<template>
  <AppLayout>
    <div class="pixel-cafe-page">
      <header class="pixel-cafe-header">
        <div v-if="pixelCafeHeaderVisible">
          <p class="pixel-cafe-kicker">PIXEL CAFE / BETA</p>
          <h1>{{ pixelCafeTitle }}</h1>
          <p v-if="pixelCafeDescription" class="pixel-cafe-subtitle">{{ pixelCafeDescription }}</p>
        </div>
        <router-link class="pixel-cafe-legacy" to="/group-buy?legacy=1">
          <Icon name="gift" size="sm" />
          旧版拼团
        </router-link>
      </header>

      <section class="pixel-cafe-my-rooms" aria-label="我的包间">
        <div class="pixel-cafe-my-rooms-heading">
          <div>
            <p class="pixel-cafe-label">我的包间</p>
            <h2>已购座位</h2>
          </div>
          <div class="pixel-cafe-my-rooms-tabs" role="tablist" aria-label="我的包间状态">
            <button
              type="button"
              :class="['pixel-cafe-my-rooms-tab', { active: myRoomsFilter === 'active,waiting' }]"
              :aria-selected="myRoomsFilter === 'active,waiting'"
              @click="loadMyRooms('active,waiting')"
            >进行中</button>
            <button
              type="button"
              :class="['pixel-cafe-my-rooms-tab', { active: myRoomsFilter === 'history' }]"
              :aria-selected="myRoomsFilter === 'history'"
              @click="loadMyRooms('history')"
            >历史</button>
          </div>
        </div>
        <div v-if="myRoomsLoading" class="pixel-cafe-my-rooms-state" data-testid="pixel-cafe-my-rooms-loading">正在加载我的包间...</div>
        <div v-else-if="myRoomsError" class="pixel-cafe-my-rooms-state pixel-cafe-my-rooms-error" data-testid="pixel-cafe-my-rooms-error">
          <span>{{ myRoomsError }}</span>
          <button type="button" class="pixel-cafe-my-rooms-retry" @click="loadMyRooms(myRoomsFilter)">重试</button>
        </div>
        <p v-else-if="myRooms.length === 0" class="pixel-cafe-my-rooms-state" data-testid="pixel-cafe-my-rooms-empty">
          {{ myRoomsFilter === 'history' ? '暂无历史包间记录。' : '暂无进行中的包间座位。' }}
        </p>
        <ul v-else class="pixel-cafe-my-rooms-list" data-testid="pixel-cafe-my-rooms-list">
          <li v-for="membership in myRooms" :key="membership.membership_id" class="pixel-cafe-my-room">
            <div class="pixel-cafe-my-room-title">
              <span class="pixel-cafe-my-room-code">{{ membership.room.code }}</span>
              <strong>{{ membership.room.name }}</strong>
            </div>
            <span :class="['pixel-cafe-my-room-state', `state-${membership.seat.status}`]">{{ myRoomStateLabel(membership.seat.status) }}</span>
            <span class="pixel-cafe-my-room-meta">座位 {{ membership.seat.seat_no ?? '-' }} · {{ membership.plan.validity_days }} 天</span>
            <span v-if="membership.managed_api_key" class="pixel-cafe-my-room-key">
              {{ membership.managed_api_key.name }} · {{ membership.managed_api_key.status }} · 额度 {{ membership.managed_api_key.quota }}
            </span>
          </li>
        </ul>
      </section>

      <nav class="pixel-cafe-zones" aria-label="网吧区域">
        <button
          v-for="zone in zones"
          :key="zone.key"
          type="button"
          :class="['pixel-cafe-zone', { active: selectedZone === zone.key }]"
          :aria-pressed="selectedZone === zone.key"
          @click="selectZone(zone.key)"
        >
          <span class="pixel-cafe-zone-dot" :style="{ backgroundColor: zoneColor(zone.key) }" aria-hidden="true"></span>
          {{ zone.name }}
        </button>
      </nav>

      <section class="pixel-cafe-workbench" aria-label="像素网吧大厅">
        <div class="pixel-cafe-scene">
          <div class="pixel-cafe-scene-topline">
            <span>大厅 / {{ activeZone.name }}</span>
            <span class="pixel-cafe-status"><i aria-hidden="true"></i>{{ loading ? '加载中' : '实时房间' }}</span>
          </div>
          <div v-if="loading" class="pixel-cafe-empty" data-testid="pixel-cafe-loading">正在加载房间...</div>
          <div v-else-if="errorMessage" class="pixel-cafe-empty pixel-cafe-error" data-testid="pixel-cafe-error">
            <p>{{ errorMessage }}</p>
            <button type="button" class="pixel-cafe-retry" @click="loadOverview">重试</button>
          </div>
          <div v-else-if="rooms.length === 0" class="pixel-cafe-empty" data-testid="pixel-cafe-empty">当前区域暂时没有可展示的房间。</div>
          <CafeScene
            v-else
            :rooms="rooms"
            :lobby-avatars="lobby.available ? lobby.avatars : []"
            :active-zone-label="activeZone.name"
            :selected-room-id="selectedRoom?.id"
            @select-room="openRoom"
          />
        </div>

        <aside class="pixel-cafe-inspector">
          <div class="pixel-cafe-inspector-heading">
            <span class="pixel-cafe-label">当前焦点</span>
            <Icon name="sparkles" size="sm" />
          </div>
          <template v-if="selectedRoom">
            <h2>{{ selectedRoom.name }}</h2>
            <p class="pixel-cafe-room-code">{{ selectedRoom.code }} / {{ activeZone.name }}</p>
            <dl class="pixel-cafe-stats">
              <div><dt>座位</dt><dd>{{ roomSeatLabel(selectedRoom) }}</dd></div>
              <div><dt>状态</dt><dd>{{ purchaseStateLabel(selectedRoom.purchase_state) }}</dd></div>
              <div><dt>周期</dt><dd>{{ selectedRoom.plan.validity_days }} 天</dd></div>
            </dl>
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
            <template v-else-if="selectedRoom.purchase_state === 'available'">
              <div class="pixel-cafe-seat-picker" aria-label="选择座位">
                <button
                  v-for="seat in availableSeats(selectedRoom)"
                  :key="seat.seat_no"
                  type="button"
                  :class="['pixel-cafe-seat-button', { active: selectedSeatNo === seat.seat_no }]"
                  :aria-pressed="selectedSeatNo === seat.seat_no"
                  @click="selectedSeatNo = seat.seat_no"
                >
                  {{ seat.seat_no }}
                </button>
              </div>
              <p v-if="availableSeats(selectedRoom).length === 0" class="pixel-cafe-inline-error">当前没有可选座位。</p>
              <label class="pixel-cafe-payment-label">
                支付方式
                <select v-model="selectedPaymentMethod" class="pixel-cafe-payment-select">
                  <option v-for="method in paymentMethods" :key="method" :value="method">{{ paymentMethodLabel(method) }}</option>
                </select>
              </label>
              <label class="pixel-cafe-agreement">
                <input v-model="agreementAccepted" type="checkbox" />
                <span>我已确认当前包间在满员并激活后生效。</span>
              </label>
              <p v-if="orderError" class="pixel-cafe-inline-error" data-testid="pixel-cafe-order-error">{{ orderError }}</p>
              <button
                type="button"
                class="pixel-cafe-primary"
                :disabled="submitting || !selectedSeatNo || !agreementAccepted"
                @click="submitOrder"
              >
                {{ submitting ? '正在创建订单' : '确认座位并付款' }}
              </button>
            </template>
            <p v-else class="pixel-cafe-muted">当前房间暂不接受新座位。</p>
          </template>
          <template v-else>
            <h2>选择一间包间</h2>
            <p class="pixel-cafe-muted">房间进度来自服务端；选择包间可查看当前座位与开团状态。</p>
          </template>
        </aside>
      </section>

      <section class="pixel-cafe-notice" aria-label="阶段状态">
        <div class="pixel-cafe-notice-icon"><Icon name="infoCircle" size="sm" /></div>
        <div>
          <strong>房间发现已接入</strong>
          <p>当前仅展示公开房间信息；选座、支付和受管 Key 会在后续阶段接入，旧版拼团入口始终保留。</p>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import CafeScene from './components/CafeScene.vue'
