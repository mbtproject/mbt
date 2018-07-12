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

	"github.com/mbtproject/mbt/e"
)

func (s *stdSystem) RunInBranch(command, name string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByBranch(name)
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInPr(command, src, dst string, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByPr(src, dst)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInDiff(command, from, to string, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByDiff(from, to)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInCurrentBranch(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByCurrentBranch()
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInCommit(command, commit string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByCommit(commit)
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInCommitContent(command, commit string, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByCommitContent(commit)
	if err != nil {
		return nil, err
	}

	return s.checkoutAndRunManifest(command, m, options)
}

func (s *stdSystem) RunInWorkspace(command string, filterOptions *FilterOptions, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByWorkspace()
	if err != nil {
		return nil, err
	}

	m, err = m.ApplyFilters(filterOptions)
	if err != nil {
		return nil, err
	}

	return s.runManifest(command, m, options)
}

func (s *stdSystem) RunInWorkspaceChanges(command string, options *CmdOptions) (*RunResult, error) {
	m, err := s.ManifestByWorkspaceChanges()
	if err != nil {
		return nil, err
	}

	return s.runManifest(command, m, options)
}

func (s *stdSystem) checkoutAndRunManifest(command string, m *Manifest, options *CmdOptions) (*RunResult, error) {
	r, err := s.WorkspaceManager.CheckoutAndRun(m.Sha, func() (interface{}, error) {
		return s.runManifest(command, m, options)
	})

	if err != nil {
		return nil, err
	}

	return r.(*RunResult), nil
}

func (s *stdSystem) runManifest(command string, m *Manifest, options *CmdOptions) (*RunResult, error) {
	completed := make([]*Module, 0)
	skipped := make([]*Module, 0)
	failed := make([]*CmdFailure, 0)

	var err error
	for _, a := range m.Modules {
		cmd, canRun := s.canRunHere(command, a)
		if !canRun || (err != nil && options.FailFast) {
			skipped = append(skipped, a)
			options.Callback(a, CmdStageSkipBuild, nil)
			continue
		}

		options.Callback(a, CmdStageBeforeBuild, nil)
		err = s.execCommand(cmd, m, a, options)
		if err != nil {
			failed = append(failed, &CmdFailure{Err: err, Module: a})
			options.Callback(a, CmdStageFailedBuild, err)
		} else {
			completed = append(completed, a)
			options.Callback(a, CmdStageAfterBuild, nil)
		}
	}

	return &RunResult{Manifest: m, Failures: failed, Completed: completed, Skipped: skipped}, nil
}

func (s *stdSystem) execCommand(command *UserCmd, manifest *Manifest, module *Module, options *CmdOptions) error {
	err := s.ProcessManager.Exec(manifest, module, options, command.Cmd, command.Args...)
	if err != nil {
		return e.Wrap(ErrClassUser, err)
	}
	return nil
}

func (s *stdSystem) canRunHere(command string, mod *Module) (*UserCmd, bool) {
	c, ok := mod.Commands()[command]
	if !ok {
		return nil, false
	}

	if len(c.OS) == 0 {
		return c, true
	}

	for _, os := range c.OS {
		if os == runtime.GOOS {
			return c, true
		}
	}

	return nil, false
}
