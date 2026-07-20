<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0">
          <h1 class="page-title">代理中心</h1>
          <p class="page-description">经营数据与佣金概览</p>
        </div>
        <button class="btn btn-secondary btn-icon shrink-0" :disabled="loading" title="刷新" aria-label="刷新" @click="load">
          <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
        </button>
      </header>

      <DistributionNav />

      <div v-if="loading && !overview" class="card flex min-h-48 items-center justify-center">
        <LoadingSpinner />
      </div>

      <template v-else-if="overview">
        <section
          v-if="overview.agent.status !== 'active'"
          class="flex items-start gap-3 rounded-panel border border-warning/40 bg-warning-subtle p-4 text-sm text-warning-foreground"
        >
          <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" />
          <div>
            <p class="font-medium">代理账号{{ overview.agent.status === 'suspended' ? '已暂停' : '已撤销' }}</p>
            <p class="mt-1">推广链接暂时不能绑定新客户，结算和团队操作也不可用。请联系平台管理员处理。</p>
          </div>
        </section>

        <section class="grid grid-cols-2 gap-3 lg:grid-cols-3 xl:grid-cols-6" aria-label="代理经营指标">
          <article v-for="metric in metrics" :key="metric.label" class="card min-w-0 p-4">
            <div class="flex items-center justify-between gap-2">
              <p class="truncate text-sm text-foreground-subtle">{{ metric.label }}</p>
              <Icon :name="metric.icon" size="sm" class="shrink-0 text-foreground-muted" />
            </div>
            <p class="mt-2 truncate text-xl font-semibold tabular-nums" :class="metric.emphasis ? 'text-success-foreground' : ''">
              {{ metric.value }}
            </p>
            <p v-if="metric.hint" class="mt-1 truncate text-xs text-foreground-muted">{{ metric.hint }}</p>
          </article>
        </section>

        <section class="grid gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(280px,0.6fr)]">
          <div class="card p-4 sm:p-5">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="text-base font-semibold">本月经营</h2>
                <p class="mt-1 text-sm text-foreground-subtle">客户充值与新增关系</p>
              </div>
              <span class="badge" :class="statusClass">{{ statusText }}</span>
            </div>
            <dl class="mt-5 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <div><dt class="text-xs text-foreground-subtle">本月客户实付</dt><dd class="mt-1 font-semibold tabular-nums">{{ money(overview.this_month_customer_paid_cny) }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">本月新增客户</dt><dd class="mt-1 font-semibold tabular-nums">{{ overview.new_customers_this_month }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">付费客户</dt><dd class="mt-1 font-semibold tabular-nums">{{ overview.paying_customer_count }}</dd></div>
              <div><dt class="text-xs text-foreground-subtle">下级代理</dt><dd class="mt-1 font-semibold tabular-nums">{{ overview.team_count }}</dd></div>
            </dl>
          </div>

          <div class="card p-4 sm:p-5">
            <h2 class="text-base font-semibold">快捷操作</h2>
            <nav class="mt-3 divide-y divide-outline" aria-label="代理快捷操作">
              <RouterLink v-for="action in quickActions" :key="action.path" :to="action.path" class="flex items-center justify-between gap-3 py-3 text-sm hover:text-brand">
                <span class="flex items-center gap-2"><Icon :name="action.icon" size="sm" />{{ action.label }}</span>
                <Icon name="chevronRight" size="sm" class="text-foreground-muted" />
              </RouterLink>
            </nav>
          </div>
        </section>

        <section v-if="Number(overview.agent.debt_cny) > 0" class="rounded-panel border border-danger/40 bg-danger/5 p-4 text-sm text-danger-foreground">
          当前退款负债 {{ money(overview.agent.debt_cny) }}，后续佣金会优先抵扣。
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DistributionNav from '@/components/distribution/DistributionNav.vue'
import { getDistributionOverview, type DistributionOverview } from '@/api/distribution'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const router = useRouter()
const overview = ref<DistributionOverview | null>(null)
const loading = ref(false)

const money = (value: string) => `¥${Number(value || 0).toFixed(2)}`
const statusText = computed(() => ({ active: '正常经营', suspended: '已暂停', revoked: '已撤销' })[overview.value?.agent.status || 'active'])
const statusClass = computed(() => overview.value?.agent.status === 'active' ? 'badge-success' : overview.value?.agent.status === 'suspended' ? 'badge-warning' : 'badge-gray')
const metrics = computed(() => overview.value ? [
  { label: '可用佣金', value: money(overview.value.agent.available_cny), hint: '可结算', icon: 'creditCard' as const, emphasis: true },
  { label: '冻结佣金', value: money(overview.value.agent.frozen_cny), hint: '等待释放', icon: 'clock' as const },
  { label: '本月佣金', value: money(overview.value.this_month_commission_cny), hint: '含冻结佣金', icon: 'calendar' as const },
  { label: '累计佣金', value: money(overview.value.agent.total_earned_cny), hint: '扣除冲正', icon: 'gift' as const },
  { label: '累计实付', value: money(overview.value.customer_paid_cny), hint: '计佣客户订单', icon: 'creditCard' as const },
  { label: '直属客户', value: String(overview.value.customer_count), hint: `${overview.value.paying_customer_count} 位已付费`, icon: 'users' as const },
] : [])
const quickActions = computed(() => [
  { path: '/distribution/promotion', label: '分享推广链接', icon: 'link' as const },
  { path: '/distribution/customers', label: '查看客户贡献', icon: 'users' as const },
  { path: '/distribution/withdrawals', label: '结算可用佣金', icon: 'creditCard' as const },
])

async function load() {
  loading.value = true
  try {
    overview.value = await getDistributionOverview()
  } catch {
    appStore.showError('当前账号没有代理权限')
    void router.replace('/dashboard')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
