# GoAdmin

GoAdmin is a clean-room, modular backend project. This repository currently includes the Phase 1 core backend skeleton and local development helpers.

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

1. Start the backend with Docker Compose:

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

This starts the HTTP server with the current config loading rules:

- `GOADMIN_ENV` / `APP_ENV` default to `dev`
- `GOADMIN_CONFIG_DIR` defaults to `backend/config`

The repository also ships a Phase 15 frontend delivery path under `web/` and `deploy/docker/web.Dockerfile`, which builds the Vue app into static assets and serves them through Nginx with API proxying to the backend.

### Tenant configuration

The backend exposes a runtime tenant toggle in `backend/config/*.yaml`:

- `tenant.enabled: true` enables multi-tenant behavior
- `tenant.enabled: false` degrades the system to single-tenant mode

When tenant is disabled, create/update operations ignore `tenant_id`, query paths stop injecting tenant filters, and the authorization runtime falls back to role-only evaluation.

### Authorization runtime configuration

The authorization layer supports a configurable policy source in `backend/config/*.yaml` and `deploy/docker-compose/.env.example`. The current implementation is Casbin-backed, but the module boundary is intentionally replaceable:

- `auth.casbin.source: file` loads the authorization model and policy from the local files under `core/auth/casbin/`
- `auth.casbin.source: db` loads authorization policy from the database, auto-migrates the authorization tables, and seeds the initial model/policy from the configured file paths on first boot
- Supported values for `auth.casbin.source` are `file` and `db`

For DB mode, the server must be started with the shared database connection so auth bootstrap can initialize the authorization runtime against the same store used by the rest of the backend.

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

The backend is exposed on port `8080` by default.

### Deployment artifacts

The repository also includes the delivery artifacts needed for Phase 9 and the current deployment layout:

- `deploy/docker/Dockerfile` for image builds
- `deploy/docker/web.Dockerfile` and `deploy/docker/web-nginx.conf` for frontend image builds and Nginx SPA hosting
- `deploy/docker-compose/docker-compose.yaml` for local container orchestration
- `deploy/k8s/` for plain Kubernetes manifests
- `deploy/helm/goadmin/` for Helm-based deployments
- `.github/workflows/ci-cd.yml` for the unified backend + frontend CI/CD automation
- `.github/workflows/web-ci-cd.yml` as a manual backup for frontend-only image publishing

Environment toggles are exposed through `deploy/docker-compose/.env.example` and the YAML config files under `backend/config/`, including `tenant.enabled`.

### Health check

The Compose service includes a health check that verifies the HTTP health endpoint:

- `GET /api/v1/health`

You can inspect status with:

```bash
docker compose ps
```

## Build and test

Build the backend binary:

```bash
make backend-build
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
make backend-build-cli
```

Run the generator from the backend module:

```bash
make backend-run-cli ARGS="generate module user"
make backend-run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
make backend-run-cli ARGS="generate plugin demo"
```

The generator will:

- create module skeletons under `backend/modules/<name>`
- generate CRUD application / transport / infrastructure layers
- append Casbin policy lines into `backend/core/auth/casbin/adapter/policy.csv`
- generate frontend API / router / view files when `--frontend` is enabled