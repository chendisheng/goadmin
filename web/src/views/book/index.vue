<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createBook, deleteBook, listbooks, updateBook } from '@/api/book';
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

type EnumOption = {
  value: string;
  label: string;
  color?: string;
  disabled?: boolean;
  order?: number;
};
const categoryEnumLabelMap: Record<string, string> = {
  ["tech"]: "技术",
  ["novel"]: "小说",
  ["history"]: "历史",
  ["other"]: "其他",
};

const categoryEnumOptions: EnumOption[] = [
  { value: "tech", label: "技术", color: "", disabled: false, order: 1 },
  { value: "novel", label: "小说", color: "", disabled: false, order: 2 },
  { value: "history", label: "历史", color: "", disabled: false, order: 3 },
  { value: "other", label: "其他", color: "", disabled: false, order: 4 },
];
const statusEnumLabelMap: Record<string, string> = {
  ["draft"]: "草稿",
  ["published"]: "已发布",
  ["off_shelf"]: "已下架",
};

const statusEnumOptions: EnumOption[] = [
  { value: "draft", label: "草稿", color: "", disabled: false, order: 1 },
  { value: "published", label: "已发布", color: "", disabled: false, order: 2 },
  { value: "off_shelf", label: "已下架", color: "", disabled: false, order: 3 },
];

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
      ElMessage.success('Book 已更新');
    } else {
      await createBook(payload);
      ElMessage.success('Book 已创建');
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: BookItem) {
  await ElMessageBox.confirm('确认删除 Book ' + row.id + ' 吗？', '删除Book', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteBook(row.id);
  ElMessage.success('Book 已删除');
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
</script>

<template>
  <div class="admin-page">
    <AdminTable
      title="Book管理"
      description="由 goadmin-cli 生成的 CRUD 页面，可直接用于列表、编辑和删除。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">刷新</el-button>
        <el-button v-permission="'book:create'" type="primary" @click="openCreate">新增Book</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="搜索Book数据" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="id" label="ID" min-width="160" />
        <el-table-column
          prop="tenant_id"
          label="Tenant Id"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.tenant_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="title"
          label="Title"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.title || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="author"
          label="Author"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.author || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="isbn"
          label="Isbn"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.isbn || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="publisher"
          label="Publisher"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.publisher || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="publish_date"
          label="Publish Date"
          min-width="180"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.publish_date) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="category"
          label="Category"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.category, categoryEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="description"
          label="Description"
          min-width="220"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          label="Status"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.status, statusEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="price"
          label="Price"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.price || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="stock_quantity"
          label="Stock Quantity"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.stock_quantity || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="cover_image_url"
          label="Cover Image Url"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.cover_image_url || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="tags"
          label="Tags"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.tags || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button v-permission="'book:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'book:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑Book' : '新增Book'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="Tenant Id">
          <el-input v-model="form.tenant_id" :placeholder="'请输入Tenant Id'" />
        </el-form-item>
        <el-form-item label="Title">
          <el-input v-model="form.title" :placeholder="'请输入Title'" />
        </el-form-item>
        <el-form-item label="Author">
          <el-input v-model="form.author" :placeholder="'请输入Author'" />
        </el-form-item>
        <el-form-item label="Isbn">
          <el-input v-model="form.isbn" :placeholder="'请输入Isbn'" />
        </el-form-item>
        <el-form-item label="Publisher">
          <el-input v-model="form.publisher" :placeholder="'请输入Publisher'" />
        </el-form-item>
        <el-form-item label="Publish Date">
          <el-date-picker
            v-model="form.publish_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择Publish Date"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="Category">
          <el-select v-model="form.category" filterable clearable :multiple="false" placeholder="请选择Category" style="width: 100%">
            <el-option :label="'技术'" :value="'tech'" :disabled="false" />
            <el-option :label="'小说'" :value="'novel'" :disabled="false" />
            <el-option :label="'历史'" :value="'history'" :disabled="false" />
            <el-option :label="'其他'" :value="'other'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Description">
          <el-input v-model="form.description" type="textarea" :rows="4" :placeholder="'请输入Description'" />
        </el-form-item>
        <el-form-item label="Status">
          <el-select v-model="form.status" filterable clearable :multiple="false" placeholder="请选择Status" style="width: 100%">
            <el-option :label="'草稿'" :value="'draft'" :disabled="false" />
            <el-option :label="'已发布'" :value="'published'" :disabled="false" />
            <el-option :label="'已下架'" :value="'off_shelf'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Price">
          <el-input-number v-model="form.price" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Stock Quantity">
          <el-input-number v-model="form.stock_quantity" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Cover Image Url">
          <el-input v-model="form.cover_image_url" :placeholder="'请输入Cover Image Url'" />
        </el-form-item>
        <el-form-item label="Tags">
          <el-input v-model="form.tags" :placeholder="'请输入Tags'" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
