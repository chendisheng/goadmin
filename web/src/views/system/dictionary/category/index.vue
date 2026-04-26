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
  return status === 'disabled' ? t('dictionary.category.disabled', '禁用') : t('dictionary.category.enabled', '启用');
}

async function submitForm() {
  if (form.code.trim() === '' || form.name.trim() === '') {
    ElMessage.warning(t('dictionary.category.validation_required', '请输入字典编码和名称'));
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
      ElMessage.success(t('dictionary.category.updated', '字典分类已更新'));
    } else {
      await createDictionaryCategory(payload);
      ElMessage.success(t('dictionary.category.created', '字典分类已创建'));
    }

    dialogVisible.value = false;
    await loadCategories();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: DictionaryCategoryItem) {
  await ElMessageBox.confirm(t('dictionary.category.confirm_delete', '确认删除字典分类 {name} 吗？', { name: row.name }), t('dictionary.category.delete_title', '删除分类'), {
    type: 'warning',
    confirmButtonText: t('common.delete', '删除'),
    cancelButtonText: t('common.cancel', '取消'),
  });
  await deleteDictionaryCategory(row.id);
  ElMessage.success(t('dictionary.category.deleted', '字典分类已删除'));
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
      :title="t('dictionary.category.title', '字典分类管理')"
      :description="t('dictionary.category.description', '维护字典分类编码、名称与启停状态，供系统内其他模块复用。')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadCategories">{{ t('common.refresh', '刷新') }}</el-button>
        <el-button v-permission="'dictionary:category:create'" type="primary" @click="openCreate">{{ t('dictionary.category.create', '新增分类') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('dictionary.category.keyword', '关键字')">
            <el-input v-model="query.keyword" clearable :placeholder="t('dictionary.category.keyword_placeholder', '编码 / 名称 / 备注')" />
          </el-form-item>
          <el-form-item :label="t('dictionary.category.status', '状态')">
            <el-select v-model="query.status" clearable :placeholder="t('dictionary.category.all_status', '全部状态')" style="width: 180px">
              <el-option :label="t('dictionary.category.enabled', '启用')" value="enabled" />
              <el-option :label="t('dictionary.category.disabled', '禁用')" value="disabled" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', '查询') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', '重置') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="code" :label="t('dictionary.category.code', '分类编码')" min-width="160" />
        <el-table-column prop="name" :label="t('dictionary.category.name', '分类名称')" min-width="160" />
        <el-table-column prop="description" :label="t('dictionary.category.description_label', '描述')" min-width="220" show-overflow-tooltip />
        <el-table-column :label="t('dictionary.category.status', '状态')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="plain">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort" :label="t('dictionary.category.sort', '排序')" width="90" />
        <el-table-column prop="remark" :label="t('dictionary.category.remark', '备注')" min-width="180" show-overflow-tooltip />
        <el-table-column :label="t('dictionary.category.updated_at', '更新时间')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', '操作')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'dictionary:category:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', '编辑') }}</el-button>
            <el-button v-permission="'dictionary:category:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', '删除') }}</el-button>
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
      :title="editingId ? t('dictionary.category.edit_title', '编辑字典分类') : t('dictionary.category.create_title', '新增字典分类')"
      :loading="dialogLoading"
      width="720px"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('dictionary.category.code', '分类编码')" required>
          <el-input v-model="form.code" :placeholder="t('dictionary.category.code_placeholder', '请输入分类编码')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.name', '分类名称')" required>
          <el-input v-model="form.name" :placeholder="t('dictionary.category.name_placeholder', '请输入分类名称')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.description_label', '描述')">
          <el-input v-model="form.description" type="textarea" :rows="3" :placeholder="t('dictionary.category.description_placeholder', '请输入描述')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.status', '状态')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('dictionary.category.enabled', '启用')" value="enabled" />
            <el-option :label="t('dictionary.category.disabled', '禁用')" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('dictionary.category.sort', '排序')">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('dictionary.category.remark', '备注')">
          <el-input v-model="form.remark" type="textarea" :rows="3" :placeholder="t('dictionary.category.remark_placeholder', '请输入备注')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
