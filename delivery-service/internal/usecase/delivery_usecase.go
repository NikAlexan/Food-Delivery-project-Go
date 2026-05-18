package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"food-delivery/delivery-service/internal/email"
	"food-delivery/delivery-service/internal/model"
	"food-delivery/delivery-service/internal/repository"
)

var (
	ErrInvalidDelivery    = errors.New("invalid delivery request")
	ErrInvalidCoordinates = errors.New("invalid driver coordinates")
	ErrAlreadyRegistered  = repository.ErrAlreadyRegistered
)

type EventPublisher interface {
	PublishDeliveryCompleted(ctx context.Context, delivery *model.Delivery) error
}

type LocationCache interface {
	SetDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error
	GetDriverLocation(ctx context.Context, driverID int64) (float64, float64, bool, error)
}

type DeliveryUsecase interface {
	AssignDriver(ctx context.Context, input AssignInput) (*model.Delivery, error)
	GetDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error)
	UpdateDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error
	TrackDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error)
	CompleteDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error)
	ListDriverDeliveries(ctx context.Context, driverID int64) ([]model.Delivery, error)
	GetDeliveryHistory(ctx context.Context, userID int64) ([]model.Delivery, error)
	CancelDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error)
	HandleOrderCreated(ctx context.Context, event model.OrderEvent) error
	HandleOrderPaid(ctx context.Context, event model.OrderEvent) error
	HandleOrderCancelled(ctx context.Context, event model.OrderEvent) error
	RegisterDriver(ctx context.Context, userID int64, name, email, phone string) (*model.Driver, error)
	GetMyDriver(ctx context.Context, userID int64) (*model.Driver, error)
	SetAvailability(ctx context.Context, userID int64, available bool) (*model.Driver, error)
}

type AssignInput struct {
	OrderID         int64
	UserID          int64
	UserEmail       string
	DeliveryAddress string
}

type deliveryUsecase struct {
	repo      repository.DeliveryRepository
	mailer    email.Sender
	publisher EventPublisher
	cache     LocationCache
}

func NewDeliveryUsecase(repo repository.DeliveryRepository, mailer email.Sender, publisher EventPublisher, cache LocationCache) DeliveryUsecase {
	return &deliveryUsecase{
		repo:      repo,
		mailer:    mailer,
		publisher: publisher,
		cache:     cache,
	}
}

func (u *deliveryUsecase) AssignDriver(ctx context.Context, input AssignInput) (*model.Delivery, error) {
	if input.OrderID <= 0 || input.UserID <= 0 || input.DeliveryAddress == "" {
		return nil, ErrInvalidDelivery
	}

	delivery, changed, err := u.repo.AssignDriver(ctx, input.OrderID, input.UserID, input.UserEmail, input.DeliveryAddress)
	if err != nil {
		return nil, err
	}

	if changed {
		u.notifyBestEffort(ctx, delivery.UserEmail,
			fmt.Sprintf("Order #%d confirmed", delivery.OrderID),
			fmt.Sprintf("Your order #%d has been assigned to courier %s.", delivery.OrderID, delivery.DriverName),
		)
	}
	return delivery, nil
}

func (u *deliveryUsecase) GetDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	return u.repo.GetByID(ctx, deliveryID)
}

func (u *deliveryUsecase) UpdateDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error {
	if driverID <= 0 || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return ErrInvalidCoordinates
	}
	if err := u.repo.UpdateDriverLocation(ctx, driverID, latitude, longitude); err != nil {
		return err
	}

	if u.cache != nil {
		if err := u.cache.SetDriverLocation(ctx, driverID, latitude, longitude); err != nil {
			log.Printf("redis location cache update failed: %v", err)
		}
	}

	return nil
}

func (u *deliveryUsecase) TrackDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	delivery, err := u.repo.GetByID(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	u.hydrateLocation(ctx, delivery)
	return delivery, nil
}

func (u *deliveryUsecase) CompleteDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	delivery, changed, err := u.repo.CompleteDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	if changed {
		u.notifyBestEffort(ctx, delivery.UserEmail,
			fmt.Sprintf("Order #%d delivered", delivery.OrderID),
			fmt.Sprintf("Delivery #%d has been completed. Enjoy your meal.", delivery.ID),
		)
		u.publishBestEffort(ctx, delivery)
	}

	return delivery, nil
}

