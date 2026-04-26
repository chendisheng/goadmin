import { computed, nextTick, onActivated, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminTable from '@/components/admin/AdminTable.vue';
import { formatDateTime } from '@/utils/admin';
import { canSubmitUploadForm, isBrowserDirectPublicUrl, formatUploadFileSize, isPreviewableImage, resolveUploadPreviewKind, resolveUploadStatusLabel, resolveUploadStatusTagType, resolveUploadVisibilityLabel, resolveUploadVisibilityTagType, } from '@/utils/upload';
import { bindUploadFile, createUploadFilePreviewUrl, deleteUploadFile, downloadUploadFile, fetchUploadFiles, fetchUploadStorageSetting, previewUploadFile, unbindUploadFile, updateUploadStorageSetting, uploadUploadFile, } from '@/api/upload';
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
const uploadFormRef = ref();
const bindFormRef = ref();
const rows = ref([]);
const total = ref(0);
const selectedFile = ref(null);
const fileInputRef = ref(null);
const previewItem = ref(null);
const previewTargetId = ref('');
const bindTarget = ref(null);
const storageDriverOptions = [
    { value: 'local', label: '本地存储' },
    { value: 'db', label: '数据库存储' },
    { value: 's3-compatible', label: 'S3 兼容' },
    { value: 'oss', label: '阿里云 OSS' },
    { value: 'cos', label: '腾讯云 COS' },
    { value: 'qiniu', label: '七牛云' },
    { value: 'minio', label: 'MinIO' },
];
const normalizeStorageDriver = (driver) => {
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
const defaultStorageSettingForm = () => ({
    driver: 'local',
});
const storageSetting = reactive(defaultStorageSettingForm());
const query = reactive({
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
const defaultUploadForm = () => ({
    visibility: 'private',
    biz_module: '',
    biz_type: '',
    biz_id: '',
    biz_field: '',
    remark: '',
});
const defaultBindForm = () => ({
    biz_module: '',
    biz_type: '',
    biz_id: '',
    biz_field: '',
});
const uploadForm = reactive(defaultUploadForm());
const bindForm = reactive(defaultBindForm());
const uploadRules = {
    visibility: [{ required: true, message: '请选择文件可见性', trigger: 'change' }],
};
const bindRules = {
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
        return '文件预览';
    }
    const kindLabel = previewKind.value === 'image' ? '图片预览' : previewKind.value === 'pdf' ? 'PDF 预览' : previewKind.value === 'text' ? '文本预览' : '仅下载';
    return `${previewItem.value.original_name || '文件预览'} · ${kindLabel}`;
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
    }
    finally {
        tableLoading.value = false;
    }
}
async function loadStorageSetting() {
    storageSettingLoading.value = true;
    try {
        const response = await fetchUploadStorageSetting();
        storageSetting.driver = normalizeStorageDriver(response.driver);
    }
    catch {
        storageSetting.driver = 'local';
    }
    finally {
        storageSettingLoading.value = false;
    }
}
async function submitStorageSetting() {
    const driver = normalizeStorageDriver(storageSetting.driver);
    storageSetting.driver = driver;
    storageSettingSaving.value = true;
    try {
        const response = await updateUploadStorageSetting({ driver });
        storageSetting.driver = normalizeStorageDriver(response.driver || driver);
        ElMessage.success('默认存储驱动已保存');
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '保存失败');
    }
    finally {
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
function handleFileChange(event) {
    const input = event.target;
    const file = input?.files?.[0] ?? null;
    selectedFile.value = file;
}
function openBindDialog(row) {
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
    }
    catch {
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
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '上传失败');
    }
    finally {
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
    }
    catch {
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
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '绑定失败');
    }
    finally {
        bindLoading.value = false;
    }
}
async function openPreview(row) {
    previewLoading.value = true;
    previewTargetId.value = row.id;
    try {
        revokePreviewBrowserUrl();
        previewItem.value = await previewUploadFile(row.id);
        const publicUrl = previewItem.value.public_url ?? '';
        if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
            previewBrowserUrl.value = '';
            previewBrowserUrlIsObjectUrl.value = false;
        }
        else if (previewMode.value === 'public_url' && publicUrl && isBrowserDirectPublicUrl(publicUrl)) {
            previewBrowserUrl.value = publicUrl;
            previewBrowserUrlIsObjectUrl.value = false;
        }
        else if (previewItem.value.download_url) {
            previewBrowserUrl.value = await createUploadFilePreviewUrl(row.id);
            previewBrowserUrlIsObjectUrl.value = true;
        }
        previewDialogVisible.value = true;
    }
    catch (error) {
        revokePreviewBrowserUrl();
        ElMessage.error(error instanceof Error ? error.message : '预览失败');
    }
    finally {
        previewLoading.value = false;
        previewTargetId.value = '';
    }
}
async function copyPreviewUrl(url) {
    if (!url) {
        return;
    }
    try {
        await navigator.clipboard.writeText(url);
        ElMessage.success('公开地址已复制');
    }
    catch {
        ElMessage.warning('复制失败，请手动复制');
    }
}
async function openPreviewWindow() {
    try {
        const previewUrl = previewBrowserUrl.value;
        if (!previewUrl) {
            ElMessage.warning('暂无可用的预览地址');
            return;
        }
        const opened = window.open(previewUrl, '_blank', 'noopener,noreferrer');
        if (!opened) {
            ElMessage.warning('浏览器拦截了新窗口，请允许弹窗后重试');
            return;
        }
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '打开预览失败');
    }
}
function getPreviewSourceLabel() {
    if (!previewItem.value) {
        return '-';
    }
    if (previewMode.value === 'download_only' || previewKind.value === 'download-only') {
        return '仅下载';
    }
    if (previewMode.value === 'public_url' && isBrowserDirectPublicUrl(previewItem.value.public_url)) {
        return '公开直连';
    }
    return '鉴权下载';
}
async function handleDownload(row) {
    try {
        await downloadUploadFile(row.id, row.original_name || row.storage_name || 'upload-file');
        ElMessage.success('文件已开始下载');
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '下载失败');
    }
}
async function handleDelete(row) {
    await ElMessageBox.confirm(`确认删除文件 ${row.original_name || row.id} 吗？`, '删除文件', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
    });
    await deleteUploadFile(row.id);
    ElMessage.success('文件已删除');
    await loadFiles();
}
async function handleUnbind(row) {
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
function handlePageChange(page) {
    query.page = page;
    void loadFiles();
}
function handleSizeChange(pageSize) {
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
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-page" },
});
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "mb-4" },
    shadow: "never",
}));
const __VLS_2 = __VLS_1({
    ...{ class: "mb-4" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-setting-card" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-setting-content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "upload-setting-title" },
});
const __VLS_4 = {}.ElText;
/** @type {[typeof __VLS_components.ElText, typeof __VLS_components.elText, typeof __VLS_components.ElText, typeof __VLS_components.elText, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    type: "info",
}));
const __VLS_6 = __VLS_5({
    type: "info",
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
var __VLS_7;
const __VLS_8 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    wrap: true,
}));
const __VLS_10 = __VLS_9({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
const __VLS_12 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    modelValue: (__VLS_ctx.storageSetting.driver),
    loading: (__VLS_ctx.storageSettingLoading),
    disabled: (__VLS_ctx.storageSettingSaving),
    ...{ style: {} },
}));
const __VLS_14 = __VLS_13({
    modelValue: (__VLS_ctx.storageSetting.driver),
    loading: (__VLS_ctx.storageSettingLoading),
    disabled: (__VLS_ctx.storageSettingSaving),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
for (const [option] of __VLS_getVForSourceType((__VLS_ctx.storageDriverOptions))) {
    const __VLS_16 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }));
    const __VLS_18 = __VLS_17({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_17));
}
var __VLS_15;
const __VLS_20 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.storageSettingSaving),
}));
const __VLS_22 = __VLS_21({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.storageSettingSaving),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
let __VLS_24;
let __VLS_25;
let __VLS_26;
const __VLS_27 = {
    onClick: (__VLS_ctx.submitStorageSetting)
};
__VLS_23.slots.default;
var __VLS_23;
var __VLS_11;
var __VLS_3;
/** @type {[typeof AdminTable, typeof AdminTable, ]} */ ;
// @ts-ignore
const __VLS_28 = __VLS_asFunctionalComponent(AdminTable, new AdminTable({
    title: "文件管理",
    description: "管理上传文件、绑定业务对象、下载文件与查看文件元数据。",
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_29 = __VLS_28({
    title: "文件管理",
    description: "管理上传文件、绑定业务对象、下载文件与查看文件元数据。",
    loading: (__VLS_ctx.tableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_28));
__VLS_30.slots.default;
{
    const { actions: __VLS_thisSlot } = __VLS_30.slots;
    const __VLS_31 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }));
    const __VLS_33 = __VLS_32({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    let __VLS_35;
    let __VLS_36;
    let __VLS_37;
    const __VLS_38 = {
        onClick: (__VLS_ctx.loadFiles)
    };
    __VLS_34.slots.default;
    var __VLS_34;
    const __VLS_39 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_41 = __VLS_40({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    let __VLS_43;
    let __VLS_44;
    let __VLS_45;
    const __VLS_46 = {
        onClick: (__VLS_ctx.openUploadDialog)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:create') }, null, null);
    __VLS_42.slots.default;
    var __VLS_42;
}
{
    const { filters: __VLS_thisSlot } = __VLS_30.slots;
    const __VLS_47 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }));
    const __VLS_49 = __VLS_48({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_48));
    __VLS_50.slots.default;
    const __VLS_51 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        label: "关键字",
    }));
    const __VLS_53 = __VLS_52({
        label: "关键字",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    __VLS_54.slots.default;
    const __VLS_55 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: "文件名 / 存储键 / 备注",
    }));
    const __VLS_57 = __VLS_56({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: "文件名 / 存储键 / 备注",
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
    var __VLS_54;
    const __VLS_59 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        label: "可见性",
    }));
    const __VLS_61 = __VLS_60({
        label: "可见性",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    __VLS_62.slots.default;
    const __VLS_63 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        modelValue: (__VLS_ctx.query.visibility),
        clearable: true,
        placeholder: "全部可见性",
        ...{ style: {} },
    }));
    const __VLS_65 = __VLS_64({
        modelValue: (__VLS_ctx.query.visibility),
        clearable: true,
        placeholder: "全部可见性",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
    __VLS_66.slots.default;
    for (const [option] of __VLS_getVForSourceType((__VLS_ctx.visibilityOptions))) {
        const __VLS_67 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }));
        const __VLS_69 = __VLS_68({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }, ...__VLS_functionalComponentArgsRest(__VLS_68));
    }
    var __VLS_66;
    var __VLS_62;
    const __VLS_71 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
        label: "状态",
    }));
    const __VLS_73 = __VLS_72({
        label: "状态",
    }, ...__VLS_functionalComponentArgsRest(__VLS_72));
    __VLS_74.slots.default;
    const __VLS_75 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: "全部状态",
        ...{ style: {} },
    }));
    const __VLS_77 = __VLS_76({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: "全部状态",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    __VLS_78.slots.default;
    for (const [option] of __VLS_getVForSourceType((__VLS_ctx.statusOptions))) {
        const __VLS_79 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }));
        const __VLS_81 = __VLS_80({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }, ...__VLS_functionalComponentArgsRest(__VLS_80));
    }
    var __VLS_78;
    var __VLS_74;
    const __VLS_83 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
        label: "业务模块",
    }));
    const __VLS_85 = __VLS_84({
        label: "业务模块",
    }, ...__VLS_functionalComponentArgsRest(__VLS_84));
    __VLS_86.slots.default;
    const __VLS_87 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
        modelValue: (__VLS_ctx.query.biz_module),
        clearable: true,
        placeholder: "biz_module",
    }));
    const __VLS_89 = __VLS_88({
        modelValue: (__VLS_ctx.query.biz_module),
        clearable: true,
        placeholder: "biz_module",
    }, ...__VLS_functionalComponentArgsRest(__VLS_88));
    var __VLS_86;
    const __VLS_91 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
        label: "业务类型",
    }));
    const __VLS_93 = __VLS_92({
        label: "业务类型",
    }, ...__VLS_functionalComponentArgsRest(__VLS_92));
    __VLS_94.slots.default;
    const __VLS_95 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
        modelValue: (__VLS_ctx.query.biz_type),
        clearable: true,
        placeholder: "biz_type",
    }));
    const __VLS_97 = __VLS_96({
        modelValue: (__VLS_ctx.query.biz_type),
        clearable: true,
        placeholder: "biz_type",
    }, ...__VLS_functionalComponentArgsRest(__VLS_96));
    var __VLS_94;
    const __VLS_99 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
        label: "业务ID",
    }));
    const __VLS_101 = __VLS_100({
        label: "业务ID",
    }, ...__VLS_functionalComponentArgsRest(__VLS_100));
    __VLS_102.slots.default;
    const __VLS_103 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
        modelValue: (__VLS_ctx.query.biz_id),
        clearable: true,
        placeholder: "biz_id",
    }));
    const __VLS_105 = __VLS_104({
        modelValue: (__VLS_ctx.query.biz_id),
        clearable: true,
        placeholder: "biz_id",
    }, ...__VLS_functionalComponentArgsRest(__VLS_104));
    var __VLS_102;
    const __VLS_107 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
        label: "上传人",
    }));
    const __VLS_109 = __VLS_108({
        label: "上传人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_108));
    __VLS_110.slots.default;
    const __VLS_111 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
        modelValue: (__VLS_ctx.query.uploaded_by),
        clearable: true,
        placeholder: "uploaded_by",
    }));
    const __VLS_113 = __VLS_112({
        modelValue: (__VLS_ctx.query.uploaded_by),
        clearable: true,
        placeholder: "uploaded_by",
    }, ...__VLS_functionalComponentArgsRest(__VLS_112));
    var __VLS_110;
    const __VLS_115 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({}));
    const __VLS_117 = __VLS_116({}, ...__VLS_functionalComponentArgsRest(__VLS_116));
    __VLS_118.slots.default;
    const __VLS_119 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_121 = __VLS_120({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    let __VLS_123;
    let __VLS_124;
    let __VLS_125;
    const __VLS_126 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_122.slots.default;
    var __VLS_122;
    const __VLS_127 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
        ...{ 'onClick': {} },
    }));
    const __VLS_129 = __VLS_128({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_128));
    let __VLS_131;
    let __VLS_132;
    let __VLS_133;
    const __VLS_134 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_130.slots.default;
    var __VLS_130;
    var __VLS_118;
    var __VLS_50;
}
const __VLS_135 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}));
const __VLS_137 = __VLS_136({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_138.slots.default;
const __VLS_139 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
    prop: "original_name",
    label: "文件名",
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_141 = __VLS_140({
    prop: "original_name",
    label: "文件名",
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
__VLS_142.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_142.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.original_name || '-');
}
var __VLS_142;
const __VLS_143 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    prop: "visibility",
    label: "可见性",
    width: "100",
}));
const __VLS_145 = __VLS_144({
    prop: "visibility",
    label: "可见性",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
__VLS_146.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_146.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_147 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
        type: (__VLS_ctx.resolveUploadVisibilityTagType(row.visibility)),
        effect: "plain",
    }));
    const __VLS_149 = __VLS_148({
        type: (__VLS_ctx.resolveUploadVisibilityTagType(row.visibility)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_148));
    __VLS_150.slots.default;
    (__VLS_ctx.resolveUploadVisibilityLabel(row.visibility));
    var __VLS_150;
}
var __VLS_146;
const __VLS_151 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    prop: "status",
    label: "状态",
    width: "110",
}));
const __VLS_153 = __VLS_152({
    prop: "status",
    label: "状态",
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_154.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_155 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
        type: (__VLS_ctx.resolveUploadStatusTagType(row.status)),
        effect: "plain",
    }));
    const __VLS_157 = __VLS_156({
        type: (__VLS_ctx.resolveUploadStatusTagType(row.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_156));
    __VLS_158.slots.default;
    (__VLS_ctx.resolveUploadStatusLabel(row.status));
    var __VLS_158;
}
var __VLS_154;
const __VLS_159 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    prop: "mime_type",
    label: "MIME 类型",
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_161 = __VLS_160({
    prop: "mime_type",
    label: "MIME 类型",
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_162.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.mime_type || '-');
}
var __VLS_162;
const __VLS_163 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    prop: "extension",
    label: "扩展名",
    width: "110",
}));
const __VLS_165 = __VLS_164({
    prop: "extension",
    label: "扩展名",
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
__VLS_166.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_166.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.extension || '-');
}
var __VLS_166;
const __VLS_167 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    prop: "size_bytes",
    label: "大小",
    width: "120",
}));
const __VLS_169 = __VLS_168({
    prop: "size_bytes",
    label: "大小",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_170.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatUploadFileSize(row.size_bytes));
}
var __VLS_170;
const __VLS_171 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    prop: "storage_driver",
    label: "存储驱动",
    width: "130",
}));
const __VLS_173 = __VLS_172({
    prop: "storage_driver",
    label: "存储驱动",
    width: "130",
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
__VLS_174.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_174.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.storage_driver || '-');
}
var __VLS_174;
const __VLS_175 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    prop: "biz_module",
    label: "业务模块",
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_177 = __VLS_176({
    prop: "biz_module",
    label: "业务模块",
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
__VLS_178.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_178.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.biz_module || '-');
}
var __VLS_178;
const __VLS_179 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    prop: "biz_type",
    label: "业务类型",
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_181 = __VLS_180({
    prop: "biz_type",
    label: "业务类型",
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
__VLS_182.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_182.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.biz_type || '-');
}
var __VLS_182;
const __VLS_183 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    prop: "biz_id",
    label: "业务ID",
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_185 = __VLS_184({
    prop: "biz_id",
    label: "业务ID",
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_186.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.biz_id || '-');
}
var __VLS_186;
const __VLS_187 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    prop: "uploaded_by",
    label: "上传人",
    width: "140",
}));
const __VLS_189 = __VLS_188({
    prop: "uploaded_by",
    label: "上传人",
    width: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
__VLS_190.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_190.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.uploaded_by || '-');
}
var __VLS_190;
const __VLS_191 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    label: "更新时间",
    minWidth: "180",
}));
const __VLS_193 = __VLS_192({
    label: "更新时间",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
__VLS_194.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_194.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_194;
const __VLS_195 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
    label: "操作",
    width: "300",
    fixed: "right",
}));
const __VLS_197 = __VLS_196({
    label: "操作",
    width: "300",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_196));
