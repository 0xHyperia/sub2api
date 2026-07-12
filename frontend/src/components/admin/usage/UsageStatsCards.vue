<template>
  <div class="grid grid-cols-1 gap-3 min-[420px]:grid-cols-2 lg:grid-cols-4">
    <article class="flex min-w-0 items-start gap-3 rounded-panel border border-outline bg-surface p-4">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-info-subtle text-info-foreground">
        <Icon name="document" size="md" />
      </div>
      <div class="min-w-0">
        <p class="text-xs font-medium text-foreground-muted">{{ t('usage.totalRequests') }}</p>
        <p class="break-all text-xl font-bold text-foreground">{{ stats?.total_requests?.toLocaleString() || '0' }}</p>
        <p class="text-xs text-foreground-subtle">{{ t('usage.inSelectedRange') }}</p>
      </div>
    </article>
    <article class="flex min-w-0 items-start gap-3 rounded-panel border border-outline bg-surface p-4">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-warning-subtle text-warning-foreground">
        <Icon name="cube" size="md" />
      </div>
      <div class="min-w-0">
        <p class="text-xs font-medium text-foreground-muted">{{ t('usage.totalTokens') }}</p>
        <p class="break-all text-xl font-bold text-foreground">{{ formatTokens(stats?.total_tokens || 0) }}</p>
        <div class="flex flex-wrap items-center gap-x-1 text-xs text-foreground-muted">
          <span>{{ t('usage.in') }}: {{ formatTokens(stats?.total_input_tokens || 0) }}</span>
          <span>/</span>
          <span>{{ t('usage.out') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}</span>
          <span>/</span>
          <details class="group relative inline-block">
            <summary class="inline-flex cursor-pointer list-none items-center gap-0.5 rounded-control focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus">
              <span>{{ cacheLabel() }}: {{ formatTokens(stats?.total_cache_tokens || 0) }}</span>
              <Icon name="infoCircle" size="xs" class="text-foreground-subtle" aria-hidden="true" />
            </summary>
            <span
              class="invisible absolute right-0 top-full z-30 mt-2 w-56 max-w-[calc(100vw-3rem)] rounded-panel border border-outline bg-surface-raised p-3 text-left text-xs text-foreground opacity-0 shadow-lg transition-opacity group-open:visible group-open:opacity-100"
            >
              <span class="mb-2 block font-medium text-foreground">
                {{ cacheDetailLabel() }}
              </span>
              <span class="flex items-center justify-between gap-3">
                <span>{{ t('usage.cacheCreationTokensLabel') }}</span>
                <span class="tabular-nums">
                  {{ formatTokens(stats?.total_cache_creation_tokens || 0) }}
                </span>
              </span>
              <span class="mt-1 flex items-center justify-between gap-3">
                <span>{{ t('usage.cacheReadTokensLabel') }}</span>
                <span class="tabular-nums">
                  {{ formatTokens(stats?.total_cache_read_tokens || 0) }}
                </span>
              </span>
            </span>
          </details>
        </div>
      </div>
    </article>
    <article class="flex min-w-0 items-start gap-3 rounded-panel border border-outline bg-surface p-4">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-success-subtle text-success-foreground">
        <Icon name="dollar" size="md" />
      </div>
      <div class="min-w-0 flex-1">
        <p class="text-xs font-medium text-foreground-muted">{{ t('usage.totalCost') }}</p>
        <p class="break-all text-xl font-bold text-success-foreground">
          ${{ (stats?.total_actual_cost || 0).toFixed(4) }}
        </p>
        <p class="break-words text-xs text-foreground-subtle">
          <template v-if="showAccountCost && totalAccountCost != null">
            <span class="text-warning-foreground">{{ t('usage.accountCost') }} ${{ totalAccountCost.toFixed(4) }}</span>
            <span> · </span>
          </template>
          <span>
            {{ t('usage.standardCost') }}
            <span :class="{ 'line-through': strikeStandardCost }">${{ (stats?.total_cost || 0).toFixed(4) }}</span>
          </span>
        </p>
      </div>
    </article>
    <article class="flex min-w-0 items-start gap-3 rounded-panel border border-outline bg-surface p-4">
      <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
        <Icon name="clock" size="md" />
      </div>
      <div class="min-w-0">
        <p class="text-xs font-medium text-foreground-muted">{{ t('usage.avgDuration') }}</p>
        <p class="break-all text-xl font-bold text-foreground">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import type { UsageStatsResponse } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  stats: (AdminUsageStatsResponse | UsageStatsResponse) | null
  showAccountCost?: boolean
  strikeStandardCost?: boolean
}>(), {
  showAccountCost: true,
  strikeStandardCost: false,
})

const { t } = useI18n()

const totalAccountCost = computed(() => {
  const stats = props.stats as (AdminUsageStatsResponse & { total_account_cost?: number }) | null
  return stats?.total_account_cost ?? null
})
const showAccountCost = computed(() => props.showAccountCost)
const strikeStandardCost = computed(() => props.strikeStandardCost)

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}

const cacheLabel = () => t('usage.cacheTotal')
const cacheDetailLabel = () => t('usage.cacheBreakdown')
</script>
