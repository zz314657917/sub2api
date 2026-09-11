<template>
  <div class="service-store-page public-page-shell">
    <PublicTopNav />

    <main class="store-main">
      <aside class="store-sidebar" aria-label="商店信息">
        <div class="store-brand-mark" aria-hidden="true">✦</div>
        <span class="store-eyebrow">SERVICE STORE</span>
        <h1>数字服务商店</h1>
        <p>挑选适合你的 AI 服务与接入方案</p>
        <div class="store-sidebar-stat"><strong>{{ products.length }}</strong><span>项精选服务</span></div>
        <RouterLink class="store-sidebar-link" to="/tutorial">不知道怎么选？查看教程 <span>→</span></RouterLink>
      </aside>

      <section class="store-content">
        <div class="store-hero-banner">
          <div>
            <span class="store-hero-kicker">轻量接入 · 即买即用</span>
            <h2>把需要的服务，<em>一次配齐</em></h2>
            <p>从 Codex 接码到团队 API 套餐，透明价格，清晰权益。</p>
          </div>
          <div class="store-hero-orbit" aria-hidden="true"><span>✦</span><b>AI</b></div>
        </div>

        <div class="store-tabs" role="tablist" aria-label="服务分类">
          <button v-for="tab in tabs" :key="tab.id" type="button" role="tab" :aria-selected="selectedTab === tab.id" :class="{ active: selectedTab === tab.id }" @click="selectedTab = tab.id">
            {{ tab.label }} <small>{{ tab.count }}</small>
          </button>
        </div>

        <div class="store-toolbar">
          <div class="store-section-title"><span class="store-dot" /> <span>选择服务</span><small>为你的工作流找到合适的工具</small></div>
          <label class="store-search"><span aria-hidden="true">⌕</span><input v-model="search" type="search" placeholder="搜索服务名称" aria-label="搜索服务名称" /><button v-if="search" type="button" aria-label="清空搜索" @click="search = ''">×</button></label>
        </div>

        <div v-if="filteredProducts.length" class="product-grid">
          <article v-for="product in filteredProducts" :key="product.id" class="product-card">
            <div class="product-cover" :class="`tone-${product.tone}`"><span class="product-cover-spark">✦</span><strong>{{ product.icon }}</strong><small>{{ product.cover }}</small></div>
            <div class="product-body">
              <div class="product-category">{{ product.category }}</div>
              <h3>{{ product.title }}</h3>
              <p>{{ product.description }}</p>
              <div class="product-footer"><strong class="product-price">¥{{ product.price }}<small>{{ product.unit }}</small></strong><span class="stock" :class="{ low: product.stock === '库存紧张' }">{{ product.stock }}</span></div>
              <button type="button" class="product-buy" @click="selectedProduct = product">立即了解 <span>→</span></button>
            </div>
          </article>
        </div>
        <div v-else class="store-empty"><span>⌕</span><strong>没有找到相关服务</strong><p>换个关键词试试看</p></div>
      </section>
    </main>

    <div v-if="selectedProduct" class="store-dialog-backdrop" role="presentation" @click.self="selectedProduct = null">
      <section class="store-dialog" role="dialog" aria-modal="true" :aria-label="selectedProduct.title">
        <button class="dialog-close" type="button" aria-label="关闭" @click="selectedProduct = null">×</button>
        <div class="dialog-icon" :class="`tone-${selectedProduct.tone}`">{{ selectedProduct.icon }}</div>
        <span class="product-category">{{ selectedProduct.category }}</span>
        <h2>{{ selectedProduct.title }}</h2>
        <p>{{ selectedProduct.description }}</p>
        <div class="dialog-meta"><strong>¥{{ selectedProduct.price }}</strong><span>{{ selectedProduct.stock }}</span></div>
        <RouterLink class="dialog-action" to="/login">登录后继续 <span>→</span></RouterLink>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import PublicTopNav from './components/PublicTopNav.vue'

type Product = { id: string; category: string; title: string; description: string; price: string; unit: string; stock: string; icon: string; cover: string; tone: string; kind: string }

