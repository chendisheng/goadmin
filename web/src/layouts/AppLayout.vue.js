import { computed } from 'vue';
import AppHeader from './components/AppHeader.vue';
import AppSidebar from './components/AppSidebar.vue';
import TabsBar from './components/TabsBar.vue';
import { useTabsStore } from '@/store/tabs';
const tabsStore = useTabsStore();
const cachedViewNames = computed(() => tabsStore.cachedViewNames);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElContainer;
/** @type {[typeof __VLS_components.ElContainer, typeof __VLS_components.elContainer, typeof __VLS_components.ElContainer, typeof __VLS_components.elContainer, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "app-layout" },
}));
const __VLS_2 = __VLS_1({
    ...{ class: "app-layout" },
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
/** @type {[typeof AppSidebar, ]} */ ;
// @ts-ignore
const __VLS_5 = __VLS_asFunctionalComponent(AppSidebar, new AppSidebar({}));
const __VLS_6 = __VLS_5({}, ...__VLS_functionalComponentArgsRest(__VLS_5));
const __VLS_8 = {}.ElContainer;
/** @type {[typeof __VLS_components.ElContainer, typeof __VLS_components.elContainer, typeof __VLS_components.ElContainer, typeof __VLS_components.elContainer, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ class: "app-layout__content" },
    direction: "vertical",
}));
const __VLS_10 = __VLS_9({
    ...{ class: "app-layout__content" },
    direction: "vertical",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
__VLS_11.slots.default;
/** @type {[typeof AppHeader, ]} */ ;
// @ts-ignore
const __VLS_12 = __VLS_asFunctionalComponent(AppHeader, new AppHeader({}));
const __VLS_13 = __VLS_12({}, ...__VLS_functionalComponentArgsRest(__VLS_12));
/** @type {[typeof TabsBar, ]} */ ;
// @ts-ignore
const __VLS_15 = __VLS_asFunctionalComponent(TabsBar, new TabsBar({}));
const __VLS_16 = __VLS_15({}, ...__VLS_functionalComponentArgsRest(__VLS_15));
const __VLS_18 = {}.ElMain;
/** @type {[typeof __VLS_components.ElMain, typeof __VLS_components.elMain, typeof __VLS_components.ElMain, typeof __VLS_components.elMain, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(__VLS_18, new __VLS_18({
    ...{ class: "app-layout__main" },
}));
const __VLS_20 = __VLS_19({
    ...{ class: "app-layout__main" },
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
__VLS_21.slots.default;
const __VLS_22 = {}.RouterView;
/** @type {[typeof __VLS_components.RouterView, typeof __VLS_components.routerView, typeof __VLS_components.RouterView, typeof __VLS_components.routerView, ]} */ ;
// @ts-ignore
const __VLS_23 = __VLS_asFunctionalComponent(__VLS_22, new __VLS_22({}));
const __VLS_24 = __VLS_23({}, ...__VLS_functionalComponentArgsRest(__VLS_23));
{
    const { default: __VLS_thisSlot } = __VLS_25.slots;
    const [{ Component, route }] = __VLS_getSlotParams(__VLS_thisSlot);
    const __VLS_26 = {}.KeepAlive;
    /** @type {[typeof __VLS_components.KeepAlive, typeof __VLS_components.KeepAlive, ]} */ ;
    // @ts-ignore
    const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
        include: (__VLS_ctx.cachedViewNames),
    }));
    const __VLS_28 = __VLS_27({
        include: (__VLS_ctx.cachedViewNames),
    }, ...__VLS_functionalComponentArgsRest(__VLS_27));
    __VLS_29.slots.default;
    const __VLS_30 = ((Component));
    // @ts-ignore
    const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
        key: (route.fullPath),
    }));
    const __VLS_32 = __VLS_31({
        key: (route.fullPath),
    }, ...__VLS_functionalComponentArgsRest(__VLS_31));
    var __VLS_29;
    __VLS_25.slots['' /* empty slot name completion */];
}
var __VLS_25;
var __VLS_21;
var __VLS_11;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['app-layout']} */ ;
/** @type {__VLS_StyleScopedClasses['app-layout__content']} */ ;
/** @type {__VLS_StyleScopedClasses['app-layout__main']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            AppHeader: AppHeader,
            AppSidebar: AppSidebar,
            TabsBar: TabsBar,
            cachedViewNames: cachedViewNames,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
