# Program version
VERSION := $(shell grep "const Version " version.go | sed -E 's/.*"(.+)"$$/\1/')

default: test

deps:
	go get -d -v ./...

test: deps
	go test -v ./...

.PHONY: test
