import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import en from 'element-plus/es/locale/lang/en';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import { resolveInitialI18nLanguage, resolvePreferredI18nLanguage } from '@/i18n/language';
const LANGUAGE_KEY = 'goadmin.language';
const LANGUAGE_PREFERENCE_KEY = 'goadmin.language_preference';
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}
function readStoredLanguage() {
    if (!canUseStorage()) {
        return resolveInitialI18nLanguage();
    }
    return resolveInitialI18nLanguage(window.localStorage.getItem(LANGUAGE_KEY));
}
function hasStoredLanguagePreference() {
    if (!canUseStorage()) {
        return false;
    }
    return window.localStorage.getItem(LANGUAGE_PREFERENCE_KEY) === 'explicit';
}
function persistLanguagePreference(explicit) {
    if (!canUseStorage()) {
        return;
    }
    if (explicit) {
        window.localStorage.setItem(LANGUAGE_PREFERENCE_KEY, 'explicit');
        return;
    }
    window.localStorage.removeItem(LANGUAGE_PREFERENCE_KEY);
}
function persistLanguage(language) {
    if (!canUseStorage()) {
        return;
    }
    window.localStorage.setItem(LANGUAGE_KEY, language);
}
export const useLocaleStore = defineStore('locale', () => {
    const language = ref(readStoredLanguage());
    const hasLanguagePreference = ref(hasStoredLanguagePreference());
    const elementLocale = computed(() => (language.value === 'en-US' ? en : zhCn));
    function applyLanguagePreference(explicitLanguage, profileLanguage, markAsUserPreference = true) {
        language.value = resolvePreferredI18nLanguage(explicitLanguage, profileLanguage);
        if (markAsUserPreference) {
            hasLanguagePreference.value = true;
            persistLanguagePreference(true);
        }
        else {
            hasLanguagePreference.value = false;
            persistLanguagePreference(false);
        }
        persistLanguage(language.value);
    }
    function hydrate() {
        language.value = readStoredLanguage();
        hasLanguagePreference.value = hasStoredLanguagePreference();
    }
    function setLanguage(value, profileLanguage) {
        applyLanguagePreference(value, profileLanguage);
    }
    function syncFromUser(value) {
        applyLanguagePreference(undefined, value?.language, false);
    }
    function clear() {
        language.value = 'zh-CN';
        hasLanguagePreference.value = false;
        persistLanguagePreference(false);
        persistLanguage(language.value);
    }
    return {
        language,
        hasLanguagePreference,
        elementLocale,
        applyLanguagePreference,
        hydrate,
        setLanguage,
        syncFromUser,
        clear,
    };
});
