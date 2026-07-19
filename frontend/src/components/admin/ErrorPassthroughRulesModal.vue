<template>
  <BaseDialog
    :show="show"
    :title="t('admin.errorPassthrough.title')"
    width="extra-wide"
    @close="$emit('close')"
  >
    <div class="space-y-4">
      <!-- Header -->
      <div class="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-sm text-foreground-subtle">
          {{ t('admin.errorPassthrough.description') }}
        </p>
        <button type="button" class="btn btn-primary btn-sm" @click="showCreateModal = true">
          <Icon name="plus" size="sm" class="mr-1" />
          {{ t('admin.errorPassthrough.createRule') }}
        </button>
      </div>

      <!-- Rules Table -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
      </div>

      <div v-else-if="rules.length === 0" class="py-8 text-center">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-surface-subtle">
          <Icon name="shield" size="lg" class="text-foreground-subtle" />
        </div>
        <h4 class="mb-1 text-sm font-medium text-foreground">
          {{ t('admin.errorPassthrough.noRules') }}
        </h4>
        <p class="text-sm text-foreground-subtle">
          {{ t('admin.errorPassthrough.createFirstRule') }}
        </p>
      </div>

      <template v-else>
        <div data-mobile-layout="rule-cards" class="max-h-[60dvh] space-y-3 overflow-y-auto sm:hidden">
          <article
            v-for="rule in rules"
            :key="`mobile-${rule.id}`"
            class="rounded-panel border border-outline bg-surface p-3 shadow-card"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <span class="inline-flex h-6 min-w-6 shrink-0 items-center justify-center rounded-control bg-surface-subtle px-1 text-xs font-semibold tabular-nums text-foreground-muted">
                    {{ rule.priority }}
                  </span>
                  <h4 class="min-w-0 break-words text-sm font-semibold text-foreground">{{ rule.name }}</h4>
                </div>
                <p v-if="rule.description" class="mt-1 break-words text-xs leading-5 text-foreground-subtle">
                  {{ rule.description }}
                </p>
              </div>
              <button
                type="button"
                role="switch"
                :aria-checked="rule.enabled"
                :aria-label="`${rule.name}: ${t('admin.errorPassthrough.columns.status')}`"
                class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors focus:outline-none focus:ring-2 focus:ring-focus"
                :class="rule.enabled ? 'bg-brand' : 'bg-outline-strong'"
                @click="toggleEnabled(rule)"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transition-transform"
                  :class="rule.enabled ? 'translate-x-5' : 'translate-x-0'"
                />
              </button>
            </div>

            <dl class="mt-3 space-y-3 border-t border-outline pt-3 text-xs">
              <div>
                <dt class="mb-1 font-medium text-foreground-muted">{{ t('admin.errorPassthrough.columns.conditions') }}</dt>
                <dd class="flex flex-wrap gap-1.5">
                  <span v-for="code in rule.error_codes" :key="code" class="badge badge-danger">{{ code }}</span>
                  <span v-for="keyword in rule.keywords" :key="keyword" class="badge badge-gray break-all">"{{ keyword }}"</span>
                  <span v-if="rule.error_codes.length === 0 && rule.keywords.length === 0" class="text-foreground-subtle">—</span>
                </dd>
                <div class="mt-1 text-foreground-subtle">{{ t('admin.errorPassthrough.matchMode.' + rule.match_mode) }}</div>
              </div>

              <div>
                <dt class="mb-1 font-medium text-foreground-muted">{{ t('admin.errorPassthrough.columns.platforms') }}</dt>
                <dd v-if="rule.platforms.length === 0" class="text-foreground-subtle">{{ t('admin.errorPassthrough.allPlatforms') }}</dd>
                <dd v-else class="flex flex-wrap gap-1.5">
                  <span v-for="platform in rule.platforms" :key="platform" class="badge badge-primary">{{ platform }}</span>
                </dd>
              </div>

              <div>
                <dt class="mb-1 font-medium text-foreground-muted">{{ t('admin.errorPassthrough.columns.behavior') }}</dt>
                <dd class="grid gap-1.5 text-foreground-subtle">
                  <span>{{ t('admin.errorPassthrough.code') }}: {{ rule.passthrough_code ? t('admin.errorPassthrough.passthrough') : (rule.response_code || '-') }}</span>
                  <span>{{ t('admin.errorPassthrough.body') }}: {{ rule.passthrough_body ? t('admin.errorPassthrough.passthrough') : t('admin.errorPassthrough.custom') }}</span>
                  <span v-if="rule.skip_monitoring" class="text-warning-foreground">{{ t('admin.errorPassthrough.skipMonitoring') }}</span>
                </dd>
              </div>
            </dl>

            <div class="mt-3 flex gap-2 border-t border-outline pt-3">
              <button type="button" class="btn btn-secondary btn-sm min-w-0 flex-1" @click="handleEdit(rule)">
                <Icon name="edit" size="sm" />
                {{ t('common.edit') }}
              </button>
              <button type="button" class="btn btn-ghost btn-sm min-w-0 flex-1 text-danger-foreground hover:bg-danger-subtle" @click="handleDelete(rule)">
                <Icon name="trash" size="sm" />
                {{ t('common.delete') }}
              </button>
            </div>
          </article>
        </div>

        <div data-desktop-layout="rules-table" class="hidden max-h-96 overflow-auto rounded-panel border border-outline-strong sm:block">
        <table class="min-w-[56rem] divide-y divide-outline">
          <thead class="sticky top-0 bg-surface-subtle">
            <tr>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.priority') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.name') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.conditions') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.platforms') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.behavior') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.status') }}
              </th>
              <th class="px-3 py-2 text-left text-xs font-medium uppercase text-foreground-subtle">
                {{ t('admin.errorPassthrough.columns.actions') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-outline bg-surface">
            <tr v-for="rule in rules" :key="rule.id" class="hover:bg-surface-subtle">
              <td class="whitespace-nowrap px-3 py-2">
                <span class="inline-flex h-5 w-5 items-center justify-center rounded bg-surface-subtle text-xs font-medium text-foreground-muted bg-outline">
                  {{ rule.priority }}
                </span>
              </td>
              <td class="px-3 py-2">
                <div class="font-medium text-foreground text-sm">{{ rule.name }}</div>
                <div v-if="rule.description" class="mt-0.5 text-xs text-foreground-subtle max-w-xs truncate">
                  {{ rule.description }}
                </div>
              </td>
              <td class="px-3 py-2">
                <div class="flex flex-wrap gap-1 max-w-48">
                  <span
                    v-for="code in rule.error_codes.slice(0, 3)"
                    :key="code"
                    class="badge badge-danger text-xs"
                  >
                    {{ code }}
                  </span>
                  <span
                    v-if="rule.error_codes.length > 3"
                    class="text-xs text-foreground-subtle"
                  >
                    +{{ rule.error_codes.length - 3 }}
                  </span>
                  <span
                    v-for="keyword in rule.keywords.slice(0, 1)"
                    :key="keyword"
                    class="badge badge-gray text-xs"
                  >
                    "{{ keyword.length > 10 ? keyword.substring(0, 10) + '...' : keyword }}"
                  </span>
                  <span
                    v-if="rule.keywords.length > 1"
                    class="text-xs text-foreground-subtle"
                  >
                    +{{ rule.keywords.length - 1 }}
                  </span>
                </div>
                <div class="mt-0.5 text-xs text-foreground-subtle">
                  {{ t('admin.errorPassthrough.matchMode.' + rule.match_mode) }}
                </div>
              </td>
              <td class="px-3 py-2">
                <div v-if="rule.platforms.length === 0" class="text-xs text-foreground-subtle">
                  {{ t('admin.errorPassthrough.allPlatforms') }}
                </div>
                <div v-else class="flex flex-wrap gap-1">
                  <span
                    v-for="platform in rule.platforms.slice(0, 2)"
                    :key="platform"
                    class="badge badge-primary text-xs"
                  >
                    {{ platform }}
                  </span>
                  <span v-if="rule.platforms.length > 2" class="text-xs text-foreground-subtle">
                    +{{ rule.platforms.length - 2 }}
                  </span>
                </div>
              </td>
              <td class="px-3 py-2">
                <div class="text-xs space-y-0.5">
                  <div class="flex items-center gap-1">
                    <Icon
                      :name="rule.passthrough_code ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="rule.passthrough_code ? 'text-success-foreground' : 'text-foreground-subtle'"
                    />
                    <span class="text-foreground-subtle">
                      {{ t('admin.errorPassthrough.code') }}:
                      {{ rule.passthrough_code ? t('admin.errorPassthrough.passthrough') : (rule.response_code || '-') }}
                    </span>
                  </div>
                  <div class="flex items-center gap-1">
                    <Icon
                      :name="rule.passthrough_body ? 'checkCircle' : 'xCircle'"
                      size="xs"
                      :class="rule.passthrough_body ? 'text-success-foreground' : 'text-foreground-subtle'"
                    />
                    <span class="text-foreground-subtle">
                      {{ t('admin.errorPassthrough.body') }}:
                      {{ rule.passthrough_body ? t('admin.errorPassthrough.passthrough') : t('admin.errorPassthrough.custom') }}
                    </span>
                  </div>
                  <div v-if="rule.skip_monitoring" class="flex items-center gap-1">
                    <Icon
                      name="checkCircle"
                      size="xs"
                      class="text-warning-foreground"
                    />
                    <span class="text-foreground-subtle">
                      {{ t('admin.errorPassthrough.skipMonitoring') }}
                    </span>
                  </div>
                </div>
              </td>
              <td class="px-3 py-2">
                <button
                  type="button"
                  role="switch"
                  :aria-checked="rule.enabled"
                  :aria-label="`${rule.name}: ${t('admin.errorPassthrough.columns.status')}`"
                  @click="toggleEnabled(rule)"
                  :class="[
                    'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-focus focus:ring-offset-2',
                    rule.enabled ? 'bg-brand' : 'bg-outline'
                  ]"
                >
                  <span
                    :class="[
                      'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                      rule.enabled ? 'translate-x-5' : 'translate-x-0'
                    ]"
                  />
                </button>
              </td>
              <td class="px-3 py-2">
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    @click="handleEdit(rule)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-panel text-foreground-subtle hover:bg-surface-subtle hover:text-brand focus:outline-none focus:ring-2 focus:ring-focus"
                    :aria-label="`${t('common.edit')} ${rule.name}`"
                    :title="t('common.edit')"
                  >
                    <Icon name="edit" size="sm" />
                  </button>
                  <button
                    type="button"
                    @click="handleDelete(rule)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-panel text-foreground-subtle hover:bg-danger-subtle hover:text-danger-foreground focus:outline-none focus:ring-2 focus:ring-danger"
                    :aria-label="`${t('common.delete')} ${rule.name}`"
                    :title="t('common.delete')"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('admin.errorPassthrough.editRule') : t('admin.errorPassthrough.createRule')"
      width="wide"
      @close="closeFormModal"
    >
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <!-- Basic Info -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label for="error-passthrough-name" class="input-label">
              {{ t('admin.errorPassthrough.form.name') }}
            </label>
            <input
              v-model="form.name"
              id="error-passthrough-name"
              type="text"
              required
              class="input"
              :placeholder="t('admin.errorPassthrough.form.namePlaceholder')"
            />
          </div>
          <div>
            <label for="error-passthrough-priority" class="input-label">
              {{ t('admin.errorPassthrough.form.priority') }}
            </label>
            <input
              v-model.number="form.priority"
              id="error-passthrough-priority"
              type="number"
              min="0"
              class="input"
            />
            <p class="input-hint">{{ t('admin.errorPassthrough.form.priorityHint') }}</p>
          </div>
        </div>

        <div>
          <label for="error-passthrough-description" class="input-label">
            {{ t('admin.errorPassthrough.form.description') }}
          </label>
          <input
            v-model="form.description"
            id="error-passthrough-description"
            type="text"
            class="input"
            :placeholder="t('admin.errorPassthrough.form.descriptionPlaceholder')"
          />
        </div>

        <!-- Match Conditions -->
        <div class="rounded-panel border p-3 border-outline-strong">
          <h4 class="mb-2 text-sm font-medium text-foreground">
            {{ t('admin.errorPassthrough.form.matchConditions') }}
          </h4>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label for="error-passthrough-codes" class="input-label text-xs">
                {{ t('admin.errorPassthrough.form.errorCodes') }}
              </label>
              <input
                v-model="errorCodesInput"
                id="error-passthrough-codes"
                type="text"
                class="input text-sm"
                :placeholder="t('admin.errorPassthrough.form.errorCodesPlaceholder')"
              />
              <p class="input-hint text-xs">{{ t('admin.errorPassthrough.form.errorCodesHint') }}</p>
            </div>
            <div>
              <label for="error-passthrough-keywords" class="input-label text-xs">
                {{ t('admin.errorPassthrough.form.keywords') }}
              </label>
              <textarea
                v-model="keywordsInput"
                id="error-passthrough-keywords"
                rows="2"
                class="input font-mono text-xs"
                :placeholder="t('admin.errorPassthrough.form.keywordsPlaceholder')"
              />
              <p class="input-hint text-xs">{{ t('admin.errorPassthrough.form.keywordsHint') }}</p>
            </div>
          </div>

          <div class="mt-3">
            <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.matchMode') }}</label>
            <div class="mt-1 space-y-2">
              <label
                v-for="option in matchModeOptions"
                :key="option.value"
                class="flex items-start gap-2 cursor-pointer"
              >
                <input
                  type="radio"
                  :value="option.value"
                  v-model="form.match_mode"
                  class="mt-0.5 h-3.5 w-3.5 border-outline-strong text-brand focus:ring-focus"
                />
                <div class="flex-1">
                  <span class="text-xs font-medium text-foreground-muted">{{ option.label }}</span>
                  <p class="text-xs text-foreground-subtle">{{ option.description }}</p>
                </div>
              </label>
            </div>
          </div>

          <div class="mt-3">
            <label class="input-label text-xs">{{ t('admin.errorPassthrough.form.platforms') }}</label>
            <div class="flex flex-wrap gap-3">
              <label
                v-for="platform in platformOptions"
                :key="platform.value"
                class="inline-flex items-center gap-1.5"
              >
                <input
                  type="checkbox"
                  :value="platform.value"
                  v-model="form.platforms"
                  class="h-3.5 w-3.5 rounded border-outline-strong text-brand focus:ring-focus"
                />
                <span class="text-xs text-foreground-muted">{{ platform.label }}</span>
              </label>
            </div>
            <p class="input-hint text-xs mt-1">{{ t('admin.errorPassthrough.form.platformsHint') }}</p>
          </div>
        </div>

        <!-- Response Behavior -->
        <div class="rounded-panel border p-3 border-outline-strong">
          <h4 class="mb-2 text-sm font-medium text-foreground">
            {{ t('admin.errorPassthrough.form.responseBehavior') }}
          </h4>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label class="flex items-center gap-1.5">
                <input
                  type="checkbox"
                  v-model="form.passthrough_code"
                  class="h-3.5 w-3.5 rounded border-outline-strong text-brand focus:ring-focus"
                />
                <span class="text-xs font-medium text-foreground-muted">
                  {{ t('admin.errorPassthrough.form.passthroughCode') }}
                </span>
              </label>
              <div v-if="!form.passthrough_code" class="mt-2">
                <label for="error-passthrough-response-code" class="input-label text-xs">
                  {{ t('admin.errorPassthrough.form.responseCode') }}
                </label>
                <input
                  v-model.number="form.response_code"
                  id="error-passthrough-response-code"
                  type="number"
                  min="100"
                  max="599"
                  class="input text-sm"
                  placeholder="422"
                />
              </div>
            </div>
            <div>
              <label class="flex items-center gap-1.5">
                <input
                  type="checkbox"
                  v-model="form.passthrough_body"
                  class="h-3.5 w-3.5 rounded border-outline-strong text-brand focus:ring-focus"
                />
                <span class="text-xs font-medium text-foreground-muted">
                  {{ t('admin.errorPassthrough.form.passthroughBody') }}
                </span>
              </label>
              <div v-if="!form.passthrough_body" class="mt-2">
                <label for="error-passthrough-custom-message" class="input-label text-xs">
                  {{ t('admin.errorPassthrough.form.customMessage') }}
                </label>
                <input
                  v-model="form.custom_message"
                  id="error-passthrough-custom-message"
                  type="text"
                  class="input text-sm"
                  :placeholder="t('admin.errorPassthrough.form.customMessagePlaceholder')"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Skip Monitoring -->
        <label class="flex items-center gap-1.5">
          <input
            type="checkbox"
            v-model="form.skip_monitoring"
            class="h-3.5 w-3.5 rounded border-outline-strong text-warning-foreground focus:ring-warning/20"
          />
          <span class="text-xs font-medium text-foreground-muted">
            {{ t('admin.errorPassthrough.form.skipMonitoring') }}
          </span>
        </label>
        <p class="input-hint text-xs -mt-3">{{ t('admin.errorPassthrough.form.skipMonitoringHint') }}</p>

        <!-- Enabled -->
        <label class="flex items-center gap-1.5">
          <input
            type="checkbox"
            v-model="form.enabled"
            class="h-3.5 w-3.5 rounded border-outline-strong text-brand focus:ring-focus"
          />
          <span class="text-xs font-medium text-foreground-muted">
            {{ t('admin.errorPassthrough.form.enabled') }}
          </span>
        </label>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeFormModal" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleSubmit" :disabled="submitting" class="btn btn-primary">
            <Icon v-if="submitting" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ showEditModal ? t('common.update') : t('common.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.errorPassthrough.deleteRule')"
      :message="t('admin.errorPassthrough.deleteConfirm', { name: deletingRule?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { ErrorPassthroughRule } from '@/api/admin/errorPassthrough'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

