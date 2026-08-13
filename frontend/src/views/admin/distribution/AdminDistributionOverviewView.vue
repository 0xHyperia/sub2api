<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
        <div class="min-w-0"><h1 class="page-title">{{ t('admin.distribution.analytics.overviewTitle') }}</h1><p class="page-description">{{ t('admin.distribution.analytics.overviewDescription') }}</p></div>
        <div class="flex min-w-0 flex-1 flex-wrap items-center justify-end gap-2 xl:flex-none">
          <DistributionAnalyticsRange v-model="range" @change="load" />
          <button class="btn btn-secondary btn-icon shrink-0" :title="t('common.refresh')" :aria-label="t('common.refresh')" :disabled="loading" @click="load"><Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" /></button>
        </div>
      </header>
      <AdminDistributionNav />

      <div v-if="loading && !overview" class="flex min-h-56 items-center justify-center"><LoadingSpinner /></div>
      <template v-else-if="overview && analytics">
        <section class="kpi-strip" :aria-label="t('common.distributionAnalytics.range')">
          <RouterLink v-for="item in kpis" :key="item.label" :to="item.to" class="kpi-item group">
            <div class="flex items-center justify-between gap-2"><p>{{ item.label }}</p><Icon :name="item.icon" size="sm" /></div>
            <strong>{{ item.value }}</strong>
            <span :class="growthClass(item.growth)">{{ growthText(item.growth) }}</span>
          </RouterLink>
        </section>

        <DistributionBusinessChart :direct="analytics.daily_direct" :team="analytics.daily_team" :resolution="analytics.trend_resolution" :date-from="analytics.date_from" :date-to="analytics.date_to" :title="t('admin.distribution.analytics.channelTrend')" :subtitle="rangeLabel" />

        <section class="grid gap-4 xl:grid-cols-[minmax(0,1.25fr)_minmax(360px,.75fr)]">
          <div class="panel">
            <div class="panel-heading"><div><h2>{{ t('admin.distribution.analytics.channelComposition') }}</h2><p>{{ t('admin.distribution.analytics.channelCompositionHint') }}</p></div></div>
            <div class="hidden overflow-x-auto sm:block">
              <table class="business-table">
                <thead><tr><th>{{ t('common.distributionAnalytics.businessScope') }}</th><th>{{ t('common.distributionAnalytics.newCustomers') }}</th><th>{{ t('common.distributionAnalytics.payingCustomers') }}</th><th>{{ t('common.distributionAnalytics.conversionRate') }}</th><th>{{ t('common.distributionAnalytics.customerPaid') }}</th><th>{{ t('common.distributionAnalytics.commission') }}</th><th>{{ t('common.distributionAnalytics.commissionCostRate') }}</th></tr></thead>
                <tbody><tr v-for="row in businessRows" :key="row.label"><td><span :class="row.badge">{{ row.label }}</span></td><td>{{ row.data.new_customers }}</td><td>{{ row.data.paying_customers }}</td><td>{{ percent(row.data.conversion_rate) }}</td><td>{{ money(row.data.customer_paid_cny) }}</td><td>{{ money(row.data.commission_cny) }}</td><td>{{ commissionRate(row.data) }}</td></tr></tbody>
              </table>
            </div>
            <div class="business-summaries sm:hidden">
              <article v-for="row in businessRows" :key="row.label">
                <div class="flex items-center justify-between gap-3"><span :class="row.badge">{{ row.label }}</span><strong>{{ percent(row.data.conversion_rate) }}</strong></div>
                <dl><div><dt>{{ t('common.distributionAnalytics.newCustomers') }} / {{ t('common.distributionAnalytics.payingCustomers') }}</dt><dd>{{ row.data.new_customers }} / {{ row.data.paying_customers }}</dd></div><div><dt>{{ t('common.distributionAnalytics.customerPaid') }}</dt><dd>{{ money(row.data.customer_paid_cny) }}</dd></div><div><dt>{{ t('common.distributionAnalytics.commission') }}</dt><dd>{{ money(row.data.commission_cny) }}</dd></div><div><dt>{{ t('common.distributionAnalytics.commissionCostRate') }}</dt><dd>{{ commissionRate(row.data) }}</dd></div></dl>
              </article>
            </div>
          </div>

          <div class="panel">
            <div class="panel-heading"><div><h2>{{ t('admin.distribution.analytics.topAgents') }}</h2><p>{{ t('admin.distribution.analytics.topAgentsHint') }}</p></div><RouterLink to="/admin/distribution/agents" class="text-xs font-medium text-brand">{{ t('admin.distribution.analytics.allAgents') }}</RouterLink></div>
            <div v-if="overview.agent_ranking.length" class="divide-y divide-outline">
              <RouterLink v-for="(agent, index) in overview.agent_ranking" :key="agent.agent_id" :to="`/admin/distribution/agents?agent_id=${agent.agent_id}`" class="ranking-row">
                <span class="rank">{{ index + 1 }}</span><span class="min-w-0 flex-1"><strong class="block truncate text-sm">{{ agent.username || agent.email }}</strong><small class="block truncate">{{ agent.email }}</small></span><span class="text-right"><strong class="block text-sm tabular-nums">{{ money(agent.customer_paid_cny) }}</strong><small class="block">{{ t('admin.distribution.analytics.rankingCommission', { amount: money(agent.commission_cny) }) }}</small></span>
              </RouterLink>
            </div><div v-else class="empty">{{ t('admin.distribution.analytics.noAgentSales') }}</div>
          </div>
        </section>

        <section class="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,.8fr)]">
          <div class="panel">
            <div class="panel-heading"><div><h2>{{ t('admin.distribution.analytics.liability') }}</h2><p>{{ t('admin.distribution.analytics.liabilityHint') }}</p></div><RouterLink to="/admin/distribution/commissions" class="text-xs font-medium text-brand">{{ t('admin.distribution.analytics.ledger') }}</RouterLink></div>
            <dl class="detail-grid"><div><dt>{{ t('admin.distribution.analytics.availableCommission') }}</dt><dd class="text-success-foreground">{{ money(overview.available_commission_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.frozenCommission') }}</dt><dd>{{ money(overview.frozen_commission_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.reservedCommission') }}</dt><dd>{{ money(overview.reserved_commission_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.registrationRewards') }}</dt><dd>{{ money(overview.period_registration_reward_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.rechargeRewards') }}</dt><dd>{{ money(overview.period_recharge_reward_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.paidThisMonth') }}</dt><dd>{{ money(overview.paid_this_month_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.reversals') }}</dt><dd :class="Number(overview.reversed_commission_cny) ? 'text-danger-foreground' : ''">{{ money(overview.reversed_commission_cny) }}</dd></div><div><dt>{{ t('admin.distribution.analytics.agentDebt') }}</dt><dd :class="Number(overview.debt_commission_cny) ? 'text-danger-foreground' : ''">{{ money(overview.debt_commission_cny) }}</dd></div></dl>
          </div>
          <div class="panel">
            <div class="panel-heading"><div><h2>{{ t('admin.distribution.analytics.operations') }}</h2><p>{{ t('admin.distribution.analytics.operationsHint') }}</p></div><RouterLink to="/admin/distribution/anomalies" class="text-xs font-medium text-brand">{{ t('admin.distribution.analytics.reconciliation') }}</RouterLink></div>
            <div class="divide-y divide-outline"><RouterLink v-for="todo in todos" :key="todo.label" :to="todo.to" class="todo-row"><span><strong>{{ todo.label }}</strong><small>{{ todo.hint }}</small></span><span class="badge" :class="todo.count ? todo.badge : 'badge-gray'">{{ todo.count }}</span></RouterLink></div>
            <div v-if="maturity" class="mt-3 flex items-center justify-between gap-3 border-t border-outline pt-3 text-xs"><span class="min-w-0 truncate text-foreground-subtle">{{ t('admin.distribution.analytics.maturity') }}：{{ maturityDescription }}</span><span class="badge shrink-0" :class="maturityStatusClass">{{ maturityStatusText }}</span></div>
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
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import AdminDistributionNav from '@/components/admin/distribution/AdminDistributionNav.vue'
import DistributionBusinessChart from '@/components/distribution/DistributionBusinessChart.vue'
import DistributionAnalyticsRange from '@/components/distribution/DistributionAnalyticsRange.vue'
import { analyticsRangeParams, defaultDistributionAnalyticsRange, formatAnalyticsRangeLabel } from '@/components/distribution/distributionAnalyticsRange'
import Icon from '@/components/icons/Icon.vue'
import { getMaturityStatus, getOverview, type DistributionAdminOverview, type DistributionMaturityStatus } from '@/api/admin/distribution'
import type { DistributionBusinessMetrics } from '@/api/distribution'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const app = useAppStore()
const { t } = useI18n()
const range = ref(defaultDistributionAnalyticsRange())
const overview = ref<DistributionAdminOverview | null>(null)
const maturity = ref<DistributionMaturityStatus | null>(null)
const loading = ref(false)
const analytics = computed(() => overview.value?.analytics)
const rangeLabel = computed(() => formatAnalyticsRangeLabel(range.value))
const money = (value: string | number) => new Intl.NumberFormat(undefined, { style: 'currency', currency: 'CNY', minimumFractionDigits: 2 }).format(Number(value || 0))
const percent = (value: string | number) => `${Number(value || 0).toFixed(1)}%`
const growthText = (value?: string) => value == null ? t('common.distributionAnalytics.previousUnavailable') : t(Number(value) >= 0 ? 'common.distributionAnalytics.comparedUp' : 'common.distributionAnalytics.comparedDown', { value: Number(value).toFixed(1) })
const growthClass = (value?: string) => value == null ? 'growth-neutral' : Number(value) > 0 ? 'growth-positive' : Number(value) < 0 ? 'growth-negative' : 'growth-neutral'
const commissionRate = (item: DistributionBusinessMetrics) => Number(item.customer_paid_cny) > 0 ? percent(Number(item.commission_cny) / Number(item.customer_paid_cny) * 100) : '—'
const kpis = computed(() => analytics.value ? [
  { label: t('common.distributionAnalytics.currentPaid'), value: money(analytics.value.total.current.customer_paid_cny), growth: analytics.value.total.paid_growth_rate, icon: 'creditCard' as const, to: '/admin/distribution/customers' },
  { label: t('common.distributionAnalytics.currentCommission'), value: money(analytics.value.total.current.commission_cny), growth: analytics.value.total.commission_growth_rate, icon: 'gift' as const, to: '/admin/distribution/commissions' },
  { label: t('common.distributionAnalytics.newCustomers'), value: String(analytics.value.total.current.new_customers), growth: undefined, icon: 'userPlus' as const, to: '/admin/distribution/customers' },
  { label: t('common.distributionAnalytics.conversionRate'), value: percent(analytics.value.total.current.conversion_rate), growth: undefined, icon: 'trendingUp' as const, to: '/admin/distribution/customers' },
  { label: t('common.distributionAnalytics.activeAgents'), value: String(analytics.value.active_agents), growth: undefined, icon: 'users' as const, to: '/admin/distribution/agents' },
] : [])
const businessRows = computed(() => analytics.value ? [
  { label: t('common.distributionAnalytics.directBusiness'), badge: 'badge badge-primary', data: analytics.value.direct.current },
  { label: t('common.distributionAnalytics.teamBusiness'), badge: 'badge badge-gray', data: analytics.value.team.current },
  { label: t('common.distributionAnalytics.businessTotal'), badge: 'badge badge-success', data: analytics.value.total.current },
] : [])
const todos = computed(() => overview.value ? [
  { label: t('admin.distribution.analytics.pendingWithdrawals'), hint: t('admin.distribution.analytics.withdrawalTotal', { amount: money(overview.value.pending_withdrawal_cny) }), count: overview.value.pending_withdrawals, badge: 'badge-warning', to: '/admin/distribution/withdrawals?status=pending' },
  { label: t('admin.distribution.analytics.payingWithdrawals'), hint: t('admin.distribution.analytics.withdrawalTotal', { amount: money(overview.value.paying_withdrawal_cny) }), count: overview.value.paying_withdrawals, badge: 'badge-primary', to: '/admin/distribution/withdrawals?status=paying' },
  { label: t('admin.distribution.analytics.overdueReview'), hint: t('admin.distribution.analytics.overdueHint'), count: overview.value.overdue_pending_count, badge: 'badge-danger', to: '/admin/distribution/anomalies?type=overdue_withdrawal' },
] : [])
const maturityStatusText = computed(() => {
  if (maturity.value?.running) return t('admin.distribution.analytics.maturityRunning')
  const labels: Record<string, string> = { success: t('admin.distribution.analytics.maturityHealthy'), standby: t('admin.distribution.analytics.maturityStandby'), error: t('admin.distribution.analytics.maturityError'), disabled: t('admin.distribution.analytics.maturityDisabled') }
  return labels[maturity.value?.last_outcome || ''] || t('admin.distribution.analytics.maturityPending')
})
const maturityStatusClass = computed(() => maturity.value?.last_outcome === 'error' ? 'badge-danger' : maturity.value?.running ? 'badge-primary' : maturity.value?.last_outcome === 'success' ? 'badge-success' : 'badge-gray')
const maturityDescription = computed(() => maturity.value?.last_error || (maturity.value?.last_success_at ? t('admin.distribution.analytics.maturityReleased', { time: new Date(maturity.value.last_success_at).toLocaleString(), count: maturity.value.last_released }) : t('admin.distribution.analytics.maturityWaiting')))
let requestSequence = 0
async function load() {
  const sequence = ++requestSequence; loading.value = true
  try { const [nextOverview, nextMaturity] = await Promise.all([getOverview(analyticsRangeParams(range.value)), getMaturityStatus()]); if (sequence === requestSequence) { overview.value = nextOverview; maturity.value = nextMaturity } }
  catch (error) { if (sequence === requestSequence) app.showError(extractApiErrorMessage(error, t('admin.distribution.analytics.loadFailed'))) }
  finally { if (sequence === requestSequence) loading.value = false }
}
onMounted(load)
</script>

