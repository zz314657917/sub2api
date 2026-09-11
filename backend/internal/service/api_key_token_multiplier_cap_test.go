package service

import (
	"context"
	"testing"
)

func TestTokenMultiplierCap(t *testing.T) {
	for _, tc := range []struct {
		name      string
		rate, cap float64
		rejected  bool
	}{
		{"reject rather than discount", .25, .01, true},
		{"equal allowed", .15, .15, false},
		{"lower allowed", .08, .15, false},
		{"disabled", .25, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &BillingCacheService{}
			group := &Group{ID: 1, RateMultiplier: tc.rate}
			key := &APIKey{TokenMultiplierCap: tc.cap}
			err := s.checkTokenMultiplierCap(context.Background(), &User{ID: 1}, key, group)
			if (err != nil) != tc.rejected {
				t.Fatalf("error = %v, rejected = %v", err, tc.rejected)
			}
			if group.RateMultiplier != tc.rate {
				t.Fatal("protection modified billing rate")
			}
		})
	}
}
