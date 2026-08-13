import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { parse } from "vue/compiler-sfc";

const view = (name: string) => {
  const filename = resolve(process.cwd(), "src/views/admin/distribution", name);
  const source = readFileSync(filename, "utf8");
  const { descriptor, errors } = parse(source, { filename });
  expect(errors).toEqual([]);
  return { source, template: descriptor.template?.content || "" };
};

describe("enterprise distribution administration contracts", () => {
  it("provides an operating and audit overview", () => {
    const { template, source } = view("AdminDistributionOverviewView.vue");
    expect(template).toContain("admin.distribution.analytics.liability");
    expect(template).toContain("admin.distribution.analytics.operations");
    expect(source).toContain("admin.distribution.analytics.overdueReview");
    expect(template).not.toContain("打款或已付款但缺少凭证");
    expect(source).toContain("getOverview");
    expect(template).toContain("admin.distribution.analytics.maturity");
    expect(source).toContain("getMaturityStatus");
  });

  it("provides a dedicated reconciliation workbench with actionable anomaly records", () => {
    const { template, source } = view("AdminDistributionAnomaliesView.vue");
    expect(template).toContain("集中处理佣金、提现、代理归属和资金负债异常");
    expect(template).toContain("提现审核超时");
    expect(template).not.toContain("缺少打款凭证");
    expect(template).toContain("代理钱包负债");
    expect(template).toContain("lg:hidden");
    expect(source).toContain("listAnomalies");
    expect(source).toContain("targetFor");
  });

  it("supports audited rate editing and server-side financial sorting for agents", () => {
    const { template, source } = view("AdminDistributionAgentsView.vue");
    expect(template).toContain("admin.distribution.agents.editRate");
    expect(template).toContain("admin.distribution.agents.inheritDefault");
    expect(template).toContain("admin.distribution.agents.changeReason");
    expect(template).toContain(':server-side-sort="true"');
    expect(source).toContain("updateAgentRate");
    expect(source).toContain("sort_by: sortBy.value");
    expect(source).toMatch(/key:\s*["']customer_paid["']/);
    expect(source).toMatch(/key:\s*["']total_earned["']/);
    expect(source).toMatch(/class:\s*["']text-right["']/);
    expect(source).toMatch(/class:\s*["']text-center["']/);
    expect(source).toMatch(/class:\s*["']w-16 text-center["']/);
    expect(template).toContain("admin.distribution.agents.recruitment");
    expect(template).toContain("admin.distribution.agents.grantRecruitment");
    expect(source).toContain("updateAgentRecruitmentPermission");
  });

  it("shows sortable customer payment, commission and refund metrics", () => {
    const { template, source } = view("AdminDistributionCustomersView.vue");
    expect(source).toContain("label: '累计实付'");
    expect(template).toContain("贡献佣金");
    expect(source).toContain("label: '退款金额'");
    expect(template).toContain(':server-side-sort="true"');
    expect(source).toContain("sort_by: sortBy.value");
  });

  it("uses paginated commission ledger and withdrawal workflow tables", () => {
    const commission = view("AdminDistributionCommissionsView.vue");
    expect(commission.template).toContain("逐笔核对订单实付");
    expect(commission.template).toContain("<Pagination");
    expect(commission.template).toContain(':server-side-sort="true"');
    expect(commission.template).toContain("计佣基数");
    expect(commission.template).toContain("冲正金额");
    expect(commission.template).toContain("已到账");
    expect(commission.template).not.toMatch(/已提现|已转余额/);
    expect(commission.template).toContain('aria-label="导出佣金台账"');

    const withdrawal = view("AdminDistributionWithdrawalsView.vue");
    expect(withdrawal.template).toContain("打款凭证（可选）");
    expect(withdrawal.template).toContain("处理记录");
    expect(withdrawal.template).toContain("开始打款");
    expect(withdrawal.template).toContain("确认已付款");
    expect(withdrawal.template).toContain("选择图片或 PDF");
    expect(withdrawal.template).toContain("<Pagination");
    expect(withdrawal.source).toContain("uploadWithdrawalEvidence");
    expect(withdrawal.source).toContain('nextStatus.value !== "paid"');
    expect(withdrawal.template).not.toContain(
      ':disabled="detail.attachments.length === 0"',
    );
    expect(withdrawal.template).toContain("批量审核只支持全部处于待审核状态");
    expect(withdrawal.template).toContain("整个批次都会失败");
    expect(withdrawal.source).toContain("batchReviewWithdrawals");
    expect(withdrawal.template).toContain('aria-label="导出提现记录"');
    expect(withdrawal.template).toContain("双人复核已启用");
    expect(withdrawal.source).toContain("cannotStartPayment");
  });

  it("exports full filtered agent and customer ledgers", () => {
    const agents = view("AdminDistributionAgentsView.vue");
    const customers = view("AdminDistributionCustomersView.vue");
    expect(agents.template).toContain("admin.distribution.agents.exportFiltered");
    expect(agents.source).toContain("exportAgents");
    expect(customers.template).toContain('aria-label="导出代理客户"');
    expect(customers.source).toContain("exportCustomers");
  });

  it("keeps all advanced filters accessible on mobile", () => {
    for (const name of [
      "AdminDistributionCommissionsView.vue",
      "AdminDistributionWithdrawalsView.vue",
    ]) {
      const { template } = view(name);
      expect(template, name).toContain("lg:hidden");
      expect(template, name).toContain('name="filter"');
      expect(template, name).toContain('type="date"');
    }
  });

  it("uses scalable promotion analysis with independent snapshot failures", () => {
    const { template, source } = view("AdminDistributionPromotionView.vue");
    expect(source).toContain("lookupAgents");
    expect(source).not.toContain("listAgents");
    expect(source).toContain("snapshotSequence");
    expect(source).toMatch(/await Promise\.allSettled\(\[\s*getPromotionAnalytics[\s\S]*listPromotionVisits/);
    expect(source).toContain("analyticsError");
    expect(source).toContain("visitsError");
    expect(source).toContain("if (sequence !== snapshotSequence) return");
    expect(template).toContain("当前筛选条件");
    expect(template).toContain("香港时间");
    expect(template).toContain("转化耗时");
  });
});
