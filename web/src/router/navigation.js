import { defineAsyncComponent, defineComponent, h } from 'vue';
import RouteGroupView from '@/views/RouteGroupView.vue';
import RoutePlaceholderView from '@/views/RoutePlaceholderView.vue';
function isButtonMenu(node) {
    return String(node.type || '').trim() === 'button';
}
const viewModules = import.meta.glob('../views/**/*.vue');
function normalizeComponentKey(componentName) {
    const normalized = componentName.trim();
    const safeName = normalized
        .replace(/[^A-Za-z0-9_-]+/g, '-')
        .replace(/-+/g, '-')
        .replace(/^[-_]+|[-_]+$/g, '');
    return safeName === '' ? 'route-view' : safeName;
}
function createNamedView(name, component) {
    return defineComponent({
        name,
        setup() {
            return () => h(component);
        },
    });
}
export function normalizeMenuRoots(nodes) {
    return nodes;
}
function resolveTitleKey(meta) {
    return (meta.titleKey || meta.titleDefault || meta.title || '').trim();
}
function componentNameToModulePath(componentName) {
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
function normalizePath(value) {
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
function relativePath(value, parentPath) {
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
function canAccessMenu(permission, canAccess) {
    const normalized = (permission || '').trim();
    if (normalized === '') {
        return true;
    }
    return canAccess(normalized);
}
export function filterMenuRoutesByPermission(nodes, canAccess) {
    const result = [];
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
function buildSidebarNodes(nodes) {
    const roots = normalizeMenuRoots(nodes);
    const result = [];
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
function resolveLeafComponent(componentName) {
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
function buildRouteRecord(node, parentPath = '/') {
    const currentPath = normalizePath(node.path);
    const childPath = relativePath(currentPath, parentPath);
    const children = (node.children ?? []).map((child) => buildRouteRecord(child, currentPath));
    const hasChildren = children.length > 0;
    const componentName = (node.component || '').trim();
    const record = {
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
export function registerBackendRoutes(router, nodes, canAccess = () => true) {
    const roots = filterMenuRoutesByPermission(nodes, canAccess);
    const routeNames = [];
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
export function buildMenusOnly(items, canAccess = () => true) {
    return buildSidebarNodes(filterMenuRoutesByPermission(items, canAccess));
}
export function mapPluginMenusToBackendRoutes(items) {
    return items.map((item) => ({
        name: item.plugin ? `${item.plugin}:${item.id}` : item.id,
        path: item.path,
        component: item.component,
        redirect: item.redirect,
        hidden: !item.visible || !item.enabled,
        type: item.type,
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
        },
        children: mapPluginMenusToBackendRoutes(item.children ?? []),
    }));
}
