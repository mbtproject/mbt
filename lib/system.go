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
	"io"
	"os"
)

// This file defines the interfaces and types that make up MBT system.
// Other source files basically contain the implementations of one
// or more of these components.

/** GIT Integration **/

// Blob stored in git.
type Blob interface {
	// ID of the blob.
	ID() string
	// Name of the blob.
	Name() string
	// Path (relative) to the blob.
	Path() string
	//String returns a printable id.
	String() string
}

// DiffDelta is a single delta in a git diff.
type DiffDelta struct {
	// NewFile path of the delta
	NewFile string
	// OldFile path of the delta
	OldFile string
}

// BlobWalkCallback used for discovering blobs in a commit tree.
type BlobWalkCallback func(Blob) error

// Commit in the repo.
// All commit based APIs in Repo interface accepts this interface.
// It gives the implementations the ability to optimise the access to commits.
// For example, a libgit2 based Repo implementation we use by default caches the access to commit tree.
type Commit interface {
	ID() string
	String() string
}

// Reference to a tree in the repository.
// For example, if you consider a git repository
// this could be pointing to a branch, tag or commit.
type Reference interface {
	Name() string
}

// Repo defines the set of interactions with the git repository.
type Repo interface {
	// GetCommit returns the commit object for the specified SHA.
	GetCommit(sha string) (Commit, error)
	// Path of the repository.
	Path() string
	// Diff gets the diff between two commits.
	Diff(a, b Commit) ([]*DiffDelta, error)
	// DiffMergeBase gets the diff between the merge base of from and to and, to.
	// In other words, diff contains the deltas of changes occurred in 'to' commit tree
	// since it diverged from 'from' commit tree.
	DiffMergeBase(from, to Commit) ([]*DiffDelta, error)
	// DiffWorkspace gets the changes in current workspace.
	// This should include untracked changes.
	DiffWorkspace() ([]*DiffDelta, error)
	// Changes returns a an array of DiffDelta objects representing the changes
	// in the specified commit.
	// Return an empty array if the specified commit is the first commit
	// in the repo.
	Changes(c Commit) ([]*DiffDelta, error)
	// WalkBlobs invokes the callback for each blob reachable from the commit tree.
	WalkBlobs(a Commit, callback BlobWalkCallback) error
	// BlobContents of specified blob.
	BlobContents(blob Blob) ([]byte, error)
	// BlobContentsByPath gets the blob contents from a specific git tree.
	BlobContentsFromTree(commit Commit, path string) ([]byte, error)
	// EntryID of a git object in path.
	// ID is resolved from the commit tree of the specified commit.
	EntryID(commit Commit, path string) (string, error)
	// BranchCommit returns the last commit for the specified branch.
	BranchCommit(name string) (Commit, error)
	// CurrentBranch returns the name of current branch.
	CurrentBranch() (string, error)
	// CurrentBranchCommit returns the last commit for the current branch.
	CurrentBranchCommit() (Commit, error)
	// IsEmpty informs if the current repository is empty or not.
	IsEmpty() (bool, error)
	// FindAllFilesInWorkspace returns all files in repository matching given pathSpec, including untracked files.
	FindAllFilesInWorkspace(pathSpec []string) ([]string, error)
	// EnsureSafeWorkspace returns an error workspace is in a safe state
	// for operations requiring a checkout.
	// For example, in git repositories we consider uncommitted changes or
	// a detached head is an unsafe state.
	EnsureSafeWorkspace() error
	// Checkout specified commit into workspace.
	// Also returns a reference to the previous tree pointed by current workspace.
	Checkout(commit Commit) (Reference, error)
	// CheckoutReference checks out the specified reference into workspace.
	CheckoutReference(Reference) error
	// MergeBase returns the merge base of two commits.
	MergeBase(a, b Commit) (Commit, error)
}

/** Module Discovery **/

// Cmd represents the structure of a command appears in .mbt.yml.
type Cmd struct {
	Cmd  string
	Args []string `yaml:",flow"`
}

// UserCmd represents the structure of a user defined command in .mbt.yml
type UserCmd struct {
	Cmd  string
	Args []string `yaml:",flow"`
	OS   []string `yaml:"os"`
}

