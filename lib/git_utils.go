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
