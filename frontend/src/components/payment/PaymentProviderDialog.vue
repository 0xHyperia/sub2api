<template>
  <BaseDialog
    :show="show"
    :title="editing ? t('admin.settings.payment.editProvider') : t('admin.settings.payment.createProvider')"
    width="wide"
    @close="emit('close')"
  >
    <form id="provider-form" @submit.prevent="handleSave" class="space-y-4">
      <!-- Name + Key -->
      <div class="grid grid-cols-1 gap-4 min-[420px]:grid-cols-2">
        <div>
          <label for="payment-provider-name" class="input-label">
            {{ t('admin.settings.payment.providerName') }}
            <span class="text-danger-foreground" aria-hidden="true">*</span>
          </label>
          <input id="payment-provider-name" v-model="form.name" type="text" class="input" required />
        </div>
        <div>
          <label for="payment-provider-key" class="input-label">
            {{ t('admin.settings.payment.providerKey') }}
            <span class="text-danger-foreground" aria-hidden="true">*</span>
          </label>
          <Select
            v-model="form.provider_key"
            id="payment-provider-key"
            :options="(!!editing ? allKeyOptions : enabledKeyOptions) as SelectOption[]"
            :disabled="!!editing"
            @change="onKeyChange"
          />
        </div>
      </div>

      <!-- Toggles + Payment mode + Supported types (single row) -->
      <div class="flex flex-wrap items-center gap-x-5 gap-y-2">
        <ToggleSwitch :label="t('common.enabled')" :checked="form.enabled" @toggle="form.enabled = !form.enabled" />
        <ToggleSwitch :label="t('admin.settings.payment.refundEnabled')" :checked="form.refund_enabled" @toggle="form.refund_enabled = !form.refund_enabled; if (!form.refund_enabled) form.allow_user_refund = false" />
        <ToggleSwitch v-if="form.refund_enabled" :label="t('admin.settings.payment.allowUserRefund')" :checked="form.allow_user_refund" @toggle="form.allow_user_refund = !form.allow_user_refund" />
        <div v-if="supportsPaymentMode" class="flex flex-wrap items-center gap-2">
          <span class="text-xs font-medium text-foreground-muted">{{ t('admin.settings.payment.paymentMode') }}</span>
          <div class="flex flex-wrap gap-1.5" role="group" :aria-label="t('admin.settings.payment.paymentMode')">
            <button
              v-for="mode in paymentModeOptions"
              :key="mode.value"
              type="button"
              :aria-pressed="form.payment_mode === mode.value"
              @click="form.payment_mode = mode.value"
              :class="[
                'rounded-control border px-2.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus',
                form.payment_mode === mode.value
                  ? 'border-brand bg-brand text-brand-foreground'
                  : 'border-outline bg-surface text-foreground-muted hover:border-outline-strong hover:bg-surface-subtle hover:text-foreground',
              ]"
            >{{ mode.label }}</button>
          </div>
        </div>
        <div v-if="availableTypes.length > 1" class="flex items-center gap-2">
          <span class="text-xs font-medium text-foreground-muted">{{ t('admin.settings.payment.supportedTypes') }}</span>
          <div class="flex flex-wrap gap-1.5" role="group" :aria-label="t('admin.settings.payment.supportedTypes')">
            <button
              v-for="pt in availableTypes"
              :key="pt.value"
              type="button"
              :aria-pressed="isTypeSelected(pt.value)"
              @click="toggleType(pt.value)"
              :class="[
                'rounded-control border px-2.5 py-1 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus',
                isTypeSelected(pt.value)
                  ? 'border-brand bg-brand text-brand-foreground'
                  : 'border-outline bg-surface text-foreground-muted hover:border-outline-strong hover:bg-surface-subtle hover:text-foreground',
              ]"
            >{{ pt.label }}</button>
          </div>
        </div>
      </div>

      <div v-if="form.provider_key === 'easypay'" class="space-y-3 rounded-panel border border-outline p-3">
        <div class="flex flex-col gap-3 min-[420px]:flex-row min-[420px]:items-center min-[420px]:justify-between">
          <div class="min-w-0">
            <h5 class="text-sm font-medium text-foreground">
              {{ t('admin.settings.payment.easypayCustomMethods') }}
            </h5>
            <p class="mt-1 text-xs text-foreground-muted">
              {{ t('admin.settings.payment.easypayCustomMethodsHint') }}
            </p>
          </div>
          <button type="button" class="btn btn-secondary btn-sm w-full min-[420px]:w-auto" @click="addEasyPayCustomMethod">
            {{ t('admin.settings.payment.addCustomMethod') }}
          </button>
        </div>
        <div v-if="easyPayCustomMethods.length" class="space-y-2">
          <div
            v-for="(method, index) in easyPayCustomMethods"
            :key="index"
            class="grid grid-cols-1 items-end gap-2 sm:grid-cols-3 lg:grid-cols-[1fr_1fr_1fr_auto]"
          >
            <div>
              <label :for="`easypay-method-type-${index}`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.customMethodType') }}</label>
              <input :id="`easypay-method-type-${index}`" v-model="method.type" type="text" class="input mt-0.5" placeholder="credit_card" />
            </div>
            <div>
              <label :for="`easypay-method-upstream-${index}`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.customMethodUpstreamType') }}</label>
              <input :id="`easypay-method-upstream-${index}`" v-model="method.upstreamType" type="text" class="input mt-0.5" placeholder="credit_card" />
            </div>
            <div>
              <label :for="`easypay-method-name-${index}`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.customMethodDisplayName') }}</label>
              <input :id="`easypay-method-name-${index}`" v-model="method.displayName" type="text" class="input mt-0.5" placeholder="信用卡" />
            </div>
            <button
              type="button"
              class="btn btn-danger btn-sm w-full sm:col-span-3 lg:col-span-1 lg:w-auto"
              @click="removeEasyPayCustomMethod(index)"
            >
              {{ t('common.delete') }}
            </button>
          </div>
        </div>
      </div>


      <!-- Config fields -->
      <div class="border-t border-outline pt-4">
        <div class="mb-3 flex items-center gap-2">
          <h4 class="text-sm font-semibold text-foreground">
            {{ t('admin.settings.payment.providerConfig') }}
          </h4>
          <HelpTooltip v-if="paymentGuide" trigger="click" width-class="w-80">
            <template #trigger>
              <button
                type="button"
                class="inline-flex h-7 w-7 items-center justify-center rounded-control border border-outline text-foreground-subtle transition-colors hover:border-outline-strong hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus"
                :aria-label="t('admin.settings.payment.paymentGuideTrigger')"
                :title="t('admin.settings.payment.paymentGuideTrigger')"
              >
                <Icon name="infoCircle" size="sm" aria-hidden="true" />
              </button>
            </template>
            <div class="space-y-3">
              <p class="font-medium text-white">{{ paymentGuide.summary }}</p>
              <div
                v-for="item in paymentGuide.items"
                :key="item.title"
                class="space-y-1.5 border-t border-white/10 pt-2 first:border-t-0 first:pt-0"
              >
                <p class="font-medium text-white">{{ item.title }}</p>
                <p><span class="text-foreground-subtle">{{ t('admin.settings.payment.guideOpenLabel') }}</span>{{ item.open }}</p>
                <p><span class="text-foreground-subtle">{{ t('admin.settings.payment.guideCallLabel') }}</span>{{ item.call }}</p>
                <p><span class="text-foreground-subtle">{{ t('admin.settings.payment.guideFallbackLabel') }}</span>{{ item.fallback }}</p>
              </div>
              <p v-if="paymentGuide.note" class="border-t border-white/10 pt-2 text-[11px] text-foreground-subtle">
                {{ paymentGuide.note }}
              </p>
            </div>
          </HelpTooltip>
        </div>
        <p v-if="paymentGuide" class="mb-3 text-xs text-foreground-muted">
          {{ paymentGuide.summary }}
        </p>
        <div class="space-y-3">
          <div v-for="field in resolvedFields" :key="field.key">
            <label :for="`payment-provider-config-${field.key}`" class="input-label">
              {{ field.label }}
              <span v-if="field.optional" class="text-xs text-foreground-subtle">({{ t('common.optional') }})</span>
              <span v-else class="text-danger-foreground" aria-hidden="true"> *</span>
            </label>
            <textarea
              v-if="field.sensitive && field.key.toLowerCase().includes('key') && field.key !== 'pkey'"
              :id="`payment-provider-config-${field.key}`"
              v-model="config[field.key]"
              rows="3"
              class="input font-mono text-xs"
              autocomplete="new-password"
              data-1p-ignore
              data-lpignore="true"
              data-bwignore="true"
              spellcheck="false"
              :aria-describedby="field.hintKey ? `payment-provider-config-${field.key}-hint` : undefined"
              :placeholder="editing ? t('admin.accounts.leaveEmptyToKeep') : ''"
            />
            <div v-else-if="field.sensitive" class="relative">
              <input
                :id="`payment-provider-config-${field.key}`"
                :type="visibleFields[field.key] ? 'text' : 'password'"
                v-model="config[field.key]"
                class="input pr-10"
                autocomplete="new-password"
                data-1p-ignore
                data-lpignore="true"
                data-bwignore="true"
                spellcheck="false"
                :aria-describedby="field.hintKey ? `payment-provider-config-${field.key}-hint` : undefined"
                :placeholder="editing ? t('admin.accounts.leaveEmptyToKeep') : (field.defaultValue || '')"
              />
              <button
                type="button"
                @click="visibleFields[field.key] = !visibleFields[field.key]"
                class="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-foreground-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus"
                :aria-label="visibleFields[field.key] ? t('common.hidePassword') : t('common.showPassword')"
                :aria-pressed="!!visibleFields[field.key]"
              >
                <Icon :name="visibleFields[field.key] ? 'eyeOff' : 'eye'" size="sm" aria-hidden="true" />
              </button>
            </div>
            <Select
              v-else-if="field.options?.length"
              :id="`payment-provider-config-${field.key}`"
              v-model="config[field.key]"
              :options="field.options"
              :searchable="field.options.length > 5"
              :aria-describedby="field.hintKey ? `payment-provider-config-${field.key}-hint` : undefined"
            />
            <input
              v-else
              :id="`payment-provider-config-${field.key}`"
              type="text"
              v-model="config[field.key]"
              class="input"
              :aria-describedby="field.hintKey ? `payment-provider-config-${field.key}-hint` : undefined"
              :placeholder="field.defaultValue || ''"
            />
            <p v-if="field.hintKey" :id="`payment-provider-config-${field.key}-hint`" class="mt-1 text-xs leading-relaxed text-foreground-muted">
              {{ t(field.hintKey) }}
            </p>
          </div>
        </div>

        <!-- Callback URLs (each = editable URL + fixed path) -->
        <div v-if="callbackPaths" class="mt-4 space-y-3">
          <div v-if="callbackPaths.notifyUrl">
            <label for="payment-provider-notify-url" class="input-label">{{ t('admin.settings.payment.field_notifyUrl') }} <span class="text-danger-foreground" aria-hidden="true">*</span></label>
            <div class="flex min-w-0 flex-col gap-1 min-[480px]:flex-row min-[480px]:gap-0">
              <input id="payment-provider-notify-url" v-model="notifyBaseUrl" type="text" class="input min-w-0 flex-1 min-[480px]:!rounded-r-none min-[480px]:!border-r-0" :placeholder="defaultBaseUrl" />
              <span class="inline-flex min-w-0 items-center break-all rounded-control border border-outline bg-surface-subtle px-3 py-2 text-xs text-foreground-muted min-[480px]:rounded-l-none">{{ callbackPaths.notifyUrl }}</span>
            </div>
          </div>
          <div v-if="callbackPaths.returnUrl">
            <label for="payment-provider-return-url" class="input-label">{{ t('admin.settings.payment.field_returnUrl') }} <span class="text-danger-foreground" aria-hidden="true">*</span></label>
            <div class="flex min-w-0 flex-col gap-1 min-[480px]:flex-row min-[480px]:gap-0">
              <input id="payment-provider-return-url" v-model="returnBaseUrl" type="text" class="input min-w-0 flex-1 min-[480px]:!rounded-r-none min-[480px]:!border-r-0" :placeholder="defaultBaseUrl" />
              <span class="inline-flex min-w-0 items-center break-all rounded-control border border-outline bg-surface-subtle px-3 py-2 text-xs text-foreground-muted min-[480px]:rounded-l-none">{{ callbackPaths.returnUrl }}</span>
            </div>
          </div>
        </div>

        <!-- 服务商 Webhook 提示 -->
        <div v-if="providerWebhookUrl" class="mt-3 rounded-panel border border-info/30 bg-info-subtle p-3 text-info-foreground">
          <p class="text-xs">
            {{ t(providerWebhookHint) }}
          </p>
          <code class="mt-1 block break-all rounded-control bg-surface/60 px-2 py-1 text-xs">
            {{ providerWebhookUrl }}
          </code>
          <p v-if="form.provider_key === 'stripe'" class="mt-2 text-xs leading-relaxed">
            {{ t('admin.settings.payment.stripeWebhookApiVersionHint', { version: STRIPE_SDK_API_VERSION }) }}
          </p>
        </div>
      </div>

      <!-- Per-type limits (collapsible) -->
      <div v-if="limitableTypes.length" class="border-t border-outline pt-4">
        <button type="button" :aria-expanded="limitsExpanded" aria-controls="payment-provider-limits" @click="limitsExpanded = !limitsExpanded" class="flex w-full items-center justify-between rounded-control focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus">
          <h4 class="text-sm font-semibold text-foreground">
            {{ t('admin.settings.payment.limitsTitle') }}
          </h4>
          <Icon name="chevronDown" size="sm" :class="['text-foreground-subtle transition-transform', limitsExpanded && 'rotate-180']" aria-hidden="true" />
        </button>
        <div id="payment-provider-limits" v-show="limitsExpanded" class="mt-3 space-y-3">
          <div
            v-for="lt in limitableTypes"
            :key="lt.value"
            class="rounded-panel border border-outline p-3"
          >
            <p class="mb-2 text-xs font-medium text-foreground">{{ lt.label }}</p>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <div>
                <label :for="`payment-limit-${lt.value}-single-min`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.limitSingleMin') }}</label>
                <input
                  :id="`payment-limit-${lt.value}-single-min`"
                  type="number"
                  :value="getLimitVal(lt.value, 'singleMin')"
                  @input="setLimitVal(lt.value, 'singleMin', ($event.target as HTMLInputElement).value)"
                  class="input mt-0.5" min="1" step="0.01" :placeholder="limitPlaceholder(lt.value)"
                />
              </div>
              <div>
                <label :for="`payment-limit-${lt.value}-single-max`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.limitSingleMax') }}</label>
                <input
                  :id="`payment-limit-${lt.value}-single-max`"
                  type="number"
                  :value="getLimitVal(lt.value, 'singleMax')"
                  @input="setLimitVal(lt.value, 'singleMax', ($event.target as HTMLInputElement).value)"
                  class="input mt-0.5" min="1" step="0.01" :placeholder="limitPlaceholder(lt.value)"
                />
              </div>
              <div>
                <label :for="`payment-limit-${lt.value}-daily`" class="text-xs text-foreground-muted">{{ t('admin.settings.payment.limitDaily') }}</label>
                <input
                  :id="`payment-limit-${lt.value}-daily`"
                  type="number"
                  :value="getLimitVal(lt.value, 'dailyLimit')"
                  @input="setLimitVal(lt.value, 'dailyLimit', ($event.target as HTMLInputElement).value)"
                  class="input mt-0.5" min="1" step="0.01" :placeholder="limitPlaceholder(lt.value)"
                />
              </div>
            </div>
          </div>
          <p class="text-xs text-foreground-subtle">{{ t('admin.settings.payment.limitsHint') }}</p>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex w-full flex-col-reverse gap-2 min-[420px]:flex-row min-[420px]:justify-end">
        <button type="button" @click="emit('close')" class="btn btn-secondary w-full min-[420px]:w-auto">{{ t('common.cancel') }}</button>
        <button type="submit" form="provider-form" :disabled="saving" class="btn btn-primary w-full min-[420px]:w-auto">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import ToggleSwitch from './ToggleSwitch.vue'
