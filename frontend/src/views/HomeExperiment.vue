<template>
  <div ref="homeRoot" class="usa-home" :data-theme="homeTheme">
    <header class="site-header" aria-label="主导航" @keydown.esc="closeMobileNav(true)">
      <div class="nav-shell">
        <a class="brand" href="#top" :aria-label="`${brandName} 首页`">
          <span class="brand-mark" aria-hidden="true"><img src="/home-experiment/logo.png" alt="" /></span>
          <span class="brand-name">{{ brandName }}</span>
        </a>
        <nav
          id="home-navigation"
          ref="mobileNavRef"
          class="nav-links"
          :class="{ 'is-open': mobileNavOpen }"
          aria-label="页面导航"
          @click="handleMobileNavNavigation"
        >
          <a href="#overview">功能总览</a>
          <a href="#providers">模型能力</a>
          <a href="#routes">接入端点</a>
          <a href="#pricing">价格估算</a>
          <RouterLink to="/key-usage">Key 用量</RouterLink>
          <RouterLink to="/download">客户端下载</RouterLink>
        </nav>
        <div class="nav-actions">
          <button
            ref="themeToggleRef"
            class="icon-button"
            type="button"
            @click="toggleHomeTheme"
            :aria-label="homeTheme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
            :title="homeTheme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
          >
            <svg class="icon sun" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 18.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13Zm0-2a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9ZM11 1h2v3h-2V1Zm0 19h2v3h-2v-3ZM1 11h3v2H1v-2Zm19 0h3v2h-3v-2ZM4.22 2.81l2.12 2.12-1.41 1.41-2.12-2.12 1.41-1.41Zm14.85 14.85 2.12 2.12-1.41 1.41-2.12-2.12 1.41-1.41Zm.71-14.85 1.41 1.41-2.12 2.12-1.41-1.41 2.12-2.12ZM4.93 17.66l1.41 1.41-2.12 2.12-1.41-1.41 2.12-2.12Z" /></svg>
            <svg class="icon moon" viewBox="0 0 24 24" aria-hidden="true"><path d="M21 14.69A8.5 8.5 0 0 1 9.31 3 7.1 7.1 0 1 0 21 14.69ZM12.36 1.25a10.5 10.5 0 1 0 10.39 10.39 1 1 0 0 0-1.63-.77A6.5 6.5 0 0 1 13.13 2.88a1 1 0 0 0-.77-1.63Z" /></svg>
          </button>
          <RouterLink class="primary-link" :to="entryPath">{{ entryLabel }}</RouterLink>
          <button
            ref="mobileNavToggleRef"
            class="icon-button mobile-nav-toggle"
            type="button"
            aria-controls="home-navigation"
            :aria-expanded="mobileNavOpen"
            :aria-label="mobileNavOpen ? '关闭页面导航' : '打开页面导航'"
            :title="mobileNavOpen ? '关闭导航' : '打开导航'"
            @click="toggleMobileNav"
          >
            <Icon :name="mobileNavOpen ? 'x' : 'menu'" size="md" aria-hidden="true" />
          </button>
        </div>
      </div>
      <button
        v-if="mobileNavOpen"
        class="mobile-nav-backdrop"
        type="button"
        tabindex="-1"
        aria-label="关闭页面导航"
        @click="closeMobileNav(true)"
      ></button>
    </header>

    <main id="top">
      <section class="hero" aria-labelledby="hero-title">
        <div class="hero-grid" aria-hidden="true"></div>
        <canvas class="code-flow" id="codeFlow" aria-hidden="true"></canvas>
        <div class="hero-content">
          <p class="eyebrow">{{ heroEyebrowText }}</p>
          <h1 id="hero-title">{{ brandName }}</h1>
          <p class="hero-copy">{{ subtitle }}</p>
          <div class="hero-actions" aria-label="主要操作">
            <RouterLink class="button primary" :to="entryPath">
              <svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h12.17l-5.58-5.59L13 5l8 8-8 8-1.41-1.41L17.17 14H5v-2Z" /></svg>
              {{ entryButtonLabel }}</RouterLink>
            <a class="button secondary" href="#routes">
              <svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 4h10a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Zm0 2v12h10V6H7Zm2 3h6v2H9V9Zm0 4h6v2H9v-2Z" /></svg>
              查看规范
            </a>
          </div>
        </div>

        <section class="console-shell" aria-label="USA-零 API 网关控制台预览">
          <div class="console-topbar">
            <div class="window-dots" aria-hidden="true"><span></span><span></span><span></span></div>
            <div class="route-pill">POST /v1/chat/completions</div>
            <div class="status-pill"><span></span> Live</div>
          </div>
          <div class="gateway-board">
            <div class="request-panel">
              <div class="panel-head">
                <span>标准请求</span>
                <button class="copy-button" type="button" :data-copy="`curl -X POST ${baseUrl}/chat/completions`" aria-label="复制请求示例" title="复制请求">
                  <svg class="icon" viewBox="0 0 24 24" aria-hidden="true"><path d="M8 7a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3h-1v-2h1a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1h-7a1 1 0 0 0-1 1v1H8V7Zm-5 4a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3H6a3 3 0 0 1-3-3v-7Zm3-1a1 1 0 0 0-1 1v7a1 1 0 0 0 1 1h7a1 1 0 0 0 1-1v-7a1 1 0 0 0-1-1H6Z" /></svg>
                </button>
              </div>
              <pre><code><span class="muted">curl</span> {{ baseUrl }}/chat/completions \
  -H <span class="string">"Authorization: Bearer sk-usa0"</span> \
  -H <span class="string">"Content-Type: application/json"</span> \
  -d <span class="string">'{
    "model": "auto:reasoning",
    "messages": [{"role":"user","content":"生成部署计划"}],
    "stream": true
  }'</span></code></pre>
            </div>
            <div class="flow-panel">
              <div class="provider-row">
                <span class="provider-chip openai">OpenAI</span>
                <span class="line"><i></i></span>
                <span class="provider-chip claude">Claude</span>
                <span class="line"><i></i></span>
                <span class="provider-chip gemini">Gemini</span>
              </div>
              <div class="gateway-core">
                <strong>USA-零 智能路由</strong>
              </div>
              <div class="metric-grid">
                <div><span class="metric-value digit-ticker">{{ availabilityLabel }}</span><small>可用性</small></div>
                <div><span class="metric-value digit-ticker">{{ routeLatencyLabel }}</span><small>路由延迟</small></div>
                <div><span class="metric-value digit-ticker">{{ todayTokensLabel }}</span><small>今日 Tokens</small></div>
              </div>
            </div>
          </div>
        </section>
        <div class="brand-reveal" aria-hidden="true">{{ brandName }}</div>
      </section>

      <section id="overview" class="section band-light">
        <div class="section-inner">
          <div class="section-title">
            <h2>把不同模型供应商整理成同一种调用方式</h2>
          </div>
          <div class="feature-grid">
            <article class="feature-card">
              <span class="feature-icon blue"><svg viewBox="0 0 24 24"><path d="M4 7a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v10a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3V7Zm3-1a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1H7Zm2 4h6v2H9v-2Zm0 4h4v2H9v-2Z" /></svg></span>
              <h3>统一协议</h3>
              <p>用 OpenAI 兼容格式接入多家模型，减少 SDK、鉴权和响应结构差异。</p>
            </article>
            <article class="feature-card">
              <span class="feature-icon green"><svg viewBox="0 0 24 24"><path d="M12 2 3 7v10l9 5 9-5V7l-9-5Zm0 2.29 5.77 3.2L12 10.69 6.23 7.49 12 4.29ZM5 9.18l6 3.33v6.7l-6-3.33v-6.7Zm8 10.03v-6.7l6-3.33v6.7l-6 3.33Z" /></svg></span>
              <h3>智能路由</h3>
              <p>按能力、延迟、成本和余额选择上游，失败时自动切换备用通道。</p>
            </article>
            <article class="feature-card">
              <span class="feature-icon amber"><svg viewBox="0 0 24 24"><path d="M12 2a10 10 0 1 0 10 10h-2a8 8 0 1 1-8-8V2Zm1 1v9h8v-1a8 8 0 0 0-8-8Zm2 3.34A6.02 6.02 0 0 1 18.66 10H15V6.34Z" /></svg></span>
              <h3>用量计量</h3>
              <p>聚合 token、请求、错误和余额数据，让团队能看懂每一次消耗。</p>
            </article>
            <article class="feature-card">
              <span class="feature-icon rose"><svg viewBox="0 0 24 24"><path d="M12 2 4 5v6c0 5.55 3.84 10.74 8 12 4.16-1.26 8-6.45 8-12V5l-8-3Zm0 2.13L18 6.4V11c0 4.36-2.75 8.47-6 9.86C8.75 19.47 6 15.36 6 11V6.4l6-2.27Zm3.54 5.33L11 14l-2.04-2.04-1.42 1.42L11 16.83l5.96-5.95-1.42-1.42Z" /></svg></span>
              <h3>密钥治理</h3>
              <p>按项目、用户、额度和模型组发放 Key，便于隔离测试和生产流量。</p>
            </article>
          </div>
        </div>
        <div class="section-inner model-pricing-inner">
          <section v-for="group in modelPricingGroups" :key="group.key" class="model-group">
            <h3 :id="`pricing-${group.key}-title`" class="model-group-title">
              <span class="provider-symbol" :class="group.symbolClass" aria-hidden="true">{{ group.symbol }}</span>
              <strong>{{ group.name }}</strong>
              <span>· {{ group.models.length }} 个模型</span>
            </h3>
            <div class="pricing-table-wrap">
              <table class="pricing-table" :aria-labelledby="`pricing-${group.key}-title`">
                <thead>
                  <tr>
                    <th scope="col">模型</th>
                    <th v-for="label in priceColumnLabels" :key="label" scope="col">{{ label }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="model in group.models" :key="model.id">
                    <th scope="row" class="model-cell">
                      <strong>
                        <span>{{ model.id }}</span>
                        <em v-if="model.badge">{{ model.badge }}</em>
                        <button class="copy-id" type="button" :data-copy="model.id"></button>
                      </strong>
                      <small>{{ model.description }}</small>
                    </th>
                    <td v-for="(price, index) in model.prices" :key="index" :data-label="priceColumnLabels[index]">
                      <span>{{ price.value }}</span>
                      <small v-if="price.note" class="price-note">{{ price.note }}</small>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>
      </section>

      <section id="providers" class="section">
        <div class="section-inner api-capabilities-inner">
          <div class="api-capabilities-heading">
            <p class="eyebrow">CORE COMPETENCIES</p>
            <h2>专注模型 API 接入的核心能力</h2>
            <p>一个 API Key，改一下 Base URL，就能接入常用大模型能力。按量计费、调用记录可查，适合个人开发、团队工具和服务端项目快速上线。</p>
          </div>
          <div class="api-capabilities-grid">
            <div class="capability-list" aria-label="模型 API 接入能力">
              <article class="capability-item">
                <span class="capability-index">01</span>
                <div>
                  <h3>官方同源，稳定直连</h3>
                  <p>面向高频调用场景优化连接稳定性，减少 429、超时和长连接中断带来的接入成本。</p>
                </div>
              </article>
              <article class="capability-item">
                <span class="capability-index">02</span>
                <div>
                  <h3>兼容 OpenAI 接口</h3>
                  <p>保留熟悉的请求格式和鉴权方式，常见 SDK、IDE 插件和服务端集成通常只需要替换基础地址。</p>
                </div>
              </article>
              <article class="capability-item">
                <span class="capability-index">03</span>
                <div>
                  <h3>GPT / Claude / Gemini 模型可用</h3>
                  <p>集中查看可用模型与适用场景，按任务选择对话、代码、长上下文或轻量模型。</p>
                </div>
              </article>
              <article class="capability-item">
                <span class="capability-index">04</span>
                <div>
                  <h3>余额与调用记录透明</h3>
                  <p>按量消费、余额可查，调用明细记录请求状态、模型和用量，方便排查与对账。</p>
                </div>
              </article>
            </div>
            <aside class="api-access-panel" aria-label="API 接入参数示例">
              <div class="access-panel-head">
                <span>接入参数</span>
                <strong>OpenAI 兼容</strong>
              </div>
              <div class="access-credentials" aria-label="基础接入信息">
                <div>
                  <span>Base URL</span>
                  <code>{{ baseUrl }}</code>
                </div>
                <div>
                  <span>Authorization</span>
                  <code>Bearer sk-usa0...</code>
                </div>
              </div>
              <div class="access-status-grid" aria-label="平台状态摘要">
                <div><i class="status-dot green"></i><strong>接口可用</strong><span>稳定直连</span></div>
                <div><i class="status-dot amber"></i><strong>按量计费</strong><span>余额消费</span></div>
                <div><i class="status-dot blue"></i><strong>用量可查</strong><span>明细记录</span></div>
              </div>
              <div class="access-terminal" aria-label="终端调用示例">
                <div class="terminal-bar" aria-hidden="true"><span></span><span></span><span></span></div>
                <pre class="access-code"><code><span class="terminal-prompt">$</span> curl {{ baseUrl }}/chat/completions \
  -H "Authorization: Bearer sk-usa0..." \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.3-codex","messages":[{"role":"user","content":"生成接口测试用例"}]}'</code></pre>
              </div>
            </aside>
          </div>
        </div>
      </section>

      <section id="routes" class="section band-muted">
        <div class="section-inner">
          <div class="section-title">
            <p class="eyebrow">Standard routes</p>
            <h2>常用端点保持兼容，内部路由保持弹性</h2>
          </div>
          <div class="tabs" role="tablist" aria-label="请求示例" @keydown="handleTablistKeydown">
            <button
              v-for="tab in routeTabs"
              :id="`route-tab-${tab.key}`"
              :key="tab.key"
              class="tab"
              :class="{ active: activeTab === tab.key }"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.key"
              :aria-controls="`route-panel-${tab.key}`"
              :tabindex="activeTab === tab.key ? 0 : -1"
              @click="activeTab = tab.key"
            >
              {{ tab.label }}
            </button>
          </div>
          <div class="route-demo">
            <div class="route-list" aria-label="标准端点">
              <div><span>POST</span><strong>/v1/chat/completions</strong><small>OpenAI-compatible</small></div>
              <div><span>POST</span><strong>/v1/responses</strong><small>multi-modal</small></div>
              <div><span>GET</span><strong>/v1/models</strong><small>normalized catalog</small></div>
              <div><span>POST</span><strong>/v1/images/generations</strong><small>provider mapped</small></div>
            </div>
            <pre
              :id="`route-panel-${activeTab}`"
              class="code-window"
              role="tabpanel"
              :aria-labelledby="`route-tab-${activeTab}`"
              tabindex="0"
            ><code>{{ routeExamples[activeTab] }}</code></pre>
          </div>
        </div>
      </section>

      <section id="pricing" class="section">
        <div class="section-inner pricing-layout">
          <div class="section-title align-left">
            <p class="eyebrow">Usage control</p>
            <h2>按团队节奏分配额度</h2>
            <p>适合开发者、小团队和需要统一管理多模型账户的服务端项目。</p>
          </div>
          <div class="pricing-card">
            <div class="pricing-head">
              <span>用量估算</span>
              <strong>按量计费</strong>
            </div>
            <label class="range-label" for="tokenRange">每月 Token 使用量 <span id="usageTokenValue">40 百万</span></label>
            <input id="tokenRange" type="range" min="1" max="400" value="40" step="1" />
            <div class="estimate token-estimate" aria-label="Token 额度估算">
              <div><span id="inputCostValue">约 ¥10</span><small>预估输入 Token 费用</small></div>
            </div>
            <p class="pricing-promo">百万输入 Token 低至 ¥0.25，适合高频调用与团队统一管理。</p>
            <ul class="check-list">
              <li>先估算每月 Token 使用量，再按输入单价折算预算</li>
              <li>人民币与美元额度 1:1 折算，¥1 等于 $1 可用余额</li>
              <li>Token 单位按万、百万、亿自动切换展示</li>
            </ul>
          </div>
        </div>
      </section>

      <section id="docs" class="section band-cta">
        <canvas class="dot-matrix" id="ctaDotMatrix" aria-hidden="true"></canvas>
        <div class="section-inner cta-layout">
          <div class="cta-copy">
            <div class="cta-brand">
              <img src="/home-experiment/logo.png" alt="" />
              <span>{{ brandName }}</span></div><h2>把应用里的模型供应商切换成 {{ brandName }}</h2>
            <p>统一 OpenAI、Claude、Gemini 等接口，按额度或路由稳定接入。</p>
          </div>
          <div class="endpoint-card">
            <span>Base URL</span>
            <code>{{ baseUrl }}</code>
            <button class="copy-button text" type="button" :data-copy="baseUrl">复制</button>
          </div>
        </div>
      </section>
    </main>

    <footer class="site-footer">
      <span>{{ brandName }}</span><span>Unified Service API ZERO（零号统一智能服务 API）</span>
    </footer>
    <div class="copy-toast" :class="{ 'is-visible': copyToastVisible }" role="status" aria-live="polite">
      {{ copyToastMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { getHomeMetrics } from '@/api/home'
import Icon from '@/components/icons/Icon.vue'
import { useTheme } from '@/composables/useTheme'

const props = defineProps<{
  siteName: string
  siteSubtitle: string
  isAuthenticated: boolean
  dashboardPath: string
}>()

const brandName = computed(() => props.siteName || 'USA-零')
const heroEyebrowText = 'Unified Service API'
const subtitle = computed(() => props.siteSubtitle || '统一 OpenAI、Claude、Gemini 等不同接口，把多模型调用规范成一个稳定、可计量、可治理的标准 API。')
const entryPath = computed(() => props.isAuthenticated ? props.dashboardPath : '/login')
const entryLabel = computed(() => props.isAuthenticated ? '进入控制台' : '开始接入')
const entryButtonLabel = computed(() => props.isAuthenticated ? '进入控制台' : '获取 API Key')
const baseUrl = computed(() => `${window.location.origin}/v1`)
const { resolvedTheme: homeTheme, toggleTheme } = useTheme()

interface ModelPrice {
  value: string
  note?: string
}

interface ModelPricingItem {
  id: string
  badge?: string
  description: string
  prices: [ModelPrice, ModelPrice, ModelPrice]
}

interface ModelPricingGroup {
  key: string
  name: string
  symbol: string
  symbolClass: string
  models: ModelPricingItem[]
}

const priceColumnLabels = ['输入 / 百万', '输出 / 百万', '缓存 / 百万']
const modelPricingGroups: ModelPricingGroup[] = [
  {
    key: 'openai',
    name: 'OPENAI',
    symbol: '◎',
    symbolClass: 'openai-symbol',
    models: [
      { id: 'gpt-5.6', badge: '最新', description: '新一代 · 通用', prices: [{ value: '$5.00' }, { value: '$30.00' }, { value: '$0.50' }] },
      { id: 'gpt-5.5', badge: '热门', description: '旗舰 · 通用', prices: [{ value: '$5.00' }, { value: '$30.00' }, { value: '$0.50' }] },
      { id: 'gpt-5.4', description: '通用 · 高性能', prices: [{ value: '$2.50' }, { value: '$15.00' }, { value: '$0.25' }] },
      { id: 'gpt-5.4-mini', description: '高性价比 · 轻量', prices: [{ value: '$0.75' }, { value: '$4.50' }, { value: '$0.075' }] },
      { id: 'gpt-5.3-codex', description: '编程 · Codex', prices: [{ value: '$1.75' }, { value: '$14.00' }, { value: '$0.175' }] }
    ]
  },
  {
    key: 'claude',
    name: 'CLAUDE CODE',
    symbol: '✣',
    symbolClass: 'claude-symbol',
    models: [
      { id: 'claude-sonnet-5', badge: '热门', description: '新一代 · 通用', prices: [{ value: '$3.00' }, { value: '$15.00' }, { value: '$0.30' }] },
      { id: 'claude-fable-5', badge: '热门', description: '新一代 · 旗舰', prices: [{ value: '$10.00' }, { value: '$50.00' }, { value: '$1.00' }] },
      { id: 'claude-opus-4-8', badge: '热门', description: '旗舰 · 编程', prices: [{ value: '$5.00' }, { value: '$25.00' }, { value: '$0.50' }] },
      { id: 'claude-opus-4-7', description: '旗舰 · 编程', prices: [{ value: '$5.00' }, { value: '$25.00' }, { value: '$0.50' }] },
      { id: 'claude-opus-4-6', description: '旗舰 · 编程', prices: [{ value: '$5.00' }, { value: '$25.00' }, { value: '$0.50' }] },
      { id: 'claude-sonnet-4-6', description: '通用 · 平衡', prices: [{ value: '$3.00' }, { value: '$15.00' }, { value: '$0.30' }] },
      { id: 'claude-haiku-4-5-20251001', description: '高性价比 · 轻量', prices: [{ value: '$1.00' }, { value: '$5.00' }, { value: '$0.10' }] }
    ]
  },
  {
    key: 'gemini',
    name: 'GEMINI',
    symbol: '✦',
    symbolClass: 'gemini-symbol',
    models: [
      { id: 'gemini-3.5-flash', badge: '最新', description: '速度优先 · 搜索与 grounding', prices: [{ value: '$1.50' }, { value: '$9.00' }, { value: '$0.15' }] },
      { id: 'gemini-3.1-pro-preview', badge: '旗舰', description: '多模态 · Agent 与复杂任务', prices: [{ value: '$2.00', note: '≤200k' }, { value: '$12.00', note: '≤200k' }, { value: '$0.20', note: '≤200k' }] },
      { id: 'gemini-3.1-flash-lite', badge: '低价', description: '高吞吐 · 翻译与轻量处理', prices: [{ value: '$0.25' }, { value: '$1.50' }, { value: '$0.025' }] },
      { id: 'gemini-2.5-pro', description: '推理 · 编程与复杂任务', prices: [{ value: '$1.25', note: '≤200k' }, { value: '$10.00', note: '≤200k' }, { value: '$0.125', note: '≤200k' }] },
      { id: 'gemini-2.5-flash', description: '平衡 · 1M 上下文', prices: [{ value: '$0.30' }, { value: '$2.50' }, { value: '$0.03' }] },
      { id: 'gemini-2.5-flash-lite', description: '批量 · 极低成本', prices: [{ value: '$0.10' }, { value: '$0.40' }, { value: '$0.01' }] }
    ]
  }
]

const homeRoot = ref<HTMLElement | null>(null)
const themeToggleRef = ref<HTMLButtonElement | null>(null)
const mobileNavRef = ref<HTMLElement | null>(null)
const mobileNavToggleRef = ref<HTMLButtonElement | null>(null)
const mobileNavOpen = ref(false)
const activeTab = ref<RouteTabKey>('chat')
const monthlyTokenMillions = ref(40)
const HOME_METRICS_REFRESH_MS = 60_000
const HOME_TOKEN_STORAGE_KEY = 'usa_home_today_tokens_state'
let fallbackTokenDateKey = getLocalDateKey()
let fallbackTokenProfile = createFallbackTokenProfile(fallbackTokenDateKey)
let storedTokenState = readStoredTokenState(fallbackTokenDateKey)
let lastFallbackTokens = storedTokenState.tokens
let lastFallbackMinute = storedTokenState.minute
let lastDisplayedTokenDateKey = fallbackTokenDateKey
let lastDisplayedTokens = storedTokenState.tokens
const todayTokens = ref<number | null>(calculateFallbackTokens())
const routeLatencyMs = ref(randomRouteLatency())
const availabilityLabel = ref('99.98%')
const todayTokensLabel = ref(todayTokens.value == null ? '--' : formatMetricTokens(todayTokens.value))
const routeLatencyLabel = ref(`${routeLatencyMs.value}ms`)
const copyToastVisible = ref(false)
const copyToastMessage = ref('已复制')
let cleanupCallbacks: Array<() => void> = []
let copyToastTimer: number | undefined
let homeMetricsRefreshing = false
let availabilityMetricAnimationFrame: number | undefined
let tokenMetricAnimationFrame: number | undefined
let latencyMetricAnimationFrame: number | undefined

function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

function randomRouteLatency(): number {
  return randomInt(18, 110)
}

interface FallbackTokenProfile {
  seed: number
  dailyTargetTokens: number
  backgroundShare: number
  morningShare: number
  afternoonShare: number
  eveningShare: number
  lateShare: number
  morningStart: number
  morningEnd: number
  afternoonStart: number
  afternoonEnd: number
  eveningStart: number
  eveningEnd: number
  lateStart: number
  lateEnd: number
}

interface StoredTokenState {
  tokens: number
  minute: number
}

function getLocalDateKey(date = new Date()): string {
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

function getLocalDayStartedAt(date = new Date()): number {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime()
}

function getElapsedTokenMinute(date = new Date()): number {
  return Math.floor(Math.max(0, (date.getTime() - getLocalDayStartedAt(date)) / HOME_METRICS_REFRESH_MS))
}

function readStoredTokenState(dateKey: string): StoredTokenState {
  try {
    const raw = localStorage.getItem(HOME_TOKEN_STORAGE_KEY)
    if (!raw) return { tokens: 0, minute: 0 }
    const parsed = JSON.parse(raw) as { dateKey?: unknown; tokens?: unknown; minute?: unknown }
    const tokens = Number(parsed.tokens)
    const minute = Number(parsed.minute)
    if (parsed.dateKey !== dateKey || !Number.isFinite(tokens)) return { tokens: 0, minute: 0 }
    return {
      tokens: Math.max(0, Math.round(tokens)),
      minute: Number.isFinite(minute) ? Math.max(0, Math.floor(minute)) : 0
    }
  } catch {
    return { tokens: 0, minute: 0 }
  }
}

function persistTokenState(dateKey: string, tokens: number, minute: number) {
  try {
    localStorage.setItem(HOME_TOKEN_STORAGE_KEY, JSON.stringify({
      dateKey,
      tokens: Math.max(0, Math.round(tokens)),
      minute: Math.max(0, Math.floor(minute))
    }))
  } catch {
    // Storage can be unavailable in private or restricted browser contexts.
  }
}

function hashDateKey(dateKey: string): number {
  let hash = 2166136261
  for (let i = 0; i < dateKey.length; i += 1) {
    hash ^= dateKey.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return hash >>> 0
}

function seededUnit(seed: number, salt: number): number {
  let value = (seed + Math.imul(salt, 0x9e3779b1)) >>> 0
  value ^= value >>> 16
  value = Math.imul(value, 0x85ebca6b) >>> 0
  value ^= value >>> 13
  value = Math.imul(value, 0xc2b2ae35) >>> 0
  value ^= value >>> 16
  return value / 0xffffffff
}

function seededRange(seed: number, salt: number, min: number, max: number): number {
  return min + (max - min) * seededUnit(seed, salt)
}

function seededInt(seed: number, salt: number, min: number, max: number): number {
  return Math.round(seededRange(seed, salt, min, max))
}

function smoothstep(edge0: number, edge1: number, value: number): number {
  const x = Math.min(1, Math.max(0, (value - edge0) / (edge1 - edge0)))
  return x * x * (3 - 2 * x)
}

function calculateCumulativeMinuteTexture(seed: number, wholeMinutes: number): number {
  let total = 0
  for (let minute = 1; minute <= wholeMinutes; minute += 1) {
    total += seededInt(seed, minute + 101, 60_000, 380_000)
  }
  return total
}

function createFallbackTokenProfile(dateKey: string): FallbackTokenProfile {
  const seed = hashDateKey(dateKey)
  const morningShare = seededRange(seed, 2, 0.20, 0.29)
  const afternoonShare = seededRange(seed, 3, 0.27, 0.38)
  const eveningShare = seededRange(seed, 4, 0.19, 0.31)
  const backgroundShare = seededRange(seed, 5, 0.08, 0.15)
  const lateShare = Math.max(0.06, 1 - morningShare - afternoonShare - eveningShare - backgroundShare)
  return {
    seed,
    dailyTargetTokens: seededInt(seed, 1, 2_800_000_000, 9_600_000_000),
    backgroundShare,
    morningShare,
    afternoonShare,
    eveningShare,
    lateShare,
    morningStart: seededRange(seed, 6, 0.20, 0.28),
    morningEnd: seededRange(seed, 7, 0.43, 0.52),
    afternoonStart: seededRange(seed, 8, 0.36, 0.46),
    afternoonEnd: seededRange(seed, 9, 0.64, 0.74),
    eveningStart: seededRange(seed, 10, 0.58, 0.68),
    eveningEnd: seededRange(seed, 11, 0.82, 0.92),
    lateStart: seededRange(seed, 12, 0.76, 0.84),
    lateEnd: 1
  }
}

function syncTokenDayState(now = new Date()): string {
  const dateKey = getLocalDateKey(now)
  if (dateKey !== fallbackTokenDateKey) {
    fallbackTokenDateKey = dateKey
    fallbackTokenProfile = createFallbackTokenProfile(dateKey)
    storedTokenState = readStoredTokenState(dateKey)
    lastFallbackTokens = storedTokenState.tokens
    lastFallbackMinute = storedTokenState.minute
  }
  if (dateKey !== lastDisplayedTokenDateKey) {
    lastDisplayedTokenDateKey = dateKey
    storedTokenState = readStoredTokenState(dateKey)
    lastDisplayedTokens = storedTokenState.tokens
    lastFallbackTokens = Math.max(lastFallbackTokens, storedTokenState.tokens)
    lastFallbackMinute = Math.max(lastFallbackMinute, storedTokenState.minute)
  }
  return dateKey
}

function rememberDisplayedTokens(tokens: number, dateKey: string, minute = getElapsedTokenMinute()): number {
  if (dateKey !== lastDisplayedTokenDateKey) {
    lastDisplayedTokenDateKey = dateKey
    storedTokenState = readStoredTokenState(dateKey)
    lastDisplayedTokens = storedTokenState.tokens
    lastFallbackTokens = Math.max(lastFallbackTokens, storedTokenState.tokens)
    lastFallbackMinute = Math.max(lastFallbackMinute, storedTokenState.minute)
  }
  lastDisplayedTokens = Math.max(lastDisplayedTokens, Math.max(0, Math.round(tokens)))
  lastFallbackTokens = Math.max(lastFallbackTokens, lastDisplayedTokens)
  lastFallbackMinute = Math.max(lastFallbackMinute, minute)
  persistTokenState(dateKey, lastDisplayedTokens, lastFallbackMinute)
  return lastDisplayedTokens
}

function calculateFallbackTokens(now = new Date()): number {
  const dateKey = syncTokenDayState(now)
  const elapsedMinutes = Math.max(0, (now.getTime() - getLocalDayStartedAt(now)) / HOME_METRICS_REFRESH_MS)
  const wholeMinutes = Math.floor(elapsedMinutes)
  const dayProgress = Math.min(1, elapsedMinutes / 1440)
  const curve =
    fallbackTokenProfile.backgroundShare * dayProgress +
    fallbackTokenProfile.morningShare * smoothstep(fallbackTokenProfile.morningStart, fallbackTokenProfile.morningEnd, dayProgress) +
    fallbackTokenProfile.afternoonShare * smoothstep(fallbackTokenProfile.afternoonStart, fallbackTokenProfile.afternoonEnd, dayProgress) +
    fallbackTokenProfile.eveningShare * smoothstep(fallbackTokenProfile.eveningStart, fallbackTokenProfile.eveningEnd, dayProgress) +
    fallbackTokenProfile.lateShare * smoothstep(fallbackTokenProfile.lateStart, fallbackTokenProfile.lateEnd, dayProgress)
  const minuteTexture = calculateCumulativeMinuteTexture(fallbackTokenProfile.seed, wholeMinutes)
  const candidate = Math.round(fallbackTokenProfile.dailyTargetTokens * Math.min(1, curve) + minuteTexture)
  const minimum = lastFallbackTokens > 0 && wholeMinutes > lastFallbackMinute
    ? lastFallbackTokens + seededInt(fallbackTokenProfile.seed, wholeMinutes + 10_001, 800_000, 4_800_000)
    : lastFallbackTokens
  return rememberDisplayedTokens(Math.max(candidate, minimum), dateKey, wholeMinutes)
}

function formatMetricTokens(tokens: number): string {
  if (!Number.isFinite(tokens) || tokens <= 0) return '0'
  if (tokens >= 1_000_000_000) return `${(tokens / 1_000_000_000).toFixed(3)}B`
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(2)}M`
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}K`
  return Math.round(tokens).toString()
}

async function refreshHomeMetrics() {
  syncTokenDayState()
  if (homeMetricsRefreshing) {
    todayTokens.value = calculateFallbackTokens()
    return
  }

  homeMetricsRefreshing = true
  try {
    const metrics = await getHomeMetrics()
    const dateKey = syncTokenDayState()
    const backendTokens = Number(metrics.today_tokens)
    todayTokens.value = Number.isFinite(backendTokens) && backendTokens >= 0
      ? rememberDisplayedTokens(Math.max(backendTokens, lastDisplayedTokens), dateKey, getElapsedTokenMinute())
      : calculateFallbackTokens()
  } catch (error) {
    todayTokens.value = calculateFallbackTokens()
  } finally {
    homeMetricsRefreshing = false
  }
}

function startHomeMetricsRefresh() {
  void refreshHomeMetrics()
  const metricsTimer = window.setInterval(() => void refreshHomeMetrics(), HOME_METRICS_REFRESH_MS)
  const latencyTimer = window.setInterval(() => {
    animateAvailability()
    routeLatencyMs.value = randomRouteLatency()
  }, HOME_METRICS_REFRESH_MS)
  cleanupCallbacks.push(() => window.clearInterval(metricsTimer))
  cleanupCallbacks.push(() => window.clearInterval(latencyTimer))
}

function startMetricIntroAnimation() {
  const timer = window.setTimeout(() => {
    animateAvailability()
    animateRouteLatency(`${routeLatencyMs.value}ms`)
    animateTodayTokens(todayTokens.value == null ? '--' : formatMetricTokens(todayTokens.value))
  }, 220)
  cleanupCallbacks.push(() => window.clearTimeout(timer))
}

function scrambleMetricLabel(options: {
  target: string
  duration: number
  cancelCurrent: () => void
  setFrame: (frame: number | undefined) => void
  setValue: (value: string) => void
}) {
  options.cancelCurrent()
  const reducedMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  if (reducedMotion) {
    options.setValue(options.target)
    options.setFrame(undefined)
    return
  }

  const chars = [...options.target]
  const digitIndexes = chars
    .map((char, index) => (/\d/.test(char) ? index : -1))
    .filter((index) => index >= 0)
  const startedAt = performance.now()
  const tick = (now: number) => {
    const progress = Math.min(1, (now - startedAt) / options.duration)
    const settledDigits = Math.floor(Math.max(0, progress - 0.34) / 0.66 * (digitIndexes.length + 1))
    const next = chars.map((char, index) => {
      if (!/\d/.test(char)) return char
      const digitOrder = digitIndexes.indexOf(index)
      return digitOrder >= 0 && digitOrder < settledDigits ? char : String(randomInt(0, 9))
    }).join('')
    options.setValue(progress >= 1 ? options.target : next)
    if (progress < 1) {
      options.setFrame(window.requestAnimationFrame(tick))
      return
    }
    options.setFrame(undefined)
  }

  options.setFrame(window.requestAnimationFrame(tick))
}

function animateAvailability() {
  scrambleMetricLabel({
    target: '99.98%',
    duration: 760,
    cancelCurrent: () => {
      if (availabilityMetricAnimationFrame) window.cancelAnimationFrame(availabilityMetricAnimationFrame)
    },
    setFrame: (frame) => { availabilityMetricAnimationFrame = frame },
    setValue: (value) => { availabilityLabel.value = value }
  })
}

function animateTodayTokens(target: string) {
  scrambleMetricLabel({
    target,
    duration: 980,
    cancelCurrent: () => {
      if (tokenMetricAnimationFrame) window.cancelAnimationFrame(tokenMetricAnimationFrame)
    },
    setFrame: (frame) => { tokenMetricAnimationFrame = frame },
    setValue: (value) => { todayTokensLabel.value = value }
  })
}

function animateRouteLatency(target: string) {
  scrambleMetricLabel({
    target,
    duration: 620,
    cancelCurrent: () => {
      if (latencyMetricAnimationFrame) window.cancelAnimationFrame(latencyMetricAnimationFrame)
    },
    setFrame: (frame) => { latencyMetricAnimationFrame = frame },
    setValue: (value) => { routeLatencyLabel.value = value }
  })
}

type RouteTabKey = 'chat' | 'responses' | 'images'

const routeTabs: Array<{ key: RouteTabKey; label: string }> = [
  { key: 'chat', label: 'Chat' },
  { key: 'responses', label: 'Responses' },
  { key: 'images', label: 'Images' }
]

const routeExamples: Record<RouteTabKey, string> = {
  chat: `{
  "model": "auto:fast",
  "messages": [{"role":"user","content":"写一段发布公告"}],
  "route": {"policy":"latency-first"}
}`,
  responses: `{
  "model": "auto:reasoning",
  "input": [{"role":"user","content":[{"type":"input_text","text":"审阅这份计划"}]}],
  "metadata": {"project":"launch"}
}`,
  images: `{
  "model": "auto:image",
  "prompt": "clean product dashboard for an API gateway",
  "size": "1024x1024"
}`,
}

type WindowListener = Parameters<Window['addEventListener']>[1]
type WindowListenerOptions = Parameters<Window['addEventListener']>[2]
type ElementListener = Parameters<Element['addEventListener']>[1]
type ElementListenerOptions = Parameters<Element['addEventListener']>[2]

function toggleHomeTheme() {
  toggleTheme()
}

function closeMobileNav(restoreFocus = false): void {
  if (!mobileNavOpen.value) return
  mobileNavOpen.value = false
  if (restoreFocus) {
    void nextTick(() => mobileNavToggleRef.value?.focus())
  }
}

function toggleMobileNav(): void {
  if (mobileNavOpen.value) {
    closeMobileNav(true)
    return
  }

  mobileNavOpen.value = true
  void nextTick(() => mobileNavRef.value?.querySelector<HTMLElement>('a')?.focus())
}

function handleMobileNavNavigation(event: MouseEvent): void {
  if ((event.target as Element | null)?.closest('a')) closeMobileNav(false)
}

function on(target: Window, type: string, listener: WindowListener, options?: WindowListenerOptions) {
  target.addEventListener(type, listener, options)
  cleanupCallbacks.push(() => target.removeEventListener(type, listener, options))
}

function onElement(target: Element, type: string, listener: ElementListener, options?: ElementListenerOptions) {
  target.addEventListener(type, listener, options)
  cleanupCallbacks.push(() => target.removeEventListener(type, listener, options))
}

function setupHeader() {
  const root = homeRoot.value
  const header = root?.querySelector('.site-header')
  if (!header) return
  const update = () => header.classList.toggle('is-scrolled', window.scrollY > 18)
  const updateViewport = () => {
    if (window.innerWidth > 900) closeMobileNav(false)
  }
  on(window, 'scroll', update, { passive: true })
  on(window, 'resize', updateViewport, { passive: true })
  update()
  updateViewport()
}

function handleTablistKeydown(event: KeyboardEvent): void {
  const currentIndex = routeTabs.findIndex((tab) => tab.key === activeTab.value)
  let nextIndex = currentIndex

  if (event.key === 'ArrowRight') nextIndex = (currentIndex + 1) % routeTabs.length
  else if (event.key === 'ArrowLeft') nextIndex = (currentIndex - 1 + routeTabs.length) % routeTabs.length
  else if (event.key === 'Home') nextIndex = 0
  else if (event.key === 'End') nextIndex = routeTabs.length - 1
  else return

  event.preventDefault()
  const nextTab = routeTabs[nextIndex]
  activeTab.value = nextTab.key
  void nextTick(() => document.getElementById(`route-tab-${nextTab.key}`)?.focus())
}

function setupPricingRange() {
  const root = homeRoot.value
  const range = root?.querySelector<HTMLInputElement>('#tokenRange')
  const usageTokenValue = root?.querySelector<HTMLElement>('#usageTokenValue')
  const inputCostValue = root?.querySelector<HTMLElement>('#inputCostValue')
  if (!range || !usageTokenValue || !inputCostValue) return
  const trimNumber = (value: number, digits = 2) => value.toFixed(digits).replace(/\.0+$/, '').replace(/(\.\d*?)0+$/, '$1')
  const formatCurrency = (value: number) => `约 ¥${trimNumber(value)}`
  const formatTokenMillions = (millionTokens: number, approximate = false) => {
    const prefix = approximate ? '约 ' : ''
    if (millionTokens >= 100) return `${prefix}${trimNumber(millionTokens / 100)} 亿`
    if (millionTokens >= 1) return `${prefix}${trimNumber(millionTokens)} 百万`
    return `${prefix}${trimNumber(millionTokens * 100, 1)} 万`
  }
  const update = () => {
    monthlyTokenMillions.value = Number(range.value)
    usageTokenValue.textContent = formatTokenMillions(monthlyTokenMillions.value)
    inputCostValue.textContent = formatCurrency(monthlyTokenMillions.value / 4)
  }
  onElement(range, 'input', update)
  update()
}

function showCopyToast(message = '复制成功') {
  copyToastMessage.value = message
  copyToastVisible.value = true
  if (copyToastTimer) window.clearTimeout(copyToastTimer)
  copyToastTimer = window.setTimeout(() => {
    copyToastVisible.value = false
    copyToastTimer = undefined
  }, 1600)
}

function setupCopyButtons() {
  const root = homeRoot.value
  if (!root) return
  root.querySelectorAll<HTMLElement>('[data-copy]').forEach((button) => {
    if (button.classList.contains('copy-id')) {
      const value = button.getAttribute('data-copy') || ''
      button.textContent = ''
      button.setAttribute('aria-label', `复制 ${value}`)
      button.setAttribute('title', `复制 ${value}`)
    }
    onElement(button, 'click', async () => {
      const value = button.getAttribute('data-copy') || ''
      let copied = false
      try {
        await navigator.clipboard.writeText(value)
        copied = true
      } catch {
        // 浏览器可能禁用剪贴板权限，此处只保留视觉反馈。
      }
      button.classList.add('copied')
      window.setTimeout(() => button.classList.remove('copied'), 1200)
      if (copied) showCopyToast('复制成功')
    })
  })
}

function startCodeFlow(canvas: HTMLCanvasElement | null) {
  if (!canvas) return
  const reducedMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const snippets = [
    'POST /v1/chat/completions',
    'model: auto:reasoning',
    'route.policy = latency-first',
    'provider.openai -> ready',
    'provider.claude -> warm',
    'provider.gemini -> vision',
    'normalize(messages)',
    'meter(tokens.input)',
    'fallback.on_error = true',
    'stream: true',
    'key.scope = project',
    'usage.rolling_24h += 1',
  ]
  let streams: Array<{ x: number; y: number; direction: number; repeatWidth: number; speed: number; alpha: number; size: number; text: string }> = []
  let raf = 0
  let lastFrameTime = 0
  const color = (alpha: number) => homeTheme.value === 'dark' ? `rgba(122, 162, 255, ${alpha})` : `rgba(37, 99, 235, ${alpha})`
  const render = (deltaSeconds = 0) => {
    const width = canvas.clientWidth
    const height = canvas.clientHeight
    ctx.clearRect(0, 0, width, height)
    ctx.textBaseline = 'top'
    streams.forEach((stream) => {
      if (stream.y < -36 || stream.y > height + 24) return
      ctx.font = `${stream.size}px ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`
      const rowFade = Math.min(1, Math.max(0, stream.y / 120), Math.max(0, (height - stream.y) / 140))
      ctx.fillStyle = color(stream.alpha * rowFade)
      const copies = Math.ceil(width / stream.repeatWidth) + 3
      for (let copy = -1; copy < copies; copy += 1) {
        const x = stream.x + copy * stream.repeatWidth
        ctx.fillText(stream.text, x, stream.y)
        ctx.fillStyle = color(stream.alpha * rowFade * 0.48)
        ctx.fillText('{ normalized: true, billable: tokens }', x + stream.repeatWidth * 0.42, stream.y)
        ctx.fillStyle = color(stream.alpha * rowFade)
      }
      if (!reducedMotion) stream.x += stream.speed * deltaSeconds * stream.direction
      if (stream.direction === 1 && stream.x > stream.repeatWidth) stream.x -= stream.repeatWidth
      if (stream.direction === -1 && stream.x < -stream.repeatWidth) stream.x += stream.repeatWidth
    })
  }
  const draw = (time: number) => {
    const deltaSeconds = lastFrameTime ? Math.min((time - lastFrameTime) / 1000, 0.05) : 0
    lastFrameTime = time
    render(deltaSeconds)
    if (!reducedMotion) raf = requestAnimationFrame(draw)
  }
  const resize = () => {
    const rect = canvas.getBoundingClientRect()
    const ratio = Math.min(window.devicePixelRatio || 1, 2)
    canvas.width = Math.max(1, Math.floor(rect.width * ratio))
    canvas.height = Math.max(1, Math.floor(rect.height * ratio))
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
    const rowGap = rect.width < 640 ? 28 : 32
    const count = Math.max(22, Math.ceil(rect.height / rowGap) + 8)
    streams = Array.from({ length: count }, (_, index) => {
      const direction = index % 2 === 0 ? 1 : -1
      const size = rect.width < 640 ? 10.5 : 12 + Math.random() * 2
      const text = snippets[index % snippets.length]
      const repeatWidth = Math.max(260, text.length * size * 0.68 + 90)
      return { x: direction === 1 ? -repeatWidth + Math.random() * repeatWidth : Math.random() * repeatWidth, y: -rowGap * 2 + index * rowGap + Math.random() * 8, direction, repeatWidth, speed: 11 + Math.random() * 25, alpha: 0.08 + Math.random() * 0.17, size, text }
    })
    render(0)
  }
  resize()
  if (!reducedMotion) raf = requestAnimationFrame(draw)
  on(window, 'resize', resize, { passive: true })
  cleanupCallbacks.push(() => cancelAnimationFrame(raf))
}

function startCtaDotMatrix(canvas: HTMLCanvasElement | null) {
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const mask = document.createElement('canvas')
  const maskCtx = mask.getContext('2d', { willReadFrequently: true })
  if (!ctx || !maskCtx) return
  const reducedMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
  let width = 0
  let height = 0
  let startedAt = performance.now()
  let raf = 0
  let visible = true
  const ripples: Array<{ x: number; y: number; start: number; duration: number }> = []
  let nextRippleAt = startedAt + 500
  const draw = (time: number) => {
    if (!width || !height) return
    ctx.clearRect(0, 0, width, height)
    maskCtx.clearRect(0, 0, width, height)
    const spacing = width < 640 ? 15 : 16
    const radius = width < 640 ? 4.2 : 4.8
    const fontSize = Math.min(Math.max(height * 0.72, 180), 320)
    const text = brandName.value
    const duration = 9000
    const solidColor = homeTheme.value === 'dark' ? '248, 250, 252' : '71, 85, 105'
    const ringColor = homeTheme.value === 'dark' ? '226, 232, 240' : '15, 23, 42'
    maskCtx.font = `900 ${fontSize}px ${getComputedStyle(homeRoot.value || document.body).fontFamily}`
    maskCtx.textBaseline = 'middle'
    const textWidth = maskCtx.measureText(text).width
    const progress = reducedMotion ? 0.45 : ((time - startedAt) % duration) / duration
    const x = width + textWidth * 0.18 - progress * (width + textWidth * 1.35)
    const y = height * 0.52
    if (!reducedMotion && time >= nextRippleAt) {
      ripples.push({ x: width * (0.16 + Math.random() * 0.68), y: height * (0.18 + Math.random() * 0.58), start: time, duration: 1700 + Math.random() * 700 })
      nextRippleAt = time + 1300 + Math.random() * 1900
    }
    for (let i = ripples.length - 1; i >= 0; i -= 1) {
      if (time - ripples[i].start > ripples[i].duration) ripples.splice(i, 1)
    }
    maskCtx.fillStyle = '#000'
    maskCtx.fillText(text, x, y)
    const maskData = maskCtx.getImageData(0, 0, width, height).data
    for (let dotY = spacing * 0.75; dotY < height; dotY += spacing) {
      const rowFade = Math.min(1, Math.max(0, dotY / 58), Math.max(0, (height - dotY) / 72))
      for (let dotX = spacing * 0.5; dotX < width; dotX += spacing) {
        const index = ((Math.floor(dotY) * width) + Math.floor(dotX)) * 4 + 3
        const hit = maskData[index] > 28
        const ripple = ripples.reduce((max, source) => {
          const age = (time - source.start) / source.duration
          if (age < 0 || age > 1) return max
          const distance = Math.hypot(dotX - source.x, dotY - source.y)
          const ringRadius = 18 + age * 150
          const ringWidth = 8 + age * 18
          const primaryRing = Math.max(0, 1 - Math.abs(distance - ringRadius) / ringWidth)
          const secondaryRing = Math.max(0, 1 - Math.abs(distance - ringRadius * 0.58) / (ringWidth * 0.72)) * 0.45
          return Math.max(max, (primaryRing + secondaryRing) * Math.sin(age * Math.PI))
        }, 0)
        ctx.beginPath()
        ctx.arc(dotX, dotY, hit ? radius : radius + ripple * 2.8, 0, Math.PI * 2)
        if (hit) {
          ctx.fillStyle = `rgba(${solidColor}, ${0.66 * rowFade})`
          ctx.fill()
        } else {
          ctx.strokeStyle = `rgba(${ringColor}, ${(0.1 + ripple * 0.18) * rowFade})`
          ctx.lineWidth = 1
          ctx.stroke()
        }
      }
    }
    if (!reducedMotion && visible) raf = requestAnimationFrame(draw)
  }
  const resize = () => {
    const rect = canvas.getBoundingClientRect()
    width = Math.max(1, Math.floor(rect.width))
    height = Math.max(1, Math.floor(rect.height))
    const ratio = Math.min(window.devicePixelRatio || 1, 2)
    canvas.width = Math.floor(width * ratio)
    canvas.height = Math.floor(height * ratio)
    mask.width = width
    mask.height = height
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0)
    draw(performance.now())
  }
  resize()
  if (!reducedMotion) raf = requestAnimationFrame(draw)
  on(window, 'resize', resize, { passive: true })
  cleanupCallbacks.push(() => cancelAnimationFrame(raf))
  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver((entries) => {
      visible = entries.some((entry) => entry.isIntersecting)
      if (visible && !raf && !reducedMotion) {
        startedAt = performance.now()
        raf = requestAnimationFrame(draw)
      }
      if (!visible && raf) {
        cancelAnimationFrame(raf)
        raf = 0
      }
    }, { threshold: 0.05 })
    observer.observe(canvas)
    cleanupCallbacks.push(() => observer.disconnect())
  }
}

onMounted(() => {
  startHomeMetricsRefresh()
  startMetricIntroAnimation()
  setupHeader()
  setupPricingRange()
  setupCopyButtons()
  startCodeFlow(homeRoot.value?.querySelector<HTMLCanvasElement>('#codeFlow') || null)
  startCtaDotMatrix(homeRoot.value?.querySelector<HTMLCanvasElement>('#ctaDotMatrix') || null)
})

watch(todayTokens, (value) => {
  animateTodayTokens(value == null ? '--' : formatMetricTokens(value))
})

watch(routeLatencyMs, (value) => {
  animateRouteLatency(`${value}ms`)
})

onUnmounted(() => {
  cleanupCallbacks.forEach((cleanup) => cleanup())
  cleanupCallbacks = []
  if (copyToastTimer) window.clearTimeout(copyToastTimer)
  if (availabilityMetricAnimationFrame) window.cancelAnimationFrame(availabilityMetricAnimationFrame)
  if (tokenMetricAnimationFrame) window.cancelAnimationFrame(tokenMetricAnimationFrame)
  if (latencyMetricAnimationFrame) window.cancelAnimationFrame(latencyMetricAnimationFrame)
})
</script>

<style scoped>
.usa-home {
  color-scheme: light;
  --background: #ffffff;
  --foreground: #0f172a;
  --muted: #f4f7fb;
  --muted-2: #eef3f8;
  --muted-foreground: #667085;
  --border: #dbe3ee;
  --border-strong: #c7d3e2;
  --primary: #475569;
  --primary-strong: #111827;
  --primary-soft: #f2f5f8;
  --hero-cta-bg: #334155;
  --hero-cta-fg: #ffffff;
  --hero-cta-border: #334155;
  --hero-cta-hover-bg: #1f2937;
  --hero-cta-shadow: rgba(15, 23, 42, 0.16);
  --overview-bg: color-mix(in srgb, var(--muted) 62%, var(--background));
  --green: #059669;
  --amber: #d97706;
  --rose: #e11d48;
  --panel: rgba(255, 255, 255, 0.86);
  --surface: rgba(255, 255, 255, 0.9);
  --surface-raised: rgba(255, 255, 255, 0.96);
  --ring: rgba(15, 23, 42, 0.08);
  --soft-shadow: 0 18px 44px rgba(15, 23, 42, 0.06);
  --shadow: 0 24px 60px rgba(15, 23, 42, 0.12);
  --radius: 8px;
  --font: "Public Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.usa-home[data-theme="dark"] {
  color-scheme: dark;
  --background: #101418;
  --foreground: #eef4fb;
  --muted: #161c22;
  --muted-2: #1b232b;
  --muted-foreground: #9caab9;
  --border: #2d3946;
  --border-strong: #405064;
  --primary: #cbd5e1;
  --primary-strong: #f8fafc;
  --primary-soft: rgba(203, 213, 225, 0.1);
  --hero-cta-bg: #f8fafc;
  --hero-cta-fg: #101418;
  --hero-cta-border: rgba(248, 250, 252, 0.92);
  --hero-cta-hover-bg: #e2e8f0;
  --hero-cta-shadow: rgba(248, 250, 252, 0.18);
  --panel: rgba(16, 20, 24, 0.82);
  --surface: rgba(22, 28, 34, 0.88);
  --surface-raised: rgba(25, 32, 39, 0.96);
  --ring: rgba(238, 244, 251, 0.1);
  --soft-shadow: 0 18px 44px rgba(0, 0, 0, 0.22);
  --shadow: 0 24px 60px rgba(0, 0, 0, 0.38);
}

* { box-sizing: border-box; }

.usa-home { scroll-behavior: smooth; }

.usa-home {
  margin: 0;
  min-width: 320px;
  background: var(--background);
  color: var(--foreground);
  font-family: var(--font);
  letter-spacing: 0;
}

.usa-home::selection { background: var(--primary); color: #fff; }

.usa-home a { color: inherit; text-decoration: none; }

.usa-home button, .usa-home input { font: inherit; }

.usa-home svg { width: 1em; height: 1em; fill: currentColor; display: block; }

.usa-home :where(a, button, input, [tabindex]):focus-visible {
  outline: 2px solid color-mix(in srgb, var(--primary) 72%, #2563eb);
  outline-offset: 3px;
}

.site-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 20;
  padding: 18px 18px;
  pointer-events: none;
  transition: padding .32s ease, transform .32s ease;
}

.nav-shell {
  position: relative;
  z-index: 2;
  max-width: 1180px;
  min-height: 54px;
  margin: 0 auto;
  padding: 7px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid transparent;
  background: transparent;
  box-shadow: 0 0 0 rgba(15, 23, 42, 0);
  backdrop-filter: blur(0);
  border-radius: 0;
  pointer-events: auto;
  transition: background .38s ease, border-color .38s ease, box-shadow .38s ease, border-radius .38s ease, backdrop-filter .38s ease, transform .38s ease;
}

.site-header.is-scrolled {
  position: fixed;
  padding: 14px 18px;
}

.site-header.is-scrolled .nav-shell {
  border-color: color-mix(in srgb, var(--border) 86%, transparent);
  background: color-mix(in srgb, var(--panel) 92%, transparent);
  box-shadow: 0 14px 36px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(16px);
  border-radius: 999px;
  transform: translateY(0);
}

.brand {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-width: max-content;
  padding: 4px 10px 4px 5px;
  font-weight: 750;
}

.brand-mark {
  width: 36px;
  height: 36px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  overflow: hidden;
  background: color-mix(in srgb, var(--background) 72%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--border) 76%, transparent);
}

.brand-mark img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.nav-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  flex: 1;
  color: var(--muted-foreground);
  font-size: 14px;
  font-weight: 600;
}

.nav-links a {
  min-height: 44px;
  padding: 8px 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
}

.nav-links a:hover { background: var(--muted); color: var(--foreground); }

.nav-actions { display: flex; align-items: center; gap: 8px; }

.icon-button.mobile-nav-toggle,
.mobile-nav-backdrop { display: none; }

.icon-button, .copy-button {
  border: 1px solid var(--border);
  background: var(--background);
  color: var(--foreground);
  border-radius: 999px;
  display: inline-grid;
  place-items: center;
  cursor: pointer;
}

.icon-button { width: 44px; height: 44px; position: relative; }
.icon-button .icon { position: absolute; transition: transform .18s ease, opacity .18s ease; }
.mobile-nav-toggle svg { width: 20px; height: 20px; fill: none; }
.usa-home[data-theme="light"] .moon { opacity: 0; transform: scale(.6) rotate(-30deg); }
.usa-home[data-theme="dark"] .sun { opacity: 0; transform: scale(.6) rotate(30deg); }

.usa-home .primary-link {
  min-height: 44px;
  padding: 9px 15px;
  border-radius: 999px;
  background: var(--foreground);
  color: var(--background);
  font-size: 14px;
  font-weight: 700;
}

.hero {
  width: 100%;
  min-height: calc(100svh - 24px);
  position: relative;
  overflow: hidden;
  padding: clamp(88px, 10svh, 112px) 18px clamp(16px, 2.5svh, 26px);
  display: grid;
  grid-template-rows: auto auto;
  align-content: center;
  row-gap: clamp(8px, 1.6svh, 18px);
  align-items: start;
  justify-items: center;
  isolation: isolate;
}

.hero-grid {
  position: absolute;
  inset: 0;
  z-index: -3;
  background-image:
    linear-gradient(color-mix(in srgb, var(--border) 72%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--border) 72%, transparent) 1px, transparent 1px);
  background-size: 44px 44px;
  mask-image: linear-gradient(to bottom, transparent, black 10%, black 80%, transparent);
  opacity: .48;
}

