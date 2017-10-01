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
	"github.com/sirupsen/logrus"
)

type BuildStage = int

var DefaultCheckoutOptions = &git.CheckoutOpts{
	Strategy: git.CheckoutSafe,
}

const (
	BUILD_STAGE_BEFORE_BUILD = iota
	BUILD_STAGE_AFTER_BUILD
	BUILD_STAGE_SKIP_BUILD
)

func Build(m *Manifest, stdin io.Reader, stdout, stderr io.Writer, buildStageCallback func(app *Application, s BuildStage)) error {
	repo, err := git.OpenRepository(m.Dir)
	if err != nil {
		return err
	}

	oid, err := git.NewOid(m.Sha)
	if err != nil {
		return err
	}

	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	// TODO: Confirm the strategy is correct
	err = repo.CheckoutTree(tree, DefaultCheckoutOptions)
	if err != nil {
		return err
	}

	defer checkoutHead(repo)

	for _, a := range m.Applications {
		if !canBuildHere(a) {
			buildStageCallback(a, BUILD_STAGE_SKIP_BUILD)
			continue
		}

		buildStageCallback(a, BUILD_STAGE_BEFORE_BUILD)
		err := buildOne(m.Dir, a, stdin, stdout, stderr)
		if err != nil {
			return err
		}
		buildStageCallback(a, BUILD_STAGE_AFTER_BUILD)
	}

	return nil
}

func setupAppBuildEnvironment(app *Application) []string {
	r := []string{
		fmt.Sprintf("MBT_APP_VERSION=%s", app.Version),
	}

	for k, v := range app.Properties {
		if value, ok := v.(string); ok {
			r = append(r, fmt.Sprintf("MBT_APP_PROPERTY_%s=%s", strings.ToUpper(k), value))
		}
	}

	return r
}

func buildOne(dir string, app *Application, stdin io.Reader, stdout, stderr io.Writer) error {
	build := app.Build[runtime.GOOS]
	cmd := exec.Command(build.Cmd)
	cmd.Env = append(os.Environ(), setupAppBuildEnvironment(app)...)
	cmd.Dir = path.Join(dir, app.Path)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Args = append(cmd.Args, build.Args...)
	err := cmd.Run()
	return err
}

func canBuildHere(app *Application) bool {
	_, ok := app.Build[runtime.GOOS]
	return ok
}

func checkoutHead(repo *git.Repository) {
	err := repo.CheckoutHead(DefaultCheckoutOptions)
	if err != nil {
		logrus.Warnf("failed to checkout head: %s", err)
	}
}
