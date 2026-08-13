<template>
  <section class="business-chart" :aria-label="title">
    <div class="chart-heading">
      <div class="min-w-0">
        <h2 class="text-sm font-semibold text-foreground">{{ title }}</h2>
        <p class="mt-1 text-xs text-foreground-subtle">{{ subtitle }}</p>
      </div>
      <div class="chart-controls">
        <div v-if="showScope" class="chart-control-group">
          <span>{{ t('common.distributionAnalytics.businessScope') }}</span>
          <div class="chart-segments" role="group" :aria-label="t('common.distributionAnalytics.businessScope')">
            <button v-for="item in scopes" :key="item.key" type="button" :aria-pressed="scope === item.key" @click="scope = item.key">{{ item.label }}</button>
          </div>
        </div>
        <div class="chart-control-group">
          <span>{{ t('common.distributionAnalytics.metric') }}</span>
          <div class="chart-segments" role="group" :aria-label="t('common.distributionAnalytics.metric')">
            <button v-for="item in metrics" :key="item.key" type="button" :aria-pressed="metric === item.key" @click="metric = item.key">{{ item.label }}</button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="points.length" class="chart-body">
      <div class="chart-current">
        <span>{{ activeMetricLabel }}</span>
        <strong>{{ formatMetric(totalValue) }}</strong>
      </div>
      <div class="chart-canvas">
        <div class="chart-y-axis" aria-hidden="true">
          <span v-for="tick in yTicks" :key="tick">{{ formatAxis(tick) }}</span>
        </div>
        <div
          ref="plotRef"
          class="chart-plot"
          @pointerenter="onPointerMove"
          @pointermove="onPointerMove"
          @pointerleave="hoveredIndex = null"
        >
          <svg class="chart-svg" viewBox="0 0 1000 200" preserveAspectRatio="none" role="img" :aria-label="`${activeMetricLabel}: ${formatMetric(totalValue)}`">
            <line v-for="tick in yTicks" :key="tick" x1="0" x2="1000" :y1="yAt(tick)" :y2="yAt(tick)" class="chart-grid" />
            <path :d="areaPath" class="chart-area" />
            <path :d="linePath" class="chart-line" />
            <circle v-for="(value, index) in markerValues" :key="index" :cx="xAt(index)" :cy="yAt(value)" r="3" class="chart-marker" />
            <template v-if="hoveredIndex !== null">
              <line :x1="xAt(hoveredIndex)" :x2="xAt(hoveredIndex)" y1="0" y2="200" class="chart-cursor" />
              <circle :cx="xAt(hoveredIndex)" :cy="yAt(values[hoveredIndex] || 0)" r="5" class="chart-point" />
            </template>
          </svg>
          <div v-if="hoveredPoint" class="chart-tooltip" :class="tooltipSide" :style="tooltipStyle">
            <p class="tooltip-title">{{ formatBucketRange(hoveredPoint.date) }}</p>
            <dl>
              <div :class="{ active: metric === 'paid' }"><dt>{{ t('common.distributionAnalytics.customerPaid') }}</dt><dd>{{ money(hoveredPoint.customer_paid_cny) }}</dd></div>
              <div :class="{ active: metric === 'commission' }"><dt>{{ t('common.distributionAnalytics.commission') }}</dt><dd>{{ money(hoveredPoint.commission_cny) }}</dd></div>
              <div :class="{ active: metric === 'new' }"><dt>{{ t('common.distributionAnalytics.newCustomers') }}</dt><dd>{{ hoveredPoint.new_customers }}</dd></div>
              <div :class="{ active: metric === 'paying' }"><dt>{{ t('common.distributionAnalytics.payingCustomers') }}</dt><dd>{{ hoveredPoint.paying_customers }}</dd></div>
            </dl>
          </div>
        </div>
        <div class="chart-x-axis" aria-hidden="true">
          <span v-for="label in xLabels" :key="label.index" :style="{ left: `${label.position}%` }">{{ label.text }}</span>
        </div>
      </div>
    </div>
    <div v-else class="chart-empty">{{ t('common.distributionAnalytics.noTrend') }}</div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DistributionDailyMetric } from '@/api/distribution'

