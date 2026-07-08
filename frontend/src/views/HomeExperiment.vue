<template>
  <div ref="homeRoot" class="usa-home" :data-theme="homeTheme">
<header class="site-header" aria-label="主导航">
      <div class="nav-shell">
        <a class="brand" href="#top" :aria-label="`${brandName} 首页`">
          <span class="brand-mark" aria-hidden="true"><img src="/home-experiment/logo.png" alt="" /></span>
          <span class="brand-name">{{ brandName }}</span>
        </a>
        <nav class="nav-links" aria-label="页面导航">
          <a href="#overview">总览</a>
          <a href="#providers">模型</a>
          <a href="#routes">路由</a>
          <a href="#pricing">价格</a>
          <a href="#docs">文档</a>
        </nav>
        <div class="nav-actions">
          <button class="icon-button" ref="themeToggleRef" type="button" @click="toggleHomeTheme" aria-label="切换深浅色主题" title="切换主题">
            <svg class="icon sun" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 18.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13Zm0-2a4.5 4.5 0 1 0 0-9 4.5 4.5 0 0 0 0 9ZM11 1h2v3h-2V1Zm0 19h2v3h-2v-3ZM1 11h3v2H1v-2Zm19 0h3v2h-3v-2ZM4.22 2.81l2.12 2.12-1.41 1.41-2.12-2.12 1.41-1.41Zm14.85 14.85 2.12 2.12-1.41 1.41-2.12-2.12 1.41-1.41Zm.71-14.85 1.41 1.41-2.12 2.12-1.41-1.41 2.12-2.12ZM4.93 17.66l1.41 1.41-2.12 2.12-1.41-1.41 2.12-2.12Z" /></svg>
            <svg class="icon moon" viewBox="0 0 24 24" aria-hidden="true"><path d="M21 14.69A8.5 8.5 0 0 1 9.31 3 7.1 7.1 0 1 0 21 14.69ZM12.36 1.25a10.5 10.5 0 1 0 10.39 10.39 1 1 0 0 0-1.63-.77A6.5 6.5 0 0 1 13.13 2.88a1 1 0 0 0-.77-1.63Z" /></svg>
          </button>
          <RouterLink class="primary-link" :to="entryPath">{{ entryLabel }}</RouterLink>
        </div>
      </div>
    </header>

    <main id="top">
      <section class="hero" aria-labelledby="hero-title">
        <div class="hero-grid" aria-hidden="true"></div>
        <canvas class="code-flow" id="codeFlow" aria-hidden="true"></canvas>
        <div class="hero-content">
          <p class="eyebrow">Unified Service API</p>
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

        <div class="console-shell" role="img" aria-label="USA-零 API 网关控制台预览">
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
                <span class="core-ring"></span>
                <strong>USA-零 Router</strong>
                <small>normalize · route · meter</small>
              </div>
              <div class="metric-grid">
                <div><span>99.98%</span><small>可用性</small></div>
                <div><span>138ms</span><small>路由延迟</small></div>
                <div><span>3.2M</span><small>今日 Tokens</small></div>
              </div>
            </div>
          </div>
        </div>
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
      </section>

      <section id="providers" class="section">
        <div class="section-inner split-layout">
          <div class="section-title align-left">
            <p class="eyebrow">Provider matrix</p>
            <h2>同一套 API，面向不同模型能力</h2>
            <p>保留各家模型的优势，同时把开发者入口收敛到一个域名、一种鉴权、一套日志。</p>
          </div>
          <div class="provider-matrix" aria-label="供应商能力矩阵">
            <div class="matrix-row head"><span>Provider</span><span>Text</span><span>Vision</span><span>Tools</span><span>Fallback</span></div>
            <div class="matrix-row"><strong>OpenAI</strong><span>on</span><span>on</span><span>on</span><span>hot</span></div>
            <div class="matrix-row"><strong>Claude</strong><span>on</span><span>on</span><span>on</span><span>warm</span></div>
            <div class="matrix-row"><strong>Gemini</strong><span>on</span><span>on</span><span>beta</span><span>warm</span></div>
            <div class="matrix-row"><strong>自定义上游</strong><span>map</span><span>map</span><span>map</span><span>cold</span></div>
          </div>
        </div>
      </section>

      <section id="routes" class="section band-muted">
        <div class="section-inner">
          <div class="section-title">
            <p class="eyebrow">Standard routes</p>
            <h2>常用端点保持兼容，内部路由保持弹性</h2>
          </div>
          <div class="tabs" role="tablist" aria-label="请求示例">
            <button class="tab active" type="button" role="tab" aria-selected="true" data-tab="chat">Chat</button>
            <button class="tab" type="button" role="tab" aria-selected="false" data-tab="responses">Responses</button>
            <button class="tab" type="button" role="tab" aria-selected="false" data-tab="images">Images</button>
          </div>
          <div class="route-demo">
            <div class="route-list" aria-label="标准端点">
              <div><span>POST</span><strong>/v1/chat/completions</strong><small>OpenAI-compatible</small></div>
              <div><span>POST</span><strong>/v1/responses</strong><small>multi-modal</small></div>
              <div><span>GET</span><strong>/v1/models</strong><small>normalized catalog</small></div>
              <div><span>POST</span><strong>/v1/images/generations</strong><small>provider mapped</small></div>
            </div>
            <pre class="code-window" id="tabCode"><code>{
  "model": "auto:fast",
  "messages": [{"role":"user","content":"写一段发布公告"}],
  "route": {"policy":"latency-first"}
}</code></pre>
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
              <span>Developer</span>
              <strong>按量计费</strong>
            </div>
            <label class="range-label" for="tokenRange">每月 token 预算 <span id="tokenValue">10M</span></label>
            <input id="tokenRange" type="range" min="5" max="100" value="10" step="5" />
            <div class="estimate"><span id="estimateValue">$18</span><small>预估月消耗</small></div>
            <ul class="check-list">
              <li>模型组、用户组、项目 Key 独立额度</li>
              <li>失败重试和熔断策略内置</li>
              <li>日志、账单和错误追踪统一导出</li>
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
      <span>{{ brandName }}</span><span>面向多模型产品的统一服务 API。</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

