package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "go-server-comon/api/proto/user"
	"go-server-comon/internal/server/user"
)

func main() {
	// 创建监听器
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 创建 gRPC 服务器
	srv := grpc.NewServer()

	// 创建并注册 User 服务
	userServer := user.NewServer()
	pb.RegisterUserServer(srv, userServer)

	// 启用 gRPC 反射服务（用于 grpcurl 等工具）
	reflection.Register(srv)

	// 启动服务器（在 goroutine 中）
	go func() {
		log.Println("User service starting on :50051")
		if err := srv.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	srv.GracefulStop()
	log.Println("Server stopped")
}
