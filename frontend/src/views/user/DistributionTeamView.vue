<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">{{ t('common.distributionCenter.teamTitle') }}</h1>
              <p class="page-description">{{ t('common.distributionCenter.teamDescription') }}</p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button
                v-if="overview?.agent.can_recruit_subagents"
                class="btn btn-primary h-10 px-3 sm:px-4"
                :disabled="overview?.agent.status !== 'active'"
                :aria-label="t('common.distributionCenter.addChild')"
                @click="openAddDialog"
              >
                <Icon name="userPlus" size="sm" /><span class="hidden sm:inline"
                  >{{ t('common.distributionCenter.addChild') }}</span
                >
              </button>
              <span v-else class="badge badge-gray whitespace-nowrap"
                >{{ t('common.distributionCenter.recruitDisabled') }}</span
              >
              <button
                class="btn btn-secondary btn-icon"
                :title="t('common.refresh')"
                :aria-label="t('common.refresh')"
                :disabled="loading"
                @click="load"
              >
                <Icon
                  name="refresh"
                  size="sm"
                  :class="loading ? 'animate-spin' : ''"
                />
              </button>
            </div>
          </header>
          <DistributionNav class="mt-5" />
        </template>

        <template #summary>
          <div v-if="overview?.analytics" class="space-y-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div><h2 class="text-sm font-semibold">{{ t('common.distributionCenter.teamBusiness') }}</h2><p class="mt-1 text-xs text-foreground-subtle">{{ t('common.distributionCenter.teamBusinessHint') }}</p></div>
              <DistributionAnalyticsRange v-model="range" @change="onRangeChange" />
            </div>
            <section class="team-kpis">
              <article><p>{{ t('common.distributionCenter.childAgents') }}</p><strong>{{ overview.team_count }}</strong><small>{{ t('common.distributionCenter.activeCount', { count: overview.analytics.active_agents }) }}</small></article>
              <article><p>{{ t('common.distributionAnalytics.newCustomers') }}</p><strong>{{ overview.analytics.team.current.new_customers }}</strong><small>{{ t('common.distributionCenter.payingCount', { count: overview.analytics.team.current.paying_customers }) }}</small></article>
              <article><p>{{ t('common.distributionAnalytics.conversionRate') }}</p><strong>{{ percent(overview.analytics.team.current.conversion_rate) }}</strong><small>{{ t('common.distributionAnalytics.cohortHint') }}</small></article>
              <article><p>{{ t('common.distributionCenter.teamCustomerPaid') }}</p><strong>{{ money(overview.analytics.team.current.customer_paid_cny) }}</strong><small>{{ growthText(overview.analytics.team.paid_growth_rate) }}</small></article>
              <article><p>{{ t('common.distributionCenter.myTeamCommission') }}</p><strong>{{ money(overview.analytics.team.current.commission_cny) }}</strong><small>{{ growthText(overview.analytics.team.commission_growth_rate) }}</small></article>
            </section>
          </div>
        </template>

        <template #filters>
          <div
            class="grid grid-cols-[minmax(0,1fr)_104px] gap-2 sm:max-w-2xl sm:grid-cols-[minmax(280px,1fr)_140px]"
          >
            <div class="relative min-w-0">
              <Icon
                name="search"
                size="sm"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle"
              /><input
                v-model="search"
                class="input pl-9"
                type="search"
                :placeholder="t('common.distributionCenter.searchChildPlaceholder')"
                :aria-label="t('common.distributionCenter.searchChild')"
              />
            </div>
            <select v-model="status" class="input" :aria-label="t('common.distributionCenter.childStatus')">
              <option value="">{{ t('common.distributionCenter.allStatuses') }}</option>
              <option value="active">{{ t('common.distributionCenter.statusActive') }}</option>
              <option value="suspended">{{ t('common.distributionCenter.statusSuspended') }}</option>
              <option value="revoked">{{ t('common.distributionCenter.statusRevoked') }}</option>
            </select>
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
            :actions-count="1"
            clickable-rows
            @row-click="openTeamDetail"
          >
            <template #cell-agent="{ row }"
              ><div class="min-w-44 max-w-56">
                <p class="truncate font-medium">
                  {{ row.email || t('common.distributionCenter.emailMissing') }}
                </p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                  {{ row.username || t('common.distributionCenter.usernameMissing') }}
                </p>
              </div></template
            >
            <template #cell-policy="{ row }"
              ><div class="min-w-32">
                <p class="font-semibold tabular-nums">
                  {{ rate(row.effective_rate_bps) }}
                </p>
                <p class="mt-0.5 font-mono text-xs text-foreground-subtle">
                  {{ row.promotion_code }}
                </p>
              </div></template
            >
            <template #cell-customers="{ row }"><div class="text-right"><p class="font-medium tabular-nums">{{ row.period_customer_count || 0 }} / {{ row.period_paying_customers || 0 }}</p><p class="mt-0.5 text-xs text-foreground-subtle">{{ t('common.distributionCenter.periodNewPaying') }}</p></div></template>
            <template #cell-business="{ row }"><div class="text-right"><p class="font-semibold tabular-nums">{{ money(row.period_customer_paid_cny) }}</p><p class="mt-0.5 text-xs text-foreground-subtle">{{ t('common.distributionAnalytics.currentPaid') }}</p></div></template>
            <template #cell-commission="{ row }"><div class="text-right"><p class="font-semibold tabular-nums text-success-foreground">{{ money(row.period_commission_cny) }}</p><p class="mt-0.5 text-xs text-foreground-subtle">{{ t('common.distributionCenter.periodDirectCommission') }}</p></div></template>
            <template #cell-status="{ row }"
              ><span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span></template
            >
            <template #cell-actions="{ row }"
              ><div class="flex items-center justify-end gap-1"><button class="btn btn-ghost btn-icon btn-sm" :title="t('common.distributionCenter.viewChild')" :aria-label="t('common.distributionCenter.viewChild')" @click.stop="openTeamDetail(row)"><Icon name="eye" size="sm" /></button><button class="btn btn-ghost btn-icon btn-sm" :title="t('common.distributionRewards.teamTitle')" :aria-label="t('common.distributionRewards.teamTitle')" @click.stop="openRewardRuleDialog(row)"><Icon name="gift" size="sm" /></button></div
            ></template>
            <template #mobile-card="{ row }"
              ><div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium">
                      {{ row.email || t('common.distributionCenter.emailMissing') }}
                    </p>
                    <p class="truncate text-xs text-foreground-subtle">
                      {{ row.username || t('common.distributionCenter.usernameMissing') }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="statusClass(row.status)"
                    >{{ statusText(row.status) }}</span
                  >
                </div>
                <div
                  class="flex items-center justify-between rounded-control bg-surface-subtle px-3 py-2"
                >
                  <div>
                    <p class="text-xs text-foreground-subtle">{{ t('common.distributionCenter.rate') }}</p>
                    <p class="mt-1 font-semibold">
                      {{ rate(row.effective_rate_bps) }}
                    </p>
                  </div>
                  <div class="min-w-0 text-right">
                    <p class="text-xs text-foreground-subtle">{{ t('common.distributionCenter.promotionCode') }}</p>
                    <p class="mt-1 max-w-40 truncate font-mono text-xs">
                      {{ row.promotion_code }}
                    </p>
                  </div>
                </div>
                <dl class="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <dt class="text-foreground-subtle">{{ t('common.distributionCenter.customerPaying') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ row.period_customer_count || 0 }} / {{ row.period_paying_customers || 0 }}
                    </dd>
                  </div>
                  <div class="text-right">
                    <dt class="text-foreground-subtle">{{ t('common.distributionAnalytics.customerPaid') }}</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(row.period_customer_paid_cny) }}
                    </dd>
                  </div>
                  <div><dt class="text-foreground-subtle">{{ t('common.distributionCenter.periodCommission') }}</dt><dd class="mt-1 font-semibold text-success-foreground">{{ money(row.period_commission_cny) }}</dd></div>
                </dl>
                <button class="btn btn-secondary w-full" @click="openTeamDetail(row)"><Icon name="eye" size="sm" />{{ t('common.distributionCenter.viewChild') }}</button>
              </div></template
            >
            <template #empty
              ><div class="py-12 text-center">
                <p class="font-medium">{{ t('common.distributionCenter.noChildren') }}</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  {{ t('common.distributionCenter.noChildrenHint') }}
                </p>
              </div></template
            >
          </DataTable>
        </template>
        <template #pagination
          ><Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="changePage"
            @update:pageSize="changePageSize"
        /></template>
      </TablePageLayout>
    </div>

    <BaseDialog
      :show="addDialog"
      :title="t('common.distributionCenter.addChild')"
      width="normal"
      @close="closeAddDialog"
    >
      <form
        id="add-team-agent-form"
        class="space-y-5"
        @submit.prevent="submitAgent"
      >
        <div>
          <label for="team-agent-email" class="input-label">{{ t('common.distributionCenter.userEmail') }}</label>
          <input
            id="team-agent-email"
            v-model.trim="childEmail"
            class="input"
            type="email"
            autocomplete="off"
            required
            :placeholder="t('common.distributionCenter.emailPlaceholder')"
          />
          <p class="input-hint">{{ t('common.distributionCenter.exactEmailHint') }}</p>
        </div>
        <div>
          <label for="team-agent-rate" class="input-label">{{ t('common.distributionCenter.rate') }}</label>
          <div class="relative">
            <input
              id="team-agent-rate"
              v-model="ratePercent"
              class="input pr-10"
              type="number"
              min="0"
              :max="maxRatePercent"
              step="0.01"
              :placeholder="t('common.distributionCenter.defaultRatePlaceholder')"
            /><span
              class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-foreground-muted"
              >%</span
            >
          </div>
          <p class="input-hint">
            {{ t('common.distributionCenter.maxRateHint', { rate: maxRatePercent.toFixed(2) }) }}
          </p>
        </div>
        <div class="border-t border-outline pt-4">
          <button
            type="button"
            class="flex items-center gap-2 text-sm font-medium text-foreground-muted hover:text-foreground"
            :aria-expanded="showAdvanced"
            @click="showAdvanced = !showAdvanced"
          >
            <Icon
              name="chevronDown"
              size="sm"
              :class="showAdvanced ? 'rotate-180' : ''"
            />{{ t('common.distributionCenter.advanced') }}
          </button>
          <div v-if="showAdvanced" class="mt-4">
            <label for="team-agent-code" class="input-label">{{ t('common.distributionCenter.customCode') }}</label
            ><input
              id="team-agent-code"
              v-model="promotionCode"
              class="input font-mono uppercase"
              maxlength="32"
              :placeholder="t('common.distributionCenter.autoCodePlaceholder')"
            />
          </div>
        </div>
      </form>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button
            class="btn btn-secondary"
            type="button"
            @click="closeAddDialog"
          >
            {{ t('common.distributionCenter.cancel') }}</button
          ><button
            class="btn btn-primary"
            type="submit"
            form="add-team-agent-form"
            :disabled="creating || !childEmail"
          >
            <Icon name="userPlus" size="sm" />{{
              creating ? t('common.distributionCenter.adding') : t('common.distributionCenter.confirmAdd')
            }}
          </button>
        </div></template
      >
    </BaseDialog>

    <BaseDrawer :show="teamDetailDialog" :title="t('common.distributionCenter.detailTitle')" :description="t('common.distributionCenter.detailDescription')" @close="closeTeamDetail">
      <div v-if="detailAgent" class="space-y-5">
        <div class="flex flex-wrap items-start justify-between gap-3 border-b border-outline pb-4">
          <div class="flex min-w-0 items-center gap-3"><div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-brand-subtle text-brand"><Icon name="user" size="md" /></div><div class="min-w-0"><p class="truncate font-semibold">{{ detailAgent.email || t('common.distributionCenter.emailMissing') }}</p><p class="mt-1 truncate text-sm text-foreground-subtle">{{ detailAgent.username || t('common.distributionCenter.usernameMissing') }}</p></div></div>
          <span class="badge" :class="statusClass(detailAgent.status)">{{ statusText(detailAgent.status) }}</span>
        </div>
        <DistributionAnalyticsRange v-model="detailRange" @change="loadTeamDetailAnalytics" />
        <div v-if="detailLoading" class="flex min-h-48 items-center justify-center"><LoadingSpinner /></div>
        <template v-else-if="detailAnalytics">
          <section>
            <h3 class="text-sm font-semibold">{{ t('common.distributionCenter.detailOverview') }}</h3>
            <dl class="detail-kpis mt-3 grid grid-cols-2 gap-px overflow-hidden rounded-panel border border-outline bg-outline sm:grid-cols-4">
              <div class="bg-surface p-3"><dt>{{ t('common.distributionAnalytics.currentPaid') }}</dt><dd>{{ money(detailAnalytics.analytics.direct.current.customer_paid_cny) }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionAnalytics.currentCommission') }}</dt><dd>{{ money(detailAnalytics.analytics.direct.current.commission_cny) }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionAnalytics.acquisitionTotal') }}</dt><dd>{{ detailAnalytics.analytics.direct.current.new_customers }} / {{ detailAnalytics.analytics.direct.current.paying_customers }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionAnalytics.conversionRate') }}</dt><dd>{{ percent(detailAnalytics.analytics.direct.current.conversion_rate) }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionCenter.periodOrders') }}</dt><dd>{{ detailAnalytics.analytics.direct.current.paid_orders }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionCenter.averageOrder') }}</dt><dd>{{ money(detailAnalytics.analytics.direct.current.average_order_cny) }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionCenter.directCustomers') }}</dt><dd>{{ detailAgent.customer_count || 0 }}</dd></div>
              <div class="bg-surface p-3"><dt>{{ t('common.distributionAnalytics.payingCustomers') }}</dt><dd>{{ detailAgent.paying_customer_count || 0 }}</dd></div>
            </dl>
          </section>
          <DistributionBusinessChart :direct="detailAnalytics.analytics.daily_direct" :resolution="detailAnalytics.analytics.trend_resolution" :date-from="detailAnalytics.analytics.date_from" :date-to="detailAnalytics.analytics.date_to" :show-scope="false" :title="t('common.distributionAnalytics.trend')" :subtitle="detailRangeLabel" />
          <section class="grid gap-4 sm:grid-cols-2">
            <div class="rounded-panel border border-outline p-4"><h3 class="text-sm font-semibold">{{ t('common.distributionCenter.detailCustomers') }}</h3><dl class="detail-list mt-3"><div><dt>{{ t('common.distributionCenter.cumulativePaid') }}</dt><dd>{{ money(detailAgent.customer_paid_cny) }}</dd></div><div><dt>{{ t('common.distributionCenter.joinedAt') }}</dt><dd>{{ detailAgent.created_at ? formatDate(detailAgent.created_at) : '-' }}</dd></div></dl></div>
            <div class="rounded-panel border border-outline p-4"><h3 class="text-sm font-semibold">{{ t('common.distributionCenter.detailCommissionPolicy') }}</h3><dl class="detail-list mt-3"><div><dt>{{ t('common.distributionCenter.rate') }}</dt><dd>{{ rate(detailAgent.effective_rate_bps) }}</dd></div><div><dt>{{ t('common.distributionCenter.promotionCode') }}</dt><dd class="font-mono">{{ detailAgent.promotion_code }}</dd></div><div><dt>{{ t('common.distributionCenter.cumulativeCommission') }}</dt><dd>{{ money(detailAgent.total_earned_cny) }}</dd></div><div><dt>{{ t('common.distributionCenter.availableCommission') }} / {{ t('common.distributionCenter.frozenCommission') }}</dt><dd>{{ money(detailAgent.available_cny) }} / {{ money(detailAgent.frozen_cny) }}</dd></div></dl></div>
          </section>
        </template>
        <div class="flex flex-wrap justify-end gap-2 border-t border-outline pt-4">
          <button class="btn btn-secondary" type="button" @click="openRewardRuleDialog(detailAgent)"><Icon name="gift" size="sm" />{{ t('common.distributionRewards.teamTitle') }}</button>
          <button v-if="detailAgent.status !== 'revoked'" class="btn" :class="detailAgent.status === 'active' ? 'btn-secondary' : 'btn-primary'" type="button" @click="openStatusConfirm(detailAgent)"><Icon :name="detailAgent.status === 'active' ? 'ban' : 'play'" size="sm" />{{ t(detailAgent.status === 'active' ? 'common.distributionCenter.suspendAgent' : 'common.distributionCenter.restoreAgent') }}</button>
        </div>
      </div>
    </BaseDrawer>

    <BaseDialog :show="rewardRuleDialog" :title="t('common.distributionRewards.teamTitle')" width="normal" @close="rewardRuleDialog = false">
      <form id="team-reward-rule-form" class="space-y-4" @submit.prevent="saveRewardRule">
        <p class="rounded-control bg-surface-subtle px-3 py-2 text-sm text-foreground-subtle">{{ t('common.distributionRewards.teamHint') }}</p>
        <div class="flex items-center justify-between"><span class="form-label">{{ t('common.distributionRewards.registration') }}</span><Toggle v-model="rewardRule.registration_enabled" :aria-label="t('common.distributionRewards.registration')" /></div>
        <label class="form-field"><span class="form-label">{{ t('common.distributionRewards.registrationShare') }}</span><div class="relative"><input v-model="rewardRule.registration_reward_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <div class="flex items-center justify-between"><span class="form-label">{{ t('common.distributionRewards.recharge') }}</span><Toggle v-model="rewardRule.recharge_enabled" :aria-label="t('common.distributionRewards.recharge')" /></div>
        <label class="form-field"><span class="form-label">{{ t('common.distributionRewards.rechargeShare') }}</span><div class="relative"><input v-model="rewardRule.recharge_reward_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
      </form>
      <template #footer><div class="flex justify-end gap-2"><button type="button" class="btn btn-secondary" @click="rewardRuleDialog = false">{{ t('common.distributionRewards.cancel') }}</button><button type="submit" form="team-reward-rule-form" class="btn btn-primary" :disabled="rewardRuleSaving">{{ rewardRuleSaving ? t('common.distributionRewards.saving') : t('common.distributionRewards.saveShare') }}</button></div></template>
    </BaseDialog>

    <ConfirmDialog
      :show="statusConfirm"
      :title="
        t(pendingAgent?.status === 'active' ? 'common.distributionCenter.suspendTitle' : 'common.distributionCenter.restoreTitle')
      "
      :message="statusConfirmMessage"
      :confirm-text="
        t(pendingAgent?.status === 'active' ? 'common.distributionCenter.confirmSuspend' : 'common.distributionCenter.confirmRestore')
      "
      :danger="pendingAgent?.status === 'active'"
      @confirm="confirmStatus"
      @cancel="statusConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import BaseDrawer from "@/components/common/BaseDrawer.vue";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import Toggle from "@/components/common/Toggle.vue";
