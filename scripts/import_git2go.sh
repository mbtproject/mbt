#!/bin/sh

#
# Use this utility to import git2go sources into the directory.
#

set -e

DIR="$(pwd)"
GIT2GO_PATH="$DIR/git2go"
GIT2GO_VERSION="v34.0.0"


if [ -d $GIT2GO_PATH ]; then
  rm -rf $GIT2GO_PATH
fi

git clone https://github.com/libgit2/git2go.git $GIT2GO_PATH

cd $GIT2GO_PATH
git checkout $GIT2GO_VERSION

cd $DIR
rm -rf $GIT2GO_PATH/.git
