import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import en from 'element-plus/es/locale/lang/en';
import zhCn from 'element-plus/es/locale/lang/zh-cn';

export type AppLanguage = 'zh-CN' | 'en-US';

const LANGUAGE_KEY = 'goadmin.language';

function canUseStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}

function normalizeLanguage(value: string | null | undefined): AppLanguage {
  const raw = typeof value === 'string' ? value.trim().toLowerCase() : '';
  if (raw.startsWith('en')) {
    return 'en-US';
  }
  return 'zh-CN';
}

function readStoredLanguage(): AppLanguage {
  if (!canUseStorage()) {
    return 'zh-CN';
  }
  return normalizeLanguage(window.localStorage.getItem(LANGUAGE_KEY));
}

function persistLanguage(language: AppLanguage): void {
  if (!canUseStorage()) {
    return;
  }
  window.localStorage.setItem(LANGUAGE_KEY, language);
}

export const useLocaleStore = defineStore('locale', () => {
  const language = ref<AppLanguage>(readStoredLanguage());

  const elementLocale = computed(() => (language.value === 'en-US' ? en : zhCn));

  function hydrate(): void {
    language.value = readStoredLanguage();
  }

  function setLanguage(value: string | null | undefined): void {
    language.value = normalizeLanguage(value);
    persistLanguage(language.value);
  }

  function syncFromUser(value: { language?: string | null } | null | undefined): void {
    if (value && typeof value.language === 'string' && value.language.trim() !== '') {
      setLanguage(value.language);
    }
  }

  function clear(): void {
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
