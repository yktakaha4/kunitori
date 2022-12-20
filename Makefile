#!/usr/bin/make -f

.PHONY: test
test:
	go test -v -short ./...

.PHONY: testall
testall:
	go test -v ./...

.PHONY: build
build:
	go build -v ./cmd/kunitori/

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...
