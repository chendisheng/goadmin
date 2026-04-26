import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { pingExamplePlugin } from '@/api/plugins';
const route = useRoute();
const loading = ref(false);
const pingResult = ref(null);
const pageTitle = computed(() => (typeof route.meta.title === 'string' && route.meta.title.trim() !== '' ? route.meta.title : '插件示例'));
const componentName = computed(() => String(route.meta.componentName || 'view/plugin/example/index'));
const routePath = computed(() => route.path);
const routePermission = computed(() => String(route.meta.permission || 'plugin:example:view'));
async function handlePing() {
    loading.value = true;
    try {
        pingResult.value = await pingExamplePlugin();
        ElMessage.success('插件接口调用成功');
    }
    catch (error) {
        ElMessage.error(error instanceof Error ? error.message : '插件接口调用失败');
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
    var __VLS_7;
}
const __VLS_8 = {}.ElAlert;
/** @type {[typeof __VLS_components.ElAlert, typeof __VLS_components.elAlert, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    title: "这是一个由插件注册的动态页面",
    description: "页面组件路径来自后端菜单配置 `view/plugin/example/index`，并通过前端动态组件映射加载。",
    type: "success",
    showIcon: true,
    closable: (false),
}));
const __VLS_10 = __VLS_9({
    title: "这是一个由插件注册的动态页面",
    description: "页面组件路径来自后端菜单配置 `view/plugin/example/index`，并通过前端动态组件映射加载。",
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
    label: "路由路径",
}));
const __VLS_18 = __VLS_17({
    label: "路由路径",
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
(__VLS_ctx.routePath);
var __VLS_19;
const __VLS_20 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    label: "组件标识",
}));
const __VLS_22 = __VLS_21({
    label: "组件标识",
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
(__VLS_ctx.componentName);
var __VLS_23;
const __VLS_24 = {}.ElDescriptionsItem;
/** @type {[typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, typeof __VLS_components.ElDescriptionsItem, typeof __VLS_components.elDescriptionsItem, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
    label: "权限标识",
}));
const __VLS_26 = __VLS_25({
    label: "权限标识",
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
var __VLS_31;
if (__VLS_ctx.pingResult) {
    const __VLS_36 = {}.ElResult;
    /** @type {[typeof __VLS_components.ElResult, typeof __VLS_components.elResult, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        icon: "success",
        title: "插件接口返回成功",
        subTitle: (`${__VLS_ctx.pingResult.message} (${__VLS_ctx.pingResult.plugin})`),
    }));
    const __VLS_38 = __VLS_37({
        icon: "success",
        title: "插件接口返回成功",
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
