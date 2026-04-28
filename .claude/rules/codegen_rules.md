# GoAdmin CodeGen Rules

## 1. Source of Truth

- 优先遵循 `docs/teckdesign/3. GoAdmin CodeGen 架构设计.md`。
- 目录拆分、输入模型、生成流程、增量保护、框架适配与实施计划必须以该文档为准。
- 若 CodeGen 设计与现有实现冲突，不要臆测；先确认当前仓库真实行为，再决定是修正文档、调整实现，还是更新规则。
- 严格遵守 clean-room 原则：**不要复制 Gin-Vue-Admin 的生成器、模板、目录、接口、字段命名或页面实现**。

## 2. CodeGen Architecture Rules

- CodeGen 的核心职责是 **schema → model(IR) → planner → generator → merger → postprocess**，不要把所有逻辑塞进单一生成器类或单个模板目录。
- 任何输入源（DSL、DB、CLI、插件清单）都必须先统一到 IR，再进入生成计划与输出阶段。
- 生成器必须支持 **增量生成**，默认不覆盖人工手写代码。
- 生成器必须支持 **dry run / preview**，能够先展示差异，再落盘。
- 生成器必须支持 **插件化扩展**，后端框架、前端框架、模板包、字段类型映射都应可扩展。
- CodeGen 不是运行时业务层，不能承担权限校验、数据库事务编排、事件投递或插件运行管理。

## 3. Repository Layout Rules

- 当前兼容层仍位于 `server/cli/generate/`，但新架构能力应优先沉淀到 `server/codegen/`。
- 建议优先按以下边界组织：
  - `server/codegen/driver/`
  - `server/codegen/schema/`
  - `server/codegen/model/`
  - `server/codegen/planner/`
  - `server/codegen/generator/`
  - `server/codegen/merger/`
  - `server/codegen/templates/`
  - `server/codegen/postprocess/`
  - `server/codegen/runtime/`
- `server/cli/generate/` 只适合保留为兼容入口和迁移过渡层，不要继续堆叠新职责。
- 模板资源应与生成逻辑分离，避免把模板字符串散落在业务代码中。
- 不要把 CodeGen 的输入 DSL、输出 artifact、运行快照混在同一目录里。

## 4. Schema and Model Rules

- `schema/` 只负责解析与校验，不负责写文件、生成代码或做框架适配。
- `model/` 只负责统一 IR、资源模型与差异模型，不直接依赖 CLI、HTTP、ORM 或具体前端框架。
- DSL 与 DB 是平级输入源，最终必须收敛到统一 IR；不要让模板直接读取原始 DSL 或数据库表结构。
- 生成模型中必须显式包含：模块、实体、字段、关系、路由、权限、页面、插件、manifest、前后端绑定信息。
- 所有命名转换、字段类型映射、权限资源命名必须由统一规则控制，禁止在模板里硬编码。

## 5. Planner and Merger Rules

- `planner/` 只负责“生成什么、跳过什么、更新什么”，不负责输出代码内容。
- `planner/` 必须能输出生成计划、冲突项、差异项、跳过项和 dry run 结果。
- `merger/` 只负责增量合并与手写保护，不负责框架判断和模板渲染。
- 默认覆盖策略应是：**新文件可创建，受控区块可更新，手写区块必须保留**。
- 必须支持块级标记、局部替换、文件保护、冲突提示；不要用“整文件重写”作为默认行为。
- 若需要破坏性覆盖，必须显式加 `--force` 或等价确认开关，并在文档与命令中保持一致。

## 6. Generator Rules

- `generator/` 按目标拆分：后端、前端、插件、策略、配置等，避免一个生成器同时负责所有平台。
- 后端适配至少要为 `gin`、`go-zero`、`kratos` 预留独立适配层。
- 前端适配至少要为 `Vue3` 预留稳定输出层，`React` 通过同一生成抽象可扩展接入。
- 生成输出必须保持与平台架构兼容：后端遵循模块化分层，前端遵循路由、API、页面、状态分离。
- 插件相关生成必须同时考虑运行时插件系统与 CodeGen 插件系统的边界，不要混淆二者职责。
- 生成策略应尽量可测试、可复现、可重放。

## 7. Templates Rules

- 模板必须版本化，并记录来源与适配框架。
- 模板应尽量参数化，避免模板中出现不可维护的复杂逻辑。
- 模板输出后必须经过格式化与后处理。
- 模板资源不得包含依赖本地路径的硬编码，必须使用生成上下文提供的相对路径或统一映射。
- 任何会影响可读性或可维护性的模板变更，都必须同步补充测试样例或生成结果样例。

## 8. CLI / API / Job Rules

- `server/cmd/cli` 作为当前兼容入口可以保留，但新增能力应优先接入新的 CodeGen 核心。
- CodeGen 的外部入口应支持：CLI、HTTP API、异步任务三种形态中的至少一种；如扩展新入口，不得破坏现有命令。
- CLI 参数必须与文档中的生成行为一致，不允许偷偷改变默认覆盖策略或默认输出目录。
- 对于耗时生成任务，优先支持 dry run、预览与异步执行，不要在主进程里做不可控的大规模同步生成。

## 9. Testing and Verification Rules

- 任何 CodeGen 改动都必须至少补充：
  - 模型/解析测试
  - 模板渲染测试
  - 生成结果快照测试
  - 增量合并或覆盖保护测试
- 修改 CLI 或生成逻辑后，优先运行对应包测试，再考虑全量测试。
- 如果改动影响了现有生成器输出，必须验证：
  - module scaffold
  - CRUD scaffold
  - plugin scaffold
  - Casbin policy 追加
  - 前端生成结果
- 如果改动影响目录结构、输入 DSL、模板路径或输出路径，必须同步回看 `docs/teckdesign/3. GoAdmin CodeGen 架构设计.md` 和目录草案文档。

## 10. Do / Don't

### Do

- 保持 IR 先行。
- 保持生成计划可预览。
- 保持模板与逻辑分离。
- 保持增量更新和手写保护。
- 保持多框架适配器清晰可扩展。
- 保持文档、目录、实现一致。

### Don't

- 不要把生成器写成单个巨型模板函数。
- 不要从模板直接读取原始 DSL 或数据库结构。
- 不要默认覆盖人工修改。
- 不要把运行时业务逻辑塞进 CodeGen。
- 不要复制旧项目的生成规则、目录或页面实现。
- 不要在没有版本与测试的情况下替换模板。
