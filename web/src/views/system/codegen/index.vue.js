import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { downloadCodegenArtifact, executeCodegenDelete, generateCodegenDatabase, generateCodegenDsl, generateDownloadCodegenDsl, generateDownloadCodegenDatabase, installCodegenManifest, previewCodegenDatabase, previewCodegenDsl, previewCodegenDelete, } from '@/api/codegen';
import { fetchPublicConfig } from '@/api/health';
import { fetchMenuTree } from '@/api/system-menus';
import { ApiError } from '@/api/types';
import { useAppI18n } from '@/i18n';
const activeMode = ref('dsl');
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
const dslReport = ref(null);
const dbReport = ref(null);
const deletePreviewReport = ref(null);
const deleteResult = ref(null);
const deleteRequestCache = ref(null);
const artifactInfo = ref(null);
const artifactForceExpired = ref(false);
const lastDownloadAt = ref('');
const lastArtifactError = ref('');
const lastArtifactErrorAt = ref('');
const lastArtifactErrorType = ref('unknown');
const lastDownloadDuration = ref(0);
const downloadStartAt = ref(0);
const copyFeedbackActive = ref(false);
const operationStatus = ref('');
const lastRunSuccess = ref(false);
const fileInputRef = ref(null);
const currentTime = ref(Date.now());
let artifactTicker = null;
let copyFeedbackTimer = null;
const dbMountMenuOptions = ref([]);
const publicConfig = ref(null);
const { t } = useAppI18n();
const currentReport = computed(() => (activeMode.value === 'db' ? dbReport.value : dslReport.value));
const dbParsedTables = computed(() => parseTableNames(dbTablesText.value));
const dbDriverLabel = computed(() => {
    const map = {
        mysql: 'MySQL',
        postgres: 'PostgreSQL',
        sqlite: 'SQLite',
    };
    return map[dbDriver.value] || dbDriver.value || '—';
});
const dbOptionSummary = computed(() => {
    const parts = [];
    parts.push(dbForce.value ? t('codegen.cover_existing', '覆盖现有文件') : t('codegen.safe_cover_off', '安全覆盖关闭'));
    parts.push(dbGenerateFrontend.value ? t('codegen.generate_frontend', '生成前端') : t('codegen.skip_frontend', '跳过前端'));
    parts.push(dbGeneratePolicy.value ? t('codegen.generate_policy', '生成权限') : t('codegen.skip_policy', '跳过权限'));
    parts.push(dbMountParentPath.value ? `${t('codegen.mount_root_prefix', '挂载：')}${dbMountMenuLabel.value}` : t('codegen.mount_root_top_label', '挂载：顶层根菜单'));
    return parts.join(' · ');
});
const dbGeneratedModuleName = computed(() => dbReport.value?.resources?.[0]?.module?.trim() || '');
const deletePreviewPlan = computed(() => deletePreviewReport.value?.plan ?? null);
const deletePlanSummary = computed(() => deletePreviewPlan.value?.summary ?? null);
const deletePlanItems = computed(() => mapDeletePlanToPreviewRows(deletePreviewPlan.value));
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
    const messages = [];
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
const messagePanelSubtitle = computed(() => activeMode.value === 'delete'
    ? t('codegen.delete_messages_subtitle', '汇总删除预览提示、执行提示与异常信息。')
    : t('codegen.messages_subtitle', '包含 dry-run 提示、生成摘要和校验信息。'));
