<template>
  <div
    ref="dialogRef"
    class="fixed inset-0 z-50 overflow-y-auto p-2 sm:p-4"
    role="dialog"
    aria-modal="true"
    aria-labelledby="totp-login-title"
    aria-describedby="totp-login-description"
    tabindex="-1"
    @keydown="handleDialogKeydown"
  >
    <div class="fixed inset-0 bg-black/55 backdrop-blur-sm" aria-hidden="true"></div>

    <div class="relative flex min-h-full items-center justify-center">
      <div class="relative w-full max-w-md rounded-panel border border-outline p-3 shadow-modal bg-surface sm:p-6">
        <div class="mb-6 text-center">
          <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-panel bg-brand-subtle">
            <svg class="h-6 w-6 text-brand" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
            </svg>
          </div>
          <h2 id="totp-login-title" class="mt-4 text-xl font-semibold text-foreground">
            {{ t('profile.totp.loginTitle') }}
          </h2>
          <p id="totp-login-description" class="mt-2 text-sm text-foreground-subtle">
            {{ t('profile.totp.loginHint') }}
          </p>
          <p v-if="userEmailMasked" class="mt-1 text-sm font-medium text-foreground-muted">
            {{ userEmailMasked }}
          </p>
        </div>

        <div class="mb-6">
          <input
            ref="hiddenOtpInputRef"
            type="text"
            inputmode="numeric"
            autocomplete="one-time-code"
            maxlength="6"
            class="pointer-events-none absolute left-0 top-0 h-px w-px opacity-0"
            aria-hidden="true"
            tabindex="-1"
            @input="handleHiddenOtpInput"
          />
          <div class="grid grid-cols-6 gap-2" role="group" :aria-label="t('profile.totp.loginTitle')">
            <input
              v-for="(_, index) in 6"
              :key="index"
              :ref="(el) => setInputRef(el, index)"
              type="text"
              maxlength="1"
              inputmode="numeric"
              pattern="[0-9]"
              autocomplete="off"
              class="h-12 min-w-0 w-full rounded-panel border border-outline-strong text-center text-lg font-semibold focus:border-focus focus:ring-focus bg-surface-subtle"
              :aria-label="`${t('profile.totp.loginTitle')} ${index + 1}/6`"
              :disabled="verifying"
              @input="handleCodeInput($event, index)"
              @keydown="handleKeydown($event, index)"
              @paste="handlePaste"
            />
          </div>
          <div v-if="verifying" class="mt-3 flex items-center justify-center gap-2 text-sm text-foreground-subtle" role="status" aria-live="polite">
            <div class="h-4 w-4 animate-spin rounded-full border-2 border-outline-strong border-b-primary-500" aria-hidden="true"></div>
            {{ t('common.verifying') }}
          </div>
        </div>

        <button
          type="button"
          class="btn btn-secondary min-h-11 w-full"
          :disabled="verifying"
          @click="cancelDialog"
        >
          {{ t('common.cancel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

defineProps<{
  tempToken: string
  userEmailMasked?: string
}>()

const emit = defineEmits<{
  verify: [code: string]
  cancel: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const verifying = ref(false)
const code = ref<string[]>(['', '', '', '', '', ''])
const inputRefs = ref<(HTMLInputElement | null)[]>([])
const hiddenOtpInputRef = ref<HTMLInputElement | null>(null)
const dialogRef = ref<HTMLElement | null>(null)
let previouslyFocusedElement: HTMLElement | null = null
let previousBodyOverflow = ''

watch(
  () => code.value.join(''),
  (newCode) => {
    if (newCode.length === 6 && !verifying.value) emit('verify', newCode)
  }
)

defineExpose({
  setVerifying: (value: boolean) => { verifying.value = value },
  setError: (message: string) => {
    if (message) appStore.showError(message)
    code.value = ['', '', '', '', '', '']
    inputRefs.value.forEach((input) => {
      if (input) input.value = ''
    })
    if (hiddenOtpInputRef.value) hiddenOtpInputRef.value.value = ''
    void nextTick(() => inputRefs.value[0]?.focus())
  }
})

function setInputRef(el: unknown, index: number): void {
  inputRefs.value[index] = el as HTMLInputElement | null
}

function handleCodeInput(event: Event, index: number): void {
  const input = event.target as HTMLInputElement
  const value = input.value.replace(/[^0-9]/g, '')
  input.value = value
  code.value[index] = value

  if (value && index < 5) void nextTick(() => inputRefs.value[index + 1]?.focus())
}

function handleHiddenOtpInput(event: Event): void {
  const input = event.target as HTMLInputElement
  applyDigits(input.value.replace(/[^0-9]/g, '').slice(0, 6).split(''))
}

function handleKeydown(event: KeyboardEvent, index: number): void {
  if (event.key !== 'Backspace') return
  const input = event.target as HTMLInputElement
  if (!input.value && index > 0) {
    event.preventDefault()
    inputRefs.value[index - 1]?.focus()
  }
}

function handlePaste(event: ClipboardEvent): void {
  event.preventDefault()
  const digits = (event.clipboardData?.getData('text') || '').replace(/[^0-9]/g, '').slice(0, 6).split('')
  applyDigits(digits)
  const focusIndex = Math.min(digits.length, 5)
  void nextTick(() => inputRefs.value[focusIndex]?.focus())
}

function applyDigits(digits: string[]): void {
  for (let index = 0; index < 6; index += 1) {
    const digit = digits[index] || ''
    code.value[index] = digit
    if (inputRefs.value[index]) inputRefs.value[index]!.value = digit
  }
}

function cancelDialog(): void {
  if (!verifying.value) emit('cancel')
}

function getFocusableElements(): HTMLElement[] {
  if (!dialogRef.value) return []
  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'button:not([disabled]), input:not([disabled]):not([tabindex="-1"]), [href], [tabindex]:not([tabindex="-1"])'
    )
  ).filter((element) => element.getAttribute('aria-hidden') !== 'true')
}

function handleDialogKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    cancelDialog()
    return
  }
  if (event.key !== 'Tab') return

  const focusableElements = getFocusableElements()
  if (!focusableElements.length) {
    event.preventDefault()
    dialogRef.value?.focus()
    return
  }

  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]
  if (event.shiftKey && document.activeElement === firstElement) {
    event.preventDefault()
    lastElement.focus()
  } else if (!event.shiftKey && document.activeElement === lastElement) {
    event.preventDefault()
    firstElement.focus()
  }
}

onMounted(() => {
  previouslyFocusedElement = document.activeElement as HTMLElement | null
  previousBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
  void nextTick(() => inputRefs.value[0]?.focus())
})

onBeforeUnmount(() => {
  document.body.style.overflow = previousBodyOverflow
  previouslyFocusedElement?.focus()
})
</script>
