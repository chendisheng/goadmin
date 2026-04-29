# GoAdmin CodeGen 模块使用说明

> **用途**：本文面向 GoAdmin 的实际使用者，说明 CodeGen 模块如何从 CLI、HTTP API 或前端控制台完成代码生成、数据库反查生成和删除卸载。
> **当前能力范围**：
> - `module` / `crud` / `plugin` 生成
> - `dsl` 驱动生成
> - `db` 驱动预览与生成
> - `remove` 删除预览与执行
> - DSL 下载交付、数据库下载交付、生成产物下载

---

# 1. 先理解 CodeGen 的三个入口

GoAdmin 的 CodeGen 目前有三个主要入口：

- **CLI**：适合本地开发、脚本化执行、CI 任务。
- **HTTP API**：适合前端控制台、远程调用、服务化部署。
- **前端控制台**：适合在系统界面里完成 DSL 编辑、预览、生成与删除。

CodeGen 的核心目标不是单纯“写文件”，而是把下面这些动作统一起来：

- 读取输入源
- 生成计划
- 预览差异
- 写入文件
- 追加或同步权限策略
- 安装菜单、路由和运行时元数据
- 删除已生成资产

---

# 2. 当前支持的命令与接口总览

## 2.1 CLI 命令总览

CLI 入口对应 `server/cmd/cli`，命令形态如下：

```text
goadmin-cli generate module <name> [--force]
goadmin-cli generate crud <name> [--fields name:string,status:string] [--primary id] [--index name] [--unique code] [--frontend] [--policy] [--force]
goadmin-cli generate plugin <name> [--force]
goadmin-cli generate dsl <dsl.yaml> [--force]
goadmin-cli generate db preview --driver mysql --dsn "..." --database goadmin [--table books] [--schema public] [--generate_frontend] [--generate_policy]
goadmin-cli generate db generate --driver mysql --dsn "..." --database goadmin [--table books] [--schema public] [--generate_frontend] [--generate_policy]
goadmin-cli remove preview <module> [--kind crud] [--force] [--with-policy] [--with-runtime] [--with-frontend] [--with-registry] [--policy-store csv|db]
goadmin-cli remove execute <module> [--kind crud] [--force] [--with-policy] [--with-runtime] [--with-frontend] [--with-registry] [--policy-store csv|db]
```

如果你习惯从根目录 `Makefile` 调用，可以用：

```bash
make server-run-cli ARGS="generate module user"
make server-run-cli ARGS="generate dsl docs/examples/demo.yaml"
make server-run-cli ARGS="generate db preview --driver mysql --dsn '...' --database goadmin --table books"
make server-run-cli ARGS="remove preview book --policy-store db"
```

## 2.2 HTTP API 总览

当前 CodeGen HTTP 路由前缀是 `/api/v1/codegen`，主要接口包括：

- `POST /api/v1/codegen/dsl/preview`
- `POST /api/v1/codegen/dsl/generate`
- `POST /api/v1/codegen/dsl/generate-download`
- `POST /api/v1/codegen/db/preview`
- `POST /api/v1/codegen/db/generate`
- `POST /api/v1/codegen/db/generate-download`
- `POST /api/v1/codegen/delete/preview`
- `POST /api/v1/codegen/delete/execute`
- `POST /api/v1/codegen/install/manifest`
- `GET /api/v1/codegen/artifacts/:taskID`

---

# 3. 运行前准备

在开始使用 CodeGen 之前，建议先确认以下前提：

- **项目根目录正确**：CLI 和服务端都需要能定位仓库根目录。
- **后端配置可加载**：数据库和认证配置应可正常读取。
- **数据库可连接**：尤其是 `db` 驱动和删除模式中的 policy 清理。
- **前端依赖已安装**：如果你要在页面中操作 CodeGen 控制台。
- **权限已配置**：CodeGen 相关接口通常会受到权限控制。

如果你是本地开发，推荐先确认这些入口：

```bash
make host-dev
make host-test
make server-run-cli ARGS="generate module demo"
```

---

# 4. CodeGen 的三种核心模式

## 4.1 DSL 驱动生成

DSL 模式适合“设计态驱动生成”。

特点：

- 输入来源是 YAML/DSL 文件
- 适合纳入 Git 管理
- 便于评审、回放和版本控制
- 最适合团队协作和稳定复现

适用场景：

- 从设计文档生成模块
- 从手工编写 DSL 快速生成 CRUD
- 需要先预览再执行的场景
- 需要生成下载包的远程服务场景

## 4.2 DB 驱动生成

DB 模式适合“从现有数据库结构反查生成”。

特点：

