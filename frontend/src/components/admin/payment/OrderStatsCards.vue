<template>
  <div class="grid grid-cols-1 gap-3 min-[420px]:grid-cols-2 lg:grid-cols-4">
    <!-- Today Revenue -->
    <article class="min-w-0 rounded-panel border border-outline bg-surface p-4">
      <div class="flex min-w-0 items-start gap-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-success-subtle text-success-foreground">
          <Icon name="dollar" size="md" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-foreground-muted">{{ t('payment.admin.todayRevenue') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.today_amount)" :key="currency" class="break-all text-xl font-bold text-foreground">
            {{ formatMoney(currency, amount) }}
          </p>
          <p class="text-xs text-foreground-subtle">
            {{ stats.today_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </article>

    <!-- Total Revenue -->
    <article class="min-w-0 rounded-panel border border-outline bg-surface p-4">
      <div class="flex min-w-0 items-start gap-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-info-subtle text-info-foreground">
          <Icon name="creditCard" size="md" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-foreground-muted">{{ t('payment.admin.totalRevenue') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.total_amount)" :key="currency" class="break-all text-xl font-bold text-foreground">
            {{ formatMoney(currency, amount) }}
          </p>
          <p class="text-xs text-foreground-subtle">
            {{ stats.total_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </article>

    <!-- Today Orders -->
    <article class="min-w-0 rounded-panel border border-outline bg-surface p-4">
      <div class="flex min-w-0 items-start gap-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-surface-subtle text-foreground-muted">
          <Icon name="chart" size="md" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-foreground-muted">{{ t('payment.admin.todayOrders') }}</p>
          <p class="break-all text-xl font-bold text-foreground">{{ stats.today_count }}</p>
        </div>
      </div>
    </article>

    <!-- Average Amount -->
    <article class="min-w-0 rounded-panel border border-outline bg-surface p-4">
      <div class="flex min-w-0 items-start gap-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-warning-subtle text-warning-foreground">
          <Icon name="chart" size="md" :stroke-width="2" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-foreground-muted">{{ t('payment.admin.avgAmount') }}</p>
          <p v-for="[currency, amount] in sortedAmounts(stats.avg_amount)" :key="currency" class="break-all text-xl font-bold text-foreground">
            {{ formatMoney(currency, amount) }}
          </p>
        </div>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { CurrencyAmounts, DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function sortedAmounts(amounts: CurrencyAmounts): [string, number][] {
  return Object.entries(amounts).sort(([left], [right]) => left.localeCompare(right))
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}
</script>
