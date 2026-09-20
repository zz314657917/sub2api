package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireObjectInputSchema(t *testing.T, schema json.RawMessage) map[string]json.RawMessage {
	t.Helper()

	require.NotEmpty(t, schema)

	var parsed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(schema, &parsed))
	require.JSONEq(t, `"object"`, string(parsed["type"]))
	require.Contains(t, parsed, "properties")

	var properties map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(parsed["properties"], &properties))

	return parsed
}

func TestResponsesToAnthropic_CustomToolNormalizesToObjectSchema(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-5.2",
		Input: json.RawMessage(`[{"role":"user","content":"Hello"}]`),
		Tools: []ResponsesTool{{
			Type:        "custom",
			Name:        "custom_shell",
			Description: "Run shell command",
			Parameters:  json.RawMessage(`{"type":"string"}`),
		}},
	}

	resp, err := ResponsesToAnthropicRequest(req)
	require.NoError(t, err)
	require.Len(t, resp.Tools, 1)
	require.Empty(t, resp.Tools[0].Type)
	require.Equal(t, "custom_shell", resp.Tools[0].Name)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(resp.Tools[0].InputSchema, &schema))
	require.Equal(t, "object", schema["type"])
	require.Contains(t, schema, "properties")
}

func TestResponsesToAnthropic_UnknownToolNormalizesInvalidSchema(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-5.2",
		Input: json.RawMessage(`[{"role":"user","content":"Hello"}]`),
		Tools: []ResponsesTool{{
			Type:       "local_shell",
			Name:       "run",
			Parameters: json.RawMessage(`not-json`),
		}},
	}

	resp, err := ResponsesToAnthropicRequest(req)
	require.NoError(t, err)
	require.Len(t, resp.Tools, 1)
	require.Equal(t, "local_shell", resp.Tools[0].Type)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(resp.Tools[0].InputSchema, &schema))
	require.Equal(t, "object", schema["type"])
	require.Contains(t, schema, "properties")
}

// Codex 的 codex_app 命名空间工具（如 automation_update）把 parameters 根节点声明为
// 对象分支的 oneOf/anyOf；Anthropic 拒绝 input_schema 顶层的 oneOf/anyOf/allOf，
// 转换时必须摊平成单个 object schema。
func TestResponsesToAnthropic_ObjectUnionRootFlattened(t *testing.T) {
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type: "function",
		Name: "codex_app__automation_update",
		Parameters: json.RawMessage(`{
			"oneOf": [
				{"type":"object","properties":{"id":{"type":"string"}}},
				{"anyOf":[{"type":"object"},{"type":"object","properties":{}}]}
			]
		}`),
	}})

	require.Len(t, tools, 1)
	schema := requireObjectInputSchema(t, tools[0].InputSchema)
	assert.NotContains(t, schema, "oneOf")
	assert.NotContains(t, schema, "anyOf")
	assert.NotContains(t, schema, "allOf")
	assert.JSONEq(t, `{"id":{"type":"string"}}`, string(schema["properties"]))
	assert.NotContains(t, schema, "required")
}

func TestResponsesToAnthropic_ObjectUnionRootMergesBranches(t *testing.T) {
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type: "function",
		Name: "codex_app__automation_update",
		Parameters: json.RawMessage(`{
			"anyOf": [
				{"type":"object","properties":{"mode":{"enum":["view"]},"id":{"type":"string"}},"required":["mode","id"]},
				{"type":"object","properties":{"mode":{"enum":["update"]},"prompt":{"type":"string"}},"required":["mode","prompt"]}
			]
		}`),
	}})

	require.Len(t, tools, 1)
	schema := requireObjectInputSchema(t, tools[0].InputSchema)
	assert.NotContains(t, schema, "anyOf")

	// 属性取各分支并集；同名属性保留两侧约束（嵌套联合本身是允许的）。
	var properties map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(schema["properties"], &properties))
	require.Contains(t, properties, "id")
	require.Contains(t, properties, "prompt")
	var mode map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(properties["mode"], &mode))
	assert.Contains(t, mode, "anyOf")

	// required 取各分支交集，只保留所有分支都要求的字段。
	assert.JSONEq(t, `["mode"]`, string(schema["required"]))
}

