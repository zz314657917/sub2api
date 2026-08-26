<template>
  <component :is="embedded ? 'div' : AppLayout">
    <div :class="[
      'admin-group-buy min-h-[calc(100vh-4rem)]',
      embedded ? 'py-1' : '-m-4 md:-m-[1.35rem] lg:-m-[1.6rem]',
    ]">
      <div class="mx-auto flex w-full max-w-[1440px] flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
        <header class="admin-group-buy-header admin-group-buy-toolbar">
          <div class="admin-group-buy-actions">
            <button type="button" class="admin-group-buy-secondary" :disabled="loading" @click="refreshAll">
              <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
              刷新
            </button>
            <button type="button" class="admin-group-buy-primary" @click="openCreatePlan">
              <Icon name="plus" size="sm" />
              新建房间计划
            </button>
          </div>
        </header>

        <div v-if="roomManagedGroups.length === 0" class="admin-group-buy-alert">
          <Icon name="exclamationTriangle" size="sm" />
          <span>需要先创建并启用网吧房间托管分组，房间计划才能绑定真实可用额度。</span>
        </div>

        <section class="admin-group-buy-panel">
          <div class="admin-group-buy-panel-head">
            <div>
              <h2>网吧房间计划</h2>
              <p>配置份额、价格、开团模式和权益档位区间规则。</p>
            </div>
          </div>

          <div v-if="plans.length === 0 && !loading" class="admin-group-buy-empty">
            <Icon name="gift" size="xl" />
            <p>暂无房间计划</p>
          </div>
          <div v-else class="admin-group-buy-plan-grid">
            <article v-for="plan in plans" :key="plan.id" class="admin-group-buy-plan">
              <div class="admin-group-buy-plan-top">
                <div class="min-w-0">
                  <p class="admin-group-buy-eyebrow">{{ fulfillmentModeLabel(plan.fulfillment_mode) }} · {{ capacityLabel(plan) }}{{ isRoomPlan(plan) ? '' : ` · ${launchModeLabel(plan.launch_mode)}` }}</p>
                  <h3>{{ plan.title }}</h3>
                  <p>{{ plan.description || 'Token拼拼拼 平台托管容量份额' }}</p>
                </div>
                <span class="admin-group-buy-badge" :class="plan.status === 'active' ? 'ok' : 'idle'">
                  {{ plan.status === 'active' ? '已上架' : '已下架' }}
                </span>
              </div>
              <div class="admin-group-buy-price">{{ priceDisplay(plan) }}<small v-if="!plan.price_label">/份</small></div>
              <dl class="admin-group-buy-meta">
                <div><dt>当前团</dt><dd>{{ currentRoundSummary(plan) }}</dd></div>
                <div><dt>单份额度</dt><dd>{{ plan.quota_per_share_label || plan.quota_label || '未填写' }}</dd></div>
                <div><dt>权益档位</dt><dd>{{ tierSummary(plan) }}</dd></div>
                <div><dt>有效期</dt><dd>{{ plan.validity_days }} 天</dd></div>
                <div><dt>截止</dt><dd>{{ timeoutLabel(plan.timeout_minutes) }}</dd></div>
                <div><dt>退款</dt><dd>{{ refundModeLabel(plan.refund_mode) }}</dd></div>
              </dl>
              <div class="admin-group-buy-row-actions">
                <button type="button" class="admin-group-buy-ghost" @click="openEditPlan(plan)">
                  <Icon name="edit" size="sm" />
                  编辑
                </button>
                <button type="button" class="admin-group-buy-ghost" :disabled="isRoomPlan(plan) || plan.launch_mode !== 'manual' || actionPlanId === plan.id" @click="createRound(plan)">
                  <Icon name="plus" size="sm" />
                  手动开团
                </button>
                <button type="button" class="admin-group-buy-danger" :disabled="deletingPlanId === plan.id" @click="deletePlan(plan)">
                  <Icon name="trash" size="sm" />
                  {{ deletingPlanId === plan.id ? '删除中' : '删除' }}
                </button>
              </div>
            </article>
          </div>
        </section>

        <section class="admin-group-buy-panel">
          <div class="admin-group-buy-panel-head admin-group-buy-panel-head-wrap">
            <div>
              <h2>团次处理</h2>
              <p>手动关闭、重试激活和处理退款都通过后端事务收口。</p>
            </div>
            <div class="admin-group-buy-filter">
              <select v-model="roundStatusFilter" @change="loadRounds">
                <option value="">全部状态</option>
                <option value="open">拼团中</option>
                <option value="activating">成团中</option>
                <option value="active">已成团</option>
                <option value="failed">未满份</option>
                <option value="cancelled">已关闭</option>
              </select>
              <button type="button" class="admin-group-buy-secondary" :disabled="roundsLoading" @click="loadRounds">
                <Icon name="refresh" size="sm" :class="{ 'animate-spin': roundsLoading }" />
              </button>
            </div>
          </div>

          <div v-if="rounds.length === 0 && !roundsLoading" class="admin-group-buy-empty admin-group-buy-empty-soft">
            <Icon name="clipboard" size="xl" />
            <p>暂无团次</p>
          </div>
          <div v-else class="admin-group-buy-table">
            <div class="admin-group-buy-table-row admin-group-buy-table-head">
              <span>团次</span>
              <span>计划</span>
              <span>状态</span>
              <span>份额</span>
              <span>退款</span>
              <span>截止时间</span>
              <span>操作</span>
            </div>
            <div v-for="round in rounds" :key="round.id" class="admin-group-buy-table-row">
              <span data-label="团次">#{{ round.id }}</span>
              <span data-label="计划">{{ planTitle(round.plan_id) }}</span>
              <span data-label="状态"><b class="admin-group-buy-badge" :class="roundStatusClass(round.status)">{{ roundStatusLabel(round.status) }}</b></span>
              <span data-label="份额">{{ round.paid_shares ?? round.paid_seats }} 已付 / {{ round.reserved_shares ?? round.reserved_seats }} 预留 / {{ round.total_shares ?? round.total_seats }} 总份</span>
              <span data-label="退款">{{ refundSummaryLabel(round) }}</span>
              <span data-label="截止时间">{{ formatDateTime(round.deadline_at) }}</span>
              <span class="admin-group-buy-table-actions">
                <button type="button" :disabled="detailLoading && detailRound?.id === round.id" @click="openRoundDetail(round)">
                  <Icon name="eye" size="sm" />
                  详情
                </button>
                <button type="button" :disabled="round.status !== 'open' || actionRoundId === round.id" @click="closeRound(round)">关闭</button>
                <button type="button" :disabled="!canRetryActivation(round) || actionRoundId === round.id" @click="retryActivation(round)">重试成团</button>
                <button type="button" :disabled="!canProcessRefunds(round) || actionRoundId === round.id" @click="processRefunds(round)">处理退款</button>
              </span>
            </div>
          </div>

          <div v-if="roundPagination.total > roundPagination.page_size" class="admin-group-buy-pagination">
            <button type="button" :disabled="roundPagination.page <= 1 || roundsLoading" @click="changeRoundPage(roundPagination.page - 1)">上一页</button>
            <span>{{ roundPagination.page }} / {{ roundPagination.pages || 1 }}</span>
            <button type="button" :disabled="roundPagination.page >= roundPagination.pages || roundsLoading" @click="changeRoundPage(roundPagination.page + 1)">下一页</button>
          </div>
        </section>
      </div>
    </div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="planDialogOpen" class="admin-group-buy-modal-backdrop">
          <form class="admin-group-buy-modal" @submit.prevent="savePlan">
            <button type="button" class="admin-group-buy-modal-close" @click="closePlanDialog">
              <Icon name="x" size="sm" />
            </button>
            <p class="admin-group-buy-eyebrow">{{ editingPlan ? '编辑房间计划' : '新建房间计划' }}</p>
            <h2>{{ editingPlan ? editingPlan.title : 'Token拼拼拼' }}</h2>

            <div class="admin-group-buy-form-grid">
              <label>
                <span>标题</span>
                <input v-model.trim="planForm.title" required />
              </label>
              <label>
                <span>计划类型</span>
                <div class="admin-group-buy-readonly-field">网吧房间</div>
              </label>
              <label>
                <span>房间人数上限（1=独享）</span>
                <input v-model.number="planForm.total_shares" type="number" min="1" max="10" required />
              </label>
              <label>
                <span>单份价格</span>
                <input v-model.number="planForm.price_per_share" type="number" min="0.01" step="0.01" required />
              </label>
              <label>
                <span>价格展示文案</span>
                <input v-model.trim="planForm.price_label" maxlength="120" placeholder="例如：每份 128 元 / 首月特价" />
              </label>
              <label>
                <span>每用户最大份额</span>
                <input v-model.number="planForm.max_shares_per_user" type="number" min="1" max="10" required />
              </label>
              <label>
                <span>有效期天数</span>
                <input v-model.number="planForm.validity_days" type="number" min="1" required />
              </label>
              <label>
                <span>成团超时分钟</span>
                <input v-model.number="planForm.timeout_minutes" type="number" min="5" required />
              </label>
              <label>
                <span>退款策略</span>
                <select v-model="planForm.refund_mode">
                  <option value="balance_credit">退回余额</option>
                  <option value="provider_refund">原路退款</option>
                </select>
              </label>
              <label>
                <span>上下架</span>
                <select v-model="planForm.status">
                  <option value="active">上架</option>
                  <option value="disabled">下架</option>
                </select>
              </label>
              <label>
                <span>排序</span>
                <input v-model.number="planForm.sort_order" type="number" min="0" />
              </label>
            </div>

            <label class="admin-group-buy-full">
              <span>单份月额度展示文案</span>
              <input v-model.trim="planForm.quota_per_share_label" placeholder="例如：单份约 50 USD 月额度" />
            </label>

            <section class="admin-group-buy-tiers">
              <div class="admin-group-buy-tier-head">
                <div>
                  <h3>房间托管配置</h3>
                  <p>人数上限为 1 时是独享房间，设置为 2-10 时允许多人共享同一个 Pro 账号。</p>
                </div>
              </div>
              <div class="admin-group-buy-tier-grid">
                <label>
                  <span>托管订阅分组</span>
                  <select v-model.number="planForm.target_group_id" required>
                    <option :value="0" disabled>选择网吧房间托管分组</option>
                    <option v-for="group in roomManagedGroups" :key="group.id" :value="group.id">
                      {{ group.name }} · {{ group.platform }}{{ group.monthly_limit_usd != null ? ` · 月 ${group.monthly_limit_usd}` : '' }}
                    </option>
                  </select>
                </label>
                <label>
                  <span>每座 Key 总额度（USD）</span>
                  <input v-model.number="planForm.room_key_quota_usd" type="number" min="0" step="0.01" />
                </label>
                <label>
                  <span>每座 Key 5 小时限额</span>
                  <input v-model.number="planForm.room_key_rate_limit_5h" type="number" min="0" step="0.01" />
                </label>
                <label>
                  <span>每座 Key 1 日限额</span>
                  <input v-model.number="planForm.room_key_rate_limit_1d" type="number" min="0" step="0.01" />
                </label>
                <label>
                  <span>每座 Key 7 日限额</span>
                  <input v-model.number="planForm.room_key_rate_limit_7d" type="number" min="0" step="0.01" />
                </label>
              </div>
              <p class="admin-group-buy-tier-note">受管 Key 会在座位激活时自动创建，不能关闭。</p>
            </section>

            <label class="admin-group-buy-full">
              <span>说明</span>
              <textarea v-model.trim="planForm.description" rows="2" placeholder="Token拼拼拼 平台托管容量份额，满份后开通。" />
            </label>
            <label class="admin-group-buy-full">
              <span>协议文案</span>
              <textarea v-model.trim="planForm.agreement_text" rows="3" />
            </label>

            <div class="admin-group-buy-modal-actions">
              <button type="button" class="admin-group-buy-secondary" @click="closePlanDialog">取消</button>
              <button type="submit" class="admin-group-buy-primary" :disabled="submitting || !canSubmitPlan">
                {{ submitting ? '保存中' : '保存房间计划' }}
              </button>
            </div>
          </form>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="detailRound" class="admin-group-buy-modal-backdrop" @click.self="closeRoundDetail">
          <section class="admin-group-buy-modal admin-group-buy-detail" role="dialog" aria-modal="true" :aria-label="`团次 #${detailRound.id} 详情`">
            <button type="button" class="admin-group-buy-modal-close" aria-label="关闭" @click="closeRoundDetail">
              <Icon name="x" size="sm" />
            </button>
            <p class="admin-group-buy-eyebrow">团次 #{{ detailRound.id }}</p>
            <h2>{{ planTitle(detailRound.plan_id) }}</h2>

            <div class="admin-group-buy-detail-summary">
              <div><span>状态</span><b>{{ roundStatusLabel(detailRound.status) }}</b></div>
              <div><span>已付份额</span><b>{{ detailRound.paid_shares ?? detailRound.paid_seats }}</b></div>
              <div><span>退款成功</span><b>{{ detailRound.refund_summary?.succeeded || 0 }}</b></div>
              <div><span>待确认</span><b>{{ pendingRefundCount(detailRound) }}</b></div>
              <div><span>失败</span><b>{{ detailRound.refund_summary?.failed || 0 }}</b></div>
              <div><span>人工复核</span><b>{{ detailRound.refund_summary?.needs_review || 0 }}</b></div>
            </div>

            <div v-if="detailLoading" class="admin-group-buy-empty admin-group-buy-empty-soft">
              <Icon name="refresh" size="lg" class="animate-spin" />
            </div>
            <div v-else-if="roundSeats.length === 0" class="admin-group-buy-empty admin-group-buy-empty-soft">
              <p>暂无参与记录</p>
            </div>
            <div v-else class="admin-group-buy-detail-list">
              <article v-for="seat in roundSeats" :key="seat.id" class="admin-group-buy-detail-row">
                <div>
                  <b>{{ seat.user?.username || seat.user?.email || `用户 #${seat.user_id}` }}</b>
                  <span>{{ seat.user?.email || `用户 ID ${seat.user_id}` }}</span>
                </div>
                <dl>
                  <div><dt>份额</dt><dd>{{ seat.share_count }}</dd></div>
                  <div><dt>份额状态</dt><dd>{{ seatStatusLabel(seat.status) }}</dd></div>
                  <div><dt>订单</dt><dd>{{ seat.order ? `#${seat.order.id} · ${seat.order.status}` : '-' }}</dd></div>
                  <div><dt>实付</dt><dd>{{ seat.order ? formatMoney(seat.order.pay_amount) : '-' }}</dd></div>
                  <div><dt>退款方式</dt><dd>{{ seat.refund ? refundModeLabel(seat.refund.mode) : '-' }}</dd></div>
                  <div><dt>退款状态</dt><dd>{{ seat.refund ? refundStatusLabel(seat.refund.status) : '-' }}</dd></div>
                </dl>
                <p v-if="seat.refund?.note">{{ seat.refund.note }}</p>
              </article>
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>
  </component>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { resolveGroupBuyProductName } from '@/utils/groupBuyProduct'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { AdminGroup, BasePaginationResponse } from '@/types'
