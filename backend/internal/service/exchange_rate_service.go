package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultUSDToCNYRate       = 7.2
	exchangeRateProvider      = "Frankfurter / ECB"
	exchangeRateSyncInterval  = 6 * time.Hour
	exchangeRateStaleAfter    = 12 * time.Hour
	exchangeRateMaxChangeRate = 0.10
)

type ExchangeRateSettings struct {
	ManualUSDToCNY float64
	AutoSync       bool
	AutoUSDToCNY   float64
	Effective      float64
	Source         string
	Provider       string
	ProviderAsOf   string
	LastSyncedAt   string
	LastError      string
	Stale          bool
}

func validUSDToCNYRate(rate float64) bool {
	return !math.IsNaN(rate) && !math.IsInf(rate, 0) && rate > 0 && rate < 100
}

func parseValidUSDToCNYRate(raw string) float64 {
	rate, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || !validUSDToCNYRate(rate) {
		return 0
	}
	return rate
}

func formatUSDToCNYRate(rate float64) string {
	return strconv.FormatFloat(rate, 'f', 6, 64)
}

func (s *SettingService) GetExchangeRateSettings(ctx context.Context) (*ExchangeRateSettings, error) {
	keys := []string{
		SettingKeyCurrencyUSDToCNYManualRate,
		SettingKeyCurrencyExchangeRateAutoSync,
		SettingKeyCurrencyUSDToCNYAutoRate,
		SettingKeyCurrencyExchangeRateProvider,
		SettingKeyCurrencyExchangeRateProviderAsOf,
		SettingKeyCurrencyExchangeRateLastSyncedAt,
		SettingKeyCurrencyExchangeRateLastError,
		SettingSubscriptionUSDToCNYRate,
	}
	values, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("get exchange rate settings: %w", err)
	}
	result := parseExchangeRateSettingsMap(values)
	return &result, nil
}

func parseExchangeRateSettingsMap(values map[string]string) ExchangeRateSettings {
	manual := parseValidUSDToCNYRate(values[SettingKeyCurrencyUSDToCNYManualRate])
	if manual == 0 {
		manual = parseValidUSDToCNYRate(values[SettingSubscriptionUSDToCNYRate])
	}
	if manual == 0 {
		manual = DefaultUSDToCNYRate
	}
	auto := parseValidUSDToCNYRate(values[SettingKeyCurrencyUSDToCNYAutoRate])
	autoEnabled := values[SettingKeyCurrencyExchangeRateAutoSync] == "true"
	effective, source := manual, "manual"
	if autoEnabled && auto > 0 {
		effective, source = auto, "auto"
	}
	lastSynced := strings.TrimSpace(values[SettingKeyCurrencyExchangeRateLastSyncedAt])
	stale := false
	if autoEnabled {
		parsed, parseErr := time.Parse(time.RFC3339, lastSynced)
		stale = auto == 0 || parseErr != nil || time.Since(parsed) > exchangeRateStaleAfter
	}
	provider := strings.TrimSpace(values[SettingKeyCurrencyExchangeRateProvider])
	if provider == "" {
		provider = exchangeRateProvider
	}
	return ExchangeRateSettings{
		ManualUSDToCNY: manual,
		AutoSync:       autoEnabled,
		AutoUSDToCNY:   auto,
		Effective:      effective,
		Source:         source,
		Provider:       provider,
		ProviderAsOf:   strings.TrimSpace(values[SettingKeyCurrencyExchangeRateProviderAsOf]),
		LastSyncedAt:   lastSynced,
		LastError:      strings.TrimSpace(values[SettingKeyCurrencyExchangeRateLastError]),
		Stale:          stale,
	}
}

func (s *SettingService) EffectiveUSDToCNY(ctx context.Context) (float64, error) {
	settings, err := s.GetExchangeRateSettings(ctx)
	if err != nil {
		return 0, err
	}
	return settings.Effective, nil
}

