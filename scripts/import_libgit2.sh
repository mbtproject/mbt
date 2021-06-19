#!/bin/sh

#
# Use this utility to import libgit2 sources into the directory.
#

set -e

DIR="$(pwd)"
LIBGIT2_PATH="$DIR/libgit2"
LIBGIT2_VERSION="v0.28.5"

if [ -n "$LIBGIT2_VERSION" ]; then
  LIBGIT2_BRANCH="-b ${LIBGIT2_VERSION}"
fi

if [ -d $LIBGIT2_PATH ]; then
  rm -rf $LIBGIT2_PATH
fi

git clone $LIBGIT2_BRANCH https://github.com/libgit2/libgit2.git $LIBGIT2_PATH

rm -rf $LIBGIT2_PATH/.git
