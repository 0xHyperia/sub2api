<template>
  <div class="space-y-8">
    <section v-if="showDetailedPerformance">
      <div class="mb-3 flex items-end justify-between gap-4">
        <div>
          <h3 class="flex items-center gap-2 text-sm font-semibold text-foreground">
            <Icon name="clock" size="sm" class="text-foreground-subtle" />
            {{ t('modelMarketplace.details.latencyTrend') }}
          </h3>
          <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.latencyTrendHint') }}</p>
        </div>
      </div>
      <div v-if="hasLatencyData" class="h-64 min-w-0 sm:h-72">
        <Line :data="latencyChartData" :options="latencyChartOptions" :plugins="[hoverColumnPlugin]" />
      </div>
      <div v-else class="flex h-52 items-center justify-center border-y border-outline text-sm text-foreground-subtle">
        {{ t('modelMarketplace.details.noPerformanceData') }}
      </div>
    </section>

    <section>
      <div class="mb-3 flex items-end justify-between gap-4">
        <div>
          <h3 class="flex items-center gap-2 text-sm font-semibold text-foreground">
            <Icon name="badge" size="sm" class="text-foreground-subtle" />
            {{ t('modelMarketplace.details.availabilityTrend') }}
          </h3>
          <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.availabilityTrendHint') }}</p>
        </div>
      </div>
      <div v-if="hasAvailabilityData" class="h-64 min-w-0 sm:h-72">
        <Line :data="availabilityChartData" :options="availabilityChartOptions" :plugins="[hoverColumnPlugin]" />
      </div>
      <div v-else class="flex h-52 items-center justify-center border-y border-outline text-sm text-foreground-subtle">
        {{ t('modelMarketplace.details.noPerformanceData') }}
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import Icon from '@/components/icons/Icon.vue'
import { useChartTheme } from '@/composables/useChartTheme'
import type { UserModelMonitorMetricBucket } from '@/api/channels'

const hoverColumnPlugin = {
  id: 'marketplaceHoverColumn',
  beforeDatasetsDraw(chart: ChartJS) {
    const active = chart.tooltip?.getActiveElements() ?? []
    if (!active.length) return
    const { ctx, chartArea } = chart
    const pointCount = chart.data.labels?.length ?? 0
    if (pointCount === 0) return
    const activeIndex = active[0].index
    const bandWidth = (chartArea.right - chartArea.left) / pointCount
    const left = chartArea.left + (activeIndex * bandWidth)
    ctx.save()
    ctx.fillStyle = 'rgba(148, 163, 184, 0.10)'
    ctx.fillRect(left, chartArea.top, bandWidth, chartArea.bottom - chartArea.top)
    ctx.restore()
  },
  beforeTooltipDraw(chart: ChartJS) {
    chart.ctx.save()
    chart.ctx.shadowColor = 'rgba(15, 23, 42, 0.14)'
    chart.ctx.shadowBlur = 16
    chart.ctx.shadowOffsetY = 5
  },
  afterTooltipDraw(chart: ChartJS) {
    chart.ctx.restore()
  },
}

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const props = withDefaults(defineProps<{
  buckets?: UserModelMonitorMetricBucket[] | null
  showDetailedPerformance?: boolean
}>(), {
  buckets: () => [],
  showDetailedPerformance: true,
})

const { t, locale } = useI18n()
const { chartTheme } = useChartTheme()

const orderedBuckets = computed(() => [...(props.buckets ?? [])].slice(-24))
const latencyBuckets = computed(() => orderedBuckets.value.filter(bucket => bucket.ttft_ms != null))
const availabilityBuckets = computed(() => orderedBuckets.value.filter(bucket => bucket.success_rate != null))
const hasLatencyData = computed(() => latencyBuckets.value.length > 0)
const hasAvailabilityData = computed(() => availabilityBuckets.value.length > 0)
const timeFormatter = computed(() => new Intl.DateTimeFormat(locale.value, {
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
}))
function labelsFor(buckets: UserModelMonitorMetricBucket[]) {
  return buckets.map(bucket => timeFormatter.value.format(new Date(bucket.started_at)))
}