const selectedTab = ref('all')
const search = ref('')
const selectedProduct = ref<Product | null>(null)
const products: Product[] = [
  { id: 'codex-sms', category: '接码服务', title: 'Codex 接码 · 一次性', description: 'OpenAI Codex 注册验证码，自动接收，使用简单。', price: '1.54', unit: ' / 次', stock: '库存充足', icon: '✉', cover: 'CODEX SMS', tone: 'mint', kind: 'coding' },
  { id: 'codex-team', category: '账号服务', title: 'Codex 稳定渠道 · 1 天', description: '稳定可用的 Codex 使用渠道，适合短期项目。', price: '3.75', unit: ' / 份', stock: '库存一般', icon: '✦', cover: 'CODEX TEAM', tone: 'violet', kind: 'coding' },
  { id: 'api-basic', category: 'API 套餐', title: 'AI API 轻量包', description: '统一 API 入口，适合个人开发和快速试跑。', price: '9.90', unit: ' / 套', stock: '库存充足', icon: '⌁', cover: 'API ACCESS', tone: 'blue', kind: 'api' },
  { id: 'image-pack', category: '创作服务', title: '图像模型体验包', description: '图像生成额度，覆盖海报、头像与产品草图。', price: '12.80', unit: ' / 套', stock: '库存充足', icon: '◈', cover: 'IMAGE LAB', tone: 'peach', kind: 'creative' },
  { id: 'team-pro', category: '团队服务', title: '团队 API 专业版', description: '团队共享额度、用量可查，适合小型项目协作。', price: '49.00', unit: ' / 月', stock: '库存紧张', icon: '◉', cover: 'TEAM PRO', tone: 'dark', kind: 'team' },
  { id: 'setup-help', category: '技术服务', title: '接入配置协助', description: '手把手完成 Base URL、Key 与常用工具配置。', price: '19.90', unit: ' / 次', stock: '可预约', icon: '↗', cover: 'SETUP HELP', tone: 'yellow', kind: 'support' }
]
const tabs = computed(() => [
  { id: 'all', label: '全部服务', count: products.length },
  { id: 'coding', label: 'Coding', count: products.filter(p => p.kind === 'coding').length },
  { id: 'api', label: 'API 套餐', count: products.filter(p => p.kind === 'api').length },
  { id: 'creative', label: '创作工具', count: products.filter(p => p.kind === 'creative').length },
  { id: 'team', label: '团队协作', count: products.filter(p => p.kind === 'team').length }
])
const filteredProducts = computed(() => {
  const query = search.value.trim().toLowerCase()
  return products.filter(product => (selectedTab.value === 'all' || product.kind === selectedTab.value) && (!query || `${product.title} ${product.description} ${product.category}`.toLowerCase().includes(query)))
})
</script>

