<template>
  <aside
    class="sidebar console-sidebar"
    :class="[
      sidebarCollapsed ? 'w-[72px]' : 'w-64',
      { '-translate-x-full lg:translate-x-0': !mobileOpen }
    ]"
  >
    <!-- Logo/Brand -->
    <div class="sidebar-header console-sidebar-header" :class="{ 'sidebar-header-collapsed': sidebarCollapsed }">
      <!-- Custom Logo or Default Logo -->
      <div class="sidebar-logo console-logo-frame flex h-9 w-9 items-center justify-center overflow-hidden">
        <img v-if="settingsLoaded" :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
      </div>
      <div class="sidebar-brand" :class="{ 'sidebar-brand-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
        <span class="sidebar-brand-title console-brand-title text-lg font-bold">
          {{ siteName }}
        </span>
        <!-- Version Badge -->
        <VersionBadge :version="siteVersion" />
      </div>
    </div>

    <!-- Navigation -->
    <nav class="sidebar-nav scrollbar-hide">
      <!-- Admin View: Admin menu first, then personal menu -->
      <template v-if="isAdmin">
        <!-- Admin Section -->
        <div class="sidebar-section">
          <template v-for="item in adminNavItems" :key="item.path">
            <!-- Collapsible group (has children) -->
            <template v-if="item.children?.length">
              <button
                type="button"
                class="sidebar-link mb-1 w-full"
                :class="{
                  'sidebar-link-active': isGroupActive(item) && !isGroupExpanded(item),
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :title="sidebarCollapsed ? item.label : undefined"
                @click="handleGroupClick(item)"
              >
                <span class="console-nav-icon-frame">
                  <Icon :name="item.icon ?? 'document'" size="sm" />
                </span>
                <span
                  class="sidebar-label sidebar-label-flex"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                  >
                    <span class="min-w-0 truncate">{{ item.label }}</span>
                    <span
                      class="flex-shrink-0 transition-transform duration-200"
                      :class="isGroupExpanded(item) ? 'rotate-180' : ''"
                    >
                    <Icon name="chevronDown" size="xs" />
                  </span>
                </span>
              </button>
              <!-- Children -->
              <div v-if="!sidebarCollapsed && isGroupExpanded(item)" class="mb-1 ml-4 border-l-2 border-accent-300 pl-2 dark:border-accent-700">
                <template v-for="child in item.children" :key="child.path">
                  <template v-if="child.children?.length">
                    <button
                      type="button"
                      class="sidebar-link mb-0.5 w-full py-1.5 text-sm"
                      :class="{ 'sidebar-link-active': isGroupActive(child) && !isGroupExpanded(child) }"
                      @click="handleGroupClick(child)"
                    >
                      <span class="console-nav-icon-frame">
                        <Icon :name="child.icon ?? 'document'" size="sm" />
                      </span>
                      <span class="sidebar-label sidebar-label-flex">
                        <span class="min-w-0 truncate">{{ child.label }}</span>
                        <span
                          class="flex-shrink-0 transition-transform duration-200"
                          :class="isGroupExpanded(child) ? 'rotate-180' : ''"
                        >
                          <Icon name="chevronDown" size="xs" />
                        </span>
                      </span>
                    </button>
                    <div v-if="isGroupExpanded(child)" class="mb-0.5 ml-4 border-l-2 border-accent-200 pl-2 dark:border-accent-800">
                      <router-link
                        v-for="grandchild in child.children"
                        :key="grandchild.path"
                        :to="grandchild.path"
                        v-bind="navLinkAttrs(grandchild)"
                        class="sidebar-link mb-0.5 py-1.5 text-sm"
                        :class="{ 'sidebar-link-active': isNavItemActive(grandchild) }"
                        :id="getAdminMenuItemId(grandchild.path)"
                        @click="handleMenuItemClick(grandchild, $event)"
                      >
                        <span class="console-nav-icon-frame">
                          <Icon :name="grandchild.icon ?? 'document'" size="sm" />
                        </span>
                        <span class="min-w-0 truncate">{{ grandchild.label }}</span>
                      </router-link>
                    </div>
                  </template>
                  <router-link
                    v-else
                    :to="child.path"
                    v-bind="navLinkAttrs(child)"
                    class="sidebar-link mb-0.5 py-1.5 text-sm"
                    :class="{ 'sidebar-link-active': isNavItemActive(child) }"
                    :id="getAdminMenuItemId(child.path)"
                    @click="handleMenuItemClick(child, $event)"
                  >
                    <span class="console-nav-icon-frame">
                      <Icon :name="child.icon ?? 'document'" size="sm" />
                    </span>
                    <span class="min-w-0 truncate">{{ child.label }}</span>
                  </router-link>
                </template>
              </div>
            </template>
            <!-- Normal item (no children) -->
            <router-link
              v-else
              :to="item.path"
              v-bind="navLinkAttrs(item)"
              class="sidebar-link mb-1"
              :class="{ 'sidebar-link-active': isNavItemActive(item), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :id="getAdminMenuItemId(item.path)"
              @click="handleMenuItemClick(item, $event)"
            >
              <span class="console-nav-icon-frame">
                <span v-if="item.iconSvg" class="sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
                <Icon v-else :name="item.icon ?? 'document'" size="sm" />
              </span>
              <span class="sidebar-label sidebar-label-flex" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
                <span class="min-w-0 truncate">{{ item.label }}</span>
                <span v-if="showWelfareClaimBadge(item)" class="sidebar-claim-badge">{{ t('nav.claimQuota') }}</span>
                <span v-if="showInvoiceClaimBadge(item)" class="sidebar-claim-badge sidebar-invoice-claim-badge">{{ t('nav.claimInvoice') }}</span>
                <span v-if="showTicketUnreadBadge(item)" class="sidebar-unread-badge">{{ ticketUnreadBadgeLabel }}</span>
                <span v-if="showAdminTicketAttentionBadge(item)" class="sidebar-unread-badge sidebar-ticket-attention-badge">{{ adminTicketAttentionBadgeLabel }}</span>
              </span>
            </router-link>
          </template>
        </div>

        <!-- Personal Section for Admin (hidden in simple mode) -->
        <div v-if="!authStore.isSimpleMode" class="sidebar-section">
          <div class="sidebar-section-title" :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
            <span class="sidebar-section-title-text" :class="{ 'sidebar-section-title-text-collapsed': sidebarCollapsed }">
              {{ t('nav.myAccount') }}
            </span>
          </div>

          <template v-for="item in personalNavItems" :key="item.path">
            <template v-if="item.children?.length">
              <button
                type="button"
                class="sidebar-link mb-1 w-full"
                :class="{
                  'sidebar-link-active': isGroupActive(item) && !isGroupExpanded(item),
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :title="sidebarCollapsed ? item.label : undefined"
                @click="handleGroupClick(item)"
              >
                <span class="console-nav-icon-frame">
                  <Icon :name="item.icon ?? 'document'" size="sm" />
                </span>
                <span
                  class="sidebar-label sidebar-label-flex"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                >
                  <span class="min-w-0 truncate">{{ item.label }}</span>
                  <span
                    class="flex-shrink-0 transition-transform duration-200"
                    :class="isGroupExpanded(item) ? 'rotate-180' : ''"
                  >
                    <Icon name="chevronDown" size="xs" />
                  </span>
                </span>
              </button>
              <div v-if="!sidebarCollapsed && isGroupExpanded(item)" class="mb-1 ml-4 border-l-2 border-accent-300 pl-2 dark:border-accent-700">
                <router-link
                  v-for="child in item.children"
                  :key="child.path"
                  :to="child.path"
                  v-bind="navLinkAttrs(child)"
                  class="sidebar-link mb-0.5 py-1.5 text-sm"
                  :class="{ 'sidebar-link-active': isNavItemActive(child) }"
                  :data-tour="getSelfMenuItemTour(child.path)"
                  @click="handleMenuItemClick(child, $event)"
                >
                  <span class="console-nav-icon-frame">
                    <Icon :name="child.icon ?? 'document'" size="sm" />
                  </span>
                  <span class="sidebar-label-flex min-w-0 flex-1">
                    <span class="min-w-0 truncate">{{ child.label }}</span>
                    <span v-if="showWelfareClaimBadge(child)" class="sidebar-claim-badge">{{ t('nav.claimQuota') }}</span>
                    <span v-if="showInvoiceClaimBadge(child)" class="sidebar-claim-badge sidebar-invoice-claim-badge">{{ t('nav.claimInvoice') }}</span>
                    <span v-if="showTicketUnreadBadge(child)" class="sidebar-unread-badge">{{ ticketUnreadBadgeLabel }}</span>
                    <span v-if="showAdminTicketAttentionBadge(child)" class="sidebar-unread-badge sidebar-ticket-attention-badge">{{ adminTicketAttentionBadgeLabel }}</span>
                  </span>
                </router-link>
              </div>
            </template>
            <component
              v-else
              :is="navItemElement(item)"
              v-bind="navItemAttrs(item)"
              class="sidebar-link mb-1"
              :class="{ 'sidebar-link-active': isNavItemActive(item), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
              @click="handleMenuItemClick(item, $event)"
            >
              <span class="console-nav-icon-frame">
                <span v-if="item.iconSvg" class="sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
                <Icon v-else :name="item.icon ?? 'document'" size="sm" />
              </span>
              <span class="sidebar-label sidebar-label-flex" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
                <span class="min-w-0 truncate">{{ item.label }}</span>
                <span v-if="showWelfareClaimBadge(item)" class="sidebar-claim-badge">{{ t('nav.claimQuota') }}</span>
                <span v-if="showInvoiceClaimBadge(item)" class="sidebar-claim-badge sidebar-invoice-claim-badge">{{ t('nav.claimInvoice') }}</span>
                <span v-if="showTicketUnreadBadge(item)" class="sidebar-unread-badge">{{ ticketUnreadBadgeLabel }}</span>
                <span v-if="showAdminTicketAttentionBadge(item)" class="sidebar-unread-badge sidebar-ticket-attention-badge">{{ adminTicketAttentionBadgeLabel }}</span>
              </span>
            </component>
          </template>
        </div>
      </template>

      <!-- Regular User View -->
      <template v-else-if="!appStore.backendModeEnabled">
        <div class="sidebar-section">
          <template v-for="item in userNavItems" :key="item.path">
            <template v-if="item.children?.length">
              <button
                type="button"
                class="sidebar-link mb-1 w-full"
                :class="{
                  'sidebar-link-active': isGroupActive(item) && !isGroupExpanded(item),
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :title="sidebarCollapsed ? item.label : undefined"
                @click="handleGroupClick(item)"
              >
                <span class="console-nav-icon-frame">
                  <Icon :name="item.icon ?? 'document'" size="sm" />
                </span>
                <span
                  class="sidebar-label sidebar-label-flex"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                >
                  <span class="min-w-0 truncate">{{ item.label }}</span>
                  <span
                    class="flex-shrink-0 transition-transform duration-200"
                    :class="isGroupExpanded(item) ? 'rotate-180' : ''"
                  >
                    <Icon name="chevronDown" size="xs" />
                  </span>
                </span>
              </button>
              <div v-if="!sidebarCollapsed && isGroupExpanded(item)" class="mb-1 ml-4 border-l-2 border-accent-300 pl-2 dark:border-accent-700">
                <router-link
                  v-for="child in item.children"
                  :key="child.path"
                  :to="child.path"
                  v-bind="navLinkAttrs(child)"
                  class="sidebar-link mb-0.5 py-1.5 text-sm"
                  :class="{ 'sidebar-link-active': isNavItemActive(child) }"
                  :data-tour="getSelfMenuItemTour(child.path)"
                  @click="handleMenuItemClick(child, $event)"
                >
                  <span class="console-nav-icon-frame">
                    <Icon :name="child.icon ?? 'document'" size="sm" />
                  </span>
                  <span class="sidebar-label-flex min-w-0 flex-1">
                    <span class="min-w-0 truncate">{{ child.label }}</span>
                    <span v-if="showWelfareClaimBadge(child)" class="sidebar-claim-badge">{{ t('nav.claimQuota') }}</span>
                    <span v-if="showInvoiceClaimBadge(child)" class="sidebar-claim-badge sidebar-invoice-claim-badge">{{ t('nav.claimInvoice') }}</span>
                    <span v-if="showTicketUnreadBadge(child)" class="sidebar-unread-badge">{{ ticketUnreadBadgeLabel }}</span>
                    <span v-if="showAdminTicketAttentionBadge(child)" class="sidebar-unread-badge sidebar-ticket-attention-badge">{{ adminTicketAttentionBadgeLabel }}</span>
                  </span>
                </router-link>
              </div>
            </template>
            <component
              v-else
              :is="navItemElement(item)"
              v-bind="navItemAttrs(item)"
              class="sidebar-link mb-1"
              :class="{ 'sidebar-link-active': isNavItemActive(item), 'sidebar-link-collapsed': sidebarCollapsed }"
              :title="sidebarCollapsed ? item.label : undefined"
              :data-tour="getSelfMenuItemTour(item.path)"
              @click="handleMenuItemClick(item, $event)"
            >
              <span class="console-nav-icon-frame">
                <span v-if="item.iconSvg" class="sidebar-svg-icon" v-html="sanitizeSvg(item.iconSvg)"></span>
                <Icon v-else :name="item.icon ?? 'document'" size="sm" />
              </span>
              <span class="sidebar-label sidebar-label-flex" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">
                <span class="min-w-0 truncate">{{ item.label }}</span>
                <span v-if="showWelfareClaimBadge(item)" class="sidebar-claim-badge">{{ t('nav.claimQuota') }}</span>
                <span v-if="showInvoiceClaimBadge(item)" class="sidebar-claim-badge sidebar-invoice-claim-badge">{{ t('nav.claimInvoice') }}</span>
                <span v-if="showTicketUnreadBadge(item)" class="sidebar-unread-badge">{{ ticketUnreadBadgeLabel }}</span>
                <span v-if="showAdminTicketAttentionBadge(item)" class="sidebar-unread-badge sidebar-ticket-attention-badge">{{ adminTicketAttentionBadgeLabel }}</span>
              </span>
            </component>
          </template>
        </div>
      </template>
    </nav>

    <!-- Bottom Section -->
    <div class="console-sidebar-footer mt-auto p-3">
      <!-- Theme Toggle -->
      <button
        @click="toggleTheme"
        class="sidebar-link mb-2 w-full"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? (isDark ? t('nav.lightMode') : t('nav.darkMode')) : undefined"
      >
        <span class="console-nav-icon-frame">
          <Icon v-if="isDark" name="sun" size="sm" />
          <Icon v-else name="moon" size="sm" />
        </span>
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{
          isDark ? t('nav.lightMode') : t('nav.darkMode')
        }}</span>
      </button>

      <!-- Collapse Button -->
      <button
        @click="toggleSidebar"
        class="sidebar-link w-full"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? t('nav.expand') : t('nav.collapse')"
      >
        <span class="console-nav-icon-frame">
          <Icon v-if="!sidebarCollapsed" name="chevronLeft" size="sm" />
          <Icon v-else name="chevronRight" size="sm" />
        </span>
        <span class="sidebar-label" :class="{ 'sidebar-label-collapsed': sidebarCollapsed }" :aria-hidden="sidebarCollapsed ? 'true' : 'false'">{{ t('nav.collapse') }}</span>
      </button>
    </div>
  </aside>

  <!-- Mobile Overlay -->
  <transition name="fade">
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-30 bg-black/50 lg:hidden"
      @click="closeMobile"
    ></div>
  </transition>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore, useWelfareStore } from '@/stores'
