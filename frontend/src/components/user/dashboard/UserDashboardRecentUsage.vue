<template>
  <section class="card min-w-0 overflow-hidden" aria-labelledby="dashboard-recent-usage-title">
    <header class="flex min-h-[3.25rem] items-center justify-between gap-3 border-b border-outline bg-surface-subtle px-4 py-3 sm:px-5">
      <h2 id="dashboard-recent-usage-title" class="text-sm font-semibold text-foreground">{{ t('dashboard.recentUsage') }}</h2>
      <span class="badge badge-gray">{{ t('dashboard.last7Days') }}</span>
    </header>
    <div>
      <div v-if="error" class="px-4 py-8 sm:px-5" role="alert">
        <EmptyState
          :title="t('dashboard.recentUsageLoadFailed')"
          :description="t('errors.tryAgain')"
          :action-text="t('common.retry')"
          :action-icon="false"
          @action="$emit('retry')"
        />
      </div>
      <div v-else-if="loading" class="flex min-h-56 items-center justify-center" role="status">
        <LoadingSpinner size="lg" />
        <span class="sr-only">{{ t('common.loading') }}</span>
      </div>
      <div v-else-if="data.length === 0" class="px-4 py-8 sm:px-5">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <template v-else>
        <ul class="divide-y divide-outline">
          <li v-for="log in data" :key="log.id" class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 transition-colors hover:bg-surface-subtle sm:px-5">
            <div class="flex min-w-0 items-center gap-3">
              <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-muted" aria-hidden="true">
                <Icon name="beaker" size="sm" />
              </span>
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-foreground" :title="log.model">{{ log.model }}</p>
                <p class="text-xs text-foreground-subtle">{{ formatDateTime(log.created_at) }}</p>
              </div>
            </div>
            <div class="min-w-0 max-w-48 text-right tabular-nums">
              <p class="break-all text-sm font-semibold text-success-foreground">
                <span :title="t('dashboard.actual')">${{ formatCost(log.actual_cost) }}</span>
                <span class="font-normal text-foreground-subtle" :title="t('dashboard.standard')"> / ${{ formatCost(log.total_cost) }}</span>
              </p>
              <p class="text-xs text-foreground-subtle">{{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens</p>
            </div>
          </li>
        </ul>

        <router-link to="/usage" class="flex min-h-11 items-center justify-center gap-2 border-t border-outline px-4 py-2 text-sm font-medium text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus">
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
  error?: boolean
}>()
defineEmits<{ retry: [] }>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
</script>
