package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"

	"food-delivery/restaurant-service/internal/model"
	"food-delivery/restaurant-service/internal/repository"
)

const (
	cacheRestaurantTTL = 5 * time.Minute
	cacheMenuTTL       = 2 * time.Minute
	natsTopicMenu      = "restaurant.menu_updated"
)

var (
	// ErrNotFound is surfaced from the repository layer.
	ErrNotFound = repository.ErrNotFound
	// ErrInvalidInput is returned for bad request data.
	ErrInvalidInput = errors.New("invalid input")
)

// RestaurantUsecase defines all business operations.
type RestaurantUsecase interface {
	CreateRestaurant(ctx context.Context, r *model.Restaurant) (*model.Restaurant, error)
	GetRestaurant(ctx context.Context, id int64) (*model.Restaurant, error)
	GetMyRestaurant(ctx context.Context, ownerID int64) (*model.Restaurant, error)
	UpdateRestaurant(ctx context.Context, r *model.Restaurant) (*model.Restaurant, error)
	DeleteRestaurant(ctx context.Context, id, ownerID int64) error
	ListRestaurants(ctx context.Context, f model.ListFilter) ([]*model.Restaurant, int, error)
	SearchRestaurants(ctx context.Context, f model.SearchFilter) ([]*model.Restaurant, int, error)

	CreateMenuItem(ctx context.Context, item *model.MenuItem) (*model.MenuItem, error)
	UpdateMenuItem(ctx context.Context, item *model.MenuItem) (*model.MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int64) error
	GetMenu(ctx context.Context, restaurantID int64) ([]*model.MenuItem, error)
}

type restaurantUsecase struct {
	repo   repository.RestaurantRepository
	cache  *redis.Client
	js     jetstream.JetStream
	logger *slog.Logger
}

// NewRestaurantUsecase wires up the usecase with all dependencies.
func NewRestaurantUsecase(
	repo repository.RestaurantRepository,
	cache *redis.Client,
	js jetstream.JetStream,
	logger *slog.Logger,
) RestaurantUsecase {
	return &restaurantUsecase{repo: repo, cache: cache, js: js, logger: logger}
}

// ── Restaurants ───────────────────────────────────────────────────────────────

func (u *restaurantUsecase) CreateRestaurant(ctx context.Context, r *model.Restaurant) (*model.Restaurant, error) {
	if r.Name == "" {
		return nil, ErrInvalidInput
	}
	if err := u.repo.CreateRestaurant(ctx, r); err != nil {
		u.logger.ErrorContext(ctx, "CreateRestaurant repo error", "error", err)
		return nil, err
	}
	u.logger.InfoContext(ctx, "restaurant created", "id", r.ID, "name", r.Name)
	return r, nil
}

func (u *restaurantUsecase) GetRestaurant(ctx context.Context, id int64) (*model.Restaurant, error) {
	cacheKey := restaurantCacheKey(id)

	// Try cache first.
	cached, err := u.cache.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var r model.Restaurant
		if jsonErr := json.Unmarshal(cached, &r); jsonErr == nil {
			return &r, nil
		}
	}

	r, err := u.repo.GetRestaurantByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Populate cache; ignore cache errors.
	if data, jsonErr := json.Marshal(r); jsonErr == nil {
		u.cache.Set(ctx, cacheKey, data, cacheRestaurantTTL)
	}

	return r, nil
}

func (u *restaurantUsecase) GetMyRestaurant(ctx context.Context, ownerID int64) (*model.Restaurant, error) {
	return u.repo.GetByOwnerID(ctx, ownerID)
}

func (u *restaurantUsecase) UpdateRestaurant(ctx context.Context, r *model.Restaurant) (*model.Restaurant, error) {
	if r.Name == "" {
		return nil, ErrInvalidInput
	}
	if err := u.repo.UpdateRestaurant(ctx, r); err != nil {
		return nil, err
	}
	u.cache.Del(ctx, restaurantCacheKey(r.ID))
	u.logger.InfoContext(ctx, "restaurant updated", "id", r.ID)
	return u.repo.GetRestaurantByID(ctx, r.ID)
}

