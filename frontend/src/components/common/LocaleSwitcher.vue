<template>
  <div class="relative" ref="dropdownRef">
    <button
      :id="triggerId"
      ref="triggerRef"
      type="button"
      :disabled="switching"
      class="flex items-center gap-1.5 rounded-panel px-2 py-1.5 text-sm font-medium text-foreground-muted transition-colors hover:bg-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40"
      :title="currentLocale?.name"
      :aria-label="`${t('common.language')}: ${currentLocale?.name || currentLocaleCode}`"
      :aria-expanded="isOpen"
      :aria-controls="isOpen ? menuId : undefined"
      aria-haspopup="menu"
      @click="toggleMenu"
      @keydown="handleTriggerKeydown"
    >
      <span class="text-base">{{ currentLocale?.flag }}</span>
      <span class="hidden sm:inline">{{ currentLocale?.code.toUpperCase() }}</span>
      <Icon
        name="chevronDown"
        size="xs"
        class="text-foreground-subtle transition-transform duration-200"
        :class="{ 'rotate-180': isOpen }"
      />
    </button>

    <transition name="dropdown">
      <div
        v-if="isOpen"
        :id="menuId"
        ref="menuRef"
        role="menu"
        :aria-labelledby="triggerId"
        class="absolute right-0 z-50 mt-1 w-32 overflow-hidden rounded-panel border border-outline shadow-floating bg-surface"
        @keydown="handleMenuKeydown"
      >
        <button
          v-for="locale in availableLocales"
          :key="locale.code"
          type="button"
          role="menuitemradio"
          tabindex="-1"
          :disabled="switching"
          :aria-checked="locale.code === currentLocaleCode"
          @click="selectLocale(locale.code)"
          class="flex w-full items-center gap-2 px-3 py-2 text-sm text-foreground-muted transition-colors hover:bg-surface-subtle focus-visible:outline-none focus-visible:bg-surface-subtle"
          :class="{
            'bg-brand-subtle text-brand':
              locale.code === currentLocaleCode
          }"
        >
          <span class="text-base">{{ locale.flag }}</span>
          <span>{{ locale.name }}</span>
          <Icon v-if="locale.code === currentLocaleCode" name="check" size="sm" class="ml-auto text-brand" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'
import { useDropdownMenu } from '@/composables/useDropdownMenu'

const { locale, t } = useI18n()

const dropdownRef = ref<HTMLElement | null>(null)
const switching = ref(false)
const {
  open: isOpen,
  triggerRef,
  menuRef,
  triggerId,
  menuId,
  closeMenu,
  toggleMenu,
  handleTriggerKeydown,
  handleMenuKeydown
} = useDropdownMenu('locale-menu')

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

async function selectLocale(code: string) {
  if (switching.value) return
  if (code === currentLocaleCode.value) {
    await closeMenu(true)
    return
  }
  switching.value = true
  try {
    await setLocale(code)
    await closeMenu(true)
  } finally {
    switching.value = false
  }
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    void closeMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
