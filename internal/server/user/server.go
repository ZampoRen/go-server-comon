package user

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "go-server-comon/api/proto/user"
)

// Server implements the User gRPC server
type Server struct {
	pb.UnimplementedUserServer
	mu     sync.RWMutex
	users  map[int64]*pb.UserInfo
	nextID int64
}

// NewServer creates a new User server instance
func NewServer() *Server {
	return &Server{
		users:  make(map[int64]*pb.UserInfo),
		nextID: 1,
	}
}

// GetUser 获取用户信息
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user_id must be greater than 0")
	}

	s.mu.RLock()
	user, exists := s.users[req.UserId]
	s.mu.RUnlock()

	if !exists {
		return nil, status.Errorf(codes.NotFound, "user with id %d not found", req.UserId)
	}

	return &pb.GetUserResponse{
		UserId:    user.UserId,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// CreateUser 创建用户
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Username == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username is required")
	}
	if req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查用户名是否已存在
	for _, user := range s.users {
		if user.Username == req.Username {
			return nil, status.Errorf(codes.AlreadyExists, "username %s already exists", req.Username)
		}
		if user.Email == req.Email {
			return nil, status.Errorf(codes.AlreadyExists, "email %s already exists", req.Email)
		}
	}

	// 创建新用户
	userID := s.nextID
	s.nextID++

	newUser := &pb.UserInfo{
		UserId:   userID,
		Username: req.Username,
		Email:    req.Email,
	}

	s.users[userID] = newUser

	return &pb.CreateUserResponse{
		UserId:   userID,
		Username: req.Username,
		Email:    req.Email,
		Success:  true,
	}, nil
}

// ListUsers 获取用户列表
func (s *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	s.mu.RLock()
	allUsers := make([]*pb.UserInfo, 0, len(s.users))
	for _, user := range s.users {
		allUsers = append(allUsers, user)
	}
	total := int32(len(allUsers))
	s.mu.RUnlock()

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= int32(len(allUsers)) {
		return &pb.ListUsersResponse{
			Users: []*pb.UserInfo{},
			Total: total,
			Page:  page,
		}, nil
	}

	if end > int32(len(allUsers)) {
		end = int32(len(allUsers))
	}

	return &pb.ListUsersResponse{
		Users: allUsers[start:end],
		Total: total,
		Page:  page,
	}, nil
}
