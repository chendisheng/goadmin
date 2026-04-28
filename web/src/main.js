import { createApp, watch } from 'vue';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import { restoreAuthenticatedSession } from '@/auth/bootstrap';
import { setUnauthorizedHandler } from '@/api/http';
import { permissionDirective } from '@/directives/permission';
import { initializeI18n, setI18nLanguage } from '@/i18n';
import App from './App.vue';
import router from './router';
import pinia from './store';
import { useAppStore } from './store/app';
import { useLocaleStore } from './store/locale';
import { useMenuStore } from './store/menu';
import { useSessionStore } from './store/session';
import { useTabsStore } from './store/tabs';
import './styles/index.css';
const app = createApp(App);
app.use(pinia);
app.directive('permission', permissionDirective);
const sessionStore = useSessionStore(pinia);
const appStore = useAppStore(pinia);
const localeStore = useLocaleStore(pinia);
const menuStore = useMenuStore(pinia);
const tabsStore = useTabsStore(pinia);
sessionStore.hydrate();
appStore.hydrate();
localeStore.hydrate();
tabsStore.hydrate();
await initializeI18n(localeStore.language);
watch(() => localeStore.language, (language) => {
    void setI18nLanguage(language);
});
setUnauthorizedHandler(() => {
    menuStore.clear(router);
    tabsStore.clearTabs();
    sessionStore.clearSession();
    const currentPath = router.currentRoute.value.fullPath;
    if (router.currentRoute.value.path !== '/login') {
        void router.replace({
            path: '/login',
            query: {
                redirect: currentPath,
            },
        });
    }
});
if (sessionStore.isAuthenticated) {
    try {
        await restoreAuthenticatedSession();
        await menuStore.ensureLoaded(router);
        tabsStore.reconcilePersistedTabs(router);
    }
    catch {
        menuStore.clear(router);
        tabsStore.clearTabs();
        sessionStore.clearSession();
    }
}
app.use(router);
app.use(ElementPlus);
await router.isReady();
app.mount('#app');
