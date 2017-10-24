package lib

import git "github.com/libgit2/git2go"

// IntersectionByCommit returns the manifest of intersection of applications modified
// between two commits.
// If we consider M as the merge base of first and second commits,
// intersection contains the applications that have been changed
// between M and first and M and second.
func IntersectionByCommit(dir, first, second string) (Applications, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, wrap(err)
	}

	fc, err := getCommit(repo, first)
	if err != nil {
		return nil, err
	}

	sc, err := getCommit(repo, second)
	if err != nil {
		return nil, err
	}

	return intersectionCore(repo, fc, sc)
}

// IntersectionByBranch returns the manifest of intersection of applications modified
// between two branches.
// If we consider M as the merge base of first and second branches,
// intersection contains the applications that have been changed
// between M and first and M and second.
func IntersectionByBranch(dir, first, second string) (Applications, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, wrap(err)
	}

	fc, err := getBranchCommit(repo, first)
	if err != nil {
		return nil, err
	}

	sc, err := getBranchCommit(repo, second)
	if err != nil {
		return nil, err
	}

	return intersectionCore(repo, fc, sc)
}

func intersectionCore(repo *git.Repository, first, second *git.Commit) (Applications, error) {
	baseOid, err := repo.MergeBase(first.Id(), second.Id())
	if err != nil {
		return nil, wrap(err)
	}

	base, err := repo.LookupCommit(baseOid)
	if err != nil {
		return nil, wrap(err)
	}

	firstSet, err := applicationsInDiffWithDependencies(repo, first, base)
	if err != nil {
		return nil, err
	}

	secondSet, err := applicationsInDiffWithDependencies(repo, second, base)
	if err != nil {
		return nil, err
	}

	result := Applications{}
	firstMap := firstSet.indexByName()

	for _, app := range secondSet {
		if _, ok := firstMap[app.Name()]; ok {
			result = append(result, app)
		}
	}

	return result, nil
}
