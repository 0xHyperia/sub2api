<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-6">
      <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div><p class="text-xs font-semibold uppercase text-foreground-subtle">Content management</p><h1 class="mt-1 text-2xl font-semibold">软件中心</h1><p class="mt-1 text-sm text-foreground-muted">统一管理 GitHub Release、官网和自托管软件下载。</p></div>
        <div class="flex gap-2"><a class="btn btn-secondary" href="/download" target="_blank" rel="noopener noreferrer"><Icon name="externalLink" size="sm" />公开页面</a><button ref="addButton" class="btn btn-primary" type="button" @click="openCreate"><Icon name="plus" size="sm" />添加软件</button></div>
      </header>

      <section aria-labelledby="catalog-heading">
        <div class="mb-3 flex items-center justify-between"><div><h2 id="catalog-heading" class="font-semibold">已上架软件</h2><p class="mt-0.5 text-xs text-foreground-muted">{{ items.length }} 款软件</p></div><button class="icon-action" type="button" :disabled="loading" aria-label="刷新列表" title="刷新列表" @click="load"><Icon name="refresh" :class="{ 'animate-spin': loading }" /></button></div>
        <div v-if="loading && !items.length" class="empty-state">正在加载软件目录...</div>
        <div v-else-if="loadError" class="rounded-panel border border-danger/30 bg-danger-subtle p-4 text-sm text-danger-foreground">{{ loadError }}</div>
        <div v-else-if="!items.length" class="empty-state"><Icon name="grid" size="xl" class="mx-auto text-foreground-subtle" /><p class="mt-3 font-medium">还没有软件</p><button class="btn btn-primary mt-4" type="button" @click="openCreate">添加第一款软件</button></div>
        <div v-else class="overflow-hidden rounded-panel border border-outline bg-surface">
          <article v-for="(item, index) in items" :key="item.id" class="flex flex-col gap-4 p-4 sm:flex-row sm:items-center" :class="{ 'border-t border-outline': index }">
            <div class="flex min-w-0 flex-1 items-center gap-3"><img class="h-12 w-12 shrink-0 rounded-control border border-outline bg-surface-subtle object-cover" :src="item.logo_url || '/favicon.svg'" :alt="`${item.name} 图标`" /><div class="min-w-0"><div class="flex flex-wrap items-center gap-2"><h3 class="truncate text-sm font-semibold">{{ item.name }}</h3><span class="badge badge-gray">{{ item.source_type === 'github' ? 'GitHub' : '手动' }}</span><span v-if="item.featured" class="badge badge-warning">推荐</span><span :class="['badge', item.enabled ? 'badge-success' : 'badge-gray']">{{ item.enabled ? '展示中' : '已隐藏' }}</span></div><a class="mt-1 block truncate text-xs text-foreground-muted hover:text-primary" :href="item.source_url" target="_blank" rel="noopener noreferrer">{{ item.repository || item.source_url }}</a><p class="mt-1 text-xs text-foreground-subtle">{{ item.release?.tag_name || item.version || '未填写版本' }} · {{ platformText(item.supported_platforms) }}</p></div></div>
            <div class="flex items-center justify-between gap-2 border-t border-outline pt-3 sm:border-0 sm:pt-0"><Toggle :model-value="item.enabled" :disabled="busyId === item.id" :aria-label="`${item.enabled ? '隐藏' : '展示'} ${item.name}`" @update:model-value="toggleEnabled(item, $event)" /><div class="flex items-center gap-1"><button v-if="item.source_type === 'github'" class="icon-action" type="button" title="同步 Release" :disabled="busyId === item.id" @click="refresh(item)"><Icon name="refresh" size="sm" :class="{ 'animate-spin': busyId === item.id && busyAction === 'refresh' }" /></button><button class="icon-action" type="button" title="编辑" @click="openEdit(item)"><Icon name="edit" size="sm" /></button><button class="icon-action danger" type="button" title="移除" :disabled="busyId === item.id" @click="remove(item)"><Icon name="trash" size="sm" /></button></div></div>
          </article>
        </div>
      </section>
    </div>

    <div v-if="dialogOpen" class="dialog-backdrop" @mousedown.self="closeDialog">
      <form ref="dialog" class="dialog-panel" role="dialog" aria-modal="true" aria-labelledby="software-dialog-title" @submit.prevent="save">
        <header class="dialog-header"><div><h2 id="software-dialog-title" class="font-semibold">{{ editing ? `编辑 ${editing.name}` : '添加软件' }}</h2><p class="mt-1 text-xs text-foreground-muted">{{ dialogSubtitle }}</p></div><button class="icon-action" type="button" aria-label="关闭" @click="closeDialog"><Icon name="x" size="sm" /></button></header>

        <div v-if="!editing && !sourceType" class="source-picker">
          <button type="button" @click="chooseSource('github')"><span class="source-icon"><Icon name="link" size="lg" /></span><span><strong>从 GitHub 导入</strong><small>识别公开仓库的最新 Release 和安装包</small></span><Icon name="chevronRight" /></button>
          <button type="button" @click="chooseSource('manual')"><span class="source-icon"><Icon name="cog" size="lg" /></span><span><strong>手动配置</strong><small>适用于官网、自托管地址和非 GitHub 软件</small></span><Icon name="chevronRight" /></button>
        </div>

        <div v-else class="dialog-body">
          <button v-if="!editing" class="source-back" type="button" @click="sourceType = null"><Icon name="arrowLeft" size="sm" />重新选择来源</button>
          <div v-if="sourceType === 'github' && !editing" class="rounded-control border border-outline bg-surface-subtle p-4"><label class="form-label" for="github-url">GitHub 项目地址</label><div class="mt-2 flex gap-2"><input id="github-url" v-model.trim="repositoryUrl" class="input min-w-0 flex-1" type="url" required placeholder="https://github.com/owner/repository" /><button class="btn btn-secondary shrink-0" type="button" :disabled="identifying || !repositoryUrl" @click="identify">{{ identifying ? '识别中' : '识别' }}</button></div><p v-if="identifyError" class="mt-2 text-xs text-danger-foreground">{{ identifyError }}</p></div>

          <fieldset :disabled="sourceType === 'github' && !editing && !preview" class="space-y-4 disabled:opacity-50">
            <div class="grid gap-4 sm:grid-cols-2"><div><label class="form-label" for="software-name">显示名称</label><input id="software-name" v-model.trim="draft.name" class="input mt-1.5" required maxlength="120" /></div><div><label class="form-label" for="software-logo">图标地址</label><input id="software-logo" v-model.trim="draft.logo_url" class="input mt-1.5" type="url" /></div></div>
            <div><label class="form-label" for="software-description">简介</label><textarea id="software-description" v-model.trim="draft.description" class="input mt-1.5 min-h-20 resize-y" maxlength="2000"></textarea></div>

            <template v-if="sourceType === 'manual'">
              <div><label class="form-label" for="source-url">官网或来源页</label><input id="source-url" v-model.trim="draft.source_url" class="input mt-1.5" type="url" required placeholder="https://chatgpt.com/download/" /></div>
              <div class="grid gap-4 sm:grid-cols-2"><div><label class="form-label" for="version">当前版本</label><input id="version" v-model.trim="draft.version" class="input mt-1.5" required placeholder="例如 1.2.0" /></div><div><label class="form-label" for="published-at">发布日期</label><input id="published-at" v-model="draft.published_at" class="input mt-1.5" type="date" /></div></div>
              <div><label class="form-label" for="release-notes">更新日志（Markdown）</label><textarea id="release-notes" v-model="draft.release_notes" class="input mt-1.5 min-h-28 resize-y" placeholder="本版本的主要变化"></textarea></div>
              <div class="asset-editor"><div class="flex items-center justify-between"><div><h3 class="text-sm font-semibold">下载项</h3><p class="mt-0.5 text-xs text-foreground-muted">每个平台和架构可配置独立地址。</p></div><button class="btn btn-secondary" type="button" @click="addAsset"><Icon name="plus" size="sm" />添加下载项</button></div>
                <div v-for="(asset, index) in draft.asset_variants" :key="index" class="asset-row"><div class="asset-row-head"><strong>下载项 {{ index + 1 }}</strong><button class="icon-action danger" type="button" aria-label="删除下载项" @click="draft.asset_variants.splice(index, 1)"><Icon name="trash" size="sm" /></button></div><div class="grid gap-3 sm:grid-cols-2"><input v-model.trim="asset.label" class="input" required placeholder="显示名称，如 Windows 安装程序" /><input v-model.trim="asset.name" class="input" required placeholder="文件名" /><select v-model="asset.platform" class="input"><option value="windows">Windows</option><option value="macos">macOS</option><option value="linux">Linux</option><option value="android">Android</option></select><select v-model="asset.architecture" class="input"><option value="x64">x64</option><option value="arm64">ARM64</option><option value="universal">通用</option></select><input v-model.trim="asset.format" class="input" required placeholder="格式，如 EXE、DMG" /><select v-model="asset.kind" class="input"><option value="installer">安装包</option><option value="portable">便携版</option><option value="archive">归档</option></select></div><input v-model.trim="asset.url" class="input" type="url" required placeholder="https://... 下载地址" /><input v-model.trim="asset.accelerated_url" class="input" type="url" placeholder="可选：加速下载地址" /></div>
              </div>
            </template>

            <div class="grid grid-cols-2 gap-3"><label class="switch-field"><span>公开展示</span><Toggle v-model="draft.enabled" /></label><label class="switch-field"><span>核心推荐</span><Toggle v-model="draft.featured" /></label></div>
            <div><label class="form-label" for="sort-order">排序</label><input id="sort-order" v-model.number="draft.sort_order" class="input mt-1.5" type="number" min="-10000" max="10000" /></div>
          </fieldset>
        </div>

        <footer v-if="sourceType" class="dialog-footer"><button class="btn btn-secondary" type="button" @click="closeDialog">取消</button><button class="btn btn-primary" type="submit" :disabled="saving || (sourceType === 'github' && !editing && !preview)">{{ saving ? '保存中' : editing ? '保存修改' : '发布软件' }}</button></footer>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'
