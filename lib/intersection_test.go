package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntersectionWithElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	mods, err := IntersectionByCommit(".tmp/repo", second.String(), third.String())
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())

	// This operation should be commutative
	mods, err = IntersectionByCommit(".tmp/repo", third.String(), second.String())
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())
}

func TestIntersectionWithoutElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	mods, err := IntersectionByCommit(".tmp/repo", second.String(), third.String())
	check(t, err)

	assert.Len(t, mods, 0)

	// This operation should be commutative
	mods, err = IntersectionByCommit(".tmp/repo", third.String(), second.String())
	check(t, err)

	assert.Len(t, mods, 0)
}

func TestIntersectionByBranchWithElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	mods, err := IntersectionByBranch(".tmp/repo", "feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())

	// This operation should be commutative
	mods, err = IntersectionByBranch(".tmp/repo", "feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-a", mods[0].Name())
}

func TestIntersectionWithDependencies(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	mods, err := IntersectionByBranch(".tmp/repo", "feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-c", mods[0].Name())

	// This operation should be commutative
	mods, err = IntersectionByBranch(".tmp/repo", "feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 1)
	assert.Equal(t, "app-c", mods[0].Name())
}

func TestIntersctionOfTwoChangesWithSharedDependency(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

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

	mods, err := IntersectionByBranch(".tmp/repo", "feature-a", "feature-b")
	check(t, err)

	assert.Len(t, mods, 0)

	// This operation should be commutative
	mods, err = IntersectionByBranch(".tmp/repo", "feature-b", "feature-a")
	check(t, err)

	assert.Len(t, mods, 0)
}
