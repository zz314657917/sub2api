package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func keyFast20261105Context(policy string, groupForce bool) context.Context {
	return context.WithValue(groupFast20261103Context(PlatformComposite, groupForce), ctxkey.APIKeyOpenAIFastPolicy, policy)
}

func TestKeyFast20261105PrecedenceHTTPAndWS(t *testing.T) {
	for _, policy := range []string{"", "passthrough", "force", "filter"} {
		for _, groupForce := range []bool{false, true} {
			for _, action := range []string{"pass", "filter", "block"} {
				for _, tier := range []string{"", "default", "flex", "ultrafast", "unknown", "fast", "priority", "  PrIoRiTy  "} {
					t.Run(fmt.Sprintf("%s/group=%v/%s/%s", policy, groupForce, action, tier), func(t *testing.T) {
						ctx := keyFast20261105Context(policy, groupForce)
						svc := newOpenAIGatewayServiceWithSettings(t, &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{
							ServiceTier: OpenAIFastTierPriority, Action: action, Scope: BetaPolicyScopeAll,
						}}})
						input := map[string]any{"model": "gpt-5.5", "input": "keep"}
						if tier != "" {
							input["service_tier"] = tier
						}
						want := strings.ToLower(strings.TrimSpace(tier))
						if want == "fast" {
							want = "priority"
						}
						if groupForce || policy == "force" {
							want = "priority"
						} else if policy == "filter" && want == "priority" {
							want = ""
						}
						blocked := want == "priority" && action == "block"
						if want == "priority" && action == "filter" {
							want = ""
						}
						raw, err := json.Marshal(input)
						require.NoError(t, err)
						account := &Account{Platform: PlatformOpenAI}
						out, err := svc.applyOpenAIFastPolicyToBody(ctx, account, "gpt-5.5", raw)
						if blocked {
							require.IsType(t, &OpenAIFastBlockedError{}, err)
						} else {
							require.NoError(t, err)
							require.Equal(t, want, gjson.GetBytes(out, "service_tier").String())
							require.Equal(t, "keep", gjson.GetBytes(out, "input").String())
						}
						input["type"] = "response.create"
						frame, err := json.Marshal(input)
						require.NoError(t, err)
						ws, wsBlocked, err := svc.applyOpenAIFastPolicyToWSResponseCreate(ctx, account, "gpt-5.5", frame)
						require.NoError(t, err)
						require.Equal(t, blocked, wsBlocked != nil)
						if !blocked {
							require.Equal(t, want, gjson.GetBytes(ws, "service_tier").String())
						}
					})
				}
			}
		}
	}
}

func TestKeyFast20261105TrustProvidersAndOtherWSEvents(t *testing.T) {
	svc := newOpenAIGatewayServiceWithSettings(t, DefaultOpenAIFastPolicySettings())
	for _, ctx := range []context.Context{
		context.Background(),
		context.WithValue(context.Background(), "ctx_api_key_openai_fast_policy", "force"),
		context.WithValue(context.Background(), ctxkey.APIKeyOpenAIFastPolicy, map[string]string{"policy": "force"}),
	} {
		raw := []byte(`{"model":"gpt-5.5","openai_fast_policy":"force"}`)
		out, err := svc.applyOpenAIFastPolicyToBody(ctx, &Account{Platform: PlatformOpenAI}, "gpt-5.5", raw)
		require.NoError(t, err)
		require.Equal(t, raw, out)
	}
	for _, platform := range []string{PlatformAnthropic, PlatformGemini, PlatformGrok, PlatformAntigravity} {
		for _, policy := range []string{"force", "filter"} {
			raw := []byte(`{"model":"test","service_tier":"flex"}`)
			out, err := svc.applyOpenAIFastPolicyToBody(keyFast20261105Context(policy, true), &Account{Platform: platform}, "test", raw)
			require.NoError(t, err)
			require.Equal(t, raw, out)
		}
	}
	for _, raw := range []string{
		`{"type":"response.cancel","service_tier":"priority"}`,
		`{"type":"session.update","service_tier":"flex"}`,
		`{"type":"conversation.item.create","service_tier":"priority"}`,
		`{"service_tier":"priority"}`, `null`, `invalid`, ``,
	} {
		out, blocked, err := svc.applyOpenAIFastPolicyToWSResponseCreate(keyFast20261105Context("force", true), &Account{Platform: PlatformOpenAI}, "", []byte(raw))
		require.NoError(t, err)
		require.Nil(t, blocked)
		require.Equal(t, raw, string(out))
	}
	for _, value := range []string{`null`, `true`, `1`, `[]`, `{}`} {
		out, err := svc.applyOpenAIFastPolicyToBody(keyFast20261105Context("force", false),
			&Account{Platform: PlatformOpenAI}, "", []byte(`{"service_tier":`+value+`}`))
		require.NoError(t, err)
		require.Equal(t, "priority", gjson.GetBytes(out, "service_tier").String())
	}
	// The existing global flex filter remains authoritative for key filter.
	svc = newOpenAIGatewayServiceWithSettings(t, &OpenAIFastPolicySettings{Rules: []OpenAIFastPolicyRule{{
		ServiceTier: OpenAIFastTierFlex, Action: BetaPolicyActionFilter, Scope: BetaPolicyScopeAll,
	}}})
	out, err := svc.applyOpenAIFastPolicyToBody(keyFast20261105Context("filter", false),
		&Account{Platform: PlatformOpenAI}, "", []byte(`{"service_tier":"flex"}`))
	require.NoError(t, err)
	require.False(t, gjson.GetBytes(out, "service_tier").Exists())
}

