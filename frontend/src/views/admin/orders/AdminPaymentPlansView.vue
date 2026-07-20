<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="commerce-toolbar flex items-center justify-end gap-2">
          <button
            type="button"
            @click="loadPlans"
            :disabled="plansLoading"
            class="btn btn-secondary px-2.5"
            :title="t('common.refresh')"
            :aria-label="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="plansLoading ? 'animate-spin' : ''" />
          </button>
          <button type="button" @click="openPlanEdit(null)" class="btn btn-primary">
            <Icon name="plus" size="sm" />
            {{ t('payment.admin.createPlan') }}
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="planColumns" :data="plans" :loading="plansLoading">
          <template #mobile-card="{ row }">
            <article class="space-y-3" :data-test="`payment-plan-mobile-card-${row.id}`">
              <header class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0">
                  <h3 class="truncate text-sm font-semibold" :class="getPlanNameClass(row.group_id)" :title="row.name">{{ row.name }}</h3>
                  <p class="mt-0.5 font-mono text-xs text-foreground-subtle">#{{ row.id }}</p>
                </div>
                <span :class="['badge flex-none', row.for_sale ? 'badge-success' : 'badge-gray']">
                  {{ row.for_sale ? t('payment.admin.onSale') : t('payment.admin.offSale') }}
                </span>
              </header>

              <dl class="grid grid-cols-2 overflow-hidden rounded-panel border border-outline bg-surface-subtle">
                <div class="min-w-0 px-3 py-2.5">
                  <dt class="text-[11px] text-foreground-subtle">{{ t('payment.admin.price') }}</dt>
                  <dd class="mt-1 text-base font-semibold tabular-nums text-foreground">
                    {{ planCurrencySymbol(row.currency) }}{{ (row.price ?? 0).toFixed(2) }}
                    <span v-if="row.currency" class="text-xs font-normal text-foreground-subtle">{{ row.currency }}</span>
                  </dd>
                  <p v-if="row.original_price" class="mt-0.5 text-xs text-foreground-subtle line-through">{{ planCurrencySymbol(row.currency) }}{{ row.original_price.toFixed(2) }}</p>
                </div>
                <div class="min-w-0 border-l border-outline px-3 py-2.5">
                  <dt class="text-[11px] text-foreground-subtle">{{ t('payment.admin.validity') }}</dt>
                  <dd class="mt-1 text-sm font-semibold tabular-nums text-foreground">
                    {{ row.validity_days }} {{ t('payment.admin.' + (row.validity_unit || 'days')) }}
                  </dd>
                </div>
              </dl>

              <div class="rounded-control border border-outline px-3 py-2.5">
                <p class="mb-1.5 text-[11px] text-foreground-subtle">{{ t('payment.admin.group') }}</p>
                <div v-if="isGroupMissing(row.group_id)" class="flex items-center gap-1.5 text-sm">
                  <span class="text-foreground-subtle">#{{ row.group_id }}</span>
                  <span class="badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
                </div>
                <GroupBadge
                  v-else-if="getGroup(row.group_id)"
                  :name="getGroup(row.group_id)!.name"
                  :platform="getGroup(row.group_id)!.platform"
                  :rate-multiplier="getGroup(row.group_id)!.rate_multiplier"
                />
                <span v-else class="text-sm text-foreground-subtle">-</span>
              </div>

              <footer class="flex items-center gap-2 border-t border-outline pt-3">
                <button type="button" class="btn btn-secondary min-w-0 flex-1" :data-test="`payment-plan-mobile-edit-${row.id}`" @click="openPlanEdit(row)">
                  <Icon name="edit" size="sm" />
                  {{ t('common.edit') }}
                </button>
                <details class="group/menu relative" @click.stop>
                  <summary
                    class="inline-flex h-10 w-10 cursor-pointer list-none items-center justify-center rounded-control border border-outline text-foreground-muted hover:bg-surface-subtle hover:text-foreground"
                    :title="t('common.actions')"
                    :aria-label="`${t('common.actions')}: ${row.name}`"
                    :data-test="`payment-plan-mobile-more-${row.id}`"
                  >
                    <Icon name="more" size="sm" />
                  </summary>
                  <div class="absolute bottom-full right-0 z-30 mb-1 w-48 overflow-hidden rounded-panel border border-outline bg-surface py-1 shadow-floating">
                    <button type="button" class="flex min-h-10 w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground-muted hover:bg-surface-subtle hover:text-foreground" :data-test="`payment-plan-mobile-toggle-${row.id}`" @click="toggleForSale(row)">
                      <Icon :name="row.for_sale ? 'xCircle' : 'checkCircle'" size="sm" />
                      {{ row.for_sale ? t('payment.admin.offSale') : t('payment.admin.onSale') }}
                    </button>
                    <button type="button" class="flex min-h-10 w-full items-center gap-2 px-3 py-2 text-left text-sm text-danger-foreground hover:bg-danger-subtle" :data-test="`payment-plan-mobile-delete-${row.id}`" @click="confirmDeletePlan(row)">
                      <Icon name="trash" size="sm" />
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </details>
              </footer>
            </article>
          </template>
          <template #cell-name="{ value, row }">
            <span class="text-sm font-medium" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
          </template>
          <template #cell-group_id="{ value }">
            <span v-if="isGroupMissing(value)" class="text-sm">
              <span class="text-foreground-subtle">#{{ value }}</span>
              <span class="ml-1 badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
            </span>
            <GroupBadge
              v-else-if="getGroup(value)"
              :name="getGroup(value)!.name"
              :platform="getGroup(value)!.platform"
              :rate-multiplier="getGroup(value)!.rate_multiplier"
            />
            <span v-else class="text-sm text-foreground-subtle">-</span>
          </template>
          <template #cell-price="{ value, row }">
            <div class="text-sm tabular-nums">
              <span class="font-semibold text-foreground">{{ planCurrencySymbol(row.currency) }}{{ (value ?? 0).toFixed(2) }}</span>
              <span v-if="row.currency" class="ml-1 text-xs text-foreground-subtle">{{ row.currency }}</span>
              <span v-if="row.original_price" class="ml-1 text-xs text-foreground-subtle line-through">{{ planCurrencySymbol(row.currency) }}{{ row.original_price.toFixed(2) }}</span>
            </div>
          </template>
          <template #cell-validity_days="{ value, row }">
            <span class="text-sm">{{ value }} {{ t('payment.admin.' + (row.validity_unit || 'days')) }}</span>
          </template>
          <template #cell-for_sale="{ value, row }">
            <Toggle
              :model-value="Boolean(value)"
              :aria-label="`${t('payment.admin.forSale')}: ${row.name}`"
              @update:model-value="toggleForSale(row)"
            />
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                type="button"
                @click="openPlanEdit(row)"
                class="plan-action"
                :title="t('common.edit')"
                :aria-label="`${t('common.edit')}: ${row.name}`"
              >
              <Icon name="edit" size="sm" />
              </button>
              <button
                type="button"
                @click="confirmDeletePlan(row)"
                class="plan-action plan-action-danger"
                :title="t('common.delete')"
                :aria-label="`${t('common.delete')}: ${row.name}`"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
      </template>
    </TablePageLayout>

    <!-- Plan Edit Dialog -->
    <PlanEditDialog :show="showPlanDialog" :plan="editingPlan" :groups="groups" :payment-config="paymentConfig" @close="showPlanDialog = false" @saved="loadPlans" />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { AdminPaymentConfig } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import adminAPI from '@/api/admin'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlanEditDialog from './PlanEditDialog.vue'
