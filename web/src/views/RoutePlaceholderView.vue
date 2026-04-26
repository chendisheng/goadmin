<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useAppI18n, resolveRouteLocaleMeta } from '@/i18n';

const route = useRoute();
const router = useRouter();
const { t } = useAppI18n();

const pageTitle = computed(() => {
  const localized = resolveRouteLocaleMeta(route);
  return localized.title.trim() !== '' ? localized.title : t('common.placeholder_route', '页面占位');
});
const componentName = computed(() => String(route.meta.componentName || t('common.unknown', '未知')));
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
          :title="t('route.placeholder.info', '该路由已由后端菜单驱动注册')"
          :description="t('route.placeholder.description', '当前页面用于承接尚未实现的业务模块，占位逻辑会在后续 Phase 13/14 中替换为真实页面。')"
          type="info"
          show-icon
          :closable="false"
        />

        <el-descriptions :column="1" border size="small" class="route-placeholder-page__meta">
          <el-descriptions-item :label="t('route.placeholder.route_path', '路由路径')">{{ routePath }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.component_name', '组件标识')">{{ componentName }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.permission', '权限标识')">{{ routePermission }}</el-descriptions-item>
          <el-descriptions-item :label="t('route.placeholder.link', '外链地址')">{{ routeLink }}</el-descriptions-item>
        </el-descriptions>

        <div class="route-placeholder-page__actions">
          <el-button type="primary" @click="goDashboard">{{ t('route.placeholder.back', '返回工作台') }}</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>
