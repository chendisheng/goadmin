import type { RouteRecordRaw } from 'vue-router';

import AppLayout from '@/layouts/AppLayout.vue';
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
      requiresAuth: true,
      hideInMenu: true,
    },
    children: [
      {
        path: 'system/plugins/:name',
        name: 'plugin-center-detail',
        component: PluginCenterDetailView,
        meta: {
          title: '插件详情',
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
      hideInMenu: true,
    },
  },
];
