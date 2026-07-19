import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminPaymentPlansView from '../AdminPaymentPlansView.vue'

const { getPlans, getConfig, updatePlan, deletePlan, getAllGroups, showError, showSuccess } = vi.hoisted(() => ({
  getPlans: vi.fn(),
  getConfig: vi.fn(),
  updatePlan: vi.fn(),
  deletePlan: vi.fn(),
  getAllGroups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn()
}))

vi.mock('@/api/admin/payment', () => ({
  adminPaymentAPI: { getPlans, getConfig, updatePlan, deletePlan }
}))

vi.mock('@/api/admin', () => ({
  default: { groups: { getAll: getAllGroups } }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.id">
        <slot name="mobile-card" :row="row" :index="0" :selected="false" :expanded="false" />
        <div :data-test="'desktop-plan-name-' + row.id"><slot name="cell-name" :value="row.name" :row="row" /></div>
      </div>
    </div>
  `
}

const PlanEditDialogStub = {
  props: ['show', 'plan'],
  template: '<div v-if="show" data-test="plan-edit-dialog">{{ plan?.name }}</div>'
}

const ConfirmDialogStub = {
  props: ['show'],
  template: '<div v-if="show" data-test="delete-plan-dialog"></div>'
}

function mountView() {
  return mount(AdminPaymentPlansView, {
    global: {
      stubs: {
        AppLayout: { template: '<main><slot /></main>' },
        TablePageLayout: { template: '<div><slot name="filters" /><slot name="table" /></div>' },
        DataTable: DataTableStub,
        PlanEditDialog: PlanEditDialogStub,
        ConfirmDialog: ConfirmDialogStub,
        GroupBadge: { props: ['name'], template: '<span data-test="plan-group">{{ name }}</span>' },
        Toggle: true,
        Icon: true
      }
    }
  })
}

describe('AdminPaymentPlansView mobile cards', () => {
  beforeEach(() => {
    getPlans.mockReset().mockResolvedValue({
      data: [{
        id: 8,
        group_id: 4,
        name: 'Pro Monthly',
        description: 'Production plan',
        price: 29.9,
        original_price: 39.9,
        currency: 'USD',
        validity_days: 1,
        validity_unit: 'months',
        features: 'Feature A\nFeature B',
        for_sale: true,
        sort_order: 10
      }]
    })
    getConfig.mockReset().mockResolvedValue({ data: {} })
    updatePlan.mockReset().mockResolvedValue(undefined)
    deletePlan.mockReset().mockResolvedValue(undefined)
    getAllGroups.mockReset().mockResolvedValue([{ id: 4, name: 'OpenAI Pro', platform: 'openai', rate_multiplier: 1 }])
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('shows name, price, period, group and sale status while keeping desktop name cell', async () => {
    const wrapper = mountView()
    await flushPromises()

    const card = wrapper.get('[data-test="payment-plan-mobile-card-8"]')
    expect(card.text()).toContain('Pro Monthly')
    expect(card.text()).toContain('$29.90')
    expect(card.text()).toContain('$39.90')
    expect(card.text()).toContain('1 payment.admin.months')
    expect(card.text()).toContain('OpenAI Pro')
    expect(card.text()).toContain('payment.admin.onSale')
    expect(wrapper.get('[data-test="desktop-plan-name-8"]').text()).toContain('Pro Monthly')
  })

  it('keeps edit explicit and layers sale toggle and delete under More', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="payment-plan-mobile-edit-8"]').trigger('click')
    expect(wrapper.get('[data-test="plan-edit-dialog"]').text()).toContain('Pro Monthly')

    expect(wrapper.get('[data-test="payment-plan-mobile-more-8"]').exists()).toBe(true)
    await wrapper.get('[data-test="payment-plan-mobile-toggle-8"]').trigger('click')
    await flushPromises()
    expect(updatePlan).toHaveBeenCalledWith(8, { for_sale: false })
    expect(wrapper.get('[data-test="payment-plan-mobile-card-8"]').text()).toContain('payment.admin.offSale')

    await wrapper.get('[data-test="payment-plan-mobile-delete-8"]').trigger('click')
    expect(wrapper.get('[data-test="delete-plan-dialog"]').exists()).toBe(true)
  })
})
