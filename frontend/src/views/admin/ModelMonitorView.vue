<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1520px] space-y-3 pb-12">
      <header class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 class="text-xl font-semibold text-foreground">{{ t('admin.modelMonitor.title') }}</h1>
          <p class="mt-1 text-sm text-foreground-muted">{{ t('admin.modelMonitor.description') }}</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="inline-flex h-9 rounded-control border border-outline bg-surface p-0.5" :aria-label="t('admin.modelMonitor.resolution')">
            <button v-for="option in resolutionOptions" :key="option.value" type="button" class="min-w-16 rounded-control px-3 text-xs font-medium transition-colors" :class="resolution === option.value ? 'bg-foreground text-surface' : 'text-foreground-muted hover:bg-surface-subtle hover:text-foreground'" @click="setResolution(option.value)">{{ option.label }}</button>
          </div>
          <button type="button" class="btn btn-secondary btn-sm" @click="toggleAllModels">
            <Icon :name="allModelsExpanded ? 'chevronUp' : 'chevronDown'" size="sm" />
            {{ allModelsExpanded ? t('admin.modelMonitor.collapseAll') : t('admin.modelMonitor.expandAll') }}
          </button>
          <button type="button" class="btn btn-secondary btn-icon h-9 w-9" :title="t('common.refresh')" :disabled="loading || refreshing" @click="load()">
            <Icon name="refresh" size="sm" :class="loading || refreshing ? 'animate-spin' : ''" />
          </button>
        </div>
      </header>

      <div v-if="!featureEnabled" class="flex flex-col gap-2 border-l-2 border-warning bg-warning-subtle px-4 py-3 text-sm text-warning-foreground sm:flex-row sm:items-center sm:justify-between">
        <span>{{ t('admin.modelMonitor.disabledBanner') }}</span>
        <router-link to="/admin/settings?tab=features" class="shrink-0 font-medium text-brand hover:underline">{{ t('admin.modelMonitor.openSettings') }}</router-link>
      </div>

      <div class="grid gap-2 rounded-panel border border-outline bg-surface p-2 sm:grid-cols-2 lg:grid-cols-[minmax(280px,1fr)_150px_150px_150px]">
        <div class="relative sm:col-span-2 lg:col-span-1">
          <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
          <input v-model="search" class="input h-10 pl-9" :placeholder="t('admin.modelMonitor.searchPlaceholder')" />
        </div>
        <select v-model="platform" class="input h-10"><option value="">{{ t('admin.modelMonitor.allPlatforms') }}</option><option v-for="item in platforms" :key="item" :value="item">{{ item }}</option></select>
        <select v-model="status" class="input h-10"><option value="">{{ t('admin.modelMonitor.allStatuses') }}</option><option value="operational">{{ t('admin.modelMonitor.operational') }}</option><option value="degraded">{{ t('admin.modelMonitor.degraded') }}</option><option value="failed">{{ t('admin.modelMonitor.failed') }}</option><option value="unknown">{{ t('admin.modelMonitor.unknown') }}</option></select>
        <select v-model="enabled" class="input h-10"><option value="">{{ t('admin.modelMonitor.allStates') }}</option><option value="true">{{ t('admin.modelMonitor.enabledOnly') }}</option><option value="false">{{ t('admin.modelMonitor.disabledOnly') }}</option></select>
      </div>

      <div v-if="loading" class="py-24 text-center text-sm text-foreground-subtle">{{ t('common.loading') }}</div>
      <div v-else-if="filteredRows.length" class="max-h-[calc(100dvh-15rem)] min-h-80 overflow-y-auto overscroll-contain rounded-panel border border-outline bg-surface">
        <div class="sticky top-0 z-20 hidden grid-cols-[minmax(240px,1.5fr)_80px_96px_104px_minmax(180px,1.1fr)_76px] items-center gap-3 border-b border-outline bg-surface-subtle px-4 py-2 text-[10px] font-medium text-foreground-subtle shadow-sm lg:grid">
          <span>{{ t('admin.modelMonitor.model') }}</span>
          <button type="button" class="inline-flex items-center justify-end gap-1 text-right transition-colors hover:text-foreground" @click="toggleMetricSort('tps')">TPS<Icon :name="sortIcon('tps')" size="xs" /></button>
          <button type="button" class="inline-flex items-center justify-end gap-1 text-right transition-colors hover:text-foreground" @click="toggleMetricSort('ttft')">TTFT<Icon :name="sortIcon('ttft')" size="xs" /></button>
          <button type="button" class="inline-flex items-center justify-end gap-1 text-right transition-colors hover:text-foreground" @click="toggleMetricSort('latency')">{{ t('admin.modelMonitor.averageLatency') }}<Icon :name="sortIcon('latency')" size="xs" /></button>
          <button type="button" class="inline-flex items-center justify-end gap-1 text-right transition-colors hover:text-foreground" @click="toggleMetricSort('successRate')">{{ t('admin.modelMonitor.successRate') }}<Icon :name="sortIcon('successRate')" size="xs" /></button>
          <span class="text-right">{{ t('admin.modelMonitor.actions') }}</span>
        </div>
        <article v-for="row in filteredRows" :key="modelKey(row)" class="border-b border-outline last:border-b-0">
          <header class="grid min-w-0 grid-cols-[minmax(0,1fr)_76px] items-center gap-2 px-3 py-3 transition-colors hover:bg-surface-subtle/60 lg:grid-cols-[minmax(240px,1.5fr)_80px_96px_104px_minmax(180px,1.1fr)_76px] lg:gap-3 lg:px-4 lg:py-2.5">
            <div class="flex min-w-0 items-center gap-2">
              <button type="button" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-control text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground" :aria-expanded="isExpanded(row)" @click="toggleExpanded(row)">
                <Icon name="chevronRight" size="xs" class="transition-transform" :class="isExpanded(row) ? 'rotate-90' : ''" />
              </button>
              <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-control bg-surface-subtle" :class="platformIconClass(row.platform)">
                <PlatformIcon :platform="row.platform as GroupPlatform" size="md" />
              </span>
              <button type="button" class="min-w-0 flex-1 text-left" @click="toggleExpanded(row)">
                <span class="flex min-w-0 items-center gap-2">
                  <strong class="truncate font-mono text-sm font-semibold text-foreground" :title="row.model">{{ row.model }}</strong>
                  <span v-if="row.label" class="badge max-w-20 shrink-0 truncate bg-black text-white">{{ row.label }}</span>
                  <span v-if="!row.catalog_available" class="badge shrink-0 bg-warning-subtle text-warning-foreground">{{ t('admin.modelMonitor.unavailable') }}</span>
                </span>
                <span class="mt-0.5 block truncate text-[10px] text-foreground-subtle">{{ row.platform }} · {{ t('admin.modelMonitor.groupTotal', { count: row.groups.length }) }}</span>
              </button>
            </div>
            <div class="hidden text-right font-mono text-xs tabular-nums text-foreground lg:block">{{ formatTPS(row.summary?.metrics?.tps) }}</div>
            <div class="hidden text-right font-mono text-xs tabular-nums text-foreground lg:block">{{ formatLatency(row.summary?.metrics?.ttft_ms) }}</div>
            <div class="hidden text-right font-mono text-xs tabular-nums text-foreground lg:block">{{ formatLatency(row.summary?.metrics?.average_latency_ms) }}</div>
            <div class="hidden justify-end lg:flex"><SuccessRateTimeline :buckets="row.summary?.metrics?.buckets" :success-rate="row.summary?.metrics?.success_rate" :resolution="resolution" /></div>
            <div class="flex justify-end gap-1">
              <button type="button" class="btn btn-ghost btn-icon h-7 w-7" :title="t('admin.modelMonitor.presentation')" @click="openPresentation(row)"><Icon name="edit" size="xs" /></button>
              <button type="button" class="btn btn-ghost btn-icon h-7 w-7" :title="t('admin.modelMonitor.history')" :disabled="!row.id" @click="openHistory(row)"><Icon name="clock" size="xs" /></button>
            </div>
            <div class="col-span-2 grid grid-cols-3 gap-px overflow-hidden rounded-control bg-outline lg:hidden">
              <ModelMetric label="TPS" :value="formatTPS(row.summary?.metrics?.tps)" />
              <ModelMetric label="TTFT" :value="formatLatency(row.summary?.metrics?.ttft_ms)" />
              <ModelMetric :label="t('admin.modelMonitor.averageLatency')" :value="formatLatency(row.summary?.metrics?.average_latency_ms)" />
            </div>
            <div class="col-span-2 flex items-center justify-between lg:hidden">
              <span class="text-[10px] text-foreground-subtle">{{ t('admin.modelMonitor.successRate') }}</span>
              <SuccessRateTimeline :buckets="row.summary?.metrics?.buckets" :success-rate="row.summary?.metrics?.success_rate" :resolution="resolution" />
            </div>
          </header>

          <section v-if="isExpanded(row)" class="border-t border-outline bg-surface-subtle/60 lg:pl-9">
            <div v-if="row.groups.length" class="divide-y divide-outline">
              <div class="hidden grid-cols-[minmax(150px,1fr)_72px_86px_96px_minmax(175px,1.2fr)_88px_104px_52px_72px] items-center gap-3 px-4 py-2 text-[9px] font-medium text-foreground-subtle xl:grid">
                <span>{{ t('admin.modelMonitor.group') }}</span><span class="text-right">TPS</span><span class="text-right">TTFT</span><span class="text-right">{{ t('admin.modelMonitor.averageLatency') }}</span><span>{{ t('admin.modelMonitor.successRate') }}</span><span class="text-right">{{ t('admin.modelMonitor.probeCost') }}</span><span>{{ t('admin.modelMonitor.interval') }}</span><span class="text-center">{{ t('admin.modelMonitor.enabled') }}</span><span class="text-right">{{ t('admin.modelMonitor.actions') }}</span>
              </div>
              <div v-for="group in row.groups" :key="group.group_id" class="px-3 py-3 transition-colors hover:bg-surface sm:px-4 xl:grid xl:min-h-14 xl:grid-cols-[minmax(150px,1fr)_72px_86px_96px_minmax(175px,1.2fr)_88px_104px_52px_72px] xl:items-center xl:gap-3 xl:py-2">
                <div class="flex min-w-0 items-center justify-between gap-3 xl:block">
                  <div class="flex min-w-0 items-center gap-2"><span class="h-2 w-2 shrink-0 rounded-full" :class="statusDotClass(group)"></span><span class="truncate text-sm font-medium text-foreground" :title="group.name">{{ group.name }}</span></div>
                  <div class="flex shrink-0 items-center gap-2 xl:hidden"><span class="text-[10px] text-foreground-subtle">{{ t('admin.modelMonitor.enabled') }}</span><Toggle :model-value="group.enabled" :disabled="isGroupSaving(row, group)" @update:model-value="saveGroup(row, group, $event)" /></div>
                </div>
                <dl class="mt-3 grid grid-cols-3 gap-px bg-outline xl:contents">
                  <MetricCell label="TPS" :value="formatTPS(group.metrics?.tps)" />
                  <MetricCell label="TTFT" :value="formatLatency(group.metrics?.ttft_ms)" />
                  <MetricCell :label="t('admin.modelMonitor.averageLatency')" :value="formatLatency(group.metrics?.average_latency_ms)" />
                </dl>
                <div class="mt-3 flex min-w-0 items-center justify-between gap-3 xl:mt-0"><span class="text-[10px] text-foreground-subtle xl:hidden">{{ t('admin.modelMonitor.successRate') }}</span><SuccessRateTimeline :buckets="group.metrics?.buckets" :success-rate="group.metrics?.success_rate" :resolution="resolution" /></div>
                <div class="mt-3 flex items-center justify-between xl:mt-0 xl:block xl:text-right"><span class="text-[10px] text-foreground-subtle xl:hidden">{{ t('admin.modelMonitor.probeCost') }}</span><span class="font-mono text-xs tabular-nums text-foreground-muted">{{ formatCost(group.metrics?.probe_cost) }}</span></div>
                <div class="mt-3 xl:mt-0">
                  <Select
                    :model-value="group.interval_seconds || 300"
                    :options="intervalSelectOptions"
                    :searchable="false"
                    :disabled="isGroupSaving(row, group)"
                    class="w-full xl:w-[104px]"
                    :aria-label="t('admin.modelMonitor.interval')"
                    @update:model-value="saveGroup(row, group, group.enabled, Number($event))"
                  />
                </div>
                <div class="hidden justify-center xl:flex"><Toggle :model-value="group.enabled" :disabled="isGroupSaving(row, group)" @update:model-value="saveGroup(row, group, $event)" /></div>
                <div class="mt-3 flex justify-end gap-1 xl:mt-0">
                  <button type="button" class="btn btn-ghost btn-icon h-8 w-8" :title="t('admin.modelMonitor.runNow')" :disabled="!featureEnabled || runningKey !== null" @click="runGroup(row, group)"><Icon name="play" size="sm" :class="runningKey === groupKey(row, group) ? 'animate-pulse' : ''" /></button>
                  <button type="button" class="btn btn-ghost btn-icon h-8 w-8" :title="t('admin.modelMonitor.history')" :disabled="!row.id" @click="openHistory(row, group)"><Icon name="clock" size="sm" /></button>
                </div>
              </div>
            </div>
            <div v-else class="px-5 py-8 text-center text-sm text-foreground-subtle">{{ t('admin.modelMonitor.noGroups') }}</div>
          </section>
        </article>
      </div>
      <EmptyState v-else :title="t('admin.modelMonitor.empty')" />
    </div>

    <BaseDialog :show="historyOpen" :title="t('admin.modelMonitor.historyTitle', { model: historyRow?.model || '' })" width="wide" @close="historyOpen = false">
      <div class="mb-4 flex items-center justify-between gap-3">
        <select v-model.number="historyGroupID" class="input h-9 max-w-56 text-sm"><option :value="0">{{ t('admin.modelMonitor.allGroups') }}</option><option v-for="group in historyRow?.groups || []" :key="group.group_id" :value="group.group_id">{{ group.name }}</option></select>
        <span class="text-xs text-foreground-subtle">{{ t('admin.modelMonitor.historyCount', { count: filteredHistoryItems.length }) }}</span>
      </div>
      <div v-if="historyLoading" class="py-12 text-center text-sm text-foreground-subtle">{{ t('common.loading') }}</div>
      <div v-else-if="filteredHistoryItems.length" class="max-h-[62vh] divide-y divide-outline overflow-y-auto border-y border-outline">
        <article v-for="item in filteredHistoryItems" :key="item.id" class="grid gap-3 py-3 sm:grid-cols-[110px_minmax(100px,1fr)_90px_90px_160px] sm:items-center">
          <div class="flex items-center gap-2 text-sm font-medium" :class="statusTextClass(item.status)"><span class="h-2 w-2 rounded-full bg-current"></span>{{ statusLabel(item.status) }}</div>
          <div class="min-w-0"><div class="truncate text-sm text-foreground">{{ item.group_name || '—' }}</div><p v-if="item.message" class="mt-0.5 truncate text-[10px] text-foreground-subtle" :title="item.message">{{ item.message }}</p></div>
          <div><span class="text-[9px] text-foreground-subtle">{{ t('admin.modelMonitor.latency') }}</span><div class="font-mono text-xs">{{ formatLatency(item.latency_ms) }}</div></div>
          <div><span class="text-[9px] text-foreground-subtle">{{ t('admin.modelMonitor.probeCost') }}</span><div class="font-mono text-xs">{{ formatCost(item.probe_cost) }}</div></div>
          <time class="text-xs text-foreground-subtle sm:text-right">{{ formatTime(item.checked_at) }}</time>
        </article>
      </div>
      <div v-else class="py-12 text-center text-sm text-foreground-subtle">{{ t('admin.modelMonitor.noHistory') }}</div>
    </BaseDialog>

    <BaseDialog :show="presentationOpen" :title="t('admin.modelMonitor.presentationTitle', { model: presentationRow?.model || '' })" @close="presentationOpen = false">
      <div class="space-y-5"><div><label class="mb-1.5 block text-sm font-medium">{{ t('admin.modelMonitor.label') }}</label><input v-model="presentationLabel" class="input" maxlength="12" /></div><div><label class="mb-1.5 block text-sm font-medium">{{ t('admin.modelMonitor.displayOrder') }}</label><input v-model.number="presentationOrder" type="number" min="0" max="9999" class="input" /><p class="mt-1.5 text-xs leading-5 text-foreground-subtle">{{ t('admin.modelMonitor.displayOrderHint') }}</p></div></div>
      <template #footer><div class="flex justify-end gap-3"><button type="button" class="btn btn-secondary" @click="presentationOpen = false">{{ t('common.cancel') }}</button><button type="button" class="btn btn-primary" :disabled="presentationSaving" @click="savePresentation">{{ t('common.save') }}</button></div></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { ModelMonitorGroupOption, ModelMonitorHistoryItem, ModelMonitorResolution, ModelMonitorRow } from '@/api/admin/modelMonitor'
