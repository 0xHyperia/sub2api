<template>
  <component
    :is="props.embedded ? 'div' : AuthLayout"
    :class="{ 'email-verification-step': props.embedded }"
  >
    <div class="space-y-6" :aria-busy="isLoading || isSendingCode">
      <!-- Title -->
      <div class="auth-form-heading">
        <h2 v-if="!props.embedded" class="text-2xl font-bold text-foreground">
          {{ t('auth.verifyYourEmail') }}
        </h2>
        <p class="verification-email-copy mt-2 text-sm text-foreground-muted">
          {{ t('auth.sendCodeDesc') }}
          <span class="font-medium text-foreground-muted">{{ email }}</span>
        </p>
      </div>

      <div
        v-if="errorMessage"
        class="auth-flow-alert"
        role="alert"
        aria-live="assertive"
      >
        {{ errorMessage }}
      </div>

      <!-- No Data Warning -->
      <div
        v-if="!hasRegisterData"
        class="verification-notice rounded-panel border bg-warning-subtle p-4 border-warning/30"
        role="alert"
      >
        <div class="flex items-start gap-3">
          <div class="flex-shrink-0">
            <Icon name="exclamationCircle" size="md" class="text-warning-foreground" />
          </div>
          <div class="text-sm text-warning-foreground">
            <p class="font-medium">{{ t('auth.sessionExpired') }}</p>
            <p class="mt-1">{{ t('auth.sessionExpiredDesc') }}</p>
          </div>
        </div>
      </div>

      <!-- Verification Form -->
      <form v-else class="space-y-5" novalidate @submit.prevent="handleVerify">
        <!-- Verification Code Input -->
        <div>
          <label for="code" class="input-label text-center">
            {{ t('auth.verificationCode') }}
          </label>
          <div
            class="verification-code-grid"
            role="group"
            :aria-label="t('auth.verificationCode')"
            :aria-describedby="errors.code ? 'email-verification-code-error' : undefined"
            @paste="handleVerificationPaste"
          >
            <input
              v-for="(_, index) in verificationDigits"
              :id="index === 0 ? 'code' : `code-${index + 1}`"
              ref="codeInputRefs"
              :key="index"
              :value="verificationDigits[index]"
              type="text"
              required
              :autocomplete="index === 0 ? 'one-time-code' : 'off'"
              inputmode="numeric"
              pattern="[0-9]*"
              maxlength="1"
              :autofocus="index === 0"
              :disabled="isLoading"
              class="verification-code-cell input"
              :class="{ 'input-error': errors.code }"
              :aria-label="`${t('auth.verificationCode')} ${index + 1}`"
              :aria-invalid="Boolean(errors.code)"
              @input="handleVerificationInput(index, $event)"
              @keydown="handleVerificationKeydown(index, $event)"
              @focus="($event.target as HTMLInputElement).select()"
            />
          </div>
          <p
            v-if="errors.code"
            id="email-verification-code-error"
            class="input-error-text text-center"
            role="alert"
          >
            {{ errors.code }}
          </p>
          <p class="input-hint text-center">{{ t('auth.verificationCodeHint') }}</p>
        </div>

        <!-- Code Status -->
        <div
          v-if="codeSent"
          class="verification-notice rounded-panel border p-4 border-success/30 bg-success-subtle"
          role="status"
          aria-live="polite"
        >
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <Icon name="checkCircle" size="md" class="text-success-foreground" />
            </div>
            <p class="text-sm text-success-foreground">
              {{ t('auth.codeSentSuccess') }}
            </p>
          </div>
        </div>

        <!-- Turnstile Widget for Resend -->
        <div
          v-if="actionCaptchaEnabled || (turnstileEnabled && showResendTurnstile)"
          role="group"
          :aria-describedby="errors.turnstile ? 'email-verification-turnstile-error' : undefined"
        >
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            :turnstile-enabled="turnstileEnabled"
            :turnstile-site-key="turnstileSiteKey"
            :tencent-enabled="tencentCaptchaEnabled"
            :tencent-app-id="tencentCaptchaAppId"
            :tencent-region="tencentCaptchaRegion"
            :aliyun-enabled="aliyunCaptchaEnabled"
            :aliyun-scene-id="aliyunCaptchaSceneId"
            :aliyun-prefix="aliyunCaptchaPrefix"
            :aliyun-region="aliyunCaptchaRegion"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p
            v-if="errors.turnstile"
            id="email-verification-turnstile-error"
            class="input-error-text"
            role="alert"
          >
            {{ errors.turnstile }}
          </p>
        </div>

        <div v-if="pendingOAuthCreateCaptchaEnabled" class="space-y-2">
          <TurnstileWidget
            ref="createAccountTurnstileRef"
            :site-key="turnstileSiteKey"
            :turnstile-enabled="turnstileEnabled"
            :turnstile-site-key="turnstileSiteKey"
            :tencent-enabled="tencentCaptchaEnabled"
            :tencent-app-id="tencentCaptchaAppId"
            :tencent-region="tencentCaptchaRegion"
            :aliyun-enabled="aliyunCaptchaEnabled"
            :aliyun-scene-id="aliyunCaptchaSceneId"
            :aliyun-prefix="aliyunCaptchaPrefix"
            :aliyun-region="aliyunCaptchaRegion"
            @verify="onCreateAccountTurnstileVerify"
            @expire="onCreateAccountTurnstileExpire"
            @error="onCreateAccountTurnstileError"
          />
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading || !verifyCode || (pendingOAuthCreateTurnstileRequired && !createAccountTurnstileToken)"
          class="auth-submit btn btn-primary w-full"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-inverse-foreground"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="checkCircle" size="md" class="mr-2" />
          {{ isLoading ? t('auth.verifying') : t('auth.verifyAndCreate') }}
        </button>

        <!-- Resend Code -->
        <div class="text-center">
          <button
            v-if="countdown > 0"
            type="button"
            disabled
            class="cursor-not-allowed text-sm text-foreground-subtle"
          >
            {{ t('auth.resendCountdown', { countdown }) }}
          </button>
          <button
            v-else
            type="button"
            @click="handleResendCode"
            :disabled="
              isSendingCode || (turnstileEnabled && showResendTurnstile && !resendTurnstileToken)
            "
            class="text-sm text-brand transition-colors hover:text-brand disabled:cursor-not-allowed disabled:opacity-50"
          >
            <span v-if="isSendingCode">{{ t('auth.sendingCode') }}</span>
            <span v-else-if="captchaEnabled && !showResendTurnstile">
              {{ t('auth.clickToResend') }}
            </span>
            <span v-else>{{ t('auth.resendCode') }}</span>
          </button>
        </div>
      </form>
    </div>

    <!-- Footer -->
    <template v-if="!props.embedded" #footer>
      <button
        type="button"
        @click="handleBack"
        class="flex items-center gap-2 transition-colors text-foreground-muted hover:text-foreground-muted"
      >
        <Icon name="arrowLeft" size="sm" />
        {{ t('auth.backToRegistration') }}
      </button>
    </template>

    <button
      v-if="props.embedded"
      type="button"
      class="email-verification-back"
      @click="handleBack"
    >
      <Icon name="arrowLeft" size="sm" />
      {{ t('auth.backToRegistration') }}
    </button>
  </component>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AuthLayout from '@/components/auth/AuthFlowLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/CaptchaChallenge.vue'