import type { ProviderInstance } from '@/types/payment'
import type { EasyPayCustomMethod, TypeOption } from './providerConfig'
import {
  PROVIDER_CONFIG_FIELDS,
  PROVIDER_SUPPORTED_TYPES,
  PROVIDER_CALLBACK_PATHS,
  WEBHOOK_PATHS,
  PAYMENT_MODE_QRCODE,
  PAYMENT_MODE_POPUP,
  PAYMENT_MODE_REDIRECT,
  STRIPE_SDK_API_VERSION,
  getAvailableTypes,
  extractBaseUrl,
  parseEasyPayCustomMethods,
  serializeEasyPayCustomMethods,
} from './providerConfig'

/** Default payment_mode per provider key — "" means "no preference, use
 * provider's built-in default behavior". */
function defaultPaymentMode(providerKey: string): string {
  if (providerKey === 'easypay') return PAYMENT_MODE_QRCODE
  return ''
}

/** Provider keys whose admin UI exposes a payment_mode selector.
 * Other providers always send payment_mode = ''. */
function providerSupportsPaymentMode(providerKey: string): boolean {
  return providerKey === 'easypay' || providerKey === 'alipay'
}

/** Allowed payment_mode values per provider. Used to coerce DB values
 * from a different provider (or stale data) back to the default. */
