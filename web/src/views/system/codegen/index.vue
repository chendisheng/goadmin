<template>
  <div class="codegen-page">
    <el-row :gutter="20" class="codegen-grid">
      <el-col :xs="24" :lg="11" :xl="11">
        <el-card shadow="never" class="codegen-card">
          <template #header>
            <div class="card-header">
              <div>
                <div class="title">CodeGen Console</div>
                <div class="subtitle">上传 DSL、预览 dry-run，并一键执行生成。</div>
              </div>
              <el-space wrap>
                <el-button @click="loadSample">载入示例</el-button>
                <el-button @click="clearDsl">清空</el-button>
                <el-button @click="triggerFileSelect">上传 DSL</el-button>
                <el-button type="primary" :loading="previewLoading" @click="handlePreview">Dry-run 预览</el-button>
                <el-button type="success" :loading="generateLoading" @click="handleGenerate">一键生成</el-button>
              </el-space>
            </div>
          </template>

          <el-form label-position="top" class="codegen-form">
            <el-form-item label="Force overwrite">
              <el-switch v-model="force" inline-prompt active-text="On" inactive-text="Off" />
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
        </el-space>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';

import { generateCodegenDsl, previewCodegenDsl, type CodegenDslExecutionReport, type CodegenDslPreviewItem } from '@/api/codegen';

const dslText = ref('');
const force = ref(false);
const previewLoading = ref(false);
const generateLoading = ref(false);
const report = ref<CodegenDslExecutionReport | null>(null);
const lastRunSuccess = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

const previewItems = computed<CodegenDslPreviewItem[]>(() => report.value?.items ?? []);
const messages = computed(() => report.value?.messages ?? []);
const statusMessage = computed(() => {
  const lines = messages.value;
  if (!lines.length) {
    return '';
  }
  return lines.join(' · ');
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
  report.value = null;
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
  lastRunSuccess.value = false;
  try {
    report.value = await previewCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
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
  lastRunSuccess.value = false;
  try {
    report.value = await generateCodegenDsl({ dsl: dslText.value, force: force.value });
    lastRunSuccess.value = true;
    ElMessage.success('生成已完成');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '生成失败');
  } finally {
    generateLoading.value = false;
  }
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
</style>
