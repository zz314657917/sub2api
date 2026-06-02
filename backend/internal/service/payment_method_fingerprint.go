package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const paymentMethodFingerprintPrefix = "pmf_"

func paymentMethodFingerprint(providerKey string, metadata map[string]string) string {
	scope, value := paymentMethodFingerprintSource(providerKey, metadata)
	if scope == "" || value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(scope + "|" + value))
	return paymentMethodFingerprintPrefix + hex.EncodeToString(sum[:])
}

func paymentMethodFingerprintSource(providerKey string, metadata map[string]string) (string, string) {
	if len(metadata) == 0 {
		return "", ""
	}
	providerKey = strings.ToLower(strings.TrimSpace(providerKey))
	if providerKey == "" {
		providerKey = "unknown"
	}

	if v := metadataValue(metadata, "payment_method_fingerprint"); v != "" {
		return providerKey + "|provided_payment_method_fingerprint", v
	}
	if v := metadataValue(metadata, "stripe_card_fingerprint", "card_fingerprint"); v != "" {
		return providerKey + "|stripe_card_fingerprint", v
	}
	if v := metadataValue(metadata, "alipay_buyer_id", "buyer_id", "buyer_user_id"); v != "" {
		return providerKey + "|alipay_buyer_id", v
	}
	if v := metadataValue(metadata, "alipay_buyer_open_id", "buyer_open_id"); v != "" {
		return providerKey + "|alipay_buyer_open_id", v
	}
	if v := metadataValue(metadata, "wxpay_openid", "payer_openid", "openid"); v != "" {
		appID := metadataValue(metadata, "appid", "app_id")
		merchantID := metadataValue(metadata, "mchid", "merchant_id")
		return providerKey + "|wxpay_openid|" + appID + "|" + merchantID, v
	}
	return "", ""
}

func metadataValue(metadata map[string]string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(metadata[key]); v != "" {
			return v
		}
	}
	return ""
}
