default: install

.PHONY: install
install: build_libgit2 build_git2go
	go install ./...

.PHONY: build_libgit2
build_libgit2:
	./scripts/build_libgit2.sh

.PHONY: build_git2go
build_git2go:
	./scripts/build_git2go.sh

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
	go test -tags static,system_libgit2 -covermode=count ./e
	go test -tags static,system_libgit2 -covermode=count ./lib
	go test -tags static,system_libgit2 -covermode=count ./cmd
	go test -tags static,system_libgit2 -covermode=count ./trie
	go test -tags static,system_libgit2 -covermode=count ./intercept
	go test -tags static,system_libgit2 -covermode=count ./graph
	go test -tags static,system_libgit2 -covermode=count ./utils
	go test -tags static,system_libgit2 -covermode=count .

.PHONY: showcover
showcover:
	go tool cover -html=coverage.out

.PHONY: doc
doc:
	go run ./scripts/gendoc.go

.PHONY: lint
lint:
	gofmt -s -w **/*.go
