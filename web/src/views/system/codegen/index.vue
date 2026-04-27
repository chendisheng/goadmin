<template>
  <div class="codegen-page">
    <el-row :gutter="20" class="codegen-grid">
      <el-col :xs="24" :lg="11" :xl="11">
        <el-card shadow="never" class="codegen-card">
          <template #header>
            <div class="card-header">
              <div>
                <div class="title">{{ t('codegen.console_title', 'CodeGen Console') }}</div>
                <div class="subtitle">{{ t('codegen.console_subtitle', '在同一页面中切换 DSL、DB 与删除模式，复用统一结果区。') }}</div>
              </div>
              <el-space wrap>
                <el-button v-if="activeMode === 'dsl'" @click="loadSample">{{ t('codegen.load_sample', '载入示例') }}</el-button>
                <el-button v-else-if="activeMode === 'db'" @click="loadDbSample">{{ t('codegen.load_sample', '载入示例') }}</el-button>
                <el-button v-else @click="loadDeleteSample">{{ t('codegen.load_sample', '载入示例') }}</el-button>
                <el-button @click="clearCurrentInputs">{{ t('codegen.clear', '清空') }}</el-button>
                <el-button v-if="activeMode === 'dsl'" @click="triggerFileSelect">{{ t('codegen.upload_dsl', '上传 DSL') }}</el-button>
                <el-button v-if="activeMode === 'dsl'" type="primary" :loading="previewLoading" @click="handlePreview">{{ t('codegen.preview_dry_run', 'Dry-run 预览') }}</el-button>
                <el-button v-if="activeMode === 'dsl'" type="success" :loading="generateLoading" @click="handleGenerate">{{ t('codegen.generate_once', '一键生成') }}</el-button>
                <el-button v-if="activeMode === 'dsl'" type="warning" :loading="downloadLoading" @click="handleGenerateDownload">{{ t('codegen.generate_download', '生成并下载') }}</el-button>
                <el-button v-if="activeMode === 'db'" type="primary" :loading="previewLoading" @click="handlePreview">{{ t('codegen.preview_dry_run', 'Dry-run 预览') }}</el-button>
                <el-button v-if="activeMode === 'db'" type="success" :loading="generateLoading || installLoading" @click="handleGenerateAndInstall">{{ t('codegen.generate_install', '生成并安装') }}</el-button>
                <el-button v-if="activeMode === 'db'" type="warning" :loading="downloadLoading" @click="handleGenerateDownload">{{ t('codegen.generate_download', '生成并下载') }}</el-button>
                <el-button v-if="activeMode === 'delete'" @click="loadDeleteSample">{{ t('codegen.load_sample', '载入示例') }}</el-button>
                <el-button v-if="activeMode === 'delete'" type="primary" :loading="previewLoading" @click="handleDeletePreview">{{ t('codegen.delete_preview', '删除预览') }}</el-button>
                <el-button v-if="activeMode === 'delete'" type="danger" :loading="deleteLoading" :disabled="!deleteExecuteEnabled" @click="handleDeleteExecute">{{ t('codegen.confirm_delete', '确认删除') }}</el-button>
              </el-space>
            </div>
          </template>

          <el-tabs v-model="activeMode" class="codegen-tabs" stretch>
            <el-tab-pane :label="t('codegen.mode.dsl', 'DSL')" name="dsl">
              <el-form label-position="top" class="codegen-form">
                <el-form-item :label="t('codegen.force_overwrite', '强制覆盖')">
                  <el-switch v-model="force" inline-prompt :active-text="t('common.on', 'On')" :inactive-text="t('common.off', 'Off')" />
                </el-form-item>
                <el-form-item :label="t('codegen.package_name', '下载包名称')">
                  <el-input v-model="packageName" :placeholder="t('codegen.package_name_placeholder', '留空则由系统自动生成 zip 名称')" />
                </el-form-item>
                <el-form-item :label="t('codegen.package_content', '下载包内容')">
                  <el-space wrap>
                    <el-switch v-model="includeReadme" inline-prompt active-text="README" inactive-text="README" />
                    <el-switch v-model="includeReport" inline-prompt active-text="Report" inactive-text="Report" />
                    <el-switch v-model="includeDsl" inline-prompt active-text="DSL" inactive-text="DSL" />
                  </el-space>
                </el-form-item>
                <el-form-item :label="t('codegen.dsl_content', 'DSL 内容')">
                  <el-input
                    v-model="dslText"
                    type="textarea"
                    :rows="28"
                    resize="none"
                    :placeholder="t('codegen.dsl_placeholder', '在这里粘贴或编辑 DSL YAML')"
                  />
                </el-form-item>
              </el-form>
              <input ref="fileInputRef" class="hidden-file-input" type="file" accept=".yaml,.yml,.json,.txt" @change="handleFileChange" />
            </el-tab-pane>

            <el-tab-pane :label="t('codegen.mode.db', 'DB')" name="db">
              <div class="db-mode-panel">
                <div class="db-hero">
                  <div>
                    <div class="db-hero-title">{{ t('codegen.db_guide_title', '数据库输入向导') }}</div>
                    <div class="db-hero-subtitle">{{ t('codegen.db_guide_subtitle', '先选驱动，再填连接串与扫描范围。建议先预览，确认结果后再生成。') }}</div>
                  </div>
                  <el-space wrap>
                    <el-tag type="info" effect="light">{{ dbDriverLabel }}</el-tag>
                    <el-tag type="success" effect="light">{{ dbParsedTables.length ? `${dbParsedTables.length} 个表` : t('common.all', '全部表') }}</el-tag>
                  </el-space>
                </div>

                <el-alert
                  :title="t('codegen.db_recommended_preview', '推荐先执行 Dry-run 预览，确认文件计划和冲突后再执行生成。')"
                  type="info"
                  :closable="false"
                  show-icon
                />

                <div class="db-preset-row">
                  <div>
                    <div class="db-section-title">{{ t('codegen.db_fast_template', '快速模板') }}</div>
                    <div class="db-section-hint">{{ t('codegen.db_fast_template_hint', '一键预填常见数据库的连接格式。') }}</div>
                  </div>
                  <el-space wrap>
                    <el-button size="small" :type="dbDriver === 'mysql' ? 'primary' : 'default'" plain @click="applyDbPreset('mysql')">MySQL</el-button>
                    <el-button size="small" :type="dbDriver === 'postgres' ? 'primary' : 'default'" plain @click="applyDbPreset('postgres')">PostgreSQL</el-button>
                    <el-button size="small" :type="dbDriver === 'sqlite' ? 'primary' : 'default'" plain @click="applyDbPreset('sqlite')">SQLite</el-button>
                  </el-space>
                </div>

                <el-form label-position="top" class="codegen-form db-form">
                  <el-row :gutter="16">
                    <el-col :xs="24" :md="6">
                      <el-form-item :label="t('codegen.db_driver', '数据库驱动')">
                        <el-select v-model="dbDriver" :placeholder="t('codegen.db_driver_placeholder', '请选择数据库驱动')" filterable>
                          <el-option label="MySQL" value="mysql" />
                          <el-option label="PostgreSQL" value="postgres" />
                          <el-option label="SQLite" value="sqlite" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="6">
                      <el-form-item :label="t('codegen.db_name', '数据库名')">
                        <el-input v-model="dbDatabase" :placeholder="t('codegen.db_name_placeholder', '请输入数据库名称')" />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="6">
                      <el-form-item :label="t('codegen.db_schema', 'Schema')">
                        <el-input v-model="dbSchema" :placeholder="t('codegen.db_schema_placeholder', '可选，PostgreSQL 等场景使用')" />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="6">
                      <el-form-item :label="t('codegen.db_mount_root', '挂载根菜单')">
                        <el-select v-model="dbMountParentPath" clearable filterable :placeholder="t('codegen.db_mount_root_placeholder', '留空为顶层根菜单')">
                          <el-option :label="t('codegen.db_mount_root_top', '顶层根菜单')" value="" />
                          <el-option
                            v-for="option in dbMountMenuOptions"
                            :key="option.value"
                            :label="option.label"
                            :value="option.value"
                          />
                        </el-select>
                      </el-form-item>
                    </el-col>
                  </el-row>

                    <el-form-item :label="t('codegen.db_table_range', '表名范围')" class="db-form-item db-form-item--wide">
                      <el-input
                        v-model="dbTablesText"
                        type="textarea"
                        :rows="6"
                        resize="none"
                        :placeholder="t('codegen.db_table_range_placeholder', '支持逗号、换行分隔，例如：books, orders')"
                      />
                      <div class="db-form-row">
                        <span class="db-field-help">{{ t('codegen.db_table_range_help', '留空则表示扫描全部表；建议优先预填少量表进行预览。') }}</span>
                        <el-space wrap>
                          <el-button text size="small" @click="loadDbSample">{{ t('codegen.load_sample', '载入示例') }}</el-button>
                          <el-button text size="small" @click="clearDbTables">{{ t('codegen.clear', '清空') }}</el-button>
                        </el-space>
                      </div>
                      <div v-if="dbParsedTables.length" class="db-table-chip-list">
                        <el-tag
                          v-for="table in dbParsedTables"
                          :key="table"
                          size="small"
                          effect="plain"
                          class="db-table-chip"
                        >
                          {{ table }}
                        </el-tag>
                      </div>
                    </el-form-item>

                  <div class="db-advanced">
                    <div class="db-section-header">
                      <div>
                        <div class="db-section-title">{{ t('codegen.generate_options', '生成选项') }}</div>
                        <div class="db-section-hint">{{ t('codegen.generate_options_hint', '控制是否覆盖现有文件、是否输出前端和权限策略。') }}</div>
                      </div>
                      <el-tag size="small" type="success" effect="light">{{ dbOptionSummary }}</el-tag>
                    </div>
                    <el-space wrap>
                      <el-switch v-model="dbForce" inline-prompt active-text="Force" inactive-text="Force" />
                      <el-switch v-model="dbGenerateFrontend" inline-prompt active-text="Frontend" inactive-text="Frontend" />
                      <el-switch v-model="dbGeneratePolicy" inline-prompt active-text="Policy" inactive-text="Policy" />
                    </el-space>
                  </div>
                </el-form>
              </div>
            </el-tab-pane>

            <el-tab-pane :label="t('codegen.mode.delete', 'Delete')" name="delete">
              <div class="delete-mode-panel">
                <el-alert
                  :title="t('codegen.delete_preview_required', '请先执行删除预览，确认计划、风险与冲突后，再点击确认删除。')"
                  type="warning"
                  :closable="false"
                  show-icon
                />

                <el-form label-position="top" class="codegen-form db-form delete-form">
                  <el-row :gutter="16">
                    <el-col :xs="24" :md="8">
                      <el-form-item :label="t('codegen.delete_module', '模块名')">
                        <el-input v-model="deleteModule" placeholder="例如 book" />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="8">
                      <el-form-item :label="t('codegen.delete_kind', '模块类型')">
                        <el-input v-model="deleteKind" placeholder="例如 crud" />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="8">
                      <el-form-item :label="t('codegen.delete_policy_store', 'Policy Store')">
                        <el-select v-model="deletePolicyStore" clearable filterable :placeholder="t('codegen.delete_policy_store_placeholder', '自动识别或手动指定')">
                          <el-option :label="t('codegen.delete_policy_store_auto_detect', '自动识别')" value="" />
                          <el-option label="CSV" value="csv" />
                          <el-option label="DB" value="db" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-form-item :label="t('codegen.delete_scope', '删除范围')">
                    <el-space wrap>
                      <el-switch v-model="deleteWithRuntime" inline-prompt active-text="Runtime" inactive-text="Runtime" />
                      <el-switch v-model="deleteWithPolicy" inline-prompt active-text="Policy" inactive-text="Policy" />
                      <el-switch v-model="deleteWithFrontend" inline-prompt active-text="Frontend" inactive-text="Frontend" />
                      <el-switch v-model="deleteWithRegistry" inline-prompt active-text="Registry" inactive-text="Registry" />
                      <el-switch v-model="deleteForce" inline-prompt active-text="Force" inactive-text="Force" />
                    </el-space>
                  </el-form-item>

                  <el-form-item :label="t('codegen.execute_notes', '执行说明')">
                    <el-input
                      v-model="deleteNotes"
                      type="textarea"
                      :rows="5"
                      resize="none"
                      :placeholder="t('codegen.execute_notes_placeholder', '可选：补充删除说明，仅用于界面记录，不会直接传给后端核心')"
                    />
                  </el-form-item>
                </el-form>
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="13" :xl="13">
        <el-space direction="vertical" :size="16" fill class="side-stack">
          <el-card shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.result_title', '执行结果') }}</div>
                  <div class="subtitle">{{ t('codegen.result_subtitle', '预览和生成都会回传资源级动作。') }}</div>
                </div>
              </div>
            </template>

            <el-alert
              v-if="statusMessage"
              :title="statusMessage"
              :type="lastRunSuccess ? 'success' : 'info'"
              :closable="false"
              show-icon
            />
            <div v-else class="result-empty">{{ t('codegen.no_result', '尚未执行预览或生成。') }}</div>

            <div class="preview-table-wrap">
              <el-table :data="previewItems" class="preview-table" size="small" border>
                <el-table-column prop="index" :label="t('codegen.preview.index', '#')" width="60" />
                <el-table-column prop="kind" :label="t('codegen.preview.kind', 'Kind')" min-width="130" show-overflow-tooltip />
                <el-table-column prop="name" :label="t('codegen.preview.name', 'Name')" min-width="160" show-overflow-tooltip />
                <el-table-column :label="t('codegen.preview.force', 'Force')" width="88">
                  <template #default="scope">
                    <el-tag :type="scope.row.force ? 'warning' : 'info'" effect="light">
                      {{ scope.row.force ? t('common.yes', 'Yes') : t('common.no', 'No') }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column :label="t('codegen.preview.actions', 'Actions')" min-width="280">
                  <template #default="scope">
                    <div class="action-tags">
                      <el-tag
                        v-for="action in scope.row.actions"
                        :key="action"
                        size="small"
                        effect="plain"
                        class="action-tag"
                      >
                        {{ action }}
                      </el-tag>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-card>

          <el-card v-if="activeMode === 'db'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.file_plan_title', '文件计划') }}</div>
                  <div class="subtitle">{{ t('codegen.file_plan_subtitle', '显示数据库预览阶段推导出的文件清单。') }}</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!filePlans.length" :description="t('codegen.no_file_plan', '暂无文件计划')" />
            <el-table v-else :data="filePlans" size="small" border class="preview-table">
              <el-table-column prop="path" :label="t('codegen.file_plan.path', 'Path')" min-width="220" show-overflow-tooltip />
              <el-table-column prop="action" :label="t('codegen.file_plan.action', 'Action')" width="120" />
              <el-table-column prop="kind" :label="t('codegen.file_plan.kind', 'Kind')" width="140" show-overflow-tooltip />
              <el-table-column :label="t('codegen.file_plan.exists', 'Exists')" width="88">
                <template #default="scope">
                  <el-tag :type="scope.row.exists ? 'warning' : 'success'" effect="light">
                    {{ scope.row.exists ? t('common.yes', 'Yes') : t('common.no', 'No') }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('codegen.file_plan.conflict', 'Conflict')" width="96">
                <template #default="scope">
                  <el-tag :type="scope.row.conflict ? 'danger' : 'success'" effect="light">
                    {{ scope.row.conflict ? t('common.yes', 'Yes') : t('common.no', 'No') }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card v-if="activeMode === 'delete'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.risk_title', '风险与冲突') }}</div>
                  <div class="subtitle">{{ t('codegen.risk_subtitle', '展示删除预览中的风险提示和阻断冲突。') }}</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!deleteConflicts.length" :description="t('codegen.no_risk', 'No conflicts')" />
            <el-table v-else :data="deleteConflicts" size="small" border class="preview-table">
              <el-table-column prop="kind" :label="t('codegen.conflict.kind', 'Kind')" min-width="140" show-overflow-tooltip />
              <el-table-column prop="severity" :label="t('codegen.conflict.severity', 'Severity')" width="110" />
              <el-table-column prop="path" :label="t('codegen.conflict.path', 'Path')" min-width="200" show-overflow-tooltip />
              <el-table-column prop="message" :label="t('codegen.conflict.message', 'Message')" min-width="280" show-overflow-tooltip />
            </el-table>
          </el-card>

          <el-card v-if="activeMode === 'db'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.conflict_title', '冲突') }}</div>
                  <div class="subtitle">{{ t('codegen.conflict_subtitle', '展示文件覆盖风险与路径冲突信息。') }}</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!conflicts.length" :description="t('codegen.no_conflicts', 'No conflicts')" />
            <el-table v-else :data="conflicts" size="small" border class="preview-table">
              <el-table-column prop="path" :label="t('codegen.conflict.path', 'Path')" min-width="220" show-overflow-tooltip />
              <el-table-column prop="resource" :label="t('codegen.conflict.resource', 'Resource')" min-width="160" show-overflow-tooltip />
              <el-table-column prop="reason" :label="t('codegen.conflict.reason', 'Reason')" min-width="260" show-overflow-tooltip />
            </el-table>
          </el-card>

          <el-card v-if="activeMode === 'db'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.audit_title', '审计') }}</div>
                  <div class="subtitle">{{ t('codegen.audit_subtitle', '记录输入、执行步骤与输出统计。') }}</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!auditRecord" :description="t('codegen.no_audit', '暂无审计记录')" />
            <div v-else class="artifact-panel">
              <div class="artifact-summary-grid">
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.record_time', '记录时间') }}</span>
                  <span class="artifact-summary-value">{{ formatDateTime(auditRecord.recorded_at) }}</span>
                  <span class="artifact-summary-meta">{{ auditRecord.recorded_at }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.input', '输入') }}</span>
                  <span class="artifact-summary-value">{{ auditRecord.input.driver }} / {{ auditRecord.input.database }}</span>
                  <span class="artifact-summary-meta">dry-run: {{ auditRecord.input.dry_run ? t('common.yes', 'Yes') : t('common.no', 'No') }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.output_file_count', '输出文件') }}</span>
                  <span class="artifact-summary-value">{{ auditRecord.output.file_count }}</span>
                  <span class="artifact-summary-meta">{{ t('codegen.conflict_count_label', 'Conflicts: {count}', { count: auditRecord.output.conflict_count }) }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.table_scope', '表范围') }}</span>
                  <span class="artifact-summary-value">{{ auditRecord.input.tables?.length ?? 0 }}</span>
                  <span class="artifact-summary-meta">{{ (auditRecord.input.tables ?? []).join(', ') || '全部表' }}</span>
                </div>
              </div>
              <div class="artifact-grid artifact-grid--compact">
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.execution_steps', '执行步骤') }}</span>
                  <ul class="message-list compact-list">
                    <li v-for="step in auditRecord.steps" :key="step.name">
                      {{ step.name }} [{{ step.status }}]{{ step.detail ? ` - ${step.detail}` : '' }}
                    </li>
                  </ul>
                </div>
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.output_overview', '输出概览') }}</span>
                  <span class="artifact-value">文件 {{ auditRecord.output.files.length }} · 冲突 {{ auditRecord.output.conflicts.length }}</span>
                </div>
              </div>
            </div>
          </el-card>

          <el-card v-if="activeMode === 'delete' && deleteResult" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ t('codegen.delete_result_title', '删除结果') }}</div>
                  <div class="subtitle">{{ t('codegen.delete_result_subtitle', '展示本次删除的执行概览、异常情况与处理明细。') }}</div>
                </div>
              </div>
            </template>

            <div class="artifact-panel">
              <el-alert
                :title="deleteResultStatusMessage"
                :type="deleteResultStatusType"
                :closable="false"
                show-icon
              />
              <div class="artifact-summary-grid">
                <div class="artifact-summary-card" :class="`artifact-summary-card--${deleteResultStatusTone}`">
                  <span class="artifact-summary-label">{{ t('codegen.result_status', '结果状态') }}</span>
                  <span class="artifact-summary-value">{{ deleteResultStatusLabel }}</span>
                  <span class="artifact-summary-meta">{{ deleteResultSummaryText }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.processed', '已处理') }}</span>
                  <span class="artifact-summary-value">{{ deleteResultSummary?.total_deleted ?? 0 }}</span>
                  <span class="artifact-summary-meta">源文件 {{ deleteResultSummary?.deleted_source_files ?? 0 }} · 运行时 {{ deleteResultSummary?.deleted_runtime_assets ?? 0 }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.skipped_failed', '跳过 / 异常') }}</span>
                  <span class="artifact-summary-value">{{ deleteResultSummary?.skipped ?? 0 }} / {{ deleteResultSummary?.failed ?? 0 }}</span>
                  <span class="artifact-summary-meta">权限 {{ deleteResultSummary?.deleted_policy_changes ?? 0 }} · 前端 {{ deleteResultSummary?.deleted_frontend_changes ?? 0 }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.elapsed', '执行耗时') }}</span>
                  <span class="artifact-summary-value">{{ deleteResultElapsedText }}</span>
                  <span class="artifact-summary-meta">开始于 {{ formatDateTime(deleteResult?.started_at ?? '') }} · 结束于 {{ formatDateTime(deleteResult?.finished_at ?? '') }}</span>
                </div>
              </div>
              <div class="artifact-grid artifact-grid--compact">
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.delete_detail', '删除明细') }}</span>
                  <ul class="message-list compact-list">
                    <li v-for="item in deleteResultDeleted" :key="deleteItemKey(item)">
                      {{ describeDeleteItem(item) }}
                    </li>
                  </ul>
                </div>
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.skip_detail', '跳过明细') }}</span>
                  <ul class="message-list compact-list">
                    <li v-for="item in deleteResultSkipped" :key="deleteItemKey(item)">
                      {{ describeDeleteItem(item) }}
                    </li>
                  </ul>
                </div>
                <div class="artifact-item artifact-item--wide" v-if="deleteResultFailures.length">
                  <span class="artifact-label">{{ t('codegen.failure_detail', '异常明细') }}</span>
                  <ul class="message-list compact-list">
                    <li v-for="failure in deleteResultFailures" :key="describeDeleteFailureKey(failure)">
                      {{ describeDeleteFailure(failure) }}
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </el-card>

          <el-card shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">{{ messagePanelTitle }}</div>
                  <div class="subtitle">{{ messagePanelSubtitle }}</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!messages.length" :description="messagePanelEmptyText" />
            <ul v-else class="message-list">
              <li v-for="message in messages" :key="message">{{ message }}</li>
            </ul>
          </el-card>

          <el-card v-if="artifactInfo" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact artifact-header">
                <div>
                  <div class="title">{{ t('codegen.artifact_title', '下载产物') }}</div>
                  <div class="subtitle">{{ t('codegen.artifact_subtitle', '展示最近一次服务端打包结果，并支持重新下载。') }}</div>
                </div>
                <el-space v-if="artifactInfo" wrap>
                  <el-button text :class="{ 'artifact-copy-button--active': copyFeedbackActive }" :disabled="!canCopyArtifactUrl" @click="handleCopyDownloadUrl">{{ copyButtonText }}</el-button>
                  <el-button text type="primary" :disabled="isArtifactExpired" :loading="downloadLoading" @click="handleArtifactDownload">{{ t('common.refresh', '重新下载') }}</el-button>
                </el-space>
              </div>
            </template>

            <el-empty v-if="!artifactInfo" :description="t('codegen.no_artifact', '尚未生成下载包')" />
            <div v-else class="artifact-panel">
              <el-alert
                :title="artifactStatusMessage"
                :type="artifactStatusType"
                :closable="false"
                show-icon
              />
              <div class="artifact-summary-grid">
                <div class="artifact-summary-card" :class="`artifact-summary-card--${artifactStatusTone}`">
                  <span class="artifact-summary-label">{{ t('codegen.artifact_status', '产物状态') }}</span>
                  <span class="artifact-summary-value">{{ artifactStatusLabel }}</span>
                  <span class="artifact-summary-meta">{{ artifactStatusSummary }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.file_overview', '文件概览') }}</span>
                  <span class="artifact-summary-value">{{ artifactInfo.filename }}</span>
                  <span class="artifact-summary-meta">{{ artifactSizeText }} · {{ artifactInfo.file_count }} 个文件</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.valid_until', '有效期') }}</span>
                  <span class="artifact-summary-value">{{ artifactRemainingText }}</span>
                  <span class="artifact-summary-meta">到期于 {{ artifactExpireText }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">{{ t('codegen.recent_activity', '最近活动') }}</span>
                  <span class="artifact-summary-value">{{ artifactLastDownloadText }}</span>
                  <span class="artifact-summary-meta">
                    <template v-if="lastDownloadDuration > 0">耗时 {{ artifactLastDownloadDurationText }} · </template>
                    失败时间：{{ artifactLastFailureText }}
                    <el-tag v-if="lastArtifactError" :type="artifactLastErrorTypeTag" size="small" effect="light" style="margin-left: 6px;">{{ artifactLastErrorTypeText }}</el-tag>
                  </span>
                </div>
              </div>
              <div class="artifact-grid artifact-grid--compact">
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.task_id', '任务 ID') }}</span>
                  <span class="artifact-value monospace">{{ artifactInfo.task_id }}</span>
                </div>
                <div class="artifact-item">
                  <span class="artifact-label">{{ t('codegen.filename', '文件名') }}</span>
                  <span class="artifact-value">{{ artifactInfo.filename }}</span>
                </div>
                <div class="artifact-item">
                  <span class="artifact-label">{{ t('codegen.last_failure_reason', '最近一次失败原因') }}</span>
                  <span class="artifact-value">{{ artifactLastErrorText }}</span>
                </div>
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">{{ t('codegen.download_url', '下载地址') }}</span>
                  <template v-if="canCopyArtifactUrl">
                    <el-button plain size="small" :class="{ 'artifact-copy-button--active': copyFeedbackActive }" @click="handleCopyDownloadUrl">{{ copyButtonText }}</el-button>
                    <span class="artifact-hint">{{ artifactDownloadUrlSummary }}</span>
                  </template>
                  <span v-else class="artifact-hint">{{ t('codegen.expired_hidden', '下载包已过期，旧下载地址已隐藏，请重新生成新的代码包。') }}</span>
                </div>
              </div>
            </div>
          </el-card>
        </el-space>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import {
  downloadCodegenArtifact,
  executeCodegenDelete,
  generateCodegenDatabase,
  generateCodegenDsl,
  generateDownloadCodegenDsl,
  generateDownloadCodegenDatabase,
  installCodegenManifest,
  previewCodegenDatabase,
  previewCodegenDsl,
  previewCodegenDelete,
  type CodegenArtifactInfo,
  type CodegenDatabasePreviewReport,
  type CodegenDatabasePreviewResource,
  type CodegenDatabaseRequest,
  type CodegenDeletePlan,
  type CodegenDeletePlanItem,
  type CodegenDeletePreviewReport,
  type CodegenDeleteRequest,
  type CodegenDeleteResult,
  type CodegenDslExecutionReport,
} from '@/api/codegen';
import { fetchPublicConfig } from '@/api/health';
import { fetchMenuTree } from '@/api/system-menus';
import { ApiError } from '@/api/types';
import { useAppI18n } from '@/i18n';
import type { PublicConfigPayload } from '@/api/health';
import type { MenuItem } from '@/types/admin';

type CodegenMode = 'dsl' | 'db' | 'delete';

type PreviewRow = {
  index: number;
  kind: string;
  name: string;
  force: boolean;
  managed?: boolean;
  actions: string[];
};

type MenuMountOption = {
  label: string;
  value: string;
};

const activeMode = ref<CodegenMode>('dsl');

const dslText = ref('');
const force = ref(false);
const packageName = ref('');
const includeReadme = ref(true);
const includeReport = ref(true);
const includeDsl = ref(true);

const dbDriver = ref('mysql');
const dbDatabase = ref('');
const dbSchema = ref('');
const dbTablesText = ref('');
const dbForce = ref(false);
const dbGenerateFrontend = ref(true);
const dbGeneratePolicy = ref(true);
const dbMountParentPath = ref('');

const deleteModule = ref('');
const deleteKind = ref('crud');
const deletePolicyStore = ref('');
const deleteWithRuntime = ref(true);
const deleteWithPolicy = ref(true);
const deleteWithFrontend = ref(true);
const deleteWithRegistry = ref(true);
const deleteForce = ref(false);
const deleteNotes = ref('');

const previewLoading = ref(false);
const generateLoading = ref(false);
const downloadLoading = ref(false);
const installLoading = ref(false);
const deleteLoading = ref(false);

const dslReport = ref<CodegenDslExecutionReport | null>(null);
const dbReport = ref<CodegenDatabasePreviewReport | null>(null);
const deletePreviewReport = ref<CodegenDeletePreviewReport | null>(null);
const deleteResult = ref<CodegenDeleteResult | null>(null);
const deleteRequestCache = ref<CodegenDeleteRequest | null>(null);
const artifactInfo = ref<CodegenArtifactInfo | null>(null);
const artifactForceExpired = ref(false);
const lastDownloadAt = ref('');
const lastArtifactError = ref('');
const lastArtifactErrorAt = ref('');
const lastArtifactErrorType = ref<'auth' | 'notfound' | 'expired' | 'server' | 'unknown'>('unknown');
const lastDownloadDuration = ref(0);
const downloadStartAt = ref(0);
const copyFeedbackActive = ref(false);
const operationStatus = ref('');
const lastRunSuccess = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);
const currentTime = ref(Date.now());

let artifactTicker: ReturnType<typeof window.setInterval> | null = null;
let copyFeedbackTimer: ReturnType<typeof window.setTimeout> | null = null;
const dbMountMenuOptions = ref<MenuMountOption[]>([]);
const publicConfig = ref<PublicConfigPayload | null>(null);
const { t } = useAppI18n();

const currentReport = computed(() => (activeMode.value === 'db' ? dbReport.value : dslReport.value));
const dbParsedTables = computed(() => parseTableNames(dbTablesText.value));
const dbDriverLabel = computed(() => {
  const map: Record<string, string> = {
    mysql: 'MySQL',
    postgres: 'PostgreSQL',
    sqlite: 'SQLite',
  };
  return map[dbDriver.value] || dbDriver.value || '—';
});
const dbOptionSummary = computed(() => {
  const parts: string[] = [];
  parts.push(dbForce.value ? t('codegen.cover_existing', '覆盖现有文件') : t('codegen.safe_cover_off', '安全覆盖关闭'));
  parts.push(dbGenerateFrontend.value ? t('codegen.generate_frontend', '生成前端') : t('codegen.skip_frontend', '跳过前端'));
  parts.push(dbGeneratePolicy.value ? t('codegen.generate_policy', '生成权限') : t('codegen.skip_policy', '跳过权限'));
  parts.push(dbMountParentPath.value ? `${t('codegen.mount_root_prefix', '挂载：')}${dbMountMenuLabel.value}` : t('codegen.mount_root_top_label', '挂载：顶层根菜单'));
  return parts.join(' · ');
});
const dbGeneratedModuleName = computed(() => dbReport.value?.resources?.[0]?.module?.trim() || '');
const deletePreviewPlan = computed(() => deletePreviewReport.value?.plan ?? null);
const deletePlanSummary = computed(() => deletePreviewPlan.value?.summary ?? null);
const deletePlanItems = computed<PreviewRow[]>(() => mapDeletePlanToPreviewRows(deletePreviewPlan.value));
const deleteConflicts = computed(() => deletePreviewPlan.value?.conflicts ?? []);
const deleteRequestSnapshotLabel = computed(() => {
  if (!deleteRequestCache.value) {
    return t('codegen.no_preview', '未预览');
  }
  return `${deleteRequestCache.value.module || '-'} · ${deleteRequestCache.value.dry_run ? 'dry-run' : 'execute'}`;
});
const deletePolicyStoreLabel = computed(() => deleteRequestCache.value?.policy_store || 'auto');
const deletePlanTotalText = computed(() => String(deletePlanSummary.value?.total ?? deletePlanItems.value.length));
const deletePlanStatusMessage = computed(() => {
  const warnings = deletePreviewPlan.value?.warnings ?? [];
  const conflicts = deleteConflicts.value.length;
  const total = deletePlanItems.value.length;
  if (conflicts > 0) {
    return t('codegen.delete_preview_complete_conflict', '删除预览完成：{total} 项计划中包含 {conflicts} 个冲突，请确认后再执行。', { total, conflicts });
  }
  if (warnings.length > 0) {
    return t('codegen.delete_preview_complete_warning', '删除预览完成：{total} 项计划，存在 {warnings} 条风险提示。', { total, warnings: warnings.length });
  }
  return t('codegen.delete_preview_complete_normal', '删除预览完成：共 {total} 项计划，可继续确认执行。', { total });
});
const deletePlanStatusType = computed(() => {
  if (deleteConflicts.value.length > 0) {
    return 'warning';
  }
  if ((deletePreviewPlan.value?.warnings ?? []).length > 0) {
    return 'info';
  }
  return 'success';
});
const deleteResultSummary = computed(() => deleteResult.value?.summary ?? null);
const deleteResultDeleted = computed(() => deleteResult.value?.deleted ?? []);
const deleteResultSkipped = computed(() => deleteResult.value?.skipped ?? []);
const deleteResultFailures = computed(() => deleteResult.value?.failures ?? []);
const deleteResultElapsedText = computed(() => formatElapsedMillis(deleteResultSummary.value?.elapsed_millis ?? 0));
const deleteResultSummaryText = computed(() => {
  const summary = deleteResultSummary.value;
  if (!summary) {
    return '-';
  }
  return t('codegen.delete_result_summary', '已处理 {deleted} 项 · 跳过 {skipped} 项 · 异常 {failed} 项', {
    deleted: summary.total_deleted ?? 0,
    skipped: summary.skipped ?? 0,
    failed: summary.failed ?? 0,
  });
});
const deleteResultStatusLabel = computed(() => {
  switch (deleteResult.value?.status ?? '') {
    case 'succeeded':
      return t('common.ok', '成功');
    case 'partial':
      return t('codegen.delete_partial', '部分成功');
    case 'failed':
      return t('codegen.delete_failed', '失败');
    case 'dry_run':
      return 'Dry-run';
    case 'planned':
      return t('codegen.delete_planned', '已计划');
    default:
      return t('common.unknown', '未知');
  }
});
const deleteResultStatusType = computed(() => {
  switch (deleteResult.value?.status ?? '') {
    case 'succeeded':
      return 'success';
    case 'partial':
      return 'warning';
    case 'failed':
      return 'error';
    case 'dry_run':
    case 'planned':
      return 'info';
    default:
      return 'info';
  }
});
const deleteResultStatusTone = computed(() => {
  switch (deleteResult.value?.status ?? '') {
    case 'succeeded':
      return 'success';
    case 'partial':
      return 'warning';
    case 'failed':
      return 'danger';
    default:
      return 'warning';
  }
});
const deleteResultStatusMessage = computed(() => {
  const result = deleteResult.value;
  if (!result) {
    return '';
  }
  const summary = result.summary;
  if (result.status === 'succeeded') {
    return t('codegen.delete_result_completed_summary', '删除已完成，共处理 {total} 项。', { total: summary?.total_deleted ?? 0 });
  }
  if (result.status === 'partial') {
    return t('codegen.delete_result_partial_summary', '删除已完成，但有 {skipped} 项被跳过、{failed} 项出现异常。', {
      skipped: summary?.skipped ?? 0,
      failed: summary?.failed ?? 0,
    });
  }
  if (result.status === 'failed') {
    return t('codegen.delete_failed', '删除执行未完成，请先查看异常明细。');
  }
  return t('codegen.delete_result_ended_status', '删除执行已结束，当前状态为 {status}。', { status: deleteResultStatusLabel.value });
});
const deleteMessages = computed(() => {
  const messages: string[] = [];
  if (deletePreviewPlan.value) {
    messages.push(...(deletePreviewPlan.value.warnings ?? []).map((message) => `${t('codegen.delete_preview_hint_prefix', '预览提示：')}${message}`));
    if ((deletePreviewPlan.value.warnings ?? []).length === 0 && (deletePreviewPlan.value.conflicts?.length ?? 0) === 0) {
      messages.push(t('codegen.delete_preview_no_extra_risk', '预览提示：未发现额外风险提示。'));
    }
  }
  if (deleteResult.value) {
    messages.push(...(deleteResult.value.warnings ?? []).map((message) => `${t('codegen.delete_execute_hint_prefix', '执行提示：')}${message}`));
    for (const failure of deleteResult.value.failures ?? []) {
      const detail = failure.reason || t('codegen.delete_failure_reason_unknown', '未提供原因');
      messages.push(`${t('codegen.delete_failure_prefix', '异常：')}${detail}${failure.item ? `（${describeDeleteItem(failure.item)}）` : ''}`);
    }
    if ((deleteResult.value.warnings ?? []).length === 0 && (deleteResult.value.failures ?? []).length === 0) {
      messages.push(t('codegen.delete_execute_no_extra_exception', '执行提示：未发现额外异常。'));
    }
  }
  return messages;
});
const messagePanelTitle = computed(() => (activeMode.value === 'delete' ? t('codegen.delete_messages_title', '删除提示') : t('codegen.messages_title', '消息')));
const messagePanelSubtitle = computed(() =>
  activeMode.value === 'delete'
    ? t('codegen.delete_messages_subtitle', '汇总删除预览提示、执行提示与异常信息。')
    : t('codegen.messages_subtitle', '包含 dry-run 提示、生成摘要和校验信息。'),
);
const messagePanelEmptyText = computed(() => (activeMode.value === 'delete' ? t('codegen.no_delete_messages', '暂无删除提示') : t('codegen.no_messages', '暂无消息')));
const deleteExecuteEnabled = computed(() => activeMode.value === 'delete' && deletePreviewReport.value !== null);
const dbMountMenuLabel = computed(() => {
  if (!dbMountParentPath.value) {
    return t('codegen.db_mount_root_top_label', '顶层根菜单');
  }
  const option = dbMountMenuOptions.value.find((item) => item.value === dbMountParentPath.value);
  return option?.label || dbMountParentPath.value;
});

const previewItems = computed<PreviewRow[]>(() => {
  if (activeMode.value === 'delete') {
    return deletePlanItems.value;
  }
  if (activeMode.value === 'db') {
    return mapDatabaseResourcesToPreviewRows(dbReport.value?.resources ?? [], dbForce.value);
  }
  return (dslReport.value?.items ?? []).map((item) => ({
    index: item.index,
    kind: item.kind,
    name: item.name,
    force: item.force,
    actions: item.actions,
  }));
});

const messages = computed(() => {
  if (activeMode.value === 'delete') {
    return deleteMessages.value;
  }
  const base = currentReport.value?.messages ?? [];
  if (activeMode.value === 'db') {
    return [...base, ...(dbReport.value?.planner.messages ?? [])];
  }
  return base;
});
const filePlans = computed(() => (activeMode.value === 'db' ? dbReport.value?.files ?? [] : []));
const conflicts = computed(() => (activeMode.value === 'db' ? dbReport.value?.conflicts ?? [] : []));
const auditRecord = computed(() => (activeMode.value === 'db' ? dbReport.value?.audit ?? null : null));
const statusMessage = computed(() => {
  if (operationStatus.value) {
    return operationStatus.value;
  }
  const lines = messages.value;
  if (!lines.length) {
    return '';
  }
  return lines.join(' · ');
});
const artifactSizeText = computed(() => formatBytes(artifactInfo.value?.size_bytes ?? 0));
const artifactExpireText = computed(() => formatDateTime(artifactInfo.value?.expire_at ?? ''));
const isArtifactExpired = computed(() => {
  if (artifactForceExpired.value) {
    return true;
  }
  const value = artifactInfo.value?.expire_at ?? '';
  if (!value) {
    return false;
  }
  const expiresAt = new Date(value).getTime();
  if (Number.isNaN(expiresAt)) {
    return false;
  }
  return expiresAt <= currentTime.value;
});
const artifactStatusType = computed(() => {
  if (downloadLoading.value) {
    return 'warning';
  }
  if (isArtifactExpired.value) {
    return 'error';
  }
  return 'success';
});
const artifactStatusMessage = computed(() => {
  if (!artifactInfo.value) {
    return '';
  }
  if (downloadLoading.value) {
    return t('codegen.downloading', '正在准备下载包，请稍候，浏览器即将开始下载。');
  }
  if (isArtifactExpired.value) {
    return t('codegen.artifact_expired', '当前下载包已过期，请重新执行“生成并下载”以获得新的代码包。');
  }
  return t('codegen.artifact_ready', '下载包已就绪，你可以重新下载，或复制下载地址用于当前登录态调试。');
});
const artifactStatusLabel = computed(() => {
  if (downloadLoading.value) {
    return t('codegen.download_preparing', '下载准备中');
  }
  if (isArtifactExpired.value) {
    return t('codegen.artifact_expired_label', '已过期');
  }
  return t('codegen.artifact_downloadable', '可下载');
});
const artifactStatusSummary = computed(() => {
  if (downloadLoading.value) {
    return t('codegen.download_preparing_detail', '浏览器即将开始下载');
  }
  if (isArtifactExpired.value) {
    return t('codegen.artifact_expired_detail', '需要重新生成新的代码包');
  }
  return t('codegen.artifact_ready_detail', '支持重新下载和复制完整地址');
});
const artifactStatusTone = computed(() => {
  if (downloadLoading.value) {
    return 'warning';
  }
  if (isArtifactExpired.value) {
    return 'danger';
  }
  return 'success';
});
const artifactDownloadUrlText = computed(() => {
  if (isArtifactExpired.value) {
    return '';
  }
  return toAbsoluteUrl(artifactInfo.value?.download_url ?? '');
});
const artifactDownloadUrlSummary = computed(() => summarizeDownloadUrl(artifactDownloadUrlText.value));
const canCopyArtifactUrl = computed(() => artifactDownloadUrlText.value !== '');
const artifactRemainingText = computed(() => formatRemainingTime(artifactInfo.value?.expire_at ?? '', currentTime.value, isArtifactExpired.value));
const artifactLastDownloadText = computed(() => formatDateTime(lastDownloadAt.value));
const artifactLastErrorText = computed(() => lastArtifactError.value || '无');
const artifactLastFailureText = computed(() => formatDateTime(lastArtifactErrorAt.value));
const artifactLastErrorTypeText = computed(() => {
  const map: Record<typeof lastArtifactErrorType.value, string> = {
    auth: t('codegen.error_auth', '登录失效'),
    notfound: t('codegen.error_notfound', '资源不存在'),
    expired: t('codegen.error_expired', '已过期'),
    server: t('codegen.error_server', '服务异常'),
    unknown: t('codegen.error_unknown', '其他'),
  };
  return map[lastArtifactErrorType.value] || t('codegen.error_unknown', '其他');
});
const artifactLastErrorTypeTag = computed(() => {
  const map: Record<typeof lastArtifactErrorType.value, 'info' | 'warning' | 'danger' | 'success'> = {
    auth: 'warning',
    notfound: 'info',
    expired: 'danger',
    server: 'danger',
    unknown: 'info',
  };
  return map[lastArtifactErrorType.value] || 'info';
});
const artifactLastDownloadDurationText = computed(() => {
  if (lastDownloadDuration.value <= 0) {
    return '-';
  }
  const ms = lastDownloadDuration.value;
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const seconds = Math.floor(ms / 1000);
  const remainMs = ms % 1000;
  return `${seconds}.${String(remainMs).padStart(3, '0')}s`;
});
const copyButtonText = computed(() => (copyFeedbackActive.value ? '已复制' : '复制下载地址'));

onMounted(() => {
  currentTime.value = Date.now();
  artifactTicker = window.setInterval(() => {
    currentTime.value = Date.now();
  }, 1000);
  void loadPublicConfig();
  void loadDbMountMenuOptions();
});

onBeforeUnmount(() => {
  if (artifactTicker !== null) {
    window.clearInterval(artifactTicker);
    artifactTicker = null;
  }
  if (copyFeedbackTimer !== null) {
    window.clearTimeout(copyFeedbackTimer);
    copyFeedbackTimer = null;
  }
});

function loadSample() {
  activeMode.value = 'dsl';
  dslText.value = `version: v1
module: codegen
framework:
  backend: gin
  frontend: vue3
resources:
  - kind: frontend-page
    name: codegen-console
    module: codegen
    pages:
      - name: console
        path: /system/codegen
        component: system/codegen/index
`;
  ElMessage.success(t('codegen.load_sample_success', '示例已载入'));
}

function loadDbSample() {
  activeMode.value = 'db';
  applyDbPreset('sqlite');
  dbSchema.value = '';
  dbTablesText.value = 'books, orders';
  dbForce.value = false;
  dbGenerateFrontend.value = true;
  dbGeneratePolicy.value = true;
  dbMountParentPath.value = '';
  ElMessage.success(t('codegen.load_db_sample_success', '示例数据库配置已载入'));
}

function loadDeleteSample() {
  activeMode.value = 'delete';
  deleteModule.value = 'book';
  deleteKind.value = 'crud';
  deletePolicyStore.value = 'db';
  deleteWithRuntime.value = true;
  deleteWithPolicy.value = true;
  deleteWithFrontend.value = true;
  deleteWithRegistry.value = true;
  deleteForce.value = false;
  deleteNotes.value = '先预览再确认执行';
  ElMessage.success(t('codegen.load_delete_sample_success', '示例删除配置已载入'));
}

function applyDbPreset(driver: 'mysql' | 'postgres' | 'sqlite') {
  const presets: Record<'mysql' | 'postgres' | 'sqlite', { database: string; schema: string }> = {
    mysql: {
      database: 'goadmin',
      schema: '',
    },
    postgres: {
      database: 'goadmin',
      schema: 'public',
    },
    sqlite: {
      database: 'goadmin',
      schema: '',
    },
  };
  const preset = presets[driver];
  dbDriver.value = driver;
  dbDatabase.value = preset.database;
  dbSchema.value = preset.schema;
}

function applyDbConfigDefaults() {
  const database = publicConfig.value?.database;
  if (!database) {
    return;
  }
  const name = database.name?.trim();
  if (name && !dbDatabase.value.trim()) {
    dbDatabase.value = name;
  }
}

async function handleDeleteExecute() {
  if (!deleteExecuteEnabled.value || !deletePreviewReport.value || !deleteRequestCache.value) {
    ElMessage.warning(t('codegen.delete_preview_required', '请先完成删除预览'));
    return;
  }
  const preview = deletePreviewReport.value;
  const conflicts = preview.plan.conflicts?.length ?? 0;
  const warnings = preview.plan.warnings?.length ?? 0;
  const total = preview.plan.summary?.total ?? deletePlanItems.value.length;
  try {
    await ElMessageBox.confirm(
      t('codegen.delete_execute_confirm', '即将对模块 {module} 执行删除，共 {total} 项。当前方案包含 {warnings} 条提示和 {conflicts} 个冲突。确认后将调用后端删除执行接口。', {
        module: preview.plan.module || deleteModule.value,
        total,
        warnings,
        conflicts,
      }),
      t('codegen.confirm_delete_title', '确认删除方案'),
      {
        confirmButtonText: t('codegen.confirm_execute', '确认执行'),
        cancelButtonText: t('codegen.return_modify', '返回修改'),
        type: 'warning',
        distinguishCancelAndClose: true,
      },
    );
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      ElMessage.info(t('codegen.delete_cancelled', '已取消删除操作'));
      return;
    }
    throw error;
  }
  deleteLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    const request = {
      ...deleteRequestCache.value,
      dry_run: false,
    } satisfies CodegenDeleteRequest;
    deleteRequestCache.value = request;
    deleteResult.value = await executeCodegenDelete(request);
    const status = deleteResult.value.status || '';
    lastRunSuccess.value = status === 'succeeded' || status === 'partial';
    operationStatus.value = deleteResultStatusMessage.value;
    if (status === 'failed') {
      ElMessage.error(deleteResultStatusMessage.value || '删除执行失败');
      return;
    }
    if (status === 'partial') {
      ElMessage.warning(deleteResultStatusMessage.value || '删除部分完成');
      return;
    }
    ElMessage.success(deleteResultStatusMessage.value || '删除执行完成');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('codegen.delete_failed', '删除执行失败'));
  } finally {
    deleteLoading.value = false;
  }
}

async function loadPublicConfig() {
  try {
    publicConfig.value = await fetchPublicConfig();
    applyDbConfigDefaults();
  } catch {
    publicConfig.value = null;
  }
}

function clearCurrentInputs() {
  if (activeMode.value === 'db') {
    clearDbInputs();
    return;
  }
  if (activeMode.value === 'delete') {
    clearDeleteInputs();
    return;
  }
  clearDslInputs();
}

function clearDslInputs() {
  dslText.value = '';
  packageName.value = '';
  dslReport.value = null;
  artifactInfo.value = null;
  artifactForceExpired.value = false;
  lastDownloadAt.value = '';
  lastArtifactError.value = '';
  lastArtifactErrorAt.value = '';
  lastArtifactErrorType.value = 'unknown';
  lastDownloadDuration.value = 0;
  downloadStartAt.value = 0;
  resetCopyFeedback();
  installLoading.value = false;
  operationStatus.value = '';
  lastRunSuccess.value = false;
}

function clearDeleteInputs() {
  deleteModule.value = '';
  deleteKind.value = 'crud';
  deletePolicyStore.value = '';
  deleteWithRuntime.value = true;
  deleteWithPolicy.value = true;
  deleteWithFrontend.value = true;
  deleteWithRegistry.value = true;
  deleteForce.value = false;
  deleteNotes.value = '';
  deletePreviewReport.value = null;
  deleteResult.value = null;
  deleteRequestCache.value = null;
  operationStatus.value = '';
  lastRunSuccess.value = false;
}

function clearDbTables() {
  dbTablesText.value = '';
}

function clearDbInputs() {
  dbDatabase.value = '';
  dbSchema.value = '';
  dbTablesText.value = '';
  dbForce.value = false;
  dbGenerateFrontend.value = true;
  dbGeneratePolicy.value = true;
  dbMountParentPath.value = '';
  dbReport.value = null;
  artifactInfo.value = null;
  artifactForceExpired.value = false;
  lastDownloadAt.value = '';
  lastArtifactError.value = '';
  lastArtifactErrorAt.value = '';
  lastArtifactErrorType.value = 'unknown';
  lastDownloadDuration.value = 0;
  downloadStartAt.value = 0;
  resetCopyFeedback();
  operationStatus.value = '';
  lastRunSuccess.value = false;
}

function triggerFileSelect() {
  fileInputRef.value?.click();
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0];
  if (!file) {
    return;
  }
  try {
    const content = await file.text();
    dslText.value = content;
    ElMessage.success(t('codegen.file_loaded', '已载入 {name}', { name: file.name }));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('codegen.file_read_failed', '读取 DSL 文件失败'));
  } finally {
    if (input) {
      input.value = '';
    }
  }
}

async function handlePreview() {
  if (activeMode.value === 'delete') {
    await handleDeletePreview();
    return;
  }
  if (activeMode.value === 'db') {
    const validationError = validateDatabaseInputs();
    if (validationError) {
      ElMessage.warning(validationError);
      return;
    }
  }
  previewLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    if (activeMode.value === 'db') {
      dbReport.value = await previewCodegenDatabase(buildDatabaseRequest());
      operationStatus.value = t('codegen.db_dry_run_complete', '数据库 Dry-run 预览完成');
      ElMessage.success(t('codegen.db_dry_run_complete', '数据库 Dry-run 预览完成'));
      lastRunSuccess.value = true;
      return;
    }
    if (!dslText.value.trim()) {
      ElMessage.warning(t('codegen.fill_dsl', '请先填写 DSL 内容'));
      return;
    }
    dslReport.value = await previewCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
    operationStatus.value = t('codegen.dry_run_complete', 'Dry-run 预览完成');
    ElMessage.success(t('codegen.dry_run_complete', 'Dry-run 预览完成'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('codegen.dry_run_failed', 'Dry-run 预览失败'));
  } finally {
    previewLoading.value = false;
  }
}

async function handleGenerate() {
  if (activeMode.value === 'db') {
    const validationError = validateDatabaseInputs();
    if (validationError) {
      ElMessage.warning(validationError);
      return;
    }
  }
  generateLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    if (activeMode.value === 'db') {
      dbReport.value = await generateCodegenDatabase(buildDatabaseRequest());
      lastRunSuccess.value = true;
      operationStatus.value = t('codegen.db_generated_status', '数据库代码已直接生成到服务端工程');
      ElMessage.success(t('codegen.db_generate_complete', '数据库生成已完成'));
      return;
    }
    if (!dslText.value.trim()) {
      ElMessage.warning(t('codegen.fill_dsl', '请先填写 DSL 内容'));
      return;
    }
    dslReport.value = await generateCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
    operationStatus.value = t('codegen.generated_status', '代码已直接生成到服务端工程');
    ElMessage.success(t('codegen.generate_complete', '生成已完成'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('codegen.generate_failed', '生成失败'));
  } finally {
    generateLoading.value = false;
  }
}

async function handleGenerateAndInstall() {
  if (activeMode.value !== 'db') {
    ElMessage.warning(t('codegen.switch_db_mode', '请先切换到 DB 模式'));
    return;
  }
  const validationError = validateDatabaseInputs();
  if (validationError) {
    ElMessage.warning(validationError);
    return;
  }
  generateLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    dbReport.value = await generateCodegenDatabase(buildDatabaseRequest());
    lastRunSuccess.value = true;
    operationStatus.value = t('codegen.db_generated_waiting_install', '数据库代码已生成，等待确认安装到系统');
    ElMessage.success(t('codegen.db_generate_complete', '数据库生成已完成'));

    if (!dbGeneratedModuleName.value) {
      ElMessage.warning(t('codegen.db_cannot_identify_module', '无法识别生成模块，请先重新生成'));
      return;
    }

    await ElMessageBox.confirm(
      t('codegen.db_install_prompt', '即将把模块 {module} 的 manifest 安装到系统菜单中，是否继续？', {
        module: dbGeneratedModuleName.value,
      }),
      t('codegen.install_confirm_title', '确认安装到系统'),
      {
        confirmButtonText: t('codegen.install_confirm_continue', '继续安装'),
        cancelButtonText: t('common.cancel', '取消'),
        type: 'warning',
        distinguishCancelAndClose: true,
      },
    );

    installLoading.value = true;
    const result = await installCodegenManifest({ module: dbGeneratedModuleName.value });
    lastRunSuccess.value = true;
    operationStatus.value = t('codegen.install_result_summary', '模块 {module} 已安装到系统，共 {total} 个菜单', {
      module: result.module || dbGeneratedModuleName.value,
      total: result.menu_total,
    });
    ElMessage.success(t('codegen.db_install_complete', '安装到系统完成'));
    await loadDbMountMenuOptions();
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      ElMessage.info(t('codegen.install_cancelled', '已取消安装'));
      return;
    }
    ElMessage.error(error instanceof Error ? error.message : t('codegen.generate_install_failed', '生成并安装失败'));
  } finally {
    installLoading.value = false;
    generateLoading.value = false;
  }
}

async function handleDeletePreview() {
  const validationError = validateDeleteInputs();
  if (validationError) {
    ElMessage.warning(validationError);
    return;
  }
  previewLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    const request = buildDeleteRequest(true);
    const report = await previewCodegenDelete(request);
    deletePreviewReport.value = report;
    deleteRequestCache.value = report.request ?? request;
    deleteResult.value = null;
    lastRunSuccess.value = true;
    operationStatus.value = deletePlanStatusMessage.value;
    ElMessage.success(t('codegen.delete_preview_complete', '删除预览完成'));
  } catch (error) {
    deletePreviewReport.value = null;
    deleteRequestCache.value = null;
    deleteResult.value = null;
    ElMessage.error(error instanceof Error ? error.message : t('codegen.delete_preview_failed', '删除预览失败'));
  } finally {
    previewLoading.value = false;
  }
}

