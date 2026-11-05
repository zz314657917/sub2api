package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func groupFast20261103Context(platform string, enabled bool) context.Context {
	return context.WithValue(context.Background(), ctxkey.Group, &Group{
		ID: 7, Platform: platform, Status: StatusActive, Hydrated: true, ForceOpenAIFast: enabled,
	})
}

func TestGroupFast20261103PolicyMatrix(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformComposite} {
		for _, tier := range []string{"", "default", "flex", "ultrafast", "unknown", "priority"} {
			for _, action := range []string{BetaPolicyActionPass, BetaPolicyActionFilter, BetaPolicyActionBlock} {
				t.Run(platform+"/"+tier+"/"+action, func(t *testing.T) {
					policy := &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{
						ServiceTier: OpenAIFastTierPriority, Action: action, Scope: BetaPolicyScopeAll,
					}}}
					svc := newOpenAIGatewayServiceWithSettings(t, policy)
					account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
					ctx := groupFast20261103Context(platform, true)
					body := map[string]any{"model": "gpt-5.5", "input": "keep"}
					if tier != "" {
						body["service_tier"] = tier
					}
					raw, err := json.Marshal(body)
					require.NoError(t, err)
					got, httpErr := svc.applyOpenAIFastPolicyToBody(ctx, account, "gpt-5.5", raw)
					if action == BetaPolicyActionBlock {
						require.IsType(t, &OpenAIFastBlockedError{}, httpErr)
					} else {
						require.NoError(t, httpErr)
						want := "priority"
						if action == BetaPolicyActionFilter {
							want = ""
						}
						require.Equal(t, want, gjson.GetBytes(got, "service_tier").String())
						require.Equal(t, "keep", gjson.GetBytes(got, "input").String())
					}
					body["type"] = "response.create"
					frame, err := json.Marshal(body)
					require.NoError(t, err)
					ws, blocked, wsErr := svc.applyOpenAIFastPolicyToWSResponseCreate(ctx, account, "gpt-5.5", frame)
					require.NoError(t, wsErr)
					if action == BetaPolicyActionBlock {
						require.NotNil(t, blocked)
					} else {
						require.Nil(t, blocked)
						require.JSONEq(t, string(got), string(removeGroupFastFrameType(t, ws)))
					}
				})
			}
		}
	}
}

func removeGroupFastFrameType(t *testing.T, raw []byte) []byte {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	delete(body, "type")
	got, err := json.Marshal(body)
	require.NoError(t, err)
	return got
}

func TestGroupFast20261103TrustAndLegacy(t *testing.T) {
	valid := &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ForceOpenAIFast: true}
	for _, tc := range []struct {
		name    string
		ctx     context.Context
		account *Account
	}{
		{"nil-context", nil, &Account{Platform: PlatformOpenAI}},
		{"missing-context", context.Background(), &Account{Platform: PlatformOpenAI}},
		{"disabled", groupFast20261103Context(PlatformOpenAI, false), &Account{Platform: PlatformOpenAI}},
		{"unsupported-group", groupFast20261103Context(PlatformAnthropic, true), &Account{Platform: PlatformOpenAI}},
		{"wrong-provider", groupFast20261103Context(PlatformComposite, true), &Account{Platform: PlatformGrok}},
		{"nil-account", groupFast20261103Context(PlatformOpenAI, true), nil},
		{"wrong-context-type", context.WithValue(context.Background(), ctxkey.Group, map[string]any{"force_openai_fast": true}), &Account{Platform: PlatformOpenAI}},
		{"unhydrated", context.WithValue(context.Background(), ctxkey.Group, &Group{ID: 7, Platform: PlatformOpenAI, Status: StatusActive, ForceOpenAIFast: true}), &Account{Platform: PlatformOpenAI}},
		{"zero-id", context.WithValue(context.Background(), ctxkey.Group, &Group{Platform: PlatformOpenAI, Status: StatusActive, Hydrated: true, ForceOpenAIFast: true}), &Account{Platform: PlatformOpenAI}},
		{"empty-status", context.WithValue(context.Background(), ctxkey.Group, &Group{ID: 7, Platform: PlatformOpenAI, Hydrated: true, ForceOpenAIFast: true}), &Account{Platform: PlatformOpenAI}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.False(t, openAIGroupForcesFast(tc.ctx, tc.account))
			if tc.ctx == nil {
				return
			}
			svc := newOpenAIGatewayServiceWithSettings(t, DefaultOpenAIFastPolicySettings())
			for _, tier := range []string{"", "default", "flex", "ultrafast", "unknown"} {
				body := []byte(`{"model":"gpt-5.5","group_id":7,"force_openai_fast":true}`)
				if tier != "" {
					body = []byte(`{"model":"gpt-5.5","service_tier":"` + tier + `"}`)
				}
				got, err := svc.applyOpenAIFastPolicyToBody(tc.ctx, tc.account, "gpt-5.5", body)
				require.NoError(t, err)
				require.Equal(t, body, got)
				frame := append([]byte(`{"type":"response.create",`), body[1:]...)
				ws, blocked, err := svc.applyOpenAIFastPolicyToWSResponseCreate(tc.ctx, tc.account, "gpt-5.5", frame)
				require.NoError(t, err)
				require.Nil(t, blocked)
				require.Equal(t, frame, ws)
			}
		})
	}
	require.True(t, openAIGroupForcesFast(context.WithValue(context.Background(), ctxkey.Group, valid), &Account{Platform: PlatformOpenAI}))
}

