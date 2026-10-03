<template>
  <section data-testid="recharge-value-estimator" class="border-t border-outline pt-5" :aria-labelledby="titleId">
    <button
      ref="triggerButton"
      type="button"
      class="group flex w-full items-center justify-between gap-4 text-left"
      :aria-expanded="drawerOpen"
      :aria-controls="drawerId"
      @click="drawerOpen = true"
    >
      <span class="flex min-w-0 items-center gap-3">
        <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-info-subtle text-info-foreground transition-colors group-hover:bg-info/15">
          <Icon name="calculator" size="sm" aria-hidden="true" />
        </span>
        <span class="min-w-0">
          <span :id="titleId" class="block text-sm font-semibold text-foreground">{{ t('payment.estimator.title') }}</span>
          <span class="mt-0.5 block text-xs leading-5 text-foreground-subtle">{{ t('payment.estimator.summary') }}</span>
        </span>
      </span>
      <span class="flex shrink-0 items-center gap-1 text-xs font-medium text-foreground-muted transition-colors group-hover:text-foreground">
        <span class="hidden sm:inline">{{ t('payment.estimator.open') }}</span>
        <Icon name="chevronRight" size="sm" aria-hidden="true" />
      </span>
    </button>

    <Teleport to="body">
      <Transition name="estimator-drawer">
        <div
          v-if="drawerOpen"
          class="fixed inset-0 z-50 flex items-end bg-black/25 backdrop-blur-[1px] lg:items-stretch lg:justify-end"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="drawerTitleId"
          @click.self="closeDrawer"
        >
          <aside
            :id="drawerId"
            ref="drawerPanel"
            tabindex="-1"
            class="flex max-h-[88dvh] w-full min-h-0 flex-col overflow-hidden rounded-t-panel border border-b-0 border-outline bg-surface-raised shadow-floating outline-none lg:h-full lg:max-h-none lg:max-w-[480px] lg:rounded-none lg:border-b lg:border-r-0"
          >
            <header class="flex min-h-[72px] shrink-0 items-center gap-3 border-b border-outline px-4 sm:px-5">
              <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control bg-info-subtle text-info-foreground">
                <Icon name="calculator" size="md" aria-hidden="true" />
              </span>
              <div class="min-w-0 flex-1">
                <h2 :id="drawerTitleId" class="text-base font-semibold text-foreground">{{ t('payment.estimator.title') }}</h2>
                <p class="mt-0.5 text-xs leading-5 text-foreground-subtle">{{ t('payment.estimator.summary') }}</p>
              </div>
              <button
                ref="closeButton"
                type="button"
                class="btn btn-ghost btn-icon shrink-0"
                :aria-label="t('common.close')"
                @click="closeDrawer"
              >
                <Icon name="x" size="md" aria-hidden="true" />
              </button>
            </header>

            <div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-5">
              <div v-if="loading" class="flex min-h-48 items-center justify-center" aria-live="polite">
                <LoadingSpinner />
              </div>
              <div v-else-if="loadFailed" class="flex items-center justify-between gap-3 rounded-control bg-warning-subtle px-3 py-2.5 text-sm text-warning-foreground" role="status">
                <span>{{ t('payment.estimator.loadFailed') }}</span>
                <button type="button" class="btn btn-ghost btn-sm shrink-0" @click="loadOptions">{{ t('common.retry') }}</button>
              </div>
              <div v-else-if="models.length === 0" class="rounded-control bg-surface-subtle px-3 py-3 text-sm text-foreground-subtle">
                {{ t('payment.estimator.empty') }}
              </div>
              <template v-else>
                <div class="space-y-4">
                  <div class="rounded-panel border border-outline bg-surface-subtle p-3.5 sm:p-4">
                    <label for="estimator-recharge-amount" class="text-sm font-medium text-foreground">
                      {{ t('payment.estimator.amount') }}
                    </label>
                    <div class="relative mt-2">
                      <span class="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm font-semibold text-foreground-subtle">¥</span>
                      <input
                        id="estimator-recharge-amount"
                        data-testid="estimator-amount-input"
                        type="number"
                        min="0"
                        step="0.01"
                        inputmode="decimal"
                        class="input h-11 w-full pl-8 text-base font-semibold tabular-nums"
                        :value="estimateAmount || ''"
                        :placeholder="t('payment.estimator.amountPlaceholder')"
                        @input="updateEstimateAmount"
                      />
                    </div>
                    <div class="mt-2.5 grid grid-cols-3 gap-2 sm:grid-cols-5" :aria-label="t('payment.estimator.quickAmounts')">
                      <button
                        v-for="quickAmount in estimatorQuickAmounts"
                        :key="quickAmount"
                        type="button"
                        class="min-h-9 rounded-control border px-2 text-xs font-semibold tabular-nums transition-colors"
                        :class="estimateAmount === quickAmount
                          ? 'border-info/35 bg-info-subtle text-info-foreground'
                          : 'border-outline bg-surface text-foreground-muted hover:border-outline-strong hover:text-foreground'"
                        :aria-pressed="estimateAmount === quickAmount"
                        @click="estimateAmount = quickAmount"
                      >
                        ¥{{ formatPlainAmount(quickAmount) }}
                      </button>
                    </div>
                  </div>

                  <div class="block min-w-0 text-sm font-medium text-foreground">
                    <span id="estimator-model-label" class="mb-2 block">{{ t('payment.estimator.model') }}</span>
                    <Select
                      data-testid="estimator-model-select"
                      :model-value="selectedModelKey"
                      :options="modelOptions"
                      searchable
                      :search-placeholder="t('payment.estimator.searchModel')"
                      aria-labelledby="estimator-model-label"
                      @update:model-value="selectedModelKey = String($event ?? '')"
                    />
                  </div>
                  <div class="block min-w-0 text-sm font-medium text-foreground">
                    <span id="estimator-group-label" class="mb-2 block">{{ t('payment.estimator.group') }}</span>
                    <Select
                      data-testid="estimator-group-select"
                      :model-value="selectedGroupId"
                      :options="groupOptions"
                      searchable
                      :search-placeholder="t('payment.estimator.searchGroup')"
                      aria-labelledby="estimator-group-label"
                      @update:model-value="selectedGroupId = $event == null ? null : Number($event)"
                    />
                  </div>
                </div>

                <section class="mt-5 overflow-hidden rounded-panel border border-outline bg-surface" :aria-label="t('payment.estimator.valueComparison')">
                  <header class="flex items-center justify-between gap-3 border-b border-outline px-4 py-3">
                    <h3 class="text-sm font-semibold text-foreground">{{ t('payment.estimator.valueComparison') }}</h3>
                    <span class="rounded-full bg-success-subtle px-2.5 py-1 text-xs font-semibold text-success-foreground">
                      {{ t('payment.estimator.estimatedSavings', { percent: formatPercent(estimatedSavingsPercent) }) }}
                    </span>
                  </header>

                  <dl class="grid grid-cols-2 divide-x divide-outline border-b border-outline">
                    <div class="min-w-0 px-4 py-3.5">
                      <dt class="text-xs text-foreground-subtle">{{ t('payment.estimator.platformBalance') }}</dt>
                      <dd data-testid="estimated-platform-balance" class="mt-1 text-lg font-semibold tabular-nums text-foreground">{{ formatUSD(estimatedPlatformBalance) }}</dd>
                      <p class="mt-1 text-[11px] leading-4 text-foreground-subtle">
                        {{ t('payment.estimator.platformExchangeHint', { usd: formatRate(safeBalanceMultiplier) }) }}
                      </p>
                    </div>
                    <div class="min-w-0 px-4 py-3.5">
                      <dt class="text-xs text-foreground-subtle">{{ t('payment.estimator.officialExchange') }}</dt>
                      <dd data-testid="official-exchange-usd" class="mt-1 text-lg font-semibold tabular-nums text-foreground">≈ {{ formatUSD(officialExchangeUsd) }}</dd>
                      <p class="mt-1 text-[11px] leading-4 text-foreground-subtle">
                        {{ t('payment.estimator.officialExchangeHint', { cny: formatRate(safeOfficialUsdToCnyRate) }) }}
                      </p>
                    </div>
                  </dl>

                  <div class="bg-info-subtle px-4 py-4">
                    <div class="flex items-start justify-between gap-4">
                      <div class="min-w-0">
                        <p class="text-xs font-medium text-info-foreground">{{ t('payment.estimator.officialEquivalent') }}</p>
                        <p data-testid="official-equivalent" class="mt-1 text-2xl font-semibold tabular-nums text-info-foreground">≈ {{ formatUSD(officialEquivalent) }}</p>
                      </div>
                      <span class="shrink-0 rounded-full bg-surface/80 px-2.5 py-1 text-xs font-semibold text-info-foreground">
                        {{ t('payment.estimator.purchasePower', { multiple: formatMultiple(purchasePowerMultiple) }) }}
                      </span>
                    </div>
                    <dl class="mt-3 grid gap-2 border-t border-info/15 pt-3 text-sm">
                      <div class="flex items-center justify-between gap-4">
                        <dt class="text-info-foreground/75">{{ t('payment.estimator.platformCost') }}</dt>
                        <dd class="font-medium tabular-nums text-info-foreground">{{ formatCNY(estimatedPayAmount) }}</dd>
                      </div>
                      <div class="flex items-center justify-between gap-4">
                        <dt class="text-info-foreground/75">{{ t('payment.estimator.officialCost') }}</dt>
                        <dd data-testid="official-cny-cost" class="font-semibold tabular-nums text-info-foreground">≈ {{ formatCNY(officialEquivalentCostCny) }}</dd>
                      </div>
                    </dl>
                  </div>
                </section>
                <p class="mt-3 text-xs leading-5 text-foreground-subtle">
                  {{ t('payment.estimator.disclaimer', { rate: formatRate(selectedGroupRate), cny: formatRate(safeOfficialUsdToCnyRate) }) }}
                </p>
              </template>
            </div>
          </aside>
        </div>
      </Transition>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import userChannelsAPI from '@/api/channels'
