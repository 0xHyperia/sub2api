import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAffiliateRewards } from '@/composables/useAffiliateRewards'

const getAffiliateDetail = vi.hoisted(() => vi.fn())
const transferAffiliateQuota = vi.hoisted(() => vi.fn())
const refreshUser = vi.hoisted(() => vi.fn())
const copyToClipboard = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/api/user', () => ({
  default: { getAffiliateDetail, transferAffiliateQuota },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ refreshUser }),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess, showError }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard }),
}))

const detail = {
  user_id: 1,
  aff_code: 'INVITE123',
  aff_count: 3,
  aff_quota: 8.5,
  aff_frozen_quota: 1,
  aff_history_quota: 24,
  effective_rebate_rate_percent: 12.5,
  invitees: [],
}

describe('useAffiliateRewards', () => {
  beforeEach(() => {
    getAffiliateDetail.mockReset().mockResolvedValue(detail)
    transferAffiliateQuota.mockReset().mockResolvedValue({ transferred_quota: 8.5, balance: 20 })
    refreshUser.mockReset().mockResolvedValue(undefined)
    copyToClipboard.mockReset().mockResolvedValue(true)
    showSuccess.mockReset()
    showError.mockReset()
  })

  it('loads reward data and builds the registration link', async () => {
    const affiliate = useAffiliateRewards()
    await affiliate.loadAffiliateDetail()

    expect(affiliate.detail.value).toEqual(detail)
    expect(affiliate.formattedRebateRate.value).toBe('12.5')
    expect(affiliate.inviteLink.value).toContain('/register?aff=INVITE123')

    await affiliate.copyInviteLink()
    expect(copyToClipboard).toHaveBeenCalledWith(affiliate.inviteLink.value, 'affiliate.linkCopied')
  })

  it('transfers available quota and refreshes both rewards and balance', async () => {
    const affiliate = useAffiliateRewards()
    await affiliate.loadAffiliateDetail()
    await affiliate.transferQuota()

    expect(transferAffiliateQuota).toHaveBeenCalledOnce()
    expect(getAffiliateDetail).toHaveBeenCalledTimes(2)
    expect(refreshUser).toHaveBeenCalledOnce()
    expect(showSuccess).toHaveBeenCalledOnce()
  })
})
