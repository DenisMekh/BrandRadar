import { useState, useMemo, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { ChevronDown, Check } from "lucide-react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";
import { useBrands } from "@/hooks/use-brands";
import { useMentions } from "@/hooks/use-mentions";
import { subDays, format, startOfDay, parseISO, addMinutes } from "date-fns";
import { ru } from "date-fns/locale";
import { Skeleton } from "@/components/shared/Skeleton";

const periods = [
  { label: "1 день", value: 1 },
  { label: "7 дней", value: 7 },
  { label: "14 дней", value: 14 },
];

const spans = [
  { label: "15 мин", value: 15 },
  { label: "30 мин", value: 30 },
  { label: "1 час", value: 60 },
  { label: "6 часов", value: 360 },
  { label: "1 день", value: 1440 },
];

const LINE_COLORS = [
  "hsl(263 70% 58%)",   // violet
  "hsl(292 84% 61%)",   // fuchsia
  "hsl(217 91% 60%)",   // blue
  "hsl(160 84% 39%)",   // emerald
  "hsl(38 92% 50%)",    // amber
  "hsl(347 77% 50%)",   // rose
];

interface TooltipEntry {
  name: string;
  value: number;
  color: string;
  dataKey: string;
}

const CustomTooltip = ({ active, payload, label }: { active?: boolean; payload?: TooltipEntry[]; label?: string }) => {
  if (!active || !payload?.length) return null;
  return (
    <div className="bg-card border border-border rounded-lg p-3 text-sm shadow-lg">
      <p className="text-muted-foreground mb-1.5">{label}</p>
      {payload.map((p) => (
        <p key={p.dataKey} className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full shrink-0" style={{ background: p.color }} />
          <span className="text-foreground">{p.name}: {p.value?.toLocaleString("ru-RU")}</span>
        </p>
      ))}
      <p className="text-[10px] text-muted-foreground/50 mt-1.5">Нажмите для деталей</p>
    </div>
  );
};

export function LiveMentionsChart({ 
  period: propsPeriod, 
  onPeriodChange 
}: { 
  period?: number; 
  onPeriodChange?: (p: number) => void 
}) {
  const [localPeriod, setLocalPeriod] = useState(1);
  const period = propsPeriod !== undefined ? propsPeriod : localPeriod;
  const setPeriod = (p: number) => {
    if (onPeriodChange) onPeriodChange(p);
    else setLocalPeriod(p);
  };

  const [span, setSpan] = useState(60); // Default bucket: 1 hour
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [hoveredLegend, setHoveredLegend] = useState<string | null>(null);
  const { data: brands } = useBrands();
  const navigate = useNavigate();
  const dropdownRef = useRef<HTMLDivElement>(null);

  const brandList = useMemo(() => brands ?? [], [brands]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const dateFrom = useMemo(() => subDays(new Date(), period).toISOString(), [period]);

  const mentionsQuery = useMentions({
    date_from: dateFrom,
    brand_ids: selectedIds.length > 0 ? selectedIds : undefined,
    limit: 2000,
  });

  const visibleBrands = useMemo(() => {
    if (selectedIds.length === 0) return brandList;
    return brandList.filter((b) => selectedIds.includes(b.id));
  }, [brandList, selectedIds]);

  const chartData = useMemo(() => {
    const mentions = mentionsQuery.data?.data ?? [];
    if (visibleBrands.length === 0) return [];

    const now = new Date();
    const start = subDays(now, period);

    // Align the start to a clean bucket boundary
    const current = new Date(start);
    if (span >= 1440) {
      // daily bucket → align to day start
      current.setHours(0, 0, 0, 0);
    } else {
      const mins = current.getMinutes();
      current.setMinutes(mins - (mins % span), 0, 0);
    }

    // Pre-parse mention timestamps once
    const parsed = mentions.map((m) => ({
      brandId: m.brand_id ?? "",
      ts: parseISO(m.created_at).getTime(),
    }));

    const points: Record<string, string | number | Date>[] = [];

    while (current <= now) {
      const bucketStart = current.getTime();
      const bucketEnd = addMinutes(current, span).getTime();

      const label = span < 1440
        ? format(current, "HH:mm", { locale: ru })
        : format(current, "d MMM", { locale: ru });

      const point: Record<string, string | number | Date> = { date: label, rawDate: new Date(current) };

      visibleBrands.forEach((b) => {
        point[b.name] = parsed.filter(
          (p) => (p.brandId === b.id || p.brandId.includes(b.id)) && p.ts >= bucketStart && p.ts < bucketEnd
        ).length;
      });

      points.push(point);
      current.setTime(addMinutes(current, span).getTime());
    }

    return points;
  }, [period, span, visibleBrands, mentionsQuery.data?.data]);

  const handleChartClick = (data: { activePayload?: { payload: Record<string, unknown> }[] }) => {
    const rawDate = data?.activePayload?.[0]?.payload?.rawDate;
    if (rawDate instanceof Date) {
      const end = addMinutes(rawDate, span);
      navigate(`/mentions?date_from=${rawDate.toISOString()}&date_to=${end.toISOString()}`);
    }
  };

  const toggleBrand = (id: string) => {
    setSelectedIds((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]);
  };

  const subtitle = selectedIds.length === 0
    ? "Все бренды"
    : visibleBrands.map((b) => b.name).join(", ");

  return (
    <>
      <div className="flex flex-col gap-2 mb-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium text-foreground">Детальная динамика</h3>
            <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>
          </div>

          {/* Desktop controls */}
          <div className="hidden sm:flex items-center gap-2">
            {/* Brand picker */}
            <div className="relative" ref={dropdownRef}>
              <button
                onClick={() => setDropdownOpen(!dropdownOpen)}
                className="flex items-center gap-1.5 bg-secondary border border-border text-sm rounded-lg px-3 py-1.5 hover:bg-secondary/80 transition-colors cursor-pointer min-w-[140px]"
              >
                <span className="text-muted-foreground truncate">
                  {selectedIds.length === 0 ? "Все бренды" : `${visibleBrands.length} выбрано`}
                </span>
                <ChevronDown className="h-3.5 w-3.5 text-muted-foreground ml-auto shrink-0" />
              </button>
              {dropdownOpen && (
                <div className="absolute right-0 top-full mt-1 z-50 bg-card border border-border rounded-lg shadow-lg py-1 min-w-[200px] max-h-[240px] overflow-y-auto">
                  {brandList.map((b) => {
                    const sel = selectedIds.includes(b.id);
                    return (
                      <button key={b.id} onClick={() => toggleBrand(b.id)} className="w-full flex items-center gap-2 px-3 py-2 text-sm text-left hover:bg-secondary/60 transition-colors cursor-pointer">
                        <span className={`w-4 h-4 rounded border flex items-center justify-center shrink-0 ${sel ? "bg-primary border-primary" : "border-border"}`}>
                          {sel && <Check className="h-3 w-3 text-primary-foreground" />}
                        </span>
                        <span className="text-foreground truncate">{b.name}</span>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Span selector */}
            <select
              value={span}
              onChange={(e) => setSpan(Number(e.target.value))}
              className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-1.5 focus:outline-none cursor-pointer"
            >
              {spans.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
            </select>

            {/* Period selector */}
            <select
              value={period}
              onChange={(e) => setPeriod(Number(e.target.value))}
              className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-1.5 focus:outline-none cursor-pointer"
            >
              {periods.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </div>
        </div>

        {/* Mobile controls */}
        <div className="flex sm:hidden items-center gap-2">
          <select
            value={span}
            onChange={(e) => setSpan(Number(e.target.value))}
            className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none cursor-pointer flex-1"
          >
            {spans.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
          </select>
          <select
            value={period}
            onChange={(e) => setPeriod(Number(e.target.value))}
            className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none cursor-pointer"
          >
            {periods.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
          </select>
        </div>
      </div>

      {/* Chart area with explicit height */}
      <div className="h-[200px] sm:h-[280px] relative">
        {mentionsQuery.isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
            <Skeleton className="w-full h-full" />
          </div>
        )}
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData} onClick={handleChartClick} style={{ cursor: "pointer" }}>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(240 3.7% 15.9%)" />
            <XAxis
              dataKey="date"
              tick={{ fill: "hsl(240 5% 64.9%)", fontSize: 11 }}
              axisLine={{ stroke: "hsl(240 3.7% 15.9%)" }}
              tickLine={false}
              interval="preserveStartEnd"
              minTickGap={40}
            />
            <YAxis tick={{ fill: "hsl(240 5% 64.9%)", fontSize: 11 }} axisLine={false} tickLine={false} width={30} />
            <Tooltip content={<CustomTooltip />} />
            {visibleBrands.map((b, i) => (
              <Line
                key={b.id}
                type="monotone"
                dataKey={b.name}
                stroke={LINE_COLORS[i % LINE_COLORS.length]}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 5, cursor: "pointer" }}
                strokeOpacity={hoveredLegend && hoveredLegend !== b.name ? 0.2 : 1}
                isAnimationActive
                animationDuration={800}
                animationBegin={i * 100}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-x-3 sm:gap-x-4 gap-y-1 mt-3 justify-center">
        {visibleBrands.map((b, i) => (
          <button
            key={b.id}
            className="flex items-center gap-1 sm:gap-1.5 text-[11px] sm:text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer py-1"
            onMouseEnter={() => setHoveredLegend(b.name)}
            onMouseLeave={() => setHoveredLegend(null)}
          >
            <span className="w-2 h-2 rounded-full shrink-0" style={{ background: LINE_COLORS[i % LINE_COLORS.length] }} />
            {b.name}
          </button>
        ))}
      </div>
    </>
  );
}