import type { GroupBuyAdminSeat, GroupBuyFulfillmentMode, GroupBuyLaunchMode, GroupBuyPlan, GroupBuyPlanStatus, GroupBuyRefundMode, GroupBuyRound, GroupBuyRoundStatus, GroupBuySeatStatus, GroupBuyTier } from '@/types/groupBuy'
import type { GroupBuyPlanPayload } from '@/api/admin/groupBuy'

withDefaults(defineProps<{
  embedded?: boolean
}>(), {
  embedded: false,
})

const appStore = useAppStore()

const groupBuyProductName = computed(() => resolveGroupBuyProductName(appStore.cachedPublicSettings))
const defaultAgreementText = computed(() => `我理解 ${groupBuyProductName.value} 为平台托管容量份额权益，不是官方 OpenAI Pro 子账号，不共享任何上游账号或官方 API Key。`)

const plans = ref<GroupBuyPlan[]>([])
const allPlans = ref<GroupBuyPlan[]>([])
const groups = ref<AdminGroup[]>([])
const rounds = ref<GroupBuyRound[]>([])
const loading = ref(false)
const roundsLoading = ref(false)
const submitting = ref(false)
const planDialogOpen = ref(false)
const editingPlan = ref<GroupBuyPlan | null>(null)
const deletingPlanId = ref<number | null>(null)
const actionRoundId = ref<number | null>(null)
const actionPlanId = ref<number | null>(null)
const roundStatusFilter = ref('')
const detailRound = ref<GroupBuyRound | null>(null)
const roundSeats = ref<GroupBuyAdminSeat[]>([])
const detailLoading = ref(false)

const roundPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0,
})

const planForm = reactive({
  title: '',
  description: '',
  total_shares: 10,
  price_per_share: 0,
  price_label: '',
  quota_per_share_label: '',
  max_shares_per_user: 10,
  target_group_id: 0,
  fulfillment_mode: 'room_subscription' as GroupBuyFulfillmentMode,
  room_key_quota_usd: 0,
  room_key_rate_limit_5h: 0,
  room_key_rate_limit_1d: 0,
  room_key_rate_limit_7d: 0,
  auto_create_room_key: true,
  tier_group_ids: {} as Record<string, number>,
  tier_rules: [] as GroupBuyTier[],
  launch_mode: 'manual' as GroupBuyLaunchMode,
  validity_days: 30,
  timeout_minutes: 1440,
  refund_mode: 'balance_credit' as GroupBuyRefundMode,
  agreement_text: defaultAgreementText.value,
  status: 'active' as GroupBuyPlanStatus,
  sort_order: 0,
})

const roomManagedGroups = computed(() =>
  groups.value.filter((group) => group.status === 'active' && group.subscription_type === 'subscription' && group.access_mode === 'room_managed'),
)

const canSubmitPlan = computed(() =>
  !!planForm.title.trim()
  && planForm.total_shares > 0
  && planForm.total_shares <= 10
  && planForm.price_per_share > 0
  && planForm.max_shares_per_user > 0
  && planForm.max_shares_per_user <= 10
  && planForm.validity_days > 0
  && planForm.timeout_minutes >= 5
  && planForm.target_group_id > 0
  && planForm.room_key_quota_usd >= 0
  && planForm.room_key_rate_limit_5h >= 0
  && planForm.room_key_rate_limit_1d >= 0
  && planForm.room_key_rate_limit_7d >= 0,
)

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([loadGroups(), loadPlans(), loadRounds()])
  } finally {
    loading.value = false
  }
}

async function loadPlans() {
  try {
    const res = await adminAPI.groupBuy.listPlans()
    allPlans.value = res.data || []
    plans.value = allPlans.value.filter(isRoomPlan)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '房间计划加载失败'))
  }
}

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '分组加载失败'))
  }
}

async function loadRounds() {
  roundsLoading.value = true
  try {
    const params = {
      page: roundPagination.page,
      page_size: roundPagination.page_size,
      status: roundStatusFilter.value || undefined,
    }
    const res = await adminAPI.groupBuy.listRounds(params)
    const data = res.data as BasePaginationResponse<GroupBuyRound>
    rounds.value = data?.items || []
    roundPagination.total = data?.total || 0
    roundPagination.pages = data?.pages || 0
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '团次加载失败'))
  } finally {
    roundsLoading.value = false
  }
}

function openCreatePlan() {
  editingPlan.value = null
  resetPlanForm()
  planDialogOpen.value = true
}