const props = defineProps<{
  siteName: string
  siteSubtitle: string
  isAuthenticated: boolean
  dashboardPath: string
}>()

const brandName = computed(() => props.siteName || 'USA-零')
const subtitle = computed(() => props.siteSubtitle || '统一 OpenAI、Claude、Gemini 等不同接口，把多模型调用规范成一个稳定、可计量、可治理的标准 API。')
const entryPath = computed(() => props.isAuthenticated ? props.dashboardPath : '/login')
const entryLabel = computed(() => props.isAuthenticated ? '进入控制台' : '开始接入')
const entryButtonLabel = computed(() => props.isAuthenticated ? '进入控制台' : '获取 API Key')
const baseUrl = computed(() => `${window.location.origin}/v1`)

const homeRoot = ref<HTMLElement | null>(null)
const themeToggleRef = ref<HTMLButtonElement | null>(null)
const homeTheme = ref<'light' | 'dark'>('light')
const activeTab = ref('chat')
const tokenBudget = ref(10)
let cleanupCallbacks: Array<() => void> = []

const routeExamples: Record<string, string> = {
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

function toggleHomeTheme() {
  homeTheme.value = homeTheme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('usa-zero-theme', homeTheme.value)
}

function on<K extends keyof WindowEventMap>(target: Window, type: K, listener: (event: WindowEventMap[K]) => void, options?: AddEventListenerOptions) {
  target.addEventListener(type, listener as EventListener, options)
  cleanupCallbacks.push(() => target.removeEventListener(type, listener as EventListener, options))
}

function onElement(target: Element, type: string, listener: EventListener, options?: AddEventListenerOptions) {
  target.addEventListener(type, listener, options)
  cleanupCallbacks.push(() => target.removeEventListener(type, listener, options))
}

function setupHeader() {
  const root = homeRoot.value
  const header = root?.querySelector('.site-header')
  if (!header) return
  const update = () => header.classList.toggle('is-scrolled', window.scrollY > 18)
  on(window, 'scroll', update, { passive: true })
  update()
}

function setupTabs() {
  const root = homeRoot.value
  if (!root) return
  const code = root.querySelector<HTMLElement>('#tabCode code')
  root.querySelectorAll<HTMLButtonElement>('.tab').forEach((button) => {
    onElement(button, 'click', () => {
      const key = button.dataset.tab || 'chat'
      activeTab.value = key
      root.querySelectorAll<HTMLButtonElement>('.tab').forEach((tab) => {
        const selected = tab === button
        tab.classList.toggle('active', selected)
        tab.setAttribute('aria-selected', selected ? 'true' : 'false')
      })
      if (code) code.textContent = routeExamples[key] || routeExamples.chat
    })
  })
}

function setupPricingRange() {
  const root = homeRoot.value
  const range = root?.querySelector<HTMLInputElement>('#tokenRange')
  const tokenValue = root?.querySelector<HTMLElement>('#tokenValue')
  const estimateValue = root?.querySelector<HTMLElement>('#estimateValue')
  if (!range || !tokenValue || !estimateValue) return
  const update = () => {
    tokenBudget.value = Number(range.value)
    tokenValue.textContent = `${tokenBudget.value}M`
    estimateValue.textContent = `$${Math.round(tokenBudget.value * 1.8)}`
  }
  onElement(range, 'input', update)
  update()
}

function setupCopyButtons() {
  const root = homeRoot.value
  if (!root) return
  root.querySelectorAll<HTMLElement>('[data-copy]').forEach((button) => {
    onElement(button, 'click', async () => {
      const value = button.getAttribute('data-copy') || ''
      try {
        await navigator.clipboard.writeText(value)
      } catch {
        // 浏览器可能禁用剪贴板权限，此处只保留视觉反馈。
      }
      button.classList.add('copied')
      window.setTimeout(() => button.classList.remove('copied'), 1200)
    })
  })
}

function setupTheme() {
  const stored = localStorage.getItem('usa-zero-theme')
  if (stored === 'dark' || stored === 'light') {
    homeTheme.value = stored
  } else if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
    homeTheme.value = 'dark'
  }
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
  const color = (alpha: number) => homeTheme.value === 'dark' ? `rgba(122, 162, 255, ${alpha})` : `rgba(37, 99, 235, ${alpha})`
  const draw = () => {
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
      if (!reducedMotion) stream.x += stream.speed * stream.direction
      if (stream.direction === 1 && stream.x > stream.repeatWidth) stream.x -= stream.repeatWidth
      if (stream.direction === -1 && stream.x < -stream.repeatWidth) stream.x += stream.repeatWidth
    })
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
      return { x: direction === 1 ? -repeatWidth + Math.random() * repeatWidth : Math.random() * repeatWidth, y: -rowGap * 2 + index * rowGap + Math.random() * 8, direction, repeatWidth, speed: 0.18 + Math.random() * 0.42, alpha: 0.08 + Math.random() * 0.17, size, text }
    })
    draw()
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
  setupTheme()
  setupHeader()
  setupTabs()
  setupPricingRange()
  setupCopyButtons()
  startCodeFlow(homeRoot.value?.querySelector<HTMLCanvasElement>('#codeFlow') || null)
  startCtaDotMatrix(homeRoot.value?.querySelector<HTMLCanvasElement>('#ctaDotMatrix') || null)
})

