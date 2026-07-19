<template>
  <section class="space-y-4" aria-labelledby="dashboard-analytics-title">
    <div class="border-b border-outline pb-4">
      <h2 id="dashboard-analytics-title" class="sr-only">{{ t('dashboard.tokenUsageTrend') }}</h2>

      <div class="grid min-w-0 grid-cols-[minmax(0,1fr)_112px_40px] items-end gap-2 sm:flex sm:flex-wrap">
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

        <div class="min-w-0 sm:w-32">
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
          <div class="w-full min-w-0 flex-1">
            <ol v-if="mobileModels.length" class="divide-y divide-outline sm:hidden" :aria-label="t('dashboard.modelDistribution')">
              <li v-for="(model, index) in mobileModels" :key="model.model" class="flex min-w-0 items-center gap-3 py-2.5 first:pt-0 last:pb-0">
                <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-control bg-surface-subtle text-xs font-semibold tabular-nums text-foreground-subtle" aria-hidden="true">
                  {{ index + 1 }}
                </span>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-medium text-foreground" :title="model.model">{{ model.model }}</p>
                  <p class="mt-0.5 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-foreground-subtle">
                    <span>{{ formatNumber(model.requests) }} {{ t('dashboard.requests') }}</span>
                    <span>{{ formatTokens(model.total_tokens) }} {{ t('dashboard.tokens') }}</span>
                  </p>
                </div>
                <div class="shrink-0 text-right">
                  <p class="text-sm font-semibold tabular-nums text-success-foreground">${{ formatCost(model.actual_cost) }}</p>
                  <p class="text-[10px] text-foreground-subtle">{{ t('dashboard.actual') }}</p>
                </div>
              </li>
            </ol>
            <div class="hidden max-h-48 overflow-auto sm:block">
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

const mobileModels = computed(() =>
  [...(props.models ?? [])]
    .sort((a, b) => (b.actual_cost ?? 0) - (a.actual_cost ?? 0))
    .slice(0, 5)
)

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
