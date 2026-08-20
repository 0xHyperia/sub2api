package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type modelMonitorRepository struct{ db *sql.DB }

func NewModelMonitorRepository(db *sql.DB) service.ModelMonitorRepository {
	return &modelMonitorRepository{db: db}
}

const modelMonitorColumns = `id, platform, model, enabled, interval_seconds, display_order, label, last_checked_at, created_by, created_at, updated_at`

// These predicates intentionally require positive evidence. Model Monitor is
// an upstream model-health signal, not a generic gateway error counter: client,
// account, routing and transport failures must not lower its success rate.
const modelMonitorSuccessfulUsagePredicate = `COALESCE(ul.request_type,0) NOT IN (4,6)
    AND (COALESCE(ul.actual_cost,0)>0 OR COALESCE(ul.total_cost,0)>0
      OR COALESCE(ul.input_tokens,0)>0 OR COALESCE(ul.output_tokens,0)>0
      OR COALESCE(ul.cache_creation_tokens,0)>0 OR COALESCE(ul.cache_read_tokens,0)>0
      OR COALESCE(ul.image_count,0)>0 OR COALESCE(ul.video_count,0)>0)`

const modelMonitorConfirmedFailurePredicate = `(
    (
      COALESCE(e.upstream_status_code,e.status_code,0) IN (500,503,529)
      AND (LOWER(COALESCE(e.error_owner,''))='provider'
        OR e.upstream_status_code IS NOT NULL
        OR (jsonb_typeof(e.upstream_errors)='array' AND jsonb_array_length(e.upstream_errors)>0))
    )
    OR (
      (LOWER(COALESCE(e.error_owner,''))='provider'
        OR e.upstream_status_code IS NOT NULL
        OR (jsonb_typeof(e.upstream_errors)='array' AND jsonb_array_length(e.upstream_errors)>0))
      AND LOWER(CONCAT_WS(' ',e.error_message,e.error_body,e.upstream_error_message,
                          e.upstream_error_detail,e.provider_error_code,e.provider_error_type)) LIKE ANY(ARRAY[
        '%selected model is at capacity%','%model_capacity_exhausted%',
        '%no capacity available for model%','%server_is_overloaded%',
        '%servers are currently overloaded%','%server is overloaded%',
        '%upstream service overloaded%','%engine_overloaded%','%overloaded_error%'])
    )
  ) AND NOT (
    LOWER(CONCAT_WS(' ',e.error_message,e.error_body,e.upstream_error_message,
                    e.upstream_error_detail,e.provider_error_code,e.provider_error_type)) LIKE ANY(ARRAY[
      '%model not found%','%model_not_found%','%unsupported model%','%invalid model%',
      '%invalid request%','%invalid_request%','%context length%','%context_length%',
      '%maximum context%','%insufficient balance%','%insufficient quota%','%quota exceeded%',
      '%unauthorized%','%forbidden%','%authentication%','%invalid api key%',
      '%content policy%','%content_policy%','%safety policy%',
      '%timeout%','%timed out%','%deadline exceeded%','%tls handshake%',
      '%network error%','%transport error%','%connection%','%dial tcp%','%read tcp%','%write tcp%',
      '%no such host%','%name resolution%','%x509%','%certificate%','%broken pipe%','%eof%','%stream error%',
      '%client disconnected%','%cancelled%','%canceled%'])
  )`

