import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import RedeemCodeForm from '@/components/payment/RedeemCodeForm.vue'

const redeem = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const fetchActiveSubscriptions = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

vi.mock('@/api/redeem', () => ({
  redeemAPI: { redeem },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ refreshUser }),
}))

vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({ fetchActiveSubscriptions }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess, showWarning, showError }),
}))

describe('RedeemCodeForm', () => {
  beforeEach(() => {
    redeem.mockReset()
    refreshUser.mockReset().mockResolvedValue(undefined)
    fetchActiveSubscriptions.mockReset().mockResolvedValue(undefined)
    showSuccess.mockReset()
    showWarning.mockReset()
    showError.mockReset()
  })

  it('redeems a balance code and refreshes the current user', async () => {
    redeem.mockResolvedValue({ message: 'ok', type: 'balance', value: 12 })
    const wrapper = mount(RedeemCodeForm)

    await wrapper.get('input').setValue(' BALANCE-12 ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(redeem).toHaveBeenCalledWith('BALANCE-12')
    expect(refreshUser).toHaveBeenCalledOnce()
    expect(fetchActiveSubscriptions).not.toHaveBeenCalled()
    expect(wrapper.emitted('redeemed')?.[0]?.[0]).toMatchObject({ type: 'balance', value: 12 })
    expect(wrapper.text()).toContain('$12.00')
  })

  it('refreshes active subscriptions after a subscription code', async () => {
    redeem.mockResolvedValue({ message: 'ok', type: 'subscription', value: 30, group_name: 'Pro' })
    const wrapper = mount(RedeemCodeForm)

    await wrapper.get('input').setValue('SUB-30')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(fetchActiveSubscriptions).toHaveBeenCalledWith(true)
    expect(wrapper.text()).toContain('Pro')
  })

  it('shows an inline error when redemption fails', async () => {
    redeem.mockRejectedValue(new Error('invalid code'))
    const wrapper = mount(RedeemCodeForm)

    await wrapper.get('input').setValue('BAD')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('invalid code')
    expect(showError).toHaveBeenCalledWith('redeem.redeemFailed')
  })
})
