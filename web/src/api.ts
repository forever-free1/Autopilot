export type VehicleStatus = {
  vehicle_id: number;
  battery: number;
  temperature: number;
  speed: number;
  location: string;
};

export type Vehicle = {
  id: number;
  vehicle_model: string;
  vin: string;
  status: VehicleStatus;
};

export type ToolCall = {
  id: number;
  tool_name: string;
  input: string;
  output: string;
  latency_ms: number;
  success: boolean;
};

export type AgentTask = {
  id: string;
  command: string;
  status: "pending" | "running" | "finished" | "failed";
  result: string;
  error_message: string;
  tool_calls?: ToolCall[];
};

let token = "";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api/v1${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });
  const text = response.status === 204 ? "" : await response.text();
  let body: any = null;
  try {
    body = text ? JSON.parse(text) : null;
  } catch {
    body = { error: `服务暂不可用 (${response.status})` };
  }
  if (!response.ok) throw new Error(body?.error ?? `请求失败 (${response.status})`);
  return body as T;
}

async function authenticate() {
  const credentials = { username: "autopilot_demo", password: "autopilot123" };
  try {
    const result = await request<{ token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(credentials),
    });
    token = result.token;
  } catch {
    const result = await request<{ token: string }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(credentials),
    });
    token = result.token;
  }
}

export async function prepareDemo(): Promise<Vehicle> {
  await authenticate();
  const list = await request<{ items: Vehicle[] }>("/vehicles");
  if (list.items.length) return list.items[0];
  return request<Vehicle>("/vehicles", {
    method: "POST",
    body: JSON.stringify({ vehicle_model: "AutoPilot S7", vin: "APDEMO20260000001" }),
  });
}

export const getVehicleStatus = (id: number) => request<VehicleStatus>(`/vehicles/${id}/status`);

export const createTask = (vehicleId: number, command: string) =>
  request<AgentTask>("/agent/tasks", {
    method: "POST",
    body: JSON.stringify({ vehicle_id: vehicleId, command }),
  });

export const getTask = (id: string) => request<AgentTask>(`/agent/tasks/${id}`);
