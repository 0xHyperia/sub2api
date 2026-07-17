import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { formatCurrency } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'
import { useClipboard } from '@/composables/useClipboard'

export function useAffiliateRewards() {
  const { t } = useI18n()
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const { copyToClipboard } = useClipboard()

  const loading = ref(true)
  const loadFailed = ref(false)
  const transferring = ref(false)
  const detail = ref<UserAffiliateDetail | null>(null)

  const inviteLink = computed(() => {
    if (!detail.value) return ''
    const path = `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
    return typeof window === 'undefined' ? path : `${window.location.origin}${path}`
  })

  const formattedRebateRate = computed(() => {
    const value = detail.value?.effective_rebate_rate_percent ?? 0
    const rounded = Math.round(value * 100) / 100
    return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
  })

  async function loadAffiliateDetail(silent = false): Promise<void> {
    if (!silent) {
      loading.value = true
      loadFailed.value = false
    }
    try {
      detail.value = await userAPI.getAffiliateDetail()
    } catch (error) {
      loadFailed.value = detail.value === null
      if (!silent) {
        appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
      }
    } finally {
      if (!silent) loading.value = false
    }
  }

  async function copyCode(): Promise<void> {
    if (!detail.value?.aff_code) return
    await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
  }

  async function copyInviteLink(): Promise<void> {
    if (!inviteLink.value) return
    await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
  }

  async function transferQuota(): Promise<void> {
    if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
    transferring.value = true
    try {
      const response = await userAPI.transferAffiliateQuota()
      appStore.showSuccess(t('affiliate.transfer.success', {
        amount: formatCurrency(response.transferred_quota),
      }))
      await Promise.all([
        loadAffiliateDetail(true),
        authStore.refreshUser().catch(() => undefined),
      ])
    } catch (error) {
      appStore.showError(extractApiErrorMessage(error, t('affiliate.transferFailed')))
    } finally {
      transferring.value = false
    }
  }

  return {
    detail,
    loading,
    loadFailed,
    transferring,
    inviteLink,
    formattedRebateRate,
    loadAffiliateDetail,
    copyCode,
    copyInviteLink,
    transferQuota,
  }
}
