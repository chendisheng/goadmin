<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchPlugin, updatePlugin } from '@/api/plugins';
import { useAppI18n } from '@/i18n';
import PluginMenuTreeEditor from '@/views/plugin/center/components/PluginMenuTreeEditor.vue';
import type { PluginFormState, PluginItem, PluginPermission } from '@/types/plugin';
import {
  buildPluginPermissionDiffRows,
  buildPluginPermissionOrphans,
  clonePluginMenuTree,
  createPluginMenuNode,
  createPluginPermissionNode,
  flattenPluginMenus,
  generatePluginPermissions,
  generatePluginPermissionsFromTemplate,
  groupPluginPermissionPresets,
  mergePluginPermissions,
  movePluginMenuNode,
  normalizePluginMenuTree,
  pluginPermissionTemplates,
  readPluginPermissionPresets,
  removePluginPermissionPreset,
  savePluginPermissionPreset,
  type PluginPermissionPreset,
} from '@/utils/plugin';
import { formatDateTime } from '@/utils/admin';

const route = useRoute();
const router = useRouter();
const { t } = useAppI18n();

const loading = ref(false);
const saving = ref(false);
const activeTab = ref<'overview' | 'menus' | 'permissions'>('overview');
const selectedActions = ref<string[]>(['view', 'create', 'update', 'delete']);
const selectedTemplateKey = ref('crud');
const presetName = ref('');
const presetSearchQuery = ref('');
const diffFilter = ref<'all' | 'missing' | 'covered'>('all');
const presets = ref<PluginPermissionPreset[]>(readPluginPermissionPresets());
const sortNotice = ref('');
const plugin = ref<PluginItem | null>(null);

const actionOptions = computed(() => [
  { label: t('plugin.action.view', 'View'), value: 'view' },
  { label: t('plugin.action.create', 'Create'), value: 'create' },
  { label: t('plugin.action.edit', 'Edit'), value: 'update' },
  { label: t('plugin.action.delete', 'Delete'), value: 'delete' },
]);

const permissionTemplateOptions = computed(() =>
  pluginPermissionTemplates.map((template) => ({
    ...template,
    label: t(`plugin.template.${template.key}.label`, template.label),
    description: t(`plugin.template.${template.key}.description`, template.description),
  })),
);

function defaultForm(): PluginFormState {
  return {
    name: '',
    description: '',
    enabled: true,
    menus: [],
    permissions: [],
  };
}

const form = reactive<PluginFormState>(defaultForm());

