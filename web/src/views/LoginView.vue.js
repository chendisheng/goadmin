import { computed, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { login } from '@/api/auth';
import { useAppI18n } from '@/i18n';
import { useLocaleStore } from '@/store/locale';
import { useMenuStore } from '@/store/menu';
import { useSessionStore } from '@/store/session';
const router = useRouter();
const route = useRoute();
const sessionStore = useSessionStore();
const localeStore = useLocaleStore();
const menuStore = useMenuStore();
const { t } = useAppI18n();
const formRef = ref();
const loading = ref(false);
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const form = reactive({
    username: 'admin',
    password: 'admin123',
});
const redirectTarget = computed(() => {
    const redirect = route.query.redirect;
    if (typeof redirect === 'string' && redirect.trim() !== '' && redirect !== '/login') {
        return redirect;
    }
    return '/dashboard';
});
const rules = computed(() => ({
    username: [{ required: true, message: t('login.form.username_required', '请输入用户名'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.form.password_required', '请输入密码'), trigger: 'blur' }],
}));
async function onSubmit() {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate(async (valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        try {
            const response = await login({ username: form.username.trim(), password: form.password });
            sessionStore.applyLoginResponse(response);
            localeStore.syncFromUser(response.user);
            await menuStore.ensureLoaded(router);
            ElMessage.success(t('login.success', '登录成功'));
            await router.replace(redirectTarget.value);
        }
        catch (error) {
            const message = error instanceof Error ? error.message : t('login.failure', '登录失败');
            ElMessage.error(message);
        }
        finally {
            loading.value = false;
        }
    });
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-top']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-top']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-body']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-body']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__highlights']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__highlights']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__stats']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__stats']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__stats']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel-header']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel-header']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form']} */ ;
/** @type {__VLS_StyleScopedClasses['login-page']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['el-card__body']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-body']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__stats']} */ ;
/** @type {__VLS_StyleScopedClasses['login-page']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-top']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel-header']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-body']} */ ;
// CSS variable injection 
// CSS variable injection end 
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-page" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div)({
    ...{ class: "login-page__backdrop" },
});
const __VLS_0 = {}.ElCard;
/** @type {[typeof __VLS_components.ElCard, typeof __VLS_components.elCard, typeof __VLS_components.ElCard, typeof __VLS_components.elCard, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    ...{ class: "login-card" },
    shadow: "never",
}));
const __VLS_2 = __VLS_1({
    ...{ class: "login-card" },
    shadow: "never",
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
__VLS_3.slots.default;
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "login-card__brand" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__brand-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__logo" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h1, __VLS_intrinsicElements.h1)({});
(__VLS_ctx.t('app.title', __VLS_ctx.appTitle));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('app.subtitle', 'Frontend Core'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__brand-body" },
});
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
(__VLS_ctx.t('login.title', '登录'));
var __VLS_7;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('login.welcome', '欢迎使用 GoAdmin'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('login.description', '请输入后端创建的账号登录系统。'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "login-card__highlights" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__stats" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.apiBaseUrl);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.api_base_url', 'API 基址'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.username', '用户名'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.default_account', '默认账号：admin / admin123'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "login-card__panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__panel-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('login.title', '登录'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('login.description', '请输入后端创建的账号登录系统。'));
const __VLS_8 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    ...{ class: "login-form" },
    labelPosition: "top",
}));
const __VLS_10 = __VLS_9({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    ...{ class: "login-form" },
    labelPosition: "top",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onKeyup: (__VLS_ctx.onSubmit)
};
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_16 = {};
__VLS_11.slots.default;
const __VLS_18 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_19 = __VLS_asFunctionalComponent(__VLS_18, new __VLS_18({
    label: (__VLS_ctx.t('login.username', '用户名')),
    prop: "username",
}));
const __VLS_20 = __VLS_19({
    label: (__VLS_ctx.t('login.username', '用户名')),
    prop: "username",
}, ...__VLS_functionalComponentArgsRest(__VLS_19));
__VLS_21.slots.default;
const __VLS_22 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_23 = __VLS_asFunctionalComponent(__VLS_22, new __VLS_22({
    modelValue: (__VLS_ctx.form.username),
    autocomplete: "username",
    placeholder: (__VLS_ctx.t('login.username_placeholder', '请输入用户名')),
}));
const __VLS_24 = __VLS_23({
    modelValue: (__VLS_ctx.form.username),
    autocomplete: "username",
    placeholder: (__VLS_ctx.t('login.username_placeholder', '请输入用户名')),
}, ...__VLS_functionalComponentArgsRest(__VLS_23));
var __VLS_21;
const __VLS_26 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_27 = __VLS_asFunctionalComponent(__VLS_26, new __VLS_26({
    label: (__VLS_ctx.t('login.password', '密码')),
    prop: "password",
}));
const __VLS_28 = __VLS_27({
    label: (__VLS_ctx.t('login.password', '密码')),
    prop: "password",
}, ...__VLS_functionalComponentArgsRest(__VLS_27));
__VLS_29.slots.default;
const __VLS_30 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_31 = __VLS_asFunctionalComponent(__VLS_30, new __VLS_30({
    modelValue: (__VLS_ctx.form.password),
    autocomplete: "current-password",
    placeholder: (__VLS_ctx.t('login.password_placeholder', '请输入密码')),
    type: "password",
    showPassword: true,
}));
const __VLS_32 = __VLS_31({
    modelValue: (__VLS_ctx.form.password),
    autocomplete: "current-password",
    placeholder: (__VLS_ctx.t('login.password_placeholder', '请输入密码')),
    type: "password",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_31));
var __VLS_29;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-form__meta" },
});
const __VLS_34 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_35 = __VLS_asFunctionalComponent(__VLS_34, new __VLS_34({
    effect: "plain",
    round: true,
    type: "info",
}));
const __VLS_36 = __VLS_35({
    effect: "plain",
    round: true,
    type: "info",
}, ...__VLS_functionalComponentArgsRest(__VLS_35));
__VLS_37.slots.default;
(__VLS_ctx.t('login.api_base_url', 'API 基址'));
(__VLS_ctx.apiBaseUrl);
var __VLS_37;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.default_account', '默认账号：admin / admin123'));
const __VLS_38 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_39 = __VLS_asFunctionalComponent(__VLS_38, new __VLS_38({
    ...{ 'onClick': {} },
    ...{ class: "login-form__submit" },
    type: "primary",
    loading: (__VLS_ctx.loading),
}));
const __VLS_40 = __VLS_39({
    ...{ 'onClick': {} },
    ...{ class: "login-form__submit" },
    type: "primary",
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_39));
let __VLS_42;
let __VLS_43;
let __VLS_44;
const __VLS_45 = {
    onClick: (__VLS_ctx.onSubmit)
};
__VLS_41.slots.default;
(__VLS_ctx.t('login.submit', '登录'));
var __VLS_41;
var __VLS_11;
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['login-page']} */ ;
/** @type {__VLS_StyleScopedClasses['login-page__backdrop']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-top']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__logo']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__brand-body']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__highlights']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__stats']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__panel-header']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form__meta']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form__submit']} */ ;
// @ts-ignore
var __VLS_17 = __VLS_16;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            t: t,
            formRef: formRef,
            loading: loading,
            appTitle: appTitle,
            apiBaseUrl: apiBaseUrl,
            form: form,
            rules: rules,
            onSubmit: onSubmit,
        };
    },
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
});
; /* PartiallyEnd: #4569/main.vue */
