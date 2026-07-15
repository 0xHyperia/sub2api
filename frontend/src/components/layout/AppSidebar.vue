<template>
  <aside
    id="app-sidebar"
    ref="sidebarRef"
    class="admin-sidebar"
    :class="[
      { 'admin-sidebar-collapsed': sidebarCollapsed },
      mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
    ]"
    :role="isDesktopViewport ? undefined : 'dialog'"
    :aria-modal="isDesktopViewport ? undefined : true"
    :aria-label="localText('主导航', 'Primary navigation')"
    :aria-hidden="isDesktopViewport ? undefined : !mobileOpen"
    :inert="!isDesktopViewport && !mobileOpen"
    tabindex="-1"
    @keydown="handleDrawerKeydown"
  >
    <div class="admin-sidebar-header">
      <router-link
        :to="homePath"
        class="sidebar-logo"
        :aria-label="siteName"
        @click="handleMenuItemClick(homePath)"
      >
        <img
          v-if="settingsLoaded"
          :src="siteLogo || '/logo.png'"
          alt=""
          class="h-full w-full object-contain"
        >
      </router-link>

      <div
        class="sidebar-brand"
        :class="{ 'sidebar-brand-collapsed': sidebarCollapsed }"
        :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
        :inert="sidebarCollapsed"
      >
        <router-link
          :to="homePath"
          class="sidebar-brand-title"
          @click="handleMenuItemClick(homePath)"
        >
          {{ siteName }}
        </router-link>
        <div class="mt-0.5 flex items-center gap-2">
          <span v-if="showAdminNavigation" class="admin-context-label">Admin</span>
          <VersionBadge :version="siteVersion" />
        </div>
      </div>

      <button
        ref="mobileCloseRef"
        type="button"
        class="sidebar-header-action lg:hidden"
        :aria-label="localText('关闭导航', 'Close navigation')"
        @click="closeMobile"
      >
        <Icon name="x" size="md" />
      </button>
    </div>

    <nav
      ref="sidebarNavRef"
      class="admin-sidebar-nav"
      :aria-label="localText('后台导航', 'Admin navigation')"
    >
      <template v-if="showAdminNavigation">
        <section
          v-for="section in adminNavSections"
          :key="section.id"
          class="sidebar-section"
          :aria-labelledby="`sidebar-section-${section.id}`"
        >
          <h2
            :id="`sidebar-section-${section.id}`"
            class="sidebar-section-title"
            :class="{ 'sidebar-section-title-collapsed': sidebarCollapsed }"
            :title="sidebarCollapsed ? section.label : undefined"
          >
            <span>{{ section.label }}</span>
          </h2>

          <template v-for="item in section.items" :key="item.path">
            <div v-if="item.children?.length" class="sidebar-group">
              <button
                type="button"
                class="sidebar-link w-full"
                :class="{
                  'sidebar-group-active': isGroupActive(item) && !sidebarCollapsed,
                  'sidebar-link-active': isGroupActive(item) && sidebarCollapsed,
                  'sidebar-link-collapsed': sidebarCollapsed
                }"
                :title="sidebarCollapsed ? item.label : undefined"
                :aria-label="sidebarCollapsed ? item.label : undefined"
                :aria-expanded="!sidebarCollapsed && isGroupExpanded(item)"
                :aria-controls="`sidebar-group-${section.id}-${slugify(item.path)}`"
                @click="handleGroupClick(item)"
              >
                <Icon :name="item.icon || 'document'" size="md" class="sidebar-link-icon" />
                <span
                  class="sidebar-label sidebar-group-label"
                  :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                  :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
                >
                  <span class="truncate">{{ item.label }}</span>
                  <Icon
                    name="chevronDown"
                    size="xs"
                    class="transition-transform duration-150"
                    :class="isGroupExpanded(item) ? 'rotate-180' : ''"
                  />
                </span>
              </button>

              <div
                v-if="!sidebarCollapsed && isGroupExpanded(item)"
                :id="`sidebar-group-${section.id}-${slugify(item.path)}`"
                class="sidebar-children"
              >
                <router-link
                  v-for="child in item.children"
                  :key="child.path"
                  :to="child.path"
                  class="sidebar-child-link"
                  :class="{ 'sidebar-child-link-active': route.path === child.path }"
                  :aria-current="route.path === child.path ? 'page' : undefined"
                  @click="handleMenuItemClick(child.path)"
                >
                  <Icon :name="child.icon || 'document'" size="sm" class="flex-shrink-0" />
                  <span class="truncate">{{ child.label }}</span>
                </router-link>
              </div>
            </div>

            <router-link
              v-else
              :to="item.path"
              class="sidebar-link"
              :class="{
                'sidebar-link-active': isActive(item.path),
                'sidebar-link-collapsed': sidebarCollapsed
              }"
              :title="sidebarCollapsed ? item.label : undefined"
              :aria-label="sidebarCollapsed ? item.label : undefined"
              :id="tourId(item.path)"
              :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
              :aria-current="isActive(item.path) ? 'page' : undefined"
              @click="handleMenuItemClick(item.path)"
            >
              <span
                v-if="item.iconSvg"
                class="sidebar-link-icon sidebar-svg-icon"
                aria-hidden="true"
                v-html="sanitizeSvg(item.iconSvg)"
              ></span>
              <Icon v-else :name="item.icon || 'document'" size="md" class="sidebar-link-icon" />
              <span
                class="sidebar-label"
                :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
                :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
              >
                {{ item.label }}
              </span>
            </router-link>
          </template>
        </section>
      </template>

      <section v-else-if="isAdmin || !appStore.backendModeEnabled" class="sidebar-section">
        <router-link
          v-for="item in userNavItems"
          :key="item.path"
          :to="item.path"
          class="sidebar-link"
          :class="{
            'sidebar-link-active': isActive(item.path),
            'sidebar-link-collapsed': sidebarCollapsed
          }"
          :title="sidebarCollapsed ? item.label : undefined"
          :aria-label="sidebarCollapsed ? item.label : undefined"
          :data-tour="item.path === '/keys' ? 'sidebar-my-keys' : undefined"
          :aria-current="isActive(item.path) ? 'page' : undefined"
          @click="handleMenuItemClick(item.path)"
        >
          <span
            v-if="item.iconSvg"
            class="sidebar-link-icon sidebar-svg-icon"
            aria-hidden="true"
            v-html="sanitizeSvg(item.iconSvg)"
          ></span>
          <Icon v-else :name="item.icon || 'document'" size="md" class="sidebar-link-icon" />
          <span
            class="sidebar-label"
            :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
            :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
          >
            {{ item.label }}
          </span>
        </router-link>
      </section>
    </nav>

    <div class="admin-sidebar-footer">
      <button
        type="button"
        class="sidebar-link hidden w-full lg:flex"
        :class="{ 'sidebar-link-collapsed': sidebarCollapsed }"
        :title="sidebarCollapsed ? t('nav.expand') : t('nav.collapse')"
        :aria-label="sidebarCollapsed ? t('nav.expand') : t('nav.collapse')"
        @click="toggleSidebar"
      >
        <Icon :name="sidebarCollapsed ? 'chevronRight' : 'chevronLeft'" size="md" class="sidebar-link-icon" />
        <span
          class="sidebar-label"
          :class="{ 'sidebar-label-collapsed': sidebarCollapsed }"
          :aria-hidden="sidebarCollapsed ? 'true' : 'false'"
        >
          {{ t('nav.collapse') }}
        </span>
      </button>
    </div>
  </aside>

  <Transition name="drawer-backdrop">
    <button
      v-if="mobileOpen && !isDesktopViewport"
      type="button"
      class="drawer-backdrop"
      :aria-label="localText('关闭导航', 'Close navigation')"
      @click="closeMobile"
    ></button>
  </Transition>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminSettingsStore, useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import VersionBadge from '@/components/common/VersionBadge.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeSvg } from '@/utils/sanitize'