// Spec represents the structure of .mbt.yml contents.
type Spec struct {
	Name             string                 `yaml:"name"`
	Build            map[string]*Cmd        `yaml:"build"`
	Commands         map[string]*UserCmd    `yaml:"commands"`
	Properties       map[string]interface{} `yaml:"properties"`
	Dependencies     []string               `yaml:"dependencies"`
	FileDependencies []string               `yaml:"fileDependencies"`
}

// Module represents a single module in the repository.
type Module struct {
	metadata   *moduleMetadata
	version    string
	requires   Modules
	requiredBy Modules
}

// Modules is an array of Module.
type Modules []*Module

// Discover module metadata for various conditions
type Discover interface {
	// ModulesInCommit walks the git tree at a specific commit looking for
	// directories with .mbt.yml file. Returns discovered Modules.
	ModulesInCommit(commit Commit) (Modules, error)
	// ModulesInWorkspace walks current workspace looking for
	// directories with .mbt.yml file. Returns discovered Modules.
	ModulesInWorkspace() (Modules, error)
}

// Reducer reduces a given modules set to impacted set from a diff delta
type Reducer interface {
	Reduce(modules Modules, deltas []*DiffDelta) (Modules, error)
}

// Manifest represents a collection modules in the repository.
type Manifest struct {
	Dir     string
	Sha     string
	Modules Modules
}

// ManifestBuilder builds Manifest for various conditions
type ManifestBuilder interface {
	// ByDiff creates the manifest for diff between two commits
	ByDiff(from, to Commit) (*Manifest, error)
	// ByPr creates the manifest for diff between two branches
	ByPr(src, dst string) (*Manifest, error)
	// ByCommit creates the manifest for the specified commit
	ByCommit(sha Commit) (*Manifest, error)
	// ByCommitContent creates the manifest for the content of the
	// specified commit.
	ByCommitContent(sha Commit) (*Manifest, error)
	// ByBranch creates the manifest for the specified branch
	ByBranch(name string) (*Manifest, error)
	// ByCurrentBranch creates the manifest for the current branch
	ByCurrentBranch() (*Manifest, error)
	// ByWorkspace creates the manifest for the current workspace
	ByWorkspace() (*Manifest, error)
	// ByWorkspaceChanges creates the manifest for the changes in workspace
	ByWorkspaceChanges() (*Manifest, error)
}

/** Workspace Management **/

// WorkspaceManager contains various functions to manipulate workspace.
type WorkspaceManager interface {
	// CheckoutAndRun checks out the given commit and executes the specified function.
	// Returns an error if current workspace is dirty.
	// Otherwise returns the output from fn.
	CheckoutAndRun(commit string, fn func() (interface{}, error)) (interface{}, error)
}

/** Process Manager **/

// ProcessManager manages the execution of build and user defined commands.
type ProcessManager interface {
	// Exec runs an external command in the context of a module in a manifest.
	// Following actions are performed prior to executing the command:
	// - Current working directory of the target process is set to module path
	// - Initialises important information in the target process environment
	Exec(manifest *Manifest, module *Module, options *CmdOptions, command string, args ...string) error
}

/** Build **/

// CmdStage is an enum to indicate various stages of a command.
type CmdStage = int

// BuildSummary is a summary of a successful build.
type BuildSummary struct {
	// Manifest used to trigger the build
	Manifest *Manifest
	// Completed list of the modules built. This list does not
	// include the modules that were skipped due to
	// the unavailability of a build command for the
	// host platform.
	Completed []*BuildResult
	// Skipped modules due to the unavailability of a build command for
	// the host platform
	Skipped []*Module
}

// BuildResult is summary for a single module build
type BuildResult struct {
	// Module of the build result
	Module *Module
}

const (
	// CmdStageBeforeBuild is the stage before executing module build command
	CmdStageBeforeBuild = iota

	// CmdStageAfterBuild is the stage after executing the module build command
	CmdStageAfterBuild

	// CmdStageSkipBuild is when module building is skipped due to lack of matching building command
	CmdStageSkipBuild

	// CmdStageFailedBuild is when module command is failed
	CmdStageFailedBuild
)

