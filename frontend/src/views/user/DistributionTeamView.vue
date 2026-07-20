<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">团队管理</h1>
              <p class="page-description">管理下级代理及返佣策略</p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button
                v-if="overview?.agent.can_recruit_subagents"
                class="btn btn-primary h-10 px-3 sm:px-4"
                :disabled="overview?.agent.status !== 'active'"
                aria-label="添加下级代理"
                @click="openAddDialog"
              >
                <Icon name="userPlus" size="sm" /><span class="hidden sm:inline"
                  >添加下级代理</span
                >
              </button>
              <span v-else class="badge badge-gray whitespace-nowrap"
                >招募权限未开通</span
              >
              <button
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
          <DistributionNav class="mt-5" />
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
                placeholder="搜索邮箱、用户名或推广码"
                aria-label="搜索下级代理"
              />
            </div>
            <select v-model="status" class="input" aria-label="下级代理状态">
              <option value="">全部状态</option>
              <option value="active">有效</option>
              <option value="suspended">已暂停</option>
              <option value="revoked">已撤销</option>
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
          >
            <template #cell-agent="{ row }"
              ><div class="min-w-44 max-w-56">
                <p class="truncate font-medium">
                  {{ row.email || "未设置邮箱" }}
                </p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                  {{ row.username || "未设置用户名" }}
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
            <template #cell-business="{ row }"
              ><div>
                <p class="font-medium">{{ row.customer_count || 0 }} 位客户</p>
                <p class="mt-0.5 text-xs text-foreground-subtle">
                  累计 {{ money(row.total_earned_cny) }}
                </p>
              </div></template
            >
            <template #cell-status="{ row }"
              ><span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span></template
            >
            <template #cell-actions="{ row }"
              ><button
                v-if="row.status !== 'revoked'"
                class="btn btn-ghost btn-icon btn-sm"
                :title="row.status === 'active' ? '暂停代理' : '恢复代理'"
                :aria-label="row.status === 'active' ? '暂停代理' : '恢复代理'"
                @click="openStatusConfirm(row)"
              >
                <Icon
                  :name="row.status === 'active' ? 'ban' : 'play'"
                  size="sm"
                /></button
            ></template>
            <template #mobile-card="{ row }"
              ><div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium">
                      {{ row.email || "未设置邮箱" }}
                    </p>
                    <p class="truncate text-xs text-foreground-subtle">
                      {{ row.username || "未设置用户名" }}
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
                    <p class="text-xs text-foreground-subtle">返佣比例</p>
                    <p class="mt-1 font-semibold">
                      {{ rate(row.effective_rate_bps) }}
                    </p>
                  </div>
                  <div class="min-w-0 text-right">
                    <p class="text-xs text-foreground-subtle">推广码</p>
                    <p class="mt-1 max-w-40 truncate font-mono text-xs">
                      {{ row.promotion_code }}
                    </p>
                  </div>
                </div>
                <dl class="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <dt class="text-foreground-subtle">直属客户</dt>
                    <dd class="mt-1 font-semibold">
                      {{ row.customer_count || 0 }}
                    </dd>
                  </div>
                  <div class="text-right">
                    <dt class="text-foreground-subtle">累计佣金</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(row.total_earned_cny) }}
                    </dd>
                  </div>
                </dl>
                <button
                  v-if="row.status !== 'revoked'"
                  class="btn btn-secondary w-full"
                  @click="openStatusConfirm(row)"
                >
                  <Icon
                    :name="row.status === 'active' ? 'ban' : 'play'"
                    size="sm"
                  />{{ row.status === "active" ? "暂停代理" : "恢复代理" }}
                </button>
              </div></template
            >
            <template #empty
              ><div class="py-12 text-center">
                <p class="font-medium">暂无下级代理</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  使用完整邮箱添加符合条件的现有用户
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
      title="添加下级代理"
      width="normal"
      @close="closeAddDialog"
    >
      <form
        id="add-team-agent-form"
        class="space-y-5"
        @submit.prevent="submitAgent"
      >
        <div>
          <label for="team-agent-email" class="input-label">用户邮箱</label>
          <input
            id="team-agent-email"
            v-model.trim="childEmail"
            class="input"
            type="email"
            autocomplete="off"
            required
            placeholder="输入完整邮箱"
          />
          <p class="input-hint">仅支持完整邮箱精确匹配</p>
        </div>
        <div>
          <label for="team-agent-rate" class="input-label">返佣比例</label>
          <div class="relative">
            <input
              id="team-agent-rate"
              v-model="ratePercent"
              class="input pr-10"
              type="number"
              min="0"
              :max="maxRatePercent"
              step="0.01"
              placeholder="使用系统默认比例"
            /><span
              class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-foreground-muted"
              >%</span
            >
          </div>
          <p class="input-hint">
            最高可设置 {{ maxRatePercent.toFixed(2) }}%，留空使用系统默认比例
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
            />高级设置
          </button>
          <div v-if="showAdvanced" class="mt-4">
            <label for="team-agent-code" class="input-label">自定义推广码</label
            ><input
              id="team-agent-code"
              v-model="promotionCode"
              class="input font-mono uppercase"
              maxlength="32"
              placeholder="留空由系统自动生成"
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
            取消</button
          ><button
            class="btn btn-primary"
            type="submit"
            form="add-team-agent-form"
            :disabled="creating || !childEmail"
          >
            <Icon name="userPlus" size="sm" />{{
              creating ? "添加中..." : "确认添加"
            }}
          </button>
        </div></template
      >
    </BaseDialog>

    <ConfirmDialog
      :show="statusConfirm"
      :title="
        pendingAgent?.status === 'active' ? '暂停下级代理' : '恢复下级代理'
      "
      :message="statusConfirmMessage"
      :confirm-text="
        pendingAgent?.status === 'active' ? '确认暂停' : '确认恢复'
      "
      :danger="pendingAgent?.status === 'active'"
      @confirm="confirmStatus"
      @cancel="statusConfirm = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import AppLayout from "@/components/layout/AppLayout.vue";
