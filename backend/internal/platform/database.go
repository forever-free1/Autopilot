package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/config"
	"github.com/forever-free1/Autopilot/backend/internal/model"
)

// Connect 建立基础设施连接并执行轻量自动迁移，便于作品项目一键复现。
func Connect(cfg config.Config) (*gorm.DB, *redis.Client, error) {
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("连接 MySQL: %w", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Vehicle{}, &model.VehicleStatus{}, &model.AgentTask{}, &model.ToolCall{}, &model.Conversation{}, &model.ConversationMessage{}, &model.TripPlan{}); err != nil {
		return nil, nil, fmt.Errorf("迁移数据库: %w", err)
	}

	cache := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cache.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("连接 Redis: %w", err)
	}
	return db, cache, nil
}
