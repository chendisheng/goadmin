import type { RouteMeta } from 'vue-router';

import { I18N_DEFAULT_NS } from './language';

export const I18N_BASE_NAMESPACES = [I18N_DEFAULT_NS, 'common', 'route'] as const;

const NAMESPACE_PATTERN = /^[a-z0-9][a-z0-9_-]*$/i;

function normalizeNamespaceCandidate(value: string): string {
  return value.trim().toLowerCase();
}

export function isValidNamespace(value: string): boolean {
  return NAMESPACE_PATTERN.test(value.trim());
}

export function namespaceFromKey(key: string): string {
  const normalizedKey = key.trim();
  if (normalizedKey === '' || !normalizedKey.includes('.')) {
    return I18N_DEFAULT_NS;
  }
  const token = normalizeNamespaceCandidate(normalizedKey.split('.')[0] || '');
  return token !== '' && isValidNamespace(token) ? token : I18N_DEFAULT_NS;
}

export function collectNamespacesFromKeys(keys: Array<string | null | undefined>): string[] {
  const result = new Set<string>(I18N_BASE_NAMESPACES);
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

export function namespacesFromComponentName(componentName: string): string[] {
  const normalized = componentName.trim();
  if (normalized === '' || normalized === 'Layout') {
    return [];
  }

  const segments = normalized.split('/').filter(Boolean);
  if (segments.length === 0) {
    return [];
  }

  const viewIndex = segments[0] === 'view' || segments[0] === 'views' ? 1 : 0;
  if (viewIndex >= segments.length) {
    return [];
  }

  const namespace = segments[viewIndex] === 'system' && viewIndex + 1 < segments.length
    ? segments[viewIndex + 1]
    : segments[viewIndex];

  const normalizedNamespace = namespace.trim().toLowerCase();
  return normalizedNamespace === '' ? [] : [normalizedNamespace];
}

function singularizeNamespaceCandidate(value: string): string {
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

function splitRouteTokens(value: string): string[] {
  return value
    .trim()
    .split(/[^A-Za-z0-9]+|(?<=[a-z0-9])(?=[A-Z])/)
    .map((token) => token.trim())
    .filter(Boolean);
}

export function namespacesFromRouteName(routeName: string): string[] {
  const tokens = splitRouteTokens(routeName);
  if (tokens.length === 0) {
    return [];
  }

  const candidateIndex = tokens[0]?.toLowerCase() === 'system' && tokens.length > 1 ? 1 : 0;
  const candidate = singularizeNamespaceCandidate(tokens[candidateIndex] || '');
  return candidate !== '' && isValidNamespace(candidate) ? [candidate] : [];
}

export function namespacesFromRoutePath(routePath: string): string[] {
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

export function collectNamespacesFromRouteMeta(meta: RouteMeta): string[] {
  const typedMeta = meta as RouteMeta & {
    titleKey?: string;
    subtitleKey?: string;
    componentName?: string;
    title?: string;
    i18nNamespaces?: string[];
  };
  const namespaces = new Set<string>(collectNamespacesFromKeys([
    typeof typedMeta.titleKey === 'string' ? typedMeta.titleKey : '',
    typeof typedMeta.subtitleKey === 'string' ? typedMeta.subtitleKey : '',
    typeof typedMeta.componentName === 'string' ? typedMeta.componentName : '',
    typeof typedMeta.title === 'string' ? typedMeta.title : '',
  ]));

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

  return [...namespaces];
}
