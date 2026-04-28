import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createCasbinModel, deleteCasbinModel, listcasbin_models, updateCasbinModel } from '@/api/casbin_model';
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
    content: '',
});
const form = reactive(defaultForm());
function getRowKey(row) {
    return row.id || row.name || '';
}
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
        const response = await listcasbin_models({ ...query });
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
    editingId.value = getRowKey(row);
    Object.assign(form, {
        content: row.content ?? '',
    });
    dialogVisible.value = true;
}
async function submitForm() {
    dialogLoading.value = true;
    try {
        const payload = {
            content: form.content.trim(),
        };
        if (editingId.value) {
            await updateCasbinModel(editingId.value, payload);
            ElMessage.success(t('casbin_model.updated', 'CasbinModel updated'));
        }
        else {
            await createCasbinModel(payload);
            ElMessage.success(t('casbin_model.created', 'CasbinModel created'));
        }
        dialogVisible.value = false;
        await loadItems();
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('casbin_model.save_failed', 'Save failed'));
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    const rowKey = getRowKey(row);
    await ElMessageBox.confirm(t('casbin_model.confirm_delete', 'Delete CasbinModel {name}?', { name: rowKey }), t('casbin_model.delete_title', 'Delete CasbinModel'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteCasbinModel(rowKey);
    ElMessage.success(t('casbin_model.deleted', 'CasbinModel deleted'));
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
    title: (__VLS_ctx.t('casbin_model.title', 'Model management')),
    description: (__VLS_ctx.t('casbin_model.description', 'Manage authorization model configuration, edit entries, and delete entries.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('casbin_model.title', 'Model management')),
    description: (__VLS_ctx.t('casbin_model.description', 'Manage authorization model configuration, edit entries, and delete entries.')),
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_model:create') }, null, null);
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
        placeholder: (__VLS_ctx.t('casbin_model.keyword_placeholder', 'Search CasbinModel data')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('casbin_model.keyword_placeholder', 'Search CasbinModel data')),
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
    rowKey: (__VLS_ctx.getRowKey),
}));
const __VLS_53 = __VLS_52({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: (__VLS_ctx.getRowKey),
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_54.slots.default;
const __VLS_55 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
    label: (__VLS_ctx.t('casbin_model.id', 'ID')),
    minWidth: "160",
}));
const __VLS_57 = __VLS_56({
    label: (__VLS_ctx.t('casbin_model.id', 'ID')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
__VLS_58.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_58.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.getRowKey(row) || '-');
}
var __VLS_58;
const __VLS_59 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    prop: "content",
    label: (__VLS_ctx.t('casbin_model.content', 'Content')),
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_61 = __VLS_60({
    prop: "content",
    label: (__VLS_ctx.t('casbin_model.content', 'Content')),
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_62.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.content || '-');
}
var __VLS_62;
const __VLS_63 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    label: (__VLS_ctx.t('casbin_model.created_at', 'Created at')),
    minWidth: "180",
}));
const __VLS_65 = __VLS_64({
    label: (__VLS_ctx.t('casbin_model.created_at', 'Created at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_66.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_66;
const __VLS_67 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    label: (__VLS_ctx.t('casbin_model.updated_at', 'Updated at')),
    minWidth: "180",
}));
const __VLS_69 = __VLS_68({
    label: (__VLS_ctx.t('casbin_model.updated_at', 'Updated at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_70.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_70;
const __VLS_71 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    label: (__VLS_ctx.t('casbin_model.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_73 = __VLS_72({
    label: (__VLS_ctx.t('casbin_model.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_74.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_75 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_77 = __VLS_76({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_76));
    let __VLS_79;
    let __VLS_80;
    let __VLS_81;
    const __VLS_82 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_model:update') }, null, null);
    __VLS_78.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_78;
    const __VLS_83 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_85 = __VLS_84({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_84));
    let __VLS_87;
    let __VLS_88;
    let __VLS_89;
    const __VLS_90 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('casbin_model:delete') }, null, null);
    __VLS_86.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_86;
}
var __VLS_74;
var __VLS_54;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_91 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_93 = __VLS_92({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_92));
    let __VLS_95;
    let __VLS_96;
    let __VLS_97;
    const __VLS_98 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_99 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_94;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('casbin_model.edit_title', 'Edit model') : __VLS_ctx.t('casbin_model.create_title', 'New model')),
    loading: (__VLS_ctx.dialogLoading),
}));
const __VLS_101 = __VLS_100({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('casbin_model.edit_title', 'Edit model') : __VLS_ctx.t('casbin_model.create_title', 'New model')),
    loading: (__VLS_ctx.dialogLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
let __VLS_103;
let __VLS_104;
let __VLS_105;
const __VLS_106 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_102.slots.default;
const __VLS_107 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_109 = __VLS_108({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
const __VLS_111 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('casbin_model.content', 'Content')),
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('casbin_model.content', 'Content')),
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
__VLS_114.slots.default;
const __VLS_115 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
    modelValue: (__VLS_ctx.form.content),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('casbin_model.content_placeholder', 'Enter content')),
}));
const __VLS_117 = __VLS_116({
    modelValue: (__VLS_ctx.form.content),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('casbin_model.content_placeholder', 'Enter content')),
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
var __VLS_114;
var __VLS_110;
var __VLS_102;
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
            getRowKey: getRowKey,
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