.code-flow {
  position: absolute;
  inset: 0;
  z-index: -2;
  width: 100%;
  height: 100%;
  opacity: .66;
  pointer-events: none;
  mix-blend-mode: multiply;
  mask-image: linear-gradient(to bottom, transparent 1%, black 9%, black 86%, transparent 98%);
}

.usa-home[data-theme="dark"] .code-flow {
  opacity: .48;
  mix-blend-mode: screen;
}

.hero::before {
  content: "";
  position: absolute;
  inset: 0;
  z-index: -4;
  background:
    linear-gradient(180deg, var(--background) 0%, color-mix(in srgb, var(--primary-soft) 52%, var(--background)) 45%, var(--overview-bg) 100%),
    radial-gradient(circle at 50% 34%, color-mix(in srgb, var(--primary) 18%, transparent), transparent 38%);
}

.hero::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: clamp(150px, 22svh, 260px);
  z-index: -1;
  pointer-events: none;
  background: linear-gradient(180deg, transparent 0%, color-mix(in srgb, var(--overview-bg) 72%, transparent) 58%, var(--overview-bg) 100%);
}

.hero-content {
  width: min(900px, 100%);
  text-align: center;
  margin: 0 auto;
  position: relative;
  z-index: 2;
  transform: translateY(clamp(12px, 2.6svh, 30px));
}

