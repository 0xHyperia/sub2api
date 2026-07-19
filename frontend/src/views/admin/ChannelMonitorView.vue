<template>
  <AppLayout>
    <TablePageLayout class="channel-monitor-page">
      <template #filters>
        <MonitorFiltersBar
          v-model:search="searchQuery"
          v-model:provider="providerFilter"
          v-model:enabled="enabledFilter"
          :loading="loading"
          @reload="reload"
          @create="openCreateDialog"
          @manage-templates="showTemplateManager = true"
          @search-input="handleSearch"
        />
      </template>

      <template #table>
        <DataTable :columns="columns" :data="monitors" :loading="loading">
          <template #mobile-card="{ row }">
            <article class="space-y-3">
              <header class="flex min-w-0 items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex min-w-0 items-center gap-1.5">
                    <h3 class="truncate text-sm font-semibold text-foreground">{{ row.name }}</h3>
                    <Icon v-if="row.api_key_decrypt_failed" name="exclamationTriangle" size="xs" class="shrink-0 text-danger-foreground" :title="t('admin.channelMonitor.apiKeyDecryptFailed')" />
                  </div>
                  <span class="mt-1 inline-flex items-center rounded-control px-2 py-0.5 text-[10px] font-medium" :class="providerBadgeClass(row.provider)">
                    {{ providerLabel(row.provider) }}
                  </span>
                </div>
                <Toggle :modelValue="row.enabled" :aria-label="`${t('admin.channelMonitor.columns.enabled')}: ${row.name}`" @update:modelValue="toggleEnabled(row)" />
              </header>

              <div class="rounded-panel border border-outline bg-surface-subtle px-3 py-2.5">
                <div class="flex min-w-0 items-center justify-between gap-3">
                  <div class="min-w-0">
                    <p class="text-[10px] text-foreground-subtle">{{ t('admin.channelMonitor.columns.primaryModel') }}</p>
                    <p class="mt-0.5 truncate font-mono text-xs font-semibold text-foreground">{{ row.primary_model }}</p>
                  </div>
                  <span class="inline-flex shrink-0 items-center gap-1.5 rounded-control px-2 py-1 text-xs" :class="statusBadgeClass(row.primary_status)">
                    {{ statusLabel(row.primary_status) }}
                  </span>
                </div>
              </div>

              <dl class="grid grid-cols-2 gap-px overflow-hidden rounded-panel border border-outline bg-outline">
                <div class="bg-surface-subtle px-3 py-2.5">
                  <dt class="text-[10px] text-foreground-subtle">{{ t('admin.channelMonitor.columns.availability7d') }}</dt>
                  <dd class="mt-0.5 font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatAvailability(row) }}</dd>
                </div>
                <div class="bg-surface-subtle px-3 py-2.5">
                  <dt class="text-[10px] text-foreground-subtle">{{ t('admin.channelMonitor.columns.latency') }}</dt>
                  <dd class="mt-0.5 font-mono text-sm font-semibold tabular-nums text-foreground">{{ formatLatency(row.primary_latency_ms) }}</dd>
                </div>
              </dl>

              <dl class="grid grid-cols-3 gap-3 text-xs">
                <div class="min-w-0">
                  <dt class="text-[9px] text-foreground-subtle">{{ t('admin.channelMonitor.form.groupName') }}</dt>
                  <dd class="mt-0.5 truncate text-foreground">{{ row.group_name || '—' }}</dd>
                </div>
                <div>
                  <dt class="text-[9px] text-foreground-subtle">{{ t('admin.channelMonitor.form.interval') }}</dt>
                  <dd class="mt-0.5 text-foreground">{{ row.interval_seconds }}s</dd>
                </div>
                <div>
                  <dt class="text-[9px] text-foreground-subtle">{{ t('admin.channelMonitor.form.extraModels') }}</dt>
                  <dd class="mt-0.5 tabular-nums text-foreground">{{ row.extra_models?.length || 0 }}</dd>
                </div>
              </dl>

              <footer class="flex items-center gap-2 border-t border-outline pt-3">
                <button type="button" class="btn btn-secondary min-w-0 flex-1" :disabled="runningId === row.id" @click="handleRunNow(row)">
                  <Icon name="refresh" size="sm" :class="runningId === row.id ? 'animate-spin' : ''" />
                  {{ t('admin.channelMonitor.runNow') }}
                </button>
                <button type="button" class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline text-foreground-muted hover:bg-surface-subtle hover:text-foreground" :title="t('common.edit')" :aria-label="t('common.edit')" @click="openEditDialog(row)">
                  <Icon name="edit" size="sm" />
                </button>
                <details class="group/menu relative" @click.stop>
                  <summary class="inline-flex h-10 w-10 cursor-pointer list-none items-center justify-center rounded-control border border-outline text-foreground-muted hover:bg-surface-subtle hover:text-foreground" :title="t('common.actions')" :aria-label="t('common.actions')">
                    <Icon name="more" size="sm" />
                  </summary>
                  <div class="absolute bottom-full right-0 z-30 mb-1 w-40 overflow-hidden rounded-panel border border-outline bg-surface py-1 shadow-floating">
                    <button type="button" class="flex min-h-10 w-full items-center gap-2 px-3 py-2 text-left text-sm text-foreground-muted hover:bg-surface-subtle hover:text-foreground disabled:opacity-50" :disabled="duplicatingIds.has(row.id) || Boolean(row.api_key_decrypt_failed)" @click="handleDuplicate(row)">
                      <Icon name="copy" size="sm" />{{ duplicatingIds.has(row.id) ? t('admin.channelMonitor.duplicating') : t('admin.channelMonitor.duplicate') }}
                    </button>
                    <button type="button" class="flex min-h-10 w-full items-center gap-2 px-3 py-2 text-left text-sm text-danger-foreground hover:bg-danger-subtle" @click="handleDelete(row)">
                      <Icon name="trash" size="sm" />{{ t('common.delete') }}
                    </button>
                  </div>
                </details>
              </footer>
            </article>
          </template>

          <template #cell-name="{ row, value }">
            <div class="flex items-center gap-1.5">
              <span class="font-medium text-foreground">{{ value }}</span>
              <HelpTooltip v-if="row.api_key_decrypt_failed" :content="t('admin.channelMonitor.apiKeyDecryptFailed')">
                <Icon name="exclamationTriangle" size="sm" class="text-danger-foreground" />
              </HelpTooltip>
            </div>
          </template>

          <template #cell-provider="{ row }">
            <span class="inline-flex items-center rounded-control px-2 py-0.5 text-xs font-medium" :class="providerBadgeClass(row.provider)">
              {{ providerLabel(row.provider) }}
            </span>
          </template>

          <template #cell-primary_model="{ row }">
            <MonitorPrimaryModelCell :row="row" />
          </template>

          <template #cell-availability_7d="{ row }">
            <span class="font-mono text-sm tabular-nums text-foreground">{{ formatAvailability(row) }}</span>
          </template>

          <template #cell-latency="{ row }">
            <span class="font-mono text-sm tabular-nums text-foreground">{{ formatLatency(row.primary_latency_ms) }}</span>
          </template>

          <template #cell-enabled="{ row }">
            <Toggle
              :modelValue="row.enabled"
              :aria-label="`${t('admin.channelMonitor.columns.enabled')}: ${row.name}`"
              @update:modelValue="toggleEnabled(row)"
            />
          </template>

          <template #cell-actions="{ row }">
            <MonitorActionsCell
              :row="row"
              :running="runningId === row.id"
              :duplicating="duplicatingIds.has(row.id)"
              @run="handleRunNow"
              @duplicate="handleDuplicate"
              @edit="openEditDialog"
              @delete="handleDelete"
            />
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.channelMonitor.noMonitorsYet')"
              :description="t('admin.channelMonitor.createFirstMonitor')"
              :action-text="t('admin.channelMonitor.createButton')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="onPageChange"
          @update:pageSize="onPageSizeChange"
        />
      </template>
    </TablePageLayout>

    <MonitorFormDialog
      :show="showDialog"
      :monitor="editing"
      @close="closeDialog"
      @saved="reload"
    />

    <MonitorTemplateManagerDialog
      :show="showTemplateManager"
      @close="showTemplateManager = false"
      @updated="reload"
    />

    <MonitorRunResultDialog
      :show="showRunResult"
      :results="runResults"
      @close="showRunResult = false"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('common.delete')"
      :message="deleteConfirmMessage"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { adminAPI } from '@/api/admin'
