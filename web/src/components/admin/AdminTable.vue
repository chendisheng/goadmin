<script setup lang="ts">
defineProps<{
  title: string;
  description?: string;
  loading?: boolean;
}>();
</script>

<template>
  <el-card class="page-card admin-table" shadow="never">
    <template #header>
      <div class="page-card__header admin-table__header">
        <div class="admin-table__title">
          <strong>{{ title }}</strong>
          <p v-if="description">{{ description }}</p>
        </div>
        <div class="admin-table__actions">
          <slot name="actions" />
        </div>
      </div>
    </template>

    <div v-if="$slots.filters" class="admin-table__filters">
      <slot name="filters" />
    </div>

    <slot />

    <div v-if="$slots.footer" class="admin-table__footer">
      <slot name="footer" />
    </div>
  </el-card>
</template>
