<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

import { fetchAuthorizationStatus, reloadAuthorizationPolicies, seedAuthorizationPolicies, type AuthorizationModuleStatus } from '@/api/casbin';
import AdminTable from '@/components/admin/AdminTable.vue';

const router = useRouter();
const loading = ref(false);
const actionLoading = ref(false);
const status = ref<AuthorizationModuleStatus>({});

const statusTag = computed(() => (status.value.enabled ? 'success' : 'info'));
const statusText = computed(() => (status.value.enabled ? '已启用' : '未启用'));

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
    ElMessage.success('授权模块已重新加载');
    await loadStatus();
  } finally {
    actionLoading.value = false;
  }
}

async function handleSeed() {
  actionLoading.value = true;
  try {
    await seedAuthorizationPolicies();
    ElMessage.success('授权模块默认策略已补齐');
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
      title="授权治理"
      description="统一管理授权运行时、默认策略与模型、策略入口。"
      :loading="loading"
    >
      <template #actions>
        <el-button :loading="loading" @click="loadStatus">刷新状态</el-button>
        <el-button :loading="actionLoading" type="primary" @click="handleReload">重载运行时</el-button>
        <el-button :loading="actionLoading" @click="handleSeed">补齐默认策略</el-button>
        <el-button @click="openModels">模型管理</el-button>
        <el-button @click="openRules">策略管理</el-button>
      </template>

      <el-row :gutter="16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>运行状态</template>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="启用状态">
                <el-tag :type="statusTag" effect="plain">{{ statusText }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="来源">{{ status.source || '-' }}</el-descriptions-item>
              <el-descriptions-item label="模型路径">{{ status.model_path || '-' }}</el-descriptions-item>
              <el-descriptions-item label="策略路径">{{ status.policy_path || '-' }}</el-descriptions-item>
            </el-descriptions>
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never" class="mb-16">
            <template #header>治理摘要</template>
            <p class="casbin-summary">{{ status.summary || '暂无摘要' }}</p>
            <el-divider />
            <div>
              <strong>关联入口</strong>
              <ul class="casbin-list">
                <li v-for="item in status.legacy_modules || []" :key="item">{{ item }}</li>
              </ul>
            </div>
            <div>
              <strong>可用接口</strong>
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
