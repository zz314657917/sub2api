<template>
  <AppLayout
    ><main class="service-store-page console-store-page">
      <header class="store-heading">
        <div>
          <h2>数字商店</h2>
          <p>价格与交付内容均由服务端确认；数字商品付款后不支持自助退款。</p>
        </div>
        <RouterLink class="orders-link" to="/store/orders"
          >我的商店订单</RouterLink
        >
      </header>
      <section class="store-toolbar" aria-label="商品筛选">
        <div class="store-tabs" aria-label="商品分类">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            type="button"
            :class="{ active: kind === tab.value }"
            :aria-pressed="kind === tab.value"
            @click="kind = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>
        <label class="store-search"
          ><span>⌕</span
          ><input
            v-model.trim="search"
            type="search"
            placeholder="搜索商品"
            @keyup.enter="page = 1; loadProducts()"
        /></label>
      </section>
      <p v-if="loadError" class="store-error">{{ loadError }}</p>
      <p v-if="visibleError" class="store-error" role="alert">
        {{ visibleError }}
      </p>
      <div v-if="loading" class="store-empty">正在加载商品…</div>
      <div v-else-if="products.length" class="product-grid">
        <article
          v-for="product in products"
          :key="product.id"
          class="product-card"
        >
          <div class="product-cover" :class="`kind-${product.kind}`">
            <strong>{{ productKindIcon(product.kind) }}</strong
            ><small>{{ product.kind.toUpperCase() }}</small>
          </div>
          <div class="product-body">
            <p class="product-category">{{ productKindLabel(product.kind) }}</p>
            <h3>{{ product.name }}</h3>
            <p class="description">{{ product.description }}</p>
            <div class="product-footer">
              <strong>¥{{ formatCents(product.price_cents) }}</strong
              ><span :class="{ out: product.stock_available <= 0 }">{{
                product.stock_available > 0
                  ? product.kind === 'file' ? '付款后下载' : `剩余 ${product.stock_available}`
                  : "暂时缺货"
              }}</span>
            </div>
            <button
              class="product-buy"
              type="button"
              :disabled="product.stock_available <= 0"
              @click="openCheckout(product)"
            >
              购买并付款 <span>→</span>
            </button>
          </div>
        </article>
      </div>
      <div v-else class="store-empty">暂无可购买商品</div>
      <nav class="mt-4 flex items-center justify-end gap-3" aria-label="商品分页">
        <span>共 {{ total }} 件，第 {{ page }} 页</span>
        <button class="btn btn-secondary" :disabled="loading || page <= 1" @click="page--; loadProducts()">上一页</button>
        <button class="btn btn-secondary" :disabled="loading || page * pageSize >= total" @click="page++; loadProducts()">下一页</button>
      </nav>
      <BaseDialog
        :show="!!checkoutProduct"
        title="确认购买"
        width="narrow"
        @close="closeCheckout"
        ><template v-if="checkoutProduct"
          ><div class="checkout-summary">
            <strong>{{ checkoutProduct.name }}</strong
            ><span>{{ productKindLabel(checkoutProduct.kind) }}</span
            ><b>¥{{ formatCents(checkoutProduct.price_cents) }}</b>
          </div>
          <p class="dialog-notice">
            价格由服务端锁定。提交后请等待当前订单结果；网络异常时请重试同一订单，不会自动创建新订单。
          </p>
          <p v-if="terminalOrderId" class="dialog-notice">上次订单 #{{ terminalOrderId }} 已结束。只有点击“再次购买”才会创建新订单。</p>
          <label class="method-label"
            >支付方式<select v-model="paymentType">
              <option
                v-for="method in availableMethods"
                :key="method"
                :value="method"
              >
                {{ method }}
              </option>
            </select></label
          >
          <p v-if="checkoutError" class="store-error">
            {{ checkoutError }}
          </p></template
        ><template #footer
          ><button class="btn btn-secondary" @click="closeCheckout">取消</button
          ><button
            class="btn btn-primary"
            :disabled="creating || !paymentType"
            @click="terminalOrderId ? buyAgain() : createOrder()"
          >
            {{ creating ? "正在处理…" : terminalOrderId ? "再次购买" : "提交订单并付款" }}
          </button></template
        ></BaseDialog
      >
      <StorePaymentLauncher ref="launcher" @refresh="loadProducts" /></main
  ></AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { stripWechatResumeQuery } from './paymentWechatResume';
