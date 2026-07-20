<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1200px] space-y-4">
      <header class="page-header">
        <div class="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-start gap-3">
          <h1 class="page-title min-w-0">{{ localText('工单支持', 'Support tickets') }}</h1>
          <button class="btn btn-primary shrink-0" type="button" @click="showCreate = true"><Icon name="plus" size="sm" />{{ localText('创建工单', 'Create ticket') }}</button>
        </div>
        <p class="page-description">{{ localText('遇到充值、模型、延迟或账户问题时，在这里和后台沟通。', 'Contact support about billing, models, latency, or account issues.') }}</p>
      </header>
      <section class="rounded-panel border border-outline bg-surface p-3 shadow-card">
        <div class="grid grid-cols-[minmax(0,1fr)_104px_40px] gap-2 sm:flex">
          <input v-model="search" class="input min-w-0 sm:flex-1" :placeholder="localText('搜索工单号或问题', 'Search tickets')" @keyup.enter="load">
          <select v-model="status" class="input min-w-0 sm:w-40" @change="load"><option value="">{{ localText('全部状态', 'All statuses') }}</option><option value="open">{{ localText('待处理', 'Open') }}</option><option value="answered">{{ localText('已回复', 'Answered') }}</option><option value="closed">{{ localText('已关闭', 'Closed') }}</option></select>
          <button class="btn btn-secondary btn-icon" :title="localText('刷新', 'Refresh')" :aria-label="localText('刷新', 'Refresh')" @click="load"><Icon name="refresh" size="sm" /></button>
        </div>
      </section>
      <section class="space-y-2">
        <button v-for="ticket in tickets" :key="ticket.id" type="button" class="w-full rounded-panel border border-outline bg-surface p-4 text-left shadow-card transition-colors hover:border-info" @click="router.push(`/support/${ticket.number}`)"><div class="flex items-center gap-3"><div class="min-w-0 flex-1"><div class="flex min-w-0 items-center gap-2"><span class="truncate text-xs font-semibold text-info-foreground">{{ ticket.number }}</span><span v-if="ticket.user_unread_count" class="inline-flex min-w-5 flex-none items-center justify-center rounded-full bg-danger px-1.5 py-0.5 text-[10px] font-bold text-danger-solid-foreground">{{ ticket.user_unread_count }}</span></div><h2 class="mt-1 truncate text-sm font-semibold text-foreground">{{ ticket.subject }}</h2><p class="mt-2 text-xs text-foreground-muted">{{ categoryName(ticket.category) }} · {{ formatDate(ticket.last_message_at) }}</p></div><div class="flex shrink-0 flex-col items-end gap-2"><span class="status-badge" :class="statusClass(ticket.status)">{{ statusLabel(ticket.status) }}</span><Icon name="chevronRight" size="sm" class="text-foreground-subtle" aria-hidden="true" /></div></div></button>
        <div v-if="loading" class="flex min-h-28 items-center justify-center gap-2 rounded-panel border border-outline bg-surface text-sm text-foreground-muted"><Icon name="refresh" size="sm" class="animate-spin" />{{ localText('正在加载工单', 'Loading tickets') }}</div>
        <div v-if="!loading && tickets.length === 0" class="rounded-panel border border-dashed border-outline p-12 text-center text-sm text-foreground-muted">{{ localText('还没有工单', 'No tickets yet') }}</div>
      </section>
      <nav v-if="pages > 1" class="flex items-center justify-end gap-2" :aria-label="localText('分页', 'Pagination')"><button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="page--; load()"><Icon name="chevronLeft" size="sm" /></button><span class="text-sm text-foreground-muted">{{ page }} / {{ pages }}</span><button class="btn btn-secondary btn-sm" :disabled="page >= pages" @click="page++; load()"><Icon name="chevronRight" size="sm" /></button></nav>
    </div>
    <BaseDialog :show="showCreate" :title="localText('告诉我们你遇到的问题', 'Tell us what happened')" width="wide" @close="showCreate = false">
      <form id="ticket-create-form" class="space-y-4" @submit.prevent="create">
        <div><label class="input-label">{{ localText('问题分类', 'Category') }}</label><select v-model.number="categoryId" class="input" required><option :value="0" disabled>{{ localText('请选择', 'Select a category') }}</option><option v-for="category in categories" :key="category.id" :value="category.id">{{ categoryName(category) }}</option></select></div>
        <div><label class="input-label">{{ localText('问题描述', 'Description') }}</label><textarea v-model="description" class="input min-h-44" maxlength="10000" required></textarea></div>
        <div v-if="capabilities.attachments_available" class="space-y-2"><label class="btn btn-secondary cursor-pointer"><Icon name="upload" size="sm" />{{ localText('上传图片 / PDF / 日志', 'Upload images / PDF / logs') }}<input class="sr-only" type="file" multiple :accept="capabilities.allowed_extensions.join(',')" @change="selectFiles"></label><p class="text-xs text-foreground-subtle">{{ attachmentLimitText }}</p></div>
        <div v-if="files.length" class="flex flex-wrap gap-2"><span v-for="(file, index) in files" :key="`${file.name}-${index}`" class="inline-flex max-w-full items-center gap-1.5 rounded-control bg-surface-subtle px-2 py-1 text-xs text-foreground-muted"><span class="max-w-52 truncate">{{ file.name }}</span><span class="flex-none text-foreground-subtle">{{ formatBytes(file.size) }}</span><button type="button" :aria-label="localText('移除附件', 'Remove attachment')" @click="files.splice(index, 1)"><Icon name="x" size="xs" /></button></span></div>
        <div v-if="creating && files.length" class="space-y-1"><div class="flex justify-between text-xs text-foreground-muted"><span>{{ localText('正在上传', 'Uploading') }}</span><span>{{ uploadProgress }}%</span></div><progress class="h-2 w-full accent-info" max="100" :value="uploadProgress"></progress></div>
      </form>
      <template #footer><button class="btn btn-secondary" @click="showCreate = false">{{ localText('取消', 'Cancel') }}</button><button form="ticket-create-form" class="btn btn-primary" :disabled="creating || !categoryId || !description.trim()">{{ localText('提交工单', 'Submit ticket') }}</button></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { ticketsAPI } from '@/api/tickets'
