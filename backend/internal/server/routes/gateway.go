package routes

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// RegisterGatewayRoutes 注册 API 网关路由（Claude/OpenAI/Gemini 兼容）
func RegisterGatewayRoutes(
	r *gin.Engine,
	h *handler.Handlers,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	opsService *service.OpsService,
	settingService *service.SettingService,
	cfg *config.Config,
) {
	bodyLimit := middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize)
	clientRequestID := middleware.ClientRequestID()
	opsErrorLogger := handler.OpsErrorLoggerMiddleware(opsService)
	endpointNorm := handler.InboundEndpointMiddleware()
	gatewayAuth := gatewayAuthMiddleware(apiKeyAuth, apiKeyService, settingService)

	// 未分组 Key 拦截中间件（按协议格式区分错误响应）
	requireGroupAnthropic := middleware.RequireGroupAssignment(settingService, middleware.AnthropicErrorWriter)
	requireGroupGoogle := middleware.RequireGroupAssignment(settingService, middleware.GoogleErrorWriter)

	// API网关（Claude API兼容）
	gateway := r.Group("/v1")
	gateway.Use(bodyLimit)
	gateway.Use(clientRequestID)
	gateway.Use(opsErrorLogger)
	gateway.Use(endpointNorm)
	gateway.Use(gatewayAuth)
	gateway.Use(requireGroupAnthropic)
	{
		// /v1/messages: auto-route based on group platform
		gateway.POST("/messages", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/messages", false) {
				return
			}
			if getGroupPlatform(c) == service.PlatformOpenAI {
				h.OpenAIGateway.Messages(c)
				return
			}
			h.Gateway.Messages(c)
		})
		// /v1/messages/count_tokens: OpenAI groups get 404
		gateway.POST("/messages/count_tokens", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/messages/count_tokens", false) {
				return
			}
			if getGroupPlatform(c) == service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"type": "error",
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Token counting is not supported for this platform",
					},
				})
				return
			}
			h.Gateway.CountTokens(c)
		})
		gateway.GET("/models", h.Gateway.Models)
		gateway.GET("/model-catalog", h.Gateway.ModelCatalog)
		gateway.GET("/usage", h.Gateway.Usage)
		// OpenAI Responses API: auto-route based on group platform
		gateway.POST("/responses", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/responses", false) {
				return
			}
			if getGroupPlatform(c) == service.PlatformOpenAI {
				h.OpenAIGateway.Responses(c)
				return
			}
			h.Gateway.Responses(c)
		})
		gateway.POST("/responses/*subpath", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/responses", false) {
				return
			}
			if getGroupPlatform(c) == service.PlatformOpenAI {
				h.OpenAIGateway.Responses(c)
				return
			}
			h.Gateway.Responses(c)
		})
		gateway.GET("/responses", h.OpenAIGateway.ResponsesWebSocket)
		// OpenAI Chat Completions API: auto-route based on group platform
		gateway.POST("/chat/completions", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/chat/completions", false) {
				return
			}
			if getGroupPlatform(c) == service.PlatformOpenAI {
				h.OpenAIGateway.ChatCompletions(c)
				return
			}
			h.Gateway.ChatCompletions(c)
		})
		gateway.POST("/embeddings", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/embeddings", false) {
				return
			}
			if getGroupPlatform(c) != service.PlatformOpenAI {
				service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Embeddings API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Embeddings(c)
		})
		gateway.POST("/images/generations", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/images/generations", true) {
				return
			}
			if getGroupPlatform(c) != service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Images API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Images(c)
		})
		gateway.POST("/midjourney/generations", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/midjourney/generations", true) {
				return
			}
			if getGroupPlatform(c) != service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Images API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Images(c)
		})
		gateway.POST("/images/edits", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/images/edits", true) {
				return
			}
			if getGroupPlatform(c) != service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Images API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Images(c)
		})
		gateway.POST("/videos/generations", func(c *gin.Context) {
			if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/videos/generations", false) {
				return
			}
			if getGroupPlatform(c) != service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Videos API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.Videos(c)
		})
		gateway.GET("/tasks/:task_id", func(c *gin.Context) {
			if getGroupPlatform(c) != service.PlatformOpenAI {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"type":    "not_found_error",
						"message": "Tasks API is not supported for this platform",
					},
				})
				return
			}
			h.OpenAIGateway.VideoTask(c)
		})
	}

	// Gemini 原生 API 兼容层（Gemini SDK/CLI 直连）
	gemini := r.Group("/v1beta")
	gemini.Use(bodyLimit)
	gemini.Use(clientRequestID)
	gemini.Use(opsErrorLogger)
	gemini.Use(endpointNorm)
	gemini.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	gemini.Use(requireGroupGoogle)
	{
		gemini.GET("/models", h.Gateway.GeminiV1BetaListModels)
		gemini.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		// Gin treats ":" as a param marker, but Gemini uses "{model}:{action}" in the same segment.
		gemini.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

	// OpenAI Responses API（不带v1前缀的别名）— auto-route based on group platform
	responsesHandler := func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/responses", false) {
			return
		}
		if getGroupPlatform(c) == service.PlatformOpenAI {
			h.OpenAIGateway.Responses(c)
			return
		}
		h.Gateway.Responses(c)
	}
	r.POST("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, responsesHandler)
	r.POST("/responses/*subpath", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, responsesHandler)
	r.GET("/responses", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, h.OpenAIGateway.ResponsesWebSocket)
	codexDirect := r.Group("/backend-api/codex")
	codexDirect.Use(bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic)
	{
		codexDirect.POST("/responses", responsesHandler)
		codexDirect.POST("/responses/*subpath", responsesHandler)
		codexDirect.GET("/responses", h.OpenAIGateway.ResponsesWebSocket)
	}
	// OpenAI Chat Completions API（不带v1前缀的别名）— auto-route based on group platform
	r.POST("/chat/completions", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/chat/completions", false) {
			return
		}
		if getGroupPlatform(c) == service.PlatformOpenAI {
			h.OpenAIGateway.ChatCompletions(c)
			return
		}
		h.Gateway.ChatCompletions(c)
	})
	r.POST("/embeddings", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/embeddings", false) {
			return
		}
		if getGroupPlatform(c) != service.PlatformOpenAI {
			service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalFeatureGate)
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Embeddings API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Embeddings(c)
	})
	r.POST("/images/generations", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/images/generations", true) {
			return
		}
		if getGroupPlatform(c) != service.PlatformOpenAI {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Images API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Images(c)
	})
	r.POST("/midjourney/generations", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/midjourney/generations", true) {
			return
		}
		if getGroupPlatform(c) != service.PlatformOpenAI {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Images API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Images(c)
	})
	r.POST("/images/edits", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/images/edits", true) {
			return
		}
		if getGroupPlatform(c) != service.PlatformOpenAI {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Images API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Images(c)
	})
	r.POST("/videos/generations", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if !resolveAPIKeyRouteForJSONModel(c, apiKeyService, "/v1/videos/generations", false) {
			return
		}
		if getGroupPlatform(c) != service.PlatformOpenAI {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Videos API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.Videos(c)
	})
	r.GET("/tasks/:task_id", bodyLimit, clientRequestID, opsErrorLogger, endpointNorm, gatewayAuth, requireGroupAnthropic, func(c *gin.Context) {
		if getGroupPlatform(c) != service.PlatformOpenAI {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"type":    "not_found_error",
					"message": "Tasks API is not supported for this platform",
				},
			})
			return
		}
		h.OpenAIGateway.VideoTask(c)
	})

	// Antigravity 模型列表
	r.GET("/antigravity/models", gatewayAuth, requireGroupAnthropic, h.Gateway.AntigravityModels)

	// Antigravity 专用路由（仅使用 antigravity 账户，不混合调度）
	antigravityV1 := r.Group("/antigravity/v1")
	antigravityV1.Use(bodyLimit)
	antigravityV1.Use(clientRequestID)
	antigravityV1.Use(opsErrorLogger)
	antigravityV1.Use(endpointNorm)
	antigravityV1.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1.Use(gatewayAuth)
	antigravityV1.Use(requireGroupAnthropic)
	{
		antigravityV1.POST("/messages", h.Gateway.Messages)
		antigravityV1.POST("/messages/count_tokens", h.Gateway.CountTokens)
		antigravityV1.GET("/models", h.Gateway.AntigravityModels)
		antigravityV1.GET("/usage", h.Gateway.Usage)
	}

	antigravityV1Beta := r.Group("/antigravity/v1beta")
	antigravityV1Beta.Use(bodyLimit)
	antigravityV1Beta.Use(clientRequestID)
	antigravityV1Beta.Use(opsErrorLogger)
	antigravityV1Beta.Use(endpointNorm)
	antigravityV1Beta.Use(middleware.ForcePlatform(service.PlatformAntigravity))
	antigravityV1Beta.Use(middleware.APIKeyAuthWithSubscriptionGoogle(apiKeyService, subscriptionService, cfg))
	antigravityV1Beta.Use(requireGroupGoogle)
	{
		antigravityV1Beta.GET("/models", h.Gateway.GeminiV1BetaListModels)
		antigravityV1Beta.GET("/models/:model", h.Gateway.GeminiV1BetaGetModel)
		antigravityV1Beta.POST("/models/*modelAction", h.Gateway.GeminiV1BetaModels)
	}

}

