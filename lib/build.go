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
	"github.com/mbtproject/mbt/e"
)

var defaultCheckoutOptions = &git.CheckoutOpts{
	Strategy: git.CheckoutSafe,
}

func (s *stdSystem) BuildBranch(name string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByBranch(name)
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildPr(src, dst string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByPr(src, dst)
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildDiff(from, to string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByDiff(from, to)
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildCurrentBranch(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByCurrentBranch()
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildCommit(commit string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByCommit(commit)
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildCommitContent(commit string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByCommitContent(commit)
	if err != nil {
		return err
	}

	return build(s.Repo, m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildWorkspace(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByWorkspace()
	if err != nil {
		return err
	}

	return buildDir(m, stdin, stdout, stderr, callback, s.Log)
}

func (s *stdSystem) BuildWorkspaceChanges(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	m, err := s.ManifestByWorkspaceChanges()
	if err != nil {
		return err
	}

	return buildDir(m, stdin, stdout, stderr, callback, s.Log)
}

// Build runs the build for the modules in specified manifest.
func build(repo Repo, m *Manifest, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback, log Log) error {
	dirty, err := repo.IsDirtyWorkspace()
	if err != nil {
		return err
	}

	if dirty {
		return e.NewError(ErrClassUser, "dirty working dir")
	}

	commit, err := repo.GetCommit(m.Sha)
	if err != nil {
		return err
	}

	defer checkoutHead(repo, log)

	// TODO: Confirm the strategy is correct
	err = repo.Checkout(commit)
	if err != nil {
		return e.Wrap(ErrClassInternal, err)
	}

	for _, a := range m.Modules {
		if !canBuildHere(a) {
			callback(a, BuildStageSkipBuild)
			continue
		}

		callback(a, BuildStageBeforeBuild)
		err := buildOne(m, a, stdin, stdout, stderr)
		if err != nil {
			return err
		}
		callback(a, BuildStageAfterBuild)
	}

	return nil
}

// BuildDir runs the build for the modules in the specified directory.
func buildDir(m *Manifest, stdin io.Reader, stdout, stderr io.Writer, buildStageCallback BuildStageCallback, log Log) error {
	for _, a := range m.Modules {
		if !canBuildHere(a) {
			buildStageCallback(a, BuildStageSkipBuild)
			continue
		}

		buildStageCallback(a, BuildStageBeforeBuild)
		err := buildOne(m, a, stdin, stdout, stderr)
		if err != nil {
			return err
		}
		buildStageCallback(a, BuildStageAfterBuild)
	}

	return nil
}

func setupModBuildEnvironment(manifest *Manifest, mod *Module) []string {
	r := []string{
		fmt.Sprintf("MBT_BUILD_COMMIT=%s", manifest.Sha),
		fmt.Sprintf("MBT_MODULE_VERSION=%s", mod.Version()),
		fmt.Sprintf("MBT_MODULE_NAME=%s", mod.Name()),
	}

	for k, v := range mod.Properties() {
		if value, ok := v.(string); ok {
			r = append(r, fmt.Sprintf("MBT_MODULE_PROPERTY_%s=%s", strings.ToUpper(k), value))
		}
	}

	return r
}

func buildOne(manifest *Manifest, mod *Module, stdin io.Reader, stdout, stderr io.Writer) error {
	build := mod.Build()[runtime.GOOS]
	cmd := exec.Command(build.Cmd)
	cmd.Env = append(os.Environ(), setupModBuildEnvironment(manifest, mod)...)
	cmd.Dir = path.Join(manifest.Dir, mod.Path())
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Args = append(cmd.Args, build.Args...)
	err := cmd.Run()
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedBuild, mod.Name())
	}
	return nil
}

func canBuildHere(mod *Module) bool {
	_, ok := mod.Build()[runtime.GOOS]
	return ok
}

func checkoutHead(repo Repo, log Log) {
	err := repo.CheckoutHead()
	if err != nil {
		log.Warnf("Failed to checkout head: %s", err)
	}
}