.eyebrow {
  margin: 0 0 12px;
  color: var(--muted-foreground);
  font-size: 12px;
  line-height: 1.35;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0;
}

.usa-home h1, .usa-home h2, .usa-home h3, .usa-home p { margin-top: 0; }

.usa-home h1 {
  margin-bottom: 16px;
  font-size: 8rem;
  line-height: .85;
  font-weight: 900;
  letter-spacing: 0;
  overflow-wrap: anywhere;
}

.hero-copy {
  max-width: 720px;
  margin: 0 auto;
  color: var(--muted-foreground);
  font-size: 19px;
  line-height: 1.7;
}

.hero-actions {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
}

.button {
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  gap: 9px;
  border-radius: 999px;
  padding: 11px 17px;
  font-size: 14px;
  font-weight: 800;
  border: 1px solid var(--border);
  transition: background .18s ease, border-color .18s ease, color .18s ease, box-shadow .18s ease, transform .18s ease;
}

.button.primary {
  background: var(--hero-cta-bg);
  color: var(--hero-cta-fg);
  border-color: var(--hero-cta-border);
  box-shadow: 0 12px 30px var(--hero-cta-shadow);
}

.button.primary:hover {
  background: var(--hero-cta-hover-bg);
  border-color: var(--hero-cta-hover-bg);
  transform: translateY(-1px);
}

