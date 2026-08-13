<template>
  <div v-if="items.length" class="flex flex-wrap items-center gap-2 text-xs" aria-live="polite">
    <span class="text-foreground-subtle">{{ labelText() }}</span>
    <button v-for="item in items" :key="item.key" type="button" class="badge badge-gray max-w-full" :aria-label="`${item.label}：${item.value}`" @click="emit('remove', item.key)">
      <span class="max-w-48 truncate">{{ item.label }}：{{ item.value }}</span><Icon name="x" size="xs" />
    </button>
    <button type="button" class="text-brand hover:underline" @click="emit('clear')">{{ clearText() }}</button>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import { useI18n } from 'vue-i18n'

export interface DistributionFilterItem { key: string; label: string; value: string }
const props = defineProps<{ items: DistributionFilterItem[]; label?: string; clearLabel?: string }>()
const { t } = useI18n()
const labelText = () => props.label || t('admin.distribution.filterCurrent')
const clearText = () => props.clearLabel || t('admin.distribution.clearFilters')
const emit = defineEmits<{ (event: 'remove', key: string): void; (event: 'clear'): void }>()
</script>
