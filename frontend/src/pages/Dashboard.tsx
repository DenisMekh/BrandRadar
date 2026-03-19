import { useState, useMemo } from "react";
import { MessageSquare, TrendingDown, AlertTriangle, Copy, TrendingUp, ArrowRight, ArrowUpRight, Plus, Tag, Play, Globe } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { LineChart, Line, ResponsiveContainer } from "recharts";
import { useQueries } from "@tanstack/react-query";
import { MentionsChart } from "@/components/dashboard/MentionsChart";
import { LiveMentionsChart } from "@/components/dashboard/LiveMentionsChart";
import { SentimentDonut } from "@/components/dashboard/SentimentDonut";
import { Skeleton } from "@/components/shared/Skeleton";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { DegradationBanner } from "@/components/shared/DegradationBanner";
import { OnboardingFlow } from "@/components/onboarding/OnboardingFlow";
import { BrandDialog, type Brand } from "@/components/settings/BrandDialog";
import { EmptyState } from "@/components/shared/EmptyState";
import { useConnection } from "@/contexts/ConnectionContext";
import { useHealth } from "@/hooks/use-health";
import { useMentions } from "@/hooks/use-mentions";
import { useAlertHistory } from "@/hooks/use-alerts";
import { useBrands, useCreateBrand } from "@/hooks/use-brands";
import { getBrandDashboard, type BrandDashboardData } from "@/lib/api/brands";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum, fmtRelative } from "@/lib/format";
import { toast } from "sonner";
import { DEFAULT_PROJECT_ID } from "@/lib/constants";
import { cn } from "@/lib/utils";
import { subDays, startOfDay, isWithinInterval, parseISO } from "date-fns";
import type { Mention } from "@/lib/api/mentions";

function generateSparkline(brandId: string, mentions: Mention[]) {
  const days: number[] = [];
  const now = new Date();
  const brandMentions = mentions.filter(m => m.brand_id === brandId);

  for (let i = 6; i >= 0; i--) {
    const d = startOfDay(subDays(now, i));
    const dayEnd = d.getTime() + 86400000;
    days.push(brandMentions.filter((m) => {
      const t = parseISO(m.created_at).getTime();
      return t >= d.getTime() && t < dayEnd;
    }).length);
  }
  return days.map((v, i) => ({ v, i }));
}

function getBrandStats(brandId: string, mentions: Mention[]) {
  const brandMentions = mentions.filter(m => m.brand_id === brandId);
  const total = brandMentions.length;
  const negative = brandMentions.filter((m) => (m.sentiment ?? m.ml?.label) === "negative").length;
  return { total, negative };
}

