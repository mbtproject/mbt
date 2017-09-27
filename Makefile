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
	go test ./...
