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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	// Create indirect dependency graph
	check(t, repo.InitModule("lib-a"))

	check(t, repo.InitModuleWithOptions("lib-b", &Spec{
		Name:         "lib-b",
		Dependencies: []string{"lib-a"},
	}))

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name:         "app-a",
		Dependencies: []string{"lib-b"},
	}))

	// Create an isolated node
	check(t, repo.InitModule("app-b"))

	check(t, repo.Commit("first"))

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	s := m.Modules.SerializeAsDot()

	assert.Equal(t, `digraph mbt {
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  "lib-a"
  "lib-b" -> "lib-a"
  "app-a" -> "lib-b"
  "app-b"
}`, s)
}

func TestSerializeAsDotOfAnEmptyRepo(t *testing.T) {
	clean()
	_, err := createTestRepository(".tmp/repo")
	check(t, err)

	m, err := NewWorld(t, ".tmp/repo").ManifestBuilder.ByCurrentBranch()
	check(t, err)

	s := m.Modules.SerializeAsDot()

	assert.Equal(t, `digraph mbt {
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  
}`, s)
}
