import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { Router } from 'vue-router';

import { fetchMenuRoutes } from '@/api/menu';
import { fetchPluginMenus } from '@/api/plugins';
import { buildMenusOnly, mapPluginMenusToBackendRoutes, registerBackendRoutes } from '@/router/navigation';
import type { BackendMenuRoute, SidebarMenuNode } from '@/types/menu';

export const useMenuStore = defineStore('menu', () => {
  const menuRoutes = ref<BackendMenuRoute[]>([]);
  const sidebarMenus = ref<SidebarMenuNode[]>([]);
  const routeNames = ref<string[]>([]);
  const loaded = ref(false);
  const loading = ref(false);
  let loadPromise: Promise<void> | null = null;

  const menuCount = computed(() => menuRoutes.value.length);
  const hasMenus = computed(() => sidebarMenus.value.length > 0);

  async function ensureLoaded(router: Router): Promise<void> {
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
      menuRoutes.value = mergedItems;
      sidebarMenus.value = buildMenusOnly(mergedItems);
      routeNames.value = registerBackendRoutes(router, mergedItems);
      loaded.value = true;
    })();

    try {
      await loadPromise;
    } finally {
      loading.value = false;
      loadPromise = null;
    }
  }

  function clear(router?: Router): void {
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
