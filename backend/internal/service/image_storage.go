package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultImageMaxDownloadBytes int64 = 32 << 20

// ImageStorage persists image bytes and returns a caller-accessible URL.
type ImageStorage interface {
	Save(ctx context.Context, key, contentType string, data []byte) (url string, err error)
}

// ImageResultUploader replaces upstream image payloads with object-storage
// URLs before the task result reaches Redis.
type ImageResultUploader struct {
	storage          ImageStorage
	httpClient       *http.Client
	prefix           string
	maxDownloadBytes int64
}

func NewImageResultUploader(storage ImageStorage, prefix string, maxDownloadBytes int64, httpClient *http.Client) *ImageResultUploader {
	if httpClient == nil {
		httpClient = defaultImageDownloadHTTPClient()
	}
	if maxDownloadBytes <= 0 {
		maxDownloadBytes = defaultImageMaxDownloadBytes
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "images/"
	} else if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return &ImageResultUploader{
		storage:          storage,
		httpClient:       httpClient,
		prefix:           prefix,
		maxDownloadBytes: maxDownloadBytes,
	}
}

// The default client is deliberately separate from the gateway's generic
// upstream clients. Image URLs are untrusted response data, so downloads are
// HTTPS-only, bypass proxy environment variables, cap redirects, and resolve
// only public addresses immediately before dialing.
func defaultImageDownloadHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = imageDownloadDialContext
	return &http.Client{
		Timeout:       60 * time.Second,
		Transport:     transport,
		CheckRedirect: imageDownloadCheckRedirect,
	}
}

func (u *ImageResultUploader) Rewrite(ctx context.Context, taskID string, result json.RawMessage) (json.RawMessage, error) {
	if u == nil || u.storage == nil {
		return result, nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(result, &top); err != nil {
		return nil, fmt.Errorf("parse image response: %w", err)
	}
	rawData, ok := top["data"]
	if !ok {
		return result, nil
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(rawData, &items); err != nil {
		return nil, fmt.Errorf("parse image response data: %w", err)
	}
	for i, item := range items {
		data, contentType, err := u.fetchImageBytes(ctx, item)
		if err != nil {
			return nil, fmt.Errorf("image %d: %w", i, err)
		}
		key := u.buildKey(taskID, i, contentType)
		storedURL, err := u.storage.Save(ctx, key, contentType, data)
		if err != nil {
			return nil, fmt.Errorf("image %d: upload to object storage: %w", i, err)
		}
		urlRaw, err := json.Marshal(storedURL)
		if err != nil {
			return nil, fmt.Errorf("image %d: encode object URL: %w", i, err)
		}
		item["url"] = urlRaw
		delete(item, "b64_json")
		items[i] = item
	}
	newData, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("encode image response data: %w", err)
	}
	top["data"] = newData
	out, err := json.Marshal(top)
	if err != nil {
		return nil, fmt.Errorf("encode image response: %w", err)
	}
	return out, nil
}

func (u *ImageResultUploader) fetchImageBytes(ctx context.Context, item map[string]json.RawMessage) ([]byte, string, error) {
	if raw, ok := item["b64_json"]; ok {
		var encoded string
		if err := json.Unmarshal(raw, &encoded); err == nil {
			encoded = strings.TrimSpace(encoded)
			if encoded != "" {
				if int64(base64.StdEncoding.DecodedLen(len(encoded))) > u.maxBytes() {
					return nil, "", fmt.Errorf("decoded image exceeds %d bytes", u.maxBytes())
				}
				data, err := base64.StdEncoding.DecodeString(encoded)
				if err != nil {
					return nil, "", fmt.Errorf("decode b64_json: %w", err)
				}
				contentType, err := imageContentType(data, "")
				if err != nil {
					return nil, "", err
				}
				return data, contentType, nil
			}
		}
	}
	if raw, ok := item["url"]; ok {
		var rawURL string
		if err := json.Unmarshal(raw, &rawURL); err == nil && strings.TrimSpace(rawURL) != "" {
			return u.download(ctx, rawURL)
		}
	}
	return nil, "", errors.New("image item has neither b64_json nor url")
}

func (u *ImageResultUploader) download(ctx context.Context, rawURL string) ([]byte, string, error) {
	if err := validateImageDownloadURL(rawURL); err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build image download request: %w", err)
	}
	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("download image: unexpected status %d", resp.StatusCode)
	}
	limit := u.maxBytes()
	if resp.ContentLength > limit {
		return nil, "", fmt.Errorf("downloaded image exceeds %d bytes", limit)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", fmt.Errorf("read image body: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, "", fmt.Errorf("downloaded image exceeds %d bytes", limit)
	}
	contentType, err := imageContentType(data, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, "", err
	}
	return data, contentType, nil
}

func (u *ImageResultUploader) maxBytes() int64 {
	if u == nil || u.maxDownloadBytes <= 0 {
		return defaultImageMaxDownloadBytes
	}
	return u.maxDownloadBytes
}

func (u *ImageResultUploader) buildKey(taskID string, index int, contentType string) string {
	return u.prefix + taskID + "-" + strconv.Itoa(index) + extensionForImageContentType(contentType)
}

func imageContentType(data []byte, header string) (string, error) {
	contentType := strings.TrimSpace(strings.Split(header, ";")[0])
	if strings.HasPrefix(contentType, "image/") {
		return contentType, nil
	}
	contentType = strings.TrimSpace(strings.Split(http.DetectContentType(data), ";")[0])
	if !strings.HasPrefix(contentType, "image/") {
		return "", errors.New("downloaded content is not an image")
	}
	return contentType, nil
}

func extensionForImageContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	default:
		return ".img"
	}
}

func imageDownloadCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 {
		return errors.New("image download exceeded redirect limit")
	}
	return validateImageDownloadURL(req.URL.String())
}

func validateImageDownloadURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil {
		return errors.New("invalid image download URL")
	}
	if !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || strings.TrimSpace(parsed.Hostname()) == "" {
		return errors.New("image download URL must use HTTPS with a host and no userinfo")
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > 65535 {
			return errors.New("image download URL has an invalid port")
		}
	}
	host := strings.TrimSpace(parsed.Hostname())
	if strings.EqualFold(host, "localhost") {
		return errors.New("image download URL must not target localhost")
	}
	if address, err := netip.ParseAddr(host); err == nil && !isPublicImageDownloadIP(address) {
		return errors.New("image download URL must not target a private address")
	}
	return nil
}

func imageDownloadDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("split image download address: %w", err)
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve image download host: %w", err)
	}
	if len(ips) == 0 {
		return nil, errors.New("image download host did not resolve")
	}
	for _, ip := range ips {
		if !isPublicImageDownloadIP(ip) {
			return nil, errors.New("image download host resolved to a private address")
		}
	}
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	var lastErr error
	for _, ip := range ips {
		conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("dial image download host: %w", lastErr)
}

func isPublicImageDownloadIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsValid() && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsUnspecified()
}
