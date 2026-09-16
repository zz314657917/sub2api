package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/securityaudit"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestWireProviderOpenAIGatewayHandlerPreservesOpsAndCoordinator(t *testing.T) {
	opsService := &service.OpsService{}
	coordinator := &securityaudit.Coordinator{}

	h := ProvideOpenAIGatewayHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, coordinator, opsService)

	if h == nil {
		t.Fatal("provider returned nil handler")
	}
	if h.opsService != opsService {
		t.Fatal("provider did not preserve the operations service")
	}
	if h.securityAuditCoordinator != coordinator {
		t.Fatal("provider did not preserve the security audit coordinator")
	}
}
