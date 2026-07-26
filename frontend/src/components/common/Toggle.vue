<template>
  <button
    type="button"
    @click="toggle"
    class="toggle-control"
    :data-checked="modelValue"
    :disabled="disabled"
    role="switch"
    :aria-checked="modelValue"
    :aria-label="ariaLabel"
    :aria-labelledby="ariaLabelledby"
  >
    <span
      class="toggle-thumb"
      aria-hidden="true"
    />
  </button>
</template>

<script setup lang="ts">
import { watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  ariaLabel?: string
  ariaLabelledby?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

if (import.meta.env.DEV) {
  watch(
    [() => props.ariaLabel, () => props.ariaLabelledby],
    ([ariaLabel, ariaLabelledby]) => {
      if (!ariaLabel?.trim() && !ariaLabelledby?.trim()) {
        console.warn(
          '[Toggle] Accessible name required: provide a non-empty aria-label or aria-labelledby.'
        )
      }
    },
    { immediate: true }
  )
}

function toggle() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>

<style scoped>
.toggle-control {
  position: relative;
  display: inline-flex;
  width: 52px;
  height: 44px;
  flex-shrink: 0;
  padding: 0;
  border: 0;
  border-radius: 999px;
  background: transparent;
  cursor: pointer;
  outline: none;
}

.toggle-control::before {
  position: absolute;
  inset: 10px 0;
  border: 1px solid var(--ui-border-strong, #c7d3e2);
  border-radius: 999px;
  background: var(--ui-surface-subtle, #f4f7fb);
  content: '';
  transition: background-color 150ms ease, border-color 150ms ease;
}

.toggle-control[data-checked='true']::before {
  border-color: var(--ui-text, #0f172a);
  background: var(--ui-text, #0f172a);
}

.toggle-thumb {
  position: absolute;
  top: 12px;
  left: 2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--ui-surface-raised, #fff);
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.24);
  pointer-events: none;
  transition: transform 150ms ease;
}

.toggle-control[data-checked='true'] .toggle-thumb {
  transform: translateX(28px);
}

.toggle-control:focus-visible {
  outline: 2px solid var(--ui-focus, #2563eb);
  outline-offset: 1px;
}

.toggle-control:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

@media (prefers-reduced-motion: reduce) {
  .toggle-control::before,
  .toggle-thumb {
    transition-duration: 1ms;
  }
}
</style>
