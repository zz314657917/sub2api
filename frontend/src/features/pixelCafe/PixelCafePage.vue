<template>
  <AppLayout>
    <div class="pixel-cafe-page">
      <header class="pixel-cafe-header">
        <div v-if="pixelCafeHeaderVisible">
          <p class="pixel-cafe-kicker">PIXEL CAFE / BETA</p>
          <h1>{{ pixelCafeTitle }}</h1>
          <p v-if="pixelCafeDescription" class="pixel-cafe-subtitle">{{ pixelCafeDescription }}</p>
        </div>
      </header>

      <section class="pixel-cafe-my-rooms" aria-label="我的包间">
        <div class="pixel-cafe-my-rooms-heading">
          <div>
            <p class="pixel-cafe-label">我的包间</p>
            <h2>已加入房间</h2>
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
          {{ myRoomsFilter === 'history' ? '暂无历史包间记录。' : '暂无进行中的包间。' }}
        </p>
        <ul v-else class="pixel-cafe-my-rooms-list" data-testid="pixel-cafe-my-rooms-list">
          <li v-for="membership in myRooms" :key="membership.membership_id" class="pixel-cafe-my-room">
            <div class="pixel-cafe-my-room-title">
              <span class="pixel-cafe-my-room-code">{{ membership.room.code }}</span>
              <strong>{{ membership.room.name }}</strong>
            </div>
            <span :class="['pixel-cafe-my-room-state', `state-${membership.seat.status}`]">{{ myRoomStateLabel(membership.seat.status) }}</span>
            <span class="pixel-cafe-my-room-meta">房间名额 {{ membership.seat.seat_no ?? '-' }} · {{ membership.plan.validity_days }} 天</span>
            <span v-if="membership.account" class="pixel-cafe-my-room-account">
              绑定账号：{{ membership.account.name }} · {{ membership.account.platform }}<span v-if="membership.account.email_masked"> · {{ membership.account.email_masked }}</span>
            </span>
            <span v-else class="pixel-cafe-my-room-account pixel-cafe-my-room-empty">未绑定账号</span>
            <span v-if="membership.managed_api_key" class="pixel-cafe-my-room-key">
              {{ membership.managed_api_key.name }} · {{ membership.managed_api_key.status }} · 总额度 {{ formatQuota(membership.managed_api_key.quota, membership.managed_api_key.quota_used) }}
              · 5H {{ formatWindowQuota(membership.managed_api_key.usage_5h, membership.managed_api_key.rate_limit_5h) }}
              · 7D {{ formatWindowQuota(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d) }}
            </span>
            <span v-else class="pixel-cafe-my-room-key pixel-cafe-my-room-empty">暂无受管 Key</span>
          </li>
        </ul>
      </section>

      <nav class="pixel-cafe-zones" aria-label="房间筛选">
        <button
          v-for="zone in zones"
          :key="zone.key"
          type="button"
          :class="['pixel-cafe-zone', { active: selectedZone === zone.key }]"
          :aria-pressed="selectedZone === zone.key"
          @click="selectZone(zone.key)"
        >
          <span class="pixel-cafe-zone-dot" :style="{ backgroundColor: zoneColor(zone.key) }" aria-hidden="true"></span>
          {{ zone.name }} · {{ zone.room_count }}
        </button>
      </nav>

      <section class="pixel-cafe-room-list" aria-label="房间列表" data-testid="pixel-cafe-room-list">
        <div class="pixel-cafe-room-list-heading">
          <div>
            <p class="pixel-cafe-label">ROOMS / 包间</p>
            <h2>{{ activeZone.name }}</h2>
          </div>
          <span class="pixel-cafe-room-list-count">{{ rooms.length }} 间包间</span>
        </div>
        <div v-if="!loading && !errorMessage && rooms.length > 0" class="pixel-cafe-room-cards">
          <button
            v-for="room in rooms"
            :key="room.id"
            type="button"
            :class="['pixel-cafe-room-card', { active: selectedRoom?.id === room.id }]"
            :aria-pressed="selectedRoom?.id === room.id"
            @click="openRoom(room)"
          >
            <span class="pixel-cafe-room-card-topline">
              <span class="pixel-cafe-room-card-code">{{ room.code }}</span>
              <span :class="['pixel-cafe-room-card-state', `state-${roomTone(room)}`]">{{ roomProgressLabel(room) }}</span>
            </span>
            <strong>{{ room.name }}</strong>
            <span class="pixel-cafe-room-card-plan">{{ room.plan.title }} · {{ roomCapacityModeLabel(room) }}</span>
            <span class="pixel-cafe-room-card-stats">
              <span>{{ room.round?.remaining_seats ?? 0 }}/{{ room.plan.total_seats }} 空位</span>
              <span>{{ room.plan.validity_days }} 天</span>
              <span>{{ room.plan.price_label || `${room.plan.price_per_seat} CNY` }}</span>
            </span>
            <span class="pixel-cafe-room-card-action">{{ room.plan.total_seats === 1 ? '预订独享房间' : '查看并加入' }} <span aria-hidden="true">→</span></span>
          </button>
        </div>
        <p v-else-if="!loading && !errorMessage" class="pixel-cafe-room-list-empty">当前区域暂无可加入的包间。</p>
      </section>

      <section class="pixel-cafe-workbench" aria-label="像素网吧大厅">
        <div class="pixel-cafe-scene">
          <div class="pixel-cafe-scene-topline">
            <span>大厅 / {{ activeZone.name }}</span>
            <span class="pixel-cafe-status"><i aria-hidden="true"></i>{{ loading ? '加载中' : isLocalDemoMode ? '本地演示' : '实时房间' }}</span>
          </div>
          <CafeScene
            :rooms="rooms"
            :lobby-avatars="sceneAvatars"
            @select-room="openRoom"
          />
          <div class="pixel-cafe-front-desk" aria-label="网吧前台">
            <span class="pixel-cafe-front-desk-lamp" aria-hidden="true"></span>
            <div>
              <p>前台</p>
              <strong>选择独享房间，或加入一个共享房间。</strong>
            </div>
          </div>
          <div v-if="loading" class="pixel-cafe-empty pixel-cafe-scene-overlay" data-testid="pixel-cafe-loading">正在加载房间...</div>
          <div v-else-if="errorMessage" class="pixel-cafe-empty pixel-cafe-error pixel-cafe-scene-overlay" data-testid="pixel-cafe-error">
            <p>{{ errorMessage }}</p>
            <button type="button" class="pixel-cafe-retry" @click="loadOverview">重试</button>
          </div>
          <div v-else-if="rooms.length === 0" class="pixel-cafe-empty pixel-cafe-scene-overlay" data-testid="pixel-cafe-empty">当前区域暂时没有可展示的房间。</div>
          <p v-if="isLocalDemoMode" class="pixel-cafe-demo-badge" data-testid="pixel-cafe-demo-badge">本地演示数据 · 不创建真实订单</p>
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
              <div><dt>模式</dt><dd>{{ roomCapacityModeLabel(selectedRoom) }}</dd></div>
              <div><dt>名额</dt><dd>{{ roomSeatLabel(selectedRoom) }}</dd></div>
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
              <div v-if="selectedRoom.plan.total_seats > 1" class="pixel-cafe-seat-picker" aria-label="选择房间名额">
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
              <p v-else class="pixel-cafe-single-room-note">这是独享房间，支付后占用唯一名额。</p>
              <p v-if="availableSeats(selectedRoom).length === 0" class="pixel-cafe-inline-error">当前没有可用名额。</p>
              <label class="pixel-cafe-payment-label">
                支付方式
                <select v-model="selectedPaymentMethod" class="pixel-cafe-payment-select">
                  <option v-for="method in paymentMethods" :key="method" :value="method">{{ paymentMethodLabel(method) }}</option>
                </select>
              </label>
              <label class="pixel-cafe-agreement">
                <input v-model="agreementAccepted" type="checkbox" />
                <span>我已确认加入该房间，具体开通时间以房间状态为准。</span>
              </label>
              <p v-if="orderError" class="pixel-cafe-inline-error" data-testid="pixel-cafe-order-error">{{ orderError }}</p>
              <button
                type="button"
                class="pixel-cafe-primary"
                :disabled="isLocalDemoMode || submitting || !selectedSeatNo || !agreementAccepted"
                @click="submitOrder"
              >
                {{ isLocalDemoMode ? '本地演示不创建订单' : submitting ? '正在创建订单' : selectedRoom.plan.total_seats === 1 ? '预订独享房间' : '加入房间并付款' }}
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
          <p>房间人数上限由计划决定：1 人为独享，2-10 人为共享；支付和受管 Key 按房间状态及账户权益开放。</p>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import CafeScene from './components/CafeScene.vue'
