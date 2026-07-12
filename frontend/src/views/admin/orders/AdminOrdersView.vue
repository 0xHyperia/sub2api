<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="commerce-toolbar flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-64">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="orderSearch"
              type="search"
              :placeholder="t('payment.admin.searchOrders')"
              :aria-label="t('payment.admin.searchOrders')"
              autocomplete="off"
              class="input pl-10"
              @input="debounceLoadOrders"
            />
          </div>
          <div class="w-full sm:w-40">
            <Select v-model="orderFilters.status" :options="statusFilterOptions" @change="loadOrders" />
          </div>
          <div class="w-full sm:w-44">
            <Select v-model="orderFilters.payment_type" :options="paymentTypeFilterOptions" @change="loadOrders" />
          </div>
          <div class="w-full sm:w-40">
            <Select v-model="orderFilters.order_type" :options="orderTypeFilterOptions" @change="loadOrders" />
          </div>
          <div class="ml-auto flex items-center">
            <button
              type="button"
              @click="loadOrders"
              :disabled="ordersLoading"
              class="btn btn-secondary px-2.5"
              :title="t('common.refresh')"
              :aria-label="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="ordersLoading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div v-if="ordersLoadError" class="orders-load-error" role="alert">
          <Icon name="exclamationTriangle" size="md" aria-hidden="true" />
          <p>{{ ordersLoadError }}</p>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="ordersLoading"
            @click="loadOrders"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': ordersLoading }" />
            {{ t('common.retry') }}
          </button>
        </div>
        <OrderTable
          v-if="orders.length > 0 || !ordersLoadError"
          :orders="orders"
          :loading="ordersLoading"
          show-user
        >
          <template #actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                type="button"
                @click="showOrderDetail(row)"
                class="order-action"
                :title="t('common.view')"
                :aria-label="t('common.view')"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                v-if="row.status === 'PENDING'"
                type="button"
                @click="handleCancelOrder(row)"
                class="order-action order-action-warning"
                :title="t('payment.orders.cancel')"
                :aria-label="t('payment.orders.cancel')"
              >
                <Icon name="x" size="sm" />
              </button>
              <button
                v-if="row.status === 'FAILED'"
                type="button"
                @click="handleRetryOrder(row)"
                class="order-action"
                :title="t('payment.admin.retry')"
                :aria-label="t('payment.admin.retry')"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <template v-if="row.status === 'REFUND_REQUESTED'">
                <span v-if="row.refund_amount" class="badge badge-warning tabular-nums">{{ creditedAmountSymbol }}{{ row.refund_amount.toFixed(2) }}</span>
                <button
                  type="button"
                  @click="openRefundDialog(row)"
                  class="order-action order-action-warning"
                  :title="t('payment.admin.approveRefund')"
                  :aria-label="t('payment.admin.approveRefund')"
                >
                  <Icon name="check" size="sm" />
                </button>
              </template>
              <button
                v-else-if="row.status === 'REFUND_FAILED'"
                type="button"
                @click="openRefundDialog(row)"
                class="order-action order-action-warning"
                :title="t('payment.admin.retryRefund')"
                :aria-label="t('payment.admin.retryRefund')"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                v-else-if="row.status === 'REFUND_PENDING'"
                type="button"
                :disabled="refundQueryingIds.has(row.id)"
                @click="handleQueryRefund(row)"
                class="order-action order-action-warning"
                :title="t('payment.admin.queryRefundStatus')"
                :aria-label="t('payment.admin.queryRefundStatus')"
              >
                <Icon name="refresh" size="sm" :class="refundQueryingIds.has(row.id) ? 'animate-spin' : ''" />
              </button>
              <button
                v-else-if="row.status === 'COMPLETED' || row.status === 'PARTIALLY_REFUNDED'"
                type="button"
                @click="openRefundDialog(row)"
                class="order-action order-action-danger"
                :title="t('payment.admin.refund')"
                :aria-label="t('payment.admin.refund')"
              >
                <Icon name="dollar" size="sm" />
              </button>
            </div>
          </template>
        </OrderTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="orderPagination.total > 0"
          :page="orderPagination.page"
          :total="orderPagination.total"
          :page-size="orderPagination.page_size"
          @update:page="handleOrderPageChange"
          @update:pageSize="handleOrderPageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog :show="showDetailDialog" :title="t('payment.admin.orderDetail')" width="wide" @close="showDetailDialog = false">
      <div v-if="selectedOrder" class="space-y-5">
        <dl class="order-detail-grid">
          <div><dt>{{ t('payment.orders.orderId') }}</dt><dd class="font-mono">#{{ selectedOrder.id }}</dd></div>
          <div><dt>{{ t('payment.orders.orderNo') }}</dt><dd class="break-all">{{ selectedOrder.out_trade_no }}</dd></div>
          <div><dt>{{ t('payment.orders.status') }}</dt><dd><OrderStatusBadge :status="selectedOrder.status" /></dd></div>
          <div><dt>{{ t('payment.orders.amount') }}</dt><dd class="tabular-nums">{{ creditedAmountSymbol }}{{ selectedOrder.amount.toFixed(2) }}</dd></div>
          <div><dt>{{ t('payment.orders.payAmount') }}</dt><dd class="tabular-nums">{{ paymentAmountSymbol(selectedOrder) }}{{ selectedOrder.pay_amount.toFixed(2) }}</dd></div>
          <div><dt>{{ t('payment.orders.paymentMethod') }}</dt><dd>{{ t('payment.methods.' + selectedOrder.payment_type, selectedOrder.payment_type) }}</dd></div>
          <div><dt>{{ t('payment.admin.feeRate') }}</dt><dd>{{ selectedOrder.fee_rate }}%</dd></div>
          <div><dt>{{ t('payment.orders.createdAt') }}</dt><dd>{{ formatDateTime(selectedOrder.created_at) }}</dd></div>
          <div><dt>{{ t('payment.admin.expiresAt') }}</dt><dd>{{ formatDateTime(selectedOrder.expires_at) }}</dd></div>
          <div v-if="selectedOrder.paid_at"><dt>{{ t('payment.admin.paidAt') }}</dt><dd>{{ formatDateTime(selectedOrder.paid_at) }}</dd></div>
          <div v-if="selectedOrder.refund_amount"><dt>{{ t('payment.admin.refundAmount') }}</dt><dd class="text-red-600 dark:text-red-400">{{ creditedAmountSymbol }}{{ selectedOrder.refund_amount.toFixed(2) }}</dd></div>
          <div v-if="selectedOrder.refund_reason" class="sm:col-span-2"><dt>{{ t('payment.admin.refundReason') }}</dt><dd>{{ selectedOrder.refund_reason }}</dd></div>
        </dl>

        <section v-if="selectedOrder.refund_requested_at" class="order-detail-section">
          <h3>{{ t('payment.admin.refundRequestInfo') }}</h3>
          <dl class="order-detail-grid mt-3">
            <div><dt>{{ t('payment.admin.refundRequestedAt') }}</dt><dd>{{ formatDateTime(selectedOrder.refund_requested_at) }}</dd></div>
            <div><dt>{{ t('payment.admin.refundRequestedBy') }}</dt><dd>#{{ selectedOrder.refund_requested_by }}</dd></div>
            <div class="sm:col-span-2"><dt>{{ t('payment.admin.refundRequestReason') }}</dt><dd>{{ selectedOrder.refund_request_reason }}</dd></div>
          </dl>
        </section>

        <section v-if="orderAuditLogs.length > 0" class="order-detail-section">
          <h3>{{ t('payment.admin.auditLogs') }}</h3>
          <ol class="order-audit-list mt-2">
            <li v-for="log in orderAuditLogs" :key="log.id">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <strong>{{ log.action }}</strong>
                <time>{{ formatDateTime(log.created_at) }}</time>
              </div>
              <p v-if="log.detail" class="break-all">{{ log.detail }}</p>
              <p v-if="log.operator">{{ t('payment.admin.operator') }}: {{ log.operator }}</p>
            </li>
          </ol>
        </section>
      </div>
    </BaseDialog>

    <AdminRefundDialog :show="showRefundDialog" :order="selectedOrder" :submitting="refundSubmitting" @confirm="handleRefund" @cancel="showRefundDialog = false" />
    <ConfirmDialog
      :show="cancelOrderTarget !== null"
      :title="t('payment.orders.cancel')"
      :message="t('payment.confirmCancel')"
      :confirm-text="t('payment.orders.cancel')"
      danger
      @confirm="confirmCancelOrder"
      @cancel="cancelOrderTarget = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminRefundDialog from '@/components/admin/payment/AdminRefundDialog.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import { currencySymbol } from '@/components/payment/currency'

