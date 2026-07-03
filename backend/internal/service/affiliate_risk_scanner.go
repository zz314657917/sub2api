package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	riskip "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	affiliateRiskScannerJobName = "affiliate_risk_scanner"

	affiliateRiskWindow          = 12 * time.Hour
	affiliateRiskTimeout         = 45 * time.Second
	affiliateRiskRewardFastAfter = 30 * time.Minute
	affiliateRiskLeaderLockKey   = "ops:affiliate:risk:leader"
	affiliateRiskLeaderLockTTL   = 90 * time.Second
)

var affiliateRiskReleaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`)

type AffiliateRiskScannerService struct {
	repo           AffiliateRiskRepository
	opsRepo        OpsRepository
	opsService     *OpsService
	settingService *SettingService
	emailService   *EmailService
	redisClient    *redis.Client
	cfg            *config.Config
	instanceID     string

	stopCh    chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup

	skipLogMu sync.Mutex
	skipLogAt time.Time
}

type affiliateRiskScoreResult struct {
	Score       int
	Severity    string
	Fingerprint string
	Title       string
	Description string
	Reasons     []string
}

func NewAffiliateRiskScannerService(
	repo AffiliateRiskRepository,
	opsRepo OpsRepository,
	opsService *OpsService,
	settingService *SettingService,
	emailService *EmailService,
	redisClient *redis.Client,
	cfg *config.Config,
) *AffiliateRiskScannerService {
	return &AffiliateRiskScannerService{
		repo:           repo,
		opsRepo:        opsRepo,
		opsService:     opsService,
		settingService: settingService,
		emailService:   emailService,
		redisClient:    redisClient,
		cfg:            cfg,
		instanceID:     uuid.NewString(),
	}
}

func ProvideAffiliateRiskScannerService(
	repo AffiliateRiskRepository,
	opsRepo OpsRepository,
	opsService *OpsService,
	settingService *SettingService,
	emailService *EmailService,
	redisClient *redis.Client,
	cfg *config.Config,
) *AffiliateRiskScannerService {
	svc := NewAffiliateRiskScannerService(repo, opsRepo, opsService, settingService, emailService, redisClient, cfg)
	svc.Start()
	return svc
}

func (s *AffiliateRiskScannerService) Start() {
	if s == nil {
		return
	}
	s.startOnce.Do(func() {
		if s.stopCh == nil {
			s.stopCh = make(chan struct{})
		}
		s.wg.Add(1)
		go s.run()
	})
}

func (s *AffiliateRiskScannerService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
		}
	})
	s.wg.Wait()
}

func (s *AffiliateRiskScannerService) run() {
	defer s.wg.Done()
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			interval := s.getInterval()
			s.scanOnce(interval)
			timer.Reset(interval)
		case <-s.stopCh:
			return
		}
	}
}

func (s *AffiliateRiskScannerService) getInterval() time.Duration {
	minutes := AffiliateRiskScanIntervalDefaultMin
	if s != nil && s.settingService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		minutes = s.settingService.GetAffiliateRiskScanIntervalMinutes(ctx)
	}
	if minutes < AffiliateRiskScanIntervalMin || minutes > AffiliateRiskScanIntervalMax {
		minutes = AffiliateRiskScanIntervalDefaultMin
	}
	return time.Duration(minutes) * time.Minute
}

func (s *AffiliateRiskScannerService) scanOnce(interval time.Duration) {
	if s == nil || s.repo == nil || s.opsRepo == nil {
		return
	}
	if s.cfg != nil && !s.cfg.Ops.Enabled {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), affiliateRiskTimeout)
	defer cancel()

	if s.opsService != nil && !s.opsService.IsMonitoringEnabled(ctx) {
		return
	}

	release, ok := s.tryAcquireLeaderLock(ctx)
	if !ok {
		return
	}
	if release != nil {
		defer release()
	}

	startedAt := time.Now().UTC()
	runAt := startedAt
	windowEnd := startedAt
	windowStart := windowEnd.Add(-affiliateRiskWindow)

	clusters, err := s.repo.ListAffiliateRiskClusters(ctx, windowStart, windowEnd)
	if err != nil {
		s.recordHeartbeatError(runAt, time.Since(startedAt), err)
		logger.LegacyPrintf("service.affiliate_risk", "[AffiliateRiskScanner] list clusters failed: %v", err)
		return
	}

	evaluated := 0
	alerts := 0
	freezes := 0
	for _, cluster := range clusters {
		result := scoreAffiliateRiskCluster(cluster, windowStart, windowEnd)
		if result.Score < 50 {
			continue
		}
		evaluated++
		if created, err := s.createRiskAlert(ctx, cluster, result, windowStart, windowEnd); err != nil {
			logger.LegacyPrintf("service.affiliate_risk", "[AffiliateRiskScanner] create alert failed inviter=%d: %v", cluster.InviterID, err)
			continue
		} else if created {
			alerts++
		}
		if result.Score >= 70 {
			ok, err := s.repo.UpsertAffiliateRiskFreeze(ctx, AffiliateRiskFreeze{
				InviterID:         cluster.InviterID,
				Fingerprint:       result.Fingerprint,
				Severity:          result.Severity,
				Score:             result.Score,
				Reason:            result.Description,
				SourceWindowStart: windowStart,
				SourceWindowEnd:   windowEnd,
			})
			if err != nil {
				logger.LegacyPrintf("service.affiliate_risk", "[AffiliateRiskScanner] upsert freeze failed inviter=%d: %v", cluster.InviterID, err)
				continue
			}
			if ok {
				freezes++
			}
		}
	}

	result := fmt.Sprintf("clusters=%d evaluated=%d alerts=%d freezes=%d interval=%s", len(clusters), evaluated, alerts, freezes, interval)
	s.recordHeartbeatSuccess(runAt, time.Since(startedAt), result)
}

func (s *AffiliateRiskScannerService) createRiskAlert(ctx context.Context, cluster AffiliateRiskCluster, result affiliateRiskScoreResult, windowStart, windowEnd time.Time) (bool, error) {
	if s == nil || s.opsRepo == nil {
		return false, nil
	}
	active, err := s.findActiveRiskAlert(ctx, cluster.InviterID, result.Fingerprint)
	if err != nil {
		return false, err
	}
	if active != nil {
		return false, nil
	}

	score := float64(result.Score)
	threshold := 50.0
	event := &OpsAlertEvent{
		RuleID:         0,
		Severity:       result.Severity,
		Status:         OpsAlertStatusFiring,
		Title:          result.Title,
		Description:    result.Description,
		MetricValue:    &score,
		ThresholdValue: &threshold,
		Dimensions: map[string]any{
			"kind":          "affiliate_risk",
			"inviter_id":    cluster.InviterID,
			"fingerprint":   result.Fingerprint,
			"window_start":  windowStart.Format(time.RFC3339),
			"window_end":    windowEnd.Format(time.RFC3339),
			"invitee_count": len(cluster.Invitees),
		},
		FiredAt:   time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}
	created, err := s.opsRepo.CreateAlertEvent(ctx, event)
	if err != nil {
		return false, err
	}
	if created != nil && created.ID > 0 {
		s.sendRiskAlertEmailBestEffort(ctx, created)
	}
	return created != nil, nil
}

func (s *AffiliateRiskScannerService) findActiveRiskAlert(ctx context.Context, inviterID int64, fingerprint string) (*OpsAlertEvent, error) {
	return s.opsRepo.GetActiveAlertEventByDimension(ctx, "affiliate_risk", inviterID, fingerprint)
}

func (s *AffiliateRiskScannerService) sendRiskAlertEmailBestEffort(ctx context.Context, event *OpsAlertEvent) {
	if s == nil || s.emailService == nil || s.opsService == nil || s.opsRepo == nil || event == nil || event.EmailSent {
		return
	}
	cfg, err := s.opsService.GetEmailNotificationConfig(ctx)
	if err != nil || cfg == nil || !cfg.Alert.Enabled || len(cfg.Alert.Recipients) == 0 {
		return
	}
	if !shouldSendOpsAlertEmailByMinSeverity(strings.TrimSpace(cfg.Alert.MinSeverity), event.Severity) {
		return
	}
	subject := fmt.Sprintf("[Ops Alert][%s] %s", strings.TrimSpace(event.Severity), strings.TrimSpace(event.Title))
	body := fmt.Sprintf("<h2>Ops Alert</h2><p><b>Severity</b>: %s</p><p><b>Status</b>: %s</p><p><b>Fired at</b>: %s</p><p><b>Description</b>: %s</p>",
		htmlEscape(event.Severity),
		htmlEscape(event.Status),
		event.FiredAt.Format(time.RFC3339),
		htmlEscape(event.Description),
	)
	anySent := false
	for _, to := range cfg.Alert.Recipients {
		addr := strings.TrimSpace(to)
		if addr == "" {
			continue
		}
		if err := s.emailService.SendEmail(ctx, addr, subject, body); err == nil {
			anySent = true
		}
	}
	if anySent {
		_ = s.opsRepo.UpdateAlertEventEmailSent(context.Background(), event.ID, true)
	}
}

func scoreAffiliateRiskCluster(cluster AffiliateRiskCluster, windowStart, windowEnd time.Time) affiliateRiskScoreResult {
	reasons := make([]string, 0)
	score := 0
	if len(cluster.Invitees) >= 3 {
		score += 25
		reasons = append(reasons, fmt.Sprintf("12h内邀请%d个账号", len(cluster.Invitees)))
	}

	inviterLoginKey := riskip.NormalizeIPForRiskKey(cluster.InviterLastLoginIP)
	inviterLoginExact := strings.TrimSpace(cluster.InviterLastLoginIP)
	sharedLoginIP := 0
	sharedIPv64 := 0
	for _, invitee := range cluster.Invitees {
		if inviterLoginExact != "" &&
			(strings.TrimSpace(invitee.LastLoginIP) == inviterLoginExact ||
				strings.TrimSpace(invitee.FirstUsageIP) == inviterLoginExact) {
			sharedLoginIP++
		}
		if inviterLoginKey != "" &&
			(riskip.NormalizeIPForRiskKey(invitee.LastLoginIP) == inviterLoginKey ||
				riskip.NormalizeIPForRiskKey(invitee.FirstUsageIP) == inviterLoginKey) {
			sharedIPv64++
		}
	}
	if sharedLoginIP > 0 {
		score += 40
		reasons = append(reasons, fmt.Sprintf("%d个被邀请账号与邀请人登录或API使用IP相同", sharedLoginIP))
	}
	if sharedIPv64 > 0 && strings.Contains(inviterLoginKey, "/64") {
		score += 35
		reasons = append(reasons, fmt.Sprintf("%d个被邀请账号与邀请人IPv6 /64相同", sharedIPv64))
	}

	if registerDispersedLoginAggregated(cluster.Invitees) {
		score += 25
		reasons = append(reasons, "注册IP分散但登录/API使用IP或IPv6 /64聚合")
	}

	fastRewards := 0
	revokedRewarded := 0
	for _, invitee := range cluster.Invitees {
		if invitee.APICallRewardAt != nil &&
			!invitee.CreatedAt.IsZero() &&
			!invitee.APICallRewardAt.Before(invitee.CreatedAt) &&
			invitee.APICallRewardAt.Sub(invitee.CreatedAt) <= affiliateRiskRewardFastAfter {
			fastRewards++
		}
		if invitee.AffiliateRevokedAt != nil && invitee.APICallRewardAt != nil {
			revokedRewarded++
		}
	}
	if fastRewards > 0 {
		score += 20
		reasons = append(reasons, fmt.Sprintf("%d个账号注册30分钟内触发API奖励", fastRewards))
	}
	if batchGeneratedEmails(cluster.Invitees) {
		score += 10
		reasons = append(reasons, "多个被邀请账号邮箱疑似批量生成")
	}
	if revokedRewarded > 0 {
		score += 30
		reasons = append(reasons, fmt.Sprintf("%d个已撤销邀请关系仍存在API调用奖励", revokedRewarded))
	}

	severity := ""
	switch {
	case score >= 90:
		severity = "P1"
	case score >= 70:
		severity = "P2"
	case score >= 50:
		severity = "P3"
	}
	fingerprint := affiliateRiskFingerprint(cluster, reasons)
	title := buildAffiliateRiskTitle(cluster, sharedIPv64, score)
	description := buildAffiliateRiskDescription(cluster, score, severity, reasons, windowStart, windowEnd)
	return affiliateRiskScoreResult{
		Score:       score,
		Severity:    severity,
		Fingerprint: fingerprint,
		Title:       title,
		Description: description,
		Reasons:     reasons,
	}
}

func registerDispersedLoginAggregated(invitees []AffiliateRiskInvitee) bool {
	if len(invitees) < 3 {
		return false
	}
	registerKeys := map[string]struct{}{}
	loginKeys := map[string]int{}
	for _, invitee := range invitees {
		if key := riskip.NormalizeIPForRiskKey(invitee.RegisterIP); key != "" {
			registerKeys[key] = struct{}{}
		}
		inviteeLoginKeys := map[string]struct{}{}
		if key := riskip.NormalizeIPForRiskKey(invitee.LastLoginIP); key != "" {
			inviteeLoginKeys[key] = struct{}{}
		}
		if key := riskip.NormalizeIPForRiskKey(invitee.FirstUsageIP); key != "" {
			inviteeLoginKeys[key] = struct{}{}
		}
		for key := range inviteeLoginKeys {
			loginKeys[key]++
		}
	}
	if len(registerKeys) < 2 {
		return false
	}
	for _, count := range loginKeys {
		if count >= 2 {
			return true
		}
	}
	return false
}

func batchGeneratedEmails(invitees []AffiliateRiskInvitee) bool {
	if len(invitees) < 3 {
		return false
	}
	domains := map[string]int{}
	prefixGroups := map[string]int{}
	for _, invitee := range invitees {
		local, domain, ok := splitEmail(invitee.Email)
		if !ok {
			continue
		}
		domains[domain]++
		prefix := emailAlphaPrefix(local)
		if prefix != "" {
			prefixGroups[prefix]++
		}
	}
	for _, count := range domains {
		if count >= 3 {
			for _, prefixCount := range prefixGroups {
				if prefixCount >= 3 {
					return true
				}
			}
		}
	}
	return false
}

func splitEmail(email string) (local, domain string, ok bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func emailAlphaPrefix(local string) string {
	var b strings.Builder
	for _, r := range local {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
			continue
		}
		break
	}
	if b.Len() < 2 {
		return ""
	}
	return b.String()
}

func affiliateRiskFingerprint(cluster AffiliateRiskCluster, reasons []string) string {
	ids := make([]int64, 0, len(cluster.Invitees))
	for _, invitee := range cluster.Invitees {
		ids = append(ids, invitee.UserID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	reasonCopy := append([]string(nil), reasons...)
	sort.Strings(reasonCopy)
	raw := fmt.Sprintf("%d|%v|%s", cluster.InviterID, ids, strings.Join(reasonCopy, "|"))
	sum := sha1.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])[:24]
}

func buildAffiliateRiskTitle(cluster AffiliateRiskCluster, sharedIPv64 int, score int) string {
	email := strings.TrimSpace(cluster.InviterEmail)
	if email == "" {
		email = fmt.Sprintf("user#%d", cluster.InviterID)
	}
	if sharedIPv64 > 0 {
		return fmt.Sprintf("疑似刷邀请奖励：%s 12小时内关联%d个小号，%d个共享登录IPv6", email, len(cluster.Invitees), sharedIPv64)
	}
	return fmt.Sprintf("疑似刷邀请奖励：%s 12小时内关联%d个小号，风险分%d", email, len(cluster.Invitees), score)
}

func buildAffiliateRiskDescription(cluster AffiliateRiskCluster, score int, severity string, reasons []string, windowStart, windowEnd time.Time) string {
	if len(reasons) == 0 {
		reasons = []string{"风险评分达到告警阈值"}
	}
	return fmt.Sprintf("%s，severity=%s，score=%d，window=%s~%s，原因：%s",
		buildAffiliateRiskTitle(cluster, countSharedInviteeIPv64(cluster), score),
		severity,
		score,
		windowStart.Format(time.RFC3339),
		windowEnd.Format(time.RFC3339),
		strings.Join(reasons, "；"),
	)
}

func countSharedInviteeIPv64(cluster AffiliateRiskCluster) int {
	key := riskip.NormalizeIPForRiskKey(cluster.InviterLastLoginIP)
	if key == "" || !strings.Contains(key, "/64") {
		return 0
	}
	count := 0
	for _, invitee := range cluster.Invitees {
		if riskip.NormalizeIPForRiskKey(invitee.LastLoginIP) == key {
			count++
		}
	}
	return count
}

func (s *AffiliateRiskScannerService) tryAcquireLeaderLock(ctx context.Context) (func(), bool) {
	if s == nil || s.redisClient == nil {
		return nil, true
	}
	ok, err := s.redisClient.SetNX(ctx, affiliateRiskLeaderLockKey, s.instanceID, affiliateRiskLeaderLockTTL).Result()
	if err != nil {
		logger.LegacyPrintf("service.affiliate_risk", "[AffiliateRiskScanner] leader lock SetNX failed; skipping this cycle: %v", err)
		return nil, false
	}
	if !ok {
		s.maybeLogSkip()
		return nil, false
	}
	return func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer releaseCancel()
		_, _ = affiliateRiskReleaseScript.Run(releaseCtx, s.redisClient, []string{affiliateRiskLeaderLockKey}, s.instanceID).Result()
	}, true
}

func (s *AffiliateRiskScannerService) maybeLogSkip() {
	s.skipLogMu.Lock()
	defer s.skipLogMu.Unlock()
	now := time.Now()
	if !s.skipLogAt.IsZero() && now.Sub(s.skipLogAt) < time.Minute {
		return
	}
	s.skipLogAt = now
	logger.LegacyPrintf("service.affiliate_risk", "[AffiliateRiskScanner] leader lock held by another instance; skipping")
}

func (s *AffiliateRiskScannerService) recordHeartbeatSuccess(runAt time.Time, duration time.Duration, result string) {
	if s == nil || s.opsRepo == nil {
		return
	}
	now := time.Now().UTC()
	durMs := duration.Milliseconds()
	msg := truncateString(result, 2048)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.opsRepo.UpsertJobHeartbeat(ctx, &OpsUpsertJobHeartbeatInput{
		JobName:        affiliateRiskScannerJobName,
		LastRunAt:      &runAt,
		LastSuccessAt:  &now,
		LastDurationMs: &durMs,
		LastResult:     &msg,
	})
}

func (s *AffiliateRiskScannerService) recordHeartbeatError(runAt time.Time, duration time.Duration, err error) {
	if s == nil || s.opsRepo == nil || err == nil {
		return
	}
	now := time.Now().UTC()
	durMs := duration.Milliseconds()
	msg := truncateString(err.Error(), 2048)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.opsRepo.UpsertJobHeartbeat(ctx, &OpsUpsertJobHeartbeatInput{
		JobName:        affiliateRiskScannerJobName,
		LastRunAt:      &runAt,
		LastErrorAt:    &now,
		LastError:      &msg,
		LastDurationMs: &durMs,
	})
}
