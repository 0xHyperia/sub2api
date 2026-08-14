<template>
  <AppLayout
    ><main class="workspace">
      <section
        class="analytics-toolbar"
        :aria-label="t('admin.business.analysisScope')"
      >
        <div class="scope-filter">
          <RemoteEntityCombobox
            :model-value="selectedAgent"
            class="agent-picker"
            input-id="business-agent-filter"
            :label="t('admin.business.analysisScope')"
            :placeholder="t('admin.business.agentSearchPlaceholder')"
            :search="searchAgents"
            @update:model-value="applyAgent"
          />
          <div v-if="selectedAgent" class="scope-controls">
            <div
              v-if="selectedAgent.depth === 1"
              class="segments"
              role="group"
              :aria-label="t('admin.business.customerScope')"
            >
              <button
                :aria-pressed="agentScope === 'team'"
                @click="setAgentScope('team')"
              >
                {{ t("admin.business.teamCustomers") }}</button
              ><button
                :aria-pressed="agentScope === 'direct'"
                @click="setAgentScope('direct')"
              >
                {{ t("admin.business.directCustomers") }}
              </button>
            </div>
            <span v-else class="scope-badge">{{
              t("admin.business.directCustomers")
            }}</span>
          </div>
        </div>
        <div class="toolbar">
          <DistributionAnalyticsRange v-model="range" @change="load" /><select
            v-if="!selectedAgent"
            v-model="channel"
            class="input channel"
            :aria-label="t('admin.business.channel')"
            @change="syncAndLoad"
          >
            <option value="">{{ t("admin.business.allChannels") }}</option>
            <option v-for="item in channels" :key="item" :value="item">
              {{ channelLabel(item) }}
            </option></select
          ><button
            class="btn btn-secondary btn-icon"
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
      </section>
      <nav class="section-nav" :aria-label="t('admin.business.title')">
        <RouterLink
          v-for="item in sections"
          :key="item.key"
          :to="routeFor(item.key)"
          >{{ item.label }}</RouterLink
        >
      </nav>
      <div
        v-if="delayedLoading && (!data || dataSection !== section)"
        class="skeleton"
        aria-live="polite"
      >
        <span v-for="n in 6" :key="n" />
      </div>
      <section v-else-if="loadError" class="load-error" role="alert">
        <Icon name="exclamationCircle" size="lg" /><strong>{{
          t("admin.business.loadFailed")
        }}</strong>
        <p>{{ loadError }}</p>
        <button type="button" class="btn btn-secondary" @click="load">
          {{ t("common.retry") }}
        </button>
      </section>
      <template v-else-if="data && dataSection === section">
        <div v-if="loading" class="updating" aria-live="polite">
          {{ t("admin.business.updating") }}
        </div>
        <div
          v-if="data.warnings?.includes('usage_history_incomplete')"
          class="coverage-warning"
          role="status"
        >
          <Icon name="infoCircle" size="sm" />
          {{
            t("admin.business.usageHistoryIncomplete", {
              date: data.usage_data_from || "—",
            })
          }}
        </div>
        <section class="kpis">
          <article v-for="metric in visibleMetrics" :key="metric.key">
            <span
              >{{ metric.label
              }}<MetricInfo
                :label="metric.label"
                :description="metricHelp(metric.key)" /></span
            ><strong>{{ formatMetric(metric.key, metric.value.value) }}</strong
            ><small
              :class="growthClass(metric.key, metric.value.change_value)"
              >{{ comparison(metric.value) }}</small
            >
          </article>
        </section>
        <section v-if="section === 'overview'" class="trend-grid">
          <BusinessAnalyticsChart
            :points="data.trend"
            :previous-points="data.previous_trend"
            :totals="trendTotals"
            :title="t('admin.business.revenueTrend')"
            :subtitle="rangeLabel"
            :granularity="data.granularity"
            :default-metrics="['net_paid_cny']"
            :available-metrics="[
              'net_paid_cny',
              'gross_paid_cny',
              'refunded_cny',
              'consumed_revenue',
            ]"
          /><BusinessAnalyticsChart
            :points="data.trend"
            :previous-points="data.previous_trend"
            :totals="trendTotals"
            :title="t('admin.business.userTrend')"
            :subtitle="rangeLabel"
            :granularity="data.granularity"
            :default-metrics="['new_users', 'active_users']"
            :available-metrics="[
              'new_users',
              'active_users',
              'first_paid_users',
              'paying_users',
              'repurchase_users',
            ]"
          />
        </section>
        <BusinessAnalyticsChart
          v-else-if="
            section !== 'retention' &&
            !(section === 'finance' && financeTab === 'balance')
          "
          :points="data.trend"
          :previous-points="data.previous_trend"
          :totals="trendTotals"
          :title="page.chartTitle"
          :subtitle="rangeLabel"
          :granularity="data.granularity"
          :default-metrics="page.defaultMetrics"
          :available-metrics="page.trendMetrics"
        />
        <template v-if="section === 'overview'"
          ><section class="two-cols">
            <div class="panel">
              <div class="panel-title">
                <h2>{{ t("admin.business.funnel") }}</h2>
                <p>{{ t("admin.business.funnelHint") }}</p>
              </div>
              <div class="funnel">
                <div v-for="(step, index) in data.funnel" :key="step.key">
                  <span>{{ t(`admin.business.funnel_${step.key}`) }}</span
                  ><strong>{{ step.count }}</strong
                  ><small v-if="index">{{
                    funnelConversion(step.key, step.count)
                  }}</small
                  ><i :style="{ width: `${funnelWidth(step.count)}%` }" />
                </div>
              </div>
            </div>
            <BusinessChannelTable
              :rows="data.channels"
              compact
              @select="openChannel"
            /></section
        ></template>
        <template v-else-if="section === 'growth'"
          ><section class="two-cols">
            <div class="panel">
              <div class="panel-title">
                <h2>{{ t("admin.business.growthStructure") }}</h2>
                <p>{{ t("admin.business.growthStructureHint") }}</p>
              </div>
              <dl class="detail-grid">
                <div
                  v-for="key in [
                    'new_users',
                    'activated_users',
                    'first_paid_users',
                    'repurchase_users',
                    'active_users',
                  ]"
                  :key="key"
                >
                  <dt>{{ metricLabel(key) }}</dt>
                  <dd>
                    {{ formatMetric(key, data.metrics[key]?.value || 0) }}
                  </dd>
                </div>
              </dl>
            </div>
            <div class="panel">
              <div class="panel-title">
                <h2>{{ t("admin.business.conversionTime") }}</h2>
                <p>{{ t("admin.business.conversionTimeHint") }}</p>
              </div>
              <div class="duration-bars">
                <div v-for="item in data.duration_distribution" :key="item.key">
                  <span>{{ t(`admin.business.duration_${item.key}`) }}</span
                  ><i
                    ><b
                      :style="{
                        width: durationWidth(item.activation, 'activation'),
                      }" /></i
                  ><strong>{{ item.activation }}</strong
                  ><i
                    ><b
                      class="paid"
                      :style="{
                        width: durationWidth(item.first_paid, 'first_paid'),
                      }" /></i
                  ><strong>{{ item.first_paid }}</strong>
                </div>
                <footer>
                  <span>{{ t("admin.business.activatedUsers") }}</span
                  ><span>{{ t("admin.business.firstPaidUsers") }}</span>
                </footer>
              </div>
            </div>
          </section>
          <BusinessChannelTable
            :rows="data.channels"
            compact
            @select="openChannel"
        /></template>
        <template v-else-if="section === 'finance'"
          ><section class="panel">
            <div class="panel-title with-tabs">
              <div>
                <h2>{{ financeTitle }}</h2>
                <p>{{ financeHint }}</p>
              </div>
              <div class="segments">
                <button
                  v-for="tab in financeTabs"
                  :key="tab"
                  :aria-pressed="financeTab === tab"
                  @click="setFinanceTab(tab)"
                >
                  {{
                    t(
                      `admin.business.${tab === "cash" ? "cashBusiness" : tab === "profit" ? "consumptionProfit" : "userBalance"}`,
                    )
                  }}
                </button>
              </div>
            </div>
            <dl v-if="financeTab !== 'balance'" class="finance-grid">
              <div v-for="key in financeKeys" :key="key">
                <dt>
                  {{ metricLabel(key)
                  }}<MetricInfo
                    :label="metricLabel(key)"
                    :description="metricHelp(key)"
                  />
                </dt>
                <dd>{{ formatMetric(key, data.metrics[key]?.value || 0) }}</dd>
                <small
                  >{{ t("admin.business.previous") }}
                  {{
                    formatMetric(key, data.metrics[key]?.previous || 0)
                  }}</small
                >
              </div>
            </dl>
            <template v-else
              ><div class="balance-toolbar">
                <div class="segments">
                  <button
                    v-for="window in balanceWindows"
                    :key="window"
                    :aria-pressed="balanceWindow === window"
                    @click="setBalanceWindow(window)"
                  >
                    {{ t(`admin.business.balanceWindow${window}`) }}
                  </button>
                </div>
                <span v-if="balance">{{
                  t("admin.business.balanceAsOf", {
                    time: new Date(balance.as_of).toLocaleString(),
                  })
                }}</span>
              </div>
              <section v-if="balanceError" class="balance-error" role="alert">
                <Icon name="exclamationCircle" size="sm" />
                <span>{{ balanceError }}</span>
                <button type="button" class="btn btn-secondary" @click="loadBalance">
                  {{ t("common.retry") }}
                </button>
              </section>
              <div v-else-if="balanceLoading && !balance" class="balance-loading">
                {{ t("admin.business.updating") }}
              </div>
              <template v-else-if="balance"
                ><dl class="finance-grid balance-kpis">
                  <div v-for="item in balanceMetrics" :key="item.key">
                    <dt>
                      {{ item.label
                      }}<MetricInfo
                        :label="item.label"
                        :description="t(`admin.business.help_${item.key}`)"
                      />
                    </dt>
                    <dd>{{ item.value }}</dd>
                    <small v-if="item.reference">{{ item.reference }}</small>
                  </div>
                </dl>
                <div class="balance-note">
                  {{ t("admin.business.balanceCurrentHint") }}
                </div>
                <div class="table-scroll">
                  <table>
                    <thead>
                      <tr>
                        <th>{{ t("admin.business.lifecycle") }}</th>
                        <th>{{ t("admin.business.users") }}</th>
                        <th>{{ t("admin.business.availableBalance") }}</th>
                        <th>{{ t("admin.business.frozenBalance") }}</th>
                        <th>{{ t("admin.business.averageBalance") }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="row in balance.segments" :key="row.key">
                        <td>
                          {{ t(`admin.business.balanceSegment_${row.key}`) }}
                        </td>
                        <td>{{ row.users }}</td>
                        <td>
                          {{ formatUSD(row.balance_usd)
                          }}<small>{{
                            formatCNY(balanceCNY(row.balance_usd))
                          }}</small>
                        </td>
                        <td>
                          {{ formatUSD(row.frozen_usd)
                          }}<small>{{
                            formatCNY(balanceCNY(row.frozen_usd))
                          }}</small>
                        </td>
                        <td>
                          {{
                            formatUSD(
                              row.users ? row.balance_usd / row.users : 0,
                            )
                          }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div></template
              ></template
            >
          </section>
          <section v-if="financeTab === 'cash'" class="panel">
            <div class="panel-title">
              <h2>{{ t("admin.business.originalCurrency") }}</h2>
              <p>{{ t("admin.business.originalCurrencyHint") }}</p>
            </div>
            <div class="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>{{ t("admin.business.currency") }}</th>
                    <th>{{ metricLabel("gross_paid") }}</th>
                    <th>{{ metricLabel("refunded") }}</th>
                    <th>{{ metricLabel("net_paid") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="row in data.currency_breakdown"
                    :key="row.currency"
                  >
                    <td>{{ row.currency }}</td>
                    <td>{{ formatOriginal(row.gross, row.currency) }}</td>
                    <td>{{ formatOriginal(row.refunded, row.currency) }}</td>
                    <td>{{ formatOriginal(row.net, row.currency) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section></template
        >
        <section v-else-if="section === 'retention'" class="retention-layout">
          <div class="panel">
            <div class="panel-title with-tabs">
              <div>
                <h2>{{ t("admin.business.retentionMatrix") }}</h2>
                <p>{{ t("admin.business.retentionHint") }}</p>
              </div>
              <div class="segments">
                <button
                  :aria-pressed="cohort === 'registration'"
                  @click="setCohort('registration')"
                >
                  {{ t("admin.business.registrationRetention") }}</button
                ><button
                  :aria-pressed="cohort === 'activation'"
                  @click="setCohort('activation')"
                >
                  {{ t("admin.business.activationRetention") }}
                </button>
              </div>
            </div>
            <div class="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>
                      {{
                        t(
                          cohort === "registration"
                            ? "admin.business.registrationDate"
                            : "admin.business.firstUseDate",
                        )
                      }}
                    </th>
                    <th>
                      {{
                        t(
                          cohort === "registration"
                            ? "admin.business.registeredUsers"
                            : "admin.business.firstUseUsers",
                        )
                      }}
                    </th>
                    <th>D1</th>
                    <th>D7</th>
                    <th>D30</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in cohortRows" :key="row.cohort_date">
                    <td>{{ row.cohort_date }}</td>
                    <td>{{ row.users }}</td>
                    <td v-for="key in ['d1', 'd7', 'd30']" :key="key">
                      <span
                        class="heat"
                        :style="
                          heatStyle(
                            row[key as keyof typeof row] as number | undefined,
                          )
                        "
                        >{{
                          row[key as keyof typeof row] == null
                            ? "—"
                            : `${Number(row[key as keyof typeof row]).toFixed(1)}%`
                        }}</span
                      >
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="panel lifecycle">
            <div class="panel-title">
              <h2>{{ t("admin.business.lifecycle") }}</h2>
              <p>{{ t("admin.business.lifecycleHint") }}</p>
            </div>
            <dl>
              <div v-for="item in lifecycleItems" :key="item.key">
                <dt>
                  {{ item.label
                  }}<MetricInfo
                    :label="item.label"
                    :description="
                      t(`admin.business.help_lifecycle_${item.key}`)
                    "
                  />
                </dt>
                <dd>{{ item.value }}</dd>
                <small v-if="item.detail">{{ item.detail }}</small>
              </div>
            </dl>
          </div>
        </section>
        <BusinessChannelTable
          v-else
          :rows="data.channels"
          @select="openChannel"
        />
        <footer class="data-footer">
          <span>{{
            t("admin.business.updatedAt", {
              time: new Date(data.updated_at).toLocaleString(),
            })
          }}</span
          ><span v-if="data.estimated" class="badge badge-warning">{{
            t("admin.business.estimatedFx")
          }}</span
          ><button class="btn btn-ghost" @click="exportCSV">
            <Icon name="download" size="sm" />{{
              t("admin.business.exportData")
            }}
          </button>
        </footer>
      </template>
    </main></AppLayout
  >
</template>
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import Icon from "@/components/icons/Icon.vue";
import RemoteEntityCombobox from "@/components/admin/distribution/RemoteEntityCombobox.vue";
import type { DistributionPickerOption } from "@/components/admin/distribution/types";
import DistributionAnalyticsRange from "@/components/distribution/DistributionAnalyticsRange.vue";
import {
  analyticsRangeParams,
  defaultDistributionAnalyticsRange,
  formatAnalyticsRangeLabel,
  type DistributionAnalyticsRangeValue,
} from "@/components/distribution/distributionAnalyticsRange";
import BusinessAnalyticsChart from "@/components/business/BusinessAnalyticsChart.vue";
import BusinessChannelTable from "@/components/business/BusinessChannelTable.vue";
import MetricInfo from "@/components/business/MetricInfo.vue";
import { lookupAgents } from "@/api/admin/distribution";
import {
  getBusinessAnalytics,
  getBusinessBalance,
  type BusinessAnalyticsSection,
  type BusinessAnalyticsSnapshot,
  type BusinessBalanceSnapshot,
  type BusinessChannel,
  type BusinessDurationBucket,
} from "@/api/businessAnalytics";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { businessFunnelRate } from "./businessFunnel";
type FinanceTab = "cash" | "profit" | "balance";
type AgentScope = "team" | "direct";
const props = defineProps<{ section: BusinessAnalyticsSection }>();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const app = useAppStore();
const data = ref<BusinessAnalyticsSnapshot | null>(null);
const dataSection = ref<BusinessAnalyticsSection | null>(null);
const dataSnapshotKey = ref("");
const loadError = ref("");
const balance = ref<BusinessBalanceSnapshot | null>(null);
const balanceError = ref("");
const balanceSnapshotKey = ref("");
const loading = ref(false);
const balanceLoading = ref(false);
const delayedLoading = ref(false);
const financeTab = ref<FinanceTab>((route.query.tab as FinanceTab) || "cash");
const balanceWindow = ref<1 | 7 | 30>(
  Number(route.query.balance_window) === 1
    ? 1
    : Number(route.query.balance_window) === 30
      ? 30
      : 7,
);
const cohort = ref<"registration" | "activation">(
  (route.query.cohort as "registration" | "activation") || "registration",
);
const channel = ref(String(route.query.channel || ""));
const selectedAgent = ref<DistributionPickerOption | null>(null);
const agentScope = ref<AgentScope>(
  route.query.agent_scope === "direct" ? "direct" : "team",
);
let timer: number | undefined;
let sequence = 0;
let balanceSequence = 0;
let routeRestoreSequence = 0;
const parseRange = (): DistributionAnalyticsRangeValue => {
  const fallback = defaultDistributionAnalyticsRange();
  const from = String(route.query.date_from || fallback.date_from),
    to = String(route.query.date_to || fallback.date_to);
  return {
    granularity: "custom",
    preset: "custom",
    date_from: from,
    date_to: to,
  };
};
const range = ref(parseRange());
const section = computed(() => props.section);
const channels: BusinessChannel[] = [
  "distribution",
  "affiliate",
  "campaign",
  "organic",
  "unknown",
];
const configs = computed(() => ({
  overview: {
    title: t("admin.business.overviewTitle"),
    chartTitle: t("admin.business.operationTrend"),
    defaultMetrics: ["net_paid_cny"] as const,
    trendMetrics: ["net_paid_cny", "new_users", "active_users"] as const,
    metrics: [
      "new_users",
      "paid_conversion_rate",
      "net_paid",
      "consumption_margin_rate",
      "repurchase_rate",
      "active_rate",
    ],
  },
  growth: {
    title: t("admin.business.growthTitle"),
    chartTitle: t("admin.business.userTrend"),
    defaultMetrics: ["new_users", "active_users"] as const,
    trendMetrics: [
      "new_users",
      "active_users",
      "first_paid_users",
      "paying_users",
      "repurchase_users",
    ] as const,
    metrics: [
      "new_users",
      "activated_users",
      "first_paid_users",
      "paid_conversion_rate",
      "repurchase_users",
      "active_users",
    ],
  },
  finance: {
    title: t("admin.business.financeTitle"),
    chartTitle: t("admin.business.financeTrend"),
    defaultMetrics: ["net_paid_cny"] as const,
    trendMetrics: [
      "net_paid_cny",
      "consumed_revenue",
      "supplier_cost",
    ] as const,
    metrics: [
      "gross_paid",
      "net_paid",
      "refunded",
      "refund_rate",
      "consumed_revenue",
      "consumption_margin_rate",
    ],
  },
  retention: {
    title: t("admin.business.retentionTitle"),
    chartTitle: "",
    defaultMetrics: ["active_users"] as const,
    trendMetrics: ["active_users"] as const,
    metrics: [
      "new_users",
      "activated_users",
      "active_users",
      "active_rate",
      "repurchase_users",
      "repurchase_rate",
    ],
  },
  channels: {
    title: t("admin.business.channelsTitle"),
    chartTitle: t("admin.business.channelTrend"),
    defaultMetrics: ["net_paid_cny"] as const,
    trendMetrics: ["net_paid_cny", "new_users", "active_users"] as const,
    metrics: [
      "new_users",
      "activated_users",
      "first_paid_users",
      "paying_users",
      "net_paid",
      "active_users",
    ],
  },
}));
const page = computed(() => configs.value[section.value]);
const sections = computed(() => [
  { key: "overview" as const, label: t("admin.business.overview") },
  { key: "growth" as const, label: t("admin.business.growth") },
  { key: "finance" as const, label: t("admin.business.finance") },
  { key: "retention" as const, label: t("admin.business.retention") },
  { key: "channels" as const, label: t("admin.business.channels") },
]);
const rangeLabel = computed(() => formatAnalyticsRangeLabel(range.value));
const visibleMetrics = computed(() =>
  page.value.metrics.map((key) => ({
    key,
    label: metricLabel(key),
    value: data.value?.metrics[key] || { value: 0, previous: 0 },
  })),
);
const financeTabs: FinanceTab[] = ["cash", "profit", "balance"];
const balanceWindows: Array<1 | 7 | 30> = [1, 7, 30];
const financeKeys = computed(() =>
  financeTab.value === "cash"
    ? ["gross_paid", "refunded", "net_paid", "refund_rate"]
    : [
        "consumed_revenue",
        "supplier_cost",
        "consumption_profit",
        "consumption_margin_rate",
      ],
);
const financeTitle = computed(() =>
  financeTab.value === "cash"
    ? t("admin.business.cashBusiness")
    : financeTab.value === "profit"
      ? t("admin.business.consumptionProfit")
      : t("admin.business.userBalance"),
);
const financeHint = computed(() =>
  financeTab.value === "cash"
    ? t("admin.business.cashHint")
    : financeTab.value === "profit"
      ? t("admin.business.profitHint")
      : t("admin.business.balanceHint"),
);
const cohortRows = computed(() =>
  cohort.value === "registration"
    ? data.value?.registration_cohorts || []
    : data.value?.activation_cohorts || [],
);
const lifecycleItems = computed(() => {
  const value = data.value?.lifecycle;
  if (!value) return [];
  return [
    { key: "new", label: t("admin.business.lifecycle_new"), value: value.new },
    {
      key: "active",
      label: t("admin.business.lifecycle_active"),
      value: value.newly_activated + value.continuously_active,
      detail: t("admin.business.lifecycleActiveDetail", {
        newly: value.newly_activated,
        continuous: value.continuously_active,
      }),
    },
    {
      key: "recalled",
      label: t("admin.business.lifecycle_recalled"),
      value: value.silent_reactivated + value.churned_reactivated,
      detail: t("admin.business.lifecycleRecallDetail", {
        silent: value.silent_reactivated,
        churned: value.churned_reactivated,
      }),
    },
    {
      key: "silent",
      label: t("admin.business.lifecycle_silent"),
      value: value.silent,
    },
    {
      key: "churned",
      label: t("admin.business.lifecycle_churned"),
      value: value.churned,
    },
    {
      key: "unactivated",
      label: t("admin.business.lifecycle_unactivated"),
      value: value.unactivated,
    },
  ];
});
const trendTotals = computed(() => {
  const m = data.value?.metrics;
  return {
    net_paid_cny: {
      value: m?.net_paid?.value || 0,
      previous: m?.net_paid?.previous || 0,
    },
    gross_paid_cny: {
      value: m?.gross_paid?.value || 0,
      previous: m?.gross_paid?.previous || 0,
    },
    refunded_cny: {
      value: m?.refunded?.value || 0,
      previous: m?.refunded?.previous || 0,
    },
    consumed_revenue: {
      value: m?.consumed_revenue?.value || 0,
      previous: m?.consumed_revenue?.previous || 0,
    },
    supplier_cost: {
      value: m?.supplier_cost?.value || 0,
      previous: m?.supplier_cost?.previous || 0,
    },
    new_users: {
      value: m?.new_users?.value || 0,
      previous: m?.new_users?.previous || 0,
    },
    active_users: {
      value: m?.active_users?.value || 0,
      previous: m?.active_users?.previous || 0,
    },
    first_paid_users: {
      value: m?.first_paid_users?.value || 0,
      previous: m?.first_paid_users?.previous || 0,
    },
    paying_users: {
      value: m?.paying_users?.value || 0,
      previous: m?.paying_users?.previous || 0,
    },
    repurchase_users: {
      value: m?.repurchase_users?.value || 0,
      previous: m?.repurchase_users?.previous || 0,
    },
  };
});
function metricLabel(key: string) {
  return t(`admin.business.metric_${key}`);
}
function metricHelp(key: string) {
  return t(`admin.business.help_${key}`);
}
function channelLabel(key: string) {
  return t(`admin.business.channel_${key}`);
}
function fixedCurrency(symbol: string, v: number) {
  return `${symbol}${new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)}`;
}
function formatCNY(v: number) {
  return fixedCurrency("￥", v);
}
function formatUSD(v: number) {
  return fixedCurrency("$", v);
}
function balanceCNY(v: number) {
  const multiplier = balance.value?.balance_recharge_multiplier || 1;
  return v / multiplier;
}
function formatOriginal(v: number, currency: string) {
  if (currency === "CNY") return formatCNY(v);
  if (currency === "USD") return formatUSD(v);
  return `${currency} ${new Intl.NumberFormat(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)}`;
}
function formatMetric(key: string, v: number) {
  if (["gross_paid", "refunded", "net_paid"].includes(key)) return formatCNY(v);
  if (["consumed_revenue", "supplier_cost", "consumption_profit"].includes(key))
    return formatUSD(v);
  if (key.includes("rate")) return `${v.toFixed(1)}%`;
  return new Intl.NumberFormat().format(Math.round(v));
}
function comparison(metric: import("@/api/businessAnalytics").BusinessMetric) {
  if (metric.comparison_type === "percentage_point")
    return t("admin.business.pointsChange", {
      value: `${Number(metric.change_value || 0) >= 0 ? "+" : ""}${Number(metric.change_value || 0).toFixed(1)}`,
    });
  if (metric.comparison_type === "turned_positive")
    return t("admin.business.turnedPositive");
  if (metric.comparison_type === "turned_negative")
    return t("admin.business.turnedNegative");
  return metric.change_rate == null
    ? t("admin.business.noComparison")
    : t("admin.business.vsPrevious", {
        value: `${metric.change_rate >= 0 ? "+" : ""}${metric.change_rate.toFixed(1)}%`,
      });
}
const inverseMetrics = new Set([
  "refunded",
  "refund_rate",
  "supplier_cost",
  "commission_cost_rate",
  "reward_cost_rate",
]);
function growthClass(key: string, v?: number) {
  if (v == null || v === 0) return "neutral";
  const increased = v > 0;
  return inverseMetrics.has(key)
    ? increased
      ? "negative"
      : "positive"
    : increased
      ? "positive"
      : "negative";
}
function funnelConversion(key: string, value: number) {
  const result = businessFunnelRate(data.value?.funnel || [], key, value);
  if (result.rate == null) return "—";
  return t(
    result.base === "first_paid"
      ? "admin.business.funnelRateOfFirstPaid"
      : "admin.business.funnelRateOfRegistered",
    { value: result.rate.toFixed(1) + "%" },
  );
}
function funnelWidth(v: number) {
  const first = data.value?.funnel[0]?.count || 0;
  return first ? Math.max(8, (v * 100) / first) : 8;
}
function durationWidth(
  value: number,
  key: keyof Pick<BusinessDurationBucket, "activation" | "first_paid">,
) {
  const max = Math.max(
    ...(data.value?.duration_distribution.map((x) => x[key]) || [1]),
    1,
  );
  return `${Math.max(value ? 5 : 0, (value * 100) / max)}%`;
}
function routeFor(key: BusinessAnalyticsSection) {
  return {
    path: `/admin/business-analytics/${key}`,
    query: {
      date_from: range.value.date_from,
      date_to: range.value.date_to,
      ...(!selectedAgent.value && channel.value
        ? { channel: channel.value }
        : {}),
      ...(selectedAgent.value
        ? { agent_id: selectedAgent.value.id, agent_scope: agentScope.value }
        : {}),
    },
  };
}
function openChannel(value: string) {
  if (value === "distribution")
    void router.push({
      path: "/admin/distribution/overview",
      query: { date_from: range.value.date_from, date_to: range.value.date_to },
    });
  else {
    channel.value = value;
    syncAndLoad();
  }
}
function heatStyle(v?: number) {
  return v == null
    ? {}
    : {
        background: `color-mix(in srgb,var(--ui-success) ${Math.max(8, Math.min(70, v))}%,var(--ui-surface-subtle))`,
      };
}
async function setFinanceTab(value: FinanceTab) {
  financeTab.value = value;
  await router.replace({ query: { ...route.query, tab: value } });
  if (value === "balance" && !balance.value) void loadBalance();
}
function setBalanceWindow(value: 1 | 7 | 30) {
  balanceWindow.value = value;
  void router.replace({ query: { ...route.query, balance_window: value } });
  void loadBalance();
}
function setCohort(value: "registration" | "activation") {
  cohort.value = value;
  void router.replace({ query: { ...route.query, cohort: value } });
}
const balanceMetrics = computed(() => {
  if (!balance.value) return [];
  const b = balance.value;
  const items = [
    {
      key: "available_balance",
      label: t("admin.business.availableBalance"),
      value: formatUSD(b.available_balance_usd),
      reference: formatCNY(balanceCNY(b.available_balance_usd)),
    },
    {
      key: "frozen_balance",
      label: t("admin.business.frozenBalance"),
      value: formatUSD(b.frozen_balance_usd),
      reference: formatCNY(balanceCNY(b.frozen_balance_usd)),
    },
    {
      key: "positive_balance_users",
      label: t("admin.business.positiveBalanceUsers"),
      value: String(b.positive_balance_users),
      reference: "",
    },
    {
      key: "average_balance",
      label: t("admin.business.averageBalance"),
      value: formatUSD(b.average_balance_usd),
      reference: formatCNY(balanceCNY(b.average_balance_usd)),
    },
  ];
  if (b.low_balance_enabled)
    items.push({
      key: "low_balance_users",
      label: t("admin.business.lowBalanceUsers"),
      value: String(b.low_balance_users),
      reference: "",
    });
  return items;
});
async function searchAgents(
  query: string,
): Promise<DistributionPickerOption[]> {
  const result = await lookupAgents(query, { include_inactive: true });
  return result
    .filter((agent) => agent.status !== "revoked")
    .map((agent) => ({
      id: agent.agent_id,
      email: agent.email,
      username: agent.username,
      depth: agent.depth,
      status: agent.status,
      meta: `${agent.depth === 1 ? t("admin.business.levelOneAgent") : t("admin.business.levelTwoAgent")} · ${agent.promotion_code}`,
    }));
}
function analyticsParams() {
  return {
    ...analyticsRangeParams(range.value),
    channel: selectedAgent.value ? undefined : channel.value || undefined,
    agent_id: selectedAgent.value?.id,
    agent_scope: selectedAgent.value ? agentScope.value : undefined,
  };
}
function analyticsSnapshotKey() {
  return [
    section.value,
    range.value.date_from,
    range.value.date_to,
    selectedAgent.value?.id || 0,
    selectedAgent.value ? agentScope.value : channel.value || "platform",
  ].join(":");
}
function syncURL() {
  return router.replace({
    query: {
      ...route.query,
      date_from: range.value.date_from,
      date_to: range.value.date_to,
      channel: selectedAgent.value ? undefined : channel.value || undefined,
      agent_id: selectedAgent.value?.id,
      agent_scope: selectedAgent.value ? agentScope.value : undefined,
    },
  });
}
async function applyAgent(value: DistributionPickerOption | null) {
  if (!value && !selectedAgent.value) return;
  selectedAgent.value = value;
  if (!value) agentScope.value = "team";
  else {
    channel.value = "";
    agentScope.value = value.depth === 2 ? "direct" : "team";
  }
  balance.value = null;
  await syncURL();
  void load();
  if (section.value === "finance" && financeTab.value === "balance")
    void loadBalance();
}
async function setAgentScope(value: AgentScope) {
  agentScope.value = value;
  balance.value = null;
  await syncURL();
  void load();
  if (section.value === "finance" && financeTab.value === "balance")
    void loadBalance();
}
async function restoreAgent() {
  const id = Number(route.query.agent_id || 0);
  if (!id) return;
  const agents = await lookupAgents(String(id), { include_inactive: true });
  const agent = agents.find((item) => item.agent_id === id);
  if (agent) {
    selectedAgent.value = {
      id: agent.agent_id,
      email: agent.email,
      username: agent.username,
      depth: agent.depth,
      status: agent.status,
      meta: `${agent.depth === 1 ? t("admin.business.levelOneAgent") : t("admin.business.levelTwoAgent")} · ${agent.promotion_code}`,
    };
    channel.value = "";
    agentScope.value =
      agent.depth === 2
        ? "direct"
        : route.query.agent_scope === "direct"
          ? "direct"
          : "team";
    await syncURL();
  }
}
async function restoreStateFromRoute() {
  const restoreID = ++routeRestoreSequence;
  const routeRange = parseRange();
  const routeAgentID = Number(route.query.agent_id || 0);
  const routeScope: AgentScope =
    route.query.agent_scope === "direct" ? "direct" : "team";
  const routeChannel = String(route.query.channel || "");
  const routeFinanceTab: FinanceTab = ["cash", "profit", "balance"].includes(
    String(route.query.tab),
  )
    ? (route.query.tab as FinanceTab)
    : "cash";
  const routeBalanceWindow: 1 | 7 | 30 =
    Number(route.query.balance_window) === 1
      ? 1
      : Number(route.query.balance_window) === 30
        ? 30
        : 7;
  const routeCohort: "registration" | "activation" =
    route.query.cohort === "activation" ? "activation" : "registration";
  const changed =
    range.value.date_from !== routeRange.date_from ||
    range.value.date_to !== routeRange.date_to ||
    (selectedAgent.value?.id || 0) !== routeAgentID ||
    agentScope.value !== routeScope ||
    (!routeAgentID && channel.value !== routeChannel) ||
    financeTab.value !== routeFinanceTab ||
    balanceWindow.value !== routeBalanceWindow ||
    cohort.value !== routeCohort;
  if (!changed) return;

  range.value = routeRange;
  agentScope.value = routeScope;
  channel.value = routeAgentID ? "" : routeChannel;
  financeTab.value = routeFinanceTab;
  balanceWindow.value = routeBalanceWindow;
  cohort.value = routeCohort;
  if (!routeAgentID) selectedAgent.value = null;
  else if (selectedAgent.value?.id !== routeAgentID) {
    selectedAgent.value = null;
    const agents = await lookupAgents(String(routeAgentID), {
      include_inactive: true,
    });
    if (restoreID !== routeRestoreSequence) return;
    const agent = agents.find((item) => item.agent_id === routeAgentID);
    if (agent)
      selectedAgent.value = {
        id: agent.agent_id,
        email: agent.email,
        username: agent.username,
        depth: agent.depth,
        status: agent.status,
        meta: `${agent.depth === 1 ? t("admin.business.levelOneAgent") : t("admin.business.levelTwoAgent")} · ${agent.promotion_code}`,
      };
  }
  balance.value = null;
  void load();
  if (section.value === "finance" && financeTab.value === "balance")
    void loadBalance();
}
async function loadBalance() {
  const id = ++balanceSequence;
  const requestedWindow = balanceWindow.value;
  const requestedAgentID = selectedAgent.value?.id;
  const requestedScope = selectedAgent.value ? agentScope.value : undefined;
  const requestedKey = `${requestedWindow}:${requestedAgentID || 0}:${requestedScope || "platform"}`;
  balanceLoading.value = true;
  balanceError.value = "";
  if (balanceSnapshotKey.value !== requestedKey) balance.value = null;
  try {
    const next = await getBusinessBalance(
      requestedWindow,
      requestedAgentID,
      requestedScope,
    );
    if (id === balanceSequence) {
      balance.value = next;
      balanceSnapshotKey.value = requestedKey;
    }
  } catch (e) {
    if (id === balanceSequence) {
      balance.value = null;
      balanceSnapshotKey.value = "";
      balanceError.value = extractApiErrorMessage(
        e,
        t("admin.business.balanceLoadFailed"),
      );
      app.showError(balanceError.value);
    }
  } finally {
    if (id === balanceSequence) balanceLoading.value = false;
  }
}
async function load() {
  const id = ++sequence;
  const requestedSection = section.value;
  const requestedKey = analyticsSnapshotKey();
  loading.value = true;
  loadError.value = "";
  if (dataSnapshotKey.value !== requestedKey) {
    data.value = null;
    dataSection.value = null;
  }
  clearTimeout(timer);
  timer = window.setTimeout(() => {
    if (id === sequence) delayedLoading.value = true;
  }, 300);
  try {
    const next = await getBusinessAnalytics(requestedSection, analyticsParams());
    if (id === sequence) {
      data.value = next;
      dataSection.value = requestedSection;
      dataSnapshotKey.value = requestedKey;
    }
  } catch (e) {
    if (id === sequence) {
      data.value = null;
      dataSection.value = null;
      dataSnapshotKey.value = "";
      loadError.value = extractApiErrorMessage(
        e,
        t("admin.business.loadFailed"),
      );
      app.showError(loadError.value);
    }
  } finally {
    if (id === sequence) {
      loading.value = false;
      delayedLoading.value = false;
    }
    clearTimeout(timer);
  }
}
function syncAndLoad() {
  void syncURL();
  void load();
}
function csvEscape(value: unknown) {
  const text = String(value ?? "");
  return /[",\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text;
}
function exportCSV() {
  if (!data.value) return;
  let rows: unknown[][] = [];
  if (section.value === "retention") {
    rows = [
      [
        t(
          cohort.value === "registration"
            ? "admin.business.registrationDate"
            : "admin.business.firstUseDate",
        ),
        t(
          cohort.value === "registration"
            ? "admin.business.registeredUsers"
            : "admin.business.firstUseUsers",
        ),
        "D1",
        "D7",
        "D30",
      ],
      ...cohortRows.value.map((x) => [
        x.cohort_date,
        x.users,
        x.d1 ?? "",
        x.d7 ?? "",
        x.d30 ?? "",
      ]),
    ];
  } else if (section.value === "channels") {
    rows = [
      [
        t("admin.business.channel"),
        t("admin.business.newUsers"),
        t("admin.business.activatedUsers"),
        t("admin.business.firstPaidUsers"),
        t("admin.business.netPaid"),
        t("admin.business.arppu"),
      ],
      ...data.value.channels.map((x) => [
        channelLabel(x.channel),
        x.new_users,
        x.activated_users,
        x.first_paid_users,
        x.net_paid_cny,
        x.arppu_cny,
      ]),
    ];
  } else {
    rows = [
      [
        t("admin.business.timeBucket"),
        t("admin.business.newUsers"),
        t("admin.business.activeUsers"),
        t("admin.business.netPaid"),
        t("admin.business.consumedRevenue"),
        t("admin.business.supplierCost"),
      ],
      ...data.value.trend.map((x) => [
        x.bucket,
        x.new_users,
        x.active_users,
        x.net_paid_cny,
        x.consumed_revenue,
        x.supplier_cost,
      ]),
    ];
  }
  const blob = new Blob(
    ["\uFEFF" + rows.map((r) => r.map(csvEscape).join(",")).join("\n")],
    { type: "text/csv;charset=utf-8" },
  );
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = `business-${section.value}-${data.value.date_from}-${data.value.date_to}.csv`;
  link.click();
  URL.revokeObjectURL(link.href);
}
watch(
  () => props.section,
  () => {
    void load();
    if (props.section === "finance" && financeTab.value === "balance")
      void loadBalance();
  },
);
watch(
  range,
  () => {
    void router.replace({
      query: {
        ...route.query,
        date_from: range.value.date_from,
        date_to: range.value.date_to,
      },
    });
  },
  { deep: true },
);
watch(
  () => [
    route.query.date_from,
    route.query.date_to,
    route.query.channel,
    route.query.agent_id,
    route.query.agent_scope,
    route.query.tab,
    route.query.balance_window,
    route.query.cohort,
  ],
  () => void restoreStateFromRoute(),
);
onMounted(async () => {
  await restoreAgent();
  void load();
  if (section.value === "finance" && financeTab.value === "balance")
    void loadBalance();
});
onBeforeUnmount(() => clearTimeout(timer));
</script>
<style scoped>
.workspace {
  grid-template-columns: minmax(0, 1fr);
}
.workspace > * {
  min-width: 0;
}
.trend-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.analytics-toolbar {
  display: flex;
  min-width: 0;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--ui-border);
  padding-bottom: 12px;
}
.scope-filter {
  display: flex;
  min-width: 260px;
  flex: 1;
  align-items: flex-end;
  gap: 10px;
}
.agent-picker {
  width: min(400px, 100%);
}
.scope-controls {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}
.scope-badge {
  border: 1px solid var(--ui-border);
  border-radius: 6px;
  background: var(--ui-surface-raised);
  padding: 7px 10px;
  color: var(--ui-text);
  font-size: 12px;
  font-weight: 650;
}
.workspace {
  position: relative;
  margin: 0 auto;
  width: 100%;
  min-width: 0;
  max-width: 1440px;
  display: grid;
  gap: 16px;
}
.toolbar {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}
.channel {
  width: 150px;
  min-height: 36px;
  font-size: 12px;
}
.section-nav {
  display: flex;
  max-width: 100%;
  overflow-x: auto;
  border-bottom: 1px solid var(--ui-border);
  gap: 2px;
}
.section-nav a {
  flex: none;
  border-bottom: 2px solid transparent;
  padding: 9px 13px;
  color: var(--ui-text-muted);
  font-size: 13px;
}
.section-nav a.router-link-active {
  border-color: var(--ui-focus);
  color: var(--ui-text);
  font-weight: 650;
}
.kpis {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface);
}
.kpis article {
  min-width: 0;
  padding: 15px;
  border-right: 1px solid var(--ui-border);
}
.kpis article:last-child {
  border: 0;
}
.kpis span,
.kpis small {
  display: block;
  color: var(--ui-text-muted);
  font-size: 11px;
}
.kpis strong {
  display: block;
  margin: 8px 0 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 20px;
  font-variant-numeric: tabular-nums;
}
.kpis small.positive {
  color: rgb(var(--color-success-foreground));
}
.kpis small.negative {
  color: rgb(var(--color-danger-foreground));
}
.two-cols {
  display: grid;
  grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
  gap: 16px;
}
.panel {
  min-width: 0;
  max-width: 100%;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface);
  padding: 16px;
}
.panel-title {
  margin-bottom: 13px;
}
.panel-title h2 {
  font-size: 14px;
  font-weight: 650;
}
.panel-title p {
  margin-top: 3px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.with-tabs {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}
.segments {
  display: flex;
  gap: 2px;
}
.segments button {
  min-height: 32px;
  border-radius: 6px;
  padding: 0 10px;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.segments button[aria-pressed="true"] {
  border: 1px solid var(--ui-border);
  background: var(--ui-surface-raised);
  color: var(--ui-text);
  font-weight: 650;
}
.funnel {
  display: grid;
  gap: 10px;
}
.funnel div {
  position: relative;
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 10px;
  overflow: hidden;
  padding: 10px 12px;
}
.funnel i {
  position: absolute;
  inset: 0 auto 0 0;
  z-index: 0;
  background: color-mix(in srgb, var(--ui-focus) 9%, transparent);
}
.funnel span,
.funnel strong,
.funnel small {
  z-index: 1;
  font-size: 12px;
}
.funnel strong {
  font-size: 15px;
}
.funnel small {
  color: var(--ui-text-muted);
}
.detail-grid,
.finance-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}
.detail-grid dt,
.finance-grid dt {
  color: var(--ui-text-muted);
  font-size: 11px;
}
.detail-grid dd,
.finance-grid dd {
  margin-top: 5px;
  font-size: 17px;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}
.finance-grid small {
  display: block;
  margin-top: 4px;
  color: var(--ui-text-subtle);
  font-size: 10px;
}
.table-scroll {
  display: block;
  width: 100%;
  max-width: 100%;
  overflow-x: auto;
}
table {
  width: 100%;
  min-width: 720px;
  border-collapse: collapse;
  font-size: 12px;
}
th,
td {
  padding: 11px 12px;
  border-bottom: 1px solid var(--ui-border);
  text-align: right;
  font-variant-numeric: tabular-nums;
}
th:first-child,
td:first-child {
  text-align: left;
}
th {
  color: var(--ui-text-muted);
  font-weight: 600;
}
.clickable {
  cursor: pointer;
}
.clickable:hover {
  background: var(--ui-surface-subtle);
}
.heat {
  display: inline-block;
  min-width: 64px;
  border-radius: 5px;
  padding: 5px;
  text-align: center;
}
.data-footer {
  display: flex;
  min-height: 42px;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  color: var(--ui-text-muted);
  font-size: 11px;
}
.data-footer .btn {
  margin-left: auto;
}
.updating {
  position: absolute;
  z-index: 10;
  top: 48px;
  right: 0;
  border: 1px solid var(--ui-border);
  border-radius: 6px;
  background: var(--ui-surface-raised);
  padding: 6px 10px;
  font-size: 11px;
  box-shadow: var(--ui-shadow-sm);
}
.coverage-warning {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 8px;
  border: 1px solid color-mix(in srgb, var(--ui-warning) 32%, var(--ui-border));
  border-radius: 6px;
  background: color-mix(in srgb, var(--ui-warning) 7%, var(--ui-surface));
  padding: 8px 10px;
  color: var(--ui-text-muted);
  font-size: 12px;
  line-height: 1.5;
  overflow-wrap: anywhere;
}
.coverage-warning :deep(svg) {
  flex: none;
  margin-top: 1px;
  color: var(--ui-warning-strong);
}
.skeleton {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.skeleton span {
  height: 116px;
  border-radius: 8px;
  background: var(--ui-surface-subtle);
  animation: pulse 1.2s ease-in-out infinite;
}
.load-error {
  display: grid;
  min-height: 220px;
  place-items: center;
  align-content: center;
  gap: 8px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface);
  padding: 24px;
  color: var(--ui-text-muted);
  text-align: center;
}
.load-error strong {
  color: var(--ui-text);
  font-size: 14px;
}
.load-error p {
  max-width: 520px;
  font-size: 12px;
}
@keyframes pulse {
  50% {
    opacity: 0.5;
  }
}
@media (max-width: 1100px) {
  .kpis {
    grid-template-columns: repeat(3, 1fr);
  }
  .kpis article:nth-child(3) {
    border-right: 0;
  }
  .kpis article:nth-child(-n + 3) {
    border-bottom: 1px solid var(--ui-border);
  }
}
@media (max-width: 767px) {
  .toolbar {
    width: 100%;
  }
  .toolbar {
    flex-wrap: wrap;
  }
  .toolbar > :first-child {
    width: 100%;
  }
  .channel {
    min-height: 40px;
    flex: 1;
  }
  .kpis {
    grid-template-columns: repeat(2, 1fr);
  }
  .kpis article:nth-child(3) {
    border-right: 1px solid var(--ui-border);
  }
  .kpis article:nth-child(2n) {
    border-right: 0;
  }
  .two-cols {
    grid-template-columns: minmax(0, 1fr);
  }
  .detail-grid,
  .finance-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .with-tabs {
    flex-direction: column;
  }
  .segments {
    width: 100%;
  }
  .segments button {
    min-height: 40px;
    flex: 1;
  }
  .data-footer {
    flex-wrap: wrap;
    justify-content: flex-start;
  }
  .data-footer .btn {
    margin-left: 0;
  }
}
@media (prefers-reduced-motion: reduce) {
  .skeleton span {
    animation: none;
  }
}
.retention-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 16px;
}
.lifecycle dl {
  display: grid;
}
.lifecycle dl div {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 2px 10px;
  border-bottom: 1px solid var(--ui-border);
  padding: 10px 0;
}
.lifecycle dt {
  display: flex;
  align-items: center;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.lifecycle dd {
  font-size: 16px;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
}
.lifecycle small {
  grid-column: 1/-1;
  color: var(--ui-text-subtle);
  font-size: 10px;
}
.balance-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  color: var(--ui-text-muted);
  font-size: 11px;
}
.balance-note {
  margin: 14px 0;
  border-left: 3px solid var(--ui-border);
  padding: 8px 10px;
  background: var(--ui-surface-subtle);
  color: var(--ui-text-muted);
  font-size: 11px;
}
.balance-loading {
  display: grid;
  min-height: 160px;
  place-items: center;
  color: var(--ui-text-muted);
  font-size: 12px;
}
.balance-error {
  display: flex;
  min-height: 72px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border: 1px solid var(--ui-danger-border, var(--ui-border));
  border-radius: 7px;
  background: var(--ui-danger-bg, var(--ui-surface-subtle));
  color: var(--ui-text-muted);
  font-size: 12px;
}
.balance-kpis {
  padding-top: 4px;
}
.table-scroll td small {
  display: block;
  margin-top: 2px;
  color: var(--ui-text-subtle);
  font-size: 10px;
}
@media (max-width: 1100px) {
  .trend-grid {
    grid-template-columns: 1fr;
  }
}
@media (max-width: 767px) {
  .scope-filter,
  .scope-controls {
    align-items: stretch;
    flex-direction: column;
  }
  .agent-picker {
    width: 100%;
  }
  .scope-controls .segments {
    width: 100%;
  }
  .scope-controls small {
    line-height: 18px;
  }
  .retention-layout {
    grid-template-columns: 1fr;
  }
  .balance-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
  .balance-toolbar .segments {
    width: 100%;
  }
}
.duration-bars {
  display: grid;
  gap: 9px;
}
.duration-bars > div {
  display: grid;
  grid-template-columns: 64px minmax(60px, 1fr) 28px minmax(60px, 1fr) 28px;
  align-items: center;
  gap: 7px;
  font-size: 11px;
}
.duration-bars i {
  height: 7px;
  overflow: hidden;
  border-radius: 3px;
  background: var(--ui-surface-subtle);
}
.duration-bars b {
  display: block;
  height: 100%;
  background: var(--ui-focus);
}
.duration-bars b.paid {
  background: rgb(var(--color-success-foreground));
}
.duration-bars strong {
  text-align: right;
  font-variant-numeric: tabular-nums;
}
.duration-bars footer {
  display: flex;
  justify-content: flex-end;
  gap: 18px;
  color: var(--ui-text-subtle);
  font-size: 10px;
}
.duration-bars footer span:before {
  display: inline-block;
  width: 7px;
  height: 7px;
  margin-right: 5px;
  border-radius: 2px;
  background: var(--ui-focus);
  content: "";
}
.duration-bars footer span:last-child:before {
  background: rgb(var(--color-success-foreground));
}
@media (max-width: 767px) {
  .duration-bars > div {
    grid-template-columns: 58px minmax(40px, 1fr) 24px minmax(40px, 1fr) 24px;
  }
}
.kpis article > span {
  display: flex;
  min-width: 0;
  align-items: center;
  line-height: 18px;
}
.kpis article > small {
  display: block;
}
.workspace {
  gap: 14px;
}
.analytics-toolbar {
  flex-wrap: wrap;
  gap: 12px 16px;
}
.scope-filter {
  flex: 1 1 260px;
}
.toolbar {
  flex: 1 1 680px;
  flex-wrap: wrap;
}
.analytics-toolbar + .section-nav {
  margin-top: -2px;
}
@media (max-width: 1100px) {
  .analytics-toolbar {
    align-items: stretch;
    flex-direction: column;
  }
  .scope-filter,
  .toolbar {
    flex: none;
  }
  .toolbar {
    justify-content: flex-start;
  }
}
@media (max-width: 767px) {
  .analytics-toolbar {
    gap: 12px;
  }
  .scope-filter,
  .scope-controls {
    align-items: stretch;
    flex-direction: column;
  }
  .agent-picker,
  .toolbar {
    width: 100%;
  }
  .toolbar > :first-child {
    width: 100%;
  }
  .scope-controls .segments {
    width: 100%;
  }
}
</style>
