package lib

import (
	"os"
	"os/exec"
	"path"
	"runtime"

	git "github.com/libgit2/git2go"
)

func buildOne(dir string, app *VersionedApplication, args []string) error {
	cmd := exec.Command(app.Application.Build)
	cmd.Env = os.Environ()
	cmd.Dir = path.Join(dir, app.Application.Path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = append(cmd.Args, append(args, "--version", app.Version)...)
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

func Build(m *Manifest, args []string) error {
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
	err = repo.CheckoutTree(tree, &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	})
	if err != nil {
		return err
	}

	for _, a := range m.Applications {
		if !canBuildHere(a.Application) {
			continue
		}

		err := buildOne(m.Dir, a, args)
		if err != nil {
			return err
		}
	}

	return nil
}
