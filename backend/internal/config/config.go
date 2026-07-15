package config

import "os"

// Config 保存服务运行所需配置，所有字段均可通过环境变量覆盖。
type Config struct {
	Port          string
	MySQLDSN      string
	RedisAddr     string
	JWTSecret     string
	RabbitMQURL   string
	InternalToken string
	CallbackURL   string
}

// Load 读取配置并提供仅适用于本地开发的默认值。
func Load() Config {
	return Config{
		Port:          env("BACKEND_PORT", "8080"),
		MySQLDSN:      env("MYSQL_DSN", "autopilot:autopilot_dev@tcp(localhost:3306)/autopilot?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		JWTSecret:     env("JWT_SECRET", "local-development-secret-change-me"),
		RabbitMQURL:   env("RABBITMQ_URL", "amqp://autopilot:autopilot_dev@localhost:5672/"),
		InternalToken: env("INTERNAL_TOKEN", "local-agent-callback-token"),
		CallbackURL:   env("BACKEND_CALLBACK_URL", "http://localhost:8080"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
