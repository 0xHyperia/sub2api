<template>
  <div class="min-w-0 space-y-3">
    <section v-if="ticket" class="rounded-panel border border-outline bg-surface p-3 shadow-card sm:p-4">
      <div class="flex min-w-0 items-start justify-between gap-3">
        <div class="min-w-0 flex-1">
          <button type="button" class="btn btn-secondary btn-sm mb-3" @click="$emit('back')">
            <Icon name="arrowLeft" size="sm" />
            {{ localText('返回', 'Back') }}
          </button>
          <div class="break-all font-mono text-xs font-semibold text-info-foreground">{{ ticket.number }}</div>
          <h2 class="mt-1 break-words text-lg font-semibold text-foreground">{{ ticket.subject }}</h2>
          <p class="mt-1 min-w-0 break-words text-xs leading-5 text-foreground-muted">
            {{ categoryName(ticket.category) }} · {{ formatDate(ticket.created_at) }}
            <span v-if="admin" class="break-all"> · {{ ticket.user_email }}</span>
          </p>
        </div>
        <span class="status-badge flex-none" :class="statusClass(ticket.status)">{{ statusLabel(ticket.status) }}</span>
      </div>
    </section>

    <section class="flex h-[calc(100dvh-268px)] min-h-[420px] max-h-[760px] min-w-0 flex-col overflow-hidden rounded-panel border border-outline bg-surface shadow-card sm:h-[calc(100dvh-276px)]">
      <div class="min-w-0 border-b border-outline px-3 py-2.5 sm:px-4">
        <div class="min-w-0">
          <h3 class="text-sm font-semibold text-foreground">{{ localText('对话记录', 'Conversation') }}</h3>
          <p v-if="ticket" class="mt-0.5 text-xs text-foreground-subtle">
            {{ localText(`${messages.length} 条消息`, `${messages.length} messages`) }}
          </p>
        </div>
      </div>

      <div class="relative min-h-0 flex-1">
        <div
          ref="messageViewport"
          class="h-full overflow-y-auto overscroll-contain px-3 py-4 sm:px-4"
          data-test="ticket-message-viewport"
          @scroll.passive="handleScroll"
        >
          <div v-if="loading" class="flex min-h-full items-center justify-center">
            <Icon name="refresh" size="lg" class="animate-spin" />
          </div>
          <div v-else-if="ticket && messages.length === 0" class="flex min-h-full items-center justify-center text-sm text-foreground-muted">
            {{ localText('暂无消息', 'No messages yet') }}
          </div>
          <div v-else-if="ticket" class="space-y-4" aria-live="polite">
            <template v-for="message in messages" :key="message.id">
              <div v-if="message.id === firstNewMessageId" class="flex items-center gap-3 py-1" data-test="new-message-divider">
                <span class="h-px flex-1 bg-info"></span>
                <span class="text-xs font-semibold text-info-foreground">{{ localText('以下为新消息', 'New messages') }}</span>
                <span class="h-px flex-1 bg-info"></span>
              </div>

              <div v-if="message.sender_type === 'system'" class="flex items-center gap-3 py-1">
                <span class="h-px flex-1 bg-outline"></span>
                <span class="rounded-full bg-surface-subtle px-3 py-1 text-xs text-foreground-muted">{{ eventLabel(message.event_type) }}</span>
                <span class="h-px flex-1 bg-outline"></span>
              </div>

              <article
                v-else
                class="flex min-w-0 items-end gap-2"
                :class="isMine(message.sender_type) ? 'flex-row-reverse' : 'flex-row'"
                :data-test="`ticket-message-${message.id}`"
              >
                <span
                  class="flex h-8 w-8 flex-none items-center justify-center rounded-full text-xs font-bold"
                  :class="message.sender_type === 'admin' ? 'bg-info text-info-solid-foreground' : 'bg-surface-subtle text-foreground-muted'"
                  aria-hidden="true"
                >
                  {{ senderInitial(message.sender_type) }}
                </span>
                <div
                  class="min-w-0 max-w-[82%] rounded-panel border px-3 py-2.5 sm:max-w-[72%]"
                  :class="message.sender_type === 'admin' ? 'border-info bg-info-subtle' : 'border-outline bg-canvas'"
                >
                  <div class="mb-1.5 flex min-w-0 flex-wrap items-center gap-x-3 gap-y-0.5 text-xs text-foreground-muted">
                    <span class="font-semibold text-foreground">{{ senderLabel(message.sender_type) }}</span>
                    <time class="flex-none">{{ formatDate(message.created_at) }}</time>
                  </div>
                  <p v-if="message.content" class="whitespace-pre-wrap break-words text-sm leading-6 text-foreground">{{ message.content }}</p>
                  <div v-if="message.attachments.length" class="mt-3 flex flex-wrap gap-2">
                    <button
                      v-for="attachment in message.attachments"
                      :key="attachment.id"
                      type="button"
                      class="btn btn-secondary btn-sm max-w-full"
                      @click="download(attachment.id)"
                    >
                      <Icon name="download" size="sm" />
                      <span class="truncate">{{ attachment.original_name }}</span>
                      <span class="text-xs text-foreground-subtle">{{ formatBytes(attachment.size_bytes) }}</span>
                    </button>
                  </div>
                </div>
              </article>
            </template>
          </div>
        </div>

        <button
          v-if="unseenMessageCount > 0"
          type="button"
          class="btn btn-primary btn-sm absolute bottom-3 left-1/2 z-10 -translate-x-1/2 shadow-card"
          data-test="jump-to-latest"
          @click="jumpToLatest"
        >
          <Icon name="arrowDown" size="sm" />
          {{ localText(`跳到最新消息（${unseenMessageCount}）`, `Jump to latest (${unseenMessageCount})`) }}
        </button>
      </div>

      <div v-if="ticket" class="flex-none border-t border-outline bg-surface p-3 sm:p-4" data-test="ticket-composer">
        <textarea
          v-model="content"
          class="input min-h-20 resize-y sm:min-h-24"
          :disabled="ticket.status === 'closed' || sending"
          maxlength="10000"
          :placeholder="ticket.status === 'closed' ? localText('重新打开工单后可继续回复', 'Reopen the ticket to continue') : localText('输入回复内容', 'Write a reply')"
        ></textarea>
        <div v-if="files.length" class="mt-2 flex flex-wrap gap-2">
          <span v-for="(file, index) in files" :key="`${file.name}-${index}`" class="inline-flex max-w-full items-center gap-1 rounded-control bg-surface-subtle px-2 py-1 text-xs">
            <span class="truncate">{{ file.name }}</span>
            <button type="button" :aria-label="localText('移除附件', 'Remove attachment')" @click="files.splice(index, 1)">
              <Icon name="x" size="xs" />
            </button>
          </span>
        </div>
        <div v-if="sending && files.length" class="mt-2 space-y-1">
          <div class="flex justify-between text-xs text-foreground-muted"><span>{{ localText('正在上传', 'Uploading') }}</span><span>{{ uploadProgress }}%</span></div>
          <progress class="h-2 w-full accent-info" max="100" :value="uploadProgress"></progress>
        </div>
        <div class="mt-3 flex min-w-0 items-center justify-between gap-2">
          <label class="btn btn-secondary btn-icon flex-none cursor-pointer" :class="{ 'pointer-events-none opacity-50': !capabilities.attachments_available || ticket.status === 'closed' }" :title="localText('添加附件', 'Add attachment')">
            <Icon name="upload" size="sm" />
            <span class="sr-only">{{ localText('添加附件', 'Add attachment') }}</span>
            <input class="sr-only" type="file" multiple :accept="accept" :disabled="!capabilities.attachments_available || ticket.status === 'closed'" @change="selectFiles">
          </label>
          <div class="flex min-w-0 items-center justify-end gap-2">
            <button v-if="ticket.status === 'closed'" type="button" class="btn btn-secondary" :disabled="sending" @click="changeStatus(true)">{{ localText('重新打开', 'Reopen') }}</button>
            <button v-else type="button" class="btn btn-secondary" :disabled="sending" @click="changeStatus(false)">{{ localText('关闭工单', 'Close ticket') }}</button>
            <button type="button" class="btn btn-primary btn-icon sm:w-auto sm:px-4" :title="localText('发送', 'Send')" :disabled="sending || ticket.status === 'closed' || (!content.trim() && files.length === 0)" @click="send">
              <Icon name="arrowRight" size="sm" />
              <span class="sr-only sm:not-sr-only">{{ localText('发送', 'Send') }}</span>
            </button>
          </div>
        </div>
        <p v-if="!capabilities.attachments_available" class="mt-2 text-xs text-foreground-subtle">{{ localText('附件存储未配置，当前仅支持文本消息。', 'Attachment storage is unavailable; text messages remain available.') }}</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTicketNotificationStore } from '@/stores/ticketNotifications'
