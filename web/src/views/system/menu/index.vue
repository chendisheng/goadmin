<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { useAppI18n } from '@/i18n';
import { createMenu, deleteMenu, fetchMenuTree, fetchMenus, updateMenu } from '@/api/system-menus';
import type { MenuFormState, MenuItem } from '@/types/admin';
import { flattenMenuItems, formatDateTime, menuTypeTagType, statusTagType } from '@/utils/admin';

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const parentLoading = ref(false);
const rows = ref<MenuItem[]>([]);
const total = ref(0);
const menuTree = ref<MenuItem[]>([]);
const editingId = ref('');
const { t } = useAppI18n();

const query = reactive({
  keyword: '',
  parent_id: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): MenuFormState => ({
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
});

const form = reactive<MenuFormState>(defaultForm());

function getMenuDisplayTitle(item: Pick<MenuItem, 'name' | 'titleKey' | 'titleDefault'>): string {
  return t(item.titleKey || '', item.titleDefault || item.name || t('menu.unnamed', 'Unnamed menu'));
}

const parentOptions = computed(() =>
  flattenMenuItems(menuTree.value).map((item) => ({
    label: `${item.path} - ${getMenuDisplayTitle(item)}`,
    value: item.id,
  })),
);

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadMenus() {
  tableLoading.value = true;
  try {
    const response = await fetchMenus({ ...query });
    rows.value = response.items;
    total.value = response.total;
  } finally {
    tableLoading.value = false;
  }
}

async function loadMenuTree() {
  parentLoading.value = true;
  try {
    const response = await fetchMenuTree();
    menuTree.value = response.items ?? [];
  } finally {
    parentLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: MenuItem) {
  editingId.value = row.id;
  Object.assign(form, {
    ...defaultForm(),
    parent_id: row.parent_id ?? '',
    name: row.name,
    titleKey: row.titleKey ?? '',
    titleDefault: row.titleDefault ?? '',
    path: row.path,
    component: row.component ?? '',
    icon: row.icon ?? '',
    sort: row.sort ?? 0,
    permission: row.permission ?? '',
    type: row.type || 'menu',
    visible: row.visible,
    enabled: row.enabled,
    redirect: row.redirect ?? '',
    external_url: row.external_url ?? '',
  });
  dialogVisible.value = true;
}

function typeLabel(type: string): string {
  switch (type) {
    case 'directory':
      return t('menu.type.directory', 'Directory');
    case 'button':
      return t('menu.type.button', 'Button');
    default:
      return t('menu.type.menu', 'Menu');
  }
}

function statusLabel(flag: boolean): string {
  return flag ? t('menu.status.active', 'Enabled') : t('menu.status.inactive', 'Disabled');
}

async function submitForm() {
  if (form.name.trim() === '' || form.path.trim() === '') {
    ElMessage.warning(t('menu.validate_required', 'Enter the menu name and path'));
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: MenuFormState = {
      ...form,
      parent_id: form.parent_id.trim(),
      name: form.name.trim(),
      titleKey: form.titleKey.trim(),
      titleDefault: form.titleDefault.trim(),
      path: form.path.trim(),
      component: form.component.trim(),
      icon: form.icon.trim(),
      sort: Number(form.sort) || 0,
      permission: form.permission.trim(),
      type: form.type.trim() || 'menu',
      visible: Boolean(form.visible),
      enabled: Boolean(form.enabled),
      redirect: form.redirect.trim(),
      external_url: form.external_url.trim(),
    };

    if (editingId.value) {
      await updateMenu(editingId.value, payload);
      ElMessage.success(t('menu.updated', 'Menu updated'));
    } else {
      await createMenu(payload);
      ElMessage.success(t('menu.created', 'Menu created'));
    }

    dialogVisible.value = false;
    await Promise.all([loadMenus(), loadMenuTree()]);
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: MenuItem) {
  await ElMessageBox.confirm(t('menu.confirm_delete', 'Delete menu {name}?', { name: row.name }), t('menu.delete_title', 'Delete menu'), {
    type: 'warning',
    confirmButtonText: t('menu.delete_confirm', 'Delete'),
    cancelButtonText: t('menu.delete_cancel', 'Cancel'),
  });
  await deleteMenu(row.id);
  ElMessage.success(t('menu.deleted', 'Menu deleted'));
  await Promise.all([loadMenus(), loadMenuTree()]);
}

function handleSearch() {
  query.page = 1;
  void loadMenus();
}

function handleReset() {
  query.keyword = '';
  query.parent_id = '';
  query.page = 1;
  void loadMenus();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadMenus();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadMenus();
}

onMounted(() => {
  void Promise.all([loadMenus(), loadMenuTree()]);
});
</script>

<template>
  <div class="admin-page">
    <AdminTable :title="t('menu.title', 'Menu management')" :description="t('menu.description', 'Maintain system menus, routes, and permission metadata.')" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadMenus">{{ t('menu.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'menu:create'" type="primary" @click="openCreate">{{ t('menu.create', 'New menu') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('common.search', 'Search')">
            <el-input v-model="query.keyword" clearable :placeholder="t('menu.keyword_placeholder', 'Menu name / path / permission')" />
          </el-form-item>
          <el-form-item :label="t('menu.parent', 'Parent menu')">
            <el-select v-model="query.parent_id" clearable filterable :loading="parentLoading" :placeholder="t('menu.parent_placeholder', 'All parents')" style="width: 220px">
              <el-option :label="t('menu.top_level', 'Top level')" value="" />
              <el-option v-for="menu in parentOptions" :key="menu.value" :label="menu.label" :value="menu.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column :label="t('menu.name', 'Name')" min-width="140">
          <template #default="{ row }">
            {{ getMenuDisplayTitle(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="path" :label="t('menu.path', 'Path')" min-width="160" />
        <el-table-column prop="component" :label="t('menu.component', 'Component')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('menu.type', 'Type')" width="100">
          <template #default="{ row }">
            <el-tag :type="menuTypeTagType(row.type)" effect="plain">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('menu.permission', 'Permission')" min-width="160">
          <template #default="{ row }">
            {{ row.permission || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('menu.visible', 'Visible')" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.visible ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.visible) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('menu.enabled', 'Enabled')" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.enabled ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('menu.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('menu.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'menu:update'" link type="primary" @click="openEdit(row)">{{ t('menu.edit', 'Edit') }}</el-button>
            <el-button v-permission="'menu:delete'" link type="danger" @click="removeRow(row)">{{ t('menu.delete', 'Delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <div class="admin-pagination">
          <el-pagination
            background
            layout="total, sizes, prev, pager, next, jumper"
            :total="total"
            :current-page="query.page"
            :page-size="query.page_size"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="handlePageChange"
            @size-change="handleSizeChange"
          />
        </div>
      </template>
    </AdminTable>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingId ? t('menu.edit_title', 'Edit menu') : t('menu.create_title', 'New menu')"
      :loading="dialogLoading"
      width="860px"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form admin-form--two-col">
        <el-form-item :label="t('menu.parent', 'Parent menu')">
          <el-select v-model="form.parent_id" clearable filterable :loading="parentLoading" :placeholder="t('menu.parent_placeholder', 'Select a parent menu')">
            <el-option :label="t('menu.top_level', 'Top level')" value="" />
            <el-option v-for="menu in parentOptions" :key="menu.value" :label="menu.label" :value="menu.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('menu.name', 'Menu name')" required>
          <el-input v-model="form.name" :placeholder="t('menu.name_placeholder', 'Enter the menu name')" />
        </el-form-item>
        <el-form-item :label="t('menu.title_key', 'Title key')">
          <el-input v-model="form.titleKey" :placeholder="t('menu.title_key_placeholder', 'For example, route.dashboard')" />
        </el-form-item>
        <el-form-item :label="t('menu.title_default', 'Default title')">
          <el-input v-model="form.titleDefault" :placeholder="t('menu.title_default_placeholder', 'For example, Dashboard')" />
        </el-form-item>
        <el-form-item :label="t('menu.path', 'Path')" required>
          <el-input v-model="form.path" :placeholder="t('menu.path_placeholder', 'Enter the route path')" />
        </el-form-item>
        <el-form-item :label="t('menu.component', 'Component path')">
          <el-input v-model="form.component" :placeholder="t('menu.component_placeholder', 'For example, view/system/user/index')" />
        </el-form-item>
        <el-form-item :label="t('menu.icon', 'Icon')">
          <el-input v-model="form.icon" :placeholder="t('menu.icon_placeholder', 'For example, user / setting')" />
        </el-form-item>
        <el-form-item :label="t('menu.sort', 'Sort')">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('menu.permission', 'Permission key')">
          <el-input v-model="form.permission" :placeholder="t('menu.permission_placeholder', 'For example, user:list')" />
        </el-form-item>
        <el-form-item :label="t('menu.type', 'Type')">
          <el-select v-model="form.type" style="width: 100%">
            <el-option :label="t('menu.type.directory', 'Directory')" value="directory" />
            <el-option :label="t('menu.type.menu', 'Menu')" value="menu" />
            <el-option :label="t('menu.type.button', 'Button')" value="button" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('menu.redirect', 'Redirect')">
          <el-input v-model="form.redirect" :placeholder="t('menu.redirect_placeholder', 'For example, /system/users')" />
        </el-form-item>
        <el-form-item :label="t('menu.external_url', 'External URL')">
          <el-input v-model="form.external_url" :placeholder="t('menu.external_url_placeholder', 'Fill this in for external links')" />
        </el-form-item>
        <el-form-item :label="t('menu.visible', 'Visible')">
          <el-switch v-model="form.visible" />
        </el-form-item>
        <el-form-item :label="t('menu.enabled', 'Enabled')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