- 从数据库表结构读取元数据
- 自动提取字段、主键、索引、注释等信息
- 适合存量数据库的代码补齐
- 也适合从运行态结构导出生成草案

适用场景：

- 已有业务表，需要快速生成 CRUD
- 想把数据库结构转成标准化代码骨架
- 想先预览数据库反查结果，再决定是否落盘

## 4.3 删除模式

删除模式用于清理 CodeGen 生成的模块资产。

特点：

- 先预览，再执行
- 默认只清理生成器可确认归属的资源
- 支持清理源码、运行时菜单、权限、注册项和前端生成物
- 支持 `csv` 和 `db` 两种 policy 存储模式

适用场景：

- 模块试验失败，需要回收生成结果
- 业务模块已废弃，需要卸载
- 需要清理系统菜单、权限和策略残留

---

# 5. 模块生成模式

虽然本文重点是 DSL、DB 和删除模式，但先给出最常用的三个生成命令，方便整体理解。

## 5.1 生成模块骨架

```bash
make server-run-cli ARGS="generate module user"
```

用途：

- 生成模块目录骨架
- 创建模块级基础文件
- 生成基础清单或元数据入口

## 5.2 生成 CRUD

```bash
make server-run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
```

用途：

- 生成领域模型相关代码
- 生成前端页面和路由
- 追加 Casbin policy

常见参数：

- `--fields`：字段定义
- `--primary`：主键字段
- `--index`：索引字段
- `--unique`：唯一字段
- `--frontend`：生成前端脚手架
- `--policy`：追加策略行
- `--force`：允许覆盖受控区块或已有文件

## 5.3 生成插件

```bash
make server-run-cli ARGS="generate plugin demo"
```

用途：

- 生成插件骨架
- 生成插件所需的基础文件和策略结构

---

# 6. DSL 驱动方式详解

## 6.1 DSL 模式是什么

DSL 模式是 CodeGen 的首选输入方式之一。

你可以把 DSL 理解成一份“机器可读的生成说明书”。它描述：

- 要生成什么模块
- 模块里有哪些实体
- 字段怎么定义
- 是否需要前端页面
- 是否需要权限策略
- 路由和菜单如何挂载

DSL 的优势是：

- 易于版本化
- 易于审阅
- 易于重复执行
- 易于和文档、示例一起保存

## 6.2 DSL 的基本流程

DSL 模式在内部通常会经历下面几步：

1. 读取 YAML/DSL 文件
2. 解析为标准文档对象
3. 解析文档中的资源列表
4. 交给 planner 做校验和预览
5. `dry-run` 时只输出结果，不落盘
6. 真实执行时写入文件并生成相关资产

## 6.3 DSL CLI 用法

### 6.3.1 基本生成

```bash
make server-run-cli ARGS="generate dsl path/to/codegen.yaml"
```

### 6.3.2 强制生成

```bash
make server-run-cli ARGS="generate dsl path/to/codegen.yaml --force"
```

### 6.3.3 预览模式

当前实现支持 `--dry-run` 预览：

```bash
make server-run-cli ARGS="generate dsl path/to/codegen.yaml --dry-run"
```

说明：

- 预览模式不会写入文件
- 会返回资源级执行计划
- 适合在正式生成前检查输出范围

## 6.4 DSL HTTP 用法

### 6.4.1 预览

`POST /api/v1/codegen/dsl/preview`

请求体示例：

```json
{
  "dsl": "module: demo\nkind: module\n",
  "force": false
}
```

### 6.4.2 生成

`POST /api/v1/codegen/dsl/generate`

请求体和预览一致。

### 6.4.3 生成并下载

`POST /api/v1/codegen/dsl/generate-download`

这个接口适合远程服务端场景：

- 先在服务端生成到临时工作区
- 再打包为下载产物
- 最后由前端或用户本地保存

### 6.4.4 DSL 输入建议

建议 DSL 采用 YAML，并尽量满足以下原则：

- 一个模块一个文件
- 文件名与模块名保持一致
- 每个资源都有清晰的 `kind` 和 `name`
- 需要重复生成的内容尽量保持稳定顺序
- 不要把临时值和环境相关值写进 DSL

## 6.5 DSL 使用建议

- **适合设计先行**：先把结构写清楚，再生成代码。
- **适合团队协作**：DSL 可以像配置文件一样进行 review。
- **适合代码回放**：同一份 DSL 可以在不同环境重放。
- **适合下载交付**：远程服务器生成后可以直接导出。

## 6.6 DSL 常见注意点

- `force` 不等于“无脑覆盖”，受控区块仍应谨慎处理。
- DSL 文件内容应保持可解析、可版本化。
- 预览与生成应尽量使用同一份输入，避免结果偏差。

---

# 7. DB 驱动方式详解

