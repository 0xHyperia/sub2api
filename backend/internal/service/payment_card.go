package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/payment/provider"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

type CardCheckoutInfo struct {
	Shops []CardCheckoutShop `json:"shops"`
}

type CardCheckoutShop struct {
	ProviderInstanceID string                  `json:"provider_instance_id"`
	Name               string                  `json:"name"`
	Shop               *provider.LdxpShopInfo  `json:"shop"`
	Categories         []provider.LdxpCategory `json:"categories"`
	Goods              []provider.LdxpGoods    `json:"goods"`
	Channels           []provider.LdxpChannel  `json:"channels"`
}

type CardPriceRequest struct {
	ProviderInstanceID string `json:"provider_instance_id"`
	GoodsKey           string `json:"goods_key"`
	Quantity           int    `json:"quantity"`
	ChannelID          int64  `json:"channel_id"`
	CouponCode         string `json:"coupon_code"`
}

type CardOrderRequest struct {
	UserID             int64
	ProviderInstanceID string
	GoodsKey           string
	Quantity           int
	ChannelID          int64
	CouponCode         string
	Contact            string
	QueryPassword      string
	AutoRedeem         *bool
	ReturnURL          string
	ClientIP           string
	SrcHost            string
	SrcURL             string
	Locale             string
}

func (s *PaymentService) GetCardCheckoutInfo(ctx context.Context) (*CardCheckoutInfo, error) {
	if err := s.ensureCardPaymentEnabled(ctx); err != nil {
		return nil, err
	}
	instances, err := s.configService.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true), paymentproviderinstance.ProviderKeyEQ(payment.TypeLdxp)).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query card providers: %w", err)
	}
	out := &CardCheckoutInfo{Shops: make([]CardCheckoutShop, 0, len(instances))}
	for _, inst := range instances {
		ldxp, err := s.ldxpProviderFromInstance(ctx, inst)
		if err != nil {
			slog.Warn("跳过链动小铺发卡店铺：服务商实例配置不可用", "instance_id", inst.ID, "error", err)
			continue
		}
		shop, shopErr := ldxp.GetShopInfo(ctx)
		categories, categoryErr := ldxp.ListCategories(ctx)
		channels, channelErr := ldxp.ListChannels(ctx)
		if shopErr != nil {
			slog.Warn("读取链动小铺店铺信息失败，使用服务商实例名称兜底", "instance_id", inst.ID, "error", shopErr)
			shop = &provider.LdxpShopInfo{Token: ldxp.ShopToken(), Nickname: inst.Name, Link: firstNonEmpty(inst.Name, ldxp.ShopToken())}
		}
		if categoryErr != nil {
			slog.Warn("读取链动小铺分类失败", "instance_id", inst.ID, "error", categoryErr)
			categories = nil
		}
		if channelErr != nil {
			slog.Warn("读取链动小铺支付通道失败", "instance_id", inst.ID, "error", channelErr)
			channels = nil
		}
		goods := make([]provider.LdxpGoods, 0)
		if len(categories) == 0 {
			if list, err := ldxp.ListGoods(ctx, 0); err == nil {
				goods = append(goods, list...)
			} else {
				slog.Warn("读取链动小铺商品失败", "instance_id", inst.ID, "category_id", 0, "error", err)
			}
		}
		for _, category := range categories {
			list, err := ldxp.ListGoods(ctx, category.ID)
			if err != nil {
				slog.Warn("读取链动小铺分类商品失败", "instance_id", inst.ID, "category_id", category.ID, "error", err)
				continue
			}
			goods = append(goods, list...)
		}
		out.Shops = append(out.Shops, CardCheckoutShop{
			ProviderInstanceID: strconv.FormatInt(int64(inst.ID), 10),
			Name:               inst.Name,
			Shop:               shop,
			Categories:         categories,
			Goods:              goods,
			Channels:           channels,
		})
	}
	return out, nil
}

