# 开发环境一键启动说明

## 推荐方式

使用 Docker Compose 一键启动完整开发环境：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up --build
```

这会同时启动：

- `mysql`
- `redis`
- `backend`

其中 `backend` 会从 `backend/config` 读取配置，并在启动时执行 Gorm 自动迁移。

## 本地脚本方式

如果你已经单独启动了 MySQL 和 Redis，也可以直接在仓库根目录执行：

```bash
./backend/scripts/dev.sh
```

这个脚本默认会把 Go 模块缓存和编译缓存放到 `backend/.cache/go-mod` 与 `backend/.cache/go-build`，避免占用 Docker 构建缓存空间。

这个脚本会：

- 自动读取 `deploy/docker-compose/.env`，若不存在则回退到根目录 `.env`
- 设置 `GOADMIN_ENV=dev`
- 设置 `GOADMIN_CONFIG_DIR=backend/config`
- 设置 `GOMODCACHE=backend/.cache/go-mod`
- 设置 `GOCACHE=backend/.cache/go-build`
- 启动 `go run ./cmd/server`

## 启动前准备

1. 复制 `deploy/docker-compose/.env.example` 为 `.env`
2. 按需修改数据库账号、端口和密码
3. 确保本地 Docker 可用

## 数据库初始化顺序

MySQL 容器首次启动时会执行：

- `deploy/docker/mysql/001-charset.sql`
- `deploy/docker/mysql/010-databases.sql`
- `deploy/docker/mysql/020-user-grants.sql`

后端启动后会再执行应用层 Gorm 迁移。
