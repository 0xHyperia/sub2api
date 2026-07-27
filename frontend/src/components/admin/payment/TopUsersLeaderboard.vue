<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-foreground">
      {{ t('payment.admin.topUsers') }}
    </h3>
    <div
      v-if="!hasUsers(users)"
      class="flex h-32 items-center justify-center text-sm text-foreground-subtle"
    >
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else class="space-y-3">
      <div v-for="[currency, currencyUsers] in sortedUsers(users)" :key="currency" class="space-y-2">
        <p class="text-xs font-semibold text-foreground-subtle">{{ currency }}</p>
        <div
          v-for="(user, idx) in currencyUsers"
          :key="user.user_id"
          class="flex min-w-0 items-center justify-between gap-3 rounded-panel px-3 py-2 hover:bg-surface-subtle"
        >
          <div class="flex min-w-0 items-center gap-3">
            <span
              :class="[
                'flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold',
                rankClass(idx),
              ]"
            >
              {{ idx + 1 }}
            </span>
            <span class="truncate text-sm text-foreground-muted">{{ user.email }}</span>
          </div>
          <span class="shrink-0 text-sm font-medium text-foreground">
            {{ formatMoney(currency, user.amount) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { TopUserPaymentStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  users: Record<string, TopUserPaymentStats[]>
}>()

function rankClass(idx: number): string {
  if (idx === 0) return 'bg-warning-subtle text-warning-foreground'
  if (idx === 1) return 'bg-surface-subtle text-foreground-muted'
  if (idx === 2) return 'bg-warning-subtle text-warning-foreground'
  return 'bg-surface-subtle text-foreground-subtle'
}

function hasUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): boolean {
  return Object.values(usersByCurrency).some(users => users.length > 0)
}

function sortedUsers(usersByCurrency: Record<string, TopUserPaymentStats[]>): [string, TopUserPaymentStats[]][] {
  return Object.entries(usersByCurrency).sort(([left], [right]) => left.localeCompare(right))
}

function formatMoney(currency: string, amount: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount)
}
</script>
