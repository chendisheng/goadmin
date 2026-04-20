<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';

import AdminTable from '@/components/admin/AdminTable.vue';
import { formatDateTime } from '@/utils/admin';
import {
  canSubmitUploadForm,
  formatUploadFileSize,
  isPreviewableImage,
  resolveUploadStatusLabel,
  resolveUploadStatusTagType,
  resolveUploadVisibilityLabel,
  resolveUploadVisibilityTagType,
} from '@/utils/upload';
import {
  bindUploadFile,
  deleteUploadFile,
  downloadUploadFile,
  fetchUploadFiles,
  previewUploadFile,
  unbindUploadFile,
  uploadUploadFile,
} from '@/api/upload';
import type { UploadFileBindFormState, UploadFileFormState, UploadFileItem, UploadFileQuery } from '@/types/upload';

const tableLoading = ref(false);
const uploadLoading = ref(false);
const bindLoading = ref(false);
const previewLoading = ref(false);
const uploadDialogVisible = ref(false);
const bindDialogVisible = ref(false);
const previewDialogVisible = ref(false);
const uploadFormRef = ref<FormInstance>();
const bindFormRef = ref<FormInstance>();
const rows = ref<UploadFileItem[]>([]);
const total = ref(0);
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);
const previewItem = ref<UploadFileItem | null>(null);
const previewTargetId = ref('');
const bindTarget = ref<UploadFileItem | null>(null);

const query = reactive<UploadFileQuery>({
  keyword: '',
  visibility: '',
  status: '',
  biz_module: '',
  biz_type: '',
  biz_id: '',
  uploaded_by: '',
  page: 1,
  page_size: 10,
});

const defaultUploadForm = (): UploadFileFormState => ({
  visibility: 'private',
  biz_module: '',
  biz_type: '',
  biz_id: '',
  biz_field: '',
  remark: '',
});

const defaultBindForm = (): UploadFileBindFormState => ({
  biz_module: '',
  biz_type: '',
  biz_id: '',
  biz_field: '',
});

const uploadForm = reactive<UploadFileFormState>(defaultUploadForm());
const bindForm = reactive<UploadFileBindFormState>(defaultBindForm());

const uploadRules: FormRules<UploadFileFormState> = {
  visibility: [{ required: true, message: '请选择文件可见性', trigger: 'change' }],
};

const bindRules: FormRules<UploadFileBindFormState> = {
  biz_module: [{ required: true, message: '请输入业务模块', trigger: 'blur' }],
  biz_type: [{ required: true, message: '请输入业务类型', trigger: 'blur' }],
  biz_id: [{ required: true, message: '请输入业务ID', trigger: 'blur' }],
  biz_field: [{ required: true, message: '请输入业务字段', trigger: 'blur' }],
};

const visibilityOptions = [
  { value: 'private', label: '私有' },
  { value: 'public', label: '公开' },
];

const statusOptions = [
  { value: 'active', label: '有效' },
  { value: 'archived', label: '已归档' },
  { value: 'deleted', label: '已删除' },
];

const selectedFileLabel = computed(() => {
  if (!selectedFile.value) {
    return '未选择文件';
  }
  return `${selectedFile.value.name} · ${formatUploadFileSize(selectedFile.value.size)}`;
});

const uploadReady = computed(() => canSubmitUploadForm(selectedFile.value, uploadForm));

function resetUploadForm() {
  Object.assign(uploadForm, defaultUploadForm());
  selectedFile.value = null;
  if (fileInputRef.value) {
    fileInputRef.value.value = '';
  }
}

function resetBindForm() {
  Object.assign(bindForm, defaultBindForm());
  bindTarget.value = null;
}

async function loadFiles() {
  tableLoading.value = true;
  try {
    const response = await fetchUploadFiles({ ...query });
    rows.value = response.items ?? [];
    total.value = response.total ?? 0;
  } finally {
    tableLoading.value = false;
  }
}

function openUploadDialog() {
  resetUploadForm();
  uploadDialogVisible.value = true;
  void nextTick(() => {
    fileInputRef.value?.focus?.();
  });
}

function triggerFileSelect() {
  fileInputRef.value?.click();
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null;
  const file = input?.files?.[0] ?? null;
  selectedFile.value = file;
}

