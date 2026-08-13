<template>
  <section class="card" :aria-busy="loading || saving">
    <header class="border-b border-outline px-4 py-4 sm:px-6">
      <div class="flex min-w-0 items-center justify-between gap-4">
        <div class="min-w-0">
          <h2 class="text-base font-semibold text-foreground sm:text-lg">分销体系</h2>
          <p class="mt-1 text-sm text-foreground-subtle">代理层级、返佣结算和提现规则</p>
        </div>
        <div v-if="settings" class="flex items-center gap-3">
          <span v-if="dirty" class="badge badge-warning" aria-live="polite">有未保存修改</span>
          <span class="hidden text-xs text-foreground-subtle sm:inline">{{ settings.enabled ? '运行中' : '已暂停' }}</span>
          <Toggle v-model="settings.enabled" aria-label="启用分销体系" />
        </div>
      </div>
    </header>

    <div v-if="loading" class="flex min-h-40 items-center justify-center" role="status" aria-label="正在加载分销设置"><LoadingSpinner /></div>
    <div v-else-if="loadError" class="p-4 sm:p-6">
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-panel border border-danger/30 bg-danger-subtle p-4" role="alert">
        <div><p class="text-sm font-medium text-danger-foreground">分销设置加载失败</p><p class="mt-1 text-xs text-foreground-subtle">未显示可能过期的设置，请检查网络后重试。</p></div>
        <button type="button" class="btn btn-secondary min-h-11" @click="load">重试</button>
      </div>
    </div>
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

      <section class="space-y-4 border-t border-outline pt-5" aria-labelledby="distribution-reward-switches-title">
        <div>
          <h3 id="distribution-reward-switches-title" class="font-semibold text-foreground">{{ t('common.distributionRewards.title') }}</h3>
          <p class="mt-1 text-sm text-foreground-subtle">{{ t('common.distributionRewards.description') }}</p>
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="flex min-h-10 items-center justify-between gap-3 rounded-panel border border-outline px-4 py-3">
            <div><p class="form-label">{{ t('common.distributionRewards.registration') }}</p><p class="text-xs text-foreground-subtle">{{ t('common.distributionRewards.registrationHint') }}</p></div>
            <Toggle v-model="settings.registration_reward_enabled" :aria-label="t('common.distributionRewards.registration')" />
          </div>
          <div class="flex min-h-10 items-center justify-between gap-3 rounded-panel border border-outline px-4 py-3">
            <div><p class="form-label">{{ t('common.distributionRewards.recharge') }}</p><p class="text-xs text-foreground-subtle">{{ t('common.distributionRewards.rechargeHint') }}</p></div>
            <Toggle v-model="settings.recharge_reward_enabled" :aria-label="t('common.distributionRewards.recharge')" />
          </div>
        </div>
      </section>

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

      <section class="space-y-4 border-t border-outline pt-5" aria-labelledby="promotion-tracking-settings-title">
        <div class="flex min-w-0 items-center justify-between gap-4">
          <div class="min-w-0">
            <h3 id="promotion-tracking-settings-title" class="font-semibold text-foreground">推广访问追踪</h3>
            <p class="mt-1 text-sm text-foreground-subtle">统计去标识化访问并在有效期内完成注册归因</p>
          </div>
          <div class="flex items-center gap-3"><span class="hidden text-xs text-foreground-subtle sm:inline">{{ settings.promotion_tracking_enabled ? '正在采集' : '已停止采集' }}{{ dirty ? '（未保存）' : '' }}</span><Toggle v-model="settings.promotion_tracking_enabled" aria-label="启用推广访问追踪" /></div>
        </div>
        <p class="rounded-control bg-surface-subtle px-3 py-2 text-xs text-foreground-subtle">关闭后停止记录新访问，但不会删除已有历史数据。已有 Cookie 的历史归因由下方开关独立控制。</p>
        <details class="group rounded-panel border border-outline">
          <summary class="flex min-h-11 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 text-sm font-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-focus"><span>高级采集与归因设置</span><Icon name="chevronDown" size="sm" class="shrink-0 transition-transform group-open:rotate-180" /></summary>
          <div class="grid gap-4 border-t border-outline p-4 sm:grid-cols-2 lg:grid-cols-4">
            <div class="flex min-h-10 items-center justify-between gap-3"><div><p class="form-label">历史访问归因</p><p class="text-xs text-foreground-subtle">注册链接不再带码时使用已有访客 Cookie</p></div><Toggle v-model="settings.promotion_attribution_enabled" aria-label="启用历史访问归因" /></div>
            <label class="form-field"><span class="form-label">归因模式</span><select v-model="settings.promotion_attribution_model" class="input" :disabled="!settings.promotion_attribution_enabled"><option value="first_touch">首次访问</option><option value="last_touch">最近访问</option></select></label>
            <label class="form-field"><span class="form-label">归因有效期</span><div class="relative"><input v-model.number="settings.promotion_attribution_days" class="input pr-12" type="number" min="1" max="365" :disabled="!settings.promotion_attribution_enabled"><span class="input-suffix">天</span></div></label>
            <label class="form-field"><span class="form-label">访问明细保留</span><div class="relative"><input v-model.number="settings.promotion_detail_retention_days" class="input pr-12" type="number" min="30" max="730" :disabled="!settings.promotion_tracking_enabled"><span class="input-suffix">天</span></div><span class="input-hint">到期隐藏并清理敏感明细；原始记录按统计周期延后物理删除</span></label>
            <div class="flex min-h-10 items-center justify-between gap-3"><div><p class="form-label">记录访问来源</p><p class="text-xs text-foreground-subtle">来源域名及 UTM 渠道</p></div><Toggle v-model="settings.promotion_collect_source" :disabled="!settings.promotion_tracking_enabled" aria-label="记录访问来源" /></div>
            <div class="flex min-h-10 items-center justify-between gap-3"><div><p class="form-label">记录设备类型</p><p class="text-xs text-foreground-subtle">仅区分桌面、手机和平板</p></div><Toggle v-model="settings.promotion_collect_device" :disabled="!settings.promotion_tracking_enabled" aria-label="记录设备类型" /></div>
            <div class="flex min-h-10 items-center justify-between gap-3"><div><p class="form-label">过滤机器人</p><p class="text-xs text-foreground-subtle">汇总指标排除已识别机器人</p></div><Toggle v-model="settings.promotion_bot_filter_enabled" :disabled="!settings.promotion_tracking_enabled" aria-label="过滤机器人访问" /></div>
          </div>
        </details>
      </section>

      <div class="space-y-3" :class="dirty ? 'sticky bottom-0 z-20 -mx-4 border-t border-outline bg-surface/95 px-4 py-3 shadow-floating backdrop-blur sm:-mx-6 sm:px-6' : 'border-t border-outline pt-5'">
        <p v-if="validationError" id="distribution-settings-validation" class="rounded-control border border-danger/30 bg-danger-subtle px-3 py-2 text-sm text-danger-foreground" role="alert">{{ validationError }}</p>
        <div class="flex justify-end">
        <button type="button" class="btn btn-primary min-h-11 w-full sm:w-auto" :disabled="saving || !dirty || !!validationError" :aria-describedby="validationError ? 'distribution-settings-validation' : undefined" @click="save">
          <Icon name="check" size="sm" /><span>{{ saving ? '保存中...' : '保存分销设置' }}</span>
        </button>
        </div>
      </div>
    </div>
  </section>
  <ConfirmDialog
    :show="saveConfirmOpen"
    title="确认调整推广追踪"
    :message="saveConfirmMessage"
    confirm-text="确认并保存"
    cancel-text="继续编辑"
    danger
    @confirm="confirmSave"
    @cancel="cancelSaveConfirm"
  />
  <ConfirmDialog
    :show="leaveConfirmOpen"
    title="放弃未保存修改？"
    message="当前分销设置尚未保存，离开后本次修改将丢失。"
    confirm-text="放弃修改"
    cancel-text="留在当前页面"
    danger
    @confirm="confirmLeave"
    @cancel="cancelLeave"
  />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { onBeforeRouteLeave, type NavigationGuardNext } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Toggle from '@/components/common/Toggle.vue'
