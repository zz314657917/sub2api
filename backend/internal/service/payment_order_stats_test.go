package service

import (
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func TestBuildPaymentOrderStatsFiltersPaidRevenueByPaidAtRange(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	seriesStart := start
	insidePaidAt := start.AddDate(0, 0, 2)
	outsidePaidAt := start.AddDate(0, 0, -1)

	stats := buildPaymentOrderStats([]*dbent.PaymentOrder{
		{
			Status:      OrderStatusCompleted,
			PayAmount:   100,
			PaymentType: "stripe",
			UserID:      10,
			UserEmail:   "paid@example.test",
			PaidAt:      &insidePaidAt,
		},
		{
			Status:      OrderStatusPaid,
			PayAmount:   40,
			PaymentType: "stripe",
			UserID:      11,
			UserEmail:   "outside@example.test",
			PaidAt:      &outsidePaidAt,
		},
		{
			Status:      OrderStatusPending,
			PayAmount:   60,
			PaymentType: "alipay",
		},
		{
			Status:      OrderStatusFailed,
			PayAmount:   80,
			PaymentType: "wxpay",
			PaidAt:      &insidePaidAt,
		},
	}, paymentStatsRange{
		Start:       &start,
		End:         &end,
		SeriesStart: &seriesStart,
		Days:        7,
	})

	if stats.TotalAmount != 100 {
		t.Fatalf("TotalAmount = %v, want 100", stats.TotalAmount)
	}
	if stats.TotalCount != 1 {
		t.Fatalf("TotalCount = %v, want 1", stats.TotalCount)
	}
	if stats.PendingOrders != 1 {
		t.Fatalf("PendingOrders = %v, want 1", stats.PendingOrders)
	}
	if len(stats.PaymentMethods) != 1 || stats.PaymentMethods[0].Type != "stripe" || stats.PaymentMethods[0].Amount != 100 {
		t.Fatalf("PaymentMethods = %#v, want only stripe amount 100", stats.PaymentMethods)
	}
	if len(stats.DailySeries) != 7 {
		t.Fatalf("DailySeries len = %v, want 7", len(stats.DailySeries))
	}
}
