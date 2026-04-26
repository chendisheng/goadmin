import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createMenu, deleteMenu, fetchMenuTree, fetchMenus, updateMenu } from '@/api/system-menus';
import { flattenMenuItems, formatDateTime, menuTypeTagType, statusTagType } from '@/utils/admin';
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const parentLoading = ref(false);
const rows = ref([]);
const total = ref(0);
const menuTree = ref([]);
const editingId = ref('');
const { t } = useAppI18n();
const query = reactive({
    keyword: '',
    parent_id: '',
    page: 1,
    page_size: 10,
});
const defaultForm = () => ({
    parent_id: '',
    name: '',
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
});
const form = reactive(defaultForm());
function getMenuDisplayTitle(item) {
    return t(item.titleKey || '', item.titleDefault || item.name);
}
const parentOptions = computed(() => flattenMenuItems(menuTree.value).map((item) => ({
    label: `${item.path} - ${getMenuDisplayTitle(item)}`,
    value: item.id,
})));
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadMenus() {
    tableLoading.value = true;
    try {
        const response = await fetchMenus({ ...query });
        rows.value = response.items;
        total.value = response.total;
    }
    finally {
        tableLoading.value = false;
    }
}
async function loadMenuTree() {
    parentLoading.value = true;
    try {
        const response = await fetchMenuTree();
        menuTree.value = response.items ?? [];
    }
    finally {
        parentLoading.value = false;
    }
}
function openCreate() {
    editingId.value = '';
    resetForm();
    dialogVisible.value = true;
}
function openEdit(row) {
    editingId.value = row.id;
    Object.assign(form, {
        ...defaultForm(),
        parent_id: row.parent_id ?? '',
        name: row.name,
        path: row.path,
        component: row.component ?? '',
        icon: row.icon ?? '',
        sort: row.sort ?? 0,
        permission: row.permission ?? '',
        type: row.type || 'menu',
        visible: row.visible,
        enabled: row.enabled,
        redirect: row.redirect ?? '',
        external_url: row.external_url ?? '',
    });
    dialogVisible.value = true;
}
function typeLabel(type) {
    switch (type) {
        case 'directory':
            return t('menu.type.directory', '目录');
        case 'button':
            return t('menu.type.button', '按钮');
        default:
            return t('menu.type.menu', '菜单');
    }
}
function statusLabel(flag) {
    return flag ? t('menu.status.active', '启用') : t('menu.status.inactive', '禁用');
}
async function submitForm() {
    if (form.name.trim() === '' || form.path.trim() === '') {
        ElMessage.warning(t('menu.validate_required', '请输入菜单名称和路径'));
        return;
    }
    dialogLoading.value = true;
    try {
        const payload = {
            ...form,
            parent_id: form.parent_id.trim(),
            name: form.name.trim(),
            path: form.path.trim(),
            component: form.component.trim(),
            icon: form.icon.trim(),
            sort: Number(form.sort) || 0,
            permission: form.permission.trim(),
            type: form.type.trim() || 'menu',
            visible: Boolean(form.visible),
            enabled: Boolean(form.enabled),
            redirect: form.redirect.trim(),
            external_url: form.external_url.trim(),
        };
        if (editingId.value) {
            await updateMenu(editingId.value, payload);
            ElMessage.success(t('menu.updated', '菜单已更新'));
        }
        else {
            await createMenu(payload);
            ElMessage.success(t('menu.created', '菜单已创建'));
        }
        dialogVisible.value = false;
        await Promise.all([loadMenus(), loadMenuTree()]);
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('menu.confirm_delete', '确认删除菜单 {name} 吗？', { name: row.name }), t('menu.delete_title', '删除菜单'), {
        type: 'warning',
        confirmButtonText: t('menu.delete_confirm', '删除'),
        cancelButtonText: t('menu.delete_cancel', '取消'),
    });
    await deleteMenu(row.id);
    ElMessage.success(t('menu.deleted', '菜单已删除'));
    await Promise.all([loadMenus(), loadMenuTree()]);
}
function handleSearch() {
    query.page = 1;
    void loadMenus();
}
function handleReset() {
    query.keyword = '';
    query.parent_id = '';
    query.page = 1;
    void loadMenus();
}
function handlePageChange(page) {
    query.page = page;
    void loadMenus();
}
function handleSizeChange(pageSize) {
    query.page_size = pageSize;
    query.page = 1;
    void loadMenus();
}
onMounted(() => {
    void Promise.all([loadMenus(), loadMenuTree()]);
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
    title: (__VLS_ctx.t('menu.title', '菜单管理')),
    description: (__VLS_ctx.t('menu.description', '维护系统菜单、路由和权限元数据。')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('menu.title', '菜单管理')),
    description: (__VLS_ctx.t('menu.description', '维护系统菜单、路由和权限元数据。')),
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
        onClick: (__VLS_ctx.loadMenus)
    };
    __VLS_6.slots.default;
    (__VLS_ctx.t('menu.refresh', '刷新'));
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('menu:create') }, null, null);
    __VLS_14.slots.default;
    (__VLS_ctx.t('menu.create', '新增菜单'));
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
        placeholder: (__VLS_ctx.t('menu.keyword_placeholder', '菜单名称 / 路径 / 权限')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('menu.keyword_placeholder', '菜单名称 / 路径 / 权限')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: (__VLS_ctx.t('menu.parent', '父级菜单')),
    }));
    const __VLS_33 = __VLS_32({
        label: (__VLS_ctx.t('menu.parent', '父级菜单')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.parent_id),
        clearable: true,
        filterable: true,
        loading: (__VLS_ctx.parentLoading),
        placeholder: (__VLS_ctx.t('menu.parent_placeholder', '全部父级')),
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.parent_id),
        clearable: true,
        filterable: true,
        loading: (__VLS_ctx.parentLoading),
        placeholder: (__VLS_ctx.t('menu.parent_placeholder', '全部父级')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    const __VLS_39 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: (__VLS_ctx.t('menu.top_level', '顶级菜单')),
        value: "",
    }));
    const __VLS_41 = __VLS_40({
        label: (__VLS_ctx.t('menu.top_level', '顶级菜单')),
        value: "",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    for (const [menu] of __VLS_getVForSourceType((__VLS_ctx.parentOptions))) {
        const __VLS_43 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
            key: (menu.value),
            label: (menu.label),
            value: (menu.value),
        }));
        const __VLS_45 = __VLS_44({
            key: (menu.value),
            label: (menu.label),
            value: (menu.value),
        }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    }
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
const __VLS_67 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}));
const __VLS_69 = __VLS_68({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_70.slots.default;
const __VLS_71 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    label: (__VLS_ctx.t('menu.name', '名称')),
    minWidth: "140",
}));
const __VLS_73 = __VLS_72({
    label: (__VLS_ctx.t('menu.name', '名称')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_74.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.getMenuDisplayTitle(row));
}
var __VLS_74;
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "path",
    label: (__VLS_ctx.t('menu.path', '路径')),
    minWidth: "160",
}));
const __VLS_77 = __VLS_76({
    prop: "path",
    label: (__VLS_ctx.t('menu.path', '路径')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "component",
    label: (__VLS_ctx.t('menu.component', '组件')),
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_81 = __VLS_80({
    prop: "component",
    label: (__VLS_ctx.t('menu.component', '组件')),
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    label: (__VLS_ctx.t('menu.type', '类型')),
    width: "100",
}));
const __VLS_85 = __VLS_84({
    label: (__VLS_ctx.t('menu.type', '类型')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_86.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_87 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
        type: (__VLS_ctx.menuTypeTagType(row.type)),
        effect: "plain",
    }));
    const __VLS_89 = __VLS_88({
        type: (__VLS_ctx.menuTypeTagType(row.type)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_88));
    __VLS_90.slots.default;
    (__VLS_ctx.typeLabel(row.type));
    var __VLS_90;
}
var __VLS_86;
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: (__VLS_ctx.t('menu.permission', '权限')),
    minWidth: "160",
}));
const __VLS_93 = __VLS_92({
    label: (__VLS_ctx.t('menu.permission', '权限')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_94.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.permission || '-');
}
var __VLS_94;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: (__VLS_ctx.t('menu.visible', '可见')),
    width: "90",
}));
const __VLS_97 = __VLS_96({
    label: (__VLS_ctx.t('menu.visible', '可见')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_98.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_99 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
        type: (__VLS_ctx.statusTagType(row.visible ? 'active' : 'inactive')),
        effect: "plain",
    }));
    const __VLS_101 = __VLS_100({
        type: (__VLS_ctx.statusTagType(row.visible ? 'active' : 'inactive')),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_100));
    __VLS_102.slots.default;
    (__VLS_ctx.statusLabel(row.visible));
    var __VLS_102;
}
var __VLS_98;
const __VLS_103 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    label: (__VLS_ctx.t('menu.enabled', '启用')),
    width: "90",
}));
const __VLS_105 = __VLS_104({
    label: (__VLS_ctx.t('menu.enabled', '启用')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
__VLS_106.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_106.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_107 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
        type: (__VLS_ctx.statusTagType(row.enabled ? 'active' : 'inactive')),
        effect: "plain",
    }));
    const __VLS_109 = __VLS_108({
        type: (__VLS_ctx.statusTagType(row.enabled ? 'active' : 'inactive')),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_108));
    __VLS_110.slots.default;
    (__VLS_ctx.statusLabel(row.enabled));
    var __VLS_110;
}
var __VLS_106;
const __VLS_111 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('menu.created_at', '创建时间')),
    minWidth: "180",
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('menu.created_at', '创建时间')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
__VLS_114.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_114.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_114;
const __VLS_115 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
    label: (__VLS_ctx.t('menu.actions', '操作')),
    width: "180",
    fixed: "right",
}));
const __VLS_117 = __VLS_116({
    label: (__VLS_ctx.t('menu.actions', '操作')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
__VLS_118.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_118.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_119 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_121 = __VLS_120({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    let __VLS_123;
    let __VLS_124;
    let __VLS_125;
    const __VLS_126 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('menu:update') }, null, null);
    __VLS_122.slots.default;
    (__VLS_ctx.t('menu.edit', '编辑'));
    var __VLS_122;
    const __VLS_127 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_129 = __VLS_128({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_128));
    let __VLS_131;
    let __VLS_132;
    let __VLS_133;
    const __VLS_134 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('menu:delete') }, null, null);
    __VLS_130.slots.default;
    (__VLS_ctx.t('menu.delete', '删除'));
    var __VLS_130;
}
var __VLS_118;
var __VLS_70;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_135 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_137 = __VLS_136({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_136));
    let __VLS_139;
    let __VLS_140;
    let __VLS_141;
    const __VLS_142 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_143 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_138;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('menu.edit_title', '编辑菜单') : __VLS_ctx.t('menu.create_title', '新增菜单')),
    loading: (__VLS_ctx.dialogLoading),
    width: "860px",
}));
const __VLS_145 = __VLS_144({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('menu.edit_title', '编辑菜单') : __VLS_ctx.t('menu.create_title', '新增菜单')),
    loading: (__VLS_ctx.dialogLoading),
    width: "860px",
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
let __VLS_147;
let __VLS_148;
let __VLS_149;
const __VLS_150 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_146.slots.default;
const __VLS_151 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    labelWidth: "110px",
    ...{ class: "admin-form admin-form--two-col" },
}));
const __VLS_153 = __VLS_152({
    labelWidth: "110px",
    ...{ class: "admin-form admin-form--two-col" },
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
const __VLS_155 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    label: (__VLS_ctx.t('menu.parent', '父级菜单')),
}));
const __VLS_157 = __VLS_156({
    label: (__VLS_ctx.t('menu.parent', '父级菜单')),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
__VLS_158.slots.default;
const __VLS_159 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    modelValue: (__VLS_ctx.form.parent_id),
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.parentLoading),
    placeholder: (__VLS_ctx.t('menu.parent_placeholder', '选择父级菜单')),
}));
const __VLS_161 = __VLS_160({
    modelValue: (__VLS_ctx.form.parent_id),
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.parentLoading),
    placeholder: (__VLS_ctx.t('menu.parent_placeholder', '选择父级菜单')),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    label: (__VLS_ctx.t('menu.top_level', '顶级菜单')),
    value: "",
}));
const __VLS_165 = __VLS_164({
    label: (__VLS_ctx.t('menu.top_level', '顶级菜单')),
    value: "",
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
for (const [menu] of __VLS_getVForSourceType((__VLS_ctx.parentOptions))) {
    const __VLS_167 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
        key: (menu.value),
        label: (menu.label),
        value: (menu.value),
    }));
    const __VLS_169 = __VLS_168({
        key: (menu.value),
        label: (menu.label),
        value: (menu.value),
    }, ...__VLS_functionalComponentArgsRest(__VLS_168));
}
var __VLS_162;
var __VLS_158;
const __VLS_171 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    label: (__VLS_ctx.t('menu.name', '菜单名称')),
    required: true,
}));
const __VLS_173 = __VLS_172({
    label: (__VLS_ctx.t('menu.name', '菜单名称')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
__VLS_174.slots.default;
const __VLS_175 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('menu.name_placeholder', '请输入菜单名称')),
}));
const __VLS_177 = __VLS_176({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('menu.name_placeholder', '请输入菜单名称')),
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
var __VLS_174;
const __VLS_179 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    label: (__VLS_ctx.t('menu.path', '路径')),
    required: true,
}));
const __VLS_181 = __VLS_180({
    label: (__VLS_ctx.t('menu.path', '路径')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
__VLS_182.slots.default;
const __VLS_183 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    modelValue: (__VLS_ctx.form.path),
    placeholder: (__VLS_ctx.t('menu.path_placeholder', '请输入路由路径')),
}));
const __VLS_185 = __VLS_184({
    modelValue: (__VLS_ctx.form.path),
    placeholder: (__VLS_ctx.t('menu.path_placeholder', '请输入路由路径')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
var __VLS_182;
const __VLS_187 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    label: (__VLS_ctx.t('menu.component', '组件路径')),
}));
const __VLS_189 = __VLS_188({
    label: (__VLS_ctx.t('menu.component', '组件路径')),
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
__VLS_190.slots.default;
const __VLS_191 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    modelValue: (__VLS_ctx.form.component),
    placeholder: (__VLS_ctx.t('menu.component_placeholder', '例如 view/system/user/index')),
}));
const __VLS_193 = __VLS_192({
    modelValue: (__VLS_ctx.form.component),
    placeholder: (__VLS_ctx.t('menu.component_placeholder', '例如 view/system/user/index')),
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
var __VLS_190;
const __VLS_195 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
    label: (__VLS_ctx.t('menu.icon', '图标')),
}));
const __VLS_197 = __VLS_196({
    label: (__VLS_ctx.t('menu.icon', '图标')),
}, ...__VLS_functionalComponentArgsRest(__VLS_196));
__VLS_198.slots.default;
const __VLS_199 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_200 = __VLS_asFunctionalComponent(__VLS_199, new __VLS_199({
    modelValue: (__VLS_ctx.form.icon),
    placeholder: (__VLS_ctx.t('menu.icon_placeholder', '例如 user / setting')),
}));
const __VLS_201 = __VLS_200({
    modelValue: (__VLS_ctx.form.icon),
    placeholder: (__VLS_ctx.t('menu.icon_placeholder', '例如 user / setting')),
}, ...__VLS_functionalComponentArgsRest(__VLS_200));
var __VLS_198;
const __VLS_203 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({
    label: (__VLS_ctx.t('menu.sort', '排序')),
}));
const __VLS_205 = __VLS_204({
    label: (__VLS_ctx.t('menu.sort', '排序')),
}, ...__VLS_functionalComponentArgsRest(__VLS_204));
__VLS_206.slots.default;
const __VLS_207 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_208 = __VLS_asFunctionalComponent(__VLS_207, new __VLS_207({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}));
const __VLS_209 = __VLS_208({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_208));
var __VLS_206;
const __VLS_211 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
    label: (__VLS_ctx.t('menu.permission', '权限标识')),
}));
const __VLS_213 = __VLS_212({
    label: (__VLS_ctx.t('menu.permission', '权限标识')),
}, ...__VLS_functionalComponentArgsRest(__VLS_212));
__VLS_214.slots.default;
const __VLS_215 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_216 = __VLS_asFunctionalComponent(__VLS_215, new __VLS_215({
    modelValue: (__VLS_ctx.form.permission),
    placeholder: (__VLS_ctx.t('menu.permission_placeholder', '例如 user:list')),
}));
const __VLS_217 = __VLS_216({
    modelValue: (__VLS_ctx.form.permission),
    placeholder: (__VLS_ctx.t('menu.permission_placeholder', '例如 user:list')),
}, ...__VLS_functionalComponentArgsRest(__VLS_216));
var __VLS_214;
const __VLS_219 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_220 = __VLS_asFunctionalComponent(__VLS_219, new __VLS_219({
    label: (__VLS_ctx.t('menu.type', '类型')),
}));
const __VLS_221 = __VLS_220({
    label: (__VLS_ctx.t('menu.type', '类型')),
}, ...__VLS_functionalComponentArgsRest(__VLS_220));
__VLS_222.slots.default;
const __VLS_223 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_224 = __VLS_asFunctionalComponent(__VLS_223, new __VLS_223({
    modelValue: (__VLS_ctx.form.type),
    ...{ style: {} },
}));
const __VLS_225 = __VLS_224({
    modelValue: (__VLS_ctx.form.type),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_224));