import { sanitizeUrl } from '@/utils/url'
import { FeatureFlags, makeSidebarFlag } from '@/utils/featureFlags'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
import {
  filterNavigationItems,
  resolveGroupClickAction,
  type ShellNavItem,
  type ShellNavSection
} from './navigation'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const onboardingStore = useOnboardingStore()
const adminSettingsStore = useAdminSettingsStore()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

const sidebarRef = ref<HTMLElement | null>(null)
const mobileCloseRef = ref<HTMLButtonElement | null>(null)
const expandedGroups = ref<Set<string>>(new Set())
const isDesktopViewport = ref(true)
let desktopMediaQuery: MediaQueryList | null = null
let previousActiveElement: HTMLElement | null = null
let previousBodyOverflow = ''

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.isAdmin)
const sidebarNavRef = ref<HTMLElement | null>(null)
const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteVersion = computed(() => appStore.siteVersion)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

const flagChannelMonitor = makeSidebarFlag(FeatureFlags.channelMonitor)
const flagAvailableChannels = makeSidebarFlag(FeatureFlags.availableChannels)
const flagModelMarketplace = makeSidebarFlag(FeatureFlags.modelMarketplace)
const flagAffiliate = makeSidebarFlag(FeatureFlags.affiliate)
const flagRiskControl = makeSidebarFlag(FeatureFlags.riskControl)
const flagOpsMonitoring = () => adminSettingsStore.opsMonitoringEnabled
const flagAdminPayment = () => adminSettingsStore.paymentEnabled
const flagBatchImageAccess = () => canUseBatchImage.value
const flagPurchase = () =>
  appStore.cachedPublicSettings?.payment_enabled === true &&
  (appStore.cachedPublicSettings?.payment_instant_enabled !== false ||
    appStore.cachedPublicSettings?.payment_card_enabled === true ||
    appStore.cachedPublicSettings?.purchase_subscription_enabled === true)
