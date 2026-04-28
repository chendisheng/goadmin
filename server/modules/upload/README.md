# GoAdmin 文件上传基础模块

> 该文档用于说明 `upload` 基础模块的定位、能力边界、路由、权限、配置和验证方式。
>
> 设计依据：
> - `docs/teckdesign/1.GoAdmin 架构设计 v1.1.md`
> - `docs/teckdesign/26. GoAdmin 文件上传基础模块设计方案.md`
> - `docs/teckdesign/27. GoAdmin 文件上传基础模块实施任务拆解清单.md`

## 1. 模块定位

`upload` 是 GoAdmin 的基础模块之一，用于统一管理文件上传、查询、预览、下载、删除与业务绑定能力。

### 目标

- 提供统一的文件资产管理入口
- 抽象本地存储与对象存储后端
- 让业务模块复用同一套文件治理能力
- 与菜单、权限、bootstrap 机制保持一致

### 模块边界

模块负责：

- 文件上传
- 文件列表与详情查询
- 文件预览与下载
- 文件删除
- 文件与业务对象绑定 / 解绑
- 文件元数据持久化
- 存储驱动适配

模块不负责：

- 业务表单校验
- 业务对象生命周期管理
- 复杂媒体处理
- 前端页面的通用布局逻辑

## 2. 目录结构

```text
server/modules/upload/
  module.go
  bootstrap.go
  manifest.yaml
  README.md
  application/
    service/
  domain/
    model/
    repository/
  infrastructure/
    persistence/
    storage/
  transport/
    http/
      handler/
      request/
      response/
      router.go
```

## 3. 核心能力

### 文件资产

当前模块以 `FileAsset` 作为核心实体，包含：

- 文件 ID
- 租户 ID
- 原始文件名
- 存储文件名 / 存储键
- 存储驱动
- 存储路径
- 公开访问地址
- MIME 类型
- 扩展名
- 文件大小
- SHA256 校验值
- 可见性
- 绑定业务信息
- 上传人
- 状态
- 备注
- 创建 / 更新 / 删除时间

### 可见性

- `private`：私有文件，需要授权访问
- `public`：公开文件，可通过公开地址访问

### 存储驱动

当前配置层与 factory 已实现以下驱动：

- `local`
- `s3-compatible`
- `oss`
- `cos`
- `minio`

默认实现使用本地存储；`s3-compatible`、`oss`、`cos`、`minio` 通过 `server/modules/upload/infrastructure/storage/objectstore/` 下的统一对象存储驱动实现，并共享同一 storage 契约。

对象存储实现进一步拆分为更贴近云 SDK 的适配层：

- `adapter/client.go`：云 SDK 风格的对象客户端门面，负责 Put/Get/Head/Delete/Exists/PublicURL/SignedURL
- `adapter/key.go`：对象 key、URL、配置派生和通用校验 helpers
- `adapter/response.go`：对象元信息、元数据文件读写和结果回填
- `adapter/sign.go`：签名 URL 组装逻辑
- 根目录 `client.go`：只保留仓库内部的门面转发，避免上层直接依赖适配细节

这样做的目的是让后续接入真实云 SDK 时，只需要替换 `adapter/client.go` 中的 SDK 调用，不必改动上层 service、factory 和仓储契约。

## 4. HTTP 路由

模块对外暴露的 API 前缀为：`/api/v1/uploads/files`

### 路由清单

- `GET /api/v1/uploads/files`
- `GET /api/v1/uploads/files/:id`
- `POST /api/v1/uploads/files`
- `DELETE /api/v1/uploads/files/:id`
- `GET /api/v1/uploads/files/:id/download`
- `GET /api/v1/uploads/files/:id/preview`
- `POST /api/v1/uploads/files/:id/bind`
- `DELETE /api/v1/uploads/files/:id/bind`

### 路由职责

- `list`：分页查询文件列表
- `get`：查询文件详情
- `upload`：上传文件
- `delete`：删除文件及物理对象
- `download`：下载文件
- `preview`：预览文件元数据
- `bind`：绑定业务对象
- `unbind`：解绑业务对象

