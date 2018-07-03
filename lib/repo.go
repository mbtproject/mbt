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

	git "github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
)

type libgitBlob struct {
	path   string
	commit *libgitCommit
	entry  *git.TreeEntry
}

func (b *libgitBlob) ID() string {
	return b.entry.Id.String()
}

func (b *libgitBlob) Name() string {
	return b.entry.Name
}

func (b *libgitBlob) Path() string {
	return b.path
}

func (b *libgitBlob) String() string {
	return fmt.Sprintf("%s%s", b.Path(), b.Name())
}

type libgitCommit struct {
	commit *git.Commit
	tree   *git.Tree
}

func (c *libgitCommit) ID() string {
	return c.commit.Id().String()
}

func (c *libgitCommit) String() string {
	return c.ID()
}

type libgitReference struct {
	reference *git.Reference
}

func (r *libgitReference) Name() string {
	return r.reference.Name()
}

type libgitRepo struct {
	path string
	Repo *git.Repository
	Log  Log
}

func (c *libgitCommit) Tree() (*git.Tree, error) {
	if c.tree == nil {
		tree, err := c.commit.Tree()
		if err != nil {
			return nil, e.Wrapf(ErrClassInternal, err, msgFailedTreeLoad, c.commit.Id())
		}
		c.tree = tree
	}

	return c.tree, nil
}

// NewLibgitRepo creates a libgit2 repo instance
func NewLibgitRepo(path string, log Log) (Repo, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, e.Wrapf(ErrClassUser, err, msgFailedOpenRepo, path)
	}

	return &libgitRepo{
		path: path,
		Repo: repo,
		Log:  log,
	}, nil
}

func (r *libgitRepo) GetCommit(commitSha string) (Commit, error) {
	commitOid, err := git.NewOid(commitSha)
	if err != nil {
		return nil, e.Wrapf(ErrClassUser, err, msgInvalidSha, commitSha)
	}

	commit, err := r.Repo.LookupCommit(commitOid)
	if err != nil {
		return nil, e.Wrapf(ErrClassUser, err, msgCommitShaNotFound, commitSha)
	}

	return &libgitCommit{commit: commit}, nil
}

func (r *libgitRepo) Path() string {
	return r.path
}

func (r *libgitRepo) Diff(a, b Commit) ([]*DiffDelta, error) {
	diff, err := diff(r.Repo, a, b)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return deltas(diff)
}

func (r *libgitRepo) DiffMergeBase(from, to Commit) ([]*DiffDelta, error) {
	bc, err := r.MergeBase(from, to)
	if err != nil {
		return nil, err
	}

	diff, err := diff(r.Repo, bc, to)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return deltas(diff)
}

func (r *libgitRepo) DiffWorkspace() ([]*DiffDelta, error) {
	index, err := r.Repo.Index()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	// Diff flags below are essential to get a list of
	// untracked files (including the ones inside new directories)
	// in the diff.
	// Without git.DiffRecurseUntracked option, if a new file is added inside
	// a new directory, we only get the path to the directory.
	// This option is same as running git status -uall
	diff, err := r.Repo.DiffIndexToWorkdir(index, &git.DiffOptions{
		Flags: git.DiffIncludeUntracked | git.DiffRecurseUntracked,
	})

	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return deltas(diff)
}

func (r *libgitRepo) Changes(c Commit) ([]*DiffDelta, error) {
	commit := c.(*libgitCommit).commit
	repo := r.Repo
	t1, err := c.(*libgitCommit).Tree()
	if err != nil {
		return nil, err
	}

	np := commit.ParentCount()
	r.Log.Debug("Commit %v has %v parents", commit, np)
	if np == 0 {
		return []*DiffDelta{}, nil
	}

	p := commit.Parent(0)
	r.Log.Debug("Changes are based on parent %v", p)

	t2, err := p.Tree()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	d, err := repo.DiffTreeToTree(t1, t2, &git.DiffOptions{})
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return deltas(d)
}

func (r *libgitRepo) WalkBlobs(commit Commit, callback BlobWalkCallback) error {
	tree, err := commit.(*libgitCommit).Tree()
	if err != nil {
		return err
	}

	var (
		walkErr error
	)

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Type == git.ObjectBlob {
			b := &libgitBlob{
				entry:  entry,
				path:   path,
				commit: commit.(*libgitCommit),
			}
			walkErr = callback(b)
			if walkErr != nil {
				return -1
			}
		}
		return 0
	})

	if walkErr != nil {
		return walkErr
	}

	if err != nil {
		return e.Wrapf(ErrClassInternal, err, msgFailedTreeWalk, tree.Id())
	}

	return nil
}

func (r *libgitRepo) BlobContents(blob Blob) ([]byte, error) {
	bl, err := r.Repo.LookupBlob(blob.(*libgitBlob).entry.Id)
	if err != nil {
		return nil, e.Wrapf(ErrClassInternal, err, "error while fetching the blob object for %s%s", blob.Path, blob.Name)
	}

	return bl.Contents(), nil
}

func (r *libgitRepo) EntryID(commit Commit, path string) (string, error) {
	tree, err := commit.(*libgitCommit).Tree()
	if err != nil {
		return "", err
	}

	entry, err := tree.EntryByPath(path)
	if err != nil {
		return "", e.Wrapf(ErrClassInternal, err, "error while fetching the tree entry for %s", path)
	}

	return entry.Id.String(), nil
}

