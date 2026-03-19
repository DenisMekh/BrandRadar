import api from "@/lib/api";

import type { components } from "./types";

export type Brand = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.BrandResponse"];
export type BrandInput = components["schemas"]["prod-pobeda-2026_internal_controller_http_v1_dto.CreateBrandRequest"];

export const getBrands = (limit = 50, offset = 0) =>
  api.get<{ data: { items: Brand[] } }>(`/brands`, { params: { limit, offset } }).then((r) => r.data.data?.items ?? []);

export const createBrand = (data: BrandInput) =>
  api.post<{ data: Brand }>(`/brands`, data).then((r) => r.data.data);

export const updateBrand = (brandId: string, data: Partial<BrandInput>) =>
  api.put<{ data: Brand }>(`/brands/${brandId}`, data).then((r) => r.data.data);

export const deleteBrand = (brandId: string) =>
  api.delete(`/brands/${brandId}`);

export interface BrandDashboardData {
  brand_id: string;
  brand_name: string;
  total_mentions: number;
  sentiment: {
    positive: number;
    negative: number;
    neutral: number;
  };
  by_source: {
    source: string;
    count: number;
  }[];
  by_date: {
    date: string;
    count: number; // legacy
    total?: number; // actual backend
    sentiment?: {
      positive: number;
      negative: number;
      neutral: number;
    };
  }[];
  recent_alerts: number;
}

export const getBrandDashboard = (brandId: string, filters?: { date_from?: string; date_to?: string }) =>
  api.get<{ data: BrandDashboardData }>(`/brands/${brandId}/dashboard`, { params: filters }).then((r) => r.data.data);
