<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="commerce-toolbar flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-72">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle" />
            <input
              v-model="searchQuery"
              type="search"
              :placeholder="t('admin.promo.searchCodes')"
              :aria-label="t('admin.promo.searchCodes')"
              autocomplete="off"
              class="input pl-10"
              @input="handleSearch"
            />
          </div>
          <div class="w-full sm:w-40">
            <Select
              v-model="filters.status"
              :options="filterStatusOptions"
              @change="loadCodes"
            />
          </div>

          <div class="ml-auto flex w-full flex-wrap items-center justify-end gap-2 sm:w-auto">
            <button
              type="button"
              @click="loadCodes"
              :disabled="loading"
              class="btn btn-secondary px-2.5"
              :title="t('common.refresh')"
              :aria-label="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button type="button" @click="showCreateDialog = true" class="btn btn-primary min-w-0 flex-1 sm:flex-none">
              <Icon name="plus" size="sm" />
              {{ t('admin.promo.createCode') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div data-mobile-layout="promo-cards" class="space-y-3 md:hidden">
          <div v-if="loading" class="flex items-center justify-center py-12 text-foreground-subtle">
            <Icon name="refresh" size="lg" class="animate-spin" />
          </div>
          <div v-else-if="codes.length === 0" class="rounded-panel border border-dashed border-outline px-4 py-10 text-center text-sm text-foreground-subtle">
            {{ t('admin.promo.noCodesYet') }}
          </div>
          <article
            v-for="row in codes"
            v-else
            :key="`mobile-${row.id}`"
            class="rounded-panel border border-outline bg-surface p-3 shadow-card"
          >
            <div class="flex min-w-0 items-start justify-between gap-2">
              <button
                type="button"
                class="flex min-w-0 items-center gap-2 text-left"
                :title="t('keys.copyToClipboard')"
                @click="copyToClipboard(row.code)"
              >
                <code class="truncate font-mono text-sm font-semibold text-foreground">{{ row.code }}</code>
                <Icon :name="copiedCode === row.code ? 'check' : 'copy'" size="sm" class="shrink-0 text-foreground-subtle" />
              </button>
              <span class="badge shrink-0" :class="getStatusClass(row.status, row)">
                {{ getStatusLabel(row.status, row) }}
              </span>
            </div>

            <div class="mt-3 rounded-control border border-info/20 bg-info-subtle p-3">
              <div class="flex items-end justify-between gap-3">
                <div>
                  <div class="text-xs font-medium text-info-foreground">{{ t('admin.promo.columns.bonusAmount') }}</div>
                  <div class="mt-1 text-xs text-foreground-subtle">{{ localText('适用于新注册用户', 'For new registrations') }}</div>
                </div>
                <div class="text-xl font-semibold tabular-nums text-foreground">+${{ row.bonus_amount.toFixed(2) }}</div>
              </div>
            </div>

            <dl class="mt-3 space-y-2 text-xs">
              <div class="flex justify-between gap-3">
                <dt class="text-foreground-subtle">{{ t('admin.promo.columns.usage') }}</dt>
                <dd class="font-medium tabular-nums text-foreground-muted">{{ row.used_count }} / {{ row.max_uses === 0 ? '∞' : row.max_uses }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-foreground-subtle">{{ t('admin.promo.columns.expiresAt') }}</dt>
                <dd class="text-right text-foreground-muted">{{ row.expires_at ? formatDateTime(row.expires_at) : t('admin.promo.neverExpires') }}</dd>
              </div>
              <div v-if="row.notes" class="flex justify-between gap-3">
                <dt class="shrink-0 text-foreground-subtle">{{ t('admin.promo.notes') }}</dt>
                <dd class="break-words text-right text-foreground-muted">{{ row.notes }}</dd>
              </div>
            </dl>

            <div class="mt-3 flex items-center gap-2 border-t border-outline pt-3">
              <button type="button" data-mobile-action="copy-register-link" class="btn btn-primary btn-sm min-w-0 flex-1" @click="copyRegisterLink(row)">
                <Icon name="link" size="sm" />
                {{ t('admin.promo.copyRegisterLink') }}
              </button>
              <button type="button" data-mobile-action="edit" class="btn btn-secondary btn-sm min-w-0 flex-1" @click="handleEdit(row)">
                <Icon name="edit" size="sm" />
                {{ t('common.edit') }}
              </button>
              <details class="group relative shrink-0">
                <summary class="inline-flex h-10 w-10 cursor-pointer list-none items-center justify-center rounded-control text-foreground-muted hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus marker:hidden" :aria-label="`${t('common.more')} ${row.code}`">
                  <Icon name="more" size="md" />
                </summary>
                <div class="dropdown bottom-12 right-0 top-auto w-48">
                  <button type="button" class="dropdown-item w-full" @click="handleViewUsages(row)">
                    <Icon name="eye" size="sm" />
                    {{ t('admin.promo.viewUsages') }}
                  </button>
                  <button type="button" class="dropdown-item w-full text-danger-foreground hover:bg-danger-subtle" @click="handleDelete(row)">
                    <Icon name="trash" size="sm" />
                    {{ t('common.delete') }}
                  </button>
                </div>
              </details>
            </div>
          </article>
        </div>

        <div data-desktop-layout="promo-table" class="hidden md:block">
        <DataTable
          :columns="columns"
          :data="codes"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-code="{ value }">
            <div class="flex items-center space-x-2">
              <code class="font-mono text-sm text-foreground">{{ value }}</code>
              <button
                type="button"
                @click="copyToClipboard(value)"
                :class="[
                  'promo-copy',
                  copiedCode === value
                    ? 'text-success-foreground'
                    : 'text-foreground-subtle'
                ]"
                :title="copiedCode === value ? t('admin.promo.copied') : t('keys.copyToClipboard')"
                :aria-label="copiedCode === value ? t('admin.promo.copied') : t('keys.copyToClipboard')"
              >
                <Icon v-if="copiedCode !== value" name="copy" size="sm" :stroke-width="2" />
                <Icon v-else name="check" size="sm" :stroke-width="2" />
              </button>
            </div>
          </template>

          <template #cell-bonus_amount="{ value }">
            <span class="text-sm font-medium text-foreground">
              ${{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-usage="{ row }">
            <span class="text-sm text-foreground-muted">
              {{ row.used_count }} / {{ row.max_uses === 0 ? '∞' : row.max_uses }}
            </span>
          </template>

          <template #cell-status="{ value, row }">
            <span
              :class="[
                'badge',
                getStatusClass(value, row)
              ]"
            >
              {{ getStatusLabel(value, row) }}
            </span>
          </template>

          <template #cell-expires_at="{ value }">
            <span class="text-sm text-foreground-muted">
              {{ value ? formatDateTime(value) : t('admin.promo.neverExpires') }}
            </span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-foreground-muted">
              {{ formatDateTime(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                type="button"
                @click="copyRegisterLink(row)"
                class="promo-action"
                :title="t('admin.promo.copyRegisterLink')"
                :aria-label="`${t('admin.promo.copyRegisterLink')}: ${row.code}`"
              >
                <Icon name="link" size="sm" />
              </button>
              <button
                type="button"
                @click="handleViewUsages(row)"
                class="promo-action"
                :title="t('admin.promo.viewUsages')"
                :aria-label="`${t('admin.promo.viewUsages')}: ${row.code}`"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                type="button"
                @click="handleEdit(row)"
                class="promo-action"
                :title="t('common.edit')"
                :aria-label="`${t('common.edit')}: ${row.code}`"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                type="button"
                @click="handleDelete(row)"
                class="promo-action promo-action-danger"
                :title="t('common.delete')"
                :aria-label="`${t('common.delete')}: ${row.code}`"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>
        </DataTable>
        </div>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create Dialog -->
    <BaseDialog
      :show="showCreateDialog"
      :title="t('admin.promo.createCode')"
      width="normal"
      @close="showCreateDialog = false"
    >
      <form id="create-promo-form" @submit.prevent="handleCreate" class="space-y-4">
        <div>
          <label for="create-promo-code" class="input-label">
            {{ t('admin.promo.code') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('admin.promo.autoGenerate') }})</span>
          </label>
          <input
            id="create-promo-code"
            v-model="createForm.code"
            type="text"
            class="input font-mono uppercase"
            :placeholder="t('admin.promo.codePlaceholder')"
          />
        </div>
        <div>
          <label for="create-promo-bonus" class="input-label">{{ t('admin.promo.bonusAmount') }}</label>
          <input
            id="create-promo-bonus"
            v-model.number="createForm.bonus_amount"
            type="number"
            step="0.01"
            min="0"
            required
            class="input"
          />
        </div>
        <div>
          <label for="create-promo-max-uses" class="input-label">
            {{ t('admin.promo.maxUses') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('admin.promo.zeroUnlimited') }})</span>
          </label>
          <input
            id="create-promo-max-uses"
            v-model.number="createForm.max_uses"
            type="number"
            min="0"
            class="input"
          />
        </div>
        <div>
          <label for="create-promo-expires" class="input-label">
            {{ t('admin.promo.expiresAt') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('common.optional') }})</span>
          </label>
          <input
            id="create-promo-expires"
            v-model="createForm.expires_at_str"
            type="datetime-local"
            class="input"
          />
        </div>
        <div>
          <label for="create-promo-notes" class="input-label">
            {{ t('admin.promo.notes') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('common.optional') }})</span>
          </label>
          <textarea
            id="create-promo-notes"
            v-model="createForm.notes"
            rows="2"
            class="input"
            :placeholder="t('admin.promo.notesPlaceholder')"
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showCreateDialog = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="create-promo-form" :disabled="creating" class="btn btn-primary">
            {{ creating ? t('common.creating') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Edit Dialog -->
    <BaseDialog
      :show="showEditDialog"
      :title="t('admin.promo.editCode')"
      width="normal"
      @close="closeEditDialog"
    >
      <form id="edit-promo-form" @submit.prevent="handleUpdate" class="space-y-4">
        <div>
          <label for="edit-promo-code" class="input-label">{{ t('admin.promo.code') }}</label>
          <input
            id="edit-promo-code"
            v-model="editForm.code"
            type="text"
            class="input font-mono uppercase"
          />
        </div>
        <div>
          <label for="edit-promo-bonus" class="input-label">{{ t('admin.promo.bonusAmount') }}</label>
          <input
            id="edit-promo-bonus"
            v-model.number="editForm.bonus_amount"
            type="number"
            step="0.01"
            min="0"
            required
            class="input"
          />
        </div>
        <div>
          <label for="edit-promo-max-uses" class="input-label">
            {{ t('admin.promo.maxUses') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('admin.promo.zeroUnlimited') }})</span>
          </label>
          <input
            id="edit-promo-max-uses"
            v-model.number="editForm.max_uses"
            type="number"
            min="0"
            class="input"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.promo.status') }}</label>
          <Select v-model="editForm.status" :options="statusOptions" />
        </div>
        <div>
          <label for="edit-promo-expires" class="input-label">
            {{ t('admin.promo.expiresAt') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('common.optional') }})</span>
          </label>
          <input
            id="edit-promo-expires"
            v-model="editForm.expires_at_str"
            type="datetime-local"
            class="input"
          />
        </div>
        <div>
          <label for="edit-promo-notes" class="input-label">
            {{ t('admin.promo.notes') }}
            <span class="ml-1 text-xs font-normal text-foreground-subtle">({{ t('common.optional') }})</span>
          </label>
          <textarea
            id="edit-promo-notes"
            v-model="editForm.notes"
            rows="2"
            class="input"
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="closeEditDialog" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="edit-promo-form" :disabled="updating" class="btn btn-primary">
            {{ updating ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Usages Dialog -->
    <BaseDialog
      :show="showUsagesDialog"
      :title="t('admin.promo.usageRecords')"
      width="wide"
      @close="showUsagesDialog = false"
    >
      <div v-if="usagesLoading" class="flex items-center justify-center py-8">
        <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
      </div>
      <div v-else-if="usages.length === 0" class="py-8 text-center text-foreground-subtle">
        {{ t('admin.promo.noUsages') }}
      </div>
      <div v-else class="promo-usage-list">
        <div
          v-for="usage in usages"
          :key="usage.id"
          class="flex items-center justify-between gap-4 px-3 py-3"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div class="promo-avatar">
              <Icon name="user" size="sm" />
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-foreground">
                {{ usage.user?.email || t('admin.promo.userPrefix', { id: usage.user_id }) }}
              </p>
              <p class="text-xs text-foreground-subtle">
                {{ formatDateTime(usage.used_at) }}
              </p>
            </div>
          </div>
          <div class="shrink-0 text-right tabular-nums">
            <span class="text-sm font-semibold text-foreground">
              +${{ usage.bonus_amount.toFixed(2) }}
            </span>
          </div>
        </div>
        <!-- Usages Pagination -->
        <div v-if="usagesTotal > usagesPageSize" class="mt-4">
          <Pagination
            :page="usagesPage"
            :total="usagesTotal"
            :page-size="usagesPageSize"
            @update:page="handleUsagesPageChange"
            @update:page-size="(size: number) => { usagesPageSize = size; usagesPage = 1; loadUsages() }"
          />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <button type="button" @click="showUsagesDialog = false" class="btn btn-secondary">
            {{ t('common.close') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.promo.deleteCode')"
      :message="t('admin.promo.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { PromoCode, PromoCodeUsage } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t, locale } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en

// State
const codes = ref<PromoCode[]>([])
const loading = ref(false)
const creating = ref(false)
const updating = ref(false)
const searchQuery = ref('')
const copiedCode = ref<string | null>(null)

const filters = reactive({
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0
})
const sortState = reactive({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

// Dialogs
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const showUsagesDialog = ref(false)

const editingCode = ref<PromoCode | null>(null)
const deletingCode = ref<PromoCode | null>(null)

// Usages
const usages = ref<PromoCodeUsage[]>([])
const usagesLoading = ref(false)
const currentViewingCode = ref<PromoCode | null>(null)
const usagesPage = ref(1)
const usagesPageSize = ref(20)
const usagesTotal = ref(0)

// Forms
const createForm = reactive({
  code: '',
  bonus_amount: 1,
  max_uses: 0,
  expires_at_str: '',
  notes: ''
})

const editForm = reactive({
  code: '',
  bonus_amount: 0,
  max_uses: 0,
  status: 'active' as 'active' | 'disabled',
  expires_at_str: '',
  notes: ''
})

// Options
const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.promo.allStatus') },
  { value: 'active', label: t('admin.promo.statusActive') },
  { value: 'disabled', label: t('admin.promo.statusDisabled') }
])

const statusOptions = computed(() => [
  { value: 'active', label: t('admin.promo.statusActive') },
  { value: 'disabled', label: t('admin.promo.statusDisabled') }
])

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.promo.columns.code') },
  { key: 'bonus_amount', label: t('admin.promo.columns.bonusAmount'), sortable: true },
  { key: 'usage', label: t('admin.promo.columns.usage') },
  { key: 'status', label: t('admin.promo.columns.status'), sortable: true },
  { key: 'expires_at', label: t('admin.promo.columns.expiresAt'), sortable: true },
  { key: 'created_at', label: t('admin.promo.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.promo.columns.actions') }
])

// Helpers
const getStatusClass = (status: string, row: PromoCode) => {
  if (row.expires_at && new Date(row.expires_at) < new Date()) {
    return 'badge-danger'
  }
  if (row.max_uses > 0 && row.used_count >= row.max_uses) {
    return 'badge-gray'
  }
  return status === 'active' ? 'badge-success' : 'badge-gray'
}

const getStatusLabel = (status: string, row: PromoCode) => {
  if (row.expires_at && new Date(row.expires_at) < new Date()) {
    return t('admin.promo.statusExpired')
  }
  if (row.max_uses > 0 && row.used_count >= row.max_uses) {
    return t('admin.promo.statusMaxUsed')
  }
  return status === 'active' ? t('admin.promo.statusActive') : t('admin.promo.statusDisabled')
}

// API calls
let abortController: AbortController | null = null

const loadCodes = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true

  try {
    const response = await adminAPI.promo.list(
      pagination.page,
      pagination.page_size,
      {
        status: filters.status || undefined,
        search: searchQuery.value || undefined,
        sort_by: sortState.sort_by,
        sort_order: sortState.sort_order
      },
      { signal: currentController.signal }
    )
    if (currentController.signal.aborted || abortController !== currentController) return

    codes.value = response.items
    pagination.total = response.total
  } catch (error: any) {
    if (
      currentController.signal.aborted ||
      abortController !== currentController ||
      error?.name === 'AbortError' ||
      error?.code === 'ERR_CANCELED'
    ) {
      return
    }
    appStore.showError(t('admin.promo.failedToLoad'))
    console.error('Error loading promo codes:', error)
  } finally {
    if (abortController === currentController) {
      loading.value = false
      abortController = null
    }
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadCodes()
  }, 300)
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadCodes()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadCodes()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  loadCodes()
}

const copyToClipboard = async (text: string) => {
  const success = await clipboardCopy(text, t('admin.promo.copied'))
  if (success) {
    copiedCode.value = text
    setTimeout(() => {
      copiedCode.value = null
    }, 2000)
  }
}

// Create
const handleCreate = async () => {
  creating.value = true
  try {
    await adminAPI.promo.create({
      code: createForm.code || undefined,
      bonus_amount: createForm.bonus_amount,
      max_uses: createForm.max_uses,
      expires_at: createForm.expires_at_str ? Math.floor(new Date(createForm.expires_at_str).getTime() / 1000) : undefined,
      notes: createForm.notes || undefined
    })
    appStore.showSuccess(t('admin.promo.codeCreated'))
    showCreateDialog.value = false
    resetCreateForm()
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.promo.failedToCreate'))
  } finally {
    creating.value = false
  }
}

const resetCreateForm = () => {
  createForm.code = ''
  createForm.bonus_amount = 1
  createForm.max_uses = 0
  createForm.expires_at_str = ''
  createForm.notes = ''
}

// Edit
const handleEdit = (code: PromoCode) => {
  editingCode.value = code
  editForm.code = code.code
  editForm.bonus_amount = code.bonus_amount
  editForm.max_uses = code.max_uses
  editForm.status = code.status
  editForm.expires_at_str = code.expires_at ? new Date(code.expires_at).toISOString().slice(0, 16) : ''
  editForm.notes = code.notes || ''
  showEditDialog.value = true
}

const closeEditDialog = () => {
  showEditDialog.value = false
  editingCode.value = null
}

const handleUpdate = async () => {
  if (!editingCode.value) return

  updating.value = true
  try {
    await adminAPI.promo.update(editingCode.value.id, {
      code: editForm.code,
      bonus_amount: editForm.bonus_amount,
      max_uses: editForm.max_uses,
      status: editForm.status,
      expires_at: editForm.expires_at_str ? Math.floor(new Date(editForm.expires_at_str).getTime() / 1000) : 0,
      notes: editForm.notes
    })
    appStore.showSuccess(t('admin.promo.codeUpdated'))
    closeEditDialog()
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.promo.failedToUpdate'))
  } finally {
    updating.value = false
  }
}

// Copy Register Link
const copyRegisterLink = async (code: PromoCode) => {
  const baseUrl = window.location.origin
  const registerLink = `${baseUrl}/register?promo=${encodeURIComponent(code.code)}`

  try {
    await navigator.clipboard.writeText(registerLink)
    appStore.showSuccess(t('admin.promo.registerLinkCopied'))
  } catch (error) {
    // Fallback for older browsers
    const textArea = document.createElement('textarea')
    textArea.value = registerLink
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    appStore.showSuccess(t('admin.promo.registerLinkCopied'))
  }
}

// Delete
const handleDelete = (code: PromoCode) => {
  deletingCode.value = code
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingCode.value) return

  try {
    await adminAPI.promo.delete(deletingCode.value.id)
    appStore.showSuccess(t('admin.promo.codeDeleted'))
    showDeleteDialog.value = false
    deletingCode.value = null
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.promo.failedToDelete'))
  }
}

// View Usages
const handleViewUsages = async (code: PromoCode) => {
  currentViewingCode.value = code
  showUsagesDialog.value = true
  usagesPage.value = 1
  await loadUsages()
}

const loadUsages = async () => {
  if (!currentViewingCode.value) return
  usagesLoading.value = true
  usages.value = []

  try {
    const response = await adminAPI.promo.getUsages(
      currentViewingCode.value.id,
      usagesPage.value,
      usagesPageSize.value
    )
    usages.value = response.items
    usagesTotal.value = response.total
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.promo.failedToLoadUsages'))
  } finally {
    usagesLoading.value = false
  }
}

const handleUsagesPageChange = (page: number) => {
  usagesPage.value = page
  loadUsages()
}

onMounted(() => {
  loadCodes()
})

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>

<style scoped>
.commerce-toolbar {
  position: relative;
  padding: 12px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface, #fff);
  box-shadow: var(--ui-shadow-xs, 0 1px 2px rgba(15, 23, 42, 0.04));
}

.promo-copy,
.promo-action {
  display: inline-flex;
  width: 40px;
  height: 40px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  transition: color 150ms ease, background-color 150ms ease;
}

.promo-copy {
  width: 32px;
  height: 32px;
}

.promo-copy:hover,
.promo-action:hover {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.promo-action-danger:hover {
  color: rgb(var(--color-danger-foreground, 185 28 28));
  background: rgb(var(--color-danger-subtle, 254 242 242));
}

.promo-copy:focus-visible,
.promo-action:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: 1px;
}

.promo-usage-list {
  overflow: hidden;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
}

.promo-usage-list > div + div {
  border-top: 1px solid var(--ui-border, #dbe3ee);
}

.promo-avatar {
  display: flex;
  width: 32px;
  height: 32px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  background: var(--ui-surface-subtle, #f4f7fb);
}

@media (max-width: 639px) {
  .commerce-toolbar {
    padding: 10px;
  }
}
</style>
