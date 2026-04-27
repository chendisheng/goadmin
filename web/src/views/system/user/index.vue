<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchRoles } from '@/api/roles';
import { createUser, deleteUser, fetchUsers, updateUser } from '@/api/users';
import { useAppI18n } from '@/i18n';
import { useSessionStore } from '@/store/session';
import type { RoleItem, UserFormState, UserItem } from '@/types/admin';
import { formatDateTime, statusTagType } from '@/utils/admin';

const sessionStore = useSessionStore();
const { t } = useAppI18n();
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
  return status === 'inactive' ? t('user.status_inactive', 'Disabled') : t('user.status_active', 'Enabled');
}

async function submitForm() {
  if (form.username.trim() === '') {
    ElMessage.warning(t('user.username_required', 'Enter username'));
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
      ElMessage.success(t('user.updated', 'User updated'));
    } else {
      await createUser(payload);
      ElMessage.success(t('user.created', 'User created'));
    }

    dialogVisible.value = false;
    await loadUsers();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: UserItem) {
  await ElMessageBox.confirm(t('user.confirm_delete', 'Delete user {name}?', { name: row.username }), t('user.delete_title', 'Delete user'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteUser(row.id);
  ElMessage.success(t('user.deleted', 'User deleted'));
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
      :title="t('user.title', 'User management')"
      :description="t('user.description', 'Maintain users, role bindings, and basic profile data.')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadUsers">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'user:create'" type="primary" @click="openCreate">{{ t('common.create', 'Create') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('user.keyword_label', 'Keyword')">
            <el-input v-model="query.keyword" clearable :placeholder="t('user.keyword_placeholder', 'Username / display name / email')" />
          </el-form-item>
          <el-form-item :label="t('user.status_label', 'Status')">
            <el-select v-model="query.status" clearable :placeholder="t('user.status_placeholder', 'All statuses')" style="width: 180px">
              <el-option :label="t('user.status_active', 'Enabled')" value="active" />
              <el-option :label="t('user.status_inactive', 'Disabled')" value="inactive" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="username" :label="t('user.username', 'Username')" min-width="140" />
        <el-table-column prop="display_name" :label="t('user.display_name', 'Display name')" min-width="140" />
        <el-table-column prop="mobile" :label="t('user.mobile', 'Mobile')" min-width="140" />
        <el-table-column prop="email" :label="t('user.email', 'Email')" min-width="200" />
        <el-table-column :label="t('user.status', 'Status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('user.role', 'Roles')" min-width="220">
          <template #default="{ row }">
            <el-space wrap>
              <el-tag v-for="code in row.role_codes || []" :key="code" effect="plain">
                {{ resolveRoleLabel(code) }}
              </el-tag>
              <span v-if="!row.role_codes || row.role_codes.length === 0">-</span>
            </el-space>
          </template>
        </el-table-column>
        <el-table-column :label="t('user.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'user:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'user:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      :title="editingId ? t('user.edit_title', 'Edit user') : t('user.create_title', 'New user')"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('user.username', 'Username')" required>
          <el-input v-model="form.username" :placeholder="t('user.username_placeholder', 'Enter username')" />
        </el-form-item>
        <el-form-item :label="t('user.display_name', 'Display name')">
          <el-input v-model="form.display_name" :placeholder="t('user.display_name_placeholder', 'Enter display name')" />
        </el-form-item>
        <el-form-item :label="t('user.mobile', 'Mobile')">
          <el-input v-model="form.mobile" :placeholder="t('user.mobile_placeholder', 'Enter mobile number')" />
        </el-form-item>
        <el-form-item :label="t('user.email', 'Email')">
          <el-input v-model="form.email" :placeholder="t('user.email_placeholder', 'Enter email')" />
        </el-form-item>
        <el-form-item :label="t('user.status', 'Status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('user.status_active', 'Enabled')" value="active" />
            <el-option :label="t('user.status_inactive', 'Disabled')" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('user.role', 'Roles')">
          <el-select v-model="form.role_codes" multiple clearable filterable :loading="roleLoading" :placeholder="t('user.role_placeholder', 'Select roles')">
            <el-option
              v-for="role in roleOptions"
              :key="role.id"
              :label="`${role.name} (${role.code})`"
              :value="role.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('user.password_hash', 'Password hash')">
          <el-input v-model="form.password_hash" type="password" show-password :placeholder="t('user.password_hash_placeholder', 'Optional, leave blank to keep unchanged')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
