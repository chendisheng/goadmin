<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchRoles } from '@/api/roles';
import { createUser, deleteUser, fetchUsers, updateUser } from '@/api/users';
import { useSessionStore } from '@/store/session';
import type { RoleItem, UserFormState, UserItem } from '@/types/admin';
import { formatDateTime, statusTagType } from '@/utils/admin';

const sessionStore = useSessionStore();
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const roleLoading = ref(false);
const total = ref(0);
const rows = ref<UserItem[]>([]);
const roleOptions = ref<RoleItem[]>([]);
const editingId = ref('');

const query = reactive({
  keyword: '',
  status: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): UserFormState => ({
  tenant_id: sessionStore.currentUser?.tenant_id ?? '',
  username: '',
  display_name: '',
  mobile: '',
  email: '',
  status: 'active',
  role_codes: [],
  password_hash: '',
});

const form = reactive<UserFormState>(defaultForm());

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadUsers() {
  tableLoading.value = true;
  try {
    const response = await fetchUsers({ ...query });
    rows.value = response.items;
    total.value = response.total;
  } finally {
    tableLoading.value = false;
  }
}

async function loadRoles() {
  roleLoading.value = true;
  try {
    const response = await fetchRoles({ keyword: '', status: '', tenant_id: '', page: 1, page_size: 200 });
    roleOptions.value = response.items;
  } finally {
    roleLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: UserItem) {
  editingId.value = row.id;
  Object.assign(form, {
    ...defaultForm(),
    tenant_id: row.tenant_id ?? sessionStore.currentUser?.tenant_id ?? '',
    username: row.username,
    display_name: row.display_name ?? '',
    mobile: row.mobile ?? '',
    email: row.email ?? '',
    status: row.status || 'active',
    role_codes: [...(row.role_codes ?? [])],
    password_hash: '',
  });
  dialogVisible.value = true;
}

function resolveRoleLabel(code: string): string {
  const role = roleOptions.value.find((item) => item.code === code);
  return role ? `${role.name} (${role.code})` : code;
}

function statusLabel(status: string): string {
  return status === 'inactive' ? '禁用' : '启用';
}

async function submitForm() {
  if (form.username.trim() === '') {
    ElMessage.warning('请输入用户名');
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: UserFormState = {
      ...form,
      tenant_id: form.tenant_id.trim(),
      username: form.username.trim(),
      display_name: form.display_name.trim(),
      mobile: form.mobile.trim(),
      email: form.email.trim(),
      status: form.status.trim() || 'active',
      role_codes: [...form.role_codes],
      password_hash: form.password_hash.trim(),
    };

    if (editingId.value) {
      await updateUser(editingId.value, payload);
      ElMessage.success('用户已更新');
    } else {
      await createUser(payload);
      ElMessage.success('用户已创建');
    }

    dialogVisible.value = false;
    await loadUsers();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: UserItem) {
  await ElMessageBox.confirm(`确认删除用户 ${row.username} 吗？`, '删除用户', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteUser(row.id);
  ElMessage.success('用户已删除');
  await loadUsers();
}

function handleSearch() {
  query.page = 1;
  void loadUsers();
}

function handleReset() {
  query.keyword = '';
  query.status = '';
  query.page = 1;
  void loadUsers();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadUsers();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadUsers();
}

onMounted(() => {
  void Promise.all([loadUsers(), loadRoles()]);
});

</script>

<template>
  <div class="admin-page">
    <AdminTable
      title="用户管理"
      description="维护系统用户、角色绑定与基础资料。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadUsers">刷新</el-button>
        <el-button v-permission="'user:create'" type="primary" @click="openCreate">新增用户</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="用户名 / 显示名 / 邮箱" />
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
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="display_name" label="显示名称" min-width="140" />
        <el-table-column prop="mobile" label="手机号" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="200" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="220">
          <template #default="{ row }">
            <el-space wrap>
              <el-tag v-for="code in row.role_codes || []" :key="code" effect="plain">
                {{ resolveRoleLabel(code) }}
              </el-tag>
              <span v-if="!row.role_codes || row.role_codes.length === 0">-</span>
            </el-space>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'user:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'user:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑用户' : '新增用户'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="form.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="form.mobile" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width: 100%">
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role_codes" multiple clearable filterable :loading="roleLoading" placeholder="选择角色">
            <el-option
              v-for="role in roleOptions"
              :key="role.id"
              :label="`${role.name} (${role.code})`"
              :value="role.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="密码哈希">
          <el-input v-model="form.password_hash" type="password" show-password placeholder="可选，留空表示不修改" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
