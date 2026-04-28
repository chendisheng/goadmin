import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { fetchHealth } from '@/api';
import { useAppI18n } from '@/i18n';
import { useAppStore } from '@/store/app';
import { useSessionStore } from '@/store/session';
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const appStore = useAppStore();
const sessionStore = useSessionStore();
const { t } = useAppI18n();
const healthState = ref(null);
const loading = ref(false);
const errorMessage = ref('');
const shellStatus = computed(() => (appStore.sidebarCollapsed ? t('dashboard.sidebar_collapsed', 'Sidebar collapsed') : t('dashboard.sidebar_expanded', 'Sidebar expanded')));
const currentUser = computed(() => sessionStore.currentUser);
const dashboardMetrics = computed(() => [
    {
        label: t('dashboard.metric.api_base_url', 'API base URL'),
        value: apiBaseUrl,
        note: t('dashboard.metric.api_base_url_note', 'Unified Axios request entry'),
    },
    {
        label: t('dashboard.metric.layout_state', 'Layout state'),
        value: shellStatus.value,
        note: t('dashboard.metric.layout_state_note', 'Sidebar collapse state persisted'),
    },
    {
        label: t('dashboard.metric.current_user', 'Current user'),
        value: sessionStore.displayName || t('dashboard.metric.default_user', 'System administrator'),
        note: t('dashboard.metric.current_user_note', 'Session information loaded'),
    },
    {
        label: t('dashboard.metric.login_mode', 'Login mode'),
        value: t('dashboard.metric.login_mode_value', 'JWT / RBAC'),
        note: t('dashboard.metric.login_mode_note', 'Button and menu permissions will be extended'),
    },
]);
async function onPingHealth() {
    loading.value = true;
    errorMessage.value = '';
    try {
        healthState.value = await fetchHealth();
        ElMessage.success(t('dashboard.health_success', 'Health check request succeeded'));
    }
    catch (error) {
        const message = error instanceof Error ? error.message : t('dashboard.health_failed', 'Health check request failed');
        errorMessage.value = message;
        ElMessage.error(message);
    }
    finally {
        loading.value = false;
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "dashboard-page" },
});
const __VLS_0 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    gutter: (16),
    ...{ class: "dashboard-metrics" },
}));
const __VLS_2 = __VLS_1({
    gutter: (16),
    ...{ class: "dashboard-metrics" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
for (const [metric] of __VLS_getVForSourceType((__VLS_ctx.dashboardMetrics))) {
    const __VLS_4 = {}.ElCol;
    /** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
        key: (metric.label),
        xs: (24),
        sm: (12),
        lg: (6),
    }));
    const __VLS_6 = __VLS_5({
        key: (metric.label),
        xs: (24),
        sm: (12),
        lg: (6),
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    __VLS_7.slots.default;
    const __VLS_8 = {}.ElCard;
    /** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
    // @ts-ignore
    const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
        ...{ class: "page-card dashboard-metric-card" },
        shadow: "never",
    }));
    const __VLS_10 = __VLS_9({
        ...{ class: "page-card dashboard-metric-card" },
        shadow: "never",
    }, ...__VLS_functionalComponentArgsRest(__VLS_9));
    __VLS_11.slots.default;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dashboard-metric-card__label" },
    });
    (metric.label);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dashboard-metric-card__value" },
    });
    (metric.value);
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dashboard-metric-card__note" },
    });
    (metric.note);
    var __VLS_11;
    var __VLS_7;
}
var __VLS_3;
const __VLS_12 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    gutter: (16),
}));
const __VLS_14 = __VLS_13({
    gutter: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
const __VLS_16 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    xs: (24),
    lg: (16),
}));
const __VLS_18 = __VLS_17({
    xs: (24),
    lg: (16),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
const __VLS_20 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    ...{ class: "page-card dashboard-hero" },
    shadow: "never",
}));
const __VLS_22 = __VLS_21({
    ...{ class: "page-card dashboard-hero" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_23.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('dashboard.hero_title', 'System overview'));
    const __VLS_24 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
        effect: "plain",
        round: true,
        type: "success",
    }));
    const __VLS_26 = __VLS_25({
        effect: "plain",
        round: true,
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    (__VLS_ctx.t('dashboard.status_online', 'Online'));
    var __VLS_27;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "dashboard-hero__content" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('dashboard.hero_heading', '{title} admin console', { title: __VLS_ctx.appTitle }));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('dashboard.hero_description', 'Unified sidebar, top navigation, and dashboard home, ready to host Auth, CRUD, and Plugin modules later.'));
const __VLS_28 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    column: (1),
    border: true,
    size: "small",
}));
const __VLS_30 = __VLS_29({
    column: (1),
    border: true,
    size: "small",
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
const __VLS_32 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    label: (__VLS_ctx.t('dashboard.hero_app_title', 'Application title')),
}));
const __VLS_34 = __VLS_33({
    label: (__VLS_ctx.t('dashboard.hero_app_title', 'Application title')),
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
__VLS_35.slots.default;
(__VLS_ctx.appTitle);
var __VLS_35;
const __VLS_36 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
    label: (__VLS_ctx.t('dashboard.hero_api_base_url', 'API base URL')),
}));
const __VLS_38 = __VLS_37({
    label: (__VLS_ctx.t('dashboard.hero_api_base_url', 'API base URL')),
}, ...__VLS_functionalComponentArgsRest(__VLS_37));
__VLS_39.slots.default;
(__VLS_ctx.apiBaseUrl);
var __VLS_39;
const __VLS_40 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    label: (__VLS_ctx.t('dashboard.hero_layout_state', 'Layout state')),
}));
const __VLS_42 = __VLS_41({
    label: (__VLS_ctx.t('dashboard.hero_layout_state', 'Layout state')),
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
__VLS_43.slots.default;
(__VLS_ctx.shellStatus);
var __VLS_43;
const __VLS_44 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
    label: (__VLS_ctx.t('dashboard.hero_current_user', 'Current user')),
}));
const __VLS_46 = __VLS_45({
    label: (__VLS_ctx.t('dashboard.hero_current_user', 'Current user')),
}, ...__VLS_functionalComponentArgsRest(__VLS_45));
__VLS_47.slots.default;
(__VLS_ctx.sessionStore.displayName);
var __VLS_47;
var __VLS_31;
var __VLS_23;
var __VLS_19;
const __VLS_48 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
    xs: (24),
    lg: (8),
}));
const __VLS_50 = __VLS_49({
    xs: (24),
    lg: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_49));
__VLS_51.slots.default;
const __VLS_52 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
    ...{ class: "page-card dashboard-quick-actions" },
    shadow: "never",
}));
const __VLS_54 = __VLS_53({
    ...{ class: "page-card dashboard-quick-actions" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_53));
__VLS_55.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_55.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('dashboard.quick_actions_title', 'API validation'));
    const __VLS_56 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_57 = __VLS_asFunctionalComponent(__VLS_56, new __VLS_56({
        effect: "plain",
        round: true,
        type: "info",
    }));
    const __VLS_58 = __VLS_57({
        effect: "plain",
        round: true,
        type: "info",
    }, ...__VLS_functionalComponentArgsRest(__VLS_57));
    __VLS_59.slots.default;
    (__VLS_ctx.t('dashboard.quick_actions_tag', 'Connectivity'));
    var __VLS_59;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "dashboard-quick-actions__body" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('dashboard.quick_actions_description', 'Click the button to send a health check request and quickly verify frontend-backend connectivity and Axios interceptors.'));
const __VLS_60 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.loading),
}));
const __VLS_62 = __VLS_61({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_61));
let __VLS_64;
let __VLS_65;
let __VLS_66;
const __VLS_67 = {
    onClick: (__VLS_ctx.onPingHealth)
};
__VLS_63.slots.default;
(__VLS_ctx.t('dashboard.health_check_button', 'Send health check'));
var __VLS_63;
const __VLS_68 = {}.ElDivider;
/** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
// @ts-ignore
const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({}));
const __VLS_70 = __VLS_69({}, ...__VLS_functionalComponentArgsRest(__VLS_69));
if (__VLS_ctx.healthState) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "dashboard-health-result" },
    });
    const __VLS_72 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
        column: (1),
        size: "small",
    }));
    const __VLS_74 = __VLS_73({
        column: (1),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_73));
    __VLS_75.slots.default;
    const __VLS_76 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
        label: (__VLS_ctx.t('dashboard.health_status', 'status')),
    }));
    const __VLS_78 = __VLS_77({
        label: (__VLS_ctx.t('dashboard.health_status', 'status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_77));
    __VLS_79.slots.default;
    (__VLS_ctx.healthState.status);
    var __VLS_79;
    const __VLS_80 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
        label: (__VLS_ctx.t('dashboard.health_uptime', 'uptime')),
    }));
    const __VLS_82 = __VLS_81({
        label: (__VLS_ctx.t('dashboard.health_uptime', 'uptime')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_81));
    __VLS_83.slots.default;
    (__VLS_ctx.healthState.uptime);
    var __VLS_83;
    const __VLS_84 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_85 = __VLS_asFunctionalComponent(__VLS_84, new __VLS_84({
        label: (__VLS_ctx.t('dashboard.health_timestamp', 'timestamp')),
    }));
    const __VLS_86 = __VLS_85({
        label: (__VLS_ctx.t('dashboard.health_timestamp', 'timestamp')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_85));
    __VLS_87.slots.default;
    (__VLS_ctx.healthState.timestamp);
    var __VLS_87;
    var __VLS_75;
}
if (__VLS_ctx.errorMessage) {
    const __VLS_88 = {}.ElAlert;
    /** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
    // @ts-ignore
    const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
        title: (__VLS_ctx.errorMessage),
        type: "error",
        showIcon: true,
        closable: (false),
    }));
    const __VLS_90 = __VLS_89({
        title: (__VLS_ctx.errorMessage),
        type: "error",
        showIcon: true,
        closable: (false),
    }, ...__VLS_functionalComponentArgsRest(__VLS_89));
}
if (__VLS_ctx.currentUser) {
    const __VLS_92 = {}.ElDivider;
    /** @type {[typeof __VLS_components.ElDivider, typeof __VLS_components.elDivider, ]} */ ;
    // @ts-ignore
    const __VLS_93 = __VLS_asFunctionalComponent(__VLS_92, new __VLS_92({}));
    const __VLS_94 = __VLS_93({}, ...__VLS_functionalComponentArgsRest(__VLS_93));
    const __VLS_96 = {}.ElDescriptions;
    /** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_97 = __VLS_asFunctionalComponent(__VLS_96, new __VLS_96({
        column: (1),
        size: "small",
    }));
    const __VLS_98 = __VLS_97({
        column: (1),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_97));
    __VLS_99.slots.default;
    const __VLS_100 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_101 = __VLS_asFunctionalComponent(__VLS_100, new __VLS_100({
        label: (__VLS_ctx.t('dashboard.user_username', 'username')),
    }));
    const __VLS_102 = __VLS_101({
        label: (__VLS_ctx.t('dashboard.user_username', 'username')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_101));
    __VLS_103.slots.default;
    (__VLS_ctx.currentUser.username);
    var __VLS_103;
    const __VLS_104 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_105 = __VLS_asFunctionalComponent(__VLS_104, new __VLS_104({
        label: (__VLS_ctx.t('dashboard.user_id', 'user_id')),
    }));
    const __VLS_106 = __VLS_105({
        label: (__VLS_ctx.t('dashboard.user_id', 'user_id')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_105));
    __VLS_107.slots.default;
    (__VLS_ctx.currentUser.user_id);
    var __VLS_107;
    const __VLS_108 = {}.ElDescriptionsItem;
    /** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_109 = __VLS_asFunctionalComponent(__VLS_108, new __VLS_108({
        label: (__VLS_ctx.t('dashboard.user_roles', 'roles')),
    }));
    const __VLS_110 = __VLS_109({
        label: (__VLS_ctx.t('dashboard.user_roles', 'roles')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_109));
    __VLS_111.slots.default;
    (__VLS_ctx.currentUser.roles?.join(', ') || '-');
    var __VLS_111;
    var __VLS_99;
}
var __VLS_55;
var __VLS_51;
var __VLS_15;
const __VLS_112 = {}.ElRow;
/** @type {[typeof __VLS_components.ElRow, typeof __VLS_components.elRow, typeof __VLS_components.ElRow, typeof __VLS_components.elRow, ]} */ ;
// @ts-ignore
const __VLS_113 = __VLS_asFunctionalComponent(__VLS_112, new __VLS_112({
    gutter: (16),
    ...{ class: "dashboard-secondary-row" },
}));
const __VLS_114 = __VLS_113({
    gutter: (16),
    ...{ class: "dashboard-secondary-row" },
}, ...__VLS_functionalComponentArgsRest(__VLS_113));
__VLS_115.slots.default;
const __VLS_116 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_117 = __VLS_asFunctionalComponent(__VLS_116, new __VLS_116({
    xs: (24),
    md: (8),
}));
const __VLS_118 = __VLS_117({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_117));
__VLS_119.slots.default;
const __VLS_120 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
    ...{ class: "page-card" },
    shadow: "never",
}));
const __VLS_122 = __VLS_121({
    ...{ class: "page-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_121));
__VLS_123.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_123.slots;
    (__VLS_ctx.t('dashboard.section.engineering', 'Engineering standards'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "dashboard-list" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.engineering.vue', 'Vue 3 + TypeScript'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.engineering.vite', 'Vite build and hot reload'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.engineering.element_plus', 'Unified Element Plus UI'));
var __VLS_123;
var __VLS_119;
const __VLS_124 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
    xs: (24),
    md: (8),
}));
const __VLS_126 = __VLS_125({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_125));
__VLS_127.slots.default;
const __VLS_128 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
    ...{ class: "page-card" },
    shadow: "never",
}));
const __VLS_130 = __VLS_129({
    ...{ class: "page-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_129));
__VLS_131.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_131.slots;
    (__VLS_ctx.t('dashboard.section.status', 'Status center'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "dashboard-list" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.status.pinia', 'Pinia global store initialized'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.status.sidebar_persisted', 'Sidebar collapse state persisted'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.status.token_reserved', 'Session token foundation reserved'));
var __VLS_131;
var __VLS_127;
const __VLS_132 = {}.ElCol;
/** @type {[typeof __VLS_components.ElCol, typeof __VLS_components.elCol, typeof __VLS_components.ElCol, typeof __VLS_components.elCol, ]} */ ;
// @ts-ignore
const __VLS_133 = __VLS_asFunctionalComponent(__VLS_132, new __VLS_132({
    xs: (24),
    md: (8),
}));
const __VLS_134 = __VLS_133({
    xs: (24),
    md: (8),
}, ...__VLS_functionalComponentArgsRest(__VLS_133));
__VLS_135.slots.default;
const __VLS_136 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_137 = __VLS_asFunctionalComponent(__VLS_136, new __VLS_136({
    ...{ class: "page-card" },
    shadow: "never",
}));
const __VLS_138 = __VLS_137({
    ...{ class: "page-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_137));
__VLS_139.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_139.slots;
    (__VLS_ctx.t('dashboard.section.plan', 'Feature roadmap'));
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "dashboard-list" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.plan.modules', 'Admin Modules base management page'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.plan.permissions', 'Permission control and button-level authorization'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('dashboard.plan.plugin_ui', 'Plugin UI and dynamic extension'));
var __VLS_139;
var __VLS_135;
var __VLS_115;
/** @type {__VLS_StyleScopedClasses['dashboard-page']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-metrics']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-metric-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-metric-card__label']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-metric-card__value']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-metric-card__note']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-hero']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-hero__content']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-quick-actions']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-quick-actions__body']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-health-result']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-secondary-row']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-list']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-list']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['dashboard-list']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            appTitle: appTitle,
            apiBaseUrl: apiBaseUrl,
            sessionStore: sessionStore,
            t: t,
            healthState: healthState,
            loading: loading,
            errorMessage: errorMessage,
            shellStatus: shellStatus,
            currentUser: currentUser,
            dashboardMetrics: dashboardMetrics,
            onPingHealth: onPingHealth,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