import { cafeAPI } from '@/api/cafe'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  normalizeVisibleMethod,
  writePaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { CafeLobbyActivity, CafeMyRoom, CafeMyRoomFilter, CafePublicRoom, CafePublicSeatVisual, CafePublicZone, CreateCafeRoomOrderResult } from '@/types/pixelCafe'

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const router = useRouter()

const zones = ref<CafePublicZone[]>([])
const rooms = ref<CafePublicRoom[]>([])
const lobby = ref<CafeLobbyActivity>(emptyLobbyActivity())
const myRooms = ref<CafeMyRoom[]>([])
const myRoomsFilter = ref<CafeMyRoomFilter>('active,waiting')
const myRoomsLoading = ref(false)
const myRoomsError = ref('')
const selectedZone = ref('featured')
const selectedRoom = ref<CafePublicRoom | null>(null)
const loading = ref(false)
const errorMessage = ref('')
const selectedSeatNo = ref<number | null>(null)
const selectedPaymentMethod = ref('alipay')
const agreementAccepted = ref(false)
const submitting = ref(false)
const orderError = ref('')
const paymentPhase = ref<'selecting' | 'paying'>('selecting')
const paymentMethods = ['alipay', 'wxpay', 'stripe', 'airwallex']
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
let roomRequest = 0
let myRoomsRequest = 0
let previousDocumentOverflowX = ''
let previousBodyOverflowX = ''

