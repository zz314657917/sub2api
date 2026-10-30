import { createApp, h } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory, RouterView } from 'vue-router'
import PelicanTestsView from '@/views/user/PelicanTestsView.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { apiClient } from '@/api/client'
import i18n, { initI18n, setLocale } from '@/i18n'
import '@/style.css'
import '@/styles/pixel-ui.css'
import '@/styles/console-ui.css'
import radarFixture from './radar-fixture.json'

// Preview entry only. The real app entry, router and API implementation are untouched.
const demoGroups = [
  { id: 1, name: '官方 GPT · 稳定组', platform: 'openai', status: 'active' },
  { id: 2, name: 'Claude Pro · 优选组', platform: 'anthropic', status: 'active' },
  { id: 3, name: 'Gemini · 日常组', platform: 'gemini', status: 'active' }
]
const demoPlans = [
  { id: 101, group_id: 1, group_name: demoGroups[0].name, model_id: 'gpt-6.1-sol', model_ids: ['gpt-6.1-sol', 'gpt-6-astra'] },
  { id: 102, group_id: 2, group_name: demoGroups[1].name, model_id: 'claude-opus-4-6', model_ids: ['claude-opus-4-6'] }
].map(plan => ({
  ...plan, preferred_account_ids: [], enabled: false, interval_minutes: 60, timeout_seconds: 180,
  max_results: 5, min_chars: 9366, daily_call_limit: 0, failure_pause_threshold: 0, retention_days: 0,
  reasoning_effort: '', daily_calls_used: 0, usage_day: '', consecutive_failed_runs: 0, pause_reason: '',
  estimated_calls_per_run: plan.model_ids.length, last_run_calls: 0,
  created_at: '2026-10-30T06:00:00Z', updated_at: '2026-10-30T06:00:00Z'
}))
const settings = {
  site_name: 'Sub2API',
  site_logo: '',
  model_plaza_enabled: true,
  payment_enabled: false,
  group_buy_enabled: false,
  pixel_cafe_enabled: false,
  welfare_enabled: false,
  custom_menu_items: []
}
const demoSettings = { display_name: '鹈鹕测试', enabled: true, prompt: '生成完整的鹈鹕骑行动态 SVG 作品。', radar: radarFixture.radar }
const previewRequests: { method: string; url: string }[] = []
;(window as any).__PELICAN_PREVIEW_REQUESTS__ = previewRequests
apiClient.defaults.adapter = async config => {
  const url = config.url || ''
  previewRequests.push({ method: config.method || 'get', url })
  if (config.method !== 'get') throw new Error('当前是布局预览，不能保存设置或运行真实测试。')
  let data: unknown = {}
  if (url === '/pelican-tests/reference-scores') {
    const scenario = new URLSearchParams(location.search).get('radar')
    const payload = structuredClone(radarFixture.payload)
    if (scenario === 'off') { payload.enabled = false; payload.status = 'disabled'; payload.items = [] }
    if (scenario === 'stale') payload.status = 'stale'
    if (scenario === 'unavailable') {
      Object.assign(payload, { status: 'unavailable', synced_at: null, test_started_at: null, test_completed_at: null, source_snapshot_at: null, round_id: null, score_max: null })
      payload.items = payload.items.map(item => ({ ...item, status: 'unavailable', score: null })) as any
    }
    if (scenario === 'missing') payload.items[0] = { model: 'unknown-model', display_name: 'unknown-model', effort: 'high', status: 'not_recorded', score: null } as any
    if (scenario === 'unsupported') payload.items[0] = { model: 'gpt-6-luna', display_name: 'GPT-6 Luna', effort: 'ultra', status: 'unsupported', score: null } as any
    data = payload
  }
  else if (url === '/admin/pelican-reference-catalog') data = radarFixture.catalog
  else if (url === '/pelican-tests/metadata' || url === '/admin/pelican-test-settings') data = demoSettings
  else if (url === '/pelican-tests') data = { items: [], groups: demoGroups, total: 3, page: 1, page_size: 12 }
  else if (url === '/admin/pelican-test-plans') data = new URLSearchParams(location.search).get('plans') === 'empty' ? [] : demoPlans
  else if (url === '/admin/pelican-test-plans/models') data = { models: [...new Set(demoPlans.filter(plan => plan.group_id === Number(config.params?.group_id)).flatMap(plan => plan.model_ids))] }
  else if (url === '/admin/pelican-test-plans/accounts') data = { accounts: [] }
  else if (url.includes('/groups')) data = url.endsWith('/all') ? demoGroups : { items: demoGroups, total: 3 }
  else if (url.includes('/subscriptions') || url === '/announcements') data = []
  else if (url === '/settings/public' || url === '/admin/settings') data = settings
  else if (url.includes('unread')) data = { unread_count: 0, user_unread_count: 0, admin_unread_count: 0, total: 0 }
  else if (url.includes('/payment')) data = { enabled: false }
  return { data: { code: 0, data }, status: 200, statusText: 'OK', headers: {}, config }
}

const app = createApp({ render: () => h(RouterView) })
app.use(createPinia())
const auth = useAuthStore()
const previewRole = new URLSearchParams(location.search).get('role') === 'user' ? 'user' : 'admin'
auth.$patch({
  token: 'local-layout-preview',
  user: { id: 999999, username: '预览账号', email: 'preview@example.invalid', role: previewRole, balance: 0 } as any
})
localStorage.setItem(`${previewRole === 'admin' ? 'admin_guide' : 'user_guide'}_999999_${previewRole}_v4_interactive`, 'true')
const appStore = useAppStore()
;(window as any).__APP_CONFIG__ = settings
appStore.initFromInjectedConfig()
document.documentElement.classList.toggle('dark', localStorage.getItem('pelican-preview-theme') === 'dark')
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/pelican-tests', name: 'PelicanTests', component: PelicanTestsView, meta: { title: '鹈鹕测试' } },
    { path: '/:pathMatch(.*)*', redirect: '/pelican-tests' }
  ]
})
router.beforeEach(to => to.path === '/pelican-tests' || false)
app.use(router)
app.use(i18n)
await initI18n()
await setLocale('zh')
await router.isReady()
app.mount('#app')
