#!/usr/bin/make -f

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	go build -o ./bin/kunitori -v ./cmd/kunitori/

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...