import type {
  ChannelMonitor,
  CheckResult,
  ListParams,
  Provider,
} from '@/api/admin/channelMonitor'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import MonitorFiltersBar from '@/components/admin/monitor/MonitorFiltersBar.vue'
import MonitorFormDialog from '@/components/admin/monitor/MonitorFormDialog.vue'
import MonitorTemplateManagerDialog from '@/components/admin/monitor/MonitorTemplateManagerDialog.vue'
import MonitorRunResultDialog from '@/components/admin/monitor/MonitorRunResultDialog.vue'
import MonitorPrimaryModelCell from '@/components/admin/monitor/MonitorPrimaryModelCell.vue'
import MonitorActionsCell from '@/components/admin/monitor/MonitorActionsCell.vue'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const { t } = useI18n()
const appStore = useAppStore()
const {
  providerLabel,
  providerBadgeClass,
  statusLabel,
  statusBadgeClass,
  formatLatency,
  formatAvailability,
} = useChannelMonitorFormat()

const monitors = ref<ChannelMonitor[]>([])
const loading = ref(false)
const runningId = ref<number | null>(null)
const searchQuery = ref('')
const providerFilter = ref<Provider | ''>('')
const enabledFilter = ref<'' | 'true' | 'false'>('')
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0 })

const showDialog = ref(false)
const showTemplateManager = ref(false)
const editing = ref<ChannelMonitor | null>(null)
const showDeleteDialog = ref(false)
const deleting = ref<ChannelMonitor | null>(null)
const showRunResult = ref(false)
const runResults = ref<CheckResult[]>([])
const duplicatingIds = reactive(new Set<number>())

