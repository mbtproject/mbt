/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mbtproject/mbt/e"
)

// TemplateData is the data passed into template.
type TemplateData struct {
	Args    map[string]interface{}
	Sha     string
	Env     map[string]string
	Modules map[string]*Module
}

func (s *stdSystem) ApplyBranch(templatePath, branch string, output io.Writer) error {
	commit, err := s.Repo.BranchCommit(branch)
	if err != nil {
		return err
	}
	return s.applyCore(commit, templatePath, output)
}

func (s *stdSystem) ApplyCommit(commit string, templatePath string, output io.Writer) error {
	c, err := s.Repo.GetCommit(commit)
	if err != nil {
		return err
	}
	return s.applyCore(c, templatePath, output)
}

// ApplyHead applies the repository manifest to specified template.
func (s *stdSystem) ApplyHead(templatePath string, output io.Writer) error {
	branch, err := s.Repo.CurrentBranch()
	if err != nil {
		return err
	}

	return s.ApplyBranch(templatePath, branch, output)
}

func (s *stdSystem) ApplyLocal(templatePath string, output io.Writer) error {
	absDir, err := filepath.Abs(s.Repo.Path())
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedLocalPath)
	}

	absTemplatePath := filepath.Join(absDir, templatePath)
	c, err := ioutil.ReadFile(absTemplatePath)
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedReadFile, absTemplatePath)
	}

	m, err := s.ManifestBuilder().ByWorkspace()
	if err != nil {
		return err
	}

	return processTemplate(c, m, output)
}

func (s *stdSystem) applyCore(commit Commit, templatePath string, output io.Writer) error {
	b, err := s.Repo.BlobContentsFromTree(commit, templatePath)
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgTemplateNotFound, templatePath, commit)
	}

	m, err := s.MB.ByCommit(commit)
	if err != nil {
		return err
	}

	return processTemplate(b, m, output)
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

			return resolveProperty(m.Properties(), strings.Split(n, "."), nil)
		},
		"propertyOr": func(m *Module, n string, def interface{}) interface{} {
			if m == nil {
				return def
			}

			return resolveProperty(m.Properties(), strings.Split(n, "."), def)
		},
		"contains": func(container interface{}, item interface{}) bool {
			if container == nil {
				return false
			}

			a, ok := container.([]interface{})
			if !ok {
				return false
			}

			for _, w := range a {
				if w == item {
					return true
				}
			}

			return false
		},
		"join": func(container interface{}, format string, sep string) string {
			if container == nil {
				return ""
			}

			a, ok := container.([]interface{})
			if !ok {
				return ""
			}

			strs := make([]string, 0, len(a))
			for _, i := range a {
				strs = append(strs, fmt.Sprintf(format, i))
			}

			return strings.Join(strs, sep)
		},
	}).Parse(string(buffer))
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedTemplateParse)
	}

	data := &TemplateData{
		Sha:     m.Sha,
		Env:     getEnvMap(),
		Modules: m.Modules.indexByName(),
	}

	return temp.Execute(output, data)
}

func resolveProperty(in interface{}, path []string, def interface{}) interface{} {
	if in == nil || len(path) == 0 {
		return def
	}
	if currentMap, ok := in.(map[string]interface{}); ok {
		v, ok := currentMap[path[0]]
		if !ok {
			return def
		}
		rest := path[1:]
		if len(rest) > 0 {
			return resolveProperty(v, rest, def)
		}
		return v
	}
	return def
}

func getEnvMap() map[string]string {
	m := make(map[string]string)

	for _, v := range os.Environ() {
		p := strings.Split(v, "=")
		m[p[0]] = p[1]
	}

	return m
}