import TablePageLayout from "@/components/layout/TablePageLayout.vue";
import DataTable from "@/components/common/DataTable.vue";
import Pagination from "@/components/common/Pagination.vue";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import Icon from "@/components/icons/Icon.vue";
import DistributionNav from "@/components/distribution/DistributionNav.vue";
import type { Column } from "@/components/common/types";
import {
  getDistributionOverview,
  grantL2Agent,
  listTeam,
  updateTeamAgentStatus,
  type DistributionAgent,
  type DistributionOverview,
} from "@/api/distribution";
import { useDistributionAccess } from "@/composables/useDistributionAccess";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";

const app = useAppStore();
const router = useRouter();
const { loadDistributionAccess } = useDistributionAccess();
const overview = ref<DistributionOverview | null>(null);
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
const statusConfirm = ref(false);
const pendingAgent = ref<DistributionAgent | null>(null);
const columns: Column[] = [
  { key: "agent", label: "下级代理" },
  { key: "policy", label: "推广策略" },
  { key: "business", label: "业务贡献" },
  { key: "status", label: "状态" },
  { key: "actions", label: "操作" },
];
const maxRatePercent = computed(
  () => (overview.value?.agent.max_child_rate_bps || 0) / 100,
);
const statusConfirmMessage = computed(() =>
  pendingAgent.value?.status === "active"
    ? "暂停后，该代理不能绑定新客户或产生新佣金，已有账目不会丢失。"
    : "恢复后，该代理可以继续使用原推广链接并产生佣金。",
);
const money = (value: string) => `¥${Number(value || 0).toFixed(2)}`;
const rate = (bps: number) => `${(bps / 100).toFixed(2)}%`;
const statusText = (value: string) =>
  ({ active: "有效", suspended: "已暂停", revoked: "已撤销" })[value] || value;
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
      getDistributionOverview(),
      listTeam({
        page: pagination.value.page,
        page_size: pagination.value.page_size,
        search: search.value.trim() || undefined,
        status: status.value || undefined,
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
    app.showError(extractApiErrorMessage(error, "加载团队失败"));
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
function openAddDialog() {
  addDialog.value = true;
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
  const parsed =
    ratePercent.value.trim() === "" ? undefined : Number(ratePercent.value);
  if (
    parsed !== undefined &&
    (!Number.isFinite(parsed) || parsed < 0 || parsed > maxRatePercent.value)
  ) {
    app.showError(
      `返佣比例必须在 0% 到 ${maxRatePercent.value.toFixed(2)}% 之间`,
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
    app.showSuccess("下级代理已添加");
    closeAddDialog();
    pagination.value.page = 1;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "添加下级代理失败"));
  } finally {
    creating.value = false;
  }
}
function openStatusConfirm(agent: DistributionAgent) {
  pendingAgent.value = agent;
  statusConfirm.value = true;
}
async function confirmStatus() {
  if (!pendingAgent.value) return;
  const next = pendingAgent.value.status === "active" ? "suspended" : "active";
  try {
    await updateTeamAgentStatus(pendingAgent.value.id, next);
    app.showSuccess(next === "active" ? "下级代理已恢复" : "下级代理已暂停");
    statusConfirm.value = false;
    pendingAgent.value = null;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "状态更新失败"));
  }
}
watch([search, status], scheduleLoad);
onMounted(load);
</script>
