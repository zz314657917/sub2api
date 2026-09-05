<template>
  <AppLayout>
    <div class="pixel-cafe-page">
      <header class="pixel-cafe-header">
        <div v-if="pixelCafeHeaderVisible">
          <p class="pixel-cafe-kicker">PIXEL CAFE / BETA</p>
          <h1>{{ pixelCafeTitle }}</h1>
        </div>
      </header>

      <section
        v-if="pixelCafeHeaderVisible && pixelCafeDescription"
        class="pixel-cafe-notice"
        aria-label="像素网吧说明"
        data-testid="pixel-cafe-description"
      >
        <div class="pixel-cafe-notice-icon"><Icon name="infoCircle" size="sm" /></div>
        <p>{{ pixelCafeDescription }}</p>
      </section>

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
            <div class="pixel-cafe-my-room-head">
              <div class="pixel-cafe-my-room-title">
              <strong>{{ membership.room.name }}</strong>
              </div>
              <div
                v-if="myRoomMemberCount(membership) > 0"
                class="pixel-cafe-my-room-members"
                :aria-label="`${myRoomMemberCount(membership)} 人已加入`"
                data-testid="pixel-cafe-my-room-members"
              >
                <span
                  v-for="member in visibleMyRoomMembers(membership)"
                  :key="member.avatar_seed"
                  class="pixel-cafe-my-room-member-avatar"
                  data-testid="pixel-cafe-my-room-member-avatar"
                  aria-hidden="true"
                >
                  <img :src="roomMemberAvatarUrl(member.avatar_seed)" alt="" />
                </span>
                <span v-if="myRoomMemberCount(membership) > MAX_ROOM_MEMBER_AVATARS" class="pixel-cafe-my-room-member-more" aria-hidden="true">
                  +{{ myRoomMemberCount(membership) - MAX_ROOM_MEMBER_AVATARS }}
                </span>
              </div>
            </div>
            <span class="pixel-cafe-my-room-account">绑定账号：{{ myRoomAccountCopy(membership) }}</span>
            <span class="pixel-cafe-my-room-lifetime" data-testid="pixel-cafe-my-room-lifetime">{{ myRoomValidityCopy(membership) }}</span>
            <div
              v-if="membership.managed_api_key && myRoomHasActiveUsage(membership)"
              class="pixel-cafe-my-room-usage"
              data-testid="pixel-cafe-my-room-usage"
            >
              <section class="pixel-cafe-my-room-window" data-testid="pixel-cafe-account-limit">
                <div class="pixel-cafe-my-room-window-heading">
                  <strong>账号 7D 剩余</strong>
                  <span>{{ formatAccountRemaining(membership.account?.remaining_7d_percent) }}</span>
                </div>
                <div
                  :class="['pixel-cafe-my-room-progress', accountRemainingAvailable(membership.account?.remaining_7d_percent) ? `tone-${accountRemainingTone(membership.account?.remaining_7d_percent)}` : 'tone-unavailable']"
                  role="progressbar"
                  aria-label="账号 7D 剩余额度"
                  aria-valuemin="0"
                  aria-valuemax="100"
                  :aria-valuenow="accountRemainingAvailable(membership.account?.remaining_7d_percent) ? membership.account?.remaining_7d_percent : undefined"
                  :aria-valuetext="formatAccountRemaining(membership.account?.remaining_7d_percent)"
                >
                  <span :style="{ width: `${accountRemainingAvailable(membership.account?.remaining_7d_percent) ? membership.account?.remaining_7d_percent : 0}%` }"></span>
                </div>
              </section>
              <section class="pixel-cafe-my-room-window" data-testid="pixel-cafe-my-limit">
                <div class="pixel-cafe-my-room-window-heading">
                  <strong>我的限额</strong>
                  <span>{{ formatWindowQuota(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d) }}</span>
                </div>
                <div
                  :class="['pixel-cafe-my-room-progress', `tone-${usageTone(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d)}`]"
                  role="progressbar"
                  aria-label="我的限额使用进度"
                  aria-valuemin="0"
                  :aria-valuemax="membership.managed_api_key.rate_limit_7d > 0 ? membership.managed_api_key.rate_limit_7d : undefined"
                  :aria-valuenow="membership.managed_api_key.rate_limit_7d > 0 ? Math.min(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d) : undefined"
                  :aria-valuetext="formatWindowQuota(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d)"
                >
                  <span :style="{ width: `${usagePercent(membership.managed_api_key.usage_7d, membership.managed_api_key.rate_limit_7d)}%` }"></span>
                </div>
              </section>
            </div>
            <span v-else class="pixel-cafe-my-room-limit-empty">{{ myRoomLimitFallback(membership) }}</span>
          </li>
        </ul>
      </section>

      <section class="pixel-cafe-workbench" aria-label="像素网吧大厅">
        <div class="pixel-cafe-scene">
          <div class="pixel-cafe-scene-topline">
            <span class="pixel-cafe-status"><i aria-hidden="true"></i>{{ loading ? '加载中' : isLocalDemoMode ? '本地演示' : '实时房间' }}</span>
          </div>
          <CafeScene
            :rooms="rooms"
            :lobby-avatars="sceneAvatars"
            :workstations="sceneWorkstations"
            @select-room="openRoom"
          />
          <section class="pixel-cafe-room-list" aria-label="房间列表" data-testid="pixel-cafe-room-list">
            <div class="pixel-cafe-room-list-heading">
              <div>
                <p class="pixel-cafe-label">ROOMS / 包间</p>
                <h2>全部包间</h2>
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
                <span class="pixel-cafe-room-card-titleline">
                  <strong>{{ room.name }}</strong>
                  <span class="pixel-cafe-room-card-badges" aria-label="订阅套餐">
                    <span class="pixel-cafe-room-card-badge">GPT</span>
                    <span class="pixel-cafe-room-card-badge tier">{{ tierLabel(room) }}</span>
                  </span>
                </span>
                <span class="pixel-cafe-room-card-stats">
                  <span>已售 {{ room.round?.paid_shares ?? 0 }}/{{ room.plan.total_shares }} 份</span>
                  <span>{{ room.round?.joined_buyers ?? 0 }}/{{ room.plan.max_buyers }} 人</span>
                  <span>{{ room.plan.validity_days }} 天</span>
                  <span>{{ room.plan.price_label || `${room.plan.price_per_share} CNY` }}</span>
                </span>
                <span
                  v-if="roomMemberCount(room) > 0"
                  class="pixel-cafe-room-card-members"
                  :aria-label="`${roomMemberCount(room)} 人已加入`"
                  data-testid="pixel-cafe-room-card-members"
                >
                  <span
                    v-for="member in visibleRoomMembers(room)"
                    :key="member.avatar_seed"
                    class="pixel-cafe-room-member-avatar"
                    data-testid="pixel-cafe-room-member-avatar"
                    aria-hidden="true"
                  >
                    <img :src="roomMemberAvatarUrl(member.avatar_seed)" alt="" />
                  </span>
                  <span v-if="roomMemberCount(room) > MAX_ROOM_MEMBER_AVATARS" class="pixel-cafe-room-member-more" aria-hidden="true">
                    +{{ roomMemberCount(room) - MAX_ROOM_MEMBER_AVATARS }}
                  </span>
                </span>
              </button>
            </div>
            <p v-else-if="!loading && !errorMessage" class="pixel-cafe-room-list-empty">当前暂无可加入的包间。</p>
          </section>
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
          <div v-else-if="rooms.length === 0" class="pixel-cafe-empty pixel-cafe-scene-overlay" data-testid="pixel-cafe-empty">当前暂时没有可展示的房间。</div>
          <p v-if="isLocalDemoMode" class="pixel-cafe-demo-badge" data-testid="pixel-cafe-demo-badge">本地演示数据 · 不创建真实订单</p>
        </div>

        <Teleport to="body">
        <div
          v-if="roomDialogOpen && selectedRoom"
          class="pixel-cafe-dialog-backdrop"
          data-testid="pixel-cafe-room-dialog"
          @click.self="closeRoomDialog"
          @keydown.esc="closeRoomDialog"
        >
        <aside
          class="pixel-cafe-inspector"
          role="dialog"
          aria-modal="true"
          aria-labelledby="pixel-cafe-room-dialog-title"
        >
          <div class="pixel-cafe-inspector-heading">
            <span class="pixel-cafe-label">包间详情</span>
            <button
              ref="roomDialogClose"
              type="button"
              class="pixel-cafe-dialog-close"
              aria-label="关闭包间详情"
              @click="closeRoomDialog"
            >×</button>
          </div>
          <template v-if="selectedRoom">
            <h2 id="pixel-cafe-room-dialog-title">{{ selectedRoom.name }}</h2>
            <p class="pixel-cafe-room-code">{{ selectedRoom.code }}</p>
            <dl class="pixel-cafe-stats">
              <div><dt>套餐</dt><dd>ChatGPT {{ tierLabel(selectedRoom) }}</dd></div>
              <div><dt>进度</dt><dd>{{ roomShareLabel(selectedRoom) }}</dd></div>
              <div><dt>周期</dt><dd>{{ selectedRoom.plan.validity_days }} 天</dd></div>
            </dl>
            <section class="pixel-cafe-purchase-limits" data-testid="pixel-cafe-purchase-limits" aria-label="每份额度">
              <div class="pixel-cafe-purchase-limits-heading">
                <strong>每份额度</strong>
                <span v-if="selectedRoom.plan.quota_per_share_label">{{ selectedRoom.plan.quota_per_share_label }}</span>
              </div>
              <dl>
                <div><dt>总额度</dt><dd>{{ formatPurchaseLimit(selectedRoom.plan.room_key_quota_usd) }}</dd></div>
                <div><dt>5H 限额</dt><dd>{{ formatPurchaseLimit(selectedRoom.plan.room_key_rate_limit_5h) }}</dd></div>
                <div><dt>1D 限额</dt><dd>{{ formatPurchaseLimit(selectedRoom.plan.room_key_rate_limit_1d) }}</dd></div>
                <div><dt>7D 限额</dt><dd>{{ formatPurchaseLimit(selectedRoom.plan.room_key_rate_limit_7d) }}</dd></div>
              </dl>
            </section>
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
            <template v-else-if="selectedRoom.purchase_state === 'available' || selectedRoom.purchase_state === 'reserved' || selectedRoom.purchase_state === 'awaiting_payment'">
              <div class="pixel-cafe-seat-picker" aria-label="购买份额">
                <button type="button" class="pixel-cafe-seat-button" :disabled="selectedShareCount <= 1" @click="selectedShareCount -= 1">−</button>
                <strong class="pixel-cafe-share-count" data-testid="pixel-cafe-share-count">{{ selectedShareCount }} 份</strong>
                <button type="button" class="pixel-cafe-seat-button" :disabled="selectedShareCount >= maxPurchasableShares(selectedRoom)" @click="selectedShareCount += 1">+</button>
              </div>
              <p class="pixel-cafe-single-room-note">每份 {{ selectedRoom.plan.price_label || `${selectedRoom.plan.price_per_share} CNY` }}，合计 {{ selectedShareCount * selectedRoom.plan.price_per_share }} CNY。{{ selectedRoom.my_paid_shares ? ` 已持有 ${selectedRoom.my_paid_shares} 份` : '' }}{{ selectedRoom.my_reserved_shares ? `，已预约 ${selectedRoom.my_reserved_shares} 份` : '' }}</p>
              <label v-if="selectedRoom.purchase_state === 'awaiting_payment' || !canReserveRooms()" class="pixel-cafe-payment-label">
                支付方式
                <select v-if="paymentMethods.length" v-model="selectedPaymentMethod" class="pixel-cafe-payment-select">
                  <option v-for="method in paymentMethods" :key="method" :value="method">{{ paymentMethodLabel(method) }}</option>
                </select>
                <span v-else class="pixel-cafe-payment-unavailable">当前没有可用的支付方式，请联系管理员。</span>
              </label>
              <label class="pixel-cafe-agreement">
                <input v-model="agreementAccepted" type="checkbox" />
                <span>我已确认加入该房间，具体开通时间以房间状态为准。</span>
              </label>
              <p v-if="orderError" class="pixel-cafe-inline-error" data-testid="pixel-cafe-order-error">{{ orderError }}</p>
              <button v-if="selectedRoom.purchase_state !== 'awaiting_payment' && canReserveRooms()"
                type="button"
                class="pixel-cafe-primary"
                :disabled="isLocalDemoMode || submitting || selectedShareCount < 1 || !agreementAccepted"
                @click="reserveShares"
              >
                {{ isLocalDemoMode ? '本地演示不创建订单' : submitting ? '正在提交预约' : '预约份额' }}
              </button>
              <button v-else
                type="button"
                class="pixel-cafe-primary"
                :disabled="isLocalDemoMode || submitting || !paymentMethods.length || selectedShareCount < 1 || !agreementAccepted"
                @click="submitOrder"
              >
                {{ isLocalDemoMode ? '本地演示不创建订单' : submitting ? '正在创建订单' : '确认份额并付款' }}
              </button>
            </template>
            <p v-else class="pixel-cafe-muted">{{ roomUnavailableCopy(selectedRoom) }}</p>
          </template>
          <template v-else>
            <h2>选择一间包间</h2>
            <p class="pixel-cafe-muted">房间进度来自服务端；选择包间可查看当前座位与开团状态。</p>
          </template>
        </aside>
        </div>
        </Teleport>
      </section>

    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import CafeScene from './components/CafeScene.vue'
