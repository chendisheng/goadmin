import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchPlugin, updatePlugin } from '@/api/plugins';
import { useAppI18n } from '@/i18n';
import PluginMenuTreeEditor from '@/views/plugin/center/components/PluginMenuTreeEditor.vue';
import { buildPluginPermissionDiffRows, buildPluginPermissionOrphans, clonePluginMenuTree, createPluginMenuNode, createPluginPermissionNode, flattenPluginMenus, generatePluginPermissions, generatePluginPermissionsFromTemplate, groupPluginPermissionPresets, mergePluginPermissions, movePluginMenuNode, normalizePluginMenuTree, pluginPermissionTemplates, readPluginPermissionPresets, removePluginPermissionPreset, savePluginPermissionPreset, } from '@/utils/plugin';
import { formatDateTime } from '@/utils/admin';
const route = useRoute();
const router = useRouter();
const { t } = useAppI18n();
const loading = ref(false);
const saving = ref(false);
const activeTab = ref('overview');
const selectedActions = ref(['view', 'create', 'update', 'delete']);
const selectedTemplateKey = ref('crud');
const presetName = ref('');
const presetSearchQuery = ref('');
const diffFilter = ref('all');
const presets = ref(readPluginPermissionPresets());
const sortNotice = ref('');
const plugin = ref(null);
const actionOptions = computed(() => [
    { label: t('plugin.action.view', 'View'), value: 'view' },
    { label: t('plugin.action.create', 'Create'), value: 'create' },
    { label: t('plugin.action.edit', 'Edit'), value: 'update' },
    { label: t('plugin.action.delete', 'Delete'), value: 'delete' },
]);
const permissionTemplateOptions = computed(() => pluginPermissionTemplates.map((template) => ({
    ...template,
    label: t(`plugin.template.${template.key}.label`, template.label),
    description: t(`plugin.template.${template.key}.description`, template.description),
})));
function defaultForm() {
    return {
        name: '',
        description: '',
        enabled: true,
        menus: [],
        permissions: [],
    };
}
const form = reactive(defaultForm());
const pluginName = computed(() => String(route.params.name || '').trim());
const pageTitle = computed(() => plugin.value?.name || pluginName.value || t('plugin.detail_title', 'Plugin details'));
const menuCount = computed(() => flattenPluginMenus(form.menus).length);
const permissionCount = computed(() => form.permissions.length);
const generatedPermissions = computed(() => generatePluginPermissions(form.name || pluginName.value, form.menus, selectedActions.value));
const generatedTemplatePermissions = computed(() => generatePluginPermissionsFromTemplate(form.name || pluginName.value, form.menus, selectedTemplateKey.value));
const menuPreviewRows = computed(() => flattenPluginMenus(form.menus));
const selectedTemplate = computed(() => permissionTemplateOptions.value.find((item) => item.key === selectedTemplateKey.value) ?? permissionTemplateOptions.value[1]);
const permissionDiffRows = computed(() => buildPluginPermissionDiffRows(form.name || pluginName.value, form.menus, form.permissions, selectedActions.value));
const orphanPermissions = computed(() => buildPluginPermissionOrphans(form.name || pluginName.value, form.menus, form.permissions));
const groupedPresets = computed(() => groupPluginPermissionPresets(presets.value));
const filteredGroupedPresets = computed(() => {
    const query = presetSearchQuery.value.trim().toLowerCase();
    if (query === '') {
        return groupedPresets.value;
    }
    return groupedPresets.value
        .map((group) => {
        const groupName = group.pluginName.toLowerCase();
        const groupMatches = groupName.includes(query);
        const presetsInGroup = groupMatches
            ? group.presets
            : group.presets.filter((preset) => {
                const haystack = [preset.name, preset.templateKey, preset.actions.join(' ')].join(' ').toLowerCase();
                return haystack.includes(query);
            });
        return {
            ...group,
            presets: presetsInGroup,
        };
    })
        .filter((group) => group.presets.length > 0);
});
const filteredPermissionDiffRows = computed(() => {
    if (diffFilter.value === 'missing') {
        return permissionDiffRows.value.filter((row) => row.missingActions.length > 0);
    }
    if (diffFilter.value === 'covered') {
        return permissionDiffRows.value.filter((row) => row.missingActions.length === 0);
    }
    return permissionDiffRows.value;
});
const coverageStats = computed(() => {
    const total = permissionDiffRows.value.length;
    const covered = permissionDiffRows.value.filter((item) => item.missingActions.length === 0).length;
    const missing = permissionDiffRows.value.filter((item) => item.missingActions.length > 0).length;
    const coverageRate = total === 0 ? 0 : Math.round((covered / total) * 100);
    return {
        total,
        covered,
        missing,
        orphan: orphanPermissions.value.length,
        coverageRate,
    };
});
const coverageLevel = computed(() => {
    if (coverageStats.value.coverageRate >= 100) {
        return 'complete';
    }
    if (coverageStats.value.coverageRate >= 75) {
        return 'high';
    }
    if (coverageStats.value.coverageRate >= 40) {
        return 'medium';
    }
    return 'low';
});
const coverageProgressColor = computed(() => {
    if (coverageLevel.value === 'complete') {
        return '#67c23a';
    }
    if (coverageLevel.value === 'high') {
        return '#409eff';
    }
    if (coverageLevel.value === 'medium') {
        return '#e6a23c';
    }
    return '#f56c6c';
});
const coverageLevelLabel = computed(() => {
    if (coverageLevel.value === 'complete') {
        return t('plugin.coverage.complete', 'Fully covered');
    }
    if (coverageLevel.value === 'high') {
        return t('plugin.coverage.high', 'High coverage');
    }
    if (coverageLevel.value === 'medium') {
        return t('plugin.coverage.medium', 'Completing coverage');
    }
    return t('plugin.coverage.low', 'Needs coverage');
});
let lastGeneratedPermissionKeys = new Set();
function buildSortSummary() {
    if (form.menus.length === 0) {
        return t('plugin.sort.empty', 'No menus need sorting yet');
    }
    const summary = form.menus
        .slice(0, 5)
        .map((menu, index) => `${index + 1}. ${t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id || t('plugin.menu_unnamed', 'Unnamed menu'))}`)
        .join(' / ');
    return t('plugin.sort.summary', 'Auto-sorted: {summary}{more}', {
        summary,
        more: form.menus.length > 5 ? ' ...' : '',
    });
}
function ensureSeedRows() {
    if (form.menus.length === 0) {
        form.menus.push(createPluginMenuNode(form.name || pluginName.value));
    }
    if (form.permissions.length === 0) {
        form.permissions.push(createPluginPermissionNode(form.name || pluginName.value));
    }
}
function syncFromPlugin(item) {
    plugin.value = item;
    Object.assign(form, defaultForm(), {
        name: item.name,
        description: item.description ?? '',
        enabled: item.enabled,
        menus: clonePluginMenuTree(item.menus ?? []),
        permissions: (item.permissions ?? []).map((permission) => ({ ...permission })),
    });
    normalizePluginMenuTree(form.menus, item.name);
    ensureSeedRows();
}
async function loadPlugin() {
    if (pluginName.value === '') {
        ElMessage.warning(t('plugin.no_name', 'Plugin name cannot be empty'));
        await router.replace('/system/plugins');
        return;
    }
    loading.value = true;
    try {
        const item = await fetchPlugin(pluginName.value);
        syncFromPlugin(item);
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('plugin.load_failed', 'Failed to load plugin details'));
        await router.replace('/system/plugins');
    }
    finally {
        loading.value = false;
    }
}
function appendPermissionRow() {
    form.permissions.push(createPluginPermissionNode(form.name || pluginName.value));
}
function removePermissionRow(index) {
    form.permissions.splice(index, 1);
}
function fillGeneratedPermissions(permissions = generatedPermissions.value) {
    const generated = permissions;
    if (generated.length === 0) {
        ElMessage.warning(t('plugin.generate_hint', 'Select generation actions first and make sure the menu is not empty'));
        return;
    }
    form.permissions = mergePluginPermissions(form.permissions, generated);
    lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
    ElMessage.success(t('plugin.generated_success', 'Generated {count} permissions', { count: generated.length }));
}
function clearGeneratedPermissions() {
    if (lastGeneratedPermissionKeys.size === 0) {
        return;
    }
    form.permissions = form.permissions.filter((item) => !lastGeneratedPermissionKeys.has(`${item.object}:${item.action}`));
    lastGeneratedPermissionKeys = new Set();
}
function completeDiffRow(row) {
    if (row.missingActions.length === 0) {
        ElMessage.info(t('plugin.coverage.already_complete', 'This row is already covered'));
        return;
    }
    const generated = row.missingActions.map((action) => ({
        plugin: form.name || pluginName.value,
        object: row.object,
        action,
        description: `${row.menuName} ${action}`,
    }));
    form.permissions = mergePluginPermissions(form.permissions, generated);
    lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
    ElMessage.success(t('plugin.coverage.completed_count', 'Completed {count} missing permissions', { count: generated.length }));
}
function completeAllMissingPermissions() {
    const generated = permissionDiffRows.value.flatMap((row) => row.missingActions.map((action) => ({
        plugin: form.name || pluginName.value,
        object: row.object,
        action,
        description: `${row.menuName} ${action}`,
    })));
    if (generated.length === 0) {
        ElMessage.info(t('plugin.coverage.no_missing', 'No missing differences to complete'));
        return;
    }
    form.permissions = mergePluginPermissions(form.permissions, generated);
    lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
    ElMessage.success(t('plugin.coverage.completed_all_count', 'Completed all {count} missing permissions', { count: generated.length }));
}
function refreshPresets() {
    presets.value = readPluginPermissionPresets();
}
function saveCurrentPreset() {
    const name = presetName.value.trim();
    if (name === '') {
        ElMessage.warning(t('plugin.preset_name_required', 'Enter a preset name'));
        return;
    }
    presets.value = savePluginPermissionPreset(form.name || pluginName.value, name, selectedTemplateKey.value, selectedActions.value);
    presetName.value = '';
    ElMessage.success(t('plugin.preset_saved', 'Preset "{name}" saved', { name }));
}
function applyPreset(preset) {
    selectedTemplateKey.value = preset.templateKey || 'crud';
    selectedActions.value = preset.actions.length > 0 ? preset.actions.slice() : ['view'];
    fillGeneratedPermissions(generatePluginPermissions(form.name || pluginName.value, form.menus, selectedActions.value));
}
function deletePreset(presetId) {
    presets.value = removePluginPermissionPreset(presetId);
}
function applyPermissionTemplate(templateKey) {
    const template = pluginPermissionTemplates.find((item) => item.key === templateKey);
    if (!template) {
        return;
    }
    selectedTemplateKey.value = template.key;
    selectedActions.value = template.actions.slice();
    fillGeneratedPermissions(generatePluginPermissions(form.name || pluginName.value, form.menus, template.actions));
}
function handleMoveNode(sourceId, targetId, position) {
    const moved = movePluginMenuNode(form.menus, sourceId, targetId, position);
    if (!moved) {
        ElMessage.warning(t('plugin.menu_move_failed', 'The current menu cannot be moved to the target position'));
        return;
    }
    normalizePluginMenuTree(form.menus, form.name || pluginName.value);
    sortNotice.value = buildSortSummary();
    ElMessage.success(t('plugin.menu_reordered', 'Menu reordered'));
}
async function savePlugin() {
    const name = form.name.trim();
    if (name === '') {
        ElMessage.warning(t('plugin.validation_name', 'Enter the plugin name'));
        return;
    }
    if (form.menus.some((menu) => menu.name.trim() === '' || menu.path.trim() === '')) {
        ElMessage.warning(t('plugin.validation_menu', 'Complete the plugin menu name and path'));
        return;
    }
    if (form.permissions.some((permission) => permission.object.trim() === '' || permission.action.trim() === '')) {
        ElMessage.warning(t('plugin.validation_permission', 'Complete the plugin permission object and action'));
        return;
    }
    saving.value = true;
    try {
        await updatePlugin(pluginName.value, {
            name,
            description: form.description.trim(),
            enabled: Boolean(form.enabled),
            menus: form.menus,
            permissions: form.permissions,
        });
        ElMessage.success(t('plugin.save_success', 'Plugin saved'));
        await loadPlugin();
    }
    finally {
        saving.value = false;
    }
}
function goBack() {
    void router.push('/system/plugins');
}
watch(() => route.params.name, () => {
    void loadPlugin();
}, { immediate: true });
onMounted(() => {
    if (selectedActions.value.length === 0) {
        selectedActions.value = ['view'];
    }
    if (selectedTemplateKey.value === '') {
        selectedTemplateKey.value = 'crud';
    }
    refreshPresets();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__coverage-metrics']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__coverage-metrics']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__coverage-metrics']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-page plugin-detail-page" },
});
/** @type {[typeof AdminTable, typeof AdminTable, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AdminTable, new AdminTable({
    title: (__VLS_ctx.pageTitle),
    description: (__VLS_ctx.t('plugin.detail_description', 'Plugin details, menu-tree editing, and batch permission generation.')),
    loading: (__VLS_ctx.loading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.pageTitle),
    description: (__VLS_ctx.t('plugin.detail_description', 'Plugin details, menu-tree editing, and batch permission generation.')),
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_2.slots.default;
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_3 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_4 = __VLS_asFunctionalComponent(__VLS_3, new __VLS_3({
        ...{ 'onClick': {} },
    }));
    const __VLS_5 = __VLS_4({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_4));
    let __VLS_7;
    let __VLS_8;
    let __VLS_9;
    const __VLS_10 = {
        onClick: (__VLS_ctx.goBack)
    };
    __VLS_6.slots.default;
    (__VLS_ctx.t('common.back', 'Back'));
    var __VLS_6;
    const __VLS_11 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.loading),
    }));
    const __VLS_13 = __VLS_12({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    let __VLS_15;
    let __VLS_16;
    let __VLS_17;
    const __VLS_18 = {
        onClick: (__VLS_ctx.loadPlugin)
    };
    __VLS_14.slots.default;
    (__VLS_ctx.t('common.refresh', 'Refresh'));
    var __VLS_14;
    const __VLS_19 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.saving),
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (__VLS_ctx.savePlugin)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('plugin:update') }, null, null);
    __VLS_22.slots.default;
    (__VLS_ctx.t('plugin.save_plugin', 'Save plugin'));
    var __VLS_22;
}
const __VLS_27 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
    gutter: (16),
    ...{ class: "mb-16" },
}));
const __VLS_29 = __VLS_28({
    gutter: (16),
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_28));
__VLS_30.slots.default;
const __VLS_31 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
    xs: (24),
    md: (8),
}));
const __VLS_33 = __VLS_32({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_32));
__VLS_34.slots.default;
const __VLS_35 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
    shadow: "never",
}));
const __VLS_37 = __VLS_36({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_36));
__VLS_38.slots.default;
const __VLS_39 = {}.ElStatistic;
/** @type {[typeof __VLS_components.ElStatistic, typeof __VLS_components.elStatistic, ]} */ ;
// @ts-ignore
const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
    title: (__VLS_ctx.t('plugin.menu_nodes', 'Menu nodes')),
    value: (__VLS_ctx.menuCount),
}));
const __VLS_41 = __VLS_40({
    title: (__VLS_ctx.t('plugin.menu_nodes', 'Menu nodes')),
    value: (__VLS_ctx.menuCount),
}, ...__VLS_functionalComponentArgsRest(__VLS_40));
var __VLS_38;
var __VLS_34;
const __VLS_43 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
    xs: (24),
    md: (8),
}));
const __VLS_45 = __VLS_44({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_44));
__VLS_46.slots.default;
const __VLS_47 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
    shadow: "never",
}));
const __VLS_49 = __VLS_48({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
__VLS_50.slots.default;
const __VLS_51 = {}.ElStatistic;
/** @type {[typeof __VLS_components.ElStatistic, typeof __VLS_components.elStatistic, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    title: (__VLS_ctx.t('plugin.permission_items', 'Permission items')),
    value: (__VLS_ctx.permissionCount),
}));
const __VLS_53 = __VLS_52({
    title: (__VLS_ctx.t('plugin.permission_items', 'Permission items')),
    value: (__VLS_ctx.permissionCount),
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
var __VLS_50;
var __VLS_46;
const __VLS_55 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
    xs: (24),
    md: (8),
}));
const __VLS_57 = __VLS_56({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
__VLS_58.slots.default;
const __VLS_59 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    shadow: "never",
}));
const __VLS_61 = __VLS_60({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElStatistic;
/** @type {[typeof __VLS_components.ElStatistic, typeof __VLS_components.elStatistic, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    title: (__VLS_ctx.t('plugin.status', 'Status')),
    value: (__VLS_ctx.form.enabled ? 1 : 0),
    formatter: ((value) => (value === 1 ? __VLS_ctx.t('plugin.enabled', 'Enabled') : __VLS_ctx.t('plugin.disabled', 'Disabled'))),
}));
const __VLS_65 = __VLS_64({
    title: (__VLS_ctx.t('plugin.status', 'Status')),
    value: (__VLS_ctx.form.enabled ? 1 : 0),
    formatter: ((value) => (value === 1 ? __VLS_ctx.t('plugin.enabled', 'Enabled') : __VLS_ctx.t('plugin.disabled', 'Disabled'))),
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
var __VLS_62;
var __VLS_58;
var __VLS_30;
const __VLS_67 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    modelValue: (__VLS_ctx.activeTab),
}));
const __VLS_69 = __VLS_68({
    modelValue: (__VLS_ctx.activeTab),
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
const __VLS_71 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    label: (__VLS_ctx.t('plugin.detail_overview_tab', 'Basic info')),
    name: "overview",
}));
const __VLS_73 = __VLS_72({
    label: (__VLS_ctx.t('plugin.detail_overview_tab', 'Basic info')),
    name: "overview",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
const __VLS_75 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    shadow: "never",
}));
const __VLS_77 = __VLS_76({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
const __VLS_79 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    labelWidth: "120px",
    ...{ class: "admin-form admin-form--two-col" },
}));
const __VLS_81 = __VLS_80({
    labelWidth: "120px",
    ...{ class: "admin-form admin-form--two-col" },
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
const __VLS_83 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    label: (__VLS_ctx.t('plugin.name', 'Plugin name')),
}));
const __VLS_85 = __VLS_84({
    label: (__VLS_ctx.t('plugin.name', 'Plugin name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
const __VLS_87 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    modelValue: (__VLS_ctx.form.name),
    disabled: true,
}));
const __VLS_89 = __VLS_88({
    modelValue: (__VLS_ctx.form.name),
    disabled: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
var __VLS_86;
const __VLS_91 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: (__VLS_ctx.t('plugin.enabled_status', 'Enabled status')),
}));
const __VLS_93 = __VLS_92({
    label: (__VLS_ctx.t('plugin.enabled_status', 'Enabled status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
const __VLS_95 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    modelValue: (__VLS_ctx.form.enabled),
}));
const __VLS_97 = __VLS_96({
    modelValue: (__VLS_ctx.form.enabled),
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
var __VLS_94;
const __VLS_99 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    label: (__VLS_ctx.t('plugin.description_label', 'Plugin description')),
    ...{ class: "admin-form__full-row" },
}));
const __VLS_101 = __VLS_100({
    label: (__VLS_ctx.t('plugin.description_label', 'Plugin description')),
    ...{ class: "admin-form__full-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_102.slots.default;
const __VLS_103 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('plugin.description_placeholder', 'Enter plugin description')),
}));
const __VLS_105 = __VLS_104({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('plugin.description_placeholder', 'Enter plugin description')),
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
var __VLS_102;
const __VLS_107 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    label: (__VLS_ctx.t('plugin.created_at', 'Created at')),
}));
const __VLS_109 = __VLS_108({
    label: (__VLS_ctx.t('plugin.created_at', 'Created at')),
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.plugin?.created_at ? __VLS_ctx.formatDateTime(__VLS_ctx.plugin.created_at) : '-');
var __VLS_110;
const __VLS_111 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('plugin.updated_at', 'Updated at')),
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('plugin.updated_at', 'Updated at')),
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
__VLS_114.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.plugin?.updated_at ? __VLS_ctx.formatDateTime(__VLS_ctx.plugin.updated_at) : '-');
var __VLS_114;
const __VLS_115 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
    label: (__VLS_ctx.t('plugin.menu_tree_total', 'Menu tree total')),
}));
const __VLS_117 = __VLS_116({
    label: (__VLS_ctx.t('plugin.menu_tree_total', 'Menu tree total')),
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
__VLS_118.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.menuCount);
var __VLS_118;
const __VLS_119 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
    label: (__VLS_ctx.t('plugin.permission_total', 'Permission total')),
}));
const __VLS_121 = __VLS_120({
    label: (__VLS_ctx.t('plugin.permission_total', 'Permission total')),
}, ...__VLS_functionalComponentArgsRest(__VLS_120));
__VLS_122.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.permissionCount);
var __VLS_122;
var __VLS_82;
var __VLS_78;
var __VLS_74;
const __VLS_123 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
    label: (__VLS_ctx.t('plugin.menu_tree_editor_tab', 'Menu tree editor')),
    name: "menus",
}));
const __VLS_125 = __VLS_124({
    label: (__VLS_ctx.t('plugin.menu_tree_editor_tab', 'Menu tree editor')),
    name: "menus",
}, ...__VLS_functionalComponentArgsRest(__VLS_124));
__VLS_126.slots.default;
const __VLS_127 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
    shadow: "never",
}));
const __VLS_129 = __VLS_128({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_128));
__VLS_130.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_130.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.menu_tree_editor_title', 'Menu tree editor'));
    const __VLS_131 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
        wrap: true,
    }));
    const __VLS_133 = __VLS_132({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_132));
    __VLS_134.slots.default;
    const __VLS_135 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
        effect: "plain",
        type: "success",
    }));
    const __VLS_137 = __VLS_136({
        effect: "plain",
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_136));
    __VLS_138.slots.default;
    (__VLS_ctx.t('plugin.recursive_edit', 'Recursive editing'));
    var __VLS_138;
    const __VLS_139 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
        effect: "plain",
        type: "info",
    }));
    const __VLS_141 = __VLS_140({
        effect: "plain",
        type: "info",
    }, ...__VLS_functionalComponentArgsRest(__VLS_140));
    __VLS_142.slots.default;
    (__VLS_ctx.t('plugin.drag_sorting', 'Drag sorting supported'));
    var __VLS_142;
    var __VLS_134;
}
const __VLS_143 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    title: (__VLS_ctx.t('plugin.drag_instructions_title', 'Drag instructions')),
    description: (__VLS_ctx.t('plugin.drag_instructions_description', 'Drag the menu card into the before/after/inside drop zones to adjust the tree structure. Menu changes update the permission-linked preview in real time.')),
    type: "info",
    showIcon: true,
    closable: (false),
    ...{ class: "mb-12" },
}));
const __VLS_145 = __VLS_144({
    title: (__VLS_ctx.t('plugin.drag_instructions_title', 'Drag instructions')),
    description: (__VLS_ctx.t('plugin.drag_instructions_description', 'Drag the menu card into the before/after/inside drop zones to adjust the tree structure. Menu changes update the permission-linked preview in real time.')),
    type: "info",
    showIcon: true,
    closable: (false),
    ...{ class: "mb-12" },
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
if (__VLS_ctx.sortNotice) {
    const __VLS_147 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
        title: (__VLS_ctx.sortNotice),
        description: (__VLS_ctx.t('plugin.sort_notice', 'The menu hierarchy has been renumbered automatically. Saving will submit using the latest order.')),
        type: "success",
        showIcon: true,
        closable: (false),
        ...{ class: "mb-12" },
    }));
    const __VLS_149 = __VLS_148({
        title: (__VLS_ctx.sortNotice),
        description: (__VLS_ctx.t('plugin.sort_notice', 'The menu hierarchy has been renumbered automatically. Saving will submit using the latest order.')),
        type: "success",
        showIcon: true,
        closable: (false),
        ...{ class: "mb-12" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_148));
}
/** @type {[typeof PluginMenuTreeEditor, ]} */ ;
// @ts-ignore
const __VLS_151 = __VLS_asFunctionalComponent(PluginMenuTreeEditor, new PluginMenuTreeEditor({
    ...{ 'onMoveNode': {} },
    menus: (__VLS_ctx.form.menus),
    pluginName: (__VLS_ctx.form.name || __VLS_ctx.pluginName),
}));
const __VLS_152 = __VLS_151({
    ...{ 'onMoveNode': {} },
    menus: (__VLS_ctx.form.menus),
    pluginName: (__VLS_ctx.form.name || __VLS_ctx.pluginName),
}, ...__VLS_functionalComponentArgsRest(__VLS_151));
let __VLS_154;
let __VLS_155;
let __VLS_156;
const __VLS_157 = {
    onMoveNode: (__VLS_ctx.handleMoveNode)
};
var __VLS_153;
var __VLS_130;
var __VLS_126;
const __VLS_158 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_159 = __VLS_asFunctionalComponent(__VLS_158, new __VLS_158({
    label: (__VLS_ctx.t('plugin.permissions_batch_tab', 'Batch permission generation')),
    name: "permissions",
}));
const __VLS_160 = __VLS_159({
    label: (__VLS_ctx.t('plugin.permissions_batch_tab', 'Batch permission generation')),
    name: "permissions",
}, ...__VLS_functionalComponentArgsRest(__VLS_159));
__VLS_161.slots.default;
const __VLS_162 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_163 = __VLS_asFunctionalComponent(__VLS_162, new __VLS_162({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_164 = __VLS_163({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_163));
__VLS_165.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_165.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.permission_template_header', 'Generate permission template'));
    const __VLS_166 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_167 = __VLS_asFunctionalComponent(__VLS_166, new __VLS_166({
        effect: "plain",
        type: "success",
    }));
    const __VLS_168 = __VLS_167({
        effect: "plain",
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_167));
    __VLS_169.slots.default;
    (__VLS_ctx.t('plugin.generate_tag', 'Generate'));
    var __VLS_169;
}
const __VLS_170 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_171 = __VLS_asFunctionalComponent(__VLS_170, new __VLS_170({
    wrap: true,
    ...{ class: "mb-12" },
}));
const __VLS_172 = __VLS_171({
    wrap: true,
    ...{ class: "mb-12" },
}, ...__VLS_functionalComponentArgsRest(__VLS_171));
__VLS_173.slots.default;
for (const [template] of __VLS_getVForSourceType((__VLS_ctx.permissionTemplateOptions))) {
    const __VLS_174 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_175 = __VLS_asFunctionalComponent(__VLS_174, new __VLS_174({
        ...{ 'onClick': {} },
        key: (template.key),
        type: (__VLS_ctx.selectedTemplateKey === template.key ? 'primary' : 'default'),
        plain: true,
    }));
    const __VLS_176 = __VLS_175({
        ...{ 'onClick': {} },
        key: (template.key),
        type: (__VLS_ctx.selectedTemplateKey === template.key ? 'primary' : 'default'),
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_175));
    let __VLS_178;
    let __VLS_179;
    let __VLS_180;
    const __VLS_181 = {
        onClick: (...[$event]) => {
            __VLS_ctx.applyPermissionTemplate(template.key);
        }
    };
    __VLS_177.slots.default;
    (template.label);
    var __VLS_177;
}
var __VLS_173;
const __VLS_182 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_183 = __VLS_asFunctionalComponent(__VLS_182, new __VLS_182({
    gutter: (16),
    ...{ class: "mb-16" },
}));
const __VLS_184 = __VLS_183({
    gutter: (16),
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_183));
__VLS_185.slots.default;
const __VLS_186 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_187 = __VLS_asFunctionalComponent(__VLS_186, new __VLS_186({
    xs: (24),
    md: (12),
}));
const __VLS_188 = __VLS_187({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_187));
__VLS_189.slots.default;
const __VLS_190 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_191 = __VLS_asFunctionalComponent(__VLS_190, new __VLS_190({
    shadow: "never",
}));
const __VLS_192 = __VLS_191({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_191));
__VLS_193.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_193.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.preset_save_title', 'Save as preset'));
    const __VLS_194 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_195 = __VLS_asFunctionalComponent(__VLS_194, new __VLS_194({
        effect: "plain",
        type: "info",
    }));
    const __VLS_196 = __VLS_195({
        effect: "plain",
        type: "info",
    }, ...__VLS_functionalComponentArgsRest(__VLS_195));
    __VLS_197.slots.default;
    (__VLS_ctx.t('plugin.local_storage', 'Local storage'));
    var __VLS_197;
}
const __VLS_198 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_199 = __VLS_asFunctionalComponent(__VLS_198, new __VLS_198({
    labelWidth: "92px",
    ...{ class: "admin-form" },
}));
const __VLS_200 = __VLS_199({
    labelWidth: "92px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_199));
__VLS_201.slots.default;
const __VLS_202 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_203 = __VLS_asFunctionalComponent(__VLS_202, new __VLS_202({
    label: (__VLS_ctx.t('plugin.preset_name', 'Preset name')),
}));
const __VLS_204 = __VLS_203({
    label: (__VLS_ctx.t('plugin.preset_name', 'Preset name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_203));
__VLS_205.slots.default;
const __VLS_206 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_207 = __VLS_asFunctionalComponent(__VLS_206, new __VLS_206({
    modelValue: (__VLS_ctx.presetName),
    placeholder: (__VLS_ctx.t('plugin.preset_name_placeholder', 'For example: plugin detail CRUD preset')),
}));
const __VLS_208 = __VLS_207({
    modelValue: (__VLS_ctx.presetName),
    placeholder: (__VLS_ctx.t('plugin.preset_name_placeholder', 'For example: plugin detail CRUD preset')),
}, ...__VLS_functionalComponentArgsRest(__VLS_207));
var __VLS_205;
const __VLS_210 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_211 = __VLS_asFunctionalComponent(__VLS_210, new __VLS_210({
    label: (__VLS_ctx.t('plugin.current_template', 'Current template')),
}));
const __VLS_212 = __VLS_211({
    label: (__VLS_ctx.t('plugin.current_template', 'Current template')),
}, ...__VLS_functionalComponentArgsRest(__VLS_211));
__VLS_213.slots.default;
const __VLS_214 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_215 = __VLS_asFunctionalComponent(__VLS_214, new __VLS_214({
    effect: "plain",
}));
const __VLS_216 = __VLS_215({
    effect: "plain",
}, ...__VLS_functionalComponentArgsRest(__VLS_215));
__VLS_217.slots.default;
(__VLS_ctx.selectedTemplate.label);
var __VLS_217;
var __VLS_213;
const __VLS_218 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_219 = __VLS_asFunctionalComponent(__VLS_218, new __VLS_218({
    label: (__VLS_ctx.t('plugin.action_set', 'Action set')),
}));
const __VLS_220 = __VLS_219({
    label: (__VLS_ctx.t('plugin.action_set', 'Action set')),
}, ...__VLS_functionalComponentArgsRest(__VLS_219));
__VLS_221.slots.default;
const __VLS_222 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_223 = __VLS_asFunctionalComponent(__VLS_222, new __VLS_222({
    wrap: true,
}));
const __VLS_224 = __VLS_223({
    wrap: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_223));
__VLS_225.slots.default;
for (const [action] of __VLS_getVForSourceType((__VLS_ctx.selectedActions))) {
    const __VLS_226 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_227 = __VLS_asFunctionalComponent(__VLS_226, new __VLS_226({
        key: (action),
        effect: "plain",
    }));
    const __VLS_228 = __VLS_227({
        key: (action),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_227));
    __VLS_229.slots.default;
    (action);
    var __VLS_229;
}
var __VLS_225;
var __VLS_221;
const __VLS_230 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_231 = __VLS_asFunctionalComponent(__VLS_230, new __VLS_230({}));
const __VLS_232 = __VLS_231({}, ...__VLS_functionalComponentArgsRest(__VLS_231));
__VLS_233.slots.default;
const __VLS_234 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_235 = __VLS_asFunctionalComponent(__VLS_234, new __VLS_234({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_236 = __VLS_235({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_235));
let __VLS_238;
let __VLS_239;
let __VLS_240;
const __VLS_241 = {
    onClick: (__VLS_ctx.saveCurrentPreset)
};
__VLS_237.slots.default;
(__VLS_ctx.t('plugin.save_current_preset', 'Save current config'));
var __VLS_237;
const __VLS_242 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_243 = __VLS_asFunctionalComponent(__VLS_242, new __VLS_242({
    ...{ 'onClick': {} },
}));
const __VLS_244 = __VLS_243({
    ...{ 'onClick': {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_243));
let __VLS_246;
let __VLS_247;
let __VLS_248;
const __VLS_249 = {
    onClick: (__VLS_ctx.refreshPresets)
};
__VLS_245.slots.default;
(__VLS_ctx.t('plugin.refresh_presets', 'Refresh presets'));
var __VLS_245;
var __VLS_233;
var __VLS_201;
var __VLS_193;
var __VLS_189;
const __VLS_250 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_251 = __VLS_asFunctionalComponent(__VLS_250, new __VLS_250({
    xs: (24),
    md: (12),
}));
const __VLS_252 = __VLS_251({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_251));
__VLS_253.slots.default;
const __VLS_254 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_255 = __VLS_asFunctionalComponent(__VLS_254, new __VLS_254({
    shadow: "never",
}));
const __VLS_256 = __VLS_255({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_255));
__VLS_257.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_257.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header plugin-detail-page__preset-header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.existing_presets', 'Existing presets'));
    const __VLS_258 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_259 = __VLS_asFunctionalComponent(__VLS_258, new __VLS_258({
        effect: "plain",
        type: "success",
    }));
    const __VLS_260 = __VLS_259({
        effect: "plain",
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_259));
    __VLS_261.slots.default;
    (__VLS_ctx.presets.length);
    var __VLS_261;
    const __VLS_262 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_263 = __VLS_asFunctionalComponent(__VLS_262, new __VLS_262({
        modelValue: (__VLS_ctx.presetSearchQuery),
        clearable: true,
        size: "small",
        placeholder: (__VLS_ctx.t('plugin.preset_search_placeholder', 'Search plugin or preset names')),
        ...{ class: "plugin-detail-page__preset-search" },
    }));
    const __VLS_264 = __VLS_263({
        modelValue: (__VLS_ctx.presetSearchQuery),
        clearable: true,
        size: "small",
        placeholder: (__VLS_ctx.t('plugin.preset_search_placeholder', 'Search plugin or preset names')),
        ...{ class: "plugin-detail-page__preset-search" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_263));
}
if (__VLS_ctx.presets.length === 0) {
    const __VLS_266 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_267 = __VLS_asFunctionalComponent(__VLS_266, new __VLS_266({
        description: (__VLS_ctx.t('plugin.no_presets', 'No presets yet, save one template configuration first')),
    }));
    const __VLS_268 = __VLS_267({
        description: (__VLS_ctx.t('plugin.no_presets', 'No presets yet, save one template configuration first')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_267));
}
else if (__VLS_ctx.filteredGroupedPresets.length === 0) {
    const __VLS_270 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_271 = __VLS_asFunctionalComponent(__VLS_270, new __VLS_270({
        description: (__VLS_ctx.t('plugin.no_matching_presets', 'No matching presets found')),
    }));
    const __VLS_272 = __VLS_271({
        description: (__VLS_ctx.t('plugin.no_matching_presets', 'No matching presets found')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_271));
}
else {
    const __VLS_274 = {}.ElCollapse;
    /** @type {[typeof __VLS_components.ElCollapse, typeof __VLS_components.elCollapse, typeof __VLS_components.ElCollapse, typeof __VLS_components.elCollapse, ]} */ ;
    // @ts-ignore
    const __VLS_275 = __VLS_asFunctionalComponent(__VLS_274, new __VLS_274({
        accordion: true,
        ...{ class: "plugin-detail-page__preset-groups" },
    }));
    const __VLS_276 = __VLS_275({
        accordion: true,
        ...{ class: "plugin-detail-page__preset-groups" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_275));
    __VLS_277.slots.default;
    for (const [group] of __VLS_getVForSourceType((__VLS_ctx.filteredGroupedPresets))) {
        const __VLS_278 = {}.ElCollapseItem;
        /** @type {[typeof __VLS_components.ElCollapseItem, typeof __VLS_components.elCollapseItem, typeof __VLS_components.ElCollapseItem, typeof __VLS_components.elCollapseItem, ]} */ ;
        // @ts-ignore
        const __VLS_279 = __VLS_asFunctionalComponent(__VLS_278, new __VLS_278({
            key: (group.pluginName),
            name: (group.pluginName),
        }));
        const __VLS_280 = __VLS_279({
            key: (group.pluginName),
            name: (group.pluginName),
        }, ...__VLS_functionalComponentArgsRest(__VLS_279));
        __VLS_281.slots.default;
        {
            const { title: __VLS_thisSlot } = __VLS_281.slots;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "plugin-detail-page__group-title" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
            (group.pluginName);
            const __VLS_282 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_283 = __VLS_asFunctionalComponent(__VLS_282, new __VLS_282({
                effect: "plain",
                size: "small",
            }));
            const __VLS_284 = __VLS_283({
                effect: "plain",
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_283));
            __VLS_285.slots.default;
            (__VLS_ctx.t('plugin.preset_count', '{count} presets', { count: group.presets.length }));
            var __VLS_285;
        }
        const __VLS_286 = {}.ElSpace;
        /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
        // @ts-ignore
        const __VLS_287 = __VLS_asFunctionalComponent(__VLS_286, new __VLS_286({
            direction: "vertical",
            fill: true,
            ...{ style: {} },
        }));
        const __VLS_288 = __VLS_287({
            direction: "vertical",
            fill: true,
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_287));
        __VLS_289.slots.default;
        for (const [preset] of __VLS_getVForSourceType((group.presets))) {
            const __VLS_290 = {}.ElCard;
            /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
            // @ts-ignore
            const __VLS_291 = __VLS_asFunctionalComponent(__VLS_290, new __VLS_290({
                key: (preset.id),
                shadow: "never",
                ...{ class: "plugin-detail-page__preset-card" },
            }));
            const __VLS_292 = __VLS_291({
                key: (preset.id),
                shadow: "never",
                ...{ class: "plugin-detail-page__preset-card" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_291));
            __VLS_293.slots.default;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "page-card__header" },
            });
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
            __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
            (preset.name);
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                ...{ class: "plugin-detail-page__preset-meta" },
            });
            const __VLS_294 = {}.ElTag;
            /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
            // @ts-ignore
            const __VLS_295 = __VLS_asFunctionalComponent(__VLS_294, new __VLS_294({
                effect: "plain",
                size: "small",
            }));
            const __VLS_296 = __VLS_295({
                effect: "plain",
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_295));
            __VLS_297.slots.default;
            (preset.templateKey);
            var __VLS_297;
            __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
            (preset.actions.join(', ') || __VLS_ctx.t('plugin.no_actions', 'No actions'));
            const __VLS_298 = {}.ElSpace;
            /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
            // @ts-ignore
            const __VLS_299 = __VLS_asFunctionalComponent(__VLS_298, new __VLS_298({}));
            const __VLS_300 = __VLS_299({}, ...__VLS_functionalComponentArgsRest(__VLS_299));
            __VLS_301.slots.default;
            const __VLS_302 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_303 = __VLS_asFunctionalComponent(__VLS_302, new __VLS_302({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }));
            const __VLS_304 = __VLS_303({
                ...{ 'onClick': {} },
                size: "small",
                type: "primary",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_303));
            let __VLS_306;
            let __VLS_307;
            let __VLS_308;
            const __VLS_309 = {
                onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.presets.length === 0))
                        return;
                    if (!!(__VLS_ctx.filteredGroupedPresets.length === 0))
                        return;
                    __VLS_ctx.applyPreset(preset);
                }
            };
            __VLS_305.slots.default;
            (__VLS_ctx.t('plugin.apply_preset', 'Apply'));
            var __VLS_305;
            const __VLS_310 = {}.ElButton;
            /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
            // @ts-ignore
            const __VLS_311 = __VLS_asFunctionalComponent(__VLS_310, new __VLS_310({
                ...{ 'onClick': {} },
                size: "small",
                type: "danger",
                plain: true,
            }));
            const __VLS_312 = __VLS_311({
                ...{ 'onClick': {} },
                size: "small",
                type: "danger",
                plain: true,
            }, ...__VLS_functionalComponentArgsRest(__VLS_311));
            let __VLS_314;
            let __VLS_315;
            let __VLS_316;
            const __VLS_317 = {
                onClick: (...[$event]) => {
                    if (!!(__VLS_ctx.presets.length === 0))
                        return;
                    if (!!(__VLS_ctx.filteredGroupedPresets.length === 0))
                        return;
                    __VLS_ctx.deletePreset(preset.id);
                }
            };
            __VLS_313.slots.default;
            (__VLS_ctx.t('common.delete', 'Delete'));
            var __VLS_313;
            var __VLS_301;
            var __VLS_293;
        }
        var __VLS_289;
        var __VLS_281;
    }
    var __VLS_277;
}
var __VLS_257;
var __VLS_253;
var __VLS_185;
const __VLS_318 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_319 = __VLS_asFunctionalComponent(__VLS_318, new __VLS_318({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_320 = __VLS_319({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_319));
__VLS_321.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_321.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.coverage_preview_title', 'Menu / permission linked preview'));
    const __VLS_322 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_323 = __VLS_asFunctionalComponent(__VLS_322, new __VLS_322({
        effect: "plain",
        type: (__VLS_ctx.coverageLevel === 'complete' ? 'success' : __VLS_ctx.coverageLevel === 'high' ? 'primary' : __VLS_ctx.coverageLevel === 'medium' ? 'warning' : 'danger'),
    }));
    const __VLS_324 = __VLS_323({
        effect: "plain",
        type: (__VLS_ctx.coverageLevel === 'complete' ? 'success' : __VLS_ctx.coverageLevel === 'high' ? 'primary' : __VLS_ctx.coverageLevel === 'medium' ? 'warning' : 'danger'),
    }, ...__VLS_functionalComponentArgsRest(__VLS_323));
    __VLS_325.slots.default;
    (__VLS_ctx.t('plugin.live_update', 'Live update'));
    var __VLS_325;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "plugin-detail-page__coverage-visual mb-12" },
});
const __VLS_326 = {}.ElProgress;
/** @type {[typeof __VLS_components.ElProgress, typeof __VLS_components.elProgress, ]} */ ;
// @ts-ignore
const __VLS_327 = __VLS_asFunctionalComponent(__VLS_326, new __VLS_326({
    type: "dashboard",
    percentage: (__VLS_ctx.coverageStats.coverageRate),
    color: (__VLS_ctx.coverageProgressColor),
    strokeWidth: (12),
}));
const __VLS_328 = __VLS_327({
    type: "dashboard",
    percentage: (__VLS_ctx.coverageStats.coverageRate),
    color: (__VLS_ctx.coverageProgressColor),
    strokeWidth: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_327));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "plugin-detail-page__coverage-metrics" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: ({ borderColor: __VLS_ctx.coverageProgressColor }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.coverageStats.covered);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('plugin.coverage.covered', 'Covered'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: ({ borderColor: __VLS_ctx.coverageProgressColor }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.coverageStats.missing);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('plugin.coverage.missing', 'Missing'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ style: ({ borderColor: __VLS_ctx.coverageProgressColor }) },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.coverageStats.orphan);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('plugin.coverage.orphan', 'Orphan permissions'));
const __VLS_330 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_331 = __VLS_asFunctionalComponent(__VLS_330, new __VLS_330({
    title: (__VLS_ctx.coverageLevelLabel),
    description: (__VLS_ctx.t('plugin.coverage.rate_description', 'Current coverage rate: {rate}%', { rate: __VLS_ctx.coverageStats.coverageRate })),
    type: (__VLS_ctx.coverageLevel === 'complete' ? 'success' : __VLS_ctx.coverageLevel === 'high' ? 'info' : __VLS_ctx.coverageLevel === 'medium' ? 'warning' : 'error'),
    showIcon: true,
    closable: (false),
    ...{ class: "mb-12" },
}));
const __VLS_332 = __VLS_331({
    title: (__VLS_ctx.coverageLevelLabel),
    description: (__VLS_ctx.t('plugin.coverage.rate_description', 'Current coverage rate: {rate}%', { rate: __VLS_ctx.coverageStats.coverageRate })),
    type: (__VLS_ctx.coverageLevel === 'complete' ? 'success' : __VLS_ctx.coverageLevel === 'high' ? 'info' : __VLS_ctx.coverageLevel === 'medium' ? 'warning' : 'error'),
    showIcon: true,
    closable: (false),
    ...{ class: "mb-12" },
}, ...__VLS_functionalComponentArgsRest(__VLS_331));
const __VLS_334 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_335 = __VLS_asFunctionalComponent(__VLS_334, new __VLS_334({
    column: (2),
    border: true,
    size: "small",
    ...{ class: "mb-12" },
}));
const __VLS_336 = __VLS_335({
    column: (2),
    border: true,
    size: "small",
    ...{ class: "mb-12" },
}, ...__VLS_functionalComponentArgsRest(__VLS_335));
__VLS_337.slots.default;
const __VLS_338 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_339 = __VLS_asFunctionalComponent(__VLS_338, new __VLS_338({
    label: (__VLS_ctx.t('plugin.menu_count', 'Menu count')),
}));
const __VLS_340 = __VLS_339({
    label: (__VLS_ctx.t('plugin.menu_count', 'Menu count')),
}, ...__VLS_functionalComponentArgsRest(__VLS_339));
__VLS_341.slots.default;
(__VLS_ctx.menuPreviewRows.length);
var __VLS_341;
const __VLS_342 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_343 = __VLS_asFunctionalComponent(__VLS_342, new __VLS_342({
    label: (__VLS_ctx.t('plugin.template_action_count', 'Template action count')),
}));
const __VLS_344 = __VLS_343({
    label: (__VLS_ctx.t('plugin.template_action_count', 'Template action count')),
}, ...__VLS_functionalComponentArgsRest(__VLS_343));
__VLS_345.slots.default;
(__VLS_ctx.selectedActions.length);
var __VLS_345;
const __VLS_346 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_347 = __VLS_asFunctionalComponent(__VLS_346, new __VLS_346({
    label: (__VLS_ctx.t('plugin.template_permission_count', 'Template permission count')),
}));
const __VLS_348 = __VLS_347({
    label: (__VLS_ctx.t('plugin.template_permission_count', 'Template permission count')),
}, ...__VLS_functionalComponentArgsRest(__VLS_347));
__VLS_349.slots.default;
(__VLS_ctx.generatedTemplatePermissions.length);
var __VLS_349;
const __VLS_350 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_351 = __VLS_asFunctionalComponent(__VLS_350, new __VLS_350({
    label: (__VLS_ctx.t('plugin.current_permission_count', 'Current permission count')),
}));
const __VLS_352 = __VLS_351({
    label: (__VLS_ctx.t('plugin.current_permission_count', 'Current permission count')),
}, ...__VLS_functionalComponentArgsRest(__VLS_351));
__VLS_353.slots.default;
(__VLS_ctx.permissionCount);
var __VLS_353;
var __VLS_337;
const __VLS_354 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_355 = __VLS_asFunctionalComponent(__VLS_354, new __VLS_354({
    data: (__VLS_ctx.generatedTemplatePermissions),
    border: true,
    size: "small",
}));
const __VLS_356 = __VLS_355({
    data: (__VLS_ctx.generatedTemplatePermissions),
    border: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_355));
__VLS_357.slots.default;
const __VLS_358 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_359 = __VLS_asFunctionalComponent(__VLS_358, new __VLS_358({
    prop: "object",
    label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
    minWidth: "220",
}));
const __VLS_360 = __VLS_359({
    prop: "object",
    label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
    minWidth: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_359));
const __VLS_362 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_363 = __VLS_asFunctionalComponent(__VLS_362, new __VLS_362({
    prop: "action",
    label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
    width: "120",
}));
const __VLS_364 = __VLS_363({
    prop: "action",
    label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_363));
const __VLS_366 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_367 = __VLS_asFunctionalComponent(__VLS_366, new __VLS_366({
    prop: "description",
    label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
    minWidth: "220",
}));
const __VLS_368 = __VLS_367({
    prop: "description",
    label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
    minWidth: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_367));
var __VLS_357;
var __VLS_321;
var __VLS_165;
const __VLS_370 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_371 = __VLS_asFunctionalComponent(__VLS_370, new __VLS_370({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_372 = __VLS_371({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_371));
__VLS_373.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_373.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.coverage_diff_title', 'Menu / permission diff comparison'));
    const __VLS_374 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_375 = __VLS_asFunctionalComponent(__VLS_374, new __VLS_374({
        wrap: true,
    }));
    const __VLS_376 = __VLS_375({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_375));
    __VLS_377.slots.default;
    const __VLS_378 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_379 = __VLS_asFunctionalComponent(__VLS_378, new __VLS_378({
        effect: "plain",
        type: "warning",
    }));
    const __VLS_380 = __VLS_379({
        effect: "plain",
        type: "warning",
    }, ...__VLS_functionalComponentArgsRest(__VLS_379));
    __VLS_381.slots.default;
    (__VLS_ctx.t('plugin.coverage.missing_items', '{count} items missing', { count: __VLS_ctx.coverageStats.missing }));
    var __VLS_381;
    const __VLS_382 = {}.ElRadioGroup;
    /** @type {[typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, typeof __VLS_components.ElRadioGroup, typeof __VLS_components.elRadioGroup, ]} */ ;
    // @ts-ignore
    const __VLS_383 = __VLS_asFunctionalComponent(__VLS_382, new __VLS_382({
        modelValue: (__VLS_ctx.diffFilter),
        size: "small",
    }));
    const __VLS_384 = __VLS_383({
        modelValue: (__VLS_ctx.diffFilter),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_383));
    __VLS_385.slots.default;
    const __VLS_386 = {}.ElRadioButton;
    /** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
    // @ts-ignore
    const __VLS_387 = __VLS_asFunctionalComponent(__VLS_386, new __VLS_386({
        label: "all",
    }));
    const __VLS_388 = __VLS_387({
        label: "all",
    }, ...__VLS_functionalComponentArgsRest(__VLS_387));
    __VLS_389.slots.default;
    (__VLS_ctx.t('common.all', 'All'));
    var __VLS_389;
    const __VLS_390 = {}.ElRadioButton;
    /** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
    // @ts-ignore
    const __VLS_391 = __VLS_asFunctionalComponent(__VLS_390, new __VLS_390({
        label: "missing",
    }));
    const __VLS_392 = __VLS_391({
        label: "missing",
    }, ...__VLS_functionalComponentArgsRest(__VLS_391));
    __VLS_393.slots.default;
    (__VLS_ctx.t('plugin.coverage.missing', 'Missing'));
    var __VLS_393;
    const __VLS_394 = {}.ElRadioButton;
    /** @type {[typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, typeof __VLS_components.ElRadioButton, typeof __VLS_components.elRadioButton, ]} */ ;
    // @ts-ignore
    const __VLS_395 = __VLS_asFunctionalComponent(__VLS_394, new __VLS_394({
        label: "covered",
    }));
    const __VLS_396 = __VLS_395({
        label: "covered",
    }, ...__VLS_functionalComponentArgsRest(__VLS_395));
    __VLS_397.slots.default;
    (__VLS_ctx.t('plugin.coverage.covered', 'Covered'));
    var __VLS_397;
    var __VLS_385;
    if (__VLS_ctx.coverageStats.missing > 0) {
        const __VLS_398 = {}.ElButton;
        /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
        // @ts-ignore
        const __VLS_399 = __VLS_asFunctionalComponent(__VLS_398, new __VLS_398({
            ...{ 'onClick': {} },
            type: "primary",
            plain: true,
        }));
        const __VLS_400 = __VLS_399({
            ...{ 'onClick': {} },
            type: "primary",
            plain: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_399));
        let __VLS_402;
        let __VLS_403;
        let __VLS_404;
        const __VLS_405 = {
            onClick: (__VLS_ctx.completeAllMissingPermissions)
        };
        __VLS_401.slots.default;
        (__VLS_ctx.t('plugin.coverage.complete_all', 'Complete all'));
        var __VLS_401;
    }
    var __VLS_377;
}
const __VLS_406 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_407 = __VLS_asFunctionalComponent(__VLS_406, new __VLS_406({
    column: (4),
    border: true,
    size: "small",
    ...{ class: "mb-12" },
}));
const __VLS_408 = __VLS_407({
    column: (4),
    border: true,
    size: "small",
    ...{ class: "mb-12" },
}, ...__VLS_functionalComponentArgsRest(__VLS_407));
__VLS_409.slots.default;
const __VLS_410 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_411 = __VLS_asFunctionalComponent(__VLS_410, new __VLS_410({
    label: (__VLS_ctx.t('plugin.menu_total', 'Menu total')),
}));
const __VLS_412 = __VLS_411({
    label: (__VLS_ctx.t('plugin.menu_total', 'Menu total')),
}, ...__VLS_functionalComponentArgsRest(__VLS_411));
__VLS_413.slots.default;
(__VLS_ctx.coverageStats.total);
var __VLS_413;
const __VLS_414 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_415 = __VLS_asFunctionalComponent(__VLS_414, new __VLS_414({
    label: (__VLS_ctx.t('plugin.coverage.covered', 'Covered')),
}));
const __VLS_416 = __VLS_415({
    label: (__VLS_ctx.t('plugin.coverage.covered', 'Covered')),
}, ...__VLS_functionalComponentArgsRest(__VLS_415));
__VLS_417.slots.default;
(__VLS_ctx.coverageStats.covered);
var __VLS_417;
const __VLS_418 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_419 = __VLS_asFunctionalComponent(__VLS_418, new __VLS_418({
    label: (__VLS_ctx.t('plugin.coverage.missing', 'Missing')),
}));
const __VLS_420 = __VLS_419({
    label: (__VLS_ctx.t('plugin.coverage.missing', 'Missing')),
}, ...__VLS_functionalComponentArgsRest(__VLS_419));
__VLS_421.slots.default;
(__VLS_ctx.coverageStats.missing);
var __VLS_421;
const __VLS_422 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_423 = __VLS_asFunctionalComponent(__VLS_422, new __VLS_422({
    label: (__VLS_ctx.t('plugin.coverage.orphan', 'Orphan permissions')),
}));
const __VLS_424 = __VLS_423({
    label: (__VLS_ctx.t('plugin.coverage.orphan', 'Orphan permissions')),
}, ...__VLS_functionalComponentArgsRest(__VLS_423));
__VLS_425.slots.default;
(__VLS_ctx.coverageStats.orphan);
var __VLS_425;
var __VLS_409;
const __VLS_426 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_427 = __VLS_asFunctionalComponent(__VLS_426, new __VLS_426({
    data: (__VLS_ctx.filteredPermissionDiffRows),
    border: true,
    size: "small",
    ...{ class: "mb-16" },
}));
const __VLS_428 = __VLS_427({
    data: (__VLS_ctx.filteredPermissionDiffRows),
    border: true,
    size: "small",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_427));
__VLS_429.slots.default;
const __VLS_430 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_431 = __VLS_asFunctionalComponent(__VLS_430, new __VLS_430({
    prop: "menuName",
    label: (__VLS_ctx.t('plugin.menu', 'Menu')),
    minWidth: "180",
}));
const __VLS_432 = __VLS_431({
    prop: "menuName",
    label: (__VLS_ctx.t('plugin.menu', 'Menu')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_431));
const __VLS_434 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_435 = __VLS_asFunctionalComponent(__VLS_434, new __VLS_434({
    prop: "object",
    label: (__VLS_ctx.t('plugin.permission_object_full', 'Permission object')),
    minWidth: "240",
    showOverflowTooltip: true,
}));
const __VLS_436 = __VLS_435({
    prop: "object",
    label: (__VLS_ctx.t('plugin.permission_object_full', 'Permission object')),
    minWidth: "240",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_435));
const __VLS_438 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_439 = __VLS_asFunctionalComponent(__VLS_438, new __VLS_438({
    label: (__VLS_ctx.t('plugin.existing_actions', 'Existing actions')),
    minWidth: "160",
}));
const __VLS_440 = __VLS_439({
    label: (__VLS_ctx.t('plugin.existing_actions', 'Existing actions')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_439));
__VLS_441.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_441.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_442 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_443 = __VLS_asFunctionalComponent(__VLS_442, new __VLS_442({
        wrap: true,
    }));
    const __VLS_444 = __VLS_443({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_443));
    __VLS_445.slots.default;
    for (const [action] of __VLS_getVForSourceType((row.existingActions))) {
        const __VLS_446 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_447 = __VLS_asFunctionalComponent(__VLS_446, new __VLS_446({
            key: (action),
            effect: "plain",
        }));
        const __VLS_448 = __VLS_447({
            key: (action),
            effect: "plain",
        }, ...__VLS_functionalComponentArgsRest(__VLS_447));
        __VLS_449.slots.default;
        (action);
        var __VLS_449;
    }
    if (row.existingActions.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    }
    var __VLS_445;
}
var __VLS_441;
const __VLS_450 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_451 = __VLS_asFunctionalComponent(__VLS_450, new __VLS_450({
    label: (__VLS_ctx.t('plugin.missing_actions', 'Missing actions')),
    minWidth: "160",
}));
const __VLS_452 = __VLS_451({
    label: (__VLS_ctx.t('plugin.missing_actions', 'Missing actions')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_451));
__VLS_453.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_453.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_454 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_455 = __VLS_asFunctionalComponent(__VLS_454, new __VLS_454({
        wrap: true,
    }));
    const __VLS_456 = __VLS_455({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_455));
    __VLS_457.slots.default;
    for (const [action] of __VLS_getVForSourceType((row.missingActions))) {
        const __VLS_458 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_459 = __VLS_asFunctionalComponent(__VLS_458, new __VLS_458({
            key: (action),
            type: "warning",
            effect: "plain",
        }));
        const __VLS_460 = __VLS_459({
            key: (action),
            type: "warning",
            effect: "plain",
        }, ...__VLS_functionalComponentArgsRest(__VLS_459));
        __VLS_461.slots.default;
        (action);
        var __VLS_461;
    }
    if (row.missingActions.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    }
    var __VLS_457;
}
var __VLS_453;
const __VLS_462 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_463 = __VLS_asFunctionalComponent(__VLS_462, new __VLS_462({
    label: (__VLS_ctx.t('plugin.status', 'Status')),
    width: "120",
}));
const __VLS_464 = __VLS_463({
    label: (__VLS_ctx.t('plugin.status', 'Status')),
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_463));
__VLS_465.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_465.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    if (row.missingActions.length === 0) {
        const __VLS_466 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_467 = __VLS_asFunctionalComponent(__VLS_466, new __VLS_466({
            type: "success",
            effect: "plain",
        }));
        const __VLS_468 = __VLS_467({
            type: "success",
            effect: "plain",
        }, ...__VLS_functionalComponentArgsRest(__VLS_467));
        __VLS_469.slots.default;
        (__VLS_ctx.t('plugin.coverage.covered', 'Covered'));
        var __VLS_469;
    }
    else {
        const __VLS_470 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_471 = __VLS_asFunctionalComponent(__VLS_470, new __VLS_470({
            type: "warning",
            effect: "plain",
        }));
        const __VLS_472 = __VLS_471({
            type: "warning",
            effect: "plain",
        }, ...__VLS_functionalComponentArgsRest(__VLS_471));
        __VLS_473.slots.default;
        (__VLS_ctx.t('plugin.coverage.missing', 'Missing'));
        var __VLS_473;
    }
}
var __VLS_465;
const __VLS_474 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_475 = __VLS_asFunctionalComponent(__VLS_474, new __VLS_474({
    label: (__VLS_ctx.t('plugin.actions', 'Actions')),
    width: "130",
    fixed: "right",
}));
const __VLS_476 = __VLS_475({
    label: (__VLS_ctx.t('plugin.actions', 'Actions')),
    width: "130",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_475));
__VLS_477.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_477.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_478 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_479 = __VLS_asFunctionalComponent(__VLS_478, new __VLS_478({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        disabled: (row.missingActions.length === 0),
    }));
    const __VLS_480 = __VLS_479({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
        disabled: (row.missingActions.length === 0),
    }, ...__VLS_functionalComponentArgsRest(__VLS_479));
    let __VLS_482;
    let __VLS_483;
    let __VLS_484;
    const __VLS_485 = {
        onClick: (...[$event]) => {
            __VLS_ctx.completeDiffRow(row);
        }
    };
    __VLS_481.slots.default;
    (__VLS_ctx.t('plugin.coverage.complete_one', 'Complete one'));
    var __VLS_481;
}
var __VLS_477;
var __VLS_429;
const __VLS_486 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_487 = __VLS_asFunctionalComponent(__VLS_486, new __VLS_486({
    gutter: (16),
}));
const __VLS_488 = __VLS_487({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_487));
__VLS_489.slots.default;
const __VLS_490 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_491 = __VLS_asFunctionalComponent(__VLS_490, new __VLS_490({
    xs: (24),
    md: (12),
}));
const __VLS_492 = __VLS_491({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_491));
__VLS_493.slots.default;
const __VLS_494 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_495 = __VLS_asFunctionalComponent(__VLS_494, new __VLS_494({
    shadow: "never",
}));
const __VLS_496 = __VLS_495({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_495));
__VLS_497.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_497.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.current_menu_preview', 'Current menu preview'));
    const __VLS_498 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_499 = __VLS_asFunctionalComponent(__VLS_498, new __VLS_498({
        effect: "plain",
    }));
    const __VLS_500 = __VLS_499({
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_499));
    __VLS_501.slots.default;
    (__VLS_ctx.t('plugin.row_count', '{count} items', { count: __VLS_ctx.menuPreviewRows.length }));
    var __VLS_501;
}
const __VLS_502 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_503 = __VLS_asFunctionalComponent(__VLS_502, new __VLS_502({
    data: (__VLS_ctx.menuPreviewRows),
    border: true,
    size: "small",
}));
const __VLS_504 = __VLS_503({
    data: (__VLS_ctx.menuPreviewRows),
    border: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_503));
__VLS_505.slots.default;
const __VLS_506 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_507 = __VLS_asFunctionalComponent(__VLS_506, new __VLS_506({
    prop: "sort",
    label: (__VLS_ctx.t('plugin.sort', 'Sort')),
    width: "90",
}));
const __VLS_508 = __VLS_507({
    prop: "sort",
    label: (__VLS_ctx.t('plugin.sort', 'Sort')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_507));
const __VLS_510 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_511 = __VLS_asFunctionalComponent(__VLS_510, new __VLS_510({
    prop: "name",
    label: (__VLS_ctx.t('plugin.menu_name', 'Menu name')),
    minWidth: "180",
}));
const __VLS_512 = __VLS_511({
    prop: "name",
    label: (__VLS_ctx.t('plugin.menu_name', 'Menu name')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_511));
const __VLS_514 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_515 = __VLS_asFunctionalComponent(__VLS_514, new __VLS_514({
    prop: "id",
    label: (__VLS_ctx.t('plugin.menu_id', 'Menu ID')),
    minWidth: "200",
}));
const __VLS_516 = __VLS_515({
    prop: "id",
    label: (__VLS_ctx.t('plugin.menu_id', 'Menu ID')),
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_515));
const __VLS_518 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_519 = __VLS_asFunctionalComponent(__VLS_518, new __VLS_518({
    prop: "type",
    label: (__VLS_ctx.t('plugin.menu_type', 'Type')),
    width: "100",
}));
const __VLS_520 = __VLS_519({
    prop: "type",
    label: (__VLS_ctx.t('plugin.menu_type', 'Type')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_519));
var __VLS_505;
var __VLS_497;
var __VLS_493;
const __VLS_522 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_523 = __VLS_asFunctionalComponent(__VLS_522, new __VLS_522({
    xs: (24),
    md: (12),
}));
const __VLS_524 = __VLS_523({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_523));
__VLS_525.slots.default;
const __VLS_526 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_527 = __VLS_asFunctionalComponent(__VLS_526, new __VLS_526({
    shadow: "never",
}));
const __VLS_528 = __VLS_527({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_527));
__VLS_529.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_529.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.coverage.orphan_title', 'Orphan permissions'));
    const __VLS_530 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_531 = __VLS_asFunctionalComponent(__VLS_530, new __VLS_530({
        effect: "plain",
        type: "danger",
    }));
    const __VLS_532 = __VLS_531({
        effect: "plain",
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_531));
    __VLS_533.slots.default;
    (__VLS_ctx.t('plugin.row_count', '{count} items', { count: __VLS_ctx.orphanPermissions.length }));
    var __VLS_533;
}
if (__VLS_ctx.orphanPermissions.length === 0) {
    const __VLS_534 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_535 = __VLS_asFunctionalComponent(__VLS_534, new __VLS_534({
        description: (__VLS_ctx.t('plugin.coverage.no_orphan', 'No orphan permissions')),
    }));
    const __VLS_536 = __VLS_535({
        description: (__VLS_ctx.t('plugin.coverage.no_orphan', 'No orphan permissions')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_535));
}
else {
    const __VLS_538 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_539 = __VLS_asFunctionalComponent(__VLS_538, new __VLS_538({
        data: (__VLS_ctx.orphanPermissions),
        border: true,
        size: "small",
    }));
    const __VLS_540 = __VLS_539({
        data: (__VLS_ctx.orphanPermissions),
        border: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_539));
    __VLS_541.slots.default;
    const __VLS_542 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_543 = __VLS_asFunctionalComponent(__VLS_542, new __VLS_542({
        prop: "object",
        label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
        minWidth: "220",
        showOverflowTooltip: true,
    }));
    const __VLS_544 = __VLS_543({
        prop: "object",
        label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
        minWidth: "220",
        showOverflowTooltip: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_543));
    const __VLS_546 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_547 = __VLS_asFunctionalComponent(__VLS_546, new __VLS_546({
        prop: "action",
        label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
        width: "120",
    }));
    const __VLS_548 = __VLS_547({
        prop: "action",
        label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
        width: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_547));
    const __VLS_550 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_551 = __VLS_asFunctionalComponent(__VLS_550, new __VLS_550({
        prop: "description",
        label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
        minWidth: "220",
        showOverflowTooltip: true,
    }));
    const __VLS_552 = __VLS_551({
        prop: "description",
        label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
        minWidth: "220",
        showOverflowTooltip: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_551));
    var __VLS_541;
}
var __VLS_529;
var __VLS_525;
var __VLS_489;
var __VLS_373;
const __VLS_554 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_555 = __VLS_asFunctionalComponent(__VLS_554, new __VLS_554({
    shadow: "never",
}));
const __VLS_556 = __VLS_555({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_555));
__VLS_557.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_557.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('plugin.permission_detail_title', 'Permission details'));
    const __VLS_558 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_559 = __VLS_asFunctionalComponent(__VLS_558, new __VLS_558({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
    }));
    const __VLS_560 = __VLS_559({
        ...{ 'onClick': {} },
        type: "primary",
        plain: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_559));
    let __VLS_562;
    let __VLS_563;
    let __VLS_564;
    const __VLS_565 = {
        onClick: (__VLS_ctx.appendPermissionRow)
    };
    __VLS_561.slots.default;
    (__VLS_ctx.t('plugin.add_permission_row', 'Add permission row'));
    var __VLS_561;
}
const __VLS_566 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_567 = __VLS_asFunctionalComponent(__VLS_566, new __VLS_566({
    data: (__VLS_ctx.form.permissions),
    border: true,
    rowKey: "object",
    size: "small",
}));
const __VLS_568 = __VLS_567({
    data: (__VLS_ctx.form.permissions),
    border: true,
    rowKey: "object",
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_567));
__VLS_569.slots.default;
const __VLS_570 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_571 = __VLS_asFunctionalComponent(__VLS_570, new __VLS_570({
    label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
    minWidth: "220",
}));
const __VLS_572 = __VLS_571({
    label: (__VLS_ctx.t('plugin.permission_object', 'Object')),
    minWidth: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_571));
__VLS_573.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_573.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_574 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_575 = __VLS_asFunctionalComponent(__VLS_574, new __VLS_574({
        modelValue: (row.object),
        placeholder: (__VLS_ctx.t('plugin.permission_object_placeholder_detail', 'plugin:example:menu-home')),
    }));
    const __VLS_576 = __VLS_575({
        modelValue: (row.object),
        placeholder: (__VLS_ctx.t('plugin.permission_object_placeholder_detail', 'plugin:example:menu-home')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_575));
}
var __VLS_573;
const __VLS_578 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_579 = __VLS_asFunctionalComponent(__VLS_578, new __VLS_578({
    label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
    minWidth: "140",
}));
const __VLS_580 = __VLS_579({
    label: (__VLS_ctx.t('plugin.permission_action', 'Action')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_579));
__VLS_581.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_581.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_582 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_583 = __VLS_asFunctionalComponent(__VLS_582, new __VLS_582({
        modelValue: (row.action),
        placeholder: (__VLS_ctx.t('plugin.permission_action_placeholder_detail', 'view')),
    }));
    const __VLS_584 = __VLS_583({
        modelValue: (row.action),
        placeholder: (__VLS_ctx.t('plugin.permission_action_placeholder_detail', 'view')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_583));
}
var __VLS_581;
const __VLS_586 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_587 = __VLS_asFunctionalComponent(__VLS_586, new __VLS_586({
    label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
    minWidth: "260",
}));
const __VLS_588 = __VLS_587({
    label: (__VLS_ctx.t('plugin.permission_description', 'Description')),
    minWidth: "260",
}, ...__VLS_functionalComponentArgsRest(__VLS_587));
__VLS_589.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_589.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_590 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_591 = __VLS_asFunctionalComponent(__VLS_590, new __VLS_590({
        modelValue: (row.description),
        placeholder: (__VLS_ctx.t('plugin.permission_description_placeholder_detail', 'Permission description')),
    }));
    const __VLS_592 = __VLS_591({
        modelValue: (row.description),
        placeholder: (__VLS_ctx.t('plugin.permission_description_placeholder_detail', 'Permission description')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_591));
}
var __VLS_589;
const __VLS_594 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_595 = __VLS_asFunctionalComponent(__VLS_594, new __VLS_594({
    label: (__VLS_ctx.t('plugin.actions', 'Actions')),
    width: "90",
    fixed: "right",
}));
const __VLS_596 = __VLS_595({
    label: (__VLS_ctx.t('plugin.actions', 'Actions')),
    width: "90",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_595));
__VLS_597.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_597.slots;
    const [{ $index }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_598 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_599 = __VLS_asFunctionalComponent(__VLS_598, new __VLS_598({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_600 = __VLS_599({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_599));
    let __VLS_602;
    let __VLS_603;
    let __VLS_604;
    const __VLS_605 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removePermissionRow($index);
        }
    };
    __VLS_601.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_601;
}
var __VLS_597;
var __VLS_569;
var __VLS_557;
var __VLS_161;
var __VLS_70;
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form--two-col']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form__full-row']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__preset-header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__preset-search']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__preset-groups']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__group-title']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__preset-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__preset-meta']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__coverage-visual']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-detail-page__coverage-metrics']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AdminTable: AdminTable,
            PluginMenuTreeEditor: PluginMenuTreeEditor,
            formatDateTime: formatDateTime,
            t: t,
            loading: loading,
            saving: saving,
            activeTab: activeTab,
            selectedActions: selectedActions,
            selectedTemplateKey: selectedTemplateKey,
            presetName: presetName,
            presetSearchQuery: presetSearchQuery,
            diffFilter: diffFilter,
            presets: presets,
            sortNotice: sortNotice,
            plugin: plugin,
            permissionTemplateOptions: permissionTemplateOptions,
            form: form,
            pluginName: pluginName,
            pageTitle: pageTitle,
            menuCount: menuCount,
            permissionCount: permissionCount,
            generatedTemplatePermissions: generatedTemplatePermissions,
            menuPreviewRows: menuPreviewRows,
            selectedTemplate: selectedTemplate,
            orphanPermissions: orphanPermissions,
            filteredGroupedPresets: filteredGroupedPresets,
            filteredPermissionDiffRows: filteredPermissionDiffRows,
            coverageStats: coverageStats,
            coverageLevel: coverageLevel,
            coverageProgressColor: coverageProgressColor,
            coverageLevelLabel: coverageLevelLabel,
            loadPlugin: loadPlugin,
            appendPermissionRow: appendPermissionRow,
            removePermissionRow: removePermissionRow,
            completeDiffRow: completeDiffRow,
            completeAllMissingPermissions: completeAllMissingPermissions,
            refreshPresets: refreshPresets,
            saveCurrentPreset: saveCurrentPreset,
            applyPreset: applyPreset,
            deletePreset: deletePreset,
            applyPermissionTemplate: applyPermissionTemplate,
            handleMoveNode: handleMoveNode,
            savePlugin: savePlugin,
            goBack: goBack,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
