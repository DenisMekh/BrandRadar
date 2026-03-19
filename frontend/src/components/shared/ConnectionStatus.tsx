import { AlertTriangle, WifiOff, Clock } from "lucide-react";
import { useConnection } from "@/contexts/ConnectionContext";
import { getCache, fmtCacheTime } from "@/lib/api-cache";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

export function ConnectionStatus() {
  const { state, failedDeps } = useConnection();

  if (state === "online") return null;

  const cached = getCache("health");
  const cacheTimeStr = cached ? fmtCacheTime(cached.ts) : null;

  if (state === "offline") {
    return (
      <div className="bg-destructive/10 border-b border-destructive/20 px-4 py-2.5 flex items-center gap-3 text-sm">
        <WifiOff className="h-4 w-4 text-destructive shrink-0" />
        <span className="text-destructive font-medium flex-1">
          Сервер недоступен. Показаны кэшированные данные.
        </span>
        {cacheTimeStr && (
          <span className="text-destructive/60 flex items-center gap-1 shrink-0">
            <Clock className="h-3 w-3" /> Данные от {cacheTimeStr}
          </span>
        )}
      </div>
    );
  }

  // degraded
  return (
    <div className="bg-warning/10 border-b border-warning/20 px-4 py-2.5 flex items-center gap-3 text-sm">
      <AlertTriangle className="h-4 w-4 text-warning shrink-0" />
      <span className="text-warning font-medium flex-1">
        ⚠ Часть сервисов недоступна. Данные могут быть неполными.
      </span>
      {failedDeps.length > 0 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <span className="text-warning/70 underline underline-offset-2 cursor-help shrink-0">
              {failedDeps.length} {failedDeps.length === 1 ? "сервис" : "сервисов"} недоступно
            </span>
          </TooltipTrigger>
          <TooltipContent side="bottom">
            <ul className="text-xs space-y-0.5">
              {failedDeps.map((d) => (
                <li key={d}>• {d}</li>
              ))}
            </ul>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );
}