import { createPixelCafeDemoOverview, isLocalPixelCafeDemo } from './demoData'
import { cafeAPI } from '@/api/cafe'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { resolvePixelCafeTitle } from '@/utils/groupBuyProduct'
import avatarGoldUrl from './assets/sprites/avatar-gold.png'
import avatarTealUrl from './assets/sprites/avatar-teal.png'
import avatarWineUrl from './assets/sprites/avatar-wine.png'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  writePaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import type { CafeLobbyActivity, CafeLobbyAvatar, CafeMyRoom, CafeMyRoomFilter, CafePublicMemberAvatar, CafePublicRoom, CreateCafeRoomOrderResult } from '@/types/pixelCafe'

type PaymentOutcome = 'success' | 'cancelled' | 'expired'

const router = useRouter()
const route = useRoute()

const rooms = ref<CafePublicRoom[]>([])
const lobby = ref<CafeLobbyActivity>(emptyLobbyActivity())
const myRooms = ref<CafeMyRoom[]>([])
const myRoomsFilter = ref<CafeMyRoomFilter>('active,waiting')
const myRoomsLoading = ref(false)
const myRoomsError = ref('')
const selectedRoom = ref<CafePublicRoom | null>(null)
const roomDialogOpen = ref(false)
const roomDialogClose = ref<HTMLButtonElement>()
let roomDialogReturnFocus: HTMLElement | null = null
const countdownNow = ref(Date.now())
let countdownTimer: number | undefined
const loading = ref(false)
const errorMessage = ref('')
const selectedShareCount = ref(1)
const selectedPaymentMethod = ref('')
const agreementAccepted = ref(false)
const submitting = ref(false)
const orderError = ref('')
const paymentPhase = ref<'selecting' | 'paying'>('selecting')
const paymentMethods = ref<string[]>([])
const MAX_ROOM_MEMBER_AVATARS = 5
const roomMemberAvatarUrls = [avatarTealUrl, avatarGoldUrl, avatarWineUrl] as const
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
const sceneWorkstations = computed(() => appStore.cachedPublicSettings?.pixel_cafe_workstation_layout || [])
const isLocalDemoMode = computed(() => isLocalPixelCafeDemo(route.query))
const sceneAvatars = computed<CafeLobbyAvatar[]>(() => {
  if (lobby.value.available && lobby.value.avatars.length > 0) return lobby.value.avatars

  return rooms.value.flatMap((room) => (room.member_avatars || [])
    .map((member, index) => ({
      avatar_seed: member.avatar_seed || `room-${room.id}-member-${index}`,
      seat_index: (room.id * 17 + index) % 12,
      activity: 'recent' as const,
    })),
  ).slice(0, 12)
})

