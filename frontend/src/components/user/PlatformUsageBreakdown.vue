<template>
  <div
    class="group/usage relative min-w-0 text-sm"
    @mouseenter="tooltipOpen = true"
    @mouseleave="tooltipOpen = false"
  >
    <div class="flex items-center gap-1.5">
      <span class="text-foreground-muted">{{ t('admin.users.today') }}:</span>
      <span class="break-all font-medium text-foreground">${{ today.toFixed(4) }}</span>
      <button
        v-if="hasBreakdown"
        type="button"
        class="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-control text-foreground-subtle hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
        :aria-label="t('admin.users.platformBreakdown')"
        :aria-describedby="tooltipId"
        :aria-expanded="tooltipOpen"
        @focus="tooltipOpen = true"
        @blur="tooltipOpen = false"
        @keydown.esc.stop="tooltipOpen = false"
      >
        <Icon name="infoCircle" size="xs" aria-hidden="true" />
      </button>
    </div>
    <div class="mt-0.5 flex items-center gap-1.5">
      <span class="text-foreground-muted">{{ t('admin.users.total') }}:</span>
      <span class="break-all font-medium text-foreground">${{ total.toFixed(4) }}</span>
    </div>

    <div
      v-if="hasBreakdown"
      :id="tooltipId"
      role="tooltip"
      :class="[
        'pointer-events-none absolute right-0 top-full z-50 mt-2 w-[min(18rem,calc(100vw-2rem))] rounded-panel border border-outline bg-surface-raised px-3 py-2 text-xs text-foreground shadow-lg transition-opacity duration-100',
        tooltipOpen ? 'visible opacity-100' : 'invisible opacity-0'
      ]"
    >
      <div class="mb-1.5 flex flex-wrap items-center justify-between gap-x-3 gap-y-1 border-b border-outline pb-1 text-[11px] text-foreground-muted">
        <span>{{ t('admin.users.platformBreakdown') }}</span>
        <span class="font-mono">{{ t('admin.users.today') }} / {{ t('admin.users.total') }}</span>
      </div>
      <div
        v-for="item in sortedBreakdown"
        :key="item.platform"
        class="flex min-w-0 items-start justify-between gap-3 py-0.5"
        :class="{ 'opacity-70 italic': item.isOther }"
      >
        <span class="min-w-0 break-words capitalize">
          {{ item.isOther ? t('admin.users.platformOther') : platformLabel(item.platform) }}
        </span>
        <span class="shrink-0 text-right font-mono">
          ${{ item.today_actual_cost.toFixed(4) }}
          <span class="opacity-50">/</span>
          ${{ item.total_actual_cost.toFixed(4) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { PlatformUsage } from '@/api/admin/dashboard'

const props = defineProps<{
  today: number
  total: number
  byPlatform?: PlatformUsage[]
}>()

const { t } = useI18n()
const tooltipId = `platform-usage-breakdown-${getCurrentInstance()?.uid ?? 0}`
const tooltipOpen = ref(false)

// 与 UserDashboardStats 保持一致：把"总值 - 各平台之和"的差作为"其他"行展示，
// 避免 tooltip 内各平台费用加总与列首总值对不上。
const OTHER_THRESHOLD = 0.0001

interface BreakdownRow {
  platform: string
  today_actual_cost: number
  total_actual_cost: number
  isOther?: boolean
}

const sortedBreakdown = computed<BreakdownRow[]>(() => {
  const list = props.byPlatform ?? []
  const rows: BreakdownRow[] = [...list]
    .sort((a, b) => b.total_actual_cost - a.total_actual_cost)
    .map((p) => ({ ...p }))

  const sumTotal = rows.reduce((s, r) => s + r.total_actual_cost, 0)
  const sumToday = rows.reduce((s, r) => s + r.today_actual_cost, 0)
  const diffTotal = Math.max(0, props.total - sumTotal)
  const diffToday = Math.max(0, props.today - sumToday)
  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    rows.push({
      platform: '__other__',
      today_actual_cost: diffToday,
      total_actual_cost: diffTotal,
      isOther: true
    })
  }
  return rows
})

const hasBreakdown = computed(() => sortedBreakdown.value.length > 0)

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity'
}

function platformLabel(platform: string): string {
  return PLATFORM_LABELS[platform] ?? platform
}
</script>
