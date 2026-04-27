<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import {
  createDictionaryItem,
  deleteDictionaryItem,
  fetchDictionaryCategories,
  fetchDictionaryItem,
  fetchDictionaryItems,
  fetchDictionaryLookupItem,
  fetchDictionaryLookupItems,
  updateDictionaryItem,
} from '@/api/dictionary';
import { useAppI18n } from '@/i18n';
import type { DictionaryCategoryItem, DictionaryItem, DictionaryItemFormState } from '@/types/dictionary';
import { formatDateTime, statusTagType } from '@/utils/admin';

const { t } = useAppI18n();
const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const categoryLoading = ref(false);
const lookupLoading = ref(false);
const rows = ref<DictionaryItem[]>([]);
const total = ref(0);
const categoryOptions = ref<DictionaryCategoryItem[]>([]);
const editingId = ref('');
const lookupResult = ref<DictionaryItem | null>(null);
const lookupItems = ref<DictionaryItem[]>([]);

const query = reactive({
  category_id: '',
  category_code: '',
  keyword: '',
  status: '',
  page: 1,
  page_size: 10,
});

const lookupForm = reactive({
  category_code: '',
  value: '',
});

const defaultForm = (): DictionaryItemFormState => ({
  category_id: '',
  value: '',
  label: '',
  tag_type: '',
  tag_color: '',
  extra: '',
  is_default: false,
  status: 'enabled',
  sort: 0,
  remark: '',
});

const form = reactive<DictionaryItemFormState>(defaultForm());

const categoryMap = computed(() => {
  const map = new Map<string, DictionaryCategoryItem>();
  for (const item of categoryOptions.value) {
    map.set(item.id, item);
  }
  return map;
});

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadCategories() {
  categoryLoading.value = true;
  try {
    const response = await fetchDictionaryCategories({ keyword: '', status: '', page: 1, page_size: 200 });
    categoryOptions.value = response.items;
  } finally {
    categoryLoading.value = false;
  }
}

async function loadItems() {
  tableLoading.value = true;
  try {
    const response = await fetchDictionaryItems({ ...query });
    rows.value = response.items;
    total.value = response.total;
  } finally {
    tableLoading.value = false;
  }
}

function categoryLabel(categoryId: string): string {
  const category = categoryMap.value.get(categoryId);
  if (!category) {
    return categoryId || '-';
  }
  return `${category.name} (${category.code})`;
}

function openCreate() {
  editingId.value = '';
  resetForm();
  if (query.category_id) {
    form.category_id = query.category_id;
  }
  dialogVisible.value = true;
}

async function openEdit(row: DictionaryItem) {
  editingId.value = row.id;
  const detail = await fetchDictionaryItem(row.id);
  Object.assign(form, {
    ...defaultForm(),
    category_id: detail.category_id,
    value: detail.value,
    label: detail.label,
    tag_type: detail.tag_type ?? '',
    tag_color: detail.tag_color ?? '',
    extra: detail.extra ?? '',
    is_default: detail.is_default,
    status: detail.status || 'enabled',
    sort: detail.sort ?? 0,
    remark: detail.remark ?? '',
  });
  dialogVisible.value = true;
}

function statusLabel(status: string): string {
  return status === 'disabled' ? t('dictionary.item.disabled', 'Disabled') : t('dictionary.item.enabled', 'Enabled');
}

function defaultLabel(value: boolean): string {
  return value ? t('dictionary.item.yes', 'Yes') : t('dictionary.item.no', 'No');
}