import { getSettings, updateSettings, type DistributionSettings } from '@/api/admin/distribution'
const emit = defineEmits<{ success: [message: string]; error: [message: string] }>()
const { t } = useI18n()
const loading = ref(false)
const loadError = ref(false)
const saving = ref(false)
const settings = ref<DistributionSettings | null>(null)
const baseline = ref('')
const baselineTracking = ref({ tracking: false, attribution: false })
const saveConfirmOpen = ref(false)
const saveConfirmMessage = ref('')
const pendingSavePayload = ref<DistributionSettings | null>(null)
const pendingSaveFingerprint = ref('')
const leaveConfirmOpen = ref(false)
let pendingNavigation: NavigationGuardNext | null = null
let allowLeave = false
const rates = reactive({ l1Default: 0, l1MaxChild: 0, l2Default: 0, withdrawalFee: 0 })
const toPercent = (bps: number) => Number((bps / 100).toFixed(2))
const toBPS = (percent: number) => Math.round(Number(percent) * 100)
const ratesFor = (value: DistributionSettings) => ({
  l1Default: toPercent(value.l1_default_rate_bps),
  l1MaxChild: toPercent(value.l1_max_child_rate_bps),
  l2Default: toPercent(value.l2_default_rate_bps),
  withdrawalFee: toPercent(value.withdrawal_fee_rate_bps),
})
const serialize = (value: DistributionSettings, rateValues: typeof rates) => JSON.stringify({ settings: value, rates: { ...rateValues } })

