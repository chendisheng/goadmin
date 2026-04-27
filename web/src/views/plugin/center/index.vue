<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createPlugin, deletePlugin, fetchPlugins, updatePlugin } from '@/api/plugins';
import type { PluginFormState, PluginItem, PluginMenu, PluginPermission } from '@/types/plugin';
import { formatDateTime, statusTagType } from '@/utils/admin';

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<PluginItem[]>([]);
const activeTab = ref<'basic' | 'menus' | 'permissions'>('basic');
const editingName = ref('');
const router = useRouter();
const { t } = useAppI18n();

const query = reactive({
  keyword: '',
  enabled: '',
});

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

const filteredRows = computed(() => {
  const keyword = query.keyword.trim().toLowerCase();
  return rows.value.filter((row) => {
    const matchesKeyword =
      keyword === '' ||
      [row.name, row.description ?? ''].some((value) => value.toLowerCase().includes(keyword));
    const matchesEnabled = query.enabled === '' || String(row.enabled) === query.enabled;
    return matchesKeyword && matchesEnabled;
  });
});

const pluginCount = computed(() => filteredRows.value.length);
const enabledCount = computed(() => filteredRows.value.filter((item) => item.enabled).length);

function resetForm() {
  Object.assign(form, defaultForm());
}

function createMenuRow(): PluginMenu {
  const pluginName = editingName.value.trim() || form.name.trim();
  return {
    plugin: pluginName,
    id: `menu-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    parent_id: '',
    name: '',
    titleKey: '',
    titleDefault: '',
    path: '',
    component: '',
    icon: '',
    sort: 0,
    permission: '',
    type: 'menu',
    visible: true,
    enabled: true,
    redirect: '',
    external_url: '',
    children: [],
  };
}

function createPermissionRow(): PluginPermission {
  const pluginName = editingName.value.trim() || form.name.trim();
  return {
    plugin: pluginName,
    object: '',
    action: '',
    description: '',
  };
}

function parentMenuOptions(currentIndex: number): Array<{ label: string; value: string }> {
  return form.menus
    .filter((_, index) => index !== currentIndex)
    .map((menu) => ({
      label: `${t(menu.titleKey || '', menu.titleDefault || menu.name || menu.id)} (${menu.id})`,
      value: menu.id,
    }));
}

function normalizeMenuRow(row: PluginMenu): PluginMenu {
  const pluginName = editingName.value.trim() || form.name.trim();
  return {
    ...row,
    plugin: pluginName,
    id: row.id.trim(),
    parent_id: row.parent_id?.trim() ?? '',
    name: row.name.trim(),
    titleKey: row.titleKey?.trim() ?? '',
    titleDefault: row.titleDefault?.trim() ?? '',
    path: row.path.trim(),
    component: row.component?.trim() ?? '',
    icon: row.icon?.trim() ?? '',
    sort: Number(row.sort) || 0,
    permission: row.permission?.trim() ?? '',
    type: row.type?.trim() || 'menu',
    visible: Boolean(row.visible),
    enabled: Boolean(row.enabled),
    redirect: row.redirect?.trim() ?? '',
    external_url: row.external_url?.trim() ?? '',
    children: [],
  };
}

function normalizePermissionRow(row: PluginPermission): PluginPermission {
  const pluginName = editingName.value.trim() || form.name.trim();
  return {
    plugin: pluginName,
    object: row.object.trim(),
    action: row.action.trim(),
    description: row.description.trim(),
  };
}

async function loadPlugins() {
  tableLoading.value = true;
  try {
    const response = await fetchPlugins();
    rows.value = response.items ?? [];
  } finally {
    tableLoading.value = false;
  }
}

function openDetail(row: PluginItem) {
  void router.push(`/system/plugins/${encodeURIComponent(row.name)}`);
}

function openCreate() {
  editingName.value = '';
  resetForm();
  activeTab.value = 'basic';
  dialogVisible.value = true;
}

function openEdit(row: PluginItem) {
  editingName.value = row.name;
  Object.assign(form, defaultForm(), {
    name: row.name,
    description: row.description ?? '',
    enabled: row.enabled,
    menus: (row.menus ?? []).map((menu) => ({ ...menu, children: [] })),
    permissions: (row.permissions ?? []).map((permission) => ({ ...permission })),
  });
  if (form.menus.length === 0) {
    form.menus.push(createMenuRow());
  }
  if (form.permissions.length === 0) {
    form.permissions.push(createPermissionRow());
  }
  activeTab.value = 'basic';
  dialogVisible.value = true;
}

function handleSearch() {
  void loadPlugins();
}

function handleReset() {
  query.keyword = '';
  query.enabled = '';
}

function appendMenuRow() {
  form.menus.push(createMenuRow());
}

function removeMenuRow(index: number) {
  form.menus.splice(index, 1);
}

function appendPermissionRow() {
  form.permissions.push(createPermissionRow());
}

function removePermissionRow(index: number) {
  form.permissions.splice(index, 1);
}

async function submitForm() {
  const name = form.name.trim();
  if (name === '') {
    ElMessage.warning(t('plugin.validation_name', 'Enter the plugin name'));
    return;
  }

  const menus = form.menus
    .map((item) => normalizeMenuRow(item))
    .filter((item) => item.name !== '' || item.path !== '' || item.component !== '' || item.permission !== '');
  if (menus.some((item) => item.name === '' || item.path === '')) {
    ElMessage.warning(t('plugin.validation_menu', 'Complete the plugin menu name and path'));
    return;
  }

  const permissions = form.permissions
    .map((item) => normalizePermissionRow(item))
    .filter((item) => item.object !== '' || item.action !== '' || item.description !== '');
  if (permissions.some((item) => item.object === '' || item.action === '')) {
    ElMessage.warning(t('plugin.validation_permission', 'Complete the plugin permission object and action'));
    return;
  }

  const payload: PluginFormState = {
    name,
    description: form.description.trim(),
    enabled: Boolean(form.enabled),
    menus,
    permissions,
  };

  dialogLoading.value = true;
  try {
    if (editingName.value) {
      await updatePlugin(editingName.value, payload);
      ElMessage.success(t('plugin.updated', 'Plugin updated'));
    } else {
      await createPlugin(payload);
      ElMessage.success(t('plugin.created', 'Plugin created'));
    }
    dialogVisible.value = false;
    await loadPlugins();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: PluginItem) {
  await ElMessageBox.confirm(t('plugin.confirm_delete', 'Delete plugin {name}?', { name: row.name }), t('plugin.delete_title', 'Delete plugin'), {
    type: 'warning',
    confirmButtonText: t('plugin.delete_confirm', 'Delete'),
    cancelButtonText: t('plugin.delete_cancel', 'Cancel'),
  });
  await deletePlugin(row.name);
  ElMessage.success(t('plugin.deleted', 'Plugin deleted'));
  await loadPlugins();
}

function statusLabel(enabled: boolean): string {
  return enabled ? t('menu.status.active', 'Enabled') : t('menu.status.inactive', 'Disabled');
}

function menuTypeLabel(type: string): string {
  switch (type) {
    case 'directory':
      return t('menu.type.directory', 'Directory');
    case 'button':
      return t('menu.type.button', 'Button');
    default:
      return t('menu.type.menu', 'Menu');
  }
}

onMounted(() => {
  void loadPlugins();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable :title="t('plugin.title', 'Plugin center')" :description="t('plugin.description', 'Manage plugin metadata, plugin menus, and plugin permission definitions.')" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadPlugins">{{ t('plugin.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'plugin:create'" type="primary" @click="openCreate">{{ t('common.create', 'Create') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('common.search', 'Search')">
            <el-input v-model="query.keyword" clearable :placeholder="t('plugin.keyword_placeholder', 'Plugin name / description')" />
          </el-form-item>
          <el-form-item :label="t('plugin.status', 'Status')">
            <el-select v-model="query.enabled" clearable :placeholder="t('plugin.all_status', 'All statuses')" style="width: 180px">
              <el-option :label="t('menu.status.active', 'Enabled')" value="true" />
              <el-option :label="t('menu.status.inactive', 'Disabled')" value="false" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-alert
        :title="t('plugin.info_title', 'Plugin center notes')"
        :description="t('plugin.info_description', 'This page manages plugin definitions, plugin menus, and plugin permissions. The plugin page itself is loaded dynamically from the menu entry view/plugin/center/index.')"
        type="info"
        show-icon
        :closable="false"
        class="mb-16"
      />

      <el-row :gutter="16" class="mb-16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.total_label', 'Total plugins')" :value="pluginCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.enabled_label', 'Enabled plugins')" :value="enabledCount" />
          </el-card>
        </el-col>
      </el-row>

      <el-table :data="filteredRows" border row-key="name" v-loading="tableLoading">
        <el-table-column prop="name" :label="t('plugin.name', 'Plugin name')" min-width="160" />
        <el-table-column prop="description" :label="t('plugin.description_column', 'Description')" min-width="220" show-overflow-tooltip />
        <el-table-column :label="t('plugin.status', 'Status')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.enabled ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.menus', 'Menu count')" width="100">
          <template #default="{ row }">
            {{ row.menus?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.permissions', 'Permission count')" width="100">
          <template #default="{ row }">
            {{ row.permissions?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click="openDetail(row)">{{ t('plugin.detail', 'Details') }}</el-button>
            <el-button v-permission="'plugin:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'plugin:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </AdminTable>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingName ? t('plugin.edit_title', 'Edit plugin') : t('plugin.create_title', 'New plugin')"
      :loading="dialogLoading"
      width="1180px"
      @confirm="submitForm"
    >
      <el-tabs v-model="activeTab" class="plugin-center-tabs">
        <el-tab-pane :label="t('plugin.basic_tab', 'Basic info')" name="basic">
          <el-form label-width="110px" class="admin-form">
            <el-form-item :label="t('plugin.name', 'Plugin name')" required>
              <el-input v-model="form.name" :disabled="Boolean(editingName)" :placeholder="t('plugin.validation_name', 'Enter the plugin name')" />
            </el-form-item>
            <el-form-item :label="t('plugin.description_label', 'Plugin description')">
              <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="t('plugin.description_placeholder', 'Enter plugin description')" />
            </el-form-item>
            <el-form-item :label="t('plugin.enabled_status', 'Enabled status')">
              <el-switch v-model="form.enabled" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.menus_tab', 'Plugin menus')" name="menus">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendMenuRow">{{ t('plugin.add_menu_row', 'Add menu row') }}</el-button>
          </div>

          <el-table :data="form.menus" border row-key="id" size="small">
            <el-table-column :label="t('plugin.menu_name', 'Name')" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.name" :placeholder="t('plugin.menu_name_placeholder', 'Menu name')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_title_key', 'Title key')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.titleKey" :placeholder="t('plugin.menu_title_key_placeholder', 'For example, route.dashboard')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_title_default', 'Default title')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.titleDefault" :placeholder="t('plugin.menu_title_default_placeholder', 'For example, Dashboard')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_path', 'Path')" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.path" :placeholder="t('plugin.menu_path_placeholder', '/plugin/xxx')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_component', 'Component')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.component" :placeholder="t('plugin.menu_component_placeholder', 'view/plugin/example/index')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_parent', 'Parent')" min-width="160">
              <template #default="{ row, $index }">
                <el-select v-model="row.parent_id" clearable filterable :placeholder="t('plugin.menu_parent_placeholder', 'No parent')">
                  <el-option :label="t('plugin.menu_parent_placeholder', 'No parent')" value="" />
                  <el-option
                    v-for="option in parentMenuOptions($index)"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_type', 'Type')" width="110">
              <template #default="{ row }">
                <el-select v-model="row.type" style="width: 100%">
                  <el-option :label="t('plugin.menu_type_directory', 'Directory')" value="directory" />
                  <el-option :label="t('plugin.menu_type_menu', 'Menu')" value="menu" />
                  <el-option :label="t('plugin.menu_type_button', 'Button')" value="button" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_permission', 'Permission')" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.permission" :placeholder="t('plugin.menu_permission_placeholder', 'plugin:xxx:view')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_sort', 'Sort')" width="90">
              <template #default="{ row }">
                <el-input-number v-model="row.sort" :min="0" :step="1" style="width: 100%" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_visible', 'Visible')" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.visible" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_enabled', 'Enabled')" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.actions', 'Actions')" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removeMenuRow($index)">{{ t('common.delete', 'Delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.permissions_tab', 'Plugin permissions')" name="permissions">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendPermissionRow">{{ t('plugin.add_permission_row', 'Add permission row') }}</el-button>
          </div>

          <el-table :data="form.permissions" border row-key="object" size="small">
            <el-table-column :label="t('plugin.permission_object', 'Object')" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.object" :placeholder="t('plugin.permission_object_placeholder', 'plugin:example')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.permission_action', 'Action')" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.action" :placeholder="t('plugin.permission_action_placeholder', 'view')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.permission_description', 'Description')" min-width="260">
              <template #default="{ row }">
                <el-input v-model="row.description" :placeholder="t('plugin.permission_description_placeholder', 'Permission description')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.actions', 'Actions')" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removePermissionRow($index)">{{ t('common.delete', 'Delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </AdminFormDialog>
  </div>
</template>
