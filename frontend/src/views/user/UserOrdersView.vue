<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-4">
      <header class="page-header flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0">
          <h1 class="page-title">{{ t('payment.orders.title') }}</h1>
        </div>
        <button type="button" class="btn btn-primary w-full sm:w-auto" @click="router.push('/purchase')">
          <Icon name="plus" size="sm" aria-hidden="true" />
          <span>{{ t('payment.result.backToRecharge') }}</span>
        </button>
      </header>

      <div class="flex flex-col gap-3 rounded-panel border border-outline bg-surface p-3 shadow-card sm:flex-row sm:items-center sm:justify-between">
        <Select
          v-model="currentFilter"
          :options="statusFilters"
          :placeholder="t('payment.orders.status')"
          class="w-full sm:w-44"
          @change="handleFilterChange"
        />
        <div class="flex items-center justify-end gap-2">
            <button
              type="button"
              class="btn btn-secondary btn-icon"
              :disabled="loading"
              :title="t('common.refresh')"
              :aria-label="t('common.refresh')"
              @click="fetchOrders"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
            </button>
        </div>
      </div>

      <div v-if="fetchFailed" class="flex flex-col gap-3 rounded-panel border border-danger/20 bg-danger-subtle px-4 py-3 text-sm text-danger-foreground sm:flex-row sm:items-center sm:justify-between" role="alert">
        <span>{{ t('common.error') }}. {{ t('errors.tryAgain') }}</span>
        <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="fetchOrders">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" aria-hidden="true" />
          <span>{{ t('common.refresh') }}</span>
        </button>
      </div>

      <section class="min-w-0 lg:overflow-hidden lg:rounded-panel lg:border lg:border-outline lg:bg-surface lg:shadow-card" :aria-label="t('payment.orders.title')">
        <OrderTable :orders="orders" :loading="loading">
          <template #actions="{ row }">
            <div class="flex flex-wrap items-center justify-end gap-1.5">
              <button
                v-if="row.status === 'PENDING'"
                type="button"
                class="btn btn-ghost btn-sm text-warning-foreground"
                @click="handleCancel(row.id)"
              >
                <Icon name="x" size="sm" aria-hidden="true" />
                <span>{{ t('payment.orders.cancel') }}</span>
              </button>
              <button
                v-if="canRequestRefund(row)"
                type="button"
                class="btn btn-ghost btn-sm"
                @click="openRefundDialog(row)"
              >
                <Icon name="dollar" size="sm" aria-hidden="true" />
                <span>{{ t('payment.orders.requestRefund') }}</span>
              </button>
            </div>
          </template>
        </OrderTable>
      </section>

      <div class="flex justify-end">
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </div>
    </div>

    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
      <p class="text-sm leading-6 text-foreground-muted">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="actionLoading" @click="cancelTargetId = null">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">
          <Icon v-if="actionLoading" name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
          <span>{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</span>
        </button>
      </template>
    </BaseDialog>

    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" width="narrow" @close="refundTarget = null">
      <div v-if="refundTarget" class="space-y-4">
        <dl class="divide-y divide-outline rounded-panel border border-outline bg-surface-subtle px-4">
          <div class="flex items-center justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-subtle">{{ t('payment.orders.orderId') }}</dt>
            <dd class="font-mono text-foreground">#{{ refundTarget.id }}</dd>
          </div>
          <div class="flex items-center justify-between gap-4 py-3 text-sm">
            <dt class="text-foreground-subtle">{{ t('payment.orders.amount') }}</dt>
            <dd class="tabular-nums text-foreground">${{ refundTarget.amount.toFixed(2) }}</dd>
          </div>
        </dl>
        <div>
          <label for="refund-reason" class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea
            id="refund-reason"
            v-model="refundReason"
            rows="4"
            class="input w-full resize-y"
            :placeholder="t('payment.refundReasonPlaceholder')"
            :disabled="actionLoading"
            required
          />
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" :disabled="actionLoading" @click="refundTarget = null">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-primary" :disabled="actionLoading || !refundReason.trim()" @click="confirmRefund">
          <Icon v-if="actionLoading" name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
          <span>{{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}</span>
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const fetchFailed = ref(false)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

async function fetchOrders() {
  loading.value = true
  fetchFailed.value = false
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    fetchFailed.value = true
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handleFilterChange() {
  pagination.page = 1
  void fetchOrders()
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

function handleCancel(orderId: number) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { refundTarget.value = order; refundReason.value = '' }

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(res.data.provider_instance_ids || [])
  } catch { /* ignore — default to hiding refund button */ }
}

onMounted(() => { fetchOrders(); loadRefundEligibility() })
</script>