async function submitForm() {
  if (form.category_id.trim() === '' || form.value.trim() === '' || form.label.trim() === '') {
    ElMessage.warning(t('dictionary.item.validation_required', 'Enter the category, value, and label'));
    return;
  }
  dialogLoading.value = true;
  try {
    const payload: DictionaryItemFormState = {
      ...form,
      category_id: form.category_id.trim(),
      value: form.value.trim(),
      label: form.label.trim(),
      tag_type: form.tag_type.trim(),
      tag_color: form.tag_color.trim(),
      extra: form.extra.trim(),
      is_default: Boolean(form.is_default),
      status: form.status.trim() || 'enabled',
      sort: Number(form.sort) || 0,
      remark: form.remark.trim(),
    };

    if (editingId.value) {
      await updateDictionaryItem(editingId.value, payload);
      ElMessage.success(t('dictionary.item.updated', 'Dictionary item updated'));
    } else {
      await createDictionaryItem(payload);
      ElMessage.success(t('dictionary.item.created', 'Dictionary item created'));
    }

    dialogVisible.value = false;
    await loadItems();
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: DictionaryItem) {
  await ElMessageBox.confirm(t('dictionary.item.confirm_delete', 'Delete dictionary item {label} / {value}?', { label: row.label, value: row.value }), t('dictionary.item.delete_title', 'Delete dictionary item'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteDictionaryItem(row.id);
  ElMessage.success(t('dictionary.item.deleted', 'Dictionary item deleted'));
  await loadItems();
}

function handleSearch() {
  query.page = 1;
  void loadItems();
}

function handleReset() {
  query.category_id = '';
  query.category_code = '';
  query.keyword = '';
  query.status = '';
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

async function runLookupList() {
  if (lookupForm.category_code.trim() === '') {
    ElMessage.warning(t('dictionary.item.lookup_category_required', 'Enter the category code'));
    return;
  }
  lookupLoading.value = true;
  try {
    const response = await fetchDictionaryLookupItems(lookupForm.category_code.trim());
    lookupItems.value = response.items;
    lookupResult.value = response.items[0] ?? null;
  } finally {
    lookupLoading.value = false;
  }
}

async function runLookupItem() {
  if (lookupForm.category_code.trim() === '' || lookupForm.value.trim() === '') {
    ElMessage.warning(t('dictionary.item.lookup_required', 'Enter the category code and value'));
    return;
  }
  lookupLoading.value = true;
  try {
    lookupResult.value = await fetchDictionaryLookupItem(lookupForm.category_code.trim(), lookupForm.value.trim());
    lookupItems.value = lookupResult.value ? [lookupResult.value] : [];
  } finally {
    lookupLoading.value = false;
  }
}

onMounted(async () => {
  await loadCategories();
  await loadItems();
});
</script>

<template>
  <div class="admin-page">
    <el-row :gutter="20">
      <el-col :xs="24" :xl="16">
        <AdminTable
          :title="t('dictionary.item.title', 'Dictionary items')"
          :description="t('dictionary.item.description', 'Maintain dictionary item values, labels, default flags, and enable/disable status with category-based filtering.')"
          :loading="tableLoading"
        >
          <template #actions>
            <el-button :loading="tableLoading" @click="loadItems">{{ t('common.refresh', 'Refresh') }}</el-button>
            <el-button v-permission="'dictionary:item:create'" type="primary" @click="openCreate">{{ t('dictionary.item.create', 'Add item') }}</el-button>
          </template>

          <template #filters>
            <el-form :inline="true" label-width="88px" class="admin-filters">
              <el-form-item :label="t('dictionary.item.category', 'Category')">
                <el-select v-model="query.category_id" clearable filterable :placeholder="t('dictionary.item.all_categories', 'All categories')" style="width: 240px" :loading="categoryLoading">
                  <el-option v-for="item in categoryOptions" :key="item.id" :label="`${item.name} (${item.code})`" :value="item.id" />
                </el-select>
              </el-form-item>
              <el-form-item :label="t('dictionary.item.keyword', 'Keyword')">
                <el-input v-model="query.keyword" clearable :placeholder="t('dictionary.item.keyword_placeholder', 'Value / label / remark')" />
              </el-form-item>
              <el-form-item :label="t('dictionary.item.status', 'Status')">
                <el-select v-model="query.status" clearable :placeholder="t('dictionary.item.all_status', 'All statuses')" style="width: 180px">
                  <el-option :label="t('dictionary.item.enabled', 'Enabled')" value="enabled" />
                  <el-option :label="t('dictionary.item.disabled', 'Disabled')" value="disabled" />
                </el-select>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
                <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
              </el-form-item>
            </el-form>
          </template>

          <el-table :data="rows" border row-key="id" v-loading="tableLoading">
            <el-table-column :label="t('dictionary.item.category', 'Category')" min-width="200" show-overflow-tooltip>
              <template #default="{ row }">
                {{ categoryLabel(row.category_id) }}
              </template>
            </el-table-column>
            <el-table-column prop="value" :label="t('dictionary.item.value', 'Value')" min-width="160" />
            <el-table-column prop="label" :label="t('dictionary.item.label', 'Label')" min-width="160" />
            <el-table-column prop="tag_type" :label="t('dictionary.item.tag_type', 'Tag type')" width="120" />
            <el-table-column :label="t('dictionary.item.default', 'Default')" width="90">
              <template #default="{ row }">
                <el-tag :type="row.is_default ? 'success' : 'info'" effect="plain">
                  {{ defaultLabel(row.is_default) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('dictionary.item.status', 'Status')" width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" effect="plain">
                  {{ statusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sort" :label="t('dictionary.item.sort', 'Sort')" width="90" />
            <el-table-column prop="remark" :label="t('dictionary.item.remark', 'Remark')" min-width="180" show-overflow-tooltip />
            <el-table-column :label="t('dictionary.item.updated_at', 'Updated at')" min-width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.updated_at) }}
              </template>
            </el-table-column>
            <el-table-column :label="t('common.actions', 'Actions')" width="180" fixed="right">
              <template #default="{ row }">
                <el-button v-permission="'dictionary:item:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
                <el-button v-permission="'dictionary:item:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      </el-col>

      <el-col :xs="24" :xl="8">
        <el-space direction="vertical" fill :size="16" style="width: 100%">
          <el-card shadow="never">
            <template #header>
              <strong>{{ t('dictionary.item.lookup_title', 'Dictionary lookup') }}</strong>
            </template>
            <el-form label-width="96px">
              <el-form-item :label="t('dictionary.item.lookup_category_code', 'Category code')">
                <el-input v-model="lookupForm.category_code" :placeholder="t('dictionary.item.lookup_category_code_placeholder', 'For example: system_status')" />
              </el-form-item>
              <el-form-item :label="t('dictionary.item.lookup_value', 'Value')">
                <el-input v-model="lookupForm.value" :placeholder="t('dictionary.item.lookup_value_placeholder', 'Only fill this when querying a single item')" />
              </el-form-item>
              <el-form-item>
                <el-space>
                  <el-button :loading="lookupLoading" type="primary" @click="runLookupList">{{ t('dictionary.item.lookup_list', 'Query category list') }}</el-button>
                  <el-button :loading="lookupLoading" @click="runLookupItem">{{ t('dictionary.item.lookup_item', 'Query single item') }}</el-button>
                </el-space>
              </el-form-item>
            </el-form>
          </el-card>

          <el-card shadow="never">
            <template #header>
              <strong>{{ t('dictionary.item.lookup_result_title', 'Query results') }}</strong>
            </template>
            <el-empty v-if="!lookupItems.length" :description="t('common.no_result', 'No results')" />
            <el-table v-else :data="lookupItems" size="small" border>
              <el-table-column prop="value" :label="t('dictionary.item.value', 'Value')" min-width="110" />
              <el-table-column prop="label" :label="t('dictionary.item.label', 'Label')" min-width="120" />
              <el-table-column :label="t('dictionary.item.status', 'Status')" width="90">
                <template #default="{ row }">
                  {{ statusLabel(row.status) }}
                </template>
              </el-table-column>
            </el-table>
            <div v-if="lookupResult" class="lookup-result-card">
              <el-divider>{{ t('dictionary.item.lookup_single_result', 'Single item result') }}</el-divider>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item :label="t('dictionary.item.value', 'Value')">{{ lookupResult.value }}</el-descriptions-item>
                <el-descriptions-item :label="t('dictionary.item.label', 'Label')">{{ lookupResult.label }}</el-descriptions-item>
                <el-descriptions-item :label="t('dictionary.item.default', 'Default')">{{ defaultLabel(lookupResult.is_default) }}</el-descriptions-item>
                <el-descriptions-item :label="t('dictionary.item.status', 'Status')">{{ statusLabel(lookupResult.status) }}</el-descriptions-item>
              </el-descriptions>
            </div>
          </el-card>
        </el-space>
      </el-col>
    </el-row>

    <AdminFormDialog
      v-model="dialogVisible"
      :title="editingId ? t('dictionary.item.edit_title', 'Edit dictionary item') : t('dictionary.item.create_title', 'New dictionary item')"
      :loading="dialogLoading"
      width="760px"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('dictionary.item.category', 'Category')" required>
          <el-select v-model="form.category_id" filterable :placeholder="t('dictionary.item.category_placeholder', 'Select a category')" style="width: 100%" :loading="categoryLoading">
            <el-option v-for="item in categoryOptions" :key="item.id" :label="`${item.name} (${item.code})`" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('dictionary.item.value', 'Value')" required>
          <el-input v-model="form.value" :placeholder="t('dictionary.item.value_placeholder', 'Enter value')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.label', 'Label')" required>
          <el-input v-model="form.label" :placeholder="t('dictionary.item.label_placeholder', 'Enter label')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.tag_type', 'Tag type')">
          <el-input v-model="form.tag_type" :placeholder="t('dictionary.item.tag_type_placeholder', 'For example: success / warning / info')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.tag_color', 'Tag color')">
          <el-input v-model="form.tag_color" :placeholder="t('dictionary.item.tag_color_placeholder', 'For example: #67C23A')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.extra', 'Extra')">
          <el-input v-model="form.extra" type="textarea" :rows="3" :placeholder="t('dictionary.item.extra_placeholder', 'Enter JSON or text')" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.default', 'Default')">
          <el-switch v-model="form.is_default" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.status', 'Status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('dictionary.item.enabled', 'Enabled')" value="enabled" />
            <el-option :label="t('dictionary.item.disabled', 'Disabled')" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('dictionary.item.sort', 'Sort')">
          <el-input-number v-model="form.sort" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('dictionary.item.remark', 'Remark')">
          <el-input v-model="form.remark" type="textarea" :rows="3" :placeholder="t('dictionary.item.remark_placeholder', 'Enter remark')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
