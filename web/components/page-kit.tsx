import { ReactNode } from "react";
export function PageHeader({ title, description, action }: { title: string; description: string; action?: ReactNode }) { return <div className="page-header"><div><h1>{title}</h1><p>{description}</p></div>{action}</div>; }
export function Badge({ children, tone = "blue" }: { children: ReactNode; tone?: string }) { return <span className={`badge ${tone}`}>{children}</span>; }
export function Metric({ label, value, meta, icon }: { label: string; value: string; meta: string; icon: ReactNode }) { return <article className="metric"><div className="metric-icon">{icon}</div><span>{label}</span><strong>{value}</strong><small>{meta}</small></article>; }
