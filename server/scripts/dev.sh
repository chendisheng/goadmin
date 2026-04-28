#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
ROOT_DIR="$(cd "${SERVER_DIR}/.." && pwd)"
GO_BIN="${GO_BIN:-go}"
GOADMIN_ENV="${GOADMIN_ENV:-dev}"
APP_ENV="${APP_ENV:-${GOADMIN_ENV}}"
GOADMIN_CONFIG_DIR="${GOADMIN_CONFIG_DIR:-${SERVER_DIR}/config}"
GOMODCACHE="${GOMODCACHE:-${SERVER_DIR}/.cache/go-mod}"
GOCACHE="${GOCACHE:-${SERVER_DIR}/.cache/go-build}"
GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"
GOSUMDB="${GOSUMDB:-off}"
ENV_FILE="${ENV_FILE:-}"

if [ -z "${ENV_FILE}" ]; then
	if [ -f "${ROOT_DIR}/deploy/docker-compose/.env" ]; then
		ENV_FILE="${ROOT_DIR}/deploy/docker-compose/.env"
	elif [ -f "${ROOT_DIR}/.env" ]; then
		ENV_FILE="${ROOT_DIR}/.env"
	fi
fi

if [ -n "${ENV_FILE}" ] && [ -f "${ENV_FILE}" ]; then
	set -a
	# shellcheck disable=SC1090
	. "${ENV_FILE}"
	set +a
fi

export GOADMIN_ENV
export APP_ENV
export GOADMIN_CONFIG_DIR
export GOMODCACHE
export GOCACHE
export GOPROXY
export GOSUMDB

cd "${SERVER_DIR}"

echo "[dev] server dir: ${SERVER_DIR}"
echo "[dev] env: ${GOADMIN_ENV}"
echo "[dev] config dir: ${GOADMIN_CONFIG_DIR}"
echo "[dev] cache dir: ${SERVER_DIR}/.cache"
echo "[dev] running: ${GO_BIN} run ./cmd/server"

exec "${GO_BIN}" run ./cmd/server "$@"