import AppLayout from "@/components/layout/AppLayout.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import StorePaymentLauncher from "@/components/store/StorePaymentLauncher.vue";
import { useAuthStore } from "@/stores/auth";
import { digitalStoreAPI } from "@/api/digitalStore";
import { paymentAPI } from "@/api/payment";
import {
  normalizeVisibleMethod,
} from "@/components/payment/paymentFlow";
import type { StoreProduct } from "@/types/digitalStore";
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const products = ref<StoreProduct[]>([]);
const loading = ref(false);
const loadError = ref("");
const search = ref("");
const kind = ref("");
const checkoutProduct = ref<StoreProduct | null>(null);
const paymentType = ref("");
const availableMethods = ref<string[]>([]);
const creating = ref(false);
const checkoutError = ref("");
const launcher = ref<InstanceType<typeof StorePaymentLauncher> | null>(null);
const page = ref(1);
const pageSize = 20;
const total = ref(0);
const terminalOrderId = ref<number | null>(null);
const terminalStatuses = new Set(['COMPLETED', 'CANCELLED', 'EXPIRED']);
const visibleError = ref("");
const tabs = [
  { value: "", label: "全部" },
  { value: "card", label: "卡密" },
  { value: "account", label: "账号" },
  { value: "file", label: "文件" },
];
const checkoutKeyStoragePrefix = "store.checkout.idempotency.";
function formatCents(cents: number) {
  return (Math.max(0, cents) / 100).toFixed(2);
}
function productKindLabel(value: string) {
  return (
    ({ card: "卡密", account: "账号", file: "文件" } as Record<string, string>)[
      value
    ] || value
  );
}
function productKindIcon(value: string) {
  return (
    ({ card: "⌘", account: "◉", file: "↓" } as Record<string, string>)[value] ||
    "□"
  );
}
function errorText(error: unknown) {
  return typeof error === "object" && error && "message" in error
    ? String(error.message)
    : "请求未完成，请稍后重试。";
}
function checkoutStorageKey(productId: number) {
  if (!auth.user?.id) throw new Error('登录状态失效，请重新登录。');
  return `${checkoutKeyStoragePrefix}${auth.user.id}.${productId}`;
}
function checkoutKey(productId: number) {
  const key = checkoutStorageKey(productId);
  const stored = localStorage.getItem(key);
  if (stored) return stored;
  const generated = crypto.randomUUID();
  localStorage.setItem(key, generated);
  return generated;
}
function queryValue(key: string) {
  const value = route.query[key];
  return typeof value === "string" ? value : "";
}
async function loadProducts() {
  loading.value = true;
  loadError.value = "";
  try {
    const result =
      (
        await digitalStoreAPI.getProducts({
          search: search.value || undefined,
          kind: kind.value || undefined,
          page: page.value,
          page_size: pageSize,
        })
      ).data;
    products.value = result.items || [];
    total.value = result.total;
    const resumeProductId = Number(queryValue("product_id"));
    if (
      queryValue("wechat_resume_token") &&
      Number.isSafeInteger(resumeProductId)
    ) {
      const product = products.value.find(
        (item) => item.id === resumeProductId,
      );
      if (product) openCheckout(product);
      else loadError.value = "该微信续单上下文无效或商品已不可购买。";
    }
  } catch (error) {
    loadError.value = errorText(error);
  } finally {
    loading.value = false;
  }
}
async function loadMethods() {
  try {
    const methods = (await paymentAPI.getLimits()).data.methods;
    availableMethods.value = Object.entries(methods)
      .filter(
        ([, limit]) =>
          limit.available &&
          (!limit.currency || limit.currency.toUpperCase() === "CNY"),
      )
      .map(([name]) => normalizeVisibleMethod(name) || name);
    paymentType.value =
      queryValue("payment_type") &&
      availableMethods.value.includes(
        normalizeVisibleMethod(queryValue("payment_type")) ||
          queryValue("payment_type"),
      )
        ? normalizeVisibleMethod(queryValue("payment_type")) ||
          queryValue("payment_type")
        : availableMethods.value[0] || "";
    if (!paymentType.value)
      checkoutError.value = "当前没有可用于人民币商品的支付方式。";
  } catch {
    availableMethods.value = [];
    paymentType.value = "";
    checkoutError.value = "暂时无法取得支付方式。";
  }
}
function openCheckout(product: StoreProduct) {
  checkoutProduct.value = product;
  terminalOrderId.value = null;
  checkoutError.value = "";
  void loadMethods();
}
function closeCheckout() {
  if (!creating.value) checkoutProduct.value = null;
}
async function createOrder() {
  const product = checkoutProduct.value;
  if (!product || !paymentType.value || creating.value) return;
  creating.value = true;
  checkoutError.value = "";
  try {
    const idempotencyKey = checkoutKey(product.id);
    const storageKey = checkoutStorageKey(product.id);
    // A signed OAuth continuation takes precedence over an old browser-only
    // recovery pointer. The server binds the token to its original key.
    const previousID = queryValue('wechat_resume_token') ? 0 : Number(localStorage.getItem(`${storageKey}.order`));
    if (Number.isSafeInteger(previousID) && previousID > 0) {
      const previous = (await digitalStoreAPI.getOrder(previousID)).data;
      if (terminalStatuses.has(previous.payment_status)) {
        terminalOrderId.value = previousID;
        return;
      }
      if (previous.payment_status !== 'PENDING') {
        checkoutError.value = '原订单正在处理，请到我的商店订单查看交付结果。';
        return;
      }
      const existing = (await digitalStoreAPI.resumeOrder(previousID)).data;
      checkoutProduct.value = null;
      await launcher.value?.launch(existing, paymentType.value);
      return;
    }
    const result = (
      await digitalStoreAPI.createOrder({
        product_id: product.id,
        idempotency_key: idempotencyKey,
        payment_type: paymentType.value,
        is_mobile: window.matchMedia("(max-width: 640px)").matches,
        payment_source: "store_checkout",
        return_url: `${window.location.origin}/payment/result`,
        wechat_resume_token: queryValue("wechat_resume_token") || undefined,
      })
    ).data;
    if (!result || typeof result !== 'object' ||
      !(result.result_type === 'oauth_required' || (Number.isSafeInteger(result.order_id) && result.order_id > 0))) {
      throw new Error('订单响应异常，请到我的商店订单核对状态后重试，不要重复下单。');
    }
    if (result.order_id > 0) localStorage.setItem(`${storageKey}.order`, String(result.order_id));
    if ('status' in result && terminalStatuses.has(String(result.status || ''))) {
      terminalOrderId.value = result.order_id;
      return;
    }
    checkoutProduct.value = null;
    await launcher.value?.launch(result, paymentType.value);
  } catch (error) {
    visibleError.value = checkoutError.value = errorText(error);
  } finally {
    creating.value = false;
  }
}
async function buyAgain() {
  const product = checkoutProduct.value;
  if (!product || !terminalOrderId.value || creating.value) return;
  creating.value = true;
  checkoutError.value = '';
  try {
    const previous = (await digitalStoreAPI.getOrder(terminalOrderId.value)).data;
    if (!terminalStatuses.has(previous.payment_status)) throw new Error('原订单尚未结束，请刷新订单状态。');
    if (queryValue('wechat_resume_token')) await router.replace({ query: stripWechatResumeQuery(route.query) });
    const key = checkoutStorageKey(product.id);
    localStorage.setItem(key, crypto.randomUUID());
    localStorage.removeItem(`${key}.order`);
    terminalOrderId.value = null;
  } catch (error) {
    checkoutError.value = errorText(error);
    return;
  } finally {
    creating.value = false;
  }
  await createOrder();
}
watch(kind, () => {
  page.value = 1;
  void loadProducts();
});
onMounted(loadProducts);
</script>

