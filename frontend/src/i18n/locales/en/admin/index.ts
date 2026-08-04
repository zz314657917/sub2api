import dashboard from './dashboard'
import backup from './backup'
import dataManagement from './dataManagement'
import affiliates from './affiliates'
import users from './users'
import groups from './groups'
import availableChannels from './availableChannels'
import channels from './channels'
import riskControl from './riskControl'
import channelMonitor from './channelMonitor'
import subscriptions from './subscriptions'
import accounts from './accounts'
import sharedAccounts from './sharedAccounts'
import scheduledTests from './scheduledTests'
import proxies from './proxies'
import redeem from './redeem'
import announcements from './announcements'
import tutorials from './tutorials'
import promo from './promo'
import usage from './usage'
import ops from './ops'
import settings from './settings'
import errorPassthrough from './errorPassthrough'
import tlsFingerprintProfiles from './tlsFingerprintProfiles'
import tickets from './tickets'
import audit from './audit'
import promptAudit from './promptAudit'
import pixelCafe from './pixelCafe'

export default {
  dashboard,
  backup,
  dataManagement,
  affiliates,
  users,
  groups,
  availableChannels,
  channels,
  riskControl,
  channelMonitor,
  subscriptions,
  accounts,
  sharedAccounts,
  scheduledTests,
  proxies,
  redeem,
  announcements,
  tutorials,
  promo,
  usage,
  ops,
  settings,
  errorPassthrough,
  tlsFingerprintProfiles,
  tickets,
  audit,
  ...promptAudit,
  pixelCafe,
}
