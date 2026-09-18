package handler

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AvailableChannelHandler 处理用户侧「可用渠道」查询。
//
// 用户侧接口委托 ChannelService.ListAvailable，并在返回前做四层过滤：
//  1. 行过滤：只保留状态为 Active 且与当前用户可访问分组有交集的渠道；
//  2. 分组过滤：渠道的 Groups 只保留用户可访问的那些；
//  3. 平台过滤：普通分组只保留自身平台模型；Composite 分组按渠道已配置的具体模型平台
//     展开。这样既防止普通分组跨平台泄漏，也让 Composite 正确展示其多平台能力；
//  4. 字段白名单：仅返回用户需要的字段（省略 BillingModelSource / RestrictModels
//     / 内部 ID / Status 等管理字段）。
type AvailableChannelHandler struct {
	channelService      *service.ChannelService
	apiKeyService       *service.APIKeyService
	settingService      *service.SettingService
	groupRepo           service.GroupRepository
	accountRepo         service.AccountRepository
	pricingService      *service.PricingService
	billingService      *service.BillingService
	pricingResolver     *service.ModelPricingResolver
	modelMonitorService *service.ModelMonitorService
}

// NewAvailableChannelHandler 创建用户侧可用渠道 handler。
func NewAvailableChannelHandler(
	channelService *service.ChannelService,
	apiKeyService *service.APIKeyService,
	settingService *service.SettingService,
	groupRepo service.GroupRepository,
	accountRepo service.AccountRepository,
	pricingService *service.PricingService,
	billingService *service.BillingService,
	pricingResolver *service.ModelPricingResolver,
	modelMonitorService *service.ModelMonitorService,
) *AvailableChannelHandler {
	return &AvailableChannelHandler{
		channelService:      channelService,
		apiKeyService:       apiKeyService,
		settingService:      settingService,
		groupRepo:           groupRepo,
		accountRepo:         accountRepo,
		pricingService:      pricingService,
		billingService:      billingService,
		pricingResolver:     pricingResolver,
		modelMonitorService: modelMonitorService,
	}
}

// featureEnabled 返回 available-channels 开关是否启用。默认关闭（opt-in）。
func (h *AvailableChannelHandler) featureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetAvailableChannelsRuntime(c.Request.Context()).Enabled
}

// marketplaceEnabled returns whether the independently gated model marketplace
// is enabled. It fails closed when settings are unavailable.
func (h *AvailableChannelHandler) marketplaceEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetModelMarketplaceRuntime(c.Request.Context()).Enabled
}

// userAvailableGroup 用户可见的分组概要（白名单字段）。
//
// 前端据此区分专属 vs 公开分组（IsExclusive）、订阅 vs 标准分组（SubscriptionType，
// 订阅视觉加深），并展示默认倍率与高峰倍率规则；用户专属倍率前端走
// /groups/rates，和 API 密钥页面保持一致。
type userAvailableGroup struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	Platform           string  `json:"platform"`
	SubscriptionType   string  `json:"subscription_type"`
	RateMultiplier     float64 `json:"rate_multiplier"`
	PeakRateEnabled    bool    `json:"peak_rate_enabled"`
	PeakStart          string  `json:"peak_start"`
	PeakEnd            string  `json:"peak_end"`
	PeakRateMultiplier float64 `json:"peak_rate_multiplier"`
	IsExclusive        bool    `json:"is_exclusive"`
	RPMLimit           int     `json:"rpm_limit"`
}

// userMarketplaceGroup extends the regular user-visible group summary with
// image-generation billing controls used only by /models/marketplace.
type userMarketplaceGroup struct {
	userAvailableGroup
	ImageRateIndependent bool                       `json:"image_rate_independent"`
	ImageRateMultiplier  float64                    `json:"image_rate_multiplier"`
	ImagePrice1K         *float64                   `json:"image_price_1k"`
	ImagePrice2K         *float64                   `json:"image_price_2k"`
	ImagePrice4K         *float64                   `json:"image_price_4k"`
	Pricing              *userSupportedModelPricing `json:"pricing"`
}

