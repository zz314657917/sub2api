package service

import (
	"bytes"
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func TestDigitalStorePostgresAccountAndFileDelivery(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	ctx := context.Background()
	account, err := env.store.AdminCreateProduct(ctx, DigitalStoreProductInput{Name: "fixture account", Kind: "account", PriceCents: 100, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.store.AdminImportStock(ctx, account.ID, []DigitalStoreStockInput{{Username: "buyer", Password: " pass with spaces ", Notes: "first\nsecond"}}); err != nil {
		t.Fatal(err)
	}
	binary := []byte{0, 255, 1, 128, 13, 10}
	file, err := env.store.UploadFile(ctx, "fixture.bin", binary)
	if err != nil {
		t.Fatal(err)
	}
	product, err := env.store.AdminCreateProduct(ctx, DigitalStoreProductInput{Name: "fixture file", Kind: "file", PriceCents: 100, Enabled: true, FileID: &file.ID})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []*DigitalStoreProduct{account, product} {
		order, err := env.store.CreateStoreOrder(ctx, storeCheckoutInput(item.ID, "delivery-"+item.Kind))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := env.db.Exec(`UPDATE payment_orders SET status='PAID',paid_at=NOW() WHERE id=$1`, order.OrderID); err != nil {
			t.Fatal(err)
		}
		if err := env.store.FulfillStoreOrder(ctx, order.OrderID); err != nil {
			t.Fatal(err)
		}
		delivery, err := env.store.GetDelivery(ctx, 7, order.OrderID)
		if err != nil {
			t.Fatal(err)
		}
		if item.Kind == "account" {
			if delivery.Username != "buyer" || delivery.Password != " pass with spaces " || delivery.Notes != "first\nsecond" {
				t.Fatal("account fields changed on delivery")
			}
		} else {
			_, body, err := env.store.GetDownload(ctx, 7, order.OrderID)
			if err != nil || !bytes.Equal(body, binary) || delivery.Filename != "fixture.bin" {
				t.Fatalf("binary delivery mismatch: %v", err)
			}
			if _, err := env.db.Exec(`UPDATE store_files SET content_sha256=repeat('0',64) WHERE id=$1`, file.ID); err != nil {
				t.Fatal(err)
			}
			if _, _, err := env.store.GetDownload(ctx, 7, order.OrderID); infraerrors.Reason(err) != "STORE_FILE_INVALID" {
				t.Fatalf("corrupt file must fail closed: %v", err)
			}
		}
	}
}

func TestDigitalStorePostgresRefundEntrypointsRejectBeforeMutation(t *testing.T) {
	env := newStoreCheckoutEnv(t)
	ctx := context.Background()
	product := env.addCardProduct(t, 1000, 1)
	created, err := env.store.CreateStoreOrder(ctx, storeCheckoutInput(product, "refund-guards"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.db.Exec(`UPDATE payment_orders SET status='PAID',paid_at=NOW() WHERE id=$1`, created.OrderID); err != nil {
		t.Fatal(err)
	}
	if err := env.store.FulfillStoreOrder(ctx, created.OrderID); err != nil {
		t.Fatal(err)
	}
	check := func(err error) {
		t.Helper()
		if infraerrors.Reason(err) != "STORE_ORDER_REFUND_UNSUPPORTED" {
			t.Fatalf("unexpected refund error: %v", err)
		}
	}
	check(env.payment.RequestRefund(ctx, created.OrderID, 7, "fixture"))
	_, _, err = env.payment.PrepareRefund(ctx, created.OrderID, 10, "fixture", true, true)
	check(err)
	_, err = env.payment.ExecuteRefund(ctx, &RefundPlan{OrderID: created.OrderID, DeductionType: payment.DeductionTypeBalance, BalanceToDeduct: 10})
	check(err)
	_, err = env.payment.QueryAndFinalizeRefund(ctx, created.OrderID)
	check(err)
	var state, delivery string
	if err := env.db.QueryRow(`SELECT po.status,so.delivery_status FROM payment_orders po JOIN store_orders so ON so.payment_order_id=po.id WHERE po.id=$1`, created.OrderID).Scan(&state, &delivery); err != nil {
		t.Fatal(err)
	}
	if state != "COMPLETED" || delivery != "delivered" || env.provider.calls.Load() != 1 {
		t.Fatalf("refund mutated delivered order: %s/%s provider=%d", state, delivery, env.provider.calls.Load())
	}
}
