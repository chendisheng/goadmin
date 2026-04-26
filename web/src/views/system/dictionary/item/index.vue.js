import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createDictionaryItem, deleteDictionaryItem, fetchDictionaryCategories, fetchDictionaryItem, fetchDictionaryItems, fetchDictionaryLookupItem, fetchDictionaryLookupItems, updateDictionaryItem, } from '@/api/dictionary';
import { formatDateTime, statusTagType } from '@/utils/admin';
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const categoryLoading = ref(false);
const lookupLoading = ref(false);
const rows = ref([]);
const total = ref(0);
const categoryOptions = ref([]);
const editingId = ref('');
const lookupResult = ref(null);
const lookupItems = ref([]);
const query = reactive({
    category_id: '',
    category_code: '',
    keyword: '',
    status: '',
    page: 1,
    page_size: 10,
});
const lookupForm = reactive({
    category_code: '',
    value: '',
});
const defaultForm = () => ({
    category_id: '',
    value: '',
    label: '',
    tag_type: '',
    tag_color: '',
    extra: '',
    is_default: false,
    status: 'enabled',
    sort: 0,
    remark: '',
});
const form = reactive(defaultForm());
const categoryMap = computed(() => {
    const map = new Map();
    for (const item of categoryOptions.value) {
        map.set(item.id, item);
    }
    return map;
});
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadCategories() {
    categoryLoading.value = true;
    try {
        const response = await fetchDictionaryCategories({ keyword: '', status: '', page: 1, page_size: 200 });
        categoryOptions.value = response.items;
    }
    finally {
        categoryLoading.value = false;
    }
}
async function loadItems() {
    tableLoading.value = true;
    try {
        const response = await fetchDictionaryItems({ ...query });
        rows.value = response.items;
        total.value = response.total;
    }
    finally {
        tableLoading.value = false;
    }
}
function categoryLabel(categoryId) {
    const category = categoryMap.value.get(categoryId);
    if (!category) {
        return categoryId || '-';
    }
    return `${category.name} (${category.code})`;
}
function openCreate() {
    editingId.value = '';
    resetForm();
    if (query.category_id) {
        form.category_id = query.category_id;
    }
    dialogVisible.value = true;
}
async function openEdit(row) {
    editingId.value = row.id;
    const detail = await fetchDictionaryItem(row.id);
    Object.assign(form, {
        ...defaultForm(),
        category_id: detail.category_id,
        value: detail.value,
        label: detail.label,
        tag_type: detail.tag_type ?? '',
        tag_color: detail.tag_color ?? '',
        extra: detail.extra ?? '',
        is_default: detail.is_default,
        status: detail.status || 'enabled',
        sort: detail.sort ?? 0,
        remark: detail.remark ?? '',
    });
    dialogVisible.value = true;
}
function statusLabel(status) {
    return status === 'disabled' ? '禁用' : '启用';
}
function defaultLabel(value) {
    return value ? '是' : '否';
}
async function submitForm() {
    if (form.category_id.trim() === '' || form.value.trim() === '' || form.label.trim() === '') {
        ElMessage.warning('请输入分类、项值和标签');
        return;
    }
    dialogLoading.value = true;
    try {
        const payload = {
            ...form,
            category_id: form.category_id.trim(),
            value: form.value.trim(),
            label: form.label.trim(),
            tag_type: form.tag_type.trim(),
            tag_color: form.tag_color.trim(),
            extra: form.extra.trim(),
            is_default: Boolean(form.is_default),
            status: form.status.trim() || 'enabled',
            sort: Number(form.sort) || 0,
            remark: form.remark.trim(),
        };
        if (editingId.value) {
            await updateDictionaryItem(editingId.value, payload);
            ElMessage.success('字典项已更新');
        }
        else {
            await createDictionaryItem(payload);
            ElMessage.success('字典项已创建');
        }
        dialogVisible.value = false;
        await loadItems();
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(`确认删除字典项 ${row.label} / ${row.value} 吗？`, '删除字典项', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
    });
    await deleteDictionaryItem(row.id);
    ElMessage.success('字典项已删除');
    await loadItems();
}
function handleSearch() {
    query.page = 1;
    void loadItems();
}
function handleReset() {
    query.category_id = '';
    query.category_code = '';
    query.keyword = '';
    query.status = '';
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
async function runLookupList() {
    if (lookupForm.category_code.trim() === '') {
        ElMessage.warning('请输入分类编码');
        return;
    }
    lookupLoading.value = true;
    try {
        const response = await fetchDictionaryLookupItems(lookupForm.category_code.trim());
        lookupItems.value = response.items;
        lookupResult.value = response.items[0] ?? null;
    }
    finally {
        lookupLoading.value = false;
    }
}
async function runLookupItem() {
    if (lookupForm.category_code.trim() === '' || lookupForm.value.trim() === '') {
        ElMessage.warning('请输入分类编码和项值');
        return;
    }
    lookupLoading.value = true;
    try {
        lookupResult.value = await fetchDictionaryLookupItem(lookupForm.category_code.trim(), lookupForm.value.trim());
        lookupItems.value = lookupResult.value ? [lookupResult.value] : [];
    }
    finally {
        lookupLoading.value = false;
    }
}
onMounted(async () => {
    await loadCategories();
    await loadItems();
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "admin-page" },
});
const __VLS_0 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    gutter: (20),
}));
const __VLS_2 = __VLS_1({
    gutter: (20),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
const __VLS_4 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
    xs: (24),
    xl: (16),
}));
const __VLS_6 = __VLS_5({
    xs: (24),
    xl: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_5));
