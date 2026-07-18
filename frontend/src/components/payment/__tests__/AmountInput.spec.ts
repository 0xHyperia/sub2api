import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AmountInput from '../AmountInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, params?: Record<string, string>) =>
      key === 'payment.quickAmountBonus' ? `赠送 $${params?.bonus}` : key,
  }),
}))

function mountAmountInput(customEnabled = true) {
  return mount(AmountInput, {
    props: {
      modelValue: null,
      customEnabled,
      currency: 'USD',
      amounts: [
        { amount: 50, bonus: 5 },
        { amount: 100, bonus: 0 },
      ],
    },
  })
}

describe('AmountInput', () => {
  it('shows preset bonuses and emits the selected payment amount', async () => {
    const wrapper = mountAmountInput()

    expect(wrapper.text()).toContain('赠送 $5.00')
    await wrapper.findAll('button')[0].trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([[50]])
  })

  it('hides the custom amount input when custom recharge is disabled', () => {
    const wrapper = mountAmountInput(false)

    expect(wrapper.find('#custom-payment-amount').exists()).toBe(false)
    expect(wrapper.findAll('button')).toHaveLength(2)
  })
})
