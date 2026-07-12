<template>
  <BaseDialog :show="show" :title="t('admin.users.balanceHistoryTitle')" width="wide" :close-on-click-outside="true" :z-index="40" @close="handleClose">
    <div v-if="user" class="space-y-4" :aria-busy="loading">
      <!-- User header: two-row layout with full user info -->
      <div class="rounded-panel border border-outline bg-surface-subtle p-3 sm:p-4">
        <!-- Row 1: avatar + email/username/created_at (left) + current balance (right) -->
        <div class="grid grid-cols-[auto_minmax(0,1fr)] items-center gap-x-3 gap-y-2 sm:grid-cols-[auto_minmax(0,1fr)_auto]">
          <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-control border border-outline bg-surface" aria-hidden="true">
            <span class="text-sm font-semibold text-foreground-muted">
              {{ user.email.charAt(0).toUpperCase() }}
            </span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <p class="min-w-0 break-all text-sm font-medium text-foreground">{{ user.email }}</p>
              <span v-if="user.deleted_at" class="badge badge-danger flex-shrink-0 text-[10px]">
                {{ t('admin.usage.userDeletedBadge') }}
              </span>
              <span
                v-if="user.username"
                class="badge badge-gray max-w-full"
              >
                <span class="min-w-0 break-all">{{ user.username }}</span>
              </span>
            </div>
            <p class="mt-0.5 break-words text-xs text-foreground-subtle">
              {{ t('admin.users.createdAt') }}: {{ formatDateTime(user.created_at) }}
            </p>
          </div>
          <!-- Current balance: prominent display on the right -->
          <div class="col-start-2 text-left sm:col-start-auto sm:text-right">
            <p class="text-xs text-foreground-subtle">{{ t('admin.users.currentBalance') }}</p>
            <p class="break-all text-lg font-semibold tabular-nums text-foreground">
              ${{ user.balance?.toFixed(2) || '0.00' }}
            </p>
          </div>
        </div>
        <!-- Row 2: notes + total recharged -->
        <div class="mt-2.5 flex flex-col items-start gap-1 border-t border-outline pt-2.5 sm:flex-row sm:items-center sm:justify-between">
          <p class="min-w-0 flex-1 break-words text-xs text-foreground-subtle" :title="user.notes || ''">
            <template v-if="user.notes">{{ t('admin.users.notes') }}: {{ user.notes }}</template>
            <template v-else>&nbsp;</template>
          </p>
          <p class="min-w-0 max-w-full break-all text-xs text-foreground-subtle sm:ml-4 sm:shrink-0">
            {{ t('admin.users.totalRecharged') }}:
            <span v-if="loadSucceeded" class="font-semibold tabular-nums text-success-foreground">${{ totalRecharged.toFixed(2) }}</span>
            <span v-else>--</span>
          </p>
        </div>
      </div>

      <!-- Type filter + Action buttons -->
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
        <Select
          v-model="typeFilter"
          :options="typeOptions"
          :aria-label="t('admin.users.allTypes')"
          class="w-full sm:w-56 sm:flex-none"
          @change="loadHistory(1)"
        />
        <div v-if="!hideActions" class="grid grid-cols-2 gap-2 sm:flex sm:items-center">
          <!-- Deposit button - matches menu style -->
          <button
            type="button"
            @click="emit('deposit')"
            class="btn btn-secondary min-h-touch min-w-0 sm:min-h-control"
          >
            <Icon name="plus" size="sm" class="shrink-0 text-success-foreground" :stroke-width="2" aria-hidden="true" />
            {{ t('admin.users.deposit') }}
          </button>
          <!-- Withdraw button - matches menu style -->
          <button
            type="button"
            @click="emit('withdraw')"
            class="btn btn-secondary min-h-touch min-w-0 sm:min-h-control"
          >
            <span class="text-lg leading-none text-warning-foreground" aria-hidden="true">-</span>
            {{ t('admin.users.withdraw') }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center gap-2 py-8 text-sm text-foreground-subtle" role="status">
        <Icon name="refresh" size="md" class="animate-spin" aria-hidden="true" />
        <span>{{ t('common.loading') }}</span>
      </div>

      <!-- Error state -->
      <div v-else-if="loadError" class="flex flex-col items-center gap-3 py-8 text-center" role="alert">
        <p class="text-sm text-danger-foreground">{{ loadError }}</p>
        <button type="button" class="btn btn-secondary" data-testid="balance-history-retry" @click="loadHistory(currentPage)">
          {{ t('admin.users.retry') }}
        </button>
      </div>

      <!-- Empty state -->
      <div v-else-if="history.length === 0" class="py-8 text-center">
        <p class="text-sm text-foreground-subtle">{{ t('admin.users.noBalanceHistory') }}</p>
      </div>

      <!-- History list -->
      <div v-else class="max-h-[28rem] divide-y divide-outline overflow-y-auto rounded-panel border border-outline bg-surface">
        <div
          v-for="item in history"
          :key="item.id"
          class="p-3 sm:p-4"
        >
          <div class="grid grid-cols-[minmax(0,1fr)_minmax(0,8rem)] items-start gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(8rem,auto)]">
            <!-- Left: type icon + description -->
            <div class="flex min-w-0 items-start gap-3">
              <div
                :class="[
                  'flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-control',
                  getIconBg(item)
                ]"
              >
                <Icon :name="getIconName(item)" size="sm" :class="getIconColor(item)" aria-hidden="true" />
              </div>
              <div class="min-w-0">
                <p class="break-words text-sm font-medium text-foreground">
                  {{ getItemTitle(item) }}
                </p>
                <!-- Notes (admin adjustment reason) -->
                <p
                  v-if="item.notes"
                  class="mt-0.5 break-words text-xs text-foreground-subtle"
                  :title="item.notes"
                >
                  {{ item.notes.length > 60 ? item.notes.substring(0, 55) + '...' : item.notes }}
                </p>
                <p class="mt-0.5 break-words text-xs text-foreground-subtle">
                  {{ formatDateTime(item.used_at || item.created_at) }}
                </p>
              </div>
            </div>
            <!-- Right: value -->
            <div class="min-w-0 text-right">
              <p :class="['break-all text-sm font-semibold tabular-nums', getValueColor(item)]">
                {{ formatValue(item) }}
              </p>
              <p
                v-if="isAdminType(item.type)"
                class="break-words text-xs text-foreground-subtle"
              >
                {{ t('redeem.adminAdjustment') }}
              </p>
              <p
                v-else
                class="break-all font-mono text-xs text-foreground-subtle"
              >
                {{ item.code.slice(0, 8) }}...
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 pt-2">
        <button
          type="button"
          :disabled="currentPage <= 1"
          class="btn btn-secondary min-h-touch px-3 text-sm sm:min-h-control"
          @click="loadHistory(currentPage - 1)"
        >
          {{ t('pagination.previous') }}
        </button>
        <span class="whitespace-nowrap text-sm tabular-nums text-foreground-subtle">
          {{ currentPage }} / {{ totalPages }}
        </span>
        <button
          type="button"
          :disabled="currentPage >= totalPages"
          class="btn btn-secondary min-h-touch px-3 text-sm sm:min-h-control"
          @click="loadHistory(currentPage + 1)"
        >
          {{ t('pagination.next') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type BalanceHistoryItem } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null; hideActions?: boolean }>()
