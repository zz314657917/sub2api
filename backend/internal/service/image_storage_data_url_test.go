package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

var imageDataURLPNG = []byte("\x89PNG\r\n\x1a\nfake-png-payload")

type imageDataURLStorage struct {
	key         string
	contentType string
	data        []byte
}

func (s *imageDataURLStorage) Save(_ context.Context, key, contentType string, data []byte) (string, error) {
	s.key = key
	s.contentType = contentType
	s.data = append([]byte(nil), data...)
	return "https://cdn.test/" + key, nil
}

type imageDataURLRoundTripper func(*http.Request) (*http.Response, error)

func (f imageDataURLRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newDataURLHTTPClient(t *testing.T, calls *int) *http.Client {
	t.Helper()
	return &http.Client{Transport: imageDataURLRoundTripper(func(*http.Request) (*http.Response, error) {
		(*calls)++
		return nil, errors.New("HTTP must not be called for data URLs")
	})}
}

func TestImageResultUploaderRewritesImageDataURLWithoutHTTP(t *testing.T) {
	httpCalls := 0
	storage := &imageDataURLStorage{}
	uploader := NewImageResultUploader(storage, "images/", 0, newDataURLHTTPClient(t, &httpCalls))
	payload := base64.StdEncoding.EncodeToString(imageDataURLPNG)
	result := json.RawMessage(`{"data":[{"url":"DATA:image/jpeg;name=photo.jpg;BaSe64,` + payload + `","revised_prompt":"kept"}]}`)

	out, err := uploader.Rewrite(context.Background(), "imgtask_data", result)
	require.NoError(t, err)
	require.Zero(t, httpCalls)
	require.Equal(t, imageDataURLPNG, storage.data)
	require.Equal(t, "image/png", storage.contentType)
	require.Equal(t, "images/imgtask_data-0.png", storage.key)

	var parsed struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(out, &parsed))
	require.JSONEq(t, `"https://cdn.test/images/imgtask_data-0.png"`, string(parsed.Data[0]["url"]))
	require.JSONEq(t, `"kept"`, string(parsed.Data[0]["revised_prompt"]))
}

func TestImageResultUploaderRejectsInvalidDataURLWithoutHTTP(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{name: "missing comma", url: "data:image/png;base64", wantErr: "missing comma"},
		{name: "non image", url: "data:text/plain;base64,aGVsbG8=", wantErr: "is not an image"},
		{name: "non base64", url: "data:image/png,raw", wantErr: "not base64"},
		{name: "invalid base64", url: "data:image/png;base64,%%%", wantErr: "base64 payload"},
		{name: "invalid media type", url: "data:image/png;bad parameter;base64,aGVsbG8=", wantErr: "invalid media type"},
		{name: "parameter after base64", url: "data:image/png;base64;name=photo.png,aGVsbG8=", wantErr: "base64 marker must be the final header token"},
		{name: "duplicate base64 marker", url: "data:image/png;base64;base64,aGVsbG8=", wantErr: "duplicate base64 marker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCalls := 0
			uploader := NewImageResultUploader(&imageDataURLStorage{}, "images/", 0, newDataURLHTTPClient(t, &httpCalls))
			result, err := json.Marshal(map[string]any{"data": []map[string]string{{"url": tt.url}}})
			require.NoError(t, err)

			_, err = uploader.Rewrite(context.Background(), "imgtask_bad", result)
			require.ErrorContains(t, err, tt.wantErr)
			require.Zero(t, httpCalls)
		})
	}
}

func TestImageResultUploaderRejectsOversizedImageDataURL(t *testing.T) {
	storage := &imageDataURLStorage{}
	uploader := NewImageResultUploader(storage, "images/", 3, nil)
	payload := base64.StdEncoding.EncodeToString([]byte("four"))
	result := json.RawMessage(`{"data":[{"url":"data:image/png;base64,` + payload + `"}]}`)

	_, err := uploader.Rewrite(context.Background(), "imgtask_large", result)
	require.ErrorContains(t, err, "decoded image data URL exceeds 3 bytes")
	require.Empty(t, storage.data)
}

func TestImageResultUploaderB64JSONTakesPrecedenceOverDataURL(t *testing.T) {
	storage := &imageDataURLStorage{}
	uploader := NewImageResultUploader(storage, "images/", 0, nil)
	b64 := base64.StdEncoding.EncodeToString(imageDataURLPNG)
	result := json.RawMessage(`{"data":[{"b64_json":"` + b64 + `","url":"data:text/plain,not-an-image"}]}`)

	_, err := uploader.Rewrite(context.Background(), "imgtask_precedence", result)
	require.NoError(t, err)
	require.Equal(t, imageDataURLPNG, storage.data)
}
