package lib

import (
	"encoding/hex"
	"fmt"
	"strings"

	git "github.com/libgit2/git2go"
	yaml "gopkg.in/yaml.v2"
)

type Application struct {
	Name           string
	Path           string
	Args           []string `yaml:",flow"`
	BuildPlatforms []string `yaml:"buildPlatforms,flow"`
	Build          string
	Version        string
	Properties     map[string]interface{}
}

type Applications []*Application

type Manifest struct {
	Dir          string
	Sha          string
	Applications Applications
}

type TemplateData struct {
	Args         map[string]interface{}
	Sha          string
	Applications map[string]*Application
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

	base, err := repo.MergeBase(srcC.Id(), dstC.Id())
	if err != nil {
		return nil, err
	}

	baseC, err := repo.LookupCommit(base)
	if err != nil {
		return nil, err
	}

	baseTree, err := baseC.Tree()
	if err != nil {
		return nil, err
	}

	srcTree, err := getBranchTree(repo, src)
	if err != nil {
		return nil, err
	}

	diff, err := repo.DiffTreeToTree(baseTree, srcTree, &git.DiffOptions{})
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

// Sort interface to sort applications by path
func (a Applications) Len() int {
	return len(a)
}

func (a Applications) Less(i, j int) bool {
	return a[i].Path < a[j].Path
}

func (a Applications) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (m *Manifest) indexByName() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range m.Applications {
		q[a.Name] = a
	}
	return q
}

func (m *Manifest) indexByPath() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range m.Applications {
		q[fmt.Sprintf("%s/", a.Path)] = a
	}
	return q
}

func fromCommit(repo *git.Repository, dir string, commit *git.Commit) (*Manifest, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	vapps := []*Application{}

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Name == "appspec.yaml" && entry.Type == git.ObjectBlob {
			blob, err := repo.LookupBlob(entry.Id)
			if err != nil {
				return 1
			}

			p := strings.TrimRight(path, "/")
			dirEntry, err := tree.EntryByPath(p)
			if err != nil {
				return 1
			}

			a, err := newApplication(p, dirEntry.Id.String(), blob.Contents())
			if err != nil {
				// TODO log this or fail
				return 1
			}

			vapps = append(vapps, a)
		}
		return 0
	})

	if err != nil {
		return nil, err
	}

	return &Manifest{dir, commit.Id().String(), vapps}, nil
}

func newApplication(dir, version string, spec []byte) (*Application, error) {
	a := &Application{
		Properties: make(map[string]interface{}),
		Args:       make([]string, 0),
	}

	err := yaml.Unmarshal(spec, a)
	if err != nil {
		return nil, err
	}

	a.Path = dir
	a.Version = version
	return a, nil
}

func newEmptyManifest(dir string) *Manifest {
	return &Manifest{Applications: []*Application{}, Dir: dir, Sha: ""}
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

func reduceToDiff(manifest *Manifest, diff *git.Diff) (*Manifest, error) {
	q := manifest.indexByPath()
	filtered := make(map[string]*Application)
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

	apps := []*Application{}
	for _, v := range filtered {
		apps = append(apps, v)
	}

	return &Manifest{
		Dir:          manifest.Dir,
		Sha:          manifest.Sha,
		Applications: apps,
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
