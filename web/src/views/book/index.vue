<script setup lang="ts">
import { onActivated, onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createBook, deleteBook, listbooks, updateBook } from '@/api/book';
import { useAppI18n } from '@/i18n';
import { formatDateTime } from '@/utils/admin';

type BookItem = {
  id: string;
  tenant_id?: string;
  title?: string;
  author?: string;
  isbn?: string;
  publisher?: string;
  publish_date?: string;
  category?: string;
  description?: string;
  status?: string;
  price?: number;
  stock_quantity?: number;
  cover_image_url?: string;
  tags?: string;
  created_at?: string;
  updated_at?: string;
};

type BookFormState = {
  tenant_id: string;
  title: string;
  author: string;
  isbn: string;
  publisher: string;
  publish_date: string;
  category: string;
  description: string;
  status: string;
  price: number;
  stock_quantity: number;
  cover_image_url: string;
  tags: string;
};

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<BookItem[]>([]);
const total = ref(0);
const editingId = ref('');
const { t } = useAppI18n();

const query = reactive({
  keyword: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): BookFormState => ({
  tenant_id: '',
  title: '',
  author: '',
  isbn: '',
  publisher: '',
  publish_date: '',
  category: '',
  description: '',
  status: '',
  price: 0,
  stock_quantity: 0,
  cover_image_url: '',
  tags: '',
});

const form = reactive<BookFormState>(defaultForm());

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

function categoryEnumLabel(value: unknown) {
  return formatEnumLabel(value, {
    tech: t('book.category.tech', '技术'),
    novel: t('book.category.novel', '小说'),
    history: t('book.category.history', '历史'),
    other: t('book.category.other', '其他'),
  });
}

function statusEnumLabel(value: unknown) {
  return formatEnumLabel(value, {
    draft: t('book.status.draft', '草稿'),
    published: t('book.status.published', '已发布'),
    off_shelf: t('book.status.off_shelf', '已下架'),
  });
}

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadItems() {
  tableLoading.value = true;
  try {
    const response = await listbooks({ ...query });
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

function openEdit(row: BookItem) {
  editingId.value = row.id;
  Object.assign(form, {
    tenant_id: row.tenant_id ?? '',
    title: row.title ?? '',
    author: row.author ?? '',
    isbn: row.isbn ?? '',
    publisher: row.publisher ?? '',
    publish_date: row.publish_date ?? '',
    category: row.category ?? '',
    description: row.description ?? '',
    status: row.status ?? '',
    price: Number(row.price ?? 0),
    stock_quantity: Number(row.stock_quantity ?? 0),
    cover_image_url: row.cover_image_url ?? '',
    tags: row.tags ?? '',
  });
  dialogVisible.value = true;
}

async function submitForm() {
  dialogLoading.value = true;
  try {
    const payload: BookFormState = {
      tenant_id: form.tenant_id.trim(),
      title: form.title.trim(),
      author: form.author.trim(),
      isbn: form.isbn.trim(),
      publisher: form.publisher.trim(),
      publish_date: form.publish_date,
      category: form.category,
      description: form.description.trim(),
      status: form.status,
      price: Number(form.price ?? 0),
      stock_quantity: Number(form.stock_quantity ?? 0),
      cover_image_url: form.cover_image_url.trim(),
      tags: form.tags.trim(),
    };

    if (editingId.value) {
      await updateBook(editingId.value, payload);
      ElMessage.success(t('book.updated', 'Book 已更新'));
    } else {
      await createBook(payload);
      ElMessage.success(t('book.created', 'Book 已创建'));
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('book.save_failed', '保存失败'));
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: BookItem) {
  await ElMessageBox.confirm(t('book.confirm_delete', '确认删除 Book {id} 吗？', { id: row.id }), t('book.delete_title', '删除Book'), {
    type: 'warning',
    confirmButtonText: t('common.delete', '删除'),
    cancelButtonText: t('common.cancel', '取消'),
  });
  await deleteBook(row.id);
  ElMessage.success(t('book.deleted', 'Book 已删除'));
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

onMounted(() => {
  void loadItems();
});

onActivated(() => {
  void loadItems();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable
      :title="t('book.title', 'Book管理')"
      :description="t('book.description', '由 goadmin-cli 生成的 CRUD 页面，可直接用于列表、编辑和删除。')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">{{ t('common.refresh', '刷新') }}</el-button>
        <el-button v-permission="'book:create'" type="primary" @click="openCreate">{{ t('book.create', '新增Book') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('book.keyword', '关键字')">
            <el-input v-model="query.keyword" clearable :placeholder="t('book.keyword_placeholder', '搜索Book数据')" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', '查询') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', '重置') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="id" :label="t('book.id', 'ID')" min-width="160" />
        <el-table-column
          prop="tenant_id"
          :label="t('book.tenant_id', 'Tenant Id')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.tenant_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="title"
          :label="t('book.title_field', 'Title')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.title || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="author"
          :label="t('book.author', 'Author')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.author || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="isbn"
          :label="t('book.isbn', 'Isbn')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.isbn || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="publisher"
          :label="t('book.publisher', 'Publisher')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.publisher || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="publish_date"
          :label="t('book.publish_date', 'Publish Date')"
          min-width="180"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.publish_date) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="category"
          :label="t('book.category', 'Category')"
          min-width="140"
        >
          <template #default="{ row }">
            {{ categoryEnumLabel(row.category) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="description"
          :label="t('book.description_field', 'Description')"
          min-width="220"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="t('book.status', 'Status')"
          min-width="140"
        >
          <template #default="{ row }">
            {{ statusEnumLabel(row.status) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="price"
          :label="t('book.price', 'Price')"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.price || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="stock_quantity"
          :label="t('book.stock_quantity', 'Stock Quantity')"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.stock_quantity || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="cover_image_url"
          :label="t('book.cover_image_url', 'Cover Image Url')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.cover_image_url || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="tags"
          :label="t('book.tags', 'Tags')"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.tags || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('book.created_at', '创建时间')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('book.updated_at', '更新时间')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', '操作')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'book:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', '编辑') }}</el-button>
            <el-button v-permission="'book:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', '删除') }}</el-button>
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
      :title="editingId ? t('book.edit_title', '编辑Book') : t('book.create_title', '新增Book')"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('book.tenant_id', 'Tenant Id')">
          <el-input v-model="form.tenant_id" :placeholder="t('book.tenant_id_placeholder', '请输入Tenant Id')" />
        </el-form-item>
        <el-form-item :label="t('book.title_field', 'Title')">
          <el-input v-model="form.title" :placeholder="t('book.title_placeholder', '请输入Title')" />
        </el-form-item>
        <el-form-item :label="t('book.author', 'Author')">
          <el-input v-model="form.author" :placeholder="t('book.author_placeholder', '请输入Author')" />
        </el-form-item>
        <el-form-item :label="t('book.isbn', 'Isbn')">
          <el-input v-model="form.isbn" :placeholder="t('book.isbn_placeholder', '请输入Isbn')" />
        </el-form-item>
        <el-form-item :label="t('book.publisher', 'Publisher')">
          <el-input v-model="form.publisher" :placeholder="t('book.publisher_placeholder', '请输入Publisher')" />
        </el-form-item>
        <el-form-item :label="t('book.publish_date', 'Publish Date')">
          <el-date-picker
            v-model="form.publish_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            :placeholder="t('book.publish_date_placeholder', '请选择Publish Date')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('book.category', 'Category')">
          <el-select v-model="form.category" filterable clearable :multiple="false" :placeholder="t('book.category_placeholder', '请选择Category')" style="width: 100%">
            <el-option :label="t('book.category.tech', '技术')" :value="'tech'" :disabled="false" />
            <el-option :label="t('book.category.novel', '小说')" :value="'novel'" :disabled="false" />
            <el-option :label="t('book.category.history', '历史')" :value="'history'" :disabled="false" />
            <el-option :label="t('book.category.other', '其他')" :value="'other'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('book.description_field', 'Description')">
          <el-input v-model="form.description" type="textarea" :rows="4" :placeholder="t('book.description_placeholder', '请输入Description')" />
        </el-form-item>
        <el-form-item :label="t('book.status', 'Status')">
          <el-select v-model="form.status" filterable clearable :multiple="false" :placeholder="t('book.status_placeholder', '请选择Status')" style="width: 100%">
            <el-option :label="t('book.status.draft', '草稿')" :value="'draft'" :disabled="false" />
            <el-option :label="t('book.status.published', '已发布')" :value="'published'" :disabled="false" />
            <el-option :label="t('book.status.off_shelf', '已下架')" :value="'off_shelf'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('book.price', 'Price')">
          <el-input-number v-model="form.price" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('book.stock_quantity', 'Stock Quantity')">
          <el-input-number v-model="form.stock_quantity" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('book.cover_image_url', 'Cover Image Url')">
          <el-input v-model="form.cover_image_url" :placeholder="t('book.cover_image_url_placeholder', '请输入Cover Image Url')" />
        </el-form-item>
        <el-form-item :label="t('book.tags', 'Tags')">
          <el-input v-model="form.tags" :placeholder="t('book.tags_placeholder', '请输入Tags')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
