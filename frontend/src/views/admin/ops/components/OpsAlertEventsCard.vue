<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { opsAPI, type AlertEventsQuery } from '@/api/admin/ops'
import type { AlertEvent } from '../types'
import { formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()
const appStore = useAppStore()

const PAGE_SIZE = 10

const loading = ref(false)
const loadingMore = ref(false)
const loadError = ref('')
const loadMoreError = ref('')
const events = ref<AlertEvent[]>([])
const hasMore = ref(true)
let listGeneration = 0
let detailRequestSequence = 0
let historyRequestSequence = 0

// Detail modal
const showDetail = ref(false)
const selected = ref<AlertEvent | null>(null)
const detailLoading = ref(false)
const detailActionLoading = ref(false)
const historyLoading = ref(false)
const history = ref<AlertEvent[]>([])
const historyRange = ref('7d')
const historyRangeOptions = computed(() => [
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const silenceDuration = ref('1h')
const silenceDurationOptions = computed(() => [
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') }
])

// Filters
const timeRange = ref('24h')
const timeRangeOptions = computed(() => [
  { value: '5m', label: t('admin.ops.timeRange.5m') },
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '6h', label: t('admin.ops.timeRange.6h') },
  { value: '24h', label: t('admin.ops.timeRange.24h') },
  { value: '7d', label: t('admin.ops.timeRange.7d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const severity = ref<string>('')
const severityOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'P0', label: 'P0' },
  { value: 'P1', label: 'P1' },
  { value: 'P2', label: 'P2' },
  { value: 'P3', label: 'P3' }
])

const status = ref<string>('')
const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'firing', label: t('admin.ops.alertEvents.status.firing') },
  { value: 'resolved', label: t('admin.ops.alertEvents.status.resolved') },
  { value: 'manual_resolved', label: t('admin.ops.alertEvents.status.manualResolved') }
])

const emailSent = ref<string>('')
const emailSentOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'true', label: t('admin.ops.alertEvents.table.emailSent') },
  { value: 'false', label: t('admin.ops.alertEvents.table.emailIgnored') }
])

function buildQuery(overrides: Partial<AlertEventsQuery> = {}): AlertEventsQuery {
  const q: AlertEventsQuery = {
    limit: PAGE_SIZE,
    time_range: timeRange.value
  }
  if (severity.value) q.severity = severity.value
  if (status.value) q.status = status.value
  if (emailSent.value === 'true') q.email_sent = true
  if (emailSent.value === 'false') q.email_sent = false
  return { ...q, ...overrides }
}

async function loadFirstPage() {
  const requestId = ++listGeneration
  const query = buildQuery()
  loading.value = true
  loadingMore.value = false
  loadError.value = ''
  loadMoreError.value = ''
  try {
    const data = await opsAPI.listAlertEvents(query)
    if (requestId !== listGeneration) return
    events.value = data
    hasMore.value = data.length === PAGE_SIZE
  } catch (err: any) {
    if (requestId !== listGeneration) return
    console.error('[OpsAlertEventsCard] Failed to load alert events', err)
    loadError.value = err?.response?.data?.detail || t('admin.ops.alertEvents.loadFailed')
    appStore.showError(loadError.value)
  } finally {
    if (requestId === listGeneration) loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || loading.value) return
  if (!hasMore.value) return
  const last = events.value[events.value.length - 1]
  if (!last) return

  const requestId = listGeneration
  const query = buildQuery({ before_fired_at: last.fired_at || last.created_at, before_id: last.id })
  loadingMore.value = true
  loadMoreError.value = ''
  try {
    const data = await opsAPI.listAlertEvents(query)
    if (requestId !== listGeneration) return
    if (!data.length) {
      hasMore.value = false
      return
    }
    events.value = [...events.value, ...data]
    if (data.length < PAGE_SIZE) hasMore.value = false
  } catch (err: any) {
    if (requestId !== listGeneration) return
    console.error('[OpsAlertEventsCard] Failed to load more alert events', err)
    loadMoreError.value = err?.response?.data?.detail || t('admin.ops.alertEvents.loadFailed')
  } finally {
    if (requestId === listGeneration) loadingMore.value = false
  }
}

function retryLoadMore() {
  loadMoreError.value = ''
  void loadMore()
}