function isValidPaymentMode(providerKey: string, mode: string): boolean {
  if (providerKey === 'easypay') {
    return mode === PAYMENT_MODE_QRCODE || mode === PAYMENT_MODE_POPUP
  }
  if (providerKey === 'alipay') {
    return mode === '' || mode === PAYMENT_MODE_REDIRECT
  }
  return mode === ''
}

const props = defineProps<{
  show: boolean
  saving: boolean
  editing: ProviderInstance | null
  allKeyOptions: TypeOption[]
  enabledKeyOptions: TypeOption[]
  allPaymentTypes: TypeOption[]
  redirectLabel: string
}>()

const emit = defineEmits<{
  close: []
  save: [payload: {
    provider_key: string
    name: string
    supported_types: string[]
    enabled: boolean
    payment_mode: string
    refund_enabled: boolean
    allow_user_refund: boolean
    config: Record<string, string>
    limits: string
  }]
}>()

const { t } = useI18n()

interface PaymentGuideItem {
  title: string
  open: string
  call: string
  fallback: string
}

interface PaymentGuide {
  summary: string
  items: PaymentGuideItem[]
  note?: string
}

// --- Form state ---
const form = reactive({
  name: '',
  provider_key: 'easypay',
  supported_types: [] as string[],
  enabled: true,
  payment_mode: PAYMENT_MODE_QRCODE,
  refund_enabled: false,
  allow_user_refund: false,
})
const config = reactive<Record<string, string>>({})
const limits = reactive<Record<string, Record<string, number>>>({})
const notifyBaseUrl = ref('')
const returnBaseUrl = ref('')
const limitsExpanded = ref(false)
const visibleFields = reactive<Record<string, boolean>>({})
const easyPayCustomMethods = reactive<EasyPayCustomMethod[]>([])

