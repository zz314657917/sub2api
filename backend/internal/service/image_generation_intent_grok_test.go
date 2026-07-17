package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsImageGenerationIntentForPlatform_GrokPassiveAndExplicitSignals(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		model    string
		body     string
		want     bool
	}{
		{
			name:     "chat completions namespace with automatic choice is passive",
			endpoint: "/v1/chat/completions",
			model:    "grok-4.5",
			body:     `{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":"auto","input":"write code"}`,
		},
		{
			name:  "top-level namespace without choice is passive",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"input":"write code"}`,
		},
		{
			name:  "Responses Lite additional tools with automatic choice is passive",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tool_choice":"auto","input":[{"type":"additional_tools","tools":[{"type":"namespace","name":"image_gen"}]},{"type":"message","role":"user","content":"write code"}]}`,
		},
		{
			name:  "Responses Lite additional tools without choice is passive",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","input":[{"type":"additional_tools","tools":[{"type":"namespace","name":"image_gen"}]}]}`,
		},
		{
			name:  "flattened namespace declaration is passive",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"function","name":"image_gen.imagegen"}],"tool_choice":"auto"}`,
		},
		{
			name:  "native image_generation tool is explicit",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"image_generation"}],"input":"draw"}`,
			want:  true,
		},
		{
			name:  "Responses Lite native image_generation tool is explicit",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","input":[{"type":"additional_tools","tools":[{"type":"image_generation"}]}]}`,
			want:  true,
		},
		{
			name:  "image model is explicit",
			model: "grok-4.5",
			body:  `{"model":"gpt-image-2","input":"draw"}`,
			want:  true,
		},
		{
			name:  "namespace tool choice is explicit",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":{"type":"namespace","name":"image_gen"}}`,
			want:  true,
		},
		{
			name:  "flattened namespace function choice is explicit",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"function","name":"image_gen.imagegen"}],"tool_choice":{"type":"function","name":"image_gen.imagegen"}}`,
			want:  true,
		},
		{
			name:  "wrapped namespace function choice is explicit",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tool_choice":{"tool":{"type":"function","name":"image_gen.imagegen"}}}`,
			want:  true,
		},
		{
			name:  "ordinary function and vision input are not image generation",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","tools":[{"type":"function","name":"lookup"}],"input":[{"type":"input_image","image_url":"data:image/png;base64,AA=="}]}`,
		},
		{
			name:  "function call history is not current intent",
			model: "grok-4.5",
			body:  `{"model":"grok-4.5","input":[{"type":"function_call","namespace":"image_gen","name":"imagegen"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint := tt.endpoint
			if endpoint == "" {
				endpoint = openAIResponsesEndpoint
			}
			got := IsImageGenerationIntentForPlatform(
				endpoint,
				tt.model,
				[]byte(tt.body),
				PlatformGrok,
			)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestIsImageGenerationIntentForPlatform_PreservesOpenAIDeclarationSemantics(t *testing.T) {
	body := []byte(`{"model":"gpt-5.5","tools":[{"type":"namespace","name":"image_gen"}],"tool_choice":"auto","input":"write code"}`)

	require.True(t, IsImageGenerationIntentForPlatform(openAIResponsesEndpoint, "gpt-5.5", body, PlatformOpenAI))
	require.True(t, IsImageGenerationIntentForPlatform(openAIResponsesEndpoint, "gpt-5.5", body, ""))
}

func TestIsImageGenerationIntentMapForPlatform_GrokPassiveAndExplicitSignals(t *testing.T) {
	passive := map[string]any{
		"model": "grok-4.5",
		"tools": []any{
			map[string]any{"type": "namespace", "name": "image_gen"},
		},
		"input": []any{
			map[string]any{
				"type":  "additional_tools",
				"tools": []any{map[string]any{"type": "namespace", "name": "image_gen"}},
			},
		},
	}
	require.False(t, IsImageGenerationIntentMapForPlatform(openAIResponsesEndpoint, "grok-4.5", passive, PlatformGrok))

	explicit := map[string]any{
		"model": "grok-4.5",
		"tools": []any{map[string]any{"type": "namespace", "name": "image_gen"}},
		"tool_choice": map[string]any{
			"type": "function",
			"name": "image_gen.imagegen",
		},
	}
	require.True(t, IsImageGenerationIntentMapForPlatform(openAIResponsesEndpoint, "grok-4.5", explicit, PlatformGrok))
}

func TestIsImageGenerationIntentMapForPlatform_OtherPlatformsUseLegacySemantics(t *testing.T) {
	body := map[string]any{
		"model": "gpt-5.5",
		"tools": []any{map[string]any{"type": "namespace", "name": "image_gen"}},
	}
	require.True(t, IsImageGenerationIntentMapForPlatform(openAIResponsesEndpoint, "gpt-5.5", body, PlatformOpenAI))
}
