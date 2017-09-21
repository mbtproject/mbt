package lib

import (
	"fmt"
	"strings"

	git "github.com/libgit2/git2go"
)

type VersionedApplication struct {
	Application *Application
	Version     string
}

type Manifest struct {
	Dir          string
	Sha          string
	Applications []*VersionedApplication
}

func ResolveChanges(path string) ([]string, error) {
	repo, _ := git.OpenRepository(path)
	head, _ := repo.Head()
	if head != nil {
		println("head is found")
	}
	return nil, nil
}

func fromCommit(repo *git.Repository, dir string, commit *git.Commit) (*Manifest, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	vapps := []*VersionedApplication{}

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Name == "appspec.yaml" && entry.Type == git.ObjectBlob {
			blob, err := repo.LookupBlob(entry.Id)
			if err != nil {
				return 1
			}

			p := strings.TrimRight(path, "/")
			a, err := NewApplication(p, blob.Contents())
			if err != nil {
				return 1
			}

			vapps = append(vapps, &VersionedApplication{
				Application: a,
				Version:     entry.Id.String(),
			})
		}
		return 0
	})

	if err != nil {
		return nil, err
	}

	return &Manifest{dir, commit.Id().String(), vapps}, nil
}

func getBranchCommit(repo *git.Repository, branch string) (*git.Commit, error) {
	ref, err := repo.References.Dwim(branch)
	if err != nil {
		return nil, err
	}

	oid := ref.Target()
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func getBranchTree(repo *git.Repository, branch string) (*git.Tree, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func fromBranch(repo *git.Repository, dir string, branch string) (*Manifest, error) {
	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return nil, err
	}

	return fromCommit(repo, dir, commit)
}

func ManifestByBranch(dir, branch string) (*Manifest, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, err
	}

	return fromBranch(repo, dir, branch)
}

func reduceToDiff(manifest *Manifest, diff *git.Diff) (*Manifest, error) {
	q := make(map[string]*VersionedApplication)
	for _, a := range manifest.Applications {
		q[fmt.Sprintf("%s/", a.Application.Path)] = a
	}

	filtered := make(map[string]*VersionedApplication)
	err := diff.ForEach(func(delta git.DiffDelta, num float64) (git.DiffForEachHunkCallback, error) {
		for k, _ := range q {
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

	apps := []*VersionedApplication{}
	for _, v := range filtered {
		apps = append(apps, v)
	}

	return &Manifest{
		Dir:          manifest.Dir,
		Sha:          manifest.Sha,
		Applications: apps,
	}, nil
}

func ManifestByPr(dir, from, to string) (*Manifest, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, err
	}

	m, err := fromBranch(repo, dir, from)
	if err != nil {
		return nil, err
	}

	fromTree, err := getBranchTree(repo, from)
	if err != nil {
		return nil, err
	}

	toTree, err := getBranchTree(repo, to)
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(toTree, fromTree, &git.DiffOptions{})
	if err != nil {
		return nil, err
	}

	return reduceToDiff(m, diff)
}