// --- Computed ---
const defaultBaseUrl = typeof window !== 'undefined' ? window.location.origin : ''

const providerWebhookHintMap: Record<string, string> = {
  stripe: 'admin.settings.payment.stripeWebhookHint',
  airwallex: 'admin.settings.payment.airwallexWebhookHint',
}

const providerWebhookUrl = computed(() => {
  const path = WEBHOOK_PATHS[form.provider_key]
  return providerWebhookHintMap[form.provider_key] && path ? defaultBaseUrl + path : ''
})

const providerWebhookHint = computed(() =>
  providerWebhookHintMap[form.provider_key] || 'admin.settings.payment.stripeWebhookHint',
)

const callbackPaths = computed(() => PROVIDER_CALLBACK_PATHS[form.provider_key] || null)

const supportsPaymentMode = computed(() => providerSupportsPaymentMode(form.provider_key))

const paymentModeOptions = computed(() => {
  if (form.provider_key === 'alipay') {
    // For Alipay official: "" = default (precreate → page.pay fallback);
    // "redirect" = always open the Alipay checkout page in a new tab.
    return [
      { value: '', label: t('admin.settings.payment.modeQRCode') },
      { value: PAYMENT_MODE_REDIRECT, label: t('admin.settings.payment.modeRedirect') },
    ]
  }
  return [
    { value: PAYMENT_MODE_QRCODE, label: t('admin.settings.payment.modeQRCode') },
    { value: PAYMENT_MODE_POPUP, label: t('admin.settings.payment.modePopup') },
  ]
})