function onScroll(e: Event) {
  const el = e.target as HTMLElement | null
  if (!el) return
  const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 120
  if (nearBottom && !loadMoreError.value) void loadMore()
}

function getDimensionString(event: AlertEvent | null | undefined, key: string): string {
  const v = event?.dimensions?.[key]
  if (v == null) return ''
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  return ''
}

function formatDurationMs(ms: number): string {
  const safe = Math.max(0, Math.floor(ms))
  const sec = Math.floor(safe / 1000)
  if (sec < 60) return `${sec}s`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h`
  const day = Math.floor(hr / 24)
  return `${day}d`
}

function formatDurationLabel(event: AlertEvent): string {
  const firedAt = new Date(event.fired_at || event.created_at)
  if (Number.isNaN(firedAt.getTime())) return '-'
  const resolvedAtStr = event.resolved_at || null
  const status = String(event.status || '').trim().toLowerCase()

  if (resolvedAtStr) {
    const resolvedAt = new Date(resolvedAtStr)
    if (!Number.isNaN(resolvedAt.getTime())) {
      const ms = resolvedAt.getTime() - firedAt.getTime()
      const prefix = status === 'manual_resolved'
        ? t('admin.ops.alertEvents.status.manualResolved')
        : t('admin.ops.alertEvents.status.resolved')
      return `${prefix} ${formatDurationMs(ms)}`
    }
  }

  const now = Date.now()
  const ms = now - firedAt.getTime()
  return `${t('admin.ops.alertEvents.status.firing')} ${formatDurationMs(ms)}`
}

function formatDimensionsSummary(event: AlertEvent): string {
  const parts: string[] = []
  const platform = getDimensionString(event, 'platform')
  if (platform) parts.push(`platform=${platform}`)
  const groupId = event.dimensions?.group_id
  if (groupId != null && groupId !== '') parts.push(`group_id=${String(groupId)}`)
  const region = getDimensionString(event, 'region')
  if (region) parts.push(`region=${region}`)
  return parts.length ? parts.join(' ') : '-'
}

function closeDetail() {
  detailRequestSequence += 1
  historyRequestSequence += 1
  showDetail.value = false
  selected.value = null
  history.value = []
}

async function openDetail(row: AlertEvent) {
  const requestId = ++detailRequestSequence
  historyRequestSequence += 1
  showDetail.value = true
  selected.value = row
  history.value = []
  detailLoading.value = true

  try {
    const detail = await opsAPI.getAlertEvent(row.id)
    if (requestId !== detailRequestSequence) return
    selected.value = detail
  } catch (err: any) {
    if (requestId !== detailRequestSequence) return
    console.error('[OpsAlertEventsCard] Failed to load alert detail', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.loadFailed'))
  } finally {
    if (requestId === detailRequestSequence) detailLoading.value = false
  }

  if (requestId === detailRequestSequence) await loadHistory()
}

async function loadHistory() {
  const ev = selected.value
  const requestId = ++historyRequestSequence
  const requestedRange = historyRange.value
  if (!ev) {
    history.value = []
    historyLoading.value = false
    return
  }

  historyLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'number' ? groupIdRaw : undefined

    const items = await opsAPI.listAlertEvents({
      limit: 20,
      time_range: requestedRange,
      platform: platform || undefined,
      group_id: groupId,
      status: ''
    })
    if (requestId !== historyRequestSequence || selected.value?.id !== ev.id) return

    // Best-effort: narrow to same rule_id + dimensions
    history.value = items.filter((it) => {
      if (it.rule_id !== ev.rule_id) return false
      const p1 = getDimensionString(it, 'platform')
      const p2 = getDimensionString(ev, 'platform')
      if ((p1 || '') !== (p2 || '')) return false
      const g1 = it.dimensions?.group_id
      const g2 = ev.dimensions?.group_id
      return (g1 ?? null) === (g2 ?? null)
    })
  } catch (err: any) {
    if (requestId !== historyRequestSequence) return
    console.error('[OpsAlertEventsCard] Failed to load alert history', err)
    history.value = []
  } finally {
    if (requestId === historyRequestSequence) historyLoading.value = false
  }
}

function durationToUntilRFC3339(duration: string): string {
  const now = Date.now()
  if (duration === '1h') return new Date(now + 60 * 60 * 1000).toISOString()
  if (duration === '24h') return new Date(now + 24 * 60 * 60 * 1000).toISOString()
  if (duration === '7d') return new Date(now + 7 * 24 * 60 * 60 * 1000).toISOString()
  return new Date(now + 60 * 60 * 1000).toISOString()
}

async function silenceAlert() {
  const ev = selected.value
  if (!ev) return
  if (detailActionLoading.value) return
  detailActionLoading.value = true
  try {
    const platform = getDimensionString(ev, 'platform')
    const groupIdRaw = ev.dimensions?.group_id
    const groupId = typeof groupIdRaw === 'number' ? groupIdRaw : null
    const region = getDimensionString(ev, 'region') || null

    await opsAPI.createAlertSilence({
      rule_id: ev.rule_id,
      platform: platform || '',
      group_id: groupId ?? undefined,
      region: region ?? undefined,
      until: durationToUntilRFC3339(silenceDuration.value),
      reason: `silence from UI (${silenceDuration.value})`
    })

    appStore.showSuccess(t('admin.ops.alertEvents.detail.silenceSuccess'))
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to silence alert', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.silenceFailed'))
  } finally {
    detailActionLoading.value = false
  }
}

async function manualResolve() {
  if (!selected.value) return
  if (detailActionLoading.value) return
  detailActionLoading.value = true
  try {
    await opsAPI.updateAlertEventStatus(selected.value.id, 'manual_resolved')
    appStore.showSuccess(t('admin.ops.alertEvents.detail.manualResolvedSuccess'))

    // Refresh detail + first page to reflect new status
    const detail = await opsAPI.getAlertEvent(selected.value.id)
    selected.value = detail
    await loadFirstPage()
    await loadHistory()
  } catch (err: any) {
    console.error('[OpsAlertEventsCard] Failed to resolve alert', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.alertEvents.detail.manualResolvedFailed'))
  } finally {
    detailActionLoading.value = false
  }
}

onMounted(() => {
  loadFirstPage()
})

watch([timeRange, severity, status, emailSent], () => {
  events.value = []
  hasMore.value = true
  loadError.value = ''
  loadMoreError.value = ''
  void loadFirstPage()
})

watch(historyRange, () => {
  if (showDetail.value) loadHistory()
})

function severityBadgeClass(severity: string | undefined): string {
  const s = String(severity || '').trim().toLowerCase()
  if (s === 'p0' || s === 'critical') return 'bg-danger-subtle text-danger-foreground'
  if (s === 'p1' || s === 'warning') return 'bg-warning-subtle text-warning-foreground'
  if (s === 'p2' || s === 'info') return 'bg-info-subtle text-info-foreground'
  if (s === 'p3') return 'bg-surface-subtle text-foreground-muted'
  return 'bg-surface-subtle text-foreground-muted'
}

function statusBadgeClass(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (s === 'firing') return 'bg-danger-subtle text-danger-foreground ring-danger/20'
  if (s === 'resolved') return 'ring-success/20 bg-success-subtle text-success-foreground'
  if (s === 'manual_resolved') return 'bg-surface-subtle text-foreground-subtle ring-outline'
  return 'bg-surface-subtle text-foreground-muted ring-outline'
}

function formatStatusLabel(status: string | undefined): string {
  const s = String(status || '').trim().toLowerCase()
  if (!s) return '-'
  if (s === 'firing') return t('admin.ops.alertEvents.status.firing')
  if (s === 'resolved') return t('admin.ops.alertEvents.status.resolved')
  if (s === 'manual_resolved') return t('admin.ops.alertEvents.status.manualResolved')
  return s.toUpperCase()
}

const empty = computed(() => events.value.length === 0 && !loading.value && !loadError.value)
</script>

<template>
  <section class="rounded-panel border border-outline bg-surface p-4 shadow-card sm:p-5" aria-labelledby="ops-alert-events-title">
    <div class="mb-4 flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
      <div>
        <h3 id="ops-alert-events-title" class="text-sm font-semibold text-foreground">{{ t('admin.ops.alertEvents.title') }}</h3>
        <p class="mt-1 text-xs text-foreground-subtle">{{ t('admin.ops.alertEvents.description') }}</p>
      </div>

      <div class="grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:items-center">
        <Select :model-value="timeRange" :options="timeRangeOptions" class="w-full sm:w-[120px]" @change="timeRange = String($event || '24h')" />
        <Select :model-value="severity" :options="severityOptions" class="w-full sm:w-[100px]" @change="severity = String($event || '')" />
        <Select :model-value="status" :options="statusOptions" class="w-full sm:w-[120px]" @change="status = String($event || '')" />
        <Select :model-value="emailSent" :options="emailSentOptions" class="w-full sm:w-[120px]" @change="emailSent = String($event || '')" />
        <button
          type="button"
          class="btn btn-secondary col-span-2 sm:col-span-1"
          :disabled="loading"
          @click="loadFirstPage"
        >
          <svg class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <div
      v-if="loadError && events.length > 0"
      data-testid="alert-events-refresh-error"
      role="alert"
      class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-panel border border-danger/20 bg-danger-subtle px-3 py-2 text-xs text-danger-foreground"
    >
      <span>{{ loadError }}</span>
      <button type="button" class="btn btn-secondary btn-sm" :disabled="loading" @click="loadFirstPage">
        {{ t('common.retry') }}
      </button>
    </div>

    <div v-if="loading && events.length === 0" class="flex items-center gap-2 py-8 text-sm text-foreground-muted">
      <svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      {{ t('admin.ops.alertEvents.loading') }}
    </div>

    <EmptyState
      v-else-if="loadError && events.length === 0"
      data-testid="alert-events-load-error"
      role="alert"
      :title="loadError"
      :action-text="t('common.retry')"
      :action-icon="false"
      @action="loadFirstPage"
    />

    <div v-else-if="empty" class="rounded-panel border border-dashed border-outline p-8 text-center text-sm text-foreground-muted">
      {{ t('admin.ops.alertEvents.empty') }}
    </div>

    <div v-else class="overflow-x-auto rounded-panel border border-outline">
      <div class="max-h-[600px] min-w-[900px] overflow-y-auto" @scroll="onScroll">
        <table class="min-w-full divide-y divide-outline">
          <thead class="sticky top-0 z-10 bg-canvas">
            <tr>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.time') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.severity') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.platform') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.ruleId') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.title') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.duration') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.dimensions') }}
              </th>
              <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">
                {{ t('admin.ops.alertEvents.table.email') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline bg-surface">
            <tr
              v-for="row in events"
              :key="row.id"
              class="cursor-pointer hover:bg-surface-subtle focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus"
              role="button"
              tabindex="0"
              @click="openDetail(row)"
              @keydown.enter.prevent="openDetail(row)"
              @keydown.space.prevent="openDetail(row)"
              :title="row.title || ''"
              :aria-label="row.title || t('admin.ops.alertEvents.detail.title')"
            >
              <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                {{ formatDateTime(row.fired_at || row.created_at) }}
              </td>
              <td class="whitespace-nowrap px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="rounded-full px-2 py-1 text-[10px] font-bold" :class="severityBadgeClass(String(row.severity || ''))">
                    {{ row.severity || '-' }}
                  </span>
                  <span class="inline-flex items-center rounded-full px-2 py-1 text-[10px] font-bold ring-1 ring-inset" :class="statusBadgeClass(row.status)">
                    {{ formatStatusLabel(row.status) }}
                  </span>
                </div>
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                {{ getDimensionString(row, 'platform') || '-' }}
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                <span class="font-mono">#{{ row.rule_id }}</span>
              </td>
              <td class="min-w-[260px] px-4 py-3 text-xs text-foreground-muted">
                <div class="font-semibold truncate max-w-[360px]">{{ row.title || '-' }}</div>
                <div v-if="row.description" class="mt-0.5 line-clamp-2 text-[11px] text-foreground-subtle">
                  {{ row.description }}
                </div>
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-xs text-foreground-muted">
                {{ formatDurationLabel(row) }}
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-[11px] text-foreground-subtle">
                {{ formatDimensionsSummary(row) }}
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-right text-xs">
                <span
                  class="inline-flex items-center justify-end gap-1.5"
                  :title="row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored')"
                >
                  <Icon
                    v-if="row.email_sent"
                    name="checkCircle"
                    size="sm"
                    class="text-success-foreground"
                  />
                  <Icon
                    v-else
                    name="ban"
                    size="sm"
                    class="text-foreground-subtle"
                  />
                  <span class="text-[11px] font-bold text-foreground-muted">
                    {{ row.email_sent ? t('admin.ops.alertEvents.table.emailSent') : t('admin.ops.alertEvents.table.emailIgnored') }}
                  </span>
                </span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="loadingMore" class="flex items-center justify-center gap-2 py-3 text-xs text-foreground-subtle">
          <svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ t('admin.ops.alertEvents.loading') }}
        </div>
        <div
          v-else-if="loadMoreError"
          data-testid="alert-events-load-more-error"
          role="alert"
          class="flex items-center justify-center gap-3 bg-danger-subtle px-3 py-2 text-xs text-danger-foreground"
        >
          <span>{{ loadMoreError }}</span>
          <button type="button" class="btn btn-secondary btn-sm" @click="retryLoadMore">
            {{ t('common.retry') }}
          </button>
        </div>
        <div v-else-if="!hasMore && events.length > 0" class="py-3 text-center text-xs text-foreground-subtle">
          -
        </div>
      </div>
    </div>

    <BaseDialog
      :show="showDetail"
      :title="t('admin.ops.alertEvents.detail.title')"
      width="wide"
      :close-on-click-outside="true"
      @close="closeDetail"
    >
      <div v-if="detailLoading" class="flex items-center justify-center py-10 text-sm text-foreground-subtle">
        {{ t('admin.ops.alertEvents.detail.loading') }}
      </div>

      <div v-else-if="!selected" class="py-10 text-center text-sm text-foreground-subtle">
        {{ t('admin.ops.alertEvents.detail.empty') }}
      </div>

      <div v-else class="space-y-5">
        <section class="border-b border-outline pb-4">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center rounded-full px-2 py-1 text-[10px] font-bold" :class="severityBadgeClass(String(selected.severity || ''))">
                  {{ selected.severity || '-' }}
                </span>
                <span class="inline-flex items-center rounded-full px-2 py-1 text-[10px] font-bold ring-1 ring-inset" :class="statusBadgeClass(selected.status)">
                  {{ formatStatusLabel(selected.status) }}
                </span>
              </div>
              <div class="mt-2 text-sm font-semibold text-foreground">
                {{ selected.title || '-' }}
              </div>
              <div v-if="selected.description" class="mt-1 whitespace-pre-wrap text-xs text-foreground-muted">
                {{ selected.description }}
              </div>
            </div>

            <div class="flex flex-wrap gap-2">
              <div class="flex items-center gap-2 rounded-panel px-2 py-1 ring-1 ring-outline bg-surface">
                <span class="text-[11px] font-bold text-foreground-muted">{{ t('admin.ops.alertEvents.detail.silence') }}</span>
                <Select
                  :model-value="silenceDuration"
                  :options="silenceDurationOptions"
                  class="w-[110px]"
                  @change="silenceDuration = String($event || '1h')"
                />
                <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="silenceAlert">
                  <Icon name="ban" size="sm" />
                  {{ t('common.apply') }}
                </button>
              </div>

              <button type="button" class="btn btn-secondary btn-sm" :disabled="detailActionLoading" @click="manualResolve">
                <Icon name="checkCircle" size="sm" />
                {{ t('admin.ops.alertEvents.detail.manualResolve') }}
              </button>
            </div>
          </div>
        </section>

          <div class="alert-detail-grid grid grid-cols-1 overflow-hidden rounded-panel border border-outline bg-surface-subtle sm:grid-cols-2">
            <div class="alert-detail-field p-4">
              <div class="text-xs font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.detail.firedAt') }}</div>
              <div class="mt-1 text-sm font-medium text-foreground">{{ formatDateTime(selected.fired_at || selected.created_at) }}</div>
            </div>
            <div class="alert-detail-field p-4">
              <div class="text-xs font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.detail.resolvedAt') }}</div>
              <div class="mt-1 text-sm font-medium text-foreground">{{ selected.resolved_at ? formatDateTime(selected.resolved_at) : '-' }}</div>
            </div>
            <div class="alert-detail-field p-4">
              <div class="text-xs font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.detail.ruleId') }}</div>
              <div class="mt-1 flex flex-wrap items-center gap-2">
                <div class="font-mono text-sm font-bold text-foreground">#{{ selected.rule_id }}</div>
                <a
                  class="inline-flex items-center gap-1 rounded-control px-2 py-1 text-[11px] font-bold text-foreground-muted ring-1 ring-outline hover:bg-surface"
                  :href="`/admin/ops?open_alert_rules=1&alert_rule_id=${selected.rule_id}`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewRule') }}
                </a>
                <a
                  class="inline-flex items-center gap-1 rounded-control px-2 py-1 text-[11px] font-bold text-foreground-muted ring-1 ring-outline hover:bg-surface"
                  :href="`/admin/ops?platform=${encodeURIComponent(getDimensionString(selected,'platform')||'')}&group_id=${selected.dimensions?.group_id || ''}&error_type=request&open_error_details=1`"
                >
                  <Icon name="externalLink" size="xs" />
                  {{ t('admin.ops.alertEvents.detail.viewLogs') }}
                </a>
              </div>
            </div>
            <div class="alert-detail-field p-4">
              <div class="text-xs font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.detail.dimensions') }}</div>
              <div class="mt-1 text-sm text-foreground">
                <div v-if="getDimensionString(selected, 'platform')">platform={{ getDimensionString(selected, 'platform') }}</div>
                <div v-if="selected.dimensions?.group_id">group_id={{ selected.dimensions.group_id }}</div>
                <div v-if="getDimensionString(selected, 'region')">region={{ getDimensionString(selected, 'region') }}</div>
              </div>
            </div>
          </div>


        <section class="border-t border-outline pt-4">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div>
              <div class="text-sm font-bold text-foreground">{{ t('admin.ops.alertEvents.detail.historyTitle') }}</div>
              <div class="mt-0.5 text-xs text-foreground-subtle">{{ t('admin.ops.alertEvents.detail.historyHint') }}</div>
            </div>
            <Select :model-value="historyRange" :options="historyRangeOptions" class="w-[140px]" @change="historyRange = String($event || '7d')" />
          </div>

          <div v-if="historyLoading" class="py-6 text-center text-xs text-foreground-subtle">
            {{ t('admin.ops.alertEvents.detail.historyLoading') }}
          </div>
          <div v-else-if="history.length === 0" class="py-6 text-center text-xs text-foreground-subtle">
            {{ t('admin.ops.alertEvents.detail.historyEmpty') }}
          </div>
          <div v-else class="overflow-hidden rounded-panel border border-outline">
            <table class="min-w-full divide-y divide-outline">
              <thead class="bg-canvas">
                <tr>
                  <th class="px-3 py-2 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.table.time') }}</th>
                  <th class="px-3 py-2 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.table.status') }}</th>
                  <th class="px-3 py-2 text-left text-[11px] font-bold uppercase tracking-wider text-foreground-subtle">{{ t('admin.ops.alertEvents.table.metric') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-outline">
                <tr v-for="it in history" :key="it.id" class="hover:bg-surface-subtle/50">
                  <td class="px-3 py-2 text-xs text-foreground-muted">{{ formatDateTime(it.fired_at || it.created_at) }}</td>
                  <td class="px-3 py-2 text-xs">
                    <span class="inline-flex items-center rounded-full px-2 py-1 text-[10px] font-bold ring-1 ring-inset" :class="statusBadgeClass(it.status)">
                      {{ formatStatusLabel(it.status) }}
                    </span>
                  </td>
                  <td class="px-3 py-2 text-xs text-foreground-muted">
                    <span v-if="typeof it.metric_value === 'number' && typeof it.threshold_value === 'number'">
                      {{ it.metric_value.toFixed(2) }} / {{ it.threshold_value.toFixed(2) }}
                    </span>
                    <span v-else>-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </BaseDialog>
  </section>
</template>

<style scoped>
.alert-detail-field {
  min-width: 0;
  border-bottom: 1px solid var(--ui-border, #dbe3ee);
}

.alert-detail-field:last-child {
  border-bottom: 0;
}

@media (min-width: 640px) {
  .alert-detail-field {
    border-right: 1px solid var(--ui-border, #dbe3ee);
  }

  .alert-detail-field:nth-child(2n) {
    border-right: 0;
  }

  .alert-detail-field:nth-last-child(-n + 2) {
    border-bottom: 0;
  }
}
</style>
