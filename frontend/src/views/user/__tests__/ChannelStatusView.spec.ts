import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ChannelStatusView from '@/views/user/ChannelStatusView.vue'

const mocks = vi.hoisted(() => ({
  list: vi.fn(),
  status: vi.fn(),
  showError: vi.fn(),
  setEnabled: vi.fn(),
  start: vi.fn(),
  stop: vi.fn(),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { channel_monitor_enabled: true },
    showError: mocks.showError,
  }),
}))

vi.mock('@/api/channelMonitor', () => ({
  list: mocks.list,
  status: mocks.status,
}))

vi.mock('@/composables/useAutoRefresh', async () => {
  const { ref } = await import('vue')
  return {
    useAutoRefresh: () => ({
      enabled: ref(false),
      intervalSeconds: ref(30),
      countdown: ref(30),
      intervals: [30, 60, 120] as const,
      setEnabled: mocks.setEnabled,
      setInterval: vi.fn(),
      start: mocks.start,
      stop: mocks.stop,
    }),
  }
})

describe('ChannelStatusView load state', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    mocks.list.mockReset()
    mocks.status.mockReset()
    mocks.showError.mockReset()
    mocks.setEnabled.mockReset()
    mocks.start.mockReset()
    mocks.stop.mockReset()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined)
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  it('does not report operational status on failure and recovers through retry', async () => {
    mocks.list
      .mockRejectedValueOnce(new Error('monitor unavailable'))
      .mockResolvedValueOnce({
        items: [{ id: 1, name: 'Primary', primary_status: 'operational' }],
      })

    const wrapper = mount(ChannelStatusView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          MonitorHero: {
            props: ['overallStatus'],
            template: '<div data-testid="monitor-hero">{{ overallStatus }}</div>',
          },
          MonitorCardGrid: { template: '<div data-testid="monitor-grid-stub" />' },
          MonitorDetailDialog: true,
        },
      },
    })

    await flushPromises()

    const errorState = wrapper.get('[data-testid="channel-status-load-error"]')
    expect(errorState.text()).toContain('channelStatus.loadError')
    expect(wrapper.find('[data-testid="monitor-hero"]').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('operational')

    await errorState.get('button').trigger('click')
    await flushPromises()

    expect(mocks.list).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="channel-status-load-error"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="monitor-hero"]').text()).toBe('operational')
  })
})
