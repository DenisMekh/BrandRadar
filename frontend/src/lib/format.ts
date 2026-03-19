import { formatDistanceToNow, parseISO } from "date-fns";
import { ru } from "date-fns/locale";

/** Format number with locale separators: 1247 → "1 247" */
export function fmtNum(n: number | string | null | undefined): string {
  if (n == null) return "—";
  const num = typeof n === "string" ? Number(n) : n;
  if (isNaN(num)) return "—";
  return num.toLocaleString("ru-RU");
}

/** Relative time: "2 минуты назад" */
export function fmtRelative(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  try {
    const date = typeof dateStr === "string" ? parseISO(dateStr) : dateStr;
    return formatDistanceToNow(date, { addSuffix: true, locale: ru });
  } catch {
    return dateStr;
  }
}
