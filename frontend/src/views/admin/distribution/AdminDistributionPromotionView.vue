<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5" :aria-busy="loading || visitsLoading">
      <header class="flex min-w-0 items-start justify-between gap-3">
        <div class="min-w-0">
          <h1 class="page-title">推广追踪</h1>
          <p class="page-description">链接访问、注册归因与去标识化转化分析</p>
        </div>
        <button class="btn btn-secondary btn-icon min-h-11 min-w-11 shrink-0" :disabled="loading || visitsLoading" title="刷新数据" aria-label="刷新当前页数据" @click="refreshSnapshot">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </header>

      <section class="card p-4 sm:p-5" aria-labelledby="promotion-filters-title">
        <div class="flex items-center justify-between gap-3">
          <div class="min-w-0">
            <h2 id="promotion-filters-title" class="text-sm font-semibold">分析范围</h2>
            <p class="mt-1 text-xs text-foreground-subtle">修改条件后统一应用，避免不同口径的数据混合显示。</p>
          </div>
          <button class="btn btn-secondary min-h-11 shrink-0 whitespace-nowrap sm:hidden" type="button" @click="openMobileFilters">
            <Icon name="filter" size="sm" />筛选
          </button>
        </div>

        <div class="mt-4 hidden sm:block">
          <div class="grid items-end gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <FilterFields :model="filterDraft" @update:model="updateDesktopDraft" />
          </div>
          <div class="mt-4 grid items-end gap-3 border-t border-outline pt-4 sm:grid-cols-2 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto_auto]">
            <label class="form-field"><span class="form-label">开始日期</span><input ref="desktopDateFromRef" v-model="filterDraft.date_from" class="input" type="date" :aria-invalid="!!dateRangeError" :aria-describedby="dateRangeError ? 'promotion-date-error' : undefined" @input="clearDateRangeError" /></label>
            <label class="form-field"><span class="form-label">结束日期</span><input ref="desktopDateToRef" v-model="filterDraft.date_to" class="input" type="date" :aria-invalid="!!dateRangeError" :aria-describedby="dateRangeError ? 'promotion-date-error' : undefined" @input="clearDateRangeError" /></label>
            <button class="btn btn-secondary" type="button" :disabled="loading || (!hasFilters && !draftChanged)" @click="resetFilters">重置</button>
            <button class="btn btn-primary" type="button" :disabled="loading || !draftChanged" @click="applyDesktopFilters">应用筛选</button>
          </div>
          <p v-if="dateRangeError" id="promotion-date-error" class="mt-2 text-sm text-danger-foreground" role="alert">{{ dateRangeError }}</p>
        </div>

        <div class="mt-3 flex flex-wrap items-center gap-2 sm:hidden" aria-label="当前筛选条件">
          <span v-for="chip in activeFilterChips" :key="chip" class="badge badge-gray max-w-full break-all">{{ chip }}</span>
          <button v-if="hasFilters" class="min-h-11 px-2 text-xs text-foreground-subtle underline" type="button" @click="resetFilters">清除筛选</button>
        </div>

        <div class="mt-4 flex flex-wrap items-center gap-2 border-t border-outline pt-4" role="group" aria-label="快捷统计周期">
          <span class="text-xs text-foreground-subtle">快捷周期</span>
          <button
            v-for="option in periods"
            :key="option.value"
            type="button"
            class="min-h-11 rounded-control border px-3 text-xs font-medium transition-colors"
            :class="activePreset === option.value ? 'border-brand bg-brand text-brand-foreground' : 'border-outline text-foreground-muted hover:bg-surface-subtle'"
            :aria-pressed="activePreset === option.value"
            :disabled="loading"
            @click="setDays(option.value)"
          >{{ option.label }}</button>
          <span class="ml-auto text-xs text-foreground-subtle">{{ periodLabel }} · 香港时间</span>
        </div>
      </section>

      <section v-if="analytics?.meta" class="rounded-panel border border-outline bg-surface-subtle px-4 py-3" aria-label="推广追踪运行状态">
        <div class="flex flex-wrap items-center gap-2 text-xs">
          <span class="font-medium text-foreground">系统状态</span>
          <span class="badge" :class="analytics.meta.tracking_enabled ? 'badge-success' : 'badge-warning'">{{ analytics.meta.tracking_enabled ? '正在采集' : '已停止采集' }}</span>
          <span class="badge" :class="analytics.meta.attribution_enabled ? 'badge-success' : 'badge-gray'">{{ analytics.meta.attribution_enabled ? `${attributionModelLabel(analytics.meta.attribution_model)} · 归因窗口 ${analytics.meta.attribution_days} 天` : '历史归因关闭' }}</span>
          <span class="badge badge-gray">{{ analytics.meta.bot_filter_enabled ? '汇总排除机器人' : '汇总包含机器人' }}</span>
          <span class="badge badge-gray">明细保留 {{ analytics.meta.detail_retention_days }} 天</span>
          <span class="badge badge-gray">原始记录 {{ analytics.meta.raw_retention_days }} 天后删除</span>
          <RouterLink to="/admin/settings" class="ml-auto inline-flex min-h-11 items-center py-2 font-medium text-brand hover:underline">调整设置</RouterLink>
        </div>
      </section>

      <section class="rounded-panel border border-outline bg-surface-subtle px-4 py-3 text-xs text-foreground-subtle" aria-label="指标口径说明">
        <div class="flex items-start gap-2">
          <Icon name="infoCircle" size="sm" class="mt-0.5 shrink-0" />
          <p><strong class="font-medium text-foreground">口径说明：</strong>筛选采用访问批次口径，注册计入其获归因的推广访问发生日。访问次数按一分钟内同一访客与代理去重；独立访客按去标识化访客标识去重。“尚未注册”表示该代理下的访客当前没有关联注册。直接注册表示注册请求携带推广码，历史归因表示注册时使用有效历史 Cookie。访问不是广告展示次数。</p>
        </div>
      </section>

      <section v-if="analyticsError" class="rounded-panel border border-danger/30 bg-danger-subtle p-4" role="alert">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div><p class="text-sm font-medium text-danger-foreground">推广概览加载失败</p><p class="mt-1 text-xs text-foreground-subtle">{{ analytics ? '当前保留最近一次成功概览，可能与现有筛选条件不同。' : '当前没有可展示的成功概览，请重试或调整分析范围。' }}</p></div>
          <button class="btn btn-secondary min-h-11" type="button" :disabled="loading" @click="refreshSnapshot">重试</button>
        </div>
      </section>
       <div v-if="loading && !analytics" class="flex min-h-40 items-center justify-center" role="status" aria-label="正在加载推广数据"><LoadingSpinner /></div>

      <template v-if="analytics">
        <section class="grid grid-cols-2 gap-3 xl:grid-cols-4" aria-label="推广核心指标">
          <article v-for="card in cards" :key="card.label" class="card min-w-0 p-4">
            <p class="text-xs text-foreground-subtle">{{ card.label }}</p>
            <p class="mt-2 break-words text-2xl font-semibold tabular-nums">{{ card.value }}</p>
            <p class="mt-1 text-xs leading-5 text-foreground-muted">{{ card.hint }}</p>
          </article>
        </section>

        <section class="border-y border-outline py-4" aria-labelledby="admin-registration-results-title">
          <div class="flex flex-wrap items-end justify-between gap-2">
            <div><h2 id="admin-registration-results-title" class="font-medium">访问批次的注册结果</h2><p class="mt-1 text-xs text-foreground-subtle">归属于当前访问人群，按获归因的推广访问日期统计</p></div>
            <span class="text-xs text-foreground-muted">{{ periodLabel }}</span>
          </div>
          <dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-3 sm:grid-cols-4">
            <div v-for="item in registrationResults" :key="item.label" class="min-w-0"><dt class="text-xs text-foreground-subtle">{{ item.label }}</dt><dd class="mt-1 break-words text-lg font-semibold tabular-nums">{{ item.value }}</dd></div>
          </dl>
          <p v-if="analytics.summary.untracked_direct > 0" class="mt-4 rounded-panel border border-warning/30 bg-warning-subtle px-3 py-2 text-xs leading-5 text-warning-foreground">{{ analytics.summary.untracked_direct.toLocaleString() }} 人携推广码注册但未匹配到有效访问，因此计入总注册，不计入访客转化率。</p>
        </section>

        <section class="grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(280px,0.6fr)]">
          <div class="border-y border-outline py-4">
            <div class="mb-4 flex items-center justify-between gap-3">
              <div><h2 class="font-medium">访问量与归因注册</h2><p class="mt-1 text-xs text-foreground-subtle">{{ trendUnit }}聚合 · 同一访问批次口径，注册显示在获归因的推广访问日期</p></div>
              <span class="text-xs text-foreground-subtle">{{ periodLabel }}</span>
            </div>
            <div class="mb-3 flex flex-wrap items-center gap-4 text-xs text-foreground-subtle" aria-hidden="true"><span class="inline-flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-info"></span>访问</span><span class="inline-flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-success"></span>注册</span></div>
            <div class="max-h-72 space-y-2 overflow-y-auto pr-1" role="list" aria-label="访问与注册趋势">
              <div v-for="item in trend" :key="item.date" class="grid grid-cols-[52px_minmax(0,1fr)] items-center gap-x-2 gap-y-1 text-xs sm:grid-cols-[68px_minmax(0,1fr)_minmax(64px,auto)]" role="listitem" :aria-label="`${formatTrendLabel(item.date)}，${item.visits} 次访问，${item.registrations} 次注册`">
                <span class="text-foreground-subtle">{{ formatTrendLabel(item.date) }}</span>
                <div class="space-y-1" aria-hidden="true"><div class="h-1.5 overflow-hidden rounded bg-surface-subtle"><div class="h-full rounded bg-info" :style="{ width: `${Math.max(item.visits ? 2 : 0, item.visits / maxTrendValue * 100)}%` }"></div></div><div class="h-1.5 overflow-hidden rounded bg-surface-subtle"><div class="h-full rounded bg-success" :style="{ width: `${Math.max(item.registrations ? 2 : 0, item.registrations / maxTrendValue * 100)}%` }"></div></div></div>
                <span class="col-start-2 min-w-0 break-words text-right tabular-nums sm:col-start-auto" aria-hidden="true">访问 {{ item.visits }} · 注册 {{ item.registrations }}</span>
              </div>
              <p v-if="!trend.length" class="py-8 text-center text-sm text-foreground-subtle">当前范围暂无趋势数据</p>
            </div>
          </div>
          <div class="border-y border-outline py-4">
            <div class="flex items-center justify-between gap-2"><h2 class="font-medium">访问来源排行</h2><span class="text-xs text-foreground-subtle">访问 / 已转化访客 / 访问转化率</span></div>
            <div class="mt-4 space-y-3">
              <div v-for="source in analytics.sources" :key="source.source" class="flex items-start justify-between gap-3 text-sm"><button type="button" class="min-h-11 min-w-11 break-words rounded-control pr-2 text-left font-medium text-brand hover:underline focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/40" :aria-label="`筛选来源 ${source.source}`" @click="filterBySource(source.source)">{{ source.source }}</button><span class="shrink-0 py-3 text-right tabular-nums">{{ source.visits }} / {{ source.conversions }}<small class="ml-2 text-foreground-muted">{{ sourceConversion(source.visits, source.conversions) }}</small></span></div>
              <div v-if="!analytics.sources.length" class="py-8 text-center text-sm text-foreground-subtle"><p>当前筛选条件下暂无来源数据</p><button v-if="hasFilters" class="btn btn-secondary mt-3 min-h-11" type="button" @click="resetFilters">清除筛选</button></div>
            </div>
          </div>
        </section>
      </template>

      <section ref="visitsSectionRef" aria-labelledby="promotion-visits-title">
        <div class="mb-3 flex items-end justify-between gap-3">
          <div><h2 id="promotion-visits-title" class="font-medium">访问明细</h2><p class="mt-1 text-xs text-foreground-subtle">按代理下的去标识化访客展示注册状态；时间均为香港时间</p></div>
          <span class="shrink-0 text-xs text-foreground-muted">{{ visitsLoading && visits.length ? '正在更新 · ' : '' }}共 {{ total.toLocaleString() }} 条</span>
        </div>
        <div v-if="visitsError" class="mb-3 flex flex-wrap items-center justify-between gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4" role="alert"><div><p class="text-sm font-medium text-danger-foreground">访问明细加载失败</p><p class="mt-1 text-xs text-foreground-subtle">{{ visits.length ? '当前保留最近一次成功明细，可能与现有筛选条件不同。' : '当前没有可展示的成功明细。' }}</p></div><button class="btn btn-secondary min-h-11" type="button" :disabled="visitsLoading" @click="refreshSnapshot">重试</button></div>
        <div v-if="visitsLoading && !visits.length" class="flex min-h-32 items-center justify-center" role="status" aria-label="正在加载访问明细"><LoadingSpinner /></div>
        <template v-else>
          <div v-if="visits.length" class="hidden overflow-x-auto rounded-panel border border-outline sm:block">
            <table class="w-full min-w-[920px] text-left text-sm"><caption class="sr-only">推广链接去标识化访问明细</caption><thead class="bg-surface-subtle text-xs text-foreground-subtle"><tr><th scope="col" class="px-4 py-3 font-medium">访问时间</th><th scope="col" class="px-4 py-3 font-medium">代理</th><th scope="col" class="px-4 py-3 font-medium">来源</th><th scope="col" class="px-4 py-3 font-medium">设备</th><th scope="col" class="px-4 py-3 font-medium">访客注册结果</th></tr></thead>
              <tbody class="divide-y divide-outline"><tr v-for="visit in visits" :key="visit.id"><td class="px-4 py-3 text-xs text-foreground-subtle">{{ formatHongKongDate(visit.visited_at) }}</td><td class="px-4 py-3 text-xs">{{ visit.agent_email || `代理 #${visit.agent_id}` }}</td><td class="max-w-56 break-words px-4 py-3">{{ visit.source }}</td><td class="px-4 py-3 text-xs text-foreground-subtle">{{ deviceLabel(visit.device_type) }}</td><td class="px-4 py-3"><span class="badge" :class="visit.registered_at ? 'badge-success' : 'badge-gray'">{{ visit.registered_at ? attributionLabel(visit.attribution_type) : '尚未注册' }}</span><p v-if="visit.registered_at" class="mt-1 text-xs text-foreground-subtle">{{ formatHongKongDate(visit.registered_at) }} · {{ conversionDuration(visit) }}</p></td></tr></tbody>
            </table>
          </div>
          <div v-if="visits.length" class="grid gap-2 sm:hidden">
            <article v-for="visit in visits" :key="visit.id" class="rounded-panel border border-outline p-3">
              <div class="flex items-center justify-between gap-2"><span class="badge" :class="visit.registered_at ? 'badge-success' : 'badge-gray'">{{ visit.registered_at ? attributionLabel(visit.attribution_type) : '尚未注册' }}</span><span class="text-xs text-foreground-subtle">{{ formatHongKongDate(visit.visited_at) }}</span></div>
              <dl class="mt-3 grid grid-cols-[auto_minmax(0,1fr)] gap-x-3 gap-y-1 text-xs"><dt class="text-foreground-subtle">代理</dt><dd class="break-all text-right">{{ visit.agent_email || `代理 #${visit.agent_id}` }}</dd><dt class="text-foreground-subtle">来源</dt><dd class="break-words text-right">{{ visit.source }}</dd><dt class="text-foreground-subtle">设备</dt><dd class="text-right">{{ deviceLabel(visit.device_type) }}</dd><template v-if="visit.registered_at"><dt class="text-foreground-subtle">注册</dt><dd class="text-right">{{ formatHongKongDate(visit.registered_at) }}</dd><dt class="text-foreground-subtle">转化耗时</dt><dd class="text-right">{{ conversionDuration(visit) }}</dd></template></dl>
            </article>
          </div>
          <div v-else-if="!visitsError" class="rounded-panel border border-dashed border-outline p-8 text-center text-sm text-foreground-subtle"><p>当前范围暂无访问</p><button v-if="hasFilters" class="btn btn-secondary mt-3 min-h-11" type="button" @click="resetFilters">清除筛选</button></div>
          <nav v-if="total > pagination.page_size" class="mt-4 flex items-center justify-between gap-3" aria-label="访问明细分页"><span class="text-xs text-foreground-subtle" aria-live="polite">第 {{ pagination.page }} / {{ pageCount }} 页</span><div class="flex gap-2"><button class="btn btn-secondary min-h-11" type="button" :disabled="pagination.page <= 1 || visitsLoading" @click="changePage(-1)">上一页</button><button class="btn btn-secondary min-h-11" type="button" :disabled="pagination.page >= pageCount || visitsLoading" @click="changePage(1)">下一页</button></div></nav>
        </template>
      </section>

      <p class="sr-only" role="status" aria-live="polite">{{ loading || visitsLoading ? '正在更新推广数据' : '' }}</p>
      <p v-if="lastUpdated" class="text-right text-xs text-foreground-muted" aria-live="polite">数据更新于 {{ lastUpdated }}</p>
    </div>

    <BaseDialog :show="mobileFilters" title="筛选推广数据" width="normal" @close="mobileFilters = false">
      <div class="space-y-4">
        <FilterFields :model="mobileFilterDraft" :stacked="true" @update:model="updateMobileFilterDraft" />
        <div class="grid grid-cols-1 gap-3 min-[380px]:grid-cols-2"><label class="form-field min-w-0"><span class="form-label">开始日期</span><input ref="mobileDateFromRef" v-model="mobileFilterDraft.date_from" class="input min-h-11 min-w-0" type="date" :aria-invalid="!!dateRangeError" :aria-describedby="dateRangeError ? 'promotion-mobile-date-error' : undefined" @input="clearDateRangeError" /></label><label class="form-field min-w-0"><span class="form-label">结束日期</span><input ref="mobileDateToRef" v-model="mobileFilterDraft.date_to" class="input min-h-11 min-w-0" type="date" :aria-invalid="!!dateRangeError" :aria-describedby="dateRangeError ? 'promotion-mobile-date-error' : undefined" @input="clearDateRangeError" /></label></div>
        <p v-if="dateRangeError" id="promotion-mobile-date-error" class="text-sm text-danger-foreground" role="alert">{{ dateRangeError }}</p>
      </div>
      <template #footer><div class="grid w-full grid-cols-2 gap-2"><button class="btn btn-secondary min-h-11 min-w-0" type="button" @click="resetMobileDraft">重置</button><button class="btn btn-primary min-h-11 min-w-0" type="button" :disabled="loading || !mobileDraftChanged" @click="applyMobileFilters">应用筛选</button></div></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import RemoteEntityCombobox from '@/components/admin/distribution/RemoteEntityCombobox.vue'