// eslint-disable-next-line @typescript-eslint/no-unused-vars
void emit // suppress unused warning - emit is used via $emit in template

const { t } = useI18n()
const appStore = useAppStore()

const rules = ref<ErrorPassthroughRule[]>([])
const loading = ref(false)
const submitting = ref(false)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const editingRule = ref<ErrorPassthroughRule | null>(null)
const deletingRule = ref<ErrorPassthroughRule | null>(null)

// Form inputs for arrays
const errorCodesInput = ref('')
const keywordsInput = ref('')

const form = reactive({
  name: '',
  enabled: true,
  priority: 0,
  match_mode: 'any' as 'any' | 'all',
  platforms: [] as string[],
  passthrough_code: true,
  response_code: null as number | null,
  passthrough_body: true,
  custom_message: null as string | null,
  skip_monitoring: false,
  description: null as string | null
})

const matchModeOptions = computed(() => [
  { value: 'any', label: t('admin.errorPassthrough.matchMode.any'), description: t('admin.errorPassthrough.matchMode.anyHint') },
  { value: 'all', label: t('admin.errorPassthrough.matchMode.all'), description: t('admin.errorPassthrough.matchMode.allHint') }
])

const platformOptions = [
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'antigravity', label: 'Antigravity' },
  { value: 'grok', label: 'Grok' }
]

