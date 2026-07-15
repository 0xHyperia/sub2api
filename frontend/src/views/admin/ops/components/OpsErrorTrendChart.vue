<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Title,
  Tooltip
} from 'chart.js'
import { Line } from 'vue-chartjs'
import type { OpsErrorTrendPoint } from '@/api/admin/ops'
import type { ChartState } from '../types'
import { formatHistoryLabel, sumNumbers } from '../utils/opsFormatters'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useOpsChartTheme } from '../utils/useOpsChartTheme'

ChartJS.register(Title, Tooltip, Legend, LineElement, LinearScale, PointElement, CategoryScale, Filler)

interface Props {
  points: OpsErrorTrendPoint[]
  loading: boolean
  timeRange: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'openRequestErrors'): void
  (e: 'openUpstreamErrors'): void
}>()
const { t } = useI18n()

const { chartTheme } = useOpsChartTheme()
const colors = computed(() => ({
  red: chartTheme.value.danger,
  redAlpha: chartTheme.value.dangerAlpha,
  purple: chartTheme.value.info,
  purpleAlpha: chartTheme.value.infoAlpha,
  gray: chartTheme.value.warning,
  grid: chartTheme.value.outline,
  text: chartTheme.value.foregroundSubtle,
  tooltip: chartTheme.value.surfaceRaised,
  tooltipTitle: chartTheme.value.foreground,
  tooltipBody: chartTheme.value.foregroundMuted,
}))

const totalRequestErrors = computed(() => sumNumbers(props.points.map((p) => p.error_count_sla ?? 0)))

const totalUpstreamErrors = computed(() =>
  sumNumbers(
    props.points.map((p) => (p.upstream_error_count_excl_429_529 ?? 0) + (p.upstream_429_count ?? 0) + (p.upstream_529_count ?? 0))
  )
)

const totalDisplayed = computed(() =>
  sumNumbers(props.points.map((p) => (p.error_count_sla ?? 0) + (p.upstream_error_count_excl_429_529 ?? 0) + (p.business_limited_count ?? 0)))
)

const hasRequestErrors = computed(() => totalRequestErrors.value > 0)
const hasUpstreamErrors = computed(() => totalUpstreamErrors.value > 0)

const chartData = computed(() => {
  if (!props.points.length || totalDisplayed.value <= 0) return null
  return {
    labels: props.points.map((p) => formatHistoryLabel(p.bucket_start, props.timeRange)),
    datasets: [
      {
        label: t('admin.ops.errorsSla'),
        data: props.points.map((p) => p.error_count_sla ?? 0),
        borderColor: colors.value.red,
        backgroundColor: colors.value.redAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.upstreamExcl429529'),
        data: props.points.map((p) => p.upstream_error_count_excl_429_529 ?? 0),
        borderColor: colors.value.purple,
        backgroundColor: colors.value.purpleAlpha,
        fill: true,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      },
      {
        label: t('admin.ops.businessLimited'),
        data: props.points.map((p) => p.business_limited_count ?? 0),
        borderColor: colors.value.gray,
        backgroundColor: 'transparent',
        borderDash: [6, 6],
        fill: false,
        tension: 0.35,
        pointRadius: 0,
        pointHitRadius: 10
      }
    ]
  }
})

const state = computed<ChartState>(() => {
  if (chartData.value) return 'ready'
  if (props.loading) return 'loading'
  return 'empty'
})

const options = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { intersect: false, mode: 'index' as const },
    plugins: {
      legend: {
        position: 'top' as const,
        align: 'end' as const,
        labels: { color: c.text, usePointStyle: true, boxWidth: 6, font: { size: 10 } }
      },
      tooltip: {
        backgroundColor: c.tooltip,
        titleColor: c.tooltipTitle,
        bodyColor: c.tooltipBody,
        borderColor: c.grid,
        borderWidth: 1,
        padding: 10,
        displayColors: true
      }
    },
    scales: {
      x: {
        type: 'category' as const,
        grid: { display: false },
        ticks: {
          color: c.text,
          font: { size: 10 },
          maxTicksLimit: 8,
          autoSkip: true,
          autoSkipPadding: 10
        }
      },
      y: {
        type: 'linear' as const,
        display: true,
        position: 'left' as const,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 }, precision: 0 }
      }
    }
  }
})
</script>

<template>
  <section class="flex min-h-[300px] flex-col rounded-panel border border-outline bg-surface p-4 shadow-card sm:p-5" aria-labelledby="ops-error-trend-title">
    <div class="mb-4 flex shrink-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h3 id="ops-error-trend-title" class="flex items-center gap-2 text-sm font-semibold text-foreground">
        <svg class="h-4 w-4 text-danger" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"
          />
        </svg>
        {{ t('admin.ops.errorTrend') }}
        <HelpTooltip :content="t('admin.ops.tooltips.errorTrend')" />
      </h3>
      <div class="grid grid-cols-2 gap-2 sm:flex sm:items-center">
        <button
          type="button"
          class="btn btn-secondary min-h-control px-2 text-xs sm:min-h-control-sm"
          :disabled="!hasRequestErrors"
          @click="emit('openRequestErrors')"
        >
          {{ t('admin.ops.errorDetails.requestErrors') }}
        </button>
        <button
          type="button"
          class="btn btn-secondary min-h-control px-2 text-xs sm:min-h-control-sm"
          :disabled="!hasUpstreamErrors"
          @click="emit('openUpstreamErrors')"
        >
          {{ t('admin.ops.errorDetails.upstreamErrors') }}
        </button>
      </div>
    </div>

    <div class="min-h-0 flex-1">
      <Line v-if="state === 'ready' && chartData" :data="chartData" :options="options" />
      <div v-else class="flex h-full items-center justify-center">
        <div v-if="state === 'loading'" class="animate-pulse text-sm text-foreground-subtle">{{ t('common.loading') }}</div>
        <EmptyState v-else :title="t('common.noData')" :description="t('admin.ops.charts.emptyError')" />
      </div>
    </div>
  </section>
</template>
