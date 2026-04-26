import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router';

import { translate } from '@/i18n';
import type { WorkspaceTabRecord, WorkspaceTabSnapshot, WorkspaceTabState } from '@/types/tabs';

const TAB_SNAPSHOT_KEY = 'goadmin.workspace.tabs.v1';
const TAB_SNAPSHOT_VERSION = 1 as const;
const DEFAULT_TAB_COMPONENT_KEY = 'route-view';
const DEFAULT_FIXED_ROUTES = new Set(['/dashboard']);
const DEFAULT_NON_TAB_ROUTES = new Set(['/login']);
const DEFAULT_NON_TAB_ROUTE_NAMES = new Set(['login', 'not-found']);

function defaultTabTitle(): string {
  return translate('tabs.page', 'Page');
}

function canUseStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.sessionStorage !== 'undefined';
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function normalizeScalar(value: unknown): string | null {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed === '' ? null : trimmed;
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value);
  }
  return null;
}

function normalizeState(input: unknown): WorkspaceTabState {
  if (!isPlainObject(input)) {
    return {};
  }

  const result: WorkspaceTabState = {};
  for (const [key, value] of Object.entries(input)) {
    if (Array.isArray(value)) {
      const items = value
        .map((item) => normalizeScalar(item))
        .filter((item): item is string => item !== null);
      if (items.length > 0) {
        result[key] = items;
      }
      continue;
    }

    const normalized = normalizeScalar(value);
    if (normalized !== null) {
      result[key] = normalized;
    }
  }
  return result;
}

function encodeState(state: WorkspaceTabState): string {
  const values = Object.entries(state)
    .flatMap(([key, value]) => {
      if (Array.isArray(value)) {
        return value.map((item) => `${key}=${item}`);
      }
      return [`${key}=${value}`];
    })
    .sort();
  return values.join('&');
}

