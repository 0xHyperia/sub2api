<template>
  <article class="dashboard-metric" :data-tone="tone">
    <div class="dashboard-metric-heading">
      <span class="dashboard-metric-icon" aria-hidden="true">
        <Icon :name="icon" size="sm" :stroke-width="1.8" />
      </span>
      <span class="dashboard-metric-label">{{ label }}</span>
    </div>

    <div class="dashboard-metric-value" :title="value">{{ value }}</div>
    <p v-if="detail" class="dashboard-metric-detail">{{ detail }}</p>
  </article>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

type IconName = InstanceType<typeof Icon>['$props']['name']
type MetricTone = 'neutral' | 'success' | 'warning' | 'danger'

withDefaults(
  defineProps<{
    label: string
    value: string
    detail?: string
    icon: IconName
    tone?: MetricTone
  }>(),
  {
    detail: '',
    tone: 'neutral'
  }
)
</script>

<style scoped>
.dashboard-metric {
  min-width: 0;
  min-height: 116px;
  padding: 16px;
  border-right: 1px solid var(--ui-border, #dbe3ee);
  background: var(--ui-surface, #fff);
}

.dashboard-metric:last-child {
  border-right: 0;
}

.dashboard-metric-heading {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.dashboard-metric-icon {
  display: inline-grid;
  width: 28px;
  height: 28px;
  flex: 0 0 28px;
  place-items: center;
  border: 1px solid var(--ui-border, #dbe3ee);
  border-radius: 7px;
  background: var(--ui-surface-subtle, #f4f7fb);
  color: var(--ui-text-muted, #667085);
}

.dashboard-metric[data-tone='success'] .dashboard-metric-icon {
  border-color: color-mix(in srgb, var(--ui-success, #059669) 24%, transparent);
  background: color-mix(in srgb, var(--ui-success, #059669) 9%, transparent);
  color: var(--ui-success, #059669);
}

.dashboard-metric[data-tone='warning'] .dashboard-metric-icon {
  border-color: color-mix(in srgb, var(--ui-warning, #d97706) 24%, transparent);
  background: color-mix(in srgb, var(--ui-warning, #d97706) 9%, transparent);
  color: var(--ui-warning, #d97706);
}

.dashboard-metric[data-tone='danger'] .dashboard-metric-icon {
  border-color: color-mix(in srgb, var(--ui-danger, #e11d48) 24%, transparent);
  background: color-mix(in srgb, var(--ui-danger, #e11d48) 9%, transparent);
  color: var(--ui-danger, #e11d48);
}

.dashboard-metric-label {
  overflow: hidden;
  color: var(--ui-text-muted, #667085);
  font-size: 12px;
  font-weight: 650;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-metric-value {
  margin-top: 13px;
  overflow: hidden;
  color: var(--ui-text, #0f172a);
  font-size: 25px;
  font-weight: 720;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0;
  line-height: 1.05;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dashboard-metric-detail {
  margin: 7px 0 0;
  overflow: hidden;
  color: var(--ui-text-subtle, #8793a3);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .dashboard-metric {
    min-height: 108px;
    border-right: 0;
    border-bottom: 1px solid var(--ui-border, #dbe3ee);
  }

  .dashboard-metric:nth-child(odd) {
    border-right: 1px solid var(--ui-border, #dbe3ee);
  }

  .dashboard-metric:nth-last-child(-n + 2) {
    border-bottom: 0;
  }

  .dashboard-metric-value {
    font-size: 22px;
  }
}
</style>
