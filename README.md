# GoAdmin

GoAdmin is a clean-room, modular backend project. This repository currently includes the Phase 1 core backend skeleton and local development helpers.

## Prerequisites

- Go 1.22+
- Docker and Docker Compose v2

## Local development

1. Copy the example environment file:

```bash
cp .env.example .env
```

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

When tenant is disabled, create/update operations ignore `tenant_id`, query paths stop injecting tenant filters, and Casbin authorization falls back to role-only evaluation.

### Available endpoints

- `GET /api/v1/health`
- `GET /api/v1/meta/version`
- `GET /api/v1/meta/config`

## Docker Compose

Start the service from the repository root:

```bash
cp .env.example .env
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

Environment toggles are exposed through `.env.example` and the YAML config files under `backend/config/`, including `tenant.enabled`.

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
make -C backend build
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
make -C backend test
```

## CLI generator

Build the CLI binary:

```bash
make -C backend build-cli
```

Run the generator from the backend module:

```bash
make -C backend run-cli ARGS="generate module user"
make -C backend run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
make -C backend run-cli ARGS="generate plugin demo"
```

The generator will:

- create module skeletons under `backend/modules/<name>`
- generate CRUD application / transport / infrastructure layers
- append Casbin policy lines into `backend/core/auth/casbin/adapter/policy.csv`
- generate frontend API / router / view files when `--frontend` is enabled