import type { DistributionPickerOption } from '@/components/admin/distribution/types'
import { getPromotionAnalytics, listPromotionVisits, lookupAgents, type DistributionPromotionAnalytics, type DistributionPromotionVisit } from '@/api/admin/distribution'
import { useAppStore } from '@/stores/app'

type FilterState = { agent: DistributionPickerOption | null; source: string; device: string; attribution_type: string; date_from: string; date_to: string }
const HONG_KONG_TIME_ZONE = 'Asia/Hong_Kong'
const periods = [{ label: '7天', value: 7 }, { label: '30天', value: 30 }, { label: '90天', value: 90 }]
const app = useAppStore()
const analytics = ref<DistributionPromotionAnalytics | null>(null)
const visits = ref<DistributionPromotionVisit[]>([])
const total = ref(0)
const loading = ref(false)
const visitsLoading = ref(false)
const analyticsError = ref(false)
const visitsError = ref(false)
const dateRangeError = ref('')
const lastUpdated = ref('')
const mobileFilters = ref(false)
const activePreset = ref<number | null>(30)
const pagination = reactive({ page: 1, page_size: 20 })
const desktopDateFromRef = ref<HTMLInputElement | null>(null)
const desktopDateToRef = ref<HTMLInputElement | null>(null)
const mobileDateFromRef = ref<HTMLInputElement | null>(null)
const mobileDateToRef = ref<HTMLInputElement | null>(null)
const visitsSectionRef = ref<HTMLElement | null>(null)
let snapshotSequence = 0
let detailSequence = 0