func scanModelMonitor(scanner interface{ Scan(...any) error }) (*service.ModelMonitor, error) {
	var m service.ModelMonitor
	if err := scanner.Scan(&m.ID, &m.Platform, &m.Model, &m.Enabled, &m.IntervalSeconds, &m.DisplayOrder, &m.Label, &m.LastCheckedAt, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *modelMonitorRepository) List(ctx context.Context) ([]service.ModelMonitor, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+modelMonitorColumns+` FROM model_monitors ORDER BY platform, model`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.ModelMonitor, 0)
	for rows.Next() {
		m, scanErr := scanModelMonitor(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *modelMonitorRepository) ListEnabledDue(ctx context.Context, now time.Time) ([]service.ModelMonitor, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+modelMonitorColumns+` FROM model_monitors WHERE enabled = TRUE AND (last_checked_at IS NULL OR last_checked_at + interval '1 second' * interval_seconds <= $1) ORDER BY COALESCE(last_checked_at, to_timestamp(0)) LIMIT 100`, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.ModelMonitor, 0)
	for rows.Next() {
		m, scanErr := scanModelMonitor(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *modelMonitorRepository) ListEnabledDueGroups(ctx context.Context, now time.Time, limit int) ([]service.ModelMonitorDueGroup, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT m.id,m.platform,m.model,m.enabled,m.interval_seconds,m.display_order,m.label,m.last_checked_at,m.created_by,m.created_at,m.updated_at,
       mg.group_id,mg.priority,mg.enabled,mg.interval_seconds,
       mg.failure_compensation_enabled,mg.failure_compensation_pending,
       mg.last_traffic_at,mg.last_probe_at,mg.last_scheduled_slot_at,
       mg.next_compensation_at,mg.consecutive_probe_failures,g.name
FROM model_monitor_groups mg
JOIN model_monitors m ON m.id=mg.monitor_id
JOIN groups g ON g.id=mg.group_id AND g.deleted_at IS NULL
WHERE mg.enabled=TRUE
  AND (mg.probe_claimed_until IS NULL OR mg.probe_claimed_until <= $1)
  AND (
    (mg.failure_compensation_enabled=TRUE AND mg.failure_compensation_pending=TRUE AND mg.next_compensation_at <= $1)
    OR (
      mg.failure_compensation_pending=FALSE
      AND to_timestamp(floor(extract(epoch FROM $1::timestamptz) / mg.interval_seconds) * mg.interval_seconds)
          > COALESCE(mg.last_scheduled_slot_at,to_timestamp(0))
    )
  )
ORDER BY ROW_NUMBER() OVER (
           PARTITION BY mg.failure_compensation_pending
           ORDER BY COALESCE(mg.next_compensation_at,mg.last_scheduled_slot_at,to_timestamp(0)),m.id,mg.priority
         ),
         CASE WHEN mg.failure_compensation_pending THEN 0 ELSE 1 END
LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.ModelMonitorDueGroup, 0)
	for rows.Next() {
		var item service.ModelMonitorDueGroup
		m, scanErr := scanModelMonitorWithTail(rows,
			&item.Group.GroupID, &item.Group.Priority, &item.Group.Enabled, &item.Group.IntervalSeconds,
			&item.Group.FailureCompensationEnabled, &item.Group.FailureCompensationPending,
			&item.Group.LastTrafficAt, &item.Group.LastProbeAt, &item.Group.LastScheduledSlotAt,
			&item.Group.NextCompensationAt, &item.Group.ConsecutiveProbeFailures, &item.Name)
		if scanErr != nil {
			return nil, scanErr
		}
		item.Monitor = *m
		item.Group.MonitorID = m.ID
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanModelMonitorWithTail(scanner interface{ Scan(...any) error }, tail ...any) (*service.ModelMonitor, error) {
	var m service.ModelMonitor
	args := []any{&m.ID, &m.Platform, &m.Model, &m.Enabled, &m.IntervalSeconds, &m.DisplayOrder, &m.Label, &m.LastCheckedAt, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt}
	args = append(args, tail...)
	if err := scanner.Scan(args...); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *modelMonitorRepository) ClaimDueGroup(ctx context.Context, monitorID, groupID int64, now, claimedUntil, scheduledSlot time.Time, compensation bool) (bool, error) {
	query := `
UPDATE model_monitor_groups mg
SET probe_claimed_until=$4
WHERE mg.monitor_id=$1 AND mg.group_id=$2 AND mg.enabled=TRUE
  AND (mg.probe_claimed_until IS NULL OR mg.probe_claimed_until <= $3)
  AND mg.failure_compensation_pending=FALSE
  AND $5 > COALESCE(mg.last_scheduled_slot_at,to_timestamp(0))`
	if compensation {
		query = `
UPDATE model_monitor_groups mg
SET probe_claimed_until=$4
WHERE mg.monitor_id=$1 AND mg.group_id=$2 AND mg.enabled=TRUE
  AND mg.failure_compensation_enabled=TRUE AND mg.failure_compensation_pending=TRUE
  AND mg.next_compensation_at <= $3
  AND (mg.probe_claimed_until IS NULL OR mg.probe_claimed_until <= $3)
  AND $5 >= date_trunc('minute',mg.next_compensation_at)`
	}
	result, err := r.db.ExecContext(ctx, query, monitorID, groupID, now, claimedUntil, scheduledSlot)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (r *modelMonitorRepository) GetByID(ctx context.Context, id int64) (*service.ModelMonitor, error) {
	m, err := scanModelMonitor(r.db.QueryRowContext(ctx, `SELECT `+modelMonitorColumns+` FROM model_monitors WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *modelMonitorRepository) GetByKey(ctx context.Context, platform, model string) (*service.ModelMonitor, error) {
	m, err := scanModelMonitor(r.db.QueryRowContext(ctx, `SELECT `+modelMonitorColumns+` FROM model_monitors WHERE platform=$1 AND model=$2`, platform, model))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *modelMonitorRepository) Upsert(ctx context.Context, m *service.ModelMonitor) error {
	if m.IntervalSeconds == 0 {
		m.IntervalSeconds = service.ModelMonitorDefaultIntervalSeconds
	}
	row := r.db.QueryRowContext(ctx, `INSERT INTO model_monitors (platform, model, enabled, interval_seconds, display_order, label, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (platform, model) DO UPDATE SET enabled=EXCLUDED.enabled, interval_seconds=EXCLUDED.interval_seconds, display_order=EXCLUDED.display_order, label=EXCLUDED.label, updated_at=NOW() RETURNING `+modelMonitorColumns, m.Platform, m.Model, m.Enabled, m.IntervalSeconds, m.DisplayOrder, m.Label, m.CreatedBy)
	updated, err := scanModelMonitor(row)
	if err != nil {
		return err
	}
	*m = *updated
	return nil
}

func (r *modelMonitorRepository) ListGroupConfigs(ctx context.Context) (map[int64][]service.ModelMonitorGroupConfig, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT monitor_id,group_id,priority,enabled,interval_seconds,
failure_compensation_enabled,failure_compensation_pending,last_traffic_at,last_probe_at,last_scheduled_slot_at,
next_compensation_at,consecutive_probe_failures FROM model_monitor_groups ORDER BY monitor_id,priority`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int64][]service.ModelMonitorGroupConfig)
	for rows.Next() {
		var config service.ModelMonitorGroupConfig
		if err := rows.Scan(&config.MonitorID, &config.GroupID, &config.Priority, &config.Enabled, &config.IntervalSeconds,
			&config.FailureCompensationEnabled, &config.FailureCompensationPending, &config.LastTrafficAt, &config.LastProbeAt,
			&config.LastScheduledSlotAt, &config.NextCompensationAt, &config.ConsecutiveProbeFailures); err != nil {
			return nil, err
		}
		out[config.MonitorID] = append(out[config.MonitorID], config)
	}
	return out, rows.Err()
}

func (r *modelMonitorRepository) ReplaceGroups(ctx context.Context, monitorID int64, groupIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `DELETE FROM model_monitor_groups WHERE monitor_id=$1`, monitorID); err != nil {
		return err
	}
	for priority, groupID := range groupIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO model_monitor_groups (monitor_id,group_id,priority,enabled,interval_seconds) VALUES ($1,$2,$3,FALSE,$4)`, monitorID, groupID, priority, service.ModelMonitorDefaultIntervalSeconds); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *modelMonitorRepository) UpsertGroupConfig(ctx context.Context, config service.ModelMonitorGroupConfig) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO model_monitor_groups (monitor_id,group_id,priority,enabled,interval_seconds,failure_compensation_enabled,last_scheduled_slot_at)
VALUES ($1,$2,COALESCE((SELECT MAX(priority)+1 FROM model_monitor_groups WHERE monitor_id=$1),0),$3,$4::integer,$5,
        to_timestamp(floor(extract(epoch FROM NOW()) / (($4::integer)::double precision)) * (($4::integer)::double precision)))
ON CONFLICT (monitor_id,group_id) DO UPDATE
SET enabled=EXCLUDED.enabled,
    interval_seconds=EXCLUDED.interval_seconds,
    failure_compensation_enabled=EXCLUDED.failure_compensation_enabled,
    last_scheduled_slot_at=CASE
      WHEN model_monitor_groups.enabled IS DISTINCT FROM EXCLUDED.enabled
        OR model_monitor_groups.interval_seconds IS DISTINCT FROM EXCLUDED.interval_seconds
      THEN EXCLUDED.last_scheduled_slot_at
      ELSE model_monitor_groups.last_scheduled_slot_at
    END,
    failure_compensation_pending=CASE
      WHEN EXCLUDED.enabled=FALSE OR EXCLUDED.failure_compensation_enabled=FALSE THEN FALSE
      ELSE model_monitor_groups.failure_compensation_pending
    END,
    next_compensation_at=CASE
      WHEN EXCLUDED.enabled=FALSE OR EXCLUDED.failure_compensation_enabled=FALSE THEN NULL
      ELSE model_monitor_groups.next_compensation_at
    END,
    consecutive_probe_failures=CASE
      WHEN EXCLUDED.enabled=FALSE OR EXCLUDED.failure_compensation_enabled=FALSE THEN 0
      ELSE model_monitor_groups.consecutive_probe_failures
    END,
    probe_claimed_until=CASE WHEN EXCLUDED.enabled=FALSE THEN NULL ELSE model_monitor_groups.probe_claimed_until END`, config.MonitorID, config.GroupID, config.Enabled, config.IntervalSeconds, config.FailureCompensationEnabled)
	return err
}

func (r *modelMonitorRepository) ReleaseGroupProbeClaim(ctx context.Context, monitorID, groupID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitor_groups SET probe_claimed_until=NULL WHERE monitor_id=$1 AND group_id=$2`, monitorID, groupID)
	return err
}

// PersistGroupProbeResult makes the history, metrics, model timestamp and
// scheduling state one durable state transition. A failed transition is safe
// to retry because the history row is only committed with the metric update.
func (r *modelMonitorRepository) PersistGroupProbeResult(ctx context.Context, monitor *service.ModelMonitor, h *service.ModelMonitorHistory, scheduledSlot *time.Time) error {
	if h == nil || h.GroupID == nil || monitor == nil {
		return fmt.Errorf("monitor group result is incomplete")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = tx.QueryRowContext(ctx, `INSERT INTO model_monitor_histories
	 (monitor_id,status,latency_ms,attempts,message,group_id,group_name,first_token_ms,input_tokens,output_tokens,generation_ms,probe_cost,checked_at,scheduled_slot_at)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	 ON CONFLICT (monitor_id,group_id,scheduled_slot_at) WHERE scheduled_slot_at IS NOT NULL AND group_id IS NOT NULL DO NOTHING
	 RETURNING id`,
		h.MonitorID, h.Status, h.LatencyMs, h.Attempts, h.Message, h.GroupID, h.GroupName,
		h.FirstTokenMs, h.InputTokens, h.OutputTokens, h.GenerationMs, h.ProbeCost, h.CheckedAt, scheduledSlot).Scan(&h.ID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if scheduledSlot == nil {
			return fmt.Errorf("manual probe history insert returned no row")
		}
		if err = tx.QueryRowContext(ctx, `SELECT id,status,checked_at FROM model_monitor_histories
		 WHERE monitor_id=$1 AND group_id=$2 AND scheduled_slot_at=$3`, monitor.ID, *h.GroupID, *scheduledSlot).
			Scan(&h.ID, &h.Status, &h.CheckedAt); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE model_monitor_groups SET probe_claimed_until=NULL WHERE monitor_id=$1 AND group_id=$2`, monitor.ID, *h.GroupID); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		monitor.LastCheckedAt = &h.CheckedAt
		return nil
	}
	cost := 0.0
	if h.ProbeCost != nil {
		cost = *h.ProbeCost
	}
	requestCount, successCount := int64(1), int64(0)
	var latencySum, latencyCount, ttftSum, ttftCount, outputTokens, generationMs int64
	if h.Status == service.MonitorStatusOperational || h.Status == service.MonitorStatusDegraded {
		successCount = 1
		if h.LatencyMs != nil {
			latencySum, latencyCount = int64(*h.LatencyMs), 1
		}
		if h.FirstTokenMs != nil {
			ttftSum, ttftCount = int64(*h.FirstTokenMs), 1
		}
		outputTokens, generationMs = int64(h.OutputTokens), h.GenerationMs
	}
	if h.Status != service.MonitorStatusError {
		if _, err = tx.ExecContext(ctx, `INSERT INTO model_monitor_metric_buckets
	 (monitor_id,group_id,resolution,source,bucket_start,request_count,success_count,latency_sum_ms,latency_count,ttft_sum_ms,ttft_count,output_tokens,generation_ms,probe_cost,probe_cost_known)
	 VALUES ($1,$2,'minute','probe',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	 ON CONFLICT (monitor_id,group_id,resolution,source,bucket_start) DO UPDATE SET
	 request_count=model_monitor_metric_buckets.request_count+EXCLUDED.request_count,
	 success_count=model_monitor_metric_buckets.success_count+EXCLUDED.success_count,
	 latency_sum_ms=model_monitor_metric_buckets.latency_sum_ms+EXCLUDED.latency_sum_ms,
	 latency_count=model_monitor_metric_buckets.latency_count+EXCLUDED.latency_count,
	 ttft_sum_ms=model_monitor_metric_buckets.ttft_sum_ms+EXCLUDED.ttft_sum_ms,
	 ttft_count=model_monitor_metric_buckets.ttft_count+EXCLUDED.ttft_count,
	 output_tokens=model_monitor_metric_buckets.output_tokens+EXCLUDED.output_tokens,
	 generation_ms=model_monitor_metric_buckets.generation_ms+EXCLUDED.generation_ms,
	 probe_cost=model_monitor_metric_buckets.probe_cost+EXCLUDED.probe_cost,
	 probe_cost_known=model_monitor_metric_buckets.probe_cost_known AND EXCLUDED.probe_cost_known,
	 updated_at=NOW()`, h.MonitorID, *h.GroupID, h.CheckedAt.UTC().Truncate(time.Minute), requestCount, successCount,
			latencySum, latencyCount, ttftSum, ttftCount, outputTokens, generationMs, cost, h.ProbeCostKnown); err != nil {
			return err
		}
	}
	success := h.Status == service.MonitorStatusOperational || h.Status == service.MonitorStatusDegraded
	retryable := h.Status == service.MonitorStatusFailed
	if _, err = tx.ExecContext(ctx, `UPDATE model_monitor_groups
	 SET last_probe_at=GREATEST(COALESCE(last_probe_at,$4),$4),
	 last_scheduled_slot_at=CASE WHEN $3::timestamptz IS NULL THEN last_scheduled_slot_at ELSE GREATEST(COALESCE(last_scheduled_slot_at,$3),$3) END,
	 failure_compensation_pending=CASE WHEN $5 THEN FALSE WHEN enabled AND failure_compensation_enabled AND $6 THEN TRUE ELSE FALSE END,
	 next_compensation_at=CASE WHEN NOT $5 AND enabled AND failure_compensation_enabled AND $6 THEN date_trunc('minute',$4::timestamptz)+interval '1 minute' ELSE NULL END,
	 consecutive_probe_failures=CASE WHEN $5 THEN 0 WHEN enabled AND failure_compensation_enabled AND $6 THEN consecutive_probe_failures+1 ELSE 0 END,
	 probe_claimed_until=CASE WHEN $3::timestamptz IS NULL THEN probe_claimed_until ELSE NULL END
	 WHERE monitor_id=$1 AND group_id=$2`, monitor.ID, *h.GroupID, scheduledSlot, h.CheckedAt, success, retryable); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE model_monitors SET last_checked_at=GREATEST(COALESCE(last_checked_at,$2),$2),updated_at=NOW() WHERE id=$1`, monitor.ID, h.CheckedAt); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	monitor.LastCheckedAt = &h.CheckedAt
	return nil
}

func (r *modelMonitorRepository) CompleteGroupSlotWithoutProbe(ctx context.Context, monitorID, groupID int64, scheduledSlot time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitor_groups
SET last_scheduled_slot_at=GREATEST(COALESCE(last_scheduled_slot_at,$3),$3),probe_claimed_until=NULL
WHERE monitor_id=$1 AND group_id=$2`, monitorID, groupID, scheduledSlot)
	return err
}

func (r *modelMonitorRepository) UpdateLastChecked(ctx context.Context, id int64, checkedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitors SET last_checked_at=$2, updated_at=NOW() WHERE id=$1`, id, checkedAt)
	return err
}

func (r *modelMonitorRepository) InsertHistory(ctx context.Context, h *service.ModelMonitorHistory) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO model_monitor_histories (monitor_id,status,latency_ms,attempts,message,group_id,group_name,first_token_ms,input_tokens,output_tokens,generation_ms,probe_cost,checked_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`, h.MonitorID, h.Status, h.LatencyMs, h.Attempts, h.Message, h.GroupID, h.GroupName, h.FirstTokenMs, h.InputTokens, h.OutputTokens, h.GenerationMs, h.ProbeCost, h.CheckedAt).Scan(&h.ID)
}

