default: install

install: 
	go install ./...

build: clean 
	./build.sh

clean:
	rm -rf build