import type { GroupPlatform } from '@/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Toggle from '@/components/common/Toggle.vue'
import Select from '@/components/common/Select.vue'
import SuccessRateTimeline from '@/components/common/SuccessRateTimeline.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import { platformIconClass } from '@/utils/platformColors'
import { compareModelMonitorRows, type ModelMonitorMetricSortField } from './modelMonitorSorting'

const ModelMetric = defineComponent({ props: { label: String, value: String }, setup: props => () => h('div', { class: 'min-w-0 bg-surface px-3 py-2.5' }, [h('div', { class: 'truncate text-[9px] text-foreground-subtle' }, props.label), h('div', { class: 'mt-0.5 truncate font-mono text-sm font-semibold tabular-nums text-foreground' }, props.value)]) })
const MetricCell = defineComponent({ props: { label: String, value: String }, setup: props => () => h('div', { class: 'min-w-0 bg-surface px-2 py-2 text-center xl:bg-transparent xl:p-0 xl:text-right' }, [h('dt', { class: 'text-[9px] text-foreground-subtle xl:hidden' }, props.label), h('dd', { class: 'mt-0.5 truncate font-mono text-xs tabular-nums text-foreground xl:mt-0' }, props.value)]) })

const { t } = useI18n()
const appStore = useAppStore()
const rows = ref<ModelMonitorRow[]>([])
const loading = ref(false)
const refreshing = ref(false)
const search = ref('')
const platform = ref('')
const status = ref('')
const enabled = ref('')
const sortField = ref<ModelMonitorMetricSortField | null>(null)
const sortDirection = ref<'asc' | 'desc'>('desc')
const resolution = ref<ModelMonitorResolution>(localStorage.getItem('usa0:model-monitor-resolution') === 'minute' ? 'minute' : 'hour')
const expandedKeys = ref(new Set<string>())
const savingGroups = ref(new Set<string>())
const runningKey = ref<string | null>(null)
const historyOpen = ref(false)
const historyLoading = ref(false)
const historyRow = ref<ModelMonitorRow | null>(null)
const historyGroupID = ref(0)
const historyItems = ref<ModelMonitorHistoryItem[]>([])
const presentationOpen = ref(false)
const presentationSaving = ref(false)
const presentationRow = ref<ModelMonitorRow | null>(null)
const presentationLabel = ref('')
const presentationOrder = ref(0)
let autoRefreshTimer: number | null = null

