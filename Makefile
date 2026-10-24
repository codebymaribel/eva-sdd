.PHONY: build test golden lint install clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/eva ./cmd/eva

test:
	go test ./...

# Regenerate golden files after changing anything under templates/.
golden:
	go test ./internal/installer -update

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

install:
	go install ./cmd/eva

clean:
	rm -rf bin
