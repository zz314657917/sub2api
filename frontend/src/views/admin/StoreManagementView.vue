<template>
  <AppLayout
    ><main class="space-y-4">
      <header class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold">数字商店管理</h2>
          <p class="mt-1 text-sm text-gray-500">
            商品、库存和异常已付款订单管理。
          </p>
        </div>
        <button
          class="btn btn-primary"
          :disabled="uploading"
          @click="openCreate"
        >
          新建商品
        </button>
      </header>
      <p v-if="pageError" class="text-sm text-red-600" role="alert">
        {{ pageError }}
      </p>

      <section class="card overflow-x-auto">
        <div class="flex items-center justify-between gap-3 p-3">
          <h3 class="font-semibold">商品</h3>
          <span v-if="productsLoading" class="text-sm text-gray-500"
            >加载中...</span
          >
        </div>
        <table class="min-w-full text-sm">
          <thead>
            <tr class="border-b text-left text-gray-500">
              <th class="p-3">商品</th>
              <th class="p-3">类型/价格</th>
              <th class="p-3">库存</th>
              <th class="p-3">状态</th>
              <th class="p-3"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in products" :key="item.id" class="border-b">
              <td class="p-3">
                <b>{{ item.name }}</b
                ><small class="block text-gray-500">{{
                  item.description
                }}</small>
              </td>
              <td class="p-3">
                {{ item.kind }} / ¥{{ cents(item.price_cents) }}
              </td>
              <td class="p-3">{{ item.stock_available }}</td>
              <td class="p-3">{{ item.enabled ? "上架" : "下架" }}</td>
              <td class="p-3 whitespace-nowrap">
                <button
                  class="btn btn-secondary btn-sm"
                  :disabled="uploading"
                  @click="edit(item)"
                >
                  编辑</button
                ><button
                  v-if="item.kind !== 'file'"
                  class="btn btn-secondary btn-sm ml-2"
                  :disabled="uploading"
                  @click="openStock(item)"
                >
                  管理库存
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="flex items-center justify-end gap-2 p-3 text-sm">
          <span>共 {{ productTotal }} 件</span
          ><button
            class="btn btn-secondary btn-sm"
            aria-label="商品上一页"
            :disabled="productsLoading || productPage <= 1"
            @click="loadProducts(productPage - 1)"
          >
            上一页</button
          ><span>{{ productPage }}</span
          ><button
            class="btn btn-secondary btn-sm"
            aria-label="商品下一页"
            :disabled="productsLoading || !hasNext(productPage, productTotal)"
            @click="loadProducts(productPage + 1)"
          >
            下一页
          </button>
        </div>
      </section>

      <section class="card p-4">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h3 class="font-semibold">商店订单</h3>
          <form class="flex flex-wrap gap-2" @submit.prevent="loadOrders(1)">
            <select
              v-model="orderFilters.status"
              class="input"
              aria-label="订单状态"
            >
              <option value="">全部状态</option>
              <option value="PENDING">待支付</option>
              <option value="PAID">已支付</option>
              <option value="COMPLETED">已完成</option>
              <option value="FAILED">失败</option></select
            ><input
              v-model.trim="orderFilters.user"
              class="input"
              inputmode="numeric"
              placeholder="用户 ID"
              aria-label="用户 ID"
            /><input
              v-model.trim="orderFilters.product"
              class="input"
              inputmode="numeric"
              placeholder="商品 ID"
              aria-label="商品 ID"
            /><button
              class="btn btn-secondary btn-sm"
              :disabled="ordersLoading"
            >
              筛选
            </button>
          </form>
        </div>
        <p v-if="ordersError" class="mt-3 text-sm text-red-600" role="alert">
          {{ ordersError }}
        </p>
        <div class="mt-3 overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead>
              <tr class="border-b text-left text-gray-500">
                <th class="p-2">订单</th>
                <th class="p-2">用户/商品</th>
                <th class="p-2">支付/交付</th>
                <th class="p-2"></th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="order in orders"
                :key="order.order_id"
                class="border-b"
              >
                <td class="p-2">
                  #{{ order.order_id }} {{ order.product_name }}
                </td>
                <td class="p-2">
                  {{ order.user_id ?? "-" }} / {{ order.product_id ?? "-" }}
                </td>
                <td class="p-2">
                  {{ order.payment_status }} / {{ order.delivery_status }}
                </td>
                <td class="p-2 text-right">
                  <button
                    v-if="order.delivery_status === 'needs_attention'"
                    class="btn btn-secondary btn-sm"
                    :disabled="retryingOrder === order.order_id"
                    @click="retry(order.order_id)"
                  >
                    {{
                      retryingOrder === order.order_id
                        ? "重试中..."
                        : "重试交付"
                    }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="mt-3 flex items-center justify-end gap-2 text-sm">
          <span>共 {{ orderTotal }} 件</span
          ><button
            class="btn btn-secondary btn-sm"
            aria-label="订单上一页"
            :disabled="ordersLoading || orderPage <= 1"
            @click="loadOrders(orderPage - 1)"
          >
            上一页</button
          ><span>{{ orderPage }}</span
          ><button
            class="btn btn-secondary btn-sm"
            aria-label="订单下一页"
            :disabled="ordersLoading || !hasNext(orderPage, orderTotal)"
            @click="loadOrders(orderPage + 1)"
          >
            下一页
          </button>
        </div>
      </section>

      <BaseDialog
        :show="showForm"
        :title="editing ? '编辑商品' : '新建商品'"
        @close="closeForm"
        ><div class="space-y-3">
          <p v-if="formError" class="text-sm text-red-600" role="alert">
            {{ formError }}
          </p>
          <input
            v-model.trim="form.name"
            class="input w-full"
            placeholder="商品名称"
          /><textarea
            v-model.trim="form.description"
            class="input w-full"
            placeholder="商品描述"
          ></textarea
          ><select
            v-model="form.kind"
            class="input w-full"
            :disabled="!!editing || uploading"
            @change="onKindChange"
          >
            <option value="card">卡密</option>
            <option value="account">账号</option>
            <option value="file">文件</option></select
          ><input
            v-model.number="form.price_cents"
            class="input w-full"
            type="number"
            min="1"
            placeholder="价格（分）"
          /><label class="flex gap-2"
            ><input v-model="form.enabled" type="checkbox" />上架</label
          ><template v-if="form.kind === 'file'"
            ><p v-if="form.file_id" class="text-sm text-gray-500">
              已保留文件 #{{ form.file_id }}
            </p>
            <input type="file" :disabled="uploading" @change="uploadFile" />
            <p v-if="uploading" class="text-sm text-gray-500">文件上传中...</p>
            <p v-if="uploadError" class="text-sm text-red-600" role="alert">
              {{ uploadError }}
            </p></template
          >
        </div>
        <template #footer
          ><button
            class="btn btn-primary"
            :disabled="saving || uploading"
            @click="save"
          >
            {{ saving ? "保存中..." : "保存" }}
          </button></template
        ></BaseDialog
      >

      <BaseDialog :show="!!stockProduct" title="库存管理" @close="closeStock"
        ><div class="space-y-3">
          <p class="text-sm text-gray-500">
            库存列表仅显示脱敏状态，不显示库存明文。
          </p>
          <p v-if="stockError" class="text-sm text-red-600" role="alert">
            {{ stockError }}
          </p>
          <table class="min-w-full text-sm">
            <thead>
              <tr class="border-b text-left text-gray-500">
                <th class="p-2">ID</th>
                <th class="p-2">类型</th>
                <th class="p-2">状态</th>
                <th class="p-2">创建时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in stockRows" :key="row.id" class="border-b">
                <td class="p-2">{{ row.id }}</td>
                <td class="p-2">{{ row.kind }}</td>
                <td class="p-2">{{ row.state }}</td>
                <td class="p-2">{{ row.created_at }}</td>
              </tr>
            </tbody>
          </table>
          <div class="flex items-center justify-end gap-2 text-sm">
            <span>共 {{ stockTotal }} 件</span
            ><button
              class="btn btn-secondary btn-sm"
              aria-label="库存上一页"
              :disabled="stockLoading || stockPage <= 1"
              @click="loadStock(stockPage - 1)"
            >
              上一页</button
            ><span>{{ stockPage }}</span
            ><button
              class="btn btn-secondary btn-sm"
              aria-label="库存下一页"
              :disabled="stockLoading || !hasNext(stockPage, stockTotal)"
              @click="loadStock(stockPage + 1)"
            >
              下一页
            </button>
          </div>
          <textarea
            v-model="stockText"
            class="input min-h-40 w-full"
            :placeholder="
              stockProduct?.kind === 'account'
                ? '每行：用户名|密码|备注'
                : '每行一个卡密'
            "
          ></textarea>
          <p v-if="importError" class="text-sm text-red-600" role="alert">
            {{ importError }}
          </p>
        </div>
        <template #footer
          ><button
            class="btn btn-primary"
            :disabled="importing"
            @click="importStock"
          >
            {{ importing ? "导入中..." : "导入" }}
          </button></template
        ></BaseDialog
      >
    </main></AppLayout
  >
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import { digitalStoreAPI } from "@/api/digitalStore";
import type {
  AdminStoreOrder,
  AdminStoreStockItem,
  AdminStoreStockRow,
  StoreProduct,
} from "@/types/digitalStore";

const pageSize = 20;
const maxUploadBytes = 20 * 1024 * 1024;
const products = ref<StoreProduct[]>([]);
const orders = ref<AdminStoreOrder[]>([]);
const stockRows = ref<AdminStoreStockRow[]>([]);
const productPage = ref(1);
const orderPage = ref(1);
const stockPage = ref(1);
const productTotal = ref(0);
const orderTotal = ref(0);
const stockTotal = ref(0);
const productsLoading = ref(false);
const ordersLoading = ref(false);
const stockLoading = ref(false);
const pageError = ref("");
const ordersError = ref("");
const stockError = ref("");
const showForm = ref(false);
const editing = ref<StoreProduct | null>(null);
const saving = ref(false);
const formError = ref("");
const uploading = ref(false);
const uploadError = ref("");
const stockProduct = ref<StoreProduct | null>(null);
const stockText = ref("");
const importing = ref(false);
const importError = ref("");
const retryingOrder = ref<number | null>(null);
const orderFilters = reactive({ status: "", user: "", product: "" });
const form = reactive({
  name: "",
  description: "",
  kind: "card" as StoreProduct["kind"],
  price_cents: 100,
  enabled: true,
  file_id: undefined as number | undefined,
});
const cents = (value: number) => (value / 100).toFixed(2);
const hasNext = (page: number, total: number) => page * pageSize < total;
const messageFor = (error: unknown, fallback: string) =>
  error instanceof Error && error.message ? error.message : fallback;
async function loadProducts(page = productPage.value) {
  productsLoading.value = true;
  pageError.value = "";
  try {
    const { data } = await digitalStoreAPI.adminProducts({
      page,
      page_size: pageSize,
    });
    products.value = data.items || [];
    productTotal.value = data.total || 0;
    productPage.value = data.page || page;
  } catch (error) {
    pageError.value = messageFor(error, "商品加载失败，请稍后重试。");
  } finally {
    productsLoading.value = false;
  }
}
function positiveId(value: string) {
  const id = Number(value);
  return Number.isInteger(id) && id > 0 ? id : undefined;
}
async function loadOrders(page = orderPage.value) {
  ordersLoading.value = true;
  ordersError.value = "";
  try {
    const userId = positiveId(orderFilters.user);
    const productId = positiveId(orderFilters.product);
    const { data } = await digitalStoreAPI.adminOrders({
      page,
      page_size: pageSize,
      ...(orderFilters.status ? { status: orderFilters.status } : {}),
      ...(userId ? { user_id: userId } : {}),
      ...(productId ? { product_id: productId } : {}),
    });
    orders.value = data.items || [];
    orderTotal.value = data.total || 0;
    orderPage.value = data.page || page;
  } catch (error) {
    ordersError.value = messageFor(error, "订单加载失败，请稍后重试。");
  } finally {
    ordersLoading.value = false;
  }
}
async function loadStock(page = stockPage.value) {
  if (!stockProduct.value) return;
  stockLoading.value = true;
  stockError.value = "";
  try {
    const { data } = await digitalStoreAPI.getStock(stockProduct.value.id, {
      page,
      page_size: pageSize,
    });
    stockRows.value = data.items || [];
    stockTotal.value = data.total || 0;
    stockPage.value = data.page || page;
  } catch (error) {
    stockError.value = messageFor(error, "库存加载失败，请稍后重试。");
  } finally {
    stockLoading.value = false;
  }
}
function reset() {
  Object.assign(form, {
    name: "",
    description: "",
    kind: "card",
    price_cents: 100,
    enabled: true,
    file_id: undefined,
  });
  formError.value = "";
  uploadError.value = "";
}
function openCreate() {
  if (uploading.value) return;
  editing.value = null;
  reset();
  showForm.value = true;
}
function edit(item: StoreProduct) {
  if (uploading.value) return;
  editing.value = item;
  Object.assign(form, {
    name: item.name,
    description: item.description,
    kind: item.kind,
    price_cents: item.price_cents,
    enabled: item.enabled,
    file_id: item.file_id,
  });
  formError.value = "";
  uploadError.value = "";
  showForm.value = true;
}
function closeForm() {
  if (!uploading.value && !saving.value) showForm.value = false;
}
function onKindChange() {
  if (form.kind !== "file") form.file_id = undefined;
}
async function uploadFile(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  uploadError.value = "";
  if (!file) return;
  if (file.size > maxUploadBytes) {
    uploadError.value = "文件不能超过 20 MiB。";
    return;
  }
  uploading.value = true;
  try {
    form.file_id = (await digitalStoreAPI.uploadFile(file)).data.id;
  } catch (error) {
    uploadError.value = messageFor(error, "文件上传失败，请重新选择文件。");
  } finally {
    uploading.value = false;
  }
}
async function save() {
  if (saving.value || uploading.value) return;
  formError.value = "";
  if (form.kind === "file" && !form.file_id) {
    formError.value = "文件商品必须先上传文件。";
    return;
  }
  saving.value = true;
  const data = {
    ...form,
    ...(form.kind === "file" ? {} : { file_id: undefined }),
  };
  try {
    if (editing.value)
      await digitalStoreAPI.updateProduct(editing.value.id, data);
    else await digitalStoreAPI.createProduct(data);
    showForm.value = false;
    await loadProducts(1);
  } catch (error) {
    formError.value = messageFor(error, "商品保存失败，请检查后重试。");
  } finally {
    saving.value = false;
  }
}
async function openStock(item: StoreProduct) {
  if (uploading.value) return;
  stockProduct.value = item;
  stockText.value = "";
  stockRows.value = [];
  stockTotal.value = 0;
  stockPage.value = 1;
  importError.value = "";
  await loadStock(1);
}
function closeStock() {
  if (!importing.value) stockProduct.value = null;
}
function parseStock(): AdminStoreStockItem[] | null {
  const lines = stockText.value
    .split(/\r?\n/)
    .filter((line) => line.trim().length > 0);
  if (!lines.length || lines.length > 500) {
    importError.value = "请提供 1 至 500 条库存。";
    return null;
  }
  if (stockProduct.value?.kind !== "account")
    return lines.map((content) => ({ content }));
  const items: AdminStoreStockItem[] = [];
  for (const line of lines) {
    const [username, password, ...noteParts] = line.split("|");
    if (
      username === undefined ||
      password === undefined ||
      !username.trim() ||
      !password.trim()
    ) {
      importError.value =
        "账号库存每行必须是“用户名|密码|备注”；不会自动删除密码中的空格。";
      return null;
    }
    items.push({ username, password, notes: noteParts.join("|") });
  }
  return items;
}
async function importStock() {
  if (!stockProduct.value || importing.value) return;
  importError.value = "";
  const items = parseStock();
  if (!items) return;
  importing.value = true;
  try {
    await digitalStoreAPI.importStock(stockProduct.value.id, items);
    stockText.value = "";
    await Promise.all([loadStock(1), loadProducts(productPage.value)]);
  } catch (error) {
    importError.value = `${messageFor(error, "库存导入请求失败。")} 请求可能已被服务端处理，请先刷新并核实库存后再决定是否重试。`;
  } finally {
    importing.value = false;
  }
}
async function retry(id: number) {
  if (retryingOrder.value !== null) return;
  ordersError.value = "";
  retryingOrder.value = id;
  try {
    await digitalStoreAPI.retryOrder(id);
    await loadOrders(orderPage.value);
  } catch (error) {
    ordersError.value = messageFor(error, "订单重试失败，请稍后重试。");
  } finally {
    retryingOrder.value = null;
  }
}
onMounted(() => {
  void Promise.all([loadProducts(), loadOrders()]);
});
</script>
