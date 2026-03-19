import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getAllAlertConfigs, updateAlertConfig, deleteAlertConfig, createAlertConfig, getAlerts } from "@/lib/api/alerts";
import type { AlertConfig } from "@/lib/api/alerts";
import { mockAlertConfigs, mockAlertHistory } from "@/lib/mock-data";
import { ENABLE_MOCKS } from "@/lib/mock-config";

export const useAlertConfigs = () =>
  useQuery({
    queryKey: ["alertConfigs"],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockAlertConfigs;
      return await getAllAlertConfigs();
    },
  });

export const useCreateAlertConfig = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<AlertConfig, "id" | "brand_name">) => createAlertConfig(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["alertConfigs"] }),
  });
};

export const useUpdateAlertConfig = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) => updateAlertConfig(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["alertConfigs"] }),
  });
};

export const useDeleteAlertConfig = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteAlertConfig(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["alertConfigs"] }),
  });
};

export const useAlertHistory = (brandId?: string) =>
  useQuery({
    queryKey: ["alertHistory", brandId],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockAlertHistory;
      if (brandId) return await getAlerts(brandId);

      // Fallback: fetch alerts for all brands if no specific brandId is requested
      const { getBrands } = await import("@/lib/api/brands");
      const brands = await getBrands(100, 0);
      const allAlertsArrays = await Promise.all(
        brands.map((b) => getAlerts(b.id || "").catch(() => []))
      );

      const combined = allAlertsArrays.flat();
      // Sort by fired_at or created_at descending
      combined.sort((a, b) => {
        const timeA = new Date(a.fired_at || a.created_at || 0).getTime();
        const timeB = new Date(b.fired_at || b.created_at || 0).getTime();
        return timeB - timeA;
      });
      return combined;
    },
  });
