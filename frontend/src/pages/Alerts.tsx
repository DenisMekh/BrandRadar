import { useState, useEffect } from "react";
import { Gauge, Timer, Pencil, ArrowRight, CheckCircle, Save, X, Flame, Plus } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Slider } from "@/components/ui/slider";
import { cn } from "@/lib/utils";
import { Link } from "react-router-dom";
import { Skeleton } from "@/components/shared/Skeleton";
import { EmptyState } from "@/components/shared/EmptyState";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useAlertConfigs, useUpdateAlertConfig, useAlertHistory, useDeleteAlertConfig, useCreateAlertConfig } from "@/hooks/use-alerts";
import { useBrands } from "@/hooks/use-brands";
import { useQueryClient } from "@tanstack/react-query";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum, fmtRelative } from "@/lib/format";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { Trash2 } from "lucide-react";
import type { AlertConfig } from "@/lib/api/alerts";

const sentimentFilterOptions = [
  { value: "all", label: "Любая" },
  { value: "negative", label: "Негативная" },
  { value: "positive", label: "Позитивная" },
  { value: "neutral", label: "Нейтральная" },
];

const aggressionLabels: Record<number, string> = {
  1: "Спокойный",
  2: "Тихий",
  3: "Нормальный",
  4: "Активный",
  5: "Агрессивный",
};


