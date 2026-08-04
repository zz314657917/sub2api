<template>
  <AppLayout v-if="showLegacyGroupBuy">
    <div class="group-buy-page -m-4 min-h-[calc(100vh-4rem)] md:-m-[1.35rem] lg:-m-[1.6rem]">
      <div class="mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
        <header class="group-buy-header group-buy-toolbar">
          <div class="group-buy-header-copy">
            <p class="group-buy-eyebrow">平台托管容量拼团</p>
            <p>{{ groupBuyDescription }}</p>
          </div>
          <button type="button" class="group-buy-refresh" :disabled="loading" aria-label="刷新" @click="refreshAll">
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          </button>
        </header>

        <div v-if="errorMessage" class="group-buy-alert">
          <Icon name="exclamationTriangle" size="sm" />
          <span>{{ errorMessage }}</span>
        </div>

        <nav class="group-buy-tabs" :aria-label="`${groupBuyProductName} tabs`">
          <button v-for="tab in tabs" :key="tab.id" type="button" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">
            {{ tab.label }}
          </button>
        </nav>

        <div v-if="activeTab === 'hall'" class="group-buy-shell">
          <main class="group-buy-main">
            <section v-if="loading && plans.length === 0" class="group-buy-empty">
              <span class="group-buy-spinner"></span>
            </section>
            <section v-else-if="plans.length === 0" class="group-buy-empty">
              <Icon name="gift" size="xl" />
              <p>暂无可参与的 {{ groupBuyProductName }} 拼团</p>
            </section>
            <section v-else class="group-buy-plan-grid">
              <article v-for="plan in plans" :key="plan.id" class="group-buy-plan">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="group-buy-plan-kicker">总 {{ totalShares(plan) }} 份 · {{ launchModeLabel(plan.launch_mode) }}</p>
                    <h2>{{ plan.title }}</h2>
                    <p class="group-buy-plan-desc">{{ plan.description || `${groupBuyProductName} 平台托管容量份额，满份后成团。` }}</p>
                  </div>
                  <span class="group-buy-status" :class="roundStatusClass(plan.current_round?.status, plan)">
                    {{ roundStatusLabel(plan.current_round?.status, plan) }}
                  </span>
                </div>

                <div class="group-buy-price-row">
                  <span>{{ priceDisplay(plan) }}</span>
                  <small v-if="!plan.price_label">/ 份</small>
                </div>

                <div class="group-buy-quota">{{ quotaPerShareLabel(plan) }}</div>

                <div class="group-buy-progress">
                  <div class="flex items-center justify-between text-xs">
                    <span>已付 {{ paidShares(plan) }} / {{ totalShares(plan) }} 份</span>
                    <span>{{ canJoin(plan) ? `剩余 ${availableShares(plan)} 份` : unavailableReason(plan) }}</span>
                  </div>
                  <div class="group-buy-progress-track">
                    <div class="group-buy-progress-fill" :style="{ width: `${roundProgress(plan)}%` }"></div>
                  </div>
                </div>

                <div class="group-buy-meta">
                  <span>{{ plan.validity_days }} 天权益</span>
                  <span>{{ timeoutLabel(plan.timeout_minutes) }} 截止</span>
                  <span>最多 {{ plan.max_shares_per_user || 10 }} 份</span>
                  <span>{{ refundModeLabel(plan.refund_mode) }}</span>
                </div>

                <button type="button" class="group-buy-primary" :disabled="submitting || !canJoin(plan)" @click="openOrderDialog(plan)">
                  <Icon name="creditCard" size="sm" />
                  {{ canJoin(plan) ? '选择份额并付款' : '暂不可参与' }}
                </button>
              </article>
            </section>
          </main>

          <aside class="group-buy-side">
            <section class="group-buy-panel">
              <div class="group-buy-panel-title">
                <Icon name="sync" size="sm" />
                拼团动态
              </div>
              <div v-if="activity.length === 0" class="group-buy-side-empty">暂无动态</div>
              <ol v-else class="group-buy-activity">
                <li v-for="event in activity" :key="event.id">
                  <span></span>
                  <div>
                    <p>{{ event.message }}</p>
                    <time>{{ formatDateTime(event.created_at) }}</time>
                  </div>
                </li>
              </ol>
            </section>

            <section class="group-buy-panel">
              <div class="group-buy-panel-title">
                <Icon name="shield" size="sm" />
                使用边界
              </div>
              <ul class="group-buy-rules">
                <li>只使用自己的平台 API Key。</li>
                <li>满份成团后按有效份额开通权益。</li>
              </ul>
            </section>
          </aside>
        </div>

        <section v-else-if="activeTab === 'binding'" class="group-buy-panel group-buy-wide-panel">
          <div class="group-buy-panel-head">
            <div>
              <h2>使用与绑定</h2>
              <p>满份成团后，系统会汇总你的有效份额并开通对应权益档位。绑定记录会保留，下次成团后自动恢复。</p>
            </div>
            <button type="button" class="group-buy-secondary" :disabled="seatsLoading" aria-label="刷新份额" @click="fetchSeats">
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': seatsLoading }" />
            </button>
          </div>

          <div class="group-buy-entitlement">
            <div>
              <p class="group-buy-caption">当前有效份额</p>
              <strong>{{ entitlement?.active_share_count ?? 0 }} 份</strong>
            </div>
            <div>
              <p class="group-buy-caption">当前权益档位</p>
              <strong>{{ entitlement?.entitlement_label || (entitlement?.status === 'active' ? `${entitlement.active_share_count} 份权益` : '未生效') }}</strong>
            </div>
            <div>
              <p class="group-buy-caption">绑定 API Key</p>
              <strong>{{ entitlement?.bound_api_key_id ? `#${entitlement.bound_api_key_id}` : '未绑定' }}</strong>
            </div>
            <div>
              <p class="group-buy-caption">权益到期</p>
              <strong>{{ entitlement?.expires_at ? formatDateTime(entitlement.expires_at) : '-' }}</strong>
            </div>
          </div>

          <p v-if="entitlement && entitlement.status !== 'active' && entitlement.bound_api_key_id" class="group-buy-inline-note">
            当前有效份额为 0，{{ groupBuyProductName }} 权益已失效；绑定记录保留，后续成团会自动恢复权益。
          </p>

          <div v-if="seats.length === 0" class="group-buy-empty group-buy-empty-soft">
            <Icon name="key" size="xl" />
            <p>暂无份额批次</p>
          </div>
          <div v-else class="group-buy-seat-list">
            <article v-for="seat in seats" :key="seat.id" class="group-buy-seat">
              <div>
                <div class="flex flex-wrap items-center gap-2">
                  <h3>{{ seat.plan?.title || `份额批次 #${seat.id}` }}</h3>
                  <span class="group-buy-status" :class="seatStatusClass(seat.status)">{{ seatStatusLabel(seat.status) }}</span>
                </div>
                <p>{{ seat.share_count || 1 }} 份 · {{ seat.plan?.quota_per_share_label || seat.plan?.quota_label || '单份额度待填写' }}</p>
                <p v-if="seat.round" class="group-buy-caption">团次 #{{ seat.round.id }} · {{ roundStatusLabel(seat.round.status, seat.plan) }}</p>
                <p class="group-buy-caption">到期：{{ seat.expires_at ? formatDateTime(seat.expires_at) : '-' }}</p>
              </div>
              <div class="group-buy-bind-box">
                <input v-model.number="bindKeyIds[seat.id]" type="number" min="1" placeholder="API Key ID" :disabled="seat.status !== 'active' || bindingSeatId === seat.id" />
                <button type="button" class="group-buy-primary" :disabled="seat.status !== 'active' || !bindKeyIds[seat.id] || bindingSeatId === seat.id" @click="bindKey(seat)">
                  {{ bindingSeatId === seat.id ? '绑定中' : '绑定 Key' }}
                </button>
              </div>
            </article>
          </div>
        </section>

        <section v-else class="group-buy-panel group-buy-wide-panel">
          <div class="group-buy-panel-head">
            <div>
              <h2>拼团订单</h2>
              <p>这里仅展示 {{ groupBuyProductName }} 订单及份额状态，不替代普通订单列表。</p>
            </div>
            <button type="button" class="group-buy-secondary" :disabled="ordersLoading" aria-label="刷新订单" @click="fetchOrders">
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': ordersLoading }" />
            </button>
          </div>
          <div v-if="orders.length === 0" class="group-buy-empty group-buy-empty-soft">
            <Icon name="clipboard" size="xl" />
            <p>暂无拼团订单</p>
          </div>
          <div v-else class="group-buy-order-table">
            <div class="group-buy-order-row group-buy-order-head">
              <span>订单</span>
              <span>金额</span>
              <span>支付方式</span>
              <span>状态</span>
              <span>创建时间</span>
            </div>
            <div v-for="order in orders" :key="order.id" class="group-buy-order-row">
              <span>#{{ order.id }}</span>
              <span>{{ formatMoney(order.pay_amount || order.amount, order.currency) }}</span>
              <span>{{ paymentMethodLabel(order.payment_type) }}</span>
              <span><b :class="orderStatusClass(order.status)">{{ orderStatusLabel(order.status) }}</b></span>
              <span>{{ formatDateTime(order.created_at) }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="orderDialogOpen && selectedPlan" class="group-buy-modal-backdrop" @click.self="closeOrderDialog">
          <div class="group-buy-modal">
            <button type="button" class="group-buy-modal-close" aria-label="关闭" @click="closeOrderDialog">
              <Icon name="x" size="sm" />
            </button>
            <template v-if="paymentPhase === 'paying'">
              <PaymentStatusPanel
                :order-id="paymentState.orderId"
                :qr-code="paymentState.qrCode"
                :expires-at="paymentState.expiresAt"
                :payment-type="paymentState.paymentType"
                :pay-url="paymentState.payUrl"
                order-type="group_buy"
                :currency="paymentState.currency || 'CNY'"
                @done="onPaymentDone"
                @success="onPaymentSuccess"
                @settled="onPaymentSettled"
              />
            </template>
            <template v-else>
              <p class="group-buy-eyebrow">确认份额</p>
              <h2>{{ selectedPlan.title }}</h2>
              <p class="group-buy-modal-desc">{{ quotaPerShareLabel(selectedPlan) }}</p>

              <div class="group-buy-stepper">
                <button type="button" :disabled="selectedShareCount <= 1" @click="setSelectedShareCount(selectedShareCount - 1)">-</button>
                <div>
                  <strong>{{ selectedShareCount }} 份</strong>
                  <span>最多可选 {{ selectedPlanMaxShareCount }} 份</span>
                </div>
                <button type="button" :disabled="selectedShareCount >= selectedPlanMaxShareCount" @click="setSelectedShareCount(selectedShareCount + 1)">+</button>
              </div>

              <div class="group-buy-modal-summary">
                <div><span>单份价格</span><b>{{ priceDisplay(selectedPlan) }}</b></div>
                <div><span>份额数量</span><b>{{ selectedShareCount }} 份</b></div>
                <div><span>预计总额度</span><b>{{ estimatedQuotaLabel(selectedPlan, selectedShareCount) }}</b></div>
                <div><span>应付总价</span><b>{{ formatMoney(selectedOrderAmount) }}</b></div>
                <div><span>付款方式</span><b>{{ selectedPaymentLabel }}</b></div>
                <div><span>权益有效期</span><b>{{ selectedPlan.validity_days }} 天</b></div>
                <div><span>成团截止</span><b>{{ timeoutLabel(selectedPlan.timeout_minutes) }}</b></div>
              </div>

              <div class="group-buy-methods">
                <button v-for="method in paymentMethods" :key="method" type="button" :class="{ active: selectedPaymentMethod === method }" @click="selectedPaymentMethod = method">
                  {{ paymentMethodLabel(method) }}
                </button>
              </div>

              <label class="group-buy-agreement">
                <input v-model="agreementAccepted" type="checkbox" />
                <span>{{ selectedPlan.agreement_text || defaultAgreementText }}</span>
              </label>

              <button type="button" class="group-buy-primary group-buy-modal-submit" :disabled="submitting || !agreementAccepted" @click="submitOrder">
                <span v-if="submitting" class="group-buy-spinner group-buy-spinner-small"></span>
                <Icon v-else name="shield" size="sm" />
                {{ submitting ? '处理中' : '确认份额并付款' }}
              </button>
            </template>
          </div>
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
  <PixelCafePage v-else />
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import PixelCafePage from '@/features/pixelCafe/PixelCafePage.vue'
import { groupBuyAPI } from '@/api/groupBuy'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { normalizePaymentCurrency, formatPaymentAmount } from '@/components/payment/currency'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import { resolveGroupBuyProductName } from '@/utils/groupBuyProduct'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  normalizeVisibleMethod,
  writePaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import type { GroupBuyEntitlement, GroupBuyEvent, GroupBuyMySeatsResponse, GroupBuyPlan, GroupBuySeat, PaymentOrderLite } from '@/types/groupBuy'
import type { CreateOrderResult } from '@/types/payment'

type TabID = 'hall' | 'binding' | 'orders'
type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const appStore = useAppStore()
const route = useRoute()
const router = useRouter()

const pixelCafeEnabled = computed(() => appStore.cachedPublicSettings?.pixel_cafe_enabled === true)
const showLegacyGroupBuy = computed(() => !pixelCafeEnabled.value || route.query.legacy === '1')

const tabs: { id: TabID; label: string }[] = [
  { id: 'hall', label: '拼团大厅' },
  { id: 'binding', label: '使用与绑定' },
  { id: 'orders', label: '拼团订单' },
]

const activeTab = ref<TabID>('hall')
const plans = ref<GroupBuyPlan[]>([])
const activity = ref<GroupBuyEvent[]>([])
const entitlement = ref<GroupBuyEntitlement | null>(null)
const seats = ref<GroupBuySeat[]>([])
const orders = ref<PaymentOrderLite[]>([])
const loading = ref(false)
const seatsLoading = ref(false)
const ordersLoading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const orderDialogOpen = ref(false)
const selectedPlan = ref<GroupBuyPlan | null>(null)
const selectedShareCount = ref(1)
const selectedPaymentMethod = ref('alipay')
const agreementAccepted = ref(false)
const paymentPhase = ref<'confirm' | 'paying'>('confirm')
const bindingSeatId = ref<number | null>(null)
const bindKeyIds = reactive<Record<number, number | undefined>>({})
const paymentMethods = ['alipay', 'wxpay', 'stripe', 'airwallex']
const groupBuyProductName = computed(() => resolveGroupBuyProductName(appStore.cachedPublicSettings))
const defaultGroupBuyDescription = computed(() => `按份额拼团，满份后开通 ${groupBuyProductName.value} 权益；使用自己的平台 API Key。`)
const defaultAgreementText = computed(() => `我理解 ${groupBuyProductName.value} 为平台托管容量份额权益，不是官方 OpenAI Pro 子账号，不共享任何上游账号或官方 API Key。`)
const groupBuyDescription = computed(() => appStore.cachedPublicSettings?.group_buy_description?.trim() || defaultGroupBuyDescription.value)

const paymentState = ref<PaymentRecoverySnapshot>({
  orderId: 0,
  amount: 0,
  qrCode: '',
  expiresAt: '',
  paymentType: '',
  payUrl: '',
  outTradeNo: '',
  clientSecret: '',
  intentId: '',
  currency: '',
  countryCode: '',
  paymentEnv: '',
  payAmount: 0,
  orderType: 'group_buy',
  paymentMode: '',
  resumeToken: '',
  createdAt: 0,
})

const selectedPaymentLabel = computed(() => paymentMethodLabel(selectedPaymentMethod.value))
const selectedPlanMaxShareCount = computed(() => selectedPlan.value ? maxPurchasableShares(selectedPlan.value) : 1)
const selectedOrderAmount = computed(() => selectedPlan.value ? pricePerShare(selectedPlan.value) * selectedShareCount.value : 0)

async function refreshAll() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [plansRes, activityRes] = await Promise.all([
      groupBuyAPI.listPlans(),
      groupBuyAPI.activity(20),
      fetchSeats(),
      fetchOrders(),
    ])
    plans.value = plansRes.data || []
    activity.value = activityRes.data || []
  } catch (err: unknown) {
    errorMessage.value = extractApiErrorMessage(err, `${groupBuyProductName.value} 信息加载失败`)
  } finally {
    loading.value = false
  }
}

