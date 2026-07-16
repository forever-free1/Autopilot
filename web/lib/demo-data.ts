export const trend = [
  { time: "08:00", success: 18, failed: 2 }, { time: "10:00", success: 31, failed: 3 },
  { time: "12:00", success: 27, failed: 2 }, { time: "14:00", success: 44, failed: 4 },
  { time: "16:00", success: 51, failed: 3 }, { time: "18:00", success: 38, failed: 2 },
];
export const tasks = [
  { id: "tsk_8f21", request: "规划最近充电站路线", vehicle: "沪A·A1027", status: "completed", step: "导航已启动", tools: 4, duration: "1.8s", created: "14:32:18" },
  { id: "tsk_8f20", request: "诊断动力电池告警", vehicle: "沪A·D8831", status: "tool_calling", step: "检索维修手册", tools: 3, duration: "进行中", created: "14:31:42" },
  { id: "tsk_8f19", request: "座舱温度调至 22°C", vehicle: "沪A·A1027", status: "completed", step: "状态已确认", tools: 2, duration: "0.9s", created: "14:29:06" },
  { id: "tsk_8f18", request: "查询剩余续航", vehicle: "沪A·B7719", status: "failed", step: "车辆离线", tools: 1, duration: "3.0s", created: "14:27:51" },
];
export const vehicles = [
  { id: 1, plate: "沪A·A1027", model: "AutoPilot S7", vin: "APDEMO20260000001", battery: 72, range: 386, status: "在线", health: "良好", location: "上海市浦东新区" },
  { id: 2, plate: "沪A·D8831", model: "AutoPilot X9", vin: "APDEMO20260000002", battery: 41, range: 206, status: "在线", health: "需检查", location: "上海市徐汇区" },
  { id: 3, plate: "沪A·B7719", model: "AutoPilot S7", vin: "APDEMO20260000003", battery: 88, range: 462, status: "离线", health: "良好", location: "上海市静安区" },
];