func (s *PaymentService) GetCardPrice(ctx context.Context, req CardPriceRequest) (*provider.LdxpPrice, error) {
	if err := s.ensureCardPaymentEnabled(ctx); err != nil {
		return nil, err
	}
	ldxp, err := s.ldxpProviderByInstanceID(ctx, req.ProviderInstanceID)
	if err != nil {
		return nil, err
	}
	return ldxp.GetGoodsPrice(ctx, strings.TrimSpace(req.GoodsKey), req.Quantity, req.CouponCode, req.ChannelID)
}

func (s *PaymentService) CreateCardOrder(ctx context.Context, req CardOrderRequest) (*CreateOrderResponse, error) {
	if err := s.ensureCardPaymentEnabled(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GoodsKey) == "" || req.ChannelID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CARD_ORDER", "goods_key and channel_id are required")
	}
	if req.Quantity <= 0 {
		return nil, infraerrors.BadRequest("INVALID_CARD_ORDER", "quantity must be greater than 0")
	}
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != payment.EntityStatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}
	inst, ldxp, err := s.ldxpInstanceAndProvider(ctx, req.ProviderInstanceID)
	if err != nil {
		return nil, err
	}
	if err := s.validateCardGoodsStock(ctx, ldxp, strings.TrimSpace(req.GoodsKey), req.Quantity); err != nil {
		return nil, err
	}
	price, err := ldxp.GetGoodsPrice(ctx, strings.TrimSpace(req.GoodsKey), req.Quantity, req.CouponCode, req.ChannelID)
	if err != nil {
		return nil, err
	}
	order, err := s.createCardOrderRecord(ctx, req, user, inst, ldxp.ShopToken(), price.TotalAmount)
	if err != nil {
		return nil, err
	}
	resolvedContact := firstNonEmpty(req.Contact, user.Email, order.OutTradeNo)
	resolvedQueryPassword := firstNonEmpty(req.QueryPassword, s.ldxpConfiguredQueryPassword(ctx, inst), order.OutTradeNo)
	if err := s.updateCardOrderQuerySnapshot(ctx, order.ID, resolvedContact, resolvedQueryPassword); err != nil {
		return nil, err
	}
	upstream, err := ldxp.CreateCardOrder(ctx, provider.LdxpCardOrderRequest{
		GoodsKey: req.GoodsKey, Quantity: req.Quantity, ChannelID: req.ChannelID,
		CouponCode: req.CouponCode, Contact: resolvedContact,
		QueryPassword: resolvedQueryPassword, OrderID: order.OutTradeNo,
	})
	if err != nil {
		_, _ = s.entClient.PaymentOrder.UpdateOneID(order.ID).SetStatus(OrderStatusFailed).SetFailedReason(err.Error()).SetFailedAt(time.Now()).Save(ctx)
		return nil, err
	}
	_, err = s.entClient.PaymentOrder.UpdateOneID(order.ID).
		SetPaymentTradeNo(upstream.TradeNo).
		SetPayAmount(upstream.TotalAmount).
		SetNillablePayURL(psNilIfEmpty(upstream.PayURL)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update card order payment details: %w", err)
	}
	s.writeAuditLog(ctx, order.ID, "CARD_ORDER_CREATED", fmt.Sprintf("user:%d", req.UserID), map[string]any{
		"providerInstanceID": req.ProviderInstanceID,
		"goodsKey":           req.GoodsKey,
		"quantity":           req.Quantity,
		"channelID":          req.ChannelID,
		"payAmount":          upstream.TotalAmount,
	})
	return &CreateOrderResponse{
		OrderID: order.ID, Amount: order.Amount, PayAmount: upstream.TotalAmount, FeeRate: 0,
		Status: OrderStatusPending, ResultType: payment.CreatePaymentResultOrderCreated,
		PaymentType: payment.TypeLdxp, OutTradeNo: order.OutTradeNo, PayURL: upstream.PayURL,
		ExpiresAt: order.ExpiresAt, PaymentMode: inst.PaymentMode,
	}, nil
}