import VersionBadge from '@/components/common/VersionBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { studioBridgeAPI } from '@/api'
import { ticketsAPI } from '@/api/tickets'
import { paymentAPI } from '@/api/payment'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { sanitizeSvg } from '@/utils/sanitize'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import { resolveGroupBuyProductName } from '@/utils/groupBuyProduct'

type IconName = InstanceType<typeof Icon>['$props']['name']
type NavItemAction = 'studioBridgeLaunch'

interface NavItem {
  path: string
  label: string
  icon?: IconName
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: NavItem[]
  exactActive?: boolean
  openInNewTab?: boolean
  action?: NavItemAction
  /**
   * When true, the parent item only toggles the expand/collapse state and
   * does NOT navigate to its `path`. The `path` is purely a stable key.
   */
  expandOnly?: boolean
  /**
   * 可选的功能开关 getter。返回 false 时菜单项被隐藏；返回 undefined/true 时显示。
   * 宽容策略（undefined → 显示）避免 public settings 未加载完成时菜单闪烁消失。
   * Getter 里访问的 reactive 来源（store / composable）会被 computed 自动追踪，
   * 开关切换时菜单自动更新。
   */
  featureFlag?: () => boolean | undefined
}

// applyFeatureFlags 递归过滤掉 featureFlag() === false 的节点（含子节点）。
// 使用 `!== false` 宽容语义：undefined（设置未加载）或 true 都视为显示。
function applyFeatureFlags(items: NavItem[]): NavItem[] {
  const out: NavItem[] = []
  for (const item of items) {
    if (item.featureFlag && item.featureFlag() === false) continue
    if (item.children) {
      const children = applyFeatureFlags(item.children)
      if (children.length === 0 && item.expandOnly) continue
      out.push({ ...item, children })
    } else {
      out.push(item)
    }
  }
  return out
}