async function fetchSeats() {
  seatsLoading.value = true
  try {
    const res = await groupBuyAPI.mySeats()
    const data = res.data as GroupBuyMySeatsResponse | GroupBuySeat[] | undefined
    if (Array.isArray(data)) {
      entitlement.value = null
      seats.value = data
    } else {
      entitlement.value = data?.entitlement || null
      seats.value = data?.seats || []
    }
  } finally {
    seatsLoading.value = false
  }
}

async function fetchOrders() {
  ordersLoading.value = true
  try {
    const res = await groupBuyAPI.myOrders({ page: 1, page_size: 20 })
    orders.value = res.data?.items || []
  } finally {
    ordersLoading.value = false
  }
}

function openOrderDialog(plan: GroupBuyPlan) {
  selectedPlan.value = plan
  selectedShareCount.value = 1
  agreementAccepted.value = false
  paymentPhase.value = 'confirm'
  orderDialogOpen.value = true
}

function closeOrderDialog() {
  if (submitting.value) return
  orderDialogOpen.value = false
}

function setSelectedShareCount(value: number) {
  selectedShareCount.value = Math.min(Math.max(1, Math.floor(value || 1)), selectedPlanMaxShareCount.value)
}

async function submitOrder() {
  if (!selectedPlan.value || !agreementAccepted.value || submitting.value) return
  submitting.value = true
  errorMessage.value = ''
  const visibleMethod = normalizeVisibleMethod(selectedPaymentMethod.value) || selectedPaymentMethod.value
  try {
    const result = (await groupBuyAPI.createOrder({
      plan_id: selectedPlan.value.id,
      share_count: selectedShareCount.value,
      payment_type: visibleMethod,
      return_url: `${window.location.origin}/payment/result`,
      payment_source: visibleMethod === 'wxpay' && isWechatBrowser() ? 'wechat_in_app_resume' : 'hosted_redirect',
      is_mobile: isMobileDevice(),
    })).data as CreateOrderResult & { resume_token?: string }

    await Promise.allSettled([fetchSeats(), fetchOrders(), refreshActivity()])
    const stripeMethod = visibleMethod === 'stripe' ? '' : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: 'group_buy',
      isMobile: isMobileDevice(),
      isWechatBrowser: isWechatBrowser(),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = decision.oauth.authorize_url
      return
    }
    if (decision.kind === 'unhandled') {
      appStore.showError('当前支付方式暂不可用，请换一种方式。')
      return
    }
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    writePaymentRecoverySnapshot(window.localStorage, decision.recovery, PAYMENT_RECOVERY_STORAGE_KEY)
    if (decision.kind === 'stripe_popup' && decision.paymentState.payUrl) {
      openPopup(decision.paymentState.payUrl)
      return
    }
    if ((decision.kind === 'stripe_route' || decision.kind === 'airwallex_route') && decision.paymentState.payUrl) {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
      } else {
        openPopup(decision.paymentState.payUrl)
      }
    }
  } catch (err: unknown) {
    const message = extractApiErrorMessage(err, `创建 ${groupBuyProductName.value} 订单失败`)
    errorMessage.value = message
    appStore.showError(message)
  } finally {
    submitting.value = false
  }
}

