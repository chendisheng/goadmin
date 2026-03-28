import type { Component } from 'vue';
import {
  House,
  Menu,
  MoreFilled,
  Odometer,
  Box,
  Setting,
  User,
  UserFilled,
} from '@element-plus/icons-vue';

const iconMap: Record<string, Component> = {
  home: House,
  dashboard: Odometer,
  user: UserFilled,
  setting: Setting,
  role: User,
  menu: Menu,
  box: Box,
  circle: MoreFilled,
  dot: MoreFilled,
};

export function resolveMenuIcon(iconName?: string): Component {
  const key = (iconName || '').trim().toLowerCase();
  return iconMap[key] || Menu;
}
