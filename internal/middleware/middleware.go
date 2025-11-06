package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// TODO: Add gRPC interceptor/middleware implementations

// UnaryServerInterceptor example
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: Implement middleware logic
		return handler(ctx, req)
	}
}
