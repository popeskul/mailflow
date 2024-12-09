package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/email-service-platform/user-service/internal/domain"
	"github.com/popeskul/email-service-platform/user-service/internal/ports"
	pb "github.com/popeskul/email-service-platform/user-service/pkg/api/user/v1"
)

type Service interface {
	UserService() ports.UserService
}

type UserServer struct {
	pb.UnimplementedUserServiceServer
	userService ports.UserService
	logger      *zap.Logger
}

func NewUserServer(userService Service, logger *zap.Logger) *UserServer {
	return &UserServer{
		userService: userService.UserService(),
		logger:      logger,
	}
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.GetEmail() == "" || req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and username are required")
	}

	user, err := s.userService.Create(ctx, req.GetEmail(), req.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &pb.CreateUserResponse{
		Id:   user.ID,
		User: toProtoUser(user),
	}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	user, err := s.userService.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserResponse{
		User: toProtoUser(user),
	}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, nextPageToken, err := s.userService.List(ctx, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	var protoUsers []*pb.User
	for _, user := range users {
		protoUsers = append(protoUsers, toProtoUser(user))
	}

	return &pb.ListUsersResponse{
		Users:         protoUsers,
		NextPageToken: nextPageToken,
	}, nil
}

func toProtoUser(user *domain.User) *pb.User {
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
