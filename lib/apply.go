package lib

import (
	"io"
	"os"
	"strings"
	"text/template"

	git "github.com/libgit2/git2go"
)

type TemplateData struct {
	Args         map[string]interface{}
	Sha          string
	Env          map[string]string
	Applications map[string]*Application
}

func ApplyBranch(dir, templatePath, branch, output string) error {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return err
	}

	m, err := fromBranch(repo, dir, branch)
	if err != nil {
		return err
	}

	t, err := getBranchTree(repo, branch)
	if err != nil {
		return err
	}

	e, err := t.EntryByPath(templatePath)
	if err != nil {
		return err
	}

	b, err := repo.LookupBlob(e.Id)
	if err != nil {
		return err
	}

	temp, err := template.New("template").Parse(string(b.Contents()))
	if err != nil {
		return err
	}

	var writer io.Writer = os.Stdout
	if output != "" {
		writer, err = os.Create(output)
		if err != nil {
			return err
		}
	}

	data := &TemplateData{
		Sha:          m.Sha,
		Env:          getEnvMap(),
		Applications: m.indexByName(),
	}

	return temp.Execute(writer, data)
}

func getEnvMap() map[string]string {
	m := make(map[string]string)

	for _, v := range os.Environ() {
		p := strings.Split(v, "=")
		m[p[0]] = p[1]
	}

	return m
}