function normalizeComponentKey(componentName: string | undefined, routeName: string | null): string {
  const candidate = componentName?.trim() || routeName?.trim() || DEFAULT_TAB_COMPONENT_KEY;
  const safeName = candidate
    .replace(/[^A-Za-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^[-_]+|[-_]+$/g, '');
  return safeName.length > 0 ? safeName : DEFAULT_TAB_COMPONENT_KEY;
}

function normalizeRouteName(value: unknown): string | null {
  if (typeof value !== 'string') {
    return null;
  }
  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

function makeTabId(routePath: string, routeFullPath: string): string {
  return routeFullPath.trim() !== '' ? routeFullPath : routePath;
}

function readSnapshot(): WorkspaceTabSnapshot | null {
  if (!canUseStorage()) {
    return null;
  }

  const raw = window.sessionStorage.getItem(TAB_SNAPSHOT_KEY);
  if (typeof raw !== 'string' || raw.trim() === '') {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<WorkspaceTabSnapshot>;
    if (parsed.version !== TAB_SNAPSHOT_VERSION || !Array.isArray(parsed.tabs) || typeof parsed.activeId !== 'string') {
      return null;
    }
    return {
      version: TAB_SNAPSHOT_VERSION,
      activeId: parsed.activeId,
      tabs: parsed.tabs.map(normalizeTabRecord).filter((item): item is WorkspaceTabRecord => item !== null),
    };
  } catch {
    return null;
  }
}

function persistSnapshot(snapshot: WorkspaceTabSnapshot): void {
  if (!canUseStorage()) {
    return;
  }
  window.sessionStorage.setItem(TAB_SNAPSHOT_KEY, JSON.stringify(snapshot));
}

function removeSnapshot(): void {
  if (!canUseStorage()) {
    return;
  }
  window.sessionStorage.removeItem(TAB_SNAPSHOT_KEY);
}

function normalizeTabRecord(input: unknown): WorkspaceTabRecord | null {
  if (!isPlainObject(input)) {
    return null;
  }

  const routePath = typeof input.routePath === 'string' ? input.routePath.trim() : '';
  const routeFullPath = typeof input.routeFullPath === 'string' ? input.routeFullPath.trim() : '';
  const title = typeof input.title === 'string' && input.title.trim() !== '' ? input.title.trim() : defaultTabTitle();
  const titleKey = typeof input.titleKey === 'string' && input.titleKey.trim() !== '' ? input.titleKey.trim() : undefined;
  const titleDefault = typeof input.titleDefault === 'string' && input.titleDefault.trim() !== '' ? input.titleDefault.trim() : undefined;
  const componentKey = normalizeComponentKey(typeof input.componentKey === 'string' ? input.componentKey : undefined, normalizeRouteName(input.routeName));
  const fixed = input.fixed === true;
  const closable = input.closable !== false && !fixed;
  const cacheable = input.cacheable === true;
  const restorable = input.restorable !== false;
  const openedAt = typeof input.openedAt === 'number' && Number.isFinite(input.openedAt) ? input.openedAt : Date.now();
  const activatedAt = typeof input.activatedAt === 'number' && Number.isFinite(input.activatedAt) ? input.activatedAt : openedAt;

  if (routePath === '' || routeFullPath === '') {
    return null;
  }

  return {
    id: typeof input.id === 'string' && input.id.trim() !== '' ? input.id.trim() : makeTabId(routePath, routeFullPath),
    routeName: normalizeRouteName(input.routeName),
    routePath,
    routeFullPath,
    title,
    titleKey,
    titleDefault,
    icon: typeof input.icon === 'string' && input.icon.trim() !== '' ? input.icon.trim() : null,
    componentKey,
    fixed,
    closable,
    cacheable,
    restorable,
    query: normalizeState(input.query),
    params: normalizeState(input.params),
    openedAt,
    activatedAt,
  };
}

function buildTabFromRoute(route: RouteLocationNormalizedLoaded): WorkspaceTabRecord | null {
  if (route.meta.public === true) {
    return null;
  }
  if (DEFAULT_NON_TAB_ROUTES.has(route.path)) {
    return null;
  }
  if (typeof route.name === 'string' && DEFAULT_NON_TAB_ROUTE_NAMES.has(route.name)) {
    return null;
  }

  const routeName = normalizeRouteName(route.name);
  const title = typeof route.meta.title === 'string' && route.meta.title.trim() !== '' ? route.meta.title.trim() : defaultTabTitle();
  const titleKey = typeof route.meta.titleKey === 'string' && route.meta.titleKey.trim() !== '' ? route.meta.titleKey.trim() : undefined;
  const titleDefault = typeof route.meta.titleDefault === 'string' && route.meta.titleDefault.trim() !== ''
    ? route.meta.titleDefault.trim()
    : title;
  const icon = typeof route.meta.icon === 'string' && route.meta.icon.trim() !== '' ? route.meta.icon.trim() : null;
  const fixed = route.meta.affix === true || DEFAULT_FIXED_ROUTES.has(route.path);
  const cacheable = route.meta.keepAlive !== false;
  const restorable = route.meta.restorable !== false;
  const componentKey = normalizeComponentKey(route.meta.componentName, routeName);
  const routeFullPath = route.fullPath.trim() !== '' ? route.fullPath : route.path;
  const routePath = route.path;

  return {
    id: makeTabId(routePath, routeFullPath),
    routeName,
    routePath,
    routeFullPath,
    title,
    titleKey,
    titleDefault,
    icon,
    componentKey,
    fixed,
    closable: !fixed,
    cacheable,
    restorable,
    query: normalizeState(route.query),
    params: normalizeState(route.params),
    openedAt: Date.now(),
    activatedAt: Date.now(),
  };
}

function isSameTab(a: WorkspaceTabRecord, b: WorkspaceTabRecord): boolean {
  return a.id === b.id || a.routePath === b.routePath || a.routeFullPath === b.routeFullPath;
}

function resolveFallbackTabIndex(tabs: WorkspaceTabRecord[], currentIndex: number, preferPrevious = true): number {
  if (tabs.length === 0) {
    return -1;
  }
  if (currentIndex < 0) {
    return Math.max(0, tabs.length - 1);
  }
  if (preferPrevious) {
    return currentIndex > 0 ? currentIndex - 1 : Math.min(1, tabs.length - 1);
  }
  return currentIndex < tabs.length - 1 ? currentIndex + 1 : Math.max(0, currentIndex - 1);
}

function cloneTab(tab: WorkspaceTabRecord): WorkspaceTabRecord {
  return {
    ...tab,
    query: { ...tab.query },
    params: { ...tab.params },
  };
}

function isRouteAvailable(router: Router, tab: WorkspaceTabRecord): boolean {
  const matched = router.resolve(tab.routeFullPath).matched;
  return matched.some((record: (typeof matched)[number]) => record.name !== 'not-found');
}

export const useTabsStore = defineStore('tabs', () => {
  const tabs = ref<WorkspaceTabRecord[]>([]);
  const activeId = ref('');
  const hydrated = ref(false);

  const activeTab = computed<WorkspaceTabRecord | null>(() => tabs.value.find((tab: WorkspaceTabRecord) => tab.id === activeId.value) ?? null);
  const tabCount = computed(() => tabs.value.length);
  const fixedTabs = computed<WorkspaceTabRecord[]>(() => tabs.value.filter((tab: WorkspaceTabRecord) => tab.fixed));
  const closableTabs = computed<WorkspaceTabRecord[]>(() => tabs.value.filter((tab: WorkspaceTabRecord) => tab.closable));
  const cacheableTabs = computed<WorkspaceTabRecord[]>(() => tabs.value.filter((tab: WorkspaceTabRecord) => tab.cacheable));
  const cachedViewNames = computed<string[]>(() => tabs.value.filter((tab: WorkspaceTabRecord) => tab.cacheable).map((tab: WorkspaceTabRecord) => tab.componentKey));
  const visitedIds = computed<string[]>(() => tabs.value.map((tab: WorkspaceTabRecord) => tab.id));

  function persist(): void {
    persistSnapshot({
      version: TAB_SNAPSHOT_VERSION,
      activeId: activeId.value,
      tabs: tabs.value.map(cloneTab),
    });
  }

  function upsertTab(route: RouteLocationNormalizedLoaded): WorkspaceTabRecord | null {
    const incoming = buildTabFromRoute(route);
    if (!incoming) {
      return null;
    }

    const existingIndex = tabs.value.findIndex((tab: WorkspaceTabRecord) => isSameTab(tab, incoming));
    if (existingIndex >= 0) {
      const existing = tabs.value[existingIndex];
      const updated: WorkspaceTabRecord = {
        ...existing,
        title: incoming.title || existing.title,
        icon: incoming.icon ?? existing.icon,
        componentKey: incoming.componentKey || existing.componentKey,
        fixed: existing.fixed || incoming.fixed,
        closable: existing.fixed ? false : incoming.closable,
        cacheable: incoming.cacheable,
        restorable: incoming.restorable,
        query: incoming.query,
        params: incoming.params,
        activatedAt: Date.now(),
      };
      tabs.value.splice(existingIndex, 1, updated);
      activeId.value = updated.id;
      persist();
      return updated;
    }

    tabs.value.push(incoming);
    activeId.value = incoming.id;
    persist();
    return incoming;
  }

  function syncFromRoute(route: RouteLocationNormalizedLoaded): WorkspaceTabRecord | null {
    const tab = upsertTab(route);
    if (tab) {
      hydrated.value = true;
    }
    return tab;
  }

  function setActiveTab(tabId: string): void {
    const found = tabs.value.find((tab: WorkspaceTabRecord) => tab.id === tabId);
    if (!found) {
      return;
    }
    activeId.value = found.id;
    found.activatedAt = Date.now();
    persist();
  }

  function closeTab(tabId: string): WorkspaceTabRecord | null {
    const index = tabs.value.findIndex((tab: WorkspaceTabRecord) => tab.id === tabId);
    if (index < 0) {
      return null;
    }

    const target = tabs.value[index];
    if (!target.closable) {
      return null;
    }

    const wasActive = target.id === activeId.value;
    tabs.value.splice(index, 1);

    if (tabs.value.length === 0) {
      activeId.value = '';
      persist();
      return target;
    }

    if (wasActive) {
      const fallbackIndex = resolveFallbackTabIndex(tabs.value, index, true);
      if (fallbackIndex >= 0) {
        activeId.value = tabs.value[fallbackIndex].id;
      }
    }

    if (activeId.value !== '' && !tabs.value.some((tab: WorkspaceTabRecord) => tab.id === activeId.value)) {
      activeId.value = tabs.value[tabs.value.length - 1].id;
    }

    persist();
    return target;
  }

  function closeOthers(tabId: string): void {
    const keep = tabs.value.filter((tab: WorkspaceTabRecord) => tab.fixed || tab.id === tabId);
    tabs.value = keep;
    if (!tabs.value.some((tab: WorkspaceTabRecord) => tab.id === activeId.value)) {
      activeId.value = tabs.value.find((tab: WorkspaceTabRecord) => tab.id === tabId)?.id ?? tabs.value.find((tab: WorkspaceTabRecord) => tab.fixed)?.id ?? tabs.value[0]?.id ?? '';
    }
    persist();
  }

  function closeTabsToLeft(tabId: string): void {
    const index = tabs.value.findIndex((tab: WorkspaceTabRecord) => tab.id === tabId);
    if (index <= 0) {
      return;
    }
    const target = tabs.value[index];
    tabs.value = tabs.value.filter((tab: WorkspaceTabRecord, currentIndex: number) => tab.fixed || currentIndex >= index || tab.id === target.id);
    if (!tabs.value.some((tab: WorkspaceTabRecord) => tab.id === activeId.value)) {
      activeId.value = target.id;
    }
    persist();
  }

  function closeTabsToRight(tabId: string): void {
    const index = tabs.value.findIndex((tab: WorkspaceTabRecord) => tab.id === tabId);
    if (index < 0 || index >= tabs.value.length - 1) {
      return;
    }
    const target = tabs.value[index];
    tabs.value = tabs.value.filter((tab: WorkspaceTabRecord, currentIndex: number) => tab.fixed || currentIndex <= index || tab.id === target.id);
    if (!tabs.value.some((tab: WorkspaceTabRecord) => tab.id === activeId.value)) {
      activeId.value = target.id;
    }
    persist();
  }

  function closeAll(): void {
    tabs.value = tabs.value.filter((tab: WorkspaceTabRecord) => tab.fixed);
    activeId.value = tabs.value[0]?.id ?? '';
    persist();
  }

  function refreshTab(tabId: string): WorkspaceTabRecord | null {
    const tab = tabs.value.find((item: WorkspaceTabRecord) => item.id === tabId);
    if (!tab) {
      return null;
    }
    tab.activatedAt = Date.now();
    persist();
    return tab;
  }

  function hydrate(): void {
    const snapshot = readSnapshot();
    if (!snapshot) {
      tabs.value = [];
      activeId.value = '';
      hydrated.value = true;
      return;
    }

    const restoredTabs = snapshot.tabs.filter((tab: WorkspaceTabRecord) => tab.restorable);
    tabs.value = restoredTabs;
    activeId.value = restoredTabs.some((tab: WorkspaceTabRecord) => tab.id === snapshot.activeId) ? snapshot.activeId : restoredTabs[0]?.id ?? '';
    hydrated.value = true;
    persist();
  }

  function reconcilePersistedTabs(router: Router): void {
    const beforeCount = tabs.value.length;
    const restoredTabs = tabs.value.filter((tab: WorkspaceTabRecord) => tab.restorable && isRouteAvailable(router, tab));

    tabs.value = restoredTabs;
    if (restoredTabs.length === 0) {
      activeId.value = '';
      removeSnapshot();
      return;
    }

    if (!restoredTabs.some((tab: WorkspaceTabRecord) => tab.id === activeId.value)) {
      activeId.value = restoredTabs.find((tab: WorkspaceTabRecord) => tab.fixed)?.id ?? restoredTabs[0].id;
    }

    if (beforeCount !== restoredTabs.length || activeId.value !== '') {
      persist();
    }
  }

  function clearTabs(): void {
    tabs.value = [];
    activeId.value = '';
    hydrated.value = true;
    removeSnapshot();
  }

  function resolveTabsFromRoute(route: RouteLocationNormalizedLoaded): WorkspaceTabRecord | null {
    return syncFromRoute(route);
  }

  return {
    tabs,
    activeId,
    hydrated,
    activeTab,
    tabCount,
    fixedTabs,
    closableTabs,
    cacheableTabs,
    cachedViewNames,
    visitedIds,
    hydrate,
    reconcilePersistedTabs,
    clearTabs,
    syncFromRoute,
    setActiveTab,
    closeTab,
    closeOthers,
    closeTabsToLeft,
    closeTabsToRight,
    closeAll,
    refreshTab,
    resolveTabsFromRoute,
  };
});
