<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">佣金明细</h1>
              <p class="page-description">直属与团队差额佣金账单</p>
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
          <DistributionNav class="mt-5" />
        </template>
        <template #filters>
          <div
            class="grid grid-cols-[minmax(0,1fr)_92px_92px] gap-2 sm:max-w-3xl sm:grid-cols-[minmax(280px,1fr)_140px_140px]"
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
                placeholder="搜索客户邮箱或用户名"
                aria-label="搜索佣金"
              />
            </div>
            <select v-model="entryType" class="input" aria-label="佣金来源">
              <option value="">全部来源</option>
              <option value="direct">直属</option>
              <option value="team">团队差额</option>
            </select>
            <select v-model="status" class="input" aria-label="佣金到账状态">
              <option value="">全部状态</option>
              <option value="frozen">冻结中</option>
              <option value="available">已到账</option>
              <option value="reversed">已冲正</option>
            </select>
          </div>
          <DistributionFilterSummary class="mt-2" :items="filterSummary" @remove="removeFilter" @clear="clearFilters" />
        </template>
        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
          >
            <template #cell-customer="{ row }"
              ><div class="min-w-44 max-w-56">
                <p class="truncate font-medium">
                  {{ row.customer_email || "未设置邮箱" }}
                </p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                  {{ row.customer_username || "未设置用户名" }}
                </p>
              </div></template
            >
            <template #cell-source="{ row }"
              ><div>
                <span
                  class="badge"
                  :class="
                    row.entry_type === 'team' ? 'badge-gray' : 'badge-primary'
                  "
                  >{{ row.entry_type === "team" ? "团队差额" : "直属" }}</span
                >
                <p class="mt-1 text-xs text-foreground-subtle">
                  {{ paymentText(row.payment_type) }}
                </p>
              </div></template
            >
            <template #cell-calculation="{ row }"
              ><div class="min-w-40">
                <p class="font-semibold tabular-nums text-success-foreground">
                  +{{ money(net(row)) }}
                </p>
                <p class="mt-0.5 text-xs text-foreground-subtle">
                  基数 {{ money(row.commission_base_cny) }} ×
                  {{ rate(row.rate_bps) }}
                </p>
              </div></template
            >
            <template #cell-status="{ row }"
              ><div class="min-w-32">
                <span class="badge" :class="statusClass(row.status)">{{
                  statusText(row.status)
                }}</span>
                <p class="mt-1 text-xs text-foreground-subtle">
                  {{ statusTime(row) }}
                </p>
              </div></template
            >
            <template #mobile-card="{ row }"
              ><div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium">
                      {{ row.customer_email || "未设置邮箱" }}
                    </p>
                    <p class="truncate text-xs text-foreground-subtle">
                      {{ row.customer_username || "未设置用户名" }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="statusClass(row.status)"
                    >{{ statusText(row.status) }}</span
                  >
                </div>
                <div
                  class="flex items-center justify-between gap-3 rounded-control bg-surface-subtle px-3 py-2"
                >
                  <div>
                    <p class="text-xs text-foreground-subtle">
                      {{ row.entry_type === "team" ? "团队差额" : "直属佣金" }}
                    </p>
                    <p class="mt-1 text-xs">
                      {{ money(row.commission_base_cny) }} ×
                      {{ rate(row.rate_bps) }}
                    </p>
                  </div>
                  <p class="font-semibold text-success-foreground">
                    +{{ money(net(row)) }}
                  </p>
                </div>
                <p class="text-xs text-foreground-muted">
                  {{ statusTime(row) }} · {{ paymentText(row.payment_type) }}
                </p>
              </div></template
            >
            <template #empty
              ><div class="py-12 text-center">
                <p class="font-medium">暂无佣金记录</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  客户完成符合条件的实付后，佣金会显示在这里
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
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import Icon from "@/components/icons/Icon.vue";
import DistributionNav from "@/components/distribution/DistributionNav.vue";
import {
  listCommissions,
  type DistributionCommission,
} from "@/api/distribution";
import type { Column } from "@/components/common/types";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import DistributionFilterSummary, { type DistributionFilterItem } from "@/components/distribution/DistributionFilterSummary.vue";

const app = useAppStore();
const items = ref<DistributionCommission[]>([]);
const loading = ref(false);
const search = ref("");
const status = ref("");
const filterSummary = computed<DistributionFilterItem[]>(() => [
  search.value ? { key: 'search', label: '搜索', value: search.value } : null,
  entryType.value ? { key: 'type', label: '来源', value: entryType.value === 'direct' ? '直属' : '团队差额' } : null,
  status.value ? { key: 'status', label: '状态', value: statusText(status.value) } : null,
].filter((item): item is DistributionFilterItem => Boolean(item)));
function removeFilter(key: string) { if (key === 'search') search.value = ''; if (key === 'type') entryType.value = ''; if (key === 'status') status.value = ''; }
function clearFilters() { search.value = ''; entryType.value = ''; status.value = ''; }
const entryType = ref("");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
let searchTimer: number | null = null;
const columns: Column[] = [
  { key: "customer", label: "客户" },
  { key: "source", label: "来源" },
  { key: "calculation", label: "佣金计算" },
  { key: "status", label: "状态与时间" },
];
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`;
const rate = (bps: number) => `${(bps / 100).toFixed(2)}%`;
const net = (item: DistributionCommission) =>
  Math.max(
    0,
    Number(item.original_amount_cny) - Number(item.reversed_amount_cny),
  );
const statusText = (value: string) =>
  value === "frozen" ? "冻结中" : value === "reversed" ? "已冲正" : "已到账";
const statusClass = (value: string) =>
  value === "frozen"
    ? "badge-warning"
    : value === "reversed"
      ? "badge-danger"
      : "badge-success";
const paymentText = (value: string) =>
  ({
    balance: "余额支付",
    cash: "在线支付",
    alipay: "支付宝",
    wechat: "微信支付",
  })[value] ||
  value ||
  "实付订单";
const statusTime = (item: DistributionCommission) =>
  item.status === "frozen"
    ? `预计 ${new Date(item.available_at).toLocaleString()} 到账`
    : item.status === "reversed"
      ? new Date(item.created_at).toLocaleString()
      : `${new Date(item.available_at).toLocaleString()} 到账`;
async function load() {
  loading.value = true;
  try {
    const result = await listCommissions({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      entry_type: entryType.value || undefined,
    });
    items.value = result.items;
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    };
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载佣金失败"));
  } finally {
    loading.value = false;
  }
}
function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = window.setTimeout(load, 300);
}
function changePage(page: number) {
  pagination.value.page = page;
  void load();
}
function changePageSize(pageSize: number) {
  pagination.value.page = 1;
  pagination.value.page_size = pageSize;
  void load();
}
watch([search, status, entryType], scheduleLoad);
onMounted(load);
</script>
