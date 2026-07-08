package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

const (
	ldxpDefaultAPIBase   = "https://pay.ldxp.cn"
	ldxpHTTPTimeout      = 10 * time.Second
	maxLdxpResponseSize  = 1 << 20
	ldxpCodeSuccess      = 1
	ldxpOrderStatusPaid  = 1
	ldxpOrderSendoutDone = 1
)

// Ldxp 实现链动小铺店铺级发卡支付能力。
type Ldxp struct {
	instanceID string
	config     map[string]string
	httpClient *http.Client
	shopToken  string
}

func NewLdxp(instanceID string, config map[string]string) (*Ldxp, error) {
	cfg := cloneStringMap(config)
	if strings.TrimSpace(cfg["apiBase"]) == "" {
		cfg["apiBase"] = ldxpDefaultAPIBase
	}
	cfg["apiBase"] = normalizeLdxpAPIBase(cfg["apiBase"])
	token, err := ParseLdxpShopToken(cfg["shopUrl"])
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("ldxp config missing required key: shopUrl")
	}
	return &Ldxp{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: ldxpHTTPTimeout},
		shopToken:  token,
	}, nil
}

func ParseLdxpShopToken(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if !strings.Contains(raw, "://") && !strings.Contains(raw, "/") {
		return raw, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("ldxp shopUrl is invalid")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, part := range parts {
		if part == "shop" && i+1 < len(parts) && strings.TrimSpace(parts[i+1]) != "" {
			return strings.TrimSpace(parts[i+1]), nil
		}
	}
	return "", fmt.Errorf("ldxp shopUrl must look like https://pay.ldxp.cn/shop/{token}")
}

func normalizeLdxpAPIBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return ldxpDefaultAPIBase
	}
	parsed, err := url.Parse(base)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		parsed.RawQuery = ""
		parsed.Fragment = ""
		parsed.RawPath = ""
		parsed.Path = strings.TrimRight(parsed.Path, "/")
		return strings.TrimRight(parsed.String(), "/")
	}
	return strings.TrimRight(base, "/")
}

func (l *Ldxp) Name() string        { return "链动小铺" }
func (l *Ldxp) ProviderKey() string { return payment.TypeLdxp }
func (l *Ldxp) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeAlipay}
}

func (l *Ldxp) MerchantIdentityMetadata() map[string]string {
	if l.shopToken == "" {
		return nil
	}
	return map[string]string{"shop_token": l.shopToken}
}

func (l *Ldxp) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	channelID, _ := strconv.Atoi(strings.TrimSpace(l.config["channelId"]))
	goodsKey := strings.TrimSpace(l.config["goodsKey"])
	if goodsKey == "" || channelID <= 0 {
		return nil, fmt.Errorf("ldxp immediate payment requires goodsKey and channelId")
	}
	body := map[string]any{
		"goods_key":        goodsKey,
		"quantity":         ldxpConfigInt(l.config["quantity"], 1),
		"coupon_code":      strings.TrimSpace(l.config["couponCode"]),
		"channel_id":       channelID,
		"contact":          ldxpContact(req, l.config),
		"query_password":   strings.TrimSpace(l.config["queryPassword"]),
		"select_cards_ids": []string{},
		"extend": map[string]string{
			"sub2api_order_id": req.OrderID,
		},
	}
	raw, err := l.postJSON(ctx, "/shopApi/Pay/order", body)
	if err != nil {
		return nil, fmt.Errorf("ldxp create: %w", err)
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TradeNo     string  `json:"trade_no"`
			TotalAmount float64 `json:"total_amount"`
			PayURL      string  `json:"payurl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse create: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp create failed: %s", strings.TrimSpace(resp.Msg))
	}
	if strings.TrimSpace(resp.Data.TradeNo) == "" || strings.TrimSpace(resp.Data.PayURL) == "" {
		return nil, fmt.Errorf("ldxp create returned incomplete payment data")
	}
	return &payment.CreatePaymentResponse{
		TradeNo: resp.Data.TradeNo,
		PayURL:  resp.Data.PayURL,
	}, nil
}

type LdxpShopInfo struct {
	Token       string `json:"token"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

type LdxpCategory struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	GoodsCount int    `json:"goods_count"`
}

