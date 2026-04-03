<template>
  <div class="codegen-page">
    <el-row :gutter="20" class="codegen-grid">
      <el-col :xs="24" :lg="11" :xl="11">
        <el-card shadow="never" class="codegen-card">
          <template #header>
            <div class="card-header">
              <div>
                <div class="title">CodeGen Console</div>
                <div class="subtitle">上传 DSL、预览 dry-run，并支持服务端生成后下载代码包。</div>
              </div>
              <el-space wrap>
                <el-button @click="loadSample">载入示例</el-button>
                <el-button @click="clearDsl">清空</el-button>
                <el-button @click="triggerFileSelect">上传 DSL</el-button>
                <el-button type="primary" :loading="previewLoading" @click="handlePreview">Dry-run 预览</el-button>
                <el-button type="success" :loading="generateLoading" @click="handleGenerate">一键生成</el-button>
                <el-button type="warning" :loading="downloadLoading" @click="handleGenerateDownload">生成并下载</el-button>
              </el-space>
            </div>
          </template>

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

          <el-card shadow="never" class="codegen-card">
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
  generateCodegenDsl,
  generateDownloadCodegenDsl,
  previewCodegenDsl,
  type CodegenArtifactInfo,
  type CodegenDslExecutionReport,
  type CodegenDslPreviewItem,
} from '@/api/codegen';
import { ApiError } from '@/api/types';

const dslText = ref('');
const force = ref(false);
const packageName = ref('');
const includeReadme = ref(true);
const includeReport = ref(true);
const includeDsl = ref(true);
const previewLoading = ref(false);
const generateLoading = ref(false);
const downloadLoading = ref(false);
const report = ref<CodegenDslExecutionReport | null>(null);
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

const previewItems = computed<CodegenDslPreviewItem[]>(() => report.value?.items ?? []);
const messages = computed(() => report.value?.messages ?? []);
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

function clearDsl() {
  dslText.value = '';
  packageName.value = '';
  report.value = null;
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
  if (!dslText.value.trim()) {
    ElMessage.warning('请先填写 DSL 内容');
    return;
  }
  previewLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    report.value = await previewCodegenDsl({ dsl: dslText.value, force: force.value });
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
  if (!dslText.value.trim()) {
    ElMessage.warning('请先填写 DSL 内容');
    return;
  }
  generateLoading.value = true;
  operationStatus.value = '';
  lastRunSuccess.value = false;
  try {
    report.value = await generateCodegenDsl({ dsl: dslText.value, force: force.value });
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
