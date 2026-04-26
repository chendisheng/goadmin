import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Close, MoreFilled, RefreshRight } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { useAppI18n } from '@/i18n';
import { useTabsStore } from '@/store/tabs';
const router = useRouter();
const tabsStore = useTabsStore();
const contextMenuVisible = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuTabId = ref('');
const tabItemRefs = new Map();
const { t } = useAppI18n();
const contextMenuTab = computed(() => tabsStore.tabs.find((tab) => tab.id === contextMenuTabId.value) ?? null);
const canCloseContextMenuTab = computed(() => contextMenuTab.value?.closable === true);
const canCloseContextMenuOthers = computed(() => (contextMenuTab.value ? hasClosableOthers(contextMenuTab.value.id) : false));
const canCloseContextMenuLeft = computed(() => (contextMenuTab.value ? hasClosableLeft(contextMenuTab.value.id) : false));
const canCloseContextMenuRight = computed(() => (contextMenuTab.value ? hasClosableRight(contextMenuTab.value.id) : false));
const canCloseContextMenuAll = computed(() => tabsStore.tabs.some((tab) => !tab.fixed));
function getTabTitle(tab) {
    return t(tab.titleKey || '', tab.titleDefault || tab.title || t('tabs.page', '页面'));
}
function isActive(tabId) {
    return tabsStore.activeId === tabId;
}
function setTabRef(tabId, element) {
    if (element) {
        tabItemRefs.set(tabId, element);
        return;
    }
    tabItemRefs.delete(tabId);
}
async function scrollTabIntoView(tabId) {
    await nextTick();
    const element = tabItemRefs.get(tabId);
    element?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' });
}
function clampMenuPosition(x, y) {
    const menuWidth = 196;
    const menuHeight = 240;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    return {
        x: Math.max(8, Math.min(x, viewportWidth - menuWidth)),
        y: Math.max(8, Math.min(y, viewportHeight - menuHeight)),
    };
}
function openContextMenu(tabId, x, y) {
    const position = clampMenuPosition(x, y);
    contextMenuTabId.value = tabId;
    contextMenuX.value = position.x;
    contextMenuY.value = position.y;
    contextMenuVisible.value = true;
}
function closeContextMenu() {
    contextMenuVisible.value = false;
}
function onTabContextMenu(event, tabId) {
    event.preventDefault();
    event.stopPropagation();
    openContextMenu(tabId, event.clientX, event.clientY);
}
function onTabMenuButtonClick(event, tabId) {
    event.stopPropagation();
    openContextMenu(tabId, event.clientX, event.clientY);
}
async function onTabClick(tabId) {
    closeContextMenu();
    const target = tabsStore.tabs.find((tab) => tab.id === tabId);
    if (!target) {
        return;
    }
    if (router.currentRoute.value.fullPath === target.routeFullPath) {
        tabsStore.setActiveTab(tabId);
        return;
    }
    tabsStore.setActiveTab(tabId);
    await router.push(target.routeFullPath);
}
async function onTabClose(tabId) {
    closeContextMenu();
    const closedTab = tabsStore.closeTab(tabId);
    if (!closedTab) {
        ElMessage.warning(t('tabs.not_closable', '该标签页不可关闭'));
        return;
    }
    const nextTab = tabsStore.activeTab;
    if (nextTab && router.currentRoute.value.fullPath !== nextTab.routeFullPath) {
        await router.push(nextTab.routeFullPath);
    }
}
async function onRefreshTab(tabId) {
    const target = tabsStore.tabs.find((tab) => tab.id === tabId);
    if (!target) {
        closeContextMenu();
        return;
    }
    tabsStore.setActiveTab(tabId);
    if (router.currentRoute.value.fullPath !== target.routeFullPath) {
        await router.push(target.routeFullPath);
    }
    tabsStore.refreshTab(tabId);
    closeContextMenu();
    window.location.reload();
}
async function syncRouteAfterMutation(previousFullPath) {
    const activeTab = tabsStore.activeTab;
    if (activeTab && router.currentRoute.value.fullPath !== activeTab.routeFullPath) {
        await router.push(activeTab.routeFullPath);
        return;
    }
    if (!activeTab && previousFullPath !== '/dashboard' && router.currentRoute.value.path !== '/dashboard') {
        await router.push('/dashboard');
    }
}
async function onTabCommand(tabId, command) {
    closeContextMenu();
    const previousFullPath = router.currentRoute.value.fullPath;
    if (command === 'refresh') {
        await onRefreshTab(tabId);
        return;
    }
    if (command === 'close') {
        await onTabClose(tabId);
        return;
    }
    if (command === 'close-others') {
        tabsStore.closeOthers(tabId);
        await syncRouteAfterMutation(previousFullPath);
        return;
    }
    if (command === 'close-left') {
        tabsStore.closeTabsToLeft(tabId);
        await syncRouteAfterMutation(previousFullPath);
        return;
    }
    if (command === 'close-right') {
        tabsStore.closeTabsToRight(tabId);
        await syncRouteAfterMutation(previousFullPath);
        return;
    }
    if (command === 'close-all') {
        tabsStore.closeAll();
        await syncRouteAfterMutation(previousFullPath);
    }
}
function hasClosableLeft(tabId) {
    const index = tabsStore.tabs.findIndex((tab) => tab.id === tabId);
    if (index <= 0) {
        return false;
    }
    return tabsStore.tabs.slice(0, index).some((tab) => !tab.fixed);
}
function hasClosableRight(tabId) {
    const index = tabsStore.tabs.findIndex((tab) => tab.id === tabId);
    if (index < 0 || index >= tabsStore.tabs.length - 1) {
        return false;
    }
    return tabsStore.tabs.slice(index + 1).some((tab) => !tab.fixed);
}
function hasClosableOthers(tabId) {
    return tabsStore.tabs.some((tab) => !tab.fixed && tab.id !== tabId);
}
function syncContextMenuTab() {
    if (!contextMenuTabId.value) {
        return;
    }
    if (!tabsStore.tabs.some((tab) => tab.id === contextMenuTabId.value)) {
        closeContextMenu();
    }
}
function onGlobalInteraction() {
    closeContextMenu();
}
function onGlobalKeydown(event) {
    if (event.key === 'Escape') {
        closeContextMenu();
    }
}
watch(() => tabsStore.activeId, (tabId) => {
    if (tabId) {
        void scrollTabIntoView(tabId);
    }
    closeContextMenu();
}, { immediate: true });
watch(() => tabsStore.tabs.map((tab) => tab.id).join('|'), () => {
    syncContextMenuTab();
});
onMounted(() => {
    window.addEventListener('click', onGlobalInteraction);
    window.addEventListener('scroll', onGlobalInteraction, true);
    window.addEventListener('resize', onGlobalInteraction);
    window.addEventListener('keydown', onGlobalKeydown);
});
onBeforeUnmount(() => {
    window.removeEventListener('click', onGlobalInteraction);
    window.removeEventListener('scroll', onGlobalInteraction, true);
    window.removeEventListener('resize', onGlobalInteraction);
    window.removeEventListener('keydown', onGlobalKeydown);
});
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tabs-bar" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tabs-bar__scroll" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "tabs-bar__list" },
    role: "tablist",
    'aria-label': (__VLS_ctx.t('tabs.aria', '已打开页面')),
});
for (const [tab] of __VLS_getVForSourceType((__VLS_ctx.tabsStore.tabs))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onContextmenu: (...[$event]) => {
                __VLS_ctx.onTabContextMenu($event, tab.id);
            } },
        ...{ onClick: (...[$event]) => {
                void __VLS_ctx.onTabClick(tab.id);
            } },
        key: (tab.id),
        ref: ((element) => __VLS_ctx.setTabRef(tab.id, element)),
        ...{ class: "tabs-bar__item" },
        ...{ class: ({ 'is-active': __VLS_ctx.isActive(tab.id), 'is-fixed': tab.fixed }) },
        role: "tab",
        'aria-selected': (__VLS_ctx.isActive(tab.id)),
        title: (__VLS_ctx.getTabTitle(tab)),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
        ...{ class: "tabs-bar__title" },
    });
    (__VLS_ctx.getTabTitle(tab));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "tabs-bar__actions" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                __VLS_ctx.onTabMenuButtonClick($event, tab.id);
            } },
        ...{ class: "tabs-bar__more" },
        type: "button",
        'aria-label': (__VLS_ctx.t('tabs.more', '更多操作 {title}', { title: __VLS_ctx.getTabTitle(tab) })),
    });
    const __VLS_0 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({}));
    const __VLS_2 = __VLS_1({}, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    const __VLS_4 = {}.MoreFilled;
    /** @type {[typeof __VLS_components.MoreFilled, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({}));
    const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
    var __VLS_3;
    if (tab.closable) {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
            ...{ onClick: (...[$event]) => {
                    if (!(tab.closable))
                        return;
                    void __VLS_ctx.onTabClose(tab.id);
                } },
            ...{ class: "tabs-bar__close" },
            type: "button",
            'aria-label': (__VLS_ctx.t('tabs.close', '关闭 {title}', { title: __VLS_ctx.getTabTitle(tab) })),
        });
        const __VLS_8 = {}.ElIcon;
        /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
        // @ts-ignore
        const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({}));
        const __VLS_10 = __VLS_9({}, ...__VLS_functionalComponentArgsRest(__VLS_9));
        __VLS_11.slots.default;
        const __VLS_12 = {}.Close;
        /** @type {[typeof __VLS_components.Close, ]} */ ;
        // @ts-ignore
        const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({}));
        const __VLS_14 = __VLS_13({}, ...__VLS_functionalComponentArgsRest(__VLS_13));
        var __VLS_11;
    }
}
if (__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ onClick: () => { } },
        ...{ class: "tabs-bar__context-menu" },
        ...{ style: ({ left: `${__VLS_ctx.contextMenuX}px`, top: `${__VLS_ctx.contextMenuY}px` }) },
        role: "menu",
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'refresh');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
    });
    const __VLS_16 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
    const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
    __VLS_19.slots.default;
    const __VLS_20 = {}.RefreshRight;
    /** @type {[typeof __VLS_components.RefreshRight, ]} */ ;
    // @ts-ignore
    const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({}));
    const __VLS_22 = __VLS_21({}, ...__VLS_functionalComponentArgsRest(__VLS_21));
    var __VLS_19;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.refresh_current', '刷新当前页'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
        ...{ class: "tabs-bar__context-menu-divider" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'close');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
        disabled: (!__VLS_ctx.canCloseContextMenuTab),
    });
    const __VLS_24 = {}.ElIcon;
    /** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
    // @ts-ignore
    const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
    const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
    __VLS_27.slots.default;
    const __VLS_28 = {}.Close;
    /** @type {[typeof __VLS_components.Close, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
    const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
    var __VLS_27;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.close_current', '关闭当前'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'close-others');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
        disabled: (!__VLS_ctx.canCloseContextMenuOthers),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.close_others', '关闭其他'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'close-left');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
        disabled: (!__VLS_ctx.canCloseContextMenuLeft),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.close_left', '关闭左侧'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'close-right');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
        disabled: (!__VLS_ctx.canCloseContextMenuRight),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.close_right', '关闭右侧'));
    __VLS_asFunctionalElement(__VLS_intrinsicElements.button, __VLS_intrinsicElements.button)({
        ...{ onClick: (...[$event]) => {
                if (!(__VLS_ctx.contextMenuVisible && __VLS_ctx.contextMenuTab))
                    return;
                void __VLS_ctx.onTabCommand(__VLS_ctx.contextMenuTab.id, 'close-all');
            } },
        ...{ class: "tabs-bar__context-menu-item" },
        type: "button",
        disabled: (!__VLS_ctx.canCloseContextMenuAll),
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.t('common.close_all', '关闭全部'));
}
/** @type {__VLS_StyleScopedClasses['tabs-bar']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__scroll']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__list']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__title']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__actions']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__more']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__close']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-divider']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
/** @type {__VLS_StyleScopedClasses['tabs-bar__context-menu-item']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            Close: Close,
            MoreFilled: MoreFilled,
            RefreshRight: RefreshRight,
            tabsStore: tabsStore,
            contextMenuVisible: contextMenuVisible,
            contextMenuX: contextMenuX,
            contextMenuY: contextMenuY,
            t: t,
            contextMenuTab: contextMenuTab,
            canCloseContextMenuTab: canCloseContextMenuTab,
            canCloseContextMenuOthers: canCloseContextMenuOthers,
            canCloseContextMenuLeft: canCloseContextMenuLeft,
            canCloseContextMenuRight: canCloseContextMenuRight,
            canCloseContextMenuAll: canCloseContextMenuAll,
            getTabTitle: getTabTitle,
            isActive: isActive,
            setTabRef: setTabRef,
            onTabContextMenu: onTabContextMenu,
            onTabMenuButtonClick: onTabMenuButtonClick,
            onTabClick: onTabClick,
            onTabClose: onTabClose,
            onTabCommand: onTabCommand,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
