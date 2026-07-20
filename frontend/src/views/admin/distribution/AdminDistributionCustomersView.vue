<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0">
              <h1 class="page-title">代理客户</h1>
              <p class="page-description">查询客户绑定关系并审计归属调整</p>
            </div>
            <div class="flex shrink-0 items-center gap-2"><button type="button" class="btn btn-secondary h-10 px-3" title="导出当前筛选" aria-label="导出代理客户" :disabled="exporting" @click="downloadExport"><Icon name="download" size="sm" /><span class="hidden sm:inline">导出</span></button><button type="button" class="btn btn-secondary btn-icon" title="刷新" aria-label="刷新" :disabled="loading" @click="load"><Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" /></button></div>
          </header>
        </template>

        <template #filters>
          <div class="flex min-w-0 items-center gap-2 sm:max-w-2xl">
            <div class="relative min-w-0 flex-1">
              <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
              <input
                v-model="search"
                class="input pl-9"
                type="search"
                placeholder="搜索客户、代理或推广码"
                aria-label="搜索代理客户"
              />
            </div>
            <button
              type="button"
              class="btn btn-secondary h-10 shrink-0 px-3"
              title="筛选"
              aria-label="筛选"
              @click="openFilterDialog"
            >
              <Icon name="filter" size="sm" />
              <span class="hidden sm:inline">筛选</span>
              <span v-if="activeFilterCount" class="badge badge-primary ml-1">{{ activeFilterCount }}</span>
            </button>
          </div>
        </template>

        <template #table>
          <DataTable
            :columns="columns"
            :data="items"
            :loading="loading"
            row-key="user_id"
            :actions-count="1"
			:server-side-sort="true"
			default-sort-key="last_paid_at"
			default-sort-order="desc"
			sort-storage-key="admin-distribution-customers"
			@sort="changeSort"
          >
			<template #cell-email="{ row }">
              <div class="min-w-52">
                <p class="truncate font-medium text-foreground">{{ row.email || '未设置邮箱' }}</p>
                <p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ row.username || '未设置用户名' }}</p>
              </div>
            </template>
			<template #cell-agent="{ row }">
              <div class="min-w-52">
                <p class="truncate font-medium text-foreground">{{ row.agent_email || '未知代理' }}</p>
				<p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ row.agent_depth === 1 ? '一级代理' : '二级代理' }} · {{ row.agent_promotion_code || '-' }}</p>
              </div>
            </template>
			<template #cell-order_count="{ row }"><span class="tabular-nums">{{ row.order_count || 0 }}</span></template>
			<template #cell-total_paid="{ row }"><span class="font-medium tabular-nums">{{ money(row.total_paid_cny) }}</span></template>
			<template #cell-commission="{ row }"><span class="font-semibold tabular-nums text-success-foreground">{{ money(row.commission_cny) }}</span></template>
			<template #cell-refunded="{ row }"><span class="tabular-nums" :class="Number(row.refunded_cny) > 0 ? 'text-danger-foreground' : 'text-foreground-subtle'">{{ money(row.refunded_cny) }}</span></template>
			<template #cell-last_paid_at="{ row }"><span class="whitespace-nowrap text-sm text-foreground-muted">{{ row.last_paid_at ? formatDate(row.last_paid_at) : '尚未付费' }}</span></template>
            <template #cell-actions="{ row }">
              <button
                type="button"
                class="btn btn-ghost btn-icon btn-sm"
                title="调整归属"
                aria-label="调整归属"
                @click.stop="openCorrectionDialog(row)"
              >
                <Icon name="edit" size="sm" />
              </button>
            </template>

            <template #mobile-card="{ row }">
              <div class="space-y-3">
                <div class="min-w-0">
                  <p class="truncate font-medium text-foreground">{{ row.email || '未设置邮箱' }}</p>
                  <p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ row.username || '未设置用户名' }}</p>
                </div>
                <div class="rounded-control bg-surface-subtle px-3 py-2">
                  <p class="text-xs text-foreground-subtle">当前归属</p>
                  <div class="mt-1 flex min-w-0 items-center justify-between gap-2">
                    <div class="min-w-0">
                      <p class="truncate text-sm font-medium">{{ row.agent_email || '未知代理' }}</p>
                      <p class="truncate font-mono text-xs text-foreground-muted">{{ row.agent_promotion_code || '-' }}</p>
                    </div>
                    <span class="badge shrink-0" :class="row.agent_depth === 1 ? 'badge-primary' : 'badge-gray'">
                      {{ row.agent_depth === 1 ? '一级' : '二级' }}
                    </span>
                  </div>
                </div>
				<dl class="grid grid-cols-3 gap-2 rounded-control bg-surface-subtle px-3 py-2.5 text-xs"><div><dt class="text-foreground-subtle">实付金额</dt><dd class="mt-1 font-semibold">{{ money(row.total_paid_cny) }}</dd></div><div class="text-center"><dt class="text-foreground-subtle">贡献佣金</dt><dd class="mt-1 font-semibold text-success-foreground">{{ money(row.commission_cny) }}</dd></div><div class="text-right"><dt class="text-foreground-subtle">订单 / 退款</dt><dd class="mt-1 font-semibold">{{ row.order_count || 0 }} / {{ money(row.refunded_cny) }}</dd></div></dl>
                <div class="flex items-center justify-between gap-3 border-t border-outline pt-3">
				  <time class="text-xs text-foreground-muted">{{ row.last_paid_at ? `最近付费 ${formatDate(row.last_paid_at)}` : `绑定于 ${formatDate(row.bound_at)}` }}</time>
                  <button type="button" class="btn btn-secondary btn-sm" @click="openCorrectionDialog(row)">
                    <Icon name="edit" size="sm" />调整归属
                  </button>
                </div>
              </div>
            </template>

            <template #empty>
              <div class="py-10 text-center">
                <p class="font-medium text-foreground">暂无客户绑定记录</p>
                <p class="mt-1 text-sm text-foreground-subtle">调整搜索或筛选条件后重试</p>
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

    <BaseDialog :show="filterDialog" title="筛选代理客户" width="narrow" @close="filterDialog = false">
      <div class="space-y-5">
        <div>
          <label for="distribution-customer-depth" class="input-label">代理等级</label>
          <select id="distribution-customer-depth" v-model="draftDepth" class="input">
            <option value="">全部等级</option>
            <option value="1">一级代理</option>
            <option value="2">二级代理</option>
          </select>
        </div>
        <RemoteEntityCombobox
          v-model="draftAgent"
          input-id="distribution-customer-agent-filter"
          label="所属代理"
          placeholder="输入代理邮箱、用户名或推广码"
          :search="searchAgents"
        />
      </div>
      <template #footer>
        <div class="flex w-full items-center justify-between gap-2">
          <button type="button" class="btn btn-ghost" @click="clearFilters">清除筛选</button>
          <div class="flex gap-2">
            <button type="button" class="btn btn-secondary" @click="filterDialog = false">取消</button>
            <button type="button" class="btn btn-primary" @click="applyFilters">应用筛选</button>
          </div>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="correctionDialog" title="调整客户归属" width="normal" @close="closeCorrectionDialog">
      <form id="distribution-correction-form" class="space-y-5" @submit.prevent="reviewCorrection">
        <div v-if="selectedCustomer" class="space-y-3">
          <div>
            <p class="input-label">客户</p>
            <div class="rounded-control bg-surface-subtle px-3 py-2.5">
              <p class="truncate text-sm font-medium text-foreground">{{ selectedCustomer.email || '未设置邮箱' }}</p>
              <p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ selectedCustomer.username || '未设置用户名' }}</p>
            </div>
          </div>
          <div>
            <p class="input-label">当前代理</p>
            <div class="rounded-control border border-outline px-3 py-2.5">
              <p class="truncate text-sm font-medium text-foreground">{{ selectedCustomer.agent_email || '未知代理' }}</p>
              <p class="mt-0.5 truncate text-xs text-foreground-subtle">
                {{ selectedCustomer.agent_username || '未设置用户名' }} · {{ selectedCustomer.agent_promotion_code }}
              </p>
            </div>
          </div>
        </div>
		<section v-if="bindingEvents.length" class="border-y border-outline py-3"><div class="flex items-center justify-between gap-3"><h3 class="text-sm font-semibold">归属调整历史</h3><span class="text-xs text-foreground-subtle">{{ bindingEvents.length }} 条</span></div><ol class="mt-3 space-y-3"><li v-for="event in bindingEvents" :key="event.id" class="text-xs"><div class="flex min-w-0 items-center gap-2"><span class="min-w-0 flex-1 truncate">{{ event.old_agent_email || '首次绑定' }}</span><Icon name="arrowRight" size="xs" class="shrink-0 text-foreground-subtle" /><span class="min-w-0 flex-1 truncate text-right font-medium">{{ event.new_agent_email }}</span></div><p class="mt-1 text-foreground-subtle">{{ formatDate(event.created_at) }} · {{ event.actor_email || '系统' }} · {{ event.reason }}</p></li></ol></section>

        <RemoteEntityCombobox
          v-model="targetAgent"
          input-id="distribution-correction-target"
          label="目标代理"
          placeholder="输入代理邮箱、用户名或推广码"
          :search="searchCorrectionTargets"
        />

        <div>
          <label for="distribution-correction-reason" class="input-label">调整原因</label>
          <textarea
            id="distribution-correction-reason"
            v-model="correctionReason"
            class="input min-h-24 resize-y"
            maxlength="200"
            placeholder="说明为什么需要调整客户归属"
          ></textarea>
          <p class="input-hint">原因会保存到审计记录中</p>
        </div>

        <div v-if="selectedCustomer && targetAgent" class="rounded-control border border-warning/30 bg-warning-subtle px-3 py-3">
          <p class="text-xs font-medium text-warning-foreground">归属变化</p>
          <div class="mt-2 flex min-w-0 items-center gap-2 text-sm">
            <span class="min-w-0 flex-1 truncate">{{ selectedCustomer.agent_email }}</span>
            <Icon name="arrowRight" size="sm" class="shrink-0 text-warning-foreground" />
            <span class="min-w-0 flex-1 truncate text-right font-medium">{{ targetAgent.email }}</span>
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeCorrectionDialog">取消</button>
          <button
            type="submit"
            form="distribution-correction-form"
            class="btn btn-primary"
            :disabled="!canReviewCorrection"
          >
            继续确认
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="confirmationDialog"
      title="确认调整客户归属"
      :message="confirmationMessage"
      confirm-text="确认调整"
      :danger="true"
      @confirm="submitCorrection"
      @cancel="confirmationDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import RemoteEntityCombobox from '@/components/admin/distribution/RemoteEntityCombobox.vue'