## 7.1 DB 模式是什么

DB 模式用于从已有数据库结构反查生成代码。

核心思路是：

- 读取数据库表结构
- 整理为统一中间模型
- 再转回 DSL / IR / 生成计划
- 复用现有生成链路

也就是说，DB 不是另一套完全独立的生成器，而是 CodeGen 的一个输入源。

## 7.2 DB CLI 用法

### 7.2.1 预览

```bash
make server-run-cli ARGS="generate db preview --driver sqlite --dsn 'file:./tmp/codegen.db?cache=shared&mode=rwc' --database codegen --table books --generate_frontend --generate_policy"
```

### 7.2.2 生成

```bash
make server-run-cli ARGS="generate db generate --driver sqlite --dsn 'file:./tmp/codegen.db?cache=shared&mode=rwc' --database codegen --table books --generate_frontend --generate_policy"
```

### 7.2.3 参数说明

- `--driver`
  - 数据库驱动名
  - 例如 `mysql`、`sqlite`
- `--dsn`
  - 数据库连接串
  - CLI 模式必须提供
- `--database`
  - 数据库名
  - 必填
- `--schema`
  - 可选 schema 名
  - 对 PostgreSQL 等场景更常见
- `--table`
  - 指定要生成的表
  - 可重复传入多个表
- `--force`
  - 允许更强覆盖
- `--generate_frontend`
  - 是否生成前端脚手架
- `--generate_policy`
  - 是否生成 policy 相关内容
- `--mount_parent_path`
  - 生成菜单挂载的父级路径

## 7.3 DB HTTP 用法

### 7.3.1 预览

`POST /api/v1/codegen/db/preview`

### 7.3.2 生成

`POST /api/v1/codegen/db/generate`

### 7.3.3 生成并下载

`POST /api/v1/codegen/db/generate-download`

## 7.4 HTTP 与 CLI 的差异

DB 模式在 CLI 和 HTTP 中有一个很重要的区别：

- **CLI**：需要你提供 `--dsn`
- **HTTP**：使用服务端已经建立好的数据库连接，不需要在请求里传 `dsn`

HTTP 请求体示例：

```json
{
  "driver": "mysql",
  "database": "goadmin",
  "schema": "public",
  "tables": ["books"],
  "force": false,
  "generate_frontend": true,
  "generate_policy": true,
  "mount_parent_path": "/system"
}
```

## 7.5 DB 模式的适用边界

适合：

- 现有数据库反查生成
- 快速补齐 CRUD
- 把已有表结构转成统一代码骨架

不适合：

- 把数据库当成唯一配置来源而不做审查
- 忽略字段命名和注释的可维护性
- 在不明确归属的情况下直接大规模覆盖已有文件

## 7.6 DB 模式的使用建议

- 先 `preview`
- 再决定是否 `generate`
- 先选少量表做验证
- 确认 `mount_parent_path` 是否符合菜单挂载结构
- 生成前先确认表结构命名是否满足项目规范

---

# 8. 删除模式详解

## 8.1 删除模式是什么

删除模式用于回收 CodeGen 生成的模块资产。

它不是简单的“删目录”，而是先做删除计划，再执行删除，并在结束后做验证。

删除模式会尽量处理这些内容：

- 生成的源码文件
- 生成的前端页面和路由
- 生成的菜单、权限和运行时注册项
- 生成的 policy 记录
- 空目录清理

## 8.2 删除模式的基本流程

### 8.2.1 预览

```bash
make server-run-cli ARGS="remove preview book --with-policy --with-runtime --with-frontend --with-registry --policy-store db"
```

### 8.2.2 执行

```bash
make server-run-cli ARGS="remove execute book --with-frontend --with-registry"
```

### 8.2.3 参数说明

- `--kind`
  - 删除类型
  - 默认 `crud`
  - 常见值：`crud`、`module`、`plugin`
- `--force`
  - 允许在部分非致命冲突下继续
- `--with-policy`
  - 是否删除 policy 相关内容
- `--with-runtime`
  - 是否删除运行时菜单、权限、注册项
- `--with-frontend`
  - 是否删除前端生成物
- `--with-registry`
  - 是否删除 bootstrap 注册项
- `--policy-store`
  - policy 存储模式
  - 支持 `csv` 和 `db`

## 8.3 删除模式的安全原则

删除模式默认遵循以下原则：

- **先预览，后执行**
- **只删生成器能确认归属的资产**
- **不默认删除业务数据**
- **不默认删除人工手写代码**
- **共享资源优先保护**

## 8.4 policy 清理说明

`--policy-store` 需要和当前系统的 policy 存储模式一致：

- `csv`
  - 直接清理文件侧 policy 记录
