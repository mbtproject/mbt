default: install

.PHONY: install
install: build_libgit2
	go install ./...

.PHONY: build_libgit2
build_libgit2:
	./scripts/build_libgit2.sh

.PHONY: build
build: clean 
	./scripts/build.sh

.PHONY: clean
clean:
	rm -rf build

.PHONY: restore
restore:
	go get -t
	go get github.com/stretchr/testify

.PHONY: test
test: build_libgit2
	go test -coverprofile=coverage.out ./lib
	go test -coverprofile=coverage.out ./cmd
	go test -coverprofile=coverage.out .

.PHONY: showcover
showcover:
	go tool cover -html=coverage.out

.PHONY: doc
doc:
	go run ./scripts/gendoc.go

.PHONY: lint
lint:
	gofmt -s -w **/*.go
