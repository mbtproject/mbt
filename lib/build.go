package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	git "github.com/libgit2/git2go"
	"gopkg.in/sirupsen/logrus.v1"
)

// BuildStage is an enum to indicate various stages of the build.
type BuildStage = int

var defaultCheckoutOptions = &git.CheckoutOpts{
	Strategy: git.CheckoutSafe,
}

const (
	BuildStageBeforeBuild = iota
	BuildStageAfterBuild
	BuildStageSkipBuild
)

// Build runs the build for the modules in specified manifest.
func Build(m *Manifest, stdin io.Reader, stdout, stderr io.Writer, buildStageCallback func(app *Module, s BuildStage)) error {
	repo, err := git.OpenRepository(m.Dir)
	if err != nil {
		return wrap(err)
	}

	dirty, err := isWorkingDirDirty(repo)
	if err != nil {
		return err
	}

	if dirty {
		return newError("dirty working dir")
	}

	oid, err := git.NewOid(m.Sha)
	if err != nil {
		return wrap(err)
	}

	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return wrap(err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return wrap(err)
	}

	// TODO: Confirm the strategy is correct
	err = repo.CheckoutTree(tree, defaultCheckoutOptions)
	if err != nil {
		return wrap(err)
	}

	defer checkoutHead(repo)

	for _, a := range m.Modules {
		if !canBuildHere(a) {
			buildStageCallback(a, BuildStageSkipBuild)
			continue
		}

		buildStageCallback(a, BuildStageBeforeBuild)
		err := buildOne(m, a, stdin, stdout, stderr)
		if err != nil {
			return wrap(err)
		}
		buildStageCallback(a, BuildStageAfterBuild)
	}

	return nil
}

func setupAppBuildEnvironment(manifest *Manifest, app *Module) []string {
	r := []string{
		fmt.Sprintf("MBT_BUILD_COMMIT=%s", manifest.Sha),
		fmt.Sprintf("MBT_APP_VERSION=%s", app.Version()),
		fmt.Sprintf("MBT_APP_NAME=%s", app.Name()),
	}

	for k, v := range app.Properties() {
		if value, ok := v.(string); ok {
			r = append(r, fmt.Sprintf("MBT_APP_PROPERTY_%s=%s", strings.ToUpper(k), value))
		}
	}

	return r
}

func buildOne(manifest *Manifest, app *Module, stdin io.Reader, stdout, stderr io.Writer) error {
	build := app.Build()[runtime.GOOS]
	cmd := exec.Command(build.Cmd)
	cmd.Env = append(os.Environ(), setupAppBuildEnvironment(manifest, app)...)
	cmd.Dir = path.Join(manifest.Dir, app.Path())
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Args = append(cmd.Args, build.Args...)
	err := cmd.Run()
	return err
}

func canBuildHere(app *Module) bool {
	_, ok := app.Build()[runtime.GOOS]
	return ok
}

func checkoutHead(repo *git.Repository) {
	err := repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})
	if err != nil {
		logrus.Warnf("failed to checkout head: %s", err)
	}
}
