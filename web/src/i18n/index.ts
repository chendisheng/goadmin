import { computed, ref } from 'vue';
import i18next from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import { useLocaleStore } from '@/store/locale';
import type { RouteLocationNormalizedLoaded, RouteMeta } from 'vue-router';

import {
  I18N_DEFAULT_LANGUAGE,
  I18N_DEFAULT_NS,
  I18N_SUPPORTED_LANGUAGES,
  normalizeI18nLanguage,
  type I18nLanguage,
} from './language';
import { ensureNamespaces, getLoadedNamespaces } from './loader';
import {
  collectNamespacesFromRouteMeta,
  isValidNamespace,
  I18N_BASE_NAMESPACES,
  namespaceFromKey,
  namespacesFromRouteName,
  namespacesFromRoutePath,
  namespacesFromComponentName,
} from './namespaces';

const APP_NAMESPACE = I18N_DEFAULT_NS;

let initPromise: Promise<typeof i18next> | null = null;
const runtimeNamespaces = new Set<string>(I18N_BASE_NAMESPACES);
const runtimeVersion = ref(0);

function bumpRuntimeVersion(): void {
  runtimeVersion.value += 1;
}

function normalizeNamespaces(namespaces: string[]): string[] {
  const result = new Set<string>(I18N_BASE_NAMESPACES);
  for (const ns of namespaces) {
    const normalized = ns.trim().toLowerCase();
    if (normalized === '' || !isValidNamespace(normalized)) {
      continue;
    }
    result.add(normalized);
  }
  return [...result];
}

async function addNamespacesToI18next(language: I18nLanguage, namespaces: string[]): Promise<void> {
  const normalizedNamespaces = normalizeNamespaces(namespaces);
  for (const namespace of normalizedNamespaces) {
    runtimeNamespaces.add(namespace);
  }
  const resourceBundle = await ensureNamespaces(language, normalizedNamespaces);
  for (const [namespace, resources] of Object.entries(resourceBundle)) {
    i18next.addResourceBundle(language, namespace, resources, true, true);
  }
}

