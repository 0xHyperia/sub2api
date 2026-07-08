package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestNewLdxpValidatesConfig(t *testing.T) {
	t.Parallel()

	if _, err := NewLdxp("1", map[string]string{"channelId": "1"}); err == nil {
		t.Fatal("NewLdxp missing shopUrl returned nil error")
	}
	if _, err := NewLdxp("1", map[string]string{"shopUrl": "https://pay.ldxp.cn/item/abc"}); err == nil {
		t.Fatal("NewLdxp invalid shopUrl returned nil error")
	}

	prov, err := NewLdxp("1", map[string]string{"shopUrl": "https://pay.ldxp.cn/shop/FWW9YE0U"})
	if err != nil {
		t.Fatalf("NewLdxp returned error: %v", err)
	}
	if prov.config["apiBase"] != ldxpDefaultAPIBase {
		t.Fatalf("apiBase = %q, want %q", prov.config["apiBase"], ldxpDefaultAPIBase)
	}
	if prov.ProviderKey() != payment.TypeLdxp {
		t.Fatalf("ProviderKey = %q, want %q", prov.ProviderKey(), payment.TypeLdxp)
	}
}

func TestLdxpCreatePayment(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotPayload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ReadAll: %v", err)
		}
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Errorf("Unmarshal request: %v", err)
		}
		if got := r.Header.Get("Referer"); got != "https://shop.example.com" {
			t.Errorf("Referer = %q, want configured referer", got)
		}
		if got := r.Header.Get("visitorid"); got != "visitor-1" {
			t.Errorf("visitorid = %q, want configured visitor id", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"trade_no":"ldxp-123","payurl":"https://pay.example.com/pay/ldxp-123","total_amount":12.34}}`))
	}))
	defer server.Close()

	prov := newTestLdxp(t, server.URL)
	resp, err := prov.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID: "sub2-123",
		Contact: "buyer@example.com",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}
	if resp.TradeNo != "ldxp-123" || resp.PayURL != "https://pay.example.com/pay/ldxp-123" {
		t.Fatalf("CreatePayment response = %+v", resp)
	}
	if gotPath != "/shopApi/Pay/order" {
		t.Fatalf("path = %q, want /shopApi/Pay/order", gotPath)
	}
	if gotPayload["goods_key"] != "goods-1" {
		t.Fatalf("goods_key = %v, want goods-1", gotPayload["goods_key"])
	}
	if gotPayload["channel_id"] != float64(7) {
		t.Fatalf("channel_id = %v, want 7", gotPayload["channel_id"])
	}
	if gotPayload["contact"] != "buyer@example.com" {
		t.Fatalf("contact = %v, want buyer@example.com", gotPayload["contact"])
	}
	if gotPayload["quantity"] != float64(2) {
		t.Fatalf("quantity = %v, want 2", gotPayload["quantity"])
	}
}

func TestLdxpQueryOrderPaid(t *testing.T) {
	t.Parallel()

	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/shopApi/Pay/query":
			_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"trade_no":"ldxp-123"}}`))
		case "/shopApi/Order/info":
			_, _ = w.Write([]byte(`{"code":1,"msg":"ok","data":{"total_amount":12.34,"status":1,"sendout":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	prov := newTestLdxp(t, server.URL)
	resp, err := prov.QueryOrder(context.Background(), "ldxp-123")
	if err != nil {
		t.Fatalf("QueryOrder returned error: %v", err)
	}
	if resp.Status != payment.ProviderStatusPaid {
		t.Fatalf("status = %q, want %q", resp.Status, payment.ProviderStatusPaid)
	}
	if resp.Amount != 12.34 {
		t.Fatalf("amount = %v, want 12.34", resp.Amount)
	}
	if resp.Metadata["shop_token"] != "FWW9YE0U" {
		t.Fatalf("metadata shop_token = %q, want FWW9YE0U", resp.Metadata["shop_token"])
	}
	if len(paths) != 2 || paths[0] != "/shopApi/Pay/query" || paths[1] != "/shopApi/Order/info" {
		t.Fatalf("paths = %v, want query then order info", paths)
	}
}

func TestLdxpQueryOrderPending(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"订单不存在"}`))
	}))
	defer server.Close()

	prov := newTestLdxp(t, server.URL)
	resp, err := prov.QueryOrder(context.Background(), "ldxp-404")
	if err != nil {
		t.Fatalf("QueryOrder returned error: %v", err)
	}
	if resp.Status != payment.ProviderStatusPending {
		t.Fatalf("status = %q, want %q", resp.Status, payment.ProviderStatusPending)
	}
}

func newTestLdxp(t *testing.T, apiBase string) *Ldxp {
	t.Helper()

	prov, err := NewLdxp("1", map[string]string{
		"apiBase":   apiBase + "/",
		"shopUrl":   "https://pay.ldxp.cn/shop/FWW9YE0U",
		"goodsKey":  "goods-1",
		"channelId": "7",
		"quantity":  "2",
		"referer":   "https://shop.example.com",
		"visitorId": "visitor-1",
	})
	if err != nil {
		t.Fatalf("NewLdxp returned error: %v", err)
	}
	return prov
}