let abortController: AbortController | null = null
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.channelMonitor.columns.name'), sortable: false },
  { key: 'provider', label: t('admin.channelMonitor.columns.provider'), sortable: false },
  { key: 'primary_model', label: t('admin.channelMonitor.columns.primaryModel'), sortable: false },
  { key: 'availability_7d', label: t('admin.channelMonitor.columns.availability7d'), sortable: false },
  { key: 'latency', label: t('admin.channelMonitor.columns.latency'), sortable: false },
  { key: 'enabled', label: t('admin.channelMonitor.columns.enabled'), sortable: false },
  { key: 'actions', label: t('admin.channelMonitor.columns.actions'), sortable: false },
])

const deleteConfirmMessage = computed(() => {
  const name = deleting.value?.name || ''
  return t('admin.channelMonitor.deleteConfirm', { name })
})

async function reload() {
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  loading.value = true
  try {
    const params: ListParams = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (providerFilter.value) params.provider = providerFilter.value
    if (enabledFilter.value === 'true') params.enabled = true
    if (enabledFilter.value === 'false') params.enabled = false
    if (searchQuery.value.trim()) params.search = searchQuery.value.trim()

    const res = await adminAPI.channelMonitor.list(params, { signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    monitors.value = res.items || []
    pagination.total = res.total
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.loadError')))
  } finally {
    if (abortController === ctrl) {
      loading.value = false
      abortController = null
    }
  }
}

function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    reload()
  }, 300)
}

function onPageChange(page: number) {
  pagination.page = page
  reload()
}

function onPageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  reload()
}

function openCreateDialog() {
  editing.value = null
  showDialog.value = true
}

function openEditDialog(row: ChannelMonitor) {
  editing.value = row
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editing.value = null
}

async function toggleEnabled(row: ChannelMonitor) {
  const next = !row.enabled
  try {
    await adminAPI.channelMonitor.update(row.id, { enabled: next })
    row.enabled = next
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

async function handleRunNow(row: ChannelMonitor) {
  if (runningId.value != null) return
  runningId.value = row.id
  try {
    const res = await adminAPI.channelMonitor.runNow(row.id)
    runResults.value = res.results || []
    showRunResult.value = true
    appStore.showSuccess(t('admin.channelMonitor.runSuccess'))
    // Refresh row to get latest status from backend
    void reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.runFailed')))
  } finally {
    runningId.value = null
  }
}

async function handleDuplicate(row: ChannelMonitor) {
  if (row.api_key_decrypt_failed) {
    appStore.showError(t('admin.channelMonitor.duplicateKeyUnavailable'))
    return
  }
  if (duplicatingIds.has(row.id)) return

  duplicatingIds.add(row.id)
  try {
    const duplicate = await adminAPI.channelMonitor.duplicate(row.id)
    appStore.showSuccess(t('admin.channelMonitor.duplicateSuccess', { name: duplicate.name }))
    await reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('admin.channelMonitor.duplicateFailed')))
  } finally {
    duplicatingIds.delete(row.id)
  }
}

function handleDelete(row: ChannelMonitor) {
  deleting.value = row
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deleting.value) return
  try {
    await adminAPI.channelMonitor.del(deleting.value.id)
    appStore.showSuccess(t('admin.channelMonitor.deleteSuccess'))
    showDeleteDialog.value = false
    deleting.value = null
    reload()
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

onMounted(reload)
onUnmounted(() => {
  if (searchTimeout) clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>