.button.primary:focus-visible {
  outline: 3px solid color-mix(in srgb, var(--hero-cta-bg) 34%, transparent);
  outline-offset: 3px;
}
.button.secondary { background: color-mix(in srgb, var(--panel) 96%, transparent); color: var(--foreground); }

.console-shell {
  width: min(1040px, 100%);
  min-height: min(350px, 38svh);
  max-height: min(410px, 43svh);
  margin-top: clamp(28px, 4svh, 48px);
  position: relative;
  z-index: 1;
  border: 1px solid color-mix(in srgb, var(--border) 88%, transparent);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--panel) 94%, transparent);
  box-shadow: var(--shadow);
  backdrop-filter: blur(18px);
  overflow: hidden;
}

@supports (height: 100dvh) {
  .hero {
    min-height: calc(100dvh - 24px);
  }

  .console-shell {
    min-height: min(350px, 38dvh);
    max-height: min(410px, 43dvh);
  }
}

.console-topbar {
  height: 52px;
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--border);
}

.window-dots { display: flex; gap: 7px; }
.window-dots span { width: 10px; height: 10px; border-radius: 50%; background: var(--border-strong); }
.window-dots span:nth-child(1) { background: #ef4444; }
.window-dots span:nth-child(2) { background: #f59e0b; }
.window-dots span:nth-child(3) { background: #22c55e; }

.route-pill, .status-pill {
  min-height: 28px;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border-radius: 999px;
  border: 1px solid var(--border);
  padding: 5px 10px;
  color: var(--muted-foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: nowrap;
}

.status-pill span { width: 8px; height: 8px; border-radius: 50%; background: #22c55e; box-shadow: 0 0 0 4px rgba(34, 197, 94, .14); }

.gateway-board {
  padding: 16px;
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(280px, .85fr);
  gap: 16px;
}

.request-panel, .flow-panel, .pricing-card, .endpoint-card, .code-window {
  min-width: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--background);
}

.panel-head {
  min-height: 44px;
  padding: 10px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border);
  color: var(--muted-foreground);
  font-size: 13px;
  font-weight: 700;
}

.copy-button { width: 40px; height: 40px; }
.copy-button.text { width: auto; min-height: 40px; padding: 8px 12px; font-size: 13px; font-weight: 800; }
.copy-button.copied {
  border-color: color-mix(in srgb, var(--foreground) 24%, var(--border));
  background: color-mix(in srgb, var(--muted) 78%, var(--surface-raised));
  color: var(--foreground);
}

.copy-toast {
  position: fixed;
  left: 50%;
  bottom: 28px;
  z-index: 80;
  transform: translate(-50%, 12px);
  border: 1px solid color-mix(in srgb, var(--border) 82%, transparent);
  border-radius: 999px;
  padding: 9px 14px;
  background: color-mix(in srgb, var(--surface-raised) 94%, transparent);
  color: var(--foreground);
  box-shadow: var(--soft-shadow);
  backdrop-filter: blur(12px);
  font-size: 13px;
  font-weight: 820;
  opacity: 0;
  pointer-events: none;
  transition: opacity .18s ease, transform .18s ease;
}
.copy-toast.is-visible {
  opacity: 1;
  transform: translate(-50%, 0);
}

.usa-home pre {
  margin: 0;
  overflow: auto;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 1.7;
}

.request-panel pre { padding: 15px; min-height: 238px; }
.muted { color: var(--muted-foreground); }
.string { color: var(--green); }

.flow-panel {
  padding: 16px;
  min-height: 282px;
  display: grid;
  gap: 14px;
  align-content: center;
}

.provider-row { display: flex; align-items: center; gap: 8px; min-width: 0; }
.provider-chip {
  flex: 0 0 auto;
  min-height: 30px;
  padding: 7px 10px;
  border-radius: 999px;
  border: 1px solid var(--border);
  font-size: 12px;
  font-weight: 800;
}
.provider-chip.openai,
.provider-chip.claude,
.provider-chip.gemini {
  color: color-mix(in srgb, var(--foreground) 72%, var(--muted-foreground));
  background: color-mix(in srgb, var(--muted) 64%, transparent);
}
.line { flex: 1 1 28px; height: 1px; background: var(--border); overflow: hidden; }
.line i { display: block; width: 60%; height: 1px; background: color-mix(in srgb, var(--foreground) 36%, transparent); animation: flow 2.7s ease-in-out infinite; }

.gateway-core {
  min-height: 116px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 6px;
  border: 1px dashed var(--border-strong);
  border-radius: var(--radius);
  position: relative;
  overflow: hidden;
}

.gateway-core::before {
  content: "async route(req) -> policy.pick(ai.fast)  const mux = normalize(req.body)\A stream.pipe(openai.chat)  fallback.to(claude)  meter.add(tokens)\A if p95.latency > 110ms { provider.next() }  cache.warm(gemini.flash)\A edge/api.in -> auth.ok -> quota.ok -> mux.ok  audit.trace(route.id)\A await router.balance({ cost, latency, context })  codec.openai(delta)\A response.delta += stream.chunk  retry.backoff(32ms)  circuit.half_open\A export /v1/chat/completions as stable.api  quota.window.sync()\A provider.score = latency * .62 + cost * .38  model.alias('auto:fast')\A mux.write({ ok: true, route, usage })  headers.set('x-usa-route')\A warmup.embedding_pool()  token.bucket.take(req.user)  policy.guard()\A cache.hit ? stream.from(cache) : upstream.fetch(req)  trace.flush()\A route.lock.release()  billing.commit(usage.total)  proxy.keepalive()\A const next = health.pick(['openai','claude','gemini'])  done(true)";
  position: absolute;
  inset: 6px 8px;
  color: color-mix(in srgb, var(--muted-foreground) 42%, transparent);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 8px;
  font-weight: 800;
  line-height: 1.28;
  white-space: pre;
  opacity: .43;
  transform: skewX(-6deg);
  pointer-events: none;
}

.gateway-core::after {
  content: "";
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 50% 50%, color-mix(in srgb, var(--background) 78%, transparent) 0%, transparent 43%);
  pointer-events: none;
}

.gateway-core::before {
  content: "";
  inset: 5px;
  background:
    repeating-linear-gradient(0deg, color-mix(in srgb, var(--muted-foreground) 22%, transparent) 0 1px, transparent 1px 12px),
    repeating-linear-gradient(90deg, color-mix(in srgb, var(--muted-foreground) 18%, transparent) 0 1px, transparent 1px 9px),
    url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='220' height='96' viewBox='0 0 220 96'%3E%3Ctext x='0' y='9' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Eroute(req)=pick(ai.fast)%3C/text%3E%3Ctext x='16' y='24' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Emux.write(delta) quota.ok%3C/text%3E%3Ctext x='4' y='39' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Efallback.to(claude) cache.hit%3C/text%3E%3Ctext x='22' y='54' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3E/v1/chat -%26gt; stream.ok%3C/text%3E%3Ctext x='8' y='69' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Etokens += usage.total%3C/text%3E%3Ctext x='28' y='84' fill='%23667085' fill-opacity='.5' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Ehealth.pick(provider)%3C/text%3E%3C/svg%3E");
  background-size: 100% 12px, 9px 100%, 220px 96px;
  opacity: .56;
  transform: skewX(-4deg) scale(1.08);
}

.gateway-core::after {
  background: radial-gradient(circle at 50% 50%, color-mix(in srgb, var(--background) 72%, transparent) 0%, transparent 37%);
}

.usa-home[data-theme="dark"] .gateway-core::before {
  background:
    repeating-linear-gradient(0deg, color-mix(in srgb, var(--muted-foreground) 28%, transparent) 0 1px, transparent 1px 12px),
    repeating-linear-gradient(90deg, color-mix(in srgb, var(--muted-foreground) 22%, transparent) 0 1px, transparent 1px 9px),
    url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='220' height='96' viewBox='0 0 220 96'%3E%3Ctext x='0' y='9' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Eroute(req)=pick(ai.fast)%3C/text%3E%3Ctext x='16' y='24' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Emux.write(delta) quota.ok%3C/text%3E%3Ctext x='4' y='39' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Efallback.to(claude) cache.hit%3C/text%3E%3Ctext x='22' y='54' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3E/v1/chat -%26gt; stream.ok%3C/text%3E%3Ctext x='8' y='69' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Etokens += usage.total%3C/text%3E%3Ctext x='28' y='84' fill='%239caab9' fill-opacity='.58' font-family='Consolas,monospace' font-size='8' font-weight='700'%3Ehealth.pick(provider)%3C/text%3E%3C/svg%3E");
  background-size: 100% 12px, 9px 100%, 220px 96px;
  opacity: .62;
}

.gateway-core::before {
  inset: 0;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='174' height='72' viewBox='0 0 174 72'%3E%3Ctext x='0' y='7' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Eroute(req).pick(ai.fast)%3C/text%3E%3Ctext x='12' y='16' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Emux.write(delta) usage+=tokens%3C/text%3E%3Ctext x='3' y='25' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Eif latency%26gt;110ms next()%3C/text%3E%3Ctext x='18' y='34' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3E/v1/chat -%26gt; stream.ok%3C/text%3E%3Ctext x='6' y='43' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Ecache.hit ? edge : upstream%3C/text%3E%3Ctext x='24' y='52' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Efallback.to(claude)%3C/text%3E%3Ctext x='2' y='61' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Ehealth.pick(provider) quota.ok%3C/text%3E%3Ctext x='15' y='70' fill='%23667085' fill-opacity='.52' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Etrace.flush() billing.commit()%3C/text%3E%3C/svg%3E");
  background-size: 174px 72px;
  background-position: 0 0;
  opacity: .72;
  transform: skewX(-4deg) scale(1.12);
}

.gateway-core::after {
  background: radial-gradient(circle at 50% 50%, color-mix(in srgb, var(--background) 48%, transparent) 0%, transparent 28%);
}

.usa-home[data-theme="dark"] .gateway-core::before {
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='174' height='72' viewBox='0 0 174 72'%3E%3Ctext x='0' y='7' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Eroute(req).pick(ai.fast)%3C/text%3E%3Ctext x='12' y='16' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Emux.write(delta) usage+=tokens%3C/text%3E%3Ctext x='3' y='25' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Eif latency%26gt;110ms next()%3C/text%3E%3Ctext x='18' y='34' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3E/v1/chat -%26gt; stream.ok%3C/text%3E%3Ctext x='6' y='43' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Ecache.hit ? edge : upstream%3C/text%3E%3Ctext x='24' y='52' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Efallback.to(claude)%3C/text%3E%3Ctext x='2' y='61' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Ehealth.pick(provider) quota.ok%3C/text%3E%3Ctext x='15' y='70' fill='%239caab9' fill-opacity='.66' font-family='Consolas,monospace' font-size='7' font-weight='700'%3Etrace.flush() billing.commit()%3C/text%3E%3C/svg%3E");
  background-size: 174px 72px;
  opacity: .78;
}

.gateway-core strong {
  position: relative;
  z-index: 1;
  color: var(--foreground);
  font-size: 18px;
  font-weight: 950;
  text-shadow: 0 1px 14px var(--background);
}

.metric-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.metric-grid div { border: 1px solid var(--border); border-radius: var(--radius); padding: 11px 8px; text-align: center; }
.metric-grid span { display: block; font-size: 17px; font-weight: 850; }
.metric-grid .metric-value { min-height: 1.2em; white-space: nowrap; font-variant-numeric: tabular-nums; }
.digit-ticker { transition: color .18s ease, text-shadow .18s ease; }
.metric-grid small { color: var(--muted-foreground); font-size: 11px; }

.brand-reveal {
  position: absolute;
  bottom: -2.4vw;
  left: 50%;
  transform: translateX(-50%);
  z-index: 0;
  color: color-mix(in srgb, var(--foreground) 13%, transparent);
  font-size: 14rem;
  font-weight: 950;
  line-height: .75;
  white-space: nowrap;
  pointer-events: none;
}

.section {
  position: relative;
  padding: 96px 18px;
  background: linear-gradient(180deg, color-mix(in srgb, var(--muted) 42%, var(--background)) 0%, var(--background) 44%, color-mix(in srgb, var(--muted) 28%, var(--background)) 100%);
}
.band-light {
  position: relative;
  margin-top: -1px;
  padding-top: clamp(8px, 1.5vw, 18px);
  background: linear-gradient(180deg, var(--overview-bg) 0%, color-mix(in srgb, var(--muted) 36%, var(--background)) 100%);
}

.band-light::before {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 140px;
  pointer-events: none;
  background: linear-gradient(180deg, color-mix(in srgb, var(--background) 18%, transparent), transparent);
}
.band-muted {
  background: linear-gradient(180deg, color-mix(in srgb, var(--muted) 40%, var(--background)) 0%, var(--muted) 52%, color-mix(in srgb, var(--muted-2) 38%, var(--background)) 100%);
}
.band-cta {
  overflow: hidden;
  isolation: isolate;
  min-height: 340px;
  background: linear-gradient(180deg, color-mix(in srgb, var(--muted-2) 36%, var(--background)) 0%, rgba(255, 255, 255, 0.86) 46%, var(--background) 100%);
  color: #111418;
}

.usa-home[data-theme="dark"] .band-cta {
  background: linear-gradient(180deg, color-mix(in srgb, var(--muted-2) 44%, var(--background)) 0%, #11161b 48%, var(--background) 100%);
  color: #f8fafc;
}



.dot-matrix {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.section-inner {
  position: relative;
  z-index: 1;
  max-width: 1120px;
  margin: 0 auto;
}
.section-title {
  max-width: 760px;
  margin: 0 auto 38px;
  text-align: center;
}
.section-title.align-left { margin: 0; text-align: left; }
.section-title h2 {
  margin-bottom: 12px;
  font-size: 3.25rem;
  line-height: 1.08;
  font-weight: 880;
  letter-spacing: 0;
}
.section-title p:not(.eyebrow) { color: var(--muted-foreground); line-height: 1.75; }

#providers {
  overflow: hidden;
  background: linear-gradient(180deg, color-mix(in srgb, var(--muted) 30%, var(--background)) 0%, var(--background) 48%, color-mix(in srgb, var(--muted) 34%, var(--background)) 100%);
  color: var(--foreground);
}

#providers::before {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
    linear-gradient(color-mix(in srgb, var(--border) 48%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--border) 48%, transparent) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: linear-gradient(to bottom, transparent, black 18%, black 82%, transparent);
}

#providers .eyebrow { color: var(--muted-foreground); }
#providers .section-title p:not(.eyebrow) { color: var(--muted-foreground); }
#providers .api-access-panel {
  border-color: color-mix(in srgb, var(--border) 86%, transparent);
  background: color-mix(in srgb, var(--surface-raised) 94%, transparent);
  box-shadow: var(--soft-shadow);
  backdrop-filter: blur(14px);
}

.feature-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 18px; }
.feature-card {
  min-height: 238px;
  position: relative;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--border) 84%, transparent);
  border-radius: var(--radius);
  background: linear-gradient(180deg, var(--surface-raised), color-mix(in srgb, var(--surface) 86%, var(--background)));
  box-shadow: var(--soft-shadow);
  padding: 20px;
}

