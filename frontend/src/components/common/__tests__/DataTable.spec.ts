import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import DataTable from '../DataTable.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

const stubMatchMedia = (matches: boolean) => {
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn()
    }))
  })
}

describe('DataTable', () => {
  beforeEach(() => {
    stubMatchMedia(true)
    localStorage.clear()
  })

  it('renders paired sort arrows and highlights the active direction', async () => {
    const wrapper = mount(DataTable, {
      props: {
        columns: [
          { key: 'name', label: 'Name', sortable: true },
          { key: 'created_at', label: 'Created', sortable: true }
        ],
        data: [
          { id: 1, name: 'Beta', created_at: '2026-01-02T00:00:00Z' },
          { id: 2, name: 'Alpha', created_at: '2026-01-01T00:00:00Z' }
        ],
        defaultSortKey: 'name',
        defaultSortOrder: 'asc'
      }
    })

    await wrapper.vm.$nextTick()

    const nameHeader = wrapper.findAll('th')[0]
    const sortButton = nameHeader.get('button')
    expect(nameHeader.attributes('aria-sort')).toBe('ascending')
    expect(sortButton.attributes('aria-label')).toBe('Sort Name descending')
    expect(nameHeader.findAll('svg')).toHaveLength(2)
    expect(nameHeader.findAll('svg')[0].classes()).toContain('text-foreground')
    expect(nameHeader.findAll('svg')[1].classes()).toContain('text-foreground-subtle/45')

    await sortButton.trigger('click')
    await wrapper.vm.$nextTick()

    expect(nameHeader.attributes('aria-sort')).toBe('descending')
    expect(sortButton.attributes('aria-label')).toBe('Sort Name ascending')
    expect(nameHeader.findAll('svg')[0].classes()).toContain('text-foreground-subtle/45')
    expect(nameHeader.findAll('svg')[1].classes()).toContain('text-foreground')
  })

  it('uses a separate native trigger for clickable mobile cards with interactive content', async () => {
    stubMatchMedia(false)
    const row = { id: 1, name: 'Alpha', status: 'Ready' }
    const wrapper = mount(DataTable, {
      attachTo: document.body,
      props: {
        columns: [
          { key: 'name', label: 'Name' },
          { key: 'status', label: 'Status' }
        ],
        data: [row],
        clickableRows: true
      },
      slots: {
        'cell-status': '<button type="button" data-test="status-action">Inspect</button>'
      }
    })

    await wrapper.vm.$nextTick()

    expect(window.matchMedia).toHaveBeenCalledWith('(min-width: 1024px)')
    expect(wrapper.find('table').exists()).toBe(false)
    const card = wrapper.get('.mobile-row-card')
    expect(card.classes()).toContain('p-3')
    expect(card.attributes('role')).toBeUndefined()
    expect(card.attributes('tabindex')).toBeUndefined()

    const rowTrigger = card.get('button.mobile-row-trigger')
    const nestedAction = card.get('[data-test="status-action"]')
    expect(rowTrigger.attributes('aria-label')).toContain('Name: Alpha')
    expect(rowTrigger.element.contains(nestedAction.element)).toBe(false)

    rowTrigger.element.focus()
    expect(document.activeElement).toBe(rowTrigger.element)
    await rowTrigger.trigger('click')
    expect(wrapper.emitted('rowClick')).toEqual([[row]])

    await nestedAction.trigger('click')
    expect(wrapper.emitted('rowClick')).toEqual([[row]])

    await card.trigger('click')
    expect(wrapper.emitted('rowClick')).toEqual([[row], [row]])
  })

  it('makes desktop clickable rows keyboard activatable without changing row events', async () => {
    const row = { id: 7, name: 'Desktop row' }
    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data: [row],
        clickableRows: true
      }
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()

    const tableRow = wrapper.get('tbody tr[data-row-id="7"]')
    expect(tableRow.attributes('tabindex')).toBe('0')
    expect(tableRow.attributes('aria-label')).toContain('Desktop row')
    await tableRow.trigger('keydown', { key: 'Enter' })
    expect(wrapper.emitted('rowClick')).toEqual([[row]])
  })

  it('renders every row with no virtual padding spacer for small datasets (virtualization off)', async () => {
    const data = Array.from({ length: 8 }, (_, i) => ({ id: i + 1, name: `Row ${i + 1}` }))
    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data
      }
    })

    await wrapper.vm.$nextTick()

    // Virtualization is OFF for a small list…
    expect((wrapper.vm as any).shouldVirtualize).toBe(false)
    // …every row is in the DOM…
    expect(wrapper.findAll('tbody tr[data-index]')).toHaveLength(data.length)
    // …and there are no aria-hidden virtual padding spacer rows.
    expect(wrapper.findAll('tbody tr[aria-hidden="true"]')).toHaveLength(0)
  })

  it('switches to windowed rendering once row count exceeds virtualizeThreshold', async () => {
    const data = Array.from({ length: 12 }, (_, i) => ({ id: i + 1, name: `Row ${i + 1}` }))
    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data,
        virtualizeThreshold: 3
      }
    })

    await wrapper.vm.$nextTick()

    // Virtualization is ON: the mode-switch decision flipped…
    expect((wrapper.vm as any).shouldVirtualize).toBe(true)
    // …and the virtualizer drives off the full row count.
    const exposed = (wrapper.vm as any).virtualizer
    const instance = exposed?.value ?? exposed
    expect(instance.options.count).toBe(data.length)
  })

  it('keys the virtualizer size cache by row identity, not index (avoids stale heights on sort/filter)', async () => {
    const data = Array.from({ length: 12 }, (_, i) => ({ id: 100 + i, name: `Row ${i + 1}` }))
    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data,
        rowKey: 'id',
        virtualizeThreshold: 3
      }
    })

    await wrapper.vm.$nextTick()

    const exposed = (wrapper.vm as any).virtualizer
    const instance = exposed?.value ?? exposed
    // getItemKey must resolve to the row's stable key (id), not the positional index.
    expect(instance.options.getItemKey(0)).toBe(100)
    expect(instance.options.getItemKey(5)).toBe(105)
  })
})
