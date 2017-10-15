package lib

import (
	"encoding/hex"
	"sort"
	"strings"

	git "github.com/libgit2/git2go"
)

// Manifest represents a collection applications in the repository.
type Manifest struct {
	Dir          string
	Sha          string
	Applications Applications
}

func ManifestByPr(dir, src, dst string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	srcC, err := getBranchCommit(repo, src)
	if err != nil {
		return nil, err
	}

	dstC, err := getBranchCommit(repo, dst)
	if err != err {
		return nil, err
	}

	diff, err := getDiffFromMergeBase(repo, srcC, dstC)
	if err != nil {
		return nil, err
	}

	m, err = fromBranch(repo, dir, src)
	if err != nil {
		return nil, err
	}

	return reduceToDiff(m, diff)
}

func ManifestBySha(dir, sha string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	bytes, err := hex.DecodeString(sha)
	if err != nil {
		return nil, err
	}

	oid := git.NewOidFromBytes(bytes)
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	return fromCommit(repo, dir, commit)
}

func ManifestByBranch(dir, branch string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	return fromBranch(repo, dir, branch)
}

func ManifestByDiff(dir, from, to string) (*Manifest, error) {
	repo, m, err := openRepo(dir)
	if err != nil {
		return nil, err
	}

	if m != nil {
		return m, nil
	}

	fromOid, err := git.NewOid(from)
	if err != nil {
		return nil, err
	}

	toOid, err := git.NewOid(to)
	if err != nil {
		return nil, err
	}

	fromC, err := repo.LookupCommit(fromOid)
	if err != nil {
		return nil, err
	}

	toC, err := repo.LookupCommit(toOid)
	if err != nil {
		return nil, err
	}

	diff, err := getDiffFromMergeBase(repo, toC, fromC)
	if err != nil {
		return nil, err
	}

	m, err = fromCommit(repo, dir, toC)
	if err != nil {
		return nil, err
	}

	return reduceToDiff(m, diff)
}

func (m *Manifest) indexByName() map[string]*Application {
	return m.Applications.indexByName()
}

func (m *Manifest) indexByPath() map[string]*Application {
	return m.Applications.indexByPath()
}

func fromCommit(repo *git.Repository, dir string, commit *git.Commit) (*Manifest, error) {
	metadataSet, err := discoverMetadata(repo, commit)
	if err != nil {
		return nil, err
	}

	vapps, err := metadataSet.toApplications(true)
	if err != nil {
		return nil, err
	}

	sort.Sort(vapps)
	return &Manifest{dir, commit.Id().String(), vapps}, nil
}

func newEmptyManifest(dir string) *Manifest {
	return &Manifest{Applications: []*Application{}, Dir: dir, Sha: ""}
}

func fromBranch(repo *git.Repository, dir string, branch string) (*Manifest, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}

	return fromCommit(repo, dir, commit)
}

func reduceToDiff(manifest *Manifest, diff *git.Diff) (*Manifest, error) {
	q := manifest.indexByPath()
	filtered := make(map[string]*Application)
	err := diff.ForEach(func(delta git.DiffDelta, num float64) (git.DiffForEachHunkCallback, error) {
		for k := range q {
			if _, ok := filtered[k]; ok {
				continue
			}
			if strings.HasPrefix(delta.NewFile.Path, k) {
				filtered[k] = q[k]
			}
		}
		return nil, nil
	}, git.DiffDetailFiles)

	if err != nil {
		return nil, err
	}

	apps := Applications{}
	for _, v := range filtered {
		apps = append(apps, v)
	}

	expandedApps, err := apps.expandRequiredByDependencies()
	if err != nil {
		return nil, err
	}

	return &Manifest{
		Dir:          manifest.Dir,
		Sha:          manifest.Sha,
		Applications: expandedApps,
	}, nil
}

func openRepo(dir string) (*git.Repository, *Manifest, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, nil, err
	}
	empty, err := repo.IsEmpty()
	if err != nil {
		return nil, nil, err
	}

	if empty {
		return nil, newEmptyManifest(dir), nil
	}

	return repo, nil, nil
}
