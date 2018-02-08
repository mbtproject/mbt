package lib

import (
	git "github.com/libgit2/git2go"
)

func openRepo(dir string) (*git.Repository, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, wrapm(ErrClassUser, err, msgFailedOpenRepo, dir)
	}
	return repo, nil
}

func statusCount(repo *git.Repository) (int, error) {
	status, err := repo.StatusList(&git.StatusOptions{
		Flags: git.StatusOptIncludeUntracked,
	})

	if err != nil {
		return 0, wrap(ErrClassInternal, err)
	}

	defer status.Free()

	return status.EntryCount()
}

func isWorkingDirDirty(repo *git.Repository) (bool, error) {
	count, err := statusCount(repo)
	return count > 0, err
}

func getBranchCommit(repo *git.Repository, branch string) (*git.Commit, error) {
	ref, err := repo.References.Dwim(branch)
	if err != nil {
		return nil, wrapm(ErrClassUser, err, msgFailedBranchLookup, branch)
	}

	oid := ref.Target()
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	return commit, nil
}

func getBranchTree(repo *git.Repository, branch string) (*git.Tree, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	return tree, nil
}

func getDiffFromMergeBase(repo *git.Repository, srcC, dstC *git.Commit) (*git.Diff, error) {
	baseOid, err := repo.MergeBase(srcC.Id(), dstC.Id())
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	baseC, err := repo.LookupCommit(baseOid)
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	baseTree, err := baseC.Tree()
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	srcTree, err := srcC.Tree()
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	diff, err := repo.DiffTreeToTree(baseTree, srcTree, &git.DiffOptions{})
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	return diff, err
}

func getDiffFromIndex(repo *git.Repository) (*git.Diff, error) {
	index, err := repo.Index()
	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	// Diff flags below are essential to get a list of
	// untracked files (including the ones inside new directories)
	// in the diff.
	// Without git.DiffRecurseUntracked option, if a new file is added inside
	// a new directory, we only get the path to the directory.
	// This option is same as running git status -uall
	diff, err := repo.DiffIndexToWorkdir(index, &git.DiffOptions{
		Flags: git.DiffIncludeUntracked | git.DiffRecurseUntracked,
	})

	if err != nil {
		return nil, wrap(ErrClassInternal, err)
	}

	return diff, err
}

func getCommit(repo *git.Repository, commitSha string) (*git.Commit, error) {
	commitOid, err := git.NewOid(commitSha)
	if err != nil {
		return nil, wrapm(ErrClassUser, err, msgInvalidSha, commitSha)
	}

	commit, err := repo.LookupCommit(commitOid)
	if err != nil {
		return nil, wrapm(ErrClassUser, err, msgCommitShaNotFound, commitSha)
	}

	return commit, nil
}

func getBranchName(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", wrap(ErrClassInternal, err)
	}

	name, err := head.Branch().Name()
	if err != nil {
		return "", wrap(ErrClassInternal, err)
	}

	return name, nil
}

func getHeadCommit(repo *git.Repository) (*git.Commit, error) {
	branch, err := getBranchName(repo)
	if err != nil {
		return nil, err
	}

	return getBranchCommit(repo, branch)
}
