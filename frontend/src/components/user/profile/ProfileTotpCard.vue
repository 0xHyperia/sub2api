<template>
  <section class="card min-w-0 overflow-hidden" aria-labelledby="profile-totp-title">
    <header class="card-header">
      <h2 id="profile-totp-title" class="text-base font-semibold text-foreground">
        {{ t('profile.totp.title') }}
      </h2>
      <p class="mt-1 text-sm text-foreground-muted">
        {{ t('profile.totp.description') }}
      </p>
    </header>
    <div class="p-4 sm:p-5">
      <div v-if="loading" class="flex min-h-24 items-center justify-center" role="status" :aria-label="t('common.loading')">
        <span class="h-7 w-7 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true" />
      </div>

      <div
        v-else-if="loadFailed"
        data-testid="profile-totp-load-error"
        class="flex min-w-0 flex-col gap-4 rounded-panel border border-danger/20 bg-danger-subtle p-4 sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-danger/20 bg-surface/60 text-danger-foreground" aria-hidden="true">
            <Icon name="exclamationTriangle" size="md" />
          </span>
          <div class="min-w-0">
            <p class="text-sm font-medium text-danger-foreground">{{ t('profile.totp.loadFailed') }}</p>
            <p class="mt-0.5 text-sm text-danger-foreground/80">{{ t('errors.tryAgain') }}</p>
          </div>
        </div>
        <button type="button" class="btn btn-secondary shrink-0" @click="loadStatus">
          <Icon name="refresh" size="sm" aria-hidden="true" />
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-else-if="status && !status.feature_enabled" class="flex min-w-0 items-start gap-3">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-subtle" aria-hidden="true">
          <Icon name="exclamationTriangle" size="md" />
        </span>
        <div class="min-w-0">
          <p class="text-sm font-medium text-foreground">
            {{ t('profile.totp.featureDisabled') }}
          </p>
          <p class="mt-0.5 text-sm text-foreground-muted">
            {{ t('profile.totp.featureDisabledHint') }}
          </p>
        </div>
      </div>

      <div v-else-if="status?.enabled" class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-success/20 bg-success-subtle text-success-foreground" aria-hidden="true">
            <Icon name="shield" size="md" />
          </span>
          <div class="min-w-0">
            <p class="text-sm font-medium text-foreground">
              {{ t('profile.totp.enabled') }}
            </p>
            <p v-if="status.enabled_at" class="mt-0.5 text-sm text-foreground-muted">
              {{ t('profile.totp.enabledAt') }}: {{ formatDate(status.enabled_at) }}
            </p>
          </div>
        </div>
        <button
          type="button"
          class="btn btn-secondary shrink-0 text-danger-foreground"
          @click="showDisableDialog = true"
        >
          {{ t('profile.totp.disable') }}
        </button>
      </div>

      <div v-else-if="status" class="flex min-w-0 flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div class="flex min-w-0 items-start gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-muted" aria-hidden="true">
            <Icon name="shield" size="md" />
          </span>
          <div class="min-w-0">
            <p class="text-sm font-medium text-foreground">
              {{ t('profile.totp.notEnabled') }}
            </p>
            <p class="mt-0.5 text-sm text-foreground-muted">
              {{ t('profile.totp.notEnabledHint') }}
            </p>
          </div>
        </div>
        <button
          type="button"
          class="btn btn-primary shrink-0"
          @click="showSetupModal = true"
        >
          {{ t('profile.totp.enable') }}
        </button>
      </div>
    </div>

    <TotpSetupModal
      v-if="showSetupModal"
      @close="showSetupModal = false"
      @success="handleSetupSuccess"
    />

    <TotpDisableDialog
      v-if="showDisableDialog"
      @close="showDisableDialog = false"
      @success="handleDisableSuccess"
    />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { totpAPI } from '@/api'
import Icon from '@/components/icons/Icon.vue'
import type { TotpStatus } from '@/types'
import TotpSetupModal from './TotpSetupModal.vue'
import TotpDisableDialog from './TotpDisableDialog.vue'

const { t } = useI18n()

const loading = ref(true)
const status = ref<TotpStatus | null>(null)
const loadFailed = ref(false)
const showSetupModal = ref(false)
const showDisableDialog = ref(false)

const loadStatus = async () => {
  loading.value = true
  loadFailed.value = false
  try {
    status.value = await totpAPI.getStatus()
  } catch (error) {
    loadFailed.value = true
    console.error('Failed to load TOTP status:', error)
  } finally {
    loading.value = false
  }
}

const handleSetupSuccess = () => {
  showSetupModal.value = false
  loadStatus()
}

const handleDisableSuccess = () => {
  showDisableDialog.value = false
  loadStatus()
}

const formatDate = (timestamp: number) => {
  // Backend returns Unix timestamp in seconds, convert to milliseconds
  const date = new Date(timestamp * 1000)
  return date.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  loadStatus()
})
</script>
