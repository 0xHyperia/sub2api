-- Model Monitor v2 only counts confirmed successes and confirmed upstream
-- model failures. Derived buckets created with the previous generic-error
-- semantics must not remain visible for the next seven days.
ALTER TABLE model_monitor_runtime_state
    ADD COLUMN IF NOT EXISTS metric_semantics_version INTEGER NOT NULL DEFAULT 1;

UPDATE model_monitor_histories
SET status = 'error'
WHERE status = 'failed'
  AND EXISTS (
      SELECT 1 FROM model_monitor_runtime_state
      WHERE id = TRUE AND metric_semantics_version < 2
  )
  AND NOT (
      (
          LOWER(COALESCE(message, '')) LIKE ANY(ARRAY[
              '%selected model is at capacity%', '%model_capacity_exhausted%',
              '%no capacity available for model%', '%server_is_overloaded%',
              '%servers are currently overloaded%', '%server is overloaded%',
              '%upstream service overloaded%', '%engine_overloaded%', '%overloaded_error%',
              '%api returned 500:%', '%api returned 503:%', '%api returned 529:%',
              '%http 500%', '%http 503%', '%http 529%',
              '%status 500%', '%status 503%', '%status 529%',
              '%status=500%', '%status=503%', '%status=529%',
              '%status_code=500%', '%status_code=503%', '%status_code=529%',
              '%"status_code":500%', '%"status_code":503%', '%"status_code":529%',
              '%(500)%', '%(503)%', '%(529)%', '%500:%', '%503:%', '%529:%'
          ])
      )
      AND NOT (
          LOWER(COALESCE(message, '')) LIKE ANY(ARRAY[
              '%model not found%', '%model_not_found%', '%unsupported model%', '%invalid model%',
              '%invalid request%', '%invalid_request%', '%context length%', '%context_length%',
              '%maximum context%', '%insufficient balance%', '%insufficient quota%', '%quota exceeded%',
              '%unauthorized%', '%forbidden%', '%authentication%', '%invalid api key%',
              '%content policy%', '%content_policy%', '%safety policy%',
              '%timeout%', '%timed out%', '%deadline exceeded%', '%tls handshake%',
              '%network error%', '%transport error%', '%connection%', '%dial tcp%', '%read tcp%', '%write tcp%',
              '%no such host%', '%name resolution%', '%x509%', '%certificate%', '%broken pipe%', '%eof%', '%stream error%',
              '%client disconnected%', '%cancelled%', '%canceled%'
          ])
      )
  );

DELETE FROM model_monitor_metric_buckets
WHERE EXISTS (
    SELECT 1 FROM model_monitor_runtime_state
    WHERE id = TRUE AND metric_semantics_version < 2
);

UPDATE model_monitor_runtime_state
SET traffic_cursor_at = NULL,
    traffic_claimed_until = NULL,
    metric_semantics_version = 2,
    updated_at = NOW()
WHERE id = TRUE AND metric_semantics_version < 2;
