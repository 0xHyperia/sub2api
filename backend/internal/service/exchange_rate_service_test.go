package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type exchangeRateRepoStub struct {
	values map[string]string
}

func (r *exchangeRateRepoStub) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}
func (r *exchangeRateRepoStub) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}
func (r *exchangeRateRepoStub) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}
func (r *exchangeRateRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}
func (r *exchangeRateRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		r.values[key] = value
	}
	return nil
}
func (r *exchangeRateRepoStub) GetAll(context.Context) (map[string]string, error) {
	return r.values, nil
}
func (r *exchangeRateRepoStub) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}

func TestExchangeRateSettingsMigrationAndEffectiveRate(t *testing.T) {
	repo := &exchangeRateRepoStub{values: map[string]string{SettingSubscriptionUSDToCNYRate: "7.16"}}
	service := NewSettingService(repo, nil)

	settings, err := service.GetExchangeRateSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 7.16, settings.ManualUSDToCNY)
	require.Equal(t, 7.16, settings.Effective)
	require.Equal(t, "manual", settings.Source)

	repo.values[SettingKeyCurrencyExchangeRateAutoSync] = "true"
	repo.values[SettingKeyCurrencyUSDToCNYAutoRate] = "7.25"
	settings, err = service.GetExchangeRateSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, 7.25, settings.Effective)
	require.Equal(t, "auto", settings.Source)
}

func TestEnsureExchangeRateSettingsMaterializesLegacyRate(t *testing.T) {
	repo := &exchangeRateRepoStub{values: map[string]string{SettingSubscriptionUSDToCNYRate: "7.16"}}
	service := NewSettingService(repo, nil)

	require.NoError(t, service.EnsureExchangeRateSettings(context.Background()))
	require.Equal(t, "7.160000", repo.values[SettingKeyCurrencyUSDToCNYManualRate])
	require.Equal(t, "false", repo.values[SettingKeyCurrencyExchangeRateAutoSync])

	repo.values[SettingSubscriptionUSDToCNYRate] = "8.88"
	require.NoError(t, service.EnsureExchangeRateSettings(context.Background()))
	require.Equal(t, "7.160000", repo.values[SettingKeyCurrencyUSDToCNYManualRate])
}

func TestSyncExchangeRateStoresValidProviderSnapshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"date":"2026-08-04","rates":{"CNY":7.22}}`))
	}))
	defer server.Close()
	repo := &exchangeRateRepoStub{values: map[string]string{
		SettingKeyCurrencyUSDToCNYManualRate:   "7.2",
		SettingKeyCurrencyExchangeRateAutoSync: "true",
	}}
	service := NewSettingService(repo, nil)
	service.exchangeRateEndpoint = server.URL

	settings, err := service.SyncExchangeRate(context.Background(), true)
	require.NoError(t, err)
	require.Equal(t, 7.22, settings.Effective)
	require.Equal(t, "auto", settings.Source)
	require.Equal(t, "2026-08-04", settings.ProviderAsOf)
	require.Empty(t, settings.LastError)
	require.NotEmpty(t, settings.LastSyncedAt)
}

func TestSyncExchangeRateRejectsSuspiciousJumpAndKeepsLastRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"date":"2026-08-04","rates":{"CNY":9.5}}`))
	}))
	defer server.Close()
	repo := &exchangeRateRepoStub{values: map[string]string{
		SettingKeyCurrencyUSDToCNYManualRate:   "7.2",
		SettingKeyCurrencyExchangeRateAutoSync: "true",
		SettingKeyCurrencyUSDToCNYAutoRate:     "7.21",
	}}
	service := NewSettingService(repo, nil)
	service.exchangeRateEndpoint = server.URL

	_, err := service.SyncExchangeRate(context.Background(), true)
	require.ErrorContains(t, err, "more than 10%")
	require.Equal(t, "7.21", repo.values[SettingKeyCurrencyUSDToCNYAutoRate])
	require.NotEmpty(t, repo.values[SettingKeyCurrencyExchangeRateLastError])
}
