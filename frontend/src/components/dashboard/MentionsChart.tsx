import { useState, useMemo, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { X, ChevronDown, Check } from "lucide-react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";
import { useBrands } from "@/hooks/use-brands";
import { useQueries } from "@tanstack/react-query";
import { getBrandDashboard } from "@/lib/api/brands";
import { subDays, format, startOfDay } from "date-fns";
import { ru } from "date-fns/locale";
import { Skeleton } from "@/components/shared/Skeleton";

const periods = [
  { label: "1 день", value: 1 },
  { label: "7 дней", value: 7 },
  { label: "14 дней", value: 14 },
  { label: "30 дней", value: 30 },
];

const LINE_COLORS = [
  "hsl(263 70% 58%)",   // violet
  "hsl(292 84% 61%)",   // fuchsia
  "hsl(217 91% 60%)",   // blue
  "hsl(160 84% 39%)",   // emerald
  "hsl(38 92% 50%)",    // amber
  "hsl(347 77% 50%)",   // rose
];


const CustomTooltip = ({ active, payload, label }: { active?: boolean; payload?: { name: string; value: number; color: string; dataKey: string; payload: { rawDate: Date } }[]; label?: string }) => {
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

interface MentionsChartProps {
  selectedIds: string[];
  onChangeSelectedIds: (ids: string[]) => void;
  period: number;
  onPeriodChange: (p: number) => void;
  showSelection?: boolean;
}

export function MentionsChart({ selectedIds, onChangeSelectedIds, period, onPeriodChange, showSelection = true }: MentionsChartProps) {
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [hoveredLegend, setHoveredLegend] = useState<string | null>(null);
  const { data: brands } = useBrands();
  const navigate = useNavigate();
  const desktopDropdownRef = useRef<HTMLDivElement>(null);
  const mobileDropdownRef = useRef<HTMLDivElement>(null);

  const brandList = brands ?? [];

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      const isDesktopClickOutside = desktopDropdownRef.current && !desktopDropdownRef.current.contains(target);
      const isMobileClickOutside = mobileDropdownRef.current && !mobileDropdownRef.current.contains(target);
      if (isDesktopClickOutside && isMobileClickOutside) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const dateFrom = useMemo(() => startOfDay(subDays(new Date(), period)).toISOString(), [period]);

  const visibleBrands = useMemo(() => {
    const brandList = brands ?? [];
    if (selectedIds.length === 0) return brandList;
    return brandList.filter((b) => selectedIds.includes(b.id));
  }, [brands, selectedIds]);

  const brandQueries = useQueries({
    queries: visibleBrands.map((b) => ({
      queryKey: ["brandDashboard", b.id, { date_from: dateFrom }],
      queryFn: () => getBrandDashboard(b.id, { date_from: dateFrom }),
      staleTime: 60000,
    }))
  });

  const isLoading = brandQueries.some(q => q.isLoading) || !brands;

  const chartData = useMemo(() => {
    if (visibleBrands.length === 0) return [];

    const daysArr = [];
    const now = new Date();
    for (let i = period - 1; i >= 0; i--) {
      daysArr.push(startOfDay(subDays(now, i)));
    }

    return daysArr.map((d) => {
      const dateStr = format(d, "d MMM", { locale: ru });
      const yyyyMmDd = format(d, "yyyy-MM-dd");
      
      const point: Record<string, string | number | Date> = { date: dateStr, rawDate: d };

      visibleBrands.forEach((b, i) => {
        const dashboardData = brandQueries[i].data;
        if (dashboardData?.by_date) {
          const dayEntry = dashboardData.by_date.find(entry => entry.date.startsWith(yyyyMmDd));
          point[b.name] = dayEntry ? (dayEntry.total ?? dayEntry.count) : 0;
        } else {
          point[b.name] = 0;
        }
      });

      return point;
    });
  }, [period, visibleBrands, brandQueries]);

  const handleChartClick = (data: { activePayload?: { payload: { rawDate: Date } }[] }) => {
    if (data?.activePayload?.[0]?.payload?.rawDate) {
      const d = data.activePayload[0].payload.rawDate as Date;
      const from = new Date(d); from.setHours(0, 0, 0, 0);
      const to = new Date(d); to.setHours(23, 59, 59, 999);
      navigate(`/mentions?date_from=${from.toISOString()}&date_to=${to.toISOString()}`);
    }
  };

  const toggleBrand = (id: string) => {
    onChangeSelectedIds(selectedIds.includes(id) ? selectedIds.filter((x) => x !== id) : [...selectedIds, id]);
  };

  const subtitle = selectedIds.length === 0
    ? "Все бренды"
    : selectedIds.length === 1
      ? brandList.find((b) => b.id === selectedIds[0])?.name ?? ""
      : `Сравнение: ${visibleBrands.map((b) => b.name).join(", ")}`;

  return (
    <>
      <div className="flex flex-col gap-2 mb-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium text-foreground">Динамика упоминаний</h3>
            {showSelection && <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>}
          </div>
          {/* Desktop selects inline */}
          <div className="flex items-center gap-2">
            {showSelection && (
              <div className="hidden sm:flex items-center gap-2">
                <div className="relative" ref={desktopDropdownRef}>
                  <button
                    onClick={() => setDropdownOpen(!dropdownOpen)}
                    className="flex items-center gap-1.5 bg-secondary border border-border text-sm rounded-lg px-3 py-1.5 hover:bg-secondary/80 transition-colors cursor-pointer min-w-[160px] max-w-[280px]"
                  >
                    {selectedIds.length === 0 ? (
                      <span className="text-muted-foreground">Все бренды</span>
                    ) : (
                      <span className="flex items-center gap-1 flex-wrap overflow-hidden">
                        {visibleBrands.slice(0, 2).map((b) => (
                          <span key={b.id} className="inline-flex items-center gap-1 bg-primary/15 text-primary text-xs rounded-full px-2 py-0.5 shrink-0">
                            {b.name}
                            <X className="h-3 w-3 cursor-pointer hover:text-destructive" onClick={(e) => { e.stopPropagation(); toggleBrand(b.id); }} />
                          </span>
                        ))}
                        {visibleBrands.length > 2 && <span className="text-xs text-muted-foreground">+{visibleBrands.length - 2}</span>}
                      </span>
                    )}
                    {selectedIds.length > 0 ? (
                      <X className="h-3.5 w-3.5 text-muted-foreground hover:text-foreground ml-auto shrink-0 cursor-pointer" onClick={(e) => { e.stopPropagation(); onChangeSelectedIds([]); }} />
                    ) : (
                      <ChevronDown className="h-3.5 w-3.5 text-muted-foreground ml-auto shrink-0" />
                    )}
                  </button>
                  {dropdownOpen && (
                    <div className="absolute right-0 top-full mt-1 z-50 bg-card border border-border rounded-lg shadow-lg py-1 min-w-[200px] max-h-[240px] overflow-y-auto">
                      {brandList.map((b) => {
                        const selected = selectedIds.includes(b.id);
                        return (
                          <button key={b.id} onClick={() => toggleBrand(b.id)} className="w-full flex items-center gap-2 px-3 py-2 text-sm text-left hover:bg-secondary/60 transition-colors cursor-pointer">
                            <span className={`w-4 h-4 rounded border flex items-center justify-center shrink-0 ${selected ? "bg-primary border-primary" : "border-border"}`}>
                              {selected && <Check className="h-3 w-3 text-primary-foreground" />}
                            </span>
                            <span className="text-foreground truncate">{b.name}</span>
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            )}
            <select
              value={period}
              onChange={(e) => onPeriodChange(Number(e.target.value))}
              className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none cursor-pointer"
            >
              {periods.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </div>
        </div>

        {/* Mobile selects on second row */}
        {showSelection && (
          <div className="flex sm:hidden items-center gap-2">
            <div className="relative flex-1" ref={mobileDropdownRef}>
              <button
                onClick={() => setDropdownOpen(!dropdownOpen)}
                className="w-full flex items-center gap-1.5 bg-secondary border border-border text-sm rounded-lg px-3 py-2 hover:bg-secondary/80 transition-colors cursor-pointer"
              >
                {selectedIds.length === 0 ? (
                  <span className="text-muted-foreground">Все бренды</span>
                ) : (
                  <span className="flex items-center gap-1 overflow-hidden">
                    {visibleBrands.slice(0, 1).map((b) => (
                      <span key={b.id} className="inline-flex items-center gap-1 bg-primary/15 text-primary text-xs rounded-full px-2 py-0.5 shrink-0">{b.name}</span>
                    ))}
                    {visibleBrands.length > 1 && <span className="text-xs text-muted-foreground">+{visibleBrands.length - 1}</span>}
                  </span>
                )}
                {selectedIds.length > 0 ? (
                  <X className="h-3.5 w-3.5 text-muted-foreground ml-auto shrink-0 cursor-pointer" onClick={(e) => { e.stopPropagation(); onChangeSelectedIds([]); }} />
                ) : (
                  <ChevronDown className="h-3.5 w-3.5 text-muted-foreground ml-auto shrink-0" />
                )}
              </button>
              {dropdownOpen && (
                <div className="absolute left-0 top-full mt-1 z-50 bg-card border border-border rounded-lg shadow-lg py-1 min-w-[200px] max-h-[240px] overflow-y-auto">
                  {brandList.map((b) => {
                    const selected = selectedIds.includes(b.id);
                    return (
                      <button key={b.id} onClick={() => toggleBrand(b.id)} className="w-full flex items-center gap-2 px-3 py-2 text-sm text-left hover:bg-secondary/60 transition-colors cursor-pointer">
                        <span className={`w-4 h-4 rounded border flex items-center justify-center shrink-0 ${selected ? "bg-primary border-primary" : "border-border"}`}>
                          {selected && <Check className="h-3 w-3 text-primary-foreground" />}
                        </span>
                        <span className="text-foreground truncate">{b.name}</span>
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
            <select
              value={period}
              onChange={(e) => onPeriodChange(Number(e.target.value))}
              className="bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none cursor-pointer"
            >
              {periods.map((p) => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </div>
        )}
      </div>

      <div className="h-[200px] sm:h-[280px] relative">
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
            <Skeleton className="w-full h-full" />
          </div>
        )}
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData} onClick={handleChartClick} style={{ cursor: "pointer" }}>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(240 3.7% 15.9%)" />
            <XAxis dataKey="date" tick={{ fill: "hsl(240 5% 64.9%)", fontSize: 11 }} axisLine={{ stroke: "hsl(240 3.7% 15.9%)" }} tickLine={false} interval={period > 14 ? 4 : 2} />
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

      {/* Custom legend */}
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