async function handleGenerateDownload() {
  const isDbMode = activeMode.value === 'db';
  if (isDbMode) {
    const validationError = validateDatabaseInputs();
    if (validationError) {
      ElMessage.warning(validationError);
      return;
    }
  } else if (!dslText.value.trim()) {
    ElMessage.warning(t('codegen.fill_dsl', '请先填写 DSL 内容'));
    return;
  }
  downloadLoading.value = true;
  downloadStartAt.value = Date.now();
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    const artifact = isDbMode
      ? await generateDownloadCodegenDatabase(buildDatabaseRequest())
      : await generateDownloadCodegenDsl({
          dsl: dslText.value,
          force: force.value,
          package_name: packageName.value.trim() || undefined,
          include_readme: includeReadme.value,
          include_report: includeReport.value,
          include_dsl: includeDsl.value,
        });
    artifactInfo.value = artifact;
    artifactForceExpired.value = false;
    lastRunSuccess.value = true;
    lastArtifactError.value = '';
    lastArtifactErrorAt.value = '';
    lastArtifactErrorType.value = 'unknown';
    operationStatus.value = isDbMode
      ? `数据库代码包已生成，共 ${artifact.file_count} 个文件，浏览器将开始下载 ${artifact.filename}`
      : `代码包已生成，共 ${artifact.file_count} 个文件，浏览器将开始下载 ${artifact.filename}`;
    await downloadCodegenArtifact(artifact.download_url, artifact.filename);
    lastDownloadAt.value = new Date().toISOString();
    lastDownloadDuration.value = downloadStartAt.value ? Date.now() - downloadStartAt.value : 0;
    ElMessage.success(t('codegen.download_ready_prefix', '下载已开始：') + artifact.filename);
  } catch (error) {
    handleArtifactError(
      error,
      isDbMode ? t('codegen.generate_database_download_failed', '生成数据库下载包失败') : t('codegen.generate_download_failed', '生成下载包失败'),
    );
  } finally {
    downloadLoading.value = false;
    downloadStartAt.value = 0;
  }
}

