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

OUT="mbt_${OS}_${ARCH}"
VERSION=$VERSION
if [ -z $VERSION ]; then
  OUT="mbt_dev_${OS}_${ARCH}"
  if [ ! -z $TRAVIS_COMMIT ]; then
    VERSION="dev-$(echo $TRAVIS_COMMIT | head -c8)"
    go run scripts/update_version.go -custom "$VERSION"
  else
    VERSION="dev-$(git rev-parse HEAD | head -c8)"
  fi
fi


make restore

# Run linter
lint *.go ./e ./dtrace ./trie ./intercept ./lib

# Run tests
go test ./e -v -covermode=count
go test ./dtrace -v -covermode=count
go test ./trie -v -covermode=count
go test ./intercept -v -covermode=count
go test ./graph -v -covermode=count
go test ./utils -v -covermode=count
go test ./lib -v -covermode=count -coverprofile=coverage.out
if [ ! -z $COVERALLS_TOKEN ] && [ -f ./coverage.out ]; then
  $HOME/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $COVERALLS_TOKEN
fi

go build -o "build/${OUT}"
shasum -a 1 -p "build/${OUT}" | cut -d ' ' -f 1 > "build/${OUT}.sha1"

# Run go vet (this should happen after the build)
go vet ./*.go
go vet ./e ./dtrace ./trie ./intercept ./lib ./graph

echo "testing the bin"
"./build/${OUT}" version

if [ ! -z $PUBLISH_TOKEN ]; then
  echo "publishing $OUT"
  curl \
    -ubuddyspike:$PUBLISH_TOKEN \
    -X POST \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$VERSION\", \"description\": \"$VERSION\", \"published\": true}" \
    "https://api.bintray.com/packages/buddyspike/bin/$OUT/versions"

  curl \
    -ubuddyspike:$PUBLISH_TOKEN \
    --progress-bar \
    -T "build/$OUT" \
    -H "X-Bintray-Package:$OUT" \
    -H "X-Bintray-Version:$VERSION" \
    "https://api.bintray.com/content/buddyspike/bin/$OUT/$VERSION/$VERSION/$OUT?publish=1"
fi
