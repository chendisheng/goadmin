APP_NAME := goadmin
CLI_NAME := goadmin-cli
GO ?= go
PKG := ./...
BIN_DIR := bin
SERVER_PKG := ./cmd/server
CLI_PKG := ./cmd/cli
CACHE_DIR ?= $(abspath .cache)
GOMODCACHE ?= $(CACHE_DIR)/go-mod
GOCACHE ?= $(CACHE_DIR)/go-build
GOPROXY ?= https://proxy.golang.org,direct
GOSUMDB ?= off
GOENV := GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) GOSUMDB=$(GOSUMDB)

.PHONY: all dev run build build-cli run-cli test fmt tidy clean cache-init

all: test build

dev:
	bash ./scripts/dev.sh

run:
	$(GOENV) $(GO) run $(SERVER_PKG)

build:
	$(MAKE) cache-init
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/$(APP_NAME) $(SERVER_PKG)

build-cli:
	$(MAKE) cache-init
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/$(CLI_NAME) $(CLI_PKG)

run-cli:
	$(GOENV) $(GO) run $(CLI_PKG) $(ARGS)

test:
	$(MAKE) cache-init
	$(GOENV) $(GO) test $(PKG)

fmt:
	$(GOENV) $(GO) fmt $(PKG)

tidy:
	$(MAKE) cache-init
	$(GOENV) $(GO) mod tidy

clean:
	rm -rf $(BIN_DIR)

cache-init:
	mkdir -p $(GOMODCACHE) $(GOCACHE)