const appStore = useAppStore()
const pixelCafeHeaderVisible = computed(() => appStore.cachedPublicSettings?.pixel_cafe_header_visible !== false)
const pixelCafeTitle = computed(() => {
  const value = appStore.cachedPublicSettings?.pixel_cafe_title
  return typeof value === 'string' && value.trim() ? value.trim() : '像素网吧'
})
const pixelCafeDescription = computed(() => {
  const value = appStore.cachedPublicSettings?.pixel_cafe_description
  return typeof value === 'string' ? value.trim() : '把每个模型分组变成一间可订阅的数字包间。'
})

const activeZone = computed<CafePublicZone>(() => zones.value.find(zone => zone.key === selectedZone.value) ?? {
  key: selectedZone.value,
  name: '精选大厅',
  room_count: 0,
  open_seat_count: 0,
})

async function loadOverview(): Promise<void> {
  loading.value = true
  errorMessage.value = ''
  selectedRoom.value = null
  try {
    const response = await cafeAPI.overview({ room_limit: 8 })
    zones.value = Array.isArray(response.data.zones) ? response.data.zones : []
    lobby.value = normalizeLobbyActivity(response.data.lobby)
    const preferredZone = zones.value.some(zone => zone.key === 'featured')
      ? 'featured'
      : zones.value[0]?.key || 'featured'
    selectedZone.value = preferredZone
    await loadRooms(preferredZone, response.data.rooms)
  } catch (error) {
    rooms.value = []
    errorMessage.value = extractApiErrorMessage(error, '加载房间失败')
  } finally {
    loading.value = false
  }
}

function emptyLobbyActivity(): CafeLobbyActivity {
  return {
    available: false,
    date: '',
    timezone: '',
    label: '今日使用用户',
    unique_users: 0,
    successful_requests: 0,
    display_max: 50,
    avatars: [],
  }
}

