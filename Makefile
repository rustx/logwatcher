SHELL := /bin/bash
BIN := logwatcher

default: all

gofmt:
	gofmt -w .

gotest:
	go test -v --covermode=count -coverprofile=/tmp/count.out

gocover:
	go tool cover -func=/tmp/count.out

gohtml:
	go tool cover -html=/tmp/count.out

gobuild:
	go build -i -o ${GOPATH}/bin/${BIN} .

test: gofmt gotest gocover gohtml

build: gobuild

all: test build