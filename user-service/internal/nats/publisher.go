package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, subject string, payload any) error
	Close()
}

type publisher struct {
	nc *nats.Conn
}

func NewPublisher(url string) (Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &publisher{nc: nc}, nil
}

func (p *publisher) Publish(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return p.nc.Publish(subject, data)
}

func (p *publisher) Close() {
	p.nc.Close()
}
