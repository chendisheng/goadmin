import type { MenuItem } from '@/types/admin';
import { getAppLanguage } from '@/i18n';

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
  const locale = getAppLanguage();
  return date.toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export function formatRemainingTime(value: string, now: number, expired: boolean): string {
  if (!value) {
    return '-';
  }
  const locale = getAppLanguage();
  const isZh = locale.toLowerCase().startsWith('zh');
  if (expired) {
    return isZh ? '已过期' : 'Expired';
  }
  const expiresAt = new Date(value).getTime();
  if (Number.isNaN(expiresAt)) {
    return '-';
  }
  const diff = Math.max(0, expiresAt - now);
  if (diff <= 0) {
    return isZh ? '已过期' : 'Expired';
  }
  const totalSeconds = Math.floor(diff / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  const parts: string[] = [];
  if (days > 0) {
    parts.push(isZh ? `${days}天` : `${days}d`);
  }
  if (hours > 0 || parts.length > 0) {
    parts.push(isZh ? `${hours}小时` : `${hours}h`);
  }
  if (minutes > 0 || parts.length > 0) {
    parts.push(isZh ? `${minutes}分钟` : `${minutes}m`);
  }
  parts.push(isZh ? `${seconds}秒` : `${seconds}s`);
  return isZh ? parts.join('') : parts.join(' ');
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
