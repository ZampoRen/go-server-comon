package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/ZampoRen/go-server-comon/api/router"
)

func main() {
	// 创建 Hertz 服务器
	h := server.Default(
		server.WithHostPorts(":8888"),
		server.WithHandleMethodNotAllowed(true),
	)

	// 注册路由（使用 hz 生成的路由注册函数）
	router.GeneratedRegister(h)

	// 启动服务器（在 goroutine 中）
	go func() {
		h.Spin()
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	hlog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		hlog.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
