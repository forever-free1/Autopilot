import { FormEvent, StrictMode, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import { AgentTask, createTask, getTask, getVehicleStatus, prepareDemo, Vehicle, VehicleStatus } from "./api";
import "./styles.css";

const defaultCommand = "把温度调到22度，并打开座椅加热";

function App() {
  const [mode, setMode] = useState<"control" | "diagnosis">("control");
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [status, setStatus] = useState<VehicleStatus | null>(null);
  const [task, setTask] = useState<AgentTask | null>(null);
  const [command, setCommand] = useState(defaultCommand);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    prepareDemo()
      .then((item) => {
        setVehicle(item);
        setStatus(item.status);
      })
      .catch((reason: Error) => setError(reason.message))
      .finally(() => setLoading(false));
  }, []);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!vehicle || !command.trim() || task?.status === "running" || task?.status === "pending") return;
    setError("");
    try {
      let current = await createTask(vehicle.id, command.trim());
      setTask(current);
      // 异步任务通常在数百毫秒内完成，短轮询能让 Demo 清楚展示状态变化。
      for (let attempt = 0; attempt < 40; attempt += 1) {
        await new Promise((resolve) => setTimeout(resolve, 250));
        current = await getTask(current.id);
        setTask(current);
        if (current.status === "finished" || current.status === "failed") break;
      }
      setStatus(await getVehicleStatus(vehicle.id));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "执行指令失败");
    }
  }

  const busy = task?.status === "pending" || task?.status === "running";
  const stateLabel = loading ? "正在连接" : error ? "连接异常" : busy ? "Agent 执行中" : "系统就绪";
  const citations = extractCitations(task);

  function switchMode(next: "control" | "diagnosis") {
    setMode(next);
    setTask(null);
    setError("");
    setCommand(next === "control" ? defaultCommand : "诊断故障码 P0A80，车辆加速无力且仪表提示检查动力电池");
  }

  return (
    <main className="shell">
      <aside className="rail">
        <div className="brand">AUTOPILOT</div>
        <nav aria-label="主导航">
          <button className={mode === "control" ? "active" : ""} onClick={() => switchMode("control")}><span>01</span> 座舱控制</button>
          <button className={mode === "diagnosis" ? "active" : ""} onClick={() => switchMode("diagnosis")}><span>02</span> 故障诊断</button>
          <a href="http://localhost:15672" target="_blank"><span>03</span> 消息队列</a>
        </nav>
        <div className="vehicle-id">
          <span>{vehicle?.vehicle_model ?? "连接车辆中"}</span>
          <small>{vehicle?.vin ?? "—"}</small>
        </div>
      </aside>

      <section className="workspace" id="console">
        <header>
          <div><span className="eyebrow">实时车辆 · {status?.location ?? "—"}</span><h1>{mode === "control" ? "座舱控制台" : "车辆健康诊断"}</h1></div>
          <span className={`system-status ${busy ? "working" : ""}`}><i /> {stateLabel}</span>
        </header>

        <section className="telemetry" aria-label="车辆状态">
          <div><span>剩余电量</span><strong>{status?.battery ?? "—"}<em>%</em></strong></div>
          <div><span>座舱温度</span><strong>{status?.temperature ?? "—"}<em>°C</em></strong></div>
          <div><span>当前速度</span><strong>{status?.speed ?? "—"}<em> km/h</em></strong></div>
        </section>

        <form className="command" onSubmit={submit}>
          <label htmlFor="command">{mode === "control" ? "向车辆发送自然语言指令" : "输入故障码与车辆症状"}</label>
          <div className="prompt">
            <input id="command" value={command} onChange={(event) => setCommand(event.target.value)} disabled={loading || busy} />
            <button disabled={loading || busy || !vehicle}>
              {busy ? <><span className="spinner" />执行中</> : mode === "control" ? "执行指令" : "开始诊断"}
            </button>
          </div>
          {mode === "control" ? (
            <div className="suggestions">
              <button type="button" onClick={() => setCommand(defaultCommand)}>22°C + 座椅加热</button>
              <button type="button" onClick={() => setCommand("将座舱温度调到24度")}>设为 24°C</button>
              <button type="button" onClick={() => setCommand("关闭座椅加热")}>关闭座椅加热</button>
            </div>
          ) : (
            <div className="suggestions">
              <button type="button" onClick={() => setCommand("诊断故障码 P0A80，车辆加速无力且仪表提示检查动力电池")}>P0A80 · 动力电池</button>
              <button type="button" onClick={() => setCommand("诊断车辆无法启动，低压系统提示故障")}>无法启动 · 低压系统</button>
            </div>
          )}
        </form>

        {(task || error) && (
          <section className={`result ${task?.status ?? "failed"}`}>
            <span>{task?.status === "finished" ? "执行完成" : task?.status === "failed" || error ? "执行失败" : "正在规划工具"}</span>
            <p>{error || task?.error_message || task?.result || "Agent 已接收任务，等待 Worker 返回结果。"}</p>
            {task && <code>{task.id}</code>}
            {citations.length > 0 && <p className="disclaimer">辅助诊断结果不替代专业维修判断。</p>}
          </section>
        )}
      </section>

      <aside className="inspector" id="trace">
        <div className="inspector-head"><span>AGENT TRACE</span><b>{task?.tool_calls?.length ?? 0} TOOLS</b></div>
        {!task && <div className="empty-trace"><span>等待指令</span><p>工具调用与执行耗时将在这里实时出现。</p></div>}
        {task && (
          <div className="timeline">
            <TraceItem index="01" title="Intent Detection" detail={mode === "control" ? "识别座舱控制意图" : "识别车辆故障诊断意图"} done />
            <TraceItem index="02" title="Task Queue" detail={task.status === "pending" ? "等待 Worker" : "RabbitMQ 已投递"} done={task.status !== "pending"} />
            {(task.tool_calls ?? []).map((call, index) => (
              <TraceItem key={call.id} index={String(index + 3).padStart(2, "0")} title={call.tool_name} detail={`${formatInput(call.input)} · ${call.latency_ms} ms`} done={call.success} />
            ))}
            <TraceItem index={String((task.tool_calls?.length ?? 0) + 3).padStart(2, "0")} title="Verification" detail={task.status === "finished" ? mode === "control" ? "车辆状态已确认" : "诊断依据已确认" : "等待执行结果"} done={task.status === "finished"} />
            {citations.length > 0 && <section className="citations"><h2>引用依据</h2>{citations.map((citation) => <article key={citation.source}><b>{citation.title}</b><p>{citation.content}</p><span>{Math.round(citation.score * 100)}% · {citation.source}</span></article>)}</section>}
          </div>
        )}
      </aside>
    </main>
  );
}

type Citation = { title: string; content: string; source: string; score: number };

function extractCitations(task: AgentTask | null): Citation[] {
  const searchCall = task?.tool_calls?.find((call) => call.tool_name === "search_manual");
  if (!searchCall) return [];
  try {
    return JSON.parse(searchCall.output).citations ?? [];
  } catch {
    return [];
  }
}

function TraceItem({ index, title, detail, done }: { index: string; title: string; detail: string; done: boolean }) {
  return <div className={`trace-item ${done ? "done" : "pending"}`}><span>{index}</span><div><b>{title}</b><p>{detail}</p></div><i /></div>;
}

function formatInput(raw: string) {
  try {
    const input = JSON.parse(raw);
    return Object.values(input).join(" · ");
  } catch {
    return raw;
  }
}

createRoot(document.getElementById("root")!).render(<StrictMode><App /></StrictMode>);