function normalizeLobbyActivity(value: CafeLobbyActivity | undefined): CafeLobbyActivity {
  if (!value || !value.available) return emptyLobbyActivity()
  return {
    ...emptyLobbyActivity(),
    ...value,
    avatars: Array.isArray(value.avatars) ? value.avatars.slice(0, 50) : [],
  }
}

async function loadMyRooms(filter: CafeMyRoomFilter = myRoomsFilter.value): Promise<void> {
  const requestID = ++myRoomsRequest
  myRoomsFilter.value = filter
  myRoomsLoading.value = true
  myRoomsError.value = ''
  try {
    const response = await cafeAPI.listMyRooms({ page: 1, page_size: 20, status: filter })
    if (requestID === myRoomsRequest) myRooms.value = Array.isArray(response.data.items) ? response.data.items : []
  } catch (error) {
    if (requestID === myRoomsRequest) {
      myRooms.value = []
      myRoomsError.value = extractApiErrorMessage(error, '加载我的包间失败')
    }
  } finally {
    if (requestID === myRoomsRequest) myRoomsLoading.value = false
  }
}

async function selectZone(zone: string): Promise<void> {
  if (zone === selectedZone.value && rooms.value.length > 0) return
  selectedZone.value = zone
  selectedRoom.value = null
  await loadRooms(zone)
}

async function loadRooms(zone: string, overviewRooms?: CafePublicRoom[]): Promise<void> {
  const requestID = ++roomRequest
  loading.value = true
  errorMessage.value = ''
  try {
    if (zone === 'featured' && overviewRooms) {
      if (requestID === roomRequest) rooms.value = overviewRooms
      return
    }
    const params = zone === 'featured'
      ? { page: 1, page_size: 24, featured: true }
      : { page: 1, page_size: 24, zone }
    const response = await cafeAPI.listRooms(params)
    if (requestID === roomRequest) rooms.value = Array.isArray(response.data.items) ? response.data.items : []
  } catch (error) {
    if (requestID === roomRequest) {
      rooms.value = []
      errorMessage.value = extractApiErrorMessage(error, '加载房间失败')
    }
  } finally {
    if (requestID === roomRequest) loading.value = false
  }
}

function roomSeatLabel(room: CafePublicRoom): string {
  if (!room.round) return `${room.plan.total_seats} 席 · 暂未开团`
  return `${room.round.remaining_seats}/${room.plan.total_seats} 空位`
}

function myRoomStateLabel(status: string): string {
  const labels: Record<string, string> = {
    locked: '待付款',
    paid: '等待开通',
    active: '使用中',
    released: '已释放',
    cancelled: '已取消',
    refund_pending: '退款处理中',
    refund_processing: '退款处理中',
    refunded: '已退款',
  }
  return labels[status] || status
}

function openRoom(room: CafePublicRoom): void {
  selectedRoom.value = room
  selectedSeatNo.value = null
  agreementAccepted.value = false
  orderError.value = ''
  paymentPhase.value = 'selecting'
}

function availableSeats(room: CafePublicRoom): CafePublicSeatVisual[] {
  return room.seat_visuals.filter(seat => seat.state === 'empty')
}

async function submitOrder(): Promise<void> {
  const room = selectedRoom.value
  if (!room || !selectedSeatNo.value || !agreementAccepted.value || submitting.value) return
  submitting.value = true
  orderError.value = ''
  const visibleMethod = normalizeVisibleMethod(selectedPaymentMethod.value) || selectedPaymentMethod.value
  try {
    const result = (await cafeAPI.createOrder(room.id, {
      seat_no: selectedSeatNo.value,
      payment_type: visibleMethod,
      return_url: `${window.location.origin}/payment/result`,
      payment_source: visibleMethod === 'wxpay' && isWechatBrowser() ? 'wechat_in_app_resume' : 'hosted_redirect',
      is_mobile: isMobileDevice(),
      agreement_accepted: true,
    }, createIdempotencyKey())).data as CreateCafeRoomOrderResult
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
      orderError.value = '当前支付方式暂不可用，请换一种方式。'
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
      if (isMobileDevice()) window.location.href = decision.paymentState.payUrl
      else openPopup(decision.paymentState.payUrl)
    }
  } catch (error) {
    orderError.value = extractApiErrorMessage(error, '创建座位订单失败')
  } finally {
    submitting.value = false
  }
}

