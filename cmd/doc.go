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

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

var templates = map[string]string{
	"main-summary": `mbt - The most flexible build orchestration tool for monorepo`,
	"main": `mbt is a build orchestration utility for monorepo.

In mbt, we use the term {{c "module"}} to identify an independently buildable unit of code.
A collection of modules stored in a repository is called a manifest.

You can turn any directory into a module by placing a spec file called {{c ".mbt.yml"}}.
Spec file is written in {{c "yaml" }} following the schema specified below.

{{c "" }}
name: Unique module name (required)
build: Dictionary of build commands specific to a platform (optional)
  default: (optional)
    cmd: Default command to run when os specific command is not found (required)
    args: Array of arguments to default build command (optional)
  linux|darwin|windows:
    cmd: Operating system specific command name (required)
    args: Array of arguments (optional)
dependencies: An array of modules that this module's build depend on (optional)
fileDependencies: An array of file names that this module's build depend on (optional)
commands: Optional dictionary of custom commands (optional)
  name: Custom command name (required)
  cmd: Command name (required)
  args: Array of arguments (optional)
	os: Array of os identifiers where this command should run (optional)
properties: Custom dictionary to hold any module specific information (optional)
{{c ""}}

{{h2 "Build Command"}}
Build command is operating system specific. When executing {{c "mbt build xxx" }}
commands, it skips the modules that do not specify a build command for the operating 
system build is running on.
Full list of operating systems names that can be used can be 
found {{link "here" "https://golang.org/doc/install/source#environment"}}

When the command is applicable for multiple operating systems, you could list it as
the default command. Operating system specific commands take precedence.

{{h2 "Dependencies"}}
{{ c "mbt"}} comes with a set of primitives to manage build dependencies. Current build
tools do a good job in managing dependencies between source files/projects.
GNU {{c "make" }} for example does a great job in deciding what parts require building.
{{c "mbt"}} dependencies work one level above that type of tooling to manage build
relationship between two independent modules.

{{h2 "Module Dependencies"}}
One advantage of a single repo is that you can share code between multiple modules
more effectively. Building this type of dependencies requires some thought. mbt provides
an easy way to define dependencies between modules and automatically builds the impacted modules
in topological order.

Dependencies are defined in {{c ".mbt.yml" }} file under 'dependencies' property.
It accepts an array of module names.
For example, {{ c "module-a" }} could define a dependency on {{c "module-b" }},
so that any time {{c "module-b"}} is changed, build command for {{c "module-a" }} is also executed.

An example of where this could be useful is, shared libraries. Shared library
could be developed independently of its consumers. However, all consumers
are automatically built whenever the shared library is modified.

{{h2 "File Dependencies"}}
File dependencies are useful in situations where a module should be built
when a file(s) stored outside the module directory is modified. For instance,
a build script that is used to build multiple modules could define a file
dependency in order to trigger the build whenever there's a change in build.

File dependencies should specify the path of the file relative to the root
of the repository.

{{h2 "Module Version"}}
For each module stored within a repository, {{c "mbt"}} generates a unique
stable version string. It is calculated based on three source attributes in
module.

- Content stored within the module directory
- Versions of dependent modules
- Content of file dependencies

As a result, version of a module is changed only when any of those attributes
are changed making it a safe attribute to use for tagging the 
build artifacts (i.e. tar balls, container images).

{{h2 "Document Generation"}}
{{ c "mbt" }} has a powerful feature that exposes the module state inferred from
the repository to a template engine. This could be quite useful for generating
documents out of that information. For example, you may have several modules
packaged as docker containers. By using this feature, you could generate a
Kubernetes deployment specification for all of them as of a particular commit
sha.

See {{c "apply"}} command for more details.

`,
	"apply-summary": `Apply repository manifest over a go template`,
	"apply": `{{cli "Apply repository manifest over a go template\n" }}
{{c "mbt apply branch [name] --to <path>"}}{{br}}
Apply the manifest of the tip of the branch to a template.
Template path should be relative to the repository root and must be available
in the commit tree. Assume master if name is not specified.

{{c "mbt apply commit <commit> --to <path>"}}{{br}}
Apply the manifest of a commit to a template.
Template path should be relative to the repository root and must be available
in the commit tree.

{{c "mbt apply head --to <path>"}}{{br}}
Apply the manifest of current head to a template.
Template path should be relative to the repository root and must be available
in the commit tree.

{{c "mbt apply local --to <path>"}}{{br}}
Apply the manifest of local workspace to a template.
Template path should be relative to the repository root and must be available
in the workspace.

{{h2 "Template Helpers"}}
Following helper functions are available when writing templates.

{{ c "module <name>" }}{{br}}
Find the specified module from modules set discovered in the repo.

{{ c "property <module> <name>" }}{{br}}
Find the specified property in the given module. Standard dot notation can be used to access nested properties (e.g. {{ c "a.b.c" }}).

{{ c "propertyOr <module> <name> <default>" }}{{br}}
Find specified property in the given module or return the designated default value.

{{ c "contains <array> <item>" }}{{br}}
Return true if the given item is present in the array.

{{ c "join <array> <format> <separator>" }}{{br}}
Format each item in the array according the format specified and then join them with the specified separator.

{{ c "kvplist <map>" }}{{br}}
Take a map of {{ c "map[string]interface{}" }} and return the items in the map as a list of key/value pairs sorted by the key. They can be accessed via {{ c ".Key" }} and {{ c ".Value" }} properties.

{{ c "add <int> <int>" }}{{br}}
Add two integers.

{{ c "sub <int> <int>" }}{{br}}
Subtract two integers.

{{ c "mul <int> <int>" }}{br}
Multiply two integers.

{{ c "div <int> <int>" }}{{br}}
Divide two integers.

{{ c "head <array|slice>" }}{{br}}
Select the first item in an array/slice. Return {{ c "\"\"" }} if the input is {{ c "nil" }} or an empty array/slice.

{{ c "tail <array|slice>" }}{{br}}
Select the last item in an array/slice. Return {{ c "\"\"" }} if the input is {{ c "nil" }} or an empty array/slice.

{{ c "ishead <array> <value>" }}{{br}}
Return true if the specified value is the head of the array/slice.

{{ c "istail <array> <value>" }}{{br}}
Return true if the specified value is the tail of the array/slice.

These functions can be pipelined to simplify complex template expressions. Below is an example of emitting the word "foo"
when module "app-a" has the value "a" in it's tags property.
`,
	"build-summary": `Run build command`,
	"build": `{{cli "Run build command \n"}}
{{c "mbt build branch [name] [--content] [--name <name>] [--fuzzy]"}}{{br}}
Build modules in a branch. Assume master if branch name is not specified.
Build just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt build commit <commit> [--content] [--name <name>] [--fuzzy]"}}{{br}}
Build modules in a commit. Full commit sha is required.
Build just the modules modified in the commit when {{c "--content"}} flag is used.
Build just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt build diff --from <commit> --to <commit>"}}{{br}}
Build modules changed between {{c "from"}} and {{c "to"}} commits.
In this mode, mbt works out the merge base between {{c "from"}} and {{c "to"}} and
evaluates the modules changed between the merge base and {{c "to"}}.

{{c "mbt build head [--content] [--name <name>] [--fuzzy]"}}{{br}}
Build modules in current head.
Build just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt build pr --src <name> --dst <name>"}}{{br}}
Build modules changed between {{c "--src"}} and {{c "--dst"}} branches.
In this mode, mbt works out the merge base between {{c "--src"}} and {{c "--dst"}} and
evaluates the modules changed between the merge base and {{c "--src"}}.

{{c "mbt build local [--all] [--content] [--name <name>] [--fuzzy]"}}{{br}}
Build modules modified in current workspace. All modules in the workspace are
built if {{c "--all"}} option is specified.
Build just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{h2 "Build Environment"}}

When executing build, following environment variables are initialised and can be
used by the command being executed.

- {{c "MBT_MODULE_NAME"}} Name of the module
- {{c "MBT_MODULE_PATH"}} Relative path to the module directory
- {{c "MBT_MODULE_VERSION"}} Module version
- {{c "MBT_BUILD_COMMIT"}} Git commit SHA of the commit being built
- {{c "MBT_REPO_PATH"}} Absolute path to the repository directory

In addition to the variables listed above, module properties are also populated 
in the form of {{c "MBT_MODULE_PROPERTY_XXX"}} where {{c "XXX"}} denotes the key.
`,
	"describe-summary": `Describe repository manifest`,
	"describe": `{{cli "Describe repository manifest \n"}}
{{c "mbt describe branch [name] [--content] [--name <name>] [--fuzzy] [--graph] [--json]"}}{{br}}
Describe modules in a branch. Assume master if branch name is not specified.
Describe just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt describe commit <commit> [--content] [--name <name>] [--fuzzy] [--graph] [--json]"}}{{br}}
Describe modules in a commit. Full commit sha is required.
Describe just the modules modified in the commit when {{c "--content"}} flag is used.
Describe just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt describe diff --from <commit> --to <commit> [--graph] [--json]"}}{{br}}
Describe modules changed between {{c "from"}} and {{c "to"}} commits.
In this mode, mbt works out the merge base between {{c "from"}} and {{c "to"}} and
evaluates the modules changed between the merge base and {{c "to"}}.

{{c "mbt describe head [--content] [--name <name>] [--fuzzy] [--graph] [--json]"}}{{br}}
Describe modules in current head.
Describe just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt describe pr --src <name> --dst <name> [--graph] [--json]"}}{{br}}
Describe modules changed between {{c "--src"}} and {{c "--dst"}} branches.
In this mode, mbt works out the merge base between {{c "--src"}} and {{c "--dst"}} and
evaluates the modules changed between the merge base and {{c "--src"}}.

{{c "mbt describe local [--all] [--content] [--name <name>] [--fuzzy] [--graph] [--json]"}}{{br}}
Describe modules modified in current workspace. All modules in the workspace are
described if {{c "--all"}} option is specified.
Describe just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{h2 "Output Formats"}}
Use {{c "--graph"}} option to output the manifest in graphviz dot format. This can
be useful to visualise build dependencies.

Use {{c "--json"}} option to output the manifest in json format.

`,
	"run-in-summary": `Run user defined command`,
	"run-in": `{{cli "Run user defined command \n"}}
{{c "mbt run-in branch [name] [--content] [--name <name>] [--fuzzy]"}}{{br}}
Run user defined command in modules in a branch. Assume master if branch name is not specified.
Consider just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt run-in commit <commit> [--content] [--name <name>] [--fuzzy]"}}{{br}}
Run user defined command in modules in a commit. Full commit sha is required.
Consider just the modules modified in the commit when {{c "--content"}} flag is used.
Consider just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt run-in diff --from <commit> --to <commit>"}}{{br}}
Run user defined command in modules changed between {{c "from"}} and {{c "to"}} commits.
In this mode, mbt works out the merge base between {{c "from"}} and {{c "to"}} and
evaluates the modules changed between the merge base and {{c "to"}}.

{{c "mbt run-in head [--content] [--name <name>] [--fuzzy]"}}{{br}}
Run user defined command in modules in current head.
Consider just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{c "mbt run-in pr --src <name> --dst <name>"}}{{br}}
Run user defined command in modules changed between {{c "--src"}} and {{c "--dst"}} branches.
In this mode, mbt works out the merge base between {{c "--src"}} and {{c "--dst"}} and
evaluates the modules changed between the merge base and {{c "--src"}}.

{{c "mbt run-in local [--all] [--content] [--name <name>] [--fuzzy]"}}{{br}}
Run user defined command in modules modified in current workspace. All modules in the workspace are
considered if {{c "--all"}} option is specified.
Consider just the modules matching the {{c "--name"}} filter if specified.
Default {{c "--name"}} filter is a prefix match. You can change this to a subsequence
match by using {{c "--fuzzy"}} option.

{{h2 "Execution Environment"}}

When executing a command, following environment variables are initialised and can be
used by the command being executed.

{{c "MBT_MODULE_NAME"}} Name of the module
{{c "MBT_MODULE_VERSION"}} Module version
{{c "MBT_BUILD_COMMIT"}} Git commit SHA of the commit being built

In addition to the variables listed above, module properties are also populated 
in the form of {{c "MBT_MODULE_PROPERTY_XXX"}} where {{c "XXX"}} denotes the key.
`,
}

