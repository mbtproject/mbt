default: install

install: 
	go install ./...

build: clean 
	./build.sh

clean:
	rm -rf build

restore:
	go get -t
	go get github.com/stretchr/testify

test:
	go test -coverprofile=coverage.out ./lib
	go test -coverprofile=coverage.out ./cmd
	go test -coverprofile=coverage.out .

showcover:
	go tool cover -html=coverage.out
