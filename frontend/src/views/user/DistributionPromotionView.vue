<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5" :aria-busy="loading || statsLoading || visitsLoading">
      <header class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0">
          <h1 class="page-title">推广中心</h1>
          <p class="page-description">分享专属链接，查看注册与客户转化</p>
        </div>
        <button class="btn btn-secondary btn-icon min-h-11 min-w-11 shrink-0" :disabled="loading" title="刷新" aria-label="刷新" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </header>

      <DistributionNav />

      <section v-if="overviewError" class="flex flex-wrap items-center justify-between gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4" role="alert">
        <div><p class="text-sm font-medium text-danger-foreground">推广信息加载失败</p><p class="mt-1 text-xs text-foreground-subtle">页面未显示可能过期的数据，请检查网络后重试。</p></div>
        <button class="btn btn-secondary min-h-11" type="button" :disabled="loading" @click="load">重试</button>
      </section>
      <div v-if="loading && !overview" class="flex min-h-48 items-center justify-center" role="status" aria-label="正在加载推广信息"><LoadingSpinner /></div>
      <template v-if="overview">
        <section v-if="!linkEnabled" class="flex items-start gap-3 rounded-panel border border-warning/40 bg-warning-subtle p-4 text-sm text-warning-foreground" role="status">
          <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" />
          <div>
            <p class="font-medium">{{ statusText }}</p>
            <p class="mt-1">推广链接暂时不能绑定新客户，请联系平台管理员处理。</p>
          </div>
        </section>
        <section v-else-if="!trackingEnabled" class="flex items-start gap-3 rounded-panel border border-warning/40 bg-warning-subtle p-4 text-sm text-warning-foreground" role="status">
          <Icon name="exclamationCircle" size="sm" class="mt-0.5 shrink-0" />
          <div>
            <p class="font-medium">推广访问追踪已暂停</p>
            <p class="mt-1">链接仍可用于注册，历史数据保留；追踪重新开启后才会记录新的访问。</p>
          </div>
        </section>

        <section class="grid min-w-0 gap-4" :class="linkEnabled ? 'lg:grid-cols-[minmax(220px,260px)_minmax(0,1fr)]' : ''" aria-label="推广工具">
          <div v-if="linkEnabled" class="card flex min-w-0 flex-col items-center p-5 text-center">
            <h2 id="promotion-tools-title" class="text-base font-semibold">推广二维码</h2>
            <p class="mt-1 text-sm text-foreground-subtle">扫码后进入注册页面</p>
            <div class="mt-4 flex size-44 shrink-0 items-center justify-center rounded-panel border border-outline bg-white p-3">
              <canvas v-show="!qrError" ref="qrCanvas" class="size-full" role="img" aria-label="打开专属注册页面的推广二维码">专属注册链接二维码</canvas>
              <div v-if="qrError" class="flex size-full flex-col items-center justify-center gap-2 rounded-control bg-inverse text-xs text-inverse-foreground" role="alert"><span>二维码生成失败</span><button type="button" class="min-h-11 text-inverse-foreground underline" @click="renderQr">重试</button></div>
            </div>
            <button class="btn btn-secondary mt-4 min-h-11 w-full" :disabled="!linkEnabled" @click="downloadQr"><Icon name="download" size="sm" />下载二维码</button>
          </div>

          <div class="card min-w-0 p-4 sm:p-5">
            <div class="flex min-w-0 items-start justify-between gap-3">
              <div class="min-w-0"><h2 class="text-base font-semibold">专属推广信息</h2><p class="mt-1 text-sm text-foreground-subtle">当前返佣比例 {{ rate }}</p></div>
              <span class="badge shrink-0" :class="linkEnabled ? (trackingEnabled ? 'badge-success' : 'badge-warning') : 'badge-gray'">{{ linkEnabled ? (trackingEnabled ? '推广中' : '可注册 · 不统计') : statusText }}</span>
            </div>
            <div v-if="linkEnabled" class="mt-5 grid min-w-0 gap-4 sm:grid-cols-2">
              <div class="min-w-0 sm:col-span-2"><p class="input-label">推广码</p><div class="flex min-w-0 items-center gap-2 rounded-control border border-outline bg-surface-subtle p-2 pl-3"><code class="min-w-0 flex-1 break-all font-mono text-sm">{{ overview.agent.promotion_code }}</code><button class="btn btn-secondary btn-icon min-h-11 min-w-11 shrink-0" title="复制推广码" aria-label="复制推广码" :disabled="!linkEnabled" @click="copyText(overview.agent.promotion_code, '推广码已复制')"><Icon name="copy" size="sm" /></button></div></div>
              <div class="min-w-0 sm:col-span-2"><p class="input-label">推广链接</p><div class="flex min-w-0 items-center gap-2 rounded-control border border-outline bg-surface-subtle p-2 pl-3"><code class="min-w-0 flex-1 break-all text-sm leading-5" :title="promotionLink">{{ promotionLink }}</code><button class="btn btn-primary btn-icon min-h-11 min-w-11 shrink-0" title="复制推广链接" aria-label="复制推广链接" :disabled="!linkEnabled" @click="copyText(promotionLink, '推广链接已复制')"><Icon name="copy" size="sm" /></button></div></div>
            </div>
            <div v-else class="mt-5 flex items-start gap-3 rounded-panel border border-outline bg-surface-subtle p-4 text-left text-sm text-foreground-subtle">
              <Icon name="lock" size="sm" class="mt-0.5 shrink-0" />
              <p>账号恢复可推广状态后，这里才会重新提供二维码、推广码和注册链接。</p>
            </div>
            <dl class="mt-5 grid grid-cols-1 gap-3 border-t border-outline pt-4 sm:grid-cols-3">
              <div><dt class="text-xs text-foreground-subtle">直属客户</dt><dd class="mt-1 text-lg font-semibold tabular-nums">{{ overview.customer_count }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">团队付费客户</dt><dd class="mt-1 text-lg font-semibold tabular-nums">{{ overview.paying_customer_count }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">团队累计实付</dt><dd class="mt-1 whitespace-nowrap text-lg font-semibold tabular-nums">¥{{ Number(overview.customer_paid_cny).toFixed(2) }}</dd></div>
            </dl>
          </div>
        </section>

        <section v-if="!overview.agent.can_view_promotion_stats" class="rounded-panel border border-outline bg-surface-subtle p-4 text-sm text-foreground-subtle" role="status">
          当前账号未开通推广数据权限。推广链接仍可使用，请联系管理员开通统计权限。
        </section>

        <template v-else>
          <section class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-outline pb-3">
            <div class="min-w-0"><h2 class="font-semibold">推广表现</h2><p class="text-xs text-foreground-subtle">独立访客按浏览器去重；一分钟内重复打开只计一次有效访问</p></div>
            <div class="flex shrink-0 rounded-control border border-outline bg-surface-subtle p-0.5" role="group" aria-label="统计周期">
              <button v-for="option in periodOptions" :key="option.value" type="button" class="min-h-11 min-w-11 rounded px-2 text-xs font-medium transition-colors" :class="days === option.value ? 'bg-surface text-foreground shadow-sm' : 'text-foreground-muted hover:text-foreground'" :aria-pressed="days === option.value" :disabled="statsLoading || visitsLoading" @click="setDays(option.value)">{{ option.label }}</button>
            </div>
          </section>

          <p class="sr-only" role="status" aria-live="polite">{{ statsLoading || visitsLoading ? '正在更新推广统计' : '' }}</p>

          <section v-if="analytics?.meta" class="rounded-panel border border-outline bg-surface-subtle px-4 py-3 text-xs text-foreground-subtle" aria-label="当前归因规则">
            <strong class="font-medium text-foreground">归因规则：</strong>{{ attributionPolicyText }}。筛选采用访问批次口径，注册计入其获归因的推广访问发生日。
          </section>

          <section v-if="statsError" class="flex items-center justify-between gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground" role="alert">
            <span>推广统计暂时加载失败，历史数据未受影响。</span><button class="btn btn-secondary min-h-11 shrink-0" :disabled="statsLoading" @click="loadStats">重试</button>
          </section>
          <template v-else-if="analytics">
            <section class="grid grid-cols-2 gap-3 xl:grid-cols-4" aria-label="访问转化指标">
              <article v-for="card in statCards" :key="card.label" class="card min-w-0 p-4"><p class="text-xs text-foreground-subtle">{{ card.label }}</p><p class="mt-2 break-words text-xl font-semibold tabular-nums sm:text-2xl">{{ card.value }}</p><p class="mt-1 text-xs leading-5 text-foreground-muted">{{ card.hint }}</p></article>
            </section>
            <p class="text-xs text-foreground-muted">数据质量：识别到 {{ analytics.summary.bot_visits.toLocaleString() }} 次机器人访问；{{ botPolicyText }}。</p>

            <section class="border-y border-outline py-4" aria-labelledby="registration-results-title">
              <div class="flex flex-wrap items-end justify-between gap-2">
                <div><h3 id="registration-results-title" class="font-medium">访问批次的注册结果</h3><p class="mt-1 text-xs text-foreground-subtle">归属于当前访问人群，按获归因的推广访问日期统计</p></div>
                <span class="text-xs text-foreground-muted">最近 {{ days }} 天</span>
              </div>
              <dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-4">
                <div v-for="item in registrationResults" :key="item.label" class="min-w-0"><dt class="text-xs text-foreground-subtle">{{ item.label }}</dt><dd class="mt-1 break-words text-lg font-semibold tabular-nums">{{ item.value }}</dd></div>
              </dl>
              <p v-if="analytics.summary.untracked_direct > 0" class="mt-4 rounded-panel border border-warning/30 bg-warning-subtle px-3 py-2 text-xs leading-5 text-warning-foreground">{{ analytics.summary.untracked_direct.toLocaleString() }} 人携推广码注册但未匹配到有效访问，因此计入总注册，不计入访客转化率。</p>
            </section>

            <section class="grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(260px,0.6fr)]">
              <div class="border-y border-outline py-4">
                <div class="mb-4 flex flex-wrap items-end justify-between gap-3"><div><h3 class="font-medium">访问量与归因注册</h3><p class="mt-1 text-xs text-foreground-subtle">同一访问批次口径；注册显示在获归因的推广访问日期</p></div><span class="shrink-0 text-xs text-foreground-subtle">{{ days === 90 ? '按周' : '按日' }} · 最近 {{ days }} 天</span></div>
                <div class="mb-3 flex flex-wrap items-center gap-4 text-xs text-foreground-subtle" aria-hidden="true"><span class="inline-flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-info"></span>访问</span><span class="inline-flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-success"></span>注册</span></div>
                <div class="max-h-72 space-y-2 overflow-y-auto pr-1" role="list" aria-label="访问与注册趋势图">
                  <div v-for="item in chartRows" :key="item.date" class="grid grid-cols-[52px_minmax(0,1fr)] items-center gap-x-2 gap-y-1 text-xs sm:grid-cols-[64px_minmax(0,1fr)_minmax(58px,auto)]" role="listitem" :aria-label="`${chartLabel(item.date)}，${item.visits} 次访问，${item.registrations} 次注册`"><span class="text-foreground-subtle">{{ chartLabel(item.date) }}</span><div class="space-y-1" aria-hidden="true"><div class="h-1.5 overflow-hidden rounded bg-surface-subtle"><div class="h-full rounded bg-info" :style="{ width: `${Math.max(item.visits ? 2 : 0, item.visits / maxChartValue * 100)}%` }"></div></div><div class="h-1.5 overflow-hidden rounded bg-surface-subtle"><div class="h-full rounded bg-success" :style="{ width: `${Math.max(item.registrations ? 2 : 0, item.registrations / maxChartValue * 100)}%` }"></div></div></div><span class="col-start-2 min-w-0 break-words text-right tabular-nums sm:col-start-auto" aria-hidden="true">访问 {{ item.visits }} · 注册 {{ item.registrations }}</span></div>
                  <p v-if="!chartRows.length" class="py-5 text-center text-sm text-foreground-subtle">当前周期暂无访问</p>
                </div>
              </div>
              <div class="border-y border-outline py-4"><h3 class="font-medium">主要来源</h3><p class="mt-1 text-xs text-foreground-subtle">访问 / 已转化访客</p><div class="mt-4 space-y-3"><div v-for="source in analytics.sources" :key="source.source" class="flex min-w-0 items-center justify-between gap-3 text-sm"><span class="min-w-0 break-words text-foreground-subtle">{{ source.source }}</span><span class="shrink-0 text-right tabular-nums">{{ source.visits }} / {{ source.conversions }}</span></div><p v-if="!analytics.sources.length" class="text-sm text-foreground-subtle">暂无来源数据</p></div></div>
            </section>
          </template>
          <div v-else-if="statsLoading" class="flex min-h-32 items-center justify-center" role="status" aria-label="正在加载推广统计"><LoadingSpinner /></div>

          <section>
            <div class="mb-3 flex min-w-0 items-center justify-between gap-3"><div class="min-w-0"><h3 class="font-medium">最近访问</h3><p class="text-xs text-foreground-subtle">默认展示最新 8 条；可继续加载更早记录</p></div><span class="shrink-0 text-xs text-foreground-muted">已显示 {{ visits.length }} / {{ visitsTotal }} 条</span></div>
            <div v-if="visitsError && !visits.length" class="flex items-center justify-between gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground" role="alert"><span>访问明细暂时加载失败。</span><button class="btn btn-secondary min-h-11 shrink-0" :disabled="visitsLoading" @click="loadVisits">重试</button></div>
            <div v-else-if="visitsLoading && !visits.length" class="flex min-h-24 items-center justify-center" role="status" aria-label="正在加载最近访问"><LoadingSpinner /></div>
            <template v-else-if="visits.length"><div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
              <article v-for="visit in visits" :key="visit.id" class="min-w-0 rounded-panel border border-outline p-3"><div class="flex items-center justify-between gap-2"><span class="badge" :class="visit.registered_at ? 'badge-success' : 'badge-gray'">{{ visit.registered_at ? '已转化访客' : '尚未注册' }}</span><span class="shrink-0 text-xs text-foreground-muted">{{ deviceLabel(visit.device_type) }}</span></div><p class="mt-3 break-words text-sm font-medium">{{ visit.source }}</p><p class="mt-1 text-xs text-foreground-subtle">访问 {{ formatDateTime(visit.visited_at) }}</p><p v-if="visit.registered_at" class="mt-1 text-xs text-success-foreground">{{ attributionLabel(visit.attribution_type) }} · 注册 {{ formatDateTime(visit.registered_at) }}</p></article>
            </div><div class="mt-3 flex flex-col items-center gap-2"><p v-if="visitsLoadMoreError" class="text-sm text-danger-foreground" role="alert">更早记录加载失败，请重试。</p><button v-if="visits.length < visitsTotal" class="btn btn-secondary min-h-11" type="button" :disabled="visitsLoading" @click="loadMoreVisits">{{ visitsLoading ? '加载中...' : '查看更多访问' }}</button></div></template>
            <div v-else class="rounded-panel border border-dashed border-outline p-8 text-center text-sm text-foreground-subtle">当前周期暂无推广访问</div>
          </section>
        </template>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import QRCode from 'qrcode'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DistributionNav from '@/components/distribution/DistributionNav.vue'
import { getDistributionOverview, getPromotionAnalytics, listPromotionVisits, type DistributionOverview, type DistributionPromotionAnalytics, type DistributionPromotionVisit } from '@/api/distribution'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const overview = ref<DistributionOverview | null>(null)
const analytics = ref<DistributionPromotionAnalytics | null>(null)
const visits = ref<DistributionPromotionVisit[]>([])
const visitsTotal = ref(0)
const visitsPage = ref(1)
const loading = ref(false)
const overviewError = ref(false)
const statsLoading = ref(false)
const visitsLoading = ref(false)
const statsError = ref(false)
const visitsError = ref(false)
const visitsLoadMoreError = ref(false)
const qrError = ref(false)
const days = ref(30)
const qrCanvas = ref<HTMLCanvasElement | null>(null)
let overviewSequence = 0
let statsSequence = 0
let visitsSequence = 0
const periodOptions = [{ label: '7天', value: 7 }, { label: '30天', value: 30 }, { label: '90天', value: 90 }]
const promotionLink = computed(() => overview.value ? `${window.location.origin}/register?agent=${encodeURIComponent(overview.value.agent.promotion_code)}` : '')
const linkEnabled = computed(() => overview.value?.distribution_enabled === true && overview.value?.agent.status === 'active')
const trackingEnabled = computed(() => overview.value?.distribution_enabled === true && overview.value?.promotion_tracking_enabled === true)
const statusText = computed(() => overview.value?.distribution_enabled === false ? '分销已暂停' : (overview.value?.agent.status === 'suspended' ? '代理账号已暂停' : '代理账号已撤销'))
const rate = computed(() => `${((overview.value?.agent.effective_rate_bps || 0) / 100).toFixed(2)}%`)
const dateParams = computed(() => ({ date_from: hongKongDate(-(days.value - 1)), date_to: hongKongDate() }))
const statCards = computed(() => analytics.value ? [
  { label: '有效访问', value: analytics.value.summary.total_visits.toLocaleString(), hint: '一分钟内重复打开已去重' },
  { label: '独立访客', value: analytics.value.summary.unique_visitors.toLocaleString(), hint: '按去标识化访客标识去重' },
  { label: '已转化访客', value: analytics.value.summary.converted_visitors.toLocaleString(), hint: '当前访问人群中已完成注册' },
  { label: '访客转化率', value: `${analytics.value.summary.conversion_rate.toFixed(2)}%`, hint: '已转化访客 ÷ 独立访客' },
] : [])
const registrationResults = computed(() => analytics.value ? [
  { label: '总注册', value: analytics.value.summary.registrations.toLocaleString() },
  { label: '直接携码', value: analytics.value.summary.direct_registrations.toLocaleString() },
  { label: '历史归因', value: analytics.value.summary.persisted_registrations.toLocaleString() },
  { label: '未匹配有效访问', value: analytics.value.summary.untracked_direct.toLocaleString() },
] : [])
const chartRows = computed(() => {
  const daily = [...(analytics.value?.daily || [])].sort((left, right) => right.date.localeCompare(left.date))
  if (days.value !== 90) return daily
  const grouped = new Map<string, { date: string; visits: number; registrations: number }>()
  for (const item of daily) {
    const date = parseLocalDate(item.date)
    const day = date.getDay() || 7
    date.setDate(date.getDate() - day + 1)
    const key = formatLocalDate(date)
    const row = grouped.get(key) || { date: key, visits: 0, registrations: 0 }
    row.visits += item.visits; row.registrations += item.registrations; grouped.set(key, row)
  }
  return Array.from(grouped.values())
})
const maxChartValue = computed(() => Math.max(1, ...chartRows.value.flatMap((item) => [item.visits, item.registrations])))
const botPolicyText = computed(() => analytics.value?.meta?.bot_filter_enabled === true ? '当前不计入汇总指标' : analytics.value?.meta?.bot_filter_enabled === false ? '当前计入汇总指标' : '过滤规则以平台设置为准')
const attributionPolicyText = computed(() => {
  const meta = analytics.value?.meta
  if (!meta?.attribution_enabled) return '历史访问归因已关闭；携推广码注册仍按直接归因记录'
  const model = meta.attribution_model === 'last_touch' ? '最近访问归因' : '首次访问归因'
  return `${model}，归因窗口 ${meta.attribution_days} 天`
})

async function renderQr() {
  await nextTick()
  if (!qrCanvas.value || !promotionLink.value) return
  try { await QRCode.toCanvas(qrCanvas.value, promotionLink.value, { width: 160, margin: 1, errorCorrectionLevel: 'M' }); qrError.value = false }
  catch { qrError.value = true; app.showError('二维码生成失败，请重试') }
}
async function loadStats() {
  const sequence = ++statsSequence
  if (!overview.value?.agent.can_view_promotion_stats) {
    analytics.value = null
    statsError.value = false
    statsLoading.value = false
    return
  }
  statsLoading.value = true; statsError.value = false
  try {
    const result = await getPromotionAnalytics({ ...dateParams.value })
    if (sequence !== statsSequence) return
    analytics.value = result
  } catch {
    if (sequence !== statsSequence) return
    analytics.value = null
    statsError.value = true
  } finally {
    if (sequence === statsSequence) statsLoading.value = false
  }
}
async function loadVisits() {
  const sequence = ++visitsSequence
  if (!overview.value?.agent.can_view_promotion_stats) {
    visits.value = []
    visitsTotal.value = 0
    visitsPage.value = 1
    visitsError.value = false
    visitsLoadMoreError.value = false
    visitsLoading.value = false
    return
  }
  visitsLoading.value = true; visitsError.value = false; visitsLoadMoreError.value = false
  try {
    const recent = await listPromotionVisits({ ...dateParams.value, page: 1, page_size: 8 })
    if (sequence !== visitsSequence) return
    visits.value = recent.items
    visitsTotal.value = recent.total
    visitsPage.value = 1
  } catch {
    if (sequence !== visitsSequence) return
    visits.value = []
    visitsTotal.value = 0
    visitsError.value = true
  } finally {
    if (sequence === visitsSequence) visitsLoading.value = false
  }
}
async function loadMoreVisits() {
  if (visitsLoading.value || visits.value.length >= visitsTotal.value) return
  const sequence = ++visitsSequence
  visitsLoading.value = true
  visitsLoadMoreError.value = false
  const nextPage = visitsPage.value + 1
  try {
    const recent = await listPromotionVisits({ ...dateParams.value, page: nextPage, page_size: 8 })
    if (sequence !== visitsSequence) return
    const known = new Set(visits.value.map((item) => item.id))
    visits.value.push(...recent.items.filter((item) => !known.has(item.id)))
    visitsTotal.value = recent.total
    visitsPage.value = nextPage
  } catch {
    if (sequence === visitsSequence) visitsLoadMoreError.value = true
  } finally {
    if (sequence === visitsSequence) visitsLoading.value = false
  }
}
async function load() {
  const sequence = ++overviewSequence
  statsSequence += 1
  visitsSequence += 1
  loading.value = true
  overviewError.value = false
  try {
    const result = await getDistributionOverview()
    if (sequence !== overviewSequence) return
    overview.value = result
    void renderQr()
    if (sequence !== overviewSequence) return
    await Promise.all([loadStats(), loadVisits()])
  } catch {
    if (sequence !== overviewSequence) return
    statsSequence += 1
    visitsSequence += 1
    overview.value = null
    analytics.value = null
    visits.value = []
    visitsTotal.value = 0
    statsLoading.value = false
    visitsLoading.value = false
    overviewError.value = true
  } finally {
    if (sequence === overviewSequence) loading.value = false
  }
}
async function setDays(value: number) {
  if (days.value === value || statsLoading.value || visitsLoading.value) return
  days.value = value
  await Promise.all([loadStats(), loadVisits()])
}
async function copyText(value: string, message: string) {
  try { await navigator.clipboard.writeText(value); app.showSuccess(message) }
  catch { app.showError('复制失败，请手动选择内容复制') }
}
function downloadQr() { if (!qrCanvas.value) return; const anchor = document.createElement('a'); anchor.download = `agent-${overview.value?.agent.promotion_code || 'promotion'}.png`; anchor.href = qrCanvas.value.toDataURL('image/png'); anchor.click() }
function chartLabel(value: string) { return days.value === 90 ? `${value.slice(5)}周` : value.slice(5).replace('-', '/') }
function parseLocalDate(value: string) { const [year, month, day] = value.split('-').map(Number); return new Date(year, month - 1, day) }
function formatLocalDate(value: Date) { const year = value.getFullYear(); const month = String(value.getMonth() + 1).padStart(2, '0'); const day = String(value.getDate()).padStart(2, '0'); return `${year}-${month}-${day}` }
function hongKongDate(offset = 0) { const date = new Date(Date.now() + offset * 86400000); const parts = new Intl.DateTimeFormat('en-US', { timeZone: 'Asia/Hong_Kong', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(date); const values = Object.fromEntries(parts.map((part) => [part.type, part.value])); return `${values.year}-${values.month}-${values.day}` }
function formatDateTime(value: string) { return new Date(value).toLocaleString('zh-CN', { timeZone: 'Asia/Hong_Kong', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }) }
function attributionLabel(value?: string) { return value === 'persisted' ? '历史归因' : '直接归因' }
function deviceLabel(value: string) { return ({ mobile: '手机', tablet: '平板', desktop: '桌面端', bot: '机器人', unknown: '未知设备' } as Record<string, string>)[value] || '未知设备' }
onMounted(load)
onBeforeUnmount(() => {
  overviewSequence += 1
  statsSequence += 1
  visitsSequence += 1
})
</script>