async function refreshActivity() {
  const res = await groupBuyAPI.activity(20)
  activity.value = res.data || []
}

function onPaymentDone() {
  orderDialogOpen.value = false
  paymentPhase.value = 'confirm'
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
  void refreshAll()
}

function onPaymentSuccess() {
  appStore.showSuccess('付款成功，份额状态已更新。')
}

function onPaymentSettled(outcome: PaymentOutcome) {
  if (outcome !== 'success') {
    void refreshAll()
  }
}

async function bindKey(seat: GroupBuySeat) {
  const apiKeyId = Number(bindKeyIds[seat.id])
  if (!apiKeyId || bindingSeatId.value) return
  bindingSeatId.value = seat.id
  try {
    await groupBuyAPI.bindKey(seat.id, { api_key_id: apiKeyId })
    appStore.showSuccess('绑定成功')
    await fetchSeats()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '绑定失败'))
  } finally {
    bindingSeatId.value = null
  }
}

function openPopup(url: string) {
  const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
  if (!win || win.closed) {
    window.location.href = url
  }
}

function isWechatBrowser(): boolean {
  return /MicroMessenger/i.test(window.navigator.userAgent)
}

function totalShares(plan: GroupBuyPlan): number {
  return Number(plan.total_shares || plan.seat_count || 10)
}

