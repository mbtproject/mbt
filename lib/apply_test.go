package lib

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

type TC struct {
	Template string
	Expected string
}

func TestApplyBranch(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
{{- $mod.Name }},
{{- end}}
`))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	output := new(bytes.Buffer)
	check(t, NewWorld(t, ".tmp/repo").System.ApplyBranch("template.tmpl", "feature", output))

	assert.Equal(t, "app-a,app-b,\n", output.String())
}

func TestApplyCommit(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
{{- $mod.Name }},
{{- end}}
`))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	output := new(bytes.Buffer)
	check(t, NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "template.tmpl", output))

	assert.Equal(t, "app-a,app-b,\n", output.String())
}

func TestApplyHead(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
{{- $mod.Name }},
{{- end}}
`))
	check(t, repo.Commit("first"))

	check(t, repo.SwitchToBranch("feature"))
	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	output := new(bytes.Buffer)
	check(t, NewWorld(t, ".tmp/repo").System.ApplyHead("template.tmpl", output))

	assert.Equal(t, "app-a,app-b,\n", output.String())
}

func TestApplyLocal(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
{{- $mod.Name }},
{{- end}}
`))
	check(t, repo.Commit("first"))

	check(t, repo.InitModule("app-b"))
	check(t, repo.Commit("second"))

	output := new(bytes.Buffer)
	check(t, NewWorld(t, ".tmp/repo").System.ApplyLocal("template.tmpl", output))

	assert.Equal(t, "app-a,app-b,\n", output.String())
}

func TestIncorrectTemplatePath(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
{{- $mod.Name }},
{{- end}}
`))
	check(t, repo.Commit("first"))

	output := new(bytes.Buffer)
	err = NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "foo/template.tmpl", output)

	assert.EqualError(t, err, fmt.Sprintf(msgTemplateNotFound, "foo/template.tmpl", repo.LastCommit.String()))
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
	assert.Equal(t, "", output.String())
}

func TestBadTemplate(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", `
{{- range $i, $mod := .Modules}}
`))
	check(t, repo.Commit("first"))

	output := new(bytes.Buffer)
	err = NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "template.tmpl", output)

	assert.EqualError(t, err, msgFailedTemplateParse)
	assert.EqualError(t, (err.(*e.E)).InnerError(), "template: template:2: unexpected EOF")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
	assert.Equal(t, "", output.String())
}

func TestEnvironmentVariables(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModule("app-a"))
	check(t, repo.WriteContent("template.tmpl", "{{.Env.EXTERNAL_VALUE}}"))
	check(t, repo.Commit("first"))

	os.Setenv("EXTERNAL_VALUE", "FOO")

	output := new(bytes.Buffer)
	check(t, NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "template.tmpl", output))

	assert.Equal(t, "FOO", output.String())
}

func TestCustomTemplateFuncs(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModuleWithOptions("app-a", &Spec{
		Name: "app-a",
		Properties: map[string]interface{}{
			"tags":    []string{"a", "b", "c"},
			"numbers": []int{1, 2, 3},
			"nested": map[string]interface{}{
				"tags": []string{"a", "b", "c"},
			},
			"empty": []int{},
			"foo":   "bar",
		},
	}))

	cases := []TC{
		{Template: `{{- if contains (property (module "app-a") "tags") "a"}}yes{{- end}}`, Expected: "yes"},
		{Template: `{{- if contains (property (module "app-b") "tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "dags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "tags") "d"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "numbers") "1"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "numbers") 1}}yes{{- end}}`, Expected: "yes"},
		{Template: `{{- if contains (property (module "app-a") "empty") 1}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested.tags") "a"}}yes{{- end}}`, Expected: "yes"},
		{Template: `{{- if contains (property (module "app-a") "nested.bags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "tags.tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested.") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- propertyOr (module "app-a") "foo" "car"}}`, Expected: "bar"},
		{Template: `{{- propertyOr (module "app-a") "foo.bar" "car"}}`, Expected: "car"},
		{Template: `{{- propertyOr (module "app-b") "foo" "car"}}`, Expected: "car"},
		{Template: `{{- join (property (module "app-a") "tags") "%v" "-"}}`, Expected: "a-b-c"},
		{Template: `{{- join (property (module "app-b") "tags") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "numbers") "%v" "-"}}`, Expected: "1-2-3"},
		{Template: `{{- join (property (module "app-a") "empty") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "bar") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "foo") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "") "%v" "-"}}`, Expected: ""},
	}

	for _, c := range cases {
		check(t, repo.WriteContent("template.tmpl", c.Template))
		check(t, repo.Commit("Update"))

		output := new(bytes.Buffer)
		err = NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "template.tmpl", output)
		check(t, err)

		assert.Equal(t, c.Expected, output.String(), "Failed test case %s", c.Template)
	}
}

