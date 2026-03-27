import type { RouteRecordRaw } from 'vue-router';

import AppLayout from '@/layouts/AppLayout.vue';
import LoginView from '@/views/LoginView.vue';
import DashboardView from '@/views/DashboardView.vue';
import NotFoundView from '@/views/NotFoundView.vue';

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
    component: AppLayout,
    meta: {
      title: 'GoAdmin',
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
          subtitle: 'Phase 10 Frontend Core',
          icon: 'Odometer',
          inMenu: true,
          requiresAuth: true,
          order: 1,
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
