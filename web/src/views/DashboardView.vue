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
    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <el-card class="page-card dashboard-hero" shadow="never">
          <template #header>
            <div class="page-card__header">
              <span>Frontend Core</span>
              <el-tag effect="plain" round type="success">Phase 10</el-tag>
            </div>
          </template>

          <div class="dashboard-hero__content">
            <div>
              <h2>{{ appTitle }} 前端骨架</h2>
              <p>当前项目已具备 Router、Pinia、Axios、Layout 和基础主题入口，可直接承接后续 Auth / Menu / CRUD 阶段。</p>
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
              <el-tag effect="plain" round type="info">/health</el-tag>
            </div>
          </template>

          <div class="dashboard-quick-actions__body">
            <p>点击按钮发送一次健康检查请求，验证 Axios 拦截器和后端联通性。</p>
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
          <template #header>工程标准</template>
          <ul class="dashboard-list">
            <li>Vue 3 + TypeScript</li>
            <li>Vite 构建与热更新</li>
            <li>Element Plus 统一界面</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>状态管理</template>
          <ul class="dashboard-list">
            <li>Pinia 全局 Store 已初始化</li>
            <li>侧栏折叠状态持久化</li>
            <li>预留会话 Token 基础能力</li>
          </ul>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="page-card" shadow="never">
          <template #header>下一阶段</template>
          <ul class="dashboard-list">
            <li>Auth 登录与 Token 管理</li>
            <li>后端菜单驱动路由</li>
            <li>管理页 CRUD 与权限控制</li>
          </ul>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
