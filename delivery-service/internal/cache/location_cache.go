package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var ErrInvalidRedisConfig = errors.New("invalid redis configuration")

type LocationCache interface {
	SetDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error
	GetDriverLocation(ctx context.Context, driverID int64) (float64, float64, bool, error)
	Close() error
}

type driverLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RedisLocationCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisLocationCache(addr string, ttl time.Duration) (*RedisLocationCache, error) {
	if addr == "" {
		return nil, fmt.Errorf("%w: missing REDIS_ADDR", ErrInvalidRedisConfig)
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("%w: ttl must be positive", ErrInvalidRedisConfig)
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisLocationCache{client: client, ttl: ttl}, nil
}

func (c *RedisLocationCache) SetDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error {
	if c == nil || c.client == nil {
		return ErrInvalidRedisConfig
	}

	payload, err := json.Marshal(driverLocation{
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		return err
	}

	return c.client.Set(ctx, locationKey(driverID), payload, c.ttl).Err()
}

func (c *RedisLocationCache) GetDriverLocation(ctx context.Context, driverID int64) (float64, float64, bool, error) {
	if c == nil || c.client == nil {
		return 0, 0, false, ErrInvalidRedisConfig
	}

	value, err := c.client.Get(ctx, locationKey(driverID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}

	var location driverLocation
	if err := json.Unmarshal([]byte(value), &location); err != nil {
		return 0, 0, false, err
	}

	return location.Latitude, location.Longitude, true, nil
}

func (c *RedisLocationCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func locationKey(driverID int64) string {
	return fmt.Sprintf("driver_location:%d", driverID)
}