func TestGroupFast20261103UnrelatedFramesUntouched(t *testing.T) {
	svc := newOpenAIGatewayServiceWithSettings(t, openAIFastFilterPriorityPolicy())
	ctx := groupFast20261103Context(PlatformOpenAI, true)
	for _, frame := range []string{
		`{"type":"response.cancel","service_tier":"flex"}`,
		`{"type":"session.update","session":{"model":"gpt-5.5"}}`,
		`{"type":"conversation.item.create","service_tier":"priority"}`,
		`{"service_tier":"priority"}`, `invalid`, ``, `null`,
	} {
		got, blocked, err := svc.applyOpenAIFastPolicyToWSResponseCreate(ctx, &Account{Platform: PlatformOpenAI}, "gpt-5.5", []byte(frame))
		require.NoError(t, err)
		require.Nil(t, blocked)
		require.Equal(t, frame, string(got))
	}
}

func TestGroupFast20261103AuthSnapshotAndDuplicate(t *testing.T) {
	ctx := groupFast20261103Context(PlatformOpenAI, true)
	group := ctx.Value(ctxkey.Group).(*Group)
	group.LongContextPricingEnabled = true
	svc := &APIKeyService{}
	key := &APIKey{ID: 1, User: &User{ID: 2}, Group: group, MultiGroupRouteGroups: []*Group{group}}
	snapshot := svc.snapshotFromAPIKey(ctx, key)
	require.Equal(t, 21, snapshot.Version)
	wire, err := json.Marshal(snapshot)
	require.NoError(t, err)
	var decoded APIKeyAuthSnapshot
	require.NoError(t, json.Unmarshal(wire, &decoded))
	restored, hit, err := svc.applyAuthCacheEntry("test", &APIKeyAuthCacheEntry{Snapshot: &decoded})
	require.NoError(t, err)
	require.True(t, hit)
	require.True(t, restored.Group.ForceOpenAIFast)
	require.True(t, restored.Group.LongContextPricingEnabled)
	require.True(t, restored.Group.Hydrated)
	require.Len(t, restored.MultiGroupRouteGroups, 1)
	require.True(t, restored.MultiGroupRouteGroups[0].ForceOpenAIFast)
	require.True(t, restored.MultiGroupRouteGroups[0].Hydrated)
	decoded.Version = 19
	stale, hit, err := svc.applyAuthCacheEntry("test", &APIKeyAuthCacheEntry{Snapshot: &decoded})
	require.NoError(t, err)
	require.False(t, hit)
	require.Nil(t, stale)
	group.Name = "Fast group"
	repo := newDuplicateGroupRepoStub(group)
	admin := &adminServiceImpl{groupRepo: repo, groupDuplicateRepo: repo}
	duplicate, err := admin.DuplicateGroup(ctx, group.ID, "admin:7", "group-fast-20261103")
	require.NoError(t, err)
	require.True(t, duplicate.ForceOpenAIFast)
	require.True(t, group.ForceOpenAIFast)
}
