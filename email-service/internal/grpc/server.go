package grpc

import (
	"context"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/domain"
	"github.com/popeskul/mailflow/email-service/internal/metrics"
	"github.com/popeskul/mailflow/email-service/internal/services"
	pb "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
)

type EmailServer struct {
	pb.UnimplementedEmailServiceServer
	emailService services.EmailService
	metrics      *metrics.EmailMetrics
	logger       logger.Logger
	isDown       int32 // atomic
}

func NewEmailServer(emailService services.EmailService, metrics *metrics.EmailMetrics, l logger.Logger) *EmailServer {
	return &EmailServer{
		emailService: emailService,
		metrics:      metrics,
		logger:       l.Named("email_server"),
	}
}

func (s *EmailServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	if atomic.LoadInt32(&s.isDown) == 1 {
		return nil, status.Error(codes.Unavailable, "service is in maintenance mode")
	}

	if err := validateSendEmailRequest(req); err != nil {
		return nil, err
	}

	start := time.Now()
	email, err := s.emailService.SendEmail(ctx, req.To, req.Subject, req.Body)
	s.metrics.ObserveProcessingDuration(time.Since(start).Seconds())

	if err != nil {
		s.logger.Error("failed to send email",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "to", Value: req.To},
		)
		s.metrics.RecordEmailFailed()
		return nil, status.Error(codes.Internal, "failed to send email")
	}

	s.metrics.RecordEmailSent()
	return &pb.SendEmailResponse{
		Id:     email.ID,
		Status: email.Status,
	}, nil
}

func (s *EmailServer) GetEmailStatus(ctx context.Context, req *pb.GetEmailStatusRequest) (*pb.GetEmailStatusResponse, error) {
	if atomic.LoadInt32(&s.isDown) == 1 {
		return nil, status.Error(codes.Unavailable, "service is in maintenance mode")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "email id is required")
	}

	email, err := s.emailService.GetEmailStatus(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get email status",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "email_id", Value: req.Id},
		)
		return nil, status.Error(codes.NotFound, "email not found")
	}

	var sentAt string
	if email.SentAt != nil {
		sentAt = email.SentAt.Format(time.RFC3339)
	}

	return &pb.GetEmailStatusResponse{
		Id:     email.ID,
		Status: email.Status,
		SentAt: sentAt,
	}, nil
}

func (s *EmailServer) ListEmails(ctx context.Context, req *pb.ListEmailsRequest) (*pb.ListEmailsResponse, error) {
	if atomic.LoadInt32(&s.isDown) == 1 {
		return nil, status.Error(codes.Unavailable, "service is in maintenance mode")
	}

	emails, nextPageToken, err := s.emailService.ListEmails(ctx, int(req.PageSize), req.PageToken)
	if err != nil {
		s.logger.Error("failed to list emails",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "page_size", Value: req.PageSize},
			logger.Field{Key: "page_token", Value: req.PageToken},
		)
		return nil, status.Error(codes.Internal, "failed to list emails")
	}

	var protoEmails []*pb.Email
	for _, email := range emails {
		protoEmails = append(protoEmails, toProtoEmail(email))
	}

	return &pb.ListEmailsResponse{
		Emails:        protoEmails,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *EmailServer) SetDowntime(isDown bool) {
	if isDown {
		atomic.StoreInt32(&s.isDown, 1)
		s.metrics.RecordDowntimePeriod()
		s.logger.Info("service entering maintenance mode")
	} else {
		atomic.StoreInt32(&s.isDown, 0)
		s.logger.Info("service exiting maintenance mode")
	}
}

func validateSendEmailRequest(req *pb.SendEmailRequest) error {
	if req.To == "" {
		return status.Error(codes.InvalidArgument, "recipient email is required")
	}
	if req.Subject == "" {
		return status.Error(codes.InvalidArgument, "subject is required")
	}
	if req.Body == "" {
		return status.Error(codes.InvalidArgument, "body is required")
	}
	return nil
}

func toProtoEmail(email *domain.Email) *pb.Email {
	result := &pb.Email{
		Id:        email.ID,
		To:        email.To,
		Subject:   email.Subject,
		Body:      email.Body,
		Status:    email.Status,
		CreatedAt: email.CreatedAt.Format(time.RFC3339),
	}

	if email.SentAt != nil {
		result.SentAt = email.SentAt.Format(time.RFC3339)
	}

	return result
}
