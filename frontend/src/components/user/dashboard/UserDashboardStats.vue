<template>
  <section class="dashboard-summary" :aria-label="t('dashboard.title')">
    <h2 class="sr-only">{{ t('dashboard.title') }}</h2>

    <div :class="['dashboard-metric-grid dashboard-metric-grid-core', { 'dashboard-metric-grid-simple': isSimple }]">
      <article v-if="!isSimple" class="dashboard-metric">
        <span class="dashboard-metric-icon dashboard-metric-icon-success" aria-hidden="true">
          <Icon name="dollar" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.balance') }}</p>
          <p class="dashboard-metric-value text-success-foreground">${{ formatBalance(balance) }}</p>
          <p class="dashboard-metric-meta">{{ t('common.available') }}</p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="key" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.apiKeys') }}</p>
          <p class="dashboard-metric-value">{{ stats?.total_api_keys || 0 }}</p>
          <p class="dashboard-metric-meta text-success-foreground">
            {{ stats?.active_api_keys || 0 }} {{ t('common.active') }}
          </p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="chart" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.todayRequests') }}</p>
          <p class="dashboard-metric-value">{{ formatNumber(stats?.today_requests || 0) }}</p>
          <p class="dashboard-metric-meta">
            {{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}
          </p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="dollar" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.todayCost') }}</p>
          <p class="dashboard-metric-value dashboard-cost-value">
            <span :title="t('dashboard.actual')">${{ formatCost(stats?.today_actual_cost || 0) }}</span>
            <span class="dashboard-cost-standard" :title="t('dashboard.standard')"> / ${{ formatCost(stats?.today_cost || 0) }}</span>
          </p>
          <p class="dashboard-metric-meta">
            {{ t('common.total') }}:
            <span :title="t('dashboard.actual')">${{ formatCost(stats?.total_actual_cost || 0) }}</span>
            <span class="text-foreground-subtle" :title="t('dashboard.standard')"> / ${{ formatCost(stats?.total_cost || 0) }}</span>
          </p>
        </div>
      </article>
    </div>

    <div class="dashboard-metric-grid dashboard-metric-grid-secondary">
      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="cube" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.todayTokens') }}</p>
          <p class="dashboard-metric-value">{{ formatTokens(stats?.today_tokens || 0) }}</p>
          <p class="dashboard-metric-meta">
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}
          </p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="database" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.totalTokens') }}</p>
          <p class="dashboard-metric-value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
          <p class="dashboard-metric-meta">
            {{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
            {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }}
          </p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="bolt" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.performance') }}</p>
          <p class="dashboard-metric-value">{{ formatTokens(stats?.rpm || 0) }} <span class="dashboard-unit">RPM</span></p>
          <p class="dashboard-metric-meta">{{ formatTokens(stats?.tpm || 0) }} TPM</p>
        </div>
      </article>

      <article class="dashboard-metric">
        <span class="dashboard-metric-icon" aria-hidden="true">
          <Icon name="clock" size="md" :stroke-width="2" />
        </span>
        <div class="min-w-0">
          <p class="dashboard-metric-label">{{ t('dashboard.avgResponse') }}</p>
          <p class="dashboard-metric-value">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
          <p class="dashboard-metric-meta">{{ t('dashboard.averageTime') }}</p>
        </div>
      </article>
    </div>

    <div v-if="!isSimple && platformCards.length > 0" class="dashboard-platforms">
      <header class="dashboard-platforms-header">
        <h3>{{ t('dashboard.platformBreakdown') }}</h3>
        <span>{{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}</span>
      </header>

      <div class="dashboard-platform-list">
        <article
          v-for="item in platformCards"
          :key="item.platform"
          :class="['dashboard-platform-row', { 'dashboard-platform-row-other': item.isOther }]"
        >
          <div class="dashboard-platform-name">
            <span>{{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}</span>
            <strong :title="t('dashboard.actual')">${{ formatCost(item.total_actual_cost) }}</strong>
          </div>

          <dl class="dashboard-platform-data">
            <div>
              <dt>{{ t('dashboard.todayCost') }}</dt>
              <dd>${{ formatCost(item.today_actual_cost) }}</dd>
            </div>
            <div>
              <dt>{{ t('dashboard.requests') }}</dt>
              <dd>{{ item.total_requests > 0 ? formatNumber(item.total_requests) : '-' }}</dd>
            </div>
            <div>
              <dt>{{ t('dashboard.tokens') }}</dt>
              <dd>{{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}</dd>
            </div>
          </dl>

          <div v-if="hasAnyLimit(item.quota) && !item.isOther" class="dashboard-quota">
            <p class="dashboard-quota-title">{{ t('dashboard.platformQuota.title') }}</p>
            <template v-for="w in (['daily', 'weekly', 'monthly'] as const)" :key="w">
              <div v-if="quotaVal(item.quota, `${w}_limit_usd`) != null" class="dashboard-quota-window">
                <template v-if="(quotaVal(item.quota, `${w}_limit_usd`) as number) === 0">
                  <div class="dashboard-quota-labels">
                    <span>{{ t(`dashboard.platformQuota.${w}`) }}</span>
                    <span class="text-danger-foreground">{{ t('dashboard.platformQuota.disabled') }}</span>
                  </div>
                  <div class="dashboard-quota-track" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow="100" :aria-label="t(`dashboard.platformQuota.${w}`)">
                    <div class="h-full w-full bg-danger" />
                  </div>
                </template>
                <template v-else>
                  <div class="dashboard-quota-labels">
                    <span>{{ t(`dashboard.platformQuota.${w}`) }}</span>
                    <span>
                      ${{ formatUsd((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0) }} /
                      ${{ formatUsd(quotaVal(item.quota, `${w}_limit_usd`) as number) }}
                    </span>
                  </div>
                  <div
                    class="dashboard-quota-track"
                    role="progressbar"
                    aria-valuemin="0"
                    aria-valuemax="100"
                    :aria-valuenow="calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number)"
                    :aria-label="t(`dashboard.platformQuota.${w}`)"
                  >
                    <div
                      class="h-full transition-[width] motion-reduce:transition-none"
                      :class="quotaBarClass(calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number))"
                      :style="{ width: calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number) + '%' }"
                    />
                  </div>
                  <p v-if="quotaVal(item.quota, `${w}_window_resets_at`)" class="dashboard-quota-reset">
                    {{ t('dashboard.platformQuota.resetsAt', { time: formatResetTime(quotaVal(item.quota, `${w}_window_resets_at`) as string) }) }}
                  </p>
                </template>
              </div>
            </template>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { PlatformQuotaItem } from '@/types'

interface FusedPlatformCard {
  platform: string
  total_actual_cost: number
  today_actual_cost: number
  total_requests: number
  total_tokens: number
  isOther?: boolean
  quota?: PlatformQuotaItem
}

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
  platformQuotas?: PlatformQuotaItem[] | null
}>()
const { t } = useI18n()

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity'
}

const platformLabel = (p: string) => PLATFORM_LABELS[p] ?? p

const sortedPlatforms = computed(() => {
  const list = props.stats?.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

// 处理"各平台之和 < 总值"的差值：后端按平台聚合时过滤了无法归属平台的行
// （group 与 account 都缺 platform）。这里把差值作为"其他"卡片显式展示，
// 避免 Row 1 总值与 Row 3 平台拆分加总对不上、用户困惑。
const OTHER_THRESHOLD = 0.0001
const platformCards = computed<FusedPlatformCard[]>(() => {
  // 建立 by_platform Map
  const byPlat = new Map<string, (typeof sortedPlatforms.value)[number]>()
  for (const item of props.stats?.by_platform ?? []) byPlat.set(item.platform, item)

  // 建立 quota Map
  const byQuota = new Map<string, PlatformQuotaItem>()
  for (const q of props.platformQuotas ?? []) byQuota.set(q.platform, q)

  // union 平台集合。后端 by_platform / quota 接口均不会返回 platform='__other__'，
  // 无需显式排除；__other__ 由下方差值补差逻辑单独追加。
  const platforms = new Set<string>([...byPlat.keys(), ...byQuota.keys()])

  const PLATFORM_ORDER = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
  const cards: FusedPlatformCard[] = []

  for (const p of platforms) {
    const stat = byPlat.get(p)
    cards.push({
      platform: p,
      total_actual_cost: stat?.total_actual_cost ?? 0,
      today_actual_cost: stat?.today_actual_cost ?? 0,
      total_requests: stat?.total_requests ?? 0,
      total_tokens: stat?.total_tokens ?? 0,
      quota: byQuota.get(p),
    })
  }

  // 排序：按 PLATFORM_ORDER，未知平台按名称排序
  cards.sort((a, b) => {
    const ai = PLATFORM_ORDER.indexOf(a.platform)
    const bi = PLATFORM_ORDER.indexOf(b.platform)
    if (ai === -1 && bi === -1) return a.platform.localeCompare(b.platform)
    if (ai === -1) return 1
    if (bi === -1) return -1
    return ai - bi
  })

  // __other__ 补差逻辑：只对 by_platform 有 usage 数据的总和计算
  const total = props.stats?.total_actual_cost ?? 0
  const today = props.stats?.today_actual_cost ?? 0
  const sumTotal = cards.reduce((s, c) => s + c.total_actual_cost, 0)
  const sumToday = cards.reduce((s, c) => s + c.today_actual_cost, 0)
  const diffTotal = Math.max(0, total - sumTotal)
  const diffToday = Math.max(0, today - sumToday)

  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    cards.push({
      platform: '__other__',
      total_actual_cost: diffTotal,
      today_actual_cost: diffToday,
      total_requests: 0,
      total_tokens: 0,
      isOther: true,
    })
  }

  return cards
})

// Quota helpers

type QuotaWindow = 'daily' | 'weekly' | 'monthly'
type QuotaField = `${QuotaWindow}_limit_usd` | `${QuotaWindow}_usage_usd` | `${QuotaWindow}_window_resets_at`

function quotaVal(q: PlatformQuotaItem | undefined, key: QuotaField): PlatformQuotaItem[QuotaField] {
  return q?.[key]
}

function hasAnyLimit(q: PlatformQuotaItem | undefined): boolean {
  if (!q) return false
  return q.daily_limit_usd != null || q.weekly_limit_usd != null || q.monthly_limit_usd != null
}

function calcPercent(usage: number, limit: number): number {
  if (!limit || limit <= 0) return 0
  return Math.min(100, Math.max(0, Math.round((usage / limit) * 100)))
}

function quotaBarClass(p: number): string {
  if (p >= 95) return 'bg-red-500'
  if (p >= 75) return 'bg-amber-500'
  return 'bg-green-500'
}

// 与 formatBalance 一致使用 Intl.NumberFormat 做半偶舍入，避免 toFixed 在不同 JS 引擎
// 下偶发截断而非四舍五入（与后端展示精度不一致）。
const usdFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})
function formatUsd(n: number): string {
  if (!Number.isFinite(n)) return '0.00'
  return usdFormatter.format(n)
}

function formatResetTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>

<style scoped>
.dashboard-summary {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--ui-border);
  border-radius: var(--radius-md);
  background: var(--ui-surface);
  box-shadow: var(--ui-shadow-xs);
}

.dashboard-metric-grid {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1px;
  background: var(--ui-border);
}

.dashboard-metric-grid + .dashboard-metric-grid {
  border-top: 1px solid var(--ui-border);
}

.dashboard-metric-grid-simple .dashboard-metric:last-child {
  grid-column: 1 / -1;
}

.dashboard-metric {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.875rem;
  background: var(--ui-surface);
}

.dashboard-metric-icon {
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--ui-border);
  border-radius: var(--radius-sm);
  background: var(--ui-surface-subtle);
  color: var(--ui-text-muted);
}

