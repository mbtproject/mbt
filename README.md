# mbt
>> Build utility for monorepos

Experimental and work in progress.

![build status](https://travis-ci.org/buddyspike/mbt.svg?branch=master)

## Format of appspec.yaml

```yaml
name: my-cool-app   # name of the app
buildPlaforms:      # list of supported build platforms
  - linux
  - darwin
build: ./build.sh   # command to execute when running mbt build 
properties:         # dict of arbitrary values that can be used in templates when running mbt apply
  foo: bar
```

## Usage Examples
```sh
# Display manifest in default branch 
mbt describe branch --in [path to repo]

# Display manifest for a specific branch
mbt describe branch [branch name] --in [path to repo]

# Display manifest for a commit
mbt describe commit [full commit sha] --in [path to repo]

# Display manifest for a pull request
mbt describe pr --src [source branch name] --dst [destination branch name] --in [path to repo]

# Build default branch
mbt build branch --in [path to repo]

# Build specific branch 
mbt build branch [branch name] --in .

# Build a pull request
mbt build pr --src [source branch name] --dst [destination branch name] --in [path to repo]

# Apply the manifest from default branch over a go template
# Template is read out from git storage. Therefore must be committed.
mbt apply branch --to [relative path to template in tree] --in . 

# Apply the manifest from a branch over a go template
mbt apply branch [branch name] --to [relative path to template in tree] --in .

# Apply the manifest and write the output to a file
mbt apply branch --to [relative path to template in tree] --out [path to output file] --in .
```

## Builds

|OS               |Download|
|-----------------|--------|
|darwin x86_64    |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_darwin_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_darwin_x86_64/_latestVersion)|
|linux x86_64     |[![Download](https://api.bintray.com/packages/buddyspike/bin/mbt_linux_x86_64/images/download.svg)](https://bintray.com/buddyspike/bin/mbt_linux_x86_64/_latestVersion)|

