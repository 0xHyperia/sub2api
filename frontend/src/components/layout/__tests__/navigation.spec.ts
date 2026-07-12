import { describe, expect, it, vi } from 'vitest'

import {
  filterNavigationItems,
  resolveGroupClickAction,
  type ShellNavItem
} from '../navigation'

describe('layout navigation helpers', () => {
  it('filters feature-gated and simple-mode items recursively', () => {
    const enabled = vi.fn(() => true)
    const disabled = vi.fn(() => false)
    const items: ShellNavItem[] = [
      { path: '/always', label: 'Always', icon: 'home' },
      { path: '/advanced', label: 'Advanced', icon: 'cog', hideInSimpleMode: true },
      { path: '/disabled', label: 'Disabled', icon: 'shield', featureFlag: disabled },
      {
        path: '/group',
        label: 'Group',
        icon: 'grid',
        featureFlag: enabled,
        children: [
          { path: '/group/visible', label: 'Visible', icon: 'document' },
          { path: '/group/hidden', label: 'Hidden', icon: 'document', featureFlag: disabled }
        ]
      }
    ]

    expect(filterNavigationItems(items, true)).toEqual([
      { path: '/always', label: 'Always', icon: 'home' },
      {
        path: '/group',
        label: 'Group',
        icon: 'grid',
        featureFlag: enabled,
        children: [{ path: '/group/visible', label: 'Visible', icon: 'document' }]
      }
    ])
  })

  it('removes groups with no visible children', () => {
    const items: ShellNavItem[] = [
      {
        path: '/empty',
        label: 'Empty',
        icon: 'grid',
        children: [
          { path: '/empty/hidden', label: 'Hidden', icon: 'document', featureFlag: () => false }
        ]
      }
    ]

    expect(filterNavigationItems(items, false)).toEqual([])
  })

  it('resolves collapsed groups to an explicit sidebar expansion', () => {
    expect(resolveGroupClickAction(true, true)).toBe('expand-sidebar')
    expect(resolveGroupClickAction(false, true)).toBe('toggle-group')
    expect(resolveGroupClickAction(false, false)).toBe('navigate-and-expand')
  })
})
