<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { fetchMenuTree } from '@/api/system-menus';
import { createRole, deleteRole, fetchRoles, updateRole } from '@/api/roles';
import { useAppI18n } from '@/i18n';
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
const { t } = useAppI18n();

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

function getMenuDisplayTitle(item: Pick<MenuItem, 'name' | 'titleKey' | 'titleDefault'>): string {
  return t(item.titleKey || '', item.titleDefault || item.name || t('menu.unnamed', 'Unnamed menu'));
}

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
  return status === 'inactive' ? t('role.status.inactive', 'Disabled') : t('role.status.active', 'Enabled');
}

async function submitForm() {
  if (form.name.trim() === '' || form.code.trim() === '') {
    ElMessage.warning(t('role.validation_required', 'Enter the role name and code'));
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
      ElMessage.success(t('role.updated', 'Role updated'));
    } else {
      await createRole(payload);
      ElMessage.success(t('role.created', 'Role created'));
    }

    dialogVisible.value = false;
    await loadRoles();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: RoleItem) {
  await ElMessageBox.confirm(t('role.confirm_delete', 'Delete role {name}?', { name: row.name }), t('role.delete_title', 'Delete role'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteRole(row.id);
  ElMessage.success(t('role.deleted', 'Role deleted'));
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
    <AdminTable :title="t('role.title', 'Role management')" :description="t('role.description', 'Maintain role basics and menu bindings.')" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadRoles">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'role:create'" type="primary" @click="openCreate">{{ t('common.create', 'Create') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('role.keyword_label', 'Keyword')">
            <el-input v-model="query.keyword" clearable :placeholder="t('role.keyword_placeholder', 'Role name / code')" />
          </el-form-item>
          <el-form-item :label="t('role.status_label', 'Status')">
            <el-select v-model="query.status" clearable :placeholder="t('role.status_placeholder', 'All statuses')" style="width: 180px">
              <el-option :label="t('role.status.active', 'Enabled')" value="active" />
              <el-option :label="t('role.status.inactive', 'Disabled')" value="inactive" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="name" :label="t('role.name', 'Role name')" min-width="140" />
        <el-table-column prop="code" :label="t('role.code', 'Role code')" min-width="140" />
        <el-table-column :label="t('role.status', 'Status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" :label="t('role.remark', 'Remark')" min-width="220" show-overflow-tooltip />
        <el-table-column :label="t('role.menu_count', 'Menu count')" width="110">
          <template #default="{ row }">
            {{ row.menu_ids?.length ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column :label="t('role.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('role.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'role:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'role:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      :title="editingId ? t('role.edit_title', 'Edit role') : t('role.create_title', 'New role')"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('role.name', 'Role name')" required>
          <el-input v-model="form.name" :placeholder="t('role.name_placeholder', 'Enter the role name')" />
        </el-form-item>
        <el-form-item :label="t('role.code', 'Role code')" required>
          <el-input v-model="form.code" :placeholder="t('role.code_placeholder', 'Enter the role code')" />
        </el-form-item>
        <el-form-item :label="t('role.status', 'Status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('role.status.active', 'Enabled')" value="active" />
            <el-option :label="t('role.status.inactive', 'Disabled')" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('role.remark', 'Remark')">
          <el-input v-model="form.remark" type="textarea" :rows="3" :placeholder="t('role.remark_placeholder', 'Enter a remark')" />
        </el-form-item>
        <el-form-item :label="t('role.menu_permissions', 'Menu permissions')">
          <el-select
            v-model="form.menu_ids"
            multiple
            clearable
            filterable
            :loading="menuLoading"
            :placeholder="t('role.menu_permissions_placeholder', 'Select the menus this role can access')"
          >
            <el-option
              v-for="menu in menuOptions"
              :key="menu.id"
              :label="`${getMenuDisplayTitle(menu)} (${menu.path})`"
              :value="menu.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
