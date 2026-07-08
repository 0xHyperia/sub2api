<template>
  <div :class="embedded ? 'space-y-6' : 'mx-auto max-w-[1500px] space-y-6 p-4 sm:p-6'">
    <div v-if="!embedded" class="flex items-center justify-between gap-3">
      <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('payment.card.title') }}</h1>
      <button
        class="btn btn-secondary"
        :disabled="loading"
        :title="t('payment.card.refreshGoods')"
        :aria-label="t('payment.card.refreshGoods')"
        @click="loadCheckoutInfo"
      >
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
      </button>
    </div>

    <PaymentStatusPanel
      v-if="createdOrder"
      :order-id="createdOrder.order_id"
      :qr-code="createdOrder.qr_code || ''"
      :pay-url="createdOrder.pay_url"
      :expires-at="createdOrder.expires_at"
      :payment-type="createdOrder.payment_type || 'ldxp'"
      :order-type="'card'"
      currency="CNY"
      @done="createdOrder = null"
    />

    <div v-else-if="loading && shops.length === 0" class="rounded-lg border border-gray-200 bg-white p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800">
      {{ t('common.loading') }}
    </div>

    <div v-else-if="!selectedShop" class="rounded-lg border border-gray-200 bg-white p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800">
      {{ t('payment.card.empty') }}
    </div>

    <div v-else class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_380px]">
      <section class="space-y-4">
        <div v-if="embedded" class="flex items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('payment.card.title') }}</h2>
          <button
            class="btn btn-secondary"
            :disabled="loading"
            :title="t('payment.card.refreshGoods')"
            :aria-label="t('payment.card.refreshGoods')"
            @click="loadCheckoutInfo"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="rounded-md px-3 py-1.5 text-sm transition"
              :class="selectedCategoryId === 0 ? activeChipClass : idleChipClass"
              @click="selectedCategoryId = 0"
            >
              {{ t('common.all') }}
            </button>
            <button
              v-for="category in selectedShop.categories || []"
              :key="category.id"
              type="button"
              class="rounded-md px-3 py-1.5 text-sm transition"
              :class="selectedCategoryId === category.id ? activeChipClass : idleChipClass"
              @click="selectedCategoryId = category.id"
            >
              {{ category.name }}
            </button>
          </div>
        </div>

        <div class="rounded-lg border border-primary-100 bg-primary-50 px-4 py-3 text-sm text-primary-700 dark:border-primary-900/40 dark:bg-primary-900/20 dark:text-primary-300">
          {{ t('payment.card.largeRechargeTip') }}
        </div>

        <div class="grid max-w-[1158px] grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
          <article
            v-for="goods in filteredGoods"
            :key="goods.goods_key"
            class="relative h-[149px] overflow-hidden rounded-[10px] border bg-white p-4 text-left transition dark:bg-dark-800"
            :class="[
              selectedGoods?.goods_key === goods.goods_key
                ? 'border-primary-500 ring-2 ring-primary-100 dark:border-primary-400 dark:ring-primary-900/40'
                : 'border-gray-200 hover:border-gray-300 dark:border-dark-700',
              goods.stock_count <= 0 ? 'opacity-60' : 'cursor-pointer'
            ]"
            @click="selectGoods(goods)"
          >
            <div class="grid h-full grid-rows-[20px_30px_18px_24px] gap-2">
              <div class="flex items-start justify-between gap-4">
                <h3 class="line-clamp-1 min-w-0 pr-16 text-[15px] font-bold leading-5 text-gray-900 dark:text-white">
                  {{ goodsTitle(goods) }}
                </h3>
                <span v-if="goods.badge" class="absolute right-4 top-4 rounded bg-gray-900 px-2 py-0.5 text-[10px] font-bold leading-4 text-white dark:bg-white dark:text-gray-900">
                  {{ goods.badge }}
                </span>
              </div>
              <div class="flex min-w-0 items-baseline gap-2">
                <span class="text-xs font-extrabold uppercase text-gray-900 dark:text-white">CNY</span>
                <span class="font-mono text-2xl font-extrabold tracking-normal text-gray-900 dark:text-white">{{ goods.price.toFixed(2) }}</span>
                <span v-if="shouldShowReferencePrice(goods)" class="font-mono text-xl font-bold text-gray-900 line-through dark:text-white">{{ goods.reference_price!.toFixed(2) }}</span>
              </div>
              <p v-if="goodsDescription(goods)" class="min-w-0 truncate text-xs leading-[18px] text-gray-500 dark:text-gray-400">
                {{ goodsDescription(goods) }}
              </p>
              <div class="flex max-h-6 min-w-0 flex-wrap gap-1.5 overflow-hidden pr-8">
                <span class="rounded border border-gray-200 bg-gray-50 px-2 py-0.5 text-[11px] text-gray-500 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300">
                  {{ stockStatusLabel(goods.stock_count) }}
                </span>
                <span
                  v-for="tag in goods.tags?.slice(0, 2)"
                  :key="tag"
                  class="rounded border border-gray-200 bg-gray-50 px-2 py-0.5 text-[11px] text-gray-500 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300"
                >
                  {{ tag }}
                </span>
              </div>
            </div>

            <div v-if="authStore.isAdmin" class="absolute bottom-3 right-3">
              <button
                type="button"
                class="rounded-md p-1.5 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                :title="t('payment.card.editGoods')"
                @click.stop="openEditDialog(goods)"
              >
                <Icon name="edit" size="sm" />
              </button>
            </div>
          </article>
        </div>
      </section>

      <aside class="space-y-4">
        <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4">
            <div class="text-xs font-medium uppercase tracking-wide text-gray-400">{{ t('payment.card.selectedGoods') }}</div>
            <div v-if="selectedGoods" class="mt-2">
              <div class="flex items-start justify-between gap-3">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ goodsTitle(selectedGoods) }}</h2>
                <div class="shrink-0 text-right">
                  <div class="font-mono text-lg font-extrabold text-gray-900 dark:text-white">CNY {{ selectedGoods.price.toFixed(2) }}</div>
                  <div v-if="shouldShowReferencePrice(selectedGoods)" class="font-mono text-xs text-gray-400 line-through">CNY {{ selectedGoods.reference_price!.toFixed(2) }}</div>
                </div>
              </div>
              <p v-if="goodsDescription(selectedGoods)" class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
                {{ goodsDescription(selectedGoods) }}
              </p>
              <div v-if="selectedGoods.tags?.length" class="mt-3 flex flex-wrap gap-1.5">
                <span
                  v-for="tag in selectedGoods.tags"
                  :key="tag"
                  class="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/20 dark:text-primary-300"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>

          <div class="space-y-4">
            <label class="block">
              <span class="input-label">{{ t('payment.card.quantity') }}</span>
              <input v-model.number="quantity" class="input" type="number" min="1" :max="quantityMax || undefined" />
              <span v-if="quantityError" class="mt-1 block text-xs text-amber-600 dark:text-amber-300">{{ quantityError }}</span>
            </label>
            <label class="block">
              <span class="input-label">{{ t('payment.card.coupon') }}</span>
              <input v-model="couponCode" class="input" type="text" />
              <span
                v-if="couponMessage"
                class="mt-1 block text-xs"
                :class="couponStatus === 'valid'
                  ? 'text-emerald-600 dark:text-emerald-300'
                  : couponStatus === 'invalid'
                    ? 'text-red-600 dark:text-red-300'
                    : 'text-gray-500 dark:text-gray-400'"
              >
                {{ couponMessage }}
              </span>
            </label>

            <div>
              <span class="input-label">{{ t('payment.card.channel') }}</span>
              <div class="grid gap-2">
                <button
                  v-for="channel in enabledChannels"
                  :key="channel.id"
                  type="button"
                  class="flex min-h-[52px] items-center gap-3 rounded-lg border px-3 py-2 text-left transition"
                  :class="channelId === channel.id
                    ? 'border-primary-500 bg-primary-50 ring-2 ring-primary-100 dark:border-primary-400 dark:bg-primary-900/20 dark:ring-primary-900/40'
                    : 'border-gray-200 hover:border-gray-300 dark:border-dark-700'"
                  @click="channelId = channel.id"
                >
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md bg-gray-100 dark:bg-dark-700">
                    <img v-if="channelIconUrl(channel.icon)" :src="channelIconUrl(channel.icon)" alt="" class="h-5 w-5 object-contain" />
                    <span v-else class="text-xs font-semibold text-gray-500">{{ channelInitial(channel) }}</span>
                  </span>
                  <span class="min-w-0">
                    <span class="block truncate text-sm font-medium text-gray-900 dark:text-white">{{ channel.show_name || channel.name }}</span>
                    <span v-if="channel.code" class="block truncate text-xs text-gray-500 dark:text-gray-400">{{ channel.code }}</span>
                  </span>
                </button>
              </div>
            </div>

            <div class="flex items-center justify-between gap-4 rounded-lg border border-gray-100 p-3 dark:border-dark-700">
              <span>
                <span class="block text-sm font-medium text-gray-900 dark:text-white">{{ t('payment.card.autoRedeem') }}</span>
                <span class="mt-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('payment.card.autoRedeemHint') }}</span>
              </span>
              <button
                type="button"
                role="switch"
                :aria-checked="autoRedeem"
                class="relative h-7 w-12 shrink-0 rounded-full transition"
                :class="autoRedeem ? 'bg-primary-600' : 'bg-gray-300 dark:bg-dark-600'"
                @click="autoRedeem = !autoRedeem"
              >
                <span class="absolute top-1 h-5 w-5 rounded-full bg-white shadow transition" :class="autoRedeem ? 'left-6' : 'left-1'" />
              </button>
            </div>
          </div>

          <div class="mt-4 flex items-center justify-between gap-3 border-t border-gray-100 pt-4 dark:border-dark-700">
            <div class="min-w-0 text-sm text-gray-600 dark:text-gray-300">
              <span>{{ t('payment.card.payAmount') }}</span>
              <span class="ml-2 text-xl font-semibold text-gray-900 dark:text-white">¥{{ totalAmount.toFixed(2) }}</span>
              <span class="ml-2 whitespace-nowrap text-xs text-gray-400 dark:text-gray-500">
                ({{ t('payment.card.feeIncluded', { fee: feeAmount.toFixed(2) }) }})
              </span>
            </div>
            <button class="btn btn-primary" :disabled="submitting || !canSubmit" @click="createOrder">
              {{ submitting ? t('common.processing') : t('payment.card.buy') }}
            </button>
          </div>
        </div>
      </aside>
    </div>

    <div v-if="editingGoods" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="w-full max-w-lg rounded-lg bg-white p-5 shadow-xl dark:bg-dark-800">
        <div class="mb-4 flex items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.card.editGoods') }}</h2>
          <button class="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200" @click="editingGoods = null">
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="space-y-4">
          <label class="block">
            <span class="input-label">{{ t('payment.card.goodsTitle') }}</span>
            <input v-model="editForm.title" class="input" type="text" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('payment.card.goodsDescription') }}</span>
            <textarea v-model="editForm.description" class="input min-h-[96px]" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('payment.card.goodsBadge') }}</span>
            <input v-model="editForm.badge" class="input" type="text" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('payment.card.goodsTags') }}</span>
            <input v-model="editForm.tagsText" class="input" type="text" :placeholder="t('payment.card.goodsTagsPlaceholder')" />
          </label>
        </div>
        <div class="mt-5 flex justify-end gap-2">
          <button class="btn btn-secondary" :disabled="savingGoods" @click="editingGoods = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="savingGoods" @click="saveGoodsOverride">
            {{ savingGoods ? t('common.processing') : t('payment.card.saveGoods') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminPaymentAPI } from '@/api/admin/payment'
import { paymentAPI } from '@/api/payment'
import Icon from '@/components/icons/Icon.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { CardCheckoutShop, CardGoods, CardPrice, CreateOrderResult } from '@/types/payment'
import { extractApiErrorMessage } from '@/utils/apiError'

defineProps<{ embedded?: boolean }>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const loading = ref(false)
const submitting = ref(false)
const savingGoods = ref(false)
const shops = ref<CardCheckoutShop[]>([])
const selectedProviderId = ref('')
const selectedCategoryId = ref(0)
const selectedGoodsKey = ref('')
const quantity = ref(1)
const channelId = ref(0)
const couponCode = ref('')
const couponStatus = ref<'idle' | 'checking' | 'valid' | 'invalid'>('idle')
const couponMessage = ref('')
const autoRedeem = ref(true)
const price = ref<CardPrice | null>(null)
const createdOrder = ref<CreateOrderResult | null>(null)
const editingGoods = ref<CardGoods | null>(null)
const editForm = reactive({
  title: '',
  description: '',
  badge: '',
  tagsText: '',
})

const activeChipClass = 'bg-primary-600 text-white'
const idleChipClass = 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300'

const selectedShop = computed(() => shops.value.find(shop => shop.provider_instance_id === selectedProviderId.value) || shops.value[0] || null)
const filteredGoods = computed(() => {
  const goods = selectedShop.value?.goods || []
  if (!selectedCategoryId.value) return goods
  return goods.filter(item => item.category_id === selectedCategoryId.value)
})
const selectedGoods = computed<CardGoods | null>(() => selectedShop.value?.goods.find(item => item.goods_key === selectedGoodsKey.value) || null)
const enabledChannels = computed(() => (selectedShop.value?.channels || []).filter(channel => channel.status === 1))
const trimmedCouponCode = computed(() => couponCode.value.trim())
const totalAmount = computed(() => price.value?.total_amount ?? (selectedGoods.value ? selectedGoods.value.price * Math.max(1, quantity.value || 1) : 0))
const feeAmount = computed(() => price.value?.fee ?? 0)
const quantityMax = computed(() => {
  if (!selectedGoods.value) return 0
  const limits = [selectedGoods.value.stock_count]
  if (selectedGoods.value.limit_count > 0) limits.push(selectedGoods.value.limit_count)
  return Math.max(0, Math.min(...limits))
})
const stockEnough = computed(() => !!selectedGoods.value && selectedGoods.value.stock_count > 0 && quantity.value <= selectedGoods.value.stock_count)
const limitEnough = computed(() => !selectedGoods.value || selectedGoods.value.limit_count <= 0 || quantity.value <= selectedGoods.value.limit_count)
const quantityError = computed(() => {
  if (!selectedGoods.value || quantity.value <= 0) return ''
  if (selectedGoods.value.stock_count <= 0) return t('payment.card.outOfStock')
  if (!stockEnough.value) return t('payment.card.stockNotEnough')
  if (!limitEnough.value) return t('payment.card.limitExceeded')
  return ''
})
const canPreviewPrice = computed(() => !!selectedShop.value && !!selectedGoods.value && quantity.value > 0 && channelId.value > 0 && stockEnough.value && limitEnough.value)
const couponReady = computed(() => !trimmedCouponCode.value || couponStatus.value === 'valid')
const canSubmit = computed(() => canPreviewPrice.value && couponReady.value && couponStatus.value !== 'checking')
let pricePreviewSeq = 0
let pricePreviewTimer: ReturnType<typeof setTimeout> | undefined

function unwrapAPI<T>(value: T | { data: T }): T {
  return value && typeof value === 'object' && 'data' in value ? (value as { data: T }).data : value as T
}

async function loadCheckoutInfo() {
  loading.value = true
  try {
    const data = unwrapAPI(await paymentAPI.getCardCheckoutInfo())
    shops.value = data.shops || []
    if (shops.value.length) {
      const currentStillExists = shops.value.some(shop => shop.provider_instance_id === selectedProviderId.value)
      selectShop(currentStillExists ? selectedProviderId.value : shops.value[0].provider_instance_id)
    } else {
      selectedProviderId.value = ''
      selectedGoodsKey.value = ''
      channelId.value = 0
    }
  } finally {
    loading.value = false
  }
}

function selectShop(providerId: string) {
  selectedProviderId.value = providerId
  selectedCategoryId.value = 0
  selectedGoodsKey.value = selectedShop.value?.goods.find(goods => goods.stock_count > 0)?.goods_key || selectedShop.value?.goods[0]?.goods_key || ''
  channelId.value = enabledChannels.value[0]?.id || 0
}

function selectGoods(goods: CardGoods) {
  if (goods.stock_count <= 0) return
  selectedGoodsKey.value = goods.goods_key
}

function goodsTitle(goods: CardGoods) {
  return goods.display_title || goods.name
}

function goodsDescription(goods: CardGoods) {
  return goods.display_description || goods.description || ''
}

function shouldShowReferencePrice(goods: CardGoods) {
  return typeof goods.reference_price === 'number' && goods.reference_price > goods.price
}

function channelIconUrl(icon?: string) {
  const raw = String(icon || '').trim()
  if (!raw) return ''
  if (raw.startsWith('//')) return `https:${raw}`
  if (/^https?:\/\//i.test(raw)) return raw
  return raw
}

function channelInitial(channel: { show_name?: string; name: string }) {
  return (channel.show_name || channel.name || '?').trim().slice(0, 1).toUpperCase()
}

function stockStatusLabel(stock: number) {
  if (stock >= 20) return t('payment.card.stockPlenty')
  if (stock > 5) return t('payment.card.stockNormal')
  if (stock > 0) return t('payment.card.stockLow')
  return t('payment.card.stockEmpty')
}

function openEditDialog(goods: CardGoods) {
  editingGoods.value = goods
  editForm.title = goods.display_title || goods.name
  editForm.description = goods.display_description || goods.description || ''
  editForm.badge = goods.badge || ''
  editForm.tagsText = (goods.tags || []).join(', ')
}

function parseTags(text: string) {
  const seen = new Set<string>()
  return text.split(/[,，]/)
    .map(tag => tag.trim())
    .filter(tag => {
      if (!tag || seen.has(tag)) return false
      seen.add(tag)
      return true
    })
}

async function saveGoodsOverride() {
  if (!editingGoods.value || !selectedShop.value) return
  savingGoods.value = true
  try {
    await adminPaymentAPI.updateCardGoodsOverride({
      provider_instance_id: selectedShop.value.provider_instance_id,
      goods_key: editingGoods.value.goods_key,
      title: editForm.title,
      description: editForm.description,
      badge: editForm.badge,
      tags: parseTags(editForm.tagsText),
    })
    appStore.showSuccess(t('payment.card.goodsSaved'))
    editingGoods.value = null
    await loadCheckoutInfo()
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : t('common.error'))
  } finally {
    savingGoods.value = false
  }
}

async function previewPrice() {
  const seq = ++pricePreviewSeq
  const coupon = trimmedCouponCode.value
  if (!coupon) {
    couponStatus.value = 'idle'
    couponMessage.value = ''
  } else {
    couponStatus.value = 'checking'
    couponMessage.value = t('payment.card.couponChecking')
  }

  if (!canPreviewPrice.value || !selectedGoods.value) {
    price.value = null
    if (coupon && seq === pricePreviewSeq) {
      couponStatus.value = 'idle'
      couponMessage.value = ''
    }
    return
  }
  try {
    const nextPrice = unwrapAPI(await paymentAPI.getCardPrice({
      provider_instance_id: selectedProviderId.value,
      goods_key: selectedGoods.value.goods_key,
      quantity: Math.max(1, quantity.value || 1),
      channel_id: channelId.value,
      coupon_code: coupon,
    }))
    if (seq !== pricePreviewSeq) return
    price.value = nextPrice
    if (coupon) {
      if (nextPrice.coupon_available === 1) {
        couponStatus.value = 'valid'
        couponMessage.value = t('payment.card.couponValid', { amount: (nextPrice.coupon_price || 0).toFixed(2) })
      } else {
        couponStatus.value = 'invalid'
        couponMessage.value = t('payment.card.couponInvalid')
      }
    }
  } catch (error) {
    if (seq !== pricePreviewSeq) return
    price.value = null
    if (coupon) {
      couponStatus.value = 'invalid'
      couponMessage.value = extractApiErrorMessage(error, t('payment.card.couponInvalid'))
    }
  }
}

async function createOrder() {
  if (!canSubmit.value || !selectedGoods.value) return
  submitting.value = true
  try {
    const order = unwrapAPI(await paymentAPI.createCardOrder({
      provider_instance_id: selectedProviderId.value,
      goods_key: selectedGoods.value.goods_key,
      quantity: Math.max(1, quantity.value || 1),
      channel_id: channelId.value,
      coupon_code: trimmedCouponCode.value,
      auto_redeem: autoRedeem.value,
      return_url: window.location.href,
    }))
    createdOrder.value = order
    if (order.pay_url && order.pay_amount > 0) {
      window.open(order.pay_url, '_blank', 'noopener,noreferrer')
    }
  } finally {
    submitting.value = false
  }
}

watch([selectedProviderId, selectedGoodsKey, quantity, channelId, couponCode], () => {
  if (pricePreviewTimer) {
    clearTimeout(pricePreviewTimer)
  }
  pricePreviewTimer = setTimeout(() => {
    void previewPrice()
  }, 350)
})

watch(selectedGoods, () => {
  if (quantityMax.value > 0 && quantity.value > quantityMax.value) {
    quantity.value = quantityMax.value
  }
  if (quantity.value <= 0) {
    quantity.value = 1
  }
})

onMounted(loadCheckoutInfo)

onUnmounted(() => {
  if (pricePreviewTimer) {
    clearTimeout(pricePreviewTimer)
  }
})
</script>
