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
    expect(template).toContain('title="添加下级代理"');
    expect(template).toContain('type="email"');
    expect(template).toContain('placeholder="输入完整邮箱"');
    expect(template).toContain("仅支持完整邮箱精确匹配");
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
});
