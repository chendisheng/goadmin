<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { Close, MoreFilled, RefreshRight } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

import { useAppI18n } from '@/i18n';
import { useTabsStore } from '@/store/tabs';
import type { WorkspaceTabRecord } from '@/types/tabs';

const router = useRouter();
const tabsStore = useTabsStore();
const contextMenuVisible = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);
const contextMenuTabId = ref('');
const tabItemRefs = new Map<string, HTMLElement>();
const { t } = useAppI18n();

type TabCommand = 'refresh' | 'close' | 'close-others' | 'close-left' | 'close-right' | 'close-all';

const contextMenuTab = computed<WorkspaceTabRecord | null>(() => tabsStore.tabs.find((tab: WorkspaceTabRecord) => tab.id === contextMenuTabId.value) ?? null);
const canCloseContextMenuTab = computed(() => contextMenuTab.value?.closable === true);
const canCloseContextMenuOthers = computed(() => (contextMenuTab.value ? hasClosableOthers(contextMenuTab.value.id) : false));
const canCloseContextMenuLeft = computed(() => (contextMenuTab.value ? hasClosableLeft(contextMenuTab.value.id) : false));
const canCloseContextMenuRight = computed(() => (contextMenuTab.value ? hasClosableRight(contextMenuTab.value.id) : false));
const canCloseContextMenuAll = computed(() => tabsStore.tabs.some((tab: WorkspaceTabRecord) => !tab.fixed));

function getTabTitle(tab: WorkspaceTabRecord): string {
  return t(tab.titleKey || '', tab.titleDefault || tab.title || t('tabs.page', '页面'));
}

function isActive(tabId: string): boolean {
  return tabsStore.activeId === tabId;
}

function setTabRef(tabId: string, element: HTMLElement | null): void {
  if (element) {
    tabItemRefs.set(tabId, element);
    return;
  }
  tabItemRefs.delete(tabId);
}

async function scrollTabIntoView(tabId: string): Promise<void> {
  await nextTick();
  const element = tabItemRefs.get(tabId);
  element?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' });
}

function clampMenuPosition(x: number, y: number): { x: number; y: number } {
  const menuWidth = 196;
  const menuHeight = 240;
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;
  return {
    x: Math.max(8, Math.min(x, viewportWidth - menuWidth)),
    y: Math.max(8, Math.min(y, viewportHeight - menuHeight)),
  };
}

function openContextMenu(tabId: string, x: number, y: number): void {
  const position = clampMenuPosition(x, y);
  contextMenuTabId.value = tabId;
  contextMenuX.value = position.x;
  contextMenuY.value = position.y;
  contextMenuVisible.value = true;
}

function closeContextMenu(): void {
  contextMenuVisible.value = false;
}

function onTabContextMenu(event: MouseEvent, tabId: string): void {
  event.preventDefault();
  event.stopPropagation();
  openContextMenu(tabId, event.clientX, event.clientY);
}

function onTabMenuButtonClick(event: MouseEvent, tabId: string): void {
  event.stopPropagation();
  openContextMenu(tabId, event.clientX, event.clientY);
}

async function onTabClick(tabId: string): Promise<void> {
  closeContextMenu();
  const target = tabsStore.tabs.find((tab: WorkspaceTabRecord) => tab.id === tabId);
  if (!target) {
    return;
  }
  if (router.currentRoute.value.fullPath === target.routeFullPath) {
    tabsStore.setActiveTab(tabId);
    return;
  }
  tabsStore.setActiveTab(tabId);
  await router.push(target.routeFullPath);
}

async function onTabClose(tabId: string): Promise<void> {
  closeContextMenu();
  const closedTab = tabsStore.closeTab(tabId);
  if (!closedTab) {
    ElMessage.warning(t('tabs.not_closable', '该标签页不可关闭'));
    return;
  }

  const nextTab = tabsStore.activeTab;
  if (nextTab && router.currentRoute.value.fullPath !== nextTab.routeFullPath) {
    await router.push(nextTab.routeFullPath);
  }
}

async function onRefreshTab(tabId: string): Promise<void> {
  const target = tabsStore.tabs.find((tab: WorkspaceTabRecord) => tab.id === tabId);
  if (!target) {
    closeContextMenu();
    return;
  }

  tabsStore.setActiveTab(tabId);
  if (router.currentRoute.value.fullPath !== target.routeFullPath) {
    await router.push(target.routeFullPath);
  }

  tabsStore.refreshTab(tabId);
  closeContextMenu();
  window.location.reload();
}

async function syncRouteAfterMutation(previousFullPath: string): Promise<void> {
  const activeTab = tabsStore.activeTab;
  if (activeTab && router.currentRoute.value.fullPath !== activeTab.routeFullPath) {
    await router.push(activeTab.routeFullPath);
    return;
  }

  if (!activeTab && previousFullPath !== '/dashboard' && router.currentRoute.value.path !== '/dashboard') {
    await router.push('/dashboard');
  }
}

async function onTabCommand(tabId: string, command: TabCommand): Promise<void> {
  closeContextMenu();
  const previousFullPath = router.currentRoute.value.fullPath;

  if (command === 'refresh') {
    await onRefreshTab(tabId);
    return;
  }

  if (command === 'close') {
    await onTabClose(tabId);
    return;
  }

  if (command === 'close-others') {
    tabsStore.closeOthers(tabId);
    await syncRouteAfterMutation(previousFullPath);
    return;
  }

  if (command === 'close-left') {
    tabsStore.closeTabsToLeft(tabId);
    await syncRouteAfterMutation(previousFullPath);
    return;
  }

  if (command === 'close-right') {
    tabsStore.closeTabsToRight(tabId);
    await syncRouteAfterMutation(previousFullPath);
    return;
  }

  if (command === 'close-all') {
    tabsStore.closeAll();
    await syncRouteAfterMutation(previousFullPath);
  }
}

