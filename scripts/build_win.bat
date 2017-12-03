SET MBT=%GOPATH%\src\github.com\mbtproject\mbt
SET GIT2GO_PATH=%GOPATH%\src\github.com\libgit2\git2go
SET GIT2GO_VENDOR_PATH=%GIT2GO_PATH%\vendor\libgit2
SET OS=windows
SET ARCH=x86
set OUT="mbt.exe"

go version
go get -d github.com/libgit2/git2go

cd %GIT2GO_PATH%
git checkout v26
git submodule update --init

cd %GIT2GO_VENDOR_PATH%
mkdir install
mkdir install/lib
mkdir build
cd build
cmake -DTHREADSAFE=ON -DBUILD_CLAR=OFF -DBUILD_SHARED_LIBS=ON -DCMAKE_C_FLAGS=-fPIC -DCMAKE_INSTALL_PREFIX=../install -DUSE_SSH=OFF -DCURL=OFF ..

cmake --build .
cmake --build . --target install

SET PKG_CONFIG_PATH=%GIT2GO_VENDOR_PATH%/build
for /f %%i in ('pkg-config --libs %GOPATH%\src\github.com\libgit2\git2go\vendor\libgit2\build\libgit2.pc') do set CGO_LDFLAGS=%%i

cd %MBT%

rd /s /q build
mkdir build

go get -t
go get github.com/stretchr/testify

copy %GIT2GO_VENDOR_PATH%\install\bin\git2.dll lib\git2.dll

go test -v ./...

go build -o "build/%OUT%"
copy %GIT2GO_VENDOR_PATH%\install\bin\git2.dll build\git2.dll

powershell -Command "Compress-Archive -Path build\* -DestinationPath build\mbt_windows_x86.zip"

dir build
