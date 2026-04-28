import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import en from 'element-plus/es/locale/lang/en';
import zhCn from 'element-plus/es/locale/lang/zh-cn';

import { resolveInitialI18nLanguage, resolvePreferredI18nLanguage, tryNormalizeI18nLanguage } from '@/i18n/language';

export type AppLanguage = 'zh-CN' | 'en-US';

const LANGUAGE_KEY = 'goadmin.language';
const LANGUAGE_PREFERENCE_KEY = 'goadmin.language_preference';

function canUseStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}

function readStoredLanguage(): AppLanguage {
  if (!canUseStorage()) {
    return resolveInitialI18nLanguage();
  }
  return resolveInitialI18nLanguage(window.localStorage.getItem(LANGUAGE_KEY));
}

function hasStoredLanguagePreference(): boolean {
  if (!canUseStorage()) {
    return false;
  }
  return window.localStorage.getItem(LANGUAGE_PREFERENCE_KEY) === 'explicit';
}

function persistLanguagePreference(explicit: boolean): void {
  if (!canUseStorage()) {
    return;
  }
  if (explicit) {
    window.localStorage.setItem(LANGUAGE_PREFERENCE_KEY, 'explicit');
    return;
  }
  window.localStorage.removeItem(LANGUAGE_PREFERENCE_KEY);
}

function persistLanguage(language: AppLanguage): void {
  if (!canUseStorage()) {
    return;
  }
  window.localStorage.setItem(LANGUAGE_KEY, language);
}

export const useLocaleStore = defineStore('locale', () => {
  const language = ref<AppLanguage>(readStoredLanguage());
  const hasLanguagePreference = ref(hasStoredLanguagePreference());

  const elementLocale = computed(() => (language.value === 'en-US' ? en : zhCn));

  function applyLanguagePreference(
    explicitLanguage?: string | null,
    profileLanguage?: string | null,
    markAsUserPreference = true,
  ): void {
    language.value = resolvePreferredI18nLanguage(explicitLanguage, profileLanguage);
    if (markAsUserPreference) {
      hasLanguagePreference.value = true;
      persistLanguagePreference(true);
    } else {
      hasLanguagePreference.value = false;
      persistLanguagePreference(false);
    }
    persistLanguage(language.value);
  }

  function hydrate(): void {
    language.value = readStoredLanguage();
    hasLanguagePreference.value = hasStoredLanguagePreference();
  }

  function setLanguage(value: string | null | undefined, profileLanguage?: string | null): void {
    applyLanguagePreference(value, profileLanguage);
  }

  function syncFromUser(value: { language?: string | null } | null | undefined): void {
    applyLanguagePreference(undefined, value?.language, false);
  }

  function clear(): void {
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
