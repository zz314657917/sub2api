//go:build unit

package admin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func bindSMTPRequest[T any](t *testing.T, body string) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	var payload T
	if err := c.ShouldBindJSON(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestTestSMTPRequestUseTLSOmissionFallsBackToSavedSetting(t *testing.T) {
	req := bindSMTPRequest[TestSMTPRequest](t, `{}`)
	if req.SMTPUseTLS != nil || !resolveSMTPUseTLS(req.SMTPUseTLS, &service.SMTPConfig{UseTLS: true}) {
		t.Fatal("omission did not fallback")
	}
}
func TestTestSMTPRequestExplicitFalseOverridesSavedUseTLS(t *testing.T) {
	req := bindSMTPRequest[TestSMTPRequest](t, `{"smtp_use_tls":false}`)
	if req.SMTPUseTLS == nil || resolveSMTPUseTLS(req.SMTPUseTLS, &service.SMTPConfig{UseTLS: true}) {
		t.Fatal("false did not override")
	}
}
func TestTestSMTPRequestExplicitTrueOverridesMissingSavedUseTLS(t *testing.T) {
	req := bindSMTPRequest[TestSMTPRequest](t, `{"smtp_use_tls":true}`)
	if req.SMTPUseTLS == nil || !resolveSMTPUseTLS(req.SMTPUseTLS, nil) {
		t.Fatal("true did not override")
	}
}
func TestSendTestEmailRequestPreservesUseTLSOmissionSemantics(t *testing.T) {
	omitted := bindSMTPRequest[SendTestEmailRequest](t, `{"email":"admin@example.com"}`)
	explicitFalse := bindSMTPRequest[SendTestEmailRequest](t, `{"email":"admin@example.com","smtp_use_tls":false}`)
	explicitTrue := bindSMTPRequest[SendTestEmailRequest](t, `{"email":"admin@example.com","smtp_use_tls":true}`)
	saved := &service.SMTPConfig{UseTLS: true}
	if !resolveSMTPUseTLS(omitted.SMTPUseTLS, saved) || resolveSMTPUseTLS(explicitFalse.SMTPUseTLS, saved) {
		t.Fatal("send-test-email TLS semantics incorrect")
	}
	if !resolveSMTPUseTLS(explicitTrue.SMTPUseTLS, &service.SMTPConfig{UseTLS: false}) || !resolveSMTPUseTLS(explicitTrue.SMTPUseTLS, nil) {
		t.Fatal("send-test-email explicit true did not override saved false or missing config")
	}
}