import { useAuthStore, useAppStore } from '@/stores'
import {
  persistOAuthTokenContext,
  getPublicSettings,
  isOAuthLoginCompletion,
  type PendingOAuthSendVerifyCodeResponse,
  sendPendingOAuthVerifyCode,
  sendVerifyCode,
} from '@/api/auth'
import { apiClient } from '@/api/client'
import { buildAuthErrorMessage } from '@/utils/authError'
import { extractApiErrorCode } from '@/utils/apiError'
import {
  formatRegistrationEmailSuffixWhitelistForMessage,
  isRegistrationEmailSuffixAllowed,
  normalizeRegistrationEmailSuffixWhitelist
} from '@/utils/registrationEmailPolicy'
import {
  clearAllAffiliateReferralCodes,
  loadAffiliateReferralCode,
  oauthAffiliatePayload
} from '@/utils/oauthAffiliate'

const { t, locale } = useI18n()
const props = withDefaults(defineProps<{
  embedded?: boolean
}>(), {
  embedded: false
})
const emit = defineEmits<{
  back: []
}>()

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const isSendingCode = ref<boolean>(false)
const errorMessage = ref<string>('')
const codeSent = ref<boolean>(false)
const verifyCode = ref<string>('')
const verificationDigits = ref<string[]>(Array.from({ length: 6 }, () => ''))
const codeInputRefs = ref<HTMLInputElement[]>([])
const countdown = ref<number>(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

// Registration data from sessionStorage
type PendingAuthTokenField = 'pending_auth_token' | 'pending_oauth_token'
type PendingAuthSessionSummary = {
  token: string
  token_field: PendingAuthTokenField
  provider: string
  redirect?: string
}
type PendingOAuthCreateAccountResponse = {
  auth_result?: string
  access_token: string
  refresh_token?: string
  expires_in?: number
  token_type?: string
  provider?: string
  redirect?: string
}

const email = ref<string>('')
const password = ref<string>('')
const initialTurnstileToken = ref<string>('')
const initialTencentCaptchaRandstr = ref<string>('')
const promoCode = ref<string>('')
const invitationCode = ref<string>('')
const affCode = ref<string>('')
const distributionCode = ref<string>('')
const pendingAuthToken = ref<string>('')
const pendingAuthTokenField = ref<PendingAuthTokenField>('pending_auth_token')
const pendingProvider = ref<string>('')
const pendingRedirect = ref<string>('')
const pendingAdoptionDecision = ref<{
  adoptDisplayName?: boolean
  adoptAvatar?: boolean
} | null>(null)
const hasRegisterData = ref<boolean>(false)

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const tencentCaptchaEnabled = ref<boolean>(false)
const tencentCaptchaAppId = ref<string>('')
const tencentCaptchaRegion = ref<string>('cn')
const aliyunCaptchaEnabled = ref<boolean>(false)
const aliyunCaptchaSceneId = ref<string>('')
const aliyunCaptchaPrefix = ref<string>('')
const aliyunCaptchaRegion = ref<string>('cn')
const siteName = ref<string>('Sub2API')
const registrationEmailSuffixWhitelist = ref<string[]>([])
// 域名限量注册开关：开启时非白名单域名可注册 1 个账户（由后端判定），前端不做白名单预检。
const emailDomainQuotaEnabled = ref<boolean>(false)

// Turnstile for resend
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const createAccountTurnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const resendTurnstileToken = ref<string>('')
const resendTencentCaptchaRandstr = ref<string>('')
const createAccountTurnstileToken = ref<string>('')
const createAccountTencentCaptchaRandstr = ref<string>('')
const showResendTurnstile = ref<boolean>(false)
const aliyunCaptchaReady = computed(
  () =>
    aliyunCaptchaEnabled.value &&
    Boolean(aliyunCaptchaSceneId.value) &&
    Boolean(aliyunCaptchaPrefix.value)
)
// 动作触发式验证码（腾讯/阿里云）：重发验证码、创建账号时弹窗验证
const actionCaptchaEnabled = computed(
  () =>
    (tencentCaptchaEnabled.value && Boolean(tencentCaptchaAppId.value)) ||
    aliyunCaptchaReady.value
)
const captchaEnabled = computed(
  () =>
    (turnstileEnabled.value && Boolean(turnstileSiteKey.value)) || actionCaptchaEnabled.value
)

const errors = ref({
  code: '',
  turnstile: ''
})

const validationToastMessage = computed(
  () => errors.value.code || errors.value.turnstile || ''
)
const pendingOAuthCreateTurnstileRequired = computed(
  () => isPendingOAuthFlow() && turnstileEnabled.value
)
const pendingOAuthCreateCaptchaEnabled = computed(
  () => isPendingOAuthFlow() && captchaEnabled.value
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

function updateVerificationCode(digits: string[]): void {
  verificationDigits.value = digits.slice(0, 6)
  verifyCode.value = verificationDigits.value.join('')
  if (verifyCode.value) errors.value.code = ''
}

function focusVerificationCell(index: number): void {
  const targetIndex = Math.max(0, Math.min(index, verificationDigits.value.length - 1))
  void nextTick(() => {
    codeInputRefs.value[targetIndex]?.focus()
    codeInputRefs.value[targetIndex]?.select()
  })
}

function applyVerificationCode(value: string): void {
  const normalized = value.replace(/\D/g, '').slice(0, 6)
  const digits = Array.from({ length: 6 }, (_, index) => normalized[index] || '')
  updateVerificationCode(digits)
  focusVerificationCell(Math.min(normalized.length, 6) - 1)
}

function handleVerificationInput(index: number, event: Event): void {
  const input = event.target as HTMLInputElement
  const normalized = input.value.replace(/\D/g, '')

  if (normalized.length > 1) {
    applyVerificationCode(normalized)
    return
  }

  const digits = [...verificationDigits.value]
  digits[index] = normalized.slice(-1)
  updateVerificationCode(digits)

  if (digits[index] && index < digits.length - 1) {
    focusVerificationCell(index + 1)
  }
}

function handleVerificationKeydown(index: number, event: KeyboardEvent): void {
  if (event.key === 'Backspace' && !verificationDigits.value[index] && index > 0) {
    event.preventDefault()
    const digits = [...verificationDigits.value]
    digits[index - 1] = ''
    updateVerificationCode(digits)
    focusVerificationCell(index - 1)
    return
  }

  if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    focusVerificationCell(index - 1)
  } else if (event.key === 'ArrowRight' && index < verificationDigits.value.length - 1) {
    event.preventDefault()
    focusVerificationCell(index + 1)
  }
}

function handleVerificationPaste(event: ClipboardEvent): void {
  const value = event.clipboardData?.getData('text') || ''
  if (!/\d/.test(value)) return

  event.preventDefault()
  applyVerificationCode(value)
}

// ==================== Lifecycle ====================

onMounted(async () => {
  const activePendingSession = authStore.pendingAuthSession as PendingAuthSessionSummary | null

  // Load registration data from sessionStorage
  const registerDataStr = sessionStorage.getItem('register_data')
  if (registerDataStr) {
    try {
      const registerData = JSON.parse(registerDataStr)
      email.value = registerData.email || ''
      password.value = registerData.password || ''
      initialTurnstileToken.value =
        registerData.tencent_captcha_ticket || registerData.turnstile_token || ''
      initialTencentCaptchaRandstr.value = registerData.tencent_captcha_randstr || ''
      promoCode.value = registerData.promo_code || ''
      invitationCode.value = registerData.invitation_code || ''
	      distributionCode.value = registerData.distribution_code || ''
	      affCode.value = distributionCode.value ? '' : (registerData.aff_code || loadAffiliateReferralCode())
      pendingAuthToken.value = registerData.pending_auth_token || activePendingSession?.token || ''
      pendingAuthTokenField.value = registerData.pending_auth_token_field || activePendingSession?.token_field || 'pending_auth_token'
      pendingProvider.value = registerData.pending_provider || activePendingSession?.provider || ''
      pendingRedirect.value = registerData.pending_redirect || activePendingSession?.redirect || ''
      pendingAdoptionDecision.value = registerData.pending_adoption_decision
        ? {
            adoptDisplayName: registerData.pending_adoption_decision.adopt_display_name === true,
            adoptAvatar: registerData.pending_adoption_decision.adopt_avatar === true
          }
        : null
      hasRegisterData.value = !!(email.value && password.value)
    } catch {
      hasRegisterData.value = false
    }
  } else if (activePendingSession) {
    pendingAuthToken.value = activePendingSession.token
    pendingAuthTokenField.value = activePendingSession.token_field
    pendingProvider.value = activePendingSession.provider
    pendingRedirect.value = activePendingSession.redirect || ''
  }

  // Load public settings
  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    tencentCaptchaEnabled.value = settings.tencent_captcha_enabled === true
    tencentCaptchaAppId.value = settings.tencent_captcha_app_id || ''
    tencentCaptchaRegion.value = settings.tencent_captcha_region || 'cn'
    aliyunCaptchaEnabled.value = settings.aliyun_captcha_enabled === true
    aliyunCaptchaSceneId.value = settings.aliyun_captcha_scene_id || ''
    aliyunCaptchaPrefix.value = settings.aliyun_captcha_prefix || ''
    aliyunCaptchaRegion.value = settings.aliyun_captcha_region || 'cn'
    siteName.value = settings.site_name || 'Sub2API'
    registrationEmailSuffixWhitelist.value = normalizeRegistrationEmailSuffixWhitelist(
      settings.registration_email_suffix_whitelist || []
    )
    emailDomainQuotaEnabled.value = settings.registration_email_domain_quota_enabled === true
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }

  // Auto-send verification code if we have valid data
  if (hasRegisterData.value) {
    await sendCode()
  }
})

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})

