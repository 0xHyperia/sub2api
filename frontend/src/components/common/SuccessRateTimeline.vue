<template>
  <div class="inline-flex min-w-0 items-center gap-2" :class="variant === 'availability' ? 'w-full' : ''">
    <div
      class="flex shrink-0 items-end"
      :class="variant === 'availability'
        ? 'h-7 w-full gap-[3px] rounded-control border border-outline bg-surface px-1.5 py-1.5 shadow-inner'
        : 'h-3.5 w-[119px] gap-px'"
      role="img"
      :aria-label="ariaLabel"
    >
      <span
        v-for="(bucket, index) in normalizedBuckets"
        :key="`${bucket.started_at}-${index}`"
        class="flex h-full items-end rounded-sm"
        :class="variant === 'availability' ? 'min-w-0 flex-1' : 'w-[3px]'"
      >
        <button
          type="button"
          class="block w-full rounded-sm outline-none transition-[filter,opacity] hover:brightness-110 focus-visible:ring-2 focus-visible:ring-focus"
          :class="barClass(bucket.success_rate)"
          :style="barStyle(bucket.success_rate)"
          :aria-label="bucketTitle(bucket)"
          @mouseenter="showTooltip($event, bucket)"
          @mouseleave="hideTooltip"
          @focus="showTooltip($event, bucket)"
          @blur="hideTooltip"
        ></button>
      </span>
    </div>
    <span v-if="showOverall" class="w-12 shrink-0 text-right font-mono text-xs font-semibold tabular-nums" :class="textClass">
      {{ formattedOverall }}
    </span>
    <Teleport to="body">
      <Transition name="success-rate-tooltip">
        <div
          v-if="tooltip"
          role="tooltip"
          class="pointer-events-none fixed z-[100] flex w-52 -translate-x-1/2 -translate-y-full items-center justify-between gap-3 rounded-control bg-inverse px-3 py-2 font-mono text-[11px] font-semibold tabular-nums text-inverse-foreground shadow-floating ring-1 ring-inverse-foreground/10"
          :style="{ left: `${tooltip.left}px`, top: `${tooltip.top}px` }"
        >
          <span class="whitespace-nowrap">{{ tooltip.time }}</span>
          <span class="whitespace-nowrap">{{ tooltip.rate }}</span>
          <span class="absolute top-full h-2 w-2 -translate-x-1/2 -translate-y-1 rotate-45 bg-inverse" :style="{ left: `${tooltip.arrowLeft}px` }"></span>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

export interface SuccessRateBucket {
  started_at: string
  success_rate: number | null
}

const props = withDefaults(defineProps<{
  buckets?: SuccessRateBucket[] | null
  successRate?: number | null
  resolution?: 'minute' | 'hour'
  variant?: 'rate' | 'availability'
  showOverall?: boolean
}>(), {
  buckets: () => [],
  successRate: null,
  resolution: 'hour',
  variant: 'rate',
  showOverall: true,
})
const tooltip = ref<{ left: number; top: number; arrowLeft: number; time: string; rate: string } | null>(null)

const normalizedBuckets = computed<SuccessRateBucket[]>(() => {
  const step = props.resolution === 'minute' ? 60_000 : 3_600_000
  const values = props.buckets ?? []
  const latestValue = values.reduce((latest, bucket) => Math.max(latest, new Date(bucket.started_at).getTime()), 0)
  const anchorValue = latestValue || Date.now()
  const anchor = Math.floor(anchorValue / step) * step
  const byTime = new Map(values.map(bucket => [Math.floor(new Date(bucket.started_at).getTime() / step) * step, bucket]))
  return Array.from({ length: 30 }, (_, index) => {
    const startedAt = anchor - ((29 - index) * step)
    return byTime.get(startedAt) ?? { started_at: new Date(startedAt).toISOString(), success_rate: null }
  })
})

const formattedOverall = computed(() => {
  if (props.successRate == null || !Number.isFinite(props.successRate)) return '—'
  return `${props.successRate.toFixed(1)}%`
})

const textClass = computed(() => rateClass(props.successRate, true))
const ariaLabel = computed(() => `Success rate ${formattedOverall.value}`)

function rateClass(rate: number | null | undefined, text = false): string {
  if (rate == null || !Number.isFinite(rate)) return text ? 'text-foreground-subtle' : 'bg-outline-strong'
  if (props.variant === 'availability') {
    if (rate >= 99.9) return text ? 'text-success-foreground' : 'bg-success'
    return rate >= 70 ? (text ? 'text-success-foreground' : 'bg-success/70') : (text ? 'text-danger-foreground' : 'bg-danger')
  }
  if (rate >= 99.9) return text ? 'text-success-foreground' : 'bg-success'
  if (rate >= 90) return text ? 'text-success-foreground' : 'bg-success/70'
  if (rate >= 70) return text ? 'text-warning-foreground' : 'bg-warning'
  return text ? 'text-danger-foreground' : 'bg-danger'
}

function barClass(rate: number | null): string {
  return rateClass(rate)
}

function barStyle(rate: number | null): Record<string, string> {
  if (props.variant === 'availability') return { height: '100%' }
  if (rate == null || !Number.isFinite(rate)) return { height: '40%' }
  if (rate >= 99.9) return { height: '100%' }
  if (rate >= 99) return { height: '88%' }
  if (rate >= 95) return { height: '72%' }
  if (rate >= 90) return { height: '55%' }
  if (rate >= 70) return { height: '48%' }
  return { height: '40%' }
}

function bucketLabel(bucket: SuccessRateBucket): { time: string; rate: string } {
  const locale = typeof document === 'undefined' ? undefined : document.documentElement.lang || navigator.language
  const startedAt = new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(new Date(bucket.started_at))
  const rate = bucket.success_rate == null ? '—' : `${bucket.success_rate.toFixed(2)}%`
  return { time: startedAt, rate }
}

function bucketTitle(bucket: SuccessRateBucket): string {
  const label = bucketLabel(bucket)
  return `${label.time} ${label.rate}`
}

function showTooltip(event: MouseEvent | FocusEvent, bucket: SuccessRateBucket) {
  const target = event.currentTarget
  if (!(target instanceof HTMLElement)) return
  const rect = target.getBoundingClientRect()
  const tooltipWidth = 208
  const viewportPadding = 12
  const anchorX = rect.left + (rect.width / 2)
  const left = Math.min(window.innerWidth - (tooltipWidth / 2) - viewportPadding, Math.max((tooltipWidth / 2) + viewportPadding, anchorX))
  const label = bucketLabel(bucket)
  tooltip.value = {
    left,
    top: rect.top - 8,
    arrowLeft: Math.min(tooltipWidth - 12, Math.max(12, anchorX - left + (tooltipWidth / 2))),
    ...label,
  }
}

function hideTooltip() {
  tooltip.value = null
}
</script>

<style scoped>
.success-rate-tooltip-enter-active,
.success-rate-tooltip-leave-active { transition: opacity 120ms ease, transform 120ms ease; }
.success-rate-tooltip-enter-from,
.success-rate-tooltip-leave-to { opacity: 0; transform: translate(-50%, calc(-100% + 4px)); }
</style>