const flagPaymentOrders = () =>
  appStore.cachedPublicSettings?.payment_enabled === true &&
  (appStore.cachedPublicSettings?.payment_instant_enabled !== false ||
    appStore.cachedPublicSettings?.payment_card_enabled === true)

const customMenuItemsForUser = computed(() => {
  const items = appStore.cachedPublicSettings?.custom_menu_items ?? []
  return items
    .filter((item) => item.visibility === 'user')
    .sort((a, b) => a.sort_order - b.sort_order)
})

const customMenuItemsForAdmin = computed(() =>
  adminSettingsStore.customMenuItems
    .filter((item) => item.visibility === 'admin')
    .sort((a, b) => a.sort_order - b.sort_order)
)

const isAdminCustomPage = computed(() => {
  if (route.name !== 'CustomPage') return false
  const id = String(route.params.id || '')
  return customMenuItemsForAdmin.value.some((item) => item.id === id)
})
const showAdminNavigation = computed(
  () => isAdmin.value && (route.path.startsWith('/admin') || isAdminCustomPage.value)
)
const homePath = computed(() => (showAdminNavigation.value ? '/admin/dashboard' : '/dashboard'))

function buildSelfNavItems(): ShellNavItem[] {
  return [
    { path: '/dashboard', label: t('nav.dashboard'), icon: 'home' },
    { path: '/keys', label: t('nav.apiKeys'), icon: 'key' },
    {
      path: '/batch-image',
      label: t('nav.batchImage'),
      icon: 'sparkles',
      hideInSimpleMode: true,
      featureFlag: flagBatchImageAccess
    },
    { path: '/usage', label: t('nav.usage'), icon: 'chart', hideInSimpleMode: true },
    {
      path: '/models',
      label: t('nav.modelMarketplace'),
      icon: 'grid',
      hideInSimpleMode: true,
      featureFlag: flagModelMarketplace
    },
    {
      path: '/available-channels',
      label: t('nav.availableChannels'),
      icon: 'cube',
      hideInSimpleMode: true,
      featureFlag: flagAvailableChannels
    },
    { path: '/monitor', label: t('nav.channelStatus'), icon: 'trendingUp', featureFlag: flagChannelMonitor },
    {
      path: '/subscriptions',
      label: t('nav.mySubscriptions'),
      icon: 'creditCard',
      hideInSimpleMode: true
    },
    {
      path: '/purchase',
      label: t('nav.buySubscription'),
      icon: 'dollar',
      hideInSimpleMode: true,
      featureFlag: flagPurchase
    },
    {
      path: '/orders',
      label: t('nav.myOrders'),
      icon: 'clipboard',
      hideInSimpleMode: true,
      featureFlag: flagPaymentOrders
    },
    { path: '/redeem', label: t('nav.redeem'), icon: 'gift', hideInSimpleMode: true },
    {
      path: '/affiliate',
      label: t('nav.affiliate'),
      icon: 'users',
      hideInSimpleMode: true,
      featureFlag: flagAffiliate
    },
    { path: '/profile', label: t('nav.profile'), icon: 'user' },
    ...customMenuItemsForUser.value.map((item): ShellNavItem => ({
      path: `/custom/${item.id}`,
      label: item.label,
      iconSvg: item.icon_svg
    }))
  ]
}

