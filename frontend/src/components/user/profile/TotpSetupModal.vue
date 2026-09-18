<template>
  <BaseDialog
    :show="true"
    :title="t('profile.totp.setupTitle')"
    width="narrow"
    :close-on-click-outside="true"
    @close="emit('close')"
  >
    <p class="mb-5 border-b border-outline pb-4 text-sm text-foreground-muted">
      {{ stepDescription }}
    </p>

    <div ref="stepContent">

        <!-- Step 0: Identity Verification -->
        <div v-if="step === 0" class="space-y-6">
          <!-- Loading verification method -->
          <div v-if="methodLoading" class="flex items-center justify-center py-8" role="status" :aria-label="t('common.loading')">
            <span class="h-7 w-7 animate-spin rounded-full border-2 border-outline-strong border-t-foreground" aria-hidden="true" />
          </div>

          <template v-else>
            <!-- Email verification -->
            <div v-if="verificationMethod === 'email'" class="space-y-4">
              <div>
                <label for="totp-setup-email-code" class="input-label">{{ t('profile.totp.emailCode') }}</label>
                <div class="flex flex-col gap-2 sm:flex-row">
                  <input
                    id="totp-setup-email-code"
                    v-model="verifyForm.emailCode"
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
            </div>

            <!-- Password verification -->
            <div v-else class="space-y-4">
              <div>
                <label for="totp-setup-password" class="input-label">{{ t('profile.currentPassword') }}</label>
                <input
                  id="totp-setup-password"
                  v-model="verifyForm.password"
                  type="password"
                  autocomplete="current-password"
                  class="input"
                  :placeholder="t('profile.totp.enterPassword')"
                />
              </div>
            </div>

            <div class="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
              <button type="button" class="btn btn-secondary" @click="$emit('close')">
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn btn-primary"
                :disabled="!canProceedFromVerify || setupLoading"
                @click="handleVerifyAndSetup"
              >
                {{ setupLoading ? t('common.loading') : t('common.next') }}
              </button>
            </div>
          </template>
        </div>

        <!-- Step 1: Show QR Code -->
        <div v-if="step === 1" class="space-y-6">
          <!-- QR Code and Secret -->
          <template v-if="setupData">
            <div class="flex justify-center">
              <div class="rounded-panel border border-outline bg-white p-3">
                <img :src="qrCodeDataUrl" :alt="t('profile.totp.setupStep1')" class="h-48 w-48" />
              </div>
            </div>

            <div class="text-center">
              <p class="mb-2 text-sm text-foreground-muted">
                {{ t('profile.totp.manualEntry') }}
              </p>
              <div class="flex min-w-0 items-center justify-center gap-2">
                <code class="min-w-0 break-all rounded-control border border-outline bg-surface-subtle px-3 py-2 font-mono text-sm text-foreground">
                  {{ setupData.secret }}
                </code>
                <button
                  type="button"
                  class="btn btn-ghost btn-icon shrink-0"
                  :aria-label="t('common.copy')"
                  :title="t('common.copy')"
                  @click="copySecret"
                >
                  <Icon name="copy" size="md" aria-hidden="true" />
                </button>
              </div>
            </div>
          </template>

          <div class="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:justify-end">
            <button type="button" class="btn btn-secondary" @click="$emit('close')">
              {{ t('common.cancel') }}
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="!setupData"
              @click="step = 2"
            >
              {{ t('common.next') }}
            </button>
          </div>
        </div>

        <!-- Step 2: Verify Code -->
        <div v-if="step === 2" class="space-y-6">
          <form @submit.prevent="handleVerify">
            <div class="mb-6">
              <p id="totp-code-label" class="mb-3 text-center text-sm font-medium text-foreground">
                {{ t('profile.totp.enterCode') }}
              </p>
              <div class="flex justify-center gap-1.5 sm:gap-2" role="group" aria-labelledby="totp-code-label">
                <input
                  v-for="(_, index) in 6"
                  :key="index"
                  :ref="(el) => setInputRef(el, index)"
                  :value="code[index]"
                  type="text"
                  maxlength="1"
                  inputmode="numeric"
                  pattern="[0-9]"
                  autocomplete="one-time-code"
                  class="h-11 w-9 rounded-control border border-outline-strong bg-surface text-center text-lg font-semibold text-foreground outline-none focus:border-focus focus:ring-2 focus:ring-focus/20 sm:w-10"
                  :aria-label="`${t('profile.totp.enterCode')} ${index + 1}`"
                  @input="handleCodeInput($event, index)"
                  @keydown="handleKeydown($event, index)"
                  @paste="handlePaste"
                />
              </div>
            </div>

            <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <button type="button" class="btn btn-secondary" @click="step = 1">
                {{ t('common.back') }}
              </button>
              <button
                type="submit"
                class="btn btn-primary"
                :disabled="verifying || code.join('').length !== 6"
              >
                {{ verifying ? t('common.verifying') : t('profile.totp.verify') }}
              </button>
            </div>
          </form>
        </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { totpAPI } from '@/api'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { TotpSetupResponse } from '@/types'
