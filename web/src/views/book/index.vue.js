import { onActivated, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createBook, deleteBook, listbooks, updateBook } from '@/api/book';
import { useAppI18n } from '@/i18n';
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
    tenant_id: '',
    title: '',
    author: '',
    isbn: '',
    publisher: '',
    publish_date: '',
    category: '',
    description: '',
    status: '',
    price: 0,
    stock_quantity: 0,
    cover_image_url: '',
    tags: '',
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
function categoryEnumLabel(value) {
    return formatEnumLabel(value, {
        tech: t('book.category.tech', 'Technology'),
        novel: t('book.category.novel', 'Novel'),
        history: t('book.category.history', 'History'),
        other: t('book.category.other', 'Other'),
    });
}
function statusEnumLabel(value) {
    return formatEnumLabel(value, {
        draft: t('book.status.draft', 'Draft'),
        published: t('book.status.published', 'Published'),
        off_shelf: t('book.status.off_shelf', 'Off shelf'),
    });
}
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadItems() {
    tableLoading.value = true;
    try {
        const response = await listbooks({ ...query });
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
        tenant_id: row.tenant_id ?? '',
        title: row.title ?? '',
        author: row.author ?? '',
        isbn: row.isbn ?? '',
        publisher: row.publisher ?? '',
        publish_date: row.publish_date ?? '',
        category: row.category ?? '',
        description: row.description ?? '',
        status: row.status ?? '',
        price: Number(row.price ?? 0),
        stock_quantity: Number(row.stock_quantity ?? 0),
        cover_image_url: row.cover_image_url ?? '',
        tags: row.tags ?? '',
    });
    dialogVisible.value = true;
}
async function submitForm() {
    dialogLoading.value = true;
    try {
        const payload = {
            tenant_id: form.tenant_id.trim(),
            title: form.title.trim(),
            author: form.author.trim(),
            isbn: form.isbn.trim(),
            publisher: form.publisher.trim(),
            publish_date: form.publish_date,
            category: form.category,
            description: form.description.trim(),
            status: form.status,
            price: Number(form.price ?? 0),
            stock_quantity: Number(form.stock_quantity ?? 0),
            cover_image_url: form.cover_image_url.trim(),
            tags: form.tags.trim(),
        };
        if (editingId.value) {
            await updateBook(editingId.value, payload);
            ElMessage.success(t('book.updated', 'Book updated'));
        }
        else {
            await createBook(payload);
            ElMessage.success(t('book.created', 'Book created'));
        }
        dialogVisible.value = false;
        await loadItems();
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('book.save_failed', 'Save failed'));
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('book.confirm_delete', 'Delete Book {id}?', { id: row.id }), t('book.delete_title', 'Delete Book'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteBook(row.id);
    ElMessage.success(t('book.deleted', 'Book deleted'));
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
onActivated(() => {
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
    title: (__VLS_ctx.t('book.title', 'Book management')),
    description: (__VLS_ctx.t('book.description', 'CRUD page generated by goadmin-cli, ready for listing, editing, and deletion.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('book.title', 'Book management')),
    description: (__VLS_ctx.t('book.description', 'CRUD page generated by goadmin-cli, ready for listing, editing, and deletion.')),
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('book:create') }, null, null);
    __VLS_14.slots.default;
    (__VLS_ctx.t('book.create', 'Add Book'));
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
        label: (__VLS_ctx.t('book.keyword', 'Keyword')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('book.keyword', 'Keyword')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('book.keyword_placeholder', 'Search Book data')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('book.keyword_placeholder', 'Search Book data')),
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
    label: (__VLS_ctx.t('book.id', 'ID')),
    minWidth: "160",
}));
const __VLS_57 = __VLS_56({
    prop: "id",
    label: (__VLS_ctx.t('book.id', 'ID')),
    minWidth: "160",
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
const __VLS_59 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    prop: "tenant_id",
    label: (__VLS_ctx.t('book.tenant_id', 'Tenant ID')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_61 = __VLS_60({
    prop: "tenant_id",
    label: (__VLS_ctx.t('book.tenant_id', 'Tenant ID')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_62.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.tenant_id || '-');
}
var __VLS_62;
const __VLS_63 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    prop: "title",
    label: (__VLS_ctx.t('book.title_field', 'Title')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_65 = __VLS_64({
    prop: "title",
    label: (__VLS_ctx.t('book.title_field', 'Title')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_66.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.title || '-');
}
var __VLS_66;
const __VLS_67 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    prop: "author",
    label: (__VLS_ctx.t('book.author', 'Author')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_69 = __VLS_68({
    prop: "author",
    label: (__VLS_ctx.t('book.author', 'Author')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_70.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.author || '-');
}
var __VLS_70;
const __VLS_71 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    prop: "isbn",
    label: (__VLS_ctx.t('book.isbn', 'Isbn')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_73 = __VLS_72({
    prop: "isbn",
    label: (__VLS_ctx.t('book.isbn', 'Isbn')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_74.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.isbn || '-');
}
var __VLS_74;
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "publisher",
    label: (__VLS_ctx.t('book.publisher', 'Publisher')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_77 = __VLS_76({
    prop: "publisher",
    label: (__VLS_ctx.t('book.publisher', 'Publisher')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_78.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.publisher || '-');
}
var __VLS_78;
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "publish_date",
    label: (__VLS_ctx.t('book.publish_date', 'Publish Date')),
    minWidth: "180",
}));
const __VLS_81 = __VLS_80({
    prop: "publish_date",
    label: (__VLS_ctx.t('book.publish_date', 'Publish Date')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_82.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.publish_date));
}
var __VLS_82;
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    prop: "category",
    label: (__VLS_ctx.t('book.category', 'Category')),
    minWidth: "140",
}));
const __VLS_85 = __VLS_84({
    prop: "category",
    label: (__VLS_ctx.t('book.category', 'Category')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_86.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.categoryEnumLabel(row.category));
}
var __VLS_86;
const __VLS_87 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    prop: "description",
    label: (__VLS_ctx.t('book.description_field', 'Description')),
    minWidth: "220",
    showOverflowTooltip: true,
}));
const __VLS_89 = __VLS_88({
    prop: "description",
    label: (__VLS_ctx.t('book.description_field', 'Description')),
    minWidth: "220",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_90.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.description || '-');
}
var __VLS_90;
const __VLS_91 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
    prop: "status",
    label: (__VLS_ctx.t('book.status', 'Status')),
    minWidth: "140",
}));
const __VLS_93 = __VLS_92({
    prop: "status",
    label: (__VLS_ctx.t('book.status', 'Status')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_92));
__VLS_94.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_94.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.statusEnumLabel(row.status));
}
var __VLS_94;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    prop: "price",
    label: (__VLS_ctx.t('book.price', 'Price')),
    minWidth: "120",
}));
const __VLS_97 = __VLS_96({
    prop: "price",
    label: (__VLS_ctx.t('book.price', 'Price')),
    minWidth: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_98.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.price || '-');
}
var __VLS_98;
const __VLS_99 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
    prop: "stock_quantity",
    label: (__VLS_ctx.t('book.stock_quantity', 'Stock Quantity')),
    minWidth: "120",
}));
const __VLS_101 = __VLS_100({
    prop: "stock_quantity",
    label: (__VLS_ctx.t('book.stock_quantity', 'Stock Quantity')),
    minWidth: "120",
}, ...__VLS_functionalComponentArgsRest(__VLS_100));
__VLS_102.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_102.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.stock_quantity || '-');
}
var __VLS_102;
const __VLS_103 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
    prop: "cover_image_url",
    label: (__VLS_ctx.t('book.cover_image_url', 'Cover Image Url')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_105 = __VLS_104({
    prop: "cover_image_url",
    label: (__VLS_ctx.t('book.cover_image_url', 'Cover Image Url')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_104));
__VLS_106.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_106.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.cover_image_url || '-');
}
var __VLS_106;
const __VLS_107 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    prop: "tags",
    label: (__VLS_ctx.t('book.tags', 'Tags')),
    minWidth: "140",
    showOverflowTooltip: true,
}));
const __VLS_109 = __VLS_108({
    prop: "tags",
    label: (__VLS_ctx.t('book.tags', 'Tags')),
    minWidth: "140",
    showOverflowTooltip: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_110.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (row.tags || '-');
}
var __VLS_110;
const __VLS_111 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('book.created_at', 'Created at')),
    minWidth: "180",
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('book.created_at', 'Created at')),
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
    label: (__VLS_ctx.t('book.updated_at', 'Updated at')),
    minWidth: "180",
}));
const __VLS_117 = __VLS_116({
    label: (__VLS_ctx.t('book.updated_at', 'Updated at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_116));
__VLS_118.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_118.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.updated_at));
}
var __VLS_118;
const __VLS_119 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_120 = __VLS_asFunctionalComponent(__VLS_119, new __VLS_119({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_121 = __VLS_120({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_120));
__VLS_122.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_122.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_123 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_125 = __VLS_124({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_124));
    let __VLS_127;
    let __VLS_128;
    let __VLS_129;
    const __VLS_130 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('book:update') }, null, null);
    __VLS_126.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_126;
    const __VLS_131 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_133 = __VLS_132({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_132));
    let __VLS_135;
    let __VLS_136;
    let __VLS_137;
    const __VLS_138 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('book:delete') }, null, null);
    __VLS_134.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_134;
}
var __VLS_122;
var __VLS_54;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_139 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_140 = __VLS_asFunctionalComponent(__VLS_139, new __VLS_139({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_141 = __VLS_140({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_140));
    let __VLS_143;
    let __VLS_144;
    let __VLS_145;
    const __VLS_146 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_147 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_142;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('book.edit_title', 'Edit Book') : __VLS_ctx.t('book.create_title', 'Add Book')),
    loading: (__VLS_ctx.dialogLoading),
}));
const __VLS_149 = __VLS_148({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('book.edit_title', 'Edit Book') : __VLS_ctx.t('book.create_title', 'Add Book')),
    loading: (__VLS_ctx.dialogLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
let __VLS_151;
let __VLS_152;
let __VLS_153;
const __VLS_154 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_150.slots.default;
const __VLS_155 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_157 = __VLS_156({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
__VLS_158.slots.default;
const __VLS_159 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    label: (__VLS_ctx.t('book.tenant_id', 'Tenant ID')),
}));
const __VLS_161 = __VLS_160({
    label: (__VLS_ctx.t('book.tenant_id', 'Tenant ID')),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    modelValue: (__VLS_ctx.form.tenant_id),
    placeholder: (__VLS_ctx.t('book.tenant_id_placeholder', 'Enter Tenant ID')),
}));
const __VLS_165 = __VLS_164({
    modelValue: (__VLS_ctx.form.tenant_id),
    placeholder: (__VLS_ctx.t('book.tenant_id_placeholder', 'Enter Tenant ID')),
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
var __VLS_162;
const __VLS_167 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('book.title_field', 'Title')),
    required: true,
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('book.title_field', 'Title')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
const __VLS_171 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    modelValue: (__VLS_ctx.form.title),
    placeholder: (__VLS_ctx.t('book.title_placeholder', 'Enter title')),
}));
const __VLS_173 = __VLS_172({
    modelValue: (__VLS_ctx.form.title),
    placeholder: (__VLS_ctx.t('book.title_placeholder', 'Enter title')),
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
var __VLS_170;
const __VLS_175 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    label: (__VLS_ctx.t('book.author', 'Author')),
}));
const __VLS_177 = __VLS_176({
    label: (__VLS_ctx.t('book.author', 'Author')),
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
__VLS_178.slots.default;
const __VLS_179 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    modelValue: (__VLS_ctx.form.author),
    placeholder: (__VLS_ctx.t('book.author_placeholder', 'Enter author')),
}));
const __VLS_181 = __VLS_180({
    modelValue: (__VLS_ctx.form.author),
    placeholder: (__VLS_ctx.t('book.author_placeholder', 'Enter author')),
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
var __VLS_178;
const __VLS_183 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    label: (__VLS_ctx.t('book.isbn', 'Isbn')),
}));
const __VLS_185 = __VLS_184({
    label: (__VLS_ctx.t('book.isbn', 'Isbn')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
const __VLS_187 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    modelValue: (__VLS_ctx.form.isbn),
    placeholder: (__VLS_ctx.t('book.isbn_placeholder', 'Enter isbn')),
}));
const __VLS_189 = __VLS_188({
    modelValue: (__VLS_ctx.form.isbn),
    placeholder: (__VLS_ctx.t('book.isbn_placeholder', 'Enter isbn')),
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
var __VLS_186;
const __VLS_191 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    label: (__VLS_ctx.t('book.publisher', 'Publisher')),
}));
const __VLS_193 = __VLS_192({
    label: (__VLS_ctx.t('book.publisher', 'Publisher')),
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
__VLS_194.slots.default;
const __VLS_195 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
    modelValue: (__VLS_ctx.form.publisher),
    placeholder: (__VLS_ctx.t('book.publisher_placeholder', 'Enter publisher')),
}));
const __VLS_197 = __VLS_196({
    modelValue: (__VLS_ctx.form.publisher),
    placeholder: (__VLS_ctx.t('book.publisher_placeholder', 'Enter publisher')),
}, ...__VLS_functionalComponentArgsRest(__VLS_196));
var __VLS_194;
const __VLS_199 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_200 = __VLS_asFunctionalComponent(__VLS_199, new __VLS_199({
    label: (__VLS_ctx.t('book.publish_date', 'Publish Date')),
}));
const __VLS_201 = __VLS_200({
    label: (__VLS_ctx.t('book.publish_date', 'Publish Date')),
}, ...__VLS_functionalComponentArgsRest(__VLS_200));
__VLS_202.slots.default;
const __VLS_203 = {}.ElDatePicker;
/** @type {[typeof __VLS_components.ElDatePicker, typeof __VLS_components.elDatePicker, ]} */ ;
// @ts-ignore
const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({
    modelValue: (__VLS_ctx.form.publish_date),
    type: "datetime",
    format: "YYYY-MM-DD HH:mm:ss",
    valueFormat: "YYYY-MM-DDTHH:mm:ssZ",
    placeholder: (__VLS_ctx.t('book.publish_date_placeholder', 'Select publish date')),
    ...{ style: {} },
}));
const __VLS_205 = __VLS_204({
    modelValue: (__VLS_ctx.form.publish_date),
    type: "datetime",
    format: "YYYY-MM-DD HH:mm:ss",
    valueFormat: "YYYY-MM-DDTHH:mm:ssZ",
    placeholder: (__VLS_ctx.t('book.publish_date_placeholder', 'Select publish date')),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_204));
