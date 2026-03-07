BINARY     := ipinfo
CMD        := ./cmd/ipinfo
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"
BUILD_DIR  := dist

.PHONY: all build clean test lint install

all: build

## build: compile for the current platform
build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

## install: install to GOPATH/bin
install:
	go install $(LDFLAGS) $(CMD)

## test: run all tests
test:
	go test ./...

## lint: run go vet
lint:
	go vet ./...

## cross: build for all release platforms
cross:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64   $(CMD)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64   $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64  $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64  $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe $(CMD)
	cd $(BUILD_DIR) && sha256sum * > checksums.txt

## clean: remove build artefacts
clean:
	rm -rf $(BINARY) $(BUILD_DIR)

help:
	@grep -E '^## ' Makefile | sed 's/## //'
