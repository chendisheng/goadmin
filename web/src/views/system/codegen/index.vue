<template>
  <div class="codegen-page">
    <el-row :gutter="20" class="codegen-grid">
      <el-col :xs="24" :lg="11" :xl="11">
        <el-card shadow="never" class="codegen-card">
          <template #header>
            <div class="card-header">
              <div>
                <div class="title">CodeGen Console</div>
                <div class="subtitle">在同一页面中切换 DSL 与 DB 输入模式，复用统一结果区。</div>
              </div>
              <el-space wrap>
                <el-button v-if="activeMode === 'dsl'" @click="loadSample">载入示例</el-button>
                <el-button v-else @click="loadDbSample">载入示例</el-button>
                <el-button @click="clearCurrentInputs">清空</el-button>
                <el-button v-if="activeMode === 'dsl'" @click="triggerFileSelect">上传 DSL</el-button>
                <el-button v-if="activeMode === 'dsl'" type="primary" :loading="previewLoading" @click="handlePreview">Dry-run 预览</el-button>
                <el-button v-if="activeMode === 'dsl'" type="success" :loading="generateLoading" @click="handleGenerate">一键生成</el-button>
                <el-button v-if="activeMode === 'dsl'" type="warning" :loading="downloadLoading" @click="handleGenerateDownload">生成并下载</el-button>
                <el-button v-if="activeMode === 'db'" type="primary" :loading="previewLoading" @click="handlePreview">Dry-run 预览</el-button>
                <el-button v-if="activeMode === 'db'" type="success" :loading="generateLoading" @click="handleGenerate">一键生成</el-button>
              </el-space>
            </div>
          </template>

          <el-tabs v-model="activeMode" class="codegen-tabs" stretch>
            <el-tab-pane label="DSL" name="dsl">
              <el-form label-position="top" class="codegen-form">
                <el-form-item label="Force overwrite">
                  <el-switch v-model="force" inline-prompt active-text="On" inactive-text="Off" />
                </el-form-item>
                <el-form-item label="下载包名称">
                  <el-input v-model="packageName" placeholder="留空则由系统自动生成 zip 名称" />
                </el-form-item>
                <el-form-item label="下载包内容">
                  <el-space wrap>
                    <el-switch v-model="includeReadme" inline-prompt active-text="README" inactive-text="README" />
                    <el-switch v-model="includeReport" inline-prompt active-text="Report" inactive-text="Report" />
                    <el-switch v-model="includeDsl" inline-prompt active-text="DSL" inactive-text="DSL" />
                  </el-space>
                </el-form-item>
                <el-form-item label="DSL 内容">
                  <el-input
                    v-model="dslText"
                    type="textarea"
                    :rows="28"
                    resize="none"
                    placeholder="在这里粘贴或编辑 DSL YAML"
                  />
                </el-form-item>
              </el-form>
              <input ref="fileInputRef" class="hidden-file-input" type="file" accept=".yaml,.yml,.json,.txt" @change="handleFileChange" />
            </el-tab-pane>

            <el-tab-pane label="DB" name="db">
              <div class="db-mode-panel">
                <div class="db-hero">
                  <div>
                    <div class="db-hero-title">数据库输入向导</div>
                    <div class="db-hero-subtitle">先选驱动，再填连接串与扫描范围。建议先预览，确认结果后再生成。</div>
                  </div>
                  <el-space wrap>
                    <el-tag type="info" effect="light">{{ dbDriverLabel }}</el-tag>
                    <el-tag type="success" effect="light">{{ dbParsedTables.length ? `${dbParsedTables.length} 个表` : '全部表' }}</el-tag>
                  </el-space>
                </div>

                <el-alert
                  title="推荐先执行 Dry-run 预览，确认文件计划和冲突后再执行生成。"
                  type="info"
                  :closable="false"
                  show-icon
                />

                <div class="db-preset-row">
                  <div>
                    <div class="db-section-title">快速模板</div>
                    <div class="db-section-hint">一键预填常见数据库的连接格式。</div>
                  </div>
                  <el-space wrap>
                    <el-button size="small" :type="dbDriver === 'mysql' ? 'primary' : 'default'" plain @click="applyDbPreset('mysql')">MySQL</el-button>
                    <el-button size="small" :type="dbDriver === 'postgres' ? 'primary' : 'default'" plain @click="applyDbPreset('postgres')">PostgreSQL</el-button>
                    <el-button size="small" :type="dbDriver === 'sqlite' ? 'primary' : 'default'" plain @click="applyDbPreset('sqlite')">SQLite</el-button>
                  </el-space>
                </div>

                <el-form label-position="top" class="codegen-form db-form">
                  <el-row :gutter="16">
                    <el-col :xs="24" :md="8">
                      <el-form-item label="数据库驱动">
                        <el-select v-model="dbDriver" placeholder="请选择数据库驱动" filterable>
                          <el-option label="MySQL" value="mysql" />
                          <el-option label="PostgreSQL" value="postgres" />
                          <el-option label="SQLite" value="sqlite" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="8">
                      <el-form-item label="数据库名">
                        <el-input v-model="dbDatabase" placeholder="请输入数据库名称" />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="8">
                      <el-form-item label="Schema">
                        <el-input v-model="dbSchema" placeholder="可选，PostgreSQL 等场景使用" />
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-form-item label="DSN" class="db-form-item">
                    <el-input
                      v-model="dbDsn"
                      type="password"
                      show-password
                      placeholder="请输入数据库连接串"
                    >
                      <template #append>
                        <el-button text @click="dbDsn = ''">清空</el-button>
                      </template>
                    </el-input>
                    <div class="db-field-help">
                      <span>连接串仅用于当前页面请求，不会在结果区明文回显。</span>
                    </div>
                  </el-form-item>

                  <el-form-item label="表名范围" class="db-form-item db-form-item--wide">
                    <el-input
                      v-model="dbTablesText"
                      type="textarea"
                      :rows="6"
                      resize="none"
                      placeholder="支持逗号、换行分隔，例如：books, orders"
                    />
                    <div class="db-form-row">
                      <span class="db-field-help">留空则表示扫描全部表；建议优先预填少量表进行预览。</span>
                      <el-space wrap>
                        <el-button text size="small" @click="loadDbSample">载入示例</el-button>
                        <el-button text size="small" @click="clearDbTables">清空范围</el-button>
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
                        <div class="db-section-title">生成选项</div>
                        <div class="db-section-hint">控制是否覆盖现有文件、是否输出前端和权限策略。</div>
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
          </el-tabs>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="13" :xl="13">
        <el-space direction="vertical" size="16" fill class="side-stack">
          <el-card shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">执行结果</div>
                  <div class="subtitle">预览和生成都会回传资源级动作。</div>
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
            <div v-else class="result-empty">尚未执行预览或生成。</div>

            <div class="preview-table-wrap">
              <el-table :data="previewItems" class="preview-table" size="small" border>
                <el-table-column prop="index" label="#" width="60" />
                <el-table-column prop="kind" label="Kind" min-width="130" show-overflow-tooltip />
                <el-table-column prop="name" label="Name" min-width="160" show-overflow-tooltip />
                <el-table-column label="Force" width="88">
                  <template #default="scope">
                    <el-tag :type="scope.row.force ? 'warning' : 'info'" effect="light">
                      {{ scope.row.force ? 'Yes' : 'No' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="Actions" min-width="280">
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
                  <div class="title">文件计划</div>
                  <div class="subtitle">显示数据库预览阶段推导出的文件清单。</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!filePlans.length" description="暂无文件计划" />
            <el-table v-else :data="filePlans" size="small" border class="preview-table">
              <el-table-column prop="path" label="Path" min-width="220" show-overflow-tooltip />
              <el-table-column prop="action" label="Action" width="120" />
              <el-table-column prop="kind" label="Kind" width="140" show-overflow-tooltip />
              <el-table-column label="Exists" width="88">
                <template #default="scope">
                  <el-tag :type="scope.row.exists ? 'warning' : 'success'" effect="light">
                    {{ scope.row.exists ? 'Yes' : 'No' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="Conflict" width="96">
                <template #default="scope">
                  <el-tag :type="scope.row.conflict ? 'danger' : 'success'" effect="light">
                    {{ scope.row.conflict ? 'Yes' : 'No' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card v-if="activeMode === 'db'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">冲突</div>
                  <div class="subtitle">展示文件覆盖风险与路径冲突信息。</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!conflicts.length" description="暂无冲突" />
            <el-table v-else :data="conflicts" size="small" border class="preview-table">
              <el-table-column prop="path" label="Path" min-width="220" show-overflow-tooltip />
              <el-table-column prop="resource" label="Resource" min-width="160" show-overflow-tooltip />
              <el-table-column prop="reason" label="Reason" min-width="260" show-overflow-tooltip />
            </el-table>
          </el-card>

          <el-card v-if="activeMode === 'db'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">审计</div>
                  <div class="subtitle">记录输入、执行步骤与输出统计。</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!auditRecord" description="暂无审计记录" />
            <div v-else class="artifact-panel">
              <div class="artifact-summary-grid">
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">记录时间</span>
                  <span class="artifact-summary-value">{{ formatDateTime(auditRecord.recorded_at) }}</span>
                  <span class="artifact-summary-meta">{{ auditRecord.recorded_at }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">输入</span>
                  <span class="artifact-summary-value">{{ auditRecord.input.driver }} / {{ auditRecord.input.database }}</span>
                  <span class="artifact-summary-meta">dry-run: {{ auditRecord.input.dry_run ? 'yes' : 'no' }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">输出文件</span>
                  <span class="artifact-summary-value">{{ auditRecord.output.file_count }}</span>
                  <span class="artifact-summary-meta">冲突：{{ auditRecord.output.conflict_count }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">表范围</span>
                  <span class="artifact-summary-value">{{ auditRecord.input.tables?.length ?? 0 }}</span>
                  <span class="artifact-summary-meta">{{ (auditRecord.input.tables ?? []).join(', ') || '全部表' }}</span>
                </div>
              </div>
              <div class="artifact-grid artifact-grid--compact">
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">执行步骤</span>
                  <ul class="message-list compact-list">
                    <li v-for="step in auditRecord.steps" :key="step.name">
                      {{ step.name }} [{{ step.status }}]{{ step.detail ? ` - ${step.detail}` : '' }}
                    </li>
                  </ul>
                </div>
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">输出概览</span>
                  <span class="artifact-value">文件 {{ auditRecord.output.files.length }} · 冲突 {{ auditRecord.output.conflicts.length }}</span>
                </div>
              </div>
            </div>
          </el-card>

          <el-card shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact">
                <div>
                  <div class="title">消息</div>
                  <div class="subtitle">包含 dry-run 提示、生成摘要和校验信息。</div>
                </div>
              </div>
            </template>

            <el-empty v-if="!messages.length" description="暂无消息" />
            <ul v-else class="message-list">
              <li v-for="message in messages" :key="message">{{ message }}</li>
            </ul>
          </el-card>

          <el-card v-if="activeMode === 'dsl'" shadow="never" class="codegen-card">
            <template #header>
              <div class="card-header compact artifact-header">
                <div>
                  <div class="title">下载产物</div>
                  <div class="subtitle">展示最近一次服务端打包结果，并支持重新下载。</div>
                </div>
                <el-space v-if="artifactInfo" wrap>
                  <el-button text :class="{ 'artifact-copy-button--active': copyFeedbackActive }" :disabled="!canCopyArtifactUrl" @click="handleCopyDownloadUrl">{{ copyButtonText }}</el-button>
                  <el-button text type="primary" :disabled="isArtifactExpired" :loading="downloadLoading" @click="handleArtifactDownload">重新下载</el-button>
                </el-space>
              </div>
            </template>

            <el-empty v-if="!artifactInfo" description="尚未生成下载包" />
            <div v-else class="artifact-panel">
              <el-alert
                :title="artifactStatusMessage"
                :type="artifactStatusType"
                :closable="false"
                show-icon
              />
              <div class="artifact-summary-grid">
                <div class="artifact-summary-card" :class="`artifact-summary-card--${artifactStatusTone}`">
                  <span class="artifact-summary-label">产物状态</span>
                  <span class="artifact-summary-value">{{ artifactStatusLabel }}</span>
                  <span class="artifact-summary-meta">{{ artifactStatusSummary }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">文件概览</span>
                  <span class="artifact-summary-value">{{ artifactInfo.filename }}</span>
                  <span class="artifact-summary-meta">{{ artifactSizeText }} · {{ artifactInfo.file_count }} 个文件</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">有效期</span>
                  <span class="artifact-summary-value">{{ artifactRemainingText }}</span>
                  <span class="artifact-summary-meta">到期于 {{ artifactExpireText }}</span>
                </div>
                <div class="artifact-summary-card">
                  <span class="artifact-summary-label">最近活动</span>
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
                  <span class="artifact-label">任务 ID</span>
                  <span class="artifact-value monospace">{{ artifactInfo.task_id }}</span>
                </div>
                <div class="artifact-item">
                  <span class="artifact-label">文件名</span>
                  <span class="artifact-value">{{ artifactInfo.filename }}</span>
                </div>
                <div class="artifact-item">
                  <span class="artifact-label">最近一次失败原因</span>
                  <span class="artifact-value">{{ artifactLastErrorText }}</span>
                </div>
                <div class="artifact-item artifact-item--wide">
                  <span class="artifact-label">下载地址</span>
                  <template v-if="canCopyArtifactUrl">
                    <el-button plain size="small" :class="{ 'artifact-copy-button--active': copyFeedbackActive }" @click="handleCopyDownloadUrl">{{ copyButtonText }}</el-button>
                    <span class="artifact-hint">{{ artifactDownloadUrlSummary }}</span>
                  </template>
                  <span v-else class="artifact-hint">下载包已过期，旧下载地址已隐藏，请重新生成新的代码包。</span>
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
import { ElMessage } from 'element-plus';

import {
  downloadCodegenArtifact,
  generateCodegenDatabase,
  generateCodegenDsl,
  generateDownloadCodegenDsl,
  previewCodegenDatabase,
  previewCodegenDsl,
  type CodegenArtifactInfo,
  type CodegenDatabasePreviewReport,
  type CodegenDatabasePreviewResource,
  type CodegenDatabaseRequest,
  type CodegenDslExecutionReport,
} from '@/api/codegen';
import { ApiError } from '@/api/types';

type CodegenMode = 'dsl' | 'db';

type PreviewRow = {
  index: number;
  kind: string;
  name: string;
  force: boolean;
  actions: string[];
};

const activeMode = ref<CodegenMode>('dsl');

const dslText = ref('');
const force = ref(false);
const packageName = ref('');
const includeReadme = ref(true);
const includeReport = ref(true);
const includeDsl = ref(true);

const dbDriver = ref('mysql');
const dbDsn = ref('');
const dbDatabase = ref('');
const dbSchema = ref('');
const dbTablesText = ref('');
const dbForce = ref(false);
const dbGenerateFrontend = ref(true);
const dbGeneratePolicy = ref(true);

const previewLoading = ref(false);
const generateLoading = ref(false);
const downloadLoading = ref(false);

const dslReport = ref<CodegenDslExecutionReport | null>(null);
const dbReport = ref<CodegenDatabasePreviewReport | null>(null);
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

const currentReport = computed(() => (activeMode.value === 'db' ? dbReport.value : dslReport.value));
const dbParsedTables = computed(() => parseTableNames(dbTablesText.value));
const dbDriverLabel = computed(() => {
  const map: Record<string, string> = {
    mysql: 'MySQL',
    postgres: 'PostgreSQL',
    sqlite: 'SQLite',
  };
  return map[dbDriver.value] || dbDriver.value || '未选择';
});
const dbOptionSummary = computed(() => {
  const parts: string[] = [];
  parts.push(dbForce.value ? '覆盖现有文件' : '安全覆盖关闭');
  parts.push(dbGenerateFrontend.value ? '生成前端' : '跳过前端');
  parts.push(dbGeneratePolicy.value ? '生成权限' : '跳过权限');
  return parts.join(' · ');
});

const previewItems = computed<PreviewRow[]>(() => {
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
    return '正在准备下载包，请稍候，浏览器即将开始下载。';
  }
  if (isArtifactExpired.value) {
    return '当前下载包已过期，请重新执行“生成并下载”以获得新的代码包。';
  }
  return '下载包已就绪，你可以重新下载，或复制下载地址用于当前登录态调试。';
});
const artifactStatusLabel = computed(() => {
  if (downloadLoading.value) {
    return '下载准备中';
  }
  if (isArtifactExpired.value) {
    return '已过期';
  }
  return '可下载';
});
const artifactStatusSummary = computed(() => {
  if (downloadLoading.value) {
    return '浏览器即将开始下载';
  }
  if (isArtifactExpired.value) {
    return '需要重新生成新的代码包';
  }
  return '支持重新下载和复制完整地址';
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
    auth: '登录失效',
    notfound: '资源不存在',
    expired: '已过期',
    server: '服务异常',
    unknown: '其他',
  };
  return map[lastArtifactErrorType.value] || '其他';
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
  ElMessage.success('示例 DSL 已载入');
}

function loadDbSample() {
  activeMode.value = 'db';
  applyDbPreset('sqlite');
  dbDatabase.value = 'codegen';
  dbSchema.value = '';
  dbTablesText.value = 'books, orders';
  dbForce.value = false;
  dbGenerateFrontend.value = true;
  dbGeneratePolicy.value = true;
  ElMessage.success('示例数据库配置已载入');
}

function applyDbPreset(driver: 'mysql' | 'postgres' | 'sqlite') {
  const presets: Record<'mysql' | 'postgres' | 'sqlite', { dsn: string; database: string; schema: string }> = {
    mysql: {
      dsn: 'root:password@tcp(127.0.0.1:3306)/goadmin?charset=utf8mb4&parseTime=True&loc=Local',
      database: 'goadmin',
      schema: '',
    },
    postgres: {
      dsn: 'postgres://postgres:password@127.0.0.1:5432/goadmin?sslmode=disable',
      database: 'goadmin',
      schema: 'public',
    },
    sqlite: {
      dsn: 'file:./tmp/codegen.db?cache=shared&mode=rwc',
      database: 'codegen',
      schema: '',
    },
  };
  const preset = presets[driver];
  dbDriver.value = driver;
  dbDsn.value = preset.dsn;
  dbDatabase.value = preset.database;
  dbSchema.value = preset.schema;
}

function clearCurrentInputs() {
  if (activeMode.value === 'db') {
    clearDbInputs();
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
  operationStatus.value = '';
  lastRunSuccess.value = false;
}

function clearDbTables() {
  dbTablesText.value = '';
}

function clearDbInputs() {
  dbDsn.value = '';
  dbDatabase.value = '';
  dbSchema.value = '';
  dbTablesText.value = '';
  dbForce.value = false;
  dbGenerateFrontend.value = true;
  dbGeneratePolicy.value = true;
  dbReport.value = null;
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
    ElMessage.success(`已载入 ${file.name}`);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '读取 DSL 文件失败');
  } finally {
    if (input) {
      input.value = '';
    }
  }
}

async function handlePreview() {
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
      operationStatus.value = '数据库 Dry-run 预览完成';
      ElMessage.success('数据库 Dry-run 预览完成');
      lastRunSuccess.value = true;
      return;
    }
    if (!dslText.value.trim()) {
      ElMessage.warning('请先填写 DSL 内容');
      return;
    }
    dslReport.value = await previewCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
    operationStatus.value = 'Dry-run 预览完成';
    ElMessage.success('Dry-run 预览完成');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : 'Dry-run 预览失败');
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
      operationStatus.value = '数据库代码已直接生成到服务端工程';
      ElMessage.success('数据库生成已完成');
      return;
    }
    if (!dslText.value.trim()) {
      ElMessage.warning('请先填写 DSL 内容');
      return;
    }
    dslReport.value = await generateCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
    operationStatus.value = '代码已直接生成到服务端工程';
    ElMessage.success('生成已完成');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '生成失败');
  } finally {
    generateLoading.value = false;
  }
}

async function handleGenerateDownload() {
  if (!dslText.value.trim()) {
    ElMessage.warning('请先填写 DSL 内容');
    return;
  }
  downloadLoading.value = true;
  downloadStartAt.value = Date.now();
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    const artifact = await generateDownloadCodegenDsl({
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
    operationStatus.value = `代码包已生成，共 ${artifact.file_count} 个文件，浏览器将开始下载 ${artifact.filename}`;
    await downloadCodegenArtifact(artifact.download_url, artifact.filename);
    lastDownloadAt.value = new Date().toISOString();
    lastDownloadDuration.value = downloadStartAt.value ? Date.now() - downloadStartAt.value : 0;
    ElMessage.success(`下载已开始：${artifact.filename}`);
  } catch (error) {
    handleArtifactError(error, '生成下载包失败');
  } finally {
    downloadLoading.value = false;
    downloadStartAt.value = 0;
  }
}

async function handleArtifactDownload() {
  if (!artifactInfo.value) {
    ElMessage.warning('暂无可下载产物');
    return;
  }
  if (isArtifactExpired.value) {
    artifactForceExpired.value = true;
    ElMessage.warning('下载包已过期，请重新执行“生成并下载”');
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
    ElMessage.success(`下载已开始：${artifactInfo.value.filename}`);
  } catch (error) {
    handleArtifactError(error, '下载失败');
  } finally {
    downloadLoading.value = false;
    downloadStartAt.value = 0;
  }
}

async function handleCopyDownloadUrl() {
  if (!artifactInfo.value || !canCopyArtifactUrl.value) {
    ElMessage.warning('暂无可复制地址');
    return;
  }
  const value = artifactDownloadUrlText.value;
  if (!value) {
    ElMessage.warning('下载地址为空');
    return;
  }
  try {
    await copyText(value);
    triggerCopyFeedback();
    ElMessage.success('下载地址已复制');
  } catch (error) {
    resetCopyFeedback();
    ElMessage.error(error instanceof Error ? error.message : '复制下载地址失败');
  }
}

function handleArtifactError(error: unknown, fallbackMessage: string) {
  if (error instanceof ApiError) {
    switch (normalizeHttpStatus(error.code)) {
      case 401:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'auth';
        rememberArtifactError('登录状态已失效，请重新登录后再下载代码包');
        ElMessage.error('登录状态已失效，请重新登录后再下载代码包');
        return;
      case 404:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'notfound';
        rememberArtifactError('下载包不存在，可能已被清理，请重新执行“生成并下载”');
        ElMessage.error('下载包不存在，可能已被清理，请重新执行“生成并下载”');
        return;
      case 410:
        artifactForceExpired.value = true;
        operationStatus.value = '下载包已过期，请重新执行“生成并下载”。';
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'expired';
        rememberArtifactError('下载包已过期，请重新执行“生成并下载”');
        ElMessage.warning('下载包已过期，请重新执行“生成并下载”');
        return;
      case 500:
        lastRunSuccess.value = false;
        lastArtifactErrorType.value = 'server';
        rememberArtifactError('下载服务暂时不可用，请稍后重试');
        ElMessage.error('下载服务暂时不可用，请稍后重试');
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
    dsn: dbDsn.value.trim(),
    database: dbDatabase.value.trim(),
    schema: dbSchema.value.trim() || undefined,
    tables: tables.length > 0 ? tables : undefined,
    force: dbForce.value,
    generate_frontend: dbGenerateFrontend.value,
    generate_policy: dbGeneratePolicy.value,
  };
}

function validateDatabaseInputs(): string {
  if (!dbDriver.value.trim()) {
    return '请先选择数据库驱动';
  }
  if (!dbDsn.value.trim()) {
    return '请先填写 DSN';
  }
  if (!dbDatabase.value.trim()) {
    return '请先填写数据库名';
  }
  return '';
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
