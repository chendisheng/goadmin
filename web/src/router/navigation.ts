import { defineAsyncComponent, defineComponent, h } from 'vue';
import type { AsyncComponentLoader, Component } from 'vue';
import type { RouteRecordRaw, Router } from 'vue-router';

import RouteGroupView from '@/views/RouteGroupView.vue';
import RoutePlaceholderView from '@/views/RoutePlaceholderView.vue';
import type { BackendMenuRoute, SidebarMenuNode } from '@/types/menu';
import type { PluginMenu } from '@/types/plugin';

type ViewModuleLoader = AsyncComponentLoader<Component>;
type PermissionChecker = (permission: string) => boolean;

function isButtonMenu(node: Pick<BackendMenuRoute, 'type' | 'meta'> | Pick<PluginMenu, 'type' | 'visible' | 'enabled'>): boolean {
  return String(node.type || '').trim() === 'button';
}

const viewModules = import.meta.glob('../views/**/*.vue') as Record<string, ViewModuleLoader>;

function normalizeComponentKey(componentName: string): string {
  const normalized = componentName.trim();
  const safeName = normalized
    .replace(/[^A-Za-z0-9_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^[-_]+|[-_]+$/g, '');
  return safeName === '' ? 'route-view' : safeName;
}

function createNamedView(name: string, component: Component): Component {
  return defineComponent({
    name,
    setup() {
      return () => h(component);
    },
  });
}

export function normalizeMenuRoots(nodes: BackendMenuRoute[]): BackendMenuRoute[] {
  return nodes;
}

function resolveTitleKey(meta: Pick<BackendMenuRoute['meta'], 'title' | 'titleKey' | 'titleDefault'>): string {
  return (meta.titleKey || meta.titleDefault || meta.title || '').trim();
}

function namespacesFromComponentName(componentName: string): string[] {
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

function canAccessMenu(permission: string | undefined, canAccess: PermissionChecker): boolean {
  const normalized = (permission || '').trim();
  if (normalized === '') {
    return true;
  }
  return canAccess(normalized);
}

export function filterMenuRoutesByPermission(nodes: BackendMenuRoute[], canAccess: PermissionChecker): BackendMenuRoute[] {
  const result: BackendMenuRoute[] =[];

  for (const node of normalizeMenuRoots(nodes)) {
    if (isButtonMenu(node)) {
      continue;
    }
    const children = filterMenuRoutesByPermission(node.children ?? [], canAccess);
    const allowed = canAccessMenu(node.meta.permission, canAccess);
    const hidden = node.hidden || !allowed;

    if (!allowed && children.length === 0) {
      continue;
    }

    result.push({
      ...node,
      hidden,
      children,
    });
  }

  return result;
}

function buildSidebarNodes(nodes: BackendMenuRoute[]): SidebarMenuNode[] {
  const roots = normalizeMenuRoots(nodes);
  const result: SidebarMenuNode[] = [];
  for (const node of roots) {
    if (isButtonMenu(node)) {
      continue;
    }
    const children = buildSidebarNodes(node.children ?? []);
    if (node.hidden) {
      result.push(...children);
      continue;
    }
    result.push({
      name: node.name,
      path: normalizePath(node.path),
      title: node.meta.title,
      titleKey: node.meta.titleKey,
      titleDefault: node.meta.titleDefault,
      icon: node.meta.icon,
      component: node.component,
      redirect: node.redirect,
      permission: node.meta.permission,
      hidden: node.hidden,
      subtitle: node.meta.subtitle,
      subtitleKey: node.meta.subtitleKey,
      subtitleDefault: node.meta.subtitleDefault,
      children,
    });
  }
  return result;
}

function resolveLeafComponent(componentName: string | undefined): Component {
  const normalized = (componentName || '').trim();
  if (normalized === '') {
    return createNamedView('route-view', RoutePlaceholderView);
  }
  if (normalized === 'Layout') {
    return RouteGroupView;
  }
  const modulePath = componentNameToModulePath(normalized);
  if (modulePath && viewModules[modulePath]) {
    const loadedComponent = defineAsyncComponent(viewModules[modulePath]);
    return createNamedView(normalizeComponentKey(normalized), loadedComponent);
  }
  return createNamedView(normalizeComponentKey(normalized), RoutePlaceholderView);
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
      title: resolveTitleKey(node.meta),
      titleKey: node.meta.titleKey,
      titleDefault: node.meta.titleDefault,
      icon: node.meta.icon,
      permission: node.meta.permission,
      link: node.meta.link,
      componentName: componentName || undefined,
      i18nNamespaces: namespacesFromComponentName(componentName),
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

export function registerBackendRoutes(router: Router, nodes: BackendMenuRoute[], canAccess: PermissionChecker = () => true): string[] {
  const roots = filterMenuRoutesByPermission(nodes, canAccess);
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

export function buildMenusOnly(items: BackendMenuRoute[], canAccess: PermissionChecker = () => true): SidebarMenuNode[] {
  return buildSidebarNodes(filterMenuRoutesByPermission(items, canAccess));
}

export function mapPluginMenusToBackendRoutes(items: PluginMenu[]): BackendMenuRoute[] {
  return items.map((item) => ({
    name: item.plugin ? `${item.plugin}:${item.id}` : item.id,
    path: item.path,
    component: item.component,
    redirect: item.redirect,
    hidden: !item.visible || !item.enabled,
    type: item.type as BackendMenuRoute['type'],
    alwaysShow: item.type === 'directory',
    meta: {
      title: item.titleDefault || item.name,
      titleKey: item.titleKey,
      titleDefault: item.titleDefault,
      icon: item.icon,
      permission: item.permission,
      hidden: !item.visible || !item.enabled,
      noCache: false,
      affix: false,
      link: item.external_url,
      subtitle: item.subtitle,
      subtitleKey: item.subtitleKey,
      subtitleDefault: item.subtitleDefault,
      componentName: item.component,
      i18nNamespaces: namespacesFromComponentName(item.component || ''),
    },
    children: mapPluginMenusToBackendRoutes(item.children ?? []),
  }));
}
