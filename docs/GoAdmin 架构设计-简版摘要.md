# GoAdmin 架构设计简版摘要

> 这是一份面向快速阅读的摘要版文档，用于在不展开全部细节的情况下，快速理解 GoAdmin 的总体架构、技术选型和演进方向。

---

# 1. 项目定位

GoAdmin 是一个面向企业级后台管理平台的 clean-room 项目，采用 `server/ + web/` 的仓库结构，目标是构建一个可持续演进的平台内核，而不是单纯的 CRUD 系统。

## 核心能力

- 认证与授权
- 动态菜单与路由
- 用户、角色、菜单、字典等基础模块
- 插件扩展
- 文件上传
- 代码生成与删除
- 国际化
- DevOps 交付

---

# 2. 架构原则

## 2.1 Clean-Room

- 不复用外部项目源码
- 不沿用外部项目目录组织
- 不复用外部项目接口和表结构
- 只参考通用工程范式

## 2.2 模块优先

- 模块是第一公民
- 分层是模块内部实现
- 模块之间通过契约、服务或事件协作

## 2.3 分层架构

后端统一采用四层：

- `transport`
- `application`
- `domain`
- `infrastructure`

依赖方向必须单向：

- `transport` → `application` → `domain`
- `infrastructure` 实现 `domain` 定义的接口

---

# 3. 总体架构

```text
前端（Vue 3）
    ↓
HTTP 入口（transport）
    ↓
应用层（application）
    ↓
领域层（domain）
    ↓
基础设施层（infrastructure）
```

横切能力由 `core/` 提供：

- 配置
- 日志
- 认证与授权
- 事件总线
- 国际化
- 多租户
- 启动编排

---

# 4. 目录规范

## 4.1 顶层目录

- `server/`：后端工作区
- `web/`：前端工程
- `deploy/`：部署资源
- `docs/`：设计与说明文档
- `Makefile`：根级入口命令

## 4.2 后端结构

- `server/core/`：平台级能力
- `server/modules/`：业务模块
- `server/plugin/`：插件系统
- `server/codegen/`：代码生成与删除
- `server/cmd/`：服务与 CLI 入口

## 4.3 前端结构

- `web/src/api/`：API 封装
- `web/src/router/`：路由与菜单
- `web/src/store/`：状态管理
- `web/src/views/`：页面视图
- `web/src/i18n/`：国际化资源

---

# 5. 运行与启动

标准启动流程：

1. 读取配置
2. 初始化日志
3. 打开数据库
4. 初始化鉴权运行时
5. 初始化认证服务
6. 初始化事件总线
7. 迁移模块
8. 安装 manifest
9. 加载插件
10. 启动 HTTP 服务

配置遵循：

- 默认值 + 文件 + 环境变量
- 后端配置位于 `server/config/`
- 生产、开发、默认环境分文件管理

---

# 6. 扩展能力

## 6.1 插件系统

插件可以扩展：

- 路由
- 菜单
- 权限
- 页面
- 事件订阅

## 6.2 事件驱动

适合用于：

- 用户创建后的同步处理
- 权限变更后的刷新通知
- 插件加载后的同步动作

## 6.3 代码生成

生成器支持：

- 模块生成
- CRUD 生成
- 插件生成
- 预览与执行
- 删除与清理

---

# 7. 前端、上传与交付

## 7.1 前端

前端采用：

- Vue 3
- TypeScript
- Vite
- Pinia
- Vue Router
- Axios
- Element Plus
- i18next 生态

## 7.2 上传

上传采用“元数据 + 存储驱动”模式，支持本地、数据库和对象存储等实现。

## 7.3 DevOps

仓库提供：

- Docker / Compose
- Kubernetes manifests
- Helm chart
- CI/CD workflow

---

# 8. 演进方向

GoAdmin 当前的核心目标是保持：

- 清晰的模块边界
- 稳定的统一契约
- 可扩展的插件体系
- 可维护的代码生成体系
- 可交付的部署能力

未来可以在不破坏主干的前提下，继续扩展多租户、事件驱动和多框架适配能力。