const messagePanelEmptyText = computed(() => (activeMode.value === 'delete' ? t('codegen.no_delete_messages', '暂无删除提示') : t('codegen.no_messages', '暂无消息')));
const deleteExecuteEnabled = computed(() => activeMode.value === 'delete' && deletePreviewReport.value !== null);
const dbMountMenuLabel = computed(() => {
    if (!dbMountParentPath.value) {
        return t('codegen.db_mount_root_top_label', '顶层根菜单');
    }
    const option = dbMountMenuOptions.value.find((item) => item.value === dbMountParentPath.value);
    return option?.label || dbMountParentPath.value;
});
const previewItems = computed(() => {
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
    const map = {
        auth: t('codegen.error_auth', '登录失效'),
        notfound: t('codegen.error_notfound', '资源不存在'),
        expired: t('codegen.error_expired', '已过期'),
        server: t('codegen.error_server', '服务异常'),
        unknown: t('codegen.error_unknown', '其他'),
    };
    return map[lastArtifactErrorType.value] || t('codegen.error_unknown', '其他');
});
const artifactLastErrorTypeTag = computed(() => {
    const map = {
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
function applyDbPreset(driver) {
    const presets = {
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
        await ElMessageBox.confirm(t('codegen.delete_execute_confirm', '即将对模块 {module} 执行删除，共 {total} 项。当前方案包含 {warnings} 条提示和 {conflicts} 个冲突。确认后将调用后端删除执行接口。', {
            module: preview.plan.module || deleteModule.value,
            total,
            warnings,
            conflicts,
        }), t('codegen.confirm_delete_title', '确认删除方案'), {
            confirmButtonText: t('codegen.confirm_execute', '确认执行'),
            cancelButtonText: t('codegen.return_modify', '返回修改'),
            type: 'warning',
            distinguishCancelAndClose: true,
        });
    }
    catch (error) {
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
        };
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
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('codegen.delete_failed', '删除执行失败'));
    }
    finally {
        deleteLoading.value = false;
    }
}
async function loadPublicConfig() {
    try {
        publicConfig.value = await fetchPublicConfig();
        applyDbConfigDefaults();
    }
    catch {
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
async function handleFileChange(event) {
    const input = event.target;
    const file = input?.files?.[0];
    if (!file) {
        return;
    }
    try {
        const content = await file.text();
        dslText.value = content;
        ElMessage.success(t('codegen.file_loaded', '已载入 {name}', { name: file.name }));
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('codegen.file_read_failed', '读取 DSL 文件失败'));
    }
    finally {
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
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('codegen.dry_run_failed', 'Dry-run 预览失败'));
    }
    finally {
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
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('codegen.generate_failed', '生成失败'));
    }
    finally {
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
        await ElMessageBox.confirm(t('codegen.db_install_prompt', '即将把模块 {module} 的 manifest 安装到系统菜单中，是否继续？', {
            module: dbGeneratedModuleName.value,
        }), t('codegen.install_confirm_title', '确认安装到系统'), {
            confirmButtonText: t('codegen.install_confirm_continue', '继续安装'),
            cancelButtonText: t('common.cancel', '取消'),
            type: 'warning',
            distinguishCancelAndClose: true,
        });
        installLoading.value = true;
        const result = await installCodegenManifest({ module: dbGeneratedModuleName.value });
        lastRunSuccess.value = true;
        operationStatus.value = t('codegen.install_result_summary', '模块 {module} 已安装到系统，共 {total} 个菜单', {
            module: result.module || dbGeneratedModuleName.value,
            total: result.menu_total,
        });
        ElMessage.success(t('codegen.db_install_complete', '安装到系统完成'));
        await loadDbMountMenuOptions();
    }
    catch (error) {
        if (error === 'cancel' || error === 'close') {
            ElMessage.info(t('codegen.install_cancelled', '已取消安装'));
            return;
        }
        ElMessage.error(error instanceof Error ? error.message : t('codegen.generate_install_failed', '生成并安装失败'));
    }
    finally {
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
    }
    catch (error) {
        deletePreviewReport.value = null;
        deleteRequestCache.value = null;
        deleteResult.value = null;
        ElMessage.error(error instanceof Error ? error.message : t('codegen.delete_preview_failed', '删除预览失败'));
    }
    finally {
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
    }
    else if (!dslText.value.trim()) {
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
    }
    catch (error) {
        handleArtifactError(error, isDbMode ? t('codegen.generate_database_download_failed', '生成数据库下载包失败') : t('codegen.generate_download_failed', '生成下载包失败'));
    }
    finally {
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
    }
    catch (error) {
        handleArtifactError(error, t('codegen.download_failed', '下载失败'));
    }
    finally {
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
    }
    catch (error) {
        resetCopyFeedback();
        ElMessage.error(error instanceof Error ? error.message : t('codegen.copy_failed', '复制下载地址失败'));
    }
}
function handleArtifactError(error, fallbackMessage) {
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
function buildDatabaseRequest() {
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
    }
    catch {
        dbMountMenuOptions.value = [];
    }
}
function flattenMenuMountOptions(items, depth = 0) {
    const options = [];
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
function validateDatabaseInputs() {
    if (!dbDriver.value.trim()) {
        return t('codegen.db_validate_driver', '请先选择数据库驱动');
    }
    if (!dbDatabase.value.trim()) {
        return t('codegen.db_validate_name', '请先填写数据库名');
    }
    return '';
}
function validateDeleteInputs() {
    if (!deleteModule.value.trim()) {
        return t('codegen.delete_validate_module', '请先填写要删除的模块名');
    }
    return '';
}
function buildDeleteRequest(dryRun) {
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
function parseTableNames(value) {
    return value
        .split(/[\n,]/)
        .map((item) => item.trim())
        .filter((item) => item.length > 0);
}
function mapDatabaseResourcesToPreviewRows(resources, forceValue) {
    return resources.map((resource, index) => ({
        index: index + 1,
        kind: resource.kind || 'resource',
        name: resource.name || resource.entity_name || resource.table_name,
        force: forceValue,
        actions: resource.actions ?? [],
    }));
}
function mapDeletePlanToPreviewRows(plan) {
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
function buildDeleteItemActions(item) {
    const actions = [];
    if (item.origin) {
        actions.push(`origin:${item.origin}`);
    }
    actions.push(item.managed ? 'managed' : 'manual');
    if (item.store) {
        actions.push(`store:${item.store}`);
    }
    return actions;
}
function describeDeleteItem(item) {
    const parts = [item.kind || 'asset'];
    if (item.path) {
        parts.push(item.path);
    }
    else if (item.ref) {
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
function deleteItemKey(item) {
    return [item.kind, item.path, item.ref, item.origin].filter((value) => Boolean(value)).join('::');
}
function describeDeleteFailure(failure) {
    const parts = [failure.reason || '删除失败'];
    if (failure.item) {
        parts.push(describeDeleteItem(failure.item));
    }
    return parts.join(' · ');
}
function describeDeleteFailureKey(failure) {
    return `${failure.reason || 'failure'}::${failure.item ? deleteItemKey(failure.item) : 'none'}`;
}
function formatElapsedMillis(value) {
    if (!Number.isFinite(value) || value <= 0) {
        return '-';
    }
    if (value < 1000) {
        return `${Math.round(value)}ms`;
    }
    const seconds = value / 1000;
    return `${seconds.toFixed(seconds >= 10 ? 1 : 2)}s`;
}
function rememberArtifactError(message) {
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
function normalizeHttpStatus(code) {
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
async function copyText(value) {
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
function toAbsoluteUrl(value) {
    const path = value.trim();
    if (!path) {
        return '';
    }
    if (/^https?:\/\//i.test(path)) {
        return path;
    }
    return new URL(path, window.location.origin).toString();
}
function summarizeDownloadUrl(value) {
    if (!value) {
        return '';
    }
    try {
        const parsed = new URL(value);
        return `${parsed.host} · ${summarizePath(parsed.pathname)}${parsed.search}`;
    }
    catch {
        return value;
    }
}
function summarizePath(value) {
    if (!value) {
        return '/';
    }
    const normalized = value.length > 48 ? `${value.slice(0, 24)}...${value.slice(-18)}` : value;
    return normalized;
}
function formatBytes(value) {
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
function formatDateTime(value) {
    if (!value) {
        return '-';
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return value;
    }
    return date.toLocaleString();
}
function formatRemainingTime(value, now, expired) {
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
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['codegen-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['el-col']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-field-help']} */ ;
/** @type {__VLS_StyleScopedClasses['db-advanced']} */ ;
/** @type {__VLS_StyleScopedClasses['side-stack']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "codegen-page" },
});
const __VLS_0 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    gutter: (20),
    ...{ class: "codegen-grid" },
}));
const __VLS_2 = __VLS_1({
    gutter: (20),
    ...{ class: "codegen-grid" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    xs: (24),
    lg: (11),
    xl: (11),
}));
const __VLS_6 = __VLS_5({
    xs: (24),
    lg: (11),
    xl: (11),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
const __VLS_8 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    shadow: "never",
    ...{ class: "codegen-card" },
}));
const __VLS_10 = __VLS_9({
    shadow: "never",
    ...{ class: "codegen-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_11.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "title" },
    });
    (__VLS_ctx.t('codegen.console_title', 'CodeGen Console'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "subtitle" },
    });
    (__VLS_ctx.t('codegen.console_subtitle', '在同一页面中切换 DSL、DB 与删除模式，复用统一结果区。'));
    const __VLS_12 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
        wrap: true,
    }));
    const __VLS_14 = __VLS_13({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
    __VLS_15.slots.default;
    if (__VLS_ctx.activeMode === 'dsl') {
        const __VLS_16 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
            ...{ 'onClick': {} },
        }));
        const __VLS_18 = __VLS_17({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_17));
        let __VLS_20;
        let __VLS_21;
        let __VLS_22;
        const __VLS_23 = {
            onClick: (__VLS_ctx.loadSample)
        };
        __VLS_19.slots.default;
        (__VLS_ctx.t('codegen.load_sample', '载入示例'));
        var __VLS_19;
    }
    else if (__VLS_ctx.activeMode === 'db') {
        const __VLS_24 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            ...{ 'onClick': {} },
        }));
        const __VLS_26 = __VLS_25({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        let __VLS_28;
        let __VLS_29;
        let __VLS_30;
        const __VLS_31 = {
            onClick: (__VLS_ctx.loadDbSample)
        };
        __VLS_27.slots.default;
        (__VLS_ctx.t('codegen.load_sample', '载入示例'));
        var __VLS_27;
    }
    else {
        const __VLS_32 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            ...{ 'onClick': {} },
        }));
        const __VLS_34 = __VLS_33({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        let __VLS_36;
        let __VLS_37;
        let __VLS_38;
        const __VLS_39 = {
            onClick: (__VLS_ctx.loadDeleteSample)
        };
        __VLS_35.slots.default;
        (__VLS_ctx.t('codegen.load_sample', '载入示例'));
        var __VLS_35;
    }
    const __VLS_40 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
        ...{ 'onClick': {} },
    }));
    const __VLS_42 = __VLS_41({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_41));
    let __VLS_44;
    let __VLS_45;
    let __VLS_46;
    const __VLS_47 = {
        onClick: (__VLS_ctx.clearCurrentInputs)
    };
    __VLS_43.slots.default;
    (__VLS_ctx.t('codegen.clear', '清空'));
    var __VLS_43;
    if (__VLS_ctx.activeMode === 'dsl') {
        const __VLS_48 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
            ...{ 'onClick': {} },
        }));
        const __VLS_50 = __VLS_49({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_49));
        let __VLS_52;
        let __VLS_53;
        let __VLS_54;
        const __VLS_55 = {
            onClick: (__VLS_ctx.triggerFileSelect)
        };
        __VLS_51.slots.default;
        (__VLS_ctx.t('codegen.upload_dsl', '上传 DSL'));
        var __VLS_51;
    }
    if (__VLS_ctx.activeMode === 'dsl') {
        const __VLS_56 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }));
        const __VLS_58 = __VLS_57({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_57));
        let __VLS_60;
        let __VLS_61;
        let __VLS_62;
        const __VLS_63 = {
            onClick: (__VLS_ctx.handlePreview)
        };
        __VLS_59.slots.default;
        (__VLS_ctx.t('codegen.preview_dry_run', 'Dry-run 预览'));
        var __VLS_59;
    }
    if (__VLS_ctx.activeMode === 'dsl') {
        const __VLS_64 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
            ...{ 'onClick': {} },
            type: "success",
            loading: (__VLS_ctx.generateLoading),
        }));
        const __VLS_66 = __VLS_65({
            ...{ 'onClick': {} },
            type: "success",
            loading: (__VLS_ctx.generateLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_65));
        let __VLS_68;
        let __VLS_69;
        let __VLS_70;
        const __VLS_71 = {
            onClick: (__VLS_ctx.handleGenerate)
        };
        __VLS_67.slots.default;
        (__VLS_ctx.t('codegen.generate_once', '一键生成'));
        var __VLS_67;
    }
    if (__VLS_ctx.activeMode === 'dsl') {
        const __VLS_72 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
            ...{ 'onClick': {} },
            type: "warning",
            loading: (__VLS_ctx.downloadLoading),
        }));
        const __VLS_74 = __VLS_73({
            ...{ 'onClick': {} },
            type: "warning",
            loading: (__VLS_ctx.downloadLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_73));
        let __VLS_76;
        let __VLS_77;
        let __VLS_78;
        const __VLS_79 = {
            onClick: (__VLS_ctx.handleGenerateDownload)
        };
        __VLS_75.slots.default;
        (__VLS_ctx.t('codegen.generate_download', '生成并下载'));
        var __VLS_75;
    }
    if (__VLS_ctx.activeMode === 'db') {
        const __VLS_80 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }));
        const __VLS_82 = __VLS_81({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_81));
        let __VLS_84;
        let __VLS_85;
        let __VLS_86;
        const __VLS_87 = {
            onClick: (__VLS_ctx.handlePreview)
        };
        __VLS_83.slots.default;
        (__VLS_ctx.t('codegen.preview_dry_run', 'Dry-run 预览'));
        var __VLS_83;
    }
    if (__VLS_ctx.activeMode === 'db') {
        const __VLS_88 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
            ...{ 'onClick': {} },
            type: "success",
            loading: (__VLS_ctx.generateLoading || __VLS_ctx.installLoading),
        }));
        const __VLS_90 = __VLS_89({
            ...{ 'onClick': {} },
            type: "success",
            loading: (__VLS_ctx.generateLoading || __VLS_ctx.installLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_89));
        let __VLS_92;
        let __VLS_93;
        let __VLS_94;
        const __VLS_95 = {
            onClick: (__VLS_ctx.handleGenerateAndInstall)
        };
        __VLS_91.slots.default;
        (__VLS_ctx.t('codegen.generate_install', '生成并安装'));
        var __VLS_91;
    }
    if (__VLS_ctx.activeMode === 'db') {
        const __VLS_96 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
            ...{ 'onClick': {} },
            type: "warning",
            loading: (__VLS_ctx.downloadLoading),
        }));
        const __VLS_98 = __VLS_97({
            ...{ 'onClick': {} },
            type: "warning",
            loading: (__VLS_ctx.downloadLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_97));
        let __VLS_100;
        let __VLS_101;
        let __VLS_102;
        const __VLS_103 = {
            onClick: (__VLS_ctx.handleGenerateDownload)
        };
        __VLS_99.slots.default;
        (__VLS_ctx.t('codegen.generate_download', '生成并下载'));
        var __VLS_99;
    }
    if (__VLS_ctx.activeMode === 'delete') {
        const __VLS_104 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
            ...{ 'onClick': {} },
        }));
        const __VLS_106 = __VLS_105({
            ...{ 'onClick': {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_105));
        let __VLS_108;
        let __VLS_109;
        let __VLS_110;
        const __VLS_111 = {
            onClick: (__VLS_ctx.loadDeleteSample)
        };
        __VLS_107.slots.default;
        (__VLS_ctx.t('codegen.load_sample', '载入示例'));
        var __VLS_107;
    }
    if (__VLS_ctx.activeMode === 'delete') {
        const __VLS_112 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }));
        const __VLS_114 = __VLS_113({
            ...{ 'onClick': {} },
            type: "primary",
            loading: (__VLS_ctx.previewLoading),
        }, ...__VLS_functionalComponentArgsRest(__VLS_113));
        let __VLS_116;
        let __VLS_117;
        let __VLS_118;
        const __VLS_119 = {
            onClick: (__VLS_ctx.handleDeletePreview)
        };
        __VLS_115.slots.default;
        (__VLS_ctx.t('codegen.delete_preview', '删除预览'));
        var __VLS_115;
    }
    if (__VLS_ctx.activeMode === 'delete') {
        const __VLS_120 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
            ...{ 'onClick': {} },
            type: "danger",
            loading: (__VLS_ctx.deleteLoading),
            disabled: (!__VLS_ctx.deleteExecuteEnabled),
        }));
        const __VLS_122 = __VLS_121({
            ...{ 'onClick': {} },
            type: "danger",
            loading: (__VLS_ctx.deleteLoading),
            disabled: (!__VLS_ctx.deleteExecuteEnabled),
        }, ...__VLS_functionalComponentArgsRest(__VLS_121));
        let __VLS_124;
        let __VLS_125;
        let __VLS_126;
        const __VLS_127 = {
            onClick: (__VLS_ctx.handleDeleteExecute)
        };
        __VLS_123.slots.default;
        (__VLS_ctx.t('codegen.confirm_delete', '确认删除'));
        var __VLS_123;
    }
    var __VLS_15;
}
const __VLS_128 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
    modelValue: (__VLS_ctx.activeMode),
    ...{ class: "codegen-tabs" },
    stretch: true,
}));
const __VLS_130 = __VLS_129({
    modelValue: (__VLS_ctx.activeMode),
    ...{ class: "codegen-tabs" },
    stretch: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_129));
__VLS_131.slots.default;
const __VLS_132 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
    label: (__VLS_ctx.t('codegen.mode.dsl', 'DSL')),
    name: "dsl",
}));
const __VLS_134 = __VLS_133({
    label: (__VLS_ctx.t('codegen.mode.dsl', 'DSL')),
    name: "dsl",
}, ...__VLS_functionalComponentArgsRest(__VLS_133));
__VLS_135.slots.default;
const __VLS_136 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
    labelPosition: "top",
    ...{ class: "codegen-form" },
}));
const __VLS_138 = __VLS_137({
    labelPosition: "top",
    ...{ class: "codegen-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_137));
__VLS_139.slots.default;
const __VLS_140 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_141 = __VLS_asFunctionalComponent(__VLS_140, new __VLS_140({
    label: (__VLS_ctx.t('codegen.force_overwrite', '强制覆盖')),
}));
const __VLS_142 = __VLS_141({
    label: (__VLS_ctx.t('codegen.force_overwrite', '强制覆盖')),
}, ...__VLS_functionalComponentArgsRest(__VLS_141));
__VLS_143.slots.default;
const __VLS_144 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_145 = __VLS_asFunctionalComponent(__VLS_144, new __VLS_144({
    modelValue: (__VLS_ctx.force),
    inlinePrompt: true,
    activeText: (__VLS_ctx.t('common.on', 'On')),
    inactiveText: (__VLS_ctx.t('common.off', 'Off')),
}));
const __VLS_146 = __VLS_145({
    modelValue: (__VLS_ctx.force),
    inlinePrompt: true,
    activeText: (__VLS_ctx.t('common.on', 'On')),
    inactiveText: (__VLS_ctx.t('common.off', 'Off')),
}, ...__VLS_functionalComponentArgsRest(__VLS_145));
var __VLS_143;
const __VLS_148 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_149 = __VLS_asFunctionalComponent(__VLS_148, new __VLS_148({
    label: (__VLS_ctx.t('codegen.package_name', '下载包名称')),
}));
const __VLS_150 = __VLS_149({
    label: (__VLS_ctx.t('codegen.package_name', '下载包名称')),
}, ...__VLS_functionalComponentArgsRest(__VLS_149));
__VLS_151.slots.default;
const __VLS_152 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_153 = __VLS_asFunctionalComponent(__VLS_152, new __VLS_152({
    modelValue: (__VLS_ctx.packageName),
    placeholder: (__VLS_ctx.t('codegen.package_name_placeholder', '留空则由系统自动生成 zip 名称')),
}));
const __VLS_154 = __VLS_153({
    modelValue: (__VLS_ctx.packageName),
    placeholder: (__VLS_ctx.t('codegen.package_name_placeholder', '留空则由系统自动生成 zip 名称')),
}, ...__VLS_functionalComponentArgsRest(__VLS_153));
var __VLS_151;
const __VLS_156 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_157 = __VLS_asFunctionalComponent(__VLS_156, new __VLS_156({
    label: (__VLS_ctx.t('codegen.package_content', '下载包内容')),
}));
const __VLS_158 = __VLS_157({
    label: (__VLS_ctx.t('codegen.package_content', '下载包内容')),
}, ...__VLS_functionalComponentArgsRest(__VLS_157));
__VLS_159.slots.default;
const __VLS_160 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_161 = __VLS_asFunctionalComponent(__VLS_160, new __VLS_160({
    wrap: true,
}));
const __VLS_162 = __VLS_161({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_161));
__VLS_163.slots.default;
const __VLS_164 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    modelValue: (__VLS_ctx.includeReadme),
    inlinePrompt: true,
    activeText: "README",
    inactiveText: "README",
}));
const __VLS_166 = __VLS_165({
    modelValue: (__VLS_ctx.includeReadme),
    inlinePrompt: true,
    activeText: "README",
    inactiveText: "README",
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
const __VLS_168 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
    modelValue: (__VLS_ctx.includeReport),
    inlinePrompt: true,
    activeText: "Report",
    inactiveText: "Report",
}));
const __VLS_170 = __VLS_169({
    modelValue: (__VLS_ctx.includeReport),
    inlinePrompt: true,
    activeText: "Report",
    inactiveText: "Report",
}, ...__VLS_functionalComponentArgsRest(__VLS_169));
const __VLS_172 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
    modelValue: (__VLS_ctx.includeDsl),
    inlinePrompt: true,
    activeText: "DSL",
    inactiveText: "DSL",
}));
const __VLS_174 = __VLS_173({
    modelValue: (__VLS_ctx.includeDsl),
    inlinePrompt: true,
    activeText: "DSL",
    inactiveText: "DSL",
}, ...__VLS_functionalComponentArgsRest(__VLS_173));
var __VLS_163;
var __VLS_159;
const __VLS_176 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
    label: (__VLS_ctx.t('codegen.dsl_content', 'DSL 内容')),
}));
const __VLS_178 = __VLS_177({
    label: (__VLS_ctx.t('codegen.dsl_content', 'DSL 内容')),
}, ...__VLS_functionalComponentArgsRest(__VLS_177));
__VLS_179.slots.default;
const __VLS_180 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({
    modelValue: (__VLS_ctx.dslText),
    type: "textarea",
    rows: (28),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.dsl_placeholder', '在这里粘贴或编辑 DSL YAML')),
}));
const __VLS_182 = __VLS_181({
    modelValue: (__VLS_ctx.dslText),
    type: "textarea",
    rows: (28),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.dsl_placeholder', '在这里粘贴或编辑 DSL YAML')),
}, ...__VLS_functionalComponentArgsRest(__VLS_181));
var __VLS_179;
var __VLS_139;
__VLS_asFunctionalElement(__VLS_intrinsicElements.input)({
    ...{ onChange: (__VLS_ctx.handleFileChange) },
    ref: "fileInputRef",
    ...{ class: "hidden-file-input" },
    type: "file",
    accept: ".yaml,.yml,.json,.txt",
});
/** @type {typeof __VLS_ctx.fileInputRef} */ ;
var __VLS_135;
const __VLS_184 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
    label: (__VLS_ctx.t('codegen.mode.db', 'DB')),
    name: "db",
}));
const __VLS_186 = __VLS_185({
    label: (__VLS_ctx.t('codegen.mode.db', 'DB')),
    name: "db",
}, ...__VLS_functionalComponentArgsRest(__VLS_185));
__VLS_187.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-mode-panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-hero" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-hero-title" },
});
(__VLS_ctx.t('codegen.db_guide_title', '数据库输入向导'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-hero-subtitle" },
});
(__VLS_ctx.t('codegen.db_guide_subtitle', '先选驱动，再填连接串与扫描范围。建议先预览，确认结果后再生成。'));
const __VLS_188 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_189 = __VLS_asFunctionalComponent(__VLS_188, new __VLS_188({
    wrap: true,
}));
const __VLS_190 = __VLS_189({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_189));
__VLS_191.slots.default;
const __VLS_192 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
    type: "info",
    effect: "light",
}));
const __VLS_194 = __VLS_193({
    type: "info",
    effect: "light",
}, ...__VLS_functionalComponentArgsRest(__VLS_193));
__VLS_195.slots.default;
(__VLS_ctx.dbDriverLabel);
var __VLS_195;
const __VLS_196 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({
    type: "success",
    effect: "light",
}));
const __VLS_198 = __VLS_197({
    type: "success",
    effect: "light",
}, ...__VLS_functionalComponentArgsRest(__VLS_197));
__VLS_199.slots.default;
(__VLS_ctx.dbParsedTables.length ? `${__VLS_ctx.dbParsedTables.length} 个表` : __VLS_ctx.t('common.all', '全部表'));
var __VLS_199;
var __VLS_191;
const __VLS_200 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({
    title: (__VLS_ctx.t('codegen.db_recommended_preview', '推荐先执行 Dry-run 预览，确认文件计划和冲突后再执行生成。')),
    type: "info",
    closable: (false),
    showIcon: true,
}));
const __VLS_202 = __VLS_201({
    title: (__VLS_ctx.t('codegen.db_recommended_preview', '推荐先执行 Dry-run 预览，确认文件计划和冲突后再执行生成。')),
    type: "info",
    closable: (false),
    showIcon: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_201));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-preset-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-section-title" },
});
(__VLS_ctx.t('codegen.db_fast_template', '快速模板'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-section-hint" },
});
(__VLS_ctx.t('codegen.db_fast_template_hint', '一键预填常见数据库的连接格式。'));
const __VLS_204 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
    wrap: true,
}));
const __VLS_206 = __VLS_205({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_205));
__VLS_207.slots.default;
const __VLS_208 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_209 = __VLS_asFunctionalComponent(__VLS_208, new __VLS_208({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'mysql' ? 'primary' : 'default'),
    plain: true,
}));
const __VLS_210 = __VLS_209({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'mysql' ? 'primary' : 'default'),
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_209));
let __VLS_212;
let __VLS_213;
let __VLS_214;
const __VLS_215 = {
    onClick: (...[$event]) => {
        __VLS_ctx.applyDbPreset('mysql');
    }
};
__VLS_211.slots.default;
var __VLS_211;
const __VLS_216 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_217 = __VLS_asFunctionalComponent(__VLS_216, new __VLS_216({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'postgres' ? 'primary' : 'default'),
    plain: true,
}));
const __VLS_218 = __VLS_217({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'postgres' ? 'primary' : 'default'),
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_217));
let __VLS_220;
let __VLS_221;
let __VLS_222;
const __VLS_223 = {
    onClick: (...[$event]) => {
        __VLS_ctx.applyDbPreset('postgres');
    }
};
__VLS_219.slots.default;
var __VLS_219;
const __VLS_224 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'sqlite' ? 'primary' : 'default'),
    plain: true,
}));
const __VLS_226 = __VLS_225({
    ...{ 'onClick': {} },
    size: "small",
    type: (__VLS_ctx.dbDriver === 'sqlite' ? 'primary' : 'default'),
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_225));
let __VLS_228;
let __VLS_229;
let __VLS_230;
const __VLS_231 = {
    onClick: (...[$event]) => {
        __VLS_ctx.applyDbPreset('sqlite');
    }
};
__VLS_227.slots.default;
var __VLS_227;
var __VLS_207;
const __VLS_232 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({
    labelPosition: "top",
    ...{ class: "codegen-form db-form" },
}));
const __VLS_234 = __VLS_233({
    labelPosition: "top",
    ...{ class: "codegen-form db-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_233));
__VLS_235.slots.default;
const __VLS_236 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
    gutter: (16),
}));
const __VLS_238 = __VLS_237({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_237));
__VLS_239.slots.default;
const __VLS_240 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_241 = __VLS_asFunctionalComponent(__VLS_240, new __VLS_240({
    xs: (24),
    md: (6),
}));
const __VLS_242 = __VLS_241({
    xs: (24),
    md: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_241));
