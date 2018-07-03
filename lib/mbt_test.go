/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	yaml "github.com/go-yaml/yaml"
	git "github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/intercept"
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
		Build: map[string]*Cmd{
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

func (r *TestRepository) AppendContent(file, content string) error {
	fpath := path.Join(r.Dir, file)
	dir := path.Dir(fpath)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	fd, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.WriteString(content)
	return err
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
	branch, err := r.Repo.LookupBranch(name, git.BranchAll)
	if err != nil {
		head, err := r.Repo.Head()
		if err != nil {
			return err
		}

		hc, err := r.Repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		branch, err = r.Repo.CreateBranch(name, hc, false)
		if err != nil {
			return err
		}
	}

	commit, err := r.Repo.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	err = r.Repo.CheckoutTree(tree, &git.CheckoutOpts{
		Strategy: git.CheckoutForce,
	})

	if err != nil {
		return err
	}

	return r.Repo.SetHead(fmt.Sprintf("refs/heads/%s", name))
}

func (r *TestRepository) CheckoutAndDetach(commit string) error {
	oid, err := git.NewOid(commit)
	if err != nil {
		return err
	}

	gitCommit, err := r.Repo.LookupCommit(oid)
	if err != nil {
		return err
	}

	tree, err := gitCommit.Tree()
	if err != nil {
		return err
	}

	err = r.Repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
	if err != nil {
		return err
	}

	return r.Repo.SetHeadDetached(oid)
}

func (r *TestRepository) SimpleMerge(src, dst string) (*git.Oid, error) {
	srcRef, err := r.Repo.References.Dwim(src)
	if err != nil {
		return nil, err
	}

	srcCommit, err := r.Repo.LookupCommit(srcRef.Target())
	if err != nil {
		return nil, err
	}

	dstRef, err := r.Repo.References.Dwim(dst)
	if err != nil {
		return nil, err
	}

	dstCommit, err := r.Repo.LookupCommit(dstRef.Target())
	if err != nil {
		return nil, err
	}

	index, err := r.Repo.MergeCommits(dstCommit, srcCommit, nil)
	if err != nil {
		return nil, err
	}

	treeID, err := index.WriteTreeTo(r.Repo)
	if err != nil {
		return nil, err
	}

	mergeTree, err := r.Repo.LookupTree(treeID)
	if err != nil {
		return nil, err
	}

	sig := &git.Signature{
		Email: "alice@wonderland.com",
		Name:  "alice",
		When:  time.Now(),
	}

	head, err := r.Repo.CreateCommit(dstRef.Name(), sig, sig, "Merged", mergeTree, dstCommit, srcCommit)
	if err != nil {
		return nil, err
	}

	err = r.Repo.CheckoutHead(&git.CheckoutOpts{Strategy: git.CheckoutForce})
	if err != nil {
		return nil, err
	}

	return head, err
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

func NewTestRepoForBench(b *testing.B, dir string) *TestRepository {
	repo, err := createTestRepository(dir)
	if err != nil {
		b.Fatalf("%v", err)
	}
	return repo
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
	Log              Log
	Repo             *TestRepo
	Discover         *TestDiscover
	Reducer          *TestReducer
	ManifestBuilder  *TestManifestBuilder
	WorkspaceManager *TestWorkspaceManager
	ProcessManager   *TestProcessManager
	System           *TestSystem
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

func sBuildSummary(e interface{}) *BuildSummary {
	if e == nil {
		return nil
	}

	return e.(*BuildSummary)
}

func sRunResult(e interface{}) *RunResult {
	if e == nil {
		return nil
	}

	return e.(*RunResult)
}

func sReference(e interface{}) Reference {
	if e == nil {
		return nil
	}

	return e.(Reference)
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

func (r *TestRepo) Changes(c Commit) ([]*DiffDelta, error) {
	ret := r.Interceptor.Call("Changes", c)
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

func (r *TestRepo) FindAllFilesInWorkspace(pathSpec []string) ([]string, error) {
	ret := r.Interceptor.Call("FindAllFilesInWorkspace", pathSpec)
	return ret[0].([]string), sErr(ret[1])
}

func (r *TestRepo) EnsureSafeWorkspace() error {
	ret := r.Interceptor.Call("EnsureSafeWorkspace")
	return sErr(ret[0])
}

func (r *TestRepo) Checkout(commit Commit) (Reference, error) {
	ret := r.Interceptor.Call("Checkout", commit)
	return sReference(ret[0]), sErr(ret[1])
}

func (r *TestRepo) CheckoutReference(reference Reference) error {
	ret := r.Interceptor.Call("CheckoutReference", reference)
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

func (b *TestManifestBuilder) ByCommitContent(sha Commit) (*Manifest, error) {
	ret := b.Interceptor.Call("ByCommitContent", sha)
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

func (s *TestSystem) BuildBranch(name string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildBranch", name, filterOptions, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildPr(src, dst string, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildPr", src, dst, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildDiff(from, to string, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildDiff", from, to, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildCurrentBranch(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildCurrentBranch", filterOptions, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildCommit(commit string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildCommit", commit, filterOptions, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildCommitContent(commit string, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildCommitContent", commit, options)
	return sBuildSummary(ret[0]), sErr(ret[1])

}

func (s *TestSystem) BuildWorkspace(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildWorkspace", filterOptions, options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) BuildWorkspaceChanges(options *CmdOptions) (*BuildSummary, error) {
	ret := s.Interceptor.Call("BuildWorkspaceChanges", options)
	return sBuildSummary(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInBranch(command, name string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInBranch", command, name, filterOptions, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInPr(command, src, dst string, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInPr", command, src, dst, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInDiff(command, from, to string, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInDiff", command, from, to, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInCurrentBranch(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInCurrentBranch", command, filterOptions, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInCommit(command, commit string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInCommit", command, commit, filterOptions, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInCommitContent(command, commit string, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInCommitContent", command, commit, options)
	return sRunResult(ret[0]), sErr(ret[1])

}

func (s *TestSystem) RunInWorkspace(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInWorkspace", command, filterOptions, options)
	return sRunResult(ret[0]), sErr(ret[1])
}

func (s *TestSystem) RunInWorkspaceChanges(command string, options *CmdOptions) (*RunResult, error) {
	ret := s.Interceptor.Call("RunInWorkspaceChanges", command, options)
	return sRunResult(ret[0]), sErr(ret[1])
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

func (s *TestSystem) ManifestByCommitContent(sha string) (*Manifest, error) {
	ret := s.Interceptor.Call("ManifestByCommitContent", sha)
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

type TestWorkspaceManager struct {
	Interceptor *intercept.Interceptor
}

func (w *TestWorkspaceManager) CheckoutAndRun(commit string, fn func() (interface{}, error)) (interface{}, error) {
	ret := w.Interceptor.Call("CheckoutAndRun", commit, fn)
	return ret[0], sErr(ret[1])
}

type TestProcessManager struct {
	Interceptor *intercept.Interceptor
}

func (p *TestProcessManager) Exec(manifest *Manifest, module *Module, options *CmdOptions, command string, args ...string) error {
	rest := []interface{}{manifest, module, options, command}
	for _, a := range args {
		rest = append(rest, a)
	}
	ret := p.Interceptor.Call("Exec", rest...)
	return sErr(ret[0])
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
	wm := &TestWorkspaceManager{Interceptor: intercept.NewInterceptor(NewWorkspaceManager(log, r))}
	pm := &TestProcessManager{Interceptor: intercept.NewInterceptor(NewProcessManager(log))}

	return &World{
		Log:              log,
		Repo:             r,
		Discover:         discover,
		Reducer:          reducer,
		ManifestBuilder:  mb,
		WorkspaceManager: wm,
		ProcessManager:   pm,
		System:           &TestSystem{Interceptor: intercept.NewInterceptor(initSystem(log, r, mb, discover, reducer, wm, pm))},
	}
}

func NewWorld(t *testing.T, repo string) *World {
	return buildWorld(repo, func(err error) { check(t, err) })
}

func NewBenchmarkWorld(b *testing.B, repo string) *World {
	return buildWorld(repo, func(err error) { b.Fatal(err) })
}
