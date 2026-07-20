import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(resolve(process.cwd(), 'src/components/admin/DistributionSettingsCard.vue'), 'utf8')

describe('DistributionSettingsCard', () => {
  it('presents rates as percentages while preserving basis-point API values', () => {
    expect(source).not.toContain('>BP<')
    expect(source).toContain('>%</span>')
    expect(source).toContain('const toBPS')
    expect(source).toContain('l1_default_rate_bps: toBPS(rates.l1Default)')
  })

  it('explains that the USD rate only converts USD orders for commission', () => {
    expect(source).toContain('美元订单计佣汇率')
    expect(source).toContain('人民币订单固定按 1:1')
  })

  it('offers maker-checker control for withdrawals', () => {
    expect(source).toContain('提现双人复核')
    expect(source).toContain('审核人与打款操作人必须不同')
    expect(source).toContain('settings.withdrawal_dual_approval_enabled')
  })
})
