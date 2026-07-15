from fastapi import FastAPI

app = FastAPI(title="AutoPilot Agent", version="0.1.0")


@app.get("/health")
def health() -> dict[str, str]:
    """提供轻量健康检查，供 Docker 和 Go 后端判断 Agent 是否可用。"""
    return {"status": "ok", "service": "agent"}
