<template>
  <div class="card" data-testid="authorized-apps-card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">{{ t('profile.authorizedApps.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('profile.authorizedApps.description') }}</p>
    </div>

    <div v-if="loading" class="flex justify-center px-6 py-10">
      <span class="h-7 w-7 animate-spin rounded-full border-2 border-gray-200 border-t-primary-600" />
    </div>
    <div v-else-if="loadError" class="px-6 py-8 text-center">
      <p class="text-sm text-red-600 dark:text-red-400">{{ loadError }}</p>
      <button type="button" class="btn btn-secondary mt-4" @click="loadDevices">{{ t('profile.authorizedApps.retry') }}</button>
    </div>
    <div v-else-if="devices.length === 0" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('profile.authorizedApps.empty') }}
    </div>
    <ul v-else class="divide-y divide-gray-100 dark:divide-dark-700">
      <li v-for="device in devices" :key="device.id" class="flex items-start gap-4 px-6 py-5">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300">
          <Icon name="key" size="md" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
            <p class="font-medium text-gray-900 dark:text-white">{{ device.device_name || t('profile.authorizedApps.unknownDevice') }}</p>
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ device.client_name || device.client_id }}</span>
          </div>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ device.platform || t('profile.authorizedApps.unknownPlatform') }}
            <span aria-hidden="true"> · </span>
            {{ t('profile.authorizedApps.authorizedAt', { date: formatDateTime(device.created_at) }) }}
          </p>
          <p v-if="device.last_used_at" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('profile.authorizedApps.lastUsedAt', { date: formatDateTime(device.last_used_at) }) }}
          </p>
        </div>
        <button
          type="button"
          class="btn-icon shrink-0 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          :aria-label="t('profile.authorizedApps.revoke')"
          :title="t('profile.authorizedApps.revoke')"
          @click="deviceToRevoke = device"
        >
          <Icon name="trash" size="md" />
        </button>
      </li>
    </ul>

    <ConfirmDialog
      :show="deviceToRevoke !== null"
      :title="t('profile.authorizedApps.revokeTitle')"
      :message="t('profile.authorizedApps.revokeMessage', { device: deviceToRevoke?.device_name || t('profile.authorizedApps.unknownDevice') })"
      :confirm-text="t('profile.authorizedApps.revoke')"
      danger
      @confirm="revokeDevice"
      @cancel="deviceToRevoke = null"
    />
  </div>
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
