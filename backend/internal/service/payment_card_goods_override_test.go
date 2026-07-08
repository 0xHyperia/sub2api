package service

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment/provider"
)

func TestParseCardGoodsOverrides(t *testing.T) {
	overrides, err := parseCardGoodsOverrides(`{"g1":{"title":"标题","description":"描述","badge":"热门","tags":["自动开通"]}}`)
	if err != nil {
		t.Fatalf("parseCardGoodsOverrides returned error: %v", err)
	}
	item := overrides["g1"]
	if item.Title != "标题" || item.Description != "描述" {
		t.Fatalf("override text mismatch: %+v", item)
	}
	if item.Badge != "热门" {
		t.Fatalf("override badge mismatch: %q", item.Badge)
	}
	if len(item.Tags) != 1 || item.Tags[0] != "自动开通" {
		t.Fatalf("override tags mismatch: %+v", item.Tags)
	}
}

func TestNormalizeCardGoodsOverride(t *testing.T) {
	override := normalizeCardGoodsOverride(UpdateCardGoodsOverrideRequest{
		Title:       "  标题  ",
		Description: "  描述  ",
		Badge:       " 热门 ",
		Tags:        []string{"自动开通", " 自动开通 ", "", "30天有效"},
	})
	if override.Title != "标题" || override.Description != "描述" {
		t.Fatalf("normalized text mismatch: %+v", override)
	}
	if override.Badge != "热门" {
		t.Fatalf("normalized badge mismatch: %q", override.Badge)
	}
	if len(override.Tags) != 2 || override.Tags[0] != "自动开通" || override.Tags[1] != "30天有效" {
		t.Fatalf("normalized tags mismatch: %+v", override.Tags)
	}
}

func TestApplyCardGoodsOverrides(t *testing.T) {
	svc := &PaymentService{configService: &PaymentConfigService{}}
	inst := &dbent.PaymentProviderInstance{
		ID:     1,
		Config: `{"goodsOverrides":"{\"g1\":{\"title\":\"自定义标题\",\"description\":\"自定义描述\",\"badge\":\"热门\",\"tags\":[\"自动开通\"]}}"}`,
	}
	goods := []provider.LdxpGoods{
		{GoodsKey: "g1", Name: "原始标题", Description: "原始描述", Price: 10, ReferencePrice: 20},
		{GoodsKey: "g2", Name: "不展示参考价", Price: 10, ReferencePrice: 5},
	}
	out := svc.applyCardGoodsOverrides(inst, goods)
	if out[0].DisplayTitle != "自定义标题" || out[0].DisplayDescription != "自定义描述" {
		t.Fatalf("display fields mismatch: %+v", out[0])
	}
	if out[0].Badge != "热门" {
		t.Fatalf("badge mismatch: %q", out[0].Badge)
	}
	if len(out[0].Tags) != 1 || out[0].Tags[0] != "自动开通" {
		t.Fatalf("tags mismatch: %+v", out[0].Tags)
	}
	if out[0].ReferencePrice != 20 {
		t.Fatalf("reference price = %v, want 20", out[0].ReferencePrice)
	}
	if out[1].ReferencePrice != 5 {
		t.Fatalf("provider reference price should pass through, got %v", out[1].ReferencePrice)
	}
}
