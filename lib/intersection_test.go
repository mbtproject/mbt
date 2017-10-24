package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntersectionWithElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.InitApplication("app-b"))
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

	apps, err := IntersectionByCommit(".tmp/repo", second.String(), third.String())
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-a", apps[0].Name())

	// This operation should be commutative
	apps, err = IntersectionByCommit(".tmp/repo", third.String(), second.String())
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-a", apps[0].Name())
}

func TestIntersectionWithoutElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.InitApplication("app-b"))
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

	apps, err := IntersectionByCommit(".tmp/repo", second.String(), third.String())
	check(t, err)

	assert.Len(t, apps, 0)

	// This operation should be commutative
	apps, err = IntersectionByCommit(".tmp/repo", third.String(), second.String())
	check(t, err)

	assert.Len(t, apps, 0)
}

func TestIntersectionByBranchWithElements(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplication("app-a"))
	check(t, repo.InitApplication("app-b"))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-a/foo", "hello"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-a/bar", "hello"))
	check(t, repo.WriteContent("app-b/foo", "hello"))
	check(t, repo.Commit("third"))

	apps, err := IntersectionByBranch(".tmp/repo", "feature-a", "feature-b")
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-a", apps[0].Name())

	// This operation should be commutative
	apps, err = IntersectionByBranch(".tmp/repo", "feature-b", "feature-a")
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-a", apps[0].Name())
}

func TestIntersectionWithDependencies(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitApplicationWithOptions("app-a", &Spec{Name: "app-a", Dependencies: []string{"app-c"}}))
	check(t, repo.InitApplication("app-b"))
	check(t, repo.InitApplicationWithOptions("app-c", &Spec{Name: "app-c"}))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature-a"))
	check(t, repo.WriteContent("app-c/foo", "hello"))
	check(t, repo.Commit("second"))

	check(t, repo.SwitchToBranch("master"))
	check(t, repo.SwitchToBranch("feature-b"))
	check(t, repo.WriteContent("app-a/bar", "hello"))
	check(t, repo.Commit("third"))

	apps, err := IntersectionByBranch(".tmp/repo", "feature-a", "feature-b")
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-c", apps[0].Name())

	// This operation should be commutative
	apps, err = IntersectionByBranch(".tmp/repo", "feature-b", "feature-a")
	check(t, err)

	assert.Len(t, apps, 1)
	assert.Equal(t, "app-c", apps[0].Name())
}
