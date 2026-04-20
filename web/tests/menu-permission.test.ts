// @vitest-environment jsdom
import { describe, expect, it } from 'vitest';

import { buildMenusOnly, filterMenuRoutesByPermission } from '../src/router/navigation';
import type { BackendMenuRoute } from '../src/types/menu';

function createMenuRoute(overrides: Partial<BackendMenuRoute>): BackendMenuRoute {
  return {
    name: 'default',
    path: '/',
    component: 'Layout',
    redirect: '',
    hidden: false,
    alwaysShow: false,
    meta: {
      title: '默认菜单',
      permission: undefined,
      hidden: false,
      noCache: false,
      affix: false,
      link: '',
    },
    children: [],
    ...overrides,
  };
}

const menuTree: BackendMenuRoute[] = [
  createMenuRoute({
    name: 'dashboard',
    path: '/dashboard',
    component: 'view/dashboard/index',
    meta: {
      title: '工作台',
      permission: 'dashboard:view',
      hidden: false,
      noCache: false,
      affix: true,
      link: '',
    },
  }),
  createMenuRoute({
    name: 'system',
    path: '/system',
    component: 'Layout',
    meta: {
      title: '系统管理',
      permission: 'system:view',
      hidden: false,
      noCache: false,
      affix: false,
      link: '',
    },
    children: [
      createMenuRoute({
        name: 'system-user',
        path: '/system/user',
        component: 'view/system/user/index',
        meta: {
          title: '用户管理',
          permission: 'user:list',
          hidden: false,
          noCache: false,
          affix: false,
          link: '',
        },
      }),
      createMenuRoute({
        name: 'system-user-create-btn',
        path: '/system/user/create',
        component: '',
        type: 'button',
        meta: {
          title: '新增用户',
          permission: 'user:create',
          hidden: false,
          noCache: true,
          affix: false,
          link: '',
        },
      }),
      createMenuRoute({
        name: 'system-role',
        path: '/system/role',
        component: 'view/system/role/index',
        meta: {
          title: '角色管理',
          permission: 'role:list',
          hidden: false,
          noCache: false,
          affix: false,
          link: '',
        },
      }),
    ],
  }),
  createMenuRoute({
    name: 'audit',
    path: '/audit',
    component: 'view/audit/index',
    meta: {
      title: '审计中心',
      permission: 'audit:view',
      hidden: false,
      noCache: false,
      affix: false,
      link: '',
    },
  }),
];

describe('menu permission filtering', () => {
  it('hides unauthorized menus the same way buttons are hidden', () => {
    const canAccess = (permission: string) => ['dashboard:view', 'system:view', 'user:list'].includes(permission);

    const filtered = filterMenuRoutesByPermission(menuTree, canAccess);
    const sidebarMenus = buildMenusOnly(menuTree, canAccess);

    expect(filtered).toHaveLength(2);
    expect(filtered[0].name).toBe('dashboard');
    expect(filtered[1].name).toBe('system');
    expect(filtered[1].hidden).toBe(true);
    expect(filtered[1].children).toHaveLength(1);
    expect(filtered[1].children[0].name).toBe('system-user');
    expect(filtered[1].children[0].hidden).toBe(false);

    expect(sidebarMenus.map((item) => item.path)).toEqual(['/dashboard', '/system/user']);
    expect(sidebarMenus.map((item) => item.title)).toEqual(['工作台', '用户管理']);
  });

  it('ignores button-type menu children when building page routes', () => {
    const canAccess = () => true;

    const filtered = filterMenuRoutesByPermission(menuTree, canAccess);
    const systemRoute = filtered.find((item) => item.name === 'system');
    const childNames = systemRoute?.children.map((child: BackendMenuRoute) => child.name) ?? [];

    expect(systemRoute).toBeDefined();
    expect(childNames).toEqual(['system-user', 'system-role']);
    expect(systemRoute?.children.some((child: BackendMenuRoute) => child.type === 'button')).toBe(false);
  });
});
