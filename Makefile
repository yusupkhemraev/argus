.PHONY: build dev clean web-build go-build build-linux build-all

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build: web-build go-build

web-build:
	cd web && npm run build

go-build:
	go build $(LDFLAGS) -o bin/argus ./cmd/argus

dev:
	cd web && npm run dev

clean:
	rm -rf bin/ web/dist/

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/argus-linux-amd64 ./cmd/argus

build-all: web-build
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o bin/argus-linux-amd64   ./cmd/argus
	GOOS=linux  GOARCH=arm64 go build $(LDFLAGS) -o bin/argus-linux-arm64   ./cmd/argus
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/argus-darwin-amd64  ./cmd/argus
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/argus-darwin-arm64  ./cmd/argus