import DistributionNav from "@/components/distribution/DistributionNav.vue";
import DistributionAnalyticsRange from "@/components/distribution/DistributionAnalyticsRange.vue";
import DistributionBusinessChart from "@/components/distribution/DistributionBusinessChart.vue";
import { analyticsRangeParams, defaultDistributionAnalyticsRange, formatAnalyticsRangeLabel, type DistributionAnalyticsRangeValue } from "@/components/distribution/distributionAnalyticsRange";
import type { Column } from "@/components/common/types";
import {
  getDistributionOverview,
  grantL2Agent,
  listTeam,
  updateTeamAgentStatus,
  type DistributionAgent,
  type DistributionOverview,
  getTeamAgentRewardRule,
  getTeamAgentAnalytics,
  updateTeamAgentRewardRule,
  type DistributionAgentAnalytics,
} from "@/api/distribution";
import { useDistributionAccess } from "@/composables/useDistributionAccess";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage, extractI18nErrorMessage } from "@/utils/apiError";

const app = useAppStore();
const { t } = useI18n();
const router = useRouter();
const { loadDistributionAccess } = useDistributionAccess();
const overview = ref<DistributionOverview | null>(null);
const range = ref(defaultDistributionAnalyticsRange());
const items = ref<DistributionAgent[]>([]);
const loading = ref(false);
const search = ref("");
const status = ref("");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
let searchTimer: number | null = null;
const addDialog = ref(false);
const childEmail = ref("");
const ratePercent = ref("");
const promotionCode = ref("");
const showAdvanced = ref(false);
const creating = ref(false);
const rewardRuleDialog = ref(false);
const rewardRuleSaving = ref(false);
const rewardRuleAgent = ref<DistributionAgent | null>(null);
const rewardRule = ref({ registration_enabled: false, registration_reward_cny: "0", recharge_enabled: false, recharge_threshold_cny: "0", recharge_reward_cny: "0" });
const statusConfirm = ref(false);
const pendingAgent = ref<DistributionAgent | null>(null);
const teamDetailDialog = ref(false);
const detailAgent = ref<DistributionAgent | null>(null);
const detailAnalytics = ref<DistributionAgentAnalytics | null>(null);
const detailLoading = ref(false);
const detailRange = ref(defaultDistributionAnalyticsRange());
const detailRangeLabel = computed(() => formatAnalyticsRangeLabel(detailRange.value));
const columns = computed<Column[]>(() => [
  { key: "agent", label: t('common.distributionCenter.childAgents') },
  { key: "policy", label: t('common.distributionCenter.promotionPolicy') },
  { key: "customers", label: t('common.distributionCenter.customerContribution'), class: "text-right" },
  { key: "business", label: t('common.distributionCenter.paidContribution'), class: "text-right" },
  { key: "commission", label: t('common.distributionCenter.commissionContribution'), class: "text-right" },
  { key: "status", label: t('common.distributionCenter.status') },
  { key: "actions", label: t('common.distributionCenter.actions') },
]);
const maxRatePercent = computed(
  () => (overview.value?.agent.max_child_rate_bps || 0) / 100,
);
const statusConfirmMessage = computed(() =>
  pendingAgent.value?.status === "active"
    ? t('common.distributionCenter.suspendMessage')
    : t('common.distributionCenter.restoreMessage'),
);
const money = (value: string) => `¥${Number(value || 0).toFixed(2)}`;
const percent = (value: string | number) => `${Number(value || 0).toFixed(1)}%`;
const growthText = (value?: string) => value == null ? t('common.distributionAnalytics.previousUnavailable') : t(Number(value) >= 0 ? 'common.distributionAnalytics.comparedUp' : 'common.distributionAnalytics.comparedDown', { value: Number(value).toFixed(1) });
const rate = (bps: number) => `${(bps / 100).toFixed(2)}%`;
const formatDate = (value: string) => new Date(value).toLocaleString();
const statusText = (value: string) =>
  ({ active: t('common.distributionCenter.statusActive'), suspended: t('common.distributionCenter.statusSuspended'), revoked: t('common.distributionCenter.statusRevoked') })[value] || value;
