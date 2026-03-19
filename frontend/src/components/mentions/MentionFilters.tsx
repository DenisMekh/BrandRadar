import { useState } from "react";
import { Search, CalendarIcon, SlidersHorizontal } from "lucide-react";
import { format } from "date-fns";
import { ru } from "date-fns/locale";
import { cn } from "@/lib/utils";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useBrands } from "@/hooks/use-brands";
import { useSources } from "@/hooks/use-sources";

export type Sentiment = "positive" | "negative" | "neutral";

const sentimentOptions = [
  { value: "all", label: "Все" },
  { value: "positive", label: "Позитивная" },
  { value: "negative", label: "Негативная" },
  { value: "neutral", label: "Нейтральная" },
] as const;


interface MentionFiltersProps {
  search: string;
  onSearchChange: (v: string) => void;
  sentiment: "all" | Sentiment;
  onSentimentChange: (v: "all" | Sentiment) => void;
  onlyRelevant: boolean;
  onRelevantChange: (v: boolean) => void;
  dateFrom: Date | undefined;
  onDateFromChange: (v: Date | undefined) => void;
  dateTo: Date | undefined;
  onDateToChange: (v: Date | undefined) => void;
  brandId: string;
  onBrandIdChange: (v: string) => void;
  sourceId: string;
  onSourceIdChange: (v: string) => void;
}

export function MentionFilters({
  search, onSearchChange,
  sentiment, onSentimentChange,
  onlyRelevant, onRelevantChange,
  dateFrom, onDateFromChange,
  dateTo, onDateToChange,
  brandId, onBrandIdChange,
  sourceId, onSourceIdChange,
}: MentionFiltersProps) {
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [fromOpen, setFromOpen] = useState(false);
  const [toOpen, setToOpen] = useState(false);
  const { data: brands } = useBrands();
  const { data: sources } = useSources();

  const activeCount = [sentiment !== "all", onlyRelevant, !!dateFrom, !!dateTo, brandId !== "all", sourceId !== "all"].filter(Boolean).length;

  return (
    <div className="sticky top-0 z-10 bg-background/80 backdrop-blur-md border-b border-border -mx-4 sm:-mx-6 px-4 sm:px-6 py-3 sm:py-4 space-y-3">
      {/* Row 1: search + filter toggle (mobile) */}
      <div className="flex items-center gap-2 sm:gap-3">
        <div className="relative flex-1 min-w-0">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Поиск по тексту упоминания..."
            className="w-full bg-card border border-border rounded-lg pl-9 pr-3 py-2.5 sm:py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
          />
        </div>

        <button
          onClick={() => setFiltersOpen(!filtersOpen)}
          className={cn(
            "sm:hidden flex items-center justify-center w-11 h-11 rounded-lg border transition-colors cursor-pointer shrink-0 relative",
            activeCount > 0 ? "bg-primary/20 border-primary/30 text-primary" : "bg-card border-border text-muted-foreground hover:text-foreground"
          )}
        >
          <SlidersHorizontal className="h-4.5 w-4.5" />
          {activeCount > 0 && (
            <span className="absolute -top-1.5 -right-1.5 bg-primary text-primary-foreground text-[10px] font-bold w-4.5 h-4.5 rounded-full flex items-center justify-center">{activeCount}</span>
          )}
        </button>

        {/* Desktop sentiment buttons */}
        <div className="hidden sm:flex rounded-lg border border-border overflow-hidden">
          {sentimentOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => onSentimentChange(opt.value as "all" | Sentiment)}
              className={cn(
                "px-3 py-2 text-sm font-medium transition-all duration-200 cursor-pointer",
                sentiment === opt.value ? "bg-primary text-primary-foreground" : "bg-card text-muted-foreground hover:text-foreground"
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Collapsible filters */}
      <div className={cn("space-y-3", !filtersOpen && "hidden sm:block")}>
        {/* Mobile sentiment */}
        <div className="flex sm:hidden rounded-lg border border-border overflow-hidden">
          {sentimentOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => onSentimentChange(opt.value as "all" | Sentiment)}
              className={cn(
                "flex-1 px-2 py-2.5 text-xs font-medium transition-all duration-200 cursor-pointer",
                sentiment === opt.value ? "bg-primary text-primary-foreground" : "bg-card text-muted-foreground"
              )}
            >
              {opt.label}
            </button>
          ))}
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-3">
          {/* Brand filter */}
          <select
            value={brandId}
            onChange={(e) => onBrandIdChange(e.target.value)}
            className="bg-card border border-border text-foreground text-sm rounded-lg px-3 py-2.5 sm:py-2 focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer min-h-[44px] sm:min-h-0"
          >
            <option value="all">Все бренды</option>
            {(brands ?? []).map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
          </select>

          <select
            value={sourceId}
            onChange={(e) => onSourceIdChange(e.target.value)}
            className="bg-card border border-border text-foreground text-sm rounded-lg px-3 py-2.5 sm:py-2 focus:outline-none focus:ring-1 focus:ring-primary cursor-pointer min-h-[44px] sm:min-h-0"
          >
            <option value="all">Все источники</option>
            {(sources ?? []).map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>

          <label className="flex items-center gap-2 text-sm text-muted-foreground cursor-pointer min-h-[44px] sm:min-h-0">
            <Switch checked={onlyRelevant} onCheckedChange={onRelevantChange} />
            Только релевантные
          </label>

          <div className="flex gap-2">
            <Popover open={fromOpen} onOpenChange={setFromOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" className={cn("text-sm gap-2 cursor-pointer flex-1 sm:flex-initial min-h-[44px] sm:min-h-0", !dateFrom && "text-muted-foreground")}>
                  <CalendarIcon className="h-4 w-4" />
                  {dateFrom ? format(dateFrom, "dd.MM.yyyy", { locale: ru }) : "От"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar mode="single" selected={dateFrom} onSelect={(d) => { onDateFromChange(d); setFromOpen(false); }} initialFocus className="p-3 pointer-events-auto" />
              </PopoverContent>
            </Popover>

            <Popover open={toOpen} onOpenChange={setToOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" className={cn("text-sm gap-2 cursor-pointer flex-1 sm:flex-initial min-h-[44px] sm:min-h-0", !dateTo && "text-muted-foreground")}>
                  <CalendarIcon className="h-4 w-4" />
                  {dateTo ? format(dateTo, "dd.MM.yyyy", { locale: ru }) : "До"}
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto p-0" align="start">
                <Calendar mode="single" selected={dateTo} onSelect={(d) => { onDateToChange(d); setToOpen(false); }} initialFocus className="p-3 pointer-events-auto" />
              </PopoverContent>
            </Popover>
          </div>
        </div>
      </div>
    </div>
  );
}