__VLS_7.slots.default;
/** @type {[typeof AdminTable, typeof AdminTable, ]} */ ;
// @ts-ignore
const __VLS_8 = __VLS_asFunctionalComponent(AdminTable, new AdminTable({
    title: "字典项管理",
    description: "维护字典项值、标签、默认项和启停状态，并支持按分类快速筛选。",
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_9 = __VLS_8({
    title: "字典项管理",
    description: "维护字典项值、标签、默认项和启停状态，并支持按分类快速筛选。",
    loading: (__VLS_ctx.tableLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_8));
__VLS_10.slots.default;
{
    const { actions: __VLS_thisSlot } = __VLS_10.slots;
    const __VLS_11 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }));
    const __VLS_13 = __VLS_12({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.tableLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    let __VLS_15;
    let __VLS_16;
    let __VLS_17;
    const __VLS_18 = {
        onClick: (__VLS_ctx.loadItems)
    };
    __VLS_14.slots.default;
    var __VLS_14;
    const __VLS_19 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (__VLS_ctx.openCreate)
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:item:create') }, null, null);
    __VLS_22.slots.default;
    var __VLS_22;
}
{
    const { filters: __VLS_thisSlot } = __VLS_10.slots;
    const __VLS_27 = {}.ElForm;
    /** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }));
    const __VLS_29 = __VLS_28({
        inline: (true),
        labelWidth: "88px",
        ...{ class: "admin-filters" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    __VLS_30.slots.default;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: "分类",
    }));
    const __VLS_33 = __VLS_32({
        label: "分类",
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.category_id),
        clearable: true,
        filterable: true,
        placeholder: "全部分类",
        ...{ style: {} },
        loading: (__VLS_ctx.categoryLoading),
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.category_id),
        clearable: true,
        filterable: true,
        placeholder: "全部分类",
        ...{ style: {} },
        loading: (__VLS_ctx.categoryLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    for (const [item] of __VLS_getVForSourceType((__VLS_ctx.categoryOptions))) {
        const __VLS_39 = {}.ElOption;
        /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
        // @ts-ignore
        const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
            key: (item.id),
            label: (`${item.name} (${item.code})`),
            value: (item.id),
        }));
        const __VLS_41 = __VLS_40({
            key: (item.id),
            label: (`${item.name} (${item.code})`),
            value: (item.id),
        }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    }
    var __VLS_38;
    var __VLS_34;
    const __VLS_43 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: "关键字",
    }));
    const __VLS_45 = __VLS_44({
        label: "关键字",
    }, ...__VLS_functionalComponentArgsRest(__VLS_44));
    __VLS_46.slots.default;
    const __VLS_47 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: "项值 / 标签 / 备注",
    }));
    const __VLS_49 = __VLS_48({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: "项值 / 标签 / 备注",
    }, ...__VLS_functionalComponentArgsRest(__VLS_48));
    var __VLS_46;
    const __VLS_51 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
        label: "状态",
    }));
    const __VLS_53 = __VLS_52({
        label: "状态",
    }, ...__VLS_functionalComponentArgsRest(__VLS_52));
    __VLS_54.slots.default;
    const __VLS_55 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: "全部状态",
        ...{ style: {} },
    }));
    const __VLS_57 = __VLS_56({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: "全部状态",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_56));
    __VLS_58.slots.default;
    const __VLS_59 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
        label: "启用",
        value: "enabled",
    }));
    const __VLS_61 = __VLS_60({
        label: "启用",
        value: "enabled",
    }, ...__VLS_functionalComponentArgsRest(__VLS_60));
    const __VLS_63 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
        label: "禁用",
        value: "disabled",
    }));
    const __VLS_65 = __VLS_64({
        label: "禁用",
        value: "disabled",
    }, ...__VLS_functionalComponentArgsRest(__VLS_64));
    var __VLS_58;
    var __VLS_54;
    const __VLS_67 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({}));
    const __VLS_69 = __VLS_68({}, ...__VLS_functionalComponentArgsRest(__VLS_68));
    __VLS_70.slots.default;
    const __VLS_71 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_73 = __VLS_72({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_72));
    let __VLS_75;
    let __VLS_76;
    let __VLS_77;
    const __VLS_78 = {
        onClick: (__VLS_ctx.handleSearch)
    };
    __VLS_74.slots.default;
    var __VLS_74;
    const __VLS_79 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
        ...{ 'onClick': {} },
    }));
    const __VLS_81 = __VLS_80({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_80));
    let __VLS_83;
    let __VLS_84;
    let __VLS_85;
    const __VLS_86 = {
        onClick: (__VLS_ctx.handleReset)
    };
    __VLS_82.slots.default;
    var __VLS_82;
    var __VLS_70;
    var __VLS_30;
}
const __VLS_87 = {}.ElTable;
/** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}));
const __VLS_89 = __VLS_88({
    data: (__VLS_ctx.rows),
    border: true,
    rowKey: "id",
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_asFunctionalDirective(__VLS_directives.vLoading)(null, { ...__VLS_directiveBindingRestFields, value: (__VLS_ctx.tableLoading) }, null, null);
__VLS_90.slots.default;
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    label: "分类",
    minWidth: "200",
    showOverflowTooltip: true,
}));
const __VLS_93 = __VLS_92({
    label: "分类",
    minWidth: "200",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_94.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.categoryLabel(row.category_id));
}
var __VLS_94;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    prop: "value",
    label: "项值",
    minWidth: "160",
}));
const __VLS_97 = __VLS_96({
    prop: "value",
    label: "项值",
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
const __VLS_99 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    prop: "label",
    label: "标签",
    minWidth: "160",
}));
const __VLS_101 = __VLS_100({
    prop: "label",
    label: "标签",
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
const __VLS_103 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    prop: "tag_type",
    label: "标签类型",
    width: "120",
}));
const __VLS_105 = __VLS_104({
    prop: "tag_type",
    label: "标签类型",
    width: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
const __VLS_107 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    label: "默认",
    width: "90",
}));
const __VLS_109 = __VLS_108({
    label: "默认",
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_110.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_111 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
        type: (row.is_default ? 'success' : 'info'),
        effect: "plain",
    }));
    const __VLS_113 = __VLS_112({
        type: (row.is_default ? 'success' : 'info'),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_112));
    __VLS_114.slots.default;
    (__VLS_ctx.defaultLabel(row.is_default));
    var __VLS_114;
}
var __VLS_110;
const __VLS_115 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
    label: "状态",
    width: "100",
}));
const __VLS_117 = __VLS_116({
    label: "状态",
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
__VLS_118.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_118.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_119 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }));
    const __VLS_121 = __VLS_120({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_120));
    __VLS_122.slots.default;
    (__VLS_ctx.statusLabel(row.status));
    var __VLS_122;
}
var __VLS_118;
const __VLS_123 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
    prop: "sort",
    label: "排序",
    width: "90",
}));
const __VLS_125 = __VLS_124({
    prop: "sort",
    label: "排序",
    width: "90",
}, ...__VLS_functionalComponentArgsRest(__VLS_124));
const __VLS_127 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_128 = __VLS_asFunctionalComponent(__VLS_127, new __VLS_127({
    prop: "remark",
    label: "备注",
    minWidth: "180",
    showOverflowTooltip: true,
}));
const __VLS_129 = __VLS_128({
    prop: "remark",
    label: "备注",
    minWidth: "180",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_128));
const __VLS_131 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
    label: "更新时间",
    minWidth: "180",
}));
const __VLS_133 = __VLS_132({
    label: "更新时间",
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_132));
__VLS_134.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_134.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_134;
const __VLS_135 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_136 = __VLS_asFunctionalComponent(__VLS_135, new __VLS_135({
    label: "操作",
    width: "180",
    fixed: "right",
}));
const __VLS_137 = __VLS_136({
    label: "操作",
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_136));
__VLS_138.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_138.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_139 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_141 = __VLS_140({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_140));
    let __VLS_143;
    let __VLS_144;
    let __VLS_145;
    const __VLS_146 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:item:update') }, null, null);
    __VLS_142.slots.default;
    var __VLS_142;
    const __VLS_147 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_149 = __VLS_148({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_148));
    let __VLS_151;
    let __VLS_152;
    let __VLS_153;
    const __VLS_154 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('dictionary:item:delete') }, null, null);
    __VLS_150.slots.default;
    var __VLS_150;
}
var __VLS_138;
var __VLS_90;
{
    const { footer: __VLS_thisSlot } = __VLS_10.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_155 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_157 = __VLS_156({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_156));
    let __VLS_159;
    let __VLS_160;
    let __VLS_161;
    const __VLS_162 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_163 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_158;
}
var __VLS_10;
var __VLS_7;
const __VLS_164 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_165 = __VLS_asFunctionalComponent(__VLS_164, new __VLS_164({
    xs: (24),
    xl: (8),
}));
const __VLS_166 = __VLS_165({
    xs: (24),
    xl: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_165));
__VLS_167.slots.default;
const __VLS_168 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_169 = __VLS_asFunctionalComponent(__VLS_168, new __VLS_168({
    direction: "vertical",
    fill: true,
    size: (16),
    ...{ style: {} },
}));
const __VLS_170 = __VLS_169({
    direction: "vertical",
    fill: true,
    size: (16),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_169));
__VLS_171.slots.default;
const __VLS_172 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_173 = __VLS_asFunctionalComponent(__VLS_172, new __VLS_172({
    shadow: "never",
}));
const __VLS_174 = __VLS_173({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_173));
__VLS_175.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_175.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
}
const __VLS_176 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_177 = __VLS_asFunctionalComponent(__VLS_176, new __VLS_176({
    labelWidth: "96px",
}));
const __VLS_178 = __VLS_177({
    labelWidth: "96px",
}, ...__VLS_functionalComponentArgsRest(__VLS_177));
__VLS_179.slots.default;
const __VLS_180 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_181 = __VLS_asFunctionalComponent(__VLS_180, new __VLS_180({
    label: "分类编码",
}));
const __VLS_182 = __VLS_181({
    label: "分类编码",
}, ...__VLS_functionalComponentArgsRest(__VLS_181));
__VLS_183.slots.default;
const __VLS_184 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_185 = __VLS_asFunctionalComponent(__VLS_184, new __VLS_184({
    modelValue: (__VLS_ctx.lookupForm.category_code),
    placeholder: "例如 system_status",
}));
const __VLS_186 = __VLS_185({
    modelValue: (__VLS_ctx.lookupForm.category_code),
    placeholder: "例如 system_status",
}, ...__VLS_functionalComponentArgsRest(__VLS_185));
var __VLS_183;
const __VLS_188 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_189 = __VLS_asFunctionalComponent(__VLS_188, new __VLS_188({
    label: "项值",
}));
const __VLS_190 = __VLS_189({
    label: "项值",
}, ...__VLS_functionalComponentArgsRest(__VLS_189));
__VLS_191.slots.default;
const __VLS_192 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_193 = __VLS_asFunctionalComponent(__VLS_192, new __VLS_192({
    modelValue: (__VLS_ctx.lookupForm.value),
    placeholder: "仅查单项时填写",
}));
const __VLS_194 = __VLS_193({
    modelValue: (__VLS_ctx.lookupForm.value),
    placeholder: "仅查单项时填写",
}, ...__VLS_functionalComponentArgsRest(__VLS_193));
var __VLS_191;
const __VLS_196 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_197 = __VLS_asFunctionalComponent(__VLS_196, new __VLS_196({}));
const __VLS_198 = __VLS_197({}, ...__VLS_functionalComponentArgsRest(__VLS_197));
__VLS_199.slots.default;
const __VLS_200 = {}.ElSpace;
/** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
// @ts-ignore
const __VLS_201 = __VLS_asFunctionalComponent(__VLS_200, new __VLS_200({}));
const __VLS_202 = __VLS_201({}, ...__VLS_functionalComponentArgsRest(__VLS_201));
__VLS_203.slots.default;
const __VLS_204 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_205 = __VLS_asFunctionalComponent(__VLS_204, new __VLS_204({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.lookupLoading),
    type: "primary",
}));
const __VLS_206 = __VLS_205({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.lookupLoading),
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_205));
let __VLS_208;
let __VLS_209;
let __VLS_210;
const __VLS_211 = {
    onClick: (__VLS_ctx.runLookupList)
};
__VLS_207.slots.default;
var __VLS_207;
const __VLS_212 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_213 = __VLS_asFunctionalComponent(__VLS_212, new __VLS_212({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.lookupLoading),
}));
const __VLS_214 = __VLS_213({
    ...{ 'onClick': {} },
    loading: (__VLS_ctx.lookupLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_213));
let __VLS_216;
let __VLS_217;
let __VLS_218;
const __VLS_219 = {
    onClick: (__VLS_ctx.runLookupItem)
};
__VLS_215.slots.default;
var __VLS_215;
var __VLS_203;
var __VLS_199;
var __VLS_179;
var __VLS_175;
const __VLS_220 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_221 = __VLS_asFunctionalComponent(__VLS_220, new __VLS_220({
    shadow: "never",
}));
const __VLS_222 = __VLS_221({
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_221));
__VLS_223.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_223.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
}
if (!__VLS_ctx.lookupItems.length) {
    const __VLS_224 = {}.ElEmpty;
    /** @type {[typeof __VLS_components.ElEmpty, typeof __VLS_components.elEmpty, ]} */ ;
    // @ts-ignore
    const __VLS_225 = __VLS_asFunctionalComponent(__VLS_224, new __VLS_224({
        description: "暂无结果",
    }));
    const __VLS_226 = __VLS_225({
        description: "暂无结果",
    }, ...__VLS_functionalComponentArgsRest(__VLS_225));
}
else {
    const __VLS_228 = {}.ElTable;
    /** @type {[typeof __VLS_components.ElTable, typeof __VLS_components.elTable, typeof __VLS_components.ElTable, typeof __VLS_components.elTable, ]} */ ;
    // @ts-ignore
    const __VLS_229 = __VLS_asFunctionalComponent(__VLS_228, new __VLS_228({
        data: (__VLS_ctx.lookupItems),
        size: "small",
        border: true,
    }));
    const __VLS_230 = __VLS_229({
        data: (__VLS_ctx.lookupItems),
        size: "small",
        border: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_229));
    __VLS_231.slots.default;
    const __VLS_232 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_233 = __VLS_asFunctionalComponent(__VLS_232, new __VLS_232({
        prop: "value",
        label: "项值",
        minWidth: "110",
    }));
    const __VLS_234 = __VLS_233({
        prop: "value",
        label: "项值",
        minWidth: "110",
    }, ...__VLS_functionalComponentArgsRest(__VLS_233));
    const __VLS_236 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_237 = __VLS_asFunctionalComponent(__VLS_236, new __VLS_236({
        prop: "label",
        label: "标签",
        minWidth: "120",
    }));
    const __VLS_238 = __VLS_237({
        prop: "label",
        label: "标签",
        minWidth: "120",
    }, ...__VLS_functionalComponentArgsRest(__VLS_237));
    const __VLS_240 = {}.ElTableColumn;
    /** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
    // @ts-ignore
    const __VLS_241 = __VLS_asFunctionalComponent(__VLS_240, new __VLS_240({
        prop: "status",
        label: "状态",
        width: "90",
    }));
    const __VLS_242 = __VLS_241({
        prop: "status",
        label: "状态",
        width: "90",
    }, ...__VLS_functionalComponentArgsRest(__VLS_241));
    var __VLS_231;
}
if (__VLS_ctx.lookupResult) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "lookup-result-card" },
    });
    const __VLS_244 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_245 = __VLS_asFunctionalComponent(__VLS_244, new __VLS_244({}));
    const __VLS_246 = __VLS_245({}, ...__VLS_functionalComponentArgsRest(__VLS_245));
    __VLS_247.slots.default;
    var __VLS_247;
    const __VLS_248 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_249 = __VLS_asFunctionalComponent(__VLS_248, new __VLS_248({
        column: (1),
        border: true,
        size: "small",
    }));
    const __VLS_250 = __VLS_249({
        column: (1),
        border: true,
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_249));
    __VLS_251.slots.default;
    const __VLS_252 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_253 = __VLS_asFunctionalComponent(__VLS_252, new __VLS_252({
        label: "项值",
    }));
    const __VLS_254 = __VLS_253({
        label: "项值",
    }, ...__VLS_functionalComponentArgsRest(__VLS_253));
    __VLS_255.slots.default;
    (__VLS_ctx.lookupResult.value);
    var __VLS_255;
    const __VLS_256 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_257 = __VLS_asFunctionalComponent(__VLS_256, new __VLS_256({
        label: "标签",
    }));
    const __VLS_258 = __VLS_257({
        label: "标签",
    }, ...__VLS_functionalComponentArgsRest(__VLS_257));
    __VLS_259.slots.default;
    (__VLS_ctx.lookupResult.label);
    var __VLS_259;
    const __VLS_260 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_261 = __VLS_asFunctionalComponent(__VLS_260, new __VLS_260({
        label: "默认",
    }));
    const __VLS_262 = __VLS_261({
        label: "默认",
    }, ...__VLS_functionalComponentArgsRest(__VLS_261));
    __VLS_263.slots.default;
    (__VLS_ctx.defaultLabel(__VLS_ctx.lookupResult.is_default));
    var __VLS_263;
    const __VLS_264 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_265 = __VLS_asFunctionalComponent(__VLS_264, new __VLS_264({
        label: "状态",
    }));
    const __VLS_266 = __VLS_265({
        label: "状态",
    }, ...__VLS_functionalComponentArgsRest(__VLS_265));
    __VLS_267.slots.default;
    (__VLS_ctx.lookupResult.status);
    var __VLS_267;
    var __VLS_251;
}
var __VLS_223;
var __VLS_171;
var __VLS_167;
var __VLS_3;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_268 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? '编辑字典项' : '新增字典项'),
    loading: (__VLS_ctx.dialogLoading),
    width: "760px",
}));
const __VLS_269 = __VLS_268({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? '编辑字典项' : '新增字典项'),
    loading: (__VLS_ctx.dialogLoading),
    width: "760px",
}, ...__VLS_functionalComponentArgsRest(__VLS_268));
let __VLS_271;
let __VLS_272;
let __VLS_273;
const __VLS_274 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_270.slots.default;
const __VLS_275 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_276 = __VLS_asFunctionalComponent(__VLS_275, new __VLS_275({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_277 = __VLS_276({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_276));
__VLS_278.slots.default;
const __VLS_279 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_280 = __VLS_asFunctionalComponent(__VLS_279, new __VLS_279({
    label: "分类",
    required: true,
}));
const __VLS_281 = __VLS_280({
    label: "分类",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_280));
__VLS_282.slots.default;
const __VLS_283 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_284 = __VLS_asFunctionalComponent(__VLS_283, new __VLS_283({
    modelValue: (__VLS_ctx.form.category_id),
    filterable: true,
    placeholder: "请选择分类",
    ...{ style: {} },
    loading: (__VLS_ctx.categoryLoading),
}));
const __VLS_285 = __VLS_284({
    modelValue: (__VLS_ctx.form.category_id),
    filterable: true,
    placeholder: "请选择分类",
    ...{ style: {} },
    loading: (__VLS_ctx.categoryLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_284));
__VLS_286.slots.default;
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.categoryOptions))) {
    const __VLS_287 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_288 = __VLS_asFunctionalComponent(__VLS_287, new __VLS_287({
        key: (item.id),
        label: (`${item.name} (${item.code})`),
        value: (item.id),
    }));
    const __VLS_289 = __VLS_288({
        key: (item.id),
        label: (`${item.name} (${item.code})`),
        value: (item.id),
    }, ...__VLS_functionalComponentArgsRest(__VLS_288));
}
var __VLS_286;
var __VLS_282;
const __VLS_291 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_292 = __VLS_asFunctionalComponent(__VLS_291, new __VLS_291({
    label: "项值",
    required: true,
}));
const __VLS_293 = __VLS_292({
    label: "项值",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_292));
__VLS_294.slots.default;
const __VLS_295 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_296 = __VLS_asFunctionalComponent(__VLS_295, new __VLS_295({
    modelValue: (__VLS_ctx.form.value),
    placeholder: "请输入项值",
}));
const __VLS_297 = __VLS_296({
    modelValue: (__VLS_ctx.form.value),
    placeholder: "请输入项值",
}, ...__VLS_functionalComponentArgsRest(__VLS_296));
var __VLS_294;
const __VLS_299 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_300 = __VLS_asFunctionalComponent(__VLS_299, new __VLS_299({
    label: "标签",
    required: true,
}));
const __VLS_301 = __VLS_300({
    label: "标签",
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_300));
__VLS_302.slots.default;
const __VLS_303 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_304 = __VLS_asFunctionalComponent(__VLS_303, new __VLS_303({
    modelValue: (__VLS_ctx.form.label),
    placeholder: "请输入标签",
}));
const __VLS_305 = __VLS_304({
    modelValue: (__VLS_ctx.form.label),
    placeholder: "请输入标签",
}, ...__VLS_functionalComponentArgsRest(__VLS_304));
var __VLS_302;
const __VLS_307 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_308 = __VLS_asFunctionalComponent(__VLS_307, new __VLS_307({
    label: "标签类型",
}));
const __VLS_309 = __VLS_308({
    label: "标签类型",
}, ...__VLS_functionalComponentArgsRest(__VLS_308));
__VLS_310.slots.default;
const __VLS_311 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_312 = __VLS_asFunctionalComponent(__VLS_311, new __VLS_311({
    modelValue: (__VLS_ctx.form.tag_type),
    placeholder: "例如 success / warning / info",
}));
const __VLS_313 = __VLS_312({
    modelValue: (__VLS_ctx.form.tag_type),
    placeholder: "例如 success / warning / info",
}, ...__VLS_functionalComponentArgsRest(__VLS_312));
var __VLS_310;
const __VLS_315 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_316 = __VLS_asFunctionalComponent(__VLS_315, new __VLS_315({
    label: "标签颜色",
}));
const __VLS_317 = __VLS_316({
    label: "标签颜色",
}, ...__VLS_functionalComponentArgsRest(__VLS_316));
__VLS_318.slots.default;
const __VLS_319 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_320 = __VLS_asFunctionalComponent(__VLS_319, new __VLS_319({
    modelValue: (__VLS_ctx.form.tag_color),
    placeholder: "例如 #67C23A",
}));
const __VLS_321 = __VLS_320({
    modelValue: (__VLS_ctx.form.tag_color),
    placeholder: "例如 #67C23A",
}, ...__VLS_functionalComponentArgsRest(__VLS_320));
var __VLS_318;
const __VLS_323 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_324 = __VLS_asFunctionalComponent(__VLS_323, new __VLS_323({
    label: "扩展值",
}));
const __VLS_325 = __VLS_324({
    label: "扩展值",
}, ...__VLS_functionalComponentArgsRest(__VLS_324));
__VLS_326.slots.default;
const __VLS_327 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_328 = __VLS_asFunctionalComponent(__VLS_327, new __VLS_327({
    modelValue: (__VLS_ctx.form.extra),
    type: "textarea",
    rows: (3),
    placeholder: "请输入扩展 JSON 或文本",
}));
const __VLS_329 = __VLS_328({
    modelValue: (__VLS_ctx.form.extra),
    type: "textarea",
    rows: (3),
    placeholder: "请输入扩展 JSON 或文本",
}, ...__VLS_functionalComponentArgsRest(__VLS_328));
var __VLS_326;
const __VLS_331 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_332 = __VLS_asFunctionalComponent(__VLS_331, new __VLS_331({
    label: "默认项",
}));
const __VLS_333 = __VLS_332({
    label: "默认项",
}, ...__VLS_functionalComponentArgsRest(__VLS_332));
__VLS_334.slots.default;
const __VLS_335 = {}.ElSwitch;
/** @type {[typeof __VLS_components.ElSwitch, typeof __VLS_components.elSwitch, ]} */ ;
// @ts-ignore
const __VLS_336 = __VLS_asFunctionalComponent(__VLS_335, new __VLS_335({
    modelValue: (__VLS_ctx.form.is_default),
}));
const __VLS_337 = __VLS_336({
    modelValue: (__VLS_ctx.form.is_default),
}, ...__VLS_functionalComponentArgsRest(__VLS_336));
var __VLS_334;
const __VLS_339 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_340 = __VLS_asFunctionalComponent(__VLS_339, new __VLS_339({
    label: "状态",
}));
const __VLS_341 = __VLS_340({
    label: "状态",
}, ...__VLS_functionalComponentArgsRest(__VLS_340));
__VLS_342.slots.default;
const __VLS_343 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_344 = __VLS_asFunctionalComponent(__VLS_343, new __VLS_343({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}));
const __VLS_345 = __VLS_344({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_344));
__VLS_346.slots.default;
const __VLS_347 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_348 = __VLS_asFunctionalComponent(__VLS_347, new __VLS_347({
    label: "启用",
    value: "enabled",
}));
const __VLS_349 = __VLS_348({
    label: "启用",
    value: "enabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_348));
const __VLS_351 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_352 = __VLS_asFunctionalComponent(__VLS_351, new __VLS_351({
    label: "禁用",
    value: "disabled",
}));
const __VLS_353 = __VLS_352({
    label: "禁用",
    value: "disabled",
}, ...__VLS_functionalComponentArgsRest(__VLS_352));
var __VLS_346;
var __VLS_342;
const __VLS_355 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_356 = __VLS_asFunctionalComponent(__VLS_355, new __VLS_355({
    label: "排序",
}));
const __VLS_357 = __VLS_356({
    label: "排序",
}, ...__VLS_functionalComponentArgsRest(__VLS_356));
__VLS_358.slots.default;
const __VLS_359 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_360 = __VLS_asFunctionalComponent(__VLS_359, new __VLS_359({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}));
const __VLS_361 = __VLS_360({
    modelValue: (__VLS_ctx.form.sort),
    min: (0),
    step: (1),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_360));
var __VLS_358;
const __VLS_363 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_364 = __VLS_asFunctionalComponent(__VLS_363, new __VLS_363({
    label: "备注",
}));
const __VLS_365 = __VLS_364({
    label: "备注",
}, ...__VLS_functionalComponentArgsRest(__VLS_364));
__VLS_366.slots.default;
const __VLS_367 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_368 = __VLS_asFunctionalComponent(__VLS_367, new __VLS_367({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: "请输入备注",
}));
const __VLS_369 = __VLS_368({
    modelValue: (__VLS_ctx.form.remark),
    type: "textarea",
    rows: (3),
    placeholder: "请输入备注",
}, ...__VLS_functionalComponentArgsRest(__VLS_368));
var __VLS_366;
var __VLS_278;
var __VLS_270;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-filters']} */ ;
/** @type {__VLS_StyleScopedClasses['admin-pagination']} */ ;
/** @type {__VLS_StyleScopedClasses['lookup-result-card']} */ ;
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
            categoryLoading: categoryLoading,
            lookupLoading: lookupLoading,
            rows: rows,
            total: total,
            categoryOptions: categoryOptions,
            editingId: editingId,
            lookupResult: lookupResult,
            lookupItems: lookupItems,
            query: query,
            lookupForm: lookupForm,
            form: form,
            loadItems: loadItems,
            categoryLabel: categoryLabel,
            openCreate: openCreate,
            openEdit: openEdit,
            statusLabel: statusLabel,
            defaultLabel: defaultLabel,
            submitForm: submitForm,
            removeRow: removeRow,
            handleSearch: handleSearch,
            handleReset: handleReset,
            handlePageChange: handlePageChange,
            handleSizeChange: handleSizeChange,
            runLookupList: runLookupList,
            runLookupItem: runLookupItem,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
