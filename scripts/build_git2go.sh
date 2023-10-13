#!/bin/sh

set -e

DIR=$(pwd)
GIT2GO_PATH=$DIR/git2go

# First ensure that git2go source tree is available
if [ ! -d git2go ]
then
    ./scripts/import_git2go.sh
fi

if [ ! -L $GIT2GO_PATH/vendor/libgit2 ]; then
    rm -rf $GIT2GO_PATH/vendor/libgit2
    ln -s $DIR/libgit2 $GIT2GO_PATH/vendor/libgit2
fi

cd $GIT2GO_PATH
make install-static

cd ${DIR}