async function handleArtifactDownload() {
  if (!artifactInfo.value) {
    ElMessage.warning(t('codegen.no_artifact', '暂无可下载产物'));
    return;
  }
  if (isArtifactExpired.value) {
    artifactForceExpired.value = true;
    ElMessage.warning(t('codegen.artifact_expired_short', '下载包已过期，请重新执行“生成并下载”'));
    return;
  }
  downloadLoading.value = true;
  downloadStartAt.value = Date.now();
  try {
    await downloadCodegenArtifact(artifactInfo.value.download_url, artifactInfo.value.filename);
    lastDownloadAt.value = new Date().toISOString();
    lastDownloadDuration.value = downloadStartAt.value ? Date.now() - downloadStartAt.value : 0;
    lastArtifactError.value = '';
    lastArtifactErrorAt.value = '';
    lastArtifactErrorType.value = 'unknown';
    ElMessage.success(t('codegen.download_ready_prefix', '下载已开始：') + artifactInfo.value.filename);
  } catch (error) {
    handleArtifactError(error, t('codegen.download_failed', '下载失败'));
  } finally {
    downloadLoading.value = false;
    downloadStartAt.value = 0;
  }
}

async function handleCopyDownloadUrl() {
  if (!artifactInfo.value || !canCopyArtifactUrl.value) {
    ElMessage.warning(t('codegen.no_copy_url', '暂无可复制地址'));
    return;
  }
  const value = artifactDownloadUrlText.value;
  if (!value) {
    ElMessage.warning(t('codegen.empty_download_url', '下载地址为空'));
    return;
  }
  try {
    await copyText(value);
    triggerCopyFeedback();
    ElMessage.success(t('codegen.copied', '下载地址已复制'));
  } catch (error) {
    resetCopyFeedback();
    ElMessage.error(error instanceof Error ? error.message : t('codegen.copy_failed', '复制下载地址失败'));
  }
}