type Resolution = 'hour' | 'day' | 'week' | 'month'
type Scope = 'total' | 'direct' | 'team'
type Metric = 'paid' | 'commission' | 'new' | 'paying'

const props = withDefaults(defineProps<{
  direct: DistributionDailyMetric[]
  team?: DistributionDailyMetric[]
  resolution?: Resolution
  dateFrom?: string
  dateTo?: string
  title?: string
  subtitle?: string
  showScope?: boolean
}>(), { team: () => [], resolution: 'day', dateFrom: '', dateTo: '', title: '', subtitle: '', showScope: true })

const { t } = useI18n()
const plotRef = ref<HTMLElement | null>(null)
const scope = ref<Scope>('total')
const metric = ref<Metric>('paid')
const hoveredIndex = ref<number | null>(null)
const scopes = computed(() => [
  { key: 'total' as const, label: t('common.distributionAnalytics.total') },
  { key: 'direct' as const, label: t('common.distributionAnalytics.direct') },
  { key: 'team' as const, label: t('common.distributionAnalytics.team') },
])
const metrics = computed(() => [
  { key: 'paid' as const, label: t('common.distributionAnalytics.paid') },
  { key: 'commission' as const, label: t('common.distributionAnalytics.commission') },
  { key: 'new' as const, label: t('common.distributionAnalytics.newCustomers') },
  { key: 'paying' as const, label: t('common.distributionAnalytics.payingCustomers') },
])
const activeMetricLabel = computed(() => metrics.value.find(item => item.key === metric.value)?.label || '')

const points = computed(() => props.direct.map((item, index) => {
  const team = props.team[index]
  if (scope.value === 'direct' || !team) return item
  if (scope.value === 'team') return team
  return {
    date: item.date,
    new_customers: item.new_customers + team.new_customers,
    paying_customers: item.paying_customers + team.paying_customers,
    customer_paid_cny: String(Number(item.customer_paid_cny) + Number(team.customer_paid_cny)),
    commission_cny: String(Number(item.commission_cny) + Number(team.commission_cny)),
  }
}))
const values = computed(() => points.value.map(item => metric.value === 'paid'
  ? Number(item.customer_paid_cny)
  : metric.value === 'commission'
    ? Number(item.commission_cny)
    : metric.value === 'new' ? item.new_customers : item.paying_customers))
const totalValue = computed(() => values.value.reduce((sum, value) => sum + value, 0))

function niceStep(maximum: number) {
  if (maximum <= 0) return 1
  const rough = maximum / 4
  const power = 10 ** Math.floor(Math.log10(rough))
  const normalized = rough / power
  return (normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10) * power
}
const tickStep = computed(() => niceStep(Math.max(...values.value, 0)))
const axisMaximum = computed(() => tickStep.value * 4)
const yTicks = computed(() => [4, 3, 2, 1, 0].map(index => index * tickStep.value))
const xAt = (index: number) => points.value.length <= 1 ? 500 : index * (1000 / (points.value.length - 1))
const yAt = (value: number) => 200 - Math.min(Math.max(value / axisMaximum.value, 0), 1) * 188
const linePath = computed(() => values.value.map((value, index) => `${index ? 'L' : 'M'} ${xAt(index)} ${yAt(value)}`).join(' '))
const areaPath = computed(() => points.value.length ? `${linePath.value} L ${xAt(points.value.length - 1)} 200 L ${xAt(0)} 200 Z` : '')
const markerValues = computed(() => values.value.length <= 31 ? values.value : [])
const hoveredPoint = computed(() => hoveredIndex.value === null ? null : points.value[hoveredIndex.value])

const xLabels = computed(() => {
  const count = points.value.length
  if (!count) return []
  const labels = Math.min(count, 6)
  const indexes = Array.from({ length: labels }, (_, index) => labels === 1 ? 0 : Math.round(index * (count - 1) / (labels - 1)))
  return [...new Set(indexes)].map(index => ({ index, position: count === 1 ? 50 : index * 100 / (count - 1), text: formatAxisDate(points.value[index]?.date) }))
})

