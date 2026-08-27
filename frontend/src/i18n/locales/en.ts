import home from './en/home'
import keyUsage from './en/keyUsage'
import setup from './en/setup'
import common from './en/common'
import nav from './en/nav'
import auth from './en/auth'
import dashboard from './en/dashboard'
import leaderboard from './en/leaderboard'
import welfare from './en/welfare'
import groups from './en/groups'
import keys from './en/keys'
import chatStudio from './en/chatStudio'
import chatImageStudio from './en/chatImageStudio'
import imageCreator from './en/imageCreator'
import imageManager from './en/imageManager'
import canvas from './en/canvas'
import usage from './en/usage'
import monitorCommon from './en/monitorCommon'
import channelStatus from './en/channelStatus'
import availableChannels from './en/availableChannels'
import myAccounts from './en/myAccounts'
import affiliate from './en/affiliate'
import redeem from './en/redeem'
import profile from './en/profile'
import empty from './en/empty'
import table from './en/table'
import pagination from './en/pagination'
import errors from './en/errors'
import dates from './en/dates'
import admin from './en/admin'
import subscriptionProgress from './en/subscriptionProgress'
import version from './en/version'
import purchase from './en/purchase'
import customPage from './en/customPage'
import announcements from './en/announcements'
import userSubscriptions from './en/userSubscriptions'
import onboarding from './en/onboarding'
import payment from './en/payment'
import tickets from './en/tickets'

const usageWithCyberPolicy = {
  ...usage,
  cyber: 'Cyber',
}

const adminWithCyberPolicy = {
  ...admin,
  riskControl: {
    ...admin.riskControl,
    cyberPolicyExcludeBan: 'Exclude Cyber Policy Hits from Ban Count',
    cyberPolicyExcludeBanHint: 'When enabled, cyber_policy hits are excluded from the current and historical auto-ban count. Logs and notice emails are unaffected.',
    violationNotCounted: 'Not counted',
    action: {
      ...admin.riskControl.action,
      cyberPolicy: 'Cyber policy',
    },
  },
  settings: {
    ...admin.settings,
    features: {
      ...admin.settings.features,
      riskControl: {
        ...admin.settings.features.riskControl,
        cyberSessionBlock: 'Cyber session auto-block',
        cyberSessionBlockHint: 'Block only the exact API key and explicit session after an upstream cyber_policy hit; other sessions on the same key remain available.',
        cyberSessionBlockTTL: 'Block TTL (seconds)',
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
