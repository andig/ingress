.PHONY: all release clean test run

PWD := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
BIN := $(PWD)/bin

all:
	@mkdir -p bin
	tools/hash.sh
	GOBIN=$(BIN) go install ./...

release:
	tools/build.sh

clean:
	tools/clean.sh

test:
	go test ./...
	
run:
	go run -race github.com/andig/ingress/cmd/ingress
