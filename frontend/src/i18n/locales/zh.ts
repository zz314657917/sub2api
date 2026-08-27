import home from './zh/home'
import keyUsage from './zh/keyUsage'
import setup from './zh/setup'
import common from './zh/common'
import nav from './zh/nav'
import auth from './zh/auth'
import dashboard from './zh/dashboard'
import leaderboard from './zh/leaderboard'
import welfare from './zh/welfare'
import groups from './zh/groups'
import keys from './zh/keys'
import chatStudio from './zh/chatStudio'
import chatImageStudio from './zh/chatImageStudio'
import imageCreator from './zh/imageCreator'
import imageManager from './zh/imageManager'
import canvas from './zh/canvas'
import usage from './zh/usage'
import monitorCommon from './zh/monitorCommon'
import channelStatus from './zh/channelStatus'
import availableChannels from './zh/availableChannels'
import myAccounts from './zh/myAccounts'
import affiliate from './zh/affiliate'
import redeem from './zh/redeem'
import profile from './zh/profile'
import empty from './zh/empty'
import table from './zh/table'
import pagination from './zh/pagination'
import errors from './zh/errors'
import dates from './zh/dates'
import admin from './zh/admin'
import subscriptionProgress from './zh/subscriptionProgress'
import version from './zh/version'
import purchase from './zh/purchase'
import customPage from './zh/customPage'
import announcements from './zh/announcements'
import userSubscriptions from './zh/userSubscriptions'
import onboarding from './zh/onboarding'
import payment from './zh/payment'
import tickets from './zh/tickets'

const usageWithCyberPolicy = {
  ...usage,
  cyber: '安全策略',
}

const adminWithCyberPolicy = {
  ...admin,
  riskControl: {
    ...admin.riskControl,
    cyberPolicyExcludeBan: 'cyber_policy 不计入封号次数',
    cyberPolicyExcludeBanHint: '开启后，cyber_policy 拦截不再计入当次及历史自动封号累计；风控日志与通知邮件照常记录。',
    violationNotCounted: '未计入封号',
    action: {
      ...admin.riskControl.action,
      cyberPolicy: '网络安全策略',
    },
  },
  settings: {
    ...admin.settings,
    features: {
      ...admin.settings.features,
      riskControl: {
        ...admin.settings.features.riskControl,
        cyberSessionBlock: 'cyber 会话自动屏蔽',
        cyberSessionBlockHint: '上游命中 cyber_policy 后，只屏蔽当前 API Key 与显式会话的精确组合；同一 Key 的其他会话不受影响。',
        cyberSessionBlockTTL: '屏蔽时长（秒）',
      },
    },
  },
}

export default {
  home,
  keyUsage,
  setup,
  common,
  nav,
  auth,
  dashboard,
  leaderboard,
  welfare,
  groups,
  keys,
  chatStudio,
  chatImageStudio,
  imageCreator,
  imageManager,
  canvas,
  usage: usageWithCyberPolicy,
  monitorCommon,
  channelStatus,
  availableChannels,
  myAccounts,
  affiliate,
  redeem,
  profile,
  empty,
  table,
  pagination,
  errors,
  dates,
  admin: adminWithCyberPolicy,
  subscriptionProgress,
  version,
  purchase,
  customPage,
  announcements,
  userSubscriptions,
  onboarding,
  payment,
  tickets,
}
