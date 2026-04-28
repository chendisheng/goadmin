<script setup lang="ts">
import { computed, nextTick, onActivated, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';

import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { formatDateTime } from '@/utils/admin';
import {
  canSubmitUploadForm,
  isBrowserDirectPublicUrl,
  formatUploadFileSize,
  isPreviewableImage,
  resolveUploadPreviewKind,
  resolveUploadStatusTagType,
  resolveUploadVisibilityTagType,
} from '@/utils/upload';
import {
  bindUploadFile,
  createUploadFilePreviewUrl,
  deleteUploadFile,
  downloadUploadFile,
  fetchUploadFiles,
  fetchUploadStorageSetting,
  previewUploadFile,
  unbindUploadFile,
  updateUploadStorageSetting,
  uploadUploadFile,
} from '@/api/upload';
import type { UploadFileBindFormState, UploadFileFormState, UploadFileItem, UploadFilePreviewItem, UploadFileQuery, UploadStorageSettingFormState } from '@/types/upload';

const { t } = useAppI18n();

const tableLoading = ref(false);
const storageSettingLoading = ref(false);
const storageSettingSaving = ref(false);
const uploadLoading = ref(false);
const bindLoading = ref(false);
const previewLoading = ref(false);
const uploadDialogVisible = ref(false);
const bindDialogVisible = ref(false);
const previewDialogVisible = ref(false);
const previewBrowserUrl = ref('');
const previewBrowserUrlIsObjectUrl = ref(false);
const uploadFormRef = ref<FormInstance>();
const bindFormRef = ref<FormInstance>();
const rows = ref<UploadFileItem[]>([]);
const total = ref(0);
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);
const previewItem = ref<UploadFilePreviewItem | null>(null);
const previewTargetId = ref('');
const bindTarget = ref<UploadFileItem | null>(null);

const storageDriverOptions = computed(() => [
  { value: 'local', label: t('upload.storage.local', 'Local storage') },
  { value: 'db', label: t('upload.storage.db', 'Database storage') },
  { value: 's3-compatible', label: t('upload.storage.s3', 'S3 compatible') },
  { value: 'oss', label: t('upload.storage.oss', 'Alibaba Cloud OSS') },
  { value: 'cos', label: t('upload.storage.cos', 'Tencent Cloud COS') },
  { value: 'qiniu', label: t('upload.storage.qiniu', 'Qiniu Cloud') },
  { value: 'minio', label: t('upload.storage.minio', 'MinIO') },
]);

const normalizeStorageDriver = (driver: string | null | undefined): string => {
  const normalized = (driver ?? '').trim().toLowerCase();
  switch (normalized) {
    case 'database':
      return 'db';
    case 'local':
    case 'db':
    case 's3-compatible':
    case 'oss':
    case 'cos':
    case 'qiniu':
    case 'minio':
      return normalized;
    default:
      return 'local';
  }
};

const defaultStorageSettingForm = (): UploadStorageSettingFormState => ({
  driver: 'local',
});

const storageSetting = reactive<UploadStorageSettingFormState>(defaultStorageSettingForm());

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
  visibility: [{ required: true, message: t('upload.validation.visibility_required', 'Select file visibility'), trigger: 'change' }],
};

const bindRules: FormRules<UploadFileBindFormState> = {
  biz_module: [{ required: true, message: t('upload.validation.biz_module_required', 'Enter business module'), trigger: 'blur' }],
  biz_type: [{ required: true, message: t('upload.validation.biz_type_required', 'Enter business type'), trigger: 'blur' }],
  biz_id: [{ required: true, message: t('upload.validation.biz_id_required', 'Enter business ID'), trigger: 'blur' }],
  biz_field: [{ required: true, message: t('upload.validation.biz_field_required', 'Enter business field'), trigger: 'blur' }],
};

const visibilityOptions = computed(() => [
  { value: 'private', label: t('upload.visibility.private', 'Private') },
  { value: 'public', label: t('upload.visibility.public', 'Public') },
]);