func (s *PaymentService) ensureCardPaymentEnabled(ctx context.Context) error {
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}
	if !cfg.CardEnabled {
		return infraerrors.Forbidden("PAYMENT_CARD_DISABLED", "card payment is disabled")
	}
	return nil
}

func (s *PaymentService) createCardOrderRecord(ctx context.Context, req CardOrderRequest, user *User, inst *dbent.PaymentProviderInstance, shopToken string, payAmount float64) (*dbent.PaymentOrder, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.checkPendingLimit(ctx, tx, req.UserID, cfg.MaxPendingOrders); err != nil {
		return nil, err
	}
	outTradeNo, err := s.allocateOutTradeNo(ctx, tx)
	if err != nil {
		return nil, err
	}
	timeout := cfg.OrderTimeoutMin
	if timeout <= 0 {
		timeout = defaultOrderTimeoutMin
	}
	instanceID := strconv.FormatInt(int64(inst.ID), 10)
	snapshot := map[string]any{
		"schema_version":       2,
		"provider_instance_id": instanceID,
		"provider_key":         payment.TypeLdxp,
		"payment_kind":         "card",
		"shop_token":           shopToken,
		"goods_key":            strings.TrimSpace(req.GoodsKey),
		"channel_id":           req.ChannelID,
		"quantity":             req.Quantity,
		"coupon_code":          strings.TrimSpace(req.CouponCode),
		"currency":             payment.DefaultPaymentCurrency,
		"auto_redeem":          req.AutoRedeem == nil || *req.AutoRedeem,
	}
	order, err := tx.PaymentOrder.Create().
		SetUserID(req.UserID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetNillableUserNotes(psNilIfEmpty(user.Notes)).
		SetAmount(payAmount).
		SetPayAmount(payAmount).
		SetFeeRate(0).
		SetRechargeCode("").
		SetOutTradeNo(outTradeNo).
		SetPaymentType(payment.TypeLdxp).
		SetPaymentTradeNo("").
		SetProviderInstanceID(instanceID).
		SetProviderKey(payment.TypeLdxp).
		SetProviderSnapshot(snapshot).
		SetOrderType(payment.OrderTypeCard).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Duration(timeout) * time.Minute)).
		SetClientIP(req.ClientIP).
		SetSrcHost(req.SrcHost).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create card order: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit card order transaction: %w", err)
	}
	return order, nil
}

func (s *PaymentService) ldxpConfiguredQueryPassword(ctx context.Context, inst *dbent.PaymentProviderInstance) string {
	if s == nil || s.loadBalancer == nil || inst == nil {
		return ""
	}
	cfg, err := s.loadBalancer.GetInstanceConfig(ctx, int64(inst.ID))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cfg["queryPassword"])
}

func (s *PaymentService) updateCardOrderQuerySnapshot(ctx context.Context, orderID int64, contact string, queryPassword string) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return fmt.Errorf("reload card order: %w", err)
	}
	snapshot := map[string]any{}
	for key, value := range order.ProviderSnapshot {
		snapshot[key] = value
	}
	snapshot["contact"] = strings.TrimSpace(contact)
	snapshot["query_password"] = strings.TrimSpace(queryPassword)
	if _, err := s.entClient.PaymentOrder.UpdateOneID(orderID).SetProviderSnapshot(snapshot).Save(ctx); err != nil {
		return fmt.Errorf("update card order query snapshot: %w", err)
	}
	return nil
}

func (s *PaymentService) validateCardGoodsStock(ctx context.Context, ldxp *provider.Ldxp, goodsKey string, quantity int) error {
	goods, err := s.findLdxpGoods(ctx, ldxp, goodsKey)
	if err != nil {
		return err
	}
	if goods == nil {
		return infraerrors.BadRequest("CARD_GOODS_NOT_FOUND", "card goods not found")
	}
	if goods.StockCount <= 0 {
		return infraerrors.BadRequest("CARD_GOODS_OUT_OF_STOCK", "card goods is out of stock")
	}
	if quantity > goods.StockCount {
		return infraerrors.BadRequest("CARD_GOODS_STOCK_NOT_ENOUGH", "card goods stock is not enough")
	}
	if goods.LimitCount > 0 && quantity > goods.LimitCount {
		return infraerrors.BadRequest("CARD_GOODS_LIMIT_EXCEEDED", "card goods purchase limit exceeded")
	}
	return nil
}