type LdxpGoods struct {
	ProviderInstanceID    string  `json:"provider_instance_id"`
	ShopToken             string  `json:"shop_token"`
	GoodsKey              string  `json:"goods_key"`
	Name                  string  `json:"name"`
	Price                 float64 `json:"price"`
	Description           string  `json:"description"`
	Image                 string  `json:"image"`
	CategoryID            int64   `json:"category_id"`
	CategoryName          string  `json:"category_name"`
	StockCount            int     `json:"stock_count"`
	LimitCount            int     `json:"limit_count"`
	QueryPasswordRequired bool    `json:"query_password_required"`
}

type LdxpChannel struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ShowName string `json:"show_name"`
	Code     string `json:"code"`
	Icon     string `json:"icon"`
	Status   int    `json:"status"`
}

type LdxpPrice struct {
	OriginalAmount float64 `json:"original_amount"`
	TotalAmount    float64 `json:"total_amount"`
	Fee            float64 `json:"fee"`
}

type LdxpCardOrderRequest struct {
	GoodsKey      string
	Quantity      int
	ChannelID     int64
	CouponCode    string
	Contact       string
	QueryPassword string
	OrderID       string
	VisitorID     string
}

type LdxpCardOrderResponse struct {
	TradeNo     string  `json:"trade_no"`
	TotalAmount float64 `json:"total_amount"`
	PayURL      string  `json:"payurl"`
}

type LdxpOrderInfo struct {
	TradeNo     string   `json:"trade_no"`
	GoodsName   string   `json:"goods_name"`
	Quantity    int      `json:"quantity"`
	TotalAmount float64  `json:"total_amount"`
	Status      int      `json:"status"`
	Sendout     int      `json:"sendout"`
	Cards       []string `json:"cards"`
}

func (l *Ldxp) ShopToken() string { return l.shopToken }

func (l *Ldxp) GetShopInfo(ctx context.Context) (*LdxpShopInfo, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Shop/info", map[string]any{"token": l.shopToken, "category_key": ""})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code int          `json:"code"`
		Msg  string       `json:"msg"`
		Data LdxpShopInfo `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse shop info: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp shop info failed: %s", strings.TrimSpace(resp.Msg))
	}
	resp.Data.Token = firstNonEmpty(resp.Data.Token, l.shopToken)
	resp.Data.Link = firstNonEmpty(resp.Data.Link, l.config["shopUrl"])
	return &resp.Data, nil
}

func (l *Ldxp) ListCategories(ctx context.Context) ([]LdxpCategory, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Shop/categoryList", map[string]any{"token": l.shopToken, "goods_type": "card", "category_key": ""})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data []LdxpCategory `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse categories: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp categories failed: %s", strings.TrimSpace(resp.Msg))
	}
	return resp.Data, nil
}

func (l *Ldxp) ListGoods(ctx context.Context, categoryID int64) ([]LdxpGoods, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Shop/goodsList", map[string]any{
		"token": l.shopToken, "keywords": "", "category_id": categoryID,
		"goods_type": "card", "current": 1, "pageSize": 100,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			List []struct {
				GoodsKey    string  `json:"goods_key"`
				Name        string  `json:"name"`
				Price       float64 `json:"price"`
				Description string  `json:"description"`
				Image       string  `json:"image"`
				Category    struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				} `json:"category"`
				Extend struct {
					StockCount          int `json:"stock_count"`
					LimitCount          int `json:"limit_count"`
					QueryPasswordStatus int `json:"query_password_status"`
				} `json:"extend"`
			} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse goods: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp goods failed: %s", strings.TrimSpace(resp.Msg))
	}
	out := make([]LdxpGoods, 0, len(resp.Data.List))
	for _, item := range resp.Data.List {
		out = append(out, LdxpGoods{
			ProviderInstanceID:    l.instanceID,
			ShopToken:             l.shopToken,
			GoodsKey:              item.GoodsKey,
			Name:                  item.Name,
			Price:                 item.Price,
			Description:           item.Description,
			Image:                 item.Image,
			CategoryID:            item.Category.ID,
			CategoryName:          item.Category.Name,
			StockCount:            item.Extend.StockCount,
			LimitCount:            item.Extend.LimitCount,
			QueryPasswordRequired: item.Extend.QueryPasswordStatus == 1,
		})
	}
	return out, nil
}

