function normalizeName(value) {
    return value.trim();
}
function canUseStorage() {
    return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}
function readJsonValue(key, fallback) {
    if (!canUseStorage()) {
        return fallback;
    }
    const raw = window.localStorage.getItem(key);
    if (!raw) {
        return fallback;
    }
    try {
        return JSON.parse(raw);
    }
    catch {
        return fallback;
    }
}
function writeJsonValue(key, value) {
    if (!canUseStorage()) {
        return;
    }
    window.localStorage.setItem(key, JSON.stringify(value));
}
const PLUGIN_PERMISSION_PRESET_KEY = 'goadmin.plugin.permission.presets';
export const pluginPermissionTemplates = [
    {
        key: 'read_only',
        label: '只读模板',
        description: '为菜单批量生成查看权限',
        actions: ['view'],
    },
    {
        key: 'crud',
        label: 'CRUD 模板',
        description: '为菜单批量生成查看、创建、编辑和删除权限',
        actions: ['view', 'create', 'update', 'delete'],
    },
    {
        key: 'manage',
        label: '管理模板',
        description: '为菜单批量生成查看、编辑和删除权限',
        actions: ['view', 'update', 'delete'],
    },
    {
        key: 'button_ops',
        label: '按钮模板',
        description: '为按钮类菜单批量生成创建、编辑和删除权限',
        actions: ['create', 'update', 'delete'],
    },
];
export function readPluginPermissionPresets() {
    return readJsonValue(PLUGIN_PERMISSION_PRESET_KEY, []).map((item) => ({
        ...item,
        pluginName: normalizeName(item.pluginName || ''),
    }));
}
export function savePluginPermissionPreset(pluginName, name, templateKey, actions) {
    const trimmedName = normalizeName(name);
    if (trimmedName === '') {
        return readPluginPermissionPresets();
    }
    const normalizedPluginName = normalizeName(pluginName);
    const normalizedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
    const current = readPluginPermissionPresets();
    const preset = {
        id: `preset-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        name: trimmedName,
        pluginName: normalizedPluginName,
        templateKey: normalizeName(templateKey),
        actions: normalizedActions,
        createdAt: new Date().toISOString(),
    };
    const next = [preset, ...current.filter((item) => item.name !== trimmedName || item.pluginName !== normalizedPluginName)];
    writeJsonValue(PLUGIN_PERMISSION_PRESET_KEY, next);
    return next;
}
export function removePluginPermissionPreset(id) {
    const normalizedId = normalizeName(id);
    const next = readPluginPermissionPresets().filter((item) => item.id !== normalizedId);
    writeJsonValue(PLUGIN_PERMISSION_PRESET_KEY, next);
    return next;
}
export function groupPluginPermissionPresets(presets) {
    const groups = new Map();
    for (const preset of presets) {
        const key = normalizeName(preset.pluginName) || '未分组';
        const current = groups.get(key) ?? [];
        current.push(preset);
        groups.set(key, current);
    }
    return Array.from(groups.entries()).map(([pluginName, groupedPresets]) => ({
        pluginName,
        presets: groupedPresets.sort((left, right) => right.createdAt.localeCompare(left.createdAt)),
    }));
}
export function createPluginMenuNode(pluginName = '', parentId = '') {
    const normalizedPlugin = normalizeName(pluginName);
    return {
        plugin: normalizedPlugin,
        id: `menu-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
        parent_id: parentId,
        name: '',
        path: '',
        component: '',
        icon: '',
        sort: 0,
        permission: '',
        type: 'menu',
        visible: true,
        enabled: true,
        redirect: '',
        external_url: '',
        children: [],
    };
}
export function createPluginPermissionsForMenu(pluginName, menu, actions) {
    const normalizedPlugin = normalizeName(pluginName);
    const normalizedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
    if (normalizedPlugin === '' || normalizedActions.length === 0) {
        return [];
    }
    const objectKey = `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`;
    return normalizedActions.map((action) => ({
        plugin: normalizedPlugin,
        object: objectKey,
        action,
        description: `${menu.name || menu.id} ${action}`,
    }));
}
export function clonePluginMenuTree(items) {
    return items.map((item) => ({
        ...item,
        children: clonePluginMenuTree(item.children ?? []),
    }));
}
export function normalizePluginMenuTree(items, pluginName = '', parentId = '') {
    const normalizedPlugin = normalizeName(pluginName);
    items.forEach((item, index) => {
        const currentPlugin = normalizeName(item.plugin || normalizedPlugin);
        item.plugin = currentPlugin;
        item.parent_id = normalizeName(parentId);
        item.sort = index + 1;
        item.type = normalizeName(item.type) || 'menu';
        item.children = item.children ?? [];
        normalizePluginMenuTree(item.children, currentPlugin, item.id);
    });
    return items;
}
export function findPluginMenuLocation(items, nodeId, parentId = '') {
    const normalizedId = normalizeName(nodeId);
    for (let index = 0; index < items.length; index += 1) {
        const node = items[index];
        if (normalizeName(node.id) === normalizedId) {
            return {
                list: items,
                index,
                node,
                parentId,
            };
        }
        const children = node.children ?? [];
        const found = findPluginMenuLocation(children, normalizedId, node.id);
        if (found) {
            return found;
        }
    }
    return null;
}
function containsPluginMenuNode(node, targetId) {
    if (normalizeName(node.id) === normalizeName(targetId)) {
        return true;
    }
    for (const child of node.children ?? []) {
        if (containsPluginMenuNode(child, targetId)) {
            return true;
        }
    }
    return false;
}
export function movePluginMenuNode(items, sourceId, targetId, position) {
    const sourceLocation = findPluginMenuLocation(items, sourceId);
    const targetLocation = findPluginMenuLocation(items, targetId);
    if (!sourceLocation || !targetLocation) {
        return false;
    }
    if (normalizeName(sourceLocation.node.id) === normalizeName(targetLocation.node.id)) {
        return false;
    }
    if (containsPluginMenuNode(sourceLocation.node, targetId)) {
        return false;
    }
    const [movingNode] = sourceLocation.list.splice(sourceLocation.index, 1);
    const refreshedTargetLocation = findPluginMenuLocation(items, targetId);
    if (!refreshedTargetLocation) {
        sourceLocation.list.splice(sourceLocation.index, 0, movingNode);
        return false;
    }
    if (position === 'inside') {
        refreshedTargetLocation.node.children = refreshedTargetLocation.node.children ?? [];
        refreshedTargetLocation.node.children.push(movingNode);
    }
    else {
        const insertIndex = position === 'before' ? refreshedTargetLocation.index : refreshedTargetLocation.index + 1;
        refreshedTargetLocation.list.splice(insertIndex, 0, movingNode);
    }
    normalizePluginMenuTree(items, movingNode.plugin || '');
    return true;
}
export function createPluginPermissionNode(pluginName = '') {
    return {
        plugin: normalizeName(pluginName),
        object: '',
        action: '',
        description: '',
    };
}
export function flattenPluginMenus(items) {
    const result = [];
    const walk = (nodes) => {
        for (const node of nodes) {
            result.push(node);
            if (node.children && node.children.length > 0) {
                walk(node.children);
            }
        }
    };
    walk(items);
    return result;
}
export function buildPluginPermissionDiffRows(pluginName, menus, permissions, actions) {
    const normalizedPlugin = normalizeName(pluginName);
    const expectedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
    if (normalizedPlugin === '' || expectedActions.length === 0) {
        return [];
    }
    return flattenPluginMenus(menus).map((menu) => {
        const object = `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`;
        const existingActions = [...new Set(permissions
                .filter((permission) => normalizeName(permission.object) === object)
                .map((permission) => normalizeName(permission.action))
                .filter((action) => action !== ''))];
        const missingActions = expectedActions.filter((action) => !existingActions.includes(action));
        const extraActions = existingActions.filter((action) => !expectedActions.includes(action));
        return {
            menuId: menu.id,
            menuName: menu.name || menu.id,
            object,
            expectedActions,
            existingActions,
            missingActions,
            extraActions,
        };
    });
}
export function buildPluginPermissionOrphans(pluginName, menus, permissions) {
    const normalizedPlugin = normalizeName(pluginName);
    if (normalizedPlugin === '') {
        return permissions;
    }
    const menuObjects = new Set(flattenPluginMenus(menus).map((menu) => `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`));
    return permissions.filter((permission) => !menuObjects.has(normalizeName(permission.object)));
}
export function generatePluginPermissions(pluginName, menus, actions) {
    const normalizedPlugin = normalizeName(pluginName);
    const normalizedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
    if (normalizedPlugin === '' || normalizedActions.length === 0) {
        return [];
    }
    const result = [];
    const seen = new Set();
    const walk = (nodes) => {
        for (const menu of nodes) {
            const objectKey = `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`;
            for (const action of normalizedActions) {
                const key = `${objectKey}:${action}`;
                if (seen.has(key)) {
                    continue;
                }
                seen.add(key);
                result.push({
                    plugin: normalizedPlugin,
                    object: objectKey,
                    action,
                    description: `${menu.name || menu.id} ${action}`,
                });
            }
            if (menu.children && menu.children.length > 0) {
                walk(menu.children);
            }
        }
    };
    walk(menus);
    return result;
}
export function generatePluginPermissionsFromTemplate(pluginName, menus, templateKey) {
    const template = pluginPermissionTemplates.find((item) => normalizeName(item.key) === normalizeName(templateKey));
    if (!template) {
        return [];
    }
    return generatePluginPermissions(pluginName, menus, template.actions);
}
export function mergePluginPermissions(existing, generated) {
    const map = new Map();
    for (const item of existing) {
        map.set(`${normalizeName(item.object)}:${normalizeName(item.action)}`, item);
    }
    for (const item of generated) {
        map.set(`${normalizeName(item.object)}:${normalizeName(item.action)}`, item);
    }
    return Array.from(map.values());
}
