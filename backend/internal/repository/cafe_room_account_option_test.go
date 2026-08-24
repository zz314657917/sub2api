package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestCafeAccountOptionMasksOnlyCredentialsEmail(t *testing.T) {
	option := cafeRoomAccountEmailMasked(&dbent.Account{Credentials: map[string]interface{}{
		"email":        "owner@example.com",
		"api_key":      "must-not-leak",
		"access_token": "also-must-not-leak",
	}})
	require.Equal(t, "o***r@example.com", option)
	require.Empty(t, cafeRoomAccountEmailMasked(&dbent.Account{Credentials: map[string]interface{}{"email": "malformed"}}))
	require.Empty(t, cafeRoomAccountEmailMasked(&dbent.Account{Credentials: map[string]interface{}{"email": 123}}))
}