// userSupportedModelPricing 用户可见的定价字段白名单。
type userSupportedModelPricing struct {
	BillingMode                  string                      `json:"billing_mode"`
	InputPrice                   *float64                    `json:"input_price"`
	OutputPrice                  *float64                    `json:"output_price"`
	CacheWritePrice              *float64                    `json:"cache_write_price"`
	CacheWrite1hPrice            *float64                    `json:"cache_write_1h_price"`
	CacheReadPrice               *float64                    `json:"cache_read_price"`
	MaxReasoningEffortMultiplier *float64                    `json:"max_reasoning_effort_multiplier,omitempty"`
	ImageInputPrice              *float64                    `json:"image_input_price"`
	ImageOutputPrice             *float64                    `json:"image_output_price"`
	PerRequestPrice              *float64                    `json:"per_request_price"`
	Intervals                    []userPricingIntervalDTO    `json:"intervals"`
	TimePricing                  *service.ChannelTimePricing `json:"time_pricing,omitempty"`
}

// userPricingIntervalDTO 定价区间白名单（去掉内部 ID、SortOrder 等前端不渲染的字段）。
type userPricingIntervalDTO struct {
	MinTokens            int      `json:"min_tokens"`
	MaxTokens            *int     `json:"max_tokens"`
	TierLabel            string   `json:"tier_label,omitempty"`
	InputPrice           *float64 `json:"input_price"`
	OutputPrice          *float64 `json:"output_price"`
	CacheWritePrice      *float64 `json:"cache_write_price"`
	CacheWrite1hPrice    *float64 `json:"cache_write_1h_price"`
	CacheReadPrice       *float64 `json:"cache_read_price"`
	InputMultiplier      *float64 `json:"input_multiplier"`
	OutputMultiplier     *float64 `json:"output_multiplier"`
	CacheWriteMultiplier *float64 `json:"cache_write_multiplier"`
	CacheReadMultiplier  *float64 `json:"cache_read_multiplier"`
	PerRequestPrice      *float64 `json:"per_request_price"`
}

// userSupportedModel 用户可见的支持模型条目。
type userSupportedModel struct {
	Name     string                     `json:"name"`
	Platform string                     `json:"platform"`
	Pricing  *userSupportedModelPricing `json:"pricing"`
}

type userMarketplaceModel struct {
	Name          string                       `json:"name"`
	Platform      string                       `json:"platform"`
	Pricing       *userSupportedModelPricing   `json:"pricing"`
	Capabilities  []string                     `json:"capabilities"`
	Groups        []userMarketplaceGroup       `json:"groups"`
	MonitorStatus *service.ModelMonitorSummary `json:"monitor_status"`
}

// userChannelPlatformSection 单渠道内某个平台的子视图：用户可见的分组 + 该平台
// 支持的模型。按 platform 聚合后让前端可以把渠道名作为 row-group 一次渲染，
// 后面的平台行按 sections 顺序铺开。
type userChannelPlatformSection struct {
	Platform        string               `json:"platform"`
	Groups          []userAvailableGroup `json:"groups"`
	SupportedModels []userSupportedModel `json:"supported_models"`
}

// userAvailableChannel 用户可见的渠道条目（白名单字段）。
//
// 每个渠道聚合为一条记录，内嵌 platforms 子数组：每个 section 对应一个平台，
// 包含该平台的 groups 和 supported_models。
type userAvailableChannel struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Platforms   []userChannelPlatformSection `json:"platforms"`
}

// userMarketplacePlatform is the channel-independent marketplace view. Models
// are aggregated from accessible groups and their schedulable accounts.
type userMarketplacePlatform struct {
	Platform        string                 `json:"platform"`
	Groups          []userMarketplaceGroup `json:"groups"`
	SupportedModels []userMarketplaceModel `json:"supported_models"`
}

const publicShowcaseModelLimit = 6

