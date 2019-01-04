SHELL:=/bin/bash

default:
	make clean && make build && make test

prepare:
	if ! [[ $$(which go) ]]; then \
		echo Go not found, please make sure that go is installed and in the \$$PATH; \
		echo For example, on ubuntu this package would be obtained with: apt install golang; \
	else \
		echo "Go found, should be OK to run: make build && make install"; \
	fi

clean:
	rm -rf tunnelfun*

build:
	make build_x64 && make build_arm && make build_386

build_x64:
	go build

build_arm:
	env GOOS=linux GOARCH=arm GOARM=5 go build -o tunnelfun.arm

build_386:
	env GOOS=linux GOARCH=386 go build -o tunnelfun.32bit

test:
	go test

format:
	go fmt
