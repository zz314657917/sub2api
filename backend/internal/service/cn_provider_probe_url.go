package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/util/urlvalidator"
)

func cnValidateProbeURL(cfg *config.Config, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("probe url is required")
	}
	if cfg != nil && cfg.Security.URLAllowlist.Enabled {
		value, err := urlvalidator.ValidateHTTPSURL(raw, urlvalidator.ValidationOptions{AllowedHosts: cfg.Security.URLAllowlist.UpstreamHosts, RequireAllowlist: true, AllowPrivate: cfg.Security.URLAllowlist.AllowPrivateHosts})
		if err != nil {
			return "", fmt.Errorf("probe target rejected by URL security policy: %w", err)
		}
		return value, nil
	}
	allowHTTP := cfg != nil && cfg.Security.URLAllowlist.AllowInsecureHTTP
	value, err := urlvalidator.ValidateURLFormat(raw, allowHTTP)
	if err != nil {
		return "", fmt.Errorf("probe target rejected by URL security policy: %w", err)
	}
	return value, nil
}
