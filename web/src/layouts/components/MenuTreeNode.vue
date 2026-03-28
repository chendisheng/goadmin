<script setup lang="ts">
import { computed } from 'vue';

import { resolveMenuIcon } from '@/utils/menu-icon';
import type { SidebarMenuNode } from '@/types/menu';

defineOptions({ name: 'MenuTreeNode' });

const props = defineProps<{
  node: SidebarMenuNode;
}>();

const hasChildren = computed(() => props.node.children.length > 0);
const iconComponent = computed(() => resolveMenuIcon(props.node.icon));
</script>

<template>
  <el-sub-menu v-if="hasChildren" :index="node.path">
    <template #title>
      <el-icon>
        <component :is="iconComponent" />
      </el-icon>
      <span>{{ node.title }}</span>
    </template>

    <MenuTreeNode v-for="child in node.children" :key="child.path" :node="child" />
  </el-sub-menu>

  <el-menu-item v-else :index="node.path">
    <el-icon>
      <component :is="iconComponent" />
    </el-icon>
    <span>{{ node.title }}</span>
  </el-menu-item>
</template>