import userGroupsAPI from '@/api/groups'
import {
  buildMarketplaceEntries,
  compareMarketplaceDisplayOrder,
  compareMarketplaceModelRecency,
  compareMarketplaceProviders,
  DEFAULT_USD_TO_CNY_RATE,
  sortedEntryGroups,
  type MarketplaceModelEntry,
} from '@/views/user/modelMarketplace'
import type { QuickRechargeAmount, RechargeBonusTier } from '@/types/payment'
import { quoteRechargeBonus, type RechargeBonusMode } from '@/utils/rechargeBonus'
import { formatPaymentAmount } from './currency'

const props = withDefaults(defineProps<{
  creditedAmount: number
  rechargeAmount?: number
  balanceRechargeMultiplier?: number
  officialUsdToCnyRate?: number
  quickRechargeAmounts?: QuickRechargeAmount[]
  bonusTiers?: RechargeBonusTier[]
  bonusMode?: RechargeBonusMode
}>(), {
  rechargeAmount: 0,
  balanceRechargeMultiplier: 1,
  officialUsdToCnyRate: 0,
  quickRechargeAmounts: () => [],
  bonusTiers: () => [],
  bonusMode: 'bonus',
})

const { t } = useI18n()
const titleId = 'recharge-estimator-title'
const drawerId = 'recharge-estimator-drawer'
const drawerTitleId = 'recharge-estimator-drawer-title'
const drawerOpen = ref(false)
const triggerButton = ref<HTMLButtonElement | null>(null)
const closeButton = ref<HTMLButtonElement | null>(null)
const drawerPanel = ref<HTMLElement | null>(null)
const loading = ref(false)
const loadFailed = ref(false)
const loaded = ref(false)
const models = ref<MarketplaceModelEntry[]>([])
const userGroupRates = ref<Record<number, number>>({})
const selectedModelKey = ref('')
const selectedGroupId = ref<number | null>(null)
const estimateAmount = ref(0)
let loadController: AbortController | null = null
let previousBodyOverflow = ''