function applySimpleMode(items: NavItem[]): NavItem[] {
  const out: NavItem[] = []
  for (const item of items) {
    if (item.hideInSimpleMode) continue
    if (item.children) {
      const children = applySimpleMode(item.children)
      if (children.length === 0 && item.expandOnly) continue
      out.push({ ...item, children })
    } else {
      out.push(item)
    }
  }
  return out
}

const { t } = useI18n()

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const adminSettingsStore = useAdminSettingsStore()
const welfareStore = useWelfareStore()

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)
const isDark = ref(document.documentElement.classList.contains('dark'))

// Track which parent nav groups are expanded
const expandedGroups = ref<Set<string>>(new Set())

// Site settings from appStore (cached, no flicker)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => appStore.siteLogo)
const siteVersion = computed(() => appStore.siteVersion)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const groupBuyProductName = computed(() => resolveGroupBuyProductName(appStore.cachedPublicSettings))

// Console navigation uses line icons for legibility; the public homepage keeps the pixel icon system.
const DashboardIcon: IconName = 'grid'
const KeyIcon: IconName = 'key'
const ChartIcon: IconName = 'chart'
const OpsIcon: IconName = 'cpu'
const UsageIcon: IconName = 'chartBar'
const TrophyIcon: IconName = 'trendingUp'
const GiftIcon: IconName = 'gift'
const WelfareIcon: IconName = 'sparkles'
const ChatIcon: IconName = 'chatBubble'
const UserIcon: IconName = 'user'
const UsersIcon: IconName = 'users'
const TeamIcon: IconName = 'userPlus'
const FolderIcon: IconName = 'folder'
const ChannelIcon: IconName = 'terminal'
const CreditCardIcon: IconName = 'creditCard'
const RechargeSubscriptionIcon: IconName = 'dollar'
const GlobeIcon: IconName = 'globe'
const TeamShareIcon: IconName = 'userPlus'
const ServerIcon: IconName = 'server'
const BellIcon: IconName = 'bell'
const TicketIcon: IconName = 'ticket'
const CogIcon: IconName = 'cog'
const OrderIcon: IconName = 'clipboard'
const SignalIcon: IconName = 'sync'
const ShieldIcon: IconName = 'shield'
const PriceTagIcon: IconName = 'tag'
const ContentIcon: IconName = 'book'