const Alerts = () => {
  usePageTitle("Алерты");

  const configsQuery = useAlertConfigs();
  const historyQuery = useAlertHistory();
  const updateConfig = useUpdateAlertConfig();
  const deleteConfig = useDeleteAlertConfig();
  const createConfig = useCreateAlertConfig();
  const { data: brands = [] } = useBrands();
  const qc = useQueryClient();

  useEffect(() => {
    // Mark as read when entering the page
    localStorage.setItem("lastViewedAlerts", new Date().toISOString());
    qc.invalidateQueries({ queryKey: ["alertHistory"] });
  }, [qc]);

  const configs = (configsQuery.data as AlertConfig[]) ?? [];
  const history = historyQuery.data ?? [];

  const isLoadingConfigs = configsQuery.isLoading;
  const isLoadingHistory = historyQuery.isLoading && historyQuery.fetchStatus !== 'idle';

  const hasConfigError = configsQuery.isError;
  const hasHistoryError = historyQuery.isError && historyQuery.fetchStatus !== 'idle';

  const [editingId, setEditingId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<AlertConfig | null>(null);
  const [editForm, setEditForm] = useState<{
    threshold: number;
    cooldown_minutes: number;
    sentiment_filter: string;
    aggression: number;
    anomaly_window_size: number;
    percentile: number;
    window_minutes: number;
  }>({
    threshold: 10,
    cooldown_minutes: 60,
    sentiment_filter: "all",
    aggression: 0.5,
    anomaly_window_size: 24,
    percentile: 95,
    window_minutes: 60,
  });
  const [createForm, setCreateForm] = useState<{
    brand_id: string;
    threshold: number;
    cooldown_minutes: number;
    sentiment_filter: string;
    aggression: number;
    anomaly_window_size: number;
    percentile: number;
    window_minutes: number;
  }>({
    brand_id: "",
    threshold: 10,
    cooldown_minutes: 60,
    sentiment_filter: "all",
    aggression: 0.5,
    anomaly_window_size: 24,
    percentile: 95,
    window_minutes: 60,
  });

  const toggleConfig = (id: string, enabled: boolean) => {
    updateConfig.mutate({ id, data: { enabled: !enabled } }, {
      onSuccess: () => toast.success(enabled ? "Алерт отключён" : "Алерт включён"),
      onError: () => toast.error("Не удалось обновить конфигурацию"),
    });
  };

  const handleDelete = () => {
    if (!deleteTarget) return;
    deleteConfig.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("Конфигурация удалена");
        setDeleteTarget(null);
      },
      onError: (e) => toast.error(`Ошибка при удалении: ${e.message}`),
    });
  };

  const handleCreate = () => {
    if (!createForm.brand_id) {
      toast.error("Выберите бренд");
      return;
    }
    createConfig.mutate({ ...createForm, enabled: true }, {
      onSuccess: () => {
        toast.success("Конфигурация создана");
        setIsCreating(false);
        setCreateForm({
          brand_id: "",
          threshold: 10,
          cooldown_minutes: 60,
          sentiment_filter: "all",
          aggression: 0.5,
          anomaly_window_size: 24,
          percentile: 95,
          window_minutes: 60
        });
      },
      onError: (e) => toast.error(`Ошибка при создании: ${e.message}`),
    });
  };

  const startEdit = (config: AlertConfig) => {
    setEditingId(config.id);
    setEditForm({
      threshold: config.threshold,
      cooldown_minutes: config.cooldown_minutes,
      sentiment_filter: config.sentiment_filter || "all",
      aggression: config.aggression ?? 0.5,
      anomaly_window_size: config.anomaly_window_size ?? 24,
      percentile: config.percentile ?? 95,
      window_minutes: config.window_minutes ?? 60,
    });
  };

  const saveEdit = (id: string) => {
    updateConfig.mutate({
      id, data: {
        threshold: editForm.threshold,
        cooldown_minutes: editForm.cooldown_minutes,
        sentiment_filter: editForm.sentiment_filter,
        aggression: editForm.aggression,
        anomaly_window_size: editForm.anomaly_window_size,
        percentile: editForm.percentile,
        window_minutes: editForm.window_minutes
      }
    }, {
      onSuccess: () => { toast.success("Конфигурация сохранена"); setEditingId(null); },
      onError: () => toast.error("Не удалось сохранить"),
    });
  };

  return (
    <div className="space-y-8">
      <section>
        <h3 className="text-lg font-medium text-foreground mb-4">Настройки алертов</h3>
        {hasConfigError ? (
          <ErrorBanner onRetry={() => configsQuery.refetch()} />
        ) : isLoadingConfigs ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="bg-card border border-border rounded-xl p-5 space-y-4">
                <div className="flex justify-between"><Skeleton className="w-32 h-5" /><Skeleton className="w-10 h-5 rounded-full" /></div>
                <div className="grid grid-cols-2 gap-3">{Array.from({ length: 4 }).map((_, j) => <Skeleton key={j} className="w-full h-4" />)}</div>
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {configs.map((config) => {
              const isEditing = editingId === config.id;
              return (
                <div key={config.id} className="bg-card border border-border rounded-xl p-5 relative hover:scale-[1.005] transition-transform duration-200">
                  <div className="flex items-start justify-between mb-4">
                    <h4 className="text-lg font-medium text-foreground">
                      {brands.find(b => b.id === config.brand_id)?.name || config.brand_name || "Неизвестный"}
                    </h4>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setDeleteTarget(config)}
                        className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg hover:bg-destructive/10 transition-colors cursor-pointer"
                        title="Удалить конфигурацию"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                      <Switch checked={config.enabled} onCheckedChange={() => toggleConfig(config.id, config.enabled)} />
                    </div>
                  </div>

                  {isEditing ? (
                    <div className="space-y-4">

                      {/* Cooldown */}
                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1.5">
                          <label className="text-xs text-muted-foreground font-medium">Кулдаун (мин)</label>
                          <input
                            type="number"
                            min={1}
                            value={editForm.cooldown_minutes}
                            onChange={(e) => setEditForm({ ...editForm, cooldown_minutes: Number(e.target.value) })}
                            className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                          />
                        </div>
                        <div className="space-y-1.5">
                          <label className="text-xs text-muted-foreground font-medium">Окно (мин)</label>
                          <input
                            type="number"
                            min={1}
                            value={editForm.window_minutes}
                            onChange={(e) => setEditForm({ ...editForm, window_minutes: Number(e.target.value) })}
                            className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                          />
                        </div>
                      </div>

                      {/* Anomaly parameters */}
                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1.5">
                          <label className="text-xs text-muted-foreground font-medium">Окно (упоминания)</label>
                          <input
                            type="number"
                            min={1}
                            value={editForm.anomaly_window_size}
                            onChange={(e) => setEditForm({ ...editForm, anomaly_window_size: Number(e.target.value) })}
                            className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                          />
                        </div>
                        <div className="space-y-1.5">
                          <label className="text-xs text-muted-foreground font-medium">Перцентиль</label>
                          <input
                            type="number"
                            min={1} max={100}
                            value={editForm.percentile}
                            onChange={(e) => setEditForm({ ...editForm, percentile: Number(e.target.value) })}
                            className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary"
                          />
                        </div>
                      </div>

                      {/* Sentiment */}
                      <div className="space-y-1.5">
                        <label className="text-xs text-muted-foreground">Тональность</label>
                        <select
                          value={editForm.sentiment_filter}
                          onChange={(e) => setEditForm({ ...editForm, sentiment_filter: e.target.value })}
                          className="w-full bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer"
                        >
                          {sentimentFilterOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                        </select>
                      </div>

                      {/* Aggression slider */}
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <label className="text-xs text-muted-foreground">Агрессивность (порог ML score)</label>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <span className="text-xs text-muted-foreground/60 cursor-help">ℹ️</span>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-[260px] text-xs">
                              ML-модель оценивает агрессивность от 0 до 1. Алерт сработает если score агрессивности упоминания выше этого порога.
                            </TooltipContent>
                          </Tooltip>
                        </div>
                        <div className="flex items-center justify-center">
                          <span className="text-sm font-medium text-foreground tabular-nums">{editForm.aggression.toFixed(1)}</span>
                        </div>
                        <div className="relative">
                          <div className="absolute inset-x-0 top-1/2 -translate-y-1/2 h-2 rounded-full" style={{ background: "linear-gradient(to right, hsl(var(--success)), hsl(var(--warning)), hsl(var(--destructive)))" }} />
                          <Slider
                            value={[editForm.aggression]}
                            onValueChange={([v]) => setEditForm({ ...editForm, aggression: Math.round(v * 10) / 10 })}
                            min={0} max={1} step={0.1}
                            className="relative z-10"
                          />
                        </div>
                        <div className="flex justify-between text-[10px] text-muted-foreground/60">
                          <span>0 — любые</span>
                          <span>1 — только агрессивные</span>
                        </div>
                      </div>

                      <div className="flex gap-2 pt-1">
                        <button onClick={() => saveEdit(config.id)} className="flex items-center gap-1 px-3 py-1.5 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">
                          <Save className="h-3.5 w-3.5" /> Сохранить
                        </button>
                        <button onClick={() => setEditingId(null)} className="flex items-center gap-1 px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer">
                          <X className="h-3.5 w-3.5" /> Отмена
                        </button>
                      </div>
                    </div>
                  ) : (
                    <>
                      <div className="space-y-2">
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Timer className="h-4 w-4 shrink-0" />
                          <span>Кулдаун: {config.cooldown_minutes} минут</span>
                        </div>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <span className="w-4 h-4 shrink-0 text-center text-xs">🎯</span>
                          <span>Тональность: {sentimentFilterOptions.find((o) => o.value === config.sentiment_filter)?.label ?? config.sentiment_filter}</span>
                        </div>
                        <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground/80 mt-1">
                          <div className="flex items-center gap-1.5">
                            <span className="w-1.5 h-1.5 rounded-full bg-primary/40" />
                            <span>Окно: {config.window_minutes}м</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <span className="w-1.5 h-1.5 rounded-full bg-primary/40" />
                            <span>Anomaly: {config.anomaly_window_size}ч</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            <span className="w-1.5 h-1.5 rounded-full bg-primary/40" />
                            <span>Percentile: {config.percentile}%</span>
                          </div>
                        </div>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Flame className="h-4 w-4 shrink-0" />
                          <span>Агрессивность: {(config.aggression ?? 0.5).toFixed(1)}</span>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <div className="w-[100px] h-2 rounded-full overflow-hidden bg-secondary cursor-help">
                                <div
                                  className="h-full rounded-full transition-all"
                                  style={{
                                    width: `${(config.aggression ?? 0.5) * 100}%`,
                                    background: (config.aggression ?? 0.5) <= 0.3 ? "hsl(var(--success))" : (config.aggression ?? 0.5) <= 0.6 ? "hsl(var(--warning))" : "hsl(var(--destructive))",
                                  }}
                                />
                              </div>
                            </TooltipTrigger>
                            <TooltipContent className="max-w-[260px] text-xs">
                              ML-модель оценивает агрессивность от 0 до 1. Алерт сработает если score агрессивности упоминания выше этого порога.
                            </TooltipContent>
                          </Tooltip>
                        </div>
                      </div>
                      <button onClick={() => startEdit(config)} className="mt-4 text-sm text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer">
                        <Pencil className="h-3.5 w-3.5" /> Редактировать
                      </button>
                    </>
                  )}
                </div>
              );
            })}

            {/* Create Card */}
            {isCreating ? (
              <div className="bg-card border-2 border-primary/30 border-dashed rounded-xl p-5 space-y-4 animate-in fade-in zoom-in duration-300">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-lg font-medium text-foreground">Новая конфигурация</h4>
                  <button onClick={() => setIsCreating(false)} className="text-muted-foreground hover:text-foreground transition-colors"><X className="h-4 w-4" /></button>
                </div>

                <div className="space-y-4">
                  {/* Brand Selector */}
                  <div className="space-y-1.5">
                    <label className="text-xs text-muted-foreground">Бренд</label>
                    <select
                      value={createForm.brand_id}
                      onChange={(e) => setCreateForm({ ...createForm, brand_id: e.target.value })}
                      className="w-full bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer"
                    >
                      <option value="" disabled>Выберите бренд...</option>
                      {brands.filter(b => !configs.some(c => c.brand_id === b.id)).map(b => (
                        <option key={b.id} value={b.id}>{b.name}</option>
                      ))}
                    </select>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-1.5">
                      <label className="text-[10px] uppercase font-bold text-muted-foreground/70">Кулдаун (минуты)</label>
                      <input type="number" min={1} value={createForm.cooldown_minutes} onChange={(e) => setCreateForm({ ...createForm, cooldown_minutes: Number(e.target.value) })} className="w-full bg-secondary border border-border rounded-lg px-2 py-1.5 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
                    </div>
                    <div className="space-y-1.5">
                      <label className="text-[10px] uppercase font-bold text-muted-foreground/70">Окно (минуты)</label>
                      <input type="number" min={1} value={createForm.window_minutes} onChange={(e) => setCreateForm({ ...createForm, window_minutes: Number(e.target.value) })} className="w-full bg-secondary border border-border rounded-lg px-2 py-1.5 text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="space-y-1.5">
                      <label className="text-[10px] uppercase font-bold text-muted-foreground/70">Окно (упоминания)</label>
                      <input type="number" min={1} value={createForm.anomaly_window_size} onChange={(e) => setCreateForm({ ...createForm, anomaly_window_size: Number(e.target.value) })} className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
                    </div>
                    <div className="space-y-1.5">
                      <label className="text-[10px] uppercase font-bold text-muted-foreground/70">Перцентиль</label>
                      <input type="number" min={1} max={100} value={createForm.percentile} onChange={(e) => setCreateForm({ ...createForm, percentile: Number(e.target.value) })} className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-primary" />
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs text-muted-foreground">Тональность</label>
                    <select value={createForm.sentiment_filter} onChange={(e) => setCreateForm({ ...createForm, sentiment_filter: e.target.value })} className="w-full bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer">
                      {sentimentFilterOptions.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                    </select>
                  </div>

                  <div className="flex gap-2 pt-1">
                    <button onClick={handleCreate} className="flex-1 flex items-center justify-center gap-1.5 px-3 py-2 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">
                      <Save className="h-4 w-4" /> Создать
                    </button>
                    <button onClick={() => setIsCreating(false)} className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground font-medium transition-colors cursor-pointer">Отмена</button>
                  </div>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setIsCreating(true)}
                className="bg-card border-2 border-border border-dashed rounded-xl p-5 flex flex-col items-center justify-center gap-3 text-muted-foreground hover:border-primary/50 hover:text-primary transition-all duration-300 group min-h-[220px] cursor-pointer"
              >
                <div className="w-12 h-12 rounded-full bg-secondary flex items-center justify-center group-hover:bg-primary/10 transition-colors">
                  <Plus className="h-6 w-6" />
                </div>
                <span className="text-sm font-medium">Добавить конфигурацию</span>
              </button>
            )}
          </div>
        )}
      </section>

      <section>
        <h3 className="text-lg font-medium text-foreground mb-4">История алертов</h3>
        {hasHistoryError ? (
          <ErrorBanner onRetry={() => historyQuery.refetch()} />
        ) : isLoadingHistory ? (
          <div className="bg-card border border-border rounded-xl overflow-hidden">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex gap-4 px-5 py-3 border-b border-border last:border-0">
                <Skeleton className="w-24 h-4" /><Skeleton className="w-12 h-4" /><Skeleton className="w-16 h-4" /><Skeleton className="w-20 h-4" />
              </div>
            ))}
          </div>
        ) : history.length === 0 ? (
          <EmptyState icon={CheckCircle} iconClassName="h-12 w-12 text-success/40 mb-4" title="Всё спокойно — алертов нет" description="Система работает в штатном режиме." />
        ) : (
          <>
            <div className="hidden md:block bg-card border border-border rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead><tr className="border-b border-border text-muted-foreground text-left">
                  <th className="px-5 py-3 font-medium">Бренд</th><th className="px-5 py-3 font-medium">Время</th><th className="px-5 py-3 font-medium"></th>
                </tr></thead>
                <tbody>
                  {history.map((h) => {
                    const brand = brands.find(b => b.id === h.brand_id);
                    return (
                      <tr key={h.id} className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors">
                        <td className="px-5 py-3 text-foreground font-medium">{brand?.name ?? "Неизвестный"}</td>
                        <td className="px-5 py-3 text-muted-foreground">{fmtRelative(h.fired_at || h.created_at)}</td>
                        <td className="px-5 py-3">
                          <Link to={`/mentions?brand_id=${h.brand_id}`} className="text-primary hover:text-primary/80 flex items-center gap-1 text-xs transition-colors cursor-pointer">
                            К упоминаниям <ArrowRight className="h-3 w-3" />
                          </Link>
                        </td>
                      </tr >
                    );
                  })}
                </tbody >
              </table >
            </div >
            <div className="md:hidden space-y-3">
              {history.map((h) => {
                const brand = brands.find(b => b.id === h.brand_id);
                return (
                  <div key={h.id} className="bg-card border border-border rounded-xl p-4 hover:scale-[1.01] transition-transform duration-200">
                    <div className="flex items-center justify-between"><span className="font-medium text-foreground">{brand?.name ?? "Неизвестный"}</span><span className="text-xs text-muted-foreground">{fmtRelative(h.fired_at || h.created_at)}</span></div>
                    <Link to={`/mentions?brand_id=${h.brand_id}`} className="text-xs text-primary hover:text-primary/80 flex items-center gap-1 mt-2 transition-colors cursor-pointer">К упоминаниям <ArrowRight className="h-3 w-3" /></Link>
                  </div >
                );
              })}
            </div >
          </>
        )}
      </section >

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => { if (!v) setDeleteTarget(null); }}
        title={`Удалить настройку алертов для «${deleteTarget?.brand_name}»?`}
        description="Алерты для этого бренда больше не будут приходить. Это действие нельзя отменить."
        confirmLabel="Удалить"
        onConfirm={handleDelete}
        destructive
      />
    </div >
  );
};

export default Alerts;
