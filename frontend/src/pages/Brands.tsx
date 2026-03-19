import { useState } from "react";
import { Link } from "react-router-dom";
import { Plus, Tag, Pencil, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Skeleton } from "@/components/shared/Skeleton";
import { EmptyState } from "@/components/shared/EmptyState";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { OfflineGuard } from "@/components/shared/OfflineGuard";
import { BrandDialog } from "@/components/settings/BrandDialog";
import { useBrands, useCreateBrand, useDeleteBrand, useUpdateBrand, useBrandDashboard } from "@/hooks/use-brands";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum } from "@/lib/format";
import { toast } from "sonner";
import type { Brand as ApiBrand } from "@/lib/api/brands";
import { DEFAULT_PROJECT_ID } from "@/lib/constants";

interface LocalBrand {
  id: string;
  name: string;
  keywords: string[];
  exclusions: string[];
  riskWords: string[];
}

function toLocal(b: ApiBrand): LocalBrand {
  return { id: b.id, name: b.name, keywords: b.keywords, exclusions: b.exclusions, riskWords: b.risk_words };
}

function TagList({ tags, variant, max = 5 }: { tags: string[]; variant: "violet" | "red" | "muted"; max?: number }) {
  const shown = tags.slice(0, max);
  const rest = tags.length - max;
  const cls = variant === "violet" ? "bg-violet-500/15 text-violet-400" : variant === "red" ? "bg-destructive/15 text-destructive" : "bg-secondary text-muted-foreground";
  return (
    <div className="flex flex-wrap gap-1.5">
      {shown.map((t) => <span key={t} className={cn("text-xs rounded-full px-2.5 py-1", cls)}>{t}</span>)}
      {rest > 0 && <span className="text-xs rounded-full px-2.5 py-1 bg-secondary text-muted-foreground">+{rest} ещё</span>}
    </div>
  );
}

function BrandCard({ b, onEdit, onDelete }: { b: ApiBrand; onEdit: (b: LocalBrand) => void; onDelete: (b: ApiBrand) => void; }) {
  const { data: dashboard, isLoading } = useBrandDashboard(b.id);
  const totalMentions = dashboard?.total_mentions ?? 0;
  const negativeMentions = dashboard?.sentiment?.negative ?? 0;

  return (
    <div className="bg-card border border-border rounded-xl p-6 flex flex-col gap-4 hover:scale-[1.01] transition-transform duration-200 relative group">
      <Link to={`/brands/${b.id}`} className="absolute inset-0 z-0" />
      <div className="flex items-center justify-between relative z-10 pointer-events-none">
        <h3 className="text-xl font-semibold text-foreground truncate max-w-[200px]" title={b.name}>{b.name}</h3>
        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto">
          <button
            onClick={(e) => { e.preventDefault(); e.stopPropagation(); onEdit(toLocal(b)); }}
            className="p-1.5 text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary transition-colors cursor-pointer"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={(e) => { e.preventDefault(); e.stopPropagation(); onDelete(b); }}
            className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg hover:bg-destructive/10 transition-colors cursor-pointer"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>

      {b.keywords.length > 0 && <div className="relative z-10 pointer-events-none"><TagList tags={b.keywords} variant="violet" /></div>}
      {b.risk_words.length > 0 && <div className="relative z-10 pointer-events-none"><TagList tags={b.risk_words} variant="red" /></div>}
      {b.exclusions.length > 0 && <div className="relative z-10 pointer-events-none"><TagList tags={b.exclusions} variant="muted" /></div>}

      {/* Stats pinned to bottom */}
      <div className="relative z-10 flex gap-3 pt-1 pointer-events-none mt-auto">
        <span className="text-sm text-muted-foreground flex items-center gap-1.5">
          {isLoading ? <Skeleton className="w-6 h-4" /> : fmtNum(totalMentions)} упоминаний
        </span>
        <span className="text-sm text-muted-foreground flex items-center gap-1.5">
          {isLoading ? <Skeleton className="w-6 h-4" /> : fmtNum(negativeMentions)} негативных
        </span>
      </div>
    </div>
  );
}

const Brands = () => {
  usePageTitle("Бренды");

  const { data: brands, isLoading, isError, refetch } = useBrands();
  const createBrand = useCreateBrand();
  const updateBrand = useUpdateBrand();
  const deleteBrand = useDeleteBrand();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingBrand, setEditingBrand] = useState<LocalBrand | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ApiBrand | null>(null);

  const handleSave = (brand: LocalBrand) => {
    const apiData = { name: brand.name, keywords: brand.keywords, exclusions: brand.exclusions, risk_words: brand.riskWords };

    const existing = brands?.find((b) => b.name.toLowerCase() === brand.name.toLowerCase() && b.id !== brand.id);
    if (existing) {
      toast.warning("Такой бренд уже существует");
      return;
    }

    if (editingBrand) {
      updateBrand.mutate({ brandId: brand.id, data: apiData }, {
        onSuccess: () => toast.success("Бренд обновлён"),
        onError: (e) => toast.error(`Не удалось сохранить: ${e.message}`),
      });
    } else {
      createBrand.mutate({ data: apiData }, {
        onSuccess: () => toast.success("Бренд создан"),
        onError: (e) => toast.error(`Не удалось сохранить: ${e.message}`),
      });
    }
    setDialogOpen(false);
    setEditingBrand(null);
  };

  const confirmDelete = () => {
    if (!deleteTarget) return;
    deleteBrand.mutate({ brandId: deleteTarget.id }, {
      onSuccess: () => toast.success("Бренд удалён"),
      onError: (e) => toast.error(`Не удалось удалить: ${e.message}`),
    });
    setDeleteTarget(null);
  };

  if (isError) return <ErrorBanner onRetry={() => refetch()} />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{isLoading ? "" : `${brands?.length ?? 0} брендов`}</p>
        <OfflineGuard>
          <button
            onClick={() => { setEditingBrand(null); setDialogOpen(true); }}
            className="text-sm text-primary hover:text-primary/80 flex items-center gap-1.5 transition-colors cursor-pointer"
          >
            <Plus className="h-4 w-4" /> Новый бренд
          </button>
        </OfflineGuard>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-card border border-border rounded-xl p-6 space-y-4">
              <Skeleton className="w-32 h-6" /><Skeleton className="w-full h-4" /><Skeleton className="w-3/4 h-4" /><Skeleton className="w-1/2 h-4" />
            </div>
          ))}
        </div>
      ) : (brands ?? []).length === 0 ? (
        <EmptyState
          icon={Tag}
          title="Брендов пока нет"
          description="Создайте первый бренд, чтобы начать мониторинг."
          actionLabel="Создать бренд"
          onAction={() => { setEditingBrand(null); setDialogOpen(true); }}
        />
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
          {(brands ?? []).map((b) => (
            <BrandCard 
              key={b.id} 
              b={b} 
              onEdit={(lb) => { setEditingBrand(lb); setDialogOpen(true); }} 
              onDelete={(sb) => setDeleteTarget(sb)} 
            />
          ))}
        </div>
      )}

      <BrandDialog open={dialogOpen} onOpenChange={setDialogOpen} brand={editingBrand} onSave={handleSave} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(v) => { if (!v) setDeleteTarget(null); }}
        title={`Удалить бренд «${deleteTarget?.name}»?`}
        description="Все упоминания и алерты будут потеряны. Это действие нельзя отменить."
        confirmLabel="Удалить"
        onConfirm={confirmDelete}
        destructive
      />
    </div>
  );
};

export default Brands;
