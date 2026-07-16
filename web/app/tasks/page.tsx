"use client";
import { Download, Filter, Search } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { Badge,PageHeader } from "@/components/page-kit";
import { AgentTask, listTasks } from "@/lib/api";

export default function Tasks(){
  const {data,isLoading,error}=useQuery({queryKey:["tasks"],queryFn:listTasks,refetchInterval:5000});
  const items=data?.items??[];
  const completed=items.filter(t=>t.status==="finished").length;
  const running=items.filter(t=>t.status==="running"||t.status==="pending").length;
  const failed=items.filter(t=>t.status==="failed").length;
  return <><PageHeader title="Agent 任务中心" description="追踪异步任务、消息队列状态与工具调用结果" action={<button className="button secondary"><Download size={16}/>导出记录</button>}/><div className="task-tabs"><button className="active">全部任务 <b>{data?.total??0}</b></button><button>执行中 <b>{running}</b></button><button>已完成 <b>{completed}</b></button><button>失败 <b>{failed}</b></button></div><div className="toolbar"><div className="input"><Search size={16}/>搜索任务 ID、用户请求</div><button className="button secondary"><Filter size={16}/>筛选</button></div><section className="panel task-table"><div className="table-wrap"><table><thead><tr><th>任务 ID / 用户请求</th><th>车辆 ID</th><th>状态</th><th>重试</th><th>工具调用</th><th>处理耗时</th><th>创建时间</th></tr></thead><tbody>{items.map(t=><tr key={t.id}><td><b>{t.id}</b><small>{t.command}</small></td><td>{(t as typeof t & {vehicle_id?:number}).vehicle_id??"—"}</td><td><Badge tone={t.status==="failed"?"red":t.status==="finished"?"green":"blue"}>{t.status}</Badge></td><td>{(t as typeof t & {retry_count?:number}).retry_count??0}</td><td>{t.tool_calls?.length??0}</td><td>{duration(t)}</td><td>{formatDate((t as typeof t & {created_at?:string}).created_at)}</td></tr>)}{!items.length&&<tr><td colSpan={7}><b>{isLoading?"正在加载任务…":error instanceof Error?error.message:"暂无任务"}</b><small>在智能座舱 Agent 页面提交任务后，记录会自动出现在这里。</small></td></tr>}</tbody></table></div></section></>
}
function duration(task: AgentTask){const start=task.started_at,end=task.finished_at;if(!start)return "—";if(!end)return "进行中";return `${Math.max(0,new Date(end).getTime()-new Date(start).getTime())} ms`}
function formatDate(value?:string){return value?new Date(value).toLocaleString("zh-CN",{hour12:false}):"—"}