function openEditPlan(plan: GroupBuyPlan) {
  editingPlan.value = plan
  const tierRules = normalizeTierRules(plan)
  const targetGroupID = targetGroupIDForShares(tierRules, totalShares(plan)) || plan.target_group_id
  Object.assign(planForm, {
    title: plan.title,
    description: plan.description || '',
    total_shares: totalShares(plan),
    price_per_share: pricePerShare(plan),
    price_label: plan.price_label || '',
    quota_per_share_label: plan.quota_per_share_label || plan.quota_label || '',
    max_shares_per_user: plan.max_shares_per_user || 10,
    target_group_id: targetGroupID,
    fulfillment_mode: 'room_subscription',
    room_key_quota_usd: plan.room_key_quota_usd || 0,
    room_key_rate_limit_5h: plan.room_key_rate_limit_5h || 0,
    room_key_rate_limit_1d: plan.room_key_rate_limit_1d || 0,
    room_key_rate_limit_7d: plan.room_key_rate_limit_7d || 0,
    auto_create_room_key: plan.auto_create_room_key === true,
    tier_group_ids: tierRulesToMap(tierRules, totalShares(plan)),
    tier_rules: tierRules,
    launch_mode: plan.launch_mode || 'auto',
    validity_days: plan.validity_days,
    timeout_minutes: plan.timeout_minutes,
    refund_mode: plan.refund_mode,
    agreement_text: plan.agreement_text || defaultAgreementText.value,
    status: plan.status,
    sort_order: plan.sort_order,
  })
  planDialogOpen.value = true
}

function resetPlanForm() {
  const firstGroupID = roomManagedGroups.value[0]?.id || 0
  const rules = firstGroupID ? [buildTierRule(1, 10, firstGroupID, '房间座位')] : []
  Object.assign(planForm, {
    title: '',
    description: `${groupBuyProductName.value} 平台托管容量份额，满份后自动开通。`,
    total_shares: 10,
    price_per_share: 0,
    price_label: '',
    quota_per_share_label: '单份月额度待填写',
    max_shares_per_user: 10,
    target_group_id: firstGroupID,
    fulfillment_mode: 'room_subscription',
    room_key_quota_usd: 0,
    room_key_rate_limit_5h: 0,
    room_key_rate_limit_1d: 0,
    room_key_rate_limit_7d: 0,
    auto_create_room_key: true,
    tier_group_ids: tierRulesToMap(rules, 10),
    tier_rules: rules,
    launch_mode: 'manual',
    validity_days: 30,
    timeout_minutes: 1440,
    refund_mode: 'balance_credit',
    agreement_text: defaultAgreementText.value,
    status: 'active',
    sort_order: 0,
  })
}

function closePlanDialog() {
  if (submitting.value) return
  planDialogOpen.value = false
}

function buildPlanPayload(): GroupBuyPlanPayload {
  const totalShares = Number(planForm.total_shares)
  const tierRules = [buildTierRule(1, totalShares, Number(planForm.target_group_id), '房间座位')]
  const tierGroupIds = tierRulesToMap(tierRules, totalShares)
  return {
    title: planForm.title.trim(),
    description: planForm.description.trim(),
    product_key: 'token_pinpinpin',
    total_shares: totalShares,
    seat_count: totalShares,
    price_per_share: Number(planForm.price_per_share),
    price_per_seat: Number(planForm.price_per_share),
    price_label: planForm.price_label.trim(),
    quota_per_share_label: planForm.quota_per_share_label.trim(),
    quota_label: planForm.quota_per_share_label.trim(),
    max_shares_per_user: Number(planForm.max_shares_per_user),
    target_group_id: Number(planForm.target_group_id),
    fulfillment_mode: 'room_subscription',
    room_key_quota_usd: Number(planForm.room_key_quota_usd) || 0,
    room_key_rate_limit_5h: Number(planForm.room_key_rate_limit_5h) || 0,
    room_key_rate_limit_1d: Number(planForm.room_key_rate_limit_1d) || 0,
    room_key_rate_limit_7d: Number(planForm.room_key_rate_limit_7d) || 0,
    auto_create_room_key: true,
    tier_group_ids: tierGroupIds,
    tier_rules: tierRules,
    launch_mode: 'manual',
    validity_days: Number(planForm.validity_days),
    timeout_minutes: Number(planForm.timeout_minutes),
    refund_mode: planForm.refund_mode,
    agreement_text: planForm.agreement_text.trim(),
    status: planForm.status,
    sort_order: Number(planForm.sort_order) || 0,
  }
}

async function savePlan() {
  if (!canSubmitPlan.value || submitting.value) return
  submitting.value = true
  try {
    const payload = buildPlanPayload()
    if (editingPlan.value) {
      await adminAPI.groupBuy.updatePlan(editingPlan.value.id, payload)
      appStore.showSuccess('房间计划已更新')
    } else {
      await adminAPI.groupBuy.createPlan(payload)
      appStore.showSuccess('房间计划已创建')
    }
    planDialogOpen.value = false
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '保存房间计划失败'))
  } finally {
    submitting.value = false
  }
}

async function deletePlan(plan: GroupBuyPlan) {
  if (deletingPlanId.value) return
  if (!window.confirm(`确认删除房间计划「${plan.title}」？`)) return
  deletingPlanId.value = plan.id
  try {
    await adminAPI.groupBuy.deletePlan(plan.id)
    appStore.showSuccess('房间计划已删除')
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '删除拼团失败'))
  } finally {
    deletingPlanId.value = null
  }
}

async function createRound(plan: GroupBuyPlan) {
  actionPlanId.value = plan.id
  try {
    await adminAPI.groupBuy.createRound(plan.id)
    appStore.showSuccess('已手动开团')
    await loadRounds()
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '手动开团失败'))
  } finally {
    actionPlanId.value = null
  }
}

function changeRoundPage(page: number) {
  roundPagination.page = Math.max(1, page)
  void loadRounds()
}

async function closeRound(round: GroupBuyRound) {
  const reason = window.prompt(`关闭团次 #${round.id} 的原因`, '后台手动关闭')
  if (reason == null) return
  actionRoundId.value = round.id
  try {
    await adminAPI.groupBuy.closeRound(round.id, reason.trim() || '后台手动关闭')
    appStore.showSuccess('团次已关闭')
    await loadRounds()
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '关闭团次失败'))
  } finally {
    actionRoundId.value = null
  }
}

