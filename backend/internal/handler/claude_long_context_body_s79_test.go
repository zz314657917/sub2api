package handler

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestS79AnthropicNormalizedBodyPropagation(t *testing.T) {
	t.Run("shared parser returns the normalized body", func(t *testing.T) {
		parsedReq, body, err := parseAnthropicGatewayRequestBody([]byte(`{"model":"claude-opus-4-8[1M][1m]","messages":[{"role":"user","content":"hi"}]}`))
		require.NoError(t, err)
		require.Equal(t, "claude-opus-4-8", parsedReq.Model)
		require.Equal(t, "claude-opus-4-8", gjson.GetBytes(body, "model").String())
		require.Equal(t, parsedReq.Body, body)
	})

	t.Run("all handler call sites overwrite the request body", func(t *testing.T) {
		targets := map[string]map[string]bool{
			"GatewayHandler.Messages": {
				"SetClaudeCodeClientContext": false,
				"setOpsRequestContext":       false,
				"checkSecurityAudit":         false,
			},
			"GatewayHandler.CountTokens": {
				"SetClaudeCodeClientContext": false,
				"setOpsRequestContext":       false,
			},
			"OpenAIGatewayHandler.CountTokens": {
				"setOpsRequestContext":          false,
				"newOpenAIModelMappedBodyCache": false,
				"GenerateSessionHash":           false,
			},
		}
		callName := func(call *ast.CallExpr) string {
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				return fn.Name
			case *ast.SelectorExpr:
				return fn.Sel.Name
			default:
				return ""
			}
		}
		hasBodyArg := func(call *ast.CallExpr) bool {
			for _, arg := range call.Args {
				if ident, ok := arg.(*ast.Ident); ok && ident.Name == "body" {
					return true
				}
			}
			return false
		}

		fset := token.NewFileSet()
		for _, filename := range []string{"gateway_handler.go", "openai_gateway_count_tokens.go"} {
			file, err := parser.ParseFile(fset, filename, nil, 0)
			require.NoError(t, err)
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
					continue
				}
				receiver := fn.Recv.List[0].Type
				if star, ok := receiver.(*ast.StarExpr); ok {
					receiver = star.X
				}
				receiverIdent, ok := receiver.(*ast.Ident)
				if !ok {
					continue
				}
				key := receiverIdent.Name + "." + fn.Name.Name
				expectedConsumers, wanted := targets[key]
				if !wanted {
					continue
				}

				var propagationPos token.Pos
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					assign, ok := node.(*ast.AssignStmt)
					if !ok || len(assign.Lhs) != 3 || len(assign.Rhs) != 1 {
						return true
					}
					bodyIdent, ok := assign.Lhs[1].(*ast.Ident)
					if !ok || bodyIdent.Name != "body" {
						return true
					}
					call, ok := assign.Rhs[0].(*ast.CallExpr)
					if !ok {
						return true
					}
					helper, ok := call.Fun.(*ast.Ident)
					if ok && helper.Name == "parseAnthropicGatewayRequestBody" {
						propagationPos = assign.Pos()
					}
					return true
				})
				require.NotEqualf(t, token.NoPos, propagationPos, "%s must overwrite body from parseAnthropicGatewayRequestBody", key)

				ast.Inspect(fn.Body, func(node ast.Node) bool {
					call, ok := node.(*ast.CallExpr)
					if !ok || call.Pos() <= propagationPos || !hasBodyArg(call) {
						return true
					}
					name := callName(call)
					if _, expected := expectedConsumers[name]; expected {
						expectedConsumers[name] = true
					}
					return true
				})
			}
		}

		for target, consumers := range targets {
			for consumer, found := range consumers {
				require.Truef(t, found, "%s must pass the normalized body to %s", target, consumer)
			}
		}
	})
}
