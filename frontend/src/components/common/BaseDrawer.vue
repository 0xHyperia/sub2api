<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="show" class="fixed inset-0 z-50 flex justify-end bg-black/35" role="presentation" @click.self="requestClose">
        <aside ref="panel" class="flex h-full w-full max-w-3xl flex-col border-l border-outline bg-surface-raised shadow-floating" role="dialog" aria-modal="true" :aria-labelledby="titleId" tabindex="-1">
          <header class="flex min-h-16 shrink-0 items-center justify-between gap-3 border-b border-outline px-4 sm:px-6">
            <div class="min-w-0"><h2 :id="titleId" class="truncate text-base font-semibold text-foreground">{{ title }}</h2><p v-if="description" class="mt-0.5 truncate text-xs text-foreground-subtle">{{ description }}</p></div>
            <button type="button" class="btn btn-ghost btn-icon shrink-0" :aria-label="t('common.close')" @click="emit('close')"><Icon name="x" size="md" /></button>
          </header>
          <div class="min-h-0 flex-1 overflow-y-auto px-4 py-5 sm:px-6"><slot /></div>
          <footer v-if="$slots.footer" class="shrink-0 border-t border-outline bg-surface-raised px-4 py-3 sm:px-6"><slot name="footer" /></footer>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{ show: boolean; title: string; description?: string; closeOnOutside?: boolean }>(), { description: '', closeOnOutside: true })
const emit = defineEmits<{ (event: 'close'): void }>()
const { t } = useI18n()
const panel = ref<HTMLElement | null>(null)
const titleId = `drawer-title-${Math.random().toString(36).slice(2)}`
let previousFocus: HTMLElement | null = null
function requestClose() { if (props.closeOnOutside) emit('close') }
function onKeydown(event: KeyboardEvent) {
  if (!props.show) return
  if (event.key === 'Escape') { event.preventDefault(); emit('close'); return }
  if (event.key !== 'Tab' || !panel.value) return
  const focusable = [...panel.value.querySelectorAll<HTMLElement>('button:not([disabled]),a[href],input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])')]
  if (!focusable.length) { event.preventDefault(); panel.value.focus(); return }
  const first = focusable[0]; const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}
watch(() => props.show, async (show) => {
  if (show) { previousFocus = document.activeElement as HTMLElement; document.body.classList.add('modal-open'); document.addEventListener('keydown', onKeydown); await nextTick(); panel.value?.focus() }
  else { document.body.classList.remove('modal-open'); document.removeEventListener('keydown', onKeydown); previousFocus?.focus(); previousFocus = null }
})
onBeforeUnmount(() => { document.removeEventListener('keydown', onKeydown); if (props.show) document.body.classList.remove('modal-open') })
</script>

<style scoped>
.drawer-enter-active,.drawer-leave-active { transition: background-color 180ms ease; }
.drawer-enter-active aside,.drawer-leave-active aside { transition: transform 180ms ease, opacity 180ms ease; }
.drawer-enter-from,.drawer-leave-to { background-color: transparent; }
.drawer-enter-from aside,.drawer-leave-to aside { opacity: 0; transform: translateX(24px); }
@media (prefers-reduced-motion: reduce) { .drawer-enter-active,.drawer-leave-active,.drawer-enter-active aside,.drawer-leave-active aside { transition: none; } }
</style>
