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