import { ticketsAPI } from '@/api/tickets'
import { adminTicketsAPI } from '@/api/admin/tickets'
import type { Ticket, TicketAttachmentCapabilities, TicketCategory, TicketMessage, TicketStatus } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{ number: string; admin?: boolean }>()
defineEmits<{ back: [] }>()

const { locale } = useI18n()
const appStore = useAppStore()
const notifications = useTicketNotificationStore()
const ticket = ref<Ticket | null>(null)
const loading = ref(true)
const sending = ref(false)
const content = ref('')
const files = ref<File[]>([])
const uploadProgress = ref(0)
const messageViewport = ref<HTMLElement | null>(null)
const firstNewMessageId = ref<number | null>(null)
const unseenMessageCount = ref(0)
const capabilities = ref<TicketAttachmentCapabilities>({ attachments_available: false, max_file_bytes: 10 * 1024 * 1024, max_files_per_message: 5, max_total_bytes: 25 * 1024 * 1024, allowed_extensions: [] })

let timer: ReturnType<typeof setInterval> | null = null
let polling = false
let markingRead = false

const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const accept = computed(() => capabilities.value.allowed_extensions.join(','))
const messages = computed<TicketMessage[]>(() => ticket.value?.messages || [])

function isNearBottom() {
  const viewport = messageViewport.value
  if (!viewport) return true
  return viewport.scrollHeight - viewport.scrollTop - viewport.clientHeight < 80
}

