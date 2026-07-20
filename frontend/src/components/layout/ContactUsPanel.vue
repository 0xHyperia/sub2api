<template>
  <template v-if="visible">
    <button
      ref="triggerRef"
      type="button"
      class="contact-trigger"
      :class="{ 'contact-trigger-collapsed': collapsed }"
      :title="collapsed ? copy.title : undefined"
      :aria-label="collapsed ? copy.title : undefined"
      @click="showPanel"
    >
      <Icon name="chatBubble" size="md" class="contact-trigger-icon" />
      <span
        class="contact-trigger-label"
        :class="{ 'contact-trigger-label-collapsed': collapsed }"
        :aria-hidden="collapsed ? 'true' : 'false'"
      >
        {{ copy.title }}
      </span>
    </button>

    <BaseDialog
      :show="open"
      :title="copy.title"
      width="narrow"
      :close-on-click-outside="true"
      @close="close"
    >
      <p class="contact-intro">{{ copy.intro }}</p>
      <div class="contact-actions">
        <a
          v-if="qqURL"
          class="contact-action"
          :href="qqURL"
          target="_blank"
          rel="noopener noreferrer"
          @click="close"
        >
          <span class="contact-icon contact-icon-qq"><Icon name="users" size="md" /></span>
          <span class="contact-copy">
            <strong>{{ settings.contact_qq_name || copy.qqTitle }}</strong>
            <span>{{ copy.qqDescription }}</span>
          </span>
          <Icon name="externalLink" size="sm" class="contact-external" />
        </a>

        <a
          v-if="telegramURL"
          class="contact-action"
          :href="telegramURL"
          target="_blank"
          rel="noopener noreferrer"
          @click="close"
        >
          <span class="contact-icon contact-icon-telegram"><Icon name="chatBubble" size="md" /></span>
          <span class="contact-copy">
            <strong>{{ settings.contact_telegram_name || copy.telegramTitle }}</strong>
            <span>{{ copy.telegramDescription }}</span>
          </span>
          <Icon name="externalLink" size="sm" class="contact-external" />
        </a>

        <button
          v-if="settings.contact_ticket_enabled"
          type="button"
          class="contact-action w-full text-left"
          @click="openTicket"
        >
          <span class="contact-icon contact-icon-ticket"><Icon name="clipboard" size="md" /></span>
          <span class="contact-copy">
            <strong>{{ copy.ticketTitle }}</strong>
            <span>{{ copy.ticketDescription }}</span>
          </span>
          <Icon name="chevronRight" size="sm" class="contact-external" />
        </button>
      </div>
    </BaseDialog>
  </template>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { PublicSettings } from '@/types'
import { sanitizeUrl } from '@/utils/url'

const props = defineProps<{
  settings: PublicSettings
  collapsed: boolean
  locale: string
}>()

const router = useRouter()
const open = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const emit = defineEmits<{ (event: 'open'): void }>()

const qqURL = computed(() => props.settings.contact_qq_enabled ? sanitizeUrl(props.settings.contact_qq_url || '') : '')
const telegramURL = computed(() => props.settings.contact_telegram_enabled ? sanitizeUrl(props.settings.contact_telegram_url || '') : '')
const visible = computed(() => props.settings.contact_us_enabled && Boolean(
  qqURL.value || telegramURL.value || props.settings.contact_ticket_enabled
))

const copy = computed(() => props.locale.startsWith('zh') ? {
  title: '联系我们',
  intro: '加入社群获取公告和交流支持，或通过工单提交问题与建议。',
  qqTitle: '加入 QQ 群',
  qqDescription: '交流使用经验，获取社群帮助',
  telegramTitle: '加入 Telegram 群',
  telegramDescription: '关注公告，与社区成员交流',
  ticketTitle: '提交问题 / 建议',
  ticketDescription: '通过工单反馈使用问题或产品建议'
} : {
  title: 'Contact us',
  intro: 'Join the community for updates and help, or send us an issue or suggestion.',
  qqTitle: 'Join the QQ group',
  qqDescription: 'Share tips and get community help',
  telegramTitle: 'Join the Telegram group',
  telegramDescription: 'Follow updates and talk with the community',
  ticketTitle: 'Submit an issue or suggestion',
  ticketDescription: 'Send product feedback through a support ticket'
})

function close(): void {
  open.value = false
}

function showPanel(): void {
  emit('open')
  open.value = true
}

async function openTicket(): Promise<void> {
  close()
  await router.push('/support')
}
</script>

<style scoped>
.contact-trigger {
  @apply flex min-h-10 w-full items-center gap-3 rounded-control px-3 py-2 text-sm font-medium text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.contact-trigger-collapsed {
  @apply justify-center px-2;
}

.contact-trigger-icon {
  @apply flex-shrink-0;
}

.contact-trigger-label {
  @apply min-w-0 flex-1 truncate text-left transition-opacity;
}

.contact-trigger-label-collapsed {
  @apply hidden;
}

.contact-intro {
  @apply mb-4 text-sm leading-6 text-foreground-subtle;
}

.contact-actions {
  @apply space-y-2.5;
}

.contact-action {
  @apply flex min-h-[72px] items-center gap-3 rounded-lg border border-outline bg-surface px-3.5 py-3 text-foreground transition-colors hover:border-outline-strong hover:bg-surface-subtle focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus/40;
}

.contact-icon {
  @apply flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-lg;
}

.contact-icon-qq {
  @apply bg-info-subtle text-info-foreground;
}

.contact-icon-telegram {
  @apply bg-brand-subtle text-brand;
}

.contact-icon-ticket {
  @apply bg-warning-subtle text-warning-foreground;
}

.contact-copy {
  @apply min-w-0 flex-1;
}

.contact-copy strong,
.contact-copy span {
  @apply block;
}

.contact-copy strong {
  @apply break-words text-sm font-semibold;
}

.contact-copy span {
  @apply mt-0.5 text-xs leading-5 text-foreground-subtle;
}

.contact-external {
  @apply flex-shrink-0 text-foreground-subtle;
}

@media (max-width: 420px) {
  .contact-action {
    @apply items-start;
  }
}
</style>
