<template>
  <Teleport to="body">
    <div v-if="show && position">
      <div
        class="fixed inset-0 z-[9998]"
        aria-hidden="true"
        @click="requestClose(true)"
      />
      <div
        ref="menuRef"
        role="menu"
        :aria-label="menuLabel"
        class="action-menu-content fixed z-[9999] w-52 max-w-[calc(100vw-1rem)] overflow-y-auto rounded-panel border border-outline bg-surface-raised p-1 shadow-floating"
        :style="menuStyle"
        @click.stop
        @keydown="handleMenuKeydown"
      >
        <template v-if="account">
          <button
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('test')"
          >
            <Icon name="play" size="sm" class="text-foreground-subtle" :stroke-width="2" />
            {{ t('admin.accounts.testConnection') }}
          </button>
          <button
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('stats')"
          >
            <Icon name="chart" size="sm" class="text-foreground-subtle" />
            {{ t('admin.accounts.viewStats') }}
          </button>
          <button
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('schedule')"
          >
            <Icon name="clock" size="sm" class="text-foreground-subtle" />
            {{ t('admin.scheduledTests.schedule') }}
          </button>
          <button
            v-if="canDuplicate"
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('duplicate')"
          >
            <Icon name="copy" size="sm" class="text-foreground-subtle" />
            {{ t('admin.accounts.duplicateAccount') }}
          </button>

          <!-- Shadow accounts do not hold credentials, so credential actions are unavailable. -->
          <template v-if="(account.type === 'oauth' || account.type === 'setup-token') && !isShadow">
            <button
              type="button"
              role="menuitem"
              tabindex="-1"
              class="dropdown-item w-full text-left"
              @click="selectAction('reauth')"
            >
              <Icon name="link" size="sm" class="text-foreground-subtle" />
              {{ t('admin.accounts.reAuthorize') }}
            </button>
            <button
              type="button"
              role="menuitem"
              tabindex="-1"
              class="dropdown-item w-full text-left"
              @click="selectAction('refresh-token')"
            >
              <Icon name="refresh" size="sm" class="text-foreground-subtle" />
              {{ t('admin.accounts.refreshToken') }}
            </button>
          </template>
          <button
            v-if="isOpenAIOAuthParent"
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('create-spark-shadow')"
          >
            <Icon name="sparkles" size="sm" class="text-foreground-subtle" />
            {{ t('admin.accounts.createSparkShadow') }}
          </button>
          <button
            v-if="supportsPrivacy"
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('set-privacy')"
          >
            <Icon name="shield" size="sm" class="text-foreground-subtle" />
            {{ t('admin.accounts.setPrivacy') }}
          </button>
          <div v-if="hasRecoverableState" role="separator" class="my-1 border-t border-outline" />
          <button
            v-if="hasRecoverableState"
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('recover-state')"
          >
            <Icon name="sync" size="sm" class="text-success" />
            {{ t('admin.accounts.recoverState') }}
          </button>
          <button
            v-if="hasQuotaLimit"
            type="button"
            role="menuitem"
            tabindex="-1"
            class="dropdown-item w-full text-left"
            @click="selectAction('reset-quota')"
          >
            <Icon name="refresh" size="sm" class="text-foreground-subtle" />
            {{ t('admin.accounts.resetQuota') }}
          </button>
        </template>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import type { Account } from '@/types'

type AccountAction =
  | 'test'
  | 'stats'
  | 'schedule'
  | 'duplicate'
  | 'reauth'
  | 'refresh-token'
  | 'recover-state'
  | 'reset-quota'
  | 'set-privacy'
  | 'create-spark-shadow'

const props = defineProps<{
  show: boolean
  account: Account | null
  position: { top: number; left: number } | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'test', account: Account): void
  (e: 'stats', account: Account): void
  (e: 'schedule', account: Account): void
  (e: 'duplicate', account: Account): void
  (e: 'reauth', account: Account): void
  (e: 'refresh-token', account: Account): void
  (e: 'recover-state', account: Account): void
  (e: 'reset-quota', account: Account): void
  (e: 'set-privacy', account: Account): void
  (e: 'create-spark-shadow', account: Account): void
}>()

const { t } = useI18n()
const menuRef = ref<HTMLElement | null>(null)
const viewportWidth = ref(typeof window === 'undefined' ? 1024 : window.innerWidth)
const viewportHeight = ref(typeof window === 'undefined' ? 768 : window.innerHeight)
let previousActiveElement: HTMLElement | null = null
let restoreFocusOnClose = true

const menuItems = () =>
  Array.from(menuRef.value?.querySelectorAll<HTMLElement>('[role="menuitem"]:not([disabled])') ?? [])

const menuLabel = computed(() => {
  const accountName = props.account?.name
  return accountName ? `${accountName}: ${t('common.more')}` : t('common.more')
})

const menuStyle = computed(() => {
  if (!props.position) return undefined

  const padding = 8
  const width = Math.min(208, Math.max(0, viewportWidth.value - padding * 2))
  const left = Math.max(
    padding,
    Math.min(props.position.left, viewportWidth.value - width - padding)
  )
  const top = Math.max(padding, Math.min(props.position.top, viewportHeight.value - 88))

  return {
    left: `${left}px`,
    top: `${top}px`,
    maxHeight: `${Math.max(80, viewportHeight.value - top - padding)}px`
  }
})

