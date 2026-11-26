GO ?= go
PREFIX ?= $(HOME)/.local
BUILD_ID ?= $(shell date -u +%Y%m%dT%H%M%SZ)

.PHONY: build install test check clean
build:
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags='-s -w -X main.buildID=$(BUILD_ID)' -o bin/keywi .
install: build
	install -d "$(PREFIX)/bin"
	install -m 755 bin/keywi "$(PREFIX)/bin/keywi"
test:
	$(GO) test -race ./...
check:
	test -z "$$(gofmt -l .)"
	$(GO) vet ./...
	$(GO) test -race ./...
	$(GO) mod verify
	sh -n install.sh
	python3 scripts/tests/test-installer.py
clean:
	rm -rf bin dist coverage.out