func docText(name string) string {
	templateString, ok := templates[name]
	if !ok {
		panic("template not found")
	}

	t := template.New(name)
	t, err := augmentWithTemplateFuncs(t).Parse(templateString)
	if err != nil {
		panic(err)
	}
	return generateHelpText(t)
}

func augmentWithTemplateFuncs(t *template.Template) *template.Template {
	heading := func(s string, level int) string {
		if os.Getenv("MBT_DOC_GEN_MARKDOWN") == "1" {
			prefix := ""
			switch level {
			case 1:
				prefix = "##"
			case 2:
				prefix = "###"
			default:
				prefix = "####"
			}
			return fmt.Sprintf("%s %s\n", prefix, s)
		}

		line := ""
		for i := 0; i < len(s); i++ {
			line = line + "="
		}

		return fmt.Sprintf("%s\n%s", s, line)
	}

	return t.Funcs(template.FuncMap{
		"c": func(s string) string {
			if os.Getenv("MBT_DOC_GEN_MARKDOWN") != "1" {
				return s
			}
			if s == "" {
				return "```"
			}
			return fmt.Sprintf("`%s`", s)
		},
		"link": func(text, link string) string {
			if os.Getenv("MBT_DOC_GEN_MARKDOWN") != "1" {
				return fmt.Sprintf("%s %s", text, link)
			}
			return fmt.Sprintf("[%s](%s)", text, link)
		},
		"h1": func(s string) string {
			return heading(s, 1)
		},
		"h2": func(s string) string {
			return heading(s, 2)
		},
		"h3": func(s string) string {
			return heading(s, 3)
		},
		"br": func() string {
			if os.Getenv("MBT_DOC_GEN_MARKDOWN") == "1" {
				return "<br>"
			}
			return ""
		},
		"cli": func(s string) string {
			if os.Getenv("MBT_DOC_GEN_MARKDOWN") == "1" {
				return ""
			}
			return s
		},
	})
}

func generateHelpText(t *template.Template) string {
	buff := new(bytes.Buffer)
	err := t.Execute(buff, nil)
	if err != nil {
		panic(err)
	}
	return buff.String()
}