const Dashboard = () => {
  usePageTitle("Дашборд");
  const navigate = useNavigate();

  const { state: connectionState } = useConnection();
  const health = useHealth();
  const [period, setPeriod] = useState(7);
  const dateFrom = useMemo(() => startOfDay(subDays(new Date(), period)).toISOString(), [period]);

  const [selectedBrandIds, setSelectedBrandIds] = useState<string[]>([]);

  const alertsQuery = useAlertHistory();
  const brandsQuery = useBrands();
  const createBrand = useCreateBrand();
  const brandList = brandsQuery.data ?? [];

  const [onboardingDone, setOnboardingDone] = useState(
    () => localStorage.getItem("onboarding_completed") === "true"
  );
  const [brandDialogOpen, setBrandDialogOpen] = useState(false);

  const brandsLoaded = !brandsQuery.isLoading;
  const hasBrands = (brandsQuery.data?.length ?? 0) > 0;
  const showOnboarding = brandsLoaded && !hasBrands && !onboardingDone;

  const visibleBrands = useMemo(() => {
    const brandList = brandsQuery.data ?? [];
    if (selectedBrandIds.length === 0) return brandList;
    return brandList.filter((b) => selectedBrandIds.includes(b.id));
  }, [brandsQuery.data, selectedBrandIds]);

  const brandQueries = useQueries({
    queries: visibleBrands.map((b) => ({
      queryKey: ["brandDashboard", b.id, { date_from: dateFrom }],
      queryFn: () => getBrandDashboard(b.id, { date_from: dateFrom }),
      staleTime: 60000,
    }))
  });

  const isBrandQueriesLoading = brandQueries.some((q) => q.isLoading) || !brandsQuery.data;
  const isBrandQueriesError = brandQueries.some((q) => q.isError);

  const isDegraded = connectionState === "degraded";
  const isLoading = health.isLoading || brandsQuery.isLoading || isBrandQueriesLoading;
  const hasError = health.isError || brandsQuery.isError || isBrandQueriesError;

  // Calculate totals and Top Brands
  let totalMentions = 0;
  let positiveMentions = 0;
  let negativeMentions = 0;
  let neutralMentions = 0;

  const brandsWithStats = visibleBrands.map((b, i) => {
    const q = brandQueries[i];
    const data = q.data as BrandDashboardData | undefined;
    let mentionCount = 0;
    let negativeCount = 0;
    let sparkline: { v: number, i: number }[] = [];

    if (data) {
      mentionCount = data.total_mentions || 0;
      negativeCount = data.sentiment?.negative || 0;
      totalMentions += mentionCount;
      positiveMentions += data.sentiment?.positive || 0;
      negativeMentions += negativeCount;
      neutralMentions += data.sentiment?.neutral || 0;

      // Extract sparkline from by_date
      if (data.by_date && data.by_date.length > 0) {
        sparkline = data.by_date.map((d, index: number) => ({ v: d.total ?? d.count ?? 0, i: index }));
      } else {
        // Fallback flat sparkline
        sparkline = Array.from({ length: 7 }).map((_, idx) => ({ v: 0, i: idx }));
      }
    }

    return { ...b, mentionCount, negativeCount, sparkline };
  });

  const topBrands = brandsWithStats
    .sort((a, b) => b.mentionCount - a.mentionCount)
    .slice(0, 5);

  // Filter alerts based on selected branch
  const filteredAlerts = useMemo(() => {
    const alerts = alertsQuery.data ?? [];
    if (selectedBrandIds.length === 0) return alerts;
    return alerts.filter(a => a.brand_id && selectedBrandIds.includes(a.brand_id));
  }, [alertsQuery.data, selectedBrandIds]);

  const recentAlerts = filteredAlerts.slice(0, 3);
  const activeAlerts = filteredAlerts.length;

  const stats = [
    { title: "Всего упоминаний за период", value: fmtNum(totalMentions), trend: null, trendUp: true, icon: MessageSquare, color: "hsl(var(--primary))", bgColor: "hsl(var(--primary) / 0.2)", href: "/mentions" },
    { title: "Негативных", value: fmtNum(negativeMentions), trend: null, trendUp: false, icon: TrendingDown, color: "hsl(var(--destructive))", bgColor: "hsl(var(--destructive) / 0.2)", href: "/mentions?sentiment=negative" },
    { title: "Позитивных", value: fmtNum(positiveMentions), trend: null, trendUp: true, icon: TrendingUp, color: "hsl(var(--success))", bgColor: "hsl(var(--success) / 0.2)", href: "/mentions?sentiment=positive" },
    { title: "Алертов", value: fmtNum(activeAlerts), trend: null, trendUp: false, icon: AlertTriangle, color: "hsl(var(--warning))", bgColor: "hsl(var(--warning) / 0.2)", href: "/alerts" },
  ];

  const handleBrandSave = (brand: Brand) => {
    createBrand.mutate({
      data: {
        name: brand.name,
        keywords: brand.keywords,
        exclusions: brand.exclusions,
        risk_words: brand.riskWords
      }
    }, {
      onSuccess: () => toast.success("Бренд создан"),
      onError: (e) => toast.error(`Не удалось сохранить: ${e.message}`),
    });
    setBrandDialogOpen(false);
  };

  return (
    <div className="space-y-4 sm:space-y-6">
      {isDegraded && <DegradationBanner />}
      {hasError && <ErrorBanner onRetry={() => { health.refetch(); }} />}

      {/* Stat cards: 2 cols on mobile, 4 on xl */}
      <div className="grid grid-cols-2 xl:grid-cols-4 gap-3 sm:gap-4">
        {isLoading
          ? Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-card border border-border rounded-xl p-4 space-y-3">
              <div className="flex items-center justify-between"><Skeleton className="w-8 h-8 sm:w-10 sm:h-10 rounded-full" /><Skeleton className="w-12 sm:w-14 h-5 rounded-full" /></div>
              <div><Skeleton className="w-16 sm:w-24 h-7 sm:h-8 mb-1.5" /><Skeleton className="w-20 sm:w-32 h-3.5 sm:h-4" /></div>
            </div>
          ))
          : stats.map((stat) => (
            <Link
              key={stat.title}
              to={stat.href}
              className="bg-card border border-border rounded-xl p-4 flex flex-col gap-2.5 sm:gap-4 hover:scale-[1.02] hover:border-muted-foreground/30 transition-all duration-200 cursor-pointer relative group"
            >
              <ArrowUpRight className="absolute top-3 right-3 sm:top-4 sm:right-4 h-3.5 w-3.5 sm:h-4 sm:w-4 text-muted-foreground/50 opacity-0 group-hover:opacity-100 transition-opacity" />
              <div className="flex items-center justify-between">
                <div className="w-8 h-8 sm:w-10 sm:h-10 rounded-full flex items-center justify-center" style={{ backgroundColor: stat.bgColor }}>
                  <stat.icon className="h-4 w-4 sm:h-5 sm:w-5" style={{ color: stat.color }} />
                </div>
                <span className={`text-[10px] sm:text-xs font-medium px-1.5 sm:px-2 py-0.5 rounded-full flex items-center gap-0.5 sm:gap-1 ${stat.trendUp ? "bg-success/20 text-success" : "bg-destructive/20 text-destructive"}`}>
                  {stat.trendUp ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}{stat.trend}
                </span>
              </div>
              <div>
                <p className="text-xl sm:text-2xl md:text-3xl font-bold text-foreground">{stat.value}</p>
                <p className="text-xs sm:text-sm text-muted-foreground mt-0.5 sm:mt-1 truncate">{stat.title}</p>
              </div>
            </Link>
          ))}
      </div>

      {/* Quick actions */}
      <div className="flex gap-2 overflow-x-auto pb-1 -mx-4 px-4 sm:mx-0 sm:px-0 sm:flex-wrap">
        <button onClick={() => setBrandDialogOpen(true)} className="flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors cursor-pointer border border-border whitespace-nowrap min-h-[44px]">
          <Plus className="h-4 w-4" /><Tag className="h-4 w-4" /> Добавить бренд
        </button>
        <button onClick={() => navigate("/sources")} className="flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors cursor-pointer border border-border whitespace-nowrap min-h-[44px]">
          <Play className="h-4 w-4" /><Globe className="h-4 w-4" /> Запустить сбор
        </button>
        <button onClick={() => navigate("/mentions?sentiment=negative")} className="flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors cursor-pointer border border-border whitespace-nowrap min-h-[44px]">
          <AlertTriangle className="h-4 w-4" /> Посмотреть негатив
        </button>
        <select
          value={period}
          onChange={(e) => setPeriod(Number(e.target.value))}
          className="ml-auto bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none cursor-pointer whitespace-nowrap"
        >
          <option value={1}>1 день</option>
          <option value={7}>7 дней</option>
          <option value={14}>14 дней</option>
          <option value={30}>30 дней</option>
        </select>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-3 sm:gap-4">
        <div className="lg:col-span-2 bg-card border border-border rounded-xl p-4 sm:p-6">
          {isLoading ? <Skeleton className="w-full h-[240px] sm:h-[340px]" /> : <MentionsChart selectedIds={selectedBrandIds} onChangeSelectedIds={setSelectedBrandIds} period={period} onPeriodChange={setPeriod} />}
        </div>
        <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
          {isLoading ? <Skeleton className="w-full h-[240px] sm:h-[340px]" /> : <SentimentDonut total={totalMentions} positive={positiveMentions} negative={negativeMentions} neutral={neutralMentions} isLoading={isLoading} />}
        </div>
      </div>

      {/* Live Dynamics Chart
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        {isLoading ? <Skeleton className="w-full h-[300px]" /> : <LiveMentionsChart period={period} onPeriodChange={setPeriod} />}
      </div>
      */}


      {/* Top Brands */}
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-base sm:text-lg font-medium text-foreground">Топ бренды</h3>
          <Link to="/brands" className="text-sm text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer">
            Все бренды <ArrowRight className="h-3.5 w-3.5" />
          </Link>
        </div>

        {brandsQuery.isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4 p-3 rounded-lg">
                <Skeleton className="w-8 h-8 rounded-full" />
                <div className="flex-1 space-y-2"><Skeleton className="w-32 h-4" /><Skeleton className="w-48 h-3" /></div>
                <Skeleton className="w-[80px] h-8" />
              </div>
            ))}
          </div>
        ) : topBrands.length === 0 ? (
          <EmptyState
            icon={Tag}
            title="Добавьте бренд"
            description="Создайте первый бренд, чтобы увидеть статистику."
            actionLabel="Создать бренд"
            onAction={() => setBrandDialogOpen(true)}
          />
        ) : (
          <div className="space-y-1">
            {topBrands.map((brand, index) => (
              <Link
                key={brand.id}
                to={`/brands/${brand.id}`}
                className="flex items-center gap-3 sm:gap-4 p-3 rounded-lg hover:bg-secondary/50 transition-all duration-200 cursor-pointer group"
              >
                {/* Position */}
                <span className={cn(
                  "text-lg sm:text-xl font-bold w-8 text-center shrink-0",
                  index === 0 ? "text-primary" : "text-muted-foreground/40"
                )}>
                  #{index + 1}
                </span>

                {/* Brand info */}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-foreground truncate group-hover:text-primary transition-colors">{brand.name}</p>
                  <div className="flex items-center gap-3 mt-0.5">
                    <span className="text-xs text-muted-foreground">{fmtNum(brand.mentionCount)} упоминаний</span>
                    {brand.negativeCount > 0 && (
                      <span className="text-xs text-destructive">{fmtNum(brand.negativeCount)} негативных</span>
                    )}
                  </div>
                </div>

                {/* Sparkline */}
                <div className="w-[80px] h-[32px] shrink-0 hidden sm:block">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={brand.sparkline}>
                      <Line
                        type="monotone"
                        dataKey="v"
                        stroke="hsl(var(--primary))"
                        strokeWidth={1.5}
                        dot={false}
                        isAnimationActive={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                <ArrowUpRight className="h-4 w-4 text-muted-foreground/30 opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
              </Link>
            ))}
          </div>
        )}
      </div>

      {/* Recent alerts */}
      <div className="bg-card border border-border rounded-xl p-4 sm:p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-base sm:text-lg font-medium text-foreground">Последние алерты</h3>
          <Link to="/alerts" className="text-sm text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer">Показать все <ArrowRight className="h-3.5 w-3.5" /></Link>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3 sm:gap-4">
          {isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="bg-secondary/30 rounded-lg p-4 space-y-2"><Skeleton className="w-24 h-4" /><Skeleton className="w-full h-4" /><Skeleton className="w-16 h-3" /></div>
            ))
            : recentAlerts.map((alert, i) => {
              const alertTime = new Date(alert.fired_at || alert.created_at || "");
              const windowStart = new Date(alert.window_start || "");
              const windowMins = Math.round((alertTime.getTime() - windowStart.getTime()) / 60000);
              const brand = brandsQuery.data?.find((b) => b.id === alert.brand_id);
              const href = `/mentions?brand_id=${alert.brand_id}&date_from=${windowStart.toISOString()}&date_to=${alertTime.toISOString()}`;
              return (
                <Link
                  key={alert.id || i}
                  to={href}
                  className="border-l-4 border-l-destructive bg-secondary/30 rounded-lg p-4 hover:scale-[1.02] hover:border-muted-foreground/30 transition-all duration-200 cursor-pointer relative group block"
                >
                  <ArrowUpRight className="absolute top-3 right-3 h-3.5 w-3.5 text-muted-foreground/50 opacity-0 group-hover:opacity-100 transition-opacity" />
                  <p className="font-medium text-foreground text-sm">{brand?.name || "Неизвестный бренд"}</p>
                  <p className="text-sm text-muted-foreground mt-1">
                    {alert.mentions_count === 0
                      ? "Алерт сработал"
                      : `${fmtNum(alert.mentions_count || 0)} упоминаний`} за {windowMins} мин
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-2">{fmtRelative(alert.fired_at || alert.created_at || "")}</p>
                </Link>
              );
            })}
        </div>
      </div>

      <BrandDialog open={brandDialogOpen} onOpenChange={setBrandDialogOpen} brand={null} onSave={handleBrandSave} />
    </div>
  );
};

export default Dashboard;
