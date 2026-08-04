package handler

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	stripe "github.com/stripe/stripe-go/v85"
	stripewebhook "github.com/stripe/stripe-go/v85/webhook"
	_ "modernc.org/sqlite"
)

func TestCafeWebhookStripeVerifiedCallbackDispatchesOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	order := createCafeWebhookGroupBuyOrder(t, client, "cafe_webhook_verified_1")
	payload := `{"id":"evt_cafe_verified","type":"payment_intent.succeeded"}`
	provider := &cafeWebhookVerifierStub{
		expectedBody:      payload,
		acceptedSignature: "cafe-valid-signature",
		notification: &payment.PaymentNotification{
			OrderID: order.OutTradeNo,
			TradeNo: "cafe-webhook-trade-1",
			Amount:  order.PayAmount,
			Status:  payment.NotificationStatusSuccess,
		},
	}
	router, recorder := newCafeWebhookStripeRouter(client, provider)

	response := performCafeWebhookStripeRequest(t, router, payload, "cafe-valid-signature")
	require.Equal(t, http.StatusOK, response.Code)
	require.Empty(t, response.Body.String())
	require.Len(t, provider.requests, 1)
	require.Equal(t, payload, provider.requests[0].rawBody)
	require.Equal(t, "cafe-valid-signature", provider.requests[0].headers["x-cafe-signature"])
	require.Equal(t, []int64{order.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, order.ID, service.OrderStatusCompleted)
	auditCount := cafeWebhookAuditCount(t, client, order.ID)
	require.NotZero(t, auditCount)

	replayResponse := performCafeWebhookStripeRequest(t, router, payload, "cafe-valid-signature")
	require.Equal(t, http.StatusOK, replayResponse.Code)
	require.Empty(t, replayResponse.Body.String())
	require.Len(t, provider.requests, 2)
	require.Equal(t, []int64{order.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, order.ID, service.OrderStatusCompleted)
	require.Equal(t, auditCount, cafeWebhookAuditCount(t, client, order.ID))
}

func TestCafeWebhookStripeRejectsForgedSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	order := createCafeWebhookGroupBuyOrder(t, client, "cafe_webhook_forged_1")
	payload := `{"id":"evt_cafe_forged","type":"payment_intent.succeeded"}`
	provider := &cafeWebhookVerifierStub{
		expectedBody:      payload,
		acceptedSignature: "cafe-valid-signature",
		notification: &payment.PaymentNotification{
			OrderID: order.OutTradeNo,
			TradeNo: "cafe-webhook-trade-forged",
			Amount:  order.PayAmount,
			Status:  payment.NotificationStatusSuccess,
		},
	}
	router, recorder := newCafeWebhookStripeRouter(client, provider)

	response := performCafeWebhookStripeRequest(t, router, payload, "forged-signature")
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Equal(t, "verify failed", response.Body.String())
	require.Len(t, provider.requests, 1)
	require.Equal(t, payload, provider.requests[0].rawBody)
	require.Equal(t, "forged-signature", provider.requests[0].headers["x-cafe-signature"])
	require.Empty(t, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, order.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, order.ID))
}

func TestCafeStripeWebhookPinnedInstance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	const firstSecret = "whsec_cafe_first_synthetic"
	const secondSecret = "whsec_cafe_second_synthetic"
	createCafeStripeWebhookInstance(t, client, "cafe-stripe-first", firstSecret, []byte(encryptionKey))
	secondInstance := createCafeStripeWebhookInstance(t, client, "cafe-stripe-second", secondSecret, []byte(encryptionKey))

	validOrder := bindCafeWebhookOrderToProvider(t, client, createCafeWebhookGroupBuyOrder(t, client, "cafe_stripe_pinned_valid_1"), secondInstance.ID)
	forgedOrder := bindCafeWebhookOrderToProvider(t, client, createCafeWebhookGroupBuyOrder(t, client, "cafe_stripe_pinned_forged_1"), secondInstance.ID)
	router, recorder := newCafePinnedStripeWebhookRouter(client, []byte(encryptionKey))

	validPayload := cafeStripeWebhookPayload(t, validOrder, "pi_cafe_pinned_valid")
	validResponse := performCafeStripeSignedWebhookRequest(t, router, validPayload, secondSecret)
	require.Equal(t, http.StatusOK, validResponse.Code)
	require.Empty(t, validResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, validOrder.ID, service.OrderStatusCompleted)
	validAuditCount := cafeWebhookAuditCount(t, client, validOrder.ID)
	require.NotZero(t, validAuditCount)

	validReplay := performCafeStripeSignedWebhookRequest(t, router, validPayload, secondSecret)
	require.Equal(t, http.StatusOK, validReplay.Code)
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	require.Equal(t, validAuditCount, cafeWebhookAuditCount(t, client, validOrder.ID))

	forgedPayload := cafeStripeWebhookPayload(t, forgedOrder, "pi_cafe_pinned_forged")
	forgedResponse := performCafeStripeSignedWebhookRequest(t, router, forgedPayload, firstSecret)
	require.Equal(t, http.StatusBadRequest, forgedResponse.Code)
	require.Equal(t, "verify failed", forgedResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, forgedOrder.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, forgedOrder.ID))
}

