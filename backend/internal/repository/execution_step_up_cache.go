package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	executionStepUpProofPrefix    = "execution:stepup:proof:"
	executionStepUpAttemptsPrefix = "execution:stepup:attempts:"
	executionStepUpAttemptsTTL    = 15 * time.Minute
)

var consumeExecutionProofScript = redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if value then
  redis.call("DEL", KEYS[1])
end
return value
`)

type ExecutionStepUpRedisCache struct {
	rdb *redis.Client
}

func NewExecutionStepUpCache(rdb *redis.Client) service.ExecutionStepUpCache {
	return &ExecutionStepUpRedisCache{rdb: rdb}
}

func (c *ExecutionStepUpRedisCache) StoreExecutionProof(
	ctx context.Context,
	proofHash string,
	grant service.ExecutionStepUpGrant,
	ttl time.Duration,
) error {
	data, err := json.Marshal(grant)
	if err != nil {
		return fmt.Errorf("marshal execution step-up grant: %w", err)
	}
	return c.rdb.Set(ctx, executionStepUpProofPrefix+proofHash, data, ttl).Err()
}

func (c *ExecutionStepUpRedisCache) ConsumeExecutionProof(ctx context.Context, proofHash string) (*service.ExecutionStepUpGrant, error) {
	value, err := consumeExecutionProofScript.Run(ctx, c.rdb, []string{executionStepUpProofPrefix + proofHash}).Text()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var grant service.ExecutionStepUpGrant
	if err := json.Unmarshal([]byte(value), &grant); err != nil {
		return nil, fmt.Errorf("unmarshal execution step-up grant: %w", err)
	}
	return &grant, nil
}

func executionStepUpAttemptsKey(userID int64) string {
	return fmt.Sprintf("%s%d", executionStepUpAttemptsPrefix, userID)
}

func (c *ExecutionStepUpRedisCache) GetExecutionStepUpAttempts(ctx context.Context, userID int64) (int, error) {
	value, err := c.rdb.Get(ctx, executionStepUpAttemptsKey(userID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}

func (c *ExecutionStepUpRedisCache) IncrementExecutionStepUpAttempts(ctx context.Context, userID int64) (int, error) {
	key := executionStepUpAttemptsKey(userID)
	pipe := c.rdb.TxPipeline()
	count := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, executionStepUpAttemptsTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return int(count.Val()), nil
}

func (c *ExecutionStepUpRedisCache) ClearExecutionStepUpAttempts(ctx context.Context, userID int64) error {
	return c.rdb.Del(ctx, executionStepUpAttemptsKey(userID)).Err()
}
