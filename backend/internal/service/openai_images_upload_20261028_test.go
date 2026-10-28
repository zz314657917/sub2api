package service

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func upstreamAMediaMultipart(t *testing.T, field string, size int) ([]byte, string, []byte) {
	t.Helper()
	image := bytes.Repeat([]byte{0xab}, size)
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	require.NoError(t, w.WriteField("model", "gpt-image-1"))
	require.NoError(t, w.WriteField("prompt", "keep every byte"))
	require.NoError(t, w.WriteField("transparent_background", "true"))
	if field == "mask" {
		p, err := w.CreateFormFile("image", "image.png")
		require.NoError(t, err)
		_, err = p.Write([]byte("image-bytes"))
		require.NoError(t, err)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="source.png"`, field))
	if field == "prompt" || field == "" {
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, field))
	}
	h.Set("Content-Type", "image/png")
	p, err := w.CreatePart(h)
	require.NoError(t, err)
	_, err = p.Write(image)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return body.Bytes(), w.FormDataContentType(), image
}

func upstreamAMediaParse(t *testing.T, body []byte, contentType string) (*OpenAIImagesRequest, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", contentType)
	return (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
}

func TestUpstreamAFixesMediaMultipartBoundary(t *testing.T) {
	for _, size := range []int{21 << 20, 50 << 20} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			body, ct, image := upstreamAMediaMultipart(t, "image", size)
			req, err := upstreamAMediaParse(t, body, ct)
			require.NoError(t, err)
			require.Len(t, req.Uploads, 1)
			require.Equal(t, size, len(req.Uploads[0].Data))
			require.True(t, bytes.Equal(image, req.Uploads[0].Data))
			require.Equal(t, "image/png", req.Uploads[0].ContentType)
			require.Equal(t, "transparent", req.Background)
			rewritten, rewrittenCT, err := rewriteOpenAIImagesModel(body, ct, "gpt-image-2")
			require.NoError(t, err)
			got, err := upstreamAMediaParse(t, rewritten, rewrittenCT)
			require.NoError(t, err)
			require.Equal(t, "gpt-image-2", got.Model)
			require.Equal(t, "transparent", got.Background)
			require.Len(t, got.Uploads, 1)
			require.Equal(t, "source.png", got.Uploads[0].FileName)
			require.Equal(t, "image/png", got.Uploads[0].ContentType)
			require.True(t, bytes.Equal(image, got.Uploads[0].Data))
		})
	}
	require.Equal(t, 20<<20, openAIImageMaxDownloadBytes)
}

func TestUpstreamAFixesMediaMultipartOversize(t *testing.T) {
	for _, field := range []string{"image", "image[0]", "mask", "ignored_file", "prompt", ""} {
		t.Run(field, func(t *testing.T) {
			body, ct, _ := upstreamAMediaMultipart(t, field, (50<<20)+1)
			req, err := upstreamAMediaParse(t, body, ct)
			require.Nil(t, req)
			require.ErrorContains(t, err, fmt.Sprintf(`multipart field %q exceeds the 50 MB per-part limit`, field))
		})
	}
}

func TestUpstreamAFixesMediaMultipartMaskRewrite(t *testing.T) {
	for _, size := range []int{21 << 20, 50 << 20} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			body, ct, mask := upstreamAMediaMultipart(t, "mask", size)
			rewritten, rewrittenCT, err := rewriteOpenAIImagesModelAndN(body, ct, "gpt-image-2", 2, true)
			require.NoError(t, err)
			req, err := upstreamAMediaParse(t, rewritten, rewrittenCT)
			require.NoError(t, err)
			require.True(t, req.HasMask)
			require.NotNil(t, req.MaskUpload)
			require.True(t, bytes.Equal(mask, req.MaskUpload.Data))
			require.Equal(t, "image/png", req.MaskUpload.ContentType)
			require.Equal(t, 2, req.N)
			require.Equal(t, "transparent", req.Background)
		})
	}
}