async function retryActivation(round: GroupBuyRound) {
  actionRoundId.value = round.id
  try {
    await adminAPI.groupBuy.retryActivation(round.id)
    appStore.showSuccess('已提交重试成团')
    await loadRounds()
    await loadPlans()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '重试成团失败'))
  } finally {
    actionRoundId.value = null
  }
}

async function processRefunds(round: GroupBuyRound) {
  actionRoundId.value = round.id
  try {
    const res = await adminAPI.groupBuy.processRefunds(round.id)
    const result = res.data
    appStore.showSuccess(`退款处理完成：成功 ${result?.succeeded ?? 0}，待确认 ${result?.pending ?? 0}，失败 ${result?.failed ?? 0}`)
    if ((result?.failed ?? 0) > 0) {
      appStore.showError(result.failures?.[0]?.message || '部分退款需要重试或人工复核')
    }
    await loadRounds()
    if (detailRound.value?.id === round.id) {
      detailRound.value = rounds.value.find((item) => item.id === round.id) || detailRound.value
      await loadRoundSeats(round.id)
    }
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '处理退款失败'))
  } finally {
    actionRoundId.value = null
  }
}

async function openRoundDetail(round: GroupBuyRound) {
  detailRound.value = round
  await loadRoundSeats(round.id)
}

async function loadRoundSeats(roundId: number) {
  detailLoading.value = true
  try {
    const res = await adminAPI.groupBuy.listRoundSeats(roundId)
    roundSeats.value = res.data || []
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, '团次详情加载失败'))
    roundSeats.value = []
  } finally {
    detailLoading.value = false
  }
}

function closeRoundDetail() {
  detailRound.value = null
  roundSeats.value = []
}

function canRetryActivation(round: GroupBuyRound): boolean {
  const paid = round.paid_shares ?? round.paid_seats
  const total = round.total_shares ?? round.total_seats
  return round.status === 'activating' || (round.status === 'open' && paid >= total)
}

function canProcessRefunds(round: GroupBuyRound): boolean {
  return round.status === 'failed' || round.status === 'cancelled'
}

function pendingRefundCount(round: GroupBuyRound): number {
  const summary = round.refund_summary
  return (summary?.pending || 0) + (summary?.processing || 0) + (summary?.pending_provider || 0)
}

function refundSummaryLabel(round: GroupBuyRound): string {
  const summary = round.refund_summary
  if (!summary?.total) return '-'
  return `${summary.succeeded} 已退 / ${pendingRefundCount(round)} 待处理 / ${summary.failed + summary.needs_review} 异常`
}

function totalShares(plan: GroupBuyPlan): number {
  return Number(plan.total_shares || plan.seat_count || 10)
}

function capacityLabel(plan: GroupBuyPlan): string {
  const total = totalShares(plan)
  if (!isRoomPlan(plan)) return `总 ${total} 份`
  return total === 1 ? '独享房间' : `${total} 人共享`
}

function buildTierRule(minShares: number, maxShares: number, groupID: number, label: string): GroupBuyTier {
  return {
    min_shares: Number(minShares),
    max_shares: Number(maxShares),
    target_group_id: Number(groupID || 0),
    label,
  }
}

function normalizeTierRules(plan: GroupBuyPlan): GroupBuyTier[] {
  const total = totalShares(plan)
  const source = (plan.tier_rules?.length ? plan.tier_rules : plan.tier_groups) || []
  const rules = source
    .map((tier) => {
      const minShares = Number(tier.min_shares || tier.share_count || 0)
      const maxShares = Number(tier.max_shares || tier.share_count || minShares)
      return buildTierRule(minShares, maxShares, Number(tier.target_group_id || 0), tier.label || tierLabel(minShares, maxShares))
    })
    .filter((tier) => tier.min_shares > 0 && tier.max_shares >= tier.min_shares && tier.target_group_id > 0)
    .sort((a, b) => a.min_shares - b.min_shares || a.max_shares - b.max_shares)
  if (rules.length > 0) return rules
  return exactTierMapToRules(plan.tier_group_ids, total, plan.target_group_id)
}

function exactTierMapToRules(raw?: Record<string, number>, totalShares = 10, fallback = 0): GroupBuyTier[] {
  const rules: GroupBuyTier[] = []
  let currentGroup = 0
  let start = 0
  for (let share = 1; share <= totalShares; share++) {
    const groupID = Number(raw?.[String(share)] || fallback || 0)
    if (groupID !== currentGroup) {
      if (currentGroup > 0) rules.push(buildTierRule(start, share - 1, currentGroup, tierLabel(start, share - 1)))
      currentGroup = groupID
      start = share
    }
  }
  if (currentGroup > 0) rules.push(buildTierRule(start, totalShares, currentGroup, tierLabel(start, totalShares)))
  return rules
}

function tierRulesToMap(rules: GroupBuyTier[], totalShares: number): Record<string, number> {
  const out: Record<string, number> = {}
  for (const tier of rules) {
    for (let share = tier.min_shares; share <= Math.min(tier.max_shares, totalShares); share++) {
      out[String(share)] = Number(tier.target_group_id || 0)
    }
  }
  return out
}

function targetGroupIDForShares(rules: GroupBuyTier[], shares: number): number {
  const target = rules.find((tier) => shares >= tier.min_shares && shares <= tier.max_shares)
  return Number(target?.target_group_id || 0)
}

function tierLabel(minShares: number, maxShares: number): string {
  return minShares === maxShares ? `${minShares} 份权益` : `${minShares}-${maxShares} 份权益`
}

function tierSummary(plan: GroupBuyPlan): string {
  const rules = normalizeTierRules(plan)
  if (!rules.length) return '未配置'
  if (rules.length === 1) return rules[0].label || tierLabel(rules[0].min_shares, rules[0].max_shares)
  return `${rules.length} 个档位 · 最高 ${rules[rules.length - 1].label || tierLabel(rules[rules.length - 1].min_shares, rules[rules.length - 1].max_shares)}`
}

function pricePerShare(plan: GroupBuyPlan): number {
  return Number(plan.price_per_share || plan.price_per_seat || 0)
}

function priceDisplay(plan: GroupBuyPlan): string {
  return plan.price_label || formatMoney(pricePerShare(plan))
}

function currentRoundSummary(plan: GroupBuyPlan): string {
  const round = plan.current_round
  if (!round) return plan.launch_mode === 'manual' ? '待手动开团' : '自动创建'
  return `${round.paid_shares ?? round.paid_seats}/${round.total_shares ?? round.total_seats} 份 · ${roundStatusLabel(round.status)}`
}

function planTitle(planId: number): string {
  return allPlans.value.find((plan) => plan.id === planId)?.title || `计划 #${planId}`
}

