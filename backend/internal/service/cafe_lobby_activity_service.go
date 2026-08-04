package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/redis/go-redis/v9"
)

const (
	cafeLobbyLabel          = "今日使用用户"
	cafeLobbyDisplayMax     = 50
	cafeLobbyFactTTL        = 72 * time.Hour
	cafeLobbyRecentWindow   = 15 * time.Minute
	cafeLobbyEventQueueSize = 4096
)

// CafeLobbyUsageRecorder is the repository-to-Cafe boundary for persisted usage.
// Implementations must return immediately; observation must never affect usage writes.
type CafeLobbyUsageRecorder interface {
	RecordPersistedUsage(userID int64, occurredAt time.Time)
}

type cafeLobbyUsageFact struct {
	userID     int64
	occurredAt time.Time
}

// CafeLobbyAvatar is an anonymous, bounded display projection.
type CafeLobbyAvatar struct {
	AvatarSeed string `json:"avatar_seed"`
	SeatIndex  int    `json:"seat_index"`
	Activity   string `json:"activity"`
}

// CafeLobbyActivity contains no user, request, account, key, or exact-time fields.
type CafeLobbyActivity struct {
	Available          bool              `json:"available"`
	Date               string            `json:"date"`
	Timezone           string            `json:"timezone"`
	Label              string            `json:"label"`
	UniqueUsers        int64             `json:"unique_users"`
	SuccessfulRequests int64             `json:"successful_requests"`
	DisplayMax         int               `json:"display_max"`
	Avatars            []CafeLobbyAvatar `json:"avatars"`
}

// CafeLobbyActivityService keeps the raw user IDs inside Redis and only returns
// date-scoped HMAC projections to callers.
type CafeLobbyActivityService struct {
	redis    *redis.Client
	secret   []byte
	location *time.Location
	tzName   string
	now      func() time.Time
	events   chan cafeLobbyUsageFact
}

func NewCafeLobbyActivityService(redisClient *redis.Client, cfg *config.Config) *CafeLobbyActivityService {
	loc := timezone.Location()
	tzName := timezone.Name()
	if cfg != nil && strings.TrimSpace(cfg.Timezone) != "" {
		if configured, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone)); err == nil {
			loc = configured
			tzName = strings.TrimSpace(cfg.Timezone)
		}
	}
	secret := []byte(nil)
	if cfg != nil {
		secret = []byte(cfg.JWT.Secret)
	}
	s := &CafeLobbyActivityService{
		redis:    redisClient,
		secret:   secret,
		location: loc,
		tzName:   tzName,
		now:      func() time.Time { return time.Now().In(loc) },
	}
	if redisClient != nil {
		s.events = make(chan cafeLobbyUsageFact, cafeLobbyEventQueueSize)
		go s.runRecorder()
	}
	return s
}

func (s *CafeLobbyActivityService) runRecorder() {
	for fact := range s.events {
		s.persistFact(fact)
	}
}

// RecordPersistedUsage is deliberately non-blocking and best effort.
func (s *CafeLobbyActivityService) RecordPersistedUsage(userID int64, occurredAt time.Time) {
	if s == nil || s.redis == nil || userID <= 0 {
		return
	}
	if occurredAt.IsZero() {
		occurredAt = s.now()
	}
	select {
	case s.events <- cafeLobbyUsageFact{userID: userID, occurredAt: occurredAt}:
	default:
		// Lobby is observational; a saturated queue must not slow the gateway.
	}
}

func (s *CafeLobbyActivityService) persistFact(fact cafeLobbyUsageFact) {
	if s == nil || s.redis == nil || fact.userID <= 0 {
		return
	}
	localTime := fact.occurredAt.In(s.location)
	date := localTime.Format("2006-01-02")
	userKey := cafeLobbyUsersKey(date)
	requestKey := cafeLobbyRequestsKey(date)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	pipe := s.redis.TxPipeline()
	// GT prevents delayed async events from moving the user's last-success score backwards.
	pipe.ZAddArgs(ctx, userKey, redis.ZAddArgs{GT: true, Members: []redis.Z{{Score: float64(fact.occurredAt.Unix()), Member: strconv.FormatInt(fact.userID, 10)}}})
	pipe.Incr(ctx, requestKey)
	pipe.Expire(ctx, userKey, cafeLobbyFactTTL)
	pipe.Expire(ctx, requestKey, cafeLobbyFactTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		logger.LegacyPrintf("service.cafe_lobby", "daily activity observation unavailable: %v", err)
	}
}

func (s *CafeLobbyActivityService) Snapshot(ctx context.Context) CafeLobbyActivity {
	if s == nil {
		return unavailableCafeLobbyActivity(time.Now())
	}
	now := s.now()
	result := CafeLobbyActivity{
		Available:  false,
		Date:       now.In(s.location).Format("2006-01-02"),
		Timezone:   s.tzName,
		Label:      cafeLobbyLabel,
		DisplayMax: cafeLobbyDisplayMax,
		Avatars:    []CafeLobbyAvatar{},
	}
	if s == nil || s.redis == nil {
		return result
	}
	date := result.Date
	userKey := cafeLobbyUsersKey(date)
	requestKey := cafeLobbyRequestsKey(date)
	pipe := s.redis.Pipeline()
	cardCmd := pipe.ZCard(ctx, userKey)
	requestCmd := pipe.Get(ctx, requestKey)
	rangeCmd := pipe.ZRevRangeWithScores(ctx, userKey, 0, int64(cafeLobbyDisplayMax-1))
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return result
	}
	uniqueUsers, err := cardCmd.Result()
	if err != nil {
		return result
	}
	requestCount, err := requestCmd.Int64()
	if err != nil && err != redis.Nil {
		return result
	}
	members, err := rangeCmd.Result()
	if err != nil && err != redis.Nil {
		return result
	}
	result.Available = true
	result.UniqueUsers = uniqueUsers
	result.SuccessfulRequests = requestCount
	result.Avatars = make([]CafeLobbyAvatar, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseInt(fmt.Sprint(member.Member), 10, 64)
		if err != nil || userID <= 0 {
			continue
		}
		seed, seatIndex := s.avatarProjection(date, userID)
		activity := "today"
		if !time.Unix(int64(member.Score), 0).Before(now.Add(-cafeLobbyRecentWindow)) {
			activity = "recent"
		}
		result.Avatars = append(result.Avatars, CafeLobbyAvatar{AvatarSeed: seed, SeatIndex: seatIndex, Activity: activity})
	}
	return result
}

func (s *CafeLobbyActivityService) avatarProjection(date string, userID int64) (string, int) {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte("pixel-cafe-lobby:" + date + ":" + strconv.FormatInt(userID, 10)))
	digest := mac.Sum(nil)
	seed := base64.RawURLEncoding.EncodeToString(digest[:])
	seatIndex := int(digest[0])%cafeLobbyDisplayMax + 1
	return seed[:16], seatIndex
}

func cafeLobbyUsersKey(date string) string {
	return "cafe:daily-users:" + date
}

func cafeLobbyRequestsKey(date string) string {
	return "cafe:daily-requests:" + date
}