// EnsureExchangeRateSettings materializes the legacy subscription rate once so
// the reusable benchmark becomes independent from payment configuration.
func (s *SettingService) EnsureExchangeRateSettings(ctx context.Context) error {
	manual, err := s.settingRepo.GetValue(ctx, SettingKeyCurrencyUSDToCNYManualRate)
	if err == nil && parseValidUSDToCNYRate(manual) > 0 {
		return nil
	}
	if err != nil && !errors.Is(err, ErrSettingNotFound) {
		return fmt.Errorf("read exchange rate migration state: %w", err)
	}
	legacy, legacyErr := s.settingRepo.GetValue(ctx, SettingSubscriptionUSDToCNYRate)
	rate := parseValidUSDToCNYRate(legacy)
	if legacyErr != nil && !errors.Is(legacyErr, ErrSettingNotFound) {
		return fmt.Errorf("read legacy exchange rate: %w", legacyErr)
	}
	if rate == 0 {
		rate = DefaultUSDToCNYRate
	}
	return s.settingRepo.SetMultiple(ctx, map[string]string{
		SettingKeyCurrencyUSDToCNYManualRate:   formatUSDToCNYRate(rate),
		SettingKeyCurrencyExchangeRateAutoSync: "false",
		SettingKeyCurrencyExchangeRateProvider: exchangeRateProvider,
	})
}

func (s *SettingService) SyncExchangeRate(ctx context.Context, force bool) (*ExchangeRateSettings, error) {
	value, err, _ := s.exchangeRateSyncSF.Do("usd-cny", func() (any, error) {
		current, getErr := s.GetExchangeRateSettings(ctx)
		if getErr != nil {
			return nil, getErr
		}
		if !current.AutoSync && !force {
			return current, nil
		}
		if !force && current.LastSyncedAt != "" {
			if syncedAt, parseErr := time.Parse(time.RFC3339, current.LastSyncedAt); parseErr == nil && time.Since(syncedAt) < exchangeRateSyncInterval {
				return current, nil
			}
		}
		rate, asOf, fetchErr := s.fetchUSDToCNYRate(ctx)
		if fetchErr != nil {
			_ = s.settingRepo.Set(ctx, SettingKeyCurrencyExchangeRateLastError, fetchErr.Error())
			return nil, fetchErr
		}
		baseline := current.Effective
		if baseline > 0 && math.Abs(rate-baseline)/baseline > exchangeRateMaxChangeRate {
			jumpErr := fmt.Errorf("provider rate %.6f differs from current rate %.6f by more than 10%%", rate, baseline)
			_ = s.settingRepo.Set(ctx, SettingKeyCurrencyExchangeRateLastError, jumpErr.Error())
			return nil, jumpErr
		}
		now := time.Now().UTC().Format(time.RFC3339)
		if setErr := s.settingRepo.SetMultiple(ctx, map[string]string{
			SettingKeyCurrencyUSDToCNYAutoRate:         formatUSDToCNYRate(rate),
			SettingKeyCurrencyExchangeRateProvider:     exchangeRateProvider,
			SettingKeyCurrencyExchangeRateProviderAsOf: asOf,
			SettingKeyCurrencyExchangeRateLastSyncedAt: now,
			SettingKeyCurrencyExchangeRateLastError:    "",
		}); setErr != nil {
			return nil, fmt.Errorf("store exchange rate: %w", setErr)
		}
		return s.GetExchangeRateSettings(ctx)
	})
	if err != nil {
		return nil, err
	}
	return value.(*ExchangeRateSettings), nil
}

func (s *SettingService) fetchUSDToCNYRate(ctx context.Context) (float64, string, error) {
	client := s.exchangeRateHTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.exchangeRateEndpoint, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("exchange rate provider returned HTTP %d", resp.StatusCode)
	}
	var payload struct {
		Date  string             `json:"date"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, "", fmt.Errorf("decode exchange rate: %w", err)
	}
	rate := payload.Rates["CNY"]
	if !validUSDToCNYRate(rate) {
		return 0, "", errors.New("exchange rate provider returned an invalid USD/CNY rate")
	}
	return rate, strings.TrimSpace(payload.Date), nil
}

type ExchangeRateRunner struct {
	service *SettingService
	stopCh  chan struct{}
	doneCh  chan struct{}
}

func NewExchangeRateRunner(service *SettingService) *ExchangeRateRunner {
	return &ExchangeRateRunner{service: service, stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

func (r *ExchangeRateRunner) Start() {
	go func() {
		defer close(r.doneCh)
		r.check()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.check()
			case <-r.stopCh:
				return
			}
		}
	}()
}

func (r *ExchangeRateRunner) check() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := r.service.EnsureExchangeRateSettings(ctx); err != nil {
		return
	}
	_, _ = r.service.SyncExchangeRate(ctx, false)
}

func (r *ExchangeRateRunner) Stop() {
	select {
	case <-r.stopCh:
		return
	default:
		close(r.stopCh)
		<-r.doneCh
	}
}
