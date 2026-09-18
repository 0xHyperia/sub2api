//go:build unit

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserAvailableChannel_Unauthenticated401(t *testing.T) {
	// 没有 AuthSubject 注入时，handler 应返回 401 且不触达 service 依赖。
	gin.SetMode(gin.TestMode)
	h := &AvailableChannelHandler{} // nil services — 401 路径不会调用它们
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/channels/available", nil)

	h.List(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestModelMarketplace_Unauthenticated401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AvailableChannelHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/models/marketplace", nil)

	h.ListMarketplace(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestModelShowcase_DisabledReturnsEmptyWithoutAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &AvailableChannelHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/models/showcase", nil)

	h.ListShowcase(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"code":0,"message":"success","data":[]}`, w.Body.String())
}

func TestModelShowcase_FieldWhitelistOmitsGroupIdentity(t *testing.T) {
	row := publicModelShowcasePlatform{
		Platform:   service.PlatformOpenAI,
		ModelCount: 1,
		Models: []publicModelShowcaseModel{{
			Name:           "gpt-test",
			Platform:       service.PlatformOpenAI,
			RateMultiplier: 0.8,
		}},
	}
	raw, err := json.Marshal(row)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "group")
	require.NotContains(t, string(raw), "exclusive")
	require.Contains(t, string(raw), `"rate_multiplier":0.8`)
}

func TestShowcasePlatformRankPrioritizesPrimaryProviders(t *testing.T) {
	require.Less(t, showcasePlatformRank(service.PlatformOpenAI), showcasePlatformRank(service.PlatformAnthropic))
	require.Less(t, showcasePlatformRank(service.PlatformAnthropic), showcasePlatformRank(service.PlatformGemini))
	require.Less(t, showcasePlatformRank(service.PlatformGemini), showcasePlatformRank("custom"))
}

func TestMarketplaceModelIDs_CustomGroupListTakesPriority(t *testing.T) {
	group := service.Group{
		Platform: service.PlatformOpenAI,
		ModelAllowlist: service.GroupModelAllowlist{
			Enabled: true,
			Models:  []string{"gpt-custom", " gpt-custom ", "gpt-other", "gpt-*"},
		},
	}
	accounts := []service.Account{{
		Platform:    service.PlatformOpenAI,
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-custom": "gpt-upstream", "gpt-other": "gpt-other"}},
	}}

	require.Equal(t, []string{"gpt-custom", "gpt-other"}, service.ModelCatalogModels(group, accounts))
}

func TestMarketplaceModelIDs_UsesOnlyExplicitAccountMappings(t *testing.T) {
	group := service.Group{Platform: service.PlatformOpenAI}
	accounts := []service.Account{{
		Platform:    service.PlatformOpenAI,
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-marketplace-only": "gpt-upstream"}},
	}}

	models := service.ModelCatalogModels(group, accounts)
	require.Equal(t, []string{"gpt-marketplace-only"}, models)
}

func TestMarketplaceModelIDs_CompositeCollectsConcreteAccountMappings(t *testing.T) {
	group := service.Group{Platform: service.PlatformComposite}
	accounts := []service.Account{
		{Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-public": "gpt-upstream"}}},
		{Platform: service.PlatformAnthropic, Credentials: map[string]any{"model_mapping": map[string]any{"claude-public": "claude-upstream"}}},
	}

	require.Equal(t, []string{"claude-public", "gpt-public"}, service.ModelCatalogModels(group, accounts))
}

func TestFilterUserVisibleGroups_IntersectionOnly(t *testing.T) {
	// 渠道挂在 {g1, g2, g3}，用户只允许 {g1, g3} —— 响应必须仅含 g1/g3。
	groups := []service.AvailableGroupRef{
		{ID: 1, Name: "g1", Platform: "anthropic"},
		{ID: 2, Name: "g2", Platform: "anthropic"},
		{ID: 3, Name: "g3", Platform: "openai"},
	}
	allowed := map[int64]struct{}{1: {}, 3: {}}

	visible := filterUserVisibleGroups(groups, allowed)
	require.Len(t, visible, 2)
	ids := []int64{visible[0].ID, visible[1].ID}
	require.ElementsMatch(t, []int64{1, 3}, ids)
}

func TestToUserSupportedModels_FiltersByAllowedPlatforms(t *testing.T) {
	// 用户可访问分组只覆盖 anthropic；anthropic 平台的模型保留，openai 模型被剔除。
	src := []service.SupportedModel{
		{Name: "claude-sonnet-4-6", Platform: "anthropic", Pricing: nil},
		{Name: "gpt-4o", Platform: "openai", Pricing: nil},
	}
	allowed := map[string]struct{}{"anthropic": {}}
	out := toUserSupportedModels(src, allowed)
	require.Len(t, out, 1)
	require.Equal(t, "claude-sonnet-4-6", out[0].Name)
}

func TestToUserSupportedModels_NilAllowedPlatformsKeepsAll(t *testing.T) {
	// 显式传 nil allowedPlatforms 表示不做过滤。
	src := []service.SupportedModel{
		{Name: "a", Platform: "anthropic"},
		{Name: "b", Platform: "openai"},
	}
	require.Len(t, toUserSupportedModels(src, nil), 2)
}

func TestUserAvailableChannel_FieldWhitelist(t *testing.T) {
	// 通过序列化 userAvailableChannel 结构体验证响应形状：
	// 只有 name / description / platforms；不含管理端字段。
	row := userAvailableChannel{
		Name:        "ch",
		Description: "d",
		Platforms: []userChannelPlatformSection{
			{
				Platform:        "anthropic",
				Groups:          []userAvailableGroup{{ID: 1, Name: "g1", Platform: "anthropic"}},
				SupportedModels: []userSupportedModel{},
			},
		},
	}
	raw, err := json.Marshal(row)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	for _, key := range []string{"id", "status", "billing_model_source", "restrict_models"} {
		_, exists := decoded[key]
		require.Falsef(t, exists, "user DTO must not expose %q", key)
	}
	for _, key := range []string{"name", "description", "platforms"} {
		_, exists := decoded[key]
		require.Truef(t, exists, "user DTO must expose %q", key)
	}

	// 验证 section 的字段（platform / groups / supported_models）。
	rawSection, err := json.Marshal(row.Platforms[0])
	require.NoError(t, err)
	var sectionDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawSection, &sectionDecoded))
	for _, key := range []string{"platform", "groups", "supported_models"} {
		_, exists := sectionDecoded[key]
		require.Truef(t, exists, "platform section must expose %q", key)
	}

	// Group DTO 暴露区分专属/公开、订阅类型、默认倍率和高峰倍率规则所需的字段，
	// 前端据此渲染 GroupBadge 并与 API 密钥页保持一致的视觉。
	rawGroup, err := json.Marshal(row.Platforms[0].Groups[0])
	require.NoError(t, err)
	var groupDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawGroup, &groupDecoded))
	for _, key := range []string{"id", "name", "platform", "subscription_type", "rate_multiplier", "peak_rate_enabled", "peak_start", "peak_end", "peak_rate_multiplier", "is_exclusive", "rpm_limit"} {
		_, exists := groupDecoded[key]
		require.Truef(t, exists, "group DTO must expose %q", key)
	}

	// pricing interval 白名单：不应暴露 id / sort_order。
	inputMultiplier := 2.0
	outputMultiplier := 1.5
	cacheWriteMultiplier := 2.0
	cacheReadMultiplier := 2.0
	pricing := toUserPricing(&service.ChannelModelPricing{
		BillingMode: service.BillingModeToken,
		TimePricing: &service.ChannelTimePricing{Timezone: "Asia/Shanghai", WeekdaysOnly: true, Periods: []service.ChannelTimePricingPeriod{
			{StartTime: "09:00:00", EndTime: "18:00:00", Multiplier: 2},
		}},
		Intervals: []service.PricingInterval{
			{
				ID: 7, MinTokens: 0, MaxTokens: nil, SortOrder: 3,
				InputMultiplier: &inputMultiplier, OutputMultiplier: &outputMultiplier,
				CacheWriteMultiplier: &cacheWriteMultiplier, CacheReadMultiplier: &cacheReadMultiplier,
			},
		},
	})
	require.NotNil(t, pricing)
	rawPricing, err := json.Marshal(pricing)
	require.NoError(t, err)
	var pricingDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawPricing, &pricingDecoded))
	require.Equal(t, map[string]any{
		"timezone": "Asia/Shanghai", "weekdays_only": true,
		"periods": []any{map[string]any{"start_time": "09:00:00", "end_time": "18:00:00", "multiplier": float64(2)}},
	}, pricingDecoded["time_pricing"])
	require.Len(t, pricing.Intervals, 1)
	rawIv, err := json.Marshal(pricing.Intervals[0])
	require.NoError(t, err)
	var ivDecoded map[string]any
	require.NoError(t, json.Unmarshal(rawIv, &ivDecoded))
	for _, key := range []string{"id", "pricing_id", "sort_order"} {
		_, exists := ivDecoded[key]
		require.Falsef(t, exists, "user pricing interval must not expose %q", key)
	}
	for key, want := range map[string]float64{
		"input_multiplier": inputMultiplier, "output_multiplier": outputMultiplier,
		"cache_write_multiplier": cacheWriteMultiplier, "cache_read_multiplier": cacheReadMultiplier,
	} {
		got, exists := ivDecoded[key]
		require.Truef(t, exists, "user pricing interval must expose %q", key)
		require.InDelta(t, want, got.(float64), 1e-12)
	}
}

func TestUserMarketplaceGroup_ExposesImagePricingWithoutChangingAvailableGroup(t *testing.T) {
	price1K := 0.04
	price2K := 0.08
	price4K := 0.12
	group := service.Group{
		ID:                   7,
		Name:                 "image",
		Platform:             service.PlatformOpenAI,
		RateMultiplier:       0.8,
		ImageRateIndependent: true,
		ImageRateMultiplier:  0.5,
		ImagePrice1K:         &price1K,
		ImagePrice2K:         &price2K,
		ImagePrice4K:         &price4K,
	}

	marketplaceRaw, err := json.Marshal(toUserMarketplaceGroup(group))
	require.NoError(t, err)
	var marketplace map[string]any
	require.NoError(t, json.Unmarshal(marketplaceRaw, &marketplace))
	for _, key := range []string{"image_rate_independent", "image_rate_multiplier", "image_price_1k", "image_price_2k", "image_price_4k"} {
		_, exists := marketplace[key]
		require.Truef(t, exists, "marketplace group DTO must expose %q", key)
	}
	require.Equal(t, 0.04, marketplace["image_price_1k"])
	require.Equal(t, 0.5, marketplace["image_rate_multiplier"])

	availableRaw, err := json.Marshal(toUserAvailableGroup(group))
	require.NoError(t, err)
	var available map[string]any
	require.NoError(t, json.Unmarshal(availableRaw, &available))
	for _, key := range []string{"image_rate_independent", "image_rate_multiplier", "image_price_1k", "image_price_2k", "image_price_4k"} {
		_, exists := available[key]
		require.Falsef(t, exists, "available channel group DTO must not expose %q", key)
	}
}

func TestBuildPlatformSections_GroupsByPlatform(t *testing.T) {
	// 一个渠道横跨 anthropic / openai / 空平台：应该生成 2 个 section，
	// 按 platform 字母序排序，各自 groups 和 supported_models 只含同平台条目。
	ch := service.AvailableChannel{
		Name: "ch",
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: "anthropic"},
			{Name: "gpt-4o", Platform: "openai"},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "g-openai", Platform: "openai"},
		{ID: 2, Name: "g-ant", Platform: "anthropic"},
		{ID: 3, Name: "g-empty", Platform: ""},
	}
	sections := buildPlatformSections(ch, visible)
	require.Len(t, sections, 2)
	require.Equal(t, "anthropic", sections[0].Platform)
	require.Equal(t, "openai", sections[1].Platform)
	require.Len(t, sections[0].Groups, 1)
	require.Equal(t, int64(2), sections[0].Groups[0].ID)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}

func TestBuildPlatformSections_CompositeGroupExpandsAcrossConfiguredModelPlatforms(t *testing.T) {
	anthropicPrice := 3e-6
	openAIPrice := 2.5e-6
	ch := service.AvailableChannel{
		Name: "composite-channel",
		SupportedModels: []service.SupportedModel{
			{
				Name:     "claude-sonnet-4-6",
				Platform: service.PlatformAnthropic,
				Pricing:  &service.ChannelModelPricing{InputPrice: &anthropicPrice},
			},
			{
				Name:     "gpt-5",
				Platform: service.PlatformOpenAI,
				Pricing:  &service.ChannelModelPricing{InputPrice: &openAIPrice},
			},
		},
	}
	visible := []userAvailableGroup{
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 2)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Equal(t, service.PlatformOpenAI, sections[1].Platform)
	for _, section := range sections {
		require.Len(t, section.Groups, 1)
		require.Equal(t, int64(9), section.Groups[0].ID)
		require.Equal(t, service.PlatformComposite, section.Groups[0].Platform)
		require.Len(t, section.SupportedModels, 1)
		require.Equal(t, section.Platform, section.SupportedModels[0].Platform)
		require.NotNil(t, section.SupportedModels[0].Pricing)
	}
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
	require.Equal(t, "gpt-5", sections[1].SupportedModels[0].Name)
}

func TestBuildPlatformSections_OrdinaryGroupRemainsPlatformIsolated(t *testing.T) {
	ch := service.AvailableChannel{
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic},
			{Name: "gpt-5", Platform: service.PlatformOpenAI},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "anthropic-only", Platform: service.PlatformAnthropic},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}

func TestBuildPlatformSections_CompositeAndOrdinaryGroupsShareConcreteSection(t *testing.T) {
	ch := service.AvailableChannel{
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic},
			{Name: "gpt-5", Platform: service.PlatformOpenAI},
		},
	}
	visible := []userAvailableGroup{
		{ID: 1, Name: "anthropic-only", Platform: service.PlatformAnthropic},
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(ch, visible)

	require.Len(t, sections, 2)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Equal(t, []int64{1, 9}, []int64{
		sections[0].Groups[0].ID,
		sections[0].Groups[1].ID,
	})
	require.Equal(t, service.PlatformOpenAI, sections[1].Platform)
	require.Len(t, sections[1].Groups, 1)
	require.Equal(t, int64(9), sections[1].Groups[0].ID)
}

func TestBuildPlatformSections_CompositeWithoutModelsKeepsEmptyCompositeSection(t *testing.T) {
	visible := []userAvailableGroup{
		{ID: 9, Name: "composite", Platform: service.PlatformComposite},
	}

	sections := buildPlatformSections(service.AvailableChannel{
		SupportedModels: []service.SupportedModel{{Name: "invalid-without-platform"}},
	}, visible)

	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformComposite, sections[0].Platform)
	require.Len(t, sections[0].Groups, 1)
	require.Empty(t, sections[0].SupportedModels)
}
