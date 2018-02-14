#!/bin/bash

set -e

# Utility functions
lint() {
  paths=$@
  for path in "${paths[@]}"
  do
    r=$(gofmt -s -d $path)
    if [ ! -z "$r" ]; then
      echo "ERROR: Linting failed. Review errors and try running gofmt -s -w $path"
      echo $r
      exit 1
    fi
  done
} 

DIR=$(pwd)
LIBGIT2_PATH=$DIR/vendor/libgit2
OS=$(uname -s | awk '{print tolower($0)}')
ARCH=$(uname -m)

# Restore build dependencies
go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls

# Build libgit2
./scripts/build_libgit2.sh 

# Set environment so to static link libgit2 when building git2go
export PKG_CONFIG_PATH="$LIBGIT2_PATH/build"
export CGO_LDFLAGS="$(pkg-config --libs --static $LIBGIT2_PATH/build/libgit2.pc)"

# All preparation is done at this point.
# Move on to building mbt
cd $DIR

VERSION=$TRAVIS_TAG
if [ -z $VERSION ]; then
  VERSION="0.0.0"
fi

OUT="mbt_${OS}_${ARCH}"

make restore 

# Run linter
lint *.go ./e ./dtrace ./trie ./lib

# Run tests
go test ./e -v -covermode=count 
go test ./dtrace -v -covermode=count 
go test ./trie -v -covermode=count 
go test ./lib -v -covermode=count -coverprofile=coverage.out 
if [ ! -z $COVERALLS_TOKEN ] && [ -f ./coverage.out ]; then 
  $HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $COVERALLS_TOKEN
fi

go build -o "build/${OUT}" 
shasum -a 1 -p "build/${OUT}" | cut -d ' ' -f 1 > "build/${OUT}.sha1" 

# Run go vet (this should happen after the build)
go tool vet ./*.go
go tool vet ./e ./dtrace ./trie ./lib

echo "testing the bin"
"./build/${OUT}" version

cat >build/bintray.json << EOL
{
    "package": {
        "name": "${OUT}",
        "repo": "bin",
        "subject": "buddyspike",
        "desc": "Monorepo Build Tool",
        "website_url": "https://github.com/mbtproject/mbt", "issue_tracker_url": "https://github.com/mbtproject/mbt/issues", "vcs_url": "https://github.com/buddyspike/mbt.git", "public_download_numbers": true, "public_stats": true }, "version": {
        "name": "${VERSION}",
        "gpgSign": false
    },
    "files": [ {"includePattern": "build/${OUT}", "uploadPattern": "/${OUT}"} ],
    "publish": true
}
EOL