import type { DistributionPickerOption } from '@/components/admin/distribution/types'
import type { Column } from '@/components/common/types'
import type { DistributionCustomer } from '@/api/distribution'
import {
  correctCustomerBinding,
	exportCustomers,
	listBindingEvents,
  listCustomers,
  lookupAgents,
	type DistributionBindingEvent,
} from '@/api/admin/distribution'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { saveDistributionExport } from '@/utils/distributionExport'

const app = useAppStore()
const route = useRoute()
const items = ref<DistributionCustomer[]>([])
const loading = ref(false)
const exporting = ref(false)
const search = ref(String(route.query.search || ''))
const sortBy = ref('last_paid_at')
const sortOrder = ref<'asc' | 'desc'>('desc')
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 })
let searchTimer: number | null = null

const filterDialog = ref(false)
const appliedDepth = ref('')
const appliedAgent = ref<DistributionPickerOption | null>(null)
const draftDepth = ref('')
const draftAgent = ref<DistributionPickerOption | null>(null)

const correctionDialog = ref(false)
const confirmationDialog = ref(false)
const selectedCustomer = ref<DistributionCustomer | null>(null)
const targetAgent = ref<DistributionPickerOption | null>(null)
const correctionReason = ref('')
const correctionSaving = ref(false)
const bindingEvents = ref<DistributionBindingEvent[]>([])

