import { defineAsyncComponent } from 'vue';
import type { AsyncComponentLoader, Component } from 'vue';
import type { RouteRecordRaw, Router } from 'vue-router';

import RouteGroupView from '@/views/RouteGroupView.vue';
import RoutePlaceholderView from '@/views/RoutePlaceholderView.vue';
import type { BackendMenuRoute, SidebarMenuNode } from '@/types/menu';
import type { PluginMenu } from '@/types/plugin';

type ViewModuleLoader = AsyncComponentLoader<Component>;

const viewModules = import.meta.glob('../views/**/*.vue') as Record<string, ViewModuleLoader>;

export function normalizeMenuRoots(nodes: BackendMenuRoute[]): BackendMenuRoute[] {
  return nodes;
}

function componentNameToModulePath(componentName: string): string | null {
  const normalized = componentName.trim();
  if (normalized === '') {
    return null;
  }
  if (normalized === 'view/dashboard/index') {
    return '../views/DashboardView.vue';
  }
  if (normalized.startsWith('view/')) {
    return `../views/${normalized.slice(5)}.vue`;
  }
  if (normalized.startsWith('system/')) {
    return `../views/system/${normalized.slice(7)}.vue`;
  }
  return null;
}

function normalizePath(value: string): string {
  const trimmed = value.trim();
  if (trimmed === '') {
    return '/';
  }
  if (trimmed === '/') {
    return '/';
  }
  const prefixed = trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
  return prefixed.length > 1 ? prefixed.replace(/\/+$/, '') : prefixed;
}

function relativePath(value: string, parentPath: string): string {
  const current = normalizePath(value);
  const parent = normalizePath(parentPath);
  if (parent === '/') {
    return current === '/' ? '' : current.replace(/^\//, '');
  }
  if (current === parent) {
    return '';
  }
  const prefix = parent.endsWith('/') ? parent : `${parent}/`;
  if (current.startsWith(prefix)) {
    return current.slice(prefix.length);
  }
  return current.replace(/^\//, '');
}

function buildSidebarNodes(nodes: BackendMenuRoute[]): SidebarMenuNode[] {
  const roots = normalizeMenuRoots(nodes);
  const result: SidebarMenuNode[] = [];
  for (const node of roots) {
    const children = buildSidebarNodes(node.children ?? []);
    if (node.hidden) {
      result.push(...children);
      continue;
    }
    result.push({
      name: node.name,
      path: normalizePath(node.path),
      title: node.meta.title,
      icon: node.meta.icon,
      component: node.component,
      redirect: node.redirect,
      permission: node.meta.permission,
      hidden: node.hidden,
      children,
    });
  }
  return result;
}

function resolveLeafComponent(componentName: string | undefined): Component {
  const normalized = (componentName || '').trim();
  if (normalized === '') {
    return RoutePlaceholderView;
  }
  if (normalized === 'Layout') {
    return RouteGroupView;
  }
  const modulePath = componentNameToModulePath(normalized);
  if (modulePath && viewModules[modulePath]) {
    return defineAsyncComponent(viewModules[modulePath]);
  }
  return RoutePlaceholderView;
}

function buildRouteRecord(node: BackendMenuRoute, parentPath = '/'): RouteRecordRaw {
  const currentPath = normalizePath(node.path);
  const childPath = relativePath(currentPath, parentPath);
  const children = (node.children ?? []).map((child) => buildRouteRecord(child, currentPath));
  const hasChildren = children.length > 0;
  const componentName = (node.component || '').trim();

  const record: RouteRecordRaw = {
    path: childPath,
    name: node.name,
    component: componentName === 'Layout' || hasChildren ? RouteGroupView : resolveLeafComponent(componentName),
    redirect: node.redirect || undefined,
    meta: {
      title: node.meta.title,
      icon: node.meta.icon,
      componentName,
      permission: node.meta.permission,
      link: node.meta.link,
      hidden: node.hidden,
      inMenu: !node.hidden,
      hideInMenu: node.hidden,
      requiresAuth: true,
      alwaysShow: node.alwaysShow ?? hasChildren,
      order: node.meta.affix ? 0 : 10,
    },
    children: hasChildren ? children : [],
  };

  return record;
}

export function registerBackendRoutes(router: Router, nodes: BackendMenuRoute[]): string[] {
  const roots = normalizeMenuRoots(nodes);
  const routeNames: string[] = [];

  for (const item of roots) {
    const record = buildRouteRecord(item, '/');
    const routeName = String(record.name);
    if (!router.hasRoute(routeName)) {
      router.addRoute('app-shell', record);
      routeNames.push(routeName);
    }
  }

  return routeNames;
}

export function buildMenusOnly(items: BackendMenuRoute[]): SidebarMenuNode[] {
  return buildSidebarNodes(items);
}

export function mapPluginMenusToBackendRoutes(items: PluginMenu[]): BackendMenuRoute[] {
  return items.map((item) => ({
    name: item.plugin ? `${item.plugin}:${item.id}` : item.id,
    path: item.path,
    component: item.component,
    redirect: item.redirect,
    hidden: !item.visible || !item.enabled,
    alwaysShow: item.type === 'directory',
    meta: {
      title: item.name,
      icon: item.icon,
      permission: item.permission,
      hidden: !item.visible || !item.enabled,
      noCache: false,
      affix: false,
      link: item.external_url,
    },
    children: mapPluginMenusToBackendRoutes(item.children ?? []),
  }));
}
