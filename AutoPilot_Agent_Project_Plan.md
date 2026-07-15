# AutoPilot Agent：面向智能座舱的多工具协同任务执行平台

> 项目定位：用于求职展示的完整工程项目。优先保证可复现、可演示、可测试，避免为了“企业级”而引入不必要的复杂度。

## 0. 最终交付与验收标准

- 完整 GitHub 仓库：克隆后按 README 可使用 Docker Compose 一键启动。
- 3 分钟以内 Demo 视频：展示车辆控制、故障诊断、RAG 引用和 Agent Trace 主链路。
- 压力测试：保留 k6 脚本和包含 P50、P95、P99、吞吐量、错误率的测试报告。
- 核心代码使用中文注释：重点解释业务规则、状态流转、幂等、重试与异常处理；变量和接口仍使用规范英文命名。
- 自动化质量检查：核心单元测试、主链路集成测试和 GitHub Actions。

### 范围约束

保留 Go、Python Agent、Web Demo、MySQL、Redis、RabbitMQ 和向量检索；不引入 Kubernetes、服务网格、复杂 RBAC、多租户或自研工作流引擎。

## 1. 项目定位

目标：构建一个接近真实车企智能座舱 Agent 架构的系统，补齐 Go
后端工程能力，同时结合大模型 Agent 能力。

核心目标：

-   Go Gin 负责业务后端、用户系统、车辆服务、任务调度和基础设施。
-   Python Agent 负责大模型推理、RAG、工具调用和任务规划。
-   MySQL 保存长期业务数据。
-   Redis 提供缓存、状态管理、限流和幂等能力。
-   RabbitMQ 实现异步 Agent Worker 调度。

最终形成：

用户 → Go 服务 → MQ → Python Agent → Vehicle Tools → 数据服务 → 返回结果

------------------------------------------------------------------------

# 2. 项目功能范围

## 2.1 智能座舱助手

支持自然语言控制车辆功能：

示例：

"帮我把温度调到22度，打开座椅加热。"

Agent 调用：

-   set_temperature
-   seat_control
-   window_control
-   music_control
-   navigation

目标能力：

-   Tool Calling
-   参数解析
-   工具执行
-   执行结果反馈

------------------------------------------------------------------------

## 2.2 车辆健康诊断 Agent

模拟真实车企售后诊断场景。

输入：

-   故障码
-   电池状态
-   温度
-   里程
-   历史维修记录

Agent：

1.  获取车辆状态
2.  查询维修知识库
3.  分析可能原因
4.  输出维修建议

评价指标：

-   Diagnosis Accuracy
-   Recommendation Quality
-   Response Latency

------------------------------------------------------------------------

## 2.3 汽车知识库 RAG

数据来源：

-   用户手册
-   维修手册
-   OTA 更新说明
-   保养规范

支持：

-   文档切分
-   向量检索
-   多轮问答
-   引用来源

------------------------------------------------------------------------

## 2.4 Agent任务执行平台

记录：

-   用户请求
-   Agent规划过程
-   Tool调用过程
-   执行结果
-   错误信息

形成 Agent Trace。

------------------------------------------------------------------------

# 3. 系统架构

    Frontend
        |
        |
    Go Gin Backend
        |
        +------ MySQL
        |
        +------ Redis
        |
        +------ RabbitMQ
                    |
                    |
              Python Agent Worker
                    |
            +-------+-------+
            |
         Tool System
            |
     Vehicle API / RAG / Planner

------------------------------------------------------------------------

# 4. 技术栈

## Backend

-   Go
-   Gin
-   GORM
-   MySQL
-   Redis
-   RabbitMQ
-   Docker

## Agent

-   Python
-   FastAPI
-   LangGraph 或自研状态机
-   LangChain组件（可选）
-   向量数据库

------------------------------------------------------------------------

# 5. 数据库设计

## users

用户信息。

字段：

-   id
-   username
-   password_hash
-   created_at

## vehicles

车辆信息。

字段：

-   id
-   user_id
-   vehicle_model
-   vin

## vehicle_status

车辆实时状态。

字段：

-   vehicle_id
-   battery
-   temperature
-   speed
-   location

## conversations

会话。

字段：

-   id
-   user_id
-   created_at

## messages

聊天记录。

字段：

-   conversation_id
-   role
-   content
-   created_at

## agent_tasks

Agent任务。

字段：

-   task_id
-   status
-   retry_count
-   result
-   error_message

## tool_calls

工具调用记录。

字段：

-   task_id
-   tool_name
-   input
-   output
-   latency
-   success