function scrollToLatest(behavior: 'auto' | 'smooth' = 'auto') {
  const viewport = messageViewport.value
  if (!viewport) return
  if (typeof viewport.scrollTo === 'function') viewport.scrollTo({ top: viewport.scrollHeight, behavior })
  else viewport.scrollTop = viewport.scrollHeight
}

function clearNewMessages() {
  firstNewMessageId.value = null
  unseenMessageCount.value = 0
}

async function markRead() {
  if (markingRead) return
  markingRead = true
  try {
    await (props.admin ? adminTicketsAPI.markRead(props.number) : ticketsAPI.markRead(props.number))
    void notifications.refresh()
  } catch {
    // A failed read receipt must not hide an otherwise successfully loaded conversation.
  } finally {
    markingRead = false
  }
}

async function load(silent = false) {
  if (silent && polling) return
  if (!silent) loading.value = true
  if (silent) polling = true

  const previousIds = new Set(messages.value.map(message => message.id))
  const wasNearBottom = isNearBottom()

  try {
    const nextTicket = props.admin ? await adminTicketsAPI.get(props.number) : await ticketsAPI.get(props.number)
    const appendedMessages = (nextTicket.messages || []).filter(message => !previousIds.has(message.id))
    ticket.value = nextTicket
    if (!silent) loading.value = false
    await nextTick()

    if (!silent) {
      clearNewMessages()
      scrollToLatest()
      await markRead()
    } else if (appendedMessages.length > 0) {
      if (wasNearBottom) {
        clearNewMessages()
        scrollToLatest('smooth')
        await markRead()
      } else {
        firstNewMessageId.value ||= appendedMessages[0].id
        unseenMessageCount.value += appendedMessages.length
      }
    }
  } catch (error) {
    if (!silent) appStore.showError(extractApiErrorMessage(error, localText('加载工单失败', 'Failed to load ticket')))
  } finally {
    loading.value = false
    if (silent) polling = false
  }
}

