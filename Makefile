#!/usr/bin/make -f

.PHONY: test
test:
	go test -v ./internal/...

.PHONY: build
build:
	go build -v ./cmd/kunitori/

.PHONY: fmt
fmt:
	go fmt -x ./internal/... ./cmd/...

.PHONY: vet
vet:
	go vet ./internal/... ./cmd/...