func (r *modelMonitorRepository) ListHistory(ctx context.Context, monitorID int64, limit int) ([]service.ModelMonitorHistory, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,monitor_id,status,latency_ms,attempts,message,group_id,group_name,first_token_ms,input_tokens,output_tokens,generation_ms,probe_cost,checked_at FROM model_monitor_histories WHERE monitor_id=$1 ORDER BY checked_at DESC LIMIT $2`, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.ModelMonitorHistory, 0)
	for rows.Next() {
		var h service.ModelMonitorHistory
		if err := rows.Scan(&h.ID, &h.MonitorID, &h.Status, &h.LatencyMs, &h.Attempts, &h.Message, &h.GroupID, &h.GroupName, &h.FirstTokenMs, &h.InputTokens, &h.OutputTokens, &h.GenerationMs, &h.ProbeCost, &h.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *modelMonitorRepository) Summaries(ctx context.Context, keys []service.ModelCatalogEntry, timelineLimit int) (map[string]service.ModelMonitorSummary, error) {
	out := make(map[string]service.ModelMonitorSummary, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	if timelineLimit <= 0 || timelineLimit > 100 {
		timelineLimit = service.ModelMonitorTimelinePoints
	}
	platforms := make([]string, 0, len(keys))
	models := make([]string, 0, len(keys))
	for _, key := range keys {
		platforms = append(platforms, key.Platform)
		models = append(models, key.Model)
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT m.id,m.platform,m.model,m.enabled,m.display_order,m.label,h.status,h.latency_ms,h.checked_at,
		CASE WHEN COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days' AND h7.status IN ('operational','degraded','failed'))=0 THEN NULL
		     ELSE 100.0*COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days' AND h7.status IN ('operational','degraded'))/COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days' AND h7.status IN ('operational','degraded','failed')) END
FROM model_monitors m
LEFT JOIN LATERAL (SELECT status,latency_ms,checked_at FROM model_monitor_histories WHERE monitor_id=m.id AND status IN ('operational','degraded','failed') ORDER BY checked_at DESC LIMIT 1) h ON TRUE
LEFT JOIN model_monitor_histories h7 ON h7.monitor_id=m.id
WHERE (m.platform,m.model) IN (SELECT * FROM unnest($1::text[],$2::text[]))
GROUP BY m.id,m.platform,m.model,m.enabled,m.display_order,m.label,h.status,h.latency_ms,h.checked_at`, pq.Array(platforms), pq.Array(models))
	if err != nil {
		return nil, err
	}
	monitorKeys := make(map[int64]string)
	monitorIDs := make([]int64, 0)
	for rows.Next() {
		var monitorID int64
		var platform, model string
		var enabled bool
		var summary service.ModelMonitorSummary
		var status sql.NullString
		var latency sql.NullInt64
		var checkedAt sql.NullTime
		var availability sql.NullFloat64
		if err := rows.Scan(&monitorID, &platform, &model, &enabled, &summary.DisplayOrder, &summary.Label, &status, &latency, &checkedAt, &availability); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if status.Valid {
			summary.Status = status.String
		}
		if latency.Valid {
			value := int(latency.Int64)
			summary.LatencyMs = &value
		}
		if checkedAt.Valid {
			value := checkedAt.Time
			summary.LastCheckedAt = &value
		}
		if availability.Valid {
			value := availability.Float64
			summary.Availability7d = &value
		}
		key := service.ModelMonitorKey(platform, model)
		out[key] = summary
		monitorKeys[monitorID] = key
		monitorIDs = append(monitorIDs, monitorID)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(monitorIDs) == 0 {
		return out, nil
	}
	timelineRows, err := r.db.QueryContext(ctx, `SELECT monitor_id,status,latency_ms,checked_at,group_id,group_name FROM (
SELECT monitor_id,status,latency_ms,checked_at,group_id,group_name,ROW_NUMBER() OVER (PARTITION BY monitor_id ORDER BY checked_at DESC) AS row_num
FROM model_monitor_histories WHERE monitor_id = ANY($1) AND status IN ('operational','degraded','failed')) ranked WHERE row_num <= $2 ORDER BY monitor_id,checked_at DESC`, pq.Array(monitorIDs), timelineLimit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = timelineRows.Close() }()
	for timelineRows.Next() {
		var monitorID int64
		var point service.ModelMonitorTimelinePoint
		if err := timelineRows.Scan(&monitorID, &point.Status, &point.LatencyMs, &point.CheckedAt, &point.GroupID, &point.GroupName); err != nil {
			return nil, err
		}
		key, ok := monitorKeys[monitorID]
		if !ok {
			continue
		}
		summary := out[key]
		summary.Timeline = append(summary.Timeline, point)
		out[key] = summary
	}
	return out, timelineRows.Err()
}

