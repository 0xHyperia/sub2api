import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import { parse } from "vue/compiler-sfc";

const view = (name: string) => {
  const filename = resolve(process.cwd(), "src/views/user", name);
  const source = readFileSync(filename, "utf8");
  const { descriptor, errors } = parse(source, { filename });
  expect(errors).toEqual([]);
  return { source, template: descriptor.template?.content || "" };
};

describe("distribution user view contracts", () => {
  it("never asks agents to type or work with internal user IDs or BP", () => {
    const sources = [
      "DistributionCustomersView.vue",
      "DistributionCommissionsView.vue",
      "DistributionTeamView.vue",
    ]
      .map((name) => view(name).template)
      .join("\n");

    expect(sources).not.toMatch(/用户\s*ID|用户\s*#|返佣比例\s*BP/i);
  });

  it("adds team agents only through an exact email and granted permission", () => {
    const { template, source } = view("DistributionTeamView.vue");
    expect(template).toContain("common.distributionCenter.addChild");
    expect(template).toContain('type="email"');
    expect(template).toContain("common.distributionCenter.emailPlaceholder");
    expect(template).toContain("common.distributionCenter.exactEmailHint");
    expect(template).toContain("overview?.agent.can_recruit_subagents");
    expect(template).not.toContain("<RemoteEntityCombobox");
    expect(source).not.toContain("lookupTeamCandidates");
    expect(template).toContain("absolute right-3 top-1/2");
    expect(template).not.toMatch(/一级代理|二级代理/);
  });

  it("uses server pagination for every growing business list", () => {
    for (const name of [
      "DistributionCustomersView.vue",
      "DistributionCommissionsView.vue",
      "DistributionTeamView.vue",
      "DistributionWithdrawalsView.vue",
    ]) {
      const { template, source } = view(name);
      expect(template, name).toContain("<Pagination");
      expect(source, name).toContain("page_size: pagination.value.page_size");
    }
  });

  it("uses commission status only for the arrival lifecycle", () => {
    const { template, source } = view("DistributionCommissionsView.vue");
    expect(template).toContain("佣金到账状态");
    expect(template).toContain("已到账");
    expect(`${template}\n${source}`).not.toMatch(/已提现|已转余额|结算中/);
  });

  it("previews and confirms both settlement modes", () => {
    const { template, source } = view("DistributionWithdrawalsView.vue");
    expect(template).toContain('aria-label="结算方式"');
    expect(template).toContain("预计到账");
    expect(template).toContain("转入平台余额");
    expect(template).toContain("absolute right-3 top-1/2");
    expect(template).toContain("<ConfirmDialog");
    expect(source).toContain("calculateWithdrawalPreview");
  });

  it("provides a dedicated promotion page with QR download", () => {
    const { template, source } = view("DistributionPromotionView.vue");
    expect(template).toContain("推广二维码");
    expect(template).toContain("下载二维码");
    expect(source).toContain("QRCode.toCanvas");
  });

  it("keeps historical promotion data visible and degrades independent requests", () => {
    const { template, source } = view("DistributionPromotionView.vue");
    expect(template).toContain("历史数据保留");
    expect(template).toContain("statsError");
    expect(template).toContain("visitsError");
    expect(template).not.toContain('v-else-if="trackingEnabled"');
    expect(source).toContain("async function loadStats");
    expect(source).toContain("async function loadVisits");
    expect(source).toContain("statsSequence");
    expect(source).toContain("visitsSequence");
    expect(source).toContain("sequence !== statsSequence");
  });

  it("aggregates the 90 day promotion trend by week", () => {
    const { template, source } = view("DistributionPromotionView.vue");
    expect(template).toContain("按周");
    expect(source).toContain("days.value !== 90");
    expect(source).toContain("grouped.set");
    expect(source).toContain("parseLocalDate");
    expect(source).toContain("formatLocalDate");
    expect(source).not.toContain("toISOString().slice(0, 10)");
    expect(source).not.toContain("slice(days.value > 30 ? -30");
  });

  it("lists promotion trend dates from newest to oldest", () => {
    const { source } = view("DistributionPromotionView.vue");
    expect(source).toContain("right.date.localeCompare(left.date)");
  });

  it("does not truncate money or promotion links on small screens", () => {
    const { template } = view("DistributionPromotionView.vue");
    expect(template).toContain("break-all");
    expect(template).toContain("whitespace-nowrap");
    expect(template).toContain("团队付费客户");
    expect(template).toContain("团队累计实付");
    expect(template).not.toContain("累计实付</dt><dd class=\"mt-1 truncate");
  });

  it("reports clipboard failures instead of silently rejecting", () => {
    const { source } = view("DistributionPromotionView.vue");
    expect(source).toContain("navigator.clipboard.writeText");
    expect(source).toContain("复制失败，请手动选择内容复制");
  });

  it("keeps failure, disabled assets and conversion semantics explicit", () => {
    const { template, source } = view("DistributionPromotionView.vue");
    expect(template).toContain("overviewError");
    expect(template).toContain('v-if="linkEnabled"');
    expect(template).toContain('role="img"');
    expect(source).toContain("有效访问");
    expect(source).toContain("访客转化率");
    expect(source).toContain("已转化访客");
    expect(source).toContain("已转化访客 ÷ 独立访客");
    expect(template).toContain("注册结果");
    expect(template).toContain("携推广码注册但未匹配到有效访问");
    expect(template).toContain("source.conversions");
    expect(template).not.toContain("source.registrations }} 注册");
    expect(template).toContain("数据质量");
    expect(source).toContain("timeZone: 'Asia/Hong_Kong'");
  });

  it("isolates QR failures, explains attribution and progressively reveals visits", () => {
    const { template, source } = view("DistributionPromotionView.vue");
    expect(template).toContain("二维码生成失败");
    expect(source).toContain("qrError.value = true");
    expect(source).toContain("void renderQr()");
    expect(template).toContain("归因规则");
    expect(source).toContain("attributionPolicyText");
    expect(template).toContain("筛选采用访问批次口径");
    expect(template).toContain("注册显示在获归因的推广访问日期");
    expect(template).toContain("默认展示最新 8 条");
    expect(template).toContain("查看更多访问");
    expect(source).toContain("async function loadMoreVisits");
    expect(template).toContain("bg-info");
    expect(template).toContain("bg-success");
  });

  it("separates direct and team operations for L1 agents", () => {
    const overview = view("DistributionView.vue");
    const team = view("DistributionTeamView.vue");

    expect(overview.template).toContain("common.distributionCenter.directTeamTitle");
    expect(overview.template).toContain("common.distributionCenter.childRanking");
    expect(overview.template).toContain("common.distributionCenter.settle");
    expect(overview.template).toContain("<DistributionBusinessChart");
    expect(overview.template).toContain("<DistributionAnalyticsRange");
    expect(overview.source).toContain("analyticsRangeParams(range.value)");
    expect(team.template).toContain("common.distributionCenter.teamBusiness");
    expect(team.template).toContain("common.distributionCenter.teamCustomerPaid");
    expect(team.template).toContain("<DistributionAnalyticsRange");
    expect(team.source).toContain("...analyticsRangeParams(range.value)");
  });
});
