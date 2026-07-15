# RAGFlow 可选部署

RAGFlow 作为独立知识库服务运行，不加入 AutoPilot 默认 Compose。这样默认 Demo 和压力测试保持轻量，真实文档解析与检索可按需启用。

## 准备步骤

1. 按 [RAGFlow 官方 Quickstart](https://ragflow.io/docs/) 部署稳定版本。
2. 在 RAGFlow 创建汽车维修知识库，上传 `../../knowledge/manuals/` 中的示例文档。
3. 等待文档解析完成，并在 RAGFlow 中测试检索效果。
4. 创建 API Key，复制知识库 ID。
5. 修改项目根目录 `.env`：

```text
RAG_PROVIDER=ragflow
RAGFLOW_BASE_URL=http://host.docker.internal:9380
RAGFLOW_API_KEY=你的 API Key
RAGFLOW_DATASET_IDS=知识库ID
```

6. 重建 Worker：`docker compose up -d --build agent-worker`

多个知识库 ID 使用英文逗号分隔。恢复稳定演示模式时，将 `RAG_PROVIDER` 改回 `mock`。

## 集成边界

AutoPilot 只调用 RAGFlow 的 `POST /api/v1/retrieval` 检索接口，文档上传、解析、切分和索引管理均留在 RAGFlow。Worker 将检索分块转换为统一 Citation，再写入 Tool Trace。
