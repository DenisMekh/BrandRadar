import type { Mention } from "@/lib/api/mentions";
import type { AlertEvent } from "@/lib/api/alerts";

// Simple deterministic random generator
function createRng(seed: string) {
  let h = 0;
  for (let i = 0; i < seed.length; i++) {
    h = (Math.imul(31, h) + seed.charCodeAt(i)) | 0;
  }
  return function() {
    h = Math.imul(h ^ (h >>> 16), 0x85ebca6b);
    h = Math.imul(h ^ (h >>> 13), 0xc2b2ae35);
    h = (h ^ (h >>> 16)) >>> 0;
    return h / 0xffffffff;
  };
}

// Generate X days of diverse mention data for analytics
function generateAnalyticsMentions(brandId: string, daysCount = 30): Mention[] {
  const sources: Array<"Telegram" | "Web" | "RSS"> = ["Telegram", "Web", "RSS"];
  const sentiments: Array<"positive" | "negative" | "neutral"> = ["positive", "negative", "neutral"];
  const statuses: Array<"new" | "processing" | "processed" | "archived"> = ["new", "processed", "processed", "archived"];
  const mentions: Mention[] = [];

  const titles: Record<string, string[]> = {
    b1: ["BrandRadar обновил платформу", "Проблемы с BrandRadar", "Обзор BrandRadar", "Отзыв о BrandRadar", "BrandRadar vs конкуренты", "Баг в BrandRadar", "Новая фича BrandRadar"],
    b2: ["Конкурент А запуск", "Конкурент А сбои", "Обзор Конкурента А", "Конкурент А цены", "Конкурент А отзыв"],
    b3: ["Продукт X уязвимость", "Обновление Продукта X", "Продукт X обзор", "Продукт X на рынке"],
  };
  const brandTitles = titles[brandId] ?? titles.b1;
  const anchorDate = new Date("2026-03-15T12:00:00Z");

  for (let day = 0; day < daysCount; day++) {
    const rng = createRng(`${brandId}-${day}`);
    const count = 3 + Math.floor(rng() * 8); // 3-10 per day
    for (let j = 0; j < count; j++) {
      const date = new Date(anchorDate);
      date.setDate(date.getDate() - day);
      date.setHours(Math.floor(rng() * 24), Math.floor(rng() * 60));

      const sentimentWeights = [0.35, 0.25, 0.4]; // pos, neg, neutral
      const r = rng();
      const sentiment = r < sentimentWeights[0] ? sentiments[0] : r < sentimentWeights[0] + sentimentWeights[1] ? sentiments[1] : sentiments[2];
      const score = sentiment === "positive" ? 0.6 + rng() * 0.4 : sentiment === "negative" ? 0.05 + rng() * 0.3 : 0.35 + rng() * 0.3;

      mentions.push({
        id: `am-${brandId}-${day}-${j}`,
        author: ["news_bot", "tech_review", "user_" + j, "social_media", "forum"][j % 5],
        source: sources[j % 3],
        title: brandTitles[j % brandTitles.length],
        text: `Автоматически сгенерированное упоминание для аналитики бренда. День ${daysCount - day}, запись ${j + 1}.`,
        sentiment,
        sentiment_score: Math.round(score * 100) / 100,
        is_relevant: rng() > 0.15,
        similar_count: rng() > 0.7 ? Math.floor(rng() * 5) + 1 : 0,
        status: statuses[Math.floor(rng() * statuses.length)],
        brand_id: brandId,
        created_at: date.toISOString(),
      });
    }
  }

  return mentions.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
}

export function getMockBrandMentions(brandId: string, daysCount?: number): Mention[] {
  return generateAnalyticsMentions(brandId, daysCount);
}

export function getMockBrandAlerts(brandName: string): AlertEvent[] {
  return [
    { id: "ba1", brand_name: brandName, mention_count: 15, window_minutes: 30, triggered_at: "2026-03-14T08:00:00Z", is_negative: true },
    { id: "ba2", brand_name: brandName, mention_count: 12, window_minutes: 30, triggered_at: "2026-03-12T20:15:00Z", is_negative: true },
    { id: "ba3", brand_name: brandName, mention_count: 8, window_minutes: 60, triggered_at: "2026-03-11T14:00:00Z", is_negative: false },
    { id: "ba4", brand_name: brandName, mention_count: 22, window_minutes: 60, triggered_at: "2026-03-09T10:30:00Z", is_negative: true },
    { id: "ba5", brand_name: brandName, mention_count: 6, window_minutes: 15, triggered_at: "2026-03-07T16:45:00Z", is_negative: false },
  ];
}
