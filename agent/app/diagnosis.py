import re
import time

from app.rag import RAGProvider
from app.tools import ToolResult


async def diagnose(command: str, provider: RAGProvider) -> tuple[list[ToolResult], str]:
    """组合故障码提取和手册检索，输出可解释的诊断建议而不是确定性维修结论。"""
    fault_match = re.search(r"\b([PBCU][0-9A-F]{4})\b", command.upper())
    fault_code = fault_match.group(1) if fault_match else "UNKNOWN"
    started = time.perf_counter()
    fault_call = ToolResult(
        "query_fault_code",
        {"fault_code": fault_code},
        {"found": fault_code != "UNKNOWN", "severity": "warning"},
        int((time.perf_counter() - started) * 1000),
    )

    started = time.perf_counter()
    citations = await provider.search(command)
    search_call = ToolResult(
        "search_manual",
        {"query": command},
        {"citations": [citation.to_dict() for citation in citations]},
        int((time.perf_counter() - started) * 1000),
    )
    if not citations:
        raise ValueError("知识库未检索到足够依据，请补充故障码或车辆症状")
    advice = (
        f"检测到故障码 {fault_code}。可能与动力电池性能衰退或模组压差异常有关；"
        "建议降低高负载行驶，检查电池冷却与高压连接，并预约具备高压资质的维修中心。"
        "本结果仅用于辅助诊断，不替代专业检修。"
    )
    return [fault_call, search_call], advice
