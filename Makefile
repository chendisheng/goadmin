SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

ROOT_DIR := $(abspath .)
BACKEND_DIR := $(ROOT_DIR)/backend
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
	backend-dev backend-run backend-build backend-build-cli backend-run-cli backend-test backend-fmt backend-tidy backend-clean \
	host-dev host-run host-build host-build-cli host-run-cli host-test host-fmt host-tidy host-cache-init \
	web-install web-dev web-build web-preview web-typecheck web-clean \
	compose-init compose-up compose-down compose-logs compose-ps compose-build compose-reset \
	compose-build-local compose-up-local \
	docker-builder-init docker-build-backend docker-build-web

help:
	@printf "GoAdmin root Makefile\n"
	@printf "\nCommon targets:\n"
	@printf "  make dev                 Start local compose environment\n"
	@printf "  make build               Build backend and frontend bundles\n"
	@printf "  make test                Run backend tests and frontend typecheck\n"
	@printf "  make clean               Clean generated build outputs\n"
	@printf "\nHost-cache targets:\n"
	@printf "  make host-dev            Run backend dev server with host-local Go cache\n"
	@printf "  make host-run            Run backend server with host-local Go cache\n"
	@printf "  make host-build          Build backend binary with host-local Go cache\n"
	@printf "  make host-build-cli      Build CLI binary with host-local Go cache\n"
	@printf "  make host-run-cli        Run CLI with host-local Go cache\n"
	@printf "  make host-test           Run backend tests with host-local Go cache\n"
	@printf "  make host-fmt            Format backend Go files with host-local Go cache\n"
	@printf "  make host-tidy           Run go mod tidy with host-local Go cache\n"
	@printf "  make host-cache-init     Create backend/.cache directories\n"
	@printf "\nBackend targets:\n"
	@printf "  make backend-dev         Run backend dev server via scripts/dev.sh\n"
	@printf "  make backend-run         Run backend server with go run\n"
	@printf "  make backend-build       Build backend binary\n"
	@printf "  make backend-build-cli   Build CLI binary\n"
	@printf "  make backend-run-cli     Run CLI, e.g. make backend-run-cli ARGS=\"generate plugin demo\"\n"
	@printf "  make backend-test        Run backend tests\n"
	@printf "  make backend-fmt         Format backend Go files\n"
	@printf "  make backend-tidy        Run go mod tidy in backend\n"
	@printf "  make backend-clean       Remove backend bin outputs\n"
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
	@printf "  make docker-build-backend Build backend image without cache\n"
	@printf "  make docker-build-web     Build frontend image without cache\n"

all: test build

dev: compose-up

build: backend-build web-build

test: backend-test web-typecheck

clean: backend-clean web-clean

backend-dev:
	$(MAKE) -C $(BACKEND_DIR) dev

backend-run:
	$(MAKE) -C $(BACKEND_DIR) run

backend-build:
	$(MAKE) -C $(BACKEND_DIR) build

backend-build-cli:
	$(MAKE) -C $(BACKEND_DIR) build-cli

backend-run-cli:
	$(MAKE) -C $(BACKEND_DIR) run-cli ARGS="$(ARGS)"

backend-test:
	$(MAKE) -C $(BACKEND_DIR) test

backend-fmt:
	$(MAKE) -C $(BACKEND_DIR) fmt

backend-tidy:
	$(MAKE) -C $(BACKEND_DIR) tidy

backend-clean:
	$(MAKE) -C $(BACKEND_DIR) clean

host-dev:
	$(MAKE) backend-dev

host-run:
	$(MAKE) backend-run

host-build:
	$(MAKE) backend-build

host-build-cli:
	$(MAKE) backend-build-cli

host-run-cli:
	$(MAKE) backend-run-cli ARGS="$(ARGS)"

host-test:
	$(MAKE) backend-test

host-fmt:
	$(MAKE) backend-fmt

host-tidy:
	$(MAKE) backend-tidy

host-cache-init:
	$(MAKE) -C $(BACKEND_DIR) cache-init

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

docker-build-backend:
	$(DOCKER) build --no-cache -f deploy/docker/Dockerfile -t goadmin/backend:local .

docker-build-web:
	$(DOCKER) build --no-cache -f deploy/docker/web.Dockerfile -t goadmin/web:local .
