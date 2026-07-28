<template>
  <header class="admin-header">
    <div class="flex h-[var(--app-header-height)] min-w-0 items-center justify-between gap-3 px-3 sm:px-4 md:px-6">
      <div class="flex min-w-0 items-center gap-2.5">
        <button
          type="button"
          class="header-action inline-flex lg:hidden"
          :aria-label="localText('打开导航', 'Open navigation')"
          :aria-expanded="appStore.mobileOpen"
          aria-controls="app-sidebar"
          @click="toggleMobileSidebar"
        >
          <Icon name="menu" size="md" />
        </button>

        <div class="min-w-0">
          <div class="flex min-w-0 items-center gap-2">
            <span
              v-if="authStore.isAdmin && isAdminContext"
              class="admin-mode-badge hidden sm:inline-flex"
            >Admin</span>
            <h1 class="truncate text-base font-semibold text-foreground lg:text-lg">
              {{ pageTitle }}
            </h1>
          </div>
          <p
            v-if="pageDescription"
            class="hidden truncate text-xs text-foreground-muted md:block"
          >
            {{ pageDescription }}
          </p>
        </div>
      </div>

      <div class="flex flex-shrink-0 items-center gap-1 sm:gap-1.5">
        <AnnouncementBell v-if="user" />

        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="header-action hidden sm:inline-flex"
          :aria-label="t('nav.docs')"
          :title="t('nav.docs')"
        >
          <Icon name="book" size="md" />
        </a>

        <div class="hidden min-[340px]:block">
          <LocaleSwitcher />
        </div>

        <button
          type="button"
          class="header-action inline-flex"
          :aria-label="isDark ? t('nav.lightMode') : t('nav.darkMode')"
          :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
          @click="toggleTheme"
        >
          <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
        </button>

        <div v-if="user && !authStore.isAdmin" class="hidden xl:block">
          <SubscriptionProgressMini />
        </div>

        <div
          v-if="user && !authStore.isAdmin"
          class="hidden items-center gap-1.5 rounded-control border border-outline px-2.5 py-1.5 text-sm bg-canvas lg:flex"
          :title="`${balanceAvailableText}: ${formatHeaderMoney(availableBalance)}`"
        >
          <Icon name="dollar" size="sm" class="text-foreground-subtle" />
          <span class="font-semibold tabular-nums text-foreground-muted">
            {{ formatHeaderMoney(availableBalance) }}
          </span>
          <span
            v-if="frozenBalance > 0"
            class="rounded bg-warning-subtle px-1 py-0.5 text-[10px] font-semibold text-warning-foreground"
          >
            {{ balanceFrozenText }}
          </span>
        </div>

        <div v-if="user" ref="dropdownRef" class="relative">
          <button
            :id="userMenuTriggerId"
            ref="userMenuButtonRef"
            type="button"
            class="user-menu-trigger"
            :aria-label="localText(`${displayName}的用户菜单`, `User menu for ${displayName}`)"
            :aria-expanded="dropdownOpen"
            :aria-controls="dropdownOpen ? userMenuId : undefined"
            aria-haspopup="menu"
            @click="toggleDropdown"
            @keydown="handleTriggerKeydown"
          >
            <div class="user-avatar">
              <img
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                class="h-full w-full object-cover"
              >
              <span v-else>{{ userInitials }}</span>
            </div>
            <div class="hidden min-w-0 text-left xl:block">
              <div class="max-w-28 truncate text-sm font-medium text-foreground">
                {{ displayName }}
              </div>
              <div class="text-[11px] capitalize text-foreground-muted">
                {{ user.role }}
              </div>
            </div>
            <Icon name="chevronDown" size="xs" class="hidden text-foreground-subtle xl:block" />
          </button>

          <Transition name="dropdown">
            <div
              v-if="dropdownOpen"
              :id="userMenuId"
              ref="userMenuRef"
              role="menu"
              :aria-labelledby="userMenuTriggerId"
              class="user-dropdown"
              @keydown="handleMenuKeydown"
            >
              <div role="presentation" class="border-b border-outline px-3.5 py-3">
                <div class="flex items-center gap-2">
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-semibold text-foreground">
                      {{ displayName }}
                    </div>
                    <div class="truncate text-xs text-foreground-muted">{{ user.email }}</div>
                  </div>
                  <span v-if="authStore.isAdmin" class="admin-role-badge">Admin</span>
                </div>
              </div>

              <div
                v-if="!authStore.isAdmin"
                role="presentation"
                class="border-b border-outline px-3.5 py-2 lg:hidden"
              >
                <div class="text-[11px] text-foreground-muted">{{ balanceAvailableText }}</div>
                <div class="text-sm font-semibold tabular-nums text-foreground">
                  {{ formatHeaderMoney(availableBalance) }}
                </div>
              </div>

              <div role="group" class="py-1" :aria-label="t('nav.myAccount')">
                <router-link
                  v-if="authStore.isAdmin"
                  :to="isAdminContext ? '/dashboard' : '/admin/dashboard'"
                  role="menuitem"
                  tabindex="-1"
                  class="user-dropdown-item"
                  @click="closeDropdown"
                >
                  <Icon name="swap" size="sm" />
                  {{ isAdminContext ? t('nav.myAccount') : t('admin.dashboard.title') }}
                </router-link>

                <router-link
                  to="/profile"
                  role="menuitem"
                  tabindex="-1"
                  class="user-dropdown-item"
                  @click="closeDropdown"
                >
                  <Icon name="user" size="sm" />
                  {{ t('nav.profile') }}
                </router-link>

                <router-link
                  to="/keys"
                  role="menuitem"
                  tabindex="-1"
                  class="user-dropdown-item"
                  @click="closeDropdown"
                >
                  <Icon name="key" size="sm" />
                  {{ t('nav.apiKeys') }}
                </router-link>

              </div>

              <div
                v-if="contactInfo"
                role="presentation"
                class="border-t border-outline px-3.5 py-2.5 text-xs text-foreground-muted"
              >
                <span>{{ t('common.contactSupport') }}:</span>
                <span class="ml-1 break-all font-medium text-foreground-muted">{{ contactInfo }}</span>
              </div>

              <div v-if="showOnboardingButton" role="group" class="border-t border-outline py-1">
                <button
                  type="button"
                  role="menuitem"
                  tabindex="-1"
                  class="user-dropdown-item w-full"
                  @click="handleReplayGuide"
                >
                  <Icon name="lightbulb" size="sm" />
                  {{ t('onboarding.restartTour') }}
                </button>
              </div>

              <div role="group" class="border-t border-outline py-1">
                <button
                  type="button"
                  role="menuitem"
                  tabindex="-1"
                  class="user-dropdown-item w-full text-danger-foreground hover:bg-danger-subtle"
                  @click="handleLogout"
                >
                  <Icon name="login" size="sm" class="rotate-180" />
                  {{ t('nav.logout') }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore, useOnboardingStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import { useTheme } from '@/composables/useTheme'
import { useDropdownMenu } from '@/composables/useDropdownMenu'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import SubscriptionProgressMini from '@/components/common/SubscriptionProgressMini.vue'
import AnnouncementBell from '@/components/common/AnnouncementBell.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const adminSettingsStore = useAdminSettingsStore()
const onboardingStore = useOnboardingStore()
const { isDark, toggleTheme } = useTheme()

const user = computed(() => authStore.user)
const dropdownRef = ref<HTMLElement | null>(null)
const {
  open: dropdownOpen,
  triggerRef: userMenuButtonRef,
  menuRef: userMenuRef,
  triggerId: userMenuTriggerId,
  menuId: userMenuId,
  closeMenu,
  toggleMenu,
  handleTriggerKeydown,
  handleMenuKeydown
} = useDropdownMenu('header-user-menu')
const contactInfo = computed(() => appStore.contactInfo)
const docUrl = computed(() => sanitizeUrl(appStore.docUrl))
const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const availableBalance = computed(() => Number(user.value?.balance || 0))
const frozenBalance = computed(() => Number(user.value?.frozen_balance || 0))

const isAdminCustomPage = computed(() => {
  if (route.name !== 'CustomPage') return false
  const id = String(route.params.id || '')
  return adminSettingsStore.customMenuItems.some((item) => item.id === id && item.visibility === 'admin')
})
const isAdminContext = computed(
  () => authStore.isAdmin && (route.path.startsWith('/admin') || isAdminCustomPage.value)
)

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

const balanceAvailableText = computed(() =>
  t('common.availableBalance') === 'common.availableBalance'
    ? localText('可用余额', 'Available balance')
    : t('common.availableBalance')
)
const balanceFrozenText = computed(() =>
  t('common.frozenBalance') === 'common.frozenBalance'
    ? localText('冻结', 'Frozen')
    : t('common.frozenBalance')
)

const showOnboardingButton = computed(
  () => !authStore.isSimpleMode && user.value?.role === 'admin'
)

const userInitials = computed(() => {
  if (!user.value) return ''
  const source = user.value.username || user.value.email?.split('@')[0] || ''
  return source.substring(0, 2).toUpperCase()
})

const displayName = computed(() =>
  user.value?.username || user.value?.email?.split('@')[0] || ''
)

const pageTitle = computed(() => {
  if (route.name === 'CustomPage') {
    const id = route.params.id as string
    const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
    const menuItem = publicItems.find((item) => item.id === id)
      ?? (authStore.isAdmin
        ? adminSettingsStore.customMenuItems.find((item) => item.id === id)
        : undefined)
    if (menuItem?.label) return menuItem.label
  }
  const titleKey = route.meta.titleKey as string
  if (titleKey) return t(titleKey)
  return (route.meta.title as string) || appStore.siteName
})

const pageDescription = computed(() => {
  const descriptionKey = route.meta.descriptionKey as string
  if (descriptionKey) return t(descriptionKey)
  return (route.meta.description as string) || ''
})

function toggleMobileSidebar(): void {
  appStore.toggleMobileSidebar()
}

function toggleDropdown(): void {
  toggleMenu()
}

function closeDropdown(): void {
  void closeMenu()
}

async function handleLogout(): Promise<void> {
  closeDropdown()
  try {
    await authStore.logout()
  } catch (error) {
    console.error('Logout error:', error)
  }
  await router.push('/login')
}

function handleReplayGuide(): void {
  closeDropdown()
  onboardingStore.replay()
}

function formatHeaderMoney(value: number): string {
  return Number.isFinite(value) ? `$${value.toFixed(2)}` : '$0.00'
}

function handleClickOutside(event: MouseEvent): void {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) closeDropdown()
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && dropdownOpen.value) {
    event.preventDefault()
    void closeMenu(true)
  }
}

