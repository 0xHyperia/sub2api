<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0">
          <h1 class="page-title">分销总览</h1>
          <p class="page-description">代理经营、佣金负债、提现队列与审计风险</p>
        </div>
        <button
          class="btn btn-secondary btn-icon shrink-0"
          title="刷新"
          aria-label="刷新"
          :disabled="loading"
          @click="load"
        >
          <Icon
            name="refresh"
            size="sm"
            :class="loading ? 'animate-spin' : ''"
          />
        </button>
      </header>
      <div
        v-if="loading && !overview"
        class="flex min-h-56 items-center justify-center"
      >
        <LoadingSpinner />
      </div>
      <template v-else-if="overview">
        <section class="grid grid-cols-2 gap-3 xl:grid-cols-4">
          <router-link
            to="/admin/distribution/agents"
            class="card p-4 transition-colors hover:border-outline-strong"
            ><div class="flex items-start justify-between gap-2">
              <p class="text-sm text-foreground-subtle">代理规模</p>
              <Icon name="users" size="sm" class="text-brand" />
            </div>
            <p class="mt-2 text-2xl font-semibold tabular-nums">
              {{ overview.total_agents }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">
              有效 {{ overview.active_agents }} · 本月新增
              {{ overview.new_agents_this_month }}
            </p></router-link
          >
          <router-link
            to="/admin/distribution/customers"
            class="card p-4 transition-colors hover:border-outline-strong"
            ><div class="flex items-start justify-between gap-2">
              <p class="text-sm text-foreground-subtle">代理客户</p>
              <Icon name="user" size="sm" class="text-brand" />
            </div>
            <p class="mt-2 text-2xl font-semibold tabular-nums">
              {{ overview.total_customers }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">
              付费 {{ overview.paying_customers }} · 本月新增
              {{ overview.new_customers_this_month }}
            </p></router-link
          >
          <router-link
            to="/admin/distribution/commissions"
            class="card p-4 transition-colors hover:border-outline-strong"
            ><div class="flex items-start justify-between gap-2">
              <p class="text-sm text-foreground-subtle">累计佣金</p>
              <Icon
                name="trendingUp"
                size="sm"
                class="text-success-foreground"
              />
            </div>
            <p class="mt-2 text-2xl font-semibold tabular-nums">
              {{ money(overview.total_commission_cny) }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">
              本月 {{ money(overview.month_commission_cny) }}
            </p></router-link
          >
          <router-link
            to="/admin/distribution/withdrawals"
            class="card p-4 transition-colors hover:border-outline-strong"
            ><div class="flex items-start justify-between gap-2">
              <p class="text-sm text-foreground-subtle">提现待办</p>
              <Icon
                name="creditCard"
                size="sm"
                class="text-warning-foreground"
              />
            </div>
            <p class="mt-2 text-2xl font-semibold tabular-nums">
              {{ overview.pending_withdrawals + overview.paying_withdrawals }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">
              待付
              {{
                money(
                  Number(overview.pending_withdrawal_cny) +
                    Number(overview.paying_withdrawal_cny),
                )
              }}
            </p></router-link
          >
        </section>

        <section
          v-if="maturity"
          class="flex flex-wrap items-center justify-between gap-3 border-y border-outline py-3 text-sm"
        >
          <div class="flex min-w-0 items-center gap-3">
            <Icon name="clock" size="sm" class="shrink-0 text-brand" />
            <div class="min-w-0">
              <p class="font-medium">佣金自动解冻任务</p>
              <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                {{ maturityDescription }}
              </p>
            </div>
          </div>
          <span class="badge shrink-0" :class="maturityStatusClass">{{
            maturityStatusText
          }}</span>
        </section>

        <section
          class="grid gap-5 lg:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.75fr)]"
        >
          <div class="border-y border-outline py-5">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="font-semibold">资金与佣金负债</h2>
                <p class="mt-1 text-sm text-foreground-subtle">
                  平台对全部代理的当前应付构成
                </p>
              </div>
              <router-link
                to="/admin/distribution/commissions"
                class="text-sm font-medium text-brand"
                >查看台账</router-link
              >
            </div>
            <dl class="mt-5 grid grid-cols-2 gap-x-4 gap-y-5 sm:grid-cols-4">
              <div>
                <dt class="text-xs text-foreground-subtle">客户累计实付</dt>
                <dd class="mt-1 font-semibold tabular-nums">
                  {{ money(overview.customer_paid_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">可结算佣金</dt>
                <dd
                  class="mt-1 font-semibold tabular-nums text-success-foreground"
                >
                  {{ money(overview.available_commission_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">冻结佣金</dt>
                <dd class="mt-1 font-semibold tabular-nums">
                  {{ money(overview.frozen_commission_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">提现占用</dt>
                <dd class="mt-1 font-semibold tabular-nums">
                  {{ money(overview.reserved_commission_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">累计冲正</dt>
                <dd
                  class="mt-1 font-semibold tabular-nums"
                  :class="
                    Number(overview.reversed_commission_cny) > 0
                      ? 'text-danger-foreground'
                      : ''
                  "
                >
                  {{ money(overview.reversed_commission_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">负债欠款</dt>
                <dd
                  class="mt-1 font-semibold tabular-nums"
                  :class="
                    Number(overview.debt_commission_cny) > 0
                      ? 'text-danger-foreground'
                      : ''
                  "
                >
                  {{ money(overview.debt_commission_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">本月已付款</dt>
                <dd class="mt-1 font-semibold tabular-nums">
                  {{ money(overview.paid_this_month_cny) }}
                </dd>
              </div>
              <div>
                <dt class="text-xs text-foreground-subtle">暂停代理</dt>
                <dd class="mt-1 font-semibold tabular-nums">
                  {{ overview.suspended_agents }}
                </dd>
              </div>
            </dl>
          </div>

          <div class="border-y border-outline py-5">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="font-semibold">运营与审计待办</h2>
                <p class="mt-1 text-sm text-foreground-subtle">
                  优先处理可能影响结算时效的事项
                </p>
              </div>
              <router-link
                to="/admin/distribution/anomalies"
                class="text-sm font-medium text-brand"
                >异常对账</router-link
              >
            </div>
            <div class="mt-4 divide-y divide-outline">
              <router-link
                to="/admin/distribution/withdrawals?status=pending"
                class="flex items-center justify-between gap-3 py-3"
                ><div>
                  <p class="text-sm font-medium">待审核提现</p>
                  <p class="mt-0.5 text-xs text-foreground-subtle">
                    合计 {{ money(overview.pending_withdrawal_cny) }}
                  </p>
                </div>
                <span
                  class="badge"
                  :class="
                    overview.pending_withdrawals
                      ? 'badge-warning'
                      : 'badge-gray'
                  "
                  >{{ overview.pending_withdrawals }}</span
                ></router-link
              >
              <router-link
                to="/admin/distribution/withdrawals?status=paying"
                class="flex items-center justify-between gap-3 py-3"
                ><div>
                  <p class="text-sm font-medium">打款处理中</p>
                  <p class="mt-0.5 text-xs text-foreground-subtle">
                    合计 {{ money(overview.paying_withdrawal_cny) }}
                  </p>
                </div>
                <span
                  class="badge"
                  :class="
                    overview.paying_withdrawals ? 'badge-primary' : 'badge-gray'
                  "
                  >{{ overview.paying_withdrawals }}</span
                ></router-link
              >
              <router-link
                to="/admin/distribution/anomalies?type=overdue_withdrawal"
                class="flex items-center justify-between gap-3 py-3"
                ><div>
                  <p class="text-sm font-medium">审核超时超过 24 小时</p>
                  <p class="mt-0.5 text-xs text-foreground-subtle">
                    需确认处理人和阻塞原因
                  </p>
                </div>
                <span
                  class="badge"
                  :class="
                    overview.overdue_pending_count
                      ? 'badge-danger'
                      : 'badge-gray'
                  "
                  >{{ overview.overdue_pending_count }}</span
                ></router-link
              >
            </div>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import AppLayout from "@/components/layout/AppLayout.vue";
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import Icon from "@/components/icons/Icon.vue";
import {
  getMaturityStatus,
  getOverview,
  type DistributionAdminOverview,
  type DistributionMaturityStatus,
} from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";

const app = useAppStore();
const overview = ref<DistributionAdminOverview | null>(null);
const maturity = ref<DistributionMaturityStatus | null>(null);
const loading = ref(false);
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`;
const maturityStatusText = computed(() => {
  if (maturity.value?.running) return "运行中";
  const labels: Record<string, string> = {
    success: "正常",
    standby: "待机",
    error: "异常",
    disabled: "未启用",
  };
  return labels[maturity.value?.last_outcome || ""] || "待运行";
});
const maturityStatusClass = computed(() =>
  maturity.value?.last_outcome === "error"
    ? "badge-danger"
    : maturity.value?.running
      ? "badge-primary"
      : maturity.value?.last_outcome === "success"
        ? "badge-success"
        : "badge-gray",
);
const maturityDescription = computed(() => {
  if (!maturity.value) return "-";
  if (maturity.value.last_error)
    return `最近执行失败：${maturity.value.last_error}`;
  if (!maturity.value.last_success_at) return "等待首次执行";
  return `最近成功 ${new Date(maturity.value.last_success_at).toLocaleString()} · 本次释放 ${maturity.value.last_released} 条佣金`;
});
async function load() {
  loading.value = true;
  try {
    [overview.value, maturity.value] = await Promise.all([
      getOverview(),
      getMaturityStatus(),
    ]);
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载分销总览失败"));
  } finally {
    loading.value = false;
  }
}
onMounted(load);
</script>
