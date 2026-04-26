import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAppI18n, resolveRouteLocaleMeta } from '@/i18n';
const route = useRoute();
const router = useRouter();
const { t } = useAppI18n();
const pageTitle = computed(() => {
    const localized = resolveRouteLocaleMeta(route);
    return localized.title.trim() !== '' ? localized.title : t('common.placeholder_route', '页面占位');
});
const componentName = computed(() => String(route.meta.componentName || t('common.unknown', '未知')));
const routePermission = computed(() => String(route.meta.permission || '-'));
const routeLink = computed(() => String(route.meta.link || '-'));
const routePath = computed(() => route.path);
function goDashboard() {
    void router.push('/dashboard');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "route-placeholder-page" },
});
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "page-card" },
    shadow: "never",
}));
const __VLS_2 = __VLS_1({
    ...{ class: "page-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
{
    const { header: __VLS_thisSlot } = __VLS_3.slots;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "page-card__header" },
    });
    __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
    (__VLS_ctx.pageTitle);
    const __VLS_4 = {}.ElTag;
    /** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
        effect: "plain",
        round: true,
        type: "warning",
    }));
    const __VLS_6 = __VLS_5({
        effect: "plain",
        round: true,
        type: "warning",
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    __VLS_7.slots.default;
    (__VLS_ctx.t('common.dynamic_route', 'Dynamic Route'));
    var __VLS_7;
}
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "route-placeholder-page__body" },
});
const __VLS_8 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    title: (__VLS_ctx.t('route.placeholder.info', '该路由已由后端菜单驱动注册')),
    description: (__VLS_ctx.t('route.placeholder.description', '当前页面用于承接尚未实现的业务模块，占位逻辑会在后续 Phase 13/14 中替换为真实页面。')),
    type: "info",
    showIcon: true,
    closable: (false),
}));
const __VLS_10 = __VLS_9({
    title: (__VLS_ctx.t('route.placeholder.info', '该路由已由后端菜单驱动注册')),
    description: (__VLS_ctx.t('route.placeholder.description', '当前页面用于承接尚未实现的业务模块，占位逻辑会在后续 Phase 13/14 中替换为真实页面。')),
    type: "info",
    showIcon: true,
    closable: (false),
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
const __VLS_12 = {}.ElDescriptions;
/** @type {[typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, typeof __VLS_components.ElDescriptions, typeof __VLS_components.elDescriptions, ]} */ ;
// @ts-ignore
const __VLS_13 = __VLS_asFunctionalComponent(__VLS_12, new __VLS_12({
    column: (1),
    border: true,
    size: "small",
    ...{ class: "route-placeholder-page__meta" },
}));
const __VLS_14 = __VLS_13({
    column: (1),
    border: true,
    size: "small",
    ...{ class: "route-placeholder-page__meta" },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
const __VLS_16 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: (__VLS_ctx.t('route.placeholder.route_path', '路由路径')),
}));
const __VLS_18 = __VLS_17({
    label: (__VLS_ctx.t('route.placeholder.route_path', '路由路径')),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
(__VLS_ctx.routePath);
var __VLS_19;
const __VLS_20 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: (__VLS_ctx.t('route.placeholder.component_name', '组件标识')),
}));
const __VLS_22 = __VLS_21({
    label: (__VLS_ctx.t('route.placeholder.component_name', '组件标识')),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
(__VLS_ctx.componentName);
var __VLS_23;
const __VLS_24 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: (__VLS_ctx.t('route.placeholder.permission', '权限标识')),
}));
const __VLS_26 = __VLS_25({
    label: (__VLS_ctx.t('route.placeholder.permission', '权限标识')),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
(__VLS_ctx.routePermission);
var __VLS_27;
const __VLS_28 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    label: (__VLS_ctx.t('route.placeholder.link', '外链地址')),
}));
const __VLS_30 = __VLS_29({
    label: (__VLS_ctx.t('route.placeholder.link', '外链地址')),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
__VLS_31.slots.default;
(__VLS_ctx.routeLink);
var __VLS_31;
var __VLS_15;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "route-placeholder-page__actions" },
});
const __VLS_32 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
    ...{ 'onClick': {} },
    type: "primary",
}));
const __VLS_34 = __VLS_33({
    ...{ 'onClick': {} },
    type: "primary",
}, ...__VLS_functionalComponentArgsRest(__VLS_33));
let __VLS_36;
let __VLS_37;
let __VLS_38;
const __VLS_39 = {
    onClick: (__VLS_ctx.goDashboard)
};
__VLS_35.slots.default;
(__VLS_ctx.t('route.placeholder.back', '返回工作台'));
var __VLS_35;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['route-placeholder-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['route-placeholder-page__body']} */ ;
/** @type {__VLS_StyleScopedClasses['route-placeholder-page__meta']} */ ;
/** @type {__VLS_StyleScopedClasses['route-placeholder-page__actions']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            pageTitle: pageTitle,
            componentName: componentName,
            routePermission: routePermission,
            routeLink: routeLink,
            routePath: routePath,
            goDashboard: goDashboard,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