func (s *PaymentService) findLdxpGoods(ctx context.Context, ldxp *provider.Ldxp, goodsKey string) (*provider.LdxpGoods, error) {
	if ldxp == nil || strings.TrimSpace(goodsKey) == "" {
		return nil, nil
	}
	categories, err := ldxp.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取链动小铺分类失败: %w", err)
	}
	categoryIDs := []int64{0}
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
	}
	seen := map[int64]bool{}
	for _, categoryID := range categoryIDs {
		if seen[categoryID] {
			continue
		}
		seen[categoryID] = true
		list, err := ldxp.ListGoods(ctx, categoryID)
		if err != nil {
			return nil, fmt.Errorf("读取链动小铺商品失败: %w", err)
		}
		for _, goods := range list {
			if strings.EqualFold(strings.TrimSpace(goods.GoodsKey), strings.TrimSpace(goodsKey)) {
				return &goods, nil
			}
		}
	}
	return nil, nil
}

func (s *PaymentService) ldxpProviderByInstanceID(ctx context.Context, instanceID string) (*provider.Ldxp, error) {
	inst, ldxp, err := s.ldxpInstanceAndProvider(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, infraerrors.NotFound("PROVIDER_NOT_FOUND", "provider instance not found")
	}
	return ldxp, nil
}

func (s *PaymentService) ldxpInstanceAndProvider(ctx context.Context, instanceID string) (*dbent.PaymentProviderInstance, *provider.Ldxp, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(instanceID), 10, 64)
	if err != nil || id <= 0 {
		return nil, nil, infraerrors.BadRequest("INVALID_PROVIDER_INSTANCE", "provider_instance_id is invalid")
	}
	inst, err := s.configService.entClient.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.IDEQ(id), paymentproviderinstance.EnabledEQ(true), paymentproviderinstance.ProviderKeyEQ(payment.TypeLdxp)).
		Only(ctx)
	if err != nil {
		return nil, nil, infraerrors.NotFound("PROVIDER_NOT_FOUND", "provider instance not found")
	}
	ldxp, err := s.ldxpProviderFromInstance(ctx, inst)
	if err != nil {
		return nil, nil, err
	}
	return inst, ldxp, nil
}

func (s *PaymentService) ldxpProviderFromInstance(ctx context.Context, inst *dbent.PaymentProviderInstance) (*provider.Ldxp, error) {
	prov, err := s.createProviderFromInstance(ctx, inst)
	if err != nil {
		return nil, err
	}
	ldxp, ok := prov.(*provider.Ldxp)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("PAYMENT_PROVIDER_MISCONFIGURED", "provider is not ldxp")
	}
	return ldxp, nil
}

func (s *PaymentService) ExecuteCardFulfillment(ctx context.Context, oid int64) error {
	o, err := s.entClient.PaymentOrder.Get(ctx, oid)
	if err != nil {
		return infraerrors.NotFound("NOT_FOUND", "order not found")
	}
	if o.Status == OrderStatusCompleted {
		return nil
	}
	if o.Status != OrderStatusPaid && o.Status != OrderStatusFailed {
		return infraerrors.BadRequest("INVALID_STATUS", "order cannot fulfill in status "+o.Status)
	}
	c, err := s.entClient.PaymentOrder.Update().Where(paymentorder.IDEQ(oid), paymentorder.StatusIn(OrderStatusPaid, OrderStatusFailed)).SetStatus(OrderStatusRecharging).Save(ctx)
	if err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	if c == 0 {
		return nil
	}
	if err := s.doCardFulfillment(ctx, o); err != nil {
		s.markFailed(ctx, oid, err)
		return err
	}
	return nil
}