function syncRates(value: DistributionSettings) {
  rates.l1Default = toPercent(value.l1_default_rate_bps)
  rates.l1MaxChild = toPercent(value.l1_max_child_rate_bps)
  rates.l2Default = toPercent(value.l2_default_rate_bps)
  rates.withdrawalFee = toPercent(value.withdrawal_fee_rate_bps)
}
const dirty = computed(() => {
  if (!settings.value) return false
  return serialize(settings.value, rates) !== baseline.value
})
const validationError = computed(() => {
  const value = settings.value
  if (!value) return ''
  const validRate = (rate: number) => Number.isFinite(rate) && rate >= 0 && rate <= 100
  if (!validRate(rates.l1Default)) return '一级代理返佣比例必须在 0% 至 100% 之间。'
  if (!validRate(rates.l1MaxChild) || rates.l1MaxChild > rates.l1Default) return '一级代理可设置上限不能超过其返佣比例。'
  if (!validRate(rates.l2Default) || rates.l2Default > rates.l1MaxChild) return '下级代理默认比例不能超过一级代理可设置上限。'
  if (!Number.isInteger(Number(value.freeze_hours)) || Number(value.freeze_hours) < 0) return '佣金冻结时间必须是大于或等于 0 的整数小时。'
  if (!Number.isFinite(Number(value.usd_to_cny)) || Number(value.usd_to_cny) <= 0) return '美元订单计佣汇率必须大于 0。'
  const minimum = Number(value.minimum_withdrawal_cny)
  const maximum = Number(value.maximum_withdrawal_cny)
  if (!Number.isFinite(minimum) || minimum < 0) return '最低提现金额不能小于 0。'
  if (!Number.isFinite(maximum) || maximum < minimum) return '单笔提现上限不能低于最低提现金额。'
  if (!validRate(rates.withdrawalFee)) return '提现手续费率必须在 0% 至 100% 之间。'
  for (const [amount, label] of [
    [value.withdrawal_fee_fixed_cny, '固定手续费'],
    [value.daily_withdrawal_limit_cny, '每日提现限额'],
    [value.monthly_withdrawal_limit_cny, '每月提现限额'],
  ] as const) {
    if (!Number.isFinite(Number(amount)) || Number(amount) < 0) return `${label}不能小于 0。`
  }
  if (!Number.isInteger(Number(value.promotion_attribution_days)) || value.promotion_attribution_days < 1 || value.promotion_attribution_days > 365) return '归因有效期必须是 1 至 365 天的整数。'
  if (!Number.isInteger(Number(value.promotion_detail_retention_days)) || value.promotion_detail_retention_days < 30 || value.promotion_detail_retention_days > 730) return '访问明细保留期必须是 30 至 730 天的整数。'
  return ''
})
function snapshot() {
  if (!settings.value) return
  baseline.value = serialize(settings.value, rates)
  baselineTracking.value = {
    tracking: !!settings.value.promotion_tracking_enabled,
    attribution: !!settings.value.promotion_attribution_enabled,
  }
}