.dashboard-metric-icon-success {
  border-color: rgb(var(--color-success) / 0.2);
  background: rgb(var(--color-success-subtle));
  color: rgb(var(--color-success-foreground));
}

.dashboard-metric-label,
.dashboard-metric-meta {
  overflow-wrap: anywhere;
  color: var(--ui-text-subtle);
  font-size: 0.75rem;
  line-height: 1.25rem;
}

.dashboard-metric-label {
  font-weight: 500;
}

.dashboard-metric-value {
  margin-top: 0.125rem;
  overflow-wrap: anywhere;
  color: var(--ui-text);
  font-size: 1.25rem;
  font-weight: 650;
  font-variant-numeric: tabular-nums;
  line-height: 1.75rem;
}

.dashboard-cost-value {
  font-size: 1rem;
}

.dashboard-cost-standard,
.dashboard-unit {
  color: var(--ui-text-subtle);
  font-size: 0.75rem;
  font-weight: 400;
}

.dashboard-platforms {
  border-top: 1px solid var(--ui-border);
}

.dashboard-platforms-header {
  display: flex;
  min-height: 3.25rem;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--ui-surface-subtle);
}

.dashboard-platforms-header h3 {
  color: var(--ui-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.dashboard-platforms-header span {
  color: var(--ui-text-subtle);
  font-size: 0.75rem;
}

.dashboard-platform-row {
  display: grid;
  min-width: 0;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid var(--ui-border);
}

.dashboard-platform-row:first-child {
  border-top: 0;
}

.dashboard-platform-row-other {
  background: var(--ui-surface-subtle);
}

.dashboard-platform-name {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  color: var(--ui-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.dashboard-platform-name span {
  overflow-wrap: anywhere;
}

.dashboard-platform-name strong {
  flex-shrink: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.8125rem;
  font-variant-numeric: tabular-nums;
}

.dashboard-platform-data {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.dashboard-platform-data div {
  min-width: 0;
}

.dashboard-platform-data dt,
.dashboard-quota-title,
.dashboard-quota-reset {
  color: var(--ui-text-subtle);
  font-size: 0.6875rem;
  line-height: 1rem;
}

.dashboard-platform-data dd {
  margin-top: 0.125rem;
  overflow-wrap: anywhere;
  color: var(--ui-text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
}

.dashboard-quota {
  min-width: 0;
}

.dashboard-quota-title {
  margin-bottom: 0.375rem;
  font-weight: 600;
}

.dashboard-quota-window + .dashboard-quota-window {
  margin-top: 0.5rem;
}

.dashboard-quota-labels {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--ui-text-muted);
  font-size: 0.6875rem;
  font-variant-numeric: tabular-nums;
  line-height: 1rem;
}

.dashboard-quota-track {
  height: 0.375rem;
  margin-top: 0.25rem;
  overflow: hidden;
  border-radius: 999px;
  background: var(--ui-surface-subtle);
}

.dashboard-quota-reset {
  margin-top: 0.125rem;
}

@media (min-width: 640px) {
  .dashboard-metric-grid-core.dashboard-metric-grid-simple {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .dashboard-metric-grid-simple .dashboard-metric:last-child {
    grid-column: auto;
  }
}

@media (min-width: 768px) {
  .dashboard-metric {
    padding: 1rem;
  }

  .dashboard-platform-row {
    grid-template-columns: minmax(9rem, 0.75fr) minmax(0, 1.25fr);
    align-items: start;
  }

  .dashboard-quota {
    grid-column: 1 / -1;
  }
}

@media (min-width: 1024px) {
  .dashboard-metric-grid-core,
  .dashboard-metric-grid-secondary {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .dashboard-metric-grid-core.dashboard-metric-grid-simple {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .dashboard-platform-row {
    grid-template-columns: minmax(9rem, 0.7fr) minmax(18rem, 1.2fr) minmax(18rem, 1.5fr);
  }

  .dashboard-quota {
    grid-column: auto;
  }
}
</style>
