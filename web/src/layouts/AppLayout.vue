<script setup lang="ts">
import AppHeader from './components/AppHeader.vue';
import AppSidebar from './components/AppSidebar.vue';
</script>

<template>
  <el-container class="app-layout">
    <AppSidebar />
    <el-container class="app-layout__content">
      <AppHeader />
      <el-main class="app-layout__main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
