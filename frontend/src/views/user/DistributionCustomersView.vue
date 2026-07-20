<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px]">
      <TablePageLayout>
        <template #actions>
          <header class="flex min-w-0 items-center justify-between gap-3">
            <div class="min-w-0"><h1 class="page-title">我的客户</h1><p class="page-description">直属客户及其付费贡献</p></div>
            <button class="btn btn-secondary btn-icon shrink-0" title="刷新" aria-label="刷新" :disabled="loading" @click="load"><Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" /></button>
          </header>
          <DistributionNav class="mt-5" />
        </template>

        <template #filters>
          <div class="relative min-w-0 sm:max-w-xl">
            <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
            <input v-model="search" class="input pl-9" type="search" placeholder="搜索邮箱或用户名" aria-label="搜索代理客户" />
          </div>
        </template>

        <template #table>
          <DataTable :columns="columns" :data="items" :loading="loading" row-key="user_id">
            <template #cell-customer="{ row }">
              <div class="min-w-44 max-w-56"><p class="truncate font-medium">{{ row.email || '未设置邮箱' }}</p><p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ row.username || '未设置用户名' }}</p></div>
            </template>
            <template #cell-contribution="{ row }">
              <div class="min-w-36"><p class="font-medium tabular-nums">实付 {{ money(row.total_paid_cny) }}</p><p class="mt-0.5 text-xs text-foreground-subtle">{{ row.order_count }} 笔计佣订单</p></div>
            </template>
            <template #cell-commission="{ row }"><span class="font-semibold tabular-nums text-success-foreground">{{ money(row.commission_cny) }}</span></template>
            <template #cell-bound_at="{ row }"><time class="whitespace-nowrap text-sm text-foreground-muted">{{ formatDate(row.bound_at) }}</time></template>
            <template #mobile-card="{ row }">
              <div class="space-y-3">
                <div class="min-w-0"><p class="truncate font-medium">{{ row.email || '未设置邮箱' }}</p><p class="mt-0.5 truncate text-xs text-foreground-subtle">{{ row.username || '未设置用户名' }}</p></div>
                <dl class="grid grid-cols-3 gap-3 border-y border-outline py-3 text-xs">
                  <div><dt class="text-foreground-subtle">累计实付</dt><dd class="mt-1 font-semibold">{{ money(row.total_paid_cny) }}</dd></div>
                  <div><dt class="text-foreground-subtle">订单</dt><dd class="mt-1 font-semibold">{{ row.order_count }}</dd></div>
                  <div><dt class="text-foreground-subtle">贡献佣金</dt><dd class="mt-1 font-semibold text-success-foreground">{{ money(row.commission_cny) }}</dd></div>
                </dl>
                <time class="block text-xs text-foreground-muted">绑定于 {{ formatDate(row.bound_at) }}</time>
              </div>
            </template>
            <template #empty><div class="py-12 text-center"><p class="font-medium">暂无客户</p><p class="mt-1 text-sm text-foreground-subtle">分享推广链接后，新注册客户会显示在这里</p><RouterLink to="/distribution/promotion" class="btn btn-secondary mt-4"><Icon name="link" size="sm" />前往推广</RouterLink></div></template>
          </DataTable>
        </template>

        <template #pagination>
          <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="changePage" @update:pageSize="changePageSize" />
        </template>
      </TablePageLayout>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import DistributionNav from '@/components/distribution/DistributionNav.vue'
import { listCustomers, type DistributionCustomer } from '@/api/distribution'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

const app = useAppStore()
const items = ref<DistributionCustomer[]>([])
const loading = ref(false)
const search = ref('')
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 })
let searchTimer: number | null = null
const columns: Column[] = [
  { key: 'customer', label: '客户' },
  { key: 'contribution', label: '付费贡献' },
  { key: 'commission', label: '贡献佣金' },
  { key: 'bound_at', label: '绑定时间' },
]
const money = (value: string) => `¥${Number(value || 0).toFixed(2)}`
const formatDate = (value: string) => new Date(value).toLocaleString()

async function load() {
  loading.value = true
  try {
    const result = await listCustomers({ page: pagination.value.page, page_size: pagination.value.page_size, search: search.value.trim() || undefined })
    items.value = result.items
    pagination.value = { page: result.page, page_size: result.page_size, total: result.total, pages: result.pages }
  } catch (error) {
    app.showError(extractApiErrorMessage(error, '加载客户失败'))
  } finally { loading.value = false }
}
function scheduleLoad() { pagination.value.page = 1; if (searchTimer) clearTimeout(searchTimer); searchTimer = window.setTimeout(load, 300) }
function changePage(page: number) { pagination.value.page = page; void load() }
function changePageSize(pageSize: number) { pagination.value.page = 1; pagination.value.page_size = pageSize; void load() }
watch(search, scheduleLoad)
onMounted(load)
</script>
