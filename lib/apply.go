package lib

import (
	"io"
	"os"
	"text/template"

	git "github.com/libgit2/git2go"
)

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
		Applications: m.IndexByName(),
	}

	return temp.Execute(writer, data)
}
