<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { ArrowDown, Expand, Fold, RefreshRight, UserFilled } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

import { logout as logoutApi } from '@/api/auth';
import { useAppStore } from '@/store/app';
import { useMenuStore } from '@/store/menu';
import { useSessionStore } from '@/store/session';
import { useTabsStore } from '@/store/tabs';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const buildMode = import.meta.env.MODE;
const appStore = useAppStore();
const menuStore = useMenuStore();
const sessionStore = useSessionStore();
const tabsStore = useTabsStore();
const router = useRouter();
const route = useRoute();

const pageTitle = computed(() => {
  if (typeof route.meta.title === 'string' && route.meta.title.trim() !== '') {
    return route.meta.title;
  }
  return appTitle;
});

const pageSubtitle = computed(() => {
  if (typeof route.meta.subtitle === 'string' && route.meta.subtitle.trim() !== '') {
    return route.meta.subtitle;
  }
  return 'Vue 3 + TypeScript + Vite + Pinia + Axios';
});

const currentUserName = computed(() => sessionStore.displayName || 'System Admin');

const currentUserRole = computed(() => {
  const role = sessionStore.currentUser?.roles?.[0];
  return typeof role === 'string' && role.trim() !== '' ? role : '管理员';
});

const currentUserInitial = computed(() => {
  const source = currentUserName.value.trim();
  if (source.length === 0) {
    return 'G';
  }
  return source.slice(0, 1).toUpperCase();
});

function refreshPage() {
  window.location.reload();
}

async function onLogout() {
  try {
    await logoutApi();
  } catch {
    // 退出时即使后端已失效也继续清理本地会话
  } finally {
    menuStore.clear(router);
    tabsStore.clearTabs();
    sessionStore.clearSession();
    ElMessage.success('已退出登录');
    await router.push({ path: '/login' });
  }
}

function onCommand(command: string) {
  if (command === 'refresh') {
    refreshPage();
  }
  if (command === 'logout') {
    void onLogout();
  }
}
</script>

<template>
  <el-header class="app-header">
    <div class="app-header__left">
      <el-button class="app-header__toggle" circle text @click="appStore.toggleSidebar()">
        <el-icon>
          <Fold v-if="!appStore.sidebarCollapsed" />
          <Expand v-else />
        </el-icon>
      </el-button>

      <div class="app-header__titles">
        <h1>{{ pageTitle }}</h1>
        <p>{{ pageSubtitle }}</p>
      </div>
    </div>

    <div class="app-header__right">
      <el-tag effect="plain" round type="info">{{ buildMode }}</el-tag>
      <el-tag effect="plain" round type="success">{{ apiBaseUrl }}</el-tag>
      <el-button circle text @click="refreshPage">
        <el-icon><RefreshRight /></el-icon>
      </el-button>

      <el-dropdown trigger="click" @command="onCommand">
        <button class="app-header__user" type="button">
          <el-avatar class="app-header__avatar" :size="32">{{ currentUserInitial }}</el-avatar>
          <span class="app-header__user-text">
            <strong>{{ currentUserName }}</strong>
            <small>{{ currentUserRole }}</small>
          </span>
          <el-icon class="app-header__user-arrow"><ArrowDown /></el-icon>
        </button>

        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><UserFilled /></el-icon>
              个人中心
            </el-dropdown-item>
            <el-dropdown-item command="refresh">刷新页面</el-dropdown-item>
            <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>
