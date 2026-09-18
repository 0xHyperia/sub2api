<template>
  <BaseDialog
    :show="true"
    :title="t('profile.totp.disableTitle')"
    width="narrow"
    :close-on-click-outside="true"
    @close="emit('close')"
  >
    <div class="mb-5 flex items-start gap-3 rounded-panel border border-danger/20 bg-danger-subtle p-3 text-danger-foreground" role="note">
      <Icon name="exclamationTriangle" size="md" class="mt-0.5 shrink-0" aria-hidden="true" />
      <p class="text-sm leading-5">{{ t('profile.totp.disableWarning') }}</p>
    </div>

    <div v-if="methodLoading" class="flex items-center justify-center py-8" role="status" :aria-label="t('common.loading')">
      <span class="h-7 w-7 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true" />
    </div>

    <form v-else class="space-y-4" @submit.prevent="handleDisable">
      <div v-if="verificationMethod === 'email'">
        <label for="totp-disable-email-code" class="input-label">{{ t('profile.totp.emailCode') }}</label>
        <div class="flex flex-col gap-2 sm:flex-row">
          <input
            id="totp-disable-email-code"
            v-model="form.emailCode"
            type="text"
            maxlength="6"
            inputmode="numeric"
            class="input min-w-0 flex-1"
            :placeholder="t('profile.totp.enterEmailCode')"
          />
          <button
            type="button"
            class="btn btn-secondary whitespace-nowrap"
            :disabled="sendingCode || codeCooldown > 0"
            @click="handleSendCode"
          >
            {{ codeCooldown > 0 ? `${codeCooldown}s` : (sendingCode ? t('common.sending') : t('profile.totp.sendCode')) }}
          </button>
        </div>
      </div>

      <div v-else>
        <label for="totp-disable-password" class="input-label">
          {{ t('profile.currentPassword') }}
        </label>
        <input
          id="totp-disable-password"
          v-model="form.password"
          type="password"
          autocomplete="current-password"
          class="input"
          :placeholder="t('profile.totp.enterPassword')"
        />
      </div>

      <div class="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" class="btn btn-danger" :disabled="loading || !canSubmit">
          {{ loading ? t('common.processing') : t('profile.totp.confirmDisable') }}
        </button>
      </div>
    </form>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { totpAPI } from '@/api'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const methodLoading = ref(true)
const verificationMethod = ref<'email' | 'password'>('password')
const loading = ref(false)
const sendingCode = ref(false)
const codeCooldown = ref(0)
const cooldownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const form = ref({
  emailCode: '',
  password: ''
})

const canSubmit = computed(() => {
  if (verificationMethod.value === 'email') {
    return form.value.emailCode.length === 6
  }
  return form.value.password.length > 0
})

const loadVerificationMethod = async () => {
  methodLoading.value = true
  try {
    const method = await totpAPI.getVerificationMethod()
    verificationMethod.value = method.method
  } catch (err: any) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
    emit('close')
  } finally {
    methodLoading.value = false
  }
}

const handleSendCode = async () => {
  sendingCode.value = true
  try {
    await totpAPI.sendVerifyCode()
    appStore.showSuccess(t('profile.totp.codeSent'))
    // Start cooldown
    codeCooldown.value = 60
    if (cooldownTimer.value) {
      clearInterval(cooldownTimer.value)
      cooldownTimer.value = null
    }
    cooldownTimer.value = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) {
        if (cooldownTimer.value) {
          clearInterval(cooldownTimer.value)
          cooldownTimer.value = null
        }
      }
    }, 1000)
  } catch (err: any) {
    appStore.showError(extractApiErrorMessage(err, t('profile.totp.sendCodeFailed')))
  } finally {
    sendingCode.value = false
  }
}

const handleDisable = async () => {
  if (!canSubmit.value) return

  loading.value = true

  try {
    const request = verificationMethod.value === 'email'
      ? { email_code: form.value.emailCode }
      : { password: form.value.password }

    await totpAPI.disable(request)
    appStore.showSuccess(t('profile.totp.disableSuccess'))
    emit('success')
  } catch (err: any) {
    appStore.showError(extractApiErrorMessage(err, t('profile.totp.disableFailed')))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadVerificationMethod()
})

onUnmounted(() => {
  if (cooldownTimer.value) {
    clearInterval(cooldownTimer.value)
    cooldownTimer.value = null
  }
})
</script>
