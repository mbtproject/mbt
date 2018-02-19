package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	git "github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/intercept"
	yaml "gopkg.in/yaml.v2"
)

type TestRepository struct {
	Dir           string
	Repo          *git.Repository
	LastCommit    *git.Oid
	CurrentBranch string
}

func (r *TestRepository) InitModule(p string) error {
	return r.InitModuleWithOptions(p, &Spec{
		Name: path.Base(p),
		Build: map[string]*BuildCmd{
			"darwin":  {"./build.sh", []string{}},
			"linux":   {"./build.sh", []string{}},
			"windows": {"powershell", []string{"-ExecutionPolicy", "Bypass", "-File", ".\\build.ps1"}},
		},
		Properties: map[string]interface{}{"foo": "bar", "jar": "car"},
	})
}

func (r *TestRepository) InitModuleWithOptions(p string, mod *Spec) error {
	modDir := path.Join(r.Dir, p)
	err := os.MkdirAll(modDir, 0755)
	if err != nil {
		return err
	}

	buff, err := yaml.Marshal(mod)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(modDir, ".mbt.yml"), buff, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *TestRepository) WriteContent(file, content string) error {
	fpath := path.Join(r.Dir, file)
	dir := path.Dir(fpath)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(fpath, []byte(content), 0744)
}