const userNavItems = computed(() =>
  filterNavigationItems(buildSelfNavItems(), authStore.isSimpleMode)
)

const adminNavSections = computed<ShellNavSection[]>(() => {
  const sections: ShellNavSection[] = [
    {
      id: 'overview',
      label: localText('概览', 'Overview'),
      items: [{ path: '/admin/dashboard', label: t('nav.dashboard'), icon: 'home' }]
    },
    {
      id: 'resources',
      label: localText('资源与路由', 'Resources'),
      items: [
        { path: '/admin/users', label: t('nav.users'), icon: 'users', hideInSimpleMode: true },
        { path: '/admin/groups', label: t('nav.groups'), icon: 'grid', hideInSimpleMode: true },
        {
          path: '/admin/channels',
          label: t('nav.channelManagement'),
          icon: 'cube',
          hideInSimpleMode: true,
          expandOnly: true,
          children: [
            { path: '/admin/channels/pricing', label: t('nav.channelPricing'), icon: 'dollar' },
            {
              path: '/admin/channels/monitor',
              label: t('nav.channelMonitor'),
              icon: 'trendingUp',
              featureFlag: flagChannelMonitor
            }
          ]
        },
        { path: '/admin/accounts', label: t('nav.accounts'), icon: 'globe' },
        { path: '/admin/proxies', label: t('nav.proxies'), icon: 'server' }
      ]
    },
    {
      id: 'observability',
      label: localText('可观测性与安全', 'Observability'),
      items: [
        { path: '/admin/ops', label: t('nav.ops'), icon: 'terminal', featureFlag: flagOpsMonitoring },
        { path: '/admin/usage', label: t('nav.usage'), icon: 'chart' },
        {
          path: '/admin/risk-control',
          label: t('nav.riskControl'),
          icon: 'shield',
          hideInSimpleMode: true,
          featureFlag: flagRiskControl
        }
      ]
    },
    {
      id: 'commerce',
      label: localText('商业化', 'Commerce'),
      items: [
        {
          path: '/admin/subscriptions',
          label: t('nav.subscriptions'),
          icon: 'creditCard',
          hideInSimpleMode: true
        },
        {
          path: '/admin/orders',
          label: t('nav.orderManagement'),
          icon: 'clipboard',
          hideInSimpleMode: true,
          expandOnly: true,
          featureFlag: flagAdminPayment,
          children: [
            { path: '/admin/orders/dashboard', label: t('nav.paymentDashboard'), icon: 'chart' },
            { path: '/admin/orders', label: t('nav.orderManagement'), icon: 'clipboard' },
            { path: '/admin/orders/plans', label: t('nav.paymentPlans'), icon: 'creditCard' }
          ]
        },
        { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: 'badge', hideInSimpleMode: true },
        { path: '/admin/promo-codes', label: t('nav.promoCodes'), icon: 'gift', hideInSimpleMode: true },
        {
          path: '/admin/affiliates',
          label: t('nav.affiliateManagement'),
          icon: 'users',
          hideInSimpleMode: true,
          expandOnly: true,
          featureFlag: flagAffiliate,
          children: [
            { path: '/admin/affiliates/invites', label: t('nav.affiliateInviteRecords'), icon: 'users' },
            { path: '/admin/affiliates/rebates', label: t('nav.affiliateRebateRecords'), icon: 'gift' },
            { path: '/admin/affiliates/transfers', label: t('nav.affiliateTransferRecords'), icon: 'creditCard' }
          ]
        }
      ]
    },
    {
      id: 'communication',
      label: localText('沟通', 'Communication'),
      items: [{ path: '/admin/announcements', label: t('nav.announcements'), icon: 'bell' }]
    },
    {
      id: 'system',
      label: localText('系统', 'System'),
      items: [
        ...(authStore.isSimpleMode
          ? [{ path: '/keys', label: t('nav.apiKeys'), icon: 'key' as const }]
          : []),
        { path: '/admin/settings', label: t('nav.settings'), icon: 'cog' }
      ]
    }
  ]

  const visibleSections = sections
    .map((section) => ({
      ...section,
      items: filterNavigationItems(section.items, authStore.isSimpleMode)
    }))
    .filter((section) => section.items.length > 0)

  if (customMenuItemsForAdmin.value.length > 0) {
    visibleSections.push({
      id: 'custom',
      label: localText('自定义', 'Custom'),
      items: customMenuItemsForAdmin.value.map((item): ShellNavItem => ({
        path: `/custom/${item.id}`,
        label: item.label,
        iconSvg: item.icon_svg
      }))
    })
  }

  return visibleSections
})

