<template>
  <div ref="rootRef" class="analytics-range">
    <div class="analytics-range-group">
      <span class="analytics-range-label">{{ t('common.distributionAnalytics.granularity') }}</span>
      <div class="analytics-segment" role="group" :aria-label="t('common.distributionAnalytics.granularity')">
        <button v-for="item in granularities" :key="item.key" type="button" :aria-pressed="value.granularity === item.key || (item.key === 'custom' && open)" @click="selectGranularity(item.key)">{{ item.label }}</button>
      </div>
    </div>

    <div class="analytics-range-group">
      <span class="analytics-range-label">{{ t('common.distributionAnalytics.analysisPeriod') }}</span>
      <div v-if="value.granularity !== 'custom' && !open" class="analytics-segment" role="group" :aria-label="t('common.distributionAnalytics.analysisPeriod')">
        <button v-for="(label, index) in presetLabels" :key="label" type="button" :aria-pressed="value.preset === `${value.granularity}-${index}`" @click="applyPreset(index)">{{ label }}</button>
      </div>
      <button v-else ref="triggerRef" type="button" class="custom-range-trigger" aria-haspopup="dialog" :aria-expanded="open" @click="open = !open">
        <span>{{ value.date_from }}</span><span aria-hidden="true" class="range-arrow">→</span><span>{{ value.date_to }}</span><Icon name="calendar" size="sm" />
      </button>
    </div>

    <Transition name="range-popover">
      <div v-if="open" class="range-popover" role="dialog" :aria-label="t('common.distributionAnalytics.customRange')">
        <div class="range-popover-heading">
          <div><strong>{{ t('common.distributionAnalytics.customRange') }}</strong><small>{{ t('common.distributionAnalytics.customRangeHint') }}</small></div>
          <button type="button" class="btn btn-ghost btn-icon" :aria-label="t('common.close')" @click="close(true)"><Icon name="x" size="sm" /></button>
        </div>
        <div class="quick-ranges">
          <button v-for="item in quickRanges" :key="item.key" type="button" @click="useQuickRange(item.value)">{{ item.label }}</button>
        </div>
        <div class="date-fields">
          <label><span>{{ t('common.distributionAnalytics.startDate') }}</span><input ref="startInputRef" v-model="draftFrom" class="input" type="date" :max="draftTo || today" @input="error = ''" /></label>
          <span class="date-field-arrow" aria-hidden="true">→</span>
          <label><span>{{ t('common.distributionAnalytics.endDate') }}</span><input v-model="draftTo" class="input" type="date" :min="draftFrom" :max="today" @input="error = ''" /></label>
        </div>
        <p v-if="error" class="range-error" role="alert">{{ error }}</p>
        <div class="range-actions"><button type="button" class="btn btn-secondary" @click="close(true)">{{ t('common.cancel') }}</button><button type="button" class="btn btn-primary" @click="applyCustom">{{ t('common.confirm') }}</button></div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { formatDateOnly, rangeForPreset, type DistributionAnalyticsGranularity, type DistributionAnalyticsRangeValue } from './distributionAnalyticsRange'

const props = defineProps<{ modelValue: DistributionAnalyticsRangeValue }>()
const emit = defineEmits<{ (event: 'update:modelValue', value: DistributionAnalyticsRangeValue): void; (event: 'change', value: DistributionAnalyticsRangeValue): void }>()
const { t } = useI18n()
const value = computed(() => props.modelValue)
const rootRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const startInputRef = ref<HTMLInputElement | null>(null)
const open = ref(false)
const draftFrom = ref(value.value.date_from)
const draftTo = ref(value.value.date_to)
const error = ref('')
const today = formatDateOnly(new Date(Date.now() + 8 * 3_600_000))

