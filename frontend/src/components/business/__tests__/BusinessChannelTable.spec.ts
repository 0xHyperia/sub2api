import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import BusinessChannelTable from '../BusinessChannelTable.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key.split('.').at(-1) || key }) }))

const row = (channel: 'distribution' | 'affiliate', net: number) => ({
  channel, new_users: 1, activated_users: 1, first_paid_users: 1, paying_users: 1,
  repurchase_users: 0, active_users: 1, net_paid_cny: net, arppu_cny: net, d7_retention_rate: 0,
})

describe('BusinessChannelTable', () => {
  it('sorts by net paid by default and toggles another sortable column', async () => {
    const wrapper = mount(BusinessChannelTable, { props: { rows: [row('distribution', 10), row('affiliate', 30)] } })
    expect(wrapper.find('tbody tr .channel-link').text()).toBe('channel_affiliate')
    await wrapper.findAll('thead button')[1].trigger('click')
    await wrapper.findAll('thead button')[1].trigger('click')
    expect(wrapper.find('tbody tr .channel-link').text()).toBe('channel_distribution')
  })
})
