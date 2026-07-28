package service

import (
	"context"
	"strconv"
	"strings"
)

const defaultAuditLogRetentionDays = 180

func (s *SettingService) IsSessionBindingEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeySessionBindingEnabled)
	return err == nil && value == "true"
}

func (s *SettingService) IsStepUpEnabled(ctx context.Context) bool {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyStepUpEnabled)
	return err == nil && value == "true"
}

func (s *SettingService) GetAuditLogRetentionDays(ctx context.Context) int {
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAuditLogRetentionDays)
	if err != nil {
		return defaultAuditLogRetentionDays
	}
	return parseAuditLogRetentionDays(value)
}

func parseAuditLogRetentionDays(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultAuditLogRetentionDays
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultAuditLogRetentionDays
	}
	if parsed < 0 {
		return 0
	}
	return parsed
}

func normalizeAuditLogRetentionDays(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