// ==================== Countdown ====================

function startCountdown(seconds: number): void {
  countdown.value = seconds

  if (countdownTimer) {
    clearInterval(countdownTimer)
  }

  countdownTimer = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    } else {
      if (countdownTimer) {
        clearInterval(countdownTimer)
        countdownTimer = null
      }
    }
  }, 1000)
}

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string, randstr = ''): void {
  resendTurnstileToken.value = token
  resendTencentCaptchaRandstr.value = randstr
  errors.value.turnstile = ''
}

function onTurnstileExpire(): void {
  resendTurnstileToken.value = ''
  resendTencentCaptchaRandstr.value = ''
  errors.value.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  resendTurnstileToken.value = ''
  resendTencentCaptchaRandstr.value = ''
  errors.value.turnstile = t('auth.turnstileFailed')
}

function onCreateAccountTurnstileVerify(token: string, randstr = ''): void {
  createAccountTurnstileToken.value = token
  createAccountTencentCaptchaRandstr.value = randstr
  errors.value.turnstile = ''
}

function onCreateAccountTurnstileExpire(): void {
  createAccountTurnstileToken.value = ''
  createAccountTencentCaptchaRandstr.value = ''
  errors.value.turnstile = t('auth.turnstileExpired')
}

function onCreateAccountTurnstileError(): void {
  createAccountTurnstileToken.value = ''
  createAccountTencentCaptchaRandstr.value = ''
  errors.value.turnstile = t('auth.turnstileFailed')
}