type keyFast20261105Repo struct {
	APIKeyRepository
	key       *APIKey
	authLoads int
	writes    []APIKeyUpdateFields
}

func (r *keyFast20261105Repo) Create(_ context.Context, key *APIKey) error {
	key.ID = 7
	clone := *key
	r.key = &clone
	return nil
}
func (r *keyFast20261105Repo) GetByID(context.Context, int64) (*APIKey, error) {
	clone := *r.key
	return &clone, nil
}
func (r *keyFast20261105Repo) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	r.authLoads++
	return r.GetByID(context.Background(), 7)
}
func (r *keyFast20261105Repo) ListByUserID(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return []APIKey{*r.key}, &pagination.PaginationResult{Total: 1}, nil
}
func (r *keyFast20261105Repo) Update(_ context.Context, key *APIKey, mask APIKeyUpdateFields) error {
	r.writes = append(r.writes, mask)
	if mask.OpenAIFastPolicy {
		r.key.OpenAIFastPolicy = key.OpenAIFastPolicy
	}
	return nil
}
func (*keyFast20261105Repo) UpdateLastUsed(context.Context, int64, time.Time) error { return nil }

type keyFast20261105UserRepo struct{ UserRepository }

func (*keyFast20261105UserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	return &User{ID: id, Status: StatusActive}, nil
}

type keyFast20261105Cache struct {
	APIKeyCache
	entry     *APIKeyAuthCacheEntry
	deletes   int
	published []string
}

func (c *keyFast20261105Cache) GetAuthCache(context.Context, string) (*APIKeyAuthCacheEntry, error) {
	return c.entry, nil
}
func (c *keyFast20261105Cache) SetAuthCache(_ context.Context, _ string, entry *APIKeyAuthCacheEntry, _ time.Duration) error {
	wire, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	c.entry = &APIKeyAuthCacheEntry{}
	return json.Unmarshal(wire, c.entry)
}
func (c *keyFast20261105Cache) DeleteAuthCache(context.Context, string) error {
	c.entry = nil
	c.deletes++
	return nil
}
func (c *keyFast20261105Cache) PublishAuthCacheInvalidation(_ context.Context, key string) error {
	c.published = append(c.published, key)
	return nil
}

