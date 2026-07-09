package domain

// GroupBuyTierRule maps an inclusive share range to one subscription group.
type GroupBuyTierRule struct {
	MinShares     int    `json:"min_shares"`
	MaxShares     int    `json:"max_shares"`
	TargetGroupID int64  `json:"target_group_id"`
	Label         string `json:"label,omitempty"`
}

// GroupBuyPolicySnapshot stores the entitlement policy used for a share batch.
type GroupBuyPolicySnapshot struct {
	ProductKey          string             `json:"product_key,omitempty"`
	PlanID              int64              `json:"plan_id,omitempty"`
	TotalShares         int                `json:"total_shares,omitempty"`
	QuotaPerShareLabel  string             `json:"quota_per_share_label,omitempty"`
	TierRules           []GroupBuyTierRule `json:"tier_rules,omitempty"`
	LegacyTierGroupIDs  map[string]int64   `json:"tier_group_ids,omitempty"`
	TargetGroupID       int64              `json:"target_group_id,omitempty"`
	CapturedAtUnixMilli int64              `json:"captured_at_unix_milli,omitempty"`
}
