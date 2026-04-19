<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createCasbinModel, deleteCasbinModel, listcasbin_models, updateCasbinModel } from '@/api/casbin_model';
import { formatDateTime } from '@/utils/admin';

type CasbinModelItem = {
  id?: string;
  name?: string;
  content?: string;
  created_at?: string;
  updated_at?: string;
};

type CasbinModelFormState = {
  content: string;
};

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<CasbinModelItem[]>([]);
const total = ref(0);
const editingId = ref('');

const query = reactive({
  keyword: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): CasbinModelFormState => ({
  content: '',
});

const form = reactive<CasbinModelFormState>(defaultForm());

function getRowKey(row: CasbinModelItem) {
  return row.id || row.name || '';
}

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
    const response = await listcasbin_models({ ...query });
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

function openEdit(row: CasbinModelItem) {
  editingId.value = getRowKey(row);
  Object.assign(form, {
    content: row.content ?? '',
  });
  dialogVisible.value = true;
}

async function submitForm() {
  dialogLoading.value = true;
  try {
    const payload: CasbinModelFormState = {
      content: form.content.trim(),
    };

    if (editingId.value) {
      await updateCasbinModel(editingId.value, payload);
      ElMessage.success('CasbinModel 已更新');
    } else {
      await createCasbinModel(payload);
      ElMessage.success('CasbinModel 已创建');
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: CasbinModelItem) {
  const rowKey = getRowKey(row);
  await ElMessageBox.confirm('确认删除 CasbinModel ' + rowKey + ' 吗？', '删除CasbinModel', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteCasbinModel(rowKey);
  ElMessage.success('CasbinModel 已删除');
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
      title="模型管理"
      description="授权模型配置的列表、编辑和删除入口。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">刷新</el-button>
        <el-button v-permission="'casbin_model:create'" type="primary" @click="openCreate">新增模型</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="搜索CasbinModel数据" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </template>

      <el-table :data="rows" border :row-key="getRowKey" v-loading="tableLoading">
        <el-table-column label="ID" min-width="160">
          <template #default="{ row }">
            {{ getRowKey(row) || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="content"
          label="Content"
          min-width="220"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.content || '-' }}
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
            <el-button v-permission="'casbin_model:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'casbin_model:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑模型' : '新增模型'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="Content">
          <el-input v-model="form.content" type="textarea" :rows="4" :placeholder="'请输入Content'" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
