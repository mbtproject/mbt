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
	"strings"

	"github.com/mbtproject/mbt/utils"
)

func (s *stdSystem) ManifestByDiff(from, to string) (*Manifest, error) {
	f, err := s.Repo.GetCommit(from)
	if err != nil {
		return nil, err
	}

	t, err := s.Repo.GetCommit(to)
	if err != nil {
		return nil, err
	}

	return s.MB.ByDiff(f, t)
}

func (s *stdSystem) ManifestByPr(src, dst string) (*Manifest, error) {
	return s.MB.ByPr(src, dst)
}

func (s *stdSystem) ManifestByCommit(sha string) (*Manifest, error) {
	c, err := s.Repo.GetCommit(sha)
	if err != nil {
		return nil, err
	}
	return s.MB.ByCommit(c)
}

func (s *stdSystem) ManifestByCommitContent(sha string) (*Manifest, error) {
	c, err := s.Repo.GetCommit(sha)
	if err != nil {
		return nil, err
	}
	return s.MB.ByCommitContent(c)
}

func (s *stdSystem) ManifestByBranch(name string) (*Manifest, error) {
	return s.MB.ByBranch(name)
}

func (s *stdSystem) ManifestByCurrentBranch() (*Manifest, error) {
	return s.MB.ByCurrentBranch()
}

func (s *stdSystem) ManifestByWorkspace() (*Manifest, error) {
	return s.MB.ByWorkspace()
}

func (s *stdSystem) ManifestByWorkspaceChanges() (*Manifest, error) {
	return s.MB.ByWorkspaceChanges()
}

// FilterByName reduces the modules in a Manifest to the
// ones that are matching the terms specified in filter.
// Multiple terms can be specified as a comma separated
// string.
// If fuzzy argument is true, comparison is a case insensitive
// subsequence comparison. Otherwise, it's a case insensitive
// exact match.
func (m *Manifest) FilterByName(filterOptions *FilterOptions) *Manifest {
	filteredModules := make(Modules, 0)
	filter := strings.ToLower(filterOptions.Name)
	filters := strings.Split(filter, ",")

	for _, m := range m.Modules {
		lowerModuleName := strings.ToLower(m.Name())

		match := matches(lowerModuleName, filters, filterOptions.Fuzzy)

		if match {
			filteredModules = append(filteredModules, m)
		}
	}

	return &Manifest{Dir: m.Dir, Modules: filteredModules, Sha: m.Sha}
}

// ApplyFilters will filter the modules in the manifest to the ones that
// matches the specified filter. If filter is not specified, original
// manifest is returned.
func (m *Manifest) ApplyFilters(filterOptions *FilterOptions) (*Manifest, error) {
	if filterOptions == nil {
		panic("filterOptions cannot be nil")
	}

	if filterOptions.Name != "" {
		m = m.FilterByName(filterOptions)
	}

	if filterOptions.Dependents {
		var err error

		m.Modules, err = m.Modules.expandRequiredByDependencies()

		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func matches(value string, filters []string, fuzzy bool) bool {
	match := false

	for _, f := range filters {
		if fuzzy {
			match = utils.IsSubsequence(value, f, true)
		} else {
			match = value == f
		}
		if match {
			break
		}
	}

	return match
}
