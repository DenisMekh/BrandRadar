import { useState } from "react";
import { Plus, Loader2, Globe } from "lucide-react";
import { cn } from "@/lib/utils";
import { Switch } from "@/components/ui/switch";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Skeleton } from "@/components/shared/Skeleton";
import { EmptyState } from "@/components/shared/EmptyState";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { OfflineGuard } from "@/components/shared/OfflineGuard";
import { useSources, useCreateSource, useToggleSource, useCollectorJobs } from "@/hooks/use-sources";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum, fmtRelative } from "@/lib/format";
import { toast } from "sonner";
import { DEFAULT_PROJECT_ID } from "@/lib/constants";

const typeColors: Record<string, string> = { Web: "bg-primary/15 text-primary", RSS: "bg-warning/15 text-warning", Telegram: "bg-blue-500/15 text-blue-400" };
const taskStatusConfig: Record<string, { label: string; className: string }> = {
  idle: { label: "Idle", className: "bg-muted text-muted-foreground" },
  running: { label: "Running", className: "bg-blue-500/15 text-blue-400" },
  completed: { label: "Completed", className: "bg-success/15 text-success" },
  failed: { label: "Failed", className: "bg-destructive/15 text-destructive" },
};

const TableSkeleton = ({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) => (
  <div className="bg-card border border-border rounded-xl overflow-hidden">
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex gap-4 px-5 py-3 border-b border-border last:border-0">
        {Array.from({ length: cols }).map((_, j) => <Skeleton key={j} className="flex-1 h-4" />)}
      </div>
    ))}
  </div>
);

