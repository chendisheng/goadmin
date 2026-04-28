import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchRoles } from '@/api/roles';
import { createUser, deleteUser, fetchUsers, updateUser } from '@/api/users';
import { useAppI18n } from '@/i18n';
import { useSessionStore } from '@/store/session';
import { formatDateTime, statusTagType } from '@/utils/admin';
const sessionStore = useSessionStore();
const { t } = useAppI18n();
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const roleLoading = ref(false);
const total = ref(0);
const rows = ref([]);
const roleOptions = ref([]);
const editingId = ref('');
const query = reactive({
    keyword: '',
    status: '',
    page: 1,
    page_size: 10,
});
const defaultForm = () => ({
    tenant_id: sessionStore.currentUser?.tenant_id ?? '',
    username: '',
    display_name: '',
    mobile: '',
    email: '',
    status: 'active',
    role_codes: [],
    password_hash: '',
});
const form = reactive(defaultForm());
function resetForm() {
    Object.assign(form, defaultForm());
}
async function loadUsers() {
    tableLoading.value = true;
    try {
        const response = await fetchUsers({ ...query });
        rows.value = response.items;
        total.value = response.total;
    }
    finally {
        tableLoading.value = false;
    }
}
async function loadRoles() {
    roleLoading.value = true;
    try {
        const response = await fetchRoles({ keyword: '', status: '', tenant_id: '', page: 1, page_size: 200 });
        roleOptions.value = response.items;
    }
    finally {
        roleLoading.value = false;
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
        tenant_id: row.tenant_id ?? sessionStore.currentUser?.tenant_id ?? '',
        username: row.username,
        display_name: row.display_name ?? '',
        mobile: row.mobile ?? '',
        email: row.email ?? '',
        status: row.status || 'active',
        role_codes: [...(row.role_codes ?? [])],
        password_hash: '',
    });
    dialogVisible.value = true;
}
function resolveRoleLabel(code) {
    const role = roleOptions.value.find((item) => item.code === code);
    return role ? `${role.name} (${role.code})` : code;
}
function statusLabel(status) {
    return status === 'inactive' ? t('user.status_inactive', 'Disabled') : t('user.status_active', 'Enabled');
}
async function submitForm() {
    if (form.username.trim() === '') {
        ElMessage.warning(t('user.username_required', 'Enter username'));
        return;
    }
    dialogLoading.value = true;
    try {
        const payload = {
            ...form,
            tenant_id: form.tenant_id.trim(),
            username: form.username.trim(),
            display_name: form.display_name.trim(),
            mobile: form.mobile.trim(),
            email: form.email.trim(),
            status: form.status.trim() || 'active',
            role_codes: [...form.role_codes],
            password_hash: form.password_hash.trim(),
        };
        if (editingId.value) {
            await updateUser(editingId.value, payload);
            ElMessage.success(t('user.updated', 'User updated'));
        }
        else {
            await createUser(payload);
            ElMessage.success(t('user.created', 'User created'));
        }
        dialogVisible.value = false;
        await loadUsers();
    }
    finally {
        dialogLoading.value = false;
    }
}
async function removeRow(row) {
    await ElMessageBox.confirm(t('user.confirm_delete', 'Delete user {name}?', { name: row.username }), t('user.delete_title', 'Delete user'), {
        type: 'warning',
        confirmButtonText: t('common.delete', 'Delete'),
        cancelButtonText: t('common.cancel', 'Cancel'),
    });
    await deleteUser(row.id);
    ElMessage.success(t('user.deleted', 'User deleted'));
    await loadUsers();
}
function handleSearch() {
    query.page = 1;
    void loadUsers();
}
function handleReset() {
    query.keyword = '';
    query.status = '';
    query.page = 1;
    void loadUsers();
}
function handlePageChange(page) {
    query.page = page;
    void loadUsers();
}
function handleSizeChange(pageSize) {
    query.page_size = pageSize;
    query.page = 1;
    void loadUsers();
}
onMounted(() => {
    void Promise.all([loadUsers(), loadRoles()]);
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
    title: (__VLS_ctx.t('user.title', 'User management')),
    description: (__VLS_ctx.t('user.description', 'Maintain users, role bindings, and basic profile data.')),
    loading: (__VLS_ctx.tableLoading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('user.title', 'User management')),
    description: (__VLS_ctx.t('user.description', 'Maintain users, role bindings, and basic profile data.')),
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
        onClick: (__VLS_ctx.loadUsers)
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
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('user:create') }, null, null);
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
        label: (__VLS_ctx.t('user.keyword_label', 'Keyword')),
    }));
    const __VLS_25 = __VLS_24({
        label: (__VLS_ctx.t('user.keyword_label', 'Keyword')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_24));
    __VLS_26.slots.default;
    const __VLS_27 = {}.ElInput;
    /** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('user.keyword_placeholder', 'Username / display name / email')),
    }));
    const __VLS_29 = __VLS_28({
        modelValue: (__VLS_ctx.query.keyword),
        clearable: true,
        placeholder: (__VLS_ctx.t('user.keyword_placeholder', 'Username / display name / email')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    var __VLS_26;
    const __VLS_31 = {}.ElFormItem;
    /** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
    // @ts-ignore
    const __VLS_32 = __VLS_asFunctionalComponent(__VLS_31, new __VLS_31({
        label: (__VLS_ctx.t('user.status_label', 'Status')),
    }));
    const __VLS_33 = __VLS_32({
        label: (__VLS_ctx.t('user.status_label', 'Status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_32));
    __VLS_34.slots.default;
    const __VLS_35 = {}.ElSelect;
    /** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('user.status_placeholder', 'All statuses')),
        ...{ style: {} },
    }));
    const __VLS_37 = __VLS_36({
        modelValue: (__VLS_ctx.query.status),
        clearable: true,
        placeholder: (__VLS_ctx.t('user.status_placeholder', 'All statuses')),
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    __VLS_38.slots.default;
    const __VLS_39 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_40 = __VLS_asFunctionalComponent(__VLS_39, new __VLS_39({
        label: (__VLS_ctx.t('user.status_active', 'Enabled')),
        value: "active",
    }));
    const __VLS_41 = __VLS_40({
        label: (__VLS_ctx.t('user.status_active', 'Enabled')),
        value: "active",
    }, ...__VLS_functionalComponentArgsRest(__VLS_40));
    const __VLS_43 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
        label: (__VLS_ctx.t('user.status_inactive', 'Disabled')),
        value: "inactive",
    }));
    const __VLS_45 = __VLS_44({
        label: (__VLS_ctx.t('user.status_inactive', 'Disabled')),
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
    prop: "username",
    label: (__VLS_ctx.t('user.username', 'Username')),
    minWidth: "140",
}));
const __VLS_73 = __VLS_72({
    prop: "username",
    label: (__VLS_ctx.t('user.username', 'Username')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
const __VLS_75 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    prop: "display_name",
    label: (__VLS_ctx.t('user.display_name', 'Display name')),
    minWidth: "140",
}));
const __VLS_77 = __VLS_76({
    prop: "display_name",
    label: (__VLS_ctx.t('user.display_name', 'Display name')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
const __VLS_79 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    prop: "mobile",
    label: (__VLS_ctx.t('user.mobile', 'Mobile')),
    minWidth: "140",
}));
const __VLS_81 = __VLS_80({
    prop: "mobile",
    label: (__VLS_ctx.t('user.mobile', 'Mobile')),
    minWidth: "140",
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
const __VLS_83 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    prop: "email",
    label: (__VLS_ctx.t('user.email', 'Email')),
    minWidth: "200",
}));
const __VLS_85 = __VLS_84({
    prop: "email",
    label: (__VLS_ctx.t('user.email', 'Email')),
    minWidth: "200",
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
const __VLS_87 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({
    label: (__VLS_ctx.t('user.status', 'Status')),
    width: "100",
}));
const __VLS_89 = __VLS_88({
    label: (__VLS_ctx.t('user.status', 'Status')),
    width: "100",
}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_90.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_90.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_91 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_92 = __VLS_asFunctionalComponent(__VLS_91, new __VLS_91({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }));
    const __VLS_93 = __VLS_92({
        type: (__VLS_ctx.statusTagType(row.status)),
        effect: "plain",
    }, ...__VLS_functionalComponentArgsRest(__VLS_92));
    __VLS_94.slots.default;
    (__VLS_ctx.statusLabel(row.status));
    var __VLS_94;
}
var __VLS_90;
const __VLS_95 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_96 = __VLS_asFunctionalComponent(__VLS_95, new __VLS_95({
    label: (__VLS_ctx.t('user.role', 'Roles')),
    minWidth: "220",
}));
const __VLS_97 = __VLS_96({
    label: (__VLS_ctx.t('user.role', 'Roles')),
    minWidth: "220",
}, ...__VLS_functionalComponentArgsRest(__VLS_96));
__VLS_98.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_98.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_99 = {}.ElSpace;
    /** @type {[typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, typeof __VLS_components.ElSpace, typeof __VLS_components.elSpace, ]} */ ;
    // @ts-ignore
    const __VLS_100 = __VLS_asFunctionalComponent(__VLS_99, new __VLS_99({
        wrap: true,
    }));
    const __VLS_101 = __VLS_100({
        wrap: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_100));
    __VLS_102.slots.default;
    for (const [code] of __VLS_getVForSourceType((row.role_codes || []))) {
        const __VLS_103 = {}.ElTag;
        /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
        // @ts-ignore
        const __VLS_104 = __VLS_asFunctionalComponent(__VLS_103, new __VLS_103({
            key: (code),
            effect: "plain",
        }));
        const __VLS_105 = __VLS_104({
            key: (code),
            effect: "plain",
        }, ...__VLS_functionalComponentArgsRest(__VLS_104));
        __VLS_106.slots.default;
        (__VLS_ctx.resolveRoleLabel(code));
        var __VLS_106;
    }
    if (!row.role_codes || row.role_codes.length === 0) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    }
    var __VLS_102;
}
var __VLS_98;
const __VLS_107 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_108 = __VLS_asFunctionalComponent(__VLS_107, new __VLS_107({
    label: (__VLS_ctx.t('user.created_at', 'Created at')),
    minWidth: "180",
}));
const __VLS_109 = __VLS_108({
    label: (__VLS_ctx.t('user.created_at', 'Created at')),
    minWidth: "180",
}, ...__VLS_functionalComponentArgsRest(__VLS_108));
__VLS_110.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_110.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    (__VLS_ctx.formatDateTime(row.created_at));
}
var __VLS_110;
const __VLS_111 = {}.ElTableColumn;
/** @type {[typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, typeof __VLS_components.ElTableColumn, typeof __VLS_components.elTableColumn, ]} */ ;
// @ts-ignore
const __VLS_112 = __VLS_asFunctionalComponent(__VLS_111, new __VLS_111({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}));
const __VLS_113 = __VLS_112({
    label: (__VLS_ctx.t('common.actions', 'Actions')),
    width: "180",
    fixed: "right",
}, ...__VLS_functionalComponentArgsRest(__VLS_112));
__VLS_114.slots.default;
{
    const { default: __VLS_thisSlot } = __VLS_114.slots;
    const [{ row }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_115 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_116 = __VLS_asFunctionalComponent(__VLS_115, new __VLS_115({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }));
    const __VLS_117 = __VLS_116({
        ...{ 'onClick': {} },
        link: true,
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_116));
    let __VLS_119;
    let __VLS_120;
    let __VLS_121;
    const __VLS_122 = {
        onClick: (...[$event]) => {
            __VLS_ctx.openEdit(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('user:update') }, null, null);
    __VLS_118.slots.default;
    (__VLS_ctx.t('common.edit', 'Edit'));
    var __VLS_118;
    const __VLS_123 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_124 = __VLS_asFunctionalComponent(__VLS_123, new __VLS_123({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }));
    const __VLS_125 = __VLS_124({
        ...{ 'onClick': {} },
        link: true,
        type: "danger",
    }, ...__VLS_functionalComponentArgsRest(__VLS_124));
    let __VLS_127;
    let __VLS_128;
    let __VLS_129;
    const __VLS_130 = {
        onClick: (...[$event]) => {
            __VLS_ctx.removeRow(row);
        }
    };
    __VLS_asFunctionalDirective(__VLS_directives.vPermission)(null, { ...__VLS_directiveBindingRestFields, value: ('user:delete') }, null, null);
    __VLS_126.slots.default;
    (__VLS_ctx.t('common.delete', 'Delete'));
    var __VLS_126;
}
var __VLS_114;
var __VLS_70;
{
    const { footer: __VLS_thisSlot } = __VLS_2.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "admin-pagination" },
    });
    const __VLS_131 = {}.ElPagination;
    /** @type {[typeof __VLS_components.ElPagination, typeof __VLS_components.elPagination, ]} */ ;
    // @ts-ignore
    const __VLS_132 = __VLS_asFunctionalComponent(__VLS_131, new __VLS_131({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }));
    const __VLS_133 = __VLS_132({
        ...{ 'onCurrentChange': {} },
        ...{ 'onSizeChange': {} },
        background: true,
        layout: "total, sizes, prev, pager, next, jumper",
        total: (__VLS_ctx.total),
        currentPage: (__VLS_ctx.query.page),
        pageSize: (__VLS_ctx.query.page_size),
        pageSizes: ([10, 20, 50, 100]),
    }, ...__VLS_functionalComponentArgsRest(__VLS_132));
    let __VLS_135;
    let __VLS_136;
    let __VLS_137;
    const __VLS_138 = {
        onCurrentChange: (__VLS_ctx.handlePageChange)
    };
    const __VLS_139 = {
        onSizeChange: (__VLS_ctx.handleSizeChange)
    };
    var __VLS_134;
}
var __VLS_2;
/** @type {[typeof AdminFormDialog, typeof AdminFormDialog, ]} */ ;
// @ts-ignore
const __VLS_140 = __VLS_asFunctionalComponent(AdminFormDialog, new AdminFormDialog({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('user.edit_title', 'Edit user') : __VLS_ctx.t('user.create_title', 'New user')),
    loading: (__VLS_ctx.dialogLoading),
}));
const __VLS_141 = __VLS_140({
    ...{ 'onConfirm': {} },
    modelValue: (__VLS_ctx.dialogVisible),
    title: (__VLS_ctx.editingId ? __VLS_ctx.t('user.edit_title', 'Edit user') : __VLS_ctx.t('user.create_title', 'New user')),
    loading: (__VLS_ctx.dialogLoading),
}, ...__VLS_functionalComponentArgsRest(__VLS_140));
let __VLS_143;
let __VLS_144;
let __VLS_145;
const __VLS_146 = {
    onConfirm: (__VLS_ctx.submitForm)
};
__VLS_142.slots.default;
const __VLS_147 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_148 = __VLS_asFunctionalComponent(__VLS_147, new __VLS_147({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}));
const __VLS_149 = __VLS_148({
    labelWidth: "110px",
    ...{ class: "admin-form" },
}, ...__VLS_functionalComponentArgsRest(__VLS_148));
__VLS_150.slots.default;
const __VLS_151 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_152 = __VLS_asFunctionalComponent(__VLS_151, new __VLS_151({
    label: (__VLS_ctx.t('user.username', 'Username')),
    required: true,
}));
const __VLS_153 = __VLS_152({
    label: (__VLS_ctx.t('user.username', 'Username')),
    required: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_152));
