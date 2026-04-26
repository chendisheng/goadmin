<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

import { fetchAuthorizationStatus, reloadAuthorizationPolicies, seedAuthorizationPolicies, type AuthorizationModuleStatus } from '@/api/casbin';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';

const router = useRouter();
const loading = ref(false);
const actionLoading = ref(false);
const status = ref<AuthorizationModuleStatus>({});
const { t } = useAppI18n();

const statusTag = computed(() => (status.value.enabled ? 'success' : 'info'));
const statusText = computed(() => (status.value.enabled ? t('casbin.enabled', '已启用') : t('casbin.disabled', '未启用')));

async function loadStatus() {
  loading.value = true;
  try {
    status.value = await fetchAuthorizationStatus();
  } finally {
    loading.value = false;
  }
}

async function handleReload() {
  actionLoading.value = true;
  try {
    await reloadAuthorizationPolicies();
    ElMessage.success(t('casbin.reload_success', '授权模块已重新加载'));
    await loadStatus();
  } finally {
    actionLoading.value = false;
  }
}

async function handleSeed() {
  actionLoading.value = true;
  try {
    await seedAuthorizationPolicies();
    ElMessage.success(t('casbin.seed_success', '授权模块默认策略已补齐'));
    await loadStatus();
  } finally {
    actionLoading.value = false;
  }
}

function openModels() {
  void router.push('/system/casbin/models');
}

function openRules() {
  void router.push('/system/casbin/rules');
}

onMounted(() => {
  void loadStatus();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      :title="t('casbin.title', '授权治理')"
      :description="t('casbin.description', '统一管理授权运行时、默认策略与模型、策略入口。')"
      :loading="loading"
    >
      <template #actions>
        <el-button :loading="loading" @click="loadStatus">{{ t('casbin.refresh_status', '刷新状态') }}</el-button>
        <el-button :loading="actionLoading" type="primary" @click="handleReload">{{ t('casbin.reload_runtime', '重载运行时') }}</el-button>
        <el-button :loading="actionLoading" @click="handleSeed">{{ t('casbin.seed_default', '补齐默认策略') }}</el-button>
        <el-button @click="openModels">{{ t('casbin.models', '模型管理') }}</el-button>
        <el-button @click="openRules">{{ t('casbin.rules', '策略管理') }}</el-button>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>{{ t('casbin.status_panel', '运行状态') }}</template>
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="t('casbin.enabled_status', '启用状态')">
                <el-tag :type="statusTag" effect="plain">{{ statusText }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('casbin.source', '来源')">{{ status.source || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('casbin.model_path', '模型路径')">{{ status.model_path || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('casbin.policy_path', '策略路径')">{{ status.policy_path || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>{{ t('casbin.summary_title', '治理摘要') }}</template>
            <p class="casbin-summary">{{ status.summary || t('casbin.no_summary', '暂无摘要') }}</p>
            <el-divider />
            <div>
              <strong>{{ t('casbin.legacy_modules', '关联入口') }}</strong>
              <ul class="casbin-list">
                <li v-for="item in status.legacy_modules || []" :key="item">{{ item }}</li>
              </ul>
            </div>
            <div>
              <strong>{{ t('casbin.available_routes', '可用接口') }}</strong>
              <ul class="casbin-list">
                <li v-for="item in status.routes || []" :key="item">{{ item }}</li>
              </ul>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </AdminTable>
  </div>
</template>
