<template>
  <BaseDialog
    :show="show"
    :title="operation === 'add' ? t('admin.users.deposit') : t('admin.users.withdraw')"
    width="narrow"
    @close="handleClose"
  >
    <form v-if="user" id="balance-form" class="space-y-4" @submit.prevent="handleBalanceSubmit">
      <div class="flex min-w-0 items-center gap-3 rounded-panel border border-outline bg-surface-subtle p-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-control border border-outline bg-surface text-sm font-semibold text-foreground-muted" aria-hidden="true">
          {{ user.email.charAt(0).toUpperCase() }}
        </div>
        <div class="min-w-0 flex-1">
          <p class="break-all text-sm font-medium text-foreground">{{ user.email }}</p>
          <p class="mt-0.5 text-xs text-foreground-subtle">
            {{ t('admin.users.currentBalance') }}:
            <span class="font-medium tabular-nums text-foreground">${{ formatBalance(user.balance) }}</span>
          </p>
        </div>
      </div>

      <div>
        <label for="balance-amount" class="input-label">
          {{ operation === 'add' ? t('admin.users.depositAmount') : t('admin.users.withdrawAmount') }}
        </label>
        <div class="flex flex-col gap-2 sm:flex-row">
          <div class="relative min-w-0 flex-1">
            <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm font-medium text-foreground-subtle" aria-hidden="true">$</span>
            <input
              id="balance-amount"
              v-model.number="form.amount"
              type="number"
              inputmode="decimal"
              step="any"
              min="0"
              :max="operation === 'subtract' ? user.balance : undefined"
              required
              class="input pl-8 tabular-nums"
            />
          </div>
          <button
            v-if="operation === 'subtract'"
            type="button"
            class="btn btn-secondary min-h-touch whitespace-nowrap sm:min-h-control"
            @click="fillAllBalance"
          >
            {{ t('admin.users.withdrawAll') }}
          </button>
        </div>
      </div>

      <div>
        <label for="balance-notes" class="input-label">{{ t('admin.users.notes') }}</label>
        <textarea id="balance-notes" v-model="form.notes" rows="3" class="input resize-y"></textarea>
      </div>

      <output
        v-if="form.amount > 0"
        class="flex items-center justify-between gap-3 rounded-panel border border-outline bg-surface-subtle px-3 py-2.5 text-sm"
        for="balance-amount"
        aria-live="polite"
      >
        <span class="text-foreground-muted">{{ t('admin.users.newBalance') }}:</span>
        <span
          class="break-all text-right font-semibold tabular-nums"
          :class="operation === 'add' ? 'text-success-foreground' : 'text-warning-foreground'"
        >
          ${{ formatBalance(calculateNewBalance()) }}
        </span>
      </output>
    </form>

    <template #footer>
      <div class="grid w-full grid-cols-2 gap-2 sm:flex sm:justify-end">
        <button type="button" class="btn btn-secondary" @click="handleClose">{{ t('common.cancel') }}</button>
        <button
          type="submit"
          form="balance-form"
          :disabled="submitting || !form.amount"
          class="btn"
          :class="operation === 'add' ? 'btn-success' : 'btn-danger'"
        >
          {{ submitting ? t('common.saving') : t('common.confirm') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null; operation: 'add' | 'subtract' }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const submitting = ref(false)
const form = reactive({ amount: 0, notes: '' })
let submitRequestSeq = 0

function resetForm() {
  form.amount = 0
  form.notes = ''
  submitting.value = false
}

watch(
  [() => props.show, () => props.user?.id, () => props.operation],
  ([show]) => {
    submitRequestSeq += 1
    if (show) resetForm()
  },
  { immediate: true }
)

// 格式化余额：显示完整精度，去除尾部多余的0
const formatBalance = (value: number) => {
  if (value === 0) return '0.00'
  // 最多保留8位小数，去除尾部的0
  const formatted = value.toFixed(8).replace(/\.?0+$/, '')
  // 确保至少有2位小数
  const parts = formatted.split('.')
  if (parts.length === 1) return formatted + '.00'
  if (parts[1].length === 1) return formatted + '0'
  return formatted
}

// 填入全部余额
const fillAllBalance = () => {
  if (props.user) {
    form.amount = props.user.balance
  }
}

const calculateNewBalance = () => {
  if (!props.user) return 0
  const result = props.operation === 'add' ? props.user.balance + form.amount : props.user.balance - form.amount
  // 避免浮点数精度问题导致的 -0.00 显示
  return Math.abs(result) < 1e-10 ? 0 : result
}
const handleBalanceSubmit = async () => {
  const user = props.user
  if (!user || !props.show) return
  const userId = user.id
  if (!form.amount || form.amount <= 0) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }
  // 退款时验证金额不超过实际余额
  if (props.operation === 'subtract' && form.amount > user.balance) {
    appStore.showError(t('admin.users.insufficientBalance'))
    return
  }
  const requestSeq = ++submitRequestSeq
  submitting.value = true
  try {
    await adminAPI.users.updateBalance(userId, form.amount, props.operation, form.notes)
    if (requestSeq !== submitRequestSeq || !props.show || props.user?.id !== userId) return
    appStore.showSuccess(t('common.success'))
    emit('success')
    handleClose()
  } catch (e: any) {
    if (requestSeq !== submitRequestSeq || !props.show || props.user?.id !== userId) return
    console.error('Failed to update balance:', e)
    appStore.showError(extractApiErrorMessage(e, t('common.error')))
  } finally {
    if (requestSeq === submitRequestSeq) submitting.value = false
  }
}

function handleClose() {
  submitRequestSeq += 1
  resetForm()
  emit('close')
}

onUnmounted(() => {
  submitRequestSeq += 1
})
</script>