function slugify(path: string): string {
  return path.replace(/[^a-z0-9]+/gi, '-').replace(/^-|-$/g, '')
}

function tourId(path: string): string | undefined {
  if (path === '/admin/accounts') return 'sidebar-channel-manage'
  if (path === '/admin/groups') return 'sidebar-group-manage'
  if (path === '/admin/redeem') return 'sidebar-wallet'
  return undefined
}

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(`${path}/`)
}

function isGroupActive(item: ShellNavItem): boolean {
  return item.children?.some((child) => route.path === child.path) ?? false
}

function isGroupExpanded(item: ShellNavItem): boolean {
  return expandedGroups.value.has(item.path) || isGroupActive(item)
}

function toggleGroup(item: ShellNavItem): void {
  if (expandedGroups.value.has(item.path)) {
    expandedGroups.value.delete(item.path)
  } else {
    expandedGroups.value.add(item.path)
  }
}

async function handleGroupClick(item: ShellNavItem): Promise<void> {
  const action = resolveGroupClickAction(sidebarCollapsed.value, item.expandOnly)

  if (action === 'expand-sidebar') {
    appStore.setSidebarCollapsed(false)
    expandedGroups.value.add(item.path)
    await nextTick()
    return
  }

  if (action === 'toggle-group') {
    toggleGroup(item)
    return
  }

  if (route.path !== item.path) await router.push(item.path)
  expandedGroups.value.add(item.path)
}

function toggleSidebar(): void {
  appStore.toggleSidebar()
}

function closeMobile(): void {
  appStore.setMobileOpen(false)
}

function handleMenuItemClick(itemPath: string): void {
  if (mobileOpen.value) closeMobile()

  const pathToSelector: Record<string, string> = {
    '/admin/groups': '#sidebar-group-manage',
    '/admin/accounts': '#sidebar-channel-manage',
    '/keys': '[data-tour="sidebar-my-keys"]'
  }
  const selector = pathToSelector[itemPath]
  if (selector && onboardingStore.isCurrentStep(selector)) {
    onboardingStore.nextStep(500)
  }
}

