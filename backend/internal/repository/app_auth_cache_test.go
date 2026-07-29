package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestAppAuthAuthorizationCodeIsConsumedOnce(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewAppAuthCache(rdb)
	ctx := context.Background()

	require.NoError(t, cache.PutAuthorizationCode(ctx, "hash", &service.AuthorizationCode{UserID: 42}, time.Minute))
	code, err := cache.ConsumeAuthorizationCode(ctx, "hash")
	require.NoError(t, err)
	require.Equal(t, int64(42), code.UserID)
	_, err = cache.ConsumeAuthorizationCode(ctx, "hash")
	require.ErrorIs(t, err, service.ErrAppAuthCacheMiss)
}

func TestAppAuthRefreshRotationIsAtomicAndDetectsReuse(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewAppAuthCache(rdb)
	ctx := context.Background()
	record := &service.AppRefreshTokenRecord{UserID: 42, ClientID: "zeroagent-desktop", GrantID: "grant", FamilyID: "family"}
	require.NoError(t, cache.PutRefreshToken(ctx, "old", record, time.Hour))

	results := make(chan service.RefreshRotationResult, 2)
	var wg sync.WaitGroup
	for _, next := range []string{"next-a", "next-b"} {
		wg.Add(1)
		go func(next string) {
			defer wg.Done()
			result, got, err := cache.RotateRefreshToken(ctx, "old", next, nil, time.Hour, time.Hour)
			require.NoError(t, err)
			require.Equal(t, "grant", got.GrantID)
			results <- result
		}(next)
	}
	wg.Wait()
	close(results)

	counts := map[service.RefreshRotationResult]int{}
	for result := range results {
		counts[result]++
	}
	require.Equal(t, 1, counts[service.RefreshRotationSucceeded])
	require.Equal(t, 1, counts[service.RefreshRotationReused])
}
