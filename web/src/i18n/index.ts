import { computed } from 'vue';

import { useLocaleStore } from '@/store/locale';
import type { RouteLocationNormalizedLoaded, RouteMeta } from 'vue-router';

import { appMessages, type AppLocale, type AppMessageKey } from './messages';

function isMessageKey(value: string): value is AppMessageKey {
  return Object.prototype.hasOwnProperty.call(appMessages['zh-CN'], value);
}

function formatMessage(template: string, values?: Record<string, string | number>): string {
  if (!values) {
    return template;
  }
  return template.replace(/\{([a-zA-Z0-9_]+)\}/g, (_, token: string) => {
    const value = values[token];
    return value === undefined || value === null ? `{${token}}` : String(value);
  });
}

export function getAppLanguage(): AppLocale {
  const localeStore = useLocaleStore();
  return localeStore.language;
}

export function translate(key: string, fallback = '', values?: Record<string, string | number>): string {
  const localeStore = useLocaleStore();
  const locale = localeStore.language as AppLocale;
  const normalizedKey = key.trim();
  if (normalizedKey !== '' && isMessageKey(normalizedKey)) {
    const bucket = (appMessages[locale] ?? appMessages['zh-CN']) as Record<string, string>;
    const fallbackBucket = appMessages['zh-CN'] as Record<string, string>;
    const translated = bucket[normalizedKey] ?? fallbackBucket[normalizedKey];
    if (translated) {
      return formatMessage(translated, values);
    }
  }
  if (fallback.trim() !== '') {
    return formatMessage(fallback, values);
  }
  return normalizedKey;
}

export function useAppI18n() {
  const localeStore = useLocaleStore();
  const language = computed(() => localeStore.language);

  function t(key: string, fallback = '', values?: Record<string, string | number>): string {
    const currentLanguage = language.value as AppLocale;
    const normalizedKey = key.trim();
    if (normalizedKey !== '' && isMessageKey(normalizedKey)) {
      const bucket = (appMessages[currentLanguage] ?? appMessages['zh-CN']) as Record<string, string>;
      const fallbackBucket = appMessages['zh-CN'] as Record<string, string>;
      const translated = bucket[normalizedKey] ?? fallbackBucket[normalizedKey];
      if (translated) {
        return formatMessage(translated, values);
      }
    }
    if (fallback.trim() !== '') {
      return formatMessage(fallback, values);
    }
    return normalizedKey;
  }

  return {
    language,
    t,
  };
}

export function getRouteTitle(meta: Pick<RouteMeta, 'title' | 'componentName' | 'permission' | 'link'> & {
  titleKey?: string;
  titleDefault?: string;
}, fallback = ''): string {
  const localeStore = useLocaleStore();
  const locale = localeStore.language as AppLocale;
  const key = (meta.titleKey || meta.componentName || meta.title || '').trim();
  const defaultTitle = (meta.titleDefault || meta.title || fallback).trim();

  if (key !== '' && isMessageKey(key)) {
    const bucket = (appMessages[locale] ?? appMessages['zh-CN']) as Record<string, string>;
    const fallbackBucket = appMessages['zh-CN'] as Record<string, string>;
    const translated = bucket[key] ?? fallbackBucket[key];
    if (translated) {
      return translated;
    }
  }

  return defaultTitle || fallback || meta.title || '';
}

export function resolveRouteLocaleMeta(route: Pick<RouteLocationNormalizedLoaded, 'meta'>): { title: string; subtitle: string } {
  const meta = route.meta as RouteMeta & {
    titleKey?: string;
    titleDefault?: string;
    subtitle?: string;
    subtitleKey?: string;
    subtitleDefault?: string;
  };
  const title = getRouteTitle(meta, '');
  const subtitleKey = typeof meta.subtitleKey === 'string' ? meta.subtitleKey : '';
  const subtitleDefault = typeof meta.subtitleDefault === 'string' && meta.subtitleDefault.trim() !== ''
    ? meta.subtitleDefault
    : typeof meta.subtitle === 'string'
      ? meta.subtitle
      : '';
  const subtitle = subtitleKey.trim() !== '' && isMessageKey(subtitleKey.trim())
    ? translate(subtitleKey, subtitleDefault)
    : subtitleDefault;

  return {
    title,
    subtitle,
  };
}