func (r *libgitRepo) BranchCommit(name string) (Commit, error) {
	repo := r.Repo
	ref, err := repo.References.Dwim(name)
	if err != nil {
		return nil, e.Wrapf(ErrClassUser, err, msgFailedBranchLookup, name)
	}

	return r.GetCommit(ref.Target().String())
}

func (r *libgitRepo) CurrentBranch() (string, error) {
	head, err := r.Repo.Head()
	if err != nil {
		return "", e.Wrap(ErrClassInternal, err)
	}

	name, err := head.Branch().Name()
	if err != nil {
		return "", e.Wrap(ErrClassInternal, err)
	}

	return name, nil
}

func (r *libgitRepo) CurrentBranchCommit() (Commit, error) {
	b, err := r.CurrentBranch()
	if err != nil {
		return nil, err
	}

	return r.BranchCommit(b)
}

func (r *libgitRepo) IsEmpty() (bool, error) {
	empty, err := r.Repo.IsEmpty()
	if err != nil {
		return false, e.Wrap(ErrClassInternal, err)
	}

	return empty, nil
}

func (r *libgitRepo) FindAllFilesInWorkspace(pathSpec []string) ([]string, error) {
	var configPaths []string
	status, err := r.Repo.StatusList(&git.StatusOptions{
		Flags:    git.StatusOptIncludeUntracked | git.StatusOptIncludeUnmodified | git.StatusOptRecurseUntrackedDirs,
		Pathspec: pathSpec,
	})

	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	defer status.Free()

	count, err := status.EntryCount()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	r.Log.Debug("Discovered %v files matching the filter", count)
	if count > 0 {
		for c := 0; c < count; c++ {
			entry, err := status.ByIndex(c)
			if err != nil {
				r.Log.Debug("Error while getting the change at index %v - %v", c, err)
				continue
			}

			if entry.Status != git.StatusWtDeleted {
				path := entry.IndexToWorkdir.NewFile.Path
				if path == "" {
					path = entry.HeadToIndex.NewFile.Path
				}
				r.Log.Debug("Matching file detected: %v", path)
				configPaths = append(configPaths, path)
			}
		}
	}

	return configPaths, nil
}

func (r *libgitRepo) EnsureSafeWorkspace() error {
	status, err := r.Repo.StatusList(&git.StatusOptions{
		Flags: git.StatusOptIncludeUntracked,
	})

	if err != nil {
		return e.Wrap(ErrClassInternal, err)
	}

	defer status.Free()

	count, err := status.EntryCount()
	if err != nil {
		return e.Wrap(ErrClassInternal, err)
	}

	if count > 0 {
		r.Log.Debug("Workspace has %v changes", count)
		r.Log.Debug("Begin tracing all changes")
		for c := 0; c < count; c++ {
			entry, err := status.ByIndex(c)
			if err != nil {
				r.Log.Debug("Error while getting the change at index %v - %v", c, err)
				continue
			}
			r.Log.Debug("%v", entry.IndexToWorkdir.NewFile)
		}
		r.Log.Debug("End tracing all changes")
		return e.NewError(ErrClassUser, msgDirtyWorkingDir)
	}
	return nil
}

func (r *libgitRepo) BlobContentsFromTree(commit Commit, path string) ([]byte, error) {
	t, err := commit.(*libgitCommit).Tree()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	item, err := t.EntryByPath(path)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	blob, err := r.Repo.LookupBlob(item.Id)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return blob.Contents(), nil
}

func (r *libgitRepo) Checkout(commit Commit) (Reference, error) {
	ref, err := r.Repo.Head()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	reference := &libgitReference{reference: ref}

	gitCommit := commit.(*libgitCommit)
	tree, err := gitCommit.Tree()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	options := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe,
	}

	err = r.Repo.CheckoutTree(tree, options)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	err = r.Repo.SetHeadDetached(gitCommit.commit.Id())
	if err != nil {
		return reference, e.Wrap(ErrClassInternal, err)
	}

	return reference, nil
}

func (r *libgitRepo) CheckoutReference(reference Reference) error {
	gitRef := (reference.(*libgitReference)).reference
	target := gitRef.Target()
	commit, err := r.Repo.LookupCommit(target)
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	err = r.Repo.CheckoutTree(tree, &git.CheckoutOpts{Strategy: git.CheckoutForce})
	if err != nil {
		return e.Wrap(ErrClassInternal, err)
	}

	err = r.Repo.SetHead(gitRef.Name())
	if err != nil {
		return e.Wrap(ErrClassInternal, err)
	}

	return nil
}

func (r *libgitRepo) MergeBase(a, b Commit) (Commit, error) {
	bid, err := r.Repo.MergeBase(a.(*libgitCommit).commit.Id(), b.(*libgitCommit).commit.Id())
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return r.GetCommit(bid.String())
}

func diff(repo *git.Repository, ca, cb Commit) (*git.Diff, error) {
	t1, err := ca.(*libgitCommit).Tree()
	if err != nil {
		return nil, err
	}

	t2, err := cb.(*libgitCommit).Tree()
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(t1, t2, &git.DiffOptions{})
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return diff, nil
}

func deltas(diff *git.Diff) ([]*DiffDelta, error) {
	count, err := diff.NumDeltas()
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	deltas := make([]*DiffDelta, 0, count)
	err = diff.ForEach(func(delta git.DiffDelta, num float64) (git.DiffForEachHunkCallback, error) {
		deltas = append(deltas, &DiffDelta{
			OldFile: delta.OldFile.Path,
			NewFile: delta.NewFile.Path,
		})
		return nil, nil
	}, git.DiffDetailFiles)

	return deltas, err
}