.feature-card::before {
  content: "";
  position: absolute;
  inset: 0 0 auto;
  height: 3px;
  background: linear-gradient(90deg, color-mix(in srgb, var(--foreground) 22%, transparent), color-mix(in srgb, var(--foreground) 8%, transparent));
}
.feature-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: var(--radius);
  margin-bottom: 22px;
  box-shadow: inset 0 0 0 1px color-mix(in srgb, currentColor 18%, transparent);
}
.feature-icon svg { width: 21px; height: 21px; }
.feature-icon.blue,
.feature-icon.green,
.feature-icon.amber,
.feature-icon.rose {
  background: color-mix(in srgb, var(--muted) 70%, var(--background));
  color: color-mix(in srgb, var(--foreground) 72%, var(--muted-foreground));
}
.feature-card h3 { margin-bottom: 10px; font-size: 18px; }
.feature-card p { margin-bottom: 0; color: var(--muted-foreground); line-height: 1.65; font-size: 14px; }

.split-layout, .pricing-layout, .cta-layout {
  display: grid;
  grid-template-columns: minmax(0, .9fr) minmax(360px, 1.1fr);
  gap: clamp(34px, 6vw, 64px);
  align-items: center;
}
.api-capabilities-inner {
  display: grid;
  gap: clamp(34px, 6vw, 62px);
}
.api-capabilities-heading {
  max-width: 780px;
}
.api-capabilities-heading h2 {
  margin-bottom: 12px;
  font-size: 3.25rem;
  line-height: 1.08;
  font-weight: 880;
  letter-spacing: 0;
}
.api-capabilities-heading p:not(.eyebrow) {
  max-width: 720px;
  color: var(--muted-foreground);
  line-height: 1.75;
}
.api-capabilities-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(340px, .78fr);
  gap: clamp(28px, 5vw, 52px);
  align-items: start;
}
.capability-list {
  border-top: 1px solid color-mix(in srgb, var(--border) 82%, transparent);
}
.capability-item {
  display: grid;
  grid-template-columns: 74px minmax(0, 1fr);
  gap: 22px;
  align-items: start;
  padding: 26px 0;
  border-bottom: 1px solid color-mix(in srgb, var(--border) 82%, transparent);
}
.capability-index {
  padding-top: 6px;
  color: var(--muted-foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  font-weight: 820;
}
.capability-item h3 {
  margin: 0 0 10px;
  color: var(--foreground);
  font-size: 2rem;
  line-height: 1.18;
  font-weight: 860;
  letter-spacing: 0;
}
.capability-item p {
  max-width: 620px;
  margin: 0;
  color: var(--muted-foreground);
  font-size: 15px;
  line-height: 1.72;
}
.api-access-panel {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--border) 86%, transparent);
  border-radius: var(--radius);
  background: var(--surface-raised);
  box-shadow: var(--soft-shadow);
}
.access-panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 58px;
  padding: 0 18px;
  border-bottom: 1px solid color-mix(in srgb, var(--border) 76%, transparent);
  background: color-mix(in srgb, var(--muted) 58%, var(--background));
}
.access-panel-head span {
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: 860;
}
.access-panel-head strong {
  min-width: 0;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  font-weight: 820;
  overflow-wrap: anywhere;
}
.access-credentials {
  display: grid;
  gap: 12px;
  padding: 18px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--muted) 18%, transparent), transparent 58%),
    color-mix(in srgb, var(--surface-raised) 94%, transparent);
}
.access-credentials div {
  min-width: 0;
  display: grid;
  gap: 8px;
  padding: 14px;
  border: 1px solid color-mix(in srgb, var(--border) 82%, transparent);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--background) 78%, transparent);
}
.access-credentials span {
  color: var(--muted-foreground);
  font-size: 11px;
  font-weight: 840;
}
.access-credentials code {
  min-width: 0;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  font-weight: 780;
  overflow-wrap: anywhere;
}
.access-status-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: color-mix(in srgb, var(--border) 70%, transparent);
}
.access-status-grid div {
  min-width: 0;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 5px 8px;
  align-items: center;
  padding: 14px;
  background: color-mix(in srgb, var(--surface-raised) 94%, var(--background));
}
.access-status-grid strong {
  min-width: 0;
  color: var(--foreground);
  font-size: 13px;
  font-weight: 840;
}
.access-status-grid span {
  grid-column: 2;
  min-width: 0;
  color: var(--muted-foreground);
  font-size: 12px;
}
.status-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 0 4px color-mix(in srgb, currentColor 14%, transparent);
}
.status-dot.green { color: var(--green); }
.status-dot.amber { color: var(--amber); }
.status-dot.blue { color: #3b82f6; }
.access-terminal {
  overflow: hidden;
  border-top: 1px solid color-mix(in srgb, var(--border) 72%, transparent);
  background: #0f1115;
}
.terminal-bar {
  display: flex;
  align-items: center;
  gap: 7px;
  min-height: 34px;
  padding: 0 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: #151922;
}
.terminal-bar span {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #64748b;
}
.terminal-bar span:nth-child(1) { background: #ef4444; }
.terminal-bar span:nth-child(2) { background: #f59e0b; }
.terminal-bar span:nth-child(3) { background: #22c55e; }
.access-code {
  margin: 0;
  padding: 18px;
  background: #0f1115;
  color: #e5e7eb;
  overflow-x: auto;
  white-space: pre;
}
.access-code code {
  color: #e5e7eb;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.7;
}
.terminal-prompt {
  color: #22c55e;
  font-weight: 840;
}
.model-pricing-inner {
  display: grid;
  gap: 34px;
  margin-top: clamp(44px, 7vw, 74px);
}
.model-group {
  min-width: 0;
}
.model-group-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
  color: var(--muted-foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  letter-spacing: .18em;
}
.model-group-title strong {
  color: var(--foreground);
  font-weight: 840;
}
.provider-symbol {
  display: inline-grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border-radius: 50%;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 18px;
  letter-spacing: 0;
}
.provider-symbol.claude-symbol { color: #f97316; }
.provider-symbol.gemini-symbol { color: #3b82f6; }
.pricing-table-wrap {
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--border) 86%, transparent);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--surface-raised) 92%, transparent);
  box-shadow: var(--soft-shadow);
  backdrop-filter: blur(14px);
}

.pricing-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.pricing-table th,
.pricing-table td {
  padding: 14px 26px;
  text-align: right;
  vertical-align: middle;
}

.pricing-table th:first-child { width: 48%; text-align: left; }
.pricing-table thead th {
  height: 44px;
  padding-block: 10px;
  background: color-mix(in srgb, var(--muted) 52%, transparent);
  color: var(--muted-foreground);
  font-size: 12px;
  font-weight: 820;
}

.pricing-table tbody tr { border-top: 1px solid color-mix(in srgb, var(--border) 74%, transparent); }
.pricing-table tbody td {
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 14px;
}

.pricing-table .price-note {
  display: block;
  margin-top: 4px;
  color: var(--muted-foreground);
  font-size: 11px;
}

.model-cell {
  min-width: 0;
  text-align: left;
}
.model-cell strong {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 16px;
  font-weight: 860;
  overflow-wrap: anywhere;
}
.model-cell small {
  display: block;
  margin-top: 7px;
  color: var(--muted-foreground);
  font-size: 13px;
  font-weight: 400;
}
.model-cell em {
  border-radius: 4px;
  padding: 2px 6px;
  background: var(--foreground);
  color: var(--background);
  font-size: 11px;
  font-style: normal;
  font-weight: 820;
}
.copy-id {
  display: inline-grid;
  width: 40px;
  height: 40px;
  min-width: 40px;
  min-height: 40px;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--border) 92%, transparent);
  border-radius: 4px;
  padding: 0;
  background: color-mix(in srgb, var(--surface-raised) 88%, transparent);
  color: var(--muted-foreground);
  font-size: 0;
  line-height: 0;
  cursor: pointer;
}
.copy-id::before {
  content: "";
  width: 14px;
  height: 14px;
  background: currentColor;
  -webkit-mask: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M8 7a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3h-1v-2h1a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1h-7a1 1 0 0 0-1 1v1H8V7Zm-5 4a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3H6a3 3 0 0 1-3-3v-7Zm3-1a1 1 0 0 0-1 1v7a1 1 0 0 0 1 1h7a1 1 0 0 0 1-1v-7a1 1 0 0 0-1-1H6Z'/%3E%3C/svg%3E") center / contain no-repeat;
  mask: url("data:image/svg+xml,%3Csvg viewBox='0 0 24 24' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M8 7a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3h-1v-2h1a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1h-7a1 1 0 0 0-1 1v1H8V7Zm-5 4a3 3 0 0 1 3-3h7a3 3 0 0 1 3 3v7a3 3 0 0 1-3 3H6a3 3 0 0 1-3-3v-7Zm3-1a1 1 0 0 0-1 1v7a1 1 0 0 0 1 1h7a1 1 0 0 0 1-1v-7a1 1 0 0 0-1-1H6Z'/%3E%3C/svg%3E") center / contain no-repeat;
}
.copy-id:hover {
  border-color: color-mix(in srgb, var(--foreground) 28%, var(--border));
  color: var(--foreground);
}
.copy-id.copied {
  border-color: color-mix(in srgb, var(--foreground) 24%, var(--border));
  background: color-mix(in srgb, var(--muted) 78%, var(--surface-raised));
  color: var(--foreground);
}