const availableTypes = computed(() => {
  const base = getAvailableTypes(form.provider_key, props.allPaymentTypes, props.redirectLabel)
  if (form.provider_key === 'easypay') {
    for (const method of normalizedEasyPayCustomMethods()) {
      if (!base.some(opt => opt.value === method.type)) {
        base.push({
          value: method.type,
          label: method.displayName || method.type,
        })
      }
    }
  }
  // Resolve i18n labels for types not in allPaymentTypes (e.g. card, link inside stripe)
  return base.map(opt =>
    opt.label === opt.value
      ? { ...opt, label: t(`payment.methods.${opt.value}`, opt.value) }
      : opt,
  )
})

const resolvedFields = computed(() => {
  const fields = PROVIDER_CONFIG_FIELDS[form.provider_key] || []
  return fields.map(f => ({
    ...f,
    label: f.label || t(`admin.settings.payment.field_${f.key}`),
  }))
})

const paymentGuide = computed<PaymentGuide | null>(() => {
  if (form.provider_key === 'alipay') {
    return {
      summary: t('admin.settings.payment.alipayGuideSummary'),
      items: [
        {
          title: t('admin.settings.payment.alipayGuideFaceToFaceTitle'),
          open: t('admin.settings.payment.alipayGuideFaceToFaceOpen'),
          call: t('admin.settings.payment.alipayGuideFaceToFaceCall'),
          fallback: t('admin.settings.payment.alipayGuideFaceToFaceFallback'),
        },
        {
          title: t('admin.settings.payment.alipayGuidePagePayTitle'),
          open: t('admin.settings.payment.alipayGuidePagePayOpen'),
          call: t('admin.settings.payment.alipayGuidePagePayCall'),
          fallback: t('admin.settings.payment.alipayGuidePagePayFallback'),
        },
        {
          title: t('admin.settings.payment.alipayGuideWapTitle'),
          open: t('admin.settings.payment.alipayGuideWapOpen'),
          call: t('admin.settings.payment.alipayGuideWapCall'),
          fallback: t('admin.settings.payment.alipayGuideWapFallback'),
        },
      ],
    }
  }

  if (form.provider_key === 'wxpay') {
    return {
      summary: t('admin.settings.payment.wxpayGuideSummary'),
      note: t('admin.settings.payment.wxpayGuideNote'),
      items: [
        {
          title: t('admin.settings.payment.wxpayGuideNativeTitle'),
          open: t('admin.settings.payment.wxpayGuideNativeOpen'),
          call: t('admin.settings.payment.wxpayGuideNativeCall'),
          fallback: t('admin.settings.payment.wxpayGuideNativeFallback'),
        },
        {
          title: t('admin.settings.payment.wxpayGuideJsapiTitle'),
          open: t('admin.settings.payment.wxpayGuideJsapiOpen'),
          call: t('admin.settings.payment.wxpayGuideJsapiCall'),
          fallback: t('admin.settings.payment.wxpayGuideJsapiFallback'),
        },
        {
          title: t('admin.settings.payment.wxpayGuideH5Title'),
          open: t('admin.settings.payment.wxpayGuideH5Open'),
          call: t('admin.settings.payment.wxpayGuideH5Call'),
          fallback: t('admin.settings.payment.wxpayGuideH5Fallback'),
        },
      ],
    }
  }

  if (form.provider_key === 'airwallex') {
    return {
      summary: t('admin.settings.payment.airwallexGuideSummary'),
      note: t('admin.settings.payment.airwallexGuideNote'),
      items: [],
    }
  }

  return null
})

