<template>
  <BaseDialog :show="show" :title="t('admin.users.groupConfig')" width="wide" @close="handleClose">
    <div v-if="user" class="space-y-5" :aria-busy="loading">
      <!-- 用户信息头部 -->
      <div class="flex min-w-0 items-start gap-3 rounded-panel border border-outline bg-surface-subtle p-3 sm:items-center sm:p-4">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-outline bg-surface text-sm font-semibold text-foreground-muted" aria-hidden="true">
          {{ user.email.charAt(0).toUpperCase() }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="break-all text-sm font-medium text-foreground">{{ user.email }}</p>
          <p class="mt-0.5 break-words text-xs leading-5 text-foreground-subtle">
            {{ t('admin.users.groupConfigHint', { email: user.email }) }}
          </p>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="flex items-center justify-center gap-2 py-10 text-sm text-foreground-subtle" role="status">
        <Icon name="refresh" size="md" class="animate-spin" aria-hidden="true" />
        <span>{{ t('common.loading') }}</span>
      </div>

      <div v-else-if="loadError" class="flex flex-col items-center gap-3 py-10 text-center" role="alert">
        <p class="text-sm text-danger-foreground">{{ loadError }}</p>
        <button type="button" class="btn btn-secondary" data-testid="allowed-groups-retry" @click="load()">
          {{ t('admin.users.retry') }}
        </button>
      </div>

      <div v-else class="space-y-5">
        <p v-if="saveError" class="rounded-panel border border-danger/20 bg-danger-subtle px-3 py-2 text-sm text-danger-foreground" role="alert">
          {{ saveError }}
        </p>
        <section v-if="exclusiveGroups.length > 0" aria-labelledby="exclusive-groups-heading">
          <div class="mb-2 flex min-w-0 items-center justify-between gap-2">
            <h4 id="exclusive-groups-heading" class="min-w-0 text-sm font-semibold text-foreground">
              {{ t('admin.users.exclusiveGroups') }}
            </h4>
            <span class="badge badge-gray shrink-0 tabular-nums">
              {{ selectedExclusiveGroupCount }}/{{ exclusiveGroupConfigs.length }}
            </span>
          </div>

          <div class="divide-y divide-outline overflow-hidden rounded-panel border border-outline bg-surface">
            <div
              v-for="config in exclusiveGroupConfigs"
              :key="config.groupId"
              class="grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-center gap-x-2 gap-y-2 p-3 transition-colors focus-within:ring-2 focus-within:ring-inset focus-within:ring-focus/30 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:gap-x-3"
              :class="config.isSelected ? 'bg-surface-subtle' : 'bg-surface'"
            >
              <label :for="`exclusive-group-${config.groupId}`" class="flex h-touch w-touch cursor-pointer items-center justify-center rounded-control">
                <input
                  :id="`exclusive-group-${config.groupId}`"
                  type="checkbox"
                  :checked="config.isSelected"
                  :aria-label="`${t('admin.users.exclusiveGroups')}: ${config.groupName}`"
                  class="h-5 w-5 rounded-control border-outline-strong accent-focus focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40"
                  @change="toggleExclusiveGroup(config.groupId)"
                />
              </label>

              <div class="min-w-0">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <span class="min-w-0 break-words text-sm font-medium text-foreground">{{ config.groupName }}</span>
                  <span class="badge badge-gray shrink-0">{{ t('admin.groups.exclusive') }}</span>
                </div>
                <p :id="`group-rate-meta-${config.groupId}`" class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-foreground-subtle">
                  <span class="inline-flex items-center gap-1">
                    <PlatformIcon :platform="config.platform" size="xs" />
                    <span class="break-all">{{ config.platform }}</span>
                  </span>
                  <span aria-hidden="true">/</span>
                  <span>{{ t('admin.users.defaultRate') }}: <strong class="font-medium text-foreground-muted">{{ config.defaultRate }}x</strong></span>
                </p>
              </div>

              <div class="col-span-2 grid min-w-0 grid-cols-[minmax(0,1fr)_7rem] items-center gap-2 sm:col-span-1 sm:flex sm:w-56 sm:shrink-0">
                <label :for="`exclusive-rate-${config.groupId}`" class="min-w-0 text-xs font-medium text-foreground-muted sm:flex-1">
                  {{ t('admin.users.customRate') }}
                </label>
                <input
                  :id="`exclusive-rate-${config.groupId}`"
                  type="number"
                  step="0.001"
                  min="0.001"
                  :value="config.customRate ?? ''"
                  :placeholder="String(config.defaultRate)"
                  :aria-describedby="`group-rate-meta-${config.groupId}`"
                  class="input hide-spinner min-w-0 tabular-nums sm:w-28"
                  @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                />
              </div>
            </div>
          </div>
        </section>

        <section v-if="publicGroups.length > 0" aria-labelledby="public-groups-heading">
          <div class="mb-2 flex min-w-0 items-center justify-between gap-2">
            <h4 id="public-groups-heading" class="min-w-0 text-sm font-semibold text-foreground">
              {{ t('admin.users.publicGroups') }}
            </h4>
            <span class="badge badge-success shrink-0 tabular-nums">{{ publicGroupConfigs.length }}</span>
          </div>

          <div class="divide-y divide-outline overflow-hidden rounded-panel border border-outline bg-surface">
            <div
              v-for="config in publicGroupConfigs"
              :key="config.groupId"
              class="grid min-w-0 grid-cols-[auto_minmax(0,1fr)] items-center gap-x-2 gap-y-2 p-3 focus-within:ring-2 focus-within:ring-inset focus-within:ring-focus/30 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:gap-x-3"
            >
              <div class="flex h-touch w-touch items-center justify-center">
                <input
                  type="checkbox"
                  checked
                  disabled
                  :aria-label="`${t('admin.users.publicGroups')}: ${config.groupName}`"
                  class="h-5 w-5 rounded-control border-outline-strong accent-success"
                />
              </div>

              <div class="min-w-0">
                <span class="block min-w-0 break-words text-sm font-medium text-foreground">{{ config.groupName }}</span>
                <p :id="`group-rate-meta-${config.groupId}`" class="mt-1 flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs text-foreground-subtle">
                  <span class="inline-flex items-center gap-1">
                    <PlatformIcon :platform="config.platform" size="xs" />
                    <span class="break-all">{{ config.platform }}</span>
                  </span>
                  <span aria-hidden="true">/</span>
                  <span>{{ t('admin.users.defaultRate') }}: <strong class="font-medium text-foreground-muted">{{ config.defaultRate }}x</strong></span>
                </p>
              </div>

              <div class="col-span-2 grid min-w-0 grid-cols-[minmax(0,1fr)_7rem] items-center gap-2 sm:col-span-1 sm:flex sm:w-56 sm:shrink-0">
                <label :for="`public-rate-${config.groupId}`" class="min-w-0 text-xs font-medium text-foreground-muted sm:flex-1">
                  {{ t('admin.users.customRate') }}
                </label>
                <input
                  :id="`public-rate-${config.groupId}`"
                  type="number"
                  step="0.001"
                  min="0.001"
                  :value="config.customRate ?? ''"
                  :placeholder="String(config.defaultRate)"
                  :aria-describedby="`group-rate-meta-${config.groupId}`"
                  class="input hide-spinner min-w-0 tabular-nums sm:w-28"
                  @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                />
              </div>
            </div>
          </div>
        </section>

        <div v-if="groups.length === 0" class="flex flex-col items-center justify-center py-10 text-center">
          <div class="mb-3 flex h-12 w-12 items-center justify-center rounded-panel border border-outline bg-surface-subtle text-foreground-subtle">
            <Icon name="inbox" size="lg" aria-hidden="true" />
          </div>
          <p class="text-sm text-foreground-subtle">{{ t('common.noGroupsAvailable') }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="grid w-full grid-cols-2 gap-2 sm:flex sm:justify-end sm:gap-3">
        <button type="button" class="btn btn-secondary sm:px-5" @click="handleClose">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary sm:px-6" :disabled="submitting || !canSave" data-testid="allowed-groups-save" @click="handleSave">
          <Icon v-if="submitting" name="refresh" size="sm" class="animate-spin" aria-hidden="true" />
          {{ submitting ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser, Group, GroupPlatform } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import Icon from '@/components/icons/Icon.vue'

interface GroupRateConfig {
  groupId: number
  groupName: string
  platform: GroupPlatform
  isExclusive: boolean
  defaultRate: number
  customRate: number | null
  isSelected: boolean
}

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const groups = ref<Group[]>([])
const groupConfigs = ref<GroupRateConfig[]>([])
const originalGroupRates = ref<Record<number, number>>({}) // 记录原始专属倍率，用于检测删除
const loading = ref(false)
const submitting = ref(false)
const loadError = ref('')
const saveError = ref('')
const loadedUserId = ref<number | null>(null)
let loadRequestSeq = 0
let saveRequestSeq = 0

// 分离专属分组和公开分组
const exclusiveGroups = computed(() => groups.value.filter((g) => g.is_exclusive))
const publicGroups = computed(() => groups.value.filter((g) => !g.is_exclusive))

const exclusiveGroupConfigs = computed(() => groupConfigs.value.filter((c) => c.isExclusive))
const publicGroupConfigs = computed(() => groupConfigs.value.filter((c) => !c.isExclusive))
const selectedExclusiveGroupCount = computed(() => exclusiveGroupConfigs.value.filter((config) => config.isSelected).length)
const canSave = computed(
  () => loadedUserId.value != null && loadedUserId.value === props.user?.id && !loading.value && !loadError.value
)

watch(
  [() => props.show, () => props.user?.id],
  ([show, userId]) => {
    loadRequestSeq += 1
    saveRequestSeq += 1
    clearData()
    if (show && userId != null) void load(userId)
  },
  { immediate: true }
)

function clearData() {
  groups.value = []
  groupConfigs.value = []
  originalGroupRates.value = {}
  loading.value = false
  submitting.value = false
  loadError.value = ''
  saveError.value = ''
  loadedUserId.value = null
}

function isCurrentLoad(requestSeq: number, userId: number): boolean {
  return requestSeq === loadRequestSeq && props.show && props.user?.id === userId
}

async function load(userId = props.user?.id) {
  if (userId == null || !props.show) return
  const requestSeq = ++loadRequestSeq
  const userAllowedGroups = [...(props.user?.allowed_groups || [])]
  const userGroupRates = { ...(props.user?.group_rates || {}) }
  groups.value = []
  groupConfigs.value = []
  originalGroupRates.value = {}
  loadError.value = ''
  saveError.value = ''
  loadedUserId.value = null
  loading.value = true
  try {
    const res = await adminAPI.groups.list(1, 1000)
    if (!isCurrentLoad(requestSeq, userId)) return
    // 只显示标准类型且活跃的分组
    groups.value = res.items.filter((g) => g.subscription_type === 'standard' && g.status === 'active')

    // 初始化配置
    // 保存原始专属倍率，用于检测删除操作
    originalGroupRates.value = { ...userGroupRates }

    groupConfigs.value = groups.value.map((g) => ({
      groupId: g.id,
      groupName: g.name,
      platform: g.platform,
      isExclusive: g.is_exclusive,
      defaultRate: g.rate_multiplier,
      customRate: userGroupRates[g.id] ?? null,
      // 专属分组：检查是否在 allowed_groups 中
      // 公开分组：始终选中
      isSelected: g.is_exclusive ? userAllowedGroups.includes(g.id) : true,
    }))
    loadedUserId.value = userId
  } catch (error) {
    if (!isCurrentLoad(requestSeq, userId)) return
    console.error('Failed to load groups:', error)
    loadError.value = t('admin.users.failedToLoadGroups')
  } finally {
    if (isCurrentLoad(requestSeq, userId)) loading.value = false
  }
}

const toggleExclusiveGroup = (groupId: number) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config && config.isExclusive) {
    config.isSelected = !config.isSelected
  }
}

const updateCustomRate = (groupId: number, value: string) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config) {
    if (value === '' || value === null || value === undefined) {
      config.customRate = null
    } else {
      const numValue = parseFloat(value)
      config.customRate = isNaN(numValue) ? null : numValue
    }
  }
}

