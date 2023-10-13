#!/bin/sh

#
# Use this utility to import libgit2 sources into the directory.
#

set -e

DIR="$(pwd)"
LIBGIT2_PATH="$DIR/libgit2"
LIBGIT2_VERSION="v1.5.2"

if [ -d $LIBGIT2_PATH ]; then
  rm -rf $LIBGIT2_PATH
fi

git clone https://github.com/libgit2/libgit2.git $LIBGIT2_PATH

cd $LIBGIT2_PATH
git checkout $LIBGIT2_VERSION

cd $DIR
rm -rf $LIBGIT2_PATH/.git
