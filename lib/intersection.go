package lib

import git "github.com/libgit2/git2go"

// IntersectionByCommit returns the manifest of intersection of modules modified
// between two commits.
// If we consider M as the merge base of first and second commits,
// intersection contains the modules that have been changed
// between M and first and M and second.
func IntersectionByCommit(dir, first, second string) (Modules, error) {
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

// IntersectionByBranch returns the manifest of intersection of modules modified
// between two branches.
// If we consider M as the merge base of first and second branches,
// intersection contains the modules that have been changed
// between M and first and M and second.
func IntersectionByBranch(dir, first, second string) (Modules, error) {
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

func intersectionCore(repo *git.Repository, first, second *git.Commit) (Modules, error) {
	baseOid, err := repo.MergeBase(first.Id(), second.Id())
	if err != nil {
		return nil, wrap(err)
	}

	base, err := repo.LookupCommit(baseOid)
	if err != nil {
		return nil, wrap(err)
	}

	firstSet, err := modulesInDiff(repo, first, base)
	if err != nil {
		return nil, err
	}

	firstSetWithDeps, err := modulesInDiffWithDependencies(repo, first, base)
	if err != nil {
		return nil, err
	}

	secondSet, err := modulesInDiff(repo, second, base)
	if err != nil {
		return nil, err
	}

	secondSetWithDeps, err := modulesInDiffWithDependencies(repo, second, base)
	if err != nil {
		return nil, err
	}

	intersection := make(map[string]*Module)
	firstMap := firstSet.indexByName()
	secondMap := secondSet.indexByName()

	merge := func(changesWithDependencies Modules, otherChanges map[string]*Module, intersection map[string]*Module) {
		for _, mod := range changesWithDependencies {
			if _, ok := otherChanges[mod.Name()]; ok {
				intersection[mod.Name()] = mod
			}
		}
	}

	merge(firstSetWithDeps, secondMap, intersection)
	merge(secondSetWithDeps, firstMap, intersection)

	result := make([]*Module, len(intersection))
	i := 0
	for _, v := range intersection {
		result[i] = v
		i++
	}

	return result, nil
}
