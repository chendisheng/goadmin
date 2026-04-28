import { createRouter, createWebHistory } from 'vue-router';
import { appRoutes } from './routes';
import { preloadRouteNamespaces, resolveRouteLocaleMeta } from '@/i18n';
import { collectNamespacesFromRouteMeta } from '@/i18n/namespaces';
import { useMenuStore } from '@/store/menu';
import { useTabsStore } from '@/store/tabs';
import { useSessionStore } from '@/store/session';
const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: appRoutes,
    scrollBehavior() {
        return { top: 0 };
    },
});
router.beforeEach(async (to) => {
    const sessionStore = useSessionStore();
    const menuStore = useMenuStore();
    await preloadRouteNamespaces(to);
    const resolvedRoute = router.resolve(to.fullPath);
    const localized = resolveRouteLocaleMeta(resolvedRoute);
    const pageTitle = localized.title.trim() !== '' ? localized.title : (typeof to.meta.title === 'string' && to.meta.title.trim() !== '' ? to.meta.title : appTitle);
    document.title = `${pageTitle} | ${appTitle}`;
    const publicRoute = to.meta.public === true || to.meta.requiresAuth === false || to.path === '/login';
    if (!publicRoute && !sessionStore.isAuthenticated) {
        return {
            path: '/login',
            query: {
                redirect: to.fullPath,
            },
        };
    }
    if (!publicRoute && sessionStore.isAuthenticated && !menuStore.loaded) {
        try {
            await menuStore.ensureLoaded(router);
        }
        catch {
            menuStore.clear(router);
            useTabsStore().clearTabs();
            sessionStore.clearSession();
            return {
                path: '/login',
                query: {
                    redirect: to.fullPath,
                },
            };
        }
    }
    if (to.path === '/login' && sessionStore.isAuthenticated) {
        const redirect = typeof to.query.redirect === 'string' && to.query.redirect.trim() !== '' ? to.query.redirect : '/dashboard';
        return redirect;
    }
});
router.afterEach((to) => {
    const tabsStore = useTabsStore();
    tabsStore.resolveTabsFromRoute(to);
});
export default router;