const granularities = computed(() => [
  { key: 'day' as const, label: t('common.distributionAnalytics.day') },
  { key: 'week' as const, label: t('common.distributionAnalytics.week') },
  { key: 'month' as const, label: t('common.distributionAnalytics.month') },
  { key: 'custom' as const, label: t('common.distributionAnalytics.custom') },
])
const presetLabels = computed(() => value.value.granularity === 'day'
  ? [t('common.distributionAnalytics.today'), t('common.distributionAnalytics.yesterday'), t('common.distributionAnalytics.dayBefore')]
  : value.value.granularity === 'week'
    ? [t('common.distributionAnalytics.thisWeek'), t('common.distributionAnalytics.lastWeek'), t('common.distributionAnalytics.previousWeek')]
    : [t('common.distributionAnalytics.thisMonth'), t('common.distributionAnalytics.lastMonth'), t('common.distributionAnalytics.previousMonth')])
const quickRanges = computed(() => [
  { key: 'today', label: t('common.distributionAnalytics.today'), value: rangeForPreset('day', 0) },
  { key: 'yesterday', label: t('common.distributionAnalytics.yesterday'), value: rangeForPreset('day', 1) },
  { key: 'week', label: t('common.distributionAnalytics.thisWeek'), value: rangeForPreset('week', 0) },
  { key: 'last-week', label: t('common.distributionAnalytics.lastWeek'), value: rangeForPreset('week', 1) },
  { key: 'month', label: t('common.distributionAnalytics.thisMonth'), value: rangeForPreset('month', 0) },
  { key: 'last-month', label: t('common.distributionAnalytics.lastMonth'), value: rangeForPreset('month', 1) },
])

function commit(next: DistributionAnalyticsRangeValue) { emit('update:modelValue', next); emit('change', next) }
function selectGranularity(granularity: DistributionAnalyticsGranularity) {
  if (granularity === 'custom') {
    draftFrom.value = value.value.date_from; draftTo.value = value.value.date_to; error.value = ''; open.value = true
    void nextTick(() => startInputRef.value?.focus())
    return
  }
  open.value = false
  commit(rangeForPreset(granularity, 0))
}
function applyPreset(index: number) { if (value.value.granularity !== 'custom') commit(rangeForPreset(value.value.granularity, index)) }
function useQuickRange(range: DistributionAnalyticsRangeValue) { draftFrom.value = range.date_from; draftTo.value = range.date_to; error.value = '' }
function applyCustom() {
  const from = new Date(`${draftFrom.value}T00:00:00Z`); const to = new Date(`${draftTo.value}T00:00:00Z`)
  const days = Math.floor((to.getTime() - from.getTime()) / 86_400_000) + 1
  if (!draftFrom.value || !draftTo.value || !Number.isFinite(days) || days < 1) { error.value = t('common.distributionAnalytics.invalidRange'); return }
  if (days > 366) { error.value = t('common.distributionAnalytics.rangeTooLong'); return }
  commit({ granularity: 'custom', preset: 'custom', date_from: draftFrom.value, date_to: draftTo.value }); close(false)
}
function close(restoreFocus: boolean) { open.value = false; error.value = ''; if (restoreFocus) void nextTick(() => rootRef.value?.querySelector<HTMLButtonElement>('.analytics-segment button:last-child')?.focus()) }
function onDocumentClick(event: MouseEvent) { if (open.value && rootRef.value && !rootRef.value.contains(event.target as Node)) close(false) }
function onKeydown(event: KeyboardEvent) { if (event.key === 'Escape' && open.value) { event.preventDefault(); close(true) } }
watch(() => [props.modelValue.date_from, props.modelValue.date_to], ([from, to]) => { if (!open.value) { draftFrom.value = from; draftTo.value = to } })
onMounted(() => { document.addEventListener('click', onDocumentClick); document.addEventListener('keydown', onKeydown) })
onUnmounted(() => { document.removeEventListener('click', onDocumentClick); document.removeEventListener('keydown', onKeydown) })
</script>