function parseBucket(value: string) {
  if (value.includes('T')) return new Date(value)
  return new Date(`${value}T00:00:00+08:00`)
}
function formatAxisDate(value?: string) {
  if (!value) return '—'
  const date = parseBucket(value)
  if (props.resolution === 'hour') return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
  if (props.resolution === 'month') return new Intl.DateTimeFormat(undefined, { year: '2-digit', month: 'short' }).format(date)
  return new Intl.DateTimeFormat(undefined, { month: 'numeric', day: 'numeric' }).format(date)
}
function formatBucketRange(value: string) {
  const bucketStart = parseBucket(value)
  const start = new Date(bucketStart)
  if (props.dateFrom) {
    const rangeStart = parseBucket(props.dateFrom)
    if (start < rangeStart) start.setTime(rangeStart.getTime())
  }
  if (props.resolution === 'hour') {
    const end = new Date(bucketStart.getTime() + 3_599_999)
    const day = new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric' }).format(start)
    const time = (date: Date) => new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
    return `${day} ${time(start)}–${time(end)}`
  }
  if (props.resolution === 'day') return new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'long', day: 'numeric' }).format(start)
  const end = new Date(bucketStart)
  if (props.resolution === 'week') end.setDate(end.getDate() + 6)
  else end.setMonth(end.getMonth() + 1, 0)
  if (props.dateTo) {
    const rangeEnd = parseBucket(props.dateTo)
    if (end > rangeEnd) end.setTime(rangeEnd.getTime())
  }
  const formatter = new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  return `${formatter.format(start)} – ${formatter.format(end)}`
}
function onPointerMove(event: PointerEvent) {
  const plot = plotRef.value
  if (!plot || !points.value.length) return
  const rect = plot.getBoundingClientRect()
  const ratio = Math.min(Math.max((event.clientX - rect.left) / Math.max(rect.width, 1), 0), 1)
  hoveredIndex.value = Math.round(ratio * (points.value.length - 1))
}
const tooltipStyle = computed(() => {
  if (hoveredIndex.value === null || points.value.length <= 1) return { left: '50%' }
  return { left: `${hoveredIndex.value * 100 / (points.value.length - 1)}%` }
})
const tooltipSide = computed(() => hoveredIndex.value !== null && hoveredIndex.value >= points.value.length * .7 ? 'align-right' : hoveredIndex.value !== null && hoveredIndex.value <= points.value.length * .3 ? 'align-left' : '')
const isMoney = computed(() => metric.value === 'paid' || metric.value === 'commission')
const money = (value: string | number) => new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY', minimumFractionDigits: 2 }).format(Number(value || 0))
const formatMetric = (value: number) => isMoney.value ? money(value) : new Intl.NumberFormat().format(value)
const formatAxis = (value: number) => {
  const compact = new Intl.NumberFormat(undefined, { notation: 'compact', maximumFractionDigits: 1 }).format(value)
  return isMoney.value ? `¥${compact}` : compact
}
watch([scope, metric, () => props.direct], () => { hoveredIndex.value = null })
</script>

