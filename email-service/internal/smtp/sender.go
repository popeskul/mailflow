package smtp

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/config"
	"github.com/popeskul/mailflow/email-service/internal/domain"
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

func NewSMTPSender(config config.SMTPConfig, logger logger.Logger) *Sender {
	return &Sender{
		enabled:  config.Enabled,
		host:     config.Host,
		port:     config.Port,
		username: config.Username,
		password: config.Password,
		from:     config.SenderEmail,
		logger:   logger.Named("smtp_sender"),
	}
}

func (s *Sender) Send(ctx context.Context, email *domain.Email) error {
	l := s.logger.WithFields(logger.Fields{
		"email_id": email.ID,
		"to":       email.To,
	})

	if !s.enabled {
		l.Info("email sending skipped (SMTP disabled)",
			logger.Field{Key: "subject", Value: email.Subject},
			logger.Field{Key: "body", Value: email.Body},
		)
		return nil
	}

	l.Debug("preparing to send email",
		logger.Field{Key: "smtp_host", Value: s.host},
		logger.Field{Key: "smtp_port", Value: s.port},
	)

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, email.To, email.Subject, email.Body)

	addr := s.host + ":" + s.port
	if err := smtp.SendMail(addr, auth, s.from, []string{email.To}, []byte(msg)); err != nil {
		l.Error("failed to send email",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "smtp_addr", Value: addr},
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	l.Info("email sent successfully")
	return nil
}
