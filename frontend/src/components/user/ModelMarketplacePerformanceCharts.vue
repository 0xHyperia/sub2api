<template>
  <div class="space-y-8">
    <section>
      <div class="mb-3 flex items-end justify-between gap-4">
        <div>
          <h3 class="flex items-center gap-2 text-sm font-semibold text-foreground">
            <Icon name="clock" size="sm" class="text-foreground-subtle" />
            {{ t('modelMarketplace.details.latencyTrend') }}
          </h3>
          <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.recentMonitorSamples') }}</p>
        </div>
      </div>
      <div v-if="hasLatencyData" class="h-64 min-w-0 sm:h-72">
        <Line :data="latencyChartData" :options="latencyChartOptions" />
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
          <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('modelMarketplace.details.rollingSuccessRate') }}</p>
        </div>
      </div>
      <div v-if="orderedPoints.length" class="h-64 min-w-0 sm:h-72">
        <Line :data="availabilityChartData" :options="availabilityChartOptions" />
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
import type { UserModelMonitorTimelinePoint } from '@/api/channels'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler)

const props = defineProps<{
  points?: UserModelMonitorTimelinePoint[] | null
}>()

const { t, locale } = useI18n()
const { chartTheme } = useChartTheme()

const orderedPoints = computed(() => [...(props.points ?? [])].reverse())
const hasLatencyData = computed(() => orderedPoints.value.some(point => point.latency_ms != null))
const chartLabels = computed(() => orderedPoints.value.map(point => new Intl.DateTimeFormat(locale.value, {
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
}).format(new Date(point.checked_at))))

const latencyChartData = computed(() => ({
  labels: chartLabels.value,
  datasets: [{
    label: t('modelMarketplace.details.latency'),
    data: orderedPoints.value.map(point => point.latency_ms),
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

const rollingAvailability = computed(() => orderedPoints.value.map((_, index, points) => {
  const window = points.slice(Math.max(0, index - 11), index + 1)
  const successful = window.filter(point => point.status === 'operational' || point.status === 'degraded').length
  return window.length ? (successful / window.length) * 100 : null
}))

const availabilityChartData = computed(() => ({
  labels: chartLabels.value,
  datasets: [{
    label: t('modelMarketplace.details.successRate'),
    data: rollingAvailability.value,
    borderColor: chartTheme.value.success,
    backgroundColor: chartTheme.value.successAlpha,
    pointBackgroundColor: chartTheme.value.success,
    pointBorderWidth: 0,
    pointRadius: 2,
    pointHoverRadius: 4,
    borderWidth: 2,
    fill: false,
    tension: 0.2,
  }],
}))

const sharedOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { intersect: false, mode: 'index' as const },
  plugins: {
    legend: { display: false },
    tooltip: {
      displayColors: false,
      backgroundColor: chartTheme.value.surfaceRaised,
      borderColor: chartTheme.value.outline,
      borderWidth: 1,
      titleColor: chartTheme.value.foreground,
      bodyColor: chartTheme.value.foregroundMuted,
      padding: 10,
    },
  },
  scales: {
    x: {
      grid: { display: false },
      border: { color: chartTheme.value.outline },
      ticks: { color: chartTheme.value.foregroundSubtle, maxTicksLimit: 8, font: { size: 10 } },
    },
  },
}))

const latencyChartOptions = computed(() => ({
  ...sharedOptions.value,
  scales: {
    ...sharedOptions.value.scales,
    y: {
      beginAtZero: true,
      grid: { color: chartTheme.value.outline },
      border: { display: false },
      ticks: {
        color: chartTheme.value.foregroundSubtle,
        font: { size: 10 },
        callback: (value: string | number) => `${value} ms`,
      },
    },
  },
}))

const availabilityChartOptions = computed(() => ({
  ...sharedOptions.value,
  scales: {
    ...sharedOptions.value.scales,
    y: {
      min: 0,
      max: 100,
      grid: { color: chartTheme.value.outline },
      border: { display: false },
      ticks: {
        color: chartTheme.value.foregroundSubtle,
        stepSize: 20,
        font: { size: 10 },
        callback: (value: string | number) => `${value}%`,
      },
    },
  },
}))
</script>