.tabs {
  width: max-content;
  max-width: 100%;
  display: flex;
  gap: 6px;
  justify-content: center;
  margin: 0 auto 18px;
  padding: 5px;
  border: 1px solid var(--border);
  border-radius: 999px;
  background: var(--surface-raised);
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.06);
}
.tab {
  min-height: 40px;
  padding: 7px 14px;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--muted-foreground);
  font-weight: 800;
  cursor: pointer;
}
.tab.active { background: color-mix(in srgb, var(--foreground) 88%, var(--muted-foreground)); color: var(--background); border-color: transparent; }
.route-demo {
  display: grid;
  grid-template-columns: .95fr 1.05fr;
  gap: 18px;
  align-items: stretch;
  border: 1px solid color-mix(in srgb, var(--border) 84%, transparent);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--surface-raised) 92%, transparent);
  box-shadow: var(--soft-shadow);
  padding: 18px;
}
.route-list {
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--background);
}
.route-list div {
  min-height: 66px;
  padding: 13px 14px;
  display: grid;
  grid-template-columns: 58px minmax(0, 1fr);
  gap: 4px 12px;
  align-items: center;
  border-bottom: 1px solid color-mix(in srgb, var(--border) 78%, transparent);
}
.route-list div:last-child { border-bottom: 0; }
.route-list span { color: var(--muted-foreground); font-size: 12px; font-weight: 900; }
.route-list strong { min-width: 0; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 14px; overflow-wrap: anywhere; }
.route-list small { grid-column: 2; color: var(--muted-foreground); }
.code-window {
  min-height: 264px;
  padding: 20px;
  background: #111418;
  color: #e5e7eb;
  border-color: rgba(148, 163, 184, 0.24);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.code-window code {
  color: #e5e7eb;
}

.pricing-card {
  position: relative;
  overflow: hidden;
  padding: 24px;
  background: linear-gradient(180deg, var(--surface-raised), var(--background));
  box-shadow: var(--soft-shadow);
}

.pricing-card::before {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 4px;
  background: linear-gradient(90deg, color-mix(in srgb, var(--foreground) 20%, transparent), color-mix(in srgb, var(--foreground) 8%, transparent));
}
.pricing-head { display: flex; justify-content: space-between; align-items: baseline; gap: 16px; margin-bottom: 24px; }
.pricing-head span { color: var(--muted-foreground); font-size: 13px; font-weight: 800; }
.pricing-head strong { font-size: 30px; }
.range-label { display: flex; justify-content: space-between; gap: 12px; color: var(--muted-foreground); font-size: 14px; font-weight: 750; }
.usa-home input[type="range"] { width: 100%; margin: 18px 0; accent-color: var(--primary); }
.estimate {
  min-height: 104px;
  border: 1px solid var(--ring);
  border-radius: var(--radius);
  display: grid;
  place-items: center;
  align-content: center;
  margin-bottom: 20px;
  background: color-mix(in srgb, var(--muted) 64%, transparent);
}
.estimate span { font-size: 38px; font-weight: 900; }
.estimate small { color: var(--muted-foreground); }
.token-estimate {
  min-height: 132px;
  grid-template-columns: 1fr;
  place-items: stretch;
  align-content: stretch;
  overflow: hidden;
}
.token-estimate div {
  display: grid;
  align-content: center;
  gap: 7px;
  padding: 20px;
  text-align: center;
  border-right: 1px solid color-mix(in srgb, var(--border) 76%, transparent);
}
.token-estimate div:last-child { border-right: 0; }
.token-estimate span {
  font-size: 30px;
  line-height: 1.1;
}
.token-estimate small {
  line-height: 1.5;
}
.pricing-promo {
  margin: -4px 0 20px;
  text-align: center;
  color: var(--foreground);
  font-size: 14px;
  font-weight: 820;
  line-height: 1.6;
}
.check-list { margin: 0; padding: 0; list-style: none; display: grid; gap: 10px; color: var(--muted-foreground); }
.check-list li { position: relative; padding-left: 22px; line-height: 1.55; }
.check-list li::before { content: ""; position: absolute; left: 0; top: .58em; width: 9px; height: 9px; border-radius: 50%; background: color-mix(in srgb, var(--foreground) 42%, var(--muted-foreground)); }

.cta-layout {
  min-height: 230px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 34px;
  text-align: left;
}

.cta-copy {
  width: min(520px, 100%);
}

.cta-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
  color: #111418;
  font-size: 20px;
  font-weight: 850;
}

