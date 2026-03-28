import type { MenuItem } from '@/types/admin';

export interface TreeOption {
  label: string;
  value: string;
  children?: TreeOption[];
}

export interface FlatMenuOption {
  id: string;
  name: string;
  path: string;
}

export function formatDateTime(value: string | number | Date | null | undefined): string {
  if (value == null || value === '') {
    return '-';
  }
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function buildTreeOptions(items: MenuItem[]): TreeOption[] {
  return items.map((item) => ({
    label: `${item.name} (${item.path})`,
    value: item.id,
    children: item.children && item.children.length > 0 ? buildTreeOptions(item.children) : undefined,
  }));
}

export function flattenMenuItems(items: MenuItem[]): FlatMenuOption[] {
  const result: FlatMenuOption[] = [];

  const walk = (nodes: MenuItem[]) => {
    for (const node of nodes) {
      result.push({
        id: node.id,
        name: node.name,
        path: node.path,
      });
      if (node.children && node.children.length > 0) {
        walk(node.children);
      }
    }
  };

  walk(items);
  return result;
}

export function statusTagType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  switch (status.toLowerCase()) {
    case 'active':
    case 'enabled':
    case 'normal':
      return 'success';
    case 'inactive':
    case 'disabled':
      return 'danger';
    default:
      return 'info';
  }
}

export function menuTypeTagType(type: string): 'success' | 'warning' | 'info' {
  switch (type.toLowerCase()) {
    case 'directory':
      return 'info';
    case 'menu':
      return 'success';
    case 'button':
      return 'warning';
    default:
      return 'info';
  }
}
