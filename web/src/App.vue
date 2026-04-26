<script setup lang="ts">
import { useLocaleStore } from '@/store/locale';

const localeStore = useLocaleStore();
</script>

<template>
  <el-config-provider :locale="localeStore.elementLocale">
    <router-view />
  </el-config-provider>
</template>
