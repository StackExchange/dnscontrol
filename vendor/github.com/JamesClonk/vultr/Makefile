TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

all: prepare lint vet test build

prepare:
	go get -v github.com/golang/lint/golint
	go get -v github.com/Masterminds/glide
	go get -v github.com/goreleaser/goreleaser
	glide install

build:
	go install

lint:
	for pkg in $(TEST); do golint $$pkg; done

vet:
	@echo "go vet ."
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

check: lint vet test

release:
	goreleaser

.PHONY: all prepare build lint vet test check release
