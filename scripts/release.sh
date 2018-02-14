#!/bin/sh

set -e

VERSION=$(go run ./scripts/update_version.go $1)
git add -A
git commit -S -m "Bump version - v$VERSION"
git tag -a $VERSION -m $VERSION --sign
git push origin master --tags
