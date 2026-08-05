import { describe, expect, it, vi } from 'vitest'

const get = vi.hoisted(() => vi.fn())

vi.mock('@/api/client', () => ({
  apiClient: { get },
}))

import { list } from '@/api/admin/modelMonitor'

describe('model monitor API normalization', () => {
  it('normalizes null groups from older backend responses', async () => {
    get.mockResolvedValueOnce({
      data: {
        items: [
          { platform: 'openai', model: 'gpt-test', groups: null },
          { platform: 'anthropic', model: 'claude-test', groups: [{ group_id: 1 }] },
        ],
      },
    })

    const response = await list()

    expect(response.items[0].groups).toEqual([])
    expect(response.items[1].groups).toHaveLength(1)
  })

  it('falls back to an empty item list for an invalid response', async () => {
    get.mockResolvedValueOnce({ data: null })

    const response = await list()

    expect(response.items).toEqual([])
  })
})