function hasClosableLeft(tabId: string): boolean {
  const index = tabsStore.tabs.findIndex((tab) => tab.id === tabId);
  if (index <= 0) {
    return false;
  }
  return tabsStore.tabs.slice(0, index).some((tab) => !tab.fixed);
}

function hasClosableRight(tabId: string): boolean {
  const index = tabsStore.tabs.findIndex((tab) => tab.id === tabId);
  if (index < 0 || index >= tabsStore.tabs.length - 1) {
    return false;
  }
  return tabsStore.tabs.slice(index + 1).some((tab) => !tab.fixed);
}

function hasClosableOthers(tabId: string): boolean {
  return tabsStore.tabs.some((tab) => !tab.fixed && tab.id !== tabId);
}

function syncContextMenuTab(): void {
  if (!contextMenuTabId.value) {
    return;
  }
  if (!tabsStore.tabs.some((tab) => tab.id === contextMenuTabId.value)) {
    closeContextMenu();
  }
}

function onGlobalInteraction(): void {
  closeContextMenu();
}

function onGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    closeContextMenu();
  }
}

watch(
  () => tabsStore.activeId,
  (tabId) => {
    if (tabId) {
      void scrollTabIntoView(tabId);
    }
    closeContextMenu();
  },
  { immediate: true },
);

watch(
  () => tabsStore.tabs.map((tab) => tab.id).join('|'),
  () => {
    syncContextMenuTab();
  },
);

onMounted(() => {
  window.addEventListener('click', onGlobalInteraction);
  window.addEventListener('scroll', onGlobalInteraction, true);
  window.addEventListener('resize', onGlobalInteraction);
  window.addEventListener('keydown', onGlobalKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener('click', onGlobalInteraction);
  window.removeEventListener('scroll', onGlobalInteraction, true);
  window.removeEventListener('resize', onGlobalInteraction);
  window.removeEventListener('keydown', onGlobalKeydown);
});
</script>

<template>
  <div class="tabs-bar">
    <div class="tabs-bar__scroll">
      <div class="tabs-bar__list" role="tablist" :aria-label="t('tabs.aria', '已打开页面')">
        <div
          v-for="tab in tabsStore.tabs"
          :key="tab.id"
          :ref="(element) => setTabRef(tab.id, element as HTMLElement | null)"
          class="tabs-bar__item"
          :class="{ 'is-active': isActive(tab.id), 'is-fixed': tab.fixed }"
          role="tab"
          :aria-selected="isActive(tab.id)"
          :title="getTabTitle(tab)"
          @contextmenu="onTabContextMenu($event, tab.id)"
          @click="void onTabClick(tab.id)"
        >
          <span class="tabs-bar__title">{{ getTabTitle(tab) }}</span>
          <div class="tabs-bar__actions">
            <button class="tabs-bar__more" type="button" :aria-label="t('tabs.more', '更多操作 {title}', { title: getTabTitle(tab) })" @click.stop="onTabMenuButtonClick($event, tab.id)">
              <el-icon><MoreFilled /></el-icon>
            </button>

            <button
              v-if="tab.closable"
              class="tabs-bar__close"
              type="button"
              :aria-label="t('tabs.close', '关闭 {title}', { title: getTabTitle(tab) })"
              @click.stop="void onTabClose(tab.id)"
            >
              <el-icon><Close /></el-icon>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="contextMenuVisible && contextMenuTab"
      class="tabs-bar__context-menu"
      :style="{ left: `${contextMenuX}px`, top: `${contextMenuY}px` }"
      role="menu"
      @click.stop
    >
      <button class="tabs-bar__context-menu-item" type="button" @click="void onTabCommand(contextMenuTab.id, 'refresh')">
        <el-icon><RefreshRight /></el-icon>
        <span>{{ t('common.refresh_current', '刷新当前页') }}</span>
      </button>
      <div class="tabs-bar__context-menu-divider" />
      <button class="tabs-bar__context-menu-item" type="button" :disabled="!canCloseContextMenuTab" @click="void onTabCommand(contextMenuTab.id, 'close')">
        <el-icon><Close /></el-icon>
        <span>{{ t('common.close_current', '关闭当前') }}</span>
      </button>
      <button class="tabs-bar__context-menu-item" type="button" :disabled="!canCloseContextMenuOthers" @click="void onTabCommand(contextMenuTab.id, 'close-others')">
        <span>{{ t('common.close_others', '关闭其他') }}</span>
      </button>
      <button class="tabs-bar__context-menu-item" type="button" :disabled="!canCloseContextMenuLeft" @click="void onTabCommand(contextMenuTab.id, 'close-left')">
        <span>{{ t('common.close_left', '关闭左侧') }}</span>
      </button>
      <button class="tabs-bar__context-menu-item" type="button" :disabled="!canCloseContextMenuRight" @click="void onTabCommand(contextMenuTab.id, 'close-right')">
        <span>{{ t('common.close_right', '关闭右侧') }}</span>
      </button>
      <button class="tabs-bar__context-menu-item" type="button" :disabled="!canCloseContextMenuAll" @click="void onTabCommand(contextMenuTab.id, 'close-all')">
        <span>{{ t('common.close_all', '关闭全部') }}</span>
      </button>
    </div>
  </div>
</template>