const selectedModel = computed(() => models.value.find(model => model.key === selectedModelKey.value) ?? null)
const selectedModelGroups = computed(() => selectedModel.value ? sortedEntryGroups(selectedModel.value, userGroupRates.value) : [])
const modelOptions = computed<SelectOption[]>(() => models.value.map(model => ({
  value: model.key,
  label: `${model.name} · ${model.platform.toUpperCase()}`,
  description: `${model.name} ${model.platform}`,
})))
const groupOptions = computed<SelectOption[]>(() => selectedModelGroups.value.map(group => ({
  value: group.id,
  label: `${group.name} · ${formatRate(group.effectiveRate)}x`,
  description: `${group.name} ${formatRate(group.effectiveRate)}`,
})))
const selectedGroupRate = computed(() => selectedModelGroups.value.find(group => group.id === selectedGroupId.value)?.effectiveRate ?? 1)
const safeBalanceMultiplier = computed(() => Number.isFinite(props.balanceRechargeMultiplier) && props.balanceRechargeMultiplier > 0
  ? props.balanceRechargeMultiplier
  : 1)
const safeOfficialUsdToCnyRate = computed(() => Number.isFinite(props.officialUsdToCnyRate) && props.officialUsdToCnyRate > 0
  ? props.officialUsdToCnyRate
  : DEFAULT_USD_TO_CNY_RATE)