__VLS_198.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_198.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_199 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_200 = __VLS_asFunctionalComponent(__VLS_199, new __VLS_199({
        wrap: true,
        size: (6),
    }));
    const __VLS_201 = __VLS_200({
        wrap: true,
        size: (6),
    }, ...__VLS_functionalComponentArgsRest(__VLS_200));
    __VLS_202.slots.default;
    const __VLS_203 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        loading: (__VLS_ctx.previewLoading && __VLS_ctx.previewTargetId === row.id),
    }));
    const __VLS_205 = __VLS_204({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        loading: (__VLS_ctx.previewLoading && __VLS_ctx.previewTargetId === row.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_204));
    let __VLS_207;
    let __VLS_208;
    let __VLS_209;
    const __VLS_210 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openPreview(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:preview') }, null, null);
    __VLS_206.slots.default;
    var __VLS_206;
    const __VLS_211 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
        ...{ 'onClick': {} },
        link: true,
        type: "success",
    }));
    const __VLS_213 = __VLS_212({
        ...{ 'onClick': {} },
        link: true,
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_212));
    let __VLS_215;
    let __VLS_216;
    let __VLS_217;
    const __VLS_218 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDownload(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:download') }, null, null);
    __VLS_214.slots.default;
    var __VLS_214;
    const __VLS_219 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_220 = __VLS_asFunctionalComponent(__VLS_219, new __VLS_219({
        ...{ 'onClick': {} },
        link: true,
        type: "warning",
    }));
    const __VLS_221 = __VLS_220({
        ...{ 'onClick': {} },
        link: true,
        type: "warning",
    }, ...__VLS_functionalComponentArgsRest(__VLS_220));
    let __VLS_223;
    let __VLS_224;
    let __VLS_225;
    const __VLS_226 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openBindDialog(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:bind') }, null, null);
    __VLS_222.slots.default;
    var __VLS_222;
    const __VLS_227 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_228 = __VLS_asFunctionalComponent(__VLS_227, new __VLS_227({
        ...{ 'onClick': {} },
        link: true,
    }));
    const __VLS_229 = __VLS_228({
        ...{ 'onClick': {} },
        link: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_228));
    let __VLS_231;
    let __VLS_232;
    let __VLS_233;
    const __VLS_234 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleUnbind(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:unbind') }, null, null);
    __VLS_230.slots.default;
    var __VLS_230;
    const __VLS_235 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_236 = __VLS_asFunctionalComponent(__VLS_235, new __VLS_235({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_237 = __VLS_236({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_236));
    let __VLS_239;
    let __VLS_240;
    let __VLS_241;
    const __VLS_242 = {
        onClick: (...[$event]) => {
            __VLS_ctx.handleDelete(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('upload:file:delete') }, null, null);
    __VLS_238.slots.default;
    var __VLS_238;
    var __VLS_202;
}
var __VLS_198;
var __VLS_138;
{
    const { footer: __VLS_thisSlot } = __VLS_30.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_243 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_244 = __VLS_asFunctionalComponent(__VLS_243, new __VLS_243({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_245 = __VLS_244({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_244));
    let __VLS_247;
    let __VLS_248;
    let __VLS_249;
    const __VLS_250 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_251 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_246;
}
var __VLS_30;
const __VLS_252 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_253 = __VLS_asFunctionalComponent(__VLS_252, new __VLS_252({
    modelValue: (__VLS_ctx.uploadDialogVisible),
    title: "上传文件",
    width: "760px",
    destroyOnClose: true,
}));
const __VLS_254 = __VLS_253({
    modelValue: (__VLS_ctx.uploadDialogVisible),
    title: "上传文件",
    width: "760px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_253));
__VLS_255.slots.default;
const __VLS_256 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
    title: "支持在上传时填写业务绑定信息和备注；文件内容会按后端存储策略进行校验。",
    type: "info",
    closable: (false),
    showIcon: true,
    ...{ class: "mb-4" },
}));
const __VLS_258 = __VLS_257({
    title: "支持在上传时填写业务绑定信息和备注；文件内容会按后端存储策略进行校验。",
    type: "info",
    closable: (false),
    showIcon: true,
    ...{ class: "mb-4" },
}, ...__VLS_functionalComponentArgsRest(__VLS_257));
const __VLS_260 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_261 = __VLS_asFunctionalComponent(__VLS_260, new __VLS_260({
    ref: "uploadFormRef",
    model: (__VLS_ctx.uploadForm),
    rules: (__VLS_ctx.uploadRules),
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_262 = __VLS_261({
    ref: "uploadFormRef",
    model: (__VLS_ctx.uploadForm),
    rules: (__VLS_ctx.uploadRules),
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_261));
/** @type {typeof __VLS_ctx.uploadFormRef} */ ;
var __VLS_264 = {};
__VLS_263.slots.default;
const __VLS_266 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_267 = __VLS_asFunctionalComponent(__VLS_266, new __VLS_266({
    label: "文件",
    required: true,
}));
const __VLS_268 = __VLS_267({
    label: "文件",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_267));
__VLS_269.slots.default;
const __VLS_270 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_271 = __VLS_asFunctionalComponent(__VLS_270, new __VLS_270({
    wrap: true,
}));
const __VLS_272 = __VLS_271({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_271));
__VLS_273.slots.default;
const __VLS_274 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_275 = __VLS_asFunctionalComponent(__VLS_274, new __VLS_274({
    ...{ 'onClick': {} },
}));
const __VLS_276 = __VLS_275({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_275));
let __VLS_278;
let __VLS_279;
let __VLS_280;
const __VLS_281 = {
    onClick: (__VLS_ctx.triggerFileSelect)
};
__VLS_277.slots.default;
var __VLS_277;
if (__VLS_ctx.selectedFile) {
    const __VLS_282 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_283 = __VLS_asFunctionalComponent(__VLS_282, new __VLS_282({
        effect: "plain",
    }));
    const __VLS_284 = __VLS_283({
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_283));
    __VLS_285.slots.default;
    (__VLS_ctx.selectedFileLabel);
    var __VLS_285;
}
else {
    const __VLS_286 = {}.ElText;
    /** @type {[typeof __VLS_components.ElText, typeof __VLS_components.elText, typeof __VLS_components.ElText, typeof __VLS_components.elText, ]} */ ;
    // @ts-ignore
    const __VLS_287 = __VLS_asFunctionalComponent(__VLS_286, new __VLS_286({
        type: "info",
    }));
    const __VLS_288 = __VLS_287({
        type: "info",
    }, ...__VLS_functionalComponentArgsRest(__VLS_287));
    __VLS_289.slots.default;
    var __VLS_289;
}
var __VLS_273;
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onChange: (__VLS_ctx.handleFileChange) },
    ref: "fileInputRef",
    type: "file",
    ...{ class: "hidden-file-input" },
});
/** @type {typeof __VLS_ctx.fileInputRef} */ ;
var __VLS_269;
const __VLS_290 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_291 = __VLS_asFunctionalComponent(__VLS_290, new __VLS_290({
    label: "可见性",
    prop: "visibility",
}));
const __VLS_292 = __VLS_291({
    label: "可见性",
    prop: "visibility",
}, ...__VLS_functionalComponentArgsRest(__VLS_291));
__VLS_293.slots.default;
const __VLS_294 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_295 = __VLS_asFunctionalComponent(__VLS_294, new __VLS_294({
    modelValue: (__VLS_ctx.uploadForm.visibility),
    ...{ style: {} },
    placeholder: "请选择可见性",
}));
const __VLS_296 = __VLS_295({
    modelValue: (__VLS_ctx.uploadForm.visibility),
    ...{ style: {} },
    placeholder: "请选择可见性",
}, ...__VLS_functionalComponentArgsRest(__VLS_295));
__VLS_297.slots.default;
for (const [option] of __VLS_getVForSourceType((__VLS_ctx.visibilityOptions))) {
    const __VLS_298 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_299 = __VLS_asFunctionalComponent(__VLS_298, new __VLS_298({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }));
    const __VLS_300 = __VLS_299({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_299));
}
var __VLS_297;
var __VLS_293;
const __VLS_302 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_303 = __VLS_asFunctionalComponent(__VLS_302, new __VLS_302({
    label: "业务模块",
}));
const __VLS_304 = __VLS_303({
    label: "业务模块",
}, ...__VLS_functionalComponentArgsRest(__VLS_303));
__VLS_305.slots.default;
const __VLS_306 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_307 = __VLS_asFunctionalComponent(__VLS_306, new __VLS_306({
    modelValue: (__VLS_ctx.uploadForm.biz_module),
    placeholder: "请输入业务模块",
}));
const __VLS_308 = __VLS_307({
    modelValue: (__VLS_ctx.uploadForm.biz_module),
    placeholder: "请输入业务模块",
}, ...__VLS_functionalComponentArgsRest(__VLS_307));
var __VLS_305;
const __VLS_310 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_311 = __VLS_asFunctionalComponent(__VLS_310, new __VLS_310({
    label: "业务类型",
}));
const __VLS_312 = __VLS_311({
    label: "业务类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_311));
__VLS_313.slots.default;
const __VLS_314 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_315 = __VLS_asFunctionalComponent(__VLS_314, new __VLS_314({
    modelValue: (__VLS_ctx.uploadForm.biz_type),
    placeholder: "请输入业务类型",
}));
const __VLS_316 = __VLS_315({
    modelValue: (__VLS_ctx.uploadForm.biz_type),
    placeholder: "请输入业务类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_315));
var __VLS_313;
const __VLS_318 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_319 = __VLS_asFunctionalComponent(__VLS_318, new __VLS_318({
    label: "业务ID",
}));
const __VLS_320 = __VLS_319({
    label: "业务ID",
}, ...__VLS_functionalComponentArgsRest(__VLS_319));
__VLS_321.slots.default;
const __VLS_322 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_323 = __VLS_asFunctionalComponent(__VLS_322, new __VLS_322({
    modelValue: (__VLS_ctx.uploadForm.biz_id),
    placeholder: "请输入业务ID",
}));
const __VLS_324 = __VLS_323({
    modelValue: (__VLS_ctx.uploadForm.biz_id),
    placeholder: "请输入业务ID",
}, ...__VLS_functionalComponentArgsRest(__VLS_323));
var __VLS_321;
const __VLS_326 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_327 = __VLS_asFunctionalComponent(__VLS_326, new __VLS_326({
    label: "业务字段",
}));
const __VLS_328 = __VLS_327({
    label: "业务字段",
}, ...__VLS_functionalComponentArgsRest(__VLS_327));
__VLS_329.slots.default;
const __VLS_330 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_331 = __VLS_asFunctionalComponent(__VLS_330, new __VLS_330({
    modelValue: (__VLS_ctx.uploadForm.biz_field),
    placeholder: "请输入业务字段",
}));
const __VLS_332 = __VLS_331({
    modelValue: (__VLS_ctx.uploadForm.biz_field),
    placeholder: "请输入业务字段",
}, ...__VLS_functionalComponentArgsRest(__VLS_331));
var __VLS_329;
const __VLS_334 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_335 = __VLS_asFunctionalComponent(__VLS_334, new __VLS_334({
    label: "备注",
}));
const __VLS_336 = __VLS_335({
    label: "备注",
}, ...__VLS_functionalComponentArgsRest(__VLS_335));
__VLS_337.slots.default;
const __VLS_338 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_339 = __VLS_asFunctionalComponent(__VLS_338, new __VLS_338({
    modelValue: (__VLS_ctx.uploadForm.remark),
    type: "textarea",
    rows: (3),
    placeholder: "请输入备注",
}));
const __VLS_340 = __VLS_339({
    modelValue: (__VLS_ctx.uploadForm.remark),
    type: "textarea",
    rows: (3),
    placeholder: "请输入备注",
}, ...__VLS_functionalComponentArgsRest(__VLS_339));
var __VLS_337;
var __VLS_263;
{
    const { footer: __VLS_thisSlot } = __VLS_255.slots;
    const __VLS_342 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_343 = __VLS_asFunctionalComponent(__VLS_342, new __VLS_342({
        ...{ 'onClick': {} },
    }));
    const __VLS_344 = __VLS_343({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_343));
    let __VLS_346;
    let __VLS_347;
    let __VLS_348;
    const __VLS_349 = {
        onClick: (...[$event]) => {
            __VLS_ctx.uploadDialogVisible = false;
        }
    };
    __VLS_345.slots.default;
    var __VLS_345;
    const __VLS_350 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_351 = __VLS_asFunctionalComponent(__VLS_350, new __VLS_350({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.uploadLoading),
        disabled: (!__VLS_ctx.uploadReady),
    }));
    const __VLS_352 = __VLS_351({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.uploadLoading),
        disabled: (!__VLS_ctx.uploadReady),
    }, ...__VLS_functionalComponentArgsRest(__VLS_351));
    let __VLS_354;
    let __VLS_355;
    let __VLS_356;
    const __VLS_357 = {
        onClick: (__VLS_ctx.submitUpload)
    };
    __VLS_353.slots.default;
    var __VLS_353;
}
var __VLS_255;
const __VLS_358 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_359 = __VLS_asFunctionalComponent(__VLS_358, new __VLS_358({
    modelValue: (__VLS_ctx.bindDialogVisible),
    title: "绑定文件",
    width: "680px",
    destroyOnClose: true,
}));
const __VLS_360 = __VLS_359({
    modelValue: (__VLS_ctx.bindDialogVisible),
    title: "绑定文件",
    width: "680px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_359));
__VLS_361.slots.default;
if (__VLS_ctx.bindTarget) {
    const __VLS_362 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_363 = __VLS_asFunctionalComponent(__VLS_362, new __VLS_362({
        title: (`当前文件：${__VLS_ctx.bindTarget.original_name || __VLS_ctx.bindTarget.id}`),
        type: "info",
        closable: (false),
        showIcon: true,
        ...{ class: "mb-4" },
    }));
    const __VLS_364 = __VLS_363({
        title: (`当前文件：${__VLS_ctx.bindTarget.original_name || __VLS_ctx.bindTarget.id}`),
        type: "info",
        closable: (false),
        showIcon: true,
        ...{ class: "mb-4" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_363));
}
const __VLS_366 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_367 = __VLS_asFunctionalComponent(__VLS_366, new __VLS_366({
    ref: "bindFormRef",
    model: (__VLS_ctx.bindForm),
    rules: (__VLS_ctx.bindRules),
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_368 = __VLS_367({
    ref: "bindFormRef",
    model: (__VLS_ctx.bindForm),
    rules: (__VLS_ctx.bindRules),
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_367));
/** @type {typeof __VLS_ctx.bindFormRef} */ ;
var __VLS_370 = {};
__VLS_369.slots.default;
const __VLS_372 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_373 = __VLS_asFunctionalComponent(__VLS_372, new __VLS_372({
    label: "业务模块",
    prop: "biz_module",
}));
const __VLS_374 = __VLS_373({
    label: "业务模块",
    prop: "biz_module",
}, ...__VLS_functionalComponentArgsRest(__VLS_373));
__VLS_375.slots.default;
const __VLS_376 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_377 = __VLS_asFunctionalComponent(__VLS_376, new __VLS_376({
    modelValue: (__VLS_ctx.bindForm.biz_module),
    placeholder: "请输入业务模块",
}));
const __VLS_378 = __VLS_377({
    modelValue: (__VLS_ctx.bindForm.biz_module),
    placeholder: "请输入业务模块",
}, ...__VLS_functionalComponentArgsRest(__VLS_377));
var __VLS_375;
const __VLS_380 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_381 = __VLS_asFunctionalComponent(__VLS_380, new __VLS_380({
    label: "业务类型",
    prop: "biz_type",
}));
const __VLS_382 = __VLS_381({
    label: "业务类型",
    prop: "biz_type",
}, ...__VLS_functionalComponentArgsRest(__VLS_381));
__VLS_383.slots.default;
const __VLS_384 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_385 = __VLS_asFunctionalComponent(__VLS_384, new __VLS_384({
    modelValue: (__VLS_ctx.bindForm.biz_type),
    placeholder: "请输入业务类型",
}));
const __VLS_386 = __VLS_385({
    modelValue: (__VLS_ctx.bindForm.biz_type),
    placeholder: "请输入业务类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_385));
var __VLS_383;
const __VLS_388 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_389 = __VLS_asFunctionalComponent(__VLS_388, new __VLS_388({
    label: "业务ID",
    prop: "biz_id",
}));
const __VLS_390 = __VLS_389({
    label: "业务ID",
    prop: "biz_id",
}, ...__VLS_functionalComponentArgsRest(__VLS_389));
__VLS_391.slots.default;
const __VLS_392 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_393 = __VLS_asFunctionalComponent(__VLS_392, new __VLS_392({
    modelValue: (__VLS_ctx.bindForm.biz_id),
    placeholder: "请输入业务ID",
}));
const __VLS_394 = __VLS_393({
    modelValue: (__VLS_ctx.bindForm.biz_id),
    placeholder: "请输入业务ID",
}, ...__VLS_functionalComponentArgsRest(__VLS_393));
var __VLS_391;
const __VLS_396 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_397 = __VLS_asFunctionalComponent(__VLS_396, new __VLS_396({
    label: "业务字段",
    prop: "biz_field",
}));
const __VLS_398 = __VLS_397({
    label: "业务字段",
    prop: "biz_field",
}, ...__VLS_functionalComponentArgsRest(__VLS_397));
__VLS_399.slots.default;
const __VLS_400 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_401 = __VLS_asFunctionalComponent(__VLS_400, new __VLS_400({
    modelValue: (__VLS_ctx.bindForm.biz_field),
    placeholder: "请输入业务字段",
}));
const __VLS_402 = __VLS_401({
    modelValue: (__VLS_ctx.bindForm.biz_field),
    placeholder: "请输入业务字段",
}, ...__VLS_functionalComponentArgsRest(__VLS_401));
var __VLS_399;
var __VLS_369;
{
    const { footer: __VLS_thisSlot } = __VLS_361.slots;
    const __VLS_404 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_405 = __VLS_asFunctionalComponent(__VLS_404, new __VLS_404({
        ...{ 'onClick': {} },
    }));
    const __VLS_406 = __VLS_405({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_405));
    let __VLS_408;
    let __VLS_409;
    let __VLS_410;
    const __VLS_411 = {
        onClick: (...[$event]) => {
            __VLS_ctx.bindDialogVisible = false;
        }
    };
    __VLS_407.slots.default;
    var __VLS_407;
    const __VLS_412 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_413 = __VLS_asFunctionalComponent(__VLS_412, new __VLS_412({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.bindLoading),
    }));
    const __VLS_414 = __VLS_413({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.bindLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_413));
    let __VLS_416;
    let __VLS_417;
    let __VLS_418;
    const __VLS_419 = {
        onClick: (__VLS_ctx.submitBind)
    };
    __VLS_415.slots.default;
    var __VLS_415;
}
var __VLS_361;
const __VLS_420 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_421 = __VLS_asFunctionalComponent(__VLS_420, new __VLS_420({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.previewDialogVisible),
    title: (__VLS_ctx.previewTitle),
    width: "840px",
    destroyOnClose: true,
}));
const __VLS_422 = __VLS_421({
    ...{ 'onClosed': {} },
    modelValue: (__VLS_ctx.previewDialogVisible),
    title: (__VLS_ctx.previewTitle),
    width: "840px",
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_421));
let __VLS_424;
let __VLS_425;
let __VLS_426;
const __VLS_427 = {
    onClosed: (__VLS_ctx.revokePreviewBrowserUrl)
};
__VLS_423.slots.default;
if (__VLS_ctx.previewItem) {
    const __VLS_428 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_429 = __VLS_asFunctionalComponent(__VLS_428, new __VLS_428({
        wrap: true,
        ...{ class: "mb-4" },
    }));
    const __VLS_430 = __VLS_429({
        wrap: true,
        ...{ class: "mb-4" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_429));
    __VLS_431.slots.default;
    const __VLS_432 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_433 = __VLS_asFunctionalComponent(__VLS_432, new __VLS_432({
        type: (__VLS_ctx.resolveUploadVisibilityTagType(__VLS_ctx.previewItem.visibility)),
        effect: "plain",
    }));
    const __VLS_434 = __VLS_433({
        type: (__VLS_ctx.resolveUploadVisibilityTagType(__VLS_ctx.previewItem.visibility)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_433));
    __VLS_435.slots.default;
    (__VLS_ctx.resolveUploadVisibilityLabel(__VLS_ctx.previewItem.visibility));
    var __VLS_435;
    const __VLS_436 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_437 = __VLS_asFunctionalComponent(__VLS_436, new __VLS_436({
        type: (__VLS_ctx.resolveUploadStatusTagType(__VLS_ctx.previewItem.status)),
        effect: "plain",
    }));
    const __VLS_438 = __VLS_437({
        type: (__VLS_ctx.resolveUploadStatusTagType(__VLS_ctx.previewItem.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_437));
    __VLS_439.slots.default;
    (__VLS_ctx.resolveUploadStatusLabel(__VLS_ctx.previewItem.status));
    var __VLS_439;
    if (__VLS_ctx.isBrowserDirectPublicUrl(__VLS_ctx.previewItem.public_url)) {
        const __VLS_440 = {}.ElLink;
        /** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
        // @ts-ignore
        const __VLS_441 = __VLS_asFunctionalComponent(__VLS_440, new __VLS_440({
            ...{ 'onClick': {} },
            plain: true,
        }));
        const __VLS_442 = __VLS_441({
            ...{ 'onClick': {} },
            plain: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_441));
        let __VLS_444;
        let __VLS_445;
        let __VLS_446;
        const __VLS_447 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.previewItem))
                    return;
                if (!(__VLS_ctx.isBrowserDirectPublicUrl(__VLS_ctx.previewItem.public_url)))
                    return;
                __VLS_ctx.copyPreviewUrl(__VLS_ctx.previewItem.public_url);
            }
        };
        __VLS_443.slots.default;
        var __VLS_443;
    }
    var __VLS_431;
    if (__VLS_ctx.previewKind === 'download-only') {
        const __VLS_448 = {}.ElAlert;
        /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
        // @ts-ignore
        const __VLS_449 = __VLS_asFunctionalComponent(__VLS_448, new __VLS_448({
            title: "当前文件类型不适合在线预览，请使用下载按钮获取原文件。",
            ...{ class: "mb-4" },
        }));
        const __VLS_450 = __VLS_449({
            title: "当前文件类型不适合在线预览，请使用下载按钮获取原文件。",
            ...{ class: "mb-4" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_449));
    }
    const __VLS_452 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_453 = __VLS_asFunctionalComponent(__VLS_452, new __VLS_452({
        column: (2),
        border: true,
        ...{ class: "mb-4" },
    }));
    const __VLS_454 = __VLS_453({
        column: (2),
        border: true,
        ...{ class: "mb-4" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_453));
    __VLS_455.slots.default;
    const __VLS_456 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_457 = __VLS_asFunctionalComponent(__VLS_456, new __VLS_456({
        label: "文件名",
    }));
    const __VLS_458 = __VLS_457({
        label: "文件名",
    }, ...__VLS_functionalComponentArgsRest(__VLS_457));
    __VLS_459.slots.default;
    (__VLS_ctx.previewItem.original_name || '-');
    var __VLS_459;
    const __VLS_460 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_461 = __VLS_asFunctionalComponent(__VLS_460, new __VLS_460({
        label: "可见性",
    }));
    const __VLS_462 = __VLS_461({
        label: "可见性",
    }, ...__VLS_functionalComponentArgsRest(__VLS_461));
    __VLS_463.slots.default;
    (__VLS_ctx.resolveUploadVisibilityLabel(__VLS_ctx.previewItem.visibility));
    var __VLS_463;
    const __VLS_464 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_465 = __VLS_asFunctionalComponent(__VLS_464, new __VLS_464({
        label: "状态",
    }));
    const __VLS_466 = __VLS_465({
        label: "状态",
    }, ...__VLS_functionalComponentArgsRest(__VLS_465));
    __VLS_467.slots.default;
    (__VLS_ctx.resolveUploadStatusLabel(__VLS_ctx.previewItem.status));
    var __VLS_467;
    const __VLS_468 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_469 = __VLS_asFunctionalComponent(__VLS_468, new __VLS_468({
        label: "大小",
    }));
    const __VLS_470 = __VLS_469({
        label: "大小",
    }, ...__VLS_functionalComponentArgsRest(__VLS_469));
    __VLS_471.slots.default;
    (__VLS_ctx.formatUploadFileSize(__VLS_ctx.previewItem.size_bytes));
    var __VLS_471;
    const __VLS_472 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_473 = __VLS_asFunctionalComponent(__VLS_472, new __VLS_472({
        label: "MIME 类型",
    }));
    const __VLS_474 = __VLS_473({
        label: "MIME 类型",
    }, ...__VLS_functionalComponentArgsRest(__VLS_473));
    __VLS_475.slots.default;
    (__VLS_ctx.previewItem.mime_type || '-');
    var __VLS_475;
    const __VLS_476 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_477 = __VLS_asFunctionalComponent(__VLS_476, new __VLS_476({
        label: "扩展名",
    }));
    const __VLS_478 = __VLS_477({
        label: "扩展名",
    }, ...__VLS_functionalComponentArgsRest(__VLS_477));
    __VLS_479.slots.default;
    (__VLS_ctx.previewItem.extension || '-');
    var __VLS_479;
    const __VLS_480 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_481 = __VLS_asFunctionalComponent(__VLS_480, new __VLS_480({
        label: "存储驱动",
    }));
    const __VLS_482 = __VLS_481({
        label: "存储驱动",
    }, ...__VLS_functionalComponentArgsRest(__VLS_481));
    __VLS_483.slots.default;
    (__VLS_ctx.previewItem.storage_driver || '-');
    var __VLS_483;
    const __VLS_484 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_485 = __VLS_asFunctionalComponent(__VLS_484, new __VLS_484({
        label: "存储键",
    }));
    const __VLS_486 = __VLS_485({
        label: "存储键",
    }, ...__VLS_functionalComponentArgsRest(__VLS_485));
    __VLS_487.slots.default;
    (__VLS_ctx.previewItem.storage_key || '-');
    var __VLS_487;
    const __VLS_488 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_489 = __VLS_asFunctionalComponent(__VLS_488, new __VLS_488({
        label: "业务模块",
    }));
    const __VLS_490 = __VLS_489({
        label: "业务模块",
    }, ...__VLS_functionalComponentArgsRest(__VLS_489));
    __VLS_491.slots.default;
    (__VLS_ctx.previewItem.biz_module || '-');
    var __VLS_491;
    const __VLS_492 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_493 = __VLS_asFunctionalComponent(__VLS_492, new __VLS_492({
        label: "业务类型",
    }));
    const __VLS_494 = __VLS_493({
        label: "业务类型",
    }, ...__VLS_functionalComponentArgsRest(__VLS_493));
    __VLS_495.slots.default;
    (__VLS_ctx.previewItem.biz_type || '-');
    var __VLS_495;
    const __VLS_496 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_497 = __VLS_asFunctionalComponent(__VLS_496, new __VLS_496({
        label: "业务ID",
    }));
    const __VLS_498 = __VLS_497({
        label: "业务ID",
    }, ...__VLS_functionalComponentArgsRest(__VLS_497));
    __VLS_499.slots.default;
    (__VLS_ctx.previewItem.biz_id || '-');
    var __VLS_499;
    const __VLS_500 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_501 = __VLS_asFunctionalComponent(__VLS_500, new __VLS_500({
        label: "业务字段",
    }));
    const __VLS_502 = __VLS_501({
        label: "业务字段",
    }, ...__VLS_functionalComponentArgsRest(__VLS_501));
    __VLS_503.slots.default;
    (__VLS_ctx.previewItem.biz_field || '-');
    var __VLS_503;
    const __VLS_504 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_505 = __VLS_asFunctionalComponent(__VLS_504, new __VLS_504({
        label: "上传人",
    }));
    const __VLS_506 = __VLS_505({
        label: "上传人",
    }, ...__VLS_functionalComponentArgsRest(__VLS_505));
    __VLS_507.slots.default;
    (__VLS_ctx.previewItem.uploaded_by || '-');
    var __VLS_507;
    const __VLS_508 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_509 = __VLS_asFunctionalComponent(__VLS_508, new __VLS_508({
        label: "更新时间",
    }));
    const __VLS_510 = __VLS_509({
        label: "更新时间",
    }, ...__VLS_functionalComponentArgsRest(__VLS_509));
    __VLS_511.slots.default;
    (__VLS_ctx.formatDateTime(__VLS_ctx.previewItem.updated_at));
    var __VLS_511;
    const __VLS_512 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_513 = __VLS_asFunctionalComponent(__VLS_512, new __VLS_512({
        label: "访问方式",
    }));
    const __VLS_514 = __VLS_513({
        label: "访问方式",
    }, ...__VLS_functionalComponentArgsRest(__VLS_513));
    __VLS_515.slots.default;
    (__VLS_ctx.getPreviewSourceLabel());
    var __VLS_515;
    const __VLS_516 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_517 = __VLS_asFunctionalComponent(__VLS_516, new __VLS_516({
        label: "公开地址",
        span: (2),
    }));
    const __VLS_518 = __VLS_517({
        label: "公开地址",
        span: (2),
    }, ...__VLS_functionalComponentArgsRest(__VLS_517));
    __VLS_519.slots.default;
    if (__VLS_ctx.previewBrowserUrl && __VLS_ctx.previewKind !== 'download-only') {
        const __VLS_520 = {}.ElLink;
        /** @type {[typeof __VLS_components.ElLink, typeof __VLS_components.elLink, typeof __VLS_components.ElLink, typeof __VLS_components.elLink, ]} */ ;
        // @ts-ignore
        const __VLS_521 = __VLS_asFunctionalComponent(__VLS_520, new __VLS_520({
            ...{ 'onClick': {} },
            type: "primary",
        }));
        const __VLS_522 = __VLS_521({
            ...{ 'onClick': {} },
            type: "primary",
        }, ...__VLS_functionalComponentArgsRest(__VLS_521));
        let __VLS_524;
        let __VLS_525;
        let __VLS_526;
        const __VLS_527 = {
            onClick: (__VLS_ctx.openPreviewWindow)
        };
        __VLS_523.slots.default;
        var __VLS_523;
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    }
    var __VLS_519;
    var __VLS_455;
    if (__VLS_ctx.previewBrowserUrl && __VLS_ctx.isPreviewableImage(__VLS_ctx.previewItem.mime_type)) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "upload-preview-image-wrap" },
        });
        const __VLS_528 = {}.ElImage;
        /** @type {[typeof __VLS_components.ElImage, typeof __VLS_components.elImage, ]} */ ;
        // @ts-ignore
        const __VLS_529 = __VLS_asFunctionalComponent(__VLS_528, new __VLS_528({
            src: (__VLS_ctx.previewBrowserUrl),
            fit: "contain",
            ...{ class: "upload-preview-image" },
            previewSrcList: ([__VLS_ctx.previewBrowserUrl]),
        }));
        const __VLS_530 = __VLS_529({
            src: (__VLS_ctx.previewBrowserUrl),
            fit: "contain",
            ...{ class: "upload-preview-image" },
            previewSrcList: ([__VLS_ctx.previewBrowserUrl]),
        }, ...__VLS_functionalComponentArgsRest(__VLS_529));
    }
    else if (__VLS_ctx.previewBrowserUrl && __VLS_ctx.previewKind === 'pdf') {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "upload-preview-document-wrap" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.iframe)({
            src: (__VLS_ctx.previewBrowserUrl),
            ...{ class: "upload-preview-document" },
            title: "文件预览",
        });
    }
    else if (__VLS_ctx.previewBrowserUrl && __VLS_ctx.previewKind === 'text') {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "upload-preview-text-wrap" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.iframe)({
            src: (__VLS_ctx.previewBrowserUrl),
            ...{ class: "upload-preview-text" },
            title: "文本预览",
        });
    }
    else if (__VLS_ctx.previewBrowserUrl) {
        const __VLS_532 = {}.ElAlert;
        /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
        // @ts-ignore
        const __VLS_533 = __VLS_asFunctionalComponent(__VLS_532, new __VLS_532({
            title: "当前文件使用临时预览地址，已在上方显示元数据。",
            type: "info",
            closable: (false),
            showIcon: true,
        }));
        const __VLS_534 = __VLS_533({
            title: "当前文件使用临时预览地址，已在上方显示元数据。",
            type: "info",
            closable: (false),
            showIcon: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_533));
    }
}
{
    const { footer: __VLS_thisSlot } = __VLS_423.slots;
    const __VLS_536 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_537 = __VLS_asFunctionalComponent(__VLS_536, new __VLS_536({
        ...{ 'onClick': {} },
    }));
    const __VLS_538 = __VLS_537({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_537));
    let __VLS_540;
    let __VLS_541;
    let __VLS_542;
    const __VLS_543 = {
        onClick: (...[$event]) => {
            __VLS_ctx.previewDialogVisible = false;
        }
    };
    __VLS_539.slots.default;
    var __VLS_539;
    if (__VLS_ctx.previewItem) {
        const __VLS_544 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_545 = __VLS_asFunctionalComponent(__VLS_544, new __VLS_544({
            ...{ 'onClick': {} },
            type: "primary",
        }));
        const __VLS_546 = __VLS_545({
            ...{ 'onClick': {} },
            type: "primary",
        }, ...__VLS_functionalComponentArgsRest(__VLS_545));
        let __VLS_548;
        let __VLS_549;
        let __VLS_550;
        const __VLS_551 = {
            onClick: (...[$event]) => {
                if (!(__VLS_ctx.previewItem))
                    return;
                __VLS_ctx.handleDownload(__VLS_ctx.previewItem);
            }
        };
        __VLS_547.slots.default;
        var __VLS_547;
    }
}
var __VLS_423;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-setting-card']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-setting-content']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-setting-title']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-filters']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-pagination']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['hidden-file-input']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-4']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-image-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-image']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-document-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-document']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-text-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['upload-preview-text']} */ ;
// @ts-ignore
var __VLS_265 = __VLS_264, __VLS_371 = __VLS_370;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AdminTable: AdminTable,
            formatDateTime: formatDateTime,
            isBrowserDirectPublicUrl: isBrowserDirectPublicUrl,
            formatUploadFileSize: formatUploadFileSize,
            isPreviewableImage: isPreviewableImage,
            resolveUploadStatusLabel: resolveUploadStatusLabel,
            resolveUploadStatusTagType: resolveUploadStatusTagType,
            resolveUploadVisibilityLabel: resolveUploadVisibilityLabel,
            resolveUploadVisibilityTagType: resolveUploadVisibilityTagType,
            tableLoading: tableLoading,
            storageSettingLoading: storageSettingLoading,
            storageSettingSaving: storageSettingSaving,
            uploadLoading: uploadLoading,
            bindLoading: bindLoading,
            previewLoading: previewLoading,
            uploadDialogVisible: uploadDialogVisible,
            bindDialogVisible: bindDialogVisible,
            previewDialogVisible: previewDialogVisible,
            previewBrowserUrl: previewBrowserUrl,
            uploadFormRef: uploadFormRef,
            bindFormRef: bindFormRef,
            rows: rows,
            total: total,
            selectedFile: selectedFile,
            fileInputRef: fileInputRef,
            previewItem: previewItem,
            previewTargetId: previewTargetId,
            bindTarget: bindTarget,
            storageDriverOptions: storageDriverOptions,
            storageSetting: storageSetting,
            query: query,
            uploadForm: uploadForm,
            bindForm: bindForm,
            uploadRules: uploadRules,
            bindRules: bindRules,
            visibilityOptions: visibilityOptions,
            statusOptions: statusOptions,
            selectedFileLabel: selectedFileLabel,
            previewKind: previewKind,
            previewTitle: previewTitle,
            uploadReady: uploadReady,
            revokePreviewBrowserUrl: revokePreviewBrowserUrl,
            loadFiles: loadFiles,
            submitStorageSetting: submitStorageSetting,
            openUploadDialog: openUploadDialog,
            triggerFileSelect: triggerFileSelect,
            handleFileChange: handleFileChange,
            openBindDialog: openBindDialog,
            submitUpload: submitUpload,
            submitBind: submitBind,
            openPreview: openPreview,
            copyPreviewUrl: copyPreviewUrl,
            openPreviewWindow: openPreviewWindow,
            getPreviewSourceLabel: getPreviewSourceLabel,
            handleDownload: handleDownload,
            handleDelete: handleDelete,
            handleUnbind: handleUnbind,
            handleSearch: handleSearch,
            handleReset: handleReset,
            handlePageChange: handlePageChange,
            handleSizeChange: handleSizeChange,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
