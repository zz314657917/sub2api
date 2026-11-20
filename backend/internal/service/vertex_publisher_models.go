package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
)

const (
	vertexPublisherModelsPageSize    = 300
	vertexPublisherModelsMaxPages    = 20
	vertexPublisherPageTokenMaxBytes = 4096
	vertexGooglePublisherModelPrefix = "publishers/google/models/"
)

// The location is a DNS label component, never a supplied host or URL.
func buildVertexPublisherModelsURL(location, pageToken string) (string, error) {
	if !vertexLocationPattern.MatchString(location) ||
		strings.HasPrefix(location, "-") || strings.HasSuffix(location, "-") ||
		len(location)+len("-aiplatform") > 63 {
		return "", errors.New("invalid vertex location")
	}
	if len(pageToken) > vertexPublisherPageTokenMaxBytes {
		return "", errors.New("vertex model list page token is too large")
	}
	host := location + "-aiplatform.googleapis.com"
	if location == "global" {
		host = "aiplatform.googleapis.com"
	}
	query := url.Values{"pageSize": {strconv.Itoa(vertexPublisherModelsPageSize)}}
	if pageToken != "" {
		query.Set("pageToken", pageToken)
	}
	return "https://" + host + "/v1beta1/publishers/google/models?" + query.Encode(), nil
}

// fetchPage must read a bounded response and reject non-2xx before returning a page.
// Any failed or unfinished page discards the entire partial catalog.
func listVertexGeminiPublisherModels(location string, fetchPage func(int, string) ([]byte, error)) ([]string, error) {
	var models []string
	pageToken := ""
	seenTokens := make(map[string]bool)
	for page := 0; page < vertexPublisherModelsMaxPages; page++ {
		listURL, err := buildVertexPublisherModelsURL(location, pageToken)
		if err != nil {
			return nil, err
		}
		body, err := fetchPage(page, listURL)
		if err != nil {
			return nil, err
		}
		var response struct {
			PublisherModels []struct {
				Name string `json:"name"`
			} `json:"publisherModels"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("invalid vertex model list JSON: %w", err)
		}
		if response.PublisherModels == nil {
			return nil, errors.New("vertex model list is missing publisherModels")
		}
		for _, entry := range response.PublisherModels {
			if !strings.HasPrefix(entry.Name, vertexGooglePublisherModelPrefix) {
				continue
			}
			modelID := strings.TrimPrefix(entry.Name, vertexGooglePublisherModelPrefix)
			if strings.HasPrefix(modelID, "gemini-") && len(modelID) > len("gemini-") && IsSafeGeminiModelPathSegment(modelID) {
				models = append(models, modelID)
			}
		}
		pageToken = response.NextPageToken
		if pageToken == "" {
			if len(models) == 0 {
				return nil, errors.New("upstream returned no supported models")
			}
			return dedupeAndSortModelIDs(models), nil
		}
		if seenTokens[pageToken] {
			return nil, errors.New("vertex model list repeated page token")
		}
		seenTokens[pageToken] = true
	}
	return nil, fmt.Errorf("vertex model list pagination did not finish after %d pages", vertexPublisherModelsMaxPages)
}

func validateVertexPublisherAccount(account *Account, location string) error {
	if account == nil || !account.IsGemini() || !account.IsVertexServiceAccount() {
		return errors.New("not a Gemini Vertex JSON service account")
	}
	if _, err := buildVertexPublisherModelsURL(location, ""); err != nil {
		return err
	}
	_, err := parseVertexServiceAccountKey(account)
	return err
}

func readModelsResponseBody(body io.Reader, limit int64) ([]byte, error) {
	// LimitReader takes a signed count; avoid overflow for a misconfigured limit.
	if limit <= 0 || limit == int64(^uint64(0)>>1) {
		return nil, errors.New("invalid models response read limit")
	}
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read upstream model list: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("upstream model list response exceeds %d bytes", limit)
	}
	return data, nil
}