<style scoped>
.kpi-strip { display: grid; grid-template-columns: repeat(5,minmax(0,1fr)); overflow: hidden; border: 1px solid var(--ui-border); border-radius: 8px; background: var(--ui-surface); }
.kpi-item { min-width: 0; padding: 16px; border-right: 1px solid var(--ui-border); transition: background 150ms ease; }.kpi-item:last-child { border-right: 0; }.kpi-item:hover { background: var(--ui-surface-subtle); }.kpi-item p,.kpi-item svg { color: var(--ui-text-muted); font-size: 12px; }.kpi-item strong { display: block; margin-top: 8px; overflow: hidden; color: var(--ui-text); font-size: 20px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; font-variant-numeric: tabular-nums; }.kpi-item span { display: block; margin-top: 4px; font-size: 11px; }
.growth-positive { color: rgb(var(--color-success-foreground)); }.growth-negative { color: rgb(var(--color-danger-foreground)); }.growth-neutral { color: var(--ui-text-subtle); }
.panel { min-width: 0; border: 1px solid var(--ui-border); border-radius: 8px; background: var(--ui-surface); padding: 16px; }.panel-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }.panel-heading h2 { font-size: 14px; font-weight: 650; }.panel-heading p { margin-top: 3px; color: var(--ui-text-muted); font-size: 12px; }
.business-table { width: 100%; min-width: 680px; border-collapse: collapse; font-size: 12px; }.business-table th { padding: 10px; border-bottom: 1px solid var(--ui-border); color: var(--ui-text-muted); text-align: right; font-weight: 600; }.business-table th:first-child,.business-table td:first-child { text-align: left; }.business-table td { padding: 12px 10px; border-bottom: 1px solid var(--ui-border); text-align: right; font-variant-numeric: tabular-nums; }.business-table tbody tr:last-child td { border-bottom: 0; }
.business-summaries article { padding: 13px 0; border-bottom: 1px solid var(--ui-border); }.business-summaries article:first-child { padding-top: 2px; }.business-summaries article:last-child { padding-bottom: 0; border-bottom: 0; }.business-summaries strong { font-size: 14px; font-variant-numeric: tabular-nums; }.business-summaries dl { display: grid; grid-template-columns: repeat(2,minmax(0,1fr)); gap: 12px; margin-top: 12px; }.business-summaries dt { color: var(--ui-text-subtle); font-size: 11px; }.business-summaries dd { margin-top: 3px; overflow-wrap: anywhere; font-size: 13px; font-weight: 600; font-variant-numeric: tabular-nums; }
.ranking-row { display: flex; min-height: 58px; align-items: center; gap: 10px; }.ranking-row:hover strong { color: var(--ui-focus); }.ranking-row small { margin-top: 2px; color: var(--ui-text-subtle); font-size: 11px; }.rank { display: flex; width: 24px; height: 24px; align-items: center; justify-content: center; border-radius: 6px; color: var(--ui-text-muted); background: var(--ui-surface-subtle); font-size: 11px; font-weight: 700; }.empty { display: flex; min-height: 160px; align-items: center; justify-content: center; color: var(--ui-text-muted); font-size: 13px; }
.detail-grid { display: grid; grid-template-columns: repeat(3,minmax(0,1fr)); gap: 18px 24px; }.detail-grid dt { color: var(--ui-text-muted); font-size: 12px; }.detail-grid dd { margin-top: 5px; font-size: 15px; font-weight: 650; font-variant-numeric: tabular-nums; }.todo-row { display: flex; min-height: 58px; align-items: center; justify-content: space-between; gap: 12px; }.todo-row strong,.todo-row small { display: block; font-size: 12px; }.todo-row small { margin-top: 3px; color: var(--ui-text-subtle); }
@media (max-width: 1023px) { .kpi-strip { grid-template-columns: repeat(2,minmax(0,1fr)); }.kpi-item { border-bottom: 1px solid var(--ui-border); }.kpi-item:nth-child(2n) { border-right: 0; }.kpi-item:last-child { grid-column: 1/-1; border-bottom: 0; }.detail-grid { grid-template-columns: repeat(2,minmax(0,1fr)); } }
@media (max-width: 639px) { .kpi-item { padding: 14px; }.kpi-item strong { font-size: 17px; }.panel { padding: 14px; }.detail-grid { gap: 14px; } }
</style>
