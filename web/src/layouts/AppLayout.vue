<script setup lang="ts">
import { computed } from 'vue';

import AppHeader from './components/AppHeader.vue';
import AppSidebar from './components/AppSidebar.vue';
import TabsBar from './components/TabsBar.vue';
import { useTabsStore } from '@/store/tabs';

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
            <component :is="Component" :key="route.fullPath" />
          </KeepAlive>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>