func (s *PaymentService) doCardFulfillment(ctx context.Context, o *dbent.PaymentOrder) error {
	inst, err := s.getOrderProviderInstance(ctx, o)
	if err != nil {
		return err
	}
	ldxp, err := s.ldxpProviderFromInstance(ctx, inst)
	if err != nil {
		return err
	}
	tradeNo := strings.TrimSpace(o.PaymentTradeNo)
	if tradeNo == "" {
		tradeNo = strings.TrimSpace(o.OutTradeNo)
	}
	info, err := ldxp.GetOrderInfoWithCards(ctx, tradeNo, ldxpOrderQueryPassword(o))
	if err != nil {
		return err
	}
	if info.Status != 1 && info.Sendout != 1 {
		return fmt.Errorf("链动小铺订单尚未发卡")
	}
	if len(info.Cards) == 0 {
		return fmt.Errorf("链动小铺订单未返回卡密")
	}
	cards := normalizeCardCodes(info.Cards)
	if len(cards) == 0 {
		return fmt.Errorf("链动小铺订单未返回有效卡密")
	}
	if err := s.updateCardOrderCardsSnapshot(ctx, o.ID, cards); err != nil {
		return err
	}
	if ldxpOrderAutoRedeem(o) {
		redeemedCards := make([]string, 0, len(cards))
		for _, card := range cards {
			if err := s.redeemCardCode(ctx, o.UserID, card); err != nil {
				return err
			}
			redeemedCards = append(redeemedCards, card)
		}
		if err := s.updateCardOrderRedeemedCardsSnapshot(ctx, o.ID, redeemedCards); err != nil {
			return err
		}
		return s.markCompleted(ctx, o, "CARD_REDEEM_SUCCESS")
	}
	return s.markCompleted(ctx, o, "CARD_DELIVERED")
}

func normalizeCardCodes(cards []string) []string {
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		if code := strings.TrimSpace(card); code != "" {
			out = append(out, code)
		}
	}
	return out
}

func (s *PaymentService) updateCardOrderCardsSnapshot(ctx context.Context, orderID int64, cards []string) error {
	return s.updateCardOrderSnapshotValues(ctx, orderID, map[string]any{"card_codes": cards})
}

func (s *PaymentService) updateCardOrderRedeemedCardsSnapshot(ctx context.Context, orderID int64, cards []string) error {
	return s.updateCardOrderSnapshotValues(ctx, orderID, map[string]any{"redeemed_card_codes": cards})
}

func (s *PaymentService) updateCardOrderSnapshotValues(ctx context.Context, orderID int64, values map[string]any) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, orderID)
	if err != nil {
		return fmt.Errorf("reload card order: %w", err)
	}
	snapshot := map[string]any{}
	for key, value := range order.ProviderSnapshot {
		snapshot[key] = value
	}
	for key, value := range values {
		snapshot[key] = value
	}
	if _, err := s.entClient.PaymentOrder.UpdateOneID(orderID).SetProviderSnapshot(snapshot).Save(ctx); err != nil {
		return fmt.Errorf("update card order snapshot: %w", err)
	}
	order.ProviderSnapshot = snapshot
	return nil
}

func (s *PaymentService) redeemCardCode(ctx context.Context, userID int64, code string) error {
	if code == "" {
		return fmt.Errorf("链动小铺返回了空卡密")
	}
	existing, err := s.redeemService.GetByCode(ctx, code)
	if err == nil && existing != nil && existing.IsUsed() {
		if existing.UsedBy != nil && *existing.UsedBy == userID {
			return nil
		}
		return fmt.Errorf("兑换码已被其他用户使用")
	}
	if _, err := s.redeemService.Redeem(ctx, userID, code); err != nil {
		return fmt.Errorf("自动兑换失败: %w", err)
	}
	return nil
}
