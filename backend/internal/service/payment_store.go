package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// DigitalStoreCreateOrderInput is accepted only by the authenticated Store
// endpoint. Amount, product snapshots and inventory identifiers are absent on
// purpose: the store service obtains and locks them server-side.
type DigitalStoreCreateOrderInput struct {
	UserID          int64
	ProductID       int64
	IdempotencyKey  string
	PaymentType     string
	OpenID          string
	ClientIP        string
	IsMobile        bool
	IsWeChatBrowser bool
	SrcHost         string
	SrcURL          string
	ReturnURL       string
	PaymentSource   string
	Locale          string
}

func (in *DigitalStoreCreateOrderInput) normalize() error {
	if in == nil || in.UserID <= 0 || in.ProductID <= 0 {
		return infraerrors.BadRequest("INVALID_STORE_ORDER", "user_id and product_id are required")
	}
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if in.IdempotencyKey == "" || len(in.IdempotencyKey) > 64 {
		return infraerrors.BadRequest("INVALID_IDEMPOTENCY_KEY", "idempotency_key is required and must be at most 64 characters")
	}
	in.PaymentType = NormalizeVisibleMethod(in.PaymentType)
	if in.PaymentType == "" {
		return infraerrors.BadRequest("INVALID_PAYMENT_TYPE", "payment_type is required")
	}
	return nil
}

// ReconcileStoreOrders provides the lifecycle runner a fail-closed no-op when
// Store wiring is unavailable. Store-enabled deployments replace it through
// SetStoreFulfillment; a missing service never attempts a balance fallback.
func (s *PaymentService) ReconcileStoreOrders(ctx context.Context) (int, error) {
	if s == nil || s.storeSvc == nil {
		return 0, nil
	}
	return s.storeSvc.ReconcileStoreOrders(ctx)
}

// paymentStoreOrderType is retained here to keep store-only helpers anchored
// to the explicit payment type rather than a string literal in handlers.
const paymentStoreOrderType = payment.OrderTypeStore

// CreateStoreOrder reserves exactly one server-selected item and creates its
// payment order in the same transaction. The idempotency row is queried before
// product/stock selection and protected by its unique SQL constraint, so a
// replay never invokes the provider a second time.
func (s *DigitalStoreService) CreateStoreOrder(ctx context.Context, input DigitalStoreCreateOrderInput) (*CreateOrderResponse, error) {
	if err := s.requireEnabled(ctx); err != nil {
		return nil, err
	}
	if err := input.normalize(); err != nil {
		return nil, err
	}
	if s.paymentSvc == nil || s.paymentSvc.configService == nil || s.paymentSvc.userRepo == nil {
		return nil, infraerrors.ServiceUnavailable("STORE_CHECKOUT_UNAVAILABLE", "store checkout is not configured")
	}
	if existing, found, err := s.storeOrderReplay(ctx, input.UserID, input.ProductID, input.IdempotencyKey); err != nil {
		return nil, err
	} else if found {
		return existing, nil
	}

	user, err := s.paymentSvc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != payment.EntityStatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}
	cfg, err := s.paymentSvc.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}
	if err := s.paymentSvc.checkCancelRateLimit(ctx, input.UserID, cfg); err != nil {
		return nil, err
	}

	// The selection is later snapshotted in the same transaction as stock. Its
	// amount is revalidated after the product lock, never accepted from input.
	req := CreateOrderRequest{UserID: input.UserID, PaymentType: input.PaymentType, OpenID: input.OpenID, ClientIP: input.ClientIP, IsMobile: input.IsMobile, IsWeChatBrowser: input.IsWeChatBrowser, SrcHost: input.SrcHost, SrcURL: input.SrcURL, ReturnURL: input.ReturnURL, PaymentSource: input.PaymentSource, OrderType: payment.OrderTypeStore, Locale: input.Locale}
	if req.IsWeChatBrowser && req.OpenID == "" && req.PaymentType == payment.TypeWxpay {
		var cents int64
		if err := storeQueryRow(ctx, s.entClient, `SELECT price_cents FROM store_products WHERE id=$1 AND enabled=true`, input.ProductID).Scan(&cents); err != nil {
			return nil, infraerrors.NotFound("STORE_PRODUCT_NOT_AVAILABLE", "store product is not available")
		}
		payAmount, selection, err := s.prepareStorePayment(ctx, req, cfg, cents)
		if err != nil {
			return nil, err
		}
		oauth, err := s.paymentSvc.maybeBuildWeChatOAuthRequiredResponseForSelection(ctx, req, float64(cents)/100, payAmount, cfg.RechargeFeeRate, selection)
		if err != nil {
			return nil, err
		}
		if oauth != nil {
			token, err := s.paymentSvc.paymentResume().CreateStoreCheckoutToken(StoreCheckoutClaims{UserID: input.UserID, ProductID: input.ProductID, IdempotencyKey: input.IdempotencyKey, PaymentType: input.PaymentType, RedirectTo: "/store"})
			if err != nil {
				return nil, err
			}
			start, err := url.Parse(oauth.OAuth.AuthorizeURL)
			if err != nil {
				return nil, err
			}
			query := start.Query()
			query.Set("store_checkout_token", token)
			query.Set("redirect", "/store")
			start.RawQuery = query.Encode()
			oauth.OAuth.AuthorizeURL = start.String()
			return oauth, nil
		}
	}
	order, selection, err := s.reserveStoreOrder(ctx, req, input, user, cfg)
	if err != nil {
		// A concurrent unique-key winner is a normal replay, not a second order.
		if replay, found, replayErr := s.storeOrderReplay(ctx, input.UserID, input.ProductID, input.IdempotencyKey); replayErr == nil && found {
			return replay, nil
		}
		return nil, err
	}

	// From this point provider outcome can be ambiguous. Preserve PENDING plus
	// its reservation on every provider error; reconciliation/replay must not
	// manufacture another provider charge.
	amountText := payment.FormatAmountForCurrency(order.PayAmount, PaymentOrderCurrency(order))
	resp, err := s.paymentSvc.invokeProvider(ctx, order, req, cfg, order.Amount, amountText, order.PayAmount, nil, selection)
	if err != nil {
		return s.storePendingResponse(order), nil
	}
	if err := s.persistStoreLaunchResponse(ctx, order.ID, resp); err != nil {
		// The provider has already accepted the payment. Preserve the pending
		// reservation and return its durable basic recovery view; a retry must
		// never create another provider payment.
		return s.storePendingResponse(order), nil
	}
	return resp, nil
}

