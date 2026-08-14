<template>
  <span ref="root" class="metric-info">
    <button
      ref="trigger"
      type="button"
      :aria-label="label"
      :aria-describedby="open ? tooltipId : undefined"
      :aria-expanded="open"
      @click.stop="toggle"
      @mouseenter="show"
      @mouseleave="scheduleClose"
      @focus="onFocus"
      @blur="scheduleClose"
    >
      <Icon name="infoCircle" size="xs" />
    </button>
    <Teleport to="body">
      <Transition name="metric-tooltip">
        <div
          v-if="open"
          :id="tooltipId"
          ref="tooltip"
          class="metric-tooltip"
          :class="`is-${placement}`"
          :style="tooltipStyle"
          role="tooltip"
          @mouseenter="cancelClose"
          @mouseleave="scheduleClose"
        >
          {{ description }}
          <i class="metric-tooltip-arrow" :style="arrowStyle" aria-hidden="true" />
        </div>
      </Transition>
    </Teleport>
  </span>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{ label: string; description: string }>()

const tooltipId = `metric-tooltip-${Math.random().toString(36).slice(2, 9)}`
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)
const tooltip = ref<HTMLElement | null>(null)
const open = ref(false)
const placement = ref<'top' | 'bottom'>('top')
const left = ref(0)
const top = ref(0)
const arrowLeft = ref(0)
let closeTimer: number | undefined
let suppressNextFocus = false

const tooltipStyle = computed(() => ({ left: `${left.value}px`, top: `${top.value}px` }))
const arrowStyle = computed(() => ({ left: `${arrowLeft.value}px` }))

function cancelClose() {
  if (closeTimer !== undefined) window.clearTimeout(closeTimer)
  closeTimer = undefined
}

function scheduleClose() {
  cancelClose()
  closeTimer = window.setTimeout(() => {
    open.value = false
  }, 100)
}

async function show() {
  cancelClose()
  open.value = true
  await nextTick()
  updatePosition()
}

function onFocus() {
  if (suppressNextFocus) {
    suppressNextFocus = false
    return
  }
  void show()
}

function toggle() {
  if (open.value) {
    open.value = false
    return
  }
  void show()
}

function updatePosition() {
  const anchor = trigger.value?.getBoundingClientRect()
  const panel = tooltip.value?.getBoundingClientRect()
  if (!anchor || !panel) return

  const viewportPadding = 12
  const gap = 9
  const centeredLeft = anchor.left + anchor.width / 2 - panel.width / 2
  const maxLeft = Math.max(viewportPadding, window.innerWidth - panel.width - viewportPadding)
  const nextLeft = Math.min(Math.max(centeredLeft, viewportPadding), maxLeft)
  const hasTopSpace = anchor.top >= panel.height + gap + viewportPadding

  placement.value = hasTopSpace ? 'top' : 'bottom'
  left.value = nextLeft
  top.value = hasTopSpace ? anchor.top - panel.height - gap : anchor.bottom + gap
  const arrowMax = Math.max(12, panel.width - 12)
  arrowLeft.value = Math.min(Math.max(anchor.left + anchor.width / 2 - nextLeft, 12), arrowMax)
}

function closeFromOutside(event: PointerEvent) {
  const target = event.target as Node
  if (!root.value?.contains(target) && !tooltip.value?.contains(target)) open.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && open.value) {
    open.value = false
    if (trigger.value && document.activeElement !== trigger.value) {
      suppressNextFocus = true
      trigger.value.focus()
    }
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', closeFromOutside)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', updatePosition)
  window.addEventListener('scroll', updatePosition, true)
})

onBeforeUnmount(() => {
  cancelClose()
  document.removeEventListener('pointerdown', closeFromOutside)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', updatePosition)
  window.removeEventListener('scroll', updatePosition, true)
})
</script>

<style scoped>
.metric-info {
  display: inline-flex !important;
  flex: none;
  align-items: center;
  margin-left: 3px;
  vertical-align: middle;
}

.metric-info button {
  display: inline-grid;
  width: 28px;
  height: 28px;
  flex: none;
  place-items: center;
  border-radius: 5px;
  color: var(--ui-text-subtle);
}

.metric-info button:hover,
.metric-info button:focus-visible,
.metric-info button[aria-expanded='true'] {
  background: var(--ui-surface-subtle);
  color: var(--ui-text);
}

.metric-info button:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: 1px;
}

.metric-tooltip {
  position: fixed;
  z-index: 1000;
  width: max-content;
  max-width: min(280px, calc(100vw - 24px));
  border: 1px solid rgb(255 255 255 / 10%);
  border-radius: 6px;
  background: #1f2937;
  padding: 9px 11px;
  color: #f8fafc;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.55;
  letter-spacing: 0;
  box-shadow: 0 8px 22px rgb(15 23 42 / 22%);
  pointer-events: auto;
  white-space: normal;
}

.metric-tooltip-arrow {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #1f2937;
  transform: translateX(-50%) rotate(45deg);
}

.metric-tooltip.is-top .metric-tooltip-arrow {
  bottom: -4px;
  border-right: 1px solid rgb(255 255 255 / 10%);
  border-bottom: 1px solid rgb(255 255 255 / 10%);
}

.metric-tooltip.is-bottom .metric-tooltip-arrow {
  top: -4px;
  border-top: 1px solid rgb(255 255 255 / 10%);
  border-left: 1px solid rgb(255 255 255 / 10%);
}

.metric-tooltip-enter-active,
.metric-tooltip-leave-active {
  transition:
    opacity 120ms ease,
    transform 120ms ease;
}

.metric-tooltip-enter-from,
.metric-tooltip-leave-to {
  opacity: 0;
  transform: translateY(2px);
}

@media (prefers-reduced-motion: reduce) {
  .metric-tooltip-enter-active,
  .metric-tooltip-leave-active {
    transition: none;
  }
}
</style>
