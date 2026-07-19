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
              :placeholder="t('admin.redeem.searchCodes')"
              :aria-label="t('admin.redeem.searchCodes')"
              autocomplete="off"
              class="input pl-10"
              @input="handleSearch"
            />
          </div>
          <div class="w-full sm:w-40">
            <Select
              v-model="filters.type"
              :options="filterTypeOptions"
              @change="loadCodes"
            />
          </div>
          <div class="w-full sm:w-40">
            <Select
              v-model="filters.status"
              :options="filterStatusOptions"
              @change="loadCodes"
            />
          </div>

          <div class="ml-auto flex flex-wrap items-center justify-end gap-2">
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
            <button type="button" @click="handleExportCodes" class="btn btn-secondary">
              <Icon name="download" size="sm" />
              {{ t('admin.redeem.exportCsv') }}
            </button>
            <button
              type="button"
              data-test="batch-update-open"
              @click="openBatchUpdateDialog"
              :disabled="selectedCount === 0 || batchUpdating"
              class="btn btn-secondary"
            >
              <Icon name="edit" size="sm" />
              {{ t('admin.redeem.batchUpdate') }}
            </button>
            <button type="button" @click="showGenerateDialog = true" class="btn btn-primary">
              <Icon name="plus" size="sm" />
              {{ t('admin.redeem.generateCodes') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div data-mobile-layout="redeem-cards" class="space-y-3 md:hidden">
          <div v-if="loading" class="flex items-center justify-center py-12 text-foreground-subtle">
            <Icon name="refresh" size="lg" class="animate-spin" />
          </div>
          <div v-else-if="codes.length === 0" class="rounded-panel border border-dashed border-outline px-4 py-10 text-center text-sm text-foreground-subtle">
            {{ t('admin.redeem.noCodes') }}
          </div>
          <template v-else>
            <label class="flex min-h-10 items-center gap-2 rounded-control border border-outline bg-surface-subtle px-3 text-sm font-medium text-foreground-muted">
              <input
                type="checkbox"
                class="h-5 w-5 cursor-pointer rounded border-outline-strong text-brand focus:ring-focus"
                :checked="allVisibleSelected"
                @change="toggleSelectAllVisible($event)"
              />
              {{ t('common.selectAll') }}
            </label>
            <article
              v-for="row in codes"
              :key="`mobile-${row.id}`"
              class="rounded-panel border border-outline bg-surface p-3 shadow-card"
            >
            <div class="flex items-start gap-3">
              <input
                data-mobile-select-code
                type="checkbox"
                class="mt-0.5 h-5 w-5 shrink-0 cursor-pointer rounded border-outline-strong text-brand focus:ring-focus"
                :checked="selectedCodeIds.has(row.id)"
                :aria-label="row.code"
                @change="toggleSelectRow(row.id, $event)"
              />
              <div class="min-w-0 flex-1">
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
                  <span
                    class="badge shrink-0"
                    :class="row.status === 'unused' ? 'badge-success' : row.status === 'used' ? 'badge-gray' : 'badge-danger'"
                  >
                    {{ t('admin.redeem.status.' + row.status) }}
                  </span>
                </div>

                <div class="mt-3 flex items-end justify-between gap-3 rounded-control bg-surface-subtle p-3">
                  <div>
                    <span
                      class="badge"
                      :class="row.type === 'balance' ? 'badge-success' : row.type === 'subscription' ? 'badge-warning' : 'badge-primary'"
                    >
                      {{ t('admin.redeem.types.' + row.type) }}
                    </span>
                    <div v-if="row.type === 'subscription' && row.group" class="mt-1 text-xs text-foreground-subtle">{{ row.group.name }}</div>
                  </div>
                  <div class="text-right text-lg font-semibold tabular-nums text-foreground">
                    <template v-if="row.type === 'balance'">${{ row.value.toFixed(2) }}</template>
                    <template v-else-if="row.type === 'subscription'">{{ row.validity_days || 30 }} {{ t('admin.redeem.days') }}</template>
                    <template v-else>{{ row.value }}</template>
                  </div>
                </div>

                <dl class="mt-3 space-y-2 text-xs">
                  <div class="flex justify-between gap-3">
                    <dt class="text-foreground-subtle">{{ t('admin.redeem.columns.usedBy') }}</dt>
                    <dd class="min-w-0 break-all text-right text-foreground-muted">{{ row.user?.email || (row.used_by ? t('admin.redeem.userPrefix', { id: row.used_by }) : '—') }}</dd>
                  </div>
                  <div class="flex justify-between gap-3">
                    <dt class="text-foreground-subtle">{{ t('admin.redeem.columns.expiresAt') }}</dt>
                    <dd :class="row.status === 'expired' ? 'text-danger-foreground' : 'text-foreground-muted'">{{ row.expires_at ? formatDateTime(row.expires_at) : t('admin.redeem.neverExpires') }}</dd>
                  </div>
                  <div v-if="row.used_at" class="flex justify-between gap-3">
                    <dt class="text-foreground-subtle">{{ t('admin.redeem.columns.usedAt') }}</dt>
                    <dd class="text-foreground-muted">{{ formatDateTime(row.used_at) }}</dd>
                  </div>
                </dl>

                <details v-if="row.status === 'unused'" class="group relative mt-3 border-t border-outline pt-3">
                  <summary class="flex min-h-10 cursor-pointer list-none items-center justify-center gap-2 rounded-control text-sm font-medium text-foreground-muted hover:bg-surface-subtle hover:text-foreground marker:hidden">
                    <Icon name="more" size="sm" />
                    {{ t('common.more') }}
                  </summary>
                  <div class="dropdown bottom-12 right-0 top-auto w-44">
                    <button type="button" class="dropdown-item w-full text-danger-foreground hover:bg-danger-subtle" @click="handleDelete(row)">
                      <Icon name="trash" size="sm" />
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </details>
              </div>
            </div>
            </article>
          </template>
        </div>

        <div data-desktop-layout="redeem-table" class="hidden md:block">
        <DataTable
          :columns="columns"
          :data="codes"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="id"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #header-select>
            <input
              data-test="select-all-codes"
              type="checkbox"
              class="h-4 w-4 cursor-pointer rounded border-outline-strong text-brand focus:ring-focus"
              :checked="allVisibleSelected"
              :aria-label="t('common.selectAll')"
              @click.stop
              @change="toggleSelectAllVisible($event)"
            />
          </template>

          <template #cell-select="{ row }">
            <input
              data-test="select-code"
              type="checkbox"
              class="h-4 w-4 cursor-pointer rounded border-outline-strong text-brand focus:ring-focus"
              :checked="selectedCodeIds.has(row.id)"
              :aria-label="row.code"
              @click.stop
              @change="toggleSelectRow(row.id, $event)"
            />
          </template>

          <template #cell-code="{ value }">
            <div class="flex items-center space-x-2">
              <code class="font-mono text-sm text-foreground">{{ value }}</code>
              <button
                type="button"
                @click="copyToClipboard(value)"
                :class="[
                  'redeem-copy',
                  copiedCode === value
                    ? 'text-success-foreground'
                    : 'text-foreground-subtle'
                ]"
                :title="copiedCode === value ? t('admin.redeem.copied') : t('keys.copyToClipboard')"
                :aria-label="copiedCode === value ? t('admin.redeem.copied') : t('keys.copyToClipboard')"
              >
                <Icon v-if="copiedCode !== value" name="copy" size="sm" :stroke-width="2" />
                <Icon v-else name="check" size="sm" :stroke-width="2" />
              </button>
            </div>
          </template>

          <template #cell-type="{ value }">
            <span
              :class="[
                'badge',
                value === 'balance'
                  ? 'badge-success'
                  : value === 'subscription'
                    ? 'badge-warning'
                    : 'badge-primary'
              ]"
            >
              {{ t('admin.redeem.types.' + value) }}
            </span>
          </template>

          <template #cell-value="{ value, row }">
            <span class="text-sm font-medium text-foreground">
              <template v-if="row.type === 'balance'">${{ value.toFixed(2) }}</template>
              <template v-else-if="row.type === 'subscription'">
                {{ row.validity_days || 30 }} {{ t('admin.redeem.days') }}
                <span v-if="row.group" class="ml-1 text-xs text-foreground-subtle"
                  >({{ row.group.name }})</span
                >
              </template>
              <template v-else>{{ value }}</template>
            </span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'unused'
                  ? 'badge-success'
                  : value === 'used'
                    ? 'badge-gray'
                    : 'badge-danger'
              ]"
            >
              {{ t('admin.redeem.status.' + value) }}
            </span>
          </template>

          <template #cell-used_by="{ value, row }">
            <span class="text-sm text-foreground-muted">
              {{ row.user?.email || (value ? t('admin.redeem.userPrefix', { id: value }) : '-') }}
            </span>
          </template>

          <template #cell-used_at="{ value }">
            <span class="text-sm text-foreground-muted">{{
              value ? formatDateTime(value) : '-'
            }}</span>
          </template>

          <template #cell-expires_at="{ value, row }">
            <span
              :class="[
                'text-sm',
                row.status === 'expired'
                  ? 'text-danger-foreground'
                  : 'text-foreground-muted'
              ]"
            >
              {{ value ? formatDateTime(value) : t('admin.redeem.neverExpires') }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                v-if="row.status === 'unused'"
                type="button"
                @click="handleDelete(row)"
                class="redeem-action redeem-action-danger"
                :title="t('common.delete')"
                :aria-label="`${t('common.delete')}: ${row.code}`"
              >
                <Icon name="trash" size="sm" />
              </button>
              <span v-else class="text-foreground-subtle">-</span>
            </div>
          </template>
        </DataTable>
        </div>
      </template>

      <template #pagination>
        <div
          v-if="selectedCount > 0"
          class="selection-bar mb-4 flex flex-wrap items-center justify-between gap-3"
        >
          <span class="text-sm font-medium text-foreground">
            {{ t('admin.redeem.selectedCount', { count: selectedCount }) }}
          </span>
          <div class="flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="text-xs font-medium text-foreground-muted hover:text-foreground"
              @click="clearSelectedCodes"
            >
              {{ t('admin.redeem.clearSelection') }}
            </button>
            <button
              type="button"
              class="btn btn-primary btn-sm"
              @click="openBatchUpdateDialog"
            >
              {{ t('admin.redeem.batchUpdate') }}
            </button>
          </div>
        </div>

        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />

        <!-- Batch Actions -->
        <div v-if="filters.status === 'unused'" class="flex justify-end">
          <button type="button" @click="showDeleteUnusedDialog = true" class="btn btn-danger">
            {{ t('admin.redeem.deleteAllUnused') }}
          </button>
        </div>
      </template>
    </TablePageLayout>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.redeem.deleteCode')"
      :message="t('admin.redeem.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Delete Unused Codes Dialog -->
    <ConfirmDialog
      :show="showDeleteUnusedDialog"
      :title="t('admin.redeem.deleteAllUnused')"
      :message="t('admin.redeem.deleteAllUnusedConfirm')"
      :confirm-text="t('admin.redeem.deleteAll')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDeleteUnused"
      @cancel="showDeleteUnusedDialog = false"
    />

    <BaseDialog
      :show="showGenerateDialog"
      :title="t('admin.redeem.generateCodesTitle')"
      width="normal"
      @close="showGenerateDialog = false"
    >
          <form id="generate-redeem-form" @submit.prevent="handleGenerateCodes" class="space-y-4">
            <div>
              <label class="input-label">{{ t('admin.redeem.codeType') }}</label>
              <Select v-model="generateForm.type" :options="typeOptions" />
            </div>
            <!-- 余额/并发类型：显示数值输入 -->
            <div v-if="generateForm.type !== 'subscription' && generateForm.type !== 'invitation'">
              <label for="redeem-code-value" class="input-label">
                {{
                  generateForm.type === 'balance'
                    ? t('admin.redeem.amount')
                    : t('admin.redeem.columns.value')
                }}
              </label>
              <input
                id="redeem-code-value"
                v-model.number="generateForm.value"
                type="number"
                :step="generateForm.type === 'balance' ? '0.01' : '1'"
                :min="generateForm.type === 'balance' ? '0.01' : '1'"
                required
                class="input"
              />
            </div>
            <!-- 邀请码类型：显示提示信息 -->
            <div v-if="generateForm.type === 'invitation'" class="form-note">
              <p class="text-sm text-foreground-muted">
                {{ t('admin.redeem.invitationHint') }}
              </p>
            </div>
            <!-- 订阅类型：显示分组选择和有效天数 -->
            <template v-if="generateForm.type === 'subscription'">
              <div>
                <label class="input-label">{{ t('admin.redeem.selectGroup') }}</label>
                <Select
                  v-model="generateForm.group_id"
                  :options="subscriptionGroupOptions"
                  :placeholder="t('admin.redeem.selectGroupPlaceholder')"
                >
                  <template #selected="{ option }">
                    <GroupBadge
                      v-if="option"
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                    />
                    <span v-else class="text-foreground-subtle">{{
                      t('admin.redeem.selectGroupPlaceholder')
                    }}</span>
                  </template>
                  <template #option="{ option, selected }">
                    <GroupOptionItem
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                      :description="(option as unknown as GroupOption).description"
                      :selected="selected"
                    />
                  </template>
                </Select>
              </div>
              <div>
                <label for="redeem-validity-days" class="input-label">{{ t('admin.redeem.validityDays') }}</label>
                <input
                  id="redeem-validity-days"
                  v-model.number="generateForm.validity_days"
                  type="number"
                  min="1"
                  max="365"
                  required
                  class="input"
                />
              </div>
            </template>
            <div>
              <label class="input-label">{{ t('admin.redeem.codeExpiry') }}</label>
              <div class="grid grid-cols-2 gap-2 sm:grid-cols-5" role="group" :aria-label="t('admin.redeem.codeExpiry')">
                <button
                  v-for="option in redeemCodeExpiryOptions"
                  :key="option.value"
                  type="button"
                  @click="generateForm.expiry_option = option.value"
                  class="expiry-option"
                  :aria-pressed="generateForm.expiry_option === option.value"
                >
                  {{ option.label }}
                </button>
              </div>
              <input
                v-if="generateForm.expiry_option === 'custom'"
                v-model.number="generateForm.custom_expiry_days"
                type="number"
                min="1"
                max="3650"
                required
                class="input mt-2"
                :placeholder="t('admin.redeem.customExpiryDays')"
                :aria-label="t('admin.redeem.customExpiryDays')"
              />
            </div>
            <div>
              <label for="redeem-code-count" class="input-label">{{ t('admin.redeem.count') }}</label>
              <input
                id="redeem-code-count"
                v-model.number="generateForm.count"
                type="number"
                min="1"
                max="100"
                required
                class="input"
              />
            </div>
          </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="showGenerateDialog = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="generate-redeem-form" :disabled="generating" class="btn btn-primary">
            {{ generating ? t('admin.redeem.generating') : t('admin.redeem.generate') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showBatchUpdateDialog"
      :title="t('admin.redeem.batchUpdateTitle')"
      width="normal"
      @close="closeBatchUpdateDialog"
    >
          <p class="mb-4 text-sm text-foreground-subtle">
            {{ t('admin.redeem.selectedCount', { count: selectedCount }) }}
          </p>

          <form id="batch-update-form" data-test="batch-update-form" class="space-y-4" @submit.prevent="handleBatchUpdate">
            <div class="space-y-2">
              <label class="flex items-center gap-2 text-sm font-medium text-foreground-muted">
                <input
                  data-test="batch-field-status"
                  v-model="batchUpdateForm.update_status"
                  type="checkbox"
                  class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
                />
                {{ t('admin.redeem.batchFields.status') }}
              </label>
              <Select
                v-if="batchUpdateForm.update_status"
                v-model="batchUpdateForm.status"
                data-test="batch-status-select"
                :options="batchStatusOptions"
              />
            </div>

            <div class="space-y-2">
              <label class="flex items-center gap-2 text-sm font-medium text-foreground-muted">
                <input
                  v-model="batchUpdateForm.update_expires_at"
                  type="checkbox"
                  class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
                />
                {{ t('admin.redeem.batchFields.expiresAt') }}
              </label>
              <template v-if="batchUpdateForm.update_expires_at">
                <Select v-model="batchUpdateForm.expires_mode" :options="batchExpiryModeOptions" />
                <input
                  v-if="batchUpdateForm.expires_mode === 'custom'"
                  v-model="batchUpdateForm.expires_at_local"
                  type="datetime-local"
                  class="input"
                  :aria-label="t('admin.redeem.batchFields.expiresAt')"
                />
              </template>
            </div>

            <div class="space-y-2">
              <label class="flex items-center gap-2 text-sm font-medium text-foreground-muted">
                <input
                  data-test="batch-field-notes"
                  v-model="batchUpdateForm.update_notes"
                  type="checkbox"
                  class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
                />
                {{ t('admin.redeem.batchFields.notes') }}
              </label>
              <textarea
                v-if="batchUpdateForm.update_notes"
                data-test="batch-notes-input"
                v-model="batchUpdateForm.notes"
                rows="3"
                class="input"
                :placeholder="t('admin.redeem.batchNotesPlaceholder')"
                :aria-label="t('admin.redeem.batchFields.notes')"
              ></textarea>
            </div>

            <div class="space-y-2">
              <label class="flex items-center gap-2 text-sm font-medium text-foreground-muted">
                <input
                  v-model="batchUpdateForm.update_group_id"
                  type="checkbox"
                  class="h-4 w-4 rounded border-outline-strong text-brand focus:ring-focus"
                />
                {{ t('admin.redeem.batchFields.group') }}
              </label>
              <Select
                v-if="batchUpdateForm.update_group_id"
                v-model="batchUpdateForm.group_id"
                :options="batchGroupOptions"
                :placeholder="t('admin.redeem.selectGroupPlaceholder')"
              />
            </div>

          </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="closeBatchUpdateDialog" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            data-test="batch-update-submit"
            type="submit"
            form="batch-update-form"
            :disabled="batchUpdating"
            class="btn btn-primary"
          >
            {{ batchUpdating ? t('common.submitting') : t('admin.redeem.batchUpdate') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showResultDialog"
      :title="t('admin.redeem.generatedSuccessfully')"
      width="normal"
      @close="closeResultDialog"
    >
          <p class="mb-3 text-sm text-foreground-subtle">
            {{ t('admin.redeem.codesCreated', { count: generatedCodes.length }) }}
          </p>
            <div class="relative">
              <textarea
                readonly
                :value="generatedCodesText"
                :style="{ height: textareaHeight }"
                :aria-label="t('admin.redeem.generatedSuccessfully')"
                class="generated-code-list"
              ></textarea>
            </div>
      <template #footer>
          <div class="flex flex-wrap justify-end gap-2">
            <button
              type="button"
              @click="copyGeneratedCodes"
              :class="[
                'btn flex items-center gap-2 transition-all',
                copiedAll ? 'btn-success' : 'btn-secondary'
              ]"
            >
              <Icon v-if="!copiedAll" name="copy" size="sm" :stroke-width="2" />
              <Icon v-else name="check" size="sm" :stroke-width="2" />
              {{ copiedAll ? t('admin.redeem.copied') : t('admin.redeem.copyAll') }}
            </button>
            <button type="button" @click="downloadGeneratedCodes" class="btn btn-primary flex items-center gap-2">
              <Icon name="download" size="sm" :stroke-width="2" />
              {{ t('admin.redeem.download') }}
            </button>
          </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { useTableSelection } from '@/composables/useTableSelection'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type {
  RedeemCode,
  RedeemCodeType,
  Group,
  GroupPlatform,
  SubscriptionType,
  BatchUpdateRedeemCodeFields
} from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

interface GroupOption {
  value: number
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

const showGenerateDialog = ref(false)
const showResultDialog = ref(false)
const generatedCodes = ref<RedeemCode[]>([])
const subscriptionGroups = ref<Group[]>([])

// 订阅类型分组选项
const subscriptionGroupOptions = computed(() => {
  return subscriptionGroups.value
    .filter((g) => g.subscription_type === 'subscription')
    .map((g) => ({
      value: g.id,
      label: g.name,
      description: g.description,
      platform: g.platform,
      subscriptionType: g.subscription_type,
      rate: g.rate_multiplier
    }))
})

const batchGroupOptions = computed(() => [
  { value: null, label: t('admin.redeem.clearGroup') },
  ...subscriptionGroupOptions.value
])

const generatedCodesText = computed(() => {
  return generatedCodes.value.map((code) => code.code).join('\n')
})

const textareaHeight = computed(() => {
  const lineCount = generatedCodes.value.length
  const lineHeight = 24 // approximate line height in px
  const padding = 24 // top + bottom padding
  const minHeight = 60
  const maxHeight = 240
  const calculatedHeight = Math.min(
    Math.max(lineCount * lineHeight + padding, minHeight),
    maxHeight
  )
  return `${calculatedHeight}px`
})

const copiedAll = ref(false)

const closeResultDialog = () => {
  showResultDialog.value = false
  generatedCodes.value = []
  copiedAll.value = false
}

const copyGeneratedCodes = async () => {
  const success = await clipboardCopy(generatedCodesText.value, t('admin.redeem.copied'))
  if (success) {
    copiedAll.value = true
    setTimeout(() => {
      copiedAll.value = false
    }, 2000)
  }
}

const downloadGeneratedCodes = () => {
  const blob = new Blob([generatedCodesText.value], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `redeem-codes-${new Date().toISOString().split('T')[0]}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const columns = computed<Column[]>(() => [
  { key: 'select', label: '' },
  { key: 'code', label: t('admin.redeem.columns.code') },
  { key: 'type', label: t('admin.redeem.columns.type'), sortable: true },
  { key: 'value', label: t('admin.redeem.columns.value'), sortable: true },
  { key: 'status', label: t('admin.redeem.columns.status'), sortable: true },
  { key: 'used_by', label: t('admin.redeem.columns.usedBy') },
  { key: 'used_at', label: t('admin.redeem.columns.usedAt'), sortable: true },
  { key: 'expires_at', label: t('admin.redeem.columns.expiresAt'), sortable: true },
  { key: 'actions', label: t('admin.redeem.columns.actions') }
])

const typeOptions = computed(() => [
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterTypeOptions = computed(() => [
  { value: '', label: t('admin.redeem.allTypes') },
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.redeem.allStatus') },
  { value: 'unused', label: t('admin.redeem.unused') },
  { value: 'used', label: t('admin.redeem.used') },
  { value: 'expired', label: t('admin.redeem.status.expired') },
  { value: 'disabled', label: t('admin.redeem.status.disabled') }
])

const batchStatusOptions = computed(() => [
  { value: 'unused', label: t('admin.redeem.status.unused') },
  { value: 'disabled', label: t('admin.redeem.status.disabled') }
])

const batchExpiryModeOptions = computed(() => [
  { value: 'clear', label: t('admin.redeem.neverExpires') },
  { value: 'custom', label: t('admin.redeem.customExpiry') }
])

const codes = ref<RedeemCode[]>([])
const loading = ref(false)
const generating = ref(false)
const batchUpdating = ref(false)
const searchQuery = ref('')
const filters = reactive({
  type: '',
  status: ''
})
const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = reactive({
  sort_by: 'id',
  sort_order: 'desc' as 'asc' | 'desc'
})

let abortController: AbortController | null = null

const showDeleteDialog = ref(false)
const showDeleteUnusedDialog = ref(false)
const showBatchUpdateDialog = ref(false)
const deletingCode = ref<RedeemCode | null>(null)
const copiedCode = ref<string | null>(null)

const {
  selectedSet: selectedCodeIds,
  selectedCount,
  allVisibleSelected,
  select,
  deselect,
  clear: clearSelectedCodes,
  toggleVisible
} = useTableSelection<RedeemCode>({
  rows: codes,
  getId: (code) => code.id
})

const batchUpdateForm = reactive({
  update_status: false,
  status: 'disabled' as 'unused' | 'disabled',
  update_expires_at: false,
  expires_mode: 'clear' as 'clear' | 'custom',
  expires_at_local: '',
  update_notes: false,
  notes: '',
  update_group_id: false,
  group_id: null as number | null
})

type RedeemCodeExpiryOption = 'never' | '1' | '3' | '7' | 'custom'

const redeemCodeExpiryOptions = computed<{ value: RedeemCodeExpiryOption; label: string }[]>(() => [
  { value: 'never', label: t('admin.redeem.neverExpires') },
  { value: '1', label: t('admin.redeem.expiryPresetDays', { days: 1 }) },
  { value: '3', label: t('admin.redeem.expiryPresetDays', { days: 3 }) },
  { value: '7', label: t('admin.redeem.expiryPresetDays', { days: 7 }) },
  { value: 'custom', label: t('admin.redeem.customExpiry') }
])

const generateForm = reactive({
  type: 'balance' as RedeemCodeType,
  value: 10,
  count: 1,
  group_id: null as number | null,
  validity_days: 30,
  expiry_option: 'never' as RedeemCodeExpiryOption,
  custom_expiry_days: 7
})

// 监听类型变化，邀请码类型时自动设置 value 为 0
watch(
  () => generateForm.type,
  (newType) => {
    if (newType === 'invitation') {
      generateForm.value = 0
    } else if (generateForm.value === 0) {
      generateForm.value = 10
    }
  }
)

const buildRedeemQueryFilters = () => ({
  type: (filters.type || undefined) as RedeemCodeType | undefined,
  status: (filters.status || undefined) as 'used' | 'expired' | 'unused' | 'disabled' | undefined,
  search: searchQuery.value || undefined,
  sort_by: sortState.sort_by,
  sort_order: sortState.sort_order
})

const loadCodes = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true
  try {
    const response = await adminAPI.redeem.list(
      pagination.page,
      pagination.page_size,
      buildRedeemQueryFilters(),
      {
        signal: currentController.signal
      }
    )
    if (currentController.signal.aborted) {
      return
    }
    codes.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (
      currentController.signal.aborted ||
      error?.name === 'AbortError' ||
      error?.code === 'ERR_CANCELED'
    ) {
      return
    }
    appStore.showError(t('admin.redeem.failedToLoad'))
    console.error('Error loading redeem codes:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
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

const toggleSelectRow = (id: number, event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.checked) {
    select(id)
    return
  }
  deselect(id)
}

const toggleSelectAllVisible = (event: Event) => {
  const target = event.target as HTMLInputElement
  toggleVisible(target.checked)
}

const getRedeemCodeExpiresInDays = () => {
  if (generateForm.expiry_option === 'never') {
    return undefined
  }
  if (generateForm.expiry_option === 'custom') {
    if (
      !Number.isFinite(generateForm.custom_expiry_days) ||
      generateForm.custom_expiry_days < 1
    ) {
      return null
    }
    return Math.floor(generateForm.custom_expiry_days)
  }
  return Number(generateForm.expiry_option)
}

const toDatetimeLocalInputValue = (date: Date) => {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}`
}

const resetBatchUpdateForm = () => {
  batchUpdateForm.update_status = false
  batchUpdateForm.status = 'disabled'
  batchUpdateForm.update_expires_at = false
  batchUpdateForm.expires_mode = 'clear'
  batchUpdateForm.expires_at_local = toDatetimeLocalInputValue(
    new Date(Date.now() + 24 * 60 * 60 * 1000)
  )
  batchUpdateForm.update_notes = false
  batchUpdateForm.notes = ''
  batchUpdateForm.update_group_id = false
  batchUpdateForm.group_id = null
}

const openBatchUpdateDialog = () => {
  if (selectedCount.value === 0) {
    appStore.showInfo(t('admin.redeem.selectCodesFirst'))
    return
  }
  resetBatchUpdateForm()
  showBatchUpdateDialog.value = true
}

const closeBatchUpdateDialog = () => {
  showBatchUpdateDialog.value = false
}

const buildBatchUpdateFields = (): BatchUpdateRedeemCodeFields | null => {
  const fields: BatchUpdateRedeemCodeFields = {}

  if (batchUpdateForm.update_status) {
    fields.status = batchUpdateForm.status
  }
  if (batchUpdateForm.update_expires_at) {
    if (batchUpdateForm.expires_mode === 'clear') {
      fields.expires_at = null
    } else {
      const expiresAt = new Date(batchUpdateForm.expires_at_local)
      if (!batchUpdateForm.expires_at_local || Number.isNaN(expiresAt.getTime())) {
        appStore.showError(t('admin.redeem.expiryDaysRequired'))
        return null
      }
      fields.expires_at = expiresAt.toISOString()
    }
  }
  if (batchUpdateForm.update_notes) {
    fields.notes = batchUpdateForm.notes
  }
  if (batchUpdateForm.update_group_id) {
    fields.group_id =
      batchUpdateForm.group_id == null ? null : Number(batchUpdateForm.group_id)
  }

  return Object.keys(fields).length > 0 ? fields : null
}

const handleGenerateCodes = async () => {
  // 订阅类型必须选择分组
  if (generateForm.type === 'subscription' && !generateForm.group_id) {
    appStore.showError(t('admin.redeem.groupRequired'))
    return
  }

  const expiresInDays = getRedeemCodeExpiresInDays()
  if (expiresInDays === null) {
    appStore.showError(t('admin.redeem.expiryDaysRequired'))
    return
  }

  generating.value = true
  try {
    const result = await adminAPI.redeem.generate(
      generateForm.count,
      generateForm.type,
      generateForm.value,
      generateForm.type === 'subscription' ? generateForm.group_id : undefined,
      generateForm.type === 'subscription' ? generateForm.validity_days : undefined,
      expiresInDays
    )
    showGenerateDialog.value = false
    generatedCodes.value = result
    showResultDialog.value = true
    // 重置表单
    generateForm.group_id = null
    generateForm.validity_days = 30
    generateForm.expiry_option = 'never'
    generateForm.custom_expiry_days = 7
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToGenerate'))
    console.error('Error generating codes:', error)
  } finally {
    generating.value = false
  }
}

const copyToClipboard = async (text: string) => {
  const success = await clipboardCopy(text, t('admin.redeem.copied'))
  if (success) {
    copiedCode.value = text
    setTimeout(() => {
      copiedCode.value = null
    }, 2000)
  }
}

const handleExportCodes = async () => {
  try {
    const blob = await adminAPI.redeem.exportCodes(buildRedeemQueryFilters())

    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `redeem-codes-${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    appStore.showSuccess(t('admin.redeem.codesExported'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToExport'))
    console.error('Error exporting codes:', error)
  }
}

const handleDelete = (code: RedeemCode) => {
  deletingCode.value = code
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingCode.value) return

  try {
    await adminAPI.redeem.delete(deletingCode.value.id)
    appStore.showSuccess(t('admin.redeem.codeDeleted'))
    showDeleteDialog.value = false
    deletingCode.value = null
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDelete'))
    console.error('Error deleting code:', error)
  }
}

const confirmDeleteUnused = async () => {
  try {
    // Get all unused codes and delete them
    const unusedCodesResponse = await adminAPI.redeem.list(1, 1000, { status: 'unused' })
    const unusedCodeIds = unusedCodesResponse.items.map((code) => code.id)

    if (unusedCodeIds.length === 0) {
      appStore.showInfo(t('admin.redeem.noUnusedCodes'))
      showDeleteUnusedDialog.value = false
      return
    }

    const result = await adminAPI.redeem.batchDelete(unusedCodeIds)
    appStore.showSuccess(t('admin.redeem.codesDeleted', { count: result.deleted }))
    showDeleteUnusedDialog.value = false
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDeleteUnused'))
    console.error('Error deleting unused codes:', error)
  }
}

const handleBatchUpdate = async () => {
  const ids = Array.from(selectedCodeIds.value)
  if (ids.length === 0) {
    appStore.showInfo(t('admin.redeem.selectCodesFirst'))
    return
  }

  const hasSelectedFields =
    batchUpdateForm.update_status ||
    batchUpdateForm.update_expires_at ||
    batchUpdateForm.update_notes ||
    batchUpdateForm.update_group_id
  if (!hasSelectedFields) {
    appStore.showError(t('admin.redeem.noBatchFieldsSelected'))
    return
  }

  const fields = buildBatchUpdateFields()
  if (!fields) {
    return
  }

  batchUpdating.value = true
  try {
    const result = await adminAPI.redeem.batchUpdate(ids, fields)
    appStore.showSuccess(t('admin.redeem.batchUpdateSuccess', { count: result.updated }))
    showBatchUpdateDialog.value = false
    clearSelectedCodes()
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToBatchUpdate'))
    console.error('Error batch updating codes:', error)
  } finally {
    batchUpdating.value = false
  }
}

// 加载订阅类型分组
const loadSubscriptionGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    subscriptionGroups.value = groups
  } catch (error) {
    console.error('Error loading subscription groups:', error)
  }
}

onMounted(() => {
  loadCodes()
  loadSubscriptionGroups()
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

.redeem-copy,
.redeem-action {
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

.redeem-copy {
  width: 32px;
  height: 32px;
}

.redeem-copy:hover,
.redeem-action:hover {
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.redeem-action-danger:hover {
  color: rgb(var(--color-danger-foreground, 185 28 28));
  background: rgb(var(--color-danger-subtle, 254 242 242));
}

.redeem-copy:focus-visible,
.redeem-action:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: 1px;
}

.selection-bar,
.form-note {
  padding: 12px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 8px;
  background: var(--ui-surface-subtle, #f4f7fb);
}

.expiry-option {
  min-height: 40px;
  padding: 8px 10px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  color: var(--ui-text-muted, #667085);
  background: var(--ui-surface, #fff);
  font-size: 13px;
  transition: color 150ms ease, border-color 150ms ease, background-color 150ms ease;
}

.expiry-option:hover,
.expiry-option[aria-pressed='true'] {
  color: var(--ui-text, #0f172a);
  border-color: var(--ui-border-strong, #c7d3e2);
  background: var(--ui-surface-subtle, #f4f7fb);
}

.expiry-option[aria-pressed='true'] {
  box-shadow: inset 0 -2px 0 var(--ui-text, #0f172a);
}

.expiry-option:focus-visible {
  outline: 2px solid var(--ui-focus, #475569);
  outline-offset: 1px;
}

.generated-code-list {
  width: 100%;
  min-height: 112px;
  max-height: 360px;
  resize: none;
  padding: 12px;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 6px;
  color: var(--ui-text, #0f172a);
  background: var(--ui-surface-subtle, #f4f7fb);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 24px;
  outline: none;
}

@media (max-width: 639px) {
  .commerce-toolbar {
    padding: 10px;
  }
}
</style>
