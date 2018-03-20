package lib

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
	return b.buildManifest(func() (*Manifest, error) {
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

		return &Manifest{Dir: b.Repo.Path(), Modules: mods, Sha: to.ID()}, nil
	})
}

func (b *stdManifestBuilder) ByPr(src, dst string) (*Manifest, error) {
	return b.buildManifest(func() (*Manifest, error) {
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
	return b.buildManifest(func() (*Manifest, error) {
		mods, err := b.Discover.ModulesInCommit(sha)
		if err != nil {
			return nil, err
		}

		return &Manifest{Dir: b.Repo.Path(), Modules: mods, Sha: sha.ID()}, nil
	})
}

func (b *stdManifestBuilder) ByCommitContent(sha Commit) (*Manifest, error) {
	return b.buildManifest(func() (*Manifest, error) {
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

		return &Manifest{Dir: b.Repo.Path(), Modules: mods, Sha: sha.ID()}, nil
	})
}

func (b *stdManifestBuilder) ByBranch(name string) (*Manifest, error) {
	return b.buildManifest(func() (*Manifest, error) {
		c, err := b.Repo.BranchCommit(name)
		if err != nil {
			return nil, err
		}

		return b.ByCommit(c)
	})
}

func (b *stdManifestBuilder) ByCurrentBranch() (*Manifest, error) {
	return b.buildManifest(func() (*Manifest, error) {
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

	return &Manifest{Dir: b.Repo.Path(), Modules: mods, Sha: "local"}, nil
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

	return &Manifest{Dir: b.Repo.Path(), Modules: mods, Sha: "local"}, nil
}

func (b *stdManifestBuilder) buildManifest(builder manifestBuilder) (*Manifest, error) {
	empty, err := b.Repo.IsEmpty()
	if err != nil {
		return nil, err
	}

	if empty {
		return &Manifest{Dir: b.Repo.Path(), Modules: Modules{}, Sha: ""}, nil
	}

	return builder()
}