interface AuditLog {
  id: number
  action: string
  detail: string | null
  operator: string | null
  created_at: string
}

const { t } = useI18n()
const appStore = useAppStore()

const ordersLoading = ref(false)
const ordersLoadError = ref('')
const orders = ref<PaymentOrder[]>([])
const orderSearch = ref('')
const orderFilters = reactive({ status: '', payment_type: '', order_type: '' })
const orderPagination = reactive({ page: 1, page_size: 20, total: 0 })
const selectedOrder = ref<PaymentOrder | null>(null)
const showDetailDialog = ref(false)
const showRefundDialog = ref(false)
const cancelOrderTarget = ref<PaymentOrder | null>(null)
const refundSubmitting = ref(false)
const refundQueryingIds = ref(new Set<number>())
const orderAuditLogs = ref<AuditLog[]>([])
const creditedAmountSymbol = currencySymbol('USD')

function paymentAmountSymbol(order: PaymentOrder | null | undefined): string {
  return currencySymbol(order?.currency)
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
let ordersLoadSequence = 0
function debounceLoadOrders() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadOrders(), 300)
}

interface OrdersQuery {
  page: number
  page_size: number
  keyword?: string
  status?: string
  payment_type?: string
  order_type?: string
}

function currentOrdersQuery(): OrdersQuery {
  return {
    page: orderPagination.page,
    page_size: orderPagination.page_size,
    keyword: orderSearch.value || undefined,
    status: orderFilters.status || undefined,
    payment_type: orderFilters.payment_type || undefined,
    order_type: orderFilters.order_type || undefined,
  }
}

