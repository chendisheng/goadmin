import type { PluginMenu, PluginPermission } from '@/types/plugin';

function normalizeName(value: string): string {
  return value.trim();
}

function canUseStorage(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';
}

function readJsonValue<T>(key: string, fallback: T): T {
  if (!canUseStorage()) {
    return fallback;
  }
  const raw = window.localStorage.getItem(key);
  if (!raw) {
    return fallback;
  }
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeJsonValue<T>(key: string, value: T): void {
  if (!canUseStorage()) {
    return;
  }
  window.localStorage.setItem(key, JSON.stringify(value));
}

export interface PluginPermissionTemplate {
  key: string;
  label: string;
  description: string;
  actions: string[];
}

export interface PluginPermissionPreset {
  id: string;
  name: string;
  pluginName: string;
  templateKey: string;
  actions: string[];
  createdAt: string;
}

export interface PluginPermissionPresetGroup {
  pluginName: string;
  presets: PluginPermissionPreset[];
}

export interface PluginPermissionDiffRow {
  menuId: string;
  menuName: string;
  object: string;
  expectedActions: string[];
  existingActions: string[];
  missingActions: string[];
  extraActions: string[];
}

const PLUGIN_PERMISSION_PRESET_KEY = 'goadmin.plugin.permission.presets';

export const pluginPermissionTemplates: PluginPermissionTemplate[] = [
  {
    key: 'read_only',
    label: 'Read-only template',
    description: 'Batch-generate view permissions for menus',
    actions: ['view'],
  },
  {
    key: 'crud',
    label: 'CRUD template',
    description: 'Batch-generate view, create, edit, and delete permissions for menus',
    actions: ['view', 'create', 'update', 'delete'],
  },
  {
    key: 'manage',
    label: 'Management template',
    description: 'Batch-generate view, edit, and delete permissions for menus',
    actions: ['view', 'update', 'delete'],
  },
  {
    key: 'button_ops',
    label: 'Button template',
    description: 'Batch-generate create, edit, and delete permissions for button menus',
    actions: ['create', 'update', 'delete'],
  },
];

export function readPluginPermissionPresets(): PluginPermissionPreset[] {
  return readJsonValue<PluginPermissionPreset[]>(PLUGIN_PERMISSION_PRESET_KEY, []).map((item) => ({
    ...item,
    pluginName: normalizeName((item as PluginPermissionPreset & { pluginName?: string }).pluginName || ''),
  }));
}

export function savePluginPermissionPreset(pluginName: string, name: string, templateKey: string, actions: string[]): PluginPermissionPreset[] {
  const trimmedName = normalizeName(name);
  if (trimmedName === '') {
    return readPluginPermissionPresets();
  }
  const normalizedPluginName = normalizeName(pluginName);
  const normalizedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
  const current = readPluginPermissionPresets();
  const preset: PluginPermissionPreset = {
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

export function removePluginPermissionPreset(id: string): PluginPermissionPreset[] {
  const normalizedId = normalizeName(id);
  const next = readPluginPermissionPresets().filter((item) => item.id !== normalizedId);
  writeJsonValue(PLUGIN_PERMISSION_PRESET_KEY, next);
  return next;
}

export function groupPluginPermissionPresets(presets: PluginPermissionPreset[]): PluginPermissionPresetGroup[] {
  const groups = new Map<string, PluginPermissionPreset[]>();
  for (const preset of presets) {
    const key = normalizeName(preset.pluginName) || 'Ungrouped';
    const current = groups.get(key) ?? [];
    current.push(preset);
    groups.set(key, current);
  }

  return Array.from(groups.entries()).map(([pluginName, groupedPresets]) => ({
    pluginName,
    presets: groupedPresets.sort((left, right) => right.createdAt.localeCompare(left.createdAt)),
  }));
}

export interface PluginMenuLocation {
  list: PluginMenu[];
  index: number;
  node: PluginMenu;
  parentId: string;
}

export function createPluginMenuNode(pluginName = '', parentId = ''): PluginMenu {
  const normalizedPlugin = normalizeName(pluginName);
  return {
    plugin: normalizedPlugin,
    id: `menu-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    parent_id: parentId,
    name: '',
    titleKey: '',
    titleDefault: '',
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

export function createPluginPermissionsForMenu(pluginName: string, menu: PluginMenu, actions: string[]): PluginPermission[] {
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

export function clonePluginMenuTree(items: PluginMenu[]): PluginMenu[] {
  return items.map((item) => ({
    ...item,
    children: clonePluginMenuTree(item.children ?? []),
  }));
}

export function normalizePluginMenuTree(items: PluginMenu[], pluginName = '', parentId = ''): PluginMenu[] {
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

export function findPluginMenuLocation(items: PluginMenu[], nodeId: string, parentId = ''): PluginMenuLocation | null {
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

function containsPluginMenuNode(node: PluginMenu, targetId: string): boolean {
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

export function movePluginMenuNode(items: PluginMenu[], sourceId: string, targetId: string, position: 'before' | 'after' | 'inside'): boolean {
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
  } else {
    const insertIndex = position === 'before' ? refreshedTargetLocation.index : refreshedTargetLocation.index + 1;
    refreshedTargetLocation.list.splice(insertIndex, 0, movingNode);
  }

  normalizePluginMenuTree(items, movingNode.plugin || '');
  return true;
}

export function createPluginPermissionNode(pluginName = ''): PluginPermission {
  return {
    plugin: normalizeName(pluginName),
    object: '',
    action: '',
    description: '',
  };
}

export function flattenPluginMenus(items: PluginMenu[]): PluginMenu[] {
  const result: PluginMenu[] = [];
  const walk = (nodes: PluginMenu[]) => {
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

export function buildPluginPermissionDiffRows(pluginName: string, menus: PluginMenu[], permissions: PluginPermission[], actions: string[]): PluginPermissionDiffRow[] {
  const normalizedPlugin = normalizeName(pluginName);
  const expectedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
  if (normalizedPlugin === '' || expectedActions.length === 0) {
    return [];
  }

  return flattenPluginMenus(menus).map((menu) => {
    const object = `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`;
    const existingActions = [...new Set(
      permissions
        .filter((permission) => normalizeName(permission.object) === object)
        .map((permission) => normalizeName(permission.action))
        .filter((action) => action !== ''),
    )];
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

export function buildPluginPermissionOrphans(pluginName: string, menus: PluginMenu[], permissions: PluginPermission[]): PluginPermission[] {
  const normalizedPlugin = normalizeName(pluginName);
  if (normalizedPlugin === '') {
    return permissions;
  }
  const menuObjects = new Set(
    flattenPluginMenus(menus).map((menu) => `plugin:${normalizedPlugin}:${normalizeName(menu.id) || normalizeName(menu.path)}`),
  );
  return permissions.filter((permission) => !menuObjects.has(normalizeName(permission.object)));
}

export function generatePluginPermissions(pluginName: string, menus: PluginMenu[], actions: string[]): PluginPermission[] {
  const normalizedPlugin = normalizeName(pluginName);
  const normalizedActions = [...new Set(actions.map(normalizeName).filter((item) => item !== ''))];
  if (normalizedPlugin === '' || normalizedActions.length === 0) {
    return [];
  }

  const result: PluginPermission[] = [];
  const seen = new Set<string>();

  const walk = (nodes: PluginMenu[]) => {
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

export function generatePluginPermissionsFromTemplate(pluginName: string, menus: PluginMenu[], templateKey: string): PluginPermission[] {
  const template = pluginPermissionTemplates.find((item) => normalizeName(item.key) === normalizeName(templateKey));
  if (!template) {
    return [];
  }
  return generatePluginPermissions(pluginName, menus, template.actions);
}

export function mergePluginPermissions(existing: PluginPermission[], generated: PluginPermission[]): PluginPermission[] {
  const map = new Map<string, PluginPermission>();
  for (const item of existing) {
    map.set(`${normalizeName(item.object)}:${normalizeName(item.action)}`, item);
  }
  for (const item of generated) {
    map.set(`${normalizeName(item.object)}:${normalizeName(item.action)}`, item);
  }
  return Array.from(map.values());
}