onUnmounted(() => {
  cleanupCallbacks.forEach((cleanup) => cleanup())
  cleanupCallbacks = []
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
  min-height: 36px;
  padding: 8px 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
}

.nav-links a:hover { background: var(--muted); color: var(--foreground); }

.nav-actions { display: flex; align-items: center; gap: 8px; }

.icon-button, .copy-button {
  border: 1px solid var(--border);
  background: var(--background);
  color: var(--foreground);
  border-radius: 999px;
  display: inline-grid;
  place-items: center;
  cursor: pointer;
}

.icon-button { width: 38px; height: 38px; position: relative; }
.icon-button .icon { position: absolute; transition: transform .18s ease, opacity .18s ease; }
.usa-home[data-theme="light"] .moon { opacity: 0; transform: scale(.6) rotate(-30deg); }
.usa-home[data-theme="dark"] .sun { opacity: 0; transform: scale(.6) rotate(30deg); }

.usa-home .primary-link {
  min-height: 38px;
  padding: 9px 15px;
  border-radius: 999px;
  background: var(--foreground);
  color: var(--background);
  font-size: 14px;
  font-weight: 700;
}

.hero {
  width: 100%;
  min-height: 100svh;
  position: relative;
  overflow: hidden;
  padding: clamp(92px, 11svh, 118px) 18px clamp(18px, 3svh, 30px);
  display: grid;
  grid-template-rows: auto minmax(250px, 1fr);
  align-items: center;
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
  margin: 0 auto clamp(16px, 3svh, 28px);
  position: relative;
  z-index: 2;
  transform: translateY(clamp(24px, 4.5svh, 54px));
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
  font-size: clamp(4rem, 12vw, 9.5rem);
  line-height: .85;
  font-weight: 900;
  letter-spacing: 0;
}

.hero-copy {
  max-width: 720px;
  margin: 0 auto;
  color: var(--muted-foreground);
  font-size: 19px;
  line-height: 1.7;
}

.hero-actions {
  margin-top: 26px;
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
  min-height: min(388px, 42svh);
  max-height: calc(100svh - clamp(92px, 11svh, 118px) - clamp(18px, 3svh, 30px) - 285px);
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
    min-height: 100dvh;
  }

  .console-shell {
    min-height: min(388px, 42dvh);
    max-height: calc(100dvh - clamp(92px, 11dvh, 118px) - clamp(18px, 3dvh, 30px) - 285px);
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
  padding: 18px;
  display: grid;
  grid-template-columns: minmax(0, 1.15fr) minmax(280px, .85fr);
  gap: 18px;
}

.request-panel, .flow-panel, .pricing-card, .endpoint-card, .code-window {
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

.copy-button { width: 30px; height: 30px; }
.copy-button.text { width: auto; min-height: 34px; padding: 7px 12px; font-size: 13px; font-weight: 800; }
.copy-button.copied { color: #fff; background: var(--green); border-color: var(--green); }

.usa-home pre {
  margin: 0;
  overflow: auto;
  color: var(--foreground);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 1.7;
}

.request-panel pre { padding: 16px; min-height: 272px; }
.muted { color: var(--muted-foreground); }
.string { color: var(--green); }

.flow-panel {
  padding: 18px;
  min-height: 316px;
  display: grid;
  gap: 18px;
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
  min-height: 132px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 6px;
  border: 1px dashed var(--border-strong);
  border-radius: var(--radius);
  position: relative;
  overflow: hidden;
}

.core-ring {
  width: 48px;
  height: 48px;
  border: 2px solid color-mix(in srgb, var(--foreground) 56%, var(--border));
  border-right-color: transparent;
  border-radius: 50%;
  animation: spin 4s linear infinite;
}
.gateway-core strong { font-size: 17px; }
.gateway-core small { color: var(--muted-foreground); font-size: 12px; }

.metric-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.metric-grid div { border: 1px solid var(--border); border-radius: var(--radius); padding: 11px 8px; text-align: center; }
.metric-grid span { display: block; font-size: 17px; font-weight: 850; }
.metric-grid small { color: var(--muted-foreground); font-size: 11px; }

.brand-reveal {
  position: absolute;
  bottom: -2.4vw;
  left: 50%;
  transform: translateX(-50%);
  z-index: 0;
  color: color-mix(in srgb, var(--foreground) 13%, transparent);
  font-size: clamp(5rem, 17vw, 17rem);
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
  font-size: clamp(2rem, 5vw, 3.25rem);
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
#providers .provider-matrix {
  border-color: color-mix(in srgb, var(--border) 86%, transparent);
  background: color-mix(in srgb, var(--surface-raised) 94%, transparent);
  box-shadow: var(--soft-shadow);
  backdrop-filter: blur(14px);
}
#providers .matrix-row { border-bottom-color: color-mix(in srgb, var(--border) 76%, transparent); }
#providers .matrix-row.head { background: color-mix(in srgb, var(--muted) 72%, var(--background)); color: var(--muted-foreground); }
#providers .matrix-row span:not(:first-child) { border-color: var(--border); background: color-mix(in srgb, var(--muted) 58%, transparent); color: var(--muted-foreground); }

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
.provider-matrix {
  border: 1px solid color-mix(in srgb, var(--border) 86%, transparent);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--surface-raised);
  box-shadow: var(--soft-shadow);
}

.matrix-row {
  display: grid;
  grid-template-columns: 1.4fr repeat(4, .8fr);
  min-height: 58px;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  border-bottom: 1px solid color-mix(in srgb, var(--border) 76%, transparent);
  font-size: 14px;
}
.matrix-row:last-child { border-bottom: 0; }
.matrix-row.head {
  min-height: 44px;
  color: var(--muted-foreground);
  background: color-mix(in srgb, var(--muted) 78%, var(--background));
  font-size: 12px;
  font-weight: 800;
  text-transform: uppercase;
}
.matrix-row span:not(:first-child) {
  width: max-content;
  min-width: 42px;
  justify-self: start;
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 4px 8px;
  color: var(--muted-foreground);
  background: color-mix(in srgb, var(--muted) 64%, transparent);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
}

.matrix-row.head span:not(:first-child) {
  min-width: 0;
  border: 0;
  border-radius: 0;
  padding: 0;
  background: transparent;
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
  min-height: 34px;
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

.band-cta h2 { margin: 0 0 14px; font-size: clamp(2rem, 5vw, 3.3rem); line-height: 1.08; }
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

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes flow { 0% { transform: translateX(-120%); } 100% { transform: translateX(180%); } }

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  *, *::before, *::after { animation-duration: .01ms !important; animation-iteration-count: 1 !important; transition-duration: .01ms !important; }
}

@media (max-width: 900px) {
  .site-header { padding: 12px 10px; }
  .site-header.is-scrolled { padding: 10px; }
  .nav-shell { align-items: flex-start; flex-wrap: wrap; }
  .site-header.is-scrolled .nav-shell { border-radius: 18px; }
  .nav-links { order: 3; flex: 0 0 100%; justify-content: flex-start; overflow-x: auto; padding: 0 2px 2px; }
  .nav-actions { margin-left: auto; }
  .hero { padding-top: 142px; min-height: 100svh; grid-template-rows: auto auto; align-content: center; }
  .hero-content { transform: translateY(clamp(8px, 2svh, 18px)); }
  .console-shell { max-height: min(520px, 48svh); min-height: 0; }
  .gateway-board { max-height: calc(min(520px, 48svh) - 52px); overflow: auto; }
  .gateway-board, .split-layout, .pricing-layout, .cta-layout, .route-demo { grid-template-columns: 1fr; }
  .cta-layout { flex-direction: column; align-items: flex-start; }
  .route-demo { padding: 14px; }
  .feature-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .section-title.align-left { text-align: center; margin: 0 auto 30px; }
}

@media (max-width: 620px) {
  .brand-name { font-size: 15px; }
  .primary-link { display: none; }
  .hero { padding: 132px 12px 22px; }
  .usa-home h1 { font-size: clamp(3.25rem, 17vw, 4.8rem); }
  .hero-copy { font-size: 16px; }
  .hero-actions { align-items: stretch; }
  .button { flex: 1 1 160px; justify-content: center; }
  .console-topbar { justify-content: flex-start; }
  .route-pill { display: none; }
  .gateway-board { padding: 12px; }
  .provider-row { flex-wrap: wrap; }
  .line { display: none; }
  .metric-grid { grid-template-columns: 1fr; }
  .feature-grid { grid-template-columns: 1fr; }
  .matrix-row { grid-template-columns: 1fr repeat(2, .62fr); }
  .matrix-row span:nth-child(4), .matrix-row span:nth-child(5) { display: none; }
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
  .usa-home h1 { font-size: clamp(3.4rem, 9.5vw, 7.2rem); margin-bottom: 10px; }
  .hero-copy { font-size: 17px; line-height: 1.55; }
  .hero-actions { margin-top: 16px; }
  .console-shell { transform: scale(.9); transform-origin: top center; }
}

@media (max-height: 680px) and (min-width: 901px) {
  .hero { padding-top: 82px; }
  .console-shell { transform: scale(.78); }
}

@media (max-width: 900px) and (max-height: 760px) {
  .hero { align-content: start; }
  .console-shell { max-height: 42svh; }
  .gateway-board { max-height: calc(42svh - 52px); }
}

</style>
