<template>
  <div ref="root" class="relative inline-flex">
    <button
      type="button"
      class="btn btn-ghost btn-icon btn-sm"
      :title="t('common.more')"
      :aria-label="t('common.more')"
      :aria-expanded="open"
      aria-haspopup="menu"
      @click.stop="toggle"
    >
      <Icon name="more" size="sm" />
    </button>
    <Teleport to="body">
    <div
      v-if="open"
      ref="menu"
      class="fixed z-[9999] w-52 rounded-panel border border-outline bg-surface-raised p-1 text-left shadow-floating"
      :style="menuStyle"
      role="menu"
      @click.stop
    >
      <button v-for="item in primaryItems" :key="item.action" type="button" role="menuitem" class="dropdown-item w-full" @click="select(item.action)">
        <Icon :name="item.icon" size="sm" class="text-foreground-subtle" />
        <span>{{ item.label }}</span>
      </button>
      <div class="my-1 border-t border-outline" />
      <button type="button" role="menuitem" class="dropdown-item w-full" @click="select('view')"><Icon name="eye" size="sm" class="text-foreground-subtle" />{{ t('admin.distribution.agentViewDetails') }}</button>
      <button type="button" role="menuitem" class="dropdown-item w-full" @click="select('rate')"><Icon name="edit" size="sm" class="text-foreground-subtle" />{{ t('admin.distribution.agentEditRate') }}</button>
      <button type="button" role="menuitem" class="dropdown-item w-full" @click="select('permissions')"><Icon name="shield" size="sm" class="text-foreground-subtle" />{{ t('admin.distribution.agentPermissions') }}</button>
      <button type="button" role="menuitem" class="dropdown-item w-full" @click="select('rewards')"><Icon name="gift" size="sm" class="text-foreground-subtle" />{{ t('admin.distribution.agentRewards') }}</button>
      <div v-if="agent.status !== 'revoked'" class="my-1 border-t border-outline" />
      <button v-if="agent.status !== 'revoked'" type="button" role="menuitem" class="dropdown-item w-full" @click="select(agent.status === 'active' ? 'suspend' : 'activate')">
        <Icon :name="agent.status === 'active' ? 'ban' : 'play'" size="sm" class="text-foreground-subtle" />
        {{ agent.status === 'active' ? t('admin.distribution.agentSuspend') : t('admin.distribution.agentActivate') }}
      </button>
      <button v-if="agent.status !== 'revoked'" type="button" role="menuitem" class="dropdown-item w-full text-danger" @click="select('revoke')"><Icon name="trash" size="sm" />{{ t('admin.distribution.agentRevoke') }}</button>
    </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DistributionAgent } from '@/api/distribution'

defineProps<{ agent: DistributionAgent }>()
const emit = defineEmits<{ (event: 'action', action: string): void }>()
const { t } = useI18n()
const root = ref<HTMLElement | null>(null)
const menu = ref<HTMLElement | null>(null)
const open = ref(false)
const position = ref({ top: 0, left: 0 })
const menuStyle = computed(() => ({ top: `${position.value.top}px`, left: `${position.value.left}px`, maxHeight: `calc(100dvh - ${position.value.top + 8}px)`, overflowY: 'auto' as const }))
const primaryItems = [{ action: 'customers', icon: 'users' as const, label: t('admin.distribution.agentCustomers') }]
function select(action: string) { open.value = false; emit('action', action) }
async function toggle() {
  open.value = !open.value
  if (!open.value || !root.value) return
  const rect = root.value.getBoundingClientRect()
  position.value = { top: Math.min(rect.bottom + 4, window.innerHeight - 80), left: Math.max(8, Math.min(rect.right - 208, window.innerWidth - 216)) }
  await nextTick()
  menu.value?.querySelector<HTMLElement>('[role="menuitem"]')?.focus()
}
function close(event: MouseEvent) { const node = event.target as Node; if (!root.value?.contains(node) && !menu.value?.contains(node)) open.value = false }
onMounted(() => document.addEventListener('click', close))
onBeforeUnmount(() => document.removeEventListener('click', close))
</script>
