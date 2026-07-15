package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/auth"
	"github.com/forever-free1/Autopilot/backend/internal/model"
	"github.com/forever-free1/Autopilot/backend/internal/queue"
)

type Handler struct {
	DB            *gorm.DB
	Cache         *redis.Client
	JWTSecret     string
	InternalToken string
	CallbackURL   string
	Publisher     *queue.Publisher
}
type credentials struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required"`
}
type vehicleInput struct {
	VehicleModel string `json:"vehicle_model" binding:"required,max=80"`
	VIN          string `json:"vin" binding:"required,len=17"`
}

func (h Handler) register(c *gin.Context) {
	var input credentials
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码格式不正确"})
		return
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := model.User{Username: strings.TrimSpace(input.Username), PasswordHash: hash}
	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}
	token, _ := auth.CreateToken(user.ID, h.JWTSecret)
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": user})
}

func (h Handler) login(c *gin.Context) {
	var input credentials
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码格式不正确"})
		return
	}
	var user model.User
	if err := h.DB.Where("username = ?", input.Username).First(&user).Error; err != nil || !auth.CheckPassword(user.PasswordHash, input.Password) {
		// 登录失败统一返回相同提示，避免攻击者探测用户名是否存在。
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	token, _ := auth.CreateToken(user.ID, h.JWTSecret)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (h Handler) createVehicle(c *gin.Context) {
	var input vehicleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "车型或 VIN 格式不正确"})
		return
	}
	vehicle := model.Vehicle{UserID: c.GetUint64(userIDKey), VehicleModel: input.VehicleModel, VIN: strings.ToUpper(input.VIN), Status: model.VehicleStatus{Battery: 82, Temperature: 24, Location: "上海"}}
	if err := h.DB.Create(&vehicle).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "VIN 已存在"})
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}

func (h Handler) listVehicles(c *gin.Context) {
	var vehicles []model.Vehicle
	if err := h.DB.Preload("Status").Where("user_id = ?", c.GetUint64(userIDKey)).Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询车辆失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": vehicles})
}

func (h Handler) getVehicleStatus(c *gin.Context) {
	vehicleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "车辆 ID 无效"})
		return
	}
	var vehicle model.Vehicle
	if err := h.DB.Where("id = ? AND user_id = ?", vehicleID, c.GetUint64(userIDKey)).First(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询车辆失败"})
		return
	}

	key := fmt.Sprintf("vehicle:%d:status", vehicleID)
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
	defer cancel()
	if raw, err := h.Cache.Get(ctx, key).Bytes(); err == nil {
		var status model.VehicleStatus
		if json.Unmarshal(raw, &status) == nil {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, status)
			return
		}
	}
	var status model.VehicleStatus
	if err := h.DB.First(&status, "vehicle_id = ?", vehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "车辆状态不存在"})
		return
	}
	if raw, err := json.Marshal(status); err == nil {
		_ = h.Cache.Set(ctx, key, raw, 30*time.Second).Err()
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, status)
}
