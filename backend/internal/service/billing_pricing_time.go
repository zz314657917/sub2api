package service

import (
	"context"
	"regexp"
	"strings"
	"time"
)

type billingPricingTimeKey struct{}

// Capture once at the public recorder boundary, including requests without a
// handler timestamp. Context carries that instant through legacy helper paths.
func (s *BillingService) captureBillingPricingTime(ctx context.Context, started time.Time) context.Context {
	if started.IsZero() {
		if s != nil && s.pricingNow != nil {
			started = s.pricingNow()
		} else {
			started = billingReferenceTime(started)
		}
	}
	return context.WithValue(ctx, billingPricingTimeKey{}, started)
}

func requestPricingTime(ctx context.Context) time.Time {
	if ctx != nil {
		if at, ok := ctx.Value(billingPricingTimeKey{}).(time.Time); ok {
			return at
		}
	}
	return time.Time{}
}

var deepseekKnownPricingName = regexp.MustCompile(`^deepseek-(?:flash|v4(?:\.1)?-flash(?:-vision-exp)?|v4-pro|chat|reasoner)(?:-(?:[0-9]{4}|[0-9]{8}))?$`)

func deepseekPricingModel(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))
	if deepseekKnownPricingName.MatchString(model) {
		return model
	}
	return ""
}

// Pro uses the user-selected nominal card, not the announced Flash transition.
// Site credits map 1:1 to CNY: input 4.50/output 13.50/read .15 per million.
// These Pro amounts are charged directly in credits, without USD conversion.
func deepseekBasePricing(model string) *ModelPricing {
	model = deepseekPricingModel(model)
	if model == "" {
		return nil
	}
	input, output, read := .15e-6, .60e-6, .003e-6
	if strings.HasPrefix(model, "deepseek-v4-pro") {
		input, output, read = 4.50e-6, 13.50e-6, .15e-6
	}
	return &ModelPricing{
		InputPricePerToken: input, OutputPricePerToken: output,
		CacheReadPricePerToken: read,
	}
}

func deepseekDefaultPricingAt(model string, at time.Time) *ModelPricing {
	pricing := deepseekBasePricing(model)
	if pricing == nil {
		return nil
	}
	factor := deepseekPeakMultiplierAt(billingReferenceTime(at))
	pricing.InputPricePerToken *= factor
	pricing.OutputPricePerToken *= factor
	pricing.CacheReadPricePerToken *= factor
	return pricing
}

// The existing weekday schedule has no public-holiday calendar.
func deepseekPeakMultiplierAt(at time.Time) float64 {
	day := at.In(time.FixedZone("Asia/Shanghai", 8*3600)).Weekday()
	hour := at.UTC().Hour()
	if day != time.Saturday && day != time.Sunday &&
		((hour >= 1 && hour < 4) || (hour >= 6 && hour < 10)) {
		return 2
	}
	return 1
}

func defaultPricingAt(model string, pricing *ModelPricing, at time.Time) *ModelPricing {
	if ds := deepseekDefaultPricingAt(model, at); ds != nil {
		return ds
	}
	return pricing
}