const canDuplicate = computed(() => {
  if (!props.account || props.account.parent_account_id != null) return false
  return ['apikey', 'upstream', 'bedrock', 'service_account'].includes(props.account.type)
})
const isRateLimited = computed(() => {
  if (props.account?.rate_limit_reset_at && new Date(props.account.rate_limit_reset_at) > new Date()) {
    return true
  }
  const modelLimits = (props.account?.extra as Record<string, unknown> | undefined)?.model_rate_limits as
    | Record<string, { rate_limit_reset_at: string }>
    | undefined
  if (modelLimits) {
    const now = new Date()
    return Object.values(modelLimits).some(info => new Date(info.rate_limit_reset_at) > now)
  }
  return false
})
const isOverloaded = computed(() => props.account?.overload_until && new Date(props.account.overload_until) > new Date())
const isTempUnschedulable = computed(() => props.account?.temp_unschedulable_until && new Date(props.account.temp_unschedulable_until) > new Date())
const hasRecoverableState = computed(() => {
  return props.account?.status === 'error' || Boolean(isRateLimited.value) || Boolean(isOverloaded.value) || Boolean(isTempUnschedulable.value)
})
const isAntigravityOAuth = computed(() => props.account?.platform === 'antigravity' && props.account?.type === 'oauth')
const isOpenAIOAuth = computed(() => props.account?.platform === 'openai' && props.account?.type === 'oauth')
const isShadow = computed(() => props.account?.parent_account_id != null)
const isOpenAIOAuthParent = computed(() => isOpenAIOAuth.value && !isShadow.value)
const supportsPrivacy = computed(() => (isAntigravityOAuth.value || isOpenAIOAuth.value) && !isShadow.value)
const hasQuotaLimit = computed(() => {
  return (props.account?.type === 'apikey' || props.account?.type === 'bedrock') && (
    (props.account?.quota_limit ?? 0) > 0 ||
    (props.account?.quota_daily_limit ?? 0) > 0 ||
    (props.account?.quota_weekly_limit ?? 0) > 0
  )
})

const requestClose = (restoreFocus: boolean) => {
  restoreFocusOnClose = restoreFocus
  emit('close')
}

const selectAction = (action: AccountAction) => {
  const account = props.account
  if (!account) return

  restoreFocusOnClose = false
  switch (action) {
    case 'test': emit('test', account); break
    case 'stats': emit('stats', account); break
    case 'schedule': emit('schedule', account); break
    case 'duplicate': emit('duplicate', account); break
    case 'reauth': emit('reauth', account); break
    case 'refresh-token': emit('refresh-token', account); break
    case 'recover-state': emit('recover-state', account); break
    case 'reset-quota': emit('reset-quota', account); break
    case 'set-privacy': emit('set-privacy', account); break
    case 'create-spark-shadow': emit('create-spark-shadow', account); break
  }
  emit('close')
}

const moveFocusAfterTrigger = async (backward: boolean) => {
  const trigger = previousActiveElement
  const selector = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled]):not([type="hidden"])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])'
  ].join(',')
  const focusable = Array.from(document.querySelectorAll<HTMLElement>(selector)).filter(element => {
    if (menuRef.value?.contains(element) || element.closest('[hidden], [aria-hidden="true"]')) {
      return false
    }
    const style = window.getComputedStyle(element)
    return style.display !== 'none' && style.visibility !== 'hidden'
  })
  const currentIndex = trigger ? focusable.indexOf(trigger) : -1
  const targetIndex = backward ? currentIndex - 1 : currentIndex + 1
  const target = focusable[targetIndex]

  requestClose(false)
  await nextTick()
  target?.focus()
}

const handleMenuKeydown = (event: KeyboardEvent) => {
  const items = menuItems()
  const currentIndex = items.indexOf(event.target as HTMLElement)

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      items[currentIndex < 0 ? 0 : (currentIndex + 1) % items.length]?.focus()
      break
    case 'ArrowUp':
      event.preventDefault()
      items[currentIndex < 0 ? items.length - 1 : (currentIndex - 1 + items.length) % items.length]?.focus()
      break
    case 'Home':
      event.preventDefault()
      items[0]?.focus()
      break
    case 'End':
      event.preventDefault()
      items[items.length - 1]?.focus()
      break
    case 'Escape':
      event.preventDefault()
      event.stopPropagation()
      requestClose(true)
      break
    case 'Tab':
      event.preventDefault()
      void moveFocusAfterTrigger(event.shiftKey)
      break
  }
}

const updateViewport = () => {
  viewportWidth.value = window.innerWidth
  viewportHeight.value = window.innerHeight
}

watch(
  () => props.show && Boolean(props.position),
  async (visible, wasVisible) => {
    if (visible) {
      previousActiveElement = document.activeElement as HTMLElement | null
      restoreFocusOnClose = true
      await nextTick()
      menuItems()[0]?.focus()
      return
    }

    if (wasVisible && restoreFocusOnClose && previousActiveElement?.isConnected) {
      await nextTick()
      previousActiveElement.focus()
    }
    previousActiveElement = null
    restoreFocusOnClose = true
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener('resize', updateViewport)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewport)
})
</script>
