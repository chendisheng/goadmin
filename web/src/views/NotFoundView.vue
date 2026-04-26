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
      :title="t('common.not_found', '页面不存在')"
      :sub-title="t('common.not_found_hint', '请返回工作台继续操作')"
    >
      <template #extra>
        <el-button type="primary" @click="router.push('/dashboard')">{{ t('common.back_to_dashboard', '返回工作台') }}</el-button>
      </template>
    </el-result>
  </div>
</template>