const emit = defineEmits(['close', 'deposit', 'withdraw'])
const { t } = useI18n()

const history = ref<BalanceHistoryItem[]>([])
const loading = ref(false)
const loadError = ref('')
const loadSucceeded = ref(false)
const currentPage = ref(1)
const total = ref(0)
const totalRecharged = ref(0)
const pageSize = 15
const typeFilter = ref('')
let loadRequestSeq = 0

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

// Type filter options
const typeOptions = computed(() => [
  { value: '', label: t('admin.users.allTypes') },
  { value: 'balance', label: t('admin.users.typeBalance') },
  { value: 'affiliate_balance', label: t('admin.users.typeAffiliateBalance') },
  { value: 'admin_balance', label: t('admin.users.typeAdminBalance') },
  { value: 'concurrency', label: t('admin.users.typeConcurrency') },
  { value: 'admin_concurrency', label: t('admin.users.typeAdminConcurrency') },
  { value: 'subscription', label: t('admin.users.typeSubscription') }
])

function clearHistory(resetFilter = false) {
  history.value = []
  currentPage.value = 1
  total.value = 0
  totalRecharged.value = 0
  loading.value = false
  loadError.value = ''
  loadSucceeded.value = false
  if (resetFilter) typeFilter.value = ''
}

function isCurrentLoad(requestSeq: number, userId: number): boolean {
  return requestSeq === loadRequestSeq && props.show && props.user?.id === userId
}

