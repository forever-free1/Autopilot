# AutoPilot Agent

[![CI](https://github.com/forever-free1/Autopilot/actions/workflows/ci.yml/badge.svg)](https://github.com/forever-free1/Autopilot/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

面向智能座舱的多工具协同任务执行平台，用于展示 Go 后端、Python Agent 与异步任务系统的完整工程能力。

仓库地址：<https://github.com/forever-free1/Autopilot>

## 当前状态

异步 Agent 主链路已可运行：用户注册/登录、车辆管理、Redis 缓存、RabbitMQ 任务投递、Python Worker、模拟车辆工具和 Tool Trace。核心代码采用中文注释，重点说明业务规则与设计原因。

## 技术结构

- `backend/`：Go + Gin 业务 API
- `agent/`：Python + FastAPI Agent Worker
- `web/`：Next.js + React + TypeScript 智能座舱 Agent 管理与演示平台
- `docker-compose.yml`：MySQL、Redis、RabbitMQ 与三个应用服务

## 快速启动

1. 复制环境变量：`Copy-Item .env.example .env`（Windows PowerShell）
2. 启动完整环境：`docker compose up --build`
3. 打开控制台：<http://localhost:5173>

服务入口：

- Go 健康检查：<http://localhost:8080/health>
- Agent 健康检查：<http://localhost:8000/health>
- RabbitMQ 管理页：<http://localhost:15672>

API 契约见 [docs/openapi.yaml](./docs/openapi.yaml)。

## Phase 1 API

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/vehicles
POST /api/v1/vehicles
GET  /api/v1/vehicles/{id}/status
POST /api/v1/agent/tasks
GET  /api/v1/agent/tasks/{id}
```

车辆接口需要请求头 `Authorization: Bearer <token>`。车辆状态响应中的 `X-Cache` 为 `MISS` 或 `HIT`，可直接观察缓存是否生效。

## Agent 执行链路

```text
Go API 创建 pending 任务
  → RabbitMQ agent.tasks
  → Python Worker 解析中文指令
  → set_temperature / seat_control
  → 内部回调更新任务、Tool Trace 和车辆状态
  → Redis 车辆状态缓存失效
```

当前规则 Agent 支持座舱温度和主驾座椅加热，作为不依赖外部模型的稳定演示与压测基线。

## Web Demo

打开 <http://localhost:5173> 后，页面会自动登录固定演示账号并在首次运行时创建演示车辆。推荐录屏流程：

1. 展示默认车辆状态和系统就绪状态。
2. 执行“把温度调到22度，并打开座椅加热”。
3. 展示右侧 RabbitMQ、Tool Calling、Verification 轨迹，以及车辆温度更新。

整个主链路无需外部模型密钥，可稳定用于 3 分钟以内的 Demo 视频。

## 故障诊断与 RAGFlow

左侧切换到“故障诊断”，可演示故障码提取、维修手册检索、诊断建议和引用来源。默认 `RAG_PROVIDER=mock`；真实 RAGFlow 接入方法见 [deploy/ragflow/README.md](./deploy/ragflow/README.md)。

停止环境：`docker compose down`。如需同时清空本地演示数据，使用 `docker compose down -v`。

## 本地测试

```text
cd backend && go test ./...
cd agent && .venv/Scripts/python -m pytest
cd web && npm run build
```

## 最终交付目标

- 可复现的 GitHub 仓库
- 3 分钟以内的主链路 Demo 视频
- k6 压力测试脚本与性能报告

首版压力测试已完成，报告见 [perf/REPORT.md](./perf/REPORT.md)。当前本机基准下，缓存查询 P95 为 2.75 ms，异步任务端到端 P95 为 399.71 ms，任务完成率为 100%。

详细路线见 [AutoPilot_Agent_Project_Plan.md](./AutoPilot_Agent_Project_Plan.md)。