const pluginName = computed(() => String(route.params.name || '').trim());
const pageTitle = computed(() => plugin.value?.name || pluginName.value || t('plugin.detail_title', 'Plugin details'));
const menuCount = computed(() => flattenPluginMenus(form.menus).length);
const permissionCount = computed(() => form.permissions.length);
const generatedPermissions = computed(() => generatePluginPermissions(form.name || pluginName.value, form.menus, selectedActions.value));
const generatedTemplatePermissions = computed(() => generatePluginPermissionsFromTemplate(form.name || pluginName.value, form.menus, selectedTemplateKey.value));
const menuPreviewRows = computed(() => flattenPluginMenus(form.menus));
const selectedTemplate = computed(() => permissionTemplateOptions.value.find((item) => item.key === selectedTemplateKey.value) ?? permissionTemplateOptions.value[1]);
const permissionDiffRows = computed(() =>
  buildPluginPermissionDiffRows(form.name || pluginName.value, form.menus, form.permissions as PluginPermission[], selectedActions.value),
);
const orphanPermissions = computed(() => buildPluginPermissionOrphans(form.name || pluginName.value, form.menus, form.permissions as PluginPermission[]));
const groupedPresets = computed(() => groupPluginPermissionPresets(presets.value));
const filteredGroupedPresets = computed(() => {
  const query = presetSearchQuery.value.trim().toLowerCase();
  if (query === '') {
    return groupedPresets.value;
  }
  return groupedPresets.value
    .map((group) => {
      const groupName = group.pluginName.toLowerCase();
      const groupMatches = groupName.includes(query);
      const presetsInGroup = groupMatches
        ? group.presets
        : group.presets.filter((preset) => {
            const haystack = [preset.name, preset.templateKey, preset.actions.join(' ')].join(' ').toLowerCase();
            return haystack.includes(query);
          });
      return {
        ...group,
        presets: presetsInGroup,
      };
    })
    .filter((group) => group.presets.length > 0);
});
const filteredPermissionDiffRows = computed(() => {
  if (diffFilter.value === 'missing') {
    return permissionDiffRows.value.filter((row) => row.missingActions.length > 0);
  }
  if (diffFilter.value === 'covered') {
    return permissionDiffRows.value.filter((row) => row.missingActions.length === 0);
  }
  return permissionDiffRows.value;
});
const coverageStats = computed(() => {
  const total = permissionDiffRows.value.length;
  const covered = permissionDiffRows.value.filter((item) => item.missingActions.length === 0).length;
  const missing = permissionDiffRows.value.filter((item) => item.missingActions.length > 0).length;
  const coverageRate = total === 0 ? 0 : Math.round((covered / total) * 100);
  return {
    total,
    covered,
    missing,
    orphan: orphanPermissions.value.length,
    coverageRate,
  };
});
const coverageLevel = computed(() => {
  if (coverageStats.value.coverageRate >= 100) {
    return 'complete';
  }
  if (coverageStats.value.coverageRate >= 75) {
    return 'high';
  }
  if (coverageStats.value.coverageRate >= 40) {
    return 'medium';
  }
  return 'low';
});
const coverageProgressColor = computed(() => {
  if (coverageLevel.value === 'complete') {
    return '#67c23a';
  }
  if (coverageLevel.value === 'high') {
    return '#409eff';
  }
  if (coverageLevel.value === 'medium') {
    return '#e6a23c';
  }
  return '#f56c6c';
});
const coverageLevelLabel = computed(() => {
  if (coverageLevel.value === 'complete') {
    return t('plugin.coverage.complete', 'Fully covered');
  }
  if (coverageLevel.value === 'high') {
    return t('plugin.coverage.high', 'High coverage');
  }
  if (coverageLevel.value === 'medium') {
    return t('plugin.coverage.medium', 'Completing coverage');
  }
  return t('plugin.coverage.low', 'Needs coverage');
});

let lastGeneratedPermissionKeys = new Set<string>();

function buildSortSummary(): string {
  if (form.menus.length === 0) {
    return t('plugin.sort.empty', 'No menus need sorting yet');
  }
  const summary = form.menus
    .slice(0, 5)
    .map((menu, index) => `${index + 1}. ${t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id || t('plugin.menu_unnamed', 'Unnamed menu'))}`)
    .join(' / ');
  return t('plugin.sort.summary', 'Auto-sorted: {summary}{more}', {
    summary,
    more: form.menus.length > 5 ? ' ...' : '',
  });
}

function ensureSeedRows() {
  if (form.menus.length === 0) {
    form.menus.push(createPluginMenuNode(form.name || pluginName.value));
  }
  if (form.permissions.length === 0) {
    form.permissions.push(createPluginPermissionNode(form.name || pluginName.value));
  }
}

function syncFromPlugin(item: PluginItem) {
  plugin.value = item;
  Object.assign(form, defaultForm(), {
    name: item.name,
    description: item.description ?? '',
    enabled: item.enabled,
    menus: clonePluginMenuTree(item.menus ?? []),
    permissions: (item.permissions ?? []).map((permission) => ({ ...permission })),
  });
  normalizePluginMenuTree(form.menus, item.name);
  ensureSeedRows();
}

