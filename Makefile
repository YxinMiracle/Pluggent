APP := pluggent
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)

GO ?= go
VERSION ?= dev
LDFLAGS := -X main.version=$(VERSION)

# 默认目标
.DEFAULT_GOAL := help

.PHONY: help fmt vet test test-race check build run clean

help:
	@printf "%s\n" \
		"Pluggent development commands:" \
		"" \
		"  make fmt        Format Go source files" \
		"  make vet        Run Go static checks" \
		"  make test       Run unit tests" \
		"  make test-race  Run unit tests with the race detector" \
		"  make check      Run required local checks" \
		"  make build      Build bin/pluggent" \
		"  make run        Run Pluggent from source" \
		"  make clean      Remove build output"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

check: vet test

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/pluggent

run:
	$(GO) run -ldflags "$(LDFLAGS)" ./cmd/pluggent

clean:
	$(RM) -r $(BIN_DIR)
