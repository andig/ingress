.PHONY: all release

PWD := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
BIN := $(PWD)/bin

all:
	@mkdir -p bin
	./hash.sh
	GOBIN=$(BIN) go install ./...

release:
	./build.sh