async function send() {
  sending.value = true
  uploadProgress.value = 0
  const onProgress = (value: number) => { uploadProgress.value = value }
  try {
    ticket.value = props.admin
      ? await adminTicketsAPI.reply(props.number, content.value, files.value, onProgress)
      : await ticketsAPI.reply(props.number, content.value, files.value, onProgress)
    content.value = ''
    files.value = []
    clearNewMessages()
    await nextTick()
    scrollToLatest('smooth')
    appStore.showSuccess(localText('回复已发送', 'Reply sent'))
    void notifications.refresh()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, localText('发送失败', 'Send failed')))
  } finally {
    sending.value = false
  }
}

async function changeStatus(reopen: boolean) {
  sending.value = true
  try {
    ticket.value = props.admin
      ? (reopen ? await adminTicketsAPI.reopen(props.number) : await adminTicketsAPI.close(props.number))
      : (reopen ? await ticketsAPI.reopen(props.number) : await ticketsAPI.close(props.number))
    await nextTick()
    scrollToLatest('smooth')
    appStore.showSuccess(reopen ? localText('工单已重新打开', 'Ticket reopened') : localText('工单已关闭', 'Ticket closed'))
    void notifications.refresh()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, localText('操作失败', 'Operation failed')))
  } finally {
    sending.value = false
  }
}

function selectFiles(event: Event) {
  const input = event.target as HTMLInputElement
  const selected = Array.from(input.files || [])
  if (files.value.length + selected.length > capabilities.value.max_files_per_message) {
    appStore.showError(localText('附件数量超过限制', 'Too many attachments'))
    input.value = ''
    return
  }
  const total = [...files.value, ...selected].reduce((sum, file) => sum + file.size, 0)
  if (selected.some(file => file.size > capabilities.value.max_file_bytes) || total > capabilities.value.max_total_bytes) {
    appStore.showError(localText('附件大小超过限制', 'Attachment size limit exceeded'))
    input.value = ''
    return
  }
  files.value.push(...selected)
  input.value = ''
}

async function download(id: number) {
  try {
    const url = await ticketsAPI.attachmentURL(id)
    window.open(url, '_blank', 'noopener,noreferrer')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, localText('下载失败', 'Download failed')))
  }
}

function handleScroll() {
  if (unseenMessageCount.value > 0 && isNearBottom()) {
    clearNewMessages()
    void markRead()
  }
}

function jumpToLatest() {
  scrollToLatest('smooth')
  clearNewMessages()
  void markRead()
}

const isMine = (sender: string) => props.admin ? sender === 'admin' : sender === 'user'
const senderLabel = (sender: string) => sender === 'admin' ? localText('客服', 'Support') : localText('用户', 'Customer')
const senderInitial = (sender: string) => sender === 'admin' ? localText('客', 'S') : localText('用', 'C')
const eventLabel = (event: string) => event === 'ticket_reopened' ? localText('工单已重新打开', 'Ticket reopened') : localText('工单已关闭', 'Ticket closed')
const statusLabel = (status: TicketStatus) => ({ open: localText('待处理', 'Open'), answered: localText('已回复', 'Answered'), closed: localText('已关闭', 'Closed') }[status])
const statusClass = (status: TicketStatus) => status === 'open' ? 'status-badge-warning' : status === 'answered' ? 'status-badge-success' : 'status-badge-neutral'
const categoryName = (category: TicketCategory) => locale.value.startsWith('zh') ? category.name_zh : (category.name_en || category.name_zh)
const formatDate = (value: string) => new Intl.DateTimeFormat(locale.value, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
const formatBytes = (value: number) => value < 1024 ? `${value} B` : value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB`

onMounted(async () => {
  capabilities.value = await ticketsAPI.capabilities().catch(() => capabilities.value)
  await load()
  timer = setInterval(() => {
    if (document.visibilityState === 'visible') void load(true)
  }, 5000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>
