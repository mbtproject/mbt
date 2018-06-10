![mbt](assets/logo-m.png)

# mbt

>> The most flexible build orchestration tool for monorepo

[Documentation](https://mbtproject.github.io/mbt/) | [Twitter](https://twitter.com/mbtproject)

[![Build Status](https://travis-ci.org/mbtproject/mbt.svg?branch=release)](https://travis-ci.org/mbtproject/mbt)
[![Build status](https://ci.appveyor.com/api/projects/status/wy8rhr188t3phqvk?svg=true)](https://ci.appveyor.com/project/mbtproject/mbt)
[![Go Report Card](https://goreportcard.com/badge/github.com/mbtproject/mbt)](https://goreportcard.com/report/github.com/mbtproject/mbt)
[![Coverage Status](https://coveralls.io/repos/github/mbtproject/mbt/badge.svg)](https://coveralls.io/github/mbtproject/mbt)

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
