package email

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMessage(t *testing.T) {
	client, err := NewSMTPClient("smtp.example.com", "user", "pass", "delivery@example.com", 587)
	assert.NoError(t, err)
	msg := client.buildMessage(Message{
		To:      "user@example.com",
		Subject: "Order delivered",
		Body:    "Your order has been delivered.",
	})

	text := string(msg)
	assert.True(t, strings.Contains(text, "From: delivery@example.com"))
	assert.True(t, strings.Contains(text, "To: user@example.com"))
	assert.True(t, strings.Contains(text, "Subject: Order delivered"))
	assert.True(t, strings.Contains(text, "Your order has been delivered."))
}

func TestNewSMTPClient_ValidatesRequiredConfig(t *testing.T) {
	client, err := NewSMTPClient("", "user", "pass", "delivery@example.com", 587)
	assert.Nil(t, client)
	assert.ErrorIs(t, err, ErrInvalidSMTPConfig)
}