__VLS_243.slots.default;
const __VLS_244 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({
    label: (__VLS_ctx.t('codegen.db_driver', '数据库驱动')),
}));
const __VLS_246 = __VLS_245({
    label: (__VLS_ctx.t('codegen.db_driver', '数据库驱动')),
}, ...__VLS_functionalComponentArgsRest(__VLS_245));
__VLS_247.slots.default;
const __VLS_248 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_249 = __VLS_asFunctionalComponent(__VLS_248, new __VLS_248({
    modelValue: (__VLS_ctx.dbDriver),
    placeholder: (__VLS_ctx.t('codegen.db_driver_placeholder', '请选择数据库驱动')),
    filterable: true,
}));
const __VLS_250 = __VLS_249({
    modelValue: (__VLS_ctx.dbDriver),
    placeholder: (__VLS_ctx.t('codegen.db_driver_placeholder', '请选择数据库驱动')),
    filterable: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_249));
__VLS_251.slots.default;
const __VLS_252 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_253 = __VLS_asFunctionalComponent(__VLS_252, new __VLS_252({
    label: "MySQL",
    value: "mysql",
}));
const __VLS_254 = __VLS_253({
    label: "MySQL",
    value: "mysql",
}, ...__VLS_functionalComponentArgsRest(__VLS_253));
const __VLS_256 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
    label: "PostgreSQL",
    value: "postgres",
}));
const __VLS_258 = __VLS_257({
    label: "PostgreSQL",
    value: "postgres",
}, ...__VLS_functionalComponentArgsRest(__VLS_257));
const __VLS_260 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_261 = __VLS_asFunctionalComponent(__VLS_260, new __VLS_260({
    label: "SQLite",
    value: "sqlite",
}));
const __VLS_262 = __VLS_261({
    label: "SQLite",
    value: "sqlite",
}, ...__VLS_functionalComponentArgsRest(__VLS_261));
var __VLS_251;
var __VLS_247;
var __VLS_243;
const __VLS_264 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_265 = __VLS_asFunctionalComponent(__VLS_264, new __VLS_264({
    xs: (24),
    md: (6),
}));
const __VLS_266 = __VLS_265({
    xs: (24),
    md: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_265));
__VLS_267.slots.default;
const __VLS_268 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_269 = __VLS_asFunctionalComponent(__VLS_268, new __VLS_268({
    label: (__VLS_ctx.t('codegen.db_name', '数据库名')),
}));
const __VLS_270 = __VLS_269({
    label: (__VLS_ctx.t('codegen.db_name', '数据库名')),
}, ...__VLS_functionalComponentArgsRest(__VLS_269));
__VLS_271.slots.default;
const __VLS_272 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_273 = __VLS_asFunctionalComponent(__VLS_272, new __VLS_272({
    modelValue: (__VLS_ctx.dbDatabase),
    placeholder: (__VLS_ctx.t('codegen.db_name_placeholder', '请输入数据库名称')),
}));
const __VLS_274 = __VLS_273({
    modelValue: (__VLS_ctx.dbDatabase),
    placeholder: (__VLS_ctx.t('codegen.db_name_placeholder', '请输入数据库名称')),
}, ...__VLS_functionalComponentArgsRest(__VLS_273));
var __VLS_271;
var __VLS_267;
const __VLS_276 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_277 = __VLS_asFunctionalComponent(__VLS_276, new __VLS_276({
    xs: (24),
    md: (6),
}));
const __VLS_278 = __VLS_277({
    xs: (24),
    md: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_277));
