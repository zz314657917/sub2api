package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestResolveEasyPayReturnedRef(t *testing.T) {
	tests := []struct {
		name, base, ref, want string
	}{
		{"root relative", "https://pay.example.com/xpay/epay", "/api/pay/toapp/ORDER", "https://pay.example.com/api/pay/toapp/ORDER"},
		{"root relative query", "https://pay.example.com/xpay/epay", "/api/pay?oid=ORDER", "https://pay.example.com/api/pay?oid=ORDER"},
		{"protocol relative", "https://pay.example.com/xpay/epay", "//cashier.example.com/pay/ORDER", "//cashier.example.com/pay/ORDER"},
		{"trimmed root relative", "https://pay.example.com/xpay/epay", "  /api/pay/ORDER  ", "https://pay.example.com/api/pay/ORDER"},
		{"http base", "http://pay.example.com", "/api/pay/ORDER", "http://pay.example.com/api/pay/ORDER"},
		{"absolute", "https://pay.example.com", "https://other.example/pay", "https://other.example/pay"},
		{"deep link", "https://pay.example.com", "weixin://wxpay/bizpayurl?pr=ABC", "weixin://wxpay/bizpayurl?pr=ABC"},
		{"opaque token", "https://pay.example.com", "OrderToken123", "OrderToken123"},
		{"invalid base", "pay.example.com", "/api/pay/ORDER", "/api/pay/ORDER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveEasyPayReturnedRef(tt.base, tt.ref); got != tt.want {
				t.Fatalf("resolveEasyPayReturnedRef(%q, %q) = %q, want %q", tt.base, tt.ref, got, tt.want)
			}
		})
	}
}

func TestEasyPayCreatePaymentResolvesRelativeRefsAndPreservesMobileChoice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"trade_no":"TRADE_NO","payurl":"/pc/pay/ORDER","payurl2":"/h5/pay/ORDER","qrcode":"/api/pay/toapp/ORDER"}`))
	}))
	defer server.Close()
	provider, err := NewEasyPay("test-instance", map[string]string{
		"pid": "pid-1", "pkey": "pkey-1", "apiBase": server.URL + "/xpay/epay",
		"notifyUrl": "https://example.com/notify", "returnUrl": "https://example.com/return",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID: "order", Amount: "1.00", PaymentType: payment.TypeWxpay, Subject: "test", IsMobile: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := server.URL + "/h5/pay/ORDER"; resp.PayURL != want {
		t.Fatalf("PayURL = %q, want %q", resp.PayURL, want)
	}
	if want := server.URL + "/api/pay/toapp/ORDER"; resp.QRCode != want {
		t.Fatalf("QRCode = %q, want %q", resp.QRCode, want)
	}
	if resp.TradeNo != "TRADE_NO" {
		t.Fatalf("TradeNo = %q", resp.TradeNo)
	}
}
