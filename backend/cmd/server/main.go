package main

import (
	"log"

	"github.com/forever-free1/Autopilot/backend/internal/config"
	"github.com/forever-free1/Autopilot/backend/internal/httpapi"
	"github.com/forever-free1/Autopilot/backend/internal/platform"
	"github.com/forever-free1/Autopilot/backend/internal/queue"
)

func main() {
	cfg := config.Load()
	db, cache, err := platform.Connect(cfg)
	if err != nil {
		log.Fatalf("初始化基础设施失败: %v", err)
	}
	defer cache.Close()
	publisher, err := queue.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("初始化消息队列失败: %v", err)
	}
	defer publisher.Close()

	router := httpapi.NewApplicationRouter(db, cache, cfg.JWTSecret, cfg.InternalToken, cfg.CallbackURL, publisher)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("启动 Go 服务失败: %v", err)
	}
}
