import { useRoute } from 'vue-router';
import { useAppI18n } from '@/i18n';
import { useAppStore } from '@/store/app';
import { useMenuStore } from '@/store/menu';
import MenuTreeNode from './MenuTreeNode.vue';
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const { t } = useAppI18n();
const route = useRoute();
const appStore = useAppStore();
const menuStore = useMenuStore();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElAside;
/** @type {[typeof __VLS_components.ElAside, typeof __VLS_components.elAside, typeof __VLS_components.ElAside, typeof __VLS_components.elAside, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "app-sidebar" },
    width: (__VLS_ctx.appStore.sidebarCollapsed ? '72px' : '220px'),
}));
const __VLS_2 = __VLS_1({
    ...{ class: "app-sidebar" },
    width: (__VLS_ctx.appStore.sidebarCollapsed ? '72px' : '220px'),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-sidebar__brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-sidebar__logo" },
});
if (!__VLS_ctx.appStore.sidebarCollapsed) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "app-sidebar__brand-text" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
    (__VLS_ctx.t('app.title', __VLS_ctx.appTitle));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('app.subtitle', 'Frontend Core'));
}
const __VLS_5 = {}.ElScrollbar;
/** @type {[typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, typeof __VLS_components.ElScrollbar, typeof __VLS_components.elScrollbar, ]} */ ;
// @ts-ignore
const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
    ...{ class: "app-sidebar__scroll" },
}));
const __VLS_7 = __VLS_6({
    ...{ class: "app-sidebar__scroll" },
}, ...__VLS_functionalComponentArgsRest(__VLS_6));
__VLS_8.slots.default;
const __VLS_9 = {}.ElMenu;
/** @type {[typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, typeof __VLS_components.ElMenu, typeof __VLS_components.elMenu, ]} */ ;
// @ts-ignore
const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
    ...{ class: "app-sidebar__menu" },
    collapse: (__VLS_ctx.appStore.sidebarCollapsed),
    collapseTransition: (false),
    defaultActive: (__VLS_ctx.route.path),
    backgroundColor: "transparent",
    textColor: "inherit",
    activeTextColor: "var(--el-color-primary)",
    router: true,
}));
const __VLS_11 = __VLS_10({
    ...{ class: "app-sidebar__menu" },
    collapse: (__VLS_ctx.appStore.sidebarCollapsed),
    collapseTransition: (false),
    defaultActive: (__VLS_ctx.route.path),
    backgroundColor: "transparent",
    textColor: "inherit",
    activeTextColor: "var(--el-color-primary)",
    router: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_10));
__VLS_12.slots.default;
for (const [item] of __VLS_getVForSourceType((__VLS_ctx.menuStore.sidebarMenus))) {
    /** @type {[typeof MenuTreeNode, ]} */ ;
    // @ts-ignore
    const __VLS_13 = __VLS_asFunctionalComponent(MenuTreeNode, new MenuTreeNode({
        key: (item.path),
        node: (item),
    }));
    const __VLS_14 = __VLS_13({
        key: (item.path),
        node: (item),
    }, ...__VLS_functionalComponentArgsRest(__VLS_13));
}
var __VLS_12;
var __VLS_8;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "app-sidebar__footer" },
});
const __VLS_16 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ 'onClick': {} },
    ...{ class: "app-sidebar__toggle" },
    text: true,
}));
const __VLS_18 = __VLS_17({
    ...{ 'onClick': {} },
    ...{ class: "app-sidebar__toggle" },
    text: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
let __VLS_20;
let __VLS_21;
let __VLS_22;
const __VLS_23 = {
    onClick: (...[$event]) => {
        __VLS_ctx.appStore.toggleSidebar();
    }
};
__VLS_19.slots.default;
(__VLS_ctx.appStore.sidebarCollapsed ? __VLS_ctx.t('common.expand_sidebar', '展开侧栏') : __VLS_ctx.t('common.collapse_sidebar', '收起侧栏'));
var __VLS_19;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['app-sidebar']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__brand']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__logo']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__brand-text']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__scroll']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__menu']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__footer']} */ ;
/** @type {__VLS_StyleScopedClasses['app-sidebar__toggle']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            MenuTreeNode: MenuTreeNode,
            appTitle: appTitle,
            t: t,
            route: route,
            appStore: appStore,
            menuStore: menuStore,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
