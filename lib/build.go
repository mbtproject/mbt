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
	"runtime"

	git "github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
)

var defaultCheckoutOptions = &git.CheckoutOpts{
	Strategy: git.CheckoutSafe,
}

func (s *stdSystem) BuildBranch(name string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByBranch(name)
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildPr(src, dst string, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByPr(src, dst)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildDiff(from, to string, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByDiff(from, to)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildCurrentBranch(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByCurrentBranch()
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildCommit(commit string, filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByCommit(commit)
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildCommitContent(commit string, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByCommitContent(commit)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndBuildManifest(m, options)
}

func (s *stdSystem) BuildWorkspace(filterOptions *FilterOptions, options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByWorkspace()
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.buildManifest(m, options)
}

func (s *stdSystem) BuildWorkspaceChanges(options *CmdOptions) (*BuildSummary, error) {
	m, err := s.ManifestByWorkspaceChanges()
	if err != nil {
		return nil, err
	}

	return s.buildManifest(m, options)
}

func (s *stdSystem) checkoutAndBuildManifest(m *Manifest, options *CmdOptions) (*BuildSummary, error) {
	r, err := s.WorkspaceManager.CheckoutAndRun(m.Sha, func() (interface{}, error) {
		return s.buildManifest(m, options)
	})

	if err != nil {
		return nil, err
	}

	return r.(*BuildSummary), nil
}

func (s *stdSystem) buildManifest(m *Manifest, options *CmdOptions) (*BuildSummary, error) {
	completed := make([]*BuildResult, 0)
	skipped := make([]*Module, 0)

	for _, a := range m.Modules {
		cmd, ok := s.canBuildHere(a)
		if !ok {
			skipped = append(skipped, a)
			options.Callback(a, CmdStageSkipBuild, nil)
			continue
		}

		options.Callback(a, CmdStageBeforeBuild, nil)
		err := s.execBuild(cmd, m, a, options)
		if err != nil {
			return nil, err
		}
		options.Callback(a, CmdStageAfterBuild, nil)
		completed = append(completed, &BuildResult{Module: a})
	}

	return &BuildSummary{Manifest: m, Completed: completed, Skipped: skipped}, nil
}

func (s *stdSystem) execBuild(buildCmd *Cmd, manifest *Manifest, module *Module, options *CmdOptions) error {
	err := s.ProcessManager.Exec(manifest, module, options, buildCmd.Cmd, buildCmd.Args...)
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedBuild, module.Name())
	}
	return nil
}

func (s *stdSystem) canBuildHere(mod *Module) (*Cmd, bool) {
	c, ok := mod.Build()[runtime.GOOS]

	if !ok {
		c, ok = mod.Build()["default"]
	}

	return c, ok
}
