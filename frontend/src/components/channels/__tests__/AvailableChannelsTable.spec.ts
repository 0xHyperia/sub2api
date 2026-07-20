import { mount } from '@vue/test-utils'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it, vi } from 'vitest'

import AvailableChannelsTable from '../AvailableChannelsTable.vue'
import type { UserAvailableChannel } from '@/api/channels'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: null }),
}))

const rows: UserAvailableChannel[] = [
  {
    name: 'Primary route',
    description: 'Low latency production channel',
    platforms: [
      {
        platform: 'openai',
        groups: [
          {
            id: 7,
            name: 'Plus',
            platform: 'openai',
            subscription_type: 'standard',
            rate_multiplier: 1.2,
            peak_rate_enabled: false,
            peak_start: '',
            peak_end: '',
            peak_rate_multiplier: 1,
            is_exclusive: true,
          },
        ],
        supported_models: [
          {
            name: 'gpt-5.6',
            platform: 'openai',
            pricing: null,
          },
        ],
      },
    ],
  },
]

const mountTable = () => mount(AvailableChannelsTable, {
  props: {
    columns: {
      name: 'Channel',
      description: 'Description',
      platform: 'Platform',
      groups: 'Groups',
      supportedModels: 'Models',
    },
    rows,
    loading: false,
    userGroupRates: { 7: 1.1 },
    pricingKeyPrefix: 'pricing',
    noPricingLabel: 'No pricing',
    noModelsLabel: 'No models',
    emptyLabel: 'No channels',
  },
  global: {
    stubs: {
      Icon: { props: ['name'], template: '<i :data-icon="name" />' },
      PlatformIcon: { props: ['platform'], template: '<i :data-platform="platform" />' },
      GroupBadge: { props: ['name'], template: '<span data-test="group-badge">{{ name }}</span>' },
      SupportedModelChip: { props: ['model'], template: '<span data-test="model-chip">{{ model.name }}</span>' },
    },
  },
})

describe('AvailableChannelsTable responsive layouts', () => {
  it('renders a business-specific mobile channel card with collapsible platform details', () => {
    const wrapper = mountTable()
    const mobileRegion = wrapper.get('.lg\\:hidden')

    expect(mobileRegion.text()).toContain('Primary route')
    expect(mobileRegion.text()).toContain('Low latency production channel')
    expect(mobileRegion.get('summary').text()).toContain('Groups 1')
    expect(mobileRegion.get('summary').text()).toContain('Models 1')
    expect(mobileRegion.get('[data-test="group-badge"]').text()).toBe('Plus')
    expect(mobileRegion.get('[data-test="model-chip"]').text()).toBe('gpt-5.6')
  })

  it('keeps the dense desktop table outside the mobile experience', () => {
    const wrapper = mountTable()
    const desktopRegion = wrapper.get('.table-wrapper')

    expect(desktopRegion.classes()).toContain('hidden')
    expect(desktopRegion.classes()).toContain('lg:block')
    expect(desktopRegion.get('table').classes()).toContain('min-w-[960px]')
    expect(desktopRegion.text()).toContain('Primary route')
  })

  it('keeps the desktop table on the scroll hook without clipping content', () => {
    const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AvailableChannelsTable.vue')
    const componentSource = readFileSync(componentPath, 'utf8')

    expect(componentSource).toMatch(/class="table-wrapper[^\"]*overflow-auto[^\"]*"/)
    expect(componentSource).not.toMatch(/<div class="card overflow-hidden">/)
  })
})
