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
    ElMessage.warning(t('plugin.validation_name', '请输入插件名称'));
    return;
  }

  const menus = form.menus
    .map((item) => normalizeMenuRow(item))
    .filter((item) => item.name !== '' || item.path !== '' || item.component !== '' || item.permission !== '');
  if (menus.some((item) => item.name === '' || item.path === '')) {
    ElMessage.warning(t('plugin.validation_menu', '请补全插件菜单名称和路径'));
    return;
  }

  const permissions = form.permissions
    .map((item) => normalizePermissionRow(item))
    .filter((item) => item.object !== '' || item.action !== '' || item.description !== '');
  if (permissions.some((item) => item.object === '' || item.action === '')) {
    ElMessage.warning(t('plugin.validation_permission', '请补全插件权限的对象和动作'));
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
      ElMessage.success(t('plugin.updated', '插件已更新'));
    } else {
      await createPlugin(payload);
      ElMessage.success(t('plugin.created', '插件已创建'));
    }
    dialogVisible.value = false;
    await loadPlugins();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: PluginItem) {
  await ElMessageBox.confirm(t('plugin.confirm_delete', '确认删除插件 {name} 吗？', { name: row.name }), t('plugin.delete_title', '删除插件'), {
    type: 'warning',
    confirmButtonText: t('plugin.delete_confirm', '删除'),
    cancelButtonText: t('plugin.delete_cancel', '取消'),
  });
  await deletePlugin(row.name);
  ElMessage.success(t('plugin.deleted', '插件已删除'));
  await loadPlugins();
}

function statusLabel(enabled: boolean): string {
  return enabled ? t('menu.status.active', '启用') : t('menu.status.inactive', '禁用');
}

function menuTypeLabel(type: string): string {
  switch (type) {
    case 'directory':
      return t('menu.type.directory', '目录');
    case 'button':
      return t('menu.type.button', '按钮');
    default:
      return t('menu.type.menu', '菜单');
  }
}