function openBindDialog(row: UploadFileItem) {
  bindTarget.value = row;
  Object.assign(bindForm, {
    ...defaultBindForm(),
    biz_module: row.biz_module ?? '',
    biz_type: row.biz_type ?? '',
    biz_id: row.biz_id ?? '',
    biz_field: row.biz_field ?? '',
  });
  bindDialogVisible.value = true;
}

async function submitUpload() {
  if (!selectedFile.value) {
    ElMessage.warning('请选择要上传的文件');
    return;
  }
  try {
    await uploadFormRef.value?.validate();
  } catch {
    return;
  }
  uploadLoading.value = true;
  try {
    await uploadUploadFile(selectedFile.value, {
      visibility: uploadForm.visibility.trim() || 'private',
      biz_module: uploadForm.biz_module.trim(),
      biz_type: uploadForm.biz_type.trim(),
      biz_id: uploadForm.biz_id.trim(),
      biz_field: uploadForm.biz_field.trim(),
      remark: uploadForm.remark.trim(),
    });
    ElMessage.success('文件已上传');
    uploadDialogVisible.value = false;
    resetUploadForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '上传失败');
  } finally {
    uploadLoading.value = false;
  }
}

async function submitBind() {
  if (!bindTarget.value) {
    ElMessage.warning('请选择要绑定的文件');
    return;
  }
  try {
    await bindFormRef.value?.validate();
  } catch {
    return;
  }
  bindLoading.value = true;
  try {
    await bindUploadFile(bindTarget.value.id, {
      biz_module: bindForm.biz_module.trim(),
      biz_type: bindForm.biz_type.trim(),
      biz_id: bindForm.biz_id.trim(),
      biz_field: bindForm.biz_field.trim(),
    });
    ElMessage.success('文件已绑定');
    bindDialogVisible.value = false;
    resetBindForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '绑定失败');
  } finally {
    bindLoading.value = false;
  }
}

async function openPreview(row: UploadFileItem) {
  previewLoading.value = true;
  previewTargetId.value = row.id;
  try {
    previewItem.value = await previewUploadFile(row.id);
    previewDialogVisible.value = true;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '预览失败');
  } finally {
    previewLoading.value = false;
    previewTargetId.value = '';
  }
}

async function copyPreviewUrl(url?: string) {
  if (!url) {
    return;
  }
  try {
    await navigator.clipboard.writeText(url);
    ElMessage.success('公开地址已复制');
  } catch {
    ElMessage.warning('复制失败，请手动复制');
  }
}

async function handleDownload(row: UploadFileItem) {
  try {
    await downloadUploadFile(row.id, row.original_name || row.storage_name || 'upload-file');
    ElMessage.success('文件已开始下载');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '下载失败');
  }
}