function handleArtifactError(error: unknown, fallbackMessage: string) {
  if (error instanceof ApiError) {
    switch (normalizeHttpStatus(error.code)) {
      case 401:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'auth';
        rememberArtifactError(t('codegen.download_auth_error', '登录状态已失效，请重新登录后再下载代码包'));
        ElMessage.error(t('codegen.download_auth_error', '登录状态已失效，请重新登录后再下载代码包'));
        return;
      case 404:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'notfound';
        rememberArtifactError(t('codegen.download_notfound_error', '下载包不存在，可能已被清理，请重新执行“生成并下载”'));
        ElMessage.error(t('codegen.download_notfound_error', '下载包不存在，可能已被清理，请重新执行“生成并下载”'));
        return;
      case 410:
        artifactForceExpired.value = true;
        operationStatus.value = t('codegen.artifact_expired_short', '下载包已过期，请重新执行“生成并下载”。');
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'expired';
        rememberArtifactError(t('codegen.artifact_expired_short', '下载包已过期，请重新执行“生成并下载”'));
        ElMessage.warning(t('codegen.artifact_expired_short', '下载包已过期，请重新执行“生成并下载”'));
        return;
      case 500:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'server';
        rememberArtifactError(t('codegen.download_server_error', '下载服务暂时不可用，请稍后重试'));
        ElMessage.error(t('codegen.download_server_error', '下载服务暂时不可用，请稍后重试'));
        return;
      default:
        break;
    }
  }
  lastArtifactErrorType.value = 'unknown';
  rememberArtifactError(error instanceof Error ? error.message : fallbackMessage);
  ElMessage.error(error instanceof Error ? error.message : fallbackMessage);
}