func (s *DigitalStoreService) selectStoreProvider(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig, amount float64) (*payment.InstanceSelection, error) {
	sel, err := s.paymentSvc.selectCreateOrderInstance(ctx, req, cfg, amount)
	if err != nil {
		return nil, err
	}
	if err := s.paymentSvc.validateSelectedCreateOrderInstance(ctx, req, sel); err != nil {
		return nil, err
	}
	return sel, nil
}

// Catalogue prices are CNY cents. Provider configuration must never reinterpret
// the same numeric price in another currency; there is no implicit FX conversion.
func (s *DigitalStoreService) prepareStorePayment(ctx context.Context, req CreateOrderRequest, cfg *PaymentConfig, cents int64) (float64, *payment.InstanceSelection, error) {
	if cents <= 0 || cents > 100000000 {
		return 0, nil, infraerrors.BadRequest("STORE_PRODUCT_INVALID", "store product price is invalid")
	}
	currency, err := s.paymentSvc.configService.ValidateMethodCurrencyConsistency(ctx, req.PaymentType)
	if err != nil {
		return 0, nil, err
	}
	if currency != "CNY" {
		return 0, nil, infraerrors.BadRequest("STORE_CURRENCY_UNSUPPORTED", "store products require a CNY payment method")
	}
	amountText, amount, err := calculateCreateOrderPayAmount(float64(cents)/100, cfg.RechargeFeeRate, currency)
	if err != nil {
		return 0, nil, err
	}
	selection, err := s.selectStoreProvider(ctx, req, cfg, amount)
	if err != nil {
		return 0, nil, err
	}
	if paymentProviderConfigCurrency(selection.ProviderKey, selection.Config) != "CNY" {
		return 0, nil, infraerrors.BadRequest("STORE_CURRENCY_UNSUPPORTED", "store products require a CNY payment provider")
	}
	if err := validateSelectedCreateOrderAmountCurrency(amountText, selection); err != nil {
		return 0, nil, err
	}
	return amount, selection, nil
}

