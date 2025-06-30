package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/user-service/internal/domain"
	pb "github.com/popeskul/mailflow/user-service/pkg/api/user/v1"
)

type Services interface {
	User() domain.UserService
}

type UserServer struct {
	pb.UnimplementedUserServiceServer
	userService domain.UserService
	logger      logger.Logger
}

func NewUserServer(userService Services, logger logger.Logger) *UserServer {
	return &UserServer{
		userService: userService.User(),
		logger:      logger.Named("user_server"),
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
