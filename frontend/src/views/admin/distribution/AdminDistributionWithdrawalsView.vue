<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">提现管理</h1>
              <p class="page-description">
                审核、打款并追踪完整处理历史，付款凭证可按需留存
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <button
                class="btn btn-secondary h-10 px-3"
                title="导出当前筛选"
                aria-label="导出提现记录"
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
              class="grid grid-cols-[minmax(0,1fr)_112px] gap-2 lg:grid-cols-[minmax(280px,1fr)_150px_150px_150px]"
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
                  placeholder="搜索申请单、代理、账号或流水号"
                  aria-label="搜索提现申请"
                />
              </div>
              <button
                type="button"
                class="btn btn-secondary h-10 px-3 lg:hidden"
                aria-label="筛选提现申请"
                @click="filterDialog = true"
              >
                <Icon name="filter" size="sm" /><span
                  v-if="activeFilterCount"
                  class="badge badge-primary"
                  >{{ activeFilterCount }}</span
                >
              </button>
              <select
                v-model="status"
                class="input hidden lg:block"
                aria-label="提现状态"
              >
                <option value="">全部状态</option>
                <option value="pending">待审核</option>
                <option value="approved">待打款</option>
                <option value="paying">打款中</option>
                <option value="paid">已付款</option>
                <option value="rejected">已拒绝</option>
                <option value="failed">付款失败</option>
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
                共 {{ pagination.total }} 笔，当前页待付
                {{ money(pagePendingPayout) }}
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
          <div
            v-if="selectedIds.length"
            class="flex flex-wrap items-center justify-between gap-3 border-b border-outline bg-surface-subtle px-4 py-3"
          >
            <div>
              <p class="text-sm font-medium">
                已选择 {{ selectedIds.length }} 笔提现
              </p>
              <p
                class="mt-0.5 text-xs"
                :class="
                  selectedAllPending
                    ? 'text-foreground-subtle'
                    : 'text-warning-foreground'
                "
              >
                {{
                  selectedAllPending
                    ? `合计到账 ${money(selectedPayout)}`
                    : "批量审核只支持全部处于待审核状态的申请"
                }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <button
                class="btn btn-secondary btn-sm"
                :disabled="!selectedAllPending"
                @click="openBatch('approved')"
              >
                <Icon name="check" size="sm" />批量通过</button
              ><button
                class="btn btn-danger btn-sm"
                :disabled="!selectedAllPending"
                @click="openBatch('rejected')"
              >
                <Icon name="x" size="sm" />批量拒绝
              </button>
            </div>
          </div>
          <DataTable
            v-model:selectedKeys="selectedKeys"
            :selectable="true"
            selection-label="选择提现申请"
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="id"
            :server-side-sort="true"
            :clickable-rows="true"
            default-sort-key="created_at"
            default-sort-order="desc"
            sort-storage-key="admin-distribution-withdrawals"
            @sort="changeSort"
            @rowClick="openDetail"
          >
            <template #cell-request="{ row }"
              ><div class="min-w-40">
                <p class="font-mono text-xs font-medium">
                  {{ row.request_no }}
                </p>
                <p class="mt-1 text-xs text-foreground-subtle">
                  {{ dateTime(row.created_at) }}
                </p>
              </div></template
            >
            <template #cell-agent="{ row }"
              ><div class="min-w-44 max-w-56">
                <p class="truncate font-medium">
                  {{ row.agent_email || "未设置邮箱" }}
                </p>
                <p class="truncate text-xs text-foreground-subtle">
                  {{ row.agent_username || `用户 #${row.agent_user_id}` }}
                </p>
              </div></template
            >
            <template #cell-account="{ row }"
              ><div class="min-w-36 max-w-48">
                <p class="truncate text-sm">{{ row.alipay_name }}</p>
                <p class="truncate text-xs text-foreground-subtle">
                  {{ maskAccount(row.alipay_account) }}
                </p>
              </div></template
            >
            <template #cell-amount="{ row }"
              ><span class="tabular-nums">{{
                money(row.amount_cny)
              }}</span></template
            >
            <template #cell-payout="{ row }"
              ><div class="text-right">
                <p class="font-semibold tabular-nums">
                  {{ money(row.payout_cny) }}
                </p>
                <p
                  v-if="Number(row.fee_cny) > 0"
                  class="text-xs text-foreground-subtle"
                >
                  手续费 {{ money(row.fee_cny) }}
                </p>
              </div></template
            >
            <template #cell-status="{ row }"
              ><span class="badge" :class="statusClass(row.status)">{{
                statusText(row.status)
              }}</span></template
            >
            <template #cell-evidence="{ row }"
              ><span
                class="whitespace-nowrap text-sm"
                :class="
                  row.attachment_count
                    ? 'text-success-foreground'
                    : 'text-foreground-subtle'
                "
                >{{
                  row.attachment_count
                    ? `${row.attachment_count} 份凭证`
                    : "暂无凭证"
                }}</span
              ></template
            >
            <template #mobile-card="{ row }"
              ><div class="space-y-3">
                <div class="flex min-w-0 items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="truncate font-mono text-xs font-medium">
                      {{ row.request_no }}
                    </p>
                    <p class="mt-1 truncate text-sm">{{ row.agent_email }}</p>
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
                    <p class="text-foreground-subtle">申请金额</p>
                    <p class="mt-1 font-semibold">
                      {{ money(row.amount_cny) }}
                    </p>
                  </div>
                  <div class="text-center">
                    <p class="text-foreground-subtle">手续费</p>
                    <p class="mt-1 font-semibold">{{ money(row.fee_cny) }}</p>
                  </div>
                  <div class="text-right">
                    <p class="text-foreground-subtle">实际到账</p>
                    <p class="mt-1 font-semibold">
                      {{ money(row.payout_cny) }}
                    </p>
                  </div>
                </div>
                <div
                  class="flex items-center justify-between gap-3 text-xs text-foreground-muted"
                >
                  <span>{{
                    row.attachment_count
                      ? `${row.attachment_count} 份凭证`
                      : "暂无凭证"
                  }}</span
                  ><time>{{ dateTime(row.created_at) }}</time>
                </div>
              </div></template
            >
            <template #empty
              ><div class="py-12 text-center">
                <p class="font-medium">暂无符合条件的提现申请</p>
                <p class="mt-1 text-sm text-foreground-subtle">
                  调整筛选条件后重试
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

    <BaseDrawer
      :show="detailDialog"
      title="提现申请详情"
      description="审核状态、收款信息、凭证和处理记录"
      @close="closeDetail"
    >
      <div
        v-if="detailLoading"
        class="flex min-h-48 items-center justify-center"
      >
        <LoadingSpinner />
      </div>
      <div v-else-if="detail" class="space-y-6">
        <div class="flex min-w-0 items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="truncate font-mono text-sm font-semibold">
              {{ detail.withdrawal.request_no }}
            </p>
            <p class="mt-1 truncate text-sm text-foreground-subtle">
              {{ detail.withdrawal.agent_email }} ·
              {{ detail.withdrawal.agent_username || "未设置用户名" }}
            </p>
          </div>
          <span
            class="badge shrink-0"
            :class="statusClass(detail.withdrawal.status)"
            >{{ statusText(detail.withdrawal.status) }}</span
          >
        </div>

        <dl
          class="grid grid-cols-2 gap-x-4 gap-y-4 border-y border-outline py-4 text-sm sm:grid-cols-4"
        >
          <div>
            <dt class="text-foreground-subtle">申请金额</dt>
            <dd class="mt-1 font-semibold">
              {{ money(detail.withdrawal.amount_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">手续费</dt>
            <dd class="mt-1 font-semibold">
              {{ money(detail.withdrawal.fee_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">实际到账</dt>
            <dd class="mt-1 font-semibold text-success-foreground">
              {{ money(detail.withdrawal.payout_cny) }}
            </dd>
          </div>
          <div>
            <dt class="text-foreground-subtle">申请时间</dt>
            <dd class="mt-1 text-xs font-medium">
              {{ dateTime(detail.withdrawal.created_at) }}
            </dd>
          </div>
        </dl>

        <div
          v-if="settings?.withdrawal_dual_approval_enabled"
          class="flex items-start gap-3 rounded-control border border-outline bg-surface-subtle px-4 py-3"
        >
          <Icon name="shield" size="sm" class="mt-0.5 shrink-0 text-brand" />
          <div class="min-w-0">
            <p class="text-sm font-medium">双人复核已启用</p>
            <p class="mt-1 text-xs text-foreground-subtle">
              审核人与开始打款的管理员必须不同。<span
                v-if="detail.withdrawal.approver_email"
                >当前审批人：{{ detail.withdrawal.approver_email }} ·
                {{ dateTime(detail.withdrawal.approved_at) }}</span
              >
            </p>
          </div>
        </div>

        <section>
          <h3 class="text-sm font-semibold">收款信息</h3>
          <div
            class="mt-3 grid gap-3 rounded-control bg-surface-subtle px-4 py-3 text-sm sm:grid-cols-2"
          >
            <div>
              <p class="text-xs text-foreground-subtle">支付宝实名</p>
              <p class="mt-1 font-medium">
                {{ detail.withdrawal.alipay_name }}
              </p>
            </div>
            <div>
              <p class="text-xs text-foreground-subtle">支付宝账号</p>
              <p class="mt-1 break-all font-medium">
                {{ detail.withdrawal.alipay_account }}
              </p>
            </div>
            <div
              v-if="detail.withdrawal.payment_reference"
              class="sm:col-span-2"
            >
              <p class="text-xs text-foreground-subtle">付款流水号</p>
              <p class="mt-1 break-all font-mono font-medium">
                {{ detail.withdrawal.payment_reference }}
              </p>
            </div>
          </div>
        </section>

        <section>
          <div class="flex items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-semibold">打款凭证（可选）</h3>
              <p class="mt-1 text-xs text-foreground-subtle">
                如需留档可上传，文件使用私有存储并仅通过短期签名链接访问
              </p>
            </div>
            <span class="badge badge-gray"
              >{{ detail.attachments.length }} 份</span
            >
          </div>
          <div
            v-if="detail.attachments.length"
            class="mt-3 divide-y divide-outline border-y border-outline"
          >
            <div
              v-for="item in detail.attachments"
              :key="item.id"
              class="flex items-center justify-between gap-3 py-3"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-medium">
                  {{ item.original_name }}
                </p>
                <p class="mt-0.5 text-xs text-foreground-subtle">
                  {{ evidenceTypeText(item.evidence_type) }} ·
                  {{ fileSize(item.size_bytes) }} ·
                  {{ dateTime(item.created_at) }}
                </p>
              </div>
              <button
                class="btn btn-secondary btn-sm shrink-0"
                @click="viewEvidence(item)"
              >
                <Icon name="externalLink" size="sm" />查看
              </button>
            </div>
          </div>
          <p
            v-else
            class="mt-3 py-4 text-center text-sm text-foreground-subtle"
          >
            尚未上传打款凭证
          </p>
          <form
            v-if="detail.withdrawal.status === 'paying'"
            class="mt-4 space-y-3 border-t border-outline pt-4"
            @submit.prevent="uploadEvidence"
          >
            <div class="grid gap-3 sm:grid-cols-[160px_minmax(0,1fr)]">
              <select
                v-model="evidenceType"
                class="input"
                aria-label="凭证类型"
              >
                <option value="payment_receipt">付款回单</option>
                <option value="bank_statement">银行流水</option>
                <option value="other">其他凭证</option></select
              ><input
                v-model="evidenceNote"
                class="input"
                maxlength="200"
                placeholder="凭证备注（可选）"
              />
            </div>
            <div
              class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between"
            >
              <label
                class="btn btn-secondary cursor-pointer"
                :class="
                  !evidenceCapabilities?.available
                    ? 'pointer-events-none opacity-50'
                    : ''
                "
                ><Icon name="upload" size="sm" />选择图片或 PDF<input
                  class="sr-only"
                  type="file"
                  multiple
                  :accept="evidenceAccept"
                  :disabled="!evidenceCapabilities?.available"
                  @change="selectEvidenceFiles"
              /></label>
              <p class="min-w-0 truncate text-xs text-foreground-subtle">
                {{
                  evidenceFiles.length
                    ? `${evidenceFiles.length} 个文件：${evidenceFiles.map((file) => file.name).join("、")}`
                    : evidenceStorageHint
                }}
              </p>
              <button
                class="btn btn-primary"
                type="submit"
                :disabled="uploading || !evidenceFiles.length"
              >
                <Icon name="upload" size="sm" />{{
                  uploading ? "上传中..." : "上传凭证"
                }}
              </button>
            </div>
          </form>
        </section>

        <section>
          <h3 class="text-sm font-semibold">处理记录</h3>
          <ol class="mt-3 space-y-3 border-l border-outline pl-4">
            <li v-for="event in detail.events" :key="event.id" class="relative">
              <span
                class="absolute -left-[21px] top-1.5 h-2.5 w-2.5 rounded-full border-2 border-surface bg-brand"
              ></span>
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-sm font-medium">
                  {{
                    event.from_status
                      ? `${statusText(event.from_status)} → `
                      : ""
                  }}{{ statusText(event.to_status) }}
                </p>
                <time class="text-xs text-foreground-subtle">{{
                  dateTime(event.created_at)
                }}</time>
              </div>
              <p class="mt-1 text-xs text-foreground-subtle">
                {{ event.actor_email || "系统"
                }}<span v-if="event.note"> · {{ event.note }}</span
                ><span v-if="event.payment_reference">
                  · 流水号 {{ event.payment_reference }}</span
                >
              </p>
            </li>
          </ol>
        </section>

        <div
          class="flex flex-wrap justify-end gap-2 border-t border-outline pt-4"
        >
          <button
            v-if="detail.withdrawal.status === 'pending'"
            class="btn btn-danger"
            @click="openAction('rejected')"
          >
            拒绝申请</button
          ><button
            v-if="detail.withdrawal.status === 'pending'"
            class="btn btn-primary"
            @click="openAction('approved')"
          >
            审核通过</button
          ><button
            v-if="detail.withdrawal.status === 'approved'"
            class="btn btn-danger"
            @click="openAction('rejected')"
          >
            终止并拒绝</button
          ><button
            v-if="detail.withdrawal.status === 'approved'"
            class="btn btn-primary"
            :disabled="cannotStartPayment"
            :title="
              cannotStartPayment ? '双人复核要求由其他管理员开始打款' : ''
            "
            @click="openAction('paying')"
          >
            开始打款</button
          ><button
            v-if="detail.withdrawal.status === 'paying'"
            class="btn btn-danger"
            @click="openAction('failed')"
          >
            登记失败</button
          ><button
            v-if="detail.withdrawal.status === 'paying'"
            class="btn btn-primary"
            @click="openAction('paid')"
          >
            <Icon name="check" size="sm" />确认已付款
          </button>
        </div>
      </div>
    </BaseDrawer>

    <BaseDialog
      :show="actionDialog"
      :title="actionTitle"
      width="narrow"
      @close="actionDialog = false"
    >
      <div class="space-y-4">
        <p class="text-sm text-foreground-subtle">{{ actionHint }}</p>
        <div v-if="nextStatus === 'paid'">
          <label for="withdrawal-reference" class="input-label"
            >付款流水号</label
          ><input
            id="withdrawal-reference"
            v-model="actionReference"
            class="input font-mono"
            maxlength="255"
            placeholder="填写银行或支付平台流水号"
          />
        </div>
        <div>
          <label for="withdrawal-note" class="input-label"
            >处理备注{{ actionReasonRequired ? "" : "（可选）" }}</label
          ><textarea
            id="withdrawal-note"
            v-model="actionNote"
            class="input min-h-24 resize-y"
            maxlength="500"
            placeholder="记录审核依据、失败原因或其他说明"
          ></textarea>
        </div>
      </div>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="actionDialog = false">
            取消</button
          ><button
            class="btn"
            :class="
              nextStatus === 'rejected' || nextStatus === 'failed'
                ? 'btn-danger'
                : 'btn-primary'
            "
            :disabled="actionSaving || !canSubmitAction"
            @click="submitAction"
          >
            {{ actionSaving ? "处理中..." : "确认提交" }}
          </button>
        </div></template
      >
    </BaseDialog>
    <BaseDialog
      :show="filterDialog"
      title="筛选提现申请"
      width="narrow"
      @close="filterDialog = false"
      ><div class="space-y-4">
        <div>
          <label class="input-label">处理状态</label
          ><select v-model="status" class="input">
            <option value="">全部状态</option>
            <option value="pending">待审核</option>
            <option value="approved">待打款</option>
            <option value="paying">打款中</option>
            <option value="paid">已付款</option>
            <option value="rejected">已拒绝</option>
            <option value="failed">付款失败</option>
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
    <BaseDialog
      :show="batchDialog"
      :title="batchStatus === 'approved' ? '批量审核通过' : '批量拒绝提现'"
      width="narrow"
      @close="batchDialog = false"
      ><div class="space-y-4">
        <div
          class="rounded-control border border-outline bg-surface-subtle px-4 py-3 text-sm"
        >
          <p>
            将处理 <strong>{{ selectedIds.length }}</strong> 笔提现申请
          </p>
          <p class="mt-1 text-foreground-subtle">
            合计实际到账
            {{ money(selectedPayout) }}，提交后将作为一个原子批次处理。
          </p>
        </div>
        <div>
          <label for="batch-review-note" class="input-label">{{
            batchStatus === "approved" ? "审核依据" : "拒绝原因"
          }}</label
          ><textarea
            id="batch-review-note"
            v-model="batchNote"
            class="input min-h-24 resize-y"
            maxlength="500"
            :placeholder="
              batchStatus === 'approved'
                ? '记录身份、金额和风控核验依据'
                : '填写统一拒绝原因，资金将退回代理可用佣金'
            "
          ></textarea>
        </div>
        <p class="text-xs text-foreground-subtle">
          任一申请状态发生变化时，整个批次都会失败，不会出现部分成功。
        </p>
      </div>
      <template #footer
        ><div class="flex justify-end gap-2">
          <button class="btn btn-secondary" @click="batchDialog = false">
            取消</button
          ><button
            class="btn"
            :class="batchStatus === 'rejected' ? 'btn-danger' : 'btn-primary'"
            :disabled="batchSaving || !batchNote.trim()"
            @click="submitBatch"
          >
            {{ batchSaving ? "处理中..." : "确认处理" }}
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
import LoadingSpinner from "@/components/common/LoadingSpinner.vue";
import Icon from "@/components/icons/Icon.vue";
import DistributionFilterSummary, { type DistributionFilterItem } from "@/components/distribution/DistributionFilterSummary.vue";
import type { Column } from "@/components/common/types";
import {
  batchReviewWithdrawals,
  exportWithdrawals,
  getSettings,
  getWithdrawal,
  getWithdrawalEvidenceCapabilities,
  getWithdrawalEvidenceURL,
  listWithdrawals,
  reviewWithdrawal,
  uploadWithdrawalEvidence,
  type DistributionEvidenceCapabilities,
  type DistributionSettings,
  type DistributionWithdrawal,
  type DistributionWithdrawalAttachment,
  type DistributionWithdrawalDetail,
} from "@/api/admin/distribution";
import { useAppStore } from "@/stores/app";
import { useAuthStore } from "@/stores/auth";
import { extractApiErrorMessage } from "@/utils/apiError";
import { saveDistributionExport } from "@/utils/distributionExport";

const app = useAppStore();
const auth = useAuthStore();
const route = useRoute();
const items = ref<DistributionWithdrawal[]>([]);
const loading = ref(false);
const exporting = ref(false);
const search = ref(String(route.query.search || ""));
const requestedStatus = String(route.query.status || "");
const status = ref(
  ["pending", "approved", "paying", "paid", "rejected", "failed"].includes(
    requestedStatus,
  )
    ? requestedStatus
    : "",
);
const dateFrom = ref("");
const dateTo = ref("");
const sortBy = ref("created_at");
const sortOrder = ref<"asc" | "desc">("desc");
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 });
const detailDialog = ref(false);
const filterDialog = ref(false);
const detailLoading = ref(false);
const detail = ref<DistributionWithdrawalDetail | null>(null);
const evidenceCapabilities = ref<DistributionEvidenceCapabilities | null>(null);
const settings = ref<DistributionSettings | null>(null);
const evidenceFiles = ref<File[]>([]);
const evidenceType = ref("payment_receipt");
const evidenceNote = ref("");
const uploading = ref(false);
const actionDialog = ref(false);
const nextStatus = ref<DistributionWithdrawal["status"]>("approved");
const actionNote = ref("");
const actionReference = ref("");
const actionSaving = ref(false);
const selectedKeys = ref<Array<string | number>>([]);
const batchDialog = ref(false);
const batchStatus = ref<"approved" | "rejected">("approved");
const batchNote = ref("");
const batchSaving = ref(false);
let searchTimer: number | null = null;

const columns: Column[] = [
  { key: "request", label: "申请单" },
  { key: "agent", label: "代理", sortable: true },
  { key: "account", label: "收款账户" },
  { key: "amount", label: "申请金额", sortable: true },
  { key: "payout", label: "实际到账", sortable: true },
  { key: "status", label: "状态", sortable: true },
  { key: "evidence", label: "审计凭证" },
];
const hasFilters = computed(() =>
  Boolean(search.value || status.value || dateFrom.value || dateTo.value),
);
const filterSummary = computed<DistributionFilterItem[]>(() => [
  search.value ? { key: 'search', label: '搜索', value: search.value } : null,
  status.value ? { key: 'status', label: '状态', value: statusText(status.value) } : null,
  dateFrom.value ? { key: 'from', label: '开始', value: dateFrom.value } : null,
  dateTo.value ? { key: 'to', label: '结束', value: dateTo.value } : null,
].filter((item): item is DistributionFilterItem => Boolean(item)));
function removeFilter(key: string) { if (key === 'search') search.value = ''; if (key === 'status') status.value = ''; if (key === 'from') dateFrom.value = ''; if (key === 'to') dateTo.value = ''; }
const activeFilterCount = computed(
  () => [status.value, dateFrom.value, dateTo.value].filter(Boolean).length,
);
const pagePendingPayout = computed(() =>
  items.value
    .filter((item) => ["pending", "approved", "paying"].includes(item.status))
    .reduce((sum, item) => sum + Number(item.payout_cny), 0),
);
const selectedIds = computed(() =>
  selectedKeys.value.map(Number).filter(Number.isFinite),
);
const selectedRows = computed(() =>
  items.value.filter((item) => selectedIds.value.includes(item.id)),
);
const selectedAllPending = computed(
  () =>
    selectedRows.value.length === selectedIds.value.length &&
    selectedRows.value.length > 0 &&
    selectedRows.value.every((item) => item.status === "pending"),
);
const selectedPayout = computed(() =>
  selectedRows.value.reduce((sum, item) => sum + Number(item.payout_cny), 0),
);
const cannotStartPayment = computed(() =>
  Boolean(
    settings.value?.withdrawal_dual_approval_enabled &&
    detail.value?.withdrawal.approved_by &&
    detail.value.withdrawal.approved_by === auth.user?.id,
  ),
);
const evidenceAccept = computed(
  () =>
    evidenceCapabilities.value?.allowed_extensions.join(",") ||
    ".png,.jpg,.jpeg,.webp,.pdf",
);
const evidenceStorageHint = computed(() =>
  evidenceCapabilities.value?.available
    ? `最多 ${evidenceCapabilities.value.max_files_per_upload} 个文件，单个不超过 ${fileSize(evidenceCapabilities.value.max_file_bytes)}`
    : "私有凭证存储未配置，请先在服务端启用工单对象存储",
);
const actionReasonRequired = computed(
  () => nextStatus.value === "rejected" || nextStatus.value === "failed",
);
const canSubmitAction = computed(
  () =>
    (!actionReasonRequired.value || Boolean(actionNote.value.trim())) &&
    (nextStatus.value !== "paid" || Boolean(actionReference.value.trim())),
);
const actionTitle = computed(
  () =>
    ({
      pending: "恢复待审核",
      approved: "审核通过",
      rejected: "拒绝提现",
      paying: "开始打款",
      paid: "确认已付款",
      failed: "登记付款失败",
    })[nextStatus.value] || "更新提现状态",
);
const actionHint = computed(() =>
  nextStatus.value === "paying" &&
  settings.value?.withdrawal_dual_approval_enabled
    ? "双人复核已启用：当前操作人必须与审批人不同。开始打款后可按需上传支付回单或银行流水。"
    : {
        pending: "",
        approved: "审核通过后申请进入待打款队列，资金仍保持冻结。",
        rejected: "拒绝后冻结资金将退回代理可用佣金，请完整记录原因。",
        paying: "确认开始打款后，可按需上传支付回单或银行流水。",
        paid: "确认前请核对到账金额、收款账户和付款流水号；打款凭证为可选留档。",
        failed: "付款失败后资金仍保持冻结，便于复核后重新处理。",
      }[nextStatus.value] || "",
);
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`;
const dateTime = (value?: string) =>
  value ? new Date(value).toLocaleString() : "-";
const statusText = (value: string) =>
  ({
    pending: "待审核",
    approved: "待打款",
    paying: "打款中",
    paid: "已付款",
    rejected: "已拒绝",
    failed: "付款失败",
  })[value] || value;
const statusClass = (value: string) =>
  ({
    pending: "badge-warning",
    approved: "badge-primary",
    paying: "badge-primary",
    paid: "badge-success",
    rejected: "badge-danger",
    failed: "badge-danger",
  })[value] || "badge-gray";
const evidenceTypeText = (value: string) =>
  ({
    payment_receipt: "付款回单",
    bank_statement: "银行流水",
    other: "其他凭证",
  })[value] || value;
const fileSize = (bytes: number) =>
  bytes >= 1024 * 1024
    ? `${(bytes / 1024 / 1024).toFixed(1)} MB`
    : `${Math.max(1, Math.round(bytes / 1024))} KB`;
const maskAccount = (value: string) =>
  value.length <= 7 ? value : `${value.slice(0, 3)}****${value.slice(-3)}`;
function toStart(value: string) {
  return value ? new Date(`${value}T00:00:00`).toISOString() : undefined;
}
function toEnd(value: string) {
  return value ? new Date(`${value}T23:59:59.999`).toISOString() : undefined;
}

async function load() {
  loading.value = true;
  selectedKeys.value = [];
  try {
    const result = await listWithdrawals({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      status: status.value || undefined,
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
    app.showError(extractApiErrorMessage(error, "加载提现申请失败"));
  } finally {
    loading.value = false;
  }
}
async function downloadExport() {
  exporting.value = true;
  try {
    const blob = await exportWithdrawals({
      search: search.value.trim() || undefined,
      status: status.value || undefined,
      date_from: toStart(dateFrom.value),
      date_to: toEnd(dateTo.value),
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
    });
    saveDistributionExport(blob, "distribution-withdrawals");
    app.showSuccess("提现记录已导出");
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "导出提现记录失败"));
  } finally {
    exporting.value = false;
  }
}
async function refreshDetail() {
  if (!detail.value) return;
  detail.value = await getWithdrawal(detail.value.withdrawal.id);
}
async function openDetail(row: DistributionWithdrawal) {
  detailDialog.value = true;
  detailLoading.value = true;
  evidenceFiles.value = [];
  try {
    detail.value = await getWithdrawal(row.id);
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "加载提现详情失败"));
    detailDialog.value = false;
  } finally {
    detailLoading.value = false;
  }
}
function closeDetail() {
  detailDialog.value = false;
  detail.value = null;
  evidenceFiles.value = [];
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
  dateFrom.value = "";
  dateTo.value = "";
  pagination.value.page = 1;
  void load();
}
function selectEvidenceFiles(event: Event) {
  evidenceFiles.value = Array.from(
    (event.target as HTMLInputElement).files || [],
  );
}
async function uploadEvidence() {
  if (!detail.value || !evidenceFiles.value.length) return;
  uploading.value = true;
  try {
    await uploadWithdrawalEvidence(
      detail.value.withdrawal.id,
      evidenceFiles.value,
      evidenceType.value,
      evidenceNote.value.trim(),
    );
    app.showSuccess("打款凭证已上传");
    evidenceFiles.value = [];
    evidenceNote.value = "";
    await refreshDetail();
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "上传打款凭证失败"));
  } finally {
    uploading.value = false;
  }
}
async function viewEvidence(item: DistributionWithdrawalAttachment) {
  if (!detail.value) return;
  try {
    const url = await getWithdrawalEvidenceURL(
      detail.value.withdrawal.id,
      item.id,
    );
    window.open(url, "_blank", "noopener,noreferrer");
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "获取凭证访问链接失败"));
  }
}
function openAction(statusValue: DistributionWithdrawal["status"]) {
  nextStatus.value = statusValue;
  actionNote.value = "";
  actionReference.value = detail.value?.withdrawal.payment_reference || "";
  actionDialog.value = true;
}
async function submitAction() {
  if (!detail.value || !canSubmitAction.value) return;
  actionSaving.value = true;
  try {
    await reviewWithdrawal(detail.value.withdrawal.id, {
      status: nextStatus.value,
      note: actionNote.value.trim() || undefined,
      payment_reference: actionReference.value.trim() || undefined,
    });
    app.showSuccess("提现状态已更新");
    actionDialog.value = false;
    await refreshDetail();
    await load();
  } catch (error) {
    app.showError(extractApiErrorMessage(error, "更新提现状态失败"));
  } finally {
    actionSaving.value = false;
  }
}
function openBatch(next: "approved" | "rejected") {
  if (!selectedAllPending.value) return;
  batchStatus.value = next;
  batchNote.value = "";
  batchDialog.value = true;
}
async function submitBatch() {
  if (!selectedAllPending.value || !batchNote.value.trim()) return;
  batchSaving.value = true;
  try {
    const result = await batchReviewWithdrawals({
      withdrawal_ids: selectedIds.value,
      status: batchStatus.value,
      note: batchNote.value.trim(),
    });
    app.showSuccess(`已批量处理 ${result.updated} 笔提现申请`);
    batchDialog.value = false;
    await load();
  } catch (error) {
    app.showError(
      extractApiErrorMessage(error, "批量审核失败，所有申请均未变更"),
    );
  } finally {
    batchSaving.value = false;
  }
}

watch(search, scheduleLoad);
watch([status, dateFrom, dateTo], () => {
  pagination.value.page = 1;
  void load();
});
onMounted(async () => {
  await Promise.all([
    load(),
    getSettings().then((value) => {
      settings.value = value;
    }),
    getWithdrawalEvidenceCapabilities()
      .then((value) => {
        evidenceCapabilities.value = value;
      })
      .catch(() => {
        evidenceCapabilities.value = null;
      }),
  ]);
});
</script>
