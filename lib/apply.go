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
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/mbtproject/mbt/e"
)

// TemplateData is the data passed into template.
type TemplateData struct {
	Args        map[string]interface{}
	Sha         string
	Env         map[string]string
	Modules     map[string]*Module
	ModulesList []*Module
}

// KVP is a key value pair.
type KVP struct {
	Key   string
	Value interface{}
}

type kvpSoter []*KVP

func (a kvpSoter) Len() int {
	return len(a)
}

func (a kvpSoter) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a kvpSoter) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

type modulesByNameSorter []*Module

func (m modulesByNameSorter) Len() int {
	return len(m)
}

func (m modulesByNameSorter) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m modulesByNameSorter) Less(i, j int) bool {
	return m[i].Name() < m[j].Name()
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
	sortedModules := make(modulesByNameSorter, len(m.Modules))
	copy(sortedModules, m.Modules)
	sort.Sort(sortedModules)

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
		"kvplist": func(m map[string]interface{}) []*KVP {
			l := make([]*KVP, 0, len(m))
			for k, v := range m {
				l = append(l, &KVP{Key: k, Value: v})
			}

			sort.Sort(kvpSoter(l))
			return l
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			return a / b
		},
		"istail": func(array, value interface{}) bool {
			if array == nil {
				return false
			}

			t := reflect.TypeOf(array)
			if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
				panic(fmt.Sprintf("istail function requires an array/slice as input %d", t.Kind()))
			}

			arrayVal := reflect.ValueOf(array)
			if arrayVal.Len() == 0 {
				return false
			}

			v := arrayVal.Index(arrayVal.Len() - 1).Interface()
			return v == value
		},
		"ishead": func(array, value interface{}) bool {
			if array == nil {
				return false
			}

			t := reflect.TypeOf(array)
			if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
				panic(fmt.Sprintf("ishead function requires an array/slice as input %d", t.Kind()))
			}

			arrayVal := reflect.ValueOf(array)
			if arrayVal.Len() == 0 {
				return false
			}

			v := arrayVal.Index(0).Interface()
			return v == value
		},
		"head": func(array interface{}) interface{} {
			if array == nil {
				return ""
			}

			t := reflect.TypeOf(array)
			if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
				panic(fmt.Sprintf("head function requires an array/slice as input %d", t.Kind()))
			}

			arrayVal := reflect.ValueOf(array)
			if arrayVal.Len() == 0 {
				return ""
			}

			return arrayVal.Index(0).Interface()
		},
		"tail": func(array interface{}) interface{} {
			if array == nil {
				return ""
			}

			t := reflect.TypeOf(array)
			if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
				panic(fmt.Sprintf("tail function requires an array/slice as input %d", t.Kind()))
			}

			arrayVal := reflect.ValueOf(array)
			if arrayVal.Len() == 0 {
				return ""
			}

			return arrayVal.Index(arrayVal.Len() - 1).Interface()
		},
	}).Parse(string(buffer))
	if err != nil {
		return e.Wrapf(ErrClassUser, err, msgFailedTemplateParse)
	}

	data := &TemplateData{
		Sha:         m.Sha,
		Env:         getEnvMap(),
		Modules:     m.Modules.indexByName(),
		ModulesList: sortedModules,
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
