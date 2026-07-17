import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string, fallback?: string) => fallback ?? key,
  }),
}))

describe('PaymentMethodSelector', () => {
  it('shows the configured display name for custom EasyPay methods', () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: {
        selected: 'ldc',
        methods: [{ type: 'ldc', display_name: 'LDC Pay', fee_rate: 0, available: true }],
      },
    })

    expect(wrapper.text()).toContain('LDC Pay')
    expect(wrapper.text()).not.toContain('ldc')
    expect(wrapper.text()).not.toContain('payment.methods.ldc')
  })

  it('uses the generic selected style for custom methods that contain built-in names', () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: {
        selected: 'card_alipay',
        methods: [{ type: 'card_alipay', display_name: 'Card Pay', fee_rate: 0, available: true }],
      },
    })

    const button = wrapper.get('button')
    expect(button.classes()).toContain('border-brand/30')
    expect(button.classes()).not.toContain('border-[#02A9F1]')
  })

  it('exposes radio semantics and moves selection with arrow keys', async () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: {
        selected: 'alipay',
        methods: [
          { type: 'alipay', fee_rate: 0, available: true },
          { type: 'wxpay', fee_rate: 0, available: false },
          { type: 'stripe', fee_rate: 0, available: true },
        ],
      },
    })

    expect(wrapper.get('[role="radiogroup"]').exists()).toBe(true)
    const radios = wrapper.findAll('[role="radio"]')
    expect(radios[0].attributes('aria-checked')).toBe('true')

    await radios[0].trigger('keydown', { key: 'ArrowRight' })

    expect(wrapper.emitted('select')).toEqual([['stripe']])
  })

  it('does not repeat the payment method label inside zero-fee options', () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: {
        selected: 'alipay',
        methods: [
          { type: 'alipay', fee_rate: 0, available: true },
          { type: 'wxpay', fee_rate: 0, available: true },
        ],
      },
    })

    expect(wrapper.text()).not.toContain('payment.paymentMethodpayment.paymentMethod')
    expect(wrapper.findAll('button')).toHaveLength(2)
    expect(wrapper.findAll('button span.text-foreground-subtle')).toHaveLength(0)
  })

  it('keeps compact payment options short and reserves a stable selection indicator', () => {
    const wrapper = mount(PaymentMethodSelector, {
      props: {
        compact: true,
        selected: 'alipay',
        methods: [
          { type: 'alipay', fee_rate: 0, available: true },
          { type: 'wxpay', fee_rate: 0, available: true },
        ],
      },
    })

    const radios = wrapper.findAll('[role="radio"]')
    expect(radios).toHaveLength(2)
    expect(radios.every(radio => radio.classes().includes('min-h-[52px]'))).toBe(true)
    expect(wrapper.findAll('[data-testid="method-selection-indicator"]')).toHaveLength(2)
  })
})
