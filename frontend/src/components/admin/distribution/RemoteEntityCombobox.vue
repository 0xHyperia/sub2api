<template>
  <div ref="rootRef" class="space-y-2">
    <label :for="inputId" class="input-label">{{ label }}</label>

    <div
      v-if="modelValue"
      class="flex min-w-0 items-center justify-between gap-3 rounded-control border border-brand/40 bg-brand-subtle px-3 py-2.5"
    >
      <div class="min-w-0">
        <p class="truncate text-sm font-medium text-foreground">{{ modelValue.email }}</p>
        <p class="mt-0.5 truncate text-xs text-foreground-subtle">
          {{ modelValue.username || '未设置用户名' }}
          <span v-if="modelValue.meta"> · {{ modelValue.meta }}</span>
        </p>
      </div>
      <button
        type="button"
        class="btn btn-ghost btn-icon h-11 w-11 shrink-0"
        aria-label="更换选择"
        title="更换选择"
        @click="clearSelection"
      >
        <Icon name="x" size="sm" />
      </button>
    </div>

    <div v-else class="relative">
      <Icon
        name="search"
        size="sm"
        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle"
      />
      <input
        ref="inputRef"
        :id="inputId"
        v-model="query"
        type="search"
        class="input min-h-11 pl-9"
        :placeholder="placeholder"
        role="combobox"
        aria-autocomplete="list"
        :aria-controls="listboxId"
        :aria-expanded="open"
        :aria-activedescendant="activeOptionId"
        :aria-busy="loading"
        autocomplete="off"
        @input="scheduleSearch"
        @focus="handleFocus"
        @keydown="handleKeydown"
      />

      <div
        v-if="open"
        :id="listboxId"
        class="absolute z-40 mt-1 max-h-64 w-full overflow-y-auto rounded-panel border border-outline bg-surface py-1 shadow-lg"
        role="listbox"
      >
        <p v-if="loading" class="px-3 py-3 text-sm text-foreground-subtle">正在搜索...</p>
        <p v-else-if="error" class="px-3 py-3 text-sm text-danger-foreground" role="alert">{{ error }}</p>
        <p
          v-else-if="query.trim().length < minChars"
          class="px-3 py-3 text-sm text-foreground-subtle"
        >
          至少输入 {{ minChars }} 个字符
        </p>
        <p v-else-if="options.length === 0" class="px-3 py-3 text-sm text-foreground-subtle">
          没有匹配结果
        </p>
        <button
          v-for="(option, index) in options"
          :id="optionId(index)"
          :key="option.id"
          type="button"
          class="flex min-h-11 w-full min-w-0 items-start justify-between gap-3 px-3 py-2 text-left"
          :class="[
            index === activeIndex && option.selectable !== false ? 'bg-surface-subtle' : '',
            option.selectable === false ? 'cursor-not-allowed opacity-60' : 'hover:bg-surface-subtle',
          ]"
          role="option"
          :aria-selected="index === activeIndex"
          :aria-disabled="option.selectable === false"
          @mouseenter="option.selectable !== false && (activeIndex = index)"
          @click="selectOption(option)"
        >
          <span class="min-w-0">
            <span class="block truncate text-sm font-medium text-foreground">{{ option.email }}</span>
            <span class="mt-0.5 block truncate text-xs text-foreground-subtle">
              {{ option.username || '未设置用户名' }}
              <template v-if="option.meta"> · {{ option.meta }}</template>
            </span>
          </span>
          <span
            v-if="option.reason"
            class="max-w-32 shrink-0 text-right text-xs text-danger-foreground"
          >
            {{ option.reason }}
          </span>
        </button>
      </div>
      <p class="sr-only" role="status" aria-live="polite">{{ statusMessage }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import type {
  DistributionPickerOption,
  DistributionPickerSearch,
} from './types'

const props = withDefaults(defineProps<{
  modelValue: DistributionPickerOption | null
  label: string
  placeholder: string
  inputId: string
  search: DistributionPickerSearch
  minChars?: number
}>(), {
  minChars: 2,
})

const emit = defineEmits<{
  'update:modelValue': [value: DistributionPickerOption | null]
}>()

const rootRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const query = ref('')
const options = ref<DistributionPickerOption[]>([])
const loading = ref(false)
const error = ref('')
const open = ref(false)
const activeIndex = ref(-1)
const listboxId = `${props.inputId}-options`
let timer: number | null = null
let requestSequence = 0

const optionId = (index: number) => `${props.inputId}-option-${index}`
const activeOptionId = computed(() => open.value && activeIndex.value >= 0 ? optionId(activeIndex.value) : undefined)
const statusMessage = computed(() => {
  if (!open.value) return ''
  if (loading.value) return '正在搜索'
  if (error.value) return error.value
  if (query.value.trim().length < props.minChars) return `至少输入 ${props.minChars} 个字符`
  return options.value.length ? `找到 ${options.value.length} 个结果` : '没有匹配结果'
})

function firstSelectableIndex() {
  return options.value.findIndex((option) => option.selectable !== false)
}

function scheduleSearch() {
  emit('update:modelValue', null)
  error.value = ''
  open.value = true
  requestSequence += 1
  if (timer) window.clearTimeout(timer)
  if (query.value.trim().length < props.minChars) {
    options.value = []
    activeIndex.value = -1
    loading.value = false
    return
  }
  loading.value = true
  timer = window.setTimeout(runSearch, 250)
}

async function runSearch() {
  const value = query.value.trim()
  if (value.length < props.minChars) return
  const sequence = ++requestSequence
  loading.value = true
  try {
    const result = await props.search(value)
    if (sequence !== requestSequence) return
    options.value = result
    activeIndex.value = firstSelectableIndex()
  } catch {
    if (sequence !== requestSequence) return
    options.value = []
    activeIndex.value = -1
    error.value = '搜索失败，请稍后重试'
  } finally {
    if (sequence === requestSequence) loading.value = false
  }
}

function selectOption(option: DistributionPickerOption) {
  if (option.selectable === false) return
  emit('update:modelValue', option)
  query.value = ''
  options.value = []
  activeIndex.value = -1
  open.value = false
}

function clearSelection() {
  emit('update:modelValue', null)
  query.value = ''
  options.value = []
  activeIndex.value = -1
  open.value = false
  void nextTick(() => inputRef.value?.focus())
}

function handleFocus() {
  open.value = true
  if (query.value.trim().length >= props.minChars && options.value.length === 0) {
    void runSearch()
  }
}

function moveActive(direction: 1 | -1) {
  if (!options.value.length) return
  let index = activeIndex.value
  for (let attempts = 0; attempts < options.value.length; attempts += 1) {
    index = (index + direction + options.value.length) % options.value.length
    if (options.value[index]?.selectable !== false) {
      activeIndex.value = index
      return
    }
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
    event.preventDefault()
    open.value = true
    moveActive(event.key === 'ArrowDown' ? 1 : -1)
  } else if (event.key === 'Enter' && activeIndex.value >= 0) {
    event.preventDefault()
    const option = options.value[activeIndex.value]
    if (option) selectOption(option)
  } else if (event.key === 'Escape') {
    open.value = false
    activeIndex.value = -1
  }
}

function handleDocumentPointer(event: MouseEvent) {
  if (!rootRef.value?.contains(event.target as Node)) open.value = false
}

onMounted(() => document.addEventListener('mousedown', handleDocumentPointer))
onUnmounted(() => {
  document.removeEventListener('mousedown', handleDocumentPointer)
  if (timer) window.clearTimeout(timer)
  requestSequence += 1
})
</script>
