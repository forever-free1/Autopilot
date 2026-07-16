package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/model"
	"github.com/forever-free1/Autopilot/backend/internal/queue"
)

type taskInput struct {
	VehicleID uint64 `json:"vehicle_id" binding:"required"`
	Command   string `json:"command" binding:"required,min=2,max=500"`
}
type toolCallEvent struct {
	ToolName  string         `json:"tool_name"`
	Input     map[string]any `json:"input"`
	Output    map[string]any `json:"output"`
	LatencyMS int64          `json:"latency_ms"`
	Success   bool           `json:"success"`
}
type taskEvent struct {
	Status       string          `json:"status"`
	Result       string          `json:"result"`
	ErrorMessage string          `json:"error_message"`
	VehiclePatch map[string]any  `json:"vehicle_patch"`
	ToolCalls    []toolCallEvent `json:"tool_calls"`
}

func (h Handler) createTask(c *gin.Context) {
	var input taskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "车辆或指令格式不正确"})
		return
	}
	var count int64
	if err := h.DB.Model(&model.Vehicle{}).Where("id = ? AND user_id = ?", input.VehicleID, c.GetUint64(userIDKey)).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "车辆不存在"})
		return
	}
	task := model.AgentTask{ID: uuid.NewString(), UserID: c.GetUint64(userIDKey), VehicleID: input.VehicleID, Command: input.Command, Status: "pending"}
	if err := h.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	message := queue.TaskMessage{TaskID: task.ID, VehicleID: task.VehicleID, Command: task.Command, CallbackURL: h.CallbackURL + "/internal/agent/tasks/" + task.ID + "/events"}
	if err := h.Publisher.Publish(ctx, message); err != nil {
		_ = h.DB.Model(&task).Updates(map[string]any{"status": "failed", "error_message": "任务发布失败"}).Error
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent 队列暂不可用"})
		return
	}
	c.JSON(http.StatusAccepted, task)
}

func (h Handler) getTask(c *gin.Context) {
	var task model.AgentTask
	if err := h.DB.Preload("ToolCalls").Where("id = ? AND user_id = ?", c.Param("id"), c.GetUint64(userIDKey)).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h Handler) receiveTaskEvent(c *gin.Context) {
	if c.GetHeader("X-Internal-Token") != h.InternalToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "内部凭据无效"})
		return
	}
	var event taskEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "事件格式错误"})
		return
	}
	var vehicleID uint64
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var task model.AgentTask
		if err := tx.First(&task, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		vehicleID = task.VehicleID
		updates := map[string]any{"status": event.Status, "result": event.Result, "error_message": event.ErrorMessage}
		now := time.Now()
		if event.Status == "running" {
			updates["started_at"] = now
		}
		if event.Status == "finished" || event.Status == "failed" {
			updates["finished_at"] = now
		}
		if err := tx.Model(&task).Updates(updates).Error; err != nil {
			return err
		}
		for _, call := range event.ToolCalls {
			input, _ := json.Marshal(call.Input)
			output, _ := json.Marshal(call.Output)
			trace := model.ToolCall{TaskID: task.ID, ToolName: call.ToolName, Input: string(input), Output: string(output), LatencyMS: call.LatencyMS, Success: call.Success}
			if err := tx.Create(&trace).Error; err != nil {
				return err
			}
		}
		// 若任务来自会话消息，在终态回调中同步保存 Agent 可见回复，确保历史会话可完整回放。
		if event.Status == "finished" || event.Status == "failed" {
			var source model.ConversationMessage
			if err := tx.Where("task_id = ? AND role = ?", task.ID, "user").First(&source).Error; err == nil {
				var count int64
				tx.Model(&model.ConversationMessage{}).Where("task_id = ? AND role = ?", task.ID, "assistant").Count(&count)
				if count == 0 {
					content := event.Result
					if event.Status == "failed" { content = event.ErrorMessage }
					reply := model.ConversationMessage{ConversationID: source.ConversationID, Role: "assistant", Content: content, TaskID: &task.ID}
					if err := tx.Create(&reply).Error; err != nil { return err }
					tx.Model(&model.Conversation{}).Where("id = ?", source.ConversationID).Update("updated_at", now)
				}
			}
		}
		// 工具结果与任务状态同事务提交，避免 Trace 成功但车辆状态未更新。
		if len(event.VehiclePatch) > 0 {
			return tx.Model(&model.VehicleStatus{}).Where("vehicle_id = ?", task.VehicleID).Updates(event.VehiclePatch).Error
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务事件处理失败"})
		return
	}
	if len(event.VehiclePatch) > 0 {
		_ = h.Cache.Del(c.Request.Context(), fmt.Sprintf("vehicle:%d:status", vehicleID)).Err()
	}
	c.Status(http.StatusNoContent)
}
