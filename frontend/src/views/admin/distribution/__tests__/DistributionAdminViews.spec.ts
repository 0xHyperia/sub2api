import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'
import { parse } from 'vue/compiler-sfc'

const viewsRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')

function readView(filename: string) {
  const path = resolve(viewsRoot, filename)
  const source = readFileSync(path, 'utf8')
  const { descriptor, errors } = parse(source, { filename: path })
  expect(errors).toEqual([])
  return { source, template: descriptor.template?.content ?? '' }
}

describe('distribution admin view contracts', () => {
  it('adds agents through an email or username picker instead of an ID input', () => {
    const { template } = readView('AdminDistributionAgentsView.vue')

    expect(template).toContain('title="添加一级代理"')
    expect(template).toContain('placeholder="输入用户邮箱或用户名"')
    expect(template).toContain('<RemoteEntityCombobox')
    expect(template).not.toMatch(/用户\s*ID|user[-_]id/i)
  })

  it('offers a reasoned reactivation action for revoked agents', () => {
    const { source, template } = readView('AdminDistributionAgentsView.vue')

    expect(template).toMatch(
      /managedAgent\.status === ["']revoked["'] \? ["']重新启用["'] : ["']恢复代理["']/
    )
    expect(source).toMatch(
      /pendingStatus\.value === ["']active["']\s*&&\s*managedAgent\.value\?\.status === ["']revoked["']/
    )
    expect(template).toContain('statusReasonRequired && !statusReason.trim()')
    expect(source).not.toContain('撤销后该代理将无法恢复')
  })

  it('keeps customer correction in a row-triggered dialog', () => {
    const { template } = readView('AdminDistributionCustomersView.vue')

    expect(template).toContain('<BaseDialog :show="correctionDialog" title="调整客户归属"')
    expect(template.match(/@click(?:\.stop)?="openCorrectionDialog\(row\)"/g)).toHaveLength(2)
    expect(template).toContain('placeholder="输入代理邮箱、用户名或推广码"')
    expect(template).not.toMatch(/客户\s*ID|代理\s*ID/)
  })
})
