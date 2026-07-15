import httpx
import pytest

from app.diagnosis import diagnose
from app.rag.provider import MockRAGProvider, RAGFlowProvider


@pytest.mark.anyio
async def test_mock_diagnosis_has_citations() -> None:
    calls, result = await diagnose("诊断故障码 P0A80", MockRAGProvider())
    assert [call.tool_name for call in calls] == ["query_fault_code", "search_manual"]
    assert calls[1].output["citations"][0]["score"] == 0.96
    assert "不替代专业检修" in result


@pytest.mark.anyio
async def test_ragflow_response_mapping(monkeypatch: pytest.MonkeyPatch) -> None:
    async def fake_post(*args, **kwargs):
        return httpx.Response(
            200,
            request=httpx.Request("POST", "http://ragflow/api/v1/retrieval"),
            json={
                "data": {
                    "chunks": [
                        {
                            "content": "维修片段",
                            "document_keyword": "维修手册",
                            "document_id": "doc-1",
                            "similarity": 0.91,
                        }
                    ]
                }
            },
        )

    monkeypatch.setattr(httpx.AsyncClient, "post", fake_post)
    citations = await RAGFlowProvider("http://ragflow", "key", ["dataset"]).search("P0A80")
    assert citations[0].title == "维修手册"
    assert citations[0].score == 0.91
