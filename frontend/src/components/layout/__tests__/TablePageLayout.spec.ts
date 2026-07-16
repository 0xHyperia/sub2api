import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import TablePageLayout from '../TablePageLayout.vue'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../TablePageLayout.vue')

describe('TablePageLayout', () => {
  it('uses the same 1024px responsive boundary as DataTable', async () => {
    let changeListener: ((event: MediaQueryListEvent) => void) | null = null
    const mediaQuery = {
      matches: false,
      media: '(min-width: 1024px)',
      onchange: null,
      addEventListener: vi.fn((_type: string, listener: (event: MediaQueryListEvent) => void) => {
        changeListener = listener
      }),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn()
    }
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      writable: true,
      value: vi.fn(() => mediaQuery)
    })

    const wrapper = mount(TablePageLayout, {
      slots: {
        table: '<div>Table content</div>'
      }
    })

    expect(window.matchMedia).toHaveBeenCalledWith('(min-width: 1024px)')
    expect(wrapper.get('.table-page-layout').classes()).toContain('mobile-mode')

    changeListener?.({ matches: true } as MediaQueryListEvent)
    await wrapper.vm.$nextTick()
    expect(wrapper.get('.table-page-layout').classes()).not.toContain('mobile-mode')

    wrapper.unmount()
    expect(mediaQuery.removeEventListener).toHaveBeenCalled()
  })

  it('does not disable the table horizontal scroll container in mobile mode', () => {
    const componentSource = readFileSync(componentPath, 'utf8')
    const tableWrapperBlocks = Array.from(
      componentSource.matchAll(/([^{}]*:deep\(\.table-wrapper\)[^{}]*)\{([^{}]*)\}/g)
    )

    expect(tableWrapperBlocks.length).toBeGreaterThan(0)
    const baseBlock = tableWrapperBlocks.find(([selector]) => !selector.includes('.mobile-mode'))
    const mobileBlocks = tableWrapperBlocks.filter(([selector]) => selector.includes('.mobile-mode'))

    expect(baseBlock?.[2]).toContain('overflow-x-auto')
    expect(mobileBlocks.every(([, , declarations]) => !declarations.includes('overflow-visible'))).toBe(true)
  })
})