import { useAppStore } from '@/stores/app'
import type { Ticket, TicketAttachmentCapabilities, TicketCategory, TicketStatus } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'

const router = useRouter(); const { locale } = useI18n(); const app = useAppStore()
const tickets = ref<Ticket[]>([]); const categories = ref<TicketCategory[]>([]); const loading = ref(false)
const search = ref(''); const status = ref(''); const showCreate = ref(false); const categoryId = ref(0); const description = ref(''); const files = ref<File[]>([]); const creating = ref(false); const uploadProgress = ref(0)
const page = ref(1); const pages = ref(1)
const capabilities = ref<TicketAttachmentCapabilities>({ attachments_available: false, max_file_bytes: 0, max_files_per_message: 0, max_total_bytes: 0, allowed_extensions: [] })
const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const formatBytes = (value: number) => value < 1024 ? `${value} B` : value < 1024 * 1024 ? `${(value / 1024).toFixed(1)} KB` : `${(value / 1024 / 1024).toFixed(1)} MB`
const attachmentLimitText = computed(() => localText(`最多 ${capabilities.value.max_files_per_message} 个文件，单个不超过 ${formatBytes(capabilities.value.max_file_bytes)}`, `Up to ${capabilities.value.max_files_per_message} files, ${formatBytes(capabilities.value.max_file_bytes)} each`))
const categoryName = (category: TicketCategory) => locale.value.startsWith('zh') ? category.name_zh : (category.name_en || category.name_zh)
const statusLabel = (value: TicketStatus) => ({ open: localText('待处理', 'Open'), answered: localText('已回复', 'Answered'), closed: localText('已关闭', 'Closed') }[value])
const statusClass = (value: TicketStatus) => value === 'open' ? 'status-badge-warning' : value === 'answered' ? 'status-badge-success' : 'status-badge-neutral'
const formatDate = (value: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))
function selectFiles(event: Event) { const input = event.target as HTMLInputElement; const selected = Array.from(input.files || []); if (files.value.length + selected.length > capabilities.value.max_files_per_message) { app.showError(localText('附件数量超过限制', 'Too many attachments')); input.value = ''; return } const total = [...files.value, ...selected].reduce((sum, file) => sum + file.size, 0); if (selected.some(file => file.size > capabilities.value.max_file_bytes) || total > capabilities.value.max_total_bytes) { app.showError(localText('附件大小超过限制', 'Attachment size limit exceeded')); input.value = ''; return } files.value.push(...selected); input.value = '' }
async function load() { loading.value = true; try { const result = await ticketsAPI.list({ page: page.value, page_size: 20, status: status.value, search: search.value }); tickets.value = result.items; pages.value = result.pages } catch (error) { app.showError(extractApiErrorMessage(error, localText('加载失败', 'Load failed'))) } finally { loading.value = false } }
async function create() { creating.value = true; uploadProgress.value = 0; try { const item = await ticketsAPI.create(categoryId.value, description.value, files.value, value => { uploadProgress.value = value }); showCreate.value = false; await router.push(`/support/${item.number}`) } catch (error) { app.showError(extractApiErrorMessage(error, localText('创建失败', 'Create failed'))) } finally { creating.value = false } }
onMounted(async () => { [categories.value, capabilities.value] = await Promise.all([ticketsAPI.categories(), ticketsAPI.capabilities()]); await load() })
</script>