function getFocusableElements(): HTMLElement[] {
  if (!sidebarRef.value) return []
  return Array.from(
    sidebarRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
  ).filter((element) => !element.hasAttribute('inert') && element.offsetParent !== null)
}

function handleDrawerKeydown(event: KeyboardEvent): void {
  if (isDesktopViewport.value || !mobileOpen.value || event.key !== 'Tab') return

  const focusable = getFocusableElements()
  if (focusable.length === 0) {
    event.preventDefault()
    sidebarRef.value?.focus()
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

function handleGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && mobileOpen.value && !isDesktopViewport.value) {
    event.preventDefault()
    closeMobile()
  }
}

function lockBodyScroll(): void {
  previousBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
}

function unlockBodyScroll(): void {
  document.body.style.overflow = previousBodyOverflow
}

watch(
  mobileOpen,
  async (open) => {
    if (open && !isDesktopViewport.value) {
      previousActiveElement = document.activeElement as HTMLElement | null
      lockBodyScroll()
      await nextTick()
      mobileCloseRef.value?.focus()
      return
    }

    unlockBodyScroll()
    if (!open && previousActiveElement) {
      previousActiveElement.focus()
      previousActiveElement = null
    }
  },
  { flush: 'post' }
)

watch(
  () => route.path,
  () => {
    if (mobileOpen.value) closeMobile()
  }
)

function handleDesktopMediaChange(event: MediaQueryListEvent): void {
  isDesktopViewport.value = event.matches
  if (event.matches && mobileOpen.value) closeMobile()
}

watch(
  isAdmin,
  (admin) => {
    if (admin) void adminSettingsStore.fetch()
  },
  { immediate: true }
)

onMounted(() => {
  desktopMediaQuery = window.matchMedia('(min-width: 1024px)')
  isDesktopViewport.value = desktopMediaQuery.matches
  desktopMediaQuery.addEventListener('change', handleDesktopMediaChange)
  document.addEventListener('keydown', handleGlobalKeydown)
  void refreshBatchImageAccess()
  if (appStore.sidebarScrollTop > 0 && sidebarNavRef.value) {
    void nextTick(() => {
      if (sidebarNavRef.value) sidebarNavRef.value.scrollTop = appStore.sidebarScrollTop
    })
  }
})

onBeforeUnmount(() => {
  if (sidebarNavRef.value) appStore.sidebarScrollTop = sidebarNavRef.value.scrollTop
  desktopMediaQuery?.removeEventListener('change', handleDesktopMediaChange)
  document.removeEventListener('keydown', handleGlobalKeydown)
  unlockBodyScroll()
})
</script>

<style scoped>
.admin-sidebar {
  @apply fixed inset-y-0 left-0 z-50 flex flex-col border-r border-outline bg-surface;
  width: min(280px, calc(100vw - 40px));
  transition: width 180ms ease, transform 180ms ease;
}

.admin-sidebar-header {
  @apply flex flex-shrink-0 items-center gap-3 border-b border-outline px-4;
  height: var(--app-header-height, 64px);
}

.sidebar-logo {
  @apply flex h-8 w-8 flex-shrink-0 items-center justify-center overflow-hidden rounded-control border border-outline bg-surface-raised focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40 focus-visible:ring-offset-2 focus-visible:ring-offset-surface;
}

.sidebar-brand {
  min-width: 0;
  flex: 1 1 auto;
  white-space: nowrap;
  transition: max-width 160ms ease, opacity 120ms ease;
  max-width: 12rem;
}

.sidebar-brand-collapsed {
  max-width: 0;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
}

