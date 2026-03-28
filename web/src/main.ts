import { createApp } from 'vue';
import ElementPlus from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import 'element-plus/dist/index.css';

import { restoreAuthenticatedSession } from '@/auth/bootstrap';
import { setUnauthorizedHandler } from '@/api/http';
import { permissionDirective } from '@/directives/permission';
import App from './App.vue';
import router from './router';
import pinia from './store';
import { useAppStore } from './store/app';
import { useMenuStore } from './store/menu';
import { useSessionStore } from './store/session';
import './styles/index.css';

const app = createApp(App);

app.use(pinia);
app.directive('permission', permissionDirective);

const sessionStore = useSessionStore(pinia);
const appStore = useAppStore(pinia);
const menuStore = useMenuStore(pinia);

sessionStore.hydrate();
appStore.hydrate();

setUnauthorizedHandler(() => {
  menuStore.clear(router);
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
  } catch {
    menuStore.clear(router);
    sessionStore.clearSession();
  }
}

app.use(router);
app.use(ElementPlus, { locale: zhCn });
await router.isReady();
app.mount('#app');