func (u *restaurantUsecase) DeleteRestaurant(ctx context.Context, id, ownerID int64) error {
	if err := u.repo.DeleteRestaurant(ctx, id, ownerID); err != nil {
		return err
	}
	u.cache.Del(ctx, restaurantCacheKey(id))
	u.cache.Del(ctx, menuCacheKey(id))
	u.logger.InfoContext(ctx, "restaurant deleted", "id", id)
	return nil
}

func (u *restaurantUsecase) ListRestaurants(ctx context.Context, f model.ListFilter) ([]*model.Restaurant, int, error) {
	return u.repo.ListRestaurants(ctx, f)
}

func (u *restaurantUsecase) SearchRestaurants(ctx context.Context, f model.SearchFilter) ([]*model.Restaurant, int, error) {
	if f.Query == "" {
		return nil, 0, ErrInvalidInput
	}
	return u.repo.SearchRestaurants(ctx, f)
}

// ── Menu ──────────────────────────────────────────────────────────────────────

func (u *restaurantUsecase) CreateMenuItem(ctx context.Context, item *model.MenuItem) (*model.MenuItem, error) {
	if item.Name == "" || item.Price <= 0 {
		return nil, ErrInvalidInput
	}
	if err := u.repo.CreateMenuItem(ctx, item); err != nil {
		u.logger.ErrorContext(ctx, "CreateMenuItem repo error", "error", err)
		return nil, err
	}
	u.cache.Del(ctx, menuCacheKey(item.RestaurantID))
	u.publishMenuUpdated(ctx, item.RestaurantID)
	u.logger.InfoContext(ctx, "menu item created", "item_id", item.ID, "restaurant_id", item.RestaurantID)
	return item, nil
}

func (u *restaurantUsecase) UpdateMenuItem(ctx context.Context, item *model.MenuItem) (*model.MenuItem, error) {
	if item.Name == "" || item.Price <= 0 {
		return nil, ErrInvalidInput
	}
	if err := u.repo.UpdateMenuItem(ctx, item); err != nil {
		return nil, err
	}
	u.cache.Del(ctx, menuCacheKey(item.RestaurantID))
	u.publishMenuUpdated(ctx, item.RestaurantID)
	u.logger.InfoContext(ctx, "menu item updated", "item_id", item.ID)
	return u.repo.GetMenuItemByID(ctx, item.ID)
}

func (u *restaurantUsecase) DeleteMenuItem(ctx context.Context, id int64) error {
	item, err := u.repo.GetMenuItemByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.repo.DeleteMenuItem(ctx, id); err != nil {
		return err
	}
	u.cache.Del(ctx, menuCacheKey(item.RestaurantID))
	u.publishMenuUpdated(ctx, item.RestaurantID)
	u.logger.InfoContext(ctx, "menu item deleted", "item_id", id)
	return nil
}

func (u *restaurantUsecase) GetMenu(ctx context.Context, restaurantID int64) ([]*model.MenuItem, error) {
	cacheKey := menuCacheKey(restaurantID)

	cached, err := u.cache.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var items []*model.MenuItem
		if jsonErr := json.Unmarshal(cached, &items); jsonErr == nil {
			return items, nil
		}
	}

	items, err := u.repo.GetMenuByRestaurant(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	if data, jsonErr := json.Marshal(items); jsonErr == nil {
		u.cache.Set(ctx, cacheKey, data, cacheMenuTTL)
	}

	return items, nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func restaurantCacheKey(id int64) string {
	return fmt.Sprintf("restaurant:%d", id)
}

func menuCacheKey(restaurantID int64) string {
	return fmt.Sprintf("menu:%d", restaurantID)
}

type menuUpdatedEvent struct {
	RestaurantID int64 `json:"restaurant_id"`
}

func (u *restaurantUsecase) publishMenuUpdated(ctx context.Context, restaurantID int64) {
	if u.js == nil {
		return
	}
	data, err := json.Marshal(menuUpdatedEvent{RestaurantID: restaurantID})
	if err != nil {
		u.logger.ErrorContext(ctx, "marshal menu event", "error", err)
		return
	}
	if _, err := u.js.Publish(ctx, natsTopicMenu, data); err != nil {
		u.logger.WarnContext(ctx, "NATS publish failed", "topic", natsTopicMenu, "error", err)
	}
}
