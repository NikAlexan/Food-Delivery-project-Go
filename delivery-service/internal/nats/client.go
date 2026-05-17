package nats

import (
	"context"
	"encoding/json"
	"log"

	natsgo "github.com/nats-io/nats.go"

	"food-delivery/delivery-service/internal/model"
)

type OrderHandler interface {
	HandleOrderCreated(ctx context.Context, event model.OrderEvent) error
	HandleOrderPaid(ctx context.Context, event model.OrderEvent) error
	HandleOrderCancelled(ctx context.Context, event model.OrderEvent) error
}

type Client struct {
	nc *natsgo.Conn
	js natsgo.JetStreamContext
}

func NewClient(url string) (*Client, error) {
	if url == "" {
		return &Client{}, nil
	}

	nc, err := natsgo.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	return &Client{nc: nc, js: js}, nil
}

func (c *Client) StartOrderConsumers(handler OrderHandler) error {
	if c == nil || c.js == nil {
		return nil
	}

	if err := c.ensureStream("ORDER_EVENTS", []string{"order.*"}); err != nil {
		return err
	}

	subscriptions := []struct {
		subject string
		durable string
		handle  func(context.Context, model.OrderEvent) error
	}{
		{subject: "order.created", durable: "delivery-order-created", handle: handler.HandleOrderCreated},
		{subject: "order.paid", durable: "delivery-order-paid", handle: handler.HandleOrderPaid},
		{subject: "order.cancelled", durable: "delivery-order-cancelled", handle: handler.HandleOrderCancelled},
	}

	for _, sub := range subscriptions {
		if _, err := c.js.QueueSubscribe(
			sub.subject,
			"delivery-workers",
			c.wrapHandler(sub.subject, sub.handle),
			natsgo.Durable(sub.durable),
			natsgo.ManualAck(),
			natsgo.AckExplicit(),
		); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) PublishDeliveryCompleted(ctx context.Context, delivery *model.Delivery) error {
	if c == nil || c.js == nil || delivery == nil {
		return nil
	}

	if err := c.ensureStream("DELIVERY_EVENTS", []string{"delivery.*"}); err != nil {
		return err
	}

	payload, err := json.Marshal(model.DeliveryCompletedEvent{
		DeliveryID:  delivery.ID,
		OrderID:     delivery.OrderID,
		UserID:      delivery.UserID,
		DriverID:    delivery.DriverID,
		CompletedAt: delivery.UpdatedAt,
	})
	if err != nil {
		return err
	}

	_, err = c.js.Publish("delivery.completed", payload)
	return err
}

func (c *Client) Close() error {
	if c == nil || c.nc == nil {
		return nil
	}
	if err := c.nc.Drain(); err != nil {
		c.nc.Close()
		return err
	}
	c.nc.Close()
	return nil
}

func (c *Client) wrapHandler(subject string, handle func(context.Context, model.OrderEvent) error) natsgo.MsgHandler {
	return func(msg *natsgo.Msg) {
		var event model.OrderEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("nats %s decode failed: %v", subject, err)
			_ = msg.Ack()
			return
		}

		if err := handle(context.Background(), event); err != nil {
			log.Printf("nats %s handler failed: %v", subject, err)
			_ = msg.Nak()
			return
		}

		_ = msg.Ack()
	}
}

func (c *Client) ensureStream(name string, subjects []string) error {
	if c == nil || c.js == nil {
		return nil
	}

	if _, err := c.js.StreamInfo(name); err == nil {
		return nil
	}

	_, err := c.js.AddStream(&natsgo.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  natsgo.FileStorage,
	})
	return err
}