func TestCafeStripeWebhookExtractOutTradeNo(t *testing.T) {
	valid := `{"data":{"object":{"metadata":{"orderId":" cafe_stripe_order_1 "}}}}`
	require.Equal(t, "cafe_stripe_order_1", extractOutTradeNo(valid, payment.TypeStripe))
	require.Empty(t, extractOutTradeNo(`{"data":{"object":{"metadata":{}}}}`, payment.TypeStripe))
	require.Empty(t, extractOutTradeNo(`{"data":{"object":{"metadata":{"orderId":"   "}}}}`, payment.TypeStripe))
	require.Empty(t, extractOutTradeNo(`{"data":`, payment.TypeStripe))
}

func TestCafeAirwallexWebhookPinnedInstance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	const firstSecret = "airwallex-cafe-first-synthetic"
	const secondSecret = "airwallex-cafe-second-synthetic"
	createCafeAirwallexWebhookInstance(t, client, "cafe-airwallex-first", firstSecret, []byte(encryptionKey))
	secondInstance := createCafeAirwallexWebhookInstance(t, client, "cafe-airwallex-second", secondSecret, []byte(encryptionKey))

	validOrder := bindCafeWebhookOrderToProvider(t, client, createCafeWebhookGroupBuyOrder(t, client, "cafe_airwallex_pinned_valid_1"), secondInstance.ID)
	forgedOrder := bindCafeWebhookOrderToProvider(t, client, createCafeWebhookGroupBuyOrder(t, client, "cafe_airwallex_pinned_forged_1"), secondInstance.ID)
	router, recorder := newCafePinnedAirwallexWebhookRouter(client, []byte(encryptionKey))

	validPayload := cafeAirwallexWebhookPayload(t, validOrder, "int_cafe_pinned_valid")
	validTimestamp, validSignature := cafeAirwallexWebhookSignature(t, validPayload, secondSecret)
	validResponse := performCafeAirwallexSignedWebhookRequest(t, router, validPayload, validTimestamp, validSignature)
	require.Equal(t, http.StatusOK, validResponse.Code)
	require.Empty(t, validResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, validOrder.ID, service.OrderStatusCompleted)
	validAuditCount := cafeWebhookAuditCount(t, client, validOrder.ID)
	require.NotZero(t, validAuditCount)

	validReplay := performCafeAirwallexSignedWebhookRequest(t, router, validPayload, validTimestamp, validSignature)
	require.Equal(t, http.StatusOK, validReplay.Code)
	require.Empty(t, validReplay.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	require.Equal(t, validAuditCount, cafeWebhookAuditCount(t, client, validOrder.ID))

	forgedPayload := cafeAirwallexWebhookPayload(t, forgedOrder, "int_cafe_pinned_forged")
	forgedTimestamp, forgedSignature := cafeAirwallexWebhookSignature(t, forgedPayload, firstSecret)
	forgedResponse := performCafeAirwallexSignedWebhookRequest(t, router, forgedPayload, forgedTimestamp, forgedSignature)
	require.Equal(t, http.StatusBadRequest, forgedResponse.Code)
	require.Equal(t, "verify failed", forgedResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, forgedOrder.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, forgedOrder.ID))
}

func TestCafeAirwallexWebhookExtractOutTradeNo(t *testing.T) {
	valid := `{"data":{"object":{"merchant_order_id":" cafe_airwallex_order_1 "}}}`
	require.Equal(t, "cafe_airwallex_order_1", extractOutTradeNo(valid, payment.TypeAirwallex))
	require.Empty(t, extractOutTradeNo(`{"data":{"object":{}}}`, payment.TypeAirwallex))
	require.Empty(t, extractOutTradeNo(`{"data":{"object":{"merchant_order_id":"   "}}}`, payment.TypeAirwallex))
	require.Empty(t, extractOutTradeNo(`{"data":`, payment.TypeAirwallex))
}

func TestCafeWxpayWebhookCandidateSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	firstPrivateKey := newCafeWxpayWebhookRSAKey(t)
	secondPrivateKey := newCafeWxpayWebhookRSAKey(t)
	firstInstance := createCafeWxpayWebhookInstance(t, client, "cafe-wxpay-first", firstPrivateKey, "cafe-wxpay-key-first", "12345678901234567890123456789012", 10, []byte(encryptionKey))
	secondInstance := createCafeWxpayWebhookInstance(t, client, "cafe-wxpay-second", secondPrivateKey, "cafe-wxpay-key-second", "abcdefghijklmnopqrstuvwxyz012345", 20, []byte(encryptionKey))
	order := createCafeWxpayWebhookGroupBuyOrder(t, client, "cafe_wxpay_candidate_valid_1")
	router, paymentSvc, recorder := newCafeWxpayWebhookRouter(client, []byte(encryptionKey))

	body, headers := cafeWxpayEncryptedWebhook(t, order, secondPrivateKey, "cafe-wxpay-key-second", "abcdefghijklmnopqrstuvwxyz012345", "wxpay-cafe-candidate-trade")
	providers, err := paymentSvc.GetWebhookProviders(context.Background(), payment.TypeWxpay, "")
	require.NoError(t, err)
	require.Len(t, providers, 2)
	require.Less(t, firstInstance.SortOrder, secondInstance.SortOrder)
	_, err = providers[0].VerifyNotification(context.Background(), body, headers)
	require.Error(t, err)
	notification, err := providers[1].VerifyNotification(context.Background(), body, headers)
	require.NoError(t, err)
	require.Equal(t, order.OutTradeNo, notification.OrderID)

	response := performCafeWxpaySignedWebhookRequest(t, router, body, headers)
	require.Equal(t, http.StatusOK, response.Code)
	assertCafeWxpaySuccessResponse(t, response)
	require.Equal(t, []int64{order.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, order.ID, service.OrderStatusCompleted)
	auditCount := cafeWebhookAuditCount(t, client, order.ID)
	require.NotZero(t, auditCount)

	replayResponse := performCafeWxpaySignedWebhookRequest(t, router, body, headers)
	require.Equal(t, http.StatusOK, replayResponse.Code)
	assertCafeWxpaySuccessResponse(t, replayResponse)
	require.Equal(t, []int64{order.ID}, recorder.paidOrderIDs)
	require.Equal(t, auditCount, cafeWebhookAuditCount(t, client, order.ID))
}

func TestCafeWxpayWebhookWrongInstance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	firstPrivateKey := newCafeWxpayWebhookRSAKey(t)
	secondPrivateKey := newCafeWxpayWebhookRSAKey(t)
	createCafeWxpayWebhookInstance(t, client, "cafe-wxpay-first", firstPrivateKey, "cafe-wxpay-key-first", "12345678901234567890123456789012", 10, []byte(encryptionKey))
	createCafeWxpayWebhookInstance(t, client, "cafe-wxpay-second", secondPrivateKey, "cafe-wxpay-key-second", "abcdefghijklmnopqrstuvwxyz012345", 20, []byte(encryptionKey))
	order := createCafeWxpayWebhookGroupBuyOrder(t, client, "cafe_wxpay_wrong_instance_1")
	router, _, recorder := newCafeWxpayWebhookRouter(client, []byte(encryptionKey))

	// The transaction remains encrypted for the second instance, but the signature
	// claims the first key. The first verifier passes RSA then fails GCM; the second
	// verifier rejects the serial, so neither candidate can mutate the order.
	body, headers := cafeWxpayEncryptedWebhook(t, order, firstPrivateKey, "cafe-wxpay-key-first", "abcdefghijklmnopqrstuvwxyz012345", "wxpay-cafe-wrong-instance-trade")
	response := performCafeWxpaySignedWebhookRequest(t, router, body, headers)
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.Equal(t, "verify failed", response.Body.String())
	require.Empty(t, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, order.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, order.ID))
}

func TestCafeWxpayWebhookExtract(t *testing.T) {
	require.Empty(t, extractOutTradeNo(`{"event_type":"TRANSACTION.SUCCESS","resource":{"ciphertext":"encrypted"}}`, payment.TypeWxpay))
}

func TestCafeEasyPayWebhookPinnedInstance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	firstInstance := createCafeEasyPayWebhookInstance(t, client, "cafe-easypay-first", "cafe-easypay-pid-first", "cafe-easypay-pkey-first", 10, []byte(encryptionKey))
	secondInstance := createCafeEasyPayWebhookInstance(t, client, "cafe-easypay-second", "cafe-easypay-pid-second", "cafe-easypay-pkey-second", 20, []byte(encryptionKey))
	validOrder := bindCafeEasyPayWebhookOrder(t, client, createCafeEasyPayWebhookGroupBuyOrder(t, client, "cafe_easypay_pinned_valid_1"), secondInstance.ID)
	forgedOrder := bindCafeEasyPayWebhookOrder(t, client, createCafeEasyPayWebhookGroupBuyOrder(t, client, "cafe_easypay_pinned_forged_1"), secondInstance.ID)
	router, recorder := newCafeEasyPayWebhookRouter(client, []byte(encryptionKey))

	validPayload := cafeEasyPayWebhookPayload(t, validOrder, "cafe-easypay-pid-second", "cafe-easypay-pkey-second", "cafe-easypay-trade-valid")
	validResponse := performCafeEasyPaySignedWebhookRequest(t, router, validPayload)
	require.Equal(t, http.StatusOK, validResponse.Code)
	require.Equal(t, "success", validResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, validOrder.ID, service.OrderStatusCompleted)
	validAuditCount := cafeWebhookAuditCount(t, client, validOrder.ID)
	require.NotZero(t, validAuditCount)

	validReplay := performCafeEasyPaySignedWebhookRequest(t, router, validPayload)
	require.Equal(t, http.StatusOK, validReplay.Code)
	require.Equal(t, "success", validReplay.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	require.Equal(t, validAuditCount, cafeWebhookAuditCount(t, client, validOrder.ID))

	forgedPayload := cafeEasyPayWebhookPayload(t, forgedOrder, "cafe-easypay-pid-second", "cafe-easypay-pkey-first", "cafe-easypay-trade-forged")
	forgedResponse := performCafeEasyPaySignedWebhookRequest(t, router, forgedPayload)
	require.Equal(t, http.StatusBadRequest, forgedResponse.Code)
	require.Equal(t, "verify failed", forgedResponse.Body.String())
	require.NotEqual(t, firstInstance.ID, secondInstance.ID)
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, forgedOrder.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, forgedOrder.ID))
}