.sidebar-brand-title {
  @apply block truncate rounded-sm text-sm font-semibold text-foreground hover:text-foreground-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.admin-context-label {
  @apply rounded border border-outline bg-surface-subtle px-1.5 py-0.5 text-[10px] font-semibold uppercase leading-none text-foreground-muted;
  letter-spacing: 0;
}

.sidebar-header-action {
  @apply ml-auto flex h-10 w-10 items-center justify-center rounded-control text-foreground-muted hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.admin-sidebar-nav {
  @apply flex-1 overflow-y-auto overflow-x-hidden px-2.5 py-3;
  scrollbar-gutter: stable;
}

.sidebar-section {
  @apply mb-3;
}

.sidebar-section-title {
  @apply mb-1 flex h-6 items-center overflow-hidden px-2 text-[11px] font-semibold uppercase text-foreground-subtle;
  letter-spacing: 0;
  white-space: nowrap;
}

.sidebar-section-title-collapsed {
  @apply justify-center px-0 text-transparent;
}

.sidebar-section-title-collapsed::after {
  width: 1.25rem;
  height: 1px;
  content: '';
  background: var(--ui-border, #dbe3ee);
}

.sidebar-group {
  @apply mb-0.5;
}

.sidebar-link {
  @apply relative mb-0.5 flex min-h-10 items-center gap-2.5 overflow-hidden rounded-control px-2.5 py-2 text-sm font-medium text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.sidebar-link-active {
  @apply bg-info-subtle font-semibold text-foreground hover:bg-info-subtle;
  box-shadow: inset 0 0 0 1px rgb(var(--color-info) / 0.16);
}

.sidebar-link-active .sidebar-link-icon {
  @apply text-info-foreground;
}

.sidebar-group-active {
  @apply bg-surface-subtle text-foreground;
}

.sidebar-group-active .sidebar-link-icon {
  @apply text-info-foreground;
}

.sidebar-link-collapsed {
  @apply justify-center gap-0 px-0;
}

.sidebar-link-icon {
  @apply h-5 w-5 flex-shrink-0;
}

.sidebar-label {
  display: block;
  min-width: 0;
  max-width: 12rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: max-width 160ms ease, opacity 100ms ease;
}

.sidebar-group-label {
  @apply flex flex-1 items-center justify-between gap-2;
}

.sidebar-label-collapsed {
  max-width: 0;
  opacity: 0;
  pointer-events: none;
}

.sidebar-children {
  @apply mb-1 ml-4 border-l border-outline pl-3;
}

.sidebar-child-link {
  @apply mb-0.5 flex min-h-9 items-center gap-2 rounded-control px-2 py-1.5 text-[13px] font-medium text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.sidebar-child-link-active {
  @apply bg-info-subtle font-semibold text-foreground hover:bg-info-subtle;
  box-shadow: inset 0 0 0 1px rgb(var(--color-info) / 0.16);
}

.sidebar-child-link-active .sidebar-link-icon {
  @apply text-info-foreground;
}

.sidebar-svg-icon {
  color: currentColor;
}

.sidebar-svg-icon :deep(svg) {
  display: block;
  width: 1.25rem;
  height: 1.25rem;
}

.admin-sidebar-footer {
  @apply flex-shrink-0 border-t border-outline p-2.5;
}

.drawer-backdrop {
  @apply fixed inset-0 z-40 cursor-default bg-black/45 lg:hidden;
}

@media (min-width: 1024px) {
  .admin-sidebar {
    width: var(--sidebar-width, 256px);
  }

  .admin-sidebar-collapsed {
    width: var(--sidebar-width-collapsed, 72px);
  }
}

.drawer-backdrop-enter-active,
.drawer-backdrop-leave-active {
  transition: opacity 150ms ease;
}

.drawer-backdrop-enter-from,
.drawer-backdrop-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .admin-sidebar,
  .sidebar-brand,
  .sidebar-label,
  .drawer-backdrop-enter-active,
  .drawer-backdrop-leave-active {
    transition-duration: 1ms;
  }
}
</style>
