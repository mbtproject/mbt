package lib

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	git "github.com/libgit2/git2go"
)

// TemplateData is the data passed into template.
type TemplateData struct {
	Args    map[string]interface{}
	Sha     string
	Env     map[string]string
	Modules map[string]*Module
}

// ApplyBranch applies the repository manifest to specified template.
func ApplyBranch(dir, templatePath, branch string, output io.Writer) error {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return wrap(err)
	}

	commit, err := getBranchCommit(repo, branch)
	if err != nil {
		return err
	}

	return applyCore(repo, commit, dir, templatePath, output)
}

// ApplyCommit applies the repository manifest to specified template.
func ApplyCommit(dir, sha, templatePath string, output io.Writer) error {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return wrap(err)
	}

	shaOid, err := git.NewOid(sha)
	if err != nil {
		return wrap(err)
	}

	commit, err := repo.LookupCommit(shaOid)
	if err != nil {
		return wrap(err)
	}

	return applyCore(repo, commit, dir, templatePath, output)
}

// ApplyLocal applies local directory manifest over an specified template
func ApplyLocal(dir, templatePath string, output io.Writer) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return wrap(err)
	}

	c, err := ioutil.ReadFile(path.Join(absDir, templatePath))
	if err != nil {
		return wrap(err)
	}

	temp, err := template.New("template").Parse(string(c))
	if err != nil {
		return wrap(err)
	}

	m, err := ManifestByLocalDir(absDir, true)
	if err != nil {
		return err
	}

	data := &TemplateData{
		Sha:     m.Sha,
		Env:     getEnvMap(),
		Modules: m.indexByName(),
	}

	return temp.Execute(output, data)
}

func applyCore(repo *git.Repository, commit *git.Commit, dir, templatePath string, output io.Writer) error {
	tree, err := commit.Tree()
	if err != nil {
		return wrap(err)
	}

	e, err := tree.EntryByPath(templatePath)
	if err != nil {
		return wrap(err)
	}

	b, err := repo.LookupBlob(e.Id)
	if err != nil {
		return wrap(err)
	}

	temp, err := template.New("template").Parse(string(b.Contents()))
	if err != nil {
		return wrap(err)
	}

	m, err := fromCommit(repo, dir, commit)
	if err != nil {
		return err
	}

	data := &TemplateData{
		Sha:     m.Sha,
		Env:     getEnvMap(),
		Modules: m.indexByName(),
	}

	return temp.Execute(output, data)
}

func getEnvMap() map[string]string {
	m := make(map[string]string)

	for _, v := range os.Environ() {
		p := strings.Split(v, "=")
		m[p[0]] = p[1]
	}

	return m
}