function resetCreateAccountTurnstile(): void {
  createAccountTurnstileToken.value = ''
  createAccountTencentCaptchaRandstr.value = ''
  createAccountTurnstileRef.value?.reset()
}

async function acquireResendActionProof(): Promise<boolean> {
  if (!actionCaptchaEnabled.value) return true

  const proof = await turnstileRef.value?.verifyAction()
  if (!proof) return false

  resendTurnstileToken.value = proof.token
  resendTencentCaptchaRandstr.value = proof.randstr
  return true
}

async function acquireCreateAccountActionProof(): Promise<boolean> {
  if (!isPendingOAuthFlow() || !actionCaptchaEnabled.value) return true

  const proof = await createAccountTurnstileRef.value?.verifyAction()
  if (!proof) return false

  createAccountTurnstileToken.value = proof.token
  createAccountTencentCaptchaRandstr.value = proof.randstr
  return true
}

function isPendingOAuthFlow(): boolean {
  return Boolean(pendingProvider.value.trim())
}

// 域名限量注册开启时交由后端按额度判定；pending OAuth / 换绑流程沿用后端策略，前端不预检。
function shouldBypassRegistrationEmailPolicy(): boolean {
  return (
    emailDomainQuotaEnabled.value || isPendingOAuthFlow() || Boolean(pendingAuthToken.value.trim())
  )
}

