#!/bin/sh

#
# Use this utility to import libgit2 and git2go sources into
# vendor directory.
#

set -e

DIR="$(pwd)"
GIT2GO_PATH=vendor/github.com/libgit2/git2go
LIBGIT2_PATH=vendor/libgit2
GIT2GO_VERSION="v26"
LIBGIT2_VERSION="v0.26.0"


if [ -d $LIBGIT2_PATH ]; then
  rm -rf $LIBGIT2_PATH
fi

if [ -d $GIT2GO_PATH ]; then
  rm -rf $GIT2GO_PATH
fi

git clone https://github.com/libgit2/libgit2.git $LIBGIT2_PATH
git clone https://github.com/libgit2/git2go.git $GIT2GO_PATH

cd $LIBGIT2_PATH
git checkout $LIBGIT2_VERSION

cd $DIR
cd $GIT2GO_PATH
git checkout $GIT2GO_VERSION

cd $DIR
rm -rf $LIBGIT2_PATH/.git
rm -rf $GIT2GO_PATH/.git
