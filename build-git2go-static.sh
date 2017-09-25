#!/bin/sh

set -ex

GIT2GO_PATH=$GOPATH/src/github.com/libgit2/git2go
GIT2GO_VENDOR_PATH=$GIT2GO_PATH/vendor/libgit2

go get github.com/libgit2/git2go &&

pushd $GIT2GO_PATH &&
git checkout v26 &&
git submodule update --init &&

pushd $GIT2GO_VENDOR_PATH &&
mkdir -p install/lib &&
mkdir -p build &&
pushd build &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      .. &&

cmake --build . &&
make -j2 install &&

go install github.com/libgit2/git2go &&

popd && popd && popd 
