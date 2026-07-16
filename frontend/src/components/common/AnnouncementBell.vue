<template>
  <div>
    <button
      type="button"
      class="relative inline-flex h-10 w-10 items-center justify-center rounded-control text-foreground-muted transition-colors hover:bg-surface-subtle hover:text-foreground"
      :class="{ 'text-foreground': unreadCount > 0 }"
      :aria-label="localText('通知', 'Notifications')"
      :aria-expanded="isModalOpen"
      :title="localText('通知', 'Notifications')"
      @click="openModal"
    >
      <Icon name="bell" size="md" />
      <span
        v-if="unreadCount > 0"
        class="absolute right-0 top-0 inline-flex min-h-4 min-w-4 items-center justify-center rounded-full bg-danger px-1 text-[9px] font-bold leading-4 text-white"
        aria-hidden="true"
      >
        {{ unreadCount > 99 ? '99+' : unreadCount }}
      </span>
      <span v-if="unreadCount > 0" class="sr-only">
        {{ t('announcements.unread') }}: {{ unreadCount }}
      </span>
    </button>

    <BaseDialog
      :show="isModalOpen"
      :title="localText('通知', 'Notifications')"
      width="normal"
      :z-index="100"
      @close="closeModal"
    >
      <div class="min-w-0">
        <div v-if="unreadCount > 0" class="mb-3 flex items-center justify-between gap-3 border-b border-outline pb-3">
          <p class="text-sm text-foreground-muted">
            <span class="font-semibold text-foreground">{{ unreadCount }}</span>
            {{ t('announcements.unread') }}
          </p>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="loading"
            @click="markAllAsRead"
          >
            <Icon name="check" size="sm" />
            {{ t('announcements.markAllRead') }}
          </button>
        </div>

        <div v-if="loading" class="flex min-h-48 items-center justify-center" role="status" :aria-label="t('common.loading')">
          <Icon name="refresh" size="lg" class="animate-spin text-foreground-subtle" />
        </div>

        <ul v-else-if="ticketItems.length > 0 || announcements.length > 0" class="max-h-[58vh] divide-y divide-outline overflow-y-auto">
          <li v-for="ticket in ticketItems" :key="`ticket-${ticket.id}`">
            <button type="button" class="group flex min-h-[68px] w-full items-center gap-3 px-1 py-3 text-left transition-colors hover:bg-surface-subtle" @click="openTicket(ticket.number)">
              <span class="inline-flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-control border border-outline bg-info-subtle text-info-foreground"><Icon name="clipboard" size="sm" /></span>
              <span class="min-w-0 flex-1"><span class="block truncate text-sm font-semibold text-foreground">{{ ticket.subject }}</span><span class="mt-1 block text-xs text-foreground-subtle">{{ ticket.number }} · {{ formatRelativeTime(ticket.last_message_at) }}</span></span>
              <Icon name="chevronRight" size="sm" class="text-foreground-subtle" />
            </button>
          </li>
          <li v-for="item in announcements" :key="item.id">
            <button
              type="button"
              class="group flex min-h-[68px] w-full items-center gap-3 px-1 py-3 text-left transition-colors hover:bg-surface-subtle"
              @click="openDetail(item)"
            >
              <span
                class="inline-flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-control border border-outline"
                :class="item.read_at ? 'bg-surface text-foreground-subtle' : 'bg-info-subtle text-info-foreground'"
                aria-hidden="true"
              >
                <Icon :name="item.read_at ? 'checkCircle' : 'infoCircle'" size="sm" />
              </span>
              <span class="min-w-0 flex-1">
                <span class="flex items-center gap-2">
                  <span class="truncate text-sm font-semibold text-foreground">{{ item.title }}</span>
                  <span v-if="!item.read_at" class="h-1.5 w-1.5 flex-shrink-0 rounded-full bg-info"></span>
                </span>
                <time class="mt-1 block text-xs text-foreground-subtle">
                  {{ formatRelativeTime(item.created_at) }}
                </time>
              </span>
              <Icon name="chevronRight" size="sm" class="flex-shrink-0 text-foreground-subtle transition-transform group-hover:translate-x-0.5" />
            </button>
          </li>
        </ul>

        <div v-else class="flex min-h-48 flex-col items-center justify-center text-center">
          <span class="mb-3 inline-flex h-12 w-12 items-center justify-center rounded-panel bg-surface-subtle text-foreground-subtle">
            <Icon name="inbox" size="lg" />
          </span>
          <p class="text-sm font-semibold text-foreground">{{ t('announcements.empty') }}</p>
          <p class="mt-1 text-xs text-foreground-muted">{{ t('announcements.emptyDescription') }}</p>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <button type="button" class="btn btn-secondary" @click="closeModal">
            {{ t('common.close') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="detailModalOpen && !!selectedAnnouncement"
      :title="selectedAnnouncement?.title || t('announcements.title')"
      width="wide"
      :z-index="110"
      @close="closeDetail"
    >
      <div v-if="selectedAnnouncement" class="min-w-0">
        <div class="mb-4 flex flex-wrap items-center gap-2 text-xs text-foreground-muted">
          <span class="inline-flex items-center gap-1.5 rounded-control bg-surface-subtle px-2 py-1 font-semibold text-foreground">
            <Icon name="bell" size="xs" />
            {{ t('announcements.title') }}
          </span>
          <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
          <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
        </div>

        <div
          class="announcement-markdown max-h-[58vh] overflow-y-auto pr-2"
          v-html="renderMarkdown(selectedAnnouncement.content)"
        ></div>
      </div>

      <template #footer>
        <div class="flex w-full flex-wrap items-center justify-between gap-3">
          <span class="text-xs text-foreground-muted">
            {{ selectedAnnouncement?.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}
          </span>
          <div class="flex gap-2">
            <button type="button" class="btn btn-secondary" @click="closeDetail">
              {{ t('common.close') }}
            </button>
            <button
              v-if="selectedAnnouncement && !selectedAnnouncement.read_at"
              type="button"
              class="btn btn-primary"
              @click="markAsReadAndClose(selectedAnnouncement.id)"
            >
              <Icon name="check" size="sm" />
              {{ t('announcements.markRead') }}
            </button>
          </div>
        </div>
      </template>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { useTicketNotificationStore } from '@/stores/ticketNotifications'
import { useAuthStore } from '@/stores/auth'
import { ticketsAPI } from '@/api/tickets'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { UserAnnouncement } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t, locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()
const ticketStore = useTicketNotificationStore()
const authStore = useAuthStore()

marked.setOptions({
  breaks: true,
  gfm: true,
})

const { announcements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount + ticketStore.unreadCount)
const ticketItems = computed(() => ticketStore.items)
const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)

function renderMarkdown(content: string): string {
  if (!content) return ''
  return DOMPurify.sanitize(marked.parse(content) as string)
}

function openModal() {
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) void markAsRead(announcement.id)
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    await announcementStore.markAsRead(id)
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.unknownError')))
  }
}

