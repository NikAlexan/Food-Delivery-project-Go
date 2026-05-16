package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

var (
	ErrInvalidSMTPConfig = errors.New("invalid smtp configuration")
	ErrInvalidMessage    = errors.New("invalid smtp message")
)

type SMTPClient struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPClient(host, user, pass, from string, port int) (*SMTPClient, error) {
	switch {
	case host == "":
		return nil, fmt.Errorf("%w: missing host", ErrInvalidSMTPConfig)
	case port <= 0:
		return nil, fmt.Errorf("%w: invalid port", ErrInvalidSMTPConfig)
	case from == "":
		return nil, fmt.Errorf("%w: missing sender", ErrInvalidSMTPConfig)
	}

	return &SMTPClient{
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}, nil
}

func (c *SMTPClient) Send(ctx context.Context, msg Message) error {
	if c == nil {
		return ErrInvalidSMTPConfig
	}
	if msg.To == "" {
		return fmt.Errorf("%w: missing recipient", ErrInvalidMessage)
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	var auth smtp.Auth
	if c.user != "" && c.pass != "" {
		auth = smtp.PlainAuth("", c.user, c.pass, c.host)
	}

	payload := c.buildMessage(msg)
	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(addr, auth, c.from, []string{msg.To}, payload)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (c *SMTPClient) buildMessage(msg Message) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", c.from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", msg.To))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)
	return []byte(b.String())
}
