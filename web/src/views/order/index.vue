<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

import AdminFormDialog from '@/components/admin/AdminFormDialog.vue';
import AdminTable from '@/components/admin/AdminTable.vue';
import { createOrder, deleteOrder, listorders, updateOrder } from '@/api/order';
import { formatDateTime } from '@/utils/admin';

type OrderItem = {
  id: string;
  tenant_id?: string;
  order_no?: string;
  user_id?: string;
  customer_name?: string;
  customer_email?: string;
  customer_phone?: string;
  shipping_address?: string;
  billing_address?: string;
  order_status?: string;
  payment_status?: string;
  payment_method?: string;
  currency?: string;
  total_amount?: number;
  discount_amount?: number;
  tax_amount?: number;
  shipping_amount?: number;
  final_amount?: number;
  order_date?: string;
  shipped_date?: string;
  delivered_date?: string;
  notes?: string;
  internal_notes?: string;
  created_at?: string;
  updated_at?: string;
};

type OrderFormState = {
  tenant_id: string;
  order_no: string;
  user_id: string;
  customer_name: string;
  customer_email: string;
  customer_phone: string;
  shipping_address: string;
  billing_address: string;
  order_status: string;
  payment_status: string;
  payment_method: string;
  currency: string;
  total_amount: number;
  discount_amount: number;
  tax_amount: number;
  shipping_amount: number;
  final_amount: number;
  order_date: string;
  shipped_date: string;
  delivered_date: string;
  notes: string;
  internal_notes: string;
};

const tableLoading = ref(false);
const dialogLoading = ref(false);
const dialogVisible = ref(false);
const rows = ref<OrderItem[]>([]);
const total = ref(0);
const editingId = ref('');

const query = reactive({
  keyword: '',
  page: 1,
  page_size: 10,
});

const defaultForm = (): OrderFormState => ({
  tenant_id: '',
  order_no: '',
  user_id: '',
  customer_name: '',
  customer_email: '',
  customer_phone: '',
  shipping_address: '',
  billing_address: '',
  order_status: '',
  payment_status: '',
  payment_method: '',
  currency: '',
  total_amount: 0,
  discount_amount: 0,
  tax_amount: 0,
  shipping_amount: 0,
  final_amount: 0,
  order_date: '',
  shipped_date: '',
  delivered_date: '',
  notes: '',
  internal_notes: '',
});

const form = reactive<OrderFormState>(defaultForm());

type EnumOption = {
  value: string;
  label: string;
  color?: string;
  disabled?: boolean;
  order?: number;
};
const order_statusEnumLabelMap: Record<string, string> = {
  ["pending"]: "待处理",
  ["paid"]: "已支付",
  ["shipped"]: "已发货",
  ["delivered"]: "已完成",
  ["cancelled"]: "已取消",
};

const order_statusEnumOptions: EnumOption[] = [
  { value: "pending", label: "待处理", color: "", disabled: false, order: 1 },
  { value: "paid", label: "已支付", color: "", disabled: false, order: 2 },
  { value: "shipped", label: "已发货", color: "", disabled: false, order: 3 },
  { value: "delivered", label: "已完成", color: "", disabled: false, order: 4 },
  { value: "cancelled", label: "已取消", color: "", disabled: false, order: 5 },
];
const payment_statusEnumLabelMap: Record<string, string> = {
  ["unpaid"]: "未支付",
  ["paid"]: "已支付",
  ["refunded"]: "已退款",
  ["failed"]: "支付失败",
};

const payment_statusEnumOptions: EnumOption[] = [
  { value: "unpaid", label: "未支付", color: "", disabled: false, order: 1 },
  { value: "paid", label: "已支付", color: "", disabled: false, order: 2 },
  { value: "refunded", label: "已退款", color: "", disabled: false, order: 3 },
  { value: "failed", label: "支付失败", color: "", disabled: false, order: 4 },
];
const payment_methodEnumLabelMap: Record<string, string> = {
  ["wechat"]: "微信支付",
  ["alipay"]: "支付宝",
  ["card"]: "银行卡",
  ["paypal"]: "PayPal",
};

const payment_methodEnumOptions: EnumOption[] = [
  { value: "wechat", label: "微信支付", color: "", disabled: false, order: 1 },
  { value: "alipay", label: "支付宝", color: "", disabled: false, order: 2 },
  { value: "card", label: "银行卡", color: "", disabled: false, order: 3 },
  { value: "paypal", label: "PayPal", color: "", disabled: false, order: 4 },
];
const currencyEnumLabelMap: Record<string, string> = {
  ["CNY"]: "人民币",
  ["USD"]: "美元",
  ["EUR"]: "欧元",
  ["JPY"]: "日元",
};

