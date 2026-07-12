<template>
  <section class="card overflow-hidden" aria-labelledby="dashboard-quick-actions-title">
    <header class="flex min-h-[3.25rem] items-center border-b border-outline bg-surface-subtle px-4 py-3 sm:px-5">
      <h2 id="dashboard-quick-actions-title" class="text-sm font-semibold text-foreground">{{ t('dashboard.quickActions') }}</h2>
    </header>
    <nav :aria-label="t('dashboard.quickActions')">
      <router-link to="/keys" class="dashboard-action">
        <span class="dashboard-action-icon" aria-hidden="true">
          <Icon name="key" size="md" />
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.createApiKey') }}</p>
          <p class="text-xs text-foreground-subtle">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="shrink-0 text-foreground-subtle" aria-hidden="true" />
      </router-link>

      <router-link to="/usage" class="dashboard-action">
        <span class="dashboard-action-icon" aria-hidden="true">
          <Icon name="chart" size="md" />
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.viewUsage') }}</p>
          <p class="text-xs text-foreground-subtle">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="shrink-0 text-foreground-subtle" aria-hidden="true" />
      </router-link>

      <router-link v-if="canUseBatchImage" to="/batch-image" class="dashboard-action">
        <span class="dashboard-action-icon" aria-hidden="true">
          <Icon name="sparkles" size="md" />
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.batchImageAgent') }}</p>
          <p class="text-xs text-foreground-subtle">{{ t('dashboard.batchImageAgentDesc') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="shrink-0 text-foreground-subtle" aria-hidden="true" />
      </router-link>

      <router-link to="/redeem" class="dashboard-action">
        <span class="dashboard-action-icon" aria-hidden="true">
          <Icon name="gift" size="md" />
        </span>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">{{ t('dashboard.redeemCode') }}</p>
          <p class="text-xs text-foreground-subtle">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon name="chevronRight" size="sm" class="shrink-0 text-foreground-subtle" aria-hidden="true" />
      </router-link>
    </nav>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
const { t } = useI18n()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

onMounted(() => {
  void refreshBatchImageAccess()
})
</script>

<style scoped>
.dashboard-action {
  display: flex;
  min-height: 4.25rem;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--ui-border);
  transition: background-color var(--duration-fast) var(--ease-standard);
}

.dashboard-action:first-child {
  border-top: 0;
}

.dashboard-action:hover {
  background: var(--ui-surface-subtle);
}

.dashboard-action:focus-visible {
  outline: 2px solid var(--ui-focus);
  outline-offset: -2px;
}

.dashboard-action-icon {
  display: inline-flex;
  width: 2.25rem;
  height: 2.25rem;
  flex: 0 0 2.25rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--ui-border);
  border-radius: var(--radius-sm);
  background: var(--ui-surface-subtle);
  color: var(--ui-text-muted);
}
</style>