function buildDatabaseRequest(): CodegenDatabaseRequest {
  const tables = parseTableNames(dbTablesText.value);
  return {
    driver: dbDriver.value.trim(),
    database: dbDatabase.value.trim(),
    schema: dbSchema.value.trim() || undefined,
    tables: tables.length > 0 ? tables : undefined,
    force: dbForce.value,
    generate_frontend: dbGenerateFrontend.value,
    generate_policy: dbGeneratePolicy.value,
    mount_parent_path: dbMountParentPath.value.trim() || undefined,
  };
}

async function loadDbMountMenuOptions() {
  try {
    const response = await fetchMenuTree();
    dbMountMenuOptions.value = flattenMenuMountOptions(response.items ?? []);
  } catch {
    dbMountMenuOptions.value = [];
  }
}

function flattenMenuMountOptions(items: MenuItem[], depth = 0): MenuMountOption[] {
  const options: MenuMountOption[] = [];
  for (const item of items) {
    const name = item.name?.trim() || item.path || t('codegen.unnamed_menu', '未命名菜单');
    const labelPrefix = depth > 0 ? `${'—'.repeat(depth)} ` : '';
    if ((item.type || '').toLowerCase() === 'directory') {
      options.push({
        label: `${labelPrefix}${name}${item.path ? ` (${item.path})` : ''}`,
        value: item.path,
      });
    }
    if (item.children?.length) {
      options.push(...flattenMenuMountOptions(item.children, depth + 1));
    }
  }
  return options;
}