import { useAppStore } from '@/stores'
import { createSoftware, deleteSoftware, listSoftware, previewSoftware, refreshSoftware, updateSoftware, type SoftwareCenterItem, type SoftwareCenterPreview, type SoftwareDownloadAsset } from '@/api/admin/softwareCenter'

const appStore = useAppStore()
const items = ref<SoftwareCenterItem[]>([]); const loading = ref(false); const loadError = ref(''); const busyId = ref<number | null>(null); const busyAction = ref('')
const dialogOpen = ref(false); const sourceType = ref<'github' | 'manual' | null>(null); const editing = ref<SoftwareCenterItem | null>(null); const preview = ref<SoftwareCenterPreview | null>(null); const repositoryUrl = ref(''); const identifying = ref(false); const identifyError = ref(''); const saving = ref(false)
const addButton = ref<HTMLButtonElement | null>(null); const dialog = ref<HTMLFormElement | null>(null); let dialogTrigger: HTMLElement | null = null
const emptyDraft = () => ({ name: '', description: '', logo_url: '', source_url: '', version: '', release_name: '', release_notes: '', published_at: '', enabled: true, featured: false, sort_order: (items.value.at(-1)?.sort_order || 0) + 10, asset_variants: [] as SoftwareDownloadAsset[] })
const draft = reactive(emptyDraft())
const dialogSubtitle = computed(() => editing.value ? (sourceType.value === 'github' ? editing.value.repository : editing.value.source_url) : sourceType.value === 'github' ? '导入 GitHub 项目' : sourceType.value === 'manual' ? '配置官网或托管下载' : '请选择软件来源')

