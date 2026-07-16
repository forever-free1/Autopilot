import pytest

from app.tools import execute_command, plan_charging_route


def test_execute_multiple_tools() -> None:
    calls, patch, result = execute_command("把温度调到22度，并打开座椅加热")
    assert [call.tool_name for call in calls] == ["set_temperature", "seat_control"]
    assert patch == {"temperature": 22.0}
    assert "执行完成" in result


def test_reject_unsupported_command() -> None:
    with pytest.raises(ValueError, match="暂不支持"):
        execute_command("播放音乐")


def test_plan_charging_route_has_complete_trace() -> None:
    calls, patch, result = plan_charging_route("帮我找到最近的充电站并规划路线")
    assert [call.tool_name for call in calls] == [
        "get_vehicle_status", "search_charging_stations", "calculate_route", "start_navigation"
    ]
    assert patch == {}
    assert "启动导航" in result
