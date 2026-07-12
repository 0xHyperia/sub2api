<template>
  <section
    class="card min-w-0 overflow-hidden"
    data-testid="authorized-apps-card"
    aria-labelledby="profile-authorized-apps-title"
  >
    <header class="card-header">
      <h2 id="profile-authorized-apps-title" class="text-base font-semibold text-foreground">
        {{ t('profile.authorizedApps.title') }}
      </h2>
      <p class="mt-1 text-sm text-foreground-muted">
        {{ t('profile.authorizedApps.description') }}
      </p>
    </header>

    <div class="p-4 sm:p-5">
      <div
        v-if="loading"
        class="flex min-h-24 items-center justify-center"
        role="status"
        :aria-label="t('common.loading')"
      >
        <span
          class="h-7 w-7 animate-spin rounded-full border-2 border-outline-strong border-t-foreground"
          aria-hidden="true"
        />
      </div>

      <div
        v-else-if="loadError"
        data-testid="authorized-apps-load-error"
        class="flex min-w-0 flex-col gap-4 rounded-panel border border-danger/20 bg-danger-subtle p-4 sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <div class="flex min-w-0 items-start gap-3">
          <span
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-danger/20 bg-surface/60 text-danger-foreground"
            aria-hidden="true"
          >
            <Icon name="exclamationTriangle" size="md" />
          </span>
          <p class="min-w-0 break-words text-sm font-medium text-danger-foreground">
            {{ loadError }}
          </p>
        </div>
        <button type="button" class="btn btn-secondary shrink-0" @click="loadDevices">
          <Icon name="refresh" size="sm" aria-hidden="true" />
          {{ t('profile.authorizedApps.retry') }}
        </button>
      </div>

      <div
        v-else-if="devices.length === 0"
        class="flex min-h-28 flex-col items-center justify-center gap-3 text-center text-sm text-foreground-muted"
      >
        <span
          class="flex h-10 w-10 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-subtle"
          aria-hidden="true"
        >
          <Icon name="key" size="md" />
        </span>
        <p>{{ t('profile.authorizedApps.empty') }}</p>
      </div>

      <ul v-else class="divide-y divide-outline overflow-hidden rounded-panel border border-outline">
        <li
          v-for="device in devices"
          :key="device.id"
          class="grid min-w-0 grid-cols-[2.5rem_minmax(0,1fr)_2.25rem] items-start gap-3 p-3 sm:p-4"
        >
          <span
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-control border border-outline bg-surface-subtle text-foreground-muted"
            aria-hidden="true"
          >
            <Icon name="key" size="md" />
          </span>
          <div class="min-w-0">
            <div class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
              <p class="min-w-0 break-words text-sm font-medium text-foreground">
                {{ device.device_name || t('profile.authorizedApps.unknownDevice') }}
              </p>
              <span class="min-w-0 break-all text-xs text-foreground-subtle">
                {{ device.client_name || device.client_id }}
              </span>
            </div>
            <p class="mt-1 break-words text-sm text-foreground-muted">
              {{ device.platform || t('profile.authorizedApps.unknownPlatform') }}
              <span aria-hidden="true"> · </span>
              {{ t('profile.authorizedApps.authorizedAt', { date: formatDateTime(device.created_at) }) }}
            </p>
            <p v-if="device.last_used_at" class="mt-1 break-words text-xs text-foreground-subtle">
              {{ t('profile.authorizedApps.lastUsedAt', { date: formatDateTime(device.last_used_at) }) }}
            </p>
          </div>
          <button
            type="button"
            class="btn btn-ghost btn-icon shrink-0 text-danger-foreground"
            :aria-label="t('profile.authorizedApps.revoke')"
            :title="t('profile.authorizedApps.revoke')"
            @click="deviceToRevoke = device"
          >
            <Icon name="trash" size="md" aria-hidden="true" />
          </button>
        </li>
      </ul>
    </div>

    <ConfirmDialog
      :show="deviceToRevoke !== null"
      :title="t('profile.authorizedApps.revokeTitle')"
      :message="t('profile.authorizedApps.revokeMessage', { device: deviceToRevoke?.device_name || t('profile.authorizedApps.unknownDevice') })"
      :confirm-text="t('profile.authorizedApps.revoke')"
      danger
      @confirm="revokeDevice"
      @cancel="deviceToRevoke = null"
    />
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { listAuthorizationDevices, revokeAuthorizationDevice, type AppAuthorizationDevice } from '@/api/appAuth'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { Icon } from '@/components/icons'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const loadError = ref('')
const devices = ref<AppAuthorizationDevice[]>([])
const deviceToRevoke = ref<AppAuthorizationDevice | null>(null)

async function loadDevices(): Promise<void> {
  loading.value = true
  loadError.value = ''
  try {
    devices.value = await listAuthorizationDevices()
  } catch (error) {
    loadError.value = extractApiErrorMessage(error, t('profile.authorizedApps.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function revokeDevice(): Promise<void> {
  const device = deviceToRevoke.value
  if (!device) return
  try {
    await revokeAuthorizationDevice(device.id)
    devices.value = devices.value.filter((item) => item.id !== device.id)
    deviceToRevoke.value = null
    appStore.showSuccess(t('profile.authorizedApps.revokeSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('profile.authorizedApps.revokeFailed')))
  }
}

onMounted(loadDevices)
</script>
