package service

import "time"

// UserErrorRequest is the user-safe whitelist view of an ops error record.
type UserErrorRequest struct {
	ID              int64     `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	Model           string    `json:"model"`
	InboundEndpoint string    `json:"inbound_endpoint"`
	StatusCode      int       `json:"status_code"`
	Category        string    `json:"category"`
	Platform        string    `json:"platform"`
	Message         string    `json:"message"`
	KeyName         string    `json:"key_name"`
	KeyDeleted      bool      `json:"key_deleted"`
}

type UserErrorRequestList struct {
	Items    []*UserErrorRequest `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

func MapUserErrorCategory(phase, errType string) string {
	switch phase {
	case "auth":
		return "auth"
	case "routing":
		return "service_unavailable"
	case "upstream", "network":
		return "upstream"
	case "internal":
		return "internal"
	case "request":
		switch errType {
		case "rate_limit_error":
			return "rate_limit"
		case "billing_error", "subscription_error":
			return "quota"
		case "invalid_request_error":
			return "invalid_request"
		case "cyber_policy":
			return "cyber"
		}
	}
	return "other"
}

func CategoryToFilter(category string) (phases []string, errorTypes []string) {
	switch category {
	case "auth":
		return []string{"auth"}, nil
	case "service_unavailable":
		return []string{"routing"}, nil
	case "upstream":
		return []string{"upstream", "network"}, nil
	case "internal":
		return []string{"internal"}, nil
	case "rate_limit":
		return nil, []string{"rate_limit_error"}
	case "quota":
		return nil, []string{"billing_error", "subscription_error"}
	case "invalid_request":
		return nil, []string{"invalid_request_error"}
	case "cyber":
		return []string{"request"}, []string{"cyber_policy"}
	default:
		return nil, nil
	}
}

func ToUserErrorRequest(entry *OpsErrorLog) *UserErrorRequest {
	if entry == nil {
		return nil
	}
	model := entry.RequestedModel
	if model == "" {
		model = entry.Model
	}
	return &UserErrorRequest{
		ID:              entry.ID,
		CreatedAt:       entry.CreatedAt,
		Model:           model,
		InboundEndpoint: entry.InboundEndpoint,
		StatusCode:      entry.StatusCode,
		Category:        MapUserErrorCategory(entry.Phase, entry.Type),
		Platform:        entry.Platform,
		Message:         entry.Message,
		KeyName:         entry.APIKeyName,
		KeyDeleted:      entry.APIKeyDeleted,
	}
}

type UserErrorRequestDetail struct {
	UserErrorRequest
	ErrorBody          string `json:"error_body"`
	UpstreamStatusCode *int   `json:"upstream_status_code,omitempty"`
}

func ToUserErrorRequestDetail(entry *OpsErrorLogDetail) *UserErrorRequestDetail {
	if entry == nil {
		return nil
	}
	base := ToUserErrorRequest(&entry.OpsErrorLog)
	return &UserErrorRequestDetail{
		UserErrorRequest:   *base,
		ErrorBody:          entry.ErrorBody,
		UpstreamStatusCode: entry.UpstreamStatusCode,
	}
}
