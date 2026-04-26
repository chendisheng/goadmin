import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { ArrowDown, Expand, Fold, RefreshRight, UserFilled } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { logout as logoutApi } from '@/api/auth';
import { resolveRouteLocaleMeta, useAppI18n } from '@/i18n';
import { useAppStore } from '@/store/app';
import { useMenuStore } from '@/store/menu';
import { useSessionStore } from '@/store/session';
import { useTabsStore } from '@/store/tabs';
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const buildMode = import.meta.env.MODE;
const appStore = useAppStore();
const menuStore = useMenuStore();
const sessionStore = useSessionStore();
const tabsStore = useTabsStore();
const router = useRouter();
const route = useRoute();
const { t } = useAppI18n();
const pageTitle = computed(() => {
    const localized = resolveRouteLocaleMeta(route);
    return localized.title.trim() !== '' ? localized.title : t('app.title', appTitle);
});
const pageSubtitle = computed(() => {
    const localized = resolveRouteLocaleMeta(route);
    if (localized.subtitle.trim() !== '') {
        return localized.subtitle;
    }
    return t('app.subtitle', 'Vue 3 + TypeScript + Vite + Pinia + Axios');
});
const currentUserName = computed(() => sessionStore.displayName || t('common.visitor', '访客'));
const currentUserRole = computed(() => {
    const role = sessionStore.currentUser?.roles?.[0];
    return typeof role === 'string' && role.trim() !== '' ? role : t('common.admin_role', '管理员');
});
const currentUserInitial = computed(() => {
    const source = currentUserName.value.trim();
    if (source.length === 0) {
        return 'G';
    }
    return source.slice(0, 1).toUpperCase();
});
function refreshPage() {
    window.location.reload();
}
async function onLogout() {
    try {
        await logoutApi();
    }
    catch {
        // 退出时即使后端已失效也继续清理本地会话
    }
    finally {
        menuStore.clear(router);
        tabsStore.clearTabs();
        sessionStore.clearSession();
        ElMessage.success(t('common.logged_out', '已退出登录'));
        await router.push({ path: '/login' });
    }
}
function onCommand(command) {
    if (command === 'refresh') {
        refreshPage();
    }
    if (command === 'logout') {
        void onLogout();
    }
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElHeader;
/** @type {[typeof __VLS_components.ElHeader, typeof __VLS_components.elHeader, typeof __VLS_components.ElHeader, typeof __VLS_components.elHeader, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "app-header" },
}));
const __VLS_2 = __VLS_1({
    ...{ class: "app-header" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-header__left" },
});
const __VLS_5 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
    ...{ 'onClick': {} },
    ...{ class: "app-header__toggle" },
    circle: true,
    text: true,
}));
const __VLS_7 = __VLS_6({
    ...{ 'onClick': {} },
    ...{ class: "app-header__toggle" },
    circle: true,
    text: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_6));
let __VLS_9;
let __VLS_10;
let __VLS_11;
const __VLS_12 = {
    onClick: (...[$event]) => {
        __VLS_ctx.appStore.toggleSidebar();
    }
};
__VLS_8.slots.default;
const __VLS_13 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({}));
const __VLS_15 = __VLS_14({}, ...__VLS_functionalComponentArgsRest(__VLS_14));
__VLS_16.slots.default;
if (!__VLS_ctx.appStore.sidebarCollapsed) {
    const __VLS_17 = {}.Fold;
    /** @type {[typeof __VLS_components.Fold, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({}));
    const __VLS_19 = __VLS_18({}, ...__VLS_functionalComponentArgsRest(__VLS_18));
}
else {
    const __VLS_21 = {}.Expand;
    /** @type {[typeof __VLS_components.Expand, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({}));
    const __VLS_23 = __VLS_22({}, ...__VLS_functionalComponentArgsRest(__VLS_22));
}
var __VLS_16;
var __VLS_8;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-header__titles" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
(__VLS_ctx.pageTitle);
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.pageSubtitle);
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-header__right" },
});
const __VLS_25 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
    effect: "plain",
    round: true,
    type: "info",
}));
const __VLS_27 = __VLS_26({
    effect: "plain",
    round: true,
    type: "info",
}, ...__VLS_functionalComponentArgsRest(__VLS_26));
__VLS_28.slots.default;
(__VLS_ctx.buildMode);
var __VLS_28;
const __VLS_29 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
    effect: "plain",
    round: true,
    type: "success",
}));
const __VLS_31 = __VLS_30({
    effect: "plain",
    round: true,
    type: "success",
}, ...__VLS_functionalComponentArgsRest(__VLS_30));
__VLS_32.slots.default;
(__VLS_ctx.apiBaseUrl);
var __VLS_32;
const __VLS_33 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
    ...{ 'onClick': {} },
    circle: true,
    text: true,
}));
const __VLS_35 = __VLS_34({
    ...{ 'onClick': {} },
    circle: true,
    text: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_34));
let __VLS_37;
let __VLS_38;
let __VLS_39;
const __VLS_40 = {
    onClick: (__VLS_ctx.refreshPage)
};
__VLS_36.slots.default;
const __VLS_41 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({}));
const __VLS_43 = __VLS_42({}, ...__VLS_functionalComponentArgsRest(__VLS_42));
__VLS_44.slots.default;
const __VLS_45 = {}.RefreshRight;
/** @type {[typeof __VLS_components.RefreshRight, ]} */ ;
// @ts-ignore
const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({}));
const __VLS_47 = __VLS_46({}, ...__VLS_functionalComponentArgsRest(__VLS_46));
var __VLS_44;
var __VLS_36;
const __VLS_49 = {}.ElDropdown;
/** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
// @ts-ignore
const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
    ...{ 'onCommand': {} },
    trigger: "click",
}));
const __VLS_51 = __VLS_50({
    ...{ 'onCommand': {} },
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_50));
let __VLS_53;
let __VLS_54;
let __VLS_55;
const __VLS_56 = {
    onCommand: (__VLS_ctx.onCommand)
};
__VLS_52.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
    ...{ class: "app-header__user" },
    type: "button",
});
const __VLS_57 = {}.ElAvatar;
/** @type {[typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, typeof __VLS_components.ElAvatar, typeof __VLS_components.elAvatar, ]} */ ;
// @ts-ignore
const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
    ...{ class: "app-header__avatar" },
    size: (32),
}));
const __VLS_59 = __VLS_58({
    ...{ class: "app-header__avatar" },
    size: (32),
}, ...__VLS_functionalComponentArgsRest(__VLS_58));
__VLS_60.slots.default;
(__VLS_ctx.currentUserInitial);
var __VLS_60;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
    ...{ class: "app-header__user-text" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.currentUserName);
