package smtp

import (
	"context"
	"fmt"
	"github.com/popeskul/email-service-platform/logger"
	"net/smtp"

	"go.uber.org/zap"

	"github.com/popeskul/email-service-platform/email-service/internal/domain"
)

type Sender struct {
	enabled  bool
	host     string
	port     string
	username string
	password string
	from     string
	logger   logger.Logger
}

func NewSMTPSender(enabled bool, host, port, username, password, from string, logger logger.Logger) *Sender {
	return &Sender{
		enabled:  enabled,
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		logger:   logger.Named("smtp_sender"),
	}
}

func (s *Sender) Send(ctx context.Context, email *domain.Email) error {
	logger := s.logger.With(
		zap.String("email_id", email.ID),
		zap.String("to", email.To),
	)

	if !s.enabled {
		logger.Info("email sending skipped (SMTP disabled)",
			zap.String("subject", email.Subject),
			zap.String("body", email.Body),
		)
		return nil
	}

	logger.Debug("preparing to send email",
		zap.String("smtp_host", s.host),
		zap.String("smtp_port", s.port),
	)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, email.To, email.Subject, email.Body)

	addr := s.host + ":" + s.port
	if err := smtp.SendMail(addr, auth, s.from, []string{email.To}, []byte(msg)); err != nil {
		logger.Error("failed to send email",
			zap.Error(err),
			zap.String("smtp_addr", addr),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	logger.Info("email sent successfully")
	return nil
}
