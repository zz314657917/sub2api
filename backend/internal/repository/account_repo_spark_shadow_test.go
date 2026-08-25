package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/account"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestAccountSparkShadowEntityMapping(t *testing.T) {
	parentID := int64(42)
	mapped := accountEntityToService(&dbent.Account{
		ID:              43,
		Platform:        service.PlatformOpenAI,
		Type:            service.AccountTypeOAuth,
		ParentAccountID: &parentID,
		QuotaDimension:  account.QuotaDimensionSpark,
	})
	if mapped.ParentAccountID == nil || *mapped.ParentAccountID != parentID {
		t.Fatal("parent account id was not mapped")
	}
	if mapped.QuotaDimensionOrDefault() != service.QuotaDimensionSpark {
		t.Fatal("spark quota dimension was not mapped")
	}
}
