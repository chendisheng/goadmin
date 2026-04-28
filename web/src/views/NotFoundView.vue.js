import { useAppI18n } from '@/i18n';
import { useRouter } from 'vue-router';
const router = useRouter();
const { t } = useAppI18n();
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "not-found-page" },
});
const __VLS_0 = {}.ElResult;
/** @type {[typeof __VLS_components.ElResult, typeof __VLS_components.elResult, typeof __VLS_components.ElResult, typeof __VLS_components.elResult, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    icon: "warning",
    title: (__VLS_ctx.t('common.not_found', 'Page not found')),
    subTitle: (__VLS_ctx.t('common.not_found_hint', 'Please return to the dashboard to continue')),
}));
const __VLS_2 = __VLS_1({
    icon: "warning",
    title: (__VLS_ctx.t('common.not_found', 'Page not found')),
    subTitle: (__VLS_ctx.t('common.not_found_hint', 'Please return to the dashboard to continue')),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
{
    const { extra: __VLS_thisSlot } = __VLS_3.slots;
    const __VLS_4 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
        ...{ 'onClick': {} },
        type: "primary",
    }));
    const __VLS_6 = __VLS_5({
        ...{ 'onClick': {} },
        type: "primary",
    }, ...__VLS_functionalComponentArgsRest(__VLS_5));
    let __VLS_8;
    let __VLS_9;
    let __VLS_10;
    const __VLS_11 = {
        onClick: (...[$event]) => {
            __VLS_ctx.router.push('/dashboard');
        }
    };
    __VLS_7.slots.default;
    (__VLS_ctx.t('common.back_to_dashboard', 'Back to dashboard'));
    var __VLS_7;
}
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['not-found-page']} */ ;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            router: router,
            t: t,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