func TestCustomTemplateFuncsForModulesWithoutProperties(t *testing.T) {
	clean()
	repo, err := createTestRepository(".tmp/repo")
	check(t, err)

	check(t, repo.InitModuleWithOptions("app-a", &Spec{Name: "app-a"}))

	cases := []TC{
		{Template: `{{- if contains (property (module "app-a") "tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-b") "tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "dags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "tags") "d"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "numbers") "1"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "numbers") 1}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "empty") 1}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested.tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested.bags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "tags.tags") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested.") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- if contains (property (module "app-a") "nested") "a"}}yes{{- end}}`, Expected: ""},
		{Template: `{{- propertyOr (module "app-a") "foo" "car"}}`, Expected: "car"},
		{Template: `{{- propertyOr (module "app-a") "foo.bar" "car"}}`, Expected: "car"},
		{Template: `{{- propertyOr (module "app-b") "foo" "car"}}`, Expected: "car"},
		{Template: `{{- join (property (module "app-a") "tags") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-b") "tags") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "numbers") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "empty") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "bar") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "foo") "%v" "-"}}`, Expected: ""},
		{Template: `{{- join (property (module "app-a") "") "%v" "-"}}`, Expected: ""},
	}

	for _, c := range cases {
		check(t, repo.WriteContent("template.tmpl", c.Template))
		check(t, repo.Commit("Update"))

		output := new(bytes.Buffer)
		err = NewWorld(t, ".tmp/repo").System.ApplyCommit(repo.LastCommit.String(), "template.tmpl", output)
		check(t, err)

		assert.Equal(t, c.Expected, output.String(), "Failed test case %s", c.Template)
	}
}

func TestApplyBranchForFailedBranchResolution(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("BranchCommit").Return(nil, errors.New("doh"))

	output := new(bytes.Buffer)
	assert.EqualError(t, w.System.ApplyBranch("template.tmpl", "master", output), "doh")
}

func TestApplyCommitForFailedCommitResolution(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("GetCommit").Return(nil, errors.New("doh"))

	output := new(bytes.Buffer)
	assert.EqualError(t, w.System.ApplyCommit("abc", "template.tmpl", output), "doh")
}

func TestApplyHeadForFailedBranchResolution(t *testing.T) {
	clean()
	NewTestRepo(t, ".tmp/repo")

	w := NewWorld(t, ".tmp/repo")
	w.Repo.Interceptor.Config("CurrentBranch").Return("", errors.New("doh"))

	output := new(bytes.Buffer)
	assert.EqualError(t, w.System.ApplyHead("template.tmpl", output), "doh")
}

func TestApplyLocalForWrongTemplatePath(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	repo.InitModule("app-a")

	w := NewWorld(t, ".tmp/repo")
	output := new(bytes.Buffer)
	err := w.System.ApplyLocal("template.tmpl", output)

	absTemplatePath, err := filepath.Abs(".tmp/repo/template.tmpl")
	check(t, err)

	err = w.System.ApplyLocal("template.tmpl", new(bytes.Buffer))
	assert.EqualError(t, err, fmt.Sprintf(msgFailedReadFile, absTemplatePath))
	assert.Error(t, (err.(*e.E)).InnerError())
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
}

func TestApplyLocalForManifestBuildFailure(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	check(t, repo.WriteContent("template.tmpl", "{{.Sha}}"))

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByWorkspace").Return(nil, errors.New("doh"))

	assert.EqualError(t, w.System.ApplyLocal("template.tmpl", new(bytes.Buffer)), "doh")
}

func TestApplyForManifestBuildFailureForCommit(t *testing.T) {
	clean()
	repo := NewTestRepo(t, ".tmp/repo")
	repo.InitModule("app-a")
	check(t, repo.WriteContent("template.tmpl", "{{.Sha}}"))
	repo.Commit("first")

	w := NewWorld(t, ".tmp/repo")
	w.ManifestBuilder.Interceptor.Config("ByCommit").Return(nil, errors.New("doh"))

	assert.EqualError(t, w.System.ApplyBranch("template.tmpl", "master", new(bytes.Buffer)), "doh")
}

func TestResolvePropertyForAnEmptySource(t *testing.T) {
	assert.Equal(t, "a", resolveProperty(nil, []string{"a"}, "a"))
}
