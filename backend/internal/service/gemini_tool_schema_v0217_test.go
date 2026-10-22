package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanToolSchemaConstExactIntersectionV0217(t *testing.T) {
	for _, tc := range []struct {
		name   string
		schema map[string]any
		want   map[string]any
	}{
		{"string", map[string]any{"const": "a"}, map[string]any{"type": "STRING", "enum": []any{"a"}}},
		{"empty", map[string]any{"const": ""}, map[string]any{"type": "STRING", "enum": []any{""}}},
		{"intersection", map[string]any{"const": "a", "enum": []any{"b", "a", "a"}}, map[string]any{"type": "STRING", "enum": []any{"a"}}},
		{"disjoint", map[string]any{"const": "a", "enum": []any{"b"}}, map[string]any{"type": "STRING", "enum": []any{}}},
		{"empty enum", map[string]any{"const": "a", "enum": []any{}}, map[string]any{"type": "STRING", "enum": []any{}}},
		{"numeric mismatch", map[string]any{"const": "7", "enum": []any{float64(7)}}, map[string]any{"type": "STRING", "enum": []any{}}},
		{"exact mixed", map[string]any{"const": "7", "enum": []any{float64(7), "7"}}, map[string]any{"type": "STRING", "enum": []any{"7"}}},
		{"existing type", map[string]any{"type": "integer", "const": "7"}, map[string]any{"type": "INTEGER", "enum": []any{"7"}}},
		{"null type present", map[string]any{"type": nil, "const": "a"}, map[string]any{"type": nil, "enum": []any{"a"}}},
		{"invalid enum member", map[string]any{"const": "a", "enum": []any{"a", map[string]any{"const": "a"}}}, map[string]any{"type": "STRING"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, cleanToolSchema(tc.schema))
		})
	}
}

func TestCleanToolSchemaNonStringConstAndScalarPolicyV0217(t *testing.T) {
	for _, constant := range []any{nil, false, true, float64(7), []any{"a"}, map[string]any{"type": "string"}} {
		t.Run(fmt.Sprintf("%#v", constant), func(t *testing.T) {
			require.Equal(t, map[string]any{}, cleanToolSchema(map[string]any{"const": constant}))
			require.Equal(t, map[string]any{"type": "INTEGER", "enum": []any{"8"}},
				cleanToolSchema(map[string]any{"type": "integer", "const": constant, "enum": []any{float64(8)}}))
		})
	}
	require.Equal(t, map[string]any{"enum": []any{"enabled", "false", "7", "null"}},
		cleanToolSchema(map[string]any{"enum": []any{"enabled", false, json.Number("7"), nil}}))
	for _, invalid := range []any{map[string]any{"type": "string"}, []any{"a"}} {
		require.Equal(t, map[string]any{}, cleanToolSchema(map[string]any{"enum": []any{"a", invalid}}))
	}
}

func TestCleanToolSchemaLiteralAndPropertyNamesV0217(t *testing.T) {
	literal := map[string]any{
		"type": "object", "const": false, "deprecated": true,
		"examples": []any{map[string]any{"type": "integer", "const": "7"}},
		"enum":     []any{map[string]any{"const": "literal"}, float64(7), nil},
		"nested":   []any{map[string]any{"$ref": "literal", "properties": map[string]any{"const": "literal"}}},
	}
	properties := map[string]any{}
	for _, name := range []string{"type", "const", "enum", "examples", "deprecated", "$ref", "properties", "dependencies"} {
		properties[name] = map[string]any{"const": name, "examples": []any{"remove"}, "default": literal}
	}
	schema := map[string]any{"type": "object", "properties": properties, "examples": []any{"remove"}, "default": literal, "x-literal": literal}
	before, err := json.Marshal(schema)
	require.NoError(t, err)
	cleaned := cleanToolSchema(schema).(map[string]any)
	require.NotContains(t, cleaned, "examples")
	require.Equal(t, literal, cleaned["default"])
	require.Equal(t, literal, cleaned["x-literal"])
	gotProperties := cleaned["properties"].(map[string]any)
	require.Len(t, gotProperties, len(properties))
	for name := range properties {
		require.Equal(t, map[string]any{"type": "STRING", "enum": []any{name}, "default": literal}, gotProperties[name])
	}
	after, err := json.Marshal(schema)
	require.NoError(t, err)
	require.Equal(t, before, after, "normalization must not mutate caller input")
}

func TestCleanToolSchemaRecursiveLocationsV0217(t *testing.T) {
	for _, key := range []string{"items", "contains", "propertyNames", "if", "then", "else", "not", "contentSchema", "additionalItems", "unevaluatedItems", "unevaluatedProperties"} {
		t.Run(key, func(t *testing.T) {
			require.Equal(t, map[string]any{key: map[string]any{"type": "STRING", "enum": []any{"a"}}},
				cleanToolSchema(map[string]any{key: map[string]any{"const": "a", "examples": []any{"remove"}}}))
		})
	}
	for _, key := range []string{"anyOf", "allOf", "oneOf", "prefixItems", "items"} {
		t.Run(key+"/array", func(t *testing.T) {
			require.Equal(t, map[string]any{key: []any{map[string]any{"type": "ARRAY", "items": map[string]any{
				"type": "OBJECT", "properties": map[string]any{"const": map[string]any{"type": "STRING", "enum": []any{""}}},
			}}}}, cleanToolSchema(map[string]any{key: []any{map[string]any{"type": "array", "items": map[string]any{
				"type": "object", "properties": map[string]any{"const": map[string]any{"const": "", "examples": []any{"remove"}}},
			}}}}))
		})
	}
	for _, key := range []string{"dependencies", "dependentSchemas"} {
		t.Run(key, func(t *testing.T) {
			schema := map[string]any{key: map[string]any{"examples": map[string]any{"const": "a"}, "const": []any{"type", "examples"}}}
			require.Equal(t, map[string]any{key: map[string]any{"examples": map[string]any{"type": "STRING", "enum": []any{"a"}}, "const": []any{"type", "examples"}}}, cleanToolSchema(schema))
		})
	}
}