async function loadPlugin() {
  if (pluginName.value === '') {
    ElMessage.warning(t('plugin.no_name', 'Plugin name cannot be empty'));
    await router.replace('/system/plugins');
    return;
  }

  loading.value = true;
  try {
    const item = await fetchPlugin(pluginName.value);
    syncFromPlugin(item);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('plugin.load_failed', 'Failed to load plugin details'));
    await router.replace('/system/plugins');
  } finally {
    loading.value = false;
  }
}

function appendPermissionRow() {
  form.permissions.push(createPluginPermissionNode(form.name || pluginName.value));
}

function removePermissionRow(index: number) {
  form.permissions.splice(index, 1);
}

function fillGeneratedPermissions(permissions = generatedPermissions.value) {
  const generated = permissions;
  if (generated.length === 0) {
    ElMessage.warning(t('plugin.generate_hint', 'Select generation actions first and make sure the menu is not empty'));
    return;
  }
  form.permissions = mergePluginPermissions(form.permissions, generated);
  lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
  ElMessage.success(t('plugin.generated_success', 'Generated {count} permissions', { count: generated.length }));
}

function clearGeneratedPermissions() {
  if (lastGeneratedPermissionKeys.size === 0) {
    return;
  }
  form.permissions = form.permissions.filter((item) => !lastGeneratedPermissionKeys.has(`${item.object}:${item.action}`));
  lastGeneratedPermissionKeys = new Set();
}

function completeDiffRow(row: (typeof permissionDiffRows.value)[number]) {
  if (row.missingActions.length === 0) {
    ElMessage.info(t('plugin.coverage.already_complete', 'This row is already covered'));
    return;
  }
  const generated = row.missingActions.map((action) => ({
    plugin: form.name || pluginName.value,
    object: row.object,
    action,
    description: `${row.menuName} ${action}`,
  }));
  form.permissions = mergePluginPermissions(form.permissions, generated);
  lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
  ElMessage.success(t('plugin.coverage.completed_count', 'Completed {count} missing permissions', { count: generated.length }));
}

function completeAllMissingPermissions() {
  const generated = permissionDiffRows.value.flatMap((row) =>
    row.missingActions.map((action) => ({
      plugin: form.name || pluginName.value,
      object: row.object,
      action,
      description: `${row.menuName} ${action}`,
    })),
  );
  if (generated.length === 0) {
    ElMessage.info(t('plugin.coverage.no_missing', 'No missing differences to complete'));
    return;
  }
  form.permissions = mergePluginPermissions(form.permissions, generated);
  lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
  ElMessage.success(t('plugin.coverage.completed_all_count', 'Completed all {count} missing permissions', { count: generated.length }));
}

function refreshPresets() {
  presets.value = readPluginPermissionPresets();
}

function saveCurrentPreset() {
  const name = presetName.value.trim();
  if (name === '') {
    ElMessage.warning(t('plugin.preset_name_required', 'Enter a preset name'));
    return;
  }
  presets.value = savePluginPermissionPreset(form.name || pluginName.value, name, selectedTemplateKey.value, selectedActions.value);
  presetName.value = '';
  ElMessage.success(t('plugin.preset_saved', 'Preset "{name}" saved', { name }));
}

function applyPreset(preset: PluginPermissionPreset) {
  selectedTemplateKey.value = preset.templateKey || 'crud';
  selectedActions.value = preset.actions.length > 0 ? preset.actions.slice() : ['view'];
  fillGeneratedPermissions(generatePluginPermissions(form.name || pluginName.value, form.menus, selectedActions.value));
}

function deletePreset(presetId: string) {
  presets.value = removePluginPermissionPreset(presetId);
}

function applyPermissionTemplate(templateKey: string) {
  const template = pluginPermissionTemplates.find((item) => item.key === templateKey);
  if (!template) {
    return;
  }
  selectedTemplateKey.value = template.key;
  selectedActions.value = template.actions.slice();
  fillGeneratedPermissions(generatePluginPermissions(form.name || pluginName.value, form.menus, template.actions));
}

