<template>
  <div class="relative" ref="containerRef">
    <button
      :id="triggerId"
      ref="triggerRef"
      type="button"
      @click="toggle"
      @keydown="handleTriggerKeydown"
      :class="['date-picker-trigger', isOpen && 'date-picker-trigger-open']"
      aria-haspopup="dialog"
      :aria-expanded="isOpen"
      :aria-controls="isOpen ? dropdownId : undefined"
    >
      <span class="date-picker-icon">
        <Icon name="calendar" size="sm" />
      </span>
      <span class="date-picker-value">
        {{ displayValue }}
      </span>
      <span class="date-picker-chevron">
        <Icon
          name="chevronDown"
          size="sm"
          :class="['transition-transform duration-200', isOpen && 'rotate-180']"
        />
      </span>
    </button>

    <Transition name="date-picker-dropdown">
      <div
        v-if="isOpen"
        :id="dropdownId"
        class="date-picker-dropdown w-[20rem] max-w-[calc(100vw-2rem)]"
        role="dialog"
        :aria-labelledby="triggerId"
      >
        <!-- Quick presets -->
        <div class="date-picker-presets">
          <button
            v-for="preset in presets"
            :key="preset.value"
            type="button"
            data-date-picker-preset
            @click="selectPreset(preset)"
            :class="['date-picker-preset', isPresetActive(preset) && 'date-picker-preset-active']"
            :aria-pressed="isPresetActive(preset)"
          >
            {{ t(preset.labelKey) }}
          </button>
        </div>

        <div class="date-picker-divider"></div>

        <!-- Custom date range inputs -->
        <div class="date-picker-custom flex-col items-stretch min-[360px]:flex-row min-[360px]:items-end">
          <div class="date-picker-field">
            <label :for="startInputId" class="date-picker-label">{{ t('dates.startDate') }}</label>
            <input
              :id="startInputId"
              type="date"
              v-model="localStartDate"
              :max="localEndDate || tomorrow"
              class="date-picker-input"
              @change="onDateChange"
            />
          </div>
          <div class="date-picker-separator rotate-90 min-[360px]:rotate-0 min-[360px]:pb-1" aria-hidden="true">
            <Icon name="arrowRight" size="sm" class="text-foreground-subtle" />
          </div>
          <div class="date-picker-field">
            <label :for="endInputId" class="date-picker-label">{{ t('dates.endDate') }}</label>
            <input
              :id="endInputId"
              type="date"
              v-model="localEndDate"
              :min="localStartDate"
              :max="tomorrow"
              class="date-picker-input"
              @change="onDateChange"
            />
          </div>
        </div>

        <!-- Apply button -->
        <div class="date-picker-actions">
          <button type="button" @click="apply" class="date-picker-apply">
            {{ t('dates.apply') }}
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, getCurrentInstance, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

interface DatePreset {
  labelKey: string
  value: string
  getRange: () => { start: string; end: string }
}

interface Props {
  startDate: string
  endDate: string
}

interface Emits {
  (e: 'update:startDate', value: string): void
  (e: 'update:endDate', value: string): void
  (e: 'change', range: { startDate: string; endDate: string; preset: string | null }): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t, locale } = useI18n()

const isOpen = ref(false)
const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const instanceId = getCurrentInstance()?.uid ?? 0
const triggerId = `date-range-picker-trigger-${instanceId}`
const dropdownId = `date-range-picker-dropdown-${instanceId}`
const startInputId = `date-range-picker-start-${instanceId}`
const endInputId = `date-range-picker-end-${instanceId}`
const localStartDate = ref(props.startDate)
const localEndDate = ref(props.endDate)
const activePreset = ref<string | null>('last24Hours')

const today = computed(() => {
  // Use local timezone to avoid UTC timezone issues
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
})

// Tomorrow's date - used for max date to handle timezone differences
// When user is in a timezone behind the server, "today" on server might be "tomorrow" locally
const tomorrow = computed(() => {
  const d = new Date()
  d.setDate(d.getDate() + 1)
  return formatDateToString(d)
})

// Helper function to format date to YYYY-MM-DD using local timezone
const formatDateToString = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const presets: DatePreset[] = [
  {
    labelKey: 'dates.today',
    value: 'today',
    getRange: () => {
      const t = today.value
      return { start: t, end: t }
    }
  },
  {
    labelKey: 'dates.yesterday',
    value: 'yesterday',
    getRange: () => {
      const d = new Date()
      d.setDate(d.getDate() - 1)
      const yesterday = formatDateToString(d)
      return { start: yesterday, end: yesterday }
    }
  },
  {
    labelKey: 'dates.last24Hours',
    value: 'last24Hours',
    getRange: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      return {
        start: formatDateToString(start),
        end: formatDateToString(end)
      }
    }
  },
  {
    labelKey: 'dates.last7Days',
    value: '7days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 6)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last14Days',
    value: '14days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 13)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.last30Days',
    value: '30days',
    getRange: () => {
      const end = today.value
      const d = new Date()
      d.setDate(d.getDate() - 29)
      const start = formatDateToString(d)
      return { start, end }
    }
  },
  {
    labelKey: 'dates.thisMonth',
    value: 'thisMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 1))
      return { start, end: today.value }
    }
  },
  {
    labelKey: 'dates.lastMonth',
    value: 'lastMonth',
    getRange: () => {
      const now = new Date()
      const start = formatDateToString(new Date(now.getFullYear(), now.getMonth() - 1, 1))
      const end = formatDateToString(new Date(now.getFullYear(), now.getMonth(), 0))
      return { start, end }
    }
  }
]