const statusOptions = computed(() => [
  { value: 'active', label: t('upload.status.active', 'Active') },
  { value: 'archived', label: t('upload.status.archived', 'Archived') },
  { value: 'deleted', label: t('upload.status.deleted', 'Deleted') },
]);

const selectedFileLabel = computed(() => {
  if (!selectedFile.value) {
    return t('upload.no_file_selected', 'No file selected');
  }
  return `${selectedFile.value.name} · ${formatUploadFileSize(selectedFile.value.size)}`;
});

const previewKind = computed(() => resolveUploadPreviewKind(previewItem.value?.mime_type));

const previewMode = computed(() => previewItem.value?.preview_mode || 'download_only');

const canDirectPreview = computed(() => {
  const item = previewItem.value;
  if (!item) {
    return false;
  }
  return item.visibility === 'public' && isBrowserDirectPublicUrl(item.public_url) && previewMode.value === 'public_url';
});

const previewTitle = computed(() => {
  if (!previewItem.value) {
    return t('upload.preview_title', 'File preview');
  }
  const kindLabel = previewKind.value === 'image'
    ? t('upload.preview.image', 'Image preview')
    : previewKind.value === 'pdf'
      ? t('upload.preview.pdf', 'PDF preview')
      : previewKind.value === 'text'
        ? t('upload.preview.text', 'Text preview')
        : t('upload.preview.download_only', 'Download only');
  return `${previewItem.value.original_name || t('upload.preview_title', 'File preview')} · ${kindLabel}`;
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

function resolveUploadVisibilityLabel(value?: string): string {
  if (!value) {
    return '-';
  }
  switch (value) {
    case 'public':
      return t('upload.visibility.public', 'Public');
    case 'private':
      return t('upload.visibility.private', 'Private');
    default:
      return value;
  }
}

function resolveUploadStatusLabel(value?: string): string {
  if (!value) {
    return '-';
  }
  switch (value) {
    case 'active':
      return t('upload.status.active', 'Active');
    case 'archived':
      return t('upload.status.archived', 'Archived');
    case 'deleted':
      return t('upload.status.deleted', 'Deleted');
    default:
      return value;
  }
}

function revokePreviewBrowserUrl() {
  if (previewBrowserUrl.value && previewBrowserUrlIsObjectUrl.value) {
    window.URL.revokeObjectURL(previewBrowserUrl.value);
  }
  previewBrowserUrl.value = '';
  previewBrowserUrlIsObjectUrl.value = false;
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

async function loadStorageSetting() {
  storageSettingLoading.value = true;
  try {
    const response = await fetchUploadStorageSetting();
    storageSetting.driver = normalizeStorageDriver(response.driver);
  } catch {
    storageSetting.driver = 'local';
  } finally {
    storageSettingLoading.value = false;
  }
}

async function submitStorageSetting() {
  storageSettingSaving.value = true;
  try {
    const driver = normalizeStorageDriver(storageSetting.driver);
    const response = await updateUploadStorageSetting({ driver });
    storageSetting.driver = normalizeStorageDriver(response.driver || driver);
    ElMessage.success(t('upload.storage.saved', 'Default storage driver saved'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('common.failure', 'Save failed'));
  } finally {
    storageSettingSaving.value = false;
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
    ElMessage.warning(t('upload.choose_file_first', 'Select a file to upload first'));
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
      visibility: uploadForm.visibility,
      biz_module: uploadForm.biz_module.trim(),
      biz_type: uploadForm.biz_type.trim(),
      biz_id: uploadForm.biz_id.trim(),
      biz_field: uploadForm.biz_field.trim(),
      remark: uploadForm.remark.trim(),
    });
    ElMessage.success(t('upload.uploaded', 'File uploaded'));
    uploadDialogVisible.value = false;
    resetUploadForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.upload_failed', 'Upload failed'));
  } finally {
    uploadLoading.value = false;
  }
}

async function submitBind() {
  if (!bindTarget.value) {
    ElMessage.warning(t('upload.choose_bind_target', 'Select a file to bind'));
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
    ElMessage.success(t('upload.bound', 'File bound'));
    bindDialogVisible.value = false;
    resetBindForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.bind_failed', 'Bind failed'));
  } finally {
    bindLoading.value = false;
  }
}

async function openPreview(row: UploadFileItem) {
  previewLoading.value = true;
  previewTargetId.value = row.id;
  try {
    revokePreviewBrowserUrl();
    const item = await previewUploadFile(row.id);
    previewItem.value = item;
    const publicUrl = item.public_url ?? '';
    if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
      previewBrowserUrl.value = '';
      previewBrowserUrlIsObjectUrl.value = false;
    } else if (previewMode.value === 'public_url' && publicUrl && isBrowserDirectPublicUrl(publicUrl)) {
      previewBrowserUrl.value = publicUrl;
      previewBrowserUrlIsObjectUrl.value = false;
    } else if (item.download_url) {
      previewBrowserUrl.value = await createUploadFilePreviewUrl(row.id);
      previewBrowserUrlIsObjectUrl.value = true;
    }
    previewDialogVisible.value = true;
  } catch (error) {
    revokePreviewBrowserUrl();
    ElMessage.error(error instanceof Error ? error.message : t('upload.preview_failed', 'Preview failed'));
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
    ElMessage.success(t('upload.preview.copied', 'Public URL copied'));
  } catch {
    ElMessage.warning(t('upload.preview.copy_failed', 'Copy failed, please copy it manually'));
  }
}

async function openPreviewWindow() {
  try {
    const previewUrl = previewBrowserUrl.value;
    if (!previewUrl) {
      ElMessage.warning(t('upload.preview.no_url', 'No preview URL available'));
      return;
    }
    const opened = window.open(previewUrl, '_blank', 'noopener,noreferrer');
    if (!opened) {
      ElMessage.warning(t('upload.preview.blocked', 'The browser blocked the new window; allow popups and try again'));
      return;
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.preview.open_failed', 'Failed to open preview'));
  }
}

function getPreviewSourceLabel() {
  if (!previewItem.value) {
    return '-';
  }
  if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
    return t('upload.preview.download_only', 'Download only');
  }
  if (previewMode.value === 'public_url' && isBrowserDirectPublicUrl(previewItem.value.public_url)) {
    return t('upload.preview.public_direct', 'Public direct link');
  }
  return t('upload.preview.auth_download', 'Authenticated download');
}

async function handleDownload(row: UploadFileItem) {
  try {
    await downloadUploadFile(row.id, row.original_name || row.storage_name || 'upload-file');
    ElMessage.success(t('upload.download_started', 'Download started'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.download_failed', 'Download failed'));
  }
}

async function handleDelete(row: UploadFileItem) {
  await ElMessageBox.confirm(t('upload.confirm_delete', 'Delete file {name}?', { name: row.original_name || row.id }), t('upload.delete_title', 'Delete file'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteUploadFile(row.id);
  ElMessage.success(t('upload.deleted', 'File deleted'));
  await loadFiles();
}

async function handleUnbind(row: UploadFileItem) {
  await ElMessageBox.confirm(t('upload.confirm_unbind', 'Unbind file {name}?', { name: row.original_name || row.id }), t('upload.unbind_title', 'Unbind'), {
    type: 'warning',
    confirmButtonText: t('upload.unbind', 'Unbind'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await unbindUploadFile(row.id);
  ElMessage.success(t('upload.unbound', 'File unbound'));
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
  void loadStorageSetting();
  void loadFiles();
});

onActivated(() => {
  void loadStorageSetting();
  void loadFiles();
});
</script>

<template>
  <div class="admin-page">
    <el-card class="mb-4" shadow="never">
      <div class="upload-setting-card">
        <div class="upload-setting-content">
          <div class="upload-setting-title">{{ t('upload.storage.title', 'Default storage driver') }}</div>
          <el-text type="info">{{ t('upload.storage.description', 'New uploads will use the selected storage implementation by default. The setting is persisted to the database; database storage stores file content and metadata together.') }}</el-text>
        </div>
        <el-space wrap>
          <el-select v-model="storageSetting.driver" :loading="storageSettingLoading" :disabled="storageSettingSaving" style="width: 220px">
            <el-option v-for="option in storageDriverOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-button type="primary" :loading="storageSettingSaving" @click="submitStorageSetting">{{ t('upload.storage.save', 'Save settings') }}</el-button>
        </el-space>
      </div>
    </el-card>

    <AdminTable
      :title="t('upload.title', 'File management')"
      :description="t('upload.description', 'Manage uploaded files, bind business objects, download files, and inspect file metadata.')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadFiles">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'upload:file:create'" type="primary" @click="openUploadDialog">{{ t('upload.upload_file', 'Upload file') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('upload.keyword', 'Keyword')">
            <el-input v-model="query.keyword" clearable :placeholder="t('upload.keyword_placeholder', 'File name / storage key / remark')" />
          </el-form-item>
          <el-form-item :label="t('upload.visibility.label', 'Visibility')">
            <el-select v-model="query.visibility" clearable :placeholder="t('upload.all_visibility', 'All visibility')" style="width: 180px">
              <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('upload.status.label', 'Status')">
            <el-select v-model="query.status" clearable :placeholder="t('upload.all_status', 'All statuses')" style="width: 180px">
              <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('upload.biz_module', 'Business module')">
            <el-input v-model="query.biz_module" clearable :placeholder="t('upload.biz_module_placeholder', 'biz_module')" />
          </el-form-item>
          <el-form-item :label="t('upload.biz_type', 'Business type')">
            <el-input v-model="query.biz_type" clearable :placeholder="t('upload.biz_type_placeholder', 'biz_type')" />
          </el-form-item>
          <el-form-item :label="t('upload.biz_id', 'Business ID')">
            <el-input v-model="query.biz_id" clearable :placeholder="t('upload.biz_id_placeholder', 'biz_id')" />
          </el-form-item>
          <el-form-item :label="t('upload.uploaded_by', 'Uploaded by')">
            <el-input v-model="query.uploaded_by" clearable :placeholder="t('upload.uploaded_by_placeholder', 'uploaded_by')" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="original_name" :label="t('upload.file_name', 'File name')" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.original_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="visibility" :label="t('upload.visibility.label', 'Visibility')" width="100">
          <template #default="{ row }">
            <el-tag :type="resolveUploadVisibilityTagType(row.visibility)" effect="plain">
              {{ resolveUploadVisibilityLabel(row.visibility) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('upload.status.label', 'Status')" width="110">
          <template #default="{ row }">
            <el-tag :type="resolveUploadStatusTagType(row.status)" effect="plain">
              {{ resolveUploadStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="mime_type" :label="t('upload.mime_type', 'MIME type')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.mime_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="extension" :label="t('upload.extension', 'Extension')" width="110">
          <template #default="{ row }">
            {{ row.extension || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="size_bytes" :label="t('upload.size', 'Size')" width="120">
          <template #default="{ row }">
            {{ formatUploadFileSize(row.size_bytes) }}
          </template>
        </el-table-column>
        <el-table-column prop="storage_driver" :label="t('upload.storage_driver', 'Storage driver')" width="130">
          <template #default="{ row }">
            {{ row.storage_driver || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_module" :label="t('upload.biz_module', 'Business module')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_module || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_type" :label="t('upload.biz_type', 'Business type')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_id" :label="t('upload.biz_id', 'Business ID')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="uploaded_by" :label="t('upload.uploaded_by', 'Uploaded by')" width="140">
          <template #default="{ row }">
            {{ row.uploaded_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('upload.updated_at', 'Updated at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', 'Actions')" width="300" fixed="right">
          <template #default="{ row }">
            <el-space wrap :size="6">
              <el-button v-permission="'upload:file:preview'" link type="primary" :loading="previewLoading && previewTargetId === row.id" @click="openPreview(row)">{{ t('upload.preview', 'Preview') }}</el-button>
              <el-button v-permission="'upload:file:download'" link type="success" @click="handleDownload(row)">{{ t('upload.download', 'Download') }}</el-button>
              <el-button v-permission="'upload:file:bind'" link type="warning" @click="openBindDialog(row)">{{ t('upload.bind', 'Bind') }}</el-button>
              <el-button v-permission="'upload:file:unbind'" link @click="handleUnbind(row)">{{ t('upload.unbind', 'Unbind') }}</el-button>
              <el-button v-permission="'upload:file:delete'" link type="danger" @click="handleDelete(row)">{{ t('common.delete', 'Delete') }}</el-button>
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

    <el-dialog v-model="uploadDialogVisible" :title="t('upload.upload_title', 'Upload file')" width="760px" destroy-on-close>
      <el-alert
        :title="t('upload.upload_alert', 'You can fill in business binding info and a remark during upload; file content is validated according to the server storage policy.')"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="uploadFormRef" :model="uploadForm" :rules="uploadRules" label-width="110px" class="admin-form">
        <el-form-item :label="t('upload.file', 'File')" required>
          <el-space wrap>
            <el-button @click="triggerFileSelect">{{ t('upload.choose_file', 'Choose file') }}</el-button>
            <el-tag v-if="selectedFile" effect="plain">{{ selectedFileLabel }}</el-tag>
            <el-text v-else type="info">{{ t('upload.no_file_selected', 'No file selected') }}</el-text>
          </el-space>
          <input ref="fileInputRef" type="file" class="hidden-file-input" @change="handleFileChange" />
        </el-form-item>
        <el-form-item :label="t('upload.visibility.label', 'Visibility')" prop="visibility">
          <el-select v-model="uploadForm.visibility" style="width: 100%" :placeholder="t('upload.visibility.placeholder', 'Select visibility')">
            <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('upload.biz_module', 'Business module')">
          <el-input v-model="uploadForm.biz_module" :placeholder="t('upload.biz_module_placeholder_input', 'Enter business module')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_type', 'Business type')">
          <el-input v-model="uploadForm.biz_type" :placeholder="t('upload.biz_type_placeholder_input', 'Enter business type')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_id', 'Business ID')">
          <el-input v-model="uploadForm.biz_id" :placeholder="t('upload.biz_id_placeholder_input', 'Enter business ID')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_field', 'Business field')">
          <el-input v-model="uploadForm.biz_field" :placeholder="t('upload.biz_field_placeholder', 'Enter business field')" />
        </el-form-item>
        <el-form-item :label="t('upload.remark', 'Remark')">
          <el-input v-model="uploadForm.remark" type="textarea" :rows="3" :placeholder="t('upload.remark_placeholder', 'Enter a remark')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">{{ t('common.cancel', 'Cancel') }}</el-button>
        <el-button type="primary" :loading="uploadLoading" :disabled="!uploadReady" @click="submitUpload">{{ t('upload.confirm_upload', 'Confirm upload') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="bindDialogVisible" :title="t('upload.bind_title', 'Bind file')" width="680px" destroy-on-close>
      <el-alert
        v-if="bindTarget"
        :title="t('upload.current_file', 'Current file: {name}', { name: bindTarget.original_name || bindTarget.id })"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="bindFormRef" :model="bindForm" :rules="bindRules" label-width="110px" class="admin-form">
        <el-form-item :label="t('upload.biz_module', 'Business module')" prop="biz_module">
          <el-input v-model="bindForm.biz_module" :placeholder="t('upload.biz_module_placeholder_input', 'Enter business module')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_type', 'Business type')" prop="biz_type">
          <el-input v-model="bindForm.biz_type" :placeholder="t('upload.biz_type_placeholder_input', 'Enter business type')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_id', 'Business ID')" prop="biz_id">
          <el-input v-model="bindForm.biz_id" :placeholder="t('upload.biz_id_placeholder_input', 'Enter business ID')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_field', 'Business field')" prop="biz_field">
          <el-input v-model="bindForm.biz_field" :placeholder="t('upload.biz_field_placeholder', 'Enter business field')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bindDialogVisible = false">{{ t('common.cancel', 'Cancel') }}</el-button>
        <el-button type="primary" :loading="bindLoading" @click="submitBind">{{ t('upload.confirm_bind', 'Confirm bind') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewDialogVisible" :title="previewTitle" width="840px" destroy-on-close @closed="revokePreviewBrowserUrl">
      <template v-if="previewItem">
        <el-space wrap class="mb-4">
          <el-tag :type="resolveUploadVisibilityTagType(previewItem.visibility)" effect="plain">
            {{ resolveUploadVisibilityLabel(previewItem.visibility) }}
          </el-tag>
          <el-tag :type="resolveUploadStatusTagType(previewItem.status)" effect="plain">
            {{ resolveUploadStatusLabel(previewItem.status) }}
          </el-tag>
          <el-link v-if="isBrowserDirectPublicUrl(previewItem.public_url)" plain @click="copyPreviewUrl(previewItem.public_url)">{{ t('upload.preview.copy_public_url', 'Copy public URL') }}</el-link>
        </el-space>

        <el-alert
          v-if="previewKind === 'download-only'"
          :title="t('upload.preview.download_hint', 'This file type is not suitable for online preview. Use the download button to get the original file.')"
          class="mb-4"
        />

        <el-descriptions :column="2" border class="mb-4">
          <el-descriptions-item :label="t('upload.file_name', 'File name')">{{ previewItem.original_name || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.visibility.label', 'Visibility')">{{ resolveUploadVisibilityLabel(previewItem.visibility) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.status.label', 'Status')">{{ resolveUploadStatusLabel(previewItem.status) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.size', 'Size')">{{ formatUploadFileSize(previewItem.size_bytes) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.mime_type', 'MIME type')">{{ previewItem.mime_type || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.extension', 'Extension')">{{ previewItem.extension || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.storage_driver', 'Storage driver')">{{ previewItem.storage_driver || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.storage_key', 'Storage key')">{{ previewItem.storage_key || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_module', 'Business module')">{{ previewItem.biz_module || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_type', 'Business type')">{{ previewItem.biz_type || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_id', 'Business ID')">{{ previewItem.biz_id || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_field', 'Business field')">{{ previewItem.biz_field || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.uploaded_by', 'Uploaded by')">{{ previewItem.uploaded_by || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.updated_at', 'Updated at')">{{ formatDateTime(previewItem.updated_at) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.access_mode', 'Access mode')">{{ getPreviewSourceLabel() }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.public_url', 'Public URL')" :span="2">
            <el-link v-if="previewBrowserUrl && previewKind !== 'download-only'" type="primary" @click="openPreviewWindow">{{ t('upload.open_in_new_window', 'Open in new window') }}</el-link>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="previewBrowserUrl && isPreviewableImage(previewItem.mime_type)" class="upload-preview-image-wrap">
          <el-image :src="previewBrowserUrl" fit="contain" class="upload-preview-image" :preview-src-list="[previewBrowserUrl]" />
        </div>
        <div v-else-if="previewBrowserUrl && previewKind === 'pdf'" class="upload-preview-document-wrap">
          <iframe :src="previewBrowserUrl" class="upload-preview-document" :title="t('upload.preview.iframe_file', 'File preview')" />
        </div>
        <div v-else-if="previewBrowserUrl && previewKind === 'text'" class="upload-preview-text-wrap">
          <iframe :src="previewBrowserUrl" class="upload-preview-text" :title="t('upload.preview.iframe_text', 'Text preview')" />
        </div>
        <el-alert
          v-else-if="previewBrowserUrl"
          :title="t('upload.preview.temp_url_hint', 'This file is using a temporary preview URL; its metadata is shown above.')"
          type="info"
          :closable="false"
          show-icon
        />
      </template>
      <template #footer>
        <el-button @click="previewDialogVisible = false">{{ t('common.close', 'Close') }}</el-button>
        <el-button v-if="previewItem" type="primary" @click="handleDownload(previewItem)">{{ t('upload.download_file', 'Download file') }}</el-button>
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

.upload-setting-card {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.upload-setting-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.upload-setting-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
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

.upload-preview-document-wrap {
  margin-top: 16px;
  min-height: 480px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
}

.upload-preview-document {
  width: 100%;
  height: 480px;
  border: 0;
}

.upload-preview-text-wrap {
  margin-top: 16px;
  min-height: 360px;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
}

.upload-preview-text {
  width: 100%;
  height: 360px;
  border: 0;
  background: #fff;
}
</style>
