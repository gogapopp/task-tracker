package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type SMTPClient struct {
	logger    *zap.SugaredLogger
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string

	client *smtp.Client
	mu     sync.Mutex
}

func NewSMTPClient(logger *zap.SugaredLogger, host string, port int, username, password, fromEmail, fromName string) (*SMTPClient, error) {
	client := &SMTPClient{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
		logger:    logger,
	}

	err := client.connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *SMTPClient) connect() error {
	const op = "internal.service.email_sender.connect"

	address := fmt.Sprintf("%s:%d", s.host, s.port)
	client, err := smtp.Dial(address)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		client.Close()
		return fmt.Errorf("%s: %w", op, err)
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	if err = client.Auth(auth); err != nil {
		client.Close()
		return fmt.Errorf("%s: auth problem: %w", op, err)
	}

	s.client = client
	return nil
}

func (s *SMTPClient) ensureConnected() error {
	if s.client == nil {
		return s.connect()
	}

	if err := s.client.Noop(); err != nil {
		s.logger.Infof("SMTP connection lost, reconnecting...")
		s.client.Close()
		return s.connect()
	}

	return nil
}

func (s *SMTPClient) SendEmail(ctx context.Context, to, toName, subject, textPart, htmlPart string) error {
	const op = "internal.service.email_sender.SendEmail"

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureConnected(); err != nil {
		return fmt.Errorf("%s: connection error: %w", op, err)
	}

	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `multipart/alternative; boundary="BOUNDARY"`,
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n--BOUNDARY\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	msg.WriteString(textPart)
	msg.WriteString("\r\n--BOUNDARY\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
	msg.WriteString(htmlPart)
	msg.WriteString("\r\n--BOUNDARY--")

	s.logger.Infof("sending mail to %s (%s) topic: %s", to, toName, subject)

	// mailing...
	if err := s.client.Mail(s.fromEmail); err != nil {
		// try to reconnect
		s.client.Close()
		if err := s.connect(); err != nil {
			return fmt.Errorf("%s: reconnection failed: %w", op, err)
		}
		if err := s.client.Mail(s.fromEmail); err != nil {
			return fmt.Errorf("%s: sender problem after reconnection: %w", op, err)
		}
	}

	if err := s.client.Rcpt(to); err != nil {
		return fmt.Errorf("%s: recipient problem: %w", op, err)
	}

	w, err := s.client.Data()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = w.Write([]byte(msg.String()))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Infof("mail successfully sent to %s", to)
	return nil
}

func (s *SMTPClient) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		err := s.client.Quit()
		s.client = nil
		return err
	}
	return nil
}