import QRCode from 'qrcode'

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()
const stepContent = ref<HTMLElement | null>(null)

// Step: 0 = verify identity, 1 = QR code, 2 = verify TOTP code
const step = ref(0)
const methodLoading = ref(true)
const verificationMethod = ref<'email' | 'password'>('password')
const verifyForm = ref({ emailCode: '', password: '' })
const sendingCode = ref(false)
const codeCooldown = ref(0)
const cooldownTimer = ref<ReturnType<typeof setInterval> | null>(null)

const setupLoading = ref(false)
const setupData = ref<TotpSetupResponse | null>(null)
const verifying = ref(false)
const code = ref<string[]>(['', '', '', '', '', ''])
const inputRefs = ref<(HTMLInputElement | null)[]>([])
const qrCodeDataUrl = ref('')

const stepDescription = computed(() => {
  switch (step.value) {
    case 0:
      return verificationMethod.value === 'email'
        ? t('profile.totp.verifyEmailFirst')
        : t('profile.totp.verifyPasswordFirst')
    case 1:
      return t('profile.totp.setupStep1')
    case 2:
      return t('profile.totp.setupStep2')
    default:
      return ''
  }
})

const canProceedFromVerify = computed(() => {
  if (verificationMethod.value === 'email') {
    return verifyForm.value.emailCode.length === 6
  }
  return verifyForm.value.password.length > 0
})

// Generate QR code as base64 when setupData changes
watch(
  () => setupData.value?.qr_code_url,
  async (url) => {
    if (url) {
      try {
        qrCodeDataUrl.value = await QRCode.toDataURL(url, {
          width: 200,
          margin: 2,
          color: {
            dark: '#000000',
            light: '#ffffff'
          }
        })
      } catch (err) {
        console.error('Failed to generate QR code:', err)
      }
    }
  },
  { immediate: true }
)

watch(step, async () => {
  await nextTick()
  const firstFocusable = stepContent.value?.querySelector<HTMLElement>(
    'button:not([disabled]), input:not([disabled]), [href], [tabindex]:not([tabindex="-1"])'
  )
  firstFocusable?.focus()
})

const setInputRef = (el: any, index: number) => {
  inputRefs.value[index] = el as HTMLInputElement | null
}

const handleCodeInput = (event: Event, index: number) => {
  const input = event.target as HTMLInputElement
  const value = input.value.replace(/[^0-9]/g, '')
  code.value[index] = value

  if (value && index < 5) {
    nextTick(() => {
      inputRefs.value[index + 1]?.focus()
    })
  }
}

const handleKeydown = (event: KeyboardEvent, index: number) => {
  if (event.key === 'Backspace') {
    const input = event.target as HTMLInputElement
    // If current cell is empty and not the first, move to previous cell
    if (!input.value && index > 0) {
      event.preventDefault()
      inputRefs.value[index - 1]?.focus()
    }
    // Otherwise, let the browser handle the backspace naturally
    // The input event will sync code.value via handleCodeInput
  }
}

const handlePaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const pastedData = event.clipboardData?.getData('text') || ''
  const digits = pastedData.replace(/[^0-9]/g, '').slice(0, 6).split('')

  // Update both the ref and the input elements
  digits.forEach((digit, index) => {
    code.value[index] = digit
    if (inputRefs.value[index]) {
      inputRefs.value[index]!.value = digit
    }
  })

  // Clear remaining inputs if pasted less than 6 digits
  for (let i = digits.length; i < 6; i++) {
    code.value[i] = ''
    if (inputRefs.value[i]) {
      inputRefs.value[i]!.value = ''
    }
  }

  const focusIndex = Math.min(digits.length, 5)
  nextTick(() => {
    inputRefs.value[focusIndex]?.focus()
  })
}

const copySecret = async () => {
  if (setupData.value) {
    try {
      await navigator.clipboard.writeText(setupData.value.secret)
      appStore.showSuccess(t('common.copied'))
    } catch {
      appStore.showError(t('common.copyFailed'))
    }
  }
}

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

const handleVerifyAndSetup = async () => {
  setupLoading.value = true

  try {
    const request = verificationMethod.value === 'email'
      ? { email_code: verifyForm.value.emailCode }
      : { password: verifyForm.value.password }

    setupData.value = await totpAPI.initiateSetup(request)
    step.value = 1
  } catch (err: any) {
    appStore.showError(extractApiErrorMessage(err, t('profile.totp.setupFailed')))
  } finally {
    setupLoading.value = false
  }
}

const handleVerify = async () => {
  const totpCode = code.value.join('')
  if (totpCode.length !== 6 || !setupData.value) return

  verifying.value = true

  try {
    await totpAPI.enable({
      totp_code: totpCode,
      setup_token: setupData.value.setup_token
    })
    appStore.showSuccess(t('profile.totp.enableSuccess'))
    emit('success')
  } catch (err: any) {
    appStore.showError(extractApiErrorMessage(err, t('profile.totp.verifyFailed')))
    code.value = ['', '', '', '', '', '']
    nextTick(() => {
      inputRefs.value[0]?.focus()
    })
  } finally {
    verifying.value = false
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
