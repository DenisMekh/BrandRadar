import { useState } from "react";
import { Copy, ExternalLink } from "lucide-react";
import { cn } from "@/lib/utils";
import { fmtRelative } from "@/lib/format";

interface Mention {
  id: string;
  author: string;
  source: string;
  time: string;
  title?: string;
  text: string;
  sentiment: "positive" | "negative" | "neutral";
  sentimentScore: number;
  relevant: boolean;
  similarCount: number;
  url?: string;
}

const sentimentConfig = {
  positive: { label: "Позитивная", dotClass: "text-success", badgeClass: "bg-success/10 text-success" },
  negative: { label: "Негативная", dotClass: "text-destructive", badgeClass: "bg-destructive/10 text-destructive" },
  neutral: { label: "Нейтральная", dotClass: "text-muted-foreground", badgeClass: "bg-muted text-muted-foreground" },
};

const sourceConfig: Record<string, string> = {
  telegram: "bg-blue-500/15 text-blue-400 border-blue-500/20",
  web: "bg-primary/15 text-primary border-primary/20",
  rss: "bg-warning/15 text-warning border-warning/20",
  // Title case fallbacks
  Telegram: "bg-blue-500/15 text-blue-400 border-blue-500/20",
  Web: "bg-primary/15 text-primary border-primary/20",
  RSS: "bg-warning/15 text-warning border-warning/20",
};

interface MentionCardProps {
  mention: Mention;
}

export function MentionCard({ mention }: MentionCardProps) {
  const [expanded, setExpanded] = useState(false);
  const sc = sentimentConfig[mention.sentiment];

  const barColor = mention.sentiment === "positive"
    ? "bg-success"
    : mention.sentiment === "negative"
      ? "bg-destructive"
      : "bg-muted-foreground";

  return (
    <div className="bg-card border border-border rounded-xl p-5 hover:border-muted-foreground/30 hover:scale-[1.005] transition-all duration-200">
      {/* Top row */}
      <div className="flex items-center justify-between gap-2 flex-wrap">
        <div className="flex items-center gap-2">
          <span className="font-medium text-foreground text-sm">{mention.author}</span>
          <span className={cn("text-xs px-2 py-0.5 rounded-full border", sourceConfig[mention.source] || sourceConfig[mention.source.toLowerCase()] || "bg-muted text-muted-foreground")}>
            {mention.source}
          </span>
        </div>
        <span className="text-sm text-muted-foreground/60">{fmtRelative(mention.time)}</span>
      </div>

      {/* Text */}
      <div className="mt-3">
        {mention.title && (
          <h4 className="text-base font-semibold text-foreground mb-1 break-words">
            {mention.title}
          </h4>
        )}
        <div
          className={cn(
            "text-[15px] leading-relaxed text-foreground/90 cursor-pointer break-words whitespace-pre-wrap",
            !expanded && "line-clamp-4"
          )}
          onClick={() => setExpanded(!expanded)}
        >
          {mention.text}
        </div>
      </div>
      {!expanded && mention.text.length > 150 && (
        <button
          onClick={() => setExpanded(true)}
          className="text-xs text-primary hover:text-primary/80 mt-1 transition-colors cursor-pointer"
        >
          показать полностью
        </button>
      )}

      {/* ML block */}
      <div className="mt-3 pt-3 border-t border-border flex flex-wrap items-center gap-4">
        <span className={cn("text-xs font-medium px-2.5 py-1 rounded-full flex items-center gap-1.5", sc.badgeClass)}>
          <span className={sc.dotClass}>●</span>
          {sc.label} {mention.sentimentScore}%
        </span>

        {mention.relevant ? (
          <span className="text-xs text-success font-medium">✓ Релевантно</span>
        ) : (
          <span className="text-xs text-muted-foreground">✗ Не релевантно</span>
        )}

        {mention.similarCount > 0 && (
          <button className="text-xs text-primary flex items-center gap-1 hover:text-primary/80 transition-colors cursor-pointer">
            <Copy className="h-3 w-3" />
            {mention.similarCount} похожих
          </button>
        )}

        <div className="flex items-center gap-2">
          <span className="text-[10px] text-muted-foreground/60 uppercase tracking-wide">Score</span>
          <div className="w-[60px] h-1.5 bg-border rounded-full overflow-hidden">
            <div
              className={cn("h-full rounded-full transition-all duration-500", barColor)}
              style={{ width: `${mention.sentimentScore}%` }}
            />
          </div>
        </div>

        {mention.url && (
          <a
            href={mention.url}
            target="_blank"
            rel="noopener noreferrer"
            onClick={(e) => e.stopPropagation()}
            className="text-xs text-primary hover:text-primary/80 flex items-center gap-1 transition-colors cursor-pointer ml-auto"
          >
            <ExternalLink className="h-3 w-3" /> Открыть источник
          </a>
        )}
      </div>
    </div>
  );
}