func TestKeyFast20261105ServiceCreateUpdateOwnershipAndManaged(t *testing.T) {
	for _, policy := range []string{"", "passthrough", "force", "filter"} {
		t.Run(policy, func(t *testing.T) {
			repo := &keyFast20261105Repo{}
			svc := NewAPIKeyService(repo, &keyFast20261105UserRepo{}, nil, nil, nil, nil, &config.Config{})
			created, err := svc.Create(context.Background(), 42, CreateAPIKeyRequest{Name: "test", OpenAIFastPolicy: policy})
			require.NoError(t, err)
			require.Equal(t, NormalizeAPIKeyOpenAIFastPolicy(policy), created.OpenAIFastPolicy)
			repo.key.QuotaUsed, repo.key.Usage5h = 17, 9
			unchanged, err := svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{})
			require.NoError(t, err)
			require.Equal(t, created.OpenAIFastPolicy, unchanged.OpenAIFastPolicy)
			require.True(t, repo.writes[0].IsEmpty())
			for _, next := range []string{"force", "filter", "passthrough"} {
				updated, err := svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{OpenAIFastPolicy: &next})
				require.NoError(t, err)
				require.Equal(t, next, updated.OpenAIFastPolicy)
				require.Equal(t, APIKeyUpdateFields{OpenAIFastPolicy: true}, repo.writes[len(repo.writes)-1])
				require.Equal(t, float64(17), updated.QuotaUsed)
				require.Equal(t, float64(9), updated.Usage5h)
			}
			force := "force"
			_, err = svc.Update(context.Background(), 7, 999, UpdateAPIKeyRequest{OpenAIFastPolicy: &force})
			require.ErrorIs(t, err, ErrInsufficientPerms)
			for _, source := range []string{APIKeyManagedSourceCafeRoomSeat, APIKeyManagedSourceCafeRoomMembership} {
				repo.key.ManagedSourceType = source
				sourceID := int64(77)
				repo.key.ManagedSourceID = &sourceID
				before := len(repo.writes)
				_, err = svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{OpenAIFastPolicy: &force})
				require.ErrorIs(t, err, ErrCafeManagedKeyProtected)
				require.Len(t, repo.writes, before)
				name := "safe rename"
				_, err = svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{Name: &name})
				require.NoError(t, err)
				require.Equal(t, APIKeyUpdateFields{Name: true}, repo.writes[len(repo.writes)-1])
			}
		})
	}
	for _, invalid := range []string{"bad", "Force", " force", "null"} {
		svc := &APIKeyService{}
		_, err := svc.Create(context.Background(), 42, CreateAPIKeyRequest{OpenAIFastPolicy: invalid})
		require.ErrorIs(t, err, ErrAPIKeyOpenAIFastPolicyInvalid)
		_, err = svc.Update(context.Background(), 7, 42, UpdateAPIKeyRequest{OpenAIFastPolicy: &invalid})
		require.ErrorIs(t, err, ErrAPIKeyOpenAIFastPolicyInvalid)
	}
	empty := ""
	_, err := (&APIKeyService{}).Update(context.Background(), 7, 42, UpdateAPIKeyRequest{OpenAIFastPolicy: &empty})
	require.ErrorIs(t, err, ErrAPIKeyOpenAIFastPolicyInvalid)
}

func TestKeyFast20261105CacheVersionRoundTripAndInvalidation(t *testing.T) {
	repo := &keyFast20261105Repo{key: &APIKey{ID: 7, UserID: 42, Key: "key-fast-local",
		OpenAIFastPolicy: "force", Status: StatusActive, User: &User{ID: 42, Status: StatusActive}}}
	cache := &keyFast20261105Cache{}
	cfg := &config.Config{APIKeyAuth: config.APIKeyAuthCacheConfig{L1Size: 1024, L1TTLSeconds: 60, L2TTLSeconds: 60}}
	svc := NewAPIKeyService(repo, nil, nil, nil, nil, cache, cfg)
	t.Cleanup(svc.authCacheL1.Close)
	ctx := context.Background()
	snapshot := svc.snapshotFromAPIKey(ctx, repo.key)
	require.Equal(t, 21, snapshot.Version)
	snapshot.Version = 20
	snapshot.OpenAIFastPolicy = ""
	cache.entry = &APIKeyAuthCacheEntry{Snapshot: snapshot}
	for _, version := range []int{0, 19, 20} {
		snapshot.Version = version
		key, hit, err := svc.applyAuthCacheEntry(repo.key.Key, cache.entry)
		require.NoError(t, err)
		require.False(t, hit)
		require.Nil(t, key)
	}
	key, err := svc.GetByKey(ctx, repo.key.Key)
	require.NoError(t, err)
	require.Equal(t, "force", key.OpenAIFastPolicy)
	require.Equal(t, 1, repo.authLoads)
	svc.authCacheL1.Wait()
	require.Equal(t, "force", cache.entry.Snapshot.OpenAIFastPolicy)
	key, err = svc.GetByKey(ctx, repo.key.Key)
	require.NoError(t, err)
	require.Equal(t, "force", key.OpenAIFastPolicy)
	require.Equal(t, 1, repo.authLoads, "L1 hit must avoid repository")
	filter := "filter"
	_, err = svc.Update(ctx, 7, 42, UpdateAPIKeyRequest{OpenAIFastPolicy: &filter})
	require.NoError(t, err)
	svc.authCacheL1.Wait()
	require.Nil(t, cache.entry)
	require.Equal(t, 1, cache.deletes)
	require.Equal(t, []string{svc.authCacheKey(repo.key.Key)}, cache.published)
	key, err = svc.GetByKey(ctx, repo.key.Key)
	require.NoError(t, err)
	require.Equal(t, "filter", key.OpenAIFastPolicy)
	require.Equal(t, 2, repo.authLoads)
	svc.authCacheL1.Wait()
	svc.authCacheL1.Del(svc.authCacheKey(repo.key.Key))
	svc.authCacheL1.Wait()
	key, err = svc.GetByKey(ctx, repo.key.Key)
	require.NoError(t, err)
	require.Equal(t, "filter", key.OpenAIFastPolicy)
	require.Equal(t, 2, repo.authLoads, "L2 serialized snapshot must retain policy")
}