func (l *Ldxp) ListChannels(ctx context.Context) ([]LdxpChannel, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Shop/getUserChannel", map[string]any{"token": l.shopToken})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			ShowName string `json:"show_name"`
			Code     string `json:"code"`
			Status   int    `json:"status"`
			PayType  struct {
				Icon string `json:"icon"`
			} `json:"paytype"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse channels: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp channels failed: %s", strings.TrimSpace(resp.Msg))
	}
	out := make([]LdxpChannel, 0, len(resp.Data))
	for _, item := range resp.Data {
		out = append(out, LdxpChannel{ID: item.ID, Name: item.Name, ShowName: item.ShowName, Code: item.Code, Icon: item.PayType.Icon, Status: item.Status})
	}
	return out, nil
}

func (l *Ldxp) GetGoodsPrice(ctx context.Context, goodsKey string, quantity int, couponCode string, channelID int64) (*LdxpPrice, error) {
	if quantity <= 0 {
		quantity = 1
	}
	raw, err := l.postJSON(ctx, "/shopApi/Shop/getGoodsPrice", map[string]any{
		"goods_key": goodsKey, "quantity": quantity, "coupon_code": couponCode, "channel_id": channelID,
	})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Code int       `json:"code"`
		Msg  string    `json:"msg"`
		Data LdxpPrice `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse price: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp price failed: %s", strings.TrimSpace(resp.Msg))
	}
	return &resp.Data, nil
}

