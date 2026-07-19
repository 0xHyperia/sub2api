<script setup lang="ts">
import { computed, getCurrentInstance, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  content?: string
  trigger?: 'hover' | 'click'
  widthClass?: string
  label?: string
  closeLabel?: string
}>(), {
  trigger: 'hover',
  widthClass: 'w-64',
  label: '',
  closeLabel: 'Close',
})

const show = ref(false)
const triggerRef = useTemplateRef<HTMLElement>('trigger')
const tooltipRef = useTemplateRef<HTMLElement>('tooltip')
const tooltipStyle = ref({ top: '0px', left: '0px' })
const placement = ref<'top' | 'bottom'>('top')
const tooltipId = `help-tooltip-${getCurrentInstance()?.uid ?? 0}`
const triggerLabel = computed(() => props.label || props.content || 'More information')

function openTooltip() {
  show.value = true
  nextTick(updatePosition)
}

function closeTooltip(restoreFocus = false) {
  show.value = false
  if (restoreFocus) {
    const focusTarget = triggerRef.value?.querySelector<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
    )
    focusTarget?.focus()
  }
}

function onEnter() {
  if (props.trigger !== 'hover') return
  openTooltip()
}

function onLeave() {
  if (props.trigger !== 'hover') return
  closeTooltip()
}

function onFocusIn() {
  if (props.trigger === 'hover') openTooltip()
}

function onFocusOut(event: FocusEvent) {
  if (props.trigger !== 'hover') return
  const nextTarget = event.relatedTarget as Node | null
  if (nextTarget && triggerRef.value?.contains(nextTarget)) return
  closeTooltip()
}

function onClick(event: MouseEvent) {
  if (props.trigger !== 'click') return
  event.stopPropagation()
  if (show.value) {
    closeTooltip()
    return
  }
  openTooltip()
}

function onDocumentClick(event: MouseEvent) {
  if (props.trigger !== 'click' || !show.value) return
  const target = event.target as Node | null
  if (!target) return
  if (triggerRef.value?.contains(target) || tooltipRef.value?.contains(target)) return
  closeTooltip()
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closeTooltip(props.trigger === 'click')
  }
}

function onViewportChange() {
  if (!show.value) return
  updatePosition()
}

function updatePosition() {
  const el = triggerRef.value
  const tooltip = tooltipRef.value
  if (!el || !tooltip) return
  const rect = el.getBoundingClientRect()
  const tooltipRect = tooltip.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const margin = 16
  const gap = 8
  const tooltipWidth = Math.min(
    tooltipRect.width || 256,
    Math.max(0, viewportWidth - margin * 2),
  )
  const desiredCenter = rect.left + rect.width / 2
  const minCenter = margin + tooltipWidth / 2
  const maxCenter = viewportWidth - margin - tooltipWidth / 2
  const left = minCenter <= maxCenter
    ? Math.min(maxCenter, Math.max(minCenter, desiredCenter))
    : viewportWidth / 2
  const tooltipHeight = tooltipRect.height || 0
  const spaceAbove = rect.top - margin
  const spaceBelow = viewportHeight - rect.bottom - margin
  placement.value = spaceAbove >= tooltipHeight + gap || spaceAbove >= spaceBelow ? 'top' : 'bottom'

  tooltipStyle.value = {
    top: `${placement.value === 'top' ? rect.top - gap : rect.bottom + gap}px`,
    left: `${left}px`,
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick, true)
  document.addEventListener('keydown', onDocumentKeydown)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick, true)
  document.removeEventListener('keydown', onDocumentKeydown)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<template>
  <div
    ref="trigger"
    class="group relative ml-1 inline-flex items-center align-middle"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @focusin="onFocusIn"
    @focusout="onFocusOut"
    @click="onClick"
  >
    <!-- Trigger Icon -->
    <slot name="trigger">
      <button
        type="button"
        class="inline-flex h-7 w-7 cursor-help items-center justify-center rounded-control text-foreground-subtle transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
        :aria-label="triggerLabel"
        :aria-describedby="tooltipId"
        :aria-expanded="show"
      >
        <Icon name="infoCircle" size="sm" aria-hidden="true" />
      </button>
    </slot>

    <!-- Teleport to body to escape modal overflow clipping -->
    <Teleport to="body">
      <div
        ref="tooltip"
        v-show="show"
        :id="tooltipId"
        role="tooltip"
        :class="[
          'fixed z-[99999] max-w-[calc(100vw-2rem)] -translate-x-1/2 rounded-panel bg-inverse p-3 text-xs leading-relaxed text-inverse-foreground shadow-floating ring-1 ring-inverse-foreground/10',
          placement === 'top' && '-translate-y-full',
          props.widthClass,
        ]"
        :style="tooltipStyle"
      >
        <button
          v-if="props.trigger === 'click'"
          type="button"
          class="absolute right-1.5 top-1.5 rounded-control p-1 text-inverse-foreground/70 transition-colors hover:bg-inverse-foreground/10 hover:text-inverse-foreground"
          :aria-label="closeLabel"
          @click.stop="closeTooltip(true)"
        >
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        <slot>{{ content }}</slot>
        <div
          :class="[
            'absolute left-1/2 h-2 w-2 -translate-x-1/2 rotate-45 bg-inverse',
            placement === 'top' ? '-bottom-1' : '-top-1'
          ]"
        ></div>
      </div>
    </Teleport>
  </div>
</template>
