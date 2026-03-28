<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';

import { pingExamplePlugin } from '@/api/plugins';
import type { ExamplePluginPingResponse } from '@/types/plugin';

const route = useRoute();
const loading = ref(false);
const pingResult = ref<ExamplePluginPingResponse | null>(null);

const pageTitle = computed(() => (typeof route.meta.title === 'string' && route.meta.title.trim() !== '' ? route.meta.title : '插件示例'));
const componentName = computed(() => String(route.meta.componentName || 'view/plugin/example/index'));
const routePath = computed(() => route.path);
const routePermission = computed(() => String(route.meta.permission || 'plugin:example:view'));

async function handlePing() {
  loading.value = true;
  try {
    pingResult.value = await pingExamplePlugin();
    ElMessage.success('插件接口调用成功');
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '插件接口调用失败');
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="plugin-example-page">
    <el-card class="page-card" shadow="never">
      <template #header>
        <div class="page-card__header">
          <span>{{ pageTitle }}</span>
          <el-tag effect="plain" round type="success">Plugin UI</el-tag>
        </div>
      </template>

      <el-alert
        title="这是一个由插件注册的动态页面"
        description="页面组件路径来自后端菜单配置 `view/plugin/example/index`，并通过前端动态组件映射加载。"
        type="success"
        show-icon
        :closable="false"
      />

      <el-descriptions :column="1" border size="small" class="plugin-example-page__meta">
        <el-descriptions-item label="路由路径">{{ routePath }}</el-descriptions-item>
        <el-descriptions-item label="组件标识">{{ componentName }}</el-descriptions-item>
        <el-descriptions-item label="权限标识">{{ routePermission }}</el-descriptions-item>
      </el-descriptions>

      <div class="plugin-example-page__actions">
        <el-button type="primary" :loading="loading" @click="handlePing">调用插件接口</el-button>
      </div>

      <el-result
        v-if="pingResult"
        icon="success"
        title="插件接口返回成功"
        :sub-title="`${pingResult.message} (${pingResult.plugin})`"
      />
    </el-card>
  </div>
</template>