const displayValue = computed(() => {
  if (activePreset.value) {
    const preset = presets.find((p) => p.value === activePreset.value)
    if (preset) return t(preset.labelKey)
  }

  if (localStartDate.value && localEndDate.value) {
    if (localStartDate.value === localEndDate.value) {
      return formatDate(localStartDate.value)
    }
    return `${formatDate(localStartDate.value)} - ${formatDate(localEndDate.value)}`
  }

  return t('dates.selectDateRange')
})

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr + 'T00:00:00')
  const dateLocale = locale.value === 'zh' ? 'zh-CN' : 'en-US'
  return date.toLocaleDateString(dateLocale, { month: 'short', day: 'numeric' })
}

const isPresetActive = (preset: DatePreset): boolean => {
  return activePreset.value === preset.value
}

const selectPreset = (preset: DatePreset) => {
  const range = preset.getRange()
  localStartDate.value = range.start
  localEndDate.value = range.end
  activePreset.value = preset.value
}

const onDateChange = () => {
  // Check if current dates match any preset
  activePreset.value = null
  for (const preset of presets) {
    const range = preset.getRange()
    if (range.start === localStartDate.value && range.end === localEndDate.value) {
      activePreset.value = preset.value
      break
    }
  }
}

const focusPreset = async (position: 'first' | 'last') => {
  await nextTick()
  const presetButtons = containerRef.value?.querySelectorAll<HTMLButtonElement>(
    '[data-date-picker-preset]'
  )
  const target = position === 'first'
    ? presetButtons?.[0]
    : presetButtons?.[presetButtons.length - 1]
  target?.focus()
}

const close = async (restoreTriggerFocus = false) => {
  isOpen.value = false
  if (restoreTriggerFocus) {
    await nextTick()
    triggerRef.value?.focus()
  }
}

const toggle = () => {
  if (isOpen.value) {
    void close()
  } else {
    isOpen.value = true
  }
}

const handleTriggerKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    isOpen.value = true
    void focusPreset(event.key === 'ArrowDown' ? 'first' : 'last')
  }
}

const apply = () => {
  emit('update:startDate', localStartDate.value)
  emit('update:endDate', localEndDate.value)
  emit('change', {
    startDate: localStartDate.value,
    endDate: localEndDate.value,
    preset: activePreset.value
  })
  void close(true)
}

const handleClickOutside = (event: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    void close()
  }
}

const handleEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && isOpen.value) {
    event.preventDefault()
    void close(true)
  }
}

// Sync local state with props
watch(
  () => props.startDate,
  (val) => {
    localStartDate.value = val
    onDateChange()
  }
)

watch(
  () => props.endDate,
  (val) => {
    localEndDate.value = val
    onDateChange()
  }
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleEscape)
  // Initialize active preset detection
  onDateChange()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleEscape)
})
</script>

<style scoped>
.date-picker-trigger {
  @apply flex max-w-full min-w-0 items-center gap-2;
  @apply rounded-panel px-3 py-2 text-sm;
  @apply bg-surface;
  @apply border border-outline;
  @apply text-foreground-muted;
  @apply transition-all duration-200;
  @apply focus:border-focus focus:outline-none focus:ring-2 focus:ring-focus/30;
  @apply hover:border-outline-strong;
  @apply cursor-pointer;
}

.date-picker-trigger-open {
  @apply border-brand ring-2 ring-focus/30;
}

.date-picker-icon {
  @apply text-foreground-subtle text-foreground-muted;
}

.date-picker-value {
  @apply min-w-0 truncate font-medium;
}

.date-picker-chevron {
  @apply text-foreground-subtle text-foreground-muted;
}

.date-picker-dropdown {
  @apply absolute left-0 z-[100] mt-2;
  @apply bg-surface;
  @apply rounded-panel;
  @apply border border-outline;
  @apply shadow-floating shadow-black/10 shadow-black/10;
  @apply overflow-hidden;
}

.date-picker-presets {
  @apply grid grid-cols-2 gap-1 p-2;
}

.date-picker-preset {
  @apply rounded-control px-3 py-1.5 text-xs font-medium;
  @apply text-foreground-muted text-foreground-subtle;
  @apply hover:bg-surface-subtle;
  @apply transition-colors duration-150;
}

.date-picker-preset-active {
  @apply bg-brand-subtle;
  @apply text-brand;
}

.date-picker-divider {
  @apply border-t border-outline;
}

.date-picker-custom {
  @apply flex gap-2 p-3;
}

.date-picker-field {
  @apply min-w-0 w-full;
}

.date-picker-label {
  @apply mb-1 block text-xs font-medium text-foreground-subtle;
}

.date-picker-input {
  @apply w-full min-w-0 rounded-control px-2 py-1.5 text-sm;
  @apply bg-surface-subtle;
  @apply border border-outline;
  @apply text-foreground;
  @apply focus:border-focus focus:outline-none focus:ring-2 focus:ring-focus/30;
}

.date-picker-input::-webkit-calendar-picker-indicator {
  @apply cursor-pointer opacity-60 hover:opacity-100;
  filter: invert(0.5);
}

.dark .date-picker-input::-webkit-calendar-picker-indicator {
  filter: invert(0.7);
}

.date-picker-separator {
  @apply flex items-center justify-center;
}

.date-picker-actions {
  @apply flex justify-end p-2 pt-0;
}

.date-picker-apply {
  @apply rounded-panel px-4 py-1.5 text-sm font-medium;
  @apply bg-brand text-brand-foreground;
  @apply hover:bg-brand-hover;
  @apply transition-colors duration-150;
}

/* Dropdown animation */
.date-picker-dropdown-enter-active,
.date-picker-dropdown-leave-active {
  transition: all 0.2s ease;
}

.date-picker-dropdown-enter-from,
.date-picker-dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
