<template>
  <AppLayout><div class="mx-auto w-full max-w-[1440px] space-y-4">
    <header class="page-header flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"><div><h1 class="page-title">{{ localText('工单队列', 'Ticket queue') }}</h1><p class="page-description">{{ localText('查看并回复所有用户工单。', 'Review and reply to user tickets.') }}</p></div><button type="button" class="btn btn-secondary" @click="openCategories"><Icon name="cog" size="sm" />{{ localText('管理分类', 'Manage categories') }}</button></header>
    <section class="rounded-panel border border-outline bg-surface p-3 shadow-card"><div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_160px_180px_auto]"><input v-model="search" class="input" :placeholder="localText('搜索工单号、标题或用户', 'Search ticket or user')" @keyup.enter="load"><select v-model="status" class="input" @change="load"><option value="">{{ localText('全部状态', 'All statuses') }}</option><option value="open">{{ localText('待处理', 'Open') }}</option><option value="answered">{{ localText('已回复', 'Answered') }}</option><option value="closed">{{ localText('已关闭', 'Closed') }}</option></select><select v-model.number="categoryId" class="input" @change="load"><option :value="0">{{ localText('全部分类', 'All categories') }}</option><option v-for="category in categories" :key="category.id" :value="category.id">{{ categoryName(category) }}</option></select><button class="btn btn-secondary" @click="load"><Icon name="refresh" size="sm" />{{ localText('刷新', 'Refresh') }}</button></div></section>
    <section class="overflow-hidden rounded-panel border border-outline bg-surface shadow-card"><div class="overflow-x-auto"><table class="w-full min-w-[800px] text-left text-sm"><thead class="bg-surface-subtle text-xs text-foreground-muted"><tr><th class="px-4 py-3">{{ localText('工单', 'Ticket') }}</th><th class="px-4 py-3">{{ localText('用户', 'User') }}</th><th class="px-4 py-3">{{ localText('分类', 'Category') }}</th><th class="px-4 py-3">{{ localText('状态', 'Status') }}</th><th class="px-4 py-3">{{ localText('最后活动', 'Last activity') }}</th></tr></thead><tbody class="divide-y divide-outline"><tr v-for="ticket in tickets" :key="ticket.id" class="cursor-pointer hover:bg-surface-subtle" @click="router.push(`/admin/tickets/${ticket.number}`)"><td class="px-4 py-3"><div class="font-semibold text-info-foreground">{{ ticket.number }}</div><div class="mt-1 max-w-xl truncate text-foreground">{{ ticket.subject }}</div></td><td class="px-4 py-3"><div>{{ ticket.user_name || '-' }}</div><div class="text-xs text-foreground-muted">{{ ticket.user_email }}</div></td><td class="px-4 py-3">{{ categoryName(ticket.category) }}</td><td class="px-4 py-3"><span class="status-badge" :class="statusClass(ticket.status)">{{ statusLabel(ticket.status) }}</span><span v-if="ticket.admin_unread_count" class="ml-2 rounded-full bg-danger px-1.5 py-0.5 text-[10px] font-bold text-white">{{ ticket.admin_unread_count }}</span></td><td class="px-4 py-3 text-foreground-muted">{{ formatDate(ticket.last_message_at) }}</td></tr><tr v-if="!loading && tickets.length === 0"><td colspan="5" class="px-4 py-12 text-center text-foreground-muted">{{ localText('没有匹配的工单', 'No matching tickets') }}</td></tr></tbody></table></div></section>
    <nav v-if="pages > 1" class="flex items-center justify-end gap-2" :aria-label="localText('分页', 'Pagination')"><button class="btn btn-secondary btn-sm" :disabled="page <= 1" @click="page--; load()"><Icon name="chevronLeft" size="sm" /></button><span class="text-sm text-foreground-muted">{{ page }} / {{ pages }}</span><button class="btn btn-secondary btn-sm" :disabled="page >= pages" @click="page++; load()"><Icon name="chevronRight" size="sm" /></button></nav>
  </div>
  <BaseDialog :show="showCategories" :title="localText('工单分类', 'Ticket categories')" width="wide" @close="showCategories = false">
    <div class="space-y-3">
      <div v-for="category in categoryDrafts" :key="category.id" class="grid gap-2 rounded-control border border-outline p-3 sm:grid-cols-[1fr_1fr_100px_auto_auto]">
        <input v-model="category.name_zh" class="input" placeholder="中文名称">
        <input v-model="category.name_en" class="input" placeholder="English name">
        <input v-model.number="category.sort_order" type="number" class="input" min="0" :aria-label="localText('排序', 'Sort order')">
        <label class="inline-flex items-center gap-2 text-sm"><input v-model="category.active" type="checkbox">{{ localText('启用', 'Active') }}</label>
        <button class="btn btn-secondary btn-sm" @click="saveCategory(category)">{{ localText('保存', 'Save') }}</button>
      </div>
      <form class="grid gap-2 border-t border-outline pt-4 sm:grid-cols-[160px_1fr_1fr_auto]" @submit.prevent="createCategory">
        <input v-model="newCategory.code" class="input" pattern="[a-z][a-z0-9_-]+" placeholder="code" required>
        <input v-model="newCategory.name_zh" class="input" placeholder="中文名称" required>
        <input v-model="newCategory.name_en" class="input" placeholder="English name">
        <button class="btn btn-primary" type="submit"><Icon name="plus" size="sm" />{{ localText('添加', 'Add') }}</button>
      </form>
    </div>
  </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminTicketsAPI } from '@/api/admin/tickets'
