import { useState } from "react";
import { cn } from "@/lib/utils";
import { useHealth } from "@/hooks/use-health";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { HealthStatus } from "@/lib/api/health";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { fmtRelative } from "@/lib/format";

const statusColors = {
  ok: "bg-success",
  degraded: "bg-warning",
  down: "bg-destructive",
};

export function HealthIndicator({ compact = false }: { compact?: boolean }) {
  const { data: health } = useHealth();
  const [dialogOpen, setDialogOpen] = useState(false);

  const status = health?.status ?? "down";
  const dotColor = statusColors[status];
  const depsMap = health?.dependencies ?? {};
  const depsEntries = Object.entries(depsMap);

  const tooltipText = depsEntries.length > 0
    ? depsEntries.map(([name, status]) => `${name}: ${status === "ok" ? "✓" : "✗"}`).join(" | ") + ` | Uptime: ${health?.uptime_seconds ?? "—"}s`
    : "Нет данных";

  if (compact) {
    return (
      <>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              onClick={() => setDialogOpen(true)}
              className="flex items-center justify-center cursor-pointer p-1"
              aria-label="Статус системы"
            >
              <span className={cn("w-2 h-2 rounded-full shrink-0", dotColor, status === "ok" && "animate-pulse")} />
            </button>
          </TooltipTrigger>
          <TooltipContent side="right" className="text-xs max-w-[280px]">{tooltipText}</TooltipContent>
        </Tooltip>
        <HealthDialog open={dialogOpen} onOpenChange={setDialogOpen} health={health} />
      </>
    );
  }

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            onClick={() => setDialogOpen(true)}
            className="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity"
            aria-label="Статус системы"
          >
            <span className={cn("w-2 h-2 rounded-full shrink-0", dotColor, status === "ok" && "animate-pulse")} />
            <span className="text-xs text-muted-foreground/50">Система</span>
          </button>
        </TooltipTrigger>
        <TooltipContent className="text-xs max-w-[280px]">{tooltipText}</TooltipContent>
      </Tooltip>
      <HealthDialog open={dialogOpen} onOpenChange={setDialogOpen} health={health} />
    </>
  );
}

function HealthDialog({ open, onOpenChange, health }: { open: boolean; onOpenChange: (v: boolean) => void; health: HealthStatus | undefined }) {
  const status = health?.status ?? "down";
  const depsMap = health?.dependencies ?? {};
  const depsEntries = Object.entries(depsMap);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-card border-border sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-foreground">Статус системы</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="flex items-center gap-3">
            <span className={cn("w-3 h-3 rounded-full", statusColors[status as keyof typeof statusColors] ?? "bg-destructive")} />
            <span className="text-sm font-medium text-foreground capitalize">{status === "ok" ? "Всё работает" : status === "degraded" ? "Деградация" : "Недоступна"}</span>
          </div>

          <div className="space-y-2">
            <p className="text-xs text-muted-foreground uppercase tracking-wide font-medium">Зависимости</p>
            {depsEntries.map(([name, depStatus]) => (
              <div key={name} className="flex items-center justify-between py-1.5 border-b border-border last:border-0">
                <span className="text-sm text-foreground">{name}</span>
                <div className="flex items-center gap-2">
                  <span className={cn("text-xs font-medium", depStatus === "ok" ? "text-success" : "text-destructive")}>{depStatus === "ok" ? "✓ Online" : "✗ Degraded"}</span>
                </div>
              </div>
            ))}
          </div>

          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Версия</span>
            <span className="text-foreground font-mono text-xs">{health?.version ?? "—"}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Uptime</span>
            <span className="text-foreground">{health?.uptime_seconds ? `${health.uptime_seconds}s` : "—"}</span>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}