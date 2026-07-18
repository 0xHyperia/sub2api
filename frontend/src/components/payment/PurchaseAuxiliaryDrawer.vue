<template>
  <Teleport to="body">
    <Transition name="purchase-tools-drawer">
      <div
        v-if="mode"
        class="fixed inset-0 z-50 flex items-end bg-black/25 backdrop-blur-[1px] md:items-stretch md:justify-end"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="emit('close')"
      >
        <aside
          ref="drawerPanel"
          tabindex="-1"
          class="flex max-h-[88dvh] w-full min-h-0 flex-col overflow-hidden rounded-t-panel border border-b-0 border-outline bg-surface-raised shadow-floating outline-none md:h-full md:max-h-none md:max-w-[420px] md:rounded-none md:border-b md:border-r-0"
        >
          <header class="flex min-h-16 shrink-0 items-center gap-3 border-b border-outline px-4 sm:px-5">
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control bg-surface-subtle text-foreground">
              <Icon name="gift" size="sm" aria-hidden="true" />
            </span>
            <div class="min-w-0 flex-1">
              <h2 :id="titleId" class="text-base font-semibold text-foreground">{{ drawerTitle }}</h2>
              <p class="mt-0.5 text-xs text-foreground-subtle">{{ t('purchaseWorkspace.toolsDescription') }}</p>
            </div>
            <button
              ref="closeButton"
              type="button"
              class="btn btn-ghost btn-icon shrink-0"
              :aria-label="t('common.close')"
              @click="emit('close')"
            >
              <Icon name="x" size="md" aria-hidden="true" />
            </button>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-5">
            <div class="space-y-4">
              <AffiliateRewardPanel
                v-if="showAffiliate"
                data-testid="purchase-drawer-affiliate"
              />

              <section
                v-if="showRedeem"
                data-testid="purchase-drawer-redeem"
                class="rounded-panel border border-outline bg-surface p-4 shadow-card sm:p-5"
              >
                <RedeemCodeForm :show-description="mode === 'combined'" />
              </section>
            </div>
          </div>
        </aside>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import AffiliateRewardPanel from './AffiliateRewardPanel.vue'
import RedeemCodeForm from './RedeemCodeForm.vue'

export type PurchaseAuxiliaryMode = 'redeem' | 'affiliate' | 'combined'

const props = defineProps<{
  mode: PurchaseAuxiliaryMode | null
  affiliateEnabled: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const titleId = 'purchase-auxiliary-drawer-title'
const closeButton = ref<HTMLButtonElement | null>(null)
const drawerPanel = ref<HTMLElement | null>(null)
let previousFocus: HTMLElement | null = null
let previousBodyOverflow = ''

const showRedeem = computed(() => props.mode === 'redeem' || props.mode === 'combined')
const showAffiliate = computed(() => props.affiliateEnabled && (props.mode === 'affiliate' || props.mode === 'combined'))
const drawerTitle = computed(() => {
  if (props.mode === 'redeem') return t('redeem.redeemCodeLabel')
  if (props.mode === 'affiliate') return t('purchaseWorkspace.affiliateTitle')
  return t('purchaseWorkspace.toolsTitle')
})

function handleKeydown(event: KeyboardEvent): void {
  if (!props.mode) return
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }
  if (event.key !== 'Tab' || !drawerPanel.value) return

  const focusable = [...drawerPanel.value.querySelectorAll<HTMLElement>(
    'button:not([disabled]),a[href],input:not([disabled]),select:not([disabled]),[tabindex]:not([tabindex="-1"])',
  )].filter(element => getComputedStyle(element).visibility !== 'hidden')
  if (!focusable.length) {
    event.preventDefault()
    drawerPanel.value.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(() => props.mode, async (mode) => {
  if (mode) {
    previousFocus = document.activeElement as HTMLElement
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleKeydown)
    await nextTick()
    closeButton.value?.focus()
    return
  }

  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
  previousFocus?.focus()
  previousFocus = null
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousBodyOverflow
})
</script>

<style scoped>
.purchase-tools-drawer-enter-active,
.purchase-tools-drawer-leave-active {
  transition: opacity 180ms ease;
}

.purchase-tools-drawer-enter-active aside,
.purchase-tools-drawer-leave-active aside {
  transition: transform 180ms ease;
}

.purchase-tools-drawer-enter-from,
.purchase-tools-drawer-leave-to {
  opacity: 0;
}

.purchase-tools-drawer-enter-from aside,
.purchase-tools-drawer-leave-to aside {
  transform: translateY(100%);
}

@media (min-width: 768px) {
  .purchase-tools-drawer-enter-from aside,
  .purchase-tools-drawer-leave-to aside {
    transform: translateX(100%);
  }
}

@media (prefers-reduced-motion: reduce) {
  .purchase-tools-drawer-enter-active,
  .purchase-tools-drawer-leave-active,
  .purchase-tools-drawer-enter-active aside,
  .purchase-tools-drawer-leave-active aside {
    transition: none;
  }
}
</style>
