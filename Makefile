.DEFAULT_GOAL := help

BINARY := pg-release-scraper
PKG    := ./cmd/$(BINARY)

.PHONY: build test fmt vet lint clean help

help:
	@echo "Targets:"
	@echo "  build  Build the binary into bin/"
	@echo "  test   Run all tests with race detector"
	@echo "  fmt    Format Go source files"
	@echo "  vet    Run go vet"
	@echo "  lint   Run golangci-lint (requires installation)"
	@echo "  clean  Remove built binary and output/"

build:
	go build -o bin/$(BINARY) $(PKG)

test:
	go test ./... -v -race

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/ output/
