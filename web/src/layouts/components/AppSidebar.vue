<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Odometer } from '@element-plus/icons-vue';

import { useAppStore } from '@/store/app';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const router = useRouter();
const route = useRoute();
const appStore = useAppStore();

const menuItems = computed(() =>
  router
    .getRoutes()
    .filter((item) => item.meta.inMenu === true)
    .sort((left, right) => Number(left.meta.order || 0) - Number(right.meta.order || 0)),
);

const iconMap = {
  Odometer,
} as const;

function resolveIcon(iconName?: string) {
  return (iconName && iconName in iconMap ? iconMap[iconName as keyof typeof iconMap] : Odometer) || Odometer;
}
</script>

<template>
  <el-aside class="app-sidebar" :width="appStore.sidebarCollapsed ? '72px' : '244px'">
    <div class="app-sidebar__brand">
      <div class="app-sidebar__logo">G</div>
      <div v-if="!appStore.sidebarCollapsed" class="app-sidebar__brand-text">
        <strong>{{ appTitle }}</strong>
        <span>Frontend Core</span>
      </div>
    </div>

    <el-menu
      class="app-sidebar__menu"
      :collapse="appStore.sidebarCollapsed"
      :collapse-transition="false"
      :default-active="route.path"
      background-color="transparent"
      text-color="inherit"
      active-text-color="var(--el-color-primary)"
      router
    >
      <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
        <el-icon>
          <component :is="resolveIcon(String(item.meta.icon || 'Odometer'))" />
        </el-icon>
        <template #title>
          {{ item.meta.title }}
        </template>
      </el-menu-item>
    </el-menu>

    <div class="app-sidebar__footer">
      <el-button class="app-sidebar__toggle" text @click="appStore.toggleSidebar()">
        {{ appStore.sidebarCollapsed ? '展开侧栏' : '收起侧栏' }}
      </el-button>
    </div>
  </el-aside>
</template>