import { currencySymbol } from '@/components/payment/currency'
import { platformTextClass } from '@/utils/platformColors'

const { t } = useI18n()
const appStore = useAppStore()

function planCurrencySymbol(currency?: string): string {
  return currencySymbol(currency || 'USD')
}

// ==================== Groups ====================

const groups = ref<AdminGroup[]>([])
const paymentConfig = ref<AdminPaymentConfig | null>(null)

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

async function loadPaymentConfig() {
  try {
    const res = await adminPaymentAPI.getConfig()
    paymentConfig.value = res.data
  } catch { /* preview only */ }
}

function getGroup(id: number): AdminGroup | undefined {
  return groups.value.find(g => g.id === id)
}

function isGroupMissing(id: number): boolean {
  return id > 0 && !groups.value.find(g => g.id === id)
}

function getPlanNameClass(groupId: number): string {
  const group = getGroup(groupId)
  return group ? platformTextClass(group.platform) : 'text-foreground'
}


// ==================== Plans ====================

const plansLoading = ref(false)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<number | null>(null)

const planColumns = computed((): Column[] => [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('payment.admin.planName') },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price') },
  { key: 'validity_days', label: t('payment.admin.validity') },
  { key: 'for_sale', label: t('payment.admin.forSale') },
  { key: 'sort_order', label: t('payment.admin.sortOrder') },
  { key: 'actions', label: t('common.actions') },
])

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    // Backend returns features as newline-separated string; parse to array
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { plansLoading.value = false }
}

function openPlanEdit(plan: SubscriptionPlan | null) {
  editingPlan.value = plan
  showPlanDialog.value = true
}


/** Quick toggle for_sale from the list */
async function toggleForSale(plan: SubscriptionPlan) {
  try {
    await adminPaymentAPI.updatePlan(plan.id, { for_sale: !plan.for_sale })
    plan.for_sale = !plan.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDeletePlan(plan: SubscriptionPlan) { deletingPlanId.value = plan.id; showDeletePlanDialog.value = true }
async function handleDeletePlan() {
  if (!deletingPlanId.value) return
  try { await adminPaymentAPI.deletePlan(deletingPlanId.value); appStore.showSuccess(t('common.deleted')); showDeletePlanDialog.value = false; loadPlans() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadGroups()
  loadPaymentConfig()
  loadPlans()
})
</script>

<style scoped>
.commerce-toolbar {
  padding: 12px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.plan-action {
  display: inline-flex;
  width: 40px;
  height: 40px;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  transition: color 150ms ease, background-color 150ms ease;
}

.plan-action:hover {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.plan-action-danger:hover {
  color: rgb(var(--color-danger-foreground, 185 28 28));
  background: rgb(var(--color-danger-subtle, 254 242 242));
}

.plan-action:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: 1px;
}

@media (max-width: 639px) {
  .commerce-toolbar {
    padding: 10px;
  }
}
</style>
