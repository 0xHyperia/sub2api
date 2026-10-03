import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import MonitorDetailDialog from '../MonitorDetailDialog.vue'

const { fetchDetail, showError } = vi.hoisted(() => ({ fetchDetail: vi.fn(), showError: vi.fn() }))

vi.mock('@/api/channelMonitor', () => ({
  status: fetchDetail,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const BaseDialogStub = {
  props: ['show', 'title'],
  template: '<section v-if="show"><h2>{{ title }}</h2><slot /><footer><slot name="footer" /></footer></section>',
}

enableAutoUnmount(afterEach)

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

function deferred() {
  let resolve!: (value: unknown) => void
  let reject!: (value: unknown) => void
  const promise = new Promise((res, rej) => { resolve = res; reject = rej })
  return { promise, resolve, reject }
}
const monitorDetail = (model: string) => ({ models: [{ model, latest_status: 'operational' }] })
function openMonitor() {
  return mount(MonitorDetailDialog, { props: { show: true, monitorId: 1, title: 'Monitor' },
    global: { stubs: { BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' } } } })
}
describe('monitor detail request ownership', () => {
  beforeEach(() => { fetchDetail.mockReset(); showError.mockReset() })

  it('keeps the new monitor response when the old response arrives last', async () => {
    const old = deferred()
    fetchDetail.mockReturnValueOnce(old.promise).mockResolvedValueOnce(monitorDetail('new-model'))
    const w = openMonitor(); await w.setProps({ monitorId: 2 }); await flushPromises()
    old.resolve(monitorDetail('old-model')); await flushPromises()
    expect(w.text()).toContain('new-model'); expect(w.text()).not.toContain('old-model')
  })
  it('ignores a previous failure while the current monitor is loading', async () => {
    const old = deferred(); const current = deferred()
    fetchDetail.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const w = openMonitor(); await w.setProps({ show: false }); await w.setProps({ show: true })
    old.reject(new Error('old failure')); await flushPromises()
    expect(showError).not.toHaveBeenCalled(); expect(w.text()).toContain('common.loading')
    current.resolve(monitorDetail('current')); await flushPromises(); expect(w.text()).toContain('current')
  })
  it('reports a failure for the current monitor', async () => {
    fetchDetail.mockRejectedValueOnce(new Error('current failure'))
    const w = openMonitor(); await flushPromises()
    expect(showError).toHaveBeenCalledWith('current failure')
    expect(w.text()).toContain('channelStatus.detailLoadError')
  })
})