const intervalOptions = [60, 300, 600, 900, 1800, 3600]
const intervalSelectOptions = computed(() => intervalOptions.map(value => ({ value, label: intervalLabel(value) })))
const featureEnabled = computed(() => appStore.cachedPublicSettings?.model_monitor_enabled === true)
const platforms = computed(() => [...new Set(rows.value.map(row => row.platform))].sort())
const resolutionOptions = computed(() => [{ value: 'minute' as const, label: t('admin.modelMonitor.minute') }, { value: 'hour' as const, label: t('admin.modelMonitor.hour') }])
const filteredHistoryItems = computed(() => historyGroupID.value === 0 ? historyItems.value : historyItems.value.filter(item => item.group_id === historyGroupID.value))
const filteredRows = computed(() => rows.value.filter(row => {
  const query = search.value.trim().toLowerCase()
  if (query && !row.model.toLowerCase().includes(query) && !row.groups.some(group => group.name.toLowerCase().includes(query))) return false
  if (platform.value && row.platform !== platform.value) return false
  if (status.value && modelStatus(row) !== status.value) return false
  if (enabled.value && String(row.groups.some(group => group.enabled)) !== enabled.value) return false
  return true
}).sort((a, b) => compareModelMonitorRows(a, b, sortField.value, sortDirection.value)))
const allModelsExpanded = computed(() => filteredRows.value.length > 0 && filteredRows.value.every(row => expandedKeys.value.has(modelKey(row))))

