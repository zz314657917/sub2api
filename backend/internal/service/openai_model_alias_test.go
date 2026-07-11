package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsOpenAIGPT56ModelAliases(t *testing.T) {
	for _, model := range []string{
		"gpt-5.6-sol",
		"openai/gpt-5.6-terra",
		"gpt-5.6-luna-2026-07-09",
		"GPT5.6 LUNA",
	} {
		require.True(t, isOpenAIGPT56Model(model), model)
	}
	require.False(t, isOpenAIGPT56Model("gpt-5.4"))
}