const limitableTypes = computed(() => {
  // Stripe: single "stripe" entry (one set of shared limits)
  if (form.provider_key === 'stripe') {
    return [{ value: 'stripe', label: 'Stripe' }]
  }
  const selected = form.supported_types.filter(t => t !== 'easypay')
  return selected.map(v => {
    const found = props.allPaymentTypes.find(pt => pt.value === v)
    return found || { value: v, label: v }
  })
})

// --- Methods ---
function isTypeSelected(type: string): boolean {
  return form.supported_types.includes(type)
}

function toggleType(type: string) {
  if (form.supported_types.includes(type)) {
    form.supported_types = form.supported_types.filter(t => t !== type)
  } else {
    form.supported_types = [...form.supported_types, type]
  }
}

function normalizedEasyPayCustomMethods(): EasyPayCustomMethod[] {
  return easyPayCustomMethods
    .map(method => ({
      type: normalizeEasyPayCustomMethodCode(method.type),
      upstreamType: normalizeEasyPayCustomMethodCode(method.upstreamType),
      displayName: method.displayName.trim(),
    }))
    .filter(method => method.type || method.upstreamType || method.displayName)
}

function normalizeEasyPayCustomMethodCode(value: string): string {
  return value.trim().toLowerCase()
}

function addEasyPayCustomMethod() {
  easyPayCustomMethods.push({ type: '', upstreamType: '', displayName: '' })
}

function removeEasyPayCustomMethod(index: number) {
  easyPayCustomMethods.splice(index, 1)
}

function onKeyChange() {
  form.supported_types = [...(PROVIDER_SUPPORTED_TYPES[form.provider_key] || [])]
  form.payment_mode = defaultPaymentMode(form.provider_key)
  clearConfig()
  applyDefaults()
}

function clearConfig() {
  Object.keys(config).forEach(k => delete config[k])
  Object.keys(limits).forEach(k => delete limits[k])
  Object.keys(visibleFields).forEach(k => delete visibleFields[k])
  notifyBaseUrl.value = ''
  returnBaseUrl.value = ''
  limitsExpanded.value = false
  easyPayCustomMethods.splice(0, easyPayCustomMethods.length)
}

function applyDefaults() {
  for (const f of PROVIDER_CONFIG_FIELDS[form.provider_key] || []) {
    if (f.defaultValue && !config[f.key]) config[f.key] = f.defaultValue
  }
}

function getLimitVal(paymentType: string, field: string): string {
  const val = limits[paymentType]?.[field]
  return val && val > 0 ? String(val) : ''
}