func TestCafeEasyPayWebhookExtractOutTradeNo(t *testing.T) {
	require.Equal(t, "cafe_easypay_order_1", extractOutTradeNo("out_trade_no=cafe_easypay_order_1&trade_status=TRADE_SUCCESS", payment.TypeEasyPay))
	require.Empty(t, extractOutTradeNo("trade_status=TRADE_SUCCESS", payment.TypeEasyPay))
	require.Empty(t, extractOutTradeNo("out_trade_no=&trade_status=TRADE_SUCCESS", payment.TypeEasyPay))
	require.Empty(t, extractOutTradeNo("out_trade_no=%ZZ", payment.TypeEasyPay))
}

func TestCafeAlipayWebhookPinnedInstance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := newCafeWebhookHandlerTestClient(t)
	const encryptionKey = "0123456789abcdef0123456789abcdef"
	firstPrivateKey := newCafeAlipayWebhookRSAKey(t)
	secondPrivateKey := newCafeAlipayWebhookRSAKey(t)
	firstInstance := createCafeAlipayWebhookInstance(t, client, "cafe-alipay-first", "cafe-alipay-app-first", firstPrivateKey, 10, []byte(encryptionKey))
	secondInstance := createCafeAlipayWebhookInstance(t, client, "cafe-alipay-second", "cafe-alipay-app-second", secondPrivateKey, 20, []byte(encryptionKey))
	validOrder := bindCafeAlipayWebhookOrder(t, client, createCafeAlipayWebhookGroupBuyOrder(t, client, "cafe_alipay_pinned_valid_1"), secondInstance.ID)
	forgedOrder := bindCafeAlipayWebhookOrder(t, client, createCafeAlipayWebhookGroupBuyOrder(t, client, "cafe_alipay_pinned_forged_1"), secondInstance.ID)
	router, recorder := newCafeAlipayWebhookRouter(client, []byte(encryptionKey))

	validPayload := cafeAlipayWebhookPayload(t, validOrder, "cafe-alipay-app-second", secondPrivateKey, "cafe-alipay-trade-valid")
	validResponse := performCafeAlipaySignedWebhookRequest(t, router, validPayload)
	require.Equal(t, http.StatusOK, validResponse.Code)
	require.Equal(t, "success", validResponse.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, validOrder.ID, service.OrderStatusCompleted)
	validAuditCount := cafeWebhookAuditCount(t, client, validOrder.ID)
	require.NotZero(t, validAuditCount)

	validReplay := performCafeAlipaySignedWebhookRequest(t, router, validPayload)
	require.Equal(t, http.StatusOK, validReplay.Code)
	require.Equal(t, "success", validReplay.Body.String())
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	require.Equal(t, validAuditCount, cafeWebhookAuditCount(t, client, validOrder.ID))

	forgedPayload := cafeAlipayWebhookPayload(t, forgedOrder, "cafe-alipay-app-second", firstPrivateKey, "cafe-alipay-trade-forged")
	forgedResponse := performCafeAlipaySignedWebhookRequest(t, router, forgedPayload)
	require.Equal(t, http.StatusBadRequest, forgedResponse.Code)
	require.Equal(t, "verify failed", forgedResponse.Body.String())
	require.NotEqual(t, firstInstance.ID, secondInstance.ID)
	require.Equal(t, []int64{validOrder.ID}, recorder.paidOrderIDs)
	assertCafeWebhookOrderStatus(t, client, forgedOrder.ID, service.OrderStatusPending)
	require.Zero(t, cafeWebhookAuditCount(t, client, forgedOrder.ID))
}

func TestCafeAlipayWebhookExtractOutTradeNo(t *testing.T) {
	require.Equal(t, "cafe_alipay_order_1", extractOutTradeNo("out_trade_no=cafe_alipay_order_1&trade_status=TRADE_SUCCESS", payment.TypeAlipay))
	require.Empty(t, extractOutTradeNo("trade_status=TRADE_SUCCESS", payment.TypeAlipay))
	require.Empty(t, extractOutTradeNo("out_trade_no=&trade_status=TRADE_SUCCESS", payment.TypeAlipay))
	require.Empty(t, extractOutTradeNo("out_trade_no=%ZZ", payment.TypeAlipay))
}

