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
	"path/filepath"
)

// NewManifestBuilder creates a new ManifestBuilder
func NewManifestBuilder(repo Repo, reducer Reducer, discover Discover, log Log) ManifestBuilder {
	return &stdManifestBuilder{Repo: repo, Discover: discover, Log: log, Reducer: reducer}
}

type stdManifestBuilder struct {
	Log      Log
	Repo     Repo
	Discover Discover
	Reducer  Reducer
}

type manifestBuilder func() (*Manifest, error)

func (b *stdManifestBuilder) ByDiff(from, to Commit) (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		mods, err := b.Discover.ModulesInCommit(to)
		if err != nil {
			return nil, err
		}

		deltas, err := b.Repo.DiffMergeBase(from, to)
		if err != nil {
			return nil, err
		}

		mods, err = b.Reducer.Reduce(mods, deltas)
		if err != nil {
			return nil, err
		}

		mods, err = mods.expandRequiredByDependencies()
		if err != nil {
			return nil, err
		}

		return b.buildManifest(mods, to.ID())
	})
}

func (b *stdManifestBuilder) ByPr(src, dst string) (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		from, err := b.Repo.BranchCommit(dst)
		if err != nil {
			return nil, err
		}

		to, err := b.Repo.BranchCommit(src)
		if err != nil {
			return nil, err
		}

		return b.ByDiff(from, to)
	})
}

func (b *stdManifestBuilder) ByCommit(sha Commit) (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		mods, err := b.Discover.ModulesInCommit(sha)
		if err != nil {
			return nil, err
		}

		return b.buildManifest(mods, sha.ID())
	})
}

func (b *stdManifestBuilder) ByCommitContent(sha Commit) (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		mods, err := b.Discover.ModulesInCommit(sha)
		if err != nil {
			return nil, err
		}

		diff, err := b.Repo.Changes(sha)
		if err != nil {
			return nil, err
		}

		if len(diff) > 0 {
			mods, err = b.Reducer.Reduce(mods, diff)
			if err != nil {
				return nil, err
			}

			mods, err = mods.expandRequiredByDependencies()
			if err != nil {
				return nil, err
			}
		}

		return b.buildManifest(mods, sha.ID())
	})
}

func (b *stdManifestBuilder) ByBranch(name string) (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		c, err := b.Repo.BranchCommit(name)
		if err != nil {
			return nil, err
		}

		return b.ByCommit(c)
	})
}

func (b *stdManifestBuilder) ByCurrentBranch() (*Manifest, error) {
	return b.runManifestBuilder(func() (*Manifest, error) {
		n, err := b.Repo.CurrentBranch()
		if err != nil {
			return nil, err
		}
		return b.ByBranch(n)
	})
}

func (b *stdManifestBuilder) ByWorkspace() (*Manifest, error) {
	mods, err := b.Discover.ModulesInWorkspace()
	if err != nil {
		return nil, err
	}

	return b.buildManifest(mods, "local")
}

func (b *stdManifestBuilder) ByWorkspaceChanges() (*Manifest, error) {
	mods, err := b.Discover.ModulesInWorkspace()
	if err != nil {
		return nil, err
	}

	deltas, err := b.Repo.DiffWorkspace()
	if err != nil {
		return nil, err
	}

	mods, err = b.Reducer.Reduce(mods, deltas)
	if err != nil {
		return nil, err
	}

	mods, err = mods.expandRequiredByDependencies()
	if err != nil {
		return nil, err
	}

	return b.buildManifest(mods, "local")
}

func (b *stdManifestBuilder) runManifestBuilder(builder manifestBuilder) (*Manifest, error) {
	empty, err := b.Repo.IsEmpty()
	if err != nil {
		return nil, err
	}

	if empty {
		return b.buildManifest(Modules{}, "")
	}

	return builder()
}

func (b *stdManifestBuilder) buildManifest(modules Modules, sha string) (*Manifest, error) {
	repoPath := b.Repo.Path()
	if !filepath.IsAbs(repoPath) {
		var err error
		repoPath, err = filepath.Abs(repoPath)
		if err != nil {
			return nil, err
		}
	}
	return &Manifest{Dir: repoPath, Modules: modules, Sha: sha}, nil
}
