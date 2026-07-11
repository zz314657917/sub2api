package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountCodexImageGenerationExplicitToolPolicy(t *testing.T) {
	tests := []struct {
		name    string
		account *Account
		want    string
	}{
		{name: "nil account defaults to allow", want: codexImageGenerationExplicitToolPolicyAllow},
		{
			name:    "unset defaults to allow",
			account: &Account{Platform: PlatformOpenAI},
			want:    codexImageGenerationExplicitToolPolicyAllow,
		},
		{
			name: "unknown value defaults to allow",
			account: &Account{Platform: PlatformOpenAI, Extra: map[string]any{
				featureKeyCodexImageGenerationExplicitToolPolicy: "future-value",
			}},
			want: codexImageGenerationExplicitToolPolicyAllow,
		},
		{
			name: "strip alias is normalized",
			account: &Account{Platform: PlatformOpenAI, Extra: map[string]any{
				featureKeyCodexImageGenerationExplicitToolPolicy: " REMOVE ",
			}},
			want: codexImageGenerationExplicitToolPolicyStrip,
		},
		{
			name: "drop alias is normalized",
			account: &Account{Platform: PlatformOpenAI, Extra: map[string]any{
				featureKeyCodexImageGenerationExplicitToolPolicy: "Drop",
			}},
			want: codexImageGenerationExplicitToolPolicyStrip,
		},
		{
			name: "nested openai value is supported",
			account: &Account{Platform: PlatformOpenAI, Extra: map[string]any{
				PlatformOpenAI: map[string]any{featureKeyCodexImageGenerationExplicitToolPolicy: "strip"},
			}},
			want: codexImageGenerationExplicitToolPolicyStrip,
		},
		{
			name: "top level takes precedence",
			account: &Account{Platform: PlatformOpenAI, Extra: map[string]any{
				featureKeyCodexImageGenerationExplicitToolPolicy: "allow",
				PlatformOpenAI: map[string]any{featureKeyCodexImageGenerationExplicitToolPolicy: "strip"},
			}},
			want: codexImageGenerationExplicitToolPolicyAllow,
		},
		{
			name: "non openai account ignores policy",
			account: &Account{Platform: PlatformAnthropic, Extra: map[string]any{
				featureKeyCodexImageGenerationExplicitToolPolicy: "strip",
			}},
			want: codexImageGenerationExplicitToolPolicyAllow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.account.CodexImageGenerationExplicitToolPolicy())
		})
	}
}
