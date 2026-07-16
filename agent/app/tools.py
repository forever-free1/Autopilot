import re
import time
from dataclasses import dataclass
from typing import Any


@dataclass
class ToolResult:
    tool_name: str
    input: dict[str, Any]
    output: dict[str, Any]
    latency_ms: int
    success: bool = True


def execute_command(command: str) -> tuple[list[ToolResult], dict[str, Any], str]:
    """将自然语言转换为确定性车辆工具调用，作为后续 LLM Tool Calling 的稳定基线。"""
    calls: list[ToolResult] = []
    patch: dict[str, Any] = {}
    temperature = re.search(r"(?:温度.*?|调到\s*)(\d{1,2}(?:\.\d+)?)\s*度", command)
    if temperature:
        started = time.perf_counter()
        value = float(temperature.group(1))
        if not 16 <= value <= 30:
            raise ValueError("座舱温度必须在 16 到 30 度之间")
        patch["temperature"] = value
        calls.append(
            ToolResult(
                "set_temperature",
                {"temperature": value},
                {"applied": True},
                int((time.perf_counter() - started) * 1000),
            )
        )
    if "座椅加热" in command:
        started = time.perf_counter()
        enabled = not any(word in command for word in ("关闭", "关掉"))
        calls.append(
            ToolResult(
                "seat_control",
                {"seat": "driver", "heating": enabled},
                {"applied": True},
                int((time.perf_counter() - started) * 1000),
            )
        )
    if not calls:
        raise ValueError("暂不支持该指令，请尝试调节温度或控制座椅加热")
    names = "、".join(call.tool_name for call in calls)
    return calls, patch, f"指令执行完成，已调用 {names}。"


def plan_charging_route(command: str) -> tuple[list[ToolResult], dict[str, Any], str]:
    """以确定性工具链完成充电站规划，保证离线面试 Demo 可重复执行。"""
    calls: list[ToolResult] = []

    def record(name: str, inputs: dict[str, Any], output: dict[str, Any]) -> None:
        started = time.perf_counter()
        calls.append(ToolResult(name, inputs, output, int((time.perf_counter() - started) * 1000)))

    record("get_vehicle_status", {}, {"battery": 72, "remaining_range_km": 386, "location": "上海张江"})
    record(
        "search_charging_stations",
        {"radius_km": 10},
        {"stations": [{"name": "张江超级充电站", "distance_km": 1.2, "available": 8}]},
    )
    record(
        "calculate_route",
        {"destination": "张江超级充电站"},
        {"distance_km": 1.2, "duration_minutes": 6, "energy_percent": 1},
    )
    record("start_navigation", {"destination": "张江超级充电站"}, {"started": True})
    return calls, {}, "已找到最近的张江超级充电站，距离 1.2 km，当前续航充足，路线已规划并启动导航。"
