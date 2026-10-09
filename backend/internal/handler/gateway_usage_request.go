package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	usageRequestObject         = "usage_request"
	usageRequestDefaultWaitSec = 3
	usageRequestMaxWaitSec     = 10
	usageRequestPollInterval   = 150 * time.Millisecond
	usageRequestMaxIDBytes     = 128
)

type usageRequestTokens struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	CacheCreationTokens int `json:"cache_creation_tokens"`
	CacheReadTokens     int `json:"cache_read_tokens"`
}

type usageRequestResponse struct {
	Object     string             `json:"object"`
	RequestID  string             `json:"request_id"`
	Status     string             `json:"status"`
	Model      string             `json:"model"`
	ActualCost float64            `json:"actual_cost"`
	Unit       string             `json:"unit"`
	Usage      usageRequestTokens `json:"usage"`
	CreatedAt  time.Time          `json:"created_at"`
}

// UsageByRequestID returns the billed cost of one request made with the authenticated API key.
// GET /v1/usage/requests/:request_id?wait=<seconds>
//
// 计费是异步的：用量记录在响应结束后才写入。wait 让服务端在记录出现前短轮询，
// 下游一次调用即可拿到结果；等待超时仍无记录返回 404。
func (h *GatewayHandler) UsageByRequestID(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		h.errorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	if h.usageService == nil {
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Usage lookup is unavailable")
		return
	}

	requestID := strings.TrimSpace(c.Param("request_id"))
	if requestID == "" || len(requestID) > usageRequestMaxIDBytes {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid request_id")
		return
	}
	waitSec, ok := parseUsageRequestWait(c.Query("wait"))
	if !ok {
		h.errorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid wait, allowed range is 0-10 seconds")
		return
	}

	ctx := c.Request.Context()
	log, err := h.lookupUsageByRequestID(ctx, apiKey.ID, requestID, time.Duration(waitSec)*time.Second)
	if err != nil {
		if errors.Is(err, service.ErrUsageLogNotFound) {
			h.errorResponse(c, http.StatusNotFound, "not_found_error", "No billed usage found for this request")
			return
		}
		if ctx.Err() != nil {
			return
		}
		h.errorResponse(c, http.StatusInternalServerError, "api_error", "Failed to look up usage")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, buildUsageRequestResponse(requestID, log))
}

func (h *GatewayHandler) lookupUsageByRequestID(ctx context.Context, apiKeyID int64, requestID string, wait time.Duration) (*service.UsageLog, error) {
	deadline := time.Now().Add(wait)
	for {
		log, err := h.usageService.GetByRequestIDForAPIKey(ctx, apiKeyID, requestID)
		if err == nil {
			return log, nil
		}
		if !errors.Is(err, service.ErrUsageLogNotFound) {
			return nil, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return nil, err
		}
		timer := time.NewTimer(min(usageRequestPollInterval, remaining))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func parseUsageRequestWait(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return usageRequestDefaultWaitSec, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > usageRequestMaxWaitSec {
		return 0, false
	}
	return n, true
}

func buildUsageRequestResponse(requestID string, log *service.UsageLog) usageRequestResponse {
	return usageRequestResponse{
		Object:     usageRequestObject,
		RequestID:  requestID,
		Status:     "billed",
		Model:      log.Model,
		ActualCost: log.ActualCost,
		Unit:       "USD",
		Usage: usageRequestTokens{
			InputTokens:         log.InputTokens,
			OutputTokens:        log.OutputTokens,
			CacheCreationTokens: log.CacheCreationTokens,
			CacheReadTokens:     log.CacheReadTokens,
		},
		CreatedAt: log.CreatedAt.UTC(),
	}
}
