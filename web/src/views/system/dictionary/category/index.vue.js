import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createDictionaryCategory, deleteDictionaryCategory, fetchDictionaryCategories, updateDictionaryCategory, } from '@/api/dictionary';
import { useAppI18n } from '@/i18n';
import { formatDateTime, statusTagType } from '@/utils/admin';
const { t } = useAppI18n();
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref([]);
const total = ref(0);
const editingId = ref('');
const query = reactive({
    keyword: '',
    status: '',
    page: 1,
    page_size: 10,
});
const defaultForm = () => ({
    code: '',
    name: '',
    description: '',
    status: 'enabled',
    sort: 0,
    remark: '',
});
const form = reactive(defaultForm());
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadCategories() {
    tableLoading.value = true;
    try {
        const response = await fetchDictionaryCategories({ ...query });
        rows.value = response.items;
        total.value = response.total;
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
        ...defaultForm(),
        code: row.code,
        name: row.name,
        description: row.description ?? '',
        status: row.status || 'enabled',
        sort: row.sort ?? 0,
        remark: row.remark ?? '',
    });
    dialogVisible.value = true;
}
function statusLabel(status) {
    return status === 'disabled' ? t('dictionary.category.disabled', 'Disabled') : t('dictionary.category.enabled', 'Enabled');
}
async function submitForm() {
    if (form.code.trim() === '' || form.name.trim() === '') {
        ElMessage.warning(t('dictionary.category.validation_required', 'Enter the dictionary code and name'));
        return;
    }
    dialogLoading.value = true;
    try {
        const payload = {
            ...form,
            code: form.code.trim(),
            name: form.name.trim(),
            description: form.description.trim(),
            status: form.status.trim() || 'enabled',
            sort: Number(form.sort) || 0,
            remark: form.remark.trim(),
        };
        if (editingId.value) {
            await updateDictionaryCategory(editingId.value, payload);
            ElMessage.success(t('dictionary.category.updated', 'Dictionary category updated'));
        }
        else {
            await createDictionaryCategory(payload);
            ElMessage.success(t('dictionary.category.created', 'Dictionary category created'));
        }
        dialogVisible.value = false;
        await loadCategories();
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('dictionary.category.confirm_delete', 'Delete dictionary category {name}?', { name: row.name }), t('dictionary.category.delete_title', 'Delete category'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteDictionaryCategory(row.id);
    ElMessage.success(t('dictionary.category.deleted', 'Dictionary category deleted'));
    await loadCategories();
}
function handleSearch() {
    query.page = 1;
    void loadCategories();
}
function handleReset() {
    query.keyword = '';
    query.status = '';
    query.page = 1;
    void loadCategories();
}
function handlePageChange(page) {
    query.page = page;
    void loadCategories();
}
function handleSizeChange(pageSize) {
    query.page_size = pageSize;
    query.page = 1;
    void loadCategories();
}
onMounted(() => {
    void loadCategories();
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
    title: (__VLS_ctx.t('dictionary.category.title', 'Dictionary categories')),
    description: (__VLS_ctx.t('dictionary.category.description', 'Maintain dictionary category codes, names, and enable/disable status for reuse by other modules.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('dictionary.category.title', 'Dictionary categories')),
    description: (__VLS_ctx.t('dictionary.category.description', 'Maintain dictionary category codes, names, and enable/disable status for reuse by other modules.')),
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
        onClick: (__VLS_ctx.loadCategories)
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:category:create') }, null, null);
    __VLS_14.slots.default;
    (__VLS_ctx.t('dictionary.category.create', 'Add category'));
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
        label: (__VLS_ctx.t('dictionary.category.keyword', 'Keyword')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('dictionary.category.keyword', 'Keyword')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('dictionary.category.keyword_placeholder', 'Code / name / remark')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('dictionary.category.keyword_placeholder', 'Code / name / remark')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
    }));
    const __VLS_33 = __VLS_32({
        label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('dictionary.category.all_status', 'All statuses')),
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('dictionary.category.all_status', 'All statuses')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    const __VLS_39 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: (__VLS_ctx.t('dictionary.category.enabled', 'Enabled')),
        value: "enabled",
    }));
    const __VLS_41 = __VLS_40({
        label: (__VLS_ctx.t('dictionary.category.enabled', 'Enabled')),
        value: "enabled",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    const __VLS_43 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: (__VLS_ctx.t('dictionary.category.disabled', 'Disabled')),
        value: "disabled",
    }));
    const __VLS_45 = __VLS_44({
        label: (__VLS_ctx.t('dictionary.category.disabled', 'Disabled')),
        value: "disabled",
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
    prop: "code",
    label: (__VLS_ctx.t('dictionary.category.code', 'Category code')),
    minWidth: "160",
}));
const __VLS_73 = __VLS_72({
    prop: "code",
    label: (__VLS_ctx.t('dictionary.category.code', 'Category code')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "name",
    label: (__VLS_ctx.t('dictionary.category.name', 'Category name')),
    minWidth: "160",
}));
const __VLS_77 = __VLS_76({
    prop: "name",
    label: (__VLS_ctx.t('dictionary.category.name', 'Category name')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "description",
    label: (__VLS_ctx.t('dictionary.category.description_label', 'Description')),
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_81 = __VLS_80({
    prop: "description",
    label: (__VLS_ctx.t('dictionary.category.description_label', 'Description')),
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
    width: "100",
}));
const __VLS_85 = __VLS_84({
    label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
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
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }));
    const __VLS_89 = __VLS_88({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_88));
    __VLS_90.slots.default;
    (__VLS_ctx.statusLabel(row.status));
    var __VLS_90;
}
var __VLS_86;
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    prop: "sort",
    label: (__VLS_ctx.t('dictionary.category.sort', 'Sort')),
    width: "90",
}));
const __VLS_93 = __VLS_92({
    prop: "sort",
    label: (__VLS_ctx.t('dictionary.category.sort', 'Sort')),
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    prop: "remark",
    label: (__VLS_ctx.t('dictionary.category.remark', 'Remark')),
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_97 = __VLS_96({
    prop: "remark",
    label: (__VLS_ctx.t('dictionary.category.remark', 'Remark')),
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
const __VLS_99 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    label: (__VLS_ctx.t('dictionary.category.updated_at', 'Updated at')),
    minWidth: "180",
}));
const __VLS_101 = __VLS_100({
    label: (__VLS_ctx.t('dictionary.category.updated_at', 'Updated at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_102.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_102.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_102;
const __VLS_103 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_105 = __VLS_104({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
__VLS_106.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_106.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_107 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_109 = __VLS_108({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_108));
    let __VLS_111;
    let __VLS_112;
    let __VLS_113;
    const __VLS_114 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:category:update') }, null, null);
    __VLS_110.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_110;
    const __VLS_115 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_117 = __VLS_116({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_116));
    let __VLS_119;
    let __VLS_120;
    let __VLS_121;
    const __VLS_122 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:category:delete') }, null, null);
    __VLS_118.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_118;
}
var __VLS_106;
var __VLS_70;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_123 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_125 = __VLS_124({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_124));
    let __VLS_127;
    let __VLS_128;
    let __VLS_129;
    const __VLS_130 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_131 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_126;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_132 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('dictionary.category.edit_title', 'Edit dictionary category') : __VLS_ctx.t('dictionary.category.create_title', 'New dictionary category')),
    loading: (__VLS_ctx.dialogLoading),
    width: "720px",
}));
const __VLS_133 = __VLS_132({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('dictionary.category.edit_title', 'Edit dictionary category') : __VLS_ctx.t('dictionary.category.create_title', 'New dictionary category')),
    loading: (__VLS_ctx.dialogLoading),
    width: "720px",
}, ...__VLS_functionalComponentArgsRest(__VLS_132));
let __VLS_135;
let __VLS_136;
let __VLS_137;
const __VLS_138 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_134.slots.default;
const __VLS_139 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_141 = __VLS_140({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
__VLS_142.slots.default;
const __VLS_143 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_144 = __VLS_asFunctionalComponent(__VLS_143, new __VLS_143({
    label: (__VLS_ctx.t('dictionary.category.code', 'Category code')),
    required: true,
}));
const __VLS_145 = __VLS_144({
    label: (__VLS_ctx.t('dictionary.category.code', 'Category code')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_144));
__VLS_146.slots.default;
const __VLS_147 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
    modelValue: (__VLS_ctx.form.code),
    placeholder: (__VLS_ctx.t('dictionary.category.code_placeholder', 'Enter category code')),
}));
const __VLS_149 = __VLS_148({
    modelValue: (__VLS_ctx.form.code),
    placeholder: (__VLS_ctx.t('dictionary.category.code_placeholder', 'Enter category code')),
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
var __VLS_146;
const __VLS_151 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    label: (__VLS_ctx.t('dictionary.category.name', 'Category name')),
    required: true,
}));
const __VLS_153 = __VLS_152({
    label: (__VLS_ctx.t('dictionary.category.name', 'Category name')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
const __VLS_155 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('dictionary.category.name_placeholder', 'Enter category name')),
}));
const __VLS_157 = __VLS_156({
    modelValue: (__VLS_ctx.form.name),
    placeholder: (__VLS_ctx.t('dictionary.category.name_placeholder', 'Enter category name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
var __VLS_154;
const __VLS_159 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    label: (__VLS_ctx.t('dictionary.category.description_label', 'Description')),
}));
const __VLS_161 = __VLS_160({
    label: (__VLS_ctx.t('dictionary.category.description_label', 'Description')),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('dictionary.category.description_placeholder', 'Enter description')),
}));
const __VLS_165 = __VLS_164({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('dictionary.category.description_placeholder', 'Enter description')),
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
var __VLS_162;
const __VLS_167 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('dictionary.category.status', 'Status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
const __VLS_171 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}));
const __VLS_173 = __VLS_172({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
__VLS_174.slots.default;
const __VLS_175 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    label: (__VLS_ctx.t('dictionary.category.enabled', 'Enabled')),
    value: "enabled",
}));
const __VLS_177 = __VLS_176({
    label: (__VLS_ctx.t('dictionary.category.enabled', 'Enabled')),
    value: "enabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
const __VLS_179 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    label: (__VLS_ctx.t('dictionary.category.disabled', 'Disabled')),
    value: "disabled",
}));
const __VLS_181 = __VLS_180({
    label: (__VLS_ctx.t('dictionary.category.disabled', 'Disabled')),
    value: "disabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
var __VLS_174;
var __VLS_170;
const __VLS_183 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    label: (__VLS_ctx.t('dictionary.category.sort', 'Sort')),
}));
const __VLS_185 = __VLS_184({
    label: (__VLS_ctx.t('dictionary.category.sort', 'Sort')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
const __VLS_187 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}));
const __VLS_189 = __VLS_188({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
var __VLS_186;
const __VLS_191 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    label: (__VLS_ctx.t('dictionary.category.remark', 'Remark')),
}));
const __VLS_193 = __VLS_192({
    label: (__VLS_ctx.t('dictionary.category.remark', 'Remark')),
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
__VLS_194.slots.default;
const __VLS_195 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('dictionary.category.remark_placeholder', 'Enter remark')),
}));
const __VLS_197 = __VLS_196({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: (__VLS_ctx.t('dictionary.category.remark_placeholder', 'Enter remark')),
}, ...__VLS_functionalComponentArgsRest(__VLS_196));
var __VLS_194;
var __VLS_142;
var __VLS_134;
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
            t: t,
            tableLoading: tableLoading,
            dialogLoading: dialogLoading,
            dialogVisible: dialogVisible,
            rows: rows,
            total: total,
            editingId: editingId,
            query: query,
            form: form,
            loadCategories: loadCategories,
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
