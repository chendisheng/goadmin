#!/usr/bin/env bash
set -euo pipefail

HOST="${GOADMIN_HEALTHCHECK_HOST:-127.0.0.1}"
PORT="${GOADMIN_HEALTHCHECK_PORT:-8080}"
PATHNAME="${GOADMIN_HEALTHCHECK_PATH:-/api/v1/health}"

exec 3<>"/dev/tcp/${HOST}/${PORT}"
printf 'GET %s HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n' "${PATHNAME}" >&3
IFS=$'\r' read -r status <&3

if [[ "${status}" == *" 200 "* ]]; then
	exit 0
fi

exit 1