// publicModelShowcaseModel is the anonymous-safe model summary used by the
// landing page. Group identities are intentionally omitted.
type publicModelShowcaseModel struct {
	Name           string                       `json:"name"`
	Platform       string                       `json:"platform"`
	Pricing        *userSupportedModelPricing   `json:"pricing"`
	RateMultiplier float64                      `json:"rate_multiplier"`
	MonitorStatus  *service.ModelMonitorSummary `json:"monitor_status"`
}

type publicModelShowcasePlatform struct {
	Platform   string                     `json:"platform"`
	ModelCount int                        `json:"model_count"`
	Models     []publicModelShowcaseModel `json:"models"`
}

// List 列出当前用户可见的「可用渠道」。
// GET /api/v1/channels/available
func (h *AvailableChannelHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Feature 未启用时返回空数组（不暴露渠道信息）。检查放在认证之后，
	// 保持与未开关前的 401 行为一致：未登录先 401，登录后再按开关决定。
	if !h.featureEnabled(c) {
		response.Success(c, []userAvailableChannel{})
		return
	}

	h.listForUser(c, subject.UserID)
}

func (h *AvailableChannelHandler) marketplaceForUser(ctx context.Context, userID int64, resolution service.ModelMonitorResolution) ([]userMarketplacePlatform, error) {
	groups, err := h.apiKeyService.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	type modelAggregate struct {
		groups map[int64]userMarketplaceGroup
	}
	byPlatform := make(map[string]map[string]*modelAggregate)

	for i := range groups {
		group := groups[i]
		platform := strings.TrimSpace(group.Platform)
		if platform == "" {
			continue
		}
		accounts, listErr := marketplaceGroupAccounts(ctx, h.accountRepo, group.ID, platform)
		if listErr != nil {
			return nil, listErr
		}
		if len(accounts) == 0 {
			continue
		}

		models := marketplaceModelIDsForGroup(ctx, h.channelService, group, accounts)
		if len(models) == 0 {
			continue
		}
		if byPlatform[platform] == nil {
			byPlatform[platform] = make(map[string]*modelAggregate)
		}
		for _, model := range models {
			aggregate := byPlatform[platform][model]
			if aggregate == nil {
				aggregate = &modelAggregate{groups: make(map[int64]userMarketplaceGroup)}
				byPlatform[platform][model] = aggregate
			}
			groupView := toUserMarketplaceGroup(group)
			groupView.Pricing = h.marketplacePricing(ctx, model, platform, group)
			aggregate.groups[group.ID] = groupView
		}
	}

	platforms := make([]string, 0, len(byPlatform))
	for platform := range byPlatform {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)

	out := make([]userMarketplacePlatform, 0, len(platforms))
	for _, platform := range platforms {
		modelNames := make([]string, 0, len(byPlatform[platform]))
		allGroups := make(map[int64]userMarketplaceGroup)
		for name, aggregate := range byPlatform[platform] {
			modelNames = append(modelNames, name)
			for id, group := range aggregate.groups {
				allGroups[id] = group
			}
		}
		sort.Strings(modelNames)
		models := make([]userMarketplaceModel, 0, len(modelNames))
		for _, name := range modelNames {
			aggregate := byPlatform[platform][name]
			modelGroups := sortedMarketplaceGroups(aggregate.groups)
			var pricing *userSupportedModelPricing
			for _, group := range modelGroups {
				if group.Pricing != nil {
					pricing = group.Pricing
					break
				}
			}
			models = append(models, userMarketplaceModel{
				Name:         name,
				Platform:     platform,
				Pricing:      pricing,
				Capabilities: h.pricingService.GetModelCapabilities(name),
				Groups:       modelGroups,
			})
		}
		out = append(out, userMarketplacePlatform{
			Platform:        platform,
			Groups:          sortedMarketplaceGroups(allGroups),
			SupportedModels: models,
		})
	}
	if h.modelMonitorService != nil {
		showDetailedPerformance := h.settingService == nil || h.settingService.ModelMarketplacePerformanceVisible(ctx)
		keys := make([]service.ModelCatalogEntry, 0)
		for _, section := range out {
			for _, model := range section.SupportedModels {
				keys = append(keys, service.ModelCatalogEntry{Platform: section.Platform, Model: model.Name})
			}
		}
		summaries, summaryErr := h.modelMonitorService.PublicSummaries(ctx, keys, resolution)
		if summaryErr != nil {
			return nil, summaryErr
		}
		for i := range out {
			for j := range out[i].SupportedModels {
				if summary, ok := summaries[service.ModelMonitorKey(out[i].Platform, out[i].SupportedModels[j].Name)]; ok {
					copy := summary
					if !showDetailedPerformance {
						service.RedactModelMonitorDetailedPerformance(&copy)
					}
					out[i].SupportedModels[j].MonitorStatus = &copy
				}
			}
		}
	}
	return out, nil
}

