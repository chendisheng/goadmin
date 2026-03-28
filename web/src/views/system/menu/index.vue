<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
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

const query = reactive({
  keyword: '',
  parent_id: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): MenuFormState => ({
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
});

const form = reactive<MenuFormState>(defaultForm());

const parentOptions = computed(() =>
  flattenMenuItems(menuTree.value).map((item) => ({
    label: `${item.path} - ${item.name}`,
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
      return '目录';
    case 'button':
      return '按钮';
    default:
      return '菜单';
  }
}

function statusLabel(flag: boolean): string {
  return flag ? '启用' : '禁用';
}

async function submitForm() {
  if (form.name.trim() === '' || form.path.trim() === '') {
    ElMessage.warning('请输入菜单名称和路径');
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: MenuFormState = {
      ...form,
      parent_id: form.parent_id.trim(),
      name: form.name.trim(),
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
      ElMessage.success('菜单已更新');
    } else {
      await createMenu(payload);
      ElMessage.success('菜单已创建');
    }

    dialogVisible.value = false;
    await Promise.all([loadMenus(), loadMenuTree()]);
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: MenuItem) {
  await ElMessageBox.confirm(`确认删除菜单 ${row.name} 吗？`, '删除菜单', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteMenu(row.id);
  ElMessage.success('菜单已删除');
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
    <AdminTable title="菜单管理" description="维护系统菜单、路由和权限元数据。" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadMenus">刷新</el-button>
        <el-button v-permission="'menu:create'" type="primary" @click="openCreate">新增菜单</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="菜单名称 / 路径 / 权限" />
          </el-form-item>
          <el-form-item label="父级菜单">
            <el-select v-model="query.parent_id" clearable filterable :loading="parentLoading" placeholder="全部父级" style="width: 220px">
              <el-option label="顶级菜单" value="" />
              <el-option v-for="menu in parentOptions" :key="menu.value" :label="menu.label" :value="menu.value" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="path" label="路径" min-width="160" />
        <el-table-column prop="component" label="组件" min-width="180" show-overflow-tooltip />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="menuTypeTagType(row.type)" effect="plain">{{ typeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限" min-width="160">
          <template #default="{ row }">
            {{ row.permission || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="可见" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.visible ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.visible) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="90">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.enabled ? 'active' : 'inactive')" effect="plain">
              {{ statusLabel(row.enabled) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'menu:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'menu:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑菜单' : '新增菜单'"
      :loading="dialogLoading"
      width="860px"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form admin-form--two-col">
        <el-form-item label="父级菜单">
          <el-select v-model="form.parent_id" clearable filterable :loading="parentLoading" placeholder="选择父级菜单">
            <el-option label="顶级菜单" value="" />
            <el-option v-for="menu in parentOptions" :key="menu.value" :label="menu.label" :value="menu.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="菜单名称" required>
          <el-input v-model="form.name" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="路径" required>
          <el-input v-model="form.path" placeholder="请输入路由路径" />
        </el-form-item>
        <el-form-item label="组件路径">
          <el-input v-model="form.component" placeholder="例如 view/system/user/index" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="例如 user / setting" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="权限标识">
          <el-input v-model="form.permission" placeholder="例如 user:list" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="目录" value="directory" />
            <el-option label="菜单" value="menu" />
            <el-option label="按钮" value="button" />
          </el-select>
        </el-form-item>
        <el-form-item label="重定向">
          <el-input v-model="form.redirect" placeholder="例如 /system/users" />
        </el-form-item>
        <el-form-item label="外链地址">
          <el-input v-model="form.external_url" placeholder="外部链接时填写" />
        </el-form-item>
        <el-form-item label="可见">
          <el-switch v-model="form.visible" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
