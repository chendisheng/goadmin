import AppLayout from '@/layouts/AppLayout.vue';
import DashboardView from '@/views/DashboardView.vue';
import LoginView from '@/views/LoginView.vue';
import NotFoundView from '@/views/NotFoundView.vue';
import PluginCenterDetailView from '@/views/plugin/center/detail.vue';
export const appRoutes = [
    {
        path: '/login',
        name: 'login',
        component: LoginView,
        meta: {
            title: 'Login',
            titleKey: 'route.login',
            titleDefault: 'Login',
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
                    title: 'Dashboard',
                    titleKey: 'route.dashboard',
                    titleDefault: 'Dashboard',
                    requiresAuth: true,
                    hideInMenu: true,
                },
            },
            {
                path: 'system/plugins/:name',
                name: 'plugin-center-detail',
                component: PluginCenterDetailView,
                meta: {
                    title: 'Plugin details',
                    titleKey: 'route.plugin_detail',
                    titleDefault: 'Plugin details',
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
            title: 'Page not found',
            titleKey: 'route.not_found',
            titleDefault: 'Page not found',
            hideInMenu: true,
        },
    },
];
