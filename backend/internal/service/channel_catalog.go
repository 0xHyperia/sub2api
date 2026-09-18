package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
)

const AccountModelCatalogKey = "marketplace_model_catalog"

// Catalog IDs are observations, never an account mapping or an access grant.
func catalogAccountMapping(account *Account) map[string]string {
	if account == nil {
		return nil
	}
	// Do not call GetModelMapping for an empty mapping: some providers expose a
	// built-in model catalog, but marketplace availability must come from the
	// account's explicit mapping or a live catalog snapshot.
	rawMapping := account.Credentials["model_mapping"]
	mapping := stringMappingFromRaw(rawMapping)
	out := make(map[string]string, len(mapping))
	for source, target := range mapping {
		out[source] = target
	}
	var models []string
	if raw, err := json.Marshal(account.Extra[AccountModelCatalogKey]); err == nil {
		_ = json.Unmarshal(raw, &models)
	}
	for _, model := range models {
		if mapped, ok := resolveRequestedModelInMapping(mapping, model); ok {
			out[model] = mapped
		} else if len(mapping) == 0 {
			// Observations supply names only for unrestricted accounts.
			out[model] = model
		}
	}
	return out
}

// ChannelCatalogModels keeps aliases additive and tests each against the same
// account catalog. Mapping values are private and do not become public models.
func ChannelCatalogModels(ctx context.Context, channels *ChannelService, group Group, accounts []Account) []string {
	unfiltered := group
	unfiltered.ModelAllowlist = GroupModelAllowlist{}
	candidates := ModelCatalogModels(unfiltered, accounts)
	if group.ModelAllowlist.Enabled {
		candidates = append(candidates, group.ModelAllowlist.Models...)
	}
	if channels != nil {
		if channel, err := channels.GetChannelForGroup(ctx, group.ID); err == nil && channel != nil {
			for _, platform := range matchingPlatforms(group.Platform) {
				for source := range channel.ModelMapping[platform] {
					candidates = append(candidates, source)
				}
			}
		}
	}
	seen := make(map[string]bool)
	models := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		model := strings.TrimSpace(candidate)
		if model == "" || strings.Contains(model, "*") || seen[model] || !group.ModelAllowlist.Allows(model) {
			continue
		}
		mapped := ChannelMappingResult{MappedModel: model}
		if channels != nil {
			mapped = channels.ResolveChannelMapping(ctx, group.ID, model)
			billingModel := billingModelForRestriction(mapped.BillingModelSource, model, mapped.MappedModel)
			if billingModel != "" && channels.IsModelRestricted(ctx, group.ID, billingModel) {
				continue
			}
		}
		for i := range accounts {
			account := &accounts[i]
			observedModel := mapped.MappedModel
			accountMapping := catalogAccountMapping(account)
			if mappedModel, ok := resolveRequestedModelInMapping(accountMapping, mapped.MappedModel); ok {
				observedModel = mappedModel
			}
			if catalogSnapshotPresent(account) && !catalogSnapshotContains(account, observedModel) {
				continue
			}
			if !mappingSupportsRequestedModel(accountMapping, mapped.MappedModel) {
				continue
			}
			if mapped.BillingModelSource == BillingModelSourceUpstream && channels != nil && channels.IsModelRestricted(ctx, group.ID, resolveAccountUpstreamModel(account, mapped.MappedModel)) {
				continue
			}
			seen[model] = true
			models = append(models, model)
			break
		}
	}
	sort.Strings(models)
	return models
}

func catalogSnapshotPresent(account *Account) bool {
	if account == nil || account.Extra == nil {
		return false
	}
	var models []string
	raw, err := json.Marshal(account.Extra[AccountModelCatalogKey])
	if err != nil || json.Unmarshal(raw, &models) != nil {
		return false
	}
	return models != nil
}

func catalogSnapshotContains(account *Account, model string) bool {
	if account == nil || account.Extra == nil {
		return false
	}
	var models []string
	raw, err := json.Marshal(account.Extra[AccountModelCatalogKey])
	if err != nil || json.Unmarshal(raw, &models) != nil {
		return false
	}
	for _, candidate := range models {
		if strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(model)) {
			return true
		}
	}
	return false
}

func (c *Channel) inheritsCatalogPricing(platform, model string) bool {
	var entries map[string][]string
	raw, err := json.Marshal(c.FeaturesConfig["catalog_pricing"])
	if err != nil || json.Unmarshal(raw, &entries) != nil {
		return false
	}
	for _, pattern := range entries[platform] {
		if strings.EqualFold(pattern, model) {
			return true
		}
	}
	return false
}