__VLS_279.slots.default;
const __VLS_280 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_281 = __VLS_asFunctionalComponent(__VLS_280, new __VLS_280({
    label: (__VLS_ctx.t('codegen.db_schema', 'Schema')),
}));
const __VLS_282 = __VLS_281({
    label: (__VLS_ctx.t('codegen.db_schema', 'Schema')),
}, ...__VLS_functionalComponentArgsRest(__VLS_281));
__VLS_283.slots.default;
const __VLS_284 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_285 = __VLS_asFunctionalComponent(__VLS_284, new __VLS_284({
    modelValue: (__VLS_ctx.dbSchema),
    placeholder: (__VLS_ctx.t('codegen.db_schema_placeholder', '可选，PostgreSQL 等场景使用')),
}));
const __VLS_286 = __VLS_285({
    modelValue: (__VLS_ctx.dbSchema),
    placeholder: (__VLS_ctx.t('codegen.db_schema_placeholder', '可选，PostgreSQL 等场景使用')),
}, ...__VLS_functionalComponentArgsRest(__VLS_285));
var __VLS_283;
var __VLS_279;
const __VLS_288 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_289 = __VLS_asFunctionalComponent(__VLS_288, new __VLS_288({
    xs: (24),
    md: (6),
}));
const __VLS_290 = __VLS_289({
    xs: (24),
    md: (6),
}, ...__VLS_functionalComponentArgsRest(__VLS_289));
__VLS_291.slots.default;
const __VLS_292 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_293 = __VLS_asFunctionalComponent(__VLS_292, new __VLS_292({
    label: (__VLS_ctx.t('codegen.db_mount_root', '挂载根菜单')),
}));
const __VLS_294 = __VLS_293({
    label: (__VLS_ctx.t('codegen.db_mount_root', '挂载根菜单')),
}, ...__VLS_functionalComponentArgsRest(__VLS_293));
__VLS_295.slots.default;
const __VLS_296 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_297 = __VLS_asFunctionalComponent(__VLS_296, new __VLS_296({
    modelValue: (__VLS_ctx.dbMountParentPath),
    clearable: true,
    filterable: true,
    placeholder: (__VLS_ctx.t('codegen.db_mount_root_placeholder', '留空为顶层根菜单')),
}));
const __VLS_298 = __VLS_297({
    modelValue: (__VLS_ctx.dbMountParentPath),
    clearable: true,
    filterable: true,
    placeholder: (__VLS_ctx.t('codegen.db_mount_root_placeholder', '留空为顶层根菜单')),
}, ...__VLS_functionalComponentArgsRest(__VLS_297));
__VLS_299.slots.default;
const __VLS_300 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_301 = __VLS_asFunctionalComponent(__VLS_300, new __VLS_300({
    label: (__VLS_ctx.t('codegen.db_mount_root_top', '顶层根菜单')),
    value: "",
}));
const __VLS_302 = __VLS_301({
    label: (__VLS_ctx.t('codegen.db_mount_root_top', '顶层根菜单')),
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_301));
for (const [option] of __VLS_getVForSourceType((__VLS_ctx.dbMountMenuOptions))) {
    const __VLS_304 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_305 = __VLS_asFunctionalComponent(__VLS_304, new __VLS_304({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }));
    const __VLS_306 = __VLS_305({
        key: (option.value),
        label: (option.label),
        value: (option.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_305));
}
var __VLS_299;
var __VLS_295;
var __VLS_291;
var __VLS_239;
const __VLS_308 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_309 = __VLS_asFunctionalComponent(__VLS_308, new __VLS_308({
    label: (__VLS_ctx.t('codegen.db_table_range', '表名范围')),
    ...{ class: "db-form-item db-form-item--wide" },
}));
const __VLS_310 = __VLS_309({
    label: (__VLS_ctx.t('codegen.db_table_range', '表名范围')),
    ...{ class: "db-form-item db-form-item--wide" },
}, ...__VLS_functionalComponentArgsRest(__VLS_309));
__VLS_311.slots.default;
const __VLS_312 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_313 = __VLS_asFunctionalComponent(__VLS_312, new __VLS_312({
    modelValue: (__VLS_ctx.dbTablesText),
    type: "textarea",
    rows: (6),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.db_table_range_placeholder', '支持逗号、换行分隔，例如：books, orders')),
}));
const __VLS_314 = __VLS_313({
    modelValue: (__VLS_ctx.dbTablesText),
    type: "textarea",
    rows: (6),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.db_table_range_placeholder', '支持逗号、换行分隔，例如：books, orders')),
}, ...__VLS_functionalComponentArgsRest(__VLS_313));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-form-row" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "db-field-help" },
});
(__VLS_ctx.t('codegen.db_table_range_help', '留空则表示扫描全部表；建议优先预填少量表进行预览。'));
const __VLS_316 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_317 = __VLS_asFunctionalComponent(__VLS_316, new __VLS_316({
    wrap: true,
}));
const __VLS_318 = __VLS_317({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_317));
__VLS_319.slots.default;
const __VLS_320 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_321 = __VLS_asFunctionalComponent(__VLS_320, new __VLS_320({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}));
const __VLS_322 = __VLS_321({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_321));
let __VLS_324;
let __VLS_325;
let __VLS_326;
const __VLS_327 = {
    onClick: (__VLS_ctx.loadDbSample)
};
__VLS_323.slots.default;
(__VLS_ctx.t('codegen.load_sample', '载入示例'));
var __VLS_323;
const __VLS_328 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_329 = __VLS_asFunctionalComponent(__VLS_328, new __VLS_328({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}));
const __VLS_330 = __VLS_329({
    ...{ 'onClick': {} },
    text: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_329));
let __VLS_332;
let __VLS_333;
let __VLS_334;
const __VLS_335 = {
    onClick: (__VLS_ctx.clearDbTables)
};
__VLS_331.slots.default;
(__VLS_ctx.t('codegen.clear', '清空'));
var __VLS_331;
var __VLS_319;
if (__VLS_ctx.dbParsedTables.length) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "db-table-chip-list" },
    });
    for (const [table] of __VLS_getVForSourceType((__VLS_ctx.dbParsedTables))) {
        const __VLS_336 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_337 = __VLS_asFunctionalComponent(__VLS_336, new __VLS_336({
            key: (table),
            size: "small",
            effect: "plain",
            ...{ class: "db-table-chip" },
        }));
        const __VLS_338 = __VLS_337({
            key: (table),
            size: "small",
            effect: "plain",
            ...{ class: "db-table-chip" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_337));
        __VLS_339.slots.default;
        (table);
        var __VLS_339;
    }
}
var __VLS_311;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-advanced" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-section-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-section-title" },
});
(__VLS_ctx.t('codegen.generate_options', '生成选项'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "db-section-hint" },
});
(__VLS_ctx.t('codegen.generate_options_hint', '控制是否覆盖现有文件、是否输出前端和权限策略。'));
const __VLS_340 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_341 = __VLS_asFunctionalComponent(__VLS_340, new __VLS_340({
    size: "small",
    type: "success",
    effect: "light",
}));
const __VLS_342 = __VLS_341({
    size: "small",
    type: "success",
    effect: "light",
}, ...__VLS_functionalComponentArgsRest(__VLS_341));
__VLS_343.slots.default;
(__VLS_ctx.dbOptionSummary);
var __VLS_343;
const __VLS_344 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_345 = __VLS_asFunctionalComponent(__VLS_344, new __VLS_344({
    wrap: true,
}));
const __VLS_346 = __VLS_345({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_345));
__VLS_347.slots.default;
const __VLS_348 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_349 = __VLS_asFunctionalComponent(__VLS_348, new __VLS_348({
    modelValue: (__VLS_ctx.dbForce),
    inlinePrompt: true,
    activeText: "Force",
    inactiveText: "Force",
}));
const __VLS_350 = __VLS_349({
    modelValue: (__VLS_ctx.dbForce),
    inlinePrompt: true,
    activeText: "Force",
    inactiveText: "Force",
}, ...__VLS_functionalComponentArgsRest(__VLS_349));
const __VLS_352 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_353 = __VLS_asFunctionalComponent(__VLS_352, new __VLS_352({
    modelValue: (__VLS_ctx.dbGenerateFrontend),
    inlinePrompt: true,
    activeText: "Frontend",
    inactiveText: "Frontend",
}));
const __VLS_354 = __VLS_353({
    modelValue: (__VLS_ctx.dbGenerateFrontend),
    inlinePrompt: true,
    activeText: "Frontend",
    inactiveText: "Frontend",
}, ...__VLS_functionalComponentArgsRest(__VLS_353));
const __VLS_356 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_357 = __VLS_asFunctionalComponent(__VLS_356, new __VLS_356({
    modelValue: (__VLS_ctx.dbGeneratePolicy),
    inlinePrompt: true,
    activeText: "Policy",
    inactiveText: "Policy",
}));
const __VLS_358 = __VLS_357({
    modelValue: (__VLS_ctx.dbGeneratePolicy),
    inlinePrompt: true,
    activeText: "Policy",
    inactiveText: "Policy",
}, ...__VLS_functionalComponentArgsRest(__VLS_357));
var __VLS_347;
var __VLS_235;
var __VLS_187;
const __VLS_360 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_361 = __VLS_asFunctionalComponent(__VLS_360, new __VLS_360({
    label: (__VLS_ctx.t('codegen.mode.delete', 'Delete')),
    name: "delete",
}));
const __VLS_362 = __VLS_361({
    label: (__VLS_ctx.t('codegen.mode.delete', 'Delete')),
    name: "delete",
}, ...__VLS_functionalComponentArgsRest(__VLS_361));
__VLS_363.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "delete-mode-panel" },
});
const __VLS_364 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_365 = __VLS_asFunctionalComponent(__VLS_364, new __VLS_364({
    title: (__VLS_ctx.t('codegen.delete_preview_required', '请先执行删除预览，确认计划、风险与冲突后，再点击确认删除。')),
    type: "warning",
    closable: (false),
    showIcon: true,
}));
const __VLS_366 = __VLS_365({
    title: (__VLS_ctx.t('codegen.delete_preview_required', '请先执行删除预览，确认计划、风险与冲突后，再点击确认删除。')),
    type: "warning",
    closable: (false),
    showIcon: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_365));
