<template>
  <AppLayout>
    <TablePageLayout>
      <template #actions>
        <div v-if="!featureEnabled" class="flex items-center justify-between gap-4 border border-warning-border bg-warning-subtle px-4 py-3 text-sm text-warning-foreground">
          <span>{{ t('admin.modelMonitor.disabledBanner') }}</span>
          <router-link to="/admin/settings?tab=features" class="shrink-0 font-medium text-brand hover:underline">
            {{ t('admin.modelMonitor.openSettings') }}
          </router-link>
        </div>
      </template>

      <template #filters>
        <div class="space-y-2 md:hidden">
          <div class="flex min-w-0 items-center gap-2">
            <div class="relative min-w-0 flex-1">
              <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
              <input v-model="search" class="input pl-9" :placeholder="t('admin.modelMonitor.searchPlaceholder')" />
            </div>
            <button type="button" class="btn btn-secondary btn-icon shrink-0" :title="t('common.refresh')" :disabled="loading || refreshing" @click="load()">
              <Icon name="refresh" size="sm" :class="loading || refreshing ? 'animate-spin' : ''" />
            </button>
          </div>
          <details class="group rounded-panel border border-outline bg-surface">
            <summary class="flex min-h-10 cursor-pointer list-none items-center justify-between gap-3 px-3 py-2 text-sm font-medium text-foreground-muted">
              <span class="flex items-center gap-2">
                <Icon name="filter" size="sm" />
                {{ t('common.filter') }}
                <span v-if="activeFilterCount" class="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-brand px-1 text-[10px] font-semibold text-brand-foreground">
                  {{ activeFilterCount }}
                </span>
              </span>
              <Icon name="chevronDown" size="xs" class="transition-transform group-open:rotate-180" />
            </summary>
            <div class="grid grid-cols-1 gap-2 border-t border-outline p-3 sm:grid-cols-2">
              <select v-model="platform" class="input">
                <option value="">{{ t('admin.modelMonitor.allPlatforms') }}</option>
                <option v-for="item in platforms" :key="item" :value="item">{{ item }}</option>
              </select>
              <select v-model="status" class="input">
                <option value="">{{ t('admin.modelMonitor.allStatuses') }}</option>
                <option v-for="item in statuses" :key="item" :value="item">{{ statusLabel(item) }}</option>
              </select>
              <select v-model="enabled" class="input">
                <option value="">{{ t('admin.modelMonitor.allStates') }}</option>
                <option value="true">{{ t('admin.modelMonitor.enabledOnly') }}</option>
                <option value="false">{{ t('admin.modelMonitor.disabledOnly') }}</option>
              </select>
              <select v-model="sortMode" class="input" :aria-label="t('admin.modelMonitor.sortLabel')">
                <option value="priority">{{ t('admin.modelMonitor.sortPriority') }}</option>
                <option value="name">{{ t('admin.modelMonitor.sortName') }}</option>
                <option value="status">{{ t('admin.modelMonitor.sortStatus') }}</option>
              </select>
              <span class="inline-flex min-h-9 items-center gap-1.5 text-xs text-foreground-subtle sm:col-span-2">
                <span class="h-1.5 w-1.5 rounded-full" :class="refreshing ? 'animate-pulse bg-brand' : 'bg-success'"></span>
                {{ t('admin.modelMonitor.autoRefresh') }}
              </span>
            </div>
          </details>
        </div>

        <div class="hidden flex-wrap items-center gap-2 md:flex">
          <div class="relative min-w-56 flex-1">
            <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
            <input v-model="search" class="input pl-9" :placeholder="t('admin.modelMonitor.searchPlaceholder')" />
          </div>
          <select v-model="platform" class="input w-auto min-w-36">
            <option value="">{{ t('admin.modelMonitor.allPlatforms') }}</option>
            <option v-for="item in platforms" :key="item" :value="item">{{ item }}</option>
          </select>
          <select v-model="status" class="input w-auto min-w-36">
            <option value="">{{ t('admin.modelMonitor.allStatuses') }}</option>
            <option v-for="item in statuses" :key="item" :value="item">{{ statusLabel(item) }}</option>
          </select>
          <select v-model="enabled" class="input w-auto min-w-36">
            <option value="">{{ t('admin.modelMonitor.allStates') }}</option>
            <option value="true">{{ t('admin.modelMonitor.enabledOnly') }}</option>
            <option value="false">{{ t('admin.modelMonitor.disabledOnly') }}</option>
          </select>
          <select v-model="sortMode" class="input w-auto min-w-36" :aria-label="t('admin.modelMonitor.sortLabel')">
            <option value="priority">{{ t('admin.modelMonitor.sortPriority') }}</option>
            <option value="name">{{ t('admin.modelMonitor.sortName') }}</option>
            <option value="status">{{ t('admin.modelMonitor.sortStatus') }}</option>
          </select>
          <span class="inline-flex min-h-10 items-center gap-1.5 px-1 text-xs text-foreground-subtle">
            <span class="h-1.5 w-1.5 rounded-full" :class="refreshing ? 'animate-pulse bg-brand' : 'bg-success'"></span>
            {{ t('admin.modelMonitor.autoRefresh') }}
          </span>
          <button type="button" class="btn btn-secondary btn-icon" :title="t('common.refresh')" :disabled="loading || refreshing" @click="load()">
            <Icon name="refresh" size="sm" :class="loading || refreshing ? 'animate-spin' : ''" />
          </button>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="filteredRows" :loading="loading">
          <template #mobile-card="{ row }">
            <article class="space-y-3">
              <header class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex min-w-0 items-center gap-1.5">
                    <h3 class="truncate text-sm font-semibold text-foreground">{{ row.model }}</h3>
                    <span v-if="row.label" class="inline-flex h-[18px] max-w-20 shrink-0 items-center truncate rounded-control bg-black px-1.5 text-[9px] font-bold text-white">{{ row.label }}</span>
                  </div>
                  <p class="mt-0.5 text-[10px] uppercase text-foreground-subtle">{{ row.platform }}</p>
                </div>
                <span class="inline-flex shrink-0 items-center gap-1.5 text-xs font-medium" :class="statusTextClass(row.summary?.status)">
                  <span class="h-2 w-2 rounded-full bg-current"></span>
                  {{ statusLabel(row.summary?.status) }}
                </span>
              </header>

              <div v-if="!row.catalog_available" class="rounded-control bg-warning-subtle px-2.5 py-1.5 text-xs text-warning-foreground">
                {{ t('admin.modelMonitor.unavailable') }}
              </div>

              <dl class="grid grid-cols-2 gap-px overflow-hidden rounded-panel border border-outline bg-outline">
                <div class="bg-surface-subtle px-3 py-2">
                  <dt class="text-[10px] text-foreground-subtle">{{ t('admin.modelMonitor.availability') }}</dt>
                  <dd class="mt-0.5 font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatAvailability(row.summary?.availability_7d) }}</dd>
                </div>
                <div class="bg-surface-subtle px-3 py-2">
                  <dt class="text-[10px] text-foreground-subtle">{{ t('admin.modelMonitor.latency') }}</dt>
                  <dd class="mt-0.5 font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatLatency(row.summary?.latency_ms) }}</dd>
                </div>
              </dl>

              <div class="flex h-5 items-end gap-0.5" :aria-label="t('admin.modelMonitor.timeline')">
                <span v-for="(point, index) in timeline(row)" :key="`mobile-${point.checked_at}-${index}`" class="h-4 min-w-0 flex-1 rounded-[2px]" :class="timelineClass(point.status)" :title="timelineTitle(point)" />
                <span v-for="index in Math.max(0, 20 - timeline(row).length)" :key="`mobile-empty-${index}`" class="h-4 min-w-0 flex-1 rounded-[2px] bg-surface-emphasis" />
              </div>

              <button type="button" class="flex w-full items-center justify-between gap-2 rounded-control bg-surface-subtle px-3 py-2 text-left text-xs text-foreground-muted hover:text-brand" @click="openGroups(row)">
                <span class="truncate">{{ groupSummary(row) }}</span>
                <Icon name="cog" size="xs" class="shrink-0" />
              </button>

              <div class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
                <select :value="row.interval_seconds" class="input h-9 min-w-0 py-1 text-xs" @change="changeInterval(row, $event)">
                  <option v-for="option in intervalOptions" :key="option" :value="option">{{ intervalLabel(option) }}</option>
                </select>
                <Toggle :modelValue="row.enabled" :aria-label="`${t('admin.modelMonitor.enabled')}: ${row.model}`" @update:modelValue="toggleRow(row, $event)" />
              </div>

              <footer class="grid grid-cols-3 gap-2 border-t border-outline pt-3">
                <button type="button" class="btn btn-secondary px-2" :title="t('admin.modelMonitor.runNow')" :disabled="!featureEnabled || runningKey !== null || !row.catalog_available" @click="runRow(row)">
                  <Icon name="play" size="sm" />
                </button>
                <button type="button" class="btn btn-secondary px-2" :title="t('admin.modelMonitor.configureGroups')" :disabled="!row.catalog_available" @click="openGroups(row)">
                  <Icon name="cog" size="sm" />
                </button>
                <button type="button" class="btn btn-secondary px-2" :title="t('admin.modelMonitor.history')" :disabled="!row.configured" @click="openHistory(row)">
                  <Icon name="clock" size="sm" />
                </button>
              </footer>
            </article>
          </template>

          <template #cell-model="{ row }">
            <div class="min-w-48">
              <div class="flex items-center gap-2">
                <span class="font-medium text-foreground">{{ row.model }}</span>
                <span v-if="row.label" class="inline-flex h-[18px] max-w-24 items-center truncate rounded-control bg-black px-2 text-[10px] font-bold leading-none text-white" :title="row.label">{{ row.label }}</span>
                <span v-if="!row.catalog_available" class="badge bg-warning-subtle text-warning-foreground">{{ t('admin.modelMonitor.unavailable') }}</span>
              </div>
              <span class="mt-0.5 block text-xs uppercase text-foreground-subtle">{{ row.platform }}</span>
            </div>
          </template>
          <template #cell-status="{ row }">
            <span class="inline-flex items-center gap-1.5 text-sm" :class="statusTextClass(row.summary?.status)">
              <span class="h-2 w-2 rounded-full bg-current"></span>{{ statusLabel(row.summary?.status) }}
            </span>
          </template>
          <template #cell-groups="{ row }">
            <button type="button" class="inline-flex max-w-52 items-center gap-1.5 text-left text-sm text-foreground-muted hover:text-brand" @click="openGroups(row)">
              <span class="truncate">{{ groupSummary(row) }}</span>
              <Icon name="cog" size="xs" class="shrink-0" />
            </button>
          </template>
          <template #cell-presentation="{ row }">
            <button type="button" class="group flex min-w-28 items-center gap-2 text-left" @click="openPresentation(row)">
              <span class="font-mono text-xs tabular-nums text-foreground-muted">{{ row.display_order }}</span>
              <span class="text-xs text-foreground-subtle group-hover:text-brand">{{ row.label || t('admin.modelMonitor.noLabel') }}</span>
              <Icon name="edit" size="xs" class="shrink-0 text-foreground-subtle group-hover:text-brand" />
            </button>
          </template>
          <template #cell-availability="{ row }">
            <span class="font-mono text-sm tabular-nums text-foreground-muted">{{ formatAvailability(row.summary?.availability_7d) }}</span>
          </template>
          <template #cell-latency="{ row }">
            <span class="font-mono text-sm tabular-nums text-foreground-muted">{{ formatLatency(row.summary?.latency_ms) }}</span>
          </template>
          <template #cell-timeline="{ row }">
            <div class="flex h-5 w-40 items-end gap-0.5" :aria-label="t('admin.modelMonitor.timeline')">
              <span v-for="(point, index) in timeline(row)" :key="`${point.checked_at}-${index}`" class="h-4 min-w-0 flex-1 rounded-[2px]" :class="timelineClass(point.status)" :title="timelineTitle(point)" />
              <span v-for="index in Math.max(0, 20 - timeline(row).length)" :key="`empty-${index}`" class="h-4 min-w-0 flex-1 rounded-[2px] bg-surface-emphasis" />
            </div>
          </template>
          <template #cell-interval="{ row }">
            <select :value="row.interval_seconds" class="input h-8 w-28 py-1 text-xs" @change="changeInterval(row, $event)">
              <option v-for="option in intervalOptions" :key="option" :value="option">{{ intervalLabel(option) }}</option>
            </select>
          </template>
          <template #cell-enabled="{ row }">
            <Toggle :modelValue="row.enabled" :aria-label="`${t('admin.modelMonitor.enabled')}: ${row.model}`" @update:modelValue="toggleRow(row, $event)" />
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-1">
              <button type="button" class="btn btn-ghost btn-icon" :title="t('admin.modelMonitor.runNow')" :disabled="!featureEnabled || runningKey !== null || !row.catalog_available" @click="runRow(row)">
                <Icon name="play" size="sm" />
              </button>
              <button type="button" class="btn btn-ghost btn-icon" :title="t('admin.modelMonitor.configureGroups')" :disabled="!row.catalog_available" @click="openGroups(row)">
                <Icon name="cog" size="sm" />
              </button>
              <button type="button" class="btn btn-ghost btn-icon" :title="t('admin.modelMonitor.history')" :disabled="!row.configured" @click="openHistory(row)">
                <Icon name="clock" size="sm" />
              </button>
            </div>
          </template>
          <template #empty><EmptyState :title="t('admin.modelMonitor.empty')" /></template>
        </DataTable>
      </template>
    </TablePageLayout>

    <BaseDialog :show="groupsOpen" :title="t('admin.modelMonitor.groupsTitle', { model: groupsRow?.model || '' })" width="wide" @close="groupsOpen = false">
      <p class="mb-4 text-sm text-foreground-muted">{{ t('admin.modelMonitor.groupsHint') }}</p>
      <div class="max-h-[55vh] space-y-2 overflow-auto pr-1">
        <div v-for="(group, index) in editableGroups" :key="group.group_id" class="flex min-h-14 items-center gap-3 border-b border-outline px-2 py-2">
          <input :id="`monitor-group-${group.group_id}`" v-model="group.selected" type="checkbox" class="h-4 w-4 accent-brand" />
          <label :for="`monitor-group-${group.group_id}`" class="min-w-0 flex-1 cursor-pointer">
            <span class="block truncate text-sm font-medium text-foreground">{{ group.name }}</span>
            <span class="text-xs text-foreground-subtle">{{ t('admin.modelMonitor.groupRate', { value: formatRate(group.rate_multiplier) }) }}</span>
          </label>
          <span v-if="group.selected" class="w-8 text-center font-mono text-xs text-foreground-subtle">{{ selectedPriority(index) }}</span>
          <div class="flex gap-1">
            <button type="button" class="btn btn-ghost btn-icon h-8 w-8" :title="t('admin.modelMonitor.moveUp')" :disabled="!group.selected || !canMove(index, -1)" @click="moveGroup(index, -1)"><Icon name="arrowUp" size="sm" /></button>
            <button type="button" class="btn btn-ghost btn-icon h-8 w-8" :title="t('admin.modelMonitor.moveDown')" :disabled="!group.selected || !canMove(index, 1)" @click="moveGroup(index, 1)"><Icon name="arrowDown" size="sm" /></button>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="groupsOpen = false">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="groupsSaving || selectedGroups.length === 0" @click="saveGroups">{{ groupsSaving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="presentationOpen" :title="t('admin.modelMonitor.presentationTitle', { model: presentationRow?.model || '' })" @close="presentationOpen = false">
      <div class="space-y-5">
        <div>
          <label for="model-monitor-label" class="mb-1.5 block text-sm font-medium text-foreground">{{ t('admin.modelMonitor.label') }}</label>
          <input id="model-monitor-label" v-model="presentationLabel" class="input" maxlength="12" :placeholder="t('admin.modelMonitor.labelPlaceholder')" />
          <div class="mt-2 flex flex-wrap gap-2">
            <button v-for="preset in labelPresets" :key="preset" type="button" class="rounded-control border border-outline px-2.5 py-1 text-xs text-foreground-muted hover:border-brand/30 hover:text-brand" @click="presentationLabel = preset">
              {{ preset }}
            </button>
            <button v-if="presentationLabel" type="button" class="rounded-control px-2.5 py-1 text-xs text-foreground-subtle hover:bg-surface-subtle hover:text-foreground" @click="presentationLabel = ''">
              {{ t('admin.modelMonitor.clearLabel') }}
            </button>
          </div>
        </div>
        <div>
          <label for="model-monitor-display-order" class="mb-1.5 block text-sm font-medium text-foreground">{{ t('admin.modelMonitor.displayOrder') }}</label>
          <input id="model-monitor-display-order" v-model.number="presentationOrder" type="number" min="0" max="9999" step="1" class="input" />
          <p class="mt-1.5 text-xs text-foreground-subtle">{{ t('admin.modelMonitor.displayOrderHint') }}</p>
        </div>
        <div class="border-y border-outline py-3">
          <div class="mb-2 text-xs font-medium text-foreground-subtle">{{ t('admin.modelMonitor.preview') }}</div>
          <div class="flex items-center gap-2">
            <span class="font-medium text-foreground">{{ presentationRow?.model }}</span>
            <span v-if="presentationLabel.trim()" class="inline-flex h-5 max-w-28 items-center truncate rounded-control bg-black px-2 text-[11px] font-bold leading-none text-white">{{ presentationLabel.trim() }}</span>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="presentationOpen = false">{{ t('common.cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="presentationSaving" @click="savePresentation">{{ presentationSaving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="historyOpen" :title="t('admin.modelMonitor.historyTitle', { model: historyRow?.model || '' })" width="wide" @close="historyOpen = false">
      <div v-if="historyLoading" class="py-8 text-center text-sm text-foreground-subtle">{{ t('common.loading') }}</div>
      <div v-else class="max-h-[60vh] overflow-auto">
        <ol class="divide-y divide-outline sm:hidden">
          <li v-for="item in historyItems" :key="`mobile-history-${item.id}`" class="relative py-3 pl-5">
            <span class="absolute left-0 top-4 h-2.5 w-2.5 rounded-full bg-current" :class="statusTextClass(item.status)"></span>
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-xs font-semibold" :class="statusTextClass(item.status)">{{ statusLabel(item.status) }}</p>
                <p class="mt-1 text-xs text-foreground-muted">{{ item.group_name || '—' }}</p>
              </div>
              <time class="text-right text-[10px] leading-4 text-foreground-subtle">{{ formatTime(item.checked_at) }}</time>
            </div>
            <div class="mt-2 flex items-center gap-3 font-mono text-xs tabular-nums text-foreground-muted">
              <span>{{ formatLatency(item.latency_ms) }}</span>
              <span>{{ t('admin.modelMonitor.attempts', { value: item.attempts }) }}</span>
            </div>
          </li>
        </ol>
        <div class="hidden overflow-x-auto sm:block">
        <table class="w-full min-w-[640px] text-left text-sm">
          <thead class="sticky top-0 border-b border-outline bg-surface text-xs text-foreground-subtle">
            <tr><th class="py-2">{{ t('admin.modelMonitor.status') }}</th><th>{{ t('admin.modelMonitor.group') }}</th><th>{{ t('admin.modelMonitor.latency') }}</th><th>{{ t('admin.modelMonitor.actions') }}</th><th>{{ t('admin.modelMonitor.checkedAt') }}</th></tr>
          </thead>
          <tbody>
            <tr v-for="item in historyItems" :key="item.id" class="border-b border-outline">
              <td class="py-2" :class="statusTextClass(item.status)">{{ statusLabel(item.status) }}</td>
              <td>{{ item.group_name || '—' }}</td>
              <td>{{ formatLatency(item.latency_ms) }}</td>
              <td>{{ t('admin.modelMonitor.attempts', { value: item.attempts }) }}</td>
              <td class="text-foreground-subtle">{{ formatTime(item.checked_at) }}</td>
            </tr>
          </tbody>
        </table>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { ModelMonitorGroupOption, ModelMonitorHistoryItem, ModelMonitorRow, ModelMonitorStatus, ModelMonitorTimelinePoint } from '@/api/admin/modelMonitor'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const rows = ref<ModelMonitorRow[]>([])
const loading = ref(false)
const refreshing = ref(false)
const runningKey = ref<string | null>(null)
const search = ref('')
const platform = ref('')
const status = ref('')
const enabled = ref('')
const sortMode = ref<'priority' | 'name' | 'status'>('priority')
const historyOpen = ref(false)
const historyLoading = ref(false)
const historyRow = ref<ModelMonitorRow | null>(null)
const historyItems = ref<ModelMonitorHistoryItem[]>([])
const groupsOpen = ref(false)
const groupsSaving = ref(false)
const groupsRow = ref<ModelMonitorRow | null>(null)
const editableGroups = ref<ModelMonitorGroupOption[]>([])
const presentationOpen = ref(false)
const presentationSaving = ref(false)
const presentationRow = ref<ModelMonitorRow | null>(null)
const presentationLabel = ref('')
const presentationOrder = ref(0)
let autoRefreshTimer: number | null = null
const statuses: ModelMonitorStatus[] = ['operational', 'degraded', 'failed', 'error']
const intervalOptions = [60, 300, 600, 1800, 3600]
const labelPresets = computed(() => [t('admin.modelMonitor.labels.latest'), t('admin.modelMonitor.labels.popular'), t('admin.modelMonitor.labels.recommended')])
const featureEnabled = computed(() => appStore.cachedPublicSettings?.model_marketplace_enabled === true && appStore.cachedPublicSettings?.model_monitor_enabled === true)
const platforms = computed(() => [...new Set(rows.value.map(row => row.platform))].sort())
const activeFilterCount = computed(() => [platform.value, status.value, enabled.value].filter(Boolean).length)
const columns = computed<Column[]>(() => [
  { key: 'model', label: t('admin.modelMonitor.model') },
  { key: 'groups', label: t('admin.modelMonitor.groups') },
  { key: 'presentation', label: t('admin.modelMonitor.presentation') },
  { key: 'status', label: t('admin.modelMonitor.status') },
  { key: 'availability', label: t('admin.modelMonitor.availability') },
  { key: 'latency', label: t('admin.modelMonitor.latency') },
  { key: 'timeline', label: t('admin.modelMonitor.timeline') },
  { key: 'interval', label: t('admin.modelMonitor.interval') },
  { key: 'enabled', label: t('admin.modelMonitor.enabled') },
  { key: 'actions', label: t('admin.modelMonitor.actions') },
])
const filteredRows = computed(() => rows.value.filter(row => {
  if (search.value && !row.model.toLowerCase().includes(search.value.toLowerCase())) return false
  if (platform.value && row.platform !== platform.value) return false
  if (status.value && row.summary?.status !== status.value) return false
  if (enabled.value && String(row.enabled) !== enabled.value) return false
  return true
}).sort((a, b) => {
  if (sortMode.value === 'priority' && a.display_order !== b.display_order) return b.display_order - a.display_order
  if (sortMode.value === 'status') {
    const statusOrder: Record<string, number> = { failed: 0, error: 1, degraded: 2, operational: 3 }
    const comparison = (statusOrder[a.summary?.status || ''] ?? 4) - (statusOrder[b.summary?.status || ''] ?? 4)
    if (comparison !== 0) return comparison
  }
  return a.model.localeCompare(b.model)
}))

async function load(silent = false) {
  if (loading.value || refreshing.value) return
  if (silent) refreshing.value = true
  else loading.value = true
  try {
    rows.value = (await adminAPI.modelMonitor.list()).items || []
  } catch (error) {
    if (!silent) appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.loadError')))
  } finally {
    loading.value = false
    refreshing.value = false
  }
}
async function saveRow(row: ModelMonitorRow, nextEnabled = row.enabled, nextInterval = row.interval_seconds) { try { await adminAPI.modelMonitor.updateConfig({ platform: row.platform, model: row.model, enabled: nextEnabled, interval_seconds: nextInterval, display_order: row.display_order, label: row.label }); row.enabled = nextEnabled; row.interval_seconds = nextInterval; row.configured = true } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.saveError'))); await load() } }
function toggleRow(row: ModelMonitorRow, value: boolean) { void saveRow(row, value) }
function changeInterval(row: ModelMonitorRow, event: Event) { void saveRow(row, row.enabled, Number((event.target as HTMLSelectElement).value)) }
async function runRow(row: ModelMonitorRow) { const key = `${row.platform}:${row.model}`; runningKey.value = key; try { await adminAPI.modelMonitor.runNow(row.platform, row.model); appStore.showSuccess(t('admin.modelMonitor.runSuccess')); await load() } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.runError'))) } finally { runningKey.value = null } }
async function openHistory(row: ModelMonitorRow) { if (!row.id) return; historyRow.value = row; historyOpen.value = true; historyLoading.value = true; try { historyItems.value = (await adminAPI.modelMonitor.history(row.id)).items || [] } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.loadError'))) } finally { historyLoading.value = false } }
const selectedGroups = computed(() => editableGroups.value.filter(group => group.selected))
function openGroups(row: ModelMonitorRow) { groupsRow.value = row; editableGroups.value = row.groups.map(group => ({ ...group })); groupsOpen.value = true }
function groupSummary(row: ModelMonitorRow) { const selected = row.groups.filter(group => group.selected); if (!selected.length) return t('admin.modelMonitor.noGroups'); return selected.length === 1 ? selected[0].name : t('admin.modelMonitor.groupCount', { first: selected[0].name, count: selected.length - 1 }) }
function formatRate(value: number) { return Number(value).toLocaleString(undefined, { maximumFractionDigits: 4 }) }
function selectedPriority(index: number) { return editableGroups.value.slice(0, index + 1).filter(group => group.selected).length }
function adjacentSelectedIndex(index: number, direction: number) { for (let next = index + direction; next >= 0 && next < editableGroups.value.length; next += direction) if (editableGroups.value[next].selected) return next; return -1 }
function canMove(index: number, direction: number) { return adjacentSelectedIndex(index, direction) >= 0 }
function moveGroup(index: number, direction: number) { const target = adjacentSelectedIndex(index, direction); if (target < 0) return; [editableGroups.value[index], editableGroups.value[target]] = [editableGroups.value[target], editableGroups.value[index]] }
async function saveGroups() { if (!groupsRow.value || selectedGroups.value.length === 0) return; groupsSaving.value = true; try { await adminAPI.modelMonitor.updateGroups({ platform: groupsRow.value.platform, model: groupsRow.value.model, group_ids: selectedGroups.value.map(group => group.group_id) }); appStore.showSuccess(t('admin.modelMonitor.groupsSaved')); groupsOpen.value = false; await load() } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.saveError'))) } finally { groupsSaving.value = false } }
function openPresentation(row: ModelMonitorRow) { presentationRow.value = row; presentationLabel.value = row.label; presentationOrder.value = row.display_order; presentationOpen.value = true }
async function savePresentation() {
  const row = presentationRow.value
  if (!row) return
  presentationSaving.value = true
  try {
    const updated = await adminAPI.modelMonitor.updateConfig({
      platform: row.platform,
      model: row.model,
      enabled: row.enabled,
      interval_seconds: row.interval_seconds,
      display_order: Math.max(0, Math.min(9999, Math.trunc(Number(presentationOrder.value) || 0))),
      label: presentationLabel.value.trim(),
    })
    row.display_order = updated.display_order
    row.label = updated.label
    row.configured = true
    presentationOpen.value = false
    await load(true)
    appStore.showSuccess(t('admin.modelMonitor.presentationSaved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.modelMonitor.saveError')))
  } finally {
    presentationSaving.value = false
  }
}
function timeline(row: ModelMonitorRow) { return [...(row.summary?.timeline || [])].slice(0, 20).reverse() }
function timelineClass(value?: string) { return value === 'operational' ? 'bg-success' : value === 'degraded' ? 'bg-warning' : value === 'failed' || value === 'error' ? 'bg-danger' : 'bg-surface-emphasis' }
function statusTextClass(value?: string) { return value === 'operational' ? 'text-success-foreground' : value === 'degraded' ? 'text-warning-foreground' : value === 'failed' || value === 'error' ? 'text-danger-foreground' : 'text-foreground-subtle' }
function statusLabel(value?: string) { return value ? t(`admin.modelMonitor.${value}`) : t('admin.modelMonitor.unknown') }
function formatAvailability(value?: number | null) { return value == null ? '—' : `${value.toFixed(2)}%` }
function formatLatency(value?: number | null) { return value == null ? '—' : `${value} ms` }
function intervalLabel(value: number) { return value >= 60 ? t('admin.modelMonitor.minutes', { value: value / 60 }) : t('admin.modelMonitor.seconds', { value }) }
function timelineTitle(point: ModelMonitorTimelinePoint) { return `${statusLabel(point.status)} · ${formatLatency(point.latency_ms)} · ${formatTime(point.checked_at)}` }
function formatTime(value: string) { return new Date(value).toLocaleString() }

onMounted(() => {
  void load()
  autoRefreshTimer = window.setInterval(() => {
    if (document.visibilityState === 'visible') void load(true)
  }, 30_000)
})
onBeforeUnmount(() => {
  if (autoRefreshTimer != null) window.clearInterval(autoRefreshTimer)
})
</script>
