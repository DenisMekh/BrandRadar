import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getBrands, createBrand, updateBrand, deleteBrand, getBrandDashboard, type BrandInput } from "@/lib/api/brands";
import { mockBrands } from "@/lib/mock-data";
import { DEFAULT_PROJECT_ID } from "@/lib/constants";
import { ENABLE_MOCKS } from "@/lib/mock-config";

export const useBrands = () =>
  useQuery({
    queryKey: ["brands"],
    queryFn: async () => {
      if (ENABLE_MOCKS) return mockBrands;
      return await getBrands();
    },
  });

export const useBrandDashboard = (brandId?: string, filters?: { date_from?: string; date_to?: string }) =>
  useQuery({
    queryKey: ["brandDashboard", brandId, filters],
    queryFn: async () => {
      if (!brandId) throw new Error("Brand ID is required");
      return await getBrandDashboard(brandId, filters);
    },
    enabled: !!brandId,
  });

export const useCreateBrand = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ data }: { data: BrandInput }) => createBrand(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["brands"] }),
  });
};

export const useUpdateBrand = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ brandId, data }: { brandId: string; data: Partial<BrandInput> }) => updateBrand(brandId, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["brands"] }),
  });
};

export const useDeleteBrand = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ brandId }: { brandId: string }) => deleteBrand(brandId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["brands"] }),
  });
};
