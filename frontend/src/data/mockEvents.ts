export type EventType =
  | "mention_created"
  | "mention_duplicated"
  | "spike_detected"
  | "alert_fired"
  | "collector_started"
  | "collector_stopped"
  | "collector_failed"
  | "source_toggled";

export interface AppEvent {
  id: string;
  type: EventType;
  payload: Record<string, string | number>;
  time: string;
}

export const eventTypeConfig: Record<EventType, { label: string; colorClass: string; dotClass: string }> = {
  mention_created:    { label: "Создано",          colorClass: "bg-blue-500/20 text-blue-400",    dotClass: "bg-blue-500" },
  mention_duplicated: { label: "Дубликат",         colorClass: "bg-zinc-500/20 text-zinc-400",    dotClass: "bg-zinc-500" },
  spike_detected:     { label: "Всплеск",          colorClass: "bg-red-500/20 text-red-400",      dotClass: "bg-red-500" },
  alert_fired:        { label: "Алерт",            colorClass: "bg-amber-500/20 text-amber-400",  dotClass: "bg-amber-500" },
  collector_started:  { label: "Сбор начат",       colorClass: "bg-green-500/20 text-green-400",  dotClass: "bg-green-500" },
  collector_stopped:  { label: "Сбор остановлен",  colorClass: "bg-yellow-500/20 text-yellow-400",dotClass: "bg-yellow-500" },
  collector_failed:   { label: "Ошибка сбора",     colorClass: "bg-red-500/20 text-red-400",      dotClass: "bg-red-500" },
  source_toggled:     { label: "Источник",          colorClass: "bg-violet-500/20 text-violet-400",dotClass: "bg-violet-500" },
};

export function formatEventPayload(event: AppEvent): string {
  const p = event.payload;
  switch (event.type) {
    case "mention_created":
      return `Создано упоминание для бренда «${p.brand}»`;
    case "mention_duplicated":
      return `Обнаружен дубликат упоминания для бренда «${p.brand}»`;
    case "spike_detected":
      return `Всплеск: ${p.count} упоминаний бренда «${p.brand}»`;
    case "alert_fired":
      return `Алерт сработал для бренда «${p.brand}»: ${p.count} упоминаний`;
    case "collector_started":
      return `Запущен сбор данных из источника «${p.source}»`;
    case "collector_stopped":
      return `Сбор данных остановлен: «${p.source}»`;
    case "collector_failed":
      return `Ошибка сбора: ${p.source} — ${p.error}`;
    case "source_toggled":
      return `Источник «${p.source}» ${p.enabled === 1 ? "включён" : "отключён"}`;
    default:
      return JSON.stringify(p);
  }
}

export const mockEvents: AppEvent[] = [
  { id: "e1",  type: "spike_detected",     payload: { brand: "BrandRadar", count: 15 }, time: "5 мин назад" },
  { id: "e2",  type: "alert_fired",        payload: { brand: "BrandRadar", count: 15 }, time: "5 мин назад" },
  { id: "e3",  type: "mention_created",    payload: { brand: "BrandRadar" }, time: "8 мин назад" },
  { id: "e4",  type: "mention_created",    payload: { brand: "Конкурент А" }, time: "12 мин назад" },
  { id: "e5",  type: "mention_duplicated", payload: { brand: "BrandRadar" }, time: "15 мин назад" },
  { id: "e6",  type: "collector_started",  payload: { source: "Telegram Bot #1" }, time: "20 мин назад" },
  { id: "e7",  type: "mention_created",    payload: { brand: "Продукт X" }, time: "34 мин назад" },
  { id: "e8",  type: "collector_failed",   payload: { source: "RSS Parser", error: "Connection timeout" }, time: "45 мин назад" },
  { id: "e9",  type: "source_toggled",     payload: { source: "Twitter API", enabled: 0 }, time: "1ч назад" },
  { id: "e10", type: "spike_detected",     payload: { brand: "Конкурент А", count: 22 }, time: "2ч назад" },
  { id: "e11", type: "alert_fired",        payload: { brand: "Конкурент А", count: 22 }, time: "2ч назад" },
  { id: "e12", type: "collector_stopped",  payload: { source: "Web Crawler #3" }, time: "3ч назад" },
  { id: "e13", type: "mention_created",    payload: { brand: "BrandRadar" }, time: "5ч назад" },
  { id: "e14", type: "source_toggled",     payload: { source: "VK API", enabled: 1 }, time: "8ч назад" },
  { id: "e15", type: "collector_started",  payload: { source: "Web Crawler #3" }, time: "12ч назад" },
];