const estimatorQuickAmounts = computed(() => {
  const configured = props.quickRechargeAmounts
    .map(item => item.amount)
    .filter(amount => Number.isFinite(amount) && amount > 0)
  const unique = [...new Set(configured.length > 0 ? configured : [10, 50, 100, 200, 500])]
  return unique.slice(0, 5)
})
// 与充值页同口径按优惠阶梯报价：赠金模式到账含赠额，折扣模式实付为折后基数。
const estimateQuote = computed(() => quoteRechargeBonus(props.bonusTiers, Math.max(0, estimateAmount.value), {
  multiplier: safeBalanceMultiplier.value,
  mode: props.bonusMode,
}))
const estimatedPlatformBalance = computed(() => estimateQuote.value.credited)
const estimatedPayAmount = computed(() => estimateQuote.value.payBase)
const officialEquivalent = computed(() => {
  const rate = selectedGroupRate.value
  if (!Number.isFinite(rate) || rate <= 0) return 0
  return Math.round((estimatedPlatformBalance.value / rate) * 100) / 100
})
const officialExchangeUsd = computed(() => Math.round((Math.max(0, estimatedPayAmount.value) / safeOfficialUsdToCnyRate.value) * 100) / 100)
const officialEquivalentCostCny = computed(() => Math.round(officialEquivalent.value * safeOfficialUsdToCnyRate.value * 100) / 100)
const purchasePowerMultiple = computed(() => estimatedPayAmount.value > 0 ? officialEquivalentCostCny.value / estimatedPayAmount.value : 0)
const estimatedSavingsPercent = computed(() => officialEquivalentCostCny.value > 0
  ? Math.max(0, (1 - Math.max(0, estimatedPayAmount.value) / officialEquivalentCostCny.value) * 100)
  : 0)

watch(selectedModelGroups, (groups) => {
  if (groups.some(group => group.id === selectedGroupId.value)) return
  selectedGroupId.value = groups[0]?.id ?? null
}, { immediate: true })

function formatUSD(value: number): string {
  return formatPaymentAmount(value, 'USD')
}

function formatCNY(value: number): string {
  return formatPaymentAmount(value, 'CNY')
}

function formatPlainAmount(value: number): string {
  return Number(value.toFixed(2)).toLocaleString()
}

function formatRate(value: number): string {
  return Number(value.toFixed(4)).toString()
}

function formatPercent(value: number): string {
  return Number(value.toFixed(1)).toString()
}

function formatMultiple(value: number): string {
  return Number(value.toFixed(value >= 10 ? 1 : 2)).toString()
}

function defaultEstimateAmount(): number {
  if (Number.isFinite(props.rechargeAmount) && props.rechargeAmount > 0) return props.rechargeAmount
  if (Number.isFinite(props.creditedAmount) && props.creditedAmount > 0) {
    return Math.round((props.creditedAmount / safeBalanceMultiplier.value) * 100) / 100
  }
  return estimatorQuickAmounts.value[0] ?? 10
}