watch(
  () => route.fullPath,
  closeDropdown
)

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.admin-header {
  @apply sticky top-0 z-30 border-b border-outline bg-surface/95 backdrop-blur;
}

.header-action {
  @apply h-10 w-10 flex-shrink-0 items-center justify-center rounded-control text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.admin-mode-badge,
.admin-role-badge {
  @apply rounded border border-outline-strong bg-surface-subtle px-1.5 py-0.5 text-[10px] font-semibold uppercase leading-none text-foreground-muted;
  letter-spacing: 0;
}

.user-menu-trigger {
  @apply flex min-h-10 items-center gap-2 rounded-control px-1.5 py-1 transition-colors hover:bg-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.user-avatar {
  @apply flex h-8 w-8 flex-shrink-0 items-center justify-center overflow-hidden rounded-control bg-foreground text-xs font-semibold text-surface;
}

.user-dropdown {
  @apply absolute right-0 mt-2 w-64 origin-top-right overflow-hidden rounded-panel border border-outline bg-surface-raised shadow-floating;
}

.user-dropdown-item {
  @apply flex min-h-10 items-center gap-2.5 px-3.5 py-2 text-left text-sm text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus/40;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 120ms ease, transform 120ms ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}

@media (prefers-reduced-motion: reduce) {
  .dropdown-enter-active,
  .dropdown-leave-active {
    transition-duration: 1ms;
  }
}
</style>