function isCurrentOrdersQuery(query: OrdersQuery): boolean {
  const current = currentOrdersQuery()
  return Object.keys(current).every((key) => (
    current[key as keyof OrdersQuery] === query[key as keyof OrdersQuery]
  ))
}

async function loadOrders() {
  const currentSequence = ++ordersLoadSequence
  const query = currentOrdersQuery()
  ordersLoading.value = true
  ordersLoadError.value = ''
  try {
    const res = await adminPaymentAPI.getOrders(query)
    if (currentSequence !== ordersLoadSequence || !isCurrentOrdersQuery(query)) return
    orders.value = res.data.items || []
    orderPagination.total = res.data.total || 0
  } catch (err: unknown) {
    if (currentSequence !== ordersLoadSequence || !isCurrentOrdersQuery(query)) return
    const message = extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))
    ordersLoadError.value = message
    appStore.showError(message)
  } finally {
    if (currentSequence === ordersLoadSequence && isCurrentOrdersQuery(query)) {
      ordersLoading.value = false
    }
  }
}

function handleOrderPageChange(page: number) { orderPagination.page = page; loadOrders() }
function handleOrderPageSizeChange(size: number) { orderPagination.page_size = size; orderPagination.page = 1; loadOrders() }

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_PENDING', label: t('payment.status.refund_pending') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

async function showOrderDetail(order: PaymentOrder) {
  selectedOrder.value = order
  orderAuditLogs.value = []
  showDetailDialog.value = true
  try {
    const res = await adminPaymentAPI.getOrder(order.id)
    const data = res.data as unknown as Record<string, unknown>
    if (data.order) selectedOrder.value = data.order as PaymentOrder
    orderAuditLogs.value = ((data.auditLogs || data.audit_logs || []) as unknown) as AuditLog[]
  } catch (_err: unknown) { /* keep cached order data */ }
}

function handleCancelOrder(order: PaymentOrder) {
  cancelOrderTarget.value = order
}

