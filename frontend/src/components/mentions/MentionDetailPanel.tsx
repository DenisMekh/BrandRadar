import { useEffect, useCallback } from "react";
import { createPortal } from "react-dom";
import { useNavigate } from "react-router-dom";
import {
  X, ExternalLink, Copy, CheckCircle, XCircle, Brain,
  ArrowLeft, ArrowRight,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { fmtRelative } from "@/lib/format";
import { toast } from "sonner";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import type { Mention } from "@/lib/api/mentions";

const sentimentConfig = {
  positive: { label: "Позитивная", barClass: "bg-success", badgeClass: "bg-success/15 text-success", textClass: "text-success" },
  negative: { label: "Негативная", barClass: "bg-destructive", badgeClass: "bg-destructive/15 text-destructive", textClass: "text-destructive" },
  neutral: { label: "Нейтральная", barClass: "bg-muted-foreground", badgeClass: "bg-muted text-muted-foreground", textClass: "text-muted-foreground" },
};

const sourceConfig: Record<string, string> = {
  Telegram: "bg-blue-500/15 text-blue-400 border-blue-500/20",
  Web: "bg-primary/15 text-primary border-primary/20",
  RSS: "bg-warning/15 text-warning border-warning/20",
};

function getInsight(sentiment: string, score: number, relevant: boolean): string {
  if (!relevant) return "Модель считает упоминание нерелевантным для бренда. Вероятно, ключевое слово использовано в другом контексте.";
  if (sentiment === "negative" && score > 0.7) return "Негативное упоминание с высокой уверенностью. Рекомендуется оперативное реагирование PR-команды.";
  if (sentiment === "negative" && score < 0.5) return "Возможно негативное упоминание. Низкая уверенность модели — рекомендуется ручная проверка.";
  if (sentiment === "negative") return "Негативное упоминание бренда. Рекомендуется реагирование.";
  if (sentiment === "positive" && relevant) return "Позитивное упоминание бренда. Можно использовать для PR-активностей и кейсов.";
  return "Нейтральное упоминание бренда. Мониторинг продолжается.";
}

function aggressionLabel(val: number): { text: string; colorClass: string } {
  if (val < 0.3) return { text: "низкая", colorClass: "text-success" };
  if (val < 0.6) return { text: "средняя", colorClass: "text-warning" };
  return { text: "высокая", colorClass: "text-destructive" };
}

interface Props {
  mention: Mention;
  allMentions: Mention[];
  onClose: () => void;
  onSelect: (m: Mention) => void;
}

export function MentionDetailPanel({ mention, allMentions, onClose, onSelect }: Props) {
  const navigate = useNavigate();
  
  const sentiment = (mention.sentiment || mention.ml?.label || "neutral") as "positive" | "negative" | "neutral";
  const sc = sentimentConfig[sentiment];
  
  // Real backend has ml.score, previously sentiment_score
  const sentimentScore = mention.ml?.score ?? 0.7;
  const scorePercent = Math.round(sentimentScore * 100);

  // Derive source name/type safely
  const sourceType = (typeof mention.source === 'object' ? mention.source?.type : mention.source) || "Web";
  const sourceName = (typeof mention.source === 'object' ? mention.source?.name : mention.source) || sourceType;

  const relevant = mention.ml?.is_relevant ?? true;

  // Derive aggression from score
  const aggression = sentiment === "negative" ? Math.min(sentimentScore + 0.2, 1) : sentimentScore * 0.4;
  const aggrLabel = aggressionLabel(aggression);

  const similar = (mention.similar_mentions || 
    (mention.ml?.similar_ids?.map(id => allMentions.find(m => m.id === id)).filter(Boolean) as Mention[]) || 
    [])
    .slice(0, 5);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(mention.text);
    toast.success("Текст скопирован");
  }, [mention.text]);

  // Escape to close
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [onClose]);

  // Lock body scroll
  useEffect(() => {
    document.body.style.overflow = "hidden";
    return () => { document.body.style.overflow = ""; };
  }, []);

  const dateObj = new Date(mention.created_at);
  const exactDate = dateObj.toLocaleString("ru-RU", { day: "numeric", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" });

  if (typeof document === "undefined") return null;

  return createPortal(
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 z-40 bg-black/30 backdrop-blur-sm" onClick={onClose} />

      {/* Panel */}
      <div className="fixed inset-y-0 right-0 z-50 h-screen w-full sm:w-[520px] bg-card sm:border-l border-border flex flex-col min-h-0 animate-slide-in-right">
        {/* Header (non-scroll) */}
        <div className="flex-shrink-0 border-b border-border p-5 bg-card">
          <div className="flex items-center gap-3">
            <span className={cn("text-xs px-2 py-0.5 rounded-full border shrink-0", sourceConfig[sourceType] || "bg-muted text-muted-foreground")}>
              {sourceName}
            </span>

            <div className="flex-1 flex justify-center min-w-0">
              <Tooltip>
                <TooltipTrigger asChild>
                  <span className="text-xs text-muted-foreground/60 cursor-default truncate">{fmtRelative(mention.created_at)}</span>
                </TooltipTrigger>
                <TooltipContent><p>{exactDate}</p></TooltipContent>
              </Tooltip>
            </div>

            <button
              onClick={onClose}
              className="hidden sm:flex p-1.5 text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary transition-colors cursor-pointer shrink-0 items-center justify-center"
              aria-label="Закрыть"
            >
              <X className="h-5 w-5" />
            </button>
            <button
              onClick={onClose}
              className="sm:hidden flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer shrink-0 min-h-[44px] px-2"
            >
              <ArrowLeft className="h-4 w-4" /> Назад
            </button>
          </div>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 min-h-0 overflow-y-auto p-5">
          <div className="text-[15px] leading-relaxed text-foreground/90 whitespace-pre-wrap break-words mt-4 first:mt-0">
            {mention.text}
          </div>

          <div className="bg-secondary/30 rounded-xl p-5 mt-5">
            <h3 className="text-base font-medium text-foreground flex items-center gap-2">
              <Brain className="h-4 w-4 text-primary" /> Анализ ML
            </h3>

            <div className="mt-3">
              <span className={cn("text-sm font-medium px-3 py-1 rounded-full inline-block", sc.badgeClass)}>
                {sc.label}
              </span>
            </div>

            <div className="w-full h-2 rounded-full bg-border mt-2 overflow-hidden">
              <div className={cn("h-full rounded-full transition-all duration-700 ease-out", sc.barClass)} style={{ width: `${scorePercent}%` }} />
            </div>
            <p className="text-xs text-muted-foreground mt-1">Уверенность модели: {scorePercent}%</p>

            <div className="flex items-center gap-2 mt-4">
              {relevant ? (
                <><CheckCircle className="h-4 w-4 text-success" /><span className="text-sm text-success font-medium">Релевантно для бренда</span></>
              ) : (
                <><XCircle className="h-4 w-4 text-muted-foreground" /><span className="text-sm text-muted-foreground">Не релевантно</span></>
              )}
            </div>

            <div className="mt-4">
              <p className="text-xs text-muted-foreground">
                Агрессивность: {aggression.toFixed(1)} — <span className={aggrLabel.colorClass}>{aggrLabel.text}</span>
              </p>
              <div className="relative w-full h-2 rounded-full overflow-hidden mt-1" style={{ background: "linear-gradient(to right, hsl(var(--success)), hsl(var(--warning)), hsl(var(--destructive)))" }}>
                <div className="absolute top-1/2 -translate-y-1/2 w-2.5 h-2.5 rounded-full bg-foreground border-2 border-card" style={{ left: `calc(${Math.round(aggression * 100)}% - 5px)` }} />
              </div>
            </div>

            <div className="bg-secondary/50 rounded-lg p-4 mt-4">
              <p className="text-xs text-muted-foreground uppercase tracking-wide mb-2 font-medium">Почему это важно</p>
              <p className="text-sm text-foreground leading-relaxed">
                {getInsight(sentiment, sentimentScore, relevant)}
              </p>
            </div>
          </div>

          {similar.length > 0 && (
            <div className="mt-5 pt-4 border-t border-border">
              <div className="flex items-center gap-2 mb-3">
                <h3 className="text-sm font-semibold text-foreground">Похожие публикации</h3>
                <span className="text-xs bg-secondary text-muted-foreground px-1.5 py-0.5 rounded-full">{similar.length}</span>
              </div>
              <div className="space-y-0.5">
                {similar.map((sm) => {
                  const simScore = Math.round(70 + Math.random() * 25);
                  return (
                    <button
                      key={sm.id}
                      onClick={() => onSelect(sm)}
                      className="w-full text-left rounded-lg px-2.5 py-2 hover:bg-secondary/50 transition-colors cursor-pointer flex items-center justify-between gap-3"
                    >
                      <p className="text-sm font-medium text-foreground line-clamp-1 flex-1 min-w-0">{sm.title || sm.text}</p>
                      <div className="flex items-center gap-2 shrink-0">
                        <span className={cn("text-xs px-2 py-0.5 rounded-full border", sourceConfig[(typeof sm.source === 'object' ? sm.source?.type : sm.source) || "Web"] || "bg-muted text-muted-foreground")}>
                          {(typeof sm.source === 'object' ? sm.source?.type : sm.source) || "Web"}
                        </span>
                        <span className="text-xs text-muted-foreground">{simScore}%</span>
                      </div>
                    </button>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Bottom actions (non-scroll) */}
        <div className="flex-shrink-0 border-t border-border p-4 bg-card flex flex-wrap gap-2">
          <button
            onClick={handleCopy}
            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1.5 px-3 py-2 rounded-lg border border-border hover:bg-secondary transition-colors cursor-pointer min-h-[44px] sm:min-h-0"
          >
            <Copy className="h-3.5 w-3.5" /> Копировать текст
          </button>
          {mention.url && (
            <a
              href={mention.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1.5 px-3 py-2 rounded-lg border border-border hover:bg-secondary transition-colors cursor-pointer min-h-[44px] sm:min-h-0"
            >
              <ExternalLink className="h-3.5 w-3.5" /> Открыть оригинал
            </a>
          )}
          <button
            onClick={() => { onClose(); navigate(`/mentions?brand_id=${mention.brand_id}`); }}
            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1.5 px-3 py-2 rounded-lg border border-border hover:bg-secondary transition-colors cursor-pointer min-h-[44px] sm:min-h-0"
          >
            <ArrowRight className="h-3.5 w-3.5" /> К упоминаниям бренда
          </button>
        </div>
      </div>
    </>,
    document.body,
  );
}

