<template>
  <div class="flex min-h-[100dvh] bg-canvas px-4 py-10 text-foreground sm:px-6">
    <main
      class="mx-auto flex w-full max-w-xl flex-col items-center justify-center text-center"
      aria-labelledby="not-found-title"
    >
      <div
        class="flex h-12 w-12 items-center justify-center rounded-panel border border-outline bg-surface text-foreground-muted shadow-card"
        aria-hidden="true"
      >
        <Icon name="exclamationCircle" size="lg" />
      </div>

      <p class="mt-6 font-mono text-sm font-semibold tabular-nums text-foreground-subtle">404</p>
      <h1 id="not-found-title" class="mt-2 text-2xl font-semibold text-foreground sm:text-3xl">
        {{ t('errors.pageNotFound') }}
      </h1>
      <p class="mt-3 max-w-md text-sm leading-6 text-foreground-muted sm:text-base">
        {{ t('errors.pageNotFoundDescription') }}
      </p>

      <div class="mt-8 flex w-full max-w-sm flex-col gap-3 sm:flex-row sm:justify-center">
        <button type="button" class="btn btn-secondary w-full sm:w-auto" @click="goBack">
          <Icon name="arrowLeft" size="md" class="mr-2" aria-hidden="true" />
          {{ t('common.back') }}
        </button>
        <RouterLink :to="primaryDestination" class="btn btn-primary w-full sm:w-auto">
          <Icon :name="isAuthenticated ? 'grid' : 'home'" size="md" class="mr-2" aria-hidden="true" />
          {{ primaryLabel }}
        </RouterLink>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const primaryDestination = computed(() => {
  if (!isAuthenticated.value) return '/home'
  return authStore.isAdmin ? '/admin/dashboard' : '/dashboard'
})
const primaryLabel = computed(() =>
  isAuthenticated.value ? t('home.goToDashboard') : t('errors.backToHome')
)

function goBack(): void {
  if (typeof window !== 'undefined' && window.history.state?.back) {
    router.back()
    return
  }
  void router.push('/home')
}
</script>
