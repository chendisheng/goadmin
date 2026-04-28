<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { ArrowDown, Expand, Fold, RefreshRight, UserFilled } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

import { logout as logoutApi } from '@/api/auth';
import { preloadRouteNamespaces, resolveRouteLocaleMeta, setI18nLanguage, useAppI18n } from '@/i18n';
import { useAppStore } from '@/store/app';
import { useLocaleStore } from '@/store/locale';
import { useMenuStore } from '@/store/menu';
import { useSessionStore } from '@/store/session';
import { useTabsStore } from '@/store/tabs';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const buildMode = import.meta.env.MODE;
const appStore = useAppStore();
const localeStore = useLocaleStore();
const menuStore = useMenuStore();
const sessionStore = useSessionStore();
const tabsStore = useTabsStore();
const router = useRouter();
const route = useRoute();
const { t } = useAppI18n();

const pageTitle = computed(() => {
  const localized = resolveRouteLocaleMeta(route);
  return localized.title.trim() !== '' ? localized.title : t('app.title', appTitle);
});

const pageSubtitle = computed(() => {
  const localized = resolveRouteLocaleMeta(route);
  if (localized.subtitle.trim() !== '') {
    return localized.subtitle;
  }
  return t('app.subtitle', 'Vue 3 + TypeScript + Vite + Pinia + Axios');
});

const currentUserName = computed(() => sessionStore.displayName || t('common.visitor', '访客'));

const currentUserRole = computed(() => {
  const role = sessionStore.currentUser?.roles?.[0];
  return typeof role === 'string' && role.trim() !== '' ? role : t('common.admin_role', '管理员');
});

const currentUserInitial = computed(() => {
  const source = currentUserName.value.trim();
  if (source.length === 0) {
    return 'G';
  }
  return source.slice(0, 1).toUpperCase();
});

const currentLanguageLabel = computed(() => {
  return localeStore.language === 'en-US' ? t('common.language_en', 'English') : t('common.language_zh', 'Chinese');
});

function refreshPage() {
  window.location.reload();
}

async function switchLanguage(language: 'zh-CN' | 'en-US') {
  if (localeStore.language === language) {
    return;
  }

  const profileLanguage = sessionStore.currentUser?.language ?? null;
  await preloadRouteNamespaces(route, language);
  await setI18nLanguage(language);
  localeStore.applyLanguagePreference(language, profileLanguage);
  sessionStore.setLanguage(language, profileLanguage);
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
    ElMessage.success(t('common.logged_out', 'Logged out'));
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
      <el-dropdown trigger="click" popper-class="app-language-dropdown" @command="switchLanguage">
        <el-button class="app-header__language" text>
          {{ t('common.language', 'Language') }}：{{ currentLanguageLabel }}
          <el-icon class="app-header__language-arrow"><ArrowDown /></el-icon>
        </el-button>

        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="zh-CN">{{ t('common.language_zh', 'Chinese') }}</el-dropdown-item>
            <el-dropdown-item command="en-US">{{ t('common.language_en', 'English') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
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
              {{ t('common.personal_center', '个人中心') }}
            </el-dropdown-item>
            <el-dropdown-item command="refresh">{{ t('common.refresh_page', '刷新页面') }}</el-dropdown-item>
            <el-dropdown-item command="logout" divided>{{ t('common.logout', '退出登录') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>