/** Returns true if any limit field for this payment type has a value */
function hasAnyLimit(paymentType: string): boolean {
  const l = limits[paymentType]
  if (!l) return false
  return (l.singleMin > 0) || (l.singleMax > 0) || (l.dailyLimit > 0)
}

/** Dynamic placeholder: "不限制" if sibling has value, "使用全局配置" if all empty */
function limitPlaceholder(paymentType: string): string {
  return hasAnyLimit(paymentType)
    ? t('admin.settings.payment.limitsNoLimit')
    : t('admin.settings.payment.limitsUseGlobal')
}

function setLimitVal(paymentType: string, field: string, val: string) {
  if (!limits[paymentType]) limits[paymentType] = {}
  const num = Number(val)
  // Empty → clear the field (use global); reject ≤0
  if (val === '' || isNaN(num)) {
    delete limits[paymentType][field]
    return
  }
  if (num <= 0) return
  limits[paymentType][field] = num
}

function serializeLimits(): string {
  const result: Record<string, Record<string, number>> = {}
  for (const [pt, fields] of Object.entries(limits)) {
    const clean: Record<string, number> = {}
    for (const [k, v] of Object.entries(fields)) {
      if (v > 0) clean[k] = v
    }
    if (Object.keys(clean).length > 0) result[pt] = clean
  }
  return Object.keys(result).length > 0 ? JSON.stringify(result) : ''
}

function handleSave() {
  // Validate required fields
  if (!form.name.trim()) {
    emitValidationError(t('admin.settings.payment.validationNameRequired'))
    return
  }
  if (form.provider_key === 'easypay') {
    const validationError = validateEasyPayCustomMethods()
    if (validationError) {
      emitValidationError(validationError)
      return
    }
    syncEasyPayCustomMethods()
  }
  // Validate required config fields — all non-optional fields must be filled.
  // In edit mode, sensitive fields may be left blank to preserve the stored
  // value (backend merges blanks by preserving the existing secret).
  for (const f of PROVIDER_CONFIG_FIELDS[form.provider_key] || []) {
    if (f.optional) continue
    if (props.editing && f.sensitive) continue
    const val = (config[f.key] || '').trim()
    if (!val) {
      const label = f.label || t(`admin.settings.payment.field_${f.key}`)
      emitValidationError(t('admin.settings.payment.validationFieldRequired', { field: label }))
      return
    }
  }

  const clearableConfigKeys = new Set(
    (PROVIDER_CONFIG_FIELDS[form.provider_key] || [])
      .filter(field => field.clearable)
      .map(field => field.key),
  )
  const filteredConfig: Record<string, string> = {}
  for (const [k, v] of Object.entries(config)) {
    if (!v || !v.trim()) {
      if (clearableConfigKeys.has(k)) {
        filteredConfig[k] = ''
      }
      continue
    }
    filteredConfig[k] = v
  }
  if (form.provider_key === 'easypay') {
    filteredConfig.customMethods = serializeEasyPayCustomMethods(normalizedEasyPayCustomMethods())
  }

  // Inject computed callback URLs (each URL = independent base + fixed path)
  // If base URL is empty, auto-fill with current domain
  const paths = PROVIDER_CALLBACK_PATHS[form.provider_key]
  if (paths) {
    const notifyBase = notifyBaseUrl.value.trim() || defaultBaseUrl
    const returnBase = returnBaseUrl.value.trim() || defaultBaseUrl
    notifyBaseUrl.value = notifyBase
    returnBaseUrl.value = returnBase
    if (paths.notifyUrl) filteredConfig['notifyUrl'] = notifyBase + paths.notifyUrl
    if (paths.returnUrl) filteredConfig['returnUrl'] = returnBase + paths.returnUrl
  }

  emit('save', {
    provider_key: form.provider_key,
    name: form.name,
    supported_types: form.supported_types,
    enabled: form.enabled,
    payment_mode: supportsPaymentMode.value ? form.payment_mode : '',
    refund_enabled: form.refund_enabled,
    allow_user_refund: form.refund_enabled ? form.allow_user_refund : false,
    config: filteredConfig,
    limits: serializeLimits(),
  })
}

function syncEasyPayCustomMethods(): string[] {
  if (form.provider_key !== 'easypay') return []
  const baseTypes = new Set(PROVIDER_SUPPORTED_TYPES.easypay || [])
  const customTypes: string[] = []
  const seen = new Set<string>()
  for (const method of normalizedEasyPayCustomMethods()) {
    if (!method.type || !method.upstreamType) continue
    if (seen.has(method.type)) continue
    seen.add(method.type)
    customTypes.push(method.type)
  }
  form.supported_types = form.supported_types
    .map(type => normalizeEasyPayCustomMethodCode(type))
    .filter(type => baseTypes.has(type) || customTypes.includes(type))
  for (const customType of customTypes) {
    if (!form.supported_types.includes(customType)) {
      form.supported_types.push(customType)
    }
  }
  return customTypes
}