func (r *modelMonitorRepository) RefreshTrafficMetrics(ctx context.Context, from, to time.Time) error {
	from = from.UTC().Truncate(time.Minute)
	to = to.UTC().Truncate(time.Minute).Add(time.Minute)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `DELETE FROM model_monitor_metric_buckets WHERE resolution='minute' AND source='traffic' AND bucket_start >= $1 AND bucket_start < $2`, from, to); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
WITH successful_candidates AS (
  SELECT DISTINCT ON (m.id,ul.group_id,COALESCE(ul.api_key_id::text,'')||':'||COALESCE(NULLIF(ul.request_id,''),'usage:'||ul.id::text))
         m.id AS monitor_id,ul.group_id,date_trunc('minute',ul.created_at) AS bucket_start,
         ul.created_at,ul.duration_ms,ul.first_token_ms,ul.output_tokens,ul.api_key_id,ul.request_id,
         COALESCE(ul.api_key_id::text,'')||':'||COALESCE(NULLIF(ul.request_id,''),'usage:'||ul.id::text) AS request_key
  FROM usage_logs ul
  JOIN accounts a ON a.id=ul.account_id
  JOIN model_monitors m ON m.platform=a.platform AND m.model=COALESCE(NULLIF(ul.requested_model,''),ul.model)
  JOIN model_monitor_groups mg ON mg.monitor_id=m.id AND mg.group_id=ul.group_id
  WHERE ul.created_at >= $1 AND ul.created_at < $2 AND ul.group_id IS NOT NULL
    AND `+modelMonitorSuccessfulUsagePredicate+`
  ORDER BY m.id,ul.group_id,COALESCE(ul.api_key_id::text,'')||':'||COALESCE(NULLIF(ul.request_id,''),'usage:'||ul.id::text),ul.created_at DESC,ul.id DESC
), successful AS (
  SELECT monitor_id,group_id,bucket_start,COUNT(*)::bigint AS request_count,COUNT(*)::bigint AS success_count,
         COALESCE(SUM(duration_ms),0)::bigint AS latency_sum_ms,COUNT(duration_ms)::bigint AS latency_count,
         COALESCE(SUM(first_token_ms),0)::bigint AS ttft_sum_ms,COUNT(first_token_ms)::bigint AS ttft_count,
         COALESCE(SUM(output_tokens),0)::bigint AS output_tokens,
         COALESCE(SUM(CASE WHEN output_tokens>0 AND duration_ms>0 THEN GREATEST(duration_ms-COALESCE(first_token_ms,0),1) ELSE 0 END),0)::bigint AS generation_ms
  FROM successful_candidates GROUP BY monitor_id,group_id,bucket_start
), final_errors AS (
  SELECT DISTINCT ON (COALESCE(e.api_key_id::text,'')||':'||COALESCE(NULLIF(e.request_id,''),'error:'||e.id::text))
		 e.id,e.group_id,e.created_at,e.api_key_id,e.request_id,e.platform,e.model,e.requested_model,
		 e.status_code,e.upstream_status_code,e.error_owner,e.upstream_errors,e.error_message,e.error_body,
		 e.upstream_error_message,e.upstream_error_detail,e.provider_error_code,e.provider_error_type,
		 e.is_count_tokens,e.is_business_limited,
         COALESCE(e.api_key_id::text,'')||':'||COALESCE(NULLIF(e.request_id,''),'error:'||e.id::text) AS request_key
  FROM ops_error_logs e
  WHERE e.created_at >= $1 AND e.created_at < $2 AND e.group_id IS NOT NULL
  ORDER BY COALESCE(e.api_key_id::text,'')||':'||COALESCE(NULLIF(e.request_id,''),'error:'||e.id::text),e.created_at DESC,e.id DESC
), failed_candidates AS (
	SELECT m.id AS monitor_id,e.group_id,date_trunc('minute',e.created_at) AS bucket_start,e.created_at,
         e.api_key_id,e.request_id,e.request_key
  FROM final_errors e
	JOIN model_monitors m ON m.platform=e.platform AND m.model=COALESCE(NULLIF(e.requested_model,''),e.model)
	JOIN model_monitor_groups mg ON mg.monitor_id=m.id AND mg.group_id=e.group_id
  WHERE COALESCE(e.is_count_tokens,FALSE)=FALSE AND COALESCE(e.is_business_limited,FALSE)=FALSE
    AND `+modelMonitorConfirmedFailurePredicate+`
), failed AS (
  SELECT f.monitor_id,f.group_id,f.bucket_start,COUNT(*)::bigint AS request_count
  FROM failed_candidates f
  WHERE NOT EXISTS (
    SELECT 1 FROM successful_candidates s
	WHERE NULLIF(f.request_id,'') IS NOT NULL AND s.request_id=f.request_id
      AND (f.api_key_id IS NULL OR s.api_key_id=f.api_key_id)
  )
  GROUP BY f.monitor_id,f.group_id,f.bucket_start
), combined AS (
  SELECT COALESCE(s.monitor_id,f.monitor_id) AS monitor_id,COALESCE(s.group_id,f.group_id) AS group_id,
         COALESCE(s.bucket_start,f.bucket_start) AS bucket_start,
         COALESCE(s.request_count,0)+COALESCE(f.request_count,0) AS request_count,
         COALESCE(s.success_count,0) AS success_count,COALESCE(s.latency_sum_ms,0) AS latency_sum_ms,
         COALESCE(s.latency_count,0) AS latency_count,COALESCE(s.ttft_sum_ms,0) AS ttft_sum_ms,
         COALESCE(s.ttft_count,0) AS ttft_count,COALESCE(s.output_tokens,0) AS output_tokens,
         COALESCE(s.generation_ms,0) AS generation_ms
  FROM successful s FULL JOIN failed f USING (monitor_id,group_id,bucket_start)
)
INSERT INTO model_monitor_metric_buckets
  (monitor_id,group_id,resolution,source,bucket_start,request_count,success_count,latency_sum_ms,latency_count,ttft_sum_ms,ttft_count,output_tokens,generation_ms)
