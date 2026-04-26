import type { RouteRecordRaw } from 'vue-router';

import AppLayout from '@/layouts/AppLayout.vue';
import DashboardView from '@/views/DashboardView.vue';
import LoginView from '@/views/LoginView.vue';
import NotFoundView from '@/views/NotFoundView.vue';
import PluginCenterDetailView from '@/views/plugin/center/detail.vue';

export const appRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: {
      title: '登录',
      titleKey: 'route.login',
      titleDefault: '登录',
      public: true,
      hideInMenu: true,
      requiresAuth: false,
    },
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/',
    name: 'app-shell',
    component: AppLayout,
    meta: {
      title: 'GoAdmin',
      titleKey: 'app.title',
      titleDefault: 'GoAdmin',
      subtitle: 'Frontend Core',
      subtitleKey: 'app.subtitle',
      subtitleDefault: 'Frontend Core',
      requiresAuth: true,
      hideInMenu: true,
    },
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: DashboardView,
        meta: {
          title: '工作台',
          titleKey: 'route.dashboard',
          titleDefault: '工作台',
          requiresAuth: true,
          hideInMenu: true,
        },
      },
      {
        path: 'system/plugins/:name',
        name: 'plugin-center-detail',
        component: PluginCenterDetailView,
        meta: {
          title: '插件详情',
          titleKey: 'route.plugin_detail',
          titleDefault: '插件详情',
          requiresAuth: true,
          hideInMenu: true,
        },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: NotFoundView,
    meta: {
      title: '页面不存在',
      titleKey: 'route.not_found',
      titleDefault: '页面不存在',
      hideInMenu: true,
    },
  },
];
