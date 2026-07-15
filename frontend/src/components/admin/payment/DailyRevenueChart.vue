<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-foreground">
      {{ t('payment.admin.dailyRevenue') }}
    </h3>
    <div class="h-64">
      <div v-if="loading" class="flex h-full items-center justify-center">
        <LoadingSpinner size="md" />
      </div>
      <Line v-else-if="chartData" :data="chartData" :options="chartOptions" />
      <div
        v-else
        class="flex h-full items-center justify-center text-sm text-foreground-subtle"
      >
        {{ t('payment.admin.noData') }}
      </div>
    </div>
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
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { useChartTheme } from '@/composables/useChartTheme'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t } = useI18n()
const { chartTheme } = useChartTheme()

const props = defineProps<{
  data: { date: string; amount: number; count: number }[]
  loading?: boolean
}>()

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) return null
  return {
    labels: props.data.map(d => d.date),
    datasets: [
      {
        label: t('payment.admin.revenue'),
        data: props.data.map(d => d.amount),
        borderColor: chartTheme.value.info,
        backgroundColor: chartTheme.value.infoAlpha,
        fill: true,
        tension: 0.3,
        pointRadius: 3,
        pointHoverRadius: 5,
      },
      {
        label: t('payment.admin.orderCount'),
        data: props.data.map(d => d.count),
        borderColor: chartTheme.value.success,
        backgroundColor: chartTheme.value.successAlpha,
        fill: false,
        tension: 0.3,
        pointRadius: 3,
        pointHoverRadius: 5,
        yAxisID: 'y1',
      }
    ]
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  scales: {
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      grid: { color: chartTheme.value.outline },
      ticks: { color: chartTheme.value.foregroundMuted },
      title: {
        display: true,
        text: t('payment.admin.revenue'),
        color: chartTheme.value.foregroundMuted,
      },
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      ticks: { color: chartTheme.value.foregroundMuted },
      title: {
        display: true,
        text: t('payment.admin.orderCount'),
        color: chartTheme.value.foregroundMuted,
      },
      grid: { drawOnChartArea: false },
    }
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: { color: chartTheme.value.foregroundMuted },
    },
  }
}))
</script>