function errorMessage(error: unknown) { return (error as { message?: string })?.message || '请求失败，请稍后重试' }
function platformText(value: string[]) { const labels: Record<string, string> = { windows: 'Windows', macos: 'macOS', linux: 'Linux', android: 'Android' }; return value.map(v => labels[v] || v).join(' / ') || '暂无下载项' }
function resetDraft() { Object.assign(draft, emptyDraft()) }
async function load() { loading.value = true; loadError.value = ''; try { items.value = await listSoftware() } catch (error) { loadError.value = errorMessage(error) } finally { loading.value = false } }
function openCreate(event?: Event) { dialogTrigger = event?.currentTarget as HTMLElement || document.activeElement as HTMLElement; editing.value = null; sourceType.value = null; preview.value = null; repositoryUrl.value = ''; resetDraft(); dialogOpen.value = true }
function chooseSource(value: 'github' | 'manual') { sourceType.value = value; if (value === 'manual' && !draft.asset_variants.length) addAsset() }
function openEdit(item: SoftwareCenterItem, event?: Event) { dialogTrigger = event?.currentTarget as HTMLElement || document.activeElement as HTMLElement; editing.value = item; sourceType.value = item.source_type; Object.assign(draft, { ...emptyDraft(), name: item.name, description: item.description, logo_url: item.logo_url, source_url: item.source_url, version: item.version || item.release?.tag_name || '', release_name: item.release_name || item.release?.name || '', release_notes: item.release_notes || item.release?.body || '', published_at: (item.published_at || item.release?.published_at || '').slice(0, 10), enabled: item.enabled, featured: item.featured, sort_order: item.sort_order, asset_variants: (item.asset_variants || []).map(asset => ({ ...asset })) }); dialogOpen.value = true }
function closeDialog() { dialogOpen.value = false; nextTick(() => dialogTrigger?.focus()) }
function addAsset() { draft.asset_variants.push({ name: '', label: '', platform: 'windows', architecture: 'x64', format: 'EXE', kind: 'installer', url: '', accelerated_url: '' }) }
async function identify() { identifying.value = true; identifyError.value = ''; preview.value = null; try { const result = await previewSoftware(repositoryUrl.value); preview.value = result; Object.assign(draft, { name: result.name, description: result.description, logo_url: result.logo_url }) } catch (error) { identifyError.value = errorMessage(error) } finally { identifying.value = false } }
async function save() { if (!sourceType.value) return; saving.value = true; try { const payload = { source_url: draft.source_url, name: draft.name, description: draft.description, logo_url: draft.logo_url, featured: draft.featured, enabled: draft.enabled, sort_order: draft.sort_order, version: draft.version, release_name: draft.release_name, release_notes: draft.release_notes, published_at: draft.published_at ? new Date(`${draft.published_at}T00:00:00Z`).toISOString() : null, asset_variants: draft.asset_variants }; if (editing.value) await updateSoftware(editing.value.id, payload); else await createSoftware({ source_type: sourceType.value, repository_url: sourceType.value === 'github' ? preview.value?.repository_url : undefined, ...payload }); appStore.showSuccess(editing.value ? '软件信息已保存' : '软件已发布'); closeDialog(); await load() } catch (error) { appStore.showError(errorMessage(error)) } finally { saving.value = false } }
async function toggleEnabled(item: SoftwareCenterItem, enabled: boolean) { busyId.value = item.id; try { Object.assign(item, await updateSoftware(item.id, { enabled })) } catch (error) { appStore.showError(errorMessage(error)) } finally { busyId.value = null } }
async function refresh(item: SoftwareCenterItem) { busyId.value = item.id; busyAction.value = 'refresh'; try { Object.assign(item, await refreshSoftware(item.id)); appStore.showSuccess('Release 已同步') } catch (error) { appStore.showError(errorMessage(error)) } finally { busyId.value = null; busyAction.value = '' } }
async function remove(item: SoftwareCenterItem) { if (!window.confirm(`确定移除“${item.name}”吗？`)) return; busyId.value = item.id; try { await deleteSoftware(item.id); items.value = items.value.filter(value => value.id !== item.id) } catch (error) { appStore.showError(errorMessage(error)) } finally { busyId.value = null } }
function onKeydown(event: KeyboardEvent) { if (event.key === 'Escape' && dialogOpen.value) closeDialog() }
watch(dialogOpen, async open => { document.body.style.overflow = open ? 'hidden' : ''; if (open) { await nextTick(); dialog.value?.querySelector<HTMLElement>('button, input')?.focus() } })
onMounted(() => { load(); window.addEventListener('keydown', onKeydown) }); onBeforeUnmount(() => { document.body.style.overflow = ''; window.removeEventListener('keydown', onKeydown) })
</script>