__VLS_226.slots.default;
const __VLS_227 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_228 = __VLS_asFunctionalComponent(__VLS_227, new __VLS_227({
    label: (__VLS_ctx.t('menu.type.directory', '目录')),
    value: "directory",
}));
const __VLS_229 = __VLS_228({
    label: (__VLS_ctx.t('menu.type.directory', '目录')),
    value: "directory",
}, ...__VLS_functionalComponentArgsRest(__VLS_228));
const __VLS_231 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_232 = __VLS_asFunctionalComponent(__VLS_231, new __VLS_231({
    label: (__VLS_ctx.t('menu.type.menu', '菜单')),
    value: "menu",
}));
const __VLS_233 = __VLS_232({
    label: (__VLS_ctx.t('menu.type.menu', '菜单')),
    value: "menu",
}, ...__VLS_functionalComponentArgsRest(__VLS_232));
const __VLS_235 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_236 = __VLS_asFunctionalComponent(__VLS_235, new __VLS_235({
    label: (__VLS_ctx.t('menu.type.button', '按钮')),
    value: "button",
}));
const __VLS_237 = __VLS_236({
    label: (__VLS_ctx.t('menu.type.button', '按钮')),
    value: "button",
}, ...__VLS_functionalComponentArgsRest(__VLS_236));
var __VLS_226;
var __VLS_222;
const __VLS_239 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_240 = __VLS_asFunctionalComponent(__VLS_239, new __VLS_239({
    label: (__VLS_ctx.t('menu.redirect', '重定向')),
}));
const __VLS_241 = __VLS_240({
    label: (__VLS_ctx.t('menu.redirect', '重定向')),
}, ...__VLS_functionalComponentArgsRest(__VLS_240));
__VLS_242.slots.default;
const __VLS_243 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_244 = __VLS_asFunctionalComponent(__VLS_243, new __VLS_243({
    modelValue: (__VLS_ctx.form.redirect),
    placeholder: (__VLS_ctx.t('menu.redirect_placeholder', '例如 /system/users')),
}));
const __VLS_245 = __VLS_244({
    modelValue: (__VLS_ctx.form.redirect),
    placeholder: (__VLS_ctx.t('menu.redirect_placeholder', '例如 /system/users')),
}, ...__VLS_functionalComponentArgsRest(__VLS_244));
var __VLS_242;
const __VLS_247 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_248 = __VLS_asFunctionalComponent(__VLS_247, new __VLS_247({
    label: (__VLS_ctx.t('menu.external_url', '外链地址')),
}));
const __VLS_249 = __VLS_248({
    label: (__VLS_ctx.t('menu.external_url', '外链地址')),
}, ...__VLS_functionalComponentArgsRest(__VLS_248));
__VLS_250.slots.default;
const __VLS_251 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_252 = __VLS_asFunctionalComponent(__VLS_251, new __VLS_251({
    modelValue: (__VLS_ctx.form.external_url),
    placeholder: (__VLS_ctx.t('menu.external_url_placeholder', '外部链接时填写')),
}));
const __VLS_253 = __VLS_252({
    modelValue: (__VLS_ctx.form.external_url),
    placeholder: (__VLS_ctx.t('menu.external_url_placeholder', '外部链接时填写')),
}, ...__VLS_functionalComponentArgsRest(__VLS_252));
var __VLS_250;
const __VLS_255 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_256 = __VLS_asFunctionalComponent(__VLS_255, new __VLS_255({
    label: (__VLS_ctx.t('menu.visible', '可见')),
}));
const __VLS_257 = __VLS_256({
    label: (__VLS_ctx.t('menu.visible', '可见')),
}, ...__VLS_functionalComponentArgsRest(__VLS_256));
__VLS_258.slots.default;
const __VLS_259 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_260 = __VLS_asFunctionalComponent(__VLS_259, new __VLS_259({
    modelValue: (__VLS_ctx.form.visible),
}));
const __VLS_261 = __VLS_260({
    modelValue: (__VLS_ctx.form.visible),
}, ...__VLS_functionalComponentArgsRest(__VLS_260));
var __VLS_258;
const __VLS_263 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_264 = __VLS_asFunctionalComponent(__VLS_263, new __VLS_263({
    label: (__VLS_ctx.t('menu.enabled', '启用')),
}));
const __VLS_265 = __VLS_264({
    label: (__VLS_ctx.t('menu.enabled', '启用')),
}, ...__VLS_functionalComponentArgsRest(__VLS_264));
__VLS_266.slots.default;
const __VLS_267 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_268 = __VLS_asFunctionalComponent(__VLS_267, new __VLS_267({
    modelValue: (__VLS_ctx.form.enabled),
}));
const __VLS_269 = __VLS_268({
    modelValue: (__VLS_ctx.form.enabled),
}, ...__VLS_functionalComponentArgsRest(__VLS_268));
var __VLS_266;
var __VLS_154;
var __VLS_146;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-filters']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-pagination']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form--two-col']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AdminFormDialog: AdminFormDialog,
            AdminTable: AdminTable,
            formatDateTime: formatDateTime,
            menuTypeTagType: menuTypeTagType,
            statusTagType: statusTagType,
            tableLoading: tableLoading,
            dialogLoading: dialogLoading,
            dialogVisible: dialogVisible,
            parentLoading: parentLoading,
            rows: rows,
            total: total,
            editingId: editingId,
            t: t,
            query: query,
            form: form,
            getMenuDisplayTitle: getMenuDisplayTitle,
            parentOptions: parentOptions,
            loadMenus: loadMenus,
            openCreate: openCreate,
            openEdit: openEdit,
            typeLabel: typeLabel,
            statusLabel: statusLabel,
            submitForm: submitForm,
            removeRow: removeRow,
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