const latencyChartData = computed(() => ({
  labels: labelsFor(latencyBuckets.value),
  datasets: [{
    label: 'TTFT',
    data: latencyBuckets.value.map(bucket => bucket.ttft_ms),
    borderColor: chartTheme.value.info,
    backgroundColor: chartTheme.value.infoAlpha,
    pointBackgroundColor: chartTheme.value.info,
    pointBorderWidth: 0,
    pointRadius: 2,
    pointHoverRadius: 4,
    borderWidth: 2,
    fill: false,
    tension: 0.25,
    spanGaps: true,
  }],
}))

const availabilityChartData = computed(() => ({
  labels: labelsFor(availabilityBuckets.value),
  datasets: [{
    label: t('modelMarketplace.details.successRate'),
    data: availabilityBuckets.value.map(bucket => bucket.success_rate),
    borderColor: chartTheme.value.success,
    backgroundColor: chartTheme.value.successAlpha,
    pointBackgroundColor: chartTheme.value.success,
    pointBorderWidth: 0,
    pointRadius: 2,
    pointHoverRadius: 4,
    borderWidth: 2,
    clip: false as const,
    fill: false,
    tension: 0.2,
  }],
}))

const sharedOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const, axis: 'x' as const },
  plugins: {
    legend: { display: false },
    tooltip: {
      displayColors: true,
      usePointStyle: true,
      boxWidth: 7,
      boxHeight: 7,
      backgroundColor: chartTheme.value.surfaceRaised,
      borderColor: chartTheme.value.outline,
      borderWidth: 1,
      titleColor: chartTheme.value.foreground,
      bodyColor: chartTheme.value.foregroundMuted,
      titleFont: { size: 13, weight: 'bold' as const },
      bodyFont: { size: 12 },
      padding: 12,
      cornerRadius: 4,
      caretSize: 0,
      bodySpacing: 6,
    },
  },
  scales: {
    x: {
      offset: true,
      grid: { display: false },
      border: { color: chartTheme.value.outline },
      ticks: { color: chartTheme.value.foregroundSubtle, autoSkip: true, maxTicksLimit: 8, font: { size: 10 } },
    },
  },
}))

const latencyChartOptions = computed(() => ({
  ...sharedOptions.value,
  plugins: {
    ...sharedOptions.value.plugins,
    tooltip: {
      ...sharedOptions.value.plugins.tooltip,
      callbacks: {
        label: (context: { parsed: { y: number | null } }) => `${t('modelMarketplace.details.latency')}   ${context.parsed.y == null ? '—' : `${Math.round(context.parsed.y)} ms`}`,
      },
    },
  },
  scales: {
    ...sharedOptions.value.scales,
    y: {
      beginAtZero: true,
      grid: { color: chartTheme.value.outline },
      border: { display: false },
      ticks: {
        color: chartTheme.value.foregroundSubtle,
        maxTicksLimit: 5,
        font: { size: 10 },
        callback: (value: string | number) => `${value} ms`,
      },
    },
  },
}))

const availabilityChartOptions = computed(() => ({
  ...sharedOptions.value,
  plugins: {
    ...sharedOptions.value.plugins,
    tooltip: {
      ...sharedOptions.value.plugins.tooltip,
      callbacks: {
        label: (context: { parsed: { y: number | null } }) => `${t('modelMarketplace.details.successRate')}   ${context.parsed.y == null ? '—' : `${context.parsed.y.toFixed(2)}%`}`,
      },
    },
  },
  scales: {
    ...sharedOptions.value.scales,
    y: {
      min: availabilityAxisMin.value,
      max: 100,
      grid: { color: chartTheme.value.outline },
      border: { display: false },
      ticks: {
        color: chartTheme.value.foregroundSubtle,
        maxTicksLimit: 6,
        font: { size: 10 },
        callback: (value: string | number) => `${value}%`,
      },
    },
  },
}))

const availabilityAxisMin = computed(() => {
  const values = availabilityBuckets.value.flatMap(bucket => bucket.success_rate == null ? [] : [bucket.success_rate])
  if (!values.length) return 0
  const minimum = Math.min(...values)
  if (minimum >= 95) return 95
  if (minimum >= 80) return Math.max(0, Math.floor(minimum / 5) * 5 - 5)
  return Math.max(0, Math.floor(minimum / 10) * 10 - 10)
})
</script>
