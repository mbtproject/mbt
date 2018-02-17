package lib

func (s *stdSystem) IntersectionByCommit(first, second string) (Modules, error) {
	c1, err := s.Repo.GetCommit(first)
	if err != nil {
		return nil, err
	}

	c2, err := s.Repo.GetCommit(second)
	if err != nil {
		return nil, err
	}

	return s.intersectionCore(c1, c2)
}

func (s *stdSystem) IntersectionByBranch(first, second string) (Modules, error) {
	fc, err := s.Repo.BranchCommit(first)
	if err != nil {
		return nil, err
	}

	sc, err := s.Repo.BranchCommit(second)
	if err != nil {
		return nil, err
	}

	return s.intersectionCore(fc, sc)
}

func (s *stdSystem) intersectionCore(first, second Commit) (Modules, error) {
	repo := s.Repo
	discover := s.Discover
	reducer := s.Reducer

	base, err := repo.MergeBase(first, second)
	if err != nil {
		return nil, err
	}

	modules, err := discover.ModulesInCommit(first)
	if err != nil {
		return nil, err
	}

	diff, err := repo.Diff(first, base)
	if err != nil {
		return nil, err
	}

	firstSet, err := reducer.Reduce(modules, diff)
	if err != nil {
		return nil, err
	}

	firstSetWithDeps, err := firstSet.expandRequiresDependencies()
	if err != nil {
		return nil, err
	}

	modules, err = discover.ModulesInCommit(second)
	if err != nil {
		return nil, err
	}

	diff, err = repo.Diff(second, base)
	if err != nil {
		return nil, err
	}
	secondSet, err := reducer.Reduce(modules, diff)
	if err != nil {
		return nil, err
	}

	secondSetWithDeps, err := secondSet.expandRequiresDependencies()
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
