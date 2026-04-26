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
  { label: t('plugin.action.view', '查看'), value: 'view' },
  { label: t('plugin.action.create', '创建'), value: 'create' },
  { label: t('plugin.action.edit', '编辑'), value: 'update' },
  { label: t('plugin.action.delete', '删除'), value: 'delete' },
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
const pageTitle = computed(() => plugin.value?.name || pluginName.value || t('plugin.detail_title', '插件详情'));
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
    return t('plugin.coverage.complete', '已完全覆盖');
  }
  if (coverageLevel.value === 'high') {
    return t('plugin.coverage.high', '覆盖率较高');
  }
  if (coverageLevel.value === 'medium') {
    return t('plugin.coverage.medium', '正在补全');
  }
  return t('plugin.coverage.low', '需要补全');
});

let lastGeneratedPermissionKeys = new Set<string>();

function buildSortSummary(): string {
  if (form.menus.length === 0) {
    return t('plugin.sort.empty', '当前没有菜单需要排序');
  }
  const summary = form.menus
    .slice(0, 5)
    .map((menu, index) => `${index + 1}. ${t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id || t('plugin.menu_unnamed', '未命名菜单'))}`)
    .join(' / ');
  return t('plugin.sort.summary', '已自动重排：{summary}{more}', {
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
    ElMessage.warning(t('plugin.no_name', '插件名称不能为空'));
    await router.replace('/system/plugins');
    return;
  }

  loading.value = true;
  try {
    const item = await fetchPlugin(pluginName.value);
    syncFromPlugin(item);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('plugin.load_failed', '加载插件详情失败'));
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
    ElMessage.warning(t('plugin.generate_hint', '请先选择生成动作，并保证菜单不为空'));
    return;
  }
  form.permissions = mergePluginPermissions(form.permissions, generated);
  lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
  ElMessage.success(t('plugin.generated_success', '已生成 {count} 条权限', { count: generated.length }));
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
    ElMessage.info(t('plugin.coverage.already_complete', '当前行已覆盖，无需补全'));
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
  ElMessage.success(t('plugin.coverage.completed_count', '已补全 {count} 条缺失权限', { count: generated.length }));
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
    ElMessage.info(t('plugin.coverage.no_missing', '当前没有需要补全的差异项'));
    return;
  }
  form.permissions = mergePluginPermissions(form.permissions, generated);
  lastGeneratedPermissionKeys = new Set(generated.map((item) => `${item.object}:${item.action}`));
  ElMessage.success(t('plugin.coverage.completed_all_count', '已一键补全 {count} 条缺失权限', { count: generated.length }));
}

function refreshPresets() {
  presets.value = readPluginPermissionPresets();
}

function saveCurrentPreset() {
  const name = presetName.value.trim();
  if (name === '') {
    ElMessage.warning(t('plugin.preset_name_required', '请输入预设名称'));
    return;
  }
  presets.value = savePluginPermissionPreset(form.name || pluginName.value, name, selectedTemplateKey.value, selectedActions.value);
  presetName.value = '';
  ElMessage.success(t('plugin.preset_saved', '已保存预设「{name}」', { name }));
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
    ElMessage.warning(t('plugin.menu_move_failed', '当前菜单无法移动到目标位置'));
    return;
  }
  normalizePluginMenuTree(form.menus, form.name || pluginName.value);
  sortNotice.value = buildSortSummary();
  ElMessage.success(t('plugin.menu_reordered', '菜单已重新排序'));
}