.usa-home[data-theme="dark"] .cta-brand {
  color: #f8fafc;
}

.cta-brand img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.band-cta h2 { margin: 0 0 14px; font-size: 3.3rem; line-height: 1.08; }
.band-cta .cta-copy p { margin: 0; color: #64748b; line-height: 1.75; }
.usa-home[data-theme="dark"] .band-cta .cta-copy p { color: #94a3b8; }
.endpoint-card {
  padding: 18px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px 12px;
  align-items: center;
  background: rgba(255, 255, 255, 0.78);
  border-color: color-mix(in srgb, var(--border) 78%, transparent);
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(10px);
}

.usa-home[data-theme="dark"] .endpoint-card {
  background: rgba(248, 250, 252, 0.08);
  border-color: rgba(226, 232, 240, 0.14);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.2);
}
.endpoint-card span { color: #64748b; font-size: 12px; font-weight: 800; grid-column: 1 / -1; }
.endpoint-card code { min-width: 0; color: #111418; overflow-wrap: anywhere; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.usa-home[data-theme="dark"] .endpoint-card span { color: #94a3b8; }
.usa-home[data-theme="dark"] .endpoint-card code { color: #f8fafc; }

.site-footer { min-height: 82px; padding: 24px 18px; display: flex; justify-content: center; align-items: center; gap: 16px; color: var(--muted-foreground); border-top: 1px solid var(--border); }
.site-footer span:first-child { color: var(--foreground); font-weight: 900; }

@keyframes flow { 0% { transform: translateX(-120%); } 100% { transform: translateX(180%); } }

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  *, *::before, *::after { animation-duration: .01ms !important; animation-iteration-count: 1 !important; transition-duration: .01ms !important; }
}

@media (max-width: 900px) {
  .site-header { padding: 12px 10px; }
  .site-header.is-scrolled { padding: 10px; }
  .nav-shell { align-items: center; flex-wrap: nowrap; gap: 8px; }
  .site-header.is-scrolled .nav-shell { border-radius: var(--radius); }
  .nav-links {
    position: absolute;
    top: calc(100% + 8px);
    left: 0;
    right: 0;
    display: none;
    align-items: stretch;
    flex-direction: column;
    gap: 2px;
    max-height: min(70dvh, 460px);
    overflow-y: auto;
    padding: 8px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--panel) 96%, transparent);
    box-shadow: var(--shadow);
    backdrop-filter: blur(18px);
  }
  .nav-links.is-open { display: flex; background: var(--panel); }
  .nav-links a { width: 100%; justify-content: flex-start; border-radius: 6px; padding-inline: 14px; }
  .nav-actions { margin-left: auto; }
  .icon-button.mobile-nav-toggle { display: inline-grid; }
  .mobile-nav-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1;
    display: block;
    width: 100%;
    height: 100%;
    border: 0;
    background: rgba(15, 23, 42, 0.28);
    pointer-events: auto;
  }
  .hero { padding-top: 96px; min-height: calc(100svh - 24px); grid-template-rows: auto auto; align-content: center; row-gap: 14px; }
  .hero-content { transform: translateY(0); }
  .usa-home h1 { font-size: 5rem; }
  .section-watermark { font-size: 8rem; }
  .section-title h2,
  .api-capabilities-heading h2,
  .band-cta h2 { font-size: 2.5rem; }
  .capability-item h3 { font-size: 1.75rem; }
  .token-estimate span { font-size: 26px; }
  .hero-actions { margin-top: 18px; }
  .console-shell { max-height: min(500px, 50svh); min-height: 0; margin-top: 30px; }
  .gateway-board { max-height: calc(min(500px, 50svh) - 52px); overflow: auto; }
  .gateway-board, .split-layout, .pricing-layout, .cta-layout, .route-demo { grid-template-columns: 1fr; }
  .api-capabilities-grid { grid-template-columns: 1fr; }
  .api-capabilities-heading { max-width: 100%; text-align: center; }
  .api-capabilities-heading p:not(.eyebrow) { margin-inline: auto; }
  .cta-layout { flex-direction: column; align-items: flex-start; }
  .route-demo { padding: 14px; }
  .feature-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .section-title.align-left { text-align: center; margin: 0 auto 30px; }
  .pricing-table th,
  .pricing-table td { padding-inline: 18px; }
}

