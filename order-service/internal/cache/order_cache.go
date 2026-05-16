package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"food-delivery/order-service/internal/model"
)

const statusTTL = 5 * time.Minute

type OrderCache struct {
	client *redis.Client
}

func NewOrderCache(addr string) *OrderCache {
	return &OrderCache{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

func (c *OrderCache) GetStatus(ctx context.Context, orderID int64) (model.OrderStatus, bool) {
	val, err := c.client.Get(ctx, statusKey(orderID)).Result()
	if err != nil {
		return "", false
	}
	return model.OrderStatus(val), true
}

func (c *OrderCache) SetStatus(ctx context.Context, orderID int64, status model.OrderStatus) {
	if err := c.client.Set(ctx, statusKey(orderID), string(status), statusTTL).Err(); err != nil {
		log.Printf("cache: set order status %d: %v", orderID, err)
	}
}

func (c *OrderCache) DeleteStatus(ctx context.Context, orderID int64) {
	if err := c.client.Del(ctx, statusKey(orderID)).Err(); err != nil {
		log.Printf("cache: del order status %d: %v", orderID, err)
	}
}

func statusKey(orderID int64) string {
	return fmt.Sprintf("order:status:%d", orderID)
}