function modelKey(row: ModelMonitorRow) { return `${row.platform}:${row.model}` }
function groupKey(row: ModelMonitorRow, group: ModelMonitorGroupOption) { return `${modelKey(row)}:${group.group_id}` }
function isExpanded(row: ModelMonitorRow) { return expandedKeys.value.has(modelKey(row)) }
function toggleExpanded(row: ModelMonitorRow) { const next = new Set(expandedKeys.value); const key = modelKey(row); next.has(key) ? next.delete(key) : next.add(key); expandedKeys.value = next }
function toggleAllModels() { expandedKeys.value = allModelsExpanded.value ? new Set() : new Set(filteredRows.value.map(modelKey)) }
function isGroupSaving(row: ModelMonitorRow, group: ModelMonitorGroupOption) { return savingGroups.value.has(groupKey(row, group)) }
function modelStatus(row: ModelMonitorRow) { const rate = row.summary?.metrics?.success_rate; return rate == null ? 'unknown' : rate >= 99.9 ? 'operational' : rate >= 90 ? 'degraded' : 'failed' }
function toggleMetricSort(field: ModelMonitorMetricSortField) {
  if (sortField.value === field) sortDirection.value = sortDirection.value === 'desc' ? 'asc' : 'desc'
  else { sortField.value = field; sortDirection.value = 'desc' }
}
function sortIcon(field: ModelMonitorMetricSortField) {
  if (sortField.value !== field) return 'chevronDown'
  return sortDirection.value === 'desc' ? 'chevronDown' : 'chevronUp'
}

