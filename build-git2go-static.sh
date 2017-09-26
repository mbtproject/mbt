#!/bin/sh

set -ex

DIR=$(pwd)
GIT2GO_PATH=$GOPATH/src/github.com/libgit2/git2go
GIT2GO_VENDOR_PATH=$GIT2GO_PATH/vendor/libgit2

go get github.com/libgit2/git2go || true &&

cd $GIT2GO_PATH &&
git checkout v26 &&
git submodule update --init || true &&

cd $GIT2GO_VENDOR_PATH &&
mkdir -p install/lib &&
mkdir -p build &&
cd build &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      .. &&

cmake --build . &&
make -j2 install &&

export PKG_CONFIG_PATH=$GIT2GO_VENDOR_PATH/build
export CGO_LDFLAGS="$(pkg-config --libs --static $GOPATH/src/github.com/libgit2/git2go/vendor/libgit2/build/libgit2.pc)"
go install -x github.com/libgit2/git2go &&

cd $DIR