const currencyEnumOptions: EnumOption[] = [
  { value: "CNY", label: "人民币", color: "", disabled: false, order: 1 },
  { value: "USD", label: "美元", color: "", disabled: false, order: 2 },
  { value: "EUR", label: "欧元", color: "", disabled: false, order: 3 },
  { value: "JPY", label: "日元", color: "", disabled: false, order: 4 },
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
    const response = await listorders({ ...query });
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

function openEdit(row: OrderItem) {
  editingId.value = row.id;
  Object.assign(form, {
    tenant_id: row.tenant_id ?? '',
    order_no: row.order_no ?? '',
    user_id: row.user_id ?? '',
    customer_name: row.customer_name ?? '',
    customer_email: row.customer_email ?? '',
    customer_phone: row.customer_phone ?? '',
    shipping_address: row.shipping_address ?? '',
    billing_address: row.billing_address ?? '',
    order_status: row.order_status ?? '',
    payment_status: row.payment_status ?? '',
    payment_method: row.payment_method ?? '',
    currency: row.currency ?? '',
    total_amount: Number(row.total_amount ?? 0),
    discount_amount: Number(row.discount_amount ?? 0),
    tax_amount: Number(row.tax_amount ?? 0),
    shipping_amount: Number(row.shipping_amount ?? 0),
    final_amount: Number(row.final_amount ?? 0),
    order_date: row.order_date ?? '',
    shipped_date: row.shipped_date ?? '',
    delivered_date: row.delivered_date ?? '',
    notes: row.notes ?? '',
    internal_notes: row.internal_notes ?? '',
  });
  dialogVisible.value = true;
}

async function submitForm() {
  dialogLoading.value = true;
  try {
    const payload: OrderFormState = {
      tenant_id: form.tenant_id.trim(),
      order_no: form.order_no.trim(),
      user_id: form.user_id.trim(),
      customer_name: form.customer_name.trim(),
      customer_email: form.customer_email.trim(),
      customer_phone: form.customer_phone.trim(),
      shipping_address: form.shipping_address.trim(),
      billing_address: form.billing_address.trim(),
      order_status: form.order_status,
      payment_status: form.payment_status,
      payment_method: form.payment_method,
      currency: form.currency,
      total_amount: Number(form.total_amount ?? 0),
      discount_amount: Number(form.discount_amount ?? 0),
      tax_amount: Number(form.tax_amount ?? 0),
      shipping_amount: Number(form.shipping_amount ?? 0),
      final_amount: Number(form.final_amount ?? 0),
      order_date: form.order_date,
      shipped_date: form.shipped_date,
      delivered_date: form.delivered_date,
      notes: form.notes.trim(),
      internal_notes: form.internal_notes.trim(),
    };

    if (editingId.value) {
      await updateOrder(editingId.value, payload);
      ElMessage.success('Order 已更新');
    } else {
      await createOrder(payload);
      ElMessage.success('Order 已创建');
    }

    dialogVisible.value = false;
    await loadItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败');
  } finally {
    dialogLoading.value = false;
  }
}

async function removeRow(row: OrderItem) {
  await ElMessageBox.confirm('确认删除 Order ' + row.id + ' 吗？', '删除Order', {
    type: 'warning',
    confirmButtonText: '删除',
    cancelButtonText: '取消',
  });
  await deleteOrder(row.id);
  ElMessage.success('Order 已删除');
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
      title="Order管理"
      description="由 goadmin-cli 生成的 CRUD 页面，可直接用于列表、编辑和删除。"
      :loading="tableLoading"
    >
      <template #actions>
        <el-button :loading="tableLoading" @click="loadItems">刷新</el-button>
        <el-button v-permission="'order:create'" type="primary" @click="openCreate">新增Order</el-button>
      </template>

      <template #filters>
        <el-form :inline="true" label-width="88px" class="admin-filters">
          <el-form-item label="关键字">
            <el-input v-model="query.keyword" clearable placeholder="搜索Order数据" />
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
          prop="order_no"
          label="Order No"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.order_no || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="user_id"
          label="User Id"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.user_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="customer_name"
          label="Customer Name"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.customer_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="customer_email"
          label="Customer Email"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.customer_email || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="customer_phone"
          label="Customer Phone"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.customer_phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="shipping_address"
          label="Shipping Address"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.shipping_address || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="billing_address"
          label="Billing Address"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.billing_address || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="order_status"
          label="Order Status"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.order_status, order_statusEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="payment_status"
          label="Payment Status"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.payment_status, payment_statusEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="payment_method"
          label="Payment Method"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.payment_method, payment_methodEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="currency"
          label="Currency"
          min-width="140"
        >
          <template #default="{ row }">
            {{ formatEnumLabel(row.currency, currencyEnumLabelMap) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="total_amount"
          label="Total Amount"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.total_amount || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="discount_amount"
          label="Discount Amount"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.discount_amount || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="tax_amount"
          label="Tax Amount"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.tax_amount || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="shipping_amount"
          label="Shipping Amount"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.shipping_amount || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="final_amount"
          label="Final Amount"
          min-width="120"
        >
          <template #default="{ row }">
            {{ row.final_amount || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="order_date"
          label="Order Date"
          min-width="180"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.order_date) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="shipped_date"
          label="Shipped Date"
          min-width="180"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.shipped_date) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="delivered_date"
          label="Delivered Date"
          min-width="180"
        >
          <template #default="{ row }">
            {{ formatDateTime(row.delivered_date) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="notes"
          label="Notes"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.notes || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="internal_notes"
          label="Internal Notes"
          min-width="140"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row.internal_notes || '-' }}
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
            <el-button v-permission="'order:update'" link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-permission="'order:delete'" link type="danger" @click="removeRow(row)">删除</el-button>
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
      :title="editingId ? '编辑Order' : '新增Order'"
      :loading="dialogLoading"
      @confirm="submitForm"
    >
      <el-form label-width="110px" class="admin-form">
        <el-form-item label="Tenant Id">
          <el-input v-model="form.tenant_id" :placeholder="'请输入Tenant Id'" />
        </el-form-item>
        <el-form-item label="Order No">
          <el-input v-model="form.order_no" :placeholder="'请输入Order No'" />
        </el-form-item>
        <el-form-item label="User Id">
          <el-input v-model="form.user_id" :placeholder="'请输入User Id'" />
        </el-form-item>
        <el-form-item label="Customer Name">
          <el-input v-model="form.customer_name" :placeholder="'请输入Customer Name'" />
        </el-form-item>
        <el-form-item label="Customer Email">
          <el-input v-model="form.customer_email" :placeholder="'请输入Customer Email'" />
        </el-form-item>
        <el-form-item label="Customer Phone">
          <el-input v-model="form.customer_phone" :placeholder="'请输入Customer Phone'" />
        </el-form-item>
        <el-form-item label="Shipping Address">
          <el-input v-model="form.shipping_address" :placeholder="'请输入Shipping Address'" />
        </el-form-item>
        <el-form-item label="Billing Address">
          <el-input v-model="form.billing_address" :placeholder="'请输入Billing Address'" />
        </el-form-item>
        <el-form-item label="Order Status">
          <el-select v-model="form.order_status" filterable clearable :multiple="false" placeholder="请选择Order Status" style="width: 100%">
            <el-option :label="'待处理'" :value="'pending'" :disabled="false" />
            <el-option :label="'已支付'" :value="'paid'" :disabled="false" />
            <el-option :label="'已发货'" :value="'shipped'" :disabled="false" />
            <el-option :label="'已完成'" :value="'delivered'" :disabled="false" />
            <el-option :label="'已取消'" :value="'cancelled'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Payment Status">
          <el-select v-model="form.payment_status" filterable clearable :multiple="false" placeholder="请选择Payment Status" style="width: 100%">
            <el-option :label="'未支付'" :value="'unpaid'" :disabled="false" />
            <el-option :label="'已支付'" :value="'paid'" :disabled="false" />
            <el-option :label="'已退款'" :value="'refunded'" :disabled="false" />
            <el-option :label="'支付失败'" :value="'failed'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Payment Method">
          <el-select v-model="form.payment_method" filterable clearable :multiple="false" placeholder="请选择Payment Method" style="width: 100%">
            <el-option :label="'微信支付'" :value="'wechat'" :disabled="false" />
            <el-option :label="'支付宝'" :value="'alipay'" :disabled="false" />
            <el-option :label="'银行卡'" :value="'card'" :disabled="false" />
            <el-option :label="'PayPal'" :value="'paypal'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Currency">
          <el-select v-model="form.currency" filterable clearable :multiple="false" placeholder="请选择Currency" style="width: 100%">
            <el-option :label="'人民币'" :value="'CNY'" :disabled="false" />
            <el-option :label="'美元'" :value="'USD'" :disabled="false" />
            <el-option :label="'欧元'" :value="'EUR'" :disabled="false" />
            <el-option :label="'日元'" :value="'JPY'" :disabled="false" />
          </el-select>
        </el-form-item>
        <el-form-item label="Total Amount">
          <el-input-number v-model="form.total_amount" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Discount Amount">
          <el-input-number v-model="form.discount_amount" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Tax Amount">
          <el-input-number v-model="form.tax_amount" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Shipping Amount">
          <el-input-number v-model="form.shipping_amount" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Final Amount">
          <el-input-number v-model="form.final_amount" :controls="false" style="width: 100%" />
        </el-form-item>
        <el-form-item label="Order Date">
          <el-date-picker
            v-model="form.order_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择Order Date"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="Shipped Date">
          <el-date-picker
            v-model="form.shipped_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择Shipped Date"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="Delivered Date">
          <el-date-picker
            v-model="form.delivered_date"
            type="datetime"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            placeholder="请选择Delivered Date"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="Notes">
          <el-input v-model="form.notes" :placeholder="'请输入Notes'" />
        </el-form-item>
        <el-form-item label="Internal Notes">
          <el-input v-model="form.internal_notes" :placeholder="'请输入Internal Notes'" />
        </el-form-item>
      </el-form>
    </AdminFormDialog>
  </div>
</template>
