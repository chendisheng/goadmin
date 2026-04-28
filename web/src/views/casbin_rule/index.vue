<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { preloadRouteNamespaces, useAppI18n } from '@/i18n';
import { createCasbinRule, deleteCasbinRule, listcasbin_rules, updateCasbinRule } from '@/api/casbin_rule';
import { formatDateTime } from '@/utils/admin';

type CasbinRuleItem = {
  id: string;
  ptype?: string;
  v0?: string;
  v1?: string;
  v2?: string;
  v3?: string;
  v4?: string;
  v5?: string;
  created_at?: string;
  updated_at?: string;
};

type CasbinRuleFormState = {
  ptype: string;
  v0: string;
  v1: string;
  v2: string;
  v3: string;
  v4: string;
  v5: string;
};

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<CasbinRuleItem[]>([]);
const total = ref(0);
const editingId = ref('');
const { t } = useAppI18n();

const query = reactive({
  keyword: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): CasbinRuleFormState => ({
  ptype: '',
  v0: '',
  v1: '',
  v2: '',
  v3: '',
  v4: '',
  v5: '',
});

const form = reactive<CasbinRuleFormState>(defaultForm());

type EnumOption = {
  value: string;
  label: string;
  color?: string;
  disabled?: boolean;
  order?: number;
};

function formatEnumLabel(value: unknown, labelMap: Record<string, string>) {
  if (Array.isArray(value)) {
    if (value.length === 0) {
      return '-';
    }
    return value.map((item) => labelMap[String(item)] ?? String(item)).join(', ');
  }
  if (value === null || value === undefined || value === '') {
    return '-';
  }
  return labelMap[String(value)] ?? String(value);
}

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadItems() {
  tableLoading.value = true;
  try {
    const response = await listcasbin_rules({ ...query });
    rows.value = response.items ?? [];
    total.value = response.total ?? 0;
  } finally {
    tableLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: CasbinRuleItem) {
  editingId.value = row.id;
  Object.assign(form, {
    ptype: row.ptype ?? '',
    v0: row.v0 ?? '',
    v1: row.v1 ?? '',
    v2: row.v2 ?? '',
    v3: row.v3 ?? '',
    v4: row.v4 ?? '',
    v5: row.v5 ?? '',
  });
  dialogVisible.value = true;
}

async function submitForm() {
  dialogLoading.value = true;
  try {
    const payload: CasbinRuleFormState = {
      ptype: form.ptype.trim(),
      v0: form.v0.trim(),
      v1: form.v1.trim(),
      v2: form.v2.trim(),
      v3: form.v3.trim(),
      v4: form.v4.trim(),
      v5: form.v5.trim(),
    };

    if (editingId.value) {
      await updateCasbinRule(editingId.value, payload);
      ElMessage.success(t('casbin_rule.updated', 'CasbinRule updated'));
    } else {
      await createCasbinRule(payload);
      ElMessage.success(t('casbin_rule.created', 'CasbinRule created'));
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('casbin_rule.save_failed', 'Save failed'));
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: CasbinRuleItem) {
  await ElMessageBox.confirm(t('casbin_rule.confirm_delete', 'Delete CasbinRule {name}?', { name: row.id }), t('casbin_rule.delete_title', 'Delete CasbinRule'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteCasbinRule(row.id);
  ElMessage.success(t('casbin_rule.deleted', 'CasbinRule deleted'));
  await loadItems();
}

function handleSearch() {
  query.page = 1;
  void loadItems();
}

function handleReset() {
  query.keyword = '';
  query.page = 1;
  void loadItems();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadItems();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadItems();
}

onMounted(async () => {
  await preloadRouteNamespaces({
    meta: {
      i18nNamespaces: ['casbin_rule'],
    },
  } as any);
  await loadItems();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      :title="t('casbin_rule.title', 'Rule management')"
      :description="t('casbin_rule.description', 'Manage authorization policy rules, including listing, editing, and deletion.')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'casbin_rule:create'" type="primary" @click="openCreate">{{ t('common.create', 'Create') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('common.search', 'Search')">
            <el-input v-model="query.keyword" clearable :placeholder="t('casbin_rule.keyword_placeholder', 'Search CasbinRule data')" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="id" :label="t('casbin_rule.id', 'ID')" min-width="160" />
        <el-table-column
          prop="ptype"
          :label="t('casbin_rule.ptype', 'Ptype')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.ptype || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v0"
          :label="t('casbin_rule.v0', 'V0')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v0 || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v1"
          :label="t('casbin_rule.v1', 'V1')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v1 || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v2"
          :label="t('casbin_rule.v2', 'V2')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v2 || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v3"
          :label="t('casbin_rule.v3', 'V3')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v3 || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v4"
          :label="t('casbin_rule.v4', 'V4')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v4 || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="v5"
          :label="t('casbin_rule.v5', 'V5')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.v5 || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('casbin_rule.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('casbin_rule.updated_at', 'Updated at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('casbin_rule.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'casbin_rule:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'casbin_rule:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      :title="editingId ? t('casbin_rule.edit_title', 'Edit rule') : t('casbin_rule.create_title', 'New rule')"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('casbin_rule.ptype', 'Ptype')">
          <el-input v-model="form.ptype" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'Ptype' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v0', 'V0')">
          <el-input v-model="form.v0" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V0' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v1', 'V1')">
          <el-input v-model="form.v1" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V1' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v2', 'V2')">
          <el-input v-model="form.v2" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V2' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v3', 'V3')">
          <el-input v-model="form.v3" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V3' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v4', 'V4')">
          <el-input v-model="form.v4" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V4' })" />
        </el-form-item>
        <el-form-item :label="t('casbin_rule.v5', 'V5')">
          <el-input v-model="form.v5" :placeholder="t('casbin_rule.placeholder', 'Enter {field}', { field: 'V5' })" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