onMounted(() => {
  void loadPlugins();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable :title="t('plugin.title', '插件中心')" :description="t('plugin.description', '管理插件基础信息、插件菜单和插件权限定义。')" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadPlugins">{{ t('plugin.refresh', '刷新') }}</el-button>
        <el-button v-permission="'plugin:create'" type="primary" @click="openCreate">{{ t('common.create', '新增') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('common.search', '查询')">
            <el-input v-model="query.keyword" clearable :placeholder="t('plugin.keyword_placeholder', '插件名称 / 描述')" />
          </el-form-item>
          <el-form-item :label="t('plugin.status', '状态')">
            <el-select v-model="query.enabled" clearable :placeholder="t('plugin.all_status', '全部状态')" style="width: 180px">
              <el-option :label="t('menu.status.active', '启用')" value="true" />
              <el-option :label="t('menu.status.inactive', '禁用')" value="false" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', '查询') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', '重置') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-alert
        :title="t('plugin.info_title', '插件中心说明')"
        :description="t('plugin.info_description', '此页面用于管理插件定义、插件菜单和插件权限。插件页面本身由菜单中的 view/plugin/center/index 动态加载。')"
        type="info"
        show-icon
        :closable="false"
        class="mb-16"
      />

      <el-row :gutter="16" class="mb-16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.total_label', '插件总数')" :value="pluginCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic :title="t('plugin.enabled_label', '启用插件')" :value="enabledCount" />
          </el-card>
        </el-col>
      </el-row>

      <el-table :data="filteredRows" border row-key="name" v-loading="tableLoading">
        <el-table-column prop="name" :label="t('plugin.name', '插件名称')" min-width="160" />
        <el-table-column prop="description" :label="t('plugin.description_column', '描述')" min-width="220" show-overflow-tooltip />
        <el-table-column :label="t('plugin.status', '状态')" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.enabled ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.menus', '菜单数')" width="100">
          <template #default="{ row }">
            {{ row.menus?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.permissions', '权限数')" width="100">
          <template #default="{ row }">
            {{ row.permissions?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.created_at', '创建时间')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('plugin.actions', '操作')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click="openDetail(row)">{{ t('plugin.detail', '详情') }}</el-button>
            <el-button v-permission="'plugin:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', '编辑') }}</el-button>
            <el-button v-permission="'plugin:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', '删除') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </AdminTable>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingName ? t('plugin.edit_title', '编辑插件') : t('plugin.create_title', '新增插件')"
      :loading="dialogLoading"
      width="1180px"
      @confirm="submitForm"
    >
      <el-tabs v-model="activeTab" class="plugin-center-tabs">
        <el-tab-pane :label="t('plugin.basic_tab', '基础信息')" name="basic">
          <el-form label-width="110px" class="admin-form">
            <el-form-item :label="t('plugin.name', '插件名称')" required>
              <el-input v-model="form.name" :disabled="Boolean(editingName)" :placeholder="t('plugin.validation_name', '请输入插件名称')" />
            </el-form-item>
            <el-form-item :label="t('plugin.description_label', '插件描述')">
              <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="t('plugin.description_placeholder', '请输入插件描述')" />
            </el-form-item>
            <el-form-item :label="t('plugin.enabled_status', '启用状态')">
              <el-switch v-model="form.enabled" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.menus_tab', '插件菜单')" name="menus">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendMenuRow">{{ t('plugin.add_menu_row', '新增菜单行') }}</el-button>
          </div>

          <el-table :data="form.menus" border row-key="id" size="small">
            <el-table-column :label="t('plugin.menu_name', '名称')" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.name" :placeholder="t('plugin.menu_name_placeholder', '菜单名称')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_title_key', '标题 Key')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.titleKey" :placeholder="t('plugin.menu_title_key_placeholder', '例如 route.dashboard')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_title_default', '标题默认值')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.titleDefault" :placeholder="t('plugin.menu_title_default_placeholder', '例如 仪表盘')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_path', '路径')" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.path" :placeholder="t('plugin.menu_path_placeholder', '/plugin/xxx')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_component', '组件')" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.component" :placeholder="t('plugin.menu_component_placeholder', 'view/plugin/example/index')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_parent', '父级')" min-width="160">
              <template #default="{ row, $index }">
                <el-select v-model="row.parent_id" clearable filterable :placeholder="t('plugin.menu_parent_placeholder', '无父级')">
                  <el-option :label="t('plugin.menu_parent_placeholder', '无父级')" value="" />
                  <el-option
                    v-for="option in parentMenuOptions($index)"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_type', '类型')" width="110">
              <template #default="{ row }">
                <el-select v-model="row.type" style="width: 100%">
                  <el-option :label="t('plugin.menu_type_directory', '目录')" value="directory" />
                  <el-option :label="t('plugin.menu_type_menu', '菜单')" value="menu" />
                  <el-option :label="t('plugin.menu_type_button', '按钮')" value="button" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_permission', '权限')" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.permission" :placeholder="t('plugin.menu_permission_placeholder', 'plugin:xxx:view')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_sort', '排序')" width="90">
              <template #default="{ row }">
                <el-input-number v-model="row.sort" :min="0" :step="1" style="width: 100%" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_visible', '可见')" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.visible" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.menu_enabled', '启用')" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.actions', '操作')" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removeMenuRow($index)">{{ t('common.delete', '删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane :label="t('plugin.permissions_tab', '插件权限')" name="permissions">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendPermissionRow">{{ t('plugin.add_permission_row', '新增权限行') }}</el-button>
          </div>

          <el-table :data="form.permissions" border row-key="object" size="small">
            <el-table-column :label="t('plugin.permission_object', '对象')" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.object" :placeholder="t('plugin.permission_object_placeholder', 'plugin:example')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.permission_action', '动作')" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.action" :placeholder="t('plugin.permission_action_placeholder', 'view')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.permission_description', '描述')" min-width="260">
              <template #default="{ row }">
                <el-input v-model="row.description" :placeholder="t('plugin.permission_description_placeholder', '权限描述')" />
              </template>
            </el-table-column>
            <el-table-column :label="t('plugin.actions', '操作')" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removePermissionRow($index)">{{ t('common.delete', '删除') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </AdminFormDialog>
  </div>
</template>
