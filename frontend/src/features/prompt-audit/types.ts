export type PromptAuditMode = 'off' | 'async_audit' | 'blocking'
export type PromptDecision = 'pass' | 'flag' | 'critical'
export type PromptRiskLevel = 'low' | 'medium' | 'high' | 'critical'
export type PromptPolicyAction = 'Allow' | 'Warn' | 'Block'

export interface PromptRiskPolicyAction {
  action: PromptPolicyAction | string
  risk_level?: PromptRiskLevel | string
}

export interface PromptRiskPolicyRule {
  id: string
  priority: number
  safety?: string[]
  categories?: string[]
  groups?: number[]
  models?: string[]
  providers?: string[]
  action: PromptPolicyAction | string
  risk_level?: PromptRiskLevel | string
  message_code?: string
  owasp_tags?: string[]
}

export interface PromptRiskActionRules {
  policy_id?: string
  policy_version?: number
  defaults?: Record<string, PromptRiskPolicyAction>
  rules?: PromptRiskPolicyRule[]
  safety?: Record<string, PromptPolicyAction | string>
  categories?: Record<string, PromptPolicyAction | string>
}

export interface PromptAuditEndpoint {
  id: string
  name: string
  protocol: 'openai_compatible'
  base_url: string
  model: string
  timeout_ms: number
  input_limit: number
  enabled: boolean
  has_token: boolean
  token_status: 'configured' | 'missing' | 'invalid' | string
}

export interface PromptAuditEndpointDraft extends PromptAuditEndpoint {
  token: string
  clear_token: boolean
}

export interface PromptAuditConfig {
  enabled: boolean
  blocking_enabled: boolean
  blocking_latest_turn_only: boolean
  store_pass_events: boolean
  effective_mode: PromptAuditMode
  strategy: 'priority'
  worker_count: number
  queue_capacity: number
  scanners: string[]
  rules?: PromptRiskActionRules
  all_groups: boolean
  group_ids: number[]
  endpoints: PromptAuditEndpoint[]
  config_version: number
  updated_at: string
  updated_by: number
  change_summary: string
}

export interface PromptAuditDraft extends Omit<PromptAuditConfig, 'endpoints'> {
  endpoints: PromptAuditEndpointDraft[]
}

export interface PromptAuditUpdateRequest {
  expected_config_version: number
  enabled: boolean
  blocking_enabled: boolean
  blocking_latest_turn_only: boolean
  store_pass_events: boolean
  strategy: 'priority'
  worker_count: number
  queue_capacity: number
  scanners: string[]
  rules?: PromptRiskActionRules
  all_groups: boolean
  group_ids: number[]
  endpoints: Array<{
    id: string
    name: string
    protocol: 'openai_compatible'
    base_url: string
    model: string
    token?: string
    clear_token: boolean
    timeout_ms: number
    input_limit: number
    enabled: boolean
  }>
}

export interface PromptPolicyDraft {
  draft_version: number
  base_config_version: number
  rules: PromptRiskActionRules
  updated_at: string
  updated_by: number
}

export interface PromptPolicyVersion {
  policy_version: number
  matched_rule_id?: string
  owasp_tags?: string[]
  policy_id: string
  rules: PromptRiskActionRules
  config_version: number
  created_at: string
  created_by: number
  rollback_count?: number
}

export interface PromptPolicyHistory {
  active_version: number
  draft?: PromptPolicyDraft
  versions: PromptPolicyVersion[]
}

export interface PromptPolicyPreviewExample {
  name: string
  safety: string
  categories: string[]
  current_action: PromptPolicyAction
  current_risk_level: PromptRiskLevel
  candidate_action: PromptPolicyAction
  candidate_risk_level: PromptRiskLevel
  matched_rule_id?: string
  owasp_tags?: string[]
  would_escalate: boolean
}

export interface PromptPolicyPreview {
  policy_id: string
  rule_count: number
  category_count: number
  scoped_rule_count: number
  blocking_rule_count: number
  warning_rule_count: number
  affected_scopes: string[]
  examples?: PromptPolicyPreviewExample[]
}

