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

const shellStatus = computed(() => (appStore.sidebarCollapsed ? t('dashboard.sidebar_collapsed', '侧栏已收起') : t('dashboard.sidebar_expanded', '侧栏已展开')));
const currentUser = computed(() => sessionStore.currentUser);

const dashboardMetrics = computed(() => [
  {
    label: t('dashboard.metric.api_base_url', 'API 基址'),
    value: apiBaseUrl,
    note: t('dashboard.metric.api_base_url_note', 'Axios 统一请求入口'),
  },
  {
    label: t('dashboard.metric.layout_state', '布局状态'),
    value: shellStatus.value,
    note: t('dashboard.metric.layout_state_note', '侧边栏折叠状态已持久化'),
  },
  {
    label: t('dashboard.metric.current_user', '当前用户'),
    value: sessionStore.displayName || t('dashboard.metric.default_user', '系统管理员'),
    note: t('dashboard.metric.current_user_note', '会话信息已加载'),
  },
  {
    label: t('dashboard.metric.login_mode', '登录模式'),
    value: t('dashboard.metric.login_mode_value', 'JWT / RBAC'),
    note: t('dashboard.metric.login_mode_note', '按钮权限与菜单权限待扩展'),
  },
]);

async function onPingHealth() {
  loading.value = true;
  errorMessage.value = '';
  try {
    healthState.value = await fetchHealth();
    ElMessage.success(t('dashboard.health_success', '健康检查请求成功'));
  } catch (error) {
    const message = error instanceof Error ? error.message : t('dashboard.health_failed', '健康检查请求失败');
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
              <span>{{ t('dashboard.hero_title', '系统概览') }}</span>
              <el-tag effect="plain" round type="success">{{ t('dashboard.status_online', '在线') }}</el-tag>
            </div>
          </template>

          <div class="dashboard-hero__content">
            <div>
              <h2>{{ t('dashboard.hero_heading', '{title} 管理后台', { title: appTitle }) }}</h2>
              <p>{{ t('dashboard.hero_description', '统一的侧边栏、顶部导航与工作台首页，后续可直接承接 Auth、CRUD 和 Plugin 功能模块。') }}</p>
            </div>

            <el-descriptions :column="1" border size="small">
              <el-descriptions-item :label="t('dashboard.hero_app_title', '应用标题')">{{ appTitle }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_api_base_url', 'API 基址')">{{ apiBaseUrl }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_layout_state', '布局状态')">{{ shellStatus }}</el-descriptions-item>
              <el-descriptions-item :label="t('dashboard.hero_current_user', '当前用户')">
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
              <span>{{ t('dashboard.quick_actions_title', 'API 验证') }}</span>
              <el-tag effect="plain" round type="info">{{ t('dashboard.quick_actions_tag', '连通性') }}</el-tag>
            </div>
          </template>

          <div class="dashboard-quick-actions__body">
            <p>{{ t('dashboard.quick_actions_description', '点击按钮发送一次健康检查请求，快速验证前后端联通性与 Axios 拦截器。') }}</p>
            <el-button type="primary" :loading="loading" @click="onPingHealth">{{ t('dashboard.health_check_button', '发送健康检查') }}</el-button>
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
          <template #header>{{ t('dashboard.section.engineering', '工程规范') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.engineering.vue', 'Vue 3 + TypeScript') }}</li>
            <li>{{ t('dashboard.engineering.vite', 'Vite 构建与热更新') }}</li>
            <li>{{ t('dashboard.engineering.element_plus', 'Element Plus 统一界面') }}</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>{{ t('dashboard.section.status', '状态中心') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.status.pinia', 'Pinia 全局 Store 已初始化') }}</li>
            <li>{{ t('dashboard.status.sidebar_persisted', '侧栏折叠状态持久化') }}</li>
            <li>{{ t('dashboard.status.token_reserved', '预留会话 Token 基础能力') }}</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>{{ t('dashboard.section.plan', '功能规划') }}</template>
          <ul class="dashboard-list">
            <li>{{ t('dashboard.plan.modules', 'Admin Modules 基础管理页') }}</li>
            <li>{{ t('dashboard.plan.permissions', '权限控制与按钮级授权') }}</li>
            <li>{{ t('dashboard.plan.plugin_ui', '插件 UI 与动态扩展') }}</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
