package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/queue"
)

// NewRouter 创建 HTTP 路由。后续业务模块统一挂载在 /api/v1 下，避免接口演进时破坏客户端。
func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "backend"})
	})

	return router
}

// NewApplicationRouter 挂载需要数据库和缓存的业务接口。
func NewApplicationRouter(db *gorm.DB, cache *redis.Client, jwtSecret, internalToken, callbackURL string, publisher *queue.Publisher) *gin.Engine {
	router := NewRouter()
	h := Handler{DB: db, Cache: cache, JWTSecret: jwtSecret, InternalToken: internalToken, CallbackURL: callbackURL, Publisher: publisher}
	router.POST("/internal/agent/tasks/:id/events", h.receiveTaskEvent)
	api := router.Group("/api/v1")
	api.POST("/auth/register", h.register)
	api.POST("/auth/login", h.login)
	authorized := api.Group("")
	authorized.Use(authenticate(jwtSecret))
	authorized.POST("/vehicles", h.createVehicle)
	authorized.GET("/vehicles", h.listVehicles)
	authorized.GET("/vehicles/:id", h.getVehicle)
	authorized.GET("/vehicles/:id/status", h.getVehicleStatus)
	authorized.GET("/dashboard/summary", h.getDashboardSummary)
	authorized.POST("/agent/tasks", h.createTask)
	authorized.GET("/agent/tasks", h.listTasks)
	authorized.GET("/agent/tasks/:id", h.getTask)
	authorized.GET("/agent/tasks/:id/events", h.streamTaskEvents)
	authorized.GET("/tool-calls", h.listToolCalls)
	return router
}