import { createPixelCafeDemoOverview, isLocalPixelCafeDemo } from './demoData'
import { cafeAPI } from '@/api/cafe'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { resolvePixelCafeTitle } from '@/utils/groupBuyProduct'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  normalizeVisibleMethod,
  writePaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { CafeLobbyActivity, CafeLobbyAvatar, CafeMyRoom, CafeMyRoomFilter, CafePublicRoom, CafePublicSeatVisual, CafePublicZone, CreateCafeRoomOrderResult } from '@/types/pixelCafe'

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const router = useRouter()
const route = useRoute()

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
const pixelCafeTitle = computed(() => resolvePixelCafeTitle(appStore.cachedPublicSettings))
const pixelCafeDescription = computed(() => {
  const value = appStore.cachedPublicSettings?.pixel_cafe_description
  return typeof value === 'string' ? value.trim() : '把每个模型分组变成一间可订阅的数字包间。'
})
const isLocalDemoMode = computed(() => isLocalPixelCafeDemo(route.query))
const sceneAvatars = computed<CafeLobbyAvatar[]>(() => {
  if (lobby.value.available && lobby.value.avatars.length > 0) return lobby.value.avatars

  return rooms.value.flatMap((room) => room.seat_visuals
    .filter((seat) => seat.state !== 'empty')
    .map((seat) => ({
      avatar_seed: seat.avatar_seed || `room-${room.id}-seat-${seat.seat_no}`,
      seat_index: (room.id * 17 + seat.seat_no) % 12,
      activity: 'recent' as const,
    })),
  ).slice(0, 12)
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
    const overview = isLocalDemoMode.value
      ? createPixelCafeDemoOverview()
      : (await cafeAPI.overview({ room_limit: 8 })).data
    zones.value = Array.isArray(overview.zones) ? overview.zones : []
    lobby.value = normalizeLobbyActivity(overview.lobby)
    const preferredZone = zones.value.some(zone => zone.key === 'featured')
      ? 'featured'
      : zones.value[0]?.key || 'featured'
    selectedZone.value = preferredZone
    await loadRooms(preferredZone, overview.rooms)
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
    if (isLocalDemoMode.value) {
      const demoRooms = createPixelCafeDemoOverview().rooms
      if (requestID === roomRequest) {
        rooms.value = zone === 'featured' ? demoRooms : demoRooms.filter(room => room.zone_key === zone)
      }
      return
    }
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
  if (!room.round) return `${room.plan.total_seats} 个名额 · 暂未开放`
  return `${room.round.remaining_seats}/${room.plan.total_seats} 空位`
}

function roomCapacityModeLabel(room: CafePublicRoom): string {
  return room.plan.total_seats === 1 ? '独享房间' : `${room.plan.total_seats} 人共享`
}

function roomProgressLabel(room: CafePublicRoom): string {
  if (room.round?.status === 'activating') return '开通中'
  if (room.round?.status === 'active') return '已开通'
  if (room.round?.status === 'open') return room.plan.total_seats === 1 ? '可预订' : '可加入'
  return room.purchase_state === 'full' ? '已满' : '暂不可用'
}

function roomTone(room: CafePublicRoom): string {
  if (room.zone_key === 'openai' || room.theme_key.includes('green')) return 'green'
  if (room.zone_key === 'gemini' || room.theme_key.includes('blue')) return 'blue'
  if (room.purchase_state === 'unavailable') return 'night'
  return 'wood'
}

function formatQuota(quota: number, used: number): string {
  if (quota <= 0) return `已用 ${used.toFixed(2)} / 不限`
  return `${used.toFixed(2)} / ${quota.toFixed(2)}`
}

function formatWindowQuota(used: number, limit: number): string {
  if (limit <= 0) return `${used.toFixed(2)} / 不限`
  return `${used.toFixed(2)} / ${limit.toFixed(2)}`
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
  const firstAvailableSeat = availableSeats(room)[0]
  selectedSeatNo.value = room.plan.total_seats === 1 ? firstAvailableSeat?.seat_no ?? null : null
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
.pixel-cafe-my-rooms { max-width: 1400px; margin: 0 auto 1rem; padding: .9rem 1rem; border-top: 2px solid #cc785c; border-bottom: 1px solid #d6cbbb; background: #fffdf8; }
.pixel-cafe-my-rooms-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }.pixel-cafe-my-rooms-heading h2 { margin: .3rem 0 0; font-size: 1rem; }.pixel-cafe-my-rooms-tabs { display: flex; border: 1px solid #d7cdbf; }.pixel-cafe-my-rooms-tab { min-width: 4.25rem; padding: .45rem .65rem; border: 0; border-right: 1px solid #d7cdbf; color: #74695d; background: #fffdf8; cursor: pointer; font-size: .76rem; }.pixel-cafe-my-rooms-tab:last-child { border-right: 0; }.pixel-cafe-my-rooms-tab.active { color: #fffdf8; background: #9a644f; }.pixel-cafe-my-rooms-state { margin: .8rem 0 0; color: #776e65; font-size: .8rem; }.pixel-cafe-my-rooms-error { display: flex; align-items: center; gap: .7rem; color: #a94d48; }.pixel-cafe-my-rooms-retry { padding: .35rem .55rem; border: 1px solid #b97867; color: #824d40; background: #fffdf8; cursor: pointer; font-size: .75rem; }.pixel-cafe-my-rooms-list { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: .75rem; margin: .85rem 0 0; padding: 0; list-style: none; }.pixel-cafe-my-room { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: .3rem .7rem; padding: .65rem .75rem; border-left: 3px solid #b97960; background: #f4f0e8; }.pixel-cafe-my-room-title { min-width: 0; }.pixel-cafe-my-room-title strong { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: .82rem; }.pixel-cafe-my-room-code, .pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { color: #776e65; font-size: .7rem; }.pixel-cafe-my-room-state { align-self: center; padding: .18rem .35rem; border: 1px solid #c8b5a2; color: #745646; background: #fffdf8; font-size: .68rem; }.pixel-cafe-my-room-state.state-active { border-color: #789275; color: #3e6849; }.pixel-cafe-my-room-state.state-refunded, .pixel-cafe-my-room-state.state-cancelled, .pixel-cafe-my-room-state.state-released { border-color: #b99c98; color: #8b5d58; }.pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { grid-column: 1 / -1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pixel-cafe-zones { display: flex; gap: .5rem; max-width: 1400px; margin: 0 auto 1rem; overflow-x: auto; }
.pixel-cafe-zone { display: inline-flex; align-items: center; gap: .45rem; flex: 0 0 auto; padding: .55rem .75rem; border: 1px solid #d7cdbf; color: #776e65; background: #fffdf8; cursor: pointer; font-size: .82rem; }
.pixel-cafe-zone.active { border-color: #ad7258; color: #6b4a3d; box-shadow: inset 0 -2px #ad7258; }
.pixel-cafe-zone-dot { width: .55rem; height: .55rem; box-shadow: 2px 2px 0 rgba(43, 41, 39, .18); }
.pixel-cafe-room-list { max-width: 1400px; margin: 0 auto 1rem; padding: 1rem; border: 1px solid #d6cbbb; background: #fffdf8; overflow: hidden; }.pixel-cafe-room-list-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; min-width: 0; }.pixel-cafe-room-list-heading h2 { margin: .25rem 0 0; font-size: 1.05rem; }.pixel-cafe-room-list-count { flex: 0 0 auto; color: #776e65; font-size: .78rem; }.pixel-cafe-room-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: .75rem; margin-top: .85rem; }.pixel-cafe-room-card { display: grid; gap: .4rem; min-width: 0; padding: .85rem; border: 1px solid #d7cdbf; color: #473d36; background: #fffdf8; text-align: left; cursor: pointer; }.pixel-cafe-room-card:hover, .pixel-cafe-room-card:focus-visible, .pixel-cafe-room-card.active { border-color: #ad7258; box-shadow: 3px 3px 0 rgba(89, 67, 50, .14); outline: 0; }.pixel-cafe-room-card-topline, .pixel-cafe-room-card-stats { display: flex; justify-content: space-between; gap: .45rem; min-width: 0; }.pixel-cafe-room-card-stats { flex-wrap: wrap; }.pixel-cafe-room-card-topline > span, .pixel-cafe-room-card-stats > span { min-width: 0; }.pixel-cafe-room-card-code { color: #8f624f; font: 700 .7rem monospace; overflow-wrap: anywhere; }.pixel-cafe-room-card-state { color: #6f8f78; font: 700 .68rem monospace; text-align: right; }.pixel-cafe-room-card-state.state-blue { color: #6d7fa7; }.pixel-cafe-room-card-state.state-night { color: #a9789f; }.pixel-cafe-room-card strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: .9rem; }.pixel-cafe-room-card-plan { overflow: hidden; color: #776e65; font-size: .72rem; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-room-card-stats { color: #776e65; font-size: .7rem; }.pixel-cafe-room-card-action { margin-top: .2rem; color: #9a644f; font-size: .76rem; font-weight: 700; overflow-wrap: anywhere; }.pixel-cafe-room-list-empty { margin: .85rem 0 .1rem; color: #776e65; font-size: .8rem; }
.pixel-cafe-workbench { display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 1rem; max-width: 1400px; margin: 0 auto; }
.pixel-cafe-scene, .pixel-cafe-inspector, .pixel-cafe-notice { border: 1px solid #d6cbbb; background: #fffdf8; box-shadow: 4px 4px 0 rgba(89, 67, 50, .12); }
.pixel-cafe-scene { position: relative; min-height: 500px; overflow: hidden; background-color: #e7ded0; background-image: linear-gradient(#d8cbbc 1px, transparent 1px), linear-gradient(90deg, #d8cbbc 1px, transparent 1px); background-size: 24px 24px; }
.pixel-cafe-scene-topline { display: flex; justify-content: space-between; padding: .75rem 1rem; border-bottom: 1px solid #cfc1b2; color: #74695d; background: rgba(255, 253, 248, .88); font-size: .78rem; }
.pixel-cafe-status { display: inline-flex; align-items: center; gap: .35rem; }.pixel-cafe-status i { width: .5rem; height: .5rem; border-radius: 999px; background: #c28d4c; }
.pixel-cafe-empty { display: grid; min-height: 240px; padding: 2rem; place-items: center; color: #776e65; text-align: center; }.pixel-cafe-scene-overlay { position: absolute; z-index: 5; inset: 3rem 0 0; min-height: 0; color: #fff6e5; background: rgba(31, 40, 55, .68); text-shadow: 1px 1px 0 rgba(24, 29, 38, .8); }.pixel-cafe-demo-badge { position: absolute; z-index: 6; right: .75rem; bottom: .75rem; margin: 0; padding: .35rem .5rem; border: 1px solid rgba(255, 246, 229, .48); color: #fff6e5; background: rgba(37, 48, 65, .82); font: 700 .68rem/1 monospace; }
.pixel-cafe-error { gap: .75rem; color: #a94d48; }.pixel-cafe-error p { margin: 0; }.pixel-cafe-retry { padding: .5rem .75rem; border: 1px solid #b97867; color: #824d40; background: #fffdf8; cursor: pointer; }
.pixel-cafe-inspector { padding: 1.2rem; }.pixel-cafe-inspector-heading { display: flex; justify-content: space-between; color: #9a6a53; }.pixel-cafe-label { font: 700 .7rem monospace; text-transform: uppercase; }.pixel-cafe-inspector h2 { margin: 1.25rem 0 .3rem; font-size: 1.25rem; }.pixel-cafe-room-code { margin: 0; color: #8c8278; font: .75rem monospace; }.pixel-cafe-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: .5rem; margin: 1.5rem 0; }.pixel-cafe-stats div { padding: .55rem .4rem; border: 1px solid #e0d6c8; text-align: center; }.pixel-cafe-stats dt { color: #8c8278; font-size: .68rem; }.pixel-cafe-stats dd { margin: .25rem 0 0; font-size: .78rem; font-weight: 700; }.pixel-cafe-primary { width: 100%; padding: .7rem; border: 0; color: #fffdf8; background: #a9785d; font-weight: 700; }.pixel-cafe-primary:disabled { cursor: not-allowed; opacity: .72; }.pixel-cafe-muted { color: #82786e; font-size: .86rem; line-height: 1.6; }
.pixel-cafe-seat-picker { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: .45rem; margin: 1rem 0; }.pixel-cafe-seat-button { min-height: 2.2rem; border: 1px solid #c9bdac; color: #5d5148; background: #fffdf8; cursor: pointer; font: 700 .8rem monospace; }.pixel-cafe-seat-button:hover, .pixel-cafe-seat-button:focus-visible, .pixel-cafe-seat-button.active { border-color: #9a644f; color: #fffdf8; background: #a9785d; outline: 0; }.pixel-cafe-single-room-note { margin: 1rem 0; padding: .65rem .7rem; border-left: 3px solid #c28d4c; color: #74695d; background: #f7efe4; font-size: .78rem; line-height: 1.45; }.pixel-cafe-payment-label { display: grid; gap: .35rem; margin: 1rem 0; color: #74695d; font-size: .78rem; }.pixel-cafe-payment-select { width: 100%; min-height: 2.3rem; border: 1px solid #cfc1b2; border-radius: 0; color: #473d36; background: #fffdf8; }.pixel-cafe-agreement { display: flex; align-items: flex-start; gap: .45rem; margin: 1rem 0; color: #74695d; font-size: .76rem; line-height: 1.45; }.pixel-cafe-agreement input { margin-top: .1rem; accent-color: #9a644f; }.pixel-cafe-inline-error { margin: .75rem 0; color: #a94d48; font-size: .78rem; line-height: 1.4; }
.pixel-cafe-notice { display: flex; gap: .75rem; align-items: flex-start; max-width: 1400px; margin: 1rem auto 0; padding: .85rem 1rem; }.pixel-cafe-notice-icon { display: grid; width: 1.8rem; height: 1.8rem; place-items: center; color: #8f624f; background: #f1e0d3; }.pixel-cafe-notice strong { font-size: .82rem; }.pixel-cafe-notice p { margin: .25rem 0 0; color: #776e65; font-size: .78rem; }
@media (max-width: 900px) { .pixel-cafe-workbench { grid-template-columns: 1fr; }.pixel-cafe-inspector { min-height: 0; }.pixel-cafe-scene { min-height: 430px; } }
@media (max-width: 620px) { .pixel-cafe-page { padding: .85rem; }.pixel-cafe-header { align-items: stretch; flex-direction: column; }.pixel-cafe-my-rooms-heading { align-items: flex-start; flex-direction: column; }.pixel-cafe-my-rooms-list { grid-template-columns: 1fr; } }

/* The lobby is the primary interaction surface; room and account details follow it. */
.pixel-cafe-page { position: relative; display: flex; min-height: calc(100vh - 4rem); flex-direction: column; padding: 1.25rem; overflow: hidden; color: #edf4fb; background: #07111e; }
.pixel-cafe-header { position: absolute; z-index: 6; top: 3.4rem; left: clamp(2rem, 5vw, 5rem); width: min(34rem, calc(100% - 5rem)); margin: 0; color: #fff7e5; pointer-events: none; text-shadow: 2px 2px 0 rgba(4, 12, 21, .72); }.pixel-cafe-kicker { margin: 0 0 .55rem; color: #f3c36d; font: 700 .72rem/1 monospace; letter-spacing: .12em; }.pixel-cafe-header h1 { margin: 0; font-size: clamp(2rem, 4vw, 3.2rem); line-height: 1.05; }.pixel-cafe-subtitle { margin: .65rem 0 0; max-width: 30rem; color: #c8d6e2; font-size: .95rem; line-height: 1.5; }
.pixel-cafe-workbench { position: relative; order: 1; display: block; width: min(100%, 1600px); margin: 0 auto; }.pixel-cafe-scene { position: relative; min-height: 0; overflow: hidden; border: 1px solid #385369; background: #0b1d30; box-shadow: 0 14px 34px rgba(0, 0, 0, .32); }.pixel-cafe-scene-topline { position: absolute; z-index: 5; top: 1rem; right: 1rem; display: flex; gap: .75rem; align-items: center; padding: 0; border: 0; color: #c9d9e8; background: transparent; font: 700 .72rem/1 monospace; text-shadow: 1px 1px 0 #06101a; }.pixel-cafe-status { display: inline-flex; gap: .4rem; align-items: center; }.pixel-cafe-status i { width: .5rem; height: .5rem; border-radius: 0; background: #73d7a1; box-shadow: 0 0 0 2px rgba(115, 215, 161, .22); }
.pixel-cafe-zones { position: relative; z-index: 1; display: flex; width: min(100%, 1600px); max-width: 1400px; gap: .45rem; margin: 0 auto 1rem; overflow-x: auto; }.pixel-cafe-zone { display: inline-flex; flex: 0 0 auto; gap: .45rem; align-items: center; min-height: 2.25rem; padding: .45rem .7rem; border: 1px solid rgba(204, 224, 238, .34); color: #dbe9f2; background: rgba(5, 16, 28, .78); cursor: pointer; font-size: .8rem; backdrop-filter: blur(4px); }.pixel-cafe-zone:hover, .pixel-cafe-zone:focus-visible, .pixel-cafe-zone.active { border-color: #f2b968; color: #fff7e5; background: rgba(48, 36, 25, .82); box-shadow: none; outline: 0; }.pixel-cafe-zone-dot { width: .55rem; height: .55rem; box-shadow: 2px 2px 0 rgba(0, 0, 0, .3); }
.pixel-cafe-room-list { width: min(100%, 1600px); max-width: 1400px; border-color: rgba(164, 194, 218, .28); color: #eff5fb; background: #0c1d2e; }.pixel-cafe-room-list-heading h2 { color: #fff7e5; }.pixel-cafe-room-list-count { color: #aebfcd; }.pixel-cafe-room-card { border-color: #385369; color: #eff5fb; background: #102238; }.pixel-cafe-room-card:hover, .pixel-cafe-room-card:focus-visible, .pixel-cafe-room-card.active { border-color: #efbd68; box-shadow: 3px 3px 0 rgba(0, 0, 0, .24); }.pixel-cafe-room-card-code { color: #f1c26f; }.pixel-cafe-room-card-state { color: #73c99a; }.pixel-cafe-room-card-state.state-blue { color: #8eb4ed; }.pixel-cafe-room-card-state.state-night { color: #d0a4e4; }.pixel-cafe-room-card-plan, .pixel-cafe-room-card-stats, .pixel-cafe-room-list-empty { color: #aebfcd; }.pixel-cafe-room-card-action { color: #efbd68; }
.pixel-cafe-front-desk { position: absolute; z-index: 5; bottom: 1rem; left: 1rem; display: flex; gap: .7rem; align-items: center; max-width: 21rem; padding: .75rem .9rem; border: 1px solid rgba(245, 192, 105, .52); color: #f9efdb; background: rgba(26, 18, 13, .84); box-shadow: 4px 4px 0 rgba(0, 0, 0, .24); }.pixel-cafe-front-desk p { margin: 0 0 .2rem; color: #f2bd69; font: 700 .68rem/1 monospace; letter-spacing: .1em; }.pixel-cafe-front-desk strong { font-size: .78rem; line-height: 1.35; }.pixel-cafe-front-desk-lamp { width: .6rem; height: .6rem; flex: 0 0 auto; background: #f2bd69; box-shadow: 0 0 0 3px rgba(242, 189, 105, .18); }
.pixel-cafe-empty { color: #fff6e5; }.pixel-cafe-scene-overlay { z-index: 8; inset: 0; color: #fff6e5; background: rgba(4, 12, 21, .78); text-shadow: 1px 1px 0 #02070d; }.pixel-cafe-demo-badge { z-index: 8; right: 1rem; bottom: 1rem; background: rgba(4, 12, 21, .8); }.pixel-cafe-error { color: #ffc2bd; }.pixel-cafe-retry { border-color: #df947e; color: #fff7e5; background: #6f3d37; }
.pixel-cafe-inspector { position: absolute; z-index: 6; right: 1rem; bottom: 1rem; width: min(19rem, calc(100% - 2rem)); padding: 1rem; border: 1px solid rgba(190, 213, 230, .38); color: #eaf3fa; background: rgba(5, 16, 28, .9); box-shadow: 5px 5px 0 rgba(0, 0, 0, .26); backdrop-filter: blur(5px); }.pixel-cafe-inspector-heading { color: #f1c26f; }.pixel-cafe-inspector h2 { margin: 1rem 0 .3rem; color: #fff7e5; font-size: 1.15rem; }.pixel-cafe-room-code { color: #aebfcd; }.pixel-cafe-stats { margin: 1rem 0; }.pixel-cafe-stats div { border-color: rgba(190, 213, 230, .22); }.pixel-cafe-stats dt { color: #9fb2c1; }.pixel-cafe-primary { border: 1px solid #efbd68; color: #1c120b; background: #efbd68; font-weight: 800; }.pixel-cafe-primary:disabled { opacity: .58; }.pixel-cafe-muted { color: #c4d0da; }.pixel-cafe-seat-button { border-color: #5f788b; color: #dce8f0; background: #122a3e; }.pixel-cafe-seat-button:hover, .pixel-cafe-seat-button:focus-visible, .pixel-cafe-seat-button.active { border-color: #efbd68; color: #1c120b; background: #efbd68; }.pixel-cafe-payment-label, .pixel-cafe-agreement { color: #c8d5df; }.pixel-cafe-payment-select { border-color: #5f788b; color: #eff5fb; background: #102438; }.pixel-cafe-agreement input { accent-color: #efbd68; }.pixel-cafe-inline-error { color: #ffc2bd; }
.pixel-cafe-my-rooms { order: 2; width: min(100%, 1600px); margin: 1.25rem auto 0; padding: 1rem 0 0; border-top: 1px solid rgba(164, 194, 218, .32); border-bottom: 0; background: transparent; }.pixel-cafe-my-rooms-heading h2 { color: #fff7e5; }.pixel-cafe-my-rooms-tabs { border-color: #385369; }.pixel-cafe-my-rooms-tab { border-color: #385369; color: #aec0ce; background: #0c1d2e; }.pixel-cafe-my-rooms-tab.active { color: #1c120b; background: #efbd68; }.pixel-cafe-my-rooms-state { color: #b6c7d4; }.pixel-cafe-my-rooms-error { color: #ffc2bd; }.pixel-cafe-my-rooms-retry { border-color: #df947e; color: #fff7e5; background: #6f3d37; }.pixel-cafe-my-room { border-left-color: #e3a962; background: #102238; }.pixel-cafe-my-room-code, .pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { color: #aebfcd; }.pixel-cafe-my-room-state { border-color: #5f788b; color: #cbd8e2; background: #0b1928; }.pixel-cafe-my-room-state.state-active { border-color: #73c99a; color: #a2e0bc; }.pixel-cafe-my-room-state.state-refunded, .pixel-cafe-my-room-state.state-cancelled, .pixel-cafe-my-room-state.state-released { border-color: #d6807a; color: #ffb7b0; }
.pixel-cafe-notice { order: 3; width: min(100%, 1600px); margin: 1rem auto 0; padding: .85rem 0; border-top: 1px solid rgba(164, 194, 218, .2); color: #bdd0dc; background: transparent; box-shadow: none; }.pixel-cafe-notice-icon { color: #efbd68; background: #102238; }.pixel-cafe-notice strong { color: #eff5fb; }.pixel-cafe-notice p { color: #bdd0dc; }
@media (max-width: 900px) { .pixel-cafe-page { padding: .85rem; }.pixel-cafe-header { position: relative; top: auto; left: auto; order: 0; width: auto; margin: 0 0 .8rem; pointer-events: auto; }.pixel-cafe-zones { position: relative; top: auto; right: auto; order: 0; max-width: none; margin: 0 0 .75rem; }.pixel-cafe-workbench { width: 100%; }.pixel-cafe-scene { overflow: visible; }.pixel-cafe-scene-topline { top: .75rem; right: .75rem; }.pixel-cafe-front-desk { position: relative; right: auto; bottom: auto; left: auto; max-width: none; margin: .75rem 0 0; }.pixel-cafe-inspector { position: relative; right: auto; bottom: auto; width: auto; margin-top: 0; border-top: 0; box-shadow: none; }.pixel-cafe-demo-badge { position: relative; right: auto; bottom: auto; display: table; width: auto; margin: .65rem 0 0 auto; }.pixel-cafe-my-rooms { margin-top: 1rem; } }
.pixel-cafe-single-room-note { border-left-color: #efbd68; color: #c8d5df; background: rgba(65, 46, 26, .66); }
@media (max-width: 620px) { .pixel-cafe-page { padding: .7rem; }.pixel-cafe-header h1 { font-size: 2rem; }.pixel-cafe-front-desk { right: .75rem; bottom: .75rem; left: .75rem; max-width: none; }.pixel-cafe-scene-topline > span:first-child { display: none; }.pixel-cafe-inspector { padding: .85rem; }.pixel-cafe-my-rooms-list { grid-template-columns: 1fr; } }
.pixel-cafe-my-room-account { grid-column: 1 / -1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: #aebfcd; font-size: .7rem; }.pixel-cafe-my-room-empty { color: #8799a8; }
</style>