const handleSave = async () => {
  const userId = props.user?.id
  if (userId == null || !canSave.value || loadedUserId.value !== userId) return
  const requestSeq = ++saveRequestSeq
  saveError.value = ''
  submitting.value = true

  try {
    // 构建 allowed_groups（仅包含专属分组中被勾选的）
    const allowedGroups = groupConfigs.value.filter((c) => c.isExclusive && c.isSelected).map((c) => c.groupId)

    // 构建 group_rates
    // - 有新专属倍率: 设置为该值
    // - 原本有专属倍率但现在被清空: 设置为 null（表示删除）
    const groupRates: Record<number, number | null> = {}
    for (const c of groupConfigs.value) {
      const hadOriginalRate = originalGroupRates.value[c.groupId] !== undefined

      if (c.customRate !== null) {
        // 有专属倍率
        groupRates[c.groupId] = c.customRate
      } else if (hadOriginalRate) {
        // 原本有专属倍率，现在被清空，需要显式删除
        groupRates[c.groupId] = null
      }
    }

    await adminAPI.users.update(userId, {
      allowed_groups: allowedGroups,
      group_rates: Object.keys(groupRates).length > 0 ? groupRates : undefined,
    })

    if (requestSeq !== saveRequestSeq || !props.show || props.user?.id !== userId) return
    appStore.showSuccess(t('admin.users.groupConfigUpdated'))
    emit('success')
    emit('close')
  } catch (error) {
    if (requestSeq === saveRequestSeq && props.show && props.user?.id === userId) {
      console.error('Failed to update user group config:', error)
      saveError.value = t('admin.users.failedToUpdateAllowedGroups')
      appStore.showError(saveError.value)
    }
  } finally {
    if (requestSeq === saveRequestSeq) submitting.value = false
  }
}

function handleClose() {
  loadRequestSeq += 1
  saveRequestSeq += 1
  clearData()
  emit('close')
}

onUnmounted(() => {
  loadRequestSeq += 1
  saveRequestSeq += 1
})
</script>

<style scoped>
/* 隐藏数字输入框的箭头按钮 */
.hide-spinner::-webkit-outer-spin-button,
.hide-spinner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.hide-spinner {
  -moz-appearance: textfield;
}
</style>
