import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, post, put } = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn(), put: vi.fn() }))
vi.mock('../client', () => ({ apiClient: { get, post, put } }))

import { ticketsAPI } from '@/api/tickets'
import { adminTicketsAPI } from '@/api/admin/tickets'

describe('ticket APIs', () => {
  beforeEach(() => { get.mockReset(); post.mockReset(); put.mockReset() })

  it('creates a user ticket as multipart form data', async () => {
    post.mockResolvedValue({ data: { number: '20260716-ABC234' } })
    const file = new File(['failure'], 'error.log', { type: 'text/plain' })

    const onProgress = vi.fn()
    await ticketsAPI.create(3, 'request failed', [file], onProgress)

    const [path, form, config] = post.mock.calls[0]
    expect(path).toBe('/tickets')
    expect(form).toBeInstanceOf(FormData)
    expect(form.get('category_id')).toBe('3')
    expect(form.get('description')).toBe('request failed')
    expect(form.getAll('files')).toHaveLength(1)
    config.onUploadProgress({ loaded: 5, total: 10 })
    expect(onProgress).toHaveBeenCalledWith(50)
  })

  it('uses separate user and admin unread endpoints', async () => {
    get.mockResolvedValueOnce({ data: { count: 2 } }).mockResolvedValueOnce({ data: { count: 5 } })
    await expect(ticketsAPI.unread()).resolves.toBe(2)
    await expect(adminTicketsAPI.unread()).resolves.toBe(5)
    expect(get).toHaveBeenNthCalledWith(1, '/tickets/unread')
    expect(get).toHaveBeenNthCalledWith(2, '/admin/tickets/unread')
  })

  it('updates and reorders admin categories without deleting them', async () => {
    put.mockResolvedValue({ data: { id: 4 } })
    await adminTicketsAPI.updateCategory(4, { active: false })
    await adminTicketsAPI.reorderCategories([4, 2, 1])
    expect(put).toHaveBeenNthCalledWith(1, '/admin/ticket-categories/4', { active: false })
    expect(put).toHaveBeenNthCalledWith(2, '/admin/ticket-categories/reorder', { ids: [4, 2, 1] })
  })
})