function handleMoveNode(sourceId: string, targetId: string, position: 'before' | 'after' | 'inside') {
  const moved = movePluginMenuNode(form.menus, sourceId, targetId, position);
  if (!moved) {
    ElMessage.warning(t('plugin.menu_move_failed', 'The current menu cannot be moved to the target position'));
    return;
  }
  normalizePluginMenuTree(form.menus, form.name || pluginName.value);
  sortNotice.value = buildSortSummary();
  ElMessage.success(t('plugin.menu_reordered', 'Menu reordered'));
}

async function savePlugin() {
  const name = form.name.trim();
  if (name === '') {
    ElMessage.warning(t('plugin.validation_name', 'Enter the plugin name'));
    return;
  }
  if (form.menus.some((menu) => menu.name.trim() === '' || menu.path.trim() === '')) {
    ElMessage.warning(t('plugin.validation_menu', 'Complete the plugin menu name and path'));
    return;
  }
  if (form.permissions.some((permission) => permission.object.trim() === '' || permission.action.trim() === '')) {
    ElMessage.warning(t('plugin.validation_permission', 'Complete the plugin permission object and action'));
    return;
  }

  saving.value = true;
  try {
    await updatePlugin(pluginName.value, {
      name,
      description: form.description.trim(),
      enabled: Boolean(form.enabled),
      menus: form.menus,
      permissions: form.permissions,
    });
    ElMessage.success(t('plugin.save_success', 'Plugin saved'));
    await loadPlugin();
  } finally {
    saving.value = false;
  }
}

function goBack() {
  void router.push('/system/plugins');
}

watch(
  () => route.params.name,
  () => {
    void loadPlugin();
  },
  { immediate: true },
);

onMounted(() => {
  if (selectedActions.value.length === 0) {
    selectedActions.value = ['view'];
  }
  if (selectedTemplateKey.value === '') {
    selectedTemplateKey.value = 'crud';
  }
  refreshPresets();
});
</script>

