import type { I18nLanguage } from './language';

import { I18N_DEFAULT_LANGUAGE, I18N_DEFAULT_NS } from './language';

type NamespaceDictionary = Record<string, string>;
type NamespaceResource = Record<string, NamespaceDictionary>;

const loadedNamespaceMap = new Map<I18nLanguage, Set<string>>();
const loadingNamespaceMap = new Map<string, Promise<NamespaceDictionary>>();
const localeModules = import.meta.glob('./locales/*/*.json', { import: 'default' });

function getLoadedSet(language: I18nLanguage): Set<string> {
  const current = loadedNamespaceMap.get(language);
  if (current) {
    return current;
  }
  const created = new Set<string>();
  loadedNamespaceMap.set(language, created);
  return created;
}

function loadingKey(language: I18nLanguage, namespace: string): string {
  return `${language}::${namespace}`;
}

function normalizeNamespacePayload(payload: unknown): NamespaceDictionary {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    return {};
  }

  const raw = payload as Record<string, unknown>;
  const normalized: NamespaceDictionary = {};
  for (const [key, value] of Object.entries(raw)) {
    if (typeof key !== 'string' || key.trim() === '') {
      continue;
    }
    if (typeof value !== 'string' || value.trim() === '') {
      continue;
    }
    normalized[key] = value;
  }
  return normalized;
}

function modulePath(language: I18nLanguage, namespace: string): string {
  return `./locales/${language}/${namespace}.json`;
}

async function loadNamespaceFromFiles(language: I18nLanguage, namespace: string): Promise<NamespaceDictionary> {
  const candidates: I18nLanguage[] = language === I18N_DEFAULT_LANGUAGE
    ? [I18N_DEFAULT_LANGUAGE]
    : [language, I18N_DEFAULT_LANGUAGE];

  for (const candidate of candidates) {
    const path = modulePath(candidate, namespace);
    const resolver = localeModules[path] as (() => Promise<unknown>) | undefined;
    if (!resolver) {
      continue;
    }
    try {
      const payload = await resolver();
      const normalized = normalizeNamespacePayload(payload);
      if (Object.keys(normalized).length > 0) {
        return normalized;
      }
    } catch {
      continue;
    }
  }

  return {};
}

export function hasLoadedNamespace(language: I18nLanguage, namespace: string): boolean {
  return getLoadedSet(language).has(namespace);
}

export function getLoadedNamespaces(language: I18nLanguage): string[] {
  return [...getLoadedSet(language)];
}

export async function loadNamespace(language: I18nLanguage, namespace: string): Promise<NamespaceDictionary> {
  const normalizedNs = namespace.trim().toLowerCase();
  if (normalizedNs === '') {
    return {};
  }

  if (hasLoadedNamespace(language, normalizedNs)) {
    return loadNamespaceFromFiles(language, normalizedNs);
  }

  const key = loadingKey(language, normalizedNs);
  const existing = loadingNamespaceMap.get(key);
  if (existing) {
    return existing;
  }

  const pending = loadNamespaceFromFiles(language, normalizedNs)
    .then((resource) => {
      getLoadedSet(language).add(normalizedNs);
      return resource;
    })
    .finally(() => {
      loadingNamespaceMap.delete(key);
    });

  loadingNamespaceMap.set(key, pending);
  return pending;
}

export async function ensureNamespaces(language: I18nLanguage, namespaces: string[]): Promise<NamespaceResource> {
  const result: NamespaceResource = {};
  for (const namespace of namespaces) {
    const normalizedNs = namespace.trim().toLowerCase();
    if (normalizedNs === '') {
      continue;
    }
    result[normalizedNs] = await loadNamespace(language, normalizedNs);
  }
  if (!result[I18N_DEFAULT_NS]) {
    result[I18N_DEFAULT_NS] = {};
  }
  return result;
}
