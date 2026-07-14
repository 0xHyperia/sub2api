<template>
  <main class="flex min-h-screen items-center justify-center bg-canvas px-4 py-10 text-foreground">
    <section class="w-full max-w-xl overflow-hidden rounded-panel border border-outline bg-surface shadow-modal">
      <header class="border-b border-outline px-5 py-5 sm:px-6">
        <div class="flex items-center gap-4">
          <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-panel border border-outline bg-surface-subtle text-foreground-muted">
            <Icon name="shield" size="lg" />
          </div>
          <div class="min-w-0">
            <h1 class="text-lg font-semibold text-foreground">
              {{ t('appAuthorization.title') }}
            </h1>
            <p class="mt-1 truncate text-sm text-foreground-muted">
              {{ context?.client_name || context?.client_id || 'ZeroBox' }}
            </p>
          </div>
        </div>
      </header>

      <div v-if="loading" data-testid="authorization-loading" class="flex justify-center px-6 py-16">
        <span class="h-8 w-8 animate-spin rounded-full border-2 border-outline border-t-foreground" />
      </div>

      <div v-else-if="errorMessage" data-testid="authorization-error" class="px-6 py-10 text-center">
        <div class="mx-auto flex h-11 w-11 items-center justify-center rounded-panel bg-danger-subtle text-danger-foreground">
          <Icon name="exclamationTriangle" size="lg" />
        </div>
        <p class="mt-4 font-medium text-foreground">{{ t('appAuthorization.invalidRequest') }}</p>
        <p class="mt-2 text-sm text-foreground-muted">{{ errorMessage }}</p>
      </div>

      <template v-else-if="context">
        <div class="space-y-6 px-6 py-6">
          <p class="text-sm leading-6 text-foreground-muted">
            {{ t('appAuthorization.requestDescription', { app: context.client_name || context.client_id }) }}
          </p>

          <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-3 rounded-panel border border-outline bg-surface-subtle px-4 py-3 text-sm">
            <dt class="text-foreground-muted">{{ t('appAuthorization.device') }}</dt>
            <dd class="min-w-0 break-words text-right font-medium text-foreground">
              {{ context.device_name || t('appAuthorization.unknownDevice') }}
            </dd>
            <dt class="text-foreground-muted">{{ t('appAuthorization.platform') }}</dt>
            <dd class="text-right font-medium text-foreground">
              {{ context.platform || t('appAuthorization.unknownPlatform') }}
            </dd>
          </dl>

          <div>
            <h2 class="text-sm font-semibold text-foreground">
              {{ t('appAuthorization.permissionsTitle') }}
            </h2>
            <ul class="mt-3 space-y-2">
              <li v-for="scope in context.scopes" :key="scope" class="flex items-start gap-3 text-sm text-foreground-muted">
                <Icon name="check" size="sm" class="mt-0.5 shrink-0 text-success-foreground" />
                <span>{{ scopeLabel(scope) }}</span>
              </li>
            </ul>
          </div>
        </div>

        <footer class="flex flex-col-reverse gap-3 border-t border-outline bg-surface-subtle px-5 py-4 sm:flex-row sm:justify-end sm:px-6">
          <button type="button" class="btn btn-secondary" :disabled="submitting" @click="decide('deny')">
            {{ t('appAuthorization.deny') }}
          </button>
          <button type="button" class="btn btn-primary" :disabled="submitting" @click="decide('allow')">
            <span v-if="submitting" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white" />
            {{ t('appAuthorization.allow') }}
          </button>
        </footer>
      </template>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@/components/icons'
import { useAuthStore } from '@/stores/auth'
import {
  createAuthorizationRequest,
  getAuthorizationContext,
  submitAuthorizationDecision,
  type AppAuthorizationContext,
  type AuthorizationDecision,
  type AuthorizationRequestParams
} from '@/api/appAuth'
import { extractApiErrorMessage } from '@/utils/apiError'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const context = ref<AppAuthorizationContext | null>(null)

const requiredParameters = [
  'response_type',
  'client_id',
  'redirect_uri',
  'scope',
  'code_challenge',
  'code_challenge_method'
] as const

function queryString(name: string): string {
  const value = route.query[name]
  return typeof value === 'string' ? value : ''
}

function buildAuthorizationRequest(): AuthorizationRequestParams | null {
  if (requiredParameters.some((name) => !queryString(name))) return null
  return {
    response_type: queryString('response_type'),
    client_id: queryString('client_id'),
    redirect_uri: queryString('redirect_uri'),
    scope: queryString('scope'),
    code_challenge: queryString('code_challenge'),
    code_challenge_method: queryString('code_challenge_method'),
    state: queryString('state') || undefined,
    device_name: queryString('device_name') || undefined,
    platform: queryString('platform') || undefined
  }
}

async function loadContext(): Promise<void> {
  loading.value = true
  errorMessage.value = ''
  try {
    let requestId = queryString('request_id')
    if (!requestId) {
      const request = buildAuthorizationRequest()
      if (!request) throw new Error(t('appAuthorization.missingParameters'))
      const result = await createAuthorizationRequest(request)
      requestId = result.request_id
      await router.replace({ name: 'AppAuthorization', query: { request_id: requestId } })
    }
    if (!authStore.isAuthenticated) {
      await router.replace({
        path: '/login',
        query: { redirect: `/oauth/authorize?request_id=${encodeURIComponent(requestId)}` }
      })
      return
    }
    context.value = await getAuthorizationContext(requestId)
  } catch (error) {
    errorMessage.value = extractApiErrorMessage(error, t('appAuthorization.loadFailed'))
  } finally {
    loading.value = false
  }
}

function scopeLabel(scope: string): string {
  const key = `appAuthorization.scopes.${scope.replace(':', '_')}`
  const translated = t(key)
  return translated === key ? scope : translated
}

async function decide(decision: AuthorizationDecision): Promise<void> {
  if (!context.value || submitting.value) return
  submitting.value = true
  errorMessage.value = ''
  try {
    const result = await submitAuthorizationDecision(context.value.request_id, decision)
    window.location.assign(result.redirect_uri)
  } catch (error) {
    errorMessage.value = extractApiErrorMessage(error, t('appAuthorization.decisionFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(loadContext)
</script>
