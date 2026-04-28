import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { useAppI18n } from '@/i18n';
const TAB_SNAPSHOT_KEY = 'goadmin.workspace.tabs.v1';
const TAB_SNAPSHOT_VERSION = 1;
const DEFAULT_TAB_COMPONENT_KEY = 'route-view';
const DEFAULT_FIXED_ROUTES = new Set(['/dashboard']);
const DEFAULT_NON_TAB_ROUTES = new Set(['/login']);
const DEFAULT_NON_TAB_ROUTE_NAMES = new Set(['login', 'not-found']);
function defaultTabTitle() {
    const { t } = useAppI18n();
    return t('tabs.page', 'Page');
}
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.sessionStorage !== 'undefined';
}
function isPlainObject(value) {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}
function normalizeScalar(value) {
    if (typeof value === 'string') {
        const trimmed = value.trim();
        return trimmed === '' ? null : trimmed;
    }
    if (typeof value === 'number' || typeof value === 'boolean') {
        return String(value);
    }
    return null;
}
function normalizeState(input) {
    if (!isPlainObject(input)) {
        return {};
    }
    const result = {};
    for (const [key, value] of Object.entries(input)) {
        if (Array.isArray(value)) {
            const items = value
                .map((item) => normalizeScalar(item))
                .filter((item) => item !== null);
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
function encodeState(state) {
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
function normalizeComponentKey(componentName, routeName) {
    const candidate = componentName?.trim() || routeName?.trim() || DEFAULT_TAB_COMPONENT_KEY;
    const safeName = candidate
        .replace(/[^A-Za-z0-9_-]+/g, '-')
        .replace(/-+/g, '-')
        .replace(/^[-_]+|[-_]+$/g, '');
    return safeName.length > 0 ? safeName : DEFAULT_TAB_COMPONENT_KEY;
}
function normalizeRouteName(value) {
    if (typeof value !== 'string') {
        return null;
    }
    const trimmed = value.trim();
    return trimmed === '' ? null : trimmed;
}
function makeTabId(routePath, routeFullPath) {
    return routeFullPath.trim() !== '' ? routeFullPath : routePath;
}
function readSnapshot() {
    if (!canUseStorage()) {
        return null;
    }
    const raw = window.sessionStorage.getItem(TAB_SNAPSHOT_KEY);
    if (typeof raw !== 'string' || raw.trim() === '') {
        return null;
    }
    try {
        const parsed = JSON.parse(raw);
        if (parsed.version !== TAB_SNAPSHOT_VERSION || !Array.isArray(parsed.tabs) || typeof parsed.activeId !== 'string') {
            return null;
        }
        return {
            version: TAB_SNAPSHOT_VERSION,
            activeId: parsed.activeId,
            tabs: parsed.tabs.map(normalizeTabRecord).filter((item) => item !== null),
        };
    }
    catch {
        return null;
    }
}
function persistSnapshot(snapshot) {
    if (!canUseStorage()) {
        return;
    }
    window.sessionStorage.setItem(TAB_SNAPSHOT_KEY, JSON.stringify(snapshot));
}
function removeSnapshot() {
    if (!canUseStorage()) {
        return;
    }
    window.sessionStorage.removeItem(TAB_SNAPSHOT_KEY);
}
function normalizeTabRecord(input) {
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
function buildTabFromRoute(route) {
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
function isSameTab(a, b) {
    return a.id === b.id || a.routePath === b.routePath || a.routeFullPath === b.routeFullPath;
}
function resolveFallbackTabIndex(tabs, currentIndex, preferPrevious = true) {
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
function cloneTab(tab) {
    return {
        ...tab,
        query: { ...tab.query },
        params: { ...tab.params },
    };
}
function isRouteAvailable(router, tab) {
    const matched = router.resolve(tab.routeFullPath).matched;
    return matched.some((record) => record.name !== 'not-found');
}
export const useTabsStore = defineStore('tabs', () => {
    const tabs = ref([]);
    const activeId = ref('');
    const hydrated = ref(false);
    const activeTab = computed(() => tabs.value.find((tab) => tab.id === activeId.value) ?? null);
    const tabCount = computed(() => tabs.value.length);
    const fixedTabs = computed(() => tabs.value.filter((tab) => tab.fixed));
    const closableTabs = computed(() => tabs.value.filter((tab) => tab.closable));
    const cacheableTabs = computed(() => tabs.value.filter((tab) => tab.cacheable));
    const cachedViewNames = computed(() => tabs.value.filter((tab) => tab.cacheable).map((tab) => tab.componentKey));
    const visitedIds = computed(() => tabs.value.map((tab) => tab.id));
    function persist() {
        persistSnapshot({
            version: TAB_SNAPSHOT_VERSION,
            activeId: activeId.value,
            tabs: tabs.value.map(cloneTab),
        });
    }
    function upsertTab(route) {
        const incoming = buildTabFromRoute(route);
        if (!incoming) {
            return null;
        }
        const existingIndex = tabs.value.findIndex((tab) => isSameTab(tab, incoming));
        if (existingIndex >= 0) {
            const existing = tabs.value[existingIndex];
            const updated = {
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
    function syncFromRoute(route) {
        const tab = upsertTab(route);
        if (tab) {
            hydrated.value = true;
        }
        return tab;
    }
    function setActiveTab(tabId) {
        const found = tabs.value.find((tab) => tab.id === tabId);
        if (!found) {
            return;
        }
        activeId.value = found.id;
        found.activatedAt = Date.now();
        persist();
    }
    function closeTab(tabId) {
        const index = tabs.value.findIndex((tab) => tab.id === tabId);
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
        if (activeId.value !== '' && !tabs.value.some((tab) => tab.id === activeId.value)) {
            activeId.value = tabs.value[tabs.value.length - 1].id;
        }
        persist();
        return target;
    }
    function closeOthers(tabId) {
        const keep = tabs.value.filter((tab) => tab.fixed || tab.id === tabId);
        tabs.value = keep;
        if (!tabs.value.some((tab) => tab.id === activeId.value)) {
            activeId.value = tabs.value.find((tab) => tab.id === tabId)?.id ?? tabs.value.find((tab) => tab.fixed)?.id ?? tabs.value[0]?.id ?? '';
        }
        persist();
    }
    function closeTabsToLeft(tabId) {
        const index = tabs.value.findIndex((tab) => tab.id === tabId);
        if (index <= 0) {
            return;
        }
        const target = tabs.value[index];
        tabs.value = tabs.value.filter((tab, currentIndex) => tab.fixed || currentIndex >= index || tab.id === target.id);
        if (!tabs.value.some((tab) => tab.id === activeId.value)) {
            activeId.value = target.id;
        }
        persist();
    }
    function closeTabsToRight(tabId) {
        const index = tabs.value.findIndex((tab) => tab.id === tabId);
        if (index < 0 || index >= tabs.value.length - 1) {
            return;
        }
        const target = tabs.value[index];
        tabs.value = tabs.value.filter((tab, currentIndex) => tab.fixed || currentIndex <= index || tab.id === target.id);
        if (!tabs.value.some((tab) => tab.id === activeId.value)) {
            activeId.value = target.id;
        }
        persist();
    }
    function closeAll() {
        tabs.value = tabs.value.filter((tab) => tab.fixed);
        activeId.value = tabs.value[0]?.id ?? '';
        persist();
    }
    function refreshTab(tabId) {
        const tab = tabs.value.find((item) => item.id === tabId);
        if (!tab) {
            return null;
        }
        tab.activatedAt = Date.now();
        persist();
        return tab;
    }
    function hydrate() {
        const snapshot = readSnapshot();
        if (!snapshot) {
            tabs.value = [];
            activeId.value = '';
            hydrated.value = true;
            return;
        }
        const restoredTabs = snapshot.tabs.filter((tab) => tab.restorable);
        tabs.value = restoredTabs;
        activeId.value = restoredTabs.some((tab) => tab.id === snapshot.activeId) ? snapshot.activeId : restoredTabs[0]?.id ?? '';
        hydrated.value = true;
        persist();
    }
    function reconcilePersistedTabs(router) {
        const beforeCount = tabs.value.length;
        const restoredTabs = tabs.value.filter((tab) => tab.restorable && isRouteAvailable(router, tab));
        tabs.value = restoredTabs;
        if (restoredTabs.length === 0) {
            activeId.value = '';
            removeSnapshot();
            return;
        }
        if (!restoredTabs.some((tab) => tab.id === activeId.value)) {
            activeId.value = restoredTabs.find((tab) => tab.fixed)?.id ?? restoredTabs[0].id;
        }
        if (beforeCount !== restoredTabs.length || activeId.value !== '') {
            persist();
        }
    }
    function clearTabs() {
        tabs.value = [];
        activeId.value = '';
        hydrated.value = true;
        removeSnapshot();
    }
    function resolveTabsFromRoute(route) {
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