// CmdStageCallback is the callback function used to notify various build stages
type CmdStageCallback func(mod *Module, s CmdStage, err error)

/** Main MBT System **/

// FilterOptions describe how to filter the modules in a manifest
type FilterOptions struct {
	Name       string
	Fuzzy      bool
	Dependents bool
}

// CmdOptions defines various options required by methods executing
// user defined commands.
type CmdOptions struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Callback       CmdStageCallback
	FailFast       bool
}

// CmdFailure contains the failures occurred while running a user defined command.
type CmdFailure struct {
	Module *Module
	Err    error
}

// RunResult is the result of running a user defined command.
type RunResult struct {
	Manifest  *Manifest
	Completed []*Module
	Skipped   []*Module
	Failures  []*CmdFailure
}

// System is the interface used by users to invoke the core functionality
// of this package
type System interface {
	// ApplyBranch applies the manifest of specified branch over a template.
	// Template is retrieved from the commit tree of last commit of that branch.
	ApplyBranch(templatePath, branch string, output io.Writer) error

	// ApplyCommit applies the manifest of specified commit over a template.
	// Template is retrieved from the commit tree of the specified commit.
	ApplyCommit(sha, templatePath string, output io.Writer) error

	// ApplyHead applies the manifest of current branch over a template.
	// Template is retrieved from the commit tree of last commit current branch.
	ApplyHead(templatePath string, output io.Writer) error

	// ApplyLocal applies the manifest of local workspace.
	// Template is retrieved from the current workspace.
	ApplyLocal(templatePath string, output io.Writer) error

	// BuildBranch builds the specified branch.
	// This function accepts FilterOptions to specify which modules to be built
	// within that branch.
	BuildBranch(name string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error)

	// BuildPr builds changes in 'src' branch since it diverged from 'dst' branch.
	BuildPr(src, dst string, options *CmdOptions) (*BuildSummary, error)

	// Build builds changes between two commits
	BuildDiff(from, to string, options *CmdOptions) (*BuildSummary, error)

	// BuildCurrentBranch builds the current branch.
	// This function accepts FilterOptions to specify which modules to be built
	// within that branch.
	BuildCurrentBranch(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error)

	// BuildCommit builds specified commit.
	// This function accepts FilterOptions to specify which modules to be built
	// within that branch.
	BuildCommit(commit string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error)

	// BuildCommitChanges builds the changes in specified commit
	BuildCommitContent(commit string, options *CmdOptions) (*BuildSummary, error)

	// BuildWorkspace builds the current workspace.
	// This function accepts FilterOptions to specify which modules to be built
	// within that branch.
	BuildWorkspace(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error)

	// BuildWorkspace builds changes in current workspace.
	BuildWorkspaceChanges(options *CmdOptions) (*BuildSummary, error)

	// IntersectionByCommit returns the manifest of intersection of modules modified
	// between two commits.
	// If we consider M as the merge base of first and second commits,
	// intersection contains the modules that have been changed
	// between M and first and M and second.
	IntersectionByCommit(first, second string) (Modules, error)

	// IntersectionByBranch returns the manifest of intersection of modules modified
	// between two branches.
	// If we consider M as the merge base of first and second branches,
	// intersection contains the modules that have been changed
	// between M and first and M and second.
	IntersectionByBranch(first, second string) (Modules, error)

	// ManifestByDiff creates the manifest for diff between two commits
	ManifestByDiff(from, to string) (*Manifest, error)

	// ManifestByPr creates the manifest for diff between two branches
	ManifestByPr(src, dst string) (*Manifest, error)

	// ManifestByCommit creates the manifest for the specified commit
	ManifestByCommit(sha string) (*Manifest, error)

	// ManifestByCommitContent creates the manifest for the content in specified commit
	ManifestByCommitContent(sha string) (*Manifest, error)

	// ByBranch creates the manifest for the specified branch
	ManifestByBranch(name string) (*Manifest, error)

	// ByCurrentBranch creates the manifest for the current branch
	ManifestByCurrentBranch() (*Manifest, error)

	// ByWorkspace creates the manifest for the current workspace
	ManifestByWorkspace() (*Manifest, error)

	// ByWorkspaceChanges creates the manifest for the changes in workspace
	ManifestByWorkspaceChanges() (*Manifest, error)

	// RunInBranch runs a command in a branch.
	// This function accepts FilterOptions to specify a subset of modules.
	RunInBranch(command, name string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error)

	// RunInPr runs a command in modules that have been changed in
	// 'src' branch since it diverged from 'dst' branch.
	RunInPr(command, src, dst string, options *CmdOptions) (*RunResult, error)

	// RunInDiff runs a command in modules that have been changed in 'from'
	// commit since it diverged from 'to' commit.
	RunInDiff(command, from, to string, options *CmdOptions) (*RunResult, error)

	// RunInCurrentBranch runs a command in modules in the current branch.
	// This function accepts FilterOptions to filter the modules included in this
	// operation.
	RunInCurrentBranch(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error)

	// RunInCommit runs a command in all modules in a commit.
	// This function accepts FilterOptions to filter the modules included in this
	// operation.
	RunInCommit(command, commit string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error)

	// RunInCommitContent runs a command in modules modified in a commit.
	RunInCommitContent(command, commit string, options *CmdOptions) (*RunResult, error)

	// RunInWorkspace runs a command in all modules in workspace.
	// This function accepts FilterOptions to filter the modules included in this
	// operation.
	RunInWorkspace(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error)

	// RunInWorkspaceChanges runs a command in modules modified in workspace.
	RunInWorkspaceChanges(command string, options *CmdOptions) (*RunResult, error)
}