<style scoped>
.empty-state { border: 1px dashed var(--ui-border); border-radius: var(--radius-panel); padding: 2.5rem; text-align: center; color: var(--ui-text-muted); }
.icon-action { display: inline-flex; width: 40px; height: 40px; align-items: center; justify-content: center; border-radius: var(--radius-control); color: var(--ui-text-muted); transition: 150ms; }
.icon-action:hover { background: var(--ui-surface-subtle); color: var(--ui-text); }.icon-action.danger:hover { background: rgb(var(--color-danger-subtle)); color: var(--ui-danger); }.icon-action:disabled { opacity: .45; }
.dialog-backdrop { position: fixed; inset: 0; z-index: 50; display: flex; align-items: flex-end; justify-content: center; background: rgb(0 0 0 / .5); }
.dialog-panel { display: flex; max-height: 94vh; width: 100%; max-width: 760px; flex-direction: column; overflow: hidden; border: 1px solid var(--ui-border); border-radius: var(--radius-panel) var(--radius-panel) 0 0; background: var(--ui-surface); box-shadow: var(--shadow-floating); }
.dialog-header,.dialog-footer { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem 1.25rem; }.dialog-header { border-bottom: 1px solid var(--ui-border); }.dialog-footer { justify-content: flex-end; border-top: 1px solid var(--ui-border); }
.dialog-body { overflow-y: auto; padding: 1.25rem; }.source-picker { display: grid; gap: .75rem; padding: 1.25rem; }.source-picker>button { display: grid; grid-template-columns: 44px minmax(0,1fr) 20px; align-items: center; gap: .875rem; padding: 1rem; border: 1px solid var(--ui-border); border-radius: var(--radius-control); text-align: left; transition: 150ms; }.source-picker>button:hover { border-color: var(--ui-focus); background: var(--ui-surface-subtle); }.source-picker strong,.source-picker small { display: block; }.source-picker small { margin-top: .2rem; color: var(--ui-text-muted); }.source-icon { display: flex; width: 44px; height: 44px; align-items: center; justify-content: center; border-radius: var(--radius-control); background: rgb(var(--color-brand-subtle)); color: rgb(var(--color-brand)); }
.source-back { display: inline-flex; align-items: center; gap: .35rem; margin-bottom: 1rem; color: var(--ui-text-muted); font-size: .8rem; }.source-back:hover { color: rgb(var(--color-brand)); }.switch-field { display: flex; min-height: 44px; align-items: center; justify-content: space-between; border: 1px solid var(--ui-border); border-radius: var(--radius-control); padding: 0 .75rem; font-size: .875rem; }
.asset-editor { padding-top: .25rem; }.asset-row { display: grid; gap: .75rem; margin-top: .75rem; padding: 1rem; border: 1px solid var(--ui-border); border-radius: var(--radius-control); background: var(--ui-surface-subtle); }.asset-row-head { display: flex; align-items: center; justify-content: space-between; font-size: .8rem; }
@media (min-width: 640px) { .dialog-backdrop { align-items: center; padding: 1.25rem; }.dialog-panel { border-radius: var(--radius-panel); } }
</style>
