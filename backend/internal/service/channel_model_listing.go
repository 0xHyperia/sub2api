package service

import (
	"context"
	"sort"
	"strings"
)

type channelModelListAlias struct {
	name   string
	target string
}

// Aliases are projected from the unmodified source, never from another alias
// added during this pass: channel routing applies its mapping only once.
func (s *ChannelService) modelListAliases(ctx context.Context, groupID int64, platform string, models []string) []channelModelListAlias {
	if s == nil || len(models) == 0 {
		return nil
	}
	lk, err := s.lookupGroupChannel(ctx, groupID)
	if err != nil || lk == nil {
		return nil
	}
	if platform != "" {
		lk.platform = platform
	}
	available := make(map[string]bool, len(models))
	for _, model := range models {
		available[model] = true
	}
	candidates := make(map[string]bool)
	for _, p := range matchingPlatforms(lk.platform) {
		for name := range lk.channel.ModelMapping[p] {
			if name != "" && !strings.Contains(name, "*") && !available[name] {
				candidates[name] = true
			}
		}
	}
	names := make([]string, 0, len(candidates))
	for name := range candidates {
		names = append(names, name)
	}
	sort.Strings(names)
	aliases := make([]channelModelListAlias, 0, len(names))
	for _, name := range names {
		mapped := resolveMapping(lk, groupID, name)
		if mapped.Mapped && available[mapped.MappedModel] {
			aliases = append(aliases, channelModelListAlias{name: name, target: mapped.MappedModel})
		}
	}
	return aliases
}

// GetAvailableModelsForListing adds channel aliases after the account catalog
// cache, so channel updates cannot leave stale aliases in that shared cache.
// An unchanged empty catalog retains the handler's existing fallback behavior.
func (s *GatewayService) GetAvailableModelsForListing(ctx context.Context, groupID *int64, platform string, fallback []string) []string {
	models := s.GetAvailableModels(ctx, groupID, platform)
	if groupID == nil || s.channelService == nil {
		return models
	}
	source := models
	if len(source) == 0 {
		source = fallback
	}
	aliases := s.channelService.modelListAliases(ctx, *groupID, platform, source)
	if len(aliases) == 0 {
		return models
	}
	if len(models) == 0 {
		if _, ok := s.GetSchedulablePlatforms(ctx, groupID)[platform]; !ok {
			return models
		}
	}
	result := append([]string(nil), source...)
	for _, alias := range aliases {
		result = append(result, alias.name)
	}
	return result
}