func gatewayAuthMiddleware(apiKeyAuth middleware.APIKeyAuthMiddleware, apiKeyService *service.APIKeyService, settingService *service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader("X-Sub2API-Studio-Secret")) == "" {
			gin.HandlerFunc(apiKeyAuth)(c)
			return
		}
		if !authenticateStudioBridgeGateway(c, apiKeyService, settingService) {
			return
		}
		c.Next()
	}
}

func authenticateStudioBridgeGateway(c *gin.Context, apiKeyService *service.APIKeyService, settingService *service.SettingService) bool {
	if c == nil || c.Request == nil || apiKeyService == nil || settingService == nil {
		middleware.AbortWithError(c, http.StatusServiceUnavailable, "STUDIO_BRIDGE_GATEWAY_UNAVAILABLE", "studio bridge gateway auth is unavailable")
		return false
	}
	cfg, err := settingService.GetStudioBridgeLuoyeAISettings(c.Request.Context())
	if err != nil {
		middleware.AbortWithError(c, http.StatusInternalServerError, "STUDIO_BRIDGE_GATEWAY_CONFIG_ERROR", "failed to load studio bridge settings")
		return false
	}
	if cfg == nil || !cfg.Enabled || strings.TrimSpace(cfg.InternalSecret) == "" {
		middleware.AbortWithError(c, http.StatusForbidden, "STUDIO_BRIDGE_DISABLED", "studio bridge is disabled")
		return false
	}
	secret := strings.TrimSpace(c.GetHeader("X-Sub2API-Studio-Secret"))
	if !studioBridgeSecretEqual(cfg.InternalSecret, secret) {
		middleware.AbortWithError(c, http.StatusUnauthorized, "STUDIO_BRIDGE_INVALID_SECRET", "invalid studio bridge secret")
		return false
	}
	userID, err := parsePositiveInt64Header(c, "X-Sub2API-Studio-User-ID")
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, "STUDIO_BRIDGE_USER_REQUIRED", "studio bridge user id is required")
		return false
	}
	groupID := parseOptionalPositiveInt64Header(c, "X-Sub2API-Group-ID")
	apiKey, err := apiKeyService.BuildStudioBridgeGatewayAPIKey(c.Request.Context(), userID, groupID, c.Request.URL.Path)
	if err != nil {
		middleware.AbortWithError(c, http.StatusForbidden, "STUDIO_BRIDGE_GATEWAY_FORBIDDEN", err.Error())
		return false
	}
	if !apiKey.User.IsActive() {
		middleware.AbortWithError(c, http.StatusUnauthorized, "USER_INACTIVE", "User account is not active")
		return false
	}
	if !validateStudioBridgeGatewayGroup(c, apiKey) {
		return false
	}
	middleware.SetAPIKeyContext(c, apiKey)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxkey.StudioBridgeGateway, true))
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{
		UserID:      apiKey.User.ID,
		Concurrency: apiKey.User.Concurrency,
	})
	c.Set(string(middleware.ContextKeyUserRole), apiKey.User.Role)
	c.Set(string(middleware.ContextKeyStudioBridgeGateway), true)
	return true
}

