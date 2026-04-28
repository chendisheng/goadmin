import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { fetchAuthorizationStatus, reloadAuthorizationPolicies, seedAuthorizationPolicies } from '@/api/casbin';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
const router = useRouter();
const loading = ref(false);
const actionLoading = ref(false);
const status = ref({});
const { t } = useAppI18n();
const statusTag = computed(() => (status.value.enabled ? 'success' : 'info'));
const statusText = computed(() => (status.value.enabled ? t('casbin.enabled', 'Enabled') : t('casbin.disabled', 'Disabled')));
async function loadStatus() {
    loading.value = true;
    try {
        status.value = await fetchAuthorizationStatus();
    }
    finally {
        loading.value = false;
    }
}
async function handleReload() {
    actionLoading.value = true;
    try {
        await reloadAuthorizationPolicies();
        ElMessage.success(t('casbin.reload_success', 'Authorization module reloaded'));
        await loadStatus();
    }
    finally {
        actionLoading.value = false;
    }
}
async function handleSeed() {
    actionLoading.value = true;
    try {
        await seedAuthorizationPolicies();
        ElMessage.success(t('casbin.seed_success', 'Default authorization policies have been seeded'));
        await loadStatus();
    }
    finally {
        actionLoading.value = false;
    }
}
function openModels() {
    void router.push('/system/casbin/models');
}
function openRules() {
    void router.push('/system/casbin/rules');
}
onMounted(() => {
    void loadStatus();
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
    title: (__VLS_ctx.t('casbin.title', 'Authorization governance')),
    description: (__VLS_ctx.t('casbin.description', 'Manage authorization runtime, default policies, and model/policy entry points.')),
    loading: (__VLS_ctx.loading),
}));
const __VLS_1 = __VLS_0({
    title: (__VLS_ctx.t('casbin.title', 'Authorization governance')),
    description: (__VLS_ctx.t('casbin.description', 'Manage authorization runtime, default policies, and model/policy entry points.')),
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
        loading: (__VLS_ctx.loading),
    }));
    const __VLS_5 = __VLS_4({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_4));
    let __VLS_7;
    let __VLS_8;
    let __VLS_9;
    const __VLS_10 = {
        onClick: (__VLS_ctx.loadStatus)
    };
    __VLS_6.slots.default;
    (__VLS_ctx.t('casbin.refresh_status', 'Refresh status'));
    var __VLS_6;
    const __VLS_11 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_12 = __VLS_asFunctionalComponent(__VLS_11, new __VLS_11({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.actionLoading),
        type: "primary",
    }));
    const __VLS_13 = __VLS_12({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.actionLoading),
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_12));
    let __VLS_15;
    let __VLS_16;
    let __VLS_17;
    const __VLS_18 = {
        onClick: (__VLS_ctx.handleReload)
    };
    __VLS_14.slots.default;
    (__VLS_ctx.t('casbin.reload_runtime', 'Reload runtime'));
    var __VLS_14;
    const __VLS_19 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_20 = __VLS_asFunctionalComponent(__VLS_19, new __VLS_19({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.actionLoading),
    }));
    const __VLS_21 = __VLS_20({
        ...{ 'onClick': {} },
        loading: (__VLS_ctx.actionLoading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_20));
    let __VLS_23;
    let __VLS_24;
    let __VLS_25;
    const __VLS_26 = {
        onClick: (__VLS_ctx.handleSeed)
    };
    __VLS_22.slots.default;
    (__VLS_ctx.t('casbin.seed_default', 'Seed default policies'));
    var __VLS_22;
    const __VLS_27 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_28 = __VLS_asFunctionalComponent(__VLS_27, new __VLS_27({
        ...{ 'onClick': {} },
    }));
    const __VLS_29 = __VLS_28({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_28));
    let __VLS_31;
    let __VLS_32;
    let __VLS_33;
    const __VLS_34 = {
        onClick: (__VLS_ctx.openModels)
    };
    __VLS_30.slots.default;
    (__VLS_ctx.t('casbin.models', 'Model management'));
    var __VLS_30;
    const __VLS_35 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_36 = __VLS_asFunctionalComponent(__VLS_35, new __VLS_35({
        ...{ 'onClick': {} },
    }));
    const __VLS_37 = __VLS_36({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_36));
    let __VLS_39;
    let __VLS_40;
    let __VLS_41;
    const __VLS_42 = {
        onClick: (__VLS_ctx.openRules)
    };
    __VLS_38.slots.default;
    (__VLS_ctx.t('casbin.rules', 'Policy management'));
    var __VLS_38;
}
const __VLS_43 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_44 = __VLS_asFunctionalComponent(__VLS_43, new __VLS_43({
    gutter: (16),
}));
const __VLS_45 = __VLS_44({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_44));
__VLS_46.slots.default;
const __VLS_47 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_48 = __VLS_asFunctionalComponent(__VLS_47, new __VLS_47({
    xs: (24),
    md: (12),
}));
const __VLS_49 = __VLS_48({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_48));
__VLS_50.slots.default;
const __VLS_51 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_52 = __VLS_asFunctionalComponent(__VLS_51, new __VLS_51({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_53 = __VLS_52({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_52));
__VLS_54.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_54.slots;
    (__VLS_ctx.t('casbin.status_panel', 'Runtime status'));
}
const __VLS_55 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_56 = __VLS_asFunctionalComponent(__VLS_55, new __VLS_55({
    column: (1),
    border: true,
}));
const __VLS_57 = __VLS_56({
    column: (1),
    border: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_56));
__VLS_58.slots.default;
const __VLS_59 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_60 = __VLS_asFunctionalComponent(__VLS_59, new __VLS_59({
    label: (__VLS_ctx.t('casbin.enabled_status', 'Enabled status')),
}));
const __VLS_61 = __VLS_60({
    label: (__VLS_ctx.t('casbin.enabled_status', 'Enabled status')),
}, ...__VLS_functionalComponentArgsRest(__VLS_60));
__VLS_62.slots.default;
const __VLS_63 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_64 = __VLS_asFunctionalComponent(__VLS_63, new __VLS_63({
    type: (__VLS_ctx.statusTag),
    effect: "plain",
}));
const __VLS_65 = __VLS_64({
    type: (__VLS_ctx.statusTag),
    effect: "plain",
}, ...__VLS_functionalComponentArgsRest(__VLS_64));
__VLS_66.slots.default;
(__VLS_ctx.statusText);
var __VLS_66;
var __VLS_62;
const __VLS_67 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_68 = __VLS_asFunctionalComponent(__VLS_67, new __VLS_67({
    label: (__VLS_ctx.t('casbin.source', 'Source')),
}));
const __VLS_69 = __VLS_68({
    label: (__VLS_ctx.t('casbin.source', 'Source')),
}, ...__VLS_functionalComponentArgsRest(__VLS_68));
__VLS_70.slots.default;
(__VLS_ctx.status.source || '-');
var __VLS_70;
const __VLS_71 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_72 = __VLS_asFunctionalComponent(__VLS_71, new __VLS_71({
    label: (__VLS_ctx.t('casbin.model_path', 'Model path')),
}));
const __VLS_73 = __VLS_72({
    label: (__VLS_ctx.t('casbin.model_path', 'Model path')),
}, ...__VLS_functionalComponentArgsRest(__VLS_72));
__VLS_74.slots.default;
(__VLS_ctx.status.model_path || '-');
var __VLS_74;
const __VLS_75 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_76 = __VLS_asFunctionalComponent(__VLS_75, new __VLS_75({
    label: (__VLS_ctx.t('casbin.policy_path', 'Policy path')),
}));
const __VLS_77 = __VLS_76({
    label: (__VLS_ctx.t('casbin.policy_path', 'Policy path')),
}, ...__VLS_functionalComponentArgsRest(__VLS_76));
__VLS_78.slots.default;
(__VLS_ctx.status.policy_path || '-');
var __VLS_78;
var __VLS_58;
var __VLS_54;
var __VLS_50;
const __VLS_79 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_80 = __VLS_asFunctionalComponent(__VLS_79, new __VLS_79({
    xs: (24),
    md: (12),
}));
const __VLS_81 = __VLS_80({
    xs: (24),
    md: (12),
}, ...__VLS_functionalComponentArgsRest(__VLS_80));
__VLS_82.slots.default;
const __VLS_83 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_84 = __VLS_asFunctionalComponent(__VLS_83, new __VLS_83({
    shadow: "never",
    ...{ class: "mb-16" },
}));
const __VLS_85 = __VLS_84({
    shadow: "never",
    ...{ class: "mb-16" },
}, ...__VLS_functionalComponentArgsRest(__VLS_84));
__VLS_86.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_86.slots;
    (__VLS_ctx.t('casbin.summary_title', 'Governance summary'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({
    ...{ class: "casbin-summary" },
});
(__VLS_ctx.status.summary || __VLS_ctx.t('casbin.no_summary', 'No summary available'));
const __VLS_87 = {}.ElDivider;
/** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
// @ts-ignore
const __VLS_88 = __VLS_asFunctionalComponent(__VLS_87, new __VLS_87({}));
const __VLS_89 = __VLS_88({}, ...__VLS_functionalComponentArgsRest(__VLS_88));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.t('casbin.legacy_modules', 'Linked entries'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "casbin-list" },
});
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.status.legacy_modules || []))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
        key: (item),
    });
    (item);
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.t('casbin.available_routes', 'Available endpoints'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "casbin-list" },
});
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.status.routes || []))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({
        key: (item),
    });
    (item);
}
var __VLS_86;
var __VLS_82;
var __VLS_46;
var __VLS_2;
/** @type {__VLS_StyleScopedClasses['admin-page']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['mb-16']} */ ;
/** @type {__VLS_StyleScopedClasses['casbin-summary']} */ ;
/** @type {__VLS_StyleScopedClasses['casbin-list']} */ ;
/** @type {__VLS_StyleScopedClasses['casbin-list']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AdminTable: AdminTable,
            loading: loading,
            actionLoading: actionLoading,
            status: status,
            t: t,
            statusTag: statusTag,
            statusText: statusText,
            loadStatus: loadStatus,
            handleReload: handleReload,
            handleSeed: handleSeed,
            openModels: openModels,
            openRules: openRules,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