export type PromptPolicyShadowSample = 'safe' | 'controversial' | 'violent' | 'unsafe'

export interface PromptPolicyMatchContext {
  group_id?: number
  model?: string
  provider?: string
}

export interface PromptPolicyShadowResult {
  current: Record<string, unknown>
  candidate: Record<string, unknown>
  action_changed: boolean
  risk_changed: boolean
  would_escalate: boolean
}

export interface PromptProbeResult {
  ok: boolean
  status: string
  error_code?: string
  message: string
  latency_ms: number
  http_status: number
  retryable: boolean
  checked_at: string
  token_applied: boolean
}

export interface PromptQueueStats {
  staging: number
  queued: number
  processing: number
  retry: number
  done: number
  failed: number
  active: number
}

export interface PromptGuardMetrics {
  total: number
  allowed: number
  flagged: number
  blocked: number
  unavailable: number
  invalid: number
  timeouts: number
  failovers: number
  bulkhead_full: number
  record_failed: number
  latency_avg_ms?: number
  latency_p50_ms?: number
  latency_p95_ms?: number
  latency_p99_ms?: number
  latency_max_ms?: number
}

export interface PromptAuditRuntime {
  process_status: 'disabled' | 'running' | 'degraded' | 'error' | string
  effective_mode: PromptAuditMode
  expected_config_version: number
  active_config_version: number
  config_loaded_at?: string
  config_load_error?: string
  worker_total: number
  worker_active: number
  worker_heartbeat_at?: string
  queue_capacity: number
  queue: PromptQueueStats
  processed_total: number
  failed_total: number
  enqueued_total: number
  dropped_total: number
  last_processed_at?: string
  last_error_code?: string
  last_error_message?: string
  database_status: string
  redis_status: string
  endpoints: Record<string, PromptProbeResult>
  guard_metrics: PromptGuardMetrics
}

export interface PromptSnapshot {
  request_id: string
  user_id: number
  username: string
  user_email: string
  api_key_id: number
  api_key_name: string
  group_id?: number
  group_name: string
  provider: string
  endpoint: string
  protocol: string
  model: string
  prompt_hash: string
  redacted_preview: string
  full_prompt: string
  prompt_length: number
  message_count: number
  stage: string
}

export interface PromptIssueSummary {
  category: string
  scanner_id: string
  title: string
  description: string
  severity: string
  severity_label: string
  action: string
  action_label: string
  code: string
  score: number
  evidence: string
  evidence_hash: string
  start_rune?: number
  end_rune?: number
}

export interface PromptAuditEvent {
  id: number
  job_id: number
  snapshot: PromptSnapshot
  decision: PromptDecision
  risk_level: PromptRiskLevel
  action: 'Allow' | 'Warn' | 'Block' | string
  categories: string[]
  matched_scanners: string[]
  scanner_scores: Record<string, number>
  scanner_evidence: Record<string, string>
  scanner_backend: string
  scanner_version: string
  guard_endpoint_id: string
  policy_id: string
  policy_version: number
  matched_rule_id?: string
  owasp_tags?: string[]
  config_version: number
  chunk_total: number
  latency_ms: number
  issue_summaries: PromptIssueSummary[]
  created_at: string
}

export interface PromptEventFilters {
  decision: string
  risk_level: string
  endpoint: string
  group_id: string
  user_id: string
  api_key_id: string
  request_id: string
  prompt_hash: string
  keyword: string
  start_at: string
  end_at: string
}

export interface PromptEventPage {
  items: PromptAuditEvent[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface PromptDeleteResult {
  deleted_events: number
  deleted_jobs: number
}

export interface PromptDeletePreview {
  matched_count: number
  filter_summary: Record<string, unknown>
  snapshot_max_id: number
  filter_hash: string
  confirmation_token: string
  expires_at: string
}

export interface PromptAuditGroup {
  id: number
  name: string
  status: 'active' | 'inactive'
  platform: string
}

export interface PromptLoadErrors {
  config: string
  runtime: string
  groups: string
  events: string
}
