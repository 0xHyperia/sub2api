package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const appAuthRedisPrefix = "app_oauth:"

var consumeAppAuthValueScript = redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if not value then
  return nil
end
redis.call("DEL", KEYS[1])
return value
`)

// consumeAppRefreshTokenScript leaves a family-bearing tombstone for the full
// refresh lifetime. A subsequent use returns the tombstone with status=reused.
var consumeAppRefreshTokenScript = redis.NewScript(`
local active = redis.call("GET", KEYS[1])
if active then
  redis.call("DEL", KEYS[1])
  redis.call("SET", KEYS[2], active, "PX", ARGV[1])
  redis.call("SET", KEYS[3], active, "PX", ARGV[2])
  return {1, active}
end
local used = redis.call("GET", KEYS[2])
if used then
  return {2, used}
end
return {0, ""}
`)

type appAuthCache struct {
	rdb *redis.Client
}

func NewAppAuthCache(rdb *redis.Client) service.AppAuthCache {
	return &appAuthCache{rdb: rdb}
}

func (c *appAuthCache) PutAuthorizationRequest(ctx context.Context, request *service.AuthorizationRequest, ttl time.Duration) error {
	return c.putJSON(ctx, appAuthRequestKey(request.RequestID), request, ttl)
}

func (c *appAuthCache) GetAuthorizationRequest(ctx context.Context, requestID string) (*service.AuthorizationRequest, error) {
	var request service.AuthorizationRequest
	if err := c.getJSON(ctx, appAuthRequestKey(requestID), &request); err != nil {
		return nil, err
	}
	return &request, nil
}

func (c *appAuthCache) ConsumeAuthorizationRequest(ctx context.Context, requestID string) (*service.AuthorizationRequest, error) {
	var request service.AuthorizationRequest
	if err := c.consumeJSON(ctx, appAuthRequestKey(requestID), &request); err != nil {
		return nil, err
	}
	return &request, nil
}

func (c *appAuthCache) PutAuthorizationCode(ctx context.Context, codeHash string, code *service.AuthorizationCode, ttl time.Duration) error {
	return c.putJSON(ctx, appAuthCodeKey(codeHash), code, ttl)
}

func (c *appAuthCache) ConsumeAuthorizationCode(ctx context.Context, codeHash string) (*service.AuthorizationCode, error) {
	var code service.AuthorizationCode
	if err := c.consumeJSON(ctx, appAuthCodeKey(codeHash), &code); err != nil {
		return nil, err
	}
	return &code, nil
}

func (c *appAuthCache) PutRefreshToken(ctx context.Context, tokenHash string, record *service.AppRefreshTokenRecord, ttl time.Duration) error {
	return c.putJSON(ctx, appAuthRefreshKey(tokenHash), record, ttl)
}

func (c *appAuthCache) GetRefreshToken(ctx context.Context, tokenHash string) (*service.AppRefreshTokenRecord, error) {
	var record service.AppRefreshTokenRecord
	if err := c.getJSON(ctx, appAuthRefreshKey(tokenHash), &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (c *appAuthCache) RotateRefreshToken(ctx context.Context, oldTokenHash, newTokenHash string, _ *service.AppRefreshTokenRecord, tokenTTL, tombstoneTTL time.Duration) (service.RefreshRotationResult, *service.AppRefreshTokenRecord, error) {
	result, err := consumeAppRefreshTokenScript.Run(ctx, c.rdb,
		[]string{appAuthRefreshKey(oldTokenHash), appAuthUsedRefreshKey(oldTokenHash), appAuthRefreshKey(newTokenHash)},
		tombstoneTTL.Milliseconds(), tokenTTL.Milliseconds(),
	).Slice()
	if err != nil {
		return service.RefreshRotationMissing, nil, err
	}
	if len(result) != 2 {
		return service.RefreshRotationMissing, nil, fmt.Errorf("unexpected refresh rotation response")
	}
	status, ok := result[0].(int64)
	if !ok {
		return service.RefreshRotationMissing, nil, fmt.Errorf("unexpected refresh rotation status %T", result[0])
	}
	if status == 0 {
		return service.RefreshRotationMissing, nil, nil
	}
	payload, ok := result[1].(string)
	if !ok {
		return service.RefreshRotationMissing, nil, fmt.Errorf("unexpected refresh rotation payload %T", result[1])
	}
	var record service.AppRefreshTokenRecord
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return service.RefreshRotationMissing, nil, fmt.Errorf("decode refresh token record: %w", err)
	}
	if status == 2 {
		return service.RefreshRotationReused, &record, nil
	}
	return service.RefreshRotationSucceeded, &record, nil
}

func (c *appAuthCache) SetGrantRevoked(ctx context.Context, grantID string, ttl time.Duration) error {
	return c.rdb.Set(ctx, appAuthRevokedKey(grantID), "1", ttl).Err()
}

func (c *appAuthCache) IsGrantRevoked(ctx context.Context, grantID string) (bool, error) {
	count, err := c.rdb.Exists(ctx, appAuthRevokedKey(grantID)).Result()
	return count > 0, err
}

func (c *appAuthCache) putJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, payload, ttl).Err()
}

func (c *appAuthCache) getJSON(ctx context.Context, key string, target any) error {
	payload, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return service.ErrAppAuthCacheMiss
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func (c *appAuthCache) consumeJSON(ctx context.Context, key string, target any) error {
	raw, err := consumeAppAuthValueScript.Run(ctx, c.rdb, []string{key}).Result()
	if errors.Is(err, redis.Nil) {
		return service.ErrAppAuthCacheMiss
	}
	if err != nil {
		return err
	}
	payload, ok := raw.(string)
	if !ok {
		return fmt.Errorf("unexpected consumed app auth value %T", raw)
	}
	return json.Unmarshal([]byte(payload), target)
}

func appAuthRequestKey(id string) string       { return appAuthRedisPrefix + "req:" + id }
func appAuthCodeKey(hash string) string        { return appAuthRedisPrefix + "code:" + hash }
func appAuthRefreshKey(hash string) string     { return appAuthRedisPrefix + "rt:" + hash }
func appAuthUsedRefreshKey(hash string) string { return appAuthRedisPrefix + "used:" + hash }
func appAuthRevokedKey(grantID string) string  { return appAuthRedisPrefix + "revoked:" + grantID }