const Sources = () => {
  usePageTitle("Источники");

  const { data: sources, isLoading: sourcesLoading, isError: sourcesError, refetch: refetchSources } = useSources();
  // const { data: jobs, isLoading: jobsLoading, isError: jobsError, refetch: refetchJobs } = useCollectorJobs();
  const jobs: { id: string; source_name: string; status: string; found: number; created_at: string }[] = [];
  const jobsLoading = false;
  const jobsError = false;
  const refetchJobs = () => { };
  const createSource = useCreateSource();
  const toggleSource = useToggleSource();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [newSource, setNewSource] = useState({ name: "", type: "Web" as "Web" | "RSS" | "Telegram", url: "" });
  const [formErrors, setFormErrors] = useState<{ name?: string; url?: string }>({});
  const [toggleTarget, setToggleTarget] = useState<{ id: string; name: string; status: string } | null>(null);

  const handleAdd = () => {
    const errors: typeof formErrors = {};
    if (!newSource.name.trim()) errors.name = "Обязательное поле";

    const url = newSource.url.trim();
    if (!url) {
      errors.url = "Обязательное поле";
    } else if (newSource.type === "Telegram" && !/^[a-zA-Z0-9_]+$/.test(url)) {
      errors.url = "Идентификатор должен содержать только латиницу, цифры и «_»";
    } else if (["RSS", "Web"].includes(newSource.type) && !url.startsWith("http://") && !url.startsWith("https://")) {
      errors.url = "Для RSS и Web URL должен начинаться с http:// или https://";
    }

    setFormErrors(errors);
    if (Object.keys(errors).length > 0) return;

    createSource.mutate({ data: { type: newSource.type, name: newSource.name, url: url } }, {
      onSuccess: () => toast.success("Источник добавлен"),
      onError: (e) => toast.error(`Не удалось добавить: ${e.message}`),
    });
    setNewSource({ name: "", type: "Web", url: "" });
    setFormErrors({});
    setDialogOpen(false);
  };

  const handleToggle = (s: { id: string; name: string; status: string }) => {
    if (s.status === "active") {
      setToggleTarget(s);
    } else {
      toggleSource.mutate(s.id, {
        onSuccess: () => toast.success("Источник включён"),
        onError: (e) => toast.error(`Ошибка: ${e.message}`),
      });
    }
  };

  const handleToggleConfirm = () => {
    if (!toggleTarget) return;
    toggleSource.mutate(toggleTarget.id, {
      onSuccess: () => toast.success("Источник отключён"),
      onError: (e) => toast.error(`Ошибка: ${e.message}`),
    });
    setToggleTarget(null);
  };

  const isLoading = sourcesLoading || jobsLoading;
  const hasError = sourcesError && jobsError;

  if (hasError) return <ErrorBanner onRetry={() => { refetchSources(); refetchJobs(); }} />;

  return (
    <div className="space-y-8">
      <section>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-foreground">Источники</h3>
          <OfflineGuard>
            <button onClick={() => setDialogOpen(true)} className="text-sm text-primary hover:text-primary/80 flex items-center gap-1.5 transition-colors cursor-pointer">
              <Plus className="h-4 w-4" /> Добавить источник
            </button>
          </OfflineGuard>
        </div>
        {sourcesLoading ? <TableSkeleton rows={5} cols={4} /> : (sources ?? []).length === 0 ? (
          <EmptyState
            icon={Globe}
            title="Добавьте источник для начала сбора"
            description="Источники данных используются для мониторинга упоминаний бренда."
            actionLabel="Добавить источник"
            onAction={() => setDialogOpen(true)}
          />
        ) : (
          <>
            {/* Desktop table */}
            <div className="hidden md:block bg-card border border-border rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead><tr className="border-b border-border text-muted-foreground text-left">
                  <th className="px-5 py-3 font-medium">Имя</th><th className="px-5 py-3 font-medium">Тип</th><th className="px-5 py-3 font-medium">Статус</th><th className="px-5 py-3 font-medium">Последний сбор</th>
                </tr></thead>
                <tbody>
                  {(sources ?? []).map((s) => (
                    <tr key={s.id} className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors">
                      <td className="px-5 py-3 text-foreground font-medium truncate max-w-[200px]">{s.name}</td>
                      <td className="px-5 py-3"><span className={cn("text-xs px-2 py-0.5 rounded-full font-medium", typeColors[s.type])}>{s.type}</span></td>
                      <td className="px-5 py-3">
                        <Switch
                          checked={s.status === "active"}
                          onCheckedChange={() => handleToggle({ id: s.id || "", name: s.name || "", status: s.status || "inactive" })}
                        />
                      </td>
                      <td className="px-5 py-3 text-muted-foreground">{fmtRelative(s.updated_at || s.created_at || "")}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {/* Mobile cards */}
            <div className="md:hidden space-y-3">
              {(sources ?? []).map((s) => (
                <div key={s.id} className="bg-card border border-border rounded-xl p-4 space-y-3 hover:scale-[1.01] transition-transform duration-200">
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-foreground truncate max-w-[180px]">{s.name}</span>
                    <span className={cn("text-xs px-2 py-0.5 rounded-full font-medium", typeColors[s.type || ""])}>{s.type}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">{fmtRelative(s.updated_at || s.created_at || "")}</span>
                    <Switch
                      checked={s.status === "active"}
                      onCheckedChange={() => handleToggle({ id: s.id || "", name: s.name || "", status: s.status || "inactive" })}
                    />
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </section>

      {/* TODO: Enable when jobs polling is implemented
      <section>
        <h3 className="text-lg font-medium text-foreground mb-4">Последние задачи сбора</h3>
        {jobsLoading ? <TableSkeleton rows={5} cols={4} /> : (
          <>
            <div className="hidden md:block bg-card border border-border rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead><tr className="border-b border-border text-muted-foreground text-left">
                  <th className="px-5 py-3 font-medium">Источник</th><th className="px-5 py-3 font-medium">Статус</th><th className="px-5 py-3 font-medium">Найдено</th><th className="px-5 py-3 font-medium">Время</th>
                </tr></thead>
                <tbody>
                  {(jobs ?? []).map((t) => {
                    const cfg = taskStatusConfig[t.status] ?? taskStatusConfig.idle; return (
                      <tr key={t.id} className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors">
                        <td className="px-5 py-3 text-foreground font-medium">{t.source_name}</td>
                        <td className="px-5 py-3"><span className={cn("text-xs px-2 py-0.5 rounded-full font-medium inline-flex items-center gap-1", cfg.className)}>{t.status === "running" && <Loader2 className="h-3 w-3 animate-spin" />}{cfg.label}</span></td>
                        <td className="px-5 py-3 text-muted-foreground">{fmtNum(t.found)}</td>
                        <td className="px-5 py-3 text-muted-foreground">{fmtRelative(t.created_at)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            <div className="md:hidden space-y-3">
              {(jobs ?? []).map((t) => {
                const cfg = taskStatusConfig[t.status] ?? taskStatusConfig.idle; return (
                  <div key={t.id} className="bg-card border border-border rounded-xl p-4 hover:scale-[1.01] transition-transform duration-200">
                    <div className="flex items-center justify-between"><span className="font-medium text-foreground text-sm">{t.source_name}</span><span className={cn("text-xs px-2 py-0.5 rounded-full font-medium inline-flex items-center gap-1", cfg.className)}>{t.status === "running" && <Loader2 className="h-3 w-3 animate-spin" />}{cfg.label}</span></div>
                    <div className="flex items-center justify-between mt-2 text-xs text-muted-foreground"><span>Найдено: {fmtNum(t.found)}</span><span>{fmtRelative(t.created_at)}</span></div>
                  </div>
                );
              })}
            </div>
          </>
        )}
      </section>
      */}

      {/* Add source dialog */}
      <Dialog open={dialogOpen} onOpenChange={(v) => { setDialogOpen(v); if (!v) setFormErrors({}); }}>
        <DialogContent className="bg-card border-border sm:max-w-md">
          <DialogHeader><DialogTitle className="text-foreground">Добавить источник</DialogTitle></DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5"><label className="text-sm text-muted-foreground">Тип</label><select value={newSource.type} onChange={(e) => setNewSource({ ...newSource, type: e.target.value as "Web" | "RSS" | "Telegram" })} className="w-full bg-secondary border border-border text-foreground text-sm rounded-lg px-3 py-2 focus:outline-none focus:ring-1 focus:ring-primary"><option value="Web">Web</option><option value="RSS">RSS</option><option value="Telegram">Telegram</option></select></div>
            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">Название *</label>
              <input value={newSource.name} onChange={(e) => { setNewSource({ ...newSource, name: e.target.value }); if (formErrors.name) setFormErrors({}); }} placeholder="Название источника" className={cn("w-full bg-secondary border rounded-lg px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary", formErrors.name ? "border-destructive" : "border-border")} />
              {formErrors.name && <p className="text-xs text-destructive">{formErrors.name}</p>}
            </div>
            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">URL</label>
              <input value={newSource.url} onChange={(e) => { setNewSource({ ...newSource, url: e.target.value }); if (formErrors.url) setFormErrors({ ...formErrors, url: undefined }); }} placeholder={newSource.type === "Telegram" ? "channel_name" : "http://..."} className={cn("w-full bg-secondary border rounded-lg px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary", formErrors.url ? "border-destructive" : "border-border")} />
              {formErrors.url && <p className="text-xs text-destructive">{formErrors.url}</p>}
            </div>
          </div>
          <DialogFooter className="gap-2">
            <button onClick={() => { setDialogOpen(false); setFormErrors({}); }} className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer">Отмена</button>
            <button onClick={handleAdd} className="px-4 py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer">Добавить</button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={!!toggleTarget}
        onOpenChange={(v) => { if (!v) setToggleTarget(null); }}
        title={`Отключить источник «${toggleTarget?.name}»?`}
        description="Сбор данных будет остановлен. Вы сможете включить его обратно в любое время."
        confirmLabel="Отключить"
        onConfirm={handleToggleConfirm}
        destructive
      />
    </div>
  );
};

export default Sources;
