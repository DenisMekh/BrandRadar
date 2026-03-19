import api from "@/lib/api";

import type { components } from "./types";

export type AlertConfig = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.AlertConfigResponse"] & {
  brand_name: string; // Missing from backend but needed for UI labels
  anomaly_window_size?: number;
  percentile?: number;
};
export type AlertEvent = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.AlertResponse"];

export const getAlertConfig = (brandId: string) =>
  api.get<{ data: AlertConfig[] }>(`/brands/${brandId}/alerts/config`).then((r) => r.data.data);

export const getAllAlertConfigs = () =>
  api.get<{ data: AlertConfig[] }>("/alerts/configs").then((r) => r.data.data);

export const createAlertConfig = (data: Omit<AlertConfig, "id" | "brand_name">) =>
  api.post<{ data: AlertConfig }>("/alerts/config", data).then((r) => r.data.data);

export const updateAlertConfig = (id: string, data: Partial<AlertConfig>) =>
  api.put<{ data: AlertConfig }>(`/alerts/config/${id}`, data).then((r) => r.data.data);

export const deleteAlertConfig = (id: string) =>
  api.delete(`/alerts/config/${id}`).then((r) => r.data);

export const getAlerts = (brandId: string, limit = 50, offset = 0) =>
  api.get<{ data: { items: AlertEvent[] } }>(`/brands/${brandId}/alerts`, { params: { limit, offset } }).then((r) => r.data.data?.items ?? []);

