<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createBook, deleteBook, getBook, listbooks, updateBook } from '@/api/book';
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

function resetForm() {
  Object.assign(form, defaultForm());
}

async function loadBooks() {
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

async function openEdit(row: BookItem) {
  editingId.value = row.id;
  const detail = await getBook(row.id).catch(() => row);
  Object.assign(form, {
    ...defaultForm(),
    tenant_id: detail.tenant_id ?? '',
    title: detail.title ?? '',
    author: detail.author ?? '',
    isbn: detail.isbn ?? '',
    publisher: detail.publisher ?? '',
    publish_date: detail.publish_date ?? '',
    category: detail.category ?? '',
    description: detail.description ?? '',
    status: detail.status ?? '',
    price: Number(detail.price ?? 0),
    stock_quantity: Number(detail.stock_quantity ?? 0),
    cover_image_url: detail.cover_image_url ?? '',
    tags: detail.tags ?? '',
  });
  dialogVisible.value = true;
}

async function submitForm() {
  if (form.title.trim() === '') {
    ElMessage.warning('请输入书名');
    return;
  }
  dialogLoading.value = true;
  try {
    const payload = {
      tenant_id: form.tenant_id.trim(),
      title: form.title.trim(),
      author: form.author.trim(),
      isbn: form.isbn.trim(),
      publisher: form.publisher.trim(),
      publish_date: form.publish_date,
      category: form.category.trim(),
      description: form.description.trim(),
      status: form.status.trim(),
      price: Number(form.price ?? 0),
      stock_quantity: Number(form.stock_quantity ?? 0),
      cover_image_url: form.cover_image_url.trim(),
      tags: form.tags.trim(),
    };

    if (editingId.value) {
      await updateBook(editingId.value, payload);
      ElMessage.success('图书已更新');
    } else {
      await createBook(payload);
      ElMessage.success('图书已创建');
    }

    dialogVisible.value = false;
    await loadBooks();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: BookItem) {
  await ElMessageBox.confirm(`确认删除图书 ${row.title || row.id} 吗？`, '删除图书', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteBook(row.id);
  ElMessage.success('图书已删除');
  await loadBooks();
}

function handleSearch() {
  query.page = 1;
  void loadBooks();
}

function handleReset() {
  query.keyword = '';
  query.page = 1;
  void loadBooks();
}

function handlePageChange(page: number) {
  query.page = page;
  void loadBooks();
}

function handleSizeChange(pageSize: number) {
  query.page_size = pageSize;
  query.page = 1;
  void loadBooks();
}

onMounted(() => {
  void loadBooks();
});
</script>

<template>
  <div class="admin-page">
    <AdminTable title="Book Management" description="由 goadmin-cli 生成的图书 CRUD 页面，可直接用于列表、编辑和删除。" :loading="tableLoading">
      <template #actions>
        <el-button :loading="tableLoading" @click="loadBooks">刷新</el-button>
        <el-button v-permission="'book:create'" type="primary" @click="openCreate">新增图书</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="书名 / 作者 / ISBN" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border row-key="id" v-loading="tableLoading">
        <el-table-column prop="title" label="书名" min-width="160" />
        <el-table-column prop="author" label="作者" min-width="140" />
        <el-table-column prop="isbn" label="ISBN" min-width="160" />
        <el-table-column prop="publisher" label="出版社" min-width="140" />
        <el-table-column label="出版日期" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.publish_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" min-width="120" />
        <el-table-column prop="status" label="状态" min-width="120" />
        <el-table-column label="价格" min-width="100">
          <template #default="{ row }">
            {{ row.price ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column label="库存" min-width="100">
          <template #default="{ row }">
            {{ row.stock_quantity ?? 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="tags" label="标签" min-width="180" show-overflow-tooltip />
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
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
      :title="editingId ? '编辑图书' : '新增图书'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="租户ID">
          <el-input v-model="form.tenant_id" placeholder="请输入租户ID" />
        </el-form-item>
        <el-form-item label="书名" required>
          <el-input v-model="form.title" placeholder="请输入书名" />
        </el-form-item>
        <el-form-item label="作者">
          <el-input v-model="form.author" placeholder="请输入作者" />
        </el-form-item>
        <el-form-item label="ISBN">
          <el-input v-model="form.isbn" placeholder="请输入 ISBN" />
        </el-form-item>
        <el-form-item label="出版社">
          <el-input v-model="form.publisher" placeholder="请输入出版社" />
        </el-form-item>
        <el-form-item label="出版日期">
          <el-date-picker
            v-model="form.publish_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择出版日期"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" placeholder="请输入分类" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="4" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="状态">
          <el-input v-model="form.status" placeholder="请输入状态" />
        </el-form-item>
        <el-form-item label="价格">
          <el-input-number v-model="form.price" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="form.stock_quantity" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="封面地址">
          <el-input v-model="form.cover_image_url" placeholder="请输入封面地址" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="请输入标签，多个值请用逗号分隔" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
