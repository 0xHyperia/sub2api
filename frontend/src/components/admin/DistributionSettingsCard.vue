<template>
  <section class="card">
    <header class="border-b border-outline px-4 py-4 sm:px-6">
      <div class="flex min-w-0 items-center justify-between gap-4">
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-foreground sm:text-lg">分销体系</h2>
          <p class="mt-1 text-sm text-foreground-subtle">代理层级、返佣结算和提现规则</p>
        </div>
        <Toggle v-if="settings" v-model="settings.enabled" aria-label="启用分销体系" />
      </div>
    </header>

    <div v-if="loading" class="flex min-h-40 items-center justify-center"><LoadingSpinner /></div>
    <div v-else-if="settings" class="space-y-6 p-4 sm:p-6">
      <div class="grid gap-4 lg:grid-cols-3">
        <label class="form-field">
          <span class="form-label">一级代理返佣比例</span>
          <div class="relative"><input v-model.number="rates.l1Default" class="input pr-10" type="number" min="0" max="100" step="0.01"><span class="input-suffix">%</span></div>
        </label>
        <label class="form-field">
          <span class="form-label">一级代理可设置上限</span>
          <div class="relative"><input v-model.number="rates.l1MaxChild" class="input pr-10" type="number" min="0" :max="rates.l1Default" step="0.01"><span class="input-suffix">%</span></div>
        </label>
        <label class="form-field">
          <span class="form-label">二级代理默认比例</span>
          <div class="relative"><input v-model.number="rates.l2Default" class="input pr-10" type="number" min="0" :max="rates.l1MaxChild" step="0.01"><span class="input-suffix">%</span></div>
        </label>
      </div>

      <div class="grid gap-4 border-t border-outline pt-5 sm:grid-cols-2 lg:grid-cols-4">
        <label class="form-field"><span class="form-label">佣金冻结时间</span><div class="relative"><input v-model.number="settings.freeze_hours" class="input pr-14" type="number" min="0"><span class="input-suffix">小时</span></div></label>
        <label class="form-field"><span class="form-label">美元订单计佣汇率</span><div class="relative"><input v-model="settings.usd_to_cny" class="input pr-14" type="number" min="0.0001" step="0.0001"><span class="input-suffix">CNY</span></div><span class="input-hint">1 USD 实付折算为多少 CNY 计佣；人民币订单固定按 1:1</span></label>
        <label class="form-field"><span class="form-label">最低提现金额</span><div class="relative"><input v-model="settings.minimum_withdrawal_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <label class="form-field"><span class="form-label">单笔提现上限</span><div class="relative"><input v-model="settings.maximum_withdrawal_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
      </div>

      <div class="grid gap-4 border-t border-outline pt-5 sm:grid-cols-2 lg:grid-cols-4">
        <div class="flex min-h-10 items-center justify-between gap-3 sm:col-span-2 lg:col-span-1">
          <div><p class="form-label">允许代理提现</p><p class="text-xs text-foreground-subtle">关闭后只允许转入平台余额</p></div>
          <Toggle v-model="settings.withdrawal_enabled" aria-label="允许代理提现" />
        </div>
        <div class="flex min-h-10 items-center justify-between gap-3 sm:col-span-2 lg:col-span-1">
          <div><p class="form-label">提现双人复核</p><p class="text-xs text-foreground-subtle">审核人与打款操作人必须不同</p></div>
          <Toggle v-model="settings.withdrawal_dual_approval_enabled" aria-label="提现双人复核" />
        </div>
        <label class="form-field"><span class="form-label">提现手续费率</span><div class="relative"><input v-model.number="rates.withdrawalFee" class="input pr-10" type="number" min="0" max="100" step="0.01"><span class="input-suffix">%</span></div></label>
        <label class="form-field"><span class="form-label">固定手续费</span><div class="relative"><input v-model="settings.withdrawal_fee_fixed_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <label class="form-field"><span class="form-label">每日提现限额</span><div class="relative"><input v-model="settings.daily_withdrawal_limit_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
        <label class="form-field"><span class="form-label">每月提现限额</span><div class="relative"><input v-model="settings.monthly_withdrawal_limit_cny" class="input pl-8" type="number" min="0" step="0.01"><span class="input-prefix">¥</span></div></label>
      </div>

      <div class="flex justify-end border-t border-outline pt-5">
        <button type="button" class="btn btn-primary w-full sm:w-auto" :disabled="saving" @click="save">
          <Icon name="check" size="sm" /><span>{{ saving ? '保存中...' : '保存分销设置' }}</span>
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Toggle from '@/components/common/Toggle.vue'
import { getSettings, updateSettings, type DistributionSettings } from '@/api/admin/distribution'
const emit = defineEmits<{ success: [message: string]; error: [message: string] }>()
const loading = ref(false)
const saving = ref(false)
const settings = ref<DistributionSettings | null>(null)
const rates = reactive({ l1Default: 0, l1MaxChild: 0, l2Default: 0, withdrawalFee: 0 })
const toPercent = (bps: number) => Number((bps / 100).toFixed(2))
const toBPS = (percent: number) => Math.round(Number(percent) * 100)

function syncRates(value: DistributionSettings) {
  rates.l1Default = toPercent(value.l1_default_rate_bps)
  rates.l1MaxChild = toPercent(value.l1_max_child_rate_bps)
  rates.l2Default = toPercent(value.l2_default_rate_bps)
  rates.withdrawalFee = toPercent(value.withdrawal_fee_rate_bps)
}

async function load() {
  loading.value = true
  try { settings.value = await getSettings(); syncRates(settings.value) }
  catch { emit('error', '加载分销设置失败') }
  finally { loading.value = false }
}

async function save() {
  if (!settings.value) return
  const payload: DistributionSettings = {
    ...settings.value,
    l1_default_rate_bps: toBPS(rates.l1Default),
    l1_max_child_rate_bps: toBPS(rates.l1MaxChild),
    l2_default_rate_bps: toBPS(rates.l2Default),
    withdrawal_fee_rate_bps: toBPS(rates.withdrawalFee),
  }
  saving.value = true
  try {
    await updateSettings(payload)
    settings.value = payload
    emit('success', '分销设置已保存')
  } catch {
    emit('error', '保存分销设置失败，请检查比例和金额范围')
  } finally { saving.value = false }
}

onMounted(load)
</script>

<style scoped>
.input-prefix,.input-suffix{position:absolute;top:50%;transform:translateY(-50%);font-size:.75rem;color:var(--ui-text-muted)}
.input-prefix{left:.75rem}.input-suffix{right:.75rem}
</style>
