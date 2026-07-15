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

func (r *modelMonitorRepository) ListGroupConfigs(ctx context.Context) (map[int64][]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT monitor_id,group_id FROM model_monitor_groups ORDER BY monitor_id,priority`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int64][]int64)
	for rows.Next() {
		var monitorID, groupID int64
		if err := rows.Scan(&monitorID, &groupID); err != nil {
			return nil, err
		}
		out[monitorID] = append(out[monitorID], groupID)
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
		if _, err = tx.ExecContext(ctx, `INSERT INTO model_monitor_groups (monitor_id,group_id,priority) VALUES ($1,$2,$3)`, monitorID, groupID, priority); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *modelMonitorRepository) UpdateLastChecked(ctx context.Context, id int64, checkedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_monitors SET last_checked_at=$2, updated_at=NOW() WHERE id=$1`, id, checkedAt)
	return err
}

func (r *modelMonitorRepository) InsertHistory(ctx context.Context, h *service.ModelMonitorHistory) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO model_monitor_histories (monitor_id,status,latency_ms,attempts,message,group_id,group_name,checked_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, h.MonitorID, h.Status, h.LatencyMs, h.Attempts, h.Message, h.GroupID, h.GroupName, h.CheckedAt).Scan(&h.ID)
}

func (r *modelMonitorRepository) ListHistory(ctx context.Context, monitorID int64, limit int) ([]service.ModelMonitorHistory, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,monitor_id,status,latency_ms,attempts,message,group_id,group_name,checked_at FROM model_monitor_histories WHERE monitor_id=$1 ORDER BY checked_at DESC LIMIT $2`, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.ModelMonitorHistory, 0)
	for rows.Next() {
		var h service.ModelMonitorHistory
		if err := rows.Scan(&h.ID, &h.MonitorID, &h.Status, &h.LatencyMs, &h.Attempts, &h.Message, &h.GroupID, &h.GroupName, &h.CheckedAt); err != nil {
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
       CASE WHEN COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days')=0 THEN NULL
            ELSE 100.0*COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days' AND h7.status IN ('operational','degraded'))/COUNT(h7.id) FILTER (WHERE h7.checked_at >= NOW()-INTERVAL '7 days') END
FROM model_monitors m
LEFT JOIN LATERAL (SELECT status,latency_ms,checked_at FROM model_monitor_histories WHERE monitor_id=m.id AND m.enabled ORDER BY checked_at DESC LIMIT 1) h ON TRUE
LEFT JOIN model_monitor_histories h7 ON h7.monitor_id=m.id AND m.enabled
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
		if enabled {
			monitorKeys[monitorID] = key
			monitorIDs = append(monitorIDs, monitorID)
		}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(monitorIDs) == 0 {
		return out, nil
	}
	timelineRows, err := r.db.QueryContext(ctx, `SELECT monitor_id,status,latency_ms,checked_at FROM (
SELECT monitor_id,status,latency_ms,checked_at,ROW_NUMBER() OVER (PARTITION BY monitor_id ORDER BY checked_at DESC) AS row_num
FROM model_monitor_histories WHERE monitor_id = ANY($1)) ranked WHERE row_num <= $2 ORDER BY monitor_id,checked_at DESC`, pq.Array(monitorIDs), timelineLimit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = timelineRows.Close() }()
	for timelineRows.Next() {
		var monitorID int64
		var point service.ModelMonitorTimelinePoint
		if err := timelineRows.Scan(&monitorID, &point.Status, &point.LatencyMs, &point.CheckedAt); err != nil {
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

func (r *modelMonitorRepository) DeleteHistoryBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_monitor_histories WHERE checked_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete model monitor history: %w", err)
	}
	return result.RowsAffected()
}
