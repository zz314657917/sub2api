package domain

// APIKeyMultiGroupRoute describes one selectable group route for an API key.
type APIKeyMultiGroupRoute struct {
	GroupID         int64 `json:"group_id"`
	Priority        int   `json:"priority"`
	Weight          int   `json:"weight"`
	CooldownSeconds int   `json:"cooldown_seconds"`
	// FirstResponseTimeoutSeconds enables same-request next-group failover
	// before the first valid response. Zero keeps the legacy behavior.
	FirstResponseTimeoutSeconds int      `json:"first_response_timeout_seconds,omitempty"`
	Enabled                     bool     `json:"enabled"`
	ModelPatterns               []string `json:"model_patterns,omitempty"`
	ImageOnly                   bool     `json:"image_only,omitempty"`
	TextOnly                    bool     `json:"text_only,omitempty"`
}