function formatMoney(value: number): string {
  return formatPaymentAmount(Number(value || 0), 'CNY', 'zh-CN')
}

function timeoutLabel(minutes: number): string {
  if (minutes >= 1440 && minutes % 1440 === 0) return `${minutes / 1440} 天`
  if (minutes >= 60 && minutes % 60 === 0) return `${minutes / 60} 小时`
  return `${minutes} 分钟`
}

function refundModeLabel(mode: string): string {
  return mode === 'provider_refund' ? '原路退款' : '退回余额'
}

function refundStatusLabel(status: string): string {
  switch (status) {
    case 'processing': return '处理中'
    case 'pending_provider': return '渠道待确认'
    case 'succeeded': return '已退款'
    case 'failed': return '失败'
    case 'needs_review': return '人工复核'
    default: return status
  }
}

function seatStatusLabel(status: GroupBuySeatStatus): string {
  switch (status) {
    case 'locked': return '待支付'
    case 'released': return '已释放'
    case 'paid': return '已支付'
    case 'active': return '权益生效'
    case 'refund_pending': return '待退款'
    case 'refund_processing': return '退款处理中'
    case 'refunded': return '已退款'
    case 'cancelled': return '已取消'
    default: return status
  }
}

function launchModeLabel(mode?: string): string {
  return mode === 'manual' ? '手动开团' : '自动续开'
}

function isRoomPlan(plan: Pick<GroupBuyPlan, 'fulfillment_mode'>): boolean {
  return plan.fulfillment_mode === 'room_subscription'
}

function fulfillmentModeLabel(mode?: string): string {
  return mode === 'room_subscription' ? '网吧房间' : '普通拼团'
}

function roundStatusLabel(status: GroupBuyRoundStatus): string {
  switch (status) {
    case 'open': return '拼团中'
    case 'activating': return '成团中'
    case 'active': return '已成团'
    case 'failed': return '未满份'
    case 'cancelled': return '已关闭'
    default: return status
  }
}

function roundStatusClass(status: GroupBuyRoundStatus): string {
  if (status === 'active') return 'ok'
  if (status === 'failed' || status === 'cancelled') return 'bad'
  if (status === 'activating') return 'warn'
  return 'idle'
}

