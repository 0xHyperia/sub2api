<template>
  <BaseDialog :show="show" :title="t('usage.errors.detail.title')" width="wide" @close="emit('update:show', false)">
    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-10" role="status" :aria-label="t('common.loading')">
      <LoadingSpinner />
    </div>

    <!-- Error state -->
    <div
      v-else-if="loadError"
      class="flex min-h-40 flex-col items-center justify-center gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-5 text-center text-danger-foreground"
      role="alert"
      data-testid="user-error-detail-load-error"
    >
      <Icon name="exclamationTriangle" size="lg" aria-hidden="true" />
      <p class="text-sm font-medium">{{ t('usage.errors.detail.loadFailed') }}</p>
      <button type="button" class="btn btn-secondary btn-sm" @click="loadCurrentDetail">
        <Icon name="refresh" size="sm" aria-hidden="true" />
        {{ t('common.retry') }}
      </button>
    </div>

    <!-- Detail content -->
    <div v-else-if="detail" class="min-w-0 space-y-4 text-sm">
      <dl class="grid grid-cols-1 gap-x-6 gap-y-3 min-[420px]:grid-cols-2">
        <!-- Time -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.time') }}</dt>
          <dd class="mt-0.5 break-words text-foreground">{{ formatDateTime(detail.created_at) }}</dd>
        </div>
        <!-- Model -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.model') }}</dt>
          <dd class="mt-0.5 break-all text-foreground">{{ detail.model || '-' }}</dd>
        </div>
        <!-- Endpoint -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.endpoint') }}</dt>
          <dd class="mt-0.5 break-all text-foreground">{{ detail.inbound_endpoint || '-' }}</dd>
        </div>
        <!-- Status Code -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.status') }}</dt>
          <dd class="mt-0.5">
            <span class="badge" :class="statusClass(detail.status_code)">{{ detail.status_code || '-' }}</span>
          </dd>
        </div>
        <!-- Category -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.category') }}</dt>
          <dd class="mt-0.5 break-words text-foreground">{{ t('usage.errors.categories.' + detail.category) }}</dd>
        </div>
        <!-- Platform -->
        <div class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.platform') }}</dt>
          <dd class="mt-0.5 break-all text-foreground">{{ detail.platform || '-' }}</dd>
        </div>
        <!-- Upstream status code -->
        <div v-if="detail.upstream_status_code != null" class="min-w-0">
          <dt class="font-medium text-foreground-muted">{{ t('usage.errors.detail.upstreamStatus') }}</dt>
          <dd class="mt-0.5 text-foreground">{{ detail.upstream_status_code }}</dd>
        </div>
      </dl>

      <!-- Message -->
      <div v-if="detail.message">
        <span class="font-medium text-foreground-muted">{{ t('usage.errors.message') }}</span>
        <p class="mt-0.5 break-all text-foreground">{{ detail.message }}</p>
      </div>

      <!-- Error Body -->
      <div v-if="detail.error_body">
        <span class="font-medium text-foreground-muted">{{ t('usage.errors.detail.responseBody') }}</span>
        <pre class="mt-1 max-h-[40vh] max-w-full overflow-auto whitespace-pre-wrap break-all rounded-panel border border-outline bg-surface-subtle p-3 text-xs text-foreground">{{ detail.error_body }}</pre>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import { getMyErrorDetail } from '@/api/usage'
import { formatDateTime } from '@/utils/format'
import type { UserErrorRequestDetail } from '@/types'

const props = defineProps<{
  show: boolean
  errorId: number | null
}>()

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
}>()

const { t } = useI18n()

const loading = ref(false)
const loadError = ref(false)
const detail = ref<UserErrorRequestDetail | null>(null)
let requestSequence = 0

watch(
  () => [props.show, props.errorId] as const,
  ([show, id], _, onCleanup) => {
    onCleanup(() => { requestSequence += 1 })
    if (show && id != null) {
      detail.value = null
      void fetchDetail(id)
    } else if (!show) {
      requestSequence += 1
      detail.value = null
      loadError.value = false
      loading.value = false
    }
  },
  { immediate: true }
)

async function fetchDetail(id: number) {
  const sequence = ++requestSequence
  loading.value = true
  loadError.value = false
  try {
    const response = await getMyErrorDetail(id)
    if (sequence !== requestSequence || !props.show || props.errorId !== id) return
    detail.value = response
  } catch (e) {
    if (sequence !== requestSequence) return
    console.error('[UserErrorDetailModal] Failed to load error detail:', e)
    loadError.value = true
  } finally {
    if (sequence === requestSequence) {
      loading.value = false
    }
  }
}

function loadCurrentDetail() {
  if (props.errorId != null) {
    void fetchDetail(props.errorId)
  }
}

function statusClass(code: number) {
  if (code >= 500) return 'badge-danger'
  if (code === 429) return 'badge-warning'
  return 'badge-gray'
}
</script>