function resolvePendingOAuthCallbackRoute(provider: string): string {
  switch (provider.trim().toLowerCase()) {
    case 'linuxdo':
      return '/auth/linuxdo/callback'
    case 'oidc':
      return '/auth/oidc/callback'
    case 'wechat':
      return '/auth/wechat/callback'
    default:
      return '/auth/callback'
  }
}

function isPendingOAuthSessionResponse(data: PendingOAuthCreateAccountResponse): boolean {
  return data.auth_result === 'pending_session'
}

function getPendingOAuthSendCodeSessionResponse(
  data: PendingOAuthSendVerifyCodeResponse,
): PendingOAuthSendVerifyCodeResponse | null {
  return data.auth_result === 'pending_session' ? data : null
}

function persistPendingOAuthSession(provider: string, redirect?: string): void {
  authStore.setPendingAuthSession({
    token: pendingAuthToken.value,
    token_field: pendingAuthTokenField.value,
    provider: provider.trim() || pendingProvider.value.trim(),
    redirect: redirect || pendingRedirect.value || undefined,
  })
}

// ==================== Send Code ====================

async function sendCode(): Promise<void> {
  isSendingCode.value = true
  errorMessage.value = ''
  let requestSucceeded = false
  let captchaProofUsed = false

  try {
    if (!shouldBypassRegistrationEmailPolicy() && !isRegistrationEmailSuffixAllowed(email.value, registrationEmailSuffixWhitelist.value)) {
      errorMessage.value = buildEmailSuffixNotAllowedMessage()
      appStore.showError(errorMessage.value)
      return
    }

    const requestPayload = {
      email: email.value,
      [pendingAuthTokenField.value]: pendingAuthToken.value || undefined,
      // 优先使用重发时新获取的 token（因为初始 token 可能已被使用）
      turnstile_token:
        turnstileEnabled.value || aliyunCaptchaEnabled.value
          ? resendTurnstileToken.value || initialTurnstileToken.value || undefined
          : undefined,
      tencent_captcha_ticket: tencentCaptchaEnabled.value
        ? resendTurnstileToken.value || initialTurnstileToken.value || undefined
        : undefined,
      tencent_captcha_randstr: tencentCaptchaEnabled.value
        ? resendTencentCaptchaRandstr.value || initialTencentCaptchaRandstr.value || undefined
        : undefined
    } as Parameters<typeof sendVerifyCode>[0]
    captchaProofUsed = Boolean(
      requestPayload.turnstile_token || requestPayload.tencent_captcha_ticket
    )
    const response = isPendingOAuthFlow()
      ? await sendPendingOAuthVerifyCode(requestPayload)
      : await sendVerifyCode(requestPayload)
    requestSucceeded = true

    const pendingSendCodeSession = isPendingOAuthFlow()
      ? getPendingOAuthSendCodeSessionResponse(response as PendingOAuthSendVerifyCodeResponse)
      : null
    if (pendingSendCodeSession) {
      sessionStorage.removeItem('register_data')
      persistPendingOAuthSession(
        pendingSendCodeSession.provider || pendingProvider.value,
        pendingSendCodeSession.redirect,
      )
      await router.push(
        resolvePendingOAuthCallbackRoute(pendingSendCodeSession.provider || pendingProvider.value),
      )
      return
    }

    codeSent.value = true
    startCountdown(response.countdown)

    showResendTurnstile.value = false
  } catch (error: unknown) {
    errorMessage.value = buildRegistrationErrorMessage(error, t('auth.sendCodeFailed'))

    appStore.showError(errorMessage.value)
  } finally {
    if (captchaProofUsed) {
      clearStoredCaptchaProof()
      initialTurnstileToken.value = ''
      initialTencentCaptchaRandstr.value = ''
      resendTurnstileToken.value = ''
      resendTencentCaptchaRandstr.value = ''
      turnstileRef.value?.reset()
      if (!requestSucceeded && turnstileEnabled.value) {
        showResendTurnstile.value = true
      }
    }
    isSendingCode.value = false
  }
}

