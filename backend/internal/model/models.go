package model

import "time"

// User 是平台用户。密码只保存不可逆哈希，不持久化明文。
type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Vehicle 归属于单个用户，VIN 在全局范围内唯一。
type Vehicle struct {
	ID           uint64        `gorm:"primaryKey" json:"id"`
	UserID       uint64        `gorm:"not null;index:idx_user_vehicle" json:"user_id"`
	VehicleModel string        `gorm:"size:80;not null" json:"vehicle_model"`
	VIN          string        `gorm:"size:17;uniqueIndex;not null" json:"vin"`
	Status       VehicleStatus `gorm:"foreignKey:VehicleID" json:"status"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// VehicleStatus 保存最新车辆快照；历史遥测数据不放在本表，避免主链路过度复杂。
type VehicleStatus struct {
	VehicleID   uint64    `gorm:"primaryKey" json:"vehicle_id"`
	Battery     float64   `gorm:"not null" json:"battery"`
	Temperature float64   `gorm:"not null" json:"temperature"`
	Speed       float64   `gorm:"not null" json:"speed"`
	Location    string    `gorm:"size:120" json:"location"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AgentTask 记录异步任务的状态以及最终用户可见结果。
type AgentTask struct {
	ID           string     `gorm:"type:char(36);primaryKey" json:"id"`
	UserID       uint64     `gorm:"not null;index" json:"user_id"`
	VehicleID    uint64     `gorm:"not null;index" json:"vehicle_id"`
	Command      string     `gorm:"type:text;not null" json:"command"`
	Status       string     `gorm:"size:24;not null;index" json:"status"`
	RetryCount   int        `gorm:"not null;default:0" json:"retry_count"`
	Result       string     `gorm:"type:text" json:"result"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ToolCalls    []ToolCall `gorm:"foreignKey:TaskID" json:"tool_calls,omitempty"`
}

// ToolCall 保存工具选择、参数、输出和耗时，构成可观察的 Agent Trace。
type ToolCall struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"type:char(36);not null;index" json:"task_id"`
	ToolName  string    `gorm:"size:64;not null" json:"tool_name"`
	Input     string    `gorm:"type:json;not null" json:"input"`
	Output    string    `gorm:"type:json;not null" json:"output"`
	LatencyMS int64     `gorm:"not null" json:"latency_ms"`
	Success   bool      `gorm:"not null" json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// Conversation 保存用户与 Agent 的会话元数据，消息与异步任务保持松耦合。
type Conversation struct {
	ID        string                `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uint64                `gorm:"not null;index" json:"user_id"`
	VehicleID uint64                `gorm:"not null;index" json:"vehicle_id"`
	Title     string                `gorm:"size:120;not null" json:"title"`
	Messages  []ConversationMessage `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// ConversationMessage 只保存用户可见内容；Tool Trace 继续由 ToolCall 表承载。
type ConversationMessage struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	ConversationID string    `gorm:"type:char(36);not null;index" json:"conversation_id"`
	Role           string    `gorm:"size:16;not null" json:"role"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	TaskID         *string   `gorm:"type:char(36);index" json:"task_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// TripPlan 保存可复现的行程规划结果，地图坐标使用 WGS84。
type TripPlan struct {
	ID             string    `gorm:"type:char(36);primaryKey" json:"id"`
	UserID         uint64    `gorm:"not null;index" json:"user_id"`
	VehicleID      uint64    `gorm:"not null;index" json:"vehicle_id"`
	Origin         string    `gorm:"size:160;not null" json:"origin"`
	Destination    string    `gorm:"size:160;not null" json:"destination"`
	DistanceKM     float64   `json:"distance_km"`
	DurationMinute int       `json:"duration_minutes"`
	EnergyPercent  float64   `json:"energy_percent"`
	RemainingRange float64   `json:"remaining_range_km"`
	NeedCharge     bool      `json:"need_charge"`
	Advice         string    `gorm:"type:text" json:"advice"`
	WaypointsJSON  string    `gorm:"type:json" json:"waypoints"`
	CreatedAt      time.Time `json:"created_at"`
}
