import { X, Pencil, Trash2 } from "lucide-react";

interface Brand {
  id: string;
  name: string;
  keywords: string[];
  exclusions: string[];
  riskWords: string[];
}

interface BrandCardProps {
  brand: Brand;
  onEdit: () => void;
  onDelete: () => void;
}

export function BrandCard({ brand, onEdit, onDelete }: BrandCardProps) {
  return (
    <div className="bg-card border border-border rounded-xl p-5 space-y-3 hover:scale-[1.005] transition-transform duration-200">
      <div className="flex items-center justify-between">
        <h4 className="text-lg font-medium text-foreground">{brand.name}</h4>
        <div className="flex gap-1">
          <button
            onClick={onEdit}
            className="p-1.5 text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary transition-colors cursor-pointer"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            onClick={onDelete}
            className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg hover:bg-destructive/10 transition-colors cursor-pointer"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>

      {brand.keywords.length > 0 && (
        <div>
          <p className="text-xs text-muted-foreground mb-1.5">Ключевые слова</p>
          <div className="flex flex-wrap gap-1.5">
            {brand.keywords.map((kw) => (
              <span key={kw} className="bg-primary/10 text-primary text-xs rounded-full px-3 py-1">
                {kw}
              </span>
            ))}
          </div>
        </div>
      )}

      {brand.exclusions.length > 0 && (
        <div>
          <p className="text-xs text-muted-foreground mb-1.5">Исключения</p>
          <div className="flex flex-wrap gap-1.5">
            {brand.exclusions.map((ex) => (
              <span key={ex} className="bg-secondary text-muted-foreground text-xs rounded-full px-3 py-1 flex items-center gap-1">
                <X className="h-3 w-3" /> {ex}
              </span>
            ))}
          </div>
        </div>
      )}

      {brand.riskWords.length > 0 && (
        <div>
          <p className="text-xs text-muted-foreground mb-1.5">Risk-слова</p>
          <div className="flex flex-wrap gap-1.5">
            {brand.riskWords.map((rw) => (
              <span key={rw} className="bg-destructive/10 text-destructive text-xs rounded-full px-3 py-1">
                {rw}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