import { useAppStore } from '@/stores/app'
import type { Ticket, TicketCategory, TicketStatus } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'

const router = useRouter(); const { locale } = useI18n(); const app = useAppStore()
const tickets = ref<Ticket[]>([]); const categories = ref<TicketCategory[]>([]); const categoryDrafts = ref<TicketCategory[]>([]); const loading = ref(false)
const search = ref(''); const status = ref(''); const categoryId = ref(0); const showCategories = ref(false)
const page = ref(1); const pages = ref(1)
const newCategory = reactive({ code: '', name_zh: '', name_en: '', active: true, sort_order: 0 })
const localText = (zh: string, en: string) => locale.value.startsWith('zh') ? zh : en
const categoryName = (category: TicketCategory) => locale.value.startsWith('zh') ? category.name_zh : (category.name_en || category.name_zh)
const statusLabel = (value: TicketStatus) => ({ open: localText('待处理', 'Open'), answered: localText('已回复', 'Answered'), closed: localText('已关闭', 'Closed') }[value])
const statusClass = (value: TicketStatus) => value === 'open' ? 'status-badge-warning' : value === 'answered' ? 'status-badge-success' : 'status-badge-neutral'
const formatDate = (value: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))
async function load() { loading.value = true; try { const result = await adminTicketsAPI.list({ page: page.value, page_size: 20, status: status.value, category_id: categoryId.value || undefined, search: search.value }); tickets.value = result.items; pages.value = result.pages } catch (error) { app.showError(extractApiErrorMessage(error, localText('加载失败', 'Load failed'))) } finally { loading.value = false } }
async function loadCategories() { categories.value = await adminTicketsAPI.categories(); categoryDrafts.value = categories.value.map(category => ({ ...category })) }
async function openCategories() { await loadCategories(); showCategories.value = true }
async function saveCategory(category: TicketCategory) { try { await adminTicketsAPI.updateCategory(category.id, category); await loadCategories(); app.showSuccess(localText('分类已保存', 'Category saved')) } catch (error) { app.showError(extractApiErrorMessage(error, localText('保存失败', 'Save failed'))) } }
async function createCategory() { try { newCategory.sort_order = categories.value.reduce((max, category) => Math.max(max, category.sort_order), 0) + 10; await adminTicketsAPI.createCategory(newCategory); Object.assign(newCategory, { code: '', name_zh: '', name_en: '', active: true, sort_order: 0 }); await loadCategories() } catch (error) { app.showError(extractApiErrorMessage(error, localText('创建失败', 'Create failed'))) } }
onMounted(async () => { await loadCategories(); await load() })
</script>