function onPaymentDone(): void {
  paymentPhase.value = 'selecting'
  selectedSeatNo.value = null
  agreementAccepted.value = false
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
  void loadOverview()
  void loadMyRooms()
}

function onPaymentSuccess(): void {
  orderError.value = '付款成功，正在更新房间状态。'
}

function onPaymentSettled(outcome: PaymentOutcome): void {
  if (outcome !== 'success') {
    void loadOverview()
    void loadMyRooms()
  }
}

function createIdempotencyKey(): string {
  if (typeof window.crypto?.randomUUID === 'function') return window.crypto.randomUUID()
  return `cafe-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function openPopup(url: string): void {
  const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
  if (!win || win.closed) window.location.href = url
}

function isWechatBrowser(): boolean {
  return /MicroMessenger/i.test(window.navigator.userAgent)
}

function purchaseStateLabel(state: string): string {
  const labels: Record<string, string> = {
    available: '可购买',
    full: '已满',
    activating: '开通中',
    active: '已开通',
    unavailable: '暂不可用',
  }
  return labels[state] || '暂不可用'
}

function paymentMethodLabel(method: string): string {
  const normalized = method.toLowerCase()
  if (normalized.includes('wxpay')) return '微信支付'
  if (normalized.includes('alipay')) return '支付宝'
  if (normalized === 'stripe') return 'Stripe'
  if (normalized === 'airwallex') return 'Airwallex'
  return method
}

function zoneColor(zone: string): string {
  return ({ featured: '#cc785c', claude: '#8b6f5b', openai: '#6f8f78', gemini: '#6d7fa7' } as Record<string, string>)[zone] || '#a9785d'
}

function containHorizontalOverflow(): void {
  previousDocumentOverflowX = document.documentElement.style.overflowX
  previousBodyOverflowX = document.body.style.overflowX
  document.documentElement.style.overflowX = 'hidden'
  document.body.style.overflowX = 'hidden'
}

function restoreHorizontalOverflow(): void {
  document.documentElement.style.overflowX = previousDocumentOverflowX
  document.body.style.overflowX = previousBodyOverflowX
}

onMounted(() => {
  containHorizontalOverflow()
  void loadOverview()
  void loadMyRooms()
})

onUnmounted(() => {
  restoreHorizontalOverflow()
})
</script>

<style scoped>
.pixel-cafe-page { min-height: calc(100vh - 4rem); padding: 1.25rem; color: #2b2927; background: #f4f0e8; }
.pixel-cafe-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; max-width: 1400px; margin: 0 auto 1rem; }
.pixel-cafe-kicker { margin: 0 0 .35rem; color: #8f624f; font: 700 .72rem/1 monospace; letter-spacing: .12em; }
.pixel-cafe-header h1 { margin: 0; font-size: clamp(1.8rem, 4vw, 2.7rem); line-height: 1.05; }
.pixel-cafe-subtitle { margin: .5rem 0 0; color: #716b64; font-size: .95rem; }
.pixel-cafe-legacy { display: inline-flex; align-items: center; gap: .45rem; padding: .65rem .85rem; border: 1px solid #d3c8b8; color: #66584d; background: #fffdf8; text-decoration: none; font-size: .85rem; }
.pixel-cafe-legacy:hover { border-color: #b67b61; color: #8f624f; }
.pixel-cafe-my-rooms { max-width: 1400px; margin: 0 auto 1rem; padding: .9rem 1rem; border-top: 2px solid #cc785c; border-bottom: 1px solid #d6cbbb; background: #fffdf8; }
.pixel-cafe-my-rooms-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }.pixel-cafe-my-rooms-heading h2 { margin: .3rem 0 0; font-size: 1rem; }.pixel-cafe-my-rooms-tabs { display: flex; border: 1px solid #d7cdbf; }.pixel-cafe-my-rooms-tab { min-width: 4.25rem; padding: .45rem .65rem; border: 0; border-right: 1px solid #d7cdbf; color: #74695d; background: #fffdf8; cursor: pointer; font-size: .76rem; }.pixel-cafe-my-rooms-tab:last-child { border-right: 0; }.pixel-cafe-my-rooms-tab.active { color: #fffdf8; background: #9a644f; }.pixel-cafe-my-rooms-state { margin: .8rem 0 0; color: #776e65; font-size: .8rem; }.pixel-cafe-my-rooms-error { display: flex; align-items: center; gap: .7rem; color: #a94d48; }.pixel-cafe-my-rooms-retry { padding: .35rem .55rem; border: 1px solid #b97867; color: #824d40; background: #fffdf8; cursor: pointer; font-size: .75rem; }.pixel-cafe-my-rooms-list { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: .75rem; margin: .85rem 0 0; padding: 0; list-style: none; }.pixel-cafe-my-room { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: .3rem .7rem; padding: .65rem .75rem; border-left: 3px solid #b97960; background: #f4f0e8; }.pixel-cafe-my-room-title { min-width: 0; }.pixel-cafe-my-room-title strong { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: .82rem; }.pixel-cafe-my-room-code, .pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { color: #776e65; font-size: .7rem; }.pixel-cafe-my-room-state { align-self: center; padding: .18rem .35rem; border: 1px solid #c8b5a2; color: #745646; background: #fffdf8; font-size: .68rem; }.pixel-cafe-my-room-state.state-active { border-color: #789275; color: #3e6849; }.pixel-cafe-my-room-state.state-refunded, .pixel-cafe-my-room-state.state-cancelled, .pixel-cafe-my-room-state.state-released { border-color: #b99c98; color: #8b5d58; }.pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { grid-column: 1 / -1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pixel-cafe-zones { display: flex; gap: .5rem; max-width: 1400px; margin: 0 auto 1rem; overflow-x: auto; }
.pixel-cafe-zone { display: inline-flex; align-items: center; gap: .45rem; flex: 0 0 auto; padding: .55rem .75rem; border: 1px solid #d7cdbf; color: #776e65; background: #fffdf8; cursor: pointer; font-size: .82rem; }
.pixel-cafe-zone.active { border-color: #ad7258; color: #6b4a3d; box-shadow: inset 0 -2px #ad7258; }
.pixel-cafe-zone-dot { width: .55rem; height: .55rem; box-shadow: 2px 2px 0 rgba(43, 41, 39, .18); }
.pixel-cafe-workbench { display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 1rem; max-width: 1400px; margin: 0 auto; }
.pixel-cafe-scene, .pixel-cafe-inspector, .pixel-cafe-notice { border: 1px solid #d6cbbb; background: #fffdf8; box-shadow: 4px 4px 0 rgba(89, 67, 50, .12); }
.pixel-cafe-scene { position: relative; min-height: 500px; overflow: hidden; background-color: #e7ded0; background-image: linear-gradient(#d8cbbc 1px, transparent 1px), linear-gradient(90deg, #d8cbbc 1px, transparent 1px); background-size: 24px 24px; }
.pixel-cafe-scene-topline { display: flex; justify-content: space-between; padding: .75rem 1rem; border-bottom: 1px solid #cfc1b2; color: #74695d; background: rgba(255, 253, 248, .88); font-size: .78rem; }
.pixel-cafe-status { display: inline-flex; align-items: center; gap: .35rem; }.pixel-cafe-status i { width: .5rem; height: .5rem; border-radius: 999px; background: #c28d4c; }
.pixel-cafe-empty { display: grid; min-height: 240px; padding: 2rem; place-items: center; color: #776e65; text-align: center; }
.pixel-cafe-error { gap: .75rem; color: #a94d48; }.pixel-cafe-error p { margin: 0; }.pixel-cafe-retry { padding: .5rem .75rem; border: 1px solid #b97867; color: #824d40; background: #fffdf8; cursor: pointer; }
.pixel-cafe-inspector { padding: 1.2rem; }.pixel-cafe-inspector-heading { display: flex; justify-content: space-between; color: #9a6a53; }.pixel-cafe-label { font: 700 .7rem monospace; text-transform: uppercase; }.pixel-cafe-inspector h2 { margin: 1.25rem 0 .3rem; font-size: 1.25rem; }.pixel-cafe-room-code { margin: 0; color: #8c8278; font: .75rem monospace; }.pixel-cafe-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: .5rem; margin: 1.5rem 0; }.pixel-cafe-stats div { padding: .55rem .4rem; border: 1px solid #e0d6c8; text-align: center; }.pixel-cafe-stats dt { color: #8c8278; font-size: .68rem; }.pixel-cafe-stats dd { margin: .25rem 0 0; font-size: .78rem; font-weight: 700; }.pixel-cafe-primary { width: 100%; padding: .7rem; border: 0; color: #fffdf8; background: #a9785d; font-weight: 700; }.pixel-cafe-primary:disabled { cursor: not-allowed; opacity: .72; }.pixel-cafe-muted { color: #82786e; font-size: .86rem; line-height: 1.6; }
.pixel-cafe-seat-picker { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: .45rem; margin: 1rem 0; }.pixel-cafe-seat-button { min-height: 2.2rem; border: 1px solid #c9bdac; color: #5d5148; background: #fffdf8; cursor: pointer; font: 700 .8rem monospace; }.pixel-cafe-seat-button:hover, .pixel-cafe-seat-button:focus-visible, .pixel-cafe-seat-button.active { border-color: #9a644f; color: #fffdf8; background: #a9785d; outline: 0; }.pixel-cafe-payment-label { display: grid; gap: .35rem; margin: 1rem 0; color: #74695d; font-size: .78rem; }.pixel-cafe-payment-select { width: 100%; min-height: 2.3rem; border: 1px solid #cfc1b2; border-radius: 0; color: #473d36; background: #fffdf8; }.pixel-cafe-agreement { display: flex; align-items: flex-start; gap: .45rem; margin: 1rem 0; color: #74695d; font-size: .76rem; line-height: 1.45; }.pixel-cafe-agreement input { margin-top: .1rem; accent-color: #9a644f; }.pixel-cafe-inline-error { margin: .75rem 0; color: #a94d48; font-size: .78rem; line-height: 1.4; }
.pixel-cafe-notice { display: flex; gap: .75rem; align-items: flex-start; max-width: 1400px; margin: 1rem auto 0; padding: .85rem 1rem; }.pixel-cafe-notice-icon { display: grid; width: 1.8rem; height: 1.8rem; place-items: center; color: #8f624f; background: #f1e0d3; }.pixel-cafe-notice strong { font-size: .82rem; }.pixel-cafe-notice p { margin: .25rem 0 0; color: #776e65; font-size: .78rem; }
@media (max-width: 900px) { .pixel-cafe-workbench { grid-template-columns: 1fr; }.pixel-cafe-inspector { min-height: 0; }.pixel-cafe-scene { min-height: 430px; } }
@media (max-width: 620px) { .pixel-cafe-page { padding: .85rem; }.pixel-cafe-header { align-items: stretch; flex-direction: column; }.pixel-cafe-legacy { align-self: flex-start; }.pixel-cafe-my-rooms-heading { align-items: flex-start; flex-direction: column; }.pixel-cafe-my-rooms-list { grid-template-columns: 1fr; } }
</style>
