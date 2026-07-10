<template>
  <div class="gateway-auth" :data-theme="theme">
    <canvas ref="codeFlowCanvas" class="gateway-auth-code" aria-hidden="true"></canvas>

    <button
      type="button"
      class="gateway-auth-theme gateway-auth-theme-floating"
      :aria-label="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
      :title="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'"
      @click="toggleTheme"
    >
      <Icon :name="theme === 'dark' ? 'sun' : 'moon'" size="md" />
    </button>

    <main class="gateway-auth-main">
      <div class="gateway-auth-stack">
        <div class="gateway-auth-intro">
          <router-link to="/home" class="gateway-auth-brand" :aria-label="`${displaySiteName} 首页`">
            <span class="gateway-auth-logo" aria-hidden="true">
              <img src="/home-experiment/logo.png" alt="" />
            </span>
            <h1>{{ displaySiteName }}</h1>
          </router-link>
          <p class="gateway-auth-subtitle">统一模型 API 平台</p>
        </div>

        <section class="gateway-auth-panel">
          <div class="gateway-auth-body">
            <slot />
          </div>
        </section>

        <div class="gateway-auth-meta">
          <div v-if="$slots.footer" class="gateway-auth-switch">
            <slot name="footer" />
          </div>
          <p class="gateway-auth-copyright">
            &copy; {{ currentYear }} {{ displaySiteName }}. All rights reserved.
          </p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'

const props = withDefaults(defineProps<{
  siteName?: string
}>(), {
  siteName: ''
})

const appStore = useAppStore()
const displaySiteName = computed(() => props.siteName || appStore.siteName || 'Sub2API')
const theme = ref<'light' | 'dark'>('light')
const codeFlowCanvas = ref<HTMLCanvasElement | null>(null)
const currentYear = new Date().getFullYear()
let stopCodeFlow: (() => void) | null = null

function setupTheme(): void {
  const storedTheme = localStorage.getItem('usa-zero-theme')
  if (storedTheme === 'light' || storedTheme === 'dark') {
    theme.value = storedTheme
  } else if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
    theme.value = 'dark'
  }
}

function toggleTheme(): void {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  localStorage.setItem('usa-zero-theme', theme.value)
  stopCodeFlow?.()
  stopCodeFlow = startCodeFlow(codeFlowCanvas.value)
}

function startCodeFlow(canvas: HTMLCanvasElement | null): (() => void) | null {
  if (!canvas) return null
  const context = canvas.getContext('2d')
  if (!context) return null

  const reducedMotion = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
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
    'usage.rolling_24h += 1'
  ]
  let streams: Array<{
    x: number
    y: number
    direction: number
    repeatWidth: number
    speed: number
    alpha: number
    size: number
    text: string
  }> = []
  let animationFrame = 0
  let lastFrameTime = 0

  const color = (alpha: number): string =>
    theme.value === 'dark'
      ? `rgba(122, 162, 255, ${alpha})`
      : `rgba(37, 99, 235, ${alpha})`

  const render = (deltaSeconds = 0): void => {
    const width = canvas.clientWidth
    const height = canvas.clientHeight
    context.clearRect(0, 0, width, height)
    context.textBaseline = 'top'

    streams.forEach((stream) => {
      context.font = `${stream.size}px ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`
      const rowFade = Math.min(
        1,
        Math.max(0, stream.y / 110),
        Math.max(0, (height - stream.y) / 130)
      )
      context.fillStyle = color(stream.alpha * rowFade)
      const copies = Math.ceil(width / stream.repeatWidth) + 3

      for (let copy = -1; copy < copies; copy += 1) {
        const x = stream.x + copy * stream.repeatWidth
        context.fillText(stream.text, x, stream.y)
        context.fillStyle = color(stream.alpha * rowFade * 0.46)
        context.fillText(
          '{ normalized: true, billable: tokens }',
          x + stream.repeatWidth * 0.42,
          stream.y
        )
        context.fillStyle = color(stream.alpha * rowFade)
      }

      if (!reducedMotion) stream.x += stream.speed * deltaSeconds * stream.direction
      if (stream.direction === 1 && stream.x > stream.repeatWidth) stream.x -= stream.repeatWidth
      if (stream.direction === -1 && stream.x < -stream.repeatWidth) stream.x += stream.repeatWidth
    })
  }

  const draw = (time: number): void => {
    const deltaSeconds = lastFrameTime ? Math.min((time - lastFrameTime) / 1000, 0.05) : 0
    lastFrameTime = time
    render(deltaSeconds)
    if (!reducedMotion) animationFrame = window.requestAnimationFrame(draw)
  }

  const resize = (): void => {
    const rect = canvas.getBoundingClientRect()
    const ratio = Math.min(window.devicePixelRatio || 1, 2)
    canvas.width = Math.max(1, Math.floor(rect.width * ratio))
    canvas.height = Math.max(1, Math.floor(rect.height * ratio))
    context.setTransform(ratio, 0, 0, ratio, 0, 0)

    const rowGap = rect.width < 640 ? 28 : 32
    const count = Math.max(22, Math.ceil(rect.height / rowGap) + 8)
    streams = Array.from({ length: count }, (_, index) => {
      const direction = index % 2 === 0 ? 1 : -1
      const size = rect.width < 640 ? 10.5 : 12 + Math.random() * 2
      const text = snippets[index % snippets.length]
      const repeatWidth = Math.max(260, text.length * size * 0.68 + 90)
      return {
        x: direction === 1
          ? -repeatWidth + Math.random() * repeatWidth
          : Math.random() * repeatWidth,
        y: -rowGap * 2 + index * rowGap + Math.random() * 8,
        direction,
        repeatWidth,
        speed: 11 + Math.random() * 25,
        alpha: 0.08 + Math.random() * 0.17,
        size,
        text
      }
    })
    render(0)
  }

  resize()
  if (!reducedMotion) animationFrame = window.requestAnimationFrame(draw)
  window.addEventListener('resize', resize, { passive: true })
  const resizeObserver = typeof ResizeObserver === 'undefined'
    ? null
    : new ResizeObserver(resize)
  resizeObserver?.observe(canvas)

  return () => {
    window.cancelAnimationFrame(animationFrame)
    window.removeEventListener('resize', resize)
    resizeObserver?.disconnect()
  }
}