func studioBridgeSecretEqual(expected, provided string) bool {
	expected = strings.TrimSpace(expected)
	provided = strings.TrimSpace(provided)
	if expected == "" || provided == "" || len(expected) != len(provided) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}

func validateStudioBridgeGatewayGroup(c *gin.Context, apiKey *service.APIKey) bool {
	if apiKey == nil || apiKey.GroupID == nil || apiKey.Group == nil {
		middleware.AbortWithError(c, http.StatusForbidden, "GROUP_NOT_FOUND", "studio bridge group is not configured")
		return false
	}
	if !apiKey.Group.IsActive() {
		middleware.AbortWithError(c, http.StatusForbidden, "GROUP_DISABLED", "studio bridge group is disabled")
		return false
	}
	return true
}

func parsePositiveInt64Header(c *gin.Context, name string) (int64, error) {
	value := strings.TrimSpace(c.GetHeader(name))
	if value == "" {
		return 0, strconv.ErrSyntax
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func parseOptionalPositiveInt64Header(c *gin.Context, name string) int64 {
	parsed, err := parsePositiveInt64Header(c, name)
	if err != nil {
		return 0
	}
	return parsed
}

// getGroupPlatform extracts the group platform from the API Key stored in context.
func getGroupPlatform(c *gin.Context) string {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		return ""
	}
	return apiKey.Group.Platform
}

func resolveAPIKeyRouteForJSONModel(c *gin.Context, apiKeyService *service.APIKeyService, endpoint string, imageEndpoint bool) bool {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return true
	}
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil {
		return true
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": gin.H{
					"type":    "invalid_request_error",
					"message": "Request body too large",
				},
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"type":    "invalid_request_error",
					"message": "Failed to read request body",
				},
			})
		}
		c.Abort()
		return false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	requestedModel := ""
	if gjson.ValidBytes(body) {
		requestedModel = gjson.GetBytes(body, "model").String()
	}
	imageIntent := imageEndpoint || service.IsImageGenerationIntent(endpoint, requestedModel, body)
	if _, ok := middleware.ResolveAPIKeyForModelRequest(c, apiKeyService, apiKey, requestedModel, imageIntent); !ok {
		c.Abort()
		return false
	}
	return true
}
