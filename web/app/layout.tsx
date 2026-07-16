import type { Metadata } from "next";
import "./globals.css";
import { AppShell } from "@/components/app-shell";
import { Providers } from "@/components/providers";

export const metadata: Metadata = { title: "AutoPilot Agent Console", description: "智能座舱 Agent 管理与演示平台" };

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="zh-CN"><body><Providers><AppShell>{children}</AppShell></Providers></body></html>;
}