const __VLS_368 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_369 = __VLS_asFunctionalComponent(__VLS_368, new __VLS_368({
    labelPosition: "top",
    ...{ class: "codegen-form db-form delete-form" },
}));
const __VLS_370 = __VLS_369({
    labelPosition: "top",
    ...{ class: "codegen-form db-form delete-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_369));
__VLS_371.slots.default;
const __VLS_372 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_373 = __VLS_asFunctionalComponent(__VLS_372, new __VLS_372({
    gutter: (16),
}));
const __VLS_374 = __VLS_373({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_373));
__VLS_375.slots.default;
const __VLS_376 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_377 = __VLS_asFunctionalComponent(__VLS_376, new __VLS_376({
    xs: (24),
    md: (8),
}));
const __VLS_378 = __VLS_377({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_377));
__VLS_379.slots.default;
const __VLS_380 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_381 = __VLS_asFunctionalComponent(__VLS_380, new __VLS_380({
    label: (__VLS_ctx.t('codegen.delete_module', '模块名')),
}));
const __VLS_382 = __VLS_381({
    label: (__VLS_ctx.t('codegen.delete_module', '模块名')),
}, ...__VLS_functionalComponentArgsRest(__VLS_381));
__VLS_383.slots.default;
const __VLS_384 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_385 = __VLS_asFunctionalComponent(__VLS_384, new __VLS_384({
    modelValue: (__VLS_ctx.deleteModule),
    placeholder: "例如 book",
}));
const __VLS_386 = __VLS_385({
    modelValue: (__VLS_ctx.deleteModule),
    placeholder: "例如 book",
}, ...__VLS_functionalComponentArgsRest(__VLS_385));
var __VLS_383;
var __VLS_379;
const __VLS_388 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_389 = __VLS_asFunctionalComponent(__VLS_388, new __VLS_388({
    xs: (24),
    md: (8),
}));
const __VLS_390 = __VLS_389({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_389));
__VLS_391.slots.default;
const __VLS_392 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_393 = __VLS_asFunctionalComponent(__VLS_392, new __VLS_392({
    label: (__VLS_ctx.t('codegen.delete_kind', '模块类型')),
}));
const __VLS_394 = __VLS_393({
    label: (__VLS_ctx.t('codegen.delete_kind', '模块类型')),
}, ...__VLS_functionalComponentArgsRest(__VLS_393));
__VLS_395.slots.default;
const __VLS_396 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_397 = __VLS_asFunctionalComponent(__VLS_396, new __VLS_396({
    modelValue: (__VLS_ctx.deleteKind),
    placeholder: "例如 crud",
}));
const __VLS_398 = __VLS_397({
    modelValue: (__VLS_ctx.deleteKind),
    placeholder: "例如 crud",
}, ...__VLS_functionalComponentArgsRest(__VLS_397));
var __VLS_395;
var __VLS_391;
const __VLS_400 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_401 = __VLS_asFunctionalComponent(__VLS_400, new __VLS_400({
    xs: (24),
    md: (8),
}));
const __VLS_402 = __VLS_401({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_401));
__VLS_403.slots.default;
const __VLS_404 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_405 = __VLS_asFunctionalComponent(__VLS_404, new __VLS_404({
    label: (__VLS_ctx.t('codegen.delete_policy_store', 'Policy Store')),
}));
const __VLS_406 = __VLS_405({
    label: (__VLS_ctx.t('codegen.delete_policy_store', 'Policy Store')),
}, ...__VLS_functionalComponentArgsRest(__VLS_405));
__VLS_407.slots.default;
const __VLS_408 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_409 = __VLS_asFunctionalComponent(__VLS_408, new __VLS_408({
    modelValue: (__VLS_ctx.deletePolicyStore),
    clearable: true,
    filterable: true,
    placeholder: (__VLS_ctx.t('codegen.delete_policy_store_placeholder', '自动识别或手动指定')),
}));
const __VLS_410 = __VLS_409({
    modelValue: (__VLS_ctx.deletePolicyStore),
    clearable: true,
    filterable: true,
    placeholder: (__VLS_ctx.t('codegen.delete_policy_store_placeholder', '自动识别或手动指定')),
}, ...__VLS_functionalComponentArgsRest(__VLS_409));
__VLS_411.slots.default;
const __VLS_412 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_413 = __VLS_asFunctionalComponent(__VLS_412, new __VLS_412({
    label: (__VLS_ctx.t('codegen.delete_policy_store_auto_detect', '自动识别')),
    value: "",
}));
const __VLS_414 = __VLS_413({
    label: (__VLS_ctx.t('codegen.delete_policy_store_auto_detect', '自动识别')),
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_413));
const __VLS_416 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_417 = __VLS_asFunctionalComponent(__VLS_416, new __VLS_416({
    label: "CSV",
    value: "csv",
}));
const __VLS_418 = __VLS_417({
    label: "CSV",
    value: "csv",
}, ...__VLS_functionalComponentArgsRest(__VLS_417));
const __VLS_420 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_421 = __VLS_asFunctionalComponent(__VLS_420, new __VLS_420({
    label: "DB",
    value: "db",
}));
const __VLS_422 = __VLS_421({
    label: "DB",
    value: "db",
}, ...__VLS_functionalComponentArgsRest(__VLS_421));
var __VLS_411;
var __VLS_407;
var __VLS_403;
var __VLS_375;
const __VLS_424 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_425 = __VLS_asFunctionalComponent(__VLS_424, new __VLS_424({
    label: (__VLS_ctx.t('codegen.delete_scope', '删除范围')),
}));
const __VLS_426 = __VLS_425({
    label: (__VLS_ctx.t('codegen.delete_scope', '删除范围')),
}, ...__VLS_functionalComponentArgsRest(__VLS_425));
__VLS_427.slots.default;
const __VLS_428 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_429 = __VLS_asFunctionalComponent(__VLS_428, new __VLS_428({
    wrap: true,
}));
const __VLS_430 = __VLS_429({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_429));
__VLS_431.slots.default;
const __VLS_432 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_433 = __VLS_asFunctionalComponent(__VLS_432, new __VLS_432({
    modelValue: (__VLS_ctx.deleteWithRuntime),
    inlinePrompt: true,
    activeText: "Runtime",
    inactiveText: "Runtime",
}));
const __VLS_434 = __VLS_433({
    modelValue: (__VLS_ctx.deleteWithRuntime),
    inlinePrompt: true,
    activeText: "Runtime",
    inactiveText: "Runtime",
}, ...__VLS_functionalComponentArgsRest(__VLS_433));
const __VLS_436 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_437 = __VLS_asFunctionalComponent(__VLS_436, new __VLS_436({
    modelValue: (__VLS_ctx.deleteWithPolicy),
    inlinePrompt: true,
    activeText: "Policy",
    inactiveText: "Policy",
}));
const __VLS_438 = __VLS_437({
    modelValue: (__VLS_ctx.deleteWithPolicy),
    inlinePrompt: true,
    activeText: "Policy",
    inactiveText: "Policy",
}, ...__VLS_functionalComponentArgsRest(__VLS_437));
const __VLS_440 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_441 = __VLS_asFunctionalComponent(__VLS_440, new __VLS_440({
    modelValue: (__VLS_ctx.deleteWithFrontend),
    inlinePrompt: true,
    activeText: "Frontend",
    inactiveText: "Frontend",
}));
const __VLS_442 = __VLS_441({
    modelValue: (__VLS_ctx.deleteWithFrontend),
    inlinePrompt: true,
    activeText: "Frontend",
    inactiveText: "Frontend",
}, ...__VLS_functionalComponentArgsRest(__VLS_441));
const __VLS_444 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_445 = __VLS_asFunctionalComponent(__VLS_444, new __VLS_444({
    modelValue: (__VLS_ctx.deleteWithRegistry),
    inlinePrompt: true,
    activeText: "Registry",
    inactiveText: "Registry",
}));
const __VLS_446 = __VLS_445({
    modelValue: (__VLS_ctx.deleteWithRegistry),
    inlinePrompt: true,
    activeText: "Registry",
    inactiveText: "Registry",
}, ...__VLS_functionalComponentArgsRest(__VLS_445));
const __VLS_448 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_449 = __VLS_asFunctionalComponent(__VLS_448, new __VLS_448({
    modelValue: (__VLS_ctx.deleteForce),
    inlinePrompt: true,
    activeText: "Force",
    inactiveText: "Force",
}));
const __VLS_450 = __VLS_449({
    modelValue: (__VLS_ctx.deleteForce),
    inlinePrompt: true,
    activeText: "Force",
    inactiveText: "Force",
}, ...__VLS_functionalComponentArgsRest(__VLS_449));
var __VLS_431;
var __VLS_427;
const __VLS_452 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_453 = __VLS_asFunctionalComponent(__VLS_452, new __VLS_452({
    label: (__VLS_ctx.t('codegen.execute_notes', '执行说明')),
}));
const __VLS_454 = __VLS_453({
    label: (__VLS_ctx.t('codegen.execute_notes', '执行说明')),
}, ...__VLS_functionalComponentArgsRest(__VLS_453));
__VLS_455.slots.default;
const __VLS_456 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_457 = __VLS_asFunctionalComponent(__VLS_456, new __VLS_456({
    modelValue: (__VLS_ctx.deleteNotes),
    type: "textarea",
    rows: (5),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.execute_notes_placeholder', '可选：补充删除说明，仅用于界面记录，不会直接传给后端核心')),
}));
const __VLS_458 = __VLS_457({
    modelValue: (__VLS_ctx.deleteNotes),
    type: "textarea",
    rows: (5),
    resize: "none",
    placeholder: (__VLS_ctx.t('codegen.execute_notes_placeholder', '可选：补充删除说明，仅用于界面记录，不会直接传给后端核心')),
}, ...__VLS_functionalComponentArgsRest(__VLS_457));
var __VLS_455;
var __VLS_371;
var __VLS_363;
var __VLS_131;
var __VLS_11;
var __VLS_7;
const __VLS_460 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_461 = __VLS_asFunctionalComponent(__VLS_460, new __VLS_460({
    xs: (24),
    lg: (13),
    xl: (13),
}));
const __VLS_462 = __VLS_461({
    xs: (24),
    lg: (13),
    xl: (13),
}, ...__VLS_functionalComponentArgsRest(__VLS_461));
__VLS_463.slots.default;
const __VLS_464 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_465 = __VLS_asFunctionalComponent(__VLS_464, new __VLS_464({
    direction: "vertical",
    size: (16),
    fill: true,
    ...{ class: "side-stack" },
}));
const __VLS_466 = __VLS_465({
    direction: "vertical",
    size: (16),
    fill: true,
    ...{ class: "side-stack" },
}, ...__VLS_functionalComponentArgsRest(__VLS_465));
__VLS_467.slots.default;
const __VLS_468 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_469 = __VLS_asFunctionalComponent(__VLS_468, new __VLS_468({
    shadow: "never",
    ...{ class: "codegen-card" },
}));
const __VLS_470 = __VLS_469({
    shadow: "never",
    ...{ class: "codegen-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_469));
__VLS_471.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_471.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-header compact" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "title" },
    });
    (__VLS_ctx.t('codegen.result_title', '执行结果'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "subtitle" },
    });
    (__VLS_ctx.t('codegen.result_subtitle', '预览和生成都会回传资源级动作。'));
}
if (__VLS_ctx.statusMessage) {
    const __VLS_472 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_473 = __VLS_asFunctionalComponent(__VLS_472, new __VLS_472({
        title: (__VLS_ctx.statusMessage),
        type: (__VLS_ctx.lastRunSuccess ? 'success' : 'info'),
        closable: (false),
        showIcon: true,
    }));
    const __VLS_474 = __VLS_473({
        title: (__VLS_ctx.statusMessage),
        type: (__VLS_ctx.lastRunSuccess ? 'success' : 'info'),
        closable: (false),
        showIcon: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_473));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "result-empty" },
    });
    (__VLS_ctx.t('codegen.no_result', '尚未执行预览或生成。'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "preview-table-wrap" },
});
const __VLS_476 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_477 = __VLS_asFunctionalComponent(__VLS_476, new __VLS_476({
    data: (__VLS_ctx.previewItems),
    ...{ class: "preview-table" },
    size: "small",
    border: true,
}));
const __VLS_478 = __VLS_477({
    data: (__VLS_ctx.previewItems),
    ...{ class: "preview-table" },
    size: "small",
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_477));
__VLS_479.slots.default;
const __VLS_480 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_481 = __VLS_asFunctionalComponent(__VLS_480, new __VLS_480({
    prop: "index",
    label: (__VLS_ctx.t('codegen.preview.index', '#')),
    width: "60",
}));
const __VLS_482 = __VLS_481({
    prop: "index",
    label: (__VLS_ctx.t('codegen.preview.index', '#')),
    width: "60",
}, ...__VLS_functionalComponentArgsRest(__VLS_481));
const __VLS_484 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_485 = __VLS_asFunctionalComponent(__VLS_484, new __VLS_484({
    prop: "kind",
    label: (__VLS_ctx.t('codegen.preview.kind', 'Kind')),
    minWidth: "130",
    showOverflowTooltip: true,
}));
const __VLS_486 = __VLS_485({
    prop: "kind",
    label: (__VLS_ctx.t('codegen.preview.kind', 'Kind')),
    minWidth: "130",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_485));
