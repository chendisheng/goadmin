import { I18N_DEFAULT_NS } from './language';
export const I18N_BASE_NAMESPACES = [I18N_DEFAULT_NS, 'common', 'route'];
const NAMESPACE_PATTERN = /^[a-z0-9][a-z0-9_-]*$/i;
function normalizeNamespaceCandidate(value) {
    return value.trim().toLowerCase();
}
export function isValidNamespace(value) {
    return NAMESPACE_PATTERN.test(value.trim());
}
export function namespaceFromKey(key) {
    const normalizedKey = key.trim();
    if (normalizedKey === '' || !normalizedKey.includes('.')) {
        return I18N_DEFAULT_NS;
    }
    const token = normalizeNamespaceCandidate(normalizedKey.split('.')[0] || '');
    return token !== '' && isValidNamespace(token) ? token : I18N_DEFAULT_NS;
}
export function collectNamespacesFromKeys(keys) {
    const result = new Set(I18N_BASE_NAMESPACES);
    for (const key of keys) {
        if (typeof key !== 'string') {
            continue;
        }
        const normalizedKey = key.trim();
        if (normalizedKey === '' || !normalizedKey.includes('.')) {
            continue;
        }
        result.add(namespaceFromKey(normalizedKey));
    }
    return [...result];
}
export function namespacesFromComponentName(componentName) {
    const normalized = componentName.trim();
    if (normalized === '' || normalized === 'Layout') {
        return [];
    }
    const normalizedNamespace = namespaceFromKey(normalized).trim().toLowerCase();
    return normalizedNamespace === '' ? [] : [normalizedNamespace];
}
function singularizeNamespaceCandidate(value) {
    const normalized = value.trim().toLowerCase();
    if (normalized === '') {
        return '';
    }
    if (normalized.endsWith('ies') && normalized.length > 3) {
        return normalized.slice(0, -3) + 'y';
    }
    if (normalized.endsWith('s') && !normalized.endsWith('ss') && normalized.length > 1) {
        return normalized.slice(0, -1);
    }
    return normalized;
}
function splitRouteTokens(value) {
    return value
        .trim()
        .split(/[^A-Za-z0-9]+|(?<=[a-z0-9])(?=[A-Z])/)
        .map((token) => token.trim())
        .filter(Boolean);
}
export function namespacesFromRouteName(routeName) {
    const tokens = splitRouteTokens(routeName);
    if (tokens.length === 0) {
        return [];
    }
    const candidateIndex = tokens[0]?.toLowerCase() === 'system' && tokens.length > 1 ? 1 : 0;
    const candidate = singularizeNamespaceCandidate(tokens[candidateIndex] || '');
    return candidate !== '' && isValidNamespace(candidate) ? [candidate] : [];
}
export function namespacesFromRoutePath(routePath) {
    const segments = routePath
        .trim()
        .split('/')
        .map((segment) => segment.trim())
        .filter(Boolean)
        .filter((segment) => segment !== ':name' && segment !== ':id' && !segment.startsWith(':') && segment !== '*');
    if (segments.length === 0) {
        return [];
    }
    const candidateIndex = segments[0]?.toLowerCase() === 'system' && segments.length > 1 ? 1 : 0;
    const candidate = singularizeNamespaceCandidate(segments[candidateIndex] || '');
    return candidate !== '' && isValidNamespace(candidate) ? [candidate] : [];
}
export function collectNamespacesFromRouteMeta(meta) {
    const typedMeta = meta;
    const namespaces = collectNamespacesFromKeys([
        typeof typedMeta.titleKey === 'string' ? typedMeta.titleKey : '',
        typeof typedMeta.subtitleKey === 'string' ? typedMeta.subtitleKey : '',
        typeof typedMeta.componentName === 'string' ? typedMeta.componentName : '',
        typeof typedMeta.title === 'string' ? typedMeta.title : '',
    ]);
    return namespaces;
}
