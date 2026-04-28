import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchMenuTree } from '@/api/system-menus';
import { createRole, deleteRole, fetchRoles, updateRole } from '@/api/roles';
import { useAppI18n } from '@/i18n';
import { flattenMenuItems, formatDateTime, statusTagType } from '@/utils/admin';
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const menuLoading = ref(false);
const rows = ref([]);
const total = ref(0);
const menuTree = ref([]);
const editingId = ref('');
const { t } = useAppI18n();
const query = reactive({
    keyword: '',
    status: '',
    page: 1,
    page_size: 10,
});
const defaultForm = () => ({
    tenant_id: '',
    name: '',
    code: '',
    status: 'active',
    remark: '',
    menu_ids: [],
});
const form = reactive(defaultForm());
function getMenuDisplayTitle(item) {
    return t(item.titleKey || '', item.titleDefault || item.name || t('menu.unnamed', 'Unnamed menu'));
}
const menuOptions = computed(() => flattenMenuItems(menuTree.value));
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadRoles() {
    tableLoading.value = true;
    try {
        const response = await fetchRoles({ ...query });
        rows.value = response.items;
        total.value = response.total;
    }
    finally {
        tableLoading.value = false;
    }
}
async function loadMenuTree() {
    menuLoading.value = true;
    try {
        const response = await fetchMenuTree();
        menuTree.value = response.items ?? [];
    }
    finally {
        menuLoading.value = false;
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
        tenant_id: row.tenant_id ?? '',
        name: row.name,
        code: row.code,
        status: row.status || 'active',
        remark: row.remark ?? '',
        menu_ids: [...(row.menu_ids ?? [])],
    });
    dialogVisible.value = true;
}
function statusLabel(status) {
    return status === 'inactive' ? t('role.status.inactive', 'Disabled') : t('role.status.active', 'Enabled');
}
async function submitForm() {
    if (form.name.trim() === '' || form.code.trim() === '') {
        ElMessage.warning(t('role.validation_required', 'Enter the role name and code'));
        return;
    }
    dialogLoading.value = true;
    try {
        const payload = {
            ...form,
            tenant_id: form.tenant_id.trim(),
            name: form.name.trim(),
            code: form.code.trim(),
            status: form.status.trim() || 'active',
            remark: form.remark.trim(),
            menu_ids: [...form.menu_ids],
        };
        if (editingId.value) {
            await updateRole(editingId.value, payload);
            ElMessage.success(t('role.updated', 'Role updated'));
        }
        else {
            await createRole(payload);
            ElMessage.success(t('role.created', 'Role created'));
        }
        dialogVisible.value = false;
        await loadRoles();
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('role.confirm_delete', 'Delete role {name}?', { name: row.name }), t('role.delete_title', 'Delete role'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteRole(row.id);
    ElMessage.success(t('role.deleted', 'Role deleted'));
    await loadRoles();
}
function handleSearch() {
    query.page = 1;
    void loadRoles();
}
function handleReset() {
    query.keyword = '';
    query.status = '';
    query.page = 1;
    void loadRoles();
}
function handlePageChange(page) {
    query.page = page;
    void loadRoles();
}
function handleSizeChange(pageSize) {
    query.page_size = pageSize;
    query.page = 1;
    void loadRoles();
}
onMounted(() => {
    void Promise.all([loadRoles(), loadMenuTree()]);
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
    title: (__VLS_ctx.t('role.title', 'Role management')),
    description: (__VLS_ctx.t('role.description', 'Maintain role basics and menu bindings.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('role.title', 'Role management')),
    description: (__VLS_ctx.t('role.description', 'Maintain role basics and menu bindings.')),
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
        onClick: (__VLS_ctx.loadRoles)
    };
    __VLS_6.slots.default;
    (__VLS_ctx.t('common.refresh', 'Refresh'));
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('role:create') }, null, null);
    __VLS_14.slots.default;
    (__VLS_ctx.t('common.create', 'Create'));
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
        label: (__VLS_ctx.t('role.keyword_label', 'Keyword')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('role.keyword_label', 'Keyword')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('role.keyword_placeholder', 'Role name / code')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('role.keyword_placeholder', 'Role name / code')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: (__VLS_ctx.t('role.status_label', 'Status')),
    }));
    const __VLS_33 = __VLS_32({
        label: (__VLS_ctx.t('role.status_label', 'Status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('role.status_placeholder', 'All statuses')),
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('role.status_placeholder', 'All statuses')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    const __VLS_39 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: (__VLS_ctx.t('role.status.active', 'Enabled')),
        value: "active",
    }));
    const __VLS_41 = __VLS_40({
        label: (__VLS_ctx.t('role.status.active', 'Enabled')),
        value: "active",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    const __VLS_43 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: (__VLS_ctx.t('role.status.inactive', 'Disabled')),
        value: "inactive",
    }));
    const __VLS_45 = __VLS_44({
        label: (__VLS_ctx.t('role.status.inactive', 'Disabled')),
        value: "inactive",
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
    (__VLS_ctx.t('common.search', 'Search'));
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
    (__VLS_ctx.t('common.reset', 'Reset'));
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
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    prop: "name",
    label: (__VLS_ctx.t('role.name', 'Role name')),
    minWidth: "140",
}));
const __VLS_73 = __VLS_72({
    prop: "name",
    label: (__VLS_ctx.t('role.name', 'Role name')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "code",
    label: (__VLS_ctx.t('role.code', 'Role code')),
    minWidth: "140",
}));
const __VLS_77 = __VLS_76({
    prop: "code",
    label: (__VLS_ctx.t('role.code', 'Role code')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    label: (__VLS_ctx.t('role.status', 'Status')),
    width: "100",
}));
const __VLS_81 = __VLS_80({
    label: (__VLS_ctx.t('role.status', 'Status')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_82.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_83 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }));
    const __VLS_85 = __VLS_84({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_84));
    __VLS_86.slots.default;
    (__VLS_ctx.statusLabel(row.status));
    var __VLS_86;
}
var __VLS_82;
const __VLS_87 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    prop: "remark",
    label: (__VLS_ctx.t('role.remark', 'Remark')),
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_89 = __VLS_88({
    prop: "remark",
    label: (__VLS_ctx.t('role.remark', 'Remark')),
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: (__VLS_ctx.t('role.menu_count', 'Menu count')),
    width: "110",
}));
const __VLS_93 = __VLS_92({
    label: (__VLS_ctx.t('role.menu_count', 'Menu count')),
    width: "110",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_94.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.menu_ids?.length ?? 0);
}
var __VLS_94;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: (__VLS_ctx.t('role.created_at', 'Created at')),
    minWidth: "180",
}));
const __VLS_97 = __VLS_96({
    label: (__VLS_ctx.t('role.created_at', 'Created at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_98.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_98;
const __VLS_99 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    label: (__VLS_ctx.t('role.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_101 = __VLS_100({
    label: (__VLS_ctx.t('role.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_102.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_102.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_103 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_105 = __VLS_104({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_104));
    let __VLS_107;
    let __VLS_108;
    let __VLS_109;
    const __VLS_110 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('role:update') }, null, null);
    __VLS_106.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_106;
    const __VLS_111 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_113 = __VLS_112({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_112));
    let __VLS_115;
    let __VLS_116;
    let __VLS_117;
    const __VLS_118 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('role:delete') }, null, null);
    __VLS_114.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_114;
}
var __VLS_102;
var __VLS_70;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_119 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_121 = __VLS_120({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    let __VLS_123;
    let __VLS_124;
    let __VLS_125;
    const __VLS_126 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_127 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_122;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_128 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('role.edit_title', 'Edit role') : __VLS_ctx.t('role.create_title', 'New role')),
    loading: (__VLS_ctx.dialogLoading),
}));
const __VLS_129 = __VLS_128({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('role.edit_title', 'Edit role') : __VLS_ctx.t('role.create_title', 'New role')),
    loading: (__VLS_ctx.dialogLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_128));
let __VLS_131;
let __VLS_132;
let __VLS_133;
const __VLS_134 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_130.slots.default;
const __VLS_135 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_137 = __VLS_136({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
__VLS_138.slots.default;
const __VLS_139 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
    label: (__VLS_ctx.t('role.name', 'Role name')),
    required: true,
}));
const __VLS_141 = __VLS_140({
    label: (__VLS_ctx.t('role.name', 'Role name')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
__VLS_142.slots.default;
const __VLS_143 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('role.name_placeholder', 'Enter the role name')),
}));
const __VLS_145 = __VLS_144({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('role.name_placeholder', 'Enter the role name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
var __VLS_142;
const __VLS_147 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
    label: (__VLS_ctx.t('role.code', 'Role code')),
    required: true,
}));
const __VLS_149 = __VLS_148({
    label: (__VLS_ctx.t('role.code', 'Role code')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
__VLS_150.slots.default;
const __VLS_151 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    modelValue: (__VLS_ctx.form.code),
    placeholder: (__VLS_ctx.t('role.code_placeholder', 'Enter the role code')),
}));
const __VLS_153 = __VLS_152({
    modelValue: (__VLS_ctx.form.code),
    placeholder: (__VLS_ctx.t('role.code_placeholder', 'Enter the role code')),
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
var __VLS_150;
const __VLS_155 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    label: (__VLS_ctx.t('role.status', 'Status')),
}));
const __VLS_157 = __VLS_156({
    label: (__VLS_ctx.t('role.status', 'Status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
__VLS_158.slots.default;
const __VLS_159 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}));
const __VLS_161 = __VLS_160({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    label: (__VLS_ctx.t('role.status.active', 'Enabled')),
    value: "active",
}));
const __VLS_165 = __VLS_164({
    label: (__VLS_ctx.t('role.status.active', 'Enabled')),
    value: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
const __VLS_167 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('role.status.inactive', 'Disabled')),
    value: "inactive",
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('role.status.inactive', 'Disabled')),
    value: "inactive",
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
var __VLS_162;
var __VLS_158;
const __VLS_171 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    label: (__VLS_ctx.t('role.remark', 'Remark')),
}));
const __VLS_173 = __VLS_172({
    label: (__VLS_ctx.t('role.remark', 'Remark')),
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
__VLS_174.slots.default;
const __VLS_175 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('role.remark_placeholder', 'Enter a remark')),
}));
const __VLS_177 = __VLS_176({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('role.remark_placeholder', 'Enter a remark')),
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
var __VLS_174;
const __VLS_179 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    label: (__VLS_ctx.t('role.menu_permissions', 'Menu permissions')),
}));
const __VLS_181 = __VLS_180({
    label: (__VLS_ctx.t('role.menu_permissions', 'Menu permissions')),
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
__VLS_182.slots.default;
const __VLS_183 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    modelValue: (__VLS_ctx.form.menu_ids),
    multiple: true,
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.menuLoading),
    placeholder: (__VLS_ctx.t('role.menu_permissions_placeholder', 'Select the menus this role can access')),
}));
const __VLS_185 = __VLS_184({
    modelValue: (__VLS_ctx.form.menu_ids),
    multiple: true,
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.menuLoading),
    placeholder: (__VLS_ctx.t('role.menu_permissions_placeholder', 'Select the menus this role can access')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
for (const [menu] of __VLS_getVForSourceType((__VLS_ctx.menuOptions))) {
    const __VLS_187 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
        key: (menu.id),
        label: (`${__VLS_ctx.getMenuDisplayTitle(menu)} (${menu.path})`),
        value: (menu.id),
    }));
    const __VLS_189 = __VLS_188({
        key: (menu.id),
        label: (`${__VLS_ctx.getMenuDisplayTitle(menu)} (${menu.path})`),
        value: (menu.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_188));
}
var __VLS_186;
var __VLS_182;
var __VLS_138;
var __VLS_130;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-filters']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-pagination']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-form']} */ ;
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
            menuLoading: menuLoading,
            rows: rows,
            total: total,
            editingId: editingId,
            t: t,
            query: query,
            form: form,
            getMenuDisplayTitle: getMenuDisplayTitle,
            menuOptions: menuOptions,
            loadRoles: loadRoles,
            openCreate: openCreate,
            openEdit: openEdit,
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
