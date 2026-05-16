package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"food-delivery/user-service/internal/model"
)

const profileTTL = 5 * time.Minute

type UserCache struct {
	client *redis.Client
}

func NewUserCache(addr string) *UserCache {
	return &UserCache{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (c *UserCache) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	data, err := c.client.Get(ctx, profileKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var u model.User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *UserCache) SetProfile(ctx context.Context, user *model.User) {
	data, err := json.Marshal(user)
	if err != nil {
		log.Printf("cache: marshal user %d: %v", user.ID, err)
		return
	}
	if err := c.client.Set(ctx, profileKey(user.ID), data, profileTTL).Err(); err != nil {
		log.Printf("cache: set user %d: %v", user.ID, err)
	}
}

func (c *UserCache) DeleteProfile(ctx context.Context, userID int64) {
	if err := c.client.Del(ctx, profileKey(userID)).Err(); err != nil {
		log.Printf("cache: del user %d: %v", userID, err)
	}
}

func profileKey(userID int64) string {
	return fmt.Sprintf("user:profile:%d", userID)
}
