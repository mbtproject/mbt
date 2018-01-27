package lib

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	check(t, ApplyBranch(".tmp/repo", "template.tmpl", "feature", output))

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
	check(t, ApplyCommit(".tmp/repo", repo.LastCommit.String(), "template.tmpl", output))

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
	check(t, ApplyHead(".tmp/repo", "template.tmpl", output))

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
	check(t, ApplyLocal(".tmp/repo", "template.tmpl", output))

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
	err = ApplyCommit(".tmp/repo", repo.LastCommit.String(), "foo/template.tmpl", output)

	assert.EqualError(t, err, "mbt: the path 'foo' does not exist in the given tree")
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
	err = ApplyCommit(".tmp/repo", repo.LastCommit.String(), "template.tmpl", output)

	assert.EqualError(t, err, "mbt: template: template:2: unexpected EOF")
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
	check(t, ApplyCommit(".tmp/repo", repo.LastCommit.String(), "template.tmpl", output))

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
			"foo": "bar",
		},
	}))

	check(t, repo.WriteContent("template.tmpl", `
{{- if contains (property (module "app-a") "tags") "a"}}foo{{- end}}
{{- if contains (property (module "app-b") "tags") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "dags") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "tags") "d"}}foo{{- end}}
{{- if contains (property (module "app-a") "numbers") "1"}}foo{{- end}}
{{- if contains (property (module "app-a") "nested.tags") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "nested.bags") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "tags.tags") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "nested.") "a"}}foo{{- end}}
{{- if contains (property (module "app-a") "nested") "a"}}foo{{- end}}
{{- propertyOr (module "app-a") "foo" "car"}}
{{- propertyOr (module "app-a") "foo.bar" "foo"}}
`))
	check(t, repo.Commit("first"))

	output := new(bytes.Buffer)
	err = ApplyCommit(".tmp/repo", repo.LastCommit.String(), "template.tmpl", output)
	check(t, err)

	assert.Equal(t, "foofoobarfoo\n", output.String())
}
