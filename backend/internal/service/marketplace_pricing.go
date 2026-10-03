package service

import "context"

// MarketplacePricing returns base selling rates before group/user multipliers.
// Token tiers use billing probes; image tiers follow gateway override precedence.
func (r *ModelPricingResolver) MarketplacePricing(ctx context.Context, model, platform string, group *Group) *ChannelModelPricing {
	ctx = WithResolvedTargetPlatform(ctx, platform)
	if r.channelService != nil {
		mapping := r.channelService.ResolveChannelMapping(ctx, group.ID, model)
		// Marketplace prices are a deterministic reference price. Resolve aliases
		// to the upstream model for display even when a channel bills from the
		// upstream response; gateway billing still applies its configured source.
		if mapping.BillingModelSource != BillingModelSourceRequested && mapping.MappedModel != "" {
			model = mapping.MappedModel
		}
	}
	resolved := r.Resolve(ctx, PricingInput{Model: model, GroupID: &group.ID, Group: group})
	if resolved == nil {
		return nil
	}
	var raw *ChannelModelPricing
	if resolved.channelPricing != nil {
		cloned := resolved.channelPricing.Clone()
		raw = &cloned
	} else if base := resolved.BasePricing; base != nil {
		raw = &ChannelModelPricing{BillingMode: BillingModeToken,
			ImageInputPrice: nonZeroPtr(base.ImageInputPricePerToken), ImageOutputPrice: nonZeroPtr(base.ImageOutputPricePerToken)}
	}
	if resolved.Mode == BillingModeImage {
		raw = plazaImageDisplayPricing(raw, group)
		return raw
	}
	if resolved.Mode == BillingModePerRequest || resolved.Mode == BillingModeVideo {
		return raw
	}
	if schedule, err := r.billingService.ResolveContextPricingSchedule(ctx, r, ContextPricingScheduleInput{Model: model, Group: group, Platform: platform}); err == nil && schedule != nil && len(schedule.Tiers) > 0 {
		price := plazaPricingFromSchedule(raw, schedule)
		if raw != nil {
			price.TimePricing = raw.TimePricing
		}
		return price
	}
	return nil
}
