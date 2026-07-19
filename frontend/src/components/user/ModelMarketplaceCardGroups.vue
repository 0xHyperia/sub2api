<template>
  <div ref="containerRef" class="marketplace-card-groups relative flex min-w-0 flex-1 items-center">
    <div class="flex min-w-0 items-center gap-1.5 overflow-hidden">
      <button
        v-for="group in visibleGroups"
        :key="group.id"
        type="button"
        class="inline-flex h-6 shrink-0 items-center gap-1 rounded-control border px-1.5 text-[9px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
        :class="group.id === activeGroupId ? 'border-outline bg-brand-subtle font-semibold text-brand' : 'border-outline bg-surface text-foreground-muted hover:bg-surface-subtle hover:text-foreground'"
        :title="group.name"
        :aria-pressed="group.id === activeGroupId"
        @click.stop="emit('select', group.id)"
      >
        <span class="truncate">{{ group.name }}</span>
        <span
          class="shrink-0 rounded-control bg-surface-subtle px-1 font-mono text-[8px] tabular-nums text-foreground-subtle"
          :class="group.id === activeGroupId ? 'bg-surface text-brand' : ''"
        >{{ formatRate(group.effectiveRate) }}x</span>
      </button>
      <button
        v-if="hiddenGroups.length"
        type="button"
        class="inline-flex h-6 shrink-0 items-center rounded-control border border-outline bg-surface px-1.5 font-mono text-[9px] font-semibold tabular-nums text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
        :class="hiddenGroups.some(group => group.id === activeGroupId) ? 'bg-brand-subtle text-brand' : ''"
        :aria-label="t('modelMarketplace.details.moreGroups', { count: hiddenGroups.length })"
        :aria-expanded="showOverflow"
        aria-haspopup="true"
        @click.stop="showOverflow = !showOverflow"
      >+{{ hiddenGroups.length }}</button>
      <span v-if="groups.length === 0" class="text-[9px] text-foreground-subtle">{{ t('modelMarketplace.details.billingGroup') }} -</span>
    </div>

    <div ref="measureRef" class="pointer-events-none absolute invisible flex max-w-full items-center gap-1.5 overflow-hidden whitespace-nowrap" aria-hidden="true">
      <button
        v-for="group in groups"
        :key="group.id"
        data-group-measure
        type="button"
        class="inline-flex h-6 shrink-0 items-center gap-1 rounded-control border px-1.5 text-[9px]"
      >
        <span>{{ group.name }}</span>
        <span class="shrink-0 rounded-control px-1 font-mono text-[8px] tabular-nums">{{ formatRate(group.effectiveRate) }}x</span>
      </button>
      <button ref="measureOverflowRef" type="button" class="inline-flex h-6 shrink-0 items-center rounded-control border px-1.5 font-mono text-[9px] font-semibold tabular-nums">+{{ Math.max(groups.length, 1) }}</button>
    </div>

    <div
      v-if="showOverflow && hiddenGroups.length"
      class="absolute bottom-full left-0 z-30 mb-2 max-h-56 w-max max-w-[calc(100vw-2rem)] overflow-y-auto rounded-panel border border-outline bg-surface-raised p-2 shadow-floating"
      role="group"
      :aria-label="t('modelMarketplace.details.groups')"
    >
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="group in hiddenGroups"
          :key="group.id"
          type="button"
          class="inline-flex h-6 items-center gap-1 rounded-control border px-1.5 text-[9px] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
          :class="group.id === activeGroupId ? 'border-outline bg-brand-subtle font-semibold text-brand' : 'border-outline bg-surface text-foreground-muted hover:bg-surface-subtle hover:text-foreground'"
          @click.stop="selectHiddenGroup(group.id)"
        >
          <span class="max-w-32 truncate">{{ group.name }}</span>
          <span class="shrink-0 rounded-control bg-surface-subtle px-1 font-mono text-[8px] tabular-nums text-foreground-subtle">{{ formatRate(group.effectiveRate) }}x</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MarketplaceGroupOption } from '@/views/user/modelMarketplace'
import { visibleMarketplaceGroupCount } from '@/views/user/modelMarketplace'

const props = defineProps<{
  groups: MarketplaceGroupOption[]
  activeGroupId: number | null | undefined
}>()

const emit = defineEmits<{
  select: [groupId: number]
}>()

const { t } = useI18n()
const containerRef = ref<HTMLElement | null>(null)
const measureRef = ref<HTMLElement | null>(null)
const measureOverflowRef = ref<HTMLElement | null>(null)
const visibleCount = ref(props.groups.length)
const showOverflow = ref(false)
let resizeObserver: ResizeObserver | null = null

const visibleGroups = computed(() => props.groups.slice(0, visibleCount.value))
const hiddenGroups = computed(() => props.groups.slice(visibleCount.value))

function formatRate(rate: number): string {
  return Number(rate.toFixed(4)).toString()
}

function recalculateVisibleGroups() {
  const availableWidth = containerRef.value?.clientWidth ?? 0
  const groupWidths = Array.from(measureRef.value?.querySelectorAll<HTMLElement>('[data-group-measure]') ?? [])
    .map(element => element.offsetWidth)
  const overflowWidth = measureOverflowRef.value?.offsetWidth ?? 0
  const gap = Number.parseFloat(measureRef.value ? getComputedStyle(measureRef.value).columnGap : '') || 0
  visibleCount.value = visibleMarketplaceGroupCount(groupWidths, availableWidth, overflowWidth, gap)
}

function selectHiddenGroup(groupId: number) {
  showOverflow.value = false
  emit('select', groupId)
}

function handleDocumentPointerDown(event: PointerEvent) {
  if (!containerRef.value?.contains(event.target as Node)) showOverflow.value = false
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') showOverflow.value = false
}

watch(
  () => props.groups,
  async () => {
    showOverflow.value = false
    visibleCount.value = props.groups.length
    await nextTick()
    recalculateVisibleGroups()
  },
  { deep: true, flush: 'post' },
)

watch(showOverflow, (open) => {
  if (open) {
    document.addEventListener('pointerdown', handleDocumentPointerDown)
    document.addEventListener('keydown', handleDocumentKeydown)
  } else {
    document.removeEventListener('pointerdown', handleDocumentPointerDown)
    document.removeEventListener('keydown', handleDocumentKeydown)
  }
})

onMounted(async () => {
  await nextTick()
  recalculateVisibleGroups()
  if (typeof ResizeObserver !== 'undefined' && containerRef.value) {
    resizeObserver = new ResizeObserver(recalculateVisibleGroups)
    resizeObserver.observe(containerRef.value)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>
