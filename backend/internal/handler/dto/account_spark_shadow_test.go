package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSparkShadowAccountFromServiceShallowIncludesLinkWithoutParentPII(t *testing.T) {
	parentID := int64(42)
	account := AccountFromServiceShallow(&service.Account{
		ID:              7,
		ParentAccountID: &parentID,
		QuotaDimension:  service.QuotaDimensionSpark,
	})
	require.Equal(t, &parentID, account.ParentAccountID)
	require.Equal(t, service.QuotaDimensionSpark, account.QuotaDimension)
}
