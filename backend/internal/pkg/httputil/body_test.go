package httputil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestSevenDecompressionBoundaries(t *testing.T) {
	for _, enc := range []string{"gzip", "x-gzip", "deflate", "zstd"} {
		t.Run(enc, func(t *testing.T) {
			for _, size := range []int{64 << 20, (64 << 20) + 1} {
				payload := bytes.Repeat([]byte("x"), size)
				var buf bytes.Buffer
				switch enc {
				case "gzip", "x-gzip":
					w := gzip.NewWriter(&buf)
					_, err := w.Write(payload)
					if err != nil {
						t.Fatal(err)
					}
					if err = w.Close(); err != nil {
						t.Fatal(err)
					}
				case "deflate":
					w := zlib.NewWriter(&buf)
					_, err := w.Write(payload)
					if err != nil {
						t.Fatal(err)
					}
					if err = w.Close(); err != nil {
						t.Fatal(err)
					}
				case "zstd":
					w, err := zstd.NewWriter(&buf, zstd.WithWindowSize(1<<20))
					if err != nil {
						t.Fatal(err)
					}
					if _, err = w.Write(payload); err != nil {
						t.Fatal(err)
					}
					if err = w.Close(); err != nil {
						t.Fatal(err)
					}
				}
				req := newRequestWithBody(t, buf.Bytes(), enc)
				req.Header.Set("Content-Length", "original")
				originalLength := req.ContentLength
				got, err := ReadRequestBodyWithPrealloc(req)
				if size > 64<<20 {
					var maxErr *http.MaxBytesError
					if !errors.As(err, &maxErr) || maxErr.Limit != 64<<20 {
						t.Fatalf("expected MaxBytesError, got len=%d err=%v", len(got), err)
					}
					if req.Header.Get("Content-Encoding") != enc || req.Header.Get("Content-Length") != "original" || req.ContentLength != originalLength {
						t.Fatal("failed decode changed headers")
					}
				} else {
					if err != nil || !bytes.Equal(got, payload) {
						t.Fatalf("exact boundary failed: len=%d err=%v", len(got), err)
					}
					if req.Header.Get("Content-Encoding") != "" || req.Header.Get("Content-Length") != "" || req.ContentLength != int64(size) {
						t.Fatal("successful decode headers incorrect")
					}
				}
			}
			req := newRequestWithBody(t, []byte("corrupt"), enc)
			if _, err := ReadRequestBodyWithPrealloc(req); err == nil {
				t.Fatal("corrupt stream accepted")
			}
			if req.Header.Get("Content-Encoding") != enc {
				t.Fatal("corrupt stream changed encoding")
			}
		})
	}
}

func TestSevenZstdWindowLimit(t *testing.T) {
	// A streaming frame advertises a 16 MiB window, above the allowed 8 MiB.
	var buf bytes.Buffer
	w, err := zstd.NewWriter(&buf, zstd.WithWindowSize(16<<20))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(bytes.Repeat([]byte("window"), 2<<20)); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	req := newRequestWithBody(t, buf.Bytes(), "zstd")
	if _, err = ReadRequestBodyWithPrealloc(req); err == nil {
		t.Fatal("oversized zstd window accepted")
	}
	if req.Header.Get("Content-Encoding") != "zstd" {
		t.Fatal("failed decode changed header")
	}
}

const samplePayload = `{"model":"gpt-5.5","input":"hi","stream":false}`

func newRequestWithBody(t *testing.T, body []byte, encoding string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if encoding != "" {
		req.Header.Set("Content-Encoding", encoding)
	}
	req.ContentLength = int64(len(body))
	return req
}

func TestReadRequestBodyWithPrealloc_PassesThroughIdentity(t *testing.T) {
	req := newRequestWithBody(t, []byte(samplePayload), "")
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != samplePayload {
		t.Fatalf("body mismatch: got %q", got)
	}
}

func TestReadRequestBodyWithPrealloc_DecodesZstd(t *testing.T) {
	enc, _ := zstd.NewWriter(nil)
	compressed := enc.EncodeAll([]byte(samplePayload), nil)
	_ = enc.Close()

	req := newRequestWithBody(t, compressed, "zstd")
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != samplePayload {
		t.Fatalf("body mismatch: got %q", got)
	}
	if req.Header.Get("Content-Encoding") != "" {
		t.Fatalf("Content-Encoding should be cleared after decoding")
	}
	if req.ContentLength != int64(len(samplePayload)) {
		t.Fatalf("ContentLength not updated: %d", req.ContentLength)
	}
}

func TestReadRequestBodyWithPrealloc_DecodesGzip(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(samplePayload)); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}

	req := newRequestWithBody(t, buf.Bytes(), "gzip")
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != samplePayload {
		t.Fatalf("body mismatch: got %q", got)
	}
}

func TestReadRequestBodyWithPrealloc_DecodesDeflate(t *testing.T) {
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write([]byte(samplePayload)); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}

	req := newRequestWithBody(t, buf.Bytes(), "deflate")
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != samplePayload {
		t.Fatalf("body mismatch: got %q", got)
	}
}

func TestReadRequestBodyWithPrealloc_RejectsUnsupportedEncoding(t *testing.T) {
	req := newRequestWithBody(t, []byte(samplePayload), "br")
	_, err := ReadRequestBodyWithPrealloc(req)
	if err == nil {
		t.Fatal("expected error for unsupported encoding, got nil")
	}
	if !strings.Contains(err.Error(), "br") {
		t.Fatalf("error should mention encoding, got %v", err)
	}
}

func TestReadRequestBodyWithPrealloc_RejectsCorruptZstd(t *testing.T) {
	req := newRequestWithBody(t, []byte("not actually zstd"), "zstd")
	_, err := ReadRequestBodyWithPrealloc(req)
	if err == nil {
		t.Fatal("expected error for corrupt zstd body, got nil")
	}
}

func TestReadRequestBodyWithPrealloc_NilBody(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/v1/responses", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil body, got %q", got)
	}
}

func TestReadRequestBodyWithPrealloc_RespectsIdentityEncoding(t *testing.T) {
	req := newRequestWithBody(t, []byte(samplePayload), "identity")
	got, err := ReadRequestBodyWithPrealloc(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != samplePayload {
		t.Fatalf("body mismatch: got %q", got)
	}
}
