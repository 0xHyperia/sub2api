<template>
  <section class="space-y-4" aria-labelledby="dashboard-analytics-title">
    <div class="flex flex-col gap-3 border-b border-outline pb-4 lg:flex-row lg:items-end lg:justify-end">
      <h2 id="dashboard-analytics-title" class="sr-only">{{ t('dashboard.tokenUsageTrend') }}</h2>

      <div class="flex min-w-0 flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
        <div class="min-w-0" role="group" :aria-label="t('dashboard.timeRange')">
          <span class="mb-1.5 block text-xs font-medium text-foreground-muted">{{ t('dashboard.timeRange') }}</span>
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:startDate="$emit('update:startDate', $event)"
            @update:endDate="$emit('update:endDate', $event)"
            @change="$emit('dateRangeChange', $event)"
          />
        </div>

        <div class="w-full sm:w-32">
          <span class="mb-1.5 block text-xs font-medium text-foreground-muted">{{ t('dashboard.granularity') }}</span>
          <Select
            :model-value="granularity"
            :placeholder="t('dashboard.granularity')"
            :options="[{ value: 'day', label: t('dashboard.day') }, { value: 'hour', label: t('dashboard.hour') }]"
            @update:model-value="$emit('update:granularity', $event)"
            @change="$emit('granularityChange')"
          />
        </div>

        <button
          type="button"
          class="btn btn-secondary btn-icon"
          :disabled="loading"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          @click="$emit('refresh')"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>
    </div>

    <div
      v-if="error"
      class="flex flex-col gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4 text-danger-foreground sm:flex-row sm:items-center sm:justify-between"
      role="alert"
    >
      <span class="text-sm font-medium">{{ t('dashboard.chartsLoadFailed') }}</span>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="$emit('refresh')">
        <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
        {{ t('common.retry') }}
      </button>
    </div>

    <div class="grid min-w-0 grid-cols-1 gap-4 lg:grid-cols-2">
      <div class="card relative min-w-0 overflow-hidden p-4 sm:p-5">
        <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center bg-surface/80" role="status">
          <LoadingSpinner size="md" />
          <span class="sr-only">{{ t('common.loading') }}</span>
        </div>
        <h3 class="mb-4 text-sm font-semibold text-foreground">{{ t('dashboard.modelDistribution') }}</h3>
        <div class="flex min-w-0 flex-col items-center gap-4 sm:flex-row sm:items-center">
          <div class="h-40 w-40 shrink-0 sm:h-44 sm:w-44">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div v-else class="flex h-full items-center justify-center text-center text-sm text-foreground-subtle">{{ t('dashboard.noDataAvailable') }}</div>
          </div>
          <div class="max-h-48 w-full min-w-0 flex-1 overflow-auto">
            <table class="w-full min-w-[30rem] text-xs">
              <thead>
                <tr class="text-foreground-subtle">
                  <th class="pb-2 text-left">{{ t('dashboard.model') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.requests') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.tokens') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.actual') }}</th>
                  <th class="pb-2 text-right">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="model in models" :key="model.model" class="border-t border-outline">
                  <td class="max-w-[10rem] truncate py-2 font-medium text-foreground" :title="model.model">{{ model.model }}</td>
                  <td class="py-2 text-right text-foreground-muted">{{ formatNumber(model.requests) }}</td>
                  <td class="py-2 text-right text-foreground-muted">{{ formatTokens(model.total_tokens) }}</td>
                  <td class="py-2 text-right text-success-foreground">${{ formatCost(model.actual_cost) }}</td>
                  <td class="py-2 text-right text-foreground-subtle">${{ formatCost(model.cost) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, error?: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16']
  }]
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}
</script>
