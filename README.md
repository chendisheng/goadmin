# GoAdmin

[English](./README.md) | [简体中文](./README_zh.md)

This English README mirrors `README_zh.md` and uses the same project scope, structure, and usage notes in English.

GoAdmin is a clean-room, modular full-stack platform. It uses a `server/ + web/` repository layout and focuses on reusable platform capabilities instead of a single business app.

This file is the English version of the project overview. It keeps the same structure and meaning as `README_zh.md`, so you can switch between the two languages at any time.

## Project positioning

- **Clean-room implementation**: do not reuse source code, directory habits, or interface design from external sibling projects.
- **Modular backend**: the backend follows `transport -> application -> domain -> infrastructure`.
- **Frontend/backend separation**: the frontend runs independently and consumes REST APIs plus dynamic runtime capabilities.
- **Expandable platform**: built-in support for code generation, plugins, uploads, i18n, and authorization governance.

## Main capabilities

- **Authentication and authorization**: JWT session handling, Casbin-style policy governance, dynamic menus and routes.
- **Dynamic admin pages**: user, role, menu, dictionary, upload, and plugin center pages.
- **Code generation**: module / CRUD / plugin generation, plus DSL preview and execution.
- **File uploads**: support for local, database, and object-storage drivers with preview/download flows.
- **Internationalization**: i18next-based frontend runtime with persisted language selection and dynamic loading.
- **Delivery toolchain**: Docker, Docker Compose, Kubernetes, Helm, and CI/CD support.

## Requirements

- **Go**: 1.22+
- **Node.js**: 18+ recommended for the `web/` frontend
- **npm**: for frontend dependency installation
- **Docker / Docker Compose v2**: for containerized startup and deployment

## Quick start

### 1. Prepare local environment variables

Copy the Compose environment file:

```bash
cp deploy/docker-compose/.env.example deploy/docker-compose/.env
```

If you plan to use Docker Compose, this file is the preferred place to adjust ports, image tags, or database settings.

### 2. Start with Docker Compose

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

By default, the service is exposed on port `8080`.

### 3. Develop the backend on the host

If you do not want to run the backend in Docker, use the host-development commands provided by the repository:

```bash
make host-dev
make host-build
make host-test
make host-run-cli ARGS="generate module demo"
```

To initialize the local cache directories explicitly, run:

```bash
make host-cache-init
```

Host development reuses the following local cache directories:

- `server/.cache/go-mod`
- `server/.cache/go-build`

## Directory structure

```text
.
├── server/                 # backend workspace, including service entrypoints, CLI, modules, config, and codegen
├── web/                    # frontend project based on Vue 3 + TypeScript + Vite + Pinia + Vue Router + Element Plus
├── deploy/                 # delivery assets for Docker / Compose / Kubernetes / Helm
├── docs/                   # architecture, design, requirements, and reference documents
├── memory-bank/            # project memory and progress records
├── Makefile                # root-level entry commands
└── README.md               # English project overview
```

### Backend highlights

- `server/cmd/`: server entrypoints, CLI entrypoints, and migration tools.
- `server/config/`: backend configuration files and environment-specific settings.
- `server/core/`: authentication, authorization, configuration, startup orchestration, and other core capabilities.
- `server/modules/`: modular backend implementations.
- `server/codegen/`: code generation, deletion, installation, and DSL handling.

### Frontend highlights

- `web/src/api/`: API client wrappers.
- `web/src/views/`: page views.
- `web/src/router/`: routing and dynamic menu binding.
- `web/src/store/`: Pinia state management.
- `web/src/i18n/`: internationalization resources and runtime integration.

## Configuration

The backend loads configuration with the following defaults:

- `GOADMIN_ENV` / `APP_ENV` default to `dev`
- `GOADMIN_CONFIG_DIR` defaults to `server/config`

Common config files:

- `server/config/config.yaml`
- `server/config/config.dev.yaml`
- `server/config/config.prod.yaml`

Compose environment variables are defined in:

- `deploy/docker-compose/.env.example`

## Common commands

### Backend build and test

```bash
make server-build
make server-build-cli
make test
```

### Frontend build and test

```bash
cd web
npm install
npm run build
npm run test
```

### Frontend type checking and i18n validation

```bash
cd web
npm run typecheck
npm run i18n:check-locales
```

### CLI code generation

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

## Runtime endpoints

The backend currently exposes the following common endpoints:

- `GET /api/v1/health`
- `GET /api/v1/meta/version`
- `GET /api/v1/meta/config`

> More business endpoints will be added as modules continue to grow.

## Deployment

The repository already includes the following delivery assets:

- `deploy/docker/Dockerfile`
- `deploy/docker/web.Dockerfile`
- `deploy/docker/web-nginx.conf`
- `deploy/docker-compose/docker-compose.yaml`
- `deploy/k8s/`
- `deploy/k8s/overlays/`
- `deploy/helm/goadmin/`
- `.github/workflows/ci-cd.yml`
- `.github/workflows/web-ci-cd.yml`

The root `Makefile` also provides common deployment helpers such as:

- `make dev`
- `make compose-build`
- `make compose-up`
- `make compose-build-local`
- `make compose-up-local`
- `make docker-build-server`
- `make docker-build-web`

## Documentation navigation

Recommended reading:

- `docs/GoAdmin 架构设计.md`
- `docs/GoAdmin 架构设计-简版摘要.md`

## Development conventions

- **Follow the architecture document**: implementation should prioritize `docs/GoAdmin 架构设计.md`.
- **Keep clean-room discipline**: only reference ideas from external projects, never copy implementations.
- **Use the root Makefile first**: repository entry commands should be run from the root.
- **Keep backend config under `server/config/`**.
- **Keep the frontend stack consistent**: Vue 3 + TypeScript + Vite + Pinia + Vue Router + Axios + Element Plus.

## License

This repository follows the license defined in the root `LICENSE` file.