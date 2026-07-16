import { z } from "zod";
const statusSchema = z.object({ vehicle_id: z.number(), battery: z.number(), temperature: z.number(), speed: z.number(), location: z.string() });
const vehicleSchema = z.object({ id: z.number(), vehicle_model: z.string(), vin: z.string(), status: statusSchema });
export type Vehicle = z.infer<typeof vehicleSchema>;
export type ToolCall = { id: number; tool_name: string; input: string; output: string; latency_ms: number; success: boolean };
export type AgentTask = { id: string; vehicle_id?:number; command: string; status: "pending"|"running"|"finished"|"failed"; retry_count?:number; result: string; error_message: string; started_at?:string; finished_at?:string; created_at?:string; tool_calls?: ToolCall[] };
export type DashboardSummary = { online_vehicles:number; tasks_today:number; success_rate:number; average_response_ms:number; fault_vehicles:number; rabbitmq_backlog:number; tool_calls_today:number; queue_metric_source:string };
let token = "";
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api/v1${path}`, { ...options, headers: { "Content-Type": "application/json", ...(token ? { Authorization: `Bearer ${token}` } : {}), ...options.headers } });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error ?? `请求失败 (${response.status})`);
  return body as T;
}
export async function authenticate() {
  const credentials = { username: "autopilot_demo", password: "autopilot123" };
  try { token = (await request<{token:string}>("/auth/login", { method: "POST", body: JSON.stringify(credentials) })).token; }
  catch { token = (await request<{token:string}>("/auth/register", { method: "POST", body: JSON.stringify(credentials) })).token; }
}
export async function prepareVehicle() {
  await authenticate(); const data = await request<{items: unknown[]}>("/vehicles");
  if (data.items.length) return vehicleSchema.parse(data.items[0]);
  return vehicleSchema.parse(await request("/vehicles", { method: "POST", body: JSON.stringify({ vehicle_model: "AutoPilot S7", vin: "APDEMO20260000001" }) }));
}
export const createTask = (vehicleId: number, command: string) => request<AgentTask>("/agent/tasks", { method: "POST", body: JSON.stringify({ vehicle_id: vehicleId, command }) });
export const getTask = (id: string) => request<AgentTask>(`/agent/tasks/${id}`);
// 使用 fetch 读取 SSE，既能携带 JWT，也避免浏览器 EventSource 无法设置认证头的问题。
export async function watchTask(id:string,onTask:(task:AgentTask)=>void){
  const response=await fetch(`/api/v1/agent/tasks/${id}/events`,{headers:{Authorization:`Bearer ${token}`}});
  if(!response.ok||!response.body)throw new Error("任务事件流连接失败");
  const reader=response.body.getReader(),decoder=new TextDecoder();let buffer="",latest:AgentTask|undefined;
  while(true){const {done,value}=await reader.read();if(done)break;buffer+=decoder.decode(value,{stream:true});const events=buffer.split("\n\n");buffer=events.pop()??"";for(const event of events){const type=event.split("\n").find(line=>line.startsWith("event:"))?.slice(6).trim();const raw=event.split("\n").find(line=>line.startsWith("data:"))?.slice(5).trim();if(type==="task"&&raw){latest=JSON.parse(raw) as AgentTask;onTask(latest)}}}
  return latest;
}
export async function getDashboardSummary(){ await authenticate(); return request<DashboardSummary>("/dashboard/summary"); }
export async function listTasks(){ await authenticate(); return request<{items:AgentTask[];total:number;page:number;page_size:number}>("/agent/tasks?page_size=50"); }
export async function listVehicles(){ await authenticate(); const data=await request<{items:unknown[]}>("/vehicles"); return data.items.map(item=>vehicleSchema.parse(item)); }
export async function createDiagnostic(vehicleId:number,symptom:string){await authenticate();return request<AgentTask>("/diagnostics",{method:"POST",body:JSON.stringify({vehicle_id:vehicleId,symptom})})}
export type TripPlan={id:string;vehicle_id:number;origin:string;destination:string;distance_km:number;duration_minutes:number;energy_percent:number;remaining_range_km:number;need_charge:boolean;advice:string;waypoints:string;created_at:string};
export async function createTripPlan(vehicleId:number,origin:string,destination:string){await authenticate();return request<TripPlan>("/trips/plan",{method:"POST",body:JSON.stringify({vehicle_id:vehicleId,origin,destination})})}
export async function listTripPlans(){await authenticate();return request<{items:TripPlan[]}>("/trips")}
export type Conversation={id:string;vehicle_id:number;title:string;created_at:string;updated_at:string};
export async function listConversations(){await authenticate();return request<{items:Conversation[]}>("/conversations")}
export async function createConversation(vehicleId:number,title:string){await authenticate();return request<Conversation>("/conversations",{method:"POST",body:JSON.stringify({vehicle_id:vehicleId,title})})}
export async function sendConversationMessage(conversationId:string,content:string){await authenticate();return request<{task:AgentTask}>(`/conversations/${conversationId}/messages`,{method:"POST",body:JSON.stringify({content})})}
