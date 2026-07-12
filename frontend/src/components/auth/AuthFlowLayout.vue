<template>
  <GatewayAuthLayout :site-name="siteName" :show-register="false">
    <div class="auth-flow" :aria-busy="busy">
      <slot />

      <div
        v-if="busy"
        class="auth-flow-progress"
        role="status"
        aria-live="polite"
      >
        <span class="auth-flow-spinner" aria-hidden="true"></span>
        <span class="sr-only">{{ busyLabel }}</span>
      </div>

      <div v-if="$slots.footer" class="auth-flow-footer">
        <slot name="footer" />
      </div>
    </div>
  </GatewayAuthLayout>
</template>

<script setup lang="ts">
import GatewayAuthLayout from '@/components/auth/GatewayAuthLayout.vue'

withDefaults(defineProps<{
  busy?: boolean
  busyLabel?: string
  siteName?: string
}>(), {
  busy: false,
  busyLabel: 'Processing',
  siteName: ''
})
</script>

<style scoped>
.auth-flow {
  min-width: 0;
  overflow-wrap: anywhere;
}

.auth-flow-progress {
  display: grid;
  min-height: 104px;
  place-items: center;
}

.auth-flow-spinner {
  width: 30px;
  height: 30px;
  border: 2px solid var(--border);
  border-top-color: var(--foreground);
  border-radius: 50%;
  animation: auth-flow-spin 700ms linear infinite;
}

.auth-flow-footer {
  display: flex;
  justify-content: center;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--border);
  color: var(--muted-foreground);
  font-size: 13px;
  text-align: center;
}

.auth-flow :deep(.auth-flow-surface) {
  padding: 16px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: color-mix(in srgb, var(--muted) 72%, transparent);
}

.auth-flow :deep(.auth-flow-alert) {
  padding: 12px 14px;
  border: 1px solid color-mix(in srgb, #dc2626 30%, var(--border));
  border-radius: 8px;
  background: color-mix(in srgb, #dc2626 8%, var(--background));
  color: color-mix(in srgb, #dc2626 80%, var(--foreground));
  font-size: 13px;
  line-height: 1.55;
}

.auth-flow :deep(.btn) {
  min-height: 44px;
  border-radius: 8px;
}

.auth-flow :deep(.btn-primary) {
  border: 1px solid var(--button-background);
  background: var(--button-background);
  color: var(--button-foreground);
  box-shadow: 0 10px 24px var(--button-shadow);
  font-weight: 800;
}

.auth-flow :deep(.btn-primary:hover:not(:disabled)) {
  border-color: var(--button-hover);
  background: var(--button-hover);
}

.auth-flow :deep(.btn-secondary) {
  border-color: var(--border);
  background: var(--background);
  color: var(--foreground);
  box-shadow: none;
}

.auth-flow :deep(.btn-secondary:hover:not(:disabled)) {
  border-color: var(--border-strong);
  background: var(--muted);
}

@keyframes auth-flow-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-flow-spinner {
    animation: none;
  }
}
</style>
