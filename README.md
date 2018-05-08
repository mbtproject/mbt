![mbt](assets/logo-m.png)
# mbt
>> The most flexible build tool for monorepo

[Documentation](doc) | [Twitter](https://twitter.com/mbtproject)

[![Build Status](https://travis-ci.org/mbtproject/mbt.svg?branch=release)](https://travis-ci.org/mbtproject/mbt)
[![Build status](https://ci.appveyor.com/api/projects/status/wy8rhr188t3phqvk?svg=true)](https://ci.appveyor.com/project/mbtproject/mbt)
[![Go Report Card](https://goreportcard.com/badge/github.com/mbtproject/mbt)](https://goreportcard.com/report/github.com/mbtproject/mbt)
[![Coverage Status](https://coveralls.io/repos/github/mbtproject/mbt/badge.svg)](https://coveralls.io/github/mbtproject/mbt)

mbt is a build tool for monorepo (single repository with many applications).

## Features

- Differential Builds
- Content Based Versioning
- Build Dependency Management
- Dependency Visualisation
- Template Driven Deployments

## Status
mbt is production ready. We try our best to maintain semver.
Visit [Github issues](https://github.com/mbtproject/mbt/issues) for support.

## Install
```sh
curl -L -o /usr/local/bin/mbt [get the url for your target from the links below]
chmod +x /usr/local/bin/mbt
```

## Releases

### Stable
|OS               |Download|
|-----------------|--------|
|darwin x86_64    |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_darwin_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_darwin_x86_64/_latestVersion)|
|linux x86_64     |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_linux_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_linux_x86_64/_latestVersion)|
|windows          |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_windows_x86/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_windows_x86/_latestVersion)|

### Dev Channel
|OS               |Download|
|-----------------|--------|
|darwin x86_64    |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_dev_darwin_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_dev_darwin_x86_64/_latestVersion)|
|linux x86_64     |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_dev_linux_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_dev_linux_x86_64/_latestVersion)|
|windows          |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_dev_windows_x86/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_dev_windows_x86/_latestVersion)|

## Building Locally

### Linux/OSX

- You need `cmake` and `pkg-config` (latest of course is preferred)
- Get the code `go get github.com/mbtproject/mbt`
- Change to source directory `cd $GOPATH/src/github.com/mbtproject/mbt`

  If you haven't set `$GOPATH`, change it to `~/go` which is the default place used by `go get`.
  See [this](https://golang.org/cmd/go/#hdr-GOPATH_environment_variable) for more information about `$GOPATH`

- Run `make build` to build and run all unit tests
- Run `make install` to install the binary in `$GOPATH/bin`

  Make sure `$GOPATH/bin` is in your path in order to execute the binary

### Windows
Local builds on Windows is not currently supported. 
However, the specifics can be found in our CI scripts (`appveyor.yml` and `build_win.bat`)

## Introduction

Repositories containing source for different software modules traditionally
confronted with unique build infrastructure problems.

- Build Isolation

  Mainstream build utilities typically trigger builds
  based on the changes to the repository. 
  They normally don't provide built-in functionality to build sub-trees that have
  been modified in one or many changes. 
  Building all modules sometimes do the trick although the 
  mileage may vary depending on the size and complexity of the build process. 
  As the source gets bigger and builds get more complex, this approach 
  becomes unsustainable not only time-wise but also in a commercial sense.

- Versioning

  It is a common practice to use git commit SHA's to 
  annotate build artifacts so that they can be associated
  to a particular point in time in the the source control.
  Any tooling with the ability to build modified sub-trees
  would have to provide a similar construct for versioning
  because commit SHA speak for the entire repository as
  whole.

mbt is built exactly around those two problems but also loaded
with a few nifty utilities to make it useful for contemporary applications.

## Basics
In the context of `mbt`, term 'Module' is used to refer to a part of source
tree that can be developed, built and released independently.
Modules can be built with different programming languages and their
native build tools.

For example, you could have Java modules built with Maven/Gradle,
.NET modules built with MSBUILD and NodeJS modules built with npm scripts - all
in one repository.
The idea is, module developers should be able to use the build tools native to their stack.

Each module in a repository is stored in its own directory with a spec file
called `.mbt.yml`. Presence of the spec file indicates `mbt` that the directory
contains module and how to build it.

## Spec File
`.mbt.yml` file must be in yaml and has the following structure.

```yaml
name: module-a      # name of the module
build:              # list of build commands to execute in each platform
  darwin:
    cmd: npm        # build command
    args: [run, build]    # optional list of arguments to be passed when invoking the build command
  linux:
    cmd: npm
    args: [run, build]
  windows:
    cmd: npm
    args: [run, build]
```

In the preceding spec file, we are basically telling `mbt` that, `module-a`
can be built by running `npm run build` on OSX, Linux and Windows. With
this committed to the repository, we can start using `mbt` cli to build in
several ways as shown below.

```
# Build the current master branch
mbt build branch

# Build the current branch
mbt build head

# Build specific commit
mbt build commit <commit sha>

# Build only the changes in your local workdir without commiting
mbt build local

# Build everything in the current workdir
mbt build local --all

# Build a specific branch
mbt build branch feature

# Build only the modules changed in a specific branch relative to another
mbt build pr --src feature --dst master

# Build only the modules changed between two commits
mbt build diff --from <commit sha> --to <commit sha>

```

## Dependencies
### Module Dependencies
One advantage of a single repo is that you can share code between multiple modules
more effectively. Building this type of dependencies requires some thought. mbt provides
an easy way to define dependencies between modules and automatically builds the impacted modules
in topological order.

Dependencies are defined in `.mbt.yml` file under 'dependencies' property.
It accepts an array of module names.
For example, `module-a` could define a dependency on `module-b` as shown below,
so that any time `module-b` is changed, build command for `module-a` is also executed.

```yaml
name: module-a
dependencies: [module-b]
```

One example of where this could be useful is shared libraries. A shared library
could be developed independently of its consumers. However, all consumers
gets automatically built whenever the shared library is modified.

### File Dependencies
File dependencies are useful in situations where a module should be built
when a file(s) stored outside the module directory is modified. As an example,
a build script that is used to build multiple modules could define a file
dependency in order to trigger the build whenever there's a change in build.

File dependencies should specify the path of the file relative to the root
of git repository.

```yaml
name: module-a
fileDependencies:
  - scripts/build.sh
```

## Build Environment
When `mbt` executes the build command specified for a module, it sets up
several important attributes as environment variables. These variables can
be used by the actual build tools in the process.

|Variable Name |Description |
|---|---|
|MBT_MODULE_NAME |Name of the module |
|MBT_MODULE_VERSION |SHA1 hash calculated based on the content of the module directory and the content of any of its dependencies (recursively) |
|MBT_BUILD_COMMIT |Git commit SHA of the commit being built |

In addition to the variables listed above, module properties (arbitrary list of
key/value pairs described below) is also populated in the form of `MBT_MODULE_PROPERTY_XXX`
where `XXX` denotes the key.

One useful scenario of these variables would be, a build command that produces a
docker image. In that case, we could tag it with `MBT_MODULE_VERSION` so that the
image produced as of a particular Git commit SHA can be identified accurately.
(We will also discuss how this information can be used to automatically produce
deployment artifacts later in this document)

## User Defined Commands
User defined commands provide the ability to run arbitrary commands on a set of
modules in a similar way to build.

They are defined in `commands` section in the spec file.

```
module: app-a
commands:
  hello:
    cmd: echo
    args: [hello]
```

With preceding spec file, we could run `hello` command in relevant modules in
current branch by running:

```
mbt run-in head --command hello
```

`run-in` command maintains the symmetry with `build` and `describe` commands.

## Describing the Change
When working in a dense, modular codebase it is sometimes important to assess
the impact of your changes to the overall system. Each `mbt build` command
variation has a `mbt describe` counterpart to see what modules are going to
to be built. For example, to list modules affected between two git branches, we
could run:

```
mbt describe pr --src <source-branch-name> --dst <destination-branch-name>
```

Furthermore, you can specify `--json` flag to get the output of `describe`
formatted in `json` so that it can be easily parsed and used by tools in the
rest of your CD pipeline.

## Going Beyond the Build
`mbt` produces a data structure based on the information presented in `.mbt.yml` files.
That is what basically drives the build behind scenes. Since it contains
information about all modules in a repository, it could also be useful for producing
other assets such as deployment scripts. This can be achieved with `mbt apply` command.
It gives us the ability to apply module information discovered by `mbt` to a go template
stored within the repository.

As a simple but rather useless example of this can be illustrated with a template
as shown below.

```go
{{ $module := range .ModulesList }}
{{ $module.Name }}
{{ end }}
```

With this template committed into repo, we could run `mbt apply` in several
useful ways to produce the list of module names in the repository.

```
# Apply the state of master branch
mbt apply branch --to <path to the template>

# Apply the state of another branch
mbt apply branch <branch-name> --to <path to the template>

# Apply the state as of a particular commit
mbt apply commit <git-commit-sha> --to <path to the template>

# Apply the state of local working directory (i.e. uncommitted work)
mbt apply local --to <path to the template>
```

Output of above commands written to `stdout` by default but can be directed to a
file using `--out` argument.

It's hard to imagine useful template transformation scenarios with just the basic
information about modules. To make it little bit more flexible, we add a couple
user-defined properties the data structure used in template transformation.
First one of them is called, `.Env`. This is a map of key/value pairs containing
arbitrary environment variables provisioned for `mbt` command.

For example, running mbt command with an environment variable as shown below would
make the key `TARGET` available with value `PRODUCTION` in  `.Env` property.

```
TARGET=PRODUCTION mbt apply branch
```

Second property is available in each module and can be specified in `.mbt.yml`
file.

```
name: module-a
properties:
  a: foo
  b: bar
```

`module-a` shown in above `.mbt.yml` snippet would make properties `a` and `b`
available to templates via `.Properties` property as illustrated below.

```go
{{ $module := range .ModulesList }}
{{ property $module "a" }} {{ property $module "b" }}
{{ end }}
```

More realistic example of this capability is demonstrated in [this example repo](
  https://github.com/mbtproject/demo). It generates docker images for two web
applications hosted in nginx, pushes them to specified docker registry and
generates a Kubernetes deployment manifest using `mbt apply` command.

### Template Helpers
Following helper functions are available when writing templates.

|Helper |Description
|---|---
|`module <name>` |Find the specified module from modules set discovered in the repo.
|`property <module> <name>` |Find the specified property in the given module. Standard dot notation can be used to access nested properties (e.g. `a.b.c`).
|`propertyOr <module> <name> <default>` |Find specified property in the given module or return the designated default value.
|`contains <array> <item>` |Return true if the given item is present in the array.
|`join <array> <format> <separator>` |Format each item in the array according the format specified and then join them with the specified separator.
|`kvplist <map>` |Accepts a map of `map[string]interface{}` type and returns the items in the map as a list of key/value pairs sorted by the key. They can be accessed via `.Key` and `.Value` properties.
|`add <int> <int>` |Adds two integers.
|`sub <int> <int>` |Subtracts two integers.
|`mul <int> <int>` |Multiplies two integers.
|`div <int> <int>` |Divides two integers.
|`head <array|slice>` |Selects the first item in an array/slice. Returns `""` if the input is `nil` or an empty array/slice.
|`tail <array|slice>` |Selects the last item in an array/slice. Returns `""` if the input is `nil` or an empty array/slice.
|`ishead` <array> <value>` |Returns true if the specified value is the head of the array/slice.
|`istail` <array> <value>` |Returns true if the specified value is the tail of the array/slice.

These functions can be pipelined to simplify complex template expressions. Below is an example of emitting the word "foo"
when module "app-a" has the value "a" in it's tags property.
```
{{if contains (property (module "app-a") "tags") "a"}}foo{{end}}
```

## Demo
[![asciicast](https://asciinema.org/a/KJxXNgrTs9KZbVV4GYNN5DScC.png)](https://asciinema.org/a/KJxXNgrTs9KZbVV4GYNN5DScC)

## Credits
`mbt` is powered by these awesome libraries
- [git2go](https://github.com/libgit2/git2go)
- [libgit2](https://github.com/libgit2/libgit2)
- [yaml](https://github.com/go-yaml/yaml)
- [cobra](https://github.com/spf13/cobra)
- [logrus](https://github.com/sirupsen/logrus)

Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>