async function savePlugin() {
  const name = form.name.trim();
  if (name === '') {
    ElMessage.warning(t('plugin.validation_name', '请输入插件名称'));
    return;
  }
  if (form.menus.some((menu) => menu.name.trim() === '' || menu.path.trim() === '')) {
    ElMessage.warning(t('plugin.validation_menu', '请补全插件菜单名称和路径'));
    return;
  }
  if (form.permissions.some((permission) => permission.object.trim() === '' || permission.action.trim() === '')) {
    ElMessage.warning(t('plugin.validation_permission', '请补全插件权限的对象和动作'));
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
    ElMessage.success(t('plugin.save_success', '插件已保存'));
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
    <AdminTable :title="pageTitle" :description="t('plugin.detail_description', '插件详情、菜单树编辑和权限批量生成。')" :loading="loading">
      <template #actions>
        <el-button @click="goBack">{{ t('common.back', '返回') }}</el-button>
        <el-button :loading="loading" @click="loadPlugin">{{ t('common.refresh', '刷新') }}</el-button>
        <el-button v-permission="'plugin:update'" type="primary" :loading="saving" @click="savePlugin">{{ t('plugin.save_plugin', '保存插件') }}</el-button>
      </template>

      <el-row :gutter="16" class="mb-16">
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.menu_nodes', '菜单节点')" :value="menuCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.permission_items', '权限条目')" :value="permissionCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="8">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.status', '状态')" :value="form.enabled ? 1 : 0" :formatter="(value) => (value === 1 ? t('plugin.enabled', '启用') : t('plugin.disabled', '禁用'))" />
          </el-card>
        </el-col>
      </el-row>

      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('plugin.detail_overview_tab', '基础信息')" name="overview">
          <el-card shadow="never">
            <el-form label-width="120px" class="admin-form admin-form--two-col">
              <el-form-item :label="t('plugin.name', '插件名称')">
                <el-input v-model="form.name" disabled />
              </el-form-item>
              <el-form-item :label="t('plugin.enabled_status', '启用状态')">
                <el-switch v-model="form.enabled" />
              </el-form-item>
              <el-form-item :label="t('plugin.description_label', '插件描述')" class="admin-form__full-row">
                <el-input v-model="form.description" type="textarea" :rows="4" :placeholder="t('plugin.description_placeholder', '请输入插件描述')" />
              </el-form-item>
              <el-form-item :label="t('plugin.created_at', '创建时间')">
                <span>{{ plugin?.created_at ? formatDateTime(plugin.created_at) : '-' }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.updated_at', '更新时间')">
                <span>{{ plugin?.updated_at ? formatDateTime(plugin.updated_at) : '-' }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.menu_tree_total', '菜单树总数')">
                <span>{{ menuCount }}</span>
              </el-form-item>
              <el-form-item :label="t('plugin.permission_total', '权限总数')">
                <span>{{ permissionCount }}</span>
              </el-form-item>
            </el-form>
          </el-card>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.menu_tree_editor_tab', '菜单树编辑器')" name="menus">
          <el-card shadow="never">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.menu_tree_editor_title', '菜单树编辑器') }}</span>
                <el-space wrap>
                  <el-tag effect="plain" type="success">{{ t('plugin.recursive_edit', '可递归编辑') }}</el-tag>
                  <el-tag effect="plain" type="info">{{ t('plugin.drag_sorting', '支持拖拽排序') }}</el-tag>
                </el-space>
              </div>
            </template>

            <el-alert
              :title="t('plugin.drag_instructions_title', '拖拽说明')"
              :description="t('plugin.drag_instructions_description', '按住菜单卡片拖拽到前后或子级投放区即可调整树形结构；菜单变化会实时影响权限联动预览。')"
              type="info"
              show-icon
              :closable="false"
              class="mb-12"
            />

            <el-alert
              v-if="sortNotice"
              :title="sortNotice"
              :description="t('plugin.sort_notice', '当前菜单层级已自动重新编号，保存时会按最新层级顺序提交。')"
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

        <el-tab-pane :label="t('plugin.permissions_batch_tab', '权限批量生成')" name="permissions">
          <el-card shadow="never" class="mb-16">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.permission_template_header', '权限模板一键生成') }}</span>
                <el-tag effect="plain" type="success">{{ t('plugin.generate_tag', '一键生成') }}</el-tag>
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
                      <span>{{ t('plugin.preset_save_title', '保存为预设') }}</span>
                      <el-tag effect="plain" type="info">{{ t('plugin.local_storage', '本地存储') }}</el-tag>
                    </div>
                  </template>

                  <el-form label-width="92px" class="admin-form">
                    <el-form-item :label="t('plugin.preset_name', '预设名称')">
                      <el-input v-model="presetName" :placeholder="t('plugin.preset_name_placeholder', '例如：插件详情 CRUD 预设')" />
                    </el-form-item>
                    <el-form-item :label="t('plugin.current_template', '当前模板')">
                      <el-tag effect="plain">{{ selectedTemplate.label }}</el-tag>
                    </el-form-item>
                    <el-form-item :label="t('plugin.action_set', '动作集合')">
                      <el-space wrap>
                        <el-tag v-for="action in selectedActions" :key="action" effect="plain">{{ action }}</el-tag>
                      </el-space>
                    </el-form-item>
                    <el-form-item>
                      <el-button type="primary" @click="saveCurrentPreset">{{ t('plugin.save_current_preset', '保存当前配置') }}</el-button>
                      <el-button @click="refreshPresets">{{ t('plugin.refresh_presets', '刷新预设') }}</el-button>
                    </el-form-item>
                  </el-form>
                </el-card>
              </el-col>

              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header plugin-detail-page__preset-header">
                      <div class="page-card__header">
                        <span>{{ t('plugin.existing_presets', '已有预设') }}</span>
                        <el-tag effect="plain" type="success">{{ presets.length }}</el-tag>
                      </div>
                      <el-input
                        v-model="presetSearchQuery"
                        clearable
                        size="small"
                        :placeholder="t('plugin.preset_search_placeholder', '搜索插件或预设名称')"
                        class="plugin-detail-page__preset-search"
                      />
                    </div>
                  </template>

                  <el-empty v-if="presets.length === 0" :description="t('plugin.no_presets', '暂无预设，先保存一个模板配置吧')" />
                  <el-empty v-else-if="filteredGroupedPresets.length === 0" :description="t('plugin.no_matching_presets', '未找到匹配的预设')" />

                  <el-collapse v-else accordion class="plugin-detail-page__preset-groups">
                    <el-collapse-item v-for="group in filteredGroupedPresets" :key="group.pluginName" :name="group.pluginName">
                      <template #title>
                        <div class="plugin-detail-page__group-title">
                          <strong>{{ group.pluginName }}</strong>
                          <el-tag effect="plain" size="small">{{ t('plugin.preset_count', '{count} 个预设', { count: group.presets.length }) }}</el-tag>
                        </div>
                      </template>

                      <el-space direction="vertical" fill style="width: 100%">
                        <el-card v-for="preset in group.presets" :key="preset.id" shadow="never" class="plugin-detail-page__preset-card">
                          <div class="page-card__header">
                            <div>
                              <strong>{{ preset.name }}</strong>
                              <div class="plugin-detail-page__preset-meta">
                                <el-tag effect="plain" size="small">{{ preset.templateKey }}</el-tag>
                                <span>{{ preset.actions.join(', ') || t('plugin.no_actions', '无动作') }}</span>
                              </div>
                            </div>
                            <el-space>
                              <el-button size="small" type="primary" plain @click="applyPreset(preset)">{{ t('plugin.apply_preset', '应用') }}</el-button>
                              <el-button size="small" type="danger" plain @click="deletePreset(preset.id)">{{ t('common.delete', '删除') }}</el-button>
                            </el-space>
                          </div>
                        </el-card>
                      </el-space>
                    </el-collapse-item>
                  </el-collapse>
                </el-card>
              </el-col>
            </el-row>

            <el-row :gutter="16" class="mb-16">
              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.template_config_title', '模板配置') }}</span>
                      <el-tag effect="plain">{{ selectedTemplate.label }}</el-tag>
                    </div>
                  </template>

                  <el-alert
                    :title="selectedTemplate.description"
                    type="info"
                    show-icon
                    :closable="false"
                    class="mb-12"
                  />

                  <el-checkbox-group v-model="selectedActions">
                    <el-checkbox v-for="option in actionOptions" :key="option.value" :label="option.value">
                      {{ option.label }}
                    </el-checkbox>
                  </el-checkbox-group>

                  <div class="plugin-detail-page__template-actions">
                    <el-button type="primary" @click="fillGeneratedPermissions()">{{ t('plugin.generate_by_actions', '按当前动作生成') }}</el-button>
                    <el-button @click="clearGeneratedPermissions">{{ t('plugin.clear_generated', '清除最近生成项') }}</el-button>
                  </div>
                </el-card>
              </el-col>

              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.coverage_preview_title', '菜单 / 权限联动预览') }}</span>
                      <el-tag effect="plain" :type="coverageLevel === 'complete' ? 'success' : coverageLevel === 'high' ? 'primary' : coverageLevel === 'medium' ? 'warning' : 'danger'">{{ t('plugin.live_update', '实时更新') }}</el-tag>
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
                        <span>{{ t('plugin.coverage.covered', '已覆盖') }}</span>
                      </div>
                      <div :style="{ borderColor: coverageProgressColor }">
                        <strong>{{ coverageStats.missing }}</strong>
                        <span>{{ t('plugin.coverage.missing', '待补全') }}</span>
                      </div>
                      <div :style="{ borderColor: coverageProgressColor }">
                        <strong>{{ coverageStats.orphan }}</strong>
                        <span>{{ t('plugin.coverage.orphan', '孤儿权限') }}</span>
                      </div>
                    </div>
                  </div>

                  <el-alert
                    :title="coverageLevelLabel"
                    :description="t('plugin.coverage.rate_description', '当前覆盖率 {rate}%', { rate: coverageStats.coverageRate })"
                    :type="coverageLevel === 'complete' ? 'success' : coverageLevel === 'high' ? 'info' : coverageLevel === 'medium' ? 'warning' : 'error'"
                    show-icon
                    :closable="false"
                    class="mb-12"
                  />

                  <el-descriptions :column="2" border size="small" class="mb-12">
                    <el-descriptions-item :label="t('plugin.menu_count', '菜单数')">{{ menuPreviewRows.length }}</el-descriptions-item>
                    <el-descriptions-item :label="t('plugin.template_action_count', '模板动作数')">{{ selectedActions.length }}</el-descriptions-item>
                    <el-descriptions-item :label="t('plugin.template_permission_count', '模板权限数')">{{ generatedTemplatePermissions.length }}</el-descriptions-item>
                    <el-descriptions-item :label="t('plugin.current_permission_count', '当前权限数')">{{ permissionCount }}</el-descriptions-item>
                  </el-descriptions>

                  <el-table :data="generatedTemplatePermissions" border size="small">
                    <el-table-column prop="object" :label="t('plugin.permission_object', '对象')" min-width="220" />
                    <el-table-column prop="action" :label="t('plugin.permission_action', '动作')" width="120" />
                    <el-table-column prop="description" :label="t('plugin.permission_description', '描述')" min-width="220" />
                  </el-table>
                </el-card>
              </el-col>
            </el-row>
          </el-card>

          <el-card shadow="never" class="mb-16">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.coverage_diff_title', '菜单 / 权限差异对比') }}</span>
                <el-space wrap>
                  <el-tag effect="plain" type="warning">{{ t('plugin.coverage.missing_items', '{count} 项待补全', { count: coverageStats.missing }) }}</el-tag>
                  <el-radio-group v-model="diffFilter" size="small">
                    <el-radio-button label="all">{{ t('common.all', '全部') }}</el-radio-button>
                    <el-radio-button label="missing">{{ t('plugin.coverage.missing', '待补全') }}</el-radio-button>
                    <el-radio-button label="covered">{{ t('plugin.coverage.covered', '已覆盖') }}</el-radio-button>
                  </el-radio-group>
                  <el-button v-if="coverageStats.missing > 0" type="primary" plain @click="completeAllMissingPermissions">{{ t('plugin.coverage.complete_all', '一键补全全部') }}</el-button>
                </el-space>
              </div>
            </template>

            <el-descriptions :column="4" border size="small" class="mb-12">
              <el-descriptions-item :label="t('plugin.menu_total', '菜单总数')">{{ coverageStats.total }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.covered', '已覆盖')">{{ coverageStats.covered }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.missing', '待补全')">{{ coverageStats.missing }}</el-descriptions-item>
              <el-descriptions-item :label="t('plugin.coverage.orphan', '孤儿权限')">{{ coverageStats.orphan }}</el-descriptions-item>
            </el-descriptions>

            <el-table :data="filteredPermissionDiffRows" border size="small" class="mb-16">
              <el-table-column prop="menuName" :label="t('plugin.menu', '菜单')" min-width="180" />
              <el-table-column prop="object" :label="t('plugin.permission_object_full', '权限对象')" min-width="240" show-overflow-tooltip />
              <el-table-column :label="t('plugin.existing_actions', '已有动作')" min-width="160">
                <template #default="{ row }">
                  <el-space wrap>
                    <el-tag v-for="action in row.existingActions" :key="action" effect="plain">{{ action }}</el-tag>
                    <span v-if="row.existingActions.length === 0">-</span>
                  </el-space>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.missing_actions', '缺失动作')" min-width="160">
                <template #default="{ row }">
                  <el-space wrap>
                    <el-tag v-for="action in row.missingActions" :key="action" type="warning" effect="plain">{{ action }}</el-tag>
                    <span v-if="row.missingActions.length === 0">-</span>
                  </el-space>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.status', '状态')" width="120">
                <template #default="{ row }">
                  <el-tag v-if="row.missingActions.length === 0" type="success" effect="plain">{{ t('plugin.coverage.covered', '已覆盖') }}</el-tag>
                  <el-tag v-else type="warning" effect="plain">{{ t('plugin.coverage.missing', '待补全') }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.actions', '操作')" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" :disabled="row.missingActions.length === 0" @click="completeDiffRow(row)">
                    {{ t('plugin.coverage.complete_one', '一键补全') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-row :gutter="16">
              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.current_menu_preview', '当前菜单预览') }}</span>
                      <el-tag effect="plain">{{ t('plugin.row_count', '{count} 条', { count: menuPreviewRows.length }) }}</el-tag>
                    </div>
                  </template>

                  <el-table :data="menuPreviewRows" border size="small">
                    <el-table-column prop="sort" :label="t('plugin.sort', '排序')" width="90" />
                    <el-table-column prop="name" :label="t('plugin.menu_name', '菜单名称')" min-width="180" />
                    <el-table-column prop="id" :label="t('plugin.menu_id', '菜单 ID')" min-width="200" />
                    <el-table-column prop="type" :label="t('plugin.menu_type', '类型')" width="100" />
                  </el-table>
                </el-card>
              </el-col>

              <el-col :xs="24" :md="12">
                <el-card shadow="never">
                  <template #header>
                    <div class="page-card__header">
                      <span>{{ t('plugin.coverage.orphan_title', '孤儿权限') }}</span>
                      <el-tag effect="plain" type="danger">{{ orphanPermissions.length }} 条</el-tag>
                    </div>
                  </template>

                  <el-empty v-if="orphanPermissions.length === 0" :description="t('plugin.coverage.no_orphan', '暂无孤儿权限')" />
                  <el-table v-else :data="orphanPermissions" border size="small">
                    <el-table-column prop="object" :label="t('plugin.permission_object', '对象')" min-width="220" show-overflow-tooltip />
                    <el-table-column prop="action" :label="t('plugin.permission_action', '动作')" width="120" />
                    <el-table-column prop="description" :label="t('plugin.permission_description', '描述')" min-width="220" show-overflow-tooltip />
                  </el-table>
                </el-card>
              </el-col>
            </el-row>
          </el-card>

          <el-card shadow="never">
            <template #header>
              <div class="page-card__header">
                <span>{{ t('plugin.permission_detail_title', '权限明细') }}</span>
                <el-button type="primary" plain @click="appendPermissionRow">{{ t('plugin.add_permission_row', '新增权限行') }}</el-button>
              </div>
            </template>

            <el-table :data="form.permissions" border row-key="object" size="small">
              <el-table-column :label="t('plugin.permission_object', '对象')" min-width="220">
                <template #default="{ row }">
                  <el-input v-model="row.object" :placeholder="t('plugin.permission_object_placeholder_detail', 'plugin:example:menu-home')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.permission_action', '动作')" min-width="140">
                <template #default="{ row }">
                  <el-input v-model="row.action" :placeholder="t('plugin.permission_action_placeholder_detail', 'view')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.permission_description', '描述')" min-width="260">
                <template #default="{ row }">
                  <el-input v-model="row.description" :placeholder="t('plugin.permission_description_placeholder_detail', '权限描述')" />
                </template>
              </el-table-column>
              <el-table-column :label="t('plugin.actions', '操作')" width="90" fixed="right">
                <template #default="{ $index }">
                  <el-button link type="danger" @click="removePermissionRow($index)">{{ t('common.delete', '删除') }}</el-button>
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