async function load() {
  loading.value = true
  loadError.value = false
  try { settings.value = await getSettings(); syncRates(settings.value); snapshot() }
  catch { settings.value = null; loadError.value = true; emit('error', '加载分销设置失败') }
  finally { loading.value = false }
}

async function save() {
  if (!settings.value || validationError.value) return
  const submittedFingerprint = serialize(settings.value, rates)
  const payload: DistributionSettings = {
    ...settings.value,
    l1_default_rate_bps: toBPS(rates.l1Default),
    l1_max_child_rate_bps: toBPS(rates.l1MaxChild),
    l2_default_rate_bps: toBPS(rates.l2Default),
    withdrawal_fee_rate_bps: toBPS(rates.withdrawalFee),
  }
  const warnings: string[] = []
  if (baselineTracking.value.tracking && !settings.value.promotion_tracking_enabled) warnings.push('停止记录新的推广访问')
  if (baselineTracking.value.attribution && !settings.value.promotion_attribution_enabled) warnings.push('停止将未携带推广码的新注册归因给历史访问')
  if (warnings.length) {
    pendingSavePayload.value = payload
    pendingSaveFingerprint.value = submittedFingerprint
    saveConfirmMessage.value = `保存后将${warnings.join('，并')}。历史汇总数据不会删除，确认继续吗？`
    saveConfirmOpen.value = true
    return
  }
  await persist(payload, submittedFingerprint)
}

async function persist(payload: DistributionSettings, submittedFingerprint: string) {
  saving.value = true
  try {
    await updateSettings(payload)
    const savedRates = ratesFor(payload)
    const hasNewerEdits = !!settings.value && serialize(settings.value, rates) !== submittedFingerprint
    baseline.value = serialize(payload, savedRates)
    baselineTracking.value = { tracking: payload.promotion_tracking_enabled, attribution: payload.promotion_attribution_enabled }
    if (!hasNewerEdits) {
      settings.value = payload
      syncRates(payload)
    }
    emit('success', hasNewerEdits ? '设置已保存；保存期间的新修改仍待保存' : '分销设置已保存')
  } catch {
    emit('error', '保存分销设置失败，请检查比例和金额范围')
  } finally { saving.value = false }
}

async function confirmSave() {
  const payload = pendingSavePayload.value
  const fingerprint = pendingSaveFingerprint.value
  saveConfirmOpen.value = false
  pendingSavePayload.value = null
  pendingSaveFingerprint.value = ''
  if (payload) await persist(payload, fingerprint)
}
function cancelSaveConfirm() { saveConfirmOpen.value = false; pendingSavePayload.value = null; pendingSaveFingerprint.value = '' }
function confirmLeave() { allowLeave = true; leaveConfirmOpen.value = false; const next = pendingNavigation; pendingNavigation = null; next?.() }
function cancelLeave() { leaveConfirmOpen.value = false; const next = pendingNavigation; pendingNavigation = null; next?.(false) }

onBeforeRouteLeave((_to, _from, next) => {
  if (!dirty.value || allowLeave) { next(); return }
  pendingNavigation = next
  leaveConfirmOpen.value = true
})

function handleBeforeUnload(event: BeforeUnloadEvent) {
  if (!dirty.value) return
  event.preventDefault()
  event.returnValue = ''
}

onMounted(() => {
  window.addEventListener('beforeunload', handleBeforeUnload)
  void load()
})
onBeforeUnmount(() => window.removeEventListener('beforeunload', handleBeforeUnload))
</script>

<style scoped>
.input{min-height:44px}
.input-prefix,.input-suffix{position:absolute;top:50%;transform:translateY(-50%);font-size:.75rem;color:var(--ui-text-muted)}
.input-prefix{left:.75rem}.input-suffix{right:.75rem}
</style>
