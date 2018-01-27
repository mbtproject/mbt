package lib

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
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

// ApplyHead applies the repository manifest to specified template.
func ApplyHead(dir, templatePath string, output io.Writer) error {
	repo, err := git.OpenRepository(dir)
	if err != nil {
		return wrap(err)
	}

	commit, err := getHeadCommit(repo)
	if err != nil {
		return err
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

	m, err := ManifestByLocalDir(absDir, true)
	if err != nil {
		return err
	}

	return processTemplate(c, m, output)
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

	m, err := fromCommit(repo, dir, commit)
	if err != nil {
		return err
	}

	return processTemplate(b.Contents(), m, output)
}

func processTemplate(buffer []byte, m *Manifest, output io.Writer) error {
	modulesIndex := m.Modules.indexByName()
	temp, err := template.New("template").Funcs(template.FuncMap{
		"module": func(n string) *Module {
			return modulesIndex[n]
		},
		"property": func(m *Module, n string) interface{} {
			if m == nil {
				return nil
			}

			return resolveProperty(m.Properties(), strings.Split(n, "."))
		},
		"contains": func(container interface{}, item string) bool {
			if container == nil {
				return false
			}
			a, ok := container.([]interface{})
			if ok {
				if reflect.TypeOf(a).Elem().AssignableTo(reflect.TypeOf(item)) {
					return false
				}
			} else {
				return false
			}

			for _, w := range a {
				if w == item {
					return true
				}
			}
			return false
		},
	}).Parse(string(buffer))
	if err != nil {
		return wrap(err)
	}

	data := &TemplateData{
		Sha:     m.Sha,
		Env:     getEnvMap(),
		Modules: m.indexByName(),
	}

	return temp.Execute(output, data)
}

func resolveProperty(in interface{}, path []string) interface{} {
	if in == nil || len(path) == 0 {
		return nil
	}
	if currentMap, ok := in.(map[string]interface{}); ok {
		v, ok := currentMap[path[0]]
		if !ok {
			return nil
		}
		rest := path[1:]
		if len(rest) > 0 {
			return resolveProperty(v, rest)
		}
		return v
	}
	return nil
}

func getEnvMap() map[string]string {
	m := make(map[string]string)

	for _, v := range os.Environ() {
		p := strings.Split(v, "=")
		m[p[0]] = p[1]
	}

	return m
}