SELECT monitor_id,group_id,'minute','traffic',bucket_start,request_count,success_count,latency_sum_ms,latency_count,ttft_sum_ms,ttft_count,output_tokens,generation_ms
FROM combined WHERE request_count>0`, from, to)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
WITH final_errors AS (
  SELECT DISTINCT ON (COALESCE(e.api_key_id::text,'')||':'||COALESCE(NULLIF(e.request_id,''),'error:'||e.id::text))
		 e.id,e.group_id,e.created_at,e.api_key_id,e.request_id,e.platform,e.model,e.requested_model,
		 e.status_code,e.upstream_status_code,e.error_owner,e.upstream_errors,e.error_message,e.error_body,
		 e.upstream_error_message,e.upstream_error_detail,e.provider_error_code,e.provider_error_type,
		 e.is_count_tokens,e.is_business_limited
  FROM ops_error_logs e
  WHERE e.created_at >= $1 AND e.created_at < $2 AND e.group_id IS NOT NULL
  ORDER BY COALESCE(e.api_key_id::text,'')||':'||COALESCE(NULLIF(e.request_id,''),'error:'||e.id::text),e.created_at DESC,e.id DESC
)
UPDATE model_monitor_groups mg SET last_traffic_at=x.last_traffic_at
FROM (
  SELECT monitor_id,group_id,MAX(activity_at) AS last_traffic_at FROM (
    SELECT m.id AS monitor_id,ul.group_id,ul.created_at AS activity_at
    FROM usage_logs ul JOIN accounts a ON a.id=ul.account_id
    JOIN model_monitors m ON m.platform=a.platform AND m.model=COALESCE(NULLIF(ul.requested_model,''),ul.model)
    WHERE ul.created_at >= $1 AND ul.created_at < $2 AND ul.group_id IS NOT NULL AND `+modelMonitorSuccessfulUsagePredicate+`
    UNION ALL
	SELECT m.id AS monitor_id,e.group_id,e.created_at AS activity_at
    FROM final_errors e
	JOIN model_monitors m ON m.platform=e.platform AND m.model=COALESCE(NULLIF(e.requested_model,''),e.model)
	JOIN model_monitor_groups mg ON mg.monitor_id=m.id AND mg.group_id=e.group_id
    WHERE COALESCE(e.is_count_tokens,FALSE)=FALSE AND COALESCE(e.is_business_limited,FALSE)=FALSE
      AND `+modelMonitorConfirmedFailurePredicate+`
  ) activity GROUP BY monitor_id,group_id
) x WHERE mg.monitor_id=x.monitor_id AND mg.group_id=x.group_id
  AND (mg.last_traffic_at IS NULL OR mg.last_traffic_at < x.last_traffic_at)`, from, to)
	if err != nil {
		return err
	}
	// A successful real request confirms recovery and ends any active retry loop.
	_, err = tx.ExecContext(ctx, `
UPDATE model_monitor_groups mg
SET failure_compensation_pending=FALSE,next_compensation_at=NULL,consecutive_probe_failures=0
FROM (
  SELECT m.id AS monitor_id,ul.group_id,MAX(ul.created_at) AS recovered_at
  FROM usage_logs ul JOIN accounts a ON a.id=ul.account_id
  JOIN model_monitors m ON m.platform=a.platform AND m.model=COALESCE(NULLIF(ul.requested_model,''),ul.model)
  WHERE ul.created_at >= $1 AND ul.created_at < $2 AND ul.group_id IS NOT NULL
    AND `+modelMonitorSuccessfulUsagePredicate+`
  GROUP BY m.id,ul.group_id
) recovered
WHERE mg.monitor_id=recovered.monitor_id AND mg.group_id=recovered.group_id
  AND mg.failure_compensation_pending=TRUE
  AND (mg.last_probe_at IS NULL OR recovered.recovered_at > mg.last_probe_at)`, from, to)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *modelMonitorRepository) ClaimTrafficMetricsRefresh(ctx context.Context, now, claimedUntil time.Time) (*time.Time, bool, error) {
	var cursor *time.Time
	err := r.db.QueryRowContext(ctx, `UPDATE model_monitor_runtime_state SET traffic_claimed_until=$2,updated_at=NOW()
	 WHERE id=TRUE AND (traffic_claimed_until IS NULL OR traffic_claimed_until <= $1)
	 RETURNING traffic_cursor_at`, now, claimedUntil).Scan(&cursor)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	return cursor, err == nil, err
}