## 5. 菜单与权限

### 菜单

模块菜单在 manifest 中定义为：

- `文件管理`：`/system/upload`
- `文件列表`：`/system/upload/files`

### 权限点

当前 manifest 提供以下权限语义：

- `upload:file:list`
- `upload:file:create`
- `upload:file:read`
- `upload:file:download`
- `upload:file:preview`
- `upload:file:delete`
- `upload:file:bind`
- `upload:file:unbind`

### Casbin 默认路由策略

默认策略已覆盖上传模块全部路由，包含：

- 列表
- 详情
- 上传
- 删除
- 下载
- 预览
- 绑定
- 解绑

如果项目切换到 DB-backed Casbin，需要保证策略初始化与当前路由清单保持一致。

## 6. 配置说明

上传模块配置位于 `server/config/config.yaml` 对应的 `upload.storage.*` 结构中。

### 关键配置

- `upload.storage.driver`
- `upload.storage.local.base_dir`
- `upload.storage.local.public_base_url`（本地静态访问前缀，默认 `/uploads/files`）
- `upload.storage.local.use_proxy_download`
- `upload.storage.policy.max_upload_size`
- `upload.storage.policy.allowed_extensions`
- `upload.storage.policy.allowed_mime_types`
- `upload.storage.policy.visibility_default`
- `upload.storage.policy.path_prefix`

### 默认值

- 存储驱动：`local`
- 默认上传大小：`20mb`
- 默认可见性：`private`
- 默认路径前缀：`uploads`
- 本地存储默认目录：系统临时目录下的 `goadmin/uploads`

### 校验行为

配置会在启动时进行校验：

- `local` 模式要求 `base_dir` 与 `public_base_url`，且推荐使用真正可直开的静态前缀（默认 `/uploads/files`）
- `s3-compatible` 模式要求 endpoint、bucket、access key 等信息
- `oss`、`cos`、`minio` 也分别有必填项校验
- 对象存储驱动会额外对路径键、公开访问 URL 和签名 URL 做统一处理

## 7. 代码入口

### bootstrap

`server/modules/upload/bootstrap.go`

职责：

- 注册模块名称与 manifest 路径
- 执行数据库迁移
- 初始化仓储、存储驱动和服务
- 注册 HTTP 路由

### service

`server/modules/upload/application/service/service.go`

职责：

- 执行上传编排
- 校验文件大小、扩展名、MIME 类型和可见性
- 调用存储驱动落盘
- 写入文件资产元数据
- 支持删除、绑定与解绑

### router

`server/modules/upload/transport/http/router.go`

职责：

- 绑定 `/uploads/files` 路由组
- 注册列表、详情、上传、删除、预览、下载、绑定、解绑接口

## 8. 验证与测试

### 现有回归测试

- `server/modules/upload/application/service/service_test.go`
  - 上传成功与删除清理
  - 不允许的扩展名拒绝
- `server/modules/upload/domain/model/file_asset_test.go`
  - 表名与克隆行为
  - 状态枚举常量
- `server/modules/upload/module_test.go`
  - 模块元数据
  - manifest 路由 / 菜单 / 权限校验
  - HTTP 路由注册回归
- `server/core/auth/casbin/service/casbin_service_test.go`
  - 默认 Casbin 策略覆盖 upload 路由

### 推荐验证命令

```bash
cd server
go test ./core/auth/casbin/... ./modules/upload/...
```

## 9. 维护说明

- 新增存储驱动时，只需保持 `storage` 契约不变，并补齐对应配置与测试
- 新增上传路由时，需要同步更新 `manifest.yaml`、Casbin 默认策略和 README
- 如果菜单或权限名称变化，需要同步更新前端页面、后端 manifest 和相关回归测试

## 10. 关联文档

- `docs/teckdesign/26. GoAdmin 文件上传基础模块设计方案.md`
- `docs/teckdesign/27. GoAdmin 文件上传基础模块实施任务拆解清单.md`
- `docs/teckdesign/25. GoAdmin 按钮权限码命名规范.md`
