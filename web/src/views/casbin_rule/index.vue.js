import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createCasbinRule, deleteCasbinRule, listcasbin_rules, updateCasbinRule } from '@/api/casbin_rule';
import { formatDateTime } from '@/utils/admin';
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref([]);
const total = ref(0);
const editingId = ref('');
const { t } = useAppI18n();
const query = reactive({
    keyword: '',
    page: 1,
    page_size: 10,
});
const defaultForm = () => ({
    ptype: '',
    v0: '',
    v1: '',
    v2: '',
    v3: '',
    v4: '',
    v5: '',
});
const form = reactive(defaultForm());
function formatEnumLabel(value, labelMap) {
    if (Array.isArray(value)) {
        if (value.length === 0) {
            return '-';
        }
        return value.map((item) => labelMap[String(item)] ?? String(item)).join(', ');
    }
    if (value === null || value === undefined || value === '') {
        return '-';
    }
    return labelMap[String(value)] ?? String(value);
}
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadItems() {
    tableLoading.value = true;
    try {
        const response = await listcasbin_rules({ ...query });
        rows.value = response.items ?? [];
        total.value = response.total ?? 0;
    }
    finally {
        tableLoading.value = false;
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
        ptype: row.ptype ?? '',
        v0: row.v0 ?? '',
        v1: row.v1 ?? '',
        v2: row.v2 ?? '',
        v3: row.v3 ?? '',
        v4: row.v4 ?? '',
        v5: row.v5 ?? '',
    });
    dialogVisible.value = true;
}
async function submitForm() {
    dialogLoading.value = true;
    try {
        const payload = {
            ptype: form.ptype.trim(),
            v0: form.v0.trim(),
            v1: form.v1.trim(),
            v2: form.v2.trim(),
            v3: form.v3.trim(),
            v4: form.v4.trim(),
            v5: form.v5.trim(),
        };
        if (editingId.value) {
            await updateCasbinRule(editingId.value, payload);
            ElMessage.success(t('casbin_rule.updated', 'CasbinRule updated'));
        }
        else {
            await createCasbinRule(payload);
            ElMessage.success(t('casbin_rule.created', 'CasbinRule created'));
        }
        dialogVisible.value = false;
        await loadItems();
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('casbin_rule.save_failed', 'Save failed'));
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('casbin_rule.confirm_delete', 'Delete CasbinRule {name}?', { name: row.id }), t('casbin_rule.delete_title', 'Delete CasbinRule'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteCasbinRule(row.id);
    ElMessage.success(t('casbin_rule.deleted', 'CasbinRule deleted'));
    await loadItems();
}
function handleSearch() {
    query.page = 1;
    void loadItems();
}
function handleReset() {
    query.keyword = '';
    query.page = 1;
    void loadItems();
}
function handlePageChange(page) {
    query.page = page;
    void loadItems();
}
function handleSizeChange(pageSize) {
    query.page_size = pageSize;
    query.page = 1;
    void loadItems();
}
onMounted(() => {
    void loadItems();
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
    title: (__VLS_ctx.t('casbin_rule.title', 'Rule management')),
    description: (__VLS_ctx.t('casbin_rule.description', 'Manage authorization policy rules, including listing, editing, and deletion.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('casbin_rule.title', 'Rule management')),
    description: (__VLS_ctx.t('casbin_rule.description', 'Manage authorization policy rules, including listing, editing, and deletion.')),
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
        onClick: (__VLS_ctx.loadItems)
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_rule:create') }, null, null);
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
        label: (__VLS_ctx.t('common.search', 'Search')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('common.search', 'Search')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('casbin_rule.keyword_placeholder', 'Search CasbinRule data')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('casbin_rule.keyword_placeholder', 'Search CasbinRule data')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({}));
    const __VLS_33 = __VLS_32({}, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_37 = __VLS_36({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    let __VLS_39;
    let __VLS_40;
    let __VLS_41;
    const __VLS_42 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_38.slots.default;
    (__VLS_ctx.t('common.search', 'Search'));
    var __VLS_38;
    const __VLS_43 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        ...{ 'onClick': {} },
    }));
    const __VLS_45 = __VLS_44({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    let __VLS_47;
    let __VLS_48;
    let __VLS_49;
    const __VLS_50 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_46.slots.default;
    (__VLS_ctx.t('common.reset', 'Reset'));
    var __VLS_46;
    var __VLS_34;
    var __VLS_22;
}
const __VLS_51 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}));
const __VLS_53 = __VLS_52({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_54.slots.default;
const __VLS_55 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
    prop: "id",
    label: (__VLS_ctx.t('casbin_rule.id', 'ID')),
    minWidth: "160",
}));
const __VLS_57 = __VLS_56({
    prop: "id",
    label: (__VLS_ctx.t('casbin_rule.id', 'ID')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
const __VLS_59 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    prop: "ptype",
    label: (__VLS_ctx.t('casbin_rule.ptype', 'Ptype')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_61 = __VLS_60({
    prop: "ptype",
    label: (__VLS_ctx.t('casbin_rule.ptype', 'Ptype')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_62.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.ptype || '-');
}
var __VLS_62;
const __VLS_63 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    prop: "v0",
    label: (__VLS_ctx.t('casbin_rule.v0', 'V0')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_65 = __VLS_64({
    prop: "v0",
    label: (__VLS_ctx.t('casbin_rule.v0', 'V0')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_66.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v0 || '-');
}
var __VLS_66;
const __VLS_67 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    prop: "v1",
    label: (__VLS_ctx.t('casbin_rule.v1', 'V1')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_69 = __VLS_68({
    prop: "v1",
    label: (__VLS_ctx.t('casbin_rule.v1', 'V1')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_70.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v1 || '-');
}
var __VLS_70;
const __VLS_71 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    prop: "v2",
    label: (__VLS_ctx.t('casbin_rule.v2', 'V2')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_73 = __VLS_72({
    prop: "v2",
    label: (__VLS_ctx.t('casbin_rule.v2', 'V2')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_74.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v2 || '-');
}
var __VLS_74;
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "v3",
    label: (__VLS_ctx.t('casbin_rule.v3', 'V3')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_77 = __VLS_76({
    prop: "v3",
    label: (__VLS_ctx.t('casbin_rule.v3', 'V3')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_78.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v3 || '-');
}
var __VLS_78;
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "v4",
    label: (__VLS_ctx.t('casbin_rule.v4', 'V4')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_81 = __VLS_80({
    prop: "v4",
    label: (__VLS_ctx.t('casbin_rule.v4', 'V4')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_82.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v4 || '-');
}
var __VLS_82;
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    prop: "v5",
    label: (__VLS_ctx.t('casbin_rule.v5', 'V5')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_85 = __VLS_84({
    prop: "v5",
    label: (__VLS_ctx.t('casbin_rule.v5', 'V5')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_86.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.v5 || '-');
}
var __VLS_86;
const __VLS_87 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    label: (__VLS_ctx.t('casbin_rule.created_at', 'Created at')),
    minWidth: "180",
}));
const __VLS_89 = __VLS_88({
    label: (__VLS_ctx.t('casbin_rule.created_at', 'Created at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_90.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_90;
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: (__VLS_ctx.t('casbin_rule.updated_at', 'Updated at')),
    minWidth: "180",
}));
const __VLS_93 = __VLS_92({
    label: (__VLS_ctx.t('casbin_rule.updated_at', 'Updated at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_94.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_94;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: (__VLS_ctx.t('casbin_rule.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_97 = __VLS_96({
    label: (__VLS_ctx.t('casbin_rule.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_98.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_99 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_101 = __VLS_100({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_100));
    let __VLS_103;
    let __VLS_104;
    let __VLS_105;
    const __VLS_106 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_rule:update') }, null, null);
    __VLS_102.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_102;
    const __VLS_107 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_109 = __VLS_108({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_108));
    let __VLS_111;
    let __VLS_112;
    let __VLS_113;
    const __VLS_114 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_rule:delete') }, null, null);
    __VLS_110.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_110;
}
var __VLS_98;
var __VLS_54;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_115 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_117 = __VLS_116({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_116));
    let __VLS_119;
    let __VLS_120;
    let __VLS_121;
    const __VLS_122 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_123 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_118;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_124 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('casbin_rule.edit_title', 'Edit rule') : __VLS_ctx.t('casbin_rule.create_title', 'New rule')),
    loading: (__VLS_ctx.dialogLoading),
}));
const __VLS_125 = __VLS_124({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('casbin_rule.edit_title', 'Edit rule') : __VLS_ctx.t('casbin_rule.create_title', 'New rule')),
    loading: (__VLS_ctx.dialogLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_124));
let __VLS_127;
let __VLS_128;
let __VLS_129;
const __VLS_130 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_126.slots.default;
const __VLS_131 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_133 = __VLS_132({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_132));
__VLS_134.slots.default;
const __VLS_135 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    label: (__VLS_ctx.t('casbin_rule.ptype', 'Ptype')),
}));
const __VLS_137 = __VLS_136({
    label: (__VLS_ctx.t('casbin_rule.ptype', 'Ptype')),
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
__VLS_138.slots.default;
const __VLS_139 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
    modelValue: (__VLS_ctx.form.ptype),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'Ptype' })),
}));
const __VLS_141 = __VLS_140({
    modelValue: (__VLS_ctx.form.ptype),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'Ptype' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
var __VLS_138;
const __VLS_143 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    label: (__VLS_ctx.t('casbin_rule.v0', 'V0')),
}));
const __VLS_145 = __VLS_144({
    label: (__VLS_ctx.t('casbin_rule.v0', 'V0')),
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
__VLS_146.slots.default;
const __VLS_147 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
    modelValue: (__VLS_ctx.form.v0),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V0' })),
}));
const __VLS_149 = __VLS_148({
    modelValue: (__VLS_ctx.form.v0),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V0' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
var __VLS_146;
const __VLS_151 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    label: (__VLS_ctx.t('casbin_rule.v1', 'V1')),
}));
const __VLS_153 = __VLS_152({
    label: (__VLS_ctx.t('casbin_rule.v1', 'V1')),
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
const __VLS_155 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    modelValue: (__VLS_ctx.form.v1),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V1' })),
}));
const __VLS_157 = __VLS_156({
    modelValue: (__VLS_ctx.form.v1),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V1' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
var __VLS_154;
const __VLS_159 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    label: (__VLS_ctx.t('casbin_rule.v2', 'V2')),
}));
const __VLS_161 = __VLS_160({
    label: (__VLS_ctx.t('casbin_rule.v2', 'V2')),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    modelValue: (__VLS_ctx.form.v2),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V2' })),
}));
const __VLS_165 = __VLS_164({
    modelValue: (__VLS_ctx.form.v2),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V2' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
var __VLS_162;
const __VLS_167 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('casbin_rule.v3', 'V3')),
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('casbin_rule.v3', 'V3')),
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
const __VLS_171 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    modelValue: (__VLS_ctx.form.v3),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V3' })),
}));
const __VLS_173 = __VLS_172({
    modelValue: (__VLS_ctx.form.v3),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V3' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
var __VLS_170;
const __VLS_175 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    label: (__VLS_ctx.t('casbin_rule.v4', 'V4')),
}));
const __VLS_177 = __VLS_176({
    label: (__VLS_ctx.t('casbin_rule.v4', 'V4')),
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
__VLS_178.slots.default;
const __VLS_179 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    modelValue: (__VLS_ctx.form.v4),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V4' })),
}));
const __VLS_181 = __VLS_180({
    modelValue: (__VLS_ctx.form.v4),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V4' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
var __VLS_178;
const __VLS_183 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    label: (__VLS_ctx.t('casbin_rule.v5', 'V5')),
}));
const __VLS_185 = __VLS_184({
    label: (__VLS_ctx.t('casbin_rule.v5', 'V5')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
const __VLS_187 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    modelValue: (__VLS_ctx.form.v5),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V5' })),
}));
const __VLS_189 = __VLS_188({
    modelValue: (__VLS_ctx.form.v5),
    placeholder: (__VLS_ctx.t('casbin_rule.placeholder', 'Enter {field}', { field: 'V5' })),
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
var __VLS_186;
var __VLS_134;
var __VLS_126;
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
            tableLoading: tableLoading,
            dialogLoading: dialogLoading,
            dialogVisible: dialogVisible,
            rows: rows,
            total: total,
            editingId: editingId,
            t: t,
            query: query,
            form: form,
            loadItems: loadItems,
            openCreate: openCreate,
            openEdit: openEdit,
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
