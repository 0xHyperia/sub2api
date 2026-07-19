<template>
  <aside
    class="flex min-w-0 flex-col overflow-hidden rounded-panel border border-outline bg-surface shadow-card"
    aria-labelledby="affiliate-reward-title"
  >
    <header class="flex items-center justify-between gap-3 border-b border-outline px-4 py-3.5 sm:px-5">
      <div class="flex min-w-0 items-center gap-3">
        <Icon name="gift" size="md" class="shrink-0 text-success-foreground" aria-hidden="true" />
        <h2 id="affiliate-reward-title" class="text-sm font-semibold text-foreground">
          {{ t('purchaseWorkspace.affiliateTitle') }}
        </h2>
      </div>
      <RouterLink to="/affiliate" class="btn btn-ghost btn-sm shrink-0">
        <span>{{ t('purchaseWorkspace.details') }}</span>
        <Icon name="arrowRight" size="sm" aria-hidden="true" />
      </RouterLink>
    </header>

    <div v-if="loading" class="flex min-h-48 items-center justify-center" aria-live="polite">
      <LoadingSpinner />
    </div>

    <div v-else-if="loadFailed && !detail" class="p-5" role="alert">
      <EmptyState
        :title="t('affiliate.loadFailed')"
        :description="t('errors.tryAgain')"
        :action-text="t('common.refresh')"
        @action="loadAffiliateDetail()"
      />
    </div>

    <template v-else-if="detail">
      <section class="affiliate-stat-banner px-4 py-4 text-foreground sm:px-5" :aria-label="t('affiliate.title')">
        <div class="mb-5 flex items-center justify-between gap-3">
          <h3 class="text-base font-semibold text-foreground">{{ t('purchaseWorkspace.rewardStats') }}</h3>
          <button
            type="button"
            class="reward-transfer-button inline-flex min-h-8 shrink-0 items-center justify-center gap-1.5 rounded-control px-3 text-xs font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50"
            :disabled="transferring || detail.aff_quota <= 0"
            :aria-busy="transferring"
            @click="transferQuota"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': transferring }" aria-hidden="true" />
            <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
          </button>
        </div>
        <dl class="grid grid-cols-3 divide-x divide-outline/60 text-center">
          <div class="min-w-0 px-2">
            <dd class="break-words text-xl font-semibold tabular-nums text-foreground sm:text-2xl">{{ formatCurrency(detail.aff_quota) }}</dd>
            <dt class="mt-1.5 text-[11px] leading-4 text-foreground-muted sm:text-xs">{{ t('affiliate.stats.availableQuota') }}</dt>
          </div>
          <div class="min-w-0 px-2">
            <dd class="break-words text-xl font-semibold tabular-nums text-foreground sm:text-2xl">{{ formatCurrency(detail.aff_history_quota) }}</dd>
            <dt class="mt-1.5 text-[11px] leading-4 text-foreground-muted sm:text-xs">{{ t('affiliate.stats.totalQuota') }}</dt>
          </div>
          <div class="min-w-0 px-2">
            <dd class="text-xl font-semibold tabular-nums text-foreground sm:text-2xl">{{ formatCount(detail.aff_count) }}</dd>
            <dt class="mt-1.5 text-[11px] leading-4 text-foreground-muted sm:text-xs">{{ t('affiliate.stats.invitedUsers') }}</dt>
          </div>
        </dl>
        <p v-if="detail.aff_frozen_quota > 0" class="mt-4 text-xs text-warning-foreground">
          {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
        </p>
      </section>

      <div class="flex flex-1 flex-col">
        <section class="p-4 sm:p-5" aria-labelledby="affiliate-link-title">
          <h3 id="affiliate-link-title" class="text-sm font-medium text-foreground">{{ t('affiliate.inviteLink') }}</h3>
          <div class="mt-2 flex min-w-0 items-center gap-2 rounded-panel border border-outline-strong bg-surface-subtle p-1.5 pl-3">
            <code class="min-w-0 flex-1 truncate text-xs text-foreground-muted" :title="inviteLink">{{ inviteLink }}</code>
            <button type="button" class="btn btn-primary btn-sm shrink-0" @click="copyInviteLink">
              <Icon name="copy" size="sm" aria-hidden="true" />
              <span>{{ t('common.copy') }}</span>
            </button>
          </div>
        </section>

        <section class="border-t border-outline px-4 py-4 sm:px-5" aria-labelledby="affiliate-rules-title">
          <h3 id="affiliate-rules-title" class="text-sm font-medium text-foreground">{{ t('purchaseWorkspace.rewardRules') }}</h3>
          <ul class="mt-3 space-y-2 text-sm leading-5 text-foreground-muted">
            <li class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-success" aria-hidden="true" />
              <span>{{ t('affiliate.tips.line1') }}</span>
            </li>
            <li class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-success" aria-hidden="true" />
              <span>{{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</span>
            </li>
            <li class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-success" aria-hidden="true" />
              <span>{{ t('affiliate.tips.line3') }}</span>
            </li>
            <li v-if="detail.aff_frozen_quota > 0" class="flex gap-2">
              <span class="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-warning" aria-hidden="true" />
              <span>{{ t('affiliate.tips.line4') }}</span>
            </li>
          </ul>
        </section>
      </div>
    </template>
  </aside>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { useAffiliateRewards } from '@/composables/useAffiliateRewards'
import { formatPaymentAmount } from '@/components/payment/currency'

const { t } = useI18n()
const {
  detail,
  loading,
  loadFailed,
  transferring,
  inviteLink,
  formattedRebateRate,
  loadAffiliateDetail,
  copyInviteLink,
  transferQuota,
} = useAffiliateRewards()

function formatCount(value: number): string {
  return value.toLocaleString()
}

function formatCurrency(value: number): string {
  return formatPaymentAmount(value, 'USD')
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>

<style scoped>
.affiliate-stat-banner {
  border-bottom: 1px solid rgb(var(--color-border));
  background: rgb(var(--color-info-subtle));
}

.reward-transfer-button {
  border: 1px solid rgb(var(--color-border-strong));
  color: rgb(var(--color-foreground));
  background: rgb(var(--color-surface-raised));
  box-shadow: var(--shadow-card);
}

.reward-transfer-button:hover:not(:disabled) {
  background: rgb(var(--color-surface-subtle));
}
</style>
