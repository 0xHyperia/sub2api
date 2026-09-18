package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelCatalogModelsIsolatesGroupsAndResolvesMappingInOrder(t *testing.T) {
	ctx := context.Background()
	channels := &ChannelService{}
	channels.cache.Store(populateChannelCache([]Channel{{
		ID: 1, Status: StatusActive, GroupIDs: []int64{10, 20},
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelMapping:       map[string]map[string]string{PlatformOpenAI: {"gpt-5.6-luna": "gpt-5.6-terra"}},
	}}, map[int64]string{10: PlatformOpenAI, 20: PlatformOpenAI}))
	account := Account{Platform: PlatformOpenAI,
		Credentials: map[string]any{"model_mapping": map[string]string{"gpt-5.6-terra": "provider-terra"}},
		Extra:       map[string]any{AccountModelCatalogKey: []string{"provider-terra", "gpt-6-astra"}},
	}
	group := Group{ID: 10, Platform: PlatformOpenAI, RateMultiplier: 0.1}
	require.Equal(t, []string{"gpt-5.6-luna", "gpt-5.6-terra"}, ChannelCatalogModels(ctx, channels, group, []Account{account}))
	other := Account{Platform: PlatformOpenAI, Extra: map[string]any{AccountModelCatalogKey: []string{"gpt-6-astra"}}}
	require.Equal(t, []string{"gpt-6-astra"}, ChannelCatalogModels(ctx, channels, Group{ID: 20, Platform: PlatformOpenAI, RateMultiplier: 0.2}, []Account{other}))
	group.ModelAllowlist = GroupModelAllowlist{Enabled: true, Models: []string{"gpt-5.6-luna"}}
	require.Equal(t, []string{"gpt-5.6-luna"}, ChannelCatalogModels(ctx, channels, group, []Account{account}))
	account.Extra[AccountModelCatalogKey] = []string{}
	require.Empty(t, ChannelCatalogModels(ctx, channels, group, []Account{account}))
}

func TestChannelCatalogModelsIntersectsSnapshotWithWildcardRestriction(t *testing.T) {
	account := Account{Platform: PlatformOpenAI,
		Credentials: map[string]any{"model_mapping": map[string]string{"gpt-5.6-*": "provider-terra"}},
		Extra:       map[string]any{AccountModelCatalogKey: []string{"gpt-5.6-terra", "provider-terra", "gpt-6-astra"}},
	}
	require.Equal(t, []string{"gpt-5.6-terra"}, ChannelCatalogModels(context.Background(), nil, Group{Platform: PlatformOpenAI}, []Account{account}))
}
