import { useState, useEffect, useCallback, KeyboardEvent } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerFooter,
} from "@/components/ui/drawer";
import { useIsMobile } from "@/hooks/use-mobile";

export interface Brand {
  id: string;
  name: string;
  keywords: string[];
  exclusions: string[];
  riskWords: string[];
}

interface BrandDialogProps {
  open: boolean;
  onOpenChange: (v: boolean) => void;
  brand: Brand | null;
  onSave: (brand: Brand) => void;
}

function TagInput({
  label,
  tags,
  onChange,
  variant = "default",
  error,
}: {
  label: string;
  tags: string[];
  onChange: (tags: string[]) => void;
  variant?: "default" | "risk";
  error?: string;
}) {
  const [input, setInput] = useState("");

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && input.trim()) {
      e.preventDefault();
      if (!tags.includes(input.trim())) {
        onChange([...tags, input.trim()]);
      }
      setInput("");
    }
  };

  const chipClass =
    variant === "risk"
      ? "bg-destructive/10 text-destructive"
      : "bg-primary/10 text-primary";

  return (
    <div className="space-y-1.5">
      <label className="text-sm text-muted-foreground">{label}</label>
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-1.5">
          {tags.map((tag) => (
            <span key={tag} className={`text-xs rounded-full px-2.5 py-1 flex items-center gap-1 ${chipClass}`}>
              {tag}
              <button onClick={() => onChange(tags.filter((t) => t !== tag))} className="hover:opacity-70 cursor-pointer">
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
      )}
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Введите и нажмите Enter"
        className={cn(
          "w-full bg-secondary border rounded-lg px-3 py-2.5 sm:py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary min-h-[44px] sm:min-h-0",
          error ? "border-destructive" : "border-border"
        )}
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}

function BrandForm({ brand, onSave, onCancel }: { brand: Brand | null; onSave: (b: Brand) => void; onCancel: () => void }) {
  const [name, setName] = useState("");
  const [keywords, setKeywords] = useState<string[]>([]);
  const [exclusions, setExclusions] = useState<string[]>([]);
  const [riskWords, setRiskWords] = useState<string[]>([]);
  const [errors, setErrors] = useState<{ name?: string; keywords?: string }>({});
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    if (brand) {
      setName(brand.name);
      setKeywords([...brand.keywords]);
      setExclusions([...brand.exclusions]);
      setRiskWords([...brand.riskWords]);
    } else {
      setName("");
      setKeywords([]);
      setExclusions([]);
      setRiskWords([]);
    }
    setErrors({});
    setSubmitted(false);
  }, [brand]);

  const validate = useCallback(() => {
    const e: typeof errors = {};
    if (!name.trim()) e.name = "Обязательное поле";
    else if (name.trim().length > 50) e.name = "Максимум 50 символов";
    if (keywords.length === 0) e.keywords = "Добавьте хотя бы одно ключевое слово";
    setErrors(e);
    return Object.keys(e).length === 0;
  }, [name, keywords]);

  const handleSave = () => {
    setSubmitted(true);
    if (!validate()) return;
    onSave({
      id: brand?.id || Date.now().toString(),
      name: name.trim(),
      keywords,
      exclusions,
      riskWords,
    });
  };

  useEffect(() => {
    if (submitted) validate();
  }, [submitted, validate]);

  return (
    <>
      <div className="space-y-4 py-2 px-1">
        <div className="space-y-1.5">
          <label className="text-sm text-muted-foreground">Название бренда</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Название"
            maxLength={50}
            className={cn(
              "w-full bg-secondary border rounded-lg px-3 py-2.5 sm:py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary min-h-[44px] sm:min-h-0",
              errors.name ? "border-destructive" : "border-border"
            )}
            autoFocus
          />
          {errors.name && <p className="text-xs text-destructive">{errors.name}</p>}
        </div>

        <TagInput label="Ключевые слова *" tags={keywords} onChange={setKeywords} error={errors.keywords} />
        <TagInput label="Слова-исключения" tags={exclusions} onChange={setExclusions} />
        <TagInput label="Risk-слова" tags={riskWords} onChange={setRiskWords} variant="risk" />
      </div>

      <div className="flex gap-2 justify-end pt-2">
        <button
          onClick={onCancel}
          className="px-4 py-2.5 sm:py-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer min-h-[44px] sm:min-h-0"
        >
          Отмена
        </button>
        <button
          onClick={handleSave}
          className="px-4 py-2.5 sm:py-2 bg-primary text-primary-foreground text-sm rounded-lg hover:bg-primary/90 transition-colors cursor-pointer min-h-[44px] sm:min-h-0"
        >
          {brand ? "Сохранить" : "Создать"}
        </button>
      </div>
    </>
  );
}

export function BrandDialog({ open, onOpenChange, brand, onSave }: BrandDialogProps) {
  const isMobile = useIsMobile();
  const title = brand ? "Редактировать бренд" : "Новый бренд";

  if (isMobile) {
    return (
      <Drawer open={open} onOpenChange={onOpenChange}>
        <DrawerContent className="max-h-[90vh]">
          <DrawerHeader>
            <DrawerTitle>{title}</DrawerTitle>
          </DrawerHeader>
          <div className="px-4 pb-4 overflow-y-auto">
            <BrandForm brand={brand} onSave={onSave} onCancel={() => onOpenChange(false)} />
          </div>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-card border-border sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="text-foreground">{title}</DialogTitle>
        </DialogHeader>
        <BrandForm brand={brand} onSave={onSave} onCancel={() => onOpenChange(false)} />
      </DialogContent>
    </Dialog>
  );
}