func (s *DigitalStoreService) reserveStoreOrder(ctx context.Context, req CreateOrderRequest, input DigitalStoreCreateOrderInput, user *User, cfg *PaymentConfig) (*dbent.PaymentOrder, *payment.InstanceSelection, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin store checkout transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	var name, kind string
	var cents int64
	var fileID sql.NullInt64
	err = storeQueryRow(txCtx, tx, `SELECT name,kind,price_cents,file_id FROM store_products WHERE id=$1 AND enabled=true FOR UPDATE`, input.ProductID).Scan(&name, &kind, &cents, &fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, infraerrors.NotFound("STORE_PRODUCT_NOT_AVAILABLE", "store product is not available")
		}
		return nil, nil, fmt.Errorf("lock store product: %w", err)
	}
	if cents <= 0 || cents > 100000000 {
		return nil, nil, infraerrors.BadRequest("STORE_PRODUCT_INVALID", "store product price is invalid")
	}
	var inventoryID sql.NullInt64
	if kind == "card" || kind == "account" {
		err = storeQueryRow(txCtx, tx, `SELECT id FROM store_inventory WHERE product_id=$1 AND kind=$2 AND state='available' ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1`, input.ProductID, kind).Scan(&inventoryID)
		if err == sql.ErrNoRows {
			return nil, nil, infraerrors.Conflict("STORE_OUT_OF_STOCK", "store product is out of stock")
		}
		if err != nil {
			return nil, nil, fmt.Errorf("reserve store inventory: %w", err)
		}
	} else if kind != "file" || !fileID.Valid {
		return nil, nil, infraerrors.BadRequest("STORE_PRODUCT_INVALID", "store product fulfillment source is invalid")
	}
	amount := float64(cents) / 100
	// Provider selection only reads config; snapshot is persisted below.
	payAmount, sel, err := s.prepareStorePayment(txCtx, req, cfg, cents)
	if err != nil {
		return nil, nil, err
	}
	if sel.ProviderKey == payment.TypeWxpay && req.IsWeChatBrowser && strings.TrimSpace(req.OpenID) == "" {
		// Configuration may have changed after OAuth preflight. Roll back instead
		// of reserving stock for a payment that cannot be launched.
		return nil, nil, infraerrors.Conflict("STORE_WECHAT_AUTH_REQUIRED", "wechat authorization is required; retry the same checkout")
	}
	if err := s.paymentSvc.checkPendingLimit(txCtx, tx, input.UserID, cfg.MaxPendingOrders); err != nil {
		return nil, nil, err
	}
	if err := s.paymentSvc.checkDailyLimit(txCtx, tx, input.UserID, amount, cfg.DailyLimit); err != nil {
		return nil, nil, err
	}
	outTradeNo, err := s.paymentSvc.allocateOutTradeNo(txCtx, tx)
	if err != nil {
		return nil, nil, err
	}
	tm := cfg.OrderTimeoutMin
	if tm <= 0 {
		tm = defaultOrderTimeoutMin
	}
	order := tx.PaymentOrder.Create().SetUserID(input.UserID).SetUserEmail(user.Email).SetUserName(user.Username).SetNillableUserNotes(psNilIfEmpty(user.Notes)).SetAmount(amount).SetPayAmount(payAmount).SetFeeRate(cfg.RechargeFeeRate).SetRechargeCode("").SetOutTradeNo(outTradeNo).SetPaymentType(req.PaymentType).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeStore).SetStatus(OrderStatusPending).SetExpiresAt(time.Now().Add(time.Duration(tm) * time.Minute)).SetClientIP(req.ClientIP).SetSrcHost(req.SrcHost).SetProviderSnapshot(buildPaymentOrderProviderSnapshot(sel, req)).SetProviderInstanceID(sel.InstanceID).SetProviderKey(sel.ProviderKey)
	if req.SrcURL != "" {
		order.SetSrcURL(req.SrcURL)
	}
	paymentOrder, err := order.Save(txCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("create store payment order: %w", err)
	}
	var storeOrderID int64
	err = storeQueryRow(txCtx, tx, `INSERT INTO store_orders(payment_order_id,user_id,product_id,idempotency_key,product_name,kind,price_cents,file_id,inventory_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, paymentOrder.ID, input.UserID, input.ProductID, input.IdempotencyKey, name, kind, cents, nullInt64Value(fileID), nullInt64Value(inventoryID)).Scan(&storeOrderID)
	if err != nil {
		return nil, nil, fmt.Errorf("create store order snapshot: %w", err)
	}
	if inventoryID.Valid {
		result, updateErr := tx.ExecContext(txCtx, `UPDATE store_inventory SET state='reserved',reservation_order_id=$2,reserved_at=NOW() WHERE id=$1 AND state='available'`, inventoryID.Int64, storeOrderID)
		if updateErr != nil {
			return nil, nil, fmt.Errorf("mark store inventory reserved: %w", updateErr)
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return nil, nil, infraerrors.Conflict("STORE_OUT_OF_STOCK", "store inventory changed during checkout")
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit store checkout: %w", err)
	}
	return paymentOrder, sel, nil
}

func nullInt64Value(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func (s *DigitalStoreService) storeOrderReplay(ctx context.Context, userID, productID int64, key string) (*CreateOrderResponse, bool, error) {
	var storedProduct, paymentID int64
	var encryptedLaunch sql.NullString
	err := storeQueryRow(ctx, s.entClient, `SELECT product_id,payment_order_id,encrypted_checkout_response FROM store_orders WHERE user_id=$1 AND idempotency_key=$2`, userID, key).Scan(&storedProduct, &paymentID, &encryptedLaunch)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("lookup store checkout replay: %w", err)
	}
	if storedProduct != productID {
		return nil, false, infraerrors.Conflict("STORE_IDEMPOTENCY_CONFLICT", "idempotency_key is already bound to another product")
	}
	o, err := s.entClient.PaymentOrder.Get(ctx, paymentID)
	if err != nil {
		return nil, false, fmt.Errorf("load replay payment order: %w", err)
	}
	if o.Status != OrderStatusPending {
		return s.storePendingResponse(o), true, nil
	}
	if encryptedLaunch.Valid && strings.TrimSpace(encryptedLaunch.String) != "" {
		response, err := s.decryptStoreLaunchResponse(encryptedLaunch.String)
		if err != nil {
			return nil, false, err
		}
		response.Status = o.Status
		return response, true, nil
	}
	return s.storePendingResponse(o), true, nil
}

func (s *DigitalStoreService) persistStoreLaunchResponse(ctx context.Context, paymentOrderID int64, response *CreateOrderResponse) error {
	if response == nil || s == nil || s.encryptor == nil {
		return infraerrors.InternalServer("STORE_LAUNCH_PERSIST_FAILED", "store launch response cannot be persisted")
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal store launch response: %w", err)
	}
	encrypted, err := s.encryptor.Encrypt(string(encoded))
	if err != nil {
		return fmt.Errorf("encrypt store launch response: %w", err)
	}
	result, err := s.entClient.ExecContext(ctx, `UPDATE store_orders SET encrypted_checkout_response=$2,updated_at=NOW() WHERE payment_order_id=$1`, paymentOrderID, encrypted)
	if err != nil {
		return fmt.Errorf("persist store launch response: %w", err)
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
	}
	return nil
}

func (s *DigitalStoreService) decryptStoreLaunchResponse(encrypted string) (*CreateOrderResponse, error) {
	if s == nil || s.encryptor == nil {
		return nil, infraerrors.InternalServer("STORE_LAUNCH_UNAVAILABLE", "store launch response is unavailable")
	}
	encoded, err := s.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, infraerrors.InternalServer("STORE_LAUNCH_UNAVAILABLE", "store launch response is unavailable")
	}
	var response CreateOrderResponse
	if err := json.Unmarshal([]byte(encoded), &response); err != nil || response.OrderID <= 0 {
		return nil, infraerrors.InternalServer("STORE_LAUNCH_UNAVAILABLE", "store launch response is invalid")
	}
	return &response, nil
}

func (s *DigitalStoreService) storePendingResponse(o *dbent.PaymentOrder) *CreateOrderResponse {
	if o == nil {
		return nil
	}
	return &CreateOrderResponse{OrderID: o.ID, Amount: o.Amount, PayAmount: o.PayAmount, FeeRate: o.FeeRate, Status: o.Status, PaymentType: o.PaymentType, OutTradeNo: o.OutTradeNo, PayURL: psStringValue(o.PayURL), QRCode: psStringValue(o.QrCode), ExpiresAt: o.ExpiresAt, ResultType: payment.CreatePaymentResultOrderCreated, Currency: PaymentOrderCurrency(o)}
}

// ResumeStoreOrder is an ownership-checked recovery endpoint. It only returns
// persisted launch data for an already pending order; it never re-enters
// provider creation and remains available while the Store is disabled.
func (s *DigitalStoreService) ResumeStoreOrder(ctx context.Context, userID, paymentOrderID int64) (*CreateOrderResponse, error) {
	if userID <= 0 || paymentOrderID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_STORE_ORDER", "invalid store order")
	}
	var encryptedLaunch sql.NullString
	err := storeQueryRow(ctx, s.entClient, `SELECT encrypted_checkout_response FROM store_orders WHERE payment_order_id=$1 AND user_id=$2`, paymentOrderID, userID).Scan(&encryptedLaunch)
	if err == sql.ErrNoRows {
		return nil, infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("lookup store order resume: %w", err)
	}
	o, err := s.entClient.PaymentOrder.Get(ctx, paymentOrderID)
	if err != nil {
		return nil, infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
	}
	if o.OrderType != payment.OrderTypeStore || o.Status != OrderStatusPending {
		return nil, infraerrors.BadRequest("STORE_ORDER_NOT_PENDING", "store order is no longer pending")
	}
	if encryptedLaunch.Valid && strings.TrimSpace(encryptedLaunch.String) != "" {
		return s.decryptStoreLaunchResponse(encryptedLaunch.String)
	}
	return s.storePendingResponse(o), nil
}

// RetryStoreFulfillment is an administrator-only recovery primitive. It is
// deliberately separate from the periodic reconciler: a late-paid order must
// not silently take newly added stock. The caller must be an admin handler.
func (s *DigitalStoreService) RetryStoreFulfillment(ctx context.Context, paymentOrderID int64) error {
	if s == nil || s.entClient == nil || paymentOrderID <= 0 {
		return infraerrors.BadRequest("INVALID_STORE_ORDER", "invalid store order")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin store retry: %w", err)
	}
	defer tx.Rollback()
	txCtx := dbent.NewTxContext(ctx, tx)
	var paymentStatus string
	var paidAt sql.NullTime
	if err := storeQueryRow(txCtx, tx, `SELECT status,paid_at FROM payment_orders WHERE id=$1 AND order_type=$2 FOR UPDATE`, paymentOrderID, payment.OrderTypeStore).Scan(&paymentStatus, &paidAt); err != nil {
		if err == sql.ErrNoRows {
			return infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
		}
		return err
	}
	if paymentStatus == OrderStatusCompleted {
		return tx.Commit()
	}
	if !paidAt.Valid || (paymentStatus != OrderStatusPaid && paymentStatus != OrderStatusRecharging && paymentStatus != OrderStatusFailed) {
		return infraerrors.BadRequest("STORE_ORDER_NOT_PAID", "only confirmed paid store orders can be retried")
	}
	var storeID, productID int64
	var kind, delivery string
	if err := storeQueryRow(txCtx, tx, `SELECT id,product_id,kind,delivery_status FROM store_orders WHERE payment_order_id=$1 FOR UPDATE`, paymentOrderID).Scan(&storeID, &productID, &kind, &delivery); err != nil {
		return fmt.Errorf("lock store retry snapshot: %w", err)
	}
	if delivery == "delivered" {
		return tx.Commit()
	}
	if delivery != "needs_attention" {
		return infraerrors.BadRequest("STORE_ORDER_NOT_RETRYABLE", "store order is not awaiting administrator retry")
	}
	if kind == "file" {
		if _, err := tx.ExecContext(txCtx, `UPDATE store_orders SET delivery_status='reserved',updated_at=NOW() WHERE id=$1 AND delivery_status='needs_attention'`, storeID); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		return s.FulfillStoreOrder(ctx, paymentOrderID)
	}
	var inventoryID int64
	if err := storeQueryRow(txCtx, tx, `SELECT id FROM store_inventory WHERE product_id=$1 AND kind=$2 AND state='available' ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1`, productID, kind).Scan(&inventoryID); err != nil {
		if err == sql.ErrNoRows {
			return infraerrors.Conflict("STORE_OUT_OF_STOCK", "no replacement inventory is available")
		}
		return fmt.Errorf("claim replacement store inventory: %w", err)
	}
	result, err := tx.ExecContext(txCtx, `UPDATE store_inventory SET state='reserved',reservation_order_id=$2,reserved_at=NOW() WHERE id=$1 AND state='available'`, inventoryID, storeID)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return infraerrors.Conflict("STORE_OUT_OF_STOCK", "replacement inventory changed")
	}
	result, err = tx.ExecContext(txCtx, `UPDATE store_orders SET inventory_id=$2,delivery_status='reserved',updated_at=NOW() WHERE id=$1 AND delivery_status='needs_attention'`, storeID, inventoryID)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return infraerrors.Conflict("STORE_ORDER_CHANGED", "store order changed during retry")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return s.FulfillStoreOrder(ctx, paymentOrderID)
}

// FulfillStoreOrder performs the payment completion and delivery mutation in
// one transaction. The lock order is payment -> store snapshot -> inventory.
func (s *DigitalStoreService) FulfillStoreOrder(ctx context.Context, paymentOrderID int64) error {
	if s == nil || s.entClient == nil || paymentOrderID <= 0 {
		return infraerrors.BadRequest("INVALID_STORE_ORDER", "invalid store order")
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin store fulfillment: %w", err)
	}
	defer tx.Rollback()
	txCtx := dbent.NewTxContext(ctx, tx)
	var status string
	var paidAt sql.NullTime
	if err := storeQueryRow(txCtx, tx, `SELECT status,paid_at FROM payment_orders WHERE id=$1 AND order_type=$2 FOR UPDATE`, paymentOrderID, payment.OrderTypeStore).Scan(&status, &paidAt); err != nil {
		if err == sql.ErrNoRows {
			return infraerrors.NotFound("STORE_ORDER_NOT_FOUND", "store order not found")
		}
		return err
	}
	var storeID int64
	var delivery, kind string
	var inventory sql.NullInt64
	if err := storeQueryRow(txCtx, tx, `SELECT id,delivery_status,kind,inventory_id FROM store_orders WHERE payment_order_id=$1 FOR UPDATE`, paymentOrderID).Scan(&storeID, &delivery, &kind, &inventory); err != nil {
		return fmt.Errorf("lock store snapshot: %w", err)
	}
	if delivery == "delivered" && status == OrderStatusCompleted {
		return tx.Commit()
	}
	if status != OrderStatusPaid && status != OrderStatusRecharging && status != OrderStatusFailed {
		return infraerrors.BadRequest("INVALID_STATUS", "store order cannot fulfill in status "+status)
	}
	if status == OrderStatusFailed && !paidAt.Valid {
		return infraerrors.BadRequest("INVALID_STATUS", "unconfirmed failed store order cannot fulfill")
	}
	if delivery == "released" {
		if _, err := tx.ExecContext(txCtx, `UPDATE store_orders SET delivery_status='needs_attention',updated_at=NOW() WHERE id=$1 AND delivery_status='released'`, storeID); err != nil {
			return fmt.Errorf("mark late paid store order for attention: %w", err)
		}
		return tx.Commit()
	}
	if delivery != "reserved" {
		return infraerrors.BadRequest("INVALID_STORE_DELIVERY_STATUS", "store order delivery state is invalid")
	}
	var encrypted any
	if inventory.Valid {
		var state string
		var reservation sql.NullInt64
		var content string
		if err := storeQueryRow(txCtx, tx, `SELECT state,reservation_order_id,encrypted_content FROM store_inventory WHERE id=$1 FOR UPDATE`, inventory.Int64).Scan(&state, &reservation, &content); err != nil {
			return fmt.Errorf("lock store inventory: %w", err)
		}
		if state != "reserved" || !reservation.Valid || reservation.Int64 != storeID {
			return infraerrors.Conflict("STORE_RESERVATION_LOST", "store inventory reservation is unavailable")
		}
		encrypted = content
		result, err := tx.ExecContext(txCtx, `UPDATE store_inventory SET state='consumed',consumed_at=NOW() WHERE id=$1 AND state='reserved' AND reservation_order_id=$2`, inventory.Int64, storeID)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return infraerrors.Conflict("STORE_RESERVATION_LOST", "store inventory reservation changed")
		}
	}
	if _, err := tx.ExecContext(txCtx, `UPDATE store_orders SET delivery_status='delivered',encrypted_delivery=$2,delivered_at=NOW(),updated_at=NOW() WHERE id=$1 AND delivery_status='reserved'`, storeID, encrypted); err != nil {
		return fmt.Errorf("write store delivery: %w", err)
	}
	result, err := tx.ExecContext(txCtx, `UPDATE payment_orders SET status=$2,completed_at=NOW(),updated_at=NOW() WHERE id=$1 AND status IN ($3,$4,$5)`, paymentOrderID, OrderStatusCompleted, OrderStatusPaid, OrderStatusRecharging, OrderStatusFailed)
	if err != nil {
		return fmt.Errorf("complete store payment: %w", err)
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return infraerrors.Conflict("STORE_PAYMENT_CHANGED", "store payment changed during fulfillment")
	}
	return tx.Commit()
}

func (s *DigitalStoreService) ReleaseStoreReservation(ctx context.Context, paymentOrderID int64, reason string) error {
	if s == nil || s.entClient == nil || paymentOrderID <= 0 {
		return nil
	}
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txCtx := dbent.NewTxContext(ctx, tx)
	var paymentStatus string
	var paidAt sql.NullTime
	err = storeQueryRow(txCtx, tx, `SELECT status,paid_at FROM payment_orders WHERE id=$1 AND order_type=$2 FOR UPDATE`, paymentOrderID, payment.OrderTypeStore).Scan(&paymentStatus, &paidAt)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if (paymentStatus != OrderStatusCancelled && paymentStatus != OrderStatusExpired) || paidAt.Valid {
		return tx.Commit()
	}
	var storeID int64
	var delivery string
	var inventory sql.NullInt64
	err = storeQueryRow(txCtx, tx, `SELECT id,delivery_status,inventory_id FROM store_orders WHERE payment_order_id=$1 FOR UPDATE`, paymentOrderID).Scan(&storeID, &delivery, &inventory)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if delivery != "reserved" {
		return tx.Commit()
	}
	if inventory.Valid {
		_, err = tx.ExecContext(txCtx, `UPDATE store_inventory SET state='available',reservation_order_id=NULL,reserved_at=NULL WHERE id=$1 AND state='reserved' AND reservation_order_id=$2`, inventory.Int64, storeID)
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(txCtx, `UPDATE store_orders SET delivery_status='released',released_at=NOW(),updated_at=NOW() WHERE id=$1 AND delivery_status='reserved'`, storeID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DigitalStoreService) ReconcileStoreOrders(ctx context.Context) (int, error) {
	if s == nil || s.entClient == nil {
		return 0, nil
	}
	// A process can crash after payment cancellation/expiry commits but before
	// cancelCore calls ReleaseStoreReservation. Reclaim only reservations whose
	// own payment is conclusively unpaid/cancelled; paid orders are handled by
	// fulfillment and a released late callback becomes needs_attention.
	releaseRows, err := s.entClient.QueryContext(ctx, `SELECT so.payment_order_id FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id WHERE so.delivery_status='reserved' AND po.order_type=$1 AND po.status IN ($2,$3) ORDER BY so.id LIMIT 50`, payment.OrderTypeStore, OrderStatusCancelled, OrderStatusExpired)
	if err != nil {
		return 0, err
	}
	released := 0
	for releaseRows.Next() {
		var id int64
		if err := releaseRows.Scan(&id); err != nil {
			releaseRows.Close()
			return released, err
		}
		if err := s.ReleaseStoreReservation(ctx, id, "reconcile cancelled or expired store order"); err == nil {
			released++
		}
	}
	if err := releaseRows.Err(); err != nil {
		releaseRows.Close()
		return released, err
	}
	releaseRows.Close()
	rows, err := s.entClient.QueryContext(ctx, `SELECT payment_order_id FROM store_orders so JOIN payment_orders po ON po.id=so.payment_order_id WHERE so.delivery_status IN ('reserved','released') AND po.order_type=$1 AND (po.status IN ($2,$3) OR (po.status=$4 AND po.paid_at IS NOT NULL)) ORDER BY so.id LIMIT 50`, payment.OrderTypeStore, OrderStatusPaid, OrderStatusRecharging, OrderStatusFailed)
	if err != nil {
		return released, err
	}
	defer rows.Close()
	n := released
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return n, err
		}
		if err := s.FulfillStoreOrder(ctx, id); err == nil {
			n++
		}
	}
	return n, rows.Err()
}

type storeQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type storeRow struct {
	rows *sql.Rows
	err  error
}

func storeQueryRow(ctx context.Context, q storeQueryer, query string, args ...any) *storeRow {
	rows, err := q.QueryContext(ctx, query, args...)
	return &storeRow{rows: rows, err: err}
}

func (r *storeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	defer r.rows.Close()
	if !r.rows.Next() {
		if err := r.rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return r.rows.Scan(dest...)
}