async function confirmCancelOrder() {
  const order = cancelOrderTarget.value
  if (!order) return
  cancelOrderTarget.value = null
  try { await adminPaymentAPI.cancelOrder(order.id); appStore.showSuccess(t('payment.admin.orderCancelled')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function handleRetryOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.retryRecharge(order.id); appStore.showSuccess(t('payment.admin.retrySuccess')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

function openRefundDialog(order: PaymentOrder) { selectedOrder.value = order; showRefundDialog.value = true }

function isRefundPendingWarning(warning: string | undefined): boolean {
  return /pending|处理中|待/.test(String(warning || '').toLowerCase())
}

async function handleRefund(data: { amount: number; reason: string; deduct_balance: boolean; force: boolean }) {
  if (!selectedOrder.value) return
  refundSubmitting.value = true
  try {
    const res = await adminPaymentAPI.refundOrder(selectedOrder.value.id, { amount: data.amount, reason: data.reason, deduct_balance: data.deduct_balance, force: data.force })
    if (res.data.success) {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
      showRefundDialog.value = false
      loadOrders()
      return
    }
    if (isRefundPendingWarning(res.data.warning)) {
      appStore.showSuccess(t('payment.admin.refundPending'))
      showRefundDialog.value = false
      loadOrders()
      return
    }
    appStore.showError(res.data.warning || t('common.error'))
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { refundSubmitting.value = false }
}

async function handleQueryRefund(order: PaymentOrder) {
  refundQueryingIds.value = new Set(refundQueryingIds.value).add(order.id)
  try {
    const res = await adminPaymentAPI.queryRefund(order.id)
    if (res.data.success) {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
    } else if (isRefundPendingWarning(res.data.warning)) {
      appStore.showSuccess(t('payment.admin.refundPending'))
    } else {
      appStore.showError(res.data.warning || t('common.error'))
    }
    loadOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    const next = new Set(refundQueryingIds.value)
    next.delete(order.id)
    refundQueryingIds.value = next
  }
}

function formatDateTime(dateStr: string): string { return formatOrderDateTime(dateStr) }

onMounted(() => loadOrders())
</script>

<style scoped>
.commerce-toolbar {
  position: relative;
  padding: 12px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.orders-load-error {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid color-mix(in srgb, var(--ui-danger, #dc2626) 30%, transparent);
  color: var(--ui-danger-text, #b42318);
  background: var(--ui-danger-subtle, #fef3f2);
}

.orders-load-error p {
  min-width: 0;
  font-size: 13px;
  font-weight: 600;
  overflow-wrap: anywhere;
}

.order-action {
  display: inline-flex;
  width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  transition: color 150ms ease, background-color 150ms ease;
}

.order-action:hover {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.order-action-warning:hover {
  color: rgb(var(--color-warning-foreground, 180 83 9));
  background: rgb(var(--color-warning-subtle, 255 251 235));
}

.order-action-danger:hover {
  color: rgb(var(--color-danger-foreground, 185 28 28));
  background: rgb(var(--color-danger-subtle, 254 242 242));
}

.order-action:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: 1px;
}

.order-action:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.order-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px 24px;
}

.order-detail-grid dt {
  margin-bottom: 4px;
  color: var(--ui-text-muted, #667085);
  font-size: 12px;
  font-weight: 500;
}

.order-detail-grid dd {
  color: var(--ui-text, #0f172a);
  font-size: 14px;
  font-weight: 500;
}

.order-detail-section {
  padding-top: 16px;
  border-top: 1px solid var(--ui-border, #dbe3ee);
}

.order-detail-section h3 {
  color: var(--ui-text, #0f172a);
  font-size: 13px;
  font-weight: 600;
}

.order-audit-list {
  max-height: 224px;
  overflow-y: auto;
  border-top: 1px solid var(--ui-border, #dbe3ee);
}

.order-audit-list li {
  padding: 12px 0;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
  color: var(--ui-text-muted, #667085);
  font-size: 12px;
}

.order-audit-list strong {
  color: var(--ui-text, #0f172a);
  font-weight: 600;
}

.order-audit-list p {
  margin-top: 4px;
}

@media (max-width: 639px) {
  .commerce-toolbar {
    padding: 10px;
  }

  .order-detail-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .orders-load-error {
    grid-template-columns: auto minmax(0, 1fr);
    padding: 12px;
  }

  .orders-load-error .btn {
    grid-column: 1 / -1;
    width: 100%;
  }
}
</style>
