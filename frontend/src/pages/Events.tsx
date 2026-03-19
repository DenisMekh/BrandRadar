import { useState, useMemo } from "react";
import { cn } from "@/lib/utils";
import { Skeleton } from "@/components/shared/Skeleton";
import { EmptyState } from "@/components/shared/EmptyState";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { useEvents } from "@/hooks/use-events";
import { useSources } from "@/hooks/use-sources";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum, fmtRelative } from "@/lib/format";
import type { Brand } from "@/lib/api/brands";
import { useBrands } from "@/hooks/use-brands";

type EventType = "mention_created" | "mention_duplicated" | "spike_detected" | "alert_fired" | "collector_started" | "collector_stopped" | "collector_failed" | "source_toggled";

const eventTypeConfig: Record<EventType, { label: string; colorClass: string; dotClass: string }> = {
  mention_created:    { label: "Создано",          colorClass: "bg-blue-500/20 text-blue-400",    dotClass: "bg-blue-500" },
  mention_duplicated: { label: "Дубликат",         colorClass: "bg-zinc-500/20 text-zinc-400",    dotClass: "bg-zinc-500" },
  spike_detected:     { label: "Всплеск",          colorClass: "bg-red-500/20 text-red-400",      dotClass: "bg-red-500" },
  alert_fired:        { label: "Алерт",            colorClass: "bg-amber-500/20 text-amber-400",  dotClass: "bg-amber-500" },
  collector_started:  { label: "Сбор начат",       colorClass: "bg-green-500/20 text-green-400",  dotClass: "bg-green-500" },
  collector_stopped:  { label: "Сбор остановлен",  colorClass: "bg-yellow-500/20 text-yellow-400",dotClass: "bg-yellow-500" },
  collector_failed:   { label: "Ошибка сбора",     colorClass: "bg-red-500/20 text-red-400",      dotClass: "bg-red-500" },
  source_toggled:     { label: "Источник",          colorClass: "bg-violet-500/20 text-violet-400",dotClass: "bg-violet-500" },
};

function formatPayload(
  type: string, 
  payload: Record<string, string | number | undefined>, 
  sourcesMap: Record<string, string>,
  brandsMap: Record<string, string>
): string {
  const p = payload;
  
  const getSourceName = () => {
    const raw = (p.source || p.source_name || p.source_id) as string | undefined;
    if (!raw) return "Неизвестный источник";
    return sourcesMap[raw] || raw;
  };

  const getBrandName = () => {
    const raw = (p.brand || p.brand_name || p.brand_id) as string | undefined;
    if (!raw) return "Неизвестный бренд";
    return brandsMap[raw] || raw;
  };

  switch (type) {
    case "mention_created": return `Создано упоминание для бренда «${getBrandName()}»`;
    case "mention_duplicated": return `Обнаружен дубликат для бренда «${getBrandName()}»`;
    case "spike_detected": return `Всплеск: ${fmtNum(p.count as number)} упоминаний бренда «${getBrandName()}»`;
    case "alert_fired": return `Алерт сработал для бренда «${getBrandName()}»: ${fmtNum(p.count as number)} упоминаний`;
    case "collector_started": return `Запущен сбор из «${getSourceName()}»`;
    case "collector_stopped": return `Сбор остановлен: «${getSourceName()}»`;
    case "collector_failed": return `Ошибка сбора: ${getSourceName()} — ${p.error}`;
    case "source_toggled": {
      const isEnabled = String(p.enabled) === "1" || String(p.enabled) === "true";
      return `Источник «${getSourceName()}» ${isEnabled ? "включён" : "отключён"}`;
    }
    default: return JSON.stringify(p);
  }
}

interface AppEvent {
  id: string;
  type: EventType;
  occurred_at: string;
  payload: string;
}

// Tags visible in the filter bar (excluding duplicates, spikes, and collector events)
const visibleTypes: EventType[] = ["mention_created", "alert_fired", "source_toggled"];
const ITEMS_STEP = 15;