<style scoped>
.business-chart{min-width:0;border:1px solid var(--ui-border);border-radius:8px;background:var(--ui-surface);padding:16px}.chart-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}.chart-controls{display:flex;flex-wrap:wrap;align-items:flex-end;justify-content:flex-end;gap:14px}.chart-control-group>span{display:block;margin-bottom:4px;color:var(--ui-text-subtle);font-size:10px;font-weight:600}.chart-segments{display:inline-flex;align-items:center;gap:2px}.chart-segments button{min-height:30px;border:1px solid transparent;border-radius:6px;padding:0 9px;color:var(--ui-text-muted);background:transparent;font-size:11px;font-weight:600;transition:background-color .16s ease,border-color .16s ease,color .16s ease,box-shadow .16s ease}.chart-segments button:hover:not([aria-pressed=true]){color:var(--ui-text);background:color-mix(in srgb,var(--ui-surface-subtle) 55%,transparent)}.chart-segments button[aria-pressed=true]{border-color:var(--ui-border);color:var(--ui-text);background:var(--ui-surface-raised);box-shadow:0 1px 3px rgb(0 0 0/.08)}.chart-segments button:focus-visible{outline:2px solid var(--ui-focus);outline-offset:1px}.chart-body{margin-top:18px}.chart-current{display:flex;align-items:baseline;gap:8px;margin-left:54px}.chart-current span{color:var(--ui-text-muted);font-size:11px}.chart-current strong{font-size:16px;font-weight:680;font-variant-numeric:tabular-nums}.chart-canvas{position:relative;height:246px;margin-top:8px;padding:0 8px 26px 54px}.chart-y-axis{position:absolute;top:0;bottom:26px;left:0;width:46px;display:flex;flex-direction:column;justify-content:space-between;text-align:right}.chart-y-axis span{color:var(--ui-text-subtle);font-size:10px;line-height:1;font-variant-numeric:tabular-nums}.chart-plot{position:relative;height:220px;min-width:0;cursor:crosshair;touch-action:pan-y}.chart-svg{display:block;width:100%;height:100%;overflow:visible}.chart-grid{stroke:var(--ui-border);stroke-width:1;vector-effect:non-scaling-stroke}.chart-area{fill:color-mix(in srgb,var(--ui-focus) 7%,transparent)}.chart-line{fill:none;stroke:var(--ui-focus);stroke-width:2;vector-effect:non-scaling-stroke}.chart-marker{fill:var(--ui-surface);stroke:var(--ui-focus);stroke-width:1.5;vector-effect:non-scaling-stroke}.chart-cursor{stroke:var(--ui-text-muted);stroke-width:1;stroke-dasharray:3 3;vector-effect:non-scaling-stroke}.chart-point{fill:var(--ui-surface-raised);stroke:var(--ui-focus);stroke-width:2.5;vector-effect:non-scaling-stroke}.chart-x-axis{position:absolute;right:8px;bottom:0;left:54px;height:20px}.chart-x-axis span{position:absolute;transform:translateX(-50%);color:var(--ui-text-subtle);font-size:10px;line-height:20px;white-space:nowrap}.chart-x-axis span:first-child{transform:none}.chart-x-axis span:last-child{transform:translateX(-100%)}.chart-tooltip{position:absolute;z-index:4;top:8px;width:214px;transform:translateX(-50%);border:1px solid color-mix(in srgb,var(--ui-border) 80%,transparent);border-radius:7px;background:var(--ui-surface-raised);padding:10px;color:var(--ui-text);box-shadow:var(--ui-shadow-lg);pointer-events:none}.chart-tooltip.align-left{transform:translateX(8px)}.chart-tooltip.align-right{transform:translateX(calc(-100% - 8px))}.tooltip-title{padding-bottom:7px;border-bottom:1px solid var(--ui-border);font-size:11px;font-weight:650}.chart-tooltip dl{display:grid;gap:4px;margin-top:7px}.chart-tooltip dl>div{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:12px;padding:2px 4px;border-radius:4px;font-size:11px}.chart-tooltip dl>div.active{background:var(--ui-surface-subtle)}.chart-tooltip dt{color:var(--ui-text-muted)}.chart-tooltip dd{font-weight:650;font-variant-numeric:tabular-nums}.chart-empty{display:flex;height:246px;align-items:center;justify-content:center;color:var(--ui-text-muted);font-size:13px}
@media(max-width:767px){.chart-heading{flex-direction:column}.chart-controls{width:100%;justify-content:flex-start}.chart-control-group{min-width:0}.chart-segments{flex-wrap:wrap}.chart-segments button{min-height:38px}.chart-canvas{height:226px;padding-left:46px}.chart-current{margin-left:46px}.chart-y-axis{width:38px}.chart-plot{height:200px}.chart-x-axis{left:46px}.chart-tooltip{top:6px;width:194px}.chart-x-axis span:nth-child(even):not(:last-child){display:none}}
@media(prefers-reduced-motion:reduce){.chart-segments button{transition:none}}
</style>
