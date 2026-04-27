<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useAppI18n, resolveRouteLocaleMeta } from '@/i18n';

const route = useRoute();
const router = useRouter();
const { t } = useAppI18n();

const pageTitle = computed(() => {
  const localized = resolveRouteLocaleMeta(route);
  return localized.title.trim() !== '' ? localized.title : t('common.placeholder_route', 'Route placeholder');
});
const componentName = computed(() => String(route.meta.componentName || t('common.unknown', 'Unknown')));
const routePermission = computed(() => String(route.meta.permission || '-'));
const routeLink = computed(() => String(route.meta.link || '-'));
const routePath = computed(() => route.path);

function goDashboard() {
  void router.push('/dashboard');
}
</script>

<template>
  <div class="route-placeholder-page">
    <el-card class="page-card" shadow="never">
      <template #header>
        <div class="page-card__header">
          <span>{{ pageTitle }}</span>
          <el-tag effect="plain" round type="warning">{{ t('common.dynamic_route', 'Dynamic Route') }}</el-tag>
        </div>
      </template>

      <div class="route-placeholder-page__body">
        <el-alert
          :title="t('route.placeholder.info', 'This route has been registered from backend menu data')"
          :description="t('route.placeholder.description', 'This page serves as a placeholder for business modules that are not implemented yet. The placeholder logic will be replaced by real pages in later phases.')"
          type="info"
          show-icon
          :closable="false"
        />

        <el-descriptions :column="1" border size="small" class="route-placeholder-page__meta">
          <el-descriptions-item :label="t('route.placeholder.route_path', 'Route path')">{{ routePath }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.component_name', 'Component name')">{{ componentName }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.permission', 'Permission key')">{{ routePermission }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.link', 'External URL')">{{ routeLink }}</el-descriptions-item>
        </el-descriptions>

        <div class="route-placeholder-page__actions">
          <el-button type="primary" @click="goDashboard">{{ t('route.placeholder.back', 'Back to dashboard') }}</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>
