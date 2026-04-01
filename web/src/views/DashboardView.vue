<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';

import { fetchHealth, type HealthPayload } from '@/api';
import { useAppStore } from '@/store/app';
import { useSessionStore } from '@/store/session';

const appTitle = import.meta.env.VITE_APP_TITLE || 'GoAdmin';
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1';
const appStore = useAppStore();
const sessionStore = useSessionStore();

const healthState = ref<HealthPayload | null>(null);
const loading = ref(false);
const errorMessage = ref('');

const shellStatus = computed(() => (appStore.sidebarCollapsed ? 'Sidebar collapsed' : 'Sidebar expanded'));
const currentUser = computed(() => sessionStore.currentUser);

const dashboardMetrics = computed(() => [
  {
    label: 'API 基址',
    value: apiBaseUrl,
    note: 'Axios 统一请求入口',
  },
  {
    label: '布局状态',
    value: shellStatus.value,
    note: '侧边栏折叠状态已持久化',
  },
  {
    label: '当前用户',
    value: sessionStore.displayName || 'System Admin',
    note: '会话信息已加载',
  },
  {
    label: '登录模式',
    value: 'JWT / RBAC',
    note: '按钮权限与菜单权限待扩展',
  },
]);

async function onPingHealth() {
  loading.value = true;
  errorMessage.value = '';
  try {
    healthState.value = await fetchHealth();
    ElMessage.success('健康检查请求成功');
  } catch (error) {
    const message = error instanceof Error ? error.message : '健康检查请求失败';
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
              <span>系统概览</span>
              <el-tag effect="plain" round type="success">在线</el-tag>
            </div>
          </template>

          <div class="dashboard-hero__content">
            <div>
              <h2>{{ appTitle }} 管理后台</h2>
              <p>统一的侧边栏、顶部导航与工作台首页，后续可直接承接 Auth、CRUD 和 Plugin 功能模块。</p>
            </div>

            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="应用标题">{{ appTitle }}</el-descriptions-item>
              <el-descriptions-item label="API 基址">{{ apiBaseUrl }}</el-descriptions-item>
              <el-descriptions-item label="布局状态">{{ shellStatus }}</el-descriptions-item>
              <el-descriptions-item label="当前用户">
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
              <span>API 验证</span>
              <el-tag effect="plain" round type="info">连通性</el-tag>
            </div>
          </template>

          <div class="dashboard-quick-actions__body">
            <p>点击按钮发送一次健康检查请求，快速验证前后端联通性与 Axios 拦截器。</p>
            <el-button type="primary" :loading="loading" @click="onPingHealth">发送健康检查</el-button>
          </div>

          <el-divider />

          <div v-if="healthState" class="dashboard-health-result">
            <el-descriptions :column="1" size="small">
              <el-descriptions-item label="status">{{ healthState.status }}</el-descriptions-item>
              <el-descriptions-item label="uptime">{{ healthState.uptime }}</el-descriptions-item>
              <el-descriptions-item label="timestamp">{{ healthState.timestamp }}</el-descriptions-item>
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
              <el-descriptions-item label="username">{{ currentUser.username }}</el-descriptions-item>
              <el-descriptions-item label="user_id">{{ currentUser.user_id }}</el-descriptions-item>
              <el-descriptions-item label="roles">{{ currentUser.roles?.join(', ') || '-' }}</el-descriptions-item>
            </el-descriptions>
          </template>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="dashboard-secondary-row">
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>工程规范</template>
          <ul class="dashboard-list">
            <li>Vue 3 + TypeScript</li>
            <li>Vite 构建与热更新</li>
            <li>Element Plus 统一界面</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>状态中心</template>
          <ul class="dashboard-list">
            <li>Pinia 全局 Store 已初始化</li>
            <li>侧栏折叠状态持久化</li>
            <li>预留会话 Token 基础能力</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>功能规划</template>
          <ul class="dashboard-list">
            <li>Admin Modules 基础管理页</li>
            <li>权限控制与按钮级授权</li>
            <li>插件 UI 与动态扩展</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