func newCafeWebhookStripeRouter(client *dbent.Client, provider payment.Provider) (*gin.Engine, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	registry.Register(provider)
	paymentSvc := service.NewPaymentService(client, registry, nil, nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.POST("/api/v1/payment/webhook/stripe", webhookHandler.StripeWebhook)
	return router, recorder
}

func newCafePinnedStripeWebhookRouter(client *dbent.Client, encryptionKey []byte) (*gin.Engine, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	paymentSvc := service.NewPaymentService(client, registry, payment.NewDefaultLoadBalancer(client, encryptionKey), nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.POST("/api/v1/payment/webhook/stripe", webhookHandler.StripeWebhook)
	return router, recorder
}

func newCafePinnedAirwallexWebhookRouter(client *dbent.Client, encryptionKey []byte) (*gin.Engine, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	paymentSvc := service.NewPaymentService(client, registry, payment.NewDefaultLoadBalancer(client, encryptionKey), nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.POST("/api/v1/payment/webhook/airwallex", webhookHandler.AirwallexWebhook)
	return router, recorder
}

func newCafeWxpayWebhookRouter(client *dbent.Client, encryptionKey []byte) (*gin.Engine, *service.PaymentService, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	paymentSvc := service.NewPaymentService(client, registry, payment.NewDefaultLoadBalancer(client, encryptionKey), nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.POST("/api/v1/payment/webhook/wxpay", webhookHandler.WxpayNotify)
	return router, paymentSvc, recorder
}

func newCafeEasyPayWebhookRouter(client *dbent.Client, encryptionKey []byte) (*gin.Engine, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	paymentSvc := service.NewPaymentService(client, registry, payment.NewDefaultLoadBalancer(client, encryptionKey), nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.GET("/api/v1/payment/webhook/easypay", webhookHandler.EasyPayNotify)
	router.POST("/api/v1/payment/webhook/easypay", webhookHandler.EasyPayNotify)
	return router, recorder
}

func newCafeAlipayWebhookRouter(client *dbent.Client, encryptionKey []byte) (*gin.Engine, *cafeWebhookGroupBuyRecorder) {
	registry := payment.NewRegistry()
	paymentSvc := service.NewPaymentService(client, registry, payment.NewDefaultLoadBalancer(client, encryptionKey), nil, nil, nil, nil, nil, nil)
	recorder := &cafeWebhookGroupBuyRecorder{}
	paymentSvc.SetGroupBuyFulfillment(recorder)
	webhookHandler := NewPaymentWebhookHandler(paymentSvc, registry)
	router := gin.New()
	router.POST("/api/v1/payment/webhook/alipay", webhookHandler.AlipayNotify)
	return router, recorder
}

func performCafeWebhookStripeRequest(t *testing.T, router http.Handler, payload, signature string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook/stripe", strings.NewReader(payload))
	request.Header.Set("X-Cafe-Signature", signature)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performCafeStripeSignedWebhookRequest(t *testing.T, router http.Handler, payload, webhookSecret string) *httptest.ResponseRecorder {
	t.Helper()
	signed := stripewebhook.GenerateTestSignedPayload(&stripewebhook.UnsignedPayload{
		Payload: []byte(payload),
		Secret:  webhookSecret,
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook/stripe", strings.NewReader(payload))
	request.Header.Set("Stripe-Signature", signed.Header)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performCafeAirwallexSignedWebhookRequest(t *testing.T, router http.Handler, payload, timestamp, signature string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook/airwallex", strings.NewReader(payload))
	request.Header.Set("X-Timestamp", timestamp)
	request.Header.Set("X-Signature", signature)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performCafeWxpaySignedWebhookRequest(t *testing.T, router http.Handler, payload string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook/wxpay", strings.NewReader(payload))
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performCafeEasyPaySignedWebhookRequest(t *testing.T, router http.Handler, payload string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/payment/webhook/easypay?"+payload, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func performCafeAlipaySignedWebhookRequest(t *testing.T, router http.Handler, payload string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhook/alipay", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertCafeWxpaySuccessResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var body wxpaySuccessResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, wxpaySuccessCode, body.Code)
	require.Equal(t, wxpaySuccessMessage, body.Message)
}

func newCafeWebhookHandlerTestClient(t *testing.T) *dbent.Client {
	t.Helper()
	databaseName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := sql.Open("sqlite", "file:"+databaseName+"?mode=memory&cache=shared&_fk=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)
	driver := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(driver)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func createCafeWebhookGroupBuyOrder(t *testing.T, client *dbent.Client, outTradeNo string) *dbent.PaymentOrder {
	t.Helper()
	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail(outTradeNo + "@example.com").
		SetPasswordHash("hash").
		SetUsername(outTradeNo).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(12).
		SetPayAmount(12).
		SetFeeRate(0).
		SetRechargeCode(outTradeNo + "-code").
		SetOutTradeNo(outTradeNo).
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeGroupBuy).
		SetStatus(service.OrderStatusPending).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("cafe-webhook-test.invalid").
		SetProviderKey(payment.TypeStripe).
		Save(ctx)
	require.NoError(t, err)
	return order
}

func createCafeStripeWebhookInstance(t *testing.T, client *dbent.Client, name, webhookSecret string, encryptionKey []byte) *dbent.PaymentProviderInstance {
	t.Helper()
	configJSON, err := json.Marshal(map[string]string{
		"currency":      "USD",
		"secretKey":     "sk_test_" + name,
		"webhookSecret": webhookSecret,
	})
	require.NoError(t, err)
	encryptedConfig, err := payment.Encrypt(string(configJSON), encryptionKey)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName(name).
		SetConfig(encryptedConfig).
		SetSupportedTypes(payment.TypeStripe).
		SetEnabled(true).
		Save(context.Background())
	require.NoError(t, err)
	return instance
}

func createCafeAirwallexWebhookInstance(t *testing.T, client *dbent.Client, name, webhookSecret string, encryptionKey []byte) *dbent.PaymentProviderInstance {
	t.Helper()
	configJSON, err := json.Marshal(map[string]string{
		"apiBase":       "https://api-demo.airwallex.com/api/v1",
		"apiKey":        "cafe-airwallex-api-key-" + name,
		"clientId":      "cafe-airwallex-client-" + name,
		"currency":      "CNY",
		"webhookSecret": webhookSecret,
	})
	require.NoError(t, err)
	encryptedConfig, err := payment.Encrypt(string(configJSON), encryptionKey)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAirwallex).
		SetName(name).
		SetConfig(encryptedConfig).
		SetSupportedTypes(payment.TypeAirwallex).
		SetEnabled(true).
		Save(context.Background())
	require.NoError(t, err)
	return instance
}

func createCafeWxpayWebhookGroupBuyOrder(t *testing.T, client *dbent.Client, outTradeNo string) *dbent.PaymentOrder {
	t.Helper()
	order := createCafeWebhookGroupBuyOrder(t, client, outTradeNo)
	order, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetPaymentType(payment.TypeWxpay).
		SetProviderKey(payment.TypeWxpay).
		Save(context.Background())
	require.NoError(t, err)
	return order
}

func newCafeWxpayWebhookRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func createCafeWxpayWebhookInstance(t *testing.T, client *dbent.Client, name string, privateKey *rsa.PrivateKey, publicKeyID, apiV3Key string, sortOrder int, encryptionKey []byte) *dbent.PaymentProviderInstance {
	t.Helper()
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	configJSON, err := json.Marshal(map[string]string{
		"appId":       "wxpay-cafe-app-" + name,
		"mchId":       "wxpay-cafe-mch-" + name,
		"privateKey":  string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})),
		"apiV3Key":    apiV3Key,
		"certSerial":  "wxpay-cafe-cert-" + name,
		"publicKey":   string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})),
		"publicKeyId": publicKeyID,
	})
	require.NoError(t, err)
	encryptedConfig, err := payment.Encrypt(string(configJSON), encryptionKey)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeWxpay).
		SetName(name).
		SetConfig(encryptedConfig).
		SetSupportedTypes(payment.TypeWxpay).
		SetEnabled(true).
		SetSortOrder(sortOrder).
		Save(context.Background())
	require.NoError(t, err)
	return instance
}

func createCafeEasyPayWebhookGroupBuyOrder(t *testing.T, client *dbent.Client, outTradeNo string) *dbent.PaymentOrder {
	t.Helper()
	order := createCafeWebhookGroupBuyOrder(t, client, outTradeNo)
	order, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetPaymentType(payment.TypeAlipay).
		SetProviderKey(payment.TypeEasyPay).
		Save(context.Background())
	require.NoError(t, err)
	return order
}

func createCafeEasyPayWebhookInstance(t *testing.T, client *dbent.Client, name, pid, pkey string, sortOrder int, encryptionKey []byte) *dbent.PaymentProviderInstance {
	t.Helper()
	configJSON, err := json.Marshal(map[string]string{
		"pid":       pid,
		"pkey":      pkey,
		"apiBase":   "https://easypay-cafe-test.invalid",
		"notifyUrl": "https://cafe-test.invalid/api/v1/payment/webhook/easypay",
		"returnUrl": "https://cafe-test.invalid/payment/result",
	})
	require.NoError(t, err)
	encryptedConfig, err := payment.Encrypt(string(configJSON), encryptionKey)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeEasyPay).
		SetName(name).
		SetConfig(encryptedConfig).
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		SetSortOrder(sortOrder).
		Save(context.Background())
	require.NoError(t, err)
	return instance
}

func bindCafeEasyPayWebhookOrder(t *testing.T, client *dbent.Client, order *dbent.PaymentOrder, instanceID int64) *dbent.PaymentOrder {
	t.Helper()
	updated, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetProviderInstanceID(strconv.FormatInt(instanceID, 10)).
		SetProviderKey(payment.TypeEasyPay).
		Save(context.Background())
	require.NoError(t, err)
	return updated
}

func createCafeAlipayWebhookGroupBuyOrder(t *testing.T, client *dbent.Client, outTradeNo string) *dbent.PaymentOrder {
	t.Helper()
	order := createCafeWebhookGroupBuyOrder(t, client, outTradeNo)
	order, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetPaymentType(payment.TypeAlipay).
		SetProviderKey(payment.TypeAlipay).
		Save(context.Background())
	require.NoError(t, err)
	return order
}

func newCafeAlipayWebhookRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func createCafeAlipayWebhookInstance(t *testing.T, client *dbent.Client, name, appID string, privateKey *rsa.PrivateKey, sortOrder int, encryptionKey []byte) *dbent.PaymentProviderInstance {
	t.Helper()
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	configJSON, err := json.Marshal(map[string]string{
		"appId":      appID,
		"privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER})),
		"publicKey":  string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})),
		"notifyUrl":  "https://cafe-test.invalid/api/v1/payment/webhook/alipay",
		"returnUrl":  "https://cafe-test.invalid/payment/result",
	})
	require.NoError(t, err)
	encryptedConfig, err := payment.Encrypt(string(configJSON), encryptionKey)
	require.NoError(t, err)
	instance, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName(name).
		SetConfig(encryptedConfig).
		SetSupportedTypes(payment.TypeAlipay).
		SetEnabled(true).
		SetSortOrder(sortOrder).
		Save(context.Background())
	require.NoError(t, err)
	return instance
}

