"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Activity, Bell, Bot, CarFront, ChevronDown, CircleGauge, ClipboardList, Map, Search, Stethoscope } from "lucide-react";

const nav = [
  ["/", "系统总览", CircleGauge], ["/agent", "智能座舱 Agent", Bot], ["/vehicles", "车辆管理", CarFront],
  ["/diagnostics", "故障诊断", Stethoscope], ["/trips", "行程规划", Map], ["/tasks", "任务中心", ClipboardList],
] as const;

export function AppShell({ children }: { children: React.ReactNode }) {
  const path = usePathname();
  return <div className="app-shell">
    <aside className="sidebar">
      <div className="brand"><span className="brand-mark"><Activity size={18}/></span><div><b>AutoPilot</b><small>Agent Console</small></div></div>
      <nav>{nav.map(([href, label, Icon]) => <Link key={href} href={href} className={path === href || href !== "/" && path.startsWith(href) ? "active" : ""}><Icon size={18}/><span>{label}</span></Link>)}</nav>
      <div className="environment"><span><i/>服务运行中</span><small>7 个容器 · 本地环境</small></div>
    </aside>
    <div className="main-frame">
      <header className="topbar"><div className="search"><Search size={17}/><span>搜索车辆、任务或会话</span><kbd>⌘ K</kbd></div><div className="top-actions"><button aria-label="通知"><Bell size={18}/><i/></button><div className="avatar">FF</div><span>Demo User</span><ChevronDown size={15}/></div></header>
      <main className="page">{children}</main>
    </div>
  </div>;
}