__VLS_asFunctionalElement(__VLS_intrinsicElements.small, __VLS_intrinsicElements.small)({});
(__VLS_ctx.currentUserRole);
const __VLS_61 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
    ...{ class: "app-header__user-arrow" },
}));
const __VLS_63 = __VLS_62({
    ...{ class: "app-header__user-arrow" },
}, ...__VLS_functionalComponentArgsRest(__VLS_62));
__VLS_64.slots.default;
const __VLS_65 = {}.ArrowDown;
/** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
// @ts-ignore
const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({}));
const __VLS_67 = __VLS_66({}, ...__VLS_functionalComponentArgsRest(__VLS_66));
var __VLS_64;
{
    const { dropdown: __VLS_thisSlot } = __VLS_52.slots;
    const __VLS_69 = {}.ElDropdownMenu;
    /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
    // @ts-ignore
    const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({}));
    const __VLS_71 = __VLS_70({}, ...__VLS_functionalComponentArgsRest(__VLS_70));
    __VLS_72.slots.default;
    const __VLS_73 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
        disabled: true,
    }));
    const __VLS_75 = __VLS_74({
        disabled: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_74));
    __VLS_76.slots.default;
    const __VLS_77 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({}));
    const __VLS_79 = __VLS_78({}, ...__VLS_functionalComponentArgsRest(__VLS_78));
    __VLS_80.slots.default;
    const __VLS_81 = {}.UserFilled;
    /** @type {[typeof __VLS_components.UserFilled, ]} */ ;
    // @ts-ignore
    const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({}));
    const __VLS_83 = __VLS_82({}, ...__VLS_functionalComponentArgsRest(__VLS_82));
    var __VLS_80;
    (__VLS_ctx.t('common.personal_center', '个人中心'));
    var __VLS_76;
    const __VLS_85 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
        command: "refresh",
    }));
    const __VLS_87 = __VLS_86({
        command: "refresh",
    }, ...__VLS_functionalComponentArgsRest(__VLS_86));
    __VLS_88.slots.default;
    (__VLS_ctx.t('common.refresh_page', '刷新页面'));
    var __VLS_88;
    const __VLS_89 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
        command: "logout",
        divided: true,
    }));
    const __VLS_91 = __VLS_90({
        command: "logout",
        divided: true,
    }, ...__VLS_functionalComponentArgsRest(__VLS_90));
    __VLS_92.slots.default;
    (__VLS_ctx.t('common.logout', '退出登录'));
    var __VLS_92;
    var __VLS_72;
}
var __VLS_52;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['app-header']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__left']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__toggle']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__titles']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__right']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__user']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__avatar']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__user-text']} */ ;
/** @type {__VLS_StyleScopedClasses['app-header__user-arrow']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ArrowDown: ArrowDown,
            Expand: Expand,
            Fold: Fold,
            RefreshRight: RefreshRight,
            UserFilled: UserFilled,
            apiBaseUrl: apiBaseUrl,
            buildMode: buildMode,
            appStore: appStore,
            t: t,
            pageTitle: pageTitle,
            pageSubtitle: pageSubtitle,
            currentUserName: currentUserName,
            currentUserRole: currentUserRole,
            currentUserInitial: currentUserInitial,
            refreshPage: refreshPage,
            onCommand: onCommand,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
