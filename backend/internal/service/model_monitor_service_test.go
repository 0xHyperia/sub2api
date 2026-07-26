package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type modelMonitorTestGroupRepo struct {
	GroupRepository
	groups []Group
}

func (r modelMonitorTestGroupRepo) ListActive(context.Context) ([]Group, error) {
	return r.groups, nil
}

type modelMonitorTestAccountRepo struct {
	AccountRepository
	accounts map[int64][]Account
}

func (r modelMonitorTestAccountRepo) ListSchedulableByGroupIDAndPlatform(_ context.Context, groupID int64, _ string) ([]Account, error) {
	return r.accounts[groupID], nil
}

type modelMonitorTestRepo struct {
	ModelMonitorRepository
	configs []ModelMonitor
}

func (r modelMonitorTestRepo) List(context.Context) ([]ModelMonitor, error) {
	return r.configs, nil
}

func (r modelMonitorTestRepo) ListGroupConfigs(context.Context) (map[int64][]int64, error) {
	return map[int64][]int64{}, nil
}

func (r modelMonitorTestRepo) Summaries(context.Context, []ModelCatalogEntry, int) (map[string]ModelMonitorSummary, error) {
	return map[string]ModelMonitorSummary{}, nil
}

func TestModelMonitorCatalogModelsUsesExplicitAccountRestrictions(t *testing.T) {
	group := Group{Platform: PlatformOpenAI}
	accounts := []Account{
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5":         "gpt-5",
					"gpt-5-codex":   "gpt-5-codex",
					"gpt-internal*": "gpt-internal*",
				},
			},
		},
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5":   "gpt-5.1",
					"o3-mini": "o3-mini",
				},
			},
		},
	}

	require.Equal(t, []string{"gpt-5", "gpt-5-codex", "o3-mini"}, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsDoesNotInferDefaultsFromUnrestrictedAccount(t *testing.T) {
	group := Group{Platform: PlatformOpenAI}
	accounts := []Account{{
		Platform:    PlatformOpenAI,
		Credentials: map[string]any{"api_key": "test"},
	}}

	require.Empty(t, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsIntersectsGroupListWithAccountRestrictions(t *testing.T) {
	group := Group{
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"gpt-5", "gpt-5-codex", "o3-mini", "unsupported", " gpt-5 "},
		},
	}
	accounts := []Account{
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"gpt-5": "gpt-5", "gpt-5-*": "gpt-5-*"},
			},
		},
		{
			Platform: PlatformOpenAI,
			Credentials: map[string]any{
				"model_mapping": map[string]string{"o3-mini": "o3-mini"},
			},
		},
	}

	require.Equal(t, []string{"gpt-5", "gpt-5-codex", "o3-mini"}, ModelCatalogModels(group, accounts))
}

func TestModelMonitorCatalogModelsRequiresAccountRestrictionEvenWithGroupList(t *testing.T) {
	group := Group{
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"gpt-5"},
		},
	}
	accounts := []Account{{Platform: PlatformOpenAI, Credentials: map[string]any{}}}

	require.Empty(t, ModelCatalogModels(group, accounts))
}

func TestModelMonitorDiscoverCatalogSkipsCompositeAliases(t *testing.T) {
	group := Group{ID: 7, Name: "Composite", Platform: PlatformComposite, Status: StatusActive}
	service := NewModelMonitorService(
		modelMonitorTestRepo{},
		modelMonitorTestGroupRepo{groups: []Group{group}},
		modelMonitorTestAccountRepo{accounts: map[int64][]Account{
			group.ID: {{Platform: PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"public-alias": "gpt-upstream"}}}},
		}},
		nil,
		nil,
		nil,
	)

	catalog, err := service.DiscoverCatalog(context.Background())
	require.NoError(t, err)
	require.Empty(t, catalog)
}

func TestModelMonitorListRowsHidesConfigsOutsideCurrentAccountRestrictions(t *testing.T) {
	group := Group{ID: 1, Name: "OpenAI", Platform: PlatformOpenAI, Status: StatusActive}
	account := Account{
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-5": "gpt-5"},
		},
	}
	service := NewModelMonitorService(
		modelMonitorTestRepo{configs: []ModelMonitor{
			{ID: 10, Platform: PlatformOpenAI, Model: "gpt-5", Enabled: true},
			{ID: 11, Platform: PlatformOpenAI, Model: "gpt-4-legacy", Enabled: true},
		}},
		modelMonitorTestGroupRepo{groups: []Group{group}},
		modelMonitorTestAccountRepo{accounts: map[int64][]Account{group.ID: {account}}},
		nil,
		nil,
		nil,
	)

	rows, err := service.ListRows(context.Background())

	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "gpt-5", rows[0].Model)
	require.True(t, rows[0].Configured)
	require.True(t, rows[0].CatalogAvailable)
}
