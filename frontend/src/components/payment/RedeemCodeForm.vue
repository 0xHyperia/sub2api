<template>
  <section class="space-y-3" aria-labelledby="redeem-code-form-title">
    <div>
      <h2 id="redeem-code-form-title" class="text-sm font-semibold text-foreground">
        {{ t('redeem.redeemCodeLabel') }}
      </h2>
      <p v-if="showDescription" class="mt-1 text-xs leading-5 text-foreground-subtle">
        {{ t('redeem.description') }}
      </p>
    </div>

    <form class="flex min-w-0 flex-col gap-2 sm:flex-row" @submit.prevent="handleRedeem">
      <label for="shared-redeem-code" class="sr-only">{{ t('redeem.redeemCodeLabel') }}</label>
      <div class="relative min-w-0 flex-1">
        <Icon
          name="gift"
          size="sm"
          class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-foreground-subtle"
          aria-hidden="true"
        />
        <input
          id="shared-redeem-code"
          v-model="redeemCode"
          type="text"
          required
          autocomplete="off"
          autocapitalize="none"
          spellcheck="false"
          :placeholder="t('redeem.redeemCodePlaceholder')"
          :disabled="submitting"
          :aria-invalid="!!errorMessage"
          class="input w-full pl-9"
        />
      </div>
      <button
        type="submit"
        class="btn btn-primary shrink-0 sm:min-w-28"
        :disabled="!redeemCode.trim() || submitting"
        :aria-busy="submitting"
      >
        <LoadingSpinner v-if="submitting" size="sm" color="white" />
        <Icon v-else name="gift" size="sm" aria-hidden="true" />
        <span>{{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}</span>
      </button>
    </form>

    <div
      v-if="redeemResult"
      class="rounded-panel border border-success/20 bg-success-subtle px-3 py-2.5"
      role="status"
      aria-live="polite"
    >
      <div class="flex items-start gap-2">
        <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-success-foreground" aria-hidden="true" />
        <div class="min-w-0 text-sm leading-5 text-success-foreground">
          <p class="font-medium">{{ redeemResult.message || t('redeem.redeemSuccess') }}</p>
          <p v-if="redeemResult.type === 'balance'">
            {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
          </p>
          <p v-else-if="redeemResult.type === 'concurrency'">
            {{ t('redeem.added') }}: {{ redeemResult.value }} {{ t('redeem.concurrentRequests') }}
          </p>
          <p v-else-if="redeemResult.type === 'subscription'">
            {{ t('redeem.subscriptionAssigned') }}
            <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
          </p>
        </div>
      </div>
    </div>

    <div
      v-if="errorMessage"
      class="rounded-panel border border-danger/20 bg-danger-subtle px-3 py-2.5 text-sm text-danger-foreground"
      role="alert"
    >
      {{ errorMessage }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { redeemAPI, type RedeemResult } from '@/api/redeem'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { extractApiErrorMessage } from '@/utils/apiError'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

withDefaults(defineProps<{
  showDescription?: boolean
}>(), {
  showDescription: false,
})

const emit = defineEmits<{
  redeemed: [result: RedeemResult]
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<RedeemResult | null>(null)
const errorMessage = ref('')

async function handleRedeem(): Promise<void> {
  const code = redeemCode.value.trim()
  if (!code || submitting.value) return

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(code)
    redeemResult.value = result
    await authStore.refreshUser()

    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true)
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    redeemCode.value = ''
    emit('redeemed', result)
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: unknown) {
    errorMessage.value = extractApiErrorMessage(error, t('redeem.failedToRedeem'))
    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}
</script>
