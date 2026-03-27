<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { Expand, Fold, RefreshRight } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

import { logout as logoutApi } from '@/api/auth';
import { useAppStore } from '@/store/app';
import { useSessionStore } from '@/store/session';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const appStore = useAppStore();
const sessionStore = useSessionStore();
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

function refreshPage() {
  window.location.reload();
}

async function onLogout() {
  try {
    await logoutApi();
  } catch {
    // 退出时即使后端已失效也继续清理本地会话
  } finally {
    sessionStore.clearSession();
    ElMessage.success('已退出登录');
    await router.push({ path: '/login' });
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
      <el-tag effect="plain" round type="info">{{ import.meta.env.MODE }}</el-tag>
      <el-tag effect="plain" round type="success">{{ apiBaseUrl }}</el-tag>
      <el-tag v-if="sessionStore.displayName" effect="plain" round type="warning">
        {{ sessionStore.displayName }}
      </el-tag>
      <el-button text type="primary" @click="onLogout">退出</el-button>
      <el-button circle text @click="refreshPage">
        <el-icon><RefreshRight /></el-icon>
      </el-button>
    </div>
  </el-header>
</template>