func (r *modelMonitorRepository) FinishTrafficMetricsRefresh(ctx context.Context, cursor time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitor_runtime_state
	 SET traffic_cursor_at=GREATEST(COALESCE(traffic_cursor_at,'epoch'::timestamptz),$1),traffic_claimed_until=NULL,updated_at=NOW()
	 WHERE id=TRUE`, cursor.UTC())
	return err
}

func (r *modelMonitorRepository) ReleaseTrafficMetricsRefresh(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitor_runtime_state SET traffic_claimed_until=NULL,updated_at=NOW() WHERE id=TRUE`)
	return err
}

func (r *modelMonitorRepository) GroupMetrics(ctx context.Context, monitorID int64, groupIDs []int64, resolution service.ModelMonitorResolution, now time.Time) (map[int64]service.ModelMonitorGroupMetrics, error) {
	batch, err := r.GroupMetricsBatch(ctx, []service.ModelMonitorMetricScope{{MonitorID: monitorID, GroupIDs: groupIDs}}, resolution, now)
	if err != nil {
		return nil, err
	}
	if metrics, ok := batch[monitorID]; ok {
		return metrics, nil
	}
	return map[int64]service.ModelMonitorGroupMetrics{}, nil
}

func (r *modelMonitorRepository) GroupMetricsBatch(ctx context.Context, scopes []service.ModelMonitorMetricScope, resolution service.ModelMonitorResolution, now time.Time) (map[int64]map[int64]service.ModelMonitorGroupMetrics, error) {
	out := make(map[int64]map[int64]service.ModelMonitorGroupMetrics, len(scopes))
	monitorIDs := make([]int64, 0)
	groupIDs := make([]int64, 0)
	groupsByMonitor := make(map[int64][]int64, len(scopes))
	for _, scope := range scopes {
		if scope.MonitorID <= 0 || len(scope.GroupIDs) == 0 {
			continue
		}
		for _, groupID := range scope.GroupIDs {
			if groupID <= 0 {
				continue
			}
			monitorIDs = append(monitorIDs, scope.MonitorID)
			groupIDs = append(groupIDs, groupID)
			groupsByMonitor[scope.MonitorID] = append(groupsByMonitor[scope.MonitorID], groupID)
		}
	}
	if len(monitorIDs) == 0 {
		return out, nil
	}
	step := time.Minute
	if resolution == service.ModelMonitorResolutionHour {
		step = time.Hour
	} else {
		resolution = service.ModelMonitorResolutionMinute
	}
	current := now.UTC().Truncate(step)
	start := current.Add(-time.Duration(service.ModelMonitorMetricBucketCount-1) * step)
	rows, err := r.db.QueryContext(ctx, `
WITH requested AS (
 SELECT DISTINCT monitor_id,group_id FROM unnest($1::bigint[],$2::bigint[]) AS r(monitor_id,group_id)
), selected AS (
 SELECT b.* FROM model_monitor_metric_buckets b
 JOIN requested r ON r.monitor_id=b.monitor_id AND r.group_id=b.group_id
 WHERE $3='minute' AND b.resolution='minute' AND b.bucket_start >= $4 AND b.bucket_start < $5
 UNION ALL
 SELECT b.* FROM model_monitor_metric_buckets b
 JOIN requested r ON r.monitor_id=b.monitor_id AND r.group_id=b.group_id
 WHERE $3='hour' AND b.resolution='hour' AND b.bucket_start >= $4 AND b.bucket_start < date_trunc('hour',$5::timestamptz)-INTERVAL '1 hour'
 UNION ALL
 SELECT b.* FROM model_monitor_metric_buckets b
 JOIN requested r ON r.monitor_id=b.monitor_id AND r.group_id=b.group_id
 WHERE $3='hour' AND b.resolution='minute' AND b.bucket_start >= $4 AND b.bucket_start < $5
   AND (b.bucket_start >= date_trunc('hour',$5::timestamptz)-INTERVAL '1 hour' OR NOT EXISTS (
     SELECT 1 FROM model_monitor_metric_buckets h
     WHERE h.monitor_id=b.monitor_id AND h.group_id=b.group_id AND h.source=b.source
       AND h.resolution='hour' AND h.bucket_start=date_trunc('hour',b.bucket_start)
   ))
), raw AS (
 SELECT monitor_id,COALESCE(group_id,0)::bigint AS group_id,
        CASE WHEN $3='hour' THEN date_trunc('hour',bucket_start) ELSE bucket_start END AS bucket_start,
        SUM(request_count)::bigint AS request_count,SUM(success_count)::bigint AS success_count,
        SUM(latency_sum_ms)::bigint AS latency_sum_ms,SUM(latency_count)::bigint AS latency_count,
        SUM(ttft_sum_ms)::bigint AS ttft_sum_ms,SUM(ttft_count)::bigint AS ttft_count,
        SUM(output_tokens)::bigint AS output_tokens,SUM(generation_ms)::bigint AS generation_ms,
        SUM(CASE WHEN source='probe' THEN probe_cost ELSE 0 END)::float8 AS probe_cost,
        BOOL_AND(CASE WHEN source='probe' THEN probe_cost_known ELSE TRUE END) AS probe_cost_known
 FROM selected
 GROUP BY GROUPING SETS (
   (monitor_id,group_id,CASE WHEN $3='hour' THEN date_trunc('hour',bucket_start) ELSE bucket_start END),
   (monitor_id,CASE WHEN $3='hour' THEN date_trunc('hour',bucket_start) ELSE bucket_start END)
 )
)
SELECT monitor_id,group_id,bucket_start,request_count,success_count,latency_sum_ms,latency_count,
       ttft_sum_ms,ttft_count,output_tokens,generation_ms,probe_cost,probe_cost_known
FROM raw ORDER BY monitor_id,group_id,bucket_start`, pq.Array(monitorIDs), pq.Array(groupIDs), string(resolution), start, current.Add(step))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	type aggregate struct {
		requests, successes, latencySum, latencyCount, ttftSum, ttftCount, outputTokens, generationMs int64
		probeCost                                                                                     float64
		probeCostKnown                                                                                bool
		buckets                                                                                       map[time.Time]service.ModelMonitorMetricBucket
	}
	aggregates := make(map[int64]map[int64]*aggregate, len(groupsByMonitor))
	for monitorID, ids := range groupsByMonitor {
		aggregates[monitorID] = map[int64]*aggregate{0: {probeCostKnown: true, buckets: make(map[time.Time]service.ModelMonitorMetricBucket)}}
		for _, groupID := range ids {
			aggregates[monitorID][groupID] = &aggregate{probeCostKnown: true, buckets: make(map[time.Time]service.ModelMonitorMetricBucket)}
		}
	}
	for rows.Next() {
		var monitorID, groupID int64
		var bucket time.Time
		var requests, successes, latencySum, latencyCount, ttftSum, ttftCount, outputTokens, generationMs int64
		var probeCost float64
		var probeKnown bool
		if err := rows.Scan(&monitorID, &groupID, &bucket, &requests, &successes, &latencySum, &latencyCount, &ttftSum, &ttftCount, &outputTokens, &generationMs, &probeCost, &probeKnown); err != nil {
			return nil, err
		}
		a := aggregates[monitorID][groupID]
		if a == nil {
			continue
		}
		a.requests += requests
		a.successes += successes
		a.latencySum += latencySum
		a.latencyCount += latencyCount
		a.ttftSum += ttftSum
		a.ttftCount += ttftCount
		a.outputTokens += outputTokens
		a.generationMs += generationMs
		a.probeCost += probeCost
		a.probeCostKnown = a.probeCostKnown && probeKnown
		if requests > 0 {
			rate := float64(successes) * 100 / float64(requests)
			point := service.ModelMonitorMetricBucket{StartedAt: bucket.UTC(), SuccessRate: &rate}
			if ttftCount > 0 {
				ttft := float64(ttftSum) / float64(ttftCount)
				point.TTFTMs = &ttft
			}
			a.buckets[bucket.UTC()] = point
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for monitorID, ids := range groupsByMonitor {
		out[monitorID] = make(map[int64]service.ModelMonitorGroupMetrics, len(ids)+1)
		metricGroupIDs := append([]int64{0}, ids...)
		for _, groupID := range metricGroupIDs {
			a := aggregates[monitorID][groupID]
			metrics := service.ModelMonitorGroupMetrics{Buckets: make([]service.ModelMonitorMetricBucket, 0, service.ModelMonitorMetricBucketCount)}
			requestCount := a.requests
			successCount := a.successes
			failureCount := max(a.requests-a.successes, 0)
			metrics.RequestCount = &requestCount
			metrics.SuccessCount = &successCount
			metrics.FailureCount = &failureCount
			for i := 0; i < service.ModelMonitorMetricBucketCount; i++ {
				started := start.Add(time.Duration(i) * step)
				point, exists := a.buckets[started]
				if !exists {
					point = service.ModelMonitorMetricBucket{StartedAt: started}
				}
				metrics.Buckets = append(metrics.Buckets, point)
			}
			if a.requests > 0 {
				value := float64(a.successes) * 100 / float64(a.requests)
				metrics.SuccessRate = &value
			}
			if a.latencyCount > 0 {
				value := float64(a.latencySum) / float64(a.latencyCount)
				metrics.AverageLatencyMs = &value
			}
			if a.ttftCount > 0 {
				value := float64(a.ttftSum) / float64(a.ttftCount)
				metrics.TTFTMs = &value
			}
			if a.generationMs > 0 && a.outputTokens > 0 {
				value := float64(a.outputTokens) * 1000 / float64(a.generationMs)
				metrics.TPS = &value
			}
			if a.probeCostKnown {
				value := a.probeCost
				metrics.ProbeCost = &value
			}
			out[monitorID][groupID] = metrics
		}
	}
	return out, nil
}

func (r *modelMonitorRepository) RollupHourlyMetrics(ctx context.Context, hour time.Time) error {
	hour = hour.UTC().Truncate(time.Hour)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `DELETE FROM model_monitor_metric_buckets WHERE resolution='hour' AND bucket_start=$1`, hour); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO model_monitor_metric_buckets
 (monitor_id,group_id,resolution,source,bucket_start,request_count,success_count,latency_sum_ms,latency_count,ttft_sum_ms,ttft_count,output_tokens,generation_ms,probe_cost,probe_cost_known)
SELECT monitor_id,group_id,'hour',source,$1,SUM(request_count),SUM(success_count),SUM(latency_sum_ms),SUM(latency_count),
       SUM(ttft_sum_ms),SUM(ttft_count),SUM(output_tokens),SUM(generation_ms),SUM(probe_cost),BOOL_AND(probe_cost_known)
FROM model_monitor_metric_buckets
WHERE resolution='minute' AND bucket_start >= $1 AND bucket_start < $1+INTERVAL '1 hour'
GROUP BY monitor_id,group_id,source`, hour)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *modelMonitorRepository) DeleteMetricBucketsBefore(ctx context.Context, resolution service.ModelMonitorResolution, cutoff time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_monitor_metric_buckets WHERE resolution=$1 AND bucket_start<$2`, string(resolution), cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *modelMonitorRepository) DeleteHistoryBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_monitor_histories WHERE checked_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete model monitor history: %w", err)
	}
	return result.RowsAffected()
}
