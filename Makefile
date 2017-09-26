default: install

install: 
	go install ./...

build: 
	./build-git2go-static.sh
