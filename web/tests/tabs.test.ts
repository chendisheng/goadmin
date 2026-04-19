// @vitest-environment jsdom
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { createMemoryHistory, createRouter, defineComponent, type RouteRecordRaw } from 'vue-router';

import { useTabsStore } from '../src/store/tabs';
import type { WorkspaceTabRecord, WorkspaceTabSnapshot } from '../src/types/tabs';

const SNAPSHOT_KEY = 'goadmin.workspace.tabs.v1';

const TestView = defineComponent({
  name: 'TestView',
  setup() {
    return () => null;
  },
});

function createTestRouter() {
  const routes: RouteRecordRaw[] = [
    {
      path: '/login',
      name: 'login',
      component: TestView,
      meta: {
        title: '登录',
        public: true,
        requiresAuth: false,
      },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: TestView,
      meta: {
        title: '工作台',
        affix: true,
        keepAlive: true,
        restorable: true,
        componentName: 'view/dashboard/index',
      },
    },
    {
      path: '/system/user',
      name: 'system-user',
      component: TestView,
      meta: {
        title: '用户管理',
        keepAlive: true,
        restorable: true,
        componentName: 'view/system/user/index',
      },
    },
    {
      path: '/system/report',
      name: 'system-report',
      component: TestView,
      meta: {
        title: '报表页',
        keepAlive: false,
        restorable: true,
        componentName: 'view/system/report/index',
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: TestView,
      meta: {
        title: '404',
        public: true,
      },
    },
  ];

  return createRouter({
    history: createMemoryHistory(),
    routes,
  });
}

function createSnapshotTab(
  overrides: Partial<WorkspaceTabSnapshot['tabs'][number]> & Pick<WorkspaceTabSnapshot['tabs'][number], 'id' | 'routePath' | 'routeFullPath' | 'title' | 'componentKey'>,
): WorkspaceTabSnapshot['tabs'][number] {
  return {
    routeName: null,
    icon: null,
    fixed: false,
    closable: true,
    cacheable: true,
    restorable: true,
    query: {},
    params: {},
    openedAt: 1,
    activatedAt: 1,
    ...overrides,
  };
}

beforeEach(() => {
  sessionStorage.clear();
  setActivePinia(createPinia());
});

describe('useTabsStore', () => {
  it('ignores public and 404 routes while syncing eligible routes into the cache list', () => {
    const router = createTestRouter();
    const tabsStore = useTabsStore();

    expect(tabsStore.syncFromRoute(router.resolve('/login'))).toBeNull();
    expect(tabsStore.syncFromRoute(router.resolve('/missing'))).toBeNull();

    const dashboardTab = tabsStore.syncFromRoute(router.resolve('/dashboard'));
    const userTab = tabsStore.syncFromRoute(router.resolve('/system/user'));
    const reportTab = tabsStore.syncFromRoute(router.resolve('/system/report'));

    expect(dashboardTab?.fixed).toBe(true);
    expect(dashboardTab?.closable).toBe(false);
    expect(userTab?.cacheable).toBe(true);
    expect(reportTab?.cacheable).toBe(false);
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard', '/system/user', '/system/report']);
    expect(tabsStore.cachedViewNames).toEqual(['view-dashboard-index', 'view-system-user-index']);
  });

  it('hydrates persisted tabs and reconciles them against the live router before restore', () => {
    const router = createTestRouter();
    const snapshot: WorkspaceTabSnapshot = {
      version: 1,
      activeId: '/legacy',
      tabs: [
        createSnapshotTab({
          id: '/dashboard',
          routeName: 'dashboard',
          routePath: '/dashboard',
          routeFullPath: '/dashboard',
          title: '工作台',
          componentKey: 'view-dashboard-index',
          fixed: true,
          closable: false,
          cacheable: true,
          restorable: true,
        }),
        createSnapshotTab({
          id: '/system/user',
          routeName: 'system-user',
          routePath: '/system/user',
          routeFullPath: '/system/user',
          title: '用户管理',
          componentKey: 'view-system-user-index',
        }),
        createSnapshotTab({
          id: '/legacy',
          routeName: 'legacy-page',
          routePath: '/legacy',
          routeFullPath: '/legacy',
          title: '历史页面',
          componentKey: 'view-legacy-page',
        }),
      ],
    };

    sessionStorage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot));

    const tabsStore = useTabsStore();
    tabsStore.hydrate();

    expect(tabsStore.activeId).toBe('/legacy');
    expect(tabsStore.tabs).toHaveLength(3);

    tabsStore.reconcilePersistedTabs(router);

    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard', '/system/user']);
    expect(tabsStore.activeId).toBe('/dashboard');
    expect(tabsStore.cachedViewNames).toEqual(['view-dashboard-index', 'view-system-user-index']);
  });

  it('protects fixed tabs and falls back to the remaining page after closing the active tab', () => {
    const router = createTestRouter();
    const tabsStore = useTabsStore();

    const dashboardTab = tabsStore.syncFromRoute(router.resolve('/dashboard'));
    const userTab = tabsStore.syncFromRoute(router.resolve('/system/user'));

    expect(dashboardTab?.id).toBe('/dashboard');
    expect(userTab?.id).toBe('/system/user');

    expect(tabsStore.closeTab('/dashboard')).toBeNull();
    expect(tabsStore.activeId).toBe('/system/user');

    const closedUser = tabsStore.closeTab('/system/user');

    expect(closedUser?.id).toBe('/system/user');
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard']);
    expect(tabsStore.activeId).toBe('/dashboard');
  });

  it('keeps fixed tabs during batch close operations and refreshes timestamps in place', () => {
    const router = createTestRouter();
    const tabsStore = useTabsStore();

    const dashboardTab = tabsStore.syncFromRoute(router.resolve('/dashboard'));
    const userTab = tabsStore.syncFromRoute(router.resolve('/system/user'));
    const reportTab = tabsStore.syncFromRoute(router.resolve('/system/report'));

    expect(dashboardTab?.id).toBe('/dashboard');
    expect(userTab?.id).toBe('/system/user');
    expect(reportTab?.id).toBe('/system/report');

    const beforeRefresh = reportTab?.activatedAt ?? 0;
    const refreshed = tabsStore.refreshTab('/system/report');
    expect(refreshed?.activatedAt).toBeGreaterThanOrEqual(beforeRefresh);

    tabsStore.closeTabsToLeft('/system/user');
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard', '/system/user', '/system/report']);

    tabsStore.closeTabsToRight('/system/user');
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard', '/system/user']);

    tabsStore.closeOthers('/system/user');
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard', '/system/user']);

    tabsStore.closeAll();
    expect(tabsStore.tabs.map((tab: WorkspaceTabRecord) => tab.routePath)).toEqual(['/dashboard']);
    expect(tabsStore.activeId).toBe('/dashboard');
  });
});
