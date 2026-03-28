<script setup lang="ts">
import { ref } from 'vue';

import type { PluginMenu } from '@/types/plugin';
import { createPluginMenuNode } from '@/utils/plugin';

defineOptions({ name: 'PluginMenuTreeEditor' });

const emit = defineEmits<{
  (event: 'move-node', sourceId: string, targetId: string, position: 'before' | 'after' | 'inside'): void;
}>();

const props = defineProps<{
  menus: PluginMenu[];
  pluginName: string;
}>();

const mutableMenus = props.menus as PluginMenu[];
const draggingId = ref('');
const dropHint = ref<{ targetId: string; position: 'before' | 'after' | 'inside' } | null>(null);

function addRootMenu() {
  mutableMenus.push(createPluginMenuNode(props.pluginName));
}

function addChildMenu(menu: PluginMenu) {
  menu.children = menu.children ?? [];
  menu.children.push(createPluginMenuNode(props.pluginName, menu.id));
}

function removeMenu(list: PluginMenu[], index: number) {
  list.splice(index, 1);
}

function forwardMove(sourceId: string, targetId: string, position: 'before' | 'after' | 'inside') {
  emit('move-node', sourceId, targetId, position);
}

function onDragStart(event: DragEvent, menu: PluginMenu) {
  draggingId.value = menu.id;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', menu.id);
  }
}

function onDragEnd() {
  draggingId.value = '';
  dropHint.value = null;
}

function setDropHint(targetId: string, position: 'before' | 'after' | 'inside') {
  dropHint.value = { targetId, position };
}

function clearDropHint() {
  dropHint.value = null;
}

function isDropHint(targetId: string, position: 'before' | 'after' | 'inside') {
  return dropHint.value?.targetId === targetId && dropHint.value?.position === position;
}

function onDrop(event: DragEvent, targetId: string, position: 'before' | 'after' | 'inside') {
  event.preventDefault();
  const sourceId = event.dataTransfer?.getData('text/plain') || draggingId.value;
  clearDropHint();
  if (!sourceId || sourceId === targetId) {
    return;
  }
  emit('move-node', sourceId, targetId, position);
}
</script>

<template>
  <div class="plugin-menu-tree-editor">
    <div class="admin-table__actions mb-12">
      <el-button type="primary" plain @click="addRootMenu">新增根菜单</el-button>
    </div>

    <el-empty v-if="mutableMenus.length === 0" description="暂无菜单，请先新增根菜单" />

    <div v-else class="plugin-menu-tree-editor__list">
      <div
        v-for="(menu, index) in mutableMenus"
        :key="menu.id"
        class="plugin-menu-tree-editor__node"
      >
        <div
          class="plugin-menu-tree-editor__dropzone"
          :class="{ 'is-active': isDropHint(menu.id, 'before') }"
          @dragover.prevent="setDropHint(menu.id, 'before')"
          @dragleave="clearDropHint"
          @drop="onDrop($event, menu.id, 'before')"
        >
          拖到这里，放在 <strong>{{ menu.name || menu.id || '当前菜单' }}</strong> 之前
        </div>

        <el-card
          shadow="never"
          class="plugin-menu-tree-editor__card"
          draggable="true"
          @dragstart="onDragStart($event, menu)"
          @dragend="onDragEnd"
        >
        <template #header>
          <div class="page-card__header">
            <span>{{ menu.name || menu.id || '未命名菜单' }}</span>
            <el-space wrap>
              <el-tag effect="plain">{{ menu.type || 'menu' }}</el-tag>
              <el-tag effect="plain" type="info">拖拽排序</el-tag>
              <el-button size="small" type="primary" plain @click="addChildMenu(menu)">新增子菜单</el-button>
              <el-button size="small" type="danger" plain @click="removeMenu(mutableMenus, index)">删除</el-button>
            </el-space>
          </div>
        </template>

        <div
          class="plugin-menu-tree-editor__dropzone plugin-menu-tree-editor__dropzone--inside"
          :class="{ 'is-active': isDropHint(menu.id, 'inside') }"
          @dragover.prevent="setDropHint(menu.id, 'inside')"
          @dragleave="clearDropHint"
          @drop="onDrop($event, menu.id, 'inside')"
        >
          拖到这里，作为 <strong>{{ menu.name || menu.id || '当前菜单' }}</strong> 的子级
        </div>

        <el-form label-width="96px" class="admin-form admin-form--two-col">
          <el-form-item label="菜单 ID">
            <el-input v-model="menu.id" placeholder="唯一 ID" />
          </el-form-item>
          <el-form-item label="父级 ID">
            <el-input v-model="menu.parent_id" placeholder="父级菜单 ID" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="menu.name" placeholder="菜单名称" />
          </el-form-item>
          <el-form-item label="路径" required>
            <el-input v-model="menu.path" placeholder="/plugin/example/home" />
          </el-form-item>
          <el-form-item label="组件">
            <el-input v-model="menu.component" placeholder="view/plugin/example/index" />
          </el-form-item>
          <el-form-item label="图标">
            <el-input v-model="menu.icon" placeholder="sparkles" />
          </el-form-item>
          <el-form-item label="权限标识">
            <el-input v-model="menu.permission" placeholder="plugin:example:view" />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="menu.type" style="width: 100%">
              <el-option label="目录" value="directory" />
              <el-option label="菜单" value="menu" />
              <el-option label="按钮" value="button" />
            </el-select>
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="menu.sort" :min="0" :step="1" style="width: 100%" />
          </el-form-item>
          <el-form-item label="重定向">
            <el-input v-model="menu.redirect" placeholder="/plugin/example/home" />
          </el-form-item>
          <el-form-item label="外链地址">
            <el-input v-model="menu.external_url" placeholder="可选" />
          </el-form-item>
          <el-form-item label="可见">
            <el-switch v-model="menu.visible" />
          </el-form-item>
          <el-form-item label="启用">
            <el-switch v-model="menu.enabled" />
          </el-form-item>
        </el-form>

        <div v-if="menu.children && menu.children.length > 0" class="plugin-menu-tree-editor__children">
          <PluginMenuTreeEditor :menus="menu.children" :plugin-name="pluginName" @move-node="forwardMove" />
        </div>

        <div
          class="plugin-menu-tree-editor__dropzone"
          :class="{ 'is-active': isDropHint(menu.id, 'after') }"
          @dragover.prevent="setDropHint(menu.id, 'after')"
          @dragleave="clearDropHint"
          @drop="onDrop($event, menu.id, 'after')"
        >
          拖到这里，放在 <strong>{{ menu.name || menu.id || '当前菜单' }}</strong> 之后
        </div>
      </el-card>
    </div>
  </div>
</template>
