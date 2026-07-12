<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="page-header">
        <h1 class="page-title">{{ t('affiliate.title') }}</h1>
        <p class="page-description">{{ t('affiliate.description') }}</p>
      </header>

      <div v-if="loading" class="card flex min-h-48 items-center justify-center" aria-live="polite">
        <LoadingSpinner />
      </div>

      <div v-else-if="loadFailed && !detail" class="card" role="alert">
        <EmptyState
          :title="t('affiliate.loadFailed')"
          :description="t('errors.tryAgain')"
          :action-text="t('common.refresh')"
          @action="loadAffiliateDetail()"
        />
      </div>

      <template v-else-if="detail">
        <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4" :aria-label="t('affiliate.title')">
          <article class="card min-w-0 p-4 sm:p-5">
            <p class="flex items-center gap-2 text-sm text-foreground-subtle">
              <Icon name="dollar" size="sm" class="text-info-foreground" aria-hidden="true" />
              {{ t('affiliate.stats.rebateRate') }}
            </p>
            <p class="mt-2 text-2xl font-semibold tabular-nums text-foreground">
              {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
            </p>
            <p class="mt-1 text-xs leading-5 text-foreground-subtle">
              {{ t('affiliate.stats.rebateRateHint') }}
            </p>
          </article>
          <article class="card min-w-0 p-4 sm:p-5">
            <p class="text-sm text-foreground-subtle">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold tabular-nums text-foreground">
              {{ formatCount(detail.aff_count) }}
            </p>
          </article>
          <article class="card min-w-0 p-4 sm:p-5">
            <p class="text-sm text-foreground-subtle">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="mt-2 break-words text-2xl font-semibold tabular-nums text-success-foreground">
              {{ formatCurrency(detail.aff_quota) }}
            </p>
          </article>
          <article class="card min-w-0 p-4 sm:p-5">
            <p class="text-sm text-foreground-subtle">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="mt-2 break-words text-2xl font-semibold tabular-nums text-foreground">
              {{ formatCurrency(detail.aff_history_quota) }}
            </p>
            <p v-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-warning-foreground">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
          </article>
        </section>

        <section class="card p-4 sm:p-5" aria-labelledby="affiliate-share-title">
          <h2 id="affiliate-share-title" class="text-base font-semibold text-foreground">{{ t('affiliate.yourCode') }}</h2>
          <div class="mt-4 grid gap-4 xl:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-foreground-muted">{{ t('affiliate.yourCode') }}</p>
              <div class="flex min-w-0 items-center gap-2 rounded-panel border border-outline-strong bg-surface-subtle p-2 pl-3">
                <code class="min-w-0 flex-1 break-all text-sm font-semibold text-foreground">{{ detail.aff_code }}</code>
                <button type="button" class="btn btn-secondary btn-icon shrink-0" :title="t('affiliate.copyCode')" :aria-label="t('affiliate.copyCode')" @click="copyCode">
                  <Icon name="copy" size="sm" aria-hidden="true" />
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-foreground-muted">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex min-w-0 items-center gap-2 rounded-panel border border-outline-strong bg-surface-subtle p-2 pl-3">
                <code class="min-w-0 flex-1 truncate text-sm text-foreground-muted" :title="inviteLink">{{ inviteLink }}</code>
                <button type="button" class="btn btn-secondary btn-icon shrink-0" :title="t('affiliate.copyLink')" :aria-label="t('affiliate.copyLink')" @click="copyInviteLink">
                  <Icon name="copy" size="sm" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>

          <aside class="mt-4 border-t border-outline pt-4">
            <p class="text-sm font-medium text-foreground">{{ t('affiliate.tips.title') }}</p>
            <ol class="mt-2 list-decimal space-y-1 pl-5 text-sm leading-6 text-foreground-muted">
              <li>{{ t('affiliate.tips.line1') }}</li>
              <li>{{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>{{ t('affiliate.tips.line3') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">{{ t('affiliate.tips.line4') }}</li>
            </ol>
          </aside>
        </section>

        <section class="card p-4 sm:p-5" aria-labelledby="affiliate-transfer-title">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="min-w-0">
              <h2 id="affiliate-transfer-title" class="text-base font-semibold text-foreground">{{ t('affiliate.transfer.title') }}</h2>
              <p class="mt-1 text-sm leading-6 text-foreground-subtle">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button
              type="button"
              class="btn btn-primary w-full shrink-0 sm:w-auto"
              :disabled="transferring || detail.aff_quota <= 0"
              :aria-busy="transferring"
              @click="transferQuota"
            >
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
              <Icon v-else name="dollar" size="sm" aria-hidden="true" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="mt-3 text-sm text-warning-foreground">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </section>

        <section aria-labelledby="affiliate-invitees-title">
          <div class="mb-3 flex items-center justify-between gap-3">
            <h2 id="affiliate-invitees-title" class="text-base font-semibold text-foreground">{{ t('affiliate.invitees.title') }}</h2>
            <span class="badge badge-gray tabular-nums">{{ formatCount(detail.invitees.length) }}</span>
          </div>
          <div class="min-w-0 lg:overflow-hidden lg:rounded-panel lg:border lg:border-outline lg:bg-surface lg:shadow-card">
            <DataTable :columns="inviteeColumns" :data="detail.invitees" :loading="false" row-key="user_id">
              <template #cell-email="{ value }">
                <span class="break-all text-foreground">{{ value || '-' }}</span>
              </template>
              <template #cell-username="{ value }">
                <span class="break-words text-foreground-muted">{{ value || '-' }}</span>
              </template>
              <template #cell-total_rebate="{ value }">
                <span class="font-medium tabular-nums text-success-foreground">{{ formatCurrency(value) }}</span>
              </template>
              <template #cell-created_at="{ value }">
                <span class="whitespace-nowrap text-foreground-muted">{{ formatDateTime(value) || '-' }}</span>
              </template>
              <template #empty>
                <EmptyState :title="t('affiliate.invitees.empty')" />
              </template>
            </DataTable>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const loadFailed = ref(false)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const inviteeColumns = computed((): Column[] => [
  { key: 'email', label: t('affiliate.invitees.columns.email') },
  { key: 'username', label: t('affiliate.invitees.columns.username') },
  { key: 'total_rebate', label: t('affiliate.invitees.columns.rebate') },
  { key: 'created_at', label: t('affiliate.invitees.columns.joinedAt') },
])

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

// Rebate rate is a percentage in the range [0, 100]; backend already clamps it.
// We trim trailing zeros (e.g. 20.00 → "20", 12.50 → "12.5") for a cleaner UI.
const formattedRebateRate = computed(() => {
  const v = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(v * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

function formatCount(value: number): string {
  return value.toLocaleString()
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) {
    loading.value = true
    loadFailed.value = false
  }
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    loadFailed.value = detail.value === null
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const resp = await userAPI.transferAffiliateQuota()
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(resp.transferred_quota) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined),
    ])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>