function validateDatabaseInputs(): string {
  if (!dbDriver.value.trim()) {
    return t('codegen.db_validate_driver', '请先选择数据库驱动');
  }
  if (!dbDatabase.value.trim()) {
    return t('codegen.db_validate_name', '请先填写数据库名');
  }
  return '';
}

function validateDeleteInputs(): string {
  if (!deleteModule.value.trim()) {
    return t('codegen.delete_validate_module', '请先填写要删除的模块名');
  }
  return '';
}

function buildDeleteRequest(dryRun: boolean): CodegenDeleteRequest {
  return {
    module: deleteModule.value.trim(),
    kind: deleteKind.value.trim() || 'crud',
    dry_run: dryRun,
    force: deleteForce.value,
    with_policy: deleteWithPolicy.value,
    with_runtime: deleteWithRuntime.value,
    with_frontend: deleteWithFrontend.value,
    with_registry: deleteWithRegistry.value,
    policy_store: deletePolicyStore.value.trim() || undefined,
    metadata_hints: deleteNotes.value.trim()
      ? {
          notes: deleteNotes.value.trim(),
        }
      : undefined,
  };
}

function parseTableNames(value: string): string[] {
  return value
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter((item) => item.length > 0);
}

function mapDatabaseResourcesToPreviewRows(resources: CodegenDatabasePreviewResource[], forceValue: boolean): PreviewRow[] {
  return resources.map((resource, index) => ({
    index: index + 1,
    kind: resource.kind || 'resource',
    name: resource.name || resource.entity_name || resource.table_name,
    force: forceValue,
    actions: resource.actions ?? [],
  }));
}

