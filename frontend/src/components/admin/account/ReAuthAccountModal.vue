<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.reAuthorizeAccount')"
    width="normal"
    @close="handleClose"
  >
    <div
      v-if="account"
      class="min-w-0 space-y-4"
      :aria-busy="currentLoading"
    >
      <section
        data-testid="admin-reauth-account-summary"
        class="min-w-0 rounded-panel border border-outline bg-surface-subtle p-3"
        aria-labelledby="admin-reauth-account-name"
        aria-describedby="admin-reauth-account-platform"
      >
        <div class="flex min-w-0 items-center gap-3">
          <div
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-panel border border-outline bg-surface text-foreground-muted"
            aria-hidden="true"
          >
            <Icon name="sparkles" size="md" />
          </div>
          <div class="min-w-0">
            <h4 id="admin-reauth-account-name" class="break-words text-sm font-semibold text-foreground">
              {{ account.name }}
            </h4>
            <p id="admin-reauth-account-platform" class="mt-0.5 text-xs text-foreground-subtle">
              {{ platformLabel }}
            </p>
          </div>
        </div>
      </section>

      <fieldset v-if="isAnthropic" class="min-w-0 border-0 p-0" :disabled="currentLoading">
        <legend class="input-label">{{ t('admin.accounts.oauth.authMethod') }}</legend>
        <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <label
            :class="[
              'flex min-h-touch cursor-pointer items-center gap-2 rounded-control border px-3 py-2 text-sm text-foreground transition-colors',
              addMethod === 'oauth'
                ? 'border-outline-strong bg-surface-subtle'
                : 'border-outline bg-surface hover:bg-surface-subtle'
            ]"
          >
            <input
              v-model="addMethod"
              type="radio"
              name="admin-reauthorization-method"
              value="oauth"
              class="h-4 w-4 shrink-0 text-brand focus:ring-focus"
            />
            <span>{{ t('admin.accounts.types.oauth') }}</span>
          </label>
          <label
            :class="[
              'flex min-h-touch cursor-pointer items-center gap-2 rounded-control border px-3 py-2 text-sm text-foreground transition-colors',
              addMethod === 'setup-token'
                ? 'border-outline-strong bg-surface-subtle'
                : 'border-outline bg-surface hover:bg-surface-subtle'
            ]"
          >
            <input
              v-model="addMethod"
              type="radio"
              name="admin-reauthorization-method"
              value="setup-token"
              class="h-4 w-4 shrink-0 text-brand focus:ring-focus"
            />
            <span>{{ t('admin.accounts.setupTokenLongLived') }}</span>
          </label>
        </div>
      </fieldset>

      <section
        v-if="isGemini"
        class="min-w-0 border-t border-outline pt-3"
        aria-labelledby="admin-reauth-gemini-oauth-type-label"
        aria-describedby="admin-reauth-gemini-oauth-type-description"
      >
        <dl>
          <dt id="admin-reauth-gemini-oauth-type-label" class="text-xs font-medium text-foreground-muted">
            {{ t('admin.accounts.oauth.gemini.oauthTypeLabel') }}
          </dt>
          <dd class="mt-1 min-w-0">
            <span class="block break-words text-sm font-medium text-foreground">
              {{ geminiOAuthTypeLabel }}
            </span>
            <p
              id="admin-reauth-gemini-oauth-type-description"
              class="mt-0.5 break-words text-xs leading-5 text-foreground-subtle"
            >
              {{ geminiOAuthTypeDescription }}
            </p>
          </dd>
        </dl>
      </section>

      <div
        class="min-w-0"
        :aria-describedby="currentError ? 'admin-reauthorization-error' : undefined"
      >
        <OAuthAuthorizationFlow
          ref="oauthFlowRef"
          :add-method="addMethod"
          :auth-url="currentAuthUrl"
          :session-id="currentSessionId"
          :loading="currentLoading"
          :error="currentError"
          :show-help="isAnthropic"
          :show-proxy-warning="isAnthropic"
          :show-cookie-option="isAnthropic"
          :show-refresh-token-option="isOpenAI || isAntigravity || isGrok"
          :show-sso-option="isGrok"
          :show-email-password-option="false"
          :allow-multiple="false"
          :method-label="t('admin.accounts.inputMethod')"
          :platform="isOpenAI ? 'openai' : isGemini ? 'gemini' : isAntigravity ? 'antigravity' : isGrok ? 'grok' : 'anthropic'"
          :show-project-id="isGemini && geminiOAuthType === 'code_assist'"
          :initial-input-method="grokInitialInputMethod"
          @generate-url="handleGenerateUrl"
          @cookie-auth="handleCookieAuth"
          @validate-refresh-token="handleGrokValidateRefreshToken"
          @import-sso="handleGrokImportSSO"
        />
        <p
          v-if="currentError"
          id="admin-reauthorization-error"
          class="sr-only"
          role="alert"
          aria-live="assertive"
        >
          {{ currentError }}
        </p>
      </div>
    </div>

    <template #footer>
      <div v-if="account" class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-between">
        <button type="button" class="btn btn-secondary w-full sm:w-auto" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          v-if="isManualInputMethod"
          type="button"
          :disabled="!canExchangeCode"
          :aria-busy="currentLoading"
          class="btn btn-primary w-full sm:w-auto"
          @click="handleExchangeCode"
        >
          <Icon
            v-if="currentLoading"
            name="refresh"
            size="sm"
            class="animate-spin"
            aria-hidden="true"
          />
          {{
            currentLoading
              ? t('admin.accounts.oauth.verifying')
              : t('admin.accounts.oauth.completeAuth')
          }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import {
  useAccountOAuth,
  type AddMethod,
  type AuthInputMethod
} from '@/composables/useAccountOAuth'
import { useOpenAIOAuth } from '@/composables/useOpenAIOAuth'
import { useGeminiOAuth } from '@/composables/useGeminiOAuth'
import { useAntigravityOAuth } from '@/composables/useAntigravityOAuth'
import { useGrokOAuth } from '@/composables/useGrokOAuth'
import type { Account } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import OAuthAuthorizationFlow from '@/components/account/OAuthAuthorizationFlow.vue'

// Type for exposed OAuthAuthorizationFlow component
// Note: defineExpose automatically unwraps refs, so we use the unwrapped types
interface OAuthFlowExposed {
  authCode: string
  oauthState: string
  projectId: string
  sessionKey: string
  inputMethod: AuthInputMethod
  reset: () => void
}

interface Props {
  show: boolean
  account: Account | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  reauthorized: [account: Account]
}>()

const appStore = useAppStore()
const { t } = useI18n()

// OAuth composables
const claudeOAuth = useAccountOAuth()
const openaiOAuth = useOpenAIOAuth()
const geminiOAuth = useGeminiOAuth()
const antigravityOAuth = useAntigravityOAuth()
const grokOAuth = useGrokOAuth()

// Refs
const oauthFlowRef = ref<OAuthFlowExposed | null>(null)

// State
const addMethod = ref<AddMethod>('oauth')
const geminiOAuthType = ref<'code_assist' | 'google_one' | 'ai_studio'>('code_assist')

const isOpenAI = computed(() => props.account?.platform === 'openai')
const isOpenAILike = computed(() => isOpenAI.value)
const isGemini = computed(() => props.account?.platform === 'gemini')
const isAnthropic = computed(() => props.account?.platform === 'anthropic')
const isAntigravity = computed(() => props.account?.platform === 'antigravity')
const isGrok = computed(() => props.account?.platform === 'grok')

const platformLabel = computed(() => {
  if (isOpenAI.value) return t('admin.accounts.openaiAccount')
  if (isGemini.value) return t('admin.accounts.geminiAccount')
  if (isAntigravity.value) return t('admin.accounts.antigravityAccount')
  if (isGrok.value) return t('admin.accounts.grokAccount')
  return t('admin.accounts.claudeCodeAccount')
})

const geminiOAuthTypeLabel = computed(() => {
  if (geminiOAuthType.value === 'google_one') return 'Google One'
  if (geminiOAuthType.value === 'code_assist') {
    return t('admin.accounts.gemini.oauthType.builtInTitle')
  }
  return t('admin.accounts.gemini.oauthType.customTitle')
})

const geminiOAuthTypeDescription = computed(() => {
  if (geminiOAuthType.value === 'google_one') {
    return t('admin.accounts.gemini.oauthType.googleOneDesc')
  }
  if (geminiOAuthType.value === 'code_assist') {
    return t('admin.accounts.gemini.oauthType.builtInDesc')
  }
  return t('admin.accounts.gemini.oauthType.customDesc')
})

/**
 * Grok reauth default tab (password auth is hidden):
 * - refresh_token when RT may still work
 * - SSO cookie otherwise
 */
const grokInitialInputMethod = computed<AuthInputMethod>(() => {
  if (!isGrok.value) return 'manual'
  const creds = (props.account?.credentials || {}) as Record<string, unknown>
  const hasRT =
    (typeof creds.refresh_token === 'string' && creds.refresh_token.trim() !== '') ||
    (typeof creds.has_refresh_token === 'boolean' && creds.has_refresh_token)
  if (hasRT) return 'refresh_token'
  return 'sso_cookie'
})

// Computed - current OAuth state based on platform
const currentAuthUrl = computed(() => {
  if (isOpenAILike.value) return openaiOAuth.authUrl.value
  if (isGemini.value) return geminiOAuth.authUrl.value
  if (isAntigravity.value) return antigravityOAuth.authUrl.value
  if (isGrok.value) return grokOAuth.authUrl.value
  return claudeOAuth.authUrl.value
})
const currentSessionId = computed(() => {
  if (isOpenAILike.value) return openaiOAuth.sessionId.value
  if (isGemini.value) return geminiOAuth.sessionId.value
  if (isAntigravity.value) return antigravityOAuth.sessionId.value
  if (isGrok.value) return grokOAuth.sessionId.value
  return claudeOAuth.sessionId.value
})
const currentLoading = computed(() => {
  if (isOpenAILike.value) return openaiOAuth.loading.value
  if (isGemini.value) return geminiOAuth.loading.value
  if (isAntigravity.value) return antigravityOAuth.loading.value
  if (isGrok.value) return grokOAuth.loading.value
  return claudeOAuth.loading.value
})
const currentError = computed(() => {
  if (isOpenAILike.value) return openaiOAuth.error.value
  if (isGemini.value) return geminiOAuth.error.value
  if (isAntigravity.value) return antigravityOAuth.error.value
  if (isGrok.value) return grokOAuth.error.value
  return claudeOAuth.error.value
})

// Computed — footer "complete auth" only for code-exchange flows, not SSO/password/RT.
const isManualInputMethod = computed(() => {
  const method = oauthFlowRef.value?.inputMethod
  if (method === 'sso_cookie' || method === 'email_password' || method === 'refresh_token') {
    return false
  }
  // OpenAI/Gemini/Antigravity/Grok use manual code paste by default (no cookie auth)
  return (
    isOpenAILike.value ||
    isGemini.value ||
    isAntigravity.value ||
    isGrok.value ||
    method === 'manual'
  )
})

const canExchangeCode = computed(() => {
  const authCode = oauthFlowRef.value?.authCode || ''
  const sessionId = currentSessionId.value
  const loading = currentLoading.value
  return Boolean(authCode.trim() && sessionId && !loading)
})

let operationVersion = 0

function resetState() {
  addMethod.value = 'oauth'
  geminiOAuthType.value = 'code_assist'
  claudeOAuth.resetState()
  openaiOAuth.resetState()
  geminiOAuth.resetState()
  antigravityOAuth.resetState()
  grokOAuth.resetState()
  oauthFlowRef.value?.reset()
}

function isOperationCurrent(version: number, accountId: number) {
  return operationVersion === version && props.show && props.account?.id === accountId
}

function errorMessage(error: unknown, fallbackKey: string) {
  const apiError = error as { response?: { data?: { detail?: string } } }
  return apiError.response?.data?.detail || t(fallbackKey)
}

watch(
  () => [props.show, props.account?.id] as const,
  ([isOpen, accountId], previous) => {
    const [wasOpen, previousAccountId] = previous ?? [false, undefined]
    operationVersion += 1

    if (!isOpen || !props.account) {
      resetState()
      return
    }

    if (!wasOpen || previousAccountId !== accountId) {
      resetState()

      if (
        isAnthropic.value &&
        (props.account.type === 'oauth' || props.account.type === 'setup-token')
      ) {
        addMethod.value = props.account.type as AddMethod
      }
      if (isGemini.value) {
        const creds = (props.account.credentials || {}) as Record<string, unknown>
        geminiOAuthType.value =
          creds.oauth_type === 'google_one'
            ? 'google_one'
            : creds.oauth_type === 'ai_studio'
              ? 'ai_studio'
              : 'code_assist'
      }
    }
  },
  { immediate: true }
)

const handleClose = () => {
  operationVersion += 1
  resetState()
  emit('close')
}

const handleGenerateUrl = async () => {
  const account = props.account
  if (!account) return

  const version = operationVersion
  const oauthType = geminiOAuthType.value

  if (account.platform === 'openai') {
    await openaiOAuth.generateAuthUrl(account.proxy_id)
  } else if (account.platform === 'gemini') {
    const creds = (account.credentials || {}) as Record<string, unknown>
    const tierId = typeof creds.tier_id === 'string' ? creds.tier_id : undefined
    const projectId = oauthType === 'code_assist' ? oauthFlowRef.value?.projectId : undefined
    await geminiOAuth.generateAuthUrl(account.proxy_id, projectId, oauthType, tierId)
  } else if (account.platform === 'antigravity') {
    await antigravityOAuth.generateAuthUrl(account.proxy_id)
  } else if (account.platform === 'grok') {
    await grokOAuth.generateAuthUrl(account.proxy_id)
  } else {
    await claudeOAuth.generateAuthUrl(addMethod.value, account.proxy_id)
  }

  if (!isOperationCurrent(version, account.id)) return
}

const handleExchangeCode = async () => {
  const account = props.account
  if (!account) return

  const version = operationVersion
  const authCode = oauthFlowRef.value?.authCode || ''
  if (!authCode.trim()) return

  if (account.platform === 'openai') {
    const oauthClient = openaiOAuth
    const sessionId = oauthClient.sessionId.value
    if (!sessionId) return
    const stateToUse = (oauthFlowRef.value?.oauthState || oauthClient.oauthState.value || '').trim()
    if (!stateToUse) {
      oauthClient.error.value = t('admin.accounts.oauth.authFailed')
      appStore.showError(oauthClient.error.value)
      return
    }

    const tokenInfo = await oauthClient.exchangeAuthCode(
      authCode.trim(),
      sessionId,
      stateToUse,
      account.proxy_id
    )
    if (!tokenInfo || !isOperationCurrent(version, account.id)) return

    const credentials = oauthClient.buildCredentials(tokenInfo)
    const extra = oauthClient.buildExtraInfo(tokenInfo)

    try {
      const updatedAccount = await adminAPI.accounts.applyOAuthCredentials(account.id, {
        type: 'oauth',
        credentials,
        extra
      })
      if (!isOperationCurrent(version, account.id)) return

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error: unknown) {
      if (!isOperationCurrent(version, account.id)) return
      oauthClient.error.value = errorMessage(error, 'admin.accounts.oauth.authFailed')
      appStore.showError(oauthClient.error.value)
    }
  } else if (account.platform === 'gemini') {
    const sessionId = geminiOAuth.sessionId.value
    if (!sessionId) return

    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || geminiOAuth.state.value
    if (!stateToUse) return

    const credentialsRecord = (account.credentials || {}) as Record<string, unknown>
    const oauthType = geminiOAuthType.value
    const tokenInfo = await geminiOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId,
      state: stateToUse,
      proxyId: account.proxy_id,
      oauthType,
      tierId: typeof credentialsRecord.tier_id === 'string' ? credentialsRecord.tier_id : undefined
    })
    if (!tokenInfo || !isOperationCurrent(version, account.id)) return

    const credentials = geminiOAuth.buildCredentials(tokenInfo)

    try {
      await adminAPI.accounts.update(account.id, {
        type: 'oauth',
        credentials
      })
      const updatedAccount = await adminAPI.accounts.clearError(account.id)
      if (!isOperationCurrent(version, account.id)) return

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error: unknown) {
      if (!isOperationCurrent(version, account.id)) return
      geminiOAuth.error.value = errorMessage(error, 'admin.accounts.oauth.authFailed')
      appStore.showError(geminiOAuth.error.value)
    }
  } else if (account.platform === 'antigravity') {
    const sessionId = antigravityOAuth.sessionId.value
    if (!sessionId) return

    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || antigravityOAuth.state.value
    if (!stateToUse) return

    const tokenInfo = await antigravityOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId,
      state: stateToUse,
      proxyId: account.proxy_id
    })
    if (!tokenInfo || !isOperationCurrent(version, account.id)) return

    const credentials = antigravityOAuth.buildCredentials(tokenInfo)

    try {
      await adminAPI.accounts.update(account.id, {
        type: 'oauth',
        credentials
      })
      const updatedAccount = await adminAPI.accounts.clearError(account.id)
      if (!isOperationCurrent(version, account.id)) return

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error: unknown) {
      if (!isOperationCurrent(version, account.id)) return
      antigravityOAuth.error.value = errorMessage(error, 'admin.accounts.oauth.authFailed')
      appStore.showError(antigravityOAuth.error.value)
    }
  } else if (account.platform === 'grok') {
    const sessionId = grokOAuth.sessionId.value
    if (!sessionId) return

    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || grokOAuth.state.value
    if (!stateToUse) return

    const tokenInfo = await grokOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId,
      state: stateToUse,
      proxyId: account.proxy_id
    })
    if (!tokenInfo || !isOperationCurrent(version, account.id)) return

    const credentials = grokOAuth.buildCredentials(tokenInfo)
    const extra = grokOAuth.buildExtraInfo(tokenInfo)

    try {
      const updatedAccount = await adminAPI.accounts.applyOAuthCredentials(account.id, {
        type: 'oauth',
        credentials,
        extra
      })
      if (!isOperationCurrent(version, account.id)) return

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error: unknown) {
      if (!isOperationCurrent(version, account.id)) return
      grokOAuth.error.value = errorMessage(error, 'admin.accounts.oauth.authFailed')
      appStore.showError(grokOAuth.error.value)
    }
  } else {
    const sessionId = claudeOAuth.sessionId.value
    if (!sessionId) return

    const method = addMethod.value
    claudeOAuth.loading.value = true
    claudeOAuth.error.value = ''

    try {
      const proxyConfig = account.proxy_id ? { proxy_id: account.proxy_id } : {}
      const endpoint =
        method === 'oauth'
          ? '/admin/accounts/exchange-code'
          : '/admin/accounts/exchange-setup-token-code'

      const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
        session_id: sessionId,
        code: authCode.trim(),
        ...proxyConfig
      })
      if (!isOperationCurrent(version, account.id)) return

      const extra = claudeOAuth.buildExtraInfo(tokenInfo)

      const updatedAccount = await adminAPI.accounts.applyOAuthCredentials(account.id, {
        type: method as 'oauth' | 'setup-token',
        credentials: tokenInfo as unknown as Record<string, unknown>,
        extra
      })
      if (!isOperationCurrent(version, account.id)) return

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error: unknown) {
      if (!isOperationCurrent(version, account.id)) return
      claudeOAuth.error.value = errorMessage(error, 'admin.accounts.oauth.authFailed')
      appStore.showError(claudeOAuth.error.value)
    } finally {
      if (isOperationCurrent(version, account.id)) {
        claudeOAuth.loading.value = false
      }
    }
  }
}