var __VLS_202;
const __VLS_207 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_208 = __VLS_asFunctionalComponent(__VLS_207, new __VLS_207({
    label: (__VLS_ctx.t('book.category', 'Category')),
}));
const __VLS_209 = __VLS_208({
    label: (__VLS_ctx.t('book.category', 'Category')),
}, ...__VLS_functionalComponentArgsRest(__VLS_208));
__VLS_210.slots.default;
const __VLS_211 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
    modelValue: (__VLS_ctx.form.category),
    ...{ style: {} },
}));
const __VLS_213 = __VLS_212({
    modelValue: (__VLS_ctx.form.category),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_212));
__VLS_214.slots.default;
const __VLS_215 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_216 = __VLS_asFunctionalComponent(__VLS_215, new __VLS_215({
    label: (__VLS_ctx.t('book.category.tech', 'Technology')),
    value: "tech",
}));
const __VLS_217 = __VLS_216({
    label: (__VLS_ctx.t('book.category.tech', 'Technology')),
    value: "tech",
}, ...__VLS_functionalComponentArgsRest(__VLS_216));
const __VLS_219 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_220 = __VLS_asFunctionalComponent(__VLS_219, new __VLS_219({
    label: (__VLS_ctx.t('book.category.novel', 'Novel')),
    value: "novel",
}));
const __VLS_221 = __VLS_220({
    label: (__VLS_ctx.t('book.category.novel', 'Novel')),
    value: "novel",
}, ...__VLS_functionalComponentArgsRest(__VLS_220));
const __VLS_223 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_224 = __VLS_asFunctionalComponent(__VLS_223, new __VLS_223({
    label: (__VLS_ctx.t('book.category.history', 'History')),
    value: "history",
}));
const __VLS_225 = __VLS_224({
    label: (__VLS_ctx.t('book.category.history', 'History')),
    value: "history",
}, ...__VLS_functionalComponentArgsRest(__VLS_224));
const __VLS_227 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_228 = __VLS_asFunctionalComponent(__VLS_227, new __VLS_227({
    label: (__VLS_ctx.t('book.category.other', 'Other')),
    value: "other",
}));
const __VLS_229 = __VLS_228({
    label: (__VLS_ctx.t('book.category.other', 'Other')),
    value: "other",
}, ...__VLS_functionalComponentArgsRest(__VLS_228));
var __VLS_214;
var __VLS_210;
const __VLS_231 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_232 = __VLS_asFunctionalComponent(__VLS_231, new __VLS_231({
    label: (__VLS_ctx.t('book.description_field', 'Description')),
}));
const __VLS_233 = __VLS_232({
    label: (__VLS_ctx.t('book.description_field', 'Description')),
}, ...__VLS_functionalComponentArgsRest(__VLS_232));
__VLS_234.slots.default;
const __VLS_235 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_236 = __VLS_asFunctionalComponent(__VLS_235, new __VLS_235({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('book.description_placeholder', 'Enter description')),
}));
const __VLS_237 = __VLS_236({
    modelValue: (__VLS_ctx.form.description),
    type: "textarea",
    rows: (4),
    placeholder: (__VLS_ctx.t('book.description_placeholder', 'Enter description')),
}, ...__VLS_functionalComponentArgsRest(__VLS_236));
var __VLS_234;
const __VLS_239 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_240 = __VLS_asFunctionalComponent(__VLS_239, new __VLS_239({
    label: (__VLS_ctx.t('book.status', 'Status')),
}));
const __VLS_241 = __VLS_240({
    label: (__VLS_ctx.t('book.status', 'Status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_240));
__VLS_242.slots.default;
const __VLS_243 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_244 = __VLS_asFunctionalComponent(__VLS_243, new __VLS_243({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}));
const __VLS_245 = __VLS_244({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_244));
__VLS_246.slots.default;
const __VLS_247 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_248 = __VLS_asFunctionalComponent(__VLS_247, new __VLS_247({
    label: (__VLS_ctx.t('book.status.draft', 'Draft')),
    value: "draft",
}));
const __VLS_249 = __VLS_248({
    label: (__VLS_ctx.t('book.status.draft', 'Draft')),
    value: "draft",
}, ...__VLS_functionalComponentArgsRest(__VLS_248));
const __VLS_251 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_252 = __VLS_asFunctionalComponent(__VLS_251, new __VLS_251({
    label: (__VLS_ctx.t('book.status.published', 'Published')),
    value: "published",
}));
const __VLS_253 = __VLS_252({
    label: (__VLS_ctx.t('book.status.published', 'Published')),
    value: "published",
}, ...__VLS_functionalComponentArgsRest(__VLS_252));
const __VLS_255 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_256 = __VLS_asFunctionalComponent(__VLS_255, new __VLS_255({
    label: (__VLS_ctx.t('book.status.off_shelf', 'Off shelf')),
    value: "off_shelf",
}));
const __VLS_257 = __VLS_256({
    label: (__VLS_ctx.t('book.status.off_shelf', 'Off shelf')),
    value: "off_shelf",
}, ...__VLS_functionalComponentArgsRest(__VLS_256));
var __VLS_246;
var __VLS_242;
const __VLS_259 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_260 = __VLS_asFunctionalComponent(__VLS_259, new __VLS_259({
    label: (__VLS_ctx.t('book.price', 'Price')),
}));
const __VLS_261 = __VLS_260({
    label: (__VLS_ctx.t('book.price', 'Price')),
}, ...__VLS_functionalComponentArgsRest(__VLS_260));
__VLS_262.slots.default;
const __VLS_263 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_264 = __VLS_asFunctionalComponent(__VLS_263, new __VLS_263({
    modelValue: (__VLS_ctx.form.price),
    controls: (false),
    ...{ style: {} },
}));
const __VLS_265 = __VLS_264({
    modelValue: (__VLS_ctx.form.price),
    controls: (false),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_264));
var __VLS_262;
const __VLS_267 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_268 = __VLS_asFunctionalComponent(__VLS_267, new __VLS_267({
    label: (__VLS_ctx.t('book.stock_quantity', 'Stock Quantity')),
}));
const __VLS_269 = __VLS_268({
    label: (__VLS_ctx.t('book.stock_quantity', 'Stock Quantity')),
}, ...__VLS_functionalComponentArgsRest(__VLS_268));
__VLS_270.slots.default;
const __VLS_271 = {}.ElInputNumber;
/** @type {[typeof __VLS_components.ElInputNumber, typeof __VLS_components.elInputNumber, ]} */ ;
// @ts-ignore
const __VLS_272 = __VLS_asFunctionalComponent(__VLS_271, new __VLS_271({
    modelValue: (__VLS_ctx.form.stock_quantity),
    controls: (false),
    ...{ style: {} },
}));
const __VLS_273 = __VLS_272({
    modelValue: (__VLS_ctx.form.stock_quantity),
    controls: (false),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_272));
var __VLS_270;
const __VLS_275 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_276 = __VLS_asFunctionalComponent(__VLS_275, new __VLS_275({
    label: (__VLS_ctx.t('book.cover_image_url', 'Cover Image Url')),
}));
const __VLS_277 = __VLS_276({
    label: (__VLS_ctx.t('book.cover_image_url', 'Cover Image Url')),
}, ...__VLS_functionalComponentArgsRest(__VLS_276));
__VLS_278.slots.default;
const __VLS_279 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_280 = __VLS_asFunctionalComponent(__VLS_279, new __VLS_279({
    modelValue: (__VLS_ctx.form.cover_image_url),
    placeholder: (__VLS_ctx.t('book.cover_image_url_placeholder', 'Enter cover image url')),
}));
const __VLS_281 = __VLS_280({
    modelValue: (__VLS_ctx.form.cover_image_url),
    placeholder: (__VLS_ctx.t('book.cover_image_url_placeholder', 'Enter cover image url')),
}, ...__VLS_functionalComponentArgsRest(__VLS_280));
var __VLS_278;
const __VLS_283 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_284 = __VLS_asFunctionalComponent(__VLS_283, new __VLS_283({
    label: (__VLS_ctx.t('book.tags', 'Tags')),
}));
const __VLS_285 = __VLS_284({
    label: (__VLS_ctx.t('book.tags', 'Tags')),
}, ...__VLS_functionalComponentArgsRest(__VLS_284));
__VLS_286.slots.default;
const __VLS_287 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_288 = __VLS_asFunctionalComponent(__VLS_287, new __VLS_287({
    modelValue: (__VLS_ctx.form.tags),
    placeholder: (__VLS_ctx.t('book.tags_placeholder', 'Comma-separated, for example: ai,ml')),
}));
const __VLS_289 = __VLS_288({
    modelValue: (__VLS_ctx.form.tags),
    placeholder: (__VLS_ctx.t('book.tags_placeholder', 'Comma-separated, for example: ai,ml')),
}, ...__VLS_functionalComponentArgsRest(__VLS_288));
var __VLS_286;
var __VLS_158;
var __VLS_150;
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
            categoryEnumLabel: categoryEnumLabel,
            statusEnumLabel: statusEnumLabel,
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