const __VLS_488 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_489 = __VLS_asFunctionalComponent(__VLS_488, new __VLS_488({
    prop: "name",
    label: (__VLS_ctx.t('codegen.preview.name', 'Name')),
    minWidth: "160",
    showOverflowTooltip: true,
}));
const __VLS_490 = __VLS_489({
    prop: "name",
    label: (__VLS_ctx.t('codegen.preview.name', 'Name')),
    minWidth: "160",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_489));
const __VLS_492 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_493 = __VLS_asFunctionalComponent(__VLS_492, new __VLS_492({
    label: (__VLS_ctx.t('codegen.preview.force', 'Force')),
    width: "88",
}));
const __VLS_494 = __VLS_493({
    label: (__VLS_ctx.t('codegen.preview.force', 'Force')),
    width: "88",
}, ...__VLS_functionalComponentArgsRest(__VLS_493));
__VLS_495.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_495.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_496 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_497 = __VLS_asFunctionalComponent(__VLS_496, new __VLS_496({
        type: (scope.row.force ? 'warning' : 'info'),
        effect: "light",
    }));
    const __VLS_498 = __VLS_497({
        type: (scope.row.force ? 'warning' : 'info'),
        effect: "light",
    }, ...__VLS_functionalComponentArgsRest(__VLS_497));
    __VLS_499.slots.default;
    (scope.row.force ? __VLS_ctx.t('common.yes', 'Yes') : __VLS_ctx.t('common.no', 'No'));
    var __VLS_499;
}
var __VLS_495;
const __VLS_500 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_501 = __VLS_asFunctionalComponent(__VLS_500, new __VLS_500({
    label: (__VLS_ctx.t('codegen.preview.actions', 'Actions')),
    minWidth: "280",
}));
const __VLS_502 = __VLS_501({
    label: (__VLS_ctx.t('codegen.preview.actions', 'Actions')),
    minWidth: "280",
}, ...__VLS_functionalComponentArgsRest(__VLS_501));
__VLS_503.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_503.slots;
    const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "action-tags" },
    });
    for (const [action] of __VLS_getVForSourceType((scope.row.actions))) {
        const __VLS_504 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_505 = __VLS_asFunctionalComponent(__VLS_504, new __VLS_504({
            key: (action),
            size: "small",
            effect: "plain",
            ...{ class: "action-tag" },
        }));
        const __VLS_506 = __VLS_505({
            key: (action),
            size: "small",
            effect: "plain",
            ...{ class: "action-tag" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_505));
        __VLS_507.slots.default;
        (action);
        var __VLS_507;
    }
}
var __VLS_503;
var __VLS_479;
var __VLS_471;
if (__VLS_ctx.activeMode === 'db') {
    const __VLS_508 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_509 = __VLS_asFunctionalComponent(__VLS_508, new __VLS_508({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_510 = __VLS_509({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_509));
    __VLS_511.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_511.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.file_plan_title', '文件计划'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.file_plan_subtitle', '显示数据库预览阶段推导出的文件清单。'));
    }
    if (!__VLS_ctx.filePlans.length) {
        const __VLS_512 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_513 = __VLS_asFunctionalComponent(__VLS_512, new __VLS_512({
            description: (__VLS_ctx.t('codegen.no_file_plan', '暂无文件计划')),
        }));
        const __VLS_514 = __VLS_513({
            description: (__VLS_ctx.t('codegen.no_file_plan', '暂无文件计划')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_513));
    }
    else {
        const __VLS_516 = {}.ElTable;
        /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
        // @ts-ignore
        const __VLS_517 = __VLS_asFunctionalComponent(__VLS_516, new __VLS_516({
            data: (__VLS_ctx.filePlans),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }));
        const __VLS_518 = __VLS_517({
            data: (__VLS_ctx.filePlans),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_517));
        __VLS_519.slots.default;
        const __VLS_520 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_521 = __VLS_asFunctionalComponent(__VLS_520, new __VLS_520({
            prop: "path",
            label: (__VLS_ctx.t('codegen.file_plan.path', 'Path')),
            minWidth: "220",
            showOverflowTooltip: true,
        }));
        const __VLS_522 = __VLS_521({
            prop: "path",
            label: (__VLS_ctx.t('codegen.file_plan.path', 'Path')),
            minWidth: "220",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_521));
        const __VLS_524 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_525 = __VLS_asFunctionalComponent(__VLS_524, new __VLS_524({
            prop: "action",
            label: (__VLS_ctx.t('codegen.file_plan.action', 'Action')),
            width: "120",
        }));
        const __VLS_526 = __VLS_525({
            prop: "action",
            label: (__VLS_ctx.t('codegen.file_plan.action', 'Action')),
            width: "120",
        }, ...__VLS_functionalComponentArgsRest(__VLS_525));
        const __VLS_528 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_529 = __VLS_asFunctionalComponent(__VLS_528, new __VLS_528({
            prop: "kind",
            label: (__VLS_ctx.t('codegen.file_plan.kind', 'Kind')),
            width: "140",
            showOverflowTooltip: true,
        }));
        const __VLS_530 = __VLS_529({
            prop: "kind",
            label: (__VLS_ctx.t('codegen.file_plan.kind', 'Kind')),
            width: "140",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_529));
        const __VLS_532 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_533 = __VLS_asFunctionalComponent(__VLS_532, new __VLS_532({
            label: (__VLS_ctx.t('codegen.file_plan.exists', 'Exists')),
            width: "88",
        }));
        const __VLS_534 = __VLS_533({
            label: (__VLS_ctx.t('codegen.file_plan.exists', 'Exists')),
            width: "88",
        }, ...__VLS_functionalComponentArgsRest(__VLS_533));
        __VLS_535.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_535.slots;
            const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_536 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_537 = __VLS_asFunctionalComponent(__VLS_536, new __VLS_536({
                type: (scope.row.exists ? 'warning' : 'success'),
                effect: "light",
            }));
            const __VLS_538 = __VLS_537({
                type: (scope.row.exists ? 'warning' : 'success'),
                effect: "light",
            }, ...__VLS_functionalComponentArgsRest(__VLS_537));
            __VLS_539.slots.default;
            (scope.row.exists ? __VLS_ctx.t('common.yes', 'Yes') : __VLS_ctx.t('common.no', 'No'));
            var __VLS_539;
        }
        var __VLS_535;
        const __VLS_540 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_541 = __VLS_asFunctionalComponent(__VLS_540, new __VLS_540({
            label: (__VLS_ctx.t('codegen.file_plan.conflict', 'Conflict')),
            width: "96",
        }));
        const __VLS_542 = __VLS_541({
            label: (__VLS_ctx.t('codegen.file_plan.conflict', 'Conflict')),
            width: "96",
        }, ...__VLS_functionalComponentArgsRest(__VLS_541));
        __VLS_543.slots.default;
        {
            const { default: __VLS_thisSlot } = __VLS_543.slots;
            const [scope] = __VLS_getSlotParams(__VLS_thisSlot);
            const __VLS_544 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_545 = __VLS_asFunctionalComponent(__VLS_544, new __VLS_544({
                type: (scope.row.conflict ? 'danger' : 'success'),
                effect: "light",
            }));
            const __VLS_546 = __VLS_545({
                type: (scope.row.conflict ? 'danger' : 'success'),
                effect: "light",
            }, ...__VLS_functionalComponentArgsRest(__VLS_545));
            __VLS_547.slots.default;
            (scope.row.conflict ? __VLS_ctx.t('common.yes', 'Yes') : __VLS_ctx.t('common.no', 'No'));
            var __VLS_547;
        }
        var __VLS_543;
        var __VLS_519;
    }
    var __VLS_511;
}
if (__VLS_ctx.activeMode === 'delete') {
    const __VLS_548 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_549 = __VLS_asFunctionalComponent(__VLS_548, new __VLS_548({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_550 = __VLS_549({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_549));
    __VLS_551.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_551.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.risk_title', '风险与冲突'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.risk_subtitle', '展示删除预览中的风险提示和阻断冲突。'));
    }
    if (!__VLS_ctx.deleteConflicts.length) {
        const __VLS_552 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_553 = __VLS_asFunctionalComponent(__VLS_552, new __VLS_552({
            description: (__VLS_ctx.t('codegen.no_risk', 'No conflicts')),
        }));
        const __VLS_554 = __VLS_553({
            description: (__VLS_ctx.t('codegen.no_risk', 'No conflicts')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_553));
    }
    else {
        const __VLS_556 = {}.ElTable;
        /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
        // @ts-ignore
        const __VLS_557 = __VLS_asFunctionalComponent(__VLS_556, new __VLS_556({
            data: (__VLS_ctx.deleteConflicts),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }));
        const __VLS_558 = __VLS_557({
            data: (__VLS_ctx.deleteConflicts),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_557));
        __VLS_559.slots.default;
        const __VLS_560 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_561 = __VLS_asFunctionalComponent(__VLS_560, new __VLS_560({
            prop: "kind",
            label: (__VLS_ctx.t('codegen.conflict.kind', 'Kind')),
            minWidth: "140",
            showOverflowTooltip: true,
        }));
        const __VLS_562 = __VLS_561({
            prop: "kind",
            label: (__VLS_ctx.t('codegen.conflict.kind', 'Kind')),
            minWidth: "140",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_561));
        const __VLS_564 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_565 = __VLS_asFunctionalComponent(__VLS_564, new __VLS_564({
            prop: "severity",
            label: (__VLS_ctx.t('codegen.conflict.severity', 'Severity')),
            width: "110",
        }));
        const __VLS_566 = __VLS_565({
            prop: "severity",
            label: (__VLS_ctx.t('codegen.conflict.severity', 'Severity')),
            width: "110",
        }, ...__VLS_functionalComponentArgsRest(__VLS_565));
        const __VLS_568 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_569 = __VLS_asFunctionalComponent(__VLS_568, new __VLS_568({
            prop: "path",
            label: (__VLS_ctx.t('codegen.conflict.path', 'Path')),
            minWidth: "200",
            showOverflowTooltip: true,
        }));
        const __VLS_570 = __VLS_569({
            prop: "path",
            label: (__VLS_ctx.t('codegen.conflict.path', 'Path')),
            minWidth: "200",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_569));
        const __VLS_572 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_573 = __VLS_asFunctionalComponent(__VLS_572, new __VLS_572({
            prop: "message",
            label: (__VLS_ctx.t('codegen.conflict.message', 'Message')),
            minWidth: "280",
            showOverflowTooltip: true,
        }));
        const __VLS_574 = __VLS_573({
            prop: "message",
            label: (__VLS_ctx.t('codegen.conflict.message', 'Message')),
            minWidth: "280",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_573));
        var __VLS_559;
    }
    var __VLS_551;
}
if (__VLS_ctx.activeMode === 'db') {
    const __VLS_576 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_577 = __VLS_asFunctionalComponent(__VLS_576, new __VLS_576({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_578 = __VLS_577({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_577));
    __VLS_579.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_579.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.conflict_title', '冲突'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.conflict_subtitle', '展示文件覆盖风险与路径冲突信息。'));
    }
    if (!__VLS_ctx.conflicts.length) {
        const __VLS_580 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_581 = __VLS_asFunctionalComponent(__VLS_580, new __VLS_580({
            description: (__VLS_ctx.t('codegen.no_conflicts', 'No conflicts')),
        }));
        const __VLS_582 = __VLS_581({
            description: (__VLS_ctx.t('codegen.no_conflicts', 'No conflicts')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_581));
    }
    else {
        const __VLS_584 = {}.ElTable;
        /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
        // @ts-ignore
        const __VLS_585 = __VLS_asFunctionalComponent(__VLS_584, new __VLS_584({
            data: (__VLS_ctx.conflicts),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }));
        const __VLS_586 = __VLS_585({
            data: (__VLS_ctx.conflicts),
            size: "small",
            border: true,
            ...{ class: "preview-table" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_585));
        __VLS_587.slots.default;
        const __VLS_588 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_589 = __VLS_asFunctionalComponent(__VLS_588, new __VLS_588({
            prop: "path",
            label: (__VLS_ctx.t('codegen.conflict.path', 'Path')),
            minWidth: "220",
            showOverflowTooltip: true,
        }));
        const __VLS_590 = __VLS_589({
            prop: "path",
            label: (__VLS_ctx.t('codegen.conflict.path', 'Path')),
            minWidth: "220",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_589));
        const __VLS_592 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_593 = __VLS_asFunctionalComponent(__VLS_592, new __VLS_592({
            prop: "resource",
            label: (__VLS_ctx.t('codegen.conflict.resource', 'Resource')),
            minWidth: "160",
            showOverflowTooltip: true,
        }));
        const __VLS_594 = __VLS_593({
            prop: "resource",
            label: (__VLS_ctx.t('codegen.conflict.resource', 'Resource')),
            minWidth: "160",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_593));
        const __VLS_596 = {}.ElTableColumn;
        /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
        // @ts-ignore
        const __VLS_597 = __VLS_asFunctionalComponent(__VLS_596, new __VLS_596({
            prop: "reason",
            label: (__VLS_ctx.t('codegen.conflict.reason', 'Reason')),
            minWidth: "260",
            showOverflowTooltip: true,
        }));
        const __VLS_598 = __VLS_597({
            prop: "reason",
            label: (__VLS_ctx.t('codegen.conflict.reason', 'Reason')),
            minWidth: "260",
            showOverflowTooltip: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_597));
        var __VLS_587;
    }
    var __VLS_579;
}
if (__VLS_ctx.activeMode === 'db') {
    const __VLS_600 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_601 = __VLS_asFunctionalComponent(__VLS_600, new __VLS_600({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_602 = __VLS_601({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_601));
    __VLS_603.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_603.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.audit_title', '审计'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.audit_subtitle', '记录输入、执行步骤与输出统计。'));
    }
    if (!__VLS_ctx.auditRecord) {
        const __VLS_604 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_605 = __VLS_asFunctionalComponent(__VLS_604, new __VLS_604({
            description: (__VLS_ctx.t('codegen.no_audit', '暂无审计记录')),
        }));
        const __VLS_606 = __VLS_605({
            description: (__VLS_ctx.t('codegen.no_audit', '暂无审计记录')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_605));
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-panel" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-grid" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.record_time', '记录时间'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.formatDateTime(__VLS_ctx.auditRecord.recorded_at));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.auditRecord.recorded_at);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.input', '输入'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.auditRecord.input.driver);
        (__VLS_ctx.auditRecord.input.database);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.auditRecord.input.dry_run ? __VLS_ctx.t('common.yes', 'Yes') : __VLS_ctx.t('common.no', 'No'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.output_file_count', '输出文件'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.auditRecord.output.file_count);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.t('codegen.conflict_count_label', 'Conflicts: {count}', { count: __VLS_ctx.auditRecord.output.conflict_count }));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.table_scope', '表范围'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.auditRecord.input.tables?.length ?? 0);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        ((__VLS_ctx.auditRecord.input.tables ?? []).join(', ') || '全部表');
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-grid artifact-grid--compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item artifact-item--wide" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.execution_steps', '执行步骤'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
            ...{ class: "message-list compact-list" },
        });
        for (const [step] of __VLS_getVForSourceType((__VLS_ctx.auditRecord.steps))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
                key: (step.name),
            });
            (step.name);
            (step.status);
            (step.detail ? ` - ${step.detail}` : '');
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item artifact-item--wide" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.output_overview', '输出概览'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-value" },
        });
        (__VLS_ctx.auditRecord.output.files.length);
        (__VLS_ctx.auditRecord.output.conflicts.length);
    }
    var __VLS_603;
}
if (__VLS_ctx.activeMode === 'delete' && __VLS_ctx.deleteResult) {
    const __VLS_608 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_609 = __VLS_asFunctionalComponent(__VLS_608, new __VLS_608({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_610 = __VLS_609({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_609));
    __VLS_611.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_611.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.delete_result_title', '删除结果'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.delete_result_subtitle', '展示本次删除的执行概览、异常情况与处理明细。'));
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-panel" },
    });
    const __VLS_612 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_613 = __VLS_asFunctionalComponent(__VLS_612, new __VLS_612({
        title: (__VLS_ctx.deleteResultStatusMessage),
        type: (__VLS_ctx.deleteResultStatusType),
        closable: (false),
        showIcon: true,
    }));
    const __VLS_614 = __VLS_613({
        title: (__VLS_ctx.deleteResultStatusMessage),
        type: (__VLS_ctx.deleteResultStatusType),
        closable: (false),
        showIcon: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_613));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-summary-grid" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-summary-card" },
        ...{ class: (`artifact-summary-card--${__VLS_ctx.deleteResultStatusTone}`) },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-label" },
    });
    (__VLS_ctx.t('codegen.result_status', '结果状态'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-value" },
    });
    (__VLS_ctx.deleteResultStatusLabel);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-meta" },
    });
    (__VLS_ctx.deleteResultSummaryText);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-summary-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-label" },
    });
    (__VLS_ctx.t('codegen.processed', '已处理'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-value" },
    });
    (__VLS_ctx.deleteResultSummary?.total_deleted ?? 0);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-meta" },
    });
    (__VLS_ctx.deleteResultSummary?.deleted_source_files ?? 0);
    (__VLS_ctx.deleteResultSummary?.deleted_runtime_assets ?? 0);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-summary-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-label" },
    });
    (__VLS_ctx.t('codegen.skipped_failed', '跳过 / 异常'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-value" },
    });
    (__VLS_ctx.deleteResultSummary?.skipped ?? 0);
    (__VLS_ctx.deleteResultSummary?.failed ?? 0);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-meta" },
    });
    (__VLS_ctx.deleteResultSummary?.deleted_policy_changes ?? 0);
    (__VLS_ctx.deleteResultSummary?.deleted_frontend_changes ?? 0);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-summary-card" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-label" },
    });
    (__VLS_ctx.t('codegen.elapsed', '执行耗时'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-value" },
    });
    (__VLS_ctx.deleteResultElapsedText);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-summary-meta" },
    });
    (__VLS_ctx.formatDateTime(__VLS_ctx.deleteResult?.started_at ?? ''));
    (__VLS_ctx.formatDateTime(__VLS_ctx.deleteResult?.finished_at ?? ''));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-grid artifact-grid--compact" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-item artifact-item--wide" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-label" },
    });
    (__VLS_ctx.t('codegen.delete_detail', '删除明细'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
        ...{ class: "message-list compact-list" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.deleteResultDeleted))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
            key: (__VLS_ctx.deleteItemKey(item)),
        });
        (__VLS_ctx.describeDeleteItem(item));
    }
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "artifact-item artifact-item--wide" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "artifact-label" },
    });
    (__VLS_ctx.t('codegen.skip_detail', '跳过明细'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
        ...{ class: "message-list compact-list" },
    });
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.deleteResultSkipped))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
            key: (__VLS_ctx.deleteItemKey(item)),
        });
        (__VLS_ctx.describeDeleteItem(item));
    }
    if (__VLS_ctx.deleteResultFailures.length) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item artifact-item--wide" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.failure_detail', '异常明细'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
            ...{ class: "message-list compact-list" },
        });
        for (const [failure] of __VLS_getVForSourceType((__VLS_ctx.deleteResultFailures))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
                key: (__VLS_ctx.describeDeleteFailureKey(failure)),
            });
            (__VLS_ctx.describeDeleteFailure(failure));
        }
    }
    var __VLS_611;
}
const __VLS_616 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_617 = __VLS_asFunctionalComponent(__VLS_616, new __VLS_616({
    shadow: "never",
    ...{ class: "codegen-card" },
}));
const __VLS_618 = __VLS_617({
    shadow: "never",
    ...{ class: "codegen-card" },
}, ...__VLS_functionalComponentArgsRest(__VLS_617));
__VLS_619.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_619.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "card-header compact" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "title" },
    });
    (__VLS_ctx.messagePanelTitle);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "subtitle" },
    });
    (__VLS_ctx.messagePanelSubtitle);
}
if (!__VLS_ctx.messages.length) {
    const __VLS_620 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_621 = __VLS_asFunctionalComponent(__VLS_620, new __VLS_620({
        description: (__VLS_ctx.messagePanelEmptyText),
    }));
    const __VLS_622 = __VLS_621({
        description: (__VLS_ctx.messagePanelEmptyText),
    }, ...__VLS_functionalComponentArgsRest(__VLS_621));
}
else {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
        ...{ class: "message-list" },
    });
    for (const [message] of __VLS_getVForSourceType((__VLS_ctx.messages))) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
            key: (message),
        });
        (message);
    }
}
var __VLS_619;
if (__VLS_ctx.artifactInfo) {
    const __VLS_624 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_625 = __VLS_asFunctionalComponent(__VLS_624, new __VLS_624({
        shadow: "never",
        ...{ class: "codegen-card" },
    }));
    const __VLS_626 = __VLS_625({
        shadow: "never",
        ...{ class: "codegen-card" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_625));
    __VLS_627.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_627.slots;
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "card-header compact artifact-header" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "title" },
        });
        (__VLS_ctx.t('codegen.artifact_title', '下载产物'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "subtitle" },
        });
        (__VLS_ctx.t('codegen.artifact_subtitle', '展示最近一次服务端打包结果，并支持重新下载。'));
        if (__VLS_ctx.artifactInfo) {
            const __VLS_628 = {}.ElSpace;
            /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
            // @ts-ignore
            const __VLS_629 = __VLS_asFunctionalComponent(__VLS_628, new __VLS_628({
                wrap: true,
            }));
            const __VLS_630 = __VLS_629({
                wrap: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_629));
            __VLS_631.slots.default;
            const __VLS_632 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_633 = __VLS_asFunctionalComponent(__VLS_632, new __VLS_632({
                ...{ 'onClick': {} },
                text: true,
                ...{ class: ({ 'artifact-copy-button--active': __VLS_ctx.copyFeedbackActive }) },
                disabled: (!__VLS_ctx.canCopyArtifactUrl),
            }));
            const __VLS_634 = __VLS_633({
                ...{ 'onClick': {} },
                text: true,
                ...{ class: ({ 'artifact-copy-button--active': __VLS_ctx.copyFeedbackActive }) },
                disabled: (!__VLS_ctx.canCopyArtifactUrl),
            }, ...__VLS_functionalComponentArgsRest(__VLS_633));
            let __VLS_636;
            let __VLS_637;
            let __VLS_638;
            const __VLS_639 = {
                onClick: (__VLS_ctx.handleCopyDownloadUrl)
            };
            __VLS_635.slots.default;
            (__VLS_ctx.copyButtonText);
            var __VLS_635;
            const __VLS_640 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_641 = __VLS_asFunctionalComponent(__VLS_640, new __VLS_640({
                ...{ 'onClick': {} },
                text: true,
                type: "primary",
                disabled: (__VLS_ctx.isArtifactExpired),
                loading: (__VLS_ctx.downloadLoading),
            }));
            const __VLS_642 = __VLS_641({
                ...{ 'onClick': {} },
                text: true,
                type: "primary",
                disabled: (__VLS_ctx.isArtifactExpired),
                loading: (__VLS_ctx.downloadLoading),
            }, ...__VLS_functionalComponentArgsRest(__VLS_641));
            let __VLS_644;
            let __VLS_645;
            let __VLS_646;
            const __VLS_647 = {
                onClick: (__VLS_ctx.handleArtifactDownload)
            };
            __VLS_643.slots.default;
            (__VLS_ctx.t('common.refresh', '重新下载'));
            var __VLS_643;
            var __VLS_631;
        }
    }
    if (!__VLS_ctx.artifactInfo) {
        const __VLS_648 = {}.ElEmpty;
        /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
        // @ts-ignore
        const __VLS_649 = __VLS_asFunctionalComponent(__VLS_648, new __VLS_648({
            description: (__VLS_ctx.t('codegen.no_artifact', '尚未生成下载包')),
        }));
        const __VLS_650 = __VLS_649({
            description: (__VLS_ctx.t('codegen.no_artifact', '尚未生成下载包')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_649));
    }
    else {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-panel" },
        });
        const __VLS_652 = {}.ElAlert;
        /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
        // @ts-ignore
        const __VLS_653 = __VLS_asFunctionalComponent(__VLS_652, new __VLS_652({
            title: (__VLS_ctx.artifactStatusMessage),
            type: (__VLS_ctx.artifactStatusType),
            closable: (false),
            showIcon: true,
        }));
        const __VLS_654 = __VLS_653({
            title: (__VLS_ctx.artifactStatusMessage),
            type: (__VLS_ctx.artifactStatusType),
            closable: (false),
            showIcon: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_653));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-grid" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
            ...{ class: (`artifact-summary-card--${__VLS_ctx.artifactStatusTone}`) },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.artifact_status', '产物状态'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.artifactStatusLabel);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.artifactStatusSummary);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.file_overview', '文件概览'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.artifactInfo.filename);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.artifactSizeText);
        (__VLS_ctx.artifactInfo.file_count);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.valid_until', '有效期'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.artifactRemainingText);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        (__VLS_ctx.artifactExpireText);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-summary-card" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-label" },
        });
        (__VLS_ctx.t('codegen.recent_activity', '最近活动'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-value" },
        });
        (__VLS_ctx.artifactLastDownloadText);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-summary-meta" },
        });
        if (__VLS_ctx.lastDownloadDuration > 0) {
            (__VLS_ctx.artifactLastDownloadDurationText);
        }
        (__VLS_ctx.artifactLastFailureText);
        if (__VLS_ctx.lastArtifactError) {
            const __VLS_656 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_657 = __VLS_asFunctionalComponent(__VLS_656, new __VLS_656({
                type: (__VLS_ctx.artifactLastErrorTypeTag),
                size: "small",
                effect: "light",
                ...{ style: {} },
            }));
            const __VLS_658 = __VLS_657({
                type: (__VLS_ctx.artifactLastErrorTypeTag),
                size: "small",
                effect: "light",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_657));
            __VLS_659.slots.default;
            (__VLS_ctx.artifactLastErrorTypeText);
            var __VLS_659;
        }
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-grid artifact-grid--compact" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item artifact-item--wide" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.task_id', '任务 ID'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-value monospace" },
        });
        (__VLS_ctx.artifactInfo.task_id);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.filename', '文件名'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-value" },
        });
        (__VLS_ctx.artifactInfo.filename);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.last_failure_reason', '最近一次失败原因'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-value" },
        });
        (__VLS_ctx.artifactLastErrorText);
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
            ...{ class: "artifact-item artifact-item--wide" },
        });
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
            ...{ class: "artifact-label" },
        });
        (__VLS_ctx.t('codegen.download_url', '下载地址'));
        if (__VLS_ctx.canCopyArtifactUrl) {
            const __VLS_660 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_661 = __VLS_asFunctionalComponent(__VLS_660, new __VLS_660({
                ...{ 'onClick': {} },
                plain: true,
                size: "small",
                ...{ class: ({ 'artifact-copy-button--active': __VLS_ctx.copyFeedbackActive }) },
            }));
            const __VLS_662 = __VLS_661({
                ...{ 'onClick': {} },
                plain: true,
                size: "small",
                ...{ class: ({ 'artifact-copy-button--active': __VLS_ctx.copyFeedbackActive }) },
            }, ...__VLS_functionalComponentArgsRest(__VLS_661));
            let __VLS_664;
            let __VLS_665;
            let __VLS_666;
            const __VLS_667 = {
                onClick: (__VLS_ctx.handleCopyDownloadUrl)
            };
            __VLS_663.slots.default;
            (__VLS_ctx.copyButtonText);
            var __VLS_663;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "artifact-hint" },
            });
            (__VLS_ctx.artifactDownloadUrlSummary);
        }
        else {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                ...{ class: "artifact-hint" },
            });
            (__VLS_ctx.t('codegen.expired_hidden', '下载包已过期，旧下载地址已隐藏，请重新生成新的代码包。'));
        }
    }
    var __VLS_627;
}
var __VLS_467;
var __VLS_463;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['codegen-page']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-form']} */ ;
/** @type {__VLS_StyleScopedClasses['hidden-file-input']} */ ;
/** @type {__VLS_StyleScopedClasses['db-mode-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['db-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['db-hero-title']} */ ;
/** @type {__VLS_StyleScopedClasses['db-hero-subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['db-preset-row']} */ ;
/** @type {__VLS_StyleScopedClasses['db-section-title']} */ ;
/** @type {__VLS_StyleScopedClasses['db-section-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form-item']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form-row']} */ ;
/** @type {__VLS_StyleScopedClasses['db-field-help']} */ ;
/** @type {__VLS_StyleScopedClasses['db-table-chip-list']} */ ;
/** @type {__VLS_StyleScopedClasses['db-table-chip']} */ ;
/** @type {__VLS_StyleScopedClasses['db-advanced']} */ ;
/** @type {__VLS_StyleScopedClasses['db-section-header']} */ ;
/** @type {__VLS_StyleScopedClasses['db-section-title']} */ ;
/** @type {__VLS_StyleScopedClasses['db-section-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['delete-mode-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-form']} */ ;
/** @type {__VLS_StyleScopedClasses['db-form']} */ ;
/** @type {__VLS_StyleScopedClasses['delete-form']} */ ;
/** @type {__VLS_StyleScopedClasses['side-stack']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['result-empty']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-table-wrap']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-table']} */ ;
/** @type {__VLS_StyleScopedClasses['action-tags']} */ ;
/** @type {__VLS_StyleScopedClasses['action-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-table']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-table']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['preview-table']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid--compact']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['message-list']} */ ;
/** @type {__VLS_StyleScopedClasses['compact-list']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-value']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid--compact']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['message-list']} */ ;
/** @type {__VLS_StyleScopedClasses['compact-list']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['message-list']} */ ;
/** @type {__VLS_StyleScopedClasses['compact-list']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['message-list']} */ ;
/** @type {__VLS_StyleScopedClasses['compact-list']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['message-list']} */ ;
/** @type {__VLS_StyleScopedClasses['codegen-card']} */ ;
/** @type {__VLS_StyleScopedClasses['card-header']} */ ;
/** @type {__VLS_StyleScopedClasses['compact']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-header']} */ ;
/** @type {__VLS_StyleScopedClasses['title']} */ ;
/** @type {__VLS_StyleScopedClasses['subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-panel']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-card']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-summary-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-grid--compact']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-value']} */ ;
/** @type {__VLS_StyleScopedClasses['monospace']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-value']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-item--wide']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-label']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-hint']} */ ;
/** @type {__VLS_StyleScopedClasses['artifact-hint']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            activeMode: activeMode,
            dslText: dslText,
            force: force,
            packageName: packageName,
            includeReadme: includeReadme,
            includeReport: includeReport,
            includeDsl: includeDsl,
            dbDriver: dbDriver,
            dbDatabase: dbDatabase,
            dbSchema: dbSchema,
            dbTablesText: dbTablesText,
            dbForce: dbForce,
            dbGenerateFrontend: dbGenerateFrontend,
            dbGeneratePolicy: dbGeneratePolicy,
            dbMountParentPath: dbMountParentPath,
            deleteModule: deleteModule,
            deleteKind: deleteKind,
            deletePolicyStore: deletePolicyStore,
            deleteWithRuntime: deleteWithRuntime,
            deleteWithPolicy: deleteWithPolicy,
            deleteWithFrontend: deleteWithFrontend,
            deleteWithRegistry: deleteWithRegistry,
            deleteForce: deleteForce,
            deleteNotes: deleteNotes,
            previewLoading: previewLoading,
            generateLoading: generateLoading,
            downloadLoading: downloadLoading,
            installLoading: installLoading,
            deleteLoading: deleteLoading,
            deleteResult: deleteResult,
            artifactInfo: artifactInfo,
            lastArtifactError: lastArtifactError,
            lastDownloadDuration: lastDownloadDuration,
            copyFeedbackActive: copyFeedbackActive,
            lastRunSuccess: lastRunSuccess,
            fileInputRef: fileInputRef,
            dbMountMenuOptions: dbMountMenuOptions,
            t: t,
            dbParsedTables: dbParsedTables,
            dbDriverLabel: dbDriverLabel,
            dbOptionSummary: dbOptionSummary,
            deleteConflicts: deleteConflicts,
            deleteResultSummary: deleteResultSummary,
            deleteResultDeleted: deleteResultDeleted,
            deleteResultSkipped: deleteResultSkipped,
            deleteResultFailures: deleteResultFailures,
            deleteResultElapsedText: deleteResultElapsedText,
            deleteResultSummaryText: deleteResultSummaryText,
            deleteResultStatusLabel: deleteResultStatusLabel,
            deleteResultStatusType: deleteResultStatusType,
            deleteResultStatusTone: deleteResultStatusTone,
            deleteResultStatusMessage: deleteResultStatusMessage,
            messagePanelTitle: messagePanelTitle,
            messagePanelSubtitle: messagePanelSubtitle,
            messagePanelEmptyText: messagePanelEmptyText,
            deleteExecuteEnabled: deleteExecuteEnabled,
            previewItems: previewItems,
            messages: messages,
            filePlans: filePlans,
            conflicts: conflicts,
            auditRecord: auditRecord,
            statusMessage: statusMessage,
            artifactSizeText: artifactSizeText,
            artifactExpireText: artifactExpireText,
            isArtifactExpired: isArtifactExpired,
            artifactStatusType: artifactStatusType,
            artifactStatusMessage: artifactStatusMessage,
            artifactStatusLabel: artifactStatusLabel,
            artifactStatusSummary: artifactStatusSummary,
            artifactStatusTone: artifactStatusTone,
            artifactDownloadUrlSummary: artifactDownloadUrlSummary,
            canCopyArtifactUrl: canCopyArtifactUrl,
            artifactRemainingText: artifactRemainingText,
            artifactLastDownloadText: artifactLastDownloadText,
            artifactLastErrorText: artifactLastErrorText,
            artifactLastFailureText: artifactLastFailureText,
            artifactLastErrorTypeText: artifactLastErrorTypeText,
            artifactLastErrorTypeTag: artifactLastErrorTypeTag,
            artifactLastDownloadDurationText: artifactLastDownloadDurationText,
            copyButtonText: copyButtonText,
            loadSample: loadSample,
            loadDbSample: loadDbSample,
            loadDeleteSample: loadDeleteSample,
            applyDbPreset: applyDbPreset,
            handleDeleteExecute: handleDeleteExecute,
            clearCurrentInputs: clearCurrentInputs,
            clearDbTables: clearDbTables,
            triggerFileSelect: triggerFileSelect,
            handleFileChange: handleFileChange,
            handlePreview: handlePreview,
            handleGenerate: handleGenerate,
            handleGenerateAndInstall: handleGenerateAndInstall,
            handleDeletePreview: handleDeletePreview,
            handleGenerateDownload: handleGenerateDownload,
            handleArtifactDownload: handleArtifactDownload,
            handleCopyDownloadUrl: handleCopyDownloadUrl,
            describeDeleteItem: describeDeleteItem,
            deleteItemKey: deleteItemKey,
            describeDeleteFailure: describeDeleteFailure,
            describeDeleteFailureKey: describeDeleteFailureKey,
            formatDateTime: formatDateTime,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
