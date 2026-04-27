<script setup lang="ts">
import { ref } from 'vue';

import { useAppI18n } from '@/i18n';
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

const { t } = useAppI18n();
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

function getMenuDisplayTitle(menu: PluginMenu): string {
  return t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id || t('plugin.menu_unnamed', 'Unnamed menu'));
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
      <el-button type="primary" plain @click="addRootMenu">{{ t('plugin.add_root_menu', 'Add root menu') }}</el-button>
    </div>

    <el-empty v-if="mutableMenus.length === 0" :description="t('plugin.no_menus', 'No menus yet, please add a root menu first')" />

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
          {{ t('plugin.drop_before', 'Drop here to place before {name}', { name: menu.name || menu.id || t('plugin.current_menu', 'Current menu') }) }}
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
            <span>{{ getMenuDisplayTitle(menu) }}</span>
            <el-space wrap>
              <el-tag effect="plain">{{ menu.type || t('plugin.menu_type_menu', 'Menu') }}</el-tag>
              <el-tag effect="plain" type="info">{{ t('plugin.drag_sorting', 'Drag sorting') }}</el-tag>
              <el-button size="small" type="primary" plain @click="addChildMenu(menu)">{{ t('plugin.add_child_menu', 'Add child menu') }}</el-button>
              <el-button size="small" type="danger" plain @click="removeMenu(mutableMenus, index)">{{ t('common.delete', 'Delete') }}</el-button>
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
          {{ t('plugin.drop_inside', 'Drop here to place as a child of {name}', { name: menu.name || menu.id || t('plugin.current_menu', 'Current menu') }) }}
        </div>

        <el-form label-width="96px" class="admin-form admin-form--two-col">
          <el-form-item :label="t('plugin.menu_id', 'Menu ID')">
            <el-input v-model="menu.id" :placeholder="t('plugin.menu_id_placeholder', 'Unique ID')" />
          </el-form-item>
          <el-form-item :label="t('plugin.parent_id', 'Parent ID')">
            <el-input v-model="menu.parent_id" :placeholder="t('plugin.parent_id_placeholder', 'Parent menu ID')" />
          </el-form-item>
          <el-form-item :label="t('plugin.menu_name', 'Name')" required>
            <el-input v-model="menu.name" :placeholder="t('plugin.menu_name_placeholder', 'Menu name')" />
          </el-form-item>
          <el-form-item :label="t('plugin.menu_title_key', 'Title key')">
            <el-input v-model="menu.titleKey" :placeholder="t('plugin.menu_title_key_placeholder', 'For example, route.dashboard')" />
          </el-form-item>
          <el-form-item :label="t('plugin.menu_title_default', 'Default title')">
            <el-input v-model="menu.titleDefault" :placeholder="t('plugin.menu_title_default_placeholder', 'For example, Dashboard')" />
          </el-form-item>
          <el-form-item :label="t('plugin.path', 'Path')" required>
            <el-input v-model="menu.path" :placeholder="t('plugin.menu_path_placeholder', '/plugin/example/home')" />
          </el-form-item>
          <el-form-item :label="t('plugin.component', 'Component')">
            <el-input v-model="menu.component" :placeholder="t('plugin.menu_component_placeholder_detail', 'view/plugin/example/index')" />
          </el-form-item>
          <el-form-item :label="t('plugin.icon', 'Icon')">
            <el-input v-model="menu.icon" :placeholder="t('plugin.icon_placeholder', 'sparkles')" />
          </el-form-item>
          <el-form-item :label="t('plugin.permission_key', 'Permission key')">
            <el-input v-model="menu.permission" :placeholder="t('plugin.permission_key_placeholder', 'plugin:example:view')" />
          </el-form-item>
          <el-form-item :label="t('plugin.type', 'Type')">
            <el-select v-model="menu.type" style="width: 100%">
              <el-option :label="t('plugin.menu_type_directory', 'Directory')" value="directory" />
              <el-option :label="t('plugin.menu_type_menu', 'Menu')" value="menu" />
              <el-option :label="t('plugin.menu_type_button', 'Button')" value="button" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('plugin.sort', 'Sort')">
            <el-input-number v-model="menu.sort" :min="0" :step="1" style="width: 100%" />
          </el-form-item>
          <el-form-item :label="t('plugin.redirect', 'Redirect')">
            <el-input v-model="menu.redirect" :placeholder="t('plugin.redirect_placeholder', '/plugin/example/home')" />
          </el-form-item>
          <el-form-item :label="t('plugin.external_url', 'External URL')">
            <el-input v-model="menu.external_url" :placeholder="t('plugin.optional', 'Optional')" />
          </el-form-item>
          <el-form-item :label="t('plugin.visible', 'Visible')">
            <el-switch v-model="menu.visible" />
          </el-form-item>
          <el-form-item :label="t('plugin.enabled', 'Enabled')">
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
          {{ t('plugin.drop_after', 'Drop here to place after {name}', { name: getMenuDisplayTitle(menu) }) }}
        </div>
      </el-card>
      </div>
    </div>
  </div>
</template>
