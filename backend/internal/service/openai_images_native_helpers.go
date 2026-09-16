package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

// detectOpenAIImageResultSize reports dimensions only for formats the local
// decoder understands.  A malformed upstream image must not turn a successful
// response into a synthetic size claim.
func detectOpenAIImageResultSize(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(decoded) == 0 {
		return ""
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(decoded))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return ""
	}
	return fmt.Sprintf("%dx%d", config.Width, config.Height)
}
