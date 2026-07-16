import { apiClient } from './client'
import type { BasePaginationResponse, Ticket, TicketAttachmentCapabilities, TicketCategory } from '@/types'

export interface TicketListParams { page?: number; page_size?: number; status?: string; category_id?: number; search?: string; unread_only?: boolean }
export type TicketUploadProgress = (percent: number) => void
export async function list(params: TicketListParams = {}): Promise<BasePaginationResponse<Ticket>> { const { data } = await apiClient.get('/tickets', { params }); return data }
export async function categories(): Promise<TicketCategory[]> { const { data } = await apiClient.get('/ticket-categories'); return data }
export async function capabilities(): Promise<TicketAttachmentCapabilities> { const { data } = await apiClient.get('/tickets/capabilities'); return data }
export async function get(number: string): Promise<Ticket> { const { data } = await apiClient.get(`/tickets/${number}`); return data }
export async function create(categoryId: number, description: string, files: File[], onProgress?: TicketUploadProgress): Promise<Ticket> { const form = ticketForm({ category_id: String(categoryId), description }, files); const { data } = await apiClient.post('/tickets', form, uploadConfig(onProgress)); return data }
export async function reply(number: string, content: string, files: File[], onProgress?: TicketUploadProgress): Promise<Ticket> { const form = ticketForm({ content }, files); const { data } = await apiClient.post(`/tickets/${number}/messages`, form, uploadConfig(onProgress)); return data }
export async function markRead(number: string): Promise<void> { await apiClient.post(`/tickets/${number}/read`) }
export async function close(number: string): Promise<Ticket> { const { data } = await apiClient.post(`/tickets/${number}/close`); return data }
export async function reopen(number: string): Promise<Ticket> { const { data } = await apiClient.post(`/tickets/${number}/reopen`); return data }
export async function unread(): Promise<number> { const { data } = await apiClient.get<{ count: number }>('/tickets/unread'); return data.count }
export async function attachmentURL(id: number): Promise<string> { const { data } = await apiClient.get<{ url: string }>(`/ticket-attachments/${id}`); return data.url }
function ticketForm(fields: Record<string,string>, files: File[]) { const form = new FormData(); Object.entries(fields).forEach(([key,value]) => form.append(key,value)); files.forEach(file => form.append('files',file)); return form }
function uploadConfig(onProgress?: TicketUploadProgress) { return { headers: { 'Content-Type': undefined }, onUploadProgress: onProgress ? (event: { loaded: number; total?: number }) => { if (event.total) onProgress(Math.min(100, Math.round(event.loaded * 100 / event.total))) } : undefined } }
export const ticketsAPI = { list, categories, capabilities, get, create, reply, markRead, close, reopen, unread, attachmentURL }