func bindCafeAlipayWebhookOrder(t *testing.T, client *dbent.Client, order *dbent.PaymentOrder, instanceID int64) *dbent.PaymentOrder {
	t.Helper()
	updated, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetProviderInstanceID(strconv.FormatInt(instanceID, 10)).
		SetProviderKey(payment.TypeAlipay).
		Save(context.Background())
	require.NoError(t, err)
	return updated
}

func bindCafeWebhookOrderToProvider(t *testing.T, client *dbent.Client, order *dbent.PaymentOrder, instanceID int64) *dbent.PaymentOrder {
	t.Helper()
	updated, err := client.PaymentOrder.UpdateOneID(order.ID).
		SetProviderInstanceID(strconv.FormatInt(instanceID, 10)).
		SetProviderKey(payment.TypeStripe).
		Save(context.Background())
	require.NoError(t, err)
	return updated
}

func cafeStripeWebhookPayload(t *testing.T, order *dbent.PaymentOrder, tradeNo string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"api_version": stripe.APIVersion,
		"id":          "evt_" + tradeNo,
		"object":      "event",
		"type":        "payment_intent.succeeded",
		"data": map[string]any{
			"object": map[string]any{
				"id":       tradeNo,
				"object":   "payment_intent",
				"amount":   1200,
				"currency": "usd",
				"metadata": map[string]string{"orderId": order.OutTradeNo},
			},
		},
	})
	require.NoError(t, err)
	return string(payload)
}

