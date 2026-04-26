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
  { value: 'local', label: t('upload.storage.local', '本地存储') },
  { value: 'db', label: t('upload.storage.db', '数据库存储') },
  { value: 's3-compatible', label: t('upload.storage.s3', 'S3 兼容') },
  { value: 'oss', label: t('upload.storage.oss', '阿里云 OSS') },
  { value: 'cos', label: t('upload.storage.cos', '腾讯云 COS') },
  { value: 'qiniu', label: t('upload.storage.qiniu', '七牛云') },
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
  visibility: [{ required: true, message: t('upload.validation.visibility_required', '请选择文件可见性'), trigger: 'change' }],
};

const bindRules: FormRules<UploadFileBindFormState> = {
  biz_module: [{ required: true, message: t('upload.validation.biz_module_required', '请输入业务模块'), trigger: 'blur' }],
  biz_type: [{ required: true, message: t('upload.validation.biz_type_required', '请输入业务类型'), trigger: 'blur' }],
  biz_id: [{ required: true, message: t('upload.validation.biz_id_required', '请输入业务ID'), trigger: 'blur' }],
  biz_field: [{ required: true, message: t('upload.validation.biz_field_required', '请输入业务字段'), trigger: 'blur' }],
};

const visibilityOptions = computed(() => [
  { value: 'private', label: t('upload.visibility.private', '私有') },
  { value: 'public', label: t('upload.visibility.public', '公开') },
]);

const statusOptions = computed(() => [
  { value: 'active', label: t('upload.status.active', '有效') },
  { value: 'archived', label: t('upload.status.archived', '已归档') },
  { value: 'deleted', label: t('upload.status.deleted', '已删除') },
]);

const selectedFileLabel = computed(() => {
  if (!selectedFile.value) {
    return t('upload.no_file_selected', '未选择文件');
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
    return t('upload.preview_title', '文件预览');
  }
  const kindLabel = previewKind.value === 'image'
    ? t('upload.preview.image', '图片预览')
    : previewKind.value === 'pdf'
      ? t('upload.preview.pdf', 'PDF 预览')
      : previewKind.value === 'text'
        ? t('upload.preview.text', '文本预览')
        : t('upload.preview.download_only', '仅下载');
  return `${previewItem.value.original_name || t('upload.preview_title', '文件预览')} · ${kindLabel}`;
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
      return t('upload.visibility.public', '公开');
    case 'private':
      return t('upload.visibility.private', '私有');
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
      return t('upload.status.active', '有效');
    case 'archived':
      return t('upload.status.archived', '已归档');
    case 'deleted':
      return t('upload.status.deleted', '已删除');
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
    ElMessage.success(t('upload.storage.saved', '默认存储驱动已保存'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('common.failure', '保存失败'));
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
    ElMessage.warning(t('upload.choose_file_first', '请选择要上传的文件'));
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
    ElMessage.success(t('upload.uploaded', '文件已上传'));
    uploadDialogVisible.value = false;
    resetUploadForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.upload_failed', '上传失败'));
  } finally {
    uploadLoading.value = false;
  }
}

async function submitBind() {
  if (!bindTarget.value) {
    ElMessage.warning(t('upload.choose_bind_target', '请选择要绑定的文件'));
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
    ElMessage.success(t('upload.bound', '文件已绑定'));
    bindDialogVisible.value = false;
    resetBindForm();
    await loadFiles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.bind_failed', '绑定失败'));
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
    const publicUrl = item.public_url ?? '';
    if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
      previewBrowserUrl.value = '';
      previewBrowserUrlIsObjectUrl.value = false;
    } else if (previewMode.value === 'public_url' && publicUrl && isBrowserDirectPublicUrl(publicUrl)) {
      previewBrowserUrl.value = publicUrl;
      previewBrowserUrlIsObjectUrl.value = false;
    } else if (previewItem.value.download_url) {
      previewBrowserUrl.value = await createUploadFilePreviewUrl(row.id);
      previewBrowserUrlIsObjectUrl.value = true;
    }
    previewDialogVisible.value = true;
  } catch (error) {
    revokePreviewBrowserUrl();
    ElMessage.error(error instanceof Error ? error.message : t('upload.preview_failed', '预览失败'));
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
    ElMessage.success(t('upload.preview.copied', '公开地址已复制'));
  } catch {
    ElMessage.warning(t('upload.preview.copy_failed', '复制失败，请手动复制'));
  }
}