<template>
  <div class="admin-page plugin-detail-page">
    <AdminTable :title="pageTitle" :description="t('plugin.detail_description', 'Plugin details, menu-tree editing, and batch permission generation.')" :loading="loading">
      <template #actions>
        <el-button @click="goBack">{{ t('common.back', 'Back') }}</el-button>
        <el-button :loading="loading" @click="loadPlugin">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'plugin:update'" type="primary" :loading="saving" @click="savePlugin">{{ t('plugin.save_plugin', 'Save plugin') }}</el-button>
      </template>

      <el-row :gutter="16" class="mb-16">
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.menu_nodes', 'Menu nodes')" :value="menuCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.permission_items', 'Permission items')" :value="permissionCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.status', 'Status')" :value="form.enabled ? 1 : 0" :formatter="(value) => (value === 1 ? t('plugin.enabled', 'Enabled') : t('plugin.disabled', 'Disabled'))" />
          </el-card>
        </el-col>
      </el-row>

      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('plugin.detail_overview_tab', 'Basic info')" name="overview">
          <el-card shadow="never">
            <el-form label-width="120px" class="admin-form admin-form--two-col">
              <el-form-item :label="t('plugin.name', 'Plugin name')">
                <el-input v-model="form.name" disabled />
              </el-form-item>
              <el-form-item :label="t('plugin.enabled_status', 'Enabled status')">
                <el-switch v-model="form.enabled" />
              </el-form-item>
              <el-form-item :label="t('plugin.description_label', 'Plugin description')" class="admin-form__full-row">
                <el-input v-model="form.description" type="textarea" :rows="4" :placeholder="t('plugin.description_placeholder', 'Enter plugin description')" />
              </el-form-item>
              <el-form-item :label="t('plugin.created_at', 'Created at')">
                <span>{{ plugin?.created_at ? formatDateTime(plugin.created_at) : '-' }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.updated_at', 'Updated at')">
                <span>{{ plugin?.updated_at ? formatDateTime(plugin.updated_at) : '-' }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.menu_tree_total', 'Menu tree total')">
                <span>{{ menuCount }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.permission_total', 'Permission total')">
                <span>{{ permissionCount }}</span>
              </el-form-item>
            </el-form>
          </el-card>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.menu_tree_editor_tab', 'Menu tree editor')" name="menus">
          <el-card shadow="never">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.menu_tree_editor_title', 'Menu tree editor') }}</span>
                <el-space wrap>
                  <el-tag effect="plain" type="success">{{ t('plugin.recursive_edit', 'Recursive editing') }}</el-tag>
                  <el-tag effect="plain" type="info">{{ t('plugin.drag_sorting', 'Drag sorting supported') }}</el-tag>
                </el-space>
              </div>
            </template>

            <el-alert
              :title="t('plugin.drag_instructions_title', 'Drag instructions')"
              :description="t('plugin.drag_instructions_description', 'Drag the menu card into the before/after/inside drop zones to adjust the tree structure. Menu changes update the permission-linked preview in real time.')"
              type="info"
              show-icon
              :closable="false"
              class="mb-12"
            />

            <el-alert
              v-if="sortNotice"
              :title="sortNotice"
              :description="t('plugin.sort_notice', 'The menu hierarchy has been renumbered automatically. Saving will submit using the latest order.')"
              type="success"
              show-icon
              :closable="false"
              class="mb-12"
            />

            <PluginMenuTreeEditor
              :menus="form.menus"
              :plugin-name="form.name || pluginName"
              @move-node="handleMoveNode"
            />
          </el-card>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.permissions_batch_tab', 'Batch permission generation')" name="permissions">
          <el-card shadow="never" class="mb-16">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.permission_template_header', 'Generate permission template') }}</span>
                <el-tag effect="plain" type="success">{{ t('plugin.generate_tag', 'Generate') }}</el-tag>
              </div>
            </template>

            <el-space wrap class="mb-12">
              <el-button
                v-for="template in permissionTemplateOptions"
                :key="template.key"
                :type="selectedTemplateKey === template.key ? 'primary' : 'default'"
                plain
                @click="applyPermissionTemplate(template.key)"
              >
                {{ template.label }}
              </el-button>
            </el-space>

            <el-row :gutter="16" class="mb-16">
              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.preset_save_title', 'Save as preset') }}</span>
                      <el-tag effect="plain" type="info">{{ t('plugin.local_storage', 'Local storage') }}</el-tag>
                    </div>
                  </template>

                  <el-form label-width="92px" class="admin-form">
                    <el-form-item :label="t('plugin.preset_name', 'Preset name')">
                      <el-input v-model="presetName" :placeholder="t('plugin.preset_name_placeholder', 'For example: plugin detail CRUD preset')" />
                    </el-form-item>
                    <el-form-item :label="t('plugin.current_template', 'Current template')">
                      <el-tag effect="plain">{{ selectedTemplate.label }}</el-tag>
                    </el-form-item>
                    <el-form-item :label="t('plugin.action_set', 'Action set')">
                      <el-space wrap>
                        <el-tag v-for="action in selectedActions" :key="action" effect="plain">{{ action }}</el-tag>
                      </el-space>
                    </el-form-item>
                    <el-form-item>
                      <el-button type="primary" @click="saveCurrentPreset">{{ t('plugin.save_current_preset', 'Save current config') }}</el-button>
                      <el-button @click="refreshPresets">{{ t('plugin.refresh_presets', 'Refresh presets') }}</el-button>
                    </el-form-item>
                  </el-form>
                </el-card>
              </el-col>

              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header plugin-detail-page__preset-header">
                      <div class="page-card__header">
                        <span>{{ t('plugin.existing_presets', 'Existing presets') }}</span>
                        <el-tag effect="plain" type="success">{{ presets.length }}</el-tag>
                      </div>
                      <el-input
                        v-model="presetSearchQuery"
                        clearable
                        size="small"
                        :placeholder="t('plugin.preset_search_placeholder', 'Search plugin or preset names')"
                        class="plugin-detail-page__preset-search"
                      />
                    </div>
                  </template>

                  <el-empty v-if="presets.length === 0" :description="t('plugin.no_presets', 'No presets yet, save one template configuration first')" />
                  <el-empty v-else-if="filteredGroupedPresets.length === 0" :description="t('plugin.no_matching_presets', 'No matching presets found')" />

                  <el-collapse v-else accordion class="plugin-detail-page__preset-groups">
                    <el-collapse-item v-for="group in filteredGroupedPresets" :key="group.pluginName" :name="group.pluginName">
                      <template #title>
                        <div class="plugin-detail-page__group-title">
                          <strong>{{ group.pluginName }}</strong>
                          <el-tag effect="plain" size="small">{{ t('plugin.preset_count', '{count} presets', { count: group.presets.length }) }}</el-tag>
                        </div>
                      </template>

                      <el-space direction="vertical" fill style="width: 100%">
                        <el-card v-for="preset in group.presets" :key="preset.id" shadow="never" class="plugin-detail-page__preset-card">
                          <div class="page-card__header">
                            <div>
                              <strong>{{ preset.name }}</strong>
                              <div class="plugin-detail-page__preset-meta">
                                <el-tag effect="plain" size="small">{{ preset.templateKey }}</el-tag>
                                <span>{{ preset.actions.join(', ') || t('plugin.no_actions', 'No actions') }}</span>
                              </div>
                            </div>
                            <el-space>
                              <el-button size="small" type="primary" plain @click="applyPreset(preset)">{{ t('plugin.apply_preset', 'Apply') }}</el-button>
                              <el-button size="small" type="danger" plain @click="deletePreset(preset.id)">{{ t('common.delete', 'Delete') }}</el-button>
                            </el-space>
                          </div>
                        </el-card>
                      </el-space>
                    </el-collapse-item>
                  </el-collapse>
                </el-card>
              </el-col>
            </el-row>

            <el-card shadow="never" class="mb-16">
              <template #header>
                <div class="page-card__header">
                  <span>{{ t('plugin.coverage_preview_title', 'Menu / permission linked preview') }}</span>
                  <el-tag effect="plain" :type="coverageLevel === 'complete' ? 'success' : coverageLevel === 'high' ? 'primary' : coverageLevel === 'medium' ? 'warning' : 'danger'">{{ t('plugin.live_update', 'Live update') }}</el-tag>
                </div>
              </template>

              <div class="plugin-detail-page__coverage-visual mb-12">
                <el-progress
                  type="dashboard"
                  :percentage="coverageStats.coverageRate"
                  :color="coverageProgressColor"
                  :stroke-width="12"
                />
                <div class="plugin-detail-page__coverage-metrics">
                  <div :style="{ borderColor: coverageProgressColor }">
                    <strong>{{ coverageStats.covered }}</strong>
                    <span>{{ t('plugin.coverage.covered', 'Covered') }}</span>
                  </div>
                  <div :style="{ borderColor: coverageProgressColor }">
                    <strong>{{ coverageStats.missing }}</strong>
                    <span>{{ t('plugin.coverage.missing', 'Missing') }}</span>
                  </div>
                  <div :style="{ borderColor: coverageProgressColor }">
                    <strong>{{ coverageStats.orphan }}</strong>
                    <span>{{ t('plugin.coverage.orphan', 'Orphan permissions') }}</span>
                  </div>
                </div>
              </div>

              <el-alert
                :title="coverageLevelLabel"
                :description="t('plugin.coverage.rate_description', 'Current coverage rate: {rate}%', { rate: coverageStats.coverageRate })"
                :type="coverageLevel === 'complete' ? 'success' : coverageLevel === 'high' ? 'info' : coverageLevel === 'medium' ? 'warning' : 'error'"
                show-icon
                :closable="false"
                class="mb-12"
              />

              <el-descriptions :column="2" border size="small" class="mb-12">
                <el-descriptions-item :label="t('plugin.menu_count', 'Menu count')">{{ menuPreviewRows.length }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.template_action_count', 'Template action count')">{{ selectedActions.length }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.template_permission_count', 'Template permission count')">{{ generatedTemplatePermissions.length }}</el-descriptions-item>
                <el-descriptions-item :label="t('plugin.current_permission_count', 'Current permission count')">{{ permissionCount }}</el-descriptions-item>
              </el-descriptions>

              <el-table :data="generatedTemplatePermissions" border size="small">
                <el-table-column prop="object" :label="t('plugin.permission_object', 'Object')" min-width="220" />
                <el-table-column prop="action" :label="t('plugin.permission_action', 'Action')" width="120" />
                <el-table-column prop="description" :label="t('plugin.permission_description', 'Description')" min-width="220" />
              </el-table>
            </el-card>
      </el-card>

          <el-card shadow="never" class="mb-16">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.coverage_diff_title', 'Menu / permission diff comparison') }}</span>
                <el-space wrap>
                  <el-tag effect="plain" type="warning">{{ t('plugin.coverage.missing_items', '{count} items missing', { count: coverageStats.missing }) }}</el-tag>
                  <el-radio-group v-model="diffFilter" size="small">
                    <el-radio-button label="all">{{ t('common.all', 'All') }}</el-radio-button>
                    <el-radio-button label="missing">{{ t('plugin.coverage.missing', 'Missing') }}</el-radio-button>
                    <el-radio-button label="covered">{{ t('plugin.coverage.covered', 'Covered') }}</el-radio-button>
                  </el-radio-group>
                  <el-button v-if="coverageStats.missing > 0" type="primary" plain @click="completeAllMissingPermissions">{{ t('plugin.coverage.complete_all', 'Complete all') }}</el-button>
                </el-space>
              </div>
            </template>

            <el-descriptions :column="4" border size="small" class="mb-12">
              <el-descriptions-item :label="t('plugin.menu_total', 'Menu total')">{{ coverageStats.total }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.covered', 'Covered')">{{ coverageStats.covered }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.missing', 'Missing')">{{ coverageStats.missing }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.orphan', 'Orphan permissions')">{{ coverageStats.orphan }}</el-descriptions-item>
            </el-descriptions>

            <el-table :data="filteredPermissionDiffRows" border size="small" class="mb-16">
              <el-table-column prop="menuName" :label="t('plugin.menu', 'Menu')" min-width="180" />
              <el-table-column prop="object" :label="t('plugin.permission_object_full', 'Permission object')" min-width="240" show-overflow-tooltip />
              <el-table-column :label="t('plugin.existing_actions', 'Existing actions')" min-width="160">
                <template #default="{ row }">
                  <el-space wrap>
                    <el-tag v-for="action in row.existingActions" :key="action" effect="plain">{{ action }}</el-tag>
                    <span v-if="row.existingActions.length === 0">-</span>
                  </el-space>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.missing_actions', 'Missing actions')" min-width="160">
                <template #default="{ row }">
                  <el-space wrap>
                    <el-tag v-for="action in row.missingActions" :key="action" type="warning" effect="plain">{{ action }}</el-tag>
                    <span v-if="row.missingActions.length === 0">-</span>
                  </el-space>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.status', 'Status')" width="120">
                <template #default="{ row }">
                  <el-tag v-if="row.missingActions.length === 0" type="success" effect="plain">{{ t('plugin.coverage.covered', 'Covered') }}</el-tag>
                  <el-tag v-else type="warning" effect="plain">{{ t('plugin.coverage.missing', 'Missing') }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.actions', 'Actions')" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" :disabled="row.missingActions.length === 0" @click="completeDiffRow(row)">
                    {{ t('plugin.coverage.complete_one', 'Complete one') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-row :gutter="16">
              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.current_menu_preview', 'Current menu preview') }}</span>
                      <el-tag effect="plain">{{ t('plugin.row_count', '{count} items', { count: menuPreviewRows.length }) }}</el-tag>
                    </div>
                  </template>

                  <el-table :data="menuPreviewRows" border size="small">
                    <el-table-column prop="sort" :label="t('plugin.sort', 'Sort')" width="90" />
                    <el-table-column prop="name" :label="t('plugin.menu_name', 'Menu name')" min-width="180" />
                    <el-table-column prop="id" :label="t('plugin.menu_id', 'Menu ID')" min-width="200" />
                    <el-table-column prop="type" :label="t('plugin.menu_type', 'Type')" width="100" />
                  </el-table>
                </el-card>
              </el-col>

              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.coverage.orphan_title', 'Orphan permissions') }}</span>
                      <el-tag effect="plain" type="danger">{{ t('plugin.row_count', '{count} items', { count: orphanPermissions.length }) }}</el-tag>
                    </div>
                  </template>

                  <el-empty v-if="orphanPermissions.length === 0" :description="t('plugin.coverage.no_orphan', 'No orphan permissions')" />
                  <el-table v-else :data="orphanPermissions" border size="small">
                    <el-table-column prop="object" :label="t('plugin.permission_object', 'Object')" min-width="220" show-overflow-tooltip />
                    <el-table-column prop="action" :label="t('plugin.permission_action', 'Action')" width="120" />
                    <el-table-column prop="description" :label="t('plugin.permission_description', 'Description')" min-width="220" show-overflow-tooltip />
                  </el-table>
                </el-card>
              </el-col>
            </el-row>
          </el-card>

          <el-card shadow="never">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.permission_detail_title', 'Permission details') }}</span>
                <el-button type="primary" plain @click="appendPermissionRow">{{ t('plugin.add_permission_row', 'Add permission row') }}</el-button>
              </div>
            </template>

            <el-table :data="form.permissions" border row-key="object" size="small">
              <el-table-column :label="t('plugin.permission_object', 'Object')" min-width="220">
                <template #default="{ row }">
                  <el-input v-model="row.object" :placeholder="t('plugin.permission_object_placeholder_detail', 'plugin:example:menu-home')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.permission_action', 'Action')" min-width="140">
                <template #default="{ row }">
                  <el-input v-model="row.action" :placeholder="t('plugin.permission_action_placeholder_detail', 'view')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.permission_description', 'Description')" min-width="260">
                <template #default="{ row }">
                  <el-input v-model="row.description" :placeholder="t('plugin.permission_description_placeholder_detail', 'Permission description')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.actions', 'Actions')" width="90" fixed="right">
                <template #default="{ $index }">
                  <el-button link type="danger" @click="removePermissionRow($index)">{{ t('common.delete', 'Delete') }}</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </AdminTable>
  </div>
</template>

<style scoped>
.plugin-detail-page__template-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 16px;
}

.plugin-detail-page__coverage-visual {
  display: flex;
  gap: 24px;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  padding: 12px 0 4px;
}

.plugin-detail-page__coverage-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(80px, 1fr));
  gap: 12px;
  min-width: 280px;
}

.plugin-detail-page__coverage-metrics > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
  padding: 12px 10px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
}

.plugin-detail-page__coverage-metrics strong {
  font-size: 20px;
  line-height: 1;
}

.plugin-detail-page__coverage-metrics span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.plugin-detail-page__preset-groups :deep(.el-collapse-item__header) {
  align-items: center;
}

.plugin-detail-page__group-title {
  display: flex;
  gap: 8px;
  align-items: center;
}

.plugin-detail-page__preset-card {
  border-style: dashed;
}

.plugin-detail-page__preset-meta {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 8px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