// Public-settings flags go through the registry in utils/featureFlags.ts,
// which handles the opt-in vs opt-out fallback when settings haven't loaded
// yet. Admin-only flags (not in public settings) stay inline below.
const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagPayment = makeSidebarFlag(FeatureFlags.payment)
const flagGroupBuy = makeSidebarFlag(FeatureFlags.groupBuy)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagAccountShare = makeSidebarFlag(FeatureFlags.accountShare)
const flagRiskControl = makeSidebarFlag(FeatureFlags.riskControl)
const flagWelfare = makeSidebarFlag(FeatureFlags.welfare)
const welfareClaimBadgeVisible = computed(() => welfareStore.hasClaimableReward)
const ticketUnreadTotal = ref(0)
const adminTicketUnreadTotal = ref(0)
const invoiceClaimTotal = ref(0)
const ticketUnreadBadgeLabel = computed(() => (ticketUnreadTotal.value > 99 ? '99+' : String(ticketUnreadTotal.value)))
const adminTicketAttentionBadgeLabel = computed(() => (adminTicketUnreadTotal.value > 99 ? '99+' : String(adminTicketUnreadTotal.value)))
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled
const flagGroupBuyUser = () => flagPayment() !== false && flagGroupBuy() !== false
const WELFARE_BADGE_REFRESH_MS = 60_000
const TICKET_UNREAD_BADGE_REFRESH_MS = 60_000
const SIDEBAR_TOUR_TARGET_EVENT = 'sub2api:sidebar-tour-target'
const TICKET_UNREAD_BADGE_REFRESH_EVENT = 'sub2api:ticket-unread-updated'
const sidebarTourTargetGroups: Record<string, string[]> = {
  '#sidebar-group-manage': ['/admin/basic-management'],
  '#sidebar-channel-manage': ['/admin/basic-management'],
  '#sidebar-wallet': ['/admin/business-operations'],
}