func cafeAirwallexWebhookPayload(t *testing.T, order *dbent.PaymentOrder, tradeNo string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"accountId": "acct_cafe_airwallex_synthetic",
		"data": map[string]any{
			"object": map[string]any{
				"amount":            order.PayAmount,
				"currency":          "CNY",
				"id":                tradeNo,
				"merchant_order_id": order.OutTradeNo,
				"status":            "SUCCEEDED",
			},
		},
		"id":   "evt_" + tradeNo,
		"name": "payment_intent.succeeded",
	})
	require.NoError(t, err)
	return string(payload)
}

func cafeAirwallexWebhookSignature(t *testing.T, payload, secret string) (string, string) {
	t.Helper()
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(timestamp))
	require.NoError(t, err)
	_, err = mac.Write([]byte(payload))
	require.NoError(t, err)
	return timestamp, hex.EncodeToString(mac.Sum(nil))
}

func cafeWxpayEncryptedWebhook(t *testing.T, order *dbent.PaymentOrder, signingKey *rsa.PrivateKey, publicKeyID, apiV3Key, tradeNo string) (string, map[string]string) {
	t.Helper()
	transactionJSON, err := json.Marshal(map[string]any{
		"appid":            "wxpay-cafe-app-test",
		"mchid":            "wxpay-cafe-mch-test",
		"out_trade_no":     order.OutTradeNo,
		"transaction_id":   tradeNo,
		"trade_state":      "SUCCESS",
		"trade_state_desc": "payment success",
		"amount": map[string]any{
			"total":    1200,
			"currency": "CNY",
		},
	})
	require.NoError(t, err)
	block, err := aes.NewCipher([]byte(apiV3Key))
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	const associatedData = "transaction"
	const resourceNonce = "cafeWxpayGCM"
	ciphertext := gcm.Seal(nil, []byte(resourceNonce), transactionJSON, []byte(associatedData))
	envelope, err := json.Marshal(map[string]any{
		"id":            "evt_" + tradeNo,
		"create_time":   time.Now().UTC().Format(time.RFC3339),
		"resource_type": "encrypt-resource",
		"event_type":    "TRANSACTION.SUCCESS",
		"summary":       "payment success",
		"resource": map[string]string{
			"original_type":   "transaction",
			"algorithm":       "AEAD_AES_256_GCM",
			"ciphertext":      base64.StdEncoding.EncodeToString(ciphertext),
			"associated_data": associatedData,
			"nonce":           resourceNonce,
		},
	})
	require.NoError(t, err)
	body := string(envelope)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	const signatureNonce = "cafe-wxpay-signature-nonce"
	return body, map[string]string{
		"Wechatpay-Timestamp": timestamp,
		"Wechatpay-Nonce":     signatureNonce,
		"Wechatpay-Serial":    publicKeyID,
		"Wechatpay-Signature": cafeWxpayWebhookSignature(t, signingKey, timestamp, signatureNonce, body),
	}
}