<style scoped>
.analytics-range{position:relative;display:flex;min-width:0;align-items:center;gap:28px}.analytics-range-group{display:flex;min-width:0;align-items:center;gap:8px}.analytics-range-label{flex:none;color:var(--ui-text-muted);font-size:12px;font-weight:650;white-space:nowrap}.analytics-segment{display:inline-flex;min-height:34px;align-items:center;gap:2px}.analytics-segment button{min-width:40px;min-height:32px;border:1px solid transparent;border-radius:6px;padding:0 9px;color:var(--ui-text-muted);background:transparent;font-size:12px;white-space:nowrap;transition:background-color .16s ease,border-color .16s ease,color .16s ease,box-shadow .16s ease}.analytics-segment button:hover:not([aria-pressed=true]){color:var(--ui-text);background:color-mix(in srgb,var(--ui-surface-subtle) 55%,transparent)}.analytics-segment button[aria-pressed=true]{border-color:var(--ui-border);background:var(--ui-surface-raised);color:var(--ui-text);font-weight:650;box-shadow:0 1px 3px rgb(0 0 0 / .08)}.analytics-segment button:focus-visible,.custom-range-trigger:focus-visible{outline:2px solid var(--ui-focus);outline-offset:1px}.custom-range-trigger{display:flex;min-height:34px;min-width:288px;align-items:center;gap:10px;border:1px solid var(--ui-border);border-radius:7px;background:var(--ui-surface-raised);padding:0 10px;color:var(--ui-text);font-size:12px;font-variant-numeric:tabular-nums;box-shadow:0 1px 2px rgb(0 0 0 / .04)}.custom-range-trigger>span:first-child,.custom-range-trigger>span:nth-child(3){flex:1;text-align:left}.custom-range-trigger svg{color:var(--ui-text-muted)}.range-arrow{color:var(--ui-text-subtle)}.range-popover{position:absolute;z-index:80;top:calc(100% + 8px);right:0;width:min(520px,calc(100vw - 32px));overflow:hidden;border:1px solid var(--ui-border);border-radius:8px;background:var(--ui-surface-raised);box-shadow:var(--ui-shadow-lg)}.range-popover-heading{display:flex;align-items:flex-start;justify-content:space-between;gap:12px;border-bottom:1px solid var(--ui-border);padding:14px 14px 12px}.range-popover-heading strong,.range-popover-heading small{display:block}.range-popover-heading strong{font-size:13px}.range-popover-heading small{margin-top:3px;color:var(--ui-text-muted);font-size:11px}.quick-ranges{display:grid;grid-template-columns:repeat(3,1fr);gap:6px;padding:12px 14px 0}.quick-ranges button{min-height:34px;border-radius:6px;background:var(--ui-surface-subtle);color:var(--ui-text-muted);font-size:12px}.quick-ranges button:hover{color:var(--ui-text);filter:brightness(.98)}.date-fields{display:grid;grid-template-columns:minmax(0,1fr) 18px minmax(0,1fr);align-items:end;gap:8px;padding:14px}.date-fields label span{display:block;margin-bottom:6px;color:var(--ui-text-muted);font-size:11px;font-weight:600}.date-fields .input{width:100%;min-width:0;min-height:38px;font-size:12px}.date-field-arrow{padding-bottom:10px;color:var(--ui-text-subtle);text-align:center}.range-error{margin:-5px 14px 10px;color:rgb(var(--color-danger-foreground));font-size:11px}.range-actions{display:flex;justify-content:flex-end;gap:8px;border-top:1px solid var(--ui-border);padding:10px 14px}.range-actions .btn{min-height:36px}.range-popover-enter-active,.range-popover-leave-active{transition:opacity .16s ease,transform .16s ease}.range-popover-enter-from,.range-popover-leave-to{opacity:0;transform:translateY(-4px)}
@media(max-width:767px){.analytics-range{width:100%;flex-direction:column;align-items:stretch;gap:8px}.analytics-range-group{width:100%;justify-content:space-between}.analytics-range-label{min-width:82px}.analytics-segment{min-width:0;flex:1}.analytics-segment button{min-width:0;min-height:40px;flex:1;padding-inline:4px}.custom-range-trigger{min-width:0;min-height:40px;flex:1}.range-popover{position:fixed;right:16px;bottom:16px;top:auto;left:16px;width:auto}.quick-ranges{grid-template-columns:repeat(2,1fr)}.date-fields{grid-template-columns:1fr}.date-field-arrow{display:none}}
@media(prefers-reduced-motion:reduce){.range-popover-enter-active,.range-popover-leave-active{transition:none}}
</style>
