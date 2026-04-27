<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import {
  createDictionaryCategory,
  deleteDictionaryCategory,
  fetchDictionaryCategories,
  updateDictionaryCategory,
} from '@/api/dictionary';
import { useAppI18n } from '@/i18n';
import type { DictionaryCategoryFormState, DictionaryCategoryItem } from '@/types/dictionary';
import { formatDateTime, statusTagType } from '@/utils/admin';

const { t } = useAppI18n();
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<DictionaryCategoryItem[]>([]);
const total = ref(0);
const editingId = ref('');

const query = reactive({
  keyword: '',
  status: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): DictionaryCategoryFormState => ({
  code: '',
  name: '',
  description: '',
  status: 'enabled',
  sort: 0,
  remark: '',
});

const form = reactive<DictionaryCategoryFormState>(defaultForm());

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadCategories() {
  tableLoading.value = true;
  try {
    const response = await fetchDictionaryCategories({ ...query });
    rows.value = response.items;
    total.value = response.total;
  } finally {
    tableLoading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: DictionaryCategoryItem) {
  editingId.value = row.id;
  Object.assign(form, {
    ...defaultForm(),
    code: row.code,
    name: row.name,
    description: row.description ?? '',
    status: row.status || 'enabled',
    sort: row.sort ?? 0,
    remark: row.remark ?? '',
  });
  dialogVisible.value = true;
}

function statusLabel(status: string): string {
  return status === 'disabled' ? t('dictionary.category.disabled', 'Disabled') : t('dictionary.category.enabled', 'Enabled');
}

async function submitForm() {
  if (form.code.trim() === '' || form.name.trim() === '') {
    ElMessage.warning(t('dictionary.category.validation_required', 'Enter the dictionary code and name'));
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: DictionaryCategoryFormState = {
      ...form,
      code: form.code.trim(),
      name: form.name.trim(),
      description: form.description.trim(),
      status: form.status.trim() || 'enabled',
      sort: Number(form.sort) || 0,
      remark: form.remark.trim(),
    };

    if (editingId.value) {
      await updateDictionaryCategory(editingId.value, payload);
      ElMessage.success(t('dictionary.category.updated', 'Dictionary category updated'));
    } else {
      await createDictionaryCategory(payload);
      ElMessage.success(t('dictionary.category.created', 'Dictionary category created'));
    }

    dialogVisible.value = false;
    await loadCategories();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: DictionaryCategoryItem) {
  await ElMessageBox.confirm(t('dictionary.category.confirm_delete', 'Delete dictionary category {name}?', { name: row.name }), t('dictionary.category.delete_title', 'Delete category'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteDictionaryCategory(row.id);
  ElMessage.success(t('dictionary.category.deleted', 'Dictionary category deleted'));
  await loadCategories();
}

function handleSearch() {
  query.page = 1;
  void loadCategories();
}

function handleReset() {
  query.keyword = '';
  query.status = '';
  query.page = 1;
  void loadCategories();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadCategories();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadCategories();
}

onMounted(() => {
  void loadCategories();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      :title="t('dictionary.category.title', 'Dictionary categories')"
      :description="t('dictionary.category.description', 'Maintain dictionary category codes, names, and enable/disable status for reuse by other modules.')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadCategories">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'dictionary:category:create'" type="primary" @click="openCreate">{{ t('dictionary.category.create', 'Add category') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('dictionary.category.keyword', 'Keyword')">
            <el-input v-model="query.keyword" clearable :placeholder="t('dictionary.category.keyword_placeholder', 'Code / name / remark')" />
          </el-form-item>
          <el-form-item :label="t('dictionary.category.status', 'Status')">
            <el-select v-model="query.status" clearable :placeholder="t('dictionary.category.all_status', 'All statuses')" style="width: 180px">
              <el-option :label="t('dictionary.category.enabled', 'Enabled')" value="enabled" />
              <el-option :label="t('dictionary.category.disabled', 'Disabled')" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="code" :label="t('dictionary.category.code', 'Category code')" min-width="160" />
        <el-table-column prop="name" :label="t('dictionary.category.name', 'Category name')" min-width="160" />
        <el-table-column prop="description" :label="t('dictionary.category.description_label', 'Description')" min-width="220" show-overflow-tooltip />
        <el-table-column :label="t('dictionary.category.status', 'Status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" :label="t('dictionary.category.sort', 'Sort')" width="90" />
        <el-table-column prop="remark" :label="t('dictionary.category.remark', 'Remark')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('dictionary.category.updated_at', 'Updated at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'dictionary:category:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'dictionary:category:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      :title="editingId ? t('dictionary.category.edit_title', 'Edit dictionary category') : t('dictionary.category.create_title', 'New dictionary category')"
      :loading="dialogLoading"
      width="720px"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('dictionary.category.code', 'Category code')" required>
          <el-input v-model="form.code" :placeholder="t('dictionary.category.code_placeholder', 'Enter category code')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.name', 'Category name')" required>
          <el-input v-model="form.name" :placeholder="t('dictionary.category.name_placeholder', 'Enter category name')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.description_label', 'Description')">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="t('dictionary.category.description_placeholder', 'Enter description')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.status', 'Status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('dictionary.category.enabled', 'Enabled')" value="enabled" />
            <el-option :label="t('dictionary.category.disabled', 'Disabled')" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('dictionary.category.sort', 'Sort')">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.remark', 'Remark')">
          <el-input v-model="form.remark" type="textarea" :rows="3" :placeholder="t('dictionary.category.remark_placeholder', 'Enter remark')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
