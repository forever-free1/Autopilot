# 前端工作台 API（阶段 2）

以下接口统一位于 `/api/v1`，并要求 `Authorization: Bearer <JWT>`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/dashboard/summary` | 聚合当日任务、成功率、平均耗时、车辆和工具调用指标 |
| GET | `/vehicles/{id}` | 获取当前用户的车辆详情和最新状态 |
| GET | `/agent/tasks` | 分页查询任务，支持 `page`、`page_size`、`status` 和 `q` |
| GET | `/agent/tasks/{id}/events` | 通过 SSE 订阅 `task` 快照和 `done` 结束事件 |
| GET | `/tool-calls` | 查询当前用户最近的工具调用轨迹 |

RabbitMQ 精确积压量尚未配置管理指标来源。`dashboard/summary` 会返回
`queue_metric_source: "not_configured"`，前端显示 `N/A`，避免把模拟值误认为真实监控数据。
