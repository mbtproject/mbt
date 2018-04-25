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

func TestIntersectionWithElements(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))

	second := repo.LastCommit

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-a/bar", "hello"))
	check(t, repo.WriteContent("app-b/foo", "hello"))
	check(t, repo.Commit("third"))

	third := repo.LastCommit

	mods, err := NewWorld(t, ".tmp/repo").System.IntersectionByCommit(second.String(), third.String())
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())

	// This operation should be commutative
	mods, err = NewWorld(t, ".tmp/repo").System.IntersectionByCommit(third.String(), second.String())
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())
}

func TestIntersectionWithoutElements(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))

	second := repo.LastCommit

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-b/foo", "hello"))
	check(t, repo.Commit("third"))

	third := repo.LastCommit

	mods, err := NewWorld(t, ".tmp/repo").System.IntersectionByCommit(second.String(), third.String())
	check(t, err)

	assert.Len(t, mods, 0)

	// This operation should be commutative
	mods, err = NewWorld(t, ".tmp/repo").System.IntersectionByCommit(third.String(), second.String())
	check(t, err)

	assert.Len(t, mods, 0)
}

func TestIntersectionByBranchWithElements(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModule("app-a"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-a/bar", "hello"))
	check(t, repo.WriteContent("app-b/foo", "hello"))
	check(t, repo.Commit("third"))

	mods, err := NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())

	// This operation should be commutative
	mods, err = NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())
}

func TestIntersectionWithDependencies(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{Name: "app-a", Dependencies: []string{"app-c"}}))
	check(t, repo.InitModule("app-b"))
	check(t, repo.InitModuleWithOptions("app-c", &Spec{Name: "app-c"}))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-c/foo", "hello"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-a/bar", "hello"))
	check(t, repo.Commit("third"))

	mods, err := NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-c", mods[0].Name())

	// This operation should be commutative
	mods, err = NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-c", mods[0].Name())
}

func TestIntersctionOfTwoChangesWithSharedDependency(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")

	check(t, repo.InitModuleWithOptions("app-a", &Spec{Name: "app-a", Dependencies: []string{"app-c"}}))
	check(t, repo.InitModuleWithOptions("app-b", &Spec{Name: "app-b", Dependencies: []string{"app-c"}}))
	check(t, repo.InitModuleWithOptions("app-c", &Spec{Name: "app-c"}))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-b/bar", "hello"))
	check(t, repo.Commit("third"))

	mods, err := NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 0)

	// This operation should be commutative
	mods, err = NewWorld(t, ".tmp/repo").System.IntersectionByBranch("feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 0)
}