const loadHistory = async (page: number) => {
  const userId = props.user?.id
  if (userId == null || !props.show) return
  const requestSeq = ++loadRequestSeq
  history.value = []
  total.value = 0
  totalRecharged.value = 0
  loadError.value = ''
  loadSucceeded.value = false
  loading.value = true
  currentPage.value = page
  try {
    const res = await adminAPI.users.getUserBalanceHistory(
      userId,
      page,
      pageSize,
      typeFilter.value || undefined
    )
    if (!isCurrentLoad(requestSeq, userId)) return
    history.value = res.items || []
    total.value = res.total || 0
    totalRecharged.value = res.total_recharged || 0
    loadSucceeded.value = true
  } catch (error) {
    if (!isCurrentLoad(requestSeq, userId)) return
    console.error('Failed to load balance history:', error)
    loadError.value = t('admin.users.failedToLoadBalanceHistory')
  } finally {
    if (isCurrentLoad(requestSeq, userId)) loading.value = false
  }
}

watch(
  [() => props.show, () => props.user?.id],
  ([show, userId]) => {
    loadRequestSeq += 1
    clearHistory(true)
    if (show && userId != null) void loadHistory(1)
  },
  { immediate: true }
)

function handleClose() {
  loadRequestSeq += 1
  clearHistory(true)
  emit('close')
}

onUnmounted(() => {
  loadRequestSeq += 1
})

// Helper: check if admin type
const isAdminType = (type: string) => type === 'admin_balance' || type === 'admin_concurrency'

// Helper: check if balance type (includes admin_balance)
const isBalanceType = (type: string) => type === 'balance' || type === 'admin_balance' || type === 'affiliate_balance'

// Helper: check if subscription type
const isSubscriptionType = (type: string) => type === 'subscription'

// Icon name based on type
const getIconName = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) return 'dollar'
  if (isSubscriptionType(item.type)) return 'badge'
  return 'bolt' // concurrency
}

// Icon background color
const getIconBg = (item: BalanceHistoryItem) => {
  if (isSubscriptionType(item.type)) return 'bg-info-subtle'
  return item.value >= 0 ? 'bg-success-subtle' : 'bg-danger-subtle'
}

// Icon text color
const getIconColor = (item: BalanceHistoryItem) => {
  if (isSubscriptionType(item.type)) return 'text-info-foreground'
  return item.value >= 0 ? 'text-success-foreground' : 'text-danger-foreground'
}

// Value text color
const getValueColor = (item: BalanceHistoryItem) => {
  if (isSubscriptionType(item.type)) return 'text-info-foreground'
  return item.value >= 0 ? 'text-success-foreground' : 'text-danger-foreground'
}

// Item title
const getItemTitle = (item: BalanceHistoryItem) => {
  switch (item.type) {
    case 'balance':
      return t('redeem.balanceAddedRedeem')
    case 'affiliate_balance':
      return t('redeem.balanceAddedAffiliate')
    case 'admin_balance':
      return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
    case 'concurrency':
      return t('redeem.concurrencyAddedRedeem')
    case 'admin_concurrency':
      return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
    case 'subscription':
      return t('redeem.subscriptionAssigned')
    default:
      return t('common.unknown')
  }
}

// Format display value
const formatValue = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  }
  if (isSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}d - ${groupName}` : `${days}d`
  }
  // concurrency types
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value}`
}
</script>