async function loadOverview(): Promise<void> {
  loading.value = true
  errorMessage.value = ''
  selectedRoom.value = null
  try {
    const overview = isLocalDemoMode.value
      ? createPixelCafeDemoOverview()
      : (await cafeAPI.overview({ room_limit: 24 })).data
    lobby.value = normalizeLobbyActivity(overview.lobby)
    rooms.value = Array.isArray(overview.rooms) ? overview.rooms : []
  } catch (error) {
    rooms.value = []
    errorMessage.value = extractApiErrorMessage(error, '加载房间失败')
  } finally {
    loading.value = false
  }
}

async function loadPaymentMethods(): Promise<void> {
  // 演示模式不创建订单，保留一个示例渠道用于展示支付区布局。
  if (isLocalDemoMode.value) {
    paymentMethods.value = ['alipay']
    selectedPaymentMethod.value = 'alipay'
    return
  }

  try {
    const response = await paymentAPI.getCheckoutInfo()
    const available = Object.fromEntries(
      Object.entries(response.data?.methods || {}).filter(([, limit]) => limit?.available !== false),
    )
    const visible = getVisibleMethods(available)
    const order = ['alipay', 'wxpay', 'stripe', 'airwallex']
    paymentMethods.value = Object.entries(visible)
      .map(([method]) => method)
      .sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
    if (!paymentMethods.value.includes(selectedPaymentMethod.value)) {
      selectedPaymentMethod.value = paymentMethods.value[0] || ''
    }
  } catch (error) {
    paymentMethods.value = []
    selectedPaymentMethod.value = ''
    console.warn('[pixel-cafe] Failed to load available payment methods:', error)
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

function roomShareLabel(room: CafePublicRoom): string { return !room.round ? `${room.plan.total_shares} 份 · 暂未开放` : `已售 ${room.round.paid_shares}/${room.plan.total_shares} 份 · ${room.round.joined_buyers}/${room.plan.max_buyers} 人` }

function roomProgressLabel(room: CafePublicRoom): string {
  if (room.round?.status === 'awaiting_account') return '待配号'
  if (room.round?.status === 'activating') return '开通中'
  if (room.round?.status === 'active') return '已开通'
  if (room.round?.status === 'refunding') return '退款中'
  if (room.round?.status === 'refunded') return '已退款'
  if (room.round?.status === 'open') return '可购买'
  return room.purchase_state === 'buyers_full' ? '人数已满' : '暂不可用'
}

function roomTone(room: CafePublicRoom): string {
  if (room.zone_key === 'openai' || room.theme_key.includes('green')) return 'green'
  if (room.zone_key === 'gemini' || room.theme_key.includes('blue')) return 'blue'
  if (room.purchase_state === 'unavailable') return 'night'
  return 'wood'
}

function tierLabel(room: CafePublicRoom): string { return room.plan.subscription_tier === 'pro' ? 'PRO' : 'PLUS' }
function canReserveRooms(): boolean { return typeof cafeAPI.reserveShares === 'function' }
function formatPurchaseLimit(value?: number): string {
  if (value === undefined || value === null || !Number.isFinite(value)) return '暂未配置'
  if (value <= 0) return '不限'
  return `${value.toFixed(2)} USD`
}
function roomMembers(room: CafePublicRoom): CafePublicMemberAvatar[] { return room.member_avatars || [] }

function roomMemberCount(room: CafePublicRoom): number {
  return roomMembers(room).length
}

function visibleRoomMembers(room: CafePublicRoom): CafePublicMemberAvatar[] {
  return roomMembers(room).slice(0, MAX_ROOM_MEMBER_AVATARS)
}

function roomMemberAvatarUrl(seed: string): string {
  let hash = 0
  for (let index = 0; index < seed.length; index += 1) hash = ((hash * 31) + seed.charCodeAt(index)) >>> 0
  return roomMemberAvatarUrls[hash % roomMemberAvatarUrls.length]
}

function myRoomMembers(membership: CafeMyRoom): CafePublicMemberAvatar[] { return membership.member_avatars || [] }
function myRoomMemberCount(membership: CafeMyRoom): number { return myRoomMembers(membership).length }
function visibleMyRoomMembers(membership: CafeMyRoom): CafePublicMemberAvatar[] { return myRoomMembers(membership).slice(0, MAX_ROOM_MEMBER_AVATARS) }

function formatWindowQuota(used: number, limit: number): string {
  if (limit <= 0) return `${used.toFixed(2)} / 不限`
  return `${used.toFixed(2)} / ${limit.toFixed(2)}`
}

function accountRemainingAvailable(value?: number): value is number {
  return Number.isFinite(value) && value! >= 0 && value! <= 100
}

function formatAccountRemaining(value?: number): string {
  return accountRemainingAvailable(value) ? `${value.toFixed(2)}%` : '暂不可用'
}

function accountRemainingTone(value?: number): 'normal' | 'warning' | 'danger' {
  if (!accountRemainingAvailable(value)) return 'normal'
  if (value <= 10) return 'danger'
  if (value <= 30) return 'warning'
  return 'normal'
}

function usagePercent(used: number, limit: number): number {
  if (!Number.isFinite(used) || !Number.isFinite(limit) || limit <= 0) return 0
  return Math.min(100, Math.max(0, (used / limit) * 100))
}

function usageTone(used: number, limit: number): 'normal' | 'warning' | 'danger' | 'unlimited' {
  if (limit <= 0) return 'unlimited'
  const percent = usagePercent(used, limit)
  if (percent >= 90) return 'danger'
  if (percent >= 70) return 'warning'
  return 'normal'
}

function parseTimestamp(value?: string): number | null {
  if (!value) return null
  const timestamp = Date.parse(value)
  return Number.isFinite(timestamp) ? timestamp : null
}

function formatRemaining(value: string): string {
  const timestamp = parseTimestamp(value)
  if (timestamp === null) return '时间未知'
  const remainingMinutes = Math.max(0, Math.ceil((timestamp - countdownNow.value) / 60_000))
  const days = Math.floor(remainingMinutes / (24 * 60))
  const hours = Math.floor((remainingMinutes % (24 * 60)) / 60)
  const minutes = remainingMinutes % 60
  if (days > 0) return `${days}天${hours}小时`
  if (hours > 0) return `${hours}小时${minutes}分钟`
  return `${remainingMinutes}分钟`
}

function myRoomWithinValidity(membership: CafeMyRoom): boolean {
  const expiresAt = parseTimestamp(membership.expires_at)
  return expiresAt === null || expiresAt > countdownNow.value
}

function myRoomHasActiveUsage(membership: CafeMyRoom): boolean {
  return myRoomStatus(membership) === 'active' && membership.round.status === 'active' && Boolean(membership.account) && Boolean(membership.managed_api_key) && myRoomWithinValidity(membership)
}

function myRoomValidityCopy(membership: CafeMyRoom): string {
  if (membership.round.status !== 'active') {
    const terminalStatuses = new Set(['released', 'cancelled', 'refund_pending', 'refund_processing', 'refunded'])
    return terminalStatuses.has(myRoomStatus(membership)) ? '剩余时间：已结束' : '剩余时间：开通后开始计算'
  }
  if (!membership.expires_at) return '剩余时间：长期有效'
  if (!myRoomWithinValidity(membership)) return '剩余时间：已到期'
  return `剩余时间：${formatRemaining(membership.expires_at)}`
}

function myRoomLimitFallback(membership: CafeMyRoom): string {
  if (membership.round.status !== 'active') return '我的限额：开通后显示'
  if (!myRoomWithinValidity(membership)) return '我的限额：已到期'
  return '我的限额：暂不可用'
}

function myRoomStatus(membership: CafeMyRoom): string { return membership.status || membership.round.status }
function myRoomWaitingCopy(membership: CafeMyRoom): string {
  if (membership.round.status === 'awaiting_account') return '已成团，预计 24 小时内开通'
  if (membership.round.status === 'refunding') return '配号超时，退款处理中'
  if (membership.round.status === 'refunded') return '已退款'
  if (membership.round.status === 'active') return '账号信息暂不可用'
  return '等待成团'
}
function myRoomAccountCopy(membership: CafeMyRoom): string {
  if (membership.round.status === 'active' && membership.account) return membership.account.name
  return myRoomWaitingCopy(membership)
}
function maxPurchasableShares(room: CafePublicRoom): number {
  if (room.purchase_state === 'awaiting_payment') return Math.max(1, room.my_reserved_shares ?? 0)
  const remaining = room.round?.remaining_shares ?? 0
  const mine = (room.my_paid_shares ?? 0) + (room.my_reserved_shares ?? 0)
  return Math.max(1, Math.min(remaining, room.plan.max_shares_per_user - mine))
}
function roomUnavailableCopy(room: CafePublicRoom): string {
	if (room.round?.status === 'awaiting_payment') return room.my_reserved_shares ? '已成团，请选择支付方式完成付款。' : '该房间已成团，正在等待参与者付款。'
  if (room.round?.status === 'awaiting_account') return '已成团，管理员正在配号，预计 24 小时内开通。'
  if (room.round?.status === 'refunding') return '配号未在期限内完成，正在退款。'
  if (room.round?.status === 'refunded') return '该轮次已退款。'
  if (room.purchase_state === 'buyers_full') return '参与人数已满；已参与用户仍可在成团前补份。'
  return '当前房间暂不接受购买。'
}

async function reserveShares(): Promise<void> {
  const room = selectedRoom.value
  if (!room || selectedShareCount.value < 1 || !agreementAccepted.value || submitting.value) return
  submitting.value = true
  orderError.value = ''
  try {
    await cafeAPI.reserveShares(room.id, { share_count: selectedShareCount.value, agreement_accepted: true })
    await Promise.all([loadOverview(), loadMyRooms()])
    closeRoomDialog()
  } catch (error) {
    orderError.value = extractApiErrorMessage(error, '预约失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}

function openRoom(room: CafePublicRoom): void {
  const preservePayment = selectedRoom.value?.id === room.id && paymentPhase.value === 'paying'
  roomDialogReturnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
  selectedRoom.value = room
  roomDialogOpen.value = true
  if (!preservePayment) {
    selectedShareCount.value = room.purchase_state === 'awaiting_payment' ? Math.max(1, room.my_reserved_shares ?? 1) : 1
    agreementAccepted.value = false
    orderError.value = ''
    paymentPhase.value = 'selecting'
  }
  void nextTick(() => roomDialogClose.value?.focus())
}

function closeRoomDialog(): void {
  roomDialogOpen.value = false
  void nextTick(() => roomDialogReturnFocus?.focus())
}

async function submitOrder(): Promise<void> {
  const room = selectedRoom.value
  if (!room || selectedShareCount.value < 1 || !agreementAccepted.value || submitting.value) return
  submitting.value = true
  orderError.value = ''
  const visibleMethod = normalizeVisibleMethod(selectedPaymentMethod.value) || selectedPaymentMethod.value
  try {
    const result = (await cafeAPI.createOrder(room.id, {
      share_count: selectedShareCount.value,
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
    orderError.value = extractApiErrorMessage(error, '创建份额订单失败')
  } finally {
    submitting.value = false
  }
}

function onPaymentDone(): void {
  paymentPhase.value = 'selecting'
  selectedShareCount.value = 1
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
  countdownTimer = window.setInterval(() => { countdownNow.value = Date.now() }, 30_000)
  void loadOverview()
  void loadPaymentMethods()
  void loadMyRooms()
})

onUnmounted(() => {
  if (countdownTimer !== undefined) window.clearInterval(countdownTimer)
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
.pixel-cafe-room-list { max-width: 1400px; margin: 0 auto 1rem; padding: 1rem; border: 1px solid #d6cbbb; background: #fffdf8; overflow: hidden; }.pixel-cafe-room-list-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; min-width: 0; }.pixel-cafe-room-list-heading h2 { margin: .25rem 0 0; font-size: 1.05rem; }.pixel-cafe-room-list-count { flex: 0 0 auto; color: #776e65; font-size: .78rem; }.pixel-cafe-room-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: .75rem; margin-top: .85rem; }.pixel-cafe-room-card { display: grid; gap: .4rem; min-width: 0; padding: .85rem; border: 1px solid #d7cdbf; color: #473d36; background: #fffdf8; text-align: left; cursor: pointer; }.pixel-cafe-room-card:hover, .pixel-cafe-room-card:focus-visible, .pixel-cafe-room-card.active { border-color: #ad7258; box-shadow: 3px 3px 0 rgba(89, 67, 50, .14); outline: 0; }.pixel-cafe-room-card-topline, .pixel-cafe-room-card-stats { display: flex; justify-content: space-between; gap: .45rem; min-width: 0; }.pixel-cafe-room-card-stats { flex-wrap: wrap; }.pixel-cafe-room-card-topline > span, .pixel-cafe-room-card-stats > span { min-width: 0; }.pixel-cafe-room-card-code { color: #8f624f; font: 700 .7rem monospace; overflow-wrap: anywhere; }.pixel-cafe-room-card-state { color: #6f8f78; font: 700 .68rem monospace; text-align: right; }.pixel-cafe-room-card-state.state-blue { color: #6d7fa7; }.pixel-cafe-room-card-state.state-night { color: #a9789f; }.pixel-cafe-room-card strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: .9rem; }.pixel-cafe-room-card-plan { overflow: hidden; color: #776e65; font-size: .72rem; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-room-card-stats { color: #776e65; font-size: .7rem; }.pixel-cafe-room-card-action { margin-top: .2rem; color: #9a644f; font-size: .76rem; font-weight: 700; overflow-wrap: anywhere; }.pixel-cafe-room-list-empty { margin: .85rem 0 .1rem; color: #776e65; font-size: .8rem; }
.pixel-cafe-workbench { display: grid; grid-template-columns: minmax(0, 1fr) 280px; gap: 1rem; max-width: 1400px; margin: 0 auto; }
.pixel-cafe-scene, .pixel-cafe-inspector, .pixel-cafe-notice { border: 1px solid #d6cbbb; background: #fffdf8; box-shadow: 4px 4px 0 rgba(89, 67, 50, .12); }
.pixel-cafe-scene { position: relative; min-height: 500px; overflow: hidden; background-color: #e7ded0; background-image: linear-gradient(#d8cbbc 1px, transparent 1px), linear-gradient(90deg, #d8cbbc 1px, transparent 1px); background-size: 24px 24px; }
.pixel-cafe-scene-topline { display: flex; justify-content: space-between; padding: .75rem 1rem; border-bottom: 1px solid #cfc1b2; color: #74695d; background: rgba(255, 253, 248, .88); font-size: .78rem; }
.pixel-cafe-status { display: inline-flex; align-items: center; gap: .35rem; }.pixel-cafe-status i { width: .5rem; height: .5rem; border-radius: 999px; background: #c28d4c; }
.pixel-cafe-empty { display: grid; min-height: 240px; padding: 2rem; place-items: center; color: #776e65; text-align: center; }.pixel-cafe-scene-overlay { position: absolute; z-index: 5; inset: 3rem 0 0; min-height: 0; color: #fff6e5; background: rgba(31, 40, 55, .68); text-shadow: 1px 1px 0 rgba(24, 29, 38, .8); }.pixel-cafe-demo-badge { position: absolute; z-index: 6; right: .75rem; bottom: .75rem; margin: 0; padding: .35rem .5rem; border: 1px solid rgba(255, 246, 229, .48); color: #fff6e5; background: rgba(37, 48, 65, .82); font: 700 .68rem/1 monospace; }
.pixel-cafe-error { gap: .75rem; color: #a94d48; }.pixel-cafe-error p { margin: 0; }.pixel-cafe-retry { padding: .5rem .75rem; border: 1px solid #b97867; color: #824d40; background: #fffdf8; cursor: pointer; }
.pixel-cafe-dialog-backdrop { position: fixed; z-index: 1000; inset: 0; display: grid; padding: 1rem; place-items: center; background: rgba(2, 8, 15, .74); backdrop-filter: blur(5px); }.pixel-cafe-inspector { width: min(34rem, 100%); max-height: calc(100dvh - 2rem); overflow-y: auto; padding: 1.2rem; }.pixel-cafe-inspector-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; color: #9a6a53; }.pixel-cafe-dialog-close { display: grid; width: 2rem; height: 2rem; flex: 0 0 auto; padding: 0; border: 1px solid currentColor; color: inherit; background: transparent; cursor: pointer; place-items: center; font: 700 1.2rem/1 monospace; }.pixel-cafe-dialog-close:hover, .pixel-cafe-dialog-close:focus-visible { color: #fff7e5; border-color: #efbd68; outline: 2px solid rgba(239, 189, 104, .28); outline-offset: 2px; }.pixel-cafe-label { font: 700 .7rem monospace; text-transform: uppercase; }.pixel-cafe-inspector h2 { margin: 1.25rem 0 .3rem; font-size: 1.25rem; }.pixel-cafe-room-code { margin: 0; color: #8c8278; font: .75rem monospace; }.pixel-cafe-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: .5rem; margin: 1.5rem 0; }.pixel-cafe-stats div { padding: .55rem .4rem; border: 1px solid #e0d6c8; text-align: center; }.pixel-cafe-stats dt { color: #8c8278; font-size: .68rem; }.pixel-cafe-stats dd { margin: .25rem 0 0; font-size: .78rem; font-weight: 700; }.pixel-cafe-primary { width: 100%; padding: .7rem; border: 0; color: #fffdf8; background: #a9785d; font-weight: 700; }.pixel-cafe-primary:disabled { cursor: not-allowed; opacity: .72; }.pixel-cafe-muted { color: #82786e; font-size: .86rem; line-height: 1.6; }
.pixel-cafe-seat-picker { display: grid; grid-template-columns: repeat(3, minmax(0, 5.9rem)); justify-content: center; align-items: stretch; gap: .45rem; margin: 1rem 0; }.pixel-cafe-seat-button { min-height: 2.2rem; border: 1px solid #c9bdac; color: #5d5148; background: #fffdf8; cursor: pointer; font: 700 .8rem monospace; }.pixel-cafe-share-count { display: grid; min-height: 2.2rem; border: 1px solid #c9bdac; color: #5d5148; background: #f7efe4; place-items: center; font: 700 .8rem/1 monospace; white-space: nowrap; }.pixel-cafe-seat-button:hover, .pixel-cafe-seat-button:focus-visible, .pixel-cafe-seat-button.active { border-color: #9a644f; color: #fffdf8; background: #a9785d; outline: 0; }.pixel-cafe-single-room-note { margin: 1rem 0; padding: .65rem .7rem; border-left: 3px solid #c28d4c; color: #74695d; background: #f7efe4; font-size: .78rem; line-height: 1.45; }.pixel-cafe-payment-label { display: grid; gap: .35rem; margin: 1rem 0; color: #74695d; font-size: .78rem; }.pixel-cafe-payment-select { width: 100%; min-height: 2.3rem; border: 1px solid #cfc1b2; border-radius: 0; color: #473d36; background: #fffdf8; }.pixel-cafe-agreement { display: flex; align-items: flex-start; gap: .45rem; margin: 1rem 0; color: #74695d; font-size: .76rem; line-height: 1.45; }.pixel-cafe-agreement input { margin-top: .1rem; accent-color: #9a644f; }.pixel-cafe-inline-error { margin: .75rem 0; color: #a94d48; font-size: .78rem; line-height: 1.4; }
.pixel-cafe-payment-unavailable { color: #aebfcd; font-size: .75rem; line-height: 1.4; }
.pixel-cafe-notice { display: flex; gap: .75rem; align-items: flex-start; max-width: 1400px; margin: 1rem auto 0; padding: .85rem 1rem; }.pixel-cafe-notice-icon { display: grid; width: 1.8rem; height: 1.8rem; place-items: center; color: #8f624f; background: #f1e0d3; }.pixel-cafe-notice strong { font-size: .82rem; }.pixel-cafe-notice p { margin: .25rem 0 0; color: #776e65; font-size: .78rem; }
@media (max-width: 900px) { .pixel-cafe-workbench { grid-template-columns: 1fr; }.pixel-cafe-inspector { min-height: 0; }.pixel-cafe-scene { min-height: 430px; } }
@media (max-width: 620px) { .pixel-cafe-page { padding: .85rem; }.pixel-cafe-header { align-items: stretch; flex-direction: column; }.pixel-cafe-my-rooms-heading { align-items: flex-start; flex-direction: column; }.pixel-cafe-my-rooms-list { grid-template-columns: 1fr; } }

/* The lobby is the primary interaction surface; room and account details follow it. */
.pixel-cafe-page { position: relative; display: flex; min-height: calc(100vh - 4rem); flex-direction: column; padding: 1.25rem; overflow: hidden; color: #edf4fb; background: #07111e; }
.pixel-cafe-header { position: relative; z-index: 1; order: 0; width: min(100%, 1600px); margin: 0 auto 1rem; color: #fff7e5; pointer-events: auto; }.pixel-cafe-kicker { margin: 0 0 .45rem; color: #f3c36d; font: 700 .68rem/1 monospace; letter-spacing: .12em; }.pixel-cafe-header h1 { margin: 0; font-size: clamp(1.55rem, 3vw, 2.25rem); line-height: 1.05; }.pixel-cafe-subtitle { margin: .5rem 0 0; max-width: 38rem; color: #c8d6e2; font-size: .88rem; line-height: 1.45; }
.pixel-cafe-workbench { position: relative; order: 1; display: block; width: min(100%, 1600px); margin: 0 auto; }.pixel-cafe-scene { position: relative; min-height: 0; overflow: hidden; border: 1px solid #385369; background: #0b1d30; box-shadow: 0 14px 34px rgba(0, 0, 0, .32); }.pixel-cafe-scene-topline { position: absolute; z-index: 5; top: 1rem; right: 1rem; display: flex; gap: .75rem; align-items: center; padding: 0; border: 0; color: #c9d9e8; background: transparent; font: 700 .72rem/1 monospace; text-shadow: 1px 1px 0 #06101a; }.pixel-cafe-status { display: inline-flex; gap: .4rem; align-items: center; }.pixel-cafe-status i { width: .5rem; height: .5rem; border-radius: 0; background: #73d7a1; box-shadow: 0 0 0 2px rgba(115, 215, 161, .22); }
.pixel-cafe-room-list { position: absolute; z-index: 6; top: 3rem; right: 1rem; display: flex; width: clamp(19rem, 27vw, 23rem); max-width: none; max-height: calc(100% - 6.75rem); flex-direction: column; margin: 0; padding: .75rem; border-color: rgba(164, 194, 218, .42); color: #eff5fb; background: rgba(7, 23, 39, .88); box-shadow: 6px 6px 0 rgba(0, 0, 0, .28); backdrop-filter: blur(8px); }.pixel-cafe-room-list-heading { flex: 0 0 auto; }.pixel-cafe-room-list-heading h2 { color: #fff7e5; }.pixel-cafe-room-list-count { color: #aebfcd; }.pixel-cafe-room-cards { min-height: 0; grid-template-columns: 1fr; margin-top: .65rem; padding: 0 .2rem .2rem 0; overflow-x: hidden; overflow-y: auto; overscroll-behavior: contain; scrollbar-color: #5f788b rgba(6, 17, 29, .6); scrollbar-width: thin; }.pixel-cafe-room-card { scroll-margin-block: .5rem; border-color: #385369; color: #eff5fb; background: rgba(16, 34, 56, .94); }.pixel-cafe-room-card:hover, .pixel-cafe-room-card:focus-visible, .pixel-cafe-room-card.active { border-color: #efbd68; box-shadow: 3px 3px 0 rgba(0, 0, 0, .24); }.pixel-cafe-room-card-code { color: #f1c26f; }.pixel-cafe-room-card-state { color: #73c99a; }.pixel-cafe-room-card-state.state-blue { color: #8eb4ed; }.pixel-cafe-room-card-state.state-night { color: #d0a4e4; }.pixel-cafe-room-card-plan, .pixel-cafe-room-card-stats, .pixel-cafe-room-list-empty { color: #aebfcd; }.pixel-cafe-room-card-action { color: #efbd68; }
.pixel-cafe-front-desk { position: absolute; z-index: 5; bottom: 1rem; left: 1rem; display: flex; gap: .7rem; align-items: center; max-width: 21rem; padding: .75rem .9rem; border: 1px solid rgba(245, 192, 105, .52); color: #f9efdb; background: rgba(26, 18, 13, .84); box-shadow: 4px 4px 0 rgba(0, 0, 0, .24); }.pixel-cafe-front-desk p { margin: 0 0 .2rem; color: #f2bd69; font: 700 .68rem/1 monospace; letter-spacing: .1em; }.pixel-cafe-front-desk strong { font-size: .78rem; line-height: 1.35; }.pixel-cafe-front-desk-lamp { width: .6rem; height: .6rem; flex: 0 0 auto; background: #f2bd69; box-shadow: 0 0 0 3px rgba(242, 189, 105, .18); }
.pixel-cafe-empty { color: #fff6e5; }.pixel-cafe-scene-overlay { z-index: 8; inset: 0; color: #fff6e5; background: rgba(4, 12, 21, .78); text-shadow: 1px 1px 0 #02070d; }.pixel-cafe-demo-badge { z-index: 8; right: 1rem; bottom: 1rem; background: rgba(4, 12, 21, .8); }.pixel-cafe-error { color: #ffc2bd; }.pixel-cafe-retry { border-color: #df947e; color: #fff7e5; background: #6f3d37; }
.pixel-cafe-inspector { position: relative; border: 1px solid rgba(190, 213, 230, .38); color: #eaf3fa; background: #071727; box-shadow: 7px 7px 0 rgba(0, 0, 0, .34); }.pixel-cafe-inspector-heading { color: #f1c26f; }.pixel-cafe-inspector h2 { margin: 1rem 0 .3rem; color: #fff7e5; font-size: 1.15rem; }.pixel-cafe-room-code { color: #aebfcd; }.pixel-cafe-stats { margin: 1rem 0; }.pixel-cafe-stats div { border-color: rgba(190, 213, 230, .22); }.pixel-cafe-stats dt { color: #9fb2c1; }.pixel-cafe-primary { border: 1px solid #efbd68; color: #1c120b; background: #efbd68; font-weight: 800; }.pixel-cafe-primary:disabled { opacity: .58; }.pixel-cafe-muted { color: #c4d0da; }.pixel-cafe-seat-button { border-color: #5f788b; color: #dce8f0; background: #122a3e; }.pixel-cafe-share-count { border-color: #5f788b; color: #eaf3fa; background: #0b1928; }.pixel-cafe-seat-button:hover, .pixel-cafe-seat-button:focus-visible, .pixel-cafe-seat-button.active { border-color: #efbd68; color: #1c120b; background: #efbd68; }.pixel-cafe-payment-label, .pixel-cafe-agreement { color: #c8d5df; }.pixel-cafe-payment-select { border-color: #5f788b; color: #eff5fb; background: #102438; }.pixel-cafe-agreement input { accent-color: #efbd68; }.pixel-cafe-inline-error { color: #ffc2bd; }
.pixel-cafe-my-rooms { order: 2; width: min(100%, 1600px); margin: 1.25rem auto 0; padding: 1rem 0 0; border-top: 1px solid rgba(164, 194, 218, .32); border-bottom: 0; background: transparent; }.pixel-cafe-my-rooms-heading h2 { color: #fff7e5; }.pixel-cafe-my-rooms-tabs { border-color: #385369; }.pixel-cafe-my-rooms-tab { border-color: #385369; color: #aec0ce; background: #0c1d2e; }.pixel-cafe-my-rooms-tab.active { color: #1c120b; background: #efbd68; }.pixel-cafe-my-rooms-state { color: #b6c7d4; }.pixel-cafe-my-rooms-error { color: #ffc2bd; }.pixel-cafe-my-rooms-retry { border-color: #df947e; color: #fff7e5; background: #6f3d37; }.pixel-cafe-my-room { border-left-color: #e3a962; background: #102238; }.pixel-cafe-my-room-code, .pixel-cafe-my-room-meta, .pixel-cafe-my-room-key { color: #aebfcd; }.pixel-cafe-my-room-state { border-color: #5f788b; color: #cbd8e2; background: #0b1928; }.pixel-cafe-my-room-state.state-active { border-color: #73c99a; color: #a2e0bc; }.pixel-cafe-my-room-state.state-refunded, .pixel-cafe-my-room-state.state-cancelled, .pixel-cafe-my-room-state.state-released { border-color: #d6807a; color: #ffb7b0; }
.pixel-cafe-notice { order: 0; width: min(100%, 1600px); margin: 0 auto 1rem; padding: .8rem 1rem; border: 1px solid rgba(164, 194, 218, .28); color: #bdd0dc; background: #0c1d2e; box-shadow: none; }.pixel-cafe-notice-icon { flex: 0 0 auto; color: #efbd68; background: #102238; }.pixel-cafe-notice p { margin: .1rem 0 0; color: #bdd0dc; font-size: .82rem; line-height: 1.55; white-space: pre-line; }
@media (max-width: 900px) { .pixel-cafe-page { padding: .85rem; }.pixel-cafe-header { position: relative; top: auto; left: auto; order: 0; width: auto; margin: 0 0 .8rem; pointer-events: auto; }.pixel-cafe-workbench { width: 100%; }.pixel-cafe-scene { overflow: hidden; }.pixel-cafe-scene-topline { top: .75rem; right: .75rem; }.pixel-cafe-room-list { top: auto; right: .75rem; bottom: .75rem; left: .75rem; width: auto; height: min(10rem, calc(100% - 4.5rem)); max-height: none; padding: .6rem; }.pixel-cafe-room-list-heading { min-height: 1.65rem; }.pixel-cafe-room-list-heading h2 { display: inline; margin: 0; font-size: .9rem; }.pixel-cafe-room-list-heading .pixel-cafe-label { display: none; }.pixel-cafe-room-list-count { font-size: .68rem; }.pixel-cafe-room-cards { display: flex; gap: .55rem; margin-top: .4rem; padding: 0 0 .2rem; overflow-x: auto; overflow-y: hidden; scroll-snap-type: x proximity; touch-action: pan-x; }.pixel-cafe-room-card { width: min(72vw, 17rem); height: 100%; flex: 0 0 min(72vw, 17rem); padding: .6rem; scroll-snap-align: start; }.pixel-cafe-front-desk { top: .75rem; right: auto; bottom: auto; left: .75rem; max-width: 15rem; margin: 0; padding: .45rem .55rem; }.pixel-cafe-front-desk strong { font-size: .68rem; }.pixel-cafe-inspector { width: 100%; }.pixel-cafe-demo-badge { top: 2.85rem; right: .75rem; bottom: auto; margin: 0; }.pixel-cafe-my-rooms { margin-top: 1rem; } }
.pixel-cafe-single-room-note { border-left-color: #efbd68; color: #c8d5df; background: rgba(65, 46, 26, .66); }
@media (max-width: 620px) { .pixel-cafe-page { padding: .7rem; }.pixel-cafe-header h1 { font-size: 2rem; }.pixel-cafe-scene-topline { top: .55rem; right: .55rem; }.pixel-cafe-room-list { right: .55rem; bottom: .55rem; left: .55rem; height: min(8.25rem, calc(100% - 3.8rem)); padding: .45rem; }.pixel-cafe-room-list-heading { min-height: 1.35rem; }.pixel-cafe-room-cards { gap: .4rem; margin-top: .25rem; }.pixel-cafe-room-card { width: min(76vw, 16rem); flex-basis: min(76vw, 16rem); gap: .2rem; padding: .42rem .5rem; }.pixel-cafe-room-card strong { font-size: .76rem; }.pixel-cafe-room-card-stats { flex-wrap: nowrap; gap: .35rem; font-size: .61rem; white-space: nowrap; }.pixel-cafe-room-card-code, .pixel-cafe-room-card-state { font-size: .61rem; }.pixel-cafe-room-card-members { min-height: 1.25rem; padding-top: 0; }.pixel-cafe-room-member-avatar { width: 1.2rem; height: 1.2rem; }.pixel-cafe-room-member-more { min-width: 1.2rem; height: 1.2rem; font-size: .55rem; }.pixel-cafe-front-desk { top: .55rem; right: auto; bottom: auto; left: .55rem; max-width: 5rem; padding: .3rem .4rem; }.pixel-cafe-front-desk-lamp, .pixel-cafe-front-desk strong { display: none; }.pixel-cafe-front-desk p { margin: 0; font-size: .61rem; }.pixel-cafe-demo-badge { top: 2.15rem; right: .55rem; padding: .25rem .35rem; font-size: .56rem; }.pixel-cafe-dialog-backdrop { padding: .5rem; }.pixel-cafe-inspector { max-height: calc(100dvh - 1rem); padding: .85rem; }.pixel-cafe-my-rooms-list { grid-template-columns: 1fr; } }
.pixel-cafe-my-room { grid-template-columns: minmax(0, 1fr); gap: .4rem; min-width: 0; }.pixel-cafe-my-room-head { display: flex; min-width: 0; align-items: flex-start; justify-content: space-between; gap: .7rem; }.pixel-cafe-my-room-account, .pixel-cafe-my-room-lifetime, .pixel-cafe-my-room-limit-empty { overflow: hidden; color: #aebfcd; font-size: .7rem; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-my-room-lifetime { color: #bfd0dc; }.pixel-cafe-my-room-limit-empty { color: #8799a8; }.pixel-cafe-my-room-members { display: flex; flex: 0 0 auto; min-height: 1.7rem; align-items: center; padding: 0 0 0 .25rem; }.pixel-cafe-my-room-member-avatar { display: grid; width: 1.55rem; height: 1.55rem; margin-left: -.3rem; overflow: hidden; border: 1px solid #5d768b; background: #172b3d; box-shadow: 1px 1px 0 rgba(0, 0, 0, .35); place-items: center; }.pixel-cafe-my-room-member-avatar:first-child { margin-left: 0; }.pixel-cafe-my-room-member-avatar img { display: block; width: 100%; height: 100%; image-rendering: pixelated; object-fit: contain; }.pixel-cafe-my-room-member-more { display: grid; min-width: 1.55rem; height: 1.55rem; margin-left: -.15rem; padding: 0 .18rem; border: 1px solid #806f57; color: #f3c36d; background: #182535; place-items: center; font: 700 .58rem/1 monospace; }.pixel-cafe-my-room-usage { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .45rem; min-width: 0; margin-top: .1rem; }.pixel-cafe-my-room-window { min-width: 0; padding: .5rem .55rem; border: 1px solid rgba(95, 120, 139, .72); background: rgba(7, 25, 40, .72); }.pixel-cafe-my-room-window-heading { display: flex; justify-content: space-between; gap: .55rem; color: #dce9f2; font-size: .68rem; }.pixel-cafe-my-room-window-heading strong { color: #f3c36d; font-size: .7rem; }.pixel-cafe-my-room-window-heading span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.pixel-cafe-my-room-progress { height: .42rem; margin-top: .45rem; overflow: hidden; border: 1px solid rgba(106, 134, 153, .62); background: #07131f; }.pixel-cafe-my-room-progress > span { display: block; height: 100%; background: #64c991; }.pixel-cafe-my-room-progress.tone-warning > span { background: #e2b25f; }.pixel-cafe-my-room-progress.tone-danger > span { background: #e87970; }.pixel-cafe-my-room-progress.tone-unlimited { background: repeating-linear-gradient(135deg, rgba(100, 201, 145, .28) 0, rgba(100, 201, 145, .28) 4px, rgba(7, 19, 31, .9) 4px, rgba(7, 19, 31, .9) 8px); }.pixel-cafe-my-room-progress.tone-unavailable > span { background: #5f7180; }
.pixel-cafe-room-card-titleline { display: flex; min-width: 0; align-items: center; justify-content: space-between; gap: .5rem; }.pixel-cafe-room-card-titleline strong { min-width: 0; }.pixel-cafe-room-card-badges { display: flex; flex: 0 0 auto; gap: .2rem; }.pixel-cafe-room-card-badge { padding: .16rem .3rem; border: 1px solid #53718a; color: #a9c7dc; background: #172c3f; font: 700 .58rem/1 monospace; letter-spacing: .03em; }.pixel-cafe-room-card-badge.tier { border-color: #806f57; color: #f3c36d; background: #2a261f; }.pixel-cafe-room-card-members { display: flex; min-width: 0; min-height: 1.9rem; align-items: center; padding: .1rem 0 0 .35rem; }.pixel-cafe-room-member-avatar { display: grid; width: 1.75rem; height: 1.75rem; margin-left: -.35rem; overflow: hidden; border: 1px solid #5d768b; background: #172b3d; box-shadow: 1px 1px 0 rgba(0, 0, 0, .35); place-items: center; }.pixel-cafe-room-member-avatar:first-child { margin-left: 0; }.pixel-cafe-room-member-avatar img { display: block; width: 100%; height: 100%; image-rendering: pixelated; object-fit: contain; }.pixel-cafe-room-member-more { display: grid; min-width: 1.75rem; height: 1.75rem; margin-left: -.2rem; padding: 0 .2rem; border: 1px solid #806f57; color: #f3c36d; background: #182535; place-items: center; font: 700 .62rem/1 monospace; }
@media (max-width: 620px) { .pixel-cafe-room-card-members { min-height: 1.05rem; padding: 0 0 0 .2rem; }.pixel-cafe-room-member-avatar { width: 1rem; height: 1rem; margin-left: -.2rem; }.pixel-cafe-room-member-more { min-width: 1rem; height: 1rem; margin-left: -.12rem; padding: 0 .12rem; font-size: .5rem; } }
@media (max-width: 620px) { .pixel-cafe-my-room-usage { grid-template-columns: 1fr; } }
@media (max-width: 620px) { .pixel-cafe-my-room-members { min-height: 1.3rem; }.pixel-cafe-my-room-member-avatar { width: 1.2rem; height: 1.2rem; margin-left: -.22rem; }.pixel-cafe-my-room-member-more { min-width: 1.2rem; height: 1.2rem; margin-left: -.1rem; padding: 0 .1rem; font-size: .5rem; } }
.pixel-cafe-purchase-limits { margin: 0 0 1rem; padding: .75rem; border: 1px solid rgba(190, 213, 230, .22); background: rgba(9, 29, 45, .72); }
.pixel-cafe-purchase-limits-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .75rem; color: #f1c26f; font-size: .76rem; }
.pixel-cafe-purchase-limits-heading span { color: #9fb2c1; font-size: .68rem; text-align: right; }
.pixel-cafe-purchase-limits dl { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .45rem; margin: .65rem 0 0; }
.pixel-cafe-purchase-limits dl div { display: flex; align-items: baseline; justify-content: space-between; gap: .45rem; padding: .45rem .5rem; border: 1px solid rgba(190, 213, 230, .16); }
.pixel-cafe-purchase-limits dt { color: #9fb2c1; font-size: .68rem; }
.pixel-cafe-purchase-limits dd { margin: 0; color: #eaf3fa; font: 700 .7rem/1.2 monospace; }
@media (max-width: 420px) { .pixel-cafe-purchase-limits dl { grid-template-columns: 1fr; } }
</style>
