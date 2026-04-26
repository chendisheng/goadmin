const __VLS_props = withDefaults(defineProps(), {
    width: '720px',
    loading: false,
    confirmText: '保存',
    cancelText: '取消',
});
const emit = defineEmits();
function updateVisible(value) {
    emit('update:modelValue', value);
    if (!value) {
        emit('cancel');
    }
}
function handleCancel() {
    emit('update:modelValue', false);
    emit('cancel');
}
function handleConfirm() {
    emit('confirm');
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_withDefaultsArg = (function (t) { return t; })({
    width: '720px',
    loading: false,
    confirmText: '保存',
    cancelText: '取消',
});
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
const __VLS_0 = {}.ElDialog;
/** @type {[typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, typeof __VLS_components.ElDialog, typeof __VLS_components.elDialog, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ 'onUpdate:modelValue': {} },
    modelValue: (__VLS_ctx.modelValue),
    title: (__VLS_ctx.title),
    width: (__VLS_ctx.width),
    destroyOnClose: true,
}));
const __VLS_2 = __VLS_1({
    ...{ 'onUpdate:modelValue': {} },
    modelValue: (__VLS_ctx.modelValue),
    title: (__VLS_ctx.title),
    width: (__VLS_ctx.width),
    destroyOnClose: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
let __VLS_4;
let __VLS_5;
let __VLS_6;
const __VLS_7 = {
    'onUpdate:modelValue': (__VLS_ctx.updateVisible)
};
var __VLS_8 = {};
__VLS_3.slots.default;
var __VLS_9 = {};
{
    const { footer: __VLS_thisSlot } = __VLS_3.slots;
    var __VLS_11 = {};
    const __VLS_13 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
        ...{ 'onClick': {} },
    }));
    const __VLS_15 = __VLS_14({
        ...{ 'onClick': {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_14));
    let __VLS_17;
    let __VLS_18;
    let __VLS_19;
    const __VLS_20 = {
        onClick: (__VLS_ctx.handleCancel)
    };
    __VLS_16.slots.default;
    (__VLS_ctx.cancelText);
    var __VLS_16;
    const __VLS_21 = {}.ElButton;
    /** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.loading),
    }));
    const __VLS_23 = __VLS_22({
        ...{ 'onClick': {} },
        type: "primary",
        loading: (__VLS_ctx.loading),
    }, ...__VLS_functionalComponentArgsRest(__VLS_22));
    let __VLS_25;
    let __VLS_26;
    let __VLS_27;
    const __VLS_28 = {
        onClick: (__VLS_ctx.handleConfirm)
    };
    __VLS_24.slots.default;
    (__VLS_ctx.confirmText);
    var __VLS_24;
}
var __VLS_3;
// @ts-ignore
var __VLS_10 = __VLS_9, __VLS_12 = __VLS_11;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            updateVisible: updateVisible,
            handleCancel: handleCancel,
            handleConfirm: handleConfirm,
        };
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
const __VLS_component = (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
    props: {},
});
export default {};
; /* PartiallyEnd: #4569/main.vue */
