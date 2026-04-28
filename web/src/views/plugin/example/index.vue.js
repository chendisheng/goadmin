import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { pingExamplePlugin } from '@/api/plugins';
import { resolveRouteLocaleMeta, useAppI18n } from '@/i18n';
const route = useRoute();
const { t } = useAppI18n();
const loading = ref(false);
const pingResult = ref(null);
const pageTitle = computed(() => {
    const localized = resolveRouteLocaleMeta(route);
    return localized.title.trim() !== '' ? localized.title : t('plugin.example_title', 'Plugin example');
});
const componentName = computed(() => String(route.meta.componentName || 'view/plugin/example/index'));
const routePath = computed(() => route.path);
const routePermission = computed(() => String(route.meta.permission || 'plugin:example:view'));
async function handlePing() {
    loading.value = true;
    try {
        pingResult.value = await pingExamplePlugin();
        ElMessage.success(t('plugin.example_call_success', 'Plugin API call succeeded'));
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : t('plugin.example_call_failed', 'Plugin API call failed'));
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
    ...{ class: "plugin-example-page" },
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
        type: "success",
    }));
    const __VLS_6 = __VLS_5({
        effect: "plain",
        round: true,
        type: "success",
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    __VLS_7.slots.default;
    (__VLS_ctx.t('plugin.example_badge', 'Plugin UI'));
    var __VLS_7;
}
const __VLS_8 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    title: (__VLS_ctx.t('plugin.example_alert_title', 'This is a dynamic page registered by a plugin')),
    description: (__VLS_ctx.t('plugin.example_alert_description', 'The component path comes from backend menu configuration `view/plugin/example/index` and is loaded through the frontend dynamic component map.')),
    type: "success",
    showIcon: true,
    closable: (false),
}));
const __VLS_10 = __VLS_9({
    title: (__VLS_ctx.t('plugin.example_alert_title', 'This is a dynamic page registered by a plugin')),
    description: (__VLS_ctx.t('plugin.example_alert_description', 'The component path comes from backend menu configuration `view/plugin/example/index` and is loaded through the frontend dynamic component map.')),
    type: "success",
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
    ...{ class: "plugin-example-page__meta" },
}));
const __VLS_14 = __VLS_13({
    column: (1),
    border: true,
    size: "small",
    ...{ class: "plugin-example-page__meta" },
}, ...__VLS_functionalComponentArgsRest(__VLS_13));
__VLS_15.slots.default;
const __VLS_16 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    label: (__VLS_ctx.t('plugin.example_route_path', 'Route path')),
}));
const __VLS_18 = __VLS_17({
    label: (__VLS_ctx.t('plugin.example_route_path', 'Route path')),
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
(__VLS_ctx.routePath);
var __VLS_19;
const __VLS_20 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: (__VLS_ctx.t('plugin.example_component_name', 'Component name')),
}));
const __VLS_22 = __VLS_21({
    label: (__VLS_ctx.t('plugin.example_component_name', 'Component name')),
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
(__VLS_ctx.componentName);
var __VLS_23;
const __VLS_24 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: (__VLS_ctx.t('plugin.example_permission', 'Permission key')),
}));
const __VLS_26 = __VLS_25({
    label: (__VLS_ctx.t('plugin.example_permission', 'Permission key')),
}, ...__VLS_functionalComponentArgsRest(__VLS_25));
__VLS_27.slots.default;
(__VLS_ctx.routePermission);
var __VLS_27;
var __VLS_15;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "plugin-example-page__actions" },
});
const __VLS_28 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.loading),
}));
const __VLS_30 = __VLS_29({
    ...{ 'onClick': {} },
    type: "primary",
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_29));
let __VLS_32;
let __VLS_33;
let __VLS_34;
const __VLS_35 = {
    onClick: (__VLS_ctx.handlePing)
};
__VLS_31.slots.default;
(__VLS_ctx.t('plugin.example_call', 'Call plugin API'));
var __VLS_31;
if (__VLS_ctx.pingResult) {
    const __VLS_36 = {}.ElResult;
    /** @type {[typeof __VLS_components.ElResult, typeof __VLS_components.elResult, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        icon: "success",
        title: (__VLS_ctx.t('plugin.example_result_title', 'Plugin API returned successfully')),
        subTitle: (`${__VLS_ctx.pingResult.message} (${__VLS_ctx.pingResult.plugin})`),
    }));
    const __VLS_38 = __VLS_37({
        icon: "success",
        title: (__VLS_ctx.t('plugin.example_result_title', 'Plugin API returned successfully')),
        subTitle: (`${__VLS_ctx.pingResult.message} (${__VLS_ctx.pingResult.plugin})`),
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
}
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['plugin-example-page']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card']} */ ;
/** @type {__VLS_StyleScopedClasses['page-card__header']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-example-page__meta']} */ ;
/** @type {__VLS_StyleScopedClasses['plugin-example-page__actions']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            loading: loading,
            pingResult: pingResult,
            pageTitle: pageTitle,
            componentName: componentName,
            routePath: routePath,
            routePermission: routePermission,
            handlePing: handlePing,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
