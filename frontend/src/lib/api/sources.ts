import api from "@/lib/api";

import type { components } from "./types";

export type Source = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.SourceResponse"];

export interface CollectorJob {
  id: string;
  source_name: string;
  status: "idle" | "running" | "completed" | "failed";
  found: number | null;
  created_at: string;
}

export const getSources = (limit = 50, offset = 0) =>
  api.get<{ data: { items: Source[] } }>(`/sources`, { params: { limit, offset } }).then((r) => r.data.data?.items ?? []);

export const createSource = (data: { type: string; name: string; url: string }) =>
  api.post<{ data: Source }>(`/sources`, data).then((r) => r.data.data);

export const toggleSource = (id: string) =>
  api.post<{ data: Source }>(`/sources/${id}/toggle`).then((r) => r.data.data);

export const getCollectorJobs = (sourceId?: string, limit = 50, offset = 0) =>
  api.get<{ data: { items: CollectorJob[] } }>("/collector/jobs", { params: { source_id: sourceId, limit, offset } }).then((r) => r.data.data?.items ?? []);

export const createCollectorJob = (sourceId: string) =>
  api.post<{ data: CollectorJob }>("/collector/jobs", { source_id: sourceId }).then((r) => r.data.data);

export const startCollectorJob = (jobId: string) =>
  api.post<{ data: CollectorJob }>(`/collector/jobs/${jobId}/start`).then((r) => r.data.data);
