import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import RechargeValueEstimator from '../RechargeValueEstimator.vue'
import Select from '@/components/common/Select.vue'
import { formatPaymentAmount } from '../currency'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const getMarketplace = vi.hoisted(() => vi.fn())
const getUserGroupRates = vi.hoisted(() => vi.fn())

vi.mock('@/api/channels', () => ({
  default: {
    getMarketplace,
  },
}))

vi.mock('@/api/groups', () => ({
  default: {
    getUserGroupRates,
  },
}))

const marketplaceFixture = [
      {
        platform: 'openai',
        groups: [],
        supported_models: [
          {
            name: 'gpt-5.6',
            platform: 'openai',
            pricing: null,
            groups: [
              {
                id: 7,
                name: 'Official 10%',
                platform: 'openai',
                subscription_type: 'standard',
                rate_multiplier: 0.1,
                peak_rate_enabled: false,
                peak_start: '',
                peak_end: '',
                peak_rate_multiplier: 1,
                is_exclusive: false,
              },
            ],
          },
        ],
      },
    ]

describe('RechargeValueEstimator', () => {
  beforeEach(() => {
    getMarketplace.mockReset().mockResolvedValue(marketplaceFixture)
    getUserGroupRates.mockReset().mockResolvedValue({})
    document.body.style.overflow = ''
  })

  it('applies recharge bonus tiers to the estimated platform balance', async () => {
    const wrapper = shallowMount(RechargeValueEstimator, {
      props: {
        creditedAmount: 10,
        rechargeAmount: 10,
        bonusTiers: [{ min_amount: 10, bonus_percent: 50 }],
        bonusMode: 'bonus',
      },
      global: { stubs: { Teleport: true, Transition: false } },
    })
    await wrapper.get('button[aria-controls="recharge-estimator-drawer"]').trigger('click')
    await flushPromises()
    await flushPromises()
    // 赠金 +50%：到账 15，官方等值为未赠送时的 1.5 倍（10 -> $100.00 则 15 -> $150.00）
    expect(wrapper.get('[data-testid="official-equivalent"]').text()).toContain('$150.00')
  })

  it('lazy loads model rates and estimates official-price-equivalent usage', async () => {
    const wrapper = shallowMount(RechargeValueEstimator, {
      props: { creditedAmount: 10 },
      global: {
        stubs: {
          Teleport: true,
          Transition: false,
        },
      },
    })

    expect(wrapper.findComponent(Select).exists()).toBe(false)
    expect(getMarketplace).not.toHaveBeenCalled()
    await wrapper.get('button[aria-controls="recharge-estimator-drawer"]').trigger('click')
    await flushPromises()
    await flushPromises()

    expect(getMarketplace).toHaveBeenCalledTimes(1)
    const selectors = wrapper.findAllComponents(Select)
    expect(selectors).toHaveLength(2)
    expect(selectors.every(selector => selector.props('searchable') === true)).toBe(true)
    expect(wrapper.get('[data-testid="official-equivalent"]').text()).toContain('$100.00')
    expect(wrapper.get('[data-testid="official-exchange-usd"]').text()).toContain('$1.39')
    expect(wrapper.get('[data-testid="official-cny-cost"]').text()).toContain(formatPaymentAmount(720, 'CNY'))
    expect(document.body.style.overflow).toBe('hidden')

    await wrapper.get('[data-testid="estimator-amount-input"]').setValue('20')
    expect(wrapper.get('[data-testid="official-equivalent"]').text()).toContain('$200.00')
    expect(wrapper.get('[data-testid="official-cny-cost"]').text()).toContain(formatPaymentAmount(1440, 'CNY'))

    const quickAmount = wrapper.findAll('button').find(button => button.text() === '¥50')
    expect(quickAmount).toBeDefined()
    await quickAmount!.trigger('click')
    expect(wrapper.get('[data-testid="official-equivalent"]').text()).toContain('$500.00')

    const dropdownEscape = new KeyboardEvent('keydown', { key: 'Escape', cancelable: true })
    dropdownEscape.preventDefault()
    document.dispatchEvent(dropdownEscape)
    await flushPromises()
    expect(wrapper.find('[data-testid="official-equivalent"]').exists()).toBe(true)

    await wrapper.get('button[aria-label="common.close"]').trigger('click')
    await flushPromises()
    expect(document.body.style.overflow).toBe('')

    await wrapper.get('button[aria-controls="recharge-estimator-drawer"]').trigger('click')
    await flushPromises()
    expect(getMarketplace).toHaveBeenCalledTimes(1)
  })

  it('defaults to the first marketplace model in the lowest-rate group', async () => {
    getMarketplace.mockResolvedValueOnce([
      {
        platform: 'openai',
        groups: [],
        supported_models: [
          {
            name: 'gpt-6',
            platform: 'openai',
            pricing: null,
            monitor_status: { display_order: 100 },
            groups: [{
              id: 1,
              name: 'Premium',
              platform: 'openai',
              subscription_type: 'standard',
              rate_multiplier: 0.5,
              peak_rate_enabled: false,
              peak_start: '',
              peak_end: '',
              peak_rate_multiplier: 1,
              is_exclusive: false,
            }],
          },
          {
            name: 'gpt-5.5-beta',
            platform: 'openai',
            pricing: null,
            monitor_status: { display_order: 1 },
            groups: [{
              id: 2,
              name: 'Value',
              platform: 'openai',
              subscription_type: 'standard',
              rate_multiplier: 0.1,
              peak_rate_enabled: false,
              peak_start: '',
              peak_end: '',
              peak_rate_multiplier: 1,
              is_exclusive: false,
            }],
          },
          {
            name: 'gpt-5.5-alpha',
            platform: 'openai',
            pricing: null,
            monitor_status: { display_order: 20 },
            groups: [{
              id: 2,
              name: 'Value',
              platform: 'openai',
              subscription_type: 'standard',
              rate_multiplier: 0.1,
              peak_rate_enabled: false,
              peak_start: '',
              peak_end: '',
              peak_rate_multiplier: 1,
              is_exclusive: false,
            }],
          },
        ],
      },
    ])
    const wrapper = shallowMount(RechargeValueEstimator, {
      props: { creditedAmount: 10 },
      global: { stubs: { Teleport: true, Transition: false } },
    })

    await wrapper.get('button[aria-controls="recharge-estimator-drawer"]').trigger('click')
    await flushPromises()
    await flushPromises()

    const selectors = wrapper.findAllComponents(Select)
    expect(selectors[0].props('modelValue')).toBe('openai::gpt-5.5-alpha')
    expect(selectors[1].props('modelValue')).toBe(2)
    expect(wrapper.get('[data-testid="official-equivalent"]').text()).toContain('$100.00')
  })
})
