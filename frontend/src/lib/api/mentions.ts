import api from "@/lib/api";

// Matches the actual backend response from GET /mentions
export interface MentionSource {
  id?: string;
  name?: string;
  type?: string;
}

export interface Mention {
  id: string;
  brand_id: string;
  source?: MentionSource;
  source_id?: string;
  text: string;
  title?: string;
  url?: string;
  author?: string;
  sentiment?: string;         // "positive" | "negative" | "neutral" — top-level in real API
  published_at?: string;
  created_at: string;
  updated_at?: string;
  status?: string;
  deduplicated?: boolean;
  external_id?: string;
  similar_mentions?: Mention[];
  similar_count?: number;
  // Legacy generated-type compat (nested ML object)
  ml?: {
    is_relevant?: boolean;
    label?: string;
    score?: number;
    similar_ids?: string[];
  };
}

export interface MentionFilters {
  brand_id?: string;
  brand_ids?: string[];
  status?: string;
  sentiment?: string;
  is_relevant?: boolean;
  date_from?: string;
  date_to?: string;
  search?: string;
  source_id?: string;
  limit?: number;
  offset?: number;
}

export const getMentions = (filters: MentionFilters) =>
  api.get<{ data: { items: Mention[]; total: number } }>("/mentions", { params: filters }).then((r) => ({
    data: r.data.data?.items ?? [],
    total: r.data.data?.total ?? 0,
  }));

export const getMention = (id: string) =>
  api.get<{ data: Mention }>(`/mentions/${id}`).then((r) => r.data.data);

export const updateMentionStatus = (id: string, status: string) =>
  api.patch<{ data: Mention }>(`/mentions/${id}/status`, { status }).then((r) => r.data.data);
