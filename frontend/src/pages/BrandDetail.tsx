import { useParams, Link } from "react-router-dom";
import { ArrowLeft, TrendingUp, TrendingDown, MessageSquare, AlertTriangle, Copy, ArrowRight, ArrowUpRight, Tag, Activity } from "lucide-react";
import { AreaChart, Area, BarChart, Bar, Cell, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { cn } from "@/lib/utils";
import { Skeleton } from "@/components/shared/Skeleton";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { useBrands, useBrandDashboard } from "@/hooks/use-brands";
import { useAlertHistory } from "@/hooks/use-alerts";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum, fmtRelative } from "@/lib/format";
import { useState, useMemo } from "react";
import { subDays, startOfDay } from "date-fns";
import { SentimentDonut } from "@/components/dashboard/SentimentDonut";
import { MentionsChart } from "@/components/dashboard/MentionsChart";

const SENTIMENT_COLORS = {
  positive: "hsl(var(--success))",
  neutral: "hsl(var(--muted-foreground))",
  negative: "hsl(var(--destructive))",
};

const SOURCE_COLORS: Record<string, string> = {
  Telegram: "#3b82f6",
  Web: "hsl(263, 70%, 58%)",
  RSS: "#f59e0b",
};

const BrandDetail = () => {
  const { id } = useParams<{ id: string }>();

  // Fetch lists
  const { data: brands, isLoading: brandsLoading, isError: brandsError, refetch: refetchBrands } = useBrands();
  const [period, setPeriod] = useState(7);
  const dateFrom = useMemo(() => startOfDay(subDays(new Date(), period)).toISOString(), [period]);
  const { data: dashboard, isLoading: dashLoading, isError: dashError, refetch: refetchDash } = useBrandDashboard(id, { date_from: dateFrom });
  const { data: alerts, isLoading: alertsLoading, isError: alertsError, refetch: refetchAlerts } = useAlertHistory(id);

  const brand = brands?.find((b) => b.id === id);
  usePageTitle(brand?.name ?? "Бренд");

  const isLoading = brandsLoading || dashLoading || alertsLoading;
  const isError = brandsError || dashError || alertsError;

  // Map daily mentions for the area chart
  const dailyData = useMemo(() => {
    return (dashboard?.by_date ?? []).map((d) => {
      const parts = d.date.split('-');
      const dateStr = parts.length === 3 ? `${parts[2]}.${parts[1]}` : d.date;
      return {
        date: dateStr,
        positive: d.sentiment?.positive ?? 0,
        neutral: d.sentiment?.neutral ?? 0,
        negative: d.sentiment?.negative ?? 0,
        total: d.total ?? d.count ?? 0
      };
    });
  }, [dashboard?.by_date]);

  const sourceData = useMemo(() => (dashboard?.by_source ?? [])
    .sort((a, b) => b.count - a.count)
    .slice(0, 5)
    .map((s) => ({
      name: s.source,
      count: s.count,
      fill: SOURCE_COLORS[s.source] ?? "hsl(var(--primary))"
    })), [dashboard?.by_source]);

  const handleRetry = () => {
    refetchBrands();
    refetchDash();
    refetchAlerts();
  };

  if (isLoading) return (
    <div className="space-y-6">
      <Skeleton className="w-48 h-8" />
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => <div key={i} className="bg-card border border-border rounded-xl p-6"><Skeleton className="w-full h-20" /></div>)}
      </div>
      <Skeleton className="w-full h-[300px]" />
    </div>
  );

  if (isError) return <ErrorBanner onRetry={handleRetry} />;

  if (!brand || !dashboard) return (
    <div className="flex flex-col items-center justify-center py-20 text-center animate-page-enter">
      <Tag className="h-12 w-12 text-muted-foreground/30 mb-4" />
      <h2 className="text-xl font-semibold text-foreground">Бренд не найден</h2>
      <p className="text-sm text-muted-foreground mt-1">Возможно, он был удалён или ID некорректен.</p>
      <Link to="/brands" className="mt-4 px-4 py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">К списку брендов</Link>
    </div>
  );

  // Calculate stats from filtered dashboard
  const totalMentions = dashboard.total_mentions ?? 0;
  const negativeCount = dashboard.sentiment?.negative ?? 0;
  const positiveCount = dashboard.sentiment?.positive ?? 0;
  const neutralCount = dashboard.sentiment?.neutral ?? 0;
  const recentAlertsCount = dashboard.recent_alerts ?? 0;

  const stats = [
    { title: "Всего упоминаний за период", value: fmtNum(totalMentions), trend: null, icon: MessageSquare, color: "hsl(var(--primary))", bg: "hsl(var(--primary) / 0.2)", href: `/mentions?brand_id=${id}&date_from=${dateFrom}` },
    { title: "Негативных", value: fmtNum(negativeCount), trend: null, icon: TrendingDown, color: "hsl(var(--destructive))", bg: "hsl(var(--destructive) / 0.2)", href: `/mentions?brand_id=${id}&sentiment=negative&date_from=${dateFrom}` },
    { title: "Позитивных", value: fmtNum(positiveCount), trend: null, icon: TrendingUp, color: "hsl(var(--success))", bg: "hsl(var(--success) / 0.2)", href: `/mentions?brand_id=${id}&sentiment=positive&date_from=${dateFrom}` },
    { title: "Алертов", value: fmtNum(recentAlertsCount), trend: null, icon: AlertTriangle, color: "hsl(var(--warning))", bg: "hsl(var(--warning) / 0.2)", href: `/alerts` },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <Link to="/brands" className="p-2.5 sm:p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary transition-colors cursor-pointer min-w-[44px] min-h-[44px] flex items-center justify-center">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div className="min-w-0">
            <h1 className="text-xl sm:text-2xl font-bold text-foreground truncate">{brand.name}</h1>
            <div className="flex gap-1.5 mt-1 flex-wrap">
              {(brand.keywords ?? []).slice(0, 3).map((kw) => (
                <span key={kw} className="bg-violet-500/15 text-violet-400 text-xs rounded-full px-2 py-0.5">{kw}</span>
              ))}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={period}
            onChange={(e) => setPeriod(Number(e.target.value))}
            className="bg-card border border-border text-foreground text-sm rounded-lg px-3 py-1.5 focus:outline-none cursor-pointer"
          >
            <option value={1}>1 день</option>
            <option value={7}>7 дней</option>
            <option value={14}>14 дней</option>
            <option value={30}>30 дней</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <Link
            key={stat.title}
            to={stat.href}
            className="bg-card border border-border rounded-xl p-5 flex flex-col gap-3 hover:scale-[1.02] hover:border-muted-foreground/30 transition-all duration-200 cursor-pointer relative group"
          >
            <ArrowUpRight className="absolute top-4 right-4 h-4 w-4 text-muted-foreground/50 opacity-0 group-hover:opacity-100 transition-opacity" />
            <div className="flex items-center justify-between">
              <div className="w-9 h-9 rounded-full flex items-center justify-center" style={{ backgroundColor: stat.bg }}>
                <stat.icon className="h-4 w-4" style={{ color: stat.color }} />
              </div>
            </div>
            <div>
              <p className="text-2xl font-bold text-foreground">{stat.value}</p>
              <p className="text-xs text-muted-foreground mt-0.5">{stat.title}</p>
            </div>
          </Link>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-4 sm:p-6">
          <MentionsChart
            selectedIds={id ? [id] : []}
            onChangeSelectedIds={() => { }}
            period={period}
            onPeriodChange={setPeriod}
            showSelection={false}
          />
        </div>

        <div className="bg-card border border-border rounded-xl p-6">
          <SentimentDonut total={totalMentions} positive={positiveCount} negative={negativeCount} neutral={neutralCount} brand_id={id} />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-card border border-border rounded-xl p-6">
          <h3 className="text-sm font-medium text-foreground mb-4">Топ источников</h3>
          {sourceData.length > 0 ? (
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={sourceData} layout="vertical">
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis type="number" tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11 }} />
                <YAxis type="category" dataKey="name" tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11 }} width={80} />
                <Tooltip contentStyle={{ backgroundColor: "hsl(var(--card))", border: "1px solid hsl(var(--border))", borderRadius: 8, fontSize: 12 }} />
                <Bar dataKey="count" name="Упоминаний" isAnimationActive>
                  {sourceData.map((d, i) => <Cell key={i} fill={d.fill} />)}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[200px] flex items-center justify-center text-muted-foreground text-sm border-2 border-dashed border-border rounded-xl">
              Нет данных
            </div>
          )}
        </div>

        <div className="bg-card border border-border rounded-xl p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-foreground">Последние алерты бренда</h3>
            <Link to="/alerts" className="text-xs text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer">
              Все <ArrowRight className="h-3 w-3" />
            </Link>
          </div>
          {(alerts ?? []).length > 0 ? (
            <div className="space-y-2">
              {(alerts ?? []).slice(0, 5).map((a) => {
                const windowMins = a.window_start && a.window_end
                  ? Math.round((new Date(a.window_end).getTime() - new Date(a.window_start).getTime()) / 60000)
                  : 0;
                return (
                  <Link
                    key={a.id}
                    to="/alerts"
                    className="flex items-center justify-between p-3 rounded-lg hover:bg-secondary/50 transition-colors cursor-pointer border-l-3 border-l-warning"
                  >
                    <div>
                      <span className="text-sm text-foreground font-medium">
                        {a.mentions_count === 0 ? "Алерт сработал" : `${fmtNum(a.mentions_count ?? 0)} упоминаний`}
                      </span>
                      <span className="text-xs text-muted-foreground ml-2">за {windowMins} мин</span>
                    </div>
                    <span className="text-xs text-muted-foreground/60">{fmtRelative(a.fired_at ?? new Date().toISOString())}</span>
                  </Link>
                );
              })}
            </div>
          ) : (
            <div className="h-[200px] flex items-center justify-center text-muted-foreground text-sm border-2 border-dashed border-border rounded-xl mt-4">
              Алертов пока нет
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default BrandDetail;