------------------------------------------------------------------------

# 6. Redis设计

## 车辆状态缓存

    vehicle:{id}:status

保存：

-   当前速度
-   电量
-   温度

## Agent任务状态

    task:{task_id}:status

状态：

pending

running

tool_calling

finished

failed

## 会话缓存

    conversation:{id}:memory

保存最近上下文。

## 幂等控制

    request:{request_id}

避免重复提交。

## 分布式锁

    lock:task:{task_id}

避免多个Worker重复执行。

------------------------------------------------------------------------

# 7. RabbitMQ设计

## Producer

Go Backend：

负责：

-   创建任务
-   发布Agent任务消息

## Consumer

Python Worker：

负责：

-   消费任务
-   执行Agent
-   回写结果

支持：

-   消息确认
-   自动重试
-   死信队列
-   幂等消费

------------------------------------------------------------------------

# 8. Agent设计

状态流程：

    START

    ↓

    Intent Detection

    ↓

    Planner

    ↓

    Tool Selection

    ↓

    Tool Execution

    ↓

    Verification

    ↓

    Answer Generation

    ↓

    END

工具：

## Vehicle Tools

-   query_vehicle_status
-   set_temperature
-   seat_control
-   navigation

## Diagnostic Tools

-   query_fault_code
-   search_manual

## Business Tools

-   create_service_ticket
-   query_service_status

------------------------------------------------------------------------

# 9. 开发计划

## Phase 0：环境搭建

时间：

1-2天

完成：

-   Docker环境
-   MySQL
-   Redis
-   RabbitMQ
-   Go工程初始化
-   Python工程初始化

------------------------------------------------------------------------

# Phase 1：Go业务后端

时间：

1周

实现：

-   用户管理
-   车辆管理
-   会话管理
-   消息接口

技术重点：

-   Gin路由
-   GORM
-   MySQL设计
-   REST API

成果：

Go Backend MVP。

------------------------------------------------------------------------

# Phase 2：Redis接入

时间：

3-5天

实现：

-   登录缓存
-   会话缓存
-   车辆状态缓存
-   请求幂等
-   分布式锁

成果：

具备生产级缓存能力。

------------------------------------------------------------------------

# Phase 3：Python Agent

时间：

1周

实现：

-   Agent状态机
-   Tool Calling
-   RAG
-   Vehicle Tools

成果：

智能座舱Agent Demo。

------------------------------------------------------------------------

# Phase 4：RabbitMQ异步化

时间：

1周

实现：

-   Go Producer
-   Python Consumer
-   Task Queue
-   Retry
-   Dead Letter Queue

成果：

异步Agent执行平台。

------------------------------------------------------------------------

# Phase 5：工程优化

时间：

1周

增加：

-   Docker Compose
-   日志系统
-   API文档
-   单元测试
-   压力测试

------------------------------------------------------------------------

# 10. 面试重点准备

## 为什么Go和Python分离？

回答：

Go负责高并发业务服务和稳定接口，Python负责模型生态和Agent逻辑，通过接口解耦。

## 为什么需要MQ？

回答：

LLM调用具有高延迟和不确定性，异步队列可以削峰、重试和解耦。

## 为什么Redis不能替代MySQL？

回答：

Redis适合高速访问和临时状态，MySQL负责可靠持久化。

## 如何避免Agent任务重复执行？

回答：

task_id唯一约束 + Redis分布式锁 + 数据库条件更新。

------------------------------------------------------------------------

# 11. 最终简历描述

AutoPilot Agent：面向智能座舱的多工具协同任务执行平台

-   基于Go Gin与Python
    Agent构建智能座舱任务执行系统，实现车辆控制、故障诊断、导航规划和汽车知识问答。
-   设计Agent Tool
    Calling框架，通过车辆状态API、导航API和维修知识库实现多工具协同推理。
-   使用MySQL持久化用户、车辆状态、会话和Agent执行轨迹，Redis实现车辆状态缓存、任务状态管理和请求幂等。
-   引入RabbitMQ构建异步Agent
    Worker，实现故障诊断、车辆检查等长耗时任务可靠执行。
-   构建Agent执行评价模块，从工具调用准确率、任务完成率和响应延迟评估系统性能。

# 12. 参考项目

学习参考：

-   VehicleWorld：汽车Agent环境和工具设计
-   CAR-bench：汽车Agent任务评测
-   LangGraph：Agent状态机设计
-   CloudWeGo Eino：Go AI应用设计
-   Gin REST项目：后端工程结构参考
