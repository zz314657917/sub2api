<template>
  <AppLayout
    ><main class="space-y-4">
      <header class="flex items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold">我的商店订单</h2>
          <p class="mt-1 text-sm text-gray-500">
            商店关闭后，已购商品仍可在此查看和下载。
          </p>
        </div>
        <RouterLink class="btn btn-secondary" to="/store">返回商店</RouterLink>
      </header>
      <p v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</p>
      <button class="btn btn-secondary" :disabled="loading || busy !== null" @click="load">刷新订单</button>
      <StorePaymentLauncher ref="launcher" @refresh="load" />
      <div class="card overflow-x-auto">
        <table class="min-w-full text-sm">
          <thead>
            <tr class="border-b text-left text-gray-500">
              <th class="p-3">商品</th>
              <th class="p-3">金额</th>
              <th class="p-3">付款/交付</th>
              <th class="p-3">时间</th>
              <th class="p-3"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="5" class="p-6 text-center">加载中…</td>
            </tr>
            <tr v-else-if="!orders.length">
              <td colspan="5" class="p-6 text-center text-gray-500">
                暂无商店订单
              </td>
            </tr>
            <tr v-for="order in orders" :key="order.order_id" class="border-b">
              <td class="p-3">
                <b>{{ order.product_name }}</b
                ><small class="block text-gray-500">{{ kindLabel(order.kind) }}</small>
              </td>
              <td class="p-3">¥{{ cents(order.price_cents) }}</td>
              <td class="p-3">
                <span>{{ paymentLabel(order.payment_status) }}</span
                ><small class="block text-gray-500">{{
                  deliveryLabel(order.delivery_status)
                }}</small>
              </td>
              <td class="p-3">{{ date(order.created_at) }}</td>
              <td class="p-3 text-right">
                <template v-if="order.payment_status === 'PENDING'">
                  <button class="btn btn-primary btn-sm" :disabled="busy !== null" @click="resume(order)">继续付款</button>
                  <button class="btn btn-secondary btn-sm" :disabled="busy !== null" @click="cancel(order)">取消订单</button>
                </template>
                <button
                  v-if="order.delivery_status === 'delivered'"
                  class="btn btn-secondary btn-sm"
                  @click="showDelivery(order)"
                >
                  查看交付</button
                ><span
                  v-else-if="order.delivery_status === 'needs_attention'"
                  class="text-xs text-amber-600"
                  >请联系管理员处理</span
                >
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <nav class="flex items-center justify-end gap-3" aria-label="订单分页">
        <span class="text-sm">共 {{ total }} 条，第 {{ page }} 页</span>
        <button class="btn btn-secondary" :disabled="loading || page <= 1" @click="changePage(-1)">上一页</button>
        <button class="btn btn-secondary" :disabled="loading || page * pageSize >= total" @click="changePage(1)">下一页</button>
      </nav>
      <BaseDialog :show="!!delivery" title="商品交付" @close="delivery = null"
        ><template v-if="delivery"
          ><div v-if="delivery.kind === 'card'" class="secret">
            {{ delivery.content }}
          </div>
          <div
            v-else-if="delivery.kind === 'account'"
            class="space-y-2 text-sm"
          >
            <p>用户名：{{ delivery.username }}</p>
            <p>密码：{{ delivery.password }}</p>
            <p v-if="delivery.notes">备注：{{ delivery.notes }}</p>
          </div>
          <div v-else>
            <p>{{ delivery.filename }}</p>
            <button class="btn btn-primary mt-4" :disabled="downloading" @click="downloadDelivery">
              下载文件
            </button>
          </div></template
        ></BaseDialog
      >
    </main></AppLayout
  >
</template>
<script setup lang="ts">
import { onMounted, ref } from "vue";
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import StorePaymentLauncher from "@/components/store/StorePaymentLauncher.vue";
import { digitalStoreAPI } from "@/api/digitalStore";
import { paymentAPI } from "@/api/payment";
import type { StoreDelivery, StoreOrder } from "@/types/digitalStore";
const orders = ref<StoreOrder[]>([]);
const loading = ref(false);
const error = ref("");
const busy = ref<number | null>(null);
const downloading = ref(false);
const page = ref(1);
const pageSize = 20;
const total = ref(0);
const launcher = ref<InstanceType<typeof StorePaymentLauncher> | null>(null);
const delivery = ref<StoreDelivery | null>(null);
const deliveryId = ref<number | null>(null);
const cents = (v: number) => (v / 100).toFixed(2);
const date = (v: string) => (v ? new Date(v).toLocaleString() : "");
const kindLabel = (value: string) => (({ card: '卡密', account: '账号', file: '文件' } as Record<string, string>)[value] || value);
const paymentLabel = (value: string) => (({ PENDING: '待付款', PAID: '已付款', RECHARGING: '交付处理中', COMPLETED: '已完成', CANCELLED: '已取消', EXPIRED: '已过期', FAILED: '处理异常' } as Record<string, string>)[value] || value);
const deliveryLabel = (value: string) => (({ reserved: '等待付款或交付', delivered: '已交付', needs_attention: '已付款，待管理员处理', released: '库存已释放' } as Record<string, string>)[value] || value);
async function load() {
  if (loading.value) return;
  loading.value = true;
  error.value = "";
  try {
    const result = (await digitalStoreAPI.getOrders({ page: page.value, page_size: pageSize })).data;
    orders.value = result.items || [];
    total.value = result.total;
  } catch {
    error.value = "订单加载失败，请刷新重试。";
  } finally {
    loading.value = false;
  }
}
async function changePage(delta: number) {
  page.value += delta;
  await load();
}
async function resume(order: StoreOrder) {
  if (busy.value !== null) return;
  busy.value = order.order_id;
  error.value = "";
  try {
    await launcher.value?.launch((await digitalStoreAPI.resumeOrder(order.order_id)).data);
  } catch {
    error.value = "暂时无法继续付款，请刷新原订单状态，不要重复下单。";
  } finally {
    busy.value = null;
  }
}
async function cancel(order: StoreOrder) {
  if (busy.value !== null || !window.confirm("确定取消此订单？已完成的付款不会因此退款。")) return;
  busy.value = order.order_id;
  error.value = "";
  try {
    await paymentAPI.cancelOrder(order.order_id);
    await load();
  } catch {
    error.value = "取消未确认成功，请刷新订单状态。";
  } finally {
    busy.value = null;
  }
}
async function showDelivery(order: StoreOrder) {
  if (busy.value !== null) return;
  busy.value = order.order_id;
  error.value = "";
  try {
    delivery.value = (await digitalStoreAPI.getDelivery(order.order_id)).data;
    deliveryId.value = order.order_id;
  } catch {
    error.value = "交付内容暂时无法读取，请刷新订单后重试。";
  } finally {
    busy.value = null;
  }
}
async function downloadDelivery() {
  if (!deliveryId.value || !delivery.value || downloading.value) return;
  const filename = delivery.value.filename || "download";
  downloading.value = true;
  error.value = "";
  try {
    const blob = (await digitalStoreAPI.download(deliveryId.value)).data;
    const url = URL.createObjectURL(blob);
    try {
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
    } finally {
      URL.revokeObjectURL(url);
    }
  } catch {
    error.value = "文件下载失败，请稍后重试。";
    delivery.value = null;
  } finally {
    downloading.value = false;
  }
}
onMounted(load);
</script>
<style scoped>
.secret {
  overflow-wrap: anywhere;
  border-radius: 0.5rem;
  background: #f8fafc;
  padding: 1rem;
  font-family: ui-monospace, monospace;
}
</style>