- `db`
  - 清理数据库中的 policy 记录

如果你不确定当前使用的是哪种模式，先查看系统配置或先做 `preview`。

## 8.5 删除模式的适用场景

- 临时模块试验后回收
- CRUD 页面验证后清理
- 插件模块卸载
- 权限和菜单同步清理

## 8.6 删除模式的注意点

- `preview` 中出现冲突时，不要直接 `execute`。
- `force` 不是无限制删除开关。
- 对共享菜单、共享权限和手写代码，默认应该保守处理。
- 删除后应检查 `policy.csv` 或 DB policy 记录是否仍合法。

---

# 9. HTTP 与 CLI 的建议工作流

## 9.1 DSL 工作流

推荐流程：

1. 准备 DSL 文件
2. 先执行预览
3. 确认生成项和冲突项
4. 再执行生成
5. 如果是远程环境，使用下载接口获取产物

示例：

```bash
make server-run-cli ARGS="generate dsl docs/examples/demo.yaml --dry-run"
make server-run-cli ARGS="generate dsl docs/examples/demo.yaml"
```

## 9.2 DB 工作流

推荐流程：

1. 确认数据库连接可用
2. 指定表名范围
3. 先执行预览
4. 检查字段映射、页面建议和 policy 计划
5. 再执行生成

示例：

```bash
make server-run-cli ARGS="generate db preview --driver mysql --dsn '...' --database goadmin --table books --table authors"
make server-run-cli ARGS="generate db generate --driver mysql --dsn '...' --database goadmin --table books --table authors"
```

## 9.3 删除工作流

推荐流程：

1. 执行 `remove preview`
2. 核对删除范围
3. 检查是否有共享资源或人工改动
4. 再执行 `remove execute`
5. 删除后检查菜单、权限和策略是否正常

示例：

```bash
make server-run-cli ARGS="remove preview book --kind crud --policy-store csv"
make server-run-cli ARGS="remove execute book --kind crud --policy-store csv"
```

---

# 10. 常见问题

## 10.1 DSL 和 DB 该怎么选

如果你希望：

- 结果可版本化、可 review、可重复执行

优先选 DSL。

如果你希望：

- 从已有表结构快速补代码
- 不想手工写太多字段定义

优先选 DB。

## 10.2 生成后文件会不会覆盖手写代码

默认不应该。

CodeGen 的设计原则是增量生成、保留手写内容、尽量只更新受控区块。即使使用 `--force`，也应谨慎看待受控文件和共享文件。

## 10.3 为什么删除前一定要 preview

因为删除模式会同时涉及：

- 源码文件
- 菜单与权限
- policy 记录
- registry 注册项

先 preview 能最大限度降低误删风险。

## 10.4 HTTP DB 入口为什么没有 `dsn`

因为 HTTP 模式下，CodeGen 服务端已经持有数据库连接，前端只需要告诉它：

- 生成哪个数据库
- 生成哪些表
- 用什么 schema
- 是否生成前端和 policy

## 10.5 `generate-download` 是做什么的

它适合远程或容器化环境：

- 服务端先生成产物
- 再打包为下载包
- 用户本地下载后再合并到工程中

---

# 11. 推荐命令速查

## 11.1 模块生成

```bash
make server-run-cli ARGS="generate module user"
```

## 11.2 CRUD 生成

```bash
make server-run-cli ARGS="generate crud order --fields id:string,name:string,status:string --policy --frontend"
```

## 11.3 DSL 生成

```bash
make server-run-cli ARGS="generate dsl path/to/codegen.yaml --dry-run"
make server-run-cli ARGS="generate dsl path/to/codegen.yaml"
```

## 11.4 DB 预览与生成

```bash
make server-run-cli ARGS="generate db preview --driver mysql --dsn '...' --database goadmin --table books"
make server-run-cli ARGS="generate db generate --driver mysql --dsn '...' --database goadmin --table books"
```

## 11.5 删除预览与执行

```bash
make server-run-cli ARGS="remove preview book --kind crud --with-policy --with-runtime --with-frontend --with-registry --policy-store db"
make server-run-cli ARGS="remove execute book --kind crud --with-policy --with-runtime --with-frontend --with-registry --policy-store db"
```

---

# 12. 小结

GoAdmin 的 CodeGen 可以按下面的思路来使用：

- **DSL**：适合设计态、版本化和可审阅输入
- **DB**：适合从现有数据库结构反查生成
- **remove**：适合安全回收生成资产

如果你只记住一句话，可以记住：

- **先 preview，再 generate 或 execute**
- **DSL 适合规范化输入，DB 适合反查生成**
- **删除模式只清理生成器负责的资产**

这三点就是当前 CodeGen 模块的最重要使用原则。
