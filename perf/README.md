# 压力测试

测试使用官方 `grafana/k6` Docker 镜像，不要求本机安装 k6。先启动项目：

```text
docker compose up -d
```

Windows PowerShell 运行示例：

```text
docker run --rm --network autopilot_default -v ${PWD}/perf:/work -w /work grafana/k6 run -e SCENARIO=cache -e DURATION=30s --summary-export=/work/results/cache.json k6/load.js
docker run --rm --network autopilot_default -v ${PWD}/perf:/work -w /work grafana/k6 run -e SCENARIO=submit -e DURATION=30s --summary-export=/work/results/submit.json k6/load.js
docker run --rm --network autopilot_default -v ${PWD}/perf:/work -w /work grafana/k6 run -e SCENARIO=e2e -e DURATION=30s --summary-export=/work/results/e2e.json k6/load.js
```

场景说明：

- `cache`：默认 200 RPS 查询同一车辆状态，主要衡量 Go、JWT 和 Redis 热缓存。
- `submit`：默认 20 RPS 创建异步 Agent 任务，衡量 MySQL 写入与 RabbitMQ 发布确认。
- `e2e`：默认 10 个并发用户提交任务并轮询完成，衡量 Worker 端到端延迟。

可通过 `RATE`、`VUS` 和 `DURATION` 环境变量调整强度。压测会自动创建独立用户和车辆，不使用 Demo 账号。
