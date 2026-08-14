<template>
  <nav class="border-b border-outline" aria-label="分销管理导航">
    <div ref="tabsRef" class="admin-distribution-tabs overflow-x-auto">
      <RouterLink v-for="item in items" :key="item.path" :to="item.path" class="admin-distribution-tab" :class="route.path === item.path ? 'admin-distribution-tab-active' : ''" :aria-current="route.path === item.path ? 'page' : undefined">
        <Icon :name="item.icon" size="sm" /><span>{{ t(item.labelKey) }}</span>
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const route = useRoute() ?? { path: '' }
const tabsRef = ref<HTMLElement | null>(null)
const { t } = useI18n()
const items = [
  { path: '/admin/distribution/overview', labelKey: 'admin.distribution.nav.overview', icon: 'home' as const },
  { path: '/admin/distribution/agent-analytics', labelKey: 'admin.distribution.nav.analysis', icon: 'trendingUp' as const },
  { path: '/admin/distribution/agents', labelKey: 'admin.distribution.nav.agents', icon: 'users' as const },
  { path: '/admin/distribution/customers', labelKey: 'admin.distribution.nav.customers', icon: 'user' as const },
  { path: '/admin/distribution/promotion', labelKey: 'admin.distribution.nav.promotion', icon: 'link' as const },
  { path: '/admin/distribution/commissions', labelKey: 'admin.distribution.nav.commissions', icon: 'gift' as const },
  { path: '/admin/distribution/withdrawals', labelKey: 'admin.distribution.nav.withdrawals', icon: 'creditCard' as const },
  { path: '/admin/distribution/anomalies', labelKey: 'admin.distribution.nav.anomalies', icon: 'shield' as const },
]

function revealActiveTab() {
  void nextTick(() => {
    const activeTab = tabsRef.value?.querySelector<HTMLElement>('[aria-current="page"]')
    activeTab?.scrollIntoView({
      behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth',
      block: 'nearest',
      inline: 'center',
    })
  })
}

onMounted(revealActiveTab)
watch(() => route.path, revealActiveTab)
</script>

<style scoped>
.admin-distribution-tabs { display: flex; width: 100%; max-width: 100%; gap: 4px; scrollbar-width: none; overscroll-behavior-inline: contain; scroll-snap-type: inline proximity; }
.admin-distribution-tabs::-webkit-scrollbar { display: none; }
.admin-distribution-tab { display: inline-flex; min-width: 60px; height: 44px; flex: 0 0 auto; scroll-snap-align: start; align-items: center; justify-content: center; gap: 6px; border-bottom: 2px solid transparent; padding: 0 12px; color: var(--ui-text-muted); font-size: 14px; white-space: nowrap; transition: color 150ms ease, border-color 150ms ease; }
.admin-distribution-tab:hover { color: var(--ui-text); }
.admin-distribution-tab-active { border-bottom-color: var(--ui-focus); color: var(--ui-text); font-weight: 600; }
@media (max-width: 639px) { .admin-distribution-tab { min-width: 56px; padding-inline: 10px; font-size: 13px; } }
</style>
