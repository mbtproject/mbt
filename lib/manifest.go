package lib

import git "github.com/libgit2/git2go"

type VersionedApplication struct {
	Application *Application
	Version     string
}

type Manifest struct {
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

func ManifestByBranch(dir, branch string) (*Manifest, error) {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return nil, err
	}

	ref, err := repo.References.Dwim(branch)
	if err != nil {
		return nil, err
	}

	oid := ref.Target()
	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	apps, err := Discover(dir)
	if err != nil {
		return nil, err
	}

	vapps := []*VersionedApplication{}

	for _, a := range apps {
		v, err := tree.EntryByPath(a.Path)
		if err != nil {
			return nil, err
		}
		vapps = append(vapps, &VersionedApplication{a, v.Id.String()})
	}

	return &Manifest{oid.String(), vapps}, nil
}