<style scoped>
@import './public-page.css';
.service-store-page { min-height: 100vh; background: #f7f9fc; color: #182230; }
.store-main { display: grid; grid-template-columns: 245px minmax(0, 1fr); gap: 0; width: min(100%, 1180px); margin: 0 auto; padding: 5rem 1.25rem 3.5rem; }
.store-sidebar { position: sticky; top: 5.25rem; align-self: start; min-height: 580px; padding: 1.1rem 1rem 2rem .2rem; }
.store-brand-mark { display: grid; place-items: center; width: 2.8rem; height: 2.8rem; margin-bottom: 1.7rem; border-radius: 12px; background: #182230; color: #ffe16a; font-size: 1.45rem; box-shadow: 0 8px 18px #18223022; }
.store-eyebrow,.store-hero-kicker { color: #718096; font-size: .68rem; font-weight: 700; letter-spacing: .14em; }
.store-sidebar h1 { margin-top: .7rem; font-size: 1.55rem; letter-spacing: -.04em; }
.store-sidebar p { margin-top: .6rem; color: #8994a4; font-size: .82rem; line-height: 1.65; }
.store-sidebar-stat { display: flex; align-items: baseline; gap: .45rem; margin-top: 2.2rem; padding-top: 1.25rem; border-top: 1px solid #e4eaf1; color: #8994a4; font-size: .75rem; }
.store-sidebar-stat strong { color: #182230; font-size: 1.5rem; }
.store-sidebar-link { display: flex; justify-content: space-between; margin-top: 2rem; padding: .8rem .85rem; border: 1px solid #e2e8f0; border-radius: 10px; color: #536174; font-size: .74rem; background: #fff; }
.store-sidebar-link:hover { border-color: #9bbcff; color: #2864d7; }
.store-content { min-width: 0; padding-left: 1rem; }
.store-hero-banner { position: relative; display: flex; justify-content: space-between; align-items: center; min-height: 178px; overflow: hidden; padding: 2rem 2.2rem; border-radius: 18px; background: linear-gradient(108deg,#dcecff,#f0f6ff 62%,#fff); }
.store-hero-banner:after { content: ''; position: absolute; width: 320px; height: 320px; right: -100px; top: -130px; border: 1px solid #fff; border-radius: 50%; box-shadow: 0 0 0 22px #ffffff45,0 0 0 44px #ffffff2e; }
.store-hero-banner h2 { position: relative; z-index: 1; margin-top: .6rem; font-size: clamp(1.55rem,3vw,2.35rem); letter-spacing: -.06em; }
.store-hero-banner h2 em { color: #2768d8; font-style: normal; }
.store-hero-banner p { position: relative; z-index: 1; margin-top: .65rem; color: #6c7b90; font-size: .84rem; }
.store-hero-orbit { position: relative; z-index: 1; display: grid; place-items: center; width: 92px; height: 92px; border: 1px solid #fff; border-radius: 28px; background: #ffffffa8; color: #2768d8; box-shadow: 0 16px 28px #86a9d833; transform: rotate(8deg); }
.store-hero-orbit b { font-size: 1.7rem; letter-spacing: -.12em; }.store-hero-orbit span { position: absolute; top: 12px; right: 14px; color: #f4bd39; }
.store-tabs { display: flex; gap: .45rem; overflow-x: auto; padding: 1.3rem 0 1rem; border-bottom: 1px solid #e2e8f0; }.store-tabs button { flex: 0 0 auto; border: 0; border-radius: 8px; padding: .62rem .8rem; color: #7d8998; background: transparent; font: inherit; font-size: .78rem; cursor: pointer; }.store-tabs button small { margin-left: .25rem; color: #a2adba; }.store-tabs button.active { color: #1e5dd0; background: #e9f1ff; font-weight: 650; }
.store-toolbar { display: flex; justify-content: space-between; align-items: center; gap: 1rem; padding: 1.25rem 0 .9rem; }.store-section-title { display: flex; align-items: center; gap: .45rem; font-size: .88rem; font-weight: 700; }.store-section-title small { margin-left: .35rem; color: #9aa5b2; font-size: .7rem; font-weight: 400; }.store-dot { width: .42rem; height: .42rem; border-radius: 50%; background: #3675e6; }.store-search { display: flex; align-items: center; gap: .4rem; width: 210px; padding: .5rem .7rem; border: 1px solid #e0e7f0; border-radius: 8px; background: #fff; color: #8492a5; }.store-search input { width: 100%; border: 0; outline: 0; color: #273549; background: transparent; font: inherit; font-size: .75rem; }.store-search button { border: 0; color: #98a3af; background: transparent; font-size: 1rem; cursor: pointer; }
.product-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: .95rem; }.product-card { overflow: hidden; border: 1px solid #e3e9f1; border-radius: 12px; background: #fff; box-shadow: 0 5px 16px #3a5b8310; transition: transform .18s ease,box-shadow .18s ease; }.product-card:hover { transform: translateY(-3px); box-shadow: 0 12px 24px #3a5b831c; }.product-cover { position: relative; display: flex; flex-direction: column; justify-content: center; min-height: 145px; overflow: hidden; padding: 1rem 1.15rem; }.product-cover:after { content: ''; position: absolute; width: 130px; height: 130px; right: -35px; top: -38px; border: 1px solid #ffffff66; border-radius: 50%; box-shadow: 0 0 0 15px #ffffff29,0 0 0 30px #ffffff1c; }.product-cover strong { position: relative; z-index: 1; font-size: 3.2rem; line-height: 1; color: #fff; text-shadow: 0 5px 14px #203b6b55; }.product-cover small { position: relative; z-index: 1; margin-top: .45rem; color: #ffffffd9; font-size: .67rem; font-weight: 700; letter-spacing: .12em; }.product-cover-spark { position: absolute; z-index: 1; right: 18px; bottom: 15px; color: #fff; }.tone-mint { background: linear-gradient(135deg,#24c6ad,#1178bd); }.tone-violet { background: linear-gradient(135deg,#8976e8,#3d4caa); }.tone-blue { background: linear-gradient(135deg,#6d9ef1,#3158cb); }.tone-peach { background: linear-gradient(135deg,#f5ad86,#c96882); }.tone-dark { background: linear-gradient(135deg,#324964,#142338); }.tone-yellow { background: linear-gradient(135deg,#f4c45d,#ef8a58); }
.product-body { padding: .9rem 1rem 1rem; }.product-category { color: #8d9aab; font-size: .67rem; }.product-body h3 { overflow: hidden; margin-top: .35rem; color: #223044; font-size: .9rem; text-overflow: ellipsis; white-space: nowrap; }.product-body p { display: -webkit-box; overflow: hidden; min-height: 2.4em; margin-top: .45rem; color: #8995a5; font-size: .73rem; line-height: 1.6; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }.product-footer { display: flex; justify-content: space-between; align-items: center; margin-top: .8rem; }.product-price { color: #2164d6; font-size: 1.08rem; }.product-price small { margin-left: .15rem; color: #9aa6b5; font-size: .64rem; font-weight: 400; }.stock { padding: .24rem .4rem; border-radius: 4px; color: #3b9b75; background: #e8f8f1; font-size: .62rem; }.stock.low { color: #dc765f; background: #fff0eb; }.product-buy { display: flex; justify-content: space-between; width: 100%; margin-top: .75rem; padding: .58rem .7rem; border: 1px solid #dce7f8; border-radius: 7px; color: #2864cf; background: #f4f8ff; font: inherit; font-size: .72rem; cursor: pointer; }.product-buy:hover { background: #e8f1ff; }
.store-empty { padding: 4rem 1rem; text-align: center; color: #8b98a8; }.store-empty span { display: block; font-size: 2rem; }.store-empty strong { display: block; margin-top: .6rem; color: #405066; }.store-empty p { margin-top: .35rem; font-size: .78rem; }.store-dialog-backdrop { position: fixed; inset: 0; z-index: 100; display: grid; place-items: center; padding: 1rem; background: #18223055; backdrop-filter: blur(5px); }.store-dialog { position: relative; width: min(100%,380px); padding: 1.5rem; border-radius: 16px; background: #fff; box-shadow: 0 24px 60px #18223033; }.dialog-close { position: absolute; top: .7rem; right: .85rem; border: 0; color: #9ba6b4; background: transparent; font-size: 1.5rem; cursor: pointer; }.dialog-icon { display: grid; place-items: center; width: 3.4rem; height: 3.4rem; margin-bottom: .9rem; border-radius: 12px; color: #fff; font-size: 1.8rem; }.store-dialog h2 { margin-top: .35rem; font-size: 1.25rem; }.store-dialog p { margin-top: .6rem; color: #78869a; font-size: .82rem; line-height: 1.7; }.dialog-meta { display: flex; justify-content: space-between; align-items: center; margin-top: 1.2rem; }.dialog-meta strong { color: #2164d6; font-size: 1.35rem; }.dialog-meta span { color: #3b9b75; font-size: .72rem; }.dialog-action { display: flex; justify-content: space-between; margin-top: 1.1rem; padding: .75rem .9rem; border-radius: 8px; color: #fff; background: #2864d6; font-size: .8rem; }
@media (max-width: 900px) { .store-main { grid-template-columns: 1fr; padding-top: 5.2rem; }.store-sidebar { position: static; min-height: 0; padding: 0 0 1rem; }.store-brand-mark { display: none; }.store-sidebar-stat,.store-sidebar-link { display: none; }.store-content { padding-left: 0; }.product-grid { grid-template-columns: repeat(2,minmax(0,1fr)); } }
@media (max-width: 560px) { .store-main { padding: 4.9rem .8rem 2.5rem; }.store-hero-banner { min-height: 158px; padding: 1.45rem; }.store-hero-orbit { width: 64px; height: 64px; border-radius: 20px; }.store-hero-orbit b { font-size: 1.25rem; }.store-hero-banner p { max-width: 15rem; font-size: .74rem; }.store-toolbar { align-items: flex-start; flex-direction: column; }.store-section-title { flex-wrap: wrap; }.store-section-title small { width: 100%; margin-left: .9rem; }.store-search { width: 100%; }.product-grid { grid-template-columns: 1fr 1fr; gap: .65rem; }.product-cover { min-height: 116px; }.product-cover strong { font-size: 2.4rem; }.product-body { padding: .75rem; }.product-body h3 { font-size: .78rem; }.product-body p { font-size: .67rem; }.product-price { font-size: .95rem; }.product-buy { font-size: .68rem; } }
</style>
