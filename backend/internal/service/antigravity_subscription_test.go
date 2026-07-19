package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
)

func TestS79NormalizeAntigravitySubscription(t *testing.T) {
	tests := []struct {
		name string
		resp *antigravity.LoadCodeAssistResponse
		want AntigravitySubscriptionResult
	}{
		{
			name: "paid Pro keeps plan and first ineligible reason",
			resp: &antigravity.LoadCodeAssistResponse{
				PaidTier: &antigravity.PaidTierInfo{ID: "g1-pro-tier"},
				IneligibleTiers: []*antigravity.IneligibleTier{
					{ReasonMessage: "location validation required"},
					{ReasonMessage: "ignored second reason"},
				},
			},
			want: AntigravitySubscriptionResult{
				PlanType:           "Pro",
				SubscriptionStatus: "abnormal",
				SubscriptionError:  "location validation required",
			},
		},
		{
			name: "paid Ultra keeps plan with ineligible tier",
			resp: &antigravity.LoadCodeAssistResponse{
				PaidTier: &antigravity.PaidTierInfo{ID: "g1-ultra-tier"},
				IneligibleTiers: []*antigravity.IneligibleTier{
					{ReasonMessage: "account verification required"},
				},
			},
			want: AntigravitySubscriptionResult{
				PlanType:           "Ultra",
				SubscriptionStatus: "abnormal",
				SubscriptionError:  "account verification required",
			},
		},
		{
			name: "free tier with ineligible result is abnormal",
			resp: &antigravity.LoadCodeAssistResponse{
				CurrentTier: &antigravity.TierInfo{ID: "free-tier"},
				IneligibleTiers: []*antigravity.IneligibleTier{
					{ReasonMessage: "some warning"},
				},
			},
			want: AntigravitySubscriptionResult{
				PlanType:           "Abnormal",
				SubscriptionStatus: "abnormal",
				SubscriptionError:  "some warning",
			},
		},
		{
			name: "unknown tier with ineligible result is abnormal",
			resp: &antigravity.LoadCodeAssistResponse{
				CurrentTier: &antigravity.TierInfo{ID: "future-tier"},
				IneligibleTiers: []*antigravity.IneligibleTier{
					{ReasonMessage: "unsupported tier"},
				},
			},
			want: AntigravitySubscriptionResult{
				PlanType:           "Abnormal",
				SubscriptionStatus: "abnormal",
				SubscriptionError:  "unsupported tier",
			},
		},
		{
			name: "no ineligible tier keeps normalized plan",
			resp: &antigravity.LoadCodeAssistResponse{
				PaidTier: &antigravity.PaidTierInfo{ID: "g1-ultra-tier"},
			},
			want: AntigravitySubscriptionResult{PlanType: "Ultra"},
		},
		{
			name: "nil response is free",
			resp: nil,
			want: AntigravitySubscriptionResult{PlanType: "Free"},
		},
		{
			name: "no tier with ineligible result is abnormal",
			resp: &antigravity.LoadCodeAssistResponse{
				IneligibleTiers: []*antigravity.IneligibleTier{
					{ReasonMessage: "unknown issue"},
				},
			},
			want: AntigravitySubscriptionResult{
				PlanType:           "Abnormal",
				SubscriptionStatus: "abnormal",
				SubscriptionError:  "unknown issue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeAntigravitySubscription(tt.resp); got != tt.want {
				t.Fatalf("NormalizeAntigravitySubscription() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
