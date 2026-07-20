<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">异常对账</h1>
              <p class="page-description">
                集中处理佣金、提现、代理归属和资金负债异常
              </p>
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
        </template>

        <template #filters>
          <div class="space-y-3">
            <div
              class="grid grid-cols-[minmax(0,1fr)_112px] gap-2 lg:grid-cols-[minmax(280px,1fr)_180px_140px]"
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
                  placeholder="搜索申请编号、订单、代理或客户"
                  aria-label="搜索分销异常"
                />
              </div>
              <button
                type="button"
                class="btn btn-secondary h-10 px-3 lg:hidden"
                aria-label="筛选分销异常"
                @click="filterDialog = true"
              >
                <Icon name="filter" size="sm" /><span
                  v-if="activeFilterCount"
                  class="badge badge-primary"
                  >{{ activeFilterCount }}</span
                >
              </button>
              <select
                v-model="type"
                class="input hidden lg:block"
                aria-label="异常类型"
              >
                <option value="">全部类型</option>
                <option value="overdue_withdrawal">提现审核超时</option>
                <option value="pending_fx">汇率待处理</option>
                <option value="agent_debt">代理钱包负债</option>
                <option value="inactive_agent_customers">非活跃代理客户</option>
                <option value="matured_commission">佣金解冻延迟</option>
              </select>
              <select
                v-model="severity"
                class="input hidden lg:block"
                aria-label="风险等级"
              >
                <option value="">全部等级</option>
                <option value="critical">严重</option>
                <option value="high">高风险</option>
                <option value="medium">需关注</option>
              </select>
            </div>
            <div
              class="flex items-center justify-between gap-3 text-xs text-foreground-subtle"
            >
              <p>共 {{ pagination.total }} 项待处理异常</p>
              <button
                v-if="hasFilters"
                type="button"
                class="text-brand hover:underline"
                @click="clearFilters"
              >
                清除筛选
              </button>
            </div>
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
            :clickable-rows="true"
            @rowClick="openTarget"
          >
            <template #cell-severity="{ row }"
              ><span class="badge" :class="severityClass(row.severity)">{{
                severityText(row.severity)
              }}</span></template
            >
            <template #cell-problem="{ row }"
              ><div class="min-w-52 max-w-80">
                <p class="font-medium">{{ typeText(row.type) }}</p>
                <p class="mt-1 text-xs text-foreground-subtle">
                  {{ row.description }}
                </p>
              </div></template
            >
            <template #cell-entity="{ row }"
              ><div class="min-w-44 max-w-64">
                <p class="truncate font-mono text-xs font-medium">
                  {{ row.reference }}
                </p>
                <p class="mt-1 truncate text-xs text-foreground-subtle">
                  {{ row.subject }}
                </p>
                <p
                  v-if="row.agent_email && row.agent_email !== row.subject"
                  class="truncate text-xs text-foreground-muted"
                >
                  代理 {{ row.agent_email }}
                </p>
              </div></template
            >
            <template #cell-amount="{ row }"
              ><span
                v-if="Number(row.amount_cny)"
                class="whitespace-nowrap font-semibold tabular-nums"
                >{{ money(row.amount_cny) }}</span
              ><span v-else class="text-foreground-subtle">-</span></template
            >
            <template #cell-detected_at="{ row }"
              ><time class="whitespace-nowrap text-sm text-foreground-muted">{{
                dateTime(row.detected_at)
              }}</time></template
            >
            <template #cell-actions="{ row }"
              ><router-link
                class="btn btn-ghost btn-icon btn-sm"
                :to="targetFor(row)"
                title="前往处理"
                aria-label="前往处理"
                @click.stop
                ><Icon name="arrowRight" size="sm" /></router-link
            ></template>
            <template #mobile-card="{ row }"
              ><div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="font-medium">{{ typeText(row.type) }}</p>
                    <p
                      class="mt-1 truncate font-mono text-xs text-foreground-subtle"
                    >
                      {{ row.reference }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="severityClass(row.severity)"
                    >{{ severityText(row.severity) }}</span
                  >
                </div>
                <p class="text-sm text-foreground-subtle">
                  {{ row.description }}
                </p>
                <div
                  class="flex items-end justify-between gap-3 border-t border-outline pt-3"
                >
                  <div class="min-w-0">
                    <p class="truncate text-sm">{{ row.subject }}</p>
                    <p class="mt-0.5 text-xs text-foreground-muted">
                      {{ dateTime(row.detected_at) }}
                    </p>
                  </div>
                  <router-link
                    class="btn btn-secondary btn-sm shrink-0"
                    :to="targetFor(row)"
                    @click.stop
                    >处理<Icon name="arrowRight" size="sm"
                  /></router-link>
                </div></div
            ></template>
            <template #empty
              ><div class="py-12 text-center">
                <Icon
                  name="checkCircle"
                  size="lg"
                  class="mx-auto text-success-foreground"
                />
                <p class="mt-3 font-medium">当前没有符合条件的异常</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  资金和业务队列暂未发现需要人工处理的问题
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
      :show="filterDialog"
      title="筛选异常"
      width="narrow"
      @close="filterDialog = false"
    >
      <div class="space-y-4">
        <div>
          <label for="anomaly-type" class="input-label">异常类型</label
          ><select id="anomaly-type" v-model="type" class="input">
            <option value="">全部类型</option>
            <option value="overdue_withdrawal">提现审核超时</option>
            <option value="pending_fx">汇率待处理</option>
            <option value="agent_debt">代理钱包负债</option>
            <option value="inactive_agent_customers">非活跃代理客户</option>
            <option value="matured_commission">佣金解冻延迟</option>
          </select>
        </div>
        <div>
          <label for="anomaly-severity" class="input-label">风险等级</label
          ><select id="anomaly-severity" v-model="severity" class="input">
            <option value="">全部等级</option>
            <option value="critical">严重</option>
            <option value="high">高风险</option>
            <option value="medium">需关注</option>
          </select>
        </div>
      </div>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button class="btn btn-secondary" type="button" @click="clearFilters">
            清除</button
          ><button
            class="btn btn-primary"
            type="button"
            @click="filterDialog = false"
          >
            完成
          </button>
        </div></template
      >
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter, type RouteLocationRaw } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import type { Column } from "@/components/common/types";
import {
  listAnomalies,
  type DistributionAdminAnomaly,
} from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";

