import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { fetchMenuRoutes } from '@/api/menu';
import { fetchPluginMenus } from '@/api/plugins';
import { buildMenusOnly, filterMenuRoutesByPermission, mapPluginMenusToBackendRoutes, registerBackendRoutes } from '@/router/navigation';
import { useSessionStore } from '@/store/session';
export const useMenuStore = defineStore('menu', () => {
    const sessionStore = useSessionStore();
    const menuRoutes = ref([]);
    const sidebarMenus = ref([]);
    const routeNames = ref([]);
    const loaded = ref(false);
    const loading = ref(false);
    let loadPromise = null;
    const menuCount = computed(() => menuRoutes.value.length);
    const hasMenus = computed(() => sidebarMenus.value.length > 0);
    async function ensureLoaded(router) {
        if (loaded.value) {
            return;
        }
        if (loadPromise) {
            await loadPromise;
            return;
        }
        loading.value = true;
        loadPromise = (async () => {
            const [menuResult, pluginResult] = await Promise.allSettled([fetchMenuRoutes(), fetchPluginMenus()]);
            const items = menuResult.status === 'fulfilled' ? menuResult.value.items ?? [] : [];
            const pluginItems = pluginResult.status === 'fulfilled' ? mapPluginMenusToBackendRoutes(pluginResult.value.items ?? []) : [];
            const mergedItems = [...items, ...pluginItems];
            const canAccessMenu = (permission) => sessionStore.hasPermission(permission);
            menuRoutes.value = filterMenuRoutesByPermission(mergedItems, canAccessMenu);
            sidebarMenus.value = buildMenusOnly(mergedItems, canAccessMenu);
            routeNames.value = registerBackendRoutes(router, mergedItems, canAccessMenu);
            loaded.value = true;
        })();
        try {
            await loadPromise;
        }
        finally {
            loading.value = false;
            loadPromise = null;
        }
    }
    function clear(router) {
        if (router) {
            for (const name of routeNames.value) {
                if (router.hasRoute(name)) {
                    router.removeRoute(name);
                }
            }
        }
        menuRoutes.value = [];
        sidebarMenus.value = [];
        routeNames.value = [];
        loaded.value = false;
        loading.value = false;
        loadPromise = null;
    }
    return {
        menuRoutes,
        sidebarMenus,
        routeNames,
        loaded,
        loading,
        menuCount,
        hasMenus,
        ensureLoaded,
        clear,
    };
});