function paidShares(plan: GroupBuyPlan): number {
  return Number(plan.current_round?.paid_shares ?? plan.current_round?.paid_seats ?? 0)
}

function availableShares(plan: GroupBuyPlan): number {
  return Number(plan.current_round?.available_shares ?? plan.current_round?.available_seats ?? totalShares(plan))
}

function pricePerShare(plan: GroupBuyPlan): number {
  return Number(plan.price_per_share || plan.price_per_seat || 0)
}

function priceDisplay(plan: GroupBuyPlan): string {
  return plan.price_label || formatMoney(pricePerShare(plan))
}

function quotaPerShareLabel(plan: GroupBuyPlan): string {
  return plan.quota_per_share_label || plan.quota_label || '单份月额度待填写'
}

function estimatedQuotaLabel(plan: GroupBuyPlan, shares: number): string {
  const label = quotaPerShareLabel(plan)
  const match = label.match(/(\d+(?:\.\d+)?)/)
  if (!match) return `${shares} 份权益`
  const value = Number(match[1])
  if (!Number.isFinite(value)) return `${shares} 份权益`
  return label.replace(match[1], String(Number((value * shares).toFixed(2))))
}

function maxPurchasableShares(plan: GroupBuyPlan): number {
  const openAvailable = canJoin(plan) ? availableShares(plan) : 0
  const maxUser = Number(plan.max_shares_per_user || 10)
  return Math.max(1, Math.min(openAvailable || maxUser, maxUser, 10))
}