func (r *TestRepository) Commit(message string) error {
	idx, err := r.Repo.Index()
	if err != nil {
		return err
	}

	err = idx.AddAll([]string{"."}, git.IndexAddCheckPathspec, func(p string, f string) int {
		return 0
	})
	if err != nil {
		return err
	}

	err = idx.Write()
	if err != nil {
		return err
	}

	oid, err := idx.WriteTree()
	if err != nil {
		return err
	}

	tree, err := r.Repo.LookupTree(oid)
	if err != nil {
		return err
	}

	sig := &git.Signature{
		Email: "alice@wonderland.com",
		Name:  "alice",
		When:  time.Now(),
	}

	parents := []*git.Commit{}
	isEmpty, err := r.Repo.IsEmpty()
	if err != nil {
		return nil
	}

	if !isEmpty {
		currentBranch, err := r.Repo.Head()
		if err != nil {
			return err
		}

		bc, err := r.Repo.LookupCommit(currentBranch.Target())
		if err != nil {
			return err
		}

		parents = append(parents, bc)
	}

	r.LastCommit, err = r.Repo.CreateCommit("HEAD", sig, sig, message, tree, parents...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TestRepository) SwitchToBranch(name string) error {
	_, err := r.Repo.LookupBranch(name, git.BranchAll)
	if err != nil {
		head, err := r.Repo.Head()
		if err != nil {
			return err
		}

		hc, err := r.Repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		_, err = r.Repo.CreateBranch(name, hc, false)
		if err != nil {
			return err
		}
	}

	err = r.Repo.SetHead(fmt.Sprintf("refs/heads/%s", name))
	if err != nil {
		return err
	}

	return r.Repo.CheckoutHead(&git.CheckoutOpts{
		Strategy: git.CheckoutForce | git.CheckoutRemoveUntracked | git.CheckoutDontWriteIndex,
	})
}

func (r *TestRepository) Remove(p string) error {
	return os.RemoveAll(path.Join(r.Dir, p))
}

func (r *TestRepository) Rename(old, new string) error {
	return os.Rename(path.Join(r.Dir, old), path.Join(r.Dir, new))
}

func (r *TestRepository) WritePowershellScript(p, content string) error {
	return r.WriteContent(p, content)
}

func (r *TestRepository) WriteShellScript(p, content string) error {
	return r.WriteContent(p, fmt.Sprintf("#!/bin/sh\n%s", content))
}

func createTestRepository(dir string) (*TestRepository, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	repo, err := git.InitRepository(dir, false)
	if err != nil {
		return nil, err
	}

	return &TestRepository{dir, repo, nil, "master"}, nil
}

func NewTestRepo(t *testing.T, dir string) *TestRepository {
	repo, err := createTestRepository(dir)
	check(t, err)
	return repo
}

func clean() {
	os.RemoveAll(".tmp")
}

func check(t *testing.T, err error) {
	if err != nil {
		ee, ok := err.(*e.E)
		if ok {
			err = ee.WithExtendedInfo()
		}
		t.Fatal(err)
	}
}

type World struct {
	Log             Log
	Repo            *TestRepo
	Discover        *TestDiscover
	Reducer         *TestReducer
	ManifestBuilder *TestManifestBuilder
	System          *TestSystem
}

/* sXxx methods below are used to safely convert interface{} to an interface type */
func sErr(e interface{}) error {
	if e == nil {
		return nil
	}
	return e.(error)
}

func sBlob(e interface{}) Blob {
	if e == nil {
		return nil
	}

	return e.(Blob)
}

func sCommit(e interface{}) Commit {
	if e == nil {
		return nil
	}

	return e.(Commit)
}

func sManifest(e interface{}) *Manifest {
	if e == nil {
		return nil
	}

	return e.(*Manifest)
}

func sModules(e interface{}) Modules {
	if e == nil {
		return nil
	}

	return e.(Modules)
}

type TestRepo struct {
	Interceptor *intercept.Interceptor
}

func (r *TestRepo) GetCommit(sha string) (Commit, error) {
	ret := r.Interceptor.Call("GetCommit", sha)
	return sCommit(ret[0]), sErr(ret[1])
}

func (r *TestRepo) Path() string {
	ret := r.Interceptor.Call("Path")
	return ret[0].(string)
}

func (r *TestRepo) Diff(a, b Commit) ([]*DiffDelta, error) {
	ret := r.Interceptor.Call("Diff", a, b)
	return ret[0].([]*DiffDelta), sErr(ret[1])
}

func (r *TestRepo) DiffMergeBase(from, to Commit) ([]*DiffDelta, error) {
	ret := r.Interceptor.Call("DiffMergeBase", from, to)
	return ret[0].([]*DiffDelta), sErr(ret[1])
}

func (r *TestRepo) DiffWorkspace() ([]*DiffDelta, error) {
	ret := r.Interceptor.Call("DiffWorkspace")
	return ret[0].([]*DiffDelta), sErr(ret[1])
}

func (r *TestRepo) WalkBlobs(a Commit, callback BlobWalkCallback) error {
	ret := r.Interceptor.Call("WalkBlobs", a, callback)
	return sErr(ret[0])
}

func (r *TestRepo) BlobContents(blob Blob) ([]byte, error) {
	ret := r.Interceptor.Call("BlobContents", blob)
	return ret[0].([]byte), sErr(ret[1])
}

func (r *TestRepo) BlobContentsFromTree(commit Commit, path string) ([]byte, error) {
	ret := r.Interceptor.Call("BlobContentsFromTree", commit, path)
	return ret[0].([]byte), sErr(ret[1])
}

func (r *TestRepo) EntryID(commit Commit, path string) (string, error) {
	ret := r.Interceptor.Call("EntryID", commit, path)
	return ret[0].(string), sErr(ret[1])
}

func (r *TestRepo) BranchCommit(name string) (Commit, error) {
	ret := r.Interceptor.Call("BranchCommit", name)
	return sCommit(ret[0]), sErr(ret[1])
}

func (r *TestRepo) CurrentBranch() (string, error) {
	ret := r.Interceptor.Call("CurrentBranch")
	return ret[0].(string), sErr(ret[1])
}

func (r *TestRepo) CurrentBranchCommit() (Commit, error) {
	ret := r.Interceptor.Call("CurrentBranchCommit")
	return sCommit(ret[0]), sErr(ret[1])
}

func (r *TestRepo) IsEmpty() (bool, error) {
	ret := r.Interceptor.Call("IsEmpty")
	return ret[0].(bool), sErr(ret[1])
}

func (r *TestRepo) IsDirtyWorkspace() (bool, error) {
	ret := r.Interceptor.Call("IsDirtyWorkspace")
	return ret[0].(bool), sErr(ret[1])
}

func (r *TestRepo) Checkout(commit Commit) error {
	ret := r.Interceptor.Call("Checkout", commit)
	return sErr(ret[0])
}

func (r *TestRepo) CheckoutHead() error {
	ret := r.Interceptor.Call("CheckoutHead")
	return sErr(ret[0])
}

func (r *TestRepo) MergeBase(a, b Commit) (Commit, error) {
	ret := r.Interceptor.Call("MergeBase", a, b)
	return sCommit(ret[0]), sErr(ret[1])
}

type TestManifestBuilder struct {
	Interceptor *intercept.Interceptor
}

func (b *TestManifestBuilder) ByDiff(from, to Commit) (*Manifest, error) {
	ret := b.Interceptor.Call("ByDiff", from, to)
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByPr(src, dst string) (*Manifest, error) {
	ret := b.Interceptor.Call("ByPr", src, dst)
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByCommit(sha Commit) (*Manifest, error) {
	ret := b.Interceptor.Call("ByCommit", sha)
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByBranch(name string) (*Manifest, error) {
	ret := b.Interceptor.Call("ByBranch", name)
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByCurrentBranch() (*Manifest, error) {
	ret := b.Interceptor.Call("ByCurrentBranch")
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByWorkspace() (*Manifest, error) {
	ret := b.Interceptor.Call("ByWorkspace")
	return sManifest(ret[0]), sErr(ret[1])
}

func (b *TestManifestBuilder) ByWorkspaceChanges() (*Manifest, error) {
	ret := b.Interceptor.Call("ByWorkspaceChanges")
	return sManifest(ret[0]), sErr(ret[1])
}

type TestSystem struct {
	Interceptor *intercept.Interceptor
}

func (s *TestSystem) ApplyBranch(templatePath, branch string, output io.Writer) error {
	ret := s.Interceptor.Call("ApplyBranch", templatePath, branch, output)
	return sErr(ret[0])
}

func (s *TestSystem) ApplyCommit(sha, templatePath string, output io.Writer) error {
	ret := s.Interceptor.Call("ApplyCommit", sha, templatePath, output)
	return sErr(ret[0])
}

func (s *TestSystem) ApplyHead(templatePath string, output io.Writer) error {
	ret := s.Interceptor.Call("ApplyHead", templatePath, output)
	return sErr(ret[0])
}

func (s *TestSystem) ApplyLocal(templatePath string, output io.Writer) error {
	ret := s.Interceptor.Call("ApplyLocal", templatePath, output)
	return sErr(ret[0])
}

func (s *TestSystem) BuildBranch(name string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildBranch", name, stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildPr(src, dst string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildPr", src, dst, stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildDiff(from, to string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildDiff", from, to, stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildCurrentBranch(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildCurrentBranch", stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildCommit(commit string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildCommit", commit, stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildWorkspace(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildWorkspace", stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) BuildWorkspaceChanges(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) error {
	ret := s.Interceptor.Call("BuildWorkspaceChanges", stdin, stdout, stderr, callback)
	return sErr(ret[0])
}

func (s *TestSystem) IntersectionByCommit(first, second string) (Modules, error) {
	ret := s.Interceptor.Call("IntersectionByCommit", first, second)
	return sModules(ret[0]), sErr(ret[1])
}

func (s *TestSystem) IntersectionByBranch(first, second string) (Modules, error) {
	ret := s.Interceptor.Call("IntersectionByBranch", first, second)
	return sModules(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByDiff(from, to string) (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByDiff", from, to)
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByPr(src, dst string) (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByPr", src, dst)
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByCommit(sha string) (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByCommit", sha)
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByBranch(name string) (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByBranch", name)
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByCurrentBranch() (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByCurrentBranch")
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByWorkspace() (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByWorkspace")
	return sManifest(ret[0]), sErr(ret[1])
}

func (s *TestSystem) ManifestByWorkspaceChanges() (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByWorkspaceChanges")
	return sManifest(ret[0]), sErr(ret[1])
}

type TestDiscover struct {
	Interceptor *intercept.Interceptor
}

func (d *TestDiscover) ModulesInCommit(commit Commit) (Modules, error) {
	ret := d.Interceptor.Call("ModulesInCommit", commit)
	return sModules(ret[0]), sErr(ret[1])
}

func (d *TestDiscover) ModulesInWorkspace() (Modules, error) {
	ret := d.Interceptor.Call("ModulesInWorkspace")
	return sModules(ret[0]), sErr(ret[1])
}

type TestReducer struct {
	Interceptor *intercept.Interceptor
}

func (r *TestReducer) Reduce(modules Modules, deltas []*DiffDelta) (Modules, error) {
	ret := r.Interceptor.Call("Reduce", modules, deltas)
	return sModules(ret[0]), sErr(ret[1])
}

func buildWorld(repo string, failureCallback func(error)) *World {
	log := NewStdLog(LogLevelNormal)
	libgitRepo, err := NewLibgitRepo(repo, log)
	if err != nil {
		failureCallback(err)
		return nil
	}

	r := &TestRepo{Interceptor: intercept.NewInterceptor(libgitRepo)}

	discover := &TestDiscover{Interceptor: intercept.NewInterceptor(NewDiscover(r, log))}
	reducer := &TestReducer{Interceptor: intercept.NewInterceptor(NewReducer(log))}
	mb := &TestManifestBuilder{Interceptor: intercept.NewInterceptor(NewManifestBuilder(r, reducer, discover, log))}

	return &World{
		Log:             log,
		Repo:            r,
		Discover:        discover,
		Reducer:         reducer,
		ManifestBuilder: mb,
		System:          &TestSystem{Interceptor: intercept.NewInterceptor(initSystem(log, r, mb, discover, reducer))},
	}
}

func NewWorld(t *testing.T, repo string) *World {
	return buildWorld(repo, func(err error) { check(t, err) })
}

func NewBenchmarkWorld(b *testing.B, repo string) *World {
	return buildWorld(repo, func(err error) { b.Fatal(err) })
}
