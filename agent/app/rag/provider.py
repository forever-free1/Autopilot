import os
from dataclasses import asdict, dataclass
from typing import Protocol

import httpx


@dataclass
class Citation:
    title: str
    content: str
    source: str
    score: float

    def to_dict(self) -> dict:
        return asdict(self)


class RAGProvider(Protocol):
    async def search(self, question: str) -> list[Citation]: ...


class MockRAGProvider:
    """内置少量汽车手册片段，保证 Demo 和压测不依赖外部模型或网络。"""

    async def search(self, question: str) -> list[Citation]:
        normalized = question.upper()
        if "P0A80" in normalized:
            return [
                Citation(
                    title="高压动力电池维修手册",
                    content="故障码 P0A80 表示动力电池组性能衰退。应读取各电池模组压差，并检查冷却系统与高压连接器。",
                    source="manuals/high_voltage_battery.md#P0A80",
                    score=0.96,
                ),
                Citation(
                    title="新能源汽车保养规范",
                    content="若单体压差持续超过制造商阈值，应停止高负载行驶并预约具备高压资质的维修中心检测。",
                    source="manuals/maintenance_guide.md#battery-safety",
                    score=0.89,
                ),
            ]
        return [
            Citation(
                title="车辆故障诊断通则",
                content="确认仪表故障提示，记录故障码和环境条件；涉及制动、高压或转向系统时，应优先停车并联系维修服务。",
                source="manuals/diagnostic_basics.md#workflow",
                score=0.78,
            )
        ]


class RAGFlowProvider:
    def __init__(self, base_url: str, api_key: str, dataset_ids: list[str]) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.dataset_ids = dataset_ids

    async def search(self, question: str) -> list[Citation]:
        """调用 RAGFlow 检索 API；只消费稳定字段，并兼容文档标题字段的版本差异。"""
        async with httpx.AsyncClient(timeout=30) as client:
            response = await client.post(
                f"{self.base_url}/api/v1/retrieval",
                headers={"Authorization": f"Bearer {self.api_key}"},
                json={"question": question, "dataset_ids": self.dataset_ids, "top_n": 3},
            )
            response.raise_for_status()
            payload = response.json()
        chunks = (payload.get("data") or {}).get("chunks") or []
        return [
            Citation(
                title=chunk.get("document_keyword") or chunk.get("docnm_kwd") or "RAGFlow 文档",
                content=chunk.get("content_with_weight") or chunk.get("content") or "",
                source=chunk.get("document_id") or chunk.get("doc_id") or "ragflow",
                score=float(chunk.get("similarity") or 0),
            )
            for chunk in chunks
            if chunk.get("content_with_weight") or chunk.get("content")
        ]


def create_provider() -> RAGProvider:
    provider = os.getenv("RAG_PROVIDER", "mock").lower()
    if provider == "mock":
        return MockRAGProvider()
    if provider == "ragflow":
        base_url = os.getenv("RAGFLOW_BASE_URL", "")
        api_key = os.getenv("RAGFLOW_API_KEY", "")
        dataset_ids = [
            item.strip() for item in os.getenv("RAGFLOW_DATASET_IDS", "").split(",") if item.strip()
        ]
        if not base_url or not api_key or not dataset_ids:
            raise RuntimeError("RAGFlow 配置不完整，请检查地址、API Key 和知识库 ID")
        return RAGFlowProvider(base_url, api_key, dataset_ids)
    raise RuntimeError(f"不支持的 RAG Provider: {provider}")
