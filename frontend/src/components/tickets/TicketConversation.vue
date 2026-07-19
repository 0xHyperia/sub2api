<template>
  <div class="space-y-4">
    <section v-if="ticket" class="rounded-panel border border-outline bg-surface p-4 shadow-card">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0">
          <button type="button" class="btn btn-secondary btn-sm mb-3" @click="$emit('back')"><Icon name="arrowLeft" size="sm" />{{ localText('返回', 'Back') }}</button>
          <div class="text-xs font-semibold text-info-foreground">{{ ticket.number }}</div>
          <h2 class="mt-1 break-words text-lg font-semibold text-foreground">{{ ticket.subject }}</h2>
          <p class="mt-1 text-xs text-foreground-muted">{{ categoryName(ticket.category) }} · {{ formatDate(ticket.created_at) }}<span v-if="admin"> · {{ ticket.user_email }}</span></p>
        </div>
        <span class="status-badge" :class="statusClass(ticket.status)">{{ statusLabel(ticket.status) }}</span>
      </div>
    </section>

    <section class="min-h-[360px] rounded-panel border border-outline bg-surface p-4 shadow-card">
      <div v-if="loading" class="flex min-h-[320px] items-center justify-center"><Icon name="refresh" size="lg" class="animate-spin" /></div>
      <div v-else-if="ticket" class="space-y-4">
        <template v-for="message in ticket.messages" :key="message.id">
          <div v-if="message.sender_type === 'system'" class="flex justify-center"><span class="rounded-full bg-surface-subtle px-3 py-1 text-xs text-foreground-muted">{{ eventLabel(message.event_type) }}</span></div>
          <article v-else class="flex" :class="isMine(message.sender_type) ? 'justify-end' : 'justify-start'">
            <div class="max-w-[88%] rounded-panel border p-3 sm:max-w-[72%]" :class="isMine(message.sender_type) ? 'border-info bg-info-subtle' : 'border-outline bg-canvas'">
              <div class="mb-2 flex flex-col items-start gap-0.5 text-xs text-foreground-muted sm:flex-row sm:items-center sm:justify-between sm:gap-4"><span class="font-semibold">{{ senderLabel(message.sender_type) }}</span><time>{{ formatDate(message.created_at) }}</time></div>
              <p v-if="message.content" class="whitespace-pre-wrap break-words text-sm leading-6 text-foreground">{{ message.content }}</p>
              <div v-if="message.attachments.length" class="mt-3 flex flex-wrap gap-2">
                <button v-for="attachment in message.attachments" :key="attachment.id" type="button" class="btn btn-secondary btn-sm max-w-full" @click="download(attachment.id)"><Icon name="download" size="sm" /><span class="truncate">{{ attachment.original_name }}</span><span class="text-xs text-foreground-subtle">{{ formatBytes(attachment.size_bytes) }}</span></button>
              </div>
            </div>
          </article>
        </template>
      </div>
    </section>

    <section v-if="ticket" class="rounded-panel border border-outline bg-surface p-4 shadow-card">
      <textarea v-model="content" class="input min-h-28 resize-y" :disabled="ticket.status === 'closed' || sending" maxlength="10000" :placeholder="ticket.status === 'closed' ? localText('重新打开工单后可继续回复', 'Reopen the ticket to continue') : localText('输入回复内容', 'Write a reply')"></textarea>
      <div v-if="files.length" class="mt-2 flex flex-wrap gap-2"><span v-for="(file,index) in files" :key="`${file.name}-${index}`" class="inline-flex max-w-full items-center gap-1 rounded-control bg-surface-subtle px-2 py-1 text-xs"><span class="truncate">{{ file.name }}</span><button type="button" :aria-label="localText('移除附件','Remove attachment')" @click="files.splice(index,1)"><Icon name="x" size="xs" /></button></span></div>
      <div v-if="sending && files.length" class="mt-2 space-y-1"><div class="flex justify-between text-xs text-foreground-muted"><span>{{ localText('正在上传','Uploading') }}</span><span>{{ uploadProgress }}%</span></div><progress class="h-2 w-full accent-info" max="100" :value="uploadProgress"></progress></div>
      <div class="mt-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <label class="btn btn-secondary w-full cursor-pointer sm:w-auto" :class="{ 'pointer-events-none opacity-50': !capabilities.attachments_available || ticket.status === 'closed' }"><Icon name="upload" size="sm" />{{ localText('添加附件','Add attachment') }}<input class="sr-only" type="file" multiple :accept="accept" :disabled="!capabilities.attachments_available || ticket.status === 'closed'" @change="selectFiles"></label>
        <div class="grid w-full gap-2 sm:flex sm:w-auto sm:justify-end"><button v-if="ticket.status === 'closed'" type="button" class="btn btn-secondary w-full sm:w-auto" :disabled="sending" @click="changeStatus(true)">{{ localText('重新打开','Reopen') }}</button><button v-else type="button" class="btn btn-secondary w-full sm:w-auto" :disabled="sending" @click="changeStatus(false)">{{ localText('关闭工单','Close ticket') }}</button><button type="button" class="btn btn-primary w-full sm:w-auto" :disabled="sending || ticket.status === 'closed' || (!content.trim() && files.length === 0)" @click="send"><Icon name="arrowRight" size="sm" />{{ localText('发送','Send') }}</button></div>
      </div>
      <p v-if="!capabilities.attachments_available" class="mt-2 text-xs text-foreground-subtle">{{ localText('附件存储未配置，当前仅支持文本消息。','Attachment storage is unavailable; text messages remain available.') }}</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTicketNotificationStore } from '@/stores/ticketNotifications'
