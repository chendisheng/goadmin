# GoAdmin

[English](./README.md) | [简体中文](./README_zh.md)

本中文 README 与 `README.md` 互为双语版本，内容结构保持一致，便于在中英文之间切换阅读。

GoAdmin 是一个 clean-room、模块化的全栈管理平台，采用 `server/ + web/` 的仓库组织方式，后端负责认证、权限、代码生成、插件、上传等核心能力，前端提供基于 Vue 3 的管理界面。

本文件是项目的中文说明，和根目录 `README.md` 保持内容一致但以中文组织，便于快速了解项目结构、启动方式和开发约定。

> **作者声明：本项目完全使用 AI 生成，一个字符的代码都没有人为改动，包括文档；人类（我）只参与了与 AI 的沟通。**

## 项目定位

- **clean-room 实现**：不直接复用外部同类项目的源码、目录和接口设计。
- **模块化后端**：后端按 `transport -> application -> domain -> infrastructure` 分层组织。
- **前后端分离**：前端独立运行，后端提供 REST API 和动态能力支撑。
- **可扩展能力**：内置代码生成、插件、上传、国际化、授权治理等能力。

## 主要能力

- **认证与授权**：JWT 登录态、Casbin 风格的权限治理、动态菜单与路由控制。
- **动态管理页**：用户、角色、菜单、字典、上传、插件中心等管理页面。
- **代码生成**：支持模块 / CRUD / 插件生成，以及 DSL 预览与执行。
- **文件上传**：支持本地、数据库、对象存储等上传驱动，并提供预览/下载能力。
- **国际化**：前端采用 i18next 生态，支持语言持久化与动态加载。
- **交付体系**：提供 Docker、Docker Compose、Kubernetes、Helm 和 CI/CD 配置。

## 环境要求

- **Go**：1.22+
- **Node.js**：建议 18+，用于 `web/` 前端构建
- **npm**：用于安装前端依赖
- **Docker / Docker Compose v2**：用于容器化启动与部署

## 快速开始

### 1. 配置本地环境变量

复制 Compose 环境文件：

```bash
cp deploy/docker-compose/.env.example deploy/docker-compose/.env
```

如果你准备使用 Docker Compose 启动，建议先生成一个本地 `.env`，这样可以按需调整端口、镜像标签和数据库参数。

### 2. 使用 Docker Compose 启动

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

默认情况下，服务会暴露在 `8080` 端口。

### 3. 使用宿主机方式开发后端

如果你不想通过 Docker 启动后端，可以使用仓库提供的宿主机开发命令：

```bash
make host-dev
make host-build
make host-test
make host-run-cli ARGS="generate module demo"
```

如果你想显式初始化本地缓存目录，可以运行：

```bash
make host-cache-init
```

宿主机开发会复用以下本地缓存目录：

- `server/.cache/go-mod`
- `server/.cache/go-build`

## 目录结构

```text
.
├── server/                 # 后端工作区，包含主服务、CLI、模块、配置与代码生成相关能力
├── web/                    # 前端工程，基于 Vue 3 + TypeScript + Vite + Pinia + Vue Router + Element Plus
├── deploy/                 # Docker / Compose / Kubernetes / Helm 等交付资源
├── docs/                   # 架构设计、需求说明、参考资料与实施拆解文档
├── memory-bank/            # 项目记忆与状态记录
├── Makefile                # 根级入口命令
└── README.md               # 英文项目说明
```

### 后端目录要点

- `server/cmd/`：服务端入口、CLI 入口与迁移入口。
- `server/config/`：后端配置文件与环境配置。
- `server/core/`：认证、授权、配置、启动引导等核心能力。
- `server/modules/`：按业务模块组织的后端实现。
- `server/codegen/`：代码生成、删除、安装与 DSL 处理相关能力。

### 前端目录要点

- `web/src/api/`：API 调用封装。
- `web/src/views/`：页面视图。
- `web/src/router/`：路由与动态菜单绑定。
- `web/src/store/`：Pinia 状态管理。
- `web/src/i18n/`：国际化资源与运行时接入。

## 配置说明

后端默认配置加载规则如下：

- `GOADMIN_ENV` / `APP_ENV` 默认值为 `dev`
- `GOADMIN_CONFIG_DIR` 默认值为 `server/config`

常用配置文件包括：

- `server/config/config.yaml`
- `server/config/config.dev.yaml`
- `server/config/config.prod.yaml`

Docker Compose 的环境变量示例位于：

- `deploy/docker-compose/.env.example`

## 常用命令

### 后端构建与测试

```bash
make server-build
make server-build-cli
make test
```

### 前端构建与测试

```bash
cd web
npm install
npm run build
npm run test
```

### 前端类型检查与国际化校验

```bash
cd web
npm run typecheck
npm run i18n:check-locales
```

### CLI 代码生成

```bash
make server-run-cli ARGS="generate module user"
make server-run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
make server-run-cli ARGS="generate plugin demo"
```

生成器会创建：

- `server/modules/<name>` 下的模块骨架
- CRUD 所需的 application / transport / infrastructure 层
- `server/core/auth/casbin/adapter/policy.csv` 中的 Casbin 策略行
- 需要时生成前端 API、路由和视图文件

## 运行时接口

当前后端已提供的常用接口包括：

- `GET /api/v1/health`
- `GET /api/v1/meta/version`
- `GET /api/v1/meta/config`

> 更多业务接口会随着模块持续扩展。

## 部署说明

仓库已经包含完整的交付资源：

- `deploy/docker/Dockerfile`
- `deploy/docker/web.Dockerfile`
- `deploy/docker/web-nginx.conf`
- `deploy/docker-compose/docker-compose.yaml`
- `deploy/k8s/`
- `deploy/k8s/overlays/`
- `deploy/helm/goadmin/`
- `.github/workflows/ci-cd.yml`
- `.github/workflows/web-ci-cd.yml`

根目录 `Makefile` 也提供了常用的部署辅助命令，例如：

- `make dev`
- `make compose-build`
- `make compose-up`
- `make compose-build-local`
- `make compose-up-local`
- `make docker-build-server`
- `make docker-build-web`

## 文档导航

推荐优先阅读以下文档：

- [`docs/GoAdmin 架构设计.md`](./docs/GoAdmin%20%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1.md)
- [`docs/GoAdmin 架构设计-简版摘要.md`](./docs/GoAdmin%20%E6%9E%B6%E6%9E%84%E8%AE%BE%E8%AE%A1-%E7%AE%80%E7%89%88%E6%91%98%E8%A6%81.md)
- [`docs/GoAdmin 功能演示.md`](./docs/GoAdmin%20%E5%8A%9F%E8%83%BD%E6%BC%94%E7%A4%BA.md)
- [`docs/GoAdmin CodeGen 模块使用说明.md`](./docs/GoAdmin%20CodeGen%20%E6%A8%A1%E5%9D%97%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E.md)

## 开发约定

- **以架构文档为准**：项目实现应优先遵循 `docs/GoAdmin 架构设计.md`。
- **保持 clean-room**：仅参考外部项目的思路，不直接复制实现。
- **优先使用根级 Makefile**：仓库入口统一从根目录执行。
- **后端配置放在 `server/config/`**。
- **前端技术栈保持统一**：Vue 3 + TypeScript + Vite + Pinia + Vue Router + Axios + Element Plus。

## 许可证

本仓库遵循项目根目录 `LICENSE` 所声明的许可协议。
