package service

import (
	"bytes"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http/httptest"
)

func TestParseAndBuildGrokImagine20ExtPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"grok-imagine-2.0-ext","prompt":"red apple","n":2,"size":"1:1","resolution":"quality","response_format":"url"}`)
	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	payload, err := buildAPIMartImagesPayload(parsed, parsed.Model, nil, "")
	require.NoError(t, err)
	require.Equal(t, "grok-imagine-2.0-ext", gjson.GetBytes(payload, "model").String())
	require.Equal(t, "quality", gjson.GetBytes(payload, "resolution").String())
	require.Equal(t, "url", gjson.GetBytes(payload, "response_format").String())
}

func TestParseAndBuildGrokImagineImage20Payload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"grok-imagine-image-2.0","prompt":"edit","n":1,"aspect_ratio":"16:9","resolution":"2k","quality":"medium","image_urls":["https://example.com/a.png"]}`)
	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	payload, err := buildAPIMartImagesPayload(parsed, parsed.Model, parsed.InputImageURLs, "")
	require.NoError(t, err)
	require.Equal(t, "16:9", gjson.GetBytes(payload, "aspect_ratio").String())
	require.Equal(t, "2k", gjson.GetBytes(payload, "resolution").String())
	require.False(t, gjson.GetBytes(payload, "quality").Exists())
	require.Len(t, gjson.GetBytes(payload, "image_urls").Array(), 1)
}

func TestGrokImagine20Validation(t *testing.T) {
	require.Error(t, validateAPIMartGrokImagine20Request(&OpenAIImagesRequest{Model: "grok-imagine-2.0-ext", N: 13}))
	require.Error(t, validateAPIMartGrokImagine20Request(&OpenAIImagesRequest{Model: "grok-imagine-image-2.0", N: 11}))
	require.Error(t, validateAPIMartGrokImagine20Request(&OpenAIImagesRequest{Model: "grok-imagine-2.0-ext", N: 1, InputImageURLs: []string{"https://example.com/a.png"}}))
}
