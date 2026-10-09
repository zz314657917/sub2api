package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	_ "github.com/lib/pq"
)

// These tests intentionally use the controller-provided PostgreSQL DSN. Each
// test creates and drops only its own schema; no user database or provider is
// contacted. The EasyPay endpoint below is a process-local fake provider.

type storeCheckoutSettingsRepo struct {
	SettingRepository
	mu     sync.RWMutex
	values map[string]string
}

func (r *storeCheckoutSettingsRepo) GetAll(context.Context) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]string, len(r.values))
	for key, value := range r.values {
		result[key] = value
	}
	return result, nil
}

func (r *storeCheckoutSettingsRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		result[key] = r.values[key]
	}
	return result, nil
}

func (r *storeCheckoutSettingsRepo) set(key, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[key] = value
}

type storeCheckoutUserRepo struct {
	UserRepository
	user *User
}

func (r *storeCheckoutUserRepo) GetByID(context.Context, int64) (*User, error) {
	copy := *r.user
	return &copy, nil
}

type storeCheckoutLoadBalancer struct {
	selection *payment.InstanceSelection
}

func (b *storeCheckoutLoadBalancer) GetInstanceConfig(context.Context, int64) (map[string]string, error) {
	return b.selection.Config, nil
}

func (b *storeCheckoutLoadBalancer) SelectInstance(context.Context, string, payment.PaymentType, payment.Strategy, float64) (*payment.InstanceSelection, error) {
	selection := *b.selection
	selection.Config = make(map[string]string, len(b.selection.Config))
	for key, value := range b.selection.Config {
		selection.Config[key] = value
	}
	return &selection, nil
}

type storeCheckoutFulfillmentRecorder struct{ calls atomic.Int64 }

func (r *storeCheckoutFulfillmentRecorder) FulfillStoreOrder(context.Context, int64) error {
	r.calls.Add(1)
	return nil
}
func (*storeCheckoutFulfillmentRecorder) ReleaseStoreReservation(context.Context, int64, string) error {
	return nil
}
func (*storeCheckoutFulfillmentRecorder) ReconcileStoreOrders(context.Context) (int, error) {
	return 0, nil
}

type storeCheckoutFakeProvider struct {
	server *httptest.Server
	calls  atomic.Int64
	fail   atomic.Bool
}

type storeCheckoutEncryptor struct{}

func (storeCheckoutEncryptor) Encrypt(value string) (string, error) {
	return "qa-cipher:" + base64.RawStdEncoding.EncodeToString([]byte(value)), nil
}

func (storeCheckoutEncryptor) Decrypt(value string) (string, error) {
	encoded := strings.TrimPrefix(value, "qa-cipher:")
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func newStoreCheckoutFakeProvider(t *testing.T) *storeCheckoutFakeProvider {
	t.Helper()
	p := &storeCheckoutFakeProvider{}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/mapi.php" {
			t.Errorf("unexpected fake provider request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		p.calls.Add(1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse fake provider form: %v", err)
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if p.fail.Load() {
			http.Error(w, "fake provider unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"trade_no":"fake-trade","payurl":"/pay/fake","qrcode":"fake-qr"}`))
	}))
	t.Cleanup(p.server.Close)
	return p
}

type storeCheckoutEnv struct {
	db       *sql.DB
	client   *dbent.Client
	settings *storeCheckoutSettingsRepo
	store    *DigitalStoreService
	payment  *PaymentService
	provider *storeCheckoutFakeProvider
}

