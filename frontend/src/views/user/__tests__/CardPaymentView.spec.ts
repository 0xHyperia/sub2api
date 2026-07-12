import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CardPaymentView from '@/views/user/CardPaymentView.vue'
import type { CardCheckoutInfo } from '@/types/payment'

const {
  getCardCheckoutInfoMock,
  getCardPriceMock,
  createCardOrderMock,
  updateCardGoodsOverrideMock,
  showErrorMock,
  showSuccessMock,
  authStoreState,
} = vi.hoisted(() => ({
  getCardCheckoutInfoMock: vi.fn(),
  getCardPriceMock: vi.fn(),
  createCardOrderMock: vi.fn(),
  updateCardGoodsOverrideMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  authStoreState: {
    isAdmin: false,
  },
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getCardCheckoutInfo: getCardCheckoutInfoMock,
    getCardPrice: getCardPriceMock,
    createCardOrder: createCardOrderMock,
  },
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: {
    updateCardGoodsOverride: updateCardGoodsOverrideMock,
  },
  default: {
    updateCardGoodsOverride: updateCardGoodsOverrideMock,
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStoreState,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

function createCheckoutInfo(): CardCheckoutInfo {
  return {
    shops: [{
      provider_instance_id: 'provider-1',
      name: 'Test shop',
      shop: {
        token: 'shop-token',
        nickname: 'Test shop',
      },
      categories: [{ id: 1, name: 'Top ups' }],
      goods: [
        {
          provider_instance_id: 'provider-1',
          shop_token: 'shop-token',
          goods_key: 'goods-a',
          name: 'Starter pack',
          price: 10,
          category_id: 1,
          stock_count: 10,
          limit_count: 5,
          query_password_required: false,
        },
        {
          provider_instance_id: 'provider-1',
          shop_token: 'shop-token',
          goods_key: 'goods-b',
          name: 'Pro pack',
          price: 20,
          category_id: 1,
          stock_count: 8,
          limit_count: 5,
          query_password_required: false,
        },
        {
          provider_instance_id: 'provider-1',
          shop_token: 'shop-token',
          goods_key: 'goods-sold-out',
          name: 'Sold out pack',
          price: 30,
          category_id: 1,
          stock_count: 0,
          limit_count: 5,
          query_password_required: false,
        },
      ],
      channels: [{
        id: 7,
        name: 'alipay',
        show_name: 'Alipay',
        status: 1,
      }],
    }],
  }
}

function mountView() {
  return mount(CardPaymentView, {
    global: {
      stubs: {
        Icon: true,
        PaymentStatusPanel: true,
      },
    },
  })
}

describe('CardPaymentView', () => {
  beforeEach(() => {
    getCardCheckoutInfoMock.mockReset()
    getCardPriceMock.mockReset()
    createCardOrderMock.mockReset()
    updateCardGoodsOverrideMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    authStoreState.isAdmin = false
    getCardPriceMock.mockResolvedValue({
      original_amount: 10,
      total_amount: 10,
      fee: 0,
    })
  })

  it('shows an inline load error with retry instead of the empty state', async () => {
    getCardCheckoutInfoMock.mockRejectedValueOnce({ message: 'Catalog unavailable' })
    const wrapper = mountView()

    await flushPromises()

    const loadError = wrapper.get('[data-testid="card-payment-load-error"]')
    expect(loadError.text()).toContain('Catalog unavailable')
    expect(wrapper.text()).not.toContain('payment.card.empty')

    getCardCheckoutInfoMock.mockResolvedValueOnce(createCheckoutInfo())
    await loadError.get('button').trigger('click')
    await flushPromises()

    expect(getCardCheckoutInfoMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="card-payment-load-error"]').exists()).toBe(false)
    expect(wrapper.get('[data-goods-key="goods-a"]').text()).toContain('Starter pack')

    wrapper.unmount()
  })

  it('shows the extracted order error and restores the submit button', async () => {
    getCardCheckoutInfoMock.mockResolvedValue(createCheckoutInfo())
    createCardOrderMock.mockRejectedValue({ message: 'Inventory changed, try again' })
    const wrapper = mountView()

    await flushPromises()

    const submitButton = wrapper.get<HTMLButtonElement>('[data-testid="card-payment-submit"]')
    expect(submitButton.attributes('disabled')).toBeUndefined()

    await submitButton.trigger('click')
    await flushPromises()

    expect(createCardOrderMock).toHaveBeenCalledWith(expect.objectContaining({
      goods_key: 'goods-a',
      channel_id: 7,
    }))
    expect(wrapper.get('[data-testid="card-payment-order-error"]').text()).toBe('Inventory changed, try again')
    expect(submitButton.attributes('disabled')).toBeUndefined()

    wrapper.unmount()
  })

  it('uses real radio buttons and supports click and arrow-key selection', async () => {
    getCardCheckoutInfoMock.mockResolvedValue(createCheckoutInfo())
    const wrapper = mountView()

    await flushPromises()

    expect(wrapper.get('[role="radiogroup"]').exists()).toBe(true)
    const radios = wrapper.findAll<HTMLButtonElement>('[role="radio"]')
    expect(radios).toHaveLength(3)
    expect(radios[0].element.tagName).toBe('BUTTON')
    expect(radios[0].attributes('aria-checked')).toBe('true')
    expect(radios[2].attributes('disabled')).toBeDefined()

    await radios[1].trigger('click')
    expect(radios[1].attributes('aria-checked')).toBe('true')

    await radios[1].trigger('keydown', { key: 'ArrowRight' })
    expect(radios[0].attributes('aria-checked')).toBe('true')

    wrapper.unmount()
  })
})
