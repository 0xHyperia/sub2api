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

    expect(template).toContain('admin.distribution.agents.addL1')
    expect(template).toContain('admin.distribution.agents.userPlaceholder')
    expect(template).toContain('<RemoteEntityCombobox')
    expect(template).not.toMatch(/用户\s*ID|user[-_]id/i)
  })

  it('offers a reasoned reactivation action for revoked agents', () => {
    const { source, template } = readView('AdminDistributionAgentsView.vue')

    expect(template).toContain('admin.distribution.agents.reactivateAgent')
    expect(template).toContain('admin.distribution.agents.restoreAgent')
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

  it('presents distribution operations as period analytics instead of lifetime balances only', () => {
    const overview = readView('AdminDistributionOverviewView.vue')
    const agents = readView('AdminDistributionAgentsView.vue')

    expect(overview.template).toContain("admin.distribution.analytics.channelTrend")
    expect(overview.source).toContain('common.distributionAnalytics.directBusiness')
    expect(overview.source).toContain('common.distributionAnalytics.teamBusiness')
    expect(overview.template).toContain('admin.distribution.analytics.topAgents')
    expect(overview.template).toContain('<DistributionAnalyticsRange')
    expect(overview.source).toContain('analyticsRangeParams(range.value)')
    expect(overview.source).toContain('commissionRate')
    expect(agents.source).toContain("admin.distribution.agents.colPeriodCustomers")
    expect(agents.source).toContain("admin.distribution.agents.tabCustomers")
    expect(agents.source).toContain("admin.distribution.agents.tabCommission")
    expect(agents.template).toContain('<DistributionBusinessChart')
    expect(agents.template.match(/<DistributionAnalyticsRange/g)).toHaveLength(2)
    expect(agents.source).not.toContain('v-if="false"')
  })

  it('keeps each L1 row on a consistent direct plus team period scope', () => {
    const agents = readView('AdminDistributionAgentsView.vue')

    expect(agents.source).toContain('periodNewCustomers(row)')
    expect(agents.source).toContain('periodPayingCustomers(row)')
    expect(agents.source).toContain('row.team_customer_count')
    expect(agents.source).toContain('row.team_paying_customers')
    expect(agents.template).toContain('admin.distribution.agents.directTeamCounts')
  })
})
