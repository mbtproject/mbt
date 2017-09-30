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

var DefaultCheckoutOptions = &git.CheckoutOpts{
	Strategy: git.CheckoutForce,
}

func Build(m *Manifest, stdin io.Reader, stdout, stderr io.Writer) error {
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
			continue
		}

		err := buildOne(m.Dir, a, stdin, stdout, stderr)
		if err != nil {
			return err
		}
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

func buildOne(dir string, app *Application, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command(app.Build)
	cmd.Env = append(os.Environ(), setupAppBuildEnvironment(app)...)
	cmd.Dir = path.Join(dir, app.Path)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Args = append(cmd.Args, app.Args...)
	err := cmd.Run()
	return err
}

func canBuildHere(app *Application) bool {
	for _, p := range app.BuildPlatforms {
		if p == runtime.GOOS {
			return true
		}
	}

	return false
}

func checkoutHead(repo *git.Repository) {
	err := repo.CheckoutHead(DefaultCheckoutOptions)
	if err != nil {
		logrus.Warnf("failed to checkout head: %s", err)
	}
}
