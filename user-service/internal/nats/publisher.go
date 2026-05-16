package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const streamName = "USER_EVENTS"

var streamSubjects = []string{"user.*"}

type Publisher interface {
	Publish(ctx context.Context, subject string, payload any) error
	Close()
}

type publisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewPublisher(url string) (Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	if err := ensureStream(js, streamName, streamSubjects); err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats ensure stream: %w", err)
	}

	return &publisher{nc: nc, js: js}, nil
}

func (p *publisher) Publish(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = p.js.Publish(subject, data)
	return err
}

func (p *publisher) Close() {
	p.nc.Drain()
	p.nc.Close()
}

func ensureStream(js nats.JetStreamContext, name string, subjects []string) error {
	if _, err := js.StreamInfo(name); err == nil {
		return nil
	}
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
	})
	return err
}