function mapDeletePlanToPreviewRows(plan: CodegenDeletePlan | null): PreviewRow[] {
  if (!plan) {
    return [];
  }
  const items = [
    ...(plan.source_files ?? []),
    ...(plan.runtime_assets ?? []),
    ...(plan.registry_changes ?? []),
    ...(plan.policy_changes ?? []),
    ...(plan.frontend_changes ?? []),
  ];
  return items.map((item, index) => ({
    index: index + 1,
    kind: item.kind || 'asset',
    name: item.path || item.ref || item.module || 'asset',
    force: plan.force ?? false,
    managed: item.managed ?? false,
    actions: buildDeleteItemActions(item),
  }));
}

function buildDeleteItemActions(item: CodegenDeletePlanItem): string[] {
  const actions: string[] = [];
  if (item.origin) {
    actions.push(`origin:${item.origin}`);
  }
  actions.push(item.managed ? 'managed' : 'manual');
  if (item.store) {
    actions.push(`store:${item.store}`);
  }
  return actions;
}

function describeDeleteItem(item: CodegenDeletePlanItem): string {
  const parts = [item.kind || 'asset'];
  if (item.path) {
    parts.push(item.path);
  } else if (item.ref) {
    parts.push(item.ref);
  }
  if (item.origin) {
    parts.push(`origin=${item.origin}`);
  }
  if (item.managed !== undefined) {
    parts.push(item.managed ? 'managed' : 'manual');
  }
  return parts.join(' · ');
}

