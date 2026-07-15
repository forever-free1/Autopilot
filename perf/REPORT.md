# AutoPilot Agent 压力测试报告

测试日期：2026-07-15  
测试工具：Grafana k6 Docker 镜像  
RAG Provider：Mock  
测试时长：每个正式场景 30 秒

## 1. 测试环境

- 操作系统：Windows + Docker Desktop Linux Engine
- CPU：20 个逻辑处理器
- 内存：31.3 GB
- Go Backend：单容器、单进程
- Python Agent Worker：单容器，RabbitMQ prefetch 8
- MySQL 8.4、Redis 7.4、RabbitMQ 4.1
- 测试流量从同一 Docker 网络发起

该结果用于项目设计验证和求职展示，不代表生产环境容量。测试未包含公网延迟、真实 LLM、真实 RAGFlow、长时间稳定性和多机网络开销。

## 2. 测试场景与结果

| 场景 | 负载 | 样本/吞吐量 | P50 | P95 | P99 | 错误率/完成率 |
|---|---:|---:|---:|---:|---:|---:|
| Redis 热缓存车辆状态查询 | 200 RPS | 6,003 HTTP 请求 | 1.64 ms | 2.75 ms | 3.35 ms | HTTP 错误率 0% |
| Agent 任务提交 | 20 RPS | 603 HTTP 请求 | 8.56 ms | 15.57 ms | 23.28 ms | HTTP 错误率 0% |
| 异步任务端到端 | 10 VUs | 1,219 任务，40.57 tasks/s | 229.26 ms | 399.71 ms | 580.87 ms | 完成率 100% |

端到端场景包含：创建数据库任务、RabbitMQ 发布确认、Python Worker 消费、工具执行、内部回调、MySQL 状态更新及 Go API 轮询。

## 3. 阈值验收

| 指标 | 阈值 | 结果 |
|---|---:|---:|
| 缓存查询 P95 | < 100 ms | 2.75 ms，通过 |
| 缓存查询 P99 | < 200 ms | 3.35 ms，通过 |
| 任务提交 P95 | < 200 ms | 15.57 ms，通过 |
| 任务提交 P99 | < 400 ms | 23.28 ms，通过 |
| HTTP 错误率 | < 1% | 0%，通过 |
| 端到端任务完成率 | > 99% | 100%，通过 |
| 端到端任务 P95 | < 2,000 ms | 399.71 ms，通过 |

## 4. 分析

车辆状态查询主要由 JWT 校验和 Redis 读取组成，在 200 RPS 下延迟稳定，说明热缓存路径适合高频状态展示。

任务提交比缓存查询慢，主要包含 MySQL 写入与 RabbitMQ Publisher Confirm；P99 仍低于 25 ms，队列发布没有形成明显瓶颈。

单 Worker 在规则 Agent 和 Mock RAG 条件下达到约 40 tasks/s。端到端耗时包含 50 ms 轮询间隔，因此测得延迟略高于实际 Worker 执行时间。Worker 使用 RabbitMQ QoS，可通过增加 Worker 副本继续扩展。

## 5. 局限与后续测试

- Mock Provider 没有真实模型和 RAGFlow 的网络及推理延迟。
- 测试持续 30 秒，只属于短时基准，不是稳定性测试。
- 缓存场景使用热数据，没有覆盖缓存击穿或 MySQL 降级。
- 当前只运行一个 Backend 和一个 Worker，没有测试横向扩容。
- 没有注入 RabbitMQ、Redis、MySQL 故障。

后续可以增加 10 分钟稳定性测试、真实 RAGFlow 延迟对照、多个 Worker 的吞吐量曲线和故障恢复测试。

## 6. 复现

运行命令和参数说明见 [README.md](./README.md)，原始汇总结果位于 `results/cache.json`、`results/submit.json` 和 `results/e2e.json`。
