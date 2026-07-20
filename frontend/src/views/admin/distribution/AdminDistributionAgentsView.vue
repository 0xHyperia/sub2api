<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">代理管理</h1>
              <p class="page-description">管理一级代理及其团队关系</p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button
                type="button"
                class="btn btn-primary h-10 px-3 sm:px-4"
                aria-label="添加一级代理"
                @click="openAddDialog"
              >
                <Icon name="userPlus" size="sm" />
                <span class="hidden sm:inline">添加一级代理</span>
              </button>
              <button
                type="button"
                class="btn btn-secondary h-10 px-3"
                title="导出当前筛选"
                aria-label="导出代理"
                :disabled="exporting"
                @click="downloadExport"
              >
                <Icon name="download" size="sm" /><span class="hidden lg:inline"
                  >导出</span
                >
              </button>
              <button
                type="button"
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
        </template>

        <template #filters>
          <div
            class="grid grid-cols-[minmax(0,1fr)_96px_96px] gap-2 sm:max-w-3xl sm:grid-cols-[minmax(280px,1fr)_140px_140px]"
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
                placeholder="搜索邮箱、用户名或推广码"
                aria-label="搜索代理"
              />
            </div>
            <select v-model="depth" class="input" aria-label="代理等级">
              <option value="">全部等级</option>
              <option value="1">一级代理</option>
              <option value="2">二级代理</option>
            </select>
            <select v-model="status" class="input" aria-label="代理状态">
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
            :server-side-sort="true"
            default-sort-key="created_at"
            default-sort-order="desc"
            sort-storage-key="admin-distribution-agents"
            @sort="changeSort"
          >
            <template #cell-email="{ row }">
              <div class="w-36 max-w-36">
                <p class="truncate font-medium text-foreground">
                  {{ row.email || "未设置邮箱" }}
                </p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                  {{ row.username || "未设置用户名" }}
                </p>
              </div>
            </template>
            <template #cell-relationship="{ row }">
              <div class="w-28 max-w-28">
                <span
                  class="badge"
                  :class="row.depth === 1 ? 'badge-primary' : 'badge-gray'"
                >
                  {{ row.depth === 1 ? "一级代理" : "二级代理" }}
                </span>
                <p class="mt-1 truncate text-xs text-foreground-subtle">
                  {{
                    row.depth === 1
                      ? "平台直属"
                      : `上级：${row.parent_email || "未知"}`
                  }}
                </p>
              </div>
            </template>
            <template #cell-effective_rate="{ row }">
              <div class="w-24 text-right">
                <p class="font-medium tabular-nums">
                  {{ formatPercent(row.effective_rate_bps) }}
                </p>
                <p class="mt-0.5 text-xs text-foreground-subtle">
                  {{
                    row.rate_override_bps == null ? "系统默认" : "自定义比例"
                  }}
                </p>
              </div>
            </template>
            <template #cell-customer_paid="{ row }"
              ><div class="text-right">
                <p class="font-medium tabular-nums">
                  {{ money(row.customer_paid_cny) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  {{ row.paying_customer_count || 0 }} 位付费客户
                </p>
              </div></template
            >
            <template #cell-total_earned="{ row }"
              ><div class="text-right">
                <p class="font-semibold tabular-nums text-success-foreground">
                  {{ money(row.total_earned_cny) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  本月 {{ money(row.this_month_commission_cny) }}
                </p>
              </div></template
            >
            <template #cell-available="{ row }"
              ><div class="text-right">
                <p class="font-medium tabular-nums">
                  {{ money(row.available_cny) }}
                </p>
                <p class="text-xs text-foreground-subtle">
                  冻结 {{ money(row.frozen_cny) }}
                </p>
              </div></template
            >
            <template #cell-status="{ row }">
              <span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span>
            </template>
            <template #cell-actions="{ row }">
              <button
                type="button"
                class="btn btn-ghost btn-icon btn-sm"
                title="管理代理"
                aria-label="管理代理"
                @click.stop="openManageDialog(row)"
              >
                <Icon name="more" size="sm" />
              </button>
            </template>

            <template #mobile-card="{ row }">
              <div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-medium text-foreground">
                      {{ row.email || "未设置邮箱" }}
                    </p>
                    <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                      {{ row.username || "未设置用户名" }}
                    </p>
                  </div>
                  <span
                    class="badge shrink-0"
                    :class="statusClass(row.status)"
                    >{{ statusText(row.status) }}</span
                  >
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    class="badge"
                    :class="row.depth === 1 ? 'badge-primary' : 'badge-gray'"
                    >{{ row.depth === 1 ? "一级代理" : "二级代理" }}</span
                  >
                  <span class="font-mono text-xs text-foreground-muted">{{
                    row.promotion_code
                  }}</span>
                </div>
                <dl
                  class="grid grid-cols-2 gap-3 border-t border-outline pt-3 text-xs"
                >
                  <div>
                    <dt class="text-foreground-subtle">返佣比例</dt>
                    <dd class="mt-1 font-semibold">
                      {{ formatPercent(row.effective_rate_bps) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">客户实付</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(row.customer_paid_cny) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">累计佣金</dt>
                    <dd class="mt-1 font-semibold text-success-foreground">
                      {{ money(row.total_earned_cny) }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">客户 / 付费</dt>
                    <dd class="mt-1 font-semibold">
                      {{ row.customer_count || 0 }} /
                      {{ row.paying_customer_count || 0 }}
                    </dd>
                  </div>
                  <div>
                    <dt class="text-foreground-subtle">可用 / 冻结</dt>
                    <dd class="mt-1 font-semibold">
                      {{ money(row.available_cny) }} /
                      {{ money(row.frozen_cny) }}
                    </dd>
                  </div>
                </dl>
                <button
                  type="button"
                  class="btn btn-secondary w-full"
                  @click="openManageDialog(row)"
                >
                  <Icon name="more" size="sm" />管理代理
                </button>
              </div>
            </template>

            <template #empty>
              <div class="py-10 text-center">
                <p class="font-medium text-foreground">暂无符合条件的代理</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  调整筛选条件，或添加第一个一级代理
                </p>
              </div>
            </template>
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

    <BaseDialog
      :show="addDialog"
      title="添加一级代理"
      width="normal"
      @close="closeAddDialog"
    >
      <form
        id="add-distribution-agent-form"
        class="space-y-5"
        @submit.prevent="submitAgent"
      >
        <RemoteEntityCombobox
          v-model="selectedUser"
          input-id="distribution-agent-user"
          label="选择用户"
          placeholder="输入用户邮箱或用户名"
          :search="searchAgentCandidates"
        />

        <div>
          <label for="distribution-agent-rate" class="input-label"
            >返佣比例</label
          >
          <div class="relative">
            <input
              id="distribution-agent-rate"
              v-model="ratePercent"
              type="number"
              class="input pr-9"
              min="0"
              max="100"
              step="0.01"
              :placeholder="defaultRateLabel"
            />
            <span class="input-suffix">%</span>
          </div>
          <p class="input-hint">留空使用系统默认比例 {{ defaultRateLabel }}%</p>
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
            />
            高级设置
          </button>
          <div v-if="showAdvanced" class="mt-4">
            <label for="distribution-agent-code" class="input-label"
              >自定义推广码</label
            >
            <input
              id="distribution-agent-code"
              v-model="promotionCode"
              type="text"
              class="input font-mono uppercase"
              maxlength="32"
              placeholder="留空由系统自动生成"
            />
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            @click="closeAddDialog"
          >
            取消
          </button>
          <button
            type="submit"
            form="add-distribution-agent-form"
            class="btn btn-primary"
            :disabled="submitting || !selectedUser"
          >
            <Icon name="userPlus" size="sm" />{{
              submitting ? "添加中..." : "确认添加"
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="manageDialog"
      title="代理详情"
      width="normal"
      @close="closeManageDialog"
    >
      <div v-if="managedAgent" class="space-y-5">
        <div class="min-w-0">
          <p class="truncate font-medium text-foreground">
            {{ managedAgent.email || "未设置邮箱" }}
          </p>
          <p class="mt-1 truncate text-sm text-foreground-subtle">
            {{ managedAgent.username || "未设置用户名" }}
          </p>
        </div>
        <dl class="grid grid-cols-2 gap-4 border-y border-outline py-4 text-sm">
          <div>
            <dt class="text-foreground-subtle">代理等级</dt>
            <dd class="mt-1 font-medium">
              {{ managedAgent.depth === 1 ? "一级代理" : "二级代理" }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">当前状态</dt>
            <dd class="mt-1">
              <span class="badge" :class="statusClass(managedAgent.status)">{{
                statusText(managedAgent.status)
              }}</span>
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">推广码</dt>
            <dd class="mt-1 font-mono">{{ managedAgent.promotion_code }}</dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">返佣比例</dt>
            <dd class="mt-1 font-medium">
              {{ formatPercent(managedAgent.effective_rate_bps) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">直属客户</dt>
            <dd class="mt-1 font-medium">
              {{ managedAgent.customer_count || 0 }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">累计佣金</dt>
            <dd class="mt-1 font-medium">
              ¥{{ Number(managedAgent.total_earned_cny).toFixed(2) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">客户实付</dt>
            <dd class="mt-1 font-medium">
              {{ money(managedAgent.customer_paid_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">本月佣金</dt>
            <dd class="mt-1 font-medium">
              {{ money(managedAgent.this_month_commission_cny) }}
            </dd>
          </div>
          <div v-if="managedAgent.depth === 1">
            <dt class="text-foreground-subtle">下级招募</dt>
            <dd class="mt-1">
              <span
                class="badge"
                :class="
                  managedAgent.can_recruit_subagents
                    ? 'badge-success'
                    : 'badge-gray'
                "
              >
                {{ managedAgent.can_recruit_subagents ? "已授权" : "未授权" }}
              </span>
            </dd>
          </div>
        </dl>
        <section>
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold">变更记录</h3>
            <span class="text-xs text-foreground-subtle"
              >最近 {{ agentEvents.length }} 条</span
            >
          </div>
          <div v-if="eventsLoading" class="mt-3 flex justify-center py-4">
            <span class="text-sm text-foreground-subtle">加载中...</span>
          </div>
          <ol
            v-else-if="agentEvents.length"
            class="mt-3 space-y-3 border-l border-outline pl-4"
          >
            <li v-for="event in agentEvents" :key="event.id" class="relative">
              <span
                class="absolute -left-[21px] top-1.5 h-2.5 w-2.5 rounded-full border-2 border-surface bg-brand"
              ></span>
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-sm font-medium">{{ agentEventTitle(event) }}</p>
                <time class="text-xs text-foreground-subtle">{{
                  new Date(event.created_at).toLocaleString()
                }}</time>
              </div>
              <p class="mt-1 text-xs text-foreground-subtle">
                {{ event.actor_email || "系统" }} ·
                {{ event.reason || "无备注" }}
              </p>
            </li>
          </ol>
          <p v-else class="mt-3 text-sm text-foreground-subtle">暂无变更记录</p>
        </section>
        <div class="flex flex-wrap justify-end gap-2">
          <button
            v-if="managedAgent.depth === 1"
            type="button"
            class="btn btn-secondary"
            @click="openRecruitmentPermissionDialog"
          >
            <Icon name="users" size="sm" />{{
              managedAgent.can_recruit_subagents
                ? "撤销招募权限"
                : "授予招募权限"
            }}
          </button>
          <button
            type="button"
            class="btn btn-secondary"
            @click="openRateDialog"
          >
            <Icon name="edit" size="sm" />编辑返佣比例
          </button>
          <button
            v-if="managedAgent.status === 'active'"
            type="button"
            class="btn btn-secondary"
            @click="openStatusDialog('suspended')"
          >
            <Icon name="ban" size="sm" />暂停代理
          </button>
          <button
            v-else
            type="button"
            class="btn btn-primary"
            @click="openStatusDialog('active')"
          >
            <Icon name="check" size="sm" />{{
              managedAgent.status === "revoked" ? "重新启用" : "恢复代理"
            }}
          </button>
          <button
            v-if="managedAgent.status !== 'revoked'"
            type="button"
            class="btn btn-danger"
            @click="openStatusDialog('revoked')"
          >
            <Icon name="trash" size="sm" />撤销代理
          </button>
        </div>
      </div>
    </BaseDialog>

    <BaseDialog
      :show="recruitmentPermissionDialog"
      :title="
        managedAgent?.can_recruit_subagents ? '撤销招募权限' : '授予招募权限'
      "
      width="narrow"
      @close="recruitmentPermissionDialog = false"
    >
      <div class="space-y-4">
        <p class="text-sm text-foreground-subtle">
          {{
            managedAgent?.can_recruit_subagents
              ? "撤销后不能继续添加下级代理，已有团队关系不受影响。"
              : "授权后，该一级代理可以通过完整邮箱添加下级代理。"
          }}
        </p>
        <div>
          <label for="distribution-recruitment-reason" class="input-label"
            >操作原因</label
          >
          <textarea
            id="distribution-recruitment-reason"
            v-model="recruitmentPermissionReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            placeholder="记录授权或撤销依据"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            class="btn btn-secondary"
            @click="recruitmentPermissionDialog = false"
          >
            取消
          </button>
          <button
            class="btn btn-primary"
            :disabled="
              recruitmentPermissionSaving || !recruitmentPermissionReason.trim()
            "
            @click="saveRecruitmentPermission"
          >
            {{ recruitmentPermissionSaving ? "保存中..." : "确认" }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="rateDialog"
      title="编辑代理返佣比例"
      width="narrow"
      @close="rateDialog = false"
    >
      <div v-if="managedAgent" class="space-y-4">
        <div>
          <p class="text-sm font-medium">{{ managedAgent.email }}</p>
          <p class="mt-1 text-xs text-foreground-subtle">
            当前生效 {{ formatPercent(managedAgent.effective_rate_bps) }} ·
            {{
              managedAgent.rate_override_bps == null
                ? "沿用系统默认"
                : "自定义比例"
            }}
          </p>
        </div>
        <div
          class="inline-flex w-full rounded-control border border-outline bg-surface-subtle p-1"
          role="group"
          aria-label="返佣策略"
        >
          <button
            type="button"
            class="btn flex-1"
            :class="rateUseDefault ? 'btn-primary' : 'btn-ghost'"
            @click="rateUseDefault = true"
          >
            系统默认</button
          ><button
            type="button"
            class="btn flex-1"
            :class="!rateUseDefault ? 'btn-primary' : 'btn-ghost'"
            @click="rateUseDefault = false"
          >
            自定义比例
          </button>
        </div>
        <div v-if="!rateUseDefault">
          <label for="distribution-edit-rate" class="input-label"
            >返佣比例</label
          >
          <div class="relative">
            <input
              id="distribution-edit-rate"
              v-model="editRatePercent"
              class="input pr-9"
              type="number"
              min="0"
              max="100"
              step="0.01"
            /><span class="input-suffix">%</span>
          </div>
          <p class="input-hint">系统默认值 {{ managedDefaultRate }}%</p>
        </div>
        <div>
          <label for="distribution-rate-reason" class="input-label"
            >变更原因</label
          ><textarea
            id="distribution-rate-reason"
            v-model="rateReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            placeholder="说明本次比例调整依据"
          ></textarea>
          <p class="input-hint">变更前后比例和操作人将写入审计记录</p>
        </div>
      </div>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="rateDialog = false">
            取消</button
          ><button
            class="btn btn-primary"
            :disabled="rateSaving || !canSaveRate"
            @click="saveRate"
          >
            {{ rateSaving ? "保存中..." : "保存比例" }}
          </button>
        </div></template
      >
    </BaseDialog>

    <BaseDialog
      :show="statusDialog"
      :title="statusDialogTitle"
      width="narrow"
      @close="statusDialog = false"
    >
      <div class="space-y-4">
        <p class="text-sm text-foreground-subtle">{{ statusDialogMessage }}</p>
        <div>
          <label for="distribution-agent-status-reason" class="input-label">
            操作原因
          </label>
          <textarea
            id="distribution-agent-status-reason"
            v-model="statusReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            placeholder="记录本次状态调整原因"
          ></textarea>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            type="button"
            class="btn btn-secondary"
            @click="statusDialog = false"
          >
            取消
          </button>
          <button
            type="button"
            class="btn"
            :class="pendingStatus === 'revoked' ? 'btn-danger' : 'btn-primary'"
            :disabled="
              statusSaving || (statusReasonRequired && !statusReason.trim())
            "
            @click="confirmStatusChange"
          >
            {{ statusSaving ? "处理中..." : "确认" }}
          </button>
        </div>
      </template>
    </BaseDialog>
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
import Icon from "@/components/icons/Icon.vue";
import RemoteEntityCombobox from "@/components/admin/distribution/RemoteEntityCombobox.vue";
import type { DistributionPickerOption } from "@/components/admin/distribution/types";
import type { Column } from "@/components/common/types";
import type { DistributionAgent } from "@/api/distribution";
import {
  getSettings,
  grantAgent,
  exportAgents,
  listAgentEvents,
  listAgents,
  lookupAgentCandidates,
  updateAgentRate,
  updateAgentRecruitmentPermission,
  updateAgentStatus,
  type DistributionAgentEvent,
  type DistributionSettings,
} from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { extractApiErrorMessage } from "@/utils/apiError";
import { saveDistributionExport } from "@/utils/distributionExport";

const app = useAppStore();
const route = useRoute();
const items = ref<DistributionAgent[]>([]);
const loading = ref(false);
const exporting = ref(false);
const search = ref(String(route.query.search || ""));
const status = ref("");
const depth = ref("");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
let searchTimer: number | null = null;

const addDialog = ref(false);
const selectedUser = ref<DistributionPickerOption | null>(null);
const ratePercent = ref("");
const promotionCode = ref("");
const showAdvanced = ref(false);
const submitting = ref(false);
const settings = ref<DistributionSettings | null>(null);

const manageDialog = ref(false);
const managedAgent = ref<DistributionAgent | null>(null);
const statusDialog = ref(false);
const pendingStatus = ref<"active" | "suspended" | "revoked">("suspended");
const statusReason = ref("");
const statusSaving = ref(false);
const agentEvents = ref<DistributionAgentEvent[]>([]);
const eventsLoading = ref(false);
const sortBy = ref("created_at");
const sortOrder = ref<"asc" | "desc">("desc");
const rateDialog = ref(false);
const rateUseDefault = ref(true);
const editRatePercent = ref("");
const rateReason = ref("");
const rateSaving = ref(false);
const recruitmentPermissionDialog = ref(false);
const recruitmentPermissionReason = ref("");
const recruitmentPermissionSaving = ref(false);

const columns: Column[] = [
  { key: "email", label: "代理用户", sortable: true },
  { key: "relationship", label: "层级关系" },
  {
    key: "effective_rate",
    label: "返佣策略",
    sortable: true,
    class: "text-right",
  },
  {
    key: "customer_paid",
    label: "客户实付",
    sortable: true,
    class: "text-right",
  },
  {
    key: "total_earned",
    label: "佣金收入",
    sortable: true,
    class: "text-right",
  },
  {
    key: "available",
    label: "佣金余额",
    sortable: true,
    class: "text-right",
  },
  {
    key: "status",
    label: "状态",
    sortable: true,
    class: "text-center",
  },
  { key: "actions", label: "操作", class: "w-16 text-center" },
];

const defaultRateLabel = computed(() =>
  ((settings.value?.l1_default_rate_bps ?? 0) / 100).toFixed(2),
);
const isReactivatingRevoked = computed(
  () =>
    pendingStatus.value === "active" &&
    managedAgent.value?.status === "revoked",
);
const statusReasonRequired = computed(() => true);
const statusDialogTitle = computed(() =>
  isReactivatingRevoked.value
    ? "重新启用代理"
    : { active: "恢复代理", suspended: "暂停代理", revoked: "撤销代理" }[
        pendingStatus.value
      ],
);
const statusDialogMessage = computed(() =>
  pendingStatus.value === "revoked"
    ? "撤销后该代理不会再产生新的客户和佣金关系，管理员之后仍可重新启用。"
    : pendingStatus.value === "suspended"
      ? "暂停后将停止该代理的新客户绑定和新佣金产生。"
      : isReactivatingRevoked.value
        ? "重新启用后，该代理可以继续使用原推广链接并产生佣金。请填写重新启用原因。"
        : "恢复后该代理可以继续使用推广链接并产生佣金。",
);
const managedDefaultRate = computed(() => {
  if (!managedAgent.value || !settings.value) return "0.00";
  return (
    (managedAgent.value.depth === 1
      ? settings.value.l1_default_rate_bps
      : settings.value.l2_default_rate_bps) / 100
  ).toFixed(2);
});
const canSaveRate = computed(() => {
  if (!rateReason.value.trim()) return false;
  if (rateUseDefault.value) return true;
  const value = Number(editRatePercent.value);
  return Number.isFinite(value) && value >= 0 && value <= 100;
});

function statusText(value: string) {
  return (
    { active: "有效", suspended: "已暂停", revoked: "已撤销" }[value] || value
  );
}

function statusClass(value: string) {
  return value === "active"
    ? "badge-success"
    : value === "suspended"
      ? "badge-warning"
      : "badge-gray";
}

function formatPercent(bps: number) {
  return `${(bps / 100).toFixed(2)}%`;
}

function money(value: string | number) {
  return `¥${Number(value || 0).toFixed(2)}`;
}

async function load() {
  loading.value = true;
  try {
    const result = await listAgents({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      depth: depth.value ? Number(depth.value) : undefined,
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
    app.showError(extractApiErrorMessage(error, "加载代理失败"));
  } finally {
    loading.value = false;
  }
}

async function downloadExport() {
  exporting.value = true;
  try {
    const blob = await exportAgents({
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      depth: depth.value ? Number(depth.value) : undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    });
    saveDistributionExport(blob, "distribution-agents");
    app.showSuccess("代理数据已导出");
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "导出代理失败"));
  } finally {
    exporting.value = false;
  }
}

function scheduleLoad() {
  pagination.value.page = 1;
  if (searchTimer) window.clearTimeout(searchTimer);
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

function changeSort(key: string, order: "asc" | "desc") {
  sortBy.value = key;
  sortOrder.value = order;
  pagination.value.page = 1;
  void load();
}

async function searchAgentCandidates(
  query: string,
): Promise<DistributionPickerOption[]> {
  const reasons: Record<string, string> = {
    user_inactive: "用户未启用",
    already_agent: "已经是代理",
    distribution_customer: "已绑定代理",
    affiliate_invitee: "已绑定邀请人",
  };
  const result = await lookupAgentCandidates(query);
  return result.map((user) => ({
    id: user.user_id,
    email: user.email,
    username: user.username,
    meta: user.status === "active" ? "用户正常" : user.status,
    selectable: user.selectable,
    reason: user.unavailable_reason
      ? reasons[user.unavailable_reason] || "当前不可选"
      : undefined,
  }));
}

async function openAddDialog() {
  addDialog.value = true;
  if (!settings.value) {
    try {
      settings.value = await getSettings();
    } catch {
      app.showError("加载默认返佣设置失败");
    }
  }
}

function closeAddDialog() {
  addDialog.value = false;
  selectedUser.value = null;
  ratePercent.value = "";
  promotionCode.value = "";
  showAdvanced.value = false;
}

async function submitAgent() {
  if (!selectedUser.value) return;
  const parsedRate =
    ratePercent.value.trim() === "" ? undefined : Number(ratePercent.value);
  if (
    parsedRate !== undefined &&
    (!Number.isFinite(parsedRate) || parsedRate < 0 || parsedRate > 100)
  ) {
    app.showError("返佣比例必须在 0% 到 100% 之间");
    return;
  }
  submitting.value = true;
  try {
    await grantAgent({
      user_id: selectedUser.value.id,
      rate_override_bps:
        parsedRate === undefined ? undefined : Math.round(parsedRate * 100),
      promotion_code: promotionCode.value.trim().toUpperCase() || undefined,
    });
    app.showSuccess("一级代理已添加");
    closeAddDialog();
    pagination.value.page = 1;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "添加代理失败"));
  } finally {
    submitting.value = false;
  }
}

async function openManageDialog(agent: DistributionAgent) {
  managedAgent.value = agent;
  manageDialog.value = true;
  agentEvents.value = [];
  eventsLoading.value = true;
  if (!settings.value)
    void getSettings()
      .then((value) => {
        settings.value = value;
      })
      .catch(() => app.showError("加载返佣设置失败"));
  try {
    agentEvents.value = await listAgentEvents(agent.id);
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载代理变更记录失败"));
  } finally {
    eventsLoading.value = false;
  }
}

function closeManageDialog() {
  manageDialog.value = false;
  managedAgent.value = null;
  agentEvents.value = [];
}

function agentEventTitle(event: DistributionAgentEvent) {
  if (event.event_type === "created") return "创建代理";
  if (event.event_type === "permission_changed")
    return event.new_status === "true"
      ? "授予下级招募权限"
      : "撤销下级招募权限";
  if (event.event_type === "status_changed")
    return `${statusText(event.old_status || "")} → ${statusText(event.new_status || "")}`;
  const before =
    event.old_effective_rate_bps == null
      ? "-"
      : formatPercent(event.old_effective_rate_bps);
  const after =
    event.new_effective_rate_bps == null
      ? "-"
      : formatPercent(event.new_effective_rate_bps);
  return `返佣比例 ${before} → ${after}`;
}

function openRecruitmentPermissionDialog() {
  recruitmentPermissionReason.value = "";
  recruitmentPermissionDialog.value = true;
}

async function saveRecruitmentPermission() {
  if (!managedAgent.value || !recruitmentPermissionReason.value.trim()) return;
  recruitmentPermissionSaving.value = true;
  try {
    const enabled = !managedAgent.value.can_recruit_subagents;
    await updateAgentRecruitmentPermission(managedAgent.value.id, {
      enabled,
      reason: recruitmentPermissionReason.value.trim(),
    });
    app.showSuccess(enabled ? "已授予下级招募权限" : "已撤销下级招募权限");
    recruitmentPermissionDialog.value = false;
    manageDialog.value = false;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "更新招募权限失败"));
  } finally {
    recruitmentPermissionSaving.value = false;
  }
}

function openStatusDialog(nextStatus: "active" | "suspended" | "revoked") {
  pendingStatus.value = nextStatus;
  statusReason.value = "";
  statusDialog.value = true;
}

async function confirmStatusChange() {
  if (!managedAgent.value) return;
  if (statusReasonRequired.value && !statusReason.value.trim()) return;
  statusSaving.value = true;
  try {
    await updateAgentStatus(
      managedAgent.value.id,
      pendingStatus.value,
      statusReason.value.trim(),
    );
    app.showSuccess("代理状态已更新");
    statusDialog.value = false;
    closeManageDialog();
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "状态更新失败"));
  } finally {
    statusSaving.value = false;
  }
}

function openRateDialog() {
  if (!managedAgent.value) return;
  rateUseDefault.value = managedAgent.value.rate_override_bps == null;
  editRatePercent.value = (
    managedAgent.value.rate_override_bps == null
      ? managedAgent.value.effective_rate_bps
      : managedAgent.value.rate_override_bps
  ).toString();
  editRatePercent.value = (Number(editRatePercent.value) / 100).toFixed(2);
  rateReason.value = "";
  rateDialog.value = true;
}

async function saveRate() {
  if (!managedAgent.value || !canSaveRate.value) return;
  rateSaving.value = true;
  try {
    await updateAgentRate(managedAgent.value.id, {
      use_default: rateUseDefault.value,
      rate_override_bps: rateUseDefault.value
        ? undefined
        : Math.round(Number(editRatePercent.value) * 100),
      reason: rateReason.value.trim(),
    });
    app.showSuccess("返佣比例已更新");
    rateDialog.value = false;
    manageDialog.value = false;
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "更新返佣比例失败"));
  } finally {
    rateSaving.value = false;
  }
}

watch([search, status, depth], scheduleLoad);
onMounted(load);
</script>

<style scoped>
.input-suffix {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--ui-text-muted);
  font-size: 0.875rem;
}
</style>