<style scoped>
.service-store-page {
  min-height: calc(100vh - 4rem);
  padding: 0 0 2rem;
  color: var(--console-text, #182230);
}
.store-heading {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.25rem 0 1rem;
}
.store-heading h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 650;
}
.store-heading p,
.dialog-notice {
  margin: 0.35rem 0 0;
  color: #78869a;
  font-size: 0.8rem;
  line-height: 1.6;
}
.orders-link {
  border: 1px solid #dce7f8;
  border-radius: 8px;
  padding: 0.55rem 0.75rem;
  color: #2864cf;
  font-size: 0.78rem;
}
.store-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid #e2e8f0;
  padding: 0.75rem 0;
}
.store-tabs {
  display: flex;
  gap: 0.4rem;
  overflow: auto;
}
.store-tabs button {
  border: 0;
  border-radius: 7px;
  padding: 0.5rem 0.7rem;
  color: #64748b;
  background: transparent;
  font: inherit;
  cursor: pointer;
}
.store-tabs button.active {
  color: #1e5dd0;
  background: #e9f1ff;
  font-weight: 650;
}
.store-search {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  width: 220px;
  border: 1px solid #e0e7f0;
  border-radius: 8px;
  padding: 0.45rem 0.65rem;
  color: #94a3b8;
  background: #fff;
}
.store-search input {
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  font: inherit;
  font-size: 0.78rem;
}
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 250px), 1fr));
  gap: 0.95rem;
  padding-top: 1rem;
}
.product-card {
  overflow: hidden;
  border: 1px solid #e3e9f1;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 5px 16px #3a5b8310;
}
.product-cover {
  display: flex;
  min-height: 120px;
  flex-direction: column;
  justify-content: center;
  padding: 1rem;
  color: #fff;
}
.product-cover strong {
  font-size: 2.6rem;
}
.product-cover small {
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.12em;
}
.kind-card {
  background: linear-gradient(135deg, #24c6ad, #1178bd);
}
.kind-account {
  background: linear-gradient(135deg, #8976e8, #3d4caa);
}
.kind-file {
  background: linear-gradient(135deg, #6d9ef1, #3158cb);
}
.product-body {
  padding: 0.9rem 1rem 1rem;
}
.product-category {
  margin: 0;
  color: #8d9aab;
  font-size: 0.67rem;
}
.product-body h3 {
  margin: 0.35rem 0;
  font-size: 0.95rem;
}
.description {
  display: -webkit-box;
  overflow: hidden;
  min-height: 2.5em;
  margin: 0;
  color: #8995a5;
  font-size: 0.75rem;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.product-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.8rem;
}
.product-footer strong {
  color: #2164d6;
  font-size: 1.08rem;
}
.product-footer span {
  border-radius: 4px;
  padding: 0.24rem 0.4rem;
  color: #3b9b75;
  background: #e8f8f1;
  font-size: 0.62rem;
}
.product-footer span.out {
  color: #dc765f;
  background: #fff0eb;
}
.product-buy {
  display: flex;
  justify-content: space-between;
  width: 100%;
  margin-top: 0.75rem;
  border: 1px solid #dce7f8;
  border-radius: 7px;
  padding: 0.58rem 0.7rem;
  color: #2864cf;
  background: #f4f8ff;
  font: inherit;
  font-size: 0.74rem;
  cursor: pointer;
}
.product-buy:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
.store-empty,
.store-error {
  margin-top: 1rem;
  padding: 2rem;
  text-align: center;
  color: #78869a;
}
.store-error {
  border-radius: 8px;
  padding: 0.7rem;
  color: #b45309;
  background: #fff7ed;
  text-align: left;
}
.checkout-summary {
  display: grid;
  gap: 0.4rem;
  border-radius: 8px;
  padding: 0.8rem;
  background: #f4f8ff;
}
.checkout-summary span {
  color: #64748b;
  font-size: 0.8rem;
}
.checkout-summary b {
  color: #2164d6;
  font-size: 1.3rem;
}
.method-label {
  display: grid;
  gap: 0.4rem;
  margin-top: 1rem;
  font-size: 0.85rem;
}
.method-label select {
  border: 1px solid #dce7f8;
  border-radius: 7px;
  padding: 0.55rem;
  background: #fff;
}
.payment-qr {
  text-align: center;
}
.payment-qr img {
  max-width: 220px;
  margin: auto;
}
.payment-qr p {
  color: #64748b;
  font-size: 0.82rem;
}
@media (max-width: 560px) {
  .store-heading,
  .store-toolbar {
    flex-direction: column;
  }
  .store-search {
    width: 100%;
  }
  .product-grid {
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 160px), 1fr));
    gap: 0.65rem;
  }
}
</style>
