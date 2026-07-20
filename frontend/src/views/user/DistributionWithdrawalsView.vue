<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-[1440px] space-y-5">
      <header class="flex min-w-0 items-center justify-between gap-3">
        <div class="min-w-0"><h1 class="page-title">佣金结算</h1><p class="page-description">提现到收款账户或转换为平台余额</p></div>
        <button class="btn btn-secondary btn-icon shrink-0" title="刷新" aria-label="刷新" :disabled="loading" @click="load"><Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" /></button>
      </header>
      <DistributionNav />

      <div v-if="loading && !overview" class="card flex min-h-48 items-center justify-center"><LoadingSpinner /></div>
      <template v-else-if="overview && rules">
        <section v-if="overview.agent.status !== 'active'" class="rounded-panel border border-warning/40 bg-warning-subtle p-4 text-sm text-warning-foreground">代理账号当前{{ overview.agent.status === 'suspended' ? '已暂停' : '已撤销' }}，暂时不能发起新的资金结算。</section>

        <section class="grid gap-4 lg:grid-cols-[minmax(0,1.25fr)_minmax(300px,0.75fr)]">
          <div class="card p-4 sm:p-5">
            <div class="flex min-w-0 items-start justify-between gap-3">
              <div><p class="text-sm text-foreground-subtle">可结算佣金</p><p class="mt-1 text-2xl font-semibold tabular-nums text-success-foreground">{{ money(overview.agent.available_cny) }}</p></div>
              <div class="inline-flex rounded-control border border-outline bg-surface-subtle p-1" role="group" aria-label="结算方式">
                <button type="button" class="btn btn-sm" :class="mode === 'withdraw' ? 'btn-primary' : 'btn-ghost'" @click="mode = 'withdraw'">提现</button>
                <button type="button" class="btn btn-sm" :class="mode === 'convert' ? 'btn-primary' : 'btn-ghost'" @click="mode = 'convert'">转余额</button>
              </div>
            </div>

            <div class="mt-5">
              <div class="flex items-end justify-between gap-3"><label for="settlement-amount" class="input-label">结算金额</label><button type="button" class="text-sm font-medium text-brand" :disabled="availableAmount <= 0" @click="useMaximum">全部可用</button></div>
              <div class="relative"><input id="settlement-amount" v-model="amount" class="input pr-14" type="number" min="0.01" step="0.01" placeholder="输入人民币金额" :disabled="settling || overview.agent.status !== 'active'" /><span class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-foreground-muted">CNY</span></div>
              <p v-if="validationMessage" class="mt-2 text-sm text-danger-foreground">{{ validationMessage }}</p>
            </div>

            <div class="mt-4 rounded-panel border border-outline bg-surface-subtle p-4">
              <template v-if="mode === 'withdraw'">
                <div class="flex items-center justify-between text-sm"><span class="text-foreground-subtle">申请金额</span><strong>{{ money(validAmount) }}</strong></div>
                <div class="mt-2 flex items-center justify-between text-sm"><span class="text-foreground-subtle">手续费</span><span>-{{ money(withdrawalFee) }}</span></div>
                <div class="mt-3 flex items-center justify-between border-t border-outline pt-3"><span class="font-medium">预计到账</span><strong class="text-lg">{{ money(withdrawalPayout) }}</strong></div>
              </template>
              <template v-else>
                <div class="flex items-center justify-between text-sm"><span class="text-foreground-subtle">扣除佣金</span><strong>{{ money(validAmount) }}</strong></div>
                <div class="mt-3 flex items-center justify-between border-t border-outline pt-3"><span class="font-medium">转入平台余额</span><strong class="text-lg">${{ platformCredit.toFixed(2) }}</strong></div>
                <p class="mt-2 text-xs text-foreground-muted">平台换算：¥{{ Number(rules.cny_per_platform_usd).toFixed(2) }} = $1.00 平台余额</p>
              </template>
            </div>

            <button class="btn btn-primary mt-4 w-full" :disabled="!canSubmit" @click="reviewSettlement"><Icon :name="mode === 'withdraw' ? 'creditCard' : 'dollar'" size="sm" />{{ mode === 'withdraw' ? '确认提现信息' : '确认转入余额' }}</button>
          </div>

          <div class="space-y-4">
            <div class="card p-4 sm:p-5">
              <div class="flex items-start justify-between gap-3"><div><h2 class="font-semibold">支付宝收款账户</h2><p class="mt-1 text-sm text-foreground-subtle">提现申请将支付到此账户</p></div><button class="btn btn-secondary btn-sm" @click="openPayoutDialog"><Icon name="edit" size="sm" />{{ hasPayoutAccount ? '编辑' : '设置' }}</button></div>
              <div v-if="hasPayoutAccount" class="mt-4 rounded-control bg-surface-subtle px-3 py-3"><p class="font-medium">{{ payout.alipay_name }}</p><p class="mt-1 text-sm text-foreground-subtle">{{ payout.alipay_account }}</p></div>
              <p v-else class="mt-4 text-sm text-warning-foreground">申请提现前需要先设置收款账户。</p>
            </div>

            <div class="card p-4 sm:p-5">
              <h2 class="font-semibold">结算规则</h2>
              <dl class="mt-4 space-y-3 text-sm">
                <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">单笔范围</dt><dd class="text-right">{{ money(rules.minimum_withdrawal_cny) }} - {{ money(rules.maximum_withdrawal_cny) }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">手续费</dt><dd class="text-right">{{ feeRuleText }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">每日限额</dt><dd class="text-right">{{ limitText(rules.daily_withdrawal_limit_cny) }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">每月限额</dt><dd class="text-right">{{ limitText(rules.monthly_withdrawal_limit_cny) }}</dd></div>
                <div class="flex justify-between gap-3"><dt class="text-foreground-subtle">佣金冻结期</dt><dd class="text-right">{{ rules.freeze_hours }} 小时</dd></div>
              </dl>
            </div>
          </div>
        </section>

        <section class="card overflow-hidden">
          <div class="flex items-center justify-between gap-3 border-b border-outline px-4 py-3"><div><h2 class="font-semibold">提现记录</h2><p class="mt-0.5 text-xs text-foreground-subtle">仅记录支付宝提现，余额转换成功后立即计入平台余额</p></div><select v-model="recordStatus" class="input w-28" aria-label="提现记录状态"><option value="">全部状态</option><option value="pending">待审核</option><option value="approved">待付款</option><option value="paid">已付款</option><option value="rejected">已拒绝</option><option value="failed">付款失败</option></select></div>
          <DataTable :columns="recordColumns" :data="items" :loading="recordsLoading" row-key="id">
            <template #cell-amount="{ row }"><div><p class="font-semibold">申请 {{ money(row.amount_cny) }}</p><p class="mt-0.5 text-xs text-foreground-subtle">手续费 {{ money(row.fee_cny) }}</p></div></template>
            <template #cell-payout="{ row }"><span class="font-semibold">{{ money(row.payout_cny) }}</span></template>
            <template #cell-status="{ row }"><span class="badge" :class="withdrawalStatusClass(row.status)">{{ withdrawalStatusText(row.status) }}</span></template>
            <template #cell-created_at="{ row }"><time class="whitespace-nowrap text-sm text-foreground-muted">{{ new Date(row.created_at).toLocaleString() }}</time></template>
            <template #mobile-card="{ row }"><div class="space-y-3"><div class="flex items-start justify-between gap-3"><div><p class="font-medium">申请 {{ money(row.amount_cny) }}</p><p class="mt-0.5 text-xs text-foreground-subtle">{{ new Date(row.created_at).toLocaleString() }}</p></div><span class="badge shrink-0" :class="withdrawalStatusClass(row.status)">{{ withdrawalStatusText(row.status) }}</span></div><dl class="grid grid-cols-2 gap-3 border-t border-outline pt-3 text-xs"><div><dt class="text-foreground-subtle">手续费</dt><dd class="mt-1 font-semibold">{{ money(row.fee_cny) }}</dd></div><div class="text-right"><dt class="text-foreground-subtle">到账金额</dt><dd class="mt-1 font-semibold">{{ money(row.payout_cny) }}</dd></div></dl></div></template>
            <template #empty><div class="py-10 text-center"><p class="font-medium">暂无提现记录</p><p class="mt-1 text-sm text-foreground-subtle">提交提现申请后可在此跟踪处理进度</p></div></template>
          </DataTable>
          <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="changePage" @update:pageSize="changePageSize" />
        </section>
      </template>
    </div>

    <BaseDialog :show="payoutDialog" title="设置支付宝收款账户" width="normal" @close="payoutDialog = false">
      <form id="payout-account-form" class="space-y-4" @submit.prevent="savePayout"><div><label for="payout-name" class="input-label">真实姓名</label><input id="payout-name" v-model="draftPayout.alipay_name" class="input" autocomplete="name" placeholder="支付宝实名认证姓名" /></div><div><label for="payout-account" class="input-label">支付宝账号</label><input id="payout-account" v-model="draftPayout.alipay_account" class="input" autocomplete="off" placeholder="手机号或邮箱账号" /></div></form>
      <template #footer><div class="flex justify-end gap-2"><button class="btn btn-secondary" type="button" @click="payoutDialog = false">取消</button><button class="btn btn-primary" type="submit" form="payout-account-form" :disabled="savingPayout || !draftPayout.alipay_name.trim() || !draftPayout.alipay_account.trim()"><Icon name="check" size="sm" />{{ savingPayout ? '保存中...' : '保存账户' }}</button></div></template>
    </BaseDialog>

    <ConfirmDialog :show="confirmDialog" :title="mode === 'withdraw' ? '确认提现申请' : '确认转入平台余额'" :message="confirmationMessage" :confirm-text="mode === 'withdraw' ? '提交提现' : '确认转换'" :danger="mode === 'withdraw'" @confirm="submitSettlement" @cancel="confirmDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DistributionNav from '@/components/distribution/DistributionNav.vue'
import { convertToBalance, getDistributionOverview, getPayoutAccount, getSettlementRules, listMyWithdrawals, requestWithdrawal, updatePayoutAccount, type DistributionOverview, type DistributionPayoutAccount, type DistributionSettlementRules, type DistributionWithdrawal } from '@/api/distribution'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { calculatePlatformCredit, calculateWithdrawalPreview } from '@/utils/distributionSettlement'

const app = useAppStore()
const overview = ref<DistributionOverview | null>(null)
const rules = ref<DistributionSettlementRules | null>(null)
const payout = ref<DistributionPayoutAccount>({ alipay_name: '', alipay_account: '' })
const items = ref<DistributionWithdrawal[]>([])
const loading = ref(false)
const recordsLoading = ref(false)
const settling = ref(false)
const mode = ref<'withdraw' | 'convert'>('withdraw')
const amount = ref('')
const recordStatus = ref('')
const pagination = ref({ page: 1, page_size: 20, total: 0, pages: 1 })
const payoutDialog = ref(false)
const savingPayout = ref(false)
const draftPayout = ref<DistributionPayoutAccount>({ alipay_name: '', alipay_account: '' })
const confirmDialog = ref(false)
const recordColumns: Column[] = [{ key: 'amount', label: '申请金额' }, { key: 'payout', label: '到账金额' }, { key: 'status', label: '状态' }, { key: 'created_at', label: '申请时间' }]
const availableAmount = computed(() => Number(overview.value?.agent.available_cny || 0))
const validAmount = computed(() => Math.max(0, Number(amount.value) || 0))
const hasPayoutAccount = computed(() => Boolean(payout.value.alipay_name && payout.value.alipay_account))
const withdrawalPreview = computed(() => calculateWithdrawalPreview(validAmount.value, rules.value?.withdrawal_fee_rate_bps || 0, Number(rules.value?.withdrawal_fee_fixed_cny || 0)))
const withdrawalFee = computed(() => withdrawalPreview.value.fee)
const withdrawalPayout = computed(() => withdrawalPreview.value.payout)
const platformCredit = computed(() => calculatePlatformCredit(validAmount.value, Number(rules.value?.cny_per_platform_usd || 1)))
const validationMessage = computed(() => { if (!amount.value) return ''; if (validAmount.value <= 0) return '请输入大于 0 的金额'; if (validAmount.value > availableAmount.value) return '结算金额不能超过可用佣金'; if (mode.value === 'withdraw' && rules.value) { if (validAmount.value < Number(rules.value.minimum_withdrawal_cny)) return `最低提现金额为 ${money(rules.value.minimum_withdrawal_cny)}`; if (validAmount.value > Number(rules.value.maximum_withdrawal_cny)) return `单笔最高提现金额为 ${money(rules.value.maximum_withdrawal_cny)}`; if (withdrawalPayout.value <= 0) return '手续费不能高于提现金额' } return '' })
const canSubmit = computed(() => Boolean(overview.value && overview.value.agent.status === 'active' && validAmount.value > 0 && !validationMessage.value && !settling.value && (mode.value === 'convert' || (rules.value?.withdrawal_enabled && hasPayoutAccount.value))))
const feeRuleText = computed(() => { if (!rules.value) return '-'; const parts = []; if (rules.value.withdrawal_fee_rate_bps > 0) parts.push(`${(rules.value.withdrawal_fee_rate_bps / 100).toFixed(2)}%`); if (Number(rules.value.withdrawal_fee_fixed_cny) > 0) parts.push(`固定 ${money(rules.value.withdrawal_fee_fixed_cny)}`); return parts.length ? parts.join(' + ') : '免手续费' })
const confirmationMessage = computed(() => mode.value === 'withdraw' ? `申请提现 ${money(validAmount.value)}，手续费 ${money(withdrawalFee.value)}，预计到账 ${money(withdrawalPayout.value)} 至 ${payout.value.alipay_account}。` : `将 ${money(validAmount.value)} 佣金转换为 $${platformCredit.value.toFixed(2)} 平台余额，转换后不可撤销。`)
const money = (value: string | number) => `¥${Number(value || 0).toFixed(2)}`
const limitText = (value: string) => Number(value) > 0 ? money(value) : '不限'
const withdrawalStatusText = (value: string) => ({ pending: '待审核', approved: '待付款', paying: '付款中', rejected: '已拒绝', paid: '已付款', failed: '付款失败', cancelled: '已取消' })[value] || value
const withdrawalStatusClass = (value: string) => value === 'paid' ? 'badge-success' : value === 'pending' || value === 'approved' || value === 'paying' ? 'badge-warning' : 'badge-gray'

async function loadRecords() { recordsLoading.value = true; try { const result = await listMyWithdrawals({ page: pagination.value.page, page_size: pagination.value.page_size, status: recordStatus.value || undefined }); items.value = result.items; pagination.value = { page: result.page, page_size: result.page_size, total: result.total, pages: result.pages } } catch (error) { app.showError(extractApiErrorMessage(error, '加载提现记录失败')) } finally { recordsLoading.value = false } }
async function load() { loading.value = true; try { [overview.value, rules.value, payout.value] = await Promise.all([getDistributionOverview(), getSettlementRules(), getPayoutAccount()]); await loadRecords() } catch (error) { app.showError(extractApiErrorMessage(error, '加载结算信息失败')) } finally { loading.value = false } }
function useMaximum() { if (!rules.value) return; amount.value = String(mode.value === 'withdraw' ? Math.min(availableAmount.value, Number(rules.value.maximum_withdrawal_cny)) : availableAmount.value) }
function openPayoutDialog() { draftPayout.value = { ...payout.value }; payoutDialog.value = true }
async function savePayout() { savingPayout.value = true; try { await updatePayoutAccount({ alipay_name: draftPayout.value.alipay_name.trim(), alipay_account: draftPayout.value.alipay_account.trim() }); payout.value = { ...draftPayout.value }; payoutDialog.value = false; app.showSuccess('收款账户已保存') } catch (error) { app.showError(extractApiErrorMessage(error, '保存收款账户失败')) } finally { savingPayout.value = false } }
function reviewSettlement() { if (canSubmit.value) confirmDialog.value = true }
async function submitSettlement() { if (!canSubmit.value) return; settling.value = true; try { if (mode.value === 'withdraw') { await requestWithdrawal(validAmount.value.toFixed(2)); app.showSuccess('提现申请已提交') } else { const result = await convertToBalance(validAmount.value.toFixed(2)); app.showSuccess(`已转入平台余额 $${Number(result.credited_platform_usd).toFixed(2)}`) } confirmDialog.value = false; amount.value = ''; await load() } catch (error) { app.showError(extractApiErrorMessage(error, mode.value === 'withdraw' ? '提现申请失败' : '转换失败')) } finally { settling.value = false } }
function changePage(page: number) { pagination.value.page = page; void loadRecords() }
function changePageSize(pageSize: number) { pagination.value.page = 1; pagination.value.page_size = pageSize; void loadRecords() }
watch(recordStatus, () => { pagination.value.page = 1; void loadRecords() })
onMounted(load)
</script>