function hongKongDate(offset = 0) {
  const date = new Date(Date.now() + offset * 86400000)
  const parts = new Intl.DateTimeFormat('en-US', { timeZone: HONG_KONG_TIME_ZONE, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(date)
  const value = Object.fromEntries(parts.map((part) => [part.type, part.value]))
  return `${value.year}-${value.month}-${value.day}`
}
function defaultState(): FilterState { return { agent: null, source: '', device: '', attribution_type: '', date_from: hongKongDate(-29), date_to: hongKongDate() } }
const filters = reactive<FilterState>(defaultState())
const filterDraft = reactive<FilterState>(defaultState())
const mobileFilterDraft = reactive<FilterState>(defaultState())
const params = computed(() => ({ agent_id: filters.agent?.id, source: filters.source || undefined, device: filters.device || undefined, attribution_type: filters.attribution_type || undefined, date_from: filters.date_from, date_to: filters.date_to }))
const periodLabel = computed(() => activePreset.value ? `最近 ${activePreset.value} 天` : `${filters.date_from} 至 ${filters.date_to}`)
const hasFilters = computed(() => !!filters.agent || !!filters.source || !!filters.device || !!filters.attribution_type || activePreset.value !== 30)
const draftChanged = computed(() => JSON.stringify(filterDraft) !== JSON.stringify(filters))
const mobileDraftChanged = computed(() => JSON.stringify(mobileFilterDraft) !== JSON.stringify(filters))
const pageCount = computed(() => Math.max(1, Math.ceil(total.value / pagination.page_size)))
const activeFilterChips = computed(() => [periodLabel.value, filters.agent ? `代理：${filters.agent.email}` : '', filters.source ? `来源：${filters.source}` : '', filters.device ? `设备：${deviceLabel(filters.device)}` : '', filters.attribution_type ? `状态：${attributionLabel(filters.attribution_type)}` : ''].filter(Boolean))
const cards = computed(() => analytics.value ? [
  { label: '有效访问', value: analytics.value.summary.total_visits.toLocaleString(), hint: '一分钟内重复打开已去重' },
  { label: '独立访客', value: analytics.value.summary.unique_visitors.toLocaleString(), hint: '按去标识化访客标识去重' },
  { label: '已转化访客', value: analytics.value.summary.converted_visitors.toLocaleString(), hint: '当前访问人群中已完成注册' },
  { label: '访客转化率', value: `${analytics.value.summary.conversion_rate.toFixed(2)}%`, hint: `已转化访客 ÷ 独立访客；${botPolicyHint(analytics.value)}` },
] : [])
const registrationResults = computed(() => analytics.value ? [
  { label: '总注册', value: analytics.value.summary.registrations.toLocaleString() },
  { label: '直接携码', value: analytics.value.summary.direct_registrations.toLocaleString() },
  { label: '历史归因', value: analytics.value.summary.persisted_registrations.toLocaleString() },
  { label: '未匹配有效访问', value: analytics.value.summary.untracked_direct.toLocaleString() },
] : [])
const selectedRangeDays = computed(() => Math.floor((Date.parse(`${filters.date_to}T00:00:00+08:00`) - Date.parse(`${filters.date_from}T00:00:00+08:00`)) / 86400000) + 1)
const trendUsesWeeks = computed(() => selectedRangeDays.value > 31)
const trendUnit = computed(() => trendUsesWeeks.value ? '按周' : '按日')
const trend = computed(() => {
  const daily = [...(analytics.value?.daily || [])].sort((left, right) => right.date.localeCompare(left.date))
  if (!trendUsesWeeks.value) return daily
  const grouped = new Map<string, { date: string; visits: number; registrations: number }>()
  for (const item of daily) { const date = parseDate(item.date); const day = date.getDay() || 7; date.setDate(date.getDate() - day + 1); const key = formatDateOnly(date); const row = grouped.get(key) || { date: key, visits: 0, registrations: 0 }; row.visits += item.visits; row.registrations += item.registrations; grouped.set(key, row) }
  return [...grouped.values()]
})
const maxTrendValue = computed(() => Math.max(1, ...trend.value.flatMap((item) => [item.visits, item.registrations])))

function copyState(target: FilterState, source: FilterState) { Object.assign(target, source) }
function parseDate(value: string) { const [year, month, day] = value.split('-').map(Number); return new Date(year, month - 1, day) }
function formatDateOnly(value: Date) { return `${value.getFullYear()}-${String(value.getMonth() + 1).padStart(2, '0')}-${String(value.getDate()).padStart(2, '0')}` }
function presetForRange(state: FilterState) { return periods.find((option) => state.date_to === hongKongDate() && state.date_from === hongKongDate(-(option.value - 1)))?.value ?? null }
function clearDateRangeError() { dateRangeError.value = '' }
function focusDateField(target: 'from' | 'to') {
  const input = mobileFilters.value
    ? (target === 'from' ? mobileDateFromRef.value : mobileDateToRef.value)
    : (target === 'from' ? desktopDateFromRef.value : desktopDateToRef.value)
  void nextTick(() => input?.focus())
}
function validateRange(state: FilterState) {
  const from = Date.parse(`${state.date_from}T00:00:00+08:00`)
  const to = Date.parse(`${state.date_to}T00:00:00+08:00`)
  if (!state.date_from || !Number.isFinite(from)) { dateRangeError.value = '请选择有效的开始日期。'; focusDateField('from'); return false }
  if (!state.date_to || !Number.isFinite(to)) { dateRangeError.value = '请选择有效的结束日期。'; focusDateField('to'); return false }
  if (from > to) { dateRangeError.value = '结束日期不能早于开始日期。'; focusDateField('to'); return false }
  if (to - from > 365 * 86400000) { dateRangeError.value = '统计范围不能超过 366 天。'; focusDateField('to'); return false }
  dateRangeError.value = ''
  return true
}
function applyState(source: FilterState) { if (!validateRange(source)) return false; copyState(filters, source); copyState(filterDraft, source); copyState(mobileFilterDraft, source); activePreset.value = presetForRange(source); pagination.page = 1; void loadSnapshot(1); return true }
function updateDesktopDraft(value: Partial<FilterState>) { Object.assign(filterDraft, value) }
function updateMobileFilterDraft(value: Partial<FilterState>) { Object.assign(mobileFilterDraft, value) }
function applyDesktopFilters() { applyState(filterDraft) }
function openMobileFilters() { clearDateRangeError(); copyState(mobileFilterDraft, filters); mobileFilters.value = true }
function applyMobileFilters() { if (applyState(mobileFilterDraft)) mobileFilters.value = false }
function resetMobileDraft() { clearDateRangeError(); copyState(mobileFilterDraft, defaultState()) }
function resetFilters() { clearDateRangeError(); const next = defaultState(); copyState(filters, next); copyState(filterDraft, next); copyState(mobileFilterDraft, next); activePreset.value = 30; pagination.page = 1; void loadSnapshot(1) }
function setDays(value: number) { clearDateRangeError(); const next = { ...filters, date_from: hongKongDate(-(value - 1)), date_to: hongKongDate() }; copyState(filters, next); copyState(filterDraft, next); copyState(mobileFilterDraft, next); activePreset.value = value; pagination.page = 1; void loadSnapshot(1) }
function filterBySource(source: string) { applyState({ ...filters, source }) }

function agentStatusLabel(status: string) { return ({ active: '启用中', suspended: '已暂停', revoked: '已撤销' } as Record<string, string>)[status] || status }
async function searchAgents(query: string): Promise<DistributionPickerOption[]> { const result = await lookupAgents(query, { include_inactive: true }); return result.map((agent) => ({ id: agent.agent_id, email: agent.email, username: agent.username, meta: `${agent.depth === 1 ? '一级代理' : '下级代理'} · ${agentStatusLabel(agent.status)} · ${agent.promotion_code}` })) }
function refreshSnapshot() { void loadSnapshot(pagination.page) }
async function loadSnapshot(targetPage = 1) {
  const sequence = ++snapshotSequence
  detailSequence += 1
  loading.value = true; visitsLoading.value = true; analyticsError.value = false; visitsError.value = false
  const query = { ...params.value }
  const [analyticsResult, visitsResult] = await Promise.allSettled([
      getPromotionAnalytics(query),
      listPromotionVisits({ ...query, page: targetPage, page_size: pagination.page_size }),
    ])
  if (sequence !== snapshotSequence) return
  if (analyticsResult.status === 'fulfilled') analytics.value = analyticsResult.value
  else analyticsError.value = true
  if (visitsResult.status === 'fulfilled') {
    const initialVisits = visitsResult.value
    const lastPage = Math.max(1, Math.ceil(initialVisits.total / pagination.page_size))
    const resolvedPage = Math.min(Math.max(1, targetPage), lastPage)
    try {
      const nextVisits = resolvedPage === targetPage ? initialVisits : await listPromotionVisits({ ...query, page: resolvedPage, page_size: pagination.page_size })
      if (sequence !== snapshotSequence) return
      visits.value = nextVisits.items; total.value = nextVisits.total; pagination.page = resolvedPage
    } catch { if (sequence === snapshotSequence) visitsError.value = true }
  } else visitsError.value = true
  if (analyticsResult.status === 'fulfilled' || visitsResult.status === 'fulfilled') lastUpdated.value = formatHongKongDate(new Date().toISOString())
  if (sequence === snapshotSequence) { loading.value = false; visitsLoading.value = false }
}
async function changePage(delta: number) {
  const requestedPage = pagination.page + delta
  const sequence = ++detailSequence
  visitsLoading.value = true
  try {
    const query = { ...params.value }
    const initial = await listPromotionVisits({ ...query, page: requestedPage, page_size: pagination.page_size })
    if (sequence !== detailSequence) return
    const lastPage = Math.max(1, Math.ceil(initial.total / pagination.page_size))
    const resolvedPage = Math.min(Math.max(1, requestedPage), lastPage)
    const result = resolvedPage === requestedPage
      ? initial
      : await listPromotionVisits({ ...query, page: resolvedPage, page_size: pagination.page_size })
    if (sequence !== detailSequence) return
    visits.value = result.items
    total.value = result.total
    pagination.page = resolvedPage
    await nextTick()
    visitsSectionRef.value?.scrollIntoView?.({ behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth', block: 'start' })
  } catch {
    if (sequence === detailSequence) app.showError('访问明细加载失败，请重试')
  } finally {
    if (sequence === detailSequence) visitsLoading.value = false
  }
}
function formatHongKongDate(value: string) { return new Date(value).toLocaleString('zh-CN', { timeZone: HONG_KONG_TIME_ZONE, month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }) }
function formatTrendLabel(value: string) { return value.slice(5).replace('-', '/') }
function deviceLabel(value: string) { return ({ mobile: '手机', tablet: '平板', desktop: '桌面端', bot: '机器人', unknown: '未知设备' } as Record<string, string>)[value] || '未知设备' }
function attributionLabel(value?: string) { return value === 'persisted' ? '历史归因注册' : value === 'direct' ? '直接注册' : '尚未注册' }
function attributionModelLabel(value?: string) { return value === 'last_touch' ? '最近访问归因' : '首次访问归因' }
function botPolicyHint(value: DistributionPromotionAnalytics) { if (value.meta?.bot_filter_enabled === true) return `已排除 ${value.summary.bot_visits.toLocaleString()} 次机器人访问`; if (value.meta?.bot_filter_enabled === false) return `包含机器人；识别到 ${value.summary.bot_visits.toLocaleString()} 次`; return `识别到 ${value.summary.bot_visits.toLocaleString()} 次机器人访问，过滤规则以平台设置为准` }
function conversionDuration(visit: DistributionPromotionVisit) { if (!visit.registered_at) return ''; const minutes = Math.max(0, Math.round((new Date(visit.registered_at).getTime() - new Date(visit.visited_at).getTime()) / 60000)); if (minutes < 1) return '不足 1 分钟'; if (minutes < 60) return `${minutes} 分钟`; if (minutes < 1440) return `${Math.floor(minutes / 60)} 小时 ${minutes % 60} 分钟`; return `${Math.floor(minutes / 1440)} 天` }
function sourceConversion(visitsCount: number, registrations: number) { return visitsCount > 0 ? `${(registrations / visitsCount * 100).toFixed(1)}%` : '0.0%' }

const FilterFields = defineComponent({ props: { model: { type: Object, required: true }, stacked: Boolean }, emits: ['update:model'], setup(props, { emit }) { const update = (key: keyof FilterState, value: unknown) => emit('update:model', { ...props.model, [key]: value }); const controlClass = props.stacked ? 'input min-h-11 min-w-0' : 'input min-w-0'; return () => h('div', { class: props.stacked ? 'grid min-w-0 gap-4' : 'contents' }, [h(RemoteEntityCombobox, { class: 'min-w-0', modelValue: (props.model as FilterState).agent, label: '代理（含已暂停和已撤销）', placeholder: '输入代理邮箱或推广码', inputId: props.stacked ? 'promotion-agent-mobile' : 'promotion-agent-desktop', search: searchAgents, 'onUpdate:modelValue': (value: DistributionPickerOption | null) => update('agent', value) }), h('label', { class: 'form-field min-w-0' }, [h('span', { class: 'form-label' }, '来源（精确匹配）'), h('input', { class: controlClass, value: (props.model as FilterState).source, maxlength: 128, placeholder: '不区分大小写，例如 google', onInput: (event: Event) => update('source', (event.target as HTMLInputElement).value) })]), h('label', { class: 'form-field min-w-0' }, [h('span', { class: 'form-label' }, '设备'), h('select', { class: controlClass, value: (props.model as FilterState).device, onChange: (event: Event) => update('device', (event.target as HTMLSelectElement).value) }, [h('option', { value: '' }, '全部设备'), h('option', { value: 'desktop' }, '桌面端'), h('option', { value: 'mobile' }, '手机'), h('option', { value: 'tablet' }, '平板'), h('option', { value: 'bot' }, '机器人')])]), h('label', { class: 'form-field min-w-0' }, [h('span', { class: 'form-label' }, '访客注册状态'), h('select', { class: controlClass, value: (props.model as FilterState).attribution_type, onChange: (event: Event) => update('attribution_type', (event.target as HTMLSelectElement).value) }, [h('option', { value: '' }, '全部状态'), h('option', { value: 'unregistered' }, '尚未注册'), h('option', { value: 'direct' }, '直接注册'), h('option', { value: 'persisted' }, '历史归因注册')])])]) } })

onMounted(() => { void loadSnapshot(1) })
onBeforeUnmount(() => { snapshotSequence += 1; detailSequence += 1 })
</script>