const columns: Column[] = [
	{ key: 'email', label: '客户', sortable: true },
	{ key: 'agent', label: '所属代理', sortable: true },
	{ key: 'order_count', label: '订单数', sortable: true },
	{ key: 'total_paid', label: '累计实付', sortable: true },
	{ key: 'commission', label: '贡献佣金', sortable: true },
	{ key: 'refunded', label: '退款金额', sortable: true },
	{ key: 'last_paid_at', label: '最近付费', sortable: true },
  { key: 'actions', label: '操作' },
]

const activeFilterCount = computed(() => Number(Boolean(appliedDepth.value)) + Number(Boolean(appliedAgent.value)))
const canReviewCorrection = computed(() => Boolean(
  selectedCustomer.value
  && targetAgent.value
  && targetAgent.value.id !== selectedCustomer.value.agent_id
  && correctionReason.value.trim(),
))
const confirmationMessage = computed(() => {
  if (!selectedCustomer.value || !targetAgent.value) return ''
  return `将客户 ${selectedCustomer.value.email || selectedCustomer.value.username} 从 ${selectedCustomer.value.agent_email || '原代理'} 调整到 ${targetAgent.value.email}。后续佣金将按新归属计算。`
})

function formatDate(value: string) {
  return new Date(value).toLocaleString()
}