func (u *deliveryUsecase) ListDriverDeliveries(ctx context.Context, driverID int64) ([]model.Delivery, error) {
	deliveries, err := u.repo.ListDriverDeliveries(ctx, driverID)
	if err != nil {
		return nil, err
	}

	for i := range deliveries {
		u.hydrateLocation(ctx, &deliveries[i])
	}
	return deliveries, nil
}

func (u *deliveryUsecase) GetDeliveryHistory(ctx context.Context, userID int64) ([]model.Delivery, error) {
	deliveries, err := u.repo.GetDeliveryHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range deliveries {
		u.hydrateLocation(ctx, &deliveries[i])
	}
	return deliveries, nil
}

func (u *deliveryUsecase) CancelDelivery(ctx context.Context, deliveryID int64) (*model.Delivery, error) {
	delivery, _, err := u.repo.CancelDelivery(ctx, deliveryID)
	return delivery, err
}

func (u *deliveryUsecase) HandleOrderCreated(ctx context.Context, event model.OrderEvent) error {
	_, err := u.AssignDriver(ctx, AssignInput{
		OrderID:         event.OrderID,
		UserID:          event.UserID,
		UserEmail:       event.UserEmail,
		DeliveryAddress: event.DeliveryAddress,
	})
	return err
}

func (u *deliveryUsecase) HandleOrderPaid(ctx context.Context, event model.OrderEvent) error {
	delivery, changed, err := u.repo.MarkInTransitByOrderID(ctx, event.OrderID)
	if err != nil {
		return err
	}

	if changed {
		u.notifyBestEffort(ctx, delivery.UserEmail,
			fmt.Sprintf("Order #%d is on the way", delivery.OrderID),
			fmt.Sprintf("Courier %s is on the way with your order.", delivery.DriverName),
		)
	}
	return nil
}

func (u *deliveryUsecase) HandleOrderCancelled(ctx context.Context, event model.OrderEvent) error {
	_, _, err := u.repo.CancelByOrderID(ctx, event.OrderID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil
	}
	return err
}

func (u *deliveryUsecase) hydrateLocation(ctx context.Context, delivery *model.Delivery) {
	if u.cache == nil || delivery == nil || delivery.DriverID <= 0 {
		return
	}

	latitude, longitude, found, err := u.cache.GetDriverLocation(ctx, delivery.DriverID)
	if err != nil {
		log.Printf("redis location cache read failed: %v", err)
		return
	}
	if !found {
		return
	}

	delivery.CurrentLatitude = latitude
	delivery.CurrentLongitude = longitude
}

func (u *deliveryUsecase) notifyBestEffort(ctx context.Context, to, subject, body string) {
	if u.mailer == nil {
		return
	}
	if err := u.mailer.Send(ctx, email.Message{To: to, Subject: subject, Body: body}); err != nil {
		log.Printf("email send failed: %v", err)
	}
}

func (u *deliveryUsecase) RegisterDriver(ctx context.Context, userID int64, name, email, phone string) (*model.Driver, error) {
	d := &model.Driver{
		UserID: &userID,
		Name:   name,
		Email:  email,
		Phone:  phone,
	}
	return u.repo.CreateDriver(ctx, d)
}

func (u *deliveryUsecase) GetMyDriver(ctx context.Context, userID int64) (*model.Driver, error) {
	return u.repo.GetDriverByUserID(ctx, userID)
}

func (u *deliveryUsecase) publishBestEffort(ctx context.Context, delivery *model.Delivery) {
	if u.publisher == nil {
		return
	}
	if err := u.publisher.PublishDeliveryCompleted(ctx, delivery); err != nil {
		log.Printf("publish delivery.completed failed: %v", err)
	}
}

func (u *deliveryUsecase) SetAvailability(ctx context.Context, userID int64, available bool) (*model.Driver, error) {
	driver, err := u.repo.GetDriverByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u.repo.SetDriverAvailability(ctx, driver.ID, available)
}
