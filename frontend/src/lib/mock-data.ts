import type { Brand } from "@/lib/api/brands";
import type { Mention } from "@/lib/api/mentions";
import type { AlertConfig, AlertEvent } from "@/lib/api/alerts";
import type { AppEvent } from "@/lib/api/events";
import type { HealthStatus } from "@/lib/api/health";
import type { Source, CollectorJob } from "@/lib/api/sources";

export const mockBrands: Brand[] = [
  { id: "b1", name: "BrandRadar", keywords: ["brandradar", "бренд радар", "brand radar"], exclusions: ["brand new"], risk_words: ["утечка", "сбой", "взлом"], created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
  { id: "b2", name: "Конкурент А", keywords: ["competitor-a", "конкурент а"], exclusions: ["competitor analysis"], risk_words: ["судебный иск", "банкротство"], created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
  { id: "b3", name: "Продукт X", keywords: ["продукт x", "product-x"], exclusions: [], risk_words: ["отзыв продукции", "дефект"], created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
];

export const mockMentionsData: { data: Mention[]; total: number } = {
  total: 10,
  data: [
    { id: "m1", author: "news_bot", source_id: "s1", brand_id: "b1", title: "BrandRadar обновил платформу", text: "Компания BrandRadar выпустила крупное обновление своей платформы мониторинга.", ml: { label: "positive", score: 0.87, is_relevant: true, similar_ids: ["m2", "m3", "m4"] }, status: "new", created_at: "2026-03-14T08:30:00Z", published_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
    { id: "m2", author: "tech_review", source_id: "s2", brand_id: "b1", title: "Обзор систем мониторинга 2026", text: "В обзоре рассмотрены основные игроки рынка, включая BrandRadar и конкурентов.", ml: { label: "neutral", score: 0.5, is_relevant: true, similar_ids: ["m1"] }, status: "processed", created_at: "2026-03-14T07:15:00Z", published_at: "2026-03-14T07:15:00Z", updated_at: "2026-03-14T07:15:00Z" },
    { id: "m3", author: "angry_user", source_id: "s1", brand_id: "b1", title: "Проблемы с BrandRadar", text: "Уже третий день не могу нормально работать с дашбордом, всё тормозит.", ml: { label: "negative", score: 0.15, is_relevant: true, similar_ids: [] }, status: "new", created_at: "2026-03-14T06:00:00Z", published_at: "2026-03-14T06:00:00Z", updated_at: "2026-03-14T06:00:00Z" },
    { id: "m4", author: "blog_parser", source_id: "s2", brand_id: "b1", title: "Рынок мониторинга растёт", text: "Аналитики прогнозируют рост рынка мониторинга бренда на 25% в 2026 году.", ml: { label: "positive", score: 0.72, is_relevant: true, similar_ids: ["m1"] }, status: "processed", created_at: "2026-03-13T22:00:00Z", published_at: "2026-03-13T22:00:00Z", updated_at: "2026-03-13T22:00:00Z" },
    { id: "m5", author: "competitor_watch", source_id: "s3", brand_id: "b2", title: "Конкурент А запустил новый продукт", text: "Основной конкурент представил обновлённую версию своего решения для мониторинга.", ml: { label: "neutral", score: 0.48, is_relevant: true, similar_ids: [] }, status: "new", created_at: "2026-03-13T18:30:00Z", published_at: "2026-03-13T18:30:00Z", updated_at: "2026-03-13T18:30:00Z" },
    { id: "m6", author: "press_release", source_id: "s2", brand_id: "b1", title: "BrandRadar привлёк инвестиции", text: "Компания BrandRadar объявила о закрытии раунда серии A на $5M.", ml: { label: "positive", score: 0.92, is_relevant: true, similar_ids: [] }, status: "processed", created_at: "2026-03-13T14:00:00Z", published_at: "2026-03-13T14:00:00Z", updated_at: "2026-03-13T14:00:00Z" },
    { id: "m7", author: "forum_user", source_id: "s3", brand_id: "b1", title: "Сравнение BrandRadar и Конкурента", text: "Подробное сравнение двух платформ по функциональности и цене.", ml: { label: "neutral", score: 0.55, is_relevant: true, similar_ids: [] }, status: "archived", created_at: "2026-03-13T10:00:00Z", published_at: "2026-03-13T10:00:00Z", updated_at: "2026-03-13T10:00:00Z" },
    { id: "m8", author: "security_alert", source_id: "s1", brand_id: "b3", title: "Уязвимость в Продукте X", text: "Обнаружена критическая уязвимость в продукте X, рекомендуется обновление.", ml: { label: "negative", score: 0.08, is_relevant: true, similar_ids: [] }, status: "new", created_at: "2026-03-12T20:00:00Z", published_at: "2026-03-12T20:00:00Z", updated_at: "2026-03-12T20:00:00Z" },
    { id: "m9", author: "social_media", source_id: "s3", brand_id: "b1", title: "Отзыв пользователя о BrandRadar", text: "Пользуемся BrandRadar уже полгода, очень довольны результатами.", ml: { label: "positive", score: 0.85, is_relevant: true, similar_ids: [] }, status: "processed", created_at: "2026-03-12T15:00:00Z", published_at: "2026-03-12T15:00:00Z", updated_at: "2026-03-12T15:00:00Z" },
    { id: "m10", author: "news_aggregator", source_id: "s2", brand_id: "b1", title: "Дайджест новостей мониторинга", text: "Еженедельный дайджест новостей из мира мониторинга брендов и репутации.", ml: { label: "neutral", score: 0.5, is_relevant: false, similar_ids: [] }, status: "archived", created_at: "2026-03-12T08:00:00Z", published_at: "2026-03-12T08:00:00Z", updated_at: "2026-03-12T08:00:00Z" },
  ],
};

export const mockAlertConfigs: AlertConfig[] = [
  { id: "ac1", brand_id: "b1", threshold: 10, window_minutes: 30, cooldown_minutes: 60, sentiment_filter: "negative", enabled: true, aggression: 0.7, created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
  { id: "ac2", brand_id: "b2", threshold: 15, window_minutes: 60, cooldown_minutes: 120, sentiment_filter: "all", enabled: true, aggression: 0.3, created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
  { id: "ac3", brand_id: "b3", threshold: 5, window_minutes: 15, cooldown_minutes: 30, sentiment_filter: "negative", enabled: false, created_at: "2026-03-14T08:30:00Z", updated_at: "2026-03-14T08:30:00Z" },
];

export const mockAlertHistory: AlertEvent[] = [
  { id: "ah1", brand_id: "b1", mentions_count: 15, window_start: "2026-03-14T07:30:00Z", window_end: "2026-03-14T08:00:00Z", fired_at: "2026-03-14T08:00:00Z", created_at: "2026-03-14T08:00:00Z", config_id: "ac1" },
  { id: "ah2", brand_id: "b2", mentions_count: 22, window_start: "2026-03-13T13:30:00Z", window_end: "2026-03-13T14:30:00Z", fired_at: "2026-03-13T14:30:00Z", created_at: "2026-03-13T14:30:00Z", config_id: "ac2" },
  { id: "ah3", brand_id: "b1", mentions_count: 12, window_start: "2026-03-12T19:45:00Z", window_end: "2026-03-12T20:15:00Z", fired_at: "2026-03-12T20:15:00Z", created_at: "2026-03-12T20:15:00Z", config_id: "ac1" },
  { id: "ah4", brand_id: "b3", mentions_count: 8, window_start: "2026-03-12T09:45:00Z", window_end: "2026-03-12T10:00:00Z", fired_at: "2026-03-12T10:00:00Z", created_at: "2026-03-12T10:00:00Z", config_id: "ac3" },
  { id: "ah5", brand_id: "b2", mentions_count: 18, window_start: "2026-03-11T15:45:00Z", window_end: "2026-03-11T16:45:00Z", fired_at: "2026-03-11T16:45:00Z", created_at: "2026-03-11T16:45:00Z", config_id: "ac2" },
];

export const mockEvents: AppEvent[] = [
  { id: "e1", type: "spike_detected", payload: JSON.stringify({ brand: "BrandRadar", count: 15 }), occurred_at: "2026-03-14T08:55:00Z" },
  { id: "e2", type: "alert_fired", payload: JSON.stringify({ brand: "BrandRadar", count: 15 }), occurred_at: "2026-03-14T08:55:00Z" },
  { id: "e3", type: "mention_created", payload: JSON.stringify({ brand: "BrandRadar" }), occurred_at: "2026-03-14T08:52:00Z" },
  { id: "e4", type: "mention_created", payload: JSON.stringify({ brand: "Конкурент А" }), occurred_at: "2026-03-14T08:48:00Z" },
  { id: "e5", type: "mention_duplicated", payload: JSON.stringify({ brand: "BrandRadar" }), occurred_at: "2026-03-14T08:45:00Z" },
  { id: "e6", type: "collector_started", payload: JSON.stringify({ source: "Telegram Bot #1" }), occurred_at: "2026-03-14T08:40:00Z" },
  { id: "e7", type: "mention_created", payload: JSON.stringify({ brand: "Продукт X" }), occurred_at: "2026-03-14T08:26:00Z" },
  { id: "e8", type: "collector_failed", payload: JSON.stringify({ source: "RSS Parser", error: "Connection timeout" }), occurred_at: "2026-03-14T08:15:00Z" },
  { id: "e9", type: "source_toggled", payload: JSON.stringify({ source: "Twitter API", enabled: 0 }), occurred_at: "2026-03-14T07:00:00Z" },
  { id: "e10", type: "collector_started", payload: JSON.stringify({ source: "Web Crawler #3" }), occurred_at: "2026-03-14T05:00:00Z" },
];

export const mockHealth: HealthStatus = {
  status: "ok",
  uptime_seconds: 86400,
  version: "1.0.0",
  dependencies: {
    "PostgreSQL": "ok",
    "Redis": "ok",
    "Collector": "ok"
  },
};

export const mockSources: Source[] = [
  { id: "s1", name: "Telegram Bot #1", type: "telegram", status: "active", updated_at: "2026-03-14T08:40:00Z", created_at: "2026-03-14T08:40:00Z" },
  { id: "s2", name: "RSS Parser", type: "rss", status: "active", updated_at: "2026-03-14T06:00:00Z", created_at: "2026-03-14T06:00:00Z" },
  { id: "s3", name: "Web Crawler #1", type: "web", status: "inactive", updated_at: "2026-03-13T22:00:00Z", created_at: "2026-03-13T22:00:00Z" },
];

export const mockCollectorJobs: CollectorJob[] = [
  { id: "j1", source_name: "Telegram Bot #1", status: "completed", found: 12, created_at: "2026-03-14T08:40:00Z" },
  { id: "j2", source_name: "RSS Parser", status: "failed", found: null, created_at: "2026-03-14T06:00:00Z" },
  { id: "j3", source_name: "Web Crawler #1", status: "completed", found: 34, created_at: "2026-03-13T22:00:00Z" },
];