// Load rules when dialog opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadRules()
  }
})

const loadRules = async () => {
  loading.value = true
  try {
    rules.value = await adminAPI.errorPassthrough.list()
  } catch (error) {
    appStore.showError(t('admin.errorPassthrough.failedToLoad'))
    console.error('Error loading rules:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.enabled = true
  form.priority = 0
  form.match_mode = 'any'
  form.platforms = []
  form.passthrough_code = true
  form.response_code = null
  form.passthrough_body = true
  form.custom_message = null
  form.skip_monitoring = false
  form.description = null
  errorCodesInput.value = ''
  keywordsInput.value = ''
}

const closeFormModal = () => {
  showCreateModal.value = false
  showEditModal.value = false
  editingRule.value = null
  resetForm()
}

const handleEdit = (rule: ErrorPassthroughRule) => {
  editingRule.value = rule
  form.name = rule.name
  form.enabled = rule.enabled
  form.priority = rule.priority
  form.match_mode = rule.match_mode
  form.platforms = [...rule.platforms]
  form.passthrough_code = rule.passthrough_code
  form.response_code = rule.response_code
  form.passthrough_body = rule.passthrough_body
  form.custom_message = rule.custom_message
  form.skip_monitoring = rule.skip_monitoring
  form.description = rule.description
  errorCodesInput.value = rule.error_codes.join(', ')
  keywordsInput.value = rule.keywords.join('\n')
  showEditModal.value = true
}

const handleDelete = (rule: ErrorPassthroughRule) => {
  deletingRule.value = rule
  showDeleteDialog.value = true
}

const parseErrorCodes = (): number[] => {
  if (!errorCodesInput.value.trim()) return []
  return errorCodesInput.value
    .split(/[,\s]+/)
    .map(s => parseInt(s.trim(), 10))
    .filter(n => !isNaN(n) && n > 0)
}

const parseKeywords = (): string[] => {
  if (!keywordsInput.value.trim()) return []
  return keywordsInput.value
    .split('\n')
    .map(s => s.trim())
    .filter(s => s.length > 0)
}

const handleSubmit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('admin.errorPassthrough.nameRequired'))
    return
  }

  const errorCodes = parseErrorCodes()
  const keywords = parseKeywords()

  if (errorCodes.length === 0 && keywords.length === 0) {
    appStore.showError(t('admin.errorPassthrough.conditionsRequired'))
    return
  }

  submitting.value = true
  try {
    const data = {
      name: form.name.trim(),
      enabled: form.enabled,
      priority: form.priority,
      error_codes: errorCodes,
      keywords: keywords,
      match_mode: form.match_mode,
      platforms: form.platforms,
      passthrough_code: form.passthrough_code,
      response_code: form.passthrough_code ? null : form.response_code,
      passthrough_body: form.passthrough_body,
      custom_message: form.passthrough_body ? null : form.custom_message,
      skip_monitoring: form.skip_monitoring,
      description: form.description?.trim() || null
    }

    if (showEditModal.value && editingRule.value) {
      await adminAPI.errorPassthrough.update(editingRule.value.id, data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleUpdated'))
    } else {
      await adminAPI.errorPassthrough.create(data)
      appStore.showSuccess(t('admin.errorPassthrough.ruleCreated'))
    }

    closeFormModal()
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToSave'))
    console.error('Error saving rule:', error)
  } finally {
    submitting.value = false
  }
}

const toggleEnabled = async (rule: ErrorPassthroughRule) => {
  try {
    await adminAPI.errorPassthrough.toggleEnabled(rule.id, !rule.enabled)
    rule.enabled = !rule.enabled
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToToggle'))
    console.error('Error toggling rule:', error)
  }
}

const confirmDelete = async () => {
  if (!deletingRule.value) return

  try {
    await adminAPI.errorPassthrough.delete(deletingRule.value.id)
    appStore.showSuccess(t('admin.errorPassthrough.ruleDeleted'))
    showDeleteDialog.value = false
    deletingRule.value = null
    loadRules()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.errorPassthrough.failedToDelete'))
    console.error('Error deleting rule:', error)
  }
}
</script>
