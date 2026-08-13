<template>
  <nav class="border-b border-outline" aria-label="分销管理导航">
    <div class="admin-distribution-tabs overflow-x-auto">
      <RouterLink v-for="item in items" :key="item.path" :to="item.path" class="admin-distribution-tab" :class="route.path === item.path ? 'admin-distribution-tab-active' : ''" :aria-current="route.path === item.path ? 'page' : undefined">
        <Icon :name="item.icon" size="sm" /><span>{{ item.label }}</span>
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'

const route = useRoute() ?? { path: '' }
const items = [
  { path: '/admin/distribution/overview', label: '总览', icon: 'home' as const },
  { path: '/admin/distribution/agents', label: '代理', icon: 'users' as const },
  { path: '/admin/distribution/customers', label: '客户', icon: 'user' as const },
  { path: '/admin/distribution/promotion', label: '推广', icon: 'link' as const },
  { path: '/admin/distribution/commissions', label: '佣金', icon: 'gift' as const },
  { path: '/admin/distribution/withdrawals', label: '提现', icon: 'creditCard' as const },
  { path: '/admin/distribution/anomalies', label: '异常', icon: 'shield' as const },
]
</script>

<style scoped>
.admin-distribution-tabs { display: flex; width: 100%; max-width: 100%; gap: 4px; scrollbar-width: none; overscroll-behavior-inline: contain; }
.admin-distribution-tabs::-webkit-scrollbar { display: none; }
.admin-distribution-tab { display: inline-flex; min-width: 60px; height: 44px; align-items: center; justify-content: center; gap: 6px; border-bottom: 2px solid transparent; padding: 0 12px; color: var(--ui-text-muted); font-size: 14px; white-space: nowrap; transition: color 150ms ease, border-color 150ms ease; }
.admin-distribution-tab:hover { color: var(--ui-text); }
.admin-distribution-tab-active { border-bottom-color: var(--ui-focus); color: var(--ui-text); font-weight: 600; }
@media (max-width: 639px) { .admin-distribution-tab { min-width: 56px; padding-inline: 10px; font-size: 13px; } }
</style>