async function ensureI18nRuntimeResources(language: I18nLanguage, namespaces: string[]): Promise<void> {
  await addNamespacesToI18next(language, namespaces);
  if (language !== I18N_DEFAULT_LANGUAGE) {
    await addNamespacesToI18next(I18N_DEFAULT_LANGUAGE, namespaces);
  }
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

async function bootstrapI18n(language?: string | null): Promise<typeof i18next> {
  const normalizedLanguage = normalizeI18nLanguage(language);
  if (!initPromise) {
    initPromise = i18next
      .use(LanguageDetector)
      .init({
        resources: {},
        supportedLngs: [...I18N_SUPPORTED_LANGUAGES],
        fallbackLng: I18N_DEFAULT_LANGUAGE,
        lng: normalizedLanguage,
        load: 'currentOnly',
        defaultNS: APP_NAMESPACE,
        ns: [...I18N_BASE_NAMESPACES],
        interpolation: {
          escapeValue: false,
          prefix: '{',
          suffix: '}',
        },
        keySeparator: false,
        nsSeparator: ':',
      })
      .then(() => i18next);
  }

  const instance = await initPromise;
  const knownNamespaces = new Set<string>([
    ...I18N_BASE_NAMESPACES,
    ...getLoadedNamespaces(normalizedLanguage),
    ...runtimeNamespaces,
  ]);
  await ensureI18nRuntimeResources(normalizedLanguage, [...knownNamespaces]);
  bumpRuntimeVersion();
  if (instance.language !== normalizedLanguage) {
    await instance.changeLanguage(normalizedLanguage);
    bumpRuntimeVersion();
  }
  return instance;
}

function hasTranslationKey(value: string): boolean {
  const key = value.trim();
  if (key === '') {
    return false;
  }
  if (i18next.isInitialized) {
    return i18next.exists(key, { ns: namespaceFromKey(key) });
  }
  return false;
}

function collectRouteNamespaces(route: Pick<RouteLocationNormalizedLoaded, 'meta' | 'path'> & {
  name?: RouteLocationNormalizedLoaded['name'] | null;
}): string[] {
  const typedMeta = route.meta as RouteMeta & {
    i18nNamespaces?: string[];
    componentName?: string;
  };

  const namespaces = new Set<string>(collectNamespacesFromRouteMeta(route.meta));
  for (const namespace of typedMeta.i18nNamespaces ?? []) {
    const normalized = namespace.trim().toLowerCase();
    if (normalized !== '' && isValidNamespace(normalized)) {
      namespaces.add(normalized);
    }
  }

  if (typeof typedMeta.componentName === 'string') {
    for (const namespace of namespacesFromComponentName(typedMeta.componentName)) {
      namespaces.add(namespace);
    }
  }

  if (typeof route.name === 'string' && route.name.trim() !== '') {
    for (const namespace of namespacesFromRouteName(route.name)) {
      namespaces.add(namespace);
    }
  }

  if (typeof route.path === 'string' && route.path.trim() !== '') {
    for (const namespace of namespacesFromRoutePath(route.path)) {
      namespaces.add(namespace);
    }
  }

  return [...namespaces];
}

export async function initializeI18n(language?: string | null): Promise<void> {
  await bootstrapI18n(language);
}

export async function setI18nLanguage(language?: string | null): Promise<void> {
  await bootstrapI18n(language);
}

export async function preloadRouteNamespaces(route: Pick<RouteLocationNormalizedLoaded, 'meta' | 'path'> & {
  name?: RouteLocationNormalizedLoaded['name'] | null;
}, languageOverride?: string | null): Promise<void> {
  const localeStore = useLocaleStore();
  const language = normalizeI18nLanguage(languageOverride ?? localeStore.language);
  const namespaces = collectRouteNamespaces(route);
  for (const namespace of namespaces) {
    runtimeNamespaces.add(namespace);
  }
  await ensureI18nRuntimeResources(language, namespaces);
  bumpRuntimeVersion();
}

export function getAppLanguage(): I18nLanguage {
  const localeStore = useLocaleStore();
  return normalizeI18nLanguage(localeStore.language);
}

export function translate(key: string, fallback = '', values?: Record<string, string | number>): string {
  const normalizedKey = key.trim();
  const namespace = namespaceFromKey(normalizedKey);
  const localeStore = useLocaleStore();
  void localeStore.language;
  void runtimeVersion.value;
  if (normalizedKey === '') {
    return fallback.trim() === '' ? '' : formatMessage(fallback, values);
  }

  if (!i18next.isInitialized) {
    if (fallback.trim() !== '') {
      return formatMessage(fallback, values);
    }
    return normalizedKey;
  }

  const translated = i18next.t(normalizedKey, {
    ns: namespace,
    defaultValue: fallback.trim() === '' ? undefined : fallback,
    ...(values ?? {}),
  });

  if (typeof translated === 'string') {
    return translated;
  }

  if (fallback.trim() !== '') {
    return formatMessage(fallback, values);
  }
  return normalizedKey;
}

export function useAppI18n() {
  const localeStore = useLocaleStore();
  const language = computed<I18nLanguage>(() => normalizeI18nLanguage(localeStore.language));

  function t(key: string, fallback = '', values?: Record<string, string | number>): string {
    void language.value;
    void runtimeVersion.value;
    return translate(key, fallback, values);
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
  const key = (meta.titleKey || meta.componentName || meta.title || '').trim();
  const defaultTitle = (meta.titleDefault || meta.title || fallback).trim();

  if (key !== '' && hasTranslationKey(key)) {
    return translate(key, defaultTitle);
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
  const subtitle = subtitleKey.trim() !== '' && hasTranslationKey(subtitleKey.trim())
    ? translate(subtitleKey, subtitleDefault)
    : subtitleDefault;

  return {
    title,
    subtitle,
  };
}
