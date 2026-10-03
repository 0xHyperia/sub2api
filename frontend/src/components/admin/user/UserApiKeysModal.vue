<template>
  <BaseDialog :show="show" :title="t('admin.users.userApiKeys')" width="wide" @close="handleClose">
    <div v-if="user" class="space-y-4" :aria-busy="loading">
      <div class="flex min-w-0 items-center gap-3 rounded-panel border border-outline bg-surface-subtle p-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-outline bg-surface text-sm font-semibold text-foreground-muted" aria-hidden="true">
          {{ user.email.charAt(0).toUpperCase() }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="break-all text-sm font-medium text-foreground">{{ user.email }}</p>
          <p class="mt-0.5 break-all text-xs text-foreground-subtle">{{ user.username }}</p>
        </div>
      </div>

      <div v-if="loading" class="flex items-center justify-center gap-2 py-8 text-sm text-foreground-subtle" role="status">
        <Icon name="refresh" size="md" class="animate-spin" aria-hidden="true" />
        <span>{{ t('common.loading') }}</span>
      </div>
      <div v-else-if="loadError" class="flex flex-col items-center gap-3 py-8 text-center" role="alert">
        <p class="text-sm text-danger-foreground">{{ loadError }}</p>
        <button type="button" class="btn btn-secondary" data-testid="api-keys-retry" @click="loadData()">
          {{ t('admin.users.retry') }}
        </button>
      </div>
      <div v-else-if="apiKeys.length === 0" class="py-8 text-center">
        <p class="text-sm text-foreground-subtle">{{ t('admin.users.noApiKeys') }}</p>
      </div>
      <div v-else class="max-h-[28rem] divide-y divide-outline overflow-y-auto rounded-panel border border-outline bg-surface">
        <article v-for="key in apiKeys" :key="key.id" class="min-w-0 p-3 sm:p-4">
          <div class="flex min-w-0 flex-wrap items-start justify-between gap-2">
            <h4 class="min-w-0 flex-1 break-all text-sm font-medium text-foreground">{{ key.name }}</h4>
            <span :class="['badge shrink-0', key.status === 'active' ? 'badge-success' : 'badge-danger']">
              {{ key.status }}
            </span>
          </div>
          <p class="mt-1 max-w-full break-all font-mono text-xs leading-5 text-foreground-subtle" data-testid="api-key-preview">
            {{ formatKeyPreview(key.key) }}
          </p>

          <div class="mt-3 grid min-w-0 gap-3 border-t border-outline pt-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
            <div class="min-w-0">
              <span class="mb-1.5 block text-xs font-medium text-foreground-muted">{{ t('admin.users.group') }}</span>
              <div class="flex min-w-0 items-center gap-2">
                <Select
                  :model-value="key.group_id ?? null"
                  :options="groupOptions"
                  value-key="value"
                  label-key="label"
                  :searchable="allGroups.length > 5"
                  :disabled="updatingKeyIds.has(key.id)"
                  :aria-label="`${t('admin.users.group')}: ${key.name}`"
                  class="min-w-0 flex-1 sm:max-w-md"
                  data-testid="api-key-group-select"
                  @change="handleGroupChange(key, $event)"
                >
                  <template #selected>
                    <span class="block min-w-0 overflow-hidden">
                      <GroupBadge
                        v-if="key.group_id && key.group"
                        :name="key.group.name"
                        :platform="key.group.platform"
                        :subscription-type="key.group.subscription_type"
                        :rate-multiplier="key.group.rate_multiplier"
                        :peak-rate-enabled="key.group.peak_rate_enabled"
                        :peak-start="key.group.peak_start"
                        :peak-end="key.group.peak_end"
                        :peak-rate-multiplier="key.group.peak_rate_multiplier"
                        class="max-w-full"
                      />
                      <span v-else class="italic text-foreground-subtle">{{ t('admin.users.none') }}</span>
                    </span>
                  </template>
                  <template #option="{ option, selected }">
                    <span v-if="option.value === null" class="min-w-0 flex-1 text-left italic text-foreground-subtle">
                      {{ t('admin.users.none') }}
                    </span>
                    <GroupOptionItem
                      v-else
                      :name="option.group.name"
                      :platform="option.group.platform"
                      :subscription-type="option.group.subscription_type"
                      :rate-multiplier="option.group.rate_multiplier"
                      :peak-rate-enabled="option.group.peak_rate_enabled"
                      :peak-start="option.group.peak_start"
                      :peak-end="option.group.peak_end"
                      :peak-rate-multiplier="option.group.peak_rate_multiplier"
                      :description="option.group.description"
                      :selected="selected"
                    />
                  </template>
                </Select>
                <span v-if="updatingKeyIds.has(key.id)" class="shrink-0 text-foreground-subtle" role="status">
                  <Icon name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
                  <span class="sr-only">{{ t('common.saving') }}</span>
                </span>
              </div>
            </div>
            <p class="min-w-0 break-words text-xs text-foreground-subtle sm:pb-2.5 sm:text-right">
              {{ t('admin.users.columns.created') }}: {{ formatDateTime(key.created_at) }}
            </p>
          </div>
        </article>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { AdminUser, AdminGroup, ApiKey } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close'])
const { t } = useI18n()
const appStore = useAppStore()

const apiKeys = ref<ApiKey[]>([])
const allGroups = ref<AdminGroup[]>([])
const loading = ref(false)
const loadError = ref('')
const updatingKeyIds = ref(new Set<number>())
let loadRequestSeq = 0

const groupOptions = computed(() => [
  { value: null, label: t('admin.users.none'), group: null },
  ...allGroups.value.map((group) => ({ value: group.id, label: group.name, group }))
])

function clearData() {
  apiKeys.value = []
  allGroups.value = []
  loadError.value = ''
  loading.value = false
  updatingKeyIds.value.clear()
}

function isCurrentLoad(requestSeq: number, userId: number): boolean {
  return requestSeq === loadRequestSeq && props.show && props.user?.id === userId
}

async function loadData(userId = props.user?.id) {
  if (userId == null || !props.show) return
  const requestSeq = ++loadRequestSeq
  apiKeys.value = []
  allGroups.value = []
  loadError.value = ''
  loading.value = true
  try {
    const [res, groups] = await Promise.all([
      adminAPI.users.getUserApiKeys(userId),
      adminAPI.groups.getAll()
    ])
    if (!isCurrentLoad(requestSeq, userId)) return
    apiKeys.value = res.items || []
    allGroups.value = groups ?? []
  } catch (error) {
    if (!isCurrentLoad(requestSeq, userId)) return
    console.error('Failed to load API keys:', error)
    loadError.value = t('admin.users.failedToLoadApiKeys')
  } finally {
    if (isCurrentLoad(requestSeq, userId)) loading.value = false
  }
}

watch(
  [() => props.show, () => props.user?.id],
  ([show, userId]) => {
    loadRequestSeq += 1
    clearData()
    if (show && userId != null) void loadData(userId)
  },
  { immediate: true }
)

const formatKeyPreview = (value: string) => `${value.substring(0, 20)}...${value.substring(Math.max(0, value.length - 8))}`

const handleGroupChange = (key: ApiKey, value: unknown) => {
  void changeGroup(key, typeof value === 'number' ? value : null)
}

const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
  if (key.group_id === newGroupId || (!key.group_id && newGroupId === null)) return

  const requestSeq = loadRequestSeq
  const userId = props.user?.id
  if (userId == null) return
  updatingKeyIds.value.add(key.id)
  try {
    const result = await adminAPI.apiKeys.updateApiKeyGroup(key.id, newGroupId)
    if (!isCurrentLoad(requestSeq, userId)) return
    // Update local data
    const idx = apiKeys.value.findIndex((k) => k.id === key.id)
    if (idx !== -1) {
      apiKeys.value[idx] = result.api_key
    }
    if (result.auto_granted_group_access && result.granted_group_name) {
      appStore.showSuccess(t('admin.users.groupChangedWithGrant', { group: result.granted_group_name }))
    } else {
      appStore.showSuccess(t('admin.users.groupChangedSuccess'))
    }
  } catch (error: any) {
    if (isCurrentLoad(requestSeq, userId)) {
      appStore.showError(error?.message || t('admin.users.groupChangeFailed'))
    }
  } finally {
    if (isCurrentLoad(requestSeq, userId)) updatingKeyIds.value.delete(key.id)
  }
}

const handleClose = () => {
  loadRequestSeq += 1
  clearData()
  emit('close')
}

onUnmounted(() => {
  loadRequestSeq += 1
})
</script>
