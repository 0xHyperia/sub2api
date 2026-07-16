import { apiClient } from '../client'
import type { BasePaginationResponse, Ticket, TicketCategory } from '@/types'
import type { TicketListParams, TicketUploadProgress } from '../tickets'

export async function list(params: TicketListParams = {}): Promise<BasePaginationResponse<Ticket>> { const { data } = await apiClient.get('/admin/tickets', { params }); return data }
export async function get(number: string): Promise<Ticket> { const { data } = await apiClient.get(`/admin/tickets/${number}`); return data }
export async function reply(number: string, content: string, files: File[], onProgress?: TicketUploadProgress): Promise<Ticket> { const form = new FormData(); form.append('content',content); files.forEach(file=>form.append('files',file)); const { data } = await apiClient.post(`/admin/tickets/${number}/messages`,form,{headers:{'Content-Type':undefined},onUploadProgress:onProgress?(event:{loaded:number;total?:number})=>{if(event.total)onProgress(Math.min(100,Math.round(event.loaded*100/event.total)))}:undefined}); return data }
export async function markRead(number:string):Promise<void>{await apiClient.post(`/admin/tickets/${number}/read`)}
export async function close(number:string):Promise<Ticket>{const{data}=await apiClient.post(`/admin/tickets/${number}/close`);return data}
export async function reopen(number:string):Promise<Ticket>{const{data}=await apiClient.post(`/admin/tickets/${number}/reopen`);return data}
export async function unread():Promise<number>{const{data}=await apiClient.get<{count:number}>('/admin/tickets/unread');return data.count}
export async function categories():Promise<TicketCategory[]>{const{data}=await apiClient.get('/admin/ticket-categories');return data}
export async function createCategory(payload:Partial<TicketCategory>):Promise<TicketCategory>{const{data}=await apiClient.post('/admin/ticket-categories',payload);return data}
export async function updateCategory(id:number,payload:Partial<TicketCategory>):Promise<TicketCategory>{const{data}=await apiClient.put(`/admin/ticket-categories/${id}`,payload);return data}
export async function reorderCategories(ids:number[]):Promise<void>{await apiClient.put('/admin/ticket-categories/reorder',{ids})}
export const adminTicketsAPI={list,get,reply,markRead,close,reopen,unread,categories,createCategory,updateCategory,reorderCategories}
