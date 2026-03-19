import { Database, Zap, Globe, Loader2, RefreshCw, ServerOff, HelpCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Skeleton } from "@/components/shared/Skeleton";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { useHealth } from "@/hooks/use-health";
import { useConnection } from "@/contexts/ConnectionContext";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtRelative } from "@/lib/format";

const iconMap: Record<string, typeof Database> = {
  PostgreSQL: Database, postgres: Database,
  Redis: Zap, redis: Zap,
  Collector: Globe, collector: Globe,
};

const statusMap = {
  ok:       { label: "Система работает", colorClass: "bg-green-500", ringClass: "bg-green-500/40", textClass: "text-green-400" },
  degraded: { label: "Деградация",       colorClass: "bg-yellow-500", ringClass: "bg-yellow-500/40", textClass: "text-yellow-400" },
  down:     { label: "Сбой",             colorClass: "bg-red-500",    ringClass: "bg-red-500/40",    textClass: "text-red-400" },
  unknown:  { label: "Неизвестно",       colorClass: "bg-zinc-600",   ringClass: "bg-zinc-600/40",   textClass: "text-zinc-400" },
};

const Health = () => {
  usePageTitle("Здоровье системы");

  const { data, isLoading, isError, refetch } = useHealth();
  const { isOffline } = useConnection();

  if (isError && !isOffline) return <ErrorBanner onRetry={() => refetch()} />;

  const st = isOffline ? statusMap.unknown : statusMap[data?.status ?? "ok"];

  return (
    <div className="space-y-10">
      {/* Offline banner */}
      {isOffline && (
        <div className="bg-destructive/10 border border-destructive/20 rounded-xl p-5 flex flex-col sm:flex-row items-start sm:items-center gap-4">
          <ServerOff className="h-6 w-6 text-destructive shrink-0" />
          <div className="flex-1">
            <p className="text-sm font-medium text-destructive">Не удаётся подключиться к серверу</p>
            <p className="text-xs text-destructive/70 mt-0.5">Проверьте URL в настройках. Текущий: {import.meta.env.VITE_API_URL || "http://localhost:8080"}</p>
          </div>
          <button
            onClick={() => refetch()}
            className="px-4 py-2 bg-destructive text-destructive-foreground text-sm rounded-lg hover:bg-destructive/90 transition-colors flex items-center gap-2 cursor-pointer shrink-0"
          >
            <RefreshCw className="h-4 w-4" /> Проверить подключение
          </button>
        </div>
      )}

      <div className="flex flex-col items-center gap-4 py-8">
        {isLoading ? (
          <><Skeleton className="w-[120px] h-[120px] rounded-full" /><Skeleton className="w-40 h-6" /><Skeleton className="w-32 h-4" /></>
        ) : (
          <>
            <div className="relative flex items-center justify-center">
              <div className={cn("absolute w-[120px] h-[120px] rounded-full animate-pulse-ring", st.ringClass)} />
              <div className={cn("w-[120px] h-[120px] rounded-full flex items-center justify-center shadow-lg", st.colorClass)}>
                <span className="text-3xl font-bold text-white">
                  {isOffline ? "?" : data?.status === "ok" ? "✓" : data?.status === "degraded" ? "!" : "✕"}
                </span>
              </div>
            </div>
            <div className="text-center space-y-1">
              <p className={cn("text-xl font-semibold", st.textClass)}>{st.label}</p>
              <p className="text-sm text-muted-foreground">Uptime: {isOffline ? "—" : data?.uptime_seconds ? `${data.uptime_seconds}s` : "—"}</p>
              <p className="text-xs text-muted-foreground/50">Версия: {isOffline ? "—" : data?.version ?? "—"}</p>
            </div>
          </>
        )}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {isLoading
          ? Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="bg-card border border-border rounded-xl p-5 flex flex-col items-center gap-3">
                <Skeleton className="w-8 h-8 rounded" /><Skeleton className="w-20 h-4" /><Skeleton className="w-16 h-4" />
              </div>
            ))
          : isOffline
          ? /* Show unknown state for all deps when offline */
            ["PostgreSQL", "Redis", "Collector"].map((name) => {
              const Icon = iconMap[name] ?? Globe;
              return (
                <div key={name} className="bg-card border border-border rounded-xl p-5 flex flex-col items-center gap-3 opacity-60">
                  <Icon className="h-8 w-8 text-muted-foreground" />
                  <p className="font-medium text-foreground">{name}</p>
                  <div className="flex items-center gap-2">
                    <HelpCircle className="h-4 w-4 text-zinc-500" />
                    <span className="text-sm text-zinc-500">Неизвестно</span>
                  </div>
                </div>
              );
            })
          : Object.entries(data?.dependencies ?? {}).map(([name, status]) => {
              const Icon = iconMap[name] ?? Globe;
              const online = status === "ok";
              return (
                <div key={name} className="bg-card border border-border rounded-xl p-5 flex flex-col items-center gap-3 hover:scale-[1.01] transition-transform duration-200">
                  <Icon className="h-8 w-8 text-muted-foreground" />
                  <p className="font-medium text-foreground">{name}</p>
                  <div className="flex items-center gap-2">
                    <span className={cn("w-2.5 h-2.5 rounded-full", online ? "bg-green-500 animate-pulse" : "bg-red-500")} />
                    <span className={cn("text-sm", online ? "text-green-400" : "text-red-400")}>{online ? "Online" : "Offline"}</span>
                  </div>
                </div>
              );
            })}
      </div>

      <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground/50">
        <Loader2 className="h-3 w-3 animate-spin-slow" />Автообновление каждые 30 сек
      </div>
    </div>
  );
};

export default Health;
