<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchMenuTree } from '@/api/system-menus';
import { createRole, deleteRole, fetchRoles, updateRole } from '@/api/roles';
import type { MenuItem, RoleFormState, RoleItem } from '@/types/admin';
import { flattenMenuItems, formatDateTime, statusTagType } from '@/utils/admin';

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const menuLoading = ref(false);
const rows = ref<RoleItem[]>([]);
const total = ref(0);
const menuTree = ref<MenuItem[]>([]);
const editingId = ref('');

const query = reactive({
  keyword: '',
  status: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): RoleFormState => ({
  tenant_id: '',
  name: '',
  code: '',
  status: 'active',
  remark: '',
  menu_ids: [],
});

const form = reactive<RoleFormState>(defaultForm());

const menuOptions = computed(() => flattenMenuItems(menuTree.value));

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadRoles() {
  tableLoading.value = true;
  try {
    const response = await fetchRoles({ ...query });
    rows.value = response.items;
    total.value = response.total;
  } finally {
    tableLoading.value = false;
  }
}

async function loadMenuTree() {
  menuLoading.value = true;
  try {
    const response = await fetchMenuTree();
    menuTree.value = response.items ?? [];
  } finally {
    menuLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: RoleItem) {
  editingId.value = row.id;
  Object.assign(form, {
    ...defaultForm(),
    tenant_id: row.tenant_id ?? '',
    name: row.name,
    code: row.code,
    status: row.status || 'active',
    remark: row.remark ?? '',
    menu_ids: [...(row.menu_ids ?? [])],
  });
  dialogVisible.value = true;
}

function statusLabel(status: string): string {
  return status === 'inactive' ? '禁用' : '启用';
}

async function submitForm() {
  if (form.name.trim() === '' || form.code.trim() === '') {
    ElMessage.warning('请输入角色名称和编码');
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: RoleFormState = {
      ...form,
      tenant_id: form.tenant_id.trim(),
      name: form.name.trim(),
      code: form.code.trim(),
      status: form.status.trim() || 'active',
      remark: form.remark.trim(),
      menu_ids: [...form.menu_ids],
    };

    if (editingId.value) {
      await updateRole(editingId.value, payload);
      ElMessage.success('角色已更新');
    } else {
      await createRole(payload);
      ElMessage.success('角色已创建');
    }

    dialogVisible.value = false;
    await loadRoles();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: RoleItem) {
  await ElMessageBox.confirm(`确认删除角色 ${row.name} 吗？`, '删除角色', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteRole(row.id);
  ElMessage.success('角色已删除');
  await loadRoles();
}

function handleSearch() {
  query.page = 1;
  void loadRoles();
}

function handleReset() {
  query.keyword = '';
  query.status = '';
  query.page = 1;
  void loadRoles();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadRoles();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadRoles();
}

onMounted(() => {
  void Promise.all([loadRoles(), loadMenuTree()]);
});
</script>

<template>
  <div class="admin-page">
    <AdminTable title="角色管理" description="维护角色基础信息和角色绑定菜单。" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadRoles">刷新</el-button>
        <el-button v-permission="'role:create'" type="primary" @click="openCreate">新增角色</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="角色名称 / 编码" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="query.status" clearable placeholder="全部状态" style="width: 180px">
              <el-option label="启用" value="active" />
              <el-option label="禁用" value="inactive" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="name" label="角色名称" min-width="140" />
        <el-table-column prop="code" label="角色编码" min-width="140" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="220" show-overflow-tooltip />
        <el-table-column label="菜单数量" width="110">
          <template #default="{ row }">
            {{ row.menu_ids?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'role:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'role:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑角色' : '新增角色'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="角色名称" required>
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色编码" required>
          <el-input v-model="form.code" placeholder="请输入角色编码" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
        <el-form-item label="菜单权限">
          <el-select
            v-model="form.menu_ids"
            multiple
            clearable
            filterable
            :loading="menuLoading"
            placeholder="选择角色可访问的菜单"
          >
            <el-option
              v-for="menu in menuOptions"
              :key="menu.id"
              :label="`${menu.name} (${menu.path})`"
              :value="menu.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
