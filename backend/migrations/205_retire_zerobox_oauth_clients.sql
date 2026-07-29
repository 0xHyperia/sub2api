-- ZeroBox OAuth clients are permanently retired. Existing grants must not be
-- relabeled because authorization codes and refresh-token families bind to the
-- original client_id.
UPDATE app_authorizations
SET status = 'revoked',
    revoked_at = COALESCE(revoked_at, NOW()),
    updated_at = NOW()
WHERE client_id LIKE 'zerobox-%'
  AND status = 'active';
