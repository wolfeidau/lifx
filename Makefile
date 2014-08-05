# Program version
VERSION := $(shell grep "const Version " version.go | sed -E 's/.*"(.+)"$$/\1/')

test:
	go test -v ./...

.PHONY: test
