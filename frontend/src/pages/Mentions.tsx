import { useState, useMemo, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { Inbox } from "lucide-react";
import { endOfDay } from "date-fns";
import { MentionFilters, type Sentiment } from "@/components/mentions/MentionFilters";
import { MentionCard } from "@/components/mentions/MentionCard";
import { MentionDetailPanel } from "@/components/mentions/MentionDetailPanel";
import { Skeleton } from "@/components/shared/Skeleton";
import { EmptyState } from "@/components/shared/EmptyState";
import { ErrorBanner } from "@/components/shared/ErrorBanner";
import { useMentions } from "@/hooks/use-mentions";
import { usePageTitle } from "@/hooks/usePageTitle";
import { fmtNum } from "@/lib/format";
import type { Mention } from "@/lib/api/mentions";

const ITEMS_PER_PAGE = 10;

const Mentions = () => {
  usePageTitle("Упоминания");
  const [searchParams] = useSearchParams();

  const [search, setSearch] = useState("");
  const [sentiment, setSentiment] = useState<"all" | Sentiment>(() => {
    const s = searchParams.get("sentiment") || searchParams.get("label") || "all";
    return s as "all" | Sentiment;
  });
  const [brandId, setBrandId] = useState<string>(() => searchParams.get("brand_id") || "all");
  const [sourceId, setSourceId] = useState<string>(() => searchParams.get("source_id") || "all");
  const [onlyRelevant, setOnlyRelevant] = useState(false);
  const [dateFrom, setDateFrom] = useState<Date | undefined>(() => {
    const df = searchParams.get("date_from");
    return df ? new Date(df) : undefined;
  });
  const [dateTo, setDateTo] = useState<Date | undefined>(() => {
    const dt = searchParams.get("date_to");
    return dt ? new Date(dt) : undefined;
  });
  const [page, setPage] = useState(1);
  const [selectedMention, setSelectedMention] = useState<Mention | null>(null);

  // Sync state with URL params
  useEffect(() => {
    const bId = searchParams.get("brand_id");
    if (bId) setBrandId(bId);

    const snt = searchParams.get("sentiment") || searchParams.get("label");
    if (snt) setSentiment(snt as "all" | Sentiment);

    const src = searchParams.get("source_id");
    if (src) setSourceId(src);

    // Reset pagination on filter change from URL
    setPage(1);
  }, [searchParams]);

  const filters = useMemo(() => ({
    search: search || undefined,
    brand_id: brandId !== "all" ? brandId : undefined,
    sentiment: sentiment !== "all" ? sentiment : undefined,
    source_id: sourceId !== "all" ? sourceId : undefined,
    is_relevant: onlyRelevant ? true : undefined,
    date_from: dateFrom?.toISOString(),
    date_to: dateTo ? endOfDay(dateTo).toISOString() : undefined,
    limit: ITEMS_PER_PAGE,
    offset: (page - 1) * ITEMS_PER_PAGE,
  }), [search, brandId, sentiment, sourceId, onlyRelevant, dateFrom, dateTo, page]);

  const { data, isLoading, isError, refetch } = useMentions(filters);
  const mentions = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / ITEMS_PER_PAGE));

  return (
    <div className="space-y-4">
      <MentionFilters
        search={search} onSearchChange={setSearch}
        sentiment={sentiment} onSentimentChange={(v) => { setSentiment(v); setPage(1); }}
        onlyRelevant={onlyRelevant} onRelevantChange={(v) => { setOnlyRelevant(v); setPage(1); }}
        dateFrom={dateFrom} onDateFromChange={setDateFrom}
        dateTo={dateTo} onDateToChange={setDateTo}
        brandId={brandId} onBrandIdChange={(v) => { setBrandId(v); setPage(1); }}
        sourceId={sourceId} onSourceIdChange={(v) => { setSourceId(v); setPage(1); }}
      />

      {isError && <ErrorBanner onRetry={() => refetch()} />}

      {isLoading ? (
        <>
          <Skeleton className="w-40 h-4" />
          <div className="space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="bg-card border border-border rounded-xl p-5 space-y-3">
                <div className="flex justify-between"><Skeleton className="w-32 h-4" /><Skeleton className="w-16 h-4" /></div>
                <Skeleton className="w-3/4 h-5" /><Skeleton className="w-full h-12" />
                <div className="flex gap-3"><Skeleton className="w-24 h-5 rounded-full" /><Skeleton className="w-20 h-5" /><Skeleton className="w-[60px] h-3" /></div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <>
          <p className="text-sm text-muted-foreground">Найдено {fmtNum(total)} упоминаний</p>
          <div className="space-y-3">
            {mentions.map((m) => (
              <div key={m.id} onClick={() => setSelectedMention(m)} className="cursor-pointer">
                <MentionCard
                  mention={{
                    id: m.id,
                    author: m.author || "Аноним",
                    source: (typeof m.source === 'object' ? m.source?.type : m.source) || "Web",
                    time: m.created_at,
                    title: m.title || "Без заголовка",
                    text: m.text,
                    sentiment: (m.sentiment || m.ml?.label || "neutral") as Sentiment,
                    sentimentScore: Math.round((m.ml?.score || 0.7) * 100),
                    relevant: m.ml?.is_relevant ?? true,
                    similarCount: m.similar_count || 0,
                    url: m.url,
                  }}
                />
              </div>
            ))}
            {mentions.length === 0 && (
              <EmptyState
                icon={Inbox}
                iconClassName="h-12 w-12 text-muted-foreground/30 mb-4"
                title="Упоминаний пока нет"
                description="Запустите сбор данных для начала мониторинга."
                actionLabel="Перейти к источникам"
                actionHref="/sources"
              />
            )}
          </div>
          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-4 pt-4">
              <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1} className="text-sm text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors cursor-pointer min-h-[44px] px-3">← Назад</button>
              <span className="text-sm text-muted-foreground">Страница {fmtNum(page)} из {fmtNum(totalPages)}</span>
              <button onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages} className="text-sm text-muted-foreground hover:text-foreground disabled:opacity-30 transition-colors cursor-pointer min-h-[44px] px-3">Вперёд →</button>
            </div>
          )}
        </>
      )}

      {selectedMention && (
        <MentionDetailPanel
          mention={selectedMention}
          allMentions={mentions}
          onClose={() => setSelectedMention(null)}
          onSelect={(m) => setSelectedMention(m)}
        />
      )}
    </div>
  );
};

export default Mentions;
