package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/forever-free1/Autopilot/backend/internal/model"
)

type statusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// getDashboardSummary 在 Go 层聚合运营指标，避免前端直接理解数据库结构。
func (h Handler) getDashboardSummary(c *gin.Context) {
	userID := c.GetUint64(userIDKey)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var vehicles, tasksToday, completed, failed, toolCalls int64
	var averageLatency float64
	var statuses []statusCount

	h.DB.Model(&model.Vehicle{}).Where("user_id = ?", userID).Count(&vehicles)
	base := h.DB.Model(&model.AgentTask{}).Where("user_id = ? AND created_at >= ?", userID, startOfDay)
	base.Count(&tasksToday)
	base.Where("status = ?", "finished").Count(&completed)
	base.Where("status = ?", "failed").Count(&failed)
	h.DB.Model(&model.AgentTask{}).Select("status, count(*) AS count").Where("user_id = ? AND created_at >= ?", userID, startOfDay).Group("status").Scan(&statuses)
	h.DB.Model(&model.ToolCall{}).Joins("JOIN agent_tasks ON agent_tasks.id = tool_calls.task_id").Where("agent_tasks.user_id = ? AND tool_calls.created_at >= ?", userID, startOfDay).Count(&toolCalls)
	h.DB.Model(&model.AgentTask{}).Where("user_id = ? AND finished_at IS NOT NULL AND started_at IS NOT NULL", userID).
		Select("COALESCE(AVG(TIMESTAMPDIFF(MICROSECOND, started_at, finished_at) / 1000), 0)").Scan(&averageLatency)

	successRate := float64(0)
	if completed+failed > 0 {
		successRate = float64(completed) / float64(completed+failed) * 100
	}
	c.JSON(http.StatusOK, gin.H{
		"online_vehicles": vehicles, "tasks_today": tasksToday, "success_rate": successRate,
		"average_response_ms": averageLatency, "fault_vehicles": 0, "rabbitmq_backlog": 0,
		"tool_calls_today": toolCalls, "task_statuses": statuses, "generated_at": now,
		"queue_metric_source": "not_configured",
	})
}

// listTasks 提供稳定的分页任务列表，支持状态过滤和关键词检索。
func (h Handler) listTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := h.DB.Model(&model.AgentTask{}).Where("user_id = ?", c.GetUint64(userIDKey))
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := c.Query("q"); keyword != "" {
		query = query.Where("id LIKE ? OR command LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	var tasks []model.AgentTask
	query.Count(&total)
	if err := query.Preload("ToolCalls").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tasks, "total": total, "page": page, "page_size": pageSize})
}

// listToolCalls 只返回当前用户任务关联的工具调用，防止跨用户读取 Trace。
func (h Handler) listToolCalls(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var calls []model.ToolCall
	err := h.DB.Joins("JOIN agent_tasks ON agent_tasks.id = tool_calls.task_id").
		Where("agent_tasks.user_id = ?", c.GetUint64(userIDKey)).Order("tool_calls.created_at DESC").Limit(limit).Find(&calls).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询工具调用失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": calls})
}

// streamTaskEvents 以 SSE 推送任务快照。数据库是事实来源，断线重连不会丢失最终状态。
func (h Handler) streamTaskEvents(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	lastVersion := ""
	for {
		var task model.AgentTask
		err := h.DB.Preload("ToolCalls").Where("id = ? AND user_id = ?", c.Param("id"), c.GetUint64(userIDKey)).First(&task).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.SSEvent("error", gin.H{"error": "任务不存在"})
			}
			return
		}
		version := fmt.Sprintf("%s:%d:%s", task.UpdatedAt.Format(time.RFC3339Nano), len(task.ToolCalls), task.Status)
		if version != lastVersion {
			c.SSEvent("task", task)
			c.Writer.Flush()
			lastVersion = version
		}
		if task.Status == "finished" || task.Status == "failed" || task.Status == "cancelled" {
			c.SSEvent("done", gin.H{"status": task.Status})
			c.Writer.Flush()
			return
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
