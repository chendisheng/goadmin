<script setup lang="ts">
import { useAppI18n } from '@/i18n';
import { useRouter } from 'vue-router';

const router = useRouter();
const { t } = useAppI18n();
</script>

<template>
  <div class="not-found-page">
    <el-result
      icon="warning"
      :title="t('common.not_found', 'Page not found')"
      :sub-title="t('common.not_found_hint', 'Please return to the dashboard to continue')"
    >
      <template #extra>
        <el-button type="primary" @click="router.push('/dashboard')">{{ t('common.back_to_dashboard', 'Back to dashboard') }}</el-button>
      </template>
    </el-result>
  </div>
</template>
