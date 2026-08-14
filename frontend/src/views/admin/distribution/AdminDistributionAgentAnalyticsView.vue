<template>
  <AppLayout
    ><main class="mx-auto w-full max-w-[1440px] space-y-4">
      <section class="analytics-toolbar" :aria-label="t('admin.distribution.agentAnalysis.filters')">
        <DistributionAnalyticsRange
          v-model="range"
          @change="resetAndLoad"
        />
        <div class="segments">
          <button :aria-pressed="depth === 0" @click="setDepth(0)">
            {{ t("admin.distribution.agentAnalysis.all") }}</button
          ><button :aria-pressed="depth === 1" @click="setDepth(1)">
            {{ t("admin.distribution.agentAnalysis.l1") }}</button
          ><button :aria-pressed="depth === 2" @click="setDepth(2)">
            {{ t("admin.distribution.agentAnalysis.l2") }}
          </button>
        </div>
        <label class="search-field">
          <Icon name="search" size="sm" />
          <input
            v-model.trim="search"
            type="search"
            :placeholder="t('admin.distribution.agentAnalysis.searchPlaceholder')"
            @keyup.enter="resetAndLoad"
          />
        </label>
        <select v-model="status" class="input status-filter" :aria-label="t('admin.distribution.agentAnalysis.status')" @change="resetAndLoad">
          <option value="">{{ t("admin.distribution.agentAnalysis.allStatuses") }}</option>
          <option value="active">{{ t("admin.distribution.statusActive") }}</option>
          <option value="suspended">{{ t("admin.distribution.statusSuspended") }}</option>
          <option value="revoked">{{ t("admin.distribution.statusRevoked") }}</option>
        </select>
        <button class="btn btn-secondary" @click="resetAndLoad">
          {{ t("common.search") }}
        </button>
        <button
          class="btn btn-secondary btn-icon"
          :disabled="loading"
          :aria-label="t('common.refresh')"
          @click="load"
        >
          <Icon
            name="refresh"
            size="sm"
            :class="loading ? 'animate-spin' : ''"
          />
        </button>
      </section>
      <AdminDistributionNav />
      <section class="analysis-panel">
        <div class="heading">
          <div>
            <h2>{{ t("admin.distribution.agentAnalysis.comparison") }}</h2>
            <p>{{ rangeLabel }}</p>
          </div>
        </div>
        <section v-if="loadError" class="inline-error" role="alert">
          <Icon name="exclamationCircle" size="sm" />
          <span>{{ loadError }}</span>
          <button type="button" class="btn btn-secondary" @click="load">{{ t("common.retry") }}</button>
        </section>
        <div v-else-if="loading && !agents.length" class="table-loading" aria-live="polite">
          {{ t("admin.distribution.agentAnalysis.loading") }}
        </div>
        <div v-else class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>{{ t("admin.distribution.agentAnalysis.agent") }}</th>
                <th>{{ t("admin.distribution.agentAnalysis.acquisition") }}</th>
                <th>
                  {{ t("admin.distribution.agentAnalysis.paidRetention") }}
                </th>
                <th>
                  {{ t("admin.distribution.agentAnalysis.customerActivity") }}
                </th>
                <th>
                  <button @click="toggleSort('period_customer_paid')">
                    {{ t("admin.distribution.agentAnalysis.netPaid")
                    }}<Icon
                      :name="sortIcon('period_customer_paid')"
                      size="xs"
                    />
                  </button>
                </th>
                <th>
                  <button @click="toggleSort('period_commission')">
                    {{ t("admin.distribution.agentAnalysis.channelCost")
                    }}<Icon :name="sortIcon('period_commission')" size="xs" />
                  </button>
                </th>
                <th>{{ t("admin.distribution.agentAnalysis.team") }}</th>
                <th>{{ t("admin.distribution.agentAnalysis.status") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="agent in agents" :key="agent.id">
                <td>
                  <RouterLink :to="agentLink(agent.id)"
                    ><strong>{{ agent.username || agent.email }}</strong
                    ><small
                      >{{ agent.email }} · L{{ agent.depth }}</small
                    ></RouterLink
                  >
                </td>
                <td>
                  <strong
                    >{{
                      agent.period_customer_count + agent.team_customer_count
                    }}
                    / {{ agent.period_cohort_paid_customers }}</strong
                  ><small
                    >{{ t("admin.distribution.agentAnalysis.newAndPaid") }} ·
                    {{
                      t("admin.distribution.agentAnalysis.activationValue", {
                        value: rate(
                          agent.period_activated_customers,
                          agent.period_customer_count +
                            agent.team_customer_count,
                        ),
                      })
                    }}</small
                  >
                </td>
                <td>
                  <strong
                    >{{ agent.period_all_paying_customers }} /
                    {{ agent.period_repurchase_customers }}</strong
                  ><small
                    >{{
                      t("admin.distribution.agentAnalysis.payingAndRepurchase")
                    }}
                    ·
                    {{
                      rate(
                        agent.period_repurchase_customers,
                        agent.period_all_paying_customers,
                      )
                    }}</small
                  >
                </td>
                <td>
                  <strong
                    >{{
                      (agent.customer_count || 0) +
                      (agent.total_team_customers || 0)
                    }}
                    / {{ agent.period_active_customers }}</strong
                  ><small
                    >{{
                      t("admin.distribution.agentAnalysis.totalAndActive")
                    }}
                    ·
                    {{
                      rate(
                        agent.period_active_customers,
                        (agent.customer_count || 0) +
                          (agent.total_team_customers || 0),
                      )
                    }}</small
                  >
                </td>
                <td>{{ money(paid(agent)) }}</td>
                <td>
                  <strong>{{ money(cost(agent)) }}</strong
                  ><small>{{
                    t("admin.distribution.agentAnalysis.commissionReward", {
                      commission: money(
                        Number(agent.period_commission_cny) +
                          Number(agent.team_commission_cny),
                      ),
                      reward: money(Number(agent.period_reward_cny)),
                    })
                  }}</small>
                </td>
                <td>
                  {{
                    agent.depth === 1
                      ? `${agent.team_count || 0} / ${agent.total_team_customers || 0}`
                      : "—"
                  }}
                </td>
                <td>
                  <span class="badge" :class="statusBadge(agent.status)">{{
                    statusLabel(agent.status)
                  }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="!loadError" class="agent-cards">
          <RouterLink
            v-for="agent in agents"
            :key="agent.id"
            :to="agentLink(agent.id)"
            class="agent-card"
            ><header>
              <div class="agent-identity">
                <strong>{{ agent.username || agent.email }}</strong
                ><span>{{ agent.email }}</span>
              </div>
              <div class="agent-status">
                <span class="level">L{{ agent.depth }}</span
                ><span class="badge" :class="statusBadge(agent.status)">{{
                  statusLabel(agent.status)
                }}</span
                ><Icon name="chevronRight" size="sm" />
              </div>
            </header>
            <dl>
              <div>
                <dt>{{ t("admin.distribution.agentAnalysis.acquisition") }}</dt>
                <dd>
                  {{
                    agent.period_customer_count + agent.team_customer_count
                  }}
                  / {{ agent.period_cohort_paid_customers }}
                </dd>
                <small>{{
                  t("admin.distribution.agentAnalysis.activationValue", {
                    value: rate(
                      agent.period_activated_customers,
                      agent.period_customer_count + agent.team_customer_count,
                    ),
                  })
                }}</small>
              </div>
              <div>
                <dt>
                  {{ t("admin.distribution.agentAnalysis.paidRetention") }}
                </dt>
                <dd>
                  {{ agent.period_all_paying_customers }} /
                  {{ agent.period_repurchase_customers }}
                </dd>
                <small>{{
                  rate(
                    agent.period_repurchase_customers,
                    agent.period_all_paying_customers,
                  )
                }}</small>
              </div>
              <div>
                <dt>
                  {{ t("admin.distribution.agentAnalysis.customerActivity") }}
                </dt>
                <dd>
                  {{
                    (agent.customer_count || 0) +
                    (agent.total_team_customers || 0)
                  }}
                  / {{ agent.period_active_customers }}
                </dd>
                <small>{{
                  rate(
                    agent.period_active_customers,
                    (agent.customer_count || 0) +
                      (agent.total_team_customers || 0),
                  )
                }}</small>
              </div>
              <div>
                <dt>{{ t("admin.distribution.agentAnalysis.netPaid") }}</dt>
                <dd>{{ money(paid(agent)) }}</dd>
              </div>
              <div class="cost">
                <dt>{{ t("admin.distribution.agentAnalysis.channelCost") }}</dt>
                <dd>{{ money(cost(agent)) }}</dd>
                <small>{{
                  t("admin.distribution.agentAnalysis.commissionReward", {
                    commission: money(
                      Number(agent.period_commission_cny) +
                        Number(agent.team_commission_cny),
                    ),
                    reward: money(Number(agent.period_reward_cny)),
                  })
                }}</small>
              </div>
              <div>
                <dt>{{ t("admin.distribution.agentAnalysis.team") }}</dt>
                <dd>
                  {{
                    agent.depth === 1
                      ? `${agent.team_count || 0} / ${agent.total_team_customers || 0}`
                      : "—"
                  }}
                </dd>
              </div>
            </dl></RouterLink
          >
        </div>
        <div v-if="!loadError && !loading && !agents.length" class="empty">
          {{ t("admin.distribution.agentAnalysis.empty") }}
        </div>
        <footer v-if="total > pageSize" class="pagination">
          <button
            class="btn btn-secondary"
            :disabled="page <= 1 || loading"
            @click="
              page--;
              load();
            "
          >
            {{ t("common.previous") }}</button
          ><span>{{
            t("admin.distribution.agentAnalysis.page", {
              page,
              total: Math.ceil(total / pageSize),
            })
          }}</span
          ><button
            class="btn btn-secondary"
            :disabled="page * pageSize >= total || loading"
            @click="
              page++;
              load();
            "
          >
            {{ t("common.next") }}
          </button>
        </footer>
      </section>
    </main></AppLayout
  >
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import AppLayout from "@/components/layout/AppLayout.vue";
import Icon from "@/components/icons/Icon.vue";
import AdminDistributionNav from "@/components/admin/distribution/AdminDistributionNav.vue";
import DistributionAnalyticsRange from "@/components/distribution/DistributionAnalyticsRange.vue";
import {
  analyticsRangeParams,
  defaultDistributionAnalyticsRange,
  formatAnalyticsRangeLabel,
} from "@/components/distribution/distributionAnalyticsRange";
import { listAgents } from "@/api/admin/distribution";
import type { DistributionAgent } from "@/api/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";

const { t } = useI18n();
const app = useAppStore();
const range = ref(defaultDistributionAnalyticsRange());
const depth = ref(0);
const loading = ref(false);
const loadError = ref("");
const agents = ref<DistributionAgent[]>([]);
const search = ref("");
const status = ref("");
const page = ref(1);
const pageSize = 25;
const total = ref(0);
const sortBy = ref("period_customer_paid");
const sortOrder = ref<"asc" | "desc">("desc");
const rangeLabel = computed(() => formatAnalyticsRangeLabel(range.value));
let requestSequence = 0;

async function load() {
  const sequence = ++requestSequence;
  loading.value = true;
  loadError.value = "";
  agents.value = [];
  total.value = 0;
  try {
    const result = await listAgents({
      page: page.value,
      page_size: pageSize,
      search: search.value || undefined,
      status: status.value || undefined,
      depth: depth.value || undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
      ...analyticsRangeParams(range.value),
    });
    if (sequence === requestSequence) {
      agents.value = result.items;
      total.value = result.total;
    }
  } catch (error) {
    if (sequence === requestSequence) {
      loadError.value = extractApiErrorMessage(
        error,
        t("admin.distribution.agentAnalysis.loadFailed"),
      );
      app.showError(loadError.value);
    }
  } finally {
    if (sequence === requestSequence) loading.value = false;
  }
}
function paid(agent: DistributionAgent) {
  return (
    Number(agent.period_customer_paid_cny) +
    Number(agent.team_customer_paid_cny)
  );
}
function cost(agent: DistributionAgent) {
  return (
    Number(agent.period_commission_cny) +
    Number(agent.team_commission_cny) +
    Number(agent.period_reward_cny)
  );
}
function rate(value: number, total: number) {
  return total ? `${((value * 100) / total).toFixed(1)}%` : "—";
}
function money(value: number) {
  return new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 2,
  }).format(value);
}
function resetAndLoad() {
  page.value = 1;
  void load();
}
function setDepth(value: number) {
  depth.value = value;
  resetAndLoad();
}
function toggleSort(value: string) {
  if (sortBy.value === value)
    sortOrder.value = sortOrder.value === "desc" ? "asc" : "desc";
  else {
    sortBy.value = value;
    sortOrder.value = "desc";
  }
  resetAndLoad();
}
function sortIcon(value: string) {
  return sortBy.value !== value
    ? "chevronUp"
    : sortOrder.value === "desc"
      ? "chevronDown"
      : "chevronUp";
}
function statusBadge(value: string) {
  return value === "active"
    ? "badge-success"
    : value === "revoked"
      ? "badge-gray"
      : "badge-warning";
}
function statusLabel(value: string) {
  const normalized = value.charAt(0).toUpperCase() + value.slice(1);
  return t(`admin.distribution.status${normalized}`);
}
function agentLink(id: number) {
  return {
    path: "/admin/distribution/agents",
    query: {
      agent_id: id,
      date_from: range.value.date_from,
      date_to: range.value.date_to,
    },
  };
}
onMounted(load);
</script>
<style scoped>
.analysis-panel {
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface);
  padding: 16px;
}
.analytics-toolbar {
  display: flex;
  min-width: 0;
  align-items: flex-end;
  gap: 8px;
  border-bottom: 1px solid var(--ui-border);
  padding-bottom: 12px;
}
.search-field {
  display: flex;
  min-width: 180px;
  max-width: 280px;
  min-height: 36px;
  flex: 1;
  align-items: center;
  gap: 7px;
  border: 1px solid var(--ui-border);
  border-radius: 6px;
  padding: 0 10px;
  background: var(--ui-surface);
}
.search-field:focus-within {
  border-color: var(--ui-focus);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--ui-focus) 16%, transparent);
}
.search-field input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--ui-text);
  font-size: 12px;
}
.status-filter { width: 132px; min-height: 36px; font-size: 12px; }
.inline-error,
.table-loading {
  display: flex;
  min-height: 180px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.inline-error { flex-wrap: wrap; }
.heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 12px;
}
.heading h2 {
  font-size: 14px;
  font-weight: 650;
}
.heading p {
  margin-top: 3px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.segments {
  display: flex;
  gap: 2px;
}
.segments button {
  min-height: 32px;
  border-radius: 6px;
  padding: 0 12px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.segments button[aria-pressed="true"] {
  border: 1px solid var(--ui-border);
  background: var(--ui-surface-raised);
  color: var(--ui-text);
  font-weight: 650;
}
.filters {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  margin-bottom: 12px;
}
.filters label {
  display: grid;
  gap: 4px;
}
.filters label:first-child {
  width: min(320px, 100%);
}
.filters label:nth-child(2) {
  width: 150px;
}
.filters span {
  color: var(--ui-text-muted);
  font-size: 11px;
}
.filters .input {
  min-height: 36px;
  font-size: 12px;
}
.table-scroll {
  overflow-x: auto;
}
table {
  width: 100%;
  min-width: 1240px;
  border-collapse: collapse;
  font-size: 12px;
}
th,
td {
  padding: 11px 10px;
  border-bottom: 1px solid var(--ui-border);
  text-align: right;
  vertical-align: middle;
  font-variant-numeric: tabular-nums;
}
th {
  color: var(--ui-text-muted);
  font-weight: 600;
  white-space: nowrap;
}
th:first-child,
td:first-child {
  text-align: left;
}
th button {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: inherit;
  font: inherit;
}
tbody tr:hover {
  background: var(--ui-surface-subtle);
}
td strong,
td small {
  display: block;
}
td small {
  margin-top: 3px;
  color: var(--ui-text-subtle);
  white-space: nowrap;
}
td:first-child a {
  display: block;
  border-radius: 5px;
  padding: 2px;
}
td:first-child a:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: 2px;
}
.agent-cards {
  display: none;
}
.empty {
  display: grid;
  min-height: 180px;
  place-items: center;
  color: var(--ui-text-muted);
  font-size: 13px;
}
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 12px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
@media (max-width: 767px) {
  .analytics-toolbar { align-items: stretch; flex-direction: column; }
  .analytics-toolbar > *, .search-field, .status-filter { width: 100%; max-width: none; }
  .analysis-panel {
    padding: 12px;
  }
  .heading {
    flex-direction: column;
  }
  .segments {
    width: 100%;
  }
  .segments button {
    min-height: 44px;
    flex: 1;
  }
  .filters {
    align-items: stretch;
    flex-direction: column;
  }
  .filters label,
  .filters label:first-child,
  .filters label:nth-child(2) {
    width: 100%;
  }
  .filters .input,
  .filters .btn {
    min-height: 44px;
  }
  .table-scroll {
    display: none;
  }
  .agent-cards {
    display: grid;
    border-top: 1px solid var(--ui-border);
  }
  .agent-card {
    display: block;
    min-width: 0;
    padding: 14px 2px;
    border-bottom: 1px solid var(--ui-border);
    color: inherit;
    transition: background-color 150ms ease;
  }
  .agent-card:active {
    background: var(--ui-surface-subtle);
  }
  .agent-card:focus-visible {
    border-radius: 6px;
    outline: 2px solid var(--ui-focus);
    outline-offset: 2px;
  }
  .agent-card header {
    display: flex;
    min-width: 0;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }
  .agent-identity {
    min-width: 0;
  }
  .agent-identity strong,
  .agent-identity span {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .agent-identity strong {
    font-size: 14px;
  }
  .agent-identity span {
    margin-top: 2px;
    color: var(--ui-text-muted);
    font-size: 11px;
  }
  .agent-status {
    display: flex;
    flex: none;
    align-items: center;
    gap: 6px;
  }
  .agent-status .level {
    color: var(--ui-text-muted);
    font-size: 11px;
    font-weight: 650;
  }
  .agent-card dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px 18px;
    margin-top: 14px;
  }
  .agent-card dl > div {
    min-width: 0;
  }
  .agent-card dt {
    color: var(--ui-text-muted);
    font-size: 11px;
  }
  .agent-card dd {
    margin-top: 3px;
    font-size: 14px;
    font-weight: 650;
    font-variant-numeric: tabular-nums;
  }
  .agent-card small {
    display: block;
    margin-top: 2px;
    overflow: hidden;
    color: var(--ui-text-subtle);
    font-size: 10px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .agent-card .cost {
    grid-column: span 2;
  }
  .pagination {
    justify-content: space-between;
  }
}
@media (prefers-reduced-motion: reduce) {
  .agent-card {
    transition: none;
  }
}
</style>