async function openPreviewWindow() {
  try {
    const previewUrl = previewBrowserUrl.value;
    if (!previewUrl) {
      ElMessage.warning(t('upload.preview.no_url', '暂无可用的预览地址'));
      return;
    }
    const opened = window.open(previewUrl, '_blank', 'noopener,noreferrer');
    if (!opened) {
      ElMessage.warning(t('upload.preview.blocked', '浏览器拦截了新窗口，请允许弹窗后重试'));
      return;
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.preview.open_failed', '打开预览失败'));
  }
}

function getPreviewSourceLabel() {
  if (!previewItem.value) {
    return '-';
  }
  if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
    return t('upload.preview.download_only', '仅下载');
  }
  if (previewMode.value === 'public_url' && isBrowserDirectPublicUrl(previewItem.value.public_url)) {
    return t('upload.preview.public_direct', '公开直连');
  }
  return t('upload.preview.auth_download', '鉴权下载');
}

async function handleDownload(row: UploadFileItem) {
  try {
    await downloadUploadFile(row.id, row.original_name || row.storage_name || 'upload-file');
    ElMessage.success(t('upload.download_started', '文件已开始下载'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('upload.download_failed', '下载失败'));
  }
}

async function handleDelete(row: UploadFileItem) {
  await ElMessageBox.confirm(t('upload.confirm_delete', '确认删除文件 {name} 吗？', { name: row.original_name || row.id }), t('upload.delete_title', '删除文件'), {
    type: 'warning',
    confirmButtonText: t('common.delete', '删除'),
    cancelButtonText: t('common.cancel', '取消'),
  });
  await deleteUploadFile(row.id);
  ElMessage.success(t('upload.deleted', '文件已删除'));
  await loadFiles();
}

async function handleUnbind(row: UploadFileItem) {
  await ElMessageBox.confirm(t('upload.confirm_unbind', '确认解除文件 {name} 的绑定吗？', { name: row.original_name || row.id }), t('upload.unbind_title', '解除绑定'), {
    type: 'warning',
    confirmButtonText: t('upload.unbind', '解绑'),
    cancelButtonText: t('common.cancel', '取消'),
  });
  await unbindUploadFile(row.id);
  ElMessage.success(t('upload.unbound', '文件已解绑'));
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
          <div class="upload-setting-title">{{ t('upload.storage.title', '默认存储驱动') }}</div>
          <el-text type="info">{{ t('upload.storage.description', '新上传文件将默认使用当前选择的存储实现，设置会持久化到数据库；数据库存储会把文件内容与元数据一并写入数据库。') }}</el-text>
        </div>
        <el-space wrap>
          <el-select v-model="storageSetting.driver" :loading="storageSettingLoading" :disabled="storageSettingSaving" style="width: 220px">
            <el-option v-for="option in storageDriverOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
          <el-button type="primary" :loading="storageSettingSaving" @click="submitStorageSetting">{{ t('upload.storage.save', '保存设置') }}</el-button>
        </el-space>
      </div>
    </el-card>

    <AdminTable
      :title="t('upload.title', '文件管理')"
      :description="t('upload.description', '管理上传文件、绑定业务对象、下载文件与查看文件元数据。')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadFiles">{{ t('common.refresh', '刷新') }}</el-button>
        <el-button v-permission="'upload:file:create'" type="primary" @click="openUploadDialog">{{ t('upload.upload_file', '上传文件') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('upload.keyword', '关键字')">
            <el-input v-model="query.keyword" clearable :placeholder="t('upload.keyword_placeholder', '文件名 / 存储键 / 备注')" />
          </el-form-item>
          <el-form-item :label="t('upload.visibility.label', '可见性')">
            <el-select v-model="query.visibility" clearable :placeholder="t('upload.all_visibility', '全部可见性')" style="width: 180px">
              <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('upload.status.label', '状态')">
            <el-select v-model="query.status" clearable :placeholder="t('upload.all_status', '全部状态')" style="width: 180px">
              <el-option v-for="option in statusOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('upload.biz_module', '业务模块')">
            <el-input v-model="query.biz_module" clearable :placeholder="t('upload.biz_module_placeholder', 'biz_module')" />
          </el-form-item>
          <el-form-item :label="t('upload.biz_type', '业务类型')">
            <el-input v-model="query.biz_type" clearable :placeholder="t('upload.biz_type_placeholder', 'biz_type')" />
          </el-form-item>
          <el-form-item :label="t('upload.biz_id', '业务ID')">
            <el-input v-model="query.biz_id" clearable :placeholder="t('upload.biz_id_placeholder', 'biz_id')" />
          </el-form-item>
          <el-form-item :label="t('upload.uploaded_by', '上传人')">
            <el-input v-model="query.uploaded_by" clearable :placeholder="t('upload.uploaded_by_placeholder', 'uploaded_by')" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', '查询') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', '重置') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="original_name" :label="t('upload.file_name', '文件名')" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.original_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="visibility" :label="t('upload.visibility.label', '可见性')" width="100">
          <template #default="{ row }">
            <el-tag :type="resolveUploadVisibilityTagType(row.visibility)" effect="plain">
              {{ resolveUploadVisibilityLabel(row.visibility) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('upload.status.label', '状态')" width="110">
          <template #default="{ row }">
            <el-tag :type="resolveUploadStatusTagType(row.status)" effect="plain">
              {{ resolveUploadStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="mime_type" :label="t('upload.mime_type', 'MIME 类型')" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.mime_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="extension" :label="t('upload.extension', '扩展名')" width="110">
          <template #default="{ row }">
            {{ row.extension || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="size_bytes" :label="t('upload.size', '大小')" width="120">
          <template #default="{ row }">
            {{ formatUploadFileSize(row.size_bytes) }}
          </template>
        </el-table-column>
        <el-table-column prop="storage_driver" :label="t('upload.storage_driver', '存储驱动')" width="130">
          <template #default="{ row }">
            {{ row.storage_driver || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_module" :label="t('upload.biz_module', '业务模块')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_module || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_type" :label="t('upload.biz_type', '业务类型')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_id" :label="t('upload.biz_id', '业务ID')" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.biz_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="uploaded_by" :label="t('upload.uploaded_by', '上传人')" width="140">
          <template #default="{ row }">
            {{ row.uploaded_by || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('upload.updated_at', '更新时间')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', '操作')" width="300" fixed="right">
          <template #default="{ row }">
            <el-space wrap :size="6">
              <el-button v-permission="'upload:file:preview'" link type="primary" :loading="previewLoading && previewTargetId === row.id" @click="openPreview(row)">{{ t('upload.preview', '预览') }}</el-button>
              <el-button v-permission="'upload:file:download'" link type="success" @click="handleDownload(row)">{{ t('upload.download', '下载') }}</el-button>
              <el-button v-permission="'upload:file:bind'" link type="warning" @click="openBindDialog(row)">{{ t('upload.bind', '绑定') }}</el-button>
              <el-button v-permission="'upload:file:unbind'" link @click="handleUnbind(row)">{{ t('upload.unbind', '解绑') }}</el-button>
              <el-button v-permission="'upload:file:delete'" link type="danger" @click="handleDelete(row)">{{ t('common.delete', '删除') }}</el-button>
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

    <el-dialog v-model="uploadDialogVisible" :title="t('upload.upload_title', '上传文件')" width="760px" destroy-on-close>
      <el-alert
        :title="t('upload.upload_alert', '支持在上传时填写业务绑定信息和备注；文件内容会按后端存储策略进行校验。')"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="uploadFormRef" :model="uploadForm" :rules="uploadRules" label-width="110px" class="admin-form">
        <el-form-item :label="t('upload.file', '文件')" required>
          <el-space wrap>
            <el-button @click="triggerFileSelect">{{ t('upload.choose_file', '选择文件') }}</el-button>
            <el-tag v-if="selectedFile" effect="plain">{{ selectedFileLabel }}</el-tag>
            <el-text v-else type="info">{{ t('upload.no_file_selected', '未选择文件') }}</el-text>
          </el-space>
          <input ref="fileInputRef" type="file" class="hidden-file-input" @change="handleFileChange" />
        </el-form-item>
        <el-form-item :label="t('upload.visibility.label', '可见性')" prop="visibility">
          <el-select v-model="uploadForm.visibility" style="width: 100%" :placeholder="t('upload.visibility.placeholder', '请选择可见性')">
            <el-option v-for="option in visibilityOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('upload.biz_module', '业务模块')">
          <el-input v-model="uploadForm.biz_module" :placeholder="t('upload.biz_module_placeholder_input', '请输入业务模块')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_type', '业务类型')">
          <el-input v-model="uploadForm.biz_type" :placeholder="t('upload.biz_type_placeholder_input', '请输入业务类型')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_id', '业务ID')">
          <el-input v-model="uploadForm.biz_id" :placeholder="t('upload.biz_id_placeholder_input', '请输入业务ID')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_field', '业务字段')">
          <el-input v-model="uploadForm.biz_field" :placeholder="t('upload.biz_field_placeholder', '请输入业务字段')" />
        </el-form-item>
        <el-form-item :label="t('upload.remark', '备注')">
          <el-input v-model="uploadForm.remark" type="textarea" :rows="3" :placeholder="t('upload.remark_placeholder', '请输入备注')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">{{ t('common.cancel', '取消') }}</el-button>
        <el-button type="primary" :loading="uploadLoading" :disabled="!uploadReady" @click="submitUpload">{{ t('upload.confirm_upload', '确认上传') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="bindDialogVisible" :title="t('upload.bind_title', '绑定文件')" width="680px" destroy-on-close>
      <el-alert
        v-if="bindTarget"
        :title="t('upload.current_file', '当前文件：{name}', { name: bindTarget.original_name || bindTarget.id })"
        type="info"
        :closable="false"
        show-icon
        class="mb-4"
      />
      <el-form ref="bindFormRef" :model="bindForm" :rules="bindRules" label-width="110px" class="admin-form">
        <el-form-item :label="t('upload.biz_module', '业务模块')" prop="biz_module">
          <el-input v-model="bindForm.biz_module" :placeholder="t('upload.biz_module_placeholder_input', '请输入业务模块')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_type', '业务类型')" prop="biz_type">
          <el-input v-model="bindForm.biz_type" :placeholder="t('upload.biz_type_placeholder_input', '请输入业务类型')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_id', '业务ID')" prop="biz_id">
          <el-input v-model="bindForm.biz_id" :placeholder="t('upload.biz_id_placeholder_input', '请输入业务ID')" />
        </el-form-item>
        <el-form-item :label="t('upload.biz_field', '业务字段')" prop="biz_field">
          <el-input v-model="bindForm.biz_field" :placeholder="t('upload.biz_field_placeholder', '请输入业务字段')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bindDialogVisible = false">{{ t('common.cancel', '取消') }}</el-button>
        <el-button type="primary" :loading="bindLoading" @click="submitBind">{{ t('upload.confirm_bind', '确认绑定') }}</el-button>
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
          <el-link v-if="isBrowserDirectPublicUrl(previewItem.public_url)" plain @click="copyPreviewUrl(previewItem.public_url)">{{ t('upload.preview.copy_public_url', '复制公开地址') }}</el-link>
        </el-space>

        <el-alert
          v-if="previewKind === 'download-only'"
          :title="t('upload.preview.download_hint', '当前文件类型不适合在线预览，请使用下载按钮获取原文件。')"
          class="mb-4"
        />

        <el-descriptions :column="2" border class="mb-4">
          <el-descriptions-item :label="t('upload.file_name', '文件名')">{{ previewItem.original_name || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.visibility.label', '可见性')">{{ resolveUploadVisibilityLabel(previewItem.visibility) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.status.label', '状态')">{{ resolveUploadStatusLabel(previewItem.status) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.size', '大小')">{{ formatUploadFileSize(previewItem.size_bytes) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.mime_type', 'MIME 类型')">{{ previewItem.mime_type || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.extension', '扩展名')">{{ previewItem.extension || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.storage_driver', '存储驱动')">{{ previewItem.storage_driver || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.storage_key', '存储键')">{{ previewItem.storage_key || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_module', '业务模块')">{{ previewItem.biz_module || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_type', '业务类型')">{{ previewItem.biz_type || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_id', '业务ID')">{{ previewItem.biz_id || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.biz_field', '业务字段')">{{ previewItem.biz_field || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.uploaded_by', '上传人')">{{ previewItem.uploaded_by || '-' }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.updated_at', '更新时间')">{{ formatDateTime(previewItem.updated_at) }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.access_mode', '访问方式')">{{ getPreviewSourceLabel() }}</el-descriptions-item>
          <el-descriptions-item :label="t('upload.public_url', '公开地址')" :span="2">
            <el-link v-if="previewBrowserUrl && previewKind !== 'download-only'" type="primary" @click="openPreviewWindow">{{ t('upload.open_in_new_window', '新窗口打开') }}</el-link>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="previewBrowserUrl && isPreviewableImage(previewItem.mime_type)" class="upload-preview-image-wrap">
          <el-image :src="previewBrowserUrl" fit="contain" class="upload-preview-image" :preview-src-list="[previewBrowserUrl]" />
        </div>
        <div v-else-if="previewBrowserUrl && previewKind === 'pdf'" class="upload-preview-document-wrap">
          <iframe :src="previewBrowserUrl" class="upload-preview-document" :title="t('upload.preview.iframe_file', '文件预览')" />
        </div>
        <div v-else-if="previewBrowserUrl && previewKind === 'text'" class="upload-preview-text-wrap">
          <iframe :src="previewBrowserUrl" class="upload-preview-text" :title="t('upload.preview.iframe_text', '文本预览')" />
        </div>
        <el-alert
          v-else-if="previewBrowserUrl"
          :title="t('upload.preview.temp_url_hint', '当前文件使用临时预览地址，已在上方显示元数据。')"
          type="info"
          :closable="false"
          show-icon
        />
      </template>
      <template #footer>
        <el-button @click="previewDialogVisible = false">{{ t('common.close', '关闭') }}</el-button>
        <el-button v-if="previewItem" type="primary" @click="handleDownload(previewItem)">{{ t('upload.download_file', '下载文件') }}</el-button>
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
