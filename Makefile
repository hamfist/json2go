TARGETS := \
  github.com/modcloth-labs/json2go \
  github.com/modcloth-labs/json2go/json2go

all: build test

build: deps
	go install -x $(TARGETS)

test:
	go test -x -v $(TARGETS)

deps:
	go get -x $(TARGETS)

.PHONY: all build deps test