function deleteItemKey(item: CodegenDeletePlanItem): string {
  return [item.kind, item.path, item.ref, item.origin].filter((value): value is string => Boolean(value)).join('::');
}

function describeDeleteFailure(failure: { reason?: string; item?: CodegenDeletePlanItem }): string {
  const parts = [failure.reason || '删除失败'];
  if (failure.item) {
    parts.push(describeDeleteItem(failure.item));
  }
  return parts.join(' · ');
}

function describeDeleteFailureKey(failure: { reason?: string; item?: CodegenDeletePlanItem }): string {
  return `${failure.reason || 'failure'}::${failure.item ? deleteItemKey(failure.item) : 'none'}`;
}

function formatElapsedMillis(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '-';
  }
  if (value < 1000) {
    return `${Math.round(value)}ms`;
  }
  const seconds = value / 1000;
  return `${seconds.toFixed(seconds >= 10 ? 1 : 2)}s`;
}

function rememberArtifactError(message: string) {
  lastArtifactError.value = message;
  lastArtifactErrorAt.value = new Date().toISOString();
}

function triggerCopyFeedback() {
  copyFeedbackActive.value = true;
  if (copyFeedbackTimer !== null) {
    window.clearTimeout(copyFeedbackTimer);
  }
  copyFeedbackTimer = window.setTimeout(() => {
    copyFeedbackActive.value = false;
    copyFeedbackTimer = null;
  }, 1600);
}

function resetCopyFeedback() {
  copyFeedbackActive.value = false;
  if (copyFeedbackTimer !== null) {
    window.clearTimeout(copyFeedbackTimer);
    copyFeedbackTimer = null;
  }
}

function normalizeHttpStatus(code: number): number {
  if (code >= 100 && code < 600) {
    return code;
  }
  if (code >= 40000 && code < 40100) {
    return 400;
  }
  if (code >= 40100 && code < 40200) {
    return 401;
  }
  if (code >= 40300 && code < 40400) {
    return 403;
  }
  if (code >= 40400 && code < 40500) {
    return 404;
  }
  if (code >= 41000 && code < 41100) {
    return 410;
  }
  if (code >= 50000 && code < 50100) {
    return 500;
  }
  return 0;
}

async function copyText(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }
  const input = document.createElement('textarea');
  input.value = value;
  input.setAttribute('readonly', 'true');
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  input.select();
  const successful = document.execCommand('copy');
  document.body.removeChild(input);
  if (!successful) {
    throw new Error('浏览器不支持自动复制，请手动复制');
  }
}

function toAbsoluteUrl(value: string): string {
  const path = value.trim();
  if (!path) {
    return '';
  }
  if (/^https?:\/\//i.test(path)) {
    return path;
  }
  return new URL(path, window.location.origin).toString();
}

function summarizeDownloadUrl(value: string): string {
  if (!value) {
    return '';
  }
  try {
    const parsed = new URL(value);
    return `${parsed.host} · ${summarizePath(parsed.pathname)}${parsed.search}`;
  } catch {
    return value;
  }
}

function summarizePath(value: string): string {
  if (!value) {
    return '/';
  }
  const normalized = value.length > 48 ? `${value.slice(0, 24)}...${value.slice(-18)}` : value;
  return normalized;
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value <= 0) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  const digits = index === 0 ? 0 : 2;
  return `${size.toFixed(digits)} ${units[index]}`;
}

function formatDateTime(value: string): string {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function formatRemainingTime(value: string, now: number, expired: boolean): string {
  if (!value) {
    return '-';
  }
  if (expired) {
    return '已过期';
  }
  const expiresAt = new Date(value).getTime();
  if (Number.isNaN(expiresAt)) {
    return '-';
  }
  const diff = Math.max(0, expiresAt - now);
  if (diff <= 0) {
    return '已过期';
  }
  const totalSeconds = Math.floor(diff / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (days > 0) {
    return `${days}天 ${hours}小时 ${minutes}分钟`;
  }
  if (hours > 0) {
    return `${hours}小时 ${minutes}分钟 ${seconds}秒`;
  }
  if (minutes > 0) {
    return `${minutes}分钟 ${seconds}秒`;
  }
  return `${seconds}秒`;
}
</script>

<style scoped>
.codegen-page {
  padding: 16px;
}

.codegen-grid {
  align-items: flex-start;
}

.codegen-grid :deep(.el-col) {
  min-width: 0;
}

.codegen-card {
  border-radius: 14px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.card-header.compact {
  justify-content: flex-start;
}

.artifact-header {
  justify-content: space-between;
}

.title {
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.subtitle {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.db-mode-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.db-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 16px;
  background: linear-gradient(135deg, var(--el-fill-color-blank) 0%, var(--el-fill-color-extra-light) 100%);
}

.db-hero-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.db-hero-subtitle {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.db-preset-row,
.db-section-header,
.db-form-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.db-section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.db-section-hint,
.db-field-help {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.db-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.db-form :deep(.el-row) {
  margin-bottom: 0;
}

.db-form :deep(.el-col) {
  min-width: 0;
}

.db-form :deep(.el-input),
.db-form :deep(.el-select),
.db-form :deep(.el-textarea),
.db-form :deep(.el-input__wrapper),
.db-form :deep(.el-select__wrapper) {
  width: 100%;
}

.db-form-item {
  margin-bottom: 12px;
}

.db-form-item--wide {
  margin-bottom: 0;
}

.db-field-help {
  margin-top: 8px;
}

.db-table-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.db-table-chip {
  border-radius: 999px;
}

.db-advanced {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: var(--el-fill-color-extra-light);
}

.db-advanced :deep(.el-space) {
  width: 100%;
}

.codegen-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.hidden-file-input {
  display: none;
}

.side-stack {
  width: 100%;
}

.side-stack :deep(.el-card__body) {
  min-width: 0;
}

.result-empty {
  padding: 20px 0;
  color: var(--el-text-color-secondary);
  text-align: center;
}

.preview-table-wrap {
  margin-top: 16px;
  width: 100%;
  overflow-x: auto;
}

.preview-table {
  min-width: 820px;
}

.action-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-width: 100%;
}

.action-tag {
  margin-right: 0;
}

.message-list {
  margin: 0;
  padding-left: 18px;
  color: var(--el-text-color-primary);
  line-height: 1.7;
  word-break: break-word;
}

.artifact-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.artifact-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.artifact-summary-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
  padding: 14px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: linear-gradient(180deg, var(--el-fill-color-blank) 0%, var(--el-fill-color-extra-light) 100%);
}

.artifact-summary-card--success {
  border-color: var(--el-color-success-light-5);
  background: linear-gradient(180deg, var(--el-color-success-light-9) 0%, var(--el-fill-color-blank) 100%);
}

.artifact-summary-card--warning {
  border-color: var(--el-color-warning-light-5);
  background: linear-gradient(180deg, var(--el-color-warning-light-9) 0%, var(--el-fill-color-blank) 100%);
}

.artifact-summary-card--danger {
  border-color: var(--el-color-danger-light-5);
  background: linear-gradient(180deg, var(--el-color-danger-light-9) 0%, var(--el-fill-color-blank) 100%);
}

.artifact-summary-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.artifact-summary-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  word-break: break-word;
}

.artifact-summary-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-word;
}

.artifact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.artifact-grid--compact {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.artifact-item {
  padding: 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  background: var(--el-fill-color-extra-light);
}

.artifact-item--wide {
  grid-column: 1 / -1;
}

.artifact-label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.artifact-value {
  display: block;
  color: var(--el-text-color-primary);
  word-break: break-word;
}

.artifact-hint {
  display: block;
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.artifact-copy-button--active {
  color: var(--el-color-success) !important;
}

.monospace {
  font-family: var(--el-font-family-monospace, SFMono-Regular, Consolas, 'Liberation Mono', Menlo, monospace);
}

@media (max-width: 960px) {
  .artifact-summary-grid,
  .artifact-grid {
    grid-template-columns: 1fr;
  }
}
</style>
