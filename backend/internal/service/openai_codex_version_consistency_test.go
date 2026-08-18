package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

func TestCodexVersionConstants_Consistency(t *testing.T) {
	require.Equal(t, "0.144.1", codexCLIVersion)
	require.Equal(t, openai.CodexDefaultOriginator+"/0.144.1", codexCLIUserAgent)
}