const handleCookieAuth = async (sessionKey: string) => {
  const account = props.account
  if (!account || account.platform !== 'anthropic') return

  const version = operationVersion
  const method = addMethod.value
  claudeOAuth.loading.value = true
  claudeOAuth.error.value = ''

  try {
    const proxyConfig = account.proxy_id ? { proxy_id: account.proxy_id } : {}
    const endpoint =
      method === 'oauth'
        ? '/admin/accounts/cookie-auth'
        : '/admin/accounts/setup-token-cookie-auth'

    const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
      session_id: '',
      code: sessionKey.trim(),
      ...proxyConfig
    })
    if (!isOperationCurrent(version, account.id)) return

    const extra = claudeOAuth.buildExtraInfo(tokenInfo)

    const updatedAccount = await adminAPI.accounts.applyOAuthCredentials(account.id, {
      type: method as 'oauth' | 'setup-token',
      credentials: tokenInfo as unknown as Record<string, unknown>,
      extra
    })
    if (!isOperationCurrent(version, account.id)) return

    appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
    emit('reauthorized', updatedAccount)
    handleClose()
  } catch (error: unknown) {
    if (!isOperationCurrent(version, account.id)) return
    claudeOAuth.error.value = errorMessage(error, 'admin.accounts.oauth.cookieAuthFailed')
  } finally {
    if (isOperationCurrent(version, account.id)) {
      claudeOAuth.loading.value = false
    }
  }
}