const statusClass = (value: string) =>
  value === "active"
    ? "badge-success"
    : value === "suspended"
      ? "badge-warning"
      : "badge-gray";

async function load() {
  loading.value = true;
  try {
    const [access, summary, result] = await Promise.all([
      loadDistributionAccess(),
      getDistributionOverview(analyticsRangeParams(range.value)),
      listTeam({
        page: pagination.value.page,
        page_size: pagination.value.page_size,
        search: search.value.trim() || undefined,
        status: status.value || undefined,
        ...analyticsRangeParams(range.value),
      }),
    ]);
    if (access.depth !== 1) {
      void router.replace("/distribution");
      return;
    }
    overview.value = summary;
    items.value = result.items;
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    };
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('common.distributionCenter.loadFailed')));
  } finally {
    loading.value = false;
  }
}
function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = window.setTimeout(load, 300);
}
function onRangeChange() { pagination.value.page = 1; void load() }
function changePage(page: number) {
  pagination.value.page = page;
  void load();
}
function changePageSize(pageSize: number) {
  pagination.value.page = 1;
  pagination.value.page_size = pageSize;
  void load();
}
function openAddDialog() {
  addDialog.value = true;
}
async function openRewardRuleDialog(agent: DistributionAgent) {
  rewardRuleAgent.value = agent;
  try {
    const rule = await getTeamAgentRewardRule(agent.id);
    rewardRule.value = { registration_enabled: rule.registration_enabled, registration_reward_cny: rule.registration_reward_cny, recharge_enabled: rule.recharge_enabled, recharge_threshold_cny: rule.recharge_threshold_cny, recharge_reward_cny: rule.recharge_reward_cny };
    rewardRuleDialog.value = true;
  } catch (error) { app.showError(extractApiErrorMessage(error, t('common.distributionRewards.loadShareFailed'))); }
}
async function saveRewardRule() {
  if (!rewardRuleAgent.value) return;
  rewardRuleSaving.value = true;
  try {
    await updateTeamAgentRewardRule(rewardRuleAgent.value.id, rewardRule.value);
    app.showSuccess(t('common.distributionRewards.savedShare'));
    rewardRuleDialog.value = false;
  } catch (error) { app.showError(extractI18nErrorMessage(error, t, "common.errors", t('common.distributionRewards.saveShareFailed'))); }
  finally { rewardRuleSaving.value = false; }
}
function closeAddDialog() {
  addDialog.value = false;
  childEmail.value = "";
  ratePercent.value = "";
  promotionCode.value = "";
  showAdvanced.value = false;
}
async function submitAgent() {
  if (!childEmail.value) return;
  // type="number" may expose a number through v-model; normalize it before
  // checking emptiness so an explicit 0% override remains valid.
  const rateInput = String(ratePercent.value ?? "").trim();
  const parsed = rateInput === "" ? undefined : Number(rateInput);
  if (
    parsed !== undefined &&
    (!Number.isFinite(parsed) || parsed < 0 || parsed > maxRatePercent.value)
  ) {
    app.showError(
      t('common.distributionCenter.rateRangeError', { rate: maxRatePercent.value.toFixed(2) }),
    );
    return;
  }
  creating.value = true;
  try {
    await grantL2Agent({
      email: childEmail.value.toLowerCase(),
      rate_override_bps:
        parsed === undefined ? undefined : Math.round(parsed * 100),
      promotion_code: promotionCode.value.trim().toUpperCase() || undefined,
    });
    app.showSuccess(t('common.distributionCenter.addSuccess'));
    closeAddDialog();
    pagination.value.page = 1;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('common.distributionCenter.addFailed')));
  } finally {
    creating.value = false;
  }
}
function openStatusConfirm(agent: DistributionAgent) {
  pendingAgent.value = agent;
  statusConfirm.value = true;
}
async function openTeamDetail(agent: DistributionAgent) {
  detailAgent.value = agent;
  detailAnalytics.value = null;
  teamDetailDialog.value = true;
  await loadTeamDetailAnalytics();
}
async function loadTeamDetailAnalytics(_range?: DistributionAnalyticsRangeValue) {
  if (!detailAgent.value || !teamDetailDialog.value) return;
  detailLoading.value = true;
  try { detailAnalytics.value = await getTeamAgentAnalytics(detailAgent.value.id, analyticsRangeParams(detailRange.value)); }
  catch (error) { app.showError(extractApiErrorMessage(error, t('common.distributionCenter.detailLoadFailed'))); }
  finally { detailLoading.value = false; }
}
function closeTeamDetail() {
  teamDetailDialog.value = false;
  detailAgent.value = null;
  detailAnalytics.value = null;
}
async function confirmStatus() {
  if (!pendingAgent.value) return;
  const next = pendingAgent.value.status === "active" ? "suspended" : "active";
  try {
    await updateTeamAgentStatus(pendingAgent.value.id, next);
    app.showSuccess(t(next === "active" ? 'common.distributionCenter.restored' : 'common.distributionCenter.suspendedSuccess'));
    statusConfirm.value = false;
    pendingAgent.value = null;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, t('common.distributionCenter.statusUpdateFailed')));
  }
}
watch([search, status], scheduleLoad);
onMounted(load);
</script>

