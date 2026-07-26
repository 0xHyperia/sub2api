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

  it('marks drafts and uses application dialogs for destructive or abandoned changes', () => {
    expect(source).toContain('有未保存修改')
    expect(source).toContain('onBeforeRouteLeave')
    expect(source).toContain('<ConfirmDialog')
    expect(source).not.toContain('window.confirm')
  })

  it('provides recoverable loading, client validation and browser-leave protection', () => {
    expect(source).toContain('分销设置加载失败')
    expect(source).toContain('@click="load"')
    expect(source).toContain('validationError')
    expect(source).toContain("window.addEventListener('beforeunload'")
    expect(source).toContain("window.removeEventListener('beforeunload'")
  })

  it('does not overwrite edits made while a save request is running', () => {
    expect(source).toContain('submittedFingerprint')
    expect(source).toContain('hasNewerEdits')
    expect(source).toContain('保存期间的新修改仍待保存')
  })

  it('keeps mobile controls touchable and dirty save actions visible', () => {
    expect(source).toContain('.input{min-height:44px}')
    expect(source).toContain("dirty ? 'sticky bottom-0")
    expect(source).toContain('min-h-11 w-full')
  })
})