func (l *Ldxp) CreateCardOrder(ctx context.Context, req LdxpCardOrderRequest) (*LdxpCardOrderResponse, error) {
	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	queryPassword := strings.TrimSpace(req.QueryPassword)
	if queryPassword == "" {
		queryPassword = strings.TrimSpace(l.config["queryPassword"])
	}
	visitorID := strings.TrimSpace(req.VisitorID)
	if visitorID == "" {
		visitorID = strings.TrimSpace(l.config["visitorId"])
	}
	body := map[string]any{
		"goods_key":        strings.TrimSpace(req.GoodsKey),
		"quantity":         quantity,
		"coupon_code":      strings.TrimSpace(req.CouponCode),
		"channel_id":       req.ChannelID,
		"contact":          ldxpOrderContact(req.Contact, req.OrderID, l.config),
		"query_password":   queryPassword,
		"select_cards_ids": []string{},
		"extend":           map[string]string{"sub2api_order_id": req.OrderID, "juuid": visitorID},
	}
	raw, err := l.postJSON(ctx, "/shopApi/Pay/order", body)
	if err != nil {
		return nil, fmt.Errorf("ldxp create card order: %w", err)
	}
	var resp struct {
		Code int                   `json:"code"`
		Msg  string                `json:"msg"`
		Data LdxpCardOrderResponse `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse card order: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp card order failed: %s", strings.TrimSpace(resp.Msg))
	}
	if strings.TrimSpace(resp.Data.TradeNo) == "" || strings.TrimSpace(resp.Data.PayURL) == "" {
		return nil, fmt.Errorf("ldxp card order returned incomplete payment data")
	}
	return &resp.Data, nil
}

func (l *Ldxp) GetOrderInfoWithCards(ctx context.Context, tradeNo string, queryPassword ...string) (*LdxpOrderInfo, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Order/info", l.orderInfoPayload(tradeNo, firstNonEmpty(queryPassword...)))
	if err != nil {
		return nil, fmt.Errorf("ldxp order info: %w", err)
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TradeNo     string  `json:"trade_no"`
			GoodsName   string  `json:"goods_name"`
			Quantity    int     `json:"quantity"`
			TotalAmount float64 `json:"total_amount"`
			Status      int     `json:"status"`
			Sendout     int     `json:"sendout"`
			Response    struct {
				Cards []string `json:"cards"`
			} `json:"response"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse order info: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp order info failed: %s", strings.TrimSpace(resp.Msg))
	}
	return &LdxpOrderInfo{
		TradeNo: resp.Data.TradeNo, GoodsName: resp.Data.GoodsName, Quantity: resp.Data.Quantity,
		TotalAmount: resp.Data.TotalAmount, Status: resp.Data.Status, Sendout: resp.Data.Sendout,
		Cards: resp.Data.Response.Cards,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func ldxpContact(req payment.CreatePaymentRequest, cfg map[string]string) string {
	return ldxpOrderContact(req.Contact, req.OrderID, cfg)
}

func ldxpOrderContact(contact, orderID string, cfg map[string]string) string {
	for _, candidate := range []string{contact, cfg["contactFallback"], orderID} {
		if value := strings.TrimSpace(candidate); value != "" {
			return value
		}
	}
	return "sub2api"
}

func ldxpConfigInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (l *Ldxp) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	return l.QueryOrderWithQueryPassword(ctx, tradeNo, strings.TrimSpace(l.config["queryPassword"]))
}

func (l *Ldxp) QueryOrderWithQueryPassword(ctx context.Context, tradeNo string, queryPassword string) (*payment.QueryOrderResponse, error) {
	tradeNo = strings.TrimSpace(tradeNo)
	if tradeNo == "" {
		return nil, fmt.Errorf("ldxp query requires trade_no")
	}

	raw, err := l.postJSON(ctx, "/shopApi/Pay/query", map[string]string{"trade_no": tradeNo})
	if err != nil {
		return nil, fmt.Errorf("ldxp query: %w", err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return &payment.QueryOrderResponse{TradeNo: tradeNo, Status: payment.ProviderStatusPending}, nil
	}
	var queryResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data any    `json:"data"`
	}
	if err := json.Unmarshal(raw, &queryResp); err != nil {
		return nil, fmt.Errorf("ldxp parse query: %w", err)
	}
	if queryResp.Code != ldxpCodeSuccess {
		return &payment.QueryOrderResponse{
			TradeNo:  tradeNo,
			Status:   payment.ProviderStatusPending,
			Metadata: l.MerchantIdentityMetadata(),
		}, nil
	}

	info, err := l.queryOrderInfo(ctx, tradeNo, queryPassword)
	if err != nil {
		return nil, err
	}
	status := payment.ProviderStatusPending
	if info.Status == ldxpOrderStatusPaid || info.Sendout == ldxpOrderSendoutDone {
		status = payment.ProviderStatusPaid
	}
	return &payment.QueryOrderResponse{
		TradeNo:  tradeNo,
		Status:   status,
		Amount:   info.TotalAmount,
		Metadata: l.MerchantIdentityMetadata(),
	}, nil
}

type ldxpOrderInfo struct {
	TotalAmount float64
	Status      int
	Sendout     int
}

func (l *Ldxp) queryOrderInfo(ctx context.Context, tradeNo string, queryPassword string) (*ldxpOrderInfo, error) {
	raw, err := l.postJSON(ctx, "/shopApi/Order/info", l.orderInfoPayload(tradeNo, queryPassword))
	if err != nil {
		return nil, fmt.Errorf("ldxp order info: %w", err)
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			TotalAmount float64 `json:"total_amount"`
			Status      int     `json:"status"`
			Sendout     int     `json:"sendout"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("ldxp parse order info: %w", err)
	}
	if resp.Code != ldxpCodeSuccess {
		return nil, fmt.Errorf("ldxp order info failed: %s", strings.TrimSpace(resp.Msg))
	}
	return &ldxpOrderInfo{
		TotalAmount: resp.Data.TotalAmount,
		Status:      resp.Data.Status,
		Sendout:     resp.Data.Sendout,
	}, nil
}

func (l *Ldxp) orderInfoPayload(tradeNo string, queryPassword string) map[string]any {
	payload := map[string]any{
		"trade_no": strings.TrimSpace(tradeNo),
		"dump":     1,
	}
	if queryPassword = strings.TrimSpace(queryPassword); queryPassword != "" {
		payload["query_password"] = queryPassword
	}
	return payload
}

func (l *Ldxp) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, fmt.Errorf("ldxp webhook is not supported")
}

func (l *Ldxp) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, fmt.Errorf("ldxp refund is not supported")
}

func (l *Ldxp) postJSON(ctx context.Context, path string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, l.config["apiBase"]+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36")
	if origin := l.origin(); origin != "" {
		req.Header.Set("Origin", origin)
	}
	if referer := strings.TrimSpace(l.config["referer"]); referer != "" {
		req.Header.Set("Referer", referer)
	} else if shopURL := strings.TrimSpace(l.config["shopUrl"]); shopURL != "" {
		req.Header.Set("Referer", shopURL)
	}
	if visitorID := strings.TrimSpace(l.config["visitorId"]); visitorID != "" {
		req.Header.Set("visitorid", visitorID)
	}
	client := l.httpClient
	if client == nil {
		client = &http.Client{Timeout: ldxpHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLdxpResponseSize))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, summarizeEasyPayResponse(body))
	}
	return body, nil
}

func (l *Ldxp) origin() string {
	parsed, err := url.Parse(strings.TrimSpace(l.config["apiBase"]))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