function validateEasyPayCustomMethods(): string | null {
  const seen = new Set<string>()
  for (const method of normalizedEasyPayCustomMethods()) {
    const hasAnyValue = Boolean(method.type || method.upstreamType || method.displayName)
    if (!hasAnyValue) continue
    if (!method.type || !method.upstreamType) {
      return t('admin.settings.payment.validationEasyPayCustomMethodRequired')
    }
    if (!/^[a-z0-9_-]+$/.test(method.type)) {
      return t('admin.settings.payment.validationEasyPayCustomMethodTypeInvalid')
    }
    if (!/^[a-z0-9_-]+$/.test(method.upstreamType)) {
      return t('admin.settings.payment.validationEasyPayCustomMethodUpstreamTypeInvalid')
    }
    if ((PROVIDER_SUPPORTED_TYPES.easypay || []).includes(method.type)) {
      return t('admin.settings.payment.validationEasyPayCustomMethodReserved')
    }
    if (method.type.startsWith('alipay') || method.type.startsWith('wxpay')) {
      return t('admin.settings.payment.validationEasyPayCustomMethodPrefixReserved')
    }
    if (seen.has(method.type)) {
      return t('admin.settings.payment.validationEasyPayCustomMethodDuplicate')
    }
    seen.add(method.type)
  }
  return null
}

function emitValidationError(msg: string) {
  // Use a custom event or inject appStore — for now use window alert fallback
  // The parent handles this via the save event validation
  import('@/stores').then(m => m.useAppStore().showError(msg))
}

// --- Public API for parent to call ---
function reset(defaultKey: string) {
  form.name = ''
  form.provider_key = defaultKey
  form.supported_types = [...(PROVIDER_SUPPORTED_TYPES[defaultKey] || [])]
  form.enabled = true
  form.payment_mode = defaultPaymentMode(defaultKey)
  form.refund_enabled = false
  form.allow_user_refund = false
  clearConfig()
  applyDefaults()
}

function loadProvider(provider: ProviderInstance) {
  form.name = provider.name
  form.provider_key = provider.provider_key
  form.supported_types = Array.isArray(provider.supported_types)
    ? [...provider.supported_types]
    : []
  form.enabled = provider.enabled
  // Coerce to a valid value for this provider. Guards against stale data
  // (e.g. "popup" written by an older client) showing up as an unselected
  // button in the dialog.
  form.payment_mode = isValidPaymentMode(provider.provider_key, provider.payment_mode || '')
    ? (provider.payment_mode || '')
    : defaultPaymentMode(provider.provider_key)
  form.refund_enabled = provider.refund_enabled
  form.allow_user_refund = provider.allow_user_refund
  clearConfig()
  // Pre-fill config from API response. Backend omits sensitive fields entirely,
  // so those inputs stay blank — submitting blank preserves the stored secret.
  if (provider.config) {
    for (const [k, v] of Object.entries(provider.config)) {
      // Skip notifyUrl/returnUrl — they are derived from callbackBaseUrl
      if (k === 'notifyUrl' || k === 'returnUrl') continue
      if (k === 'customMethods' && provider.provider_key === 'easypay') {
        easyPayCustomMethods.push(...parseEasyPayCustomMethods(v))
        continue
      }
      config[k] = v
    }
    // Extract base URLs from existing callback URLs
    const paths = PROVIDER_CALLBACK_PATHS[provider.provider_key]
    if (paths?.notifyUrl && provider.config['notifyUrl']) {
      notifyBaseUrl.value = extractBaseUrl(provider.config['notifyUrl'], paths.notifyUrl)
    }
    if (paths?.returnUrl && provider.config['returnUrl']) {
      returnBaseUrl.value = extractBaseUrl(provider.config['returnUrl'], paths.returnUrl)
    }
  }
  applyDefaults()
  // Parse existing limits
  if (provider.limits) {
    try {
      const parsed = JSON.parse(provider.limits)
      for (const [pt, fields] of Object.entries(parsed as Record<string, Record<string, number>>)) {
        limits[pt] = { ...fields }
      }
      limitsExpanded.value = Object.keys(limits).length > 0
    } catch { /* ignore */ }
  }
}

defineExpose({ reset, loadProvider })
</script>
