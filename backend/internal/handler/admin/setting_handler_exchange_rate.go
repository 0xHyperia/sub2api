package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// SyncExchangeRate immediately refreshes the reusable USD/CNY benchmark.
// POST /api/v1/admin/settings/exchange-rate/sync
func (h *SettingHandler) SyncExchangeRate(c *gin.Context) {
	settings, err := h.settingService.SyncExchangeRate(c.Request.Context(), true)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"currency_usd_to_cny_manual_rate":          settings.ManualUSDToCNY,
		"currency_exchange_rate_auto_sync_enabled": settings.AutoSync,
		"currency_usd_to_cny_auto_rate":            settings.AutoUSDToCNY,
		"currency_usd_to_cny_effective_rate":       settings.Effective,
		"currency_exchange_rate_source":            settings.Source,
		"currency_exchange_rate_provider":          settings.Provider,
		"currency_exchange_rate_provider_as_of":    settings.ProviderAsOf,
		"currency_exchange_rate_last_synced_at":    settings.LastSyncedAt,
		"currency_exchange_rate_last_error":        settings.LastError,
		"currency_exchange_rate_stale":             settings.Stale,
	})
}
