<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';

import { fetchHealth, type HealthPayload } from '@/api';
import { useAppI18n } from '@/i18n';
import { useAppStore } from '@/store/app';
import { useSessionStore } from '@/store/session';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const appStore = useAppStore();
const sessionStore = useSessionStore();
const { t } = useAppI18n();

const healthState = ref<HealthPayload | null>(null);
const loading = ref(false);
const errorMessage = ref('');

const shellStatus = computed(() => (appStore.sidebarCollapsed ? t('dashboard.sidebar_collapsed', 'Sidebar collapsed') : t('dashboard.sidebar_expanded', 'Sidebar expanded')));
const currentUser = computed(() => sessionStore.currentUser);

const dashboardMetrics = computed(() => [
  {
    label: t('dashboard.metric.api_base_url', 'API base URL'),
    value: apiBaseUrl,
    note: t('dashboard.metric.api_base_url_note', 'Unified Axios request entry'),
  },
  {
    label: t('dashboard.metric.layout_state', 'Layout state'),
    value: shellStatus.value,
    note: t('dashboard.metric.layout_state_note', 'Sidebar collapse state persisted'),
  },
  {
    label: t('dashboard.metric.current_user', 'Current user'),
    value: sessionStore.displayName || t('dashboard.metric.default_user', 'System administrator'),
    note: t('dashboard.metric.current_user_note', 'Session information loaded'),
  },
  {
    label: t('dashboard.metric.login_mode', 'Login mode'),
    value: t('dashboard.metric.login_mode_value', 'JWT / RBAC'),
    note: t('dashboard.metric.login_mode_note', 'Button and menu permissions will be extended'),
  },
]);

async function onPingHealth() {
  loading.value = true;
  errorMessage.value = '';
  try {
    healthState.value = await fetchHealth();
    ElMessage.success(t('dashboard.health_success', 'Health check request succeeded'));
  } catch (error) {
    const message = error instanceof Error ? error.message : t('dashboard.health_failed', 'Health check request failed');
    errorMessage.value = message;
    ElMessage.error(message);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="dashboard-page">
    <el-row :gutter="16" class="dashboard-metrics">
      <el-col v-for="metric in dashboardMetrics" :key="metric.label" :xs="24" :sm="12" :lg="6">
        <el-card class="page-card dashboard-metric-card" shadow="never">
          <div class="dashboard-metric-card__label">{{ metric.label }}</div>
          <div class="dashboard-metric-card__value">{{ metric.value }}</div>
          <div class="dashboard-metric-card__note">{{ metric.note }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <el-card class="page-card dashboard-hero" shadow="never">
          <template #header>
            <div class="page-card__header">
              <span>{{ t('dashboard.hero_title', 'System overview') }}</span>
              <el-tag effect="plain" round type="success">{{ t('dashboard.status_online', 'Online') }}</el-tag>
            </div>
          </template>

          <div class="dashboard-hero__content">
            <div>
              <h2>{{ t('dashboard.hero_heading', '{title} admin console', { title: appTitle }) }}</h2>
              <p>{{ t('dashboard.hero_description', 'Unified sidebar, top navigation, and dashboard home, ready to host Auth, CRUD, and Plugin modules later.') }}</p>
            </div>

            <el-descriptions :column="1" border size="small">
              <el-descriptions-item :label="t('dashboard.hero_app_title', 'Application title')">{{ appTitle }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_api_base_url', 'API base URL')">{{ apiBaseUrl }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_layout_state', 'Layout state')">{{ shellStatus }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_current_user', 'Current user')">
                {{ sessionStore.displayName }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card class="page-card dashboard-quick-actions" shadow="never">
          <template #header>
            <div class="page-card__header">
              <span>{{ t('dashboard.quick_actions_title', 'API validation') }}</span>
              <el-tag effect="plain" round type="info">{{ t('dashboard.quick_actions_tag', 'Connectivity') }}</el-tag>
            </div>
          </template>

          <div class="dashboard-quick-actions__body">
            <p>{{ t('dashboard.quick_actions_description', 'Click the button to send a health check request and quickly verify frontend-server connectivity and Axios interceptors.') }}</p>
            <el-button type="primary" :loading="loading" @click="onPingHealth">{{ t('dashboard.health_check_button', 'Send health check') }}</el-button>
          </div>

          <el-divider />

          <div v-if="healthState" class="dashboard-health-result">
            <el-descriptions :column="1" size="small">
              <el-descriptions-item :label="t('dashboard.health_status', 'status')">{{ healthState.status }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.health_uptime', 'uptime')">{{ healthState.uptime }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.health_timestamp', 'timestamp')">{{ healthState.timestamp }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <el-alert
            v-if="errorMessage"
            :title="errorMessage"
            type="error"
            show-icon
            :closable="false"
          />

          <template v-if="currentUser">
            <el-divider />
            <el-descriptions :column="1" size="small">
              <el-descriptions-item :label="t('dashboard.user_username', 'username')">{{ currentUser.username }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.user_id', 'user_id')">{{ currentUser.user_id }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.user_roles', 'roles')">{{ currentUser.roles?.join(', ') || '-' }}</el-descriptions-item>
            </el-descriptions>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="dashboard-secondary-row">
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>{{ t('dashboard.section.engineering', 'Engineering standards') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.engineering.vue', 'Vue 3 + TypeScript') }}</li>
            <li>{{ t('dashboard.engineering.vite', 'Vite build and hot reload') }}</li>
            <li>{{ t('dashboard.engineering.element_plus', 'Unified Element Plus UI') }}</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>{{ t('dashboard.section.status', 'Status center') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.status.pinia', 'Pinia global store initialized') }}</li>
            <li>{{ t('dashboard.status.sidebar_persisted', 'Sidebar collapse state persisted') }}</li>
            <li>{{ t('dashboard.status.token_reserved', 'Session token foundation reserved') }}</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>{{ t('dashboard.section.plan', 'Feature roadmap') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.plan.modules', 'Admin Modules base management page') }}</li>
            <li>{{ t('dashboard.plan.permissions', 'Permission control and button-level authorization') }}</li>
            <li>{{ t('dashboard.plan.plugin_ui', 'Plugin UI and dynamic extension') }}</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