func TestResponsesToAnthropic_AllOfRootKeepsUnionRequired(t *testing.T) {
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type: "function",
		Name: "strict_tool",
		Parameters: json.RawMessage(`{
			"allOf": [
				{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]},
				{"type":"object","properties":{"encoding":{"type":"string"}},"required":["encoding"]}
			]
		}`),
	}})

	require.Len(t, tools, 1)
	schema := requireObjectInputSchema(t, tools[0].InputSchema)
	assert.NotContains(t, schema, "allOf")
	assert.JSONEq(t, `["path","encoding"]`, string(schema["required"]))
}

// 非对象分支无法用 Anthropic 的 object schema 表达，直接丢弃分支但保留属性词汇。
func TestResponsesToAnthropic_RootUnionDropsNonObjectBranches(t *testing.T) {
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type:       "function",
		Name:       "mixed_tool",
		Parameters: json.RawMessage(`{"oneOf":[{"type":"object","properties":{"path":{"type":"string"}}},{"type":"string"}]}`),
	}})

	require.Len(t, tools, 1)
	schema := requireObjectInputSchema(t, tools[0].InputSchema)
	assert.NotContains(t, schema, "oneOf")
	var properties map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(schema["properties"], &properties))
	assert.Contains(t, properties, "path")
}

func TestResponsesToAnthropic_NestedUnionPreserved(t *testing.T) {
	parameters := json.RawMessage(`{"type":"object","properties":{"value":{"anyOf":[{"type":"string"},{"type":"array","items":{"type":"string"}}]}}}`)
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type:       "function",
		Name:       "get_value",
		Parameters: parameters,
	}})

	require.Len(t, tools, 1)
	assert.JSONEq(t, string(parameters), string(tools[0].InputSchema))
}

// 根节点自带 properties/required 时，与分支约束合并而不是被覆盖。
func TestResponsesToAnthropic_RootUnionKeepsRootPropertiesAndRequired(t *testing.T) {
	tools := convertResponsesToAnthropicTools([]ResponsesTool{{
		Type: "function",
		Name: "merge_root_tool",
		Parameters: json.RawMessage(`{
			"type":"object",
			"properties":{"shared":{"type":"string"}},
			"required":["shared"],
			"oneOf":[{"type":"object","properties":{"extra":{"type":"integer"}}}]
		}`),
	}})

	require.Len(t, tools, 1)
	schema := requireObjectInputSchema(t, tools[0].InputSchema)
	assert.NotContains(t, schema, "oneOf")
	var properties map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(schema["properties"], &properties))
	assert.Contains(t, properties, "shared")
	assert.Contains(t, properties, "extra")
	assert.JSONEq(t, `["shared"]`, string(schema["required"]))
}

// 端到端：已由网关展开的 Codex 命名空间工具转为 Anthropic 工具后，
// schema 顶部不再有联合关键字。
func TestResponsesToAnthropic_CodexNamespaceRootUnionTool(t *testing.T) {
	body := []byte(`{
		"model": "claude-opus-5",
		"input": "create an automation",
		"tools": [{
			"type": "function",
			"name": "codex_app__automation_update",
			"description": "Create, update, view, or delete recurring automations",
			"parameters": {
				"oneOf": [
					{"type":"object","properties":{"mode":{"enum":["view"]},"id":{"type":"string"}},"required":["mode","id"]},
					{"type":"object","properties":{"mode":{"enum":["update"]},"id":{"type":"string"}},"required":["mode","id"]}
				]
			}
		}]
	}`)

	var responsesReq ResponsesRequest
	require.NoError(t, json.Unmarshal(body, &responsesReq))

	anthropicReq, err := ResponsesToAnthropicRequest(&responsesReq)
	require.NoError(t, err)
	require.Len(t, anthropicReq.Tools, 1)

	tool := anthropicReq.Tools[0]
	assert.Equal(t, "codex_app__automation_update", tool.Name)
	schema := requireObjectInputSchema(t, tool.InputSchema)
	assert.NotContains(t, schema, "oneOf")
	assert.NotContains(t, schema, "anyOf")
	assert.NotContains(t, schema, "allOf")
	assert.JSONEq(t, `["mode","id"]`, string(schema["required"]))

	wire, err := json.Marshal(tool)
	require.NoError(t, err)
	assert.NotContains(t, string(wire), `"oneOf"`)
	assert.NotContains(t, string(wire), `"allOf"`)
	// 同名属性两侧的约束用嵌套联合保留，只允许出现在 properties 内部。
	assert.Contains(t, string(wire), `"mode":{"anyOf":[{"enum":["view"]},{"enum":["update"]}]}`)
}
