<template>
  <dl class="mt-4 grid grid-cols-2 divide-x divide-outline border-y border-outline py-3">
    <div class="min-w-0 pr-3">
      <dt
        class="flex items-center gap-1.5 text-[10px] font-semibold uppercase text-foreground-subtle"
      >
        <Icon :name="primaryIcon" size="xs" />
        <span class="truncate">{{ primaryLabel }}</span>
      </dt>
      <dd class="mt-1.5 break-all font-mono text-lg font-semibold tabular-nums text-foreground">
        {{ primaryValue }}<span class="ml-0.5 text-xs font-normal text-foreground-subtle">{{ primaryUnit }}</span>
      </dd>
    </div>
    <div class="min-w-0 pl-3">
      <dt
        class="flex items-center gap-1.5 text-[10px] font-semibold uppercase text-foreground-subtle"
      >
        <Icon :name="secondaryIcon" size="xs" />
        <span class="truncate">{{ secondaryLabel }}</span>
      </dt>
      <dd class="mt-1.5 break-all font-mono text-lg font-semibold tabular-nums text-foreground">
        {{ secondaryValue }}<span class="ml-0.5 text-xs font-normal text-foreground-subtle">{{ secondaryUnit }}</span>
      </dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  primaryLabel: string
  primaryValue: string
  primaryUnit: string
  primaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
  secondaryLabel: string
  secondaryValue: string
  secondaryUnit: string
  secondaryIcon: 'bolt' | 'globe' | 'clock' | 'link'
}>()
</script>
