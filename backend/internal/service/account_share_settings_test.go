package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestSettingService_ParseAccountShareSettings(t *testing.T) {
	svc := NewSettingService(nil, &config.Config{})
	settings := svc.parseSettings(map[string]string{
		SettingKeyAccountShareEnabled:          "true",
		SettingKeyAccountShareOwnerRate:        "85.5",
		SettingKeyAccountShareFreezeHours:      "24",
		SettingKeyAccountShareAutoReview:       "true",
		SettingKeyAccountShareUserAccountLimit: "8",
	})

	if !settings.AccountShareEnabled {
		t.Fatal("expected account sharing enabled")
	}
	if settings.AccountShareOwnerRatePercent != 85.5 {
		t.Fatalf("owner rate = %v", settings.AccountShareOwnerRatePercent)
	}
	if settings.AccountShareFreezeHours != 24 {
		t.Fatalf("freeze hours = %d", settings.AccountShareFreezeHours)
	}
	if !settings.AccountShareAutoReview {
		t.Fatal("expected auto review enabled")
	}
	if settings.AccountShareUserAccountLimit != 8 {
		t.Fatalf("user account limit = %d", settings.AccountShareUserAccountLimit)
	}
}

func TestSettingService_BuildAccountShareSettingsUpdatesClampsValues(t *testing.T) {
	svc := NewSettingService(nil, &config.Config{})
	updates, err := svc.buildSystemSettingsUpdates(context.Background(), &SystemSettings{
		AccountShareEnabled:          true,
		AccountShareOwnerRatePercent: 150,
		AccountShareFreezeHours:      AccountShareFreezeHoursMax + 1,
		AccountShareAutoReview:       true,
		AccountShareUserAccountLimit: AccountShareUserAccountLimitMax + 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	if updates[SettingKeyAccountShareEnabled] != "true" {
		t.Fatalf("share enabled update = %q", updates[SettingKeyAccountShareEnabled])
	}
	if updates[SettingKeyAccountShareOwnerRate] != "100.00000000" {
		t.Fatalf("owner rate update = %q", updates[SettingKeyAccountShareOwnerRate])
	}
	if updates[SettingKeyAccountShareFreezeHours] != "2160" {
		t.Fatalf("freeze hours update = %q", updates[SettingKeyAccountShareFreezeHours])
	}
	if updates[SettingKeyAccountShareAutoReview] != "true" {
		t.Fatalf("auto review update = %q", updates[SettingKeyAccountShareAutoReview])
	}
	if updates[SettingKeyAccountShareUserAccountLimit] != "100" {
		t.Fatalf("user account limit update = %q", updates[SettingKeyAccountShareUserAccountLimit])
	}
}
