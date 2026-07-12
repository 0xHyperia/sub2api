<template>
  <div class="min-h-screen bg-canvas text-foreground">
    <a
      href="#app-main-content"
      class="fixed left-3 top-3 z-[100] -translate-y-20 rounded-md bg-gray-950 px-3 py-2 text-sm font-medium text-white shadow-lg transition-transform focus:translate-y-0 dark:bg-white dark:text-gray-950"
    >
      {{ skipLinkLabel }}
    </a>

    <AppSidebar />

    <div
      class="relative min-h-screen min-w-0 transition-[margin] duration-200 motion-reduce:transition-none"
      :class="[sidebarCollapsed ? 'lg:ml-[var(--sidebar-width-collapsed)]' : 'lg:ml-[var(--sidebar-width)]']"
    >
      <AppHeader />

      <main id="app-main-content" class="min-w-0 p-gutter" tabindex="-1">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const { locale } = useI18n()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const skipLinkLabel = computed(() =>
  locale.value.startsWith('zh') ? '跳到主要内容' : 'Skip to main content'
)

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
