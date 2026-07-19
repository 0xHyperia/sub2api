import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import MonitorDetailDialog from '../MonitorDetailDialog.vue'

const { fetchDetail } = vi.hoisted(() => ({ fetchDetail: vi.fn() }))

vi.mock('@/api/channelMonitor', () => ({
  status: fetchDetail,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn() }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<section v-if="show"><h2>{{ title }}</h2><slot /><footer><slot name="footer" /></footer></section>',
}

describe('MonitorDetailDialog responsive presentation', () => {
  beforeEach(() => {
    fetchDetail.mockReset()
    fetchDetail.mockResolvedValue({
      id: 3,
      name: 'Production',
      provider: 'openai',
      group_name: 'Plus',
      models: [
        {
          model: 'gpt-5.6',
          latest_status: 'operational',
          latest_latency_ms: 842,
          availability_7d: 99.95,
          availability_15d: 99.8,
          availability_30d: 99.5,
          avg_latency_7d_ms: 910,
        },
      ],
    })
  })

  it('renders mobile status cards and preserves the desktop comparison table', async () => {
    const wrapper = mount(MonitorDetailDialog, {
      props: { show: true, monitorId: 3, title: 'Production status' },
      global: { stubs: { BaseDialog: BaseDialogStub } },
    })

    await flushPromises()

    const mobileCard = wrapper.get('article')
    expect(mobileCard.text()).toContain('gpt-5.6')
    expect(mobileCard.text()).toContain('842')
    expect(mobileCard.text()).toContain('99.95%')
    expect(mobileCard.text()).toContain('99.50%')

    const desktopTable = wrapper.get('table')
    expect(desktopTable.element.closest('.sm\\:block')).not.toBeNull()
    expect(desktopTable.text()).toContain('gpt-5.6')
    expect(desktopTable.text()).toContain('910')
  })
})
