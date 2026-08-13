<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">佣金台账</h1>
              <p class="page-description">
                逐笔核对订单实付、计佣基数、比例、冲正与入账状态
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button
                class="btn btn-secondary h-10 px-3"
                title="导出当前筛选"
                aria-label="导出佣金台账"
                :disabled="exporting"
                @click="downloadExport"
              >
                <Icon name="download" size="sm" /><span class="hidden sm:inline"
                  >导出</span
                ></button
              ><button
                class="btn btn-secondary btn-icon"
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
            </div>
          </header>
          <AdminDistributionNav class="mt-5" />
        </template>

        <template #filters>
          <div class="space-y-3">
            <div
              class="grid grid-cols-[minmax(0,1fr)_112px] gap-2 lg:grid-cols-[minmax(280px,1fr)_140px_140px_140px_150px_150px]"
            >
              <div class="relative min-w-0">
                <Icon
                  name="search"
                  size="sm"
                  class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle"
                />
                <input
                  v-model="search"
                  class="input pl-9"
                  type="search"
                  placeholder="搜索订单、客户或代理"
                  aria-label="搜索佣金记录"
                />
              </div>
              <button
                type="button"
                class="btn btn-secondary h-10 px-3 lg:hidden"
                aria-label="筛选佣金记录"
                @click="filterDialog = true"
              >
                <Icon name="filter" size="sm" /><span
                  v-if="activeFilterCount"
                  class="badge badge-primary"
                  >{{ activeFilterCount }}</span
                >
              </button>
              <select
                v-model="entryType"
                class="input hidden lg:block"
                aria-label="佣金类型"
              >
                <option value="">全部类型</option>
                <option value="direct">直属佣金</option>
                <option value="team">团队差额</option>
              </select>
              <select
                v-model="status"
                class="input hidden lg:block"
                aria-label="入账状态"
              >
                <option value="">全部状态</option>
                <option value="frozen">冻结中</option>
                <option value="available">已到账</option>
                <option value="reversed">已冲正</option>
              </select>
              <select
                v-model="paymentType"
                class="input hidden lg:block"
                aria-label="支付方式"
              >
                <option value="">全部支付方式</option>
                <option value="alipay">支付宝</option>
                <option value="wxpay">微信支付</option>
                <option value="stripe">Stripe</option>
                <option value="airwallex">Airwallex</option>
                <option value="demo">演示数据</option>
              </select>
              <input
                v-model="dateFrom"
                class="input hidden lg:block"
                type="date"
                aria-label="开始日期"
              />
              <input
                v-model="dateTo"
                class="input hidden lg:block"
                type="date"
                aria-label="结束日期"
              />
            </div>
            <div
              class="flex items-center justify-between gap-3 text-xs text-foreground-subtle"
            >
              <p>
                共 {{ pagination.total }} 条记录，当前页净佣金
                {{ money(pageCommission) }}
              </p>
              <button
                v-if="hasFilters"
                type="button"
                class="text-brand hover:underline"
                @click="clearFilters"
              >
                清除筛选
              </button>
            </div>
            <DistributionFilterSummary :items="filterSummary" @remove="removeFilter" @clear="clearFilters" />
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
            :server-side-sort="true"
            :clickable-rows="true"
            default-sort-key="created_at"
            default-sort-order="desc"
            sort-storage-key="admin-distribution-commissions"
            @sort="changeSort"
            @rowClick="openDetail"
          >
            <template #cell-customer="{ row }">
              <div class="min-w-44 max-w-56">
                <p class="truncate font-medium">
                  {{ row.customer_email || "未设置邮箱" }}
                </p>
                <p class="truncate text-xs text-foreground-subtle">
                  {{ row.customer_username || `用户 #${row.customer_user_id}` }}
                </p>
              </div>
            </template>
            <template #cell-agent="{ row }">
              <div class="min-w-40 max-w-52">
                <p class="truncate text-sm">
                  {{ row.agent_email || "未知代理" }}
                </p>
                <p class="truncate text-xs text-foreground-subtle">
                  {{ row.entry_type === "team" ? "团队差额" : "直属佣金" }}
                </p>
              </div>
            </template>
            <template #cell-order="{ row }">
              <div class="min-w-32">
                <p class="font-mono text-xs">
                  {{ row.order_no || `#${row.payment_order_id}` }}
                </p>
                <p class="mt-1 text-xs text-foreground-subtle">
                  {{ paymentLabel(row.payment_type) }}
                </p>
              </div>
            </template>
            <template #cell-commission_base="{ row }"
              ><span class="tabular-nums">{{
                money(row.commission_base_cny)
              }}</span></template
            >
            <template #cell-rate="{ row }"
              ><span class="tabular-nums">{{
                percent(row.rate_bps)
              }}</span></template
            >
            <template #cell-commission="{ row }">
              <div class="text-right">
                <p class="font-semibold tabular-nums text-success-foreground">
                  {{ money(netCommission(row)) }}
                </p>
                <p
                  v-if="Number(row.reversed_amount_cny) > 0"
                  class="text-xs text-danger-foreground"
                >
                  冲正 {{ money(row.reversed_amount_cny) }}
                </p>
              </div>
            </template>
            <template #cell-status="{ row }"
              ><span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span></template
            >
            <template #cell-created_at="{ row }"
              ><time class="whitespace-nowrap text-sm text-foreground-muted">{{
                dateTime(row.created_at)
              }}</time></template
            >
            <template #mobile-card="{ row }">
              <div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium">
                      {{ row.customer_email || "未设置邮箱" }}
                    </p>
                    <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                      {{ row.order_no || `订单 #${row.payment_order_id}` }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="statusClass(row.status)"
                    >{{ statusText(row.status) }}</span
                  >
                </div>
                <div
                  class="grid grid-cols-3 gap-2 rounded-control bg-surface-subtle px-3 py-2.5 text-xs"
                >
                  <div>
                    <p class="text-foreground-subtle">计佣基数</p>
                    <p class="mt-1 font-semibold">
                      {{ money(row.commission_base_cny) }}
                    </p>
                  </div>
                  <div class="text-center">
                    <p class="text-foreground-subtle">比例</p>
                    <p class="mt-1 font-semibold">
                      {{ percent(row.rate_bps) }}
                    </p>
                  </div>
                  <div class="text-right">
                    <p class="text-foreground-subtle">净佣金</p>
                    <p class="mt-1 font-semibold text-success-foreground">
                      {{ money(netCommission(row)) }}
                    </p>
                  </div>
                </div>
                <div
                  class="flex items-center justify-between gap-3 text-xs text-foreground-muted"
                >
                  <span class="truncate"
                    >{{ row.agent_email || "未知代理" }} ·
                    {{ row.entry_type === "team" ? "团队差额" : "直属" }}</span
                  ><time class="shrink-0">{{ dateTime(row.created_at) }}</time>
                </div>
              </div>
            </template>
            <template #empty
              ><div class="py-12 text-center">
                <p class="font-medium">暂无符合条件的佣金记录</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  调整筛选条件后重试
                </p>
              </div></template
            >
          </DataTable>
        </template>

        <template #pagination>
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="changePage"
            @update:pageSize="changePageSize"
          />
        </template>
      </TablePageLayout>
    </div>

    <BaseDrawer
      :show="detailDialog"
      title="佣金记录详情"
      description="订单、计佣和入账信息"
      @close="detailDialog = false"
    >
      <div v-if="selected" class="space-y-5">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="font-mono text-sm font-medium">
              {{ selected.order_no || `订单 #${selected.payment_order_id}` }}
            </p>
            <p class="mt-1 text-xs text-foreground-subtle">
              {{ dateTime(selected.created_at) }}
            </p>
          </div>
          <span class="badge shrink-0" :class="statusClass(selected.status)">{{
            statusText(selected.status)
          }}</span>
        </div>
        <dl
          class="grid grid-cols-2 gap-x-4 gap-y-4 border-y border-outline py-4 text-sm"
        >
          <div>
            <dt class="text-foreground-subtle">客户</dt>
            <dd class="mt-1 break-all font-medium">
              {{ selected.customer_email }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">受益代理</dt>
            <dd class="mt-1 break-all font-medium">
              {{ selected.agent_email }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">实付</dt>
            <dd class="mt-1 font-medium">
              {{ selected.payment_currency }}
              {{ Number(selected.actual_paid_amount).toFixed(2) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">折算汇率</dt>
            <dd class="mt-1 font-medium">
              {{ Number(selected.fx_rate_to_cny).toFixed(4) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">计佣基数</dt>
            <dd class="mt-1 font-medium">
              {{ money(selected.commission_base_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">返佣比例</dt>
            <dd class="mt-1 font-medium">{{ percent(selected.rate_bps) }}</dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">原始佣金</dt>
            <dd class="mt-1 font-medium">
              {{ money(selected.original_amount_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">冲正金额</dt>
            <dd
              class="mt-1 font-medium"
              :class="
                Number(selected.reversed_amount_cny) > 0
                  ? 'text-danger-foreground'
                  : ''
              "
            >
              {{ money(selected.reversed_amount_cny) }}
            </dd>
          </div>
        </dl>
        <p class="text-xs text-foreground-subtle">
          可用时间：{{ dateTime(selected.available_at) }} · 支付方式：{{
            paymentLabel(selected.payment_type)
          }}
        </p>
      </div>
    </BaseDrawer>
    <BaseDialog
      :show="filterDialog"
      title="筛选佣金记录"
      width="narrow"
      @close="filterDialog = false"
      ><div class="space-y-4">
        <div>
          <label class="input-label">佣金类型</label
          ><select v-model="entryType" class="input">
            <option value="">全部类型</option>
            <option value="direct">直属佣金</option>
            <option value="team">团队差额</option>
          </select>
        </div>
        <div>
          <label class="input-label">到账状态</label
          ><select v-model="status" class="input">
            <option value="">全部状态</option>
            <option value="frozen">冻结中</option>
            <option value="available">已到账</option>
            <option value="reversed">已冲正</option>
          </select>
        </div>
        <div>
          <label class="input-label">支付方式</label
          ><select v-model="paymentType" class="input">
            <option value="">全部支付方式</option>
            <option value="alipay">支付宝</option>
            <option value="wxpay">微信支付</option>
            <option value="stripe">Stripe</option>
            <option value="airwallex">Airwallex</option>
            <option value="demo">演示数据</option>
          </select>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="input-label">开始日期</label
            ><input v-model="dateFrom" class="input" type="date" />
          </div>
          <div>
            <label class="input-label">结束日期</label
            ><input v-model="dateTo" class="input" type="date" />
          </div>
        </div>
      </div>
      <template #footer
        ><div class="flex w-full items-center justify-between">
          <button class="btn btn-ghost" @click="clearFilters">清除</button
          ><button class="btn btn-primary" @click="filterDialog = false">
            完成
          </button>
        </div></template
      ></BaseDialog
    >
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import BaseDrawer from "@/components/common/BaseDrawer.vue";
import AdminDistributionNav from "@/components/admin/distribution/AdminDistributionNav.vue";
import Icon from "@/components/icons/Icon.vue";
import DistributionFilterSummary, { type DistributionFilterItem } from "@/components/distribution/DistributionFilterSummary.vue";
import type { Column } from "@/components/common/types";
import type { DistributionCommission } from "@/api/distribution";
import { exportCommissions, listCommissions } from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { saveDistributionExport } from "@/utils/distributionExport";

const app = useAppStore();
const route = useRoute();
const items = ref<DistributionCommission[]>([]);
const loading = ref(false);
const exporting = ref(false);
const search = ref(String(route.query.search || ""));
const status = ref("");
const entryType = ref("");
const paymentType = ref("");
const dateFrom = ref("");
const dateTo = ref("");
const sortBy = ref("created_at");
const sortOrder = ref<"asc" | "desc">("desc");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
const detailDialog = ref(false);
const filterDialog = ref(false);
const selected = ref<DistributionCommission | null>(null);
let searchTimer: number | null = null;

const columns: Column[] = [
  { key: "customer", label: "客户", sortable: true },
  { key: "agent", label: "受益代理", sortable: true },
  { key: "order", label: "来源订单" },
  { key: "commission_base", label: "计佣基数", sortable: true },
  { key: "rate", label: "比例", sortable: true },
  { key: "commission", label: "净佣金", sortable: true },
  { key: "status", label: "状态", sortable: true },
  { key: "created_at", label: "产生时间", sortable: true },
];

const hasFilters = computed(() =>
  Boolean(
    search.value ||
    status.value ||
    entryType.value ||
    paymentType.value ||
    dateFrom.value ||
    dateTo.value,
  ),
);
const filterSummary = computed<DistributionFilterItem[]>(() => [
  search.value ? { key: 'search', label: '搜索', value: search.value } : null,
  status.value ? { key: 'status', label: '状态', value: statusText(status.value) } : null,
  entryType.value ? { key: 'type', label: '来源', value: entryType.value === 'direct' ? '直属佣金' : '团队差额' } : null,
  paymentType.value ? { key: 'payment', label: '支付', value: paymentLabel(paymentType.value) } : null,
  dateFrom.value ? { key: 'from', label: '开始', value: dateFrom.value } : null,
  dateTo.value ? { key: 'to', label: '结束', value: dateTo.value } : null,
].filter((item): item is DistributionFilterItem => Boolean(item)));
function removeFilter(key: string) {
  if (key === 'search') search.value = '';
  if (key === 'status') status.value = '';
  if (key === 'type') entryType.value = '';
  if (key === 'payment') paymentType.value = '';
  if (key === 'from') dateFrom.value = '';
  if (key === 'to') dateTo.value = '';
}
const activeFilterCount = computed(
  () =>
    [
      status.value,
      entryType.value,
      paymentType.value,
      dateFrom.value,
      dateTo.value,
    ].filter(Boolean).length,
);
const pageCommission = computed(() =>
  items.value.reduce((sum, item) => sum + netCommission(item), 0),
);
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`;
const percent = (bps: number) => `${(Number(bps || 0) / 100).toFixed(2)}%`;
const netCommission = (item: DistributionCommission) =>
  Number(item.original_amount_cny) - Number(item.reversed_amount_cny);
const dateTime = (value?: string) =>
  value ? new Date(value).toLocaleString() : "-";
const paymentLabel = (value: string) =>
  ({
    alipay: "支付宝",
    wxpay: "微信支付",
    stripe: "Stripe",
    airwallex: "Airwallex",
    demo: "演示数据",
    cash: "现金支付",
  })[value] ||
  value ||
  "-";
const statusText = (value: string) =>
  value === "frozen" ? "冻结中" : value === "reversed" ? "已冲正" : "已到账";
const statusClass = (value: string) =>
  value === "frozen"
    ? "badge-warning"
    : value === "reversed"
      ? "badge-danger"
      : "badge-success";

function toStart(value: string) {
  return value ? new Date(`${value}T00:00:00`).toISOString() : undefined;
}
function toEnd(value: string) {
  return value ? new Date(`${value}T23:59:59.999`).toISOString() : undefined;
}

async function load() {
  loading.value = true;
  try {
    const result = await listCommissions({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      entry_type: entryType.value || undefined,
      payment_type: paymentType.value || undefined,
      date_from: toStart(dateFrom.value),
      date_to: toEnd(dateTo.value),
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    });
    items.value = result.items;
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    };
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载佣金台账失败"));
  } finally {
    loading.value = false;
  }
}

async function downloadExport() {
  exporting.value = true;
  try {
    const blob = await exportCommissions({
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      entry_type: entryType.value || undefined,
      payment_type: paymentType.value || undefined,
      date_from: toStart(dateFrom.value),
      date_to: toEnd(dateTo.value),
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    });
    saveDistributionExport(blob, "distribution-commissions");
    app.showSuccess("佣金台账已导出");
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "导出佣金台账失败"));
  } finally {
    exporting.value = false;
  }
}

function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) window.clearTimeout(searchTimer);
  searchTimer = window.setTimeout(load, 300);
}
function changeSort(key: string, order: "asc" | "desc") {
  sortBy.value = key;
  sortOrder.value = order;
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
function clearFilters() {
  search.value = "";
  status.value = "";
  entryType.value = "";
  paymentType.value = "";
  dateFrom.value = "";
  dateTo.value = "";
  pagination.value.page = 1;
  void load();
}
function openDetail(row: DistributionCommission) {
  selected.value = row;
  detailDialog.value = true;
}

watch(search, scheduleLoad);
watch([status, entryType, paymentType, dateFrom, dateTo], () => {
  pagination.value.page = 1;
  void load();
});
onMounted(load);
</script>
