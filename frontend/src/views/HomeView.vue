<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <HomeExperiment
    v-else
    :site-name="siteName"
    :site-subtitle="siteSubtitle"
    :is-authenticated="isAuthenticated"
    :dashboard-path="dashboardPath"
  />
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAuthStore, useAppStore } from '@/stores'
import HomeExperiment from './HomeExperiment.vue'

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'USA-零')
const siteSubtitle = computed(() =>
  appStore.cachedPublicSettings?.site_subtitle
  || '统一 OpenAI、Claude、Gemini 等不同接口，把多模型调用规范成一个稳定、可计量、可治理的标准 API。'
)
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    void appStore.fetchPublicSettings()
  }
})
</script>
