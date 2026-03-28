<script setup lang="ts">
import { useRoute } from 'vue-router';

import { useAppStore } from '@/store/app';
import { useMenuStore } from '@/store/menu';
import MenuTreeNode from './MenuTreeNode.vue';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const route = useRoute();
const appStore = useAppStore();
const menuStore = useMenuStore();
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

    <el-scrollbar class="app-sidebar__scroll">
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
        <MenuTreeNode v-for="item in menuStore.sidebarMenus" :key="item.path" :node="item" />
      </el-menu>
    </el-scrollbar>

    <div class="app-sidebar__footer">
      <el-button class="app-sidebar__toggle" text @click="appStore.toggleSidebar()">
        {{ appStore.sidebarCollapsed ? '展开侧栏' : '收起侧栏' }}
      </el-button>
    </div>
  </el-aside>
</template>