function money(value: string | number) {
	return `¥${Number(value || 0).toFixed(2)}`
}

async function load() {
  loading.value = true
  try {
    const result = await listCustomers({
      page: pagination.value.page,
      page_size: pagination.value.page_size,
      search: search.value.trim() || undefined,
      depth: appliedDepth.value ? Number(appliedDepth.value) : undefined,
      agent_id: appliedAgent.value?.id,
	  sort_by: sortBy.value,
	  sort_order: sortOrder.value,
    })
    items.value = result.items
    pagination.value = {
      page: result.page,
      page_size: result.page_size,
      total: result.total,
      pages: result.pages,
    }
  } catch (error) {
    app.showError(extractApiErrorMessage(error, '加载客户绑定失败'))
  } finally {
    loading.value = false
  }
}

async function downloadExport() {
  exporting.value = true
  try {
    const blob = await exportCustomers({ search: search.value.trim() || undefined, depth: appliedDepth.value ? Number(appliedDepth.value) : undefined, agent_id: appliedAgent.value?.id, sort_by: sortBy.value, sort_order: sortOrder.value })
    saveDistributionExport(blob, 'distribution-customers')
    app.showSuccess('代理客户数据已导出')
  } catch (error) { app.showError(extractApiErrorMessage(error, '导出客户失败')) }
  finally { exporting.value = false }
}

function scheduleLoad() {
  pagination.value.page = 1
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(load, 300)
}

function changePage(page: number) {
  pagination.value.page = page
  void load()
}

function changePageSize(pageSize: number) {
  pagination.value.page = 1
  pagination.value.page_size = pageSize
  void load()
}

function changeSort(key: string, order: 'asc' | 'desc') {
	sortBy.value = key
	sortOrder.value = order
	pagination.value.page = 1
	void load()
}

async function searchAgents(query: string): Promise<DistributionPickerOption[]> {
  const result = await lookupAgents(query)
  return result.map((agent) => ({
    id: agent.agent_id,
    email: agent.email,
    username: agent.username,
    meta: `${agent.depth === 1 ? '一级代理' : '二级代理'} · ${agent.promotion_code}`,
  }))
}

async function searchCorrectionTargets(query: string) {
  const result = await searchAgents(query)
  return result.filter((agent) => agent.id !== selectedCustomer.value?.agent_id)
}

function openFilterDialog() {
  draftDepth.value = appliedDepth.value
  draftAgent.value = appliedAgent.value
  filterDialog.value = true
}

function clearFilters() {
  draftDepth.value = ''
  draftAgent.value = null
}

function applyFilters() {
  appliedDepth.value = draftDepth.value
  appliedAgent.value = draftAgent.value
  pagination.value.page = 1
  filterDialog.value = false
  void load()
}

async function openCorrectionDialog(customer: DistributionCustomer) {
  selectedCustomer.value = customer
  targetAgent.value = null
  correctionReason.value = ''
	bindingEvents.value = []
  correctionDialog.value = true
	try { bindingEvents.value = await listBindingEvents(customer.user_id) }
	catch (error) { app.showError(extractApiErrorMessage(error, '加载客户归属历史失败')) }
}

function closeCorrectionDialog() {
  correctionDialog.value = false
  selectedCustomer.value = null
  targetAgent.value = null
  correctionReason.value = ''
	bindingEvents.value = []
}

function reviewCorrection() {
  if (!canReviewCorrection.value) return
  confirmationDialog.value = true
}

async function submitCorrection() {
  if (!selectedCustomer.value || !targetAgent.value || correctionSaving.value) return
  correctionSaving.value = true
  try {
    await correctCustomerBinding(
      selectedCustomer.value.user_id,
      targetAgent.value.id,
      correctionReason.value.trim(),
    )
    app.showSuccess('客户归属已调整')
    confirmationDialog.value = false
    closeCorrectionDialog()
    await load()
  } catch (error) {
    app.showError(extractApiErrorMessage(error, '调整客户归属失败'))
  } finally {
    correctionSaving.value = false
  }
}

watch(search, scheduleLoad)
onMounted(load)
</script>
