<script setup lang="ts">
import { computed } from 'vue';

import AppHeader from './components/AppHeader.vue';
import AppSidebar from './components/AppSidebar.vue';
import TabsBar from './components/TabsBar.vue';
import { useLocaleStore } from '@/store/locale';
import { useTabsStore } from '@/store/tabs';

const localeStore = useLocaleStore();
const tabsStore = useTabsStore();
const cachedViewNames = computed(() => tabsStore.cachedViewNames);
</script>

<template>
  <el-container class="app-layout">
    <AppSidebar />
    <el-container class="app-layout__content" direction="vertical">
      <AppHeader />
      <TabsBar />
      <el-main class="app-layout__main">
        <router-view v-slot="{ Component, route }">
          <KeepAlive :include="cachedViewNames">
            <component :is="Component" :key="`${route.fullPath}:${localeStore.language}`" />
          </KeepAlive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>
