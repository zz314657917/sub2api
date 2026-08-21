package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func seedreamTestImageURLs(count int) []string {
	urls := make([]string, count)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://example.com/ref-%d.png", i)
	}
	return urls
}

func TestBuildAPIMartSeedream50ProPayload(t *testing.T) {
	watermark := true
	payload, err := buildAPIMartImagesPayload(&OpenAIImagesRequest{
		Prompt:       "product poster",
		N:            1,
		Size:         "16:9",
		Resolution:   "2K",
		OutputFormat: "png",
		Watermark:    &watermark,
	}, "seedream-5-0-pro", []string{"https://example.com/ref.png"}, "")
	require.NoError(t, err)
	require.Equal(t, "seedream-5-0-pro", gjson.GetBytes(payload, "model").String())
	require.Equal(t, "2K", gjson.GetBytes(payload, "resolution").String())
	require.Equal(t, "png", gjson.GetBytes(payload, "output_format").String())
	require.True(t, gjson.GetBytes(payload, "watermark").Bool())
	require.Len(t, gjson.GetBytes(payload, "image_urls").Array(), 1)
}

func TestBuildAPIMartSeedream50LitePayloadEnablesSequence(t *testing.T) {
	payload, err := buildAPIMartImagesPayload(&OpenAIImagesRequest{
		Prompt:       "storyboard",
		N:            3,
		Size:         "16:9",
		Resolution:   "3K",
		OutputFormat: "jpeg",
	}, "seedream-5-0-lite", nil, "")
	require.NoError(t, err)
	require.Equal(t, "auto", gjson.GetBytes(payload, "sequential_image_generation").String())
	require.Equal(t, "jpeg", gjson.GetBytes(payload, "output_format").String())
}

func TestValidateAPIMartSeedream50Limits(t *testing.T) {
	require.Error(t, validateAPIMartSeedream50Request(&OpenAIImagesRequest{Model: "seedream-5-0-pro", N: 2}))
	require.Error(t, validateAPIMartSeedream50Request(&OpenAIImagesRequest{
		Model:          "seedream-5-0-pro",
		N:              1,
		InputImageURLs: seedreamTestImageURLs(11),
	}))
	require.Error(t, validateAPIMartSeedream50Request(&OpenAIImagesRequest{
		Model:          "seedream-5-0-lite",
		N:              6,
		InputImageURLs: seedreamTestImageURLs(10),
	}))
	require.NoError(t, validateAPIMartSeedream50Request(&OpenAIImagesRequest{
		Model:          "seedream-5-0-lite",
		N:              5,
		InputImageURLs: seedreamTestImageURLs(10),
	}))
	require.True(t, isAPIMartImagesAsyncModel("seedream-5-0-pro"))
	require.True(t, isAPIMartImagesAsyncModel("seedream-5-0-lite"))
}