function updateEstimateAmount(event: Event): void {
  const value = Number((event.target as HTMLInputElement).value)
  estimateAmount.value = Number.isFinite(value) ? Math.max(0, value) : 0
}

function compareEstimatorModels(a: MarketplaceModelEntry, b: MarketplaceModelEntry): number {
  const byProvider = compareMarketplaceProviders(a.platform, b.platform)
  if (byProvider !== 0) return byProvider
  const byRecency = compareMarketplaceModelRecency(a.name, b.name)
  if (byRecency !== 0) return byRecency
  const byDisplayOrder = compareMarketplaceDisplayOrder(a, b)
  if (byDisplayOrder !== 0) return byDisplayOrder
  return a.name.localeCompare(b.name)
}

function selectDefaultEstimate(): void {
  const candidates = models.value.flatMap((model, modelIndex) =>
    sortedEntryGroups(model, userGroupRates.value).map((group, groupIndex) => ({ model, modelIndex, group, groupIndex })),
  )
  candidates.sort((a, b) =>
    a.group.effectiveRate - b.group.effectiveRate
    || a.modelIndex - b.modelIndex
    || a.groupIndex - b.groupIndex,
  )
  selectedModelKey.value = candidates[0]?.model.key ?? ''
  selectedGroupId.value = candidates[0]?.group.id ?? null
}

async function loadOptions(): Promise<void> {
  loadController?.abort()
  loadController = new AbortController()
  loading.value = true
  loadFailed.value = false
  try {
    const [catalog, rates] = await Promise.all([
      userChannelsAPI.getMarketplace({ signal: loadController.signal }),
      userGroupsAPI.getUserGroupRates(),
    ])
    userGroupRates.value = rates
    models.value = buildMarketplaceEntries(catalog)
      .filter(model => model.groups.length > 0)
      .sort(compareEstimatorModels)
    selectDefaultEstimate()
    loaded.value = true
  } catch (error) {
    if (loadController.signal.aborted) return
    console.error('Failed to load recharge estimator options:', error)
    loadFailed.value = true
  } finally {
    if (!loadController.signal.aborted) loading.value = false
  }
}

function closeDrawer(): void {
  drawerOpen.value = false
}

function handleKeydown(event: KeyboardEvent): void {
  if (!drawerOpen.value || event.defaultPrevented) return
  if (event.key === 'Escape') {
    event.preventDefault()
    closeDrawer()
    return
  }
  if (event.key !== 'Tab' || !drawerPanel.value) return

  const focusable = [...drawerPanel.value.querySelectorAll<HTMLElement>(
    'button:not([disabled]),input:not([disabled]),select:not([disabled]),[tabindex]:not([tabindex="-1"])',
  )].filter(element => getComputedStyle(element).visibility !== 'hidden')
  if (!focusable.length) {
    event.preventDefault()
    drawerPanel.value.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(drawerOpen, async (open) => {
  if (open) {
    estimateAmount.value = defaultEstimateAmount()
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleKeydown)
    if (!loaded.value && !loading.value) void loadOptions()
    await nextTick()
    closeButton.value?.focus()
    return
  }

  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
  triggerButton.value?.focus()
})

onBeforeUnmount(() => {
  loadController?.abort()
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
})
</script>

<style scoped>
.estimator-drawer-enter-active,
.estimator-drawer-leave-active {
  transition: opacity 180ms ease;
}

.estimator-drawer-enter-active aside,
.estimator-drawer-leave-active aside {
  transition: transform 180ms ease;
}

.estimator-drawer-enter-from,
.estimator-drawer-leave-to {
  opacity: 0;
}

.estimator-drawer-enter-from aside,
.estimator-drawer-leave-to aside {
  transform: translateY(100%);
}

@media (min-width: 1024px) {
  .estimator-drawer-enter-from aside,
  .estimator-drawer-leave-to aside {
    transform: translateX(100%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .estimator-drawer-enter-active,
  .estimator-drawer-leave-active,
  .estimator-drawer-enter-active aside,
  .estimator-drawer-leave-active aside {
    transition: none;
  }
}
</style>