// buildSelfNavItems builds user navigation; admin personal navigation reuses it.
function buildSelfNavItems(withDashboard: boolean): NavItem[] {
  const items: NavItem[] = []
  if (withDashboard) {
    items.push({ path: '/dashboard', label: t('nav.dashboard'), icon: DashboardIcon })
  }

  const welfareItem: NavItem = { path: '/welfare', label: t('nav.welfare'), icon: WelfareIcon, hideInSimpleMode: true, featureFlag: flagWelfare }
  const accountItems: NavItem[] = [
    { path: '/my-accounts', label: t('nav.myAccounts'), icon: GlobeIcon, hideInSimpleMode: true, featureFlag: flagAccountShare },
    { path: '/profile', label: t('nav.profile'), icon: UserIcon },
  ]

  const usageStatusItems: NavItem[] = [
    { path: '/leaderboard', label: t('nav.leaderboard'), icon: TrophyIcon, hideInSimpleMode: true },
    { path: '/available-channels', label: t('nav.availableChannels'), icon: ChannelIcon, hideInSimpleMode: true, featureFlag: flagAvailableChannels },
    { path: '/monitor', label: t('nav.channelStatus'), icon: SignalIcon, featureFlag: flagChannelMonitor },
  ]
  const primarySelfItems: NavItem[] = [
    { path: '/chat-images', label: t('nav.chatImageCreator'), icon: ChatIcon, hideInSimpleMode: true, openInNewTab: true, action: 'studioBridgeLaunch' },
    { path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon },
    { path: '/usage', label: t('nav.usageAndSubscriptions'), icon: UsageIcon, hideInSimpleMode: true },
    { path: '/tickets', label: t('nav.tickets'), icon: TicketIcon, hideInSimpleMode: true },
    { path: '/purchase', label: t('nav.buySubscription'), icon: RechargeSubscriptionIcon, hideInSimpleMode: true, featureFlag: flagPayment },
    { path: '/group-buy', label: groupBuyProductName.value, icon: GiftIcon, hideInSimpleMode: true, featureFlag: flagGroupBuyUser },
    { path: '/affiliate', label: t('nav.affiliate'), icon: TeamIcon, hideInSimpleMode: true, featureFlag: flagAffiliate },
    welfareItem,
  ]

  items.push(
    ...primarySelfItems,
    ...usageStatusItems,
    ...accountItems,
  )

  items.push(
    ...customMenuItemsForUser.value.map((item): NavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      iconSvg: item.icon_svg,
    })),
  )
  return items
}

// finalizeNav 合并三重过滤：featureFlag 过滤 + simple 模式过滤。
function finalizeNav(items: NavItem[]): NavItem[] {
  const visible = applyFeatureFlags(items)
  return authStore.isSimpleMode ? applySimpleMode(visible) : visible
}

// User navigation items (for regular users)
const userNavItems = computed((): NavItem[] => finalizeNav(buildSelfNavItems(true)))

// Personal navigation items (for admin's "My Account" section, without Dashboard).
// Admins access 可用渠道 from this section just like regular users — there is no
// separate admin entry, since the page is purely a user-facing view.
const personalNavItems = computed((): NavItem[] => finalizeNav(buildSelfNavItems(false)))

// Custom menu items filtered by visibility
const customMenuItemsForUser = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
})

const customMenuItemsForAdmin = computed(() => {
  return adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
})

// Admin navigation items
const adminNavItems = computed((): NavItem[] => {
  const baseItems: NavItem[] = [
    { path: '/admin/dashboard', label: t('nav.dashboard'), icon: DashboardIcon },
    { path: '/admin/ops', label: t('nav.ops'), icon: OpsIcon, featureFlag: flagOpsMonitoring },
    {
      path: '/admin/basic-management',
      label: t('nav.basicManagement'),
      icon: UsersIcon,
      expandOnly: true,
      children: [
        { path: '/admin/users', label: t('nav.users'), icon: UsersIcon, hideInSimpleMode: true },
        { path: '/admin/groups', label: t('nav.groups'), icon: FolderIcon, hideInSimpleMode: true },
        { path: '/admin/accounts', label: t('nav.accounts'), icon: GlobeIcon },
        { path: '/admin/shared-accounts', label: t('nav.sharedAccounts'), icon: TeamShareIcon, featureFlag: flagAccountShare },
        { path: '/admin/proxies', label: t('nav.proxies'), icon: ServerIcon },
      ],
    },
    {
      path: '/admin/channels',
      label: t('nav.channelManagement'),
      icon: ChannelIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      children: [
        { path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: PriceTagIcon },
        { path: '/admin/channels/monitor', label: t('nav.channelMonitor'), icon: SignalIcon, featureFlag: flagChannelMonitor },
      ],
    },
    {
      path: '/admin/resource-access',
      label: t('nav.resourceAccess'),
      icon: ShieldIcon,
      expandOnly: true,
      children: [
        { path: '/admin/risk-control', label: t('nav.riskControl'), icon: ShieldIcon, hideInSimpleMode: true, featureFlag: flagRiskControl },
      ],
    },
    {
      path: '/admin/orders',
      label: t('nav.orderManagement'),
      icon: OrderIcon,
      hideInSimpleMode: true,
      expandOnly: true,
      featureFlag: flagAdminPayment,
      children: [
        { path: '/admin/orders/dashboard', label: t('nav.paymentDashboard'), icon: ChartIcon },
        { path: '/admin/orders', label: t('nav.orderList'), icon: OrderIcon, exactActive: true },
        { path: '/admin/orders/invoices', label: t('nav.invoiceRequests'), icon: TicketIcon },
      ],
    },
    {
      path: '/admin/business-operations',
      label: t('nav.businessOperations'),
      icon: OrderIcon,
      expandOnly: true,
      children: [
        { path: '/admin/subscriptions', label: t('nav.subscriptions'), icon: CreditCardIcon, hideInSimpleMode: true },
        { path: '/admin/orders/plans', label: t('nav.paymentPlans'), icon: PriceTagIcon, hideInSimpleMode: true, featureFlag: flagAdminPayment },
        { path: '/admin/group-buy', label: t('nav.groupBuyManagement'), icon: GiftIcon, hideInSimpleMode: true, featureFlag: flagAdminPayment },
        { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: TicketIcon, hideInSimpleMode: true },
        { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: GiftIcon, hideInSimpleMode: true },
        {
          path: '/admin/affiliates',
          label: t('nav.affiliateManagement'),
          icon: UsersIcon,
          hideInSimpleMode: true,
          expandOnly: true,
          featureFlag: flagAffiliate,
          children: [
            { path: '/admin/affiliates/invites', label: t('nav.affiliateInviteRecords'), icon: TeamIcon },
            { path: '/admin/affiliates/rebates', label: t('nav.affiliateRebateRecords'), icon: OrderIcon },
            { path: '/admin/affiliates/transfers', label: t('nav.affiliateTransferRecords'), icon: CreditCardIcon },
          ],
        },
      ],
    },
    { path: '/admin/usage', label: t('nav.usage'), icon: UsageIcon },
    { path: '/admin/audit-logs', label: t('nav.auditLogs'), icon: ShieldIcon, hideInSimpleMode: true },
    { path: '/admin/tickets', label: t('nav.supportTickets'), icon: TicketIcon, hideInSimpleMode: true },
    {
      path: '/admin/content-records',
      label: t('nav.contentRecords'),
      icon: ContentIcon,
      expandOnly: true,
      children: [
        { path: '/admin/announcements', label: t('nav.announcements'), icon: BellIcon },
        { path: '/admin/tutorials', label: t('nav.tutorials'), icon: ContentIcon },
        { path: '/admin/model-market', label: t('nav.modelMarket'), icon: PriceTagIcon },
      ],
    },
  ]

  const visible = applyFeatureFlags(baseItems)

  // 简单模式下，在系统设置前插入 API密钥
  if (authStore.isSimpleMode) {
    const filtered = applySimpleMode(visible)
    filtered.push({ path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon })
    filtered.push({ path: '/admin/settings', label: t('nav.settings'), icon: CogIcon })
    for (const cm of customMenuItemsForAdmin.value) {
      filtered.push({ path: `/custom/${cm.id}`, label: cm.label, iconSvg: cm.icon_svg })
    }
    return filtered
  }

  visible.push({ path: '/admin/settings', label: t('nav.settings'), icon: CogIcon })
  for (const cm of customMenuItemsForAdmin.value) {
    visible.push({ path: `/custom/${cm.id}`, label: cm.label, iconSvg: cm.icon_svg })
  }
  return visible
})