onMounted(() => {
  setupTheme()
  stopCodeFlow = startCodeFlow(codeFlowCanvas.value)
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})

onUnmounted(() => {
  stopCodeFlow?.()
})
</script>

<style scoped>
.gateway-auth {
  color-scheme: light;
  --background: #ffffff;
  --foreground: #0f172a;
  --muted: #f4f7fb;
  --muted-foreground: #667085;
  --border: #dbe3ee;
  --border-strong: #c7d3e2;
  --surface: rgba(255, 255, 255, 0.94);
  --button-background: #334155;
  --button-foreground: #ffffff;
  --button-hover: #1f2937;
  --button-shadow: rgba(15, 23, 42, 0.16);
  --shadow: 0 24px 60px rgba(15, 23, 42, 0.12);
  position: relative;
  min-width: 320px;
  min-height: 100vh;
  overflow: hidden;
  background: linear-gradient(180deg, var(--background), color-mix(in srgb, var(--muted) 64%, var(--background)));
  color: var(--foreground);
  font-family: "Public Sans", Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  letter-spacing: 0;
}

.gateway-auth[data-theme="dark"] {
  color-scheme: dark;
  --background: #101418;
  --foreground: #eef4fb;
  --muted: #161c22;
  --muted-foreground: #9caab9;
  --border: #2d3946;
  --border-strong: #405064;
  --surface: rgba(25, 32, 39, 0.94);
  --button-background: #f8fafc;
  --button-foreground: #101418;
  --button-hover: #e2e8f0;
  --button-shadow: rgba(248, 250, 252, 0.18);
  --shadow: 0 24px 60px rgba(0, 0, 0, 0.38);
}

.gateway-auth::before {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(color-mix(in srgb, var(--border) 72%, transparent) 1px, transparent 1px),
    linear-gradient(90deg, color-mix(in srgb, var(--border) 72%, transparent) 1px, transparent 1px);
  background-size: 44px 44px;
  content: "";
  mask-image: linear-gradient(to bottom, black 0%, black 78%, transparent 100%);
  opacity: 0.48;
  pointer-events: none;
}

.gateway-auth-code {
  position: absolute;
  inset: 0;
  z-index: 0;
  width: 100%;
  height: 100%;
  opacity: 0.72;
  pointer-events: none;
  mix-blend-mode: multiply;
  mask-image: linear-gradient(to bottom, transparent 1%, black 9%, black 88%, transparent 99%);
}

.gateway-auth[data-theme="dark"] .gateway-auth-code {
  opacity: 0.52;
  mix-blend-mode: screen;
}

.gateway-auth-main {
  position: relative;
  z-index: 1;
}

.gateway-auth-brand {
  display: inline-flex;
  min-width: 0;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: inherit;
  text-decoration: none;
}

.gateway-auth-brand h1 {
  min-width: 0;
  margin: 0;
  font-size: 30px;
  line-height: 1.15;
  font-weight: 850;
  letter-spacing: 0;
  overflow-wrap: anywhere;
}

.gateway-auth-logo {
  display: grid;
  width: 64px;
  height: 64px;
  flex: 0 0 64px;
  place-items: center;
  overflow: hidden;
  border-radius: 50%;
  background: color-mix(in srgb, var(--background) 72%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--border) 76%, transparent);
}

.gateway-auth-logo img { width: 100%; height: 100%; object-fit: cover; }

