# Promotion tracking operations

## Required production configuration

Set `DISTRIBUTION_TRACKING_HASH_SECRETS` to an independent random value of at
least 32 bytes. During rotation, place the new secret first and retain prior
secrets for at least the longest configured attribution window. Do not reuse
`JWT_SECRET`. When the independent secret is absent, tracking fails closed
instead of hashing with the JWT key or an empty key.

Direct deployments should leave `SERVER_TRUSTED_PROXIES` empty and keep
`SECURITY_TRUST_FORWARDED_IP_FOR_API_KEY_ACL=false`. Behind a reverse proxy,
set `SERVER_TRUSTED_PROXIES` to only the exact IP/CIDR values that connect
directly to Sub2API. This keeps public tracking rate limits and IP hashing on
the same trusted client-IP chain.

Standalone deployments connect to external PostgreSQL and Redis. Use
`DATABASE_SSLMODE=verify-full` and `REDIS_ENABLE_TLS=true` across hosts or
untrusted networks. The Compose file requires an explicit choice; `disable`
and `false` are acceptable only on a protected private network.

The visitor hashes stored in promotion archives are pseudonymous, not fully
anonymous. They cannot be reversed to the browser token, but they remain
linkable across reporting dates. Restrict database access and include these
rows in the platform's retention and deletion policy.

## Enabling physical cleanup

Privacy scrubbing and expired-attribution removal always run. Physical archive
and deletion are a separate opt-in operation and default to disabled:

```dotenv
DISTRIBUTION_TRACKING_CLEANUP_ENABLED=false
DISTRIBUTION_TRACKING_CLEANUP_INTERVAL_MINUTES=10
DISTRIBUTION_TRACKING_CLEANUP_BATCH_SIZE=5000
```

Before changing the switch to `true`:

1. Create a database backup and verify that it can be restored.
2. Confirm migration 192 completed successfully.
3. Record the oldest raw promotion visit and the current archive/skip counts.
4. Run the first cleanup in a controlled window and compare raw plus archived
   analytics for the same date range.
5. Alert on cleanup errors, repeated protected-day skips, and a growing oldest
   raw-visit backlog.

Useful checks:

```sql
SELECT MIN(visited_at) AS oldest_raw_visit,
       COUNT(*) AS raw_visits
FROM distribution_promotion_visits;

SELECT rollup_kind, MIN(stat_date), MAX(stat_date), SUM(visits)
FROM distribution_promotion_daily_rollups
GROUP BY rollup_kind;

SELECT *
FROM distribution_promotion_archive_skips
ORDER BY last_detected_at DESC;

SELECT * FROM distribution_promotion_maintenance WHERE id = 1;

SELECT status, COUNT(*), MIN(updated_at) AS oldest
FROM distribution_binding_claims
GROUP BY status;

SELECT COUNT(*) AS overdue_unscrubbed
FROM distribution_promotion_visits v
JOIN distribution_settings s ON s.id = 1
WHERE v.detail_expired_at IS NULL
  AND v.visited_at < NOW() - make_interval(days => s.promotion_detail_retention_days);
```

`last_privacy_scrub_at` must continue advancing while physical cleanup is
disabled. Alert if it is older than twice the configured cleanup interval.
Alert on growing `pending` binding claims and investigate every `conflict`
claim; these rows make registration-time attribution failures durable rather
than silently losing ownership.

## Upgrade and rollback constraints

Promotion tracking ships as the single unpublished migration 192 baseline.
There is no supported production state between the former development patches.
After physical cleanup has deleted raw visits, those rows cannot be reconstructed
from rollups. Rollback therefore means stopping the application and cleanup
worker, restoring the pre-upgrade database backup, and then deploying the old
image. Switching only the image is not sufficient.

## Development demo data

The demo seed requires both an exact non-production database name and an
explicit confirmation token. Use a dedicated database whose name contains
`dev`, `development`, `test`, `local`, or `demo`:

```bash
psql "$DATABASE_URL" \
  -v ALLOW_DISTRIBUTION_DEMO_SEED=I_UNDERSTAND_NON_PRODUCTION_ONLY \
  -v DISTRIBUTION_DEMO_DATABASE=sub2api_local \
  -f backend/scripts/seed-distribution-demo.sql
```

Seeded users are disabled and use a non-login password marker. The script never
copies an administrator password hash.

## Legacy Rollup Recovery

The compatibility path below is only for development or rehearsal databases
that ran pre-squash versions of promotion cleanup. Production never received
those patch migrations. Inspect the compatibility view before reusing such a
database:

```sql
SELECT * FROM distribution_promotion_legacy_rollup_status;
```

If `legacy_rows` is non-zero, those aggregates cannot provide exact visitor
cohort filtering. Restore `distribution_promotion_visits` and
`distribution_promotion_conversions` for the reported date range from the last
pre-cleanup backup, then run:

```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
  -f backend/scripts/recover-distribution-promotion-legacy-rollups.sql
```

The recovery script is fail-closed: it compares every restored raw aggregate
with its legacy row and deletes legacy rows only after an exact match. If the
raw backup is unavailable, keep the legacy rows for audit and accept that exact
analytics for that date range returns `PROMOTION_ARCHIVE_UNAVAILABLE`; do not
delete or approximate them.

## Pending Registration Bindings

Registration writes `distribution_binding_claims` before applying the agent
binding. Transient failures remain `pending` and are retried whenever an access
or refresh token is issued. Permanent ownership/code conflicts are retained as
`conflict` for manual review instead of being silently discarded:

```sql
SELECT user_id,promotion_code,signup_source,status,attempts,last_error,updated_at
FROM distribution_binding_claims
WHERE status<>'completed'
ORDER BY updated_at;
```

Do not delete conflict rows until the user's affiliate/distribution ownership
has been reviewed. The `user_promotion_ownerships` row is the authoritative
database-level exclusivity decision.