<style scoped>
.team-kpis { display: grid; grid-template-columns: repeat(5,minmax(0,1fr)); overflow: hidden; border: 1px solid var(--ui-border); border-radius: 8px; background: var(--ui-surface); }
.team-kpis article { min-width: 0; padding: 14px; border-right: 1px solid var(--ui-border); }
.team-kpis article:last-child { border-right: 0; }
.team-kpis p,.team-kpis small { display: block; overflow: hidden; color: var(--ui-text-muted); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.team-kpis strong { display: block; margin: 6px 0 3px; overflow: hidden; font-size: 18px; font-weight: 650; text-overflow: ellipsis; white-space: nowrap; font-variant-numeric: tabular-nums; }
.input-prefix {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--ui-text-muted);
  font-size: 0.875rem;
  pointer-events: none;
}
.detail-list{display:grid;gap:10px}.detail-list>div{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.detail-list dt{color:var(--ui-text-muted);font-size:12px}.detail-list dd{text-align:right;font-size:12px;font-weight:650;font-variant-numeric:tabular-nums}.detail-kpis dt{color:var(--ui-text-muted);font-size:11px}.detail-kpis dd{margin-top:5px;font-size:15px;font-weight:650;font-variant-numeric:tabular-nums}
@media (max-width: 900px) { .team-kpis { grid-template-columns: repeat(2,minmax(0,1fr)); }.team-kpis article { border-bottom: 1px solid var(--ui-border); }.team-kpis article:nth-child(2n) { border-right: 0; }.team-kpis article:last-child { grid-column: 1/-1; border-bottom: 0; } }
</style>
