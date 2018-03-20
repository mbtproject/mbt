package lib

import (
	"io"
)

// Interfaces and types that makes up MBT system
// Other source files basically contains the implementations of one
// or more of these concepts.

/** GIT Integration Interface **/

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
	// IsDirtyWorkspace returns true if the current workspace has uncommitted changes.
	IsDirtyWorkspace() (bool, error)
	// Checkout specified commit in current workspace.
	Checkout(commit Commit) error
	// CheckoutHead checkout head commit in current workspace.
	CheckoutHead() error
	// MergeBase returns the merge base of two commits.
	MergeBase(a, b Commit) (Commit, error)
}

/** Module Discovery Interface **/

// BuildCmd represents the structure of build configuration in .mbt.yml.
type BuildCmd struct {
	Cmd  string
	Args []string `yaml:",flow"`
}

// Spec represents the structure of .mbt.yml contents.
type Spec struct {
	Name             string                 `yaml:"name"`
	Build            map[string]*BuildCmd   `yaml:"build"`
	Properties       map[string]interface{} `yaml:"properties"`
	Dependencies     []string               `yaml:"dependencies"`
	FileDependencies []string               `yaml:"fileDependencies"`
}

// Module represents a single module in the repository.
type Module struct {
	name             string
	path             string
	build            map[string]*BuildCmd
	hash             string
	version          string
	properties       map[string]interface{}
	requires         Modules
	requiredBy       Modules
	fileDependencies []string
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

/** Manifest Interface **/

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

/** Build Interface **/

// BuildStage is an enum to indicate various stages of the build.
type BuildStage = int

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
	// BuildStageBeforeBuild is the stage before executing module build command
	BuildStageBeforeBuild = iota

	// BuildStageAfterBuild is the stage after executing the module build command
	BuildStageAfterBuild

	// BuildStageSkipBuild is when module building is skipped due to lack of matching building command
	BuildStageSkipBuild
)

// BuildStageCallback is the callback function used to notify various build stages
type BuildStageCallback func(mod *Module, s BuildStage)

/** Main MBT System Interface **/

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

	// BuildBranch builds the specifed branch.
	BuildBranch(name string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildPr builds changes in 'src' branch since it diverged from 'dst' branch.
	BuildPr(src, dst string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// Build builds changes between two commits
	BuildDiff(from, to string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildCurrentBranch builds the current branch.
	BuildCurrentBranch(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildCommit builds specified commit.
	BuildCommit(commit string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildCommitChanges builds the changes in specified commit
	BuildCommitContent(commit string, stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildWorkspace builds the current workspace.
	BuildWorkspace(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

	// BuildWorkspace builds changes in current workspace.
	BuildWorkspaceChanges(stdin io.Reader, stdout, stderr io.Writer, callback BuildStageCallback) (*BuildSummary, error)

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
}

type stdSystem struct {
	Repo     Repo
	Log      Log
	MB       ManifestBuilder
	Discover Discover
	Reducer  Reducer
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
	return initSystem(log, repo, mb, discover, reducer), nil
}

func initSystem(log Log, repo Repo, mb ManifestBuilder, discover Discover, reducer Reducer) System {
	return &stdSystem{
		Log:      log,
		Repo:     repo,
		MB:       mb,
		Discover: discover,
		Reducer:  reducer,
	}
}

func (s *stdSystem) ManifestBuilder() ManifestBuilder {
	return s.MB
}