import { ticketsAPI } from '@/api/tickets'
import { adminTicketsAPI } from '@/api/admin/tickets'
import type { Ticket, TicketCategory, TicketAttachmentCapabilities, TicketStatus } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const props=defineProps<{number:string;admin?:boolean}>();defineEmits<{back:[]}>()
const {locale}=useI18n();const appStore=useAppStore();const notifications=useTicketNotificationStore()
const ticket=ref<Ticket|null>(null);const loading=ref(true);const sending=ref(false);const content=ref('');const files=ref<File[]>([]);const uploadProgress=ref(0)
const capabilities=ref<TicketAttachmentCapabilities>({attachments_available:false,max_file_bytes:10*1024*1024,max_files_per_message:5,max_total_bytes:25*1024*1024,allowed_extensions:[]})
let timer:ReturnType<typeof setInterval>|null=null
const localText=(zh:string,en:string)=>locale.value.startsWith('zh')?zh:en
const accept=computed(()=>capabilities.value.allowed_extensions.join(','))
async function load(silent=false){if(!silent)loading.value=true;try{ticket.value=props.admin?await adminTicketsAPI.get(props.number):await ticketsAPI.get(props.number);await (props.admin?adminTicketsAPI.markRead(props.number):ticketsAPI.markRead(props.number));void notifications.refresh()}catch(error){appStore.showError(extractApiErrorMessage(error,localText('加载工单失败','Failed to load ticket')))}finally{loading.value=false}}
async function send(){sending.value=true;uploadProgress.value=0;const onProgress=(value:number)=>{uploadProgress.value=value};try{ticket.value=props.admin?await adminTicketsAPI.reply(props.number,content.value,files.value,onProgress):await ticketsAPI.reply(props.number,content.value,files.value,onProgress);content.value='';files.value=[];appStore.showSuccess(localText('回复已发送','Reply sent'));void notifications.refresh()}catch(error){appStore.showError(extractApiErrorMessage(error,localText('发送失败','Send failed')))}finally{sending.value=false}}
async function changeStatus(reopen:boolean){sending.value=true;try{ticket.value=props.admin?(reopen?await adminTicketsAPI.reopen(props.number):await adminTicketsAPI.close(props.number)):(reopen?await ticketsAPI.reopen(props.number):await ticketsAPI.close(props.number));appStore.showSuccess(reopen?localText('工单已重新打开','Ticket reopened'):localText('工单已关闭','Ticket closed'));void notifications.refresh()}catch(error){appStore.showError(extractApiErrorMessage(error,localText('操作失败','Operation failed')))}finally{sending.value=false}}
function selectFiles(event:Event){const input=event.target as HTMLInputElement;const selected=Array.from(input.files||[]);if(files.value.length+selected.length>capabilities.value.max_files_per_message){appStore.showError(localText('附件数量超过限制','Too many attachments'));input.value='';return}const total=[...files.value,...selected].reduce((sum,file)=>sum+file.size,0);if(selected.some(file=>file.size>capabilities.value.max_file_bytes)||total>capabilities.value.max_total_bytes){appStore.showError(localText('附件大小超过限制','Attachment size limit exceeded'));input.value='';return}files.value.push(...selected);input.value=''}
async function download(id:number){try{const url=await ticketsAPI.attachmentURL(id);window.open(url,'_blank','noopener,noreferrer')}catch(error){appStore.showError(extractApiErrorMessage(error,localText('下载失败','Download failed')))}}
const isMine=(sender:string)=>props.admin?sender==='admin':sender==='user';const senderLabel=(sender:string)=>sender==='admin'?localText('后台','Support'):localText('我','Me')
const eventLabel=(event:string)=>event==='ticket_reopened'?localText('工单已重新打开','Ticket reopened'):localText('工单已关闭','Ticket closed')
const statusLabel=(status:TicketStatus)=>({open:localText('待处理','Open'),answered:localText('已回复','Answered'),closed:localText('已关闭','Closed')}[status])
const statusClass=(status:TicketStatus)=>status==='open'?'status-badge-warning':status==='answered'?'status-badge-success':'status-badge-neutral'
const categoryName=(category:TicketCategory)=>locale.value.startsWith('zh')?category.name_zh:(category.name_en||category.name_zh)
const formatDate=(value:string)=>new Intl.DateTimeFormat(locale.value,{month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit'}).format(new Date(value))
const formatBytes=(value:number)=>value<1024?`${value} B`:value<1024*1024?`${(value/1024).toFixed(1)} KB`:`${(value/1024/1024).toFixed(1)} MB`
onMounted(async()=>{capabilities.value=await ticketsAPI.capabilities().catch(()=>capabilities.value);await load();timer=setInterval(()=>{if(document.visibilityState==='visible')void load(true)},30000)})
onBeforeUnmount(()=>{if(timer)clearInterval(timer)})
</script>