func (h *AvailableChannelHandler) showcaseForPublic(ctx context.Context) ([]publicModelShowcasePlatform, error) {
	if h.groupRepo == nil || h.accountRepo == nil || h.pricingService == nil {
		return nil, errors.New("model showcase dependencies unavailable")
	}

	groups, err := h.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	type modelAggregate struct {
		rate    float64
		hasRate bool
		pricing *userSupportedModelPricing
	}
	byPlatform := make(map[string]map[string]*modelAggregate)
	for i := range groups {
		group := groups[i]
		platform := strings.TrimSpace(group.Platform)
		if platform == "" || group.IsExclusive {
			continue
		}
		accounts, listErr := marketplaceGroupAccounts(ctx, h.accountRepo, group.ID, platform)
		if listErr != nil {
			return nil, listErr
		}
		if len(accounts) == 0 {
			continue
		}
		if byPlatform[platform] == nil {
			byPlatform[platform] = make(map[string]*modelAggregate)
		}
		for _, model := range marketplaceModelIDsForGroup(ctx, h.channelService, group, accounts) {
			aggregate := byPlatform[platform][model]
			if aggregate == nil {
				aggregate = &modelAggregate{}
				byPlatform[platform][model] = aggregate
			}
			if !aggregate.hasRate || group.RateMultiplier < aggregate.rate {
				aggregate.rate = group.RateMultiplier
				aggregate.hasRate = true
				aggregate.pricing = h.marketplacePricing(ctx, model, platform, group)
			}
		}
	}

	out := make([]publicModelShowcasePlatform, 0, len(byPlatform))
	keys := make([]service.ModelCatalogEntry, 0)
	for platform, modelsByName := range byPlatform {
		models := make([]publicModelShowcaseModel, 0, len(modelsByName))
		for name, aggregate := range modelsByName {
			models = append(models, publicModelShowcaseModel{
				Name:           name,
				Platform:       platform,
				Pricing:        aggregate.pricing,
				RateMultiplier: aggregate.rate,
			})
			keys = append(keys, service.ModelCatalogEntry{Platform: platform, Model: name})
		}
		out = append(out, publicModelShowcasePlatform{Platform: platform, ModelCount: len(models), Models: models})
	}

	if h.modelMonitorService != nil {
		showDetailedPerformance := h.settingService == nil || h.settingService.ModelMarketplacePerformanceVisible(ctx)
		summaries, summaryErr := h.modelMonitorService.PublicSummaries(ctx, keys, service.ModelMonitorResolutionHour)
		if summaryErr != nil {
			return nil, summaryErr
		}
		for i := range out {
			for j := range out[i].Models {
				key := service.ModelMonitorKey(out[i].Platform, out[i].Models[j].Name)
				if summary, ok := summaries[key]; ok {
					copy := summary
					if !showDetailedPerformance {
						service.RedactModelMonitorDetailedPerformance(&copy)
					}
					out[i].Models[j].MonitorStatus = &copy
				}
			}
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		left, right := showcasePlatformRank(out[i].Platform), showcasePlatformRank(out[j].Platform)
		if left != right {
			return left < right
		}
		return out[i].Platform < out[j].Platform
	})
	for i := range out {
		sort.SliceStable(out[i].Models, func(a, b int) bool {
			left, right := out[i].Models[a], out[i].Models[b]
			leftOrder, rightOrder := 0, 0
			if left.MonitorStatus != nil {
				leftOrder = left.MonitorStatus.DisplayOrder
			}
			if right.MonitorStatus != nil {
				rightOrder = right.MonitorStatus.DisplayOrder
			}
			if leftOrder != rightOrder {
				return leftOrder > rightOrder
			}
			return left.Name > right.Name
		})
		if len(out[i].Models) > publicShowcaseModelLimit {
			out[i].Models = out[i].Models[:publicShowcaseModelLimit]
		}
	}
	return out, nil
}