async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    await Promise.all([
      announcementStore.markAllAsRead(),
      ...ticketItems.value.map((ticket) => authStore.isAdmin ? adminTicketsAPI.markRead(ticket.number) : ticketsAPI.markRead(ticket.number))
    ])
    await ticketStore.refresh()
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.unknownError')))
  }
}

async function openTicket(number: string) {
  closeModal()
  await router.push(authStore.isAdmin ? `/admin/tickets/${number}` : `/support/${number}`)
}
</script>

<style scoped>
.announcement-markdown {
  color: var(--ui-text-muted);
  overflow-wrap: anywhere;
}

.announcement-markdown :deep(h1),
.announcement-markdown :deep(h2),
.announcement-markdown :deep(h3),
.announcement-markdown :deep(h4) {
  margin: 1.25rem 0 0.5rem;
  color: var(--ui-text);
  font-weight: 700;
  letter-spacing: 0;
}

.announcement-markdown :deep(h1) {
  font-size: 1.25rem;
}

.announcement-markdown :deep(h2) {
  font-size: 1.125rem;
}

.announcement-markdown :deep(h3),
.announcement-markdown :deep(h4) {
  font-size: 1rem;
}

.announcement-markdown :deep(p),
.announcement-markdown :deep(ul),
.announcement-markdown :deep(ol) {
  margin-bottom: 0.875rem;
  font-size: 0.875rem;
  line-height: 1.7;
}

.announcement-markdown :deep(ul),
.announcement-markdown :deep(ol) {
  padding-left: 1.25rem;
}

.announcement-markdown :deep(a) {
  color: rgb(var(--color-info));
  text-decoration: underline;
  text-underline-offset: 2px;
}

.announcement-markdown :deep(blockquote) {
  margin: 1rem 0;
  padding: 0.75rem 1rem;
  border-left: 3px solid rgb(var(--color-info));
  background: var(--ui-surface-subtle);
}

.announcement-markdown :deep(code) {
  border-radius: 4px;
  background: var(--ui-surface-subtle);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.8125rem;
}

.announcement-markdown :deep(pre) {
  max-width: 100%;
  overflow-x: auto;
  padding: 0.875rem;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-surface-subtle);
}

.announcement-markdown :deep(pre code) {
  background: transparent;
}

.announcement-markdown :deep(table) {
  width: 100%;
  min-width: 560px;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.announcement-markdown :deep(th),
.announcement-markdown :deep(td) {
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--ui-border);
  text-align: left;
}

.announcement-markdown :deep(th) {
  background: var(--ui-surface-subtle);
  color: var(--ui-text);
  font-weight: 650;
}

.announcement-markdown :deep(img) {
  max-width: 100%;
  height: auto;
  margin: 1rem 0;
  border-radius: 8px;
}
</style>