async function load(silent = false) {
  if (loading.value || refreshing.value) return false
  silent ? refreshing.value = true : loading.value = true
  try {
    rows.value = (await adminAPI.modelMonitor.list(resolution.value)).items || []
    return true
  } catch (error) { if (!silent) appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.loadError'))); return false }
  finally { loading.value = false; refreshing.value = false }
}
async function setResolution(value: ModelMonitorResolution) { if (resolution.value === value) return; const previous = resolution.value; resolution.value = value; if (await load()) localStorage.setItem('usa0:model-monitor-resolution', value); else resolution.value = previous }
async function saveGroup(row: ModelMonitorRow, group: ModelMonitorGroupOption, nextEnabled = group.enabled, nextInterval = group.interval_seconds || 300) { const key = groupKey(row, group); savingGroups.value = new Set(savingGroups.value).add(key); try { await adminAPI.modelMonitor.updateGroupConfig({ platform: row.platform, model: row.model, group_id: group.group_id, enabled: nextEnabled, interval_seconds: nextInterval }); group.enabled = nextEnabled; group.interval_seconds = nextInterval; group.selected = true; row.configured = true } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.saveError'))); await load(true) } finally { const done = new Set(savingGroups.value); done.delete(key); savingGroups.value = done } }
async function runGroup(row: ModelMonitorRow, group: ModelMonitorGroupOption) { runningKey.value = groupKey(row, group); try { await adminAPI.modelMonitor.runNow(row.platform, row.model, group.group_id); appStore.showSuccess(t('admin.modelMonitor.runSuccess')); await load(true) } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.runError'))) } finally { runningKey.value = null } }
async function openHistory(row: ModelMonitorRow, group?: ModelMonitorGroupOption) { if (!row.id) return; historyRow.value = row; historyGroupID.value = group?.group_id ?? 0; historyOpen.value = true; historyLoading.value = true; try { historyItems.value = (await adminAPI.modelMonitor.history(row.id)).items || [] } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.loadError'))) } finally { historyLoading.value = false } }
function openPresentation(row: ModelMonitorRow) { presentationRow.value = row; presentationLabel.value = row.label; presentationOrder.value = row.display_order; presentationOpen.value = true }
async function savePresentation() { const row = presentationRow.value; if (!row) return; presentationSaving.value = true; try { await adminAPI.modelMonitor.updateConfig({ platform: row.platform, model: row.model, enabled: row.enabled, interval_seconds: Math.max(60, row.interval_seconds || 300), display_order: Math.max(0, Math.min(9999, Math.trunc(presentationOrder.value || 0))), label: presentationLabel.value.trim() }); presentationOpen.value = false; await load(true); appStore.showSuccess(t('admin.modelMonitor.presentationSaved')) } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.saveError'))) } finally { presentationSaving.value = false } }

function statusDotClass(group: ModelMonitorGroupOption) { const rate = group.metrics?.success_rate; return rate == null ? 'bg-surface-emphasis' : rate >= 99.9 ? 'bg-success' : rate >= 90 ? 'bg-warning' : 'bg-danger' }
function statusTextClass(value?: string) { return value === 'operational' ? 'text-success-foreground' : value === 'degraded' ? 'text-warning-foreground' : value === 'failed' || value === 'error' ? 'text-danger-foreground' : 'text-foreground-subtle' }
function statusLabel(value?: string) { return value ? t(`admin.modelMonitor.${value}`) : t('admin.modelMonitor.unknown') }
function formatTPS(value?: number | null) { return value == null ? '—' : `${value.toFixed(value < 10 ? 2 : 1)} t/s` }
function formatLatency(value?: number | null) { if (value == null) return '—'; return value >= 1000 ? `${(value / 1000).toFixed(2)}s` : `${Math.round(value)}ms` }
function formatCost(value?: number | null) { return value == null ? '—' : `$${value.toFixed(value > 0 && value < 0.01 ? 6 : 4)}` }
function intervalLabel(value: number) { return t('admin.modelMonitor.minutes', { value: value / 60 }) }
function formatTime(value: string) { return new Date(value).toLocaleString() }

onMounted(() => { void load(); autoRefreshTimer = window.setInterval(() => { if (document.visibilityState === 'visible') void load(true) }, 30_000) })
onBeforeUnmount(() => { if (autoRefreshTimer != null) window.clearInterval(autoRefreshTimer) })
</script>