func cafeWxpayWebhookSignature(t *testing.T, privateKey *rsa.PrivateKey, timestamp, nonce, body string) string {
	t.Helper()
	digest := sha256.Sum256([]byte(timestamp + "\n" + nonce + "\n" + body + "\n"))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(signature)
}

func cafeEasyPayWebhookPayload(t *testing.T, order *dbent.PaymentOrder, pid, pkey, tradeNo string) string {
	t.Helper()
	params := map[string]string{
		"pid":          pid,
		"out_trade_no": order.OutTradeNo,
		"trade_no":     tradeNo,
		"trade_status": "TRADE_SUCCESS",
		"money":        "12.00",
		"sign_type":    "MD5",
	}
	params["sign"] = cafeEasyPayWebhookSignature(params, pkey)
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values.Encode()
}

func cafeEasyPayWebhookSignature(params map[string]string, pkey string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key == "sign" || key == "sign_type" || value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var canonical strings.Builder
	for index, key := range keys {
		if index > 0 {
			canonical.WriteByte('&')
		}
		canonical.WriteString(key)
		canonical.WriteByte('=')
		canonical.WriteString(params[key])
	}
	canonical.WriteString(pkey)
	digest := md5.Sum([]byte(canonical.String()))
	return hex.EncodeToString(digest[:])
}

func cafeAlipayWebhookPayload(t *testing.T, order *dbent.PaymentOrder, appID string, privateKey *rsa.PrivateKey, tradeNo string) string {
	t.Helper()
	values := url.Values{
		"app_id":       {appID},
		"charset":      {"utf-8"},
		"notify_id":    {"cafe-alipay-notify-" + tradeNo},
		"notify_time":  {"2026-08-03 23:15:00"},
		"notify_type":  {"trade_status_sync"},
		"out_trade_no": {order.OutTradeNo},
		"sign_type":    {"RSA2"},
		"total_amount": {"12.00"},
		"trade_no":     {tradeNo},
		"trade_status": {"TRADE_SUCCESS"},
		"version":      {"1.0"},
	}
	values.Set("sign", cafeAlipayWebhookSignature(t, values, privateKey))
	return values.Encode()
}

func cafeAlipayWebhookSignature(t *testing.T, values url.Values, privateKey *rsa.PrivateKey) string {
	t.Helper()
	pairs := make([]string, 0, len(values))
	for key, items := range values {
		if key == "sign" || key == "sign_type" || key == "alipay_cert_sn" {
			continue
		}
		for _, value := range items {
			pairs = append(pairs, key+"="+value)
		}
	}
	sort.Strings(pairs)
	digest := sha256.Sum256([]byte(strings.Join(pairs, "&")))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(signature)
}

func assertCafeWebhookOrderStatus(t *testing.T, client *dbent.Client, orderID int64, wantStatus string) {
	t.Helper()
	order, err := client.PaymentOrder.Get(context.Background(), orderID)
	require.NoError(t, err)
	require.Equal(t, wantStatus, order.Status)
}

func cafeWebhookAuditCount(t *testing.T, client *dbent.Client, orderID int64) int {
	t.Helper()
	count, err := client.PaymentAuditLog.Query().Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(orderID, 10))).Count(context.Background())
	require.NoError(t, err)
	return count
}

type cafeWebhookVerifierRequest struct {
	rawBody string
	headers map[string]string
}

type cafeWebhookVerifierStub struct {
	expectedBody      string
	acceptedSignature string
	notification      *payment.PaymentNotification
	requests          []cafeWebhookVerifierRequest
}

func (p *cafeWebhookVerifierStub) Name() string { return payment.TypeStripe }

func (p *cafeWebhookVerifierStub) ProviderKey() string { return payment.TypeStripe }

func (p *cafeWebhookVerifierStub) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}

func (p *cafeWebhookVerifierStub) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, errors.New("unexpected payment creation")
}

func (p *cafeWebhookVerifierStub) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, errors.New("unexpected payment order query")
}

func (p *cafeWebhookVerifierStub) VerifyNotification(_ context.Context, rawBody string, headers map[string]string) (*payment.PaymentNotification, error) {
	clonedHeaders := make(map[string]string, len(headers))
	for key, value := range headers {
		clonedHeaders[key] = value
	}
	p.requests = append(p.requests, cafeWebhookVerifierRequest{rawBody: rawBody, headers: clonedHeaders})
	if rawBody != p.expectedBody {
		return nil, errors.New("unexpected webhook body")
	}
	if headers["x-cafe-signature"] != p.acceptedSignature {
		return nil, errors.New("invalid webhook signature")
	}
	return p.notification, nil
}

func (p *cafeWebhookVerifierStub) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, errors.New("unexpected refund")
}

type cafeWebhookGroupBuyRecorder struct {
	paidOrderIDs []int64
}

func (r *cafeWebhookGroupBuyRecorder) HandleGroupBuyOrderPaid(_ context.Context, orderID int64) error {
	r.paidOrderIDs = append(r.paidOrderIDs, orderID)
	return nil
}

func (r *cafeWebhookGroupBuyRecorder) ReleaseGroupBuySeatForOrder(context.Context, int64, string) error {
	return errors.New("unexpected group buy seat release")
}
