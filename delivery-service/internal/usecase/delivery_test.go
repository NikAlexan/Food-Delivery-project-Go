package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"food-delivery/delivery-service/internal/email"
	"food-delivery/delivery-service/internal/model"
)

type recordingMailer struct {
	messages []email.Message
	err      error
}

func (m *recordingMailer) Send(ctx context.Context, msg email.Message) error {
	m.messages = append(m.messages, msg)
	return m.err
}

type recordingPublisher struct {
	deliveries []*model.Delivery
	err        error
}

func (p *recordingPublisher) PublishDeliveryCompleted(ctx context.Context, delivery *model.Delivery) error {
	p.deliveries = append(p.deliveries, delivery)
	return p.err
}

type recordingLocationCache struct {
	locations map[int64][2]float64
	err       error
}

func (c *recordingLocationCache) SetDriverLocation(ctx context.Context, driverID int64, latitude, longitude float64) error {
	if c.err != nil {
		return c.err
	}
	if c.locations == nil {
		c.locations = make(map[int64][2]float64)
	}
	c.locations[driverID] = [2]float64{latitude, longitude}
	return nil
}

func (c *recordingLocationCache) GetDriverLocation(ctx context.Context, driverID int64) (float64, float64, bool, error) {
	if c.err != nil {
		return 0, 0, false, c.err
	}
	location, ok := c.locations[driverID]
	if !ok {
		return 0, 0, false, nil
	}
	return location[0], location[1], true, nil
}

func TestAssignDriver_SendsConfirmationEmail(t *testing.T) {
	repo := &mockDeliveryRepo{}
	mailer := &recordingMailer{}
	publisher := &recordingPublisher{}
	uc := NewDeliveryUsecase(repo, mailer, publisher, nil)
	ctx := context.Background()

	delivery := &model.Delivery{
		ID:         7,
		OrderID:    501,
		UserID:     9,
		UserEmail:  "user@example.com",
		DriverID:   3,
		DriverName: "Courier One",
		Status:     model.StatusAssigned,
	}

	repo.On("AssignDriver", ctx, int64(501), int64(9), "user@example.com", "Abay 10").Return(delivery, true, nil)

	result, err := uc.AssignDriver(ctx, AssignInput{
		OrderID:         501,
		UserID:          9,
		UserEmail:       "user@example.com",
		DeliveryAddress: "Abay 10",
	})
	require.NoError(t, err)
	assert.Equal(t, delivery, result)
	assert.Len(t, mailer.messages, 1)
	assert.Equal(t, "user@example.com", mailer.messages[0].To)
	assert.Contains(t, mailer.messages[0].Body, "Courier One")
	repo.AssertExpectations(t)
}

func TestAssignDriver_ValidatesInput(t *testing.T) {
	repo := &mockDeliveryRepo{}
	uc := NewDeliveryUsecase(repo, nil, nil, nil)

	_, err := uc.AssignDriver(context.Background(), AssignInput{})
	assert.ErrorIs(t, err, ErrInvalidDelivery)
}

func TestCompleteDelivery_PublishesEventAndSendsEmail(t *testing.T) {
	repo := &mockDeliveryRepo{}
	mailer := &recordingMailer{}
	publisher := &recordingPublisher{}
	uc := NewDeliveryUsecase(repo, mailer, publisher, nil)
	ctx := context.Background()

	delivery := &model.Delivery{
		ID:         18,
		OrderID:    777,
		UserID:     4,
		UserEmail:  "user@example.com",
		DriverID:   2,
		DriverName: "Courier Two",
		Status:     model.StatusCompleted,
	}

	repo.On("CompleteDelivery", ctx, int64(18)).Return(delivery, true, nil)

	result, err := uc.CompleteDelivery(ctx, 18)
	require.NoError(t, err)
	assert.Equal(t, delivery, result)
	assert.Len(t, mailer.messages, 1)
	assert.Len(t, publisher.deliveries, 1)
	assert.Equal(t, int64(18), publisher.deliveries[0].ID)
	repo.AssertExpectations(t)
}