function formatDateTime(value?: string): string {
  if (!value) return '-'
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

onMounted(() => {
  void refreshAll()
})
</script>

<style scoped>
.admin-group-buy,
.admin-group-buy-modal-backdrop {
  --agb-bg: var(--console-bg, #faf9f5);
  --agb-surface: var(--console-surface, #faf9f5);
  --agb-surface-soft: var(--console-surface-soft, rgba(245, 240, 232, 0.76));
  --agb-surface-hover: var(--console-surface-hover, #fffaf5);
  --agb-input: var(--console-input, rgba(250, 249, 245, 0.94));
  --agb-text: var(--console-text, #141413);
  --agb-muted: var(--console-muted, #6c6a64);
  --agb-muted-strong: var(--console-muted-strong, #3d3d3a);
  --agb-border: var(--console-border, rgba(216, 206, 194, 0.68));
  --agb-border-strong: var(--console-border-strong, rgba(160, 153, 144, 0.38));
  --agb-accent: var(--console-accent, #cc785c);
  --agb-accent-strong: #a9583e;
  --agb-ring: var(--console-ring, rgba(204, 120, 92, 0.16));
  --agb-shadow: var(--console-shadow, 0 12px 28px rgba(75, 52, 40, 0.045));
}

.admin-group-buy {
  background: transparent;
  color: var(--agb-text);
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.admin-group-buy-panel-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
}

.admin-group-buy-header {
  border-bottom: 1px solid var(--agb-border);
  display: flex;
  justify-content: flex-end;
  padding-bottom: 0.95rem;
}

.admin-group-buy-toolbar {
  min-height: 2.75rem;
}

.admin-group-buy-panel h2,
.admin-group-buy-modal h2 {
  margin: 0.2rem 0 0;
  font-size: clamp(1.65rem, 2.4vw, 2.35rem);
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.12;
  color: var(--agb-text);
}

.admin-group-buy-panel-head p,
.admin-group-buy-tier-head p {
  margin-top: 0.65rem;
  max-width: 50rem;
  color: var(--agb-muted);
  line-height: 1.65;
}

.admin-group-buy-tier-note {
  margin: 0.95rem 0 0;
  border: 1px solid rgba(204, 120, 92, 0.22);
  border-radius: 8px;
  background: rgba(255, 250, 245, 0.7);
  color: var(--agb-muted-strong);
  font-size: 0.82rem;
  line-height: 1.65;
  padding: 0.75rem 0.85rem;
}

.admin-group-buy-eyebrow {
  color: var(--agb-accent-strong);
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0;
}

.admin-group-buy-actions,
.admin-group-buy-row-actions,
.admin-group-buy-table-actions,
.admin-group-buy-modal-actions,
.admin-group-buy-filter,
.admin-group-buy-tier-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.admin-group-buy-tier-head {
  justify-content: space-between;
}

.admin-group-buy-primary,
.admin-group-buy-secondary,
.admin-group-buy-ghost,
.admin-group-buy-danger,
.admin-group-buy-table-actions button,
.admin-group-buy-pagination button {
  display: inline-flex;
  min-height: 2.45rem;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 999px;
  padding: 0 0.95rem;
  font-size: 0.88rem;
  font-weight: 700;
  transition: background-color 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease, transform 0.12s ease, opacity 0.12s ease;
}

.admin-group-buy-primary {
  background: var(--agb-accent);
  color: #fff;
}

.admin-group-buy-secondary,
.admin-group-buy-ghost,
.admin-group-buy-table-actions button,
.admin-group-buy-pagination button {
  border: 1px solid var(--agb-border-strong);
  background: var(--agb-surface-soft);
  color: var(--agb-muted-strong);
}

.admin-group-buy-danger {
  border: 1px solid rgba(185, 28, 28, 0.22);
  background: #fff1f1;
  color: #991b1b;
}

.admin-group-buy-primary:active,
.admin-group-buy-secondary:active,
.admin-group-buy-ghost:active,
.admin-group-buy-danger:active,
.admin-group-buy-table-actions button:active,
.admin-group-buy-pagination button:active {
  transform: scale(0.97);
}

.admin-group-buy-primary:hover:not(:disabled) {
  background: var(--agb-accent-strong);
  box-shadow: 0 0 0 3px var(--agb-ring);
}

.admin-group-buy-secondary:hover:not(:disabled),
.admin-group-buy-ghost:hover:not(:disabled),
.admin-group-buy-table-actions button:hover:not(:disabled),
.admin-group-buy-pagination button:hover:not(:disabled) {
  border-color: rgba(204, 120, 92, 0.34);
  background: var(--agb-surface-hover);
  color: var(--agb-text);
}

.admin-group-buy-primary:disabled,
.admin-group-buy-secondary:disabled,
.admin-group-buy-ghost:disabled,
.admin-group-buy-danger:disabled,
.admin-group-buy-table-actions button:disabled,
.admin-group-buy-pagination button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.admin-group-buy-panel,
.admin-group-buy-plan,
.admin-group-buy-modal {
  border: 1px solid var(--agb-border);
  border-radius: 0.5rem;
  background: var(--agb-surface);
  box-shadow: var(--agb-shadow);
  color: var(--agb-text);
  backdrop-filter: blur(14px) saturate(1.04);
}

.admin-group-buy-panel {
  padding: 1.2rem;
}

.admin-group-buy-plan-grid {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem;
}

.admin-group-buy-plan {
  display: flex;
  min-height: 24rem;
  flex-direction: column;
  padding: 1.2rem;
}

.admin-group-buy-plan-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-group-buy-plan h3,
.admin-group-buy-tiers h3 {
  margin-top: 0.25rem;
  color: var(--agb-text);
  font-size: 1.15rem;
  font-weight: 800;
  line-height: 1.25;
}

.admin-group-buy-plan p,
.admin-group-buy-meta dt {
  color: var(--agb-muted);
  line-height: 1.55;
}

.admin-group-buy-price {
  margin-top: 1.2rem;
  color: var(--agb-text);
  font-size: 2rem;
  font-weight: 800;
  line-height: 1;
}

.admin-group-buy-price small {
  margin-left: 0.35rem;
  color: var(--agb-muted);
  font-size: 0.85rem;
  font-weight: 600;
}

.admin-group-buy-meta {
  margin-top: 1rem;
  display: grid;
  gap: 0.55rem;
}

.admin-group-buy-meta div {
  display: grid;
  grid-template-columns: 5rem minmax(0, 1fr);
  gap: 0.75rem;
  border-top: 1px solid var(--agb-border);
  padding-top: 0.55rem;
}

.admin-group-buy-meta dd {
  min-width: 0;
  overflow: hidden;
  color: var(--agb-muted-strong);
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-group-buy-row-actions {
  margin-top: auto;
  padding-top: 1rem;
}

.admin-group-buy-badge {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  border-radius: 999px;
  padding: 0.35rem 0.6rem;
  font-size: 0.75rem;
  font-weight: 800;
}

.admin-group-buy-badge.ok { background: #dcfce7; color: #166534; }
.admin-group-buy-badge.warn { background: #fef3c7; color: #92400e; }
.admin-group-buy-badge.bad { background: #fee2e2; color: #991b1b; }
.admin-group-buy-badge.idle { background: #ebe5dc; color: #5d5146; }

.admin-group-buy-alert {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  border: 1px solid rgba(180, 83, 9, 0.25);
  border-radius: 0.5rem;
  background: #fff7ed;
  padding: 0.85rem 1rem;
  color: #92400e;
}

.admin-group-buy-empty {
  margin-top: 1rem;
  display: grid;
  min-height: 12rem;
  place-items: center;
  gap: 0.75rem;
  border: 1px dashed var(--agb-border-strong);
  border-radius: 0.5rem;
  color: var(--agb-muted);
}

.admin-group-buy-empty-soft {
  background: var(--agb-surface-soft);
}

.admin-group-buy-table {
  margin-top: 1rem;
  overflow-x: auto;
}

.admin-group-buy-table-row {
  display: grid;
  min-width: 980px;
  grid-template-columns: 0.7fr 1.2fr 0.85fr 1.35fr 1.25fr 1.2fr 1.8fr;
  gap: 1rem;
  align-items: center;
  border-top: 1px solid var(--agb-border);
  padding: 0.85rem 0;
}

.admin-group-buy-table-head {
  border-top: 0;
  color: var(--agb-muted);
  font-size: 0.8rem;
  font-weight: 800;
}

.admin-group-buy-filter select,
.admin-group-buy-form-grid input,
.admin-group-buy-form-grid select,
.admin-group-buy-readonly-field,
.admin-group-buy-full input,
.admin-group-buy-full textarea,
.admin-group-buy-tier-grid input,
.admin-group-buy-tier-grid select {
  width: 100%;
  border: 1px solid var(--agb-border-strong);
  border-radius: 0.5rem;
  background: var(--agb-input);
  color: var(--agb-text);
  outline: none;
}

.admin-group-buy-filter select:focus,
.admin-group-buy-form-grid input:focus,
.admin-group-buy-form-grid select:focus,
.admin-group-buy-full input:focus,
.admin-group-buy-full textarea:focus,
.admin-group-buy-tier-grid input:focus,
.admin-group-buy-tier-grid select:focus {
  border-color: rgba(204, 120, 92, 0.58);
  box-shadow: 0 0 0 3px var(--agb-ring);
}

.admin-group-buy-filter select {
  min-height: 2.45rem;
  padding: 0 0.8rem;
}

.admin-group-buy-pagination {
  margin-top: 1rem;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 0.75rem;
  color: var(--agb-muted);
}

.admin-group-buy-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(20, 20, 19, 0.28);
  padding: 1rem;
}

.admin-group-buy-modal {
  position: relative;
  width: min(100%, 860px);
  max-height: 90vh;
  overflow-y: auto;
  padding: 1.5rem;
  background: var(--agb-surface);
}

.admin-group-buy-detail {
  width: min(100%, 980px);
}

.admin-group-buy-detail-summary {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 0.65rem;
  margin: 1.25rem 0;
}

.admin-group-buy-detail-summary > div {
  border: 1px solid var(--agb-border);
  border-radius: 0.45rem;
  background: var(--agb-surface-soft);
  padding: 0.65rem;
}

.admin-group-buy-detail-summary span,
.admin-group-buy-detail-summary b {
  display: block;
}

.admin-group-buy-detail-summary span {
  color: var(--agb-muted);
  font-size: 0.76rem;
}

.admin-group-buy-detail-summary b {
  margin-top: 0.25rem;
  color: var(--agb-text);
}

.admin-group-buy-detail-list {
  display: grid;
  gap: 0.7rem;
  max-height: 48vh;
  overflow-y: auto;
  padding-right: 0.2rem;
}

.admin-group-buy-detail-row {
  border: 1px solid var(--agb-border);
  border-radius: 0.45rem;
  background: var(--agb-surface-soft);
  padding: 0.8rem;
}

.admin-group-buy-detail-row > div {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-group-buy-detail-row > div span,
.admin-group-buy-detail-row > p {
  color: var(--agb-muted);
  font-size: 0.8rem;
}

.admin-group-buy-detail-row dl {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.55rem;
  margin: 0.7rem 0 0;
}

.admin-group-buy-detail-row dt {
  color: var(--agb-muted);
  font-size: 0.75rem;
}

.admin-group-buy-detail-row dd {
  margin: 0.2rem 0 0;
  color: var(--agb-muted-strong);
  font-size: 0.84rem;
  overflow-wrap: anywhere;
}

.admin-group-buy-detail-row p {
  margin: 0.65rem 0 0;
}

.admin-group-buy-modal-close {
  position: absolute;
  right: 1rem;
  top: 1rem;
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--agb-surface-soft);
  color: var(--agb-muted-strong);
}

.admin-group-buy-form-grid {
  margin-top: 1.25rem;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
}

.admin-group-buy-form-grid label,
.admin-group-buy-full,
.admin-group-buy-tier-grid label {
  display: grid;
  gap: 0.35rem;
  color: var(--agb-muted-strong);
  font-size: 0.84rem;
  font-weight: 800;
}

.admin-group-buy-form-grid input,
.admin-group-buy-form-grid select,
.admin-group-buy-full input,
.admin-group-buy-full textarea,
.admin-group-buy-tier-grid input,
.admin-group-buy-tier-grid select {
  min-height: 2.55rem;
  padding: 0.65rem 0.8rem;
  font-weight: 500;
}

.admin-group-buy-readonly-field {
  display: flex;
  min-height: 2.55rem;
  align-items: center;
  border: 1px solid var(--agb-border);
  border-radius: 0.5rem;
  background: var(--agb-surface-soft);
  padding: 0.65rem 0.8rem;
  color: var(--agb-text);
  font-weight: 700;
}

.admin-group-buy-full,
.admin-group-buy-tiers {
  margin-top: 0.9rem;
}

.admin-group-buy-tiers {
  border: 1px solid var(--agb-border);
  border-radius: 0.5rem;
  background: var(--agb-surface-soft);
  padding: 1rem;
}

.admin-group-buy-tier-grid {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.admin-group-buy-tier-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  justify-content: flex-end;
}

.admin-group-buy-tier-rule {
  display: grid;
  grid-template-columns: 0.65fr 0.65fr minmax(0, 1fr) minmax(0, 1.4fr) auto;
  gap: 0.75rem;
  align-items: end;
  border-top: 1px solid var(--agb-border);
  padding-top: 0.75rem;
}

.admin-group-buy-tier-rule:first-child {
  border-top: 0;
  padding-top: 0;
}

.admin-group-buy-tier-remove {
  min-height: 2.55rem;
  padding-inline: 0.75rem;
}

.admin-group-buy-modal-actions {
  margin-top: 1.2rem;
  justify-content: flex-end;
}

@media (max-width: 1180px) {
  .admin-group-buy-plan-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .admin-group-buy-panel-head-wrap,
  .admin-group-buy-tier-head {
    align-items: flex-start;
    flex-direction: column;
  }
  .admin-group-buy-header {
    align-items: flex-start;
    justify-content: flex-start;
  }
  .admin-group-buy-plan-grid,
  .admin-group-buy-form-grid,
  .admin-group-buy-tier-grid,
  .admin-group-buy-tier-rule {
    grid-template-columns: 1fr;
  }
  .admin-group-buy-table {
    overflow-x: visible;
  }
  .admin-group-buy-table-head {
    display: none;
  }
  .admin-group-buy-table-row {
    min-width: 0;
    grid-template-columns: 1fr;
    gap: 0.55rem;
    border: 1px solid var(--agb-border);
    border-radius: 0.5rem;
    background: var(--agb-surface-soft);
    padding: 0.85rem;
  }

  .admin-group-buy-detail-summary,
  .admin-group-buy-detail-row dl {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .admin-group-buy-detail-row > div {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.15rem;
  }
  .admin-group-buy-table-row + .admin-group-buy-table-row {
    margin-top: 0.75rem;
  }
  .admin-group-buy-table-row > span:not(.admin-group-buy-table-actions) {
    display: grid;
    grid-template-columns: 5.2rem minmax(0, 1fr);
    gap: 0.75rem;
    align-items: center;
    min-width: 0;
    color: var(--agb-muted-strong);
  }
  .admin-group-buy-table-row > span:not(.admin-group-buy-table-actions)::before {
    content: attr(data-label);
    color: var(--agb-muted);
    font-size: 0.78rem;
    font-weight: 800;
  }
  .admin-group-buy-table-actions {
    justify-content: flex-start;
    padding-top: 0.25rem;
  }
}

:global(.dark) .admin-group-buy {
  --agb-accent-strong: #c4b5fd;
  background: transparent;
  color: var(--agb-text);
}

:global(.dark) .admin-group-buy-modal-backdrop {
  --agb-accent-strong: #c4b5fd;
  color: var(--agb-text);
}

:global(.dark) .admin-group-buy-panel,
:global(.dark) .admin-group-buy-plan,
:global(.dark) .admin-group-buy-modal,
:global(.dark) .admin-group-buy-secondary,
:global(.dark) .admin-group-buy-ghost,
:global(.dark) .admin-group-buy-table-actions button,
:global(.dark) .admin-group-buy-pagination button {
  border-color: var(--agb-border);
  background: var(--agb-surface);
  color: var(--agb-text);
}

:global(.dark) .admin-group-buy-panel-head p,
:global(.dark) .admin-group-buy-tier-head p,
:global(.dark) .admin-group-buy-plan p,
:global(.dark) .admin-group-buy-meta dt {
  color: var(--agb-muted);
}

:global(.dark) .admin-group-buy-tier-note {
  border-color: var(--agb-border-strong);
  background: rgba(49, 44, 39, 0.68);
  color: var(--agb-muted-strong);
}

:global(.dark) .admin-group-buy-form-grid input,
:global(.dark) .admin-group-buy-form-grid select,
:global(.dark) .admin-group-buy-full input,
:global(.dark) .admin-group-buy-full textarea,
:global(.dark) .admin-group-buy-filter select,
:global(.dark) .admin-group-buy-tier-grid input,
:global(.dark) .admin-group-buy-tier-grid select,
:global(.dark) .admin-group-buy-modal-close,
:global(.dark) .admin-group-buy-tiers {
  border-color: var(--agb-border-strong);
  background: var(--agb-input);
  color: var(--agb-text);
}
</style>