function clearStoredCaptchaProof(): void {
  const registerDataStr = sessionStorage.getItem('register_data')
  if (!registerDataStr) return

  try {
    const registerData = JSON.parse(registerDataStr) as Record<string, unknown>
    delete registerData.turnstile_token
    delete registerData.tencent_captcha_ticket
    delete registerData.tencent_captcha_randstr
    sessionStorage.setItem('register_data', JSON.stringify(registerData))
  } catch {
    // Invalid registration state is handled by the existing onMounted parser.
  }
}

// ==================== Handlers ====================

async function handleResendCode(): Promise<void> {
  // Turnstile stays staged; Tencent is acquired from this action.
  if (turnstileEnabled.value && !showResendTurnstile.value) {
    showResendTurnstile.value = true
    return
  }

  if (turnstileEnabled.value && !resendTurnstileToken.value) {
    errors.value.turnstile = t('auth.completeVerification')
    return
  }

  if (!(await acquireResendActionProof())) {
    return
  }

  await sendCode()
}

function validateForm(): boolean {
  errors.value.code = ''

  if (!verifyCode.value.trim()) {
    errors.value.code = t('auth.codeRequired')
    return false
  }

  if (!/^\d{6}$/.test(verifyCode.value.trim())) {
    errors.value.code = t('auth.invalidCode')
    return false
  }

  return true
}

