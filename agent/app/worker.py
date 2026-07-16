import asyncio
import json
import os
from dataclasses import asdict

import aio_pika
import httpx

from app.tools import execute_command, plan_charging_route
from app.diagnosis import diagnose
from app.rag import create_provider


async def callback(url: str, payload: dict) -> None:
    """回调由内部令牌保护，失败会抛出异常并让消息进入死信队列。"""
    async with httpx.AsyncClient(timeout=10) as client:
        response = await client.post(
            url,
            json=payload,
            headers={"X-Internal-Token": os.getenv("INTERNAL_TOKEN", "local-agent-callback-token")},
        )
        response.raise_for_status()


async def handle(message: aio_pika.IncomingMessage) -> None:
    async with message.process(requeue=False):
        task = json.loads(message.body)
        await callback(task["callback_url"], {"status": "running"})
        try:
            if any(keyword in task["command"] for keyword in ("故障", "诊断", "维修")):
                calls, result = await diagnose(task["command"], create_provider())
                patch = {}
            elif any(keyword in task["command"] for keyword in ("充电站", "规划路线", "启动导航")):
                calls, patch, result = plan_charging_route(task["command"])
            else:
                calls, patch, result = execute_command(task["command"])
            await callback(
                task["callback_url"],
                {
                    "status": "finished",
                    "result": result,
                    "vehicle_patch": patch,
                    "tool_calls": [asdict(call) for call in calls],
                },
            )
        except ValueError as exc:
            # 可预期的业务拒绝属于任务失败，但消息已经被正确消费，不应重复执行。
            await callback(task["callback_url"], {"status": "failed", "error_message": str(exc)})


async def run() -> None:
    connection = await aio_pika.connect_robust(
        os.getenv("RABBITMQ_URL", "amqp://autopilot:autopilot_dev@localhost:5672/")
    )
    channel = await connection.channel()
    await channel.set_qos(prefetch_count=8)
    queue = await channel.declare_queue(
        "agent.tasks", durable=True, arguments={"x-dead-letter-exchange": "agent.tasks.dlx"}
    )
    await queue.consume(handle)
    print("Agent Worker 已开始消费任务", flush=True)
    await asyncio.Future()


if __name__ == "__main__":
    asyncio.run(run())
