# GoAdmin

GoAdmin is a clean-room, modular server project. This repository currently includes the Phase 1 core server skeleton and local development helpers.

## Prerequisites

- Go 1.22+
- Docker and Docker Compose v2

## Local development

1. Copy the example environment file:

```bash
cp deploy/docker-compose/.env.example deploy/docker-compose/.env
```

If you only want the default local compose values, this step is still recommended so you can tweak port, image tag, or database settings without editing the compose file.

If you prefer to keep compose-specific env files next to the compose manifest, you can also copy `deploy/docker-compose/.env.example` to `deploy/docker-compose/.env` and start with `--env-file deploy/docker-compose/.env`.

1. Start the server with Docker Compose:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

If you want to run the server outside Docker, use the host-development workflow below.

### Host development / non-Docker server development

The server Makefile and dev script use host-local Go module and build caches under `server/.cache/go-mod` and `server/.cache/go-build`.

Recommended commands:

```bash
make host-dev
make host-build
make host-test
make host-run-cli ARGS="generate module demo"
```

These targets reuse the same host-local directories for `go run`, `go build`, `go test`, `go fmt`, and `go mod tidy`.

If you want to initialize the cache directories explicitly, run:

```bash
make host-cache-init
```

This starts the HTTP server with the current config loading rules:

- `GOADMIN_ENV` / `APP_ENV` default to `dev`
- `GOADMIN_CONFIG_DIR` defaults to `server/config`

The repository also ships a Phase 15 frontend delivery path under `web/` and `deploy/docker/web.Dockerfile`, which builds the Vue app into static assets and serves them through Nginx with API proxying to the server.

### Tenant configuration

The server exposes a runtime tenant toggle in `server/config/*.yaml`:

- `tenant.enabled: true` enables multi-tenant behavior
- `tenant.enabled: false` degrades the system to single-tenant mode

When tenant is disabled, create/update operations ignore `tenant_id`, query paths stop injecting tenant filters, and the authorization runtime falls back to role-only evaluation.

### Authorization runtime configuration

The authorization layer supports a configurable policy source in `server/config/*.yaml` and `deploy/docker-compose/.env.example`. The current implementation is Casbin-backed, but the module boundary is intentionally replaceable:

- `auth.casbin.source: file` loads the authorization model and policy from the local files under `core/auth/casbin/`
- `auth.casbin.source: db` loads authorization policy from the database, auto-migrates the authorization tables, and seeds the initial model/policy from the configured file paths on first boot
- Supported values for `auth.casbin.source` are `file` and `db`

For DB mode, the server must be started with the shared database connection so auth bootstrap can initialize the authorization runtime against the same store used by the rest of the server.

### Available endpoints

- `GET /api/v1/health`
- `GET /api/v1/meta/version`
- `GET /api/v1/meta/config`

## Docker Compose

Start the service from the repository root:

```bash
cp deploy/docker-compose/.env.example deploy/docker-compose/.env
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

The server is exposed on port `8080` by default.

### Deployment artifacts

The repository also includes the delivery artifacts needed for Phase 9 and the current deployment layout:

- `deploy/docker/Dockerfile` for image builds
- `deploy/docker/web.Dockerfile` and `deploy/docker/web-nginx.conf` for frontend image builds and Nginx SPA hosting
- `deploy/docker-compose/docker-compose.yaml` for local container orchestration
- `deploy/k8s/` for plain Kubernetes manifests
- `deploy/helm/goadmin/` for Helm-based deployments
- `.github/workflows/ci-cd.yml` for the unified server + frontend CI/CD automation
- `.github/workflows/web-ci-cd.yml` as a manual backup for frontend-only image publishing

For local development, the root `Makefile` exposes the same host-development workflow through `make host-*` targets, while `make dev` continues to start the Docker Compose stack.

The root `Makefile` also exposes `make docker-build-server` and `make docker-build-web` as plain `docker build --no-cache` commands for manual image builds.

`make compose-build-local` and `make compose-up-local` are simple aliases of `make compose-build` and `make compose-up`.

Environment toggles are exposed through `deploy/docker-compose/.env.example` and the YAML config files under `server/config/`, including `tenant.enabled`.

### Health check

The Compose service includes a health check that verifies the HTTP health endpoint:

- `GET /api/v1/health`

You can inspect status with:

```bash
docker compose ps
```

## Build and test

Build the server binary:

```bash
make server-build
```

Build the frontend bundle:

```bash
cd web
npm install
npm run build
```

Build the frontend container image:

```bash
docker build -f deploy/docker/web.Dockerfile -t goadmin/web:local .
```

Run tests:

```bash
make test
```

## CLI generator

Build the CLI binary:

```bash
make server-build-cli
```

Run the generator from the server module:

```bash
make server-run-cli ARGS="generate module user"
make server-run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
make server-run-cli ARGS="generate plugin demo"
```

The generator will:

- create module skeletons under `server/modules/<name>`
- generate CRUD application / transport / infrastructure layers
- append Casbin policy lines into `server/core/auth/casbin/adapter/policy.csv`
- generate frontend API / router / view files when `--frontend` is enabled