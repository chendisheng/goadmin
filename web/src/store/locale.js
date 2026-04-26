import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import en from 'element-plus/es/locale/lang/en';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
const LANGUAGE_KEY = 'goadmin.language';
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}
function normalizeLanguage(value) {
    const raw = typeof value === 'string' ? value.trim().toLowerCase() : '';
    if (raw.startsWith('en')) {
        return 'en-US';
    }
    return 'zh-CN';
}
function readStoredLanguage() {
    if (!canUseStorage()) {
        return 'zh-CN';
    }
    return normalizeLanguage(window.localStorage.getItem(LANGUAGE_KEY));
}
function persistLanguage(language) {
    if (!canUseStorage()) {
        return;
    }
    window.localStorage.setItem(LANGUAGE_KEY, language);
}
export const useLocaleStore = defineStore('locale', () => {
    const language = ref(readStoredLanguage());
    const elementLocale = computed(() => (language.value === 'en-US' ? en : zhCn));
    function hydrate() {
        language.value = readStoredLanguage();
    }
    function setLanguage(value) {
        language.value = normalizeLanguage(value);
        persistLanguage(language.value);
    }
    function syncFromUser(value) {
        if (value && typeof value.language === 'string' && value.language.trim() !== '') {
            setLanguage(value.language);
        }
    }
    function clear() {
        language.value = 'zh-CN';
        persistLanguage(language.value);
    }
    return {
        language,
        elementLocale,
        hydrate,
        setLanguage,
        syncFromUser,
        clear,
    };
});
