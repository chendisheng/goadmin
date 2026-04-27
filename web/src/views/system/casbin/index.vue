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
const statusText = computed(() => (status.value.enabled ? t('casbin.enabled', 'Enabled') : t('casbin.disabled', 'Disabled')));

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
    ElMessage.success(t('casbin.reload_success', 'Authorization module reloaded'));
    await loadStatus();
  } finally {
    actionLoading.value = false;
  }
}

async function handleSeed() {
  actionLoading.value = true;
  try {
    await seedAuthorizationPolicies();
    ElMessage.success(t('casbin.seed_success', 'Default authorization policies have been seeded'));
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
      :title="t('casbin.title', 'Authorization governance')"
      :description="t('casbin.description', 'Manage authorization runtime, default policies, and model/policy entry points.')"
      :loading="loading"
    >
      <template #actions>
        <el-button :loading="loading" @click="loadStatus">{{ t('casbin.refresh_status', 'Refresh status') }}</el-button>
        <el-button :loading="actionLoading" type="primary" @click="handleReload">{{ t('casbin.reload_runtime', 'Reload runtime') }}</el-button>
        <el-button :loading="actionLoading" @click="handleSeed">{{ t('casbin.seed_default', 'Seed default policies') }}</el-button>
        <el-button @click="openModels">{{ t('casbin.models', 'Model management') }}</el-button>
        <el-button @click="openRules">{{ t('casbin.rules', 'Policy management') }}</el-button>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>{{ t('casbin.status_panel', 'Runtime status') }}</template>
            <el-descriptions :column="1" border>
              <el-descriptions-item :label="t('casbin.enabled_status', 'Enabled status')">
                <el-tag :type="statusTag" effect="plain">{{ statusText }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item :label="t('casbin.source', 'Source')">{{ status.source || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('casbin.model_path', 'Model path')">{{ status.model_path || '-' }}</el-descriptions-item>
              <el-descriptions-item :label="t('casbin.policy_path', 'Policy path')">{{ status.policy_path || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>{{ t('casbin.summary_title', 'Governance summary') }}</template>
            <p class="casbin-summary">{{ status.summary || t('casbin.no_summary', 'No summary available') }}</p>
            <el-divider />
            <div>
              <strong>{{ t('casbin.legacy_modules', 'Linked entries') }}</strong>
              <ul class="casbin-list">
                <li v-for="item in status.legacy_modules || []" :key="item">{{ item }}</li>
              </ul>
            </div>
            <div>
              <strong>{{ t('casbin.available_routes', 'Available endpoints') }}</strong>
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
