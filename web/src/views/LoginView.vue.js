import { computed, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowDown } from '@element-plus/icons-vue';
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
const currentLanguageLabel = computed(() => {
    return localeStore.language === 'en-US' ? t('common.language_en', 'English') : t('common.language_zh', 'Chinese');
});
const rules = computed(() => ({
    username: [{ required: true, message: t('login.form.username_required', 'Enter username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.form.password_required', 'Enter password'), trigger: 'blur' }],
}));
function switchLanguage(language) {
    localeStore.setLanguage(language);
    sessionStore.setLanguage(language);
}
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
            ElMessage.success(t('login.success', 'Login successful'));
            await router.replace(redirectTarget.value);
        }
        catch (error) {
            const message = error instanceof Error ? error.message : t('login.failure', 'Login failed');
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
/** @type {__VLS_StyleScopedClasses['login-card__panel-header-top']} */ ;
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
(__VLS_ctx.t('login.title', 'Login'));
var __VLS_7;
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('login.welcome', 'Welcome to GoAdmin'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('login.description', 'Sign in with the account created by the backend.'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.ul, __VLS_intrinsicElements.ul)({
    ...{ class: "login-card__highlights" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('login.highlight.jwt_session', 'JWT login and session management'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('login.highlight.dynamic_menu', 'Dynamic menus and permission-driven access'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.li, __VLS_intrinsicElements.li)({});
(__VLS_ctx.t('login.highlight.element_plus', 'Unified Element Plus styling'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__stats" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
(__VLS_ctx.apiBaseUrl);
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.api_base_url', 'API base URL'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.username', 'Username'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.default_account', 'Default account: admin / admin123'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.section, __VLS_intrinsicElements.section)({
    ...{ class: "login-card__panel" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__panel-header" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-card__panel-header-top" },
});
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
__VLS_asFunctionalElement(__VLS_intrinsicElements.h2, __VLS_intrinsicElements.h2)({});
(__VLS_ctx.t('login.title', 'Login'));
__VLS_asFunctionalElement(__VLS_intrinsicElements.p, __VLS_intrinsicElements.p)({});
(__VLS_ctx.t('login.description', 'Sign in with the account created by the backend.'));
const __VLS_8 = {}.ElDropdown;
/** @type {[typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, typeof __VLS_components.ElDropdown, typeof __VLS_components.elDropdown, ]} */ ;
// @ts-ignore
const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
    ...{ 'onCommand': {} },
    trigger: "click",
}));
const __VLS_10 = __VLS_9({
    ...{ 'onCommand': {} },
    trigger: "click",
}, ...__VLS_functionalComponentArgsRest(__VLS_9));
let __VLS_12;
let __VLS_13;
let __VLS_14;
const __VLS_15 = {
    onCommand: (__VLS_ctx.switchLanguage)
};
__VLS_11.slots.default;
const __VLS_16 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({
    ...{ class: "login-card__language" },
    text: true,
}));
const __VLS_18 = __VLS_17({
    ...{ class: "login-card__language" },
    text: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_17));
__VLS_19.slots.default;
(__VLS_ctx.t('common.language', 'Language'));
(__VLS_ctx.currentLanguageLabel);
const __VLS_20 = {}.ElIcon;
/** @type {[typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, typeof __VLS_components.ElIcon, typeof __VLS_components.elIcon, ]} */ ;
// @ts-ignore
const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
    ...{ class: "login-card__language-arrow" },
}));
const __VLS_22 = __VLS_21({
    ...{ class: "login-card__language-arrow" },
}, ...__VLS_functionalComponentArgsRest(__VLS_21));
__VLS_23.slots.default;
const __VLS_24 = {}.ArrowDown;
/** @type {[typeof __VLS_components.ArrowDown, ]} */ ;
// @ts-ignore
const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({}));
const __VLS_26 = __VLS_25({}, ...__VLS_functionalComponentArgsRest(__VLS_25));
var __VLS_23;
var __VLS_19;
{
    const { dropdown: __VLS_thisSlot } = __VLS_11.slots;
    const __VLS_28 = {}.ElDropdownMenu;
    /** @type {[typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, typeof __VLS_components.ElDropdownMenu, typeof __VLS_components.elDropdownMenu, ]} */ ;
    // @ts-ignore
    const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({}));
    const __VLS_30 = __VLS_29({}, ...__VLS_functionalComponentArgsRest(__VLS_29));
    __VLS_31.slots.default;
    const __VLS_32 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
        command: "zh-CN",
    }));
    const __VLS_34 = __VLS_33({
        command: "zh-CN",
    }, ...__VLS_functionalComponentArgsRest(__VLS_33));
    __VLS_35.slots.default;
    (__VLS_ctx.t('common.language_zh', 'Chinese'));
    var __VLS_35;
    const __VLS_36 = {}.ElDropdownItem;
    /** @type {[typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, typeof __VLS_components.ElDropdownItem, typeof __VLS_components.elDropdownItem, ]} */ ;
    // @ts-ignore
    const __VLS_37 = __VLS_asFunctionalComponent(__VLS_36, new __VLS_36({
        command: "en-US",
    }));
    const __VLS_38 = __VLS_37({
        command: "en-US",
    }, ...__VLS_functionalComponentArgsRest(__VLS_37));
    __VLS_39.slots.default;
    (__VLS_ctx.t('common.language_en', 'English'));
    var __VLS_39;
    var __VLS_31;
}
var __VLS_11;
const __VLS_40 = {}.ElForm;
/** @type {[typeof __VLS_components.ElForm, typeof __VLS_components.elForm, typeof __VLS_components.ElForm, typeof __VLS_components.elForm, ]} */ ;
// @ts-ignore
const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    ...{ class: "login-form" },
    labelPosition: "top",
}));
const __VLS_42 = __VLS_41({
    ...{ 'onKeyup': {} },
    ref: "formRef",
    model: (__VLS_ctx.form),
    rules: (__VLS_ctx.rules),
    ...{ class: "login-form" },
    labelPosition: "top",
}, ...__VLS_functionalComponentArgsRest(__VLS_41));
let __VLS_44;
let __VLS_45;
let __VLS_46;
const __VLS_47 = {
    onKeyup: (__VLS_ctx.onSubmit)
};
/** @type {typeof __VLS_ctx.formRef} */ ;
var __VLS_48 = {};
__VLS_43.slots.default;
const __VLS_50 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_51 = __VLS_asFunctionalComponent(__VLS_50, new __VLS_50({
    label: (__VLS_ctx.t('login.username', 'Username')),
    prop: "username",
}));
const __VLS_52 = __VLS_51({
    label: (__VLS_ctx.t('login.username', 'Username')),
    prop: "username",
}, ...__VLS_functionalComponentArgsRest(__VLS_51));
__VLS_53.slots.default;
const __VLS_54 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_55 = __VLS_asFunctionalComponent(__VLS_54, new __VLS_54({
    modelValue: (__VLS_ctx.form.username),
    autocomplete: "username",
    placeholder: (__VLS_ctx.t('login.username_placeholder', 'Enter username')),
}));
const __VLS_56 = __VLS_55({
    modelValue: (__VLS_ctx.form.username),
    autocomplete: "username",
    placeholder: (__VLS_ctx.t('login.username_placeholder', 'Enter username')),
}, ...__VLS_functionalComponentArgsRest(__VLS_55));
var __VLS_53;
const __VLS_58 = {}.ElFormItem;
/** @type {[typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, typeof __VLS_components.ElFormItem, typeof __VLS_components.elFormItem, ]} */ ;
// @ts-ignore
const __VLS_59 = __VLS_asFunctionalComponent(__VLS_58, new __VLS_58({
    label: (__VLS_ctx.t('login.password', 'Password')),
    prop: "password",
}));
const __VLS_60 = __VLS_59({
    label: (__VLS_ctx.t('login.password', 'Password')),
    prop: "password",
}, ...__VLS_functionalComponentArgsRest(__VLS_59));
__VLS_61.slots.default;
const __VLS_62 = {}.ElInput;
/** @type {[typeof __VLS_components.ElInput, typeof __VLS_components.elInput, ]} */ ;
// @ts-ignore
const __VLS_63 = __VLS_asFunctionalComponent(__VLS_62, new __VLS_62({
    modelValue: (__VLS_ctx.form.password),
    autocomplete: "current-password",
    placeholder: (__VLS_ctx.t('login.password_placeholder', 'Enter password')),
    type: "password",
    showPassword: true,
}));
const __VLS_64 = __VLS_63({
    modelValue: (__VLS_ctx.form.password),
    autocomplete: "current-password",
    placeholder: (__VLS_ctx.t('login.password_placeholder', 'Enter password')),
    type: "password",
    showPassword: true,
}, ...__VLS_functionalComponentArgsRest(__VLS_63));
var __VLS_61;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
    ...{ class: "login-form__meta" },
});
const __VLS_66 = {}.ElTag;
/** @type {[typeof __VLS_components.ElTag, typeof __VLS_components.elTag, typeof __VLS_components.ElTag, typeof __VLS_components.elTag, ]} */ ;
// @ts-ignore
const __VLS_67 = __VLS_asFunctionalComponent(__VLS_66, new __VLS_66({
    effect: "plain",
    round: true,
    type: "info",
}));
const __VLS_68 = __VLS_67({
    effect: "plain",
    round: true,
    type: "info",
}, ...__VLS_functionalComponentArgsRest(__VLS_67));
__VLS_69.slots.default;
(__VLS_ctx.t('login.api_base_url', 'API base URL'));
(__VLS_ctx.apiBaseUrl);
var __VLS_69;
__VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
(__VLS_ctx.t('login.default_account', 'Default account: admin / admin123'));
const __VLS_70 = {}.ElButton;
/** @type {[typeof __VLS_components.ElButton, typeof __VLS_components.elButton, typeof __VLS_components.ElButton, typeof __VLS_components.elButton, ]} */ ;
// @ts-ignore
const __VLS_71 = __VLS_asFunctionalComponent(__VLS_70, new __VLS_70({
    ...{ 'onClick': {} },
    ...{ class: "login-form__submit" },
    type: "primary",
    loading: (__VLS_ctx.loading),
}));
const __VLS_72 = __VLS_71({
    ...{ 'onClick': {} },
    ...{ class: "login-form__submit" },
    type: "primary",
    loading: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_71));
let __VLS_74;
let __VLS_75;
let __VLS_76;
const __VLS_77 = {
    onClick: (__VLS_ctx.onSubmit)
};
__VLS_73.slots.default;
(__VLS_ctx.t('login.submit', 'Login'));
var __VLS_73;
var __VLS_43;
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
/** @type {__VLS_StyleScopedClasses['login-card__panel-header-top']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__language']} */ ;
/** @type {__VLS_StyleScopedClasses['login-card__language-arrow']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form__meta']} */ ;
/** @type {__VLS_StyleScopedClasses['login-form__submit']} */ ;
// @ts-ignore
var __VLS_49 = __VLS_48;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            ArrowDown: ArrowDown,
            t: t,
            formRef: formRef,
            loading: loading,
            appTitle: appTitle,
            apiBaseUrl: apiBaseUrl,
            form: form,
            currentLanguageLabel: currentLanguageLabel,
            rules: rules,
            switchLanguage: switchLanguage,
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
