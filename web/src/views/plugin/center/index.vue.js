import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createPlugin, deletePlugin, fetchPlugins, updatePlugin } from '@/api/plugins';
import { formatDateTime, statusTagType } from '@/utils/admin';
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref([]);
const activeTab = ref('basic');
const editingName = ref('');
const router = useRouter();
const { t } = useAppI18n();
const query = reactive({
    keyword: '',
    enabled: '',
});
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
const filteredRows = computed(() => {
    const keyword = query.keyword.trim().toLowerCase();
    return rows.value.filter((row) => {
        const matchesKeyword = keyword === '' ||
            [row.name, row.description ?? ''].some((value) => value.toLowerCase().includes(keyword));
        const matchesEnabled = query.enabled === '' || String(row.enabled) === query.enabled;
        return matchesKeyword && matchesEnabled;
    });
});
const pluginCount = computed(() => filteredRows.value.length);
const enabledCount = computed(() => filteredRows.value.filter((item) => item.enabled).length);
function resetForm() {
    Object.assign(form, defaultForm());
}
function createMenuRow() {
    const pluginName = editingName.value.trim() || form.name.trim();
    return {
        plugin: pluginName,
        id: `menu-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        parent_id: '',
        name: '',
        titleKey: '',
        titleDefault: '',
        path: '',
        component: '',
        icon: '',
        sort: 0,
        permission: '',
        type: 'menu',
        visible: true,
        enabled: true,
        redirect: '',
        external_url: '',
        children: [],
    };
}
function createPermissionRow() {
    const pluginName = editingName.value.trim() || form.name.trim();
    return {
        plugin: pluginName,
        object: '',
        action: '',
        description: '',
    };
}
function parentMenuOptions(currentIndex) {
    return form.menus
        .filter((_, index) => index !== currentIndex)
        .map((menu) => ({
        label: `${t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id)} (${menu.id})`,
        value: menu.id,
    }));
}
function normalizeMenuRow(row) {
    const pluginName = editingName.value.trim() || form.name.trim();
    return {
        ...row,
        plugin: pluginName,
        id: row.id.trim(),
        parent_id: row.parent_id?.trim() ?? '',
        name: row.name.trim(),
        titleKey: row.titleKey?.trim() ?? '',
        titleDefault: row.titleDefault?.trim() ?? '',
        path: row.path.trim(),
        component: row.component?.trim() ?? '',
        icon: row.icon?.trim() ?? '',
        sort: Number(row.sort) || 0,
        permission: row.permission?.trim() ?? '',
        type: row.type?.trim() || 'menu',
        visible: Boolean(row.visible),
        enabled: Boolean(row.enabled),
        redirect: row.redirect?.trim() ?? '',
        external_url: row.external_url?.trim() ?? '',
        children: [],
    };
}
function normalizePermissionRow(row) {
    const pluginName = editingName.value.trim() || form.name.trim();
    return {
        plugin: pluginName,
        object: row.object.trim(),
        action: row.action.trim(),
        description: row.description.trim(),
    };
}
async function loadPlugins() {
    tableLoading.value = true;
    try {
        const response = await fetchPlugins();
        rows.value = response.items ?? [];
    }
    finally {
        tableLoading.value = false;
    }
}
function openDetail(row) {
    void router.push(`/system/plugins/${encodeURIComponent(row.name)}`);
}
function openCreate() {
    editingName.value = '';
    resetForm();
    activeTab.value = 'basic';
    dialogVisible.value = true;
}
function openEdit(row) {
    editingName.value = row.name;
    Object.assign(form, defaultForm(), {
        name: row.name,
        description: row.description ?? '',
        enabled: row.enabled,
        menus: (row.menus ?? []).map((menu) => ({ ...menu, children: [] })),
        permissions: (row.permissions ?? []).map((permission) => ({ ...permission })),
    });
    if (form.menus.length === 0) {
        form.menus.push(createMenuRow());
    }
    if (form.permissions.length === 0) {
        form.permissions.push(createPermissionRow());
    }
    activeTab.value = 'basic';
    dialogVisible.value = true;
}
function handleSearch() {
    void loadPlugins();
}
function handleReset() {
    query.keyword = '';
    query.enabled = '';
}
function appendMenuRow() {
    form.menus.push(createMenuRow());
}
function removeMenuRow(index) {
    form.menus.splice(index, 1);
}
function appendPermissionRow() {
    form.permissions.push(createPermissionRow());
}
function removePermissionRow(index) {
    form.permissions.splice(index, 1);
}
async function submitForm() {
    const name = form.name.trim();
    if (name === '') {
        ElMessage.warning(t('plugin.validation_name', '请输入插件名称'));
        return;
    }
    const menus = form.menus
        .map((item) => normalizeMenuRow(item))
        .filter((item) => item.name !== '' || item.path !== '' || item.component !== '' || item.permission !== '');
    if (menus.some((item) => item.name === '' || item.path === '')) {
        ElMessage.warning(t('plugin.validation_menu', '请补全插件菜单名称和路径'));
        return;
    }
    const permissions = form.permissions
        .map((item) => normalizePermissionRow(item))
        .filter((item) => item.object !== '' || item.action !== '' || item.description !== '');
    if (permissions.some((item) => item.object === '' || item.action === '')) {
        ElMessage.warning(t('plugin.validation_permission', '请补全插件权限的对象和动作'));
        return;
    }
    const payload = {
        name,
        description: form.description.trim(),
        enabled: Boolean(form.enabled),
        menus,
        permissions,
    };
    dialogLoading.value = true;
    try {
        if (editingName.value) {
            await updatePlugin(editingName.value, payload);
            ElMessage.success(t('plugin.updated', '插件已更新'));
        }
        else {
            await createPlugin(payload);
            ElMessage.success(t('plugin.created', '插件已创建'));
        }
        dialogVisible.value = false;
        await loadPlugins();
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('plugin.confirm_delete', '确认删除插件 {name} 吗？', { name: row.name }), t('plugin.delete_title', '删除插件'), {
        type: 'warning',
        confirmButtonText: t('plugin.delete_confirm', '删除'),
        cancelButtonText: t('plugin.delete_cancel', '取消'),
    });
    await deletePlugin(row.name);
    ElMessage.success(t('plugin.deleted', '插件已删除'));
    await loadPlugins();
}
function statusLabel(enabled) {
    return enabled ? t('menu.status.active', '启用') : t('menu.status.inactive', '禁用');
}
function menuTypeLabel(type) {
    switch (type) {
        case 'directory':
            return t('menu.type.directory', '目录');
        case 'button':
            return t('menu.type.button', '按钮');
        default:
            return t('menu.type.menu', '菜单');
    }
}
onMounted(() => {
    void loadPlugins();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-page" },
});
/** @type {[typeof AdminTable, typeof AdminTable, ]} */ ;
// @ts-ignore
const __VLS_0 = __VLS_asFunctionalComponent(AdminTable, new AdminTable({
    title: (__VLS_ctx.t('plugin.title', '插件中心')),
    description: (__VLS_ctx.t('plugin.description', '管理插件基础信息、插件菜单和插件权限定义。')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('plugin.title', '插件中心')),
    description: (__VLS_ctx.t('plugin.description', '管理插件基础信息、插件菜单和插件权限定义。')),
    loading: (__VLS_ctx.tableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_0));
__VLS_2.slots.default;
{
    const { actions: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_3 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_4 = __VLS_asFunctionalComponent(__VLS_3, new __VLS_3({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }));
    const __VLS_5 = __VLS_4({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_4));
    let __VLS_7;
    let __VLS_8;
    let __VLS_9;
    const __VLS_10 = {
        onClick: (__VLS_ctx.loadPlugins)
    };
    __VLS_6.slots.default;
    (__VLS_ctx.t('plugin.refresh', '刷新'));
    var __VLS_6;
    const __VLS_11 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_13 = __VLS_12({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    let __VLS_15;
    let __VLS_16;
    let __VLS_17;
    const __VLS_18 = {
        onClick: (__VLS_ctx.openCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('plugin:create') }, null, null);
    __VLS_14.slots.default;
    (__VLS_ctx.t('common.create', '新增'));
    var __VLS_14;
}
{
    const { filters: __VLS_thisSlot } = __VLS_2.slots;
    const __VLS_19 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }));
    const __VLS_21 = __VLS_20({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    __VLS_22.slots.default;
    const __VLS_23 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_24 = __VLS_asFunctionalComponent(__VLS_23, new __VLS_23({
        label: (__VLS_ctx.t('common.search', '查询')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('common.search', '查询')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('plugin.keyword_placeholder', '插件名称 / 描述')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('plugin.keyword_placeholder', '插件名称 / 描述')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: (__VLS_ctx.t('plugin.status', '状态')),
    }));
    const __VLS_33 = __VLS_32({
        label: (__VLS_ctx.t('plugin.status', '状态')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.enabled),
        clearable: true,
        placeholder: (__VLS_ctx.t('plugin.all_status', '全部状态')),
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.enabled),
        clearable: true,
        placeholder: (__VLS_ctx.t('plugin.all_status', '全部状态')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    const __VLS_39 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: (__VLS_ctx.t('menu.status.active', '启用')),
        value: "true",
    }));
    const __VLS_41 = __VLS_40({
        label: (__VLS_ctx.t('menu.status.active', '启用')),
        value: "true",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    const __VLS_43 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: (__VLS_ctx.t('menu.status.inactive', '禁用')),
        value: "false",
    }));
    const __VLS_45 = __VLS_44({
        label: (__VLS_ctx.t('menu.status.inactive', '禁用')),
        value: "false",
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    var __VLS_38;
    var __VLS_34;
    const __VLS_47 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({}));
    const __VLS_49 = __VLS_48({}, ...__VLS_functionalComponentArgsRest(__VLS_48));
    __VLS_50.slots.default;
    const __VLS_51 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_53 = __VLS_52({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    let __VLS_55;
    let __VLS_56;
    let __VLS_57;
    const __VLS_58 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_54.slots.default;
    (__VLS_ctx.t('common.search', '查询'));
    var __VLS_54;
    const __VLS_59 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        ...{ 'onClick': {} },
    }));
    const __VLS_61 = __VLS_60({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    let __VLS_63;
    let __VLS_64;
    let __VLS_65;
    const __VLS_66 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_62.slots.default;
    (__VLS_ctx.t('common.reset', '重置'));
    var __VLS_62;
    var __VLS_50;
    var __VLS_22;
}
const __VLS_67 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    title: (__VLS_ctx.t('plugin.info_title', '插件中心说明')),
    description: (__VLS_ctx.t('plugin.info_description', '此页面用于管理插件定义、插件菜单和插件权限。插件页面本身由菜单中的 view/plugin/center/index 动态加载。')),
    type: "info",
    showIcon: true,
    closable: (false),
    ...{ class: "mb-16" },
}));
const __VLS_69 = __VLS_68({
    title: (__VLS_ctx.t('plugin.info_title', '插件中心说明')),
    description: (__VLS_ctx.t('plugin.info_description', '此页面用于管理插件定义、插件菜单和插件权限。插件页面本身由菜单中的 view/plugin/center/index 动态加载。')),
    type: "info",
    showIcon: true,
    closable: (false),
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
const __VLS_71 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    gutter: (16),
    ...{ class: "mb-16" },
}));
const __VLS_73 = __VLS_72({
    gutter: (16),
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
const __VLS_75 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    xs: (24),
    md: (12),
}));
const __VLS_77 = __VLS_76({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
const __VLS_79 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    shadow: "never",
}));
const __VLS_81 = __VLS_80({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
const __VLS_83 = {}.ElStatistic;
/** @type {[typeof __VLS_components.ElStatistic, typeof __VLS_components.elStatistic, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    title: (__VLS_ctx.t('plugin.total_label', '插件总数')),
    value: (__VLS_ctx.pluginCount),
}));
const __VLS_85 = __VLS_84({
    title: (__VLS_ctx.t('plugin.total_label', '插件总数')),
    value: (__VLS_ctx.pluginCount),
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
var __VLS_82;
var __VLS_78;
const __VLS_87 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    xs: (24),
    md: (12),
}));
const __VLS_89 = __VLS_88({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
const __VLS_91 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    shadow: "never",
}));
const __VLS_93 = __VLS_92({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
const __VLS_95 = {}.ElStatistic;
/** @type {[typeof __VLS_components.ElStatistic, typeof __VLS_components.elStatistic, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    title: (__VLS_ctx.t('plugin.enabled_label', '启用插件')),
    value: (__VLS_ctx.enabledCount),
}));
const __VLS_97 = __VLS_96({
    title: (__VLS_ctx.t('plugin.enabled_label', '启用插件')),
    value: (__VLS_ctx.enabledCount),
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
var __VLS_94;
var __VLS_90;
var __VLS_74;
const __VLS_99 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    data: (__VLS_ctx.filteredRows),
    border: true,
    rowKey: "name",
}));
const __VLS_101 = __VLS_100({
    data: (__VLS_ctx.filteredRows),
    border: true,
    rowKey: "name",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_102.slots.default;
const __VLS_103 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    prop: "name",
    label: (__VLS_ctx.t('plugin.name', '插件名称')),
    minWidth: "160",
}));
const __VLS_105 = __VLS_104({
    prop: "name",
    label: (__VLS_ctx.t('plugin.name', '插件名称')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
const __VLS_107 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    prop: "description",
    label: (__VLS_ctx.t('plugin.description_column', '描述')),
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_109 = __VLS_108({
    prop: "description",
    label: (__VLS_ctx.t('plugin.description_column', '描述')),
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
const __VLS_111 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('plugin.status', '状态')),
    width: "110",
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('plugin.status', '状态')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
__VLS_114.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_114.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_115 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
        type: (__VLS_ctx.statusTagType(row.enabled ? 'active' : 'inactive')),
        effect: "plain",
    }));
    const __VLS_117 = __VLS_116({
        type: (__VLS_ctx.statusTagType(row.enabled ? 'active' : 'inactive')),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_116));
    __VLS_118.slots.default;
    (__VLS_ctx.statusLabel(row.enabled));
    var __VLS_118;
}
var __VLS_114;
const __VLS_119 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
    label: (__VLS_ctx.t('plugin.menus', '菜单数')),
    width: "100",
}));
const __VLS_121 = __VLS_120({
    label: (__VLS_ctx.t('plugin.menus', '菜单数')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_120));
__VLS_122.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_122.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.menus?.length ?? 0);
}
var __VLS_122;
const __VLS_123 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
    label: (__VLS_ctx.t('plugin.permissions', '权限数')),
    width: "100",
}));
const __VLS_125 = __VLS_124({
    label: (__VLS_ctx.t('plugin.permissions', '权限数')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_124));
__VLS_126.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_126.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.permissions?.length ?? 0);
}
var __VLS_126;
const __VLS_127 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
    label: (__VLS_ctx.t('plugin.created_at', '创建时间')),
    minWidth: "180",
}));
const __VLS_129 = __VLS_128({
    label: (__VLS_ctx.t('plugin.created_at', '创建时间')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_128));
__VLS_130.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_130.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_130;
const __VLS_131 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "180",
    fixed: "right",
}));
const __VLS_133 = __VLS_132({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_132));
__VLS_134.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_134.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_135 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
        ...{ 'onClick': {} },
        link: true,
        type: "success",
    }));
    const __VLS_137 = __VLS_136({
        ...{ 'onClick': {} },
        link: true,
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_136));
    let __VLS_139;
    let __VLS_140;
    let __VLS_141;
    const __VLS_142 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openDetail(row);
        }
    };
    __VLS_138.slots.default;
    (__VLS_ctx.t('plugin.detail', '详情'));
    var __VLS_138;
    const __VLS_143 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_145 = __VLS_144({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_144));
    let __VLS_147;
    let __VLS_148;
    let __VLS_149;
    const __VLS_150 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('plugin:update') }, null, null);
    __VLS_146.slots.default;
    (__VLS_ctx.t('common.edit', '编辑'));
    var __VLS_146;
    const __VLS_151 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_153 = __VLS_152({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_152));
    let __VLS_155;
    let __VLS_156;
    let __VLS_157;
    const __VLS_158 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('plugin:delete') }, null, null);
    __VLS_154.slots.default;
    (__VLS_ctx.t('common.delete', '删除'));
    var __VLS_154;
}
var __VLS_134;
var __VLS_102;
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_159 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingName ? __VLS_ctx.t('plugin.edit_title', '编辑插件') : __VLS_ctx.t('plugin.create_title', '新增插件')),
    loading: (__VLS_ctx.dialogLoading),
    width: "1180px",
}));
const __VLS_160 = __VLS_159({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingName ? __VLS_ctx.t('plugin.edit_title', '编辑插件') : __VLS_ctx.t('plugin.create_title', '新增插件')),
    loading: (__VLS_ctx.dialogLoading),
    width: "1180px",
}, ...__VLS_functionalComponentArgsRest(__VLS_159));
let __VLS_162;
let __VLS_163;
let __VLS_164;
const __VLS_165 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_161.slots.default;
const __VLS_166 = {}.ElTabs;
/** @type {[typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, typeof __VLS_components.ElTabs, typeof __VLS_components.elTabs, ]} */ ;
// @ts-ignore
const __VLS_167 = __VLS_asFunctionalComponent(__VLS_166, new __VLS_166({
    modelValue: (__VLS_ctx.activeTab),
    ...{ class: "plugin-center-tabs" },
}));
const __VLS_168 = __VLS_167({
    modelValue: (__VLS_ctx.activeTab),
    ...{ class: "plugin-center-tabs" },
}, ...__VLS_functionalComponentArgsRest(__VLS_167));
__VLS_169.slots.default;
const __VLS_170 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_171 = __VLS_asFunctionalComponent(__VLS_170, new __VLS_170({
    label: (__VLS_ctx.t('plugin.basic_tab', '基础信息')),
    name: "basic",
}));
const __VLS_172 = __VLS_171({
    label: (__VLS_ctx.t('plugin.basic_tab', '基础信息')),
    name: "basic",
}, ...__VLS_functionalComponentArgsRest(__VLS_171));
__VLS_173.slots.default;
const __VLS_174 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_175 = __VLS_asFunctionalComponent(__VLS_174, new __VLS_174({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_176 = __VLS_175({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_175));
__VLS_177.slots.default;
const __VLS_178 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_179 = __VLS_asFunctionalComponent(__VLS_178, new __VLS_178({
    label: (__VLS_ctx.t('plugin.name', '插件名称')),
    required: true,
}));
const __VLS_180 = __VLS_179({
    label: (__VLS_ctx.t('plugin.name', '插件名称')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_179));
__VLS_181.slots.default;
const __VLS_182 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_183 = __VLS_asFunctionalComponent(__VLS_182, new __VLS_182({
    modelValue: (__VLS_ctx.form.name),
    disabled: (Boolean(__VLS_ctx.editingName)),
    placeholder: (__VLS_ctx.t('plugin.validation_name', '请输入插件名称')),
}));
const __VLS_184 = __VLS_183({
    modelValue: (__VLS_ctx.form.name),
    disabled: (Boolean(__VLS_ctx.editingName)),
    placeholder: (__VLS_ctx.t('plugin.validation_name', '请输入插件名称')),
}, ...__VLS_functionalComponentArgsRest(__VLS_183));
var __VLS_181;
const __VLS_186 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_187 = __VLS_asFunctionalComponent(__VLS_186, new __VLS_186({
    label: (__VLS_ctx.t('plugin.description_label', '插件描述')),
}));
const __VLS_188 = __VLS_187({
    label: (__VLS_ctx.t('plugin.description_label', '插件描述')),
}, ...__VLS_functionalComponentArgsRest(__VLS_187));
__VLS_189.slots.default;
const __VLS_190 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_191 = __VLS_asFunctionalComponent(__VLS_190, new __VLS_190({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('plugin.description_placeholder', '请输入插件描述')),
}));
const __VLS_192 = __VLS_191({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('plugin.description_placeholder', '请输入插件描述')),
}, ...__VLS_functionalComponentArgsRest(__VLS_191));
var __VLS_189;
const __VLS_194 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_195 = __VLS_asFunctionalComponent(__VLS_194, new __VLS_194({
    label: (__VLS_ctx.t('plugin.enabled_status', '启用状态')),
}));
const __VLS_196 = __VLS_195({
    label: (__VLS_ctx.t('plugin.enabled_status', '启用状态')),
}, ...__VLS_functionalComponentArgsRest(__VLS_195));
__VLS_197.slots.default;
const __VLS_198 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_199 = __VLS_asFunctionalComponent(__VLS_198, new __VLS_198({
    modelValue: (__VLS_ctx.form.enabled),
}));
const __VLS_200 = __VLS_199({
    modelValue: (__VLS_ctx.form.enabled),
}, ...__VLS_functionalComponentArgsRest(__VLS_199));
var __VLS_197;
var __VLS_177;
var __VLS_173;
const __VLS_202 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_203 = __VLS_asFunctionalComponent(__VLS_202, new __VLS_202({
    label: (__VLS_ctx.t('plugin.menus_tab', '插件菜单')),
    name: "menus",
}));
const __VLS_204 = __VLS_203({
    label: (__VLS_ctx.t('plugin.menus_tab', '插件菜单')),
    name: "menus",
}, ...__VLS_functionalComponentArgsRest(__VLS_203));
__VLS_205.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-table__actions mb-12" },
});
const __VLS_206 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_207 = __VLS_asFunctionalComponent(__VLS_206, new __VLS_206({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_208 = __VLS_207({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_207));
let __VLS_210;
let __VLS_211;
let __VLS_212;
const __VLS_213 = {
    onClick: (__VLS_ctx.appendMenuRow)
};
__VLS_209.slots.default;
(__VLS_ctx.t('plugin.add_menu_row', '新增菜单行'));
var __VLS_209;
const __VLS_214 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_215 = __VLS_asFunctionalComponent(__VLS_214, new __VLS_214({
    data: (__VLS_ctx.form.menus),
    border: true,
    rowKey: "id",
    size: "small",
}));
const __VLS_216 = __VLS_215({
    data: (__VLS_ctx.form.menus),
    border: true,
    rowKey: "id",
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_215));
__VLS_217.slots.default;
const __VLS_218 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_219 = __VLS_asFunctionalComponent(__VLS_218, new __VLS_218({
    label: (__VLS_ctx.t('plugin.menu_name', '名称')),
    minWidth: "150",
}));
const __VLS_220 = __VLS_219({
    label: (__VLS_ctx.t('plugin.menu_name', '名称')),
    minWidth: "150",
}, ...__VLS_functionalComponentArgsRest(__VLS_219));
__VLS_221.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_221.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_222 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_223 = __VLS_asFunctionalComponent(__VLS_222, new __VLS_222({
        modelValue: (row.name),
        placeholder: (__VLS_ctx.t('plugin.menu_name_placeholder', '菜单名称')),
    }));
    const __VLS_224 = __VLS_223({
        modelValue: (row.name),
        placeholder: (__VLS_ctx.t('plugin.menu_name_placeholder', '菜单名称')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_223));
}
var __VLS_221;
const __VLS_226 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_227 = __VLS_asFunctionalComponent(__VLS_226, new __VLS_226({
    label: (__VLS_ctx.t('plugin.menu_title_key', '标题 Key')),
    minWidth: "170",
}));
const __VLS_228 = __VLS_227({
    label: (__VLS_ctx.t('plugin.menu_title_key', '标题 Key')),
    minWidth: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_227));
__VLS_229.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_229.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_230 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_231 = __VLS_asFunctionalComponent(__VLS_230, new __VLS_230({
        modelValue: (row.titleKey),
        placeholder: (__VLS_ctx.t('plugin.menu_title_key_placeholder', '例如 route.dashboard')),
    }));
    const __VLS_232 = __VLS_231({
        modelValue: (row.titleKey),
        placeholder: (__VLS_ctx.t('plugin.menu_title_key_placeholder', '例如 route.dashboard')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_231));
}
var __VLS_229;
const __VLS_234 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_235 = __VLS_asFunctionalComponent(__VLS_234, new __VLS_234({
    label: (__VLS_ctx.t('plugin.menu_title_default', '标题默认值')),
    minWidth: "170",
}));
const __VLS_236 = __VLS_235({
    label: (__VLS_ctx.t('plugin.menu_title_default', '标题默认值')),
    minWidth: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_235));
__VLS_237.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_237.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_238 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_239 = __VLS_asFunctionalComponent(__VLS_238, new __VLS_238({
        modelValue: (row.titleDefault),
        placeholder: (__VLS_ctx.t('plugin.menu_title_default_placeholder', '例如 仪表盘')),
    }));
    const __VLS_240 = __VLS_239({
        modelValue: (row.titleDefault),
        placeholder: (__VLS_ctx.t('plugin.menu_title_default_placeholder', '例如 仪表盘')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_239));
}
var __VLS_237;
const __VLS_242 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_243 = __VLS_asFunctionalComponent(__VLS_242, new __VLS_242({
    label: (__VLS_ctx.t('plugin.menu_path', '路径')),
    minWidth: "180",
}));
const __VLS_244 = __VLS_243({
    label: (__VLS_ctx.t('plugin.menu_path', '路径')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_243));
__VLS_245.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_245.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_246 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_247 = __VLS_asFunctionalComponent(__VLS_246, new __VLS_246({
        modelValue: (row.path),
        placeholder: (__VLS_ctx.t('plugin.menu_path_placeholder', '/plugin/xxx')),
    }));
    const __VLS_248 = __VLS_247({
        modelValue: (row.path),
        placeholder: (__VLS_ctx.t('plugin.menu_path_placeholder', '/plugin/xxx')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_247));
}
var __VLS_245;
const __VLS_250 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_251 = __VLS_asFunctionalComponent(__VLS_250, new __VLS_250({
    label: (__VLS_ctx.t('plugin.menu_component', '组件')),
    minWidth: "170",
}));
const __VLS_252 = __VLS_251({
    label: (__VLS_ctx.t('plugin.menu_component', '组件')),
    minWidth: "170",
}, ...__VLS_functionalComponentArgsRest(__VLS_251));
__VLS_253.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_253.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_254 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_255 = __VLS_asFunctionalComponent(__VLS_254, new __VLS_254({
        modelValue: (row.component),
        placeholder: (__VLS_ctx.t('plugin.menu_component_placeholder', 'view/plugin/example/index')),
    }));
    const __VLS_256 = __VLS_255({
        modelValue: (row.component),
        placeholder: (__VLS_ctx.t('plugin.menu_component_placeholder', 'view/plugin/example/index')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_255));
}
var __VLS_253;
const __VLS_258 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_259 = __VLS_asFunctionalComponent(__VLS_258, new __VLS_258({
    label: (__VLS_ctx.t('plugin.menu_parent', '父级')),
    minWidth: "160",
}));
const __VLS_260 = __VLS_259({
    label: (__VLS_ctx.t('plugin.menu_parent', '父级')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_259));
__VLS_261.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_261.slots;
    const [{ row, $index }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_262 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_263 = __VLS_asFunctionalComponent(__VLS_262, new __VLS_262({
        modelValue: (row.parent_id),
        clearable: true,
        filterable: true,
        placeholder: (__VLS_ctx.t('plugin.menu_parent_placeholder', '无父级')),
    }));
    const __VLS_264 = __VLS_263({
        modelValue: (row.parent_id),
        clearable: true,
        filterable: true,
        placeholder: (__VLS_ctx.t('plugin.menu_parent_placeholder', '无父级')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_263));
    __VLS_265.slots.default;
    const __VLS_266 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_267 = __VLS_asFunctionalComponent(__VLS_266, new __VLS_266({
        label: (__VLS_ctx.t('plugin.menu_parent_placeholder', '无父级')),
        value: "",
    }));
    const __VLS_268 = __VLS_267({
        label: (__VLS_ctx.t('plugin.menu_parent_placeholder', '无父级')),
        value: "",
    }, ...__VLS_functionalComponentArgsRest(__VLS_267));
    for (const [option] of __VLS_getVForSourceType((__VLS_ctx.parentMenuOptions($index)))) {
        const __VLS_270 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_271 = __VLS_asFunctionalComponent(__VLS_270, new __VLS_270({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }));
        const __VLS_272 = __VLS_271({
            key: (option.value),
            label: (option.label),
            value: (option.value),
        }, ...__VLS_functionalComponentArgsRest(__VLS_271));
    }
    var __VLS_265;
}
var __VLS_261;
const __VLS_274 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_275 = __VLS_asFunctionalComponent(__VLS_274, new __VLS_274({
    label: (__VLS_ctx.t('plugin.menu_type', '类型')),
    width: "110",
}));
const __VLS_276 = __VLS_275({
    label: (__VLS_ctx.t('plugin.menu_type', '类型')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_275));
__VLS_277.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_277.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_278 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_279 = __VLS_asFunctionalComponent(__VLS_278, new __VLS_278({
        modelValue: (row.type),
        ...{ style: {} },
    }));
    const __VLS_280 = __VLS_279({
        modelValue: (row.type),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_279));
    __VLS_281.slots.default;
    const __VLS_282 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_283 = __VLS_asFunctionalComponent(__VLS_282, new __VLS_282({
        label: (__VLS_ctx.t('plugin.menu_type_directory', '目录')),
        value: "directory",
    }));
    const __VLS_284 = __VLS_283({
        label: (__VLS_ctx.t('plugin.menu_type_directory', '目录')),
        value: "directory",
    }, ...__VLS_functionalComponentArgsRest(__VLS_283));
    const __VLS_286 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_287 = __VLS_asFunctionalComponent(__VLS_286, new __VLS_286({
        label: (__VLS_ctx.t('plugin.menu_type_menu', '菜单')),
        value: "menu",
    }));
    const __VLS_288 = __VLS_287({
        label: (__VLS_ctx.t('plugin.menu_type_menu', '菜单')),
        value: "menu",
    }, ...__VLS_functionalComponentArgsRest(__VLS_287));
    const __VLS_290 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_291 = __VLS_asFunctionalComponent(__VLS_290, new __VLS_290({
        label: (__VLS_ctx.t('plugin.menu_type_button', '按钮')),
        value: "button",
    }));
    const __VLS_292 = __VLS_291({
        label: (__VLS_ctx.t('plugin.menu_type_button', '按钮')),
        value: "button",
    }, ...__VLS_functionalComponentArgsRest(__VLS_291));
    var __VLS_281;
}
var __VLS_277;
const __VLS_294 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_295 = __VLS_asFunctionalComponent(__VLS_294, new __VLS_294({
    label: (__VLS_ctx.t('plugin.menu_permission', '权限')),
    minWidth: "140",
}));
const __VLS_296 = __VLS_295({
    label: (__VLS_ctx.t('plugin.menu_permission', '权限')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_295));
__VLS_297.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_297.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_298 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_299 = __VLS_asFunctionalComponent(__VLS_298, new __VLS_298({
        modelValue: (row.permission),
        placeholder: (__VLS_ctx.t('plugin.menu_permission_placeholder', 'plugin:xxx:view')),
    }));
    const __VLS_300 = __VLS_299({
        modelValue: (row.permission),
        placeholder: (__VLS_ctx.t('plugin.menu_permission_placeholder', 'plugin:xxx:view')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_299));
}
var __VLS_297;
const __VLS_302 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_303 = __VLS_asFunctionalComponent(__VLS_302, new __VLS_302({
    label: (__VLS_ctx.t('plugin.menu_sort', '排序')),
    width: "90",
}));
const __VLS_304 = __VLS_303({
    label: (__VLS_ctx.t('plugin.menu_sort', '排序')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_303));
__VLS_305.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_305.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_306 = {}.ElInputNumber;
    /** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
    // @ts-ignore
    const __VLS_307 = __VLS_asFunctionalComponent(__VLS_306, new __VLS_306({
        modelValue: (row.sort),
        min: (0),
        step: (1),
        ...{ style: {} },
    }));
    const __VLS_308 = __VLS_307({
        modelValue: (row.sort),
        min: (0),
        step: (1),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_307));
}
var __VLS_305;
const __VLS_310 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_311 = __VLS_asFunctionalComponent(__VLS_310, new __VLS_310({
    label: (__VLS_ctx.t('plugin.menu_visible', '可见')),
    width: "90",
}));
const __VLS_312 = __VLS_311({
    label: (__VLS_ctx.t('plugin.menu_visible', '可见')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_311));
__VLS_313.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_313.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_314 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_315 = __VLS_asFunctionalComponent(__VLS_314, new __VLS_314({
        modelValue: (row.visible),
    }));
    const __VLS_316 = __VLS_315({
        modelValue: (row.visible),
    }, ...__VLS_functionalComponentArgsRest(__VLS_315));
}
var __VLS_313;
const __VLS_318 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_319 = __VLS_asFunctionalComponent(__VLS_318, new __VLS_318({
    label: (__VLS_ctx.t('plugin.menu_enabled', '启用')),
    width: "90",
}));
const __VLS_320 = __VLS_319({
    label: (__VLS_ctx.t('plugin.menu_enabled', '启用')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_319));
__VLS_321.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_321.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_322 = {}.ElSwitch;
    /** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
    // @ts-ignore
    const __VLS_323 = __VLS_asFunctionalComponent(__VLS_322, new __VLS_322({
        modelValue: (row.enabled),
    }));
    const __VLS_324 = __VLS_323({
        modelValue: (row.enabled),
    }, ...__VLS_functionalComponentArgsRest(__VLS_323));
}
var __VLS_321;
const __VLS_326 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_327 = __VLS_asFunctionalComponent(__VLS_326, new __VLS_326({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "90",
    fixed: "right",
}));
const __VLS_328 = __VLS_327({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "90",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_327));
__VLS_329.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_329.slots;
    const [{ $index }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_330 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_331 = __VLS_asFunctionalComponent(__VLS_330, new __VLS_330({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_332 = __VLS_331({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_331));
    let __VLS_334;
    let __VLS_335;
    let __VLS_336;
    const __VLS_337 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeMenuRow($index);
        }
    };
    __VLS_333.slots.default;
    (__VLS_ctx.t('common.delete', '删除'));
    var __VLS_333;
}
var __VLS_329;
var __VLS_217;
var __VLS_205;
const __VLS_338 = {}.ElTabPane;
/** @type {[typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, typeof __VLS_components.ElTabPane, typeof __VLS_components.elTabPane, ]} */ ;
// @ts-ignore
const __VLS_339 = __VLS_asFunctionalComponent(__VLS_338, new __VLS_338({
    label: (__VLS_ctx.t('plugin.permissions_tab', '插件权限')),
    name: "permissions",
}));
const __VLS_340 = __VLS_339({
    label: (__VLS_ctx.t('plugin.permissions_tab', '插件权限')),
    name: "permissions",
}, ...__VLS_functionalComponentArgsRest(__VLS_339));
__VLS_341.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-table__actions mb-12" },
});
const __VLS_342 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_343 = __VLS_asFunctionalComponent(__VLS_342, new __VLS_342({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}));
const __VLS_344 = __VLS_343({
    ...{ 'onClick': {} },
    type: "primary",
    plain: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_343));
let __VLS_346;
let __VLS_347;
let __VLS_348;
const __VLS_349 = {
    onClick: (__VLS_ctx.appendPermissionRow)
};
__VLS_345.slots.default;
(__VLS_ctx.t('plugin.add_permission_row', '新增权限行'));
var __VLS_345;
const __VLS_350 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_351 = __VLS_asFunctionalComponent(__VLS_350, new __VLS_350({
    data: (__VLS_ctx.form.permissions),
    border: true,
    rowKey: "object",
    size: "small",
}));
const __VLS_352 = __VLS_351({
    data: (__VLS_ctx.form.permissions),
    border: true,
    rowKey: "object",
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_351));
__VLS_353.slots.default;
const __VLS_354 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_355 = __VLS_asFunctionalComponent(__VLS_354, new __VLS_354({
    label: (__VLS_ctx.t('plugin.permission_object', '对象')),
    minWidth: "180",
}));
const __VLS_356 = __VLS_355({
    label: (__VLS_ctx.t('plugin.permission_object', '对象')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_355));
__VLS_357.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_357.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_358 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_359 = __VLS_asFunctionalComponent(__VLS_358, new __VLS_358({
        modelValue: (row.object),
        placeholder: (__VLS_ctx.t('plugin.permission_object_placeholder', 'plugin:example')),
    }));
    const __VLS_360 = __VLS_359({
        modelValue: (row.object),
        placeholder: (__VLS_ctx.t('plugin.permission_object_placeholder', 'plugin:example')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_359));
}
var __VLS_357;
const __VLS_362 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_363 = __VLS_asFunctionalComponent(__VLS_362, new __VLS_362({
    label: (__VLS_ctx.t('plugin.permission_action', '动作')),
    minWidth: "140",
}));
const __VLS_364 = __VLS_363({
    label: (__VLS_ctx.t('plugin.permission_action', '动作')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_363));
__VLS_365.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_365.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_366 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_367 = __VLS_asFunctionalComponent(__VLS_366, new __VLS_366({
        modelValue: (row.action),
        placeholder: (__VLS_ctx.t('plugin.permission_action_placeholder', 'view')),
    }));
    const __VLS_368 = __VLS_367({
        modelValue: (row.action),
        placeholder: (__VLS_ctx.t('plugin.permission_action_placeholder', 'view')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_367));
}
var __VLS_365;
const __VLS_370 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_371 = __VLS_asFunctionalComponent(__VLS_370, new __VLS_370({
    label: (__VLS_ctx.t('plugin.permission_description', '描述')),
    minWidth: "260",
}));
const __VLS_372 = __VLS_371({
    label: (__VLS_ctx.t('plugin.permission_description', '描述')),
    minWidth: "260",
}, ...__VLS_functionalComponentArgsRest(__VLS_371));
__VLS_373.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_373.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_374 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_375 = __VLS_asFunctionalComponent(__VLS_374, new __VLS_374({
        modelValue: (row.description),
        placeholder: (__VLS_ctx.t('plugin.permission_description_placeholder', '权限描述')),
    }));
    const __VLS_376 = __VLS_375({
        modelValue: (row.description),
        placeholder: (__VLS_ctx.t('plugin.permission_description_placeholder', '权限描述')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_375));
}
var __VLS_373;
const __VLS_378 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_379 = __VLS_asFunctionalComponent(__VLS_378, new __VLS_378({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "90",
    fixed: "right",
}));
const __VLS_380 = __VLS_379({
    label: (__VLS_ctx.t('plugin.actions', '操作')),
    width: "90",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_379));
__VLS_381.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_381.slots;
    const [{ $index }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_382 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_383 = __VLS_asFunctionalComponent(__VLS_382, new __VLS_382({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_384 = __VLS_383({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_383));
    let __VLS_386;
    let __VLS_387;
    let __VLS_388;
    const __VLS_389 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removePermissionRow($index);
        }
    };
    __VLS_385.slots.default;
    (__VLS_ctx.t('common.delete', '删除'));
    var __VLS_385;
}
var __VLS_381;
var __VLS_353;
var __VLS_341;
var __VLS_169;
var __VLS_161;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-filters']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-center-tabs']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-table__actions']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-table__actions']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-12']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AdminFormDialog: AdminFormDialog,
            AdminTable: AdminTable,
            formatDateTime: formatDateTime,
            statusTagType: statusTagType,
            tableLoading: tableLoading,
            dialogLoading: dialogLoading,
            dialogVisible: dialogVisible,
            activeTab: activeTab,
            editingName: editingName,
            t: t,
            query: query,
            form: form,
            filteredRows: filteredRows,
            pluginCount: pluginCount,
            enabledCount: enabledCount,
            parentMenuOptions: parentMenuOptions,
            loadPlugins: loadPlugins,
            openDetail: openDetail,
            openCreate: openCreate,
            openEdit: openEdit,
            handleSearch: handleSearch,
            handleReset: handleReset,
            appendMenuRow: appendMenuRow,
            removeMenuRow: removeMenuRow,
            appendPermissionRow: appendPermissionRow,
            removePermissionRow: removePermissionRow,
            submitForm: submitForm,
            removeRow: removeRow,
            statusLabel: statusLabel,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
