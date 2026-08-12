package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func newAccountWithAvailability(enabled bool, start, end string) *Account {
	return &Account{
		Status:      StatusActive,
		Schedulable: true,
		Extra: map[string]any{
			accountAvailabilityEnabledExtraKey: enabled,
			accountAvailabilityStartExtraKey:   start,
			accountAvailabilityEndExtraKey:     end,
		},
	}
}

func TestAccountAvailabilityBoundaries(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "18:00", "22:00")
	cases := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "before start", at: at(17, 59), want: false},
		{name: "start included", at: at(18, 0), want: true},
		{name: "inside", at: at(21, 59), want: true},
		{name: "end excluded", at: at(22, 0), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, account.IsSchedulableAt(tc.at))
		})
	}
}

func TestAccountAvailabilityRespectsExistingUnschedulableState(t *testing.T) {
	setTestTimezone(t, "UTC")
	timeInside := at(19, 0)
	account := newAccountWithAvailability(true, "18:00", "22:00")
	account.Status = "inactive"
	require.False(t, account.IsSchedulableAt(timeInside))

	account.Status = StatusActive
	account.Schedulable = false
	require.False(t, account.IsSchedulableAt(timeInside))
}

func TestAccountAvailabilityUsesRequestStartedAt(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "18:00", "22:00")
	ctx := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(21, 59))

	require.True(t, account.IsSchedulableWithContext(ctx))
	require.False(t, account.IsSchedulableAt(at(22, 0)))
}

func TestAccountAvailabilityExcludesAntigravityModelCandidateOutsideWindow(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "18:00", "22:00")
	account.Platform = PlatformAntigravity

	outsideWindow := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(17, 59))
	insideWindow := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(18, 0))

	require.False(t, account.IsSchedulableForModelWithContext(outsideWindow, "gemini-3-flash"))
	require.True(t, account.IsSchedulableForModelWithContext(insideWindow, "gemini-3-flash"))
}

func TestAccountAvailabilityInvalidLegacyConfigurationFailsOpen(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "22:00", "18:00")

	require.True(t, account.IsSchedulableAt(at(12, 0)))
	require.Error(t, ValidateAccountAvailabilityConfig(account.Extra))
}

func TestAccountAvailabilityLegacySingleDigitHourRemainsCompatible(t *testing.T) {
	setTestTimezone(t, "UTC")
	account := newAccountWithAvailability(true, "9:00", "18:00")

	require.False(t, account.IsSchedulableAt(at(8, 59)))
	require.True(t, account.IsSchedulableAt(at(9, 0)))
	require.False(t, account.IsSchedulableAt(at(18, 0)))
	require.Error(t, ValidateAccountAvailabilityConfig(account.Extra))
}

func TestValidateAccountAvailabilityConfig(t *testing.T) {
	cases := []struct {
		name    string
		extra   map[string]any
		wantErr bool
	}{
		{name: "absent", extra: nil},
		{name: "disabled without window", extra: map[string]any{accountAvailabilityEnabledExtraKey: false}},
		{name: "disabled valid window retained", extra: map[string]any{
			accountAvailabilityEnabledExtraKey: false,
			accountAvailabilityStartExtraKey:   "18:00",
			accountAvailabilityEndExtraKey:     "22:00",
		}},
		{name: "enabled valid", extra: map[string]any{
			accountAvailabilityEnabledExtraKey: true,
			accountAvailabilityStartExtraKey:   "18:00",
			accountAvailabilityEndExtraKey:     "22:00",
		}},
		{name: "enabled missing window", extra: map[string]any{accountAvailabilityEnabledExtraKey: true}, wantErr: true},
		{name: "non boolean enabled", extra: map[string]any{accountAvailabilityEnabledExtraKey: "true"}, wantErr: true},
		{name: "partial window", extra: map[string]any{accountAvailabilityStartExtraKey: "18:00"}, wantErr: true},
		{name: "bad time", extra: map[string]any{
			accountAvailabilityStartExtraKey: "25:00",
			accountAvailabilityEndExtraKey:   "22:00",
		}, wantErr: true},
		{name: "equal window", extra: map[string]any{
			accountAvailabilityStartExtraKey: "18:00",
			accountAvailabilityEndExtraKey:   "18:00",
		}, wantErr: true},
		{name: "cross midnight", extra: map[string]any{
			accountAvailabilityStartExtraKey: "22:00",
			accountAvailabilityEndExtraKey:   "02:00",
		}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAccountAvailabilityConfig(tc.extra)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestFilterAccountsForSchedulingAppliesRequestAvailability(t *testing.T) {
	setTestTimezone(t, "UTC")
	ctx := context.WithValue(context.Background(), ctxkey.RequestStartedAt, at(17, 59))
	accounts := []Account{
		*newAccountWithAvailability(true, "18:00", "22:00"),
		{Status: StatusActive, Schedulable: true},
	}

	filtered := filterAccountsForScheduling(ctx, accounts)
	require.Len(t, filtered, 1)
	require.Nil(t, filtered[0].Extra)
}
