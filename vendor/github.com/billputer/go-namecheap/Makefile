.PHONY: all fmt vet lint build test
.DEFAULT: default

all: build fmt lint test vet

build:
	@echo "+ $@"
	@go build -tags "$(BUILDTAGS) cgo" .

fmt:
	@echo "+ $@"
	@gofmt -l . | grep -v vendor | tee /dev/stderr

lint:
	@echo "+ $@"
	@golint ./... | grep -v vendor | tee /dev/stderr

test:
	@echo "+ $@"
	@go test -v

vet:
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor)