async function handleVerify(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  if (!shouldBypassRegistrationEmailPolicy() && !isRegistrationEmailSuffixAllowed(email.value, registrationEmailSuffixWhitelist.value)) {
    errorMessage.value = buildEmailSuffixNotAllowedMessage()
    appStore.showError(errorMessage.value)
    return
  }

  if (!(await acquireCreateAccountActionProof())) {
    return
  }

  isLoading.value = true

  try {
    if (isPendingOAuthFlow()) {
      const payload: Record<string, unknown> = {
        email: email.value,
        password: password.value,
        verify_code: verifyCode.value.trim(),
        ...((turnstileEnabled.value || aliyunCaptchaEnabled.value) &&
        createAccountTurnstileToken.value
          ? { turnstile_token: createAccountTurnstileToken.value }
          : {}),
        ...(tencentCaptchaEnabled.value && createAccountTurnstileToken.value
          ? {
              tencent_captcha_ticket: createAccountTurnstileToken.value,
              tencent_captcha_randstr: createAccountTencentCaptchaRandstr.value
          }
          : {}),
        ...oauthAffiliatePayload(affCode.value || loadAffiliateReferralCode()),
      }
      if (invitationCode.value) {
        payload.invitation_code = invitationCode.value
      }
      if (pendingAdoptionDecision.value?.adoptDisplayName !== undefined) {
        payload.adopt_display_name = pendingAdoptionDecision.value.adoptDisplayName
      }
      if (pendingAdoptionDecision.value?.adoptAvatar !== undefined) {
        payload.adopt_avatar = pendingAdoptionDecision.value.adoptAvatar
      }

      const { data } = await apiClient.post<PendingOAuthCreateAccountResponse>(
        '/auth/oauth/pending/create-account',
        payload
      )
      if (isPendingOAuthSessionResponse(data)) {
        sessionStorage.removeItem('register_data')
        persistPendingOAuthSession(data.provider || pendingProvider.value, data.redirect)
        await router.push(resolvePendingOAuthCallbackRoute(data.provider || pendingProvider.value))
        return
      }
      if (!isOAuthLoginCompletion(data)) {
        throw new Error(t('auth.verifyFailed'))
      }

      persistOAuthTokenContext(data)
      await authStore.setToken(data.access_token)
      authStore.clearPendingAuthSession?.()
    } else {
      // Register with verification code
      await authStore.register({
        email: email.value,
        password: password.value,
        verify_code: verifyCode.value.trim(),
        turnstile_token:
          turnstileEnabled.value || aliyunCaptchaEnabled.value
            ? initialTurnstileToken.value || undefined
            : undefined,
        tencent_captcha_ticket: tencentCaptchaEnabled.value ? initialTurnstileToken.value || undefined : undefined,
        tencent_captcha_randstr: tencentCaptchaEnabled.value ? initialTencentCaptchaRandstr.value || undefined : undefined,
        promo_code: promoCode.value || undefined,
        invitation_code: invitationCode.value || undefined,
	        ...(affCode.value && !distributionCode.value ? { aff_code: affCode.value } : {}),
	        ...(distributionCode.value ? { distribution_code: distributionCode.value } : {})
      })
    }

    // Clear session data
    sessionStorage.removeItem('register_data')
    clearAllAffiliateReferralCodes()

    // Show success toast
    appStore.showSuccess(t('auth.accountCreatedSuccess', { siteName: siteName.value }))

    // Redirect to dashboard
    await router.push(pendingRedirect.value || '/dashboard')
  } catch (error: unknown) {
    errorMessage.value = buildRegistrationErrorMessage(error, t('auth.verifyFailed'))

    appStore.showError(errorMessage.value)
  } finally {
    initialTurnstileToken.value = ''
    initialTencentCaptchaRandstr.value = ''
    if (pendingOAuthCreateCaptchaEnabled.value) {
      resetCreateAccountTurnstile()
    }
    isLoading.value = false
  }
}

