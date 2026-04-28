import { I18N_DEFAULT_LANGUAGE, I18N_DEFAULT_NS } from './language';
const loadedNamespaceMap = new Map();
const loadingNamespaceMap = new Map();
const localeModules = import.meta.glob('./locales/*/*.json', { import: 'default' });
function getLoadedSet(language) {
    const current = loadedNamespaceMap.get(language);
    if (current) {
        return current;
    }
    const created = new Set();
    loadedNamespaceMap.set(language, created);
    return created;
}
function loadingKey(language, namespace) {
    return `${language}::${namespace}`;
}
function normalizeNamespacePayload(payload) {
    if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
        return {};
    }
    const raw = payload;
    const normalized = {};
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
function modulePath(language, namespace) {
    return `./locales/${language}/${namespace}.json`;
}
async function loadNamespaceFromFiles(language, namespace) {
    const candidates = language === I18N_DEFAULT_LANGUAGE
        ? [I18N_DEFAULT_LANGUAGE]
        : [language, I18N_DEFAULT_LANGUAGE];
    for (const candidate of candidates) {
        const path = modulePath(candidate, namespace);
        const resolver = localeModules[path];
        if (!resolver) {
            continue;
        }
        try {
            const payload = await resolver();
            const normalized = normalizeNamespacePayload(payload);
            if (Object.keys(normalized).length > 0) {
                return normalized;
            }
        }
        catch {
            continue;
        }
    }
    return {};
}
export function hasLoadedNamespace(language, namespace) {
    return getLoadedSet(language).has(namespace);
}
export function getLoadedNamespaces(language) {
    return [...getLoadedSet(language)];
}
export async function loadNamespace(language, namespace) {
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
export async function ensureNamespaces(language, namespaces) {
    const result = {};
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
