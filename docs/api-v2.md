# 前端工作台 API（阶段 2）

以下接口统一位于 `/api/v1`，并要求 `Authorization: Bearer <JWT>`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/dashboard/summary` | 聚合当日任务、成功率、平均耗时、车辆和工具调用指标 |
| GET | `/vehicles/{id}` | 获取当前用户的车辆详情和最新状态 |
| GET | `/agent/tasks` | 分页查询任务，支持 `page`、`page_size`、`status` 和 `q` |
| GET | `/agent/tasks/{id}/events` | 通过 SSE 订阅 `task` 快照和 `done` 结束事件 |
| GET | `/tool-calls` | 查询当前用户最近的工具调用轨迹 |
| GET/POST | `/conversations` | 查询或创建 Agent 会话 |
| GET | `/conversations/{id}` | 获取会话及按时间排序的消息 |
| POST | `/conversations/{id}/messages` | 保存用户消息并创建关联 Agent 任务 |
| POST | `/diagnostics` | 创建经过 RabbitMQ 和 Python Agent 的诊断任务 |
| POST | `/trips/plan` | 生成并持久化可复现的行程规划 |
| GET | `/trips` | 查询最近的行程规划记录 |

RabbitMQ 精确积压量尚未配置管理指标来源。`dashboard/summary` 会返回
`queue_metric_source: "not_configured"`，前端显示 `N/A`，避免把模拟值误认为真实监控数据。

行程规划第一版使用确定性估算和固定 WGS84 演示坐标，以保证面试 Demo 在离线环境中仍可复现。
后续接入地图供应商时只替换 Go 服务内部的规划适配器，前端接口保持不变。