function canJoin(plan: GroupBuyPlan): boolean {
  const round = plan.current_round
  if (!round) return plan.launch_mode === 'auto'
  return round.status === 'open' && availableShares(plan) > 0
}

function unavailableReason(plan: GroupBuyPlan): string {
  if (!plan.current_round && plan.launch_mode === 'manual') return '待开团'
  if (plan.current_round?.status === 'open' && availableShares(plan) <= 0) return '已满份'
  return roundStatusLabel(plan.current_round?.status, plan)
}

function roundProgress(plan: GroupBuyPlan): number {
  const total = Math.max(1, totalShares(plan))
  const paid = Math.max(0, paidShares(plan))
  return Math.min(100, Math.round((paid / total) * 100))
}

function formatMoney(value: number, currency = 'CNY'): string {
  return formatPaymentAmount(Number(value || 0), normalizePaymentCurrency(currency))
}

function timeoutLabel(minutes: number): string {
  if (minutes >= 1440 && minutes % 1440 === 0) return `${minutes / 1440} 天`
  if (minutes >= 60 && minutes % 60 === 0) return `${minutes / 60} 小时`
  return `${minutes} 分钟`
}

function refundModeLabel(mode: string): string {
  return mode === 'provider_refund' ? '原路退款' : '退回余额'
}

function launchModeLabel(mode?: string): string {
  return mode === 'manual' ? '手动开团' : '自动续开'
}

function paymentMethodLabel(method: string): string {
  const m = method.toLowerCase()
  if (m.includes('wxpay')) return '微信支付'
  if (m.includes('alipay')) return '支付宝'
  if (m === 'stripe') return 'Stripe'
  if (m === 'airwallex') return 'Airwallex'
  return method || '-'
}

function roundStatusLabel(status?: string, plan?: GroupBuyPlan | null): string {
  switch (status) {
    case 'open': return '拼团中'
    case 'activating': return '成团中'
    case 'active': return '已成团'
    case 'failed': return '未满份'
    case 'cancelled': return '已关闭'
    default: return plan?.launch_mode === 'manual' ? '待开团' : '自动成团'
  }
}

function seatStatusLabel(status: string): string {
  switch (status) {
    case 'locked': return '待付款'
    case 'paid': return '待满份'
    case 'active': return '已开通'
    case 'refund_pending': return '待退款'
    case 'refund_processing': return '退款中'
    case 'refunded': return '已退款'
    case 'released': return '已释放'
    case 'cancelled': return '已取消'
    default: return status
  }
}

function orderStatusLabel(status: string): string {
  switch (status) {
    case 'PENDING': return '待支付'
    case 'PAID': return '已支付'
    case 'RECHARGING': return '处理中'
    case 'COMPLETED': return '已完成'
    case 'EXPIRED': return '已过期'
    case 'CANCELLED': return '已取消'
    case 'FAILED': return '失败'
    case 'REFUND_PENDING': return '待退款'
    case 'REFUNDED': return '已退款'
    default: return status
  }
}

function roundStatusClass(status?: string, plan?: GroupBuyPlan | null): string {
  if (status === 'active') return 'ok'
  if (status === 'failed' || status === 'cancelled') return 'bad'
  if (status === 'activating' || (!status && plan?.launch_mode === 'manual')) return 'warn'
  return 'idle'
}

function seatStatusClass(status: string): string {
  if (status === 'active' || status === 'refunded') return 'ok'
  if (status === 'refund_pending' || status === 'paid') return 'warn'
  if (status === 'released' || status === 'cancelled') return 'bad'
  return 'idle'
}

function orderStatusClass(status: string): string {
  if (status === 'COMPLETED' || status === 'PAID') return 'text-emerald-700'
  if (status === 'PENDING' || status === 'RECHARGING' || status === 'REFUND_PENDING') return 'text-amber-700'
  if (status === 'FAILED' || status === 'EXPIRED' || status === 'CANCELLED') return 'text-red-700'
  return 'text-stone-700'
}

