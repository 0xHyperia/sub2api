<template>
  <section class="rounded-lg border border-outline bg-surface px-3 py-2 shadow-card">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <span
        class="inline-flex min-h-8 items-center rounded-md px-2.5 text-xs font-semibold"
        :class="overallChipClass"
      >
        <span class="mr-1.5 h-1.5 w-1.5 rounded-full" :class="overallDotClass"></span>
        {{ overallLabel }}
      </span>

      <div class="flex min-w-0 flex-wrap items-center justify-end gap-2">
      <div
        role="tablist"
        :aria-label="t('channelStatus.detailColumns.availability7d')"
        class="inline-flex max-w-full overflow-x-auto rounded-lg border border-outline bg-surface-subtle p-0.5 text-xs"
      >
        <button
          v-for="opt in windowOptions"
          :key="opt.value"
          :id="`monitor-window-${opt.value}`"
          type="button"
          role="tab"
          :aria-selected="window === opt.value"
          :tabindex="window === opt.value ? 0 : -1"
          aria-controls="monitor-grid"
          class="min-h-8 whitespace-nowrap rounded-md px-3 py-1 transition-colors"
          :class="window === opt.value
            ? 'bg-surface text-foreground shadow-card font-semibold'
            : 'text-foreground-muted hover:text-foreground'"
          @click="emit('update:window', opt.value)"
          @keydown="handleWindowKeydown($event, opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>

      <button
        type="button"
        class="inline-flex h-9 w-9 items-center justify-center rounded-md text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground disabled:opacity-50"
        :disabled="loading"
        :title="t('common.refresh')"
        :aria-label="t('common.refresh')"
        @click="emit('refresh')"
      >
        <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
      </button>

      <AutoRefreshButton
        v-if="autoRefresh"
        :enabled="autoRefresh.enabled.value"
        :interval-seconds="autoRefresh.intervalSeconds.value"
        :countdown="autoRefresh.countdown.value"
        :intervals="autoRefresh.intervals"
        @update:enabled="autoRefresh.setEnabled"
        @update:interval="autoRefresh.setInterval"
      />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AutoRefreshButton from '@/components/common/AutoRefreshButton.vue'
export type MonitorWindow = '7d' | '15d' | '30d'
export type OverallStatus = 'operational' | 'degraded'

const props = defineProps<{
  overallStatus: OverallStatus
  intervalSeconds: number
  window: MonitorWindow
  loading: boolean
  autoRefresh?: {
    enabled: { value: boolean }
    intervalSeconds: { value: number }
    countdown: { value: number }
    intervals: readonly number[]
    setEnabled: (v: boolean) => void
    setInterval: (v: number) => void
  }
}>()

const emit = defineEmits<{
  (e: 'update:window', value: MonitorWindow): void
  (e: 'refresh'): void
}>()

const { t } = useI18n()

const windowOptions = computed<{ value: MonitorWindow; label: string }[]>(() => [
  { value: '7d', label: t('channelStatus.windowTab.7d') },
  { value: '15d', label: t('channelStatus.windowTab.15d') },
  { value: '30d', label: t('channelStatus.windowTab.30d') },
])

const overallLabel = computed(() => t(`channelStatus.overall.${props.overallStatus}`))

const overallChipClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300'
    case 'degraded':
    default:
      return 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300'
  }
})

const overallDotClass = computed(() => {
  switch (props.overallStatus) {
    case 'operational':
      return 'bg-emerald-500'
    case 'degraded':
    default:
      return 'bg-amber-500'
  }
})

const handleWindowKeydown = (event: KeyboardEvent, current: MonitorWindow) => {
  const values = windowOptions.value.map((option) => option.value)
  const index = values.indexOf(current)
  let nextIndex: number | null = null

  if (event.key === 'ArrowLeft') nextIndex = (index - 1 + values.length) % values.length
  else if (event.key === 'ArrowRight') nextIndex = (index + 1) % values.length
  else if (event.key === 'Home') nextIndex = 0
  else if (event.key === 'End') nextIndex = values.length - 1
  if (nextIndex === null) return

  event.preventDefault()
  const value = values[nextIndex]
  emit('update:window', value)
  window.requestAnimationFrame(() => document.getElementById(`monitor-window-${value}`)?.focus())
}

</script>
