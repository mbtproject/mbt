#!/bin/sh

set -e

DIR=$(pwd)
LIBGIT2_PATH=$DIR/libgit2

mkdir -p $LIBGIT2_PATH

cd $LIBGIT2_PATH &&
mkdir -p install/lib &&
mkdir -p build &&
cd build &&
cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      -DUSE_SSH=OFF \
      -DCURL=OFF \
      .. &&

cmake --build . &&
make -j2 install &&

cd $DIR
