# MySQL 初始化与迁移说明

## 目录用途

这个目录用于 MySQL 容器初始化脚本和数据库迁移说明。

## 文件说明

- `001-charset.sql`
  - 设置字符集和时区
- `010-databases.sql`
  - 创建 `goadmin` / `goadmin_dev` / `goadmin_prod`
- `020-user-grants.sql`
  - 创建 `goadmin` 账号并授权
- `01-init.sql`
  - 旧的合并版脚本占位，保留说明，不再执行
- `README.md`
  - 用于说明初始化 SQL 与应用迁移的职责划分

## 初始化流程

1. `docker compose -f deploy/docker-compose/docker-compose.yaml up` 启动 `mysql:8.0`
2. MySQL 容器挂载 `./deploy/docker/mysql`
3. 容器首次初始化时按文件名顺序自动执行 `001-charset.sql`、`010-databases.sql`、`020-user-grants.sql`
4. SQL 创建以下对象：
   - `goadmin`
   - `goadmin_dev`
   - `goadmin_prod`
   - `goadmin` 用户及授权

## 应用迁移流程

后端启动后会执行 Gorm 自动迁移：

- `users`
- `roles`
- `menus`

迁移入口在：

- `backend/cmd/server/main.go`

对应仓储实现：

- `backend/modules/user/infrastructure/repo/gorm.go`
- `backend/modules/role/infrastructure/repo/gorm.go`
- `backend/modules/menu/infrastructure/repo/gorm.go`

## 配置同步

如果你修改了数据库名或账号密码，请同步更新：

- `deploy/docker-compose/docker-compose.yaml`
- `deploy/docker-compose/.env.example`
- `backend/config/config.yaml`
- `backend/config/config.dev.yaml`
- `backend/config/config.prod.yaml`
- `backend/core/config/config.go`

## 重新执行初始化

MySQL 初始化脚本只会在数据卷为空时执行。

如果需要重新触发：

1. 删除 `mysql-data` volume
2. 重新启动 `docker-compose`

## 注意事项

- 不要把业务表结构初始化写进 `001-charset.sql` / `010-databases.sql` / `020-user-grants.sql`
- 业务表结构应由后端 Gorm 迁移维护
- 如果后续切换到 PostgreSQL，这里只保留说明文档即可