.gateway-auth-theme {
  display: inline-grid;
  width: 38px;
  height: 38px;
  place-items: center;
  border: 1px solid var(--border);
  border-radius: 50%;
  background: var(--background);
  color: var(--foreground);
  cursor: pointer;
}

.gateway-auth-theme-floating {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 10;
  backdrop-filter: blur(12px);
}

.gateway-auth-main {
  display: flex;
  width: min(480px, calc(100% - 36px));
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  padding: 54px 0;
}

.gateway-auth-stack {
  display: grid;
  width: 100%;
  gap: 26px;
}

.gateway-auth-intro {
  text-align: center;
}

.gateway-auth-subtitle {
  margin: 9px auto 0;
  color: var(--muted-foreground);
  font-size: 14px;
  line-height: 1.5;
}

.gateway-auth-panel {
  width: 100%;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--border) 90%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--surface) 94%, transparent);
  box-shadow: var(--shadow);
  backdrop-filter: blur(18px);
}

.gateway-auth-body { padding: 34px 36px 30px; }

.gateway-auth-meta {
  display: grid;
  gap: 22px;
  text-align: center;
}

.gateway-auth-switch {
  color: var(--muted-foreground);
  font-size: 13px;
}

.gateway-auth-switch :deep(p) {
  margin: 0;
}

.gateway-auth-switch :deep(a) {
  margin-left: 4px;
  color: var(--foreground) !important;
  font-weight: 800;
  text-decoration: none;
}

.gateway-auth-copyright {
  margin: 0;
  color: var(--muted-foreground);
  font-size: 12px;
}

.gateway-auth :deep(.auth-form-heading) { text-align: center; }

.gateway-auth :deep(.auth-form-heading h2) {
  margin: 0;
  color: var(--foreground);
  font-size: 28px;
  line-height: 1.2;
  font-weight: 850;
  letter-spacing: 0;
}

.gateway-auth :deep(.auth-form-heading > p:last-child) {
  margin: 8px 0 0;
  color: var(--muted-foreground);
  font-size: 14px;
  line-height: 1.6;
}

.gateway-auth :deep(.input-label) { color: var(--foreground); font-size: 13px; font-weight: 700; }
.gateway-auth :deep(.input) {
  min-height: 44px;
  border-color: var(--border);
  border-radius: 8px;
  background: var(--background);
  color: var(--foreground);
  box-shadow: none;
}

.gateway-auth :deep(.input::placeholder) { color: var(--muted-foreground); opacity: 0.72; }
.gateway-auth :deep(.input:focus) {
  border-color: var(--border-strong);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--foreground) 9%, transparent);
}

.gateway-auth :deep(.input:disabled) { background: var(--muted); }
.gateway-auth :deep(.input-hint),
.gateway-auth :deep(.text-gray-500),
.gateway-auth :deep(.text-gray-400),
.gateway-auth :deep(.dark\:text-dark-400),
.gateway-auth :deep(.dark\:text-dark-500) { color: var(--muted-foreground) !important; }

.gateway-auth :deep(.bg-gray-200),
.gateway-auth :deep(.dark\:bg-dark-700) { background: var(--border) !important; }

.gateway-auth :deep(.auth-submit.btn-primary) {
  min-height: 46px;
  border: 1px solid var(--button-background);
  border-radius: 8px;
  background: var(--button-background);
  color: var(--button-foreground);
  box-shadow: 0 12px 30px var(--button-shadow);
  font-weight: 800;
}

.gateway-auth :deep(.auth-submit.btn-primary:hover:not(:disabled)) {
  border-color: var(--button-hover);
  background: var(--button-hover);
  transform: translateY(-1px);
}

.gateway-auth :deep(.btn-secondary) {
  min-height: 44px;
  border-color: var(--border);
  border-radius: 8px;
  background: var(--background);
  color: var(--foreground);
  box-shadow: none;
}

@media (max-width: 900px) {
  .gateway-auth-main {
    width: min(480px, calc(100% - 24px));
    min-height: 100vh;
    padding: 48px 0 44px;
  }
}

@media (max-width: 520px) {
  .gateway-auth-theme-floating { top: 12px; right: 12px; }
  .gateway-auth-main { width: min(100% - 20px, 480px); padding-top: 64px; }
  .gateway-auth-stack { gap: 22px; }
  .gateway-auth-logo { width: 56px; height: 56px; flex-basis: 56px; }
  .gateway-auth-brand h1 { font-size: 24px; }
  .gateway-auth-subtitle { font-size: 13px; }
  .gateway-auth-body { padding: 26px 20px 24px; }
  .gateway-auth :deep(.auth-form-heading h2) { font-size: 24px; }
}

@media (prefers-reduced-motion: reduce) {
  .gateway-auth *,
  .gateway-auth *::before,
  .gateway-auth *::after {
    animation-duration: 1ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 1ms !important;
  }
}
</style>
