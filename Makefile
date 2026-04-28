SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

ROOT_DIR := $(abspath .)
SERVER_DIR := $(ROOT_DIR)/server
WEB_DIR := $(ROOT_DIR)/web
COMPOSE_DIR := $(ROOT_DIR)/deploy/docker-compose
COMPOSE_FILE := $(COMPOSE_DIR)/docker-compose.yaml
COMPOSE_ENV_FILE := $(COMPOSE_DIR)/.env
COMPOSE_ENV_EXAMPLE := $(COMPOSE_DIR)/.env.example

GO ?= go
NPM ?= npm
DOCKER ?= docker
COMPOSE ?= docker compose
ARGS ?=

.PHONY: help all dev build test clean \
	server-dev server-run server-build server-build-cli server-run-cli server-test server-fmt server-tidy server-clean \
	host-dev host-run host-build host-build-cli host-run-cli host-test host-fmt host-tidy host-cache-init \
	web-install web-dev web-build web-preview web-typecheck web-clean \
	compose-init compose-up compose-down compose-logs compose-ps compose-build compose-reset \
	compose-build-local compose-up-local \
	docker-builder-init docker-build-server docker-build-web

help:
	@printf "GoAdmin root Makefile\n"
	@printf "\nCommon targets:\n"
	@printf "  make dev                 Start local compose environment\n"
	@printf "  make build               Build server and frontend bundles\n"
	@printf "  make test                Run server tests and frontend typecheck\n"
	@printf "  make clean               Clean generated build outputs\n"
	@printf "\nHost-cache targets:\n"
	@printf "  make host-dev            Run server dev server with host-local Go cache\n"
	@printf "  make host-run            Run server with host-local Go cache\n"
	@printf "  make host-build          Build server binary with host-local Go cache\n"
	@printf "  make host-build-cli      Build CLI binary with host-local Go cache\n"
	@printf "  make host-run-cli        Run CLI with host-local Go cache\n"
	@printf "  make host-test           Run server tests with host-local Go cache\n"
	@printf "  make host-fmt            Format server Go files with host-local Go cache\n"
	@printf "  make host-tidy           Run go mod tidy with host-local Go cache\n"
	@printf "  make host-cache-init     Create server/.cache directories\n"
	@printf "\nServer targets:\n"
	@printf "  make server-dev          Run server dev server via scripts/dev.sh\n"
	@printf "  make server-run          Run server with go run\n"
	@printf "  make server-build        Build server binary\n"
	@printf "  make server-build-cli    Build CLI binary\n"
	@printf "  make server-run-cli      Run CLI, e.g. make server-run-cli ARGS=\"generate plugin demo\"\n"
	@printf "  make server-test         Run server tests\n"
	@printf "  make server-fmt          Format server Go files\n"
	@printf "  make server-tidy         Run go mod tidy in server\n"
	@printf "  make server-clean        Remove server bin outputs\n"
	@printf "\nFrontend targets:\n"
	@printf "  make web-install         Install frontend dependencies\n"
	@printf "  make web-dev             Run Vite dev server\n"
	@printf "  make web-build           Build frontend bundle\n"
	@printf "  make web-preview         Preview frontend bundle\n"
	@printf "  make web-typecheck       Run Vue type check\n"
	@printf "  make web-clean           Remove frontend dist output\n"
	@printf "\nCompose targets:\n"
	@printf "  make compose-init        Ensure deploy/docker-compose/.env exists\n"
	@printf "  make compose-up          Start docker compose stack\n"
	@printf "  make compose-down        Stop docker compose stack\n"
	@printf "  make compose-logs        Tail docker compose logs\n"
	@printf "  make compose-ps          Show docker compose status\n"
	@printf "  make compose-build       Build docker compose services\n"
	@printf "  make compose-build-local Build docker compose services (alias of compose-build)\n"
	@printf "  make compose-up-local    Start docker compose stack (alias of compose-up)\n"
	@printf "  make compose-reset       Stop compose and remove volumes\n"
	@printf "\nDocker targets:\n"
	@printf "  make docker-builder-init  Create and select a docker-container buildx builder\n"
	@printf "  make docker-build-server  Build server image without cache\n"
	@printf "  make docker-build-web     Build frontend image without cache\n"

all: test build

dev: compose-up

build: server-build web-build

test: server-test web-typecheck

clean: server-clean web-clean

server-dev:
	$(MAKE) -C $(SERVER_DIR) dev

server-run:
	$(MAKE) -C $(SERVER_DIR) run

server-build:
	$(MAKE) -C $(SERVER_DIR) build

server-build-cli:
	$(MAKE) -C $(SERVER_DIR) build-cli

server-run-cli:
	$(MAKE) -C $(SERVER_DIR) run-cli ARGS="$(ARGS)"

server-test:
	$(MAKE) -C $(SERVER_DIR) test

server-fmt:
	$(MAKE) -C $(SERVER_DIR) fmt

server-tidy:
	$(MAKE) -C $(SERVER_DIR) tidy

server-clean:
	$(MAKE) -C $(SERVER_DIR) clean

host-dev:
	$(MAKE) server-dev

host-run:
	$(MAKE) server-run

host-build:
	$(MAKE) server-build

host-build-cli:
	$(MAKE) server-build-cli

host-run-cli:
	$(MAKE) server-run-cli ARGS="$(ARGS)"

host-test:
	$(MAKE) server-test

host-fmt:
	$(MAKE) server-fmt

host-tidy:
	$(MAKE) server-tidy

host-cache-init:
	$(MAKE) -C $(SERVER_DIR) cache-init

web-install:
	cd $(WEB_DIR) && $(NPM) install

web-dev:
	cd $(WEB_DIR) && $(NPM) run dev

web-build:
	cd $(WEB_DIR) && $(NPM) run build

web-preview:
	cd $(WEB_DIR) && $(NPM) run preview

web-typecheck:
	cd $(WEB_DIR) && $(NPM) run typecheck

web-clean:
	rm -rf $(WEB_DIR)/dist

compose-init:
	@if [ ! -f "$(COMPOSE_ENV_FILE)" ]; then \
		cp "$(COMPOSE_ENV_EXAMPLE)" "$(COMPOSE_ENV_FILE)"; \
		echo "Created $(COMPOSE_ENV_FILE) from example"; \
	fi

compose-up: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" up --build

compose-down: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" down

compose-logs: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" logs -f

compose-ps: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" ps

compose-build: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" build

compose-build-local: compose-init
	$(MAKE) compose-build

compose-up-local: compose-init
	$(MAKE) compose-up

compose-reset: compose-init
	$(COMPOSE) --env-file "$(COMPOSE_ENV_FILE)" -f "$(COMPOSE_FILE)" down -v

docker-builder-init:
	$(DOCKER) buildx create --name goadmin-builder --driver docker-container --bootstrap --use

docker-build-server:
	$(DOCKER) build --no-cache -f deploy/docker/Dockerfile -t goadmin/server:local .

docker-build-web:
	$(DOCKER) build --no-cache -f deploy/docker/web.Dockerfile -t goadmin/web:local .