func showcasePlatformRank(platform string) int {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case service.PlatformOpenAI:
		return 0
	case service.PlatformAnthropic:
		return 1
	case service.PlatformGemini:
		return 2
	case service.PlatformGrok:
		return 3
	case service.PlatformAntigravity:
		return 4
	default:
		return 100
	}
}

func marketplaceModelIDsForGroup(ctx context.Context, channels *service.ChannelService, group service.Group, accounts []service.Account) []string {
	return service.ChannelCatalogModels(ctx, channels, group, accounts)
}

func marketplaceGroupAccounts(ctx context.Context, repo service.AccountRepository, groupID int64, platform string) ([]service.Account, error) {
	if platform == service.PlatformComposite {
		return repo.ListModelAvailabilityCandidates(ctx, &groupID, []string{
			service.PlatformAnthropic,
			service.PlatformOpenAI,
			service.PlatformGemini,
			service.PlatformAntigravity,
			service.PlatformGrok,
		}, false)
	}
	return repo.ListSchedulableByGroupIDAndPlatform(ctx, groupID, platform)
}

func toUserAvailableGroup(group service.Group) userAvailableGroup {
	return userAvailableGroup{
		ID:                 group.ID,
		Name:               group.Name,
		Platform:           group.Platform,
		SubscriptionType:   group.SubscriptionType,
		RateMultiplier:     group.RateMultiplier,
		PeakRateEnabled:    group.PeakRateEnabled,
		PeakStart:          group.PeakStart,
		PeakEnd:            group.PeakEnd,
		PeakRateMultiplier: group.PeakRateMultiplier,
		IsExclusive:        group.IsExclusive,
		RPMLimit:           group.RPMLimit,
	}
}

func toUserMarketplaceGroup(group service.Group) userMarketplaceGroup {
	return userMarketplaceGroup{
		userAvailableGroup:   toUserAvailableGroup(group),
		ImageRateIndependent: group.ImageRateIndependent,
		ImageRateMultiplier:  group.ImageRateMultiplier,
		ImagePrice1K:         group.ImagePrice1K,
		ImagePrice2K:         group.ImagePrice2K,
		ImagePrice4K:         group.ImagePrice4K,
	}
}

func (h *AvailableChannelHandler) marketplacePricing(ctx context.Context, model, platform string, group service.Group) *userSupportedModelPricing {
	if h.billingService == nil || h.pricingResolver == nil {
		return toUserPricing(h.pricingService.GetDisplayModelPricing(model))
	}
	return toUserPricing(h.pricingResolver.MarketplacePricing(ctx, model, platform, &group))
}

