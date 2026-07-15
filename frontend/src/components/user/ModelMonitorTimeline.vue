<template>
  <div class="relative min-w-0">
    <div
      class="group/timeline flex h-7 items-end gap-[3px] rounded-control border border-outline bg-surface px-1.5 py-1.5 shadow-inner"
      role="list"
      :aria-label="t('modelMarketplace.monitor.timeline')"
    >
      <span
        v-for="index in emptyPointCount"
        :key="`empty-${index}`"
        class="h-3 min-w-0 flex-1 rounded-[2px] bg-outline opacity-70 transition-[height,opacity] duration-150 group-hover/timeline:h-3.5 group-hover/timeline:opacity-100"
        aria-hidden="true"
      ></span>

      <div
        v-for="(point, index) in visiblePoints"
        :key="`${point.checked_at}-${index}`"
        class="group/point relative flex h-full min-w-0 flex-1 items-end"
        role="listitem"
      >
        <button
          type="button"
          class="h-3.5 w-full rounded-[2px] opacity-90 outline-none transition-[height,transform,filter,box-shadow,opacity] duration-150 ease-out hover:h-[18px] hover:-translate-y-0.5 hover:opacity-100 hover:brightness-110 hover:shadow-sm focus-visible:h-[18px] focus-visible:-translate-y-0.5 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-1 focus-visible:ring-offset-surface"
          :class="pointClass(point.status)"
          :aria-label="pointAriaLabel(point)"
          :aria-describedby="tooltipId(index)"
          @click.stop
          @keydown.esc.stop="blurPoint"
        ></button>

        <div
          :id="tooltipId(index)"
          role="tooltip"
          class="pointer-events-none invisible absolute bottom-full z-50 mb-3 w-56 translate-y-1 rounded-panel border border-outline-strong bg-surface-raised px-3 py-2.5 text-left opacity-0 shadow-floating transition-[opacity,transform,visibility] duration-150 group-hover/point:visible group-hover/point:translate-y-0 group-hover/point:opacity-100 group-focus-within/point:visible group-focus-within/point:translate-y-0 group-focus-within/point:opacity-100"
          :class="tooltipPositionClass(index)"
        >
          <div class="text-[10px] font-medium text-foreground-subtle">{{ formatCheckedAt(point.checked_at) }}</div>
          <div class="mt-1.5 flex items-end justify-between gap-3">
            <span class="inline-flex min-w-0 items-center gap-2 text-xs font-semibold" :class="statusTextClass(point.status)">
              <span class="h-2 w-2 shrink-0 rounded-full bg-current shadow-sm"></span>
              <span class="truncate">{{ statusLabel(point.status) }}</span>
            </span>
            <span class="shrink-0 text-right">
              <span class="mr-1 text-[9px] text-foreground-subtle">{{ t('modelMarketplace.monitor.latency') }}</span>
              <span class="font-mono text-xs font-semibold tabular-nums text-foreground">{{ latencyLabel(point.latency_ms) }}</span>
            </span>
          </div>
          <span class="absolute top-full h-2 w-2 -translate-y-1 rotate-45 border-b border-r border-outline-strong bg-surface-raised" :class="tooltipArrowClass(index)"></span>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserModelMonitorTimelinePoint } from '@/api/channels'

const props = withDefaults(defineProps<{
  points?: UserModelMonitorTimelinePoint[] | null
  limit?: number
}>(), {
  points: () => [],
  limit: 30,
})

const { t, locale } = useI18n()
const componentId = `model-monitor-timeline-${getCurrentInstance()?.uid ?? 0}`

const visiblePoints = computed(() => [...(props.points ?? [])].slice(0, props.limit).reverse())
const emptyPointCount = computed(() => Math.max(0, props.limit - visiblePoints.value.length))

function tooltipId(index: number) {
  return `${componentId}-point-${index}`
}

function pointClass(status?: string) {
  if (status === 'operational') return 'bg-success'
  if (status === 'degraded') return 'bg-warning'
  if (status === 'failed' || status === 'error') return 'bg-danger'
  return 'bg-surface-emphasis'
}

function statusTextClass(status?: string) {
  if (status === 'operational') return 'text-success-foreground'
  if (status === 'degraded') return 'text-warning-foreground'
  if (status === 'failed' || status === 'error') return 'text-danger-foreground'
  return 'text-foreground-subtle'
}

function statusLabel(status?: string) {
  return status ? t(`modelMarketplace.monitor.${status}`) : t('modelMarketplace.monitor.unknown')
}

function latencyLabel(latency: number | null) {
  return latency == null ? t('modelMarketplace.monitor.noLatency') : `${latency} ms`
}

function formatCheckedAt(value: string) {
  return new Intl.DateTimeFormat(locale.value, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}

function pointAriaLabel(point: UserModelMonitorTimelinePoint) {
  return `${statusLabel(point.status)}, ${latencyLabel(point.latency_ms)}, ${formatCheckedAt(point.checked_at)}`
}

function blurPoint(event: KeyboardEvent) {
  const target = event.currentTarget
  if (target instanceof HTMLButtonElement) target.blur()
}

function tooltipPositionClass(index: number) {
  const slotIndex = emptyPointCount.value + index
  const edgeSize = Math.min(5, Math.floor(props.limit / 2))
  if (slotIndex < edgeSize) return 'left-0'
  if (slotIndex >= props.limit - edgeSize) return 'right-0'
  return 'left-1/2 -translate-x-1/2 group-hover/point:-translate-x-1/2 group-focus-within/point:-translate-x-1/2'
}

function tooltipArrowClass(index: number) {
  const slotIndex = emptyPointCount.value + index
  const edgeSize = Math.min(5, Math.floor(props.limit / 2))
  if (slotIndex < edgeSize) return 'left-2'
  if (slotIndex >= props.limit - edgeSize) return 'right-2'
  return 'left-1/2 -translate-x-1/2'
}
</script>
