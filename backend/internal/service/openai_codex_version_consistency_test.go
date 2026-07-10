package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexVersionConstants_Consistency(t *testing.T) {
	require.Equal(t, "0.144.1", codexCLIVersion)
	require.Equal(t, "0.144.1", openAICodexProbeVersion)
	require.Equal(t, "codex_cli_rs/0.144.1", codexCLIUserAgent)
}
