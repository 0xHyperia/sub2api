<template>
  <BaseDialog
    :show="show"
    :title="t('admin.scheduledTests.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="space-y-4">
      <div class="flex justify-end">
        <button
          type="button"
          class="btn btn-primary w-full sm:w-auto"
          :aria-expanded="showAddForm"
          aria-controls="scheduled-test-add-form"
          @click="showAddForm = !showAddForm"
        >
          <Icon name="plus" size="sm" :stroke-width="2" />
          {{ t('admin.scheduledTests.addPlan') }}
        </button>
      </div>

      <!-- Add Plan Form -->
      <div
        v-if="showAddForm"
        id="scheduled-test-add-form"
        class="rounded-panel border border-outline bg-surface-subtle p-3 sm:p-4"
      >
        <div class="mb-3 text-sm font-medium text-foreground">
          {{ t('admin.scheduledTests.addPlan') }}
        </div>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label for="scheduled-test-new-model" class="mb-1 block text-xs font-medium text-foreground-muted">
              {{ t('admin.scheduledTests.model') }}
            </label>
            <Select
              id="scheduled-test-new-model"
              v-model="newPlan.model_id"
              :options="modelOptions"
              :placeholder="t('admin.scheduledTests.model')"
              :searchable="modelOptions.length > 5"
            />
          </div>
          <div>
            <div class="mb-1 flex items-center gap-1 text-xs font-medium text-foreground-muted">
              <label for="scheduled-test-new-cron">
                {{ t('admin.scheduledTests.cronExpression') }}
              </label>
              <HelpTooltip trigger="click">
                <template #trigger>
                  <button
                    type="button"
                    class="inline-flex h-6 w-6 items-center justify-center rounded-control text-foreground-subtle hover:bg-surface hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                    :aria-label="t('admin.scheduledTests.cronTooltipTitle')"
                  >
                    <Icon name="infoCircle" size="xs" :stroke-width="2" />
                  </button>
                </template>
                <div class="space-y-1.5">
                  <p class="font-medium">{{ t('admin.scheduledTests.cronTooltipTitle') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipMeaning') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleEvery30Min') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleHourly') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleDaily') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleWeekly') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipRange') }}</p>
                </div>
              </HelpTooltip>
            </div>
            <Input
              id="scheduled-test-new-cron"
              v-model="newPlan.cron_expression"
              :placeholder="'*/30 * * * *'"
              :hint="t('admin.scheduledTests.cronHelp')"
            />
          </div>
          <div>
            <div class="mb-1 flex items-center gap-1 text-xs font-medium text-foreground-muted">
              <label for="scheduled-test-new-max-results">
                {{ t('admin.scheduledTests.maxResults') }}
              </label>
              <HelpTooltip trigger="click">
                <template #trigger>
                  <button
                    type="button"
                    class="inline-flex h-6 w-6 items-center justify-center rounded-control text-foreground-subtle hover:bg-surface hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                    :aria-label="t('admin.scheduledTests.maxResultsTooltipTitle')"
                  >
                    <Icon name="infoCircle" size="xs" :stroke-width="2" />
                  </button>
                </template>
                <div class="space-y-1.5">
                  <p class="font-medium">{{ t('admin.scheduledTests.maxResultsTooltipTitle') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipMeaning') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipBody') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipExample') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipRange') }}</p>
                </div>
              </HelpTooltip>
            </div>
            <Input
              id="scheduled-test-new-max-results"
              v-model="newPlan.max_results"
              type="number"
              placeholder="100"
            />
          </div>
          <div class="flex items-end">
            <label class="flex items-center gap-2 text-sm text-foreground-muted">
              <Toggle
                v-model="newPlan.enabled"
                :aria-label="t('admin.scheduledTests.enabled')"
              />
              {{ t('admin.scheduledTests.enabled') }}
            </label>
          </div>
          <div class="flex items-end">
            <div>
              <label class="flex items-center gap-2 text-sm text-foreground-muted">
                <Toggle
                  v-model="newPlan.auto_recover"
                  :aria-label="t('admin.scheduledTests.autoRecover')"
                />
                {{ t('admin.scheduledTests.autoRecover') }}
              </label>
              <p class="mt-1 text-xs text-foreground-subtle">
                {{ t('admin.scheduledTests.autoRecoverHelp') }}
              </p>
            </div>
          </div>
        </div>
        <div class="mt-3 grid grid-cols-1 gap-2 sm:flex sm:justify-end">
          <button
            type="button"
            @click="showAddForm = false; resetNewPlan()"
            class="btn btn-secondary btn-sm w-full sm:w-auto"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            @click="handleCreate"
            :disabled="!newPlan.model_id || !newPlan.cron_expression || creating"
            class="btn btn-primary btn-sm w-full sm:w-auto"
          >
            <Icon v-if="creating" name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            {{ t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" role="status" class="flex items-center justify-center py-8">
        <Icon name="refresh" size="md" class="animate-spin text-foreground-subtle" :stroke-width="2" />
        <span class="ml-2 text-sm text-foreground-muted">{{ t('common.loading') }}...</span>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="plans.length === 0"
        class="rounded-panel border border-dashed border-outline-strong bg-surface-subtle px-3 py-10 text-center"
      >
        <Icon name="calendar" size="lg" class="mx-auto mb-2 text-foreground-subtle" :stroke-width="1.5" />
        <p class="text-sm text-foreground-muted">
          {{ t('admin.scheduledTests.noPlans') }}
        </p>
      </div>

      <!-- Plans List -->
      <div v-else class="space-y-3">
        <div
          v-for="plan in plans"
          :key="plan.id"
          class="min-w-0 overflow-hidden rounded-panel border border-outline bg-surface"
        >
          <div class="flex flex-col gap-3 px-3 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-4">
            <button
              type="button"
              class="flex min-w-0 flex-1 items-center justify-between gap-3 rounded-control text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
              :aria-expanded="expandedPlanId === plan.id"
              :aria-controls="`scheduled-test-results-${plan.id}`"
              @click="toggleExpand(plan.id)"
            >
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-foreground" :title="plan.model_id">
                  {{ plan.model_id }}
                </div>
                <div class="mt-0.5 truncate font-mono text-xs text-foreground-muted" :title="plan.cron_expression">
                  {{ plan.cron_expression }}
                </div>
              </div>
              <span v-if="plan.auto_recover" class="badge badge-success hidden shrink-0 sm:inline-flex">
                {{ t('admin.scheduledTests.autoRecover') }}
              </span>
              <Icon
                name="chevronDown"
                size="sm"
                :class="[
                  'shrink-0 text-foreground-subtle transition-transform duration-200',
                  expandedPlanId === plan.id ? 'rotate-180' : ''
                ]"
              />
            </button>

            <div class="flex flex-wrap items-center justify-between gap-2 sm:justify-end sm:gap-3">
              <div class="flex items-center gap-1.5">
                <Toggle
                  :model-value="plan.enabled"
                  :aria-label="`${plan.model_id}: ${t('admin.scheduledTests.enabled')}`"
                  @update:model-value="(val: boolean) => handleToggleEnabled(plan, val)"
                />
                <span class="text-xs text-foreground-muted">
                  {{ plan.enabled ? t('admin.scheduledTests.enabled') : '' }}
                </span>
              </div>

              <span v-if="plan.auto_recover" class="badge badge-success sm:hidden">
                {{ t('admin.scheduledTests.autoRecover') }}
              </span>

              <div v-if="plan.last_run_at" class="hidden text-right text-xs text-foreground-muted lg:block">
                <div>{{ t('admin.scheduledTests.lastRun') }}</div>
                <div>{{ formatDateTime(plan.last_run_at) }}</div>
              </div>

              <div v-if="plan.next_run_at" class="hidden text-right text-xs text-foreground-muted lg:block">
                <div>{{ t('admin.scheduledTests.nextRun') }}</div>
                <div>{{ formatDateTime(plan.next_run_at) }}</div>
              </div>

              <div class="ml-auto flex items-center gap-1 sm:ml-0">
                <button
                  type="button"
                  class="btn btn-ghost btn-sm w-control-sm px-0"
                  :aria-label="t('admin.scheduledTests.editPlan')"
                  :title="t('admin.scheduledTests.editPlan')"
                  @click="startEdit(plan)"
                >
                  <Icon name="edit" size="sm" :stroke-width="2" />
                </button>
                <button
                  type="button"
                  class="btn btn-ghost btn-sm w-control-sm px-0 text-danger-foreground"
                  :aria-label="t('admin.scheduledTests.deletePlan')"
                  :title="t('admin.scheduledTests.deletePlan')"
                  @click="confirmDeletePlan(plan)"
                >
                  <Icon name="trash" size="sm" :stroke-width="2" />
                </button>
              </div>
            </div>
          </div>

          <!-- Edit Form -->
          <div
            v-if="editingPlanId === plan.id"
            class="border-t border-outline bg-surface-subtle px-3 py-3 sm:px-4"
          >
            <div class="mb-2 text-xs font-medium text-foreground-muted">
              {{ t('admin.scheduledTests.editPlan') }}
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <label for="scheduled-test-edit-model" class="mb-1 block text-xs font-medium text-foreground-muted">
                  {{ t('admin.scheduledTests.model') }}
                </label>
                <Select
                  id="scheduled-test-edit-model"
                  v-model="editForm.model_id"
                  :options="modelOptions"
                  :placeholder="t('admin.scheduledTests.model')"
                  :searchable="modelOptions.length > 5"
                />
              </div>
              <div>
                <div class="mb-1 flex items-center gap-1 text-xs font-medium text-foreground-muted">
                  <label for="scheduled-test-edit-cron">
                    {{ t('admin.scheduledTests.cronExpression') }}
                  </label>
                  <HelpTooltip trigger="click">
                    <template #trigger>
                      <button
                        type="button"
                        class="inline-flex h-6 w-6 items-center justify-center rounded-control text-foreground-subtle hover:bg-surface hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                        :aria-label="t('admin.scheduledTests.cronTooltipTitle')"
                      >
                        <Icon name="infoCircle" size="xs" :stroke-width="2" />
                      </button>
                    </template>
                    <div class="space-y-1.5">
                      <p class="font-medium">{{ t('admin.scheduledTests.cronTooltipTitle') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipMeaning') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleEvery30Min') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleHourly') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleDaily') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleWeekly') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipRange') }}</p>
                    </div>
                  </HelpTooltip>
                </div>
                <Input
                  id="scheduled-test-edit-cron"
                  v-model="editForm.cron_expression"
                  :placeholder="'*/30 * * * *'"
                  :hint="t('admin.scheduledTests.cronHelp')"
                />
              </div>
              <div>
                <div class="mb-1 flex items-center gap-1 text-xs font-medium text-foreground-muted">
                  <label for="scheduled-test-edit-max-results">
                    {{ t('admin.scheduledTests.maxResults') }}
                  </label>
                  <HelpTooltip trigger="click">
                    <template #trigger>
                      <button
                        type="button"
                        class="inline-flex h-6 w-6 items-center justify-center rounded-control text-foreground-subtle hover:bg-surface hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                        :aria-label="t('admin.scheduledTests.maxResultsTooltipTitle')"
                      >
                        <Icon name="infoCircle" size="xs" :stroke-width="2" />
                      </button>
                    </template>
                    <div class="space-y-1.5">
                      <p class="font-medium">{{ t('admin.scheduledTests.maxResultsTooltipTitle') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipMeaning') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipBody') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipExample') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipRange') }}</p>
                    </div>
                  </HelpTooltip>
                </div>
                <Input
                  id="scheduled-test-edit-max-results"
                  v-model="editForm.max_results"
                  type="number"
                  placeholder="100"
                />
              </div>
              <div class="flex items-end">
                <label class="flex items-center gap-2 text-sm text-foreground-muted">
                  <Toggle
                    v-model="editForm.enabled"
                    :aria-label="t('admin.scheduledTests.enabled')"
                  />
                  {{ t('admin.scheduledTests.enabled') }}
                </label>
              </div>
              <div class="flex items-end">
                <div>
                  <label class="flex items-center gap-2 text-sm text-foreground-muted">
                    <Toggle
                      v-model="editForm.auto_recover"
                      :aria-label="t('admin.scheduledTests.autoRecover')"
                    />
                    {{ t('admin.scheduledTests.autoRecover') }}
                  </label>
                  <p class="mt-1 text-xs text-foreground-subtle">
                    {{ t('admin.scheduledTests.autoRecoverHelp') }}
                  </p>
                </div>
              </div>
            </div>
            <div class="mt-3 grid grid-cols-1 gap-2 sm:flex sm:justify-end">
              <button
                type="button"
                @click="cancelEdit"
                class="btn btn-secondary btn-sm w-full sm:w-auto"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                @click="handleEdit"
                :disabled="!editForm.model_id || !editForm.cron_expression || updating"
                class="btn btn-primary btn-sm w-full sm:w-auto"
              >
                <Icon v-if="updating" name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
                {{ t('common.save') }}
              </button>
            </div>
          </div>

          <!-- Expanded Results Section -->
          <div
            v-if="expandedPlanId === plan.id"
            :id="`scheduled-test-results-${plan.id}`"
            class="border-t border-outline px-3 py-3 sm:px-4"
          >
            <div class="mb-2 text-xs font-medium text-foreground-muted">
              {{ t('admin.scheduledTests.results') }}
            </div>

            <div v-if="loadingResults" role="status" class="flex items-center justify-center py-4">
              <Icon name="refresh" size="sm" class="animate-spin text-foreground-subtle" :stroke-width="2" />
              <span class="ml-2 text-xs text-foreground-muted">{{ t('common.loading') }}...</span>
            </div>

            <div
              v-else-if="results.length === 0"
              class="py-4 text-center text-xs text-foreground-muted"
            >
              {{ t('admin.scheduledTests.noResults') }}
            </div>

            <!-- Results List -->
            <div v-else class="max-h-64 space-y-2 overflow-y-auto">
              <div
                v-for="result in results"
                :key="result.id"
                class="min-w-0 rounded-panel border border-outline bg-surface-subtle p-3"
              >
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <div class="flex flex-wrap items-center gap-2">
                    <span :class="['badge', resultStatusClass(result.status)]">
                      {{ resultStatusLabel(result.status) }}
                    </span>

                    <span v-if="result.latency_ms > 0" class="text-xs text-foreground-muted">
                      {{ result.latency_ms }}ms
                    </span>
                  </div>

                  <span class="break-words text-xs text-foreground-subtle">
                    {{ formatDateTime(result.started_at) }}
                  </span>
                </div>

                <div v-if="result.error_message" class="mt-2">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-2 rounded-control py-1 text-left text-xs font-medium text-danger-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                    :aria-expanded="expandedResultIds.has(result.id)"
                    :aria-controls="`scheduled-result-detail-${result.id}`"
                    @click="toggleResultDetail(result.id)"
                  >
                    <span>{{ t('admin.scheduledTests.errorMessage') }}</span>
                    <Icon
                      name="chevronDown"
                      size="sm"
                      :class="[
                        'shrink-0 transition-transform duration-200',
                        expandedResultIds.has(result.id) ? 'rotate-180' : ''
                      ]"
                    />
                  </button>
                  <pre
                    v-if="expandedResultIds.has(result.id)"
                    :id="`scheduled-result-detail-${result.id}`"
                    class="mt-1 max-h-32 overflow-auto whitespace-pre-wrap break-all rounded-control border border-danger/20 bg-danger-subtle p-2 text-xs text-danger-foreground"
                  >{{ result.error_message }}</pre>
                </div>
                <div v-else-if="result.response_text" class="mt-2">
                  <button
                    type="button"
                    class="flex w-full items-center justify-between gap-2 rounded-control py-1 text-left text-xs font-medium text-foreground-muted focus:outline-none focus-visible:ring-2 focus-visible:ring-focus/30"
                    :aria-expanded="expandedResultIds.has(result.id)"
                    :aria-controls="`scheduled-result-detail-${result.id}`"
                    @click="toggleResultDetail(result.id)"
                  >
                    <span>{{ t('admin.scheduledTests.responseText') }}</span>
                    <Icon
                      name="chevronDown"
                      size="sm"
                      :class="[
                        'shrink-0 transition-transform duration-200',
                        expandedResultIds.has(result.id) ? 'rotate-180' : ''
                      ]"
                    />
                  </button>
                  <pre
                    v-if="expandedResultIds.has(result.id)"
                    :id="`scheduled-result-detail-${result.id}`"
                    class="mt-1 max-h-32 overflow-auto whitespace-pre-wrap break-all rounded-control border border-outline bg-surface p-2 text-xs text-foreground-muted"
                  >{{ result.response_text }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('admin.scheduledTests.deletePlan')"
      :message="t('admin.scheduledTests.confirmDelete')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Input from '@/components/common/Input.vue'
import Toggle from '@/components/common/Toggle.vue'
import { Icon } from '@/components/icons'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { ScheduledTestPlan, ScheduledTestResult } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  show: boolean
  accountId: number | null
  modelOptions: SelectOption[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

// State
const loading = ref(false)
const creating = ref(false)
const loadingResults = ref(false)
const plans = ref<ScheduledTestPlan[]>([])
const results = ref<ScheduledTestResult[]>([])
const expandedPlanId = ref<number | null>(null)
const expandedResultIds = reactive(new Set<number>())
const showAddForm = ref(false)
const showDeleteConfirm = ref(false)
const deletingPlan = ref<ScheduledTestPlan | null>(null)
const editingPlanId = ref<number | null>(null)
const updating = ref(false)
const editForm = reactive({
  model_id: '' as string,
  cron_expression: '' as string,
  max_results: '100' as string,
  enabled: true,
  auto_recover: false
})

const newPlan = reactive({
  model_id: '' as string,
  cron_expression: '' as string,
  max_results: '100' as string,
  enabled: true,
  auto_recover: false
})

const resetNewPlan = () => {
  newPlan.model_id = ''
  newPlan.cron_expression = ''
  newPlan.max_results = '100'
  newPlan.enabled = true
  newPlan.auto_recover = false
}

const resultStatusClass = (status: string) => {
  if (status === 'success') return 'badge-success'
  if (status === 'running') return 'badge-primary'
  return 'badge-danger'
}

const resultStatusLabel = (status: string) => {
  if (status === 'success') return t('admin.scheduledTests.success')
  if (status === 'running') return t('admin.scheduledTests.running')
  return t('admin.scheduledTests.failed')
}

// Load plans when dialog opens
watch(
  () => props.show,
  async (visible) => {
    if (visible && props.accountId) {
      await loadPlans()
    } else {
      plans.value = []
      results.value = []
      expandedPlanId.value = null
      expandedResultIds.clear()
      showAddForm.value = false
      showDeleteConfirm.value = false
    }
  },
  { immediate: true }
)

const loadPlans = async () => {
  if (!props.accountId) return
  loading.value = true
  try {
    plans.value = await adminAPI.scheduledTests.listByAccount(props.accountId)
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to load plans')
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  if (!props.accountId || !newPlan.model_id || !newPlan.cron_expression) return
  creating.value = true
  try {
    const maxResults = Number(newPlan.max_results) || 100
    await adminAPI.scheduledTests.create({
      account_id: props.accountId,
      model_id: newPlan.model_id,
      cron_expression: newPlan.cron_expression,
      enabled: newPlan.enabled,
      max_results: maxResults,
      auto_recover: newPlan.auto_recover
    })
    appStore.showSuccess(t('admin.scheduledTests.createSuccess'))
    showAddForm.value = false
    resetNewPlan()
    await loadPlans()
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to create plan')
  } finally {
    creating.value = false
  }
}

const handleToggleEnabled = async (plan: ScheduledTestPlan, enabled: boolean) => {
  try {
    const updated = await adminAPI.scheduledTests.update(plan.id, { enabled })
    const index = plans.value.findIndex((p) => p.id === plan.id)
    if (index !== -1) {
      plans.value[index] = updated
    }
    appStore.showSuccess(t('admin.scheduledTests.updateSuccess'))
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to update plan')
  }
}

const startEdit = (plan: ScheduledTestPlan) => {
  editingPlanId.value = plan.id
  editForm.model_id = plan.model_id
  editForm.cron_expression = plan.cron_expression
  editForm.max_results = String(plan.max_results)
  editForm.enabled = plan.enabled
  editForm.auto_recover = plan.auto_recover
}

const cancelEdit = () => {
  editingPlanId.value = null
}

const handleEdit = async () => {
  if (!editingPlanId.value || !editForm.model_id || !editForm.cron_expression) return
  updating.value = true
  try {
    const updated = await adminAPI.scheduledTests.update(editingPlanId.value, {
      model_id: editForm.model_id,
      cron_expression: editForm.cron_expression,
      max_results: Number(editForm.max_results) || 100,
      enabled: editForm.enabled,
      auto_recover: editForm.auto_recover
    })
    const index = plans.value.findIndex((p) => p.id === editingPlanId.value)
    if (index !== -1) {
      plans.value[index] = updated
    }
    appStore.showSuccess(t('admin.scheduledTests.updateSuccess'))
    editingPlanId.value = null
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to update plan')
  } finally {
    updating.value = false
  }
}

const confirmDeletePlan = (plan: ScheduledTestPlan) => {
  deletingPlan.value = plan
  showDeleteConfirm.value = true
}

const handleDelete = async () => {
  if (!deletingPlan.value) return
  try {
    await adminAPI.scheduledTests.delete(deletingPlan.value.id)
    appStore.showSuccess(t('admin.scheduledTests.deleteSuccess'))
    plans.value = plans.value.filter((p) => p.id !== deletingPlan.value!.id)
    if (expandedPlanId.value === deletingPlan.value.id) {
      expandedPlanId.value = null
      results.value = []
    }
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to delete plan')
  } finally {
    showDeleteConfirm.value = false
    deletingPlan.value = null
  }
}

const toggleExpand = async (planId: number) => {
  if (expandedPlanId.value === planId) {
    expandedPlanId.value = null
    results.value = []
    expandedResultIds.clear()
    return
  }

  expandedPlanId.value = planId
  expandedResultIds.clear()
  loadingResults.value = true
  try {
    results.value = await adminAPI.scheduledTests.listResults(planId, 20)
  } catch (error: any) {
    appStore.showError(error?.message || 'Failed to load results')
    results.value = []
  } finally {
    loadingResults.value = false
  }
}

const toggleResultDetail = (resultId: number) => {
  if (expandedResultIds.has(resultId)) {
    expandedResultIds.delete(resultId)
  } else {
    expandedResultIds.add(resultId)
  }
}
</script>