func newStoreCheckoutEnv(t *testing.T) *storeCheckoutEnv {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("DIGITAL_STORE_TEST_DSN"))
	if dsn == "" {
		t.Fatal("DIGITAL_STORE_TEST_DSN is required for DigitalStore PostgreSQL tests")
	}
	ctx := context.Background()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("digital_store_checkout_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE") })
	if _, err := db.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
		t.Fatal(err)
	}

	baseMigration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "092_payment_orders.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(baseMigration)); err != nil {
		t.Fatalf("apply payment_orders prerequisite: %v", err)
	}
	auditMigration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "093_payment_audit_logs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(auditMigration)); err != nil {
		t.Fatalf("apply payment audit prerequisite: %v", err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE payment_orders ADD COLUMN out_trade_no VARCHAR(64) NOT NULL DEFAULT ''; ALTER TABLE payment_orders ADD COLUMN provider_key VARCHAR(30); ALTER TABLE payment_orders ADD COLUMN provider_snapshot JSONB;`); err != nil {
		t.Fatalf("add current payment order columns: %v", err)
	}
	storeMigration, err := os.ReadFile(filepath.Join("..", "..", "migrations", "246_digital_store.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(storeMigration)); err != nil {
		t.Fatalf("apply migration 246: %v", err)
	}

	clientDB, err := sql.Open("postgres", dsn+"&search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = clientDB.Close() })
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, clientDB)))
	t.Cleanup(func() { _ = client.Close() })
	provider := newStoreCheckoutFakeProvider(t)
	settings := &storeCheckoutSettingsRepo{values: map[string]string{
		SettingKeyServiceStoreEnabled: "true",
		SettingPaymentEnabled:         "true",
		SettingOrderTimeoutMinutes:    "30",
		SettingMaxPendingOrders:       "10",
	}}
	configSvc := NewPaymentConfigService(nil, settings, nil)
	loadBalancer := &storeCheckoutLoadBalancer{selection: &payment.InstanceSelection{
		InstanceID:     "qa-fake-easypay",
		ProviderKey:    payment.TypeEasyPay,
		SupportedTypes: payment.TypeWxpay,
		Config: map[string]string{
			"pid": "qa-pid", "pkey": "qa-key", "apiBase": provider.server.URL,
			"notifyUrl": "https://example.test/notify", "returnUrl": "https://example.test/return",
		},
	}}
	userRepo := &storeCheckoutUserRepo{user: &User{ID: 7, Email: "qa@example.test", Username: "qa", Status: payment.EntityStatusActive}}
	paymentSvc := NewPaymentService(client, nil, loadBalancer, nil, nil, configSvc, userRepo, nil, nil)
	store := NewDigitalStoreService(client, NewSettingService(settings, &config.Config{}), storeCheckoutEncryptor{}, paymentSvc)
	return &storeCheckoutEnv{db: db, client: client, settings: settings, store: store, payment: paymentSvc, provider: provider}
}

func (e *storeCheckoutEnv) addCardProduct(t *testing.T, priceCents int64, stock int) int64 {
	t.Helper()
	product, err := e.store.AdminCreateProduct(context.Background(), DigitalStoreProductInput{Name: "QA card", Kind: "card", PriceCents: priceCents, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	items := make([]DigitalStoreStockInput, stock)
	for i := range items {
		items[i] = DigitalStoreStockInput{Content: fmt.Sprintf("QA-CARD-%d", i)}
	}
	if _, err := e.store.AdminImportStock(context.Background(), product.ID, items); err != nil {
		t.Fatal(err)
	}
	return product.ID
}

func storeCheckoutInput(productID int64, key string) DigitalStoreCreateOrderInput {
	return DigitalStoreCreateOrderInput{UserID: 7, ProductID: productID, IdempotencyKey: key, PaymentType: payment.TypeWxpay}
}

func TestDigitalStorePostgresCheckoutFinalUnitConcurrentDifferentKeys(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 1299, 1)
	keys := []string{"final-unit-a", "final-unit-b"}
	results := make(chan struct {
		response *CreateOrderResponse
		err      error
	}, len(keys))
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			response, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, key))
			results <- struct {
				response *CreateOrderResponse
				err      error
			}{response, err}
		}(key)
	}
	wg.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result.err == nil && result.response != nil {
			successes++
		}
	}
	if successes > 1 {
		t.Fatalf("final unit created %d successful orders", successes)
	}
	if successes != 1 {
		t.Fatalf("final unit created %d successful orders, want 1", successes)
	}
	if got := env.provider.calls.Load(); got != 1 {
		t.Fatalf("fake provider CreatePayment calls=%d, want 1", got)
	}
	var orders, reserved, available int
	if err := env.db.QueryRow(`SELECT count(*) FROM store_orders`).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := env.db.QueryRow(`SELECT count(*) FILTER (WHERE state='reserved'), count(*) FILTER (WHERE state='available') FROM store_inventory`).Scan(&reserved, &available); err != nil {
		t.Fatal(err)
	}
	if orders != 1 || reserved != 1 || available != 0 {
		t.Fatalf("unexpected final-unit state: orders=%d reserved=%d available=%d", orders, reserved, available)
	}
}

func TestDigitalStorePostgresCheckoutIdempotencyProviderFailureAndDisabledResume(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 2500, 2)
	responses := make(chan struct {
		response *CreateOrderResponse
		err      error
	}, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "same-key"))
			responses <- struct {
				response *CreateOrderResponse
				err      error
			}{response, err}
		}()
	}
	wg.Wait()
	close(responses)
	var first, second *CreateOrderResponse
	for result := range responses {
		if result.err != nil {
			t.Fatal(result.err)
		}
		if first == nil {
			first = result.response
		} else {
			second = result.response
		}
	}
	if first == nil || second == nil || first.OrderID != second.OrderID || env.provider.calls.Load() != 1 {
		t.Fatalf("idempotent replay order IDs=%d/%d provider calls=%d", first.OrderID, second.OrderID, env.provider.calls.Load())
	}
	if err := env.store.persistStoreLaunchResponse(context.Background(), first.OrderID, &CreateOrderResponse{OrderID: first.OrderID, ClientSecret: "qa-client-secret"}); err != nil {
		t.Fatal(err)
	}
	var encryptedLaunch string
	if err := env.db.QueryRow(`SELECT encrypted_checkout_response FROM store_orders WHERE payment_order_id=$1`, first.OrderID).Scan(&encryptedLaunch); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encryptedLaunch, "qa-client-secret") {
		t.Fatal("store launch response persisted client secret as plaintext")
	}
	resumedSecret, err := env.store.ResumeStoreOrder(context.Background(), 7, first.OrderID)
	if err != nil || resumedSecret.ClientSecret != "qa-client-secret" {
		t.Fatalf("encrypted launch recovery response=%+v err=%v", resumedSecret, err)
	}
	otherProductID := env.addCardProduct(t, 2500, 1)
	if _, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(otherProductID, "same-key")); infraerrors.Reason(err) != "STORE_IDEMPOTENCY_CONFLICT" {
		t.Fatalf("same key with different product error=%v reason=%q", err, infraerrors.Reason(err))
	}

	env.provider.fail.Store(true)
	pending, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "provider-failure"))
	if err != nil {
		t.Fatal(err)
	}
	if pending.Status != OrderStatusPending {
		t.Fatalf("provider failure response status=%q, want pending", pending.Status)
	}
	failureCalls := env.provider.calls.Load()
	replay, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "provider-failure"))
	if err != nil || replay.OrderID != pending.OrderID || env.provider.calls.Load() != failureCalls {
		t.Fatalf("provider failure replay response=%+v err=%v provider calls=%d/%d", replay, err, env.provider.calls.Load(), failureCalls)
	}
	var status, inventoryState string
	if err := env.db.QueryRow(`SELECT po.status,si.state FROM payment_orders po JOIN store_orders so ON so.payment_order_id=po.id JOIN store_inventory si ON si.id=so.inventory_id WHERE po.id=$1`, pending.OrderID).Scan(&status, &inventoryState); err != nil {
		t.Fatal(err)
	}
	if status != OrderStatusPending || inventoryState != "reserved" {
		t.Fatalf("provider failure did not preserve pending reservation: status=%q inventory=%q", status, inventoryState)
	}

	env.settings.set(SettingKeyServiceStoreEnabled, "false")
	if _, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "disabled-new")); infraerrors.Reason(err) != "STORE_DISABLED" {
		t.Fatalf("disabled new purchase error=%v reason=%q", err, infraerrors.Reason(err))
	}
	resumed, err := env.store.ResumeStoreOrder(context.Background(), 7, first.OrderID)
	if err != nil || resumed.OrderID != first.OrderID || env.provider.calls.Load() != failureCalls {
		t.Fatalf("disabled resume response=%+v err=%v provider calls=%d/%d", resumed, err, env.provider.calls.Load(), failureCalls)
	}
}

func TestDigitalStorePostgresCheckoutRejectsNonCNYSelectionWithoutMutation(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 1234, 1)
	env.payment.loadBalancer = &storeCheckoutLoadBalancer{selection: &payment.InstanceSelection{
		InstanceID: "qa-usd", ProviderKey: payment.TypeStripe, SupportedTypes: payment.TypeStripe,
		Config: map[string]string{"currency": "USD", "secretKey": "not-used"},
	}}
	request := CreateOrderRequest{UserID: 7, PaymentType: payment.TypeStripe, OrderType: payment.OrderTypeStore}
	input := DigitalStoreCreateOrderInput{UserID: 7, ProductID: productID, IdempotencyKey: "usd-accepted", PaymentType: payment.TypeStripe}
	if _, _, err := env.store.reserveStoreOrder(context.Background(), request, input, &User{ID: 7, Email: "qa@example.test", Username: "qa", Status: payment.EntityStatusActive}, &PaymentConfig{Enabled: true, MaxPendingOrders: 10, OrderTimeoutMin: 30}); infraerrors.Reason(err) != "STORE_CURRENCY_UNSUPPORTED" {
		t.Fatalf("non-CNY checkout error=%v reason=%q", err, infraerrors.Reason(err))
	}
	var orders, reservations int
	if err := env.db.QueryRow(`SELECT count(*) FROM payment_orders`).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := env.db.QueryRow(`SELECT count(*) FROM store_inventory WHERE state='reserved'`).Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if orders != 0 || reservations != 0 || env.provider.calls.Load() != 0 {
		t.Fatalf("non-CNY rejection mutated checkout: payments=%d reserved=%d provider_calls=%d", orders, reservations, env.provider.calls.Load())
	}
}

func TestDigitalStorePostgresCheckoutWechatOAuthBeforeReserve(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 1800, 1)
	env.settings.set(SettingKeyWeChatConnectEnabled, "true")
	env.settings.set(SettingKeyWeChatConnectMPEnabled, "true")
	env.settings.set(SettingKeyWeChatConnectMPAppID, "qa-mp-app")
	env.settings.set(SettingKeyWeChatConnectMPAppSecret, "qa-mp-secret")
	env.payment.resumeService = NewPaymentResumeService([]byte("qa-store-checkout-signing-key-32bytes"))
	env.payment.loadBalancer = &storeCheckoutLoadBalancer{selection: &payment.InstanceSelection{
		InstanceID: "qa-wxpay", ProviderKey: payment.TypeWxpay, SupportedTypes: payment.TypeWxpay,
		Config: map[string]string{"appId": "qa-mp-app", "mpAppId": "qa-mp-app"},
	}}

	response, err := env.store.CreateStoreOrder(context.Background(), DigitalStoreCreateOrderInput{
		UserID: 7, ProductID: productID, IdempotencyKey: "wechat-before-reserve", PaymentType: payment.TypeWxpay, IsWeChatBrowser: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || response.ResultType != payment.CreatePaymentResultOAuthRequired || response.OAuth == nil {
		t.Fatalf("wechat preflight response=%+v", response)
	}
	authorizeURL, err := url.Parse(response.OAuth.AuthorizeURL)
	if err != nil {
		t.Fatal(err)
	}
	token := authorizeURL.Query().Get("store_checkout_token")
	claims, err := env.payment.paymentResume().ParseStoreCheckoutToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.ProductID != productID || claims.IdempotencyKey != "wechat-before-reserve" || claims.PaymentType != payment.TypeWxpay || claims.RedirectTo != "/store" {
		t.Fatalf("unexpected signed store checkout claims=%+v", claims)
	}
	var payments, orders, reservations int
	if err := env.db.QueryRow(`SELECT count(*) FROM payment_orders`).Scan(&payments); err != nil {
		t.Fatal(err)
	}
	if err := env.db.QueryRow(`SELECT count(*) FROM store_orders`).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := env.db.QueryRow(`SELECT count(*) FROM store_inventory WHERE state='reserved'`).Scan(&reservations); err != nil {
		t.Fatal(err)
	}
	if payments != 0 || orders != 0 || reservations != 0 || env.provider.calls.Load() != 0 {
		t.Fatalf("wechat OAuth preflight mutated checkout: payments=%d orders=%d reserved=%d provider_calls=%d", payments, orders, reservations, env.provider.calls.Load())
	}
}

func TestDigitalStorePostgresCheckoutStoreFulfillmentDoesNotFallBackToBalanceOrSubscription(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 900, 1)
	response, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "fulfillment-dispatch"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`UPDATE payment_orders SET status='PAID',paid_at=NOW() WHERE id=$1`, response.OrderID); err != nil {
		t.Fatal(err)
	}
	recorder := &storeCheckoutFulfillmentRecorder{}
	env.payment.SetStoreFulfillment(recorder)
	if err := env.payment.executeFulfillment(context.Background(), response.OrderID); err != nil {
		t.Fatal(err)
	}
	if recorder.calls.Load() != 1 {
		t.Fatalf("store fulfillment calls=%d, want 1", recorder.calls.Load())
	}
	var planID sql.NullInt64
	if err := env.db.QueryRow(`SELECT plan_id FROM payment_orders WHERE id=$1`, response.OrderID).Scan(&planID); err != nil {
		t.Fatal(err)
	}
	if planID.Valid {
		t.Fatalf("store payment acquired subscription plan_id=%d", planID.Int64)
	}
}

func TestDigitalStorePostgresCheckoutLateStorePaymentBeyondGraceNeedsAttention(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 900, 1)
	response, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "late-beyond-grace"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`UPDATE payment_orders SET status='EXPIRED',updated_at=NOW()-INTERVAL '10 minutes' WHERE id=$1`, response.OrderID); err != nil {
		t.Fatal(err)
	}
	if err := env.store.ReleaseStoreReservation(context.Background(), response.OrderID, "expired"); err != nil {
		t.Fatal(err)
	}
	order, err := env.client.PaymentOrder.Get(context.Background(), response.OrderID)
	if err != nil {
		t.Fatal(err)
	}
	env.payment.SetStoreFulfillment(env.store)
	if err := env.payment.toPaid(context.Background(), order, "late-verified-trade", order.PayAmount, payment.TypeEasyPay, ""); err != nil {
		t.Fatal(err)
	}
	var status, delivery, inventory string
	if err := env.db.QueryRow(`SELECT po.status,so.delivery_status,si.state FROM payment_orders po JOIN store_orders so ON so.payment_order_id=po.id JOIN store_inventory si ON si.id=so.inventory_id WHERE po.id=$1`, response.OrderID).Scan(&status, &delivery, &inventory); err != nil {
		t.Fatal(err)
	}
	if status != OrderStatusPaid || delivery != "needs_attention" || inventory != "available" {
		t.Fatalf("late payment state status=%q delivery=%q inventory=%q", status, delivery, inventory)
	}
}

func TestDigitalStorePostgresCheckoutReconcilesReleasedLatePayment(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	productID := env.addCardProduct(t, 1000, 1)
	response, err := env.store.CreateStoreOrder(context.Background(), storeCheckoutInput(productID, "reconcile-released-late"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`UPDATE payment_orders SET status='EXPIRED',updated_at=NOW()-INTERVAL '10 minutes' WHERE id=$1`, response.OrderID); err != nil {
		t.Fatal(err)
	}
	if err := env.store.ReleaseStoreReservation(context.Background(), response.OrderID, "expired"); err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`UPDATE payment_orders SET status='PAID',paid_at=NOW() WHERE id=$1`, response.OrderID); err != nil {
		t.Fatal(err)
	}
	count, err := env.store.ReconcileStoreOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var delivery, inventory string
	if err := env.db.QueryRow(`SELECT so.delivery_status,si.state FROM store_orders so JOIN store_inventory si ON si.id=so.inventory_id WHERE so.payment_order_id=$1`, response.OrderID).Scan(&delivery, &inventory); err != nil {
		t.Fatal(err)
	}
	if count < 1 || delivery != "needs_attention" || inventory != "available" {
		t.Fatalf("reconciled late payment count=%d delivery=%q inventory=%q", count, delivery, inventory)
	}
}
