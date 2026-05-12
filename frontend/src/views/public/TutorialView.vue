<template>
  <div class="public-page-shell min-h-screen text-white">
    <PublicMatrixBackdrop />

    <PublicTopNav />

    <main class="relative z-10 mx-auto w-full max-w-6xl px-4 py-10 sm:px-6 lg:py-14">
      <section class="public-hero">
        <div class="public-kicker">
          <span></span>
          QUICK START
        </div>
        <h1>落叶网络接入教程</h1>
        <p>
          从注册、领取试用额度，到把密钥接入 ChatGPT、Codex、Claude Code 或你自己的 OpenAI SDK，
          这份教程按实际使用顺序整理，照着做即可开始调用。
        </p>
      </section>

      <section id="quick-start" class="tutorial-layout">
        <aside class="tutorial-index">
          <a v-for="item in sections" :key="item.id" :href="`#${item.id}`">
            {{ item.title }}
          </a>
        </aside>

        <div class="tutorial-content">
          <article id="account" class="tutorial-card">
            <PixelIcon name="user" size="md" />
            <div>
              <p class="tutorial-step">01</p>
              <h2>创建账户并领取试用额度</h2>
              <p>点击首页“注册领取试用”，完成注册后联系右下角客服，说明你的使用场景，即可领取试用额度。</p>
              <ul>
                <li>建议使用常用邮箱注册，便于找回账户和接收通知。</li>
                <li>试用额度用于验证模型可用性、延迟和工具接入流程。</li>
              </ul>
            </div>
          </article>

          <article id="key" class="tutorial-card">
            <PixelIcon name="key" size="md" />
            <div>
              <p class="tutorial-step">02</p>
              <h2>创建并复制 API 密钥</h2>
              <p>进入控制台的 API 密钥页面，新建密钥后复制保存。密钥只展示一次，建议放入本机环境变量或你的密钥管理工具。</p>
              <pre><code>OPENAI_API_KEY=你的落叶网络密钥
OPENAI_BASE_URL=https://你的站点地址/v1</code></pre>
            </div>
          </article>

          <article id="codex" class="tutorial-card">
            <PixelIcon name="cursor" size="md" />
            <div>
              <p class="tutorial-step">03</p>
              <h2>接入 Codex / OpenAI SDK</h2>
              <p>把客户端的 base URL 指向落叶网络的兼容接口，模型名按模型广场展示填写即可。</p>
              <pre><code>import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY,
  baseURL: process.env.OPENAI_BASE_URL
})

const response = await client.chat.completions.create({
  model: 'gpt-5.4',
  messages: [{ role: 'user', content: '你好，帮我写一个接口示例' }]
})</code></pre>
            </div>
          </article>

          <article id="claude" class="tutorial-card">
            <PixelIcon name="spark" size="md" />
            <div>
              <p class="tutorial-step">04</p>
              <h2>Claude Code / 兼容工具配置</h2>
              <p>如果你的工具支持 OpenAI 兼容接口，优先选择 OpenAI Compatible Provider，然后填入落叶网络密钥和接口地址。</p>
              <ul>
                <li>模型选择：日常编码建议先用快速模型，复杂重构再切换高阶模型。</li>
                <li>上下文较长时，注意控制历史消息和文件范围，避免无意义消耗。</li>
              </ul>
            </div>
          </article>

          <article id="usage" class="tutorial-card">
            <PixelIcon name="usage" size="md" />
            <div>
              <p class="tutorial-step">05</p>
              <h2>查看用量和排查问题</h2>
              <p>控制台可以查看请求记录、额度消耗和模型调用情况。遇到失败请求时，先确认密钥、模型名、base URL 和账户额度。</p>
              <ul>
                <li>401：检查密钥是否复制完整。</li>
                <li>404：检查模型名是否在模型广场中可用。</li>
                <li>429 或额度不足：降低并发或联系管理员补充额度。</li>
              </ul>
            </div>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import PixelIcon from '@/components/icons/PixelIcon.vue'
import PublicMatrixBackdrop from './components/PublicMatrixBackdrop.vue'
import PublicTopNav from './components/PublicTopNav.vue'

const sections = [
  { id: 'account', title: '创建账户' },
  { id: 'key', title: 'API 密钥' },
  { id: 'codex', title: 'Codex 接入' },
  { id: 'claude', title: '工具配置' },
  { id: 'usage', title: '用量排查' },
]
</script>

<style scoped>
@import './public-page.css';

.tutorial-layout {
  display: grid;
  grid-template-columns: 12rem minmax(0, 1fr);
  gap: 1rem;
}

.tutorial-index {
  position: sticky;
  top: 5rem;
  align-self: start;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.065);
  padding: 0.65rem;
  backdrop-filter: blur(18px);
}

.tutorial-index a {
  display: block;
  padding: 0.65rem 0.75rem;
  color: rgba(235, 245, 239, 0.72);
  font-size: 0.82rem;
  font-weight: 800;
}

.tutorial-index a:hover {
  background: rgba(255, 255, 255, 0.08);
  color: white;
}

.tutorial-content {
  display: grid;
  gap: 1rem;
}

.tutorial-card {
  display: grid;
  grid-template-columns: 2.7rem minmax(0, 1fr);
  gap: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.075);
  padding: 1.2rem;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1), 0 18px 44px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(18px);
}

.tutorial-card h2 {
  margin-top: 0.1rem;
  font-size: 1.2rem;
  font-weight: 900;
}

.tutorial-card p,
.tutorial-card li {
  color: rgba(238, 246, 240, 0.74);
  line-height: 1.8;
}

.tutorial-card ul {
  margin-top: 0.75rem;
  padding-left: 1.2rem;
}

.tutorial-step {
  color: #77ffad !important;
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.08em;
}

.tutorial-card pre {
  margin-top: 0.9rem;
  overflow-x: auto;
  border: 1px solid rgba(120, 255, 170, 0.16);
  background: rgba(2, 7, 9, 0.66);
  padding: 0.9rem;
  color: rgba(225, 255, 233, 0.86);
  font-size: 0.78rem;
  line-height: 1.7;
}

.tutorial-card .pixel-glyph {
  margin-top: 0.2rem;
}

@media (max-width: 900px) {
  .tutorial-layout {
    grid-template-columns: 1fr;
  }

  .tutorial-index {
    position: static;
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }
}
</style>
