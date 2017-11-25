![mbt](assets/logo-m.png)
# mbt
>> Build utility for monorepos

[![Build Status](https://travis-ci.org/mbtproject/mbt.svg?branch=release)](https://travis-ci.org/mbtproject/mbt)
[![Build status](https://ci.appveyor.com/api/projects/status/wy8rhr188t3phqvk?svg=true)](https://ci.appveyor.com/project/mbtproject/mbt)
[![Go Report Card](https://goreportcard.com/badge/github.com/mbtproject/mbt)](https://goreportcard.com/report/github.com/mbtproject/mbt)
[![Coverage Status](https://coveralls.io/repos/github/mbtproject/mbt/badge.svg)](https://coveralls.io/github/mbtproject/mbt)

Monorepos are a great way to organize large, modular applications.
`mbt` is a build utility to build only the modules affected 
by a particular change in a git repository. 

Modules can be built with different programming languages and their native build tools. 
For example, a repository could have .NET modules built with msbuild 
and NodeJS modules built with npm scripts. 
Developers of these modules should be able to develop with tools they are 
already familiar with.

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

### Various Ways to Invoke `mbt`
```
# Build the current branch 
mbt build branch --in .

# Build a specific branch
mbt build branch feature --in .

# Build only the modules changed in a specific branch relative to another
mbt build pr --src feature --dst master --in .

# Build only the modules changed between two commits
mbt build diff --from <commit-sha> --to <commit-sha> --in .

```

## Dependencies
Sometimes a change to a module could require the build of the modules that 
depend on it. We can define these dependencies in `.mbt.yml` files and `mbt` 
takes care of triggering the build in the topological order.

For example, `module-a` could define a dependency on `module-b`, so that any
time `module-b` is changed, build command for `module-a` is also executed.

```yaml
name: module-a 
dependencies: [module-b]
```

One example of where this could be useful is shared libraries. A shared library
could be developed independently of its consumers. However, all consumers 
gets automatically built whenever the shared library is modified. 

Another situation could be where one module is a single page web app and the 
other module is the back-end API and the host application of that. It's a 
prevalent model to develop such modules in different programming languages and 
usually host app build contains the steps to pack the web app into the deployment 
package.
Using the dependency feature, host module could depend on the client app 
module so that anytime the web app changes the host module is also built. 

## Build Environment
When `mbt` executes the build command specified for a module, it sets up 
several important attributes as environment variables. These variables can 
be used by the actual build tools in the process.

|Variable Name |Description |
|---|---|
|MBT_APP_NAME |Name of the module |
|MBT_APP_VERSION |SHA1 hash calculated based on the content of the module directory and the content of any of its dependencies (recursively) |
|MBT_BUILD_COMMIT |Git commit SHA for the commit which is resolved for the build |

In addition to the variables listed above, module properties (arbitrary list of 
key/value pairs described below) is also populated in the form of `MBT_PROPERTY_XXX`
where `XXX` denotes the key.

One useful scenario of these variables would be a build command that produces a 
docker image. In that case, we could tag it with `MBT_APP_VERSION` so that the 
image produced as of a particular Git commit SHA can be identified accurately. 
(We will also discuss how this information can be used to automatically produce 
deployment artifacts later in this document)

## Describing the Change
When working in a dense, modular codebase it is sometimes important to assess 
the impact of your changes to the overall system. Each `mbt build` command 
variation has a `mbt describe` counterpart to see what modules are going to 
to be built. For example, to list modules affected between two git branches we 
could run:

```
mbt describe pr --src <source-branch-name> --dst <destination-branch-name> --in .
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
{{ $app := range .Applications }}
{{ app.Name }}
{{ end }}
```

With this template committed into repo, we could run `mbt apply` in several 
useful ways to produce the list of module names in the repository.

```
# Apply the state of master branch
mbt apply branch --to <path to the template> --in <path to repo>

# Apply the state of another branch
mbt apply branch <branch-name> --to <path to the template> --in <path to repo>

# Apply the state as of a particular commit
mbt apply commit <git-commit-sha> --to <path to the template> --in <path to repo>
```

Output of above commands written to `stdout` by default but can be directed to a 
file using `--out` argument.

## Install
```sh
curl -L -o /usr/local/bin/mbt [get the url for your target from the links below]
chmod +x /usr/local/bin/mbt
```
## Builds

|OS               |Download|
|-----------------|--------|
|darwin x86_64    |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_darwin_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_darwin_x86_64/_latestVersion)|
|linux x86_64     |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_linux_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_linux_x86_64/_latestVersion)|
|windows          |[ ![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_windows_x86/images/download.svg) ](https://bintray.com/buddyspike/bin/mbt_windows_x86/_latestVersion)|

[This blog post](https://buddyspike.github.io/blog/post/building-modular-systems-with-mbt/) covers some initial thinking behind the tool.

## Credits
`mbt` is powered by these awesome libraries
- [git2go](https://github.com/libgit2/git2go)
- [libgit2](https://github.com/libgit2/libgit2) 
- [yaml](https://github.com/go-yaml/yaml)
- [cobra](https://github.com/spf13/cobra)
- [logrus](https://github.com/sirupsen/logrus)

Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>