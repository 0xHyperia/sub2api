import { enableAutoUnmount, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import AmountInput from '../AmountInput.vue'
enableAutoUnmount(afterEach)

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

function mountInput(value: number | null = null) {
  return mount(AmountInput, { props: { modelValue: value } })
}

describe('recharge amount input', () => {
  it.each(['10abc', '10.555', '-10', '1e2'])('restores the accepted amount after rejecting %s', async (value) => {
    const wrapper = mountInput(10)
    const input = wrapper.get('input')
    await input.setValue(value)
    expect((input.element as HTMLInputElement).value).toBe('10')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })

  it('restores the last typed amount rather than a stale prop', async () => {
    const wrapper = mountInput()
    const input = wrapper.get('input')
    await input.setValue('12.50')
    await input.setValue('12.500')
    expect((input.element as HTMLInputElement).value).toBe('12.50')
    expect(wrapper.emitted('update:modelValue')).toEqual([[12.5]])
  })

  it('preserves decimal editing and allows clearing the amount', async () => {
    const wrapper = mountInput()
    const input = wrapper.get('input')
    for (const value of ['0', '0.', '0.5', '0.50', '']) await input.setValue(value)
    expect(wrapper.emitted('update:modelValue')).toEqual([[null], [null], [0.5], [0.5], [null]])
    expect((input.element as HTMLInputElement).value).toBe('')
  })
})