/** Apply Grok Build OAuth tokens onto the existing account (never store password/SSO). */
const applyGrokReauthTokenInfo = async (tokenInfo: {
  access_token?: string
  refresh_token?: string
  email?: string
  [key: string]: unknown
}) => {
  if (!props.account) return
  const credentials = grokOAuth.buildCredentials(tokenInfo as any)
  const extra = grokOAuth.buildExtraInfo(tokenInfo as any)
  const updatedAccount = await adminAPI.accounts.applyOAuthCredentials(props.account.id, {
    type: 'oauth',
    credentials,
    extra
  })
  appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
  emit('reauthorized', updatedAccount)
  handleClose()
}

/** Re-auth with a single SSO cookie → Build OAuth (not batch create). */
const handleGrokImportSSO = async (ssoInput: string) => {
  if (!props.account || !isGrok.value) return
  const ssoToken = ssoInput
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)[0]
  if (!ssoToken) return

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''
  try {
    const tokenInfo = await grokOAuth.validateSSOToken(ssoToken, props.account.proxy_id)
    if (!tokenInfo) return
    await applyGrokReauthTokenInfo(tokenInfo)
  } catch (error: any) {
    grokOAuth.error.value =
      error.response?.data?.detail ||
      error.message ||
      t('admin.accounts.oauth.grok.failedToValidateSSO', 'Failed to validate Grok SSO')
    appStore.showError(grokOAuth.error.value)
  } finally {
    grokOAuth.loading.value = false
  }
}

/** Re-auth with a single refresh token. */
const handleGrokValidateRefreshToken = async (refreshTokenInput: string) => {
  if (!props.account || !isGrok.value) return
  const refreshToken = refreshTokenInput
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)[0]
  if (!refreshToken) {
    grokOAuth.error.value = t('admin.accounts.oauth.grok.pleaseEnterRefreshToken')
    return
  }

  grokOAuth.loading.value = true
  grokOAuth.error.value = ''
  try {
    const tokenInfo = await grokOAuth.validateRefreshToken(refreshToken, props.account.proxy_id)
    if (!tokenInfo) return
    await applyGrokReauthTokenInfo(tokenInfo)
  } catch (error: any) {
    grokOAuth.error.value =
      error.response?.data?.detail ||
      error.message ||
      t('admin.accounts.oauth.grok.failedToValidateRT')
    appStore.showError(grokOAuth.error.value)
  } finally {
    grokOAuth.loading.value = false
  }
}
</script>
