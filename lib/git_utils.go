package lib

import git "github.com/libgit2/git2go"

func statusCount(repo *git.Repository) (int, error) {
	status, err := repo.StatusList(&git.StatusOptions{
		Flags: git.StatusOptIncludeUntracked,
	})

	if err != nil {
		return 0, err
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
		return nil, err
	}

	oid := ref.Target()
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return tree, nil
}

func getDiffFromMergeBase(repo *git.Repository, srcC, dstC *git.Commit) (*git.Diff, error) {
	baseOid, err := repo.MergeBase(srcC.Id(), dstC.Id())
	if err != nil {
		return nil, err
	}

	baseC, err := repo.LookupCommit(baseOid)
	if err != nil {
		return nil, err
	}

	baseTree, err := baseC.Tree()
	if err != nil {
		return nil, err
	}

	srcTree, err := srcC.Tree()
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(baseTree, srcTree, &git.DiffOptions{})
	if err != nil {
		return nil, err
	}

	return diff, err
}