type stdSystem struct {
	Repo             Repo
	Log              Log
	MB               ManifestBuilder
	Discover         Discover
	Reducer          Reducer
	WorkspaceManager WorkspaceManager
	ProcessManager   ProcessManager
}

// NewSystem creates a new instance of core mbt system
func NewSystem(path string, logLevel int) (System, error) {
	log := NewStdLog(logLevel)
	repo, err := NewLibgitRepo(path, log)
	if err != nil {
		return nil, err
	}
	discover := NewDiscover(repo, log)
	reducer := NewReducer(log)
	mb := NewManifestBuilder(repo, reducer, discover, log)
	wm := NewWorkspaceManager(log, repo)
	pm := NewProcessManager(log)
	return initSystem(log, repo, mb, discover, reducer, wm, pm), nil
}

// NoFilter is built-in filter that represents no filtering
var NoFilter = &FilterOptions{}

// FuzzyFilter is a helper to create a fuzzy match FilterOptions
func FuzzyFilter(name string) *FilterOptions {
	return &FilterOptions{Name: name, Fuzzy: true}
}

// ExactMatchFilter is a helper to create an exact match FilterOptions
func ExactMatchFilter(name string) *FilterOptions {
	return &FilterOptions{Name: name}
}

// FuzzyDependentsFilter is a helper to create a fuzzy match FilterOptions
func FuzzyDependentsFilter(name string) *FilterOptions {
	return &FilterOptions{Name: name, Fuzzy: true, Dependents: true}
}

// ExactMatchDependentsFilter is a helper to create an exact match FilterOptions
func ExactMatchDependentsFilter(name string) *FilterOptions {
	return &FilterOptions{Name: name, Dependents: true}
}

func initSystem(log Log, repo Repo, mb ManifestBuilder, discover Discover, reducer Reducer, workspaceManager WorkspaceManager, processManager ProcessManager) System {
	return &stdSystem{
		Log:              log,
		Repo:             repo,
		MB:               mb,
		Discover:         discover,
		Reducer:          reducer,
		WorkspaceManager: workspaceManager,
		ProcessManager:   processManager,
	}
}

func (s *stdSystem) ManifestBuilder() ManifestBuilder {
	return s.MB
}

// CmdOptionsWithStdIO creates an instance of CmdOptions with
// its streams pointing to std io streams.
func CmdOptionsWithStdIO(callback CmdStageCallback) *CmdOptions {
	return &CmdOptions{
		Callback: callback,
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	}
}
