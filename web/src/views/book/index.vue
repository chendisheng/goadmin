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
    tech: t('book.category.tech', 'Technology'),
    novel: t('book.category.novel', 'Novel'),
    history: t('book.category.history', 'History'),
    other: t('book.category.other', 'Other'),
  });
}

function statusEnumLabel(value: unknown) {
  return formatEnumLabel(value, {
    draft: t('book.status.draft', 'Draft'),
    published: t('book.status.published', 'Published'),
    off_shelf: t('book.status.off_shelf', 'Off shelf'),
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
      ElMessage.success(t('book.updated', 'Book updated'));
    } else {
      await createBook(payload);
      ElMessage.success(t('book.created', 'Book created'));
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : t('book.save_failed', 'Save failed'));
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: BookItem) {
  await ElMessageBox.confirm(t('book.confirm_delete', 'Delete Book {id}?', { id: row.id }), t('book.delete_title', 'Delete Book'), {
    type: 'warning',
    confirmButtonText: t('common.delete', 'Delete'),
    cancelButtonText: t('common.cancel', 'Cancel'),
  });
  await deleteBook(row.id);
  ElMessage.success(t('book.deleted', 'Book deleted'));
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
      :title="t('book.title', 'Book management')"
      :description="t('book.description', 'CRUD page generated by goadmin-cli, ready for listing, editing, and deletion.')"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">{{ t('common.refresh', 'Refresh') }}</el-button>
        <el-button v-permission="'book:create'" type="primary" @click="openCreate">{{ t('book.create', 'Add Book') }}</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item :label="t('book.keyword', 'Keyword')">
            <el-input v-model="query.keyword" clearable :placeholder="t('book.keyword_placeholder', 'Search Book data')" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">{{ t('common.search', 'Search') }}</el-button>
            <el-button @click="handleReset">{{ t('common.reset', 'Reset') }}</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="id" :label="t('book.id', 'ID')" min-width="160" />
        <el-table-column
          prop="tenant_id"
          :label="t('book.tenant_id', 'Tenant ID')"
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
        <el-table-column :label="t('book.created_at', 'Created at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('book.updated_at', 'Updated at')" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions', 'Actions')" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'book:update'" link type="primary" @click="openEdit(row)">{{ t('common.edit', 'Edit') }}</el-button>
            <el-button v-permission="'book:delete'" link type="danger" @click="removeRow(row)">{{ t('common.delete', 'Delete') }}</el-button>
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
      :title="editingId ? t('book.edit_title', 'Edit Book') : t('book.create_title', 'Add Book')"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item :label="t('book.tenant_id', 'Tenant ID')">
          <el-input v-model="form.tenant_id" :placeholder="t('book.tenant_id_placeholder', 'Enter Tenant ID')" />
        </el-form-item>
        <el-form-item :label="t('book.title_field', 'Title')" required>
          <el-input v-model="form.title" :placeholder="t('book.title_placeholder', 'Enter title')" />
        </el-form-item>
        <el-form-item :label="t('book.author', 'Author')">
          <el-input v-model="form.author" :placeholder="t('book.author_placeholder', 'Enter author')" />
        </el-form-item>
        <el-form-item :label="t('book.isbn', 'Isbn')">
          <el-input v-model="form.isbn" :placeholder="t('book.isbn_placeholder', 'Enter isbn')" />
        </el-form-item>
        <el-form-item :label="t('book.publisher', 'Publisher')">
          <el-input v-model="form.publisher" :placeholder="t('book.publisher_placeholder', 'Enter publisher')" />
        </el-form-item>
        <el-form-item :label="t('book.publish_date', 'Publish Date')">
          <el-date-picker
            v-model="form.publish_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            :placeholder="t('book.publish_date_placeholder', 'Select publish date')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('book.category', 'Category')">
          <el-select v-model="form.category" style="width: 100%">
            <el-option :label="t('book.category.tech', 'Technology')" value="tech" />
            <el-option :label="t('book.category.novel', 'Novel')" value="novel" />
            <el-option :label="t('book.category.history', 'History')" value="history" />
            <el-option :label="t('book.category.other', 'Other')" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('book.description_field', 'Description')">
          <el-input v-model="form.description" type="textarea" :rows="4" :placeholder="t('book.description_placeholder', 'Enter description')" />
        </el-form-item>
        <el-form-item :label="t('book.status', 'Status')">
          <el-select v-model="form.status" style="width: 100%">
            <el-option :label="t('book.status.draft', 'Draft')" value="draft" />
            <el-option :label="t('book.status.published', 'Published')" value="published" />
            <el-option :label="t('book.status.off_shelf', 'Off shelf')" value="off_shelf" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('book.price', 'Price')">
          <el-input-number v-model="form.price" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('book.stock_quantity', 'Stock Quantity')">
          <el-input-number v-model="form.stock_quantity" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item :label="t('book.cover_image_url', 'Cover Image Url')">
          <el-input v-model="form.cover_image_url" :placeholder="t('book.cover_image_url_placeholder', 'Enter cover image url')" />
        </el-form-item>
        <el-form-item :label="t('book.tags', 'Tags')">
          <el-input v-model="form.tags" :placeholder="t('book.tags_placeholder', 'Comma-separated, for example: ai,ml')" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