func sortedMarketplaceGroups(groups map[int64]userMarketplaceGroup) []userMarketplaceGroup {
	out := make([]userMarketplaceGroup, 0, len(groups))
	for _, group := range groups {
		out = append(out, group)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RateMultiplier != out[j].RateMultiplier {
			return out[i].RateMultiplier < out[j].RateMultiplier
		}
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// ListMarketplace returns a channel-independent, group-derived model catalog.
// GET /api/v1/models/marketplace
func (h *AvailableChannelHandler) ListMarketplace(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if !h.marketplaceEnabled(c) {
		response.Success(c, []userMarketplacePlatform{})
		return
	}

	resolution := service.ModelMonitorResolution(strings.ToLower(strings.TrimSpace(c.DefaultQuery("resolution", "hour"))))
	out, err := h.marketplaceForUser(c.Request.Context(), subject.UserID, resolution)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

// ListShowcase returns a bounded, anonymous-safe catalog for the landing page.
// GET /api/v1/models/showcase
func (h *AvailableChannelHandler) ListShowcase(c *gin.Context) {
	if !h.marketplaceEnabled(c) {
		response.Success(c, []publicModelShowcasePlatform{})
		return
	}
	out, err := h.showcaseForPublic(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

func (h *AvailableChannelHandler) listForUser(c *gin.Context, userID int64) {
	out, err := h.availableForUser(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, out)
}

func (h *AvailableChannelHandler) availableForUser(ctx context.Context, userID int64) ([]userAvailableChannel, error) {
	userGroups, err := h.apiKeyService.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, err
	}
	allowedGroupIDs := make(map[int64]struct{}, len(userGroups))
	for i := range userGroups {
		allowedGroupIDs[userGroups[i].ID] = struct{}{}
	}

	channels, err := h.channelService.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]userAvailableChannel, 0, len(channels))
	for _, ch := range channels {
		if ch.Status != service.StatusActive {
			continue
		}
		visibleGroups := filterUserVisibleGroups(ch.Groups, allowedGroupIDs)
		if len(visibleGroups) == 0 {
			continue
		}
		sections := buildPlatformSections(ch, visibleGroups)
		if len(sections) == 0 {
			continue
		}
		out = append(out, userAvailableChannel{
			Name:        ch.Name,
			Description: ch.Description,
			Platforms:   sections,
		})
	}

	return out, nil
}

// buildPlatformSections 把一个渠道按 visibleGroups 的平台集合拆成有序的 section 列表：
// 每个 section 对应一个具体平台，只包含该平台的 groups 和 supported_models。
//
// Composite 分组可访问渠道中所有已配置的具体平台，因此会被展开到每个有支持模型的
// 平台 section。普通分组仍严格留在自身平台，避免跨平台模型信息泄漏。Composite 渠道
// 尚未配置任何模型时保留 composite section，以便前端继续展示该分组和“未配置模型”状态。
// 输出按 platform 字母序稳定排序，便于前端等效比较与回归测试。
func buildPlatformSections(
	ch service.AvailableChannel,
	visibleGroups []userAvailableGroup,
) []userChannelPlatformSection {
	groupsByPlatform := make(map[string][]userAvailableGroup, 4)
	compositeGroups := make([]userAvailableGroup, 0, 1)
	for _, g := range visibleGroups {
		if g.Platform == "" {
			continue
		}
		if g.Platform == service.PlatformComposite {
			compositeGroups = append(compositeGroups, g)
			continue
		}
		groupsByPlatform[g.Platform] = append(groupsByPlatform[g.Platform], g)
	}

	if len(compositeGroups) > 0 {
		modelPlatforms := make(map[string]struct{}, len(ch.SupportedModels))
		for i := range ch.SupportedModels {
			if platform := ch.SupportedModels[i].Platform; platform != "" {
				modelPlatforms[platform] = struct{}{}
			}
		}
		if len(modelPlatforms) == 0 {
			groupsByPlatform[service.PlatformComposite] = append(
				groupsByPlatform[service.PlatformComposite],
				compositeGroups...,
			)
		} else {
			for platform := range modelPlatforms {
				groupsByPlatform[platform] = append(groupsByPlatform[platform], compositeGroups...)
			}
		}
	}
	if len(groupsByPlatform) == 0 {
		return nil
	}

	platforms := make([]string, 0, len(groupsByPlatform))
	for p := range groupsByPlatform {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)

	sections := make([]userChannelPlatformSection, 0, len(platforms))
	for _, platform := range platforms {
		platformSet := map[string]struct{}{platform: {}}
		sections = append(sections, userChannelPlatformSection{
			Platform:        platform,
			Groups:          groupsByPlatform[platform],
			SupportedModels: toUserSupportedModels(ch.SupportedModels, platformSet),
		})
	}
	return sections
}

// filterUserVisibleGroups 仅保留用户可访问的分组。
func filterUserVisibleGroups(
	groups []service.AvailableGroupRef,
	allowed map[int64]struct{},
) []userAvailableGroup {
	visible := make([]userAvailableGroup, 0, len(groups))
	for _, g := range groups {
		if _, ok := allowed[g.ID]; !ok {
			continue
		}
		visible = append(visible, userAvailableGroup{
			ID:                 g.ID,
			Name:               g.Name,
			Platform:           g.Platform,
			SubscriptionType:   g.SubscriptionType,
			RateMultiplier:     g.RateMultiplier,
			PeakRateEnabled:    g.PeakRateEnabled,
			PeakStart:          g.PeakStart,
			PeakEnd:            g.PeakEnd,
			PeakRateMultiplier: g.PeakRateMultiplier,
			IsExclusive:        g.IsExclusive,
		})
	}
	return visible
}

// toUserSupportedModels 将 service 层支持模型转换为用户 DTO（字段白名单）。
// 仅保留平台在 allowedPlatforms 中的条目，防止跨平台模型信息泄漏。
// allowedPlatforms 为 nil 时不做平台过滤（保留全部，供测试或明确无过滤场景使用）。
func toUserSupportedModels(
	src []service.SupportedModel,
	allowedPlatforms map[string]struct{},
) []userSupportedModel {
	out := make([]userSupportedModel, 0, len(src))
	for i := range src {
		m := src[i]
		if allowedPlatforms != nil {
			if _, ok := allowedPlatforms[m.Platform]; !ok {
				continue
			}
		}
		out = append(out, userSupportedModel{
			Name:     m.Name,
			Platform: m.Platform,
			Pricing:  toUserPricing(m.Pricing),
		})
	}
	return out
}

// toUserPricingIntervals 将定价区间转换为用户 DTO 白名单形态；nil 入参返回 nil（JSON omitempty 可省略）。
func toUserPricingIntervals(src []service.PricingInterval) []userPricingIntervalDTO {
	if src == nil {
		return nil
	}
	intervals := make([]userPricingIntervalDTO, 0, len(src))
	for _, iv := range src {
		intervals = append(intervals, userPricingIntervalDTO{
			MinTokens:            iv.MinTokens,
			MaxTokens:            iv.MaxTokens,
			TierLabel:            iv.TierLabel,
			InputPrice:           iv.InputPrice,
			OutputPrice:          iv.OutputPrice,
			CacheWritePrice:      iv.CacheWritePrice,
			CacheWrite1hPrice:    iv.CacheWrite1hPrice,
			CacheReadPrice:       iv.CacheReadPrice,
			InputMultiplier:      iv.InputMultiplier,
			OutputMultiplier:     iv.OutputMultiplier,
			CacheWriteMultiplier: iv.CacheWriteMultiplier,
			CacheReadMultiplier:  iv.CacheReadMultiplier,
			PerRequestPrice:      iv.PerRequestPrice,
		})
	}
	return intervals
}

// toUserPricing 将 service 层定价转换为用户 DTO；入参为 nil 时返回 nil。
func toUserPricing(p *service.ChannelModelPricing) *userSupportedModelPricing {
	if p == nil {
		return nil
	}
	intervals := toUserPricingIntervals(p.Intervals)
	if intervals == nil {
		// 用户侧定价的 intervals 固定输出数组（空配置为 []），保持既有契约。
		intervals = []userPricingIntervalDTO{}
	}
	billingMode := string(p.BillingMode)
	if billingMode == "" {
		billingMode = string(service.BillingModeToken)
	}
	return &userSupportedModelPricing{
		BillingMode:                  billingMode,
		InputPrice:                   p.InputPrice,
		OutputPrice:                  p.OutputPrice,
		CacheWritePrice:              p.CacheWritePrice,
		CacheWrite1hPrice:            p.CacheWrite1hPrice,
		CacheReadPrice:               p.CacheReadPrice,
		MaxReasoningEffortMultiplier: p.MaxReasoningEffortMultiplier,
		ImageInputPrice:              p.ImageInputPrice,
		ImageOutputPrice:             p.ImageOutputPrice,
		PerRequestPrice:              p.PerRequestPrice,
		Intervals:                    intervals,
		TimePricing:                  p.TimePricing,
	}
}
