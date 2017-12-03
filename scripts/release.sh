#!/bin/sh

set -e

VERSION=$(go run ./scripts/update_version.go $1)
git add -A
git commit -m "Bump version"
git tag $VERSION 
git push origin master --tags
