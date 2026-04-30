SHELL:=/bin/bash

# Directory where this file resides. Allows running the Makefile from anywhere using
# the `-f` parameter
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
testfile ?= "./..."

all:


@PHONY: test
test:
	go test $(testfile)

@PHONY: test-cov
test-cov:
	go test -coverprofile=.cover $(testfile) \
	    && go tool cover -func=.cover \
		&& unlink .cover

@PHONY: vet
vet:
	go vet $(testfile)

@PHONY: fmt
fmt:
	test -z "$(shell go fmt . ./internal/...)"

@PHONY: build
build:
	go build . ./internal/...
