package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/model"
	"github.com/forever-free1/Autopilot/backend/internal/queue"
)

type conversationInput struct {
	VehicleID uint64 `json:"vehicle_id" binding:"required"`
	Title     string `json:"title" binding:"required,max=120"`
}
type messageInput struct {
	Content string `json:"content" binding:"required,min=2,max=2000"`
}
type diagnosticInput struct {
	VehicleID uint64 `json:"vehicle_id" binding:"required"`
	Symptom   string `json:"symptom" binding:"required,min=2,max=1000"`
}
type tripInput struct {
	VehicleID   uint64 `json:"vehicle_id" binding:"required"`
	Origin      string `json:"origin" binding:"required,max=160"`
	Destination string `json:"destination" binding:"required,max=160"`
}

func (h Handler) listConversations(c *gin.Context) {
	var items []model.Conversation
	h.DB.Where("user_id = ?", c.GetUint64(userIDKey)).Order("updated_at DESC").Find(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}
func (h Handler) createConversation(c *gin.Context) {
	var input conversationInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话参数无效"})
		return
	}
	var vehicle model.Vehicle
	if h.DB.Where("id = ? AND user_id = ?", input.VehicleID, c.GetUint64(userIDKey)).First(&vehicle).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在"})
		return
	}
	item := model.Conversation{ID: uuid.NewString(), UserID: c.GetUint64(userIDKey), VehicleID: input.VehicleID, Title: strings.TrimSpace(input.Title)}
	if h.DB.Create(&item).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}
	c.JSON(http.StatusCreated, item)
}
func (h Handler) getConversation(c *gin.Context) {
	var item model.Conversation
	if h.DB.Preload("Messages", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).Where("id = ? AND user_id = ?", c.Param("id"), c.GetUint64(userIDKey)).First(&item).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}
	c.JSON(http.StatusOK, item)
}
func (h Handler) createConversationMessage(c *gin.Context) {
	var input messageInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息内容无效"})
		return
	}
	var conversation model.Conversation
	if h.DB.Where("id = ? AND user_id = ?", c.Param("id"), c.GetUint64(userIDKey)).First(&conversation).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}
	task, err := h.publishWorkspaceTask(c.Request.Context(), c.GetUint64(userIDKey), conversation.VehicleID, input.Content)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	msg := model.ConversationMessage{ConversationID: conversation.ID, Role: "user", Content: input.Content, TaskID: &task.ID}
	h.DB.Create(&msg)
	h.DB.Model(&conversation).Update("updated_at", time.Now())
	c.JSON(http.StatusAccepted, gin.H{"message": msg, "task": task})
}
func (h Handler) createDiagnostic(c *gin.Context) {
	var input diagnosticInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "诊断参数无效"})
		return
	}
	command := "诊断车辆故障：" + strings.TrimSpace(input.Symptom)
	task, err := h.publishWorkspaceTask(c.Request.Context(), c.GetUint64(userIDKey), input.VehicleID, command)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, task)
}
func (h Handler) createTripPlan(c *gin.Context) {
	var input tripInput
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "行程参数无效"})
		return
	}
	var vehicle model.Vehicle
	if h.DB.Preload("Status").Where("id = ? AND user_id = ?", input.VehicleID, c.GetUint64(userIDKey)).First(&vehicle).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在"})
		return
	}
	// 第一版采用确定性估算，保证离线 Demo 可复现；后续可在此适配真实地图供应商。
	distance := 36.8
	energy := math.Round(distance/4.1*10) / 10
	rangeKM := vehicle.Status.Battery * 5.35
	remaining := math.Max(0, rangeKM-distance)
	points, _ := json.Marshal([]gin.H{{"type": "origin", "lat": 31.2042, "lng": 121.5891}, {"type": "charger", "lat": 31.2188, "lng": 121.4772}, {"type": "destination", "lat": 31.1969, "lng": 121.3271}})
	needCharge := remaining < 50
	advice := fmt.Sprintf("当前续航充足，无需中途充电。预计到达后剩余续航 %.0f km。", remaining)
	if needCharge {
		advice = "当前续航偏低，建议在推荐充电站补能后继续行程。"
	}
	plan := model.TripPlan{ID: uuid.NewString(), UserID: c.GetUint64(userIDKey), VehicleID: input.VehicleID, Origin: input.Origin, Destination: input.Destination, DistanceKM: distance, DurationMinute: 52, EnergyPercent: energy, RemainingRange: remaining, NeedCharge: needCharge, Advice: advice, WaypointsJSON: string(points)}
	if h.DB.Create(&plan).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存行程失败"})
		return
	}
	c.JSON(http.StatusCreated, plan)
}
func (h Handler) listTripPlans(c *gin.Context) {
	var items []model.TripPlan
	h.DB.Where("user_id = ?", c.GetUint64(userIDKey)).Order("created_at DESC").Limit(50).Find(&items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h Handler) publishWorkspaceTask(parent context.Context, userID, vehicleID uint64, command string) (model.AgentTask, error) {
	var count int64
	if h.DB.Model(&model.Vehicle{}).Where("id = ? AND user_id = ?", vehicleID, userID).Count(&count).Error != nil || count == 0 {
		return model.AgentTask{}, fmt.Errorf("车辆不存在")
	}
	task := model.AgentTask{ID: uuid.NewString(), UserID: userID, VehicleID: vehicleID, Command: command, Status: "pending"}
	if h.DB.Create(&task).Error != nil {
		return task, fmt.Errorf("创建任务失败")
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	message := queue.TaskMessage{TaskID: task.ID, VehicleID: vehicleID, Command: command, CallbackURL: h.CallbackURL + "/internal/agent/tasks/" + task.ID + "/events"}
	if h.Publisher.Publish(ctx, message) != nil {
		h.DB.Model(&task).Updates(map[string]any{"status": "failed", "error_message": "任务发布失败"})
		return task, fmt.Errorf("Agent 队列暂不可用")
	}
	return task, nil
}
