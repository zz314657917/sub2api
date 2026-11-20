package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/gemini"
	"github.com/Wei-Shaw/sub2api/internal/pkg/googleapi"
	"github.com/Wei-Shaw/sub2api/internal/util/responseheaders"
)

// Keep the old exported helpers for callers outside the native handler.
func (s *GeminiMessagesCompatService) SelectAccountForAIStudioEndpoints(ctx context.Context, groupID *int64) (*Account, error) {
	return s.SelectAccountForModelsEndpoints(ctx, groupID)
}

func (s *GeminiMessagesCompatService) ForwardAIStudioGET(ctx context.Context, account *Account, path string) (*UpstreamHTTPResult, error) {
	if path == "/v1beta/models" {
		return s.ForwardListModels(ctx, account)
	}
	if strings.HasPrefix(path, "/v1beta/models/") {
		return s.ForwardGetModel(ctx, account, strings.TrimPrefix(path, "/v1beta/models/"))
	}
	return s.forwardGeminiModelsGET(ctx, account, path)
}

func (s *GeminiMessagesCompatService) ForwardListModels(ctx context.Context, account *Account) (*UpstreamHTTPResult, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	if !account.IsVertexServiceAccount() {
		return s.forwardGeminiModelsGET(ctx, account, "/v1beta/models")
	}
	modelIDs, failure, err := s.listVertexGeminiModels(ctx, account, vertexDiscoveryLocation(account.VertexLocation("")))
	if err != nil || failure != nil {
		return failure, err
	}
	models := make([]gemini.Model, 0, len(modelIDs))
	for _, id := range modelIDs {
		models = append(models, gemini.FallbackModel(id))
	}
	return geminiModelsJSONResult(http.StatusOK, gemini.ModelsListResponse{Models: models})
}

func (s *GeminiMessagesCompatService) ForwardGetModel(ctx context.Context, account *Account, model string) (*UpstreamHTTPResult, error) {
	if account == nil {
		return nil, errors.New("account is nil")
	}
	if !IsSafeGeminiModelPathSegment(model) {
		return nil, errors.New("invalid model in URL")
	}
	if !account.IsVertexServiceAccount() {
		return s.forwardGeminiModelsGET(ctx, account, "/v1beta/models/"+model)
	}
	modelIDs, failure, err := s.listVertexGeminiModels(ctx, account, vertexDiscoveryLocation(account.VertexLocation(model)))
	if err != nil || failure != nil {
		return failure, err
	}
	if slices.Contains(modelIDs, model) {
		return geminiModelsJSONResult(http.StatusOK, gemini.FallbackModel(model))
	}
	return geminiModelsJSONResult(http.StatusNotFound, map[string]any{
		"error": map[string]any{
			"code":    http.StatusNotFound,
			"message": fmt.Sprintf("models/%s is not found", model),
			"status":  googleapi.HTTPStatusToGoogleStatus(http.StatusNotFound),
		},
	})
}

func (s *GeminiMessagesCompatService) listVertexGeminiModels(ctx context.Context, account *Account, location string) ([]string, *UpstreamHTTPResult, error) {
	// Validate configuration before a token provider can perform any token exchange.
	if err := validateVertexPublisherAccount(account, location); err != nil {
		return nil, nil, err
	}
	if s.tokenProvider == nil {
		return nil, nil, errors.New("gemini token provider not configured")
	}
	accessToken, err := s.tokenProvider.GetAccessToken(ctx, account)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, nil, errors.New("no Vertex access token is available")
	}
	var failure *UpstreamHTTPResult
	models, err := listVertexGeminiPublisherModels(location, func(_ int, listURL string) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		res, err := s.doGeminiModelsGET(req, account)
		if err != nil {
			return nil, err
		}
		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
			failure = res
			return nil, errors.New("upstream model list returned non-2xx status")
		}
		return res.Body, nil
	})
	if failure != nil {
		return nil, failure, nil
	}
	return models, nil, err
}

func (s *GeminiMessagesCompatService) doGeminiModelsGET(req *http.Request, account *Account) (*UpstreamHTTPResult, error) {
	if s.httpUpstream == nil {
		return nil, errors.New("upstream HTTP client not configured")
	}
	resp, err := s.httpUpstream.Do(req, upstreamModelsProxyURL(account), account.ID, account.Concurrency)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := readModelsResponseBody(resp.Body, resolveModelsListReadLimit(s.cfg))
	if err != nil {
		return nil, err
	}
	headers := responseheaders.FilterHeaders(resp.Header, s.responseHeaderFilter)
	if values := resp.Header.Values("Www-Authenticate"); len(values) > 0 {
		headers["Www-Authenticate"] = append([]string(nil), values...)
	}
	return &UpstreamHTTPResult{StatusCode: resp.StatusCode, Headers: headers, Body: body}, nil
}

func geminiModelsJSONResult(status int, payload any) (*UpstreamHTTPResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &UpstreamHTTPResult{
		StatusCode: status,
		Headers:    http.Header{"Content-Type": {"application/json; charset=utf-8"}},
		Body:       body,
	}, nil
}