const app = useAppStore();
const router = useRouter();
const route = useRoute();
const items = ref<DistributionAdminAnomaly[]>([]);
const loading = ref(false);
const search = ref(String(route.query.search || ""));
const requestedType = String(route.query.type || "");
const type = ref(
  [
    "overdue_withdrawal",
    "pending_fx",
    "agent_debt",
    "inactive_agent_customers",
    "matured_commission",
  ].includes(requestedType)
    ? requestedType
    : "",
);
const severity = ref("");
const filterDialog = ref(false);
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
let searchTimer: number | null = null;
const columns: Column[] = [
  { key: "severity", label: "等级" },
  { key: "problem", label: "异常问题" },
  { key: "entity", label: "业务对象" },
  { key: "amount", label: "涉及金额" },
  { key: "detected_at", label: "发现时间" },
  { key: "actions", label: "操作" },
];
const activeFilterCount = computed(
  () => Number(Boolean(type.value)) + Number(Boolean(severity.value)),
);
const hasFilters = computed(() =>
  Boolean(search.value || type.value || severity.value),
);
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`;
const dateTime = (value: string) => new Date(value).toLocaleString();
const severityText = (value: string) =>
  ({ critical: "严重", high: "高风险", medium: "需关注" })[value] || value;
const severityClass = (value: string) =>
  ({
    critical: "badge-danger",
    high: "badge-warning",
    medium: "badge-primary",
  })[value] || "badge-gray";
const typeText = (value: string) =>
  ({
    overdue_withdrawal: "提现审核超时",
    pending_fx: "汇率待处理",
    agent_debt: "代理钱包负债",
    inactive_agent_customers: "非活跃代理客户",
    matured_commission: "佣金解冻延迟",
  })[value] || value;

function targetFor(item: DistributionAdminAnomaly): RouteLocationRaw {
  if (item.entity_type === "withdrawal")
    return {
      path: "/admin/distribution/withdrawals",
      query: { search: item.reference },
    };
  if (item.entity_type === "agent")
    return {
      path: "/admin/distribution/agents",
      query: { search: item.agent_email || item.reference },
    };
  if (item.entity_type === "customer")
    return {
      path: "/admin/distribution/customers",
      query: { search: item.subject },
    };
  return {
    path: "/admin/distribution/commissions",
    query: { search: item.reference },
  };
}
function openTarget(item: DistributionAdminAnomaly) {
  void router.push(targetFor(item));
}
async function load() {
  loading.value = true;
  try {
    const result = await listAnomalies({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      type: type.value || undefined,
      severity: severity.value || undefined,
      sort_order: "desc",
    });
    items.value = result.items;
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    };
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载分销异常失败"));
  } finally {
    loading.value = false;
  }
}
function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(load, 300);
}
function clearFilters() {
  search.value = "";
  type.value = "";
  severity.value = "";
  filterDialog.value = false;
  pagination.value.page = 1;
  void load();
}
function changePage(page: number) {
  pagination.value.page = page;
  void load();
}
function changePageSize(size: number) {
  pagination.value.page = 1;
  pagination.value.page_size = size;
  void load();
}
watch(search, scheduleLoad);
watch([type, severity], () => {
  pagination.value.page = 1;
  void load();
});
onMounted(load);
</script>
