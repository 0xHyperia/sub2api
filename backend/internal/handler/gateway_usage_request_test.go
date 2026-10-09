package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type usageRequestRepoStub struct {
	service.UsageLogRepository
	ownerKeyID int64
	storedID   string
	appearAt   int // 第几次查询开始可见；0 表示一直不可见
	calls      int
}

func (r *usageRequestRepoStub) GetByRequestIDForAPIKey(_ context.Context, apiKeyID int64, ids []string) (*service.UsageLog, error) {
	r.calls++
	if r.appearAt == 0 || r.calls < r.appearAt || apiKeyID != r.ownerKeyID || !slices.Contains(ids, r.storedID) {
		return nil, service.ErrUsageLogNotFound
	}
	return &service.UsageLog{
		RequestID:    r.storedID,
		Model:        "gpt-test",
		InputTokens:  10,
		OutputTokens: 5,
		ActualCost:   0.0123,
		CreatedAt:    time.Date(2026, 10, 9, 8, 0, 0, 0, time.UTC),
	}, nil
}

func callUsageByRequestID(t *testing.T, repo *usageRequestRepoStub, keyID int64, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &GatewayHandler{usageService: service.NewUsageService(repo, nil, nil, nil)}
	router := gin.New()
	router.GET("/v1/usage/requests/:request_id", func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{ID: keyID})
		h.UsageByRequestID(c)
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestUsageByRequestIDReturnsBilledCost(t *testing.T) {
	repo := &usageRequestRepoStub{ownerKeyID: 5, storedID: "client:req-1", appearAt: 1}
	w := callUsageByRequestID(t, repo, 5, "/v1/usage/requests/req-1?wait=0")

	require.Equal(t, http.StatusOK, w.Code)
	var resp usageRequestResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "usage_request", resp.Object)
	require.Equal(t, "req-1", resp.RequestID)
	require.Equal(t, "billed", resp.Status)
	require.Equal(t, "USD", resp.Unit)
	require.InDelta(t, 0.0123, resp.ActualCost, 1e-12)
	require.Equal(t, 10, resp.Usage.InputTokens)
	require.Equal(t, 5, resp.Usage.OutputTokens)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
}

func TestUsageByRequestIDWaitsForAsyncBilling(t *testing.T) {
	repo := &usageRequestRepoStub{ownerKeyID: 5, storedID: "client:req-2", appearAt: 3}
	w := callUsageByRequestID(t, repo, 5, "/v1/usage/requests/req-2?wait=5")

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, 3, repo.calls)
}

func TestUsageByRequestIDNotFoundAfterWait(t *testing.T) {
	repo := &usageRequestRepoStub{ownerKeyID: 5, storedID: "client:req-3"}
	w := callUsageByRequestID(t, repo, 5, "/v1/usage/requests/req-3?wait=0")

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, 1, repo.calls)
}

func TestUsageByRequestIDIsScopedToAPIKey(t *testing.T) {
	repo := &usageRequestRepoStub{ownerKeyID: 5, storedID: "client:req-4", appearAt: 1}
	w := callUsageByRequestID(t, repo, 6, "/v1/usage/requests/req-4?wait=0")

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUsageByRequestIDRejectsInvalidWait(t *testing.T) {
	for _, wait := range []string{"-1", "11", "abc"} {
		repo := &usageRequestRepoStub{ownerKeyID: 5, storedID: "client:req-5", appearAt: 1}
		w := callUsageByRequestID(t, repo, 5, "/v1/usage/requests/req-5?wait="+wait)
		require.Equal(t, http.StatusBadRequest, w.Code, wait)
		require.Zero(t, repo.calls, wait)
	}
}
