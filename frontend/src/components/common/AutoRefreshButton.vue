<template>
  <div class="relative" ref="dropdownRef">
    <button
      :id="triggerId"
      ref="triggerRef"
      type="button"
      class="inline-flex min-h-9 items-center gap-1.5 rounded-control border border-outline bg-surface px-2.5 text-xs font-medium text-foreground-muted shadow-card transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40"
      :title="t('common.autoRefresh.title')"
      :aria-label="t('common.autoRefresh.title')"
      :aria-expanded="showDropdown"
      :aria-controls="showDropdown ? menuId : undefined"
      aria-haspopup="menu"
      @click="toggleMenu"
      @keydown="handleTriggerKeydown"
    >
      <Icon name="refresh" size="xs" />
      <span class="tabular-nums">
        {{ enabled
          ? t('common.autoRefresh.countdown', { seconds: countdown })
          : t('common.autoRefresh.title')
        }}
      </span>
    </button>

    <div
      v-if="showDropdown"
      :id="menuId"
      ref="menuRef"
      role="menu"
      :aria-labelledby="triggerId"
      class="absolute right-0 z-20 mt-1 w-48 rounded-panel border border-outline bg-surface p-1 shadow-floating"
      @keydown="handleMenuKeydown"
    >
      <div>
        <button
          type="button"
          role="menuitemcheckbox"
          tabindex="-1"
          :aria-checked="enabled"
          @click="toggleEnabled"
          class="flex min-h-10 w-full items-center justify-between rounded-control px-3 py-2 text-sm text-foreground-muted hover:bg-surface-subtle hover:text-foreground focus-visible:bg-surface-subtle focus-visible:text-foreground focus-visible:outline-none"
        >
          <span>{{ t('common.autoRefresh.enable') }}</span>
          <Icon v-if="enabled" name="check" size="sm" class="text-foreground" />
        </button>
        <div role="separator" class="my-1 border-t border-outline"></div>
        <button
          v-for="sec in intervals"
          :key="sec"
          type="button"
          role="menuitemradio"
          tabindex="-1"
          :aria-checked="intervalSeconds === sec"
          @click="selectInterval(sec)"
          class="flex min-h-10 w-full items-center justify-between rounded-control px-3 py-2 text-sm text-foreground-muted hover:bg-surface-subtle hover:text-foreground focus-visible:bg-surface-subtle focus-visible:text-foreground focus-visible:outline-none"
        >
          <span>{{ t('common.autoRefresh.seconds', { n: sec }) }}</span>
          <Icon v-if="intervalSeconds === sec" name="check" size="sm" class="text-foreground" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useDropdownMenu } from '@/composables/useDropdownMenu'

const props = defineProps<{
  enabled: boolean
  intervalSeconds: number
  countdown: number
  intervals: readonly number[]
}>()

const emit = defineEmits<{
  (e: 'update:enabled', value: boolean): void
  (e: 'update:interval', value: number): void
}>()

const { t } = useI18n()
const dropdownRef = ref<HTMLElement | null>(null)
const {
  open: showDropdown,
  triggerRef,
  menuRef,
  triggerId,
  menuId,
  closeMenu,
  toggleMenu,
  handleTriggerKeydown,
  handleMenuKeydown
} = useDropdownMenu('auto-refresh-menu')

function toggleEnabled() {
  emit('update:enabled', !props.enabled)
}

function selectInterval(seconds: number) {
  emit('update:interval', seconds)
  void closeMenu(true)
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    void closeMenu()
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>