__VLS_154.slots.default;
const __VLS_155 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_156 = __VLS_asFunctionalComponent(__VLS_155, new __VLS_155({
    modelValue: (__VLS_ctx.form.username),
    placeholder: (__VLS_ctx.t('user.username_placeholder', 'Enter username')),
}));
const __VLS_157 = __VLS_156({
    modelValue: (__VLS_ctx.form.username),
    placeholder: (__VLS_ctx.t('user.username_placeholder', 'Enter username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_156));
var __VLS_154;
const __VLS_159 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_160 = __VLS_asFunctionalComponent(__VLS_159, new __VLS_159({
    label: (__VLS_ctx.t('user.display_name', 'Display name')),
}));
const __VLS_161 = __VLS_160({
    label: (__VLS_ctx.t('user.display_name', 'Display name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_160));
__VLS_162.slots.default;
const __VLS_163 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_164 = __VLS_asFunctionalComponent(__VLS_163, new __VLS_163({
    modelValue: (__VLS_ctx.form.display_name),
    placeholder: (__VLS_ctx.t('user.display_name_placeholder', 'Enter display name')),
}));
const __VLS_165 = __VLS_164({
    modelValue: (__VLS_ctx.form.display_name),
    placeholder: (__VLS_ctx.t('user.display_name_placeholder', 'Enter display name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_164));
var __VLS_162;
const __VLS_167 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_168 = __VLS_asFunctionalComponent(__VLS_167, new __VLS_167({
    label: (__VLS_ctx.t('user.mobile', 'Mobile')),
}));
const __VLS_169 = __VLS_168({
    label: (__VLS_ctx.t('user.mobile', 'Mobile')),
}, ...__VLS_functionalComponentArgsRest(__VLS_168));
__VLS_170.slots.default;
const __VLS_171 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_172 = __VLS_asFunctionalComponent(__VLS_171, new __VLS_171({
    modelValue: (__VLS_ctx.form.mobile),
    placeholder: (__VLS_ctx.t('user.mobile_placeholder', 'Enter mobile number')),
}));
const __VLS_173 = __VLS_172({
    modelValue: (__VLS_ctx.form.mobile),
    placeholder: (__VLS_ctx.t('user.mobile_placeholder', 'Enter mobile number')),
}, ...__VLS_functionalComponentArgsRest(__VLS_172));
var __VLS_170;
const __VLS_175 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_176 = __VLS_asFunctionalComponent(__VLS_175, new __VLS_175({
    label: (__VLS_ctx.t('user.email', 'Email')),
}));
const __VLS_177 = __VLS_176({
    label: (__VLS_ctx.t('user.email', 'Email')),
}, ...__VLS_functionalComponentArgsRest(__VLS_176));
__VLS_178.slots.default;
const __VLS_179 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_180 = __VLS_asFunctionalComponent(__VLS_179, new __VLS_179({
    modelValue: (__VLS_ctx.form.email),
    placeholder: (__VLS_ctx.t('user.email_placeholder', 'Enter email')),
}));
const __VLS_181 = __VLS_180({
    modelValue: (__VLS_ctx.form.email),
    placeholder: (__VLS_ctx.t('user.email_placeholder', 'Enter email')),
}, ...__VLS_functionalComponentArgsRest(__VLS_180));
var __VLS_178;
const __VLS_183 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_184 = __VLS_asFunctionalComponent(__VLS_183, new __VLS_183({
    label: (__VLS_ctx.t('user.status', 'Status')),
}));
const __VLS_185 = __VLS_184({
    label: (__VLS_ctx.t('user.status', 'Status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_184));
__VLS_186.slots.default;
const __VLS_187 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_188 = __VLS_asFunctionalComponent(__VLS_187, new __VLS_187({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}));
const __VLS_189 = __VLS_188({
    modelValue: (__VLS_ctx.form.status),
    ...{ style: {} },
}, ...__VLS_functionalComponentArgsRest(__VLS_188));
__VLS_190.slots.default;
const __VLS_191 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_192 = __VLS_asFunctionalComponent(__VLS_191, new __VLS_191({
    label: (__VLS_ctx.t('user.status_active', 'Enabled')),
    value: "active",
}));
const __VLS_193 = __VLS_192({
    label: (__VLS_ctx.t('user.status_active', 'Enabled')),
    value: "active",
}, ...__VLS_functionalComponentArgsRest(__VLS_192));
const __VLS_195 = {}.ElOption;
/** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
// @ts-ignore
const __VLS_196 = __VLS_asFunctionalComponent(__VLS_195, new __VLS_195({
    label: (__VLS_ctx.t('user.status_inactive', 'Disabled')),
    value: "inactive",
}));
const __VLS_197 = __VLS_196({
    label: (__VLS_ctx.t('user.status_inactive', 'Disabled')),
    value: "inactive",
}, ...__VLS_functionalComponentArgsRest(__VLS_196));
var __VLS_190;
var __VLS_186;
const __VLS_199 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_200 = __VLS_asFunctionalComponent(__VLS_199, new __VLS_199({
    label: (__VLS_ctx.t('user.role', 'Roles')),
}));
const __VLS_201 = __VLS_200({
    label: (__VLS_ctx.t('user.role', 'Roles')),
}, ...__VLS_functionalComponentArgsRest(__VLS_200));
__VLS_202.slots.default;
const __VLS_203 = {}.ElSelect;
/** @type {[typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, typeof __VLS_components.ElSelect, typeof __VLS_components.elSelect, ]} */ ;
// @ts-ignore
const __VLS_204 = __VLS_asFunctionalComponent(__VLS_203, new __VLS_203({
    modelValue: (__VLS_ctx.form.role_codes),
    multiple: true,
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.roleLoading),
    placeholder: (__VLS_ctx.t('user.role_placeholder', 'Select roles')),
}));
const __VLS_205 = __VLS_204({
    modelValue: (__VLS_ctx.form.role_codes),
    multiple: true,
    clearable: true,
    filterable: true,
    loading: (__VLS_ctx.roleLoading),
    placeholder: (__VLS_ctx.t('user.role_placeholder', 'Select roles')),
}, ...__VLS_functionalComponentArgsRest(__VLS_204));
__VLS_206.slots.default;
for (const [role] of __VLS_getVForSourceType((__VLS_ctx.roleOptions))) {
    const __VLS_207 = {}.ElOption;
    /** @type {[typeof __VLS_components.ElOption, typeof __VLS_components.elOption, ]} */ ;
    // @ts-ignore
    const __VLS_208 = __VLS_asFunctionalComponent(__VLS_207, new __VLS_207({
        key: (role.id),
        label: (`${role.name} (${role.code})`),
        value: (role.code),
    }));
    const __VLS_209 = __VLS_208({
        key: (role.id),
        label: (`${role.name} (${role.code})`),
        value: (role.code),
    }, ...__VLS_functionalComponentArgsRest(__VLS_208));
}
var __VLS_206;
var __VLS_202;
const __VLS_211 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_212 = __VLS_asFunctionalComponent(__VLS_211, new __VLS_211({
    label: (__VLS_ctx.t('user.password_hash', 'Password hash')),
}));
const __VLS_213 = __VLS_212({
    label: (__VLS_ctx.t('user.password_hash', 'Password hash')),
}, ...__VLS_functionalComponentArgsRest(__VLS_212));
__VLS_214.slots.default;
const __VLS_215 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_216 = __VLS_asFunctionalComponent(__VLS_215, new __VLS_215({
    modelValue: (__VLS_ctx.form.password_hash),
    type: "password",
    showPassword: true,
    placeholder: (__VLS_ctx.t('user.password_hash_placeholder', 'Optional, leave blank to keep unchanged')),
}));
const __VLS_217 = __VLS_216({
    modelValue: (__VLS_ctx.form.password_hash),
    type: "password",
    showPassword: true,
    placeholder: (__VLS_ctx.t('user.password_hash_placeholder', 'Optional, leave blank to keep unchanged')),
}, ...__VLS_functionalComponentArgsRest(__VLS_216));
var __VLS_214;
var __VLS_150;
var __VLS_142;
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
            roleLoading: roleLoading,
            total: total,
            rows: rows,
            roleOptions: roleOptions,
            editingId: editingId,
            query: query,
            form: form,
            loadUsers: loadUsers,
            openCreate: openCreate,
            openEdit: openEdit,
            resolveRoleLabel: resolveRoleLabel,
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
