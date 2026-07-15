import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import GroupCapacityBadge from '../GroupCapacityBadge.vue'

describe('GroupCapacityBadge', () => {
  it('renders idle capacity with a quiet neutral surface', () => {
    const wrapper = mount(GroupCapacityBadge, {
      props: {
        concurrencyUsed: 0,
        concurrencyMax: 10,
        sessionsUsed: 0,
        sessionsMax: 4,
        rpmUsed: 0,
        rpmMax: 100
      }
    })

    const badges = wrapper.findAll('span.inline-flex')
    expect(badges).toHaveLength(3)
    for (const badge of badges) {
      expect(badge.classes()).toContain('border-outline')
      expect(badge.classes()).toContain('bg-surface-subtle')
      expect(badge.classes()).not.toContain('bg-foreground')
    }
  })
})