@media (max-width: 620px) {
  .brand-name { font-size: 15px; }
  .primary-link { display: none; }
  .hero { padding: 88px 12px 18px; row-gap: 12px; align-content: start; }
  .usa-home h1 { margin-bottom: 10px; font-size: 3.9rem; }
  .hero-copy { max-width: 34rem; font-size: 15px; line-height: 1.58; }
  .hero-actions { margin-top: 16px; align-items: stretch; gap: 8px; }
  .button { min-height: 40px; flex: 1 1 136px; justify-content: center; padding: 9px 12px; font-size: 13px; }
  .console-shell { max-height: min(420px, 48svh); margin-top: 28px; border-radius: 8px; }
  .console-topbar { justify-content: flex-start; }
  .route-pill { display: none; }
  .gateway-board { max-height: calc(min(420px, 48svh) - 52px); padding: 10px; gap: 10px; }
  .request-panel,
  .flow-panel { width: 100%; overflow: hidden; }
  .panel-head { min-height: 38px; padding: 8px 10px; }
  .usa-home pre { font-size: 11px; line-height: 1.55; }
  .request-panel pre { min-height: 156px; padding: 11px; }
  .flow-panel { min-height: 190px; padding: 12px; gap: 10px; }
  .gateway-core { min-height: 82px; }
  .gateway-core strong { font-size: 15px; }
  .provider-row { flex-wrap: wrap; }
  .line { display: none; }
  .metric-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 6px; }
  .metric-grid div { padding: 8px 5px; }
  .metric-grid span { font-size: 14px; }
  .metric-grid small { font-size: 10px; }
  .feature-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .section-watermark { font-size: 5rem; }
  .section-title h2,
  .api-capabilities-heading h2,
  .band-cta h2 { font-size: 2rem; }
  .feature-card { min-height: 150px; padding: 12px; }
  .feature-card::before { height: 2px; }
  .feature-icon { width: 30px; height: 30px; margin-bottom: 10px; }
  .feature-icon svg { width: 16px; height: 16px; }
  .feature-card h3 { margin-bottom: 6px; font-size: 14px; }
  .feature-card p { font-size: 12px; line-height: 1.45; }
  .api-capabilities-heading { text-align: left; }
  .capability-item { grid-template-columns: 46px minmax(0, 1fr); gap: 14px; padding: 22px 0; }
  .capability-index { padding-top: 4px; font-size: 12px; }
  .capability-item h3 { font-size: 20px; }
  .capability-item p { font-size: 14px; line-height: 1.65; }
  .access-panel-head { align-items: flex-start; flex-direction: column; justify-content: center; padding: 12px 16px; }
  .access-credentials { padding: 12px; gap: 10px; }
  .access-credentials div { padding: 12px; }
  .access-status-grid { grid-template-columns: 1fr; }
  .access-status-grid div { padding: 12px; }
  .access-code { padding: 12px; }
  .access-code code { font-size: 11px; line-height: 1.6; }
  .model-pricing-inner { gap: 28px; margin-top: 42px; }
  .model-group-title { align-items: flex-start; flex-wrap: wrap; letter-spacing: .12em; }
  .pricing-table-wrap { overflow: visible; border: 0; background: transparent; box-shadow: none; backdrop-filter: none; }
  .pricing-table,
  .pricing-table tbody { display: block; }
  .pricing-table thead {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    clip-path: inset(50%);
    white-space: nowrap;
  }
  .pricing-table tbody { display: grid; gap: 10px; }
  .pricing-table tbody tr {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    overflow: hidden;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface-raised);
    box-shadow: var(--soft-shadow);
  }
  .pricing-table tbody th.model-cell {
    display: block;
    width: auto;
    grid-column: 1 / -1;
    padding: 14px;
    border-bottom: 1px solid var(--border);
  }
  .pricing-table tbody td {
    display: flex;
    min-width: 0;
    flex-direction: column;
    align-items: flex-start;
    gap: 3px;
    padding: 12px 8px;
    text-align: left;
  }
  .pricing-table tbody td::before {
    color: var(--muted-foreground);
    content: attr(data-label);
    font-family: var(--font);
    font-size: 10px;
    font-weight: 700;
  }
  .model-cell strong { font-size: 14px; }
  .pricing-table tbody td { font-size: 12px; }
  .tabs { width: 100%; justify-content: flex-start; overflow-x: auto; }
  .section { padding: 64px 12px; }
  .band-light { padding-top: 10px; }
  .band-cta { min-height: 420px; }
  .cta-layout { min-height: 300px; gap: 22px; }
  .route-demo { padding: 12px; }
  .pricing-card { padding: 20px; }
  .endpoint-card { grid-template-columns: 1fr; }
  .site-footer { flex-direction: column; gap: 6px; text-align: center; }
}

@media (max-height: 780px) and (min-width: 901px) {
  .usa-home h1 { font-size: 6.5rem; margin-bottom: 10px; }
  .hero-copy { font-size: 17px; line-height: 1.55; }
  .hero-actions { margin-top: 16px; }
  .console-shell { transform: scale(.9); transform-origin: top center; }
}

@media (max-height: 680px) and (min-width: 901px) {
  .hero { padding-top: 82px; }
  .console-shell { transform: scale(.78); }
}

@media (max-width: 900px) and (max-height: 760px) {
  .hero { align-content: start; row-gap: 10px; }
  .console-shell { max-height: 44svh; }
  .gateway-board { max-height: calc(44svh - 52px); }
}

</style>
