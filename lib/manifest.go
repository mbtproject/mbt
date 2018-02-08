package lib

import (
	git "github.com/libgit2/git2go"
)

// Manifest represents a collection modules in the repository.
type Manifest struct {
	Dir     string
	Sha     string
	Modules Modules
}

// ManifestByPr returns the manifest of pull request.
func ManifestByPr(dir, src, dst string) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {
		srcC, err := getBranchCommit(repo, src)
		if err != nil {
			return nil, err
		}

		dstC, err := getBranchCommit(repo, dst)
		if err != err {
			return nil, err
		}

		a, err := modulesInDiffWithDependents(repo, srcC, dstC)
		if err != nil {
			return nil, err
		}

		return &Manifest{Modules: a, Dir: dir, Sha: srcC.Id().String()}, nil
	})
}

// ManifestBySha returns the manifest as of the specified commit sha.
func ManifestBySha(dir, sha string) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {
		commit, err := getCommit(repo, sha)
		if err != nil {
			return nil, err
		}

		return fromCommit(repo, dir, commit)
	})
}

// ManifestByBranch returns the manifest as of the tip of the specified branch.
func ManifestByBranch(dir, branch string) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {
		return fromBranch(repo, dir, branch)
	})
}

// ManifestByDiff returns the manifest for the diff between given two commits.
func ManifestByDiff(dir, from, to string) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {
		fromC, err := getCommit(repo, from)
		if err != nil {
			return nil, err
		}

		toC, err := getCommit(repo, to)
		if err != nil {
			return nil, err
		}

		a, err := modulesInDiffWithDependents(repo, toC, fromC)
		if err != nil {
			return nil, err
		}

		return &Manifest{Modules: a, Dir: dir, Sha: to}, nil
	})
}

// ManifestByLocalDir returns the manifest of the diff between the local branch and the working directory
func ManifestByLocalDir(dir string, all bool) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {
		return fromDirectory(repo, dir, "local", all)
	})
}

// ManifestByHead returns the manifest for head of the current branch.
func ManifestByHead(dir string) (*Manifest, error) {
	return buildManifest(dir, func(repo *git.Repository) (*Manifest, error) {

		branch, err := getBranchName(repo)
		if err != nil {
			return nil, err
		}

		return fromBranch(repo, dir, branch)
	})
}

func fromBranch(repo *git.Repository, dir string, branch string) (*Manifest, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}

	return fromCommit(repo, dir, commit)
}

func fromCommit(repo *git.Repository, dir string, commit *git.Commit) (*Manifest, error) {
	metadataSet, err := discoverMetadata(repo, commit)
	if err != nil {
		return nil, err
	}

	vmods, err := metadataSet.toModules()
	if err != nil {
		return nil, err
	}

	return &Manifest{dir, commit.Id().String(), vmods}, nil
}

func fromDirectory(repo *git.Repository, dir, sha string, all bool) (*Manifest, error) {
	var vmods Modules
	var err error
	if all == false {
		vmods, err = modulesInDirectoryDiff(repo, dir)
	} else {
		vmods, err = modulesInDirectory(repo, dir)
	}

	if err != nil {
		return nil, err
	}

	return &Manifest{Modules: vmods, Dir: dir, Sha: sha}, nil
}

func newEmptyManifest(dir string) *Manifest {
	return &Manifest{Modules: []*Module{}, Dir: dir, Sha: ""}
}

type manifestBuilder func(*git.Repository) (*Manifest, error)

func buildManifest(dir string, builder manifestBuilder) (*Manifest, error) {
	repo, m, err := openRepoSafe(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	return builder(repo)
}

func openRepoSafe(dir string) (*git.Repository, *Manifest, error) {
	repo, err := openRepo(dir)
	if err != nil {
		return nil, nil, err
	}

	empty, err := repo.IsEmpty()
	if err != nil {
		return nil, nil, wrap(ErrClassInternal, err)
	}

	if empty {
		return nil, newEmptyManifest(dir), nil
	}

	return repo, nil, nil
}