const Events = () => {
  usePageTitle("Журнал событий");

  // Empty set = show ALL events; non-empty = show only selected types (OR logic)
  const [activeTypes, setActiveTypes] = useState<Set<EventType>>(new Set());
  const [visibleCount, setVisibleCount] = useState(ITEMS_STEP);

  const { data: sources } = useSources();
  const { data: brands } = useBrands();

  const sourcesMap = useMemo(() => {
    const map: Record<string, string> = {};
    (sources ?? []).forEach(s => { if (s.id) map[s.id] = s.name ?? ""; });
    return map;
  }, [sources]);

  const brandsMap = useMemo(() => {
    const map: Record<string, string> = {};
    (brands ?? []).forEach(b => { if (b.id) map[b.id] = b.name ?? ""; });
    return map;
  }, [brands]);

  // Always fetch all events; filtering is done client-side for correct OR logic with multiple tags
  const { data: events, isLoading, isError, refetch } = useEvents(undefined, 100);

  const toggleType = (t: EventType) => {
    setActiveTypes((prev) => {
      const next = new Set(prev);
      if (next.has(t)) { next.delete(t); } else { next.add(t); }
      return next;
    });
    setVisibleCount(ITEMS_STEP);
  };

  const resetFilters = () => {
    setActiveTypes(new Set());
    setVisibleCount(ITEMS_STEP);
  };

  const filtered = useMemo(() => {
    // If no tags selected, show everything; otherwise show events matching any selected tag
    if (activeTypes.size === 0) return ((events as AppEvent[]) ?? []);
    return ((events as AppEvent[]) ?? []).filter((e) => e.type !== undefined && activeTypes.has(e.type as EventType));
  }, [events, activeTypes]);
  const visible = filtered.slice(0, visibleCount);
  const hasMore = visibleCount < filtered.length;
  const hasFiltersChanged = activeTypes.size > 0;

  return (
    <div className="space-y-4 sm:space-y-6">
      {/* Filter chips - horizontal scroll on mobile */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 -mx-4 px-4 sm:mx-0 sm:px-0 sm:flex-wrap">
        {visibleTypes.map((t) => {
          const cfg = eventTypeConfig[t];
          const active = activeTypes.has(t);
          return (
            <button key={t} onClick={() => toggleType(t)} className={cn("text-xs font-medium px-3 py-1.5 rounded-full border transition-all duration-200 cursor-pointer whitespace-nowrap min-h-[36px]", active ? `${cfg.colorClass} border-transparent` : "bg-secondary text-muted-foreground/50 border-border hover:border-muted-foreground/40")}>{cfg.label}</button>
          );
        })}
        {hasFiltersChanged && (
          <button onClick={resetFilters} className="text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer whitespace-nowrap ml-auto shrink-0">
            Сбросить
          </button>
        )}
      </div>

      {isError && <ErrorBanner onRetry={() => refetch()} />}

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-card border border-border rounded-xl p-4 space-y-2">
              <div className="flex gap-2"><Skeleton className="w-16 h-4 rounded-full" /><Skeleton className="w-20 h-4 ml-auto" /></div>
              <Skeleton className={`h-4 ${i % 2 === 0 ? "w-3/4" : "w-1/2"}`} />
            </div>
          ))}
        </div>
      ) : (
        <>
          {/* Desktop: timeline with line */}
          <div className="hidden sm:block relative pl-6">
            <div className="absolute left-[9px] top-2 bottom-2 w-0.5 bg-border" />
            <div className="space-y-4">
              {visible.map((event) => {
                const cfg = eventTypeConfig[event.type as EventType] || { label: "Неизвестно", colorClass: "bg-gray-500/20 text-gray-400", dotClass: "bg-gray-500" };
                let parsedPayload = {};
                try {
                  parsedPayload = typeof event.payload === 'string' && event.payload.trim() !== '' ? JSON.parse(event.payload) : (event.payload || {});
                } catch (e) {
                  // ignore invalid json
                }
                return (
                  <div key={event.id} className="relative flex gap-4">
                    <div className={cn("absolute -left-6 top-3 w-[18px] h-[18px] rounded-full border-[3px] border-background", cfg.dotClass)} />
                    <div className="flex-1 bg-card border border-border rounded-xl p-4 hover:scale-[1.005] transition-transform duration-200">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className={cn("text-[11px] font-medium px-2 py-0.5 rounded-full", cfg.colorClass)}>{cfg.label}</span>
                        <span className="text-xs text-muted-foreground/50 ml-auto">{fmtRelative(event.occurred_at || "")}</span>
                      </div>
                      <p className="text-sm text-foreground mt-2">{formatPayload(event.type || "", parsedPayload, sourcesMap, brandsMap)}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Mobile: simple cards without timeline line */}
          <div className="sm:hidden space-y-3">
            {visible.map((event) => {
              const cfg = eventTypeConfig[event.type as EventType] || { label: "Неизвестно", colorClass: "bg-gray-500/20 text-gray-400", dotClass: "bg-gray-500" };
              let parsedPayload = {};
              try {
                parsedPayload = typeof event.payload === 'string' && event.payload.trim() !== '' ? JSON.parse(event.payload) : (event.payload || {});
              } catch (e) {
                // ignore invalid json
              }
              return (
                <div key={event.id} className="bg-card border border-border rounded-xl p-4">
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className={cn("w-2.5 h-2.5 rounded-full shrink-0", cfg.dotClass)} />
                    <span className={cn("text-[11px] font-medium px-2 py-0.5 rounded-full", cfg.colorClass)}>{cfg.label}</span>
                    <span className="text-xs text-muted-foreground/50 ml-auto">{fmtRelative(event.occurred_at || "")}</span>
                  </div>
                  <p className="text-sm text-foreground mt-2">{formatPayload(event.type || "", parsedPayload, sourcesMap, brandsMap)}</p>
                </div>
              );
            })}
          </div>

          {visible.length === 0 && !isLoading && <EmptyState title="Журнал событий пуст" description="Здесь будут отображаться все события системы." />}

          {hasMore && (
            <div className="text-center">
              <button onClick={() => setVisibleCount((c) => c + ITEMS_STEP)} className="text-sm text-primary hover:text-primary/80 transition-colors cursor-pointer min-h-[44px] px-4">Загрузить ещё</button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default Events;
