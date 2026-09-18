package service

import (
	"context"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
)

func TestChannelModelListAliasesRemainWithinGroupAndPlatform(t *testing.T) {
	channels := &ChannelService{}
	channels.cache.Store(populateChannelCache([]Channel{{
		ID: 1, Status: StatusActive, GroupIDs: []int64{10, 20},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {"gpt-5.6-luna": "gpt-5.6-terra", "chained": "gpt-5.6-luna", "wild-*": "gpt-5.6-terra"},
			PlatformGemini: {"gemini-alias": "gpt-5.6-terra"},
		},
	}}, map[int64]string{10: PlatformOpenAI, 20: PlatformComposite}))
	for _, tt := range []struct {
		name     string
		groupID  int64
		platform string
		models   []string
		want     []channelModelListAlias
	}{
		{"additive", 10, PlatformOpenAI, []string{"gpt-5.6-terra"}, []channelModelListAlias{{"gpt-5.6-luna", "gpt-5.6-terra"}}},
		{"missing target", 10, PlatformOpenAI, []string{"gpt-6-astra"}, nil},
		{"unlinked group", 30, PlatformOpenAI, []string{"gpt-5.6-terra"}, nil},
		{"duplicate", 10, PlatformOpenAI, []string{"gpt-5.6-terra", "gpt-5.6-luna", "chained"}, nil},
		{"composite openai", 20, PlatformOpenAI, []string{"gpt-5.6-terra"}, []channelModelListAlias{{"gpt-5.6-luna", "gpt-5.6-terra"}}},
		{"composite gemini", 20, PlatformGemini, []string{"gpt-5.6-terra"}, []channelModelListAlias{{"gemini-alias", "gpt-5.6-terra"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := channels.modelListAliases(context.Background(), tt.groupID, tt.platform, tt.models)
			require.Equal(t, tt.want, append([]channelModelListAlias(nil), got...))
		})
	}
}

func TestAvailableModelsForListingProjectsCurrentChannelOverCachedAccounts(t *testing.T) {
	ctx := context.Background()
	groupID := int64(10)
	channels := &ChannelService{}
	setMapping := func(mapping map[string]string) {
		channels.cache.Store(populateChannelCache([]Channel{{
			ID: 1, Status: StatusActive, GroupIDs: []int64{groupID},
			ModelMapping: map[string]map[string]string{PlatformOpenAI: mapping},
		}}, map[int64]string{groupID: PlatformOpenAI}))
	}
	setMapping(map[string]string{"gpt-5.6-luna": "gpt-5.6-terra"})
	repo := &modelsListAccountRepoStub{byGroup: map[int64][]Account{groupID: {{
		Platform:    PlatformOpenAI,
		Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5.6-terra": "provider-terra"}},
	}}}}
	svc := &GatewayService{accountRepo: repo, channelService: channels,
		modelsListCache: gocache.New(time.Minute, time.Minute), modelsListCacheTTL: time.Minute}
	require.Equal(t, []string{"gpt-5.6-terra", "gpt-5.6-luna"}, svc.GetAvailableModelsForListing(ctx, &groupID, PlatformOpenAI, nil))
	require.Equal(t, []string{"gpt-5.6-terra"}, svc.GetAvailableModels(ctx, &groupID, PlatformOpenAI))
	setMapping(map[string]string{"new-alias": "gpt-5.6-terra"})
	require.Equal(t, []string{"gpt-5.6-terra", "new-alias"}, svc.GetAvailableModelsForListing(ctx, &groupID, PlatformOpenAI, nil))
	setMapping(nil)
	require.Equal(t, []string{"gpt-5.6-terra"}, svc.GetAvailableModelsForListing(ctx, &groupID, PlatformOpenAI, nil))
	require.EqualValues(t, 1, repo.listByGroupCalls.Load())
}