function formatDateTime(value?: string): string {
  if (!value) return '-'
  return new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

onMounted(() => {
  if (!showLegacyGroupBuy.value) return
  void refreshAll()
})
</script>

<style scoped>
.group-buy-page,
.group-buy-modal-backdrop {
  --gb-bg: var(--console-bg, #faf9f5);
  --gb-surface: var(--console-surface, #faf9f5);
  --gb-surface-soft: var(--console-surface-soft, rgba(245, 240, 232, 0.76));
  --gb-surface-hover: var(--console-surface-hover, #fffaf5);
  --gb-input: var(--console-input, rgba(250, 249, 245, 0.94));
  --gb-text: var(--console-text, #141413);
  --gb-muted: var(--console-muted, #6c6a64);
  --gb-muted-strong: var(--console-muted-strong, #3d3d3a);
  --gb-border: var(--console-border, rgba(216, 206, 194, 0.68));
  --gb-border-strong: var(--console-border-strong, rgba(160, 153, 144, 0.38));
  --gb-accent: var(--console-accent, #cc785c);
  --gb-accent-strong: #a9583e;
  --gb-ring: var(--console-ring, rgba(204, 120, 92, 0.16));
  --gb-shadow: var(--console-shadow, 0 12px 28px rgba(75, 52, 40, 0.045));
  --gb-shadow-lg: var(--console-shadow-lg, 0 18px 44px rgba(75, 52, 40, 0.07));
}

.group-buy-page {
  background: transparent;
  color: var(--gb-text);
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.group-buy-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid var(--gb-border);
  padding-bottom: 0.95rem;
}

.group-buy-toolbar {
  min-height: 2.75rem;
}

.group-buy-header-copy {
  min-width: 0;
  max-width: 44rem;
}

.group-buy-header-copy p {
  margin: 0;
}

.group-buy-header-copy p + p {
  margin-top: 0.35rem;
  color: var(--gb-muted);
  font-size: 0.95rem;
  line-height: 1.55;
}

.group-buy-eyebrow,
.group-buy-plan-kicker {
  color: var(--gb-accent-strong);
  font-size: 0.765rem;
  font-weight: 700;
  letter-spacing: 0;
}

.group-buy-panel h2,
.group-buy-modal h2 {
  margin: 0;
  color: var(--gb-text);
  font-size: clamp(1.55rem, 2.1vw, 2rem);
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.12;
}

.group-buy-refresh,
.group-buy-secondary {
  display: inline-flex;
  min-height: 2.75rem;
  min-width: 2.75rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--gb-border-strong);
  border-radius: 999px;
  background: var(--gb-surface-soft);
  color: var(--gb-muted-strong);
  transition: background-color 0.16s ease, border-color 0.16s ease, color 0.16s ease, transform 0.12s ease;
}

.group-buy-refresh:hover,
.group-buy-secondary:hover {
  border-color: rgba(204, 120, 92, 0.34);
  background: var(--gb-surface-hover);
  color: var(--gb-text);
}

.group-buy-refresh:active,
.group-buy-secondary:active,
.group-buy-tabs button:active,
.group-buy-primary:active,
.group-buy-stepper button:active,
.group-buy-methods button:active {
  transform: scale(0.96);
}

.group-buy-tabs {
  display: inline-flex;
  width: fit-content;
  max-width: 100%;
  gap: 0.25rem;
  overflow-x: auto;
  border: 1px solid var(--gb-border);
  border-radius: 999px;
  background: var(--gb-surface-soft);
  padding: 0.25rem;
  box-shadow: var(--gb-shadow);
}

.group-buy-tabs button {
  min-height: 2.4rem;
  white-space: nowrap;
  border-radius: 999px;
  padding: 0 1rem;
  color: var(--gb-muted);
  font-size: 0.875rem;
  font-weight: 700;
  transition: background-color 0.16s ease, color 0.16s ease, transform 0.12s ease;
}

.group-buy-tabs button.active {
  background: var(--gb-text);
  color: var(--gb-bg);
}

.group-buy-alert,
.group-buy-inline-note {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgba(180, 83, 9, 0.25);
  border-radius: 0.5rem;
  background: #fff7ed;
  padding: 0.85rem 1rem;
  color: #92400e;
}

.group-buy-shell {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 1.1rem;
  align-items: start;
}

.group-buy-plan-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.1rem;
}

.group-buy-plan,
.group-buy-panel,
.group-buy-modal {
  border: 1px solid var(--gb-border);
  border-radius: 0.5rem;
  background: var(--gb-surface);
  box-shadow: var(--gb-shadow);
  color: var(--gb-text);
  backdrop-filter: blur(14px) saturate(1.04);
}

.group-buy-plan {
  display: flex;
  min-height: 26rem;
  flex-direction: column;
  padding: 1.3rem;
}

.group-buy-plan h2 {
  margin-top: 0.4rem;
  color: var(--gb-text);
  font-size: 1.2rem;
  font-weight: 800;
  line-height: 1.25;
  letter-spacing: 0;
}

.group-buy-plan-desc,
.group-buy-modal-desc,
.group-buy-panel-head p,
.group-buy-seat p,
.group-buy-caption {
  color: var(--gb-muted);
  line-height: 1.65;
}

.group-buy-price-row {
  margin-top: 1.35rem;
  display: flex;
  align-items: baseline;
  gap: 0.35rem;
}

.group-buy-price-row span {
  color: var(--gb-text);
  font-size: 2.05rem;
  font-weight: 800;
  line-height: 1;
  letter-spacing: 0;
}

.group-buy-price-row small {
  color: var(--gb-muted);
}

.group-buy-quota {
  margin-top: 1rem;
  border: 1px solid var(--gb-border);
  border-radius: 0.5rem;
  background: var(--gb-surface-soft);
  padding: 0.8rem;
  color: var(--gb-muted-strong);
  font-size: 0.9rem;
  font-weight: 700;
}

.group-buy-progress {
  margin-top: 1.25rem;
  color: var(--gb-muted);
}

.group-buy-progress-track {
  margin-top: 0.5rem;
  height: 0.48rem;
  overflow: hidden;
  border-radius: 999px;
  background: var(--gb-surface-soft);
}

.group-buy-progress-fill {
  height: 100%;
  border-radius: 999px;
  background: var(--gb-accent);
}

.group-buy-meta {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.group-buy-meta span,
.group-buy-status {
  border-radius: 999px;
  background: var(--gb-surface-soft);
  padding: 0.35rem 0.6rem;
  color: var(--gb-muted);
  font-size: 0.75rem;
  font-weight: 700;
}

.group-buy-status.ok { background: #dcfce7; color: #166534; }
.group-buy-status.warn { background: #fef3c7; color: #92400e; }
.group-buy-status.bad { background: #fee2e2; color: #991b1b; }
.group-buy-status.idle { background: var(--gb-surface-soft); color: var(--gb-muted); }

.group-buy-primary {
  display: inline-flex;
  min-height: 2.65rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 999px;
  background: var(--gb-accent);
  padding: 0 1.1rem;
  color: white;
  font-weight: 800;
  transition: background-color 0.16s ease, box-shadow 0.16s ease, transform 0.12s ease, opacity 0.12s ease;
}

.group-buy-primary:hover:not(:disabled) {
  background: var(--gb-accent-strong);
  box-shadow: 0 0 0 3px var(--gb-ring);
}

.group-buy-primary:disabled,
.group-buy-stepper button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.group-buy-plan .group-buy-primary {
  margin-top: auto;
}

.group-buy-side {
  position: sticky;
  top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.group-buy-panel {
  padding: 1.2rem;
}

.group-buy-wide-panel {
  min-height: 30rem;
}

.group-buy-panel-title,
.group-buy-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--gb-text);
  font-weight: 800;
}

.group-buy-panel-title {
  justify-content: flex-start;
}

.group-buy-activity {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
}

.group-buy-activity li {
  display: grid;
  grid-template-columns: 0.6rem 1fr;
  gap: 0.7rem;
}

.group-buy-activity li > span {
  margin-top: 0.45rem;
  height: 0.45rem;
  width: 0.45rem;
  border-radius: 999px;
  background: var(--gb-accent);
}

.group-buy-activity p {
  color: var(--gb-muted-strong);
  font-size: 0.9rem;
  line-height: 1.45;
}

.group-buy-activity time,
.group-buy-side-empty,
.group-buy-rules,
.group-buy-caption {
  color: var(--gb-muted);
  font-size: 0.8rem;
}

.group-buy-rules {
  margin-top: 0.8rem;
  display: grid;
  gap: 0.45rem;
  padding-left: 1rem;
  list-style: disc;
}

.group-buy-empty {
  display: grid;
  min-height: 20rem;
  place-items: center;
  gap: 0.75rem;
  border: 1px dashed var(--gb-border-strong);
  border-radius: 0.5rem;
  color: var(--gb-muted);
}

.group-buy-empty-soft {
  margin-top: 1rem;
  background: var(--gb-surface-soft);
}

.group-buy-spinner {
  display: inline-block;
  height: 2rem;
  width: 2rem;
  border: 3px solid var(--gb-border);
  border-top-color: var(--gb-accent);
  border-radius: 999px;
  animation: group-buy-spin 0.8s linear infinite;
}

.group-buy-spinner-small {
  height: 1rem;
  width: 1rem;
  border-width: 2px;
}

.group-buy-entitlement {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.group-buy-entitlement > div {
  border: 1px solid var(--gb-border);
  border-radius: 0.5rem;
  background: var(--gb-surface-soft);
  padding: 0.9rem;
}

.group-buy-entitlement strong {
  display: block;
  margin-top: 0.25rem;
  color: var(--gb-text);
  font-size: 1.05rem;
  font-weight: 800;
  overflow-wrap: anywhere;
}

.group-buy-inline-note {
  margin-top: 1rem;
  font-size: 0.88rem;
}

.group-buy-seat-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.75rem;
}

.group-buy-seat {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(260px, 320px);
  gap: 1rem;
  align-items: center;
  border-top: 1px solid var(--gb-border);
  padding: 1rem 0;
}

.group-buy-seat h3 {
  color: var(--gb-text);
  font-size: 1rem;
  font-weight: 800;
}

.group-buy-bind-box {
  display: flex;
  gap: 0.5rem;
}

.group-buy-bind-box input {
  min-width: 0;
  flex: 1;
  min-height: 2.65rem;
  border: 1px solid var(--gb-border-strong);
  border-radius: 999px;
  background: var(--gb-input);
  color: var(--gb-text);
  padding: 0 0.85rem;
}

.group-buy-bind-box input:focus {
  outline: 0;
  border-color: rgba(204, 120, 92, 0.58);
  box-shadow: 0 0 0 3px var(--gb-ring);
}

.group-buy-order-table {
  margin-top: 1rem;
  overflow-x: auto;
}

.group-buy-order-row {
  display: grid;
  min-width: 760px;
  grid-template-columns: 1.4fr 0.8fr 0.8fr 0.8fr 1fr;
  gap: 1rem;
  border-top: 1px solid var(--gb-border);
  padding: 0.85rem 0;
  align-items: center;
  color: var(--gb-muted-strong);
}

.group-buy-order-head {
  border-top: 0;
  color: var(--gb-muted);
  font-size: 0.8rem;
  font-weight: 800;
}

.group-buy-order-row small {
  display: block;
  margin-top: 0.2rem;
  color: var(--gb-muted);
  font-size: 0.72rem;
}

.group-buy-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(20, 20, 19, 0.28);
  padding: 1rem;
}

.group-buy-modal {
  position: relative;
  width: min(100%, 560px);
  max-height: 90vh;
  overflow-y: auto;
  padding: 1.5rem;
  background: var(--gb-surface);
  box-shadow: var(--gb-shadow-lg);
}

.group-buy-modal-close {
  position: absolute;
  right: 1rem;
  top: 1rem;
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--gb-surface-soft);
  color: var(--gb-muted-strong);
  transition: transform 0.12s ease, background-color 0.16s ease;
}

.group-buy-stepper {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: 2.75rem minmax(0, 1fr) 2.75rem;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid var(--gb-border);
  border-radius: 0.5rem;
  background: var(--gb-surface-soft);
  padding: 0.75rem;
  text-align: center;
}

.group-buy-stepper button {
  min-height: 2.4rem;
  border-radius: 999px;
  background: var(--gb-text);
  color: var(--gb-bg);
  font-size: 1.2rem;
  font-weight: 800;
  transition: transform 0.12s ease, opacity 0.12s ease;
}

.group-buy-stepper strong,
.group-buy-stepper span {
  display: block;
}

.group-buy-stepper strong {
  color: var(--gb-text);
  font-size: 1.25rem;
  font-weight: 800;
}

.group-buy-stepper span {
  color: var(--gb-muted);
  font-size: 0.8rem;
}

.group-buy-modal-summary {
  margin-top: 1.25rem;
  display: grid;
  gap: 0.75rem;
  border: 1px solid var(--gb-border);
  border-radius: 0.5rem;
  background: var(--gb-surface-soft);
  padding: 1rem;
}

.group-buy-modal-summary div {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.group-buy-modal-summary span {
  color: var(--gb-muted);
}

.group-buy-methods {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.5rem;
}

.group-buy-methods button {
  min-height: 2.4rem;
  border: 1px solid var(--gb-border);
  border-radius: 999px;
  color: var(--gb-muted);
  font-weight: 700;
  transition: background-color 0.16s ease, border-color 0.16s ease, color 0.16s ease, transform 0.12s ease;
}

.group-buy-methods button.active {
  border-color: var(--gb-accent);
  background: var(--gb-accent);
  color: white;
}

.group-buy-agreement {
  margin-top: 1rem;
  display: flex;
  gap: 0.65rem;
  color: var(--gb-muted-strong);
  font-size: 0.86rem;
  line-height: 1.6;
}

.group-buy-agreement input {
  margin-top: 0.25rem;
}

.group-buy-modal-submit {
  margin-top: 1rem;
  width: 100%;
}

@keyframes group-buy-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 1180px) {
  .group-buy-shell {
    grid-template-columns: 1fr;
  }
  .group-buy-side {
    position: static;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .group-buy-plan-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .group-buy-entitlement {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .group-buy-header,
  .group-buy-panel-head {
    align-items: flex-start;
    flex-direction: column;
  }
  .group-buy-refresh {
    align-self: flex-start;
  }
  .group-buy-plan-grid,
  .group-buy-side,
  .group-buy-seat,
  .group-buy-entitlement {
    grid-template-columns: 1fr;
  }
  .group-buy-methods {
    grid-template-columns: repeat(2, 1fr);
  }
  .group-buy-bind-box {
    flex-direction: column;
  }
  .group-buy-tabs {
    width: 100%;
  }
  .group-buy-tabs button {
    flex: 1 0 auto;
  }
}

:global(.dark) .group-buy-page {
  --gb-accent-strong: #c4b5fd;
  background: transparent;
  color: var(--gb-text);
}

:global(.dark) .group-buy-modal-backdrop {
  --gb-accent-strong: #c4b5fd;
  color: var(--gb-text);
}

:global(.dark) .group-buy-plan,
:global(.dark) .group-buy-panel,
:global(.dark) .group-buy-modal,
:global(.dark) .group-buy-tabs,
:global(.dark) .group-buy-refresh,
:global(.dark) .group-buy-secondary {
  border-color: var(--gb-border);
  background: var(--gb-surface);
  color: var(--gb-text);
}

:global(.dark) .group-buy-plan-desc,
:global(.dark) .group-buy-modal-desc,
:global(.dark) .group-buy-panel-head p,
:global(.dark) .group-buy-seat p,
:global(.dark) .group-buy-caption {
  color: var(--gb-muted);
}

:global(.dark) .group-buy-quota,
:global(.dark) .group-buy-entitlement > div,
:global(.dark) .group-buy-modal-summary,
:global(.dark) .group-buy-modal-close,
:global(.dark) .group-buy-stepper {
  background: var(--gb-surface-soft);
}
</style>
