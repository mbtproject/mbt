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
		for _, f := range filters {
			match := false
			if filterOptions.Fuzzy {
				match = utils.IsSubsequence(lowerModuleName, f, true)
			} else {
				match = lowerModuleName == f
			}
			if match {
				// We've got a match. Append it to the list
				// and discard rest of the filters for this
				// module.
				filteredModules = append(filteredModules, m)
				break
			}
		}
	}

	return &Manifest{Dir: m.Dir, Modules: filteredModules, Sha: m.Sha}
}

// ApplyFilters will filter the modules in the manifest to the ones that
// matches the specified filter. If filter is not specified, original
// manifest is returned.
func (m *Manifest) ApplyFilters(filterOptions *FilterOptions) *Manifest {
	if filterOptions == nil {
		panic("filterOptions cannot be nil")
	}

	if filterOptions.Name != "" {
		return m.FilterByName(filterOptions)
	}
	return m
}
