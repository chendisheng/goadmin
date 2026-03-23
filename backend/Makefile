APP_NAME := goadmin
CLI_NAME := goadmin-cli
GO ?= go
PKG := ./...
BIN_DIR := bin
SERVER_PKG := ./cmd/server
CLI_PKG := ./cmd/cli

.PHONY: all dev run build build-cli run-cli test fmt tidy clean

all: test build

dev:
	bash ./scripts/dev.sh

run:
	$(GO) run $(SERVER_PKG)

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(SERVER_PKG)

build-cli:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(CLI_NAME) $(CLI_PKG)

run-cli:
	$(GO) run $(CLI_PKG) $(ARGS)

test:
	$(GO) test $(PKG)

fmt:
	$(GO) fmt $(PKG)

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)