async function handleDelete(row: UploadFileItem) {
  await ElMessageBox.confirm(`确认删除文件 ${row.original_name || row.id} 吗？`, '删除文件', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteUploadFile(row.id);
  ElMessage.success('文件已删除');
  await loadFiles();
}

async function handleUnbind(row: UploadFileItem) {
  await ElMessageBox.confirm(`确认解除文件 ${row.original_name || row.id} 的绑定吗？`, '解除绑定', {
    type: 'warning',
    confirmButtonText: '解绑',
    cancelButtonText: '取消',
  });
  await unbindUploadFile(row.id);
  ElMessage.success('文件已解绑');
  await loadFiles();
}

function handleSearch() {
  query.page = 1;
  void loadFiles();
}

function handleReset() {
  query.keyword = '';
  query.visibility = '';
  query.status = '';
  query.biz_module = '';
  query.biz_type = '';
  query.biz_id = '';
  query.uploaded_by = '';
  query.page = 1;
  void loadFiles();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadFiles();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadFiles();
}

onMounted(() => {
  void loadFiles();
});

onActivated(() => {
  void loadFiles();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      title="文件管理"
      description="管理上传文件、绑定业务对象、下载文件与查看文件元数据。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadFiles">刷新</el-button>
        <el-button v-permission="'upload:file:create'" type="primary" @click="openUploadDialog">上传文件</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="文件名 / 存储键 / 备注" />
          </el-form-item>
          <el-form-item label="可见性">
            <el-select v-model="query.visibility" clearable placeholder="全部可见性" style="width: 180px">
              <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="query.status" clearable placeholder="全部状态" style="width: 180px">
              <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="业务模块">
            <el-input v-model="query.biz_module" clearable placeholder="biz_module" />
          </el-form-item>
          <el-form-item label="业务类型">
            <el-input v-model="query.biz_type" clearable placeholder="biz_type" />
          </el-form-item>
          <el-form-item label="业务ID">
            <el-input v-model="query.biz_id" clearable placeholder="biz_id" />
          </el-form-item>
          <el-form-item label="上传人">
            <el-input v-model="query.uploaded_by" clearable placeholder="uploaded_by" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="original_name" label="文件名" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.original_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="visibility" label="可见性" width="100">
          <template #default="{ row }">
            <el-tag :type="resolveUploadVisibilityTagType(row.visibility)" effect="plain">
              {{ resolveUploadVisibilityLabel(row.visibility) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="resolveUploadStatusTagType(row.status)" effect="plain">
              {{ resolveUploadStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="mime_type" label="MIME 类型" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.mime_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="extension" label="扩展名" width="110">
          <template #default="{ row }">
            {{ row.extension || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="size_bytes" label="大小" width="120">
          <template #default="{ row }">
            {{ formatUploadFileSize(row.size_bytes) }}
          </template>
        </el-table-column>
        <el-table-column prop="storage_driver" label="存储驱动" width="130">
          <template #default="{ row }">
            {{ row.storage_driver || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_module" label="业务模块" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_module || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_type" label="业务类型" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_id" label="业务ID" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="uploaded_by" label="上传人" width="140">
          <template #default="{ row }">
            {{ row.uploaded_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-space wrap :size="6">
              <el-button v-permission="'upload:file:preview'" link type="primary" :loading="previewLoading && previewTargetId === row.id" @click="openPreview(row)">预览</el-button>
              <el-button v-permission="'upload:file:download'" link type="success" @click="handleDownload(row)">下载</el-button>
              <el-button v-permission="'upload:file:bind'" link type="warning" @click="openBindDialog(row)">绑定</el-button>
              <el-button v-permission="'upload:file:unbind'" link @click="handleUnbind(row)">解绑</el-button>
              <el-button v-permission="'upload:file:delete'" link type="danger" @click="handleDelete(row)">删除</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <div class="admin-pagination">
          <el-pagination
            background
            layout="total, sizes, prev, pager, next, jumper"
            :total="total"
            :current-page="query.page"
            :page-size="query.page_size"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          />
        </div>
      </template>
    </AdminTable>

    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="760px" destroy-on-close>
      <el-alert
        title="支持在上传时填写业务绑定信息和备注；文件内容会按后端存储策略进行校验。"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="uploadFormRef" :model="uploadForm" :rules="uploadRules" label-width="110px" class="admin-form">
        <el-form-item label="文件" required>
          <el-space wrap>
            <el-button @click="triggerFileSelect">选择文件</el-button>
            <el-tag v-if="selectedFile" effect="plain">{{ selectedFileLabel }}</el-tag>
            <el-text v-else type="info">未选择文件</el-text>
          </el-space>
          <input ref="fileInputRef" type="file" class="hidden-file-input" @change="handleFileChange" />
        </el-form-item>
        <el-form-item label="可见性" prop="visibility">
          <el-select v-model="uploadForm.visibility" style="width: 100%" placeholder="请选择可见性">
            <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="业务模块">
          <el-input v-model="uploadForm.biz_module" placeholder="请输入业务模块" />
        </el-form-item>
        <el-form-item label="业务类型">
          <el-input v-model="uploadForm.biz_type" placeholder="请输入业务类型" />
        </el-form-item>
        <el-form-item label="业务ID">
          <el-input v-model="uploadForm.biz_id" placeholder="请输入业务ID" />
        </el-form-item>
        <el-form-item label="业务字段">
          <el-input v-model="uploadForm.biz_field" placeholder="请输入业务字段" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="uploadForm.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" :disabled="!uploadReady" @click="submitUpload">确认上传</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="bindDialogVisible" title="绑定文件" width="680px" destroy-on-close>
      <el-alert
        v-if="bindTarget"
        :title="`当前文件：${bindTarget.original_name || bindTarget.id}`"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="bindFormRef" :model="bindForm" :rules="bindRules" label-width="110px" class="admin-form">
        <el-form-item label="业务模块" prop="biz_module">
          <el-input v-model="bindForm.biz_module" placeholder="请输入业务模块" />
        </el-form-item>
        <el-form-item label="业务类型" prop="biz_type">
          <el-input v-model="bindForm.biz_type" placeholder="请输入业务类型" />
        </el-form-item>
        <el-form-item label="业务ID" prop="biz_id">
          <el-input v-model="bindForm.biz_id" placeholder="请输入业务ID" />
        </el-form-item>
        <el-form-item label="业务字段" prop="biz_field">
          <el-input v-model="bindForm.biz_field" placeholder="请输入业务字段" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bindDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="bindLoading" @click="submitBind">确认绑定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewDialogVisible" title="文件预览" width="840px" destroy-on-close>
      <template v-if="previewItem">
        <el-space wrap class="mb-4">
          <el-tag :type="resolveUploadVisibilityTagType(previewItem.visibility)" effect="plain">
            {{ resolveUploadVisibilityLabel(previewItem.visibility) }}
          </el-tag>
          <el-tag :type="resolveUploadStatusTagType(previewItem.status)" effect="plain">
            {{ resolveUploadStatusLabel(previewItem.status) }}
          </el-tag>
          <el-button v-if="previewItem.public_url" plain @click="copyPreviewUrl(previewItem.public_url)">复制公开地址</el-button>
          <el-link v-if="previewItem.public_url" :href="previewItem.public_url" target="_blank" type="primary">新窗口打开</el-link>
        </el-space>

        <el-alert
          v-if="!previewItem.public_url"
          title="当前文件没有公开访问地址，可通过下载按钮获取文件。"
          type="warning"
          :closable="false"
          show-icon
          class="mb-4"
        />

        <el-descriptions :column="2" border class="mb-4">
          <el-descriptions-item label="文件名">{{ previewItem.original_name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="可见性">{{ resolveUploadVisibilityLabel(previewItem.visibility) }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ resolveUploadStatusLabel(previewItem.status) }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ formatUploadFileSize(previewItem.size_bytes) }}</el-descriptions-item>
          <el-descriptions-item label="MIME 类型">{{ previewItem.mime_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="扩展名">{{ previewItem.extension || '-' }}</el-descriptions-item>
          <el-descriptions-item label="存储驱动">{{ previewItem.storage_driver || '-' }}</el-descriptions-item>
          <el-descriptions-item label="存储键">{{ previewItem.storage_key || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务模块">{{ previewItem.biz_module || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务类型">{{ previewItem.biz_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务ID">{{ previewItem.biz_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="业务字段">{{ previewItem.biz_field || '-' }}</el-descriptions-item>
          <el-descriptions-item label="上传人">{{ previewItem.uploaded_by || '-' }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDateTime(previewItem.updated_at) }}</el-descriptions-item>
          <el-descriptions-item label="公开地址" :span="2">
            <el-link v-if="previewItem.public_url" :href="previewItem.public_url" target="_blank" type="primary">{{ previewItem.public_url }}</el-link>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="previewItem.public_url && isPreviewableImage(previewItem.mime_type)" class="upload-preview-image-wrap">
          <el-image :src="previewItem.public_url" fit="contain" class="upload-preview-image" :preview-src-list="[previewItem.public_url]" />
        </div>
        <el-alert
          v-else-if="previewItem.public_url"
          title="当前文件类型不是图片，已在上方显示元数据，可使用下载按钮查看原文件。"
          type="info"
          :closable="false"
          show-icon
        />
      </template>
      <template #footer>
        <el-button @click="previewDialogVisible = false">关闭</el-button>
        <el-button v-if="previewItem" type="primary" @click="handleDownload(previewItem)">下载文件</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.hidden-file-input {
  display: none;
}

.mb-4 {
  margin-bottom: 16px;
}

.upload-preview-image-wrap {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}

.upload-preview-image {
  max-width: 100%;
  max-height: 420px;
  border-radius: 8px;
}
</style>