function toggleSidebar() {
  appStore.toggleSidebar()
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function closeMobile() {
  appStore.setMobileOpen(false)
}

function expandGroupsForTourTarget(selector: string) {
  const groupPaths = sidebarTourTargetGroups[selector]
  if (!groupPaths) return
  for (const groupPath of groupPaths) {
    expandedGroups.value.add(groupPath)
  }
}

function handleSidebarTourTarget(event: Event) {
  const selector = (event as CustomEvent<{ selector?: string }>).detail?.selector
  if (selector) {
    expandGroupsForTourTarget(selector)
  }
}

function writeStudioBridgePopupPlaceholder(popup: Window): void {
  const title = t('chatImageStudio.launchingTitle')
  popup.document.title = title
  popup.document.body.style.margin = '0'
  popup.document.body.style.background = '#f8fafc'
  const placeholder = popup.document.createElement('div')
  placeholder.style.display = 'grid'
  placeholder.style.minHeight = '100vh'
  placeholder.style.placeItems = 'center'
  placeholder.style.fontFamily = "system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif"
  placeholder.style.color = '#475569'
  placeholder.textContent = title
  popup.document.body.replaceChildren(placeholder)
}

async function launchStudioBridgeFromSidebar(): Promise<void> {
  const popup = window.open('about:blank', '_blank')
  if (popup) {
    popup.opener = null
    writeStudioBridgePopupPlaceholder(popup)
  }
  try {
    const result = await studioBridgeAPI.launch()
    if (popup && !popup.closed) {
      popup.location.href = result.launch_url
      return
    }
    window.location.assign(result.launch_url)
  } catch (error) {
    if (popup && !popup.closed) {
      popup.close()
    }
    console.error('Failed to launch Luoye Creative studio:', error)
    appStore.showError(t('chatImageStudio.launchFailedDescription'))
    void router.push('/chat-images')
  }
}

function handleMenuItemClick(item: NavItem, event?: MouseEvent) {
  if (item.action === 'studioBridgeLaunch') {
    event?.preventDefault()
    void launchStudioBridgeFromSidebar()
    if (mobileOpen.value) {
      setTimeout(() => {
        appStore.setMobileOpen(false)
      }, 150)
    }
    return
  }

  if (mobileOpen.value) {
    setTimeout(() => {
      appStore.setMobileOpen(false)
    }, 150)
  }

  // Map paths to tour selectors
  const pathToSelector: Record<string, string> = {
    '/admin/groups': '#sidebar-group-manage',
    '/admin/accounts': '#sidebar-channel-manage',
    '/keys': '[data-tour="sidebar-my-keys"]'
  }

  const selector = pathToSelector[item.path]
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function getSelfMenuItemTour(path: string): string | undefined {
  if (path === '/keys') return 'sidebar-my-keys'
  return undefined
}

function getAdminMenuItemId(path: string): string | undefined {
  if (path === '/admin/accounts') return 'sidebar-channel-manage'
  if (path === '/admin/groups') return 'sidebar-group-manage'
  if (path === '/admin/redeem') return 'sidebar-wallet'
  return undefined
}

function showWelfareClaimBadge(item: NavItem): boolean {
  return item.path === '/welfare' && welfareClaimBadgeVisible.value
}

function showTicketUnreadBadge(item: NavItem): boolean {
  return item.path === '/tickets' && ticketUnreadTotal.value > 0
}

function showInvoiceClaimBadge(item: NavItem): boolean {
  return item.path === '/tickets' && invoiceClaimTotal.value > 0
}

function showAdminTicketAttentionBadge(item: NavItem): boolean {
  return item.path === '/admin/tickets' && adminTicketUnreadTotal.value > 0
}

function navLinkAttrs(item: NavItem): Record<string, string> {
  if (!item.openInNewTab || item.action) return {}
  return {
    target: '_blank',
    rel: 'noopener noreferrer',
  }
}

function navItemElement(item: NavItem): 'button' | 'router-link' {
  return item.action ? 'button' : 'router-link'
}

function navItemAttrs(item: NavItem): Record<string, string> {
  if (item.action) return { type: 'button' }
  return {
    to: item.path,
    ...navLinkAttrs(item),
  }
}

function canRefreshWelfareBadge(): boolean {
  return authStore.isAuthenticated && flagWelfare() && !authStore.isSimpleMode
}

function refreshWelfareBadge(force = false): void {
  if (!canRefreshWelfareBadge()) {
    welfareStore.reset()
    return
  }
  void welfareStore.fetchOverview(force)
}

let welfareBadgeRefreshTimer: ReturnType<typeof setInterval> | null = null
let ticketUnreadBadgeRefreshTimer: ReturnType<typeof setInterval> | null = null

function startWelfareBadgeRefreshTimer(): void {
  if (welfareBadgeRefreshTimer) return
  welfareBadgeRefreshTimer = setInterval(() => {
    refreshWelfareBadge(true)
  }, WELFARE_BADGE_REFRESH_MS)
}

function stopWelfareBadgeRefreshTimer(): void {
  if (!welfareBadgeRefreshTimer) return
  clearInterval(welfareBadgeRefreshTimer)
  welfareBadgeRefreshTimer = null
}

function handleWelfareBadgeVisibilityChange(): void {
  if (document.visibilityState === 'visible') {
    refreshWelfareBadge(true)
    void refreshTicketUnreadBadge()
  }
}

function canRefreshTicketUnreadBadge(): boolean {
  return authStore.isAuthenticated && !authStore.isAdmin && !authStore.isSimpleMode
}

async function refreshInvoiceClaimBadge(): Promise<void> {
  if (!canRefreshTicketUnreadBadge()) {
    invoiceClaimTotal.value = 0
    return
  }
  try {
    const response = await paymentAPI.getInvoiceClaimSummary()
    invoiceClaimTotal.value = Math.max(0, Number(response.data.claimable_count) || 0)
  } catch {
    invoiceClaimTotal.value = 0
  }
}

function canRefreshAdminTicketAttentionBadge(): boolean {
  return authStore.isAuthenticated && authStore.isAdmin && !authStore.isSimpleMode
}

async function refreshTicketUnreadBadge(): Promise<void> {
  if (canRefreshTicketUnreadBadge()) {
    try {
      const summary = await ticketsAPI.unreadSummary()
      ticketUnreadTotal.value = Math.max(0, Number(summary.total_unread) || 0)
    } catch {
      ticketUnreadTotal.value = 0
    }
  } else {
    ticketUnreadTotal.value = 0
  }
  await refreshInvoiceClaimBadge()

  if (canRefreshAdminTicketAttentionBadge()) {
    try {
      const response = await adminTicketsAPI.list({
        page: 1,
        page_size: 1,
        unread_only: true,
        sort_by: 'unread_first',
        sort_order: 'desc',
      })
      adminTicketUnreadTotal.value = Math.max(0, Number(response.total) || 0)
    } catch {
      adminTicketUnreadTotal.value = 0
    }
  } else {
    adminTicketUnreadTotal.value = 0
  }
}

function startTicketUnreadBadgeRefreshTimer(): void {
  if (ticketUnreadBadgeRefreshTimer) return
  ticketUnreadBadgeRefreshTimer = setInterval(() => {
    void refreshTicketUnreadBadge()
  }, TICKET_UNREAD_BADGE_REFRESH_MS)
}

function stopTicketUnreadBadgeRefreshTimer(): void {
  if (!ticketUnreadBadgeRefreshTimer) return
  clearInterval(ticketUnreadBadgeRefreshTimer)
  ticketUnreadBadgeRefreshTimer = null
}

function isActive(path: string, exact = false): boolean {
  if (exact) return route.path === path
  return route.path === path || route.path.startsWith(path + '/')
}

function isNavItemActive(item: NavItem): boolean {
  return isActive(item.path, item.exactActive)
}

function isGroupActive(item: NavItem): boolean {
  if (!item.children) return false
  return item.children.some(child => isNavItemActive(child) || isGroupActive(child))
}

function isGroupExpanded(item: NavItem): boolean {
  return expandedGroups.value.has(item.path) || isGroupActive(item)
}

function toggleGroup(item: NavItem) {
  if (expandedGroups.value.has(item.path)) {
    expandedGroups.value.delete(item.path)
  } else {
    expandedGroups.value.add(item.path)
  }
}

/**
 * Click handler for collapsible parent items.
 * - When sidebar is collapsed: do nothing (children are not visible).
 * - When `expandOnly` is true: only toggle expand state.
 * - Otherwise (default, e.g. /admin/orders): navigate to the parent path
 *   (router-link semantics) and ensure the group is expanded.
 */
function handleGroupClick(item: NavItem) {
  if (sidebarCollapsed.value) return
  if (item.expandOnly) {
    toggleGroup(item)
    return
  }
  // Push to path and ensure expanded
  if (route.path !== item.path) {
    router.push(item.path)
  }
  if (!expandedGroups.value.has(item.path)) {
    expandedGroups.value.add(item.path)
  }
}

// Initialize theme
const savedTheme = localStorage.getItem('theme')
if (savedTheme === 'dark') {
  isDark.value = true
  document.documentElement.classList.add('dark')
}

// Fetch admin settings (for feature-gated nav items like Ops).
watch(
  isAdmin,
  (v) => {
    if (v) {
      adminSettingsStore.fetch()
    }
  },
  { immediate: true }
)

watch(
  () => [authStore.isAuthenticated, flagWelfare(), authStore.isSimpleMode, route.path] as const,
  ([authenticated, welfareEnabled, simpleMode]) => {
    if (!authenticated || !welfareEnabled || simpleMode) {
      refreshWelfareBadge()
      return
    }
    refreshWelfareBadge()
  },
  { immediate: true }
)

watch(
  () => [authStore.isAuthenticated, authStore.isAdmin, authStore.isSimpleMode, route.path] as const,
  () => {
    void refreshTicketUnreadBadge()
  },
  { immediate: true }
)

onMounted(() => {
  if (isAdmin.value) {
    adminSettingsStore.fetch()
  }
  startWelfareBadgeRefreshTimer()
  startTicketUnreadBadgeRefreshTimer()
  document.addEventListener('visibilitychange', handleWelfareBadgeVisibilityChange)
  window.addEventListener(SIDEBAR_TOUR_TARGET_EVENT, handleSidebarTourTarget)
  window.addEventListener(TICKET_UNREAD_BADGE_REFRESH_EVENT, refreshTicketUnreadBadge)
})

onUnmounted(() => {
  stopWelfareBadgeRefreshTimer()
  stopTicketUnreadBadgeRefreshTimer()
  document.removeEventListener('visibilitychange', handleWelfareBadgeVisibilityChange)
  window.removeEventListener(SIDEBAR_TOUR_TARGET_EVENT, handleSidebarTourTarget)
  window.removeEventListener(TICKET_UNREAD_BADGE_REFRESH_EVENT, refreshTicketUnreadBadge)
})
</script>

<style scoped>
.sidebar-logo {
  flex: 0 0 2.25rem;
  min-width: 2.25rem;
}

.sidebar-header-collapsed {
  gap: 0;
  padding-left: 1.125rem;
  padding-right: 1.125rem;
}

.sidebar-brand {
  min-width: 0;
  flex: 1 1 auto;
  white-space: nowrap;
  transition:
    max-width 0.22s ease,
    opacity 0.14s ease,
    transform 0.14s ease;
  max-width: 12rem;
}

.sidebar-brand-collapsed {
  max-width: 0;
  overflow: hidden;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

.sidebar-brand-title {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-link {
  min-height: 2.25rem;
  padding-top: 0.45rem;
  padding-bottom: 0.45rem;
}

.sidebar-link-collapsed {
  gap: 0;
  padding-left: 0.875rem;
  padding-right: 0.875rem;
}

.sidebar-section {
  margin-bottom: 1rem;
}

.sidebar-section-title {
  position: relative;
  display: flex;
  align-items: center;
  min-height: 1rem;
  overflow: hidden;
  white-space: nowrap;
}

.sidebar-section-title-text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    opacity 0.16s ease,
    transform 0.16s ease;
}

.sidebar-section-title::after {
  content: '';
  position: absolute;
  left: 0.75rem;
  right: 0.75rem;
  top: 50%;
  height: 1px;
  background: rgb(229 231 235);
  opacity: 0;
  transform: translateY(-50%);
  transition: opacity 0.18s ease;
}

.dark .sidebar-section-title::after {
  background: rgb(55 65 81);
}

.sidebar-section-title-text-collapsed {
  opacity: 0;
  transform: translateX(-4px);
}

.sidebar-section-title-collapsed::after {
  opacity: 1;
  transition-delay: 0.08s;
}

.sidebar-label {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition:
    max-width 0.2s ease,
    opacity 0.12s ease,
    transform 0.12s ease;
  max-width: 12rem;
}

.sidebar-label-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.sidebar-claim-badge {
  flex: 0 0 auto;
  max-width: 4rem;
  overflow: hidden;
  border-radius: 9999px;
  border: 1px solid rgba(204, 120, 92, 0.18);
  background: #f3e7df;
  padding: 0.125rem 0.375rem;
  color: #a9583e;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dark .sidebar-claim-badge {
  border-color: rgba(240, 184, 158, 0.26);
  background: rgba(204, 120, 92, 0.16);
  color: #f0b89e;
}

.sidebar-invoice-claim-badge {
  max-width: 4.75rem;
  background: linear-gradient(135deg, rgb(234 88 12), rgb(202 138 4));
  color: white;
  box-shadow: 0 0 0 0 rgba(251, 146, 60, 0.36);
  animation: sidebar-ticket-attention-pulse 1.15s ease-in-out infinite;
}

.dark .sidebar-invoice-claim-badge {
  background: linear-gradient(135deg, rgb(249 115 22), rgb(234 179 8));
  color: rgb(255 251 235);
}

.sidebar-unread-badge {
  flex: 0 0 auto;
  min-width: 1.25rem;
  max-width: 2.5rem;
  overflow: hidden;
  border-radius: 9999px;
  background: rgb(239 68 68);
  padding: 0.125rem 0.375rem;
  color: white;
  font-size: 0.6875rem;
  font-weight: 700;
  line-height: 1rem;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-ticket-attention-badge {
  background: linear-gradient(135deg, rgb(239 68 68), rgb(245 158 11));
  box-shadow: 0 0 0 0 rgba(248, 113, 113, 0.45);
  transform-origin: center;
  animation: sidebar-ticket-attention-pulse 1.15s ease-in-out infinite;
}

.dark .sidebar-ticket-attention-badge {
  box-shadow: 0 0 0 0 rgba(251, 146, 60, 0.38);
}

@keyframes sidebar-ticket-attention-pulse {
  0%,
  100% {
    transform: scale(1);
    opacity: 1;
    box-shadow: 0 0 0 0 rgba(248, 113, 113, 0.45);
  }

  50% {
    transform: scale(1.14);
    opacity: 0.74;
    box-shadow: 0 0 0 0.32rem rgba(248, 113, 113, 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .sidebar-ticket-attention-badge,
  .sidebar-invoice-claim-badge {
    animation: none;
  }
}

.sidebar-label-collapsed {
  max-width: 0;
  opacity: 0;
  transform: translateX(-4px);
  pointer-events: none;
}

/* Custom SVG icon in sidebar: constrain size without overriding uploaded SVG colors */
.sidebar-svg-icon {
  color: currentColor;
}

.sidebar-svg-icon :deep(svg) {
  display: block;
  width: 1.25rem;
  height: 1.25rem;
}
</style>