func TestHandleOrderPaid_SendsInTransitEmail(t *testing.T) {
	repo := &mockDeliveryRepo{}
	mailer := &recordingMailer{}
	uc := NewDeliveryUsecase(repo, mailer, nil, nil)
	ctx := context.Background()

	repo.On("MarkInTransitByOrderID", ctx, int64(700)).Return(&model.Delivery{
		ID:         21,
		OrderID:    700,
		UserEmail:  "user@example.com",
		DriverName: "Courier Three",
		Status:     model.StatusInTransit,
	}, true, nil)

	err := uc.HandleOrderPaid(ctx, model.OrderEvent{OrderID: 700})
	require.NoError(t, err)
	assert.Len(t, mailer.messages, 1)
	assert.Contains(t, mailer.messages[0].Subject, "on the way")
	repo.AssertExpectations(t)
}

func TestUpdateDriverLocation_RejectsInvalidCoordinates(t *testing.T) {
	repo := &mockDeliveryRepo{}
	uc := NewDeliveryUsecase(repo, nil, nil, nil)

	err := uc.UpdateDriverLocation(context.Background(), 10, 200, 20)
	assert.ErrorIs(t, err, ErrInvalidCoordinates)
}

func TestAssignDriver_DoesNotResendForRedelivery(t *testing.T) {
	repo := &mockDeliveryRepo{}
	mailer := &recordingMailer{}
	uc := NewDeliveryUsecase(repo, mailer, nil, nil)
	ctx := context.Background()

	repo.On("AssignDriver", ctx, int64(501), int64(9), "user@example.com", "Abay 10").Return(&model.Delivery{
		ID:         7,
		OrderID:    501,
		UserID:     9,
		UserEmail:  "user@example.com",
		DriverID:   3,
		DriverName: "Courier One",
		Status:     model.StatusAssigned,
	}, false, nil)

	_, err := uc.AssignDriver(ctx, AssignInput{
		OrderID:         501,
		UserID:          9,
		UserEmail:       "user@example.com",
		DeliveryAddress: "Abay 10",
	})
	require.NoError(t, err)
	assert.Empty(t, mailer.messages)
	repo.AssertExpectations(t)
}

func TestCompleteDelivery_DoesNotRepublishForDuplicateCompletion(t *testing.T) {
	repo := &mockDeliveryRepo{}
	mailer := &recordingMailer{}
	publisher := &recordingPublisher{}
	uc := NewDeliveryUsecase(repo, mailer, publisher, nil)
	ctx := context.Background()

	repo.On("CompleteDelivery", ctx, int64(18)).Return(&model.Delivery{
		ID:         18,
		OrderID:    777,
		UserID:     4,
		UserEmail:  "user@example.com",
		DriverID:   2,
		DriverName: "Courier Two",
		Status:     model.StatusCompleted,
	}, false, nil)

	_, err := uc.CompleteDelivery(ctx, 18)
	require.NoError(t, err)
	assert.Empty(t, mailer.messages)
	assert.Empty(t, publisher.deliveries)
	repo.AssertExpectations(t)
}

func TestTrackDelivery_PrefersRedisLocation(t *testing.T) {
	repo := &mockDeliveryRepo{}
	cache := &recordingLocationCache{
		locations: map[int64][2]float64{
			2: {43.238949, 76.889709},
		},
	}
	uc := NewDeliveryUsecase(repo, nil, nil, cache)
	ctx := context.Background()

	repo.On("GetByID", ctx, int64(18)).Return(&model.Delivery{
		ID:               18,
		OrderID:          777,
		UserID:           4,
		UserEmail:        "user@example.com",
		DriverID:         2,
		CurrentLatitude:  0,
		CurrentLongitude: 0,
	}, nil)

	delivery, err := uc.TrackDelivery(ctx, 18)
	require.NoError(t, err)
	assert.Equal(t, 43.238949, delivery.CurrentLatitude)
	assert.Equal(t, 76.889709, delivery.CurrentLongitude)
	repo.AssertExpectations(t)
}
