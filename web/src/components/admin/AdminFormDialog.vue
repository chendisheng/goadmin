<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue: boolean;
    title: string;
    width?: string | number;
    loading?: boolean;
    confirmText?: string;
    cancelText?: string;
  }>(),
  {
    width: '720px',
    loading: false,
    confirmText: '保存',
    cancelText: '取消',
  },
);

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void;
  (event: 'confirm'): void;
  (event: 'cancel'): void;
}>();

function updateVisible(value: boolean) {
  emit('update:modelValue', value);
  if (!value) {
    emit('cancel');
  }
}

function handleCancel() {
  emit('update:modelValue', false);
  emit('cancel');
}

function handleConfirm() {
  emit('confirm');
}
</script>

<template>
  <el-dialog :model-value="modelValue" :title="title" :width="width" destroy-on-close @update:model-value="updateVisible">
    <slot />

    <template #footer>
      <slot name="footer">
        <el-button @click="handleCancel">{{ cancelText }}</el-button>
        <el-button type="primary" :loading="loading" @click="handleConfirm">{{ confirmText }}</el-button>
      </slot>
    </template>
  </el-dialog>
</template>
