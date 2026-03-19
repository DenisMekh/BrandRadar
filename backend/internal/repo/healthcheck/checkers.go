package healthcheck

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

// PostgresChecker — проверка доступности PostgreSQL.
type PostgresChecker struct {
	pool *pgxpool.Pool
}

func NewPostgresChecker(pool *pgxpool.Pool) *PostgresChecker {
	return &PostgresChecker{pool: pool}
}

func (c *PostgresChecker) Ping(ctx context.Context) error {
	if err := c.pool.Ping(ctx); err != nil {
		return fmt.Errorf("PostgresChecker.Ping: %w", err)
	}
	return nil
}

func (c *PostgresChecker) Name() string { return "postgres" }

// RedisChecker — проверка доступности Redis.
type RedisChecker struct {
	client *goredis.Client
}

func NewRedisChecker(client *goredis.Client) *RedisChecker {
	return &RedisChecker{client: client}
}

func (c *RedisChecker) Ping(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("RedisChecker.Ping: %w", err)
	}
	return nil
}

func (c *RedisChecker) Name() string { return "redis" }