function handleBack(): void {
  // Clear session data
  sessionStorage.removeItem('register_data')

  if (props.embedded) {
    emit('back')
    return
  }

  void router.push('/register')
}

function buildEmailSuffixNotAllowedMessage(): string {
  const normalizedWhitelist = normalizeRegistrationEmailSuffixWhitelist(
    registrationEmailSuffixWhitelist.value
  )
  if (normalizedWhitelist.length === 0) {
    return t('auth.emailSuffixNotAllowed')
  }
  const separator = String(locale.value || '').toLowerCase().startsWith('zh') ? '、' : ', '
  return t('auth.emailSuffixNotAllowedWithAllowed', {
    suffixes: formatRegistrationEmailSuffixWhitelistForMessage(normalizedWhitelist, {
      separator,
      more: (count) => t('auth.emailSuffixAllowedMore', { count })
    })
  })
}

function buildRegistrationErrorMessage(error: unknown, fallback: string): string {
  if (extractApiErrorCode(error) === 'EMAIL_DOMAIN_REGISTRATION_LIMIT') {
    return t('auth.emailDomainRegistrationLimit')
  }
  return buildAuthErrorMessage(error, { fallback })
}
</script>

<style scoped>
.email-verification-step {
  width: 100%;
}

.email-verification-step .verification-email-copy {
  margin: 0;
  color: var(--muted-foreground);
  line-height: 1.7;
}

.email-verification-step .verification-email-copy span {
  display: inline;
  margin-left: 4px;
  color: var(--foreground);
  font-weight: 750;
  overflow-wrap: anywhere;
}

.email-verification-step .verification-notice {
  border-radius: 8px;
}

.verification-code-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 8px;
  margin-top: 7px;
}

.verification-code-cell.input {
  width: 100%;
  min-width: 0;
  height: 50px;
  padding: 0;
  border-radius: 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 20px;
  font-weight: 750;
  line-height: 1;
  text-align: center;
}

.email-verification-back {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  margin: 20px auto 0;
  border: 0;
  background: transparent;
  color: var(--muted-foreground);
  cursor: pointer;
  font-size: 13px;
  font-weight: 700;
}

.email-verification-back:hover {
  color: var(--foreground);
}

.email-verification-back:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--foreground) 34%, transparent);
  outline-offset: 4px;
}

@media (max-width: 520px) {
  .verification-code-grid {
    gap: 6px;
  }

  .verification-code-cell.input {
    height: 46px;
    font-size: 18px;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
