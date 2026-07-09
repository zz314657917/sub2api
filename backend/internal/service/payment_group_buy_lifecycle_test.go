package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type paymentGroupBuyReleaseRecorder struct {
	orderID int64
	reason  string
	calls   int
}

type paymentGroupBuyPendingProvider struct{}

func (p *paymentGroupBuyPendingProvider) Name() string {
	return "payment-group-buy-pending-provider"
}

func (p *paymentGroupBuyPendingProvider) ProviderKey() string {
	return payment.TypeStripe
}

func (p *paymentGroupBuyPendingProvider) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (p *paymentGroupBuyPendingProvider) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	panic("unexpected call")
}

func (p *paymentGroupBuyPendingProvider) QueryOrder(_ context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	return &payment.QueryOrderResponse{
		TradeNo: tradeNo,
		Status:  payment.ProviderStatusPending,
	}, nil
}

func (p *paymentGroupBuyPendingProvider) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	panic("unexpected call")
}

func (p *paymentGroupBuyPendingProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	panic("unexpected call")
}

func (r *paymentGroupBuyReleaseRecorder) HandleGroupBuyOrderPaid(context.Context, int64) error {
	return nil
}

func (r *paymentGroupBuyReleaseRecorder) ReleaseGroupBuySeatForOrder(_ context.Context, orderID int64, reason string) error {
	r.orderID = orderID
	r.reason = reason
	r.calls++
	return nil
}

func TestPaymentCancelGroupBuyOrderReleasesLockedSeat(t *testing.T) {
	ctx := context.Background()
	client := newPaymentGroupBuyLifecycleTestClient(t)

	user, err := client.User.Create().
		SetEmail("cancel-group-buy@example.com").
		SetPasswordHash("hash").
		SetUsername("cancel-group-buy-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(128).
		SetPayAmount(128).
		SetFeeRate(0).
		SetRechargeCode("GB-CANCEL").
		SetOutTradeNo("sub2_group_buy_cancel").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	registry := payment.NewRegistry()
	registry.Register(&paymentGroupBuyPendingProvider{})
	recorder := &paymentGroupBuyReleaseRecorder{}
	svc := &PaymentService{
		entClient:       client,
		registry:        registry,
		providersLoaded: true,
		groupBuySvc:     recorder,
	}

	outcome, err := svc.CancelOrder(ctx, order.ID, user.ID)
	require.NoError(t, err)
	require.Equal(t, checkPaidResultCancelled, outcome)
	require.Equal(t, 1, recorder.calls)
	require.Equal(t, order.ID, recorder.orderID)
	require.Equal(t, "user cancelled order", recorder.reason)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCancelled, reloaded.Status)
}

func newPaymentGroupBuyLifecycleTestClient(t *testing.T) *dbent.Client {
	t.Helper()

	db, err := sql.Open("sqlite", "file:payment_group_buy_lifecycle?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}
