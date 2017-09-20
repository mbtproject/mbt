package lib

import git "github.com/libgit2/git2go"

type EnvManifest struct {
	Name    string
	Version string
}

func GetEnvManifestByBranch(path string, branch string) ([]*EnvManifest, error) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		return nil, err
	}

	_, err = repo.References.Dwim(branch)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
