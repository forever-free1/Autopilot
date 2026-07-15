import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";
import exec from "k6/execution";

const scenario = __ENV.SCENARIO || "cache";
const baseURL = __ENV.BASE_URL || "http://backend:8080";
const taskCompleted = new Rate("task_completed");
const taskE2E = new Trend("task_e2e_duration", true);

const scenarios = {
  cache: {
    executor: "constant-arrival-rate",
    rate: Number(__ENV.RATE || 200),
    timeUnit: "1s",
    duration: __ENV.DURATION || "30s",
    preAllocatedVUs: 30,
    maxVUs: 100,
    exec: "cacheQuery",
  },
  submit: {
    executor: "constant-arrival-rate",
    rate: Number(__ENV.RATE || 20),
    timeUnit: "1s",
    duration: __ENV.DURATION || "30s",
    preAllocatedVUs: 20,
    maxVUs: 60,
    exec: "taskSubmit",
  },
  e2e: {
    executor: "constant-vus",
    vus: Number(__ENV.VUS || 10),
    duration: __ENV.DURATION || "30s",
    exec: "taskRoundTrip",
  },
};

export const options = {
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
  scenarios: { [scenario]: scenarios[scenario] },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    checks: ["rate>0.99"],
    ...(scenario === "cache" ? { http_req_duration: ["p(95)<100", "p(99)<200"] } : {}),
    ...(scenario === "submit" ? { http_req_duration: ["p(95)<200", "p(99)<400"] } : {}),
    ...(scenario === "e2e" ? { task_completed: ["rate>0.99"], task_e2e_duration: ["p(95)<2000"] } : {}),
  },
};

export function setup() {
  const stamp = String(Date.now());
  const credentials = { username: `perf_${stamp}`, password: "autopilot123" };
  const register = http.post(`${baseURL}/api/v1/auth/register`, JSON.stringify(credentials), jsonHeaders());
  check(register, { "压测用户创建成功": (response) => response.status === 201 });
  const token = register.json("token");
  const vin = `PF${stamp.padStart(15, "0").slice(-15)}`;
  const vehicle = http.post(
    `${baseURL}/api/v1/vehicles`,
    JSON.stringify({ vehicle_model: "AutoPilot Performance", vin }),
    authHeaders(token),
  );
  check(vehicle, { "压测车辆创建成功": (response) => response.status === 201 });
  return { token, vehicleId: vehicle.json("id") };
}

export function cacheQuery(data) {
  const response = http.get(`${baseURL}/api/v1/vehicles/${data.vehicleId}/status`, authHeaders(data.token));
  check(response, { "缓存查询返回 200": (item) => item.status === 200 });
}

export function taskSubmit(data) {
  const response = submitTask(data);
  check(response, { "任务提交返回 202": (item) => item.status === 202 });
}

export function taskRoundTrip(data) {
  // 使用单调运行时钟，避免宿主机校时导致端到端耗时出现负值。
  const started = exec.instance.currentTestRunDuration;
  const submitted = submitTask(data);
  if (submitted.status !== 202) {
    taskCompleted.add(false);
    return;
  }
  const taskID = submitted.json("id");
  for (let attempt = 0; attempt < 40; attempt += 1) {
    sleep(0.05);
    const task = http.get(`${baseURL}/api/v1/agent/tasks/${taskID}`, authHeaders(data.token));
    if (task.status === 200 && task.json("status") === "finished") {
      taskCompleted.add(true);
      taskE2E.add(Math.max(0, exec.instance.currentTestRunDuration - started));
      return;
    }
    if (task.status === 200 && task.json("status") === "failed") break;
  }
  taskCompleted.add(false);
}

function submitTask(data) {
  return http.post(
    `${baseURL}/api/v1/agent/tasks`,
    JSON.stringify({ vehicle_id: data.vehicleId, command: "将座舱温度调到22度" }),
    authHeaders(data.token),
  );
}

function jsonHeaders() {
  return { headers: { "Content-Type": "application/json" } };
}

function authHeaders(token) {
  return { headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` } };
}
