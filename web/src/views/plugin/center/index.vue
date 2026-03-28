<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
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
      label: `${menu.name || menu.id} (${menu.id})`,
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
    ElMessage.warning('请输入插件名称');
    return;
  }

  const menus = form.menus
    .map((item) => normalizeMenuRow(item))
    .filter((item) => item.name !== '' || item.path !== '' || item.component !== '' || item.permission !== '');
  if (menus.some((item) => item.name === '' || item.path === '')) {
    ElMessage.warning('请补全插件菜单名称和路径');
    return;
  }

  const permissions = form.permissions
    .map((item) => normalizePermissionRow(item))
    .filter((item) => item.object !== '' || item.action !== '' || item.description !== '');
  if (permissions.some((item) => item.object === '' || item.action === '')) {
    ElMessage.warning('请补全插件权限的对象和动作');
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
      ElMessage.success('插件已更新');
    } else {
      await createPlugin(payload);
      ElMessage.success('插件已创建');
    }
    dialogVisible.value = false;
    await loadPlugins();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: PluginItem) {
  await ElMessageBox.confirm(`确认删除插件 ${row.name} 吗？`, '删除插件', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deletePlugin(row.name);
  ElMessage.success('插件已删除');
  await loadPlugins();
}

function statusLabel(enabled: boolean): string {
  return enabled ? '启用' : '禁用';
}

function menuTypeLabel(type: string): string {
  switch (type) {
    case 'directory':
      return '目录';
    case 'button':
      return '按钮';
    default:
      return '菜单';
  }
}

onMounted(() => {
  void loadPlugins();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable title="插件中心" description="管理插件基础信息、插件菜单和插件权限定义。" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadPlugins">刷新</el-button>
        <el-button v-permission="'plugin:create'" type="primary" @click="openCreate">新增插件</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="插件名称 / 描述" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="query.enabled" clearable placeholder="全部状态" style="width: 180px">
              <el-option label="启用" value="true" />
              <el-option label="禁用" value="false" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-alert
        title="插件中心说明"
        description="此页面用于管理插件定义、插件菜单和插件权限。插件页面本身由菜单中的 view/plugin/center/index 动态加载。"
        type="info"
        show-icon
        :closable="false"
        class="mb-16"
      />

      <el-row :gutter="16" class="mb-16">
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic title="插件总数" :value="pluginCount" />
          </el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never">
            <el-statistic title="启用插件" :value="enabledCount" />
          </el-card>
        </el-col>
      </el-row>

      <el-table :data="filteredRows" border row-key="name" v-loading="tableLoading">
        <el-table-column prop="name" label="插件名称" min-width="160" />
        <el-table-column prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.enabled ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="菜单数" width="100">
          <template #default="{ row }">
            {{ row.menus?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="权限数" width="100">
          <template #default="{ row }">
            {{ row.permissions?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="success" @click="openDetail(row)">详情</el-button>
            <el-button v-permission="'plugin:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'plugin:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </AdminTable>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingName ? '编辑插件' : '新增插件'"
      :loading="dialogLoading"
      width="1180px"
      @confirm="submitForm"
    >
      <el-tabs v-model="activeTab" class="plugin-center-tabs">
        <el-tab-pane label="基础信息" name="basic">
          <el-form label-width="110px" class="admin-form">
            <el-form-item label="插件名称" required>
              <el-input v-model="form.name" :disabled="Boolean(editingName)" placeholder="请输入插件名称" />
            </el-form-item>
            <el-form-item label="插件描述">
              <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入插件描述" />
            </el-form-item>
            <el-form-item label="启用状态">
              <el-switch v-model="form.enabled" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="插件菜单" name="menus">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendMenuRow">新增菜单行</el-button>
          </div>

          <el-table :data="form.menus" border row-key="id" size="small">
            <el-table-column label="名称" min-width="150">
              <template #default="{ row }">
                <el-input v-model="row.name" placeholder="菜单名称" />
              </template>
            </el-table-column>
            <el-table-column label="路径" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.path" placeholder="/plugin/xxx" />
              </template>
            </el-table-column>
            <el-table-column label="组件" min-width="170">
              <template #default="{ row }">
                <el-input v-model="row.component" placeholder="view/plugin/example/index" />
              </template>
            </el-table-column>
            <el-table-column label="父级" min-width="160">
              <template #default="{ row, $index }">
                <el-select v-model="row.parent_id" clearable filterable placeholder="无父级">
                  <el-option label="无父级" value="" />
                  <el-option
                    v-for="option in parentMenuOptions($index)"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="类型" width="110">
              <template #default="{ row }">
                <el-select v-model="row.type" style="width: 100%">
                  <el-option label="目录" value="directory" />
                  <el-option label="菜单" value="menu" />
                  <el-option label="按钮" value="button" />
                </el-select>
              </template>
            </el-table-column>
            <el-table-column label="权限" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.permission" placeholder="plugin:xxx:view" />
              </template>
            </el-table-column>
            <el-table-column label="排序" width="90">
              <template #default="{ row }">
                <el-input-number v-model="row.sort" :min="0" :step="1" style="width: 100%" />
              </template>
            </el-table-column>
            <el-table-column label="可见" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.visible" />
              </template>
            </el-table-column>
            <el-table-column label="启用" width="90">
              <template #default="{ row }">
                <el-switch v-model="row.enabled" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removeMenuRow($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="插件权限" name="permissions">
          <div class="admin-table__actions mb-12">
            <el-button type="primary" plain @click="appendPermissionRow">新增权限行</el-button>
          </div>

          <el-table :data="form.permissions" border row-key="object" size="small">
            <el-table-column label="对象" min-width="180">
              <template #default="{ row }">
                <el-input v-model="row.object" placeholder="plugin:example" />
              </template>
            </el-table-column>
            <el-table-column label="动作" min-width="140">
              <template #default="{ row }">
                <el-input v-model="row.action" placeholder="view" />
              </template>
            </el-table-column>
            <el-table-column label="描述" min-width="260">
              <template #default="{ row }">
                <el-input v-model="row.description" placeholder="权限描述" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ $index }">
                <el-button link type="danger" @click="removePermissionRow($index)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </AdminFormDialog>
  </div>
</template>
