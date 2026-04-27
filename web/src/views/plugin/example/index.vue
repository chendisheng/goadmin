<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';

import { pingExamplePlugin } from '@/api/plugins';
import { resolveRouteLocaleMeta, useAppI18n } from '@/i18n';
import type { ExamplePluginPingResponse } from '@/types/plugin';

const route = useRoute();
const { t } = useAppI18n();
const loading = ref(false);
const pingResult = ref<ExamplePluginPingResponse | null>(null);

const pageTitle = computed(() => {
  const localized = resolveRouteLocaleMeta(route);
  return localized.title.trim() !== '' ? localized.title : t('plugin.example_title', 'Plugin example');
});
const componentName = computed(() => String(route.meta.componentName || 'view/plugin/example/index'));
const routePath = computed(() => route.path);
const routePermission = computed(() => String(route.meta.permission || 'plugin:example:view'));

async function handlePing() {
  loading.value = true;
  try {
    pingResult.value = await pingExamplePlugin();
    ElMessage.success(t('plugin.example_call_success', 'Plugin API call succeeded'));
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('plugin.example_call_failed', 'Plugin API call failed'));
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
          <el-tag effect="plain" round type="success">{{ t('plugin.example_badge', 'Plugin UI') }}</el-tag>
        </div>
      </template>

      <el-alert
        :title="t('plugin.example_alert_title', 'This is a dynamic page registered by a plugin')"
        :description="t('plugin.example_alert_description', 'The component path comes from backend menu configuration `view/plugin/example/index` and is loaded through the frontend dynamic component map.')"
        type="success"
        show-icon
        :closable="false"
      />

      <el-descriptions :column="1" border size="small" class="plugin-example-page__meta">
        <el-descriptions-item :label="t('plugin.example_route_path', 'Route path')">{{ routePath }}</el-descriptions-item>
        <el-descriptions-item :label="t('plugin.example_component_name', 'Component name')">{{ componentName }}</el-descriptions-item>
        <el-descriptions-item :label="t('plugin.example_permission', 'Permission key')">{{ routePermission }}</el-descriptions-item>
      </el-descriptions>

      <div class="plugin-example-page__actions">
        <el-button type="primary" :loading="loading" @click="handlePing">{{ t('plugin.example_call', 'Call plugin API') }}</el-button>
      </div>

      <el-result
        v-if="pingResult"
        icon="success"
        :title="t('plugin.example_result_title', 'Plugin API returned successfully')"
        :sub-title="`${pingResult.message} (${pingResult.plugin})`"
      />
    </el-card>
  </div>
</template>
