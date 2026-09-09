<template>
  <div class="space-y-4">
    <label :for="`${idPrefix}-max-effort`" class="input-label">{{ t('admin.groups.form.maxReasoningEffort') }}</label>
    <Select :id="`${idPrefix}-max-effort`" :model-value="maxEffort" :options="reasoningEffortOptions" :placeholder="t('admin.groups.form.maxReasoningEffortUnlimited')" @update:model-value="updateMaxEffort" />
    <p class="input-hint">{{ t('admin.groups.form.maxReasoningEffortHint') }}</p>
    <div class="border-t border-outline pt-4">
      <div class="mb-3 flex items-center justify-between"><label class="input-label mb-0">{{ t('admin.groups.form.reasoningEffortMappings') }}</label><button type="button" class="inline-flex min-h-11 items-center gap-1.5 rounded-lg px-2.5 text-sm font-medium text-brand hover:bg-brand-subtle" @click="addGroup"><Icon name="plus" size="sm" />{{ t('admin.groups.form.addReasoningEffortMapping') }}</button></div>
      <div v-for="group in mappings" :key="group.id" class="space-y-3 rounded-lg border border-outline bg-surface-subtle p-3">
        <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto]">
          <div><label :for="`${idPrefix}-${group.id}-match-type`" class="input-label">{{ t('admin.groups.form.reasoningEffortMatchType') }}</label><Select :id="`${idPrefix}-${group.id}-match-type`" :model-value="group.match_type" :options="matchTypeOptions" :placeholder="t('admin.groups.form.reasoningEffortMatchTypePlaceholder')" @update:model-value="updateGroup(group.id, 'match_type', $event)" /></div>
          <div><label :for="`${idPrefix}-${group.id}-model`" class="input-label">{{ t('admin.groups.form.reasoningEffortModel') }}</label><input :id="`${idPrefix}-${group.id}-model`" :value="group.model" class="input" :placeholder="t('admin.groups.form.reasoningEffortModelPlaceholder')" @input="onModelInput(group.id, $event)" /></div>
          <button type="button" class="flex h-11 w-11 items-center justify-center text-foreground-subtle hover:text-danger-foreground" @click="removeGroup(group.id)"><Icon name="trash" size="sm" /></button>
        </div>
        <div v-for="pair in group.pairs" :key="pair.id" class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_auto]">
          <div><label :for="`${idPrefix}-${pair.id}-from`" class="input-label">{{ t('admin.groups.form.reasoningEffortFrom') }}</label><Select :id="`${idPrefix}-${pair.id}-from`" :model-value="pair.from" :options="reasoningEffortSourceOptions" @update:model-value="updatePair(group.id, pair.id, 'from', $event)" /></div>
          <div><label :for="`${idPrefix}-${pair.id}-to`" class="input-label">{{ t('admin.groups.form.reasoningEffortTo') }}</label><Select :id="`${idPrefix}-${pair.id}-to`" :model-value="pair.to" :options="reasoningEffortTargetOptions" @update:model-value="updatePair(group.id, pair.id, 'to', $event)" /></div>
          <button type="button" class="flex h-11 w-11 items-center justify-center text-foreground-subtle hover:text-danger-foreground" @click="removePair(group.id, pair.id)"><Icon name="trash" size="sm" /></button>
        </div>
        <button type="button" class="text-sm font-medium text-brand" @click="addPair(group.id)"><Icon name="plus" size="sm" /> {{ t('admin.groups.form.addReasoningEffortPair') }}</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'; import { useI18n } from 'vue-i18n'; import type { GroupPlatform } from '@/types'; import Icon from '@/components/icons/Icon.vue'; import Select from '@/components/common/Select.vue'; import { createReasoningEffortMappingPair, createReasoningEffortMappingRow, normalizeReasoningEffortMatchType, reasoningEffortOptionsForPlatform, reasoningEffortSourceOptionsForPlatform, reasoningEffortTargetOptionsForPlatform, reasoningEffortMappingDeny, validateReasoningEffortMappings, type ReasoningEffortMappingRow } from '@/views/admin/groupsReasoningEffort'
const props = defineProps<{ idPrefix: string; platform: GroupPlatform; maxEffort: string; overLimit: string; mappings: ReasoningEffortMappingRow[] }>(); const emit = defineEmits<{ (e:'update:maxEffort', v:string):void; (e:'update:overLimit', v:string):void; (e:'update:mappings', v:ReasoningEffortMappingRow[]):void }>(); const { t } = useI18n()
const showValidation = ref(false); const validationErrors = computed(() => validateReasoningEffortMappings(props.mappings, props.platform))
const reasoningEffortOptions = computed(() => reasoningEffortOptionsForPlatform(props.platform)); const reasoningEffortSourceOptions = computed(() => reasoningEffortSourceOptionsForPlatform(props.platform)); const reasoningEffortTargetOptions = computed(() => reasoningEffortTargetOptionsForPlatform(props.platform).map(option => option.value === reasoningEffortMappingDeny ? { ...option, label: t('admin.groups.form.reasoningEffortToDeny') } : option)); const matchTypeOptions = computed(() => ['exact','prefix','suffix'].map(value => ({ value, label: t(`admin.groups.form.reasoningEffortMatch${value[0].toUpperCase()}${value.slice(1)}`) })))
const value = (v: unknown) => v == null ? '' : String(v)
const updateMaxEffort = (v: unknown) => emit('update:maxEffort', value(v)); const addGroup = () => emit('update:mappings', [...props.mappings, createReasoningEffortMappingRow()]); const removeGroup = (id:string) => emit('update:mappings', props.mappings.filter(row => row.id !== id));
const updateGroup = (id:string, field:'match_type', v:unknown) => emit('update:mappings', props.mappings.map(row => row.id === id ? { ...row, [field]: normalizeReasoningEffortMatchType(value(v)) } : row));
const onModelInput = (id:string, event:Event) => emit('update:mappings', props.mappings.map(row => row.id === id ? { ...row, model: (event.target as HTMLInputElement).value } : row));
const updatePair = (groupID:string, pairID:string, field:'from'|'to', v:unknown) => emit('update:mappings', props.mappings.map(row => row.id === groupID ? { ...row, pairs: row.pairs.map(pair => pair.id === pairID ? { ...pair, [field]: value(v) } : pair) } : row));
const addPair = (groupID:string) => emit('update:mappings', props.mappings.map(row => row.id === groupID ? { ...row, pairs: [...row.pairs, createReasoningEffortMappingPair()] } : row)); const removePair = (groupID:string, pairID:string) => emit('update:mappings', props.mappings.map(row => row.id === groupID ? { ...row, pairs: row.pairs.filter(pair => pair.id !== pairID) } : row));
const validate = (): boolean => { showValidation.value = true; return Object.keys(validationErrors.value).length === 0 }
const resetValidation = () => { showValidation.value = false }
defineExpose({ validate, resetValidation })
</script>
