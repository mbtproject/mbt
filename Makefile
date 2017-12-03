default: install

.PHONY: install
install: 
	go install ./...

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
test:
	go test -coverprofile=coverage.out ./lib
	go test -coverprofile=coverage.out ./cmd
	go test -coverprofile=coverage.out .

.PHONY: showcover
showcover:
	go tool cover -html=coverage.out

.PHONY: doc
doc:
	go run ./scripts/gendoc.go
